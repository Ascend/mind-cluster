// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package resource a series of resource function
package resource

import (
	"clusterd/pkg/common/constant"
)

// DeviceInfoCollector collector device info
func DeviceInfoCollector(oldDevInfo, newDevInfo *constant.DeviceInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		saveDeviceInfoCM(newDevInfo)
	} else if operator == constant.DeleteOperator {
		delDeviceInfoCM(newDevInfo)
	}
}

// SwitchInfoCollector collector switchinfo info of 900A3
func SwitchInfoCollector(oldSwitchInfo, newSwitchInfo *constant.SwitchInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		saveSwitchInfoCM(newSwitchInfo)
	} else if operator == constant.DeleteOperator {
		delSwitchInfoCM(newSwitchInfo)
	}
}

// NodeCollector collector node info
func NodeCollector(oldNodeInfo, newNodeInfo *constant.NodeInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		saveNodeInfoCM(newNodeInfo)
	} else if operator == constant.DeleteOperator {
		deleteNodeConfigMap(newNodeInfo)
	}
}
