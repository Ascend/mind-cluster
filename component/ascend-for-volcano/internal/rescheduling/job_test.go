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
	"strconv"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
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
	var tmpPatch *gomonkey.Patches = nil
	faultTask1 := fakeReSchedulerFaultTask(true, []string{"pod0", "vcjob", "node0", "job0", "0"}, 0)
	faultTask2 := fakeReSchedulerFaultTask(false, []string{"pod1", "vcjob", "node1", "job0", "1"}, 0)
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
			schedulerJob: &schedulerJob,
			cacheFuncBefore: func() {
				tmpPatch = gomonkey.ApplyMethod(reflect.TypeOf(&FaultTask{}), "DeleteRealPodByTask",
					func(_ *FaultTask, _ kubernetes.Interface, _ int64) error { return nil })
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
	env := plugin.ScheduleEnv{}
	env.SuperPodInfo = plugin.NewSuperPodInfo()
	tests := buildFaultJobForceDeleteJobTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cacheFuncBefore()
			fJob := &FaultJob{
				ReScheduleKey:      tt.fields.ReScheduleKey,
				IsFaultJob:         tt.fields.IsFaultJob,
				JobName:            tt.fields.JobName,
				JobUID:             tt.fields.JobUID,
				JobNamespace:       tt.fields.JobNamespace,
				FaultTasks:         tt.fields.FaultTasks,
				UpdateTime:         tt.fields.UpdateTime,
				FaultTypes:         tt.fields.FaultTypes,
				DeleteExecutedFlag: tt.fields.DeleteExecutedFlag,
				ElasticScheduling:  tt.fields.ElasticScheduling,
			}
			if err := fJob.ForceDeleteJob(tt.args.schedulerJob, env); (err != nil) != tt.wantErr {
				t.Errorf("ForceDeleteJob() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.args.cacheFuncAfter()
		})
	}
}

// TestGetVirSupPodId test getVirSupPodId
func TestGetVirSupPodId(t *testing.T) {
	env := plugin.ScheduleEnv{}
	env.SuperPodInfo = plugin.NewSuperPodInfo()
	fJob := &FaultJob{
		JobUID: "test",
	}
	t.Run("getVirSupPodId", func(t *testing.T) {
		if result := fJob.getVirSupPodId("node", env); result != "" {
			t.Errorf("return empty when SuperPodReschdInfo "+
				"doesn't have job test result = %v, want %v", result, "")
		}
		env.SuperPodInfo.SuperPodReschdInfo["test"] = map[string][]plugin.SuperNode{
			"0": {
				plugin.SuperNode{
					SuperPodID: 1,
					Name:       "node1",
				},
			},
		}
		if result := fJob.getVirSupPodId("node", env); result != "" {
			t.Errorf("return empty when name != node1 result = %v, want %v", result, "")
		}
		if result := fJob.getVirSupPodId("node1", env); result != "0" {
			t.Errorf("return 0 when name == node1 result = %v, want %v", result, "0")
		}
	})
}

// TestIsContainTask test isContainTask
func TestIsContainTask(t *testing.T) {
	env := plugin.ScheduleEnv{}
	env.SuperPodInfo = plugin.NewSuperPodInfo()

	fJob := &FaultJob{
		JobUID: "test",
	}
	t.Run("isContainTask", func(t *testing.T) {
		if result := fJob.isContainTask([]string{"0"}, "node", env); result != false {
			t.Errorf("return false when SuperPodReschdInfo doesn't have job test "+
				"result = %v, want %v", result, false)
		}
		env.SuperPodInfo.SuperPodReschdInfo["test"] = map[string][]plugin.SuperNode{
			"0": {
				plugin.SuperNode{
					SuperPodID: 1,
					Name:       "node1",
				},
			},
		}
		if result := fJob.isContainTask([]string{"0"}, "node", env); result != false {
			t.Errorf("return false when name != node1 result = %v, want %v", result, false)
		}
		if result := fJob.isContainTask([]string{"0"}, "node1", env); result != true {
			t.Errorf("return true when name == node1 result = %v, want %v", result, true)
		}
	})
}

