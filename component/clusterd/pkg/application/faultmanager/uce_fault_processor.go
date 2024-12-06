// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"fmt"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

func newUceFaultProcessor(deviceCenter *deviceFaultProcessCenter) *uceFaultProcessor {
	return &uceFaultProcessor{
		JobReportRecoverTimeout:  constant.JobReportRecoverTimeout,
		JobReportCompleteTimeout: constant.JobReportCompleteTimeout,
		reportInfo: &reportInfosForAllJobs{
			InfoMap: make(map[string]map[string]map[string]reportInfo),
			RwMutex: sync.RWMutex{},
		},
		deviceCenter: deviceCenter,
	}
}

func (reportInfos *reportInfosForAllJobs) getInfo(jobId, nodeName, deviceName string) reportInfo {
	if reportInfos == nil {
		return reportInfo{
			RecoverTime:  constant.JobNotRecover,
			CompleteTime: constant.JobNotRecoverComplete,
		}
	}
	reportInfos.RwMutex.RLock()
	defer reportInfos.RwMutex.RUnlock()
	if info, ok := reportInfos.InfoMap[jobId][nodeName][deviceName]; ok {
		return info
	}
	return reportInfo{
		RecoverTime:  constant.JobNotRecover,
		CompleteTime: constant.JobNotRecoverComplete,
	}
}

func (processor *uceFaultProcessor) initUceDeviceFromNodeAndReportInfo(jobId string, nodeName string) uceNodeInfo {
	uceNode := processor.uceDeviceOfNode[nodeName]
	devicesOfJobOnNode := processor.jobServerInfoMap.InfoMap[jobId][nodeName]
	jobUceNodeInfo := uceNodeInfo{
		NodeName:   uceNode.NodeName,
		DeviceInfo: make(map[string]uceDeviceInfo),
	}

	for _, deviceOfJob := range devicesOfJobOnNode.DeviceList {
		deviceName := processor.nodeDeviceCmMap[nodeName].ServerType + "-" + deviceOfJob.DeviceID
		if uceDevice, ok := uceNode.DeviceInfo[deviceName]; ok {
			reportInfo := processor.reportInfo.getInfo(jobId, uceNode.NodeName, deviceName)
			jobUceNodeInfo.DeviceInfo[uceDevice.DeviceName] = uceDeviceInfo{
				DeviceName:   deviceName,
				FaultTime:    uceDevice.FaultTime,
				RecoverTime:  reportInfo.RecoverTime,
				CompleteTime: reportInfo.CompleteTime,
			}
		}
	}

	return jobUceNodeInfo
}

func (processor *uceFaultProcessor) process() {
	processor.jobServerInfoMap = processor.deviceCenter.jobServerInfoMap
	deviceInfos := processor.deviceCenter.getProcessingCm()
	processor.nodeDeviceCmMap = getAdvanceDeviceCmForNodeMap(deviceInfos)
	hwlog.RunLog.Debugf("current deviceInfos %s", util.ObjToString(deviceInfos))
	hwlog.RunLog.Debugf("current nodeDeviceCmMap %s", util.ObjToString(processor.nodeDeviceCmMap))

	processor.uceDeviceOfNode = processor.getUceDeviceOfNodes()
	hwlog.RunLog.Debugf("current uceDeviceOfNode %s", util.ObjToString(processor.uceDeviceOfNode))

	processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs()
	hwlog.RunLog.Debugf("current uceDevicesOfUceJob %s", util.ObjToString(processor.uceDevicesOfUceJob))

	currentTime := time.Now().UnixMilli()
	processor.processUceFaultInfo(currentTime)
	advanceDeviceCmForNodeMapToString(processor.nodeDeviceCmMap, deviceInfos)

	hwlog.RunLog.Debugf("result deviceInfos %s", util.ObjToString(deviceInfos))
	processor.deviceCenter.setProcessingCm(deviceInfos)
}

