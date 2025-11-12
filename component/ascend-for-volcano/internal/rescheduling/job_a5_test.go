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
	"strconv"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
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
					ComJob: util.ComJob{Selector: map[string]string{
						util.AcceleratorType: "910B",
					}}}},
			want: false, // 假设CheckA5Label("910B") == false
		},
		{
			name: "selector has A5",
			job: &plugin.SchedulerJob{
				SchedulerJobAttr: util.SchedulerJobAttr{
					ComJob: util.ComJob{Selector: map[string]string{
						util.AcceleratorType: "900SuperPod-A5-8",
					}}}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := is910A5Job(tt.job); got != tt.want {
				t.Errorf("is910A5Job() = %v, want %v", got, tt.want)
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
	name                     string
	TpBlock                  int
	WhetherBackToVspSchedule bool
	SuperPods                map[string][]plugin.SuperNode
	ids                      []string
	nodeName                 string
	want                     bool
}

func buildTestCase0() inTheSameVSuperPodTestCase {
	return inTheSameVSuperPodTestCase{
		name:                     "00-test The precondition failed",
		TpBlock:                  testTp2,
		WhetherBackToVspSchedule: false,
		ids:                      []string{},
		nodeName:                 "",
		want:                     false,
	}
}

func buildTestCase1() inTheSameVSuperPodTestCase {
	return inTheSameVSuperPodTestCase{
		name:                     "01-test ids don't exist",
		TpBlock:                  testTp1,
		WhetherBackToVspSchedule: true,
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

func buildTestCase2() inTheSameVSuperPodTestCase {
	return inTheSameVSuperPodTestCase{
		name:                     "02-test nodeName don't exist",
		TpBlock:                  testTp1,
		WhetherBackToVspSchedule: true,
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
		buildTestCase2(),
	}
}

func TestInTheSameVSuperPod(t *testing.T) {
	for _, tc := range buildInTheSameVSuperPodTestCases() {
		fJob := &FaultJob{
			FaultJobA5Field: FaultJobA5Field{
				TpBlock:                  tc.TpBlock,
				WhetherBackToVspSchedule: tc.WhetherBackToVspSchedule},
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

type processReschedulingSkipTaskTestCase struct {
	name              string
	PendingSessionNum int
	fTask             FaultTask
	TpBlock           int
	schedulerJob      plugin.SchedulerJob
	want1             bool
	want2             bool
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
	name                     string
	FaultTasks               []FaultTask
	PendingSessionNum        int
	TpBlock                  int
	SuperPods                map[string][]plugin.SuperNode
	IsMasterFault            bool
	WhetherBackToVspSchedule bool
	vSuperPodIds             []string
	schedulerJob             *plugin.SchedulerJob
	want                     bool
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
		TpBlock:      2,
		vSuperPodIds: []string{},
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
		name: "pendingSessionNum >= spPendingTimes, master-0 in same tp block, should set IsMasterFault true",
		FaultTasks: []FaultTask{
			{IsFaultTask: false, NodeRankIndex: "0"},
			{IsFaultTask: true, NodeRankIndex: "1"},
		},
		PendingSessionNum: spPendingTimes,
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
		WhetherBackToVspSchedule: true,
		vSuperPodIds:             []string{"vsp1"},
		schedulerJob:             &plugin.SchedulerJob{},
		want:                     true,
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

func TestJudgeJobIsMasterFault(t *testing.T) {
	tests := []JudgeJobIsMasterFaultTest{
		buildJudgeJobIsMasterFaultTestCase1(),
		buildJudgeJobIsMasterFaultTestCase2(),
		buildJudgeJobIsMasterFaultTestCase3(),
		buildJudgeJobIsMasterFaultTestCase4(),
		buildJudgeJobIsMasterFaultTestCase5(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fJob := &FaultJob{
				FaultTasks:        tt.FaultTasks,
				PendingSessionNum: tt.PendingSessionNum,
				SuperPods:         tt.SuperPods,
				FaultJobA5Field: FaultJobA5Field{
					TpBlock:                  tt.TpBlock,
					IsMasterFault:            false,
					WhetherBackToVspSchedule: tt.WhetherBackToVspSchedule,
				},
			}
			fJob.judgeJobIsMasterFault(tt.vSuperPodIds, tt.schedulerJob)
			if fJob.IsMasterFault != tt.want {
				t.Errorf("IsMasterFault = %v, want %v", fJob.IsMasterFault, tt.want)
			}
		})
	}
}
