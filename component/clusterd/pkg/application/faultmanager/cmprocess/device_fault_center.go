// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package cmprocess contain cm processor
package cmprocess

import (
	"clusterd/pkg/application/faultmanager/cmprocess/publicfault"
	"clusterd/pkg/application/faultmanager/cmprocess/uce"
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
		uce.UceProcessor,                   // this processor filter the uce faults.
	})
}
