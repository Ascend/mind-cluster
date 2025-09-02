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
	"strings"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
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
	deviceContent, deviceOk := info.(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	if deviceOk {
		for nodeName, advanceDeviceFaultCm := range deviceContent.AllConfigmap {
			hwlog.RunLog.Debugf("nodeName: %s current advanceDeviceFaultCm.FaultDeviceList: %v",
				nodeName, advanceDeviceFaultCm.FaultDeviceList)
			totalDeleteFaults := make([]constant.DeviceFault, 0)
			for deviceName, faults := range advanceDeviceFaultCm.FaultDeviceList {
				deleteFaults := dealDeviceL2Fault(faults, nodeName, deviceName)
				totalDeleteFaults = append(totalDeleteFaults, deleteFaults...)
			}
			for _, fault := range totalDeleteFaults {
				advanceDeviceFaultCm.DelFaultAndFix(fault)
			}
			deviceContent.AllConfigmap[nodeName] = advanceDeviceFaultCm
		}
		return deviceContent
	}
	switchContent, switchOK := info.(constant.OneConfigmapContent[*constant.SwitchInfo])
	if switchOK {
		for cmName, switchInfo := range switchContent.AllConfigmap {
			hwlog.RunLog.Debugf("cmName: %s current switchInfo: %v",
				cmName, switchInfo)
			nodeName := strings.TrimPrefix(cmName, constant.SwitchInfoPrefix)
			tmpFaults := dealSwitchL2Fault(switchInfo, nodeName)
			switchContent.AllConfigmap[cmName].FaultInfo = tmpFaults
		}
	}
	return switchContent
}

func shouldReportFault(level string, faultTime int64, jobId string, deviceName string, faultCode string) bool {
	if level != constant.RestartRequest {
		return true
	}

	during := time.Now().UnixMilli() - faultTime
	if during > selfrecoverFaultTimeout.Milliseconds() {
		hwlog.RunLog.Debugf("L2 fault during more than 10s, should report fault:%s", faultCode)
		return true
	}

	if jobId == "" {
		hwlog.RunLog.Debugf("mindie job does not exist, report fault")
		return true
	}

	if deviceName != "" && !job.IsJobUsedDevice(jobId, deviceName) {
		hwlog.RunLog.Debugf("mindie job:%s not uses fault npu:%s, should report fault:%s",
			jobId, deviceName, faultCode)
		return true
	}

	if !common.Publisher.IsSubscribed(jobId, constant.ControllerAppType) {
		hwlog.RunLog.Debugf("mindie job:%s does not subscribe grpc interface, should report fault:%s",
			jobId, faultCode)
		return true
	}

	hwlog.RunLog.Infof("L2 fault during less than 10s, mindie job: %s has subscribed grpc interface and "+
		"using fault npu: %s, should report fault: %s", jobId, deviceName, faultCode)
	return false
}

func dealDeviceL2Fault(faults []constant.DeviceFault,
	nodeName, deviceName string) []constant.DeviceFault {
	hwlog.RunLog.Debugf("dealDeviceL2Fault nodeName=%s, deviceName=%s", nodeName, deviceName)
	deleteFaults := make([]constant.DeviceFault, 0, len(faults))
	jobId := job.GetInferenceJobIdByNodeName(nodeName)
	for _, faultInfo := range faults {
		faultDetail, ok := faultInfo.FaultTimeAndLevelMap[faultInfo.FaultCode]
		if !ok {
			hwlog.RunLog.Warnf("faultInfo has no faultTimeAndLevel for faultCode: %v, report fault:%v",
				faultInfo.FaultCode, faultInfo)
			continue
		}

		if !shouldReportFault(faultDetail.FaultLevel, faultDetail.FaultTime, jobId, deviceName, faultInfo.FaultCode) {
			deleteFaults = append(deleteFaults, faultInfo)
		}
	}
	return deleteFaults
}

func dealSwitchL2Fault(switchInfo *constant.SwitchInfo,
	nodeName string) []constant.SimpleSwitchFaultInfo {
	hwlog.RunLog.Debugf("dealSwitchL2Fault nodeName=%s", nodeName)
	filteredFaults := make([]constant.SimpleSwitchFaultInfo, 0, len(switchInfo.FaultInfo))
	jobId := job.GetInferenceJobIdByNodeName(nodeName)
	for _, faultInfo := range switchInfo.FaultInfo {
		faultDetail, ok := switchInfo.FaultTimeAndLevelMap[faultInfo.AssembledFaultCode]
		if !ok {
			hwlog.RunLog.Warnf("switchInfo has no faultTimeAndLevel for faultCode:%s, report fault:%v",
				faultInfo.AssembledFaultCode, faultInfo)
			filteredFaults = append(filteredFaults, faultInfo)
			continue
		}

		if shouldReportFault(faultDetail.FaultLevel, faultDetail.FaultTime, jobId, "",
			faultInfo.AssembledFaultCode) {
			filteredFaults = append(filteredFaults, faultInfo)
		}
	}
	return filteredFaults
}
