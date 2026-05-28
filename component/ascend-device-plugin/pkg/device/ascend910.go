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
	networkDetectOK   = uint32(0)
	networkDetectInit = uint32(6)
	beforeRescanDelay = 3 // seconds sleep before rescan devices
	afterRescanDelay  = 2 // seconds sleep after rescan devices
	deviceA3Id0       = 0 // die id 0 of A3 card
	deviceA3Id1       = 1 // dir id 1 of A3 card
	ringNumOfA3       = 2 // device number in a ring
	otherCardIncrease = 1
	errorId           = -1
)

var (
	lastTimeNetworkRecoverDevices = sets.String{}
	hotResetManagerInitOnce       sync.Once
	isHotResetOn                        = false
	inResetDev                    int32 = -1
	isolateDevList                []int32
	resetTimeMap                  = &sync.Map{}
	offlineInBandFailLogicId      = sync.Map{}
)

// HwAscend910Manager manages huawei Ascend910 devices.
type HwAscend910Manager struct {
	AscendTools
	dpu             common.DpuInfo
	hotResetManager HotResetManager
}

func getResetTime(logicId int32) int64 {
	tmpResetTime, ok := resetTimeMap.Load(logicId)
	if !ok {
		return 0
	}
	resetTime, ok := tmpResetTime.(int64)
	if !ok {
		return 0
	}
	return resetTime
}

// getAscend910Name returns the device name based on the real card type.
// For Ascend910A5, it returns api.NPULowerCase; otherwise, it returns api.Ascend910.
func getAscend910Name() string {
	if common.ParamOption.RealCardType == api.Ascend910A5 {
		return api.NPULowerCase
	}
	return api.Ascend910
}

