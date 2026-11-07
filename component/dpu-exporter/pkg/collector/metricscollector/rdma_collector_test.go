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
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"huawei.com/dpu-exporter/pkg/device"
)

type rdmaMockCache struct {
	CacheAccessor
	rdma map[string]map[string]RdmaMetricValue
}

func newRdmaMockCache() *rdmaMockCache {
	return &rdmaMockCache{rdma: make(map[string]map[string]RdmaMetricValue)}
}
func (m *rdmaMockCache) GetAllRdmaMetrics() map[string]map[string]RdmaMetricValue    { return m.rdma }
func (m *rdmaMockCache) SetAllRdmaMetrics(all map[string]map[string]RdmaMetricValue) { m.rdma = all }

type rdmaMockDmgr struct {
	device.DeviceManager
	output string
	err    error
	args   []string
	calls  int
}

func (m *rdmaMockDmgr) ExecRdma(args ...string) (string, error) {
	m.calls++
	m.args = args
	return m.output, m.err
}

type rdmaCtx struct {
	dmgr  device.DeviceManager
	cache CacheAccessor
}

func (c *rdmaCtx) GetDmgr() device.DeviceManager { return c.dmgr }
func (c *rdmaCtx) GetCache() CacheAccessor       { return c.cache }

type rdmaRegistryCollector struct {
	inner *RdmaCollector
	ctx   CollectorContext
}

func (c *rdmaRegistryCollector) Describe(ch chan<- *prometheus.Desc) { c.inner.Describe(ch) }
func (c *rdmaRegistryCollector) Collect(ch chan<- prometheus.Metric) {
	c.inner.UpdatePrometheus(ch, c.ctx)
}
func newRdmaCtx(output string) (*rdmaMockCache, *rdmaMockDmgr, *rdmaCtx) {
	cache := newRdmaMockCache()
	dmgr := &rdmaMockDmgr{output: output}
	return cache, dmgr, &rdmaCtx{dmgr: dmgr, cache: cache}
}

const rdmaAllOutput = "0: hinic0: pd 1 cq 2 qp 3 cm_id 4 mr 5 ctx 6 srq 7\n"

