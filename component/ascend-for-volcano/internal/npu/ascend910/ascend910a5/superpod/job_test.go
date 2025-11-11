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

// Package superpod for job test
package superpod

import (
	"fmt"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	itest "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

// for test cases use
const (
	superPodSize32 = 32
	npuTaskNum1    = 1
	npuTaskNum2    = 2
	npuTaskNum3    = 3
	npuTaskNum4    = 4
	npuTaskNum5    = 5
	npuTaskNum12   = 12
)

// ValidNPUJobTestCase valid job tests use
type ValidNPUJobTestCase struct {
	Name          string
	WantErr       *api.ValidateResult
	Attr          util.SchedulerJobAttr
	SchedulerName string
	ScheduleEnv   plugin.ScheduleEnv
	SpBlockNum    int
	TpBlockNum    int
}

func buildTestJobAttr(npuTaskNum int, reqTaskNPU string) util.SchedulerJobAttr {
	job := test.FakeNormalTestJob("job01", npuTaskNum)
	test.SetFakeJobResRequest(job, util.NPU910CardName, reqTaskNPU)
	attr := itest.FakeSchedulerJobAttrByJob(job)
	return attr
}

// TestValidNPUJob for ValidNPUJob
func TestValidNPUJob(t *testing.T) {
	testCases := buildCheckSpBlockValidCase()
	testCases = append(testCases, buildCheckTpBlockNumCase()...)
	testCases = append(testCases, buildCalculateTpBlockAndCheckCase()...)
	testCases = append(testCases, buildCheckJobReqNpuNumCase()...)
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			npu := New(tt.SchedulerName)
			npu.ScheduleEnv = tt.ScheduleEnv
			tt.Attr.SpBlockNPUNum = tt.SpBlockNum
			tt.Attr.TpBlockNPUNum = tt.TpBlockNum
			npu.SetSchedulerAttr(tt.Attr)
			if err := npu.ValidNPUJob(); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("ValidNPUJob() error = %v, wantErr %+v", err, tt.WantErr)
			}
		})
	}
}

func setSuperPodSize(superpodSize int) plugin.VolcanoFrame {
	return plugin.VolcanoFrame{
		ConfigParameters: plugin.ConfigParameters{
			DynamicParameters: plugin.DynamicParameters{
				SuperPodSize: superpodSize,
			},
		},
	}
}

func buildCheckSpBlockValidCase() []ValidNPUJobTestCase {
	return []ValidNPUJobTestCase{
		{
			Name: "checkSpBlockValid-01: Parameter sp-block is invalid." +
				"should return nil",
			SchedulerName: SuperPodx8SchedulerName,
			Attr:          buildTestJobAttr(npuTaskNum1, "8"),
			SpBlockNum:    0,
			TpBlockNum:    1,
			ScheduleEnv: plugin.ScheduleEnv{
				FrameAttr: setSuperPodSize(superPodSize32),
			},
			WantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  spBlockInvalidReason,
				Message: fmt.Sprintf("Parameter sp-block(%d) is invalid.", 0),
			},
		},
		{
			Name:          "checkSpBlockValid-02: Parameter sp-block(24) is not multiple of node npu (8)",
			Attr:          buildTestJobAttr(npuTaskNum1, "8"),
			SpBlockNum:    10,
			TpBlockNum:    1,
			SchedulerName: SuperPodx8SchedulerName,
			ScheduleEnv: plugin.ScheduleEnv{
				FrameAttr: setSuperPodSize(superPodSize32),
			},
			WantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  spBlockInvalidReason,
				Message: "Parameter sp-block(10) is not multiple of node npu (8)",
			},
		},
		{
			Name: "checkSpBlockValid-03: " +
				"job require total Pod(5) should be multiple of a sp-block size 4",
			Attr:          buildTestJobAttr(npuTaskNum5, "8"),
			SpBlockNum:    32,
			TpBlockNum:    1,
			SchedulerName: SuperPodx8SchedulerName,
			ScheduleEnv: plugin.ScheduleEnv{
				FrameAttr: setSuperPodSize(superPodSize32),
			},
			WantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  "job task num is invalid",
				Message: "job require total Pod(5) should be multiple of a sp-block size 4",
			},
		},
	}
}

func buildCheckTpBlockNumCase() []ValidNPUJobTestCase {
	return []ValidNPUJobTestCase{
		{
			Name:          "checkTpBlockNum-01: Parameter tp-block is invalid, it should be a number in the range",
			Attr:          buildTestJobAttr(npuTaskNum4, "8"),
			SpBlockNum:    32,
			TpBlockNum:    128,
			SchedulerName: SuperPodx8SchedulerName,
			ScheduleEnv: plugin.ScheduleEnv{
				FrameAttr: setSuperPodSize(superPodSize32),
			},
			WantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  tpBlockInvalidReason,
				Message: "Parameter tp-block is invalid, it should be a number in the range from 1 to 64",
			},
		},
		{
			Name:          "checkTpBlockNum-02: Parameter tp-block(48) must be the power of 2",
			Attr:          buildTestJobAttr(npuTaskNum8, "8"),
			SpBlockNum:    64,
			TpBlockNum:    48,
			SchedulerName: SuperPodx8SchedulerName,
			ScheduleEnv: plugin.ScheduleEnv{
				FrameAttr: setSuperPodSize(superPodSize32),
			},
			WantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  tpBlockInvalidReason,
				Message: "Parameter tp-block(48) must be the power of 2",
			},
		},
	}
}

