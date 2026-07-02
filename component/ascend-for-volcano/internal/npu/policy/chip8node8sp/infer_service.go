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

package chip8node8sp

import (
	"container/heap"
	"fmt"
	"strconv"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func (tp *chip8node8sp) isInferServiceJobCheck() bool {
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

func (tp *chip8node8sp) getInferServiceScheduledSPs() map[int32]*inferServiceSPInfo {
	sameSPs := make(map[int32]*inferServiceSPInfo)

	if tp.inferServiceID == "" || tp.ScheduleEnv.Jobs == nil {
		return sameSPs
	}

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
				if _, exist := sameSPs[sn.SuperPodID]; !exist {
					sameSPs[sn.SuperPodID] = &inferServiceSPInfo{
						superPodID: sn.SuperPodID,
					}
				}
			}
		}
	}

	return sameSPs
}

func (tp *chip8node8sp) enrichInferServiceSPInfo(
	superPodTop map[int32]superPod,
	sameSPs map[int32]*inferServiceSPInfo,
) {
	for spID, sp := range superPodTop {
		if info, ok := sameSPs[spID]; ok {
			info.freeNodeNum = len(sp)
		}
	}
}

func (pq inferServicePQ) Len() int { return len(pq) }

func (pq inferServicePQ) Less(i, j int) bool {
	a, b := pq[i], pq[j]

	if a.group != b.group {
		return a.group < b.group
	}

	return a.freeNodes > b.freeNodes
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

func (tp *chip8node8sp) selectNodesForInferService(
	task *api.TaskInfo,
	nodes []*api.NodeInfo,
) (map[string][]plugin.SuperNode, error) {
	if tp.spBlock <= 0 {
		return nil, fmt.Errorf("invalid spBlock %d for infer service job", tp.spBlock)
	}

	superPodTop := tp.getSuperPodTop(nodes)

	sameSPs := tp.getInferServiceScheduledSPs()
	tp.enrichInferServiceSPInfo(superPodTop, sameSPs)

	spBlockCount := tp.ReqNPUNum / tp.SpBlockNPUNum
	selectedNodes := make(map[string][]plugin.SuperNode)

	pq := tp.buildInferServicePriorityQueue(superPodTop, sameSPs)
	for i := 0; i < spBlockCount; i++ {
		var item *inferServicePQItem
		for pq.Len() > 0 {
			item = heap.Pop(pq).(*inferServicePQItem)
			sp, ok := superPodTop[item.superPodID]
			if !ok || len(sp) < tp.spBlock {
				item = nil
				continue
			}
			break
		}
		if item == nil {
			break
		}

		sp := superPodTop[item.superPodID]
		spIndex := strconv.Itoa(i)
		selectedNodes[spIndex] = make([]plugin.SuperNode, 0, tp.spBlock)
		nodeCount := 0
		for nodeName, nNode := range sp {
			if nodeCount >= tp.spBlock {
				break
			}
			selectedNodes[spIndex] = append(selectedNodes[spIndex], plugin.SuperNode{
				Name:       nodeName,
				SuperPodID: nNode.SuperPodID,
			})
			delete(sp, nodeName)
			nodeCount++
		}

		sameSPs[item.superPodID] = &inferServiceSPInfo{
			superPodID: item.superPodID,
		}
		tp.enrichInferServiceSPInfo(superPodTop, sameSPs)
		pq = tp.buildInferServicePriorityQueue(superPodTop, sameSPs)
	}

	if len(selectedNodes) < spBlockCount {
		return nil, fmt.Errorf("infer service schedule failed, required %d sp-block, got %d",
			spBlockCount, len(selectedNodes))
	}

	klog.V(util.LogInfoLev).Infof("infer service schedule success, job %s, inferServiceID %s, selectedNodes %v",
		tp.Name, tp.inferServiceID, selectedNodes)

	return selectedNodes, nil
}

func (tp *chip8node8sp) buildInferServicePriorityQueue(
	superPodTop map[int32]superPod,
	sameSPs map[int32]*inferServiceSPInfo,
) *inferServicePQ {
	pq := make(inferServicePQ, 0)
	heap.Init(&pq)

	for spID, info := range sameSPs {
		sp, ok := superPodTop[spID]
		if !ok {
			continue
		}
		if len(sp) < tp.spBlock {
			continue
		}
		heap.Push(&pq, &inferServicePQItem{
			superPodID: spID,
			freeNodes:  info.freeNodeNum,
			group:      inferServiceGroupSameSP,
		})
	}

	for spID, sp := range superPodTop {
		if _, ok := sameSPs[spID]; ok {
			continue
		}
		if len(sp) < tp.spBlock {
			continue
		}
		heap.Push(&pq, &inferServicePQItem{
			superPodID: spID,
			freeNodes:  len(sp),
			group:      inferServiceGroupOtherSP,
		})
	}

	return &pq
}
