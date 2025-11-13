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

// Package rescheduling is using for Huawei Ascend pin fault rescheduling
package rescheduling

import (
	"strconv"
	"strings"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// The return value indicates whether the task is faulted
func (reScheduler *ReScheduler) getTaskHealthStateByNodeDpu(fTask *FaultTask) bool {
	// fTask.UseCardName - The card used by the pod.
	// reScheduler.FaultNodes[fTask.NodeName].dpuCMInfo - DPU status
	faultNode, ok := reScheduler.FaultNodes[fTask.NodeName]
	if !ok {
		klog.V(util.LogErrorLev).Infof("%s task<%s> get fault node<%s> failed, reScheduler.FaultNodes=%+v",
			util.DpuLogPrefix, fTask.TaskName, fTask.NodeName, reScheduler.FaultNodes)
		return false
	}
	dpuCMInfo := faultNode.dpuCMInfo
	klog.V(util.LogInfoLev).Infof("%s task<%s> use card: %v, dpu: %v", util.DpuLogPrefix, fTask.TaskName,
		fTask.UseCardName, dpuCMInfo)
	npuToDpuMap := dpuCMInfo.NpuToDpusMap
	if len(npuToDpuMap) == util.EmptyNPUToDPUMapLen {
		klog.V(util.LogDebugLev).Infof("%s task<%s> get node<%s> npu to dpu map failed:%v", util.DpuLogPrefix,
			fTask.TaskName, fTask.NodeName, npuToDpuMap)
		return false
	}
	dpuOperstateMap := make(map[string]string, util.DpuMaxNum)
	for _, dpu := range dpuCMInfo.DpuList {
		dpuOperstateMap[dpu.Name] = dpu.Operstate
	}
	return checkIsTaskFaultByDpu(npuToDpuMap, fTask, dpuOperstateMap)
}

func checkIsTaskFaultByDpu(npuToDpuMap map[string][]string, fTask *FaultTask, dpuOperstateMap map[string]string) bool {
	for npu, dpus := range npuToDpuMap {
		if !isNpuBeUsed(fTask.UseCardName, npu) {
			continue
		}
		isNpuFault := true
		// PCIe: One NPU has only one DPU, so the for loop only iterates once.
		// UB: One NPU has two DPUs, but as long as one DPU is active, it's fine.
		for _, dpu := range dpus {
			if dpuOperstateMap[dpu] == util.ActiveStatus {
				isNpuFault = false
				break
			}
		}
		if isNpuFault {
			klog.V(util.LogDebugLev).Infof("%s for npu<%s> dpu<%v> are faulty", util.DpuLogPrefix, npu, dpus)
			return true
		}
	}
	return false
}

func isNpuBeUsed(strs []string, target string) bool {
	for _, str := range strs {
		parts := strings.Split(str, "-")
		if len(parts) <= 1 {
			klog.V(util.LogErrorLev).Infof("get id by spliting npu card failed")
			return false
		}
		cardId, err := strconv.Atoi(parts[len(parts)-1])
		if err != nil {
			klog.V(util.LogErrorLev).Infof("get card id failed:%s", err)
			return false
		}
		cardIdStr := strconv.Itoa(cardId % util.NpuCountPerNode)
		if cardIdStr == target {
			return true
		}
	}
	return false
}
