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

// Package prometheushttp fetches Prometheus text exposition over HTTP and publishes validated gauge snapshots.
package prometheushttp

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"huawei.com/npu-exporter/v6/utils/logger"
)

// Fetcher obtains an HTTP response body in Prometheus text exposition format.
type Fetcher interface {
	// FetchMetrics returns the complete Prometheus text body or an error.
	FetchMetrics(context.Context) ([]byte, error)
}

// GaugeSpec defines the public descriptor and optional sample validation for a gauge.
type GaugeSpec struct {
	Name       string
	Help       string
	LabelNames []string
	Validate   func(value float64, labels []string) bool
}

// gauge binds a metric contract to its Prometheus descriptor.
type gauge struct {
	GaugeSpec
	desc *prometheus.Desc
}

// Sample is a validated gauge with the timestamp shared by its collection snapshot.
type Sample struct {
	Name      string
	Labels    map[string]string
	Value     float64
	Timestamp time.Time
}

// sample stores one validated series in descriptor label order.
type sample struct {
	gauge     gauge
	values    []string
	value     float64
	timestamp time.Time
}

// snapshot contains the samples published by one successful collection.
type snapshot struct {
	samples []sample
}

// Collector publishes HTTP metric snapshots. Its caller decides when to refresh.
type Collector struct {
	fetcher     Fetcher
	maxFamilies int
	gauges      map[string]gauge
	refreshing  atomic.Bool
	current     atomic.Pointer[snapshot]
}

// New creates a collector for the supplied gauge definitions.
func New(fetcher Fetcher, specs []GaugeSpec, maxFamilies int) (*Collector, error) {
	if fetcher == nil {
		return nil, fmt.Errorf("HTTP metrics fetcher is required")
	}
	c := &Collector{fetcher: fetcher, maxFamilies: maxFamilies, gauges: make(map[string]gauge, len(specs))}
	for _, spec := range specs {
		if _, exists := c.gauges[spec.Name]; exists {
			return nil, fmt.Errorf("duplicate gauge definition %q", spec.Name)
		}
		spec.LabelNames = slices.Clone(spec.LabelNames)
		desc := prometheus.NewDesc(spec.Name, spec.Help, spec.LabelNames, nil)
		if _, err := prometheus.NewConstMetric(desc, prometheus.GaugeValue, 0,
			make([]string, len(spec.LabelNames))...); err != nil {
			return nil, fmt.Errorf("invalid gauge definition %q: %w", spec.Name, err)
		}
		c.gauges[spec.Name] = gauge{GaugeSpec: spec, desc: desc}
	}
	return c, nil
}

// CollectToCache fetches and publishes a complete snapshot. An overlapping refresh is skipped.
func (c *Collector) CollectToCache(ctx context.Context) error {
	if !c.refreshing.CompareAndSwap(false, true) {
		return nil
	}
	defer c.refreshing.Store(false)

	body, err := c.fetcher.FetchMetrics(ctx)
	if err != nil {
		c.current.Store(nil)
		return fmt.Errorf("fetch HTTP metrics: %w", err)
	}
	var parser expfmt.TextParser
	families, err := parser.TextToMetricFamilies(bytes.NewReader(body))
	if err != nil {
		c.current.Store(nil)
		return fmt.Errorf("parse HTTP metrics: %w", err)
	}

	if c.maxFamilies > 0 && len(families) > c.maxFamilies {
		logger.Warnf("HTTP metrics family count exceeds limit: actual=%d, limit=%d", len(families), c.maxFamilies)
		c.current.Store(nil)
		return fmt.Errorf("HTTP metrics response exceeds %d metric families", c.maxFamilies)
	}

	// Use one collection timestamp for every family and both output platforms.
	timestamp := time.Now()
	next := &snapshot{}
	for name, family := range families {
		spec, ok := c.gauges[name]
		if ok && family.GetType() == dto.MetricType_GAUGE {
			next.samples = append(next.samples, gaugeSamples(spec, family.GetMetric(), timestamp)...)
		}
	}
	c.current.Store(next)
	return nil
}

// SampleCount returns the number of samples in the currently published snapshot.
func (c *Collector) SampleCount() int {
	if current := c.current.Load(); current != nil {
		return len(current.samples)
	}
	return 0
}

// gaugeSamples validates values and excludes every occurrence of duplicate series.
func gaugeSamples(spec gauge, metrics []*dto.Metric, timestamp time.Time) []sample {
	var samples []sample
	var keys []string
	counts := make(map[string]int)
	for _, metric := range metrics {
		values, key, ok := sampleLabels(spec.LabelNames, metric.GetLabel())
		if !ok {
			continue
		}
		counts[key]++
		g := metric.GetGauge()
		if g == nil || g.Value == nil {
			continue
		}
		value := g.GetValue()
		if math.IsNaN(value) || math.IsInf(value, 0) ||
			(spec.Validate != nil && !spec.Validate(value, slices.Clone(values))) {
			continue
		}
		samples = append(samples, sample{gauge: spec, values: values, value: value, timestamp: timestamp})
		keys = append(keys, key)
	}
	unique := samples[:0]
	for i, s := range samples {
		if counts[keys[i]] == 1 {
			unique = append(unique, s)
		}
	}
	return unique
}

// sampleLabels validates the exact label set and builds an unambiguous series key.
func sampleLabels(names []string, labels []*dto.LabelPair) ([]string, string, bool) {
	if len(names) != len(labels) {
		return nil, "", false
	}
	byName := make(map[string]string, len(labels))
	for _, label := range labels {
		if _, duplicate := byName[label.GetName()]; duplicate {
			return nil, "", false
		}
		byName[label.GetName()] = label.GetValue()
	}
	values := make([]string, len(names))
	var key strings.Builder
	for i, name := range names {
		value, ok := byName[name]
		if !ok || !utf8.ValidString(value) {
			return nil, "", false
		}
		values[i] = value
		key.WriteString(strconv.Quote(value))
	}
	return values, key.String(), true
}

// Describe reports all configured descriptors before the first successful refresh.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, spec := range c.gauges {
		ch <- spec.desc
	}
}

// UpdatePrometheus emits the cached snapshot to the Prometheus output channel.
func (c *Collector) UpdatePrometheus(ch chan<- prometheus.Metric) {
	current := c.current.Load()
	if current == nil {
		return
	}
	for _, s := range current.samples {
		metric, err := prometheus.NewConstMetric(s.gauge.desc, prometheus.GaugeValue, s.value, s.values...)
		if err != nil {
			continue
		}
		if !s.timestamp.IsZero() {
			metric = prometheus.NewMetricWithTimestamp(s.timestamp, metric)
		}
		ch <- metric
	}
}

// GetSamples returns a copy of the current snapshot for other outputs.
func (c *Collector) GetSamples() []Sample {
	current := c.current.Load()
	if current == nil {
		return nil
	}
	samples := make([]Sample, 0, len(current.samples))
	for _, s := range current.samples {
		labels := make(map[string]string, len(s.values))
		for i, name := range s.gauge.LabelNames {
			labels[name] = s.values[i]
		}
		samples = append(samples, Sample{Name: s.gauge.Name, Labels: labels, Value: s.value, Timestamp: s.timestamp})
	}
	return samples
}
