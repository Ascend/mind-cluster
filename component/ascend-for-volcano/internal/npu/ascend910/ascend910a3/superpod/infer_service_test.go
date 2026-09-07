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

package superpod

import (
	"fmt"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910a3"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base/inferservice"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// buildSchedulerJob builds a scheduler job carrying the given inferServiceID label and a single
// scheduled super node on the given super pod id, used to mock already-scheduled infer service jobs.
func buildSchedulerJob(inferID string, spID int32) plugin.SchedulerJob {
	return plugin.SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{Label: map[string]string{inferservice.IDLabelKey: inferID}},
			NPUJob: &util.NPUJob{},
		},
		SuperPods: map[string][]plugin.SuperNode{
			"sp0": {{Name: "node10", SuperPodID: spID}},
		},
	}
}

type isInferServiceJobCheckCase struct {
	name   string
	labels map[string]string
	want   bool
	wantID string
}

func buildIsInferServiceJobCheckCases() []isInferServiceJobCheckCase {
	return []isInferServiceJobCheckCase{
		{"01 - nil label returns false", nil, false, ""},
		{"02 - valid value returns true", map[string]string{inferservice.IDLabelKey: "svc-0"}, true, "svc-0"},
	}
}

func TestIsInferServiceJobCheck(t *testing.T) {
	for _, tt := range buildIsInferServiceJobCheckCases() {
		t.Run(tt.name, func(t *testing.T) {
			tp := &module910SuperPod{}
			tp.Label = tt.labels
			got := tp.isInferServiceJobCheck()
			if got != tt.want {
				t.Errorf("isInferServiceJobCheck() = %v, want %v", got, tt.want)
			}
			if got && tp.inferServiceID != tt.wantID {
				t.Errorf("inferServiceID = %s, want %s", tp.inferServiceID, tt.wantID)
			}
		})
	}
}

type selectNodesForInferServiceCase struct {
	name           string
	inferServiceID string
	jobs           map[api.JobID]plugin.SchedulerJob
	nodeInfos      []*api.NodeInfo
	spBlock        int
	reqNPUNum      int
	spBlockNPUNum  int
	wantSPIDs      map[int32]struct{}
}

// buildSelectNodesForInferServiceCases only keeps wrapper-wiring cases here; the selecting
// strategy itself is covered by the shared inferservice package tests.
func buildSelectNodesForInferServiceCases() []selectNodesForInferServiceCase {
	return []selectNodesForInferServiceCase{
		{
			name: "01 - resource enough, prefer same super pod", inferServiceID: "svc-0",
			jobs:      map[api.JobID]plugin.SchedulerJob{"other-job": buildSchedulerJob("svc-0", 1)},
			nodeInfos: []*api.NodeInfo{node0, node1, node10, node11}, spBlock: spBlockNum2,
			reqNPUNum: 16, spBlockNPUNum: 16, wantSPIDs: map[int32]struct{}{1: {}},
		},
		{
			name: "02 - same super pod insufficient, fallback to other super pod", inferServiceID: "svc-0",
			jobs:      map[api.JobID]plugin.SchedulerJob{"other-job": buildSchedulerJob("svc-0", 0)},
			nodeInfos: []*api.NodeInfo{node0, node1, node10, node11, node12, node13}, spBlock: 4,
			reqNPUNum: 64, spBlockNPUNum: 64, wantSPIDs: map[int32]struct{}{1: {}},
		},
	}
}

func TestSelectNodesForInferService(t *testing.T) {
	for _, cs := range buildSelectNodesForInferServiceCases() {
		t.Run(cs.name, func(t *testing.T) {
			tp := &module910SuperPod{}
			tp.Name = "my-job"
			tp.MaxNodeNPUNum = ascend910a3.NodeNPUNumber16
			tp.inferServiceID = cs.inferServiceID
			tp.spBlock = cs.spBlock
			tp.NPUJob = &util.NPUJob{ReqNPUNum: cs.reqNPUNum, SpBlockNPUNum: cs.spBlockNPUNum}
			tp.ScheduleEnv = plugin.ScheduleEnv{ClusterCache: plugin.ClusterCache{Jobs: cs.jobs}}
			tp.Nodes = newNPUNodes(npuNodes, superPodSize10)
			task := &api.TaskInfo{Job: "my-job", Name: "task0"}
			selectedNodes, err := tp.selectNodesForInferService(task, cs.nodeInfos)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got := getSelectedNodesSuperPodID(selectedNodes); !reflect.DeepEqual(got, cs.wantSPIDs) {
				t.Errorf("selectedSPIDs = %v, want %v", got, cs.wantSPIDs)
			}
		})
	}
}

