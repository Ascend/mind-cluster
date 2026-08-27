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
	"sync/atomic"
	"testing"
	"time"

	"huawei.com/dpu-exporter/pkg/collector/metricscollector"
	"huawei.com/dpu-exporter/pkg/configmanager"
)

// TestWaitForNextSignal covers all three wake-up reasons.
func TestWaitForNextSignal(t *testing.T) {
	// timer expiry
	if got := waitForNextSignal(context.Background(), 10*time.Millisecond, make(chan struct{})); got != wakeByTimer {
		t.Errorf("waitForNextSignal(timer) = %v, want %v", got, wakeByTimer)
	}

	// context cancel
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := waitForNextSignal(ctx, time.Hour, nil); got != wakeByContext {
		t.Errorf("waitForNextSignal(ctx) = %v, want %v", got, wakeByContext)
	}

	// config reload signal
	reloadCh := make(chan struct{}, 1)
	reloadCh <- struct{}{}
	if got := waitForNextSignal(context.Background(), time.Hour, reloadCh); got != wakeByConfigReload {
		t.Errorf("waitForNextSignal(reload) = %v, want %v", got, wakeByConfigReload)
	}
}

// TestBuildSchedule covers chain-to-schedule conversion including nil skipping.
func TestBuildSchedule(t *testing.T) {
	defer resetMockCollectorInterval()
	configmanager.SetCollectorInterval(mockCollectorCacheKey, 100*time.Millisecond)

	c := &mockCollector{supported: true}
	s := buildSchedule([]metricscollector.MetricsCollector{nil, c})
	if len(s.entries) != 1 {
		t.Fatalf("buildSchedule entries = %d, want 1 (nil collector skipped)", len(s.entries))
	}
	e := s.entries[0]
	if e.collector != c {
		t.Errorf("entry.collector = %v, want %v", e.collector, c)
	}
	if e.cacheKey != mockCollectorCacheKey {
		t.Errorf("entry.cacheKey = %q, want %q", e.cacheKey, mockCollectorCacheKey)
	}
	if e.interval != 100*time.Millisecond {
		t.Errorf("entry.interval = %v, want 100ms", e.interval)
	}
	if e.nextRun.IsZero() {
		t.Error("entry.nextRun not initialized")
	}
}

// TestMarkAllDueAndPopDue covers due-collector computation.
func TestMarkAllDueAndPopDue(t *testing.T) {
	defer resetMockCollectorInterval()
	configmanager.SetCollectorInterval(mockCollectorCacheKey, time.Hour)

	c1 := &mockCollector{supported: true}
	c2 := &mockCollector{supported: true}
	s := buildSchedule([]metricscollector.MetricsCollector{c1, c2})

	// nextRun == build time: both are due now
	if due := s.popDue(time.Now()); len(due) != 2 {
		t.Errorf("popDue() = %d collectors, want 2", len(due))
	}

	// push c1 into the future: only c2 due
	s.entries[0].nextRun = time.Now().Add(time.Hour)
	if due := s.popDue(time.Now()); len(due) != 1 || due[0] != c2 {
		t.Errorf("popDue() = %v, want [%v]", due, c2)
	}

	// markAllDue makes everything due again
	s.markAllDue()
	if due := s.popDue(time.Now()); len(due) != 2 {
		t.Errorf("popDue() after markAllDue = %d collectors, want 2", len(due))
	}

	// nil-collector entries are skipped
	sNil := &collectorSchedule{entries: []scheduleEntry{{collector: nil, nextRun: time.Now()}}}
	if due := sNil.popDue(time.Now()); len(due) != 0 {
		t.Errorf("popDue() with nil collector = %d, want 0", len(due))
	}
}

