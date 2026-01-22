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

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
	"container-manager/pkg/devmgr"
	"container-manager/pkg/fault/domain"
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
			cm.setCtrRelatedInfo(containerObj.ID(), ns, usedDevs)
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
			cm.setCtrRelatedInfo(containerObj.ID, "default", usedDevs)
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
	// if device have any fault, or get fault failed, need pause
	_, codes, err := devmgr.DevMgr.GetDeviceErrCode(id)
	if err != nil || utils.Contains(common.GetNeedPauseCtrFaultLevels(), domain.GetFaultLevelByCode(codes)) {
		return true
	}
	return false
}

func (cm *CtrCtl) setCtrRelatedInfo(ctrId, ns string, usedDevs []int32) {
	cm.ctrInfoMap.SetCtrInfo(ctrId, ns, usedDevs)
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
	ctrNeedPaused := cm.devInfoMap.GetNeedPausedCtr(onRing)
	ctrHasPaused := cm.ctrInfoMap.GetCtrsByStatus(common.StatusPaused)
	needPaused := utils.RemoveEleSli(ctrNeedPaused, ctrHasPaused)
	for _, id := range needPaused {
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

func (cm *CtrCtl) resumeCtr(onRing bool) {
	ctrHasPaused := cm.ctrInfoMap.GetCtrsByStatus(common.StatusPaused)
	var ctrNeedResume []string
	for _, id := range ctrHasPaused {
		if !onRing {
			if cm.isDevsNeedPause(cm.ctrInfoMap.GetCtrUsedDevs(id)) {
				continue
			}
			ctrNeedResume = append(ctrNeedResume, id)
			continue
		}
		if utils.Contains(ctrNeedResume, id) {
			continue
		}
		ctrsOnRings := cm.ctrInfoMap.GetCtrsOnRing(id)
		// can all containers on the ring be resumed.
		// as long as one of the cards used by the containers on the ring does not meet the condition,
		// the entire container on the ring cannot be resumed
		ringCtrsUsedDevs := cm.ctrInfoMap.GetCtrRelatedDevs(ctrsOnRings)
		if cm.isDevsNeedPause(cm.devInfoMap.GetDevsOnRing(ringCtrsUsedDevs)) {
			continue
		}
		for _, ctrId := range ctrsOnRings {
			if utils.Contains(ctrHasPaused, ctrId) {
				ctrNeedResume = append(ctrNeedResume, ctrId)
			}
		}
	}

	for _, id := range utils.RemoveDuplicates(ctrNeedResume) {
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
	}
}
