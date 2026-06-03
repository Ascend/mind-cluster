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

package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

type PluginConfig struct {
	PluginName string `json:"pluginName"`
	State      string `json:"state"`
}

const (
	PluginStateOn     = "ON"
	PluginStateOff    = "OFF"
	configApplyDelay  = 5 * time.Minute
	defaultConfigPath = "/usr/local/hotResetPluginConfiguration.json"
	defaultFileSize   = 10
)

type PluginConfigMgr struct {
	mu             sync.RWMutex
	configs        []PluginConfig
	pendingConfigs []PluginConfig
	watcher        *fsnotify.Watcher
	configPath     string
	ctx            context.Context
	cancel         context.CancelFunc
	applyTimer     *time.Timer
	onConfigChange func()
}

// NewPluginConfigMgr creates a new PluginConfigMgr
func NewPluginConfigMgr(onConfigChange func()) *PluginConfigMgr {
	ctx, cancel := context.WithCancel(context.Background())
	return &PluginConfigMgr{
		configPath:     defaultConfigPath,
		ctx:            ctx,
		cancel:         cancel,
		onConfigChange: onConfigChange,
	}
}

func (pcm *PluginConfigMgr) LoadConfig() {
	if pcm.applyTimer != nil {
		pcm.applyTimer.Stop()
	}
	configs, err := pcm.readConfigFile()
	if err != nil {
		hwlog.RunLog.Warnf("read config file %s failed: %v, use default config", pcm.configPath, err)
		pcm.mu.Lock()
		pcm.configs = defaultPluginConfigs()
		pcm.mu.Unlock()
		return
	}
	hwlog.RunLog.Infof("load config file %s success, configs: %v", pcm.configPath, configs)
	pcm.mu.Lock()
	pcm.configs = configs
	pcm.mu.Unlock()
}

func (pcm *PluginConfigMgr) IsPluginEnabled(name string) bool {
	pcm.mu.RLock()
	defer pcm.mu.RUnlock()
	for _, cfg := range pcm.configs {
		if cfg.PluginName == name {
			return cfg.State == PluginStateOn
		}
	}
	return false
}

func (pcm *PluginConfigMgr) GetConfigs() []PluginConfig {
	pcm.mu.RLock()
	defer pcm.mu.RUnlock()
	result := make([]PluginConfig, len(pcm.configs))
	copy(result, pcm.configs)
	return result
}

func (pcm *PluginConfigMgr) WatchConfigChange() {
	configDir := filepath.Dir(pcm.configPath)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		hwlog.RunLog.Errorf("create fsnotify watcher failed: %v", err)
		return
	}
	pcm.watcher = watcher
	if err := watcher.Add(configDir); err != nil {
		hwlog.RunLog.Errorf("watch config dir %s failed: %v", configDir, err)
		watcher.Close()
		return
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case <-pcm.ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				pcm.handleConfigFileEvent(event)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				hwlog.RunLog.Errorf("config watcher error: %v", err)
			}
		}
	}()
}

func (pcm *PluginConfigMgr) Stop() {
	pcm.cancel()
	pcm.mu.Lock()
	if pcm.applyTimer != nil {
		pcm.applyTimer.Stop()
	}
	pcm.mu.Unlock()
}

func (pcm *PluginConfigMgr) handleConfigFileEvent(event fsnotify.Event) {
	if filepath.Base(event.Name) != filepath.Base(pcm.configPath) {
		return
	}
	if event.Op&fsnotify.Remove != 0 || event.Op&fsnotify.Rename != 0 {
		hwlog.RunLog.Infof("config file %s removed or renamed, fallback to default", pcm.configPath)
		pcm.fallbackToDefault()
		return
	}
	if event.Op&fsnotify.Create != 0 || event.Op&fsnotify.Write != 0 || event.Op&fsnotify.Chmod != 0 {
		pcm.LoadConfig()
	}
}

func (pcm *PluginConfigMgr) applyPendingConfig() {
	pcm.mu.Lock()
	pcm.configs = pcm.pendingConfigs
	pcm.pendingConfigs = nil
	pcm.mu.Unlock()
	hwlog.RunLog.Infof("plugin config applied after 5 minutes delay")
	if pcm.onConfigChange != nil {
		pcm.onConfigChange()
	}
}

func (pcm *PluginConfigMgr) fallbackToDefault() {
	pcm.mu.Lock()
	pcm.pendingConfigs = defaultPluginConfigs()
	if pcm.applyTimer != nil {
		pcm.applyTimer.Stop()
	}
	pcm.applyTimer = time.AfterFunc(configApplyDelay, func() {
		pcm.applyPendingConfig()
	})
	pcm.mu.Unlock()
}

func (pcm *PluginConfigMgr) readConfigFile() ([]PluginConfig, error) {
	path, err := utils.RealFileChecker(pcm.configPath, false, false, defaultFileSize)
	if err != nil {
		return nil, fmt.Errorf("check file failed: %w", err)
	}
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}
	var configs []PluginConfig
	if err := json.Unmarshal(fileBytes, &configs); err != nil {
		return nil, fmt.Errorf("parse json failed: %w", err)
	}
	return configs, nil
}

func defaultPluginConfigs() []PluginConfig {
	return []PluginConfig{
		{PluginName: "outbandReset", State: PluginStateOn},
		{PluginName: "resetRecord", State: PluginStateOff},
	}
}