func TestGetJobFaultRescheduleLabel(t *testing.T) {
	t.Run("01-GetJobFaultRescheduleLabel return error when job is nil", func(t *testing.T) {
		fJob := &FaultJob{JobUID: "test"}
		res := fJob.GetJobFaultRescheduleLabel()
		if res != JobOffRescheduleLabelValue {
			t.Errorf("GetJobFaultRescheduleLabel() res = %v, wantRes is %v", res, JobOffRescheduleLabelValue)
		}
	})
}

type isNormalJobNeedRestartTestCase struct {
	name    string
	fJob    *FaultJob
	wantRes bool
}

func buildIsNormalJobNeedRestartTestCases() []isNormalJobNeedRestartTestCase {
	return []isNormalJobNeedRestartTestCase{
		{
			name:    "01-IsNormalJobNeedRestart return false when fJob is nil",
			fJob:    nil,
			wantRes: false,
		},
		{
			name:    "02-IsNormalJobNeedRestart return true when IsSoftwareFault is true",
			fJob:    &FaultJob{FaultTasks: []FaultTask{{IsSoftwareFault: true}}},
			wantRes: true,
		},
		{
			name: "03-IsNormalJobNeedRestart return true when FaultHandling is PreSeparateNPU",
			fJob: &FaultJob{FaultTasks: []FaultTask{{
				Reason: []FaultReasonList{
					{FaultDeviceList: FaultDeviceList{FaultHandling: PreSeparateNPU}},
				},
			}}},
			wantRes: true,
		},
	}
}

func TestIsNormalJobNeedRestart(t *testing.T) {
	testCases := buildIsNormalJobNeedRestartTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if res := tt.fJob.IsNormalJobNeedRestart(); res != tt.wantRes {
				t.Errorf("IsNormalJobNeedRestart() res = %v, wantRes is %v", res, tt.wantRes)
			}
		})
	}
}

const (
	mockJobName   = "job0"
	mockTaskUID   = "task01"
	mockCardName1 = "npu1"
	mockCardName2 = "npu2"
	mockJobUID    = "vcjob/job0"
	mockNumFive   = 5
)

func TestIsJobGraceDeleteSuccess(t *testing.T) {
	fJob := &FaultJob{FaultTasks: []FaultTask{{
		IsFaultTask: true,
		TaskUID:     mockTaskUID,
		UseCardName: []string{mockCardName1, mockCardName2},
	}}}
	jobInfo := test.FakeNormalTestJob(mockJobName, util.NPUIndex2)
	t.Run("01-isJobGraceDeleteSuccess return true when jobInfo.Tasks is nil", func(t *testing.T) {
		if res := fJob.isJobGraceDeleteSuccess(jobInfo, false); res != true {
			t.Errorf("isJobGraceDeleteSuccess() res = %v, want true", res)
		}
	})
	t.Run("02-isJobGraceDeleteSuccess return true when jobInfo.PodGroup.Labels is not nil",
		func(t *testing.T) {
			fJob.PendingSessionNum = spPendingTimes
			jobInfo.Tasks = map[api.TaskID]*api.TaskInfo{
				mockTaskUID: {Pod: &v1.Pod{}},
			}
			jobInfo.PodGroup.Labels = map[string]string{
				util.SinglePodTag: util.EnableFunc,
			}
			if res := fJob.isJobGraceDeleteSuccess(jobInfo, false); res != true {
				t.Errorf("isJobGraceDeleteSuccess() res = %v, want true", res)
			}
		})
}

func TestDeleteJobWithLabels(t *testing.T) {
	ssn := test.FakeNormalSSN(nil)
	t.Run("01-deleteJobWithLabels return error when isFaultJobCanRestarted failed", func(t *testing.T) {
		fJob := &FaultJob{JobUID: mockJobUID}
		err := fJob.deleteJobWithLabels(ssn, &ReScheduler{}, &plugin.SchedulerJob{}, plugin.ScheduleEnv{})
		if err == nil {
			t.Errorf("deleteJobWithLabels() err = %v, wantErr is nil", err)
		}
	})
	t.Run("02-deleteJobWithLabels return error when deleteJobWithSubHealthyLabels success", func(t *testing.T) {
		fJob := &FaultJob{JobUID: mockJobUID, SubHealthyStrategy: util.SubHealthyForceExit, IsSubHealthFault: true,
			IsFaultJob: true, faultReason: PodHealthy}
		env := plugin.ScheduleEnv{
			FrameAttr: plugin.VolcanoFrame{
				KubeClient: fake.NewSimpleClientset(),
			},
			ClusterCache: plugin.ClusterCache{
				SuperPodInfo: &plugin.SuperPodInfo{
					SuperPodFaultTaskNodes: map[api.JobID][]string{mockJobUID: {}},
				},
			},
		}
		if err := fJob.deleteJobWithLabels(ssn, &ReScheduler{}, &plugin.SchedulerJob{}, env); err != nil {
			t.Errorf("deleteJobWithLabels() err = %v, wantErr is nil", err)
		}
	})
}

