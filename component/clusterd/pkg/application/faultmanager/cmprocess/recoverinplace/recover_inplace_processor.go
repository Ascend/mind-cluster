// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package recoverinplace contain filtering fault handling method for single process fault
package recoverinplace

import (
	"fmt"
	"strconv"
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
	"k8s.io/apimachinery/pkg/util/sets"
)

var RecoverInplaceProcessor *recoverInplaceFaultProcessor

type recoverInplaceFaultProcessor struct {
	DevicesOfJob     map[string]*constant.SingleProcessJobInfo    // job -> node -> device -> detail
	faultDetailOfJob map[string]*constant.SingleProcessJobPodInfo // job-> podRank -> detail
	DeviceOfNode     map[string]*constant.SingleProcessNodeInfo   // node -> device -> detail
	jobServerInfoMap constant.JobServerInfoMap
	nodeDeviceCmMap  map[string]*constant.AdvanceDeviceFaultCm
}

func init() {
	RecoverInplaceProcessor = &recoverInplaceFaultProcessor{
		nodeDeviceCmMap: make(map[string]*constant.AdvanceDeviceFaultCm),
	}
}

func (processor *recoverInplaceFaultProcessor) initDeviceFromNodeAndReportInfo(jobId,
	nodeName string) *constant.SingleProcessNodeInfo {
	managerPlaneFaultNode := processor.DeviceOfNode[nodeName]
	if managerPlaneFaultNode == nil {
		return nil
	}
	devicesOfJobOnNode := processor.jobServerInfoMap.InfoMap[jobId][nodeName]
	deviceNumOfPod := pod.GetPodDeviceNumByJobId(jobId)
	jobSingleProcessNodeInfo := &constant.SingleProcessNodeInfo{NodeName: nodeName,
		DeviceInfo: make(map[string]*constant.SingleProcessDeviceInfo)}
	for _, deviceOfJob := range devicesOfJobOnNode.DeviceList {
		deviceName := processor.nodeDeviceCmMap[nodeName].DeviceType + "-" + deviceOfJob.DeviceID
		if faultDevice, ok := managerPlaneFaultNode.DeviceInfo[deviceName]; ok {
			podRank, err := common.CalculateStringDivInt(deviceOfJob.RankID, deviceNumOfPod)
			if err != nil {
				hwlog.RunLog.Errorf("job %v calculate pod rank error, %v", jobId, err)
				podRank = constant.InvalidPodRank
			}
			podRankStr := strconv.Itoa(podRank)
			reportTime := collector.ReportInfoCollector.GetSingleProcessFaultReportTime(jobId, podRankStr)
			faultDevice.FaultDetail.PodRankStr = podRankStr
			faultDevice.FaultDetail.ReportTime = reportTime
			jobSingleProcessNodeInfo.DeviceInfo[deviceName] = faultDevice
			processor.updateNormalFaultDetailOfJob(jobId, podRankStr, faultDevice.FaultDetail)
		}
	}
	return jobSingleProcessNodeInfo
}

