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
	"sort"
	"strconv"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/config"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// GetGraceDeleteTime Get the graceful delete time from configuration
func (reScheduler *ReScheduler) GetGraceDeleteTime(Conf []config.Configuration) (int64, error) {
	klog.V(util.LogInfoLev).Infof("enter GetGraceDeleteTime ...")
	defer klog.V(util.LogInfoLev).Infof("leave GetGraceDeleteTime ...")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("GetGraceDeleteTime failed: %s, nil reScheduler", util.ArgumentError)
		return DefaultGraceOverTime, errors.New(util.ArgumentError)
	}
	if len(Conf) == 0 {
		klog.V(util.LogErrorLev).Infof("GetGraceDeleteTime failed: %s, no conf", util.ArgumentError)
		return DefaultGraceOverTime, errors.New(util.ArgumentError)
	}
	// Read configmap
	configuration, err := util.GetConfigFromSchedulerConfigMap(util.CMInitParamKey, Conf)
	if err != nil {
		klog.V(util.LogErrorLev).Info("cannot get configuration, GraceOverTime will not be changed.")
		return DefaultGraceOverTime, nil
	}
	// get grace over time by user configuration
	overTimeStr, ok := configuration.Arguments[GraceOverTimeKey]
	if !ok {
		klog.V(util.LogErrorLev).Info("set GraceOverTime failed and will not be changed, " +
			"key grace-over-time doesn't exists.")
		return DefaultGraceOverTime, nil
	}
	overTime, err := strconv.ParseInt(overTimeStr, util.Base10, util.BitSize64)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("set GraceOverTime failed and will not be changed, "+
			"grace-over-time is invalid [%s].", util.SafePrint(overTimeStr))
		return DefaultGraceOverTime, err
	}
	// check time validity
	if !reScheduler.checkGraceDeleteTimeValid(overTime) {
		return DefaultGraceOverTime, errors.New("defaultGraceOverTime is out of range")
	}
	return overTime, nil
}

func (reScheduler *ReScheduler) setGraceOverTime(value int64) {
	reScheduler.GraceDeleteTime = value
}

// checkGraceDeleteTimeValid used by GetGraceDeleteTime for validity checking
func (reScheduler *ReScheduler) checkGraceDeleteTimeValid(overTime int64) bool {
	if overTime < minGraceOverTime || overTime > maxGraceOverTime {
		klog.V(util.LogErrorLev).Infof("GraceOverTime value should be range [2, 3600], configured is [%d], "+
			"GraceOverTime will not be changed", overTime)
		return false
	}
	// use user's configuration to set grace over time
	klog.V(util.LogInfoLev).Infof("set GraceOverTime to new value [%d].", overTime)
	return true
}

// createFaultTaskHandler Create FaultTask struct and set the corresponding values
func (reScheduler *ReScheduler) createFaultTaskHandler(job *api.JobInfo, cardName string) ([]FaultTask, error) {
	var faultTasks []FaultTask
	for _, task := range job.Tasks {
		faultTask := newFaultTaskDefault(task, job)
		// 2. updateNodeRankIndex by pod.Annotation
		tmpNodeRankIndex, err := faultTask.getNodeRankIndex(task)
		if err != nil {
			klog.V(util.LogInfoLev).Infof("getNodeRankIndex %s %s.", task.Name, util.SafePrint(err))
		}
		faultTask.setNodeRankIndex(tmpNodeRankIndex)
		// 3. update UseCardName
		tmpUseCardName, getErr := faultTask.getUseCardName(task, cardName)
		if getErr != nil {
			klog.V(util.LogInfoLev).Infof("getUseCardName %s %s", task.Name, util.SafePrint(getErr))
		}
		faultTask.setUseCardName(tmpUseCardName)
		isFaultTask, healthState := reScheduler.getTaskHealthState(&faultTask, task)
		klog.V(util.LogInfoLev).Infof("task %s is fault task: %v, health state: %s", task.Name, isFaultTask,
			healthState)
		faultTask.setIsFaultTask(isFaultTask)
		err = reScheduler.setTaskCardHealthCode(&faultTask)
		if err != nil {
			klog.V(util.LogDebugLev).Infof("setTaskCardHealthCode task %s err %s", task.Name, util.SafePrint(err))
		}
		faultTask.setFaultType(healthState)
		faultTasks = append(faultTasks, faultTask)
	}
	return faultTasks, nil
}

// GetRunningJobs get all the running jobs of <UseCardName> type
func (reScheduler *ReScheduler) GetRunningJobs(ssn *framework.Session) (map[api.JobID]*api.JobInfo, error) {
	var myJobs = make(map[api.JobID]*api.JobInfo, util.MapInitNum)
	for _, jobInfo := range ssn.Jobs {
		if (jobInfo.PodGroup.Status.Phase != util.PodGroupRunning) &&
			(jobInfo.PodGroup.Status.Phase != util.PodGroupUnknown) { // pending jobs would not be put into cache
			klog.V(util.LogInfoLev).Infof("job %s pod group is not running but %s, skip",
				jobInfo.Name, jobInfo.PodGroup.Status.Phase)
			continue
		}
		schedulerJob, ok := reScheduler.Jobs[jobInfo.UID]
		if !ok || schedulerJob.NPUJob == nil {
			klog.V(util.LogWarningLev).Infof("job %s not in session, skip", jobInfo.UID)
			continue
		}
		// req type is not current card type
		if schedulerJob.ReqNPUNum == 0 || len(schedulerJob.Selector) == 0 {
			klog.V(util.LogWarningLev).Infof("job %s requires npu %d and selector len is %d is illegal, skip",
				schedulerJob.Name, schedulerJob.ReqNPUNum, len(schedulerJob.Selector))
			continue
		}
		myJobs[jobInfo.UID] = jobInfo
	}
	if len(myJobs) == 0 {
		klog.V(util.LogDebugLev).Info("nil running jobs")
		return nil, fmt.Errorf("nil running jobs")
	}
	return myJobs, nil
}

func (reScheduler *ReScheduler) updateNewFaultJobAttr(faultJob FaultJob, jobInfo *api.JobInfo) FaultJob {

	npuJob := reScheduler.Jobs[faultJob.JobUID] // 1. set the value of ReScheduleKey, grace/force/off

	tmpElasticKey := faultJob.GetJobElasticSchedulingLabel(&npuJob)
	faultJob.setJobElasticReScheduleLabel(tmpElasticKey)

	tmpReScheduleKey := faultJob.GetJobFaultRescheduleLabel(&npuJob)
	faultJob.setJobFaultReScheduleLabel(tmpReScheduleKey)
	klog.V(util.LogInfoLev).Infof("job %s set rescheduleLabel %v", jobInfo.Name, tmpReScheduleKey)
	if tmpReScheduleKey == JobOffRescheduleLabelValue {
		klog.V(util.LogInfoLev).Infof("job %s rescheduleLabel off, skip rescheduling.", jobInfo.Name)
		return faultJob
	}
	npuName := util.GetNpuNameFromJobRequire(npuJob.ReqNPUName)

	// 2. create new FaultTask objects and update corresponding attributes
	tmpFaultTasks, err := reScheduler.createFaultTaskHandler(jobInfo, npuName)
	if err != nil {
		klog.V(util.LogInfoLev).Infof("job %s createFaultTaskHandler failed: %s", jobInfo.Name, util.SafePrint(err))
	}
	faultJob.setFaultTasks(tmpFaultTasks)
	tmpNodeNames := faultJob.getJobUseNodes() // 3. update the value of Job used nodeNames
	klog.V(util.LogDebugLev).Infof("job %s used nodes: %v", faultJob.JobName, tmpNodeNames)
	faultJob.setNodeNames(tmpNodeNames)
	tmpIsFaultJob := faultJob.getIsFaultJob() // 4. update the value of IsFaultJob
	klog.V(util.LogDebugLev).Infof("job %s if fault job: %v", faultJob.JobName, tmpIsFaultJob)
	faultJob.setIsFaultJob(tmpIsFaultJob)
	// 6. update FaultTypes of the job by status of FaultTasks bound on the job
	if faultJob.IsFaultJob {
		for _, fTask := range faultJob.FaultTasks {
			if fTask.IsFaultTask {
				faultJob.FaultTypes = append(faultJob.FaultTypes, fTask.faultType)
			}
		}
	}
	klog.V(util.LogDebugLev).Infof("job %s fault types: %v", faultJob.JobName, faultJob.FaultTypes)
	if npuName == util.NPU910CardName { // 5. update JobRankIds of fault cards
		tmpJobRankIds := reScheduler.getJobRankIdsFromTasks(&faultJob, util.NPU910CardNamePre)
		klog.V(util.LogDebugLev).Infof("job %s rankIds: %v", faultJob.JobName, tmpJobRankIds)
		faultJob.setJobRankIds(tmpJobRankIds)
		_, ok := reScheduler.JobRemainRetryTimes[faultJob.JobUID]
		if !ok {
			if reScheduler.JobRemainRetryTimes == nil {
				reScheduler.JobRemainRetryTimes = make(map[api.JobID]*RemainRetryTimes)
			}
			reScheduler.JobRemainRetryTimes[faultJob.JobUID] = &RemainRetryTimes{
				UUID:  faultJob.UUID,
				Times: faultJob.FaultRetryTimes,
			}
		}
	}
	return faultJob
}

