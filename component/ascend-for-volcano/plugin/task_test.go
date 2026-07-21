/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package plugin is using for HuaWei Ascend pin affinity schedule frame.
*/
package plugin

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type npuAllocateFuncArgs struct {
	task *api.TaskInfo
}

type npuAllocateFuncTest struct {
	name   string
	fields fields
	args   npuAllocateFuncArgs
	want   string
}

func buildNPUAllocateFuncTest() []npuAllocateFuncTest {
	task := test.FakeNormalTestTasks(1)[0]
	name, num := getVCTaskReqNPUTypeFromTaskInfo(task)
	tmpJobReadyTag := true
	npuTask := util.NPUTask{
		Name: task.Name, NameSpace: task.Namespace, ReqNPUName: name,
		ReqNPUNum: num,
		Label:     getTaskLabels(task), VTask: &util.VTask{}}
	tests := []npuAllocateFuncTest{
		{
			name:   "01-NPUAllocateFunc task nil test",
			fields: fields{},
			args:   npuAllocateFuncArgs{task: nil},
			want:   "",
		},
		{
			name: "02-NPUAllocateFunc no job test.",
			fields: fields{NPUPlugins: make(sets.String),
				ScheduleEnv: ScheduleEnv{
					ClusterCache: NewClusterCache(),
					FrameAttr:    VolcanoFrame{}}},
			args: npuAllocateFuncArgs{task: task},
			want: "",
		},
		{
			name: "03-NPUAllocateFunc no node test",
			fields: fields{NPUPlugins: make(sets.String),
				ScheduleEnv: ScheduleEnv{
					ClusterCache: ClusterCache{
						Jobs: map[api.JobID]SchedulerJob{task.Job: {
							JobReadyTag: &tmpJobReadyTag,
							SchedulerJobAttr: util.SchedulerJobAttr{
								NPUJob: &util.NPUJob{
									Tasks: map[api.TaskID]util.NPUTask{task.UID: npuTask},
								},
							}}},
						Nodes: map[string]NPUNode{}},
					FrameAttr: VolcanoFrame{}}},
			args: npuAllocateFuncArgs{task: task},
			want: "",
		},
		{
			name: "04-NPUAllocateFunc UseAnnotation failed test.",
			fields: fields{NPUPlugins: make(sets.String),
				ScheduleEnv: newDefaultsHandlerByFakeSsn().ScheduleEnv},
			args: npuAllocateFuncArgs{task: task},
			want: "",
		},
	}
	return tests
}

func TestNPUAllocateFunc(t *testing.T) {
	tests := buildNPUAllocateFuncTest()
	temp := func(task *api.TaskInfo) string {
		if task == nil {
			return ""
		}
		value, _ := task.Pod.Annotations[test.NPU910CardName]
		return value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			sHandle.NPUAllocateFunc(tt.args.task)
			value := temp(tt.args.task)
			if value != tt.want {
				t.Errorf("NPUAllocateFunc() got = %v, want %v", value, tt.want)
			}
		})
	}
}

type npuDeallocateFuncArgs struct {
	task *api.TaskInfo
}

type npuDeallocateFuncTest struct {
	name   string
	fields fields
	args   npuDeallocateFuncArgs
	want   string
}

func makeNPUDeallocateFuncTest01(_ *api.TaskInfo) npuDeallocateFuncTest {
	return npuDeallocateFuncTest{
		name:   "01-NPUDeallocateFunc task nil test",
		fields: fields{}, args: npuDeallocateFuncArgs{task: nil}, want: "",
	}
}

func makeNPUDeallocateFuncTest02(vTask *api.TaskInfo) npuDeallocateFuncTest {
	return npuDeallocateFuncTest{
		name: "02-NPUAllocateFunc no job test.",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{ClusterCache: NewClusterCache()}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "Ascend910-4",
	}
}

func makeNPUDeallocateFuncTest03(vTask *api.TaskInfo) npuDeallocateFuncTest {
	return npuDeallocateFuncTest{
		name: "03-NPUAllocateFunc no node test",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs:  map[api.JobID]SchedulerJob{vTask.Job: {}},
					Nodes: map[string]NPUNode{}}}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "Ascend910-4",
	}
}

func makeNPUDeallocateFuncTest04(vTask *api.TaskInfo) npuDeallocateFuncTest {
	return npuDeallocateFuncTest{
		name: "04-NPUAllocateFunc UseAnnotation failed test.",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs: map[api.JobID]SchedulerJob{
						vTask.Job: {SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: nil}}}},
					Nodes: map[string]NPUNode{vTask.NodeName: {}}}}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "Ascend910-4",
	}
}

func makeNPUDeallocateFuncTest05(vTask *api.TaskInfo) npuDeallocateFuncTest {
	return npuDeallocateFuncTest{
		name: "05-NPUAllocateFunc pod no req test.",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs: map[api.JobID]SchedulerJob{
						vTask.Job: {
							SchedulerJobAttr: util.SchedulerJobAttr{
								NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{vTask.UID: {ReqNPUName: "haha"}}}},
						},
					},
					Nodes: map[string]NPUNode{vTask.NodeName: {}}}}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "Ascend910-4",
	}
}

