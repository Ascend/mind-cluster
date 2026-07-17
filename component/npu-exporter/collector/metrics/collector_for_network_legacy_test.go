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
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/devmanager/common"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
)

// TestPromUpdateNetInfoLegacy tests legacy format emission for network metrics
func TestPromUpdateNetInfoLegacy(t *testing.T) {
	convey.Convey("TestPromUpdateNetInfoLegacy", t, func() {
		ch := make(chan prometheus.Metric, 100)
		timestamp := time.Now()
		cardLabel := []string{"card0"}
		extendedLabel := append(cardLabel, "0", "1")

		netInfo := &common.NpuNetInfo{
			LinkStatusInfo: &common.LinkStatusInfo{LinkState: "UP"},
			BandwidthInfo:  &common.BandwidthInfo{TxValue: 100, RxValue: 200},
			LinkSpeedInfo:  &common.LinkSpeedInfo{Speed: 400},
		}

		convey.Convey("When EnableLegacyMetrics is false, no metrics emitted", func() {
			colcommon.EnableLegacyMetrics = false
			initNetworkLegacyDesc()
			callCount := 0
			patches := gomonkey.ApplyFunc(doUpdateMetricWithValidateNum, func(ch chan<- prometheus.Metric,
				ts time.Time, val float64, labels []string, desc *prometheus.Desc) {
				callCount++
			})
			defer patches.Reset()

			promUpdateNetInfoLegacy(ch, timestamp, netInfo, extendedLabel, 0)
			convey.So(callCount, convey.ShouldEqual, 0)
		})

		convey.Convey("When EnableLegacyMetrics is true, correct number of metrics emitted", func() {
			colcommon.EnableLegacyMetrics = true
			initNetworkLegacyDesc()
			callCount := 0
			lastLabels := []string{}
			patches := gomonkey.ApplyFunc(doUpdateMetricWithValidateNum, func(ch chan<- prometheus.Metric,
				ts time.Time, val float64, labels []string, desc *prometheus.Desc) {
				callCount++
				lastLabels = labels
			})
			defer patches.Reset()

			promUpdateNetInfoLegacy(ch, timestamp, netInfo, extendedLabel, 0)
			convey.So(callCount, convey.ShouldEqual, ascend950NetworkMetricNum)
			// legacy format should NOT include udie/port labels
			convey.So(len(lastLabels), convey.ShouldEqual, len(cardLabel))
		})

		convey.Convey("When netInfo has nil fields, skips gracefully", func() {
			colcommon.EnableLegacyMetrics = true
			initNetworkLegacyDesc()
			nilNetInfo := &common.NpuNetInfo{
				LinkStatusInfo: nil,
				BandwidthInfo:  nil,
				LinkSpeedInfo:  nil,
			}
			callCount := 0
			patches := gomonkey.ApplyFunc(doUpdateMetricWithValidateNum, func(ch chan<- prometheus.Metric,
				ts time.Time, val float64, labels []string, desc *prometheus.Desc) {
				callCount++
			})
			defer patches.Reset()

			promUpdateNetInfoLegacy(ch, timestamp, nilNetInfo, extendedLabel, 0)
			convey.So(callCount, convey.ShouldEqual, 0)
		})
	})
}