func (processor *recoverInplaceFaultProcessor) updateNormalFaultDetailOfJob(jobId, podRankStr string,
	detail *constant.DeviceFaultDetail) {
	podDetail, ok := processor.faultDetailOfJob[jobId]
	if !ok {
		podDetail = &constant.SingleProcessJobPodInfo{
			JobId: jobId,
			Pod:   make(map[string]*constant.DeviceFaultDetail),
		}
		processor.faultDetailOfJob[jobId] = podDetail
	}
	podRankFaultDetail, ok := podDetail.Pod[podRankStr]
	if !ok {
		podRankFaultDetail = &constant.DeviceFaultDetail{
			FaultTime:      detail.FaultTime,
			ReportTime:     detail.ReportTime,
			FaultCodeLevel: make(map[string]string),
			PodRankStr:     podRankStr,
		}
		util.MergeStringMapList[string](podRankFaultDetail.FaultCodeLevel, detail.FaultCodeLevel)
		podDetail.Pod[podRankStr] = podRankFaultDetail
		return
	}
	podRankFaultDetail.FaultTime = util.MinInt(podRankFaultDetail.FaultTime, detail.FaultTime)
	podRankFaultDetail.ReportTime = util.MinInt(podRankFaultDetail.ReportTime, detail.ReportTime)
	util.MergeStringMapList[string](podRankFaultDetail.FaultCodeLevel, detail.FaultCodeLevel)
	podRankFaultDetail.PodRankStr = podRankStr
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

	processor.faultDetailOfJob = make(map[string]*constant.SingleProcessJobPodInfo)
	processor.DevicesOfJob = processor.getDevicesForTolerateJobs()
	hwlog.RunLog.Debugf("current DevicesOfJob %v", processor.DevicesOfJob)

	processor.processFaultInfo(currentTime)
	hwlog.RunLog.Debugf("faultDetailOfJob: %v", processor.faultDetailOfJob)
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
		if podgroup.JudgeRestartProcessByJobKey(jobId) {
			faultNode, ok := faultJob.Node[nodeName]
			if !ok || faultNode == nil {
				continue
			}
			subHealthStrategy := podgroup.GetSubHealthStrategyByJobKey(jobId)
			for deviceName, device := range faultNode.DeviceInfo {
				log := fmt.Sprintf("device: %s on node %s, "+
					"currentTime: %s, ", device.DeviceName, nodeName, util.ReadableMsTime(currentTime))
				detailInfo := device.FaultDetail
				fullLog := log + fmt.Sprintf("faultTime: %s", util.ReadableMsTime(detailInfo.FaultTime))
				canFilter, reason := processor.canFilterNormalDeviceFaultInfo(jobId, detailInfo.PodRankStr,
					currentTime, subHealthStrategy)
				if canFilter {
					hwlog.RunLog.Warn("Processor filter normal " + fullLog)
					processor.filterNormalDeviceFaultInfo(deviceName, deviceInfo)
					modified = true
				} else {
					hwlog.RunLog.Warn("Processor cannot filter normal " + fullLog + "," + reason)
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

func (processor *recoverInplaceFaultProcessor) canFilterNormalDeviceFaultInfo(jobId, podRankStr string,
	currentTime int64, subHealthStrategy string) (bool, string) {
	jobFaultDetail, ok := processor.faultDetailOfJob[jobId]
	if !ok {
		return false, constant.FailedReasonJobNoFault
	}
	if _, exist := jobFaultDetail.Pod[constant.InvalidPodRankStr]; exist {
		return false, constant.FailedReasonParseRankError
	}
	if podFaultDetail, exist := jobFaultDetail.Pod[constant.PodRankStrZero]; exist {
		canFilter, reason := faultDetailCanDoRestartInPlace(podFaultDetail, currentTime, subHealthStrategy)
		if !canFilter {
			return false, fmt.Sprintf("podRank0 %v", reason)
		}
	}
	if podFaultDetail, exist := jobFaultDetail.Pod[podRankStr]; exist {
		return faultDetailCanDoRestartInPlace(podFaultDetail, currentTime, subHealthStrategy)
	}
	return false, constant.FailedReasonPodRankNoFault
}

func (processor *recoverInplaceFaultProcessor) handleDeviceOfNodes() map[string]*constant.SingleProcessNodeInfo {
	faultNodes := make(map[string]*constant.SingleProcessNodeInfo)
	for nodeName, deviceInfo := range processor.nodeDeviceCmMap {
		faultDevicesOnNode := processor.getFaultDevices(nodeName, deviceInfo)

		if len(faultDevicesOnNode.DeviceInfo) == 0 {
			continue
		}
		faultNodes[nodeName] = faultDevicesOnNode
	}
	return faultNodes
}

func (processor *recoverInplaceFaultProcessor) getDevicesForTolerateJobs() map[string]*constant.SingleProcessJobInfo {
	nodeNameList := make([]string, 0)
	for key, _ := range processor.nodeDeviceCmMap {
		nodeNameList = append(nodeNameList, key)
	}
	faultJobs := make(map[string]*constant.SingleProcessJobInfo)
	for jobUid, serverList := range processor.jobServerInfoMap.InfoMap {
		if !podgroup.JudgeRestartProcessByJobKey(jobUid) {
			continue
		}
		jobInfo := &constant.SingleProcessJobInfo{
			Node:  make(map[string]*constant.SingleProcessNodeInfo),
			JobId: jobUid,
		}
		for _, nodeName := range nodeNameList {
			devicesOfJobOnNode := serverList[nodeName]
			if len(devicesOfJobOnNode.DeviceList) == 0 {
				continue
			}
			if nodeInfo := processor.initDeviceFromNodeAndReportInfo(jobUid, nodeName); nodeInfo != nil &&
				len(nodeInfo.DeviceInfo) > 0 {
				jobInfo.Node[nodeName] = nodeInfo
			}
		}
		if len(jobInfo.Node) != 0 {
			faultJobs[jobUid] = jobInfo
		}
	}
	return faultJobs
}

func (processor *recoverInplaceFaultProcessor) getFaultDevices(
	nodeName string, deviceInfo *constant.AdvanceDeviceFaultCm) *constant.SingleProcessNodeInfo {
	nodeInfo := &constant.SingleProcessNodeInfo{
		NodeName:   nodeName,
		DeviceInfo: make(map[string]*constant.SingleProcessDeviceInfo),
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
				faultDeviceInfo = &constant.SingleProcessDeviceInfo{
					DeviceName: fault.NPUName,
					FaultDetail: &constant.DeviceFaultDetail{FaultTime: faultTime,
						FaultCodeLevel: make(map[string]string)},
				}
			}
			faultDeviceInfo.FaultDetail.FaultTime = util.MinInt(faultDeviceInfo.FaultDetail.FaultTime, faultTime)
			faultDeviceInfo.FaultDetail.FaultCodeLevel[fault.FaultCode] = fault.FaultLevel
			nodeInfo.DeviceInfo[fault.NPUName] = faultDeviceInfo
		}
	}
	return nodeInfo
}

// CanDoRestartInPlace judge job can restart fault process in place
func (processor *recoverInplaceFaultProcessor) CanDoRestartInPlace(jobId, podRankStr string) bool {
	subHealthStrategy := podgroup.GetSubHealthStrategyByJobKey(jobId)
	canFilter, _ := processor.canFilterNormalDeviceFaultInfo(jobId, podRankStr, time.Now().UnixMilli(), subHealthStrategy)
	return canFilter
}

func faultDetailCanDoRestartInPlace(faultDetail *constant.DeviceFaultDetail, currentTime int64,
	subHealthStrategy string) (bool, string) {
	// Is it necessary to report the fault to volcano
	faultLevels := sets.NewString(util.GetStringMapValueList[string](faultDetail.FaultCodeLevel)...)
	if !faultdomain.IsRecoverInPlaceFaultLevels(faultLevels, subHealthStrategy) {
		return false, constant.FailedReasonHasOtherFault
	}
	if faultDetail.ReportTime != constant.JobShouldReportFault && faultDetail.ReportTime > faultDetail.FaultTime {
		return false, constant.FailedReasonShouldReport
	}
	if faultDetail.FaultTime < currentTime-constant.JobRestartInPlaceTimeout {
		return false, constant.FailedReasonFaultTimeOut
	}
	return true, ""
}

// GetFilterFaultCodeAndLevel get filtered fault info
func (processor *recoverInplaceFaultProcessor) GetFilterFaultCodeAndLevel(jobId, nodeName, deviceName string) map[string]string {
	jobInfo, found := processor.DevicesOfJob[jobId]
	if !found || jobInfo == nil {
		return nil
	}
	nodeInfo, found := jobInfo.Node[nodeName]
	if !found || nodeInfo == nil {
		return nil
	}
	device, found := nodeInfo.DeviceInfo[deviceName]
	if !found || device == nil || device.FaultDetail == nil {
		hwlog.RunLog.Debugf("job %s's fault is not on node %s device %s", jobId, nodeName, deviceName)
		return nil
	}
	return device.FaultDetail.FaultCodeLevel
}

// GetJobUnRecoverdPodRanks get job unrecoverd pod ranks
func (processor *recoverInplaceFaultProcessor) GetJobUnRecoverdPodRanks(jobId string, subHealthStrategy string) ([]string, bool) {
	jobFault, found := processor.faultDetailOfJob[jobId]
	if !found {
		return nil, true
	}
	unRecoverdPod := make([]string, 0)
	done := true
	for podRank, podRankFault := range jobFault.Pod {
		faultLevels := sets.NewString(util.GetStringMapValueList[string](podRankFault.FaultCodeLevel)...)
		if faultdomain.IsRecoverInPlaceFaultLevels(faultLevels, subHealthStrategy) {
			hwlog.RunLog.Infof("jobId %s have recoverable podRank fault, podRank %v, podRankFault %v", jobId, podRank, podRankFault)
			done = false
		}
		if faultdomain.FaultLevelsHasNpuFault(faultLevels, subHealthStrategy) {
			hwlog.RunLog.Infof("jobId %s have fault, podRank %v, podRankFault %v", jobId, podRank, podRankFault)
			unRecoverdPod = append(unRecoverdPod, podRank)
		}
	}
	return unRecoverdPod, done
}