// TestUpdateNext covers next-run updates for ran/unran collectors and collect-once.
func TestUpdateNext(t *testing.T) {
	defer resetMockCollectorInterval()
	configmanager.SetCollectorInterval(mockCollectorCacheKey, 100*time.Millisecond)

	c1 := &mockCollector{supported: true}
	c2 := &mockCollector{supported: true}
	s := buildSchedule([]metricscollector.MetricsCollector{c1, c2})
	now := time.Now()

	// empty ran list: no-op
	prev := s.entries[0].nextRun
	s.updateNext(nil, now)
	if s.entries[0].nextRun != prev {
		t.Error("updateNext(nil) changed nextRun, want no-op")
	}

	// only c1 ran: c1 moves to now+interval, c2 untouched
	prev2 := s.entries[1].nextRun
	s.updateNext([]metricscollector.MetricsCollector{c1}, now)
	if got := s.entries[0].nextRun; !got.Equal(now.Add(100 * time.Millisecond)) {
		t.Errorf("c1 nextRun = %v, want %v", got, now.Add(100*time.Millisecond))
	}
	if s.entries[1].nextRun != prev2 {
		t.Error("unran collector's nextRun changed")
	}

	// collect-once: nextRun becomes far future
	configmanager.SetCollectorInterval(mockCollectorCacheKey, configmanager.CollectOnceInterval)
	s.updateNext([]metricscollector.MetricsCollector{c1}, now)
	if !s.entries[0].nextRun.Equal(farFutureTime) {
		t.Errorf("collect-once nextRun = %v, want %v", s.entries[0].nextRun, farFutureTime)
	}
}

// TestNextWaitDuration covers all wait duration calculation branches.
func TestNextWaitDuration(t *testing.T) {
	// empty schedule: 1 second
	s := &collectorSchedule{}
	if got := s.nextWaitDuration(); got != time.Second {
		t.Errorf("empty schedule nextWaitDuration = %v, want 1s", got)
	}

	// overdue entry: minimum wait
	c := &mockCollector{supported: true}
	s = &collectorSchedule{entries: []scheduleEntry{{collector: c, nextRun: time.Now().Add(-time.Minute)}}}
	if got := s.nextWaitDuration(); got != minWaitDuration {
		t.Errorf("overdue nextWaitDuration = %v, want %v", got, minWaitDuration)
	}

	// future entry: approximately the remaining time
	s = &collectorSchedule{entries: []scheduleEntry{{collector: c, nextRun: time.Now().Add(200 * time.Millisecond)}}}
	if got := s.nextWaitDuration(); got <= 0 || got > 200*time.Millisecond {
		t.Errorf("future nextWaitDuration = %v, want (0, 200ms]", got)
	}

	// far-future entry is skipped, nil collector is skipped: default group interval
	s = &collectorSchedule{entries: []scheduleEntry{
		{collector: c, nextRun: farFutureTime},
		{collector: nil, nextRun: time.Now()},
	}}
	if got := s.nextWaitDuration(); got != configmanager.DefaultGroupInterval {
		t.Errorf("all-far-future nextWaitDuration = %v, want %v", got, configmanager.DefaultGroupInterval)
	}
}

