// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package uce contain uce process method
package uce

import (
	"fmt"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/collector"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/job"
)

var UceProcessor *UceFaultProcessor

/*
The UceFaultProcessor process uce fault reporting information.
If the device fault is UCE fault, then determine whether the job running on the device can tolerate UCE faults.
If they can tolerate it, the reporting of the UCE fault should be delayed by 10 seconds.
*/
type UceFaultProcessor struct {
	JobReportRecoverTimeout  int64
	JobReportCompleteTimeout int64

	// uceJob->jobInfo
	uceDevicesOfUceJob map[string]constant.UceJobInfo
	// node->DeviceName->uceDeviceInfo
	uceDeviceOfNode  map[string]constant.UceNodeInfo
	jobServerInfoMap constant.JobServerInfoMap
	nodeDeviceCmMap  map[string]constant.AdvanceDeviceFaultCm
}

func NewUceFaultProcessor() *UceFaultProcessor {
	UceProcessor = &UceFaultProcessor{
		JobReportRecoverTimeout:  constant.JobReportRecoverTimeout,
		JobReportCompleteTimeout: constant.JobReportCompleteTimeout,
	}
	return UceProcessor
}

func (processor *UceFaultProcessor) initUceDeviceFromNodeAndReportInfo(jobId string, nodeName string) constant.UceNodeInfo {
	managerPlaneUceNode := processor.uceDeviceOfNode[nodeName]
	devicesOfJobOnNode := processor.jobServerInfoMap.InfoMap[jobId][nodeName]
	jobUceNodeInfo := constant.UceNodeInfo{
		NodeName:   nodeName,
		DeviceInfo: make(map[string]constant.UceDeviceInfo),
	}

	for _, deviceOfJob := range devicesOfJobOnNode.DeviceList {
		deviceName := processor.nodeDeviceCmMap[nodeName].ServerType + "-" + deviceOfJob.DeviceID
		uceReportInfo := collector.ReportInfoCollector.GetInfo(jobId, nodeName, deviceName)
		jobUceDevice := constant.UceDeviceInfo{
			DeviceName:   deviceName,
			FaultTime:    constant.DeviceNotFault,
			RecoverTime:  uceReportInfo.RecoverTime,
			CompleteTime: uceReportInfo.CompleteTime,
		}
		// management plane found uce fault
		if uceDevice, ok := managerPlaneUceNode.DeviceInfo[deviceName]; ok {
			jobUceDevice.FaultTime = uceDevice.FaultTime
			jobUceNodeInfo.DeviceInfo[deviceName] = jobUceDevice
		} else if faultdomain.ValidBusinessUceReportInfo(&uceReportInfo) { // business plane found uce fault
			jobUceNodeInfo.DeviceInfo[deviceName] = jobUceDevice
		}
	}

	return jobUceNodeInfo
}

func (processor *UceFaultProcessor) Process(info any) any {
	processContent, ok := info.(constant.CenterProcessContent[*constant.DeviceInfo])
	if !ok {
		hwlog.RunLog.Errorf("%v cannot convert to DeviceInfo", info)
		return info
	}
	if len(processContent.UpdateConfigmap) == 0 {
		return info
	}
	deviceInfos := processContent.AllConfigmap

	processor.jobServerInfoMap = job.GetJobServerInfoMap()
	processor.nodeDeviceCmMap = faultdomain.GetAdvanceDeviceCmForNodeMap(deviceInfos)
	hwlog.RunLog.Debugf("current deviceInfos %s", util.ObjToString(deviceInfos))
	hwlog.RunLog.Debugf("current nodeDeviceCmMap %s", util.ObjToString(processor.nodeDeviceCmMap))

	processor.uceDeviceOfNode = processor.getUceDeviceOfNodes()
	hwlog.RunLog.Debugf("current uceDeviceOfNode %s", util.ObjToString(processor.uceDeviceOfNode))

	processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs()
	hwlog.RunLog.Debugf("current uceDevicesOfUceJob %s", util.ObjToString(processor.uceDevicesOfUceJob))

	currentTime := time.Now().UnixMilli()
	processor.processUceFaultInfo(currentTime)
	faultdomain.AdvanceDeviceCmForNodeMapToString(processor.nodeDeviceCmMap, deviceInfos)

	hwlog.RunLog.Debugf("result deviceInfos %s", util.ObjToString(deviceInfos))
	processContent.AllConfigmap = deviceInfos
	return processContent
}

func (processor *UceFaultProcessor) processUceFaultInfo(currentTime int64) {
	for nodeName, advanceDeviceInfo := range processor.nodeDeviceCmMap {
		advanceDeviceInfo = processor.processEachNodeUceFaultInfo(nodeName, advanceDeviceInfo, currentTime)
		processor.nodeDeviceCmMap[nodeName] = advanceDeviceInfo
	}
}

