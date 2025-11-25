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

// Package app realize npu reset function
package app

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"ascend-common/common-utils/hwlog"
	devmanagercommon "ascend-common/devmanager/common"
	"container-manager/pkg/common"
	containerdomain "container-manager/pkg/container/domain"
	"container-manager/pkg/devmgr"
	faultdomain "container-manager/pkg/fault/domain"
	"container-manager/pkg/reset/domain"
)

func (r *ResetMgr) processResetWork() {
	if !r.allowToResetNpu() {
		hwlog.RunLog.Debug("allowToResetNpu if false, end this loop")
		return
	}
	faults, err := getFaultCache()
	if err != nil {
		hwlog.RunLog.Errorf("get fault cache failed, error: %v", err)
		return
	}
	hwlog.RunLog.Debugf("current falut cache: %v", faults)
	r.resetResetCountCache(faults)
	rawNeedResetNpus := getNeedToHandleFaults(faults)
	if len(rawNeedResetNpus) == 0 {
		hwlog.RunLog.Debug("no fault need to reset, end loop")
		return
	}

	hwlog.RunLog.Infof("raw fault npus need to reset: %v", rawNeedResetNpus)
	needResetNpus := r.filterCountLimit(rawNeedResetNpus)
	hwlog.RunLog.Infof("filtered fault npus need to reset: %v", rawNeedResetNpus)
	if len(needResetNpus) == 0 {
		hwlog.RunLog.Info("no fault need to reset, end loop")
		return
	}
	relatedInfos := getAllRelatedNpus(needResetNpus)
	r.hotResetRelatedNpuArrays(relatedInfos)
}

func (r *ResetMgr) hotResetRelatedNpuArrays(relatedInfos []domain.ResetNpuInfos) {
	for _, info := range relatedInfos {
		hwlog.RunLog.Infof("fault npu pyhsic ID [%v], related ID %v, start to reset", info.FaultId, info.RelatedIds)
		if isNpuHoldByContainer(info.RelatedIds) {
			hwlog.RunLog.Infof("npus %v are hold by container, skip reset", info.RelatedIds)
			continue
		}
		isHold, checkErr := isNpuHoldByProcess(info.RelatedIds)
		if checkErr != nil {
			hwlog.RunLog.Infof("npus %v check process occupation failed, skip reset, error: %v",
				info.RelatedIds, checkErr)
			continue
		}
		if isHold {
			hwlog.RunLog.Infof("npus %v are hold by process, skip reset", info.RelatedIds)
			continue
		}

		if !isFaultExist(info.RelatedIds) {
			hwlog.RunLog.Infof("npus' %v fault is not exist, skip reset", info.RelatedIds)
			continue
		}
		hwlog.RunLog.Infof("npus' %v fault is exist, start to reset", info.RelatedIds)
		r.hotReset(info)
	}
}

func (r *ResetMgr) allowToResetNpu() bool {
	// check there is any npu is in resetting
	resettingCache := r.resetCache.DeepCopy()
	if len(resettingCache) != 0 {
		hwlog.RunLog.Debugf("exist npus in reset, skip this loop. npus: %v", resettingCache)
		return false
	}

	// check cooldown period
	if r.lastSuccessResetTime != nil && time.Now().Before(r.lastSuccessResetTime.Add(cooldownPeriod)) {
		hwlog.RunLog.Debugf("cooldown period, skip this loop. last successful reset time: %v", r.lastSuccessResetTime)
		return false
	}
	return true
}

// get the full fault cache
func getFaultCache() (map[int32]map[int64]map[string]*common.DevFaultInfo, error) {
	faultCache := faultdomain.GetFaultCache()
	return faultCache.DeepCopy()
}

// reset count of npu for previous faults that no longer exist
func (r *ResetMgr) resetResetCountCache(faultsMap map[int32]map[int64]map[string]*common.DevFaultInfo) {
	noFaultNpuSet := make(map[int32]struct{})
	for _, id := range r.countCache.GetAllFailedResetCountNpuId() {
		// get npu which no longer have any fault
		if _, ok := faultsMap[id]; !ok {
			noFaultNpuSet[id] = struct{}{}
		}
	}
	hwlog.RunLog.Debugf("no fault npu set: %v", noFaultNpuSet)
	for id, _ := range noFaultNpuSet {
		r.countCache.ClearFailedResetCount(id)
	}
}

