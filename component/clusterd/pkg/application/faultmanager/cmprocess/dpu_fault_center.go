// Copyright (c) Huawei Technologies Co., Ltd. 2026-2026. All rights reserved.

// Package cmprocess contain cm processor
package cmprocess

import (
	"clusterd/pkg/application/silentfault"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain/cmmanager"
)

// DpuCenter process dpu cm info and records hardware fault timeline
var DpuCenter *dpuFaultProcessCenter

func init() {
	manager := cmmanager.DpuCenterCmManager
	DpuCenter = &dpuFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(manager, constant.DpuProcessType),
	}

	DpuCenter.addProcessors([]constant.FaultProcessor{
		silentfault.FaultRecorderProcessor, // this processor records hardware fault timeline.
	})
}

// dpuFaultProcessCenter
type dpuFaultProcessCenter struct {
	baseFaultCenter[*constant.DpuInfo]
}
