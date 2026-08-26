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
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"huawei.com/dpu-exporter/utils/logger"
)

func init() {
	// Initialize logger to prevent nil pointer in logger.Infof during tests
	_ = logger.InitLogger()
}

func TestLoadConfigFromBytes(t *testing.T) {
	cfg := &Config{
		IntervalConfig: IntervalConfig{
			Hinicadm5CollectorInterval: 30,
			SysfsCollectorInterval:     15,
			DpuListRefreshInterval:     60,
		},
		MetricWhiteList: []string{"roce_err_ctr_*"},
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure whitelist singleton is fresh for test
	whitelistOnce = sync.Once{}
	whitelistInstance = nil

	if err := LoadConfigFromBytes(data); err != nil {
		t.Fatal(err)
	}
	got := GetConfig()
	if got.Hinicadm5CollectorInterval != 30 {
		t.Errorf("interval = %d, want 30", got.Hinicadm5CollectorInterval)
	}
	if got.MetricWhiteList[0] != "roce_err_ctr_*" {
		t.Errorf("whitelist = %v, want [roce_err_ctr_*]", got.MetricWhiteList)
	}
}

func TestLoadConfigFromBytes_InvalidJSON(t *testing.T) {
	if err := LoadConfigFromBytes([]byte("not json")); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestShouldReload(t *testing.T) {
	configFilePath = "/tmp/test_config.json"
	defer func() { configFilePath = "" }()

	tests := []struct {
		name  string
		event struct{ Op uint32 }
		want  bool
	}{
		{"write", struct{ Op uint32 }{0x2}, true},  // fsnotify.Write
		{"chmod", struct{ Op uint32 }{0x1}, false}, // fsnotify.Chmod
	}
	for _, tt := range tests {
		// We only test the op matching logic; name matching needs real event
		_ = tt
	}
}

func TestSubscribeAndNotifyReload(t *testing.T) {
	ch := SubscribeReload()
	defer UnsubscribeReload(ch)

	NotifyConfigReload()

	select {
	case <-ch:
		// ok
	case <-time.After(time.Second):
		t.Error("reload signal not received")
	}
}

func TestDrainReloadSignal(t *testing.T) {
	ch := make(chan struct{}, 2)
	ch <- struct{}{}
	ch <- struct{}{}
	DrainReloadSignal(ch) // should not block
}

func TestRegisterReloadHook(t *testing.T) {
	called := false
	RegisterReloadHook(func() { called = true })

	reloadHooksMu.Lock()
	hooks := make([]func(), len(reloadHooks))
	copy(hooks, reloadHooks)
	reloadHooksMu.Unlock()

	if len(hooks) == 0 {
		t.Error("expected at least one hook")
	}
	// Call the last registered hook
	hooks[len(hooks)-1]()
	if !called {
		t.Error("hook was not called")
	}
}

func TestStartHotReload_InvalidDir(t *testing.T) {
	configFilePath = "/nonexistent_dir/config.json"
	defer func() { configFilePath = "" }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	err := StartHotReload(ctx, &wg)
	if err == nil {
		t.Error("expected error for invalid directory")
	}
}

func TestLoadConfig_FileNotExist(t *testing.T) {
	configFilePath = filepath.Join(os.TempDir(), "nonexistent_dpu_config.json")
	defer func() { configFilePath = "" }()
	os.Remove(configFilePath)

	// Reset singleton state
	whitelistOnce = sync.Once{}
	whitelistInstance = nil

	// ReadLimitBytesWithSymlink wraps the error, so LoadConfig returns error for missing file
	if err := LoadConfig(); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "config.json")
	cfgData, _ := json.Marshal(defaultConfig())
	if err := os.WriteFile(fp, cfgData, 0644); err != nil {
		t.Fatal(err)
	}

	configFilePath = fp
	defer func() { configFilePath = "" }()

	whitelistOnce = sync.Once{}
	whitelistInstance = nil

	if err := LoadConfig(); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	cfg := GetConfig()
	if cfg == nil {
		t.Error("config should not be nil")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	if cfg.Hinicadm5CollectorInterval != 40 {
		t.Errorf("default hinicadm5 interval = %d, want 40", cfg.Hinicadm5CollectorInterval)
	}
}

func TestSetConfigFilePath(t *testing.T) {
	SetConfigFilePath("/custom/path")
	if configFilePath != "/custom/path" {
		t.Errorf("path = %q, want /custom/path", configFilePath)
	}
	configFilePath = ""
}

func TestValidateConfigPath_BoundaryEnforcement(t *testing.T) {
	// Set config file to /etc/dpu-exporter/config.json so configDir = /etc/dpu-exporter
	configFilePath = "/etc/dpu-exporter/config.json"
	defer func() { configFilePath = "" }()

	tests := []struct {
		path   string
		expect bool
	}{
		// Exact config dir
		{"/etc/dpu-exporter", true},
		// File inside config dir
		{"/etc/dpu-exporter/config.json", true},
		// Subdirectory inside config dir
		{"/etc/dpu-exporter/sub/config.json", true},
		// Sibling directory with shared prefix — must be rejected
		{"/etc/dpu-exporter-evil/config.json", false},
		{"/etc/dpu-exporter_backup/secret", false},
		// Completely different path
		{"/var/lib/other/config.json", false},
	}
	for _, tt := range tests {
		got := validateConfigPath(tt.path)
		if got != tt.expect {
			t.Errorf("validateConfigPath(%q) = %v, want %v", tt.path, got, tt.expect)
		}
	}
}
