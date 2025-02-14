// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package faultmanager contain fault process
package faultdomain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

// IsNodeReady returns the node ready status
func IsNodeReady(node *v1.Node) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == v1.NodeReady {
			return cond.Status == v1.ConditionTrue
		}
	}
	return false
}

// GetNodeAndDeviceFromJobIdAndRankId get node and device name from jobId and rankId
func GetNodeAndDeviceFromJobIdAndRankId(
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

// CmNameToNodeName convert cmName to nodeName
func CmNameToNodeName(cmName string) string {
	if !strings.HasPrefix(cmName, constant.DeviceInfoPrefix) {
		hwlog.RunLog.Errorf("CmName %s has not prefix %s", cmName, constant.DeviceInfoPrefix)
		return cmName
	}
	return strings.TrimPrefix(cmName, constant.DeviceInfoPrefix)
}

func nodeNameToCmName(nodeName string) string {
	return constant.DeviceInfoPrefix + nodeName
}

// GetAdvanceDeviceCmForNodeMap get advance device cm for node map
func GetAdvanceDeviceCmForNodeMap(
	deviceInfoCms map[string]*constant.DeviceInfo) map[string]constant.AdvanceDeviceFaultCm {
	advanceDeviceCmForNodeMap := make(map[string]constant.AdvanceDeviceFaultCm)
	for _, deviceInfo := range deviceInfoCms {
		advanceDeviceCmForNodeMap[CmNameToNodeName(deviceInfo.CmName)] = GetAdvanceDeviceCm(deviceInfo)
	}
	return advanceDeviceCmForNodeMap
}

// GetAdvanceDeviceCm deviceName->faults
func GetAdvanceDeviceCm(devInfo *constant.DeviceInfo) constant.AdvanceDeviceFaultCm {
	advanceDeviceCm := constant.AdvanceDeviceFaultCm{
		CmName:      devInfo.CmName,
		SuperPodID:  devInfo.SuperPodID,
		ServerIndex: devInfo.ServerIndex,
		UpdateTime:  devInfo.UpdateTime,
		ServerType:  GetServerType(devInfo),
	}
	if faultList, ok := devInfo.DeviceList[GetFaultListKey(devInfo)]; ok {
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
			deviceFaults := splitDeviceFault(deviceFault, CmNameToNodeName(devInfo.CmName))
			deviceFaultMap[deviceFault.NPUName] = append(deviceFaultMap[deviceFault.NPUName], deviceFaults...)
		}
		advanceDeviceCm.FaultDeviceList = deviceFaultMap
	} else {
		hwlog.RunLog.Infof("get fault list for node %v failed. fault list does not exist", devInfo.CmName)
	}
	if networkUnhealthyCardList, ok := devInfo.DeviceList[GetNetworkUnhealthyKey(devInfo)]; ok {
		cardList := strings.Split(networkUnhealthyCardList, ",")
		advanceDeviceCm.NetworkUnhealthy = cardList
	} else {
		hwlog.RunLog.Infof("get NetworkUnhealthy list for node %v failed. fault list does not exist",
			devInfo.CmName)
	}
	if cardUnhealthyCardList, ok := devInfo.DeviceList[GetCardUnhealthyKey(devInfo)]; ok {
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

// GetServerType get server type from DeviceInfo
func GetServerType(devInfo *constant.DeviceInfo) string {
	for key, _ := range devInfo.DeviceList {
		if strings.Contains(key, constant.Ascend910Server) {
			return constant.Ascend910Server
		}
		if strings.Contains(key, constant.Ascend310PServer) {
			return constant.Ascend310PServer
		}
		if strings.Contains(key, constant.Ascend310Server) {
			return constant.Ascend310Server
		}
	}
	hwlog.RunLog.Warn("cannot decide server type")
	return constant.Ascend910Server
}

// device plugin may merge multiple fault codes in one string
func splitDeviceFault(faultInfo constant.DeviceFault, nodeName string) []constant.DeviceFault {
	deviceFaults := make([]constant.DeviceFault, 0)
	faultInfo.FaultCode = strings.Replace(faultInfo.FaultCode, " ", "", -1)
	codes := strings.Split(faultInfo.FaultCode, ",")
	for _, code := range codes {
		var faultTimeAndLevel constant.FaultTimeAndLevel
		var found bool
		if code == "" && faultInfo.FaultLevel == constant.ManuallySeparateNPU {
			code = constant.ManuallySeparateNPU
			faultTimeAndLevel = constant.FaultTimeAndLevel{
				FaultTime:  constant.UnknownFaultTime,
				FaultLevel: constant.ManuallySeparateNPU,
			}
			found = true
		} else {
			faultTimeAndLevel, found = faultInfo.FaultTimeAndLevelMap[code]
		}
		var faultLevel string
		if !found {
			hwlog.RunLog.Warnf("cannot find faultTimeAndLevel of code %s in faultInfo %s of node %s.",
				code, util.ObjToString(faultInfo), nodeName)
			faultLevel = faultInfo.FaultLevel
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
			fautLevels = append(fautLevels, fault.FaultLevel)
			if fault.FaultCode != constant.ManuallySeparateNPU {
				faultCodeList = append(faultCodeList, fault.FaultCode)
				newTimeAndLevelMap[fault.FaultCode] = fault.FaultTimeAndLevelMap[fault.FaultCode]
			}
		}
		faultLevel := GetMostSeriousFaultLevel(fautLevels)
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

// DeleteFaultFromFaultMap delete fault from faultMap
func DeleteFaultFromFaultMap(faultMap map[string][]constant.DeviceFault,
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

// AddFaultIntoFaultMap add fault into faultMap
func AddFaultIntoFaultMap(faultMap map[string][]constant.DeviceFault,
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

// AdvanceDeviceCmForNodeMapToString convert advance device cm to original format
func AdvanceDeviceCmForNodeMapToString(
	advanceDeviceCm map[string]constant.AdvanceDeviceFaultCm, orgDeviceCm map[string]*constant.DeviceInfo) {
	for nodeName, advanceCm := range advanceDeviceCm {
		advanceCm = mergeCodeAndRemoveUnhealthy(advanceCm)
		cmName := nodeNameToCmName(nodeName)
		deviceInfo, found := orgDeviceCm[cmName]
		if !found {
			continue
		}
		faultListKey := GetFaultListKey(deviceInfo)
		if faultListKey != "" {
			orgDeviceCm[cmName].DeviceList[faultListKey] =
				util.ObjToString(faultMapToFaultList(advanceCm.FaultDeviceList))
		}

		networkUnhealthyKey := GetNetworkUnhealthyKey(deviceInfo)
		if networkUnhealthyKey != "" {
			orgDeviceCm[cmName].DeviceList[networkUnhealthyKey] = strings.Join(advanceCm.NetworkUnhealthy, ",")
		}

		cardUnhealthyKey := GetCardUnhealthyKey(deviceInfo)
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

func mergeCodeAndRemoveUnhealthy(advanceDeviceCm constant.AdvanceDeviceFaultCm) constant.AdvanceDeviceFaultCm {
	for deviceName, faults := range advanceDeviceCm.FaultDeviceList {
		if len(faults) == 0 {
			advanceDeviceCm.NetworkUnhealthy = util.DeleteStringSliceItem(advanceDeviceCm.NetworkUnhealthy, deviceName)
			advanceDeviceCm.CardUnHealthy = util.DeleteStringSliceItem(advanceDeviceCm.CardUnHealthy, deviceName)
			hwlog.RunLog.Warnf("remove device %s from unhealthy", deviceName)
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

// GetFaultListKey get FaultList key in DeviceInfo
func GetFaultListKey(devInfo *constant.DeviceInfo) string {
	for key, _ := range devInfo.DeviceList {
		if strings.Contains(key, "huawei.com/Ascend") && strings.Contains(key, "-Fault") {
			return key
		}
	}
	return ""
}

// GetNetworkUnhealthyKey get networkUnhealthy key in DeviceInfo
func GetNetworkUnhealthyKey(devInfo *constant.DeviceInfo) string {
	for key, _ := range devInfo.DeviceList {
		if strings.Contains(key, "huawei.com/Ascend") && strings.Contains(key, "-NetworkUnhealthy") {
			return key
		}
	}
	return ""
}

// GetCardUnhealthyKey get CardUnhealthy key in DeviceInfo
func GetCardUnhealthyKey(devInfo *constant.DeviceInfo) string {
	for key, _ := range devInfo.DeviceList {
		if strings.Contains(key, "huawei.com/Ascend") && strings.Contains(key, "-Unhealthy") {
			return key
		}
	}
	return ""
}

// IsUceFault check faultCode is uce
func IsUceFault(faultCode string) bool {
	if strings.Contains(faultCode, constant.UceFaultCode) {
		return true
	}
	return false
}

// IsCqeFault check faultCode is cqe fault
func IsCqeFault(faultCode string) bool {
	return strings.Contains(faultCode, constant.DevCqeFaultCode) ||
		strings.Contains(faultCode, constant.HostCqeFaultCode)
}

// IsLinkDownFault check faultCode is linkdown fault
func IsLinkDownFault(faultCode string) bool {
	return strings.Contains(faultCode, constant.LinkDownFaultCode)
}

// IsUceAccompanyFault check faultCode is uce accompany
func IsUceAccompanyFault(faultCode string) bool {
	return strings.Contains(faultCode, constant.AicFaultCode) ||
		strings.Contains(faultCode, constant.AivFaultCode)
}

// IsDeviceFaultEqual check two DeviceFault is equal
func IsDeviceFaultEqual(one, other constant.DeviceFault) bool {
	return reflect.DeepEqual(one, other)
}

// GetMostSeriousFaultLevel get most serious fault level
func GetMostSeriousFaultLevel(fautLevels []string) string {
	faultTypeSet := sets.NewString(fautLevels...)
	if faultTypeSet.Has(constant.ManuallySeparateNPU) {
		return constant.ManuallySeparateNPU
	} else if faultTypeSet.Has(constant.SeparateNPU) {
		return constant.SeparateNPU
	} else if faultTypeSet.Has(constant.PreSeparateNPU) {
		return constant.PreSeparateNPU
	} else if faultTypeSet.Has(constant.RestartNPU) {
		return constant.RestartNPU
	} else if faultTypeSet.Has(constant.FreeRestartNPU) {
		return constant.FreeRestartNPU
	} else if faultTypeSet.Has(constant.RestartBusiness) {
		return constant.RestartBusiness
	} else if faultTypeSet.Has(constant.RestartRequest) {
		return constant.RestartRequest
	} else if faultTypeSet.Has(constant.SubHealthFault) {
		return constant.SubHealthFault
	} else if faultTypeSet.Has(constant.NotHandleFault) {
		return constant.NotHandleFault
	}
	return constant.NormalNPU
}

// GetFaultTime get fault time in fault
func GetFaultTime(fault constant.DeviceFault, errorMsg string) int64 {
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

// GetContainedElementIdx get element idx in stringList
func GetContainedElementIdx(element string, stringList []string) int {
	for idx, deviceName := range stringList {
		if element == deviceName {
			return idx
		}
	}
	return -1
}

// CanDoStepRetry check UceDeviceInfo can do step retry
func CanDoStepRetry(uceDevice *constant.UceDeviceInfo) bool {
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

// ValidBusinessRecoverTime check recoverTime is valid
func ValidBusinessRecoverTime(recoverTime int64) bool {
	if recoverTime != constant.JobNotRecover &&
		time.Now().UnixMilli()-constant.JobReportInfoExpiredTimeout <= recoverTime {
		return true
	}
	return false
}

// ValidBusinessUceReportInfo check ReportInfo is valid
func ValidBusinessUceReportInfo(info *constant.ReportInfo) bool {
	return ValidBusinessRecoverTime(info.RecoverTime)
}
