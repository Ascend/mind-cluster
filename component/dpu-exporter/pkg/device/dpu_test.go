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

package device

import "testing"

// TestNewDPU covers construction and the zero-value metric accessors.
func TestNewDPU(t *testing.T) {
	d := NewDPU("hinic0", "CAL_2X400G_UB_EXP", "hinic0")
	if d.CardName != "hinic0" || d.CardType != "CAL_2X400G_UB_EXP" || d.HcaName != "hinic0" {
		t.Errorf("NewDPU fields = %+v", d)
	}
	if got := d.GetMetric("missing"); got != 0 {
		t.Errorf("GetMetric(missing) = %v, want 0", got)
	}
	if got := d.GetMetrics(); got == nil || len(got) != 0 {
		t.Errorf("GetMetrics() = %v, want empty non-nil map", got)
	}
	if got := d.GetEthList(); len(got) != 0 {
		t.Errorf("GetEthList() = %v, want empty", got)
	}
}

// TestDPUMetricsAccessors covers SetMetrics/GetMetrics/GetMetric semantics.
func TestDPUMetricsAccessors(t *testing.T) {
	d := NewDPU("hinic0", "CAL_2X400G_UB_EXP", "hinic0")

	d.SetMetrics(map[string]float64{"m1": 1, "m2": 2})
	if got := d.GetMetric("m1"); got != 1 {
		t.Errorf("GetMetric(m1) = %v, want 1", got)
	}
	if got := d.GetMetric("nope"); got != 0 {
		t.Errorf("GetMetric(nope) = %v, want 0", got)
	}

	// GetMetrics must return a copy: mutating it must not affect the DPU
	got := d.GetMetrics()
	got["m1"] = 999
	if v := d.GetMetric("m1"); v != 1 {
		t.Errorf("internal map affected by mutation of returned copy, m1=%v", v)
	}

	// SetMetrics replaces the whole map
	d.SetMetrics(map[string]float64{"m3": 3})
	if v := d.GetMetric("m1"); v != 0 {
		t.Errorf("m1 = %v after replace, want 0", v)
	}
	if v := d.GetMetric("m3"); v != 3 {
		t.Errorf("m3 = %v, want 3", v)
	}
}

// TestDPUGetEthList covers GetEthList ordering.
func TestDPUGetEthList(t *testing.T) {
	d := NewDPU("hinic0", "CAL_2X400G_UB_EXP", "hinic0")
	d.Interfaces = []Interface{
		NewInterface("ens1f0", "hinic0", "/sys/class/net/ens1f0"),
		NewInterface("ens1p1", "hinic0", "/sys/class/net/ens1p1"),
	}
	eth := d.GetEthList()
	if len(eth) != 2 || eth[0] != "ens1f0" || eth[1] != "ens1p1" {
		t.Errorf("GetEthList() = %v, want [ens1f0 ens1p1]", eth)
	}
}
