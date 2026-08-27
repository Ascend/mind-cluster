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

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/next/devicefactory/customname"
	"ascend-common/api"
	"ascend-common/api/annotation"
	"ascend-common/api/label"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/dcmi"
)

// --- Labelers (implement label.NodeLabeler) ---

// chipNameLabeler writes chip name labels (new + old dual-write).
type chipNameLabeler struct {
	hdm *HwDevManager
}

func (l *chipNameLabeler) Write(labels map[string]string, ctx *label.NodeContext) error {
	if ctx.IsHeterogeneous {
		return nil
	}
	chipInfo, err := l.hdm.manager.GetDmgr().GetValidChipInfo()
	if err != nil {
		hwlog.RunLog.Warnf("failed to get valid chip info for chip.name, skip chip.name label, err: %v", err)
		return err
	}
	writeValue(labels, chipInfo.Name, label.NPUChipNameLabel, label.NPUChipNameLabelDeprecated)
	return nil
}

// chipMemoryLabeler writes chip memory labels (new + old dual-write).
// Skipped when heterogeneous chips are present.
type chipMemoryLabeler struct {
	hdm *HwDevManager
}

func (l *chipMemoryLabeler) Write(labels map[string]string, ctx *label.NodeContext) error {
	if ctx.IsHeterogeneous {
		return nil
	}
	if !common.HasOnChipMemory() {
		hwlog.RunLog.Debugf("current device does't have high bandwidth memory , skip chip.memory label")
		return nil
	}
	if len(l.hdm.allInfo.AllDevs) == 0 {
		hwlog.RunLog.Warnf("no devices found, skip chip.memory label")
		return nil
	}
	hbmInfo, err := l.hdm.manager.GetDmgr().GetDeviceHbmInfo(l.hdm.allInfo.AllDevs[common.FirstDevice].LogicID)
	if err != nil {
		hwlog.RunLog.Warnf("failed to get node on-chip-memory info, skip chip.memory label, err: %s", err)
		return nil
	}
	memoryValue := fmt.Sprintf("%dG", hbmInfo.MemorySize/memoryRadix)
	writeValue(labels, memoryValue, label.NPUChipMemoryLabel, label.NPUChipMemoryLabelDeprecated)
	return nil
}

// chipBoardIDLabeler writes chip board ID label.
// Skipped when heterogeneous chips are present.
type chipBoardIDLabeler struct {
	hdm *HwDevManager
}

func (l *chipBoardIDLabeler) Write(labels map[string]string, ctx *label.NodeContext) error {
	if ctx.IsHeterogeneous {
		return nil
	}
	if len(l.hdm.allInfo.AllDevs) == 0 {
		hwlog.RunLog.Warnf("no devices found, skip chip.boardid label")
		return nil
	}
	boardInfo, err := l.hdm.manager.GetDmgr().GetBoardInfo(l.hdm.allInfo.AllDevs[common.FirstDevice].LogicID)
	if err != nil {
		hwlog.RunLog.Warnf("failed to get board info for chip.boardid, skip chip.boardid label, err: %s", err)
		return nil
	}
	if boardInfo.BoardId == common.EmptyBoardId {
		hwlog.RunLog.Warnf("board id is empty, skip chip.boardid label")
		return nil
	}
	boardIDValue := fmt.Sprintf("0x%02x", boardInfo.BoardId)
	writeValue(labels, boardIDValue, label.NPUChipBoardIDLabel)
	return nil
}

// chipProductTypeLabeler writes chip product type label (replaces infer-card-type).
// Skipped when heterogeneous chips are present.
type chipProductTypeLabeler struct {
	hdm *HwDevManager
}

func (l *chipProductTypeLabeler) Write(labels map[string]string, ctx *label.NodeContext) error {
	if ctx.IsHeterogeneous {
		return nil
	}
	productType := l.hdm.getProductType()
	if productType == "" || productType == "NA" {
		hwlog.RunLog.Warnf("product type is empty or NA, skip chip.product-type label")
		return nil
	}
	writeValue(labels, sanitizeLabelValue(productType), label.NPUChipProductTypeLabel)
	if common.IsContainAll300IDuo() {
		writeValue(labels, api.A300IDuoLabel, label.NPUChipProductTypeLabelDeprecated)
	}
	return nil
}

