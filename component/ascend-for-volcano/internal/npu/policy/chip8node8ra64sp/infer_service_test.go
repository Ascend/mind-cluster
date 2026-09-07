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
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func TestIsInferServiceJobCheck(t *testing.T) {
	t.Run("Label is nil should return false", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.Label = nil
		if tp.isInferServiceJobCheck() {
			t.Errorf("expected false when Label is nil")
		}
	})

	t.Run("Label without inferServiceIDLabelKey should return false", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.Label = map[string]string{"other-key": "value"}
		if tp.isInferServiceJobCheck() {
			t.Errorf("expected false when inferServiceIDLabelKey not in Label")
		}
	})

	t.Run("Label with empty inferServiceIDLabelKey should return false", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.Label = map[string]string{inferServiceIDLabelKey: ""}
		if tp.isInferServiceJobCheck() {
			t.Errorf("expected false when inferServiceIDLabelKey is empty")
		}
	})

	t.Run("Label with valid inferServiceIDLabelKey should return true", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.Label = map[string]string{inferServiceIDLabelKey: "test-infer-id"}
		if !tp.isInferServiceJobCheck() {
			t.Errorf("expected true when inferServiceIDLabelKey has value")
		}
		if tp.inferServiceID != "test-infer-id" {
			t.Errorf("expected inferServiceID=test-infer-id, got %s", tp.inferServiceID)
		}
	})
}

func TestGetInferServiceScheduledInfo(t *testing.T) {
	t.Run("empty inferServiceID should return empty maps", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.inferServiceID = ""
		tp.ScheduleEnv = plugin.ScheduleEnv{
			ClusterCache: plugin.ClusterCache{
				Jobs: map[api.JobID]plugin.SchedulerJob{},
			},
		}
		racks, sps := tp.getInferServiceScheduledInfo()
		if len(racks) != 0 || len(sps) != 0 {
			t.Errorf("expected empty maps, got racks=%d, sps=%d", len(racks), len(sps))
		}
	})
	t.Run("nil Jobs should return empty maps", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.inferServiceID = "test-id"
		tp.ScheduleEnv.Jobs = nil
		racks, sps := tp.getInferServiceScheduledInfo()
		if len(racks) != 0 || len(sps) != 0 {
			t.Errorf("expected empty maps when Jobs is nil")
		}
	})
	t.Run("current job should be skipped", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
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
								{Name: "node0", SuperPodID: 1, RackID: 10},
							},
						},
					},
				},
			},
		}
		racks, sps := tp.getInferServiceScheduledInfo()
		if len(racks) != 0 || len(sps) != 0 {
			t.Errorf("expected empty maps when only current job exists")
		}
	})

	t.Run("multi-SuperPod with same RackID should not collide", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.inferServiceID = "test-id"
		tp.Name = "current-job"
		tp.ScheduleEnv = plugin.ScheduleEnv{
			ClusterCache: plugin.ClusterCache{
				Jobs: map[api.JobID]plugin.SchedulerJob{
					"job-a": {
						SchedulerJobAttr: util.SchedulerJobAttr{
							ComJob: util.ComJob{
								Label: map[string]string{inferServiceIDLabelKey: "test-id"},
							},
						},
						SuperPods: map[string][]plugin.SuperNode{
							"sp-a": {
								{Name: "node-a-0", SuperPodID: 0, RackID: 0},
							},
						},
					},
					"job-b": {
						SchedulerJobAttr: util.SchedulerJobAttr{
							ComJob: util.ComJob{
								Label: map[string]string{inferServiceIDLabelKey: "test-id"},
							},
						},
						SuperPods: map[string][]plugin.SuperNode{
							"sp-b": {
								{Name: "node-b-0", SuperPodID: 1, RackID: 0},
							},
						},
					},
				},
			},
		}

		racks, sps := tp.getInferServiceScheduledInfo()

		if len(racks) != 2 {
			t.Errorf("expected 2 racks (one per SuperPod), got %d — possible key collision", len(racks))
		}

		infoA, okA := racks[rackKey(0, 0)]
		if !okA {
			t.Errorf("expected rackKey(0,0) from job-a to exist")
		} else {
			if infoA.superPodID != 0 || infoA.rackID != 0 {
				t.Errorf("rackKey(0,0): expected superPodID=0, rackID=0, got superPodID=%d, rackID=%d",
					infoA.superPodID, infoA.rackID)
			}
		}

		infoB, okB := racks[rackKey(1, 0)]
		if !okB {
			t.Errorf("expected rackKey(1,0) from job-b to exist — SuperPodID=1, RackID=0")
		} else {
			if infoB.superPodID != 1 || infoB.rackID != 0 {
				t.Errorf("rackKey(1,0): expected superPodID=1, rackID=0, got superPodID=%d, rackID=%d",
					infoB.superPodID, infoB.rackID)
			}
		}

		if len(sps) != 2 {
			t.Errorf("expected 2 SPs, got %d", len(sps))
		}
	})
}