func (processor *uceFaultProcessor) processUceFaultInfo(currentTime int64) {
	for nodeName, advanceDeviceInfo := range processor.nodeDeviceCmMap {
		advanceDeviceInfo = processor.processEachNodeUceFaultInfo(nodeName, advanceDeviceInfo, currentTime)
		processor.nodeDeviceCmMap[nodeName] = advanceDeviceInfo
	}
}

func (processor *uceFaultProcessor) processEachNodeUceFaultInfo(
	nodeName string, deviceInfo AdvanceDeviceFaultCm, currentTime int64) AdvanceDeviceFaultCm {
	for _, uceJob := range processor.uceDevicesOfUceJob {
		for deviceName, uceDevice := range uceJob.UceNode[nodeName].DeviceInfo {
			log := fmt.Sprintf("filter uce device: %s on node %s, "+
				"currentTime: %s, faultTime: %s, recoverTime: %s",
				uceDevice.DeviceName, nodeName,
				util.ReadableMsTime(currentTime),
				util.ReadableMsTime(uceDevice.FaultTime),
				util.ReadableMsTime(uceDevice.RecoverTime))
			if processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime) {
				hwlog.RunLog.Warn("uceFaultProcessor " + log)
				deviceInfo.FaultDeviceList = processor.filterUceDeviceFaultInfo(deviceName, deviceInfo.FaultDeviceList)
			} else {
				hwlog.RunLog.Warn("uceFaultProcessor cannot " + log)
			}
		}
	}
	return deviceInfo
}

func (processor *uceFaultProcessor) filterUceDeviceFaultInfo(
	deviceName string, deviceFaultMap map[string][]constant.DeviceFault) map[string][]constant.DeviceFault {
	for _, fault := range deviceFaultMap[deviceName] {
		// filter device's uce fault
		if isUceFault(fault) {
			deviceFaultMap = deleteFaultFromFaultMap(deviceFaultMap, fault)
		}
	}
	return deviceFaultMap
}

func (processor *uceFaultProcessor) canFilterUceDeviceFaultInfo(uceDevice uceDeviceInfo, currentTime int64) bool {
	if processor.currentTimeIsNotExceedReportRecoverTimeout(uceDevice, currentTime) {
		return true
	}
	if processor.RecoverTimeIsNotExceedRecoverTimeout(uceDevice) {
		if processor.currentTimeIsNotExceedRecoverCompleteTimeout(uceDevice, currentTime) {
			return true
		} else if processor.reportCompleteTimeIsNotExceedCompleteTimeout(uceDevice) {
			return true
		}
		return false
	}
	return false
}

func (processor *uceFaultProcessor) currentTimeIsNotExceedReportRecoverTimeout(
	uceDevice uceDeviceInfo, currentTime int64) bool {
	return uceDevice.FaultTime >= currentTime-processor.JobReportRecoverTimeout
}

func (processor *uceFaultProcessor) RecoverTimeIsNotExceedRecoverTimeout(
	uceDevice uceDeviceInfo) bool {
	return uceDevice.FaultTime >= uceDevice.RecoverTime-processor.JobReportRecoverTimeout
}

func (processor *uceFaultProcessor) currentTimeIsNotExceedRecoverCompleteTimeout(
	uceDevice uceDeviceInfo, currentTime int64) bool {
	return processor.JobReportCompleteTimeout+uceDevice.RecoverTime >= currentTime
}

func (processor *uceFaultProcessor) reportCompleteTimeIsNotExceedCompleteTimeout(
	uceDevice uceDeviceInfo) bool {
	// invalid complete time
	if uceDevice.CompleteTime < uceDevice.FaultTime || uceDevice.CompleteTime < uceDevice.RecoverTime {
		return false
	}
	return processor.JobReportCompleteTimeout+uceDevice.RecoverTime >= uceDevice.CompleteTime
}

func (processor *uceFaultProcessor) getUceDeviceOfNodes() map[string]uceNodeInfo {
	uceNodes := make(map[string]uceNodeInfo)
	for nodeName, deviceInfo := range processor.nodeDeviceCmMap {
		uceFaultDevicesOnNode := processor.getUceFaultDevices(nodeName, deviceInfo)

		if len(uceFaultDevicesOnNode.DeviceInfo) == 0 {
			continue
		}
		uceNodes[nodeName] = uceFaultDevicesOnNode
	}
	return uceNodes
}

