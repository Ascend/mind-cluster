// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package cmprocess contain cm processor
package cmprocess

import (
	"clusterd/pkg/application/faultmanager/cmprocess/publicfault"
	"clusterd/pkg/application/faultmanager/cmprocess/retry"
	"clusterd/pkg/application/faultmanager/cmprocess/stresstest"
	"clusterd/pkg/application/faultmanager/cmprocess/uceaccompany"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain/cmmanager"
)

// DeviceCenter process device cm info
var DeviceCenter *deviceFaultProcessCenter

// deviceFaultProcessCenter
type deviceFaultProcessCenter struct {
	baseFaultCenter[*constant.AdvanceDeviceFaultCm]
}

func init() {
	manager := cmmanager.DeviceCenterCmManager
	DeviceCenter = &deviceFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(manager, constant.DeviceProcessType),
	}

	DeviceCenter.addProcessors([]constant.FaultProcessor{
		publicfault.PubFaultProcessor,
		uceaccompany.UceAccompanyProcessor, // this processor filter the uce accompany faults, before processorForUceFault
		retry.RetryProcessor,               // this processor filter the uce faults.
		stresstest.StressTestProcessor,     // this processor filter the stress test faults.
	})
}
