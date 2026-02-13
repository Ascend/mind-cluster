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

// Package chip4nodex is using for HuaWei 300I A5 affinity schedule.
package chip4nodex

import (
	"reflect"
	"strings"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910a5"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
)

const requiredNodesForTestThree = 3
const requiredNodesForTestFour = 4

// TestJudgeNodeAndTaskNPU test JudgeNodeAndTaskNPU
func TestJudgeNodeAndTaskNPU(t *testing.T) {
	mod := &chip4nodex{}
	if err := mod.judgeNodeAndTaskNPU(requiredNodesForTestThree, []int{0, 1, 2}); err != nil {
		t.Errorf("expected no error when required=3 and nodeTop length is 3, got: %v", err)
	}
	if err := mod.judgeNodeAndTaskNPU(requiredNodesForTestFour, []int{0, 1, 2}); err == nil {
		t.Error("expected error when required=4 and nodeTop length is 3, got nil")
	} else if !strings.Contains(err.Error(), "not meet task in 4Pmesh") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestCheckJobModeNoTasks checkJobMode: NPUTaskNum=0 / per-task invalid / per-task valid
func TestCheckJobModeNoTasks(t *testing.T) {
	tp := &chip4nodex{
		Base910A5: ascend910a5.Base910A5{NPUHandler: base.NPUHandler{
			MaxNodeNPUNum:    8,
			SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{}},
		}},
	}
	tp.Name = "jobA"
	tp.NPUTaskNum = 0
	tp.ReqNPUNum = 0

	err := tp.checkJobMode()
	if err == nil || !strings.Contains(err.Error(), "no npu job") {
		t.Errorf("0 test should be ‘no npu job’，actually: %v", err)
	}
}

func TestCheckJobModeValid(t *testing.T) {
	tp := &chip4nodex{
		Base910A5: ascend910a5.Base910A5{NPUHandler: base.NPUHandler{
			MaxNodeNPUNum:    8,
			SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{}},
		}},
	}
	tp.Name = "jobC"
	tp.NPUTaskNum = 2
	tp.ReqNPUNum = 4

	if err := tp.checkJobMode(); err != nil {
		t.Errorf("Legal per-task should not report an error，actually: %v", err)
	}
}

// TestValid800ia5NPUJobWarningsAndModeError SpBlock/TpBlock warning + mode error
func TestValid800ia5NPUJobWarningsAndModeError(t *testing.T) {
	tp := &chip4nodex{
		Base910A5: ascend910a5.Base910A5{NPUHandler: base.NPUHandler{
			MaxNodeNPUNum: 5,
			SchedulerJobAttr: util.SchedulerJobAttr{
				NPUJob: &util.NPUJob{
					SpBlockNPUNum: 1,
					TpBlockNPUNum: util.LeastTpBlock + 1}}}},
	}
	tp.Name = "jobD"
	tp.NPUTaskNum = 0
	tp.ReqNPUNum = 0
	res := tp.ValidNPUJob()
	if res == nil {
		t.Fatal("Expected to return ValidateResult, but actually nil")
	}
	if res.Pass {
		t.Error("When the mode is incorrect, Pass should be false")
	}
	if !strings.Contains(res.Reason, "InvalidResourceReques") {
		t.Errorf("Reason should include ‘InvalidResourceReques’，actually: %q", res.Reason)
	}
}

// TestValid800ia5NPUJobSuccess Normal execution: no warnings, checkJobMode passed
func TestValid800ia5NPUJobSuccess(t *testing.T) {
	tp := &chip4nodex{
		Base910A5: ascend910a5.Base910A5{NPUHandler: base.NPUHandler{
			MaxNodeNPUNum: 10,
			SchedulerJobAttr: util.SchedulerJobAttr{
				NPUJob: &util.NPUJob{
					SpBlockNPUNum: 0,
					TpBlockNPUNum: util.LeastTpBlock}}}},
	}
	tp.Name = "jobE"
	tp.NPUTaskNum = 2
	tp.ReqNPUNum = 4
	if got := tp.ValidNPUJob(); got != nil {
		t.Errorf("The normal path should return nil, but actually: %v", got)
	}
}

