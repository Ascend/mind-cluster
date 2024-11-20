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
	"fmt"
	"reflect"
	"testing"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/config"

	"k8s.io/api/core/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func TestGetJobLabelFromVcJob(t *testing.T) {
	tJob := test.FakeNormalTestJob("test1", 1)
	test.AddTestJobLabel(tJob, "haha", "who")
	type args struct {
		job *api.JobInfo
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "01-GetJobLabelFromVcJob get ok test",
			args: args{job: tJob},
			want: map[string]string{"haha": "who"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetJobLabelFromVcJob(tt.args.job); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetJobLabelFromVcJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

type getJobNPUTasksArgs struct {
	vcJob *api.JobInfo
}

type getJobNPUTasksTest struct {
	name string
	args getJobNPUTasksArgs
	want map[api.TaskID]util.NPUTask
}

func buildGetJobNPUTasksTest() []getJobNPUTasksTest {
	tJob1 := test.FakeNormalTestJob("test1", 1)
	test.AddFakeTaskResReq(tJob1.Tasks[`"vcjob"-"pod"`], util.NPU910CardName, util.NPUIndex8)
	tests := []getJobNPUTasksTest{
		{
			name: "01-GetJobNPUTasks job nil test.",
			args: getJobNPUTasksArgs{vcJob: nil},
			want: nil,
		},
	}
	return tests
}

func TestGetJobNPUTasks(t *testing.T) {
	tests := buildGetJobNPUTasksTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetJobNPUTasks(tt.args.vcJob); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetJobNPUTasks() =%+v, want %+v", got, tt.want)
			}
		})
	}
}

type getJobSelectorFromVcJobArgs struct {
	job *api.JobInfo
}

type getJobSelectorFromVcJobTest struct {
	name string
	args getJobSelectorFromVcJobArgs
	want map[string]string
}

func buildGetJobSelectorFromVcJobTest() []getJobSelectorFromVcJobTest {
	tJob1 := test.FakeNormalTestJob("test1", 1)
	test.AddTestJobLabel(tJob1, "haha", "heihei")
	tJob2 := test.FakeNormalTestJob("test1", 1)
	tests := []getJobSelectorFromVcJobTest{
		{
			name: "01-GetJobSelectorFromVcJob nil job selector test",
			args: getJobSelectorFromVcJobArgs{job: tJob2},
			want: make(map[string]string, util.MapInitNum),
		},
		{
			name: "02-GetJobSelectorFromVcJob nil job selector test",
			args: getJobSelectorFromVcJobArgs{job: tJob1},
			want: map[string]string{"haha": "heihei"},
		},
	}
	return tests
}

func TestGetJobSelectorFromVcJob(t *testing.T) {
	tests := buildGetJobSelectorFromVcJobTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetJobSelectorFromVcJob(tt.args.job); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetJobSelectorFromVcJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTaskSelectors(t *testing.T) {
	tTasks := test.FakeNormalTestTasks(util.NPUIndex2)
	test.AddTestTaskLabel(tTasks[1], "haha", "who")
	type args struct {
		task *api.TaskInfo
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "01-GetTaskSelectors no selector test.",
			args: args{task: tTasks[0]},
			want: nil,
		},
		{
			name: "02-GetTaskSelectors ok test.",
			args: args{task: tTasks[1]},
			want: map[string]string{"haha": "who"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTaskSelectors(tt.args.task); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTaskSelectors() = %v, want %v", got, tt.want)
			}
		})
	}
}

type getVCJobReqNPUTypeFromJobInfoArgs struct {
	vcJob *api.JobInfo
}

type getVCJobReqNPUTypeFromJobInfoTest struct {
	name    string
	args    getVCJobReqNPUTypeFromJobInfoArgs
	want    string
	want1   int
	wantErr bool
}

