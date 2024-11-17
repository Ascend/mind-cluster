/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
 */

// Package faultshoot contain fault process
package faultshoot

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

func getNodeAndDeviceFromJobIdAndRankId(
	jobId, rankId string, jobServerInfoMap job.JobServerInfoMap) (string, string, error) {
	for _, server := range jobServerInfoMap.InfoMap[jobId] {
		for _, dev := range server.DeviceList {
			if dev.RankID == rankId {
				return server.ServerName, dev.DeviceID, nil
			}
		}
	}
	return "", "", fmt.Errorf("not find node and device from jobId %v and rankid %v", jobId, rankId)
}

func getNodesNameFromDeviceInfo(deviceInfos map[string]*constant.DeviceInfo) []string {
	nodesName := make([]string, 0)
	for cmName, _ := range deviceInfos {
		nodeName := cmNameToNodeName(cmName)
		nodesName = append(nodesName, nodeName)
	}
	return nodesName
}

func cmNameToNodeName(cmName string) string {
	if !strings.HasPrefix(cmName, constant.DeviceInfoPrefix) {
		hwlog.RunLog.Errorf("CmName %s has not prefix %s", cmName, constant.DeviceInfoPrefix)
		return cmName
	}
	return strings.TrimPrefix(cmName, constant.DeviceInfoPrefix)
}

func nodeNameToCmName(nodeName string) string {
	return constant.DeviceInfoPrefix + nodeName
}

func getAdvanceDeviceCmForNodeMap(nodeDeviceInfoMap map[string]*constant.DeviceInfo) map[string]AdvanceDeviceCm {
	advanceDeviceCmForNodeMap := make(map[string]AdvanceDeviceCm)
	for _, deviceInfo := range nodeDeviceInfoMap {
		advanceDeviceCmForNodeMap[cmNameToNodeName(deviceInfo.CmName)] = getAdvanceDeviceCm(deviceInfo)
	}
	return advanceDeviceCmForNodeMap
}

// deviceName->faults
func getAdvanceDeviceCm(devInfo *constant.DeviceInfo) AdvanceDeviceCm {
	advanceDeviceCm := AdvanceDeviceCm{
		CmName:      devInfo.CmName,
		SuperPodID:  devInfo.SuperPodID,
		ServerIndex: devInfo.ServerIndex,
		UpdateTime:  devInfo.UpdateTime,
		ServerType:  getServerType(devInfo),
	}
	if devInfo == nil {
		hwlog.RunLog.Error("get fault list for node failed. devInfo is nil")
		return advanceDeviceCm
	}
	if devInfo.DeviceList == nil {
		hwlog.RunLog.Errorf("get fault list for node %v failed. device list does not exist", devInfo.CmName)
		return advanceDeviceCm
	}
	if faultList, ok := devInfo.DeviceList[getFaultListKey(devInfo)]; ok {
		var devicesFault []constant.DeviceFault
		err := json.Unmarshal([]byte(faultList), &devicesFault)
		if err != nil {
			hwlog.RunLog.Errorf("get fault list for node %v failed. "+
				"Json unmarshall exception: %v", devInfo.CmName, err)
			return advanceDeviceCm
		}
		deviceFaultMap := make(map[string][]constant.DeviceFault)
		for _, deviceFault := range devicesFault {
			if _, ok := deviceFaultMap[deviceFault.NPUName]; !ok {
				deviceFaultMap[deviceFault.NPUName] = make([]constant.DeviceFault, 0)
			}
			// device plugin may merge multiple fault codes in one string
			deviceFaults := splitDeviceFault(deviceFault)
			deviceFaultMap[deviceFault.NPUName] = append(deviceFaultMap[deviceFault.NPUName], deviceFaults...)
		}
		advanceDeviceCm.DeviceList = deviceFaultMap
	} else {
		hwlog.RunLog.Infof("get fault list for node %v failed. fault list does not exist", devInfo.CmName)
	}
	if networkUnhealthyCardList, ok := devInfo.DeviceList[getNetworkUnhealthyKey(devInfo)]; ok {
		cardList := strings.Split(networkUnhealthyCardList, ",")
		advanceDeviceCm.NetworkUnhealthy = cardList
	} else {
		hwlog.RunLog.Infof("get NetworkUnhealthy list for node %v failed. fault list does not exist",
			devInfo.CmName)
	}
	if cardUnhealthyCardList, ok := devInfo.DeviceList[getCardUnhealthyKey(devInfo)]; ok {
		cardList := strings.Split(cardUnhealthyCardList, ",")
		advanceDeviceCm.CarUnHealthy = cardList
	}
	return advanceDeviceCm
}

