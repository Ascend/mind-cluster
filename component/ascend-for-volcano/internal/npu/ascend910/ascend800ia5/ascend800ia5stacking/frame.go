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

// Package ascend800ia5stacking provides NPU affinity scheduling for Huawei Ascend800i-A5 stacking architecture.
package ascend800ia5stacking

import (
	"errors"
	"fmt"
	"reflect"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// New creates and returns a new NPU plugin instance
func New(name string) base.AscendHandler {
	m := &module800ia5stacking{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetMaxNodeNPUNum(nodeNPUNumber)
	m.netUnhealthyKey = networkUnhealthyNPU
	m.SuperPodCache = make(map[int32][]plugin.NPUNode)
	m.NPUSelectedCache = make(map[api.JobID]map[int32][]int)
	m.PickedNodeCache = make(map[string]bool)
	// Score for unsatisfied requirements
	AffUnableScore := util.AffScore8 + unSchedulerOffset
	// Rows represent the number of NPUs required by the task, columns represent the remaining NPUs on the node
	m.AffScoreList = [][]int{
		{util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3, util.AffScore4, util.AffScore5,
			util.AffScore6, util.AffScore7},
		{AffUnableScore, util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3, util.AffScore4,
			util.AffScore5, util.AffScore6},
		{AffUnableScore, AffUnableScore, util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3,
			util.AffScore4, util.AffScore5},
		{AffUnableScore, AffUnableScore, AffUnableScore, util.AffScore0, util.AffScore1, util.AffScore2,
			util.AffScore3, util.AffScore4},
		{AffUnableScore, AffUnableScore, AffUnableScore, AffUnableScore, util.AffScore0, util.AffScore1,
			util.AffScore2, util.AffScore3},
		{AffUnableScore, AffUnableScore, AffUnableScore, AffUnableScore, AffUnableScore, util.AffScore0,
			util.AffScore1, util.AffScore2},
		{AffUnableScore, AffUnableScore, AffUnableScore, AffUnableScore, AffUnableScore, AffUnableScore,
			util.AffScore0, util.AffScore1},
		{AffUnableScore, AffUnableScore, AffUnableScore, AffUnableScore, AffUnableScore, AffUnableScore,
			AffUnableScore, util.AffScore0},
	}
	return m
}

// PreStartAction prepares the stacking NPU nodes before the scheduling session starts.
// It groups NPU nodes by SuperPodID and stores them in SuperPodCache for efficient access.
// Finally, it calls the underlying NPUHandler's PreStartAction method.
func (tp *module800ia5stacking) PreStartAction(ssn *framework.Session) error {
	for _, npuNode := range tp.NPUHandler.Nodes {
		if stackingNpuList, exists := tp.SuperPodCache[npuNode.SuperPodID]; !exists {
			tp.SuperPodCache[npuNode.SuperPodID] = []plugin.NPUNode{npuNode}
		} else {
			stackingNpuList = append(stackingNpuList, npuNode)
			tp.SuperPodCache[npuNode.SuperPodID] = stackingNpuList
		}

	}
	return nil
}

// ValidNPUJob validates if the task's NPU request is valid
func (tp *module800ia5stacking) ValidNPUJob() *api.ValidateResult {
	return tp.Valid800ia5NPUJob()
}

// CheckNodeNPUByTask checks if the current node can meet the task's NPU resource requirements
func (tp *module800ia5stacking) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err)
		return err
	}
	nodeAcceleratorType := node.Label[util.AcceleratorType]
	klog.V(util.LogInfoLev).Infof("node %s: accelerator-type=%s", node.Name, nodeAcceleratorType)
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	// Get available NPU resource topology information on the current node
	nodeTop, err := tp.getUsableTopFromNode(node, tp.NPUTaskNum > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	// Determine if the node and task's NPU configurations match
	if err = tp.JudgeNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogErrorLev).Infof("%s JudgeNodeAndTaskNPU err: %s", tp.GetPluginName(), err.Error())
		return fmt.Errorf("npu topology not meet job require,network unhealthy card is [ %s ]",
			node.Annotation[tp.netUnhealthyKey])
	}

	return nil
}

