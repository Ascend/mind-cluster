// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/kube"
)

type faultProcessorImpl struct {
	*jobRankFaultInfoProcessor
}

func (fpi *faultProcessorImpl) process() {
	deviceInfos := GlobalFaultProcessCenter.deviceCenter.getProcessingCm()
	hwlog.RunLog.Debugf("deviceInfos: %#v", deviceInfos)
	deviceCmForNodeMap := getAdvanceDeviceCmForNodeMap(deviceInfos)
	nodeInfos := GlobalFaultProcessCenter.nodeCenter.getProcessingCm()
	hwlog.RunLog.Debugf("nodeInfos: %#v", nodeInfos)
	switchInfos := GlobalFaultProcessCenter.switchCenter.getProcessingCm()
	hwlog.RunLog.Debugf("switchInfos: %#v", switchInfos)

	jobFaultInfos := make(map[string]JobFaultInfo)
	jobServerInfoMap := GlobalFaultProcessCenter.jobServerInfoMap
	for jobId, serverList := range jobServerInfoMap.InfoMap {
		jobFaultInfo := JobFaultInfo{
			JobId:     jobId,
			FaultList: make([]FaultRank, 0),
		}
		hwlog.RunLog.Debugf("serverList: %d", len(serverList))
		for nodeName, server := range serverList {
			hwlog.RunLog.Debugf("nodeName: %s, server: %#v", nodeName, server)
			ni, ok := nodeInfos[constant.NodeInfoPrefix+nodeName]
			if ok && ni.NodeStatus == "UnHealthy" {
				hwlog.RunLog.Infof("node %s is unhealthy", nodeName)
				jobFaultInfo.FaultList = append(jobFaultInfo.FaultList, serverHcclToFaultRank(server)...)
				continue
			}
			node := kube.GetNodeFromIndexer(nodeName)
			if node == nil || !isNodeReady(node) {
				hwlog.RunLog.Infof("node %s is not ready", nodeName)
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
	fpi.jobRankFaultInfoProcessor.setJobFaultRankInfos(jobFaultInfos)
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
