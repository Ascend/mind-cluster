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

/*
Package ascend910a3 is using for A3 affinity schedule.
*/
package ascend910a3

import (
	"fmt"
	"sort"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// GetNodeCardTopology get node card topology by npu index.
func (tp *Base910A3) GetNodeCardTopology(npuIndex []int) map[int][]int {
	cardTopology := make(map[int][]int)
	for _, index := range npuIndex {
		cardId := index / tp.MaxCardNPUNum
		_, ok := cardTopology[cardId]
		if !ok {
			cardTopology[cardId] = make([]int, 0, tp.MaxCardNPUNum)
		}
		cardTopology[cardId] = append(cardTopology[cardId], index)
	}
	return cardTopology
}

// JudgeNodeAndTaskNPU judge node and task npu is meet.
func (tp *Base910A3) JudgeNodeAndTaskNPU(taskNPU int, nodeNPUTopology []int) error {
	if err := tp.NPUHandler.JudgeNodeAndTaskNPU(taskNPU, nodeNPUTopology); err != nil {
		return err
	}
	if taskNPU == 1 {
		return nil
	}
	fitDies := 0
	cardTopology := tp.GetNodeCardTopology(nodeNPUTopology)
	for _, card := range cardTopology {
		// whole card schedule
		if len(card) == tp.MaxCardNPUNum {
			fitDies += tp.MaxCardNPUNum
		}
	}
	if fitDies < taskNPU {
		return fmt.Errorf("npu top[%v] is not meet task req(%d)", nodeNPUTopology, taskNPU)
	}
	return nil
}

// SelectNPUFromNode select npu from node.
func (tp *Base910A3) SelectNPUFromNode(task *api.TaskInfo, node plugin.NPUNode, isDistributeJob bool) ([]int, error) {
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %#v", tp.GetPluginName(), err)
		return nil, err
	}
	npuTop, err := tp.GetUsableTopFromNode(node, isDistributeJob)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s getUsableTopFromNode err: %#v", tp.GetPluginName(), err)
		return nil, err
	}
	klog.V(util.LogDebugLev).Infof("node %s usable npu list: %v", node.Name, npuTop)
	if len(npuTop) < taskNPUNum {
		return nil, fmt.Errorf("%s node<%s> don't have enough usable npu", util.NPUResourceShortageError, node.Name)
	}
	// job valid has already been carried out earlier, and the invalid number of cards is not considered here
	if len(npuTop) == taskNPUNum {
		return npuTop, nil
	}
	return tp.selectNPUForA3Job(taskNPUNum, npuTop, node)
}

func (tp *Base910A3) selectNPUForA3Job(taskNPUNum int, npuTop []int, node plugin.NPUNode) ([]int, error) {
	sort.Ints(npuTop)
	klog.V(util.LogDebugLev).Infof("%s select %d NPU Node(%s) nodeTop<%v>", tp.GetPluginName(), taskNPUNum,
		node.Name, npuTop)
	cardTop := tp.GetNodeCardTopology(npuTop)

	cardTopSlice := make([][]int, 0)
	for _, card := range cardTop {
		cardTopSlice = append(cardTopSlice, card)
	}
	sort.Slice(cardTopSlice, func(i, j int) bool {
		return len(cardTopSlice[i]) < len(cardTopSlice[j])
	})
	klog.V(util.LogDebugLev).Infof("%s selectNPUFromNode cardTopSlice<%v>", tp.GetPluginName(), cardTopSlice)
	var selected []int
	for _, card := range cardTopSlice {
		if taskNPUNum == 0 {
			break
		}
		// single die schedule
		if taskNPUNum == 1 {
			selected = append(selected, card[0])
			break
		}
		// whole card schedule
		if len(card) == tp.MaxCardNPUNum {
			selected = append(selected, card...)
			taskNPUNum -= tp.MaxCardNPUNum
		}
	}
	return selected, nil
}

