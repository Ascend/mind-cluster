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

package metricscollector

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"ascend-common/common-utils/hwlog"
	"huawei.com/dpu-exporter/pkg/configmanager"
	"huawei.com/dpu-exporter/utils/logger"
)

const (
	// rdmaMaxMissRounds is the number of consecutive missing rounds tolerated
	// before a metric or device is removed from the cache.
	rdmaMaxMissRounds = 3
	// rdmaCmdResource and rdmaCmdShow form the only supported command.
	rdmaCmdResource = "resource"
	rdmaCmdShow     = "show"
	// hwlog duplicate-log limiting categories.
	rdmaLogID             = 0
	rdmaEmptyOutputDomain = "rdma_empty_output"
	rdmaNoParsedDomain    = "rdma_no_parsed_devices"

	// Regex capture groups and field layout of one summary line.
	rdmaSummaryDeviceGroup          = 1
	rdmaSummaryFieldsGroup          = 2
	rdmaResourceFieldPair           = 2
	rdmaCurrentValueParts           = 2
	rdmaParseBase, rdmaParseBitSize = 10, 64
)

// rdmaSummaryRe matches one summary line. Device names are environment-specific;
// on the target DPU nodes they look like hrn5_0:
//
//	0: hrn5_0: pd 3 cq 130 qp 1 cm_id 0 mr 0 ctx 0 srq 2
var rdmaSummaryRe = regexp.MustCompile(`^\s*\d+:\s+(\S+):\s+(.+)$`)

// rdmaMetricDef describes one exported RDMA metric.
type rdmaMetricDef struct {
	resource   string
	metricName string
	desc       *prometheus.Desc
}

// newRdmaMetricDef keeps the exported metric name and Desc in sync.
func newRdmaMetricDef(resource, metricName, help string) rdmaMetricDef {
	return rdmaMetricDef{
		resource:   resource,
		metricName: metricName,
		desc:       BuildDescWithLabels(metricName, help, DpuRdmaLabels),
	}
}

// rdmaMetricDefs defines the RDMA metrics exported by RdmaCollector.
var rdmaMetricDefs = []rdmaMetricDef{
	newRdmaMetricDef("pd", "dpu_roce_pd",
		"Number of RDMA protection domains (PD) on the RDMA device"),
	newRdmaMetricDef("cq", "dpu_roce_cq",
		"Number of RDMA completion queues (CQ) on the RDMA device"),
	newRdmaMetricDef("qp", "dpu_roce_qp",
		"Number of RDMA queue pairs (QP) on the RDMA device"),
	newRdmaMetricDef("cm_id", "dpu_roce_cm_id",
		"Number of RDMA communication manager IDs (CM ID) on the RDMA device"),
	newRdmaMetricDef("mr", "dpu_roce_mr",
		"Number of RDMA memory regions (MR) on the RDMA device"),
	newRdmaMetricDef("ctx", "dpu_roce_ctx",
		"Number of RDMA contexts (CTX) on the RDMA device"),
	newRdmaMetricDef("srq", "dpu_roce_srq",
		"Number of RDMA shared receive queues (SRQ) on the RDMA device"),
}

// rdmaResourceWhitelist maps supported resource fields to their exported names.
var rdmaResourceWhitelist = func() map[string]string {
	m := make(map[string]string, len(rdmaMetricDefs))
	for _, def := range rdmaMetricDefs {
		m[def.resource] = def.metricName
	}
	return m
}()

// RdmaCollector collects RDMA resource counts from `rdma resource show`.
// It is a host-level collector: one invocation covers every RDMA device visible
// in the exporter's network namespace.
type RdmaCollector struct {
	MetricsCollectorAdapter
}

// Describe reports the seven RDMA metric descriptors.
func (c *RdmaCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, def := range rdmaMetricDefs {
		ch <- def.desc
	}
}