func buildGetVCJobReqNPUTypeFromJobInfoTest() []getVCJobReqNPUTypeFromJobInfoTest {
	tJob1 := test.FakeNormalTestJob("test1", 1)
	test.SetFakeJobRequestSource(tJob1, util.NPU910CardName, util.NPUIndex8)
	tests := []getVCJobReqNPUTypeFromJobInfoTest{
		{
			name:    "01-GetVCJobReqNPUTypeFromJobInfo nil job test.",
			args:    getVCJobReqNPUTypeFromJobInfoArgs{vcJob: nil},
			want:    "",
			want1:   0.0,
			wantErr: true,
		},
		{
			name:    "03-GetVCJobReqNPUTypeFromJobInfo ok test.",
			args:    getVCJobReqNPUTypeFromJobInfoArgs{vcJob: tJob1},
			want:    util.NPU910CardName,
			want1:   util.NPUIndex8,
			wantErr: false,
		},
	}
	return tests
}

func TestGetVCJobReqNPUTypeFromJobInfo(t *testing.T) {
	tests := buildGetVCJobReqNPUTypeFromJobInfoTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := GetVCJobReqNPUTypeFromJobInfo(tt.args.vcJob)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVCJobReqNPUTypeFromJobInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetVCJobReqNPUTypeFromJobInfo() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetVCJobReqNPUTypeFromJobInfo() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

type getVCTaskReqNPUTypeFromTaskInfoArgs struct {
	vcTask *api.TaskInfo
}

type getVCTaskReqNPUTypeFromTaskInfoTest struct {
	name  string
	args  getVCTaskReqNPUTypeFromTaskInfoArgs
	want  string
	want1 int
}

func buildGetVCTaskReqNPUTypeFromTaskInfoTest() []getVCTaskReqNPUTypeFromTaskInfoTest {
	tTasks := test.FakeNormalTestTasks(1)
	tests := []getVCTaskReqNPUTypeFromTaskInfoTest{
		{
			name:  "01-GetVCTaskReqNPUTypeFromTaskInfo nil test.",
			args:  getVCTaskReqNPUTypeFromTaskInfoArgs{vcTask: nil},
			want:  "",
			want1: 0,
		},
		{
			name:  "02-GetVCTaskReqNPUTypeFromTaskInfo ok test.",
			args:  getVCTaskReqNPUTypeFromTaskInfoArgs{vcTask: tTasks[0]},
			want:  util.NPU910CardName,
			want1: util.NPUIndex8,
		},
	}
	return tests
}

func TestGetVCTaskReqNPUTypeFromTaskInfo(t *testing.T) {
	tests := buildGetVCTaskReqNPUTypeFromTaskInfoTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetVCTaskReqNPUTypeFromTaskInfo(tt.args.vcTask)
			if got != tt.want {
				t.Errorf("GetVCTaskReqNPUTypeFromTaskInfo() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetVCTaskReqNPUTypeFromTaskInfo() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

type isJobInitialArgs struct {
	job *api.JobInfo
}

type isJobInitialTest struct {
	name string
	args isJobInitialArgs
	want bool
}

func buildGIsJobInitialTest() []isJobInitialTest {
	tJob := test.FakeNormalTestJob("testJob", 1)
	tJob2 := test.FakeNormalTestJob("testJob2", 1)
	test.SetFakeNPUJobStatusPending(tJob2)
	tests := []isJobInitialTest{
		{
			name: "01-IsJobInitial pending test",
			args: isJobInitialArgs{job: tJob2},
			want: true,
		},
		{
			name: "02-IsJobInitial test ok",
			args: isJobInitialArgs{job: tJob},
			want: true,
		},
	}
	return tests
}

func TestIsJobInitial(t *testing.T) {
	tests := buildGIsJobInitialTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsJobInitial(tt.args.job); got != tt.want {
				t.Errorf("IsJobInitial() = %v, want %v", got, tt.want)
			}
		})
	}
}

type jobValidArgs struct {
	obj interface{}
}

type jobValidTest struct {
	name   string
	fields fields
	args   jobValidArgs
	want   *api.ValidateResult
}

func buildJobValidTest() []jobValidTest {
	tJob := test.FakeNormalTestJob("testJob", 1)
	tJob2 := test.FakeNormalTestJob("testJob2", 1)
	test.SetFakeNPUJobStatusPending(tJob2)
	tJob3 := test.FakeNormalTestJob("testJob", 1)
	test.AddTestJobLabel(tJob3, "haha", "who")
	tests := []jobValidTest{
		{
			name:   "01-JobValid not job test.",
			fields: fields{},
			args:   jobValidArgs{obj: "haha"},
			want: &api.ValidateResult{Pass: false, Reason: "job convert failed",
				Message: `validJobFn ["haha"] failed:job convert failed`},
		},
		{
			name:   "02-JobValid job not initial test.",
			fields: fields{},
			args:   jobValidArgs{obj: tJob2},
			want:   nil,
		},
		{
			name: "03-JobValid job not in jobs test.",
			fields: fields{NPUPlugins: map[string]NPUBuilder{},
				ScheduleEnv: ScheduleEnv{Jobs: map[api.JobID]SchedulerJob{}}},
			args: jobValidArgs{obj: tJob},
			want: nil,
		},
		{
			name: "04-JobValid job no selector test.",
			fields: fields{NPUPlugins: map[string]NPUBuilder{},
				ScheduleEnv: ScheduleEnv{Jobs: map[api.JobID]SchedulerJob{tJob.
					UID: {SchedulerJobAttr: util.SchedulerJobAttr{ComJob: util.ComJob{Name: tJob.UID}}}}}},
			args: jobValidArgs{obj: tJob},
			want: &api.ValidateResult{Pass: false, Reason: "Job selector error",
				Message: fmt.Errorf("%s or vcFrame's selectors nil", tJob.UID).Error()},
		},
	}
	return tests
}

func TestJobValid(t *testing.T) {
	tests := buildJobValidTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := &ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			if got := sHandle.JobValid(tt.args.obj); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JobValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

type setJobPendReasonByNodesCaseArgs struct {
	job *api.JobInfo
}

type setJobPendReasonByNodesCaseTest struct {
	name   string
	fields fields
	args   setJobPendReasonByNodesCaseArgs
}

func buildSetJobPendReasonByNodesCaseTest() []setJobPendReasonByNodesCaseTest {
	tJob := test.FakeNormalTestJob("testJob", 1)
	tJob1 := test.FakeNormalTestJob("testJob1", 1)
	test.SetFakeNPUJobErrors(tJob1, "haha")
	tests := []setJobPendReasonByNodesCaseTest{
		{
			name: "01-SetJobPendReasonByNodesCase job no error test.",
			fields: fields{NPUPlugins: map[string]NPUBuilder{},
				ScheduleEnv: ScheduleEnv{
					Jobs:      map[api.JobID]SchedulerJob{},
					Nodes:     map[string]NPUNode{},
					FrameAttr: VolcanoFrame{}}},
			args: setJobPendReasonByNodesCaseArgs{job: tJob},
		},
		{
			name: "02-SetJobPendReasonByNodesCase test ok.",
			fields: fields{NPUPlugins: map[string]NPUBuilder{},
				ScheduleEnv: ScheduleEnv{
					Jobs:      map[api.JobID]SchedulerJob{},
					Nodes:     map[string]NPUNode{},
					FrameAttr: VolcanoFrame{}}},
			args: setJobPendReasonByNodesCaseArgs{job: tJob1},
		},
	}
	return tests
}

func TestSetJobPendReasonByNodesCase(t *testing.T) {
	tests := buildSetJobPendReasonByNodesCaseTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			sHandle.SetJobPendReasonByNodesCase(tt.args.job)
		})
	}
}