type isFaultJobCanRestartedTestCase struct {
	name       string
	fJob       *FaultJob
	reschedule *ReScheduler
	wantRes    bool
}

func buildIsFaultJobCanRestartedTestCases() []isFaultJobCanRestartedTestCase {
	return []isFaultJobCanRestartedTestCase{
		{
			name:       "01-isFaultJobCanRestarted returns false when fJob.IsFaultJob is false",
			fJob:       &FaultJob{JobUID: mockJobUID},
			reschedule: &ReScheduler{},
			wantRes:    false,
		},
		{
			name:       "02-isFaultJobCanRestarted returns true when fJob.faultReason is PodHealthy",
			fJob:       &FaultJob{JobUID: mockJobUID, IsFaultJob: true, faultReason: PodHealthy},
			reschedule: &ReScheduler{},
			wantRes:    true,
		},
		{
			name:       "03-isFaultJobCanRestarted returns false when fJob.faultReason is PodFailed",
			fJob:       &FaultJob{JobUID: mockJobUID, IsFaultJob: true, faultReason: PodFailed},
			reschedule: &ReScheduler{},
			wantRes:    false,
		},
		{
			name: "04-isFaultJobCanRestarted returns false when JobRemainRetryTimes not found key",
			fJob: &FaultJob{JobUID: mockJobUID, IsFaultJob: true, faultReason: PodFailed, FaultRetryTimes: 1},
			reschedule: &ReScheduler{
				DealReSchedulerCache: &DealReSchedulerCache{
					JobRemainRetryTimes: map[api.JobID]*RemainRetryTimes{mockJobUID: nil},
				},
			},
			wantRes: false,
		},
		{
			name: "05-isFaultJobCanRestarted returns false when remain.Times is 0",
			fJob: &FaultJob{JobUID: mockJobUID, IsFaultJob: true, faultReason: PodFailed, FaultRetryTimes: 1},
			reschedule: &ReScheduler{
				DealReSchedulerCache: &DealReSchedulerCache{
					JobRemainRetryTimes: map[api.JobID]*RemainRetryTimes{
						mockJobUID: {UUID: mockJobUID, Times: 0},
					},
				},
			},
			wantRes: false,
		},
	}
}

func TestIsFaultJobCanRestarted(t *testing.T) {
	testCases := buildIsFaultJobCanRestartedTestCases()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.fJob.isFaultJobCanRestarted(tc.reschedule)
			if res != tc.wantRes {
				t.Errorf("isFaultJobCanRestarted() res = %v, wantRes %v", res, tc.wantRes)
			}
		})
	}
}

func TestDeleteJobWithSubHealthyLabels(t *testing.T) {
	fJob := &FaultJob{JobUID: mockJobUID}
	t.Run("01-deleteJobWithSubHealthyLabels return nil when fJob.IsFaultJob is false", func(t *testing.T) {
		ssn := test.FakeNormalSSN(nil)
		env := plugin.ScheduleEnv{
			FrameAttr: plugin.VolcanoFrame{
				KubeClient: fake.NewSimpleClientset(),
			},
			ClusterCache: plugin.ClusterCache{
				SuperPodInfo: &plugin.SuperPodInfo{
					SuperPodFaultTaskNodes: map[api.JobID][]string{mockJobUID: {}},
				},
			},
		}
		err := fJob.deleteJobWithSubHealthyLabels(ssn, &plugin.SchedulerJob{}, env)
		if err != nil {
			t.Errorf("deleteJobWithSubHealthyLabels() err = %v, wantErr is nil", err)
		}
	})
	t.Run("02-deleteJobWithSubHealthyLabels return error when ssn is nil", func(t *testing.T) {
		fJob.SubHealthyStrategy = util.SubHealthyForceExit
		err := fJob.deleteJobWithSubHealthyLabels(nil, nil, plugin.ScheduleEnv{})
		if err == nil {
			t.Errorf("deleteJobWithSubHealthyLabels() err = %v, wantErr is not nil", err)
		}
	})
}

