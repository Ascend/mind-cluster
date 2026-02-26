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
Package main is using for HuaWei Ascend pin affinity schedule.
*/
package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

var sHandler *plugin.ScheduleHandler

func init() {
	sHandler = HandlerStart()
}

// HandlerStart HuaWei NPU plugin start by frame.
func HandlerStart() *plugin.ScheduleHandler {
	scheduleHandler := &plugin.ScheduleHandler{
		NPUPlugins:  sets.String{util.NPU910CardName: {}, util.NPU310CardName: {}, util.NPU310PCardName: {}},
		FaultHandle: rescheduling.NewHandler(),
		ScheduleEnv: plugin.ScheduleEnv{
			FrameAttr:               plugin.NewVolcanoFrame(),
			JobScheduleInfoRecorder: plugin.NewJobScheduleInfoRecorder(),
			ClusterCache:            plugin.NewClusterCache(),
		},
	}
	scheduleHandler.PolicyBuilder = internal.New
	return scheduleHandler
}

// New return npu plugin.
func New(arguments framework.Arguments) framework.Plugin {
	return &huaweiNPUPlugin{Scheduler: sHandler, Arguments: arguments}
}

// Name This need by volcano frame init plugin.
func (tp *huaweiNPUPlugin) Name() string {
	return PluginName
}

// OnSessionOpen HuaWei NPU Action's init session for frame.
func (tp *huaweiNPUPlugin) OnSessionOpen(ssn *framework.Session) {
	klog.V(util.LogInfoLev).Infof("enter %s OnSessionOpen.", PluginName)
	defer klog.V(util.LogInfoLev).Infof("leave %s OnSessionOpen.", PluginName)
	if tp == nil || ssn == nil {
		klog.V(util.LogInfoLev).Infof("OnSessionOpen : %s.", util.ArgumentError)
		return
	}
	// Init npu plugin and nodes.
	if err := tp.Scheduler.InitNPUSession(ssn); err != nil {
		klog.V(util.LogErrorLev).Infof("InitNPUSession : %s, npu plugin will not be initialized.", err)
		return
	}
	// check job npu resource, if illegal return failed
	ssn.AddJobValidFn(tp.Name(), func(obj interface{}) *api.ValidateResult {
		return tp.Scheduler.JobValid(obj)
	})
	// if node not meet the task require, the task will be failed. so need to intercept in advance
	addPredicateFn(ssn, tp)

	ssn.AddJobPipelinedFn(tp.Name(), func(obj interface{}) int {
		return jobPipelined(obj, tp)
	})

	ssn.AddJobOrderFn(tp.Name(), func(l interface{}, r interface{}) int {
		return jobOrderFn(l, r)
	})

	addBatchNodeOrderFn(ssn, tp)

	ssn.AddJobReadyFn(tp.Name(), func(obj interface{}) bool {
		return jobReady(obj, tp)
	})

	ssn.AddJobEnqueueableFn(tp.Name(), func(job interface{}) int {
		return jobEnqueueable(job, ssn, tp)
	})

	ssn.AddTaskOrderFn(tp.Name(), func(l interface{}, r interface{}) int {
		return tp.Scheduler.TaskOrderFn(l, r)
	})
	// Register event handlers to update task info in PodLister & nodeMap
	// for support Concurrency
	addEventHandler(ssn, tp)

	updatePgAnnotation(ssn)
}

