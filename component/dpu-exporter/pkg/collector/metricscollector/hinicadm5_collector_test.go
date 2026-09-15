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
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"huawei.com/dpu-exporter/pkg/configmanager"
	"huawei.com/dpu-exporter/pkg/device"
)

func TestExtractMetricNameFromHeader(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"ROCE_ERR_CTR_SQA_NAK_PSN_ERR(ROCE|ERR|ERR):", "roce_err_ctr_sqa_nak_psn_err"},
		{"ROCE_DP_CTR_RR_ECN_RX(ROCE|DP|KEY):", "roce_dp_ctr_rr_ecn_rx"},
		{"(ROCE|ERR|ERR):", ""}, // no name before (
		{"NO_PARENS", ""},       // no parentheses
	}
	for _, tt := range tests {
		got := extractMetricNameFromHeader(tt.input)
		if got != tt.expect {
			t.Errorf("extractMetricNameFromHeader(%q) = %q, want %q", tt.input, got, tt.expect)
		}
	}
}

func TestExtractMetricName(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"ID-0x1902::0000:ROCE_CMDQ_CTR_EXT_CMD:", "roce_cmdq_ctr_ext_cmd"},
		{"ID-0x1926::0000:ROCE_ERR_CTR_FOO:", "roce_err_ctr_foo"},
		{"NO_COLON", ""},
	}
	for _, tt := range tests {
		got := extractMetricName(tt.input)
		if got != tt.expect {
			t.Errorf("extractMetricName(%q) = %q, want %q", tt.input, got, tt.expect)
		}
	}
}

func TestParseHinicadm5CounterOutput_HeaderOnly(t *testing.T) {
	// Metric with only header line → value = 0
	output := "ROCE_ERR_CTR_SQA_NAK_PSN_ERR(ROCE|ERR|ERR):\n"
	metrics := parseHinicadm5CounterOutput(output)
	if v, ok := metrics["roce_err_ctr_sqa_nak_psn_err"]; !ok || v != 0 {
		t.Errorf("expected 0, got %v, ok=%v", v, ok)
	}
}

func TestParseHinicadm5CounterOutput_HeaderAndValue(t *testing.T) {
	// Header line sets 0, value line overwrites with 2
	output := `ROCE_CMDQ_CTR_ROCE_UPDATE_GID(ROCE|CMDQ|INFO):
ID-0x1926::0000:ROCE_CMDQ_CTR_ROCE_UPDATE_GID: 2
`
	metrics := parseHinicadm5CounterOutput(output)
	if v, ok := metrics["roce_cmdq_ctr_roce_update_gid"]; !ok || v != 2 {
		t.Errorf("expected 2, got %v, ok=%v", v, ok)
	}
}

func TestParseHinicadm5CounterOutput_Mixed(t *testing.T) {
	output := `Card num: 1
ROCE_ERR_CTR_SQA_NAK_PSN_ERR(ROCE|ERR|ERR):
ROCE_DP_CTR_RR_ECN_RX(ROCE|DP|KEY):
ID-0x1902::0000:ROCE_DP_CTR_RR_ECN_RX: 5
# comment line
`
	metrics := parseHinicadm5CounterOutput(output)
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}
	if metrics["roce_err_ctr_sqa_nak_psn_err"] != 0 {
		t.Error("header-only metric should be 0")
	}
	if metrics["roce_dp_ctr_rr_ecn_rx"] != 5 {
		t.Error("value line should overwrite to 5")
	}
}

func TestParseHinicadm5CounterOutput_Empty(t *testing.T) {
	metrics := parseHinicadm5CounterOutput("")
	if len(metrics) != 0 {
		t.Errorf("expected empty, got %d", len(metrics))
	}
}

// --- mock for CollectorContext ---

type mockCacheAccessor struct {
	dpuList      []device.DPU
	dpuMetrics   map[string]map[string]float64
	ifaceMetrics map[string]map[string]float64
}

func newMockCache() *mockCacheAccessor {
	return &mockCacheAccessor{
		dpuMetrics:   make(map[string]map[string]float64),
		ifaceMetrics: make(map[string]map[string]float64),
	}
}