// acceleratorTypeLabeler writes accelerator-type label for 910B infer cards.
// Skipped when heterogeneous chips are present.
type acceleratorTypeLabeler struct {
	hdm *HwDevManager
}

func (l *acceleratorTypeLabeler) Write(labels map[string]string, ctx *label.NodeContext) error {
	if common.ParamOption.RealCardType != api.Ascend910B {
		return nil
	}
	if len(l.hdm.allInfo.AllDevs) == 0 {
		hwlog.RunLog.Warnf("no devices found, skip accelerator-type label")
		return nil
	}
	boardInfo, err := l.hdm.manager.GetDmgr().GetBoardInfo(l.hdm.allInfo.AllDevs[common.FirstDevice].LogicID)
	if err != nil {
		hwlog.RunLog.Warnf("failed to get board info for accelerator-type, skip accelerator-type label, err: %s", err)
		return nil
	}
	if boardInfo.BoardId == common.A300IA2BoardId || boardInfo.BoardId == common.A300IA2GB64BoardId {
		writeValue(labels, label.A300IA2Label, label.AcceleratorTypeKeyDeprecated)
	}
	return nil
}

// serverTypeLabeler writes server type labels (new + old dual-write).
type serverTypeLabeler struct {
	hdm *HwDevManager
}

func (l *serverTypeLabeler) Write(labels map[string]string, ctx *label.NodeContext) error {
	cardType := common.ParamOption.RealCardType + common.MiddelLine +
		strconv.Itoa(int(common.ParamOption.AiCoreCount))
	if !customname.IsOldDeviceType(common.ParamOption.RealCardType) {
		serverTypeValue := api.AscendMinuxPrefix + strconv.Itoa(int(common.ParamOption.AiCoreCount))
		writeValue(labels, serverTypeValue, label.NPUServerTypeLabel, label.NPUServerTypeLabelDeprecated)
	} else {
		newVal, ok := label.GetNodeLabel(ctx.Node, label.NPUServerTypeLabel)
		oldVal, okOld := label.GetNodeLabel(ctx.Node, label.NPUServerTypeLabelDeprecated)
		if !ok && !okOld {
			value := customname.ReplaceDevicePublicName(l.hdm.RunMode, cardType)
			writeValue(labels, value, label.NPUServerTypeLabel, label.NPUServerTypeLabelDeprecated)
		} else if !ok {
			writeValue(labels, oldVal, label.NPUServerTypeLabel)
		} else if !okOld {
			writeValue(labels, newVal, label.NPUServerTypeLabelDeprecated)
		}

	}
	return nil
}

// driverVersionLabeler writes driver version labels (new + old dual-write).
type driverVersionLabeler struct {
	hdm *HwDevManager
}

func (l *driverVersionLabeler) Write(labels map[string]string, ctx *label.NodeContext) error {
	driverVersion := l.hdm.manager.GetDmgr().GetDcmiVersion()
	if driverVersion == "" {
		hwlog.RunLog.Warnf("failed to get dcmi driver version, skip driver.version label")
		return nil
	}
	writeValue(labels, driverVersion, label.NPUDriverVersionLabel, label.NPUDriverVersionLabelDeprecated)
	return nil
}

// acceleratorLabeler writes the accelerator label (not migrated to new prefix).
// Consumers derive chipKind from chip.name instead. Will be removed in Phase 2.
type acceleratorLabeler struct {
	hdm *HwDevManager
}

func (l *acceleratorLabeler) Write(labels map[string]string, ctx *label.NodeContext) error {
	if v, ok := acceleratorLabelMap[common.ParamOption.RealCardType]; ok {
		writeValue(labels, v, label.AcceleratorLabelKeyDeprecated)
	}
	return nil
}

// topologyLabeler writes topology-related labels (superPodId, rackId, serverId).
type topologyLabeler struct {
	hdm *HwDevManager
}

