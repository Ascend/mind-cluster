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
	"sync/atomic"
	"time"

	"github.com/containerd/containerd"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
	"container-manager/pkg/coordinator/proto"
)

var ctrCache *CtrCache = nil
var initOnce sync.Once

// CtrCache ctr cache
type CtrCache struct {
	ctrInfoMap map[string]*ctrInfo
	// dataDirty reports whether container info changed since the coordinator
	// last checked it via HasDataChanged (see dataSyncLoop).
	dataDirty atomic.Bool
	mutex     sync.RWMutex
}

// JobInfo holds distributed-task metadata parsed from container labels.
type JobInfo struct {
	JobID         string
	JobReplica    int32
	EnableRecover bool
}

type ctrInfo struct {
	Id              string  // ctr id
	Ns              string  // ctr namespace
	UsedDevs        []int32 // ctr used dev phy id
	Status          string
	StatusStartTime int64
	CtrsOnRing      []string
	DetailedInfo    containerd.Container
	JobID           string
	JobReplica      int32
	EnableRecover   bool
	PausedByPeer    string
	Action          int // 0: none, 1: wait pause, 2: wait resume
}

const (
	PeerActionNone   int = 0
	PeerActionPause  int = 1
	PeerActionResume int = 2
)

// GetCtrInfo new ctr info
func GetCtrInfo() *CtrCache {
	initOnce.Do(
		func() {
			ctrCache = &CtrCache{
				ctrInfoMap: make(map[string]*ctrInfo),
				mutex:      sync.RWMutex{},
			}
		},
	)
	return ctrCache
}

// HasDataChanged reports whether container info changed since the last call
// and clears the flag. Polled by the coordinator's dataSyncLoop every
// EventSyncInterval to decide whether a sync is needed.
func (cc *CtrCache) HasDataChanged() bool {
	return cc.dataDirty.Swap(false)
}

// GetCtrUsedDevs get ctr used devs
func (cc *CtrCache) GetCtrUsedDevs(id string) []int32 {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
	info, ok := cc.ctrInfoMap[id]
	if !ok {
		return []int32{}
	}
	return append([]int32{}, info.UsedDevs...)
}

// SetCtrsStatus set ctrs status
func (cc *CtrCache) SetCtrsStatus(ctrId, status string) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	info, ok := cc.ctrInfoMap[ctrId]
	if !ok {
		return
	}
	info.Status = status
	info.StatusStartTime = time.Now().Unix()
	cc.updateStatusFile()
	cc.dataDirty.Store(true)
}

// GetCtrStatusAndStartTime get ctr status and start time
func (cc *CtrCache) GetCtrStatusAndStartTime(id string) (string, int64) {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
	info, ok := cc.ctrInfoMap[id]
	if !ok {
		return "", 0
	}
	return info.Status, info.StatusStartTime
}

// GetCtrsByStatus get ctrs whose status is in statusMap
func (cc *CtrCache) GetCtrsByStatus(statusMap map[string]struct{}) []string {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
	var ids []string
	for id, info := range cc.ctrInfoMap {
		if _, ok := statusMap[info.Status]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}

// SetCtrInfo set ctr info
func (cc *CtrCache) SetCtrInfo(ctrId, ns string, usedDevs []int32, job JobInfo) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	_, ok := cc.ctrInfoMap[ctrId]
	if !ok {
		isSupportRecover := common.ParamOption.CtrStrategy != common.NeverStrategy
		cc.ctrInfoMap[ctrId] = &ctrInfo{
			Id:              ctrId,
			Ns:              ns,
			Status:          common.StatusRunning,
			StatusStartTime: time.Now().Unix(),
			UsedDevs:        append([]int32{}, usedDevs...),
			JobID:           job.JobID,
			JobReplica:      job.JobReplica,
			EnableRecover:   job.EnableRecover && isSupportRecover,
		}
		hwlog.RunLog.Infof("add container info, ctr: %s, UsedDevs: %v, jobInfo %v", ctrId, usedDevs, job)
		cc.dataDirty.Store(true)
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
		info.CtrsOnRing = append([]string{}, ctrsOnRing...)
	}
}

// GetCtrsOnRing get ctrs on ring
func (cc *CtrCache) GetCtrsOnRing(id string) []string {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
	info, ok := cc.ctrInfoMap[id]
	if !ok {
		return []string{}
	}
	return append([]string{}, info.CtrsOnRing...)
}

// GetCtrNs get ctr ns
func (cc *CtrCache) GetCtrNs(id string) string {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
	info, ok := cc.ctrInfoMap[id]
	if !ok {
		return ""
	}
	return info.Ns
}

