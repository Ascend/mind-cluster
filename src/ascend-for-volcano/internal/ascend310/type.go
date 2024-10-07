/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package ascend310 is using for HuaWei A800/9000 Ascend310 pin affinity schedule.
*/
package ascend310

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const (
	// PluginName the module910's plugin name.
	PluginName = util.NPU310CardName
	// Accelerator310Key accelerator key of 310
	Accelerator310Key = "npu-310-strategy"
	// Card310AcceleratorValue card value
	Card310AcceleratorValue = "card"
	// Chip310AcceleratorValue chip value
	Chip310AcceleratorValue = "chip"
)

type asend310 struct {
	// need plugin
	base.NPUHandler
	// 310 support scheduler kinds.
	Kind map[string]base.AscendHandler
	// specific job use.
	handle   base.AscendHandler
	reHandle *rescheduling.ReScheduler
}
