// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

func getNodeAndDeviceFromJobIdAndRankId(
	jobId, rankId string, jobServerInfoMap constant.JobServerInfoMap) (string, string, error) {
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

func getAdvanceDeviceCmForNodeMap(deviceInfoCms map[string]*constant.DeviceInfo) map[string]AdvanceDeviceFaultCm {
	advanceDeviceCmForNodeMap := make(map[string]AdvanceDeviceFaultCm)
	for _, deviceInfo := range deviceInfoCms {
		advanceDeviceCmForNodeMap[cmNameToNodeName(deviceInfo.CmName)] = getAdvanceDeviceCm(deviceInfo)
	}
	return advanceDeviceCmForNodeMap
}

// deviceName->faults
func getAdvanceDeviceCm(devInfo *constant.DeviceInfo) AdvanceDeviceFaultCm {
	advanceDeviceCm := AdvanceDeviceFaultCm{
		CmName:      devInfo.CmName,
		SuperPodID:  devInfo.SuperPodID,
		ServerIndex: devInfo.ServerIndex,
		UpdateTime:  devInfo.UpdateTime,
		ServerType:  getServerType(devInfo),
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
			hwlog.RunLog.Debugf("device fault: %s of cm %s, time: %s",
				util.ObjToString(deviceFault), devInfo.CmName, util.ReadableMsTime(devInfo.UpdateTime))
			// device plugin may merge multiple fault codes in one string
			deviceFaults := splitDeviceFault(deviceFault, cmNameToNodeName(devInfo.CmName))
			deviceFaultMap[deviceFault.NPUName] = append(deviceFaultMap[deviceFault.NPUName], deviceFaults...)
		}
		advanceDeviceCm.FaultDeviceList = deviceFaultMap
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
		var cardList []string
		if len(cardUnhealthyCardList) == 0 {
			cardList = make([]string, 0)
		} else {
			cardList = strings.Split(cardUnhealthyCardList, ",")
		}
		advanceDeviceCm.CardUnHealthy = cardList
	}
	return advanceDeviceCm
}

func getServerType(devInfo *constant.DeviceInfo) string {
	for key, _ := range devInfo.DeviceList {
		if strings.Contains(key, Ascend910Server) {
			return Ascend910Server
		}
		if strings.Contains(key, Ascend310PServer) {
			return Ascend310PServer
		}
		if strings.Contains(key, Ascend310Server) {
			return Ascend310Server
		}
	}
	hwlog.RunLog.Warn("cannot decide server type")
	return Ascend910Server
}

// device plugin may merge multiple fault codes in one string
func splitDeviceFault(faultInfo constant.DeviceFault, nodeName string) []constant.DeviceFault {
	deviceFaults := make([]constant.DeviceFault, 0)
	faultInfo.FaultCode = strings.Replace(faultInfo.FaultCode, " ", "", -1)
	codes := strings.Split(faultInfo.FaultCode, ",")
	for _, code := range codes {
		faultTimeAndLevel, found := faultInfo.FaultTimeAndLevelMap[code]
		var faultLevel string
		if !found {
			hwlog.RunLog.Warnf("cannot find fault level of code %s in device %s of node %s. map is %s.",
				code, util.ObjToString(faultTimeAndLevel), faultInfo.NPUName, nodeName)
			faultLevel = NormalNPU
		} else {
			faultLevel = faultTimeAndLevel.FaultLevel
		}
		newFault := constant.DeviceFault{
			FaultType:            faultInfo.FaultType,
			NPUName:              faultInfo.NPUName,
			LargeModelFaultLevel: faultLevel,
			FaultLevel:           faultLevel,
			FaultHandling:        faultLevel,
			FaultCode:            code,
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
				code: faultTimeAndLevel,
			},
		}
		deviceFaults = append(deviceFaults, newFault)
	}
	return deviceFaults
}

