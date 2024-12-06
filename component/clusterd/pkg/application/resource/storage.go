// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package resource a series of resource function
package resource

import (
	"sync"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/switchinfo"
)

var cmManager ConfigMapManager

// ConfigMapManager use for deviceInfo and nodeInfo report
type ConfigMapManager struct {
	sync.Mutex
	processCnt      int
	nodeInfoMap     map[string]*constant.NodeInfo
	nodeInfoMutex   sync.Mutex
	deviceInfoMap   map[string]*constant.DeviceInfo
	deviceInfoMutex sync.Mutex
	switchInfoMap   map[string]*constant.SwitchInfo
	switchInfoMutex sync.Mutex
}

func init() {
	cmManager.nodeInfoMap = map[string]*constant.NodeInfo{}
	cmManager.nodeInfoMutex = sync.Mutex{}
	cmManager.deviceInfoMap = map[string]*constant.DeviceInfo{}
	cmManager.deviceInfoMutex = sync.Mutex{}
	cmManager.switchInfoMap = map[string]*constant.SwitchInfo{}
	cmManager.switchInfoMutex = sync.Mutex{}
}

func delDeviceInfoCM(devInfo *constant.DeviceInfo) {
	cmManager.deviceInfoMutex.Lock()
	delete(cmManager.deviceInfoMap, devInfo.CmName)
	cmManager.deviceInfoMutex.Unlock()
	AddNewMessageTotal()
}

func delSwitchInfoCM(switchInfo *constant.SwitchInfo) {
	cmManager.switchInfoMutex.Lock()
	delete(cmManager.switchInfoMap, switchInfo.CmName)
	cmManager.switchInfoMutex.Unlock()
	AddNewMessageTotal()
}

func saveDeviceInfoCM(devInfo *constant.DeviceInfo) {
	cmManager.deviceInfoMutex.Lock()
	if len(cmManager.deviceInfoMap) > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("deviceInfoMap length=%d > %d, deviceInfo cm name=%s save failed",
			len(cmManager.deviceInfoMap), constant.MaxSupportNodeNum, devInfo.CmName)
		cmManager.deviceInfoMutex.Unlock()
		return
	}
	oldDevInfo := cmManager.deviceInfoMap[devInfo.CmName]
	cmManager.deviceInfoMap[devInfo.CmName] = devInfo
	cmManager.deviceInfoMutex.Unlock()
	// update business data will report message.if only update time，will report message with every atLeastReportCycle
	if device.BusinessDataIsNotEqual(oldDevInfo, devInfo) {
		AddNewMessageTotal()
	}
}

func saveSwitchInfoCM(newSwitchInfo *constant.SwitchInfo) {
	cmManager.switchInfoMutex.Lock()
	if len(cmManager.switchInfoMap) > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("switchInfoMap length=%d > %d, switchInfo cm name=%s save failed",
			len(cmManager.switchInfoMap), constant.MaxSupportNodeNum, newSwitchInfo.CmName)
		cmManager.switchInfoMutex.Unlock()
		return
	}
	oldSwitchInfo := cmManager.switchInfoMap[newSwitchInfo.CmName]
	cmManager.switchInfoMap[newSwitchInfo.CmName] = newSwitchInfo
	cmManager.switchInfoMutex.Unlock()
	if switchinfo.BusinessDataIsNotEqual(oldSwitchInfo, newSwitchInfo) {
		AddNewMessageTotal()
	}
}

// DeleteNodeConfigMap add CM to cache
func deleteNodeConfigMap(newDevInfo *constant.NodeInfo) {
	cmManager.nodeInfoMutex.Lock()
	delete(cmManager.nodeInfoMap, newDevInfo.CmName)
	cmManager.nodeInfoMutex.Unlock()
	AddNewMessageTotal()
}

func saveNodeInfoCM(newNodeInfo *constant.NodeInfo) {
	cmManager.nodeInfoMutex.Lock()
	if len(cmManager.nodeInfoMap) > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("nodeInfoMap length=%d > %d, nodeInfo cm name=%s save failed",
			len(cmManager.nodeInfoMap), constant.MaxSupportNodeNum, newNodeInfo.CmName)
		cmManager.nodeInfoMutex.Unlock()
		return
	}
	oldNodeInfo := cmManager.nodeInfoMap[newNodeInfo.CmName]
	cmManager.nodeInfoMap[newNodeInfo.CmName] = newNodeInfo
	cmManager.nodeInfoMutex.Unlock()
	// update business data will report message.if only update time, will report message with every 1atLeastReportCycle
	if node.BusinessDataIsNotEqual(oldNodeInfo, newNodeInfo) {
		AddNewMessageTotal()
	}
}
