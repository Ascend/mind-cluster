/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package rescheduling

import (
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/k8s"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func TestFaultNodeSetSilentUnhealthyNPU(t *testing.T) {
	tests := []struct {
		name   string
		faults []k8s.FaultDeviceList
		want   []string
	}{
		{
			name: "collects silent fault cards",
			faults: []k8s.FaultDeviceList{
				{FaultLevel: SilentFault, NPUName: "Ascend910-0"},
				{FaultLevel: "OtherFault", NPUName: "Ascend910-1"},
				{FaultLevel: SilentFault, NPUName: "Ascend910-2"},
			},
			want: []string{"Ascend910-0", "Ascend910-2"},
		},
		{
			name: "skips empty npu name",
			faults: []k8s.FaultDeviceList{
				{FaultLevel: SilentFault, NPUName: ""},
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fNode := &FaultNode{FaultDeviceList: tt.faults}
			fNode.setSilentUnhealthyNPU()
			if !reflect.DeepEqual(fNode.SilentUnhealthyNPU, tt.want) {
				t.Fatalf("SilentUnhealthyNPU = %v, want %v", fNode.SilentUnhealthyNPU, tt.want)
			}
		})
	}
}

func TestFaultNodeIsSilentFaultNode(t *testing.T) {
	if got := (*FaultNode)(nil).IsSilentFaultNode(); got {
		t.Fatal("nil FaultNode should not be a silent fault node")
	}
	if got := (&FaultNode{}).IsSilentFaultNode(); got {
		t.Fatal("FaultNode with no silent cards should not be a silent fault node")
	}
	if got := (&FaultNode{SilentUnhealthyNPU: []string{"Ascend910-0"}}).IsSilentFaultNode(); !got {
		t.Fatal("FaultNode with silent cards should be a silent fault node")
	}
}

func TestJobSilentUnhealthyTolerance(t *testing.T) {
	reScheduler := &ReScheduler{
		DealReSchedulerCache: &DealReSchedulerCache{},
		Jobs: map[api.JobID]plugin.SchedulerJob{
			"job-1": {SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{
				ToleranceConfig: util.JobToleranceConfig{SilentUnhealthyToleranceAfterTime: 120},
			}}},
		},
	}
	if got := reScheduler.jobSilentUnhealthyTolerance("job-1"); got != 120 {
		t.Fatalf("jobSilentUnhealthyTolerance(job-1) = %d, want 120", got)
	}
	if got := reScheduler.jobSilentUnhealthyTolerance("unknown"); got != util.DefaultSilentUnhealthyToleranceAfterTime {
		t.Fatalf("jobSilentUnhealthyTolerance(unknown) = %d, want %d", got, util.DefaultSilentUnhealthyToleranceAfterTime)
	}
}

func TestIsTaskUsingSilentFaultCard(t *testing.T) {
	reScheduler := &ReScheduler{
		DealReSchedulerCache: &DealReSchedulerCache{
			FaultNodes: map[string]*FaultNode{
				"node-1": {SilentUnhealthyNPU: []string{"Ascend910-0", "Ascend910-2"}},
			},
		},
	}
	tests := []struct {
		name string
		task *FaultTask
		want bool
	}{
		{"nil task", nil, false},
		{"empty node name", &FaultTask{}, false},
		{"node not in fault nodes", &FaultTask{NodeName: "node-x"}, false},
		{"task uses a silent card", &FaultTask{NodeName: "node-1", UseCardName: []string{"Ascend910-0"}}, true},
		{"task does not use a silent card", &FaultTask{NodeName: "node-1", UseCardName: []string{"Ascend910-1"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reScheduler.isTaskUsingSilentFaultCard(tt.task); got != tt.want {
				t.Fatalf("isTaskUsingSilentFaultCard() = %v, want %v", got, tt.want)
			}
		})
	}
}
