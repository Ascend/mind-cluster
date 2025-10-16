// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package constant a series of para
package constant

import (
	"maps"
	"time"

	"k8s.io/utils/strings/slices"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/util"
)

var normalFaultLevel = []string{NotHandleFault, SubHealthFault, NormalNPU, NormalNetwork}

func (cm *AdvanceDeviceFaultCm) addFault(addFault DeviceFault) bool {
	if cm.FaultDeviceList == nil {
		cm.FaultDeviceList = make(map[string][]DeviceFault)
	}
	if _, ok := cm.FaultDeviceList[addFault.NPUName]; !ok {
		cm.FaultDeviceList[addFault.NPUName] = make([]DeviceFault, 0)
	}
	deviceFaults := cm.FaultDeviceList[addFault.NPUName]
	found := false
	for _, fault := range deviceFaults {
		if equalDeviceFault(&addFault, &fault) {
			found = true
			break
		}
	}
	if !found {
		deviceFaults = append(deviceFaults, addFault)
	}
	cm.FaultDeviceList[addFault.NPUName] = deviceFaults
	return !found
}

// AddFaultAndFix add fault in the AdvanceDeviceFaultCm
// If the fault is more than normalFaultLevel, then should add into CardUnHealthy/NetworkUnhealthy
// And remove from AvailableDeviceList
func (cm *AdvanceDeviceFaultCm) AddFaultAndFix(addFault DeviceFault) {
	if !cm.addFault(addFault) {
		return
	}
	if !slices.Contains(normalFaultLevel, addFault.FaultLevel) {
		cm.AvailableDeviceList = util.DeleteStringSliceItem(cm.AvailableDeviceList, addFault.NPUName)
		if addFault.FaultType == CardUnhealthy || addFault.FaultType == PublicFaultType {
			if !slices.Contains(cm.CardUnHealthy, addFault.NPUName) {
				cm.CardUnHealthy = append(cm.CardUnHealthy, addFault.NPUName)
			}
		} else if addFault.FaultType == CardNetworkUnhealthy {
			if !slices.Contains(cm.NetworkUnhealthy, addFault.NPUName) {
				cm.NetworkUnhealthy = append(cm.NetworkUnhealthy, addFault.NPUName)
			}
		} else {
			hwlog.RunLog.Errorf("unrecognizable fault type %s", addFault.FaultType)
		}
	}
}

func (cm *AdvanceDeviceFaultCm) delFault(delFault DeviceFault) bool {
	if cm.FaultDeviceList == nil {
		return false
	}
	if _, ok := cm.FaultDeviceList[delFault.NPUName]; !ok {
		return false
	}
	deviceFaults := cm.FaultDeviceList[delFault.NPUName]

	newDeviceFaults := make([]DeviceFault, 0)
	found := false
	for _, fault := range deviceFaults {
		if equalDeviceFault(&delFault, &fault) {
			found = true
			continue
		}
		newDeviceFaults = append(newDeviceFaults, fault)
	}
	if len(newDeviceFaults) == 0 {
		delete(cm.FaultDeviceList, delFault.NPUName)
	} else {
		cm.FaultDeviceList[delFault.NPUName] = newDeviceFaults
	}
	return found
}

// DelFaultAndFix delete fault in the AdvanceDeviceFaultCm
// Delete fault cannot add npu into AvailableDeviceList, because some job run on the npu
func (cm *AdvanceDeviceFaultCm) DelFaultAndFix(delFault DeviceFault) {
	if !cm.delFault(delFault) {
		return
	}
	deviceFaults := cm.FaultDeviceList[delFault.NPUName]
	delFromCardUnhealthy := true
	delFromCardNetworkUnhealthy := true
	for _, devFault := range deviceFaults {
		if !slices.Contains(normalFaultLevel, devFault.FaultLevel) {
			if devFault.FaultType == CardUnhealthy || devFault.FaultType == PublicFaultType {
				delFromCardUnhealthy = false
			} else if devFault.FaultType == CardNetworkUnhealthy {
				delFromCardNetworkUnhealthy = false
			} else {
				hwlog.RunLog.Errorf("unrecognizable fault type %s", devFault.FaultType)
			}
		}
	}
	if delFromCardUnhealthy {
		cm.CardUnHealthy = util.DeleteStringSliceItem(cm.CardUnHealthy, delFault.NPUName)
	}
	if delFromCardNetworkUnhealthy {
		cm.NetworkUnhealthy = util.DeleteStringSliceItem(cm.NetworkUnhealthy, delFault.NPUName)
	}
}

