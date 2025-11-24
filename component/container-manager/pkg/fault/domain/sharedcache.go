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

// Package domain shared cache function for container controller
package domain

import (
	"sync"

	"ascend-common/common-utils/hwlog"
	"container-manager/pkg/common"
)

const maxFaultNumPerNPU = 100

// SharedFaultCache shared fault cache between fault manager and ctr controller
var SharedFaultCache *sharedCache

type sharedCache struct {
	faults map[int32][]*common.DevFaultInfo // key: dev phy id, value: []fault info
	// UpdateChan for unprocessed fault events
	UpdateChan chan struct{}
	mutex      sync.Mutex
}

func init() {
	SharedFaultCache = &sharedCache{
		faults:     make(map[int32][]*common.DevFaultInfo),
		UpdateChan: make(chan struct{}, 1),
		mutex:      sync.Mutex{},
	}
}

// Notify unprocessed fault events updated, notify ctr controller to process
func (sc *sharedCache) Notify() {
	if len(sc.UpdateChan) == 0 {
		sc.UpdateChan <- struct{}{}
	}
}

// GetAndClean get and clean fault cache
func (sc *sharedCache) GetAndClean() map[int32][]*common.DevFaultInfo {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	faults := sc.faults
	sc.faults = make(map[int32][]*common.DevFaultInfo)
	return faults
}

// AddFault add fault to faults in cache
func (sc *sharedCache) AddFault(newFault *common.DevFaultInfo) {
	if newFault == nil {
		hwlog.RunLog.Error("new fault is nil")
		return
	}
	if newFault.Assertion == common.FaultRecover || newFault.Assertion == common.FaultOnce {
		return
	}
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	if sc.getNPUFaultNum(newFault.PhyID) >= maxFaultNumPerNPU {
		hwlog.RunLog.Errorf("add fault to sharedCache failed, "+
			"fault number per npu in cache exceeds tupper limit %d", maxFaultNumPerNPU)
		return
	}
	defer sc.Notify()
	npuFault, npuExist := sc.faults[newFault.PhyID]
	if !npuExist {
		sc.faults[newFault.PhyID] = []*common.DevFaultInfo{newFault}
		return
	}
	sc.faults[newFault.PhyID] = append(npuFault, newFault)
}

func (sc *sharedCache) getNPUFaultNum(pyhID int32) int {
	npuFault, ok := sc.faults[pyhID]
	if !ok {
		return 0
	}
	return len(npuFault)
}

// DeepCopy deep copy faults
func (sc *sharedCache) DeepCopy() (map[int32][]*common.DevFaultInfo, error) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	result := new(map[int32][]*common.DevFaultInfo)
	if err := common.DeepCopy(result, sc.faults); err != nil {
		return nil, err
	}
	return *result, nil
}