type getNodeMeshInfoTestCase struct {
	name           string
	nodeTop        []int
	wantCardCount  []int
	wantFullMeshes int
}

func getNodeMeshInfoTestCases() []getNodeMeshInfoTestCase {
	return []getNodeMeshInfoTestCase{
		{
			name:           "single card",
			nodeTop:        []int{0},
			wantCardCount:  []int{1, 0, 0, 0},
			wantFullMeshes: 0,
		},
		{
			name:           "one mesh partially filled",
			nodeTop:        []int{0, 1},
			wantCardCount:  []int{2, 0, 0, 0},
			wantFullMeshes: 0,
		},
		{
			name:           "one mesh full",
			nodeTop:        []int{0, 1, 2, 3},
			wantCardCount:  []int{4, 0, 0, 0},
			wantFullMeshes: 1,
		},
		{
			name:           "all meshes full",
			nodeTop:        []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			wantCardCount:  []int{4, 4, 4, 4},
			wantFullMeshes: 4,
		},
		{
			name:           "spread across meshes",
			nodeTop:        []int{0, 4, 8, 12},
			wantCardCount:  []int{1, 1, 1, 1},
			wantFullMeshes: 0,
		},
		{
			name:           "three meshes full, one partial",
			nodeTop:        []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}, // 缺 15
			wantCardCount:  []int{4, 4, 4, 3},
			wantFullMeshes: 3,
		},
	}
}

func TestGetNodeMeshInfo(t *testing.T) {
	tests := getNodeMeshInfoTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCardCount, gotFullMeshes := getNodeMeshInfo(tt.nodeTop)
			if !reflect.DeepEqual(gotCardCount, tt.wantCardCount) {
				t.Errorf("got cardCount = %v, want %v", gotCardCount, tt.wantCardCount)
			}
			if gotFullMeshes != tt.wantFullMeshes {
				t.Errorf("got fullMeshes = %v, want %v", gotFullMeshes, tt.wantFullMeshes)
			}
		})
	}
}

type getAvailableMeshInNodeTestCase struct {
	name              string
	taskNPUNum        int
	nodeTop           []int
	wantNeedMesh      int
	wantAvailableMesh int
}

func getAvailableNotFullMeshCases() []getAvailableMeshInNodeTestCase {
	return []getAvailableMeshInNodeTestCase{
		{
			name:              "task < CardsNumPerMesh, enough in one mesh",
			taskNPUNum:        2,
			nodeTop:           []int{0, 1, 2, 3}, // mesh0 = 4
			wantNeedMesh:      1,
			wantAvailableMesh: 1,
		},
		{
			name:              "task < CardsNumPerMesh, spread out, not enough in any mesh",
			taskNPUNum:        3,
			nodeTop:           []int{0, 4, 8, 12}, // only one card for each mesh
			wantNeedMesh:      1,
			wantAvailableMesh: 0,
		},
		{
			name:              "task < CardsNumPerMesh, multiple meshes satisfy",
			taskNPUNum:        1,
			nodeTop:           []int{0, 1, 4, 5}, // mesh0=2, mesh1=2
			wantNeedMesh:      1,
			wantAvailableMesh: 2,
		},
	}
}

func getAvailableFullMeshCases() []getAvailableMeshInNodeTestCase {
	return []getAvailableMeshInNodeTestCase{
		{
			name:              "task >= CardsNumPerMesh, exact one mesh full",
			taskNPUNum:        4,
			nodeTop:           []int{0, 1, 2, 3}, // mesh0 = full
			wantNeedMesh:      1,
			wantAvailableMesh: 1,
		},
		{
			name:              "task >= CardsNumPerMesh, two meshes full",
			taskNPUNum:        8,
			nodeTop:           []int{0, 1, 2, 3, 4, 5, 6, 7}, // mesh0=4, mesh1=4
			wantNeedMesh:      2,
			wantAvailableMesh: 2,
		},
		{
			name:              "task >= CardsNumPerMesh, partial meshes not enough",
			taskNPUNum:        4,
			nodeTop:           []int{0, 1, 4, 5, 8}, // mesh0=2, mesh1=2, mesh2=1
			wantNeedMesh:      1,
			wantAvailableMesh: 0,
		},
		{
			name:              "all meshes full, big task",
			taskNPUNum:        16,
			nodeTop:           []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			wantNeedMesh:      4,
			wantAvailableMesh: 4,
		},
	}
}