// NewHwAscend910Manager is used to create ascend npu manager
func NewHwAscend910Manager() *HwAscend910Manager {
	return &HwAscend910Manager{
		AscendTools: AscendTools{
			name:                      getAscend910Name(),
			unHealthyKey:              common.GetAscend910Key(api.CmCardUnhealthySuffix),
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
		hnm.hotResetManager = NewHotResetManager(hnm.GetDeviceUsage(), len(classifyDevs[hnm.name]), hnm.boardId)
		if hnm.hotResetManager == nil {
			hwlog.RunLog.Error("hot reset manager is nil")
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
	if isHotResetOn {
		return nil
	}
	deviceList, ok := classifyDevs[hnm.name]
	if !ok {
		return fmt.Errorf("device list not found, %v", hnm.name)
	}
	resetDevs := make([]*common.NpuDevice, 0, len(deviceList))
	resetFaultInfos := make([]*common.DevFaultInfo, 0, len(deviceList))
	isHotResetOn = true
	resetRing := make(map[int32]struct{})
	var podList *v1.PodList = nil
	for _, dev := range deviceList {
		tempFaultInfo := hnm.getDevFaultInfo(dev.LogicID)
		if tempFaultInfo == nil {
			resetTimeMap.Delete(dev.LogicID)
			continue
		}
		idx, err := hnm.getResetIndex(dev.LogicID)
		if err != nil {
			continue
		}
		if _, exist := resetRing[idx]; exist {
			continue
		}
		if canReset, err := hnm.canBeReset(tempFaultInfo, podList); err != nil || !canReset {
			hwlog.RunLog.Infof("device %v cannot reset, it is busy, err: %v", tempFaultInfo.LogicId, err)
			continue
		}
		if restartFlag := hnm.isFaultNeedRestart(tempFaultInfo); !restartFlag {
			continue
		}
		resetRing[idx] = struct{}{}
		resetDevs = append(resetDevs, dev)
		resetFaultInfos = append(resetFaultInfos, tempFaultInfo)
		hwlog.RunLog.Debugf("found %v error on device %v, will start reset process "+
			"whenever all chips are free on ring", tempFaultInfo.Policy, dev.DeviceName)
	}
	for idx, dev := range resetDevs {
		if err := hnm.startUpHotReset(classifyDevs, resetFaultInfos[idx], dev); err != nil {
			hwlog.RunLog.Errorf("failed to start up hot reset, err: %v", err)
		}
	}
	hnm.hotResetTryOutBand(resetDevs)
	isHotResetOn = false
	return nil
}

func (hnm *HwAscend910Manager) getDevFaultInfo(logicID int32) *common.DevFaultInfo {
	tempFaultInfo, tempErr := hnm.hotResetManager.GetGlobalDevFaultInfo(logicID)
	if tempErr != nil {
		hwlog.RunLog.Errorf("failed to get global device fault info from cache, err: %v", tempErr)
		return nil
	}
	if tempFaultInfo.Policy == common.EmptyError || tempFaultInfo.Policy == common.IsolateError {
		return nil
	}
	return tempFaultInfo
}

func (hnm *HwAscend910Manager) getResetIndex(logicID int32) (int32, error) {
	if common.ParamOption.RealCardType == api.Ascend910A3 {
		if idx, err := hnm.getResetIndexForA3(logicID); err == nil {
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
	return (logicID / int32(resetDevNumOnce)) * int32(resetDevNumOnce), nil
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
	if common.ParamOption.RealCardType != api.Ascend910A3 {
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

// isFaultNeedRestart restarts when faults handling failed
func (hnm *HwAscend910Manager) isFaultNeedRestart(devFaultInfo *common.DevFaultInfo) bool {
	if common.ParamOption.RealCardType == api.Ascend910B &&
		(devFaultInfo.Policy == common.FreeResetError || devFaultInfo.Policy == common.ResetError) {
		return true
	}
	resetTime := getResetTime(devFaultInfo.LogicId)
	if resetTime == 0 {
		resetTimeMap.Store(devFaultInfo.LogicId, time.Now().Unix())
		return false
	}
	if time.Now().Unix()-resetTime > common.ResetFaultToleranceTimeInterval {
		hwlog.RunLog.Infof("device %v fault exist over 60s", devFaultInfo.LogicId)
		resetTimeMap.Delete(devFaultInfo.LogicId)
		return true
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
	devStatusList, ok := classifyDevs[hnm.name]
	if !ok {
		return fmt.Errorf("no ascend npu device needed filter")
	}
	clearDeviceStatus(devStatusList)
	if !isHotResetOn {
		return nil
	}
	if inResetDev == -1 {
		hwlog.RunLog.Debug("should not set device to unhealthy")
		return nil
	}
	if common.ParamOption.RealCardType == api.Ascend910A3 &&
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
	logicIdArr, err := hnm.GetAssociatedLogicIDs(dev.LogicID, dev.CardID, dev.DeviceID)
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

func (hnm *HwAscend910Manager) GetAssociatedLogicIDs(logicID, cardID, deviceID int32) ([]int32, error) {
	associatedCardID, err := hnm.GetDmgr().GetBrotherCardID(logicID)
	if err != nil {
		hwlog.RunLog.Debugf("get brother card failed, cardID %v deviceID %v, err: %v",
			cardID, deviceID, err)
		return nil, err
	}
	logicID0, err := hnm.GetDmgr().GetDeviceLogicID(associatedCardID, deviceA3Id0)
	if err != nil {
		hwlog.RunLog.Errorf("get logicID failed by cardID %v deviceID %v, err: %v",
			associatedCardID, deviceA3Id0, err)
		return nil, err
	}
	logicID1, err := hnm.GetDmgr().GetDeviceLogicID(associatedCardID, deviceA3Id1)
	if err != nil {
		hwlog.RunLog.Errorf("get logicID failed by cardID %v deviceID %v, err: %v",
			associatedCardID, deviceA3Id1, err)
		return nil, err
	}
	// get the other device id in a ring
	otherDeviceId := (deviceID + otherCardIncrease) % ringNumOfA3
	ringDevLogic, err := hnm.GetDmgr().GetDeviceLogicID(cardID, otherDeviceId)
	if err != nil {
		hwlog.RunLog.Errorf("get logicID failed by cardID %v deviceID %v, err: %v",
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
	if err := hnm.execHotReset(classifyDevs, devInfo); err != nil {
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

func (hnm *HwAscend910Manager) checkFaultIsExist(devs map[string][]*common.NpuDevice, logicID int32) bool {
	devList, ok := devs[hnm.name]
	if !ok {
		hwlog.RunLog.Error("no ascend npu device, upgrade hot reset error fail")
		// get error consider fault exist
		return true
	}
	resetIdx, err := hnm.getResetIndex(logicID)
	if err != nil {
		hwlog.RunLog.Errorf("get reset index failed, err: %v", err)
		// get error consider fault exist
		return true
	}
	for _, dev := range devList {
		idx, err := hnm.getResetIndex(dev.LogicID)
		if err != nil || idx != resetIdx {
			continue
		}
		_, errCodes, getErr := hnm.dmgr.GetDeviceAllErrorCode(dev.LogicID)
		if getErr != nil {
			hwlog.RunLog.Errorf("failed to get device error code, err %v", getErr)
			// get error consider fault exist
			return true
		}
		if len(errCodes) == 0 {
			continue
		}
		hwlog.RunLog.Infof("reset device id <%v> fault <%#v> is exist", dev.LogicID, errCodes)
		resetPolicy := hnm.hotResetManager.GetDevProcessPolicy(common.GetFaultType(errCodes, dev.LogicID))
		if resetPolicy == common.RestartRequestError || resetPolicy == common.RestartError ||
			resetPolicy == common.FreeResetError || resetPolicy == common.ResetError {
			hwlog.RunLog.Infof("device id <%v> fault is exist", dev.LogicID)
			return true
		}
	}
	return false
}

func (hnm *HwAscend910Manager) upgradeHotResetError(classifyDevs map[string][]*common.NpuDevice,
	npuDev *common.NpuDevice) {
	isolateDevList = append(isolateDevList, npuDev.LogicID)
	devStatusList, ok := classifyDevs[hnm.name]
	if !ok {
		hwlog.RunLog.Error("no ascend npu device, upgrade hot reset error fail")
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

func (hnm *HwAscend910Manager) execHotReset(classifyDevs map[string][]*common.NpuDevice,
	devInfo *common.DevFaultInfo) error {
	logicID := devInfo.LogicId

	shouldCheckNet := hnm.isShouldCheckNet(logicID)
	if err := hnm.tryResetDeviceOffline(classifyDevs, logicID); err != nil {
		offlineInBandFailLogicId.Store(logicID, struct{}{})
		hwlog.RunLog.Errorf("failed to reset device, err %v", err)
		FreeBusyDev(logicID)
		return err
	}
	if err := hnm.isRingResetComplete(logicID, shouldCheckNet); err != nil {
		hwlog.RunLog.Errorf("fail while waiting for hot reset complete, err %v", err)
		return err
	}
	resetTimeMap.Delete(logicID)
	FreeBusyDev(logicID)
	hwlog.RunLog.Infof("hot reset complete, logicId: %d", logicID)
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
func (hnm *HwAscend910Manager) canBeReset(dev *common.DevFaultInfo, podList *v1.PodList) (bool, error) {
	if !hnm.canResetDeviceByLogicID(dev.LogicId) {
		return false, nil
	}
	if podList == nil {
		var err error
		podList, err = hnm.client.GetAllPodList()
		if err != nil {
			hwlog.RunLog.Errorf("get pod list fail, err %v", err)
			return false, err
		}
	}
	if common.ParamOption.RealCardType == api.Ascend910A3 &&
		hnm.canA3BeReset(dev, podList) {
		return true, nil
	}
	oriLogicID := dev.LogicId
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

func (hnm *HwAscend910Manager) canA3BeReset(dev *common.DevFaultInfo, podList *v1.PodList) bool {
	if podList == nil {
		return false
	}
	cardID, deviceID, err := hnm.GetDmgr().GetCardIDDeviceID(dev.LogicId)
	if err != nil {
		hwlog.RunLog.Errorf("get cardID deviceID by logicID %v failed: %v", dev.LogicId, err)
		return false
	}
	logicIdArr, err := hnm.GetAssociatedLogicIDs(dev.LogicId, cardID, deviceID)
	if err != nil {
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
		annotationTag := fmt.Sprintf("%s%s", api.ResourceNamePrefix, hnm.name)
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
func (hnm *HwAscend910Manager) DoWithVolcanoListAndWatch(classifyDevs map[string][]*common.NpuDevice, chipMemory int) {
	devStatusSet := hnm.getDevStatesDevSet(classifyDevs, chipMemory)
	if err := hnm.UpdateNodeDeviceInfo(devStatusSet, hnm.dpu, hnm.updateDeviceInfo); err != nil {
		hwlog.RunLog.Errorf("update device info failed, err: %v", err)
	}
}

func (hnm *HwAscend910Manager) updateDeviceInfo(oldDevInfo, newDevInfo map[string]string,
	devStatusSet common.DevStatusSet) error {
	if newDevInfo == nil {
		return fmt.Errorf("invalid new device info")
	}
	nodeFmtDevRecover, nodeFmtDevNetRecover := sets.String{}, sets.String{}
	newDevRecoverLabel, newAscend910 := hnm.getHealthAndRecoverDev(devStatusSet, nodeFmtDevRecover,
		common.ConvertDevListToSets(oldDevInfo[common.GetAscend910Key(api.CmCardUnhealthySuffix)],
			common.CommaSepDev))
	newNetRecoverSets, newNetUHDevSets := hnm.getNewNetworkRecoverDev(devStatusSet.NetUnHealthyDevice,
		common.ConvertDevListToSets(oldDevInfo[common.GetAscend910Key(api.CmCardNetworkUnhealthySuffix)],
			common.CommaSepDev),
		nodeFmtDevNetRecover)
	newDevInfo[common.GetAscend910Key("")] = newAscend910
	newDevInfo[common.GetAscend910Key(api.CmRecoveringSuffix)] = common.ToString(devStatusSet.RecoveringDevices,
		common.CommaSepDev)
	// hnm.isNeedBlockAllDevice: server is A800IA2 with hccs and there are fault devices or is already in resetting,
	// no more pod should be scheduled to this node cause all npu resetting is on the way
	// if reset failed more than ResetRetryTimes times, will no longer try to reset server
	if common.ParamOption.HotReset == common.HotResetInfer &&
		hnm.GetResetFailedTimes(common.FirstDevice) <= common.MaxResetTimes &&
		hnm.isNeedBlockAllDevice(devStatusSet.DeviceFault) {

		newDevInfo[common.GetAscend910Key("")] = ""
		newDevInfo[common.GetAscend910Key(api.CmRecoveringSuffix)] = common.ToString(devStatusSet.AllDevices,
			common.CommaSepDev)
		hwlog.RunLog.Warnf("all device on node have been cleared, due to resetting all devices in process")
	}

	newDevInfo[common.GetAscend910Key(api.CmCardUnhealthySuffix)] = common.ToString(devStatusSet.UnHealthyDevice,
		common.CommaSepDev)
	newDevInfo[common.GetAscend910Key(api.CmCardNetworkUnhealthySuffix)] = common.ToString(newNetUHDevSets,
		common.CommaSepDev)
	newDevInfo[common.GetAscend910Key(api.CmCardDPUUnhealthySuffix)] = common.ToString(devStatusSet.DpuUnHealthyDevice,
		common.CommaSepDev)
	var data []byte
	if data = common.MarshalData(devStatusSet.DeviceFault); len(data) == 0 {
		return fmt.Errorf("device fault code marshal failed")
	}
	newDevInfo[common.GetAscend910Key(api.CmFaultListSuffix)] = string(data)
	if common.ParamOption.AutoStowingDevs {
		return nil
	}
	// plan to sunset 709~718
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
	newNode.Labels[common.GetAscend910Key(api.NodeLabelRecoverSuffix)] = devRecoverLabel
	newNode.Labels[common.GetAscend910Key(api.NodeLabelNetworkRecoverSuffix)] = netRecoverLabel
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
	device910 := curDevStatusSet.FreeHealthyDevice[hnm.name]
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
		curNode.Labels[common.GetAscend910Key(api.NodeLabelRecoverSuffix)], common.DotSepDev))
	// netRecoverLabel like Ascend910-0,Ascend910-2,Ascend910-3, means dev network exception
	*netRecoverLabel = hnm.toStandardDeviceFmt(common.ConvertDevListToSets(
		curNode.Labels[common.GetAscend910Key(api.NodeLabelNetworkRecoverSuffix)], common.DotSepDev))
	return curNode, nil
}

// toStandardDeviceFmt convert physical id "x" to format "Ascend910-x"
func (hnm *HwAscend910Manager) toStandardDeviceFmt(devices sets.String) sets.String {
	if devices.Len() == 0 {
		return sets.String{}
	}

	standardSets := sets.String{}
	for devID := range devices {
		deviceName := fmt.Sprintf("%s-%s", hnm.name, devID)
		standardSets.Insert(deviceName)
	}

	return standardSets
}

func (hnm *HwAscend910Manager) updateHotResetCache(classifyDevs map[string][]*common.NpuDevice) error {
	deviceList, ok := classifyDevs[hnm.name]
	if !ok {
		hwlog.RunLog.Error("ascend npu device list no found")
		return fmt.Errorf("ascend npu device list not found")
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
	deviceList, ok := classifyDevs[hnm.name]
	if !ok {
		return fmt.Errorf("no Ascend npu device found in cache")
	}
	for _, dev := range deviceList {
		index := -1
		for i := range isolateDevList {
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
		tmpNpu, ok := pod.Annotations[api.HuaweiAscend910]
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
		if common.ParamOption.RealCardType == api.Ascend910B && hnm.GetDeviceUsage() == common.Infer {
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
func (hnm *HwAscend910Manager) filterDevStatus(classifyDevs map[string][]*common.NpuDevice) error {
	devStatusList, ok := classifyDevs[hnm.name]
	if !ok {
		return fmt.Errorf("no ascend npu device needed filter")
	}
	if common.ParamOption.RealCardType == api.Ascend910A3 &&
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
		logicIdArr, err := hnm.GetAssociatedLogicIDs(dev.LogicID, dev.CardID, dev.DeviceID)
		if err != nil {
			return err
		}
		for _, id := range logicIdArr {
			devToBeSet[id] = struct{}{}
		}
	}
	for idx := range devToBeSet {
		if int(idx) >= len(devStatusList) {
			hwlog.RunLog.Errorf("device logicID %v is greater than device list length %v",
				idx, len(devStatusList))
			continue
		}
		devStatusList[idx].NetworkHealth = v1beta1.Healthy
	}
	return nil
}
func (hnm *HwAscend910Manager) getResetIndexForA3(logicID int32) (int32, error) {
	cardID, deviceID, err := hnm.GetDmgr().GetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Errorf("get cardID deviceID by logicID %v failed: %v", logicID, err)
		return errorId, err
	}
	logicIDs, err := hnm.GetAssociatedLogicIDs(logicID, cardID, deviceID)
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
func (hnm *HwAscend910Manager) canResetDeviceByLogicID(logicID int32) bool {
	return hnm.canResetDevice(logicID)
}

func (hnm *HwAscend910Manager) canResetDevice(logicID int32) bool {
	if IsDevBusy(logicID) {
		hwlog.RunLog.Infof("device is busy, can not reset, logicID %v", logicID)
		return false
	}
	if GetResetCnt(logicID) > common.MaxResetTimes {
		hwlog.RunLog.Infof("device reset count %v over limit %v, can not reset, logicID %v",
			GetResetCnt(logicID), common.MaxResetTimes, logicID)
		return false
	}
	hwlog.RunLog.Debugf("device can be reset, logicID %v", logicID)
	return true
}
func (hnm *HwAscend910Manager) execOutBandReset(inBandFailDevs, sucDevs []ResetDevice) error {
	failDevs := make([]ResetDevice, 0, len(inBandFailDevs))
	newSucDevs := make([]ResetDevice, 0, len(inBandFailDevs)+len(sucDevs))
	newSucDevs = append(newSucDevs, sucDevs...)
	var resetError error
	for _, dev := range inBandFailDevs {
		if err := hnm.resetDeviceOutBand(dev.LogicID); err != nil {
			resetError = err
			failDevs = append(failDevs, dev)
			continue
		}
		// wait for the device to reset completely
		if err := hnm.isRingResetComplete(dev.LogicID, dev.shouldCheckNet); err != nil {
			resetError = err
			continue
		}
		newSucDevs = append(newSucDevs, dev)
	}
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
		if err := hnm.GetDmgr().RescanSoc(dev.LogicID); err != nil {
			hwlog.RunLog.Errorf("fail to rescan logicID %v, error: %v", dev.LogicID, err)
			scanFailDevs = append(scanFailDevs, dev)
			continue
		}
		FreeBusyDev(dev.LogicID)
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
			return nil, fmt.Errorf("logicId %v can found", devCopy[i].LogicID)
		}
		devCopy[i].AssociatedCardId = errorId
		if common.ParamOption.RealCardType != api.Ascend910A3 {
			continue
		}
		associatedCardId, err := hnm.GetDmgr().GetBrotherCardID(devCopy[i].LogicID)
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
	if common.ParamOption.RealCardType == api.Ascend910A3 {
		resetInfo.ThirdPartyResetDevs = append(resetInfo.ThirdPartyResetDevs, filledFailDevs...)
	} else {
		resetInfo.ManualResetDevs = append(resetInfo.ManualResetDevs, filledFailDevs...)
	}
	WriteResetInfo(resetInfo, WMOverwrite, true)
}

func (hnm *HwAscend910Manager) resetDeviceOutBand(logicID int32) error {
	if err := hnm.dmgr.GetOutBandChannelState(logicID); err != nil {
		hwlog.RunLog.Warnf("out band channel state error: %v", err)
		return err
	}
	if err := hnm.dmgr.PreResetSoc(logicID); err != nil {
		hwlog.RunLog.Errorf("pre reset failed: %v", err)
		return err
	}
	if err := hnm.dmgr.SetDeviceResetOutBand(logicID); err != nil {
		hwlog.RunLog.Errorf("reset out band failed: %v", err)
		return err
	}
	time.Sleep(beforeRescanDelay * time.Second)
	if err := hnm.dmgr.RescanSoc(logicID); err != nil {
		hwlog.RunLog.Errorf("rescan device failed: %v", err)
		return err
	}
	hwlog.RunLog.Infof("out band reset success, logic id: %v", logicID)
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
func (hnm *HwAscend910Manager) tryResetDeviceOffline(classifyDevs map[string][]*common.NpuDevice, logicId int32) error {
	AddResetCnt(logicId)
	AddBusyDev(logicId)
	var realError error = nil
	for i := 0; i < common.ResetRetryTimes; i++ {
		if !hnm.checkFaultIsExist(classifyDevs, logicId) {
			hwlog.RunLog.Infof("device id <%v> fault is not exist, stop reset device", logicId)
			return nil
		}
		hwlog.RunLog.Infof("start to execute logicId %d reset", logicId)
		err := hnm.GetDmgr().SetDeviceReset(logicId)
		if err == nil {
			hwlog.RunLog.Infof("execute reset logicId %d success", logicId)
			return nil
		}
		hwlog.RunLog.Errorf("logicId(%d) failed to reset device, err: %v", logicId, err)
		realError = err
		if i != common.ResetRetryTimes-1 {
			time.Sleep(time.Duration(i+1) * common.ResetInterVal * time.Second)
		}
	}
	return realError
}
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

// SetDpu writes dpuInfo into HwAscend910Manager
func (hnm *HwAscend910Manager) SetDpu(busType string, dpuList []common.DpuCMData, npuToDpusMap map[string][]string) {
	hnm.dpu = common.DpuInfo{
		BusType:      busType,
		DPUList:      dpuList,
		NpuToDpusMap: npuToDpusMap,
		UpdateTime:   time.Now().Unix(),
	}
}
