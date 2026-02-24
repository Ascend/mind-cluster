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

// Package chip8node8ra64sp for test score functions
package chip8node8ra64sp

import (
	"reflect"
	"strconv"
	"testing"

	"k8s.io/api/core/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

const (
	testJobID  = "0"
	workerSpec = "worker"
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

const (
	batchScoreNpuTask4 = 4
	batchScoreRank0    = 0
)

func createObtainBatchScoreRankTaskInfo(jobId, rankId, spec string) *api.TaskInfo {
	task := test.FakeNormalTestTask("pod1", "node1", "acjob")
	task.Job = api.JobID(jobId)
	task.Pod.Annotations[plugin.PodRankIndexKey] = rankId
	task.Pod.Annotations[TaskSpecAnno] = spec
	return task
}

func createBatchScoreNPUTasks(n int) map[api.TaskID]util.NPUTask {
	tasks := make(map[api.TaskID]util.NPUTask, n)
	for i := 0; i < n; i++ {
		spec := workerSpec
		if batchScoreRank0 == i {
			spec = SchedulerType
		}
		tasks[api.TaskID(strconv.Itoa(i))] = util.NPUTask{
			Name:       "task" + strconv.Itoa(i),
			ReqNPUName: util.NPU910CardName,
			Annotation: map[string]string{
				plugin.PodRankIndexKey: strconv.Itoa(i),
				TaskSpecAnno:           spec,
			},
			PodStatus: v1.PodPending,
		}
	}
	return tasks
}

func fakeSchedulerJobEmptyTask(jobName, namespace string) *plugin.SchedulerJob {
	job := &plugin.SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{
				Name:      api.JobID(jobName),
				NameSpace: namespace,
				Selector:  map[string]string{},
				Label:     map[string]string{},
			},
			NPUJob: &util.NPUJob{
				ReqNPUName: util.NPU910CardName,
				ReqNPUNum:  0,
				Tasks:      make(map[api.TaskID]util.NPUTask),
			},
		},
	}
	return job
}

type obtainBatchScoreRankArgs struct {
	task *api.TaskInfo
	job  *plugin.SchedulerJob
}

type obtainBatchScoreRankTest struct {
	name string
	args obtainBatchScoreRankArgs
	want map[int]struct{}
}

func getObtainBatchScoreRankTestCases(jobId string) []obtainBatchScoreRankTest {
	schedulerJob := fakeSchedulerJobEmptyTask(jobId, "")
	schedulerJob.Tasks = createBatchScoreNPUTasks(batchScoreNpuTask4)
	taskInfoWithoutSpecAnno := createObtainBatchScoreRankTaskInfo(jobId, "1", SchedulerType)
	delete(taskInfoWithoutSpecAnno.Pod.Annotations, TaskSpecAnno)
	tests := []obtainBatchScoreRankTest{
		{
			name: "01-obtainBatchScoreRank invalid argument",
			args: obtainBatchScoreRankArgs{},
			want: nil,
		},
		{
			name: "02-obtainBatchScoreRank spec not exist",
			args: obtainBatchScoreRankArgs{
				task: taskInfoWithoutSpecAnno,
				job:  schedulerJob,
			},
			want: nil,
		},
		{
			name: "03-obtainBatchScoreRank spec " + SchedulerType,
			args: obtainBatchScoreRankArgs{
				task: createObtainBatchScoreRankTaskInfo(jobId, "1", SchedulerType),
				job:  schedulerJob,
			},
			want: map[int]struct{}{0: {}},
		},
		{
			name: "04-obtainBatchScoreRank spec " + workerSpec,
			args: obtainBatchScoreRankArgs{
				task: createObtainBatchScoreRankTaskInfo(jobId, "1", workerSpec),
				job:  schedulerJob,
			},
			want: map[int]struct{}{1: {}, 2: {}, 3: {}},
		},
	}
	return tests
}

// TestObtainBatchScoreRank test of obtainBatchScoreRank
func TestObtainBatchScoreRank(t *testing.T) {
	jobId := "job1"
	plg := New(util.SuperPodx8SchedulerName)
	plg.Name = api.JobID(jobId)
	plg.SchedulerJobAttr = util.SchedulerJobAttr{
		ComJob: util.ComJob{},
		NPUJob: &util.NPUJob{},
	}
	plg.ScheduleEnv = plugin.ScheduleEnv{}
	for _, tt := range getObtainBatchScoreRankTestCases(jobId) {
		t.Run(tt.name, func(t *testing.T) {
			if got := plg.obtainBatchScoreRank(tt.args.task, tt.args.job); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("obtainBatchScoreRank() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectNodeForStandaloneJob(t *testing.T) {
	nodeInfo := buildNodeInfos(nodeInfoIdx0, nodeInfoIdx7, nil)
	t.Run("test selectNodeForStandaloneJob", func(t *testing.T) {
		handler := New(util.SuperPodx8SchedulerName)
		handler.SchedulerJobAttr = util.SchedulerJobAttr{
			NPUJob: &util.NPUJob{
				ReqNPUName: util.NPU910CardName,
				ReqNPUNum:  npuNumber8,
			},
		}
		nodes := handler.selectNodeForStandaloneJob(nodeInfo)
		if len(nodes) != npuNumber8 {
			t.Errorf("got %v; want %v", len(nodes), npuNumber8)
		}
	})
}

func TestScoreNodeForReadyJob(t *testing.T) {
	tasks := getTaskInfos(npuTaskNum2, "job1")
	handler := New(util.SuperPodx8SchedulerName)
	handler.spBlock = 1
	t.Run("test scoreNodeForReadyJob score node success", func(t *testing.T) {
		jobs := make(map[api.JobID]plugin.SchedulerJob)
		job := plugin.SchedulerJob{
			SuperPods: map[string][]plugin.SuperNode{"0": {plugin.SuperNode{Name: "work1"}}},
		}
		jobs[tasks[0].Job] = job
		sMap := map[string]float64{"work1": 0, "work2": 0}
		handler.scoreNodeForReadyJob(tasks[0], &job, sMap)
		if sMap == nil || sMap["work1"] != scoreForNode {
			t.Errorf("got a bad score(%v) of node", sMap)
		}
	})
	t.Run("test scoreNodeForReadyJob score node failed", func(t *testing.T) {
		jobs := make(map[api.JobID]plugin.SchedulerJob)
		job := plugin.SchedulerJob{
			SuperPods: map[string][]plugin.SuperNode{},
		}
		jobs[tasks[1].Job] = job
		sMap := map[string]float64{"work1": 0}
		handler.scoreNodeForReadyJob(tasks[1], &job, sMap)
		if sMap == nil || sMap["work1"] != 0 {
			t.Errorf("got a bad score(%v) of node", sMap)
		}
	})
}