func TestGetAvailableMeshInNode(t *testing.T) {
	tests := getAvailableNotFullMeshCases()
	tests = append(tests, getAvailableFullMeshCases()...)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNeedMesh, gotAvailableMesh := getAvailableMeshInNode(tt.taskNPUNum, tt.nodeTop)
			if gotNeedMesh != tt.wantNeedMesh || gotAvailableMesh != tt.wantAvailableMesh {
				t.Errorf("getAvailableMeshInNode(%d, %v) = (%d, %d), want (%d, %d)",
					tt.taskNPUNum, tt.nodeTop, gotNeedMesh, gotAvailableMesh,
					tt.wantNeedMesh, tt.wantAvailableMesh)
			}
		})
	}
}

type judgeNodeAndTaskNpuIn4PmeshTestCase struct {
	name    string
	taskNPU int
	nodeTop []int
	want    bool
}

func getNotFullMeshTestCases() []judgeNodeAndTaskNpuIn4PmeshTestCase {
	return []judgeNodeAndTaskNpuIn4PmeshTestCase{
		{
			name:    "task < 4, enough cards in one mesh",
			taskNPU: 2,
			nodeTop: []int{0, 1, 2}, // mesh0=3
			want:    true,
		},
		{
			name:    "task < 4, no mesh has enough cards",
			taskNPU: 3,
			nodeTop: []int{0, 4, 8}, // 每个 mesh 只有1
			want:    false,
		},
		{
			name:    "task < 4, exactly enough in one mesh",
			taskNPU: 3,
			nodeTop: []int{0, 1, 2, 5, 6, 7, 9, 10, 11}, // mesh0=3, mesh1=3, mesh2=3
			want:    true,
		},
	}
}

func getFullMeshTestCases() []judgeNodeAndTaskNpuIn4PmeshTestCase {
	return []judgeNodeAndTaskNpuIn4PmeshTestCase{
		{
			name:    "task = 4, one full mesh",
			taskNPU: 4,
			nodeTop: []int{0, 1, 2, 3}, // mesh0 full
			want:    true,
		},
		{
			name:    "task = 4, no full mesh",
			taskNPU: 4,
			nodeTop: []int{0, 1, 2}, // mesh0=3
			want:    false,
		},
		{
			name:    "task = 8, two full meshes",
			taskNPU: 8,
			nodeTop: []int{0, 1, 2, 3, 4, 5, 6, 7}, // mesh0=4, mesh1=4
			want:    true,
		},
		{
			name:    "task = 8, only one full mesh",
			taskNPU: 8,
			nodeTop: []int{0, 1, 2, 3, 8, 9, 13, 15}, // mesh0=4, mesh2=2
			want:    false,
		},
		{
			name:    "task = 16, all full",
			taskNPU: 16,
			nodeTop: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			want:    true,
		},
		{
			name:    "task = 16, missing some cards",
			taskNPU: 16,
			nodeTop: []int{0, 1, 2, 3, 4, 5, 6, 7}, // only 2 meshes full
			want:    false,
		},
	}
}

func TestJudgeNodeAndTaskNpuIn4Pmesh(t *testing.T) {
	tests := getNotFullMeshTestCases()
	tests = append(tests, getFullMeshTestCases()...)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := judgeNodeAndTaskNpuIn4Pmesh(tt.taskNPU, tt.nodeTop)
			if got != tt.want {
				t.Errorf("JudgeNodeAndTaskNpuIn4Pmesh(%d, %v) = %v, want %v",
					tt.taskNPU, tt.nodeTop, got, tt.want)
			}
		})
	}
}