// get chip that needs to be reset
func getNeedToHandleFaults(faults map[int32]map[int64]map[string]*common.DevFaultInfo) []int32 {
	var needResetNpuSet []int32
	for phyId, faultInfoMap := range faults {
		if !isFaultsNeedToHandle(faultInfoMap) {
			continue
		}
		needResetNpuSet = append(needResetNpuSet, phyId)
	}
	return needResetNpuSet
}

func isFaultsNeedToHandle(faultInfoMap map[int64]map[string]*common.DevFaultInfo) bool {
	for faultCode, moduleFaultInfoMap := range faultInfoMap {
		faultLevel := faultdomain.GetFaultLevelByCode([]int64{faultCode})

		if _, ok := faultsToHandleAtOnce[faultLevel]; ok {
			hwlog.RunLog.Infof("fault need to handle at once, code: %X, level: %v", faultCode, faultLevel)
			return true
		}

		if _, ok := faultsToHandleLastingOneMinute[faultLevel]; ok && checkLastingFaultNeedToReset(moduleFaultInfoMap) {
			hwlog.RunLog.Infof("lasting fault need to handle, code: %X, level: %v", faultCode, faultLevel)
			return true
		}
	}
	return false
}

// check whether there is an L2/L3 fault lasting for more than 60 seconds
func checkLastingFaultNeedToReset(faultInfo map[string]*common.DevFaultInfo) bool {
	for _, devFaultInfo := range faultInfo {
		if time.Now().Unix()-devFaultInfo.ReceiveTime > lastingToHandlePeriodInSeconds {
			return true
		}
	}
	return false
}

// filter npu with reset times exceeding the limit
func (r *ResetMgr) filterCountLimit(faultNpus []int32) []int32 {
	var filteredOutNpus []int32
	for _, phyId := range faultNpus {
		count := r.countCache.GetFailedResetCount(phyId)
		if count >= npuContinuouslyResetCountLimit {
			hwlog.RunLog.Warnf("npu [physic Id: %v] has been reset for %v times, remove from array", phyId, count)
			continue
		}
		filteredOutNpus = append(filteredOutNpus, phyId)
	}
	return filteredOutNpus
}

// get chip information associated with the current chip through DCMI for status statistics
func getAllRelatedNpus(faultNpus []int32) []domain.ResetNpuInfos {
	var infos []domain.ResetNpuInfos
	countedNpu := make(map[int32]struct{})
	for _, phyId := range faultNpus {
		// already been associated with the previous faulty npu, skipping
		if _, ok := countedNpu[phyId]; ok {
			continue
		}
		npuInfos := devmgr.DevMgr.GetNodeNPUInfo()
		if npuInfos == nil {
			hwlog.RunLog.Error("get npu infos failed, npu info set is empty")
			continue
		}
		npuInfo, ok := npuInfos[phyId]
		if !ok || npuInfo == nil || len(npuInfo.DevsOnRing) == 0 {
			hwlog.RunLog.Error("get all related npus failed, npu info is empty")
			continue
		}

		infos = append(infos, domain.ResetNpuInfos{FaultId: phyId, RelatedIds: npuInfo.DevsOnRing})
		for _, ringId := range npuInfo.DevsOnRing {
			countedNpu[ringId] = struct{}{}
		}
	}
	hwlog.RunLog.Infof("all related npus group infos: %v", infos)
	return infos
}

// check the container occupancy status of the fault-related chip through crtmgr
func isNpuHoldByContainer(phyIds []int32) bool {
	for _, phyId := range phyIds {
		cache := containerdomain.GetDevCache()
		if cache == nil {
			hwlog.RunLog.Infof("container cache is not ready yet, cannot reset physic ID [%v]", phyId)
			return true
		}
		containerIds := cache.GetDevsRelatedCtrs(phyId)
		if len(containerIds) != 0 {
			hwlog.RunLog.Infof("physic ID [%v] is hold by running container %v", phyId, containerIds)
			return true
		}
	}
	return false
}

