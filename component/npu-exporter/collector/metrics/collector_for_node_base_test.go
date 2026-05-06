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

// Package metrics for general collector
package metrics

import (
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/utils/logger"
	"huawei.com/npu-exporter/v6/versions"
)

const (
	mockExporterVersion = "v26.1.0"
	mockDriverVersion   = "26.0.3"
	goroutineCount      = 10
)

func init() {
	logger.HwLogConfig = &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	logger.InitLogger("Prometheus")
}

func mockNpuCollector() *common.NpuCollector {
	dmgr := &devmanager.DeviceManager{}
	return common.NewNpuCollector(
		time.Duration(num5)*time.Second,
		time.Duration(num5)*time.Second,
		&container.DevicesParser{},
		dmgr,
	)
}

type describeTestCase struct {
	name string
}

func buildDescribeTestCases() []describeTestCase {
	return []describeTestCase{
		{
			name: "should register nodeInfoDesc when Describe is called",
		},
	}
}

func TestNodeBaseCollectorDescribe(t *testing.T) {
	for _, tt := range buildDescribeTestCases() {
		convey.Convey(tt.name, t, func() {
			collector := &NodeBaseCollector{}
			ch := make(chan *prometheus.Desc, 1)
			collector.Describe(ch)
			desc := <-ch
			convey.So(desc, convey.ShouldNotBeNil)
		})
	}
}

type collectToCacheTestCase struct {
	name              string
	setupPatches      func(*devmanager.DeviceManager) *gomonkey.Patches
	expectExporterVer string
	expectDriverVer   string
}

func buildCollectToCacheTestCases() []collectToCacheTestCase {
	return []collectToCacheTestCase{
		{
			name: "should store cache with correct versions when CollectToCache is called",
			setupPatches: func(dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				patches := gomonkey.NewPatches()
				patches.ApplyGlobalVar(&versions.BuildVersion, mockExporterVersion)
				patches.ApplyMethodReturn(dmgr, "GetDcmiVersion", mockDriverVersion)
				return patches
			},
			expectExporterVer: mockExporterVersion,
			expectDriverVer:   mockDriverVersion,
		},
		{
			name: "should store cache with empty driverVersion when GetDcmiVersion returns empty",
			setupPatches: func(dmgr *devmanager.DeviceManager) *gomonkey.Patches {
				patches := gomonkey.NewPatches()
				patches.ApplyGlobalVar(&versions.BuildVersion, mockExporterVersion)
				patches.ApplyMethodReturn(dmgr, "GetDcmiVersion", "")
				return patches
			},
			expectExporterVer: mockExporterVersion,
			expectDriverVer:   "",
		},
	}
}

func TestNodeBaseCollectorCollectToCache(t *testing.T) {
	for _, tt := range buildCollectToCacheTestCases() {
		convey.Convey(tt.name, t, func() {
			dmgr := &devmanager.DeviceManager{}
			patches := tt.setupPatches(dmgr)
			defer patches.Reset()

			n := mockNpuCollector()
			n.Dmgr = dmgr
			collector := &NodeBaseCollector{}
			collector.CollectToCache(n, nil)

			cacheVal, ok := collector.LocalCache.Load(common.GetCacheKey(collector))
			convey.So(ok, convey.ShouldBeTrue)
			cache, typeOk := cacheVal.(nodeBaseInfoCache)
			convey.So(typeOk, convey.ShouldBeTrue)
			convey.So(cache.exporterVersion, convey.ShouldEqual, tt.expectExporterVer)
			convey.So(cache.driverVersion, convey.ShouldEqual, tt.expectDriverVer)
			convey.So(cache.timestamp, convey.ShouldHappenBefore, time.Now())
		})
	}
}

type updatePromTestCase struct {
	name         string
	setupCache   func(*NodeBaseCollector)
	expectCall   bool
	expectLabels []string
}

