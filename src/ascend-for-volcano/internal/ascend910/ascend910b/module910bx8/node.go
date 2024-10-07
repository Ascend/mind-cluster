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
Package module910bx8 is using for HuaWei Ascend910Bx8 pin affinity schedule.
*/
package module910bx8

import (
	"fmt"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (tp *module910bx8) getUsableTopFromNode(node plugin.NPUNode, disFlag bool) ([]int, error) {
	nodeTop, err := tp.GetUsableTopFromNode(node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getUsableTopFromNode err: %s", err)
		return nil, err
	}
	if len(nodeTop) > tp.MaxNodeNPUNum {
		err := fmt.Errorf("node<%s> npu nodeTop top<%v> is invalid", node.Name, nodeTop)
		klog.V(util.LogWarningLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	if !disFlag {
		return nodeTop, nil
	}
	networkUnhealthyTopStr, ok := node.Annotation[tp.netUnhealthyKey]
	if !ok {
		err := fmt.Errorf("node<%s> don't have resource<%s>", node.Name, tp.netUnhealthyKey)
		klog.V(util.LogWarningLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	networkUnhealthyTop := util.ChangeTopToIntArray(networkUnhealthyTopStr, tp.GetAnnoPreVal())
	if len(networkUnhealthyTop) > tp.MaxNodeNPUNum {
		err := fmt.Errorf("node<%s> npu networkUnhealthy top<%v> is invalid", node.Name, networkUnhealthyTop)
		klog.V(util.LogWarningLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	res := util.RemoveCommonElement(nodeTop, networkUnhealthyTop)
	return res, nil
}

func (tp *module910bx8) getNodeBestScore(taskNPUNum int, npuTop []int) (int, error) {
	if taskNPUNum < 1 || taskNPUNum > nodeNPUNumber {
		return 0, fmt.Errorf("task req npu num<%d> is invalid", taskNPUNum)
	}
	npuNum := len(npuTop)
	if npuNum < 1 || npuNum > tp.MaxNodeNPUNum {
		return 0, fmt.Errorf("node npu num<%d> is invalid", npuNum)
	}
	return tp.AffScoreList[taskNPUNum-1][npuNum-1], nil
}
