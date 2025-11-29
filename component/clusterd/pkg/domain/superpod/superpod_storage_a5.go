/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package superpod a series of cluster device info storage function.
*/
package superpod

import (
	"fmt"
	"strconv"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

var superPodDeleteFinalManager Manager

const (
	// maxRackNumPerSuperPod for a5 ras
	maxRackNumPerSuperPod = 256
	// maxNodeNumPerSuperPodA5 for a5 ras
	maxNodeNumPerSuperPodA5 = 1024
)

// GetAllSuperPodIDWithAcceleratorType return all super pod ids with accelerator-type
func GetAllSuperPodIDWithAcceleratorType() map[int]string {
	superPodManager.rwLock.RLock()
	defer superPodManager.rwLock.RUnlock()
	result := make(map[int]string)
	for id, device := range superPodManager.snMap {
		if device == nil || device.SuperPodID == "" {
			continue
		}
		// exclude repeat one
		if intId, err := strconv.Atoi(id); err == nil && result[intId] == "" {
			result[intId] = device.AcceleratorType
		}
	}
	return result
}

func deepCopyNpuInfo(npuInfo *api.NpuInfo) *api.NpuInfo {
	if npuInfo == nil {
		return nil
	}
	copyNpuInfo := &api.NpuInfo{
		PhyId:     npuInfo.PhyId,
		LevelList: make([]api.LevelElement, len(npuInfo.LevelList)),
	}
	for _, each := range npuInfo.LevelList {
		copyNpuInfo.LevelList = append(copyNpuInfo.LevelList, each)
	}
	return copyNpuInfo
}

func deepCopyServerInfo(serverInfo *api.ServerInfo) *api.ServerInfo {
	if serverInfo == nil {
		return nil
	}
	copyServerInfo := &api.ServerInfo{
		ServerIndex: serverInfo.ServerIndex,
		NodeName:    serverInfo.NodeName,
		NpuMap:      make(map[string]*api.NpuInfo, len(serverInfo.NpuMap)),
	}
	for k, v := range serverInfo.NpuMap {
		copyServerInfo.NpuMap[k] = deepCopyNpuInfo(v)
	}
	return copyServerInfo
}

func deepCopyRackInfo(rackInfo *api.RackInfo) *api.RackInfo {
	if rackInfo == nil {
		return nil
	}
	copyRackInfo := &api.RackInfo{
		RackID:    rackInfo.RackID,
		ServerMap: make(map[string]*api.ServerInfo, len(rackInfo.ServerMap)),
	}
	for k, v := range rackInfo.ServerMap {
		copyRackInfo.ServerMap[k] = deepCopyServerInfo(v)
	}
	return copyRackInfo
}

// GetFinalDelSuperPodID return delete superpod last node info
func GetFinalDelSuperPodID() map[int]string {
	superPodDeleteFinalManager.rwLock.RLock()
	defer superPodDeleteFinalManager.rwLock.RUnlock()
	result := make(map[int]string)
	for id, device := range superPodDeleteFinalManager.snMap {
		if device == nil || device.SuperPodID == "" {
			continue
		}
		if intId, err := strconv.Atoi(id); err == nil && result[intId] == "" {
			result[intId] = device.AcceleratorType
		}
	}
	return result
}

func saveRackInfo(superPod *api.SuperPodDevice, node *api.NodeDevice) error {
	if !api.CheckIsVersionA5(superPod.Version) {
		return nil
	}
	rackID := node.RackID
	rackInfo, ok := superPod.RackMap[rackID]
	if !ok {
		if len(superPod.RackMap) >= maxRackNumPerSuperPod {
			hwlog.RunLog.Errorf("rackMap length will exceed %d, superPodID=%s, rackID=%s, nodeName=%s",
				maxRackNumPerSuperPod, superPod.SuperPodID, rackID, node.NodeName)
			return fmt.Errorf("rackMap length exceeds the limit")
		}
		rackInfo = &api.RackInfo{
			RackID:    rackID,
			ServerMap: make(map[string]*api.ServerInfo),
		}
		superPod.RackMap[rackID] = rackInfo
	}

	serverInfo := &api.ServerInfo{
		NodeName:    node.NodeName,
		ServerIndex: node.ServerID,
		NpuMap:      make(map[string]*api.NpuInfo),
	}
	for phyId, npuInfo := range node.NpuInfoMap {
		serverInfo.NpuMap[phyId] = npuInfo
	}
	rackInfo.ServerMap[serverInfo.ServerIndex] = serverInfo
	return nil
}

func canAddNodeToSuperPod(superPod *api.SuperPodDevice, node *api.NodeDevice) bool {
	if len(superPod.Version) != 0 && node.ServerType != superPod.Version {
		hwlog.RunLog.Errorf("node version (%s) is not same as super pod version (%s), node name: %s",
			node.ServerType, superPod.Version, node.NodeName)
		return false
	}
	if api.CheckIsVersionA5(superPod.Version) && len(superPod.NodeDeviceMap) > maxNodeNumPerSuperPodA5 {
		hwlog.RunLog.Errorf("nodeDeviceMap length will exceed %d, superPodID=%s, nodeName=%s",
			maxNodeNumPerSuperPod, superPod.SuperPodID, node.NodeName)
		return false
	}
	if !api.CheckIsVersionA5(superPod.Version) && len(superPod.NodeDeviceMap) > maxNodeNumPerSuperPod {
		hwlog.RunLog.Errorf("nodeDeviceMap length will exceed %d, superPodID=%s, nodeName=%s",
			maxNodeNumPerSuperPod, superPod.SuperPodID, node.NodeName)
		return false
	}
	return true
}

// DeleteNodeInRackMap delete node from RackMap in superPodManager
func DeleteNodeInRackMap(superPodID string, nodeDevice *api.NodeDevice) {
	superPodManager.rwLock.Lock()
	defer superPodManager.rwLock.Unlock()
	superPod, ok := superPodManager.snMap[superPodID]
	if !ok {
		hwlog.RunLog.Warnf("DeleteNodeInRackMap failed, superPodID<%s> not exit in snMap", superPodID)
		return
	}

	if nodeDevice == nil {
		hwlog.RunLog.Warn("DeleteNodeInRackMap failed, nodeDevice is nil")
		return
	}

	rackID := nodeDevice.RackID
	rack, ok := superPod.RackMap[rackID]
	if !ok {
		hwlog.RunLog.Warnf("DeleteNodeInRackMap failed, superPodID<%s> rackID<%s> not exit in rackMap",
			superPodID, rackID)
		return
	}

	serverID := nodeDevice.ServerID
	delete(rack.ServerMap, serverID)
	if len(rack.ServerMap) == 0 {
		delete(superPod.RackMap, rackID)
	}
	if len(superPod.RackMap) == 0 {
		delete(superPodManager.snMap, superPodID)
	}
	hwlog.RunLog.Warnf("DeleteNodeInRackMap success, superPodID<%s> rackID<%s> serverID<%s>",
		superPodID, rackID, serverID)
}
