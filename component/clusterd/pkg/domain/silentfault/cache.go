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

// Package silentfault silent fault detection domain caches
package silentfault

import (
	"sync"

	"clusterd/pkg/domain/conf"
)

// PendingEvent stage-1 event awaiting validation (stored per job, with the job node snapshot)
type PendingEvent struct {
	JobID     string   // job identifier (vcjob UID, consistent with the job cache key)
	FailNode  string   // first-error node (the faulty node of this pod-failed event)
	TaskNodes []string // all nodes occupied by the job when the event occurred (PreServerList snapshot, including FailNode)
	Timestamp int64    // reschedule occurrence time (seconds)
}

// FirstFaultEvent stage-2 node-level event (first error / non-first error)
type FirstFaultEvent struct {
	JobID     string
	IsFirst   bool  // true: first-error record; false: non-first-error record (used to break the consecutive run)
	Timestamp int64 // seconds
}

// PendingCache stage-1 cache: events awaiting validation + CM dedup set
type PendingCache struct {
	events    map[string][]*PendingEvent    // jobID -> list of events awaiting validation
	processed map[string]map[int64]struct{} // dedup: jobID -> set of processed RescheduleTimeStamp
	mutex     sync.RWMutex
}

// FirstFaultMgr stage-2 cache: node-level event storage
type FirstFaultMgr struct {
	events map[string][]*FirstFaultEvent // nodeName -> event list (appended in ascending time order)
	mutex  sync.RWMutex
}

// NewPendingCache create a pending cache
func NewPendingCache() *PendingCache {
	return &PendingCache{
		events:    make(map[string][]*PendingEvent),
		processed: make(map[string]map[int64]struct{}),
		mutex:     sync.RWMutex{},
	}
}

// NewFirstFaultMgr create a first fault manager
func NewFirstFaultMgr() *FirstFaultMgr {
	return &FirstFaultMgr{
		events: make(map[string][]*FirstFaultEvent),
		mutex:  sync.RWMutex{},
	}
}

// Add appends an event awaiting validation to the job's event list
func (c *PendingCache) Add(ev *PendingEvent) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.events[ev.JobID] = append(c.events[ev.JobID], ev)
}

// TakeDue removes and deletes all events exceeding O seconds (deleted once closed, whether later dispatched or dropped)
func (c *PendingCache) TakeDue(now int64) []*PendingEvent {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	var due []*PendingEvent
	for jobID, evs := range c.events {
		kept := evs[:0]
		for _, ev := range evs {
			if now >= ev.Timestamp+conf.GetHwWindowSeconds() {
				due = append(due, ev)
			} else {
				kept = append(kept, ev)
			}
		}
		if len(kept) == 0 {
			delete(c.events, jobID)
		} else {
			c.events[jobID] = kept
		}
	}
	return due
}

// HasProcessed reports whether the job's ts has been processed (guards against duplicate CM push)
func (c *PendingCache) HasProcessed(jobID string, ts int64) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	_, ok := c.processed[jobID][ts]
	return ok
}

// MarkProcessed marks the job's ts as processed
func (c *PendingCache) MarkProcessed(jobID string, ts int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.processed[jobID] == nil {
		c.processed[jobID] = make(map[int64]struct{})
	}
	c.processed[jobID][ts] = struct{}{}
}

// PruneProcessed deletes ts entries that have disappeared from the current CM (sliding window)
func (c *PendingCache) PruneProcessed(jobID string, currentTs map[int64]struct{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	set := c.processed[jobID]
	if len(set) == 0 {
		delete(c.processed, jobID)
		return
	}
	for ts := range set {
		if _, ok := currentTs[ts]; !ok {
			delete(set, ts)
		}
	}
	if len(set) == 0 {
		delete(c.processed, jobID)
	}
}

// PruneOrphanProcessed deletes jobIDs that have disappeared from the CM (job deleted)
func (c *PendingCache) PruneOrphanProcessed(currentJobs map[string]struct{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for jobID := range c.processed {
		if _, ok := currentJobs[jobID]; !ok {
			delete(c.processed, jobID)
		}
	}
}

// ResetEvents clears the events awaiting validation (stage-1 pending events)
func (c *PendingCache) ResetEvents() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.events = make(map[string][]*PendingEvent)
}

// ResetProcessed clears the dedup set (processed RescheduleTimeStamp)
func (c *PendingCache) ResetProcessed() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.processed = make(map[string]map[int64]struct{})
}

// AddEvent appends a first-error/non-first-error event to the node's event list
func (m *FirstFaultMgr) AddEvent(node string, ev *FirstFaultEvent) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.events[node] = append(m.events[node], ev)
}

// Nodes returns a snapshot of all node names for the detector to iterate
func (m *FirstFaultMgr) Nodes() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	nodes := make([]string, 0, len(m.events))
	for node := range m.events {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetEvents returns a deep-copy snapshot of the node's event list (avoids concurrent write conflicts with AddEvent)
func (m *FirstFaultMgr) GetEvents(node string) []*FirstFaultEvent {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	src := m.events[node]
	cp := make([]*FirstFaultEvent, len(src))
	for i, e := range src {
		ee := *e
		cp[i] = &ee
	}
	return cp
}

// ClearNode deletes all events of a node (cleanup on hit/isolated state)
func (m *FirstFaultMgr) ClearNode(node string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.events, node)
}

// PruneExpiredAll deletes events past the cutoff for all nodes, removing empty nodes
func (m *FirstFaultMgr) PruneExpiredAll(now, window int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	cutoff := now - window
	for node, events := range m.events {
		kept := events[:0]
		for _, e := range events {
			if e.Timestamp >= cutoff {
				kept = append(kept, e)
			}
		}
		if len(kept) == 0 {
			delete(m.events, node)
		} else {
			m.events[node] = kept
		}
	}
}

// Reset clears all node events
func (m *FirstFaultMgr) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.events = make(map[string][]*FirstFaultEvent)
}
