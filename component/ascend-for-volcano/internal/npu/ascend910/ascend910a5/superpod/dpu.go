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

// Package superpod is using for Huawei Ascend pin affinity schedule
package superpod

import (
	"strconv"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func filterDpuFault(npuCardIdList []int, node plugin.NPUNode) []int {
	if len(node.DpuInfo.DpuList) == 0 {
		klog.V(util.LogDebugLev).Infof("%s cannot find dpu info, do not filter dpu devices for node<%s>",
			util.DpuLogPrefix, node.Name)
		return npuCardIdList
	}
	klog.V(util.LogInfoLev).Infof("%s filter dpu fault for node<%s> from npu lists: %v", util.DpuLogPrefix, node.Name,
		npuCardIdList)

	npuToDpuMap := node.DpuInfo.NpuToDpusMap
	dpuOperstateMap := make(map[string]string, util.DpuMaxNum)
	for _, dpu := range node.DpuInfo.DpuList {
		dpuOperstateMap[dpu.Name] = dpu.Operstate
	}

	resList := filterNpuByDpu(npuCardIdList, npuToDpuMap, dpuOperstateMap)
	klog.V(util.LogInfoLev).Infof("%s node<%s> use npu: %v", util.DpuLogPrefix, node.Name, resList)
	return resList
}

func filterNpuByDpu(npuCardIdList []int, npuToDpuMap map[string][]string, dpuOperstateMap map[string]string) []int {
	var resList []int
	for _, cardId := range npuCardIdList {
		position := cardId % nodeNPUNum
		positionStr := strconv.Itoa(position)
		if _, ok := npuToDpuMap[positionStr]; !ok {
			continue
		}
		// PCIe: One NPU has only one DPU, so the for loop only iterates once.
		// UB: One NPU has two DPUs, but as long as one DPU is active, it's fine.
		for _, dpu := range npuToDpuMap[positionStr] {
			if dpuOperstateMap[dpu] == util.ActiveStatus {
				resList = append(resList, cardId)
				break
			}
		}
	}
	return resList
}
