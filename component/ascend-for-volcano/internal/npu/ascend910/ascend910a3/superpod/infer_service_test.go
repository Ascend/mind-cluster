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
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910a3"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base/inferservice"
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
