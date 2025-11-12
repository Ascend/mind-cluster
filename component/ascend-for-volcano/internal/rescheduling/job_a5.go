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
	"errors"
	"strconv"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func is910A5Job(schedulerJob *plugin.SchedulerJob) bool {
	if schedulerJob == nil {
		return false
	}
	if v, ok := schedulerJob.Selector[util.AcceleratorType]; ok {
		return util.CheckA5Label(v)
	}
	return false
}

// GraceDeleteJobFor910A5 grace delete jobs labelled to be deleted gracefully
func (fJob *FaultJob) GraceDeleteJobFor910A5(ssn *framework.Session, npuJob *plugin.SchedulerJob,
	env plugin.ScheduleEnv) error {
	if fJob == nil {
		return errors.New("getJobFaultRescheduleLabel fJob object does not exist")
	}
	if ssn == nil {
		return errors.New("session does not exist")
	}
	if npuJob == nil {
		return errors.New("schedulerJob does not exist")
	}
	reason := fJob.getRestartInfosFor910A5()
	isSuperPod := false
	vSuperPodIds := make([]string, 0)
	if _, ok := npuJob.Annotation[util.SuperPodAnnoKey]; ok {
		isSuperPod = true
		vSuperPodIds = fJob.getVSuperPodIds()
	}
	fJob.judgeJobIsMasterFault(vSuperPodIds, npuJob)
	fJob.updateSuperPodsReschdInfo(env)
	dpi := &deletePodInfo{
		isSuperPod: isSuperPod,
		ids:        vSuperPodIds,
		reason:     reason,
	}
	fJob.graceDeletePodsFor910A5(ssn, npuJob, env, dpi)
	return nil
}

// GraceDeleteJob grace delete jobs labelled to be deleted gracefully
func (fJob *FaultJob) getRestartInfosFor910A5() string {
	var reasonList []FaultReasonList
	for _, fTask := range fJob.FaultTasks {
		if fTask.Reason != nil {
			reasonList = append(reasonList, fTask.Reason...)
		}
	}
	reason := GetTaskRestartReason(reasonList)
	return reason
}

func (fJob *FaultJob) graceDeletePodsFor910A5(ssn *framework.Session, npuJob *plugin.SchedulerJob,
	env plugin.ScheduleEnv, dpi *deletePodInfo) {
	klog.V(util.LogInfoLev).Infof("whetherBackToVspSchedule is %v", fJob.WhetherBackToVspSchedule)
	for id, fTask := range fJob.FaultTasks {
		npuTask, ok := npuJob.Tasks[fTask.TaskUID]
		if !ok {
			klog.V(util.LogDebugLev).Infof(
				"rescheduling: skip grace delete because task<%s> for job <%s/%s> has been deleted in session",
				fTask.TaskName, fJob.JobNamespace, fJob.JobName)
			continue
		}
		if fJob.skipThisTask(dpi, fTask, env, npuJob) {
			klog.V(util.LogDebugLev).Infof(
				"rescheduling: skip grace delete task<%s> for job<%s/%s>", fTask.TaskName, fJob.JobNamespace,
				fJob.JobName)
			continue
		}
		klog.V(util.LogDebugLev).Infof(
			"rescheduling: start to grace delete task<%s> for job<%s/%s>", fTask.TaskName, fJob.JobNamespace,
			fJob.JobName)
		fJob.FaultTasks[id].IsSatisfiedRackAffinity = false
		fJob.updateSuperPodMapInfo(env, fTask.TaskName, fTask.NodeName)
		if delErr := npuTask.ForceDeletePodByTaskInf(ssn, dpi.reason, fTask.NodeName); delErr != nil {
			klog.V(util.LogErrorLev).Infof("rescheduling: ForceDeletePodByTaskInf %s: %s.", npuTask.Name, delErr)
			continue
		}
		klog.V(util.LogDebugLev).Infof(
			"rescheduling: grace delete task<%s> for job<%s/%s> succeeded", fTask.TaskName, fJob.JobNamespace,
			fJob.JobName)
		fJob.FaultTasks[id].IsBeingGracefulDeleted = true
	}
}

