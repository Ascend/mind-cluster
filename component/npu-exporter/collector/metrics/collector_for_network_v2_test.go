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
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/devmanager/common"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
)

const (
	ascend950NetworkMetricNum = 4
)

// TestNetworkCollectorNPUIsSupported test NetworkCollectorNPU IsSupported
func TestNetworkCollectorA5IsSupported(t *testing.T) {
	n := mockNewNpuCollector()
	cases := []testCase{
		buildTestCase("NetworkCollector: testIsSupported on Ascend910A3", &NetworkNPUCollector{}, api.Ascend910A3, false),
		buildTestCase("NetworkCollector: testIsSupported on Ascend910A5", &NetworkNPUCollector{}, api.Ascend910A5, true),
	}

	for _, c := range cases {
		patches := gomonkey.NewPatches()
		convey.Convey(c.name, t, func() {
			defer patches.Reset()
			patches.ApplyMethodReturn(n.Dmgr, "GetDevType", c.deviceType)
			isSupported := c.collectorType.IsSupported(n)
			convey.So(isSupported, convey.ShouldEqual, c.expectValue)
		})
	}
}

func TestNetworkCollectorA5CollectToCache(t *testing.T) {
	collector := &NetworkNPUCollector{}
	n := mockNewNpuCollector()
	testChips := []colcommon.HuaWeiAIChip{{PhyId: 0}}

	convey.Convey("TestNetworkCollectorNPUCollectToCache", t, func() {
		collector.CollectToCache(n, testChips)
		cacheInfo := colcommon.GetInfoFromCache[chipCache](n, colcommon.GetCacheKey(collector))
		convey.So(cacheInfo, convey.ShouldNotBeNil)
	})
}

func TestPromUpdateNetInfo(t *testing.T) {
	convey.Convey("Given promUpdateNetInfo function", t, func() {
		ch := make(chan prometheus.Metric, 100)
		timestamp := time.Now()
		cardLabel := []string{"card0"}
		patches := gomonkey.ApplyFunc(validateNotNilForEveryElement, func(objs ...interface{}) bool {
			return objs != nil
		})
		defer patches.Reset()

		convey.Convey("When cache extInfo is nil", func() {
			cache := netInfoA5Cache{extInfo: nil}
			promUpdateNetInfo(ch, cache, timestamp, cardLabel)
			convey.So(len(ch), convey.ShouldEqual, 0)
		})

		convey.Convey("When cache has valid data", func() {
			mockExtInfo := make([]*common.NpuNetInfo, maxDieId*maxPortId)
			for i := 0; i < (maxDieId * maxPortId); i++ {
				mockExtInfo[i] = &common.NpuNetInfo{
					LinkStatusInfo: &common.LinkStatusInfo{LinkState: "UP"},
					BandwidthInfo:  &common.BandwidthInfo{TxValue: 100, RxValue: 200},
					LinkSpeedInfo:  &common.LinkSpeedInfo{Speed: 400},
				}
			}
			cache := netInfoA5Cache{extInfo: mockExtInfo}

			callCount := 0
			patches.ApplyFunc(doUpdateMetricWithValidateNum, func(ch chan<- prometheus.Metric, ts time.Time, val float64, labels []string, desc *prometheus.Desc) {
				callCount++
			})

			promUpdateNetInfo(ch, cache, timestamp, cardLabel)
			expectedCalls := (maxDieId * maxPortId) * ascend950NetworkMetricNum
			convey.So(callCount, convey.ShouldEqual, expectedCalls)
		})
	})
}