func TestDeleteJobWithFaultLabels(t *testing.T) {
	t.Run("01-deleteJobWithFaultLabels return error when ssn is nil", func(t *testing.T) {
		fJob := &FaultJob{JobUID: mockJobUID, SubHealthyStrategy: util.SubHealthyForceExit}
		err := fJob.deleteJobWithFaultLabels(nil, nil, nil, plugin.ScheduleEnv{})
		if err == nil {
			t.Errorf("deleteJobWithFaultLabels() err = %v, wantErr is not nil", err)
		}
	})
	t.Run("02-deleteJobWithFaultLabels return error when ssn is nil", func(t *testing.T) {
		fJob := &FaultJob{JobUID: mockJobUID, ReScheduleKey: JobForceRescheduleLabelValue}
		err := fJob.deleteJobWithFaultLabels(nil, nil, nil, plugin.ScheduleEnv{})
		if err == nil {
			t.Errorf("deleteJobWithFaultLabels() err = %v, wantErr is not nil", err)
		}
	})
}

func TestIsJobSingleRescheduling(t *testing.T) {
	t.Run("01-IsJobSingleRescheduling return true when pod-rescheduling label is on", func(t *testing.T) {
		fJob := &FaultJob{JobUID: mockJobUID, SubHealthyStrategy: util.SubHealthyForceExit}
		sJob := &plugin.SchedulerJob{}
		sJob.Label = map[string]string{util.SinglePodTag: util.EnableFunc}
		if res := fJob.IsJobSingleRescheduling(sJob); !res {
			t.Errorf("IsJobSingleRescheduling() res = %v, wantRes is true", res)
		}
	})
}

func TestIsProcessReschedulingJob(t *testing.T) {
	t.Run("01-IsProcessReschedulingJob return true when process-recover-enable label is on", func(t *testing.T) {
		fJob := &FaultJob{JobUID: mockJobUID, SubHealthyStrategy: util.SubHealthyForceExit}
		sJob := &plugin.SchedulerJob{}
		sJob.Label = map[string]string{util.ProcessRecoverEnable: util.EnableFunc}
		if res := fJob.IsProcessReschedulingJob(sJob); !res {
			t.Errorf("IsProcessReschedulingJob() res = %v, wantRes is true", res)
		}
	})
	t.Run("02-IsProcessReschedulingJob return false when rocess-recover-enable label is not on", func(t *testing.T) {
		fJob := &FaultJob{JobUID: mockJobUID, SubHealthyStrategy: util.SubHealthyForceExit}
		sJob := &plugin.SchedulerJob{}
		sJob.Label = map[string]string{util.ProcessRecoverEnable: ""}
		if res := fJob.IsProcessReschedulingJob(sJob); res {
			t.Errorf("IsProcessReschedulingJob() res = %v, wantRes is false", res)
		}
	})
}

type isNormalTaskCanBeDeleteArgs struct {
	faultJob      *FaultJob
	deletePodInfo *deletePodInfo
	schedulerJob  *plugin.SchedulerJob
	env           plugin.ScheduleEnv
}

