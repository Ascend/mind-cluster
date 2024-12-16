package faultmanager

import (
	"context"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"

	"clusterd/pkg/application/resource"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/kube"
)

// Process faultmanager process start
func Process(ctx context.Context) {
	fJobCenter := newFaultJobProcessCenter()
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel closed")
			}
			hwlog.RunLog.Info("stop fault process")
			return
		default:
			fJobCenter.process()
			time.Sleep(5 * time.Second)
		}
	}
}

func newFaultJobProcessCenter() *faultJobProcessCenter {
	return &faultJobProcessCenter{}
}

func (fJobCenter *faultJobProcessCenter) process() {

	fJobCenter.jobServerInfoMap = kube.JobMgr.GetJobServerInfoMap()
	fJobCenter.switchInfoCm = resource.GetSwitchInfoMap()
	fJobCenter.deviceInfoCm = resource.GetDeviceInfoMap()
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
