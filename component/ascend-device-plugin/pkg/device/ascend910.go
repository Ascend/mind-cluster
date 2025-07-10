/* Copyright(C) 2022-2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package device a series of device function
package device

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

const (
	networkDetectOK              = uint32(0)
	networkDetectInit            = uint32(6)
	synchronizeWaitMagnification = 3
	podDevStatusAnnotation       = "podDevStatus"
	updateResetCMFailedPattern   = "failed to update reset cm to recovered status, err: %v"
	unsetTaskFailedPattern       = "failed to unset task in reset, err: %v"
	failedToUpdateCmPattern      = "failed to update reset cm to recover failed status, err: %v"
	beforeRescanDelay            = 3 // seconds sleep before rescan devices
	afterRescanDelay             = 2 // seconds sleep after rescan devices
	deviceA3Id0                  = 0 // die id 0 of A3 card
	deviceA3Id1                  = 1 // dir id 1 of A3 card
	ringNumOfA3                  = 2 // device number in a ring
	otherCardIncrease            = 1
	errorId                      = -1
)

var (
	lastTimeNetworkRecoverDevices = sets.String{}
	hotResetManagerInitOnce       sync.Once
	resetDevMap                         = &sync.Map{}
	inResetDev                    int32 = -1
	isolateDevList                []int32
	isL3FaultExistMap             map[int32]bool = make(map[int32]bool, common.MaxDevicesNum)
	resetGoroutine                               = &sync.Map{}
	offlineInBandFailLogicId                     = sync.Map{}
)

// HwAscend910Manager manages huawei Ascend910 devices.
type HwAscend910Manager struct {
	AscendTools
	hotResetManager HotResetManager
}

// NewHwAscend910Manager is used to create ascend 910 manager
func NewHwAscend910Manager() *HwAscend910Manager {
	return &HwAscend910Manager{
		AscendTools: AscendTools{
			name:                      common.Ascend910,
			unHealthyKey:              common.HuaweiUnHealthAscend910,
			devCount:                  common.MaxDevicesNum,
			cardInResetMap:            make(map[int32]bool, common.GeneralMapSize),
			resetFailedTimesMap:       make(map[int32]int, common.GeneralMapSize),
			lastUsedChipsContainerMap: make(map[string]sets.String),
		},
	}
}

// GetNPUs Discovers all HUAWEI Ascend910 devices by call devmanager interface
// a physical npu can be split into multiple vNPU
// vNPU is classification by computing power, like Ascend910-4c, Ascend910-8c, Ascend910-16c
// physical npu sets corresponding to the deviTypes, and vNPU is vDeviTypes
// vDeviTypes may is: [Ascend910-4c, Ascend910-4c, Ascend910-8c], also deviTypes may is: [Ascend910, Ascend910]
// one class deviType will generate a socket file, like ascend910-4c.sock or Ascend910.sock, so we deduplicate
func (hnm *HwAscend910Manager) GetNPUs() (common.NpuAllInfo, error) {
	devNum, devList, err := hnm.dmgr.GetDeviceList()
	if err != nil {
		return common.NpuAllInfo{}, err
	}

	if devNum > hnm.devCount {
		return common.NpuAllInfo{}, fmt.Errorf("invalid device num: %d", devNum)
	}
	var allDevices []common.NpuDevice
	var aiCoreDevices []*common.NpuDevice
	var allDeviceTypes = make([]string, 0)
	for i := int32(0); i < devNum; i++ {
		davinCiDev, err := hnm.getDavinCiDev(devList[i])
		if err != nil {
			return common.NpuAllInfo{}, err
		}
		vDevInfos, err := hnm.getVirtualDevice(devList[i])
		if err != nil {
			hwlog.RunLog.Warnf("The virtual device is considered not exist, please check the error: %v", err)
		}
		if vDevInfos.TotalResource.VDevNum > common.MaxVirtualDeviceNum {
			return common.NpuAllInfo{}, fmt.Errorf("invalid virtual device count")
		}
		if !common.ParamOption.PresetVDevice {
			common.FakeAiCoreDevice(davinCiDev, &aiCoreDevices)
		}
		if vDevInfos.TotalResource.VDevNum == 0 {
			hnm.assemblePhyDevices(davinCiDev, &allDevices, &allDeviceTypes)
			continue
		}
		hnm.assembleVirtualDevices(davinCiDev, vDevInfos, &allDevices, &allDeviceTypes)
	}
	allDeviceTypes = hnm.removeDuplicate(&allDeviceTypes)
	return common.NpuAllInfo{AllDevs: allDevices, AICoreDevs: aiCoreDevices, AllDevTypes: allDeviceTypes}, nil
}

// GraceTolerance process training task with device fault gracefully
func (hnm *HwAscend910Manager) GraceTolerance(ctx context.Context, classifyDevs map[string][]*common.NpuDevice) {
	hotResetManagerInitOnce.Do(func() {
		hnm.hotResetManager = NewHotResetManager(hnm.GetDeviceUsage(), len(classifyDevs[common.Ascend910]), hnm.boardId)
		if hnm.hotResetManager == nil {
			hwlog.RunLog.Errorf("hot reset manager is nil, devType: %s", common.ParamOption.RealCardType)
			return
		}
		hnm.hotResetManager.SyncResetCM(ctx, hnm.GetKubeClient())
	})
	if !common.ParamOption.GraceToleranceOn {
		return
	}

	// obtain the current device status and update the cache of hot reset manager
	if err := hnm.updateHotResetCache(classifyDevs); err != nil {
		hwlog.RunLog.Errorf("failed to update hot reset cache, err: %v", err)
		return
	}
	// online recover will process fault by task but offline recover will not
	if common.ParamOption.HotReset == common.HotResetTrainOnLine {
		// performs graceful fault tolerance for tasks to be processed based on the device information in the cache
		if err := hnm.processAllTask(classifyDevs); err != nil {
			hwlog.RunLog.Errorf("failed to process task, err: %v", err)
		}
	}
	// handling hot reset without task
	go func() {
		if err := hnm.hotResetHandler(classifyDevs); err != nil {
			hwlog.RunLog.Errorf("hot reset err: %v", err)
		}
	}()
	// filter the faulty device in the reset state in the device info cm to avoid rescheduling
	if err := hnm.filterDevStatus(classifyDevs); err != nil {
		hwlog.RunLog.Errorf("failed to filter device status,err: %v", err)
	}
	// when hot reset is on, we update device info cm so that task could not be dispatched on resetting device
	if err := hnm.setAllDevUnhealthyOnRing(classifyDevs); err != nil {
		hwlog.RunLog.Errorf("set all device on reset status fail, err %v", err)
	}
}

// hotResetHandler handling hot reset
func (hnm *HwAscend910Manager) hotResetHandler(classifyDevs map[string][]*common.NpuDevice) error {
	deviceList, ok := classifyDevs[common.Ascend910]
	if !ok {
		return fmt.Errorf("device list not found, %v", common.Ascend910)
	}
	resetDevs := make([]*common.NpuDevice, 0, len(deviceList))
	resetFaultInfos := make([]*common.DevFaultInfo, 0, len(deviceList))
	resetRing := make(map[int32]struct{})
	for _, dev := range deviceList {
		if _, ok = resetDevMap.Load(dev.LogicID); ok {
			continue
		}
		tempFaultInfo := hnm.getDevFaultInfo(dev.LogicID)
		if tempFaultInfo == nil {
			continue
		}
		idx, err := hnm.getResetIndex(dev)
		if err != nil {
			continue
		}
		if _, exist := resetRing[idx]; exist {
			continue
		}
		if canReset, err := hnm.canBeReset(tempFaultInfo); err != nil || !canReset {
			hwlog.RunLog.Infof("device %v cannot reset, it is busy, err: %v", tempFaultInfo.LogicId, err)
			continue
		}
		if l3RestartFlag := hnm.handleL2L3FaultRestart(tempFaultInfo); !l3RestartFlag &&
			(tempFaultInfo.Policy != common.ResetError && tempFaultInfo.Policy != common.FreeResetError) {
			continue
		}
		resetRing[idx] = struct{}{}
		resetDevs = append(resetDevs, dev)
		resetFaultInfos = append(resetFaultInfos, tempFaultInfo)
		hwlog.RunLog.Debugf("found %v error on device %v, will start reset process "+
			"whenever all chips are free on ring", tempFaultInfo.Policy, dev.DeviceName)
	}
	if len(resetDevs) == 0 {
		return nil
	}
	return hnm.tryDeviceReset(classifyDevs, resetDevs, resetFaultInfos)
}

func (hnm *HwAscend910Manager) tryDeviceReset(classifyDevs map[string][]*common.NpuDevice,
	resetDevs []*common.NpuDevice, resetFaultInfos []*common.DevFaultInfo) error {
	var err error
	var wg sync.WaitGroup
	wg.Add(len(resetDevs))
	for idx, dev := range resetDevs {
		resetDevMap.Store(dev.LogicID, struct{}{})
		go func(idx int, dev *common.NpuDevice) {
			if err = hnm.startUpHotReset(classifyDevs, resetFaultInfos[idx], dev); err != nil {
				hwlog.RunLog.Errorf("failed to start up hot reset, err: %v", err)
			}
			wg.Done()
		}(idx, dev)
	}
	wg.Wait()
	hnm.hotResetTryOutBand(resetDevs)
	for _, dev := range resetDevs {
		resetDevMap.Delete(dev.LogicID)
	}
	return err
}

func (hnm *HwAscend910Manager) getDevFaultInfo(logicID int32) *common.DevFaultInfo {
	tempFaultInfo, tempErr := hnm.hotResetManager.GetGlobalDevFaultInfo(logicID)
	if tempErr != nil {
		hwlog.RunLog.Errorf("failed to get global device fault info from cache, err: %v", tempErr)
		return nil
	}
	if tempFaultInfo.Policy == common.EmptyError {
		return nil
	}
	return tempFaultInfo
}

func (hnm *HwAscend910Manager) getResetIndex(dev *common.NpuDevice) (int32, error) {
	if common.ParamOption.RealCardType == common.Ascend910A3 {
		if idx, err := hnm.getResetIndexForA3(dev.LogicID); err == nil {
			return idx, nil
		}
	}
	resetDevNumOnce, err := hnm.hotResetManager.GetResetDevNumOnce()
	if err != nil {
		hwlog.RunLog.Error(err)
		return errorId, err
	}
	// Get the first card in the same ring by dividing the logicID by the number of cards in the ring,
	// taking the integer part, and then multiplying it by the number of cards in the ring,
	// for instance, logicID 7, resetDevNumOnce 4, then return 4, because 7 is in [4 5 6 7],
	// and logicID 4 is the first in that ring
	return (dev.LogicID / int32(resetDevNumOnce)) * int32(resetDevNumOnce), nil
}

func (hnm *HwAscend910Manager) hotResetTryOutBand(devs []*common.NpuDevice) {
	sucDevs := make([]ResetDevice, 0, len(devs))
	failDevs := make([]ResetDevice, 0, len(devs))
	allDevs := make([]ResetDevice, 0, len(devs))
	for _, dev := range devs {
		if _, exist := offlineInBandFailLogicId.Load(dev.LogicID); !exist {
			sucDevs = append(sucDevs, npuDevToResetDev(*dev))
		} else {
			failDevs = append(failDevs, npuDevToResetDev(*dev))
		}
		allDevs = append(allDevs, npuDevToResetDev(*dev))
	}
	if common.ParamOption.RealCardType != common.Ascend910A3 {
		hnm.updateResetInfo(failDevs, sucDevs)
		return
	}
	if err := hnm.execOutBandReset(failDevs, sucDevs); err != nil {
		hwlog.RunLog.Errorf("hot reset out band failed, err: %v", err)
	}
}

func npuDevToResetDev(dev common.NpuDevice) ResetDevice {
	return ResetDevice{
		CardId:   dev.CardID,
		DeviceId: dev.DeviceID,
		LogicID:  dev.LogicID,
	}
}

// handleL2L3FaultRestart restarts when l2l3 faults handling failed
func (hnm *HwAscend910Manager) handleL2L3FaultRestart(devFaultInfo *common.DevFaultInfo) bool {
	if devFaultInfo.Policy == common.RestartError || devFaultInfo.Policy == common.RestartRequestError {
		existFlag, ok := isL3FaultExistMap[devFaultInfo.LogicId]
		if existFlag && ok {
			isL3FaultExistMap[devFaultInfo.LogicId] = false
			return true
		}
		isL3FaultExistMap[devFaultInfo.LogicId] = true
	} else {
		isL3FaultExistMap[devFaultInfo.LogicId] = false
	}
	return false
}

// startUpHotReset starts hot reset goroutine when chips are free
func (hnm *HwAscend910Manager) startUpHotReset(classifyDevs map[string][]*common.NpuDevice,
	tempFaultInfo *common.DevFaultInfo, dev *common.NpuDevice) error {
	hwlog.RunLog.Infof("start handling fault: %s", tempFaultInfo.Policy)
	inResetDev = tempFaultInfo.LogicId
	hnm.handleResetProcess(classifyDevs, tempFaultInfo, dev)
	return nil
}

// setAllDevUnhealthyOnRing change the npu health status to unhealthy for all device on ring
func (hnm *HwAscend910Manager) setAllDevUnhealthyOnRing(classifyDevs map[string][]*common.NpuDevice) error {
	devStatusList, ok := classifyDevs[common.Ascend910]
	if !ok {
		return fmt.Errorf("no ascend 910 device needed filter")
	}
	clearDeviceStatus(devStatusList)
	if common.GetSyncMapLen(resetDevMap) == 0 {
		return nil
	}
	if inResetDev == -1 {
		hwlog.RunLog.Debug("should not set device to unhealthy")
		return nil
	}
	if common.ParamOption.RealCardType == common.Ascend910A3 &&
		hnm.setUnhealthyForA3(devStatusList) == nil {
		return nil
	}
	resetDevNumOnce, err := hnm.hotResetManager.GetResetDevNumOnce()
	if err != nil {
		hwlog.RunLog.Error(err)
		return err
	}
	ringIndex := int(inResetDev) / resetDevNumOnce
	startDevIndex := ringIndex * resetDevNumOnce
	endDevIndex := startDevIndex + resetDevNumOnce
	for devIndex := startDevIndex; devIndex < endDevIndex; devIndex++ {
		devStatusList[devIndex].NetworkHealth = v1beta1.Unhealthy
		devStatusList[devIndex].Health = v1beta1.Unhealthy
		devStatusList[devIndex].Status = common.NPUResettingStatus
	}
	return nil
}

func (hnm *HwAscend910Manager) setUnhealthyForA3(devStatusList []*common.NpuDevice) error {
	// get busy status, get associated cards, set to unhealthy
	if int(inResetDev) >= len(devStatusList) || inResetDev < 0 {
		return fmt.Errorf("invalid in reset dev id %v", inResetDev)
	}
	dev := devStatusList[inResetDev]
	logicIdArr, err := hnm.getAssociatedLogicIDs(dev.LogicID, dev.CardID, dev.DeviceID)
	if err != nil {
		return err
	}
	for _, idx := range logicIdArr {
		if int(idx) >= len(devStatusList) || int(idx) < 0 {
			hwlog.RunLog.Errorf("device logicID %v is invalid, device list length is %v",
				idx, len(devStatusList))
			continue
		}
		devStatusList[idx].NetworkHealth = v1beta1.Unhealthy
		devStatusList[idx].Health = v1beta1.Unhealthy
		devStatusList[idx].Status = common.NPUResettingStatus
	}
	return nil
}

func (hnm *HwAscend910Manager) getAssociatedLogicIDs(logicID, cardID, deviceID int32) ([]int32, error) {
	associatedCardID, err := hnm.GetDmgr().GetBrotherCardID(cardID, deviceID)
	if err != nil {
		hwlog.RunLog.Debugf("get brother card failed, cardID %v deviceID %v, err: %v",
			cardID, deviceID, err)
		return nil, err
	}
	logicID0, err := hnm.GetDmgr().GetDeviceLogicID(associatedCardID, deviceA3Id0)
	if err != nil {
		hwlog.RunLog.Errorf("get logicID faild by cardID %v deviceID %v, err: %v",
			associatedCardID, deviceA3Id0, err)
		return nil, err
	}
	logicID1, err := hnm.GetDmgr().GetDeviceLogicID(associatedCardID, deviceA3Id1)
	if err != nil {
		hwlog.RunLog.Errorf("get logicID faild by cardID %v deviceID %v, err: %v",
			associatedCardID, deviceA3Id1, err)
		return nil, err
	}
	// get the other device id in a ring
	otherDeviceId := (deviceID + otherCardIncrease) % ringNumOfA3
	ringDevLogic, err := hnm.GetDmgr().GetDeviceLogicID(cardID, otherDeviceId)
	if err != nil {
		hwlog.RunLog.Errorf("get logicID faild by cardID %v deviceID %v, err: %v",
			cardID, otherDeviceId, err)
		return nil, err
	}
	return []int32{logicID, ringDevLogic, logicID0, logicID1}, nil
}

// clearDeviceStatus clear resetting device status
func clearDeviceStatus(devList []*common.NpuDevice) {
	for _, dev := range devList {
		dev.Status = common.NPUNormalStatus
	}
}

// handleResetProcess start handling hot reset process
func (hnm *HwAscend910Manager) handleResetProcess(classifyDevs map[string][]*common.NpuDevice,
	devInfo *common.DevFaultInfo, npuDev *common.NpuDevice) {
	haveErr := false
	defer func() {
		inResetDev = -1
	}()
	if common.ParamOption.HotReset == common.HotResetTrainOffLine &&
		common.ParamOption.RealCardType == common.Ascend910A3 {
		hwlog.RunLog.Infof("sleep before hotReset case hotReset=2, devInfo=%v", devInfo)
		time.Sleep(common.SleepMinutesForA3Reset * time.Minute)
	}
	if err := hnm.execHotReset(devInfo); err != nil {
		hwlog.RunLog.Errorf("execute hot reset failed, err %v", err)
		haveErr = true
	}
	isShouldUpgrade, err := hnm.refreshDevFaultInfoForResetProcess(devInfo)
	if err != nil {
		hwlog.RunLog.Errorf("failed refresh device fault info, err: %v", err)
		haveErr = true
	}
	if isShouldUpgrade || haveErr == true {
		hnm.upgradeHotResetError(classifyDevs, npuDev)
		return
	}
	common.SetDeviceInit(devInfo.LogicId)
}

func (hnm *HwAscend910Manager) upgradeHotResetError(classifyDevs map[string][]*common.NpuDevice,
	npuDev *common.NpuDevice) {
	isolateDevList = append(isolateDevList, npuDev.LogicID)
	devStatusList, ok := classifyDevs[common.Ascend910]
	if !ok {
		hwlog.RunLog.Error("no ascend 910 device, upgrade hot reset error fail")
		return
	}
	resetDevNumOnce, err := hnm.hotResetManager.GetResetDevNumOnce()
	if err != nil {
		hwlog.RunLog.Error(err)
		return
	}
	ringIndex := int(npuDev.LogicID) / resetDevNumOnce
	startDevIndex := ringIndex * resetDevNumOnce
	endDevIndex := startDevIndex + resetDevNumOnce
	for devIndex := startDevIndex; devIndex < endDevIndex; devIndex++ {
		tempFaultInfo, err := hnm.hotResetManager.GetGlobalDevFaultInfo(int32(devIndex))
		if err != nil {
			hwlog.RunLog.Errorf("failed to get global device fault info from cache device-%d, err: %v", devIndex, err)
			continue
		}
		if tempFaultInfo.Policy != common.EmptyError && tempFaultInfo.Policy != common.IgnoreError {
			continue
		}
		devStatusList[devIndex].Health = v1beta1.Healthy
		devStatusList[devIndex].NetworkHealth = v1beta1.Healthy
	}
	hwlog.RunLog.Infof("error upgrade to isolate: device-%v", npuDev.LogicID)
}

func (hnm *HwAscend910Manager) refreshDevFaultInfoForResetProcess(devInfo *common.DevFaultInfo) (bool, error) {
	_, errorCode, err := hnm.GetDmgr().GetDeviceAllErrorCode(devInfo.LogicId)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get err code of device %d", devInfo.LogicId)
		return true, err
	}
	if len(errorCode) == 0 {
		return false, nil
	}
	devInfo.Policy = hnm.hotResetManager.GetDevProcessPolicy(common.GetFaultType(errorCode, devInfo.LogicId))
	devInfo.ErrorCode = errorCode
	return true, nil
}

func (hnm *HwAscend910Manager) execHotReset(devInfo *common.DevFaultInfo) error {
	logicID := devInfo.LogicId
	cardId, deviceId, err := hnm.GetDmgr().GetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get reset device card id and device id, err %v", err)
		return err
	}
	shouldCheckNet := hnm.isShouldCheckNet(logicID)
	if err := hnm.tryResetDevice(cardId, deviceId); err != nil {
		offlineInBandFailLogicId.Store(logicID, struct{}{})
		hwlog.RunLog.Errorf("failed to reset device, err %v", err)
		return nil
	}
	if err := hnm.isRingResetComplete(logicID, shouldCheckNet); err != nil {
		hwlog.RunLog.Errorf("fail while waiting for hot reset complete, err %v", err)
		return err
	}
	FreeBusyDev(cardId, deviceId)
	hwlog.RunLog.Infof("hot reset complete, cardId: %d, logicId: %d", cardId, logicID)
	return nil
}

// isChipActive check if there is job on chip
func (hnm *HwAscend910Manager) isChipActive(logicID int32, busyChipList []string) (bool, error) {
	chipInfo, err := hnm.AscendTools.GetDmgr().GetDevProcessInfo(logicID)
	if err != nil || chipInfo == nil {
		hwlog.RunLog.Errorf("failed to get device process, logicId: %d, err: %v, devProcessInfo: %v",
			logicID, err, chipInfo)
		return false, err
	}
	logicIDForCompare := fmt.Sprintf("Ascend910-%d", logicID)
	if chipInfo.ProcNum != 0 {
		hwlog.RunLog.Debugf("found busy chip: %v", logicIDForCompare)
		return false, nil
	}
	for _, busyChip := range busyChipList {
		if busyChip == logicIDForCompare {
			hwlog.RunLog.Debugf("found busy chip: %v", logicIDForCompare)
			return false, nil
		}
	}
	return true, nil
}

// canBeReset check if all chips are active
func (hnm *HwAscend910Manager) canBeReset(dev *common.DevFaultInfo) (bool, error) {
	if !hnm.canResetDeviceByLogicID(dev.LogicId) {
		return false, nil
	}
	if common.ParamOption.RealCardType == common.Ascend910A3 &&
		hnm.canA3BeReset(dev) {
		return true, nil
	}
	oriLogicID := dev.LogicId
	podList, err := hnm.client.GetAllPodList()
	if err != nil {
		hwlog.RunLog.Errorf("get pod list fail, err %v", err)
		return false, err
	}
	busyChipList := hnm.getBusyChipListFromPod(podList)
	resetDevNumOnce, err := hnm.hotResetManager.GetResetDevNumOnce()
	if err != nil {
		hwlog.RunLog.Error(err)
		return false, err
	}
	resetStartLogicID := oriLogicID / int32(resetDevNumOnce) * int32(resetDevNumOnce)
	for logicID := resetStartLogicID; logicID < resetStartLogicID+int32(resetDevNumOnce); logicID++ {
		chipActivity, err := hnm.isChipActive(logicID, busyChipList)
		if err != nil {
			return false, err
		}
		if !chipActivity {
			return false, nil
		}
	}
	// all chip on rings are active return true
	return true, nil
}

func (hnm *HwAscend910Manager) canA3BeReset(dev *common.DevFaultInfo) bool {
	cardID, deviceID, err := hnm.GetDmgr().GetCardIDDeviceID(dev.LogicId)
	if err != nil {
		hwlog.RunLog.Errorf("get cardID deviceID by logicID %v faild: %v", dev.LogicId, err)
		return false
	}
	logicIdArr, err := hnm.getAssociatedLogicIDs(dev.LogicId, cardID, deviceID)
	if err != nil {
		return false
	}
	podList, err := hnm.client.GetAllPodList()
	if err != nil {
		hwlog.RunLog.Errorf("get pod list fail, err %v", err)
		return false
	}
	busyChipList := hnm.getBusyChipListFromPod(podList)
	for _, logicID := range logicIdArr {
		chipActivity, err := hnm.isChipActive(logicID, busyChipList)
		if err != nil {
			return false
		}
		if !chipActivity {
			return false
		}
	}
	return true
}

// getBusyChipListFromPod is to get all busy chip from current pod list
func (hnm *HwAscend910Manager) getBusyChipListFromPod(podList *v1.PodList) []string {
	var devList = make([]string, 0)
	for _, pod := range podList.Items {
		if pod.Status.Phase == v1.PodSucceeded {
			continue
		}
		annotationTag := fmt.Sprintf("%s%s", api.ResourceNamePrefix, common.Ascend910)
		annotation, exist := pod.Annotations[annotationTag]
		if !exist {
			continue
		}
		curList := strings.Split(annotation, common.CommaSepDev)
		devList = append(devList, curList...)
	}
	return devList
}

// DoWithVolcanoListAndWatch ascend910 affinity scheduling
func (hnm *HwAscend910Manager) DoWithVolcanoListAndWatch(classifyDevs map[string][]*common.NpuDevice) {
	devStatusSet := hnm.getDevStatesDevSet(classifyDevs)
	if err := hnm.UpdateNodeDeviceInfo(devStatusSet, hnm.updateDeviceInfo); err != nil {
		hwlog.RunLog.Errorf("update device info failed, err: %v", err)
	}
}

func (tool *AscendTools) getDeviceNetworkState(logicID int32, initStatus string) (string, error) {
	healthCode, err := tool.dmgr.GetDeviceNetWorkHealth(logicID)
	if err != nil {
		hwlog.RunLog.Warnf("get logicID %d network health status failed, network health code is %d, "+
			"network health status will not change",
			logicID, healthCode)
		return initStatus, err
	}
	switch healthCode {
	case networkDetectOK, networkDetectInit:
		return v1beta1.Healthy, nil
	default:
		hwlog.RunLog.Debugf("%d network status is unhealthy, health code is %d", logicID, healthCode)
		return v1beta1.Unhealthy, nil
	}
}

func (hnm *HwAscend910Manager) updateDeviceInfo(oldDevInfo, newDevInfo map[string]string,
	devStatusSet common.DevStatusSet) error {
	if newDevInfo == nil {
		return fmt.Errorf("invalid new device info")
	}
	nodeFmtDevRecover, nodeFmtDevNetRecover := sets.String{}, sets.String{}
	newDevRecoverLabel, newAscend910 := hnm.getHealthAndRecoverDev(devStatusSet, nodeFmtDevRecover,
		common.ConvertDevListToSets(oldDevInfo[common.HuaweiUnHealthAscend910], common.CommaSepDev))
	newNetRecoverSets, newNetUHDevSets := hnm.getNewNetworkRecoverDev(devStatusSet.NetUnHealthyDevice,
		common.ConvertDevListToSets(oldDevInfo[common.HuaweiNetworkUnHealthAscend910], common.CommaSepDev),
		nodeFmtDevNetRecover)
	newDevInfo[common.HuaweiAscend910] = newAscend910
	newDevInfo[common.HuaweiRecoveringAscend910] = common.ToString(devStatusSet.RecoveringDevices, common.CommaSepDev)
	// hnm.isNeedBlockAllDevice: server is A800IA2 with hccs and there are fault devices or is already in resetting,
	// no more pod should be scheduled to this node cause all npu resetting is on the way
	// if reset failed more than ResetRetryTimes times, will no longer try to reset server
	if common.ParamOption.HotReset == common.HotResetInfer &&
		hnm.GetResetFailedTimes(common.FirstDevice) <= common.MaxResetTimes &&
		hnm.isNeedBlockAllDevice(devStatusSet.DeviceFault) {

		newDevInfo[common.HuaweiAscend910] = ""
		newDevInfo[common.HuaweiRecoveringAscend910] = common.ToString(devStatusSet.AllDevices, common.CommaSepDev)
		hwlog.RunLog.Warnf("all device on node have been cleared, due to resetting all devices in process")
	}

	newDevInfo[common.HuaweiUnHealthAscend910] = common.ToString(devStatusSet.UnHealthyDevice, common.CommaSepDev)
	newDevInfo[common.HuaweiNetworkUnHealthAscend910] = common.ToString(newNetUHDevSets, common.CommaSepDev)
	var data []byte
	if data = common.MarshalData(devStatusSet.DeviceFault); len(data) == 0 {
		return fmt.Errorf("device fault code marshal failed")
	}
	newDevInfo[common.HuaweiFaultCodeAscend910] = string(data)
	if common.ParamOption.AutoStowingDevs {
		return nil
	}
	curNode, err := hnm.getRecoverLabelFromNodeSets(&nodeFmtDevRecover, &nodeFmtDevNetRecover)
	if err != nil {
		return err
	}
	if err := hnm.update910NodeLabel(curNode, newDevRecoverLabel, hnm.getPatchLabel(newNetRecoverSets)); err != nil {
		hwlog.RunLog.Errorf("update node label failed, err: %v", err)
		return err
	}
	lastTimeNetworkRecoverDevices = newNetRecoverSets
	return nil
}

func (hnm *HwAscend910Manager) isNeedBlockAllDevice(faultDevices []common.DeviceFault) bool {
	usage, err := hnm.GetKubeClient().GetServerUsageLabelCache()
	if err != nil {
		hwlog.RunLog.Errorf("failed to get server usage label, err: %s", err.Error())
		return false
	}
	// only A800IA2 hccs server with fault device will return true
	boardId, err := hnm.GetServerBoardId(common.FirstDevice)
	if err != nil {
		return false
	}
	needBlockErr := false
	for _, device := range faultDevices {
		if device.FaultLevel != common.NotHandleFault {
			needBlockErr = true
		}
	}
	if usage == common.Infer && boardId != common.A800IA2NoneHccsBoardId &&
		boardId != common.A800IA2NoneHccsBoardIdOld &&
		(needBlockErr || hnm.GetIfCardsInResetting(common.FirstDevice)) {
		return true
	}
	return false
}

func (hnm *HwAscend910Manager) update910NodeLabel(curNode *v1.Node, devRecoverLabel, netRecoverLabel string) error {
	newNode := curNode.DeepCopy()
	newNode.Labels[common.HuaweiRecoverAscend910] = devRecoverLabel
	newNode.Labels[common.HuaweiNetworkRecoverAscend910] = netRecoverLabel
	hwlog.RunLog.Debugf("newNode.Labels: %#v", newNode.Labels)
	updatedNode, _, err := hnm.client.PatchNodeState(curNode, newNode)
	if err != nil {
		return err
	}
	hwlog.RunLog.Debugf("updatedNode.Labels: %#v", updatedNode.Labels)
	return nil
}

func (hnm *HwAscend910Manager) getHealthAndRecoverDev(curDevStatusSet common.DevStatusSet, devRecoverDev,
	recordUHDev sets.String) (string, string) {
	device910 := curDevStatusSet.FreeHealthyDevice[common.Ascend910]
	if common.ParamOption.AutoStowingDevs {
		return "", common.ToString(device910, common.CommaSepDev)
	}
	addRecoverSets := recordUHDev.Difference(curDevStatusSet.UnHealthyDevice)
	devRecoverSets := devRecoverDev.Union(addRecoverSets)
	newDevice910 := device910.Difference(devRecoverSets)
	return hnm.getPatchLabel(devRecoverSets), common.ToString(newDevice910, common.CommaSepDev)
}

// getNewNetworkRecoverDev , return new devices to be restored and network unhealthy device in this times
func (hnm *HwAscend910Manager) getNewNetworkRecoverDev(totalNetUHDev, devInfoNetUHRecord,
	labelRecoverRecord sets.String) (sets.String, sets.String) {
	// devInfoNetUHRecord means device info record network unhealthy devices
	// labelRecoverRecord means device's network is ok and to be restored
	// if there is no network unhealthy device and autoStowing devices is true
	if common.ParamOption.AutoStowingDevs {
		return sets.String{}, totalNetUHDev
	}
	// devices recovered between the last check and this check
	recoveredDevSets := lastTimeNetworkRecoverDevices.Difference(labelRecoverRecord)

	newNetworkRecoverDevSets := devInfoNetUHRecord.Difference(totalNetUHDev)
	// remove the device that network is unhealthy in this times
	newNetworkRecoverDevSets = newNetworkRecoverDevSets.Difference(labelRecoverRecord.Intersection(totalNetUHDev))
	// remove the device that recovered
	newNetworkRecoverDevSets = newNetworkRecoverDevSets.Difference(recoveredDevSets)
	newNetworkUnhealthyDevSets := devInfoNetUHRecord.Union(totalNetUHDev).Difference(recoveredDevSets)
	return newNetworkRecoverDevSets, newNetworkUnhealthyDevSets
}

// getPatchLabel get elements one by one from the sets and change the element "Ascend910-x" to "x"
// which will patch to node
func (hnm *HwAscend910Manager) getPatchLabel(chips sets.String) string {
	if chips.Len() == 0 {
		return ""
	}

	var ascendLabel = make([]string, 0)
	for devName := range chips {
		devTypeAndID := strings.Split(devName, common.MiddelLine)
		if len(devTypeAndID) != common.LabelDeviceLen {
			continue
		}
		phyID := devTypeAndID[len(devTypeAndID)-1]
		if _, isValidNum := common.IsValidNumber(phyID); !isValidNum {
			continue
		}
		ascendLabel = append(ascendLabel, phyID)
	}

	return strings.Join(ascendLabel, common.DotSepDev)
}

func (hnm *HwAscend910Manager) getRecoverLabelFromNodeSets(devRecoverLabel, netRecoverLabel *sets.String) (
	*v1.Node, error) {
	curNode, err := hnm.client.GetNode()
	if err != nil {
		hwlog.RunLog.Error("get node error")
		return nil, err
	}
	if curNode == nil || curNode.Labels == nil {
		return nil, fmt.Errorf("invalid node")
	}
	// devRecoverLabel like Ascend910-0,Ascend910-2,Ascend910-3, means dev healthy exception
	*devRecoverLabel = hnm.toStandardDeviceFmt(common.ConvertDevListToSets(
		curNode.Labels[common.HuaweiRecoverAscend910], common.DotSepDev))
	// netRecoverLabel like Ascend910-0,Ascend910-2,Ascend910-3, means dev network exception
	*netRecoverLabel = hnm.toStandardDeviceFmt(common.ConvertDevListToSets(
		curNode.Labels[common.HuaweiNetworkRecoverAscend910], common.DotSepDev))
	return curNode, nil
}

// toStandardDeviceFmt convert physical id "x" to format "Ascend910-x"
func (hnm *HwAscend910Manager) toStandardDeviceFmt(devices sets.String) sets.String {
	if devices.Len() == 0 {
		return sets.String{}
	}

	standardSets := sets.String{}
	for devID := range devices {
		deviceName := fmt.Sprintf("%s-%s", common.Ascend910, devID)
		standardSets.Insert(deviceName)
	}

	return standardSets
}

func (hnm *HwAscend910Manager) updateHotResetCache(classifyDevs map[string][]*common.NpuDevice) error {
	deviceList, ok := classifyDevs[common.Ascend910]
	if !ok {
		hwlog.RunLog.Error("ascend 910 device list no found")
		return fmt.Errorf("ascend 910 device list not found")
	}
	if err := hnm.updateUpgradeErrorInfo(classifyDevs); err != nil {
		hwlog.RunLog.Errorf("fail to update upgrade error npu info, err: %v", err)
	}
	if err := hnm.hotResetManager.UpdateGlobalDevFaultInfoCache(deviceList, isolateDevList); err != nil {
		hwlog.RunLog.Errorf("failed to update global device fault info cache, err: %v", err)
		return err
	}
	if err := hnm.setTaskDevInfoCache(); err != nil {
		hwlog.RunLog.Errorf("failed to set task device info cache, err: %v", err)
		return err
	}
	return nil
}

// updateUpgradeErrorInfo updates global variable isolateDevList
func (hnm *HwAscend910Manager) updateUpgradeErrorInfo(classifyDevs map[string][]*common.NpuDevice) error {
	if len(isolateDevList) == 0 {
		return nil
	}
	deviceList, ok := classifyDevs[common.Ascend910]
	if !ok {
		return fmt.Errorf("no Ascend 910 device found in cache")
	}
	for _, dev := range deviceList {
		index := -1
		for i, _ := range isolateDevList {
			if isolateDevList[i] != dev.LogicID {
				continue
			}
			if dev.Health == v1beta1.Unhealthy {
				continue
			}
			index = i
			break
		}
		if index != -1 {
			isolateDevList = append(isolateDevList[:index], isolateDevList[index+1:]...)
		}
	}
	return nil
}

func (hnm *HwAscend910Manager) setTaskDevInfoCache() error {
	podList := hnm.client.GetActivePodListCache()
	newTaskDevListCache := make(map[string][]int32)
	newTaskDevFaultInfoCache := make(map[string][]*common.TaskDevInfo)
	newTaskPodCache := make(map[string]v1.Pod)
	taskListUsedDevice := make(map[string]struct{})
	for _, pod := range podList {
		tmpNpu, ok := pod.Annotations[common.HuaweiAscend910]
		if !ok || len(tmpNpu) == 0 || len(tmpNpu) > common.PodAnnotationMaxLength {
			continue
		}
		devIdList, err := hnm.convertPhysicIdToLogicId(hnm.hotResetManager.GetDevIdList(tmpNpu))
		if err != nil {
			hwlog.RunLog.Errorf("failed to convert physic id to logic id, npu: %s, err: %v", tmpNpu, err)
			continue
		}
		if hnm.isReSchedulingScene(len(devIdList)) {
			continue
		}
		taskName := hnm.hotResetManager.GetTaskNameByPod(pod)
		if taskName == "" {
			continue
		}
		rankIndex, ok := pod.Annotations[api.PodRankIndexAnno]
		if common.ParamOption.RealCardType == common.Ascend910B && hnm.GetDeviceUsage() == common.Infer {
			rankIndex = common.InferRankIndex
		} else {
			if !ok {
				hwlog.RunLog.Warn("failed to get rank index by rank index key")
				continue
			}
		}
		taskListUsedDevice[taskName] = struct{}{}
		newTaskDevListCache[taskName] = devIdList
		taskDevFaultInfoList, err := hnm.hotResetManager.GenerateTaskDevFaultInfoList(devIdList, rankIndex)
		if err != nil {
			hwlog.RunLog.Errorf("failed to get task device fault info list, err: %v", err)
			return err
		}
		// podAntiAffinity make sure that there won't be multi pod in single node of one task
		newTaskDevFaultInfoCache[taskName] = taskDevFaultInfoList
		newTaskPodCache[taskName] = pod
		if err = hnm.hotResetManager.UpdateFaultDev2PodMap(devIdList, pod); err != nil {
			hwlog.RunLog.Errorf("update faultDev2PodMap error: %v", err)
		}
	}
	return hnm.handleUpdateCaches(taskListUsedDevice, newTaskDevListCache, newTaskDevFaultInfoCache, newTaskPodCache)
}

func (hnm *HwAscend910Manager) handleUpdateCaches(taskListUsedDevice map[string]struct{},
	newTaskDevListCache map[string][]int32, newTaskDevFaultInfoCache map[string][]*common.TaskDevInfo,
	newTaskPodCache map[string]v1.Pod) error {
	hnm.hotResetManager.UpdateFreeTask(taskListUsedDevice, newTaskDevListCache)
	if err := hnm.hotResetManager.UpdateTaskDevListCache(newTaskDevListCache); err != nil {
		return err
	}
	if err := hnm.hotResetManager.UpdateTaskDevFaultInfoCache(newTaskDevFaultInfoCache); err != nil {
		return err
	}
	if err := hnm.hotResetManager.UpdateTaskPodCache(newTaskPodCache); err != nil {
		return err
	}
	return nil
}

func (hnm *HwAscend910Manager) convertPhysicIdToLogicId(physicIds []int32) ([]int32, error) {
	if len(physicIds) == 0 {
		return nil, fmt.Errorf("convert physic id to logic id failed, " +
			"physic id is nil or length of physic id is 0")
	}
	var logicIds []int32
	for _, physicId := range physicIds {
		logicId, err := hnm.GetDmgr().GetLogicIDFromPhysicID(physicId)
		if err != nil {
			hwlog.RunLog.Errorf("convert physic id to logic id failed, err: %v", err)
			return nil, err
		}
		logicIds = append(logicIds, logicId)
	}
	return logicIds, nil
}

func (hnm *HwAscend910Manager) convertLogicIdToPhysicId(logicIds []int32) ([]int32, error) {
	if len(logicIds) == 0 {
		return nil, fmt.Errorf("convert logic id to physic id failed, logic id empty")
	}
	var physicIds []int32
	for _, logicId := range logicIds {
		physicId, err := hnm.GetDmgr().GetPhysicIDFromLogicID(logicId)
		if err != nil {
			hwlog.RunLog.Errorf("convert logic id to physic id failed, err: %v", err)
			return nil, err
		}
		physicIds = append(physicIds, physicId)
	}
	return physicIds, nil
}

func (hnm *HwAscend910Manager) isReSchedulingScene(npuCount int) bool {
	resetDevNumOnce, err := hnm.hotResetManager.GetResetDevNumOnce()
	if err != nil {
		hwlog.RunLog.Error(err)
		return false
	}

	if hnm.GetDeviceUsage() == common.Train && npuCount < resetDevNumOnce {
		return true
	}

	return false
}

func (hnm *HwAscend910Manager) isTaskInReset(taskName string) (bool, error) {
	pod, err := hnm.hotResetManager.GetTaskPod(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task pod, err: %v", err)
		return false, err
	}
	if hnm.hotResetManager.IsCurNodeTaskInReset(taskName) {
		hwlog.RunLog.Infof("this node task %s is resetting, skip once process", taskName)
		return true, nil
	}
	resetCM, err := hnm.hotResetManager.GetCMFromCache(pod.Namespace + "/" + common.ResetInfoCMNamePrefix + taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get reset info cm, err: %v", err)
		return false, err
	}
	resetInfoData, err := getResetInfoData(resetCM)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get reset info data, err: %v", err)
		return false, err
	}
	if len(resetInfoData) == 0 {
		return false, nil
	}
	needTolorance := false
	for _, info := range resetInfoData {
		if info.Status != common.UnrecoveredStatus {
			needTolorance = true
			break
		}
	}
	if needTolorance {
		hwlog.RunLog.Debugf("task %s not in reset", taskName)
		return false, nil
	}
	hwlog.RunLog.Infof("global task %s is resetting, skip once process", taskName)
	return true, nil
}

// filterDevStatus filters the health of the device being reset and
// the network health of the ring that the device is on
func (hnm *HwAscend910Manager) filterDevStatus(classifyDevs map[string][]*common.NpuDevice) error {
	devStatusList, ok := classifyDevs[common.Ascend910]
	if !ok {
		return fmt.Errorf("no ascend 910 device needed filter")
	}
	if common.ParamOption.RealCardType == common.Ascend910A3 &&
		hnm.filterDevStatusForA3(devStatusList) == nil {
		return nil
	}
	devInReset := hnm.hotResetManager.GetDevListInReset()
	filteredRingIndex := -1
	resetDevNumOnce, err := hnm.hotResetManager.GetResetDevNumOnce()
	if err != nil {
		hwlog.RunLog.Error(err)
		return err
	}
	for _, devStatus := range devStatusList {
		if _, ok := devInReset[devStatus.LogicID]; !ok || devStatus.Health == v1beta1.Healthy ||
			hnm.isDevShouldBeIsolate(devStatus.LogicID) {
			continue
		}
		devStatus.Health = v1beta1.Healthy
		ringIndex := int(devStatus.LogicID) / resetDevNumOnce
		if ringIndex != filteredRingIndex {
			startDevIndex := ringIndex * resetDevNumOnce
			endDevIndex := startDevIndex + resetDevNumOnce
			for devIndex := startDevIndex; devIndex < endDevIndex; devIndex++ {
				devStatusList[devIndex].NetworkHealth = v1beta1.Healthy
			}
			filteredRingIndex = ringIndex
		}
	}
	return nil
}

func (hnm *HwAscend910Manager) filterDevStatusForA3(devStatusList []*common.NpuDevice) error {
	devToBeSet := make(map[int32]struct{})
	devInReset := hnm.hotResetManager.GetDevListInReset()
	for _, dev := range devStatusList {
		if _, ok := devInReset[dev.LogicID]; !ok || dev.Health == v1beta1.Healthy ||
			hnm.isDevShouldBeIsolate(dev.LogicID) {
			continue
		}
		dev.Health = v1beta1.Healthy
		if _, exist := devToBeSet[dev.LogicID]; exist {
			continue
		}
		logicIdArr, err := hnm.getAssociatedLogicIDs(dev.LogicID, dev.CardID, dev.DeviceID)
		if err != nil {
			return err
		}
		for _, id := range logicIdArr {
			devToBeSet[id] = struct{}{}
		}
	}
	for idx, _ := range devToBeSet {
		if int(idx) >= len(devStatusList) {
			hwlog.RunLog.Errorf("device logicID %v is greater than device list length %v",
				idx, len(devStatusList))
			continue
		}
		devStatusList[idx].NetworkHealth = v1beta1.Healthy
	}
	return nil
}

func (hnm *HwAscend910Manager) processAllTask(classifyDevs map[string][]*common.NpuDevice) error {
	taskDevFaultInfoList := hnm.hotResetManager.GetAllTaskDevFaultInfoList()
	for taskName := range taskDevFaultInfoList {
		policy, policyLevel, err := hnm.hotResetManager.GetTaskProcessPolicy(taskName)
		if err != nil {
			hwlog.RunLog.Errorf("failed to get task %s process policy, err: %v", taskName, err)
			continue
		}
		if policyLevelHandle(policy, taskName, policyLevel) {
			continue
		}
		if hnm.isolateSceneHandle(taskName) {
			continue
		}
		resetInfo, err := hnm.preProcess(taskName, policy)
		if err != nil {
			return err
		}
		if err = hnm.runProcessTask(taskName, policyLevel, resetInfo, classifyDevs); err != nil {
			return err
		}
	}
	return nil
}

func policyLevelHandle(handlePolicy, handleTaskName string, handlePolicyLevel int) bool {
	switch handlePolicyLevel {
	case common.RestartRequestErrorLevel, common.RestartErrorLevel:
		hwlog.RunLog.Debugf("start handling fault: %s - %d, task name: %s", handlePolicy,
			handlePolicyLevel, handleTaskName)
	case common.ResetErrorLevel:
		hwlog.RunLog.Debugf("start handling fault: %s - %d, task name: %s", handlePolicy,
			handlePolicyLevel, handleTaskName)
	case common.FreeResetErrorLevel:
		return true
	default:
		return true
	}
	return false
}

func (hnm *HwAscend910Manager) isolateSceneHandle(handleTaskName string) bool {
	if resetFlag, err := hnm.isTaskInReset(handleTaskName); err != nil || resetFlag {
		if resetFlag && !hnm.hotResetManager.IsCurNodeTaskInReset(handleTaskName) &&
			hnm.hotResetManager.IsExistFaultyDevInTask(handleTaskName) {
			hwlog.RunLog.Infof("task %s is in reset process which is not performed on current node. NPU"+
				" faults occurred on current node, the task's process policy will be marked as isolate", handleTaskName)
			go hnm.tryWriteIsolationInfo(handleTaskName)
		}
		return true
	}
	return false
}

func (hnm *HwAscend910Manager) runProcessTask(taskName string, policyLevel int, resetInfo *common.TaskResetInfo,
	classifyDevs map[string][]*common.NpuDevice) error {
	switch policyLevel {
	case common.RestartRequestErrorLevel:
		go hnm.restartRequestProcess(taskName, resetInfo, classifyDevs)
	case common.RestartErrorLevel:
		go hnm.restartProcess(taskName, resetInfo, classifyDevs)
	case common.ResetErrorLevel:
		go hnm.resetProcess(taskName, resetInfo, classifyDevs)
	default:
		return fmt.Errorf("invalid processing policy")
	}
	return nil
}

func (hnm *HwAscend910Manager) restartRequestProcess(taskName string, resetInfo *common.TaskResetInfo,
	classifyDevs map[string][]*common.NpuDevice) {
	defer func() {
		if err := hnm.postProcess(taskName, resetInfo); err != nil {
			hwlog.RunLog.Errorf("failed to unset device in reset, err %v", err)
		}
	}()
	devFaultInfoList, err := hnm.hotResetManager.GetTaskDevFaultInfoList(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("get task fault devices info list failed, err: %v", err)
		return
	}
	hwlog.RunLog.Infof("start handle L2 fault, task name: %s", taskName)
	common.RecordFaultInfoList(devFaultInfoList)
	devFaultInfoListInReset := hnm.hotResetManager.DeepCopyDevFaultInfoList(devFaultInfoList)

	currentPolicy, needUpgrade, resetErr := hnm.checkDevErrorCode(taskName, devFaultInfoList, classifyDevs)
	if resetErr == nil {
		hnm.handleSucceedRestartRequest(taskName, currentPolicy, devFaultInfoList, devFaultInfoListInReset)
		return
	}
	hwlog.RunLog.Errorf("failed to refresh device fault info, err %v", resetErr)
	if !needUpgrade {
		return
	}
	currentPolicy, resetErr = hnm.upgradeRestartRequestProcess(taskName, devFaultInfoList, classifyDevs)
	if resetErr != nil {
		hwlog.RunLog.Errorf("failed to exec upgrade reset process, err: %v", resetErr)
		if err := hnm.updateResetCMStatus(taskName, common.IsolateError, common.RestartRequestError,
			common.RecoverFailedStatus, devFaultInfoList); err != nil {
			hwlog.RunLog.Errorf(failedToUpdateCmPattern, err)
		}
		return
	}
	hnm.handleSucceedRestartRequest(taskName, currentPolicy, devFaultInfoList, devFaultInfoListInReset)
}

func (hnm *HwAscend910Manager) handleSucceedRestartRequest(taskName, currentPolicy string,
	newDevInfo, oldDevInfo []*common.TaskDevInfo) {
	for _, devInfo := range newDevInfo {
		common.SetDeviceInit(devInfo.LogicId)
	}
	if err := hnm.updateResetCMStatus(taskName, currentPolicy, common.RestartRequestError, common.RecoveredStatus,
		oldDevInfo); err != nil {
		hwlog.RunLog.Errorf(updateResetCMFailedPattern, err)
		return
	}
	if err := hnm.hotResetManager.UnSetTaskInReset(taskName); err != nil {
		hwlog.RunLog.Errorf(unsetTaskFailedPattern, err)
	}
}

func (hnm *HwAscend910Manager) checkDevErrorCode(taskName string, devFaultInfo []*common.TaskDevInfo,
	classifyDevs map[string][]*common.NpuDevice) (string,
	bool, error) {
	timeOut := time.After(common.WaitErrorCodeCleanTime * time.Second)
	timeCost := 0
	for {
		select {
		case <-timeOut:
			return "", true, fmt.Errorf("after %d second, there still has error code on device", common.WaitErrorCodeCleanTime)
		default:
			if err := hnm.refreshDevFaultInfo(devFaultInfo, classifyDevs); err != nil {
				return "", false, err
			}
			faultInfoList, err := hnm.hotResetManager.GetDevListByPolicyLevel(devFaultInfo,
				common.RestartRequestErrorLevel)
			if err != nil {
				hwlog.RunLog.Errorf("failed to get need fault device list, err %v", err)
				return "", false, err
			}
			if len(faultInfoList) == 0 {
				hwlog.RunLog.Infof("L2 fault self-healing success, task name: %s, cost: %d second", taskName, timeCost)
				return common.RestartRequestError, false, nil
			}
			time.Sleep(time.Second)
			timeCost++
		}
	}
}

func (hnm *HwAscend910Manager) restartProcess(taskName string, resetInfo *common.TaskResetInfo,
	classifyDevs map[string][]*common.NpuDevice) {
	defer func() {
		if err := hnm.postProcess(taskName, resetInfo); err != nil {
			hwlog.RunLog.Errorf("failed to unset device in reset, err %v", err)
		}
	}()
	devFaultInfoList, err := hnm.hotResetManager.GetTaskDevFaultInfoList(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task device fault info list, err %v", err)
		if err := hnm.updateResetCMStatus(taskName, common.IsolateError, common.RestartError, common.RecoverFailedStatus,
			devFaultInfoList); err != nil {
			hwlog.RunLog.Errorf(failedToUpdateCmPattern, err)
		}
		return
	}
	hwlog.RunLog.Infof("start handle L3 fault, task name: %s", taskName)
	common.RecordFaultInfoList(devFaultInfoList)
	devFaultInfoListInReset := hnm.hotResetManager.DeepCopyDevFaultInfoList(devFaultInfoList)
	if err := hnm.waitForAllFaultyDeviceProcessesToZero(taskName, devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf("failed to check the number of processes in the fault device, err: %v", err)
		return
	}
	time.Sleep(common.WaitFaultSelfHealingTime * time.Second)
	if err := hnm.refreshDevFaultInfo(devFaultInfoList, classifyDevs); err != nil {
		hwlog.RunLog.Errorf("failed to refresh device fault info, err %v", err)
		return
	}
	currentPolicy, err := hnm.upgradeRestartProcess(taskName, devFaultInfoList, classifyDevs)
	if err != nil {
		hwlog.RunLog.Errorf("failed to exec upgrade restart process, err: %v", err)
		if err := hnm.updateResetCMStatus(taskName, common.IsolateError, common.RestartError, common.RecoverFailedStatus,
			devFaultInfoList); err != nil {
			hwlog.RunLog.Errorf(failedToUpdateCmPattern, err)
		}
		return
	}
	if err := hnm.updateResetCMStatus(taskName, currentPolicy, common.RestartError, common.RecoveredStatus,
		devFaultInfoListInReset); err != nil {
		hwlog.RunLog.Errorf(updateResetCMFailedPattern, err)
		return
	}
	if err := hnm.hotResetManager.UnSetTaskInReset(taskName); err != nil {
		hwlog.RunLog.Errorf(unsetTaskFailedPattern, err)
		return
	}
	return
}

// upgradeRestartProcess upgrade the device restart processing to the device reset processing
func (hnm *HwAscend910Manager) upgradeRestartProcess(taskName string, devFaultInfoList []*common.TaskDevInfo,
	classifyDevs map[string][]*common.NpuDevice) (string,
	error) {
	restartFaultInfoList, err := hnm.hotResetManager.GetDevListByPolicyLevel(devFaultInfoList, common.RestartErrorLevel)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get need reset device list, err %v", err)
		return "", err
	}
	if len(restartFaultInfoList) == 0 {
		hwlog.RunLog.Infof("after restart, L3 fault healing success, task name: %s", taskName)
		return common.RestartError, nil
	}
	hwlog.RunLog.Errorf("after restart, L3 fault healing failed, upgrade fault, task name: %s", taskName)
	if err := hnm.updateResetCMStatusWithoutWait(taskName, common.ResetError,
		common.RestartError, common.UnrecoveredStatus, devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf(failedToUpdateCmPattern, err)
		return "", err
	}
	if err := hnm.resetDeviceOnce(devFaultInfoList, classifyDevs); err != nil {
		return "", err
	}
	resultFaultInfoList, err := hnm.hotResetManager.GetDevListByPolicyLevel(devFaultInfoList, common.RestartErrorLevel)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get need reset device list, err: %v", err)
		return "", err
	}
	if len(resultFaultInfoList) == 0 {
		hwlog.RunLog.Infof("after reset, L3 fault healing success, task name: %s", taskName)
		return common.ResetError, nil
	}
	if err := hnm.updateResetCMStatus(taskName, common.IsolateError, common.RestartError, common.RecoverFailedStatus,
		devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf(failedToUpdateCmPattern, err)
		return "", err
	}
	return "", fmt.Errorf("failed to restart task, upgrade recovery failed status")
}

// upgradeRestartProcess upgrade the device restart processing to the device reset processing
func (hnm *HwAscend910Manager) upgradeRestartRequestProcess(taskName string,
	devFaultInfoList []*common.TaskDevInfo, classifyDevs map[string][]*common.NpuDevice) (string, error) {
	hwlog.RunLog.Warnf("L2 fault self-healing failed, upgrade fault, task name: %s", taskName)
	if err := hnm.updateResetCMStatusWithoutWait(taskName, common.ResetError, common.RestartRequestError,
		common.UnrecoveredStatus, devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf("failed to update reset cm to ResetError, err: %v", err)
		return "", err
	}
	if err := hnm.waitForAllFaultyDeviceProcessesToZero(taskName, devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf("failed to wait for all faulty devices having no process, err: %v", err)
		return "", err
	}
	if err := hnm.resetDeviceOnce(devFaultInfoList, classifyDevs); err != nil {
		hwlog.RunLog.Warnf("L2 upgrade reset failed, err:%v", err)
		return "", err
	}
	resultFaultInfoList, err := hnm.hotResetManager.GetDevListByPolicyLevel(devFaultInfoList,
		common.RestartRequestErrorLevel)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get need fault device list, err: %v", err)
		return "", err
	}
	if len(resultFaultInfoList) == 0 {
		hwlog.RunLog.Infof("after reset, L2 fault healing success, task name: %s", taskName)
		return common.ResetError, nil
	}
	if err := hnm.updateResetCMStatus(taskName, common.IsolateError, common.RestartRequestError,
		common.RecoverFailedStatus, devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf(failedToUpdateCmPattern, err)
		return "", err
	}
	return "", fmt.Errorf("after reset, L2 fault still exists, task name: %s", taskName)
}

func (hnm *HwAscend910Manager) updateResetCMStatus(taskName, policy, initPolicy, status string,
	devFaultInfoList []*common.TaskDevInfo) error {
	taskInReset, err := hnm.isTaskInReset(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to find in reset task %s, err: %v", taskName, err)
		return err
	}
	if !taskInReset && (status == common.RecoveredStatus || status == common.RecoverFailedStatus) {
		return fmt.Errorf("no need to update reset config map with failed or recovered status, " +
			"because there is no task in reset")
	}

	if status == common.RecoveredStatus && policy == common.ResetError {
		hwlog.RunLog.Infof("reset success, wait %d second for npu ready", common.WaitNpuReadyTime)
		time.Sleep(common.WaitNpuReadyTime * time.Second)
	}

	newResetInfo, err := hnm.hotResetManager.GetTaskResetInfo(devFaultInfoList, policy, initPolicy, status)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task reset info list, err: %v", err)
		return err
	}
	pod, err := hnm.hotResetManager.GetTaskPod(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task pod, err: %v", err)
		return err
	}
	if _, err := hnm.client.WriteResetInfoDataIntoCM(taskName, pod.Namespace, newResetInfo, false); err != nil {
		hwlog.RunLog.Errorf("write reset info into reset cm failed, err: %v", err)
		return err
	}
	hwlog.RunLog.Infof("sleep %d second for config map to sync", common.WaitProcessReadCMTime)
	time.Sleep(common.WaitProcessReadCMTime * time.Second)
	return nil
}

func (hnm *HwAscend910Manager) updateResetCMStatusWithoutWait(taskName, policy, initPolicy, status string,
	devFaultInfoList []*common.TaskDevInfo) error {
	newResetInfo, err := hnm.hotResetManager.GetTaskResetInfo(devFaultInfoList, policy, initPolicy, status)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task reset info, err: %v", err)
		return err
	}
	pod, err := hnm.hotResetManager.GetTaskPod(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task pod, err: %v", err)
		return err
	}
	if _, err := hnm.client.WriteResetInfoDataIntoCM(taskName, pod.Namespace, newResetInfo, false); err != nil {
		hwlog.RunLog.Errorf("failed to write reset info to config map, err: %v", err)
		return err
	}
	return nil
}

func (hnm *HwAscend910Manager) resetProcess(taskName string, resetInfo *common.TaskResetInfo,
	classifyDevs map[string][]*common.NpuDevice) {
	devFaultInfoList, err := hnm.hotResetManager.GetTaskDevFaultInfoList(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task device fault info list, err: %v", err)
		return
	}
	defer func() {
		removeResetDevMap(devFaultInfoList)
		if err := hnm.postProcess(taskName, resetInfo); err != nil {
			hwlog.RunLog.Errorf("failed to exec post process, err: %v", err)
		}
	}()
	hwlog.RunLog.Infof("start handle L5 fault, task name: %s", taskName)
	common.RecordFaultInfoList(devFaultInfoList)
	devFaultInfoListInReset := hnm.hotResetManager.DeepCopyDevFaultInfoList(devFaultInfoList)
	if err := hnm.waitForAllFaultyDeviceProcessesToZero(taskName, devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf("failed to check the number of processes in the fault device, err: %v", err)
		if err := hnm.updateResetCMStatus(taskName, common.IsolateError, common.ResetError, common.RecoverFailedStatus,
			devFaultInfoList); err != nil {
			hwlog.RunLog.Errorf(failedToUpdateCmPattern, err)
		}
		return
	}
	if err := hnm.resetDeviceOnce(devFaultInfoList, classifyDevs); err != nil {
		hwlog.RunLog.Errorf("failed to reset device, err: %v", err)
		if err := hnm.updateResetCMStatus(taskName, common.IsolateError, common.ResetError, common.RecoverFailedStatus,
			devFaultInfoList); err != nil {
			hwlog.RunLog.Errorf(failedToUpdateCmPattern, err)
		}
		return
	}
	if err := hnm.upgradeResetProcess(taskName, devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf("failed to exec upgrade reset process, err :%v", err)
		if err := hnm.updateResetCMStatus(taskName, common.IsolateError, common.ResetError, common.RecoverFailedStatus,
			devFaultInfoList); err != nil {
			hwlog.RunLog.Errorf(failedToUpdateCmPattern, err)
		}
		return
	}
	if err := hnm.updateResetCMStatus(taskName, common.ResetError, common.ResetError, common.RecoveredStatus,
		devFaultInfoListInReset); err != nil {
		hwlog.RunLog.Errorf(updateResetCMFailedPattern, err)
		return
	}
	if err := hnm.hotResetManager.UnSetTaskInReset(taskName); err != nil {
		hwlog.RunLog.Errorf(unsetTaskFailedPattern, err)
		return
	}
}

// waitForAllFaultyDeviceProcessesToZero waits for the number of processes on all devices
// that need to be hot reset under a task to be 0.
func (hnm *HwAscend910Manager) waitForAllFaultyDeviceProcessesToZero(taskName string,
	devFaultInfoList []*common.TaskDevInfo) error {
	faultDeviceLogicIdMap, err := hnm.getNeedResetDeviceLogicIdMap(devFaultInfoList)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get need reset device logic id list, err: %v", err)
		return err
	}
	hwlog.RunLog.Infof("start check the number of remaining processes on the faulty chips in the task named %v,"+
		"logic id list: %v", taskName, faultDeviceLogicIdMap)
	timeoutChan := time.After(common.WaitProcessesToZeroTime * time.Second)
	timeCount := 0
	for {
		select {
		case _, ok := <-timeoutChan:
			if !ok {
				return fmt.Errorf("after %d second, there is still remaining processes on the faulty chips",
					int(common.WaitProcessesToZeroTime))
			}
			if hnm.canContinueGraceProcess(faultDeviceLogicIdMap, taskName, true) {
				return nil
			}
			hwlog.RunLog.Warnf("more than %v seconds have elapsed, "+
				"but the number of remaining processes on the faulty chips in the task named %v is still not 0, "+
				"fault device logicId and current number of remaining processes is %+v",
				int(common.WaitProcessesToZeroTime), taskName, faultDeviceLogicIdMap)
			if err = hnm.updateResetCMStatusToIsolate(taskName, devFaultInfoList); err != nil {
				hwlog.RunLog.Errorf("failed to update status of reset configmap to isolate, taskName: %v, err: %v", taskName, err)
				return err
			}
			return fmt.Errorf("check the number of remaining processes on the faulty chips timeout")
		default:
			if hnm.canContinueGraceProcess(faultDeviceLogicIdMap, taskName, false) {
				hwlog.RunLog.Infof("cost %d second to delete all process", timeCount)
				return nil
			}
			time.Sleep(common.PollingInterval * time.Second)
			timeCount++
		}
	}
}

func (hnm *HwAscend910Manager) canContinueGraceProcess(faultDeviceLogicIdMap map[int32]int32,
	taskName string, isLastQuery bool) bool {
	isNumberOfAllProcessZero, err := hnm.checkNumberOfAllProcessIsZero(faultDeviceLogicIdMap)
	if err != nil {
		hwlog.RunLog.Errorf("failed to check whether the number of processes on the chip is zero, err is: %v", err)
		if isLastQuery {
			hwlog.RunLog.Warn("an error is reported when the DCMI interface is queried. " +
				"Therefore, continue to perform graceful fault tolerance")
			return true
		}
		return false
	}
	if isNumberOfAllProcessZero {
		hwlog.RunLog.Infof("the number of processes on all chips that require hot reset is 0 now, "+
			"the task name is %v", taskName)
		return true
	}
	return false
}

func (hnm *HwAscend910Manager) updateResetCMStatusToIsolate(taskName string,
	devFaultInfoList []*common.TaskDevInfo) error {
	initPolicy, _, err := hnm.hotResetManager.GetTaskProcessPolicy(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task %s policy, err: %v", taskName, err)
		return err
	}
	if err := hnm.updateResetCMStatus(taskName, common.IsolateError, initPolicy, common.RecoverFailedStatus,
		devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf("failed to update reset cm to isolate, err: %v", err)
		return fmt.Errorf(updateResetCMFailedPattern, err)
	}
	hwlog.RunLog.Infof("successfully updated status of reset configmap to isolate")
	return nil
}

func (hnm *HwAscend910Manager) getA3LogicMapByAssociation(devFaultInfoList []*common.TaskDevInfo) (map[int32]int32,
	error) {
	if common.ParamOption.RealCardType != common.Ascend910A3 {
		return nil, fmt.Errorf("only support A3")
	}
	ret := make(map[int32]int32)
	idArr := make([]int32, 0, len(devFaultInfoList))
	for _, devFault := range devFaultInfoList {
		cardID, deviceID, err := hnm.GetDmgr().GetCardIDDeviceID(devFault.LogicId)
		if err != nil {
			hwlog.RunLog.Errorf("get cardID deviceID by logicID %v failed: %v", devFault.LogicId, err)
			return nil, err
		}
		logicIDs, err := hnm.getAssociatedLogicIDs(devFault.LogicId, cardID, deviceID)
		if err != nil {
			return nil, err
		}
		idArr = append(idArr, logicIDs...)
	}
	for _, id := range idArr {
		if _, exist := ret[id]; !exist {
			ret[id] = id
		}
	}
	return ret, nil
}

// getNeedResetDeviceLogicIdList gets the list of logic ids of the devices that need to be reset.
func (hnm *HwAscend910Manager) getNeedResetDeviceLogicIdMap(devFaultInfoList []*common.TaskDevInfo) (
	map[int32]int32, error) {
	if devMap, err := hnm.getA3LogicMapByAssociation(devFaultInfoList); err == nil {
		return devMap, nil
	}
	resetFaultInfoMap, err := hnm.hotResetManager.GetNeedResetDevMap(devFaultInfoList)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get need reset device list, err: %v", err)
		return nil, err
	}
	faultDeviceLogicIdMap := make(map[int32]int32, len(resetFaultInfoMap))
	resetDevNumOnce, err := hnm.hotResetManager.GetResetDevNumOnce()
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, err
	}
	for resetStartLogicId := range resetFaultInfoMap {
		if err = hnm.addAllLogicIdsToFaultMap(resetStartLogicId, int32(resetDevNumOnce), faultDeviceLogicIdMap); err != nil {
			hwlog.RunLog.Errorf("failed to add logic_ids to faultDeviceLogicIdMap in the same SMP system, err: %v", err)
			return nil, err
		}
	}
	return faultDeviceLogicIdMap, nil
}

func (hnm *HwAscend910Manager) addAllLogicIdsToFaultMap(resetStartLogicId, chipCountOnRing int32,
	faultDeviceLogicIdMap map[int32]int32) error {
	if faultDeviceLogicIdMap == nil {
		faultDeviceLogicIdMap = make(map[int32]int32)
	}
	for logicID := resetStartLogicId; logicID < resetStartLogicId+chipCountOnRing; logicID++ {
		faultDeviceLogicIdMap[logicID] = common.InitialProcNum
	}
	return nil
}

// checkNumberOfAllProcessIsZero checks if the number of processes is 0 on all faulty devices.
func (hnm *HwAscend910Manager) checkNumberOfAllProcessIsZero(faultDeviceLogicIdMap map[int32]int32) (bool, error) {
	if len(faultDeviceLogicIdMap) == 0 {
		return true, nil
	}
	isNumberOfAllProcessZero := true
	for faultDeviceLogicId, num := range faultDeviceLogicIdMap {
		if num == 0 {
			continue
		}
		devProcessInfo, err := hnm.dmgr.GetDevProcessInfo(faultDeviceLogicId)
		if err != nil || devProcessInfo == nil {
			hwlog.RunLog.Errorf("failed to get device process info, logicId: %v, err: %v, devProcessInfo: %v",
				faultDeviceLogicId, err, devProcessInfo)
			return false, err
		}
		hwlog.RunLog.Debugf("current proc number of logic %v is %v", faultDeviceLogicId, devProcessInfo.ProcNum)
		faultDeviceLogicIdMap[faultDeviceLogicId] = devProcessInfo.ProcNum
		isNumberOfAllProcessZero = devProcessInfo.ProcNum == 0
		if !isNumberOfAllProcessZero {
			return false, nil
		}
	}
	return isNumberOfAllProcessZero, nil
}

// upgradeResetProcess upgrade the device reset processing to the device isolation processing
func (hnm *HwAscend910Manager) upgradeResetProcess(taskName string, devFaultInfoList []*common.TaskDevInfo) error {
	resetFaultInfoMap, err := hnm.hotResetManager.GetNeedResetDevMap(devFaultInfoList)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get need reset device list, err: %v", err)
		return err
	}
	if len(resetFaultInfoMap) == 0 {
		return nil
	}
	if err := hnm.updateResetCMStatus(taskName, common.IsolateError, common.ResetError, common.RecoverFailedStatus,
		devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf(failedToUpdateCmPattern, err)
		return err
	}
	return fmt.Errorf("failed to reset task, upgrade recovery failed status")
}

// preProcess write cm info, set task and device in reset
func (hnm *HwAscend910Manager) preProcess(taskName, policy string) (*common.TaskResetInfo, error) {
	devFaultInfoList, err := hnm.hotResetManager.GetTaskDevFaultInfoList(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task device fault info list, err: %v", err)
		return nil, err
	}
	pod, err := hnm.hotResetManager.GetTaskPod(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task pod, err: %v", err)
		return nil, err
	}
	resetInfo, err := hnm.hotResetManager.GetTaskResetInfo(devFaultInfoList, policy, policy, common.UnrecoveredStatus)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task reset info list, err: %v", err)
		return nil, err
	}
	needAddRetryTime := hnm.hotResetManager.IsCurNodeTaskInReset(taskName)
	if _, err := hnm.client.WriteResetInfoDataIntoCM(taskName, pod.Namespace, resetInfo, !needAddRetryTime); err != nil {
		hwlog.RunLog.Errorf("failed to write reset info to cm, err: %v", err)
		return nil, err
	}
	faultInfo, err := hnm.hotResetManager.GetTaskFaultRankInfo(devFaultInfoList)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task fault rank info, err: %v", err)
		return nil, err
	}
	// This CM is used for elastic-agent to generate recover strategy, which is optional for chip reset
	if _, err := hnm.client.WriteFaultInfoDataIntoCM(taskName, pod.Namespace, faultInfo); err != nil {
		hwlog.RunLog.Warnf("failed to write fault rank info to cm, err %v", err)
	}
	if err := hnm.hotResetManager.SetTaskInReset(taskName); err != nil {
		hwlog.RunLog.Errorf("failed to set task %s in reset", taskName)
		return nil, err
	}
	if err := hnm.hotResetManager.SetAllDevInReset(resetInfo); err != nil {
		hwlog.RunLog.Errorf("failed to set all device in reset, err: %v", err)
		return nil, err
	}
	return resetInfo, nil
}

// postProcess clear reset info cm and unset the reset status of all device in a task
func (hnm *HwAscend910Manager) postProcess(taskName string, resetInfo *common.TaskResetInfo) error {
	if err := hnm.hotResetManager.UnSetAllDevInReset(resetInfo); err != nil {
		hwlog.RunLog.Errorf("failed to unset all device in reset, err: %v", err)
		return err
	}
	hwlog.RunLog.Infof("grace tolerance process complete, task name: %s", taskName)
	return nil
}

func (hnm *HwAscend910Manager) refreshDevFaultInfo(devFaultInfo []*common.TaskDevInfo,
	classifyDevs map[string][]*common.NpuDevice) error {
	devStatusList, ok := classifyDevs[common.Ascend910]
	if !ok {
		return fmt.Errorf("not found %s device type in %v", common.Ascend910, devStatusList)
	}

	common.Synchronize = false
	// wait for main process to update the fault cache info. The coroutine performs grace tolerance processing based on
	// the updated fault cache info of the main process
	if err := wait.PollImmediate(time.Second, time.Duration(
		synchronizeWaitMagnification*common.ParamOption.ListAndWatchPeriod)*time.Second, func() (bool, error) {
		if !common.Synchronize {
			hwlog.RunLog.Debug("failed to synchronize the fault cache of the main process, will retry again")
			return false, nil
		}
		for _, npuDevice := range devStatusList {
			if int(npuDevice.LogicID) >= len(devFaultInfo) {
				continue
			}
			devFaultInfo[npuDevice.LogicID].ErrorCode = npuDevice.FaultCodes
			devFaultInfo[npuDevice.LogicID].Policy = hnm.hotResetManager.
				GetDevProcessPolicy(common.GetFaultType(npuDevice.FaultCodes, npuDevice.LogicID))
			hwlog.RunLog.Infof("refresh device fault info, device %d, policy %s, err code: %v", npuDevice.LogicID,
				devFaultInfo[npuDevice.LogicID].Policy, devFaultInfo[npuDevice.LogicID].ErrorCode)
		}
		return true, nil
	}); err != nil {
		return fmt.Errorf("synchronize the fault cache of the main process timeout: %v", err)
	}

	return nil
}

func (hnm *HwAscend910Manager) getNeedResetDevMapForA3(devFaultInfoList []*common.TaskDevInfo) (map[int32]int32,
	error) {
	if common.ParamOption.RealCardType != common.Ascend910A3 {
		return nil, fmt.Errorf("not A3 device")
	}
	needResetDevMap := make(map[int32]int32, len(devFaultInfoList))
	for _, devFaultInfo := range devFaultInfoList {
		policyType, ok := processPolicyTable[devFaultInfo.Policy]
		if !ok {
			err := fmt.Errorf("invalid policy str of device %d", devFaultInfo.LogicId)
			hwlog.RunLog.Error(err)
			return nil, err
		}
		if policyType == common.RestartErrorLevel || policyType == common.ResetErrorLevel ||
			policyType == common.RestartRequestErrorLevel {
			resetIndex, err := hnm.getResetIndexForA3(devFaultInfo.LogicId)
			if err != nil {
				return nil, err
			}
			if _, exist := needResetDevMap[devFaultInfo.LogicId]; !exist {
				needResetDevMap[resetIndex] = devFaultInfo.LogicId
			}
		}
	}
	return needResetDevMap, nil
}

func (hnm *HwAscend910Manager) getResetIndexForA3(logicID int32) (int32, error) {
	cardID, deviceID, err := hnm.GetDmgr().GetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Errorf("get cardID deviceID by logicID %v failed: %v", logicID, err)
		return errorId, err
	}
	logicIDs, err := hnm.getAssociatedLogicIDs(logicID, cardID, deviceID)
	if err != nil {
		return errorId, err
	}

	intNums := make([]int, len(logicIDs))
	for i, v := range logicIDs {
		intNums[i] = int(v)
	}

	sort.Ints(intNums)

	if len(intNums) <= 0 {
		hwlog.RunLog.Errorf("sort logic ids failed, logic ids %v, sorted ids %v", logicIDs, intNums)
		return errorId, fmt.Errorf("sort logic ids failed, logic ids %v, sorted ids %v", logicIDs, intNums)
	}
	hwlog.RunLog.Debugf("related devices logicIDs %v", intNums)
	return int32(intNums[0]), nil
}

func (hnm *HwAscend910Manager) resetDeviceOnce(devFaultInfoList []*common.TaskDevInfo,
	classifyDevs map[string][]*common.NpuDevice) error {
	resetFaultInfoMap, err := hnm.hotResetManager.GetNeedResetDevMap(devFaultInfoList)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get need reset device list, err: %v", err)
		return err
	}
	if devMap, err := hnm.getNeedResetDevMapForA3(devFaultInfoList); err == nil {
		resetFaultInfoMap = devMap
	}
	hwlog.RunLog.Debugf("devices %v will be reset", resetFaultInfoMap)
	processId := time.Now().UnixMilli()
	resetGoroutine.Store(processId, struct{}{})
	if err := hnm.execResetDevice(resetFaultInfoMap); err != nil {
		hwlog.RunLog.Errorf("failed to exec reset device list, err: %v", err)
		resetGoroutine.Delete(processId)
		return err
	}
	resetGoroutine.Delete(processId)
	for _, devInfo := range devFaultInfoList {
		common.SetDeviceInit(devInfo.LogicId)
	}
	if err := hnm.refreshDevFaultInfo(devFaultInfoList, classifyDevs); err != nil {
		hwlog.RunLog.Errorf("failed to refresh device fault info, err: %v", err)
		return err
	}
	return nil
}

func (hnm *HwAscend910Manager) canResetDeviceByLogicID(logicID int32) bool {
	cardID, deviceID, err := hnm.GetDmgr().GetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Errorf("get cardID deviceID by logicID %v failed: %v", logicID, err)
		return false
	}

	return hnm.canResetDevice(cardID, deviceID)
}

func (hnm *HwAscend910Manager) canResetDevice(cardID, deviceID int32) bool {
	if IsDevBusy(cardID, deviceID) {
		hwlog.RunLog.Infof("device is busy, can not reset, cardID %v, deviceID %v", cardID, deviceID)
		return false
	}
	if GetResetCnt(cardID, deviceID) > common.MaxResetTimes {
		hwlog.RunLog.Infof("device reset count %v over limit %v, can not reset, cardID %v, deviceID %v",
			GetResetCnt(cardID, deviceID), common.MaxResetTimes, cardID, deviceID)
		return false
	}
	hwlog.RunLog.Debugf("device can be reset, cardID %v, deviceID %v", cardID, deviceID)
	return true
}

func (hnm *HwAscend910Manager) execResetDevice(devMap map[int32]int32) error {
	errList := make([]error, 0, len(devMap))
	resetFailDevs := make([]ResetDevice, 0)
	resetSuccDevs := make([]ResetDevice, 0)
	var wg sync.WaitGroup
	wg.Add(len(devMap))
	for resetLogicId, faultLogicId := range devMap {
		resetDevMap.Store(resetLogicId, struct{}{})
		go func(resetLogicId, faultLogicId int32) {
			defer wg.Done()
			cardId, deviceId, err := hnm.GetDmgr().GetCardIDDeviceID(resetLogicId)
			if err != nil {
				hwlog.RunLog.Errorf("failed to get reset device card id and device id, err %v", err)
				return
			}
			if !hnm.canResetDevice(cardId, deviceId) {
				return
			}
			shouldCheckNet := hnm.isShouldCheckNet(faultLogicId)
			if err := hnm.tryResetDevice(cardId, deviceId); err != nil {
				errList = append(errList, err)
				resetFailDevs = append(resetFailDevs, ResetDevice{CardId: cardId, DeviceId: deviceId,
					LogicID: resetLogicId, shouldCheckNet: shouldCheckNet})
				return
			}
			// wait for the device to reset completely
			if err := hnm.isRingResetComplete(resetLogicId, shouldCheckNet); err != nil {
				errList = append(errList, err)
				resetFailDevs = append(resetFailDevs, ResetDevice{CardId: cardId, DeviceId: deviceId,
					LogicID: resetLogicId})
				return
			}
			FreeBusyDev(cardId, deviceId)
			resetSuccDevs = append(resetSuccDevs, ResetDevice{CardId: cardId, DeviceId: deviceId,
				LogicID: resetLogicId})
			hwlog.RunLog.Infof("hot reset complete, cardId: %d, logicId: %d", cardId, resetLogicId)
		}(resetLogicId, faultLogicId)
	}
	wg.Wait()
	if len(errList) == 0 {
		hnm.updateResetInfo(resetFailDevs, resetSuccDevs)
		return nil
	}
	if common.ParamOption.RealCardType != common.Ascend910A3 {
		hwlog.RunLog.Info("out band reset only support A3")
		hnm.updateResetInfo(resetFailDevs, resetSuccDevs)
		return errList[0]
	}
	return hnm.execOutBandReset(resetFailDevs, resetSuccDevs)
}

func (hnm *HwAscend910Manager) execOutBandReset(inBandFailDevs, sucDevs []ResetDevice) error {
	var resetError error
	failDevs := make([]ResetDevice, 0, len(inBandFailDevs))
	newSucDevs := make([]ResetDevice, 0, len(inBandFailDevs)+len(sucDevs))
	newSucDevs = append(newSucDevs, sucDevs...)
	var wg sync.WaitGroup
	wg.Add(len(inBandFailDevs))
	for _, dev := range inBandFailDevs {
		go func(dev ResetDevice) {
			defer wg.Done()
			if err := hnm.resetDeviceOutBand(dev.CardId, dev.DeviceId); err != nil {
				resetError = err
				failDevs = append(failDevs, dev)
				return
			}
			// wait for the device to reset completely
			if err := hnm.isRingResetComplete(dev.LogicID, dev.shouldCheckNet); err != nil {
				resetError = err
				return
			}
			newSucDevs = append(newSucDevs, dev)
		}(dev)
	}
	wg.Wait()
	if len(failDevs) < len(inBandFailDevs) {
		time.Sleep(afterRescanDelay * time.Second)
	}
	hnm.updateResetInfo(failDevs, newSucDevs)
	filledFailDevs, err := hnm.fillResetDevs(failDevs)
	if err != nil {
		hwlog.RunLog.Errorf("complement device info err: %v", err)
		return resetError
	}
	go hnm.scanDeviceForThirdParty(filledFailDevs)
	return resetError
}

func (hnm *HwAscend910Manager) scanDeviceForThirdParty(failDevs []ResetDevice) {
	if len(failDevs) <= 0 {
		return
	}
	delay := time.Duration(common.ParamOption.ThirdPartyScanDelay) * time.Second
	time.AfterFunc(delay, func() {
		hnm.execRescan(failDevs)
	})
}

func (hnm *HwAscend910Manager) execRescan(failDevs []ResetDevice) {
	scanFailDevs := make([]ResetDevice, 0, len(failDevs))
	sucDevs := make([]ResetDevice, 0, len(failDevs))
	for _, dev := range failDevs {
		if err := hnm.GetDmgr().RescanSoc(dev.CardId, dev.DeviceId); err != nil {
			hwlog.RunLog.Errorf("fail to rescan cardID %v deviceID %v, error: %v",
				dev.CardId, dev.DeviceId, err)
			scanFailDevs = append(scanFailDevs, dev)
			continue
		}
		FreeBusyDev(dev.CardId, dev.DeviceId)
		sucDevs = append(sucDevs, dev)
	}
	WriteResetInfo(ResetInfo{ThirdPartyResetDevs: failDevs}, WMDelete, false)
	WriteResetInfo(ResetInfo{ManualResetDevs: scanFailDevs}, WMAppend, true)
}

// fillResetDevs complement phyID and associatedCardID
func (hnm *HwAscend910Manager) fillResetDevs(devs []ResetDevice) ([]ResetDevice, error) {
	npuInfos, err := hnm.GetNPUs()
	if err != nil {
		return nil, err
	}
	logicIdMap := make(map[int32]int32)
	for _, dev := range npuInfos.AllDevs {
		logicIdMap[dev.LogicID] = dev.PhyID
	}
	devCopy := make([]ResetDevice, len(devs))
	copy(devCopy, devs)
	for i := range devCopy {
		phyId, exist := logicIdMap[devCopy[i].LogicID]
		if exist {
			devCopy[i].PhyID = phyId
		} else {
			return nil, fmt.Errorf("logicId %v can found, cardID: %v, deviceID: %v",
				devCopy[i].LogicID, devCopy[i].CardId, devCopy[i].DeviceId)
		}
		devCopy[i].AssociatedCardId = errorId
		if common.ParamOption.RealCardType != common.Ascend910A3 {
			continue
		}
		associatedCardId, err := hnm.GetDmgr().GetBrotherCardID(devCopy[i].CardId, devCopy[i].DeviceId)
		if err == nil {
			devCopy[i].AssociatedCardId = associatedCardId
		}
	}
	return devCopy, nil
}

// updateResetInfo update in rest devices, wait third party devices, wait manually devices
func (hnm *HwAscend910Manager) updateResetInfo(failDevs, sucDevs []ResetDevice) {
	if len(failDevs) <= 0 {
		return
	}
	hwlog.RunLog.Infof("reset failed devices: %v, reset success devices: %v", failDevs, sucDevs)
	resetInfo := ReadResetInfo()
	filledFailDevs, err := hnm.fillResetDevs(failDevs)
	if err != nil {
		hwlog.RunLog.Errorf("fail to complement device info, wait manually reset, err: %v", err)
		return
	}
	if common.ParamOption.RealCardType == common.Ascend910A3 {
		resetInfo.ThirdPartyResetDevs = append(resetInfo.ThirdPartyResetDevs, filledFailDevs...)
	} else {
		resetInfo.ManualResetDevs = append(resetInfo.ManualResetDevs, filledFailDevs...)
	}
	WriteResetInfo(resetInfo, WMOverwrite, true)
}

func (hnm *HwAscend910Manager) resetDeviceOutBand(cardId, deviceId int32) error {
	if err := hnm.dmgr.GetOutBandChannelState(cardId, deviceId); err != nil {
		hwlog.RunLog.Warnf("out band channel state error: %v", err)
		return err
	}
	if err := hnm.dmgr.PreResetSoc(cardId, deviceId); err != nil {
		hwlog.RunLog.Errorf("pre reset failed: %v", err)
		return err
	}
	if err := hnm.dmgr.SetDeviceResetOutBand(cardId, deviceId); err != nil {
		hwlog.RunLog.Errorf("reset out band failed: %v", err)
		return err
	}
	time.Sleep(beforeRescanDelay * time.Second)
	if err := hnm.dmgr.RescanSoc(cardId, deviceId); err != nil {
		hwlog.RunLog.Errorf("rescan device failed: %v", err)
		return err
	}
	hwlog.RunLog.Infof("out band reset success, card id: %v, device id: %v", cardId, deviceId)
	return nil
}

func (hnm *HwAscend910Manager) isNetResetCompleted(logicId int32) bool {
	netStatus, err := hnm.GetDmgr().GetDeviceNetWorkHealth(logicId)
	if err != nil {
		hwlog.RunLog.Warnf("get net status of %v error: %v", logicId, err)
		return false
	}
	switch netStatus {
	case networkDetectOK, networkDetectInit:
		return true
	default:
		hwlog.RunLog.Warnf("%d network status is unhealthy, health code is %d", logicId, netStatus)
		return false
	}
}

func (hnm *HwAscend910Manager) waitDeviceResetComplete(logicId int32, totalTime *int, shouldCheckNet bool) error {
	if err := wait.PollImmediate(time.Second, common.WaitDeviceResetTime*time.Second, func() (bool, error) {
		*totalTime += 1
		if *totalTime > common.MaxResetWaitRecoverTime {
			return true, fmt.Errorf("wait device reset recover timeout")
		}
		hwlog.RunLog.Infof("start to check card %d boot status", logicId)
		bootState, err := hnm.GetDmgr().GetDeviceBootStatus(logicId)
		if err != nil {
			hwlog.RunLog.Errorf("get device boot status failed, logic id: %d, err: %v", logicId, err)
			return false, err
		}
		if bootState != common.BootStartFinish {
			hwlog.RunLog.Debugf("device bootState(%d), starting...", bootState)
			return false, nil
		}
		hwlog.RunLog.Infof("card %d start finish", logicId)

		if !shouldCheckNet {
			return true, nil
		}

		return hnm.isNetResetCompleted(logicId), nil
	}); err != nil {
		hwlog.RunLog.Errorf("hot reset failed, timeout or err: %v, logic id: %d", err, logicId)
		return err
	}
	return nil
}

// isRunningDistributed returns true when volcano update 'distributed-job=true' to pod
func (hnm *HwAscend910Manager) isRunningDistributed(logicID int32) bool {
	podMap, err := hnm.hotResetManager.GetFaultDev2PodMap()
	if err != nil {
		hwlog.RunLog.Errorf("failed to get pod while checking the task running mode of dev %v: %v", logicID, err)
		return false
	}

	pod, ok := podMap[logicID]
	if !ok {
		hwlog.RunLog.Errorf("no task running on device %v", logicID)
		return false
	}

	isDistributed, ok := pod.Annotations[common.DistributedJob]
	if !ok {
		return false
	}

	return isDistributed == "true"
}

func (hnm *HwAscend910Manager) isShouldCheckNet(logicID int32) bool {
	// there is no need to check status of network, when running a single node task
	return hnm.isRunningDistributed(logicID)
}

func (hnm *HwAscend910Manager) isRingResetComplete(oriLogicID int32, shouldCheckNet bool) error {
	var totalTime int
	resetDevNumOnce, err := hnm.hotResetManager.GetResetDevNumOnce()
	if err != nil {
		hwlog.RunLog.Error(err)
		return err
	}
	resetStartLogicID := oriLogicID / int32(resetDevNumOnce) * int32(resetDevNumOnce)
	for logicID := resetStartLogicID; logicID < resetStartLogicID+int32(resetDevNumOnce); logicID++ {
		if err := hnm.waitDeviceResetComplete(logicID, &totalTime, shouldCheckNet); err != nil {
			return err
		}
	}
	return nil
}

func (hnm *HwAscend910Manager) tryResetDevice(cardId, deviceId int32) error {
	AddResetCnt(cardId, deviceId)
	AddBusyDev(cardId, deviceId)
	var realError error
	for i := 0; i < common.ResetRetryTimes; i++ {
		hwlog.RunLog.Infof("start to execute cardId %d reset", cardId)
		err := hnm.GetDmgr().SetDeviceReset(cardId, deviceId)
		if err == nil {
			hwlog.RunLog.Infof("execute reset cardId %d success", cardId)
			return nil
		}
		hwlog.RunLog.Errorf("cardId(%d) failed to reset device, err: %v", cardId, err)
		realError = err
		if i != common.ResetRetryTimes-1 {
			time.Sleep(time.Duration(i+1) * common.ResetInterVal * time.Second)
		}
	}
	return realError
}

// tryRescheduleTask writes the isolation info to the reset config map
// so that other nodes don't filter the health of device
func (hnm *HwAscend910Manager) tryWriteIsolationInfo(taskName string) {
	devFaultInfoList, err := hnm.hotResetManager.GetTaskDevFaultInfoList(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task device fault info list, err: %v", err)
		return
	}
	initPolicy, _, err := hnm.hotResetManager.GetTaskProcessPolicy(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task %s process policy, err: %v", taskName, err)
		return
	}
	if err := hnm.updateResetCMStatus(taskName, common.IsolateError, initPolicy, common.RecoverFailedStatus,
		devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf("failed to update reset cm to isolate, err: %v", err)
	}
}

// isDevShouldBeIsolate determines whether device should be isolated
func (hnm *HwAscend910Manager) isDevShouldBeIsolate(faultyDevLogicId int32) bool {
	faultDev2Pod, err := hnm.hotResetManager.GetFaultDev2PodMap()
	if err != nil {
		hwlog.RunLog.Warnf("get faultDev2Pod info err: %v", err)
		return false
	}
	pod, ok := faultDev2Pod[faultyDevLogicId]
	if !ok {
		hwlog.RunLog.Warnf("the dev %#v does not in cache", faultyDevLogicId)
		return false
	}

	taskName, ok := pod.Annotations[common.ResetTaskNameKey]
	if !ok {
		taskName, ok = pod.Labels[common.ResetTaskNameKeyInLabel]
		if !ok {
			hwlog.RunLog.Error("failed to get task name by task key in isDevShouldBeIsolate")
			return true
		}
	}
	resetCM, err := hnm.hotResetManager.GetCMFromCache(pod.Namespace + "/" + common.ResetInfoCMNamePrefix + taskName)
	if err != nil {
		hwlog.RunLog.Warnf("get reset cm error: %v", err)
		return true
	}
	resetInfoData, err := getResetInfoData(resetCM)
	if err != nil {
		hwlog.RunLog.Warnf("get reset info data error: %v", err)
		return true
	}
	if len(resetInfoData) == 0 {
		return true
	}
	for _, rankInfo := range resetInfoData {
		if rankInfo.Policy == common.IsolateError {
			return true
		}
	}

	return false
}

func removeResetDevMap(devFaultInfoList []*common.TaskDevInfo) {
	for _, dev := range devFaultInfoList {
		resetDevMap.Delete(dev.LogicId)
	}
}