// AddFaultJobWithSession read all running jobs of given card types and create the corresponding FaultJob objects
func (reScheduler *ReScheduler) AddFaultJobWithSession(jobs map[api.JobID]*api.JobInfo) error {
	klog.V(util.LogInfoLev).Infof("enter AddFaultJobWithSession ... ")
	defer klog.V(util.LogInfoLev).Infof("leave AddFaultJobWithSession ... ")
	if reScheduler == nil {
		klog.V(util.LogDebugLev).Infof("AddFaultJobWithSession: %s, nil reScheduler or job", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	klog.V(util.LogDebugLev).Infof("ReSchedulerCache fault jobs before add: %#v", reScheduler.FaultJobs)
	nowTime := time.Now().Unix()
	for _, jobInfo := range jobs {
		klog.V(util.LogDebugLev).Infof("ReSchedulerCache considering job %s", jobInfo.Name)
		flagInCache := false
		for _, fJob := range reScheduler.FaultJobs {
			if fJob.JobUID == jobInfo.UID ||
				(fJob.JobNamespace == jobInfo.Namespace && fJob.ReferenceName == util.ReferenceNameOfJob(jobInfo)) {
				flagInCache = true
				break
			}
		}
		// 1. jobs already in cache: go through the continue logic
		if flagInCache {
			continue
		}
		// 2. create FaultJob objects for jobs not in cache but sent by session
		klog.V(util.LogDebugLev).Infof("Add job %s to cache", jobInfo.Name)
		faultJob := newFaultJobDefault(jobInfo, nowTime)
		faultJob = reScheduler.updateNewFaultJobAttr(faultJob, jobInfo)
		faultJob.recordFaultJobsToLogs()
		reScheduler.FaultJobs = append(reScheduler.FaultJobs, faultJob)
	}
	klog.V(util.LogDebugLev).Infof("ReSchedulerCache fault jobs after add: %#v", reScheduler.FaultJobs)
	return nil
}

// getJobRankIdsFromTasks record the rankIds of fault cards on nodes of the given job
func (reScheduler ReScheduler) getJobRankIdsFromTasks(fJob *FaultJob, cardName string) []string {
	var jobRankIds []string

	for _, fTask := range fJob.FaultTasks {
		nodeRankIndex, err := strconv.Atoi(fTask.NodeRankIndex)
		if err != nil {
			klog.V(util.LogInfoLev).Infof("%s setJobRankIdsFromTasks convert %s", fTask.NodeRankIndex, util.SafePrint(err))
		}
		fNode := reScheduler.getFNodeOfGivenNameFromCache(fTask.NodeName)
		taskUseFaultCards, logicIds, err := fTask.getTaskUsedFaultCards(fNode, cardName, len(fJob.FaultTasks) > 1)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("taskUseFaultCards: %s", util.SafePrint(err))
			continue
		}
		klog.V(util.LogDebugLev).Infof("taskUseFaultCards: %v", taskUseFaultCards)
		for _, logicId := range logicIds {
			jobRankIds = append(jobRankIds, strconv.Itoa(logicId+nodeRankIndex*len(fTask.UseCardName)))
		}
	}
	klog.V(util.LogInfoLev).Infof("job %s rankIds: %v", fJob.JobName, jobRankIds)
	return jobRankIds
}

func GetTaskRestartReason(reasonList []FaultReasonList) string {
	str, err := json.Marshal(reasonList)
	if err != nil {
		klog.V(util.LogInfoLev).Infof("convertToReSchedulerJobsMapFromCM marshal: %s.", util.SafePrint(err))
		return ""
	}
	return string(str)
}

func (fTask *FaultTask) getTaskUsedFaultCards(fNode *FaultNode, cardName string, disFlag bool) ([]string, []int,
	error) {

	var cardIds []int
	for _, card := range fTask.UseCardName {
		taskUseCardID, err := strconv.Atoi(strings.TrimPrefix(card, cardName))
		if err != nil {
			klog.V(util.LogErrorLev).Infof("convert card ID %s to int failed", card)
			continue
		}
		cardIds = append(cardIds, taskUseCardID)
	}

	sort.Ints(cardIds)

	var logicIds []int
	if fTask.faultType == NodeUnhealthy || fTask.faultType == PodFailed { // node unhealthy returns all cards,
		// ahead of fNode equals to nil for node not in session
		klog.V(util.LogDebugLev).Infof("node unhealthy, return all NPUs used by task %s", fTask.TaskName)
		for index := range fTask.UseCardName {
			logicIds = append(logicIds, index)
		}
		return fTask.UseCardName, logicIds, nil
	}
	if fNode == nil {
		return nil, nil, fmt.Errorf("node is nil")
	}
	if fTask.NodeName != fNode.NodeName {
		return nil, nil, fmt.Errorf("fNode %s is not fTask %s's occupying node", fNode.NodeName, fTask.NodeName)
	}
	var taskUseFaultCard []string
	klog.V(util.LogDebugLev).Infof("task fault type: %s", fTask.faultType)
	for _, taskUseCard := range fTask.UseCardName {
		for _, fCard := range fNode.FaultCards {
			if taskUseCard != fCard.NPUName || !fCard.IsFaultCard {
				continue
			}
			if fCard.FaultType == CardNetworkUnhealthy && !disFlag {
				continue
			}
			taskUseFaultCard = append(taskUseFaultCard, taskUseCard)
		}
	}
	for _, faultCard := range taskUseFaultCard {
		taskUseCardID, err := strconv.Atoi(strings.TrimPrefix(faultCard, cardName))
		if err != nil {
			klog.V(util.LogErrorLev).Infof("convert card ID %s to int failed", faultCard)
			continue
		}
		for index, cardId := range cardIds {
			if cardId == taskUseCardID {
				logicIds = append(logicIds, index)
				break
			}
		}
	}
	return taskUseFaultCard, logicIds, nil
}

// getGraceDeleteFaultJobs get jobs needed to be deleted gracefully, only fault jobs with grace label would be selected
func (reScheduler ReScheduler) getGraceDeleteFaultJobs() []FaultJob {
	var graceDeleteJobs []FaultJob
	for _, fJob := range reScheduler.FaultJobs {
		if !fJob.IsFaultJob || fJob.ReScheduleKey != JobGraceRescheduleLabelValue {
			continue
		}
		graceDeleteJobs = append(graceDeleteJobs, fJob)
	}
	return graceDeleteJobs
}

// GetNeedForceDeleteDelayingNPUJobs get fault jobs with grace label but haven't been evicted successfully
func (reScheduler *ReScheduler) GetNeedForceDeleteDelayingNPUJobs(
	schedulerJobs map[api.JobID]plugin.SchedulerJob, ssn *framework.Session) ([]plugin.SchedulerJob, error) {
	klog.V(util.LogInfoLev).Infof("enter GetNeedForceDeleteDelayingNPUJobs ... ")
	defer klog.V(util.LogInfoLev).Infof("leave GetNeedForceDeleteDelayingNPUJobs ... ")
	if reScheduler == nil || len(schedulerJobs) == 0 || ssn == nil {
		klog.V(util.LogErrorLev).Infof("GetNeedForceDeleteDelayingNPUJobs: %s, "+
			"nil reScheduler or schedulerJobs or session", util.ArgumentError)
		return nil, errors.New(util.ArgumentError)
	}
	var forceJobs []plugin.SchedulerJob
	graceDeleteFaultJobs := reScheduler.getGraceDeleteFaultJobs()
	for _, fJob := range graceDeleteFaultJobs {
		jobInfo := fJob.jobInfoInSession(ssn.Jobs)
		if jobInfo == nil {
			klog.V(util.LogDebugLev).Infof(
				"GetNeedForceDeleteDelayingNPUJobs %v not in ssn.Jobs.", fJob.JobName)
		}
		if fJob.isJobGraceDeleteSuccess(jobInfo) { // if job successfully restarted, do not force delete
			continue
		}
		if !reScheduler.isDelayingJobTimeout(&fJob) { // if job not restarted and not time out, do not force delete
			continue
		}
		klog.V(util.LogWarningLev).Infof("grace delete job %s is time out for force delete.", fJob.JobName)
		schedulerJob, ok := schedulerJobs[fJob.JobUID]
		if !ok {
			continue
		}
		forceJobs = append(forceJobs, schedulerJob)
	}
	if len(forceJobs) == 0 {
		klog.V(util.LogInfoLev).Infof("GetNeedForceDeleteDelayingNPUJobs get nil jobs.")
		return nil, errors.New(getNoneJobsErr)
	}
	return forceJobs, nil
}

func (reScheduler *ReScheduler) isDelayingJobTimeout(fJob *FaultJob) bool {
	nowTime := time.Now().Unix()
	createTime := fJob.JobRankIdCreateTime
	klog.V(util.LogDebugLev).Infof("isDelayingJobTimeOut now: %v create: %v.", nowTime, createTime)
	if nowTime-createTime > reScheduler.GraceDeleteTime {
		klog.V(util.LogInfoLev).Infof("Time out: %v > %v", nowTime-createTime, reScheduler.GraceDeleteTime)
		return true
	}
	return false
}

// New Initialisation of ReScheduler
func New(env *plugin.ScheduleEnv, jobType string) *ReScheduler {
	klog.V(util.LogInfoLev).Infof("Creating Fault ReScheduler for %s ...", jobType)
	defer klog.V(util.LogInfoLev).Infof("Finished creating Fault ReScheduler for %s ...", jobType)
	var faultReScheduler = ReScheduler{
		GraceDeleteTime:      0,
		DealReSchedulerCache: nil,
	}
	// 1. Initialise ReScheduler.graceDeleteTime
	klog.V(util.LogDebugLev).Infof("Initialising graceDeleteTime.")
	graceDeleteTime, err := faultReScheduler.GetGraceDeleteTime(env.FrameAttr.Confs)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("GetGraceDeleteTime %s.", util.SafePrint(err))
	}
	faultReScheduler.setGraceOverTime(graceDeleteTime)
	// 2. Initialise ReScheduler.DealReSchedulerCache
	klog.V(util.LogDebugLev).Infof("Initialising ReSchedulerCache")
	reSchedulerCache := DealReSchedulerCache{
		FaultNodes: nil,
		FaultJobs:  nil,
	}
	// 2.1 Initialise ReScheduler.DealReSchedulerCache.Configmap
	klog.V(util.LogDebugLev).Infof("Initialising ReSchedulerCache.DealReSchedulerConfigmap")
	reSchedulerConfigmap := DealReSchedulerConfigmap{
		CMData: nil,
	}
	// 2.2 read configmap from kubernetes
	if cmErr := reSchedulerConfigmap.newReSchedulerCMFromEnv(env, jobType); cmErr != nil {
		klog.V(util.LogErrorLev).Infof("GetConfigmapFromKubernetes %s", util.SafePrint(cmErr))
		return nil
	}
	reSchedulerCache.DealReSchedulerConfigmap = &reSchedulerConfigmap
	// 2.2 Initialise ReScheduler.DealReSchedulerCache.FaultNodes by unmarshal data read from cm
	if setNodeErr := reSchedulerCache.SetFaultNodesFromCM(); setNodeErr != nil {
		klog.V(util.LogErrorLev).Infof("SetFaultNodesFromCM: %s", util.SafePrint(setNodeErr))
	}
	// 2.3 Initialise ReScheduler.DealReSchedulerCache.NodeHeartbeats by unmarshal data read from cm
	if setHBErr := reSchedulerCache.SetNodeHeartbeatFromCM(); setHBErr != nil {
		klog.V(util.LogErrorLev).Infof("SetNodeHeartbeatFromCM: %s", util.SafePrint(setHBErr))
	}

	if setRTErr := reSchedulerCache.SetRetryTimesFromCM(); setRTErr != nil {
		klog.V(util.LogErrorLev).Infof("SetRetryTimesFromCM: %s", util.SafePrint(setRTErr))
	}
	// 2.4 Initialise ReScheduler.DealReSchedulerCache.AllocNodeRankOccurrenceMap by unmarshal data read from cm
	if setNROErr := reSchedulerCache.SetNodeRankOccurrenceMapFromCM(); setNROErr != nil {
		klog.V(util.LogErrorLev).Infof("SetNodeRankOccurrenceMapFromCM: %s", util.SafePrint(setNROErr))
	}
	faultReScheduler.DealReSchedulerCache = &reSchedulerCache // 2.4 set DealReSchedulerCache
	faultReScheduler.Jobs = env.Jobs                          // 3 Initialise session Jobs Nodes copying data from env
	faultReScheduler.Nodes = env.Nodes
	faultReScheduler.kubeClient = env.FrameAttr.KubeClient // 4 Initialise kubeClient copying data from env
	return &faultReScheduler
}

