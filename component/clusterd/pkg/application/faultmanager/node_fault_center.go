// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"sync"

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
		baseFaultCenter: newBaseFaultCenter(&manager, constant.NodeProcessType),
	}
}
