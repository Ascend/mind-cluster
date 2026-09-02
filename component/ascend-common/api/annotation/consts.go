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

// Package annotation defines NPU node annotation keys.
package annotation

import "ascend-common/api"

// NPU annotation keys
const (
	// NPUChipVdieIDsAnnotation annotation value is vdie ids (JSON)
	NPUChipVdieIDsAnnotation = api.ResourceNamePrefix + "npu.chip.vdie-ids"
	// NPUChipDdieIDsAnnotation annotation value is ddie ids (JSON)
	NPUChipDdieIDsAnnotation = api.ResourceNamePrefix + "npu.chip.ddie-ids"
	// NPUChipSerialNumberAnnotation annotation value is chip serial number (JSON)
	NPUChipSerialNumberAnnotation = api.ResourceNamePrefix + "npu.chip.serial-number"
	// NPUBaseDevInfosAnnotation annotation value is device base info (standardized key)
	NPUBaseDevInfosAnnotation = api.ResourceNamePrefix + "npu.base-device-infos"
	// NPUResetInfoAnnotation annotation value is reset info (standardized key)
	NPUResetInfoAnnotation = api.ResourceNamePrefix + "npu.reset-info"
	// NPUTopologyAnnotation node NPU chip topology declaration (NPU ID multi-level array).
	NPUTopologyAnnotation = api.ResourceNamePrefix + "npu.topology"
)

// Old annotation keys (deprecated, dual-write in Phase 1, Will be removed in Phase 2)
const (
	// Deprecated: NodeSNAnnotationDeprecated is the old node serial number annotation key,
	// replaced by NPUServerSerialNumberLabel. Will be removed in Phase 2.
	NodeSNAnnotationDeprecated = "product-serial-number"

	// Deprecated: BaseDevInfoAnnoDeprecated is the old device base info annotation key,
	// replaced by NPUBaseDevInfosAnnotation. Will be removed in Phase 2.
	BaseDevInfoAnnoDeprecated = "baseDeviceInfos"

	// Deprecated: ServerTypeKeyDeprecated is the old node server type annotation key,
	// replaced by label.NPUChipNameLabel and label.NPUChipBoardIDLabel.
	// Phase 1: compatible retention (fallback when chipname+boardid fails).
	// Will be removed in Phase 2.
	ServerTypeKeyDeprecated = "serverType"

	// Deprecated: ServerIndexKeyDeprecated is the old server index annotation key,
	// replaced by TopoLabelServerId label. Will be removed in Phase 2.
	ServerIndexKeyDeprecated = "serverIndex"

	// Deprecated: SuperPodIDKeyDeprecated is the old super pod ID annotation key,
	// replaced by topotree.superpodid label. Will be removed in Phase 2.
	SuperPodIDKeyDeprecated = "superPodID"

	// Deprecated: ResetInfoKeyDeprecated is the old reset info annotation key,
	// replaced by NPUResetInfoAnnotation. Will be removed in Phase 2.
	ResetInfoKeyDeprecated = "ResetInfo"

	// Deprecated: RackIDKeyDeprecated is the old rack ID annotation key,
	// replaced by topotree.rackid label. Will be removed in Phase 2.
	RackIDKeyDeprecated = "rackID"

	// Deprecated: CardTypeKeyDeprecated is the old card type annotation key.
	// Phase 1: keep writing, Will be removed in Phase 2.
	CardTypeKeyDeprecated = "cardType"
)
