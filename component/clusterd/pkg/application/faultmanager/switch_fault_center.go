// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"sync"
	"time"

	"clusterd/pkg/common/constant"
)

func newSwitchFaultProcessCenter() *switchFaultProcessCenter {
	manager := faultCenterCmManager[*constant.SwitchInfo]{
		mutex:        sync.RWMutex{},
		originalCm:   configMap[*constant.SwitchInfo]{configmap: make(map[string]*constant.SwitchInfo)},
		processingCm: configMap[*constant.SwitchInfo]{configmap: make(map[string]*constant.SwitchInfo)},
		processedCm:  configMap[*constant.SwitchInfo]{configmap: make(map[string]*constant.SwitchInfo)},
	}
	return &switchFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(&manager),
	}
}

func (switchCenter *switchFaultProcessCenter) process() {
	currentTime := time.Now().UnixMilli()
	if switchCenter.isProcessLimited(currentTime) {
		return
	}
	switchCenter.lastProcessTime = currentTime
	switchCenter.setProcessingCm(switchCenter.getOriginalCm())
	switchCenter.baseFaultCenter.process()
	switchCenter.setProcessedCm(switchCenter.getProcessingCm())
}