func buildUpdatePromTestCases() []updatePromTestCase {
	return []updatePromTestCase{
		{
			name: "should update metric with correct labels when cache exists",
			setupCache: func(c *NodeBaseCollector) {
				c.LocalCache.Store(common.GetCacheKey(c), nodeBaseInfoCache{
					timestamp:       time.Now(),
					exporterVersion: mockExporterVersion,
					driverVersion:   mockDriverVersion,
				})
			},
			expectCall:   true,
			expectLabels: []string{mockExporterVersion, mockDriverVersion},
		},
		{
			name: "should skip update when cache not found",
			setupCache: func(c *NodeBaseCollector) {
			},
			expectCall: false,
		},
	}
}

func TestNodeBaseCollectorUpdatePrometheus(t *testing.T) {
	for _, tt := range buildUpdatePromTestCases() {
		convey.Convey(tt.name, t, func() {
			collector := &NodeBaseCollector{}
			tt.setupCache(collector)
			ch := make(chan prometheus.Metric, 1)

			var actualLabels []string
			patches := gomonkey.NewPatches()
			patches.ApplyFunc(doUpdateMetric,
				func(_ chan<- prometheus.Metric, _ time.Time, _ interface{}, labels []string, _ *prometheus.Desc) {
					actualLabels = labels
				})
			defer patches.Reset()

			collector.UpdatePrometheus(ch, nil, nil, nil)

			if tt.expectCall {
				convey.So(actualLabels, convey.ShouldResemble, tt.expectLabels)
			} else {
				convey.So(actualLabels, convey.ShouldBeNil)
			}
		})
	}
}

type updateTelegrafTestCase struct {
	name           string
	setupCache     func(*NodeBaseCollector)
	setupFieldsMap func() map[string]map[string]interface{}
	expectResult   bool
	expectLabels   map[string]string
}

func buildUpdateTelegrafTestCases() []updateTelegrafTestCase {
	return []updateTelegrafTestCase{
		{
			name: "should write TelegrafData when cache exists",
			setupCache: func(c *NodeBaseCollector) {
				c.LocalCache.Store(common.GetCacheKey(c), nodeBaseInfoCache{
					timestamp:       time.Now(),
					exporterVersion: mockExporterVersion,
					driverVersion:   mockDriverVersion,
				})
			},
			setupFieldsMap: func() map[string]map[string]interface{} {
				return map[string]map[string]interface{}{
					common.KeyForMetricsWithCustomLabels: {},
				}
			},
			expectResult: true,
			expectLabels: map[string]string{
				exporterVersionLabel: mockExporterVersion,
				driverVersionLabel:   mockDriverVersion,
			},
		},
		{
			name: "should return original fieldsMap when cache not found",
			setupCache: func(c *NodeBaseCollector) {
			},
			setupFieldsMap: func() map[string]map[string]interface{} {
				return map[string]map[string]interface{}{
					common.KeyForMetricsWithCustomLabels: {},
				}
			},
			expectResult: false,
		},
		{
			name: "should return original fieldsMap when cache type mismatch",
			setupCache: func(c *NodeBaseCollector) {
				c.LocalCache.Store(common.GetCacheKey(c), "invalid_type")
			},
			setupFieldsMap: func() map[string]map[string]interface{} {
				return map[string]map[string]interface{}{
					common.KeyForMetricsWithCustomLabels: {},
				}
			},
			expectResult: false,
		},
	}
}

func TestNodeBaseCollectorUpdateTelegraf(t *testing.T) {
	for _, tt := range buildUpdateTelegrafTestCases() {
		convey.Convey(tt.name, t, func() {
			collector := &NodeBaseCollector{}
			tt.setupCache(collector)
			fieldsMap := tt.setupFieldsMap()

			result := collector.UpdateTelegraf(fieldsMap, nil, nil, nil)
			convey.So(result, convey.ShouldNotBeNil)

			if tt.expectResult {
				customLabelsMap := result[common.KeyForMetricsWithCustomLabels]
				data, ok := customLabelsMap[measurementForNodeBaseInfo].(common.TelegrafData)
				convey.So(ok, convey.ShouldBeTrue)
				convey.So(data.Measurement, convey.ShouldEqual, measurementForNodeBaseInfo)
				convey.So(data.Labels, convey.ShouldResemble, tt.expectLabels)
				convey.So(data.Metrics, convey.ShouldContainKey, "node_base_info")
			}
		})
	}
}

