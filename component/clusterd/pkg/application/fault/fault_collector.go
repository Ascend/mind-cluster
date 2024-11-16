package fault

import "clusterd/pkg/common/constant"

// DeviceInfoCollector collector device info
func DeviceInfoCollector(oldDevInfo, newDevInfo *constant.DeviceInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		// GlobalFaultProcessCenter.informDeviceInfoAdd(newDevInfo)
	} else if operator == constant.DeleteOperator {
		// GlobalFaultProcessCenter.informDeviceInfoDel(newDevInfo)
	}
}

// SwitchInfoCollector collector switchinfo info of 900A3
func SwitchInfoCollector(oldSwitchInfo, newSwitchInfo *constant.SwitchInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		// GlobalFaultProcessCenter.informSwitchInfoAdd(newSwitchInfo)
	} else if operator == constant.DeleteOperator {
		// GlobalFaultProcessCenter.informSwitchInfoDel(newSwitchInfo)
	}
}

// NodeCollector collector node info
func NodeCollector(oldNodeInfo, newNodeInfo *constant.NodeInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		// GlobalFaultProcessCenter.informNodeInfoAdd(newNodeInfo)
	} else if operator == constant.DeleteOperator {
		// GlobalFaultProcessCenter.informNodeInfoDel(newNodeInfo)
	}
}