func getServerType(devInfo *constant.DeviceInfo) string {
	for key, _ := range devInfo.DeviceList {
		if strings.Contains(key, "Ascend910") {
			return "Ascend910"
		}
		if strings.Contains(key, "Ascend310") {
			return "Ascend310"
		}
	}
	hwlog.RunLog.Errorf("cannot decide server type")
	return "Ascend910"
}

// device plugin may merge multiple fault codes in one string
func splitDeviceFault(faultInfo constant.DeviceFault) []constant.DeviceFault {
	deviceFaults := make([]constant.DeviceFault, 0)
	codes := strings.Split(faultInfo.FaultCode, ",")
	for _, code := range codes {
		newFault := constant.DeviceFault{
			FaultType:            faultInfo.FaultType,
			NPUName:              faultInfo.NPUName,
			LargeModelFaultLevel: faultInfo.LargeModelFaultLevel,
			FaultLevel:           faultInfo.FaultLevel,
			FaultHandling:        faultInfo.FaultHandling,
			FaultCode:            code,
			FaultTime:            faultInfo.FaultTime,
		}
		deviceFaults = append(deviceFaults, newFault)
	}
	return deviceFaults
}

func mergeDeviceFault(deviceFaults []constant.DeviceFault) (constant.DeviceFault, error) {
	deviceName := deviceFaults[0].NPUName
	mergeFault := constant.DeviceFault{
		FaultType:            deviceFaults[0].FaultType,
		NPUName:              deviceName,
		LargeModelFaultLevel: deviceFaults[0].LargeModelFaultLevel,
		FaultLevel:           deviceFaults[0].FaultLevel,
		FaultHandling:        deviceFaults[0].FaultHandling,
		FaultTime:            deviceFaults[0].FaultTime,
	}
	faultCodeList := make([]string, 0)
	for _, fault := range deviceFaults {
		if fault.NPUName != deviceName {
			return constant.DeviceFault{}, fmt.Errorf("deviceFaults cannot merge, "+
				"they belongs to multiple devices: %s, %s", deviceName, fault.NPUName)
		}
		faultCodeList = append(faultCodeList, fault.FaultCode)
	}
	sort.SliceStable(faultCodeList, func(i, j int) bool {
		return faultCodeList[i] < faultCodeList[j]
	})
	mergeFault.FaultCode = strings.Join(faultCodeList, ",")
	return mergeFault, nil
}

func deleteFaultFromFaultMap(faultMap map[string][]constant.DeviceFault,
	delFault constant.DeviceFault) map[string][]constant.DeviceFault {
	deviceFaults, ok := faultMap[delFault.NPUName]
	if !ok {
		return faultMap
	}
	newDeviceFaults := make([]constant.DeviceFault, 0)
	for _, fault := range deviceFaults {
		if reflect.DeepEqual(delFault, fault) {
			continue
		}
		newDeviceFaults = append(newDeviceFaults, fault)
	}
	faultMap[delFault.NPUName] = newDeviceFaults
	return faultMap
}