type setJobPendingReasonArgs struct {
	vcJob  *api.JobInfo
	reason interface{}
}

type setJobPendingReasonTest struct {
	name    string
	fields  fields
	args    setJobPendingReasonArgs
	wantErr bool
}

func buildSetJobPendingReasonTest() []setJobPendingReasonTest {
	tJob := test.FakeNormalTestJob("testJob", 1)
	tests := []setJobPendingReasonTest{
		{
			name: "01-SetJobPendingReason nil test.",
			fields: fields{NPUPlugins: map[string]NPUBuilder{},
				ScheduleEnv: ScheduleEnv{
					Jobs:      map[api.JobID]SchedulerJob{},
					Nodes:     map[string]NPUNode{},
					FrameAttr: VolcanoFrame{}}},
			args:    setJobPendingReasonArgs{vcJob: nil},
			wantErr: true,
		},
		{
			name: "02-SetJobPendingReason not support type test.",
			fields: fields{NPUPlugins: map[string]NPUBuilder{},
				ScheduleEnv: ScheduleEnv{
					Jobs:      map[api.JobID]SchedulerJob{},
					Nodes:     map[string]NPUNode{},
					FrameAttr: VolcanoFrame{}}},
			args:    setJobPendingReasonArgs{vcJob: tJob, reason: api.NodeInfo{}},
			wantErr: true,
		},
		{
			name: "03-SetJobPendingReason string type test.",
			fields: fields{NPUPlugins: map[string]NPUBuilder{},
				ScheduleEnv: ScheduleEnv{
					Jobs:      map[api.JobID]SchedulerJob{},
					Nodes:     map[string]NPUNode{},
					FrameAttr: VolcanoFrame{}}},
			args:    setJobPendingReasonArgs{vcJob: tJob, reason: "haha"},
			wantErr: false,
		},
		{
			name: "04-SetJobPendingReason nodeErrors test.",
			fields: fields{NPUPlugins: map[string]NPUBuilder{},
				ScheduleEnv: ScheduleEnv{
					Jobs:      map[api.JobID]SchedulerJob{},
					Nodes:     map[string]NPUNode{},
					FrameAttr: VolcanoFrame{}}},
			args:    setJobPendingReasonArgs{vcJob: tJob, reason: map[api.TaskID]*api.FitErrors{}},
			wantErr: false,
		},
	}
	return tests
}