// IsSame compare two AdvanceDeviceFaultCm, do not care UpdateTime
func (cm *AdvanceDeviceFaultCm) IsSame(another ConfigMapInterface) bool {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
		return false
	}
	thatCm, ok := another.(*AdvanceDeviceFaultCm)
	if !ok {
		return false
	}
	eq := func(faultListOne []DeviceFault, faultListOther []DeviceFault) bool {
		if len(faultListOne) != len(faultListOther) {
			return false
		}
		for i, fault := range faultListOne {
			if !equalDeviceFault(&fault, &faultListOther[i]) {
				return false
			}
		}
		return true
	}
	return cm.DeviceType == thatCm.DeviceType &&
		cm.CmName == thatCm.CmName &&
		cm.SuperPodID == thatCm.SuperPodID &&
		cm.ServerIndex == thatCm.ServerIndex &&
		slices.Equal(cm.AvailableDeviceList, thatCm.AvailableDeviceList) &&
		slices.Equal(cm.Recovering, thatCm.Recovering) &&
		slices.Equal(cm.CardUnHealthy, thatCm.CardUnHealthy) &&
		slices.Equal(cm.NetworkUnhealthy, thatCm.NetworkUnhealthy) &&
		maps.EqualFunc(cm.FaultDeviceList, thatCm.FaultDeviceList, eq)
}

// UpdateFaultReceiveTime update fault receive time
func (cm *AdvanceDeviceFaultCm) UpdateFaultReceiveTime(oldInfo ConfigMapInterface) {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
		return
	}
	oldCm, ok := oldInfo.(*AdvanceDeviceFaultCm)
	if !ok || oldCm == nil {
		hwlog.RunLog.Error("oldInfo convert to AdvanceDeviceFaultCm failed or oldCm is nil")
		updateFaultReceiveTimeForDevices(cm.FaultDeviceList, nil)
		return
	}
	updateFaultReceiveTimeForDevices(cm.FaultDeviceList, oldCm.FaultDeviceList)
	hwlog.RunLog.Debugf("updateFaultReceiveTimeForDevices cm.CmName=%s, cm.FaultDeviceList=%#v",
		cm.CmName, cm.FaultDeviceList)
}

func updateFaultReceiveTimeForDevices(cmFaults map[string][]DeviceFault, oldFaults map[string][]DeviceFault) {
	deviceOldTimeMap := buildDeviceOldTimeMap(oldFaults)
	hwlog.RunLog.Debugf("updateFaultReceiveTimeForDevices deviceOldTimeMap=%#v", deviceOldTimeMap)
	for deviceName, faults := range cmFaults {
		updateDeviceFaults(faults, deviceOldTimeMap[deviceName])
	}
}

func buildDeviceOldTimeMap(oldFaults map[string][]DeviceFault) map[string]map[string]int64 {
	oldTimeMap := make(map[string]map[string]int64)
	if oldFaults == nil {
		return oldTimeMap
	}

	for deviceName, faults := range oldFaults {
		codeTimeMap := make(map[string]int64)
		for _, fault := range faults {
			for code, timeAndLevel := range fault.FaultTimeAndLevelMap {
				codeTimeMap[code] = timeAndLevel.FaultReceivedTime
			}
		}
		oldTimeMap[deviceName] = codeTimeMap
	}
	return oldTimeMap
}

func updateDeviceFaults(faults []DeviceFault, oldCodeTimes map[string]int64) {
	for i, fault := range faults {
		for code, timeAndLevel := range fault.FaultTimeAndLevelMap {
			if oldTime, ok := oldCodeTimes[code]; ok {
				timeAndLevel.FaultReceivedTime = oldTime
			} else if timeAndLevel.FaultReceivedTime == 0 {
				timeAndLevel.FaultReceivedTime = time.Now().UnixMilli()
			}
			faults[i].FaultTimeAndLevelMap[code] = timeAndLevel
		}
	}
}

// GetCmName return cm name
func (cm *AdvanceDeviceFaultCm) GetCmName() string {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
		return ""
	}
	return cm.CmName
}

// GetRecoveringKey return cm RecoveringKey
func (cm *AdvanceDeviceFaultCm) GetRecoveringKey() string {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
		return ""
	}
	return api.ResourceNamePrefix + cm.DeviceType + CmRecoveringSuffix
}

