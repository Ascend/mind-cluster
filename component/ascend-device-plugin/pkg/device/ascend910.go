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
	"strings"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/api"
	"ascend-common/api/label"
	"ascend-common/common-utils/hwlog"
	npuCommon "ascend-common/devmanager/common"
)

const (
	deviceA3Id0       = 0 // die id 0 of A3 card
	deviceA3Id1       = 1 // dir id 1 of A3 card
	ringNumOfA3       = 2 // device number in a ring
	otherCardIncrease = 1
)

var (
	lastTimeNetworkRecoverDevices = sets.String{}
	hotResetManagerInitOnce       sync.Once
)

// HwAscend910Manager manages huawei Ascend910 devices.
type HwAscend910Manager struct {
	AscendTools
	hotResetManager HotResetManager
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

// SyncCM sync hot reset cm with kubernetes
func (hnm *HwAscend910Manager) SyncCM(ctx context.Context, classifyDevs map[string][]*common.NpuDevice) {
	hotResetManagerInitOnce.Do(func() {
		hnm.hotResetManager = NewHotResetManager(len(classifyDevs[hnm.name]), hnm.boardId)
		if hnm.hotResetManager == nil {
			hwlog.RunLog.Error("hot reset manager is nil")
			return
		}
		hnm.hotResetManager.SyncResetCM(ctx, hnm.GetKubeClient())
	})
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

// DoWithVolcanoListAndWatch ascend910 affinity scheduling
func (hnm *HwAscend910Manager) DoWithVolcanoListAndWatch(classifyDevs map[string][]*common.NpuDevice, chipMemory int) {
	devStatusSet := hnm.getDevStatesDevSet(classifyDevs, chipMemory)
	if err := hnm.UpdateNodeDeviceInfo(devStatusSet, hnm.updateDeviceInfo); err != nil {
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

	newDevInfo[common.GetAscend910Key(api.CmCardUnhealthySuffix)] = common.ToString(devStatusSet.UnHealthyDevice,
		common.CommaSepDev)
	newDevInfo[common.GetAscend910Key(api.CmCardNetworkUnhealthySuffix)] = common.ToString(newNetUHDevSets,
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
	if boardId != common.A800IA2NoneHccsBoardId &&
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
		devTypeAndID := strings.Split(devName, npuCommon.Minus)
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
	devRecoverVal, _ := label.GetNodeLabel(curNode, common.GetAscend910Key(api.NodeLabelRecoverSuffix))
	*devRecoverLabel = hnm.toStandardDeviceFmt(common.ConvertDevListToSets(devRecoverVal, common.DotSepDev))
	// netRecoverLabel like Ascend910-0,Ascend910-2,Ascend910-3, means dev network exception
	netRecoverVal, _ := label.GetNodeLabel(curNode, common.GetAscend910Key(api.NodeLabelNetworkRecoverSuffix))
	*netRecoverLabel = hnm.toStandardDeviceFmt(common.ConvertDevListToSets(netRecoverVal, common.DotSepDev))
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
