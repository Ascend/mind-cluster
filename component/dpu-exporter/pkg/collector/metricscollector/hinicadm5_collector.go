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
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"huawei.com/dpu-exporter/pkg/configmanager"
	"huawei.com/dpu-exporter/pkg/device"
	"huawei.com/dpu-exporter/utils/logger"
)

// hinicadm5 counter command constants
const (
	counterType    = "-t"
	counterTypeVal = "1"
	cardFlag       = "-i"
	featureFlag    = "-f"
	featureRoce    = "feature:ROCE"
)

// Hinicadm5Collector collects global DPU metrics via the hinicadm5 CLI tool.
// Each invocation of `hinicadm5 counter -t 1 -i <card> -f feature:ROCE` takes ~1.5s
// and returns 300+ metric items. A whitelist is used to filter down to the desired metrics.
type Hinicadm5Collector struct {
	MetricsCollectorAdapter
	// localCache stores per-card metrics during a single collection round
	localCache sync.Map
}

// Describe reports the metric description to Prometheus
func (c *Hinicadm5Collector) Describe(ch chan<- *prometheus.Desc) {
	desc := BuildDesc(
		"dpu_hinicadm5_metric",
		"Global DPU metric collected via hinicadm5 counter command",
	)
	ch <- desc
}

// CollectToCache executes hinicadm5 counter for each DPU card and stores results in cache.
func (c *Hinicadm5Collector) CollectToCache(ctx CollectorContext) {
	dpuList := ctx.GetCache().GetDpuList()
	if len(dpuList) == 0 {
		logger.Warn("no dpu found, skip hinicadm5 collect")
		return
	}

	dmgr := ctx.GetDmgr()
	whitelist := configmanager.GetWhitelist()

	// hinicadm5 does not support concurrent execution, collect sequentially
	for i := range dpuList {
		c.collectForCard(dmgr, dpuList[i].CardName, whitelist, ctx)
	}
}

// collectForCard executes the hinicadm5 counter command for a single DPU card,
// parses the output, filters by whitelist, and writes to cache.
func (c *Hinicadm5Collector) collectForCard(
	dmgr device.DeviceManager,
	cardName string,
	whitelist *configmanager.WhitelistManager,
	ctx CollectorContext,
) {
	args := []string{
		"counter",
		counterType, counterTypeVal,
		cardFlag, cardName,
		featureFlag, featureRoce,
	}

	output, err := dmgr.ExecCommand(args...)
	if err != nil {
		logger.Errorf("hinicadm5 counter failed for card %s: %v", cardName, err)
		return
	}

	metrics := parseHinicadm5CounterOutput(output)
	if len(metrics) == 0 {
		logger.Warnf("no metrics parsed from hinicadm5 output for card %s", cardName)
		return
	}

	// Apply whitelist filter
	if whitelist != nil {
		metrics = whitelist.Filter(metrics)
	}

	logger.Debugf("hinicadm5 collected %d metrics for card %s", len(metrics), cardName)

	// Store in local cache
	c.localCache.Store(cardName, metrics)

	// Write to DpuCache
	ctx.GetCache().SetDpuMetrics(cardName, metrics)
}

// parseHinicadm5CounterOutput parses the text output of `hinicadm5 counter` command.
// Input format per value line:
//
//	ID-0x1902::0000:ROCE_CMDQ_CTR_EXT_CMD: 2
//
// We extract the metric name part (ROCE_CMDQ_CTR_EXT_CMD), lowercase it,
// and strip parentheses for Prometheus compatibility → roce_cmdq_ctr_ext_cmd
func parseHinicadm5CounterOutput(output string) map[string]float64 {
	metrics := make(map[string]float64)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "Card num") || strings.Contains(line, "Device Information") ||
			strings.Contains(line, "NPU Statistics") {
			continue
		}

		// Skip header lines like: ROCE_CMDQ_CTR_EXT_CMD(ROCE|CMDQ|INFO):
		// These have parentheses and no numeric value
		if strings.Contains(line, "(") && strings.Contains(line, ")") {
			continue
		}

		// Parse value lines: ID-0x1902::0000:ROCE_CMDQ_CTR_EXT_CMD: 2
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		valueStr := parts[len(parts)-1]
		val, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			continue
		}

		// Extract metric name from the ID field
		// Format: ID-<hex>::<instance>:<METRIC_NAME>:
		idField := parts[0]
		metricName := extractMetricName(idField)
		if metricName == "" {
			continue
		}

		metrics[metricName] = val
	}
	return metrics
}

// extractMetricName extracts the clean metric name from an ID field.
// Input:  "ID-0x1902::0000:ROCE_CMDQ_CTR_EXT_CMD:"
// Output: "roce_cmdq_ctr_ext_cmd" (lowercased, colons and ID prefix stripped)
func extractMetricName(idField string) string {
	// Strip "ID-" prefix
	s := idField
	if idx := strings.Index(s, ":"); idx != -1 {
		// Find the metric name part after the last colon-separated segment
		// Split by ":" and take the last non-empty segment before the trailing ":"
		segments := strings.Split(s, ":")
		// segments for "ID-0x1902::0000:ROCE_CMDQ_CTR_EXT_CMD:" =
		//   ["ID-0x1902", "", "0000", "ROCE_CMDQ_CTR_EXT_CMD", ""]
		for i := len(segments) - 1; i >= 0; i-- {
			seg := strings.TrimSpace(segments[i])
			if seg != "" && !strings.HasPrefix(seg, "ID-") && !strings.HasPrefix(seg, "0x") {
				return strings.ToLower(seg)
			}
		}
	}
	return ""
}

// UpdatePrometheus reads cached hinicadm5 metrics and outputs them as Prometheus metrics.
func (c *Hinicadm5Collector) UpdatePrometheus(ch chan<- prometheus.Metric, ctx CollectorContext) {
	dpuList := ctx.GetCache().GetDpuList()
	for i := range dpuList {
		metrics := ctx.GetCache().GetDpuMetrics(dpuList[i].CardName)
		for name, val := range metrics {
			desc := prometheus.NewDesc(
				fmt.Sprintf("dpu_hinicadm5_%s", name),
				fmt.Sprintf("hinicadm5 metric: %s", name),
				DpuCardLabels, nil,
			)
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, val,
				dpuList[i].CardName, dpuList[i].CardType)
		}
	}
}

// IsSupported returns true if the device manager is Huawei
func (c *Hinicadm5Collector) IsSupported(ctx CollectorContext) bool {
	return ctx.GetDmgr().GetCardType() == device.CardTypeHuawei
}

// GetInterval returns the configured interval for this collector
func (c *Hinicadm5Collector) GetInterval() time.Duration {
	return configmanager.GetCollectorInterval(configmanager.CacheKeyHinicadm5, configmanager.DefaultGroupInterval)
}
