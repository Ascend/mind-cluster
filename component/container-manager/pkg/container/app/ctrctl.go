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

// Package app container controller struct
package app

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/docker/docker/api/types"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
	ctrdomain "container-manager/pkg/container/domain"
	"container-manager/pkg/coordinator/proto"
	"container-manager/pkg/devmgr"
	"container-manager/pkg/fault/domain"
	resetdomain "container-manager/pkg/reset/domain"
)

func (cm *CtrCtl) initAndControl() {
	if err := cm.updateCtrRelatedInfo(); err != nil {
		hwlog.RunLog.Errorf("init ctr related info failed, error: %v", err)
		return
	}
	if err := cm.initRingInfo(); err != nil {
		if common.ParamOption.CtrStrategy == common.RingStrategy {
			hwlog.RunLog.Errorf("init ring info failed, error: %v", err)
			return
		}
		hwlog.RunLog.Warnf("init ring info failed, error: %v", err)
	}
	cm.ctrControl()
	cm.devInfoMap.ResetDevStatus()
}

func (cm *CtrCtl) updateForContainerd(cs map[string][]containerd.Container) []string {
	var ctrIds []string
	for ns, containers := range cs {
		ctx := namespaces.WithNamespace(context.Background(), ns)
		for _, containerObj := range containers {
			ctrIds = append(ctrIds, containerObj.ID())
			usedDevs, err := cm.client.getUsedDevs(containerObj, ctx)
			if err != nil {
				hwlog.RunLog.Errorf("get container %s used devs failed: %v", containerObj.ID(), err)
				continue
			}
			if len(usedDevs) == 0 {
				// only ctr of used dev need save to cache
				continue
			} else {
				hwlog.RunLog.Debugf("container %s,%s used devs: %v", ns, containerObj.ID(), usedDevs)
			}
			job := cm.client.getJobInfo(containerObj, ctx)
			cm.setCtrRelatedInfo(containerObj.ID(), ns, usedDevs, job)
		}
	}
	return ctrIds
}

func (cm *CtrCtl) updateCtrRelatedInfo() error {
	ctrs, err := cm.client.getAllContainers()
	if err != nil {
		return fmt.Errorf("get all ctrs failed: %v", err)
	}
	var ctrIds []string
	switch cs := ctrs.(type) {
	case map[string][]containerd.Container:
		ctrIds = cm.updateForContainerd(cs)
	case []types.Container:
		for _, containerObj := range cs {
			ctrIds = append(ctrIds, containerObj.ID)
			usedDevs, err := cm.client.getUsedDevs(containerObj, nil)
			if err != nil {
				hwlog.RunLog.Errorf("get container %s used devs failed: %v", containerObj.ID, err)
				continue
			}
			if len(usedDevs) == 0 {
				// only ctr of used dev need save to cache
				continue
			} else {
				hwlog.RunLog.Debugf("container %s used devs: %v", containerObj.ID, usedDevs)
			}
			job := cm.client.getJobInfo(containerObj, nil)
			cm.setCtrRelatedInfo(containerObj.ID, "default", usedDevs, job)
		}
	default:
		return nil
	}
	cm.removeDeletedCtr(ctrIds)
	return nil
}

func (cm *CtrCtl) ctrControl() {
	switch common.ParamOption.CtrStrategy {
	case common.NeverStrategy:
		return
	case common.SingleStrategy:
		cm.pauseCtr(false)
		cm.resumeCtr(false)
	case common.RingStrategy:
		cm.pauseCtr(true)
		cm.resumeCtr(true)
	default:
		hwlog.RunLog.Debugf("unknown ctr strategy: %s", common.ParamOption.CtrStrategy)
	}
}

// isDevsNeedPause if need pause, so cannot resume
func (cm *CtrCtl) isDevsNeedPause(usedDevs []int32) bool {
	var isNeedPause bool
	for _, id := range usedDevs {
		if cm.isSingleDevNeedPause(id) {
			cm.devInfoMap.SetDevStatus(id, common.StatusNeedPause)
			isNeedPause = true
			continue
		}
		// update device status
		cm.devInfoMap.SetDevStatus(id, common.StatusIgnorePause)
	}
	return isNeedPause
}

