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
	opticalIndexLegacyDescs    []*prometheus.Desc
	opticalTxPower0LegacyDescs []*prometheus.Desc
	opticalTxPower1LegacyDescs []*prometheus.Desc
	opticalTxPower2LegacyDescs []*prometheus.Desc
	opticalTxPower3LegacyDescs []*prometheus.Desc
	opticalRxPower0LegacyDescs []*prometheus.Desc
	opticalRxPower1LegacyDescs []*prometheus.Desc
	opticalRxPower2LegacyDescs []*prometheus.Desc
	opticalRxPower3LegacyDescs []*prometheus.Desc
)

func initOpticalLegacyDesc() {
	if !colcommon.EnableLegacyMetrics {
		return
	}
	opticalIndexLegacyDescs = buildLegacyDescSlice("optical_index_num", "the npu link optical index num on ub port")
	opticalTxPower0LegacyDescs = buildLegacyDescSlice("optical_tx_power_0", "optical tx power lane 0 on ub port")
	opticalTxPower1LegacyDescs = buildLegacyDescSlice("optical_tx_power_1", "optical tx power lane 1 on ub port")
	opticalTxPower2LegacyDescs = buildLegacyDescSlice("optical_tx_power_2", "optical tx power lane 2 on ub port")
	opticalTxPower3LegacyDescs = buildLegacyDescSlice("optical_tx_power_3", "optical tx power lane 3 on ub port")
	opticalRxPower0LegacyDescs = buildLegacyDescSlice("optical_rx_power_0", "optical rx power lane 0 on ub port")
	opticalRxPower1LegacyDescs = buildLegacyDescSlice("optical_rx_power_1", "optical rx power lane 1 on ub port")
	opticalRxPower2LegacyDescs = buildLegacyDescSlice("optical_rx_power_2", "optical rx power lane 2 on ub port")
	opticalRxPower3LegacyDescs = buildLegacyDescSlice("optical_rx_power_3", "optical rx power lane 3 on ub port")
}

func tryEmitOpticalLegacyMetric(ch chan<- prometheus.Metric, timestamp time.Time,
	value interface{}, extendedLabel []string, legacyDescs []*prometheus.Desc, i int) {
	if !colcommon.EnableLegacyMetrics {
		return
	}
	if i < len(legacyDescs) {
		doUpdateMetric(ch, timestamp, value, extendedLabel[:len(extendedLabel)-2], legacyDescs[i])
	}
}

func tryEmitOpticalLegacyMetricFloat(ch chan<- prometheus.Metric, timestamp time.Time,
	value float64, extendedLabel []string, legacyDescs []*prometheus.Desc, i int) {
	if i < len(legacyDescs) {
		doUpdateMetricWithValidateNum(ch, timestamp, value,
			extendedLabel[:len(extendedLabel)-2], legacyDescs[i])
	}
}

func promUpdateOpticalInfoLegacy(ch chan<- prometheus.Metric, timestamp time.Time,
	info *common.OpticalNpuInfo, extendedLabel []string, i int) {
	if !colcommon.EnableLegacyMetrics {
		return
	}
	tryEmitOpticalLegacyMetric(ch, timestamp, info.OpticalIndex, extendedLabel,
		opticalIndexLegacyDescs, i)
	tryEmitOpticalLegacyMetricFloat(ch, timestamp, info.OpticalTxPower0, extendedLabel,
		opticalTxPower0LegacyDescs, i)
	tryEmitOpticalLegacyMetricFloat(ch, timestamp, info.OpticalTxPower1, extendedLabel,
		opticalTxPower1LegacyDescs, i)
	tryEmitOpticalLegacyMetricFloat(ch, timestamp, info.OpticalTxPower2, extendedLabel,
		opticalTxPower2LegacyDescs, i)
	tryEmitOpticalLegacyMetricFloat(ch, timestamp, info.OpticalTxPower3, extendedLabel,
		opticalTxPower3LegacyDescs, i)
	tryEmitOpticalLegacyMetricFloat(ch, timestamp, info.OpticalRxPower0, extendedLabel,
		opticalRxPower0LegacyDescs, i)
	tryEmitOpticalLegacyMetricFloat(ch, timestamp, info.OpticalRxPower1, extendedLabel,
		opticalRxPower1LegacyDescs, i)
	tryEmitOpticalLegacyMetricFloat(ch, timestamp, info.OpticalRxPower2, extendedLabel,
		opticalRxPower2LegacyDescs, i)
	tryEmitOpticalLegacyMetricFloat(ch, timestamp, info.OpticalRxPower3, extendedLabel,
		opticalRxPower3LegacyDescs, i)
}

func addOpticalLegacyMetricsDesc(ch chan<- *prometheus.Desc) {
	// Send legacy Desc for backward compatibility
	if colcommon.EnableLegacyMetrics {
		for _, desc := range opticalIndexLegacyDescs {
			ch <- desc
		}
		for _, desc := range opticalTxPower0LegacyDescs {
			ch <- desc
		}
		for _, desc := range opticalTxPower1LegacyDescs {
			ch <- desc
		}
		for _, desc := range opticalTxPower2LegacyDescs {
			ch <- desc
		}
		for _, desc := range opticalTxPower3LegacyDescs {
			ch <- desc
		}
		for _, desc := range opticalRxPower0LegacyDescs {
			ch <- desc
		}
		for _, desc := range opticalRxPower1LegacyDescs {
			ch <- desc
		}
		for _, desc := range opticalRxPower2LegacyDescs {
			ch <- desc
		}
		for _, desc := range opticalRxPower3LegacyDescs {
			ch <- desc
		}
	}
}
