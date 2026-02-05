/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package chip4nodex for 300I server scheduling
package chip4nodex

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// New return npu plugin for chip4nodex including 300I-A5-4p-8 or 300I-A5-4p-16 with 4p mesh
func New(name string) base.AscendHandler {
	m := &chip4nodex{}
	klog.V(util.LogInfoLev).Infof("chip4nodex card type =%s", name)
	num := getNPUNumByHandler(name)
	m.SetMaxNodeNPUNum(num)
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.affScoreList = createAffScoreList(m.MaxNodeNPUNum)
	return m
}

// ValidNPUJob Verify whether the NPU's request for the task is valid
func (tp *chip4nodex) ValidNPUJob() *api.ValidateResult {
	return tp.validNPUJob()
}

// CheckNodeNPUByTask Check whether the current node can meet the task's NPU resource requirements
func (tp *chip4nodex) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err)
		return err
	}
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	nodeTop, err := tp.GetUsableTopFromNode(node, tp.NPUTaskNum > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	if err = tp.judgeNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogErrorLev).Infof("the node judgeNodeAndTaskNPU failed, node name %s, err: %s",
			node.Name, err.Error())
		return fmt.Errorf("checkNodeNPUByTask %s err: %s", util.NodeNotMeetTopologyWarning, err.Error())
	}
	return nil
}

// ScoreBestNPUNodes According to the task requirements and the available NPU resources at each node,
// score and rank the candidate nodes, with the scores stored in sMap.
func (tp *chip4nodex) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, sMap map[string]float64) error {
	if tp == nil || task == nil || len(nodes) == 0 || len(sMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %s.", err)
		return err
	}
	taskNPUNum, getErr := tp.GetTaskReqNPUNum(task)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum %s: %s", tp.GetPluginName(), task.Name, getErr)
		return getErr
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
		cardIds, err := tp.GetUsableTopFromNode(nNode, tp.NPUTaskNum > 1)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes getErr: %s", tp.GetPluginName(), err)
			continue
		}

		if is4PmeshAffinity(taskNPUNum) {
			sMap[node.Name] = tp.scoreNodeFor4Pmesh(taskNPUNum, cardIds)
			continue
		}
		sMap[node.Name] = tp.scoreNodeForGeneral(taskNPUNum, cardIds)
	}
	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes task<%s> sMap<%v>", tp.GetPluginName(),
		task.Name, sMap)
	return nil
}

// UseAnnotation Select NPU resources for the task from the specified node and update the node information
func (tp *chip4nodex) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("UseAnnotation %s.", err)
		return nil
	}
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation task<%s> node<%s> resource<%s> Annotation: %s",
		tp.GetPluginName(), task.Name, node.Name, tp.GetAnnoName(), util.SafePrint(node.Annotation))
	selectedNPU, err := tp.selectNPUFromNode(task, node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation err:%s.", tp.GetPluginName(), err)
		return nil
	}
	klog.V(util.LogInfoLev).Infof("%s UseAnnotation %s select %v.", tp.GetPluginName(), task.Name, selectedNPU)
	// Write the selected NPU topology data into the annotations of the Pod where the task is located
	tp.SetNPUTopologyToPodFn(task, selectedNPU, node)
	// Return the updated plugin.NPUNode structure
	newNode := tp.UpdateNodeInfo(node, selectedNPU)
	return newNode
}

// selectNPUFromNode Select the NPU resources that meet the task requirements from the specified nodes
func (tp *chip4nodex) selectNPUFromNode(task *api.TaskInfo, node plugin.NPUNode) ([]int, error) {
	// Get the number of NPUs required for the task
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	// Get the currently available NPU topology information on the node
	nodeTop, err := tp.GetUsableTopFromNode(node, tp.NPUTaskNum > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	if len(nodeTop) < taskNPUNum {
		err = fmt.Errorf("node<%s> top<%v> can not meet task req<%d>", node.Name, len(nodeTop), taskNPUNum)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes err: %s", err)
		return nil, err
	}
	if is4PmeshAffinity(taskNPUNum) {
		return tp.selectNPUIn4Pmesh(taskNPUNum, nodeTop), nil
	}
	return nodeTop[:taskNPUNum], nil
}

// ReleaseAnnotation Used to release allocated resources
func (tp *chip4nodex) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