// buildPodLevelJob builds a scheduler job whose pod-level rescheduling is enabled.
func buildPodLevelJob() plugin.SchedulerJob {
	return plugin.SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{
			NPUJob: &util.NPUJob{SchedulingTaskNum: 1, Tasks: newNPUTasks(2)},
		},
	}
}

// buildSuperPodTop builds a super pod topology with spNodeCount nodes on each given super pod id.
func buildSuperPodTop(spIDs []int32, spNodeCount int) map[int32]superPod {
	totalNodes := make(map[int32]superPod)
	for _, id := range spIDs {
		sp := make(superPod)
		for i := 0; i < spNodeCount; i++ {
			name := fmt.Sprintf("node-%d-%d", id, i)
			sp[name] = plugin.NPUNode{CommonNode: plugin.CommonNode{Name: name, SuperPodID: id}}
		}
		totalNodes[id] = sp
	}
	return totalNodes
}

// newPodLevelHandler builds a superpod handler with pod-level rescheduling enabled.
func newPodLevelHandler(jobs map[api.JobID]plugin.SchedulerJob) *module910SuperPod {
	tp := &module910SuperPod{}
	tp.Name = "my-job"
	tp.spBlock = 2
	tp.inferServiceID = "svc-0"
	tp.NPUJob = &util.NPUJob{NPUTaskNum: 2}
	tp.ScheduleEnv = plugin.ScheduleEnv{ClusterCache: plugin.ClusterCache{
		Jobs: jobs, SuperPodInfo: plugin.NewSuperPodInfo()}}
	// Nodes is promoted from ClusterCache: assign it after ScheduleEnv, otherwise the
	// whole-struct assignment above overwrites the node map and getSuperPodTop sees none.
	tp.Nodes = newNPUNodes(4, 2)
	return tp
}

// newPodLevelFaultJob builds a fault job of the current job with the given super pod blocks.
func newPodLevelFaultJob(spBlocks map[string][]plugin.SuperNode) *rescheduling.FaultJob {
	return &rescheduling.FaultJob{JobUID: "my-job", SuperPods: spBlocks}
}