func TestEnrichRackAndSPInfo(t *testing.T) {
	t.Run("empty superPodMap should not panic", func(t *testing.T) {
		superPodMap := map[int32]superPod{}
		sameRacks := map[int64]*inferServiceRackInfo{}
		sameSPs := map[int32]*inferServiceSPInfo{}

		tp := &chip8node8ra64sp{}
		tp.enrichRackAndSPInfo(superPodMap, sameRacks, sameSPs)
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
			{group: inferServiceGroupSameRack, freeNodes: 1},
			{group: inferServiceGroupSameSP, freeNodes: 10},
		}
		if !pq.Less(0, 1) {
			t.Errorf("expected sameRack group to be less than sameSP group")
		}
	})

	t.Run("same group sameRack: more freeNodes first", func(t *testing.T) {
		pq := inferServicePQ{
			{group: inferServiceGroupSameRack, freeNodes: 5},
			{group: inferServiceGroupSameRack, freeNodes: 3},
		}
		if !pq.Less(0, 1) {
			t.Errorf("expected more freeNodes to come first in sameRack group")
		}
	})

	t.Run("same group sameSP: more idleRackNum first", func(t *testing.T) {
		pq := inferServicePQ{
			{group: inferServiceGroupSameSP, idleRackNum: 3, freeNodes: 5, totalFree: 10},
			{group: inferServiceGroupSameSP, idleRackNum: 1, freeNodes: 8, totalFree: 20},
		}
		if !pq.Less(0, 1) {
			t.Errorf("expected more idleRackNum to come first in sameSP group")
		}
	})

	t.Run("same group sameSP same idleRackNum: more freeNodes first", func(t *testing.T) {
		pq := inferServicePQ{
			{group: inferServiceGroupSameSP, idleRackNum: 2, freeNodes: 8, totalFree: 10},
			{group: inferServiceGroupSameSP, idleRackNum: 2, freeNodes: 5, totalFree: 20},
		}
		if !pq.Less(0, 1) {
			t.Errorf("expected more freeNodes to come first when idleRackNum is equal")
		}
	})

	t.Run("same group sameSP same idleRackNum same freeNodes: more totalFree first", func(t *testing.T) {
		pq := inferServicePQ{
			{group: inferServiceGroupSameSP, idleRackNum: 2, freeNodes: 5, totalFree: 20},
			{group: inferServiceGroupSameSP, idleRackNum: 2, freeNodes: 5, totalFree: 10},
		}
		if !pq.Less(0, 1) {
			t.Errorf("expected more totalFree to come first when idleRackNum and freeNodes are equal")
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
	heap.Push(&pq, &inferServicePQItem{superPodID: 1, group: inferServiceGroupSameRack, freeNodes: 5})
	heap.Push(&pq, &inferServicePQItem{superPodID: 2, group: inferServiceGroupOtherSP, freeNodes: 3})
	item := heap.Pop(&pq).(*inferServicePQItem)
	if item.superPodID != 1 {
		t.Errorf("expected superPodID=1 (sameRack group pops first), got %d", item.superPodID)
	}
	if pq.Len() != 1 {
		t.Errorf("expected len=1 after pop, got %d", pq.Len())
	}
}

func TestCountSPMetrics(t *testing.T) {
	t.Run("empty rackGroup should return zeros", func(t *testing.T) {
		idle, total := countSPMetrics(map[int32][]nodeBaseInfo{})
		if idle != 0 || total != 0 {
			t.Errorf("expected (0,0), got (%d,%d)", idle, total)
		}
	})

	t.Run("rackGroup with full and partial racks", func(t *testing.T) {
		rackGroup := map[int32][]nodeBaseInfo{
			0: make([]nodeBaseInfo, rackNodeNum),
			1: make([]nodeBaseInfo, 4),
		}
		idle, total := countSPMetrics(rackGroup)
		if idle != 1 {
			t.Errorf("expected idleRackNum=1, got %d", idle)
		}
		if total != rackNodeNum+4 {
			t.Errorf("expected totalFree=%d, got %d", rackNodeNum+4, total)
		}
	})
}

func TestBuildInferServicePriorityQueue1(t *testing.T) {
	t.Run("empty maps should return empty queue", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.spBlock = 4
		superPodMap := map[int32]superPod{}
		sameRacks := map[int64]*inferServiceRackInfo{}
		sameSPs := map[int32]*inferServiceSPInfo{}
		pq := tp.buildInferServicePriorityQueue(superPodMap, sameRacks, sameSPs)
		if pq.Len() != 0 {
			t.Errorf("expected empty queue, got len=%d", pq.Len())
		}
	})

	t.Run("sameRack with enough freeNodes should be in queue", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.spBlock = 4
		superPodMap := buildSuperPodsByParams(map[int32]int32{0: 16})
		sameRacks := map[int64]*inferServiceRackInfo{
			rackKey(0, 0): {rackID: 0, superPodID: 0, freeNodes: 8},
		}
		sameSPs := map[int32]*inferServiceSPInfo{}
		pq := tp.buildInferServicePriorityQueue(superPodMap, sameRacks, sameSPs)
		if pq.Len() == 0 {
			t.Errorf("expected non-empty queue")
		}
		item := heap.Pop(pq).(*inferServicePQItem)
		if item.group != inferServiceGroupSameRack {
			t.Errorf("expected group=sameRack, got %d", item.group)
		}
	})
}

