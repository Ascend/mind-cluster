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

// Package domain device info struct
package domain

import (
	"sync"

	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
)

// DevCache dev cache
type DevCache struct {
	devInfoMap map[int32]*devInfo
	mutex      sync.Mutex
}

type devInfo struct {
	CtrIds     []string
	Status     string
	DevsOnRing []int32
}

// NewDevCache new dev cache
func NewDevCache(ids []int32) *DevCache {
	devMap := make(map[int32]*devInfo)
	for _, id := range ids {
		devMap[id] = &devInfo{
			CtrIds: []string{},
			Status: common.StatusIgnorePause,
		}
	}
	return &DevCache{
		devInfoMap: devMap,
		mutex:      sync.Mutex{},
	}
}

// ResetDevStatus reset dev status
func (dc *DevCache) ResetDevStatus() {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	for _, info := range dc.devInfoMap {
		info.Status = common.StatusIgnorePause
	}
}

// SetCtrRelatedInfo set ctr related info
func (dc *DevCache) SetCtrRelatedInfo(ctrId string, usedDevs []int32) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	for _, devId := range usedDevs {
		info, ok := dc.devInfoMap[devId]
		if !ok {
			dc.devInfoMap[devId] = &devInfo{
				CtrIds: []string{ctrId},
			}
			continue
		}
		info.CtrIds = append(info.CtrIds, ctrId)
		info.CtrIds = utils.RemoveDuplicates(info.CtrIds)
	}
}

// RemoveDeletedCtr remove deleted ctr
func (dc *DevCache) RemoveDeletedCtr(newCtrIds []string) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	for _, info := range dc.devInfoMap {
		info.CtrIds = utils.RemoveElementsNotInSecond(info.CtrIds, newCtrIds)
	}
}

// UpdateDevStatus update dev status
func (dc *DevCache) UpdateDevStatus(faultCache map[int32][]*common.DevFaultInfo) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	if faultCache == nil || len(faultCache) == 0 {
		return
	}
	for id, info := range dc.devInfoMap {
		faults, ok := faultCache[id]
		if !ok {
			info.Status = common.StatusIgnorePause
			continue
		}
		info.Status = common.GetDevStatus(faults)
	}
}

// GetNeedPausedCtr get ctrs used L2-L5 level fault devs
func (dc *DevCache) GetNeedPausedCtr(onRing bool) []string {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	var needPaused []string
	for _, info := range dc.devInfoMap {
		if info.Status != common.StatusNeedPause {
			continue
		}

		if !onRing {
			needPaused = append(needPaused, info.CtrIds...)
			continue
		}
		for _, idOnRing := range info.DevsOnRing {
			ringDevInfo, ok := dc.devInfoMap[idOnRing]
			if !ok {
				// unreached branch
				continue
			}
			needPaused = append(needPaused, ringDevInfo.CtrIds...)
		}
	}
	return utils.RemoveDuplicates(needPaused)
}

// SetDevStatus set dev status
func (dc *DevCache) SetDevStatus(id int32, status string) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	info, ok := dc.devInfoMap[id]
	if !ok {
		// unreached branch
		return
	}
	info.Status = status
}

// SetDevsOnRing set devs on ring
func (dc *DevCache) SetDevsOnRing(id int32, devsOnRing []int32) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	info, ok := dc.devInfoMap[id]
	if !ok {
		// unreached branch
		return
	}
	info.DevsOnRing = devsOnRing
}

// GetDevsRelatedCtrs get devs related ctrs
func (dc *DevCache) GetDevsRelatedCtrs(id int32) []string {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	info, ok := dc.devInfoMap[id]
	if !ok {
		// unreached branch
		return []string{}
	}
	return info.CtrIds
}

// DeepCopy deep copy fault cache
func (dc *DevCache) DeepCopy() (map[int32]*devInfo, error) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	result := new(map[int32]*devInfo)
	if err := common.DeepCopy(result, dc.devInfoMap); err != nil {
		return nil, err
	}
	return *result, nil
}