// OnSessionClose Close session by volcano frame.
func (tp *huaweiNPUPlugin) OnSessionClose(ssn *framework.Session) {
	klog.V(util.LogInfoLev).Infof("enter %s OnSessionClose.", PluginName)
	defer klog.V(util.LogInfoLev).Infof("leave %s OnSessionClose.", PluginName)
	if tp == nil || ssn == nil {
		klog.V(util.LogInfoLev).Infof("OnSessionClose failed: %s.", util.ArgumentError)
		return
	}
	if *tp.Scheduler.FrameAttr.IsFirstSession {
		*tp.Scheduler.FrameAttr.IsFirstSession = false
	}
	// 1、Record job's unscheduled reason;
	// 2、Update job statue;
	// 3、Handle other post-dispatch issues.
	for _, job := range ssn.Jobs {
		sjob, ok := tp.Scheduler.Jobs[job.UID]
		if !ok {
			continue
		}
		klog.V(util.LogInfoLev).Infof("job ReadyTaskNum %d, sjob.MinAvailable: %d", job.ReadyTaskNum(),
			sjob.MinAvailable)
		if job.ReadyTaskNum() >= sjob.MinAvailable {
			continue
		}
		tp.addBatchOrderFailedCondition(job, ssn)
		tp.addNodePredicateFailedCondition(job, ssn)
		tp.addJobValidFailedCondition(job, ssn)
		tp.addJobEnqueueFailedCondition(job, ssn)
	}
	tp.Scheduler.BeforeCloseHandler()
}

func addPredicateFn(ssn *framework.Session, tp *huaweiNPUPlugin) {
	// check job npu resource, if illegal return failed
	ssn.AddPredicateFn(tp.Name(), func(taskInfo *api.TaskInfo, nodeInfo *api.NodeInfo) error {
		predicateErr := tp.Scheduler.NodePredicate(taskInfo, nodeInfo)
		if predicateErr != nil {
			tp.Scheduler.NodePredicateErrors.Add(taskInfo.Job, nodeInfo.Name, predicateErr)
		}
		return predicateErr
	})
}

func jobPipelined(obj interface{}, tp *huaweiNPUPlugin) int {
	ji, ok := obj.(*api.JobInfo)
	if !ok {
		klog.V(util.LogErrorLev).Info("obj assertion failed.")
		return util.Reject
	}

	job, ok := tp.Scheduler.Jobs[ji.UID]
	if !ok {
		return util.Abstain
	}
	if *job.JobReadyTag {
		return util.Abstain
	}
	return util.Reject
}

func addBatchNodeOrderFn(ssn *framework.Session, tp *huaweiNPUPlugin) {
	ssn.AddBatchNodeOrderFn(tp.Name(), func(task *api.TaskInfo, nodes []*api.NodeInfo) (map[string]float64, error) {
		_, ok := tp.Scheduler.PredicatedNodes[task.Job]
		if !ok {
			tp.Scheduler.PredicatedNodes[task.Job] = sets.String{}
		}
		for _, node := range nodes {
			tp.Scheduler.PredicatedNodes[task.Job].Insert(node.Name)
		}
		score, err := tp.Scheduler.BatchNodeOrderFn(task, nodes)
		if err != nil {
			tp.Scheduler.BatchOrderError[task.Job] = err
		}
		return score, nil
	})
}

func jobReady(obj interface{}, tp *huaweiNPUPlugin) bool {
	ji, ok := obj.(*api.JobInfo)
	if !ok {
		klog.V(util.LogErrorLev).Info("obj assertion failed.")
		return false
	}
	job, ok := tp.Scheduler.Jobs[ji.UID]
	if !ok {
		return true
	}
	return *job.JobReadyTag && ji.ReadyTaskNum() >= job.MinAvailable
}

func addEventHandler(ssn *framework.Session, tp *huaweiNPUPlugin) {
	ssn.AddEventHandler(&framework.EventHandler{
		AllocateFunc: func(event *framework.Event) {
			if event == nil {
				klog.V(util.LogErrorLev).Infof("AllocateFunc event nil.")
				return
			}
			tp.Scheduler.NPUAllocateFunc(event.Task)
		},
		DeallocateFunc: func(event *framework.Event) {
			if event == nil {
				klog.V(util.LogErrorLev).Infof("DeallocateFunc event nil.")
				return
			}
			tp.Scheduler.NPUDeallocateFunc(event.Task)
		},
	})
}

