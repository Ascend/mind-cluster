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
	"time"

	"huawei.com/dpu-exporter/pkg/collector/metricscollector"
	"huawei.com/dpu-exporter/pkg/configmanager"
)

// waitResult represents the reason the collector loop woke up.
type waitResult int

const (
	wakeByContext waitResult = iota
	wakeByTimer
	wakeByConfigReload
)

const (
	minWaitDuration = 50 * time.Millisecond
)

var farFutureTime = time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)

// waitForNextSignal blocks until context cancel, config reload, or timer expiry
func waitForNextSignal(ctx context.Context, wait time.Duration, reloadCh <-chan struct{}) waitResult {
	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return wakeByContext
	case <-reloadCh:
		return wakeByConfigReload
	case <-timer.C:
		return wakeByTimer
	}
}

// collectorSchedule manages per-collector scheduling within a collection loop
type collectorSchedule struct {
	entries []scheduleEntry
}

type scheduleEntry struct {
	collector metricscollector.MetricsCollector
	cacheKey  string
	interval  time.Duration
	nextRun   time.Time
}

func buildSchedule(chain []metricscollector.MetricsCollector) collectorSchedule {
	s := collectorSchedule{entries: make([]scheduleEntry, 0, len(chain))}
	now := time.Now()
	for _, c := range chain {
		if c == nil {
			continue
		}
		key := metricscollector.GetCacheKey(c)
		interval := configmanager.GetCollectorInterval(key, configmanager.DefaultGroupInterval)
		s.entries = append(s.entries, scheduleEntry{
			collector: c,
			cacheKey:  key,
			interval:  interval,
			nextRun:   now,
		})
	}
	return s
}

func (s *collectorSchedule) markAllDue() {
	now := time.Now()
	for i := range s.entries {
		s.entries[i].nextRun = now
	}
}

func (s *collectorSchedule) popDue(now time.Time) []metricscollector.MetricsCollector {
	due := make([]metricscollector.MetricsCollector, 0)
	for _, e := range s.entries {
		if e.collector == nil {
			continue
		}
		if !e.nextRun.After(now) {
			due = append(due, e.collector)
		}
	}
	return due
}

func (s *collectorSchedule) updateNext(ran []metricscollector.MetricsCollector, now time.Time) {
	if len(ran) == 0 {
		return
	}
	ranSet := make(map[metricscollector.MetricsCollector]struct{}, len(ran))
	for _, c := range ran {
		ranSet[c] = struct{}{}
	}
	for i := range s.entries {
		if _, ok := ranSet[s.entries[i].collector]; !ok {
			continue
		}
		s.entries[i].interval = configmanager.GetCollectorInterval(s.entries[i].cacheKey, s.entries[i].interval)
		if s.entries[i].interval == configmanager.CollectOnceInterval {
			s.entries[i].nextRun = farFutureTime
			continue
		}
		s.entries[i].nextRun = now.Add(s.entries[i].interval)
	}
}

func (s *collectorSchedule) nextWaitDuration() time.Duration {
	if len(s.entries) == 0 {
		return time.Second
	}
	var min time.Duration = -1
	for _, e := range s.entries {
		if e.collector == nil {
			continue
		}
		if !e.nextRun.Before(farFutureTime) {
			continue
		}
		d := time.Until(e.nextRun)
		if d <= 0 {
			return minWaitDuration
		}
		if min < 0 || d < min {
			min = d
		}
	}
	if min < 0 {
		return configmanager.DefaultGroupInterval
	}
	return min
}

// collectorLoopOptions configures the behavior of a collector loop
type collectorLoopOptions struct {
	loadChain        func() []metricscollector.MetricsCollector
	onStop           func()
	runDueCollectors func([]metricscollector.MetricsCollector)
}

// runCollectorLoop is the core event loop for all collector goroutines.
// It handles three wake signals: context cancel, config reload, and timer expiry.
func runCollectorLoop(ctx context.Context, n *DpuCollector, opts collectorLoopOptions) {
	currentChain := opts.loadChain()
	for _, c := range currentChain {
		if c != nil {
			c.PreCollect(n)
		}
	}
	defer func() {
		for _, c := range currentChain {
			if c != nil {
				c.PostCollect(n)
			}
		}
	}()

	schedule := buildSchedule(currentChain)
	schedule.markAllDue()

	reloadCh := configmanager.SubscribeReload()
	defer configmanager.UnsubscribeReload(reloadCh)

	for {
		result := waitForNextSignal(ctx, schedule.nextWaitDuration(), reloadCh)
		if result == wakeByContext {
			if opts.onStop != nil {
				opts.onStop()
			}
			return
		}
		if result == wakeByConfigReload {
			for _, c := range currentChain {
				if c != nil {
					c.PostCollect(n)
				}
			}
			currentChain = opts.loadChain()
			for _, c := range currentChain {
				if c != nil {
					c.PreCollect(n)
				}
			}
			schedule = buildSchedule(currentChain)
			schedule.markAllDue()
			configmanager.DrainReloadSignal(reloadCh)
			continue
		}
		// wakeByTimer: execute due collectors
		now := time.Now()
		dueCollectors := schedule.popDue(now)
		if len(dueCollectors) == 0 {
			continue
		}
		opts.runDueCollectors(dueCollectors)
		schedule.updateNext(dueCollectors, now)
	}
}

func getCollectorNames(collectors []metricscollector.MetricsCollector) []string {
	names := make([]string, 0, len(collectors))
	for _, c := range collectors {
		names = append(names, metricscollector.GetCacheKey(c))
	}
	return names
}
