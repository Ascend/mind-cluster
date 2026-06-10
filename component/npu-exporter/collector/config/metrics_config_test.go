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

// Package config for general collector
package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/metrics"
	"huawei.com/npu-exporter/v6/utils/logger"
)

func init() {
	logger.HwLogConfig = &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	logger.InitLogger("Prometheus")
	initChain()
}

func initChain() {
	common.SetChains([]common.MetricsCollector{}, []common.MetricsCollector{}, []common.MetricsCollector{})
}

func TestInitConfiguration(t *testing.T) {
	convey.Convey("TestInitConfiguration", t, func() {
		initConfiguration([]byte("test"), &presetConfigs)
		convey.So(len(presetConfigs), convey.ShouldEqual, 0)
	})
}

func TestLoadConfiguration(t *testing.T) {
	convey.Convey("TestLoadConfiguration", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		convey.Convey("load config ok", func() {
			patches.ApplyFunc(loadFromFile, func(filePath string) []byte {
				if filePath == PresetConfigPath {
					filePath = "../../build/metricConfiguration.json"
				} else if filePath == PluginConfigPath {
					filePath = "../../build/pluginConfiguration.json"
				}
				fileBytes, _ := utils.LoadFile(filePath)
				return fileBytes
			})
			defer func() {
				presetConfigs = make([]MetricsGroupConfig, 0)
				pluginConfigs = make([]MetricsGroupConfig, 0)
			}()
			loadConfiguration()
			convey.So(len(presetConfigs), convey.ShouldBeGreaterThan, 0)
			convey.So(len(pluginConfigs), convey.ShouldBeGreaterThan, 0)
		})
		convey.Convey("load config fail", func() {
			presetConfigs = make([]MetricsGroupConfig, 0)
			pluginConfigs = make([]MetricsGroupConfig, 0)
			patches.ApplyFunc(loadFromFile, func(filePath string) []byte {
				return nil
			})
			loadConfiguration()
			convey.So(len(presetConfigs), convey.ShouldEqual, len(defaultPresetConfigs))
			convey.So(len(pluginConfigs), convey.ShouldEqual, len(defaultPluginConfigs))
		})
	})
}

