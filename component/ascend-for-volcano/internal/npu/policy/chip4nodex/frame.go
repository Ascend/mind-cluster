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
	"sort"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// New return npu plugin for chip4nodex including 300I-npu-4p-8 or 300I-npu-4p-16 with 4p mesh
func New(name string) base.AscendHandler {
	m := &chip4nodex{}
	klog.V(util.LogInfoLev).Infof("chip4nodex card type =%s", name)
	num := getNPUNumByHandler(name)
	m.SetMaxNodeNPUNum(num)
	m.SetMaxCardNPUNum(cardsNumPerMesh)
	m.SetPluginName(name)
	m.SetAnnoName(util.NPUCardName)
	m.SetAnnoPreVal(util.NPUCardNamePre)
	m.SetIsNetworkFaultAttention(true)
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
		klog.V(util.LogDebugLev).Infof("%s GetTaskReqNPUNum err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	nodeTop, err := tp.GetUsableTopFromNode(node, tp.NPUTaskNum > 1)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	if err = tp.judgeNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogDebugLev).Infof("the node judgeNodeAndTaskNPU failed, node name %s, err: %s",
			node.Name, err.Error())
		return fmt.Errorf("checkNodeNPUByTask %s, network unhealthy card is [ %s ]",
			util.NodeNotMeetTopologyWarning, node.Annotation[tp.GetNetUnhealthyNPUKey()])
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
		klog.V(util.LogDebugLev).Infof("%s GetTaskReqNPUNum %s: %s", tp.GetPluginName(), task.Name, getErr)
		return getErr
	}

	for _, node := range nodes {
		if node == nil {
			continue
		}
		nNode, ok := tp.Nodes[node.Name]
		if !ok {
			klog.V(util.LogDebugLev).Infof("%s %s ScoreBestNPUNodes %s is not npu node",
				tp.GetPluginName(), task.Name, node.Name)
			continue
		}
		// Get the list of NPUs currently available on the node
		cardIds, err := tp.GetUsableTopFromNode(nNode, tp.NPUTaskNum > 1)
		if err != nil {
			klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodes getErr: %s", tp.GetPluginName(), err)
			continue
		}

		if is4PmeshAffinity(taskNPUNum) {
			sMap[node.Name] = tp.scoreNodeFor4Pmesh(taskNPUNum, cardIds)
			continue
		}
		sMap[node.Name] = tp.scoreNodeForGeneral(taskNPUNum, cardIds)
	}
	klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodes task<%s> sMap<%v>", tp.GetPluginName(),
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
		tp.GetPluginName(), task.Name, node.Name, tp.GetAnnoName(tp.ReqNPUName), util.SafePrint(node.Annotation))
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
		err = fmt.Errorf("%s node<%s> top<%v> can not meet task req<%d>", util.NPUResourceShortageError, node.Name, len(nodeTop), taskNPUNum)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes err: %s", err)
		return nil, err
	}
	if is4PmeshAffinity(taskNPUNum) {
		return tp.selectNPUIn4Pmesh(taskNPUNum, nodeTop), nil
	}
	return nodeTop[:taskNPUNum], nil
}

// selectFeasibleCardsByMeshAffinity selects feasible mesh cards based on 4P mesh affinity.
// Returns nil if no feasible cards found.
func selectFeasibleCardsByMeshAffinity(reqNPUNum, maxCardNPUNum int,
	cardFreeCount map[int]int) map[int]struct{} {
	if !is4PmeshAffinity(reqNPUNum) {
		return nil
	}

	// Single mesh: task needs fewer NPUs than one mesh provides
	if reqNPUNum < maxCardNPUNum {
		var candidates []int
		for id, fc := range cardFreeCount {
			if fc >= reqNPUNum {
				candidates = append(candidates, id)
			}
		}
		if len(candidates) == 0 {
			return nil
		}
		// best-fit: prefer the mesh with the least extra free NPUs
		sort.Slice(candidates, func(i, j int) bool {
			extraI := cardFreeCount[candidates[i]] - reqNPUNum
			extraJ := cardFreeCount[candidates[j]] - reqNPUNum
			if extraI != extraJ {
				return extraI < extraJ
			}
			return candidates[i] < candidates[j]
		})
		return map[int]struct{}{candidates[0]: {}}
	}

	// Multi mesh: task needs one or more complete meshes
	var fullMeshIDs []int
	for id, fc := range cardFreeCount {
		if fc == maxCardNPUNum {
			fullMeshIDs = append(fullMeshIDs, id)
		}
	}
	neededMeshes := (reqNPUNum + maxCardNPUNum - 1) / maxCardNPUNum
	if len(fullMeshIDs) < neededMeshes {
		return nil
	}
	sort.Ints(fullMeshIDs)
	feasibleCards := make(map[int]struct{}, neededMeshes)
	for i := 0; i < neededMeshes; i++ {
		feasibleCards[fullMeshIDs[i]] = struct{}{}
	}
	return feasibleCards
}

