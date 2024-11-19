// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultshoot contain fault process
package faultshoot

import (
	"clusterd/pkg/common/constant"
)

// DeviceInfoCollector collects device info
func DeviceInfoCollector(oldDevInfo, newDevInfo *constant.DeviceInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		GlobalFaultProcessCenter.informDeviceInfoAdd(newDevInfo)
	} else if operator == constant.DeleteOperator {
		GlobalFaultProcessCenter.informDeviceInfoDel(newDevInfo)
	}
}

// SwitchInfoCollector collects switchinfo info of 900A3
func SwitchInfoCollector(oldSwitchInfo, newSwitchInfo *constant.SwitchInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		GlobalFaultProcessCenter.informSwitchInfoAdd(newSwitchInfo)
	} else if operator == constant.DeleteOperator {
		GlobalFaultProcessCenter.informSwitchInfoDel(newSwitchInfo)
	}
}

// NodeCollector collects node info
func NodeCollector(oldNodeInfo, newNodeInfo *constant.NodeInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		GlobalFaultProcessCenter.informNodeInfoAdd(newNodeInfo)
	} else if operator == constant.DeleteOperator {
		GlobalFaultProcessCenter.informNodeInfoDel(newNodeInfo)
	}
}
