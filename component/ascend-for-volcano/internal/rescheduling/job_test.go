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
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package rescheduling

import (
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type FaultJobTestField struct {
	ReScheduleKey       string
	IsFaultJob          bool
	IsInSession         bool
	JobName             string
	JobUID              api.JobID
	JobNamespace        string
	JobRankIds          []string
	NodeNames           []string
	FaultTasks          []FaultTask
	UpdateTime          int64
	JobRankIdCreateTime int64
	FaultTypes          []string
	DeleteExecutedFlag  bool
	ElasticScheduling   string
}

type FaultJobForceDeleteJobArgs struct {
	ssn             *framework.Session
	schedulerJob    *plugin.SchedulerJob
	cacheFuncBefore func()
	cacheFuncAfter  func()
}

type FaultJobForceDeleteJobTests struct {
	name    string
	fields  FaultJobTestField
	args    FaultJobForceDeleteJobArgs
	wantErr bool
}

func buildFaultJobForceDeleteJobTests() []FaultJobForceDeleteJobTests {
	var tmpPatch *gomonkey.Patches
	faultTask1 := fakeReSchedulerFaultTask(true, []string{"pod0", "vcjob", "node0", "job0", "0"}, 0,
		"ppppppppppppp")
	faultTask2 := fakeReSchedulerFaultTask(false, []string{"pod1", "vcjob", "node1", "job0", "1"}, 0,
		"ppppppppppppp")
	schedulerJob := fakeSchedulerJobEmptyTask("job0", "vcjob")
	test1 := FaultJobForceDeleteJobTests{
		name: "01-FaultJobForceDeleteJob()-delete success",
		fields: FaultJobTestField{
			JobName:             "job0",
			JobUID:              "vcjob/job0",
			JobNamespace:        "vcjob",
			JobRankIds:          nil,
			NodeNames:           nil,
			FaultTasks:          []FaultTask{faultTask1, faultTask2},
			UpdateTime:          0,
			JobRankIdCreateTime: 0,
			FaultTypes:          nil,
			DeleteExecutedFlag:  false,
		},
		args: FaultJobForceDeleteJobArgs{
			ssn:          test.FakeSSNReSchedule(),
			schedulerJob: &schedulerJob,
			cacheFuncBefore: func() {
				tmpPatch = gomonkey.ApplyMethod(reflect.TypeOf(&FaultTask{}), "DeleteRealPodByTask",
					func(_ *FaultTask, _ *framework.Session, _ int64) error { return nil })
			},
			cacheFuncAfter: func() {
				if tmpPatch != nil {
					tmpPatch.Reset()
				}
			},
		},
		wantErr: false,
	}
	tests := []FaultJobForceDeleteJobTests{
		test1,
	}
	return tests
}

// TestFaultJobForceDeleteJob test for force delete function
func TestFaultJobForceDeleteJob(t *testing.T) {
	tests := buildFaultJobForceDeleteJobTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFuncBefore()
			fJob := &FaultJob{
				ReScheduleKey:       tt.fields.ReScheduleKey,
				IsFaultJob:          tt.fields.IsFaultJob,
				IsInSession:         tt.fields.IsInSession,
				JobName:             tt.fields.JobName,
				JobUID:              tt.fields.JobUID,
				JobNamespace:        tt.fields.JobNamespace,
				JobRankIds:          tt.fields.JobRankIds,
				NodeNames:           tt.fields.NodeNames,
				FaultTasks:          tt.fields.FaultTasks,
				UpdateTime:          tt.fields.UpdateTime,
				JobRankIdCreateTime: tt.fields.JobRankIdCreateTime,
				FaultTypes:          tt.fields.FaultTypes,
				DeleteExecutedFlag:  tt.fields.DeleteExecutedFlag,
				ElasticScheduling:   tt.fields.ElasticScheduling,
			}
			if err := fJob.ForceDeleteJob(tt.args.ssn, tt.args.schedulerJob); (err != nil) != tt.wantErr {
				t.Errorf("ForceDeleteJob() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.args.cacheFuncAfter()
		})
	}
}
