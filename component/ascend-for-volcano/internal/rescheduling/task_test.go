/*
Copyright(C)2025-2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/k8s"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/consts"
)

type testCase struct {
	name                    string
	subHealthyStrategy      string
	hasSubHealthFault       bool
	annotations             map[string]string
	expectedIsFault         bool
	expectedFaultType       string
	expectedHotSwitchDelete bool
}

func buildTestCase(name string, hasSubHealthFault bool, annotations map[string]string,
	expectedIsFault bool, expectedFaultType string) testCase {
	return testCase{
		name:               name,
		subHealthyStrategy: util.SubHealthyHotSwitch,
		hasSubHealthFault:  hasSubHealthFault,
		annotations:        annotations,
		expectedIsFault:    expectedIsFault,
		expectedFaultType:  expectedFaultType,
	}
}

// TestGetTaskHealthStateBySubHealth tests the getTaskHealthStateBySubHealth method
func TestGetTaskHealthStateBySubHealth(t *testing.T) {
	tests := []testCase{
		buildTestCase("SubHealthyIgnore strategy should return healthy",
			true, map[string]string{}, false, PodHealthy),
		buildTestCase("No sub health fault should return healthy",
			false, map[string]string{}, false, PodHealthy),
		buildTestCase("HotSwitch strategy without delete annotation should return healthy",
			true, map[string]string{}, false, PodHealthy),
		buildTestCase("HotSwitch strategy with non-delete annotation should return healthy",
			true, map[string]string{}, false, PodHealthy),
		{
			name:               "HotSwitch delete annotation without backup pod should NOT mark hotSwitchDelete",
			subHealthyStrategy: util.SubHealthyHotSwitch,
			hasSubHealthFault:  true,
			annotations:        map[string]string{util.NeedVolcanoOpeKey: util.OpeTypeDelete},
			expectedIsFault:    true,
			expectedFaultType:  SubHealthFault,
		},
		{
			name:               "HotSwitch delete annotation with backup pod should mark hotSwitchDelete",
			subHealthyStrategy: util.SubHealthyHotSwitch,
			hasSubHealthFault:  true,
			annotations: map[string]string{util.NeedVolcanoOpeKey: util.OpeTypeDelete,
				consts.BackupNewPodNameKey: "backup-pod-1"},
			expectedIsFault:         true,
			expectedFaultType:       SubHealthFault,
			expectedHotSwitchDelete: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fTask := &FaultTask{
				HasSubHealthFault: tt.hasSubHealthFault,
				Annotations:       tt.annotations,
				TaskName:          "test-task",
				Reason:            nil,
			}

			isFault, faultType := fTask.getTaskHealthStateBySubHealth(tt.subHealthyStrategy)

			if isFault != tt.expectedIsFault {
				t.Errorf("getTaskHealthStateBySubHealth() isFault = %v, want %v", isFault, tt.expectedIsFault)
			}

			if faultType != tt.expectedFaultType {
				t.Errorf("getTaskHealthStateBySubHealth() faultType = %v, want %v", faultType, tt.expectedFaultType)
			}

			if fTask.IsHotSwitchDelete != tt.expectedHotSwitchDelete {
				t.Errorf("getTaskHealthStateBySubHealth() IsHotSwitchDelete = %v, want %v",
					fTask.IsHotSwitchDelete, tt.expectedHotSwitchDelete)
			}
		})
	}
}

func TestHasDpuNodeSeparateFault(t *testing.T) {
	tests := []struct {
		name       string
		nodeEvent  *k8s.DpuNodeEvent
		wantResult bool
	}{
		{
			name:       "01 return false when nodeEvent is nil",
			nodeEvent:  nil,
			wantResult: false,
		},
		{
			name: "02 return false when fault level is sub-health",
			nodeEvent: &k8s.DpuNodeEvent{FaultList: []k8s.DpuFaultDetail{
				{FaultLevel: util.SubHealthFault},
			}},
			wantResult: false,
		},
		{
			name: "03 return true when fault level is isolate",
			nodeEvent: &k8s.DpuNodeEvent{FaultList: []k8s.DpuFaultDetail{
				{FaultLevel: SeparateDPU},
			}},
			wantResult: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasDpuNodeSeparateFault(tt.nodeEvent); got != tt.wantResult {
				t.Errorf("hasDpuNodeSeparateFault() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func TestHasDpuNodeSubHealthFault(t *testing.T) {
	tests := []struct {
		name        string
		faultDpuList []k8s.DPUItem
		wantResult  bool
	}{
		{
			name:        "01 return false when faultDpuList is empty",
			faultDpuList: nil,
			wantResult:  false,
		},
		{
			name: "02 return false when fault level is isolate",
			faultDpuList: []k8s.DPUItem{
				{FaultList: []k8s.DpuFaultDetail{{FaultLevel: SeparateDPU}}},
			},
			wantResult: false,
		},
		{
			name: "03 return true when fault level is sub-health",
			faultDpuList: []k8s.DPUItem{
				{FaultList: []k8s.DpuFaultDetail{{FaultLevel: SubHealthFault}}},
			},
			wantResult: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasDpuNodeSubHealthFault(tt.faultDpuList); got != tt.wantResult {
				t.Errorf("hasDpuNodeSubHealthFault() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func TestIsDpuSeparateFault(t *testing.T) {
	tests := []struct {
		name       string
		dpu        k8s.DPUItem
		wantResult bool
	}{
		{
			name:       "01 return false when fault list is empty",
			dpu:        k8s.DPUItem{},
			wantResult: false,
		},
		{
			name: "02 return false when fault level is not handle",
			dpu: k8s.DPUItem{FaultList: []k8s.DpuFaultDetail{
				{FaultLevel: NotHandleFault},
			}},
			wantResult: false,
		},
		{
			name: "03 return true when fault level is isolate",
			dpu: k8s.DPUItem{FaultList: []k8s.DpuFaultDetail{
				{FaultLevel: SeparateDPU},
			}},
			wantResult: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDpuSeparateFault(tt.dpu); got != tt.wantResult {
				t.Errorf("IsDpuSeparateFault() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func TestTaskUsesFaultDpuAffectedNpu(t *testing.T) {
	fNode := &FaultNode{
		NodeName: "node0",
		FaultDpuList: []k8s.DPUItem{
			{
				HcaName:     "mlx5_0",
				AffectedNPU: []int{0, 1},
				FaultList:   []k8s.DpuFaultDetail{{FaultLevel: SeparateDPU}},
			},
		},
	}
	tests := []struct {
		name       string
		fTask      *FaultTask
		wantResult bool
	}{
		{
			name:       "01 return true when task uses NPU affected by faulty dpu",
			fTask:      &FaultTask{UseCardName: []string{"Ascend910-0"}},
			wantResult: true,
		},
		{
			name:       "02 return false when task does not use affected NPU",
			fTask:      &FaultTask{UseCardName: []string{"Ascend910-7"}},
			wantResult: false,
		},
		{
			name:       "03 return false when dpu fault is not isolate level",
			fTask:      &FaultTask{UseCardName: []string{"Ascend910-0"}},
			wantResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "03 return false when dpu fault is not isolate level" {
				fNode = &FaultNode{
					FaultDpuList: []k8s.DPUItem{
						{AffectedNPU: []int{0}, FaultList: []k8s.DpuFaultDetail{{FaultLevel: NotHandleFault}}},
					},
				}
			}
			if got := tt.fTask.taskUsesFaultDpuAffectedNpu(fNode); got != tt.wantResult {
				t.Errorf("taskUsesFaultDpuAffectedNpu() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}