func (l *topologyLabeler) Write(labels map[string]string, ctx *label.NodeContext) error {
	if common.ParamOption.RealCardType == api.Ascend910A3 {
		superPodId := l.hdm.manager.GetSuperPodID()
		if int(superPodId) >= 0 {
			hwlog.RunLog.Infof("A3 device add superid label: %d", superPodId)
			writeValue(labels, strconv.Itoa(int(superPodId)), label.TopoLabelSuperPodId)
		} else {
			hwlog.RunLog.Warnf("A3 device superPodId is invalid: %d, skip topotree.superpodid label", superPodId)
		}
	}
	if common.ParamOption.RealCardType == api.Ascend910A5 {
		superPodId := l.hdm.manager.GetSuperPodID()
		if int(superPodId) >= 0 {
			hwlog.RunLog.Infof("npu device add superid label: %d", superPodId)
			writeValue(labels, strconv.Itoa(int(superPodId)), label.TopoLabelSuperPodId)
		} else {
			hwlog.RunLog.Warnf("npu device superPodId is invalid: %d, skip topotree.superpodid label", superPodId)
		}
		superPodType := l.hdm.manager.GetSuperPodType()
		if superPodType == common.ProductType1D || superPodType == common.ProductType2D {
			rackId := l.hdm.manager.GetRackID()
			if int(rackId) >= 0 {
				hwlog.RunLog.Infof("npu device add rackid label: %d", rackId)
				writeValue(labels, strconv.Itoa(int(rackId)), label.TopoLabelRackId)
			} else {
				hwlog.RunLog.Warnf("npu device rackId is invalid: %d, skip topotree.rackid label", rackId)
			}
		}
		serverIndex := l.hdm.manager.GetServerIndex()
		if int(serverIndex) >= 0 {
			hwlog.RunLog.Infof("npu device add serverid label: %d", serverIndex)
			writeValue(labels, strconv.Itoa(int(serverIndex)), label.TopoLabelServerId)
		} else {
			hwlog.RunLog.Warnf("npu device serverIndex is invalid: %d, skip topotree.serverid label", serverIndex)
		}
	}
	return nil
}

// --- Annotators (implement annotation.NodeAnnotator) ---

// baseInfoAnnotator writes base device info annotation (old + new dual-write).
type baseInfoAnnotator struct {
	hdm *HwDevManager
}

func (a *baseInfoAnnotator) Write(annotations map[string]string, ctx *label.NodeContext) error {
	baseInfo := a.hdm.getNpuBaseInfo()
	mashaledNpuInfo, err := json.Marshal(baseInfo)
	if err != nil {
		hwlog.RunLog.Warnf("failed to marshal base device info, skip baseDeviceInfos annotation, err: %v", err)
		return fmt.Errorf("failed to marshal device ip map: %w", err)
	}
	a.hdm.baseNPUInfo = baseInfo
	newMashaledNpuInfo := customname.ReplaceDevicePublicName(a.hdm.RunMode, string(mashaledNpuInfo))
	writeValue(annotations, newMashaledNpuInfo, annotation.BaseDevInfoAnnoDeprecated,
		annotation.NPUBaseDevInfosAnnotation)
	return nil
}

// vdieDdieAnnotator writes vdie-ids and ddie-ids annotations.
type vdieDdieAnnotator struct {
	hdm *HwDevManager
}

func (a *vdieDdieAnnotator) Write(annotations map[string]string, ctx *label.NodeContext) error {
	dieIDMaps := a.hdm.getDieIDAnnotations()
	vdieIDs := dieIDMaps[dcmi.VDIE]
	ddieIDs := dieIDMaps[dcmi.DDIE]
	if len(vdieIDs) > 0 {
		vdieJSON, err := json.Marshal(vdieIDs)
		if err != nil {
			hwlog.RunLog.Warnf("failed to marshal vdie-ids, skip vdie-ids annotation, err: %v", err)
		} else {
			writeValue(annotations, string(vdieJSON), annotation.NPUChipVdieIDsAnnotation)
		}
	}
	if len(ddieIDs) > 0 {
		ddieJSON, err := json.Marshal(ddieIDs)
		if err != nil {
			hwlog.RunLog.Warnf("failed to marshal ddie-ids, skip ddie-ids annotation, err: %v", err)
		} else {
			writeValue(annotations, string(ddieJSON), annotation.NPUChipDdieIDsAnnotation)
		}
	}
	return nil
}

