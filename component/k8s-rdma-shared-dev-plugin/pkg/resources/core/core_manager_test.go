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

// Package core for common func
package core

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"ascend-common/common-utils/hwlog"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
	"github.com/smartystreets/goconvey/convey"
)

var testValidConfigJSON = `{
	"periodicUpdateInterval": 30,
	"configList": [
		{
			"resourceName": "rdma_dev_a",
			"resourcePrefix": "huawei.com",
			"rdmaHcaMax": 1000,
			"devices": [],
			"selectors": {
				"vendors": ["19e5"],
				"deviceIDs": ["a222"],
				"drivers": ["ub"]
			}
		}
	]
}`

func init() {
	ctx := context.Background()
	hwLogConfig := &hwlog.LogConfig{
		LogFileName:   "/tmp/core_manager_test.log",
		LogLevel:      0,
		MaxBackups:    3,
		MaxAge:        7,
		MaxLineLength: 1024,
	}
	_ = hwlog.InitRunLogger(hwLogConfig, ctx)
}

func newTestCoreManager() *coreResourceManager {
	return NewCoreResourceManager("/tmp/test_config.json", "huawei.com", "sock", false).(*coreResourceManager)
}

func TestNewCoreResourceManager(t *testing.T) {
	convey.Convey("When NewCoreResourceManager is called", t, func() {
		crm := NewCoreResourceManager("/tmp/test.json", "huawei.com", "sock", false).(*coreResourceManager)
		convey.Convey("Then it should return a valid coreResourceManager", func() {
			convey.So(crm, convey.ShouldNotBeNil)
			convey.So(crm.configList, convey.ShouldBeNil)
			convey.So(crm.useCdi, convey.ShouldBeFalse)
			convey.So(crm.PeriodicUpdateInterval, convey.ShouldEqual, defaultPeriodicUpdateInterval)
		})
	})
}

func TestNewCoreResourceManagerWithCDI(t *testing.T) {
	convey.Convey("When NewCoreResourceManager is called with useCdi=true", t, func() {
		crm := NewCoreResourceManager("/tmp/test.json", "huawei.com", "sock", true).(*coreResourceManager)
		convey.Convey("Then useCdi should be true", func() {
			convey.So(crm.useCdi, convey.ShouldBeTrue)
		})
	})
}

func TestReadConfigSuccess(t *testing.T) {
	crm := newTestCoreManager()
	tmpFile, err := os.CreateTemp("", "config*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.WriteString(testValidConfigJSON)
	_ = tmpFile.Close()

	convey.Convey("When config file exists and is valid", t, func() {
		crm.configFile = tmpFile.Name()
		err := crm.ReadConfig()
		convey.Convey("Then ReadConfig should succeed", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(crm.configList), convey.ShouldEqual, 1)
			convey.So(crm.configList[0].ResourceName, convey.ShouldEqual, "rdma_dev_a")
			convey.So(crm.PeriodicUpdateInterval, convey.ShouldEqual, 30*time.Second)
		})
	})
}

func TestReadConfigFileNotExist(t *testing.T) {
	crm := newTestCoreManager()
	crm.configFile = "/nonexistent/path/config.json"

	convey.Convey("When config file does not exist", t, func() {
		err := crm.ReadConfig()
		convey.Convey("Then it should use default config", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(crm.configList), convey.ShouldEqual, 1)
			convey.So(crm.configList[0].ResourceName, convey.ShouldEqual, "rdma")
		})
	})
}

func TestReadConfigInvalidJSON(t *testing.T) {
	crm := newTestCoreManager()
	tmpFile, err := os.CreateTemp("", "config*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.WriteString("invalid json")
	_ = tmpFile.Close()

	convey.Convey("When config file contains invalid JSON", t, func() {
		crm.configFile = tmpFile.Name()
		err := crm.ReadConfig()
		convey.Convey("Then it should use default config", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(crm.configList), convey.ShouldEqual, 1)
			convey.So(crm.configList[0].ResourceName, convey.ShouldEqual, "rdma")
		})
	})
}

