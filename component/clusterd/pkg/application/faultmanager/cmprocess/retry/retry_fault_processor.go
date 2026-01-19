// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package retry contain filtering fault handling method for uce, L2, and L3 faults
package retry

import (
	"fmt"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/faultdomain/collector"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podgroup"
)

var RetryProcessor *retryFaultProcessor

/*
The retryFaultProcessor process retry fault reporting information.
If the device fault is UCE fault, then determine whether the job running on the device can tolerate UCE faults.
If they can tolerate it, the reporting of the UCE fault should be delayed by 10 seconds.
*/

const linkdownTimeout = 180

type retryFaultProcessor struct {
	JobReportRecoverTimeout  int64
	JobReportCompleteTimeout int64

	// retryJob->jobInfo
	retryDevicesOfJob      map[string]constant.RetryJobInfo
	normalFaultDetailOfJob map[string]constant.DeviceFaultDetail
	// node->DeviceName->retryDeviceInfo
	retryDeviceOfNode   map[string]constant.RetryNodeInfo
	jobServerInfoMap    constant.JobServerInfoMap
	nodeDeviceCmMap     map[string]*constant.AdvanceDeviceFaultCm
	nodeSwitchCmMap     map[string]*constant.SwitchInfo
	onceRetryDeviceInfo map[string]map[string]constant.RetryDeviceInfo
	onceFaultMap        map[string]map[string]constant.DeviceFault
	linkdownDeviceFault map[string]map[string]constant.DeviceFault
	linkdownSwitchFault map[string]constant.SimpleSwitchFaultInfo
	hasLinkDown         bool
}

func init() {
	RetryProcessor = &retryFaultProcessor{
		JobReportRecoverTimeout:  constant.JobReportRecoverTimeout,
		JobReportCompleteTimeout: constant.JobReportCompleteTimeout,
		onceRetryDeviceInfo:      make(map[string]map[string]constant.RetryDeviceInfo),
		onceFaultMap:             make(map[string]map[string]constant.DeviceFault),
		linkdownDeviceFault:      make(map[string]map[string]constant.DeviceFault),
		nodeDeviceCmMap:          make(map[string]*constant.AdvanceDeviceFaultCm),
		nodeSwitchCmMap:          make(map[string]*constant.SwitchInfo),
		linkdownSwitchFault:      make(map[string]constant.SimpleSwitchFaultInfo),
	}
}

func (processor *retryFaultProcessor) initRetryDeviceFromNodeAndReportInfo(jobId, nodeName string,
	currentTime int64) constant.RetryNodeInfo {
	managerPlaneRetryNode := processor.retryDeviceOfNode[nodeName]
	devicesOfJobOnNode := processor.jobServerInfoMap.InfoMap[jobId][nodeName]
	deviceNumOfPod := pod.GetPodDeviceNumByJobId(jobId)
	jobRetryNodeInfo := constant.RetryNodeInfo{NodeName: nodeName, DeviceInfo: make(map[string]constant.RetryDeviceInfo)}
	reportTime := collector.ReportInfoCollector.GetNoRetryReportTime(jobId)
	hasReport := false
	for _, deviceOfJob := range devicesOfJobOnNode.DeviceList {
		deviceName := processor.nodeDeviceCmMap[nodeName].DeviceType + "-" + deviceOfJob.DeviceID
		retryReportInfo := collector.ReportInfoCollector.GetInfo(jobId, nodeName, deviceName)
		if retryReportInfo.FaultType == constant.HcclFaultType {
			hasReport = true
		}
		jobRetryDevice := constant.RetryDeviceInfo{
			DeviceName: deviceName, FaultDetail: make(map[string]constant.DeviceFaultDetail),
			FaultCodeLevel: make(map[string]string)}
		detailInfo := constant.DeviceFaultDetail{
			FaultTime: constant.DeviceNotFault, RecoverTime: retryReportInfo.RecoverTime,
			CompleteTime: retryReportInfo.CompleteTime, ReportTime: reportTime, FaultType: retryReportInfo.FaultType,
		}
		if retryDevice, ok := managerPlaneRetryNode.DeviceInfo[deviceName]; ok {
			jobRetryDevice.FaultCodeLevel = retryDevice.FaultCodeLevel
			// management plane found retry fault
			if retryDetail, ok1 := retryDevice.FaultDetail[constant.DeviceRetryFault]; ok1 {
				detailInfo.FaultTime = retryDetail.FaultTime
				detailInfo.FaultType = retryDetail.FaultType
				jobRetryDevice.FaultDetail[constant.DeviceRetryFault] = detailInfo
				jobRetryNodeInfo.DeviceInfo[deviceName] = jobRetryDevice
			}
			if retryDetail, ok1 := retryDevice.FaultDetail[constant.DeviceNormalFault]; ok1 {
				detailInfo.FaultTime = retryDetail.FaultTime
				detailInfo.FaultType = retryDetail.FaultType
				jobRetryDevice.FaultDetail[constant.DeviceNormalFault] = detailInfo
				jobRetryNodeInfo.DeviceInfo[deviceName] = jobRetryDevice
				podRank := common.CalculateStringDivInt(deviceOfJob.RankID, deviceNumOfPod)
				processor.updateNormalFaultDetailOfJob(jobId, &retryDetail, podRank, reportTime)
			}
		} else if faultdomain.ValidBusinessRetryReportInfo(&retryReportInfo) { // business plane found retry fault
			jobRetryDevice.FaultDetail[constant.DeviceRetryFault] = detailInfo
			jobRetryNodeInfo.DeviceInfo[deviceName] = jobRetryDevice
		}
		if hasReport {
			processor.clearFault(jobId)
			continue
		}
		processor.addOnceRetryDevices(nodeName, deviceName, currentTime, &jobRetryNodeInfo)
	}
	return jobRetryNodeInfo
}

