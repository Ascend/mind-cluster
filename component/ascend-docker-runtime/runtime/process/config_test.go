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

package process

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
)

// resetConfig resets the global config cache and sync.Once for test isolation.
func resetConfig() {
	config = nil
	configOnce = sync.Once{}
}

func init() {
	// Ensure hwlog.RunLog is initialized to avoid nil pointer in tests.
	_ = hwlog.InitRunLogger(&hwlog.LogConfig{
		LogFileName: "/dev/null",
		LogLevel:    0,
		MaxBackups:  0,
		MaxAge:      0,
		OnlyToFile:  true,
		FileMaxSize: 1,
	}, nil)
}

func TestLoadConfig_Defaults(t *testing.T) {
	resetConfig()
	// Point configFilePath to a non-existent file
	configFilePath = filepath.Join(t.TempDir(), "nonexistent.json")

	cfg := loadConfig()

	assert.Equal(t, defaultInjectionMode, cfg.InjectionMode)
}

func TestLoadConfig_CDI(t *testing.T) {
	resetConfig()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	configFilePath = cfgPath

	err := os.WriteFile(cfgPath, []byte(`{"injectionMode": "cdi"}`), 0644)
	assert.NoError(t, err)

	cfg := loadConfig()

	assert.Equal(t, cdiInjectionMode, cfg.InjectionMode)
}

func TestLoadConfig_UnknownMode(t *testing.T) {
	resetConfig()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	configFilePath = cfgPath

	err := os.WriteFile(cfgPath, []byte(`{"injectionMode": "unknown_mode"}`), 0644)
	assert.NoError(t, err)

	cfg := loadConfig()

	assert.Equal(t, defaultInjectionMode, cfg.InjectionMode, "unknown mode should fall back to legacy")
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	resetConfig()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	configFilePath = cfgPath

	err := os.WriteFile(cfgPath, []byte("{{{ garbage json }}}\n"), 0644)
	assert.NoError(t, err)

	cfg := loadConfig()

	assert.Equal(t, defaultInjectionMode, cfg.InjectionMode, "invalid JSON should fall back to legacy")
}

func TestLoadConfig_Cached(t *testing.T) {
	resetConfig()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	configFilePath = cfgPath

	// Create a valid config file
	err := os.WriteFile(cfgPath, []byte(`{"injectionMode": "cdi"}`), 0644)
	assert.NoError(t, err)

	cfg1 := loadConfig()
	assert.Equal(t, cdiInjectionMode, cfg1.InjectionMode)

	// Delete the config file — loadConfig should still return cached result
	err = os.Remove(cfgPath)
	assert.NoError(t, err)

	// Suppress expected warning via gomonkey since file is now gone,
	// but sync.Once prevents re-reading anyway.
	patches := gomonkey.ApplyFunc(hwlog.RunLog.Warnf, func(format string, args ...interface{}) {})
	defer patches.Reset()

	cfg2 := loadConfig()
	assert.Same(t, cfg1, cfg2, "cached config should return same pointer")
	assert.Equal(t, cdiInjectionMode, cfg2.InjectionMode)
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	resetConfig()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	configFilePath = cfgPath

	err := os.WriteFile(cfgPath, []byte(""), 0644)
	assert.NoError(t, err)

	cfg := loadConfig()

	// Empty JSON file: json.Unmarshal returns an error,
	// so the error path falls back to defaultInjectionMode ("legacy").
	assert.Equal(t, defaultInjectionMode, cfg.InjectionMode)
}