func TestAddPluginCollector(t *testing.T) {
	convey.Convey("TestAddPluginCollector", t, func() {
		convey.Convey("add plugin ok", func() {
			pluginCollectorMap = make(map[string]common.MetricsCollector)
			defer func() {
				pluginCollectorMap = make(map[string]common.MetricsCollector)
			}()
			err := AddPluginCollector("test", &metrics.HccsCollector{})
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("add plugin fail", func() {
			pluginCollectorMap["test"] = &metrics.HccsCollector{}
			defer func() {
				pluginCollectorMap = make(map[string]common.MetricsCollector)
			}()
			err := AddPluginCollector("test", &metrics.HccsCollector{})
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestDeletePluginCollector(t *testing.T) {
	convey.Convey("TestDeletePluginCollector", t, func() {
		convey.Convey("delete plugin ok", func() {
			pluginCollectorMap["test"] = &metrics.HccsCollector{}
			DeletePluginCollector("test")
			convey.So(pluginCollectorMap["test"], convey.ShouldBeNil)
		})
		convey.Convey("delete plugin fail", func() {
			pluginCollectorMap = make(map[string]common.MetricsCollector)
			DeletePluginCollector("test")
			convey.So(len(pluginCollectorMap), convey.ShouldEqual, 0)
		})
	})
}

func TestRegister(t *testing.T) {
	convey.Convey("TestRegister", t, func() {
		n := &common.NpuCollector{}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		// Mock IsSupported method to always return true
		patches.ApplyMethodReturn(&metrics.HccsCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.BaseInfoCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.SioCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.VersionCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.HbmCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.DdrCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.VnpuCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.PcieCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.NetworkCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.RoceCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.OpticalCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.UbCollector{}, "IsSupported", true)
		patches.ApplyFunc(loadConfiguration, func() {
			initConfiguration(loadFromFile("../../build/metricConfiguration.json"), &presetConfigs)
			initConfiguration(loadFromFile("../../build/pluginConfiguration.json"), &pluginConfigs)
		})
		patches.ApplyFunc(common.InitNpuDevNetPortInfos, func(n *common.NpuCollector) {})
		Register(n)
		convey.Convey("Should add collectors to ChainForSingleGoroutine", func() {
			convey.So(len(common.ChainForSingleGoroutine), convey.ShouldBeGreaterThan, 0)
		})
		convey.Convey("Should add collectors to ChainForMultiGoroutine", func() {
			convey.So(len(common.ChainForMultiGoroutine), convey.ShouldBeGreaterThan, 0)
		})
	})
}

func TestUnRegister(t *testing.T) {
	convey.Convey("TestUnRegister", t, func() {
		// Initialize chains with some collectors
		common.ChainForSingleGoroutine = []common.MetricsCollector{
			&metrics.HccsCollector{},
			&metrics.BaseInfoCollector{},
		}
		common.ChainForMultiGoroutine = []common.MetricsCollector{
			&metrics.NetworkCollector{},
			&metrics.RoceCollector{},
		}

		convey.Convey("When UnRegister is called with HccsCollector type", func() {
			UnRegister(reflect.TypeOf(&metrics.HccsCollector{}))

			convey.Convey("Should remove HccsCollector from ChainForSingleGoroutine", func() {
				expected := []common.MetricsCollector{
					&metrics.BaseInfoCollector{},
				}
				convey.So(len(common.ChainForSingleGoroutine), convey.ShouldEqual, len(expected))
				for i, collector := range common.ChainForSingleGoroutine {
					convey.So(reflect.TypeOf(collector), convey.ShouldEqual, reflect.TypeOf(expected[i]))
				}
			})

			convey.Convey("Should not affect ChainForMultiGoroutine", func() {
				expected := []common.MetricsCollector{
					&metrics.NetworkCollector{},
					&metrics.RoceCollector{},
				}
				convey.So(len(common.ChainForMultiGoroutine), convey.ShouldEqual, len(expected))
				for i, collector := range common.ChainForMultiGoroutine {
					convey.So(reflect.TypeOf(collector), convey.ShouldEqual, reflect.TypeOf(expected[i]))
				}
			})
		})
	})
}

func TestUnRegisterChain(t *testing.T) {
	convey.Convey("TestUnRegisterChain", t, func() {
		// Initialize a chain with some collectors
		chain := []common.MetricsCollector{
			&metrics.HccsCollector{},
			&metrics.BaseInfoCollector{},
			&metrics.NetworkCollector{},
		}

		convey.Convey("When unRegisterChain is called with BaseInfoCollector type", func() {
			unRegisterChain(reflect.TypeOf(&metrics.BaseInfoCollector{}), &chain)
			convey.Convey("Should remove BaseInfoCollector from the chain", func() {
				expected := []common.MetricsCollector{
					&metrics.HccsCollector{},
					&metrics.NetworkCollector{},
				}
				convey.So(len(chain), convey.ShouldEqual, len(expected))
				for i, collector := range chain {
					convey.So(reflect.TypeOf(collector), convey.ShouldEqual, reflect.TypeOf(expected[i]))
				}
			})
		})
	})
}

const (
	testMetricsGroup = "testGroup"
	testStateOn      = "ON"
	testInterval     = 60
	testInterval1    = 1
	testInterval5    = 5
	testIntervalNeg1 = -1
	testInterval0    = 0
	testIntervalMax  = 86401
	testPrefix       = "metricsGroup"
	testPluginPrefix = "plugin collector"
)

func TestBuildDefaultConfig(t *testing.T) {
	testCases := []struct {
		name             string
		metricsGroup     string
		state            string
		intervalSeconds  int
		expectedGroup    string
		expectedState    string
		expectedInterval int
	}{
		{
			name:             "should return config with 60s interval when interval is 60",
			metricsGroup:     testMetricsGroup,
			state:            testStateOn,
			intervalSeconds:  testInterval,
			expectedGroup:    testMetricsGroup,
			expectedState:    testStateOn,
			expectedInterval: testInterval,
		},
		{
			name:             "should return config with 1s interval when interval is 1",
			metricsGroup:     testMetricsGroup,
			state:            testStateOn,
			intervalSeconds:  testInterval1,
			expectedGroup:    testMetricsGroup,
			expectedState:    testStateOn,
			expectedInterval: testInterval1,
		},
		{
			name:             "should return config with 5s interval when interval is 5",
			metricsGroup:     testMetricsGroup,
			state:            testStateOn,
			intervalSeconds:  testInterval5,
			expectedGroup:    testMetricsGroup,
			expectedState:    testStateOn,
			expectedInterval: testInterval5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			convey.Convey(tc.name, t, func() {
				result := buildDefaultConfig(tc.metricsGroup, tc.state, tc.intervalSeconds)
				convey.So(result.MetricsGroup, convey.ShouldEqual, tc.expectedGroup)
				convey.So(result.State, convey.ShouldEqual, tc.expectedState)
				convey.So(result.IntervalSeconds, convey.ShouldNotBeNil)
				convey.So(*result.IntervalSeconds, convey.ShouldEqual, tc.expectedInterval)
			})
		})
	}
}

func TestResolveInterval(t *testing.T) {
	testCases := []struct {
		name             string
		configInterval   *int
		fallbackInterval time.Duration
		expectedInterval time.Duration
	}{
		{
			name:             "should return fallback when fallback is positive",
			configInterval:   intPtr(testInterval),
			fallbackInterval: 30 * time.Second,
			expectedInterval: 30 * time.Second,
		},
		{
			name:             "should return default when config is nil",
			configInterval:   nil,
			fallbackInterval: 0,
			expectedInterval: time.Duration(defaultIntervalSeconds) * time.Second,
		},
		{
			name:             "should return collectOnce when config is -1",
			configInterval:   intPtr(testIntervalNeg1),
			fallbackInterval: 0,
			expectedInterval: common.CollectOnceInterval(),
		},
		{
			name:             "should return disabled when config is 0",
			configInterval:   intPtr(testInterval0),
			fallbackInterval: 0,
			expectedInterval: common.DisabledInterval(),
		},
		{
			name:             "should return disabled when config exceeds max",
			configInterval:   intPtr(testIntervalMax),
			fallbackInterval: 0,
			expectedInterval: common.DisabledInterval(),
		},
		{
			name:             "should return configured interval when valid",
			configInterval:   intPtr(testInterval5),
			fallbackInterval: 0,
			expectedInterval: time.Duration(testInterval5) * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			convey.Convey(tc.name, t, func() {
				result := resolveInterval(tc.configInterval, tc.fallbackInterval)
				convey.So(result, convey.ShouldEqual, tc.expectedInterval)
			})
		})
	}
}

func intPtr(v int) *int {
	return &v
}

func TestValidateConfigs(t *testing.T) {
	t.Run("should skip off state config when state is OFF", func(t *testing.T) {
		convey.Convey("should skip off state config when state is OFF", t, func() {
			configs := []MetricsGroupConfig{
				{MetricsGroup: groupHccs, State: stateOFF, IntervalSeconds: intPtr(testInterval)},
			}
			results := validateConfigs(configs, 0, testPrefix)
			convey.So(len(results), convey.ShouldEqual, 0)
		})
	})

	t.Run("should return validated config when state is ON and interval valid", func(t *testing.T) {
		convey.Convey("should return validated config when state is ON and interval valid", t, func() {
			configs := []MetricsGroupConfig{
				{MetricsGroup: groupHccs, State: stateOn, IntervalSeconds: intPtr(testInterval5)},
			}
			results := validateConfigs(configs, 0, testPrefix)
			convey.So(len(results), convey.ShouldEqual, 1)
			convey.So(results[0].name, convey.ShouldEqual, groupHccs)
			convey.So(results[0].interval, convey.ShouldEqual,
				time.Duration(testInterval5)*time.Second)
		})
	})

	t.Run("should skip disabled config when interval is 0", func(t *testing.T) {
		convey.Convey("should skip disabled config when interval is 0", t, func() {
			configs := []MetricsGroupConfig{
				{MetricsGroup: groupHccs, State: stateOn, IntervalSeconds: intPtr(testInterval0)},
			}
			results := validateConfigs(configs, 0, testPrefix)
			convey.So(len(results), convey.ShouldEqual, 0)
		})
	})

	t.Run("should use default interval when IntervalSeconds is nil", func(t *testing.T) {
		convey.Convey("should use default interval when IntervalSeconds is nil", t, func() {
			configs := []MetricsGroupConfig{
				{MetricsGroup: groupHccs, State: stateOn, IntervalSeconds: nil},
			}
			results := validateConfigs(configs, 0, testPrefix)
			convey.So(len(results), convey.ShouldEqual, 1)
			convey.So(results[0].interval, convey.ShouldEqual,
				time.Duration(defaultIntervalSeconds)*time.Second)
		})
	})

	t.Run("should return collectOnce when interval is -1", func(t *testing.T) {
		convey.Convey("should return collectOnce when interval is -1", t, func() {
			configs := []MetricsGroupConfig{
				{MetricsGroup: groupVersion, State: stateOn, IntervalSeconds: intPtr(testIntervalNeg1)},
			}
			results := validateConfigs(configs, 0, testPrefix)
			convey.So(len(results), convey.ShouldEqual, 1)
			convey.So(results[0].interval, convey.ShouldEqual, common.CollectOnceInterval())
		})
	})

	t.Run("should use fallback when fallback is positive", func(t *testing.T) {
		convey.Convey("should use fallback when fallback is positive", t, func() {
			configs := []MetricsGroupConfig{
				{MetricsGroup: groupHccs, State: stateOn, IntervalSeconds: intPtr(testInterval5)},
			}
			results := validateConfigs(configs, 10*time.Second, testPrefix)
			convey.So(len(results), convey.ShouldEqual, 1)
			convey.So(results[0].interval, convey.ShouldEqual, 10*time.Second)
		})
	})
}

func TestMatchCollectors(t *testing.T) {
	n := &common.NpuCollector{}

	t.Run("should match collector when name exists and IsSupported", func(t *testing.T) {
		convey.Convey("should match collector when name exists and IsSupported", t, func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyMethodReturn(&metrics.HccsCollector{}, "IsSupported", true)

			validated := []validatedConfig{
				{name: groupHccs, interval: time.Duration(testInterval) * time.Second},
			}
			collectors, entries := matchCollectors(validated, singleGoroutineMap, n)
			convey.So(len(collectors), convey.ShouldEqual, 1)
			convey.So(len(entries), convey.ShouldEqual, 1)
			convey.So(entries[0].name, convey.ShouldEqual, groupHccs)
		})
	})

	t.Run("should skip collector when name not in map", func(t *testing.T) {
		convey.Convey("should skip collector when name not in map", t, func() {
			validated := []validatedConfig{
				{name: "nonExist", interval: time.Duration(testInterval) * time.Second},
			}
			collectors, entries := matchCollectors(validated, singleGoroutineMap, n)
			convey.So(len(collectors), convey.ShouldEqual, 0)
			convey.So(len(entries), convey.ShouldEqual, 0)
		})
	})

	t.Run("should skip collector when IsSupported returns false", func(t *testing.T) {
		convey.Convey("should skip collector when IsSupported returns false", t, func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyMethodReturn(&metrics.HccsCollector{}, "IsSupported", false)

			validated := []validatedConfig{
				{name: groupHccs, interval: time.Duration(testInterval) * time.Second},
			}
			collectors, entries := matchCollectors(validated, singleGoroutineMap, n)
			convey.So(len(collectors), convey.ShouldEqual, 0)
			convey.So(len(entries), convey.ShouldEqual, 0)
		})
	})

	t.Run("should return empty when validated is empty", func(t *testing.T) {
		convey.Convey("should return empty when validated is empty", t, func() {
			collectors, entries := matchCollectors([]validatedConfig{}, singleGoroutineMap, n)
			convey.So(len(collectors), convey.ShouldEqual, 0)
			convey.So(len(entries), convey.ShouldEqual, 0)
		})
	})
}

func TestLogCollectorIntervals(t *testing.T) {
	t.Run("should not panic when entries is empty", func(t *testing.T) {
		convey.Convey("should not panic when entries is empty", t, func() {
			logCollectorIntervals([]collectorIntervalEntry{})
		})
	})

	t.Run("should log grouped intervals when entries have different intervals", func(t *testing.T) {
		convey.Convey("should log grouped intervals when entries have different intervals", t, func() {
			entries := []collectorIntervalEntry{
				{name: groupHccs, interval: time.Duration(testInterval) * time.Second},
				{name: groupNpu, interval: time.Duration(testInterval) * time.Second},
				{name: groupVersion, interval: common.CollectOnceInterval()},
			}
			logCollectorIntervals(entries)
		})
	})
}