func rdmaDeviceMetrics(cache *rdmaMockCache, dev string) map[string]RdmaMetricValue {
	return cache.GetAllRdmaMetrics()[dev]
}
func rdmaAllMetrics(value float64) map[string]RdmaMetricValue {
	out := make(map[string]RdmaMetricValue, len(rdmaMetricDefs))
	for _, def := range rdmaMetricDefs {
		out[def.resource] = RdmaMetricValue{Value: value}
	}
	return out
}
func checkRdmaMetric(t *testing.T, metrics map[string]RdmaMetricValue, key string, value float64, miss int) {
	got, ok := metrics[key]
	if !ok || got.MissCount != miss || got.Value != value {
		t.Fatalf("%s = %+v, want value %v / miss %d", key, got, value, miss)
	}
}
func checkRdmaMetricAbsent(t *testing.T, metrics map[string]RdmaMetricValue, key string) {
	if got, ok := metrics[key]; ok {
		t.Fatalf("%s = %+v, want metric removed", key, got)
	}
}
func TestRdmaDefinitionsAndParse(t *testing.T) {
	for _, def := range rdmaMetricDefs {
		if def.resource == "" || def.metricName == "" || def.desc == nil ||
			rdmaResourceWhitelist[def.resource] != def.metricName {
			t.Errorf("invalid metric definition: %+v", def)
		}
	}
	all := map[string]float64{"pd": 2, "cq": 4, "qp": 4, "cm_id": 0, "mr": 2, "ctx": 1, "srq": 0}
	for _, tc := range []struct {
		name, output string
		want         map[string]map[string]float64
	}{
		{"empty", "", map[string]map[string]float64{}},
		{"fields", "0: hinic0: srq 0/10 mr 2/10 cm_id 0 qp 4/10 cq 4/10 pd 2/10 ctx 1\n" +
			"1: hinic1: pd 3 cq 8 qp 8 cm_id 0 mr 6 ctx 1 srq 0 comp_cntr 9\n" +
			"2: hinic2: qp 7 cq bad\nbad line\n", map[string]map[string]float64{
			"hinic0": all,
			"hinic1": {"pd": 3, "cq": 8, "qp": 8, "cm_id": 0, "mr": 6, "ctx": 1, "srq": 0},
			"hinic2": {"qp": 7},
		}},
		{"unknown", "0: dev0: comp_cntr 3\n", map[string]map[string]float64{"dev0": {}}},
		{"invalid", "0: dev: qp 4/10 cq bad\n", map[string]map[string]float64{"dev": {"qp": 4}}},
	} {
		if got := parseRdmaResourceSummary(tc.output); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: parse = %v, want %v", tc.name, got, tc.want)
		}
	}
}
func TestRdmaCollectAgingAndRecovery(t *testing.T) {
	cache, dmgr, ctx := newRdmaCtx(rdmaAllOutput)
	c := &RdmaCollector{}
	c.CollectToCache(ctx)
	if dmgr.calls != 1 || strings.Join(dmgr.args, " ") != "resource show" {
		t.Fatalf("ExecRdma calls/args = %d/%v", dmgr.calls, dmgr.args)
	}
	checkRdmaMetric(t, rdmaDeviceMetrics(cache, "hinic0"), "qp", 3, 0)
	dmgr.err = errors.New("rdma failed")
	for round := 1; round <= 2; round++ {
		c.CollectToCache(ctx)
		checkRdmaMetric(t, rdmaDeviceMetrics(cache, "hinic0"), "qp", 3, round)
	}
	dmgr.err = nil
	c.CollectToCache(ctx)
	checkRdmaMetric(t, rdmaDeviceMetrics(cache, "hinic0"), "qp", 3, 0)
	dmgr.output = ""
	for round := 1; round <= 2; round++ {
		c.CollectToCache(ctx)
		checkRdmaMetric(t, rdmaDeviceMetrics(cache, "hinic0"), "qp", 3, round)
	}
	dmgr.output = rdmaAllOutput
	c.CollectToCache(ctx)
	checkRdmaMetric(t, rdmaDeviceMetrics(cache, "hinic0"), "cq", 2, 0)
	dmgr.output = ""
	for round := 1; round <= 2; round++ {
		c.CollectToCache(ctx)
		checkRdmaMetric(t, rdmaDeviceMetrics(cache, "hinic0"), "qp", 3, round)
	}
	c.CollectToCache(ctx)
	if _, ok := cache.GetAllRdmaMetrics()["hinic0"]; ok {
		t.Fatal("hinic0 still cached after three missing rounds")
	}
	dmgr.output = "0: hinic0: qp 9\n"
	c.CollectToCache(ctx)
	metrics := rdmaDeviceMetrics(cache, "hinic0")
	if len(metrics) != 1 {
		t.Fatalf("partial initial metric count = %d, want 1", len(metrics))
	}
	checkRdmaMetric(t, metrics, "qp", 9, 0)
	checkRdmaMetricAbsent(t, metrics, "cq")
}
func TestRdmaMetricMissingAndRecovery(t *testing.T) {
	cache, dmgr, ctx := newRdmaCtx(rdmaAllOutput)
	c := &RdmaCollector{}
	c.CollectToCache(ctx)
	dmgr.output = "0: hinic0: qp 10\n"
	for round := 1; round <= 2; round++ {
		c.CollectToCache(ctx)
		checkRdmaMetric(t, rdmaDeviceMetrics(cache, "hinic0"), "cq", 2, round)
	}
	c.CollectToCache(ctx)
	metrics := rdmaDeviceMetrics(cache, "hinic0")
	if len(metrics) != 1 {
		t.Fatalf("metric count after removals = %d, want 1", len(metrics))
	}
	checkRdmaMetric(t, metrics, "qp", 10, 0)
	checkRdmaMetricAbsent(t, metrics, "cq")
	dmgr.output = "0: hinic0: qp 10 cq 20\n"
	c.CollectToCache(ctx)
	checkRdmaMetric(t, rdmaDeviceMetrics(cache, "hinic0"), "cq", 20, 0)
	dmgr.output = "0: hinic0: qp 10\n"
	c.CollectToCache(ctx)
	checkRdmaMetric(t, rdmaDeviceMetrics(cache, "hinic0"), "cq", 20, 1)
}

func TestRdmaPrometheus(t *testing.T) {
	cache := newRdmaMockCache()
	cache.SetAllRdmaMetrics(map[string]map[string]RdmaMetricValue{
		"hinic0": rdmaAllMetrics(3),
		"hinic1": rdmaAllMetrics(2),
	})
	reg := prometheus.NewRegistry()
	reg.MustRegister(&rdmaRegistryCollector{inner: &RdmaCollector{}, ctx: &rdmaCtx{cache: cache}})
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	if len(families) != len(rdmaMetricDefs) {
		t.Fatalf("families = %d, want %d", len(families), len(rdmaMetricDefs))
	}
	for _, family := range families {
		if family.GetHelp() == "" {
			t.Errorf("missing help for %q", family.GetName())
		}
		for _, metric := range family.GetMetric() {
			labels, gauge := metric.GetLabel(), metric.GetGauge()
			if gauge == nil || len(labels) != 1 || labels[0].GetName() != RdmaDeviceLabelName {
				t.Errorf("%s metric = %v, gauge=%v", family.GetName(), labels, gauge)
			}
		}
	}
	if (&RdmaCollector{}).GetInterval() == 0 {
		t.Error("GetInterval() = 0")
	}
}