// New910ReScheduler initialise ReScheduler.FaultJobs for 910x8
func (reScheduler *ReScheduler) New910ReScheduler() {
	klog.V(util.LogInfoLev).Infof("Initialising New 910x8 fault scheduler fault jobs")
	defer klog.V(util.LogInfoLev).Infof("Finished initialising New 910x8 fault scheduler fault jobs")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("New910ReScheduler: %s, nil reScheduler", util.ArgumentError)
		return
	}
	if setJobErr := reScheduler.DealReSchedulerCache.SetFaultJobsFromCM(CmFaultJob910x8Kind); setJobErr != nil {
		klog.V(util.LogErrorLev).Infof("SetFaultJobsFromCM: %s", util.SafePrint(setJobErr))
	}
	return
}

// NewCommonReScheduler initialise ReScheduler.FaultJobs for non 910x8
func (reScheduler *ReScheduler) NewCommonReScheduler(jobType string) {
	klog.V(util.LogInfoLev).Infof("Initialising New %s fault scheduler fault jobs", jobType)
	defer klog.V(util.LogInfoLev).Infof("Finished initialising New %s fault scheduler fault jobs", jobType)
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("NewCommonReScheduler: %s, nil reScheduler", util.ArgumentError)
		return
	}
	if setJobErr := reScheduler.DealReSchedulerCache.SetFaultJobsFromCM(jobType); setJobErr != nil {
		klog.V(util.LogErrorLev).Infof("SetFaultJobsFromCM: %s", util.SafePrint(setJobErr))
	}
	return
}

