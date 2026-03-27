/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// GetJobFaultRescheduleLabel Get job's fault reschedule label.
func (fJob *FaultJob) GetJobFaultRescheduleLabel() string {
	if fJob == nil {
		klog.V(util.LogErrorLev).Info(
			"GetJobFaultRescheduleLabel fJob or schedulerJob does not exist")
		return JobOffRescheduleLabelValue
	}
	value, ok := fJob.Labels[JobRescheduleLabelKey]
	if !ok {
		klog.V(util.LogDebugLev).Infof(
			"GetJobFaultRescheduleLabel %s. %s no job reschedule label", value, fJob.JobUID)
		return JobOffRescheduleLabelValue
	}
	klog.V(util.LogDebugLev).Infof("GetJobFaultRescheduleLabel job: %s, label: %s", fJob.JobUID, value)
	return value
}

// IsNormalJobNeedRestart is Job has the key of PreSeparateNPU os Job has software fault
func (fJob *FaultJob) IsNormalJobNeedRestart() bool {
	if fJob == nil {
		return false
	}
	for _, fTask := range fJob.FaultTasks {
		if fTask.IsSoftwareFault {
			klog.V(util.LogWarningLev).Infof("task %s has software fault, need restart by platform", fTask.TaskName)
			return true
		}
		for _, reason := range fTask.Reason {
			if reason.FaultHandling == PreSeparateNPU {
				klog.V(util.LogWarningLev).Infof("task %s has PreSeparateNPU fault, "+
					"need restart by platform", fTask.TaskName)
				return true
			}
		}
	}
	return false
}

func (fJob *FaultJob) isJobGraceDeleteSuccess(jobInfo *api.JobInfo, tag910A5 bool) bool {
	restartNum := 0
	for _, fTask := range fJob.FaultTasks {
		podCreateTimeRecord := fTask.PodCreateTime // former pods create time
		npuTask, ok := jobInfo.Tasks[fTask.TaskUID]
		if !ok { // task not in job indicates it has been restarted
			restartNum++
			klog.V(util.LogInfoLev).Infof("%s in %s has been deleted", fTask.TaskName, jobInfo.Name)
			continue
		}
		podCreateTimeCur := npuTask.Pod.CreationTimestamp.Unix() // current pod create time
		if podCreateTimeCur != podCreateTimeRecord {
			klog.V(util.LogInfoLev).Infof("pod %s restart success[new:%v---old:%v]", npuTask.Pod.Name, podCreateTimeCur,
				podCreateTimeRecord)
			restartNum++
			continue
		}

		if fTask.IsFaultTask {
			klog.V(util.LogInfoLev).Infof("fault task<%s> is still being deleted", fTask.TaskName)
			return false
		}

		if fTask.IsBeingGracefulDeleted {
			klog.V(util.LogInfoLev).Infof("normal task<%s> is still being deleted", fTask.TaskName)
			return false
		}
	}

	if len(fJob.Labels) != 0 && (fJob.Labels[util.SinglePodTag] == util.EnableFunc ||
		fJob.Labels[util.ProcessRecoverEnable] == util.EnableFunc) &&
		(tag910A5 || fJob.PendingSessionNum != pendingTimes) {
		klog.V(util.LogInfoLev).Info("grace deletion for pod-level rescheduling complete")
		return true
	}

	klog.V(util.LogInfoLev).Infof("<%d/%d> pod of job restarted", restartNum, len(fJob.FaultTasks))
	return restartNum >= len(fJob.FaultTasks)
}

// deleteJobWithLabels delete job with labels
func (fJob *FaultJob) deleteJobWithLabels(ssn *framework.Session, reschedule *ReScheduler,
	schedulerJob *plugin.SchedulerJob, env plugin.ScheduleEnv) error {
	if !fJob.isFaultJobCanRestarted(reschedule) {
		return fmt.Errorf("job <%s> is not fault job or pod failed job reach max restart time", fJob.JobName)
	}
	if fJob.IsSubHealthFault {
		return fJob.deleteJobWithSubHealthyLabels(ssn, schedulerJob, env)
	}
	return fJob.deleteJobWithFaultLabels(ssn, reschedule, schedulerJob, env)
}

func (fJob *FaultJob) deleteJobWithSubHealthyLabels(ssn *framework.Session, schedulerJob *plugin.SchedulerJob,
	env plugin.ScheduleEnv) error {
	klog.V(util.LogWarningLev).Infof("delete job %s subHealthyStrategy %s", fJob.JobName,
		fJob.SubHealthyStrategy)

	if fJob.SubHealthyStrategy == util.SubHealthyForceExit || fJob.SubHealthyStrategy == util.SubHealthyHotSwitch {
		if is910A5SuperPodJob(schedulerJob) {
			return fJob.ForceDeleteJobFor910A5(schedulerJob, env)
		}
		return fJob.ForceDeleteJob(schedulerJob, env)
	}
	if !fJob.IsProcessReschedulingJob(schedulerJob) {
		klog.V(util.LogWarningLev).Infof("reset configmap be updated with graceExit, job:%s", fJob.JobName)
		updateResetConfigMapWithGraceExit(env.FrameAttr.KubeClient, plugin.ResetInfoCMNamePrefix+
			schedulerJob.ReferenceName, schedulerJob.NameSpace, plugin.GraceExitValue)
	}
	if is910A5SuperPodJob(schedulerJob) {
		return fJob.GraceDeleteJobFor910A5(ssn, schedulerJob, env)
	}
	return fJob.GraceDeleteJob(ssn, schedulerJob, env)
}