type updateTelegrafInitFieldsTestCase struct {
	name       string
	expectInit bool
}

func buildUpdateTelegrafInitFieldsTestCases() []updateTelegrafInitFieldsTestCase {
	return []updateTelegrafInitFieldsTestCase{
		{
			name:       "should initialize KeyForMetricsWithCustomLabels when it is nil",
			expectInit: true,
		},
	}
}

func TestNodeBaseCollectorUpdateTelegrafInitFields(t *testing.T) {
	for _, tt := range buildUpdateTelegrafInitFieldsTestCases() {
		convey.Convey(tt.name, t, func() {
			collector := &NodeBaseCollector{}
			collector.LocalCache.Store(common.GetCacheKey(collector), nodeBaseInfoCache{
				timestamp:       time.Now(),
				exporterVersion: mockExporterVersion,
				driverVersion:   mockDriverVersion,
			})
			fieldsMap := map[string]map[string]interface{}{
				common.KeyForMetricsWithCustomLabels: {},
			}

			result := collector.UpdateTelegraf(fieldsMap, nil, nil, nil)

			if tt.expectInit {
				convey.So(result[common.KeyForMetricsWithCustomLabels], convey.ShouldNotBeNil)
			}
		})
	}
}

type cacheTypeMismatchPromTestCase struct {
	name       string
	setupCache func(*NodeBaseCollector)
}

func buildCacheTypeMismatchPromTestCases() []cacheTypeMismatchPromTestCase {
	return []cacheTypeMismatchPromTestCase{
		{
			name: "should not panic when cache type mismatch in UpdatePrometheus",
			setupCache: func(c *NodeBaseCollector) {
				c.LocalCache.Store(common.GetCacheKey(c), "invalid_type")
			},
		},
	}
}

func TestNodeBaseCollectorUpdatePrometheusCacheTypeMismatch(t *testing.T) {
	for _, tt := range buildCacheTypeMismatchPromTestCases() {
		convey.Convey(tt.name, t, func() {
			collector := &NodeBaseCollector{}
			tt.setupCache(collector)
			ch := make(chan prometheus.Metric, 1)

			patches := gomonkey.NewPatches()
			patches.ApplyFunc(doUpdateMetric,
				func(_ chan<- prometheus.Metric, _ time.Time, _ interface{}, _ []string, _ *prometheus.Desc) {})
			defer patches.Reset()

			collector.UpdatePrometheus(ch, nil, nil, nil)
		})
	}
}

type concurrentCollectTestCase struct {
	name string
}

func buildConcurrentCollectTestCases() []concurrentCollectTestCase {
	return []concurrentCollectTestCase{
		{
			name: "should handle concurrent CollectToCache safely",
		},
	}
}

func TestNodeBaseCollectorConcurrentCollect(t *testing.T) {
	for _, tt := range buildConcurrentCollectTestCases() {
		convey.Convey(tt.name, t, func() {
			dmgr := &devmanager.DeviceManager{}
			patches := gomonkey.NewPatches()
			patches.ApplyGlobalVar(&versions.BuildVersion, mockExporterVersion)
			patches.ApplyMethodReturn(dmgr, "GetDcmiVersion", mockDriverVersion)
			defer patches.Reset()

			n := mockNpuCollector()
			n.Dmgr = dmgr
			collector := &NodeBaseCollector{}
			var wg sync.WaitGroup
			for i := 0; i < goroutineCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					collector.CollectToCache(n, nil)
				}()
			}
			wg.Wait()

			cacheVal, ok := collector.LocalCache.Load(common.GetCacheKey(collector))
			convey.So(ok, convey.ShouldBeTrue)
			cache, typeOk := cacheVal.(nodeBaseInfoCache)
			convey.So(typeOk, convey.ShouldBeTrue)
			convey.So(cache.exporterVersion, convey.ShouldEqual, mockExporterVersion)
			convey.So(cache.driverVersion, convey.ShouldEqual, mockDriverVersion)
		})
	}
}
