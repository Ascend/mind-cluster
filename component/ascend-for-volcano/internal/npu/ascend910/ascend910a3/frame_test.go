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
Package ascend910a3 is using for A3 affinity schedule.
*/
package ascend910a3

import (
	"fmt"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
)

const (
	// TrainingWorker represents the training worker task ID for testing
	TrainingWorker = "training-worker"
	// TrainingMaster represents the training master task ID for testing
	TrainingMaster = "training-master"
	// MaxNPUNum represents the maximum number of NPUs per node
	MaxNPUNum = 16
	// IncorrectNPUNum represents an incorrect NPU number for testing failure scenarios
	IncorrectNPUNum = 8
)

// TaskConfig represents configuration for a single task
type TaskConfig struct {
	TaskID     api.TaskID
	ReqNPUNum  int
	Annotation map[string]string
}

// newBase910A3 creates a new Base910A3 instance with the given configuration
func newBase910A3(maxNodeNPUNum int, tasks []TaskConfig) *Base910A3 {
	taskMap := make(map[api.TaskID]util.NPUTask)
	for _, task := range tasks {
		taskMap[task.TaskID] = util.NPUTask{
			ReqNPUNum:  task.ReqNPUNum,
			Annotation: task.Annotation,
		}
	}

	return &Base910A3{
		NPUHandler: base.NPUHandler{
			MaxNodeNPUNum: maxNodeNPUNum,
			SchedulerJobAttr: util.SchedulerJobAttr{
				NPUJob: &util.NPUJob{
					Tasks: taskMap,
				},
			},
		},
	}
}

func TestBase910A3_CheckReqNPUEqualNodeNPU(t *testing.T) {
	tests := []struct {
		name           string
		base910A3      *Base910A3
		expectedResult *api.ValidateResult
	}{
		{
			name:           "01 All tasks have correct NPU number - should pass",
			base910A3:      newBase910A3(MaxNPUNum, []TaskConfig{{TaskID: TrainingWorker, ReqNPUNum: MaxNPUNum}, {TaskID: TrainingMaster, ReqNPUNum: MaxNPUNum}}),
			expectedResult: nil, // Should pass, return nil
		},
		{
			name:           "02 Task with zero NPU and scheduler spec annotation - should pass",
			base910A3:      newBase910A3(MaxNPUNum, []TaskConfig{{TaskID: TrainingWorker, Annotation: map[string]string{taskSpec: schedulerSpec}}}),
			expectedResult: nil, // Should pass, return nil
		},
		{
			name:           "03 Task with zero NPU and skip ascend plugin annotation - should pass",
			base910A3:      newBase910A3(MaxNPUNum, []TaskConfig{{TaskID: TrainingWorker, Annotation: map[string]string{skipAscendPlugin: skipEnabled}}}),
			expectedResult: nil, // Should pass, return nil
		},
		{
			name:           "04 Empty tasks list - should pass",
			base910A3:      newBase910A3(MaxNPUNum, []TaskConfig{}),
			expectedResult: nil, // Should pass, return nil
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.base910A3.CheckReqNPUEqualNodeNPU()
			if result != nil {
				t.Errorf("CheckReqNPUEqualNodeNPU() = %v, expected nil", result)
			}
		})
	}
}

func TestBase910A3_CheckReqNPUEqualNodeNPU_ExpectError(t *testing.T) {
	tests := []struct {
		name           string
		base910A3      *Base910A3
		expectedResult *api.ValidateResult
	}{
		{
			name:           "01 Task with zero NPU but no special annotations - should fail",
			base910A3:      newBase910A3(MaxNPUNum, []TaskConfig{{TaskID: TrainingWorker}}),
			expectedResult: &api.ValidateResult{Pass: false, Reason: JobCheckFailedReason, Message: fmt.Sprintf("distributed job require npu %d, instead of 0", MaxNPUNum)},
		},
		{
			name:           "02 Task with incorrect NPU number - should fail",
			base910A3:      newBase910A3(MaxNPUNum, []TaskConfig{{TaskID: TrainingWorker, ReqNPUNum: IncorrectNPUNum}}),
			expectedResult: &api.ValidateResult{Pass: false, Reason: JobCheckFailedReason, Message: fmt.Sprintf("distributed job require npu %d, instead of %d", MaxNPUNum, IncorrectNPUNum)},
		},
		{
			name:           "03 Mixed tasks - some pass, some fail - should fail on first failure",
			base910A3:      newBase910A3(MaxNPUNum, []TaskConfig{{TaskID: TrainingWorker, ReqNPUNum: MaxNPUNum}, {TaskID: TrainingMaster, ReqNPUNum: IncorrectNPUNum}}),
			expectedResult: &api.ValidateResult{Pass: false, Reason: JobCheckFailedReason, Message: fmt.Sprintf("distributed job require npu %d, instead of %d", MaxNPUNum, IncorrectNPUNum)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.base910A3.CheckReqNPUEqualNodeNPU()
			if result == nil {
				t.Errorf("CheckReqNPUEqualNodeNPU() = nil, expected %v", tt.expectedResult)
			} else {
				if result.Pass != tt.expectedResult.Pass {
					t.Errorf("CheckReqNPUEqualNodeNPU().Pass = %v, expected %v", result.Pass, tt.expectedResult.Pass)
				}
				if result.Reason != tt.expectedResult.Reason {
					t.Errorf("CheckReqNPUEqualNodeNPU().Reason = %v, expected %v", result.Reason, tt.expectedResult.Reason)
				}
				if result.Message != tt.expectedResult.Message {
					t.Errorf("CheckReqNPUEqualNodeNPU().Message = %v, expected %v", result.Message, tt.expectedResult.Message)
				}
			}
		})
	}
}
