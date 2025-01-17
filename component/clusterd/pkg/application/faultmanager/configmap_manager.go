// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

func (manager *faultCenterCmManager[T]) getOriginalCm() configMap[T] {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	return manager.originalCm.deepCopy()
}

func (manager *faultCenterCmManager[T]) setProcessingCm(cm configMap[T]) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	manager.processingCm = cm.deepCopy()
}

func (manager *faultCenterCmManager[T]) getProcessingCm() configMap[T] {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	return manager.processingCm.deepCopy()
}

func (manager *faultCenterCmManager[T]) setProcessedCm(cm configMap[T]) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	manager.processedCm = cm.deepCopy()
}

func (manager *faultCenterCmManager[T]) getProcessedCm() configMap[T] {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	return manager.processedCm.deepCopy()
}

func (manager *faultCenterCmManager[T]) updateOriginalCm(newInfo T, isAdd bool) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	manager.originalCm.updateCmInfo(newInfo, isAdd)
}

func (manager *faultCenterCmManager[T]) updateBatchOriginalCm() []constant.InformerCmItem[T] {
	informerCms := manager.cmBuffer.Pop()
	for _, cm := range informerCms {
		manager.updateOriginalCm(cm.Data, cm.IsAdd)
	}
	return informerCms
}

func (cm *configMap[T]) deepCopy() configMap[T] {
	result := new(map[string]T)
	err := util.DeepCopy(result, cm.configmap)
	if err != nil {
		hwlog.RunLog.Errorf("deepCopy deviceInfoCm failed: %v", err)
		return configMap[T]{}
	}
	return configMap[T]{configmap: *result}
}

func (cm *configMap[T]) updateCmInfo(newInfo T, isAdd bool) {
	if isAdd {
		if cm.configmap == nil {
			cm.configmap = make(map[string]T)
		}
		length := len(cm.configmap)
		if length > constant.MaxSupportNodeNum {
			hwlog.RunLog.Errorf("updateCmInfo %s failed, exceed length", util.ObjToString(newInfo))
			return
		}
		cm.configmap[newInfo.GetCmName()] = newInfo
		hwlog.RunLog.Debugf("add DeviceInfo: %s", util.ObjToString(newInfo))
	} else {
		delete(cm.configmap, newInfo.GetCmName())
	}
}