func (processor *uceFaultProcessor) getUceDevicesForUceTolerateJobs() map[string]uceJobInfo {
	nodeNameList := make([]string, 0)
	for key, _ := range processor.nodeDeviceCmMap {
		nodeNameList = append(nodeNameList, key)
	}
	uceJobs := make(map[string]uceJobInfo)
	for jobUid, serverList := range processor.jobServerInfoMap.InfoMap {
		if !processor.jobServerInfoMap.UceTolerate[jobUid] {
			continue
		}
		jobInfo := uceJobInfo{
			UceNode: make(map[string]uceNodeInfo),
			JobId:   jobUid,
		}
		for _, nodeName := range nodeNameList {
			devicesOfJobOnNode := serverList[nodeName]
			if len(devicesOfJobOnNode.DeviceList) == 0 {
				continue
			}
			jobInfo.UceNode[nodeName] =
				processor.initUceDeviceFromNodeAndReportInfo(jobUid, nodeName)

		}
		if len(jobInfo.UceNode) != 0 {
			uceJobs[jobUid] = jobInfo
		}
	}
	return uceJobs
}

func (processor *uceFaultProcessor) getUceFaultDevices(nodeName string, deviceInfo AdvanceDeviceFaultCm) uceNodeInfo {
	nodeInfo := uceNodeInfo{
		NodeName:   nodeName,
		DeviceInfo: make(map[string]uceDeviceInfo),
	}
	for _, deviceFaults := range deviceInfo.FaultDeviceList {
		for _, fault := range deviceFaults {
			if !isUceFault(fault) {
				continue
			}
			errorMsg := fmt.Sprintf("getUceFaultDevices cannot find uce fault time for device %s of node %s",
				deviceInfo.CmName, nodeName)
			faultTime := getFaultTime(fault, errorMsg)
			nodeInfo.DeviceInfo[fault.NPUName] = uceDeviceInfo{
				DeviceName:   fault.NPUName,
				FaultTime:    faultTime,
				RecoverTime:  constant.JobNotRecover,
				CompleteTime: constant.JobNotRecoverComplete,
			}
		}
	}
	return nodeInfo
}

func (processor *uceFaultProcessor) reportUceInfo(jobId string, rankId string, recoverTime int64) error {
	nodeName, deviceId, err := getNodeAndDeviceFromJobIdAndRankId(jobId, rankId, processor.jobServerInfoMap)
	if err != nil {
		err = fmt.Errorf("report info failed, exception: %v", err)
		hwlog.RunLog.Error(err)
		return err
	}
	deviceName := processor.nodeDeviceCmMap[nodeName].ServerType + "-" + deviceId
	processor.reportInfo.RwMutex.Lock()
	defer processor.reportInfo.RwMutex.Unlock()
	infoMap := processor.reportInfo.InfoMap
	info := reportInfo{
		RecoverTime:  recoverTime,
		CompleteTime: constant.JobNotRecoverComplete,
	}
	if infoMap == nil {
		infoMap = make(map[string]map[string]map[string]reportInfo)
	}
	if _, ok := infoMap[jobId]; !ok {
		infoMap[jobId] = make(map[string]map[string]reportInfo)
		if _, ok := infoMap[jobId][nodeName]; !ok {
			infoMap[jobId][nodeName] = make(map[string]reportInfo)
		}
		infoMap[jobId][nodeName][deviceName] = info
	} else {
		if _, ok := infoMap[jobId][nodeName]; !ok {
			infoMap[jobId][nodeName] = make(map[string]reportInfo)
		}
		infoMap[jobId][nodeName][deviceName] = info
	}
	processor.reportInfo.InfoMap = infoMap
	hwlog.RunLog.Infof("callbackForReportUceInfo receive report info(%s, %s, %d)", jobId, rankId, recoverTime)
	hwlog.RunLog.Debugf("Current reportInfo is %s", util.ObjToString(processor.reportInfo.InfoMap))
	return nil
}
