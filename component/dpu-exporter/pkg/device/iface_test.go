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

// TestNewInterface covers construction and the zero-value metric accessors.
func TestNewInterface(t *testing.T) {
	i := NewInterface("ens1f0", "hinic0", "/sys/class/net/ens1f0")
	if i.EthName != "ens1f0" || i.HcaName != "hinic0" || i.SysfsPath != "/sys/class/net/ens1f0" {
		t.Errorf("NewInterface fields = %+v", i)
	}
	if got := i.GetMetric("missing"); got != 0 {
		t.Errorf("GetMetric(missing) = %v, want 0", got)
	}
	if got := i.GetMetrics(); got == nil || len(got) != 0 {
		t.Errorf("GetMetrics() = %v, want empty non-nil map", got)
	}
}

// TestInterfaceMetricsAccessors covers SetMetrics/GetMetrics/GetMetric semantics.
func TestInterfaceMetricsAccessors(t *testing.T) {
	i := NewInterface("ens1f0", "hinic0", "/sys/class/net/ens1f0")

	i.SetMetrics(map[string]float64{"rx_packets": 10})
	if got := i.GetMetric("rx_packets"); got != 10 {
		t.Errorf("GetMetric(rx_packets) = %v, want 10", got)
	}
	if got := i.GetMetric("tx_packets"); got != 0 {
		t.Errorf("GetMetric(tx_packets) = %v, want 0", got)
	}

	// GetMetrics must return a copy: mutating it must not affect the Interface
	got := i.GetMetrics()
	got["rx_packets"] = 999
	if v := i.GetMetric("rx_packets"); v != 10 {
		t.Errorf("internal map affected by mutation of returned copy, rx_packets=%v", v)
	}

	// SetMetrics replaces the whole map
	i.SetMetrics(map[string]float64{"tx_packets": 5})
	if v := i.GetMetric("rx_packets"); v != 0 {
		t.Errorf("rx_packets = %v after replace, want 0", v)
	}
	if v := i.GetMetric("tx_packets"); v != 5 {
		t.Errorf("tx_packets = %v, want 5", v)
	}
}
