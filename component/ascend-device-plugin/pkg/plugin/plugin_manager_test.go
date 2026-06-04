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
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
)

const testConfigDir = "/tmp/test_plugin_config"

var testErr = errors.New("test error")

type testPlugin struct {
	HotResetPluginAdapter
	name string
}

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		return
	}
}

func (p *testPlugin) Name() string { return p.name }

func (p *testPlugin) PreReset(_ context.Context, _ []ResetDevice) error {
	return nil
}

func (p *testPlugin) CustomReset(_ context.Context, _ []ResetDevice, resetErr error) error {
	return resetErr
}

func (p *testPlugin) AfterReset(_ context.Context, _ []ResetDevice, _ error) error {
	return nil
}

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		return
	}
}

func createTestConfigFile(t *testing.T, configs []PluginConfig) string {
	t.Helper()
	dir := filepath.Join(testConfigDir, t.Name())
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("create test dir failed: %v", err)
	}
	configPath := filepath.Join(dir, "hotResetPluginConfiguration.json")
	data, err := json.Marshal(configs)
	if err != nil {
		t.Fatalf("marshal config failed: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("write config file failed: %v", err)
	}
	return configPath
}

func cleanupTestDir(t *testing.T) {
	t.Helper()
	os.RemoveAll(filepath.Join(testConfigDir, t.Name()))
}

func TestNewPluginConfigMgr(t *testing.T) {
	convey.Convey("test NewPluginConfigMgr", t, func() {
		convey.Convey("01-creates with default path when no path set", func() {
			mgr := NewPluginConfigMgr(nil)
			convey.So(mgr.configPath, convey.ShouldEqual, defaultConfigPath)
			mgr.Stop()
		})
		convey.Convey("02-custom path can be set", func() {
			mgr := NewPluginConfigMgr(nil)
			mgr.configPath = "/custom/path"
			convey.So(mgr.configPath, convey.ShouldEqual, "/custom/path")
			mgr.Stop()
		})
	})
}

func TestPluginConfigMgr_LoadConfig(t *testing.T) {
	convey.Convey("test PluginConfigMgr LoadConfig", t, func() {
		convey.Convey("01-file not exist uses default config", func() {
			mgr := NewPluginConfigMgr(nil)
			mgr.configPath = "/nonexistent/path/config.json"
			defer mgr.Stop()
			mgr.LoadConfig()
			convey.So(mgr.IsPluginEnabled("outbandReset"), convey.ShouldBeTrue)
			convey.So(mgr.IsPluginEnabled("resetRecord"), convey.ShouldBeFalse)
		})
		convey.Convey("02-valid config file loads correctly", func() {
			configs := []PluginConfig{
				{PluginName: "outbandReset", State: PluginStateOff},
				{PluginName: "resetRecord", State: PluginStateOn},
			}
			configPath := createTestConfigFile(t, configs)
			defer cleanupTestDir(t)
			mgr := NewPluginConfigMgr(nil)
			mgr.configPath = configPath
			defer mgr.Stop()
			mgr.LoadConfig()
			convey.So(mgr.IsPluginEnabled("outbandReset"), convey.ShouldBeFalse)
			convey.So(mgr.IsPluginEnabled("resetRecord"), convey.ShouldBeTrue)
		})
	})
}

func TestPluginConfigMgr_IsPluginEnabled(t *testing.T) {
	convey.Convey("test PluginConfigMgr IsPluginEnabled", t, func() {
		convey.Convey("01-unknown plugin returns false", func() {
			mgr := NewPluginConfigMgr(nil)
			mgr.configPath = "/nonexistent/path"
			defer mgr.Stop()
			mgr.LoadConfig()
			convey.So(mgr.IsPluginEnabled("unknownPlugin"), convey.ShouldBeFalse)
		})
	})
}

