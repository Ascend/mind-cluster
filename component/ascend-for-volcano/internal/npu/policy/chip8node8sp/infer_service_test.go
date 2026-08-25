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
	"fmt"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base/inferservice"
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
		tp.Label = map[string]string{inferservice.IDLabelKey: ""}
		if tp.isInferServiceJobCheck() {
			t.Errorf("expected false when inferServiceIDLabelKey is empty")
		}
	})

	t.Run("Label with valid inferServiceIDLabelKey should return true", func(t *testing.T) {
		tp := &chip8node8sp{}
		tp.Label = map[string]string{inferservice.IDLabelKey: "test-infer-id"}
		if !tp.isInferServiceJobCheck() {
			t.Errorf("expected true when inferServiceIDLabelKey has value")
		}
		if tp.inferServiceID != "test-infer-id" {
			t.Errorf("expected inferServiceID=test-infer-id, got %s", tp.inferServiceID)
		}
	})
}

// buildScheduledJob builds a scheduler job carrying the given inferServiceID label and a
// single scheduled super node on the given super pod id, used to mock already-scheduled
// infer service jobs.
func buildScheduledJob(inferID string, spID int32) plugin.SchedulerJob {
	return plugin.SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{Label: map[string]string{inferservice.IDLabelKey: inferID}},
			NPUJob: &util.NPUJob{},
		},
		SuperPods: map[string][]plugin.SuperNode{
			"sp0": {{Name: "node4", SuperPodID: spID}},
		},
	}
}

// buildInferNPUNodes builds NPU nodes named node<start>..node<end-1>, grouped into super
// pods of spSize nodes each (SuperPodID = index / spSize).
func buildInferNPUNodes(start, end, spSize int) map[string]plugin.NPUNode {
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
					util.AcceleratorTypeKeyDeprecated: AcceleratorType,
				},
			},
		}
	}
	return nodes
}

func buildInferNodeInfos(start, end int) []*api.NodeInfo {
	var nodeInfos []*api.NodeInfo
	for i := start; i < end; i++ {
		nodeInfos = append(nodeInfos, &api.NodeInfo{Name: fmt.Sprintf("node%d", i)})
	}
	return nodeInfos
}

type selectNodesForInferServiceCase struct {
	name          string
	jobs          map[api.JobID]plugin.SchedulerJob
	nodeInfos     []*api.NodeInfo
	spBlock       int
	reqNPUNum     int
	spBlockNPUNum int
	wantErr       bool
	wantSPIDs     map[int32]struct{}
}

// buildSelectNodesForInferServiceCases only keeps wrapper-wiring cases here; the selecting
// strategy itself is covered by the shared inferservice package tests.
func buildSelectNodesForInferServiceCases() []selectNodesForInferServiceCase {
	return []selectNodesForInferServiceCase{
		{
			name:          "01 - resource enough, prefer same super pod",
			jobs:          map[api.JobID]plugin.SchedulerJob{"other-job": buildScheduledJob("svc-0", 1)},
			nodeInfos:     buildInferNodeInfos(0, 8),
			spBlock:       2,
			reqNPUNum:     16,
			spBlockNPUNum: 16,
			wantSPIDs:     map[int32]struct{}{1: {}},
		},
		{
			name:      "02 - same super pod insufficient, fallback to other super pod",
			jobs:      map[api.JobID]plugin.SchedulerJob{"other-job": buildScheduledJob("svc-0", 0)},
			nodeInfos: []*api.NodeInfo{{Name: "node0"}, {Name: "node1"}, {Name: "node4"}, {Name: "node5"}, {Name: "node6"}, {Name: "node7"}},
			spBlock:   4,
			reqNPUNum: 64, spBlockNPUNum: 64,
			wantSPIDs: map[int32]struct{}{1: {}},
		},
	}
}

func TestSelectNodesForInferService(t *testing.T) {
	for _, cs := range buildSelectNodesForInferServiceCases() {
		t.Run(cs.name, func(t *testing.T) {
			tp := &chip8node8sp{}
			tp.Name = "my-job"
			tp.spBlock = cs.spBlock
			tp.NPUJob = &util.NPUJob{ReqNPUNum: cs.reqNPUNum, SpBlockNPUNum: cs.spBlockNPUNum}
			tp.inferServiceID = "svc-0"
			tp.ScheduleEnv = plugin.ScheduleEnv{ClusterCache: plugin.ClusterCache{Jobs: cs.jobs}}
			tp.Nodes = buildInferNPUNodes(0, 8, 4)
			task := &api.TaskInfo{Job: "my-job", Name: "task0"}
			selectedNodes, err := tp.selectNodesForInferService(task, cs.nodeInfos)
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
