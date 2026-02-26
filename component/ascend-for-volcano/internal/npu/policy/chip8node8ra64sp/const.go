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

// Package chip8node8ra64sp for constants used by a5
package chip8node8ra64sp

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// for business use
const (
	// UninitializedRestRackLenMapId not initialized RestRackLenMapId
	UninitializedRestRackLenMapId = -1
	tpRescheduleStage                   = 3
	spRescheduleStage                   = 6
	backToJobRescheduleStage            = 12
)

const (
	rackNodeNum = 8
	nodeNPUNum  = 8

	npuNumber8 = 8

	networkUnhealthyNPU = util.NPUCardName + "-NetworkUnhealthy"
	dpuUnhealthyNPU     = util.NPUCardName + "-DPUUnhealthy"
	faultNPU            = util.NPUCardName + "-Fault"
	nodeNum1            = 1

	tpBlock1 = 1

	npuTaskNum8       = 8
	miniTpBlockNum    = 1
	rackNPUNumber     = 64
	maxSuperPodNPUNum = 8192
	uBMemRackNumber   = 16

	scoreForNode         = 100000000
	jobCheckFailedReason = "npu num is invalid"

	spBlockInvalidReason      = "Parameter sp-block is invalid."
	tpBlockInvalidReason      = "Parameter tp-block is invalid."
	superPodSizeInvalidReason = "Parameter super-pod-size is invalid."

	taskSpec                   = "volcano.sh/task-spec"
	getNPUFromPodFailedPattern = "%s getUsableTopFromNode err: %s"
	uBMemory                   = "huawei.com/schedule_ubmemory"
	uBMemoryRequire            = "true"
	maxNpuNumInUBMemScene      = 1024

	// TaskSpecAnno used in pod annotation when EnableGangScheduling is true
	TaskSpecAnno = "volcano.sh/task-spec"
	// SchedulerType the type of Scheduler for mindspore
	SchedulerType string = "scheduler"
)