func (processor *retryFaultProcessor) clearFault(jobID string) {
	for node, _ := range processor.jobServerInfoMap.InfoMap[jobID] {
		delete(processor.linkdownDeviceFault, node)
		delete(processor.linkdownSwitchFault, node)
		delete(processor.onceFaultMap, node)
		delete(processor.onceRetryDeviceInfo, node)
	}
}

func (processor *retryFaultProcessor) addOnceRetryDevices(nodeName, deviceName string,
	currentTime int64, jobRetryNodeInfo *constant.RetryNodeInfo) {
	for _, fault := range processor.onceRetryDeviceInfo[nodeName] {
		if fault.DeviceName == deviceName {
			if currentTime-fault.FaultDetail[constant.DeviceRetryFault].FaultTime <= constant.HCCLStepRetryTimeout {
				continue
			}
			jobRetryNodeInfo.DeviceInfo[deviceName] = fault
			hwlog.RunLog.Debugf("jobRetryNodeInfo: nodeName %s, deviceName%s ,add: %v ",
				nodeName, deviceName, fault)
		}
	}
}

func (processor *retryFaultProcessor) updateNormalFaultDetailOfJob(jobId string, detail *constant.DeviceFaultDetail,
	podRank int, reportTime int64) {
	hasRank0Fault := podRank == 0
	jobFaultDetail, ok := processor.normalFaultDetailOfJob[jobId]
	if !ok {
		processor.normalFaultDetailOfJob[jobId] = constant.DeviceFaultDetail{
			FaultTime:       detail.FaultTime,
			ReportTime:      reportTime,
			HasFaultAboveL3: detail.HasFaultAboveL3,
			HasRank0Fault:   hasRank0Fault,
		}
		return
	}
	jobFaultDetail.FaultTime = util.MinInt(jobFaultDetail.FaultTime, detail.FaultTime)
	jobFaultDetail.ReportTime = util.MinInt(jobFaultDetail.ReportTime, reportTime)
	jobFaultDetail.HasFaultAboveL3 = jobFaultDetail.HasFaultAboveL3 || detail.HasFaultAboveL3
	jobFaultDetail.HasRank0Fault = jobFaultDetail.HasRank0Fault || hasRank0Fault
	processor.normalFaultDetailOfJob[jobId] = jobFaultDetail
}

