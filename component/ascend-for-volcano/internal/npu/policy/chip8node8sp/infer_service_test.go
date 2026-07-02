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
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func TestIsInferServiceJobCheck(t *testing.T) {
	t.Run("Label is nil should return false", func(t *testing.T) {
		tp := &chip8node8sp{}
		tp.Label = nil
		if tp.isInferServiceJobCheck() {
			t.Errorf("expected false when Label is nil")
		}
	})

	t.Run("Label without inferServiceIDLabelKey should return false", func(t *testing.T) {
		tp := &chip8node8sp{}
		tp.Label = map[string]string{"other-key": "value"}
		if tp.isInferServiceJobCheck() {
			t.Errorf("expected false when inferServiceIDLabelKey not in Label")
		}
	})

	t.Run("Label with empty inferServiceIDLabelKey should return false", func(t *testing.T) {
		tp := &chip8node8sp{}
		tp.Label = map[string]string{inferServiceIDLabelKey: ""}
		if tp.isInferServiceJobCheck() {
			t.Errorf("expected false when inferServiceIDLabelKey is empty")
		}
	})

	t.Run("Label with valid inferServiceIDLabelKey should return true", func(t *testing.T) {
		tp := &chip8node8sp{}
		tp.Label = map[string]string{inferServiceIDLabelKey: "test-infer-id"}
		if !tp.isInferServiceJobCheck() {
			t.Errorf("expected true when inferServiceIDLabelKey has value")
		}
		if tp.inferServiceID != "test-infer-id" {
			t.Errorf("expected inferServiceID=test-infer-id, got %s", tp.inferServiceID)
		}
	})
}

func TestGetInferServiceScheduledSPs1(t *testing.T) {
	t.Run("empty inferServiceID should return empty map", func(t *testing.T) {
		tp := &chip8node8sp{}
		tp.inferServiceID = ""
		tp.ScheduleEnv = plugin.ScheduleEnv{
			ClusterCache: plugin.ClusterCache{
				Jobs: map[api.JobID]plugin.SchedulerJob{},
			},
		}
		sps := tp.getInferServiceScheduledSPs()
		if len(sps) != 0 {
			t.Errorf("expected empty map, got %d", len(sps))
		}
	})

	t.Run("nil Jobs should return empty map", func(t *testing.T) {
		tp := &chip8node8sp{}
		tp.inferServiceID = "test-id"
		tp.ScheduleEnv.Jobs = nil
		sps := tp.getInferServiceScheduledSPs()
		if len(sps) != 0 {
			t.Errorf("expected empty map when Jobs is nil")
		}
	})
}
func TestGetInferServiceScheduledSPs2(t *testing.T) {
	t.Run("job with same inferServiceID should populate map", func(t *testing.T) {
		tp := &chip8node8sp{}
		tp.inferServiceID = "test-id"
		tp.Name = "current-job"
		tp.ScheduleEnv = plugin.ScheduleEnv{
			ClusterCache: plugin.ClusterCache{
				Jobs: map[api.JobID]plugin.SchedulerJob{
					"other-job": {
						SchedulerJobAttr: util.SchedulerJobAttr{
							ComJob: util.ComJob{
								Label: map[string]string{inferServiceIDLabelKey: "test-id"},
							},
						},
						SuperPods: map[string][]plugin.SuperNode{
							"sp0": {
								{Name: "node0", SuperPodID: 1},
							},
						},
					},
				},
			},
		}
		sps := tp.getInferServiceScheduledSPs()
		if len(sps) != 1 {
			t.Errorf("expected 1 sp, got %d", len(sps))
		}
		if _, ok := sps[1]; !ok {
			t.Errorf("expected superPodID=1 in sameSPs")
		}
	})
}
func TestGetInferServiceScheduledSPs3(t *testing.T) {
	t.Run("current job should be skipped", func(t *testing.T) {
		tp := &chip8node8sp{}
		tp.inferServiceID = "test-id"
		tp.Name = "my-job"
		tp.ScheduleEnv = plugin.ScheduleEnv{
			ClusterCache: plugin.ClusterCache{
				Jobs: map[api.JobID]plugin.SchedulerJob{
					"my-job": {
						SchedulerJobAttr: util.SchedulerJobAttr{
							ComJob: util.ComJob{
								Label: map[string]string{inferServiceIDLabelKey: "test-id"},
							},
						},
						SuperPods: map[string][]plugin.SuperNode{
							"sp0": {
								{Name: "node0", SuperPodID: 1},
							},
						},
					},
				},
			},
		}
		sps := tp.getInferServiceScheduledSPs()
		if len(sps) != 0 {
			t.Errorf("expected empty map when only current job exists")
		}
	})

	t.Run("job with different inferServiceID should be skipped", func(t *testing.T) {
		tp := &chip8node8sp{}
		tp.inferServiceID = "test-id"
		tp.ScheduleEnv = plugin.ScheduleEnv{
			ClusterCache: plugin.ClusterCache{
				Jobs: map[api.JobID]plugin.SchedulerJob{
					"other-job": {
						SchedulerJobAttr: util.SchedulerJobAttr{
							ComJob: util.ComJob{
								Label: map[string]string{inferServiceIDLabelKey: "different-id"},
							},
						},
						SuperPods: map[string][]plugin.SuperNode{
							"sp0": {
								{Name: "node0", SuperPodID: 1},
							},
						},
					},
				},
			},
		}
		sps := tp.getInferServiceScheduledSPs()
		if len(sps) != 0 {
			t.Errorf("expected empty map when inferServiceID does not match")
		}
	})
}