func TestReadConfigDefaultInterval(t *testing.T) {
	config := types.UserConfigList{
		ConfigList: []types.UserConfig{
			{ResourceName: "test", ResourcePrefix: "huawei.com", RdmaHcaMax: 1000},
		},
	}
	data, _ := json.Marshal(config)
	tmpFile, err := os.CreateTemp("", "config*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write(data)
	_ = tmpFile.Close()

	crm := newTestCoreManager()
	convey.Convey("When periodicUpdateInterval is not set", t, func() {
		crm.configFile = tmpFile.Name()
		err := crm.ReadConfig()
		convey.Convey("Then it should use default interval", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(crm.PeriodicUpdateInterval, convey.ShouldEqual, defaultPeriodicUpdateInterval)
		})
	})
}

func TestUseDefaultConfig(t *testing.T) {
	crm := newTestCoreManager()

	convey.Convey("When useDefaultConfig is called", t, func() {
		err := crm.useDefaultConfig()
		convey.Convey("Then default config should be set", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(crm.configList), convey.ShouldEqual, 1)
			convey.So(crm.configList[0].ResourceName, convey.ShouldEqual, "rdma")
			convey.So(crm.configList[0].ResourcePrefix, convey.ShouldEqual, "huawei.com")
			convey.So(crm.configList[0].RdmaHcaMax, convey.ShouldEqual, 1000)
		})
	})
}

