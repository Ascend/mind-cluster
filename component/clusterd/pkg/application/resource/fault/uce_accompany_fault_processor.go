package fault

import (
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/device"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"time"
)

// UceAccompanyFaultProcessor:
// aic aiv fault can be 1) accompanied by uce fault, also can 2) curr alone.
// if 1) aic aiv fault should be filtered. Once find aic fault, check if there is an uce fault 5s ago
// if 2) aic aiv fault should not be retained.
type UceAccompanyFaultProcessor struct {
	deviceCenter *DeviceFaultProcessCenter
	// maintain 5s ago device info
	DiagnosisAccompanyTimeout int64
	// nodeName -> deviceName -> faultQue
	uceAccompanyFaultQue map[string]map[string][]constant.DeviceFault
	// uceFaultTime
	uceFaultTime map[string]map[string]int64
}

func (processor *UceAccompanyFaultProcessor) uceAccompanyFaultInQue(deviceInfos map[string]*constant.DeviceInfo) {
	for _, deviceInfo := range deviceInfos {
		nodeName, err := cmNameToNodeName(deviceInfo.CmName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		processor.uceAccompanyFaultForNode(nodeName, deviceInfo)
	}
}

func (processor *UceAccompanyFaultProcessor) uceAccompanyFaultForNode(nodeName string, deviceInfo *constant.DeviceInfo) {
	if _, ok := processor.uceAccompanyFaultQue[nodeName]; !ok {
		processor.uceAccompanyFaultQue[nodeName] = make(map[string][]constant.DeviceFault)
		processor.uceFaultTime[nodeName] = make(map[string]int64)
	}
	faultMap := device.GetFaultMap(deviceInfo)
	for deviceName, deviceFaults := range faultMap {
		for _, fault := range deviceFaults {
			if device.IsUceFault(fault) {
				processor.uceFaultTime[nodeName][deviceName] = fault.FaultTime
				continue
			}
			if !device.IsUceAccompanyFault(fault) {
				continue
			}
			if _, ok := processor.uceAccompanyFaultQue[nodeName][deviceName]; !ok {
				processor.uceAccompanyFaultQue[nodeName][deviceName] = make([]constant.DeviceFault, 0)
			}

			// in que
			faultsInfo := processor.uceAccompanyFaultQue[nodeName][deviceName]
			processor.uceAccompanyFaultQue[nodeName][deviceName] = append(faultsInfo, fault)
		}
	}
}

func (processor *UceAccompanyFaultProcessor) filterFaultInfos(currentTime int64,
	deviceInfos map[string]*constant.DeviceInfo) map[string]*constant.DeviceInfo {
	for nodeName, nodeFaults := range processor.uceAccompanyFaultQue {
		faultMap := device.GetFaultMap(deviceInfos[nodeNameToCmName(nodeName)])
		for deviceName, deviceFaultQue := range nodeFaults {
			newQue, newFaultMap := processor.filterFaultDevice(faultMap, currentTime, nodeName, deviceName, deviceFaultQue)
			nodeFaults[deviceName] = newQue
			faultMap = newFaultMap
		}
		deviceInfos[nodeNameToCmName(nodeName)].DeviceList[device.GetFaultListKey()] = device.FaultMapToArrayToString(faultMap)
	}
	return deviceInfos
}

func (processor *UceAccompanyFaultProcessor) filterFaultDevice(
	faultMap map[string][]constant.DeviceFault, currentTime int64, nodeName, deviceName string,
	deviceFaultQue []constant.DeviceFault) ([]constant.DeviceFault, map[string][]constant.DeviceFault) {
	newDeviceFaultQue := make([]constant.DeviceFault, 0)
	for _, fault := range deviceFaultQue {
		uceFaultTime := processor.getDeviceUceFaultTime(nodeName, deviceName)
		accompanyFaultTime := fault.FaultTime
		// if is accompanied fault, filter
		if processor.isAccompaniedFaultByUce(uceFaultTime, accompanyFaultTime) {
			faultMap = device.DeleteFaultFromFaultMap(faultMap, fault)
			continue
		}
		// if current is not exceed diagnosis time,
		// then cannot decide fault is accompany or not, filter, and in que to decide in next turn.
		if !processor.isCurrentExceedDiagnosisTimeout(currentTime, accompanyFaultTime) {
			faultMap = device.DeleteFaultFromFaultMap(faultMap, fault)
			newDeviceFaultQue = append(newDeviceFaultQue, fault)
		}
	}
	return newDeviceFaultQue, faultMap
}

func (processor *UceAccompanyFaultProcessor) getDeviceUceFaultTime(nodeName, deviceName string) int64 {
	if faultTime, ok := processor.uceFaultTime[nodeName][deviceName]; ok {
		return faultTime
	}
	return constant.DeviceNotFault
}

func (processor *UceAccompanyFaultProcessor) isAccompaniedFaultByUce(
	uceFaultTime, uceAccompanyFaultTime int64) bool {
	return util.Abs(uceFaultTime-uceAccompanyFaultTime) <= processor.DiagnosisAccompanyTimeout
}

func (processor *UceAccompanyFaultProcessor) isCurrentExceedDiagnosisTimeout(
	currentTime, uceAccompanyFaultTime int64) bool {
	return uceAccompanyFaultTime < currentTime-processor.DiagnosisAccompanyTimeout
}

func (processor *UceAccompanyFaultProcessor) Process() {
	deviceInfos := processor.deviceCenter.GetDeviceInfos()
	processor.uceAccompanyFaultInQue(deviceInfos)
	currentTime := time.Now().UnixMilli()
	filteredFaultInfos := processor.filterFaultInfos(currentTime, deviceInfos)
	processor.deviceCenter.SetDeviceInfos(filteredFaultInfos)
}
