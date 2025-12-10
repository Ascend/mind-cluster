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
	ChipKind = "910"
	// HwPreName is the prefix of npu resource.
	HwPreName = "huawei.com/"
	// NPU910CardName for judge 910 npu resource.
	NPU910CardName = "huawei.com/Ascend910"
	// NPU910CardNamePre for getting card number.
	NPU910CardNamePre = "Ascend910-"
	// NPU310PCardName for judge 310P npu resource.
	NPU310PCardName = "huawei.com/Ascend310P"
	// NPU310CardName for judge 310 npu resource.
	NPU310CardName = "huawei.com/Ascend310"
	// NPU310CardNamePre for getting card number.
	NPU310CardNamePre = "Ascend310-"
	// NPU310PCardNamePre for getting card number.
	NPU310PCardNamePre = "Ascend310P-"
	// AscendNPUPodRealUse for NPU pod real use cards.
	AscendNPUPodRealUse = "huawei.com/AscendReal"
	// AscendNPUCore for NPU core num, like 56; Records the chip name that the scheduler assigns to the pod.
	AscendNPUCore = "huawei.com/npu-core"
	// Ascend910bName for judge Ascend910b npu resource.
	Ascend910bName = "huawei.com/Ascend910b"

	// Ascend310P device type 310P
	Ascend310P = "Ascend310P"
	// Ascend310 device type 310
	Ascend310 = "Ascend310"
	// Ascend910 device type 910
	Ascend910 = "Ascend910"
	// Pod910DeviceKey pod annotation key, for generate 910 hccl rank table
	Pod910DeviceKey = "ascend.kubectl.kubernetes.io/ascend-910-configuration"
	// JobKind910Value in ring-controller.atlas.
	JobKind910Value = "ascend-910"
	// JobKind310Value in ring-controller.atlas.
	JobKind310Value = "ascend-310"
	// JobKind310PValue 310p ring controller name
	JobKind310PValue = "ascend-310P"
	// JobKind910BValue 910B ring controller name
	JobKind910BValue = "ascend-910b"
	// Module910bx16AcceleratorType for module mode.
	Module910bx16AcceleratorType = "module-910b-16"
	// Module910bx8AcceleratorType for module mode.
	Module910bx8AcceleratorType = "module-910b-8"
	// Module910A3x16AcceleratorType for module mode.
	Module910A3x16AcceleratorType = "module-a3-16"
	// Module910A3SuperPodAcceleratorType for 910A3-SuperPod hardware
	Module910A3SuperPodAcceleratorType = "module-a3-16-super-pod"
	// Accelerator310Key accelerator key of old infer card
	Accelerator310Key = "npu-310-strategy"
)

// constants for schedule_policy
const (
	// SchedulePolicyAnnoKey annotation key for schedule policy
	SchedulePolicyAnnoKey = "huawei.com/schedule_policy"
	// SchedulePolicyA3x16 schedule policy for a3-16 server
	SchedulePolicyA3x16 = "module-a3-16"
	// SchedulePolicySuperPod schedule policy for a3 super-pod, if added, func IsSuperPodJob need adaptation.
	SchedulePolicySuperPod = "module-a3-16-super-pod"
)

// constants for ome inference service
const (
	// OmeInferenceServiceKey indicate this pod belongs to ome inference-service
	OmeInferenceServiceKey = "ome.io/inferenceservice"
)

// constants for MindIE
const (
	// SuperPodFitAnnoKey decide schedule policy of super-pod
	SuperPodFitAnnoKey = "sp-fit"
)

const (
	// MinAvailableKey decide minAvailable of task
	MinAvailableKey = "huawei.com/schedule_minAvailable"
	// RecoverPolicyPathKey decide recover policy path
	RecoverPolicyPathKey = "huawei.com/recover_policy_path"
	// ReschedulingUpperLimitPod means volcano only rescheduling fault pod rather than super-pod or job
	ReschedulingUpperLimitPod = "pod"
	// ReschedulingUpperLimitJob means volcano only rescheduling fault job
	ReschedulingUpperLimitJob = "job"
	// ReschedulingPodUpgradeToJob means volcano rescheduling fault, and then rescheduling fault Job when pod pending
	ReschedulingPodUpgradeToJob = "pod,job"
)

const (
	// SuperPodRankKey logic SuperPod id annotation key
	SuperPodRankKey = "super-pod-rank"
	// SuperPodIdKey physic SuperPod id annotation key
	SuperPodIdKey = "super-pod-id"
)
