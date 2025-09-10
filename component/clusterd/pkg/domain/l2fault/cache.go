// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package l2fault cache utils for l2 fault
package l2fault

import (
	"sync"

	"clusterd/pkg/common/constant"
)

type l2FaultCache struct {
	deletedDevL2FaultCm      map[string]*constant.AdvanceDeviceFaultCm
	deletedSwitchL2FaultCm   map[string]*constant.SwitchInfo
	deletedJobFaultDeviceMap map[string][]constant.FaultDevice
	rwMutex                  sync.RWMutex
}

// L2FaultCache l2 fault cache
var L2FaultCache *l2FaultCache

func init() {
	L2FaultCache = &l2FaultCache{
		deletedDevL2FaultCm:      make(map[string]*constant.AdvanceDeviceFaultCm),
		deletedSwitchL2FaultCm:   make(map[string]*constant.SwitchInfo),
		deletedJobFaultDeviceMap: make(map[string][]constant.FaultDevice),
		rwMutex:                  sync.RWMutex{},
	}
}

// GetDeletedDevL2FaultCmForNodeMap get deleted dev l2 fault cm for node map
func (cache *l2FaultCache) GetDeletedDevL2FaultCmForNodeMap() map[string]*constant.AdvanceDeviceFaultCm {
	cache.rwMutex.RLock()
	defer cache.rwMutex.RUnlock()
	return cache.deletedDevL2FaultCm
}

// GetDeletedSwitchL2FaultCmForNodeMap get deleted switch l2 fault cm for node map
func (cache *l2FaultCache) GetDeletedSwitchL2FaultCmForNodeMap() map[string]*constant.SwitchInfo {
	cache.rwMutex.RLock()
	defer cache.rwMutex.RUnlock()
	return cache.deletedSwitchL2FaultCm
}

// GetDeletedJobFaultDeviceMap get deleted job fault device map
func (cache *l2FaultCache) GetDeletedJobFaultDeviceMap() map[string][]constant.FaultDevice {
	cache.rwMutex.RLock()
	defer cache.rwMutex.RUnlock()
	return cache.deletedJobFaultDeviceMap
}

// SetDeletedDevL2FaultCmForNodeMap set deleted dev l2 fault cm for node map
func (cache *l2FaultCache) SetDeletedDevL2FaultCmForNodeMap(
	deletedDeviceFaultCmMap map[string]*constant.AdvanceDeviceFaultCm) {
	cache.rwMutex.Lock()
	defer cache.rwMutex.Unlock()
	cache.deletedDevL2FaultCm = deletedDeviceFaultCmMap
}

// SetDeletedSwitchL2FaultCmForNodeMap set deleted switch l2 fault cm for node map
func (cache *l2FaultCache) SetDeletedSwitchL2FaultCmForNodeMap(
	deletedSwitchFaultCmMap map[string]*constant.SwitchInfo) {
	cache.rwMutex.Lock()
	defer cache.rwMutex.Unlock()
	cache.deletedSwitchL2FaultCm = deletedSwitchFaultCmMap
}

// SetDeletedJobFaultDeviceMap set deleted job fault device map
func (cache *l2FaultCache) SetDeletedJobFaultDeviceMap(deletedJobFaultDeviceMap map[string][]constant.FaultDevice) {
	cache.rwMutex.Lock()
	defer cache.rwMutex.Unlock()
	cache.deletedJobFaultDeviceMap = deletedJobFaultDeviceMap
}
