/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package node funcs about node.
*/
package node

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

const (
	rackIdA5Max = 255
	rackIdMin   = 0
)

func getNodeAcceleratorType(node *v1.Node) string {
	acceleratorType, hasTypeKey := node.Labels[api.AcceleratorTypeKey]
	if !hasTypeKey {
		hwlog.RunLog.Debugf("empty acceleratorType, nodeName=%s", node.Name)
		return ""
	}
	return acceleratorType
}

func getRackIdFromNode(node *v1.Node) (string, error) {
	rackID, hasRackIDKey := node.Annotations[api.RackIDKey]
	rackID = strings.Trim(rackID, " ")
	if !hasRackIDKey || len(rackID) == 0 {
		hwlog.RunLog.Debugf("empty rack id, nodeName=%s", node.Name)
		return "", fmt.Errorf("invalid rack id in node anno")
	}
	return rackID, nil
}

func getRackID(node *v1.Node) string {
	acceleratorType := getNodeAcceleratorType(node)
	if acceleratorType == api.Ascend800ia5SuperPod {
		hwlog.RunLog.Debugf("getRackID is 0 for acceleratorType= %s", acceleratorType)
		return "0"
	}
	rackID, err := getRackIdFromNode(node)
	if err != nil {
		hwlog.RunLog.Debug("getRackIdFromNode is err")
		return ""
	}

	if !api.CheckIsVersionA5(getDeviceType(node)) {
		hwlog.RunLog.Debug("version is not npu")
		return ""
	}

	if api.IsA5InferServer(acceleratorType) { // A5 infer server return origin get rackID
		hwlog.RunLog.Debugf("rack id is: %s", rackID)
		return rackID
	}

	rackIDVal, err := strconv.Atoi(rackID)
	if err != nil {
		hwlog.RunLog.Errorf("rackId convert string to int failed, err %v", err)
		return ""
	}
	if rackIDVal < rackIdMin || rackIDVal > rackIdA5Max {
		hwlog.RunLog.Errorf("rackId %d out of range", rackIDVal)
		return ""
	}
	return rackID
}

func getNodeDeviceA5(
	baseDevInfos map[string]*api.NpuBaseInfo, nodeName, devType, serverIndex, acceleratorType string) *api.NodeDevice {
	if baseDevInfos == nil {
		return nil
	}
	nodeDevice := &api.NodeDevice{
		NodeName:   nodeName,
		ServerID:   serverIndex,
		ServerType: devType,
		DeviceMap:  make(map[string]string, len(baseDevInfos)),
		// npu vnic info for A5
		NpuInfoMap:      make(map[string]*api.NpuInfo, len(baseDevInfos)),
		AcceleratorType: acceleratorType,
	}
	for device, info := range baseDevInfos {
		physicID := strings.TrimPrefix(device, api.AscendMinuxPrefix)
		_, err := strconv.Atoi(physicID)
		if err != nil {
			hwlog.RunLog.Warnf("illegal device name, deviceName=%s, nodeName=%s",
				device, nodeName)
			return nil
		}
		superDeviceID := strconv.FormatUint(uint64(info.SuperDeviceID), formatBase)
		nodeDevice.DeviceMap[physicID] = superDeviceID
		var levelList []api.LevelElement
		for _, levels := range info.LevelList {
			for _, each := range levels.Info {
				levelList = append(levelList, each)
			}
		}
		nodeDevice.NpuInfoMap[physicID] = &api.NpuInfo{
			PhyId:     physicID,
			LevelList: levelList,
		}
	}
	return nodeDevice
}