func initIsNormalTaskCanBeDeleteArgs() isNormalTaskCanBeDeleteArgs {
	fJob := &FaultJob{
		JobUID:             mockJobUID,
		SubHealthyStrategy: util.SubHealthyForceExit,
		PendingSessionNum:  spPendingTimes,
	}
	dpi := &deletePodInfo{
		isMasterFault: false,
		isSuperPod:    false,
		ids:           nil,
		reason:        "",
	}
	schedulerJob := &plugin.SchedulerJob{}
	schedulerJob.Label = map[string]string{util.SinglePodTag: util.EnableFunc}
	env := plugin.ScheduleEnv{}
	env.SuperPodInfo = plugin.NewSuperPodInfo()
	return isNormalTaskCanBeDeleteArgs{
		faultJob:      fJob,
		deletePodInfo: dpi,
		schedulerJob:  schedulerJob,
		env:           env,
	}
}
func TestIsNormalTaskCanBeDelete(t *testing.T) {
	args := initIsNormalTaskCanBeDeleteArgs()
	t.Run("01-isNormalTaskCanBeDelete return false when fTask.IsFaultTask and dpi.superPod are false",
		func(t *testing.T) {
			res := args.faultJob.isNormalTaskCanBeDelete(FaultTask{}, args.schedulerJob,
				plugin.ScheduleEnv{}, args.deletePodInfo)
			if res {
				t.Errorf("isNormalTaskCanBeDelete() res = %v, wantRes is false", res)
			}
		})
	args.deletePodInfo.isSuperPod = true
	t.Run("02-isNormalTaskCanBeDelete return false when fJob.PendingSessionNum less than 6",
		func(t *testing.T) {
			args.faultJob.PendingSessionNum = mockNumFive
			res := args.faultJob.isNormalTaskCanBeDelete(FaultTask{}, args.schedulerJob,
				plugin.ScheduleEnv{}, args.deletePodInfo)
			if res {
				t.Errorf("isNormalTaskCanBeDelete() res = %v, wantRes is false", res)
			}
		})
	args.env.SuperPodInfo.SuperPodReschdInfo["test"] = map[string][]plugin.SuperNode{
		"0": {plugin.SuperNode{SuperPodID: 1, Name: "node1"}},
	}
	t.Run("03-isNormalTaskCanBeDelete return false when isContainTask function return false",
		func(t *testing.T) {
			res := args.faultJob.isNormalTaskCanBeDelete(FaultTask{}, args.schedulerJob, args.env, args.deletePodInfo)
			if res {
				t.Errorf("isNormalTaskCanBeDelete() res = %v, wantRes is false", res)
			}
		})

	t.Run("04-isNormalTaskCanBeDelete return false when IsProcessReschedulingJob function return true "+
		"and fTask.IsFaultTask is false",
		func(t *testing.T) {
			args.schedulerJob.Label = map[string]string{util.ProcessRecoverEnable: util.EnableFunc}
			res := args.faultJob.isNormalTaskCanBeDelete(FaultTask{}, args.schedulerJob, args.env, args.deletePodInfo)
			if res {
				t.Errorf("isNormalTaskCanBeDelete() res = %v, wantRes is false", res)
			}
		})

	t.Run("05-isNormalTaskCanBeDelete return true when IsProcessReschedulingJob function return false ",
		func(t *testing.T) {
			args.schedulerJob.Label = map[string]string{}
			res := args.faultJob.isNormalTaskCanBeDelete(FaultTask{}, args.schedulerJob, args.env, args.deletePodInfo)
			if !res {
				t.Errorf("isNormalTaskCanBeDelete() res = %v, wantRes is true", res)
			}
		})
}

func TestGetTaskPodUidByTaskName(t *testing.T) {
	jobInfo := test.FakeNormalTestJob(mockJobName, util.NPUIndex2)
	t.Run("01-getTaskPodUidByTaskName return empty string when taskName is empty", func(t *testing.T) {
		if res := getTaskPodUidByTaskName("", jobInfo); res != "" {
			t.Errorf("getTaskPodUidByTaskName() res = %v, wantRes is empty string", res)
		}
	})
	t.Run("02-getTaskPodUidByTaskName return non-empty string when taskName is not empty", func(t *testing.T) {
		if res := getTaskPodUidByTaskName("pod0", jobInfo); res == "" {
			t.Errorf("getTaskPodUidByTaskName() res = %v, wantRes is non-empty string", res)
		}
	})
}

