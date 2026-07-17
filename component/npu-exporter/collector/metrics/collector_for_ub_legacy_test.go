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

const (
	ubRxLegacyMetricCount   = 21
	ubTxLegacyMetricCount   = 16
	ubSumLegacyMetricCount  = 3
	ubUboeLegacyMetricCount = 8
	num6                    = 6
)

// TestBuildLegacyDescMap tests the legacy descriptor map builder
func TestBuildLegacyDescMap(t *testing.T) {
	convey.Convey("TestBuildLegacyDescMap", t, func() {
		portMap := map[int][]common.NpuDevPortInfo{
			0: {{PortID: 0}, {PortID: 1}},
			1: {{PortID: 0}},
		}
		patches := gomonkey.ApplyFunc(colcommon.NpuDevPortInfos.GetPortMap, func() map[int][]common.NpuDevPortInfo {
			return portMap
		})
		defer patches.Reset()

		result := buildLegacyDescMap("test_base", "test help")
		convey.So(len(result), convey.ShouldEqual, num6)
	})
}

// TestBuildLegacyDescSlice tests the legacy descriptor slice builder
func TestBuildLegacyDescSlice(t *testing.T) {
	convey.Convey("TestBuildLegacyDescSlice", t, func() {
		portMap := map[int][]common.NpuDevPortInfo{
			0: {{PortID: 0}, {PortID: 1}},
			1: {{PortID: 0}},
		}
		patches := gomonkey.ApplyFunc(colcommon.NpuDevPortInfos.GetPortMap, func() map[int][]common.NpuDevPortInfo {
			return portMap
		})
		defer patches.Reset()

		result := buildLegacyDescSlice("test_base", "test help")
		convey.So(len(result), convey.ShouldEqual, num6)
	})
}

// TestTryEmitUbLegacyMetric tests the legacy metric emission helper
func TestTryEmitUbLegacyMetric(t *testing.T) {
	convey.Convey("TestTryEmitUbLegacyMetric", t, func() {
		ch := make(chan prometheus.Metric, 10)
		timestamp := time.Now()
		cardLabel := []string{"card0"}

		convey.Convey("When key does not exist in map, no metric emitted", func() {
			callCount := 0
			patches := gomonkey.ApplyFunc(doUpdateMetric, func(ch chan<- prometheus.Metric,
				ts time.Time, val interface{}, labels []string, desc *prometheus.Desc) {
				callCount++
			})
			defer patches.Reset()

			tryEmitUbLegacyMetric(ch, timestamp, 1, cardLabel, "non_existent_key", 0, 0)
			convey.So(callCount, convey.ShouldEqual, 0)
		})

		convey.Convey("When port key does not exist, no metric emitted", func() {
			ubLegacyDescMap["test_key"] = map[string]*prometheus.Desc{
				"0_1": colcommon.BuildDescWithLabel("test", "help", colcommon.CardLabel),
			}

			callCount := 0
			patches := gomonkey.ApplyFunc(doUpdateMetric, func(ch chan<- prometheus.Metric,
				ts time.Time, val interface{}, labels []string, desc *prometheus.Desc) {
				callCount++
			})
			defer patches.Reset()

			tryEmitUbLegacyMetric(ch, timestamp, 1, cardLabel, "test_key", 0, 0)
			convey.So(callCount, convey.ShouldEqual, 0)
		})
	})
}

// TestPromUpdateUbRxLegacy tests legacy format emission for UB rx metrics
func TestPromUpdateUbRxLegacy(t *testing.T) {
	convey.Convey("TestPromUpdateUbRxLegacy", t, func() {
		ch := make(chan prometheus.Metric, 100)
		timestamp := time.Now()
		cardLabel := []string{"card0"}
		extendedLabel := append(cardLabel, "0", "1") // cardLabel + dieID + portID
		ubInfo := &common.UBInfo{
			UBCommonStats: initUBCommonStats(),
			Udie:          0,
			Port:          1,
		}

		initUbLegacyDesc()

		convey.Convey("When EnableLegacyMetrics is false, no metrics emitted", func() {
			colcommon.EnableLegacyMetrics = false
			callCount := 0
			patches := gomonkey.ApplyFunc(doUpdateMetric, func(ch chan<- prometheus.Metric,
				ts time.Time, val interface{}, labels []string, desc *prometheus.Desc) {
				callCount++
			})
			defer patches.Reset()

			promUpdateUbRxLegacy(ch, timestamp, ubInfo, extendedLabel)
			convey.So(callCount, convey.ShouldEqual, 0)
		})

		convey.Convey("When EnableLegacyMetrics is true, correct number of metrics emitted", func() {
			colcommon.EnableLegacyMetrics = true
			callCount := 0
			lastLabels := []string{}
			patches := gomonkey.ApplyFunc(doUpdateMetric, func(ch chan<- prometheus.Metric,
				ts time.Time, val interface{}, labels []string, desc *prometheus.Desc) {
				callCount++
				lastLabels = labels
			})
			defer patches.Reset()

			promUpdateUbRxLegacy(ch, timestamp, ubInfo, extendedLabel)
			convey.So(callCount, convey.ShouldEqual, ubRxLegacyMetricCount)
			// legacy format should NOT include udie/port labels
			convey.So(len(lastLabels), convey.ShouldEqual, len(cardLabel))
		})
	})
}

