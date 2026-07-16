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
// This file contains legacy metrics with _X_Y suffix for Atlas 350 backward compatibility.
// It will be removed after the compatibility period ends.
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"ascend-common/devmanager/common"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
)

var (
	linkStatusLegacyDescs           []*prometheus.Desc
	bandwidthTxLegacyDescs          []*prometheus.Desc
	bandwidthRxLegacyDescs          []*prometheus.Desc
	npuChipPortLinkSpeedLegacyDescs []*prometheus.Desc
)

func initNetworkLegacyDesc() {
	if !colcommon.EnableLegacyMetrics {
		return
	}
	linkStatusLegacyDescs = buildLegacyDescSlice("link_status", "the npu link status on ub port")
	bandwidthTxLegacyDescs = buildLegacyDescSlice("bandwidth_tx", "the npu port transport speed, unit is 'MB/s'")
	bandwidthRxLegacyDescs = buildLegacyDescSlice("bandwidth_rx", "the npu port receive speed, unit is 'MB/s'")
	npuChipPortLinkSpeedLegacyDescs = buildLegacyDescSlice("link_speed", "the npu port link speed, unit is 'G'")
}

func tryEmitNetworkLegacyMetric(ch chan<- prometheus.Metric, timestamp time.Time,
	value float64, extendedLabel []string, legacyDescs []*prometheus.Desc, i int) {
	if i < len(legacyDescs) {
		doUpdateMetricWithValidateNum(ch, timestamp, value,
			extendedLabel[:len(extendedLabel)-2], legacyDescs[i])
	}
}

func promUpdateNetInfoLegacy(ch chan<- prometheus.Metric, timestamp time.Time,
	netInfo *common.NpuNetInfo, extendedLabel []string, i int) {
	if !colcommon.EnableLegacyMetrics {
		return
	}
	if validateNotNilForEveryElement(netInfo.LinkStatusInfo) {
		tryEmitNetworkLegacyMetric(ch, timestamp,
			float64(getLinkStatusCode(netInfo.LinkStatusInfo.LinkState)),
			extendedLabel, linkStatusLegacyDescs, i)
	}
	if validateNotNilForEveryElement(netInfo.BandwidthInfo) {
		tryEmitNetworkLegacyMetric(ch, timestamp, netInfo.BandwidthInfo.TxValue,
			extendedLabel, bandwidthTxLegacyDescs, i)
		tryEmitNetworkLegacyMetric(ch, timestamp, netInfo.BandwidthInfo.RxValue,
			extendedLabel, bandwidthRxLegacyDescs, i)
	}
	if validateNotNilForEveryElement(netInfo.LinkSpeedInfo) {
		tryEmitNetworkLegacyMetric(ch, timestamp, netInfo.LinkSpeedInfo.Speed,
			extendedLabel, npuChipPortLinkSpeedLegacyDescs, i)
	}
}

func addNetWorkLegacyMetricsDesc(ch chan<- *prometheus.Desc) {
	// Send legacy Desc for backward compatibility
	if colcommon.EnableLegacyMetrics {
		for _, desc := range linkStatusLegacyDescs {
			ch <- desc
		}
		for _, desc := range bandwidthTxLegacyDescs {
			ch <- desc
		}
		for _, desc := range bandwidthRxLegacyDescs {
			ch <- desc
		}
		for _, desc := range npuChipPortLinkSpeedLegacyDescs {
			ch <- desc
		}
	}
}