// ForceDeleteJobFor910A5 force delete jobs includes labelled force delete ones and grace delete failed ones
func (fJob *FaultJob) ForceDeleteJobFor910A5(schedulerJob *plugin.SchedulerJob,
	env plugin.ScheduleEnv) error {
	klog.V(util.LogDebugLev).Infof("enter ForceDeleteJob")
	if fJob == nil || schedulerJob == nil {
		return errors.New("getJobFaultRescheduleLabel fJob object or ssn or schedulerJob does not exist")
	}
	isSuperPod := false
	vSuperPodIds := make([]string, 0)
	if _, ok := schedulerJob.Annotation[util.SuperPodAnnoKey]; ok {
		isSuperPod = true
		vSuperPodIds = fJob.getVSuperPodIds()
	}
	fJob.updateSuperPodsReschdInfo(env)
	fJob.judgeJobIsMasterFault(vSuperPodIds, schedulerJob)
	dpi := &deletePodInfo{
		isSuperPod: isSuperPod,
		ids:        vSuperPodIds,
	}
	fJob.forceDeletePodsFor910A5(schedulerJob, env, dpi)
	return nil
}

func (fJob *FaultJob) judgeJobIsMasterFault(vSuperPodIds []string, schedulerJob *plugin.SchedulerJob) {
	for _, fTask := range fJob.FaultTasks {
		if fTask.NodeRankIndex != util.Rank0 {
			continue
		}
		if fTask.IsFaultTask {
			fJob.IsMasterFault = true
			return
		}
		if fJob.IsProcessReschedulingJob(schedulerJob) && fJob.inTheSameTpBlock(fTask) {
			klog.V(util.LogInfoLev).Infof("master fault in process reschedule, start job reschedule")
			fJob.IsMasterFault = true
			return
		}
		if fJob.PendingSessionNum >= spPendingTimes && fJob.inTheSameTpBlock(fTask) {
			klog.V(util.LogInfoLev).Infof("master pod and fault task is in the same tpBlock")
			fJob.IsMasterFault = true
			return
		}
		if fJob.PendingSessionNum >= spPendingTimes && fJob.inTheSameVSuperPod(vSuperPodIds, fTask.NodeName) {
			klog.V(util.LogInfoLev).Infof("master pod and fault task is in the same vsuperpod")
			fJob.IsMasterFault = true
			return
		}
	}
	fJob.IsMasterFault = false
}

func (fJob *FaultJob) getVSuperPodIds() []string {
	var ids []string
	for _, task := range fJob.FaultTasks {
		if task.IsFaultTask {
			id := fJob.getVSuperPodId(task.NodeName)
			if id != "" {
				ids = append(ids, id)
			}
		}
	}
	klog.V(util.LogInfoLev).Infof("getVSuperPodIds super pod vSuperPodIds:%v", ids)
	return ids
}

// getVSuperPodId get virtual super pods pods id
func (fJob *FaultJob) getVSuperPodId(node string) string {
	for id, superNodes := range fJob.SuperPods {
		for _, superNode := range superNodes {
			if superNode.Name == node {
				return id
			}
		}
	}
	return ""
}

func (fJob *FaultJob) forceDeletePodsFor910A5(schedulerJob *plugin.SchedulerJob, env plugin.ScheduleEnv,
	dpi *deletePodInfo) {
	var waitDeleteTask = make([]FaultTask, 0)
	klog.V(util.LogInfoLev).Infof("whetherBackToVspSchedule is %v", fJob.WhetherBackToVspSchedule)
	for id, fTask := range fJob.FaultTasks {
		klog.V(util.LogDebugLev).Infof("not masterFault is %v, job single rescheduling is %v, not fault task is %v",
			!fJob.IsMasterFault, fJob.IsJobSingleRescheduling(schedulerJob), !fTask.IsFaultTask)
		if fJob.skipThisTask(dpi, fTask, env, schedulerJob) {
			klog.V(util.LogDebugLev).Infof(
				"rescheduling: skip force delete task<%s> for job<%s/%s>", fTask.TaskName, fJob.JobNamespace,
				fJob.JobName)
			continue
		}
		klog.V(util.LogDebugLev).Infof(
			"rescheduling: start to force delete task<%s> for job<%s/%s>", fTask.TaskName, fJob.JobNamespace,
			fJob.JobName)
		if fTask.NodeRankIndex == util.Rank0 {
			klog.V(util.LogInfoLev).Infof("master pod will be deleted, set fJob.IsMasterFault true")
			fJob.IsMasterFault = true
		}
		fJob.FaultTasks[id].IsSatisfiedRackAffinity = false
		fJob.updateSuperPodMapInfo(env, fTask.TaskName, fTask.NodeName)
		waitDeleteTask = append(waitDeleteTask, fTask)
		klog.V(util.LogInfoLev).Infof("superpod delete pods:%s", fTask.TaskName)
	}
	fJob.deletingTasksConcurrently(waitDeleteTask, env.FrameAttr.KubeClient)
}

