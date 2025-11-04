// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package recoverinplace contain filtering fault handling method for single process fault
package recoverinplace

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

var RecoverInplaceProcessor *recoverInplaceFaultProcessor

type recoverInplaceFaultProcessor struct {
	JobReportRecoverTimeout  int64
	JobReportCompleteTimeout int64
	DevicesOfJob             map[string]constant.SingleProcessJobInfo
	normalFaultDetailOfJob   map[string]constant.DeviceFaultDetail
	DeviceOfNode             map[string]constant.SingleProcessNodeInfo
	jobServerInfoMap         constant.JobServerInfoMap
	nodeDeviceCmMap          map[string]*constant.AdvanceDeviceFaultCm
}

func init() {
	RecoverInplaceProcessor = &recoverInplaceFaultProcessor{
		JobReportRecoverTimeout:  constant.JobReportRecoverTimeout,
		JobReportCompleteTimeout: constant.JobReportCompleteTimeout,
		nodeDeviceCmMap:          make(map[string]*constant.AdvanceDeviceFaultCm),
	}
}

func (processor *recoverInplaceFaultProcessor) initDeviceFromNodeAndReportInfo(jobId,
	nodeName string) constant.SingleProcessNodeInfo {
	managerPlaneFaultNode := processor.DeviceOfNode[nodeName]
	devicesOfJobOnNode := processor.jobServerInfoMap.InfoMap[jobId][nodeName]
	deviceNumOfPod := pod.GetPodDeviceNumByJobId(jobId)
	jobSingleProcessNodeInfo := constant.SingleProcessNodeInfo{NodeName: nodeName,
		DeviceInfo: make(map[string]constant.SingleProcessDeviceInfo)}
	reportTime := collector.ReportInfoCollector.GetSingleProcessFaultReportTime(jobId)
	for _, deviceOfJob := range devicesOfJobOnNode.DeviceList {
		deviceName := processor.nodeDeviceCmMap[nodeName].DeviceType + "-" + deviceOfJob.DeviceID
		jobDevice := constant.SingleProcessDeviceInfo{
			DeviceName: deviceName, FaultDetail: constant.DeviceFaultDetail{}, FaultCodeLevel: make(map[string]string),
		}
		detailInfo := constant.DeviceFaultDetail{
			FaultTime:  constant.DeviceNotFault,
			ReportTime: reportTime,
		}
		if faultDevice, ok := managerPlaneFaultNode.DeviceInfo[deviceName]; ok {
			jobDevice.FaultCodeLevel = faultDevice.FaultCodeLevel
			faultDetail := faultDevice.FaultDetail
			detailInfo.FaultTime = faultDetail.FaultTime
			jobDevice.FaultDetail = detailInfo
			jobSingleProcessNodeInfo.DeviceInfo[deviceName] = jobDevice
			podRank := common.CalculateStringDivInt(deviceOfJob.RankID, deviceNumOfPod)
			processor.updateNormalFaultDetailOfJob(jobId, &faultDetail, podRank, reportTime)
		}
	}
	return jobSingleProcessNodeInfo
}