func (cm *CtrCtl) isSingleDevNeedPause(id int32) bool {
	// if device is in resetting, need pause
	if resetdomain.GetNpuInResetCache().IsNpuInReset(id) {
		return true
	}
	// if device have any fault, or get fault failed, need pause
	_, codes, err := devmgr.DevMgr.GetDeviceErrCode(id)
	if err != nil || utils.Contains(common.GetNeedPauseCtrFaultLevels(), domain.GetFaultLevelByCode(codes)) {
		return true
	}
	return false
}

func (cm *CtrCtl) setCtrRelatedInfo(ctrId, ns string, usedDevs []int32, job ctrdomain.JobInfo) {
	cm.ctrInfoMap.SetCtrInfo(ctrId, ns, usedDevs, job)
	cm.devInfoMap.SetCtrRelatedInfo(ctrId, usedDevs)
}

func (cm *CtrCtl) removeDeletedCtr(newCtrIds []string) {
	cm.ctrInfoMap.RemoveDeletedCtr(newCtrIds)
	cm.devInfoMap.RemoveDeletedCtr(newCtrIds)
}

func (cm *CtrCtl) initRingInfo() error {
	devInfos, err := cm.devInfoMap.DeepCopy()
	if err != nil {
		return fmt.Errorf("deep copy dev info in cache failed: %v", err)
	}
	for id := range devInfos {
		devsOnRing, err := devmgr.DevMgr.GetPhyIdOnRing(id)
		if err != nil {
			return fmt.Errorf("failed to get dev ids on ring for %d: %v", id, err)
		}
		var ctrsOnRing []string
		for _, devId := range devsOnRing {
			cm.devInfoMap.SetDevsOnRing(devId, devsOnRing)
			ctrsOnRing = append(ctrsOnRing, cm.devInfoMap.GetDevsRelatedCtrs(devId)...)
		}
		cm.ctrInfoMap.SetCtrsOnRing(utils.RemoveDuplicates(ctrsOnRing))
	}
	return nil
}

func (cm *CtrCtl) pauseCtr(onRing bool) {
	pausedCtrList := cm.ctrInfoMap.GetCtrsByStatus(map[string]struct{}{common.StatusPaused: {}})
	pausedCtrs := sets.NewString(pausedCtrList...)

	// pause ctrs that are gated by a peer and still need pausing
	ctrsPauseByPeer := cm.ctrInfoMap.GetCtrsByPeerGate().Pause
	ctrsPauseByPeer = utils.RemoveEleSli(ctrsPauseByPeer, pausedCtrList)
	if len(ctrsPauseByPeer) > 0 {
		hwlog.RunLog.Infof("pause ctrs by peer: %v", ctrsPauseByPeer)
		cm.doPauseCtrs(ctrsPauseByPeer)
		pausedCtrs.Insert(ctrsPauseByPeer...)
	}

	// find ctr groups that need pause, grouped by faulted device (or its ring)
	ctrGroups := cm.devInfoMap.GetNeedPausedCtr(onRing)
	hwlog.RunLog.Debugf("need pause ctr groups: %v", ctrGroups)
	// pause groups one by one, deduping across groups within one cycle
	for _, ctrs := range ctrGroups {
		var ctrNeedPause []string
		for _, ctrId := range ctrs {
			if pausedCtrs.Has(ctrId) {
				continue
			}
			ctrNeedPause = append(ctrNeedPause, ctrId)
		}
		if len(ctrNeedPause) == 0 || !cm.ctrInfoMap.IsCtrsRecoverable(ctrNeedPause) {
			hwlog.RunLog.Infof("ctrs %v are not recoverable, skip pausing group", ctrNeedPause)
			continue
		}
		if !cm.pauseCtrsInGroups(ctrNeedPause) {
			continue
		}
		pausedCtrs.Insert(ctrNeedPause...)
	}
}

func (cm *CtrCtl) pauseCtrsInGroups(ctrsRelated []string) bool {
	_, distCtrs, jobs := cm.partitionCtrs(ctrsRelated)
	if len(jobs) > 0 {
		if err := cm.coord.RequestStopJobs(jobs, distCtrs); err != nil {
			hwlog.RunLog.Errorf("RequestStop for jobs %v failed: %v", jobs, err)
			return false
		}
	}
	cm.doPauseCtrs(ctrsRelated)
	return true
}

