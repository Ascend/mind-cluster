// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"fmt"
	"sync"

	"clusterd/pkg/common/constant"
)

func newBaseFaultCenter[T constant.ConfigMapInterface](cmManager *faultCenterCmManager[T]) baseFaultCenter[T] {
	return baseFaultCenter[T]{
		processorList:        make([]faultProcessor, 0),
		lastProcessTime:      0,
		subscribeChannelList: make([]chan struct{}, 0),
		mutex:                sync.Mutex{},
		processPeriod:        constant.FaultCenterProcessPeriod,
		cmManager:            cmManager,
	}
}

func (baseCenter *baseFaultCenter[T]) isProcessLimited(currentTime int64) bool {
	return baseCenter.lastProcessTime+baseCenter.processPeriod > currentTime
}

func (baseCenter *baseFaultCenter[T]) process() {
	for _, processor := range baseCenter.processorList {
		processor.process()
	}
	for _, ch := range baseCenter.subscribeChannelList {
		if ch != nil {
			ch <- struct{}{}
		}
	}
}

func (baseCenter *baseFaultCenter[T]) addProcessors(processors []faultProcessor) {
	baseCenter.processorList = append(baseCenter.processorList, processors...)
}

func (baseCenter *baseFaultCenter[T]) register(ch chan struct{}) error {
	baseCenter.mutex.Lock()
	defer baseCenter.mutex.Unlock()
	if baseCenter.subscribeChannelList == nil {
		baseCenter.subscribeChannelList = make([]chan struct{}, 0)
	}
	length := len(baseCenter.subscribeChannelList)
	if length > constant.MaxFaultCenterSubscriber {
		return fmt.Errorf("the number of registrants is %d, cannot add any more", length)
	}
	baseCenter.subscribeChannelList = append(baseCenter.subscribeChannelList, ch)
	return nil
}

func (baseCenter *baseFaultCenter[T]) getOriginalCm() map[string]T {
	return baseCenter.cmManager.getOriginalCm().configmap
}

func (baseCenter *baseFaultCenter[T]) setProcessingCm(cm map[string]T) {
	baseCenter.cmManager.setProcessingCm(configMap[T]{configmap: cm})
}

func (baseCenter *baseFaultCenter[T]) getProcessingCm() map[string]T {
	return baseCenter.cmManager.getProcessingCm().configmap
}

func (baseCenter *baseFaultCenter[T]) setProcessedCm(cm map[string]T) {
	baseCenter.cmManager.setProcessedCm(configMap[T]{configmap: cm})
}

func (baseCenter *baseFaultCenter[T]) getProcessedCm() map[string]T {
	return baseCenter.cmManager.getProcessedCm().configmap
}

func (baseCenter *baseFaultCenter[T]) updateOriginalCm(newInfo T, isAdd bool) {
	baseCenter.cmManager.originalCm.updateCmInfo(newInfo, isAdd)
}