// GetCardUnHealthyKey return cm CardUnHealthyKey
func (cm *AdvanceDeviceFaultCm) GetCardUnHealthyKey() string {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
		return ""
	}
	return api.ResourceNamePrefix + cm.DeviceType + CmCardUnhealthySuffix
}

// GetNetworkUnhealthyKey return cm NetworkUnhealthyKey
func (cm *AdvanceDeviceFaultCm) GetNetworkUnhealthyKey() string {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
		return ""
	}
	return api.ResourceNamePrefix + cm.DeviceType + CmCardNetworkUnhealthySuffix
}

// GetFaultDeviceListKey return cm FaultDeviceListKey
func (cm *AdvanceDeviceFaultCm) GetFaultDeviceListKey() string {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
		return ""
	}
	return api.ResourceNamePrefix + cm.DeviceType + CmFaultListSuffix
}

// GetAvailableDeviceListKey return cm AvailableDeviceListKey
func (cm *AdvanceDeviceFaultCm) GetAvailableDeviceListKey() string {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
		return ""
	}
	return api.ResourceNamePrefix + cm.DeviceType
}

// GetCmName get configmap name of device info
func (cm *DeviceInfo) GetCmName() string {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
		return ""
	}
	return cm.CmName
}

// GetCmName get configmap name of switch info
func (cm *SwitchInfo) GetCmName() string {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
		return ""
	}
	return cm.CmName
}

// GetCmName get configmap name of node info
func (cm *NodeInfo) GetCmName() string {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
		return ""
	}
	return cm.CmName
}

// IsSame compare with another cm
func (cm *DeviceInfo) IsSame(another ConfigMapInterface) bool {
	anotherDeviceInfo, ok := another.(*DeviceInfo)
	if !ok {
		hwlog.RunLog.Warnf("compare with cm which is not DeviceInfo")
		return false
	}
	return !DeviceInfoBusinessDataIsNotEqual(cm, anotherDeviceInfo)
}

// UpdateFaultReceiveTime update fault receive time
func (cm *DeviceInfo) UpdateFaultReceiveTime(oldInfo ConfigMapInterface) {
	return
}

// IsSame compare with another cm
func (cm *SwitchInfo) IsSame(another ConfigMapInterface) bool {
	anotherSwitchInfo, ok := another.(*SwitchInfo)
	if !ok {
		hwlog.RunLog.Warnf("compare with cm which is not SwitchInfo")
		return false
	}
	return !SwitchInfoBusinessDataIsNotEqual(cm, anotherSwitchInfo)
}

// UpdateFaultReceiveTime update fault receive time
func (cm *SwitchInfo) UpdateFaultReceiveTime(oldInfo ConfigMapInterface) {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
		return
	}
	oldCm, ok := oldInfo.(*SwitchInfo)
	if !ok || oldCm == nil {
		hwlog.RunLog.Error("oldInfo convert to SwitchInfo failed or oldCm is nil")
		updateFaultReceiveTimeForSwitchs(cm.FaultTimeAndLevelMap, nil)
		return
	}
	updateFaultReceiveTimeForSwitchs(cm.FaultTimeAndLevelMap, oldCm.FaultTimeAndLevelMap)
	hwlog.RunLog.Debugf("UpdateFaultReceiveTime cm.CmName=%s, cm.FaultTimeAndLevelMap=%#v",
		cm.CmName, cm.FaultTimeAndLevelMap)
}

func updateFaultReceiveTimeForSwitchs(cmFaultTimeAndLevelMap map[string]FaultTimeAndLevel,
	oldFaultTimeAndLevelMap map[string]FaultTimeAndLevel) {
	if cmFaultTimeAndLevelMap == nil {
		hwlog.RunLog.Error("cmFaultTimeAndLevelMap is nil")
		return
	}
	switchOldTimeMap := make(map[string]int64)
	if oldFaultTimeAndLevelMap != nil {
		for key, timeAndLevel := range oldFaultTimeAndLevelMap {
			switchOldTimeMap[key] = timeAndLevel.FaultReceivedTime
		}
	}
	hwlog.RunLog.Debugf("updateFaultReceiveTimeForSwitchs switchOldTimeMap=%#v", switchOldTimeMap)
	for key, timeAndLevel := range cmFaultTimeAndLevelMap {
		if oldTime, ok := switchOldTimeMap[key]; ok {
			timeAndLevel.FaultReceivedTime = oldTime
		} else {
			timeAndLevel.FaultReceivedTime = time.Now().UnixMilli()
		}
		cmFaultTimeAndLevelMap[key] = timeAndLevel
	}
}

