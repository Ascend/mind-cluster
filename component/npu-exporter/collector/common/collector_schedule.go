/*
 * Copyright (c) 2026. Huawei Technologies Co., Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package common for general collector
package common

import (
	"context"
	"sync"
	"time"
)

// waitResult represents the reason why the collector loop wakes up from waiting.
type waitResult int

const (
	// wakeByContext context cancelled, need to exit loop
	wakeByContext waitResult = iota
	// wakeByTimer timer expired, time to collect
	wakeByTimer
	// wakeByConfigReload received config hot-reload signal, need to rebuild schedule
	wakeByConfigReload
)

const (
	collectOnceInterval  = -1 * time.Second
	defaultGroupInterval = 60 * time.Second
	disabledInterval     = -2 * time.Second // indicates this group collection is disabled
)

var (
	farFutureTime = time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)

	collectorIntervalMap sync.Map // map[string]time.Duration, key is GetCacheKey(collector)

	// subscribers Store all hot-updated subscriber channels configured
	subscribers   []chan struct{}
	subscribersMu sync.Mutex
)

// CollectOnceInterval indicates this collector should run only once.
func CollectOnceInterval() time.Duration {
	return collectOnceInterval
}

// DisabledInterval indicates this group collection is disabled.
func DisabledInterval() time.Duration {
	return disabledInterval
}

// subscribeConfigReload Sign up for a configuration hot update subscriber。
func subscribeConfigReload() <-chan struct{} {
	ch := make(chan struct{}, 1)
	subscribersMu.Lock()
	subscribers = append(subscribers, ch)
	subscribersMu.Unlock()
	return ch
}

// unsubscribeConfigReload Sign out of a configuration hot update subscriber。
func unsubscribeConfigReload(ch <-chan struct{}) {
	subscribersMu.Lock()
	defer subscribersMu.Unlock()
	for i, s := range subscribers {
		if s == ch {
			subscribers = append(subscribers[:i], subscribers[i+1:]...)
			return
		}
	}
}

// NotifyConfigReload Send signals through all subscribers one by one。
func NotifyConfigReload() {
	subscribersMu.Lock()
	subs := make([]chan struct{}, len(subscribers))
	copy(subs, subscribers)
	subscribersMu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// drainReloadSignal Drains the reload signal channel.
func drainReloadSignal(ch <-chan struct{}) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

// waitForNextSignal blocks until one of the following occurs:
// context cancellation, config hot-reload signal, or timer expiration.
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

// collectorSchedule manages the scheduling state for a collection cycle.
// Each scheduleEntry tracks its own collector and interval independently,
// allowing collectors to run at different frequencies within the same loop.
type collectorSchedule struct {
	entries []scheduleEntry
}

type scheduleEntry struct {
	collector MetricsCollector
	cacheKey  string
	interval  time.Duration
	nextRun   time.Time
}

// buildSchedule creates a new schedule from the given collector chain.
// Each collector gets its own entry with interval and nextRun initialized.
func buildSchedule(chain []MetricsCollector) collectorSchedule {
	s := collectorSchedule{entries: make([]scheduleEntry, 0, len(chain))}
	now := time.Now()
	for _, c := range chain {
		if c == nil {
			continue
		}
		key := GetCacheKey(c)
		interval := GetCollectorInterval(key, defaultGroupInterval)
		s.entries = append(s.entries, scheduleEntry{
			collector: c,
			cacheKey:  key,
			interval:  interval,
			nextRun:   now,
		})
	}
	return s
}

// markAllDue marks all entries as due to run immediately.
func (s *collectorSchedule) markAllDue() {
	now := time.Now()
	for i := range s.entries {
		s.entries[i].nextRun = now
	}
}

// popDue returns all collectors with nextRun timestamp <= now.
func (s *collectorSchedule) popDue(now time.Time) []MetricsCollector {
	due := make([]MetricsCollector, 0)
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

// updateNext advances the nextRun time for the given collectors.
// Re-reads interval from cache to support dynamic interval changes after hot-reload.
// Collectors with interval == collectOnceInterval are marked as far future.
func (s *collectorSchedule) updateNext(ran []MetricsCollector, now time.Time) {
	if len(ran) == 0 {
		return
	}
	ranSet := make(map[MetricsCollector]struct{}, len(ran))
	for _, c := range ran {
		ranSet[c] = struct{}{}
	}
	for i := range s.entries {
		if _, ok := ranSet[s.entries[i].collector]; !ok {
			continue
		}
		// Re-read interval to support dynamic changes after hot-reload
		s.entries[i].interval = GetCollectorInterval(s.entries[i].cacheKey, s.entries[i].interval)
		if s.entries[i].interval == collectOnceInterval {
			s.entries[i].nextRun = farFutureTime
			continue
		}
		s.entries[i].nextRun = now.Add(s.entries[i].interval)
	}
}

// nextWaitDuration returns the minimum duration until the next collector is due.
// Returns 50ms minimum for due collectors, defaultGroupInterval if all in far future.
func (s *collectorSchedule) nextWaitDuration() time.Duration {
	if len(s.entries) == 0 {
		return time.Second
	}
	var min time.Duration = -1
	for _, e := range s.entries {
		if e.collector == nil {
			continue
		}
		// Skip entries that have completed their one-time collection
		if !e.nextRun.Before(farFutureTime) {
			continue
		}
		d := time.Until(e.nextRun)
		if d <= 0 {
			// Already due, but enforce minimum 50 millisecond wait
			return time.Millisecond * 50
		}
		if min < 0 || d < min {
			min = d
		}
	}
	if min < 0 {
		// All entries are in far future, use default interval
		return defaultGroupInterval
	}
	return min
}

// SetCollectorInterval stores a collector's interval in the global collectorIntervalMap (sync.Map).
func SetCollectorInterval(cacheKey string, interval time.Duration) {
	if cacheKey == "" {
		return
	}
	if interval == collectOnceInterval {
		collectorIntervalMap.Store(cacheKey, interval)
		return
	}
	if interval <= 0 {
		return
	}
	collectorIntervalMap.Store(cacheKey, interval)
}

// GetCollectorInterval loads a collector's interval from the global collectorIntervalMap.
func GetCollectorInterval(cacheKey string, fallback time.Duration) time.Duration {
	if cacheKey == "" {
		return fallback
	}
	v, ok := collectorIntervalMap.Load(cacheKey)
	if !ok {
		return fallback
	}
	d, ok := v.(time.Duration)
	if !ok {
		return fallback
	}
	if d == collectOnceInterval {
		return d
	}
	if d <= 0 {
		return fallback
	}
	return d
}
