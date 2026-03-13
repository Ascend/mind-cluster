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

	"ascend-common/api"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/hccn"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
)

// TestDescribeOpticalNpu test Describe
func TestDescribeOpticalNpu(t *testing.T) {
	convey.Convey("test prometheus desc ", t, func() {
		ch := make(chan *prometheus.Desc, maxMetricsCount)
		c := OpticalNpuCollector{}
		c.Describe(ch)
		t.Logf("Describe len(ch):%v", len(ch))
		convey.So(ch, convey.ShouldNotBeEmpty)
	})
}

// TestIsSupportedOptical tests the IsSupported method of OpticalNpuCollector
func TestIsSupportedOptical(t *testing.T) {
	n := mockNewNpuCollector()
	cases := []testCase{
		buildTestCase("UBCollector: testIsSupported on Ascend910A3", &OpticalNpuCollector{}, api.Ascend910A3, false),
		buildTestCase("UBCollector: testIsSupported on Ascend910A5", &OpticalNpuCollector{}, api.Ascend910A5, true),
	}

	for _, c := range cases {
		patches := gomonkey.NewPatches()
		convey.Convey(c.name, t, func() {
			defer patches.Reset()
			patches.ApplyMethodReturn(n.Dmgr, "GetMainBoardId", uint32(api.Atlas850MainBoardID2))
			patches.ApplyMethodReturn(n.Dmgr, "GetDevType", c.deviceType)
			isSupported := c.collectorType.IsSupported(n)
			convey.So(isSupported, convey.ShouldEqual, c.expectValue)
		})
	}
}

// TestCollectOpticalNpuInfo tests the collectOpticalNpuInfo function
func TestCollectOpticalNpuInfo(t *testing.T) {
	convey.Convey("Test collectOpticalNpuInfo function", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		mockOpticalInfo := map[string]string{
			txNpuPower0:  "-3.5",
			txNpuPower1:  "-4.2",
			txNpuPower2:  "-3.8",
			txNpuPower3:  "-4.0",
			rxNpuPower0:  "-5.1",
			rxNpuPower1:  "-5.3",
			rxNpuPower2:  "-4.9",
			rxNpuPower3:  "-5.0",
			opticalIndex: "1",
		}

		patches.ApplyFunc(hccn.GetNpuOpticalInfoNpu, func(logicID, udieID, portID int32) (map[string]string, error) {
			return mockOpticalInfo, nil
		})

		patches.ApplyFunc(hccn.GetFloatDataFromStrNpu, func(str string) (float64, error) {
			return 1.0, nil
		})

		patches.ApplyFunc(hccn.GetIntDataFromStrNpu, func(str string) (int, error) {
			return 1, nil
		})

		result := collectOpticalNpuInfo(0)

		convey.So(len(result), convey.ShouldEqual, maxDieId*maxPortId)
		convey.So(result, convey.ShouldNotBeNil)

		for _, info := range result {
			convey.So(info, convey.ShouldNotBeNil)
			convey.So(info.OpticalIndex, convey.ShouldEqual, 1)
			convey.So(info.OpticalTxPower0, convey.ShouldEqual, 1.0)
			convey.So(info.OpticalRxPower0, convey.ShouldEqual, 1.0)
		}
	})
}

func mockOpticalInfoNpu() []*common.OpticalNpuInfo {
	var newOpticalInfos []*common.OpticalNpuInfo
	for i := 0; i < (maxDieId * maxPortId); i++ {
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

// TestCollectToCacheOptical test CollectToCache
func TestCollectToCacheOpticalNpu(t *testing.T) {
	n := mockNewNpuCollector()
	convey.Convey("TestCollectToCacheOptical", t, func() {

		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFuncReturn(collectOpticalNpuInfo, mockOpticalInfoNpu())
		chips := mockGetNPUChipList()
		c := OpticalNpuCollector{}
		c.PreCollect(n, chips)
		c.CollectToCache(n, chips)
		convey.So(colcommon.GetInfoFromCache[opticalNpuCache](n, colcommon.GetCacheKey(&OpticalNpuCollector{})),
			convey.ShouldNotBeEmpty)
	})
}

func mockOpticalNpuCache(n *colcommon.NpuCollector, chips []colcommon.HuaWeiAIChip, cacheKey string) {
	localCache := sync.Map{}
	for _, chip := range chips {
		localCache.Store(chip.PhyId, opticalNpuCache{chip: chip, timestamp: time.Now(),
			extInfo: mockOpticalInfoNpu()})
	}
	colcommon.UpdateCache[opticalNpuCache](n, cacheKey, &localCache)
}

func TestUpdateTelegrafOpticalNpu(t *testing.T) {
	n := mockNewNpuCollector()
	convey.Convey("TestUpdateTelegrafOpticalNpu", t, func() {
		containerInfos := mockGetContainerNPUInfo()
		chips := mockGetNPUChipList()
		c := OpticalNpuCollector{}
		mockOpticalNpuCache(n, chips, colcommon.GetCacheKey(&c))
		fieldsMap := make(map[string]map[string]interface{})
		result := c.UpdateTelegraf(fieldsMap, n, containerInfos, chips)
		convey.So(result, convey.ShouldNotBeEmpty)
	})
}
