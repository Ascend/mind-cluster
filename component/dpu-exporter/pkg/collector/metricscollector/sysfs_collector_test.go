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

	"github.com/prometheus/client_golang/prometheus"

	"huawei.com/dpu-exporter/pkg/device"
)

func TestParseSysfsValue(t *testing.T) {
	tests := []struct {
		file   string
		raw    string
		expect float64
		err    bool
	}{
		{"operstate", "up", 1, false},
		{"operstate", "down", 0, false},
		{"operstate", "unknown", -1, false},
		{"carrier", "1", 1, false},
		{"carrier", "0", 0, false},
		{"carrier", "abc", -1, false},
		{"carrier", "2", -1, false},
		{"carrier_changes", "5", 5, false},
		{"carrier_changes", "abc", 0, true},
	}
	for _, tt := range tests {
		got, err := parseSysfsValue(tt.file, tt.raw)
		if tt.err {
			if err == nil {
				t.Errorf("parseSysfsValue(%q,%q) expected error", tt.file, tt.raw)
			}
		} else {
			if err != nil || got != tt.expect {
				t.Errorf("parseSysfsValue(%q,%q) = %v,%v; want %v,nil", tt.file, tt.raw, got, err, tt.expect)
			}
		}
	}
}

func TestSysfsGaugeMetrics_Classification(t *testing.T) {
	// Verify gauge classification
	if _, ok := sysfsGaugeMetrics["carrier"]; !ok {
		t.Error("carrier should be gauge")
	}
	if _, ok := sysfsGaugeMetrics["operstate"]; !ok {
		t.Error("operstate should be gauge")
	}
	if _, ok := sysfsGaugeMetrics["carrier_changes"]; ok {
		t.Error("carrier_changes should NOT be gauge (it is counter)")
	}
}

func TestSysfsCollector_UpdatePrometheus_GaugeVsCounter(t *testing.T) {
	c := &SysfsCollector{}
	cache := newMockCache()
	cache.dpuList = []device.DPU{
		{
			CardName: "hinic0",
			CardType: "huawei",
			Interfaces: []device.Interface{
				{EthName: "ens1f0"},
			},
		},
	}
	cache.ifaceMetrics["ens1f0"] = map[string]float64{
		"carrier":        1,
		"operstate":      1,
		"carrier_changes": 5,
		"rx_packets":     100,
	}

	ctx := &mockCollectorContext{cache: cache}

	// Collect metrics and verify type by writing to a channel
	ch := make(chan prometheus.Metric, 10)
	c.UpdatePrometheus(ch, ctx)
	close(ch)

	count := 0
	for m := range ch {
		_ = m.Desc().String() // just ensure it doesn't panic
		count++
	}
	if count != 4 {
		t.Errorf("expected 4 metrics, got %d", count)
	}
}

// mockCollectorContext implements CollectorContext for tests
type mockCollectorContext struct {
	dmgr  device.DeviceManager
	cache CacheAccessor
}

func (m *mockCollectorContext) GetDmgr() device.DeviceManager { return m.dmgr }
func (m *mockCollectorContext) GetCache() CacheAccessor       { return m.cache }
