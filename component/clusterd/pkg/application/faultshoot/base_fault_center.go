// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultshoot contain fault process
package faultshoot

import (
	"fmt"
	"sync"

	"clusterd/pkg/common/constant"
)

func newBaseFaultCenter() baseFaultCenter {
	return baseFaultCenter{
		processorList:        make([]faultProcessor, 0),
		lastProcessTime:      0,
		subscribeChannelList: make([]chan struct{}, 0),
		mutex:                sync.Mutex{},
		processPeriod:        constant.FaultCenterProcessPeriod,
	}
}

func (baseCenter *baseFaultCenter) isProcessLimited(currentTime int64) bool {
	return baseCenter.lastProcessTime+baseCenter.processPeriod > currentTime
}

func (baseCenter *baseFaultCenter) process() {
	for _, processor := range baseCenter.processorList {
		processor.process()
	}
	for _, ch := range baseCenter.subscribeChannelList {
		if ch != nil {
			ch <- struct{}{}
		}
	}
}

func (baseCenter *baseFaultCenter) addProcessors(processors []faultProcessor) {
	baseCenter.processorList = append(baseCenter.processorList, processors...)
}

func (baseCenter *baseFaultCenter) register(ch chan struct{}) error {
	baseCenter.mutex.Lock()
	defer baseCenter.mutex.Unlock()
	if baseCenter.subscribeChannelList == nil {
		baseCenter.subscribeChannelList = make([]chan struct{}, 0)
	}
	length := len(baseCenter.subscribeChannelList)
	if length > constant.MAX_FAULT_CENTER_SUBSCRIBER {
		return fmt.Errorf("the number of registrants is %d, cannot add any more", length)
	}
	baseCenter.subscribeChannelList = append(baseCenter.subscribeChannelList, ch)
	return nil
}
