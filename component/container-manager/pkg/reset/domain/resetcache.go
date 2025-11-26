/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package domain reset cache
package domain

import (
	"sync"

	"ascend-common/common-utils/hwlog"
)

var (
	inResetCache *NpuInResetCache
	initOnce     sync.Once
)

// NpuInResetCache cache if npu is in resetting
type NpuInResetCache struct {
	npuInResetLock  sync.Mutex
	npuInResetCache map[int32]struct{} // InResetPhyId
}

// GetNpuInResetCache create a new instance of NpuInResetCache
func GetNpuInResetCache() *NpuInResetCache {
	initOnce.Do(func() {
		inResetCache = &NpuInResetCache{npuInResetCache: make(map[int32]struct{})}
	})
	return inResetCache
}

// DeepCopy get a copy of NpuInResetCache
func (c *NpuInResetCache) DeepCopy() map[int32]struct{} {
	c.npuInResetLock.Lock()
	defer c.npuInResetLock.Unlock()
	copyCache := make(map[int32]struct{})
	for phyId := range c.npuInResetCache {
		copyCache[phyId] = struct{}{}
	}
	return copyCache
}

// SetNpuInReset add npus into NpuInResetCache
func (c *NpuInResetCache) SetNpuInReset(phyIds ...int32) {
	c.npuInResetLock.Lock()
	defer c.npuInResetLock.Unlock()
	hwlog.RunLog.Debugf("set physic Ids <%v> in reset status", phyIds)
	for _, phyId := range phyIds {
		c.npuInResetCache[phyId] = struct{}{}
	}
}

// IsNpuInReset get npu reset status
func (c *NpuInResetCache) IsNpuInReset(phyId int32) bool {
	c.npuInResetLock.Lock()
	defer c.npuInResetLock.Unlock()
	_, ok := c.npuInResetCache[phyId]
	return ok
}

// ClearNpuInReset remove npus from NpuInResetCache
func (c *NpuInResetCache) ClearNpuInReset(phyIds ...int32) {
	c.npuInResetLock.Lock()
	defer c.npuInResetLock.Unlock()
	hwlog.RunLog.Debugf("clear physic Ids <%v> reset status", phyIds)
	for _, phyId := range phyIds {
		delete(c.npuInResetCache, phyId)
	}
}

// FailedResetCountCache cache npu failed reset count
type FailedResetCountCache struct {
	failedResetCountLock  sync.Mutex
	failedResetCountCache map[int32]int // phyId: resetCount
}

// NewFailedResetCountCache create a new instance of FailedResetCountCache
func NewFailedResetCountCache() *FailedResetCountCache {
	return &FailedResetCountCache{failedResetCountCache: make(map[int32]int)}

}

// SetFailedResetCount set npus:count to FailedResetCountCache
func (c *FailedResetCountCache) SetFailedResetCount(phyId int32, count int) {
	c.failedResetCountLock.Lock()
	defer c.failedResetCountLock.Unlock()
	hwlog.RunLog.Debugf("set physic ID [%v] failed reset count %v", phyId, count)
	c.failedResetCountCache[phyId] = count
}

// GetFailedResetCount get failed reset count by physic ID
func (c *FailedResetCountCache) GetFailedResetCount(phyId int32) int {
	c.failedResetCountLock.Lock()
	defer c.failedResetCountLock.Unlock()
	var count int
	count, ok := c.failedResetCountCache[phyId]
	if !ok {
		count = 0
	}
	hwlog.RunLog.Debugf("physic Id [%v], current failed reset count: %v", phyId, count)
	return count
}

// GetAllFailedResetCountNpuId get all failed reset counts by physic IDs
func (c *FailedResetCountCache) GetAllFailedResetCountNpuId() []int32 {
	c.failedResetCountLock.Lock()
	defer c.failedResetCountLock.Unlock()
	var ids []int32
	for phyId, _ := range c.failedResetCountCache {
		ids = append(ids, phyId)
	}
	hwlog.RunLog.Debugf("current fault reset count npus: %v", ids)
	return ids
}

// ClearFailedResetCount clear failed reset count by physic ID
func (c *FailedResetCountCache) ClearFailedResetCount(phyId int32) {
	c.failedResetCountLock.Lock()
	defer c.failedResetCountLock.Unlock()
	hwlog.RunLog.Debugf("clear fault npu: %v reset count", phyId)
	delete(c.failedResetCountCache, phyId)
}
