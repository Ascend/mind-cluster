/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package label defines NPU node label keys and values.
package label

import "ascend-common/api"

// NPU label keys and values
const (
	// NPUChipNameLabel label value is npu chip name
	NPUChipNameLabel = api.ResourceNamePrefix + "npu.chip.name"
	// NPUChipMemoryLabel label value is npu chip memory
	NPUChipMemoryLabel = api.ResourceNamePrefix + "npu.chip.memory"
	// NPUChipBoardIDLabel label value is npu chip board id
	NPUChipBoardIDLabel = api.ResourceNamePrefix + "npu.chip.boardid"
	// NPUServerTypeLabel label value is server type
	NPUServerTypeLabel = api.ResourceNamePrefix + "npu.server.type"
	// NPUServerSerialNumberLabel label value is server serial number
	NPUServerSerialNumberLabel = api.ResourceNamePrefix + "npu.server.serial-number"
	// NPUDriverVersionLabel label value is driver version
	NPUDriverVersionLabel = api.ResourceNamePrefix + "npu.driver.version"
	// NPUChipProductTypeLabel label value is chip product type (replaces infer-card-type)
	NPUChipProductTypeLabel = api.ResourceNamePrefix + "npu.chip.product-type"
	// TopoLabelSuperPodId topological label for super-pod id
	TopoLabelSuperPodId = api.ResourceNamePrefix + "topotree.superpodid"
	// TopoLabelRackId topological label for rack id
	TopoLabelRackId = api.ResourceNamePrefix + "topotree.rackid"
	// TopoLabelServerId topological label for server id
	TopoLabelServerId = api.ResourceNamePrefix + "topotree.serverid"
)

// Accelerator-type label values
const (
	// AcceleratorTypeModule910A3x16SuperPod for 16-npu 910A3-SuperPod hardware
	AcceleratorTypeModule910A3x16SuperPod = "module-a3-16-super-pod"
	// AcceleratorTypeModule910A3x8SuperPod for 8-npu 910A3-SuperPod hardware
	AcceleratorTypeModule910A3x8SuperPod = "module-a3-8-super-pod"
)

// Accelerator label values
const (
	// Accelerator910Label accelerator label value for 910/910B/910A3
	Accelerator910Label = "huawei-Ascend910"
	// Accelerator310Label accelerator label value for 310
	Accelerator310Label = "huawei-Ascend310"
	// Accelerator310PLabel accelerator label value for 310P
	Accelerator310PLabel = "huawei-Ascend310P"
	// AcceleratorNPULabel accelerator label value for A5 and other npu
	AcceleratorNPULabel = "huawei-npu"
	// A300IA2Label the value of the A300I A2 node label
	A300IA2Label = "card-910b-infer"
	// A300IDuoLabel the value of the A300I Duo node label
	A300IDuoLabel = "card-300i-duo"
)

// Old label keys (deprecated, dual-write in Phase 1, Will be removed in Phase 2)
const (
	// Deprecated: NPUChipMemoryLabelDeprecated is the old npu chip memory label key,
	// replaced by NPUChipMemoryLabel. Will be removed in Phase 2.
	NPUChipMemoryLabelDeprecated = "mind-cluster/npu-chip-memory"

	// Deprecated: NPUChipNameLabelDeprecated is the old npu chip name label key,
	// replaced by NPUChipNameLabel. Will be removed in Phase 2.
	NPUChipNameLabelDeprecated = "node.kubernetes.io/npu.chip.name"

	// Deprecated: AcceleratorTypeKeyDeprecated is the old accelerator type label key.
	// Not migrated (no consumers). Will be removed in Phase 2.
	AcceleratorTypeKeyDeprecated = "accelerator-type"

	// AcceleratorLabelKeyDeprecated is the accelerator label key.
	// Not migrated to new prefix; consumers derive chipKind from chip.name instead.
	// Will be removed in Phase 2.
	AcceleratorLabelKeyDeprecated = "accelerator"

	// Deprecated: NPUServerTypeLabelDeprecated is the old server type label key,
	// replaced by NPUServerTypeLabel. Will be removed in Phase 2.
	NPUServerTypeLabelDeprecated = "servertype"

	// Deprecated: NPUChipProductTypeLabelDeprecated is the old infer card type label key,
	// replaced by NPUChipProductTypeLabel. Will be removed in Phase 2.
	NPUChipProductTypeLabelDeprecated = "infer-card-type"

	// Deprecated: NPUDriverVersionLabelDeprecated is the old driver version label key,
	// replaced by NPUDriverVersionLabel. Will be removed in Phase 2.
	NPUDriverVersionLabelDeprecated = "huawei.com/driver.version"
)