// Preemptable override: multi-chip tasks must select complete Dies only
func (tp *Base910A3) Preemptable(preemptor *api.TaskInfo, preemptees []*api.TaskInfo,
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

	klog.V(util.LogInfoLev).Infof("Preemptable(910A3): task<%s> req<%d> maxCardNPUNum<%d> on node<%s>, "+
		"preemptees<%d>", preemptor.Name, reqNPUNum, maxCardNPUNum, vcNode.Name, len(preemptees))

	cardFreeCount := plugin.CalcCardFreeCount(vcNode, preemptees, maxCardNPUNum)
	if len(cardFreeCount) == 0 {
		klog.V(util.LogInfoLev).Infof("Preemptable(910A3): no free cards on node<%s>", vcNode.Name)
		return nil, false
	}

	// Single chip: any Die with freeCount >= 1 is fine
	if reqNPUNum == 1 {
		klog.V(util.LogInfoLev).Infof("Preemptable(910A3): task<%s> single chip, any Die with freeCount>=1",
			preemptor.Name)
		type cardInfo struct {
			id        int
			freeCount int
		}
		cards := make([]cardInfo, 0, len(cardFreeCount))
		for id, fc := range cardFreeCount {
			if fc >= 1 {
				cards = append(cards, cardInfo{id, fc})
			}
		}
		if len(cards) == 0 {
			klog.V(util.LogInfoLev).Infof("Preemptable(910A3): no Die with freeCount>=1 on node<%s>", vcNode.Name)
			return nil, false
		}
		feasibleCards := make(map[int]struct{})
		feasibleCards[cards[0].id] = struct{}{}
		klog.V(util.LogInfoLev).Infof("Preemptable(910A3): task<%s> selected Die<%d> with freeCount<%d>",
			preemptor.Name, cards[0].id, cards[0].freeCount)
		return plugin.FilterPreempteesByFeasibleCards(vcNode, preemptees, feasibleCards, maxCardNPUNum), true
	}

	// Multi-chip: only select complete Dies (freeCount == maxCardNPUNum)
	klog.V(util.LogInfoLev).Infof("Preemptable(910A3): task<%s> multi-chip, need complete Dies only",
		preemptor.Name)
	type dieInfo struct {
		id        int
		freeCount int
	}
	var fullDies []dieInfo
	for id, fc := range cardFreeCount {
		if fc == maxCardNPUNum {
			fullDies = append(fullDies, dieInfo{id, fc})
		}
	}
	neededDies := (reqNPUNum + maxCardNPUNum - 1) / maxCardNPUNum
	klog.V(util.LogInfoLev).Infof("Preemptable(910A3): task<%s> fullDies<%d> neededDies<%d>",
		preemptor.Name, len(fullDies), neededDies)
	if len(fullDies) < neededDies {
		klog.V(util.LogInfoLev).Infof("Preemptable(910A3): not enough full Dies: %d < %d on node<%s>",
			len(fullDies), neededDies, vcNode.Name)
		return nil, false
	}
	sort.Slice(fullDies, func(i, j int) bool { return fullDies[i].id < fullDies[j].id })
	feasibleCards := make(map[int]struct{})
	for i := 0; i < neededDies; i++ {
		feasibleCards[fullDies[i].id] = struct{}{}
	}
	klog.V(util.LogInfoLev).Infof("Preemptable(910A3): task<%s> selected %d Dies on node<%s>, feasible",
		preemptor.Name, neededDies, vcNode.Name)
	return plugin.FilterPreempteesByFeasibleCards(vcNode, preemptees, feasibleCards, maxCardNPUNum), true
}

// Reclaimable override: same strategy as preempt for 910A3.
func (tp *Base910A3) Reclaimable(reclaimer *api.TaskInfo, reclaimees []*api.TaskInfo,
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

	klog.V(util.LogInfoLev).Infof("Reclaimable(910A3): task<%s> req<%d> maxCardNPUNum<%d> on node<%s>, "+
		"reclaimees<%d>", reclaimer.Name, reqNPUNum, maxCardNPUNum, vcNode.Name, len(reclaimees))

	cardFreeCount := plugin.CalcCardFreeCount(vcNode, reclaimees, maxCardNPUNum)
	if len(cardFreeCount) == 0 {
		klog.V(util.LogInfoLev).Infof("Reclaimable(910A3): no free cards on node<%s>", vcNode.Name)
		return nil, false
	}

	// Single chip: any Die with freeCount >= 1 is fine
	if reqNPUNum == 1 {
		klog.V(util.LogInfoLev).Infof("Reclaimable(910A3): task<%s> single chip, any Die with freeCount>=1",
			reclaimer.Name)
		type cardInfo struct {
			id        int
			freeCount int
		}
		cards := make([]cardInfo, 0, len(cardFreeCount))
		for id, fc := range cardFreeCount {
			if fc >= 1 {
				cards = append(cards, cardInfo{id, fc})
			}
		}
		if len(cards) == 0 {
			klog.V(util.LogInfoLev).Infof("Reclaimable(910A3): no Die with freeCount>=1 on node<%s>", vcNode.Name)
			return nil, false
		}
		feasibleCards := make(map[int]struct{})
		feasibleCards[cards[0].id] = struct{}{}
		klog.V(util.LogInfoLev).Infof("Reclaimable(910A3): task<%s> selected Die<%d> with freeCount<%d>",
			reclaimer.Name, cards[0].id, cards[0].freeCount)
		return plugin.FilterPreempteesByFeasibleCards(vcNode, reclaimees, feasibleCards, maxCardNPUNum), true
	}

	// Multi-chip: only select complete Dies (freeCount == maxCardNPUNum)
	klog.V(util.LogInfoLev).Infof("Reclaimable(910A3): task<%s> multi-chip, need complete Dies only",
		reclaimer.Name)
	type dieInfo struct {
		id        int
		freeCount int
	}
	var fullDies []dieInfo
	for id, fc := range cardFreeCount {
		if fc == maxCardNPUNum {
			fullDies = append(fullDies, dieInfo{id, fc})
		}
	}
	neededDies := (reqNPUNum + maxCardNPUNum - 1) / maxCardNPUNum
	klog.V(util.LogInfoLev).Infof("Reclaimable(910A3): task<%s> fullDies<%d> neededDies<%d>",
		reclaimer.Name, len(fullDies), neededDies)
	if len(fullDies) < neededDies {
		klog.V(util.LogInfoLev).Infof("Reclaimable(910A3): not enough full Dies: %d < %d on node<%s>",
			len(fullDies), neededDies, vcNode.Name)
		return nil, false
	}
	sort.Slice(fullDies, func(i, j int) bool { return fullDies[i].id < fullDies[j].id })
	feasibleCards := make(map[int]struct{})
	for i := 0; i < neededDies; i++ {
		feasibleCards[fullDies[i].id] = struct{}{}
	}
	klog.V(util.LogInfoLev).Infof("Reclaimable(910A3): task<%s> selected %d Dies on node<%s>, feasible",
		reclaimer.Name, neededDies, vcNode.Name)
	return plugin.FilterPreempteesByFeasibleCards(vcNode, reclaimees, feasibleCards, maxCardNPUNum), true
}
