// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package faultmanager contain fault process
package faultdomain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/strings/slices"

	"ascend-common/api"
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

// GetAdvanceFaultCm return more usable fault cm, ONLY FOR TESTCASE
func GetAdvanceFaultCm[U, T constant.ConfigMapInterface](
	cmInfos map[string]T) map[string]U {
	result := make(map[string]U)
	for _, info := range cmInfos {
		result[CmNameToNodeName(info.GetCmName())] = GetAdvanceFaultForNode(info).(U)
	}
	return result
}

// GetAdvanceFaultForNode return more usable fault cm for one node
func GetAdvanceFaultForNode[T constant.ConfigMapInterface](cmForNode T) constant.ConfigMapInterface {
	switch cm := any(cmForNode).(type) {
	case *constant.DeviceInfo:
		return GetAdvanceDeviceCm(cm)
	case *constant.NodeInfo:
		return cm
	case *constant.SwitchInfo:
		return cm
	case *constant.AdvanceDeviceFaultCm:
		return cm
	default:
		hwlog.RunLog.Errorf("cmForNode type is not support.")
		return nil
	}
}

// GetAdvanceDeviceCm return more usable device cm
func GetAdvanceDeviceCm(devInfo *constant.DeviceInfo) *constant.AdvanceDeviceFaultCm {
	advanceDeviceCm := &constant.AdvanceDeviceFaultCm{
		CmName:      devInfo.CmName,
		SuperPodID:  devInfo.SuperPodID,
		ServerIndex: devInfo.ServerIndex,
		UpdateTime:  devInfo.UpdateTime,
		DeviceType:  GetDeviceType(devInfo),
	}
	advanceDeviceCm.FaultDeviceList = getFaultListInfo(devInfo)
	advanceDeviceCm.NetworkUnhealthy = getNetworkUnhealthyCardList(devInfo)
	advanceDeviceCm.CardUnHealthy = getCardUnHealthy(devInfo)
	advanceDeviceCm.AvailableDeviceList = getAvailableDevices(devInfo)
	advanceDeviceCm.Recovering = getRecoveringDevList(devInfo)
	return advanceDeviceCm
}

