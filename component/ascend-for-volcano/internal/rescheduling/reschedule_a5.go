/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package rescheduling

import (
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

func (reScheduler *ReScheduler) updateFaultJobWhenGraceDeleteSuccess(jobInfo *api.JobInfo, faultJob *FaultJob,
	is910A5 bool) {
	for id := range faultJob.FaultTasks {
		faultJob.FaultTasks[id].IsBeingGracefulDeleted = false
	}
	faultJob.updateFaultJobWhenNewPodError(jobInfo)
	klog.V(util.LogDebugLev).Infof("%s grace deleted successful.", faultJob.JobName)
	klog.V(util.LogDebugLev).Infof(
		"rescheduling: before rescheduling upgrade, job <%s/%s> PendingSessionNum=%d,"+
			"DeleteExecutedFlag=%t", faultJob.JobNamespace, faultJob.JobName,
		faultJob.PendingSessionNum, faultJob.DeleteExecutedFlag)
	// delete cache when all pods have been allocated
	if is910A5 {
		reScheduler.singlePodReschedulingUpgradeFor910A5(jobInfo, faultJob)
	} else {
		reScheduler.singlePodReschedulingUpgrade(faultJob)
	}
	klog.V(util.LogDebugLev).Infof(
		"rescheduling: after rescheduling upgrade, job <%s/%s> PendingSessionNum=%d,"+
			"DeleteExecutedFlag=%t", faultJob.JobNamespace, faultJob.JobName,
		faultJob.PendingSessionNum, faultJob.DeleteExecutedFlag)
}

func (reScheduler *ReScheduler) singlePodReschedulingUpgradeFor910A5(jobInfo *api.JobInfo, fJob *FaultJob) {
	if jobInfo.PodGroup.Labels[util.SinglePodTag] != util.EnableFunc {
		return
	}

	if reScheduler.processPendingRules(fJob) {
		return
	}
	fJob.PendingSessionNum++

	if _, ok := jobInfo.PodGroup.Annotations[util.SuperPodAnnoKey]; ok {
		if fJob.PendingSessionNum == spPendingTimes {
			fJob.DeleteExecutedFlag = false
		}
	}

	if fJob.PendingSessionNum == pendingTimes {
		fJob.DeleteExecutedFlag = false
	}

	klog.V(util.LogDebugLev).Infof("rescheduling: WhetherBackToVspSchedule=%t",
		reScheduler.Jobs[fJob.JobUID].WhetherBackToVspSchedule)
	if reScheduler.Jobs[fJob.JobUID].WhetherBackToVspSchedule &&
		fJob.PendingSessionNum == backToVspPendingTimes {
		fJob.DeleteExecutedFlag = false
	}
}

func (reScheduler *ReScheduler) processPendingRules(fJob *FaultJob) bool {
	if sJob := reScheduler.Jobs[fJob.JobUID]; fJob.IsProcessReschedulingJob(&sJob) {
		if fJob.PendingSessionNum < pendingTimes {
			fJob.PendingSessionNum++
		}
		return true
	}
	if sJob := reScheduler.Jobs[fJob.JobUID]; fJob.IsProcessReschedulePause(&sJob) {
		if fJob.PendingSessionNum < pendingTimes {
			fJob.DeleteExecutedFlag = false
			fJob.PendingSessionNum++
			klog.V(util.LogInfoLev).Infof("process recover failed, back to job rescheduling")
		}
		return true
	}
	return false
}
