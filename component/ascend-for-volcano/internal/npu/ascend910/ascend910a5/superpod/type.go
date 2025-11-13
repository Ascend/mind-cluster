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

// Package superpod for any struct used in a5
package superpod

import (
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910a5"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

type module910a5SuperPod struct {
	ascend910a5.Base910A5
	jobParams
	isSoftSuperPodAffinity bool
}

type jobParams struct {
	spBlock                  int
	tpBlock                  int
	totalCount               int
	scheduleStrategy         int
	netUnhealthyKey          string
	faultNPUKey              string
	whetherBackToVspSchedule bool
}

type superPodsInfo struct {
	superPodTable superPodOrderTable
	spCount       int
}

type superPodOrderTable = [][][]superPod

// sample node base info record each NPUNode but more sample
type nodeBaseInfo struct {
	name       string
	superPodID int32
	rackID     int32
}

type superPod = map[string]nodeBaseInfo

type jobCheckerFunc func() *api.ValidateResult

type nodeCheckerFunc func(*api.TaskInfo, plugin.NPUNode) error

type rackNpuTopType [rackNodeNum][nodeNPUNum]bool

const (
	// SuperPodAnnoKey the key of sp-block
	SuperPodAnnoKey = "sp-block"
	// SuperPodx8 the real label of a5 node
	SuperPodx8 = "900SuperPod-A5-8"
	// SuperPodx8SchedulerName maxNodeNPUNum is 8
	SuperPodx8SchedulerName = util.HwPreName + util.Ascend910 + SuperPodx8
	superPodAffinity        = "super-pod-affinity"
	softRequire             = "soft"
)
