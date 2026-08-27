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

package dpucollector

import (
	"sync"
	"testing"

	"huawei.com/dpu-exporter/pkg/device"
)

// TestNewDpuCache verifies the cache is created with initialized internal state.
func TestNewDpuCache(t *testing.T) {
	c := NewDpuCache()
	if c == nil {
		t.Fatal("NewDpuCache returned nil")
	}
	if c.dpuData == nil || c.dpuData.CardMetrics == nil {
		t.Error("dpuData/CardMetrics not initialized")
	}
	if c.ifaceData == nil || c.ifaceData.IfaceMetrics == nil {
		t.Error("ifaceData/IfaceMetrics not initialized")
	}
	// fresh cache: dpuList is nil, GetDpuList must return non-nil empty slice
	if got := c.GetDpuList(); got == nil || len(got) != 0 {
		t.Errorf("GetDpuList() = %v, want empty non-nil slice", got)
	}
}

// TestDpuListSetGet covers SetDpuList/GetDpuList round-trip.
func TestDpuListSetGet(t *testing.T) {
	c := NewDpuCache()
	list := []device.DPU{
		{CardName: "hinic0", CardType: "huawei"},
		{CardName: "hinic1", CardType: "huawei"},
	}
	c.SetDpuList(list)
	got := c.GetDpuList()
	if len(got) != 2 || got[0].CardName != "hinic0" || got[1].CardName != "hinic1" {
		t.Errorf("GetDpuList() = %v, want %v", got, list)
	}
}

// TestDpuMetricsSetGet covers GetDpuMetrics miss/hit/copy/merge semantics.
func TestDpuMetricsSetGet(t *testing.T) {
	c := NewDpuCache()

	// miss: empty non-nil map
	if got := c.GetDpuMetrics("hinic0"); got == nil || len(got) != 0 {
		t.Errorf("GetDpuMetrics(miss) = %v, want empty non-nil map", got)
	}

	// hit + returned map is a copy
	c.SetDpuMetrics("hinic0", map[string]float64{"m1": 1})
	got := c.GetDpuMetrics("hinic0")
	if got["m1"] != 1 {
		t.Errorf("GetDpuMetrics(hit) = %v, want m1=1", got)
	}
	got["m1"] = 999 // mutate copy only
	if v := c.GetDpuMetrics("hinic0")["m1"]; v != 1 {
		t.Errorf("internal map affected by mutation of returned copy, m1=%v", v)
	}

	// merge: second Set merges into existing map
	c.SetDpuMetrics("hinic0", map[string]float64{"m2": 2})
	got = c.GetDpuMetrics("hinic0")
	if len(got) != 2 || got["m1"] != 1 || got["m2"] != 2 {
		t.Errorf("SetDpuMetrics merge failed, got %v, want m1=1,m2=2", got)
	}
}

// TestIfaceMetricsSetGet covers GetIfaceMetrics miss/hit/copy/merge semantics.
func TestIfaceMetricsSetGet(t *testing.T) {
	c := NewDpuCache()

	// miss: empty non-nil map
	if got := c.GetIfaceMetrics("eth0"); got == nil || len(got) != 0 {
		t.Errorf("GetIfaceMetrics(miss) = %v, want empty non-nil map", got)
	}

	// hit + returned map is a copy
	c.SetIfaceMetrics("eth0", map[string]float64{"carrier": 1})
	got := c.GetIfaceMetrics("eth0")
	if got["carrier"] != 1 {
		t.Errorf("GetIfaceMetrics(hit) = %v, want carrier=1", got)
	}
	got["carrier"] = 999
	if v := c.GetIfaceMetrics("eth0")["carrier"]; v != 1 {
		t.Errorf("internal map affected by mutation of returned copy, carrier=%v", v)
	}

	// merge: second Set merges into existing map
	c.SetIfaceMetrics("eth0", map[string]float64{"rx_packets": 100})
	got = c.GetIfaceMetrics("eth0")
	if len(got) != 2 || got["carrier"] != 1 || got["rx_packets"] != 100 {
		t.Errorf("SetIfaceMetrics merge failed, got %v, want carrier=1,rx_packets=100", got)
	}
}

// TestDpuCacheConcurrency exercises concurrent readers/writers (run with -race).
func TestDpuCacheConcurrency(t *testing.T) {
	c := NewDpuCache()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				c.SetDpuList([]device.DPU{{CardName: "hinic0"}})
				_ = c.GetDpuList()
				c.SetDpuMetrics("hinic0", map[string]float64{"m": 1})
				_ = c.GetDpuMetrics("hinic0")
				c.SetIfaceMetrics("eth0", map[string]float64{"m": 1})
				_ = c.GetIfaceMetrics("eth0")
			}
		}()
	}
	wg.Wait()
}