func (processor *recoverInplaceFaultProcessor) updateNormalFaultDetailOfJob(jobId string, detail *constant.DeviceFaultDetail,
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

// Process L2 and L3 fault
func (processor *recoverInplaceFaultProcessor) Process(info any) any {
	deviceContent, deviceOk := info.(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	if !deviceOk {
		hwlog.RunLog.Errorf("%v cannot convert to DeviceInfo or SwitchInfo", info)
		return info
	}

	processor.nodeDeviceCmMap = deviceContent.AllConfigmap
	processor.jobServerInfoMap = job.GetJobServerInfoMap()
	hwlog.RunLog.Debugf("current nodeDeviceCmMap %v", processor.nodeDeviceCmMap)

	processor.DeviceOfNode = processor.handleDeviceOfNodes()
	hwlog.RunLog.Debugf("current DeviceOfNode %v", processor.DeviceOfNode)

	currentTime := time.Now().UnixMilli()
	hwlog.RunLog.Debugf("currentTime %d", currentTime)

	processor.normalFaultDetailOfJob = make(map[string]constant.DeviceFaultDetail)
	processor.DevicesOfJob = processor.getDevicesForTolerateJobs()
	hwlog.RunLog.Debugf("current DevicesOfJob %v", processor.DevicesOfJob)

	processor.processFaultInfo(currentTime)
	hwlog.RunLog.Debugf("normalFaultDetailOfJob: %v", processor.normalFaultDetailOfJob)
	hwlog.RunLog.Debugf("DevicesOfJob: %v", processor.DevicesOfJob)

	hwlog.RunLog.Debugf("result deviceInfos %v", deviceContent.AllConfigmap)
	return deviceContent
}

func (processor *recoverInplaceFaultProcessor) processFaultInfo(currentTime int64) {
	for nodeName, advanceDeviceInfo := range processor.nodeDeviceCmMap {
		advanceDeviceInfo = processor.processEachNodeFaultInfo(nodeName, advanceDeviceInfo, currentTime)
		processor.nodeDeviceCmMap[nodeName] = advanceDeviceInfo
	}
}

func (processor *recoverInplaceFaultProcessor) processEachNodeFaultInfo(
	nodeName string, deviceInfo *constant.AdvanceDeviceFaultCm, currentTime int64) *constant.AdvanceDeviceFaultCm {
	modified := false
	for jobId, faultJob := range processor.DevicesOfJob {
		for deviceName, device := range faultJob.Node[nodeName].DeviceInfo {
			log := fmt.Sprintf("device: %s on node %s, "+
				"currentTime: %s, ", device.DeviceName, nodeName, util.ReadableMsTime(currentTime))
			if podgroup.JudgeRestartProcessByJobKey(jobId) {
				detailInfo := device.FaultDetail
				fullLog := log + fmt.Sprintf("faultTime: %s", util.ReadableMsTime(detailInfo.FaultTime))
				if processor.canFilterNormalDeviceFaultInfo(jobId, device, currentTime) {
					hwlog.RunLog.Warn("Processor filter normal " + fullLog)
					processor.filterNormalDeviceFaultInfo(deviceName, deviceInfo)
					modified = true
				} else {
					hwlog.RunLog.Warn("Processor cannot filter normal " + fullLog)
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

func (processor *recoverInplaceFaultProcessor) filterNormalDeviceFaultInfo(
	deviceName string, advanceDevInfo *constant.AdvanceDeviceFaultCm) {
	for _, fault := range advanceDevInfo.FaultDeviceList[deviceName] {
		if faultdomain.IsL2L3Fault(fault.FaultLevel) {
			advanceDevInfo.DelFaultAndFix(fault)
		}
	}
}

func (processor *recoverInplaceFaultProcessor) canFilterNormalDeviceFaultInfo(jobId string,
	device constant.SingleProcessDeviceInfo,
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
	detailInfo := device.FaultDetail
	return detailInfo.FaultTime >= currentTime-constant.JobRestartInPlaceTimeout
}

func (processor *recoverInplaceFaultProcessor) handleDeviceOfNodes() map[string]constant.SingleProcessNodeInfo {
	faultNodes := make(map[string]constant.SingleProcessNodeInfo)
	for nodeName, deviceInfo := range processor.nodeDeviceCmMap {
		faultDevicesOnNode := processor.getFaultDevices(nodeName, deviceInfo)

		if len(faultDevicesOnNode.DeviceInfo) == 0 {
			continue
		}
		faultNodes[nodeName] = faultDevicesOnNode
	}
	return faultNodes
}

func (processor *recoverInplaceFaultProcessor) getDevicesForTolerateJobs() map[string]constant.SingleProcessJobInfo {
	nodeNameList := make([]string, 0)
	for key, _ := range processor.nodeDeviceCmMap {
		nodeNameList = append(nodeNameList, key)
	}
	faultJobs := make(map[string]constant.SingleProcessJobInfo)
	for jobUid, serverList := range processor.jobServerInfoMap.InfoMap {
		if !podgroup.JudgeRestartProcessByJobKey(jobUid) {
			continue
		}
		jobInfo := constant.SingleProcessJobInfo{
			Node:  make(map[string]constant.SingleProcessNodeInfo),
			JobId: jobUid,
		}
		for _, nodeName := range nodeNameList {
			devicesOfJobOnNode := serverList[nodeName]
			if len(devicesOfJobOnNode.DeviceList) == 0 {
				continue
			}
			jobInfo.Node[nodeName] = processor.initDeviceFromNodeAndReportInfo(jobUid, nodeName)
		}
		if len(jobInfo.Node) != 0 {
			faultJobs[jobUid] = jobInfo
		}
	}
	return faultJobs
}

func (processor *recoverInplaceFaultProcessor) getFaultDevices(
	nodeName string, deviceInfo *constant.AdvanceDeviceFaultCm) constant.SingleProcessNodeInfo {
	nodeInfo := constant.SingleProcessNodeInfo{
		NodeName:   nodeName,
		DeviceInfo: make(map[string]constant.SingleProcessDeviceInfo),
	}
	for _, deviceFaults := range deviceInfo.FaultDeviceList {
		for _, fault := range deviceFaults {
			if faultdomain.IsL1Fault(fault.FaultLevel) {
				continue
			}
			errorMsg := fmt.Sprintf("getFaultDevices cannot find fault time for device %s of node %s",
				deviceInfo.CmName, nodeName)
			faultTime := faultdomain.GetFaultTime(fault, errorMsg)
			faultDeviceInfo, ok := nodeInfo.DeviceInfo[fault.NPUName]
			if !ok {
				faultDeviceInfo = constant.SingleProcessDeviceInfo{
					DeviceName:     fault.NPUName,
					FaultDetail:    constant.DeviceFaultDetail{FaultTime: faultTime},
					FaultCodeLevel: make(map[string]string),
				}
			}
			detailInfo := constant.DeviceFaultDetail{
				FaultTime: faultTime,
				HasFaultAboveL3: !faultdomain.IsL2L3Fault(fault.FaultLevel) &&
					!faultdomain.IsL1Fault(fault.FaultLevel),
			}
			oldDetailInfo := faultDeviceInfo.FaultDetail
			detailInfo.FaultTime = util.MinInt(oldDetailInfo.FaultTime, detailInfo.FaultTime)
			detailInfo.HasFaultAboveL3 = detailInfo.HasFaultAboveL3 || oldDetailInfo.HasFaultAboveL3
			faultDeviceInfo.FaultDetail = detailInfo
			faultDeviceInfo.FaultCodeLevel[fault.FaultCode] = fault.FaultLevel
			nodeInfo.DeviceInfo[fault.NPUName] = faultDeviceInfo
		}
	}
	return nodeInfo
}

// CanDoRestartInPlace judge job can restart fault process in place
func (processor *recoverInplaceFaultProcessor) CanDoRestartInPlace(jobId string) bool {
	jobFaultDetail, ok := processor.normalFaultDetailOfJob[jobId]
	if !ok {
		return false
	}
	// Is it necessary to report the fault to volcano
	if jobFaultDetail.HasFaultAboveL3 || jobFaultDetail.HasRank0Fault ||
		(jobFaultDetail.ReportTime != constant.JobShouldReportFault &&
			jobFaultDetail.ReportTime > jobFaultDetail.FaultTime) {
		return false
	}
	return jobFaultDetail.FaultTime >= time.Now().UnixMilli()-constant.JobRestartInPlaceTimeout
}

// GetFilterFaultCodeAndLevel get filtered fault info
func (processor *recoverInplaceFaultProcessor) GetFilterFaultCodeAndLevel(jobId, nodeName, deviceName string) map[string]string {
	jobInfo, found := processor.DevicesOfJob[jobId]
	if !found {
		return nil
	}
	device, found := jobInfo.Node[nodeName].DeviceInfo[deviceName]
	if !found {
		hwlog.RunLog.Debugf("job %s's fault is not on node %s device %s", jobId, nodeName, deviceName)
		return nil
	}
	return device.FaultCodeLevel
}

// JobHasFault judge job has fault
func (processor *recoverInplaceFaultProcessor) JobHasFault(jobId string) bool {
	filterJob, found := processor.DevicesOfJob[jobId]
	if !found {
		return false
	}
	subHealthStrategy := podgroup.GetSubHealthStrategyByJobKey(jobId)
	for _, filterNode := range filterJob.Node {
		for _, filterDevice := range filterNode.DeviceInfo {
			if faultdomain.ContainCannotIgnoreFault(filterDevice.FaultCodeLevel, subHealthStrategy) {
				return true
			}
		}
	}
	return false
}
