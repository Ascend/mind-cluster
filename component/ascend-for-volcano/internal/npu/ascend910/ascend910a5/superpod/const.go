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
	// SuperPodA5Prefix the prefix value of superpod a5 handler name
	SuperPodA5Prefix = "900SuperPod-A5"
	// RackSchedule the rack schedule strategy selecting nodes in one rack
	RackSchedule = 0
	// SuperPodSchedule the superPod schedule strategy selecting nodes in one superPod
	SuperPodSchedule = 2
	// MulSuperPodsSchedule the multiple superPod schedule strategy selecting nodes in multiple superPods
	MulSuperPodsSchedule = 3
)

const (
	rackNodeNum = 8
	nodeNPUNum  = 8

	npuNumber8  = 8
	npuNumber16 = 16
	npuNumber32 = 32
	npuNumber64 = 64

	networkUnhealthyNPU = "huawei.com/Ascend910-NetworkUnhealthy"
	faultNPU            = "huawei.com/Ascend910-Fault"
	nodeNum1    = 1
	npuTaskNum8 = 8

	miniTpBlockNum = 1
	rackNPUNumber  = 64

	scoreForNode         = 100000000
	jobCheckFailedReason = "npu num is invalid"

	spBlockInvalidReason = "Parameter sp-block is invalid."
	tpBlockInvalidReason = "Parameter tp-block is invalid."

	taskSpec = "volcano.sh/task-spec"
)
