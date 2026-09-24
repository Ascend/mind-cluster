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

// Package hardwarefault provides a public, node-keyed hardware fault timeline. It is intended to
// be reused by any feature that needs to query whether a hardware fault occurred around a point
// in time (e.g. silent fault detection).
package hardwarefault

import (
	"context"
	"sort"
	"sync"
	"time"
)

// maxFaultRecords bounds the total retained records to prevent unbounded memory growth when
// hardware faults keep being reported continuously.
const maxFaultRecords = 4096

// dailyRetention is the daily cleanup retention window (1 day) in milliseconds, matching the unit
// the fault log stores in this system.
const dailyRetention = int64(24 * time.Hour / time.Millisecond)

// FaultHardwareLog is an ascending, deduped timeline of hardware fault times keyed by node name. All
// timestamp units are caller-defined; callers must keep Append and HasFaultIn/PruneExpired
// consistent (the silent fault detection usage stores milliseconds).
type FaultHardwareLog struct {
	faults map[string][]int64 // nodeName -> fault times (ascending and deduped)
	mutex  sync.RWMutex
}

// NewFaultHardwareLog creates an empty fault log.
func NewFaultHardwareLog() *FaultHardwareLog {
	return &FaultHardwareLog{
		faults: make(map[string][]int64),
		mutex:  sync.RWMutex{},
	}
}

// Append appends a fault time for the node, keeping ascending order, deduplicating identical
// times, and dropping the oldest records once the total exceeds maxFaultRecords.
func (l *FaultHardwareLog) Append(node string, t int64) {
	if t <= 0 {
		return
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()
	times := l.faults[node]
	pos := sort.Search(len(times), func(i int) bool { return times[i] >= t })
	if pos < len(times) && times[pos] == t {
		return
	}
	times = append(times, 0)
	copy(times[pos+1:], times[pos:])
	times[pos] = t
	l.faults[node] = times
	if l.total() > maxFaultRecords {
		l.dropOldest()
	}
}

// HasFaultIn reports whether the ascending timeline of the node has a fault time within the
// closed interval [start, end].
func (l *FaultHardwareLog) HasFaultIn(node string, start, end int64) bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	for _, t := range l.faults[node] {
		if t > end {
			break
		}
		if t >= start {
			return true
		}
	}
	return false
}

// PruneExpired deletes fault times older than now-window. now and window must use the same unit
// as Append.
func (l *FaultHardwareLog) PruneExpired(now, window int64) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	cutoff := now - window
	for node, times := range l.faults {
		kept := times[:0]
		for _, t := range times {
			if t >= cutoff {
				kept = append(kept, t)
			}
		}
		if len(kept) == 0 {
			delete(l.faults, node)
		} else {
			l.faults[node] = kept
		}
	}
}

// RunDailyCleanup starts a background goroutine that, once per hour, prunes records older than
// 1 day, until ctx is canceled. It is decoupled from the fault detection loop.
func (l *FaultHardwareLog) RunDailyCleanup(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			l.PruneExpired(time.Now().UnixMilli(), dailyRetention)
		}
	}
}

// Reset clears the fault timeline.
func (l *FaultHardwareLog) Reset() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.faults = make(map[string][]int64)
}

// Len returns the total number of retained records.
func (l *FaultHardwareLog) Len() int {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.total()
}

// total returns the total number of retained records. The caller must hold the lock.
func (l *FaultHardwareLog) total() int {
	n := 0
	for _, times := range l.faults {
		n += len(times)
	}
	return n
}

// dropOldest removes the globally oldest record. The caller must hold the lock and guarantee at
// least one record exists.
func (l *FaultHardwareLog) dropOldest() {
	var oldestNode string
	first := true
	for node, times := range l.faults {
		if len(times) == 0 {
			continue
		}
		if first || times[0] < l.faults[oldestNode][0] {
			oldestNode = node
			first = false
		}
	}
	if first {
		return
	}
	times := l.faults[oldestNode]
	times = times[1:]
	if len(times) == 0 {
		delete(l.faults, oldestNode)
	} else {
		l.faults[oldestNode] = times
	}
}
