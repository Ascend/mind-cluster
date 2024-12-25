// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"sync"
	"testing"

	"clusterd/pkg/common/constant"
)

type fakeProcessor struct{}

func (f *fakeProcessor) process() {
}

func TestBaseFaultCenterProcess(t *testing.T) {
	t.Run("TestBaseFaultCenterProcess", func(t *testing.T) {
		manager := faultCenterCmManager[*constant.DeviceInfo]{
			mutex:        sync.RWMutex{},
			originalCm:   configMap[*constant.DeviceInfo]{configmap: make(map[string]*constant.DeviceInfo)},
			processingCm: configMap[*constant.DeviceInfo]{configmap: make(map[string]*constant.DeviceInfo)},
			processedCm:  configMap[*constant.DeviceInfo]{configmap: make(map[string]*constant.DeviceInfo)},
		}
		baseCenter := newBaseFaultCenter(&manager, constant.DeviceProcessType)
		deviceInfo := &constant.DeviceInfo{
			DeviceInfoNoName: constant.DeviceInfoNoName{
				DeviceList: make(map[string]string),
				UpdateTime: 0,
			},
			CmName: "test",
		}
		baseCenter.updateOriginalCm(deviceInfo, true)
		baseCenter.addProcessors([]faultProcessor{&fakeProcessor{}})
		notifyChan := make(chan int, 1)
		baseCenter.register(notifyChan)
		baseCenter.process()
		if len(baseCenter.getProcessedCm()) == 0 {
			t.Errorf("TestBaseFaultCenterProcess failed")
		}
		sign := <-notifyChan
		if sign != constant.DeviceProcessType {
			t.Errorf("TestBaseFaultCenterProcess failed")
		}
	})
}
