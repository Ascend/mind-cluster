/*
Copyright(C)2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package half910x4 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.
*/
package half910x4

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
)

const (
	// SchedulerName Ascend910half plugin name
	SchedulerName       = "huawei.com/Ascend910half"
	npuNumPerHccs       = 4
	networkUnhealthyNPU = "huawei.com/Ascend910-NetworkUnhealthy"
	nodeWeight          = 8.0
)

type half910x4 struct {
	base.NPUHandler
	netUnhealthyKey string
	affScoreList    [][]int
	reHandle        *rescheduling.ReScheduler
}