// query whether the current chip is occupied by other processes through the DCMI interface
func isNpuHoldByProcess(phyIds []int32) (bool, error) {
	for _, phyId := range phyIds {
		info, err := devmgr.DevMgr.GetDmgr().GetDevProcessInfo(devmgr.DevMgr.GetLogicIdByPhyId(phyId))
		if err != nil {
			hwlog.RunLog.Errorf("get device process info failed, %v", err)
			return false, err
		}
		if info.ProcNum != 0 {
			hwlog.RunLog.Infof("physic ID [%v] is hold by running process %v", phyId, info.DevProcArray)
			return true, nil
		}
	}
	return false, nil
}

// check npu fault exist before do reset
func isFaultExist(relatedIds []int32) bool {
	for _, phyId := range relatedIds {
		_, errCodes, getErr := devmgr.DevMgr.GetDmgr().GetDeviceAllErrorCode(devmgr.DevMgr.GetLogicIdByPhyId(phyId))
		if getErr != nil {
			hwlog.RunLog.Errorf("failed to get device error code, err %v", getErr)
			// get error consider fault exist
			continue
		}
		if len(errCodes) == 0 {
			continue
		}
		faultLevel := faultdomain.GetFaultLevelByCode(errCodes)
		hwlog.RunLog.Infof("device fault exist, physic ID: [%v], fault codes: <%#v>, level: %v",
			phyId, errCodes, faultLevel)
		_, restartExist := faultsToHandleLastingOneMinute[faultLevel]
		_, resetExist := faultsToHandleAtOnce[faultLevel]
		if !restartExist && !resetExist {
			continue
		}
		return true
	}
	return false
}

func (r *ResetMgr) hotReset(info domain.ResetNpuInfos) {
	r.resetCache.SetNpuInReset(info.RelatedIds...)
	defer r.resetCache.ClearNpuInReset(info.RelatedIds...)

	if err := execDeviceReset(info.FaultId); err != nil {
		hwlog.RunLog.Errorf("reset device %v failed, error: %v", info.FaultId, err)
		r.countCache.SetFailedResetCount(info.FaultId, r.countCache.GetFailedResetCount(info.FaultId)+1)
		return
	}
	if err := getResetSuccessfulStatus(info); err != nil {
		hwlog.RunLog.Errorf("get reset result failed, error: %v", err)
		r.countCache.SetFailedResetCount(info.FaultId, r.countCache.GetFailedResetCount(info.FaultId)+1)
		return
	}

	r.countCache.ClearFailedResetCount(info.FaultId)
	timeNow := time.Now()
	r.lastSuccessResetTime = &timeNow
	hwlog.RunLog.Infof("physic ID [%v], related IDs %v, hot reset success", info.FaultId, info.RelatedIds)
}

func execDeviceReset(faultPhyId int32) error {
	const execHotResetMaxRetryTimes = 4
	var errorInfo error
	for i := 0; i < execHotResetMaxRetryTimes; i++ {
		cardID, deviceID, err := devmgr.DevMgr.GetDmgr().GetCardIDDeviceID(devmgr.DevMgr.GetLogicIdByPhyId(faultPhyId))
		if err != nil {
			hwlog.RunLog.Errorf("failed to get cardID and deviceID by logicID(%d)", faultPhyId)
			errorInfo = err
			continue
		}
		hwlog.RunLog.Infof("start device card(%d) and deviceID(%d) reset...", cardID, deviceID)
		if err := devmgr.DevMgr.GetDmgr().SetDeviceReset(cardID, deviceID); err != nil {
			errorInfo = err
			continue
		}
		return nil
	}
	return errorInfo
}

func getResetSuccessfulStatus(info domain.ResetNpuInfos) error {
	return wait.PollImmediate(time.Second, defaultWaitDeviceResetTime, func() (bool, error) {
		// check all device state that hot reset together
		for _, phyId := range info.RelatedIds {
			bootState, err := devmgr.DevMgr.GetDmgr().GetDeviceBootStatus(devmgr.DevMgr.GetLogicIdByPhyId(phyId))
			if err != nil {
				hwlog.RunLog.Errorf("get device %v boot status failed, err: %v", phyId, err)
				return false, err
			}
			if bootState != devmanagercommon.BootStartFinish {
				hwlog.RunLog.Warnf("device %v bootState(%d), starting...", phyId, bootState)
				return false, nil
			}
		}
		return true, nil
	})
}
