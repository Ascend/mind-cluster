/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package superpod for test score functions
package superpod

import (
	"testing"

	"k8s.io/api/core/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

const (
	testJobID = "0"
)

type tasksCommonTestCase struct {
	name  string
	tasks map[api.TaskID]util.NPUTask
}

type obtainOriginalRankIdTestCase struct {
	tasksCommonTestCase
	want int
}

func buildObtainOriginalRankIdTestCase() []obtainOriginalRankIdTestCase {
	return []obtainOriginalRankIdTestCase{
		{
			tasksCommonTestCase: tasksCommonTestCase{
				name: "01-obtainOriginalRankIdMap get all pending pod rankId map",
				tasks: map[api.TaskID]util.NPUTask{
					testJobID: {
						ReqNPUName: util.HwPreName + util.Ascend910,
						PodStatus:  v1.PodPending,
						Annotation: map[string]string{
							plugin.PodRankIndexKey: "0",
						},
					},
				},
			},
			want: 1,
		},
		{
			tasksCommonTestCase: tasksCommonTestCase{
				name: "02-obtainOriginalRankIdMap get empty rankId map with all running pod",
				tasks: map[api.TaskID]util.NPUTask{
					testJobID: {
						ReqNPUName: util.HwPreName + util.Ascend910,
						PodStatus:  v1.PodRunning,
						Annotation: map[string]string{
							plugin.PodRankIndexKey: "0",
						},
					},
				},
			},
			want: 0,
		},
		{
			tasksCommonTestCase: tasksCommonTestCase{
				name: "03-obtainOriginalRankIdMap get empty rankId map with empty hccl/rankIndex of pending pod",
				tasks: map[api.TaskID]util.NPUTask{
					testJobID: {
						ReqNPUName: util.HwPreName + util.Ascend910,
						PodStatus:  v1.PodPending,
						Annotation: map[string]string{},
					},
				},
			},
			want: 1,
		},
	}
}

func TestObtainOriginalRankIdMap(t *testing.T) {
	for _, cs := range buildObtainOriginalRankIdTestCase() {
		t.Run(cs.name, func(t *testing.T) {
			job := plugin.SchedulerJob{
				JobReadyTag: new(bool),
				SchedulerJobAttr: util.SchedulerJobAttr{
					NPUJob: &util.NPUJob{
						Tasks: cs.tasks,
					},
				},
				SuperPods: map[string][]plugin.SuperNode{},
			}
			res := obtainOriginalRankIdMap(&job)
			if len(res) != cs.want {
				t.Errorf("got %v; want %v", res, cs.want)
			}
		})
	}
}

// TestScoreNodeBatchForReadyJob test of scoreNodeBatchForReadyJob
func TestScoreNodeBatchForReadyJob(t *testing.T) {
	plg := New(SuperPodx8SchedulerName)
	plg.Name = "job1"
	plg.SchedulerJobAttr = util.SchedulerJobAttr{
		ComJob: util.ComJob{},
		NPUJob: &util.NPUJob{},
	}
	plg.ScheduleEnv = plugin.ScheduleEnv{}
	type args struct {
		task *api.TaskInfo
		job  *plugin.SchedulerJob
		sMap map[string]float64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "01-scoreNodeBatchForReadyJob invalid argument",
			args: args{},
		},
		{
			name: "02-scoreNodeBatchForReadyJob rankIdMap empty",
			args: args{
				task: test.FakeNormalTestTask("pod1", "node1", "acjob"),
				job:  &plugin.SchedulerJob{},
				sMap: map[string]float64{"node1": 0},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plg.scoreNodeBatchForReadyJob(tt.args.task, tt.args.job, tt.args.sMap)
		})
	}
}