func TestBuildInferServicePriorityQueue2(t *testing.T) {
	t.Run("otherSP racks with enough nodes should be in queue", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.spBlock = 4
		superPodMap := buildSuperPodsByParams(map[int32]int32{0: 16})
		sameRacks := map[int64]*inferServiceRackInfo{}
		sameSPs := map[int32]*inferServiceSPInfo{}
		pq := tp.buildInferServicePriorityQueue(superPodMap, sameRacks, sameSPs)
		if pq.Len() == 0 {
			t.Errorf("expected non-empty queue for otherSP")
		}
		item := heap.Pop(pq).(*inferServicePQItem)
		if item.group != inferServiceGroupOtherSP {
			t.Errorf("expected group=otherSP, got %d", item.group)
		}
	})

	t.Run("sameSP racks not in sameRacks should be in queue", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.spBlock = 4
		superPodMap := buildSuperPodsByParams(map[int32]int32{0: 24})
		sameRacks := map[int64]*inferServiceRackInfo{
			rackKey(0, 0): {rackID: 0, superPodID: 0, freeNodes: 8},
		}
		sameSPs := map[int32]*inferServiceSPInfo{
			0: {superPodID: 0},
		}
		pq := tp.buildInferServicePriorityQueue(superPodMap, sameRacks, sameSPs)
		sameRackCount := 0
		sameSPCount := 0
		for pq.Len() > 0 {
			item := heap.Pop(pq).(*inferServicePQItem)
			if item.group == inferServiceGroupSameRack {
				sameRackCount++
			}
			if item.group == inferServiceGroupSameSP {
				sameSPCount++
			}
		}
		if sameRackCount != 1 {
			t.Errorf("expected 1 sameRack item, got %d", sameRackCount)
		}
		if sameSPCount < 1 {
			t.Errorf("expected at least 1 sameSP item, got %d", sameSPCount)
		}
	})
}

func TestInferServicePQ_FullHeapOrdering(t *testing.T) {
	t.Run("priority ordering: sameRack > sameSP > otherSP", func(t *testing.T) {
		pq := make(inferServicePQ, 0)
		heap.Init(&pq)
		heap.Push(&pq, &inferServicePQItem{superPodID: 3, rackID: 30, freeNodes: 8, group: inferServiceGroupOtherSP})
		heap.Push(&pq, &inferServicePQItem{superPodID: 1, rackID: 10, freeNodes: 8, group: inferServiceGroupSameSP, idleRackNum: 1, totalFree: 8})
		heap.Push(&pq, &inferServicePQItem{superPodID: 2, rackID: 20, freeNodes: 8, group: inferServiceGroupSameRack})

		first := heap.Pop(&pq).(*inferServicePQItem)
		if first.group != inferServiceGroupSameRack {
			t.Errorf("expected first pop to be sameRack, got group=%d", first.group)
		}
		second := heap.Pop(&pq).(*inferServicePQItem)
		if second.group != inferServiceGroupSameSP {
			t.Errorf("expected second pop to be sameSP, got group=%d", second.group)
		}
		third := heap.Pop(&pq).(*inferServicePQItem)
		if third.group != inferServiceGroupOtherSP {
			t.Errorf("expected third pop to be otherSP, got group=%d", third.group)
		}
	})
}

