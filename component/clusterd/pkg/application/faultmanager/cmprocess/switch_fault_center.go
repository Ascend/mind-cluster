// Copyright (c) Huawei Technologies Co., Ltd. 2024-2026. All rights reserved.

// Package cmprocess contain cm processor
package cmprocess

import (
	"clusterd/pkg/application/faultmanager/cmprocess/custom"
	"clusterd/pkg/application/faultmanager/cmprocess/preseparate"
	"clusterd/pkg/application/faultmanager/cmprocess/retry"
	"clusterd/pkg/application/silentfault"
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
		custom.CustomProcessor,
		retry.RetryProcessor,
		preseparate.PreSeparateFaultProcessor, // this processor process the preSeparate faults.
		silentfault.FaultRecorderProcessor,    // this processor records hardware fault timeline.
	})
}

type switchFaultProcessCenter struct {
	baseFaultCenter[*constant.SwitchInfo]
}