func TestValidateConfigsSuccess(t *testing.T) {
	crm := newTestCoreManager()
	err := crm.useDefaultConfig()
	if err != nil {
		t.Skip("Failed to setup test")
	}
	crm.configList[0].Selectors = types.Selectors{Vendors: []string{"19e5"}}

	convey.Convey("When config is valid", t, func() {
		err := crm.ValidateConfigs()
		convey.Convey("Then validation should succeed", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestValidateConfigsEmptyConfigList(t *testing.T) {
	crm := newTestCoreManager()

	convey.Convey("When configList is empty", t, func() {
		err := crm.ValidateConfigs()
		convey.Convey("Then validation should fail", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "no resources configuration found")
		})
	})
}

func TestValidateConfigsInvalidRdmaHcaMax(t *testing.T) {
	crm := newTestCoreManager()
	err := crm.useDefaultConfig()
	if err != nil {
		t.Skip("Failed to setup test")
	}
	crm.configList[0].RdmaHcaMax = -1

	convey.Convey("When rdmaHcaMax is negative", t, func() {
		err := crm.ValidateConfigs()
		convey.Convey("Then validation should fail", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestValidateConfigsDuplicateResourceName(t *testing.T) {
	crm := newTestCoreManager()
	crm.configList = []*types.UserConfig{
		{ResourceName: "test", ResourcePrefix: "huawei.com", RdmaHcaMax: 100, Selectors: types.Selectors{Vendors: []string{"19e5"}}},
		{ResourceName: "test", ResourcePrefix: "huawei.com", RdmaHcaMax: 100, Selectors: types.Selectors{Vendors: []string{"19e5"}}},
	}

	convey.Convey("When resource names are duplicated", t, func() {
		err := crm.ValidateConfigs()
		convey.Convey("Then validation should fail", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestValidateConfigsEmptySelectorAndDevices(t *testing.T) {
	crm := newTestCoreManager()
	crm.configList = []*types.UserConfig{
		{ResourceName: "test", ResourcePrefix: "huawei.com", RdmaHcaMax: 100},
	}

	convey.Convey("When both selectors and devices are empty", t, func() {
		err := crm.ValidateConfigs()
		convey.Convey("Then validation should fail", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestValidateConfigsBothSelectorAndDevices(t *testing.T) {
	crm := newTestCoreManager()
	crm.configList = []*types.UserConfig{
		{ResourceName: "test", ResourcePrefix: "huawei.com", RdmaHcaMax: 100, Devices: []string{"eth0"}, Selectors: types.Selectors{Vendors: []string{"19e5"}}},
	}

	convey.Convey("When both selectors and devices are provided", t, func() {
		err := crm.ValidateConfigs()
		convey.Convey("Then validation should fail", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestInitServers(t *testing.T) {
	crm := newTestCoreManager()

	convey.Convey("When InitServers is called", t, func() {
		err := crm.InitServers()
		convey.Convey("Then it should succeed", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestStartAllServersSuccess(t *testing.T) {
	crm := newTestCoreManager()

	convey.Convey("When no servers are configured", t, func() {
		err := crm.StartAllServers()
		convey.Convey("Then it should succeed", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestStopAllServersSuccess(t *testing.T) {
	crm := newTestCoreManager()

	convey.Convey("When no servers are configured", t, func() {
		err := crm.StopAllServers()
		convey.Convey("Then it should succeed", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestRestartAllServersSuccess(t *testing.T) {
	crm := newTestCoreManager()

	convey.Convey("When no servers are configured", t, func() {
		err := crm.RestartAllServers()
		convey.Convey("Then it should succeed", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetConfigList(t *testing.T) {
	crm := newTestCoreManager()

	convey.Convey("When GetConfigList is called on empty manager", t, func() {
		result := crm.GetConfigList()
		convey.Convey("Then it should return nil", func() {
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

func TestSetConfigList(t *testing.T) {
	crm := newTestCoreManager()
	testConfig := &types.UserConfig{
		ResourceName:   "test",
		ResourcePrefix: "huawei.com",
		RdmaHcaMax:     100,
	}

	convey.Convey("When SetConfigList is called", t, func() {
		crm.SetConfigList([]*types.UserConfig{testConfig})
		convey.Convey("Then configList should be updated", func() {
			convey.So(len(crm.configList), convey.ShouldEqual, 1)
			convey.So(crm.configList[0].ResourceName, convey.ShouldEqual, "test")
		})
	})
}

func TestGetResourceServers(t *testing.T) {
	crm := newTestCoreManager()

	convey.Convey("When GetResourceServers is called on empty manager", t, func() {
		result := crm.GetResourceServers()
		convey.Convey("Then it should return empty slice", func() {
			convey.So(len(result), convey.ShouldEqual, 0)
		})
	})
}

func TestResourceServersInitialState(t *testing.T) {
	crm := newTestCoreManager()

	convey.Convey("When coreResourceManager is initialized", t, func() {
		convey.Convey("Then resourceServers should be empty", func() {
			convey.So(len(crm.resourceServers), convey.ShouldEqual, 0)
		})
	})
}

func TestGetUseCdi(t *testing.T) {
	crm := newTestCoreManager()

	convey.Convey("When GetUseCdi is called", t, func() {
		result := crm.GetUseCdi()
		convey.Convey("Then it should return useCdi value", func() {
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestGetPeriodicUpdateInterval(t *testing.T) {
	crm := newTestCoreManager()

	convey.Convey("When GetPeriodicUpdateInterval is called", t, func() {
		result := crm.GetPeriodicUpdateInterval()
		convey.Convey("Then it should return PeriodicUpdateInterval value", func() {
			convey.So(result, convey.ShouldEqual, defaultPeriodicUpdateInterval)
		})
	})
}

func TestValidResourceNameSuccess(t *testing.T) {
	convey.Convey("When resource name is valid", t, func() {
		result := validResourceName("test_dev")
		convey.Convey("Then it should return true", func() {
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

func TestValidResourceNameInvalid(t *testing.T) {
	convey.Convey("When resource name contains invalid characters", t, func() {
		result := validResourceName("test-dev")
		convey.Convey("Then it should return false", func() {
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestValidResourcePrefixSuccess(t *testing.T) {
	convey.Convey("When resource prefix is valid DNS subdomain", t, func() {
		result := validResourcePrefix("huawei.com")
		convey.Convey("Then it should return true", func() {
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

func TestValidResourcePrefixInvalid(t *testing.T) {
	convey.Convey("When resource prefix contains uppercase", t, func() {
		result := validResourcePrefix("Huawei.Com")
		convey.Convey("Then it should return false", func() {
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestDetectPluginWatchModeDirExist(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "plugin_test")
	defer os.Remove(tmpDir)

	convey.Convey("When active sock dir exists", t, func() {
		result := detectPluginWatchMode(tmpDir)
		convey.Convey("Then it should return true", func() {
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

func TestDetectPluginWatchModeDirNotExist(t *testing.T) {
	convey.Convey("When active sock dir does not exist", t, func() {
		result := detectPluginWatchMode("/nonexistent/path")
		convey.Convey("Then it should return false", func() {
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestPeriodicUpdateZeroInterval(t *testing.T) {
	crm := newTestCoreManager()
	crm.PeriodicUpdateInterval = 0

	convey.Convey("When PeriodicUpdateInterval is zero", t, func() {
		stopFn := crm.PeriodicUpdate()
		convey.Convey("Then stop function should not block", func() {
			convey.So(stopFn, convey.ShouldNotBeNil)
			stopFn()
		})
	})
}
