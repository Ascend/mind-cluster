// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"fmt"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

func newUceAccompanyFaultProcessor(deviceCenter *deviceFaultProcessCenter) *uceAccompanyFaultProcessor {
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
	nodeName string, deviceInfo AdvanceDeviceFaultCm) {
	if _, ok := processor.uceAccompanyFaultQue[nodeName]; !ok {
		processor.uceAccompanyFaultQue[nodeName] = make(map[string][]constant.DeviceFault)
	}
	if _, ok := processor.uceFaultTime[nodeName]; !ok {
		processor.uceFaultTime[nodeName] = make(map[string]int64)
	}
	for deviceName, deviceFaults := range deviceInfo.FaultDeviceList {
		for _, fault := range deviceFaults {
			if isUceFault(fault) {
				errorMsg := fmt.Sprintf("uceAccompany cannot find uce fault time of device %s of node %s",
					deviceName, nodeName)
				processor.uceFaultTime[nodeName][deviceName] = getFaultTime(fault, errorMsg)
				continue
			}
			if !isUceAccompanyFault(fault) {
				continue
			}
			processor.inQue(nodeName, deviceName, fault)
		}
	}
}

func (processor *uceAccompanyFaultProcessor) inQue(nodeName, deviceName string, fault constant.DeviceFault) {
	if _, ok := processor.uceAccompanyFaultQue[nodeName][deviceName]; !ok {
		processor.uceAccompanyFaultQue[nodeName][deviceName] = make([]constant.DeviceFault, 0)
	}

	faultsInfo := processor.uceAccompanyFaultQue[nodeName][deviceName]
	found := false
	for _, anotherFault := range faultsInfo {
		if isDeviceFaultEqual(fault, anotherFault) {
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
		accompanyFaultTime := getFaultTime(fault, errorMsg)
		// if is accompanied fault, filter
		if processor.isAccompaniedFaultByUce(uceFaultTime, accompanyFaultTime) {
			hwlog.RunLog.Warnf("filter uce accompany fault %s, fault time: %s",
				util.ObjToString(fault), util.ReadableMsTime(accompanyFaultTime))
			faultMap = deleteFaultFromFaultMap(faultMap, fault)
			continue
		}
		// if current is not exceed diagnosis time,
		// then cannot decide fault is accompany or not, filter, and in que to decide in next turn.
		if !processor.isCurrentExceedDiagnosisTimeout(currentTime, accompanyFaultTime) {
			hwlog.RunLog.Warnf("filter uce accompany like fault %s, fault time: %s",
				util.ObjToString(fault), util.ReadableMsTime(accompanyFaultTime))
			faultMap = deleteFaultFromFaultMap(faultMap, fault)
			newDeviceFaultQue = append(newDeviceFaultQue, fault)
		}
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

func (processor *uceAccompanyFaultProcessor) process() {
	deviceInfos := processor.deviceCenter.getProcessingCm()
	processor.deviceCmForNodeMap = getAdvanceDeviceCmForNodeMap(deviceInfos)
	hwlog.RunLog.Debugf("current deviceInfos: %s", util.ObjToString(deviceInfos))
	hwlog.RunLog.Debugf("current deviceCmForNodeMap: %s", util.ObjToString(processor.deviceCmForNodeMap))

	processor.uceAccompanyFaultInQue()
	hwlog.RunLog.Debugf("current uceAccompanyFaultQue: %s", util.ObjToString(processor.uceAccompanyFaultQue))
	currentTime := time.Now().UnixMilli()

	processor.filterFaultInfos(currentTime)
	advanceDeviceCmForNodeMapToString(processor.deviceCmForNodeMap, deviceInfos)

	hwlog.RunLog.Debugf("uceAccompanyFaultProcessor result: %s", util.ObjToString(deviceInfos))
	processor.deviceCenter.setProcessingCm(deviceInfos)
}
