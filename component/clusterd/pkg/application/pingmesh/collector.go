// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package pingmesh a series of function handle ping mesh configmap create/update/delete
package pingmesh

import (
	"strconv"

	v1 "k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/superpod"
)

// NodeCollector collector node info
func NodeCollector(oldNodeInfo, newNodeInfo *v1.Node, operator string) {
	nodeDevice, superPodID := node.GetNodeDeviceAndSuperPodID(newNodeInfo)
	if superPodID == "" || nodeDevice == nil {
		hwlog.RunLog.Debugf("discard illegal super pod device info, superPodID=%s.", superPodID)
		return
	}
	spIdIntValue, err := strconv.Atoi(superPodID)
	if spIdIntValue < 0 || err != nil {
		hwlog.RunLog.Debugf("superPodID=%s cannot converto a natural number", superPodID)
		return
	}
	switch operator {
	case constant.AddOperator, constant.UpdateOperator:
		superpod.SaveNode(superPodID, nodeDevice)
		addEvent(superPodID, constant.UpdateOperator)
	case constant.DeleteOperator:
		superpod.DeleteNode(superPodID, newNodeInfo.Name)
		device := superpod.GetSuperPodDevice(superPodID)
		if device == nil {
			addEvent(superPodID, constant.DeleteOperator)
			return
		}
		addEvent(superPodID, constant.UpdateOperator)
	default:
		hwlog.RunLog.Errorf("error operator: %s", operator)
	}
}
