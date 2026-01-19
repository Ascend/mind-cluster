// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package cmprocess contain cm processor
package cmprocess

import (
	"clusterd/pkg/application/faultmanager/cmprocess/l2fault"
	"clusterd/pkg/application/faultmanager/cmprocess/retry"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain/cmmanager"
)

// SwitchCenter process switch cm info
var SwitchCenter *switchFaultProcessCenter

func init() {
	manager := cmmanager.SwitchCenterCmManager
	SwitchCenter = &switchFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(manager, constant.SwitchProcessType),
	}
	SwitchCenter.addProcessors([]constant.FaultProcessor{
		l2fault.L2FaultProcessor,
		retry.RetryProcessor,
	})
}

type switchFaultProcessCenter struct {
	baseFaultCenter[*constant.SwitchInfo]
}
