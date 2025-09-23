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

// Package l2fault is used to process l2 faults, when subscribed fault interface mindie job subscribed fault interface and using l2 fault npu,
// it will  l2 faults.
package l2fault

import (
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/l2fault"
	"clusterd/pkg/interface/common"
)

const (
	selfrecoverFaultTimeout = 10 * time.Second
)

// L2FaultProcessor is used to process l2 faults
var L2FaultProcessor *l2FaultProcessor

type l2FaultProcessor struct{}

func init() {
	L2FaultProcessor = &l2FaultProcessor{}
}

// Process is used to process l2 faults
func (processor *l2FaultProcessor) Process(info any) any {
	mindIeServerJobInfoMap, mindIeServerJobUsedDeviceMap := job.GetMindIeServerJobAndUsedDeviceInfoMap()
	if len(mindIeServerJobInfoMap) == 0 {
		hwlog.RunLog.Debug("no mindie server job info, skip l2 fault process")
		return info
	}
	if deviceContent, deviceOk := info.(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]); deviceOk {
		deletedL2FaultCmMap := processor.processDeviceFaults(deviceContent, mindIeServerJobInfoMap,
			mindIeServerJobUsedDeviceMap)
		l2fault.L2FaultCache.SetDeletedDevL2FaultCmForNodeMap(deletedL2FaultCmMap)
		return deviceContent
	}
	if switchContent, switchOK := info.(constant.OneConfigmapContent[*constant.SwitchInfo]); switchOK {
		deletedL2FaultCmMap := processor.processSwitchFaults(switchContent, mindIeServerJobInfoMap)
		l2fault.L2FaultCache.SetDeletedSwitchL2FaultCmForNodeMap(deletedL2FaultCmMap)
		return switchContent
	}
	return info
}

func (processor *l2FaultProcessor) processDeviceFaults(
	deviceContent constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm],
	mindIeServerJobInfoMap map[string]map[string]constant.JobInfo,
	mindIeServerJobUsedDeviceMap map[string]map[string]sets.String) map[string]*constant.AdvanceDeviceFaultCm {
	deletedL2FaultCmMap := make(map[string]*constant.AdvanceDeviceFaultCm)
	for nodeName, advanceDeviceFaultCm := range deviceContent.AllConfigmap {
		jobInfoMap, hasJobInfo := mindIeServerJobInfoMap[nodeName]
		jobUsedDeviceInfoMap, hasUsedDeviceInfo := mindIeServerJobUsedDeviceMap[nodeName]
		if !hasJobInfo || !hasUsedDeviceInfo {
			hwlog.RunLog.Debugf("nodeName: %s has no mindie server job info or used device info, "+
				"skip l2 fault process", nodeName)
			continue
		}
		hwlog.RunLog.Debugf("nodeName: %s current advanceDeviceFaultCm.FaultDeviceList: %v",
			nodeName, advanceDeviceFaultCm.FaultDeviceList)
		deletedL2FaultCm, err := copyAdvanceDeviceFaultCm(advanceDeviceFaultCm)
		if err != nil {
			continue
		}
		processor.collectAndRemoveDeviceFaults(advanceDeviceFaultCm, deletedL2FaultCm,
			jobInfoMap, jobUsedDeviceInfoMap)
		deletedL2FaultCmMap[nodeName] = deletedL2FaultCm
		hwlog.RunLog.Debugf("set nodeName: %s and device deletedL2FaultCm: %v to deletedL2FaultCmMap",
			nodeName, deletedL2FaultCm)
	}
	return deletedL2FaultCmMap
}

func (processor *l2FaultProcessor) processSwitchFaults(switchContent constant.OneConfigmapContent[*constant.SwitchInfo],
	mindIeServerJobInfoMap map[string]map[string]constant.JobInfo) map[string]*constant.SwitchInfo {
	deletedL2FaultCmMap := make(map[string]*constant.SwitchInfo)
	for cmName, switchInfo := range switchContent.AllConfigmap {
		hwlog.RunLog.Debugf("cmName: %s current switchInfo: %v", cmName, switchInfo)
		nodeName := strings.TrimPrefix(cmName, constant.SwitchInfoPrefix)
		if _, hasJobInfo := mindIeServerJobInfoMap[nodeName]; !hasJobInfo {
			hwlog.RunLog.Debugf("node %s (from cm %s) has no mindie server job info, skip l2 fault process",
				nodeName, cmName)
			continue
		}
		deletedSwitchInfo, err := copySwitchInfo(switchInfo)
		if err != nil {
			continue
		}
		deleteFaults := getDeletedSwitchL2Fault(switchInfo, mindIeServerJobInfoMap[nodeName])
		deletedSwitchInfo.FaultInfo = deleteFaults
		deletedL2FaultCmMap[cmName] = deletedSwitchInfo
		hwlog.RunLog.Debugf("set cmName: %s and switch deleteFaults: %v to deletedL2FaultCmMap",
			cmName, deleteFaults)
	}

	return deletedL2FaultCmMap
}

func copyAdvanceDeviceFaultCm(src *constant.AdvanceDeviceFaultCm) (*constant.AdvanceDeviceFaultCm, error) {
	dst := new(constant.AdvanceDeviceFaultCm)
	if err := util.DeepCopy(dst, src); err != nil {
		hwlog.RunLog.Errorf("deep copy AdvanceDeviceFaultCm failed: %v", err)
		return nil, err
	}
	return dst, nil
}

