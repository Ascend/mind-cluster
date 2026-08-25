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
	"reflect"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"huawei.com/dpu-exporter/pkg/device"
)

// Metric label names
const (
	// CardLabelName is the label for DPU card name (e.g. "hinic0")
	CardLabelName = "card"
	// CardTypeLabel is the label for DPU card type (e.g. "CAL_2X400G_UB_EXP")
	CardTypeLabel = "card_type"
	// InterfaceLabel is the label for network interface name (e.g. "ens1f0")
	InterfaceLabel = "interface"
)

// DPU labels used in Prometheus output
var (
	// DpuCardLabels are labels for DPU-level (global) metrics
	DpuCardLabels = []string{CardLabelName, CardTypeLabel}
	// DpuIfaceLabels are labels for interface-level metrics
	DpuIfaceLabels = []string{CardLabelName, InterfaceLabel}
)

// CollectorContext provides access to the collector engine's state.
// Defined here (in metricscollector) to break the import cycle:
// dpucollector imports metricscollector, but metricscollector does NOT import dpucollector.
// DpuCollector implements this interface.
type CollectorContext interface {
	// GetDmgr returns the device manager for hardware access
	GetDmgr() device.DeviceManager
	// GetCache returns the cache for reading/writing metrics
	GetCache() CacheAccessor
}

// CacheAccessor provides read/write access to the metric cache.
// DpuCache in dpucollector implements this interface.
type CacheAccessor interface {
	// Read methods
	GetDpuList() []device.DPU
	GetDpuMetrics(cardName string) map[string]float64
	GetIfaceMetrics(ethName string) map[string]float64
	// Write methods
	SetDpuMetrics(cardName string, metrics map[string]float64)
	SetIfaceMetrics(ethName string, metrics map[string]float64)
	SetDpuList(dpuList []device.DPU)
}

// MetricsCollector is the interface that all DPU metric collectors must implement.
// Per the class diagram, each collector takes CollectorContext (not *DpuCollector directly).
type MetricsCollector interface {
	// Describe registers Prometheus metric descriptions
	Describe(ch chan<- *prometheus.Desc)

	// CollectToCache collects data from hardware and writes to cache
	CollectToCache(ctx CollectorContext)

	// UpdatePrometheus reads from cache and outputs Prometheus metrics
	UpdatePrometheus(ch chan<- prometheus.Metric, ctx CollectorContext)

	// PreCollect is called once before the collect loop starts
	PreCollect(ctx CollectorContext)

	// PostCollect is called once when the collect loop exits
	PostCollect(ctx CollectorContext)

	// IsSupported returns true if the current hardware supports this collector
	IsSupported(ctx CollectorContext) bool

	// GetInterval returns the configured collection interval for this collector
	GetInterval() time.Duration
}

// MetricsCollectorAdapter provides default no-op implementations for MetricsCollector.
// Concrete collectors embed this struct and only override the methods they need.
type MetricsCollectorAdapter struct{}

// Describe default no-op
func (c *MetricsCollectorAdapter) Describe(ch chan<- *prometheus.Desc) {}

// CollectToCache default no-op
func (c *MetricsCollectorAdapter) CollectToCache(ctx CollectorContext) {}

// UpdatePrometheus default no-op
func (c *MetricsCollectorAdapter) UpdatePrometheus(ch chan<- prometheus.Metric, ctx CollectorContext) {
}

// PreCollect default no-op
func (c *MetricsCollectorAdapter) PreCollect(ctx CollectorContext) {}

// PostCollect default no-op
func (c *MetricsCollectorAdapter) PostCollect(ctx CollectorContext) {}

// IsSupported default returns true
func (c *MetricsCollectorAdapter) IsSupported(ctx CollectorContext) bool { return true }

// GetInterval default returns 0 (use scheduler default)
func (c *MetricsCollectorAdapter) GetInterval() time.Duration { return 0 }

// GetCacheKey returns the struct type name of the collector, used as the cache key
func GetCacheKey(ptr interface{}) string {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		return ""
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return ""
	}
	return v.Type().Name()
}

// BuildDesc builds a prometheus.Desc with DPU card labels
func BuildDesc(name, help string) *prometheus.Desc {
	return prometheus.NewDesc(name, help, DpuCardLabels, nil)
}

// BuildIfaceDesc builds a prometheus.Desc with interface labels
func BuildIfaceDesc(name, help string) *prometheus.Desc {
	return prometheus.NewDesc(name, help, DpuIfaceLabels, nil)
}

// BuildDescWithLabels builds a prometheus.Desc with custom labels
func BuildDescWithLabels(name, help string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(name, help, labels, nil)
}
