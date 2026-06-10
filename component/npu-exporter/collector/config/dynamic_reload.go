/* Copyright(C) 2026. Huawei Technologies Co., Ltd. All rights reserved.
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

// Package config for general collector
package config

import (
	"context"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"

	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/utils/logger"
)

const (
	reloadDelay = 200 * time.Millisecond
)

var configFileNames = map[string]struct{}{
	filepath.Base(PresetConfigPath): {},
	filepath.Base(PluginConfigPath): {},
}

// StartDynamicReload starts the config file hot-reload watcher.
// It watches for changes in the config directory and triggers Register to reload config after debouncing.
//
// Supports two deployment scenarios:
//  1. Kubernetes ConfigMap mount: watches for ..data symlink Rename event (atomic update marker)
//  2. Binary deployment: watches for Write/Create/Rename events on config files
func StartDynamicReload(ctx context.Context, n *common.NpuCollector) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Errorf("failed to create fsnotify watcher: %v", err)
		return
	}
	configDir := filepath.Dir(PresetConfigPath)
	if err := w.Add(configDir); err != nil {
		logger.Errorf("failed to watch directory %s: %v", configDir, err)
		_ = w.Close()
		return
	}
	logger.Infof("start watching config directory: %s", configDir)
	go runReloadLoop(ctx, n, w, configDir)
}

func runReloadLoop(ctx context.Context, n *common.NpuCollector, w *fsnotify.Watcher, configDir string) {
	var reloadTimer *time.Timer
	defer func() {
		if reloadTimer != nil {
			reloadTimer.Stop()
		}
		_ = w.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-w.Events:
			if !ok {
				return
			}
			handleFsEvent(ev, configDir, &reloadTimer)
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			logger.Errorf("config watcher error: %v", err)
		case <-timerCh(reloadTimer):
			reloadTimer = nil
			logger.Infof("reloading metrics/plugin config")
			Register(n)
		}
	}
}

func handleFsEvent(ev fsnotify.Event, configDir string, reloadTimer **time.Timer) {
	if *reloadTimer != nil {
		// It is already within the 200ms stabilization window and no longer handles it
		return
	}
	if !isRelevantEvent(ev, configDir) {
		return
	}
	logger.Infof("detected config change: %s", ev.Name)
	*reloadTimer = time.NewTimer(reloadDelay)
}

func timerCh(t *time.Timer) <-chan time.Time {
	if t == nil {
		return nil
	}
	return t.C
}

func isRelevantEvent(ev fsnotify.Event, configDir string) bool {
	// Only care about events in the config directory
	if filepath.Dir(ev.Name) != configDir {
		return false
	}

	baseName := filepath.Base(ev.Name)

	// Case 1: Config file changes (binary deployment)
	if _, ok := configFileNames[baseName]; ok {
		return ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0
	}

	// Case 2: ..data symlink Rename (K8s ConfigMap)
	if baseName == "..data" {
		return ev.Op&fsnotify.Rename != 0
	}

	// Case 3: Timestamp directory Create (K8s ConfigMap)
	if len(baseName) > 2 && baseName[0] == '.' && baseName[1] != '.' {
		return ev.Op&fsnotify.Create != 0
	}

	return false
}