// GetCtrRelatedDevs get ctr used devs
func (cc *CtrCache) GetCtrRelatedDevs(ctrIds []string) []int32 {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
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
	isDeleted := false
	for id, info := range cc.ctrInfoMap {
		if utils.Contains(newCtrIds, id) {
			newCtrInfoMap[id] = info
		} else {
			isDeleted = true
			hwlog.RunLog.Infof("container %s has been deleted", id)
		}
	}
	cc.ctrInfoMap = newCtrInfoMap
	cc.updateStatusFile()
	if isDeleted {
		cc.dataDirty.Store(true)
	}
}

// DeepCopy deep copy fault cache
func (cc *CtrCache) DeepCopy() (map[string]*ctrInfo, error) {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
	result := new(map[string]*ctrInfo)
	if err := common.DeepCopy(result, cc.ctrInfoMap); err != nil {
		return nil, err
	}
	return *result, nil
}

// UpdateCtrPeerMark gates the container by the peer node that issued the stop
// request
func (cc *CtrCache) UpdateCtrPeerMark(id, peerNodeID string, action int) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	info, ok := cc.ctrInfoMap[id]
	if !ok {
		return
	}
	info.PausedByPeer = peerNodeID
	info.Action = action
	cc.dataDirty.Store(true)
}

// GetCtrPausedByPeer returns the peer node that stopped the container.
func (cc *CtrCache) GetCtrPausedByPeer(id string) string {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
	info, ok := cc.ctrInfoMap[id]
	if !ok {
		return ""
	}
	return info.PausedByPeer
}

// PeerGateGroups groups gated containers by their peer gate state.
type PeerGateGroups struct {
	None   []string // PausedByPeer=="" && Action==PeerActionNone:  no gate, not gated by any peer
	Pause  []string // PausedByPeer!="" && Action==PeerActionPause:  waiting to be paused
	Resume []string // PausedByPeer!="" && Action==PeerActionResume: waiting to be resumed
}

// GetCtrsByPeerGate returns the ids of all containers grouped by their peer
// gate state. Only containers carrying a job (JobID != "") are considered.
// Callers filter by container status as needed.
func (cc *CtrCache) GetCtrsByPeerGate() *PeerGateGroups {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
	groups := &PeerGateGroups{}
	for id, info := range cc.ctrInfoMap {
		if info.JobID == "" {
			continue
		}
		switch {
		case info.PausedByPeer == "" && info.Action == PeerActionNone:
			groups.None = append(groups.None, id)
		case info.PausedByPeer != "" && info.Action == PeerActionPause:
			groups.Pause = append(groups.Pause, id)
		case info.PausedByPeer != "" && info.Action == PeerActionResume:
			groups.Resume = append(groups.Resume, id)
		}
	}
	return groups
}

func (cc *CtrCache) IsCtrsRecoverable(ctrIds []string) bool {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
	for _, id := range ctrIds {
		if info, ok := cc.ctrInfoMap[id]; ok && !info.EnableRecover {
			return false
		}
	}
	return true
}

func (cc *CtrCache) GetJobInfo(ctrId string) *JobInfo {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
	info, ok := cc.ctrInfoMap[ctrId]
	if !ok {
		return nil
	}
	return &JobInfo{
		JobID:         info.JobID,
		JobReplica:    info.JobReplica,
		EnableRecover: info.EnableRecover,
	}
}

// GetCtrsByJob returns the container IDs belonging to any of the given jobs.
func (cc *CtrCache) GetCtrsByJob(jobIDs []string) []string {
	if len(jobIDs) == 0 {
		return nil
	}
	jobSet := make(map[string]struct{}, len(jobIDs))
	for _, j := range jobIDs {
		jobSet[j] = struct{}{}
	}
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
	var ids []string
	for id, info := range cc.ctrInfoMap {
		if info.JobID == "" {
			continue
		}
		if _, ok := jobSet[info.JobID]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}

// GetDistributedSnapshots returns snapshots of containers
func (cc *CtrCache) GetDistributedSnapshots() []*proto.ContainerInfo {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
	var snaps []*proto.ContainerInfo
	for _, info := range cc.ctrInfoMap {
		if info.JobID == "" {
			continue
		}
		snaps = append(snaps, &proto.ContainerInfo{
			ContainerId:   info.Id,
			Status:        info.Status,
			PhyIds:        append([]int32{}, info.UsedDevs...),
			JobId:         info.JobID,
			JobReplica:    info.JobReplica,
			EnableRecover: info.EnableRecover,
			PausedByPeer:  info.PausedByPeer,
		})
	}
	return snaps
}
