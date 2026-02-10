// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package custom cache utils for fault filtered by custom_filter_fault_processor
package custom

import (
	"sync"

	"clusterd/pkg/common/constant"
)

type faultCache struct {
	deletedDevFaultCm        map[string]*constant.AdvanceDeviceFaultCm
	deletedSwitchFaultCm     map[string]*constant.SwitchInfo
	deletedJobFaultDeviceMap map[string][]constant.FaultDevice
	rwMutex                  sync.RWMutex
}

// FaultCache custom fault cache
var FaultCache *faultCache

func init() {
	FaultCache = &faultCache{
		deletedDevFaultCm:        make(map[string]*constant.AdvanceDeviceFaultCm),
		deletedSwitchFaultCm:     make(map[string]*constant.SwitchInfo),
		deletedJobFaultDeviceMap: make(map[string][]constant.FaultDevice),
		rwMutex:                  sync.RWMutex{},
	}
}

// GetDeletedDevFaultCmForNodeMap get deleted dev fault cm for node map
func (cache *faultCache) GetDeletedDevFaultCmForNodeMap() map[string]*constant.AdvanceDeviceFaultCm {
	cache.rwMutex.RLock()
	defer cache.rwMutex.RUnlock()
	return cache.deletedDevFaultCm
}

// GetDeletedSwitchFaultCmForNodeMap get deleted switch fault cm for node map
func (cache *faultCache) GetDeletedSwitchFaultCmForNodeMap() map[string]*constant.SwitchInfo {
	cache.rwMutex.RLock()
	defer cache.rwMutex.RUnlock()
	return cache.deletedSwitchFaultCm
}

// GetDeletedJobFaultDeviceMap get deleted job fault device map
func (cache *faultCache) GetDeletedJobFaultDeviceMap() map[string][]constant.FaultDevice {
	cache.rwMutex.RLock()
	defer cache.rwMutex.RUnlock()
	return cache.deletedJobFaultDeviceMap
}

// SetDeletedDevFaultCmForNodeMap set deleted dev fault cm for node map
func (cache *faultCache) SetDeletedDevFaultCmForNodeMap(
	deletedDeviceFaultCmMap map[string]*constant.AdvanceDeviceFaultCm) {
	cache.rwMutex.Lock()
	defer cache.rwMutex.Unlock()
	cache.deletedDevFaultCm = deletedDeviceFaultCmMap
}

// SetDeletedSwitchFaultCmForNodeMap set deleted switch fault cm for node map
func (cache *faultCache) SetDeletedSwitchFaultCmForNodeMap(
	deletedSwitchFaultCmMap map[string]*constant.SwitchInfo) {
	cache.rwMutex.Lock()
	defer cache.rwMutex.Unlock()
	cache.deletedSwitchFaultCm = deletedSwitchFaultCmMap
}

// SetDeletedJobFaultDeviceMap set deleted job fault device map
func (cache *faultCache) SetDeletedJobFaultDeviceMap(deletedJobFaultDeviceMap map[string][]constant.FaultDevice) {
	cache.rwMutex.Lock()
	defer cache.rwMutex.Unlock()
	cache.deletedJobFaultDeviceMap = deletedJobFaultDeviceMap
}
