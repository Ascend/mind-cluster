// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultshoot contain fault process
package faultshoot

import (
	"sync"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/node"
)

func newNodeFaultProcessCenter() *nodeFaultProcessCenter {
	return &nodeFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(),
		processingCm:    make(map[string]*constant.NodeInfo),
		devicePluginCm:  make(map[string]*constant.NodeInfo),
		mutex:           sync.RWMutex{},
	}
}

func (nodeCenter *nodeFaultProcessCenter) getProcessingCm() map[string]*constant.NodeInfo {
	nodeCenter.mutex.RLock()
	defer nodeCenter.mutex.RUnlock()
	return node.DeepCopyInfos(nodeCenter.processingCm)
}

func (nodeCenter *nodeFaultProcessCenter) setProcessingCm(infos map[string]*constant.NodeInfo) {
	nodeCenter.mutex.Lock()
	defer nodeCenter.mutex.Unlock()
	nodeCenter.processingCm = node.DeepCopyInfos(infos)
}

func (nodeCenter *nodeFaultProcessCenter) getProcessedCm() map[string]*constant.NodeInfo {
	nodeCenter.mutex.RLock()
	defer nodeCenter.mutex.RUnlock()
	return node.DeepCopyInfos(nodeCenter.processedCm)
}

func (nodeCenter *nodeFaultProcessCenter) setProcessedCm(infos map[string]*constant.NodeInfo) {
	nodeCenter.mutex.Lock()
	defer nodeCenter.mutex.Unlock()
	nodeCenter.processedCm = node.DeepCopyInfos(infos)
}

func (nodeCenter *nodeFaultProcessCenter) updateDevicePluginCm(newInfo *constant.NodeInfo) {
	nodeCenter.mutex.Lock()
	defer nodeCenter.mutex.Unlock()
	length := len(nodeCenter.devicePluginCm)
	if length > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("SwitchInfo length=%d > %d, SwitchInfo cm name=%s save failed",
			length, constant.MaxSupportNodeNum, newInfo.CmName)
		return
	}
	nodeCenter.devicePluginCm[newInfo.CmName] = newInfo
}

func (nodeCenter *nodeFaultProcessCenter) delDevicePluginCm(newInfo *constant.NodeInfo) {
	nodeCenter.mutex.Lock()
	defer nodeCenter.mutex.Unlock()
	delete(nodeCenter.devicePluginCm, newInfo.CmName)
}

func (nodeCenter *nodeFaultProcessCenter) process() {
	currentTime := time.Now().UnixMilli()
	if nodeCenter.isProcessLimited(currentTime) {
		return
	}
	nodeCenter.lastProcessTime = currentTime
	nodeCenter.setProcessingCm(nodeCenter.devicePluginCm)
	nodeCenter.baseFaultCenter.process()
	nodeCenter.setProcessedCm(nodeCenter.processedCm)
}