func TestPluginConfigMgr_readConfigFile(t *testing.T) {
	convey.Convey("test PluginConfigMgr readConfigFile", t, func() {
		convey.Convey("01-invalid json returns error", func() {
			dir := filepath.Join(testConfigDir, t.Name())
			os.MkdirAll(dir, 0755)
			defer os.RemoveAll(dir)
			configPath := filepath.Join(dir, "config.json")
			os.WriteFile(configPath, []byte("invalid json"), 0644)
			mgr := NewPluginConfigMgr(nil)
			mgr.configPath = configPath
			defer mgr.Stop()
			_, err := mgr.readConfigFile()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestPluginManager_RegisterPlugin(t *testing.T) {
	convey.Convey("test PluginManager RegisterPlugin", t, func() {
		convey.Convey("01-register plugin successfully", func() {
			pm := NewPluginManager()
			defer pm.Stop()
			p := &testPlugin{name: "testPlugin"}
			err := pm.RegisterPlugin(p)
			convey.So(err, convey.ShouldBeNil)
			_, ok := pm.GetPlugin("testPlugin")
			convey.So(ok, convey.ShouldBeTrue)
		})
		convey.Convey("02-duplicate registration returns error", func() {
			pm := NewPluginManager()
			defer pm.Stop()
			p := &testPlugin{name: "testPlugin"}
			pm.RegisterPlugin(p)
			err := pm.RegisterPlugin(p)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestPluginManager_RegisterPlugin_Invalid(t *testing.T) {
	convey.Convey("test PluginManager RegisterPlugin invalid input", t, func() {
		convey.Convey("01-nil plugin returns error", func() {
			pm := NewPluginManager()
			defer pm.Stop()
			err := pm.RegisterPlugin(nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-empty name plugin returns error", func() {
			pm := NewPluginManager()
			defer pm.Stop()
			p := &testPlugin{name: ""}
			err := pm.RegisterPlugin(p)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestPluginManager_BuildHookCache(t *testing.T) {
	convey.Convey("test PluginManager BuildHookCache", t, func() {
		convey.Convey("01-default config only enables outbandReset", func() {
			pm := NewPluginManager()
			defer pm.Stop()
			pm.configMgr.configPath = "/nonexistent/path"
			pm.configMgr.LoadConfig()
			pm.RegisterPlugin(&testPlugin{name: "outbandReset"})
			pm.RegisterPlugin(&testPlugin{name: "resetRecord"})
			pm.BuildHookCache()
			pre, custom, after := pm.GetHookChains()
			convey.So(len(pre), convey.ShouldEqual, 1)
			convey.So(len(custom), convey.ShouldEqual, 1)
			convey.So(len(after), convey.ShouldEqual, 1)
			convey.So(pre[0].Name(), convey.ShouldEqual, "outbandReset")
		})
		convey.Convey("02-all plugins enabled builds full chains", func() {
			pm := NewPluginManager()
			defer pm.Stop()
			pm.configMgr.mu.Lock()
			pm.configMgr.configs = []PluginConfig{
				{PluginName: "outbandReset", State: PluginStateOn},
				{PluginName: "resetRecord", State: PluginStateOn},
			}
			pm.configMgr.mu.Unlock()
			pm.RegisterPlugin(&testPlugin{name: "outbandReset"})
			pm.RegisterPlugin(&testPlugin{name: "resetRecord"})
			pm.BuildHookCache()
			pre, custom, after := pm.GetHookChains()
			convey.So(len(pre), convey.ShouldEqual, 2)
			convey.So(len(custom), convey.ShouldEqual, 2)
			convey.So(len(after), convey.ShouldEqual, 2)
		})
	})
}

func TestPluginManager_ExecutePreReset(t *testing.T) {
	convey.Convey("test PluginManager ExecutePreReset", t, func() {
		convey.Convey("01-executes preReset plugins in order", func() {
			pm := NewPluginManager()
			defer pm.Stop()
			pm.configMgr.mu.Lock()
			pm.configMgr.configs = []PluginConfig{
				{PluginName: "testPlugin", State: PluginStateOn},
			}
			pm.configMgr.mu.Unlock()
			pm.RegisterPlugin(&testPlugin{name: "testPlugin"})
			pm.BuildHookCache()
			pm.ExecutePreReset(context.Background(), []ResetDevice{{LogicID: 0}})
		})
	})
}

func TestPluginManager_ExecuteCustomReset(t *testing.T) {
	convey.Convey("test PluginManager ExecuteCustomReset", t, func() {
		convey.Convey("01-empty chain returns input error", func() {
			pm := NewPluginManager()
			defer pm.Stop()
			pm.BuildHookCache()
			err := pm.ExecuteCustomReset(context.Background(), nil, testErr)
			convey.So(err, convey.ShouldEqual, testErr)
		})
		convey.Convey("02-chain passes error through", func() {
			pm := NewPluginManager()
			defer pm.Stop()
			pm.configMgr.mu.Lock()
			pm.configMgr.configs = []PluginConfig{
				{PluginName: "testPlugin", State: PluginStateOn},
			}
			pm.configMgr.mu.Unlock()
			pm.RegisterPlugin(&testPlugin{name: "testPlugin"})
			pm.BuildHookCache()
			err := pm.ExecuteCustomReset(context.Background(), nil, testErr)
			convey.So(err, convey.ShouldEqual, testErr)
		})
	})
}

func TestPluginManager_ExecuteAfterReset(t *testing.T) {
	convey.Convey("test PluginManager ExecuteAfterReset", t, func() {
		convey.Convey("01-executes afterReset plugins", func() {
			pm := NewPluginManager()
			defer pm.Stop()
			pm.configMgr.mu.Lock()
			pm.configMgr.configs = []PluginConfig{
				{PluginName: "testPlugin", State: PluginStateOn},
			}
			pm.configMgr.mu.Unlock()
			pm.RegisterPlugin(&testPlugin{name: "testPlugin"})
			pm.BuildHookCache()
			pm.ExecuteAfterReset(context.Background(), []ResetDevice{{LogicID: 0}}, nil)
		})
	})
}

func TestPluginConfigMgr_fallbackToDefault(t *testing.T) {
	convey.Convey("test PluginConfigMgr fallbackToDefault", t, func() {
		convey.Convey("01-fallback sets pending to default", func() {
			mgr := NewPluginConfigMgr(nil)
			mgr.configPath = "/nonexistent/path"
			defer mgr.Stop()
			mgr.fallbackToDefault()
			mgr.mu.RLock()
			pending := mgr.pendingConfigs
			mgr.mu.RUnlock()
			convey.So(len(pending), convey.ShouldEqual, 2)
			convey.So(pending[0].PluginName, convey.ShouldEqual, "outbandReset")
		})
	})
}

func TestPluginConfigMgr_applyPendingConfig(t *testing.T) {
	convey.Convey("test PluginConfigMgr applyPendingConfig", t, func() {
		convey.Convey("01-applies pending config and calls callback", func() {
			called := false
			mgr := NewPluginConfigMgr(func() {
				called = true
			})
			defer mgr.Stop()
			mgr.mu.Lock()
			mgr.pendingConfigs = []PluginConfig{
				{PluginName: "testPlugin", State: PluginStateOn},
			}
			mgr.mu.Unlock()
			mgr.applyPendingConfig()
			convey.So(called, convey.ShouldBeTrue)
			convey.So(mgr.IsPluginEnabled("testPlugin"), convey.ShouldBeTrue)
		})
	})
}

func TestPluginConfigMgr_reloadConfig(t *testing.T) {
	convey.Convey("test PluginConfigMgr reloadConfig", t, func() {
		convey.Convey("01-parse failure keeps current config", func() {
			dir := filepath.Join(testConfigDir, t.Name())
			os.MkdirAll(dir, 0755)
			defer os.RemoveAll(dir)
			configPath := filepath.Join(dir, "config.json")
			os.WriteFile(configPath, []byte("invalid"), 0644)
			mgr := NewPluginConfigMgr(nil)
			mgr.configPath = configPath
			defer mgr.Stop()
			mgr.configs = defaultPluginConfigs()
			mgr.LoadConfig()
			convey.So(mgr.IsPluginEnabled("outbandReset"), convey.ShouldBeTrue)
		})
	})
}

func TestDefaultPluginConfigs(t *testing.T) {
	convey.Convey("test defaultPluginConfigs", t, func() {
		convey.Convey("01-returns correct default configs", func() {
			configs := defaultPluginConfigs()
			convey.So(len(configs), convey.ShouldEqual, 2)
			convey.So(configs[0].PluginName, convey.ShouldEqual, "outbandReset")
			convey.So(configs[0].State, convey.ShouldEqual, PluginStateOn)
			convey.So(configs[1].PluginName, convey.ShouldEqual, "resetRecord")
			convey.So(configs[1].State, convey.ShouldEqual, PluginStateOff)
		})
	})
}

func TestPluginConfigMgr_handleConfigFileEvent(t *testing.T) {
	convey.Convey("test PluginConfigMgr handleConfigFileEvent", t, func() {
		convey.Convey("01-ignore non-target file events, fallback to default", func() {
			mgr := NewPluginConfigMgr(nil)
			mgr.configPath = "/some/dir/hotResetPluginConfiguration.json"
			defer mgr.Stop()
			mgr.configs = defaultPluginConfigs()
			mgr.handleConfigFileEvent(fsnotify.Event{Name: "/some/dir/other.json", Op: fsnotify.Create})
			convey.So(mgr.IsPluginEnabled("outbandReset"), convey.ShouldBeTrue)
		})
	})
}

func TestPluginManager_Init(t *testing.T) {
	convey.Convey("test PluginManager Init", t, func() {
		convey.Convey("01-init with non-existent config uses defaults", func() {
			pm := NewPluginManager()
			pm.configMgr.configPath = "/nonexistent/path/config.json"
			defer pm.Stop()
			err := pm.Init()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestPluginManager_OnConfigChange(t *testing.T) {
	convey.Convey("test PluginManager OnConfigChange", t, func() {
		convey.Convey("01-rebuilds hook cache on config change", func() {
			pm := NewPluginManager()
			defer pm.Stop()
			pm.configMgr.configPath = "/nonexistent/path"
			pm.configMgr.LoadConfig()
			pm.RegisterPlugin(&testPlugin{name: "outbandReset"})
			pm.BuildHookCache()
			_, customBefore, _ := pm.GetHookChains()
			convey.So(len(customBefore), convey.ShouldEqual, 1)
			pm.configMgr.mu.Lock()
			pm.configMgr.configs = []PluginConfig{
				{PluginName: "outbandReset", State: PluginStateOff},
			}
			pm.configMgr.mu.Unlock()
			pm.OnConfigChange()
			_, customAfter, _ := pm.GetHookChains()
			convey.So(len(customAfter), convey.ShouldEqual, 0)
		})
	})
}

func TestPluginConfigMgr_GetConfigs(t *testing.T) {
	convey.Convey("test PluginConfigMgr GetConfigs", t, func() {
		convey.Convey("01-returns copy of configs", func() {
			mgr := NewPluginConfigMgr(nil)
			defer mgr.Stop()
			mgr.configs = []PluginConfig{
				{PluginName: "testPlugin", State: PluginStateOn},
			}
			result := mgr.GetConfigs()
			convey.So(len(result), convey.ShouldEqual, 1)
			convey.So(result[0].PluginName, convey.ShouldEqual, "testPlugin")
			result[0].State = PluginStateOff
			convey.So(mgr.configs[0].State, convey.ShouldEqual, PluginStateOn)
		})
	})
}

func TestPluginManager_GetPlugin(t *testing.T) {
	convey.Convey("test PluginManager GetPlugin", t, func() {
		convey.Convey("01-returns plugin if registered", func() {
			pm := NewPluginManager()
			defer pm.Stop()
			p := &testPlugin{name: "testPlugin"}
			pm.RegisterPlugin(p)
			got, ok := pm.GetPlugin("testPlugin")
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(got, convey.ShouldEqual, p)
		})
		convey.Convey("02-returns false for unregistered plugin", func() {
			pm := NewPluginManager()
			defer pm.Stop()
			_, ok := pm.GetPlugin("unknown")
			convey.So(ok, convey.ShouldBeFalse)
		})
	})
}

func TestPluginConfigMgr_Stop(t *testing.T) {
	convey.Convey("test PluginConfigMgr Stop", t, func() {
		convey.Convey("01-stop cancels context", func() {
			mgr := NewPluginConfigMgr(nil)
			mgr.Stop()
			select {
			case <-mgr.ctx.Done():
			default:
				t.Fatal("context should be cancelled after stop")
			}
		})
	})
}

func TestPluginManager_BuildHookCache_DisabledPlugin(t *testing.T) {
	convey.Convey("test BuildHookCache with disabled plugin", t, func() {
		convey.Convey("01-disabled plugin not added to chains", func() {
			pm := NewPluginManager()
			defer pm.Stop()
			pm.configMgr.mu.Lock()
			pm.configMgr.configs = []PluginConfig{
				{PluginName: "outbandReset", State: PluginStateOff},
			}
			pm.configMgr.mu.Unlock()
			pm.RegisterPlugin(&testPlugin{name: "outbandReset"})
			pm.BuildHookCache()
			pre, custom, after := pm.GetHookChains()
			convey.So(len(pre), convey.ShouldEqual, 0)
			convey.So(len(custom), convey.ShouldEqual, 0)
			convey.So(len(after), convey.ShouldEqual, 0)
		})
	})
}