// Process retry, L2 and L3 fault
func (processor *retryFaultProcessor) Process(info any) any {
	deviceContent, deviceOk := info.(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	switchContent, switchOK := info.(constant.OneConfigmapContent[*constant.SwitchInfo])
	if !deviceOk && !switchOK {
		hwlog.RunLog.Errorf("%v cannot convert to DeviceInfo or SwitchInfo", info)
		return info
	}
	if switchOK {
		processor.nodeSwitchCmMap = switchContent.AllConfigmap
	}
	if deviceOk {
		processor.nodeDeviceCmMap = deviceContent.AllConfigmap
	}
	processor.jobServerInfoMap = job.GetJobServerInfoMap()
	hwlog.RunLog.Debugf("current nodeDeviceCmMap %v", processor.nodeDeviceCmMap)

	processor.retryDeviceOfNode = processor.handleRetryDeviceOfNodes()
	hwlog.RunLog.Debugf("current retryDeviceOfNode %v", processor.retryDeviceOfNode)

	currentTime := time.Now().UnixMilli()
	hwlog.RunLog.Debugf("currentTime %d", currentTime)

	processor.normalFaultDetailOfJob = make(map[string]constant.DeviceFaultDetail)
	processor.retryDevicesOfJob = processor.getRetryDevicesForTolerateJobs(currentTime)
	hwlog.RunLog.Debugf("current retryDevicesOfJob %v", processor.retryDevicesOfJob)

	processor.processRetryFaultInfo(currentTime)
	hwlog.RunLog.Debugf("normalFaultDetailOfJob: %v", processor.normalFaultDetailOfJob)
	hwlog.RunLog.Debugf("retryDevicesOfJob: %v", processor.retryDevicesOfJob)

	hwlog.RunLog.Debugf("result deviceInfos %v", deviceContent.AllConfigmap)
	hwlog.RunLog.Debugf("result switchInfos %v", switchContent.AllConfigmap)
	if switchOK {
		return switchContent
	}
	return deviceContent
}

func (processor *retryFaultProcessor) processRetryFaultInfo(currentTime int64) {
	for nodeName, advanceDeviceInfo := range processor.nodeDeviceCmMap {
		advanceDeviceInfo = processor.processEachNodeRetryFaultInfo(nodeName, advanceDeviceInfo, currentTime)
		processor.nodeDeviceCmMap[nodeName] = advanceDeviceInfo
	}
	clearLinkDown := true
	for _, devs := range processor.onceFaultMap {
		if len(devs) > 0 {
			clearLinkDown = false
			break
		}
	}
	if processor.hasLinkDown && clearLinkDown {
		processor.hasLinkDown = false
		processor.linkdownDeviceFault = make(map[string]map[string]constant.DeviceFault)
		processor.linkdownSwitchFault = make(map[string]constant.SimpleSwitchFaultInfo)
	}
}

func (processor *retryFaultProcessor) processEachNodeRetryFaultInfo(
	nodeName string, deviceInfo *constant.AdvanceDeviceFaultCm, currentTime int64) *constant.AdvanceDeviceFaultCm {
	modified := false
	for jobId, retryJob := range processor.retryDevicesOfJob {
		for deviceName, retryDevice := range retryJob.RetryNode[nodeName].DeviceInfo {
			log := fmt.Sprintf("device: %s on node %s, "+
				"currentTime: %s, ", retryDevice.DeviceName, nodeName, util.ReadableMsTime(currentTime))
			if detailInfo, ok := retryDevice.FaultDetail[constant.DeviceRetryFault]; ok &&
				processor.jobServerInfoMap.RetryTolerate[jobId] {
				fullLog := log + fmt.Sprintf("faultTime: %s, recoverTime: %s , faultType: %s ",
					util.ReadableMsTime(detailInfo.FaultTime),
					util.ReadableMsTime(detailInfo.RecoverTime),
					detailInfo.FaultType)
				if processor.canFilterRetryDeviceFaultInfo(retryDevice, currentTime) {
					hwlog.RunLog.Warn("retryProcessor filter retry " + fullLog)
					processor.filterRetryDeviceFaultInfo(deviceName, deviceInfo, nodeName)
					modified = true
				} else if retryDevice.FaultDetail[constant.DeviceRetryFault].FaultType == constant.HcclFaultType {
					hwlog.RunLog.Warn("retryProcessor cannot filter retry " + fullLog)
					processor.addRetryFault(nodeName, deviceInfo, deviceName)
					modified = true
				} else {
					hwlog.RunLog.Warn("retryProcessor cannot filter retry " + fullLog)
				}
			}
			if detailInfo, ok := retryDevice.FaultDetail[constant.DeviceNormalFault]; ok &&
				podgroup.JudgeRestartProcessByJobKey(jobId) {
				fullLog := log + fmt.Sprintf("faultTime: %s, recoverTime: %s , faultType: %s ",
					util.ReadableMsTime(detailInfo.FaultTime),
					util.ReadableMsTime(detailInfo.RecoverTime),
					detailInfo.FaultType)
				if processor.canFilterNormalDeviceFaultInfo(jobId, retryDevice, currentTime) {
					hwlog.RunLog.Warn("retryProcessor filter normal " + fullLog)
					processor.filterNormalDeviceFaultInfo(deviceName, deviceInfo)
					modified = true
				} else {
					hwlog.RunLog.Warn("retryProcessor cannot filter normal " + fullLog)
				}
			}
		}
	}
	if modified {
		deviceInfo.UpdateTime = time.Now().Unix()
		faultdomain.SortDataForAdvanceDeviceInfo(deviceInfo)
	}
	return deviceInfo
}

func (processor *retryFaultProcessor) addRetryFault(nodeName string,
	deviceInfo *constant.AdvanceDeviceFaultCm, deviceName string) {
	if len(processor.onceRetryDeviceInfo[nodeName]) == 0 {
		return
	}
	hwlog.RunLog.Infof("node %s processor.linkdownSwitch %v", nodeName, processor.linkdownSwitchFault[nodeName])
	hwlog.RunLog.Infof("node %s processor.linkdownDevice %v", nodeName, processor.linkdownDeviceFault[nodeName])
	for _, faultInfo := range processor.linkdownDeviceFault {
		if len(faultInfo) > 0 {
			processor.hasLinkDown = true
			break
		}
	}
	if len(processor.linkdownSwitchFault) > 0 {
		processor.hasLinkDown = true
	}
	if fault, ok := processor.onceFaultMap[nodeName][deviceName]; ok {
		if !processor.hasLinkDown {
			fault.FaultLevel = constant.RestartBusiness
			hwlog.RunLog.Warn("fault upgrade  to restartBusiness")
		}
		deviceInfo.AddFaultAndFix(fault)
		hwlog.RunLog.Warnf("nodeName :%v deviceName: %v, add hccl error: %v ", nodeName, deviceName, fault)
	}
	if fault, ok := processor.linkdownDeviceFault[nodeName][deviceName]; ok {
		fault.ForceAdd = true
		deviceInfo.AddFaultAndFix(fault)
		hwlog.RunLog.Warnf("nodeName :%v deviceName: %v, add linkdown error: %v ", nodeName, deviceName, fault)
	}
	if fault, ok := processor.linkdownSwitchFault[nodeName]; ok {
		fault.ForceAdd = true
		processor.nodeSwitchCmMap[constant.SwitchInfoPrefix+nodeName].AddFaultAndFix(fault)
		hwlog.RunLog.Warnf("nodeName :%v deviceName: %v, add switch error: %v ", nodeName, deviceName, fault)
	}
	delete(processor.onceRetryDeviceInfo[nodeName], deviceName)
	delete(processor.onceFaultMap[nodeName], deviceName)
}

func (processor *retryFaultProcessor) filterRetryDeviceFaultInfo(
	deviceName string, advanceDevInfo *constant.AdvanceDeviceFaultCm, nodeName string) {
	for _, fault := range advanceDevInfo.FaultDeviceList[deviceName] {
		// filter device's retry fault
		if faultdomain.IsUceFault(fault.FaultCode) || faultdomain.IsHcclRetryFault(fault.FaultCode) {
			advanceDevInfo.DelFaultAndFix(fault)
		}
		if faultdomain.IsHcclRetryFault(fault.FaultCode) {
			if _, ok := processor.onceFaultMap[nodeName]; !ok {
				processor.onceFaultMap[nodeName] = make(map[string]constant.DeviceFault)
			}
			processor.onceFaultMap[nodeName][fault.NPUName] = fault
			hwlog.RunLog.Debugf("processor.onceFaultMap add: %v", fault)
		}
	}
}

func (processor *retryFaultProcessor) filterNormalDeviceFaultInfo(
	deviceName string, advanceDevInfo *constant.AdvanceDeviceFaultCm) {
	for _, fault := range advanceDevInfo.FaultDeviceList[deviceName] {
		if faultdomain.IsL2L3Fault(fault.FaultLevel) {
			advanceDevInfo.DelFaultAndFix(fault)
		}
	}
}

func (processor *retryFaultProcessor) canFilterRetryDeviceFaultInfo(retryDevice constant.RetryDeviceInfo, currentTime int64) bool {
	detailInfo, ok := retryDevice.FaultDetail[constant.DeviceRetryFault]
	if !ok {
		return false
	}
	if processor.currentTimeIsNotExceedReportRecoverTimeout(detailInfo, currentTime) {
		return true
	}
	if processor.RecoverTimeIsNotExceedRecoverTimeout(detailInfo) {
		if processor.currentTimeIsNotExceedRecoverCompleteTimeout(detailInfo, currentTime) {
			return true
		} else if processor.reportCompleteTimeIsNotExceedCompleteTimeout(detailInfo) {
			return true
		}
		return false
	}
	return false
}

func (processor *retryFaultProcessor) canFilterNormalDeviceFaultInfo(jobId string, retryDevice constant.RetryDeviceInfo,
	currentTime int64) bool {
	jobFaultDetail, ok := processor.normalFaultDetailOfJob[jobId]
	if ok {
		if jobFaultDetail.HasFaultAboveL3 || jobFaultDetail.HasRank0Fault ||
			(jobFaultDetail.ReportTime != constant.JobShouldReportFault &&
				jobFaultDetail.ReportTime >= jobFaultDetail.FaultTime) {
			return false
		}
		return jobFaultDetail.FaultTime >= currentTime-constant.JobRestartInPlaceTimeout
	}
	detailInfo, ok := retryDevice.FaultDetail[constant.DeviceNormalFault]
	if !ok {
		return false
	}
	return detailInfo.FaultTime >= currentTime-constant.JobRestartInPlaceTimeout
}

func (processor *retryFaultProcessor) currentTimeIsNotExceedReportRecoverTimeout(
	detailInfo constant.DeviceFaultDetail, currentTime int64) bool {
	return detailInfo.FaultTime >= currentTime-processor.JobReportRecoverTimeout
}

func (processor *retryFaultProcessor) RecoverTimeIsNotExceedRecoverTimeout(
	detailInfo constant.DeviceFaultDetail) bool {
	return detailInfo.FaultTime >= detailInfo.RecoverTime-processor.JobReportRecoverTimeout
}

func (processor *retryFaultProcessor) currentTimeIsNotExceedRecoverCompleteTimeout(
	detailInfo constant.DeviceFaultDetail, currentTime int64) bool {
	return processor.JobReportCompleteTimeout+detailInfo.RecoverTime >= currentTime
}

func (processor *retryFaultProcessor) reportCompleteTimeIsNotExceedCompleteTimeout(
	detailInfo constant.DeviceFaultDetail) bool {
	// invalid complete time
	if detailInfo.CompleteTime < detailInfo.FaultTime || detailInfo.CompleteTime < detailInfo.RecoverTime {
		return false
	}
	return processor.JobReportCompleteTimeout+detailInfo.RecoverTime >= detailInfo.CompleteTime
}

func (processor *retryFaultProcessor) handleRetryDeviceOfNodes() map[string]constant.RetryNodeInfo {
	retryNodes := make(map[string]constant.RetryNodeInfo)
	for nodeName, deviceInfo := range processor.nodeDeviceCmMap {
		retryFaultDevicesOnNode := processor.getRetryFaultDevices(nodeName, deviceInfo)

		if len(retryFaultDevicesOnNode.DeviceInfo) == 0 {
			continue
		}
		retryNodes[nodeName] = retryFaultDevicesOnNode
		for _, retryDevice := range retryFaultDevicesOnNode.DeviceInfo {
			if retryDevice.FaultDetail[constant.DeviceRetryFault].FaultType != constant.HcclFaultType {
				continue
			}
			if _, ok := processor.onceRetryDeviceInfo[nodeName]; !ok {
				processor.onceRetryDeviceInfo[nodeName] = make(map[string]constant.RetryDeviceInfo)
			}
			processor.onceRetryDeviceInfo[nodeName][retryDevice.DeviceName] = retryDevice
		}
	}
	return retryNodes
}

func (processor *retryFaultProcessor) getRetryDevicesForTolerateJobs(curTime int64) map[string]constant.RetryJobInfo {
	nodeNameList := make([]string, 0)
	for key, _ := range processor.nodeDeviceCmMap {
		nodeNameList = append(nodeNameList, key)
	}
	retryJobs := make(map[string]constant.RetryJobInfo)
	for jobUid, serverList := range processor.jobServerInfoMap.InfoMap {
		if !processor.jobServerInfoMap.RetryTolerate[jobUid] && !podgroup.JudgeRestartProcessByJobKey(jobUid) {
			continue
		}
		jobInfo := constant.RetryJobInfo{
			RetryNode: make(map[string]constant.RetryNodeInfo),
			JobId:     jobUid,
		}
		for _, nodeName := range nodeNameList {
			devicesOfJobOnNode := serverList[nodeName]
			if len(devicesOfJobOnNode.DeviceList) == 0 {
				continue
			}
			jobInfo.RetryNode[nodeName] = processor.initRetryDeviceFromNodeAndReportInfo(jobUid, nodeName, curTime)
			processor.processSwitchLinkDownFaults(nodeName)
		}
		if len(jobInfo.RetryNode) != 0 {
			retryJobs[jobUid] = jobInfo
		}
	}
	return retryJobs
}

func (processor *retryFaultProcessor) processSwitchLinkDownFaults(nodeName string) {
	currentTime := time.Now().Unix()
	switchKey := constant.SwitchInfoPrefix + nodeName
	switchFault, ok := processor.linkdownSwitchFault[nodeName]
	if ok && currentTime-switchFault.AlarmRaisedTime >= linkdownTimeout {
		hwlog.RunLog.Infof("switch linkdownFault delete node: %v: %v", nodeName, switchFault)
		delete(processor.linkdownSwitchFault, nodeName)
	}
	switchInfo, ok := processor.nodeSwitchCmMap[switchKey]
	if !ok {
		return
	}
	for _, info := range switchInfo.SwitchFaultInfo.FaultInfo {
		if !faultdomain.IsSwitchLinkDownFault(info.AssembledFaultCode) {
			continue
		}
		if currentTime-info.AlarmRaisedTime >= linkdownTimeout {
			continue
		}
		processor.linkdownSwitchFault[nodeName] = info
	}
}

func (processor *retryFaultProcessor) getRetryFaultDevices(
	nodeName string, deviceInfo *constant.AdvanceDeviceFaultCm) constant.RetryNodeInfo {
	nodeInfo := constant.RetryNodeInfo{
		NodeName:   nodeName,
		DeviceInfo: make(map[string]constant.RetryDeviceInfo),
	}
	for _, deviceFaults := range deviceInfo.FaultDeviceList {
		for _, fault := range deviceFaults {
			if !faultdomain.IsUceFault(fault.FaultCode) && !faultdomain.IsHcclRetryFault(fault.FaultCode) &&
				faultdomain.IsL1Fault(fault.FaultLevel) {
				continue
			}
			faultDeviceInfo, ok := nodeInfo.DeviceInfo[fault.NPUName]
			if !ok {
				faultDeviceInfo = constant.RetryDeviceInfo{
					DeviceName:     fault.NPUName,
					FaultDetail:    make(map[string]constant.DeviceFaultDetail),
					FaultCodeLevel: make(map[string]string),
				}
			}
			errorMsg := fmt.Sprintf("getRetryFaultDevices cannot find retry fault time for device %s of node %s",
				deviceInfo.CmName, nodeName)
			faultTime := faultdomain.GetFaultTime(fault, errorMsg)
			detailInfo := constant.DeviceFaultDetail{
				FaultTime:    faultTime,
				RecoverTime:  constant.JobNotRecover,
				CompleteTime: constant.JobNotRecoverComplete,
				FaultType:    faultdomain.GetRetryTypeByFaultCode(fault.FaultCode),
				HasFaultAboveL3: !faultdomain.IsL2L3Fault(fault.FaultLevel) &&
					!faultdomain.IsL1Fault(fault.FaultLevel),
			}
			if faultdomain.IsUceFault(fault.FaultCode) || faultdomain.IsHcclRetryFault(fault.FaultCode) {
				faultDeviceInfo.FaultDetail[constant.DeviceRetryFault] = detailInfo
				faultDeviceInfo.FaultCodeLevel[fault.FaultCode] = fault.FaultLevel
			}
			if faultdomain.IsLinkDownFault(fault.FaultCode) {
				if _, ok := processor.linkdownDeviceFault[nodeName]; !ok {
					processor.linkdownDeviceFault[nodeName] = make(map[string]constant.DeviceFault)
				}
				processor.linkdownDeviceFault[nodeName][fault.NPUName] = fault
			}
			if !faultdomain.IsL1Fault(fault.FaultLevel) {
				if oldDetailInfo, ok := faultDeviceInfo.FaultDetail[constant.DeviceNormalFault]; ok {
					detailInfo.FaultTime = util.MinInt(oldDetailInfo.FaultTime, detailInfo.FaultTime)
					detailInfo.HasFaultAboveL3 = detailInfo.HasFaultAboveL3 || oldDetailInfo.HasFaultAboveL3
				}
				faultDeviceInfo.FaultDetail[constant.DeviceNormalFault] = detailInfo
				faultDeviceInfo.FaultCodeLevel[fault.FaultCode] = fault.FaultLevel
			}
			nodeInfo.DeviceInfo[fault.NPUName] = faultDeviceInfo
		}
	}
	return nodeInfo
}

func (processor *retryFaultProcessor) GetRetryDeviceFromJob(jobId, nodeName, deviceName string) (constant.RetryDeviceInfo, bool) {
	jobInfo, found := processor.retryDevicesOfJob[jobId]
	if !found {
		hwlog.RunLog.Debugf("job %s has no retry fault", jobId)
		return constant.RetryDeviceInfo{}, false
	}
	retryDevice, found := jobInfo.RetryNode[nodeName].DeviceInfo[deviceName]
	if !found {
		hwlog.RunLog.Debugf("job %s's retry fault is not on node %s device %s", jobId, nodeName, deviceName)
		return constant.RetryDeviceInfo{}, false
	}
	return retryDevice, true
}

// CanDoRestartInPlace judge job can restart fault process in place
func (processor *retryFaultProcessor) CanDoRestartInPlace(jobId string) bool {
	jobFaultDetail, ok := processor.normalFaultDetailOfJob[jobId]
	if !ok {
		return false
	}
	if jobFaultDetail.HasFaultAboveL3 || jobFaultDetail.HasRank0Fault ||
		(jobFaultDetail.ReportTime != constant.JobShouldReportFault &&
			jobFaultDetail.ReportTime > jobFaultDetail.FaultTime) {
		return false
	}
	return jobFaultDetail.FaultTime >= time.Now().UnixMilli()-constant.JobRestartInPlaceTimeout
}

// GetFilterFaultCodeAndLevel get filtered fault info
func (processor *retryFaultProcessor) GetFilterFaultCodeAndLevel(jobId, nodeName, deviceName string) map[string]string {
	filterDevice, found := processor.GetRetryDeviceFromJob(jobId, nodeName, deviceName)
	if !found {
		return nil
	}
	return filterDevice.FaultCodeLevel
}

// JobHasFault judge job has fault
func (processor *retryFaultProcessor) JobHasFault(jobId string) bool {
	filterJob, found := processor.retryDevicesOfJob[jobId]
	if !found {
		return false
	}
	for _, filterNode := range filterJob.RetryNode {
		for _, filterDevice := range filterNode.DeviceInfo {
			if len(filterDevice.FaultCodeLevel) > 0 {
				return true
			}
		}
	}
	return false
}