func (cm *CtrCtl) doPauseCtrs(ids []string) {
	for _, id := range ids {
		hwlog.RunLog.Infof("start pausing container: %s", id)
		cm.ctrInfoMap.SetCtrsStatus(id, common.StatusPausing)
		ns := cm.ctrInfoMap.GetCtrNs(id)
		if ns == "" {
			hwlog.RunLog.Errorf("failed to get namespace of container: %s", id)
			continue
		}
		if err := cm.client.doStop(id, ns); err != nil {
			hwlog.RunLog.Errorf("pause container %s failed, error: %v", id, err)
			continue
		}
		hwlog.RunLog.Infof("successfully pause container: %s", id)
		cm.ctrInfoMap.SetCtrsStatus(id, common.StatusPaused)
	}
}

func (cm *CtrCtl) getNeedResumeCtrInGroups(id string, onRing bool) []string {
	if !onRing {
		if cm.isDevsNeedPause(cm.ctrInfoMap.GetCtrUsedDevs(id)) {
			return nil
		}
		return []string{id}
	}
	ctrsOnRings := cm.ctrInfoMap.GetCtrsOnRing(id)
	// can all containers on the ring be resumed.
	// as long as one of the cards used by the containers on the ring does not meet the condition,
	// the entire container on the ring cannot be resumed
	ringCtrsUsedDevs := cm.ctrInfoMap.GetCtrRelatedDevs(ctrsOnRings)
	if cm.isDevsNeedPause(cm.devInfoMap.GetDevsOnRing(ringCtrsUsedDevs)) {
		return nil
	}
	return ctrsOnRings
}

func (cm *CtrCtl) resumeCtr(onRing bool) {
	ctrHasPaused := cm.ctrInfoMap.GetCtrsByStatus(map[string]struct{}{common.StatusPaused: {}, common.StatusResuming: {}})
	gate := cm.ctrInfoMap.GetCtrsByPeerGate()
	excludeCtrs := sets.NewString(gate.Pause...)

	for _, id := range ctrHasPaused {
		if excludeCtrs.Has(id) {
			continue
		}
		ctrNeedResume := cm.getNeedResumeCtrInGroups(id, onRing)
		var newCtrNeedResume []string
		for _, ctrId := range ctrNeedResume {
			if !excludeCtrs.Has(ctrId) && utils.Contains(ctrHasPaused, ctrId) {
				newCtrNeedResume = append(newCtrNeedResume, ctrId)
			}
		}
		if len(newCtrNeedResume) == 0 {
			continue
		}
		cm.resumeCtrsInGroups(newCtrNeedResume, gate)
		excludeCtrs.Insert(newCtrNeedResume...)
	}
}

func (cm *CtrCtl) resumeCtrsInGroups(ctrNeedResume []string, gate *ctrdomain.PeerGateGroups) {
	localCtrs, distCtrs, jobs := cm.partitionCtrs(ctrNeedResume)

	var distCtrsNeedReq, jobNeedReq []string
	for _, ctrId := range distCtrs {
		if utils.Contains(gate.None, ctrId) {
			distCtrsNeedReq = append(distCtrsNeedReq, ctrId)
			jobNeedReq = append(jobNeedReq, cm.ctrInfoMap.GetJobInfo(ctrId).JobID)
		}
	}

	if len(jobNeedReq) > 0 {
		if err := cm.coord.RequestStartJobs(jobNeedReq, distCtrsNeedReq); err != nil {
			hwlog.RunLog.Errorf("RequestStart for jobs %v failed: %v; distributed containers skipped this cycle", jobs, err)
		} else {
			cm.doResumeCtrs(distCtrsNeedReq)
		}
	}
	cm.doResumeCtrs(utils.RemoveEleSli(distCtrs, distCtrsNeedReq))
	cm.doResumeCtrs(localCtrs)
}

func (cm *CtrCtl) doResumeCtrs(ids []string) {
	for _, id := range ids {
		hwlog.RunLog.Infof("start resuming container: %s", id)
		cm.ctrInfoMap.SetCtrsStatus(id, common.StatusResuming)
		ns := cm.ctrInfoMap.GetCtrNs(id)
		if ns == "" {
			hwlog.RunLog.Errorf("failed to get namespace of container: %s", id)
			continue
		}
		if err := cm.client.doStart(id, ns); err != nil {
			hwlog.RunLog.Errorf("resume container %s failed, error: %v", id, err)
			continue
		}
		hwlog.RunLog.Infof("successfully resume container: %s", id)
		cm.ctrInfoMap.SetCtrsStatus(id, common.StatusRunning)
		cm.ctrInfoMap.UpdateCtrPeerMark(id, "", ctrdomain.PeerActionNone)
	}
}