// serialNumberAnnotator writes chip serial number annotation.
type serialNumberAnnotator struct {
	hdm *HwDevManager
}

func (a *serialNumberAnnotator) Write(annotations map[string]string, ctx *label.NodeContext) error {
	serialNumbers := a.hdm.getChipSerialNumbers()
	if len(serialNumbers) == 0 {
		hwlog.RunLog.Warnf("no chip serial numbers found, skip chip.serial-number annotation")
		return nil
	}
	snJSON, err := json.Marshal(serialNumbers)
	if err != nil {
		hwlog.RunLog.Warnf("failed to marshal chip serial numbers, skip chip.serial-number annotation, err: %v", err)
		return nil
	}
	writeValue(annotations, string(snJSON), annotation.NPUChipSerialNumberAnnotation)
	return nil
}

// serverTypeAnnotator writes server-type annotation for backward compatibility.
type serverTypeAnnotator struct {
	hdm *HwDevManager
}

func (a *serverTypeAnnotator) Write(annotations map[string]string, ctx *label.NodeContext) error {
	writeValue(annotations, getDevType(common.ParamOption.RealCardType), annotation.ServerTypeKeyDeprecated)
	return nil
}

// superPodInfoAnnotator writes superPodID, serverIndex, and rackID annotations
// (Phase 1: keep writing, Will be removed in Phase 2).
type superPodInfoAnnotator struct {
	hdm *HwDevManager
}

func (a *superPodInfoAnnotator) Write(annotations map[string]string, ctx *label.NodeContext) error {
	writeValue(annotations, strconv.Itoa(int(a.hdm.manager.GetSuperPodID())), annotation.SuperPodIDKeyDeprecated)
	writeValue(annotations, strconv.Itoa(int(a.hdm.manager.GetServerIndex())), annotation.ServerIndexKeyDeprecated)
	if common.ParamOption.RealCardType == api.Ascend910A5 {
		superPodType := a.hdm.manager.GetSuperPodType()
		if superPodType == common.ProductType1D || superPodType == common.ProductType2D {
			writeValue(annotations, strconv.Itoa(int(a.hdm.manager.GetRackID())), annotation.RackIDKeyDeprecated)
		}
	}
	return nil
}

// cardTypeAnnotator writes cardType annotation (Phase 1: keep writing, Will be removed in Phase 2).
type cardTypeAnnotator struct {
	hdm *HwDevManager
}

func (a *cardTypeAnnotator) Write(annotations map[string]string, ctx *label.NodeContext) error {
	cardType, err := a.hdm.getCardType()
	if err != nil {
		hwlog.RunLog.Errorf("failed to get node board info, err: %v", err)
	}
	if cardType != "" {
		common.ParamOption.CardType = cardType
		writeValue(annotations, common.ParamOption.CardType, annotation.CardTypeKeyDeprecated)
	} else {
		hwlog.RunLog.Warnf("card type is empty, skip cardType annotation")
	}
	return nil
}

// writeValue writes value to the map for all given keys, skipping empty values.
func writeValue(m map[string]string, value string, keys ...string) {
	if value == "" {
		hwlog.RunLog.Infof("skip writing empty value for keys: %v", keys)
		return
	}
	for _, key := range keys {
		m[key] = value
	}
}

// sanitizeLabelValue sanitizes a label value to conform to K8s label value regex.
func sanitizeLabelValue(value string) string {
	invalidRegex := regexp.MustCompile(common.LabelSanitizeRegex)
	sanitized := invalidRegex.ReplaceAllString(value, "")
	spaceRegex := regexp.MustCompile(` +`)
	sanitized = spaceRegex.ReplaceAllString(sanitized, "-")
	return sanitized
}