func BenchmarkCountSPMetrics(b *testing.B) {
	rackGroup := map[int32][]nodeBaseInfo{
		0: make([]nodeBaseInfo, rackNodeNum),
		1: make([]nodeBaseInfo, rackNodeNum),
		2: make([]nodeBaseInfo, 4),
	}
	for i := 0; i < b.N; i++ {
		countSPMetrics(rackGroup)
	}
}

func BenchmarkBuildInferServicePriorityQueue(b *testing.B) {
	tp := &chip8node8ra64sp{}
	tp.spBlock = 4
	superPodMap := buildSuperPodsByParams(map[int32]int32{0: 64, 1: 64, 2: 64})
	sameRacks := map[int64]*inferServiceRackInfo{
		rackKey(0, 0): {rackID: 0, superPodID: 0, freeNodes: 8},
	}
	sameSPs := map[int32]*inferServiceSPInfo{
		0: {superPodID: 0},
	}
	for i := 0; i < b.N; i++ {
		tp.buildInferServicePriorityQueue(superPodMap, sameRacks, sameSPs)
	}
}

func TestSelectNodesForInferService_NormalCase(t *testing.T) {
	tp := &chip8node8ra64sp{}
	tp.spBlock = 4
	tp.uBMemRackNum = 8
	tp.NPUJob = &util.NPUJob{
		SpBlockNPUNum: 32,
		ReqNPUNum:     32,
	}

	superPodMap := buildSuperPodsByParams(map[int32]int32{0: 16})

	sameRacks := map[int64]*inferServiceRackInfo{}
	sameSPs := map[int32]*inferServiceSPInfo{}

	pq := tp.buildInferServicePriorityQueue(superPodMap, sameRacks, sameSPs)
	if pq.Len() == 0 {
		t.Fatal("expected non-empty priority queue")
	}

	var item *inferServicePQItem
	for pq.Len() > 0 {
		item = heap.Pop(pq).(*inferServicePQItem)
		sp, ok := superPodMap[item.superPodID]
		if !ok || len(sp) < tp.spBlock {
			continue
		}
		rackGroup := transferSuperPodToRackIdMap(sp)
		nodesInRack, rackOk := rackGroup[item.rackID]
		if !rackOk || len(nodesInRack) < tp.spBlock {
			continue
		}
		break
	}

	if item == nil {
		t.Errorf("expected item not nil")
	} else {
		t.Logf("Selected item: superPodID=%d, rackID=%d", item.superPodID, item.rackID)
	}
}

func TestSelectNodesForInferService_SkipSuperPod(t *testing.T) {
	tp := &chip8node8ra64sp{}
	tp.spBlock = 8
	tp.uBMemRackNum = 8
	tp.NPUJob = &util.NPUJob{
		SpBlockNPUNum: 64,
		ReqNPUNum:     64,
	}

	superPodMap := buildSuperPodsByParams(map[int32]int32{0: 4})

	sameRacks := map[int64]*inferServiceRackInfo{}
	sameSPs := map[int32]*inferServiceSPInfo{}

	pq := tp.buildInferServicePriorityQueue(superPodMap, sameRacks, sameSPs)

	var item *inferServicePQItem
	for pq.Len() > 0 {
		item = heap.Pop(pq).(*inferServicePQItem)
		sp, ok := superPodMap[item.superPodID]
		if !ok || len(sp) < tp.spBlock {
			continue
		}
		rackGroup := transferSuperPodToRackIdMap(sp)
		nodesInRack, rackOk := rackGroup[item.rackID]
		if !rackOk || len(nodesInRack) < tp.spBlock {
			continue
		}
		break
	}

	if item != nil {
		t.Errorf("expected item to be nil when no valid superPod found")
	}
}