func TestEnrichInferServiceSPInfo(t *testing.T) {
	t.Run("enrich freeNodeNum for sameSPs", func(t *testing.T) {
		superPodTop := map[int32]superPod{
			0: {"node0": {}, "node1": {}, "node2": {}},
			1: {"node3": {}},
		}
		sameSPs := map[int32]*inferServiceSPInfo{
			0: {superPodID: 0},
		}

		tp := &chip8node8sp{}
		tp.enrichInferServiceSPInfo(superPodTop, sameSPs)

		if sameSPs[0].freeNodeNum != 3 {
			t.Errorf("expected freeNodeNum=3, got %d", sameSPs[0].freeNodeNum)
		}
	})

	t.Run("spID not in sameSPs should not be enriched", func(t *testing.T) {
		superPodTop := map[int32]superPod{
			0: {"node0": {}, "node1": {}},
			1: {"node3": {}},
		}
		sameSPs := map[int32]*inferServiceSPInfo{
			0: {superPodID: 0},
		}

		tp := &chip8node8sp{}
		tp.enrichInferServiceSPInfo(superPodTop, sameSPs)

		if _, ok := sameSPs[1]; ok {
			t.Errorf("spID=1 should not be in sameSPs")
		}
	})

	t.Run("empty superPodTop should not panic", func(t *testing.T) {
		superPodTop := map[int32]superPod{}
		sameSPs := map[int32]*inferServiceSPInfo{}

		tp := &chip8node8sp{}
		tp.enrichInferServiceSPInfo(superPodTop, sameSPs)
	})
}

func TestInferServicePQ_Len(t *testing.T) {
	t.Run("empty queue should return 0", func(t *testing.T) {
		pq := make(inferServicePQ, 0)
		if pq.Len() != 0 {
			t.Errorf("expected len=0, got %d", pq.Len())
		}
	})

	t.Run("queue with items should return correct length", func(t *testing.T) {
		pq := make(inferServicePQ, 2)
		pq[0] = &inferServicePQItem{superPodID: 0}
		pq[1] = &inferServicePQItem{superPodID: 1}
		if pq.Len() != 2 {
			t.Errorf("expected len=2, got %d", pq.Len())
		}
	})
}