func jobEnqueueable(job interface{}, ssn *framework.Session, tp *huaweiNPUPlugin) int {
	if tp.Scheduler.NPUPlugins == nil {
		klog.V(util.LogErrorLev).Infof("AddJobEnqueueableFn : %s", util.ArgumentError)
		return util.JobEnqueueSkip
	}
	vcjob, ok := job.(*api.JobInfo)
	if !ok {
		return util.JobEnqueueSkip
	}
	jobDequeueForTimeout(vcjob, ssn)
	npuName, rNpuNum, _ := plugin.GetVCJobReqNPUTypeFromJobInfo(vcjob)
	if !tp.Scheduler.NPUPlugins.Has(npuName) {
		return util.JobEnqueueSkip
	}
	tNpuNum := getNpuNum(ssn, tp, npuName)
	if tNpuNum < rNpuNum {
		klog.V(util.LogWarningLev).Infof("job <%s> Add enqueue failed, require npu num is %v "+
			"but cluster npu num is %v", vcjob.Name, rNpuNum, tNpuNum)
		tp.Scheduler.EnqueueError[vcjob.UID] = fmt.Errorf("require npu num is %v, but cluster npu num is %v", rNpuNum,
			tNpuNum)
		return util.JobNotEnqueue
	}
	if tp.Scheduler.FrameAttr.ForceEnqueue {
		klog.V(util.LogWarningLev).Infof("job <%s> Add enqueue success will start schedule, require npu num is <%v> "+
			"and cluster npu num is <%v>.", vcjob.Name, rNpuNum, tNpuNum)
		return util.JobEnqueue
	}
	return util.JobEnqueueSkip
}

func getNpuNum(ssn *framework.Session, tp *huaweiNPUPlugin, npuName string) int {
	var tNpuNum int
	errs := util.NewErrorCollector("getNpuNum", util.DefaultPrintLimit)
	for _, node := range ssn.Nodes {
		vcNode, ok := tp.Scheduler.Nodes[node.Name]
		if !ok {
			klog.V(util.LogDebugLev).Infof("AddJobEnqueueableFn add node failed,%s is not in cache", node.Name)
			errs.Add(node.Name, errors.New("node is not in cache"))
			continue
		}
		deviceInfo, ok := vcNode.Annotation[npuName]
		if !ok || len(deviceInfo) == 0 {
			klog.V(util.LogDebugLev).Infof("AddJobEnqueueableFn add node failed,"+
				"%s deviceList is empty", node.Name)
			errs.Add(node.Name, errors.New("node deviceList is empty"))
			continue
		}
		deviceList := strings.Split(deviceInfo, ",")
		klog.V(util.LogDebugLev).Infof("Add enqueue node %s deviceList is: %#v", vcNode.Name, deviceList)
		npuNum, ok := vcNode.Idle[v1.ResourceName(npuName)]
		shareDevCount := 1
		softShareDevEnableExist := false
		if node.Node != nil {
			softShareDevEnable, softShareDevEnableExist := node.Node.Labels[util.SchedulerSoftShareDevEnableNodeLabel]
			if softShareDevEnableExist && softShareDevEnable == "true" {
				shareDevCount = util.SoftShareDevCount
			}
		}
		if !softShareDevEnableExist && (!ok || len(deviceList) > int(npuNum/util.NPUHexKilo)) {
			klog.V(util.LogDebugLev).Infof("Add enqueue node %s device info is %v and k8s is %v", vcNode.Name,
				len(deviceList), int(npuNum/util.NPUHexKilo))
			errs.Add(node.Name, fmt.Errorf("node resource is not stable, device info is %v and k8s is %v",
				len(deviceList), int(npuNum/util.NPUHexKilo)))
			continue
		}
		if capVal, exist := vcNode.Capability[v1.ResourceName(npuName)]; !exist || capVal < npuNum {
			klog.V(util.LogErrorLev).Infof("Add enqueue node %s cap<%v> is less than idle<%v>, waiting "+
				"kubelet report correctly", vcNode.Name, int(capVal/util.NPUHexKilo), int(npuNum/util.NPUHexKilo))
			errs.Add(node.Name, fmt.Errorf("node resource is not init, cap<%v> is less than idle<%v>",
				int(capVal/util.NPUHexKilo), int(npuNum/util.NPUHexKilo)))
			continue
		}
		tNpuNum += len(deviceList) * shareDevCount
	}
	errs.Print()
	return tNpuNum
}

