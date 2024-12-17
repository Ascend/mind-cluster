// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package resource a series of resource function
package resource

import (
	"bytes"
	"encoding/gob"
	"strings"
	"sync"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/switchinfo"
	"clusterd/pkg/interface/kube"
)

var cmManager ConfigMapManager

// ConfigMapManager use for deviceInfo and nodeInfo report
type ConfigMapManager struct {
	sync.Mutex
	processCnt    int
	nodeInfoMap   map[string]*constant.NodeInfo
	deviceInfoMap map[string]*constant.DeviceInfo
	switchInfoMap map[string]*constant.SwitchInfo
}

// GetDeviceInfoMap get device info map
func GetDeviceInfoMap() map[string]*constant.DeviceInfo {
	tempDeviceInfoMap := new(map[string]*constant.DeviceInfo)
	cmManager.Lock()
	defer cmManager.Unlock()
	if err := DeepCopy(tempDeviceInfoMap, cmManager.deviceInfoMap); err != nil {
		hwlog.RunLog.Errorf("deep copy cmManager failed: %s", err)
		return *tempDeviceInfoMap
	}
	return *tempDeviceInfoMap
}

// GetSwitchInfoMap get switch info map
func GetSwitchInfoMap() map[string]*constant.SwitchInfo {
	tempSwitchInfoMap := new(map[string]*constant.SwitchInfo)
	cmManager.Lock()
	defer cmManager.Unlock()
	if err := DeepCopy(tempSwitchInfoMap, cmManager.switchInfoMap); err != nil {
		hwlog.RunLog.Errorf("deep copy cmManager failed: %s", err)
		return *tempSwitchInfoMap
	}
	return *tempSwitchInfoMap
}

func init() {
	cmManager.nodeInfoMap = map[string]*constant.NodeInfo{}
	cmManager.deviceInfoMap = map[string]*constant.DeviceInfo{}
	cmManager.switchInfoMap = map[string]*constant.SwitchInfo{}
}

// DeepCopy for object using gob
func DeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func delDeviceInfoCM(devInfo *constant.DeviceInfo) {
	cmManager.Lock()
	delete(cmManager.deviceInfoMap, devInfo.CmName)
	cmManager.Unlock()
	AddNewMessageTotal()
}

func delSwitchInfoCM(switchInfo *constant.SwitchInfo) {
	cmManager.Lock()
	delete(cmManager.switchInfoMap, switchInfo.CmName)
	cmManager.Unlock()
	AddNewMessageTotal()
}

func saveDeviceInfoCM(devInfo *constant.DeviceInfo) {
	cmManager.Lock()
	oldDevInfo := cmManager.deviceInfoMap[devInfo.CmName]
	cmManager.deviceInfoMap[devInfo.CmName] = devInfo
	cmManager.Unlock()
	// update business data will report message.if only update time，will report message with every atLeastReportCycle
	if device.BusinessDataIsNotEqual(oldDevInfo, devInfo) {
		if kube.JobMgr != nil {
			nodeName := strings.TrimPrefix(devInfo.CmName, constant.DeviceInfoPrefix)
			updateJobDeviceHealth(nodeName, devInfo.DeviceList)
		}
		AddNewMessageTotal()
	}
}

func updateJobDeviceHealth(nodeName string, deviceList map[string]string) {
	if kube.JobMgr == nil {
		hwlog.RunLog.Infof("jobMgr is nil, cannot set device healthy status on node: %s", nodeName)
		return
	}
	if len(deviceList) == 0 {
		hwlog.RunLog.Infof("device list is empty, ignore set device healthy status on node: %s", nodeName)
		return
	}
	netUnhealthy, unHealthy := "", ""
	for k, v := range deviceList {
		if strings.Contains(k, "NetworkUnhealthy") {
			netUnhealthy = v
		} else if strings.Contains(k, "Unhealthy") {
			unHealthy = v
		} else {
			continue
		}
	}
	kube.JobMgr.UpdateJobDeviceStatus(nodeName, netUnhealthy, unHealthy)
}

func saveSwitchInfoCM(newSwitchInfo *constant.SwitchInfo) {
	cmManager.Lock()
	oldSwitchInfo := cmManager.switchInfoMap[newSwitchInfo.CmName]
	cmManager.switchInfoMap[newSwitchInfo.CmName] = newSwitchInfo
	cmManager.Unlock()
	if switchinfo.BusinessDataIsNotEqual(oldSwitchInfo, newSwitchInfo) {
		if kube.JobMgr != nil {
			nodeName := strings.TrimPrefix(newSwitchInfo.CmName, constant.SwitchInfoPrefix)
			updateJobNodeHealth(nodeName, newSwitchInfo.NodeStatus == "Healthy")
		}
		AddNewMessageTotal()
	}
}

// DeleteNodeConfigMap add CM to cache
func deleteNodeConfigMap(newDevInfo *constant.NodeInfo) {
	cmManager.Lock()
	delete(cmManager.nodeInfoMap, newDevInfo.CmName)
	cmManager.Unlock()
	AddNewMessageTotal()
}

func saveNodeInfoCM(newNodeInfo *constant.NodeInfo) {
	cmManager.Lock()
	oldNodeInfo := cmManager.nodeInfoMap[newNodeInfo.CmName]
	cmManager.nodeInfoMap[newNodeInfo.CmName] = newNodeInfo
	cmManager.Unlock()
	// update business data will report message.if only update time, will report message with every 1atLeastReportCycle
	if node.BusinessDataIsNotEqual(oldNodeInfo, newNodeInfo) {
		if kube.JobMgr != nil {
			nodeName := strings.TrimPrefix(newNodeInfo.CmName, constant.NodeInfoPrefix)
			updateJobNodeHealth(nodeName, newNodeInfo.NodeStatus == "Healthy")
		}
		AddNewMessageTotal()
	}
}

func updateJobNodeHealth(nodeName string, healthy bool) {
	if kube.JobMgr == nil {
		hwlog.RunLog.Infof("jobMgr is nil, cannot set node healthy status on node: %s", nodeName)
		return
	}
	kube.JobMgr.UpdateJobNodeStatus(nodeName, healthy)
}
