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
	"time"

	"k8s.io/klog/v2"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/cache"
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
	nodeTop = tp.removeBackoffCards(node, nodeTop)
	if !disFlag || !tp.IsNetworkFaultAttention || tp.ParameterPlaneUnhealthyTolerance {
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

// removeBackoffCards removes every backoff card from the node's usable top:
// the cards the current job has errored (pod-error) while the job's pod-error
// tolerance window is still active, plus the node-level silent-fault and
// sub-healthy cards while the job's corresponding tolerance window is still
// active. Both backoffs are per job: other jobs are unaffected.
func (tp *NPUHandler) removeBackoffCards(node plugin.NPUNode, nodeTop []int) []int {
	if tp == nil || node.ChipTopo == nil || tp.NPUJob == nil {
		return nodeTop
	}
	backoffTop := make([]int, 0)
	now := time.Now().Unix()
	for card, faultTime := range cache.PodErrorCards(node.Name) {
		if util.TimeGateAllows(tp.NPUJob.ToleranceConfig.PodErrorToleranceAfterTime, now-faultTime) {
			continue
		}
		for id := range node.ChipIDsFromCardNames([]string{card}) {
			backoffTop = append(backoffTop, id)
		}
	}
	if !tp.subHealthyToleranceAllows() {
		for id := range node.ChipTopo.SubHealthChips(node.Name) {
			backoffTop = append(backoffTop, id)
		}
	}
	if !tp.silentToleranceAllows() {
		for id := range node.ChipTopo.SilentFaultChips() {
			backoffTop = append(backoffTop, id)
		}
	}
	if len(backoffTop) == 0 {
		return nodeTop
	}
	klog.V(util.LogWarningLev).Infof("%s remove backoff cards %v from nodeTop %v",
		tp.GetPluginName(), backoffTop, nodeTop)
	return util.RemoveCommonElement(nodeTop, backoffTop)
}

// subHealthyToleranceAllows reports whether the job's sub-healthy tolerance has
// released the sub-healthy resources (true) or they are still withheld (false).
func (tp *NPUHandler) subHealthyToleranceAllows() bool {
	if tp == nil || tp.NPUJob == nil {
		return true
	}
	elapsed, ok := cache.SessionElapsed(tp.Name)
	if !ok {
		klog.V(util.LogWarningLev).Infof("%s cannot get job %s elapsed time", tp.GetPluginName(), tp.Name)
		return true
	}
	return util.TimeGateAllows(tp.NPUJob.ToleranceConfig.SubHealthyToleranceAfterTime, elapsed)
}

// silentToleranceAllows reports whether the job's silent-fault tolerance has
// released the silent-fault resources (true) or they are still withheld (false).
func (tp *NPUHandler) silentToleranceAllows() bool {
	if tp == nil || tp.NPUJob == nil {
		return true
	}
	elapsed, ok := cache.SessionElapsed(tp.Name)
	if !ok {
		klog.V(util.LogWarningLev).Infof("%s cannot get job %s elapsed time", tp.GetPluginName(), tp.Name)
		return true
	}
	return util.TimeGateAllows(tp.NPUJob.ToleranceConfig.SilentUnhealthyToleranceAfterTime, elapsed)
}

func (tp *NPUHandler) getNetUnhealthyNPU(node plugin.NPUNode) ([]int, error) {
	var annokey string
	if tp.ReqNPUName == "" {
		err := fmt.Errorf("reqNPUName is empty")
		klog.V(util.LogWarningLev).Infof("%s getNetUnhealthyNPU err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	if tp.ReqNPUName == util.NPUCardName {
		annokey = networkUnhealthyNPU
	} else {
		annokey = networkUnhealthy910
	}
	networkUnhealthyTopStr, ok := node.Annotation[annokey]
	if !ok {
		err := fmt.Errorf("node<%s> don't have resource<%s>", node.Name, annokey)
		klog.V(util.LogWarningLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	netUnhealthyTop := util.ChangeTopToIntArray(networkUnhealthyTopStr, tp.GetAnnoPreVal(tp.ReqNPUName))
	return netUnhealthyTop, nil
}

// GetNetUnhealthyNPUKey get network unhealthy npu annotation key
func (tp *NPUHandler) GetNetUnhealthyNPUKey() string {
	if tp.ReqNPUName == util.NPUCardName {
		return networkUnhealthyNPU
	}
	return networkUnhealthy910
}

func (tp *NPUHandler) getUnhealthyNPU(node plugin.NPUNode) []int {
	var annokey string
	if tp.ReqNPUName == "" {
		err := fmt.Errorf("reqNPUName is empty")
		klog.V(util.LogWarningLev).Infof("%s getUnhealthyNPU err: %s", tp.GetPluginName(), err.Error())
		return nil
	}
	if tp.ReqNPUName == util.NPUCardName {
		annokey = unHealthyNPU
	} else {
		annokey = unHealthy910
	}
	unhealthyTopStr, ok := node.Annotation[annokey]
	if !ok {
		klog.V(util.LogDebugLev).Infof("node<%s> don't have resource<%s>", node.Name, annokey)
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
