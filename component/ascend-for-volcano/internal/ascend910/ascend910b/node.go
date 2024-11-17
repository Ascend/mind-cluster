/*
Copyright(C)2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package ascend910b is using for HuaWei Ascend 910B pin affinity schedule.
*/
package ascend910b

import (
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// UpdateNodeInfo update node info
func (ab *Base910b) UpdateNodeInfo(node plugin.NPUNode, usedTop []int) *plugin.NPUNode {
	if ab == nil || len(node.Annotation) == 0 || len(usedTop) == 0 {
		return nil
	}
	if len(usedTop) > ab.MaxNodeNPUNum {
		klog.V(util.LogErrorLev).Infof("%s UpdateNodeInfo err: used npu num<%d> is invalid",
			ab.GetPluginName(), len(usedTop))
		return nil
	}
	klog.V(util.LogDebugLev).Infof("%s before UpdateNodeInfo node<%s> Annotation: %s",
		ab.GetPluginName(), node.Name, util.SafePrint(node.Annotation))
	healthyAnno, err := node.GetNewNPUNodeAnnotation(usedTop, ab.GetAnnoName(), ab.GetAnnoPreVal())
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UpdateNodeInfo err: %s", ab.GetPluginName(), err)
		return nil
	}
	node.Annotation[ab.GetAnnoName()] = healthyAnno
	klog.V(util.LogDebugLev).Infof("%s after UpdateNodeInfo node<%s> Annotation: %s",
		ab.GetPluginName(), node.Name, util.SafePrint(node.Annotation))
	return &node
}

func (ab *Base910b) GetNodeHccsArray(nodeTop []int) ([]int, []int) {
	var leftHccsArray []int
	var rightHccsArray []int

	idCutNum := ab.MaxNodeNPUNum / util.NPUIndex2
	if ab.MaxNodeNPUNum < util.NPUIndex8 {
		for _, v := range nodeTop {
			if v < util.NPUIndex4 {
				leftHccsArray = append(leftHccsArray, v)
				continue
			}
			rightHccsArray = append(rightHccsArray, v)
		}
		return leftHccsArray, rightHccsArray
	}
	for _, v := range nodeTop {
		if v < idCutNum {
			leftHccsArray = append(leftHccsArray, v)
			continue
		}
		rightHccsArray = append(rightHccsArray, v)
	}

	return leftHccsArray, rightHccsArray
}