// deleteJobWithLabels delete job with labels
func (fJob *FaultJob) deleteJobWithFaultLabels(ssn *framework.Session, reschedule *ReScheduler,
	schedulerJob *plugin.SchedulerJob, env plugin.ScheduleEnv) error {
	klog.V(util.LogWarningLev).Infof("delete job %s ReScheduleKey %s", fJob.JobName,
		fJob.ReScheduleKey)
	if fJob.ReScheduleKey == JobForceRescheduleLabelValue {
		if is910A5SuperPodJob(schedulerJob) {
			return fJob.ForceDeleteJobFor910A5(schedulerJob, env)
		}
		return fJob.ForceDeleteJob(schedulerJob, env)
	}
	if is910A5SuperPodJob(schedulerJob) {
		return fJob.GraceDeleteJobFor910A5(ssn, schedulerJob, env)
	}
	return fJob.GraceDeleteJob(ssn, schedulerJob, env)
}

// getVirSupPodId get virtual super pods pods id
func (fJob *FaultJob) getVirSupPodId(node string, env plugin.ScheduleEnv) string {
	if _, ok := env.SuperPodInfo.SuperPodReschdInfo[fJob.JobUID]; !ok {
		return ""
	}
	for id, v := range env.SuperPodInfo.SuperPodReschdInfo[fJob.JobUID] {
		for _, each := range v {
			if each.Name == node {
				return id
			}
		}
	}
	return ""
}

