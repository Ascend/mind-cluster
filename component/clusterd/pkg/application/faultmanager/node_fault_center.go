// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"sync"

	"clusterd/pkg/application/faultmanager/collector"
	"clusterd/pkg/common/constant"
)

func NewNodeFaultProcessCenter() *NodeFaultProcessCenter {
	manager := faultCenterCmManager[*constant.NodeInfo]{
		mutex:        sync.RWMutex{},
		originalCm:   configMap[*constant.NodeInfo]{configmap: make(map[string]*constant.NodeInfo)},
		processingCm: configMap[*constant.NodeInfo]{configmap: make(map[string]*constant.NodeInfo)},
		processedCm:  configMap[*constant.NodeInfo]{configmap: make(map[string]*constant.NodeInfo)},
		cmBuffer:     collector.NodeCmCollectBuffer,
	}
	return &NodeFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(&manager, constant.NodeProcessType),
	}
}