type PluginManager struct {
	mu               sync.RWMutex
	Plugins          map[string]HotResetPlugin
	configMgr        *PluginConfigMgr
	preResetChain    []HotResetPlugin
	customResetChain []HotResetPlugin
	afterResetChain  []HotResetPlugin
}

func NewPluginManager() *PluginManager {
	pm := &PluginManager{
		Plugins: make(map[string]HotResetPlugin),
	}
	pm.configMgr = NewPluginConfigMgr(pm.OnConfigChange)
	return pm
}

func (pm *PluginManager) Init() error {
	pm.configMgr.LoadConfig()
	pm.BuildHookCache()
	pm.configMgr.WatchConfigChange()
	return nil
}

func (pm *PluginManager) RegisterPlugin(plugin HotResetPlugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin is nil")
	}
	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name is empty")
	}
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if _, exists := pm.Plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}
	pm.Plugins[name] = plugin
	hwlog.RunLog.Infof("plugin %s registered", name)
	return nil
}

func (pm *PluginManager) GetPlugin(name string) (HotResetPlugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	p, ok := pm.Plugins[name]
	return p, ok
}

func (pm *PluginManager) BuildHookCache() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.buildHookCacheLocked()
}

func (pm *PluginManager) buildHookCacheLocked() {
	var preChain, customChain, afterChain []HotResetPlugin
	for _, cfg := range pm.configMgr.GetConfigs() {
		if cfg.State != PluginStateOn {
			hwlog.RunLog.Infof("plugin %s state is %s, skip", cfg.PluginName, cfg.State)
			continue
		}
		p, ok := pm.Plugins[cfg.PluginName]
		if !ok {
			hwlog.RunLog.Warnf("plugin %s not found, skip", cfg.PluginName)
			continue
		}
		preChain = append(preChain, p)
		customChain = append(customChain, p)
		afterChain = append(afterChain, p)
		hwlog.RunLog.Infof("plugin %s hook built in cache", p.Name())
	}
	pm.preResetChain = preChain
	pm.customResetChain = customChain
	pm.afterResetChain = afterChain
	hwlog.RunLog.Infof("hook cache built: preReset=%d, customReset=%d, afterReset=%d",
		len(preChain), len(customChain), len(afterChain))
}

func (pm *PluginManager) OnConfigChange() {
	pm.BuildHookCache()
}

func (pm *PluginManager) ExecutePreReset(ctx context.Context, deviceList []ResetDevice) {
	pm.mu.RLock()
	chain := make([]HotResetPlugin, len(pm.preResetChain))
	copy(chain, pm.preResetChain)
	pm.mu.RUnlock()

	for _, p := range chain {
		pluginCtx, cancel := context.WithTimeout(ctx, PreResetTimeout)
		hwlog.RunLog.Infof("plugin %s PreReset start", p.Name())
		err := p.PreReset(pluginCtx, deviceList)
		cancel()
		if err != nil {
			hwlog.RunLog.Warnf("plugin %s PreReset failed: %v", p.Name(), err)
		}
	}
}

func (pm *PluginManager) ExecuteCustomReset(ctx context.Context, deviceList []ResetDevice, resetErr error) error {
	pm.mu.RLock()
	chain := make([]HotResetPlugin, len(pm.customResetChain))
	copy(chain, pm.customResetChain)
	pm.mu.RUnlock()

	if len(chain) == 0 {
		return resetErr
	}
	err := resetErr
	for _, p := range chain {
		pluginCtx, cancel := context.WithTimeout(ctx, CustomResetTimeout)
		hwlog.RunLog.Infof("plugin %s CustomReset start", p.Name())
		err = p.CustomReset(pluginCtx, deviceList, err)
		cancel()
		if err != nil {
			hwlog.RunLog.Warnf("plugin %s CustomReset failed: %v", p.Name(), err)
		}
	}
	return err
}

func (pm *PluginManager) ExecuteAfterReset(ctx context.Context, deviceList []ResetDevice, resetErr error) {
	pm.mu.RLock()
	chain := make([]HotResetPlugin, len(pm.afterResetChain))
	copy(chain, pm.afterResetChain)
	pm.mu.RUnlock()

	for _, p := range chain {
		pluginCtx, cancel := context.WithTimeout(ctx, AfterResetTimeout)
		hwlog.RunLog.Infof("plugin %s AfterReset start", p.Name())
		err := p.AfterReset(pluginCtx, deviceList, resetErr)
		cancel()
		if err != nil {
			hwlog.RunLog.Warnf("plugin %s AfterReset failed: %v", p.Name(), err)
		}
	}
}

func (pm *PluginManager) Stop() {
	pm.configMgr.Stop()
}

func (pm *PluginManager) GetConfigMgr() *PluginConfigMgr {
	return pm.configMgr
}

func (pm *PluginManager) GetHookChains() ([]HotResetPlugin, []HotResetPlugin, []HotResetPlugin) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	pre := make([]HotResetPlugin, len(pm.preResetChain))
	copy(pre, pm.preResetChain)
	custom := make([]HotResetPlugin, len(pm.customResetChain))
	copy(custom, pm.customResetChain)
	after := make([]HotResetPlugin, len(pm.afterResetChain))
	copy(after, pm.afterResetChain)
	return pre, custom, after
}