func TestScheduleHandlerSetJobPendingReason(t *testing.T) {
	tests := buildSetJobPendingReasonTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := &ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			if err := sHandle.SetJobPendingReason(tt.args.vcJob, tt.args.reason); (err != nil) != tt.wantErr {
				t.Errorf("SetJobPendingReason() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type schedulerJobFields struct {
	SchedulerJobAttr util.SchedulerJobAttr
	handler          ISchedulerPlugin
}

type CheckNodeNumArgs struct {
	taskInfo *api.TaskInfo
	vcNode   NPUNode
}

type CheckNodeNumTest struct {
	name    string
	fields  schedulerJobFields
	args    CheckNodeNumArgs
	wantErr bool
}

func buildCheckNodeNumTest() []CheckNodeNumTest {
	tTasks := test.FakeNormalTestTasks(1)
	tNode1 := NPUNode{CommonNode: CommonNode{
		Name: "testNode1", Idle: map[v1.ResourceName]float64{util.NPU910CardName: util.NPUIndex2}},
	}
	tNode2 := NPUNode{CommonNode: CommonNode{Name: "testNode2", Idle: map[v1.ResourceName]float64{
		util.NPU910CardName: util.NPUIndex8 * util.NPUHexKilo}}}
	tests := []CheckNodeNumTest{
		{
			name:    "01-CheckNodeNum no task request test.",
			fields:  schedulerJobFields{SchedulerJobAttr: util.SchedulerJobAttr{}},
			args:    CheckNodeNumArgs{taskInfo: nil},
			wantErr: true,
		},
		{
			name: "02-CheckNodeNum no task test.",
			fields: schedulerJobFields{SchedulerJobAttr: util.SchedulerJobAttr{
				NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{}}}},
			args: CheckNodeNumArgs{taskInfo: tTasks[0], vcNode: NPUNode{CommonNode{Name: "testNode1", Idle: nil},
				VNode{}, 0}},
			wantErr: true,
		},
		{
			name: "03-CheckNodeNum no node idle test.",
			fields: schedulerJobFields{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.
				NPUJob{Tasks: map[api.TaskID]util.NPUTask{tTasks[0].UID: {Name: tTasks[0].Name,
				ReqNPUName: util.NPU910CardName, ReqNPUNum: util.NPUIndex8}}}}},
			args: CheckNodeNumArgs{taskInfo: tTasks[0], vcNode: NPUNode{CommonNode{Name: "testNode1", Idle: nil},
				VNode{}, 0}},
			wantErr: true,
		},
		{
			name: "04-CheckNodeNum not meet test.",
			fields: schedulerJobFields{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.
				NPUJob{Tasks: map[api.TaskID]util.NPUTask{tTasks[0].UID: {Name: tTasks[0].Name,
				ReqNPUName: util.NPU910CardName, ReqNPUNum: util.NPUIndex8}}}}},
			args:    CheckNodeNumArgs{taskInfo: tTasks[0], vcNode: tNode1},
			wantErr: true,
		},
		{
			name: "05-CheckNodeNum meet test.",
			fields: schedulerJobFields{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.
				NPUJob{Tasks: map[api.TaskID]util.NPUTask{tTasks[0].UID: {Name: tTasks[0].Name,
				ReqNPUName: util.NPU910CardName, ReqNPUNum: util.NPUIndex8}}}}},
			args:    CheckNodeNumArgs{taskInfo: tTasks[0], vcNode: tNode2},
			wantErr: false,
		},
	}
	return tests
}

