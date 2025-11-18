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

// Package superpod for constants used by a5
package superpod

// for business use
const (
	// UninitializedRestRackLenMapId not initialized RestRackLenMapId
	UninitializedRestRackLenMapId int32 = -1
	// FirstRescheduleStage the first stage of rescheduling
	FirstRescheduleStage = 6
	// SecondRescheduleStage the second stage of rescheduling
	SecondRescheduleStage = 12
)

const (
	rackNodeNum = 8
	nodeNPUNum  = 8

	npuNumber8 = 8

	networkUnhealthyNPU = "huawei.com/Ascend910-NetworkUnhealthy"
	faultNPU            = "huawei.com/Ascend910-Fault"
	nodeNum1            = 1

	tpBlock1 = 1

	npuTaskNum8       = 8
	miniTpBlockNum    = 1
	rackNPUNumber     = 64
	maxSuperPodNPUNum = 8192

	scoreForNode         = 100000000
	jobCheckFailedReason = "npu num is invalid"

	spBlockInvalidReason      = "Parameter sp-block is invalid."
	tpBlockInvalidReason      = "Parameter tp-block is invalid."
	superPodSizeInvalidReason = "Parameter super-pod-size is invalid."

	taskSpec                   = "volcano.sh/task-spec"
	getNPUFromPodFailedPattern = "%s getUsableTopFromNode err: %s"

	// TaskSpecAnno used in pod annotation when EnableGangScheduling is true
	TaskSpecAnno = "volcano.sh/task-spec"
	// SchedulerType the type of Scheduler for mindspore
	SchedulerType string = "scheduler"
)
