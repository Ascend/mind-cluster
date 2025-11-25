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

// Package app fault queue function
package app

import (
	"sync"

	"ascend-common/common-utils/hwlog"
	"container-manager/pkg/common"
)

// QueueCache fault queue cache
var QueueCache *FaultQueue

func init() {
	QueueCache = NewFaultQueue()
}

// FaultQueue fault queue
type FaultQueue struct {
	faults []common.DevFaultInfo
	mutex  sync.Mutex
}

// NewFaultQueue new fault queue
func NewFaultQueue() *FaultQueue {
	return &FaultQueue{
		faults: make([]common.DevFaultInfo, 0),
		mutex:  sync.Mutex{},
	}
}

// Push to push new item to faults
func (q *FaultQueue) Push(newItem common.DevFaultInfo) {
	const maxFaultCacheNum = 10000
	if q.Len() >= maxFaultCacheNum {
		hwlog.RunLog.Errorf("add fault to queue failed, "+
			"fault number in queue exceeds the upper limit %d", maxFaultCacheNum)
		return
	}
	if newItem.FaultLevel == common.NormalNPU || newItem.FaultLevel == common.UnknownLevel {
		hwlog.RunLog.Warn("normal or unknown fault, do not deal")
		return
	}
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.faults = append(q.faults, newItem)
}

// Pop to pop item from faults
func (q *FaultQueue) Pop() common.DevFaultInfo {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.faults) == 0 {
		return common.DevFaultInfo{}
	}
	removed := q.faults[0]
	// remove the head item
	q.faults = q.faults[1:]
	return removed
}

// Len length of faults
func (q *FaultQueue) Len() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.faults)
}
