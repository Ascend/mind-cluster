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
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
)

// Base910A3 is the struct of Base910A3.
type Base910A3 struct {
	base.NPUHandler
	// row: request npu num; column: usable npu num
	AffScoreList [][]int
}

const (
	// NodeNPUNumber16 is the number of NPU chips in a node, a3 is 16.
	NodeNPUNumber16 = 16
	// NodeNPUNumber8 is the number of NPU chips in a node, a3 is 16.
	NodeNPUNumber8 = 8
	// DieNPUNumber is the number of NPU chips in whole NPU, a3 is 2.
	DieNPUNumber = 2

	// JobCheckFailedReason is the reason of job check failed.
	JobCheckFailedReason = "npu num is invalid"
)

// annotation
const (
	// TaskSpecAnno used in pod annotation when EnableGangScheduling is true
	TaskSpecAnno = "volcano.sh/task-spec"
	// SchedulerType the type of Scheduler for mindspore
	SchedulerType string = "scheduler"
)