func TestSelectNodesForInferService_SkipRack(t *testing.T) {
	tp := &chip8node8ra64sp{}
	tp.spBlock = 9
	tp.uBMemRackNum = 8
	tp.NPUJob = &util.NPUJob{
		SpBlockNPUNum: 72,
		ReqNPUNum:     72,
	}

	superPodMap := buildSuperPodsByParams(map[int32]int32{0: 10})

	sameRacks := map[int64]*inferServiceRackInfo{}
	sameSPs := map[int32]*inferServiceSPInfo{}

	pq := tp.buildInferServicePriorityQueue(superPodMap, sameRacks, sameSPs)

	var item *inferServicePQItem
	for pq.Len() > 0 {
		item = heap.Pop(pq).(*inferServicePQItem)
		sp, ok := superPodMap[item.superPodID]
		if !ok || len(sp) < tp.spBlock {
			continue
		}
		rackGroup := transferSuperPodToRackIdMap(sp)
		nodesInRack, rackOk := rackGroup[item.rackID]
		if !rackOk || len(nodesInRack) < tp.spBlock {
			continue
		}
		break
	}

	if item != nil {
		t.Errorf("expected item to be nil when no valid rack found")
	}
}

func TestSelectNodesForInferService_EmptyPQ(t *testing.T) {
	tp := &chip8node8ra64sp{}
	tp.spBlock = 4

	superPodMap := map[int32]superPod{}
	sameRacks := map[int64]*inferServiceRackInfo{}
	sameSPs := map[int32]*inferServiceSPInfo{}

	pq := tp.buildInferServicePriorityQueue(superPodMap, sameRacks, sameSPs)

	var item *inferServicePQItem
	for pq.Len() > 0 {
		item = heap.Pop(pq).(*inferServicePQItem)
		sp, ok := superPodMap[item.superPodID]
		if !ok || len(sp) < tp.spBlock {
			continue
		}
		rackGroup := transferSuperPodToRackIdMap(sp)
		nodesInRack, rackOk := rackGroup[item.rackID]
		if !rackOk || len(nodesInRack) < tp.spBlock {
			continue
		}
		break
	}

	if item != nil {
		t.Errorf("expected item to be nil when priority queue is empty")
	}
}

// buildPodLevelNodes builds npu nodes and api nodes for pod-level rescheduling tests:
// spNodeCount nodes on one super pod, rack of each node is i/rackNodeNum.
func buildPodLevelNodes(spNodeCount int) (map[string]plugin.NPUNode, []*api.NodeInfo) {
	npuNodes := make(map[string]plugin.NPUNode)
	apiNodes := make([]*api.NodeInfo, 0, spNodeCount)
	for i := 0; i < spNodeCount; i++ {
		name := fmt.Sprintf("node-0-%d", i)
		npuNodes[name] = plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       name,
				SuperPodID: 0,
				RackID:     int32(i / rackNodeNum),
			},
		}
		apiNodes = append(apiNodes, &api.NodeInfo{Name: name})
	}
	return npuNodes, apiNodes
}

func TestPopValidInferServiceSPItem(t *testing.T) {
	t.Run("valid item is returned", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.spBlock = 4
		superPodMap := buildSuperPodsByParams(map[int32]int32{0: 16})
		pq := tp.buildInferServicePriorityQueue(superPodMap,
			map[int64]*inferServiceRackInfo{}, map[int32]*inferServiceSPInfo{})
		item := tp.popValidInferServiceSPItem(pq, superPodMap)
		if item == nil {
			t.Fatalf("expected valid item, got nil")
		}
		if item.superPodID != 0 {
			t.Errorf("expected superPodID=0, got %d", item.superPodID)
		}
	})
	t.Run("item with missing super pod is skipped", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.spBlock = 4
		superPodMap := buildSuperPodsByParams(map[int32]int32{0: 16})
		pq := tp.buildInferServicePriorityQueue(superPodMap,
			map[int64]*inferServiceRackInfo{}, map[int32]*inferServiceSPInfo{})
		heap.Push(pq, &inferServicePQItem{superPodID: 9, rackID: 0,
			freeNodes: 100, group: inferServiceGroupOtherSP})
		item := tp.popValidInferServiceSPItem(pq, superPodMap)
		if item == nil {
			t.Fatalf("expected fallback item, got nil")
		}
		if item.superPodID != 0 {
			t.Errorf("expected superPodID=0 after skipping missing SP, got %d", item.superPodID)
		}
	})
	t.Run("empty queue returns nil", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.spBlock = 4
		pq := make(inferServicePQ, 0)
		if item := tp.popValidInferServiceSPItem(&pq, map[int32]superPod{}); item != nil {
			t.Errorf("expected nil for empty queue, got %+v", item)
		}
	})
}