// TestPromUpdateUbTxLegacy tests legacy format emission for UB tx metrics
func TestPromUpdateUbTxLegacy(t *testing.T) {
	convey.Convey("TestPromUpdateUbTxLegacy", t, func() {
		ch := make(chan prometheus.Metric, 100)
		timestamp := time.Now()
		cardLabel := []string{"card0"}
		extendedLabel := append(cardLabel, "0", "1") // cardLabel + dieID + portID
		ubInfo := &common.UBInfo{
			UBCommonStats: initUBCommonStats(),
			Udie:          0,
			Port:          1,
		}

		initUbLegacyDesc()

		convey.Convey("When EnableLegacyMetrics is true, correct number of metrics emitted", func() {
			colcommon.EnableLegacyMetrics = true
			callCount := 0
			patches := gomonkey.ApplyFunc(doUpdateMetric, func(ch chan<- prometheus.Metric,
				ts time.Time, val interface{}, labels []string, desc *prometheus.Desc) {
				callCount++
			})
			defer patches.Reset()

			promUpdateUbTxLegacy(ch, timestamp, ubInfo, extendedLabel)
			convey.So(callCount, convey.ShouldEqual, ubTxLegacyMetricCount)
		})
	})
}

// TestPromUpdateUbSumLegacy tests legacy format emission for UB sum metrics
func TestPromUpdateUbSumLegacy(t *testing.T) {
	convey.Convey("TestPromUpdateUbSumLegacy", t, func() {
		ch := make(chan prometheus.Metric, 100)
		timestamp := time.Now()
		cardLabel := []string{"card0"}
		extendedLabel := append(cardLabel, "0", "1") // cardLabel + dieID + portID
		ubInfo := &common.UBInfo{
			UBCommonStats: initUBCommonStats(),
			Udie:          0,
			Port:          1,
		}

		initUbLegacyDesc()

		convey.Convey("When EnableLegacyMetrics is true, correct number of metrics emitted", func() {
			colcommon.EnableLegacyMetrics = true
			callCount := 0
			patches := gomonkey.ApplyFunc(doUpdateMetric, func(ch chan<- prometheus.Metric,
				ts time.Time, val interface{}, labels []string, desc *prometheus.Desc) {
				callCount++
			})
			defer patches.Reset()

			promUpdateUbSumLegacy(ch, timestamp, ubInfo, extendedLabel)
			convey.So(callCount, convey.ShouldEqual, ubSumLegacyMetricCount)
		})
	})
}

// TestPromUpdateUbUboeLegacy tests legacy format emission for UB UBOE metrics
func TestPromUpdateUbUboeLegacy(t *testing.T) {
	convey.Convey("TestPromUpdateUbUboeLegacy", t, func() {
		ch := make(chan prometheus.Metric, 100)
		timestamp := time.Now()
		cardLabel := []string{"card0"}
		extendedLabel := append(cardLabel, "0", "1") // cardLabel + dieID + portID
		ubInfo := &common.UBInfo{
			UBCommonStats: initUBCommonStats(),
			Udie:          0,
			Port:          1,
			UboeExtensions: &common.UBOEExtensions{
				CoreMibRxPausePkts: 1,
				CoreMibTxPausePkts: 1,
				CoreMibRxPfcPkts:   1,
				CoreMibTxPfcPkts:   1,
				CoreMibRxBadPkts:   1,
				CoreMibTxBadPkts:   1,
				CoreMibRxBadOctets: 1,
				CoreMibTxBadOctets: 1,
			},
		}

		initUbLegacyDesc()

		convey.Convey("When EnableLegacyMetrics is true, correct number of metrics emitted", func() {
			colcommon.EnableLegacyMetrics = true
			callCount := 0
			patches := gomonkey.ApplyFunc(doUpdateMetric, func(ch chan<- prometheus.Metric,
				ts time.Time, val interface{}, labels []string, desc *prometheus.Desc) {
				callCount++
			})
			defer patches.Reset()

			promUpdateUbUboeLegacy(ch, timestamp, ubInfo, extendedLabel)
			convey.So(callCount, convey.ShouldEqual, ubUboeLegacyMetricCount)
		})
	})
}

// TestAddUbLegacyMetricsDesc tests adding legacy descriptors to channel
func TestAddUbLegacyMetricsDesc(t *testing.T) {
	convey.Convey("TestAddUbLegacyMetricsDesc", t, func() {
		// Mock port map to have predictable number of descriptors
		portMap := map[int][]common.NpuDevPortInfo{
			0: {{PortID: 0}, {PortID: 1}},
			1: {{PortID: 0}},
		}
		patches := gomonkey.ApplyFunc(colcommon.NpuDevPortInfos.GetPortMap, func() map[int][]common.NpuDevPortInfo {
			return portMap
		})
		defer patches.Reset()

		// Re-initialize with mocked port map
		ubLegacyDescMap = make(map[string]map[string]*prometheus.Desc)
		initUbLegacyDesc()

		convey.Convey("When EnableLegacyMetrics is false, no descriptors sent", func() {
			// Use buffered channel with small capacity
			ch := make(chan *prometheus.Desc, 1)
			colcommon.EnableLegacyMetrics = false
			addUbLegacyMetricsDesc(ch)
			// Verify no descriptors were sent (channel should be empty)
			convey.So(len(ch), convey.ShouldEqual, 0)
		})
	})
}
