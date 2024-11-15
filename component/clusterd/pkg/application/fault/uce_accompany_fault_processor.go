package fault

import (
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

// uceAccompanyFaultProcessor:
// aic aiv fault can be 1) accompanied by uce fault, also can 2) curr alone.
// if 1) aic aiv fault should be filtered. Once find aic fault, check if there is an uce fault 5s ago
// if 2) aic aiv fault should not be retained.
type uceAccompanyFaultProcessor struct {
	deviceCenter *deviceFaultProcessCenter
	// maintain 5s ago device info
	DiagnosisAccompanyTimeout int64
	// nodeName -> deviceName -> faultQue
	uceAccompanyFaultQue map[string]map[string][]constant.DeviceFault
	// uceFaultTime
	uceFaultTime       map[string]map[string]int64
	deviceCmForNodeMap map[string]AdvanceDeviceCm
}

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
	nodeName string, deviceInfo AdvanceDeviceCm) {
	if _, ok := processor.uceAccompanyFaultQue[nodeName]; !ok {
		processor.uceAccompanyFaultQue[nodeName] = make(map[string][]constant.DeviceFault)
	}
	if _, ok := processor.uceFaultTime[nodeName]; !ok {
		processor.uceFaultTime[nodeName] = make(map[string]int64)
	}
	for deviceName, deviceFaults := range deviceInfo.DeviceList {
		for _, fault := range deviceFaults {
			if isUceFault(fault) {
				hwlog.RunLog.Debugf("find uce fault %s, on node %s", util.ObjToString(fault), nodeName)
				processor.uceFaultTime[nodeName][deviceName] = fault.FaultTime
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
		// in que
		hwlog.RunLog.Infof("find uce accompany like fault %s, on node %s", util.ObjToString(fault), nodeName)
		processor.uceAccompanyFaultQue[nodeName][deviceName] = append(faultsInfo, fault)
	}
}

func (processor *uceAccompanyFaultProcessor) filterFaultInfos(currentTime int64) {
	for nodeName, nodeFaults := range processor.uceAccompanyFaultQue {
		faultMap := processor.deviceCmForNodeMap[nodeName]
		for deviceName, deviceFaultQue := range nodeFaults {
			newQue, newFaultMap :=
				processor.filterFaultDevice(faultMap.DeviceList, currentTime, nodeName, deviceName, deviceFaultQue)
			nodeFaults[deviceName] = newQue
			faultMap.DeviceList = newFaultMap
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
		accompanyFaultTime := fault.FaultTime
		// if is accompanied fault, filter
		if processor.isAccompaniedFaultByUce(uceFaultTime, accompanyFaultTime) {
			hwlog.RunLog.Infof("filter uce accompany fault %s", util.ObjToString(fault))
			hwlog.RunLog.Infof("uceFaultTime %d, accompanyFaultTime %d", uceFaultTime, accompanyFaultTime)
			faultMap = deleteFaultFromFaultMap(faultMap, fault)
			continue
		}
		// if current is not exceed diagnosis time,
		// then cannot decide fault is accompany or not, filter, and in que to decide in next turn.
		if !processor.isCurrentExceedDiagnosisTimeout(currentTime, accompanyFaultTime) {
			hwlog.RunLog.Infof("filter uce accompany like fault %s", util.ObjToString(fault))
			hwlog.RunLog.Infof("currentTime %d, accompanyFaultTime %d", currentTime, accompanyFaultTime)
			faultMap = deleteFaultFromFaultMap(faultMap, fault)
			newDeviceFaultQue = append(newDeviceFaultQue, fault)
		}
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
	deviceInfos := processor.deviceCenter.getInfoMap()
	processor.deviceCmForNodeMap = getAdvanceDeviceCmForNodeMap(deviceInfos)
	hwlog.RunLog.Infof("current deviceInfos: %s", util.ObjToString(deviceInfos))
	hwlog.RunLog.Infof("current deviceCmForNodeMap: %s", util.ObjToString(processor.deviceCmForNodeMap))

	processor.uceAccompanyFaultInQue()
	hwlog.RunLog.Infof("current uceAccompanyFaultQue: %s", util.ObjToString(processor.uceAccompanyFaultQue))
	currentTime := time.Now().UnixMilli()
	hwlog.RunLog.Infof("currentTime: %d", currentTime)

	processor.filterFaultInfos(currentTime)
	advanceDeviceCmForNodeMapToString(processor.deviceCmForNodeMap, deviceInfos)

	hwlog.RunLog.Infof("uceAccompanyFaultProcessor result: %s", util.ObjToString(deviceInfos))
	processor.deviceCenter.setInfoMap(deviceInfos)
}
