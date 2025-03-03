// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package superpod a series of cluster device info storage function
package superpod

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

const (
	maxNodeNumPerSuperPod  = 256
	maxSuperPodNum         = 1024
	initNodeNumPerSuperPod = 64
	initSuperPodNum        = 32
	maxNodeDeviceNum       = 128
	formatBase             = 10

	deviceKey     = "baseDeviceInfos"
	superPodIDKey = "superPodID"
)

func deepCopyNodeDevice(device *api.NodeDevice) *api.NodeDevice {
	if device == nil {
		return nil
	}
	copyDevice := &api.NodeDevice{
		NodeName:  device.NodeName,
		DeviceMap: make(map[string]string, len(device.DeviceMap)),
	}
	for k, v := range device.DeviceMap {
		copyDevice.DeviceMap[k] = v
	}
	return copyDevice
}

func deepCopySuperPodDevice(superPodDevice *api.SuperPodDevice) *api.SuperPodDevice {
	if superPodDevice == nil {
		return nil
	}
	copySuperPodDevice := &api.SuperPodDevice{
		SuperPodID:    superPodDevice.SuperPodID,
		NodeDeviceMap: make(map[string]*api.NodeDevice, len(superPodDevice.NodeDeviceMap)),
	}
	for k, v := range superPodDevice.NodeDeviceMap {
		copySuperPodDevice.NodeDeviceMap[k] = deepCopyNodeDevice(v)
	}
	return copySuperPodDevice
}

type Manager struct {
	snMap  map[string]*api.SuperPodDevice
	rwLock sync.RWMutex
}

var superPodManager Manager

func init() {
	superPodManager.snMap = make(map[string]*api.SuperPodDevice, initSuperPodNum)
	superPodManager.rwLock = sync.RWMutex{}
}

// GetSuperPodDevice get superPod with lock
func GetSuperPodDevice(jobKey string) *api.SuperPodDevice {
	superPodManager.rwLock.RLock()
	defer superPodManager.rwLock.RUnlock()
	superPod, ok := superPodManager.snMap[jobKey]
	if !ok {
		return nil
	}
	return deepCopySuperPodDevice(superPod)
}

// SaveNode save node with lock
func SaveNode(superPodID string, node *api.NodeDevice) {
	if node == nil {
		hwlog.RunLog.Warn("reject add nil node device")
		return
	}
	if len(superPodID) == 0 {
		hwlog.RunLog.Warnf("reject add node device with empty superPodID, nodeName=%s",
			node.NodeName)
		return
	}
	superPodManager.rwLock.Lock()
	defer superPodManager.rwLock.Unlock()
	superPod, ok := superPodManager.snMap[superPodID]
	if !ok {
		if len(superPodManager.snMap) >= maxSuperPodNum {
			hwlog.RunLog.Errorf("snMap length will exceed %d, superPodID=%s, nodeName=%s",
				maxSuperPodNum, superPodID, node.NodeName)
			return
		}
		superPod = &api.SuperPodDevice{
			SuperPodID:    superPodID,
			NodeDeviceMap: make(map[string]*api.NodeDevice, initNodeNumPerSuperPod),
		}
		superPodManager.snMap[superPodID] = superPod
	}
	if len(superPod.NodeDeviceMap) > maxNodeNumPerSuperPod {
		hwlog.RunLog.Errorf("nodeDeviceMap length will exceed %d, superPodID=%s, nodeName=%s",
			maxNodeNumPerSuperPod, superPodID, node.NodeName)
		return
	}
	superPod.NodeDeviceMap[node.NodeName] = node
}

// DeleteNode delete node with lock
func DeleteNode(superPodID string, nodeName string) {
	superPodManager.rwLock.Lock()
	defer superPodManager.rwLock.Unlock()
	superPod, ok := superPodManager.snMap[superPodID]
	if !ok {
		return
	}
	delete(superPod.NodeDeviceMap, nodeName)
	if len(superPod.NodeDeviceMap) == 0 {
		delete(superPodManager.snMap, superPodID)
	}
	return
}

// GetNodeDeviceAndSuperPodID parse NodeDevice and superPodID from node
func GetNodeDeviceAndSuperPodID(node *v1.Node) (*api.NodeDevice, string) {
	if node == nil {
		hwlog.RunLog.Error("empty node")
		return nil, ""
	}
	if len(node.Name) == 0 {
		hwlog.RunLog.Error("empty node name")
		return nil, ""
	}
	superPodID, hasSuperPodIDKey := node.Annotations[superPodIDKey]
	if !hasSuperPodIDKey || len(superPodID) == 0 {
		hwlog.RunLog.Errorf("empty super pod id, nodeName=%s", node.Name)
		return nil, ""
	}
	baseDeviceMap := make(map[string]*api.NpuBaseInfo)
	deviceStr, hasDeviceKey := node.Annotations[deviceKey]
	if !hasDeviceKey || len(deviceStr) == 0 {
		hwlog.RunLog.Errorf("empty device info, nodeName=%s", node.Name)
		return nil, superPodID
	}
	if err := json.Unmarshal([]byte(deviceStr), &baseDeviceMap); err != nil {
		hwlog.RunLog.Errorf("unmarshal device info error, err=%v, nodeName=%s",
			err, node.Name)
		return nil, superPodID
	}
	if len(baseDeviceMap) == 0 || len(baseDeviceMap) > maxNodeDeviceNum {
		hwlog.RunLog.Errorf("illegal device length, deviceLen=%d, nodeName=%s",
			len(baseDeviceMap), node.Name)
		return nil, superPodID
	}
	nodeDevice := &api.NodeDevice{
		NodeName:  node.Name,
		DeviceMap: make(map[string]string, len(baseDeviceMap)),
	}
	for device, info := range baseDeviceMap {
		physicID := strings.TrimPrefix(device, constant.AscendDevPrefix)
		_, err := strconv.Atoi(physicID)
		if err != nil {
			hwlog.RunLog.Errorf("illegal device name, deviceName=%s, nodeName=%s",
				device, node.Name)
			return nil, superPodID
		}
		superDeviceID := strconv.FormatUint(uint64(info.SuperDeviceID), formatBase)
		nodeDevice.DeviceMap[physicID] = superDeviceID
	}
	return nodeDevice, superPodID
}

// ListClusterDevice return slice of cluster super pod device
func ListClusterDevice() []*api.SuperPodDevice {
	superPodManager.rwLock.Lock()
	defer superPodManager.rwLock.Unlock()
	superPodSlice := make([]*api.SuperPodDevice, 0, len(superPodManager.snMap))
	for _, device := range superPodManager.snMap {
		superPodSlice = append(superPodSlice, deepCopySuperPodDevice(device))
	}
	return superPodSlice
}