func (processor *l2FaultProcessor) collectAndRemoveDeviceFaults(src, dst *constant.AdvanceDeviceFaultCm,
	jobInfoMap map[string]constant.JobInfo, usedDeviceMap map[string]sets.String) {
	totalDeleteFaults := make([]constant.DeviceFault, 0)
	for deviceName, faults := range src.FaultDeviceList {
		deleteFaults := getDeletedDeviceL2Fault(faults, deviceName, jobInfoMap, usedDeviceMap)
		totalDeleteFaults = append(totalDeleteFaults, deleteFaults...)
		dst.FaultDeviceList[deviceName] = make([]constant.DeviceFault, 0, len(deleteFaults))
	}
	for _, fault := range totalDeleteFaults {
		src.DelFaultAndFix(fault)
		dst.AddFaultAndFix(fault)
	}
}

func copySwitchInfo(src *constant.SwitchInfo) (*constant.SwitchInfo, error) {
	dst := new(constant.SwitchInfo)
	if err := util.DeepCopy(dst, src); err != nil {
		hwlog.RunLog.Errorf("deep copy SwitchInfo failed: %v", err)
		return nil, err
	}
	return dst, nil
}

func shouldReportFault(faultTimeAndLevel constant.FaultTimeAndLevel, jobInfo constant.JobInfo,
	deviceName string, faultCode string) bool {
	if faultTimeAndLevel.FaultLevel != constant.RestartRequest {
		return true
	}

	durationMs := time.Now().UnixMilli() - faultTimeAndLevel.FaultTime
	if durationMs > selfrecoverFaultTimeout.Milliseconds() {
		hwlog.RunLog.Debugf("L2 fault %s during %dms more than 10s, should report fault", faultCode, durationMs)
		return true
	}

	if !common.Publisher.IsSubscribed(jobInfo.MultiInstanceJobId, constant.ControllerAppType) {
		hwlog.RunLog.Debugf("mindie job:%s not subscribed to grpc interface, should report fault:%s",
			jobInfo.Key, faultCode)
		return true
	}

	hwlog.RunLog.Infof("L2 fault %s during less than 10s, mindie job: %s has subscribed grpc interface and "+
		"using fault npu: %s, should not report fault", faultCode, jobInfo.Key, deviceName)
	return false
}

func getDeletedDeviceL2Fault(faults []constant.DeviceFault, deviceName string, jobInfoMap map[string]constant.JobInfo,
	jobUsedDeviceInfoMap map[string]sets.String) []constant.DeviceFault {
	deleteFaults := make([]constant.DeviceFault, 0, len(faults))
	for jobId, jobInfo := range jobInfoMap {
		jobUsedDeviceInfo, hasUsedDevices := jobUsedDeviceInfoMap[jobId]
		if !hasUsedDevices {
			hwlog.RunLog.Debugf("mindie server job %s has no used device info, report all faults", jobId)
			continue
		}
		for _, faultInfo := range faults {
			faultTimeAndLevel, hasTimeLevel := faultInfo.FaultTimeAndLevelMap[faultInfo.FaultCode]
			if !hasTimeLevel {
				hwlog.RunLog.Warnf("fault %s has no time and level info, report it", faultInfo.FaultCode)
				continue
			}
			if deviceName != "" && !jobUsedDeviceInfo.Has(deviceName) {
				hwlog.RunLog.Debugf("mindie job:%s does not use fault npu:%s, report fault:%v",
					jobInfo.Key, deviceName, faultInfo.FaultCode)
				continue
			}
			if !shouldReportFault(faultTimeAndLevel, jobInfo, deviceName, faultInfo.FaultCode) {
				deleteFaults = append(deleteFaults, faultInfo)
			}
		}
	}
	return deleteFaults
}

func getDeletedSwitchL2Fault(switchInfo *constant.SwitchInfo, jobInfoMap map[string]constant.JobInfo) []constant.
	SimpleSwitchFaultInfo {
	filteredFaults := make([]constant.SimpleSwitchFaultInfo, 0, len(switchInfo.FaultInfo))
	deletedFaults := make([]constant.SimpleSwitchFaultInfo, 0, len(switchInfo.FaultInfo))
	for _, jobInfo := range jobInfoMap {
		for _, faultInfo := range switchInfo.FaultInfo {
			faultTimeAndLevelKey := faultInfo.AssembledFaultCode + "_" + strconv.Itoa(int(faultInfo.SwitchChipId)) +
				"_" + strconv.Itoa(int(faultInfo.SwitchPortId))
			faultTimeAndLevel, ok := switchInfo.FaultTimeAndLevelMap[faultTimeAndLevelKey]
			if !ok {
				hwlog.RunLog.Warnf("switchInfo has no faultTimeAndLevel for faultTimeAndLevelKey:%s, "+
					"report fault:%v", faultTimeAndLevelKey, faultInfo)
				filteredFaults = append(filteredFaults, faultInfo)
				continue
			}
			if shouldReportFault(faultTimeAndLevel, jobInfo, "", faultInfo.AssembledFaultCode) {
				filteredFaults = append(filteredFaults, faultInfo)
			} else {
				delete(switchInfo.FaultTimeAndLevelMap, faultTimeAndLevelKey)
				deletedFaults = append(deletedFaults, faultInfo)
			}
		}
	}
	switchInfo.FaultInfo = filteredFaults
	if len(deletedFaults) > 0 && switchInfo.FaultLevel == constant.RestartRequest {
		switchInfo.FaultLevel = ""
		switchInfo.NodeStatus = constant.HealthyState
		if len(filteredFaults) > 0 {
			switchInfo.FaultLevel = constant.NotHandleFault
		}
		hwlog.RunLog.Debugf("update switchInfo FaultLevel=%s, NodeStatus=%s", switchInfo.FaultLevel,
			switchInfo.NodeStatus)
	}
	return deletedFaults
}
