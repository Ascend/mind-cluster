/* Copyright(C) 2026-2026. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
)

const (
	testLogicID0     = int32(0)
	defaultUtilValue = -1
	testAicUtil      = uint32(50)
	testAivUtil      = uint32(60)
	testAicoreUtil   = uint32(70)
	testNpuUtil      = uint32(80)
	testAICoreUtil   = uint32(75)
	testVectorUtil   = uint32(65)
	testOverallUtil  = uint32(85)
	apiCallFailedMsg = "api call failed"
)

// TestIsSupportNetworkHealthDevices
func TestIsSupportNetworkHealthDevices(t *testing.T) {
	cases := []struct {
		name        string
		devType     string
		mainBoardId uint32
		expected    bool
	}{
		{"Ascend910A3, 0 => true", api.Ascend910A3, 0, true},
		{"Ascend910A5, Atlas9501DMainBoardID => true", api.Ascend910A5, api.Atlas9501DMainBoardID, true},
		{"Ascend910A5, Atlas3504PMainBoardID => false", api.Ascend910A5, api.Atlas3504PMainBoardID, false},
		{"Ascend910A5, daYuMainBoardId => true", api.Ascend910A5, daYuMainBoardId, true},
		{"Ascend910A5, yinHeMainBoardId => true", api.Ascend910A5, yinHeMainBoardId, true},
		{"Ascend910A5, ubxMainBoardId => true", api.Ascend910A5, ubxMainBoardId, true},
	}
	for _, c := range cases {
		convey.Convey(c.name, t, func() {
			result := isSupportNetworkHealthDevices(c.devType, c.mainBoardId)
			convey.So(result, convey.ShouldEqual, c.expected)
		})
	}
}

const (
	testCardNum        = int32(4)
	testUnsupportedDev = "UnsupportedDevType"
	testSupportedDev   = api.Ascend910A
)

type collectCardNumTestCase struct {
	name        string
	devType     string
	cardNum     int32
	cardListErr error
	expectCache bool
	expectValue int32
}

func buildCollectCardNumTestCases() []collectCardNumTestCase {
	return []collectCardNumTestCase{
		{
			name:        "should return early when device type is not supported",
			devType:     testUnsupportedDev,
			cardNum:     0,
			cardListErr: nil,
			expectCache: false,
			expectValue: 0,
		},
		{
			name:        "should log error and return when GetCardList fails",
			devType:     testSupportedDev,
			cardNum:     0,
			cardListErr: errors.New(apiCallFailedMsg),
			expectCache: false,
			expectValue: 0,
		},
		{
			name:        "should store card number to cache when GetCardList succeeds",
			devType:     testSupportedDev,
			cardNum:     testCardNum,
			cardListErr: nil,
			expectCache: true,
			expectValue: testCardNum,
		},
	}
}

func TestCollectCardNum(t *testing.T) {
	for _, tt := range buildCollectCardNumTestCases() {
		convey.Convey(tt.name, t, func() {
			dmgr := &devmanager.DeviceManager{}
			patches := gomonkey.NewPatches()
			patches.ApplyMethodReturn(dmgr, "GetDevType", tt.devType)
			patches.ApplyMethodReturn(dmgr, "GetCardList", tt.cardNum, []int32{}, tt.cardListErr)
			defer patches.Reset()

			n := &colcommon.NpuCollector{Dmgr: dmgr}
			c := &BaseInfoCollector{}

			collectCardNum(n, c)

			cacheVal, ok := c.LocalCache.Load(colcommon.MachineInfoCardDescKey)
			convey.So(ok, convey.ShouldEqual, tt.expectCache)
			if tt.expectCache {
				convey.So(cacheVal, convey.ShouldEqual, tt.expectValue)
			}
		})
	}
}

type updateMachineInfoCardMetricTestCase struct {
	name         string
	setupCache   func(*sync.Map)
	expectMetric bool
}

func buildUpdateMachineInfoCardMetricTestCases() []updateMachineInfoCardMetricTestCase {
	return []updateMachineInfoCardMetricTestCase{
		{
			name: "should not send metric when cache key not found",
			setupCache: func(localCache *sync.Map) {
			},
			expectMetric: false,
		},
		{
			name: "should not send metric when cache value type is wrong",
			setupCache: func(localCache *sync.Map) {
				localCache.Store(colcommon.MachineInfoCardDescKey, "invalid_type")
			},
			expectMetric: false,
		},
		{
			name: "should send metric when cache value is valid int32",
			setupCache: func(localCache *sync.Map) {
				localCache.Store(colcommon.MachineInfoCardDescKey, int32(testCardNum))
			},
			expectMetric: true,
		},
	}
}

func TestUpdateMachineInfoCardMetric(t *testing.T) {
	for _, tt := range buildUpdateMachineInfoCardMetricTestCases() {
		convey.Convey(tt.name, t, func() {
			localCache := &sync.Map{}
			tt.setupCache(localCache)

			ch := make(chan prometheus.Metric, 1)
			go func() {
				updateMachineInfoCardMetric(ch, localCache)
				close(ch)
			}()

			var metric prometheus.Metric
			var received bool
			for m := range ch {
				metric = m
				received = true
			}

			convey.So(received, convey.ShouldEqual, tt.expectMetric)
			if tt.expectMetric {
				convey.So(metric, convey.ShouldNotBeNil)
			}
		})
	}
}

const (
	testLogicID           = int32(0)
	testHealthCode        = 0
	testNetworkHealthCode = 0
)

type getNetworkHealthyTestCase struct {
	name         string
	logicID      int32
	netCode      uint32
	getNetErr    error
	expectStatus string
}

func buildGetNetworkHealthyTestCases() []getNetworkHealthyTestCase {
	return []getNetworkHealthyTestCase{
		{
			name:         "should return Unknown when GetDeviceNetWorkHealth fails",
			logicID:      testLogicID,
			netCode:      0,
			getNetErr:    errors.New(apiCallFailedMsg),
			expectStatus: colcommon.Unknown,
		},
		{
			name:         "should return Healthy when netCode is NetworkInit",
			logicID:      testLogicID,
			netCode:      common.NetworkInit,
			getNetErr:    nil,
			expectStatus: colcommon.Healthy,
		},
		{
			name:         "should return Healthy when netCode is NetworkSuccess",
			logicID:      testLogicID,
			netCode:      common.NetworkSuccess,
			getNetErr:    nil,
			expectStatus: colcommon.Healthy,
		},
		{
			name:         "should return UnHealthy when netCode is other value",
			logicID:      testLogicID,
			netCode:      1,
			getNetErr:    nil,
			expectStatus: colcommon.UnHealthy,
		},
	}
}

func TestGetNetworkHealthy(t *testing.T) {
	for _, tt := range buildGetNetworkHealthyTestCases() {
		convey.Convey(tt.name, t, func() {
			dmgr := &devmanager.DeviceManager{}
			patches := gomonkey.ApplyMethodReturn(dmgr, "GetDeviceNetWorkHealth", tt.netCode, tt.getNetErr)
			defer patches.Reset()

			result := getNetworkHealthy(tt.logicID, dmgr)
			convey.So(result, convey.ShouldEqual, tt.expectStatus)
		})
	}
}

type getHealthTestCase struct {
	name         string
	logicID      int32
	healthCode   uint32
	getHealthErr error
	expectStatus string
}

func buildGetHealthTestCases() []getHealthTestCase {
	return []getHealthTestCase{
		{
			name:         "should return Unknown when GetDeviceHealth fails",
			logicID:      testLogicID,
			healthCode:   0,
			getHealthErr: errors.New(apiCallFailedMsg),
			expectStatus: colcommon.Unknown,
		},
		{
			name:         "should return Healthy when health is 0",
			logicID:      testLogicID,
			healthCode:   0,
			getHealthErr: nil,
			expectStatus: colcommon.Healthy,
		},
		{
			name:         "should return UnHealthy when health is not 0",
			logicID:      testLogicID,
			healthCode:   1,
			getHealthErr: nil,
			expectStatus: colcommon.UnHealthy,
		},
	}
}

func TestGetHealth(t *testing.T) {
	for _, tt := range buildGetHealthTestCases() {
		convey.Convey(tt.name, t, func() {
			dmgr := &devmanager.DeviceManager{}
			patches := gomonkey.ApplyMethodReturn(dmgr, "GetDeviceHealth", tt.healthCode, tt.getHealthErr)
			defer patches.Reset()

			result := getHealth(tt.logicID, dmgr)
			convey.So(result, convey.ShouldEqual, tt.expectStatus)
		})
	}
}

type getHealthCodeTestCase struct {
	name         string
	healthStatus string
	expectCode   int
}

func buildGetHealthCodeTestCases() []getHealthCodeTestCase {
	return []getHealthCodeTestCase{
		{
			name:         "should return UnRetError when health is NotReport",
			healthStatus: colcommon.NotReport,
			expectCode:   common.UnRetError,
		},
		{
			name:         "should return FailedValue when health is Unknown",
			healthStatus: colcommon.Unknown,
			expectCode:   common.FailedValue,
		},
		{
			name:         "should return HealthyCode when health is Healthy",
			healthStatus: colcommon.Healthy,
			expectCode:   colcommon.HealthyCode,
		},
		{
			name:         "should return UnhealthyCode when health is UnHealthy",
			healthStatus: colcommon.UnHealthy,
			expectCode:   colcommon.UnhealthyCode,
		},
	}
}

func TestGetHealthCode(t *testing.T) {
	for _, tt := range buildGetHealthCodeTestCases() {
		convey.Convey(tt.name, t, func() {
			result := getHealthCode(tt.healthStatus)
			convey.So(result, convey.ShouldEqual, tt.expectCode)
		})
	}
}

type countHealthStatusTestCase struct {
	name            string
	chips           []colcommon.HuaWeiAIChip
	caches          map[int32]chipCache
	expectHealthy   int32
	expectUnhealthy int32
	expectUnknown   int32
}

func buildCountHealthStatusTestCases() []countHealthStatusTestCase {
	return []countHealthStatusTestCase{
		{
			name:  "should count all healthy chips",
			chips: []colcommon.HuaWeiAIChip{{PhyId: 0}, {PhyId: 1}},
			caches: map[int32]chipCache{
				0: {HealthStatus: colcommon.Healthy},
				1: {HealthStatus: colcommon.Healthy},
			},
			expectHealthy:   2,
			expectUnhealthy: 0,
			expectUnknown:   0,
		},
		{
			name:  "should count mixed health statuses",
			chips: []colcommon.HuaWeiAIChip{{PhyId: 0}, {PhyId: 1}, {PhyId: 2}, {PhyId: 3}},
			caches: map[int32]chipCache{
				0: {HealthStatus: colcommon.Healthy},
				1: {HealthStatus: colcommon.UnHealthy},
				2: {HealthStatus: colcommon.Unknown},
				3: {HealthStatus: colcommon.UnHealthy},
			},
			expectHealthy:   1,
			expectUnhealthy: 2,
			expectUnknown:   1,
		},
		{
			name:  "should count chips missing from cache as unknown",
			chips: []colcommon.HuaWeiAIChip{{PhyId: 0}, {PhyId: 1}},
			caches: map[int32]chipCache{
				0: {HealthStatus: colcommon.Healthy},
			},
			expectHealthy:   1,
			expectUnhealthy: 0,
			expectUnknown:   1,
		},
		{
			name:  "should treat unsupported health status as unknown",
			chips: []colcommon.HuaWeiAIChip{{PhyId: 0}, {PhyId: 1}},
			caches: map[int32]chipCache{
				0: {HealthStatus: colcommon.Healthy},
				1: {HealthStatus: ""},
			},
			expectHealthy:   1,
			expectUnhealthy: 0,
			expectUnknown:   1,
		},
		{
			name:            "should return zero when no chips",
			chips:           []colcommon.HuaWeiAIChip{},
			caches:          map[int32]chipCache{0: {HealthStatus: colcommon.Healthy}},
			expectHealthy:   0,
			expectUnhealthy: 0,
			expectUnknown:   0,
		},
	}
}

func TestCountHealthStatus(t *testing.T) {
	for _, tt := range buildCountHealthStatusTestCases() {
		convey.Convey(tt.name, t, func() {
			healthy, unhealthy, unknown := countHealthStatus(tt.chips, tt.caches)
			convey.So(healthy, convey.ShouldEqual, tt.expectHealthy)
			convey.So(unhealthy, convey.ShouldEqual, tt.expectUnhealthy)
			convey.So(unknown, convey.ShouldEqual, tt.expectUnknown)
		})
	}
}

// replayCollector re-emits pre-built metrics into a registry so they can be gathered by name.
type replayCollector struct {
	metrics []prometheus.Metric
}

func (c *replayCollector) Describe(_ chan<- *prometheus.Desc) {}

func (c *replayCollector) Collect(out chan<- prometheus.Metric) {
	for _, m := range c.metrics {
		out <- m
	}
}

// gatherGaugeValues drains ch and returns gauge values keyed by metric name.
func gatherGaugeValues(ch chan prometheus.Metric) map[string]float64 {
	metrics := make([]prometheus.Metric, 0)
	for m := range ch {
		metrics = append(metrics, m)
	}
	registry := prometheus.NewRegistry()
	registry.MustRegister(&replayCollector{metrics: metrics})
	families, err := registry.Gather()
	if err != nil {
		return map[string]float64{}
	}
	values := make(map[string]float64)
	for _, family := range families {
		for _, metric := range family.GetMetric() {
			if gauge := metric.GetGauge(); gauge != nil {
				values[family.GetName()] = gauge.GetValue()
			}
		}
	}
	return values
}

func mockChipCacheWithHealth(n *colcommon.NpuCollector, c *BaseInfoCollector, chips []colcommon.HuaWeiAIChip,
	healthByPhyID map[int32]string) {
	localCache := sync.Map{}
	for _, chip := range chips {
		status := healthByPhyID[chip.PhyId]
		if status == "" {
			status = colcommon.Healthy
		}
		localCache.Store(chip.PhyId, chipCache{chip: chip, timestamp: time.Now(), HealthStatus: status,
			ErrorCodes: []int64{0}, Temperature: 0, Power: 0, Voltage: 0, AICoreCurrentFreq: 0,
			NetHealthStatus: colcommon.Healthy, DevProcessInfo: mockProcessInfo()})
	}
	colcommon.UpdateCache[chipCache](n, colcommon.GetCacheKey(c), &localCache)
}

func TestUpdatePrometheusMachineHealthMetrics(t *testing.T) {
	for _, tt := range []struct {
		name            string
		healthByPhyID   map[int32]string
		expectHealthy   float64
		expectUnhealthy float64
		expectUnknown   float64
	}{
		{
			name:            "should report all healthy chips",
			healthByPhyID:   map[int32]string{},
			expectHealthy:   float64(maxChipNum),
			expectUnhealthy: 0,
			expectUnknown:   0,
		},
		{
			name: "should report mixed health statuses",
			healthByPhyID: map[int32]string{
				0: colcommon.Healthy,
				1: colcommon.UnHealthy,
				2: colcommon.UnHealthy,
				3: colcommon.Unknown,
			},
			expectHealthy:   5,
			expectUnhealthy: 2,
			expectUnknown:   1,
		},
	} {
		convey.Convey(tt.name, t, func() {
			n := mockNewNpuCollector()
			c := &BaseInfoCollector{}
			chips := mockGetNPUChipList()
			mockChipCacheWithHealth(n, c, chips, tt.healthByPhyID)

			ch := make(chan prometheus.Metric, maxMetricsCount)
			go func() {
				defer close(ch)
				c.UpdatePrometheus(ch, n, mockGetContainerNPUInfo(), chips)
			}()

			values := gatherGaugeValues(ch)
			convey.So(values["machine_npu_nums"], convey.ShouldEqual, float64(len(chips)))
			convey.So(values["machine_healthy_npu_nums"], convey.ShouldEqual, tt.expectHealthy)
			convey.So(values["machine_unhealthy_npu_nums"], convey.ShouldEqual, tt.expectUnhealthy)
			convey.So(values["machine_unknown_npu_nums"], convey.ShouldEqual, tt.expectUnknown)
		})
	}
}
