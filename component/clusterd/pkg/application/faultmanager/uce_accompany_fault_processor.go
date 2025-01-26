// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"fmt"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/collector"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
)

func newUceAccompanyFaultProcessor(deviceCenter *DeviceFaultProcessCenter) *uceAccompanyFaultProcessor {
	return &uceAccompanyFaultProcessor{
		DiagnosisAccompanyTimeout: constant.DiagnosisAccompanyTimeout,
		uceAccompanyFaultQue:      make(map[string]map[string][]constant.DeviceFault),
		uceFaultTime:              make(map[string]map[string]int64),
		deviceCenter:              deviceCenter,
	}
}

func (processor *uceAccompanyFaultProcessor) uceAccompanyFaultInQue() {
	for nodeName, deviceInfo := range processor.deviceCmForNodeMap {
		processor.uceAccompanyFaultInQueForNode(nodeName, deviceInfo)
	}
}

func (processor *uceAccompanyFaultProcessor) uceAccompanyFaultInQueForNode(
	nodeName string, deviceInfo constant.AdvanceDeviceFaultCm) {
	if _, ok := processor.uceAccompanyFaultQue[nodeName]; !ok {
		processor.uceAccompanyFaultQue[nodeName] = make(map[string][]constant.DeviceFault)
	}
	if _, ok := processor.uceFaultTime[nodeName]; !ok {
		processor.uceFaultTime[nodeName] = make(map[string]int64)
	}
	for deviceName, deviceFaults := range deviceInfo.FaultDeviceList {
		for _, fault := range deviceFaults {
			// find uce fault in control plane
			if faultdomain.IsUceFault(fault.FaultCode) {
				errorMsg := fmt.Sprintf("uceAccompany cannot find uce fault time of device %s of node %s",
					deviceName, nodeName)
				processor.uceFaultTime[nodeName][deviceName] = faultdomain.GetFaultTime(fault, errorMsg)
				continue
			}
			// find uce fault in business plane
			if found, info := processor.isBusinessUceFault(nodeName, fault.NPUName); found {
				processor.uceFaultTime[nodeName][deviceName] = info.RecoverTime
			}
			if !faultdomain.IsUceAccompanyFault(fault.FaultCode) {
				continue
			}
			processor.inQue(nodeName, deviceName, fault)
		}
	}
}

func (processor *uceAccompanyFaultProcessor) isBusinessUceFault(nodeName, deviceName string) (bool, constant.ReportInfo) {
	info := collector.ReportInfoCollector.GetInfoWithoutJobId(nodeName, deviceName)
	if info.RecoverTime != constant.JobNotRecover {
		return true, info
	}
	return false, constant.ReportInfo{}
}

func (processor *uceAccompanyFaultProcessor) inQue(nodeName, deviceName string, fault constant.DeviceFault) {
	if _, ok := processor.uceAccompanyFaultQue[nodeName][deviceName]; !ok {
		processor.uceAccompanyFaultQue[nodeName][deviceName] = make([]constant.DeviceFault, 0)
	}

	faultsInfo := processor.uceAccompanyFaultQue[nodeName][deviceName]
	found := false
	for _, anotherFault := range faultsInfo {
		if faultdomain.IsDeviceFaultEqual(fault, anotherFault) {
			found = true
			break
		}
	}
	if !found {
		processor.uceAccompanyFaultQue[nodeName][deviceName] = append(faultsInfo, fault)
	}
}

func (processor *uceAccompanyFaultProcessor) filterFaultInfos(currentTime int64) {
	for nodeName, nodeFaults := range processor.uceAccompanyFaultQue {
		faultMap := processor.deviceCmForNodeMap[nodeName]
		for deviceName, deviceFaultQue := range nodeFaults {
			newQue, newFaultMap :=
				processor.filterFaultDevice(faultMap.FaultDeviceList, currentTime, nodeName, deviceName, deviceFaultQue)
			nodeFaults[deviceName] = newQue
			faultMap.FaultDeviceList = newFaultMap
		}
		processor.deviceCmForNodeMap[nodeName] = faultMap
	}
}

