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
Package main is using for HuaWei Ascend pin affinity schedule.
*/
package main

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

type handlerStartTest struct {
	name string
	want *plugin.ScheduleHandler
}

func buildTestHandlerStartTestCases() []handlerStartTest {
	testCases := []handlerStartTest{
		{
			name: "HandlerStart ok test",
			want: &plugin.ScheduleHandler{},
		},
	}
	return testCases
}

func TestHandlerStart(t *testing.T) {
	tests := buildTestHandlerStartTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HandlerStart(); got == nil {
				t.Errorf("HandlerStart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestName(t *testing.T) {
	tests := []struct {
		name string
		tp   *huaweiNPUPlugin
		want string
	}{
		{
			name: "01-Name ok test",
			tp:   &huaweiNPUPlugin{},
			want: PluginName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tp.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		arguments framework.Arguments
	}
	tests := []struct {
		name string
		args args
		want framework.Plugin
	}{
		{
			name: "New ok test",
			args: args{arguments: framework.Arguments{PluginName: "haha"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.arguments); got == nil {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fields struct {
	Scheduler *plugin.ScheduleHandler
	Arguments framework.Arguments
}

type args struct {
	ssn            *framework.Session
	cacheFunBefore func()
	cacheFunAfter  func()
}

type onSessionOpenTest struct {
	name   string
	fields fields
	args   args
}

func buildOnSessionOpenTestCases() []onSessionOpenTest {
	tests := []onSessionOpenTest{
		{
			name:   "OnSessionOpen test ssn nil ok",
			fields: fields{Scheduler: HandlerStart()},
			args:   args{ssn: nil, cacheFunBefore: func() {}, cacheFunAfter: func() {}},
		},
	}
	return tests
}

func TestOnSessionOpen(t *testing.T) {
	tests := buildOnSessionOpenTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &huaweiNPUPlugin{
				Scheduler: tt.fields.Scheduler,
				Arguments: tt.fields.Arguments,
			}
			tt.args.cacheFunBefore()
			tp.OnSessionOpen(tt.args.ssn)
			tt.args.cacheFunAfter()
		})
	}
}

type onSessionCloseTest struct {
	name   string
	fields fields
	args   args
}

func buildOnSessionCloseTestCases() []onSessionCloseTest {
	testSsn := test.FakeNormalSSN()
	tests := []onSessionCloseTest{
		{
			name:   "OnSessionCloseTestCases test ok",
			fields: fields{Scheduler: HandlerStart()},
			args:   args{ssn: testSsn},
		},
	}
	return tests
}

func TestOnSessionClose(t *testing.T) {
	tests := buildOnSessionCloseTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &huaweiNPUPlugin{
				Scheduler: tt.fields.Scheduler,
				Arguments: tt.fields.Arguments,
			}
			tp.OnSessionClose(tt.args.ssn)
		})
	}
}

type JobPipelinedFnTest struct {
	name string
	tp   *huaweiNPUPlugin
	obj  interface{}
	want int
}

func buildJobPipelinedFnTestCases() []JobPipelinedFnTest {
	return []JobPipelinedFnTest{
		{
			name: "01 JobPipelinedFnTest will return Reject when obj is not job info",
			tp:   &huaweiNPUPlugin{},
			obj:  nil,
			want: util.Reject,
		},
		{
			name: "02 JobPipelinedFnTest will return Abstain when scheduler job is not exist",
			tp: &huaweiNPUPlugin{Scheduler: &plugin.ScheduleHandler{ScheduleEnv: plugin.ScheduleEnv{
				Jobs: map[api.JobID]plugin.SchedulerJob{}}}},
			obj:  &api.JobInfo{},
			want: util.Abstain,
		},
		{
			name: "03 JobPipelinedFnTest will return Abstain when job ready tag is true",
			tp: &huaweiNPUPlugin{Scheduler: &plugin.ScheduleHandler{ScheduleEnv: plugin.ScheduleEnv{
				Jobs: map[api.JobID]plugin.SchedulerJob{"test-name": {JobReadyTag: true}}}}},
			obj:  &api.JobInfo{UID: "test-name"},
			want: util.Abstain,
		},
		{
			name: "04 JobPipelinedFnTest will return Reject when job ready tag is false",
			tp: &huaweiNPUPlugin{Scheduler: &plugin.ScheduleHandler{ScheduleEnv: plugin.ScheduleEnv{
				Jobs: map[api.JobID]plugin.SchedulerJob{"test-name": {JobReadyTag: false}}}}},
			obj:  &api.JobInfo{UID: "test-name"},
			want: util.Reject,
		},
	}
}

func TestJobPipelinedFn(t *testing.T) {
	tests := buildJobPipelinedFnTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tp.jobPipelinedFn(tt.obj); got != tt.want {
				t.Errorf("jobPipelinedFn() = %v, want %v", got, tt.want)
			}
		})
	}
}
