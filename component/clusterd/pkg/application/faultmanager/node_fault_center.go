// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"sync"
	"time"

	"clusterd/pkg/common/constant"
)

func newNodeFaultProcessCenter() *nodeFaultProcessCenter {
	manager := faultCenterCmManager[*constant.NodeInfo]{
		mutex:        sync.RWMutex{},
		originalCm:   configMap[*constant.NodeInfo]{configmap: make(map[string]*constant.NodeInfo)},
		processingCm: configMap[*constant.NodeInfo]{configmap: make(map[string]*constant.NodeInfo)},
		processedCm:  configMap[*constant.NodeInfo]{configmap: make(map[string]*constant.NodeInfo)},
	}
	return &nodeFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(&manager),
	}
}

func (nodeCenter *nodeFaultProcessCenter) process() {
	currentTime := time.Now().UnixMilli()
	if nodeCenter.isProcessLimited(currentTime) {
		return
	}
	nodeCenter.lastProcessTime = currentTime
	nodeCenter.setProcessingCm(nodeCenter.getOriginalCm())
	nodeCenter.baseFaultCenter.process()
	nodeCenter.setProcessedCm(nodeCenter.getProcessingCm())
}