func TestSchedulerJobCheckNodeNum(t *testing.T) {
	tests := buildCheckNodeNumTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sJob := &SchedulerJob{
				SchedulerJobAttr: tt.fields.SchedulerJobAttr,
				handler:          tt.fields.handler,
			}
			if err := sJob.CheckNodeNum(tt.args.taskInfo, tt.args.vcNode); (err != nil) != tt.wantErr {
				t.Errorf("CheckNodeNum() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type initArgs struct {
	vcJob   *api.JobInfo
	sHandle *ScheduleHandler
}

type initTest struct {
	name    string
	fields  schedulerJobFields
	args    initArgs
	wantErr bool
}

func buildInitTest() []initTest {
	tJob := test.FakeNormalTestJob("haha", 1)
	tests := []initTest{
		{
			name:    "01-Init vcJob nil test.",
			fields:  schedulerJobFields{SchedulerJobAttr: util.SchedulerJobAttr{}},
			args:    initArgs{vcJob: nil},
			wantErr: true,
		},
		{
			name:    "02-Init plugin not register test.",
			fields:  schedulerJobFields{SchedulerJobAttr: util.SchedulerJobAttr{}},
			args:    initArgs{vcJob: tJob},
			wantErr: true,
		},
	}
	return tests
}

func TestSchedulerJobInit(t *testing.T) {
	tests := buildInitTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sJob := &SchedulerJob{
				SchedulerJobAttr: tt.fields.SchedulerJobAttr,
				handler:          tt.fields.handler,
			}
			if err := sJob.Init(tt.args.vcJob, tt.args.sHandle); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type validJobSelectorArgs struct {
	vcFrame VolcanoFrame
}

type validJobSelectorTest struct {
	name    string
	fields  schedulerJobFields
	args    validJobSelectorArgs
	wantErr bool
}

func buildValidJobSelectorTest() []validJobSelectorTest {
	tests := []validJobSelectorTest{
		{
			name:    "01-ValidJobSelector nil test.",
			fields:  schedulerJobFields{},
			args:    validJobSelectorArgs{},
			wantErr: true,
		},
		{
			name: "02-ValidJobSelector selector not meet test.",
			fields: schedulerJobFields{SchedulerJobAttr: util.SchedulerJobAttr{ComJob: util.ComJob{
				Name: "haha", Selector: map[string]string{"heihei": "what?"},
			}}},
			args: validJobSelectorArgs{vcFrame: VolcanoFrame{Confs: []config.
				Configuration{{Arguments: map[string]string{"heihei": "why?"}}}}},
			wantErr: true,
		},
		{
			name: "03-ValidJobSelector ok test.",
			fields: schedulerJobFields{SchedulerJobAttr: util.SchedulerJobAttr{ComJob: util.ComJob{
				Name: "haha", Selector: map[string]string{"heihei": "oh"},
			}}},
			args: validJobSelectorArgs{vcFrame: VolcanoFrame{Confs: []config.
				Configuration{{Arguments: map[string]string{"heihei": "oh"}}}}},
			wantErr: false,
		},
	}
	return tests
}

func TestValidJobSelector(t *testing.T) {
	tests := buildValidJobSelectorTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sJob := SchedulerJob{
				SchedulerJobAttr: tt.fields.SchedulerJobAttr,
				handler:          tt.fields.handler,
			}
			if err := sJob.ValidJobSelector(tt.args.vcFrame); (err != nil) != tt.wantErr {
				t.Errorf("ValidJobSelector() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
