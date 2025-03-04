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
Package ascend310p is using for HuaWei Ascend pin affinity schedule.
*/
package ascend310p

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu"
)

type ascend310P struct {
	// base event handler
	base.NPUHandler
	vHandle *vnpu.VirtualNPU
	// 300I duo support scheduler kinds.
	Kind map[string]base.AscendHandler
	// specific job use.
	handle base.AscendHandler
}

const (
	// PluginName ascend31P plugin name
	PluginName    = util.NPU310PCardName
	maxNodeNPUNum = 64
	// Accelerator310Key accelerator key of 310
	Accelerator310Key = "npu-310-strategy"
	// Chip310AcceleratorValue chip value
	Chip310AcceleratorValue = "chip"
	// DuoKeyLabel key and label for 300i duo
	DuoKeyLabel = "duo"
	// TrueStr true or false
	TrueStr = "true"
)
