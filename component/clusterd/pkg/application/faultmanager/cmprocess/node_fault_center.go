// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package cmprocess contain cm processor
package cmprocess

import (
	"clusterd/pkg/application/faultmanager/cmprocess/preseparate"
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
	NodeCenter.addProcessors([]constant.FaultProcessor{
		preseparate.PreSeparateFaultProcessor, // this processor process the preSeparate faults.
	})
}

// nodeFaultProcessCenter
type nodeFaultProcessCenter struct {
	baseFaultCenter[*constant.NodeInfo]
}