func TestInferServicePQ_Less(t *testing.T) {
	t.Run("different groups: lower group should come first", func(t *testing.T) {
		pq := inferServicePQ{
			{group: inferServiceGroupSameSP, freeNodes: 1},
			{group: inferServiceGroupOtherSP, freeNodes: 10},
		}
		if !pq.Less(0, 1) {
			t.Errorf("expected sameSP group to be less than otherSP group")
		}
	})

	t.Run("same group: more freeNodes first", func(t *testing.T) {
		pq := inferServicePQ{
			{group: inferServiceGroupSameSP, freeNodes: 5},
			{group: inferServiceGroupSameSP, freeNodes: 3},
		}
		if !pq.Less(0, 1) {
			t.Errorf("expected more freeNodes to come first")
		}
	})

	t.Run("otherSP group: more freeNodes first", func(t *testing.T) {
		pq := inferServicePQ{
			{group: inferServiceGroupOtherSP, freeNodes: 8},
			{group: inferServiceGroupOtherSP, freeNodes: 4},
		}
		if !pq.Less(0, 1) {
			t.Errorf("expected more freeNodes to come first in otherSP group")
		}
	})
}

func TestInferServicePQ_Swap(t *testing.T) {
	pq := inferServicePQ{
		{superPodID: 0, index: 0},
		{superPodID: 1, index: 1},
	}
	pq.Swap(0, 1)
	if pq[0].superPodID != 1 || pq[1].superPodID != 0 {
		t.Errorf("expected items swapped")
	}
	if pq[0].index != 0 || pq[1].index != 1 {
		t.Errorf("expected indices updated after swap")
	}
}

func TestInferServicePQ_Push(t *testing.T) {
	pq := make(inferServicePQ, 0)
	item := &inferServicePQItem{superPodID: 42}
	heap.Push(&pq, item)
	if pq.Len() != 1 {
		t.Errorf("expected len=1 after push, got %d", pq.Len())
	}
	if pq[0].superPodID != 42 {
		t.Errorf("expected superPodID=42, got %d", pq[0].superPodID)
	}
	if pq[0].index != 0 {
		t.Errorf("expected index=0, got %d", pq[0].index)
	}
}

func TestInferServicePQ_Pop(t *testing.T) {
	pq := make(inferServicePQ, 0)
	heap.Push(&pq, &inferServicePQItem{superPodID: 1, group: inferServiceGroupSameSP, freeNodes: 5})
	heap.Push(&pq, &inferServicePQItem{superPodID: 2, group: inferServiceGroupOtherSP, freeNodes: 3})
	item := heap.Pop(&pq).(*inferServicePQItem)
	if item.superPodID != 1 {
		t.Errorf("expected superPodID=1 (sameSP group pops first), got %d", item.superPodID)
	}
	if pq.Len() != 1 {
		t.Errorf("expected len=1 after pop, got %d", pq.Len())
	}
}

func TestBuildInferServicePriorityQueue1(t *testing.T) {
	t.Run("empty maps should return empty queue", func(t *testing.T) {
		tp := &chip8node8sp{}
		tp.spBlock = 4
		superPodTop := map[int32]superPod{}
		sameSPs := map[int32]*inferServiceSPInfo{}
		pq := tp.buildInferServicePriorityQueue(superPodTop, sameSPs)
		if pq.Len() != 0 {
			t.Errorf("expected empty queue, got len=%d", pq.Len())
		}
	})

	t.Run("sameSP with enough freeNodes should be in queue", func(t *testing.T) {
		tp := &chip8node8sp{}
		tp.spBlock = 2
		superPodTop := map[int32]superPod{
			0: {"node0": {}, "node1": {}, "node2": {}},
		}
		sameSPs := map[int32]*inferServiceSPInfo{
			0: {superPodID: 0, freeNodeNum: 3},
		}
		pq := tp.buildInferServicePriorityQueue(superPodTop, sameSPs)
		if pq.Len() == 0 {
			t.Errorf("expected non-empty queue")
		}
		item := heap.Pop(pq).(*inferServicePQItem)
		if item.group != inferServiceGroupSameSP {
			t.Errorf("expected group=sameSP, got %d", item.group)
		}
	})

	t.Run("sameSP with insufficient freeNodes should be excluded", func(t *testing.T) {
		tp := &chip8node8sp{}
		tp.spBlock = 4
		superPodTop := map[int32]superPod{
			0: {"node0": {}, "node1": {}},
		}
		sameSPs := map[int32]*inferServiceSPInfo{
			0: {superPodID: 0, freeNodeNum: 2},
		}
		pq := tp.buildInferServicePriorityQueue(superPodTop, sameSPs)
		if pq.Len() != 0 {
			t.Errorf("expected empty queue when freeNodes < spBlock, got len=%d", pq.Len())
		}
	})
}

