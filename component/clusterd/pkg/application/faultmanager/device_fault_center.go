// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"fmt"
	"sync"

	"clusterd/pkg/application/faultmanager/collector"
	"clusterd/pkg/application/faultmanager/uce"
	"clusterd/pkg/common/constant"
)

func NewDeviceFaultProcessCenter() *DeviceFaultProcessCenter {
	manager := faultCenterCmManager[*constant.DeviceInfo]{
		mutex:        sync.RWMutex{},
		originalCm:   configMap[*constant.DeviceInfo]{configmap: make(map[string]*constant.DeviceInfo)},
		processingCm: configMap[*constant.DeviceInfo]{configmap: make(map[string]*constant.DeviceInfo)},
		processedCm:  configMap[*constant.DeviceInfo]{configmap: make(map[string]*constant.DeviceInfo)},
		cmBuffer:     collector.DeviceCmCollectBuffer,
	}
	deviceCenter := &DeviceFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(&manager, constant.DeviceProcessType),
	}

	var processorForUceAccompanyFault = newUceAccompanyFaultProcessor(deviceCenter)
	var processorForUceFault = uce.NewUceFaultProcessor()
	var processForJobFaultRank = newJobRankFaultInfoProcessor(deviceCenter)

	deviceCenter.addProcessors([]constant.FaultProcessor{
		processorForUceAccompanyFault, // this processor filter the uce accompany faults, before processorForUceFault
		processorForUceFault,          // this processor filter the uce faults.
		processForJobFaultRank,        // this processor need to get filtered faults
	})
	return deviceCenter
}

func (deviceCenter *DeviceFaultProcessCenter) getUceAccompanyFaultProcessor() (*uceAccompanyFaultProcessor, error) {
	for _, processor := range deviceCenter.processorList {
		if processor, ok := processor.(*uceAccompanyFaultProcessor); ok {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("can not find uceAccompanyFaultProcessor in FaultProcessCenter")
}

func (deviceCenter *DeviceFaultProcessCenter) getJobFaultRankProcessor() (*jobRankFaultInfoProcessor, error) {
	for _, processor := range deviceCenter.processorList {
		if processor, ok := processor.(*jobRankFaultInfoProcessor); ok {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("can not find jobRankFaultInfoProcessor in FaultProcessCenter")
}