func TestGraceDeleteJob(t *testing.T) {
	fJob := &FaultJob{JobUID: mockJobUID}
	env := plugin.ScheduleEnv{}
	env.SuperPodInfo = plugin.NewSuperPodInfo()
	npuJob := &plugin.SchedulerJob{}
	t.Run("01-GraceDeleteJob return error when ssn is nil", func(t *testing.T) {
		err := fJob.GraceDeleteJob(nil, &plugin.SchedulerJob{}, env)
		if err == nil {
			t.Errorf("GraceDeleteJob() err = %v, wantErr is not nil", err)
		}
	})
	t.Run("02-GraceDeleteJob return error when npuJob is nil", func(t *testing.T) {
		err := fJob.GraceDeleteJob(&framework.Session{}, nil, env)
		if err == nil {
			t.Errorf("GraceDeleteJob() err = %v, wantErr is not nil", err)
		}
	})
	t.Run("03-GraceDeleteJob return nil when ssn and npuJob are not nil", func(t *testing.T) {
		npuJob.Annotation = map[string]string{util.SuperPodAnnoKey: ""}
		err := fJob.GraceDeleteJob(&framework.Session{}, npuJob, env)
		if err != nil {
			t.Errorf("GraceDeleteJob() err = %v, wantErr is nil", err)
		}
	})
}

func mockFaultJobWithTasks() *FaultJob {
	return &FaultJob{
		JobUID: mockJobUID,
		FaultTasks: []FaultTask{
			{
				IsFaultTask: true,
				TaskUID:     mockTaskUID,
				UseCardName: []string{mockCardName1, mockCardName2},
			},
		}}
}

func TestGraceDeletePods(t *testing.T) {
	fJob := mockFaultJobWithTasks()
	ssn := test.FakeNormalSSN(nil)
	npuJob := &plugin.SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{
			NPUJob: &util.NPUJob{
				Tasks: map[api.TaskID]util.NPUTask{},
			},
		},
	}
	env := plugin.ScheduleEnv{}
	env.SuperPodInfo = plugin.NewSuperPodInfo()
	t.Run("01-graceDeletePods do not change IsBeingGracefulDeleted when npuTask not in session",
		func(t *testing.T) {
			var tmpPatch *gomonkey.Patches = nil
			tmpPatch = gomonkey.ApplyPrivateMethod(reflect.TypeOf(&util.NPUTask{}), "ForceDeletePodByTaskInf",
				func(ssn *framework.Session, reason string, nodeName string) error { return nil })
			fJob.graceDeletePods(ssn, npuJob, env, &deletePodInfo{})
			tmpPatch.Reset()
			for id := range fJob.FaultTasks {
				if fJob.FaultTasks[id].IsBeingGracefulDeleted == true {
					t.Error("graceDeletePods() return true, want false")
				}
			}
		})
	t.Run("02-graceDeletePods change IsBeingGracefulDeleted when npuTask in session", func(t *testing.T) {
		npuJob.Label = map[string]string{}
		npuJob.Tasks = map[api.TaskID]util.NPUTask{mockTaskUID: {
			VTask: &util.VTask{Allocated: util.TaskAllocated{}}}}
		var tmpPatch *gomonkey.Patches = nil
		tmpPatch = gomonkey.ApplyPrivateMethod(reflect.TypeOf(&util.NPUTask{}), "ForceDeletePodByTaskInf",
			func(ssn *framework.Session, reason string, nodeName string) error { return nil })
		fJob.graceDeletePods(ssn, npuJob, env, &deletePodInfo{})
		tmpPatch.Reset()
		for id := range fJob.FaultTasks {
			if fJob.FaultTasks[id].IsBeingGracefulDeleted == false {
				t.Error("graceDeletePods() return false, want true")
			}
		}
	})
}

func TestRestartSingleFaultJob(t *testing.T) {
	fJob := &FaultJob{JobUID: mockJobUID}
	t.Run("01-restartSingleFaultJob return error when fJob.ReScheduleKey is off",
		func(t *testing.T) {
			fJob.ReScheduleKey = JobOffRescheduleLabelValue
			err := fJob.restartSingleFaultJob(nil, nil, nil, plugin.ScheduleEnv{})
			if err == nil {
				t.Errorf("restartSingleFaultJob() err = %v, wantErr is not nil", err)
			}
		})
	t.Run("02-restartSingleFaultJob return error when fJob.ReScheduleKey is fault-scheduling",
		func(t *testing.T) {
			fJob.ReScheduleKey = JobRescheduleLabelKey
			err := fJob.restartSingleFaultJob(nil, nil, nil, plugin.ScheduleEnv{})
			if err == nil {
				t.Errorf("restartSingleFaultJob() err = %v, wantErr is not nil", err)
			}
		})
}