func (processor *UceFaultProcessor) processEachNodeUceFaultInfo(
	nodeName string, deviceInfo constant.AdvanceDeviceFaultCm, currentTime int64) constant.AdvanceDeviceFaultCm {
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

func (processor *UceFaultProcessor) filterUceDeviceFaultInfo(
	deviceName string, deviceFaultMap map[string][]constant.DeviceFault) map[string][]constant.DeviceFault {
	for _, fault := range deviceFaultMap[deviceName] {
		// filter device's uce fault
		if faultdomain.IsUceFault(fault.FaultCode) {
			deviceFaultMap = faultdomain.DeleteFaultFromFaultMap(deviceFaultMap, fault)
		}
	}
	return deviceFaultMap
}

func (processor *UceFaultProcessor) canFilterUceDeviceFaultInfo(uceDevice constant.UceDeviceInfo, currentTime int64) bool {
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

func (processor *UceFaultProcessor) currentTimeIsNotExceedReportRecoverTimeout(
	uceDevice constant.UceDeviceInfo, currentTime int64) bool {
	return uceDevice.FaultTime >= currentTime-processor.JobReportRecoverTimeout
}

func (processor *UceFaultProcessor) RecoverTimeIsNotExceedRecoverTimeout(
	uceDevice constant.UceDeviceInfo) bool {
	return uceDevice.FaultTime >= uceDevice.RecoverTime-processor.JobReportRecoverTimeout
}

func (processor *UceFaultProcessor) currentTimeIsNotExceedRecoverCompleteTimeout(
	uceDevice constant.UceDeviceInfo, currentTime int64) bool {
	return processor.JobReportCompleteTimeout+uceDevice.RecoverTime >= currentTime
}

func (processor *UceFaultProcessor) reportCompleteTimeIsNotExceedCompleteTimeout(
	uceDevice constant.UceDeviceInfo) bool {
	// invalid complete time
	if uceDevice.CompleteTime < uceDevice.FaultTime || uceDevice.CompleteTime < uceDevice.RecoverTime {
		return false
	}
	return processor.JobReportCompleteTimeout+uceDevice.RecoverTime >= uceDevice.CompleteTime
}

func (processor *UceFaultProcessor) getUceDeviceOfNodes() map[string]constant.UceNodeInfo {
	uceNodes := make(map[string]constant.UceNodeInfo)
	for nodeName, deviceInfo := range processor.nodeDeviceCmMap {
		uceFaultDevicesOnNode := processor.getUceFaultDevices(nodeName, deviceInfo)

		if len(uceFaultDevicesOnNode.DeviceInfo) == 0 {
			continue
		}
		uceNodes[nodeName] = uceFaultDevicesOnNode
	}
	return uceNodes
}

func (processor *UceFaultProcessor) getUceDevicesForUceTolerateJobs() map[string]constant.UceJobInfo {
	nodeNameList := make([]string, 0)
	for key, _ := range processor.nodeDeviceCmMap {
		nodeNameList = append(nodeNameList, key)
	}
	uceJobs := make(map[string]constant.UceJobInfo)
	for jobUid, serverList := range processor.jobServerInfoMap.InfoMap {
		if !processor.jobServerInfoMap.UceTolerate[jobUid] {
			continue
		}
		jobInfo := constant.UceJobInfo{
			UceNode: make(map[string]constant.UceNodeInfo),
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

func (processor *UceFaultProcessor) getUceFaultDevices(nodeName string, deviceInfo constant.AdvanceDeviceFaultCm) constant.UceNodeInfo {
	nodeInfo := constant.UceNodeInfo{
		NodeName:   nodeName,
		DeviceInfo: make(map[string]constant.UceDeviceInfo),
	}
	for _, deviceFaults := range deviceInfo.FaultDeviceList {
		for _, fault := range deviceFaults {
			if !faultdomain.IsUceFault(fault.FaultCode) {
				continue
			}
			errorMsg := fmt.Sprintf("getUceFaultDevices cannot find uce fault time for device %s of node %s",
				deviceInfo.CmName, nodeName)
			faultTime := faultdomain.GetFaultTime(fault, errorMsg)
			nodeInfo.DeviceInfo[fault.NPUName] = constant.UceDeviceInfo{
				DeviceName:   fault.NPUName,
				FaultTime:    faultTime,
				RecoverTime:  constant.JobNotRecover,
				CompleteTime: constant.JobNotRecoverComplete,
			}
		}
	}
	return nodeInfo
}

func (processor *UceFaultProcessor) GetUceDeviceFromJob(jobId, nodeName, deviceName string) (constant.UceDeviceInfo, bool) {
	jobInfo, found := processor.uceDevicesOfUceJob[jobId]
	if !found {
		hwlog.RunLog.Debugf("job %s has no uce fault", jobId)
		return constant.UceDeviceInfo{}, false
	}
	uceDevice, found := jobInfo.UceNode[nodeName].DeviceInfo[deviceName]
	if !found {
		hwlog.RunLog.Debugf("job %s's uce fault is not on node %s device %s", jobId, nodeName, deviceName)
		return constant.UceDeviceInfo{}, false
	}
	return uceDevice, true
}
