// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package constant a series of para
package constant

import (
	"maps"

	"k8s.io/utils/strings/slices"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

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

// GetCmName return cm name
func (cm *AdvanceDeviceFaultCm) GetCmName() string {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
	}
	return cm.CmName
}

// GetRecoveringKey return cm RecoveringKey
func (cm *AdvanceDeviceFaultCm) GetRecoveringKey() string {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
	}
	return api.ResourceNamePrefix + cm.DeviceType + CmRecoveringSuffix
}

// GetCardUnHealthyKey return cm CardUnHealthyKey
func (cm *AdvanceDeviceFaultCm) GetCardUnHealthyKey() string {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
	}
	return api.ResourceNamePrefix + cm.DeviceType + CmCardUnhealthySuffix
}

// GetNetworkUnhealthyKey return cm NetworkUnhealthyKey
func (cm *AdvanceDeviceFaultCm) GetNetworkUnhealthyKey() string {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
	}
	return api.ResourceNamePrefix + cm.DeviceType + CmCardNetworkUnhealthySuffix
}

// GetFaultDeviceListKey return cm FaultDeviceListKey
func (cm *AdvanceDeviceFaultCm) GetFaultDeviceListKey() string {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
	}
	return api.ResourceNamePrefix + cm.DeviceType + CmFaultListSuffix
}

// GetAvailableDeviceListKey return cm AvailableDeviceListKey
func (cm *AdvanceDeviceFaultCm) GetAvailableDeviceListKey() string {
	if cm == nil {
		hwlog.RunLog.Error("cm is nil")
	}
	return api.ResourceNamePrefix + cm.DeviceType
}

// GetCmName get configmap name of device info
func (cm *DeviceInfo) GetCmName() string {
	return cm.CmName
}

// GetCmName get configmap name of switch info
func (cm *SwitchInfo) GetCmName() string {
	return cm.CmName
}

// GetCmName get configmap name of node info
func (cm *NodeInfo) GetCmName() string {
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

// IsSame compare with another cm
func (cm *SwitchInfo) IsSame(another ConfigMapInterface) bool {
	anotherSwitchInfo, ok := another.(*SwitchInfo)
	if !ok {
		hwlog.RunLog.Warnf("compare with cm which is not SwitchInfo")
		return false
	}
	return !SwitchInfoBusinessDataIsNotEqual(cm, anotherSwitchInfo)
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
		len(newSwitch.FaultCode) != len(oldSwitch.FaultCode) {
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
