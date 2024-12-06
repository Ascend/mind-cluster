// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

func informInfoUpdate(newInfo any, whichToInformer int, isAdd bool) {
	switch whichToInformer {
	case constant.DeviceFaultType:
		GlobalFaultProcessCenter.deviceCenter.updateOriginalCm(newInfo.(*constant.DeviceInfo), isAdd)
	case constant.NodeFaultType:
		GlobalFaultProcessCenter.nodeCenter.updateOriginalCm(newInfo.(*constant.NodeInfo), isAdd)
	case constant.SwitchFaultType:
		GlobalFaultProcessCenter.switchCenter.updateOriginalCm(newInfo.(*constant.SwitchInfo), isAdd)
	default:
		hwlog.RunLog.Errorf("cannot process %d", whichToInformer)
		return
	}
	GlobalFaultProcessCenter.notifyFaultCenterProcess(whichToInformer)
}

// DeviceInfoCollector collects device info
func DeviceInfoCollector(oldDevInfo, newDevInfo *constant.DeviceInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		informInfoUpdate(newDevInfo, constant.DeviceFaultType, true)
	} else if operator == constant.DeleteOperator {
		informInfoUpdate(newDevInfo, constant.DeviceFaultType, false)
	}
}

// SwitchInfoCollector collects switchinfo info of 900A3
func SwitchInfoCollector(oldSwitchInfo, newSwitchInfo *constant.SwitchInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		informInfoUpdate(newSwitchInfo, constant.SwitchFaultType, true)
	} else if operator == constant.DeleteOperator {
		informInfoUpdate(newSwitchInfo, constant.SwitchFaultType, false)
	}
}

// NodeCollector collects node info
func NodeCollector(oldNodeInfo, newNodeInfo *constant.NodeInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		informInfoUpdate(newNodeInfo, constant.NodeFaultType, true)
	} else if operator == constant.DeleteOperator {
		informInfoUpdate(newNodeInfo, constant.NodeFaultType, false)
	}
}
