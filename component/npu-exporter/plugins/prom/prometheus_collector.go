/* Copyright(C) 2021-2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package prometheus for prometheus collector
package prom

import (
	"github.com/prometheus/client_golang/prometheus"

	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/utils/logger"
)

// CollectorForPrometheus Entry point for collecting and converting
type CollectorForPrometheus struct {
	collector *common.NpuCollector
}

// NewPrometheusCollector create an instance of prometheus Collector
func NewPrometheusCollector(collector *common.NpuCollector) *CollectorForPrometheus {
	promCollector := &CollectorForPrometheus{
		collector: collector,
	}
	return promCollector
}

// Describe desc metrics of prometheus
func (*CollectorForPrometheus) Describe(ch chan<- *prometheus.Desc) {
	if ch == nil {
		logger.Error("ch is nil ")
		return
	}
	describeChain(ch, common.ChainForSingleGoroutine)
	describeChain(ch, common.ChainForMultiGoroutine)
}

func describeChain(ch chan<- *prometheus.Desc, chain []common.MetricsCollector) {
	for _, collector := range chain {
		collector.Describe(ch)
	}
}

// Collect update metrics of prometheus
func (n *CollectorForPrometheus) Collect(ch chan<- prometheus.Metric) {
	containerMap := common.GetContainerNPUInfo(n.collector)
	chips := common.GetChipListWithVNPU(n.collector)
	collectChain(ch, n, containerMap, chips, common.ChainForSingleGoroutine)
	collectChain(ch, n, containerMap, chips, common.ChainForMultiGoroutine)
}

func collectChain(ch chan<- prometheus.Metric, n *CollectorForPrometheus, containerMap map[int32]container.DevicesInfo,
	chips []common.HuaWeiAIChip, chain []common.MetricsCollector) {
	if ch == nil {
		logger.Error("ch is nil")
		return
	}
	for _, collector := range chain {
		collector.UpdatePrometheus(ch, n.collector, containerMap, chips)
	}
}