func mergeDeviceFault(notGroupDeviceFaults []constant.DeviceFault) ([]constant.DeviceFault, error) {
	faultsGroupByType := faultsGroupByType(notGroupDeviceFaults)
	result := make([]constant.DeviceFault, 0)
	for _, faultsGroup := range faultsGroupByType {
		deviceName := faultsGroup[0].NPUName
		fautLevels := make([]string, 0)
		newTimeAndLevelMap := make(map[string]constant.FaultTimeAndLevel, len(faultsGroup))
		faultCodeList := make([]string, 0)
		for _, fault := range faultsGroup {
			if fault.NPUName != deviceName {
				return []constant.DeviceFault{}, fmt.Errorf("deviceFaults cannot merge, "+
					"they belongs to multiple devices: %s, %s", deviceName, fault.NPUName)
			}
			faultCodeList = append(faultCodeList, fault.FaultCode)
			fautLevels = append(fautLevels, fault.FaultLevel)
			newTimeAndLevelMap[fault.FaultCode] = fault.FaultTimeAndLevelMap[fault.FaultCode]
		}
		faultLevel := getMostSeriousFaultLevel(fautLevels)
		mergeFault := constant.DeviceFault{
			FaultType:            faultsGroup[0].FaultType,
			NPUName:              deviceName,
			FaultTimeAndLevelMap: newTimeAndLevelMap,
		}
		mergeFault.FaultLevel = faultLevel
		mergeFault.LargeModelFaultLevel = faultLevel
		mergeFault.FaultHandling = faultLevel
		mergeFault.FaultCode = strings.Join(faultCodeList, ",")
		result = append(result, mergeFault)
	}
	return result, nil
}

