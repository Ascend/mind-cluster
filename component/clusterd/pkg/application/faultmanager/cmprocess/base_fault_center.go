// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package cmprocess contain cm processor
package cmprocess

import (
	"fmt"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain/cmmanager"
)

type baseFaultCenter[T constant.ConfigMapInterface] struct {
	processorList        []constant.FaultProcessor
	lastProcessTime      int64
	subscribeChannelList []chan int
	mutex                sync.Mutex
	processPeriod        int64
	cmManager            *cmmanager.FaultCenterCmManager[T]
	centerType           int
}

func newBaseFaultCenter[T constant.ConfigMapInterface](cmManager *cmmanager.FaultCenterCmManager[T], centerType int) baseFaultCenter[T] {
	return baseFaultCenter[T]{
		processorList:        make([]constant.FaultProcessor, 0),
		lastProcessTime:      0,
		subscribeChannelList: make([]chan int, 0),
		mutex:                sync.Mutex{},
		processPeriod:        constant.FaultCenterProcessPeriod,
		cmManager:            cmManager,
		centerType:           centerType,
	}
}

func (baseCenter *baseFaultCenter[T]) isProcessLimited(currentTime int64) bool {
	return baseCenter.lastProcessTime+baseCenter.processPeriod > currentTime
}

func (baseCenter *baseFaultCenter[T]) Process() {
	currentTime := time.Now().UnixMilli()
	if baseCenter.isProcessLimited(currentTime) {
		return
	}
	baseCenter.lastProcessTime = currentTime
	updateOriginalCm := baseCenter.updateOriginalCm()
	baseCenter.setProcessingCm(baseCenter.getOriginalCm())
	for _, processor := range baseCenter.processorList {
		processingCm := baseCenter.getProcessingCm()
		info := constant.OneConfigmapContent[T]{
			AllConfigmap:    processingCm,
			UpdateConfigmap: updateOriginalCm,
		}
		processingCm = processor.Process(info).(constant.OneConfigmapContent[T]).AllConfigmap
		baseCenter.setProcessingCm(processingCm)
	}
	if baseCenter.setProcessedCm(baseCenter.getProcessingCm()) {
		baseCenter.notifySubscriber()
	}
}

func (baseCenter *baseFaultCenter[T]) notifySubscriber() {
	for _, ch := range baseCenter.subscribeChannelList {
		if ch != nil {
			select {
			case ch <- baseCenter.centerType:
			default:
				hwlog.RunLog.Warnf("send %d notify failed.", baseCenter.centerType)
			}
		}
	}
}

func (baseCenter *baseFaultCenter[T]) addProcessors(processors []constant.FaultProcessor) {
	baseCenter.processorList = append(baseCenter.processorList, processors...)
}

// Register notify chan
func (baseCenter *baseFaultCenter[T]) Register(ch chan int) error {
	baseCenter.mutex.Lock()
	defer baseCenter.mutex.Unlock()
	if baseCenter.subscribeChannelList == nil {
		baseCenter.subscribeChannelList = make([]chan int, 0)
	}
	length := len(baseCenter.subscribeChannelList)
	if length > constant.MaxFaultCenterSubscriber {
		return fmt.Errorf("the number of registrants is %d, cannot add any more", length)
	}
	baseCenter.subscribeChannelList = append(baseCenter.subscribeChannelList, ch)
	return nil
}

func (baseCenter *baseFaultCenter[T]) getOriginalCm() map[string]T {
	return baseCenter.cmManager.GetOriginalCm().Data
}

func (baseCenter *baseFaultCenter[T]) setProcessingCm(cm map[string]T) {
	baseCenter.cmManager.SetProcessingCm(cmmanager.ConfigMap[T]{Data: cm})
}

func (baseCenter *baseFaultCenter[T]) getProcessingCm() map[string]T {
	return baseCenter.cmManager.GetProcessingCm().Data
}

func (baseCenter *baseFaultCenter[T]) setProcessedCm(cm map[string]T) bool {
	return baseCenter.cmManager.SetProcessedCm(cmmanager.ConfigMap[T]{Data: cm})
}

func (baseCenter *baseFaultCenter[T]) GetProcessedCm() map[string]T {
	return baseCenter.cmManager.GetProcessedCm().Data
}

func (baseCenter *baseFaultCenter[T]) updateOriginalCm() []constant.InformerCmItem[T] {
	return baseCenter.cmManager.UpdateBatchOriginalCm()
}