func TestSelectNodesForInferServicePodLevel(t *testing.T) {
	t.Run("healthy sp block is kept in stage 1", func(t *testing.T) {
		tp := newPodLevelHandler(map[api.JobID]plugin.SchedulerJob{"my-job": buildPodLevelJob()})
		fJob := newPodLevelFaultJob(map[string][]plugin.SuperNode{"0": {
			{Name: "node0", SuperPodID: 0}, {Name: "node1", SuperPodID: 0}}})
		selectNodes, err := tp.selectNodesForInferServicePodLevel(
			&api.TaskInfo{Job: "my-job"}, []*api.NodeInfo{&api.NodeInfo{Name: "node0"},
				&api.NodeInfo{Name: "node1"}, &api.NodeInfo{Name: "node2"}, &api.NodeInfo{Name: "node3"}}, fJob)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := selectNodes["0"]
		if len(got) != 2 || got[0].Name != "node0" || got[1].Name != "node1" {
			t.Errorf("expected kept nodes [node0 node1], got %v", got)
		}
	})
	t.Run("unready sp block falls back to stage 4 same-service super pod", func(t *testing.T) {
		jobs := map[api.JobID]plugin.SchedulerJob{
			"my-job": buildPodLevelJob(), "other-job": buildSchedulerJob("svc-0", 1)}
		tp := newPodLevelHandler(jobs)
		fJob := newPodLevelFaultJob(map[string][]plugin.SuperNode{})
		selectNodes, err := tp.selectNodesForInferServicePodLevel(
			&api.TaskInfo{Job: "my-job"}, []*api.NodeInfo{&api.NodeInfo{Name: "node0"},
				&api.NodeInfo{Name: "node1"}, &api.NodeInfo{Name: "node2"}, &api.NodeInfo{Name: "node3"}}, fJob)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(selectNodes["0"]) != 2 {
			t.Fatalf("expected 2 nodes for sp block 0, got %d", len(selectNodes["0"]))
		}
		for _, sn := range selectNodes["0"] {
			if sn.SuperPodID != 1 {
				t.Errorf("expected superPodID=1 from same-service super pod, got %d", sn.SuperPodID)
			}
		}
	})
	t.Run("no available super pod returns error", func(t *testing.T) {
		tp := newPodLevelHandler(map[api.JobID]plugin.SchedulerJob{"my-job": buildPodLevelJob()})
		tp.Nodes = map[string]plugin.NPUNode{}
		fJob := newPodLevelFaultJob(map[string][]plugin.SuperNode{})
		_, err := tp.selectNodesForInferServicePodLevel(
			&api.TaskInfo{Job: "my-job"}, []*api.NodeInfo{&api.NodeInfo{Name: "node0"},
				&api.NodeInfo{Name: "node1"}}, fJob)
		if err == nil {
			t.Errorf("expected error when no super pod is available")
		}
	})
}

func TestSelectInferServiceSPForPodLevel(t *testing.T) {
	t.Run("same-service super pod is preferred for unready ids", func(t *testing.T) {
		tp := newPodLevelHandler(map[api.JobID]plugin.SchedulerJob{
			"other-job": buildSchedulerJob("svc-0", 1)})
		totalNodes := buildSuperPodTop([]int32{0, 1}, 2)
		selectNodes := make(map[string][]plugin.SuperNode)
		vSuperPodID := map[string]bool{"0": false, "1": false}
		err := inferservice.SelectInferServiceSPForPodLevel(tp.ScheduleEnv.Jobs, tp.Name,
			tp.inferServiceID, tp.spBlock, []string{"0", "1"}, totalNodes, selectNodes, vSuperPodID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, sn := range selectNodes["0"] {
			if sn.SuperPodID != 1 {
				t.Errorf("expected superPodID=1 for id 0, got %d", sn.SuperPodID)
			}
		}
		for _, sn := range selectNodes["1"] {
			if sn.SuperPodID != 0 {
				t.Errorf("expected superPodID=0 for id 1, got %d", sn.SuperPodID)
			}
		}
		if !vSuperPodID["0"] || !vSuperPodID["1"] {
			t.Errorf("expected all ids marked ready, got %v", vSuperPodID)
		}
	})
	t.Run("no valid super pod returns error", func(t *testing.T) {
		tp := newPodLevelHandler(map[api.JobID]plugin.SchedulerJob{})
		totalNodes := buildSuperPodTop([]int32{0}, 1)
		err := inferservice.SelectInferServiceSPForPodLevel(tp.ScheduleEnv.Jobs, tp.Name,
			tp.inferServiceID, tp.spBlock, []string{"0"}, totalNodes,
			make(map[string][]plugin.SuperNode), map[string]bool{"0": false})
		if err == nil {
			t.Errorf("expected error when no valid super pod exists")
		}
	})
	t.Run("empty unready ids is a no-op", func(t *testing.T) {
		tp := newPodLevelHandler(map[api.JobID]plugin.SchedulerJob{})
		totalNodes := buildSuperPodTop([]int32{0}, 2)
		selectNodes := make(map[string][]plugin.SuperNode)
		if err := inferservice.SelectInferServiceSPForPodLevel(tp.ScheduleEnv.Jobs, tp.Name,
			tp.inferServiceID, tp.spBlock, nil, totalNodes, selectNodes,
			map[string]bool{}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(selectNodes) != 0 {
			t.Errorf("expected no selection, got %v", selectNodes)
		}
	})
}
