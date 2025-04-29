// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package cmprocess contain cm processor
package cmprocess

import (
	"testing"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain/cmmanager"
)

type fakeProcessor struct{}

func (f *fakeProcessor) Process(info any) any {
	return info
}

func TestMain(m *testing.M) {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
	m.Run()
}

func TestBaseFaultCenterProcess(t *testing.T) {
	t.Run("TestBaseFaultCenterProcess", func(t *testing.T) {
		manager := cmmanager.DeviceCenterCmManager
		baseCenter := newBaseFaultCenter(manager, constant.DeviceProcessType)
		baseCenter.addProcessors([]constant.FaultProcessor{&fakeProcessor{}})
		notifyChan := make(chan int, 1)
		baseCenter.Register(notifyChan)
		baseCenter.Process()
		if baseCenter.GetProcessedCm() == nil {
			t.Errorf("TestBaseFaultCenterProcess failed")
		}
	})
}