func (fJob *FaultJob) isContainTask(ids []string, nodeName string, env plugin.ScheduleEnv) bool {
	if _, ok := env.SuperPodInfo.SuperPodReschdInfo[fJob.JobUID]; !ok {
		return false
	}
	for _, v := range ids {
		nodes, ok := env.SuperPodInfo.SuperPodReschdInfo[fJob.JobUID][v]
		if !ok {
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

func (fJob *FaultJob) getDeletePodInfo(
	schedulerJob *plugin.SchedulerJob, env plugin.ScheduleEnv) (*deletePodInfo, error) {
	if schedulerJob.IsMultiLevelJob() {
		return fJob.getMultiLevelJobDeletePodInfo(schedulerJob, env)
	}
	return fJob.getNormalJobDeletePodInfo(schedulerJob)
}

func (fJob *FaultJob) getMultiLevelJobDeletePodInfo(schedulerJob *plugin.SchedulerJob,
	env plugin.ScheduleEnv) (*deletePodInfo, error) {
	needRescheduleNodes, err := fJob.getNeedReschedulePods(schedulerJob, env)
	if err != nil {
		return &deletePodInfo{}, err
	}
	return &deletePodInfo{
		needRescheduleNodes: needRescheduleNodes,
	}, nil
}

func (fJob *FaultJob) getNormalJobDeletePodInfo(schedulerJob *plugin.SchedulerJob) (*deletePodInfo, error) {
	isSuperPod := false
	ids := make([]string, 0)
	isSuperPod, ids = fJob.getFaultJobSuperPodInfo(schedulerJob)
	return &deletePodInfo{
		isSuperPod: isSuperPod,
		ids:        ids,
	}, nil
}

func (fJob *FaultJob) getNeedReschedulePods(
	schedulerJob *plugin.SchedulerJob, env plugin.ScheduleEnv) (map[string]struct{}, error) {
	treeLevels, err := util.GetTaskTreeLevels(schedulerJob.AffinityBlocks, schedulerJob.NPUTaskNum)
	if err != nil {
		return nil, fmt.Errorf("getTaskTreeLevels failed: %v", err)
	}
	resourceLevel, err := fJob.getResourceLevels(fJob.SuperPods, env)
	if err != nil {
		return nil, fmt.Errorf("getResourceLevels failed: %v", err)
	}

	taskTree, err := plugin.GetTaskTreeFromSuperNodeMap(fJob.SuperPods, treeLevels, resourceLevel, env.Nodes)
	if err != nil {
		return nil, fmt.Errorf("getTaskTreeFromSuperNodeMap failed: %v", err)
	}

	faultNodeMap := make(map[string]struct{})
	for _, task := range fJob.FaultTasks {
		if !task.IsFaultTask {
			continue
		}
		faultNodeMap[task.NodeName] = struct{}{}
	}
	reschedulingLevel := fJob.PendingSessionNum / spPendingTimes
	if reschedulingLevel > len(taskTree.Levels)-1 {
		reschedulingLevel = len(taskTree.Levels) - 1
	}
	needRescheduleBlocks := make(map[*util.TaskNode]struct{})
	for _, task := range taskTree.GetAllBaseTasks() {
		if _, ok := faultNodeMap[task.ResourceNodeName]; !ok {
			continue
		}
		for i := 0; i < reschedulingLevel; i++ {
			task = task.Parent
		}
		needRescheduleBlocks[task] = struct{}{}
	}

	needRescheduleNodes := make(map[string]struct{})
	for block := range needRescheduleBlocks {
		blockSubTree, err := taskTree.GetSubTree(block)
		if err != nil {
			return nil, fmt.Errorf("cannot find subtree for block %d, %v", block.Index, err)
		}
		for _, task := range blockSubTree.GetAllBaseTasks() {
			needRescheduleNodes[task.ResourceNodeName] = struct{}{}
		}
	}
	return needRescheduleNodes, nil
}

func (fJob *FaultJob) getResourceLevels(superPod map[string][]plugin.SuperNode, env plugin.ScheduleEnv) ([]util.ResourceTreeLevel, error) {
	for _, nodes := range superPod {
		for _, node := range nodes {
			levels, ok := env.FrameAttr.ResourceLevelsInfo[node.TopoTreeName]
			if !ok {
				continue
			}
			return levels, nil
		}
	}
	return nil, errors.New("no match resource levels found")

}

func (fJob *FaultJob) isMasterFault(dpi *deletePodInfo, isNpuJob bool) bool {
	masterFault := false
	for _, task := range fJob.FaultTasks {
		if task.IsFaultTask && isMasterRank(task, isNpuJob) {
			masterFault = true
			break
		}
	}
	masterFault = masterFault || fJob.isMasterPodAffectedByFault(dpi, isNpuJob)
	if masterFault {
		klog.V(util.LogInfoLev).Infof("job %v/%v is master fault", fJob.JobNamespace, fJob.JobName)
	}
	return masterFault
}

func (fJob *FaultJob) isMasterPodAffectedByFault(dpi *deletePodInfo, isNpuJob bool) bool {
	return fJob.isMasterFaultWithMultiLevelJob(dpi, isNpuJob) || fJob.isMasterFaultWithSuperPodJob(dpi, isNpuJob)
}

func (fJob *FaultJob) isMasterFaultWithMultiLevelJob(dpi *deletePodInfo, isNpuJob bool) bool {
	if dpi == nil || dpi.needRescheduleNodes == nil {
		return false
	}
	for _, task := range fJob.FaultTasks {
		if !task.IsNpuTask {
			continue
		}
		_, ok := dpi.needRescheduleNodes[task.NodeName]
		if ok && isMasterRank(task, isNpuJob) {
			return true
		}
	}
	return false
}

func (fJob *FaultJob) isMasterFaultWithSuperPodJob(dpi *deletePodInfo, isNpuJob bool) bool {
	if dpi == nil || !dpi.isSuperPod {
		return false
	}
	for _, task := range fJob.FaultTasks {
		if !isMasterRank(task, isNpuJob) {
			continue
		}
		if fJob.PendingSessionNum >= spPendingTimes && fJob.inTheSameVSuperPod(dpi.ids, task.NodeName) {
			return true
		}
	}
	return false
}

func isMasterRank(fTask FaultTask, isNpuJob bool) bool {
	if fTask.NodeRankIndex != util.Rank0 || !isNpuJob {
		return false
	}
	isMaster := true
	replicaType, ok := fTask.Labels[util.ReplicaTypeKey]
	if ok && replicaType == util.ReplicaTypeValueWorker {
		isMaster = false
	}
	return isMaster
}

// ForceDeleteJob force delete jobs includes labelled force delete ones and grace delete failed ones
func (fJob *FaultJob) ForceDeleteJob(schedulerJob *plugin.SchedulerJob,
	env plugin.ScheduleEnv) error {
	klog.V(util.LogDebugLev).Infof("enter ForceDeleteJob")
	if fJob == nil || schedulerJob == nil {
		return fmt.Errorf(
			"getJobFaultRescheduleLabel fJob object or ssn or schedulerJob does not exist")
	}
	dpi, err := fJob.getDeletePodInfo(schedulerJob, env)
	if err != nil {
		return fmt.Errorf("unable to find the pods to delete, %v", err)
	}
	dpi.isMasterFault = fJob.isMasterFault(dpi, schedulerJob.IsNPUJob())
	fJob.updateSuperPodsReschdInfo(env)
	fJob.forceDeletePods(schedulerJob, env, dpi)
	return nil
}

func (fJob *FaultJob) getFaultJobSuperPodInfo(schedulerJob *plugin.SchedulerJob) (bool, []string) {
	isSuperPod := false
	ids := make([]string, 0)
	if schedulerJob.IsSuperPodJob() {
		isSuperPod = true
		ids = fJob.getVSuperPodIds()
	}
	return isSuperPod, ids
}

func (fJob *FaultJob) forceDeletePods(schedulerJob *plugin.SchedulerJob,
	env plugin.ScheduleEnv, dpi *deletePodInfo) {
	var waitDeleteTask = make([]FaultTask, 0)
	for _, fTask := range fJob.FaultTasks {
		if !fJob.isNormalTaskCanBeDelete(fTask, schedulerJob, env, dpi) {
			klog.V(util.LogDebugLev).Infof(
				"rescheduling: skip force delete task<%s> for job<%s/%s>", fTask.TaskName, fJob.JobNamespace,
				fJob.JobName)
			continue
		}
		fJob.updateSuperPodMapInfo(env, fTask.TaskName, fTask.NodeName)
		waitDeleteTask = append(waitDeleteTask, fTask)
		klog.V(util.LogInfoLev).Infof("superpod delete pods:%s", fTask.TaskName)
	}
	fJob.deletingTasksConcurrently(waitDeleteTask, env.FrameAttr.KubeClient)
}

func (fJob *FaultJob) isNormalTaskCanBeDelete(fTask FaultTask, schedulerJob *plugin.SchedulerJob,
	env plugin.ScheduleEnv, dpi *deletePodInfo) bool {
	klog.V(util.LogDebugLev).Infof("not masterFault is %v, job single rescheduling is %v ,"+
		"not fault task is %v, allow upgrade is %v", !dpi.isMasterFault, fJob.IsJobSingleRescheduling(schedulerJob),
		!fTask.IsFaultTask, fJob.allowUpgradePodRescheduling())

	// if upgrade is not allowed, only the fault task should be rescheduled
	if !fJob.allowUpgradePodRescheduling() {
		return fTask.IsFaultTask
	}

	if dpi.isMasterFault {
		return true
	}
	// when pod rescheduling, master fault try job rescheduling
	// tor affinity job, first delete fault task, when first task pending 12 session, try job rescheduling
	// super pod, first delete fault task, when first task pending 6 session, try super pod rescheduling
	// super pod, when pod pending 12 session, try job rescheduling
	if fJob.IsJobSingleRescheduling(schedulerJob) && !fTask.IsFaultTask {
		// needRescheduleNodes only used by multilevel job
		if _, ok := dpi.needRescheduleNodes[fTask.NodeName]; ok && fTask.IsNpuTask {
			return true
		}
		if !dpi.isSuperPod {
			return false
		}
		// single pod rescheduling stage, delete no pod
		// virtual super pod rescheduling stage, delete all virtual super pod where fault task in
		if fJob.PendingSessionNum < spPendingTimes {
			return false
		} else if !fJob.isContainTask(dpi.ids, fTask.NodeName, env) {
			return false
		}
	}
	// when master pod fault try job rescheduling
	// when job is process rescheduling and not in pod rescheduling label, only delete fault task.
	// when  process rescheduling is failed, process-rescheduling change to pause, try job rescheduling
	if fJob.IsProcessReschedulingJob(schedulerJob) && !fTask.IsFaultTask {
		return false
	}
	return true
}

func (fJob *FaultJob) allowUpgradePodRescheduling() bool {
	return fJob.ReScheduleLimit != util.ReschedulingUpperLimitPod
}

func (fJob *FaultJob) deletingTasksConcurrently(waitDeleteTask []FaultTask, kubeClient kubernetes.Interface) {
	if len(waitDeleteTask) == 0 {
		return
	}
	if len(waitDeleteTask) <= singleThreadDeletePodNum {
		fJob.forceDeleteTasks(waitDeleteTask, kubeClient)
		return
	}
	var deleteJobSync sync.WaitGroup
	for i := 0; i < len(waitDeleteTask); i += singleThreadDeletePodNum {
		deleteJobSync.Add(1)
		if i+singleThreadDeletePodNum > len(waitDeleteTask) {
			go fJob.forceDeleteTasksConcurrently(waitDeleteTask[i:], kubeClient, &deleteJobSync)
		} else {
			go fJob.forceDeleteTasksConcurrently(waitDeleteTask[i:i+singleThreadDeletePodNum], kubeClient,
				&deleteJobSync)
		}
	}
	deleteJobSync.Wait()
}

func (fJob *FaultJob) forceDeleteTasksConcurrently(waitDeleteTask []FaultTask, kubeClient kubernetes.Interface,
	deleteJobSync *sync.WaitGroup) {
	fJob.forceDeleteTasks(waitDeleteTask, kubeClient)
	deleteJobSync.Done()
}

func (fJob *FaultJob) forceDeleteTasks(waitDeleteTask []FaultTask, kubeClient kubernetes.Interface) {
	for _, fTask := range waitDeleteTask {
		if deleteErr := fTask.DeleteRealPodByTask(kubeClient, 0); deleteErr != nil {
			klog.V(util.LogWarningLev).Infof("ForceDeleteFaultPod %s: %s.", fTask.TaskName, util.SafePrint(deleteErr))
		}
	}
}

func (fJob *FaultJob) updateSuperPodMapInfo(env plugin.ScheduleEnv, taskName string, nodeName string) {
	if nodeName == "" {
		klog.V(util.LogInfoLev).Infof("updateSuperPodMapInfo nodeName is empty")
		return
	}
	mapInfo := make(map[string]string)
	if _, ok := env.SuperPodInfo.SuperPodMapFaultTaskNodes[fJob.JobUID]; ok {
		mapInfo = env.SuperPodInfo.SuperPodMapFaultTaskNodes[fJob.JobUID]
	}
	mapInfo[taskName] = nodeName
	env.SuperPodInfo.SuperPodMapFaultTaskNodes[fJob.JobUID] = mapInfo
}

func (fJob *FaultJob) updateSuperPodsReschdInfo(env plugin.ScheduleEnv) {
	if _, ok := env.SuperPodInfo.SuperPodFaultTaskNodes[fJob.JobUID]; !ok {
		var nodesName []string
		for _, fTask := range fJob.FaultTasks {
			if fTask.IsFaultTask && fTask.NodeName != "" {
				nodesName = append(nodesName, fTask.NodeName)
			}
		}
		env.SuperPodInfo.SuperPodFaultTaskNodes[fJob.JobUID] = nodesName
	}
	if len(fJob.SuperPods) > 0 {
		env.SuperPodInfo.SuperPodReschdInfo[fJob.JobUID] = fJob.SuperPods
	}
	klog.V(util.LogInfoLev).Infof("cache superpods %+v", env.SuperPodInfo.SuperPodReschdInfo[fJob.JobUID])
}

// IsJobSingleRescheduling valid job.
func (fJob *FaultJob) IsJobSingleRescheduling(sJob *plugin.SchedulerJob) bool {
	if sJob.Label[util.SinglePodTag] == util.EnableFunc && fJob.PendingSessionNum < pendingTimes {
		return true
	}
	return false
}

// IsProcessReschedulingJob valid job.
func (fJob *FaultJob) IsProcessReschedulingJob(sJob *plugin.SchedulerJob) bool {
	return sJob.Label[util.ProcessRecoverEnable] == util.EnableFunc
}

func (fJob *FaultJob) updateTaskPodUid(jobInfo *api.JobInfo) {
	for i := 0; i < len(fJob.FaultTasks); i++ {
		fJob.FaultTasks[i].TaskUID = getTaskPodUidByTaskName(fJob.FaultTasks[i].TaskName, jobInfo)
	}
}

func getTaskPodUidByTaskName(taskName string, jobInfo *api.JobInfo) api.TaskID {
	for _, task := range jobInfo.Tasks {
		if task.Name == taskName && task.Pod != nil {
			return api.TaskID(task.Pod.UID)
		}
	}
	return ""
}

func (fJob *FaultJob) updateFaultJobWhenNewPodError(jobInfo *api.JobInfo) {
	if fJob.Labels[util.SinglePodTag] != util.EnableFunc &&
		fJob.Labels[util.ProcessRecoverEnable] != util.EnableFunc {
		return
	}
	newFailedTask := make(map[api.TaskID]struct{})
	for taskId, task := range jobInfo.Tasks {
		if isFailedTask(task) {
			newFailedTask[taskId] = struct{}{}
		}
	}
	if len(newFailedTask) == 0 {
		return
	}
	fJob.DeleteExecutedFlag = false

	for i, fTask := range fJob.FaultTasks {
		if fTask.IsFaultTask {
			continue
		}
		if _, ok := newFailedTask[fTask.TaskUID]; ok {
			fJob.FaultTasks[i].IsFaultTask = true
			fJob.FaultTasks[i].faultType = PodFailed
		}
	}
}

func isFailedTask(task *api.TaskInfo) bool {
	if task == nil || task.Pod == nil {
		return false
	}
	// pod phase is failed, when RestartPolicy is Never, means task failed
	if task.Pod.Status.Phase == v1.PodFailed {
		klog.V(util.LogInfoLev).Infof("task [name=%v, pod phase=%v] failed", task.Name, v1.PodFailed)
		return true
	}

	// any pod's container terminated with none-zero, means task failed
	for _, status := range task.Pod.Status.ContainerStatuses {
		if status.LastTerminationState.Terminated != nil &&
			status.LastTerminationState.Terminated.ExitCode != 0 {
			klog.V(util.LogInfoLev).Infof("task [name=%v, containerName=%v] failed",
				task.Name, status.Name)
			return true
		}
	}
	return false
}

// isPodFailedCanRestarted if the pod status is failed, judge whether job can be restarted
func (fJob *FaultJob) isFaultJobCanRestarted(reScheduler *ReScheduler) bool {
	if !fJob.IsFaultJob {
		klog.V(util.LogWarningLev).Infof("fJob %s has software fault or PreSeparateNPU fault, "+
			"need restarted by platform skip delete pod", fJob.JobName)
		return false
	}
	if fJob.faultReason == PodFailed {
		if fJob.FaultRetryTimes == 0 {
			klog.V(util.LogInfoLev).Infof("job<%s> retry times is 0", fJob.JobUID)
			return false
		}
		remain, ok := reScheduler.JobRemainRetryTimes[fJob.JobUID]
		if !ok || remain == nil {
			return false
		}
		if remain.Times <= 0 {
			klog.V(util.LogInfoLev).Infof("job<%s> remain retry times: %d", fJob.JobUID,
				reScheduler.JobRemainRetryTimes[fJob.JobUID].Times)
			return false
		}
	}
	return true
}

// GraceDeleteJob grace delete jobs labelled to be deleted gracefully
func (fJob *FaultJob) GraceDeleteJob(ssn *framework.Session, npuJob *plugin.SchedulerJob,
	env plugin.ScheduleEnv) error {
	if fJob == nil {
		return fmt.Errorf("getJobFaultRescheduleLabel fJob object does not exist")
	}
	if ssn == nil {
		return fmt.Errorf("session does not exist")
	}
	if npuJob == nil {
		return fmt.Errorf("schedulerJob does not exist")
	}
	dpi, err := fJob.getDeletePodInfo(npuJob, env)
	if err != nil {
		return fmt.Errorf("unable to find the pods to delete, %v", err)
	}
	dpi.reason = fJob.getRestartInfos()
	dpi.isMasterFault = fJob.isMasterFault(dpi, npuJob.IsNPUJob())
	if dpi.isMasterFault {
		klog.V(util.LogInfoLev).Infof("job %v is master fault", npuJob.Name)
	}
	fJob.updateSuperPodsReschdInfo(env)
	err = fJob.graceDeletePods(ssn, npuJob, env, dpi)
	if err != nil {
		return err
	}
	return nil
}

func (fJob *FaultJob) graceDeletePods(ssn *framework.Session, npuJob *plugin.SchedulerJob, env plugin.ScheduleEnv,
	dpi *deletePodInfo) error {
	for id, fTask := range fJob.FaultTasks {
		npuTask, ok := npuJob.Tasks[fTask.TaskUID]
		if !ok {
			klog.V(util.LogDebugLev).Infof(
				"rescheduling: skip grace delete because task<%s> for job <%s/%s> has been deleted in session",
				fTask.TaskName, fJob.JobNamespace, fJob.JobName)
			continue
		}
		if !fJob.isNormalTaskCanBeDelete(fTask, npuJob, env, dpi) {
			klog.V(util.LogDebugLev).Infof(
				"rescheduling: skip grace delete task<%s> for job<%s/%s>", fTask.TaskName,
				fJob.JobNamespace, fJob.JobName)
			continue
		}
		klog.V(util.LogDebugLev).Infof(
			"rescheduling: start to grace delete task<%s> for job<%s/%s>", fTask.TaskName,
			fJob.JobNamespace, fJob.JobName)
		fJob.updateSuperPodMapInfo(env, fTask.TaskName, fTask.NodeName)
		if delErr := npuTask.ForceDeletePodByTaskInf(ssn, dpi.reason, fTask.NodeName); delErr != nil {
			klog.V(util.LogErrorLev).Infof("ForceDeletePodByTaskInf %s: %s.", npuTask.Name, delErr)
			continue
		}
		klog.V(util.LogDebugLev).Infof(
			"rescheduling: grace delete task<%s> for job<%s/%s> succeeded", fTask.TaskName,
			fJob.JobNamespace, fJob.JobName)
		fJob.FaultTasks[id].IsBeingGracefulDeleted = true
	}
	return nil
}

func (fJob *FaultJob) getRestartInfos() string {
	var reasonList []FaultReasonList
	for _, fTask := range fJob.FaultTasks {
		if fTask.Reason != nil {
			reasonList = append(reasonList, fTask.Reason...)
		}
	}
	reason := GetTaskRestartReason(reasonList)
	return reason
}

func (fJob *FaultJob) restartSingleFaultJob(ssn *framework.Session,
	reschedule *ReScheduler, schedulerJob *plugin.SchedulerJob, env plugin.ScheduleEnv) error {

	// delete jobs
	var deleteErr error

	switch fJob.ReScheduleKey {
	case JobForceRescheduleLabelValue, JobGraceRescheduleLabelValue:
		deleteErr = fJob.deleteJobWithLabels(ssn, reschedule, schedulerJob, env)
	case JobOffRescheduleLabelValue:
		deleteErr = fmt.Errorf("job reschedule %s", fJob.ReScheduleKey)
	default:
		deleteErr = fmt.Errorf("not support %s to reschedule job", fJob.ReScheduleKey)
	}

	return deleteErr
}

func (fJob *FaultJob) resetGraceExitCode(k8sClient kubernetes.Interface) {
	var isNotSubHealthFault bool
	for _, faultType := range fJob.FaultTypes {
		if faultType != SubHealthFault {
			isNotSubHealthFault = true
			break
		}
	}
	// if faultTypes contain other fault type, skip update exit code
	if isNotSubHealthFault || !fJob.IsFaultJob {
		return
	}
	updateResetConfigMapWithGraceExit(k8sClient, plugin.ResetInfoCMNamePrefix+
		fJob.ReferenceName, fJob.JobNamespace, plugin.DefaultExitValue)
}

func (fJob *FaultJob) jobInfoInSession(jobs map[api.JobID]*api.JobInfo) *api.JobInfo {
	if fJob.ElasticScheduling == JobOnElasticScheduling {
		for _, job := range jobs {
			// consider elastic scheduling which lead job uid/name change
			if job.Namespace == fJob.JobNamespace &&
				(job.Name == fJob.JobName || util.ReferenceNameOfJob(job) == fJob.ReferenceName) {
				return job
			}
		}
		return nil
	}

	job, ok := jobs[fJob.JobUID]
	if ok && util.UuidOfJob(job) == fJob.UUID {
		return job
	}

	return nil
}

func (fJob *FaultJob) getIsFaultJob() bool {
	for _, task := range fJob.FaultTasks {
		if task.IsFaultTask {
			klog.V(util.LogDebugLev).Infof(
				"job %s isFaultJob set true because of having fault tasks", fJob.JobName)
			return true
		}
	}
	return false
}

func (fJob *FaultJob) recordFaultJobsToLogs() {
	if !fJob.IsFaultJob {
		return
	}
	var tmpfJobInfo miniFaultJob
	tmpfJobInfo.ReferenceName = fJob.ReferenceName
	for _, fTask := range fJob.FaultTasks {
		if !fTask.IsFaultTask {
			continue
		}
		var tmpfTaskInfo miniFaultTask
		tmpfTaskInfo.Reason = fTask.Reason
		tmpfTaskInfo.TaskName = fTask.TaskName
		tmpfTaskInfo.NodeName = fTask.NodeName
		tmpfTaskInfo.FaultType = fTask.faultType
		tmpfTaskInfo.UseCardName = fTask.UseCardName
		tmpfTaskInfo.NodeRankIndex = fTask.NodeRankIndex
		if judgePublicFaultInReason(&tmpfTaskInfo) {
			tmpfTaskInfo.FaultType = PublicFaultType
		}
		tmpfJobInfo.FaultTasks = append(tmpfJobInfo.FaultTasks, tmpfTaskInfo)
	}
	str, err := json.Marshal(tmpfJobInfo)
	if err != nil {
		return
	}
	klog.V(util.LogWarningLev).Infof("Add FaultJob %s fault info: %s", fJob.JobName, string(str))
}

func (fJob *FaultJob) setJobFaultReScheduleLabel(value string) {
	fJob.ReScheduleKey = value
}

func (fJob *FaultJob) setJobElasticReScheduleLabel(value string) {
	fJob.ElasticScheduling = value
}

func (fJob *FaultJob) setFaultTasks(value []FaultTask) {
	fJob.FaultTasks = value
}

func (fJob *FaultJob) setIsSubHealthFault() {
	if !fJob.IsFaultJob {
		return
	}
	for _, faultType := range fJob.FaultTypes {
		if faultType != SubHealthFault {
			fJob.IsSubHealthFault = false
			return
		}
	}
	fJob.IsSubHealthFault = true
}

func (fJob *FaultJob) setIsFaultJob(value bool) {
	fJob.IsFaultJob = value
}

func (fJob *FaultJob) setJobSubHealthyStrategy() {
	subHealthyStrategy, exist := fJob.Labels[util.SubHealthyStrategyLabel]
	if !exist || !util.IsStrategyInSubHealthyStrategs(subHealthyStrategy) {
		subHealthyStrategy = util.SubHealthyIgnore
	}
	fJob.SubHealthyStrategy = subHealthyStrategy
}

func newFaultJobDefault(job *api.JobInfo, updateTime int64) *FaultJob {
	faultJob := &FaultJob{
		ReScheduleKey:     JobOffRescheduleLabelValue, // off/grace/force
		JobName:           job.Name,
		JobUID:            job.UID,
		JobNamespace:      job.Namespace,
		UpdateTime:        updateTime,
		FaultTypes:        make([]string, 0),
		ElasticScheduling: JobOffElasticScheduling,
		ReferenceName:     util.ReferenceNameOfJob(job),
		UUID:              util.UuidOfJob(job),
	}

	return faultJob
}

func (fJob *FaultJob) setCommonAttrFromScheduleJob(npuJob plugin.SchedulerJob) {
	fJob.Labels = npuJob.Label
	fJob.Annotations = npuJob.Annotation
	fJob.MinAvailable = npuJob.MinAvailable
}

func (fJob *FaultJob) setReScheduleLimit() {
	if _, exist := fJob.Labels[util.OmeInferenceServiceKey]; exist {
		fJob.ReScheduleLimit = util.ReschedulingUpperLimitPod
		return
	}
	if fJob.Labels[util.AppTypeLabelKey] == util.ControllerAppType ||
		fJob.Labels[util.AppTypeLabelKey] == util.CoordinatorAppType {
		fJob.ReScheduleLimit = util.ReschedulingUpperLimitPod
		return
	}
	fJob.ReScheduleLimit = fJob.Annotations[util.RecoverPolicyPathKey]
}

func (fJob *FaultJob) setFaultRetryTimeOfJob() {
	value, ok := fJob.Labels[FaultRetryTimesKey]
	if !ok {
		klog.V(util.LogDebugLev).Infof("get job<%s> label<%s> failed", fJob.JobUID, FaultRetryTimesKey)
		fJob.FaultRetryTimes = 0
		return
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("Failed to convert fault-retry-times <%s> of job <%s> into number",
			value, fJob.JobUID)
		fJob.FaultRetryTimes = 0
	}
	klog.V(util.LogDebugLev).Infof("get job: %s, fault-retry-times: %s", fJob.JobUID, value)
	fJob.FaultRetryTimes = v
}

func getRealFaultJobForCache(fJobs map[api.JobID]*FaultJob) map[api.JobID]*FaultJob {
	realFaultJobs := make(map[api.JobID]*FaultJob)
	for _, fJob := range fJobs {
		if fJob.IsFaultJob {
			realFaultJobs[fJob.JobUID] = fJob
		}
	}
	return realFaultJobs
}

func rebuildScheduledSuperPods(sJob plugin.SchedulerJob, env plugin.ScheduleEnv) map[string][]plugin.SuperNode {
	klog.V(util.LogWarningLev).Infof("rebuild scheduled super pod, job: %v/%v.", sJob.NameSpace, sJob.Name)
	if sJob.IsMultiLevelJob() {
		return reBuildMultiLevelSchedulingCache(sJob, env)
	}
	return reNormalJobSchedulingCache(sJob)
}

func reNormalJobSchedulingCache(sJob plugin.SchedulerJob) map[string][]plugin.SuperNode {
	superPods := map[string][]plugin.SuperNode{}
	for _, task := range sJob.Tasks {
		if task.Annotation == nil || task.NodeName == "" {
			continue
		}
		logicSpId, logicSpIdExist := task.Annotation[util.SuperPodRankKey]
		if !logicSpIdExist || logicSpId == "" {
			continue
		}
		superPodId, superPodIdExist := task.Annotation[util.SuperPodIdKey]
		if !superPodIdExist {
			continue
		}
		superPodID, err := strconv.Atoi(superPodId)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("rebuild scheduled super pod info failed, get task [%v] super pod id failed: %v.",
				task.Name, util.SafePrint(err))
			continue
		}
		superNodes, ok := superPods[logicSpId]
		if !ok {
			superPods[logicSpId] = []plugin.SuperNode{{SuperPodID: int32(superPodID), Name: task.NodeName}}
		} else {
			superPods[logicSpId] = append(superNodes,
				plugin.SuperNode{SuperPodID: int32(superPodID), Name: task.NodeName})
		}
	}
	klog.V(util.LogWarningLev).Infof("rebuild scheduled super pod, %v.", string(util.MarshalData(superPods)))
	sJob.SuperPods = superPods
	return superPods
}

func reBuildMultiLevelSchedulingCache(sJob plugin.SchedulerJob, env plugin.ScheduleEnv) map[string][]plugin.SuperNode {
	klog.V(util.LogWarningLev).Infof("start to rebuild multiLevel scheduling cache...")

	superPods := map[string][]plugin.SuperNode{}
	taskLevels, err := getMultiLevelAffinityConfig(sJob)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getMultiLevelAffinityConfig failed: %v", err)
		return superPods
	}
	var nodeTopoTreeName string
	for _, task := range sJob.Tasks {
		if task.Annotation == nil || task.NodeName == "" {
			continue
		}
		if nodeTopoTreeName == "" {
			nodeTopoTreeName, err = getNodeTopoTreeName(task.NodeName, env)
			if err != nil {
				klog.V(util.LogErrorLev).Infof("getNodeTopoTreeName failed: %v", err)
				continue
			}
			klog.V(util.LogInfoLev).Infof("rebuild nodeTopoTreeName: %v", nodeTopoTreeName)
		}

		hcclRankStr, rankExist := task.Annotation[plugin.PodRankIndexKey]
		if !rankExist || hcclRankStr == "" {
			continue
		}
		rankId, err := strconv.Atoi(hcclRankStr)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("getMultiLevelAffinityConfig failed: %v", err)
			continue
		}
		nodeDepth := len(taskLevels) - 1
		level1Depth := nodeDepth - util.Level1Number
		logicL1Rank := rankId / taskLevels[level1Depth].ReqNode
		localRank := rankId % taskLevels[level1Depth].ReqNode
		superNodes, ok := superPods[strconv.Itoa(logicL1Rank)]
		if !ok {
			superNodes = make([]plugin.SuperNode, taskLevels[level1Depth].ReqNode)
			superNodes[localRank] = plugin.SuperNode{Name: task.NodeName, TopoTreeName: nodeTopoTreeName}
		} else {
			superNodes[localRank] = plugin.SuperNode{Name: task.NodeName, TopoTreeName: nodeTopoTreeName}
		}
		superPods[strconv.Itoa(logicL1Rank)] = superNodes
	}
	klog.V(util.LogWarningLev).Infof("rebuild scheduled MultiLevel SchedulingCache, %v.", string(util.MarshalData(superPods)))
	sJob.SuperPods = superPods

	return superPods
}

func getMultiLevelAffinityConfig(sJob plugin.SchedulerJob) ([]util.TaskTreeLevel, error) {
	blocks, err := plugin.GetAffinityBlocks(sJob.Annotation)
	if err != nil {
		return nil, err
	}
	levels, err := util.GetTaskTreeLevels(blocks, sJob.NPUTaskNum)
	if err != nil {
		return nil, err
	}
	return levels, nil
}

func getNodeTopoTreeName(nodeName string, env plugin.ScheduleEnv) (string, error) {
	node, ok := env.Nodes[nodeName]
	if !ok {
		return "", fmt.Errorf("node %v not exist in env cache", nodeName)
	}
	if node.Label == nil {
		return "", fmt.Errorf("node %v don't have label", nodeName)
	}
	treeName, treeExist := node.Label[util.TopoTreeLabel]
	if !treeExist {
		treeName = util.DefaultTopoTree
	}
	return treeName, nil
}
