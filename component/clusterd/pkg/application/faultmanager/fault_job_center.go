// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/job"
)

func NewFaultJobProcessCenter() *faultJobProcessCenter {
	return &faultJobProcessCenter{}
}

func (fJobCenter *faultJobProcessCenter) Process() {
	currentTime := time.Now().UnixMilli()
	if fJobCenter.isProcessLimited(currentTime) {
		return
	}
	fJobCenter.lastProcessTime = currentTime

	fJobCenter.jobServerInfoMap = job.GetJobServerInfoMap()
	fJobCenter.nodeInfoCm = GlobalFaultProcessCenter.NodeCenter.baseFaultCenter.getProcessedCm()
	fJobCenter.switchInfoCm = GlobalFaultProcessCenter.SwitchCenter.baseFaultCenter.getProcessedCm()
	fJobCenter.deviceInfoCm = GlobalFaultProcessCenter.DeviceCenter.baseFaultCenter.getProcessedCm()
	fJobCenter.InitFaultJobs()
	for _, fJob := range fJobCenter.FaultJobs {
		fJob.Process()
	}

}

func (fJobCenter *faultJobProcessCenter) isProcessLimited(currentTime int64) bool {
	return fJobCenter.lastProcessTime+faultJobProcessInterval > currentTime
}

func (fJobCenter *faultJobProcessCenter) InitFaultJobs() {
	deviceCmForNodeMap := faultdomain.GetAdvanceDeviceCmForNodeMap(fJobCenter.deviceInfoCm)
	faultJobs := make(map[string]*FaultJob)
	for jobId, serverLists := range fJobCenter.jobServerInfoMap.InfoMap {
		if len(serverLists) == 0 {
			hwlog.RunLog.Warnf("job %s serverList is empty", jobId)
			continue
		}
		tmpFaultJob, ok := fJobCenter.FaultJobs[jobId]
		if !ok {
			tmpFaultJob = &FaultJob{}
		}
		tmpFaultJob.initFaultJobAttr()
		for nodeName, serverList := range serverLists {
			tmpFaultJob.IsA3Job = deviceCmForNodeMap[nodeName].SuperPodID >= 0
			tmpFaultJob.PodNames[serverList.ServerName] = serverList.PodID
			tmpFaultJob.NameSpace = serverList.PodNameSpace
			tmpFaultJob.initFaultJobBySwitchFault(fJobCenter.switchInfoCm[constant.SwitchInfoPrefix+nodeName], serverList)
			tmpFaultJob.initFaultJobByDeviceFault(deviceCmForNodeMap[nodeName], serverList)
		}
		faultJobs[jobId] = tmpFaultJob
		hwlog.RunLog.Debugf("init fault job %v", util.ObjToString(faultJobs))
	}
	fJobCenter.FaultJobs = faultJobs
}
