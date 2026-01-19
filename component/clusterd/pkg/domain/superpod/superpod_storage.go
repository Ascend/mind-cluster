// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package superpod a series of cluster device info storage function
package superpod

import (
	"sync"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

const (
	maxNodeNumPerSuperPod  = 256
	maxSuperPodNum         = 1024
	initNodeNumPerSuperPod = 64
	initSuperPodNum        = 32
)

func deepCopyNodeDevice(device *api.NodeDevice) *api.NodeDevice {
	if device == nil {
		return nil
	}
	copyDevice := &api.NodeDevice{
		ServerType: device.ServerType,
		ServerID:   device.ServerID,
		NodeName:   device.NodeName,
		DeviceMap:  make(map[string]string, len(device.DeviceMap)),
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
		Version:       superPodDevice.Version,
		SuperPodID:    superPodDevice.SuperPodID,
		NodeDeviceMap: make(map[string]*api.NodeDevice, len(superPodDevice.NodeDeviceMap)),
	}
	for k, v := range superPodDevice.NodeDeviceMap {
		copySuperPodDevice.NodeDeviceMap[k] = deepCopyNodeDevice(v)
	}
	return copySuperPodDevice
}

// Manager the manager of super pod
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
func GetSuperPodDevice(superPodID string) *api.SuperPodDevice {
	superPodManager.rwLock.RLock()
	defer superPodManager.rwLock.RUnlock()
	superPod, ok := superPodManager.snMap[superPodID]
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
		hwlog.RunLog.Debugf("reject add node device with empty superPodID, nodeName=%s",
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
			Version:       node.ServerType,
			SuperPodID:    superPodID,
			NodeDeviceMap: make(map[string]*api.NodeDevice, initNodeNumPerSuperPod),
		}
		superPodManager.snMap[superPodID] = superPod
	}
	if len(superPod.NodeDeviceMap) >= maxNodeNumPerSuperPod {
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