// SynCacheFaultNodeWithSession Synchronise FaultNodes in cache by updating the information using current session
func (reScheduler *ReScheduler) SynCacheFaultNodeWithSession() {
	klog.V(util.LogDebugLev).Infof("enter SynCacheFaultNodeWithSession ...")
	defer klog.V(util.LogDebugLev).Infof("leave SynCacheFaultNodeWithSession ...")
	klog.V(util.LogDebugLev).Infof("ReSchedulerCache fault nodes before sync: %#v", reScheduler.FaultNodes)
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("SynCacheFaultNodeWithSession: %s, nil reScheduler", util.ArgumentError)
		return
	}
	var updatedFaultNodes []FaultNode
	for _, faultNode := range reScheduler.FaultNodes {
		klog.V(util.LogDebugLev).Infof("Updating fault node %s recorded cache", faultNode.NodeName)
		// 1. nodes not in session should be kept in cache
		if !faultNode.isNodeInSessionByNpuNodes(reScheduler.Nodes) {
			klog.V(util.LogWarningLev).Infof("node %s in cache is not in session, keep without updating.",
				faultNode.NodeName)
			updatedFaultNodes = append(updatedFaultNodes, faultNode)
			continue
		}
		// 2. update attributes of cached FaultNodes utilising new information read from current session
		npuNode, _ := reScheduler.Nodes[faultNode.NodeName]
		// 2.1 read oldNodeHeartbeat value from cached nodeHeartbeat objects
		faultNode.setOldNodeHeartbeatTime(reScheduler.getLastNodeHeartbeatByNodeNameFromCache(npuNode.Name))
		// 2.2 update information sent by session NPUNodes
		faultNode.updateFaultNodesFromDeviceInfo(&npuNode, faultNode.NPUName)
		if err := faultNode.updateFaultNodesAttr(&npuNode); err != nil {
			klog.V(util.LogDebugLev).Infof("updateFaultNodesAttr: %s", util.SafePrint(err))
		}
		updatedFaultNodes = append(updatedFaultNodes, faultNode)
	}
	reScheduler.setFaultNodes(updatedFaultNodes)
	klog.V(util.LogDebugLev).Infof("ReSchedulerCache fault nodes after sync: %#v", reScheduler.FaultNodes)
}

// SynCacheFaultJobWithSession Synchronise FaultJobs in cache by updating the information using current session
func (reScheduler *ReScheduler) SynCacheFaultJobWithSession(ssn *framework.Session) {
	klog.V(util.LogInfoLev).Infof("enter SynCacheFaultJobWithSession...")
	defer klog.V(util.LogInfoLev).Infof("leave SynCacheFaultJobWithSession...")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("SynCacheFaultJobWithSession: %s, nil reScheduler", util.ArgumentError)
		return
	}
	var updatedFaultJobs []FaultJob
	nowTime := time.Now().Unix()
	for _, faultJob := range reScheduler.FaultJobs {
		// 1. cache Jobs exceeded max waiting time should be deleted and treated as normal new jobs
		if nowTime-faultJob.JobRankIdCreateTime > maxIntervalTime+reScheduler.GraceDeleteTime {
			klog.V(util.LogWarningLev).Infof("delete %s from CM for overTime %v => %v.",
				faultJob.JobName, nowTime, faultJob.JobRankIdCreateTime)
			continue
		}

		jobInfo := faultJob.jobInfoInSession(ssn.Jobs)
		if jobInfo == nil {
			klog.V(util.LogWarningLev).Infof("faultJob name: %s not in session", faultJob.JobName)
			continue
		}
		// 2. cache Jobs turned normal in session should be deleted ,meaning it has been restarted
		if faultJob.isJobGraceDeleteSuccess(jobInfo) {
			klog.V(util.LogDebugLev).Infof("%s grace deleted successful.", faultJob.JobName)
			// delete cache when all pods have been allocated
			allocated := int32(0)
			for _, task := range jobInfo.Tasks {
				if task.NodeName != "" {
					allocated++
				}
			}
			if allocated >= jobInfo.MinAvailable {
				continue
			}
		}
		if faultJob.ElasticScheduling == JobOnElasticScheduling {
			continue
		}
		if !faultJob.DeleteExecutedFlag {
			reScheduler.updateJobHealthCode(&faultJob)
		}
		str, err := json.Marshal(reScheduler.AllocNodeRankOccurrenceMap[faultJob.JobUID])
		if err != nil {
			klog.V(util.LogInfoLev).Infof("Marshal %s NodeRankOccurrence failed %s", faultJob.JobName, err)
		}
		if jobInfo.PodGroup.Annotations == nil {
			jobInfo.PodGroup.Annotations = make(map[string]string)
		}
		jobInfo.PodGroup.Annotations[plugin.JobDeleteFlag] = string(str)
		updatedFaultJobs = append(updatedFaultJobs, faultJob)
	}
	reScheduler.setFaultJobs(updatedFaultJobs)
	klog.V(util.LogDebugLev).Infof("ReSchedulerCache fault jobs after sync: %#v", reScheduler.FaultJobs)
}

// SyncJobRemainRetryTimes Synchronise job remain retry times in cache by updating the information using current session
func (reScheduler *ReScheduler) SyncJobRemainRetryTimes(ssn *framework.Session) {
	klog.V(util.LogInfoLev).Infof("enter SynJobRemainRetryTimes...")
	defer klog.V(util.LogInfoLev).Infof("leave SynJobRemainRetryTimes...")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("SynCacheFaultJobWithSession: %s, nil reScheduler", util.ArgumentError)
		return
	}

	klog.V(util.LogDebugLev).Infof("job remain retry times, sync before: %v", reScheduler.JobRemainRetryTimes)
	defer klog.V(util.LogDebugLev).Infof("job remain retry times, sync after: %v", reScheduler.JobRemainRetryTimes)

	newInfo := make(map[api.JobID]*RemainRetryTimes)
	for jobID, rt := range reScheduler.JobRemainRetryTimes {
		job, ok := ssn.Jobs[jobID]
		if !ok {
			klog.V(util.LogWarningLev).Infof("job<%s> is not session, remain retry times will be delete", jobID)
			continue
		}

		elastic, ok := job.PodGroup.Labels[ElasticSchedulingKey]
		if ok && elastic == JobOnElasticScheduling {
			continue
		}

		if util.UuidOfJob(job) != rt.UUID {
			continue
		}

		newInfo[jobID] = rt
	}
	reScheduler.JobRemainRetryTimes = newInfo
}

// SynCacheNodeRankOccMapWithSession Synchronise FaultJobs in cache by updating the information using current session
func (reScheduler *ReScheduler) SynCacheNodeRankOccMapWithSession(ssn *framework.Session) {
	klog.V(util.LogInfoLev).Infof("enter SynCacheNodeRankOccMapWithSession ...")
	defer klog.V(util.LogInfoLev).Infof("leave SynCacheNodeRankOccMapWithSession ...")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("SynCacheNodeRankOccMapWithSession: %s, nil reScheduler",
			util.ArgumentError)
		return
	}
	klog.V(util.LogDebugLev).Infof("NodeRankOccMap before sync: %#v", reScheduler.AllocNodeRankOccurrenceMap)
	newNodeRankOccMap := make(map[api.JobID][]*AllocNodeRankOccurrence, util.MapInitNum)
	for jobUID, NodeRankOcc := range reScheduler.AllocNodeRankOccurrenceMap {
		for _, fJob := range reScheduler.FaultJobs {
			if jobUID != fJob.JobUID {
				continue
			}
			if !fJob.checkJobNodeRankIndexValid() {
				newNodeRankOccMap[jobUID] = NodeRankOcc // restarted, leave the old map
			}
			ssnJob, ok := ssn.Jobs[fJob.JobUID]
			if !ok {
				newNodeRankOccMap[jobUID] = NodeRankOcc
			}
			if !fJob.IsFaultJob && plugin.IsJobRestarted(ssnJob) { // delete none faultJobs
				continue
			}
			newNodeRankOccMap[jobUID] = NodeRankOcc // only add faultJobs in the re-scheduling process
		}
	}
	reScheduler.AllocNodeRankOccurrenceMap = newNodeRankOccMap
	klog.V(util.LogDebugLev).Infof("NodeRankOccMap after sync: %#v", reScheduler.AllocNodeRankOccurrenceMap)
}

