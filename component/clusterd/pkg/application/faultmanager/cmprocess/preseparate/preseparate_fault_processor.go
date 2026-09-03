/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package preseparate is used to process preseparate faults
package preseparate

import (
	"strings"

	"k8s.io/utils/strings/slices"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/pod"
)

// PreSeparateFaultProcessor is used to process preseparate faults
var PreSeparateFaultProcessor *preSeparateFaultProcessor

type preSeparateFaultProcessor struct {
	nodeStatusMap map[string]string
}

func init() {
	PreSeparateFaultProcessor = &preSeparateFaultProcessor{
		nodeStatusMap: make(map[string]string),
	}
}

// Process is used to process preseparate faults
func (processor *preSeparateFaultProcessor) Process(info any) any {
	if deviceContent, ok := info.(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]); ok {
		for nodeName, faultCm := range deviceContent.AllConfigmap {
			processor.processDeviceFaultCm(faultCm, nodeName)
		}
		return deviceContent
	}
	if switchContent, ok := info.(constant.OneConfigmapContent[*constant.SwitchInfo]); ok {
		for cmName, switchInfo := range switchContent.AllConfigmap {
			processor.updateNodeStatusBySwitchInfo(cmName, switchInfo)
		}
		cleanNodeStatusMap(processor.nodeStatusMap, switchContent.UpdateConfigmap)
		return switchContent
	}
	if nodeContent, ok := info.(constant.OneConfigmapContent[*constant.NodeInfo]); ok {
		for cmName, nodeInfo := range nodeContent.AllConfigmap {
			processor.updateNodeStatusByNodeInfo(cmName, nodeInfo)
		}
		cleanNodeStatusMap(processor.nodeStatusMap, nodeContent.UpdateConfigmap)
		return nodeContent
	}
	return info
}

// cleanNodeStatusMap deletes the node status in nodeStatusMap when the configmap is deleted
func cleanNodeStatusMap[T constant.ConfigMapInterface](nodeStatusMap map[string]string,
	updateItems []constant.InformerCmItem[T]) {
	for _, item := range updateItems {
		if !item.IsAdd {
			delete(nodeStatusMap, item.Data.GetCmName())
		}
	}
}

func (processor *preSeparateFaultProcessor) processDeviceFaultCm(
	faultCm *constant.AdvanceDeviceFaultCm, nodeName string) {
	for deviceName, faults := range faultCm.FaultDeviceList {
		for _, faultInfo := range faults {
			if constant.ParameterPlaneFaultCodes.Has(faultInfo.FaultCode) ||
				faultInfo.FaultLevel != constant.PreSeparateNPU {
				continue
			}
			processor.updateCardUnHealthy(faultCm, nodeName, deviceName)
			if slices.Contains(faultCm.AvailableDeviceList, deviceName) {
				hwlog.RunLog.Debugf("delete deviceName: %s from AvailableDeviceList", deviceName)
				faultCm.AvailableDeviceList = util.DeleteStringSliceItem(faultCm.AvailableDeviceList, deviceName)
			}
		}
	}
}

func (processor *preSeparateFaultProcessor) updateCardUnHealthy(
	faultCm *constant.AdvanceDeviceFaultCm, nodeName, deviceName string) {
	if usedDevices := pod.GetUsedDevicesByNodeName(nodeName); usedDevices.Has(deviceName) {
		hwlog.RunLog.Debugf("%s deviceName: %s is used now, delete it from CardUnHealthy", nodeName, deviceName)
		if slices.Contains(faultCm.CardUnHealthy, deviceName) {
			faultCm.CardUnHealthy = util.DeleteStringSliceItem(faultCm.CardUnHealthy, deviceName)
		}
		return
	}
	hwlog.RunLog.Debugf("deviceName: %s is not used now, add it to CardUnHealthy", deviceName)
	if !slices.Contains(faultCm.CardUnHealthy, deviceName) {
		faultCm.CardUnHealthy = append(faultCm.CardUnHealthy, deviceName)
	}
}

func (processor *preSeparateFaultProcessor) updateNodeStatusBySwitchInfo(cmName string, faultCm *constant.SwitchInfo) {
	if faultCm.NodeStatus == constant.UnHealthyState {
		return
	}
	nodeName := strings.TrimPrefix(cmName, constant.SwitchInfoPrefix)
	if faultCm.FaultLevel != constant.PreSeparateFaultLevelStr {
		return
	}
	usedDevices := pod.GetUsedDevicesByNodeName(nodeName)
	hasVolcanoTask := pod.HasVolcanoTaskOnNode(nodeName)
	if usedDevices.Len() > 0 || hasVolcanoTask {
		faultCm.NodeStatus = constant.HealthyState
	} else {
		faultCm.NodeStatus = constant.UnHealthyState
	}
	processor.logNodeStatusChanged(cmName, nodeName, faultCm.NodeStatus, faultCm.FaultLevel,
		usedDevices.Len(), hasVolcanoTask)
}

func (processor *preSeparateFaultProcessor) updateNodeStatusByNodeInfo(cmName string, faultCm *constant.NodeInfo) {
	if faultCm.NodeStatus == constant.UnHealthyState {
		return
	}
	nodeName := strings.TrimPrefix(cmName, constant.NodeInfoPrefix)
	faultLevels := make([]string, 0, len(faultCm.FaultDevList))
	for _, faultDev := range faultCm.FaultDevList {
		faultLevels = append(faultLevels, faultDev.FaultLevel)
	}
	mostSeriousLevel := faultdomain.GetNodeMostSeriousFaultLevel(faultLevels)

	if mostSeriousLevel != constant.PreSeparateFault {
		return
	}

	usedDevices := pod.GetUsedDevicesByNodeName(nodeName)
	hasVolcanoTask := pod.HasVolcanoTaskOnNode(nodeName)
	if usedDevices.Len() > 0 || hasVolcanoTask {
		faultCm.NodeStatus = constant.HealthyState
	} else {
		faultCm.NodeStatus = constant.UnHealthyState
	}
	processor.logNodeStatusChanged(cmName, nodeName, faultCm.NodeStatus, mostSeriousLevel, usedDevices.Len(),
		hasVolcanoTask)
}

// logNodeStatusChanged logs the node status transition only when the status changes,
// to avoid log flooding caused by the periodic processing of the same configmap.
func (processor *preSeparateFaultProcessor) logNodeStatusChanged(cmName, nodeName, newStatus, faultLevel string,
	usedDevicesNum int, hasVolcanoTask bool) {
	if processor.nodeStatusMap == nil {
		processor.nodeStatusMap = make(map[string]string)
	}
	oldStatus, exist := processor.nodeStatusMap[cmName]
	if exist && oldStatus == newStatus {
		return
	}
	processor.nodeStatusMap[cmName] = newStatus
	hwlog.RunLog.Infof("node %s nodeStatus %s -> %s (level: %s) usedDevices:%d volcanoTask:%t",
		nodeName, oldStatus, newStatus, faultLevel, usedDevicesNum, hasVolcanoTask)
}