// Preemptable override: 4P mesh affinity requires complete mesh groups
func (tp *chip4nodex) Preemptable(preemptor *api.TaskInfo, preemptees []*api.TaskInfo,
	vcNode *plugin.NPUNode) ([]*api.TaskInfo, bool) {
	if tp == nil || preemptor == nil || vcNode == nil || len(preemptees) == 0 {
		klog.V(util.LogInfoLev).Infof("Preemptable: invalid arguments, handler nil=%v preemptor nil=%v "+
			"vcNode nil=%v preemptees=%d", tp == nil, preemptor == nil, vcNode == nil, len(preemptees))
		return nil, false
	}
	maxCardNPUNum := tp.GetMaxCardNPUNum()
	if maxCardNPUNum <= 0 {
		klog.V(util.LogInfoLev).Infof("Preemptable: maxCardNPUNum is 0")
		return nil, false
	}
	reqNPUNum, err := tp.GetTaskReqNPUNum(preemptor)
	if err != nil || reqNPUNum <= 0 {
		klog.V(util.LogInfoLev).Infof("Preemptable: invalid reqNPUNum %d, err %v", reqNPUNum, err)
		return nil, false
	}

	klog.V(util.LogInfoLev).Infof("Preemptable(chip4nodex): task<%s> req<%d> maxCardNPUNum<%d> on node<%s>, "+
		"preemptees<%d>", preemptor.Name, reqNPUNum, maxCardNPUNum, vcNode.Name, len(preemptees))

	availableChipIDs := util.ChangeTopToIntArray(vcNode.Annotation[tp.GetAnnoName(tp.ReqNPUName)],
		tp.GetAnnoPreVal(tp.ReqNPUName))
	cardFreeCount := plugin.CalcCardFreeCount(vcNode, preemptees, maxCardNPUNum, availableChipIDs)
	if len(cardFreeCount) == 0 {
		klog.V(util.LogInfoLev).Infof("Preemptable(chip4nodex): no free cards on node<%s>", vcNode.Name)
		return nil, false
	}

	feasibleCards := selectFeasibleCardsByMeshAffinity(reqNPUNum, maxCardNPUNum, cardFreeCount)
	if feasibleCards != nil {
		klog.V(util.LogInfoLev).Infof("Preemptable(chip4nodex): task<%s> mesh affinity feasible on node<%s>",
			preemptor.Name, vcNode.Name)
		return plugin.FilterPreempteesByFeasibleCards(vcNode, preemptees, feasibleCards, maxCardNPUNum), true
	}

	if is4PmeshAffinity(reqNPUNum) {
		klog.V(util.LogInfoLev).Infof("Preemptable(chip4nodex): task<%s> mesh affinity required but not feasible on node<%s>",
			preemptor.Name, vcNode.Name)
		return nil, false
	}

	klog.V(util.LogInfoLev).Infof("Preemptable(chip4nodex): task<%s> non-affinity path, fallback to default",
		preemptor.Name)
	return tp.NPUHandler.Preemptable(preemptor, preemptees, vcNode)
}

// Reclaimable override: same strategy as preempt for chip4nodex.
func (tp *chip4nodex) Reclaimable(reclaimer *api.TaskInfo, reclaimees []*api.TaskInfo,
	vcNode *plugin.NPUNode) ([]*api.TaskInfo, bool) {
	if tp == nil || reclaimer == nil || vcNode == nil || len(reclaimees) == 0 {
		klog.V(util.LogInfoLev).Infof("Reclaimable: invalid arguments, handler nil=%v reclaimer nil=%v "+
			"vcNode nil=%v reclaimees=%d", tp == nil, reclaimer == nil, vcNode == nil, len(reclaimees))
		return nil, false
	}
	maxCardNPUNum := tp.GetMaxCardNPUNum()
	if maxCardNPUNum <= 0 {
		klog.V(util.LogInfoLev).Infof("Reclaimable: maxCardNPUNum is 0")
		return nil, false
	}
	reqNPUNum, err := tp.GetTaskReqNPUNum(reclaimer)
	if err != nil || reqNPUNum <= 0 {
		klog.V(util.LogInfoLev).Infof("Reclaimable: invalid reqNPUNum %d, err %v", reqNPUNum, err)
		return nil, false
	}

	klog.V(util.LogInfoLev).Infof("Reclaimable(chip4nodex): task<%s> req<%d> maxCardNPUNum<%d> on node<%s>, "+
		"reclaimees<%d>", reclaimer.Name, reqNPUNum, maxCardNPUNum, vcNode.Name, len(reclaimees))

	availableChipIDs := util.ChangeTopToIntArray(vcNode.Annotation[tp.GetAnnoName(tp.ReqNPUName)],
		tp.GetAnnoPreVal(tp.ReqNPUName))
	cardFreeCount := plugin.CalcCardFreeCount(vcNode, reclaimees, maxCardNPUNum, availableChipIDs)
	if len(cardFreeCount) == 0 {
		klog.V(util.LogInfoLev).Infof("Reclaimable(chip4nodex): no free cards on node<%s>", vcNode.Name)
		return nil, false
	}

	feasibleCards := selectFeasibleCardsByMeshAffinity(reqNPUNum, maxCardNPUNum, cardFreeCount)
	if feasibleCards != nil {
		klog.V(util.LogInfoLev).Infof("Reclaimable(chip4nodex): task<%s> mesh affinity feasible on node<%s>",
			reclaimer.Name, vcNode.Name)
		return plugin.FilterPreempteesByFeasibleCards(vcNode, reclaimees, feasibleCards, maxCardNPUNum), true
	}

	if is4PmeshAffinity(reqNPUNum) {
		klog.V(util.LogInfoLev).Infof("Reclaimable(chip4nodex): task<%s> mesh affinity required but not feasible on node<%s>",
			reclaimer.Name, vcNode.Name)
		return nil, false
	}

	klog.V(util.LogInfoLev).Infof("Reclaimable(chip4nodex): task<%s> non-affinity path, fallback to default",
		reclaimer.Name)
	return tp.NPUHandler.Reclaimable(reclaimer, reclaimees, vcNode)
}

// ReleaseAnnotation Used to release allocated resources
func (tp *chip4nodex) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
