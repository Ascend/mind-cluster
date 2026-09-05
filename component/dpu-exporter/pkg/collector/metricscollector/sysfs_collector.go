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
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"huawei.com/dpu-exporter/pkg/configmanager"
	"huawei.com/dpu-exporter/pkg/device"
	"huawei.com/dpu-exporter/utils/logger"
)

// sysfsNetBase is the base path for sysfs network interface entries
const sysfsNetBase = "/sys/class/net"

// sysfsIfaceFiles are the interface-level files directly under /sys/class/net/<interface>/
// These are NOT under the statistics/ subdirectory.
var sysfsIfaceFiles = []string{"carrier", "carrier_changes", "operstate"}

// sysfsGaugeMetrics are interface metrics that represent current state (gauge type),
// not cumulative counters.
var sysfsGaugeMetrics = map[string]struct{}{
	"carrier":   {},
	"operstate": {},
}

// SysfsCollector collects interface-level metrics from /sys/class/net/.
// It reads:
//   - /sys/class/net/<interface>/{carrier, carrier_changes, operstate}
//   - /sys/class/net/<interface>/statistics/* (dynamically discovered)
//
// These reads are very fast (<1ms per file), so all interfaces are collected
// in a single goroutine sequentially.
type SysfsCollector struct {
	MetricsCollectorAdapter
	// localCache stores per-interface metrics during a single collection round
	localCache sync.Map
}

// UpdatePrometheus reads cached sysfs metrics and outputs them as Prometheus metrics.
// Descriptions are created on-the-fly since statistics files are dynamically discovered.
func (c *SysfsCollector) UpdatePrometheus(ch chan<- prometheus.Metric, ctx CollectorContext) {
	dpuList := ctx.GetCache().GetDpuList()

	for i := range dpuList {
		for j := range dpuList[i].Interfaces {
			metrics := ctx.GetCache().GetIfaceMetrics(dpuList[i].Interfaces[j].EthName)
			for name, val := range metrics {
				desc := prometheus.NewDesc(
					fmt.Sprintf("dpu_interface_%s", name),
					fmt.Sprintf("interface metric: %s", name),
					DpuIfaceLabels, nil,
				)
				metricType := prometheus.CounterValue
				if _, ok := sysfsGaugeMetrics[name]; ok {
					metricType = prometheus.GaugeValue
				}
				ch <- prometheus.MustNewConstMetric(desc, metricType, val,
					dpuList[i].CardName, dpuList[i].Interfaces[j].EthName)
			}
		}
	}
}

// CollectToCache reads sysfs metrics for all interfaces across all DPU cards.
func (c *SysfsCollector) CollectToCache(ctx CollectorContext) {
	dpuList := ctx.GetCache().GetDpuList()
	if len(dpuList) == 0 {
		logger.Warn("no dpu found, skip sysfs collect")
		return
	}

	dmgr := ctx.GetDmgr()
	totalIfaces := 0

	for i := range dpuList {
		for j := range dpuList[i].Interfaces {
			c.collectForInterface(dmgr, dpuList[i].Interfaces[j])
			totalIfaces++
		}
	}

	logger.Debugf("sysfs collected metrics for %d interfaces", totalIfaces)
}

// collectForInterface reads all sysfs metrics for a single interface:
//  1. Interface-level files: carrier, carrier_changes, operstate
//  2. Statistics files: dynamically discovered from /sys/class/net/<iface>/statistics/
func (c *SysfsCollector) collectForInterface(dmgr device.DeviceManager, iface device.Interface) {
	metrics := make(map[string]float64)

	// 1. Read interface-level files (carrier, carrier_changes, operstate)
	for _, fileName := range sysfsIfaceFiles {
		path := filepath.Join(sysfsNetBase, iface.EthName, fileName)
		raw, err := dmgr.ReadSysfs(path)
		if err != nil {
			if fileName == "carrier" {
				metrics[fileName] = -1
			}
			continue
		}
		val, parseErr := parseSysfsValue(fileName, raw)
		if parseErr != nil {
			continue
		}
		metrics[fileName] = val
	}

	// 2. Dynamically discover and read statistics files
	statsDir := filepath.Join(sysfsNetBase, iface.EthName, "statistics")
	statFiles, err := dmgr.ListDir(statsDir)
	if err != nil {
		logger.Debugf("failed to list statistics dir for %s: %v", iface.EthName, err)
	} else {
		for _, fileName := range statFiles {
			path := filepath.Join(statsDir, fileName)
			raw, readErr := dmgr.ReadSysfs(path)
			if readErr != nil {
				continue
			}
			val, parseErr := strconv.ParseFloat(raw, 64)
			if parseErr != nil {
				continue
			}
			metrics[fileName] = val
		}
	}

	// Store in local cache
	c.localCache.Store(iface.EthName, metrics)

	// Write to DpuCache via the global CollectorContext
	if globalCtx != nil {
		globalCtx.GetCache().SetIfaceMetrics(iface.EthName, metrics)
	}
}

// parseSysfsValue parses a sysfs file value into a float64.
// Special handling for non-numeric values like operstate ("up"/"down").
func parseSysfsValue(fileName, raw string) (float64, error) {
	// operstate returns string values, map them to numeric
	if fileName == "operstate" {
		switch raw {
		case "up":
			return 1, nil
		case "down":
			return 0, nil
		default:
			return -1, nil // unknown state
		}
	}
	// carrier is boolean: report 0/1 as-is, anything else as -1 (unknown)
	if fileName == "carrier" {
		switch raw {
		case "1":
			return 1, nil
		case "0":
			return 0, nil
		default:
			return -1, nil
		}
	}
	return strconv.ParseFloat(raw, 64)
}

// globalCtx is set by the engine to allow collectors to write to cache
// without requiring CollectorContext in every method signature
var globalCtx CollectorContext

// SetGlobalContext sets the global collector context (called by engine at startup)
func SetGlobalContext(ctx CollectorContext) {
	globalCtx = ctx
}

// GetInterval returns the configured interval for this collector
func (c *SysfsCollector) GetInterval() time.Duration {
	return configmanager.GetCollectorInterval(configmanager.CacheKeySysfs, configmanager.DefaultGroupInterval)
}
