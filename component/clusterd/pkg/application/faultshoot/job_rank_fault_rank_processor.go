// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultshoot contain fault process
package faultshoot

import (
	"sync"

	"clusterd/pkg/application/job"
)

func newJobRankFaultInfoProcessor(deviceCenter *deviceFaultProcessCenter) *jobRankFaultInfoProcessor {
	return &jobRankFaultInfoProcessor{
		jobFaultInfoMap: make(map[string]JobFaultInfo),
		deviceCenter:    deviceCenter,
		mutex:           sync.RWMutex{},
	}
}

func (processor *jobRankFaultInfoProcessor) getJobFaultRankInfos() map[string]JobFaultInfo {
	return processor.jobFaultInfoMap
}

func (processor *jobRankFaultInfoProcessor) setJobFaultRankInfos(faultInfos map[string]JobFaultInfo) {
	processor.jobFaultInfoMap = faultInfos
}

func (processor *jobRankFaultInfoProcessor) process() {
	deviceInfos := processor.deviceCenter.getProcessingCm()
	nodesName := getNodesNameFromDeviceInfo(deviceInfos)
	deviceCmForNodeMap := getAdvanceDeviceCmForNodeMap(deviceInfos)

	jobFaultInfos := make(map[string]JobFaultInfo)
	jobServerInfoMap := processor.deviceCenter.jobServerInfoMap
	for jobId, serverList := range jobServerInfoMap.InfoMap {
		jobFaultInfo := JobFaultInfo{
			JobId:     jobId,
			FaultList: make([]FaultRank, 0),
		}

		for _, nodeName := range nodesName {
			faultRankList := processor.findFaultRankForJob(deviceCmForNodeMap, nodeName, serverList)
			jobFaultInfo.FaultList = append(jobFaultInfo.FaultList, faultRankList...)
		}
		jobFaultInfos[jobId] = jobFaultInfo
	}
	processor.setJobFaultRankInfos(jobFaultInfos)
}

func (processor *jobRankFaultInfoProcessor) findFaultRankForJob(
	nodeDeviceInfoMap map[string]AdvanceDeviceCm, nodeName string, serverList map[string]job.ServerHccl) []FaultRank {
	advanceDeviceInfo := nodeDeviceInfoMap[nodeName]
	devicesOfJobOnNode, ok := serverList[nodeName]
	faultRankList := make([]FaultRank, 0)
	if !ok || len(devicesOfJobOnNode.DeviceList) == 0 {
		return faultRankList
	}
	for _, deviceInfo := range devicesOfJobOnNode.DeviceList {
		deviceName := advanceDeviceInfo.ServerType + "-" + deviceInfo.DeviceID
		faultList, ok := advanceDeviceInfo.DeviceList[deviceName]
		if !ok {
			continue
		}
		for _, fault := range faultList {
			faultRankList = append(faultRankList, FaultRank{
				RankId:    deviceInfo.RankID,
				FaultCode: fault.FaultCode,
			})
		}
	}
	return faultRankList
}
