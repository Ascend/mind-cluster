// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package cmprocess contain cm processor
package cmprocess

import (
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain/cmmanager"
)

// NodeCenter process node cm info
var NodeCenter *nodeFaultProcessCenter

func init() {
	manager := cmmanager.NodeCenterCmManager
	NodeCenter = &nodeFaultProcessCenter{
		baseFaultCenter: newBaseFaultCenter(manager, constant.NodeProcessType),
	}
}

// nodeFaultProcessCenter
type nodeFaultProcessCenter struct {
	baseFaultCenter[*constant.NodeInfo]
}