// CollectToCache executes rdma resource show and updates the per-device metrics.
func (c *RdmaCollector) CollectToCache(ctx CollectorContext) {
	output, err := ctx.GetDmgr().ExecRdma(rdmaCmdResource, rdmaCmdShow)
	if err != nil {
		c.applyRdmaRound(ctx, nil)
		return
	}

	if strings.TrimSpace(output) == "" {
		hwlog.RunLog.WarnfWithLimit(rdmaEmptyOutputDomain, rdmaLogID,
			"rdma resource show output is empty")
		c.applyRdmaRound(ctx, nil)
		return
	}

	parsed := parseRdmaResourceSummary(output)
	if len(parsed) == 0 {
		hwlog.RunLog.WarnfWithLimit(rdmaNoParsedDomain, rdmaLogID,
			"rdma resource show output contains no parseable devices")
	} else {
		hwlog.ResetErrCnt(rdmaNoParsedDomain, rdmaLogID)
	}
	hwlog.ResetErrCnt(rdmaEmptyOutputDomain, rdmaLogID)
	c.applyRdmaRound(ctx, parsed)
}

// parseRdmaResourceSummary parses the text output of `rdma resource show`.
func parseRdmaResourceSummary(output string) map[string]map[string]float64 {
	result := make(map[string]map[string]float64)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		m := rdmaSummaryRe.FindStringSubmatch(line)
		if m == nil {
			logger.Debugf("rdma resource: skip unexpected line %q", line)
			continue
		}

		deviceName := m[rdmaSummaryDeviceGroup]
		fields := strings.Fields(m[rdmaSummaryFieldsGroup])
		counts := make(map[string]float64, len(rdmaMetricDefs))
		for i := 0; i+1 < len(fields); i += rdmaResourceFieldPair {
			resource, raw := fields[i], fields[i+1]
			metricName, ok := rdmaResourceWhitelist[resource]
			if !ok {
				continue
			}

			current := strings.SplitN(raw, "/", rdmaCurrentValueParts)[0]
			value, err := strconv.ParseUint(current, rdmaParseBase, rdmaParseBitSize)
			if err != nil {
				logger.Debugf("rdma resource: invalid %s for %s: %v", metricName, deviceName, err)
				continue
			}
			counts[resource] = float64(value)
		}

		result[deviceName] = counts
	}
	return result
}

// applyRdmaRound updates every known or newly discovered device. `parsed` may be
// nil, in which case all known devices are treated as missing for this round.
func (c *RdmaCollector) applyRdmaRound(ctx CollectorContext, parsed map[string]map[string]float64) {
	cache := ctx.GetCache()
	prevAll := cache.GetAllRdmaMetrics()

	devices := make(map[string]struct{}, len(prevAll)+len(parsed))
	for deviceName := range prevAll {
		devices[deviceName] = struct{}{}
	}
	for deviceName := range parsed {
		devices[deviceName] = struct{}{}
	}

	nextAll := make(map[string]map[string]RdmaMetricValue, len(devices))
	for deviceName := range devices {
		prev := prevAll[deviceName]
		next := make(map[string]RdmaMetricValue, len(rdmaMetricDefs))
		for _, def := range rdmaMetricDefs {
			prevMetric, hadPrev := prev[def.resource]
			value, hasValue := parsed[deviceName][def.resource]
			if hasValue {
				next[def.resource] = RdmaMetricValue{Value: value}
				if hadPrev && prevMetric.MissCount > 0 {
					logger.Infof("rdma metric %s for device %s recovered", def.metricName, deviceName)
				}
				continue
			}

			if !hadPrev {
				continue
			}
			missCount := prevMetric.MissCount + 1
			if missCount >= rdmaMaxMissRounds {
				logger.Warnf("remove rdma metric %s for device %s after %d consecutive missing rounds",
					def.metricName, deviceName, missCount)
				continue
			}
			prevMetric.MissCount = missCount
			next[def.resource] = prevMetric
		}
		if len(next) > 0 {
			nextAll[deviceName] = next
		}
	}
	cache.SetAllRdmaMetrics(nextAll)
}

// UpdatePrometheus emits one gauge per cached RDMA device and metric.
func (c *RdmaCollector) UpdatePrometheus(ch chan<- prometheus.Metric, ctx CollectorContext) {
	all := ctx.GetCache().GetAllRdmaMetrics()
	for deviceName, metrics := range all {
		for _, def := range rdmaMetricDefs {
			if metric, ok := metrics[def.resource]; ok {
				ch <- prometheus.MustNewConstMetric(def.desc, prometheus.GaugeValue, metric.Value, deviceName)
			}
		}
	}
}

// GetInterval returns the configured interval for RDMA collection.
func (c *RdmaCollector) GetInterval() time.Duration {
	return configmanager.GetCollectorInterval(configmanager.CacheKeyRdma, configmanager.DefaultRdmaInterval)
}
