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
	"context"
	"testing"

	"Ascend-device-plugin/pkg/duplicatedetector/types"
	"ascend-common/common-utils/hwlog"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func TestNewContainerCache(t *testing.T) {
	cache := NewContainerCache()
	if cache == nil {
		t.Fatal("NewContainerCache returned nil")
	}
	if cache.containers == nil {
		t.Error("containers map is nil")
	}
	if cache.deviceMap == nil {
		t.Error("deviceMap is nil")
	}
}

func TestStoreAllAndFindDuplicates_NoDuplicates(t *testing.T) {
	cache := NewContainerCache()
	infos := map[string]*types.ContainerNPUInfo{
		"container1": {ID: "container1", Devices: []int{0}},
		"container2": {ID: "container2", Devices: []int{1}},
	}
	duplicates := cache.StoreAllAndFindDuplicates(infos)
	if len(duplicates) != 0 {
		t.Errorf("expected 0 duplicates, got %d", len(duplicates))
	}
}

func TestStoreAllAndFindDuplicates_WithDuplicates(t *testing.T) {
	cache := NewContainerCache()
	infos := map[string]*types.ContainerNPUInfo{
		"container1": {ID: "container1", Devices: []int{0}},
		"container2": {ID: "container2", Devices: []int{0}},
	}
	duplicates := cache.StoreAllAndFindDuplicates(infos)
	if len(duplicates) != 1 {
		t.Errorf("expected 1 duplicate, got %d", len(duplicates))
	}
	if duplicates[0].DeviceID != 0 {
		t.Errorf("expected device ID 0, got %d", duplicates[0].DeviceID)
	}
}

func TestStoreAllAndFindDuplicates_Empty(t *testing.T) {
	cache := NewContainerCache()
	infos := map[string]*types.ContainerNPUInfo{}
	duplicates := cache.StoreAllAndFindDuplicates(infos)
	if len(duplicates) != 0 {
		t.Errorf("expected 0 duplicates, got %d", len(duplicates))
	}
}

func TestStoreSingleAndFindDuplicates_NoDuplicates(t *testing.T) {
	cache := NewContainerCache()
	info := &types.ContainerNPUInfo{ID: "container1", Devices: []int{0}}
	duplicates := cache.StoreSingleAndFindDuplicates(info)
	if len(duplicates) != 0 {
		t.Errorf("expected 0 duplicates, got %d", len(duplicates))
	}
}

func TestStoreSingleAndFindDuplicates_WithDuplicates(t *testing.T) {
	cache := NewContainerCache()
	info1 := &types.ContainerNPUInfo{ID: "container1", Devices: []int{0}}
	cache.StoreSingleAndFindDuplicates(info1)

	info2 := &types.ContainerNPUInfo{ID: "container2", Devices: []int{0}}
	duplicates := cache.StoreSingleAndFindDuplicates(info2)
	if len(duplicates) != 1 {
		t.Errorf("expected 1 duplicate, got %d", len(duplicates))
	}
	if duplicates[0].DeviceID != 0 {
		t.Errorf("expected device ID 0, got %d", duplicates[0].DeviceID)
	}
}

func TestStoreSingleAndFindDuplicates_MultipleDevices(t *testing.T) {
	cache := NewContainerCache()
	info1 := &types.ContainerNPUInfo{ID: "container1", Devices: []int{0, 1}}
	cache.StoreSingleAndFindDuplicates(info1)

	info2 := &types.ContainerNPUInfo{ID: "container2", Devices: []int{1, 2}}
	duplicates := cache.StoreSingleAndFindDuplicates(info2)
	if len(duplicates) != 1 {
		t.Errorf("expected 1 duplicate, got %d", len(duplicates))
	}
	if duplicates[0].DeviceID != 1 {
		t.Errorf("expected device ID 1, got %d", duplicates[0].DeviceID)
	}
}

func TestRemoveContainer_Existing(t *testing.T) {
	cache := NewContainerCache()
	info := &types.ContainerNPUInfo{ID: "container1", Devices: []int{0}}
	cache.StoreSingleAndFindDuplicates(info)

	cache.RemoveContainer("container1")
	if _, exists := cache.containers["container1"]; exists {
		t.Error("container still exists after removal")
	}
	if len(cache.deviceMap[0]) != 0 {
		t.Error("device mapping not cleaned up")
	}
}

func TestRemoveContainer_NonExisting(t *testing.T) {
	cache := NewContainerCache()
	cache.RemoveContainer("nonexistent")
}

func TestRemoveContainer_MultipleDevices(t *testing.T) {
	cache := NewContainerCache()
	info := &types.ContainerNPUInfo{ID: "container1", Devices: []int{0, 1, 2}}
	cache.StoreSingleAndFindDuplicates(info)

	cache.RemoveContainer("container1")
	for _, deviceID := range []int{0, 1, 2} {
		if len(cache.deviceMap[deviceID]) != 0 {
			t.Errorf("device %d mapping not cleaned up", deviceID)
		}
	}
}

func TestRemoveContainer_WithOtherContainers(t *testing.T) {
	cache := NewContainerCache()
	info1 := &types.ContainerNPUInfo{ID: "container1", Devices: []int{0}}
	info2 := &types.ContainerNPUInfo{ID: "container2", Devices: []int{0}}
	cache.StoreSingleAndFindDuplicates(info1)
	cache.StoreSingleAndFindDuplicates(info2)

	cache.RemoveContainer("container1")
	if len(cache.deviceMap[0]) != 1 {
		t.Errorf("expected 1 container for device 0, got %d", len(cache.deviceMap[0]))
	}
	if cache.deviceMap[0][0] != "container2" {
		t.Error("wrong container remains in device mapping")
	}
}
