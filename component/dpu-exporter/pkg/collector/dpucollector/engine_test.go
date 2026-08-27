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
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"huawei.com/dpu-exporter/pkg/collector/metricscollector"
	"huawei.com/dpu-exporter/pkg/configmanager"
	"huawei.com/dpu-exporter/pkg/device"
	"huawei.com/dpu-exporter/utils/logger"
)

func init() {
	// Initialize logger to prevent nil pointer in logger.Infof during tests
	_ = logger.InitLogger()
}

// mockCollectorCacheKey is the cache key derived from the mockCollector type name.
const mockCollectorCacheKey = "mockCollector"

// mockDeviceManager implements device.DeviceManager for tests.
type mockDeviceManager struct {
	list      []device.DPU
	cardType  string
	callCount int32 // number of GetDpuList calls
}

func (m *mockDeviceManager) AutoInit() error { return nil }
func (m *mockDeviceManager) GetDpuList() []device.DPU {
	atomic.AddInt32(&m.callCount, 1)
	return m.list
}
func (m *mockDeviceManager) ExecCommand(...string) (string, error) { return "", nil }
func (m *mockDeviceManager) ReadSysfs(string) (string, error)      { return "", nil }
func (m *mockDeviceManager) ListDir(string) ([]string, error)      { return nil, nil }
func (m *mockDeviceManager) GetCardType() string                   { return m.cardType }

// mockCollector implements metricscollector.MetricsCollector for tests.
type mockCollector struct {
	supported bool
	collects  int32
	preCnt    int32
	postCnt   int32
}

func (m *mockCollector) Describe(ch chan<- *prometheus.Desc) {}
func (m *mockCollector) CollectToCache(ctx metricscollector.CollectorContext) {
	atomic.AddInt32(&m.collects, 1)
}
func (m *mockCollector) UpdatePrometheus(ch chan<- prometheus.Metric, ctx metricscollector.CollectorContext) {
}
func (m *mockCollector) PreCollect(ctx metricscollector.CollectorContext) {
	atomic.AddInt32(&m.preCnt, 1)
}
func (m *mockCollector) PostCollect(ctx metricscollector.CollectorContext) {
	atomic.AddInt32(&m.postCnt, 1)
}
func (m *mockCollector) IsSupported(ctx metricscollector.CollectorContext) bool {
	return m.supported
}
func (m *mockCollector) GetInterval() time.Duration { return 0 }

// waitUntil polls cond until it returns true or timeout elapses.
func waitUntil(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return cond()
}

// waitWg waits for the group with a timeout, failing the test on timeout.
func waitWg(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatal("wait group did not finish in time")
	}
}

// resetMockCollectorInterval restores the default interval config for mockCollector.
func resetMockCollectorInterval() {
	configmanager.SetCollectorInterval(mockCollectorCacheKey, configmanager.DefaultGroupInterval)
}

// TestNewDpuCollector covers constructor and accessors.
func TestNewDpuCollector(t *testing.T) {
	dmgr := &mockDeviceManager{cardType: "huawei"}
	n := NewDpuCollector(dmgr)

	if n == nil {
		t.Fatal("NewDpuCollector returned nil")
	}
	if Collector != n {
		t.Error("global Collector not set to the new instance")
	}
	if n.GetDmgr() != dmgr {
		t.Error("GetDmgr() does not return the device manager passed in")
	}
	if n.GetCache() == nil {
		t.Error("GetCache() returned nil")
	}
	if got := n.GetDpuList(); len(got) != 0 {
		t.Errorf("GetDpuList() = %v, want empty", got)
	}
	if got := n.GetDpuMetrics("hinic0"); len(got) != 0 {
		t.Errorf("GetDpuMetrics() = %v, want empty", got)
	}
	if got := n.GetIfaceMetrics("eth0"); len(got) != 0 {
		t.Errorf("GetIfaceMetrics() = %v, want empty", got)
	}
}

// TestSetChainsAndGetChainsSnapshot covers chain set/get and snapshot isolation.
func TestSetChainsAndGetChainsSnapshot(t *testing.T) {
	defer SetChains(nil, nil)

	c1 := &mockCollector{supported: true}
	SetChains([]metricscollector.MetricsCollector{c1}, []metricscollector.MetricsCollector{c1})

	dpu, iface := GetChainsSnapshot()
	if len(dpu) != 1 || dpu[0] != c1 {
		t.Errorf("dpu chain snapshot = %v, want [%v]", dpu, c1)
	}
	if len(iface) != 1 || iface[0] != c1 {
		t.Errorf("iface chain snapshot = %v, want [%v]", iface, c1)
	}

	// snapshot must be a copy: mutating it must not affect the global chains
	dpu = append(dpu, &mockCollector{})
	dpu2, _ := GetChainsSnapshot()
	if len(dpu2) != 1 {
		t.Errorf("global dpu chain affected by snapshot mutation, len=%d", len(dpu2))
	}

	// empty chains
	SetChains(nil, nil)
	dpu, iface = GetChainsSnapshot()
	if len(dpu) != 0 || len(iface) != 0 {
		t.Errorf("empty chains snapshot = %d,%d, want 0,0", len(dpu), len(iface))
	}
}