// AddFaultAndFix add fault in the switchCM
func (cm *SwitchInfo) AddFaultAndFix(addFault SimpleSwitchFaultInfo) {
	cm.SwitchFaultInfo.FaultInfo = append(cm.FaultInfo, addFault)
}

// IsSame compare with another cm
func (cm *NodeInfo) IsSame(another ConfigMapInterface) bool {
	anotherNodeInfo, ok := another.(*NodeInfo)
	if !ok {
		hwlog.RunLog.Warnf("compare with cm which is not NodeInfo")
		return false
	}
	return !NodeInfoBusinessDataIsNotEqual(cm, anotherNodeInfo)
}

// UpdateFaultReceiveTime update fault receive time
func (cm *NodeInfo) UpdateFaultReceiveTime(oldInfo ConfigMapInterface) {
	return
}

// DeviceInfoBusinessDataIsNotEqual determine the business data is not equal
func DeviceInfoBusinessDataIsNotEqual(oldDevInfo *DeviceInfo, devInfo *DeviceInfo) bool {
	if oldDevInfo == nil && devInfo == nil {
		hwlog.RunLog.Debug("both oldDevInfo and devInfo are nil")
		return false
	}
	if oldDevInfo == nil || devInfo == nil {
		hwlog.RunLog.Debug("one of oldDevInfo and devInfo is not empty, and the other is empty")
		return true
	}
	if len(oldDevInfo.DeviceList) != len(devInfo.DeviceList) {
		hwlog.RunLog.Debug("the length of the deviceList of oldDevInfo is not equal to that of the deviceList of devInfo")
		return true
	}
	for nKey, nValue := range oldDevInfo.DeviceList {
		oValue, exists := devInfo.DeviceList[nKey]
		if !exists || nValue != oValue {
			hwlog.RunLog.Debug("neither oldDevInfo nor devInfo is empty, but oldDevInfo is not equal to devInfo")
			return true
		}
	}
	hwlog.RunLog.Debug("oldDevInfo is equal to devInfo")
	return false
}

// SwitchInfoBusinessDataIsNotEqual judge is the faultcode and fault level is the same as known, if is not same returns true
func SwitchInfoBusinessDataIsNotEqual(oldSwitch, newSwitch *SwitchInfo) bool {
	if oldSwitch == nil && newSwitch == nil {
		return false
	}
	if (oldSwitch != nil && newSwitch == nil) || (oldSwitch == nil && newSwitch != nil) {
		return true
	}
	if newSwitch.FaultLevel != oldSwitch.FaultLevel || newSwitch.NodeStatus != oldSwitch.NodeStatus ||
		len(newSwitch.FaultInfo) != len(oldSwitch.FaultInfo) {
		return true
	}
	hwlog.RunLog.Debug("oldSwitch is equal to newSwitch")
	return false
}

// NodeInfoBusinessDataIsNotEqual determine the business data is not equal
func NodeInfoBusinessDataIsNotEqual(oldNodeInfo *NodeInfo, newNodeInfo *NodeInfo) bool {
	if oldNodeInfo == nil && newNodeInfo == nil {
		hwlog.RunLog.Debug("both oldNodeInfo and newNodeInfo are nil")
		return false
	}
	if oldNodeInfo == nil || newNodeInfo == nil {
		hwlog.RunLog.Debug("one of oldNodeInfo and newNodeInfo is not empty, and the other is empty")
		return true
	}
	if oldNodeInfo.NodeStatus != newNodeInfo.NodeStatus ||
		len(oldNodeInfo.FaultDevList) != len(newNodeInfo.FaultDevList) {
		hwlog.RunLog.Debug("neither oldNodeInfo nor newNodeInfo is empty, but oldNodeInfo is not equal to newNodeInfo")
		return true
	}
	hwlog.RunLog.Debug("oldNodeInfo is equal to newNodeInfo")
	return false
}

func equalDeviceFault(one, other *DeviceFault) bool {
	return one.FaultType == other.FaultType &&
		one.NPUName == other.NPUName &&
		one.LargeModelFaultLevel == other.LargeModelFaultLevel &&
		one.FaultLevel == other.FaultLevel &&
		one.FaultHandling == other.FaultHandling &&
		one.FaultCode == other.FaultCode &&
		maps.Equal(one.FaultTimeAndLevelMap, other.FaultTimeAndLevelMap)
}
