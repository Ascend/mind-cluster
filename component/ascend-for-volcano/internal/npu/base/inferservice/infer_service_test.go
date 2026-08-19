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

package inferservice

import (
	"container/heap"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// buildSchedulerJob builds a scheduler job carrying the given inferServiceID label and
// a single scheduled super node on the given super pod id, used to mock
// already-scheduled infer service jobs.
func buildSchedulerJob(inferID string, spID int32) plugin.SchedulerJob {
	return plugin.SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{Label: map[string]string{IDLabelKey: inferID}},
			NPUJob: &util.NPUJob{},
		},
		SuperPods: map[string][]plugin.SuperNode{
			"sp0": {{Name: "node10", SuperPodID: spID}},
		},
	}
}

// buildSuperPodTop builds a super pod topology from a superPodID -> node names spec,
// with each NPUNode carrying its real SuperPodID so that selected SuperNode entries
// can be verified by their super pod id.
func buildSuperPodTop(spec map[int32][]string) map[int32]plugin.SuperPod {
	top := make(map[int32]plugin.SuperPod, len(spec))
	for spID, nodeNames := range spec {
		sp := make(plugin.SuperPod, len(nodeNames))
		for _, name := range nodeNames {
			sp[name] = plugin.NPUNode{CommonNode: plugin.CommonNode{Name: name, SuperPodID: spID}}
		}
		top[spID] = sp
	}
	return top
}

type getInferServiceIDCase struct {
	name   string
	labels map[string]string
	want   string
}

func buildGetInferServiceIDCases() []getInferServiceIDCase {
	return []getInferServiceIDCase{
		{"01 - nil label returns empty", nil, ""},
		{"02 - label without key returns empty", map[string]string{"other": "v"}, ""},
		{"03 - empty value returns empty", map[string]string{IDLabelKey: ""}, ""},
		{"04 - valid value returned", map[string]string{IDLabelKey: "svc-0"}, "svc-0"},
	}
}

func TestGetInferServiceID(t *testing.T) {
	for _, tt := range buildGetInferServiceIDCases() {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetInferServiceID(tt.labels); got != tt.want {
				t.Errorf("GetInferServiceID() = %q, want %q", got, tt.want)
			}
		})
	}
}

type collectScheduledSPsCase struct {
	name           string
	inferServiceID string
	currentJob     api.JobID
	jobs           map[api.JobID]plugin.SchedulerJob
	want           map[int32]struct{}
}

func buildCollectScheduledSPsCases() []collectScheduledSPsCase {
	return []collectScheduledSPsCase{
		{
			name: "01 - nil Jobs returns empty", inferServiceID: "svc-0", jobs: nil,
			want: map[int32]struct{}{},
		},
		{
			name: "02 - same inferServiceID scheduled SP collected", inferServiceID: "svc-0",
			currentJob: "my-job", jobs: map[api.JobID]plugin.SchedulerJob{"other-job": buildSchedulerJob("svc-0", 1)},
			want: map[int32]struct{}{1: {}},
		},
		{
			name: "03 - current job skipped", inferServiceID: "svc-0", currentJob: "my-job",
			jobs: map[api.JobID]plugin.SchedulerJob{"my-job": buildSchedulerJob("svc-0", 1)},
			want: map[int32]struct{}{},
		},
		{
			name: "04 - different inferServiceID skipped", inferServiceID: "svc-0", currentJob: "my-job",
			jobs: map[api.JobID]plugin.SchedulerJob{"other-job": buildSchedulerJob("svc-1", 1)},
			want: map[int32]struct{}{},
		},
		{
			name: "05 - job without SuperPods skipped", inferServiceID: "svc-0", currentJob: "my-job",
			jobs: map[api.JobID]plugin.SchedulerJob{
				"other-job": {SchedulerJobAttr: util.SchedulerJobAttr{
					ComJob: util.ComJob{Label: map[string]string{IDLabelKey: "svc-0"}}}},
			},
			want: map[int32]struct{}{},
		},
		{
			name: "06 - job without label skipped", inferServiceID: "svc-0", currentJob: "my-job",
			jobs: map[api.JobID]plugin.SchedulerJob{
				"other-job": {SchedulerJobAttr: util.SchedulerJobAttr{ComJob: util.ComJob{}}},
			},
			want: map[int32]struct{}{},
		},
	}
}

func TestCollectScheduledSPs(t *testing.T) {
	for _, tt := range buildCollectScheduledSPsCases() {
		t.Run(tt.name, func(t *testing.T) {
			got := CollectScheduledSPs(tt.jobs, tt.currentJob, tt.inferServiceID)
			if len(got) != len(tt.want) {
				t.Errorf("CollectScheduledSPs() len = %d, want %d", len(got), len(tt.want))
				return
			}
			for id := range tt.want {
				if _, ok := got[id]; !ok {
					t.Errorf("expected superPodID %d in result", id)
				}
			}
		})
	}
}

