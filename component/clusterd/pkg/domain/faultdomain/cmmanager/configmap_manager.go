// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package faultmanager contain fault process
package cmmanager

import (
	"sync"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain/collector"
)

type ConfigMap[T constant.ConfigMapInterface] struct {
	Data map[string]T
}

var DeviceCenterCmManager *FaultCenterCmManager[*constant.AdvanceDeviceFaultCm]
var SwitchCenterCmManager *FaultCenterCmManager[*constant.SwitchInfo]
var NodeCenterCmManager *FaultCenterCmManager[*constant.NodeInfo]
var DpuCenterCMManager *FaultCenterCmManager[*constant.DpuInfoCM]

type FaultCenterCmManager[T constant.ConfigMapInterface] struct {
	mutex       sync.RWMutex
	cmBuffer    *collector.ConfigmapCollectBuffer[T]
	originalCm  ConfigMap[T]
	processedCm ConfigMap[T]
	isChanged   bool
}

func init() {
	DeviceCenterCmManager = &FaultCenterCmManager[*constant.AdvanceDeviceFaultCm]{
		mutex:       sync.RWMutex{},
		originalCm:  ConfigMap[*constant.AdvanceDeviceFaultCm]{Data: make(map[string]*constant.AdvanceDeviceFaultCm)},
		processedCm: ConfigMap[*constant.AdvanceDeviceFaultCm]{Data: make(map[string]*constant.AdvanceDeviceFaultCm)},
		cmBuffer:    collector.DeviceCmCollectBuffer,
	}
	SwitchCenterCmManager = &FaultCenterCmManager[*constant.SwitchInfo]{
		mutex:       sync.RWMutex{},
		originalCm:  ConfigMap[*constant.SwitchInfo]{Data: make(map[string]*constant.SwitchInfo)},
		processedCm: ConfigMap[*constant.SwitchInfo]{Data: make(map[string]*constant.SwitchInfo)},
		cmBuffer:    collector.SwitchCmCollectBuffer,
	}
	NodeCenterCmManager = &FaultCenterCmManager[*constant.NodeInfo]{
		mutex:       sync.RWMutex{},
		originalCm:  ConfigMap[*constant.NodeInfo]{Data: make(map[string]*constant.NodeInfo)},
		processedCm: ConfigMap[*constant.NodeInfo]{Data: make(map[string]*constant.NodeInfo)},
		cmBuffer:    collector.NodeCmCollectBuffer,
	}
	DpuCenterCMManager = &FaultCenterCmManager[*constant.DpuInfoCM]{
		mutex:       sync.RWMutex{},
		originalCm:  ConfigMap[*constant.DpuInfoCM]{Data: make(map[string]*constant.DpuInfoCM)},
		processedCm: ConfigMap[*constant.DpuInfoCM]{Data: make(map[string]*constant.DpuInfoCM)},
		cmBuffer:    collector.DpuCMCollectBuffer,
	}
}

// GetOriginalCm return original configmap
func (manager *FaultCenterCmManager[T]) GetOriginalCm() ConfigMap[T] {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	return manager.originalCm.deepCopy()
}

// SetProcessedCm set processed configmap
func (manager *FaultCenterCmManager[T]) SetProcessedCm(cm ConfigMap[T]) bool {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	if manager.processedCm.equal(cm) {
		manager.isChanged = false
		return false
	}
	manager.isChanged = true
	manager.processedCm = cm.deepCopy()
	return true
}

// GetProcessedCm return processed configmap
func (manager *FaultCenterCmManager[T]) GetProcessedCm() ConfigMap[T] {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	return manager.processedCm.deepCopy()
}

func (manager *FaultCenterCmManager[T]) updateOriginalCm(newInfo T, isAdd bool) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	manager.originalCm.updateCmInfo(newInfo, isAdd)
}

// UpdateBatchOriginalCm update original configmap
func (manager *FaultCenterCmManager[T]) UpdateBatchOriginalCm() []constant.InformerCmItem[T] {
	informerCms := manager.cmBuffer.Pop()
	for _, cm := range informerCms {
		manager.updateOriginalCm(cm.Data, cm.IsAdd)
	}
	return informerCms
}

// IsChanged return whether the processedCm is changed
func (manager *FaultCenterCmManager[T]) IsChanged() bool {
	return manager.isChanged
}

func (cm *ConfigMap[T]) deepCopy() ConfigMap[T] {
	result := new(map[string]T)
	err := util.DeepCopy(result, cm.Data)
	if err != nil {
		hwlog.RunLog.Errorf("deepCopy deviceInfoCm failed: %v", err)
		return ConfigMap[T]{}
	}
	return ConfigMap[T]{Data: *result}
}

func (cm *ConfigMap[T]) updateCmInfo(newInfo T, isAdd bool) {
	if isAdd {
		if cm.Data == nil {
			cm.Data = make(map[string]T)
		}
		length := len(cm.Data)
		if length > constant.MaxSupportNodeNum {
			hwlog.RunLog.Errorf("updateCmInfo %s failed, exceed length", util.ObjToString(newInfo))
			return
		}
		newInfo.UpdateFaultReceiveTime(cm.Data[newInfo.GetCmName()])
		cm.Data[newInfo.GetCmName()] = newInfo
		hwlog.RunLog.Debugf("add DeviceInfo: %s", util.ObjToString(newInfo))
		return
	}
	delete(cm.Data, newInfo.GetCmName())
}

func (cm *ConfigMap[T]) equal(other ConfigMap[T]) bool {
	if len(cm.Data) != len(other.Data) {
		return false
	}
	for cmName, info := range cm.Data {
		otherInfo, found := other.Data[cmName]
		if !found {
			return false
		}
		if !info.IsSame(otherInfo) {
			return false
		}
	}
	return true
}