// TestFilterSupported covers nil and unsupported collector filtering.
func TestFilterSupported(t *testing.T) {
	n := NewDpuCollector(&mockDeviceManager{})
	c1 := &mockCollector{supported: true}
	c2 := &mockCollector{supported: false}

	got := filterSupported([]metricscollector.MetricsCollector{nil, c1, c2}, n)
	if len(got) != 1 || got[0] != c1 {
		t.Errorf("filterSupported() = %v, want [%v]", got, c1)
	}
}

// TestReloadConfig verifies ReloadConfig notifies reload subscribers.
func TestReloadConfig(t *testing.T) {
	ch := configmanager.SubscribeReload()
	defer configmanager.UnsubscribeReload(ch)

	ReloadConfig()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("ReloadConfig did not signal the reload subscriber")
	}
}

// TestInitDpuList covers initial refresh, hot-reload interval update, and stop.
func TestInitDpuList(t *testing.T) {
	defer configmanager.SetCollectorInterval(configmanager.CacheKeyDpuListRefresh,
		configmanager.DefaultGroupInterval)

	dmgr := &mockDeviceManager{
		list:     []device.DPU{{CardName: "hinic0", CardType: "huawei"}},
		cardType: "huawei",
	}
	n := NewDpuCollector(dmgr)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	InitDpuList(&wg, ctx, n)

	// initial refresh happens immediately
	if !waitUntil(2*time.Second, func() bool { return len(n.GetDpuList()) == 1 }) {
		t.Fatalf("initial dpu list refresh did not happen, list=%v", n.GetDpuList())
	}
	firstCalls := atomic.LoadInt32(&dmgr.callCount)

	// hot-reload with a changed interval: loop must pick up the new ticker (100ms)
	configmanager.SetCollectorInterval(configmanager.CacheKeyDpuListRefresh, 100*time.Millisecond)
	configmanager.NotifyConfigReload()
	if !waitUntil(3*time.Second, func() bool {
		return atomic.LoadInt32(&dmgr.callCount) > firstCalls+1
	}) {
		t.Fatalf("dpu list was not refreshed with the new interval, calls=%d", atomic.LoadInt32(&dmgr.callCount))
	}

	cancel()
	waitWg(t, &wg, 2*time.Second)
}

// TestStartCollect covers first-time dpu list init, per-card and interface
// collection loops, collector filtering, and graceful shutdown.
func TestStartCollect(t *testing.T) {
	defer resetMockCollectorInterval()
	configmanager.SetCollectorInterval(mockCollectorCacheKey, 60*time.Millisecond)
	defer SetChains(nil, nil)

	supported := &mockCollector{supported: true}
	unsupported := &mockCollector{supported: false}
	SetChains(
		[]metricscollector.MetricsCollector{supported, unsupported},
		[]metricscollector.MetricsCollector{supported},
	)

	dmgr := &mockDeviceManager{list: []device.DPU{{CardName: "hinic0", CardType: "huawei"}}}
	n := NewDpuCollector(dmgr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	StartCollect(&wg, ctx, n)

	// dpu list must be initialized at first time (cache was empty)
	if !waitUntil(2*time.Second, func() bool { return len(n.GetDpuList()) == 1 }) {
		t.Fatalf("StartCollect did not initialize dpu list, list=%v", n.GetDpuList())
	}
	// supported collector runs in both the dpu-card loop and the iface loop
	if !waitUntil(3*time.Second, func() bool { return atomic.LoadInt32(&supported.collects) >= 1 }) {
		t.Fatalf("collect loop did not run, collects=%d", atomic.LoadInt32(&supported.collects))
	}

	cancel()
	waitWg(t, &wg, 3*time.Second)

	if got := atomic.LoadInt32(&unsupported.collects); got != 0 {
		t.Errorf("unsupported collector ran %d times, want 0", got)
	}
	if got := atomic.LoadInt32(&supported.preCnt); got < 2 {
		t.Errorf("PreCollect called %d times, want >= 2 (dpu loop + iface loop)", got)
	}
	if got := atomic.LoadInt32(&supported.postCnt); got < 2 {
		t.Errorf("PostCollect called %d times, want >= 2", got)
	}
}