func TestIs4PmeshAffinity(t *testing.T) {
	tests := []struct {
		name     string
		jobType  string
		taskNPUs int
		want     bool
	}{
		{
			name:     "single card should be true",
			jobType:  SchedulePolicy4Px8,
			taskNPUs: 1,
			want:     true,
		},
		{
			name:     "valid label, task < CardsNumPerMesh",
			jobType:  SchedulePolicy4Px16,
			taskNPUs: 2,
			want:     true,
		},
		{
			name:     "valid label, task divisible by CardsNumPerMesh",
			jobType:  SchedulePolicy4Px8,
			taskNPUs: 8,
			want:     true,
		},
		{
			name:     "valid label, task not < mesh and not divisible",
			jobType:  SchedulePolicy4Px8,
			taskNPUs: 5,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := chip4nodex{}
			tp.SetPluginName(tt.jobType)

			got := is4PmeshAffinity(tt.taskNPUs)
			if got != tt.want {
				t.Errorf("is4PmeshAffinity(%d) with jobType=%s = %v, want %v",
					tt.taskNPUs, tt.jobType, got, tt.want)
			}
		})
	}
}

func TestGetNodeBestScoreIn4Pmesh(t *testing.T) {
	tp := &chip4nodex{
		Base910A5: ascend910a5.Base910A5{
			NPUHandler: base.NPUHandler{MaxNodeNPUNum: 16},
		},
		affScoreList: make([][]int, 16),
	}

	for i := 0; i < 16; i++ {
		tp.affScoreList[i] = make([]int, 16)
		for j := 0; j < 16; j++ {
			if i > j {
				tp.affScoreList[i][j] = 16
			} else {
				tp.affScoreList[i][j] = j - i
			}
		}
	}

	tests := []struct {
		name          string
		taskNPU       int
		npuTop        []int
		expected      int
		wantErr       bool
		maxNodeNPUNum int
	}{
		// 8-card 4Pmesh scheduling, prioritizing those with fewer remaining mesh counts, then those with fewer remaining nodes.
		{"task=2, mesh={4,0,0,0}", 2, []int{0, 1, 2, 3}, 2, false, 8},
		{"task=2, mesh={2,2,0,0}", 2, []int{0, 1, 4, 5}, 10, false, 8},
		{"task=2, mesh={3,0,0,0}", 2, []int{0, 1, 2}, 1, false, 8},
		// 16-card 4Pmesh scheduling, prioritizing those with fewer remaining mesh counts, then those with fewer remaining nodes.
		{"task=8, mesh={4,4,4,4}", 8, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, 40, false, 16},
		{"task=8, mesh={4,4,4,2}", 8, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}, 22, false, 16},
		{"task=8, mesh={4,4,4,0}", 8, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, 20, false, 16},
		{"task=8, mesh={4,0,4,0}", 8, []int{0, 1, 2, 3, 8, 9, 10, 11}, 0, false, 16},
		// Boundary scenario, the required number of cards exceeds the limit
		{"task=0, mesh={4,0,0,0}", 0, []int{0, 1, 2, 3}, 0, true, 16},
		{"task=17, mesh={4,0,0,0}", 0, []int{0, 1, 2, 3}, 0, true, 16},
		// In boundary scenarios, the obtained card topology information exceeds the range
		{"task=8, mesh={4,4,4,4,1}", 8, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, 0, true, 16},
		{"task=2, mesh={0,0,0,0}", 2, []int{}, 2, true, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp.MaxNodeNPUNum = tt.maxNodeNPUNum
			got, err := tp.getNodeBestScoreIn4Pmesh(tt.taskNPU, tt.npuTop)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected err=%v, got %v", tt.wantErr, err)
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("expected score=%d, got %d", tt.expected, got)
			}
		})
	}
}

