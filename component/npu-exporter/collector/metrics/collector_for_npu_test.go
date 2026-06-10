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
