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

package plugins

import (
	"context"
	"errors"
	"os"
	"slices"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"ascend-common/enpumanager"
	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/collector/prometheushttp"
	"huawei.com/npu-exporter/v6/utils/logger"
)

const (
	// enpuGroup identifies the plugin and the Telegraf measurement.
	enpuGroup = "enpu"
	// enpuRequestTimeout bounds a complete manager metrics request.
	enpuRequestTimeout = 5 * time.Second
	// enpuFamilyLimit bounds parsed families before filtering unknown metrics.
	enpuFamilyLimit = 100
	// enpuMetricFamilyCount is the number of metric families in the manager contract.
	enpuMetricFamilyCount = 16
	// enpuMaxUtilizationPercent is the upper bound of a utilization or quota percentage.
	enpuMaxUtilizationPercent = 100
	// enpuMetricDisabled represents a false validity or enabled flag.
	enpuMetricDisabled = 0
	// enpuMetricEnabled represents a true flag or the presence of an allocation.
	enpuMetricEnabled = 1
	// enpuDialTimeout delegates the connection deadline to the request context.
	enpuDialTimeout = 0
)

// enpuCollector adapts manager Prometheus metrics to both exporter outputs.
type enpuCollector struct {
	common.MetricsCollectorAdapter
	metrics *prometheushttp.Collector
}

var _ common.MetricsCollector = (*enpuCollector)(nil)

// newEnpuCollector connects the shared client to the fixed manager metric contract.
func newEnpuCollector() (*enpuCollector, error) {
	client := enpumanager.NewClient(enpumanager.SocketPath, enpuDialTimeout)
	collector, err := prometheushttp.New(client, enpuGaugeSpecs(), enpuFamilyLimit)
	if err != nil {
		return nil, err
	}
	return &enpuCollector{metrics: collector}, nil
}

// Describe publishes descriptors before metrics become available.
func (c *enpuCollector) Describe(ch chan<- *prometheus.Desc) {
	c.metrics.Describe(ch)
}

// CollectToCache fetches Prometheus text and replaces the shared output snapshot.
func (c *enpuCollector) CollectToCache(_ *common.NpuCollector, _ []common.HuaWeiAIChip) {
	ctx, cancel := context.WithTimeout(context.Background(), enpuRequestTimeout)
	defer cancel()
	if err := c.metrics.CollectToCache(ctx); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logger.Warnf("collect enpu-manager metrics: %v", err)
		}
		return
	}
	logger.Debugf("enpu-manager metrics snapshot: samples=%d", c.metrics.SampleCount())
}

// UpdatePrometheus sends cached manager gauges to Prometheus.
func (c *enpuCollector) UpdatePrometheus(ch chan<- prometheus.Metric, _ *common.NpuCollector,
	_ map[int32][]container.DevicesInfo, _ []common.HuaWeiAIChip) {
	c.metrics.UpdatePrometheus(ch)
}

// UpdateTelegraf converts cached manager gauges to enpu measurement fields.
func (c *enpuCollector) UpdateTelegraf(ch chan<- common.TelegrafMetric, _ *common.NpuCollector,
	_ map[int32][]container.DevicesInfo, _ []common.HuaWeiAIChip) {
	for _, sample := range c.metrics.GetSamples() {
		metric := common.NewGeneralMetric()
		metric.Measurement = enpuGroup
		metric.Labels = sample.Labels
		metric.Fields[sample.Name] = sample.Value
		metric.Timestamp = sample.Timestamp
		ch <- metric
	}
}

// enpuGaugeSpecs returns the metric families and labels in the manager metrics contract.
func enpuGaugeSpecs() []prometheushttp.GaugeSpec {
	physical := []string{"phy_id"}
	allocation := []string{"allocation_id", "phy_id", "vnpu_id", "pod_uid", "container_name"}
	percent := func(value float64, _ []string) bool { return value >= 0 && value <= enpuMaxUtilizationPercent }
	nonnegative := func(value float64, _ []string) bool { return value >= 0 }
	binary := func(value float64, _ []string) bool { return value == enpuMetricDisabled || value == enpuMetricEnabled }
	one := func(value float64, _ []string) bool { return value == enpuMetricEnabled }

	physicalUtil := []string{"utilization_percent", "aicore_utilization_percent", "aicpu_utilization_percent",
		"ctrlcpu_utilization_percent", "cube_utilization_percent", "vector_utilization_percent"}
	specs := make([]prometheushttp.GaugeSpec, 0, enpuMetricFamilyCount)
	for _, name := range physicalUtil {
		specs = append(specs, prometheushttp.GaugeSpec{Name: "enpu_npu_" + name,
			Help: "Physical device utilization percentage.", LabelNames: physical, Validate: percent})
	}
	return append(specs,
		prometheushttp.GaugeSpec{Name: "enpu_npu_metric_valid", Help: "Whether a physical metric has a valid current sample.",
			LabelNames: []string{"phy_id", "metric"}, Validate: enpuValidityFor(physicalUtil)},
		prometheushttp.GaugeSpec{Name: "enpu_allocation_info", Help: "Allocation identity, metadata and assigned resource.",
			LabelNames: allocation, Validate: one},
		prometheushttp.GaugeSpec{Name: "enpu_vnpu_utilization_percent", Help: "Estimated vNPU utilization percentage.",
			LabelNames: allocation, Validate: percent},
		prometheushttp.GaugeSpec{Name: "enpu_vnpu_aicore_utilization_percent", Help: "Estimated vNPU AI Core utilization percentage.",
			LabelNames: allocation, Validate: percent},
		prometheushttp.GaugeSpec{Name: "enpu_vnpu_aicore_quota_percent",
			Help: "Configured AI Core quota percentage when limiting is enabled.", LabelNames: allocation, Validate: percent},
		prometheushttp.GaugeSpec{Name: "enpu_vnpu_core_limit_enabled", Help: "Whether AI Core limiting is enabled.",
			LabelNames: allocation, Validate: binary},
		prometheushttp.GaugeSpec{Name: "enpu_vnpu_hbm_request_bytes", Help: "Configured HBM request in bytes.",
			LabelNames: allocation, Validate: nonnegative},
		prometheushttp.GaugeSpec{Name: "enpu_vnpu_hbm_limit_bytes", Help: "Configured HBM limit in bytes.",
			LabelNames: allocation, Validate: nonnegative},
		prometheushttp.GaugeSpec{Name: "enpu_vnpu_hbm_used_bytes", Help: "Source container HBM usage on the configured device in bytes.",
			LabelNames: allocation, Validate: nonnegative},
		prometheushttp.GaugeSpec{Name: "enpu_vnpu_metric_valid", Help: "Whether a resource metric has a valid current sample.",
			LabelNames: append(slices.Clone(allocation), "metric"),
			Validate:   enpuValidityFor([]string{"hbm_used_bytes", "utilization_percent", "aicore_utilization_percent"})},
	)
}

// enpuValidityFor checks the validity flag and its referenced metric name.
func enpuValidityFor(allowed []string) func(float64, []string) bool {
	return func(value float64, labels []string) bool {
		return (value == enpuMetricDisabled || value == enpuMetricEnabled) && slices.Contains(allowed, labels[len(labels)-1])
	}
}
