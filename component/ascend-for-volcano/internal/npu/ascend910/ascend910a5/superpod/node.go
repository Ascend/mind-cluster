/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package superpod is using for HuaWei Ascend pin affinity schedule.
*/
package superpod

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// acquire usable node top
func (tp *module910a5SuperPod) getUsableTopFromNode(node plugin.NPUNode) ([]int, error) {
	if tp == nil || len(node.Annotation) == 0 {
		return nil, errors.New(util.ArgumentError)
	}

	topStr, ok := node.Annotation[tp.GetAnnoName()]
	if !ok || len(topStr) == 0 {
		return nil, fmt.Errorf("getUsableTopFromNode %s don't have %s", node.Name, tp.GetAnnoName())
	}

	nodeTop := util.ChangeTopToIntArray(topStr, tp.GetAnnoPreVal())
	// the max card num of one node is 8
	if len(nodeTop) > npuNumber8 {
		err := fmt.Errorf("node npu num is invalid, and the npus index: %v", nodeTop)
		klog.V(util.LogErrorLev).Infof(getNPUFromPodFailedPattern, tp.GetPluginName(), err.Error())
		return nil, err
	}

	networkUnhealthyTopStr, ok := node.Annotation[tp.netUnhealthyKey]
	if !ok {
		err := fmt.Errorf("node<%s> don't have resource<%s>", node.Name, tp.netUnhealthyKey)
		klog.V(util.LogErrorLev).Infof(getNPUFromPodFailedPattern, tp.GetPluginName(), err.Error())
		return nil, err
	}
	networkUnhealthyTop := util.ChangeTopToIntArray(networkUnhealthyTopStr, tp.GetAnnoPreVal())
	if len(networkUnhealthyTop) > tp.MaxNodeNPUNum {
		err := fmt.Errorf("node<%s> npu networkUnhealthy top<%v> is invalid", node.Name, networkUnhealthyTop)
		klog.V(util.LogErrorLev).Infof(getNPUFromPodFailedPattern, tp.GetPluginName(), err.Error())
		return nil, err
	}

	res := util.RemoveCommonElement(nodeTop, networkUnhealthyTop)
	// print logs to record the usable npu numbers when it's not equal to 8
	if len(res) != npuNumber8 {
		klog.V(util.LogInfoLev).Infof("the len of the final usable npus in the node<%s> is %d", node.Name, len(res))
	}
	nodeNPUTopology := filterDpuFault(res, node)
	return nodeNPUTopology, nil
}

func (tp *module910a5SuperPod) checkNodeStaticParams(_ *api.TaskInfo, node plugin.NPUNode) error {
	// node in super-pod has super-podID which is not less than 0
	if node.SuperPodID < 0 {
		klog.V(util.LogErrorLev).Infof("node super-pod-id is invalid for node=%s, id=%d", node.Name, node.SuperPodID)
		return fmt.Errorf("the super-pod-id of node is invalid for node=%s, id=%d",
			node.Name, node.SuperPodID)
	}

	if node.RackID < 0 {
		klog.V(util.LogErrorLev).Infof("node rack-id is invalid for node=%s, id=%d", node.Name, node.RackID)
		return fmt.Errorf("node rack-id is invalid for node=%s, id=%d", node.Name, node.RackID)
	}
	return nil
}

func (tp *module910a5SuperPod) checkNodeNPUNums(task *api.TaskInfo, node plugin.NPUNode) error {
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	nodeTop, err := tp.getUsableTopFromNode(node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	if err = tp.NPUHandler.JudgeNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogErrorLev).Infof("%s JudgeNodeAndTaskNPU err: %s", task.Name, err.Error())
		return fmt.Errorf("checkNodeNPUByTask %s err: %s", util.NodeNotMeetTopologyWarning, err.Error())
	}
	return nil
}

// get the [8][8]bool npu using top of one rack nodes
func (tp *module910a5SuperPod) getRackNPUTop(nodes []nodeBaseInfo) rackNpuTopType {
	var rackNPUTop rackNpuTopType
	if len(nodes) < 1 {
		return rackNPUTop
	}
	if len(nodes) > rackNodeNum {
		klog.V(util.LogErrorLev).Infof("one rack<%d> max nodes num=8,but recived nodes num=%d",
			nodes[0].rackID, len(nodes))
		return rackNPUTop
	}
	for index, node := range nodes {
		nodeTop, err := tp.getUsableTopFromNode(tp.Nodes[node.name])
		if err != nil {
			klog.V(util.LogErrorLev).Infof("the node<%s> getUsableTopFromNode err: %s in getRackNPUTop",
				node.name, err.Error())
			return rackNPUTop
		}
		rackNPUTop[index] = getUsableNPUIndex(nodeTop)
	}
	for i := len(nodes); i < rackNodeNum; i++ {
		rackNPUTop[i] = [nodeNPUNum]bool{}
	}
	klog.V(util.LogInfoLev).Infof("the npu topo of rack<%d> is %v", nodes[0].rackID, rackNPUTop)
	return rackNPUTop
}
