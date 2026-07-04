/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package chip8node8ra64sp

import (
	"container/heap"
	"fmt"
	"strconv"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

const (
	inferServiceGroupSameRack = 1
	inferServiceGroupSameSP   = 2
	inferServiceGroupOtherSP  = 3

	superPodIDShift = 32
)

func rackKey(superPodID, rackID int32) int64 {
	return (int64(superPodID) << superPodIDShift) | int64(uint32(rackID))
}

func (tp *chip8node8ra64sp) isInferServiceJobCheck() bool {
	if tp.Label == nil {
		return false
	}
	id, ok := tp.Label[inferServiceIDLabelKey]
	if !ok || id == "" {
		return false
	}
	tp.inferServiceID = id
	return true
}

func (tp *chip8node8ra64sp) getInferServiceScheduledInfo() (
	map[int64]*inferServiceRackInfo,
	map[int32]*inferServiceSPInfo,
) {
	sameRacks := make(map[int64]*inferServiceRackInfo)
	sameSPs := make(map[int32]*inferServiceSPInfo)

	if tp.inferServiceID == "" || tp.ScheduleEnv.Jobs == nil {
		return sameRacks, sameSPs
	}

	klog.V(util.LogDebugLev).Infof("infer service job %s: start collecting scheduled rack/SP info, inferServiceID=%s",
		tp.Name, tp.inferServiceID)

	for jobID, job := range tp.ScheduleEnv.Jobs {
		if job.Label == nil {
			continue
		}
		jobInferID, ok := job.Label[inferServiceIDLabelKey]
		if !ok || jobInferID != tp.inferServiceID {
			continue
		}
		if jobID == tp.Name {
			continue
		}
		if len(job.SuperPods) == 0 {
			continue
		}

		for _, spNodes := range job.SuperPods {
			for _, sn := range spNodes {
				if _, exist := sameRacks[rackKey(sn.SuperPodID, sn.RackID)]; !exist {
					klog.V(util.LogDebugLev).Infof("infer service job %s: collect rack from job %s, superPodID=%d, rackID=%d",
						tp.Name, jobID, sn.SuperPodID, sn.RackID)
					sameRacks[rackKey(sn.SuperPodID, sn.RackID)] = &inferServiceRackInfo{
						rackID:     sn.RackID,
						superPodID: sn.SuperPodID,
					}
				}
				if _, exist := sameSPs[sn.SuperPodID]; !exist {
					klog.V(util.LogDebugLev).Infof("infer service job %s: collect SP from job %s, superPodID=%d",
						tp.Name, jobID, sn.SuperPodID)
					sameSPs[sn.SuperPodID] = &inferServiceSPInfo{
						superPodID: sn.SuperPodID,
					}
				}
			}
		}
	}

	return sameRacks, sameSPs
}

func (tp *chip8node8ra64sp) enrichRackAndSPInfo(
	superPodMap map[int32]superPod,
	sameRacks map[int64]*inferServiceRackInfo,
	sameSPs map[int32]*inferServiceSPInfo,
) {
	for spID, sp := range superPodMap {
		rackGroup := transferSuperPodToRackIdMap(sp)
		spFreeNodes := 0
		spIdleRacks := 0

		for rackID, nodes := range rackGroup {
			freeNodes := len(nodes)
			spFreeNodes += freeNodes
			if freeNodes == rackNodeNum {
				spIdleRacks++
			}
			if info, ok := sameRacks[rackKey(spID, rackID)]; ok {
				klog.V(util.LogDebugLev).Infof("infer service: enrich rack, superPodID=%d, rackID=%d, freeNodes=%d",
					spID, rackID, freeNodes)
				info.freeNodes = freeNodes
			}
		}

		if info, ok := sameSPs[spID]; ok {
			info.idleRackNum = spIdleRacks
			info.freeNodeNum = spFreeNodes
		}
	}
}

type inferServicePQItem struct {
	superPodID  int32
	rackID      int32
	freeNodes   int
	idleRackNum int
	totalFree   int
	group       int
	index       int
}

type inferServicePQ []*inferServicePQItem

func (pq inferServicePQ) Len() int { return len(pq) }

func (pq inferServicePQ) Less(i, j int) bool {
	a, b := pq[i], pq[j]

	if a.group != b.group {
		return a.group < b.group
	}

	switch a.group {
	case inferServiceGroupSameRack:
		return a.freeNodes > b.freeNodes
	case inferServiceGroupSameSP, inferServiceGroupOtherSP:
		if a.idleRackNum != b.idleRackNum {
			return a.idleRackNum > b.idleRackNum
		}
		if a.freeNodes != b.freeNodes {
			return a.freeNodes > b.freeNodes
		}
		return a.totalFree > b.totalFree
	}
	return false
}

func (pq inferServicePQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *inferServicePQ) Push(x interface{}) {
	n := len(*pq)
	item := x.(*inferServicePQItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *inferServicePQ) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[:n-1]
	return item
}

func (tp *chip8node8ra64sp) selectNodesForInferService(
	task *api.TaskInfo,
	nodes []*api.NodeInfo,
) (map[string][]plugin.SuperNode, error) {
	if tp.spBlock <= 0 {
		return nil, fmt.Errorf("invalid spBlock %d for infer service job", tp.spBlock)
	}

	superPodMap := getSuperPodMap(tp.Nodes, nodes, tp.GetPluginName(), tp.uBMemRackNum)

	sameRacks, sameSPs := tp.getInferServiceScheduledInfo()
	tp.enrichRackAndSPInfo(superPodMap, sameRacks, sameSPs)

	spBlockCount := tp.ReqNPUNum / tp.SpBlockNPUNum
	selectedNodes := make(map[string][]plugin.SuperNode)

	klog.V(util.LogInfoLev).Infof("infer service job %s: start selecting nodes, spBlock=%d, spBlockCount=%d, superPodMap size=%d, sameRacks=%d, sameSPs=%d",
		tp.Name, tp.spBlock, spBlockCount, len(superPodMap), len(sameRacks), len(sameSPs))

	pq := tp.buildInferServicePriorityQueue(superPodMap, sameRacks, sameSPs)
	for i := 0; i < spBlockCount; i++ {
		var item *inferServicePQItem
		for pq.Len() > 0 {
			item = heap.Pop(pq).(*inferServicePQItem)
			sp, ok := superPodMap[item.superPodID]
			if !ok || len(sp) < tp.spBlock {
				klog.V(util.LogDebugLev).Infof("infer service: skip PQ item superPodID=%d, rackID=%d, sp not found or insufficient, spLen=%d",
					item.superPodID, item.rackID, len(sp))
				item = nil
				continue
			}
			rackGroup := transferSuperPodToRackIdMap(sp)
			nodesInRack, rackOk := rackGroup[item.rackID]
			if !rackOk || len(nodesInRack) < tp.spBlock {
				klog.V(util.LogDebugLev).Infof("infer service: skip PQ item superPodID=%d, rackID=%d, rack not found or insufficient, nodesInRack=%d",
					item.superPodID, item.rackID, len(nodesInRack))
				item = nil
				continue
			}
			break
		}
		if item == nil {
			klog.V(util.LogDebugLev).Infof("infer service job %s: PQ exhausted, no valid item found for spBlock[%d/%d]",
				tp.Name, i+1, spBlockCount)
			break
		}

		sp := superPodMap[item.superPodID]
		rackGroup := transferSuperPodToRackIdMap(sp)
		nodesInRack := rackGroup[item.rackID]

		spIndex := strconv.Itoa(i)
		klog.V(util.LogInfoLev).Infof("infer service job %s: select spBlock[%d/%d], superPodID=%d, rackID=%d, group=%d, freeNodes=%d",
			tp.Name, i+1, spBlockCount, item.superPodID, item.rackID, item.group, item.freeNodes)
		selectedNodes[spIndex] = make([]plugin.SuperNode, 0, tp.spBlock)
		for j := 0; j < tp.spBlock; j++ {
			selectedNodes[spIndex] = append(selectedNodes[spIndex], plugin.SuperNode{
				Name:       nodesInRack[j].name,
				SuperPodID: nodesInRack[j].superPodID,
				RackID:     nodesInRack[j].rackID,
			})
			delete(sp, nodesInRack[j].name)
		}

		sameRacks[rackKey(item.superPodID, item.rackID)] = &inferServiceRackInfo{
			rackID:     item.rackID,
			superPodID: item.superPodID,
		}
		sameSPs[item.superPodID] = &inferServiceSPInfo{
			superPodID: item.superPodID,
		}
		tp.enrichRackAndSPInfo(superPodMap, sameRacks, sameSPs)
		pq = tp.buildInferServicePriorityQueue(superPodMap, sameRacks, sameSPs)
	}

	if len(selectedNodes) < spBlockCount {
		klog.V(util.LogInfoLev).Infof("infer service job %s: schedule failed, required %d sp-block, got %d, superPodMap=%d",
			tp.Name, spBlockCount, len(selectedNodes), len(superPodMap))
		return nil, fmt.Errorf("infer service schedule failed, required %d sp-block, got %d",
			spBlockCount, len(selectedNodes))
	}

	klog.V(util.LogInfoLev).Infof("infer service schedule success, job %s, inferServiceID %s, selectedNodes %v",
		tp.Name, tp.inferServiceID, selectedNodes)

	return selectedNodes, nil
}

func (tp *chip8node8ra64sp) buildInferServicePriorityQueue(
	superPodMap map[int32]superPod,
	sameRacks map[int64]*inferServiceRackInfo,
	sameSPs map[int32]*inferServiceSPInfo,
) *inferServicePQ {
	pq := make(inferServicePQ, 0)
	heap.Init(&pq)

	for _, rackInfo := range sameRacks {
		sp, ok := superPodMap[rackInfo.superPodID]
		if !ok {
			continue
		}
		rackGroup := transferSuperPodToRackIdMap(sp)
		if _, rackOk := rackGroup[rackInfo.rackID]; !rackOk {
			continue
		}
		if rackInfo.freeNodes < tp.spBlock {
			klog.V(util.LogDebugLev).Infof("infer service: skip sameRack PQ, superPodID=%d, rackID=%d, freeNodes=%d < spBlock=%d",
				rackInfo.superPodID, rackInfo.rackID, rackInfo.freeNodes, tp.spBlock)
			continue
		}
		klog.V(util.LogDebugLev).Infof("infer service: add sameRack to PQ, superPodID=%d, rackID=%d, freeNodes=%d, spBlock=%d",
			rackInfo.superPodID, rackInfo.rackID, rackInfo.freeNodes, tp.spBlock)
		heap.Push(&pq, &inferServicePQItem{
			superPodID: rackInfo.superPodID,
			rackID:     rackInfo.rackID,
			freeNodes:  rackInfo.freeNodes,
			group:      inferServiceGroupSameRack,
		})
	}

	for spID := range sameSPs {
		sp, ok := superPodMap[spID]
		if !ok {
			continue
		}
		rackGroup := transferSuperPodToRackIdMap(sp)
		idleRackNum, totalFree := countSPMetrics(rackGroup)
		for rackID, nodes := range rackGroup {
			if _, excluded := sameRacks[rackKey(spID, rackID)]; excluded {
				continue
			}
			if len(nodes) < tp.spBlock {
				continue
			}
			klog.V(util.LogDebugLev).Infof("infer service: add sameSP to PQ, superPodID=%d, rackID=%d, freeNodes=%d, idleRackNum=%d, totalFree=%d",
				spID, rackID, len(nodes), idleRackNum, totalFree)
			heap.Push(&pq, &inferServicePQItem{
				superPodID:  spID,
				rackID:      rackID,
				freeNodes:   len(nodes),
				idleRackNum: idleRackNum,
				totalFree:   totalFree,
				group:       inferServiceGroupSameSP,
			})
		}
	}

	for spID, sp := range superPodMap {
		if _, ok := sameSPs[spID]; ok {
			continue
		}
		rackGroup := transferSuperPodToRackIdMap(sp)
		idleRackNum, totalFree := countSPMetrics(rackGroup)
		for rackID, nodes := range rackGroup {
			if len(nodes) < tp.spBlock {
				continue
			}
			klog.V(util.LogDebugLev).Infof("infer service: add otherSP to PQ, superPodID=%d, rackID=%d, freeNodes=%d, idleRackNum=%d, totalFree=%d",
				spID, rackID, len(nodes), idleRackNum, totalFree)
			heap.Push(&pq, &inferServicePQItem{
				superPodID:  spID,
				rackID:      rackID,
				freeNodes:   len(nodes),
				idleRackNum: idleRackNum,
				totalFree:   totalFree,
				group:       inferServiceGroupOtherSP,
			})
		}
	}

	return &pq
}

func countSPMetrics(rackGroup map[int32][]nodeBaseInfo) (int, int) {
	idleRackNum := 0
	totalFree := 0
	for _, nodes := range rackGroup {
		totalFree += len(nodes)
		if len(nodes) == rackNodeNum {
			idleRackNum++
		}
	}
	return idleRackNum, totalFree
}
