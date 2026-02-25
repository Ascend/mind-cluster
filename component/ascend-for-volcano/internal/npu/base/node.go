/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package base is using for HuaWei Ascend pin affinity schedule.
*/
package base

import (
	"errors"
	"fmt"
	"sort"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// GetUsableTopFromNode Get ascend node usable top.
func (tp *NPUHandler) GetUsableTopFromNode(node plugin.NPUNode, disFlag bool) ([]int, error) {
	if tp == nil || len(node.Annotation) == 0 {
		return nil, errors.New(util.ArgumentError)
	}
	topStr, ok := node.Annotation[tp.GetAnnoName(tp.ReqNPUName)]
	if !ok || len(topStr) == 0 {
		return nil, fmt.Errorf("getUsableTopFromNode don't have %s", tp.GetAnnoName(tp.ReqNPUName))
	}

	nodeTop := util.ChangeTopToIntArray(topStr, tp.GetAnnoPreVal(tp.ReqNPUName))
	if len(nodeTop) > tp.MaxNodeNPUNum {
		err := fmt.Errorf("node npu top<%v> is invalid", nodeTop)
		klog.V(util.LogDebugLev).Infof("%s GetUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	if !disFlag || !tp.IsNetworkFaultAttention {
		sort.Ints(nodeTop)
		return nodeTop, nil
	}
	// for distributed job, need to remove the net unhealthy npu
	netUnhealthyTop, err := tp.getNetUnhealthyNPU(node)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("getNetUnhealthyNPU err: %s", err)
		return nil, err
	}

	res := util.RemoveCommonElement(nodeTop, netUnhealthyTop)
	sort.Ints(res)
	return res, nil
}

func (tp *NPUHandler) getNetUnhealthyNPU(node plugin.NPUNode) ([]int, error) {
	networkUnhealthyTopStr, ok := node.Annotation[networkUnhealthyNPU]
	if !ok {
		err := fmt.Errorf("node<%s> don't have resource<%s>", node.Name, networkUnhealthyNPU)
		klog.V(util.LogWarningLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	netUnhealthyTop := util.ChangeTopToIntArray(networkUnhealthyTopStr, tp.GetAnnoPreVal(tp.ReqNPUName))
	return netUnhealthyTop, nil
}

func (tp *NPUHandler) getUnhealthyNPU(node plugin.NPUNode) []int {
	unhealthyTopStr, ok := node.Annotation[unHealthyNPU]
	if !ok {
		klog.V(util.LogDebugLev).Infof("node<%s> don't have resource<%s>", node.Name, unHealthyNPU)
		return make([]int, 0)
	}
	unhealthyTop := util.ChangeTopToIntArray(unhealthyTopStr, tp.GetAnnoPreVal(tp.ReqNPUName))
	return unhealthyTop
}

// GetCardNumGroupsFromTop get the chip for each card from nodeTop
func (tp *NPUHandler) GetCardNumGroupsFromTop(nodeNPUTopology []int) [][]int {
	if tp == nil || tp.MaxCardNPUNum == 0 {
		return nil
	}
	maxCardNum := 0
	for _, v := range nodeNPUTopology {
		maxCardNum = util.Max(maxCardNum, v)
	}
	cardNumGroups := make([][]int, maxCardNum/tp.MaxCardNPUNum+1)
	for _, v := range nodeNPUTopology {
		index := v / tp.MaxCardNPUNum
		if index > len(cardNumGroups)-1 {
			continue
		}
		cardNumGroups[index] = append(cardNumGroups[index], v)
	}
	return cardNumGroups
}

// UpdateNodeInfo update node info
func (tp *NPUHandler) UpdateNodeInfo(node plugin.NPUNode, usedTop []int) *plugin.NPUNode {
	if tp == nil || len(usedTop) > tp.MaxNodeNPUNum {
		klog.V(util.LogErrorLev).Infof("NPUHandler is <%#v> or UpdateNodeInfo err: used npu num<%d> is invalid",
			tp, len(usedTop))
		return nil
	}
	klog.V(util.LogDebugLev).Infof("%s before UpdateNodeInfo node<%s> Annotation: %s",
		tp.GetPluginName(), node.Name, util.SafePrint(node.Annotation))
	healthyAnno, err := node.GetNewNPUNodeAnnotation(usedTop, tp.GetAnnoName(tp.ReqNPUName), tp.GetAnnoPreVal(tp.ReqNPUName))
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UpdateNodeInfo err: %s", tp.GetPluginName(), err.Error())
		return nil
	}
	node.Annotation[tp.GetAnnoName(tp.ReqNPUName)] = healthyAnno
	klog.V(util.LogDebugLev).Infof("%s after UpdateNodeInfo node<%s> Annotation: %s",
		tp.GetPluginName(), node.Name, util.SafePrint(node.Annotation))
	return &node
}