func makeNPUDeallocateFuncTest06(vTask *api.TaskInfo) npuDeallocateFuncTest {
	tmpSchedulerJobAttr := util.SchedulerJobAttr{
		NPUJob: &util.NPUJob{
			Tasks: map[api.TaskID]util.NPUTask{
				vTask.UID: {ReqNPUName: test.NPU910CardName, ReqNPUNum: util.NPUIndex2}},
		},
	}

	return npuDeallocateFuncTest{
		name: "06-NPUAllocateFunc pod req num not meet test.",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs: map[api.JobID]SchedulerJob{
						vTask.Job: {
							SchedulerJobAttr: tmpSchedulerJobAttr,
						},
					},
					Nodes: map[string]NPUNode{vTask.NodeName: {}}}}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "Ascend910-4",
	}
}

func makeNPUDeallocateFuncTest07(vTask *api.TaskInfo) npuDeallocateFuncTest {
	tmpSchedulerJobAttr := util.SchedulerJobAttr{
		NPUJob: &util.NPUJob{
			Tasks: map[api.TaskID]util.NPUTask{
				vTask.UID: {ReqNPUName: test.NPU910CardName, ReqNPUNum: 1}},
		},
	}
	tmpNPUNode := NPUNode{
		CommonNode: CommonNode{Annotation: nil},
	}
	return npuDeallocateFuncTest{
		name: "07-NPUAllocateFunc node no annotation value test.",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs: map[api.JobID]SchedulerJob{
						vTask.Job: {SchedulerJobAttr: tmpSchedulerJobAttr},
					},
					Nodes: map[string]NPUNode{vTask.NodeName: tmpNPUNode}}}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "Ascend910-4",
	}
}

func makeNPUDeallocateFuncTest08(vTask *api.TaskInfo) npuDeallocateFuncTest {
	tmpSchedulerJobAttr := util.SchedulerJobAttr{
		NPUJob: &util.NPUJob{
			Tasks: map[api.TaskID]util.NPUTask{
				vTask.UID: {ReqNPUName: test.NPU910CardName, ReqNPUNum: 1,
					VTask: &util.VTask{}}},
		},
	}
	tmpNPUNode := NPUNode{
		CommonNode: CommonNode{
			Annotation: map[string]string{test.NPU910CardName: ""},
		},
	}
	return npuDeallocateFuncTest{
		name: "08-NPUAllocateFunc node has empty annotation value test.",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs: map[api.JobID]SchedulerJob{vTask.Job: {SchedulerJobAttr: tmpSchedulerJobAttr,
						policyHandler: New(testPluginName)}},
					Nodes: map[string]NPUNode{vTask.NodeName: tmpNPUNode}}}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "",
	}
}

func makeNPUDeallocateFuncTest09(vTask *api.TaskInfo) npuDeallocateFuncTest {
	tmpSchedulerJobAttr := util.SchedulerJobAttr{
		NPUJob: &util.NPUJob{
			Tasks: map[api.TaskID]util.NPUTask{
				vTask.UID: {ReqNPUName: test.NPU910CardName, ReqNPUNum: 1}},
		},
	}
	tmpNPUNode := NPUNode{
		CommonNode: CommonNode{
			Annotation: map[string]string{test.NPU910CardName: "Ascend910-3"},
		},
	}
	return npuDeallocateFuncTest{
		name: "09-NPUAllocateFunc ok test.",
		fields: fields{NPUPlugins: make(sets.String),
			ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs: map[api.JobID]SchedulerJob{vTask.Job: {SchedulerJobAttr: tmpSchedulerJobAttr,
						policyHandler: New(testPluginName)}},
					Nodes: map[string]NPUNode{vTask.NodeName: tmpNPUNode}}}},
		args: npuDeallocateFuncArgs{task: vTask}, want: "",
	}
}

func buildNPUDeallocateFuncTest() []npuDeallocateFuncTest {
	vTask := test.BuildTestTaskWithAnnotation(test.NPU910CardName, "1", "Ascend910-4")
	tests := []npuDeallocateFuncTest{
		makeNPUDeallocateFuncTest01(vTask),
		makeNPUDeallocateFuncTest02(vTask),
		makeNPUDeallocateFuncTest03(vTask),
		makeNPUDeallocateFuncTest04(vTask),
		makeNPUDeallocateFuncTest05(vTask),
		makeNPUDeallocateFuncTest06(vTask),
		makeNPUDeallocateFuncTest07(vTask),
		makeNPUDeallocateFuncTest08(vTask),
		makeNPUDeallocateFuncTest09(vTask),
	}
	return tests
}