func TestBuildInferServicePriorityQueue2(t *testing.T) {
	t.Run("otherSP with enough nodes should be in queue", func(t *testing.T) {
		tp := &chip8node8sp{}
		tp.spBlock = 2
		superPodTop := map[int32]superPod{
			0: {"node0": {}, "node1": {}, "node2": {}},
		}
		sameSPs := map[int32]*inferServiceSPInfo{}
		pq := tp.buildInferServicePriorityQueue(superPodTop, sameSPs)
		if pq.Len() == 0 {
			t.Errorf("expected non-empty queue for otherSP")
		}
		item := heap.Pop(pq).(*inferServicePQItem)
		if item.group != inferServiceGroupOtherSP {
			t.Errorf("expected group=otherSP, got %d", item.group)
		}
	})

	t.Run("sp in sameSPs should not appear in otherSP group", func(t *testing.T) {
		tp := &chip8node8sp{}
		tp.spBlock = 2
		superPodTop := map[int32]superPod{
			0: {"node0": {}, "node1": {}, "node2": {}},
			1: {"node3": {}, "node4": {}, "node5": {}},
		}
		sameSPs := map[int32]*inferServiceSPInfo{
			0: {superPodID: 0, freeNodeNum: 3},
		}
		pq := tp.buildInferServicePriorityQueue(superPodTop, sameSPs)
		sameSPCount := 0
		otherSPCount := 0
		for pq.Len() > 0 {
			item := heap.Pop(pq).(*inferServicePQItem)
			if item.group == inferServiceGroupSameSP {
				sameSPCount++
			}
			if item.group == inferServiceGroupOtherSP {
				otherSPCount++
			}
		}
		if sameSPCount != 1 {
			t.Errorf("expected 1 sameSP item, got %d", sameSPCount)
		}
		if otherSPCount != 1 {
			t.Errorf("expected 1 otherSP item, got %d", otherSPCount)
		}
	})
}

func TestInferServicePQ_FullHeapOrdering(t *testing.T) {
	t.Run("priority ordering: sameSP > otherSP", func(t *testing.T) {
		pq := make(inferServicePQ, 0)
		heap.Init(&pq)
		heap.Push(&pq, &inferServicePQItem{superPodID: 2, freeNodes: 8, group: inferServiceGroupOtherSP})
		heap.Push(&pq, &inferServicePQItem{superPodID: 1, freeNodes: 5, group: inferServiceGroupSameSP})

		first := heap.Pop(&pq).(*inferServicePQItem)
		if first.group != inferServiceGroupSameSP {
			t.Errorf("expected first pop to be sameSP, got group=%d", first.group)
		}
		second := heap.Pop(&pq).(*inferServicePQItem)
		if second.group != inferServiceGroupOtherSP {
			t.Errorf("expected second pop to be otherSP, got group=%d", second.group)
		}
	})
}

func newTestNPUHandler() base.NPUHandler {
	return base.NPUHandler{
		ScheduleEnv: plugin.ScheduleEnv{
			ClusterCache: plugin.ClusterCache{
				Jobs: map[api.JobID]plugin.SchedulerJob{},
			},
		},
	}
}

func buildTestNPUNodes(start, end, spSize int) map[string]plugin.NPUNode {
	nodes := make(map[string]plugin.NPUNode)
	for i := start; i < end; i++ {
		nodeName := fmt.Sprintf("node%d", i)
		nodes[nodeName] = plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       nodeName,
				SuperPodID: int32(i / spSize),
				Annotation: map[string]string{
					util.NPUCardName:    "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7",
					networkUnhealthyNPU: "",
				},
				Label: map[string]string{
					util.AcceleratorType: AcceleratorType,
				},
			},
		}
	}
	return nodes
}