func (m *mockCacheAccessor) GetDpuList() []device.DPU                       { return m.dpuList }
func (m *mockCacheAccessor) GetDpuMetrics(c string) map[string]float64      { return m.dpuMetrics[c] }
func (m *mockCacheAccessor) GetIfaceMetrics(e string) map[string]float64    { return m.ifaceMetrics[e] }
func (m *mockCacheAccessor) SetDpuMetrics(c string, v map[string]float64)   { m.dpuMetrics[c] = v }
func (m *mockCacheAccessor) SetIfaceMetrics(e string, v map[string]float64) { m.ifaceMetrics[e] = v }
func (m *mockCacheAccessor) SetDpuList(d []device.DPU)                      { m.dpuList = d }

// collectorMockDmgr implements device.DeviceManager with injectable behavior
// for exercising collector collection paths.
type collectorMockDmgr struct {
	cardType string
	dpuList  []device.DPU

	execOutput string
	execErr    error
	execCalls  int

	readVals map[string]string
	readErrs map[string]error

	listFiles []string
	listErr   error
}

func (m *collectorMockDmgr) AutoInit() error                  { return nil }
func (m *collectorMockDmgr) GetDpuList() []device.DPU         { return m.dpuList }
func (m *collectorMockDmgr) GetCardType() string              { return m.cardType }
func (m *collectorMockDmgr) ListDir(string) ([]string, error) { return m.listFiles, m.listErr }

func (m *collectorMockDmgr) ExecCommand(_ ...string) (string, error) {
	m.execCalls++
	return m.execOutput, m.execErr
}

func (m *collectorMockDmgr) ReadSysfs(path string) (string, error) {
	if err, ok := m.readErrs[path]; ok {
		return "", err
	}
	if v, ok := m.readVals[path]; ok {
		return v, nil
	}
	return "", errors.New("no such sysfs file: " + path)
}

// fakeHinicadm5CounterOutput mimics hinicadm5 counter output with one
// whitelisted metric (roce_dp_ctr_rr_ecn_rx) and several non-whitelisted ones.
const fakeHinicadm5CounterOutput = `Card num:1
Device Information:
ROCE_ERR_CTR_SQA_NAK_PSN_ERR(ROCE|ERR|ERR):
ID-0x1902::0000:ROCE_ERR_CTR_SQA_NAK_PSN_ERR: 9
ID-0x1903::0000:ROCE_DP_CTR_RR_ECN_RX: 5
ROCE_CMDQ_CTR_ROCE_UPDATE_GID(ROCE|CMDQ|INFO):
ID-0x1904::0000:ROCE_CMDQ_CTR_ROCE_UPDATE_GID: 1`

// TestHinicadm5Collector_CollectToCache_NoDpu verifies no command is executed
// when the cache contains no DPU devices.
func TestHinicadm5Collector_CollectToCache_NoDpu(t *testing.T) {
	dmgr := &collectorMockDmgr{}
	ctx := &mockCollectorContext{dmgr: dmgr, cache: newMockCache()}
	(&Hinicadm5Collector{}).CollectToCache(ctx)
	if dmgr.execCalls != 0 {
		t.Errorf("ExecCommand called %d times with empty dpu list, want 0", dmgr.execCalls)
	}
}

// TestHinicadm5Collector_CollectToCache_Success covers the happy path:
// per-card collection, whitelist filtering and cache writes.
func TestHinicadm5Collector_CollectToCache_Success(t *testing.T) {
	configmanager.GetWhitelist().LoadCustom([]string{"roce_dp_ctr_rr_ecn_rx"})
	defer configmanager.GetWhitelist().LoadCustom(nil)

	dmgr := &collectorMockDmgr{
		cardType:   device.CardTypeHuawei,
		execOutput: fakeHinicadm5CounterOutput,
	}
	cache := newMockCache()
	cache.dpuList = []device.DPU{{CardName: "hinic0", CardType: "CAL_2X400G_UB_EXP"}}
	ctx := &mockCollectorContext{dmgr: dmgr, cache: cache}

	c := &Hinicadm5Collector{}
	c.CollectToCache(ctx)

	if dmgr.execCalls != 1 {
		t.Fatalf("ExecCommand called %d times, want 1", dmgr.execCalls)
	}
	got := cache.dpuMetrics["hinic0"]
	if v, ok := got["roce_dp_ctr_rr_ecn_rx"]; !ok || v != 5 {
		t.Errorf("roce_dp_ctr_rr_ecn_rx = (%v, %v), want (5, true)", v, ok)
	}
	if len(got) != 1 {
		t.Errorf("filtered metrics = %v, want only the whitelisted one", got)
	}
	if _, ok := c.localCache.Load("hinic0"); !ok {
		t.Error("localCache does not contain hinic0 metrics")
	}
}