// InitFaultNodeMap init the node map of fault node
func (reScheduler *ReScheduler) InitFaultNodeMap() {
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("InitFaultNodeMap: %s, nil reScheduler", util.ArgumentError)
		return
	}
	reScheduler.FaultNodeMaps = make(map[string]SimpleFNodeInfo, len(reScheduler.FaultNodes))
	for _, node := range reScheduler.FaultNodes {
		reScheduler.FaultNodeMaps[node.NodeName] = initSimpleFNodeInfoByFNode(&node)
	}
}

// AddFaultNodeWithSession Add FaultNode objects for new nodes in session not in cache
func (reScheduler *ReScheduler) AddFaultNodeWithSession() {
	klog.V(util.LogInfoLev).Infof("enter AddFaultNodeWithSession ...")
	defer klog.V(util.LogInfoLev).Infof("leave AddFaultNodeWithSession ...")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("AddFaultNodeWithSession: %s, nil reScheduler", util.ArgumentError)
		return
	}
	newNodes := make(map[string]plugin.NPUNode, util.MapInitNum)
	nowTime := time.Now().Unix()
	for npuNodeName, npuNode := range reScheduler.Nodes {
		flag := false
		for _, fNode := range reScheduler.FaultNodes {
			if npuNodeName == fNode.NodeName {
				flag = true
				break
			}
		}
		if flag {
			klog.V(util.LogDebugLev).Infof("node %s is already in session, skip adding", npuNodeName)
			continue // 1. skip nodes already in cached FaultNodes
		}
		newNodes[npuNodeName] = npuNode // create new for those not in cache
	}
	for name, npuNode := range newNodes {
		klog.V(util.LogDebugLev).Infof("Adding node %s to reScheduler cache", name)
		chipKind, nameErr := npuNode.GetChipKindFromNpuNode()
		if nameErr != nil {
			klog.V(util.LogWarningLev).Infof("get chip name err by err:%s", chipKind)
			continue
		}
		npuName := util.HwPreName + chipKind
		// 0. Initialise faultNode
		faultNode := newFaultNodeDefault(npuNode.Name, nowTime)
		faultNode.NPUName = npuName
		faultNode.OldHeartbeatTime = reScheduler.getLastNodeHeartbeatByNodeNameFromCache(npuNode.Name)
		faultNode.UpdateHeartbeatTime = reScheduler.getLastNodeHeartUpdateTimeByNodeNameFromCache(npuNode.Name)
		faultNode.updateFaultNodesFromDeviceInfo(&npuNode, npuName)
		if err := faultNode.updateFaultNodesAttr(&npuNode); err != nil {
			klog.V(util.LogInfoLev).Infof("node %s updateFaultNodesAttr err: %s", npuNode.Name, util.SafePrint(err))
		}
		reScheduler.FaultNodes = append(reScheduler.FaultNodes, faultNode)
	}
	reScheduler.RealFaultNodes = reScheduler.GetRealFaultNodes()
}