func buildTestNodeInfos(start, end int) []*api.NodeInfo {
	var nodeInfos []*api.NodeInfo
	for i := start; i <= end; i++ {
		nodeName := fmt.Sprintf("node%d", i)
		nodeInfos = append(nodeInfos, &api.NodeInfo{Name: nodeName})
	}
	return nodeInfos
}

func TestSelectNodesForInferService_NormalCase(t *testing.T) {
	tp := &chip8node8sp{}
	tp.spBlock = 2
	tp.NPUJob = &util.NPUJob{
		SpBlockNPUNum: 16,
		ReqNPUNum:     16,
	}

	superPodTop := map[int32]superPod{
		0: {"node0": {}, "node1": {}, "node2": {}, "node3": {}},
	}

	sameSPs := map[int32]*inferServiceSPInfo{}

	pq := tp.buildInferServicePriorityQueue(superPodTop, sameSPs)
	if pq.Len() == 0 {
		t.Fatal("expected non-empty priority queue")
	}

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
		t.Errorf("expected item not nil")
	} else {
		t.Logf("Selected item: superPodID=%d", item.superPodID)
	}
}

func TestSelectNodesForInferService_SkipSuperPod(t *testing.T) {
	tp := &chip8node8sp{}
	tp.spBlock = 4
	tp.NPUJob = &util.NPUJob{
		SpBlockNPUNum: 32,
		ReqNPUNum:     32,
	}

	superPodTop := map[int32]superPod{
		0: {"node0": {}, "node1": {}},
	}

	sameSPs := map[int32]*inferServiceSPInfo{}

	pq := tp.buildInferServicePriorityQueue(superPodTop, sameSPs)

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

	if item != nil {
		t.Errorf("expected item to be nil when no valid superPod found")
	}
}

func TestSelectNodesForInferService_EnoughNodes(t *testing.T) {
	tp := &chip8node8sp{}
	tp.spBlock = 2
	tp.NPUJob = &util.NPUJob{
		SpBlockNPUNum: 16,
		ReqNPUNum:     16,
	}

	superPodTop := map[int32]superPod{
		0: {"node0": {}, "node1": {}},
	}

	sameSPs := map[int32]*inferServiceSPInfo{}

	pq := tp.buildInferServicePriorityQueue(superPodTop, sameSPs)

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
		t.Errorf("expected item not nil when superPod has enough nodes")
	} else if item.superPodID != 0 {
		t.Errorf("expected superPodID=0, got %d", item.superPodID)
	}
}

func TestSelectNodesForInferService_EmptyPQ(t *testing.T) {
	tp := &chip8node8sp{}
	tp.spBlock = 2

	superPodTop := map[int32]superPod{}
	sameSPs := map[int32]*inferServiceSPInfo{}

	pq := tp.buildInferServicePriorityQueue(superPodTop, sameSPs)

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

	if item != nil {
		t.Errorf("expected item to be nil when priority queue is empty")
	}
}

func TestSelectNodesForInferService_MultipleSuperPods(t *testing.T) {
	tp := &chip8node8sp{}
	tp.spBlock = 4
	tp.NPUJob = &util.NPUJob{
		SpBlockNPUNum: 32,
		ReqNPUNum:     32,
	}

	superPodTop := map[int32]superPod{
		0: {"node0": {}, "node1": {}},
		1: {"node2": {}, "node3": {}, "node4": {}, "node5": {}, "node6": {}},
	}

	sameSPs := map[int32]*inferServiceSPInfo{}

	pq := tp.buildInferServicePriorityQueue(superPodTop, sameSPs)

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
		t.Errorf("expected item not nil")
	} else if item.superPodID != 1 {
		t.Errorf("expected to select superPodID=1, got %d", item.superPodID)
	}
}