// ScoreBestNPUNodes scores and ranks candidate nodes based on task requirements and available NPU resources, with scores stored in sMap
func (tp *module800ia5stacking) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo,
	sMap map[string]float64) error {
	if tp == nil || task == nil || len(nodes) == 0 || len(sMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %s.", err)
		return err
	}
	tp.initSelectedCardCache(task)
	tp.initSelectedNodeCache(nodes)
	taskNPUNum, getErr := tp.GetTaskReqNPUNum(task)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum %s: %s", tp.GetPluginName(), task.Name, getErr)
		return getErr
	}
	for _, node := range nodes {
		if reflect.ValueOf(node).IsNil() {
			continue
		}
		nNode, ok := tp.Nodes[node.Name]
		if !ok {
			klog.V(util.LogWarningLev).Infof("%s %s ScoreBestNPUNodes %s is not npu node",
				tp.GetPluginName(), task.Name, node.Name)
			continue
		}
		// Get the number of y-axis aligned cards and available cards
		stackNpus, cardIds := tp.getTheSameSuperPodIdNodeNpuNum(nNode)

		//For the current nNode, if the corresponding stack has cards occupied by the current task (formercardids),
		// and there is overlap with cardids (indicating it's not fully occupied), then this node should be prioritized for allocation.
		// The priority level is determined by the intersection of cardids and formercardids
		bestScore, err := tp.getNodeBestScore(task, nNode, cardIds, taskNPUNum, stackNpus)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s getNodeBestScoreInStack getErr: %s", tp.GetPluginName(), err)
			continue
		}
		healthyNPUNum := tp.getHealthyNpu(nNode.SuperPodID)
		sMap[node.Name] = float64(int(healthyNPUNum/util.NPUHexKilo)*stackOffsetPower - bestScore)
	}
	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes task<%s> sMap<%v>", tp.GetPluginName(),
		task.Name, sMap)
	return nil
}

func (tp *module800ia5stacking) getNodeBestScore(task *api.TaskInfo, nNode plugin.NPUNode, cardIds []int,
	taskNPUNum int, stackNpus []int) (int, error) {
	effectScore := 0
	for _, node := range tp.SuperPodCache[nNode.SuperPodID] {
		if _, exists := tp.PickedNodeCache[node.Name]; !exists {
			// This node is neither in the selected nodes nor in the selectable nodes (e.g., cordoned)
			effectScore = unSchedulerScore
		}
	}
	if taskUsed, exists := tp.NPUSelectedCache[task.Job]; exists {
		if formerSelected, formerExists := taskUsed[nNode.SuperPodID]; formerExists {
			if potentialIds, isDuplicated := util.HasCommonElement(formerSelected, cardIds); isDuplicated {
				bScore, err := tp.getBestScoreInComplicatedStack(taskNPUNum, len(potentialIds), len(cardIds))
				return bScore + effectScore, err
			}
		}
	}
	bScore, err := tp.getNodeBestScoreInStack(taskNPUNum, len(stackNpus), len(cardIds))
	return bScore + effectScore, err
}

// UseAnnotation selects NPU resources for the task from the specified node and updates node information
func (tp *module800ia5stacking) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
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

	// Write the selected NPU topology data into the annotation of the Pod where the task is located
	tp.SetNPUTopologyToPodFn(task, selectedNPU, node)
	// Write to cache
	if taskUsed, exists := tp.NPUSelectedCache[task.Job]; exists {
		if formerSelected, formerExists := taskUsed[node.SuperPodID]; formerExists {
			taskUsed[node.SuperPodID] = util.MergeUnique(formerSelected, selectedNPU)
		} else {
			taskUsed[node.SuperPodID] = selectedNPU
		}
	} else {
		tp.NPUSelectedCache[task.Job] = map[int32][]int{
			node.SuperPodID: selectedNPU,
		}
	}
	// Return the updated plugin.NPUNode structure
	newNode := tp.UpdateNodeInfo(node, selectedNPU)
	return newNode
}

// selectNPUFromNode selects the required number of NPU resources from the specified node that meet the task requirements
func (tp *module800ia5stacking) selectNPUFromNode(task *api.TaskInfo, node plugin.NPUNode) ([]int, error) {
	// Get the number of NPUs required by the task
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	nodeTop, err := tp.getUsableTopFromNode(node, tp.NPUTaskNum > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	if taskUsed, exists := tp.NPUSelectedCache[task.Job]; exists {
		if formerSelected, formerExists := taskUsed[node.SuperPodID]; formerExists {
			if potentialIds, isDuplicated := util.HasCommonElement(formerSelected, nodeTop); isDuplicated &&
				len(potentialIds) >= taskNPUNum {
				return potentialIds[:taskNPUNum], nil
			}
		}
	}
	stackTop, err := tp.getUsableTopFromStack(node, tp.NPUTaskNum > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s getUsableTopFromStack err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	if len(stackTop) < taskNPUNum {
		if len(nodeTop) < taskNPUNum {
			err = fmt.Errorf("node<%s> top<%v> can not meet task req<%d>", node.Name, len(stackTop), taskNPUNum)
			klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes err: %s", err)
			return nil, err
		}
		return nodeTop[:taskNPUNum], nil
	}
	return stackTop[:taskNPUNum], nil
}

// ReleaseAnnotation is used to release allocated resources
func (tp *module800ia5stacking) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
