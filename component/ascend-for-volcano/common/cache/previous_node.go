/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package affinity provides pod-to-node affinity cache for preferring
// previously-used nodes when rescheduling evicted pods.
package cache

import (
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

const (
	// DefaultTTL is the default time-to-live for cache entries (72 hours).
	// Owners not updated within this duration are evicted to prevent memory leaks.
	DefaultTTL = 72 * time.Hour
)

// RankNodeEntry stores the current and previous node assignment for a single rank.
// Previous acts as a rollback anchor: when an allocation is rolled back
// (DeallocateFunc with Pending status), Node is restored from Previous.
type RankNodeEntry struct {
	Node     string // current node assignment
	Previous string // previous assignment (rollback anchor), empty if this is a first-time assignment
}

// PodNodeAffinityCache stores pod-to-node mapping for "prefer previous node" scheduling.
// The cache is held by ScheduleHandler and survives across scheduler sessions.
//
// Two-layer structure:
//
//	layer 1: ownerUID (PodGroup's owner, e.g. Deployment UID)
//	layer 2: rankIndex → RankNodeEntry{Node, Previous}
//
// Using owner UID as the first key ensures the mapping persists even when
// the PodGroup/Job is recreated with a different UID after eviction.
//
// Concurrency: no lock is needed because Volcano's scheduling framework processes
// tasks sequentially in a single goroutine.
type PodNodeAffinityCache struct {
	// key: ownerUID, value: rankIndex → RankNodeEntry
	OwnerToRankNodes map[string]map[string]*RankNodeEntry
	// key: ownerUID, value: last update timestamp (Unix seconds)
	UpdateTime map[string]int64
}

// NewPodNodeAffinityCache creates a new empty cache.
func NewPodNodeAffinityCache() *PodNodeAffinityCache {
	return &PodNodeAffinityCache{
		OwnerToRankNodes: make(map[string]map[string]*RankNodeEntry),
		UpdateTime:       make(map[string]int64),
	}
}

// RecordAssignment records a pod's node assignment.
// The old Node value is preserved in Previous as a rollback anchor.
func (c *PodNodeAffinityCache) RecordAssignment(ownerUID types.UID, rankIndex, nodeName string) {
	ownerKey := string(ownerUID)
	if c.OwnerToRankNodes[ownerKey] == nil {
		c.OwnerToRankNodes[ownerKey] = make(map[string]*RankNodeEntry)
	}
	entry, ok := c.OwnerToRankNodes[ownerKey][rankIndex]
	if !ok {
		entry = &RankNodeEntry{}
		c.OwnerToRankNodes[ownerKey][rankIndex] = entry
	}
	entry.Previous = entry.Node
	entry.Node = nodeName
	c.UpdateTime[ownerKey] = time.Now().Unix()
	klog.V(util.LogDebugLev).Infof("affinity cache: recorded owner=%s rank=%s -> %s (prev=%s)",
		ownerKey, rankIndex, nodeName, entry.Previous)
}

// RollbackAssignment restores the previous node assignment after an allocation
// rollback. If there is a Previous value, Node is restored from it and Previous
// is cleared. If there is no Previous value (first-time assignment), the entry
// is removed entirely — it never represented a real pod placement.
func (c *PodNodeAffinityCache) RollbackAssignment(ownerUID types.UID, rankIndex string) {
	ownerKey := string(ownerUID)
	rankNodes, ok := c.OwnerToRankNodes[ownerKey]
	if !ok {
		return
	}
	entry, ok := rankNodes[rankIndex]
	if !ok {
		return
	}
	if entry.Previous != "" {
		entry.Node = entry.Previous
		entry.Previous = ""
		klog.V(util.LogDebugLev).Infof("affinity cache: rolled back owner=%s rank=%s -> %s",
			ownerKey, rankIndex, entry.Node)
	} else {
		delete(rankNodes, rankIndex)
		if len(rankNodes) == 0 {
			delete(c.OwnerToRankNodes, ownerKey)
			delete(c.UpdateTime, ownerKey)
		}
		klog.V(util.LogDebugLev).Infof("affinity cache: removed stale entry owner=%s rank=%s",
			ownerKey, rankIndex)
	}
}

// GetPreferredNode returns the preferred node for a pod, or empty string if not found.
func (c *PodNodeAffinityCache) GetPreferredNode(ownerUID types.UID, rankIndex string) string {
	ownerKey := string(ownerUID)
	if rankNodes, ok := c.OwnerToRankNodes[ownerKey]; ok {
		if entry, ok := rankNodes[rankIndex]; ok {
			return entry.Node
		}
	}
	return ""
}

// GetPreferredNodeMap returns rank→nodeName for a rank range [startRank, endRank).
// The returned map uses integer rank as key for direct position-based reordering.
// Returns nil if the owner has no cached entries in this range.
func (c *PodNodeAffinityCache) GetPreferredNodeMap(ownerUID types.UID, startRank, endRank int) map[int]string {
	ownerKey := string(ownerUID)
	rankNodes, ok := c.OwnerToRankNodes[ownerKey]
	if !ok {
		return nil
	}
	result := make(map[int]string, endRank-startRank)
	for rank := startRank; rank < endRank; rank++ {
		if entry, ok := rankNodes[strconv.Itoa(rank)]; ok && entry.Node != "" {
			result[rank] = entry.Node
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// RefreshOwner updates the timestamp for an owner, marking it as still active.
// Called during InitNPUSession for each owner whose PodGroup still exists,
// so TTL counts from the last time the PG was seen rather than the last assignment.
func (c *PodNodeAffinityCache) RefreshOwner(ownerUID types.UID) {
	ownerKey := string(ownerUID)
	if _, ok := c.OwnerToRankNodes[ownerKey]; ok {
		c.UpdateTime[ownerKey] = time.Now().Unix()
	}
}

// EvictExpired removes owners whose last update exceeds the given TTL.
// Called periodically to prevent unbounded memory growth from deleted owners.
func (c *PodNodeAffinityCache) EvictExpired(ttl time.Duration) {
	now := time.Now().Unix()
	cutoff := now - int64(ttl.Seconds())
	for ownerKey, ts := range c.UpdateTime {
		if ts < cutoff {
			delete(c.OwnerToRankNodes, ownerKey)
			delete(c.UpdateTime, ownerKey)
			klog.V(util.LogDebugLev).Infof("affinity cache: evicted expired owner %s (age=%ds)",
				ownerKey, now-ts)
		}
	}
}

// Size returns the total number of cached rank→node entries.
func (c *PodNodeAffinityCache) Size() int {
	total := 0
	for _, rankNodes := range c.OwnerToRankNodes {
		total += len(rankNodes)
	}
	return total
}