func getFaultListInfo(devInfo *constant.DeviceInfo) map[string][]constant.DeviceFault {
	_, faultList := getFaultListString(devInfo)
	if len(faultList) == 0 {
		hwlog.RunLog.Infof("get fault list for node %v failed. fault list does not exist", devInfo.CmName)
		return make(map[string][]constant.DeviceFault)
	}
	var devicesFault []constant.DeviceFault
	err := json.Unmarshal([]byte(faultList), &devicesFault)
	if err != nil {
		hwlog.RunLog.Errorf("get fault list for node %v failed. "+
			"Json unmarshall exception: %v", devInfo.CmName, err)
		return make(map[string][]constant.DeviceFault)
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
	return deviceFaultMap
}

func getCardUnHealthy(devInfo *constant.DeviceInfo) []string {
	_, info := getCardUnhealthyString(devInfo)
	if len(info) == 0 {
		return make([]string, 0)
	}
	return strings.Split(info, ",")
}

func getNetworkUnhealthyCardList(devInfo *constant.DeviceInfo) []string {
	_, info := getNetworkUnhealthyString(devInfo)
	if len(info) == 0 {
		return make([]string, 0)
	}
	return strings.Split(info, ",")
}

func getAvailableDevices(devInfo *constant.DeviceInfo) []string {
	_, info := getAvailDevListString(devInfo)
	if len(info) == 0 {
		return make([]string, 0)
	}
	return strings.Split(info, ",")
}

func getRecoveringDevList(devInfo *constant.DeviceInfo) []string {
	_, info := getRecoveringString(devInfo)
	if len(info) == 0 {
		return make([]string, 0)
	}
	return strings.Split(info, ",")
}

// GetDeviceType get device type from device info
func GetDeviceType(devInfo *constant.DeviceInfo) string {
	for key, _ := range devInfo.DeviceList {
		if strings.Contains(key, constant.Ascend910) {
			return constant.Ascend910
		}
		if strings.Contains(key, constant.Ascend310P) {
			return constant.Ascend310P
		}
		if strings.Contains(key, constant.Ascend310) {
			return constant.Ascend310
		}
	}
	hwlog.RunLog.Warn("cannot decide server type")
	return constant.Ascend910
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
	faultTypes := getSortedKeys(faultsGroupByType)
	for _, faultType := range faultTypes {
		faultsGroup := faultsGroupByType[faultType]
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

func AdvanceFaultMapToOriginalFaultMap[U, T constant.ConfigMapInterface](advanceFaultCm map[string]T) map[string]U {
	orgFaultCm := make(map[string]U)
	for _, advanceCmForNode := range advanceFaultCm {
		orgFaultCm[advanceCmForNode.GetCmName()] = AdvanceCmToOriginalCm(advanceCmForNode).(U)
	}
	return orgFaultCm
}

func AdvanceCmToOriginalCm[T constant.ConfigMapInterface](advanceCmForNode T) constant.ConfigMapInterface {
	switch cm := any(advanceCmForNode).(type) {
	case *constant.AdvanceDeviceFaultCm:
		return AdvanceDevCmToOrigCm(cm)
	case *constant.SwitchInfo:
		return cm
	case *constant.NodeInfo:
		return cm
	default:
		hwlog.RunLog.Errorf("AdvanceFaultCmToOriginalCmForNode don't support this type.")
		return nil
	}
}

// AdvanceDevCmToOrigCm convert advance device cm to original format
func AdvanceDevCmToOrigCm(advanceDeviceCm *constant.AdvanceDeviceFaultCm) *constant.DeviceInfo {
	orgDeviceCm := &constant.DeviceInfo{
		DeviceInfoNoName: constant.DeviceInfoNoName{
			DeviceList: make(map[string]string),
			UpdateTime: advanceDeviceCm.UpdateTime,
		},
		CmName:      advanceDeviceCm.CmName,
		SuperPodID:  advanceDeviceCm.SuperPodID,
		ServerIndex: advanceDeviceCm.ServerIndex,
	}

	mergeCode(advanceDeviceCm)

	orgDeviceCm.DeviceList[advanceDeviceCm.GetFaultDeviceListKey()] =
		util.ObjToString(faultMapToFaultList(advanceDeviceCm.FaultDeviceList))

	orgDeviceCm.DeviceList[advanceDeviceCm.GetNetworkUnhealthyKey()] = ""
	if len(advanceDeviceCm.NetworkUnhealthy) > 0 {
		orgDeviceCm.DeviceList[advanceDeviceCm.GetNetworkUnhealthyKey()] =
			strings.Join(advanceDeviceCm.NetworkUnhealthy, ",")
	}

	orgDeviceCm.DeviceList[advanceDeviceCm.GetCardUnHealthyKey()] = ""
	if len(advanceDeviceCm.CardUnHealthy) > 0 {
		orgDeviceCm.DeviceList[advanceDeviceCm.GetCardUnHealthyKey()] =
			strings.Join(advanceDeviceCm.CardUnHealthy, ",")
	}

	orgDeviceCm.DeviceList[advanceDeviceCm.GetRecoveringKey()] = ""
	if len(advanceDeviceCm.Recovering) > 0 {
		orgDeviceCm.DeviceList[advanceDeviceCm.GetRecoveringKey()] =
			strings.Join(advanceDeviceCm.Recovering, ",")
	}

	orgDeviceCm.DeviceList[advanceDeviceCm.GetAvailableDeviceListKey()] = ""
	if len(advanceDeviceCm.AvailableDeviceList) > 0 {
		orgDeviceCm.DeviceList[advanceDeviceCm.GetAvailableDeviceListKey()] =
			strings.Join(advanceDeviceCm.AvailableDeviceList, ",")
	}
	return orgDeviceCm
}

func faultMapToFaultList(deviceFaultMap map[string][]constant.DeviceFault) []constant.DeviceFault {
	deviceFaultList := make([]constant.DeviceFault, 0)
	deviceNames := getSortedKeys(deviceFaultMap)
	for _, deviceName := range deviceNames {
		deviceFaultList = append(deviceFaultList, deviceFaultMap[deviceName]...)
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

func isFaultDeletable(faults []constant.DeviceFault, faultTypes []string, faultLevels []string) bool {
	for _, fault := range faults {
		if slices.Contains(faultTypes, fault.FaultType) && !slices.Contains(faultLevels, fault.FaultLevel) {
			return false
		}
	}
	return true
}

func mergeCode(advanceDeviceCm *constant.AdvanceDeviceFaultCm) {
	for deviceName, faults := range advanceDeviceCm.FaultDeviceList {
		if len(faults) == 0 {
			continue
		}
		mergedFaults, err := mergeDeviceFault(faults)
		if err != nil {
			hwlog.RunLog.Errorf("merge device %s faults failed, exception: %v", deviceName, err)
			continue
		}
		advanceDeviceCm.FaultDeviceList[deviceName] = mergedFaults
	}
}

func getNetworkUnhealthyString(devInfo *constant.DeviceInfo) (string, string) {
	key := api.ResourceNamePrefix + GetDeviceType(devInfo) + constant.CmCardNetworkUnhealthySuffix
	return key, devInfo.DeviceList[key]
}

func getCardUnhealthyString(devInfo *constant.DeviceInfo) (string, string) {
	key := api.ResourceNamePrefix + GetDeviceType(devInfo) + constant.CmCardUnhealthySuffix
	return key, devInfo.DeviceList[key]
}

func getRecoveringString(devInfo *constant.DeviceInfo) (string, string) {
	key := api.ResourceNamePrefix + GetDeviceType(devInfo) + constant.CmRecoveringSuffix
	return key, devInfo.DeviceList[key]
}

func getFaultListString(devInfo *constant.DeviceInfo) (string, string) {
	key := api.ResourceNamePrefix + GetDeviceType(devInfo) + constant.CmFaultListSuffix
	return key, devInfo.DeviceList[key]
}

func getAvailDevListString(devCMInfo *constant.DeviceInfo) (string, string) {
	availKey := api.ResourceNamePrefix + GetDeviceType(devCMInfo)
	availDevList, ok := devCMInfo.DeviceList[availKey]
	if !ok {
		return "", ""
	}
	return availKey, availDevList
}

// IsUceFault check faultCode is uce
func IsUceFault(faultCode string) bool {
	if strings.Contains(faultCode, constant.UceFaultCode) {
		return true
	}
	return false
}

// IsHcclRetryFault check faultCode is uce
func IsHcclRetryFault(faultCode string) bool {
	if strings.Contains(faultCode, constant.HcclRetryFaultCode) {
		return true
	}
	return false
}

// GetRetryTypeByFaultCode get retry fault type by fault code
func GetRetryTypeByFaultCode(faultCode string) string {
	if IsUceFault(faultCode) {
		return constant.UceFaultType
	} else if IsHcclRetryFault(faultCode) {
		return constant.HcclFaultType
	}
	return constant.NormalFaultType
}

// GetRetryCodeByFaultType get retry fault code by fault type
func GetRetryCodeByFaultType(faultType string) string {
	var faultCode string
	if faultType == constant.UceFaultType {
		faultCode = constant.UceFaultCode
	} else if faultType == constant.HcclFaultType {
		faultCode = constant.HcclRetryFaultCode
	}
	return faultCode
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

// IsL2L3Fault check faultLevel is L2 or L3
func IsL2L3Fault(faultLevel string) bool {
	return faultLevel == constant.RestartRequest || faultLevel == constant.RestartBusiness
}

// IsL1Fault check faultLevel is L1
func IsL1Fault(faultLevel string) bool {
	return faultLevel == constant.NotHandleFault
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

// CanDoStepRetry check DeviceFaultDetail can do step retry
func CanDoStepRetry(faultDetail *constant.DeviceFaultDetail) bool {
	if faultDetail.RecoverTime == constant.JobNotRecover {
		return false
	}
	if time.Now().UnixMilli()-constant.JobReportRecoverTimeout <= faultDetail.RecoverTime {
		return true
	}
	if faultDetail.FaultTime == constant.DeviceNotFault {
		return false
	}
	if faultDetail.FaultTime+constant.JobReportRecoverTimeout >= faultDetail.RecoverTime {
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

// ValidBusinessRetryReportInfo check ReportInfo is valid
func ValidBusinessRetryReportInfo(info *constant.ReportInfo) bool {
	return ValidBusinessRecoverTime(info.RecoverTime)
}

// SortDataForAdvanceDeviceInfo sort the field of deviceInfo
func SortDataForAdvanceDeviceInfo(deviceInfo *constant.AdvanceDeviceFaultCm) {
	sort.Strings(deviceInfo.AvailableDeviceList)
	sort.Strings(deviceInfo.CardUnHealthy)
	sort.Strings(deviceInfo.NetworkUnhealthy)
	sort.Strings(deviceInfo.Recovering)
	for _, faultList := range deviceInfo.FaultDeviceList {
		sort.Slice(faultList, func(i, j int) bool {
			if compareDeviceFault(faultList[i], faultList[j]) <= 0 {
				return true
			}
			return false
		})
	}
}

func compareDeviceFault(a, b constant.DeviceFault) int {
	if res := strings.Compare(a.FaultType, b.FaultType); res != 0 {
		return res
	}
	if res := strings.Compare(a.NPUName, b.NPUName); res != 0 {
		return res
	}
	if res := strings.Compare(a.LargeModelFaultLevel, b.LargeModelFaultLevel); res != 0 {
		return res
	}
	if res := strings.Compare(a.FaultLevel, b.FaultLevel); res != 0 {
		return res
	}
	if res := strings.Compare(a.FaultHandling, b.FaultHandling); res != 0 {
		return res
	}
	if res := strings.Compare(a.FaultCode, b.FaultCode); res != 0 {
		return res
	}
	keysA := getSortedKeys(a.FaultTimeAndLevelMap)
	keysB := getSortedKeys(b.FaultTimeAndLevelMap)
	for i := 0; i < len(keysA); i++ {
		if cmp := strings.Compare(keysA[i], keysB[i]); cmp != 0 {
			return cmp
		}
		valA := a.FaultTimeAndLevelMap[keysA[i]]
		valB := b.FaultTimeAndLevelMap[keysB[i]]
		if cmp := compareFaultTimeAndLevel(valA, valB); cmp != 0 {
			return cmp
		}
	}
	return 0
}

func compareFaultTimeAndLevel(a, b constant.FaultTimeAndLevel) int {
	if res := a.FaultTime - b.FaultTime; res != 0 {
		return int(res)
	}
	return strings.Compare(a.FaultLevel, b.FaultLevel)
}

func getSortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
