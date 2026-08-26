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
	"testing"

	"huawei.com/dpu-exporter/pkg/device"
)

func TestExtractMetricNameFromHeader(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"ROCE_ERR_CTR_SQA_NAK_PSN_ERR(ROCE|ERR|ERR):", "roce_err_ctr_sqa_nak_psn_err"},
		{"ROCE_DP_CTR_RR_ECN_RX(ROCE|DP|KEY):", "roce_dp_ctr_rr_ecn_rx"},
		{"(ROCE|ERR|ERR):", ""},  // no name before (
		{"NO_PARENS", ""},        // no parentheses
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
	dpuList    []device.DPU
	dpuMetrics map[string]map[string]float64
	ifaceMetrics map[string]map[string]float64
}

func newMockCache() *mockCacheAccessor {
	return &mockCacheAccessor{
		dpuMetrics:   make(map[string]map[string]float64),
		ifaceMetrics: make(map[string]map[string]float64),
	}
}

func (m *mockCacheAccessor) GetDpuList() []device.DPU          { return m.dpuList }
func (m *mockCacheAccessor) GetDpuMetrics(c string) map[string]float64 { return m.dpuMetrics[c] }
func (m *mockCacheAccessor) GetIfaceMetrics(e string) map[string]float64 { return m.ifaceMetrics[e] }
func (m *mockCacheAccessor) SetDpuMetrics(c string, v map[string]float64)  { m.dpuMetrics[c] = v }
func (m *mockCacheAccessor) SetIfaceMetrics(e string, v map[string]float64) { m.ifaceMetrics[e] = v }
func (m *mockCacheAccessor) SetDpuList(d []device.DPU)                     { m.dpuList = d }