func (processor *uceAccompanyFaultProcessor) filterFaultDevice(
	faultMap map[string][]constant.DeviceFault, currentTime int64, nodeName, deviceName string,
	deviceFaultQue []constant.DeviceFault) ([]constant.DeviceFault, map[string][]constant.DeviceFault) {
	newDeviceFaultQue := make([]constant.DeviceFault, 0)
	for _, fault := range deviceFaultQue {
		uceFaultTime := processor.getDeviceUceFaultTime(nodeName, deviceName)
		errorMsg := fmt.Sprintf("filterFaultDevice cannot find uce fault time for device %s of node %s",
			deviceName, nodeName)
		accompanyFaultTime := faultdomain.GetFaultTime(fault, errorMsg)
		// if is accompanied fault, filter
		if processor.isAccompaniedFaultByUce(uceFaultTime, accompanyFaultTime) {
			hwlog.RunLog.Warnf("filter uce accompany fault %s, fault time: %s",
				util.ObjToString(fault), util.ReadableMsTime(accompanyFaultTime))
			faultMap = faultdomain.DeleteFaultFromFaultMap(faultMap, fault)
			continue
		}
		// if current is not exceed diagnosis time,
		// then cannot decide fault is accompany or not, filter, and in que to decide in next turn.
		if !processor.isCurrentExceedDiagnosisTimeout(currentTime, accompanyFaultTime) {
			hwlog.RunLog.Warnf("filter uce accompany like fault %s, fault time: %s",
				util.ObjToString(fault), util.ReadableMsTime(accompanyFaultTime))
			faultMap = faultdomain.DeleteFaultFromFaultMap(faultMap, fault)
			newDeviceFaultQue = append(newDeviceFaultQue, fault)
			continue
		}
		// cannot filter, add the aic/aiv fault into faultMap
		faultMap = faultdomain.AddFaultIntoFaultMap(faultMap, fault)
		hwlog.RunLog.Warnf("cannot filter uce accompany like fault %s, uce fault time: %s",
			util.ObjToString(fault), util.ReadableMsTime(uceFaultTime))
	}
	return newDeviceFaultQue, faultMap
}

func (processor *uceAccompanyFaultProcessor) getDeviceUceFaultTime(nodeName, deviceName string) int64 {
	if faultTime, ok := processor.uceFaultTime[nodeName][deviceName]; ok {
		return faultTime
	}
	return constant.DeviceNotFault
}

func (processor *uceAccompanyFaultProcessor) isAccompaniedFaultByUce(
	uceFaultTime, uceAccompanyFaultTime int64) bool {
	return util.Abs(uceFaultTime-uceAccompanyFaultTime) <= processor.DiagnosisAccompanyTimeout
}

func (processor *uceAccompanyFaultProcessor) isCurrentExceedDiagnosisTimeout(
	currentTime, uceAccompanyFaultTime int64) bool {
	return uceAccompanyFaultTime < currentTime-processor.DiagnosisAccompanyTimeout
}

func (processor *uceAccompanyFaultProcessor) Process(info any) any {
	processContent, ok := info.(constant.CenterProcessContent[*constant.DeviceInfo])
	if !ok {
		hwlog.RunLog.Errorf("%v cannot convert to DeviceInfo", info)
		return info
	}
	if len(processContent.UpdateConfigmap) == 0 && len(processor.uceAccompanyFaultQue) == 0 {
		return info
	}
	deviceInfos := processContent.AllConfigmap
	processor.deviceCmForNodeMap = faultdomain.GetAdvanceDeviceCmForNodeMap(deviceInfos)
	hwlog.RunLog.Debugf("current deviceInfos: %s", util.ObjToString(deviceInfos))
	hwlog.RunLog.Debugf("current deviceCmForNodeMap: %s", util.ObjToString(processor.deviceCmForNodeMap))

	processor.uceAccompanyFaultInQue()
	hwlog.RunLog.Debugf("current uceAccompanyFaultQue: %s", util.ObjToString(processor.uceAccompanyFaultQue))
	currentTime := time.Now().UnixMilli()

	processor.filterFaultInfos(currentTime)
	faultdomain.AdvanceDeviceCmForNodeMapToString(processor.deviceCmForNodeMap, deviceInfos)

	hwlog.RunLog.Debugf("uceAccompanyFaultProcessor result: %s", util.ObjToString(deviceInfos))
	processContent.AllConfigmap = deviceInfos
	return processContent
}