func TestJobInfoInSession(t *testing.T) {
	fJob := &FaultJob{JobUID: mockJobUID, ElasticScheduling: JobOnElasticScheduling}
	jobs := map[api.JobID]*api.JobInfo{
		mockJobUID: {
			UID: mockJobUID,
		},
	}
	t.Run("01-jobInfoInSession return nil when fJob.ElasticScheduling is on and jobs is nil",
		func(t *testing.T) {
			if res := fJob.jobInfoInSession(nil); res != nil {
				t.Errorf("jobInfoInSession() res = %v, wantRes is nil", res)
			}
		})
	t.Run("02-jobInfoInSession return nil when fJob.ElasticScheduling is on and jobs is not nil",
		func(t *testing.T) {
			if res := fJob.jobInfoInSession(jobs); res == nil {
				t.Errorf("jobInfoInSession() res = %v, wantRes is not nil", res)
			}
		})
	fJob.ElasticScheduling = JobOffElasticScheduling
	t.Run("03-jobInfoInSession return nil when fJob.ElasticScheduling is off and jobs is nil",
		func(t *testing.T) {

			if res := fJob.jobInfoInSession(nil); res != nil {
				t.Errorf("jobInfoInSession() res = %v, wantRes is nil", res)
			}
		})
	t.Run("04-jobInfoInSession return nil when fJob.ElasticScheduling is off and jobs is not nil",
		func(t *testing.T) {
			if res := fJob.jobInfoInSession(jobs); res == nil {
				t.Errorf("jobInfoInSession() res = %v, wantRes is not nil", res)
			}
		})
}

func TestDeletingTasksConcurrently(t *testing.T) {
	waitDeleteTaskCountList := []int{201, 256, 1000, 1280, 1999, 4000}
	for _, waitDeleteTaskCount := range waitDeleteTaskCountList {
		t.Run("TestDeletingTasksConcurrently taskCount="+strconv.Itoa(waitDeleteTaskCount),
			func(t *testing.T) {
				faultJob := &FaultJob{}
				kubeClient := fake.NewSimpleClientset()

				actualDeleteTaskNameMap := map[string]string{}
				listMutex := sync.Mutex{}

				mockFunc := gomonkey.ApplyPrivateMethod(faultJob, "forceDeleteTasksConcurrently",
					func(_ *FaultJob, waitDeleteTask []FaultTask, _ kubernetes.Interface,
						deleteJobSync *sync.WaitGroup) {
						listMutex.Lock()
						defer listMutex.Unlock()

						for _, faultTask := range waitDeleteTask {
							actualDeleteTaskNameMap[faultTask.TaskName] = faultTask.TaskName
						}
						deleteJobSync.Done()
					})
				defer mockFunc.Reset()
				var waitDeleteTaskList []FaultTask
				for i := 0; i < waitDeleteTaskCount; i++ {
					waitDeleteTaskList = append(waitDeleteTaskList, FaultTask{TaskName: strconv.Itoa(i)})
				}
				faultJob.deletingTasksConcurrently(waitDeleteTaskList, kubeClient)

				if len(actualDeleteTaskNameMap) != len(waitDeleteTaskList) {
					t.Errorf("actualDeleteTaskNameMap size(%v) != waitDeleteTaskList size(%v)",
						len(actualDeleteTaskNameMap), len(waitDeleteTaskList))
				}
				for _, taskName := range actualDeleteTaskNameMap {
					found := false
					for _, faultTask := range waitDeleteTaskList {
						if taskName == faultTask.TaskName {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("actualDeleteTaskNameMap elements != waitDeleteTaskList elements")
					}
				}
			})
	}
}

func TestIsFailedTask(t *testing.T) {
	const exitCode127 = 127
	terminatedErrorState := v1.ContainerState{
		Terminated: &v1.ContainerStateTerminated{ExitCode: exitCode127},
	}
	terminatedNormalState := v1.ContainerState{
		Terminated: &v1.ContainerStateTerminated{ExitCode: 0},
	}
	normalStatus := v1.ContainerStatus{
		State:                terminatedNormalState,
		LastTerminationState: v1.ContainerState{},
		RestartCount:         0,
	}
	backoffStatus := v1.ContainerStatus{
		State:                v1.ContainerState{},
		LastTerminationState: terminatedErrorState,
		RestartCount:         0,
	}
	task := test.FakeNormalTestTask("pod1", "node1", "pg1")
	t.Run("01-isFailedTask return false when task status normal", func(t *testing.T) {
		task.Pod.Status = v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{normalStatus}}
		if isFailedTask(task) {
			t.Errorf("isFailedTask() error, want false, return true")
		}
	})
	t.Run("02-isFailedTask return true when task container terminated with none-zero", func(t *testing.T) {
		task.Pod.Status = v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{backoffStatus}}
		if !isFailedTask(task) {
			t.Errorf("isFailedTask() error, want true, return false")
		}
	})
	t.Run("03-isFailedTask return true when task phase failed", func(t *testing.T) {
		task.Pod.Status = v1.PodStatus{Phase: v1.PodFailed}
		if !isFailedTask(task) {
			t.Errorf("isFailedTask() error, want true, return false")
		}
	})
	t.Run("04-isFailedTask return false when task or pod are nil", func(t *testing.T) {
		if isFailedTask(nil) || isFailedTask(&api.TaskInfo{}) {
			t.Errorf("isFailedTask() error, want false, return true")
		}
	})
}