func deleteFaultFromFaultMap(faultMap map[string][]constant.DeviceFault,
	delFault constant.DeviceFault) map[string][]constant.DeviceFault {
	if faultMap == nil {
		return make(map[string][]constant.DeviceFault)
	}
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

func addFaultIntoFaultMap(faultMap map[string][]constant.DeviceFault,
	addFault constant.DeviceFault) map[string][]constant.DeviceFault {
	if faultMap == nil {
		faultMap = make(map[string][]constant.DeviceFault)
	}
	deviceFaults, ok := faultMap[addFault.NPUName]
	if !ok {
		deviceFaults = make([]constant.DeviceFault, 0)
	}
	isExisting := false
	for _, fault := range deviceFaults {
		if reflect.DeepEqual(addFault, fault) {
			isExisting = true
			break
		}
	}
	if !isExisting {
		deviceFaults = append(deviceFaults, addFault)
	}
	faultMap[addFault.NPUName] = deviceFaults
	return faultMap
}

func advanceDeviceCmForNodeMapToString(
	advanceDeviceCm map[string]AdvanceDeviceFaultCm, orgDeviceCm map[string]*constant.DeviceInfo) {
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
				util.ObjToString(faultMapToFaultList(advanceCm.FaultDeviceList))
		}

		networkUnhealthyKey := getNetworkUnhealthyKey(deviceInfo)
		if networkUnhealthyKey != "" {
			orgDeviceCm[cmName].DeviceList[networkUnhealthyKey] = strings.Join(advanceCm.NetworkUnhealthy, ",")
		}

		cardUnhealthyKey := getCardUnhealthyKey(deviceInfo)
		if cardUnhealthyKey != "" {
			orgDeviceCm[cmName].DeviceList[cardUnhealthyKey] = strings.Join(advanceCm.CardUnHealthy, ",")
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

func faultsGroupByType(faults []constant.DeviceFault) map[string][]constant.DeviceFault {
	result := make(map[string][]constant.DeviceFault)
	for _, fault := range faults {
		_, found := result[fault.FaultType]
		if !found {
			result[fault.FaultType] = make([]constant.DeviceFault, 0)
		}
		result[fault.FaultType] = append(result[fault.FaultType], fault)
	}
	return result
}

func mergeCodeAndRemoveUnhealthy(advanceDeviceCm AdvanceDeviceFaultCm) AdvanceDeviceFaultCm {
	for deviceName, faults := range advanceDeviceCm.FaultDeviceList {
		if len(faults) == 0 {
			advanceDeviceCm.NetworkUnhealthy = util.DeleteStringSliceItem(advanceDeviceCm.NetworkUnhealthy, deviceName)
			advanceDeviceCm.CardUnHealthy = util.DeleteStringSliceItem(advanceDeviceCm.CardUnHealthy, deviceName)
			hwlog.RunLog.Errorf("remove device %s from unhealthy", deviceName)
			continue
		}
		mergedFaults, err := mergeDeviceFault(faults)
		if err != nil {
			hwlog.RunLog.Errorf("merge device %s faults failed, exception: %v", deviceName, err)
			continue
		}
		advanceDeviceCm.FaultDeviceList[deviceName] = mergedFaults
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
	if strings.Contains(faultDevice.FaultCode, constant.UceFaultCode) {
		return true
	}
	return false
}

func isCqeFault(faultDevice constant.DeviceFault) bool {
	return strings.Contains(faultDevice.FaultCode, constant.DevCqeFaultCode) ||
		strings.Contains(faultDevice.FaultCode, constant.HostCqeFaultCode)
}

func isLinkDownFault(faultDevice constant.DeviceFault) bool {
	return strings.Contains(faultDevice.FaultCode, constant.LinkDownFaultCode)
}

func isUceAccompanyFault(faultDevice constant.DeviceFault) bool {
	return strings.Contains(faultDevice.FaultCode, constant.AicFaultCode) ||
		strings.Contains(faultDevice.FaultCode, constant.AivFaultCode)
}

func isDeviceFaultEqual(one, other constant.DeviceFault) bool {
	return reflect.DeepEqual(one, other)
}

func getMostSeriousFaultLevel(fautLevels []string) string {
	faultTypeSet := sets.NewString(fautLevels...)
	if faultTypeSet.Has(ManuallySeparateNPU) {
		return ManuallySeparateNPU
	} else if faultTypeSet.Has(SeparateNPU) {
		return SeparateNPU
	} else if faultTypeSet.Has(PreSeparateNPU) {
		return PreSeparateNPU
	} else if faultTypeSet.Has(RestartNPU) {
		return RestartNPU
	} else if faultTypeSet.Has(FreeRestartNPU) {
		return FreeRestartNPU
	} else if faultTypeSet.Has(RestartBusiness) {
		return RestartBusiness
	} else if faultTypeSet.Has(RestartRequest) {
		return RestartRequest
	} else if faultTypeSet.Has(SubHealthFault) {
		return SubHealthFault
	} else if faultTypeSet.Has(NotHandleFault) {
		return NotHandleFault
	}
	return NormalNPU
}

func getFaultTime(fault constant.DeviceFault, errorMsg string) int64 {
	faultTimeAndLevel, ok := fault.FaultTimeAndLevelMap[fault.FaultCode]
	var faultTime int64
	if !ok {
		hwlog.RunLog.Errorf("cannot find fault time of %s. bussiness info: %s",
			util.ObjToString(fault), errorMsg)
		faultTime = constant.DeviceNotFault
	} else {
		faultTime = faultTimeAndLevel.FaultTime
	}
	return faultTime
}

func getContainedElementIdx(element string, stringList []string) int {
	for idx, deviceName := range stringList {
		if element == deviceName {
			return idx
		}
	}
	return -1
}

func canDoStepRetry(uceDevice *uceDeviceInfo) bool {
	if uceDevice.RecoverTime == constant.JobNotRecover {
		return false
	}
	if time.Now().UnixMilli()-constant.JobReportRecoverTimeout <= uceDevice.RecoverTime {
		return true
	}
	if uceDevice.FaultTime == constant.DeviceNotFault {
		return false
	}
	if uceDevice.FaultTime+constant.JobReportRecoverTimeout >= uceDevice.RecoverTime {
		return true
	}
	return false
}

func expiredReportInfo(info *reportInfo) bool {
	if info.RecoverTime != constant.JobNotRecover &&
		time.Now().UnixMilli()-constant.JobReportInfoExpiredTimeout > info.RecoverTime {
		return true
	}
	return false
}