func advanceDeviceCmForNodeMapToString(
	advanceDeviceCm map[string]AdvanceDeviceCm, orgDeviceCm map[string]*constant.DeviceInfo) {
	for nodeName, advanceCm := range advanceDeviceCm {
		advanceCm = mergeCodeAndRemoveUnhealthy(advanceCm)
		cmName := nodeNameToCmName(nodeName)
		deviceInfo, found := orgDeviceCm[cmName]
		if !found {
			continue
		}
		faultListKey := getFaultListKey(deviceInfo)
		if faultListKey != "" {
			orgDeviceCm[cmName].DeviceList[faultListKey] =
				util.ObjToString(faultMapToFaultList(advanceCm.DeviceList))
		}

		networkUnhealthyKey := getNetworkUnhealthyKey(deviceInfo)
		if networkUnhealthyKey != "" {
			orgDeviceCm[cmName].DeviceList[networkUnhealthyKey] = strings.Join(advanceCm.NetworkUnhealthy, ",")
		}

		cardUnhealthyKey := getCardUnhealthyKey(deviceInfo)
		if cardUnhealthyKey != "" {
			orgDeviceCm[cmName].DeviceList[cardUnhealthyKey] = strings.Join(advanceCm.CarUnHealthy, ",")
		}
	}
}

func faultMapToFaultList(deviceFaultMap map[string][]constant.DeviceFault) []constant.DeviceFault {
	deviceFaultList := make([]constant.DeviceFault, 0)
	for _, faultList := range deviceFaultMap {
		deviceFaultList = append(deviceFaultList, faultList...)
	}
	return deviceFaultList
}

func mergeCodeAndRemoveUnhealthy(advanceDeviceCm AdvanceDeviceCm) AdvanceDeviceCm {
	for deviceName, faults := range advanceDeviceCm.DeviceList {
		if len(faults) == 0 {
			advanceDeviceCm.NetworkUnhealthy = util.DeleteStringSliceItem(advanceDeviceCm.NetworkUnhealthy, deviceName)
			advanceDeviceCm.CarUnHealthy = util.DeleteStringSliceItem(advanceDeviceCm.CarUnHealthy, deviceName)
			hwlog.RunLog.Errorf("remove device %s from unhealthy", deviceName)
			continue
		}
		mergedFaults, err := mergeDeviceFault(faults)
		if err != nil {
			hwlog.RunLog.Errorf("merge device %s faults failed, exception: %v", deviceName, err)
			continue
		}
		advanceDeviceCm.DeviceList[deviceName] = []constant.DeviceFault{mergedFaults}
	}
	return advanceDeviceCm
}

func getFaultListKey(devInfo *constant.DeviceInfo) string {
	for key, _ := range devInfo.DeviceList {
		if strings.Contains(key, "huawei.com/Ascend") && strings.Contains(key, "-Fault") {
			return key
		}
	}
	return ""
}

func getNetworkUnhealthyKey(devInfo *constant.DeviceInfo) string {
	for key, _ := range devInfo.DeviceList {
		if strings.Contains(key, "huawei.com/Ascend") && strings.Contains(key, "-NetworkUnhealthy") {
			return key
		}
	}
	return ""
}

func getCardUnhealthyKey(devInfo *constant.DeviceInfo) string {
	for key, _ := range devInfo.DeviceList {
		if strings.Contains(key, "huawei.com/Ascend") && strings.Contains(key, "-Unhealthy") {
			return key
		}
	}
	return ""
}

func isUceFault(faultDevice constant.DeviceFault) bool {
	if strings.Contains(faultDevice.FaultCode, constant.UCE_FAULT_CODE) {
		return true
	}
	return false
}

func isUceAccompanyFault(faultDevice constant.DeviceFault) bool {
	return strings.Contains(faultDevice.FaultCode, constant.AIC_FAULT_CODE) ||
		strings.Contains(faultDevice.FaultCode, constant.AIV_FAULT_CODE)
}

func isDeviceFaultEqual(one, other constant.DeviceFault) bool {
	return reflect.DeepEqual(one, other)
}
