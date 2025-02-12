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
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
)

type testIsSelectorMeetJobArgs struct {
	jobSelectors map[string]string
	conf         map[string]string
}

type testIsSelectorMeetJobTest struct {
	name string
	args testIsSelectorMeetJobArgs
	want bool
}

func buildTestIsSelectorMeetJobTest() []testIsSelectorMeetJobTest {
	tests := []testIsSelectorMeetJobTest{
		{
			name: "01-IsSelectorMeetJob nil jobSelector test.",
			args: testIsSelectorMeetJobArgs{jobSelectors: nil, conf: nil},
			want: true,
		},
		{
			name: "02-IsSelectorMeetJob conf no job selector test.",
			args: testIsSelectorMeetJobArgs{jobSelectors: map[string]string{"haha": "test"}, conf: nil},
			want: false,
		},
		{
			name: "03-IsSelectorMeetJob jobSelector no have conf test.",
			args: testIsSelectorMeetJobArgs{jobSelectors: map[string]string{"haha": "test"},
				conf: map[string]string{"haha": "what"}},
			want: false,
		},
	}
	return tests
}

// TestIsSelectorMeetJob test IsSelectorMeetJob.
func TestIsSelectorMeetJob(t *testing.T) {
	tests := buildTestIsSelectorMeetJobTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSelectorMeetJob(tt.args.jobSelectors, tt.args.conf); got != tt.want {
				t.Errorf("IsSelectorMeetJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

type SetVJobTypeTest struct {
	name string
	nJob *NPUJob
	want int
}

func buildSetVJobTypeTestCase() []SetVJobTypeTest {
	return []SetVJobTypeTest{
		{
			name: "01-SetJobType JobTypeUnknown status",
			nJob: &NPUJob{
				Tasks: map[api.TaskID]NPUTask{},
				VJob:  &VJob{},
			},
			want: JobTypeUnknown,
		},
		{
			name: "02-SetJobType JobTypeWhole status",
			nJob: &NPUJob{
				Tasks: map[api.TaskID]NPUTask{"task01": {ReqNPUNum: 1, VTask: &VTask{Type: JobTypeWhole}}},
				VJob:  &VJob{},
			},
			want: JobTypeWhole,
		},
		{
			name: "03-SetJobType ReqNPUNum is zero",
			nJob: &NPUJob{
				Tasks: map[api.TaskID]NPUTask{"task01": {ReqNPUNum: 1, VTask: &VTask{Type: JobTypeWhole}},
					"task00": {ReqNPUNum: 0, VTask: &VTask{Type: JobTypeWhole}}},
				VJob: &VJob{},
			},
			want: JobTypeWhole,
		},
	}
}

func TestSetVJobType(t *testing.T) {
	tests := buildSetVJobTypeTestCase()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.nJob.SetJobType()
			if got := tt.nJob.Type; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testSetVJobStatusByInfTest struct {
	name  string
	nJob  *NPUJob
	vcJob *api.JobInfo
	want  string
}

func buildTestSetVJobStatusByInfTest() []testSetVJobStatusByInfTest {
	tests := []testSetVJobStatusByInfTest{
		{
			name:  "01-test SetJobStatusByInf ",
			nJob:  &NPUJob{VJob: &VJob{}},
			vcJob: &api.JobInfo{PodGroup: &api.PodGroup{}},
			want:  "",
		},
	}
	return tests
}

func TestSetVJobStatusByInf(t *testing.T) {
	tests := buildTestSetVJobStatusByInfTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.nJob.SetJobStatusByInf(tt.vcJob)
			if got := tt.nJob.Status; got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSchedulerJobAttrIsJobSinglePodDelete(t *testing.T) {
	type fields struct {
		ComJob ComJob
		NPUJob *NPUJob
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "01-IsJobSinglePodDelete  SchedulingTaskNum equal npu tasks",
			fields: fields{
				NPUJob: &NPUJob{
					SchedulingTaskNum: 1,
					Tasks:             map[api.TaskID]NPUTask{"task01": {ReqNPUNum: 1}},
				},
			},
			want: false,
		},
		{
			name: "02-IsJobSinglePodDelete SchedulingTaskNum not equal npu tasks",
			fields: fields{
				NPUJob: &NPUJob{
					SchedulingTaskNum: 1,
					Tasks: map[api.TaskID]NPUTask{
						"task01": {ReqNPUNum: 1}, "task00": {ReqNPUNum: 1}},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sJob := SchedulerJobAttr{
				ComJob: tt.fields.ComJob,
				NPUJob: tt.fields.NPUJob,
			}
			if got := sJob.IsJobSinglePodDelete(); got != tt.want {
				t.Errorf("IsJobSinglePodDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNPUJobIsNPUJob(t *testing.T) {
	type fields struct {
		ReqNPUName string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "01-IsNPUJob npu job",
			fields: fields{ReqNPUName: AscendNPUCore},
			want:   true,
		},
		{
			name:   "02-IsNPUJob not npu job",
			fields: fields{ReqNPUName: ""},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nJob := &NPUJob{
				ReqNPUName: tt.fields.ReqNPUName,
			}
			if got := nJob.IsNPUJob(); got != tt.want {
				t.Errorf("IsNPUJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNPUJobGetNPUTaskNumInJob(t *testing.T) {
	type fields struct {
		ReqNPUName string
		Tasks      map[api.TaskID]NPUTask
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "01-GetNPUTaskNumInJob not npu job",
			fields: fields{ReqNPUName: ""},
			want:   0,
		},
		{
			name: "02-GetNPUTaskNumInJob npu job",
			fields: fields{
				ReqNPUName: AscendNPUCore,
				Tasks:      map[api.TaskID]NPUTask{"task01": {ReqNPUName: Ascend910bName}},
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
			if got := nJob.GetNPUTaskNumInJob(); got != tt.want {
				t.Errorf("GetNPUTaskNumInJob() = %v, want %v", got, tt.want)
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
