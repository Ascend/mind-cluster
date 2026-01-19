/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package util is using for the total variable.
*/
package util

const (
	// ChipKind is the prefix of npu resource.
	ChipKind = "A2G"
	// HwPreName is the prefix of npu resource.
	HwPreName = "npu.com/"
	// NPU910CardName for judge 910 npu resource.
	NPU910CardName = "npu.com/AlanA2G"
	// NPU910CardNamePre for getting card number.
	NPU910CardNamePre = "AlanA2G-"
	// NPU310PCardName for judge 310P npu resource.
	NPU310PCardName = "npu.com/AlanI2"
	// NPU310CardName for judge 310 npu resource.
	NPU310CardName = "npu.com/Alan310"
	// NPU310CardNamePre for getting card number.
	NPU310CardNamePre = "Alan310-"
	// NPU310PCardNamePre for getting card number.
	NPU310PCardNamePre = "AlanI2-"
	// AscendNPUPodRealUse for NPU pod real use cards.
	AscendNPUPodRealUse = "npu.com/AlanReal"
	// AscendNPUCore for NPU core num, like 56; Records the chip name that the scheduler assigns to the pod.
	AscendNPUCore = "npu.com/npu-core"
	// Ascend910bName for judge Ascend910b npu resource.
	Ascend910bName = "npu.com/AlanA2G"

	// Ascend310P device type 310P
	Ascend310P = "AlanI2"
	// Ascend310 device type 310
	Ascend310 = "Alan310"
	// Ascend910 device type 910
	Ascend910 = "AlanA2G"
	// Pod910DeviceKey pod annotation key, for generate 910 hccl rank table
	Pod910DeviceKey = "alan.kubectl.kubernetes.io/alan-a2g-configuration"
	// JobKind910Value in ring-controller.atlas.
	JobKind910Value = "alan-a1g"
	// JobKind310PValue 310p ring controller name
	JobKind310PValue = "alan-I2"
	// JobKind910BValue 910B ring controller name
	JobKind910BValue = "alan-a2g"
	// Module910bx16AcceleratorType for module mode.
	Module910bx16AcceleratorType = "module-a2g-16"
	// Module910bx8AcceleratorType for module mode.
	Module910bx8AcceleratorType = "module-a2g-8"
	// Accelerator310Key accelerator key of old infer card
	Accelerator310Key = "npu-I2-strategy"
	// A300IDuoLabel the value of the A300I Duo node label
	A300IDuoLabel = "card-i2-duo"
)
