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
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/hccn"
)

const ascend950NetworkMetricNum = 4

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// TestCollectNetworkNpuInfo test collectNetworkNpuInfo function
func TestCollectNetworkNpuInfo(t *testing.T) {
	convey.Convey("TestCollectNetworkNpuInfo", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		// Mock hccn functions to return success
		patches.ApplyFuncReturn(hccn.GetNPULinkStatusNpu, "up", nil)
		patches.ApplyFuncReturn(hccn.GetNPUInterfaceTrafficNpu, 100.0, 200.0, nil)
		patches.ApplyFuncReturn(hccn.GetNPULinkSpeedNpu, 1000, nil)

		// Call the function under test
		result := collectNetworkNpuInfo(0)

		// Basic verification
		convey.So(result, convey.ShouldNotBeEmpty)
		convey.So(len(result), convey.ShouldEqual, maxDieId*maxPortId)
		convey.So(result[0].LinkStatusInfo.LinkState, convey.ShouldEqual, "up")
		convey.So(result[0].BandwidthInfo.TxValue, convey.ShouldEqual, 100.0)
		convey.So(result[0].BandwidthInfo.RxValue, convey.ShouldEqual, 200.0)
		convey.So(result[0].LinkSpeedInfo.Speed, convey.ShouldEqual, 1000.0)
	})
}

// TestNetworkCollectorIsSupported test NetworkCollector IsSupported
func TestNetworkCollectorIsSupported(t *testing.T) {
	n := mockNewNpuCollector()
	cases := []testCase{
		buildTestCase("NetworkCollector: testIsSupported on Npu", &NetworkCollector{}, api.Ascend910A5, true),
		buildTestCase("NetworkCollector: testIsSupported on Ascend910A3", &NetworkCollector{}, api.Ascend910A3, true),
	}

	for _, c := range cases {
		patches := gomonkey.NewPatches()
		convey.Convey(c.name, t, func() {
			defer patches.Reset()
			patches.ApplyMethodReturn(n.Dmgr, "GetMainBoardId", uint32(api.Atlas9501DMainBoardID))
			patches.ApplyMethodReturn(n.Dmgr, "IsTrainingCard", true)
			patches.ApplyMethodReturn(n.Dmgr, "GetDevType", c.deviceType)
			isSupported := c.collectorType.IsSupported(n)
			convey.So(isSupported, convey.ShouldEqual, c.expectValue)
		})
	}
}

// TestNetworkCollectorIsSupported2 test NetworkCollector IsSupported
func TestNetworkCollectorIsSupported2(t *testing.T) {
	n := mockNewNpuCollector()
	cases := []testCase{
		buildTestCase("NetworkCollector: testIsSupported on Npu: false", &NetworkCollector{}, api.Ascend910A5, false),
	}

	for _, c := range cases {
		patches := gomonkey.NewPatches()
		convey.Convey(c.name, t, func() {
			defer patches.Reset()
			patches.ApplyMethodReturn(n.Dmgr, "GetMainBoardId", uint32(api.Atlas3501PMainBoardID))
			patches.ApplyMethodReturn(n.Dmgr, "IsTrainingCard", true)
			patches.ApplyMethodReturn(n.Dmgr, "GetDevType", c.deviceType)
			isSupported := c.collectorType.IsSupported(n)
			convey.So(isSupported, convey.ShouldEqual, c.expectValue)
		})
	}
}

// TestPromUpdateNetInfo test update prometheus update info
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
			cache := netInfoNPUCache{extInfo: nil}
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
			cache := netInfoNPUCache{extInfo: mockExtInfo}

			callCount := 0
			callTelCount := 0
			mockFiledMap := make(map[string]interface{})
			patches.ApplyFunc(doUpdateMetricWithValidateNum, func(ch chan<- prometheus.Metric, ts time.Time, val float64, labels []string, desc *prometheus.Desc) {
				callCount++
			})
			patches.ApplyFunc(doUpdateTelegrafWithValidateNum, func(fieldMap map[string]interface{}, desc *prometheus.Desc, value float64, extInfo string) {
				callTelCount++
			})

			promUpdateNetInfo(ch, cache, timestamp, cardLabel)
			telegrafUpdateNetInfo(cache, mockFiledMap)
			expectedCalls := (maxDieId * maxPortId) * ascend950NetworkMetricNum
			convey.So(callCount, convey.ShouldEqual, expectedCalls)
			convey.So(callTelCount, convey.ShouldEqual, expectedCalls)
		})
	})
}
