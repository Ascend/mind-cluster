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

// Package chip8node8ra64sp is using for Huawei Ascend pin affinity schedule
package chip8node8ra64sp

import (
	"fmt"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func (tp *chip8node8ra64sp) filterDpuFault(npuCardIdList []int, node plugin.NPUNode) []int {
	if len(node.Annotation) == 0 {
		klog.V(util.LogDebugLev).Infof("%s cannot find dpu info, do not filter dpu devices for node<%s>",
			util.DpuLogPrefix, node.Name)
		return npuCardIdList
	}
	klog.V(util.LogInfoLev).Infof("%s filter dpu fault for node<%s> from npu lists: %v", util.DpuLogPrefix, node.Name,
		npuCardIdList)

	dpuUnhealthyTopStr, ok := node.Annotation[tp.dpuUnhealthyKey]
	if !ok {
		err := fmt.Errorf("node<%s> don't have resource<%s>", node.Name, tp.dpuUnhealthyKey)
		klog.V(util.LogErrorLev).Infof(getNPUFromPodFailedPattern, tp.GetPluginName(), err.Error())
		return npuCardIdList
	}
	// Use a default value directly instead of relying on tp.ReqNPUName
	dpuUnhealthyTop := util.ChangeTopToIntArray(dpuUnhealthyTopStr, tp.GetAnnoPreVal(tp.ReqNPUName))
	if len(dpuUnhealthyTop) > tp.MaxNodeNPUNum {
		err := fmt.Errorf("node<%s> npu dpuUnhealthy top<%v> is invalid", node.Name, dpuUnhealthyTop)
		klog.V(util.LogErrorLev).Infof(getNPUFromPodFailedPattern, tp.GetPluginName(), err.Error())
		return npuCardIdList
	}

	res := util.RemoveCommonElement(npuCardIdList, dpuUnhealthyTop)
	// print logs to record the usable npu numbers when it's not equal to 8
	if len(res) != npuNumber8 {
		klog.V(util.LogInfoLev).Infof("the len of the final usable npus in the node<%s> is %d", node.Name, len(res))
	}

	klog.V(util.LogInfoLev).Infof("%s node<%s> use npu: %v", util.DpuLogPrefix, node.Name, res)
	return res
}
