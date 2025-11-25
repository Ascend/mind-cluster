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

// Package domain container info struct
package domain

import (
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
)

// CtrCache ctr cache
type CtrCache struct {
	ctrInfoMap map[string]*ctrInfo
	mutex      sync.Mutex
}

type ctrInfo struct {
	Id              string
	Ns              string
	UsedDevs        []int32 // phy id
	Status          string
	StatusStartTime int64
	CtrsOnRing      []string
}

// NewCtrInfo new ctr info
func NewCtrInfo() *CtrCache {
	return &CtrCache{
		ctrInfoMap: make(map[string]*ctrInfo),
		mutex:      sync.Mutex{},
	}
}

// GetCtrUsedDevs get ctr used devs
func (cc *CtrCache) GetCtrUsedDevs(id string) []int32 {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	info, ok := cc.ctrInfoMap[id]
	if !ok {
		return []int32{}
	}
	return info.UsedDevs
}

// SetCtrsStatus set ctrs status
func (cc *CtrCache) SetCtrsStatus(ctrId string, status string) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	info, ok := cc.ctrInfoMap[ctrId]
	if !ok {
		return
	}
	info.Status = status
	info.StatusStartTime = time.Now().Unix()
	cc.updateStatusFile()
}

// GetCtrStatusAndStartTime get ctr status and start time
func (cc *CtrCache) GetCtrStatusAndStartTime(id string) (string, int64) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	info, ok := cc.ctrInfoMap[id]
	if !ok {
		return "", 0
	}
	return info.Status, info.StatusStartTime
}

// GetCtrsByStatus get ctrs by status
func (cc *CtrCache) GetCtrsByStatus(status string) []string {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	var ids []string
	for id, info := range cc.ctrInfoMap {
		if info.Status == status {
			ids = append(ids, id)
		}
	}
	return ids
}

// SetCtrInfo set ctr info
func (cc *CtrCache) SetCtrInfo(ctrId, ns string, usedDevs []int32) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	_, ok := cc.ctrInfoMap[ctrId]
	if !ok {
		cc.ctrInfoMap[ctrId] = &ctrInfo{
			Id:       ctrId,
			Ns:       ns,
			Status:   common.StatusRunning,
			UsedDevs: usedDevs,
		}
	}
	cc.updateStatusFile()
}

// SetCtrsOnRing set ctrs on ring
func (cc *CtrCache) SetCtrsOnRing(ctrsOnRing []string) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	for _, ctrId := range ctrsOnRing {
		info, ok := cc.ctrInfoMap[ctrId]
		if !ok {
			// unreached branch
			continue
		}
		info.CtrsOnRing = ctrsOnRing
	}
}

// GetCtrsOnRing get ctrs on ring
func (cc *CtrCache) GetCtrsOnRing(id string) []string {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	info, ok := cc.ctrInfoMap[id]
	if !ok {
		return []string{}
	}
	return info.CtrsOnRing
}

// GetCtrNs get ctr ns
func (cc *CtrCache) GetCtrNs(id string) string {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	info, ok := cc.ctrInfoMap[id]
	if !ok {
		return ""
	}
	return info.Ns
}

// GetCtrRelatedDevs get ctr used devs
func (cc *CtrCache) GetCtrRelatedDevs(ctrIds []string) []int32 {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	var usedDevs []int32
	for id, info := range cc.ctrInfoMap {
		if utils.Contains(ctrIds, id) {
			usedDevs = append(usedDevs, info.UsedDevs...)
		}
	}
	return utils.RemoveDuplicates(usedDevs)
}

// RemoveDeletedCtr remove deleted ctr
func (cc *CtrCache) RemoveDeletedCtr(newCtrIds []string) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	var newCtrInfoMap = make(map[string]*ctrInfo)
	for id, info := range cc.ctrInfoMap {
		if utils.Contains(newCtrIds, id) {
			newCtrInfoMap[id] = info
		} else {
			hwlog.RunLog.Infof("container %s has been deleted", id)
		}
	}
	cc.ctrInfoMap = newCtrInfoMap
	cc.updateStatusFile()
}

// DeepCopy deep copy fault cache
func (cc *CtrCache) DeepCopy() (map[string]*ctrInfo, error) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	result := new(map[string]*ctrInfo)
	if err := common.DeepCopy(result, cc.ctrInfoMap); err != nil {
		return nil, err
	}
	return *result, nil
}
