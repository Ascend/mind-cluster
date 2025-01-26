// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"sync"
	"testing"

	"clusterd/pkg/application/faultmanager/collector"
	"clusterd/pkg/common/constant"
)

type fakeProcessor struct{}

func (f *fakeProcessor) Process(info any) any {
	return info
}

func TestBaseFaultCenterProcess(t *testing.T) {
	t.Run("TestBaseFaultCenterProcess", func(t *testing.T) {
		manager := faultCenterCmManager[*constant.DeviceInfo]{
			mutex:        sync.RWMutex{},
			originalCm:   configMap[*constant.DeviceInfo]{configmap: make(map[string]*constant.DeviceInfo)},
			processingCm: configMap[*constant.DeviceInfo]{configmap: make(map[string]*constant.DeviceInfo)},
			processedCm:  configMap[*constant.DeviceInfo]{configmap: make(map[string]*constant.DeviceInfo)},
			cmBuffer:     collector.DeviceCmCollectBuffer,
		}
		baseCenter := newBaseFaultCenter(&manager, constant.DeviceProcessType)
		baseCenter.addProcessors([]constant.FaultProcessor{&fakeProcessor{}})
		notifyChan := make(chan int, 1)
		baseCenter.register(notifyChan)
		baseCenter.Process()
		if baseCenter.getProcessedCm() == nil {
			t.Errorf("TestBaseFaultCenterProcess failed")
		}
	})
}