func TestEnrichSPInfo(t *testing.T) {
	t.Run("fill FreeNodeNum for sameSPs", func(t *testing.T) {
		top := map[int32]plugin.SuperPod{0: {"node0": {}, "node1": {}, "node2": {}}, 1: {"node3": {}}}
		sameSPs := map[int32]*SPInfo{0: {SuperPodID: 0}}
		EnrichSPInfo(top, sameSPs)
		if sameSPs[0].FreeNodeNum != 3 {
			t.Errorf("FreeNodeNum = %d, want 3", sameSPs[0].FreeNodeNum)
		}
	})
	t.Run("empty top not panic", func(t *testing.T) {
		EnrichSPInfo(map[int32]plugin.SuperPod{}, map[int32]*SPInfo{})
	})
}

func TestPQ_Less(t *testing.T) {
	t.Run("lower group first", func(t *testing.T) {
		pq := PQ{
			{Group: GroupSameSP, FreeNodes: 1},
			{Group: GroupOtherSP, FreeNodes: 10},
		}
		if !pq.Less(0, 1) {
			t.Errorf("expected sameSP less than otherSP")
		}
	})
	t.Run("same group more FreeNodes first", func(t *testing.T) {
		pq := PQ{
			{Group: GroupSameSP, FreeNodes: 5},
			{Group: GroupSameSP, FreeNodes: 3},
		}
		if !pq.Less(0, 1) {
			t.Errorf("expected more FreeNodes first")
		}
	})
}

func TestPQ_Swap(t *testing.T) {
	pq := PQ{{SuperPodID: 0, index: 0}, {SuperPodID: 1, index: 1}}
	pq.Swap(0, 1)
	if pq[0].SuperPodID != 1 || pq[1].SuperPodID != 0 {
		t.Errorf("expected items swapped")
	}
	if pq[0].index != 0 || pq[1].index != 1 {
		t.Errorf("expected indices updated")
	}
}

// TestPQ_Defensive covers the defensive branches of Push and Pop: non-PQItem
// items are ignored and popping an empty queue returns nil.
func TestPQ_Defensive(t *testing.T) {
	pq := make(PQ, 0)
	pq.Push("not-a-pqitem")
	if pq.Len() != 0 {
		t.Errorf("expected non-PQItem push ignored, got len %d", pq.Len())
	}
	if item := pq.Pop(); item != nil {
		t.Errorf("expected nil popping empty queue, got %v", item)
	}
}

func TestPQ_FullHeapOrdering(t *testing.T) {
	pq := make(PQ, 0)
	heap.Init(&pq)
	heap.Push(&pq, &PQItem{SuperPodID: 2, FreeNodes: 8, Group: GroupOtherSP})
	heap.Push(&pq, &PQItem{SuperPodID: 1, FreeNodes: 5, Group: GroupSameSP})
	first := heap.Pop(&pq).(*PQItem)
	if first.Group != GroupSameSP {
		t.Errorf("expected first pop to be sameSP, got group=%d", first.Group)
	}
	second := heap.Pop(&pq).(*PQItem)
	if second.Group != GroupOtherSP {
		t.Errorf("expected second pop to be otherSP, got group=%d", second.Group)
	}
}

type buildPriorityQueueCase struct {
	name       string
	spBlock    int
	top        map[int32]plugin.SuperPod
	sameSPs    map[int32]*SPInfo
	wantGroups map[int]int
}

func buildPriorityQueueCases() []buildPriorityQueueCase {
	return []buildPriorityQueueCase{
		{
			name: "01 - empty maps return empty queue", spBlock: 2, top: map[int32]plugin.SuperPod{},
			sameSPs: map[int32]*SPInfo{}, wantGroups: map[int]int{},
		},
		{
			name: "02 - sameSP with enough nodes in queue", spBlock: 2,
			top:        map[int32]plugin.SuperPod{0: {"node0": {}, "node1": {}, "node2": {}}},
			sameSPs:    map[int32]*SPInfo{0: {SuperPodID: 0, FreeNodeNum: 3}},
			wantGroups: map[int]int{GroupSameSP: 1},
		},
		{
			name: "03 - sameSP with insufficient nodes excluded", spBlock: 4,
			top:        map[int32]plugin.SuperPod{0: {"node0": {}, "node1": {}}},
			sameSPs:    map[int32]*SPInfo{0: {SuperPodID: 0, FreeNodeNum: 2}},
			wantGroups: map[int]int{},
		},
		{
			name: "04 - otherSP with enough nodes in queue", spBlock: 2,
			top:     map[int32]plugin.SuperPod{0: {"node0": {}, "node1": {}, "node2": {}}},
			sameSPs: map[int32]*SPInfo{}, wantGroups: map[int]int{GroupOtherSP: 1},
		},
		{
			name: "05 - sameSP and otherSP both in queue, no duplicate", spBlock: 2,
			top: map[int32]plugin.SuperPod{
				0: {"node0": {}, "node1": {}, "node2": {}},
				1: {"node3": {}, "node4": {}, "node5": {}},
			},
			sameSPs:    map[int32]*SPInfo{0: {SuperPodID: 0, FreeNodeNum: 3}},
			wantGroups: map[int]int{GroupSameSP: 1, GroupOtherSP: 1},
		},
	}
}

