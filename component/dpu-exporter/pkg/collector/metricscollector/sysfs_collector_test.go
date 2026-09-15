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
	"path/filepath"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"huawei.com/dpu-exporter/pkg/configmanager"
	"huawei.com/dpu-exporter/pkg/device"
	"huawei.com/dpu-exporter/utils/logger"
)

func init() {
	// Initialize logger to prevent nil pointer in logger.Warn during tests
	_ = logger.InitLogger()
}

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
		"carrier":         1,
		"operstate":       1,
		"carrier_changes": 5,
		"rx_packets":      100,
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

// sysfsIfacePath builds <sysfsNetBase>/<eth>/<parts...>
func sysfsIfacePath(eth string, parts ...string) string {
	return filepath.Join(append([]string{sysfsNetBase, eth}, parts...)...)
}

// TestSysfsCollector_CollectToCache_NoDpu verifies the empty-list early return.
func TestSysfsCollector_CollectToCache_NoDpu(t *testing.T) {
	dmgr := &collectorMockDmgr{}
	(&SysfsCollector{}).CollectToCache(&mockCollectorContext{dmgr: dmgr, cache: newMockCache()})
}

// TestSysfsCollector_CollectToCache_WithInterface covers the happy path:
// iface-level files, statistics discovery and writes to the global cache.
func TestSysfsCollector_CollectToCache_WithInterface(t *testing.T) {
	const eth0 = "eth0"
	dmgr := &collectorMockDmgr{
		cardType: device.CardTypeHuawei,
		readVals: map[string]string{
			sysfsIfacePath(eth0, "carrier"):                  "1",
			sysfsIfacePath(eth0, "carrier_changes"):          "5",
			sysfsIfacePath(eth0, "operstate"):                "up",
			sysfsIfacePath(eth0, "statistics", "rx_packets"): "100",
			sysfsIfacePath(eth0, "statistics", "bad_stat"):   "not-a-number",
		},
		listFiles: []string{"rx_packets", "bad_stat"},
	}

	gCache := newMockCache()
	SetGlobalContext(&mockCollectorContext{dmgr: dmgr, cache: gCache})
	defer SetGlobalContext(nil)

	cache := newMockCache()
	cache.dpuList = []device.DPU{{
		CardName:   "hinic0",
		CardType:   "huawei",
		Interfaces: []device.Interface{{EthName: eth0}},
	}}
	c := &SysfsCollector{}
	c.CollectToCache(&mockCollectorContext{dmgr: dmgr, cache: cache})

	got := gCache.ifaceMetrics[eth0]
	if got["carrier"] != 1 || got["carrier_changes"] != 5 || got["operstate"] != 1 {
		t.Errorf("iface-level metrics = %v, want carrier=1 carrier_changes=5 operstate=1", got)
	}
	if got["rx_packets"] != 100 {
		t.Errorf("rx_packets = %v, want 100", got["rx_packets"])
	}
	if _, ok := got["bad_stat"]; ok {
		t.Error("unparseable statistics value should be skipped")
	}
	// local cache also holds the collected metrics
	if _, ok := c.localCache.Load(eth0); !ok {
		t.Error("localCache does not contain eth0 metrics")
	}
}

// TestSysfsCollector_CollectForInterface_ErrorPaths covers the carrier-read
// failure (maps to -1) and the statistics ListDir failure branches.
func TestSysfsCollector_CollectForInterface_ErrorPaths(t *testing.T) {
	const eth0 = "eth0"
	dmgr := &collectorMockDmgr{
		readErrs: map[string]error{
			sysfsIfacePath(eth0, "carrier"): errors.New("interface down"),
		},
		listErr: errors.New("statistics dir not available"),
	}

	gCache := newMockCache()
	SetGlobalContext(&mockCollectorContext{dmgr: dmgr, cache: gCache})
	defer SetGlobalContext(nil)

	(&SysfsCollector{}).collectForInterface(dmgr, device.Interface{EthName: eth0})

	got := gCache.ifaceMetrics[eth0]
	if v, ok := got["carrier"]; !ok || v != -1 {
		t.Errorf("carrier = (%v, %v), want (-1, true) on read error", v, ok)
	}
	if _, ok := got["rx_packets"]; ok {
		t.Error("statistics should be absent when ListDir fails")
	}
}

// TestSysfsCollector_GetInterval verifies the configured collection interval.
func TestSysfsCollector_GetInterval(t *testing.T) {
	configmanager.SetCollectorInterval(configmanager.CacheKeySysfs, 21*time.Second)
	defer configmanager.SetCollectorInterval(configmanager.CacheKeySysfs, configmanager.DefaultGroupInterval)
	if got := (&SysfsCollector{}).GetInterval(); got != 21*time.Second {
		t.Errorf("GetInterval() = %v, want 21s", got)
	}
}
