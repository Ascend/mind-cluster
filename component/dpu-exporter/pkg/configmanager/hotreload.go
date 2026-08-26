/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package configmanager

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"ascend-common/common-utils/utils"

	"huawei.com/dpu-exporter/utils/logger"
)

const (
	// defaultConfigFile is the default config file path
	defaultConfigFile = "/etc/dpu-exporter/config.json"
	// configReloadDelay is the delay before processing a config change (debounce)
	configReloadDelay = 100 * time.Millisecond
)

// Config represents the dpu-exporter configuration file structure.
type Config struct {
	IntervalConfig `json:",inline"`
	// MetricWhiteList is the list of metric name patterns to keep
	MetricWhiteList []string `json:"metricWhiteList"`
}

// defaultConfig returns a Config with sensible defaults
func defaultConfig() *Config {
	return &Config{
		IntervalConfig: IntervalConfig{
			Hinicadm5CollectorInterval: 40,
			SysfsCollectorInterval:     20,
			DpuListRefreshInterval:     60,
		},
	}
}

var (
	configFilePath string
	currentConfig  *Config
	configMu       sync.RWMutex

	// reloadHooks are callbacks invoked after a successful config reload.
	// This avoids configmanager importing dpucollector/metricscollector (no cycle).
	reloadHooks   []func()
	reloadHooksMu sync.Mutex
)

// ReloadHook registers a callback to be invoked after config reload
func RegisterReloadHook(hook func()) {
	reloadHooksMu.Lock()
	defer reloadHooksMu.Unlock()
	reloadHooks = append(reloadHooks, hook)
}

// GetConfig returns the current config (thread-safe)
func GetConfig() *Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return currentConfig
}

// SetConfigFilePath sets the config file path explicitly (takes precedence over default)
func SetConfigFilePath(path string) {
	configFilePath = path
}

// resolveConfigPath returns the config file path from env or default
func resolveConfigPath() string {
	if configFilePath != "" {
		return configFilePath
	}
	return defaultConfigFile
}

// validateConfigPath checks that the resolved symlink target is under the config directory.
func validateConfigPath(resolvedPath string) bool {
	abs, err := filepath.Abs(resolvedPath)
	if err != nil {
		return false
	}
	configDir, err := filepath.Abs(filepath.Dir(resolveConfigPath()))
	if err != nil {
		return false
	}
	return abs == configDir || strings.HasPrefix(abs, configDir+string(os.PathSeparator))
}

// LoadConfig reads and parses the config file, returning an error on failure.
func LoadConfig() error {
	configFilePath = resolveConfigPath()

	data, err := utils.ReadLimitBytesWithSymlink(configFilePath, utils.Size10M, validateConfigPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Warnf("config file %s not found, using defaults", configFilePath)
			currentConfig = defaultConfig()
			return nil
		}
		return err
	}

	// Record initial hash
	h := sha256.Sum256(data)
	lastConfigHash = fmt.Sprintf("%x", h[:])

	return LoadConfigFromBytes(data)
}

// LoadConfigFromBytes parses config from raw bytes and applies it.
func LoadConfigFromBytes(data []byte) error {
	cfg := defaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return err
	}

	configMu.Lock()
	currentConfig = cfg
	configMu.Unlock()

	// Apply intervals
	cfg.IntervalConfig.Apply()

	// Load whitelist
	GetWhitelist().LoadCustom(cfg.MetricWhiteList)

	logger.Infof("config loaded from %s", configFilePath)
	return nil
}

// StartHotReload watches the config file for changes and triggers a reload.
// This is called once at startup and runs until ctx is cancelled.
func StartHotReload(ctx context.Context, group *sync.WaitGroup) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	dir := filepath.Dir(configFilePath)
	if err := watcher.Add(dir); err != nil {
		return err
	}

	group.Add(1)
	go func() {
		defer group.Done()
		defer watcher.Close()

		reloadTimer := time.NewTimer(configReloadDelay)
		reloadTimer.Stop()

		for {
			select {
			case <-ctx.Done():
				reloadTimer.Stop()
				logger.Info("config hot-reload stopped")
				return
			case event, ok := <-watcher.Events:
				if !ok {
					reloadTimer.Stop()
					return
				}
				if !shouldReload(event) {
					continue
				}
				if !reloadTimer.Stop() {
					select {
					case <-reloadTimer.C:
					default:
					}
				}
				reloadTimer.Reset(configReloadDelay)
			case err, ok := <-watcher.Errors:
				if !ok {
					reloadTimer.Stop()
					return
				}
				logger.Errorf("config watcher error: %v", err)
			case <-reloadTimer.C:
				doReloadConfig()
			}
		}
	}()

	logger.Infof("config hot-reload watching %s", dir)
	return nil
}

// shouldReload returns true if the fsnotify event indicates a config file change
func shouldReload(event fsnotify.Event) bool {
	absPath, _ := filepath.Abs(event.Name)
	cfgAbsPath, _ := filepath.Abs(configFilePath)
	if absPath != cfgAbsPath {
		return false
	}
	return event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0
}

// doReloadConfig reloads the config file and invokes registered reload hooks.
// It compares the file content hash to skip no-op reloads.
func doReloadConfig() {
	data, err := utils.ReadLimitBytesWithSymlink(configFilePath, utils.Size10M, validateConfigPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Warn("config file removed, keeping current config")
			return
		}
		logger.Errorf("config reload failed: read file error: %v", err)
		return
	}

	h := sha256.Sum256(data)
	hashStr := fmt.Sprintf("%x", h[:])
	if hashStr == lastConfigHash {
		logger.Debug("config file unchanged, skipping reload")
		return
	}
	lastConfigHash = hashStr

	logger.Info("reloading config...")
	if err := LoadConfigFromBytes(data); err != nil {
		logger.Errorf("config reload failed: %v", err)
		return
	}

	// Invoke reload hooks (e.g. chain rebuild + notify)
	reloadHooksMu.Lock()
	hooks := make([]func(), len(reloadHooks))
	copy(hooks, reloadHooks)
	reloadHooksMu.Unlock()

	for _, hook := range hooks {
		hook()
	}

	// Notify all collector goroutines to pick up changes
	NotifyConfigReload()

	logger.Info("config reload completed")
}

// --- Config reload notification (subscribe/publish) ---

var (
	reloadSubscribers   []chan struct{}
	reloadSubscribersMu sync.Mutex

	lastConfigHash string
)

// SubscribeReload registers a subscriber channel for config hot-reload signals
func SubscribeReload() <-chan struct{} {
	ch := make(chan struct{}, 1)
	reloadSubscribersMu.Lock()
	reloadSubscribers = append(reloadSubscribers, ch)
	reloadSubscribersMu.Unlock()
	return ch
}

// UnsubscribeReload removes a subscriber channel
func UnsubscribeReload(ch <-chan struct{}) {
	reloadSubscribersMu.Lock()
	defer reloadSubscribersMu.Unlock()
	for i, s := range reloadSubscribers {
		if s == ch {
			reloadSubscribers = append(reloadSubscribers[:i], reloadSubscribers[i+1:]...)
			return
		}
	}
}

// NotifyConfigReload sends a non-blocking signal to all subscribers
func NotifyConfigReload() {
	reloadSubscribersMu.Lock()
	subs := make([]chan struct{}, len(reloadSubscribers))
	copy(subs, reloadSubscribers)
	reloadSubscribersMu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- struct{}{}:
		default:
			logger.Warn("reload signal dropped: subscriber channel full")
		}
	}
}

// DrainReloadSignal drains any pending reload signals
func DrainReloadSignal(ch <-chan struct{}) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}