func TestBuildPriorityQueue(t *testing.T) {
	for _, tt := range buildPriorityQueueCases() {
		t.Run(tt.name, func(t *testing.T) {
			pq := buildPriorityQueue(tt.top, tt.sameSPs, tt.spBlock)
			got := map[int]int{}
			for pq.Len() > 0 {
				item := heap.Pop(pq).(*PQItem)
				got[item.Group]++
			}
			if !reflect.DeepEqual(got, tt.wantGroups) {
				t.Errorf("groups = %v, want %v", got, tt.wantGroups)
			}
		})
	}
}

type selectNodesCase struct {
	name           string
	inferServiceID string
	jobs           map[api.JobID]plugin.SchedulerJob
	superPodTop    map[int32]plugin.SuperPod
	spBlock        int
	reqNPUNum      int
	spBlockNPUNum  int
	wantErr        bool
	wantSPIDs      map[int32]struct{}
}

func buildSelectNodesCases() []selectNodesCase {
	return []selectNodesCase{
		{
			name: "01 - resource enough, prefer same super pod", inferServiceID: "svc-0",
			jobs: map[api.JobID]plugin.SchedulerJob{"other-job": buildSchedulerJob("svc-0", 1)},
			superPodTop: buildSuperPodTop(map[int32][]string{
				0: {"node0", "node1"},
				1: {"node10", "node11"},
			}),
			spBlock: 2, reqNPUNum: 16, spBlockNPUNum: 16, wantSPIDs: map[int32]struct{}{1: {}},
		},
		{
			name: "02 - first round without same infer service, otherSP selected", inferServiceID: "svc-0",
			jobs:        map[api.JobID]plugin.SchedulerJob{},
			superPodTop: buildSuperPodTop(map[int32][]string{0: {"node0", "node1"}}),
			spBlock:     2, reqNPUNum: 16, spBlockNPUNum: 16, wantSPIDs: map[int32]struct{}{0: {}},
		},
		{
			name: "03 - same super pod insufficient, fallback to other super pod", inferServiceID: "svc-0",
			jobs: map[api.JobID]plugin.SchedulerJob{"other-job": buildSchedulerJob("svc-0", 0)},
			superPodTop: buildSuperPodTop(map[int32][]string{
				0: {"node0", "node1"},
				1: {"node10", "node11", "node12", "node13"},
			}),
			spBlock: 4, reqNPUNum: 64, spBlockNPUNum: 64, wantSPIDs: map[int32]struct{}{1: {}},
		},
		{
			name: "04 - second sp-block falls back to other super pod", inferServiceID: "svc-0",
			jobs: map[api.JobID]plugin.SchedulerJob{"other-job": buildSchedulerJob("svc-0", 0)},
			superPodTop: buildSuperPodTop(map[int32][]string{
				0: {"node0", "node1", "node2"},
				1: {"node10", "node11", "node12", "node13"},
			}),
			spBlock: 2, reqNPUNum: 32, spBlockNPUNum: 16, wantSPIDs: map[int32]struct{}{0: {}, 1: {}},
		},
		{
			name: "05 - invalid spBlock returns error", spBlock: 0,
			superPodTop: buildSuperPodTop(map[int32][]string{0: {"node0"}}), wantErr: true,
		},
		{
			name: "06 - invalid spBlockNPUNum returns error", spBlock: 2, spBlockNPUNum: 0,
			superPodTop: buildSuperPodTop(map[int32][]string{0: {"node0"}}), wantErr: true,
		},
		{
			name: "07 - resources insufficient returns error", inferServiceID: "svc-0",
			jobs:        map[api.JobID]plugin.SchedulerJob{},
			superPodTop: buildSuperPodTop(map[int32][]string{0: {"node0"}}),
			spBlock:     2, reqNPUNum: 16, spBlockNPUNum: 16, wantErr: true,
		},
	}
}

func TestSelectNodesForInferService(t *testing.T) {
	for _, cs := range buildSelectNodesCases() {
		t.Run(cs.name, func(t *testing.T) {
			selectedNodes, err := SelectNodesForInferService(InferServiceReq{
				Jobs:           cs.jobs,
				JobName:        api.JobID("my-job"),
				InferServiceID: cs.inferServiceID,
				SpBlock:        cs.spBlock,
				ReqNPUNum:      cs.reqNPUNum,
				SpBlockNPUNum:  cs.spBlockNPUNum,
				SuperPodTop:    cs.superPodTop,
			})
			if cs.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got := getSelectedSPIDs(selectedNodes); !reflect.DeepEqual(got, cs.wantSPIDs) {
				t.Errorf("selectedSPIDs = %v, want %v", got, cs.wantSPIDs)
			}
		})
	}
}

func getSelectedSPIDs(selectedNodes map[string][]plugin.SuperNode) map[int32]struct{} {
	spIDs := make(map[int32]struct{})
	for _, nodes := range selectedNodes {
		for _, node := range nodes {
			spIDs[node.SuperPodID] = struct{}{}
		}
	}
	return spIDs
}