// RestartNeedForceDeleteJobs Restart jobs that need to be force deleted
func (reScheduler *ReScheduler) RestartNeedForceDeleteJobs(ssn *framework.Session) error {
	klog.V(util.LogInfoLev).Infof("enter RestartNeedForceDeleteJobs...")
	defer klog.V(util.LogInfoLev).Infof("leave RestartNeedForceDeleteJobs...")
	if reScheduler == nil || ssn == nil {
		klog.V(util.LogErrorLev).Infof("RestartNeedForceDeleteJobs failed: %s, nil reScheduler or session",
			util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	needDeleteNPUJobs, err := reScheduler.GetNeedForceDeleteDelayingNPUJobs(reScheduler.Jobs, ssn)
	if err != nil {
		if err.Error() == getNoneJobsErr {
			return nil
		}
		return err
	}
	klog.V(util.LogDebugLev).Infof("GetNeedForceDeleteDelayingNPUJobs: %#v", needDeleteNPUJobs)
	for _, schedulerJob := range needDeleteNPUJobs {
		for _, faultJob := range reScheduler.FaultJobs {
			if schedulerJob.Name != faultJob.JobUID {
				continue
			}
			klog.V(util.LogWarningLev).Infof("grace delete job %s is timeout,force delete.", schedulerJob.Name)
			if deleteErr := faultJob.ForceDeleteJob(ssn, &schedulerJob); deleteErr != nil {
				klog.V(util.LogErrorLev).Infof("%s ForceDeleteJob: %s", schedulerJob.Name, util.SafePrint(deleteErr))
			}
		}
	}
	return nil
}

// RestartFaultJobs Restart fault jobs by its corresponding strategy  grace,force,off
func (reScheduler *ReScheduler) RestartFaultJobs(ssn *framework.Session) error {
	klog.V(util.LogInfoLev).Infof("enter RestartFaultJobs...")
	defer klog.V(util.LogInfoLev).Infof("leave RestartFaultJobs...")
	if reScheduler == nil || ssn == nil {
		klog.V(util.LogErrorLev).Infof("RestartFaultJobs failed: %s, nil reScheduler or nil session",
			util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	// 1. Get fault jobs, only faultJobs that haven't been evicted yet should be put into list
	realFaultJobs, err := reScheduler.getRealFaultJobs()
	if err != nil {
		if err.Error() == NoFaultJobsErr {
			return nil
		}
		return fmt.Errorf("restartFaultJobs: %s", util.SafePrint(err))
	}

	restartFaultJobs := reScheduler.getJobsToBeRestarted(realFaultJobs) // each job only triggers restart once
	var newCacheJobs []FaultJob
	var flag bool
	for _, fJob := range reScheduler.FaultJobs {
		flag = false
		for _, restartFJob := range restartFaultJobs {
			if fJob.JobName == restartFJob.JobName {
				flag = true
				break
			}
		}
		if !flag {
			newCacheJobs = append(newCacheJobs, fJob) // jobs no need to restart directly put back to cache FaultJobs
		}
	}

	klog.V(util.LogDebugLev).Infof("Jobs to be restarted: %#v", restartFaultJobs)
	// 2. Restart fault jobs
	for _, restartFaultJob := range restartFaultJobs {
		schedulerJob, ok := reScheduler.Jobs[restartFaultJob.JobUID]
		if !ok {
			klog.V(util.LogWarningLev).Infof("restartFaultJob %s not in session, has already been deleted",
				schedulerJob.Name)
			continue
		}
		klog.V(util.LogInfoLev).Infof("%s need restart.", restartFaultJob.JobName)
		if restartErr := restartFaultJob.restartSingleFaultJob(
			ssn, reScheduler, &schedulerJob); restartErr != nil {
			klog.V(util.LogErrorLev).Infof("RestartJob %s %s.", schedulerJob.Name, util.SafePrint(restartErr))
		} else {
			restartFaultJob.DeleteExecutedFlag = true
			if restartFaultJob.faultReason == PodFailed {
				reScheduler.JobRemainRetryTimes[restartFaultJob.JobUID].Times -= 1
				klog.V(util.LogInfoLev).Infof("job<%s> restart success, remain retry times reduce 1", restartFaultJob.JobUID)
			}
			klog.V(util.LogWarningLev).Infof("RestartJob %s execution success, set flag true", schedulerJob.Name)
		}
		newCacheJobs = append(newCacheJobs, restartFaultJob) // modify restartFlag and put modified fJob into cache
	}
	reScheduler.setFaultJobs(newCacheJobs)
	return nil
}

// ScoreBestNPUNodes add scores on scoreMap for normal nodes used by re-scheduling tasks
func (reScheduler *ReScheduler) ScoreBestNPUNodes(task *api.TaskInfo, scoreMap map[string]float64) error {
	if reScheduler == nil || task == nil || len(scoreMap) == 0 {
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes: %s, nil reScheduler or task or scoreMap",
			util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	klog.V(util.LogDebugLev).Infof("enter rescheduling ScoreBestNPUNodes %s...", task.Name)
	klog.V(util.LogDebugLev).Infof("node score map before add rescheduling weights %#v", scoreMap)
	defer klog.V(util.LogDebugLev).Infof("leave rescheduling ScoreBestNPUNodes ...")
	fJob := reScheduler.getFaultJobOfGivenTaskInfoFromCache(task) // 2. get faultJob object given the faultTask object
	if fJob == nil {
		klog.V(util.LogInfoLev).Infof("task %s is not in rescheduler cache", task.Name)
		return nil
	}
	if !fJob.IsFaultJob { // skip adding re-scheduling score for normal jobs
		return fmt.Errorf("task %s belongs to job %s which is not a fault job", task.Name, fJob.JobName)
	}
	if k, ok := task.Pod.Labels[AcJobTag]; !ok || k != AcJobVersion {
		reScheduler.scoreVcJobNPUNodes(fJob, scoreMap)
		return nil
	}

	reScheduler.scoreAcJobNPUNodes(task, fJob, scoreMap)
	klog.V(util.LogDebugLev).Infof("node score map after add rescheduling weights %#v", scoreMap)
	return nil
}

func (reScheduler *ReScheduler) scoreVcJobNPUNodes(fJob *FaultJob, scoreMap map[string]float64) {
	for _, ftask := range fJob.FaultTasks {
		if _, ok := scoreMap[ftask.NodeName]; ok {
			klog.V(util.LogDebugLev).Infof("node<%s> score is add", ftask.NodeName)
			scoreMap[ftask.NodeName] += util.NPUIndex8 * util.NPUIndex8
		}
	}
}

func (reScheduler *ReScheduler) scoreAcJobNPUNodes(task *api.TaskInfo, fJob *FaultJob, scoreMap map[string]float64) {
	// score for ascend job
	curfTask := reScheduler.getFaultTaskOfGivenTaskNameFromCache(task.Namespace, task.Name) // 1. get faultTask object
	if curfTask == nil {
		klog.V(util.LogInfoLev).Infof("task %s is not in rescheduler cache", task.Name)
		return
	}

	if _, ok := scoreMap[curfTask.NodeName]; ok {
		klog.V(util.LogDebugLev).Infof("fault task<%s> previous used node<%s> score is increase", task.Name,
			curfTask.NodeName)
		scoreMap[curfTask.NodeName] += util.NPUIndex8 * util.NPUIndex8
		return
	}

	for _, nodeName := range fJob.NodeNames {
		if _, ok := scoreMap[nodeName]; ok {
			scoreMap[nodeName] = 0
		}
	}
}

// UseAnnotation reallocate rankIndex for the re-scheduling jobs
func (reScheduler *ReScheduler) UseAnnotation(task *api.TaskInfo, node *plugin.NPUNode) error {
	if reScheduler == nil || task == nil || node == nil {
		klog.V(util.LogErrorLev).Infof("UseAnnotation failed: %s, nil reScheduler or task or node",
			util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	klog.V(util.LogInfoLev).Infof("enter rescheduling UseAnnotation ...(%s, %s)", task.Name, node.Name)
	defer klog.V(util.LogInfoLev).Infof("leave rescheduling UseAnnotation ...(%s, %s)", task.Name, node.Name)
	if reScheduler.AllocNodeRankOccurrenceMap == nil || len(reScheduler.AllocNodeRankOccurrenceMap) == 0 {
		return nil
	}

	fJob := reScheduler.getFaultJobOfGivenTaskInfoFromCache(task)
	if fJob == nil {
		return fmt.Errorf("no fJob %s in reScheduler cache", task.Job)
	}

	if fJob.ElasticScheduling == JobOnElasticScheduling {
		klog.V(util.LogInfoLev).Infof("task %s is enabled with elastic scheduling, "+
			"skip volcano rankIndex writing process", task.Name)
		return nil
	}

	// if job is ascend job,skip add rankIndex
	if k, ok := task.Pod.Labels[AcJobTag]; ok && k == AcJobVersion {
		return nil
	}

	nodeRankTimes := reScheduler.AllocNodeRankOccurrenceMap[fJob.JobUID]
	// 1. if given node is in the nodeRankTime, keep it ,node is used by the fault job before
	for _, nodeRankTime := range nodeRankTimes {
		klog.V(util.LogInfoLev).Infof("set before: node: %s, rank: %s, occur: %d", nodeRankTime.NodeName,
			nodeRankTime.RankIndex, nodeRankTime.Occurrence)
		if node.Name == nodeRankTime.NodeName {
			klog.V(util.LogInfoLev).Infof("set old node rank index <%s>/<%s>/<%s>", node.Name,
				task.Name, nodeRankTime.RankIndex)
			task.Pod.Annotations[podRankIndex] = nodeRankTime.RankIndex
			nodeRankTime.Occurrence++
			return nil
		}
	}
	for _, nodeRankTime := range nodeRankTimes {
		if nodeRankTime.Occurrence == 0 {
			klog.V(util.LogInfoLev).Infof("set new node rank index <%s>/<%s>/<%s>", node.Name,
				task.Name, nodeRankTime.RankIndex)
			nodeRankTime.Occurrence++
			nodeRankTime.NodeName = node.Name
			task.Pod.Annotations[podRankIndex] = nodeRankTime.RankIndex
			break
		}
	}

	return nil
}

// GenerateNodeRankIndexTaskMap get the nodeName, rankIndex, and Occurrence of nodes in a job
func (reScheduler *ReScheduler) GenerateNodeRankIndexTaskMap() {
	klog.V(util.LogInfoLev).Infof("enter GenerateNodeRankIndexTaskMap ...")
	defer klog.V(util.LogInfoLev).Infof("leave GenerateNodeRankIndexTaskMap ...")
	if reScheduler == nil {
		klog.V(util.LogErrorLev).Infof("GenerateNodeRankIndexTaskMap failed: %s, nil reScheduler",
			util.ArgumentError)
		return
	}
	klog.V(util.LogDebugLev).Infof("NodeRankOccMap before add: %#v", reScheduler.AllocNodeRankOccurrenceMap)
	nodeRankIndexTaskMap := make(map[api.JobID][]*AllocNodeRankOccurrence, util.MapInitNum)
	for _, fJob := range reScheduler.FaultJobs {
		oldRecord, ok := reScheduler.AllocNodeRankOccurrenceMap[fJob.JobUID]
		if ok {
			klog.V(util.LogDebugLev).Infof("NodeRankOccMap for job %s already generated, keep it", fJob.JobName)
			nodeRankIndexTaskMap[fJob.JobUID] = oldRecord
			// continue but do not change those whose jobid already occurred
			continue
		}
		if fJob.DeleteExecutedFlag {
			klog.V(util.LogDebugLev).Infof("Create NodeRankOccMap for job %s", fJob.JobName)
			var nodeRankTimes []*AllocNodeRankOccurrence
			for _, fTask := range fJob.FaultTasks {
				nodeRankTime := &AllocNodeRankOccurrence{
					NodeName:   fTask.NodeName,
					RankIndex:  fTask.NodeRankIndex,
					IsFault:    fTask.IsFaultTask,
					Occurrence: 0,
				}
				nodeRankTimes = append(nodeRankTimes, nodeRankTime)
			}
			nodeRankIndexTaskMap[fJob.JobUID] = nodeRankTimes
		}
	}
	reScheduler.AllocNodeRankOccurrenceMap = nodeRankIndexTaskMap
	klog.V(util.LogDebugLev).Infof("NodeRankOccMap after add: %#v", reScheduler.AllocNodeRankOccurrenceMap)
}

// CheckNodeNPUByTask used in the predicate process of task and node
func (reScheduler *ReScheduler) CheckNodeNPUByTask(task *api.TaskInfo, vcNode plugin.NPUNode, npuName string) error {
	klog.V(util.LogDebugLev).Infof("enter rescheduling CheckNodeNPUByTask ...(%s, %s)", task.Name, vcNode.Name)
	defer klog.V(util.LogDebugLev).Infof("leave rescheduling CheckNodeNPUByTask ...(%s, %s)",
		task.Name, vcNode.Name)

	// 3. non faultJobs should not occupy normal nodes previously used by distributional
	if err := reScheduler.checkNodeNewJobUseFJobNormNode(vcNode, task); err != nil {
		return err
	}
	// 1. jobs should not be scheduled to faultNodes
	if err := reScheduler.checkNodeCurNodeIsFault(vcNode, task); err != nil {
		return err
	}
	if curFTask := reScheduler.getFaultTaskOfGivenTaskNameFromCache(task.Namespace, task.Name); curFTask == nil {
		klog.V(util.LogDebugLev).Infof("task %s not in reschedule cache", task.Name)
		return nil // cannot return error, or will be stuck for new job scheduling
	}
	curFJob := reScheduler.getFaultJobOfGivenTaskInfoFromCache(task)
	if curFJob == nil {
		return fmt.Errorf("task %s does not have corresponding job in cache", task.Name)
	}
	if !curFJob.IsFaultJob || curFJob.ReScheduleKey == JobOffRescheduleLabelValue || npuName != util.NPU910CardName {
		klog.V(util.LogDebugLev).Infof("CheckNodeNPUByTask job %s is not fault job, node %s check over",
			curFJob.JobName, vcNode.Name)
		return nil
	}
	curSchedulerJob := reScheduler.getSchedulerJobOfGivenUIDFromReScheduler(task.Job)
	// 0. if fTask's corresponding faultJobs' previously used normal nodes haven't be released,
	// stuck the task scheduling process
	if err := reScheduler.checkNodeFJobNormNodeRelease(curFJob, curSchedulerJob); err != nil {
		return err
	}
	// 4. job's previously old node should only be assigned to new node once
	if err := reScheduler.checkFJobFNodeRankIndexAllAllocated(curFJob, vcNode); err != nil {
		return err
	}
	klog.V(util.LogDebugLev).Infof("CheckNodeNPUByTask node %s passed rescheduling predicate for task %s",
		vcNode.Name, task.Name)
	return nil
}

// 0. stuck scheduling as long as normal nodes used by re-scheduling jobs not released
func (reScheduler *ReScheduler) checkNodeFJobNormNodeRelease(
	curFJob *FaultJob, curSchedulerJob plugin.SchedulerJob) error {
	if !curFJob.IsFaultJob { // if current job is not in faultJob cache, skip
		klog.V(util.LogDebugLev).Infof("job %s is not a fault job", curFJob.JobName)
		return nil
	}

	if curFJob.isUnstable { // if current job is not stable, return err
		return fmt.Errorf("job used normal node is still unstable")
	}
	for _, fTask := range curFJob.FaultTasks {
		fNode := reScheduler.getFNodeByNodeName(fTask.NodeName)
		if fNode == nil { // if fNode is nil then the node should be a fault one
			klog.V(util.LogDebugLev).Infof("node %s does not exist in cache", fTask.NodeName)
			continue
		}
		if !fNode.IsFaultNode { // if normal node in faultJob hasn't been released, return error
			if !isNodeInSessionByNodeName(fNode.NodeName, reScheduler.Nodes) { // case1. node not released and not sent by ssn
				return fmt.Errorf("cache fault  normal node hasn't been release")
			}
			npuNode := reScheduler.getNPUNodeOfGiveNodeNameFromReScheduler(fNode.NodeName)
			if err := npuNode.CheckNPUResourceStableReScheduling(curSchedulerJob); err != nil {
				curFJob.isUnstable = true
				return fmt.Errorf("cache fault normal node resource still unstableby err<%s>", err) // case2.node sent by ssn but still unstable
			}
		}
	}
	klog.V(util.LogDebugLev).Infof("checkNodeFJobNormNodeRelease: check ok, fault job %s task length: %d",
		curFJob.JobName, len(curFJob.FaultTasks))
	return nil
}

func (reScheduler *ReScheduler) checkNodeCurNodeIsFault(vcNode plugin.NPUNode, task *api.TaskInfo) error {
	schedulerJob, ok := reScheduler.Jobs[task.Job]
	if !ok {
		return fmt.Errorf("task corresponding job not in session")
	}
	reschKey, ok := schedulerJob.SchedulerJobAttr.Label[JobRescheduleLabelKey]
	if !ok || reschKey == JobOffRescheduleLabelValue {
		klog.V(util.LogInfoLev).Infof("job %s rescheduling not enabled, skip check node", schedulerJob.Name)
		return nil
	}
	for _, fNode := range reScheduler.RealFaultNodes {
		if vcNode.Name == fNode.NodeName && fNode.NodeHealthState == NodeUnhealthy {
			// none distributed job, npu fault considered in previous ops
			return fmt.Errorf("task cannot be assigned to %s node", NodeUnhealthy)
		}
	}
	klog.V(util.LogInfoLev).Infof("node %s is not fault node, check success", vcNode.Name)
	return nil
}

// 2. new jobs cannot take normal nodes used by old distributional jobs
func (reScheduler *ReScheduler) checkNodeNewJobUseFJobNormNode(vcNode plugin.NPUNode, task *api.TaskInfo) error {
	if reScheduler == nil {
		return errors.New(util.ArgumentError)
	}
	realFaultJobs, err := reScheduler.getRealFaultJobs()
	if err != nil {
		klog.V(util.LogDebugLev).Infof("none real fault jobs")
		return nil
	}
	usedByFaultJob := false
	// 3. non faultJobs should not occupy normal nodes previously used by distributional
	// faultJobs in the re-scheduling process
	for _, fJob := range realFaultJobs {
		if _, ok := fJob.NodeNameMaps[vcNode.Name]; !fJob.IsFaultJob || !ok {
			continue
		}
		usedByFaultJob = true
		if task.Job == fJob.JobUID ||
			(task.Namespace == fJob.JobNamespace && util.ReferenceNameOfTask(task) == fJob.ReferenceName) {
			klog.V(util.LogInfoLev).Infof("node %s is not normal node used by fault job or current task %s is in "+
				"reScheduler job, check success", vcNode.Name, task.Name)
			return nil
		}
	}
	if !usedByFaultJob {
		return nil
	}
	klog.V(util.LogDebugLev).Infof("task %s cannot use normal node %s occupied by faultJob %v",
		task.Name, vcNode.Name, realFaultJobs)
	return fmt.Errorf("task cannot use node occupied by faultJob")

}

// checkFJobFNodeRankIndexAllAllocated ensure rankIndex of fNode not allocated is enough for nodes not occurred
func (reScheduler *ReScheduler) checkFJobFNodeRankIndexAllAllocated(curFJob *FaultJob, vcNode plugin.NPUNode) error {
	nodeRankTimes, ok := reScheduler.AllocNodeRankOccurrenceMap[curFJob.JobUID]
	if !ok {
		return fmt.Errorf("job %s's rankIndexMap not found in cache", curFJob.JobUID)
	}
	countFNode := 0
	countRankId := 0
	for _, nodeRankTime := range nodeRankTimes {
		if vcNode.Name == nodeRankTime.NodeName { // node already in old node cache
			return nil
		}
		fNode := reScheduler.getFNodeOfGivenNameFromCache(nodeRankTime.NodeName)
		if fNode == nil {
			return fmt.Errorf("node %s not found in cache", nodeRankTime.NodeName)
		}
		if fNode.IsFaultNode {
			countFNode++
		}
		if nodeRankTime.Occurrence == 1 {
			countRankId++
		}
	}

	if countFNode > 0 && countFNode == countRankId {
		return fmt.Errorf("node %s cannot be assigned to job %s since job's rank index has been assigned to "+
			"other nodes", vcNode.Name, curFJob.JobName)
	}
	klog.V(util.LogDebugLev).Infof("checkFJobFNodeRankIndexAllAllocated: check ok for node %s", vcNode.Name)
	return nil
}

func (reScheduler ReScheduler) getFaultTaskOfGivenTaskNameFromCache(namespace, name string) *FaultTask {
	for _, fJob := range reScheduler.FaultJobs {
		if fJob.JobNamespace != namespace {
			continue
		}
		for _, fTask := range fJob.FaultTasks {
			if fTask.TaskName == name {
				return &fTask
			}
		}
	}
	return nil
}

func (reScheduler ReScheduler) getNPUNodeOfGiveNodeNameFromReScheduler(nodeName string) *plugin.NPUNode {
	if len(reScheduler.Nodes) == 0 {
		return nil
	}
	if node, ok := reScheduler.Nodes[nodeName]; ok {
		return &node
	}
	return nil
}

func (reScheduler ReScheduler) getSchedulerJobOfGivenUIDFromReScheduler(jobUID api.JobID) plugin.SchedulerJob {
	return reScheduler.Jobs[jobUID]
}

// getFaultJobOfGivenTaskInfoFromCache get fault job from task info
func (reScheduler ReScheduler) getFaultJobOfGivenTaskInfoFromCache(task *api.TaskInfo) *FaultJob {
	for i, fJob := range reScheduler.FaultJobs {
		if fJob.JobUID == task.Job {
			return &reScheduler.FaultJobs[i]
		}
		if task.Namespace == fJob.JobNamespace && util.ReferenceNameOfTask(task) == fJob.ReferenceName {
			return &reScheduler.FaultJobs[i]
		}
	}
	return nil
}

func (reScheduler ReScheduler) getLastNodeHeartbeatByNodeNameFromCache(nodeName string) int64 {
	for _, nodeHB := range reScheduler.NodeHeartbeats {
		if nodeHB.NodeName == nodeName {
			klog.V(util.LogDebugLev).Infof("getLastNodeHeartbeatByNodeNameFromCache: %s, %d",
				nodeName, nodeHB.HeartbeatTime)
			return nodeHB.HeartbeatTime
		}
	}
	return 0
}

func (reScheduler ReScheduler) setTaskCardHealthCode(fTask *FaultTask) error {
	klog.V(util.LogDebugLev).Infof("task %s setTaskCardHealthCode", fTask.TaskName)
	var reasonList []FaultReasonList
	if fTask.NodeName == "" {
		fTask.Reason = reasonList
		return fmt.Errorf("setTaskCardHealthCode fTask %s use node is nil", fTask.TaskName)
	}
	for _, fNode := range reScheduler.FaultNodes {
		if fNode.NodeName != fTask.NodeName {
			continue
		}
		if fNode.NodeHealthState == NodeUnhealthy {
			var reason FaultReasonList
			reason.NodeName = fNode.NodeName
			reason.FaultType = NodeUnhealthy
			reason.FaultCode = NodeFaultCode
			reason.FaultLevel = PreSeparateNPU
			reason.FaultHandling = PreSeparateNPU
			reason.LargeModelFaultLevel = PreSeparateNPU
			reasonList = append(reasonList, reason)
		}
		tmpReason := setTaskFaultReasonByFaultNode(fTask, fNode)
		reasonList = append(reasonList, tmpReason...)
		break
	}
	fTask.Reason = reasonList
	return nil
}

func setTaskFaultReasonByFaultNode(fTask *FaultTask, fNode FaultNode) []FaultReasonList {
	var reasonList []FaultReasonList
	for _, cardName := range fTask.UseCardName {
		for _, fCard := range fNode.FaultDeviceList {
			if cardName == fCard.NPUName && fCard.FaultHandling != NotHandleFault {
				var reason FaultReasonList
				reason.NodeName = fNode.NodeName
				reason.FaultDeviceList = fCard
				reasonList = append(reasonList, reason)
			}
		}
	}
	return reasonList
}

func (reScheduler ReScheduler) updateJobHealthCode(fJob *FaultJob) {
	if fJob == nil {
		return
	}
	for index := range fJob.FaultTasks {
		if err := reScheduler.setTaskCardHealthCode(&fJob.FaultTasks[index]); err != nil {
			klog.V(util.LogInfoLev).Infof("setTaskCardHealthCode err:%s", err)
		}
	}
}

func (reScheduler ReScheduler) getLastNodeHeartUpdateTimeByNodeNameFromCache(nodeName string) int64 {
	for _, nodeHB := range reScheduler.NodeHeartbeats {
		if nodeHB.NodeName == nodeName {
			klog.V(util.LogDebugLev).Infof("getLastNodeHeartbeatByNodeNameFromCache: %s, %d",
				nodeName, nodeHB.HeartbeatTime)
			return nodeHB.UpdateTime
		}
	}
	return 0
}

// getTaskHealthState return true when unhealthy
func (reScheduler ReScheduler) getTaskHealthState(fTask *FaultTask, task *api.TaskInfo) (bool, string) {
	klog.V(util.LogDebugLev).Infof("task %s getTaskHealthState", fTask.TaskName)

	if fTask.NodeName == "" {
		return false, NodeHealthy // tasks has not yet been scheduled
	}
	isFault, state := reScheduler.getTaskHealthStateByNode(fTask)
	if isFault {
		return isFault, state
	}
	return reScheduler.getTaskHealthStateByPod(task)
}

func (reScheduler *ReScheduler) getTaskHealthStateByNode(fTask *FaultTask) (bool, string) {
	var nodeUseCardHealthState []string
	realFaultNode := reScheduler.GetRealFaultNodes()
	for _, fNode := range realFaultNode {
		if fNode.NodeName == fTask.NodeName {
			if !fNode.IsFaultNode { // if task used node isFaultNode is false, return healthy
				klog.V(util.LogInfoLev).Infof("task %s use healthy node %s, thus task sets %s", fTask.TaskName,
					fNode.NodeName, NodeHealthy)
				return false, NodeHealthy
			}
			if fNode.NodeHealthState == NodeUnhealthy { // if task used node is nodeUnhealthy, return
				klog.V(util.LogInfoLev).Infof("task %s use %s node %s, thus task sets %s", fTask.TaskName,
					NodeUnhealthy, fNode.NodeName, NodeUnhealthy)
				return true, NodeUnhealthy
			}
			nodeUseCardHealthState = fTask.getTaskUseFaultCardHealthState(&fNode) // get fault NPUs on task used node
		}
	}
	if util.IsSliceContain(NodeCardUnhealthy, nodeUseCardHealthState) { // if has unhealthy npu, return in advance
		klog.V(util.LogInfoLev).Infof("task %s use %s node, thus task sets %s", fTask.TaskName,
			NodeCardUnhealthy, NodeCardUnhealthy)
		return true, NodeCardUnhealthy
	}
	klog.V(util.LogInfoLev).Infof("task %s all nodes healthy, thus task sets %s", fTask.TaskName, NodeHealthy)
	return false, NodeHealthy
}

func (reScheduler *ReScheduler) getTaskHealthStateByPod(task *api.TaskInfo) (bool, string) {
	if task.Pod.Status.Phase == v1.PodFailed {
		return true, PodFailed
	}
	return false, PodHealthy
}

func (reScheduler ReScheduler) getJobsToBeRestarted(realFaultJobs []FaultJob) []FaultJob {
	var restartFaultJobs []FaultJob
	for _, fJob := range realFaultJobs {
		if fJob.DeleteExecutedFlag {
			continue
		}

		restartFaultJobs = append(restartFaultJobs, fJob)
	}
	return restartFaultJobs
}

func (reScheduler ReScheduler) getFNodeByNodeName(nodeName string) *SimpleFNodeInfo {
	if len(reScheduler.FaultNodeMaps) == 0 {
		return nil
	}
	if node, ok := reScheduler.FaultNodeMaps[nodeName]; ok {
		return &node
	}
	return nil
}

func (reScheduler ReScheduler) getFNodeOfGivenNameFromCache(nodeName string) *FaultNode {
	for _, fNode := range reScheduler.FaultNodes {
		if fNode.NodeName == nodeName {
			return &fNode
		}
	}
	return nil
}
