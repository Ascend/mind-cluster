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
	name       string
	setupCache func(*NodeBaseCollector)
	expectCall bool
}

func buildUpdatePromTestCases() []updatePromTestCase {
	return []updatePromTestCase{
		{
			name: "should update metric when cache exists",
			setupCache: func(c *NodeBaseCollector) {
				c.LocalCache.Store(common.GetCacheKey(c), nodeBaseInfoCache{
					timestamp:       time.Now(),
					exporterVersion: mockExporterVersion,
					driverVersion:   mockDriverVersion,
				})
			},
			expectCall: true,
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

			collector.UpdatePrometheus(ch, nil, nil, nil)

			if tt.expectCall {
				select {
				case <-ch:
				default:
					convey.So("expected metric but channel was empty", convey.ShouldBeNil)
				}
			} else {
				select {
				case <-ch:
					convey.So("unexpected metric received", convey.ShouldBeNil)
				default:
				}
			}
		})
	}
}

type updateTelegrafTestCase struct {
	name         string
	setupCache   func(*NodeBaseCollector)
	expectData   bool
	expectLabels map[string]string
}

func buildUpdateTelegrafTestCases() []updateTelegrafTestCase {
	return []updateTelegrafTestCase{
		{
			name: "should write TelegrafMetric when cache exists",
			setupCache: func(c *NodeBaseCollector) {
				c.LocalCache.Store(common.GetCacheKey(c), nodeBaseInfoCache{
					timestamp:       time.Now(),
					exporterVersion: mockExporterVersion,
					driverVersion:   mockDriverVersion,
				})
			},
			expectData: true,
			expectLabels: map[string]string{
				exporterVersionLabel: mockExporterVersion,
				driverVersionLabel:   mockDriverVersion,
			},
		},
		{
			name: "should write nothing when cache not found",
			setupCache: func(c *NodeBaseCollector) {
			},
			expectData: false,
		},
		{
			name: "should write nothing when cache type mismatch",
			setupCache: func(c *NodeBaseCollector) {
				c.LocalCache.Store(common.GetCacheKey(c), "invalid_type")
			},
			expectData: false,
		},
	}
}

func TestNodeBaseCollectorUpdateTelegraf(t *testing.T) {
	for _, tt := range buildUpdateTelegrafTestCases() {
		convey.Convey(tt.name, t, func() {
			collector := &NodeBaseCollector{}
			tt.setupCache(collector)

			received := drainUpdateTelegraf(collector, nil, nil, nil)
			if tt.expectData {
				convey.So(received, convey.ShouldHaveLength, 1)
				convey.So(received[0].Labels, convey.ShouldResemble, tt.expectLabels)
				convey.So(received[0].Fields, convey.ShouldContainKey, "node_base_info")
			} else {
				convey.So(received, convey.ShouldBeEmpty)
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
