/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package cache provides cache for container NPU device information

package cache

import (
	"sync"

	"Ascend-device-plugin/pkg/duplicatedetector/types"
	"ascend-common/common-utils/hwlog"
)

// ContainerCache caches container NPU device information
type ContainerCache struct {
	// ContainerID -> ContainerNPUInfo
	containers map[string]*types.ContainerNPUInfo
	// DeviceID -> list of container IDs
	deviceMap map[int][]string
	mutex     sync.RWMutex
}

// NewContainerCache creates a new ContainerCache
func NewContainerCache() *ContainerCache {
	return &ContainerCache{
		containers: make(map[string]*types.ContainerNPUInfo),
		deviceMap:  make(map[int][]string),
	}
}

// StoreAllAndFindDuplicates stores all container NPU device information and finds duplicates
func (cc *ContainerCache) StoreAllAndFindDuplicates(infos map[string]*types.ContainerNPUInfo) []*types.DuplicateMountInfo {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	cc.containers = infos
	for _, info := range infos {
		for _, deviceID := range info.Devices {
			cc.deviceMap[deviceID] = append(cc.deviceMap[deviceID], info.ID)
		}
	}
	return cc.findDuplicates()
}

// StoreSingleAndFindDuplicates stores a single container NPU device information and finds duplicates
func (cc *ContainerCache) StoreSingleAndFindDuplicates(info *types.ContainerNPUInfo) []*types.DuplicateMountInfo {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	// Check for duplicates with existing containers
	var duplicates []*types.DuplicateMountInfo
	for _, deviceID := range info.Devices {
		existingContainers, ok := cc.deviceMap[deviceID]
		if !ok || len(existingContainers) == 0 {
			continue
		}
		// This device is already mounted by other containers
		var dupContainers []*types.ContainerNPUInfo
		for _, existingID := range existingContainers {
			if existingContainer, ok := cc.containers[existingID]; ok {
				dupContainers = append(dupContainers, existingContainer)
			}
		}
		dupContainers = append(dupContainers, info)
		duplicates = append(duplicates, &types.DuplicateMountInfo{
			DeviceID:   deviceID,
			Containers: dupContainers,
		})

	}

	// Add new container to cache
	cc.containers[info.ID] = info
	for _, deviceID := range info.Devices {
		cc.deviceMap[deviceID] = append(cc.deviceMap[deviceID], info.ID)
	}
	return duplicates
}

// RemoveContainer removes a container from the cache
func (cc *ContainerCache) RemoveContainer(containerID string) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	info, ok := cc.containers[containerID]
	if !ok {
		return
	}

	// Remove container from cache
	delete(cc.containers, containerID)

	// Remove container from device mappings
	for _, deviceID := range info.Devices {
		containers := cc.deviceMap[deviceID]
		newList := make([]string, 0, len(containers))
		for _, id := range containers {
			if id != containerID {
				newList = append(newList, id)
			}
		}
		cc.deviceMap[deviceID] = newList
	}
}

// findDuplicates finds duplicate mounts
func (cc *ContainerCache) findDuplicates() []*types.DuplicateMountInfo {
	var duplicates []*types.DuplicateMountInfo
	for deviceID, containerIDs := range cc.deviceMap {
		hwlog.RunLog.Infof("checking device %d, containers: %d", deviceID, len(containerIDs))
		if len(containerIDs) == 1 {
			continue
		}
		containers := make([]*types.ContainerNPUInfo, 0, len(containerIDs))
		for _, id := range containerIDs {
			if info, ok := cc.containers[id]; ok {
				containers = append(containers, info)
			}
		}

		duplicates = append(duplicates, &types.DuplicateMountInfo{
			DeviceID:   deviceID,
			Containers: containers,
		})

	}
	return duplicates
}
