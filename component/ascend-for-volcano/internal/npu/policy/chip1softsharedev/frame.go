/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package chip1softsharedev is using for HuaWei chip1softsharedev schedule.
package chip1softsharedev

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// New return npu plugin
func New(name string) base.AscendHandler {
	m := &chip1softsharedev{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetMaxNodeNPUNum(nodeNPUNumber)
	m.netUnhealthyKey = networkUnhealthyNPU
	return m
}

// ValidNPUJob check job req npu num and mode
func (tp *chip1softsharedev) ValidNPUJob() *api.ValidateResult { return tp.validNPUJob() }

// CheckNodeNPUByTask check nod npu meet task req
func (tp *chip1softsharedev) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err)
		return err
	}
	softShareDevEnable, ok := node.Label[util.SchedulerSoftShareDevEnableNodeLabel]
	if !ok || softShareDevEnable != "true" {
		err := fmt.Errorf("node %s label %s not exist", node.Name, util.SchedulerSoftShareDevEnableNodeLabel)
		klog.V(util.LogDebugLev).Infof("CheckNodeNPUByTask err: %s", err)
		return err
	}
	reqResource, err := tp.getSoftShareDevResource()
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s getSoftShareDevResource err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	nodeUsedResourceMap := tp.getUsedResourceMapFromNodeTasks(node.Tasks)
	nodeTop, err := tp.GetUsableTopFromNode(node, false)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	chipMemory, err := tp.getChipMemoryFromNodeLabel(node.Label)
	if err != nil {
		return err
	}
	for _, cardIdx := range nodeTop {
		used, exists := nodeUsedResourceMap[cardIdx]
		if !exists {
			return nil
		}
		if used.aicoreQuota+reqResource.aicoreQuota <= util.MaxAicoreQuota &&
			used.hbmQuota+reqResource.hbmQuota <= chipMemory &&
			used.schedulingPolicy == reqResource.schedulingPolicy {
			return nil
		}
	}
	return fmt.Errorf("npu topology not meet share job: %s require", task.Name)
}

// ScoreBestNPUNodes score node by calculate task req npu num and node npu top
func (tp *chip1softsharedev) ScoreBestNPUNodes(
	task *api.TaskInfo, nodes []*api.NodeInfo, sMap map[string]float64) error {
	if tp == nil || task == nil || len(sMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %s.", err)
		return err
	}
	reqResource, err := tp.getSoftShareDevResource()
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s getSoftShareDevResource %s: %s", tp.GetPluginName(), task.Name, err)
		return err
	}
	for _, node := range nodes {
		if node == nil {
			continue
		}
		nNode, ok := tp.Nodes[node.Name]
		if !ok {
			klog.V(util.LogWarningLev).Infof("%s %s ScoreBestNPUNodes %s is not npu node",
				tp.GetPluginName(), task.Name, node.Name)
			continue
		}
		// Get the list of NPUs currently available on the node
		cardIds, err := tp.GetUsableTopFromNode(nNode, false)
		if err != nil {
			klog.V(util.LogDebugLev).Infof("%s GetUsableTopFromNode err: %s", tp.GetPluginName(), err)
			continue
		}
		usedResourceMap := tp.getUsedResourceMapFromNodeTasks(nNode.Tasks)
		npuChipMemory, err := tp.getChipMemoryFromNodeLabel(node.Node.Labels)
		if err != nil {
			klog.V(util.LogDebugLev).Infof("%s getChipMemoryFromNodeLabel err: %s", tp.GetPluginName(), err)
			continue
		}
		bestScore := getBestScore(usedResourceMap, cardIds, reqResource, npuChipMemory)
		sMap[node.Name] = float64(bestScore * tp.MaxNodeNPUNum)
	}
	klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodes task<%s> sMap<%v>", tp.GetPluginName(),
		task.Name, sMap)
	return nil
}

// UseAnnotation select npu for task from node
func (tp *chip1softsharedev) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("UseAnnotation %s.", err)
		return nil
	}
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation task<%s> node<%s> resource<%s> Annotation: %s",
		tp.GetPluginName(), task.Name, node.Name, tp.GetAnnoName(tp.ReqNPUName), util.SafePrint(node.Annotation))
	selectedNPU, err := tp.selectNPUFromNode(task, node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("task %s UseAnnotation err:%s.", task.Name, err)
		return nil
	}
	klog.V(util.LogInfoLev).Infof("%s UseAnnotation %s select %v.", tp.GetPluginName(), task.Name, selectedNPU)
	tp.SetNPUTopologyToPodFn(task, selectedNPU, node)
	return tp.NPUHandler.UpdateNodeInfo(node, selectedNPU)
}

func (tp *chip1softsharedev) selectNPUFromNode(task *api.TaskInfo, node plugin.NPUNode) ([]int, error) {
	nodeTop, err := tp.GetUsableTopFromNode(node, false)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	if len(nodeTop) < 1 {
		return nil, fmt.Errorf("node<%s> top<%v> can not meet task req<%d>", node.Name, len(nodeTop), 1)
	}
	reqResource, err := tp.getSoftShareDevResource()
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s getSoftShareDevResource %s: %s", tp.GetPluginName(), task.Name, err)
		return nil, err
	}
	npuChipMemory, err := tp.getChipMemoryFromNodeLabel(node.Label)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s getChipMemoryFromNodeLabel %s: %s", tp.GetPluginName(), task.Name, err)
		return nil, err
	}
	nodeTopUsedResourceMap := tp.getUsedResourceMapFromNodeTasks(node.Tasks)
	requestResource := softShareDevResource{
		aicoreQuota:      reqResource.aicoreQuota,
		hbmQuota:         reqResource.hbmQuota,
		schedulingPolicy: reqResource.schedulingPolicy,
	}
	filterNodeTop := npuPrioritySort(nodeTop, nodeTopUsedResourceMap, requestResource, npuChipMemory)
	if len(filterNodeTop) < 1 {
		klog.V(util.LogDebugLev).Infof("node<%s> top<%v> can not meet task req<%d>",
			node.Name, len(filterNodeTop), 1)
		return nil, fmt.Errorf("node<%s> top<%v> can not meet task req<%d>", node.Name, len(filterNodeTop), 1)
	}
	return filterNodeTop, nil
}

// ReleaseAnnotation Release used resource.
func (tp *chip1softsharedev) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
