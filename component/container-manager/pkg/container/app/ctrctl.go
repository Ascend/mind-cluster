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
	"fmt"

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
		_, codes, err := devmgr.DevMgr.GetDeviceErrCode(id)
		if err != nil || utils.Contains(common.GetNeedPauseCtrFaultLevels(), domain.GetFaultLevelByCode(codes)) {
			cm.devInfoMap.SetDevStatus(id, common.StatusNeedPause)
			isNeedPause = true
			continue
		}
		// update device status
		cm.devInfoMap.SetDevStatus(id, common.StatusIgnorePause)
	}
	return isNeedPause
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