func buildCalculateTpBlockAndCheckCase01() ValidNPUJobTestCase {
	return ValidNPUJobTestCase{
		Name: "calculateTpBlockAndCheck-01: " +
			"Parameter tp-block(32)/8 could not be bigger than sp-block(16)/8",
		Attr:          buildTestJobAttr(npuTaskNum2, "8"),
		SpBlockNum:    16,
		TpBlockNum:    32,
		SchedulerName: SuperPodx8SchedulerName,
		ScheduleEnv: plugin.ScheduleEnv{
			FrameAttr: setSuperPodSize(superPodSize32),
		},
		WantErr: &api.ValidateResult{
			Pass:    false,
			Reason:  tpBlockInvalidReason,
			Message: "Parameter tp-block(32)/8 could not be bigger than sp-block(16)/8",
		},
	}
}

func buildCalculateTpBlockAndCheckCase02() ValidNPUJobTestCase {
	return ValidNPUJobTestCase{
		Name: "calculateTpBlockAndCheck-02: " +
			"number of tasks(3) must be multiple of nodes occupied by tp-block(2)",
		Attr:          buildTestJobAttr(npuTaskNum3, "8"),
		SpBlockNum:    24,
		TpBlockNum:    16,
		SchedulerName: SuperPodx8SchedulerName,
		ScheduleEnv: plugin.ScheduleEnv{
			FrameAttr: setSuperPodSize(superPodSize32),
		},
		WantErr: &api.ValidateResult{
			Pass:    false,
			Reason:  tpBlockInvalidReason,
			Message: "number of tasks(3) must be multiple of nodes occupied by tp-block(2)",
		},
	}
}

func buildCalculateTpBlockAndCheckCase03() ValidNPUJobTestCase {
	return ValidNPUJobTestCase{
		Name: "calculateTpBlockAndCheck-03: " +
			"Parameter sp-block(32)/8 must be multiple of nodes occupied by NPUTaskNum(16)/8",
		Attr:          buildTestJobAttr(npuTaskNum12, "8"),
		SpBlockNum:    24,
		TpBlockNum:    16,
		SchedulerName: SuperPodx8SchedulerName,
		ScheduleEnv: plugin.ScheduleEnv{
			FrameAttr: setSuperPodSize(superPodSize32),
		},
		WantErr: &api.ValidateResult{
			Pass:    false,
			Reason:  tpBlockInvalidReason,
			Message: "spBlock= 24 / 8 must be multiple of tpBlock= 16 / 8",
		},
	}
}

func buildCalculateTpBlockAndCheckCase() []ValidNPUJobTestCase {
	return []ValidNPUJobTestCase{
		buildCalculateTpBlockAndCheckCase01(),
		buildCalculateTpBlockAndCheckCase02(),
		buildCalculateTpBlockAndCheckCase03(),
	}
}

func buildCheckJobReqNpuNumCase() []ValidNPUJobTestCase {
	return []ValidNPUJobTestCase{
		{
			Name:          "checkJobReqNpuNum-01: single super-pod job require npu [1, 8]",
			Attr:          buildTestJobAttr(npuTaskNum1, "9"),
			SpBlockNum:    1,
			TpBlockNum:    1,
			SchedulerName: SuperPodx8SchedulerName,
			ScheduleEnv: plugin.ScheduleEnv{
				FrameAttr: setSuperPodSize(superPodSize32),
			},
			WantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  jobCheckFailedReason,
				Message: "single super-pod job require npu [1, 8], instead of 9",
			},
		},
		{
			Name:          "checkJobReqNpuNum-02: distributed super-pod job require npu should be multiple of sp-block",
			Attr:          buildTestJobAttr(npuTaskNum2, "8"),
			SpBlockNum:    6,
			TpBlockNum:    1,
			SchedulerName: SuperPodx8SchedulerName,
			ScheduleEnv: plugin.ScheduleEnv{
				FrameAttr: setSuperPodSize(superPodSize32),
			},
			WantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  jobCheckFailedReason,
				Message: "distributed super-pod job require npu should be multiple of sp-block, instead of 16",
			},
		},
		{
			Name:          "checkJobReqNpuNum-03: distributed super-pod job require npu should be multiple of sp-block",
			Attr:          buildTestJobAttr(npuTaskNum3, "6"),
			SpBlockNum:    1,
			TpBlockNum:    1,
			SchedulerName: SuperPodx8SchedulerName,
			ScheduleEnv: plugin.ScheduleEnv{
				FrameAttr: setSuperPodSize(superPodSize32),
			},
			WantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  jobCheckFailedReason,
				Message: "distributed job require npu 8, instead of 6",
			},
		},
	}
}