func TestSelectNodesFromRack(t *testing.T) {
	tp := &chip8node8ra64sp{}
	tp.spBlock = 4
	superPodMap := buildSuperPodsByParams(map[int32]int32{0: 16})
	item := &inferServicePQItem{superPodID: 0, rackID: 0}

	nodes := tp.selectNodesFromRack(item, superPodMap)

	if len(nodes) != tp.spBlock {
		t.Fatalf("expected %d nodes, got %d", tp.spBlock, len(nodes))
	}
	if got := len(superPodMap[0]); got != 12 {
		t.Errorf("expected 12 nodes remaining after selection, got %d", got)
	}
	for i, n := range nodes {
		if n.SuperPodID != 0 {
			t.Errorf("node %d: expected superPodID=0, got %d", i, n.SuperPodID)
		}
		if n.RackID != 0 {
			t.Errorf("node %d: expected RackID=0, got %d", i, n.RackID)
		}
	}
}

func TestUpdateInferServiceAffinity(t *testing.T) {
	tp := &chip8node8ra64sp{}
	tp.spBlock = 4
	superPodMap := buildSuperPodsByParams(map[int32]int32{0: 16})
	sameRacks := map[int64]*inferServiceRackInfo{}
	sameSPs := map[int32]*inferServiceSPInfo{}
	item := &inferServicePQItem{superPodID: 0, rackID: 0}

	tp.updateInferServiceAffinity(item, superPodMap, sameRacks, sameSPs)

	info := sameRacks[rackKey(0, 0)]
	if info == nil {
		t.Fatalf("expected sameRacks[rackKey(0,0)] to be recorded")
	}
	if info.rackID != 0 || info.superPodID != 0 {
		t.Errorf("expected rackID=0 superPodID=0, got rackID=%d superPodID=%d", info.rackID, info.superPodID)
	}
	if info.freeNodes != 8 {
		t.Errorf("expected freeNodes=8 for full rack 0, got %d", info.freeNodes)
	}
	if sp := sameSPs[0]; sp == nil || sp.superPodID != 0 {
		t.Errorf("expected sameSPs[0].superPodID=0, got %+v", sp)
	}
}

func TestSelectInferServiceSPForPodLevel(t *testing.T) {
	t.Run("selects nodes for unready sp blocks", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.spBlock = 4
		tp.Name = "my-job"
		superPodMap := buildSuperPodsByParams(map[int32]int32{0: 16})
		selectNodes := make(map[string][]plugin.SuperNode)
		err := tp.selectInferServiceSPForPodLevel([]string{"0", "1"}, superPodMap, selectNodes)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, id := range []string{"0", "1"} {
			if len(selectNodes[id]) != tp.spBlock {
				t.Errorf("sp block %s: expected %d nodes, got %d", id, tp.spBlock, len(selectNodes[id]))
			}
		}
		if got := len(superPodMap[0]); got != 8 {
			t.Errorf("expected 8 free nodes left after selection, got %d", got)
		}
	})
	t.Run("no valid sp block returns error", func(t *testing.T) {
		tp := &chip8node8ra64sp{}
		tp.spBlock = 4
		tp.Name = "my-job"
		superPodMap := buildSuperPodsByParams(map[int32]int32{})
		err := tp.selectInferServiceSPForPodLevel([]string{"0"}, superPodMap,
			make(map[string][]plugin.SuperNode))
		if err == nil {
			t.Errorf("expected error when no valid sp block exists")
		}
	})
}

// newRa64PodLevelEnv builds a chip8node8ra64sp handler and its api nodes for
// pod-level rescheduling tests: one super pod with 16 nodes, spBlock 4.
func newRa64PodLevelEnv() (*chip8node8ra64sp, []*api.NodeInfo) {
	tp := &chip8node8ra64sp{}
	tp.spBlock = 4
	tp.uBMemRackNum = uBMemRackNumber
	tp.NPUJob = &util.NPUJob{ReqNPUNum: 32, SpBlockNPUNum: 32}
	npuNodes, apiNodes := buildPodLevelNodes(16)
	tp.Nodes = npuNodes
	return tp, apiNodes
}

// buildRa64FaultJob builds a fault job with the given super pod blocks and a
// pending session number large enough to skip rack stage 2/3.
func buildRa64FaultJob(spBlocks map[string][]plugin.SuperNode) *rescheduling.FaultJob {
	return &rescheduling.FaultJob{JobUID: "my-job", PendingSessionNum: 8, SuperPods: spBlocks}
}

