// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

type faultProcessorImpl struct {
	*jobRankFaultInfoProcessor
	center *FaultProcessCenter
}

func (fpi *faultProcessorImpl) process() {
	deviceInfos := fpi.center.deviceCenter.getProcessingCm()
	hwlog.RunLog.Debugf("deviceInfos: %#v", deviceInfos)
	deviceCmForNodeMap := getAdvanceDeviceCmForNodeMap(deviceInfos)
	nodeInfos := fpi.center.nodeCenter.getProcessingCm()
	hwlog.RunLog.Debugf("nodeInfos: %#v", nodeInfos)
	switchInfos := fpi.center.switchCenter.getProcessingCm()
	hwlog.RunLog.Debugf("switchInfos: %#v", switchInfos)

	jobFaultInfos := make(map[string]JobFaultInfo)
	jobServerInfoMap := fpi.center.jobServerInfoMap
	for jobId, serverList := range jobServerInfoMap.InfoMap {
		jobFaultInfo := JobFaultInfo{
			JobId:     jobId,
			FaultList: make([]FaultRank, 0),
		}
		hwlog.RunLog.Debugf("serverList: %d", len(serverList))
		for nodeName, server := range serverList {
			hwlog.RunLog.Debugf("nodeName: %s, server: %#v", nodeName, server)
			ni, ok := nodeInfos[constant.NodeInfoPrefix+nodeName]
			if ok && (ni.NodeStatus == "UnHealthy" || time.Since(ni.UpdateTime) > time.Second*time.Duration(ni.HeartbeatInterval*3)) {
				hwlog.RunLog.Infof("node %s is unhealthy or update time is too long(%f)", nodeName,
					time.Since(ni.UpdateTime).Seconds())
				jobFaultInfo.FaultList = append(jobFaultInfo.FaultList, serverHcclToFaultRank(server)...)
				continue
			}
			si, ok := switchInfos[constant.SwitchInfoPrefix+nodeName]
			if ok && si.NodeStatus == "UnHealthy" {
				hwlog.RunLog.Infof("node %s switch is unhealthy", nodeName)
				jobFaultInfo.FaultList = append(jobFaultInfo.FaultList, serverHcclToFaultRank(server)...)
				continue
			}
			faultRankList := fpi.findFaultRankForJob(deviceCmForNodeMap, nodeName, serverList, jobId)
			jobFaultInfo.FaultList = append(jobFaultInfo.FaultList, faultRankList...)

		}
		if len(jobFaultInfo.FaultList) > 0 {
			hwlog.RunLog.Infof("jobFaultInfo: %#v", jobFaultInfo)
		}
		jobFaultInfos[jobId] = jobFaultInfo
	}
	fpi.jobRankFaultInfoProcessor.jobFaultInfoMap = jobFaultInfos
}

func serverHcclToFaultRank(server constant.ServerHccl) []FaultRank {
	faultRanks := make([]FaultRank, 0, len(server.DeviceList))
	for _, device := range server.DeviceList {
		faultRanks = append(faultRanks, FaultRank{
			RankId:      device.RankID,
			FaultCode:   "",
			FaultLevel:  SeparateNPU,
			DoStepRetry: false,
		})
	}
	return faultRanks
}