func TestScoreNodeFor4Pmesh(t *testing.T) {
	tp := &chip4nodex{
		Base910A5: ascend910a5.Base910A5{
			NPUHandler: base.NPUHandler{MaxNodeNPUNum: 16},
		},
		affScoreList: make([][]int, 16),
	}

	for i := 0; i < 16; i++ {
		tp.affScoreList[i] = make([]int, 16)
		for j := 0; j < 16; j++ {
			if i > j {
				tp.affScoreList[i][j] = 16
			} else {
				tp.affScoreList[i][j] = j - i
			}
		}
	}

	tests := []struct {
		name          string
		taskNPU       int
		npuTop        []int
		expected      float64
		maxNodeNPUNum int
	}{
		// 8-card 4Pmesh scheduling, prioritizing those with fewer remaining mesh counts, then those with fewer remaining nodes.
		{"task=2, mesh={4,0,0,0}", 2, []int{0, 1, 2, 3}, 16 * 14, 8},
		{"task=2, mesh={2,2,0,0}", 2, []int{0, 1, 4, 5}, 16 * 6, 8},
		{"task=2, mesh={3,0,0,0}", 2, []int{0, 1, 2}, 16 * 15, 8},
		// 16-card 4Pmesh scheduling, prioritizing those with fewer remaining mesh counts, then those with fewer remaining nodes.
		{"task=8, mesh={4,4,4,4}", 8, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, 64 * 24, 16},
		{"task=8, mesh={4,4,4,2}", 8, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}, 64 * 42, 16},
		{"task=8, mesh={4,4,4,0}", 8, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, 64 * 44, 16},
		{"task=8, mesh={4,0,4,0}", 8, []int{0, 1, 2, 3, 8, 9, 10, 11}, 64 * 64, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp.MaxNodeNPUNum = tt.maxNodeNPUNum
			got := tp.scoreNodeFor4Pmesh(tt.taskNPU, tt.npuTop)
			if got != tt.expected {
				t.Errorf("expected score=%f, got %f", tt.expected, got)
			}
		})
	}

}

func TestSelectNPUNotFullMesh(t *testing.T) {
	tests := []struct {
		name       string
		taskNPUNum int
		nodeTop    []int
		want       []int
	}{
		{
			name:       "task=2, mesh={3,2,3,3}",
			taskNPUNum: 2,
			nodeTop:    []int{0, 1, 2, 5, 6, 8, 9, 10, 12, 13, 14},
			want:       []int{5, 6},
		},
		{
			name:       "task=2, mesh={4,1,3,4}",
			taskNPUNum: 2,
			nodeTop:    []int{0, 1, 2, 3, 5, 8, 9, 10, 12, 13, 14, 15},
			want:       []int{8, 9},
		},
		{
			name:       "task=3, mesh={4,2,3,4}",
			taskNPUNum: 3,
			nodeTop:    []int{0, 1, 2, 3, 4, 5, 8, 9, 10, 12, 13, 14, 15},
			want:       []int{8, 9, 10},
		},
		{
			name:       "task=-1, mesh={4,2,3,4}",
			taskNPUNum: -1,
			nodeTop:    []int{},
			want:       nil,
		},
		{
			name:       "task=4, mesh={4,2,3,4}",
			taskNPUNum: 4,
			nodeTop:    []int{0, 5, 9},
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectNPUinSingleMesh(tt.taskNPUNum, tt.nodeTop)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("selectNPUinSingleMesh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectNPUMultiMesh(t *testing.T) {
	tests := []struct {
		name       string
		taskNPUNum int
		nodeTop    []int
		want       []int
	}{
		{
			name:       "task=4, mesh={3,2,3,4}",
			taskNPUNum: 4,
			nodeTop:    []int{0, 1, 2, 5, 6, 8, 9, 10, 12, 13, 14, 15},
			want:       []int{12, 13, 14, 15},
		},
		{
			name:       "task=4, mesh={3,2,4,4}",
			taskNPUNum: 4,
			nodeTop:    []int{0, 1, 2, 5, 6, 8, 9, 10, 11, 12, 13, 14, 15},
			want:       []int{8, 9, 10, 11},
		},
		{
			name:       "task=8, mesh={4,1,3,4}",
			taskNPUNum: 8,
			nodeTop:    []int{0, 1, 2, 3, 5, 8, 9, 10, 12, 13, 14, 15},
			want:       []int{0, 1, 2, 3, 12, 13, 14, 15},
		},
		{
			name:       "task=12, mesh={4,2,4,4}",
			taskNPUNum: 12,
			nodeTop:    []int{0, 1, 2, 3, 4, 5, 8, 9, 10, 11, 12, 13, 14, 15},
			want:       []int{0, 1, 2, 3, 8, 9, 10, 11, 12, 13, 14, 15},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectNPUMultiMesh(tt.taskNPUNum, tt.nodeTop)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("selectNPUinSingleMesh() = %v, want %v", got, tt.want)
			}
		})
	}
}
