// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package cmprocess contain cm processor
package cmprocess

import (
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain/cmmanager"
)

type dpuFaultProcessCenter struct {
	baseFaultCenter[*constant.DpuInfoCM]
}

// DpuCenter process dpu cm info
var DpuCenter *dpuFaultProcessCenter

func init() {
	manager := cmmanager.DpuCenterCMManager
	DpuCenter = &dpuFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(manager, constant.DpuProcessType),
	}
}
