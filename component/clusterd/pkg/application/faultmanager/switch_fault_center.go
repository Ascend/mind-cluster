// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"sync"

	"clusterd/pkg/application/faultmanager/collector"
	"clusterd/pkg/common/constant"
)

func NewSwitchFaultProcessCenter() *SwitchFaultProcessCenter {
	manager := faultCenterCmManager[*constant.SwitchInfo]{
		mutex:        sync.RWMutex{},
		originalCm:   configMap[*constant.SwitchInfo]{configmap: make(map[string]*constant.SwitchInfo)},
		processingCm: configMap[*constant.SwitchInfo]{configmap: make(map[string]*constant.SwitchInfo)},
		processedCm:  configMap[*constant.SwitchInfo]{configmap: make(map[string]*constant.SwitchInfo)},
		cmBuffer:     collector.SwitchCmCollectBuffer,
	}
	return &SwitchFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(&manager, constant.SwitchProcessType),
	}
}