// TestRunCollectorLoop_ContextCancel covers the stop path: PreCollect/PostCollect
// are called once and no collection happens when the context is already cancelled.
func TestRunCollectorLoop_ContextCancel(t *testing.T) {
	defer resetMockCollectorInterval()
	resetMockCollectorInterval()

	c := &mockCollector{supported: true}
	stopped := int32(0)
	opts := collectorLoopOptions{
		loadChain: func() []metricscollector.MetricsCollector {
			return []metricscollector.MetricsCollector{c}
		},
		onStop: func() { atomic.AddInt32(&stopped, 1) },
		runDueCollectors: func(due []metricscollector.MetricsCollector) {
			for _, cc := range due {
				cc.CollectToCache(nil)
			}
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before start for a deterministic stop path

	n := NewDpuCollector(&mockDeviceManager{})
	done := make(chan struct{})
	go func() {
		runCollectorLoop(ctx, n, opts)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runCollectorLoop did not stop after context cancel")
	}

	if atomic.LoadInt32(&stopped) != 1 {
		t.Errorf("onStop called %d times, want 1", atomic.LoadInt32(&stopped))
	}
	if got := atomic.LoadInt32(&c.preCnt); got != 1 {
		t.Errorf("PreCollect called %d times, want 1", got)
	}
	if got := atomic.LoadInt32(&c.postCnt); got != 1 {
		t.Errorf("PostCollect called %d times, want 1", got)
	}
	if got := atomic.LoadInt32(&c.collects); got != 0 {
		t.Errorf("CollectToCache called %d times, want 0", got)
	}
}

// TestRunCollectorLoop_TimerCollect covers the timer path: due collectors run.
func TestRunCollectorLoop_TimerCollect(t *testing.T) {
	defer resetMockCollectorInterval()
	configmanager.SetCollectorInterval(mockCollectorCacheKey, 60*time.Millisecond)

	c := &mockCollector{supported: true}
	opts := collectorLoopOptions{
		loadChain: func() []metricscollector.MetricsCollector {
			return []metricscollector.MetricsCollector{c}
		},
		runDueCollectors: func(due []metricscollector.MetricsCollector) {
			for _, cc := range due {
				cc.CollectToCache(nil)
			}
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	n := NewDpuCollector(&mockDeviceManager{})
	done := make(chan struct{})
	go func() {
		runCollectorLoop(ctx, n, opts)
		close(done)
	}()

	if !waitUntil(3*time.Second, func() bool { return atomic.LoadInt32(&c.collects) >= 1 }) {
		t.Fatalf("due collector did not run, collects=%d", atomic.LoadInt32(&c.collects))
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runCollectorLoop did not stop after context cancel")
	}
	if got := atomic.LoadInt32(&c.postCnt); got != 1 {
		t.Errorf("PostCollect called %d times, want 1", got)
	}
}

// TestRunCollectorLoop_ConfigReload covers the reload path: the chain is
// reloaded and the new chain's collectors get PreCollect called.
func TestRunCollectorLoop_ConfigReload(t *testing.T) {
	defer resetMockCollectorInterval()
	resetMockCollectorInterval()

	c1 := &mockCollector{supported: true}
	c2 := &mockCollector{supported: true}
	loadCnt := int32(0)
	opts := collectorLoopOptions{
		loadChain: func() []metricscollector.MetricsCollector {
			if atomic.AddInt32(&loadCnt, 1) == 1 {
				return []metricscollector.MetricsCollector{c1}
			}
			return []metricscollector.MetricsCollector{c1, c2}
		},
		runDueCollectors: func(due []metricscollector.MetricsCollector) {
			for _, cc := range due {
				cc.CollectToCache(nil)
			}
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	n := NewDpuCollector(&mockDeviceManager{})
	done := make(chan struct{})
	go func() {
		runCollectorLoop(ctx, n, opts)
		close(done)
	}()

	// keep notifying until the loop reloads its chain (subscription may race with the first notify)
	stopNotify := make(chan struct{})
	go func() {
		for {
			select {
			case <-stopNotify:
				return
			default:
				configmanager.NotifyConfigReload()
				time.Sleep(20 * time.Millisecond)
			}
		}
	}()
	if !waitUntil(3*time.Second, func() bool { return atomic.LoadInt32(&loadCnt) >= 2 }) {
		close(stopNotify)
		cancel()
		<-done
		t.Fatalf("chain was not reloaded on config reload, loadChain calls=%d", atomic.LoadInt32(&loadCnt))
	}
	close(stopNotify)

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runCollectorLoop did not stop after context cancel")
	}

	if got := atomic.LoadInt32(&c1.preCnt); got < 2 {
		t.Errorf("c1 PreCollect called %d times, want >= 2 (initial + reload)", got)
	}
	if got := atomic.LoadInt32(&c2.preCnt); got < 1 {
		t.Errorf("c2 PreCollect called %d times, want >= 1 (after reload)", got)
	}
}

// TestGetCollectorNames covers collector name listing.
func TestGetCollectorNames(t *testing.T) {
	c1 := &mockCollector{supported: true}
	c2 := &mockCollector{supported: true}

	names := getCollectorNames([]metricscollector.MetricsCollector{c1, c2})
	if len(names) != 2 || names[0] != mockCollectorCacheKey || names[1] != mockCollectorCacheKey {
		t.Errorf("getCollectorNames() = %v, want [%s %s]", names, mockCollectorCacheKey, mockCollectorCacheKey)
	}
	if got := getCollectorNames(nil); len(got) != 0 {
		t.Errorf("getCollectorNames(nil) = %v, want empty", got)
	}
}