func jobOrderFn(interfaceA interface{}, interfaceB interface{}) int {
	jobInfoA, ok := interfaceA.(*api.JobInfo)
	if !ok {
		klog.V(util.LogDebugLev).Infof("jobOrderFn failed, object is not JobInfo")
		return util.JobOrderSamePriority
	}
	jobInfoB, ok := interfaceB.(*api.JobInfo)
	if !ok {
		klog.V(util.LogDebugLev).Infof("jobOrderFn failed, object is not JobInfo")
		return util.JobOrderSamePriority
	}
	var lNum, rNum = 0, 0
	var err error = nil
	lStrNum, lExist := jobInfoA.PodGroup.Annotations[util.DequeueFrequencyAnnoKey]
	if lExist && lStrNum != "" {
		lNum, err = strconv.Atoi(lStrNum)
		if err != nil {
			klog.V(util.LogDebugLev).Infof("jobOrderFn failed, convert dequeue frequency failed, "+
				"strNum: %s, err: %v", lStrNum, err)
			return util.JobOrderSamePriority
		}
	}
	rStrNum, rExist := jobInfoB.PodGroup.Annotations[util.DequeueFrequencyAnnoKey]
	if rExist && rStrNum != "" {
		rNum, err = strconv.Atoi(rStrNum)
		if err != nil {
			klog.V(util.LogDebugLev).Infof("jobOrderFn failed, convert dequeue frequency failed, "+
				"strNum: %s, err: %v", rStrNum, err)
			return util.JobOrderSamePriority
		}
	}
	if lNum > rNum {
		return util.JobOrderLowPriority
	} else if lNum < rNum {
		return util.JobOrderHighPriority
	} else {
		return util.JobOrderSamePriority
	}
}

func updatePgAnnotation(ssn *framework.Session) {
	for _, jobInfo := range ssn.Jobs {
		if jobInfo.PodGroup == nil {
			continue
		}
		annoMap := jobInfo.PodGroup.Annotations
		if annoMap == nil {
			annoMap = make(map[string]string)
			jobInfo.PodGroup.Annotations = annoMap
		}
		if jobInfo.PodGroup.Status.Phase == util.PodGroupInqueue {
			if _, exist := annoMap[util.EnqueueTimeAnnoKey]; !exist {
				annoMap[util.EnqueueTimeAnnoKey] = strconv.FormatInt(time.Now().UnixMilli(), util.Base10)
			}
			continue
		} else if !jobInfo.IsPending() {
			delete(annoMap, util.EnqueueTimeAnnoKey)
			delete(annoMap, util.DequeueFrequencyAnnoKey)
		}
	}
}

func jobDequeueForTimeout(vcjob *api.JobInfo, ssn *framework.Session) {
	for _, job := range ssn.Jobs {
		if job.Queue != vcjob.Queue {
			continue
		}
		if job.PodGroup == nil || job.PodGroup.Status.Phase != util.PodGroupInqueue {
			continue
		}
		if val, exist := job.PodGroup.Annotations[util.EnableDequeueAnnoKey]; !exist || val != util.EnableDequeueOnVal {
			continue
		}
		enqueueTimeStr, exist := job.PodGroup.Annotations[util.EnqueueTimeAnnoKey]
		if !exist {
			continue
		}
		enqueueTime, err := strconv.ParseInt(enqueueTimeStr, util.Base10, util.BitSize64)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("convert job <%s> enqueue time failed: %v", vcjob.Name, err)
			continue
		}
		if time.Now().UnixMilli()-enqueueTime > int64(util.EnqueueTimeOut) {
			execJobDequeue(ssn, job)
		}
	}
}