func (cm *CtrCtl) partitionCtrs(ctrIds []string) ([]string, []string, []string) {
	var (
		localCtrId       []string
		distributedCtrId []string
		jobs             []string
	)
	for _, ctrId := range utils.RemoveDuplicates(ctrIds) {
		jobInfo := cm.ctrInfoMap.GetJobInfo(ctrId)
		if jobInfo == nil {
			continue
		}
		// if JobID is empty or JobReplica is 1, it is a local container.
		// JobReplica is not mandatory, this field may be 0 for all distributed ctr
		// and still requires a request coordinator
		if jobInfo.JobID == "" || jobInfo.JobReplica == 1 {
			localCtrId = append(localCtrId, ctrId)
			continue
		}
		distributedCtrId = append(distributedCtrId, ctrId)
		jobs = append(jobs, jobInfo.JobID)
	}
	return localCtrId, distributedCtrId, utils.RemoveDuplicates(jobs)
}

// GetLocalContainers implements coordinator.ContainerOps.
func (cm *CtrCtl) GetLocalContainers() []*proto.ContainerInfo {
	return cm.ctrInfoMap.GetDistributedSnapshots()
}

// HasDataChanged implements coordinator.ContainerOps.
func (cm *CtrCtl) HasDataChanged() bool {
	return cm.ctrInfoMap.HasDataChanged()
}

// PauseJobContainers gates all local containers of the given jobs by the initiating peer
func (cm *CtrCtl) PauseJobContainers(jobIDs, faultCtrIds []string, peerNodeID string) error {
	ctrIds := cm.ctrInfoMap.GetCtrsByJob(jobIDs)
	validCtrs := make([]string, 0, len(ctrIds))
	for _, ctrId := range ctrIds {
		if utils.Contains(faultCtrIds, ctrId) {
			hwlog.RunLog.Infof("container %s is faulted, skip update peer mark", ctrId)
			continue
		}
		gate := cm.ctrInfoMap.GetCtrPausedByPeer(ctrId)
		if gate != "" && gate != peerNodeID {
			return fmt.Errorf("stop initiator %s does not match stop initiator %s (container %s)", peerNodeID, gate, ctrId)
		}
		status, _ := cm.ctrInfoMap.GetCtrStatusAndStartTime(ctrId)
		if status == common.StatusResuming {
			return fmt.Errorf("container %s is resuming, stop by %s conflicts; retry next cycle", ctrId, peerNodeID)
		}
		validCtrs = append(validCtrs, ctrId)
	}
	for _, ctrId := range validCtrs {
		cm.ctrInfoMap.UpdateCtrPeerMark(ctrId, peerNodeID, ctrdomain.PeerActionPause)
	}
	return nil
}

// ResumeJobContainers releases the peer gate on all local containers of the given jobs
func (cm *CtrCtl) ResumeJobContainers(jobIDs, faultCtrIds []string, peerNodeID string) error {
	ctrIds := cm.ctrInfoMap.GetCtrsByJob(jobIDs)
	validCtrs := make([]string, 0, len(ctrIds))
	for _, ctrId := range ctrIds {
		if utils.Contains(faultCtrIds, ctrId) {
			hwlog.RunLog.Infof("container %s is faulted, skip update peer mark", ctrId)
			continue
		}
		gate := cm.ctrInfoMap.GetCtrPausedByPeer(ctrId)
		if gate != "" && gate != peerNodeID {
			return fmt.Errorf("start initiator %s does not match stop initiator %s (container %s)", peerNodeID, gate, ctrId)
		}
		status, _ := cm.ctrInfoMap.GetCtrStatusAndStartTime(ctrId)
		switch status {
		case common.StatusPaused:
			if gate == "" {
				hwlog.RunLog.Warnf("container %s paused without peer gate, follow peer %s to resume", ctrId, peerNodeID)
			}
			validCtrs = append(validCtrs, ctrId)
		case common.StatusRunning, common.StatusResuming:
		default:
			return fmt.Errorf("container %s not paused yet (status %s), stop by %s has not completed", ctrId, status, gate)
		}
	}
	for _, ctrId := range validCtrs {
		cm.ctrInfoMap.UpdateCtrPeerMark(ctrId, peerNodeID, ctrdomain.PeerActionResume)
	}
	return nil
}
