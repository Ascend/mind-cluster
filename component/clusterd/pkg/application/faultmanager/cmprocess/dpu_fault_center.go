// Copyright (c) Huawei Technologies Co., Ltd. 2026-2026. All rights reserved.

// Package cmprocess contain cm processor
package cmprocess

import (
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain/cmmanager"
)

// DpuCenter process dpu cm info, no fault processors needed for pass-through aggregation
var DpuCenter *dpuFaultProcessCenter

func init() {
	manager := cmmanager.DpuCenterCmManager
	DpuCenter = &dpuFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(manager, constant.DpuProcessType),
	}
}

// dpuFaultProcessCenter
type dpuFaultProcessCenter struct {
	baseFaultCenter[*constant.DpuInfo]
}
