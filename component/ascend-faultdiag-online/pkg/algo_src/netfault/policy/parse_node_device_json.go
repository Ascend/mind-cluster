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

// Package policy is used for processing superpod information
package policy

import (
	"fmt"
	"regexp"
	"strconv"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/netfault/algo"
)

const exetractWordIdLen = 2
const matchedNetStrMaxLen = 3
const netPlane0Constant = "netplane_0"

func parseNodeDeviceMap(nodeDeviceMap map[string]*NodeDevice) ([]string, map[string][]string) {
	if nodeDeviceMap == nil || len(nodeDeviceMap) == 0 {
		hwlog.RunLog.Errorf("nodeDeviceMap nil")
		return nil, nil
	}

	npuFullMeshInfo := make([]string, 0)
	npuNetPlaneInfo := make(map[string][]string)

	for workId, workInfo := range nodeDeviceMap {
		res := getCurWorkInfo(npuNetPlaneInfo, workInfo)
		if res == nil || len(res) == 0 {
			hwlog.RunLog.Errorf("work %s getCurWorkInfo error", workId)
		}
	}

	npuFullMeshInfo = append(npuFullMeshInfo, DiagVersionA3)
	return npuFullMeshInfo, npuNetPlaneInfo

}

func getCurWorkInfo(npuNetplaneInfo map[string][]string, workInfo *NodeDevice) map[string][]string {
	if npuNetplaneInfo == nil {
		return nil
	}

	if workInfo == nil {
		hwlog.RunLog.Error("work info is empty")
		return nil
	}

	if len(workInfo.DeviceMap) == 0 {
		hwlog.RunLog.Errorf("the DeviceMap of node %s is empty", workInfo.NodeName)
		return nil
	}

	if len(workInfo.ServerID) == 0 {
		hwlog.RunLog.Errorf("the ServerID of node %s is empty", workInfo.NodeName)
		return nil
	}

	for id, sdIdStr := range workInfo.DeviceMap {
		// A3场景worker抽象成rock用于适配算法解析
		npuNetplaneInfoStr := fmt.Sprintf("L2:0#Rack-%s:0#rack-%s.NSlot-0:0#NPU%s-%s:0",
			workInfo.ServerID, workInfo.ServerID, id, sdIdStr)
		npuNetplaneInfo[netPlane0Constant] = append(npuNetplaneInfo[netPlane0Constant], npuNetplaneInfoStr)
	}

	return npuNetplaneInfo
}

// ExtractNPUMapA3 提取NPU信息A3版本
func ExtractNPUMapA3(npuNetPlaneInfo map[string][]string) map[string]algo.NpuInfo {
	resultMap := make(map[string]algo.NpuInfo)
	netPlane0, ok := npuNetPlaneInfo[netPlane0Constant]
	if !ok {
		hwlog.RunLog.Errorf("npuNetPlaneInfo %s formatted error", netPlane0Constant)
		return nil
	}

	for _, npuNetPlaneInfoStr := range netPlane0 {
		sdId, npuInfo := extractNetStrInfo(npuNetPlaneInfoStr)
		if sdId == "" {
			hwlog.RunLog.Warnf("extractNetSet err")
			continue
		}
		resultMap[sdId] = npuInfo
	}

	return resultMap
}

func extractNetStrInfo(netStr string) (string, algo.NpuInfo) {
	npuInfo := algo.NpuInfo{}
	pattern := `#(.*?):`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(netStr, -1)
	if len(matches) < matchedNetStrMaxLen {
		hwlog.RunLog.Errorf("not enough matches for regex:%s", netStr)
		return "", algo.NpuInfo{}
	}
	npuInfo.RackName = matches[0][1]
	npuInfo.SlotName = matches[1][1]
	npuInfo.NetPlaneId = netPlane0Constant
	npuId := matches[2][1]

	exepectLen := 3 // 预期长度
	pattern = `NPU(.+?)-(\d+)`
	re = regexp.MustCompile(pattern)
	matches = re.FindAllStringSubmatch(npuId, -1)
	if len(matches) == 0 || len(matches[0]) < exepectLen {
		hwlog.RunLog.Errorf("not enough matches for regex:%s", npuId)
		return "", algo.NpuInfo{}
	}
	npuInfo.IP = matches[0][2]
	phyIdStr := matches[0][1]
	phyId, err := strconv.Atoi(phyIdStr)
	if err != nil {
		hwlog.RunLog.Error(err)
		return "", algo.NpuInfo{}
	}
	npuInfo.NpuNumber = phyId
	return npuInfo.IP, npuInfo

}
