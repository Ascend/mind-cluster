// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"sync"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/switchinfo"
)

func newSwitchFaultProcessCenter() *switchFaultProcessCenter {
	return &switchFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(),
		processingCm:    make(map[string]*constant.SwitchInfo),
		originalCm:      make(map[string]*constant.SwitchInfo),
		mutex:           sync.RWMutex{},
	}
}

func (switchCenter *switchFaultProcessCenter) getProcessingCm() map[string]*constant.SwitchInfo {
	switchCenter.mutex.RLock()
	defer switchCenter.mutex.RUnlock()
	return switchinfo.DeepCopyInfos(switchCenter.processingCm)
}

func (switchCenter *switchFaultProcessCenter) setProcessingCm(infos map[string]*constant.SwitchInfo) {
	switchCenter.mutex.Lock()
	defer switchCenter.mutex.Unlock()
	switchCenter.processingCm = switchinfo.DeepCopyInfos(infos)
}

func (switchCenter *switchFaultProcessCenter) getProcessedCm() map[string]*constant.SwitchInfo {
	switchCenter.mutex.RLock()
	defer switchCenter.mutex.RUnlock()
	return switchinfo.DeepCopyInfos(switchCenter.processedCm)
}

func (switchCenter *switchFaultProcessCenter) setProcessedCm(infos map[string]*constant.SwitchInfo) {
	switchCenter.mutex.Lock()
	defer switchCenter.mutex.Unlock()
	switchCenter.processedCm = switchinfo.DeepCopyInfos(infos)
}

func (switchCenter *switchFaultProcessCenter) updateOriginalCm(newInfo *constant.SwitchInfo) {
	switchCenter.mutex.Lock()
	defer switchCenter.mutex.Unlock()
	length := len(switchCenter.originalCm)
	if length > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("SwitchInfo length=%d > %d, SwitchInfo cm name=%s save failed",
			length, constant.MaxSupportNodeNum, newInfo.CmName)
		return
	}
	switchCenter.originalCm[newInfo.CmName] = newInfo
}

func (switchCenter *switchFaultProcessCenter) delOriginalCm(newInfo *constant.SwitchInfo) {
	switchCenter.mutex.Lock()
	defer switchCenter.mutex.Unlock()
	delete(switchCenter.originalCm, newInfo.CmName)
}

func (switchCenter *switchFaultProcessCenter) process() {
	currentTime := time.Now().UnixMilli()
	if switchCenter.isProcessLimited(currentTime) {
		return
	}
	switchCenter.lastProcessTime = currentTime
	switchCenter.setProcessingCm(switchCenter.originalCm)
	switchCenter.baseFaultCenter.process()
	switchCenter.setProcessedCm(switchCenter.processingCm)
}