// buildRa64HealthySpBlock builds a healthy sp block of spBlock nodes on super pod 0 rack 0.
func buildRa64HealthySpBlock(spBlock int) []plugin.SuperNode {
	healthyNodes := make([]plugin.SuperNode, 0, spBlock)
	for i := 0; i < spBlock; i++ {
		healthyNodes = append(healthyNodes, plugin.SuperNode{
			Name: "node-0-" + strconv.Itoa(i), SuperPodID: 0, RackID: 0})
	}
	return healthyNodes
}

func TestSelectNodesForInferServicePodLevel(t *testing.T) {
	t.Run("graceful deletion in progress refuses rescheduling", func(t *testing.T) {
		tp, apiNodes := newRa64PodLevelEnv()
		fJob := buildRa64FaultJob(map[string][]plugin.SuperNode{})
		fJob.FaultTasks = []rescheduling.FaultTask{
			{TaskName: "task0", NodeName: "node-0-0",
				FaultTaskA5Field: rescheduling.FaultTaskA5Field{IsBeingGracefulDeleted: true}}}
		_, err := tp.selectNodesForInferServicePodLevel(&api.TaskInfo{}, apiNodes, fJob)
		if err == nil {
			t.Errorf("expected error when fault task is being graceful deleted")
		}
	})
	t.Run("unready sp blocks fall back to stage 4 priority queue", func(t *testing.T) {
		tp, apiNodes := newRa64PodLevelEnv()
		selectNodes, err := tp.selectNodesForInferServicePodLevel(&api.TaskInfo{Job: "my-job"},
			apiNodes, buildRa64FaultJob(map[string][]plugin.SuperNode{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(selectNodes["0"]) != tp.spBlock {
			t.Errorf("expected %d nodes for sp block 0, got %d", tp.spBlock, len(selectNodes["0"]))
		}
		for _, sn := range selectNodes["0"] {
			if sn.SuperPodID != 0 {
				t.Errorf("expected superPodID=0, got %d", sn.SuperPodID)
			}
		}
	})
	t.Run("healthy sp block is kept in stage 1", func(t *testing.T) {
		tp, apiNodes := newRa64PodLevelEnv()
		want := buildRa64HealthySpBlock(tp.spBlock)
		selectNodes, err := tp.selectNodesForInferServicePodLevel(&api.TaskInfo{Job: "my-job"},
			apiNodes, buildRa64FaultJob(map[string][]plugin.SuperNode{"0": want}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := selectNodes["0"]
		if len(got) != len(want) {
			t.Fatalf("expected %d kept nodes, got %d", len(want), len(got))
		}
		for idx, sn := range got {
			if sn.Name != want[idx].Name {
				t.Errorf("kept node mismatch: expected %s, got %s", want[idx].Name, sn.Name)
			}
		}
	})
}

func TestSelectNodesForInferService_AllItemsSkipped(t *testing.T) {
	tp := &chip8node8ra64sp{}
	tp.spBlock = 4
	tp.uBMemRackNum = 8
	tp.NPUJob = &util.NPUJob{
		SpBlockNPUNum: 32,
		ReqNPUNum:     32,
	}

	superPodMap := buildSuperPodsByParams(map[int32]int32{0: 16})

	sameRacks := map[int64]*inferServiceRackInfo{}
	sameSPs := map[int32]*inferServiceSPInfo{}

	pq := tp.buildInferServicePriorityQueue(superPodMap, sameRacks, sameSPs)

	if pq.Len() == 0 {
		t.Fatal("expected non-empty priority queue")
	}

	for k := range superPodMap {
		delete(superPodMap, k)
	}

	var item *inferServicePQItem
	for pq.Len() > 0 {
		item = heap.Pop(pq).(*inferServicePQItem)
		sp, ok := superPodMap[item.superPodID]
		if !ok || len(sp) < tp.spBlock {
			item = nil
			continue
		}
		rackGroup := transferSuperPodToRackIdMap(sp)
		nodesInRack, rackOk := rackGroup[item.rackID]
		if !rackOk || len(nodesInRack) < tp.spBlock {
			item = nil
			continue
		}
		break
	}

	if item != nil {
		t.Errorf("expected item to be nil when all items are skipped")
	}
}
