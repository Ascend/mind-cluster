/*
Copyright(C)2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package module910bx8 is using for HuaWei Ascend910Bx8 pin affinity schedule.
*/
package module910bx8

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend910/ascend910b"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
)

type module910bx8 struct {
	ascend910b.Base910b
	reHandle        *rescheduling.ReScheduler
	netUnhealthyKey string
}

const (
	// SchedulerName name of scheduler
	SchedulerName       = "huawei.com/Ascend910module-910b-8"
	nodeNPUNumber       = 8
	networkUnhealthyNPU = "huawei.com/Ascend910-NetworkUnhealthy"
)
