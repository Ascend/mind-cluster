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
Package card910bx2 is using for HuaWei Ascend 910B(Atlas 300T A2) card pin affinity schedule.
*/
package card910bx2infer

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend910/ascend910b"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
)

type card910bx2infer struct {
	ascend910b.Base910b
	reHandle *rescheduling.ReScheduler
}

const (
	// SchedulerName name of scheduler
	SchedulerName = "huawei.com/Ascend910card-910b-infer"
	nodeNPUNumber = 2
)
