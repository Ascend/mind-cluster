// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
)

func newFaultJobProcessCenter() *faultJobProcessCenter {
	return &faultJobProcessCenter{}
}

func (fJobCenter *faultJobProcessCenter) process() {
	currentTime := time.Now().UnixMilli()
	if fJobCenter.isProcessLimited(currentTime) {
		return
	}
	fJobCenter.lastProcessTime = currentTime

	fJobCenter.jobServerInfoMap = job.GetJobServerInfoMap()
	fJobCenter.nodeInfoCm = GlobalFaultProcessCenter.nodeCenter.baseFaultCenter.getProcessedCm()
	fJobCenter.switchInfoCm = GlobalFaultProcessCenter.switchCenter.baseFaultCenter.getProcessedCm()
	fJobCenter.deviceInfoCm = GlobalFaultProcessCenter.deviceCenter.baseFaultCenter.getProcessedCm()
	fJobCenter.InitFaultJobs()
	for _, fJob := range fJobCenter.FaultJobs {
		fJob.process()
	}

}

func (fJobCenter *faultJobProcessCenter) isProcessLimited(currentTime int64) bool {
	return fJobCenter.lastProcessTime+faultJobProcessInterval > currentTime
}

func (fJobCenter *faultJobProcessCenter) InitFaultJobs() {
	deviceCmForNodeMap := getAdvanceDeviceCmForNodeMap(fJobCenter.deviceInfoCm)
	faultJobs := make(map[string]*FaultJob)
	for jobId, serverLists := range fJobCenter.jobServerInfoMap.InfoMap {
		tmpFaultJob, ok := fJobCenter.FaultJobs[jobId]
		if !ok {
			tmpFaultJob = &FaultJob{}
		}
		tmpFaultJob.TriggerFault = nil
		tmpFaultJob.AllFaultCode = sets.String{}
		tmpFaultJob.initFaultJobAttr()
		for nodeName, serverList := range serverLists {
			tmpFaultJob.IsA3Job = deviceCmForNodeMap[nodeName].SuperPodID >= 0
			tmpFaultJob.PodNames[serverList.ServerName] = serverList.PodID
			tmpFaultJob.NameSpace = serverList.PodNameSpace
			tmpFaultJob.initFaultJobBySwitchFault(fJobCenter.switchInfoCm[constant.SwitchInfoPrefix+nodeName], serverList)
			tmpFaultJob.initFaultJobByDeviceFault(deviceCmForNodeMap[nodeName], serverList)
		}
		faultJobs[jobId] = tmpFaultJob
		hwlog.RunLog.Debugf("init fault job %#v", faultJobs)
	}
	fJobCenter.FaultJobs = faultJobs
}
