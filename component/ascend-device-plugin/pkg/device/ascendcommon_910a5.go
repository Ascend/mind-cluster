/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package device a series of device function
package device

import (
	"Ascend-device-plugin/pkg/common"
	"ascend-common/common-utils/hwlog"
)

// SetSuperPodType setting the type of super pod
func (tool *AscendTools) SetSuperPodType(superPodType int8) {
	tool.superPodType = superPodType
}

// GetSuperPodType getting the type of super pod
func (tool *AscendTools) GetSuperPodType() int8 {
	return tool.superPodType
}

// SetSuperPodSize setting the type of super pod
func (tool *AscendTools) SetSuperPodSize(superPodSize int32) {
	tool.superPodSize = superPodSize
}

// GetSuperPodSize getting the type of super pod
func (tool *AscendTools) GetSuperPodSize() int32 {
	return tool.superPodSize
}

// SetNodeInternalIPInK8s setting the ip of the node server in k8s
func (tool *AscendTools) SetNodeInternalIPInK8s(nodeIp string) {
	tool.nodeInternalIP = nodeIp
}

// GetNodeInternalIPInK8s getting the ip of the node server in k8s
func (tool *AscendTools) GetNodeInternalIPInK8s() string {
	return tool.nodeInternalIP
}

// SetRackID setting the rank id
func (tool *AscendTools) SetRackID(rackID int32) {
	tool.rackID = rackID
}

// GetRackID getting the rack id
func (tool *AscendTools) GetRackID() int32 {
	return tool.rackID
}

func (tool *AscendTools) writeNodeDeviceInfoDataA5(newDeviceList map[string]string, manuallySeparateNPU string,
	switchFaultInfo common.SwitchFaultInfo) (bool, error) {
	nodeDeviceData := &common.NodeDeviceInfoCache{
		DeviceInfo: common.NodeDeviceInfo{
			DeviceList: newDeviceList,
		},
		SuperPodID:  tool.GetSuperPodID(),
		RackID:      tool.GetRackID(),
		ServerIndex: tool.GetServerIndex(),
	}
	if err := tool.client.WriteDeviceInfoDataIntoCMCacheA5(nodeDeviceData, manuallySeparateNPU,
		switchFaultInfo); err != nil {
		hwlog.RunLog.Errorf("write device info failed: %v", err)
		return false, nil
	}
	return true, nil
}
