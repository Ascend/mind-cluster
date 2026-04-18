/* Copyright(C) 2025-2025. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/hccn"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
)

const (
	opticalMetricsNum = 9
)

func init() {
	colcommon.NpuDevPortInfos.SetPortMap(mockPorts)
	colcommon.NpuDevPortInfos.Init()
}

// TestOpticalCollectorIsSupported test OpticalCollector IsSupported
func TestOpticalCollectorIsSupported(t *testing.T) {
	n := mockNewNpuCollector()
	cases := []testCase{
		buildTestCase("NetworkCollector: testIsSupported on Ascend910A3", &OpticalCollector{}, api.Ascend910A3, true),
		buildTestCase("NetworkCollector: testIsSupported on Ascend910A5", &OpticalCollector{}, api.Ascend910A5, true),
	}

	for _, c := range cases {
		patches := gomonkey.NewPatches()
		convey.Convey(c.name, t, func() {
			defer patches.Reset()
			patches.ApplyMethodReturn(n.Dmgr, "GetMainBoardId", uint32(api.Atlas850MainBoardID2))
			patches.ApplyMethodReturn(n.Dmgr, "GetDevType", c.deviceType)
			patches.ApplyMethodReturn(n.Dmgr, "IsTrainingCard", true)
			isSupported := c.collectorType.IsSupported(n)
			convey.So(isSupported, convey.ShouldEqual, c.expectValue)
		})
	}
}

// TestOpticalCollectorIsSupported2 test OpticalCollector IsSupported
func TestOpticalCollectorIsSupported2(t *testing.T) {
	n := mockNewNpuCollector()
	cases := []testCase{
		buildTestCase("NetworkCollector: testIsSupported on Ascend910A3", &OpticalCollector{}, api.Ascend910A3, true),
		buildTestCase("NetworkCollector: testIsSupported on Ascend910A5", &OpticalCollector{}, api.Ascend910A5, false),
	}

	for _, c := range cases {
		colcommon.DevType = c.deviceType
		patches := gomonkey.NewPatches()
		convey.Convey(c.name, t, func() {
			defer patches.Reset()
			patches.ApplyMethodReturn(n.Dmgr, "GetMainBoardId", uint32(api.Atlas9501DMainBoardID))
			patches.ApplyMethodReturn(n.Dmgr, "IsTrainingCard", true)
			isSupported := c.collectorType.IsSupported(n)
			convey.So(isSupported, convey.ShouldEqual, c.expectValue)
		})
	}
}

func mockOpticalInfoNpu() []*common.OpticalNpuInfo {
	var newOpticalInfos []*common.OpticalNpuInfo
	for i := 0; i < colcommon.NpuDevPortInfos.GetCount(); i++ {
		opticalInfos := common.OpticalNpuInfo{
			OpticalIndex:    num2,
			OpticalTxPower0: 1.0,
			OpticalTxPower1: 1.0,
			OpticalTxPower2: 1.0,
			OpticalTxPower3: 1.0,
			OpticalRxPower0: 1.0,
			OpticalRxPower1: 1.0,
			OpticalRxPower2: 1.0,
			OpticalRxPower3: 1.0,
		}
		newOpticalInfos = append(newOpticalInfos, &opticalInfos)
	}
	return newOpticalInfos
}

func mockOpticalNpuCache(n *colcommon.NpuCollector, chips []colcommon.HuaWeiAIChip, cacheKey string) {
	localCache := sync.Map{}
	for _, chip := range chips {
		localCache.Store(chip.PhyId, opticalNpuCache{chip: chip, timestamp: time.Now(),
			extInfo: mockOpticalInfoNpu()})
	}
	colcommon.UpdateCache[opticalNpuCache](n, cacheKey, &localCache)
}

// TestUpdateTelegrafOpticalNpu test update telegraf info
func TestUpdateTelegrafOpticalNpu(t *testing.T) {
	n := mockNewNpuCollector()
	convey.Convey("TestUpdateTelegrafOpticalNpu", t, func() {
		colcommon.DevType = api.Ascend910A5
		initNpuOpticalDesc()
		containerInfos := mockGetContainerNPUInfo()
		chips := mockGetNPUChipList()
		c := OpticalCollector{}
		mockOpticalNpuCache(n, chips, colcommon.GetCacheKey(&c))
		fieldsMap := make(map[string]map[string]interface{})
		result := c.UpdateTelegraf(fieldsMap, n, containerInfos, chips)
		convey.So(result, convey.ShouldNotBeEmpty)
	})
}

// TestCollectToCacheOpticalNpu test CollectToCache
func TestCollectToCacheOpticalNpu(t *testing.T) {
	n := mockNewNpuCollector()
	convey.Convey("TestCollectToCacheOpticalNpu", t, func() {
		colcommon.DevType = api.Ascend910A5
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFuncReturn(collectOpticalNpuInfo, mockOpticalInfoNpu())
		chips := mockGetNPUChipList()
		c := OpticalCollector{}
		c.CollectToCache(n, chips)
		convey.So(colcommon.GetInfoFromCache[opticalNpuCache](n, colcommon.GetCacheKey(&c)),
			convey.ShouldNotBeEmpty)
	})
}

// TestCollectOpticalNpuInfo test collectOpticalNpuInfo function
func TestCollectOpticalNpuInfo(t *testing.T) {
	convey.Convey("TestCollectOpticalNpuInfo", t, func() {
		// Mock the dependencies
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		// Mock GetNpuOpticalInfoNpu to return success
		testOpticalInfo := map[string]string{
			txNpuPower0:  "1.0",
			rxNpuPower0:  "5.0",
			opticalIndex: "1",
		}
		patches.ApplyFuncReturn(hccn.GetNpuOpticalInfoNpu, testOpticalInfo, nil)

		// Mock storeOpticalNpuInfos to return a valid OpticalNpuInfo
		testOpticalNpuInfo := &common.OpticalNpuInfo{
			OpticalIndex:    1,
			OpticalTxPower0: 1.0,
			OpticalRxPower0: 5.0,
		}
		patches.ApplyFuncReturn(storeOpticalNpuInfos, testOpticalNpuInfo)

		// Call the function under test
		result := collectOpticalNpuInfo(0)

		// Basic verification
		convey.So(result, convey.ShouldNotBeEmpty)
		convey.So(len(result), convey.ShouldEqual, colcommon.NpuDevPortInfos.GetCount())
		convey.So(result[0], convey.ShouldNotBeNil)
		convey.So(result[0].OpticalIndex, convey.ShouldEqual, 1)
	})
}

// TestPromUpdateOpticalInfo test update prometheus update inf
func TestPromUpdateOpticalInfo(t *testing.T) {
	convey.Convey("TestPromUpdateOpticalInfo", t, func() {
		ch := make(chan prometheus.Metric, 100)
		timestamp := time.Now()
		cardLabel := []string{"card0"}
		patches := gomonkey.ApplyFunc(validateNotNilForEveryElement, func(objs ...interface{}) bool {
			return objs != nil
		})
		defer patches.Reset()

		convey.Convey("When cache extInfo is nil", func() {
			cache := opticalNpuCache{extInfo: nil}
			promUpdateOpticalInfo(ch, cache, timestamp, cardLabel)
			convey.So(len(ch), convey.ShouldEqual, 0)
		})

		convey.Convey("When cache has valid data", func() {
			cache := opticalNpuCache{extInfo: mockOpticalInfoNpu()}
			callCount := 0
			callTelCount := 0
			mockFiledMap := make(map[string]interface{})
			patches.ApplyFunc(doUpdateMetric, func(ch chan<- prometheus.Metric, timestamp time.Time, value interface{}, cardLabel []string, desc *prometheus.Desc) {
				callCount++
			})
			patches.ApplyFunc(doUpdateMetricWithValidateNum, func(ch chan<- prometheus.Metric, ts time.Time, val float64, labels []string, desc *prometheus.Desc) {
				callCount++
			})
			patches.ApplyFunc(doUpdateTelegrafWithValidateNum, func(fieldMap map[string]interface{}, desc *prometheus.Desc, value float64, extInfo string) {
				callTelCount++
			})
			patches.ApplyFunc(doUpdateTelegraf, func(fieldMap map[string]interface{}, desc *prometheus.Desc, value interface{}, extInfo string) {
				callTelCount++
			})

			promUpdateOpticalInfo(ch, cache, timestamp, cardLabel)
			telegrafUpdateOpticalInfo(cache, mockFiledMap)
			expectedCalls := colcommon.NpuDevPortInfos.GetCount() * opticalMetricsNum
			convey.So(callCount, convey.ShouldEqual, expectedCalls)
			convey.So(callTelCount, convey.ShouldEqual, expectedCalls)
		})
	})
}

// TestStoreOpticalNpuInfos test storeOpticalNpuInfos function
func TestStoreOpticalNpuInfos(t *testing.T) {
	convey.Convey("TestStoreOpticalNpuInfos", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		// Test data
		testInfoMap := map[string]string{
			txNpuPower0:  "10.0",
			txNpuPower1:  "20.0",
			rxNpuPower0:  "50.0",
			opticalIndex: "5",
		}

		// Simple mock that returns expected values
		patches.ApplyFunc(storeSingleOpticalNpuInfo, func(str string, logicID int32, uDie, port int, convertType string) interface{} {
			if convertType == "int" {
				return 5
			}
			return 10.0 + float64(uDie*10+port*5)
		})

		// Call and verify
		result := storeOpticalNpuInfos(testInfoMap, 0, 0, 0)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result.OpticalIndex, convey.ShouldEqual, 5)
		convey.So(result.OpticalTxPower0, convey.ShouldNotEqual, 0)
		convey.So(result.OpticalRxPower0, convey.ShouldNotEqual, 0)
	})
}

// TestStoreSingleOpticalNpuInfo test storeSingleOpticalNpuInfo function
func TestStoreSingleOpticalNpuInfo(t *testing.T) {
	convey.Convey("TestStoreSingleOpticalNpuInfo", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		// Mock logging
		patches.ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {})

		convey.Convey("Test int conversion success", func() {
			patches.ApplyFuncReturn(hccn.GetIntDataFromStrNpu, 10, nil)
			result := storeSingleOpticalNpuInfo("10", 0, 0, 0, "int")
			convey.So(result, convey.ShouldEqual, 10)
		})

		convey.Convey("Test float conversion success", func() {
			patches.ApplyFuncReturn(hccn.GetFloatDataFromStrNpu, 10.5, nil)
			result := storeSingleOpticalNpuInfo("10.5", 0, 0, 0, "float")
			convey.So(result, convey.ShouldEqual, 10.5)
		})

		convey.Convey("Test int conversion failure", func() {
			patches.ApplyFuncReturn(hccn.GetIntDataFromStrNpu, 0, fmt.Errorf("parse error"))
			result := storeSingleOpticalNpuInfo("invalid", 0, 0, 0, "int")
			convey.So(result, convey.ShouldEqual, 0)
		})

		convey.Convey("Test float conversion failure", func() {
			patches.ApplyFuncReturn(hccn.GetFloatDataFromStrNpu, 0.0, fmt.Errorf("parse error"))
			result := storeSingleOpticalNpuInfo("invalid", 0, 0, 0, "float")
			convey.So(result, convey.ShouldEqual, 0.0)
		})

		convey.Convey("Test invalid convert type", func() {
			result := storeSingleOpticalNpuInfo("10", 0, 0, 0, "invalid")
			convey.So(result, convey.ShouldEqual, common.RetError)
		})
	})
}