func execJobDequeue(ssn *framework.Session, job *api.JobInfo) {
	klog.V(util.LogInfoLev).Infof(" <%s> dequeue", job.Name)
	job.PodGroup.Status.Phase = ""
	delete(job.PodGroup.Annotations, util.EnqueueTimeAnnoKey)
	ssn.Jobs[job.UID] = job
	dequeStartTimes := "1"
	dequeueTimesStr, exist := job.PodGroup.Annotations[util.DequeueFrequencyAnnoKey]
	if !exist {
		job.PodGroup.Annotations[util.DequeueFrequencyAnnoKey] = dequeStartTimes
		return
	}
	dequeueTimes, err := strconv.ParseInt(dequeueTimesStr, util.Base10, util.BitSize64)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("convert job <%s> dequeue frequency failed: %v", job.Name, err)
		job.PodGroup.Annotations[util.DequeueFrequencyAnnoKey] = dequeStartTimes
	} else {
		job.PodGroup.Annotations[util.DequeueFrequencyAnnoKey] = strconv.FormatInt(dequeueTimes+1, util.Base10)
	}
}

func addPodGroupCondition(job *api.JobInfo, sessionID types.UID, reason, message string) {
	jc := scheduling.PodGroupCondition{
		Type:               scheduling.PodGroupUnschedulableType,
		Status:             v1.ConditionTrue,
		TransitionID:       string(sessionID),
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             reason,
		Message:            message,
	}

	index := -1
	for i, cond := range job.PodGroup.Status.Conditions {
		if cond.Type == scheduling.PodGroupUnschedulableType && cond.Reason == reason {
			index = i
			break
		}
	}

	if index >= 0 {
		job.PodGroup.Status.Conditions[index] = jc
	} else {
		job.PodGroup.Status.Conditions = append(job.PodGroup.Status.Conditions, jc)
	}
}

func (tp *huaweiNPUPlugin) addBatchOrderFailedCondition(job *api.JobInfo, ssn *framework.Session) {
	batchOrderError, ok := tp.Scheduler.BatchOrderError[job.UID]
	if !ok {
		return
	}
	addPodGroupCondition(job, ssn.UID, util.BatchOrderFailedReason, batchOrderError.Error())
}

func (tp *huaweiNPUPlugin) addNodePredicateFailedCondition(job *api.JobInfo, ssn *framework.Session) {
	const maxPrint = 20
	var message string
	if nodes, ok := tp.Scheduler.PredicatedNodes[job.UID]; ok {
		message += fmt.Sprintf("Predicated-Nodes count: %d, nodes: %v..., ", len(nodes), nodes.List()[:util.Min(nodes.Len(),
			maxPrint)])
	}
	for _, fitError := range job.NodesFitErrors {
		message += fitError.Error()
	}
	nodePredicateErr := tp.Scheduler.NodePredicateErrors.Get(job.UID)
	if nodePredicateErr != nil {
		for errStr, nodes := range nodePredicateErr {
			message += fmt.Sprintf(" Reason: %s, such as: %v...", errStr, nodes.List()[:util.Min(nodes.Len(),
				maxPrint)])
		}
	}
	if message == "" {
		return
	}

	addPodGroupCondition(job, ssn.UID, util.NodePredicateFailedReason, message)
}

func (tp *huaweiNPUPlugin) addJobValidFailedCondition(job *api.JobInfo, ssn *framework.Session) {
	result, ok := tp.Scheduler.ValidResult[job.UID]
	if !ok {
		return
	}
	addPodGroupCondition(job, ssn.UID, util.JobValidateFailedReason, fmt.Sprintf("%s: %s", result.Reason, result.Message))
}

func (tp *huaweiNPUPlugin) addJobEnqueueFailedCondition(job *api.JobInfo, ssn *framework.Session) {
	enqueueError, ok := tp.Scheduler.EnqueueError[job.UID]
	if !ok {
		return
	}

	addPodGroupCondition(job, ssn.UID, util.JobEnqueueFailedReason, enqueueError.Error())
}