// TestHinicadm5Collector_CollectToCache_Errors covers exec failure and
// empty-parse-result branches.
func TestHinicadm5Collector_CollectToCache_Errors(t *testing.T) {
	// exec failure → nothing written to cache
	dmgr := &collectorMockDmgr{execErr: errors.New("hinicadm5 not available")}
	cache := newMockCache()
	cache.dpuList = []device.DPU{{CardName: "hinic0", CardType: "huawei"}}
	(&Hinicadm5Collector{}).CollectToCache(&mockCollectorContext{dmgr: dmgr, cache: cache})
	if _, ok := cache.dpuMetrics["hinic0"]; ok {
		t.Error("metrics written to cache despite exec failure")
	}

	// output without metric lines → nothing written to cache
	dmgr2 := &collectorMockDmgr{execOutput: "Card num:1\nDevice Information:\n"}
	cache2 := newMockCache()
	cache2.dpuList = []device.DPU{{CardName: "hinic0", CardType: "huawei"}}
	(&Hinicadm5Collector{}).CollectToCache(&mockCollectorContext{dmgr: dmgr2, cache: cache2})
	if _, ok := cache2.dpuMetrics["hinic0"]; ok {
		t.Error("metrics written to cache despite empty parse result")
	}
}

// TestHinicadm5Collector_Describe verifies the metric descriptor.
func TestHinicadm5Collector_Describe(t *testing.T) {
	ch := make(chan *prometheus.Desc, 1)
	(&Hinicadm5Collector{}).Describe(ch)
	close(ch)
	for desc := range ch {
		if !strings.Contains(desc.String(), "dpu_hinicadm5_metric") {
			t.Errorf("desc = %q, want it to contain dpu_hinicadm5_metric", desc.String())
		}
	}
}

// TestHinicadm5Collector_UpdatePrometheus verifies cached metrics are emitted
// with one Prometheus metric per metric name.
func TestHinicadm5Collector_UpdatePrometheus(t *testing.T) {
	cache := newMockCache()
	cache.dpuList = []device.DPU{{CardName: "hinic0", CardType: "CAL_2X400G_UB_EXP"}}
	cache.dpuMetrics["hinic0"] = map[string]float64{"roce_err_ctr_x": 3, "roce_dp_ctr_y": 7}
	ctx := &mockCollectorContext{cache: cache}

	ch := make(chan prometheus.Metric, 4)
	(&Hinicadm5Collector{}).UpdatePrometheus(ch, ctx)
	close(ch)

	count := 0
	for m := range ch {
		_ = m.Desc().String()
		count++
	}
	if count != 2 {
		t.Errorf("emitted %d metrics, want 2", count)
	}
}

// TestHinicadm5Collector_IsSupportedAndGetInterval covers IsSupported for
// matching/mismatching card types and the configurable collection interval.
func TestHinicadm5Collector_IsSupportedAndGetInterval(t *testing.T) {
	c := &Hinicadm5Collector{}
	if !c.IsSupported(&mockCollectorContext{dmgr: &collectorMockDmgr{cardType: device.CardTypeHuawei}}) {
		t.Error("IsSupported(huawei) = false, want true")
	}
	if c.IsSupported(&mockCollectorContext{dmgr: &collectorMockDmgr{cardType: "other"}}) {
		t.Error("IsSupported(other) = true, want false")
	}

	configmanager.SetCollectorInterval(configmanager.CacheKeyHinicadm5, 33*time.Second)
	defer configmanager.SetCollectorInterval(configmanager.CacheKeyHinicadm5, configmanager.DefaultGroupInterval)
	if got := c.GetInterval(); got != 33*time.Second {
		t.Errorf("GetInterval() = %v, want 33s", got)
	}
}
