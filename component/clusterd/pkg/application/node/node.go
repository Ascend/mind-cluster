// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package node funcs about node
package node

import (
	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/node"
)

// UpdateNodeInfoCache update node info cache
func UpdateNodeInfoCache(_, newNodeInfo *v1.Node, operator string) {
	if newNodeInfo == nil {
		return
	}
	switch operator {
	case constant.AddOperator, constant.UpdateOperator:
		node.SaveNodeToCache(newNodeInfo)
	case constant.DeleteOperator:
		node.DeleteNodeFromCache(newNodeInfo)
	default:
		hwlog.RunLog.Error("invalid operator")
	}
}
