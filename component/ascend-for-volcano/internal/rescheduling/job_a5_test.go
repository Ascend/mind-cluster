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

/*
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package rescheduling

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/client-go/kubernetes"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

func Test_is910A5Job(t *testing.T) {
	tests := []struct {
		name string
		job  *plugin.SchedulerJob
		want bool
	}{
		{
			name: "nil job",
			job:  nil,
			want: false,
		},
		{
			name: "selector has key but not A5",
			job: &plugin.SchedulerJob{
				SchedulerJobAttr: util.SchedulerJobAttr{
					ComJob: util.ComJob{Annotation: map[string]string{
						util.SchedulePolicyAnnoKey: util.Chip1Node2,
					}}}},
			want: false,
		},
		{
			name: "selector has A5",
			job: &plugin.SchedulerJob{
				SchedulerJobAttr: util.SchedulerJobAttr{
					ComJob: util.ComJob{Annotation: map[string]string{
						util.SchedulePolicyAnnoKey: util.Chip8Node8Ra64Sp,
					}}}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := is910A5SuperPodJob(tt.job); got != tt.want {
				t.Errorf("is910A5SuperPodJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

const (
	testTp8        = 8
	testTp4        = 4
	testTp2        = 2
	testTp1        = 1
	nodeNum        = 8
	testCreateTime = 0
)

type inTheSameTpBlockTestCase struct {
	fields  FaultJobTestField
	name    string
	tpBlock int
	wantErr [nodeNum]bool
}

func buildInTheSameTpBlockTestCases1() []inTheSameTpBlockTestCase {
	faultTask1 := fakeReSchedulerFaultTask(true, []string{"pod0", "vcjob", "node0", "job0", "0"}, testCreateTime)
	faultTask2 := fakeReSchedulerFaultTask(false, []string{"pod1", "vcjob", "node1", "job0", "1"}, testCreateTime)
	test1 := inTheSameTpBlockTestCase{
		name: "01-inTheSameTpBlock() return true when in same tp-block=16",
		fields: FaultJobTestField{
			JobName:      "job0",
			JobUID:       "vcjob/job0",
			JobNamespace: "vcjob",
			FaultTasks:   []FaultTask{faultTask1, faultTask2},
		},
		tpBlock: testTp2,
		wantErr: [nodeNum]bool{true, true, false, false, false, false, false, false},
	}
	test2 := inTheSameTpBlockTestCase{
		name: "02-inTheSameTpBlock() return true when in same tp-block=64",
		fields: FaultJobTestField{
			JobName:      "job0",
			JobUID:       "vcjob/job0",
			JobNamespace: "vcjob",
			FaultTasks:   []FaultTask{faultTask1, faultTask2},
		},
		tpBlock: testTp8,
		wantErr: [nodeNum]bool{true, true, true, true, true, true, true, true},
	}
	return []inTheSameTpBlockTestCase{
		test1, test2,
	}
}

func buildInTheSameTpBlockTestCases2() []inTheSameTpBlockTestCase {
	faultTask1 := fakeReSchedulerFaultTask(false, []string{"pod0", "vcjob", "node0", "job0", "0"}, testCreateTime)
	faultTask2 := fakeReSchedulerFaultTask(true, []string{"pod1", "vcjob", "node1", "job0", "7"}, testCreateTime)
	test1 := inTheSameTpBlockTestCase{
		name: "03-inTheSameTpBlock() return true when in same tp-block=16",
		fields: FaultJobTestField{
			JobName:      "job0",
			JobUID:       "vcjob/job0",
			JobNamespace: "vcjob",
			FaultTasks:   []FaultTask{faultTask1, faultTask2},
		},
		tpBlock: testTp2,
		wantErr: [nodeNum]bool{false, false, false, false, false, false, true, true},
	}
	test2 := inTheSameTpBlockTestCase{
		name: "04-inTheSameTpBlock() return true when in same tp-block=32",
		fields: FaultJobTestField{
			JobName:      "job0",
			JobUID:       "vcjob/job0",
			JobNamespace: "vcjob",
			FaultTasks:   []FaultTask{faultTask1, faultTask2},
		},
		tpBlock: testTp4,
		wantErr: [nodeNum]bool{false, false, false, false, true, true, true, true},
	}
	return []inTheSameTpBlockTestCase{test1, test2}
}

func buildInTheSameTpBlockTestCases3() []inTheSameTpBlockTestCase {
	faultTask1 := fakeReSchedulerFaultTask(true, []string{"pod0", "vcjob", "node0", "job0", "0"}, testCreateTime)
	faultTask2 := fakeReSchedulerFaultTask(true, []string{"pod1", "vcjob", "node1", "job0", "7"}, testCreateTime)
	test1 := inTheSameTpBlockTestCase{
		name: "05-inTheSameTpBlock() return false when in same tp-block=8 and two fault tasks",
		fields: FaultJobTestField{
			JobName:      "job0",
			JobUID:       "vcjob/job0",
			JobNamespace: "vcjob",
			FaultTasks:   []FaultTask{faultTask1, faultTask2},
		},
		tpBlock: testTp1,
		wantErr: [nodeNum]bool{false, false, false, false, false, false, false, false},
	}
	test2 := inTheSameTpBlockTestCase{
		name: "06-inTheSameTpBlock() return true when in same tp-block=16 and two fault tasks",
		fields: FaultJobTestField{
			JobName:      "job0",
			JobUID:       "vcjob/job0",
			JobNamespace: "vcjob",
			FaultTasks:   []FaultTask{faultTask1, faultTask2},
		},
		tpBlock: testTp2,
		wantErr: [nodeNum]bool{true, true, false, false, false, false, true, true},
	}
	return []inTheSameTpBlockTestCase{test1, test2}
}

func buildInTheSameTpBlockTestCases() []inTheSameTpBlockTestCase {
	result := make([]inTheSameTpBlockTestCase, 0)
	result = append(result, buildInTheSameTpBlockTestCases1()...)
	result = append(result, buildInTheSameTpBlockTestCases2()...)
	result = append(result, buildInTheSameTpBlockTestCases3()...)
	return result
}

func testCaseRunDetail(t *testing.T, tc inTheSameTpBlockTestCase) {
	fJob := &FaultJob{
		ReScheduleKey: tc.fields.ReScheduleKey,
		IsFaultJob:    tc.fields.IsFaultJob,
		JobName:       tc.fields.JobName,
		JobUID:        tc.fields.JobUID,
		JobNamespace:  tc.fields.JobNamespace,
		FaultTasks:    tc.fields.FaultTasks,
		FaultJobA5Field: FaultJobA5Field{
			TpBlock: tc.tpBlock,
		},
	}
	for i := 0; i < nodeNum; i++ {
		if ret := fJob.inTheSameTpBlock(
			FaultTask{IsFaultTask: false, NodeRankIndex: strconv.Itoa(i)}); ret != tc.wantErr[i] {
			t.Errorf("inTheSameTpBlock() when nodeRank=%d, return = %v, but want %v", i, ret, tc.wantErr[i])
		}
	}
}

// TestInTheSameTpBlock test for the same tp-block by rankIndex
func TestInTheSameTpBlock(t *testing.T) {
	tests := buildInTheSameTpBlockTestCases()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testCaseRunDetail(t, tc)
		})
	}
}

type inTheSameVSuperPodTestCase struct {
	name      string
	TpBlock   int
	SuperPods map[string][]plugin.SuperNode
	ids       []string
	nodeName  string
	want      bool
}

func buildTestCase0() inTheSameVSuperPodTestCase {
	return inTheSameVSuperPodTestCase{
		name:    "01-test ids don't exist",
		TpBlock: testTp1,
		SuperPods: map[string][]plugin.SuperNode{
			"0": {
				{Name: "work1"},
				{Name: "work2"},
				{Name: "work3"},
				{Name: "work4"},
				{Name: "work5"},
			},
		},
		ids:      []string{"1"},
		nodeName: "work1",
		want:     false,
	}
}

func buildTestCase1() inTheSameVSuperPodTestCase {
	return inTheSameVSuperPodTestCase{
		name:    "02-test nodeName don't exist",
		TpBlock: testTp1,
		SuperPods: map[string][]plugin.SuperNode{
			"0": {
				{Name: "work1"},
				{Name: "work2"},
				{Name: "work3"},
				{Name: "work4"},
				{Name: "work5"},
			},
			"1": {
				{Name: "work1"},
				{Name: "work2"},
				{Name: "work3"},
				{Name: "work4"},
				{Name: "work5"},
			},
		},
		ids:      []string{"0"},
		nodeName: "work9",
		want:     false,
	}
}

func buildInTheSameVSuperPodTestCases() []inTheSameVSuperPodTestCase {
	return []inTheSameVSuperPodTestCase{
		buildTestCase0(),
		buildTestCase1(),
	}
}

func TestInTheSameVSuperPod(t *testing.T) {
	for _, tc := range buildInTheSameVSuperPodTestCases() {
		fJob := &FaultJob{
			FaultJobA5Field: FaultJobA5Field{
				TpBlock: tc.TpBlock},
		}
		t.Run(tc.name, func(t *testing.T) {
			if ret := fJob.inTheSameVSuperPod(tc.ids, tc.nodeName); ret != tc.want {
				t.Errorf("inTheSameVSuperPod() when ids=%v nodeName=%v, return = %v, but want %v",
					tc.ids, tc.nodeName, ret, tc.want)
			}
		})
	}
}

func checkVSuperPodIds(t *testing.T, got, want []string) {
	if len(got) != len(want) {
		t.Errorf("getVSuperPodIds() = %v, want %v", got, want)
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("getVSuperPodIds() = %v, want %v", got, want)
			return
		}
	}
}

func TestGetVSuperPodIds(t *testing.T) {
	tests := []struct {
		name      string
		faultJobs []FaultTask
		superPods map[string][]plugin.SuperNode
		want      []string
	}{
		{
			name:      "fault task with no superpod id",
			faultJobs: []FaultTask{{IsFaultTask: true, NodeName: "node1"}},
			superPods: map[string][]plugin.SuperNode{"vsp1": {{Name: "node2"}}},
			want:      []string{},
		},
		{
			name: "multiple fault tasks, some with superpod id",
			faultJobs: []FaultTask{
				{IsFaultTask: true, NodeName: "node1"},
				{IsFaultTask: true, NodeName: "node2"},
				{IsFaultTask: false, NodeName: "node3"},
			},
			superPods: map[string][]plugin.SuperNode{
				"vsp1": {{Name: "node1"}},
				"vsp2": {{Name: "node2"}},
			},
			want: []string{"vsp1", "vsp2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fJob := &FaultJob{
				FaultTasks: tt.faultJobs,
				SuperPods:  tt.superPods,
			}
			got := fJob.getVSuperPodIds()
			checkVSuperPodIds(t, got, tt.want)
		})
	}
}

type JudgeJobIsMasterFaultTest struct {
	name              string
	FaultTasks        []FaultTask
	PendingSessionNum int
	TpBlock           int
	SuperPods         map[string][]plugin.SuperNode
	IsMasterFault     bool
	vSuperPodIds      []string
	schedulerJob      *plugin.SchedulerJob
	want              bool
}

func buildJudgeJobIsMasterFaultTestCase1() JudgeJobIsMasterFaultTest {
	return JudgeJobIsMasterFaultTest{
		name: "master-0 pod is fault, should set IsMasterFault true",
		FaultTasks: []FaultTask{
			{IsFaultTask: true, NodeRankIndex: "0"},
		},
		vSuperPodIds: []string{},
		schedulerJob: &plugin.SchedulerJob{},
		want:         true,
	}
}
func buildJudgeJobIsMasterFaultTestCase2() JudgeJobIsMasterFaultTest {
	return JudgeJobIsMasterFaultTest{
		name: "process rescheduling, master-0 in same tp block, should set IsMasterFault true",
		FaultTasks: []FaultTask{
			{IsFaultTask: false, NodeRankIndex: "0"},
			{IsFaultTask: true, NodeRankIndex: "1"},
		},
		PendingSessionNum: tpPendingTimes,
		TpBlock:           2,
		vSuperPodIds:      []string{},
		schedulerJob: &plugin.SchedulerJob{
			SchedulerJobAttr: util.SchedulerJobAttr{
				ComJob: util.ComJob{
					Label: map[string]string{
						util.ProcessRecoverEnable: util.EnableFunc,
					},
				},
			},
		},
		want: true,
	}
}
func buildJudgeJobIsMasterFaultTestCase3() JudgeJobIsMasterFaultTest {
	return JudgeJobIsMasterFaultTest{
		name: "pendingSessionNum >= tpPendingTimes, master-0 in same tp block, should set IsMasterFault true",
		FaultTasks: []FaultTask{
			{IsFaultTask: false, NodeRankIndex: "0"},
			{IsFaultTask: true, NodeRankIndex: "1"},
		},
		PendingSessionNum: tpPendingTimes,
		TpBlock:           2,
		vSuperPodIds:      []string{},
		schedulerJob:      &plugin.SchedulerJob{},
		want:              true,
	}
}
func buildJudgeJobIsMasterFaultTestCase4() JudgeJobIsMasterFaultTest {
	return JudgeJobIsMasterFaultTest{
		name: "pendingSessionNum >= spPendingTimes, master-0 in same vsuperpod, should set IsMasterFault true",
		FaultTasks: []FaultTask{
			{IsFaultTask: false, NodeRankIndex: "0", NodeName: "node1"},
		},
		PendingSessionNum: spPendingTimes,
		SuperPods: map[string][]plugin.SuperNode{
			"vsp1": {{Name: "node1"}},
		},
		vSuperPodIds: []string{"vsp1"},
		schedulerJob: &plugin.SchedulerJob{},
		want:         true,
	}
}
func buildJudgeJobIsMasterFaultTestCase5() JudgeJobIsMasterFaultTest {
	return JudgeJobIsMasterFaultTest{
		name: "no master fault, should set IsMasterFault false",
		FaultTasks: []FaultTask{
			{IsFaultTask: false, NodeRankIndex: "1"},
		},
		vSuperPodIds: []string{},
		schedulerJob: &plugin.SchedulerJob{},
		want:         false,
	}
}
func buildJudgeJobIsMasterFaultTestCase6() JudgeJobIsMasterFaultTest {
	return JudgeJobIsMasterFaultTest{
		name: "pendingSessionNum >= tpPendingTimes, master-0 in same tp block, should set IsMasterFault true",
		FaultTasks: []FaultTask{
			{IsFaultTask: false, NodeRankIndex: "0"},
			{IsFaultTask: true, NodeRankIndex: "1"},
		},
		TpBlock:      2,
		vSuperPodIds: []string{},
		schedulerJob: &plugin.SchedulerJob{},
		want:         false,
	}
}

func TestJudgeJobIsMasterFault(t *testing.T) {
	tests := []JudgeJobIsMasterFaultTest{
		buildJudgeJobIsMasterFaultTestCase1(),
		buildJudgeJobIsMasterFaultTestCase2(),
		buildJudgeJobIsMasterFaultTestCase3(),
		buildJudgeJobIsMasterFaultTestCase4(),
		buildJudgeJobIsMasterFaultTestCase5(),
		buildJudgeJobIsMasterFaultTestCase6(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fJob := &FaultJob{
				FaultTasks:        tt.FaultTasks,
				PendingSessionNum: tt.PendingSessionNum,
				SuperPods:         tt.SuperPods,
				FaultJobA5Field: FaultJobA5Field{
					TpBlock:       tt.TpBlock,
					IsMasterFault: false,
				},
			}
			fJob.judgeJobIsMasterFault(tt.vSuperPodIds)
			if fJob.IsMasterFault != tt.want {
				t.Errorf("IsMasterFault = %v, want %v", fJob.IsMasterFault, tt.want)
			}
		})
	}
}

func TestGraceDeleteJobFor910A5(t *testing.T) {
	fJob := &FaultJob{JobUID: mockJobUID}
	env := plugin.ScheduleEnv{}
	env.SuperPodInfo = plugin.NewSuperPodInfo()
	npuJob := &plugin.SchedulerJob{}
	t.Run("01-GraceDeleteJobFor910A5 return error when ssn is nil", func(t *testing.T) {
		err := fJob.GraceDeleteJobFor910A5(nil, &plugin.SchedulerJob{}, env)
		if err == nil {
			t.Errorf("GraceDeleteJob() err = %v, wantErr is not nil", err)
		}
	})
	t.Run("02-GraceDeleteJobFor910A5 return error when npuJob is nil", func(t *testing.T) {
		err := fJob.GraceDeleteJobFor910A5(&framework.Session{}, nil, env)
		if err == nil {
			t.Errorf("GraceDeleteJob() err = %v, wantErr is not nil", err)
		}
	})
	t.Run("03-GraceDeleteJobFor910A5 return nil when ssn and npuJob are not nil", func(t *testing.T) {
		npuJob.Annotation = map[string]string{util.SuperPodAnnoKey: ""}
		err := fJob.GraceDeleteJobFor910A5(&framework.Session{}, npuJob, env)
		if err != nil {
			t.Errorf("GraceDeleteJob() err = %v, wantErr is nil", err)
		}
	})
}

func buildFaultJobForceDeleteJobFor910A5Tests() []FaultJobForceDeleteJobTests {
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
func TestFaultJobForceDeleteJobFor910A5(t *testing.T) {
	env := plugin.ScheduleEnv{}
	env.SuperPodInfo = plugin.NewSuperPodInfo()
	tests := buildFaultJobForceDeleteJobFor910A5Tests()
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
			if err := fJob.ForceDeleteJobFor910A5(tt.args.schedulerJob, env); (err != nil) != tt.wantErr {
				t.Errorf("ForceDeleteJobFor910A5() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.args.cacheFuncAfter()
		})
	}
}

type skipThisTaskTestCase struct {
	name            string
	fJob            *FaultJob
	cacheFuncBefore func()
	cacheFuncAfter  func()
	schedulerJob    *plugin.SchedulerJob
	dpi             *deletePodInfo
	fTask           FaultTask
	want            bool
}

func buildSkipThisTaskTestCase1() skipThisTaskTestCase {
	var tmpPatch *gomonkey.Patches = nil
	fJob := &FaultJob{}
	fJob.PendingSessionNum = tpPendingTimes
	schedulerJob := &plugin.SchedulerJob{}
	schedulerJob.Label = map[string]string{
		util.ProcessRecoverEnable: util.EnableFunc,
	}
	return skipThisTaskTestCase{
		name: "01-SkipThisTask return false when in the same tp block",
		fJob: fJob,
		cacheFuncBefore: func() {
			tmpPatch = gomonkey.ApplyPrivateMethod(reflect.TypeOf(&FaultJob{}), "inTheSameTpBlock",
				func(fTask FaultTask) bool { return true })
		},
		cacheFuncAfter: func() {
			if tmpPatch != nil {
				tmpPatch.Reset()
			}
		},
		dpi: &deletePodInfo{
			isSuperPod:    true,
			isMasterFault: false,
		},
		fTask: FaultTask{
			IsFaultTask: false,
		},
		schedulerJob: schedulerJob,
		want:         false,
	}
}

func buildSkipThisTaskTestCase2() skipThisTaskTestCase {
	var tmpPatch *gomonkey.Patches = nil
	schedulerJob := &plugin.SchedulerJob{}
	schedulerJob.Label = map[string]string{
		util.ProcessRecoverEnable: util.ProcessRecoverPause,
	}
	return skipThisTaskTestCase{
		name: "01-SkipThisTask return false when process rescheduling failed",
		fJob: &FaultJob{},
		cacheFuncBefore: func() {
			tmpPatch = gomonkey.ApplyPrivateMethod(reflect.TypeOf(&FaultJob{}), "inTheSameTpBlock",
				func(fTask FaultTask) bool { return true })
		},
		cacheFuncAfter: func() {
			if tmpPatch != nil {
				tmpPatch.Reset()
			}
		},
		dpi: &deletePodInfo{
			isSuperPod:    true,
			isMasterFault: false,
		},
		fTask: FaultTask{
			IsFaultTask: false,
		},
		schedulerJob: schedulerJob,
		want:         false,
	}
}

func buildSkipThisTaskTestCase3() skipThisTaskTestCase {
	var tmpPatch *gomonkey.Patches = nil
	fJob := &FaultJob{}
	fJob.PendingSessionNum = spPendingTimes
	schedulerJob := &plugin.SchedulerJob{}
	schedulerJob.Label = map[string]string{
		util.SinglePodTag: util.EnableFunc,
	}
	return skipThisTaskTestCase{
		name: "01-SkipThisTask return false when process in the same VSuperPod",
		fJob: fJob,
		cacheFuncBefore: func() {
			tmpPatch = gomonkey.ApplyPrivateMethod(reflect.TypeOf(&FaultJob{}), "inTheSameVSuperPod",
				func(fTask FaultTask) bool { return true })
		},
		cacheFuncAfter: func() {
			if tmpPatch != nil {
				tmpPatch.Reset()
			}
		},
		dpi: &deletePodInfo{
			isSuperPod:    true,
			isMasterFault: false,
		},
		fTask: FaultTask{
			IsFaultTask: false,
		},
		schedulerJob: schedulerJob,
		want:         false,
	}
}

func buildSkipThisTaskTestCase4() skipThisTaskTestCase {
	return skipThisTaskTestCase{
		name: "04-SkipThisTask allow-upgrade-false fault-task",
		fJob: &FaultJob{
			ReScheduleLimit: util.ReschedulingUpperLimitPod, // allowUpgradePodRescheduling() returns false
		},
		cacheFuncBefore: func() {},
		cacheFuncAfter:  func() {},
		fTask: FaultTask{
			IsFaultTask: true,
		},
		dpi: &deletePodInfo{},
		schedulerJob: &plugin.SchedulerJob{
			SchedulerJobAttr: util.SchedulerJobAttr{
				ComJob: util.ComJob{Label: map[string]string{}},
			},
		},
		want: false, // Should not skip fault task when allowUpgrade is false
	}
}

func buildSkipThisTaskTestCase5() skipThisTaskTestCase {
	return skipThisTaskTestCase{
		name: "05-SkipThisTask allow-upgrade-false non-fault-task",
		fJob: &FaultJob{
			ReScheduleLimit: util.ReschedulingUpperLimitPod, // allowUpgradePodRescheduling() returns false
		},
		cacheFuncBefore: func() {},
		cacheFuncAfter:  func() {},
		fTask: FaultTask{
			IsFaultTask: false,
		},
		dpi: &deletePodInfo{},
		schedulerJob: &plugin.SchedulerJob{
			SchedulerJobAttr: util.SchedulerJobAttr{
				ComJob: util.ComJob{Label: map[string]string{}},
			},
		},
		want: true, // Should skip non-fault task when allowUpgrade is false
	}
}

func buildSkipThisTaskTestCase6() skipThisTaskTestCase {
	return skipThisTaskTestCase{
		name: "06-SkipThisTask is-master-fault",
		fJob: &FaultJob{
			ReScheduleLimit: "", // allowUpgradePodRescheduling() returns true
			FaultJobA5Field: FaultJobA5Field{
				IsMasterFault: true,
			},
		},
		cacheFuncBefore: func() {},
		cacheFuncAfter:  func() {},
		fTask: FaultTask{
			IsFaultTask: false,
		},
		dpi: &deletePodInfo{},
		schedulerJob: &plugin.SchedulerJob{
			SchedulerJobAttr: util.SchedulerJobAttr{
				ComJob: util.ComJob{Label: map[string]string{}},
			},
		},
		want: false, // Should not skip when IsMasterFault is true
	}
}

func buildSkipThisTaskTestCase7() skipThisTaskTestCase {
	return skipThisTaskTestCase{
		name: "07-SkipThisTask process-rescheduling-skip",
		fJob: &FaultJob{
			ReScheduleLimit:   "", // allowUpgradePodRescheduling() returns true
			PendingSessionNum: tpPendingTimes - 1,
			FaultTasks:        []FaultTask{{IsFaultTask: true, NodeRankIndex: "0"}},
		},
		cacheFuncBefore: func() {},
		cacheFuncAfter:  func() {},
		fTask: FaultTask{
			IsFaultTask:   false,
			NodeRankIndex: "1",
		},
		dpi: &deletePodInfo{},
		schedulerJob: &plugin.SchedulerJob{
			SchedulerJobAttr: util.SchedulerJobAttr{
				ComJob: util.ComJob{Label: map[string]string{util.ProcessRecoverEnable: util.EnableFunc}},
			},
		},
		want: true, // Should skip when process rescheduling is in first stage
	}
}

func buildSkipThisTaskTestCase8() skipThisTaskTestCase {
	return skipThisTaskTestCase{
		name: "08-SkipThisTask pod-rescheduling-skip",
		fJob: &FaultJob{
			ReScheduleLimit:   "",           // allowUpgradePodRescheduling() returns true
			PendingSessionNum: pendingTimes, // Set to a value >= spPendingTimes
			FaultTasks:        []FaultTask{{IsFaultTask: true, NodeName: "test-node"}},
		},
		cacheFuncBefore: func() {},
		cacheFuncAfter:  func() {},
		fTask: FaultTask{
			IsFaultTask: false,
			NodeName:    "test-node", // Same node as fault task
		},
		dpi: &deletePodInfo{
			isSuperPod: true,
			ids:        []string{"test-id"},
		},
		schedulerJob: &plugin.SchedulerJob{
			SchedulerJobAttr: util.SchedulerJobAttr{
				ComJob: util.ComJob{Label: map[string]string{util.SinglePodTag: util.EnableFunc}},
			},
		},
		want: false, // Should not skip when pod rescheduling
	}
}

func buildSkipThisTaskTestCases() []skipThisTaskTestCase {
	return []skipThisTaskTestCase{
		buildSkipThisTaskTestCase1(),
		buildSkipThisTaskTestCase2(),
		buildSkipThisTaskTestCase3(),
		buildSkipThisTaskTestCase4(),
		buildSkipThisTaskTestCase5(),
		buildSkipThisTaskTestCase6(),
		buildSkipThisTaskTestCase7(),
		buildSkipThisTaskTestCase8(),
	}
}

func TestSkipThisTask(t *testing.T) {
	testCases := buildSkipThisTaskTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.cacheFuncBefore()
			if res := tt.fJob.skipThisTask(tt.dpi, tt.fTask, tt.schedulerJob); res != tt.want {
				t.Errorf("skipThisTask() res = %v, want is %v", res, tt.want)
			}
			tt.cacheFuncAfter()
		})
	}
}

func TestGraceDeletePodsFor910A5(t *testing.T) {
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
	t.Run("01-graceDeletePodsFor910A5 do not change IsBeingGracefulDeleted when npuTask not in session",
		func(t *testing.T) {
			var tmpPatch *gomonkey.Patches = nil
			tmpPatch = gomonkey.ApplyPrivateMethod(reflect.TypeOf(&util.NPUTask{}), "ForceDeletePodByTaskInf",
				func(ssn *framework.Session, reason string, nodeName string) error { return nil })
			fJob.graceDeletePodsFor910A5(ssn, npuJob, env, &deletePodInfo{})
			tmpPatch.Reset()
			for id := range fJob.FaultTasks {
				if fJob.FaultTasks[id].IsBeingGracefulDeleted == true {
					t.Error("graceDeletePodsFor910A5() return true, want false")
				}
			}
		})
	t.Run("02-graceDeletePodsFor910A5 change IsBeingGracefulDeleted when npuTask in session", func(t *testing.T) {
		npuJob.Label = map[string]string{}
		npuJob.Tasks = map[api.TaskID]util.NPUTask{mockTaskUID: {
			VTask: &util.VTask{Allocated: util.TaskAllocated{}}}}
		var tmpPatch *gomonkey.Patches = nil
		tmpPatch = gomonkey.ApplyPrivateMethod(reflect.TypeOf(&util.NPUTask{}), "ForceDeletePodByTaskInf",
			func(ssn *framework.Session, reason string, nodeName string) error { return nil })
		fJob.graceDeletePodsFor910A5(ssn, npuJob, env, &deletePodInfo{})
		tmpPatch.Reset()
		for id := range fJob.FaultTasks {
			if fJob.FaultTasks[id].IsBeingGracefulDeleted == false {
				t.Error("graceDeletePodsFor910A5() return false, want true")
			}
		}
	})
}