// TestGetAllocatedChipIDsFromPod covers nil pod, nil annotations and
// multiple card-name annotations (910 / 310P / npu).
func TestGetAllocatedChipIDsFromPod(t *testing.T) {
	node := &NPUNode{}
	tests := []struct {
		name      string
		podName   string
		npuName   string
		annoVal   string
		wantLen   int
		wantFirst int
	}{
		{"01-nil pod handled by caller", "", "", "", 0, 0},
		{"02-910 single", "p1", util.NPU910CardName, "Ascend910-3", 1, 3},
		{"03-910 multi", "p2", util.NPU910CardName, "Ascend910-0,Ascend910-3", 2, 0},
		{"04-310P", "p3", util.NPU310PCardName, "Ascend310P-1,Ascend310P-2", 2, 1},
		{"05-npu", "p4", util.NPUCardName, "npu-1,npu-2", 2, 1},
		{"06-empty annotation", "p5", util.NPU910CardName, "", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.podName == "" {
				if got := getAllocatedChipIDsFromPod(nil, node); len(got) != 0 {
					t.Errorf("got=%v want empty", got)
				}
				return
			}
			task := test.BuildTestTaskWithAnnotation(tt.npuName, "1", tt.annoVal)
			got := getAllocatedChipIDsFromPod(task.Pod, node)
			if len(got) != tt.wantLen {
				t.Errorf("got=%v wantLen=%d", got, tt.wantLen)
				return
			}
			if tt.wantLen > 0 && got[0] != tt.wantFirst {
				t.Errorf("got[0]=%d want=%d", got[0], tt.wantFirst)
			}
		})
	}
}

// TestCalcCardFreeCount covers invalid args and whole-card scenarios:
// Chips map empty, preemptee chip counted, unhealthy excluded,
// non-preemptee occupied excluded, multi-card bucketing,
// and non-contiguous available chip IDs (only chips in availableChipIDs counted).
func TestCalcCardFreeCount(t *testing.T) {
	pe := test.BuildTestTaskWithAnnotation(util.NPU910CardName, "1", "Ascend910-0")
	pe.Pod.UID = "pe-uid" // override default UID "-" to distinguish from other pods
	other := test.BuildTestTaskWithAnnotation(util.NPU910CardName, "1", "Ascend910-2")
	other.Pod.UID = "other-uid"
	tests := []struct {
		name     string
		node     *NPUNode
		maxNum   int
		availIDs []int
		wantCard int
		wantCnt  int
		wantFull map[int]int // full map expectation, optional; if nil only wantCard/wantCnt checked
	}{
		{"01-nil node", nil, 0, nil, 0, 0, nil},
		{"02-preemptee+unhealthy+idle", &NPUNode{
			CommonNode: CommonNode{Name: "n1"},
			VNode:      VNode{UnhealthyChipIds: map[int]struct{}{1: {}}}}, 4, []int{0, 1, 2, 3}, 0, 3, nil},
		{"03-with non-preemptee occupied", &NPUNode{
			CommonNode: CommonNode{Name: "n2", Tasks: map[api.TaskID]*api.TaskInfo{api.TaskID("other-uid"): other}},
			VNode:      VNode{}}, 4, []int{0, 1, 2, 3}, 0, 3, nil},
		{"04-multi-card bucketing", &NPUNode{
			CommonNode: CommonNode{Name: "n3"},
			VNode:      VNode{UnhealthyChipIds: map[int]struct{}{5: {}}}}, 4, []int{0, 1, 2, 3, 4, 5, 6, 7}, 1, 3, nil},
		// Simulates real-world scenario: 8 cards in 2 meshes, but only chips 1,2,3,7 are
		// reported as available (0,4,5,6 are unschedulable and absent from node annotation).
		// preemptee pe holds chip 0, which is NOT in availableChipIDs, so it is ignored.
		// Result: mesh 0 has 3 free chips (1,2,3), mesh 1 has 1 free chip (7).
		{"05-non-contiguous available IDs", &NPUNode{
			CommonNode: CommonNode{Name: "n4"},
			VNode:      VNode{}}, 4, []int{1, 2, 3, 7}, 0, 3, map[int]int{0: 3, 1: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcCardFreeCount(tt.node, []*api.TaskInfo{pe}, tt.maxNum, tt.availIDs)
			if tt.node == nil {
				if got != nil {
					t.Errorf("nil node got=%v want nil", got)
				}
				return
			}
			if tt.wantFull != nil {
				if len(got) != len(tt.wantFull) {
					t.Errorf("got=%v want=%v", got, tt.wantFull)
					return
				}
				for k, v := range tt.wantFull {
					if got[k] != v {
						t.Errorf("card[%d]=%d want=%d, full=%v", k, got[k], v, got)
					}
				}
				return
			}
			if got[tt.wantCard] != tt.wantCnt {
				t.Errorf("card[%d]=%d want=%d, full=%v", tt.wantCard, got[tt.wantCard], tt.wantCnt, got)
			}
		})
	}
}

func TestNPUDeallocateFunc(t *testing.T) {
	tests := buildNPUDeallocateFuncTest()
	temp := func(task *api.TaskInfo) string {
		if task == nil {
			return ""
		}
		value, _ := task.Pod.Annotations[test.NPU910CardName]
		return value
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := &ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			sHandle.NPUDeallocateFunc(tt.args.task)
			value := temp(tt.args.task)
			if value != tt.want {
				t.Errorf("NPUDeallocateFunc() got = %v, want %v", value, tt.want)
			}
		})
	}
}
