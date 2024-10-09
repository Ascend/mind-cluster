/*
Copyright(C)2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package half910x4 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.
*/
package half910x4

import (
	"fmt"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (tp *half910x4) getUsableTopFromNode(node plugin.NPUNode, disFlag bool) ([]int, error) {
	nodeTop, err := tp.GetUsableTopFromNode(node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getUsableTopFromNode err: %s", err)
		return nil, err
	}

	if !disFlag {
		return nodeTop, nil
	}

	netUnhealthyTop, err := tp.getNetUnhealthyNPU(node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getNetUnhealthyNPU err: %s", err)
		return nil, err
	}

	res := util.RemoveCommonElement(nodeTop, netUnhealthyTop)
	return res, nil
}

func (tp *half910x4) getNetUnhealthyNPU(node plugin.NPUNode) ([]int, error) {
	networkUnhealthyTopStr, ok := node.Annotation[tp.netUnhealthyKey]
	if !ok {
		err := fmt.Errorf("node<%s> don't have resource<%s>", node.Name, tp.netUnhealthyKey)
		klog.V(util.LogWarningLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	netUnhealthyTop := util.ChangeTopToIntArray(networkUnhealthyTopStr, tp.GetAnnoPreVal())
	return netUnhealthyTop, nil
}

func (tp *half910x4) getNodeBestScore(taskNPUNum int, npuTop []int) (int, error) {
	if taskNPUNum < 1 || taskNPUNum > npuNumPerHccs {
		return 0, fmt.Errorf("task req npu num<%d> is invalid", taskNPUNum)
	}
	npuNum := len(npuTop)
	if npuNum < 1 || npuNum > tp.MaxNodeNPUNum {
		return 0, fmt.Errorf("node npu num<%d> is invalid", npuNum)
	}
	return tp.affScoreList[taskNPUNum-1][npuNum-1], nil
}
