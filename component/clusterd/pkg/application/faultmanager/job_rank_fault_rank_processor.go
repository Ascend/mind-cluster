// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"sync"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/application/job"
	"clusterd/pkg/common/util"
)

func newJobRankFaultInfoProcessor(deviceCenter *deviceFaultProcessCenter) *jobRankFaultInfoProcessor {
	return &jobRankFaultInfoProcessor{
		jobFaultInfoMap: make(map[string]JobFaultInfo),
		deviceCenter:    deviceCenter,
		mutex:           sync.RWMutex{},
	}
}

func (processor *jobRankFaultInfoProcessor) getJobFaultRankInfos() map[string]JobFaultInfo {
	result := new(map[string]JobFaultInfo)
	err := util.DeepCopy(result, processor.jobFaultInfoMap)
	if err != nil {
		hwlog.RunLog.Errorf("get job fault rank failed, err: %v", err)
		return nil
	}
	return *result
}

func (processor *jobRankFaultInfoProcessor) getJobFaultRankInfosFilterLevel(
	faultLevel string) map[string]JobFaultInfo {
	jobFaultRankInfos := processor.getJobFaultRankInfos()
	if jobFaultRankInfos == nil {
		return nil
	}
	for jobId, jobFaultInfo := range jobFaultRankInfos {
		faultList := make([]FaultRank, 0)
		for _, fault := range jobFaultInfo.FaultList {
			if fault.FaultLevel != faultLevel {
				faultList = append(faultList, fault)
			}
		}
		jobFaultInfo.FaultList = faultList
		jobFaultRankInfos[jobId] = jobFaultInfo
	}
	return jobFaultRankInfos
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
	nodeDeviceInfoMap map[string]AdvanceDeviceFaultCm, nodeName string, serverList map[string]job.ServerHccl) []FaultRank {
	advanceDeviceInfo := nodeDeviceInfoMap[nodeName]
	devicesOfJobOnNode, ok := serverList[nodeName]
	faultRankList := make([]FaultRank, 0)
	if !ok || len(devicesOfJobOnNode.DeviceList) == 0 {
		return faultRankList
	}
	for _, deviceInfo := range devicesOfJobOnNode.DeviceList {
		deviceName := advanceDeviceInfo.ServerType + "-" + deviceInfo.DeviceID
		faultList, ok := advanceDeviceInfo.FaultDeviceList[deviceName]
		if !ok {
			continue
		}
		for _, fault := range faultList {
			faultRankList = append(faultRankList, FaultRank{
				RankId:     deviceInfo.RankID,
				FaultCode:  fault.FaultCode,
				FaultLevel: fault.FaultLevel,
			})
		}
	}
	return faultRankList
}