// TestSetFaultRetryTimeOfJob tests the setFaultRetryTimeOfJob function
func TestSetFaultRetryTimeOfJob(t *testing.T) {
	// Test case 1: Job without fault-retry-times label
	t.Run("01-setFaultRetryTimeOfJob with no fault-retry-times label", func(t *testing.T) {
		fJob := &FaultJob{
			JobUID: "test-job-1",
			Labels: map[string]string{},
		}
		fJob.setFaultRetryTimeOfJob()
		if fJob.FaultRetryTimes != 0 {
			t.Errorf("setFaultRetryTimeOfJob() FaultRetryTimes = %v, want 0", fJob.FaultRetryTimes)
		}
	})

	// Test case 2: Job with invalid fault-retry-times label
	t.Run("02-setFaultRetryTimeOfJob with invalid fault-retry-times label", func(t *testing.T) {
		fJob := &FaultJob{
			JobUID: "test-job-2",
			Labels: map[string]string{
				FaultRetryTimesKey: "invalid-value",
			},
		}
		fJob.setFaultRetryTimeOfJob()
		if fJob.FaultRetryTimes != 0 {
			t.Errorf("setFaultRetryTimeOfJob() FaultRetryTimes = %v, want 0", fJob.FaultRetryTimes)
		}
	})

	// Test case 3: Job with valid fault-retry-times label
	t.Run("03-setFaultRetryTimeOfJob with valid fault-retry-times label", func(t *testing.T) {
		expectedRetryTimes := 3
		fJob := &FaultJob{
			JobUID: "test-job-3",
			Labels: map[string]string{
				FaultRetryTimesKey: strconv.Itoa(expectedRetryTimes),
			},
		}
		fJob.setFaultRetryTimeOfJob()
		if fJob.FaultRetryTimes != expectedRetryTimes {
			t.Errorf("setFaultRetryTimeOfJob() FaultRetryTimes = %v, want %v", fJob.FaultRetryTimes, expectedRetryTimes)
		}
	})

	// Test case 4: Job with zero fault-retry-times label
	t.Run("04-setFaultRetryTimeOfJob with zero fault-retry-times label", func(t *testing.T) {
		expectedRetryTimes := 0
		fJob := &FaultJob{
			JobUID: "test-job-4",
			Labels: map[string]string{
				FaultRetryTimesKey: strconv.Itoa(expectedRetryTimes),
			},
		}
		fJob.setFaultRetryTimeOfJob()
		if fJob.FaultRetryTimes != expectedRetryTimes {
			t.Errorf("setFaultRetryTimeOfJob() FaultRetryTimes = %v, want %v", fJob.FaultRetryTimes, expectedRetryTimes)
		}
	})
}