func (fJob *FaultJob) skipThisTask(dpi *deletePodInfo, fTask FaultTask, env plugin.ScheduleEnv,
	schedulerJob *plugin.SchedulerJob) bool {
	// when master pod fault or not pod rescheduling or fault pod, delete pod
	if fJob.IsMasterFault {
		return false
	}
	if res, continueProcess := fJob.processReschedulingSkipTaskforA5(fTask, schedulerJob); !continueProcess {
		return res
	}
	// pod rescheduling and is not fault task
	return fJob.podReschedulingSkipTask(dpi, fTask, schedulerJob)
}

func (fJob *FaultJob) processReschedulingSkipTaskforA5(fTask FaultTask,
	schedulerJob *plugin.SchedulerJob) (bool, bool) {
	// the first boolean value indicates whether the task can be skipped
	// the second boolean value indicates whether to proceed to next check
	if fJob.IsProcessReschedulingJob(schedulerJob) {
		if fTask.IsFaultTask || fJob.inTheSameTpBlock(fTask) {
			klog.V(util.LogInfoLev).Infof("delete pod in same tp block")
			return false, false
		}
		klog.V(util.LogInfoLev).Infof("process rescheduling will not delete pod, for no pod in same tp has failed")
		return true, false
	}
	if fJob.IsProcessReschedulePause(schedulerJob) {
		klog.V(util.LogInfoLev).Infof("process rescheduling failed, back to job rescheduling, delete all pods")
		return false, false
	}
	return false, true
}

func (fJob *FaultJob) IsProcessReschedulePause(sJob *plugin.SchedulerJob) bool {
	return sJob.Label[util.ProcessRecoverEnable] == util.ProcessRecoverPause
}

func (fJob *FaultJob) podReschedulingSkipTask(dpi *deletePodInfo, fTask FaultTask,
	schedulerJob *plugin.SchedulerJob) bool {
	// boolean value indicates whether the task can be skipped
	if fJob.IsJobSingleRescheduling(schedulerJob) && !fTask.IsFaultTask {
		if !dpi.isSuperPod {
			klog.V(util.LogInfoLev).Infof("skip because %s is not super pod", fJob.JobName)
			return true
		}
		// single pod rescheduling stage, delete no pod
		// virtual super pod rescheduling stage, delete all virtual super pod where fault task in
		// virtual rack rescheduling stage, delete all virtual rack where fault task in
		if fJob.PendingSessionNum < spPendingTimes {
			klog.V(util.LogInfoLev).Infof("skip because %s rescheduling is the first stage", fJob.JobName)
			return true
		}
		if fJob.WhetherBackToVspSchedule && !fJob.inTheSameVSuperPod(dpi.ids, fTask.NodeName) {
			klog.V(util.LogInfoLev).Infof("skip because %s and fault task is not inTheSameVSuperPod", fTask.TaskName)
			return true
		}
		if !fJob.WhetherBackToVspSchedule && !fJob.inTheSameTpBlock(fTask) {
			klog.V(util.LogInfoLev).Infof("skip because %s and fault task is not inTheSameTpBlock", fTask.TaskName)
			return true
		}
	}
	return false
}

func (fJob *FaultJob) inTheSameVSuperPod(ids []string, nodeName string) bool {
	if !fJob.WhetherBackToVspSchedule {
		klog.V(util.LogInfoLev).Infof("leave inTheSameVSuperPod because rescheduling hasn't back to vspschedule")
		return false
	}
	for _, v := range ids {
		nodes, ok := fJob.SuperPods[v]
		if !ok {
			klog.V(util.LogErrorLev).Infof("superpod id does not exist in fJob.SuperPods, id: %s", v)
			return false
		}
		for _, each := range nodes {
			if each.Name == nodeName {
				return true
			}
		}
	}
	return false
}

func (fJob *FaultJob) inTheSameTpBlock(fTask FaultTask) bool {
	if fJob.TpBlock <= forceRackAffinityLimit {
		return false
	}
	for _, task := range fJob.FaultTasks {
		if !task.IsFaultTask {
			continue
		}
		targetRankId, err := strconv.Atoi(task.NodeRankIndex)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("target task NodeRankIndex cannot be converted to int, targetRankId:"+
				" %s", task.NodeRankIndex)
			return false
		}
		targetRankId = targetRankId - targetRankId%fJob.TpBlock
		fTaskRankId, err := strconv.Atoi(fTask.NodeRankIndex)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("ftask NodeRankIndex cannot be converted to int, targetRankId: %s",
				task.NodeRankIndex)
			return false
		}
		if fTaskRankId >= targetRankId && fTaskRankId < targetRankId+fJob.TpBlock {
			return true
		}
	}
	return false
}
