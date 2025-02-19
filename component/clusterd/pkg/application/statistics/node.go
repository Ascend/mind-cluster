// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package statistics statistic funcs about node
package statistics

import (
	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/statistics"
)

const nodeAnnotation = "product-serial-number"

// UpdateNodeSNAndNameCache update node sn and name cache
func UpdateNodeSNAndNameCache(nodeInfo *v1.Node, operator string) {
	if nodeInfo == nil {
		return
	}
	nodeSN, ok := nodeInfo.Annotations[nodeAnnotation]
	if !ok {
		return
	}
	nodeName := nodeInfo.Name
	switch operator {
	case constant.AddOperator, constant.UpdateOperator:
		statistics.GetNodeSNAndNameCache()[nodeSN] = nodeName
	case constant.DeleteOperator:
		delete(statistics.GetNodeSNAndNameCache(), nodeSN)
	default:
		hwlog.RunLog.Error("invalid operator")
	}
}
