/*
Copyright(C)2020-2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package util is using for the total variable.
*/
package util

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
)

type testIsVJobTest struct {
	name string
	nJob *NPUJob
	want bool
}

func buildTestIsVJobTest() []testIsVJobTest {
	tests := []testIsVJobTest{
		{
			name: "01-IsVJob nJob.ReqNPUName nil test.",
			nJob: &NPUJob{},
			want: false,
		},
		{
			name: "02-IsVjob nJob.ReqNPUName > 2 test.",
			nJob: &NPUJob{
				ReqNPUName: AscendNPUCore,
			},
			want: true,
		},
	}
	return tests
}

func TestIsVJob(t *testing.T) {
	tests := buildTestIsVJobTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.nJob.IsVJob(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNPUJobIsNPUJob(t *testing.T) {
	type fields struct {
		ReqNPUName string
		ReqNPUNum  int
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "01-IsNPUJob npu job with AscendNPUCore resource.",
			fields: fields{ReqNPUName: AscendNPUCore},
			want:   true,
		},
		{
			name:   "02-IsNPUJob not npu job with empty name.",
			fields: fields{ReqNPUName: ""},
			want:   false,
		},
		{
			name:   "03-IsNPUJob annotation-based NPU job with ReqNPUNum=0.",
			fields: fields{ReqNPUName: NPU910CardName, ReqNPUNum: 0},
			want:   true,
		},
		{
			name:   "04-IsNPUJob normal NPU job with Ascend310P.",
			fields: fields{ReqNPUName: NPU310PCardName, ReqNPUNum: 8},
			want:   true,
		},
		{
			name:   "05-IsNPUJob nil NPUJob.",
			fields: fields{ReqNPUName: ""},
			want:   false,
		},
		{
			name:   "06-IsNPUJob dynamic VNPU job (npu-core).",
			fields: fields{ReqNPUName: AscendNPUCore, ReqNPUNum: 4},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nJob *NPUJob
			if tt.name != "05-IsNPUJob nil NPUJob." {
				nJob = &NPUJob{
					ReqNPUName: tt.fields.ReqNPUName,
				}
			}
			if got := nJob.IsNPUJob(); got != tt.want {
				t.Errorf("IsNPUJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNPUJobGetSchedulingTaskNum(t *testing.T) {
	type fields struct {
		Tasks      map[api.TaskID]NPUTask
		ReqNPUName string
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "01-GetSchedulingTaskNum not npu job",
			fields: fields{
				ReqNPUName: "",
			},
			want: 0,
		},
		{
			name: "02-GetSchedulingTaskNum npu job",
			fields: fields{
				ReqNPUName: AscendNPUCore,
				Tasks:      map[api.TaskID]NPUTask{"task01": {ReqNPUName: Ascend910bName}},
			},
			want: 1,
		},
		{
			name: "03-GetSchedulingTaskNum npu job",
			fields: fields{
				ReqNPUName: AscendNPUCore,
				Tasks: map[api.TaskID]NPUTask{
					"task01": {ReqNPUName: Ascend910bName, NodeName: "node01"},
					"task00": {ReqNPUName: NPU310CardName}},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nJob := &NPUJob{
				Tasks:      tt.fields.Tasks,
				ReqNPUName: tt.fields.ReqNPUName,
			}
			if got := nJob.GetSchedulingTaskNum(); got != tt.want {
				t.Errorf("GetSchedulingTaskNum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReferenceNameOfJob(t *testing.T) {
	tests := []struct {
		name string
		job  *api.JobInfo
		want string
	}{
		{
			name: "01-ReferenceNameOfJob nil job",
			job:  nil,
			want: "",
		},
		{
			name: "02-ReferenceNameOfJob nil podgroup",
			job:  &api.JobInfo{},
			want: "",
		},
		{
			name: "03-ReferenceNameOfJob nil podgroup ownerreference",
			job:  &api.JobInfo{PodGroup: &api.PodGroup{}},
			want: "",
		},
		{
			name: "04-ReferenceNameOfJob podgroup has ownerreference",
			job: &api.JobInfo{PodGroup: &api.PodGroup{
				PodGroup: scheduling.PodGroup{ObjectMeta: v1.ObjectMeta{
					OwnerReferences: []v1.OwnerReference{{Name: "test"}}}},
			}},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReferenceNameOfJob(tt.job); got != tt.want {
				t.Errorf("ReferenceNameOfJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUuidOfJob(t *testing.T) {
	tests := []struct {
		name string
		job  *api.JobInfo
		want types.UID
	}{
		{
			name: "01-UuidOfJob nil job",
			job:  nil,
			want: "",
		},
		{
			name: "02-UuidOfJob nil podgroup",
			job:  &api.JobInfo{},
			want: "",
		},
		{
			name: "03-UuidOfJob nil podgroup ownerreference",
			job:  &api.JobInfo{PodGroup: &api.PodGroup{}},
			want: "",
		},
		{
			name: "04-UuidOfJob podgroup has ownerreference",
			job: &api.JobInfo{PodGroup: &api.PodGroup{
				PodGroup: scheduling.PodGroup{ObjectMeta: v1.ObjectMeta{
					OwnerReferences: []v1.OwnerReference{{Name: "test", UID: "test-uid"}}}},
			}},
			want: "test-uid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UuidOfJob(tt.job); got != tt.want {
				t.Errorf("UuidOfJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPluginNameByReq(t *testing.T) {
	t.Run("01 will return empty string when SchedulerJobAttr is empty", func(t *testing.T) {
		sJob := SchedulerJobAttr{NPUJob: &NPUJob{}}
		if got := sJob.GetPluginNameByReq(); got != "" {
			t.Errorf("GetPluginNameByReq() = %v, want %v", got, "")
		}
	})
	t.Run("02 will return huawei.com/Ascend910 when SchedulerJobAttr require it", func(t *testing.T) {
		sJob := SchedulerJobAttr{NPUJob: &NPUJob{ReqNPUName: AscendNPUCore}}
		sJob.Label = map[string]string{JobKindKey: JobKind910Value}
		if got := sJob.GetPluginNameByReq(); got != NPU910CardName {
			t.Errorf("GetPluginNameByReq() = %v, want %v", got, NPU910CardName)
		}
	})
	t.Run("03 will return huawei.com/Ascend910 when SchedulerJobAttr require it", func(t *testing.T) {
		sJob := SchedulerJobAttr{NPUJob: &NPUJob{ReqNPUName: AscendNPUCore}}
		sJob.Label = map[string]string{JobKindKey: JobKindDefaultValue}
		if got := sJob.GetPluginNameByReq(); got != NPU910CardName {
			t.Errorf("GetPluginNameByReq() = %v, want %v", got, NPU910CardName)
		}
	})
	t.Run("04 will return empty string when dynamic job and label is nil", func(t *testing.T) {
		sJob := SchedulerJobAttr{NPUJob: &NPUJob{ReqNPUName: AscendNPUCore}}
		if got := sJob.GetPluginNameByReq(); got != "" {
			t.Errorf("GetPluginNameByReq() = %v, want %v", got, "")
		}
	})
}

func TestIsSuperPodJob(t *testing.T) {
	annotationSpBlock := map[string]string{SuperPodAnnoKey: "16"}
	annotationNotSpPolicy := map[string]string{SchedulePolicyAnnoKey: SchedulePolicyA3x16}
	selectorSp := map[string]string{AcceleratorType: Module910A3x16SuperPodAcceleratorType}
	annotationSchedulePolicySp := map[string]string{SchedulePolicyAnnoKey: Chip2Node16Sp}
	annotationSchedulePolicyAtlas9000Sp := map[string]string{SchedulePolicyAnnoKey: Chip2Node8Sp}
	t.Run("01-isSuperPodJob true, when sp-block exist",
		func(t *testing.T) {
			attr := SchedulerJobAttr{ComJob: ComJob{Annotation: annotationSpBlock}}
			if !attr.IsSuperPodJob() {
				t.Errorf("isSuperPodJob() err, want true while return false")
			}
		})
	t.Run("02-isSuperPodJob true, when sp-block is nil, and accelerator-type exist ",
		func(t *testing.T) {
			attr := SchedulerJobAttr{ComJob: ComJob{Selector: selectorSp}}
			if !attr.IsSuperPodJob() {
				t.Errorf("isSuperPodJob() err, want true while return false")
			}
		})
	t.Run("03-isSuperPodJob false, when sp-block is exist, but schedule policy exist ",
		func(t *testing.T) {
			attr := SchedulerJobAttr{ComJob: ComJob{Annotation: annotationNotSpPolicy, Selector: selectorSp}}
			if attr.IsSuperPodJob() {
				t.Errorf("isSuperPodJob() err, want false while return true")
			}
		})
	t.Run("04-isSuperPodJob false, when sp-block/accelerator-type/schedule-policy are nil ",
		func(t *testing.T) {
			attr := SchedulerJobAttr{}
			if attr.IsSuperPodJob() {
				t.Errorf("isSuperPodJob() err, want false while return true")
			}
		})
	t.Run("05-isSuperPodJob true, when schedule-policy are chip2-node16-sp",
		func(t *testing.T) {
			attr := SchedulerJobAttr{ComJob: ComJob{Annotation: annotationSchedulePolicySp}}
			if !attr.IsSuperPodJob() {
				t.Errorf("isSuperPodJob() err, want true while return false")
			}
		})
	t.Run("06-isSuperPodJob true, when schedule-policy are chip2-node8-sp",
		func(t *testing.T) {
			attr := SchedulerJobAttr{ComJob: ComJob{Annotation: annotationSchedulePolicyAtlas9000Sp}}
			if !attr.IsSuperPodJob() {
				t.Errorf("isSuperPodJob() err, want true while return false")
			}
		})
}


type countBackupTasksCase struct {
	name  string
	tasks map[api.TaskID]NPUTask
	want  int
}

func buildCountBackupTasksCases() []countBackupTasksCase {
	return []countBackupTasksCase{
		{name: "01-nil job tasks", tasks: nil, want: 0},
		{name: "02-empty tasks", tasks: map[api.TaskID]NPUTask{}, want: 0},
		{
			name: "03-no backup pods",
			tasks: map[api.TaskID]NPUTask{
				"t0": {Annotation: map[string]string{}},
				"t1": {Annotation: map[string]string{"other": "val"}},
			},
			want: 0,
		},
		{
			name: "04-one backup pod",
			tasks: map[api.TaskID]NPUTask{
				"t0": {Annotation: map[string]string{PodTypeKey: PodTypeBackup}},
				"t1": {Annotation: map[string]string{}},
			},
			want: 1,
		},
		{
			name: "05-mixed original and backup pods",
			tasks: map[api.TaskID]NPUTask{
				"t0": {Annotation: map[string]string{}},
				"t1": {Annotation: map[string]string{PodTypeKey: PodTypeBackup}},
				"t2": {Annotation: map[string]string{}},
				"t3": {Annotation: map[string]string{PodTypeKey: PodTypeBackup}},
			},
			want: 2,
		},
	}
}

func TestNPUJobCountBackupTasks(t *testing.T) {
	for _, tt := range buildCountBackupTasksCases() {
		t.Run(tt.name, func(t *testing.T) {
			nJob := &NPUJob{Tasks: tt.tasks}
			if got := nJob.CountBackupTasks(); got != tt.want {
				t.Errorf("CountBackupTasks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskValidResult(t *testing.T) {
	helper := NewTaskValidateHelper()
	expected := "valid expected message"

	// Case 1: No invalid resource requests
	result := helper.TaskValidResult(expected)
	if result != nil {
		t.Errorf("Expected nil result when there are no invalid resource requests")
	}

	// Case 2: Invalid resource requests exist
	taskID := api.TaskID("test-task")
	resourceRequest := NPUIndex10
	helper.AddInvalidResourceRequest(taskID, resourceRequest)

	result = helper.TaskValidResult(expected)
	if result == nil {
		t.Error("Expected non-nil result when there are invalid resource requests")
	}
	if result.Pass {
		t.Error("Expected Pass to be false when there are invalid resource requests")
	}
	if result.Reason != InvalidResourceRequestReason {
		t.Errorf("Expected reason %s, got %s", InvalidResourceRequestReason, result.Reason)
	}
	expectedMessage := fmt.Sprintf("expected: %s, actual: task<%s> req npu: %d, ", expected, taskID, resourceRequest)
	if result.Message != expectedMessage {
		t.Errorf("Expected message %s, got %s", expectedMessage, result.Message)
	}
}
