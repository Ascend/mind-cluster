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

	v1 "k8s.io/api/core/v1"
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
		NPUPlugins:  sets.String{util.NPUCardName: {}, util.NPU910CardName: {}, util.NPU310CardName: {}, util.NPU310PCardName: {}},
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
		return jobOrderFn(l, r, tp.Scheduler)
	})

	addBatchNodeOrderFn(ssn, tp)

	addPreemptableFn(ssn, tp)

	addReclaimableFn(ssn, tp)

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
		klog.V(util.LogInfoLev).Infof("predicateFn: task<%s> on node<%s>", taskInfo.Name, nodeInfo.Name)
		predicateErr := tp.Scheduler.NodePredicate(taskInfo, nodeInfo)
		if predicateErr != nil {
			tp.Scheduler.NodePredicateErrors.Add(taskInfo.Job, nodeInfo.Name, predicateErr)
			return convertToNPUFitError(tp, taskInfo, nodeInfo, predicateErr)
		}
		klog.V(util.LogInfoLev).Infof("predicateFn: task<%s> on node<%s> passed", taskInfo.Name, nodeInfo.Name)
		return nil
	})
}

func convertToNPUFitError(tp *huaweiNPUPlugin, taskInfo *api.TaskInfo,
	nodeInfo *api.NodeInfo, predicateErr error) error {
	if isNPUSchedulableByPreemption(tp, taskInfo, nodeInfo, predicateErr) {
		klog.V(util.LogInfoLev).Infof("predicate: task<%s> on node<%s> is unschedulable but preemptable, "+
			"can schedule after preempting lower priority tasks, reason: %s",
			taskInfo.Name, nodeInfo.Name, predicateErr.Error())
		return api.NewFitErrWithStatus(taskInfo, nodeInfo, &api.Status{
			Code:   api.Unschedulable,
			Reason: predicateErr.Error(),
		})
	}
	klog.V(util.LogInfoLev).Infof("predicate: task<%s> on node<%s> is unschedulable and unresolvable, "+
		"preemption cannot help, reason: %s",
		taskInfo.Name, nodeInfo.Name, predicateErr.Error())
	return api.NewFitErrWithStatus(taskInfo, nodeInfo, &api.Status{
		Code:   api.UnschedulableAndUnresolvable,
		Reason: predicateErr.Error(),
	})
}

func isNPUSchedulableByPreemption(tp *huaweiNPUPlugin, taskInfo *api.TaskInfo,
	nodeInfo *api.NodeInfo, predicateErr error) bool {
	if !isResourceShortageError(predicateErr) {
		klog.V(util.LogInfoLev).Infof("isNPUSchedulableByPreemption: task<%s> node<%s> not resource shortage err: %s",
			taskInfo.Name, nodeInfo.Name, predicateErr.Error())
		return false
	}
	vcNode, ok := tp.Scheduler.Nodes[nodeInfo.Name]
	if !ok {
		klog.V(util.LogInfoLev).Infof("isNPUSchedulableByPreemption: node<%s> not found in cache", nodeInfo.Name)
		return false
	}
	vcJob, ok := tp.Scheduler.Jobs[taskInfo.Job]
	if !ok {
		klog.V(util.LogInfoLev).Infof("isNPUSchedulableByPreemption: job<%s> not found in cache", taskInfo.Job)
		return false
	}
	if vcJob.NPUJob == nil {
		klog.V(util.LogInfoLev).Infof("isNPUSchedulableByPreemption: job<%s> is not NPU job", taskInfo.Job)
		return false
	}
	vcTask, ok := vcJob.NPUJob.Tasks[taskInfo.UID]
	if !ok {
		klog.V(util.LogInfoLev).Infof("isNPUSchedulableByPreemption: task<%s> not found in NPU job", taskInfo.Name)
		return false
	}
	_, total, _ := vcNode.GetChipCount(v1.ResourceName(vcTask.ReqNPUName))
	// for distributed job, need to remove the net unhealthy npu from total
	if vcJob.NPUJob.NPUTaskNum > 1 {
		total = subtractNetUnhealthyNPU(vcNode, vcTask.ReqNPUName, total)
	}
	result := vcTask.ReqNPUNum <= total
	klog.V(util.LogInfoLev).Infof("isNPUSchedulableByPreemption: task<%s> req<%d> total<%d> on node<%s>, preemptable=%v",
		taskInfo.Name, vcTask.ReqNPUNum, total, nodeInfo.Name, result)
	return result
}

func isResourceShortageError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, util.NPUResourceShortageError)
}

// subtractNetUnhealthyNPU subtracts network unhealthy npu count from total chip count.
func subtractNetUnhealthyNPU(node plugin.NPUNode, reqNPUName string, total int) int {
	netUnhealthyKey := getNetworkUnhealthyNPUKey(reqNPUName)
	netUnhealthyStr, ok := node.Annotation[netUnhealthyKey]
	if !ok || netUnhealthyStr == "" {
		return total
	}
	annoPreVal := util.NPU910CardNamePre
	if reqNPUName == util.NPUCardName {
		annoPreVal = util.NPUCardNamePre
	}
	netUnhealthyTop := util.ChangeTopToIntArray(netUnhealthyStr, annoPreVal)
	return total - len(netUnhealthyTop)
}

// getNetworkUnhealthyNPUKey returns the annotation key for network unhealthy npu.
func getNetworkUnhealthyNPUKey(reqNPUName string) string {
	if reqNPUName == util.NPUCardName {
		return util.NPUCardName + "-NetworkUnhealthy"
	}
	return util.HwPreName + util.Ascend910 + "-NetworkUnhealthy"
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
	if !*job.JobReadyTag {
		return util.Reject
	}
	klog.V(util.LogInfoLev).Infof("job %s/%s WaitingTaskNum: %d, ReadyTaskNum: %d, MinAvailable: %d", ji.Namespace,
		ji.Name, ji.WaitingTaskNum(), ji.ReadyTaskNum(), job.MinAvailable)
	if ji.WaitingTaskNum()+ji.ReadyTaskNum() < job.MinAvailable {
		return util.Reject
	}
	return util.Abstain
}

func addBatchNodeOrderFn(ssn *framework.Session, tp *huaweiNPUPlugin) {
	ssn.AddBatchNodeOrderFn(tp.Name(), func(task *api.TaskInfo, nodes []*api.NodeInfo) (map[string]float64, error) {
		klog.V(util.LogInfoLev).Infof("batchNodeOrderFn: task<%s> scoring %d nodes", task.Name, len(nodes))
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
			klog.V(util.LogInfoLev).Infof("batchNodeOrderFn: task<%s> scoring error: %v", task.Name, err)
		}
		klog.V(util.LogInfoLev).Infof("batchNodeOrderFn: task<%s> scored %d nodes", task.Name, len(score))
		return score, nil
	})
}

func addPreemptableFn(ssn *framework.Session, tp *huaweiNPUPlugin) {
	ssn.AddPreemptableFn(tp.Name(), func(preemptor *api.TaskInfo, preemptees []*api.TaskInfo) ([]*api.TaskInfo, int) {
		klog.V(util.LogInfoLev).Infof("preemptableFn: task<%s> evaluating %d preemptees", preemptor.Name, len(preemptees))
		vcJob, ok := tp.Scheduler.Jobs[preemptor.Job]
		if !ok || vcJob.NPUJob == nil || vcJob.GetPolicyHandler() == nil {
			klog.V(util.LogInfoLev).Infof("preemptableFn: task<%s> job not found or not NPU job, Abstain",
				preemptor.Name)
			return nil, util.Abstain
		}
		vcTask, ok := vcJob.NPUJob.Tasks[preemptor.UID]
		if !ok || vcTask.ReqNPUNum <= 0 {
			klog.V(util.LogInfoLev).Infof("preemptableFn: task<%s> not found or reqNPU<=0, Abstain",
				preemptor.Name)
			return nil, util.Abstain
		}
		maxCardNPUNum := vcJob.GetPolicyHandler().GetMaxCardNPUNum()
		if maxCardNPUNum <= 0 {
			klog.V(util.LogInfoLev).Infof("preemptableFn: task<%s> maxCardNPUNum=0, Abstain", preemptor.Name)
			return nil, util.Abstain
		}

		if len(preemptees) == 0 {
			klog.V(util.LogInfoLev).Infof("preemptableFn: task<%s> no preemptees, Abstain", preemptor.Name)
			return nil, util.Abstain
		}
		nodeName := preemptees[0].NodeName
		if nodeName == "" {
			klog.V(util.LogInfoLev).Infof("preemptableFn: task<%s> preemptee has no node, Abstain",
				preemptor.Name)
			return nil, util.Abstain
		}
		vcNode, ok := tp.Scheduler.Nodes[nodeName]
		if !ok {
			klog.V(util.LogInfoLev).Infof("preemptableFn: node<%s> not found in cache, Abstain", nodeName)
			return nil, util.Abstain
		}

		klog.V(util.LogInfoLev).Infof("preemptableFn: task<%s> req<%d> maxCardNPUNum<%d> on node<%s>, "+
			"preemptees<%d>", preemptor.Name, vcTask.ReqNPUNum, maxCardNPUNum, nodeName, len(preemptees))
		filtered, ok := vcJob.GetPolicyHandler().Preemptable(preemptor, preemptees, &vcNode)
		if !ok || len(filtered) == 0 {
			klog.V(util.LogInfoLev).Infof("preemptableFn: task<%s> on node<%s> no feasible victims, Abstain",
				preemptor.Name, nodeName)
			return nil, util.Abstain
		}
		klog.V(util.LogInfoLev).Infof("preemptableFn: task<%s> on node<%s>, filtered %d/%d preemptees, Permit",
			preemptor.Name, nodeName, len(filtered), len(preemptees))
		return filtered, util.Permit
	})
}

func addReclaimableFn(ssn *framework.Session, tp *huaweiNPUPlugin) {
	ssn.AddReclaimableFn(tp.Name(), func(reclaimer *api.TaskInfo, reclaimees []*api.TaskInfo) ([]*api.TaskInfo, int) {
		klog.V(util.LogInfoLev).Infof("reclaimableFn: task<%s> evaluating %d reclaimees", reclaimer.Name, len(reclaimees))
		vcJob, ok := tp.Scheduler.Jobs[reclaimer.Job]
		if !ok || vcJob.NPUJob == nil || vcJob.GetPolicyHandler() == nil {
			klog.V(util.LogInfoLev).Infof("reclaimableFn: task<%s> job not found or not NPU job, Abstain",
				reclaimer.Name)
			return nil, util.Abstain
		}
		vcTask, ok := vcJob.NPUJob.Tasks[reclaimer.UID]
		if !ok || vcTask.ReqNPUNum <= 0 {
			klog.V(util.LogInfoLev).Infof("reclaimableFn: task<%s> not found or reqNPU<=0, Abstain",
				reclaimer.Name)
			return nil, util.Abstain
		}
		maxCardNPUNum := vcJob.GetPolicyHandler().GetMaxCardNPUNum()
		if maxCardNPUNum <= 0 {
			klog.V(util.LogInfoLev).Infof("reclaimableFn: task<%s> maxCardNPUNum=0, Abstain", reclaimer.Name)
			return nil, util.Abstain
		}

		if len(reclaimees) == 0 {
			klog.V(util.LogInfoLev).Infof("reclaimableFn: task<%s> no reclaimees, Abstain", reclaimer.Name)
			return nil, util.Abstain
		}
		nodeName := reclaimees[0].NodeName
		if nodeName == "" {
			klog.V(util.LogInfoLev).Infof("reclaimableFn: task<%s> reclaimee has no node, Abstain",
				reclaimer.Name)
			return nil, util.Abstain
		}
		vcNode, ok := tp.Scheduler.Nodes[nodeName]
		if !ok {
			klog.V(util.LogInfoLev).Infof("reclaimableFn: node<%s> not found in cache, Abstain", nodeName)
			return nil, util.Abstain
		}

		klog.V(util.LogInfoLev).Infof("reclaimableFn: task<%s> req<%d> maxCardNPUNum<%d> on node<%s>, "+
			"reclaimees<%d>", reclaimer.Name, vcTask.ReqNPUNum, maxCardNPUNum, nodeName, len(reclaimees))
		filtered, ok := vcJob.GetPolicyHandler().Reclaimable(reclaimer, reclaimees, &vcNode)
		if !ok || len(filtered) == 0 {
			klog.V(util.LogInfoLev).Infof("reclaimableFn: task<%s> on node<%s> no feasible victims, Abstain",
				reclaimer.Name, nodeName)
			return nil, util.Abstain
		}
		klog.V(util.LogInfoLev).Infof("reclaimableFn: task<%s> on node<%s>, filtered %d/%d reclaimees, Permit",
			reclaimer.Name, nodeName, len(filtered), len(reclaimees))
		return filtered, util.Permit
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
			klog.V(util.LogInfoLev).Infof("AllocateFunc: task<%s> on node<%s>",
				event.Task.Name, event.Task.NodeName)
			tp.Scheduler.NPUAllocateFunc(event.Task)
		},
		DeallocateFunc: func(event *framework.Event) {
			if event == nil {
				klog.V(util.LogErrorLev).Infof("DeallocateFunc event nil.")
				return
			}
			klog.V(util.LogInfoLev).Infof("DeallocateFunc: task<%s> on node<%s>",
				event.Task.Name, event.Task.NodeName)
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
	jobInfo, exist := tp.Scheduler.Jobs[vcjob.UID]
	if !exist {
		return util.JobEnqueueSkip
	}
	if !tp.Scheduler.NPUPlugins.Has(jobInfo.ReqNPUName) {
		return util.JobEnqueueSkip
	}
	tNpuNum := getNpuNum(ssn, tp, jobInfo.ReqNPUName)
	if tNpuNum < jobInfo.ReqNPUNum {
		klog.V(util.LogWarningLev).Infof("job <%s> Add enqueue failed, require npu num is %v "+
			"but cluster npu num is %v", vcjob.Name, jobInfo.ReqNPUNum, tNpuNum)
		tp.Scheduler.EnqueueError[vcjob.UID] = fmt.Errorf("require npu num is %v, but cluster npu num is %v", jobInfo.ReqNPUNum,
			tNpuNum)
		return util.JobNotEnqueue
	}
	if tp.Scheduler.FrameAttr.ForceEnqueue {
		klog.V(util.LogWarningLev).Infof("job <%s> Add enqueue success will start schedule, require npu num is <%v> "+
			"and cluster npu num is <%v>.", vcjob.Name, jobInfo.ReqNPUNum, tNpuNum)
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
		if !ok || len(deviceList) > int(npuNum/util.NPUHexKilo) {
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
		shareDevCount := 1
		if node.Node != nil {
			softShareDevEnable, softShareDevEnableExist := node.Node.Labels[util.SchedulerSoftShareDevEnableNodeLabel]
			if softShareDevEnableExist && softShareDevEnable == "true" {
				shareDevCount = util.SoftShareDevCount
			}
		}
		tNpuNum += len(deviceList) * shareDevCount
	}
	errs.Print()
	return tNpuNum
}

// isFragileJob judges whether a job is fragile (ready task num <= minAvailable).
// For NPU jobs, use the plugin's MinAvailable from annotation; otherwise fall back to framework's.
func isFragileJob(jobInfo *api.JobInfo, sHandle *plugin.ScheduleHandler) bool {
	if sHandle != nil {
		if vcJob, ok := sHandle.Jobs[jobInfo.UID]; ok && vcJob.IsNPUJob() {
			return jobInfo.ReadyTaskNum() <= vcJob.MinAvailable
		}
	}
	return jobInfo.ReadyTaskNum() <= jobInfo.MinAvailable
}

func jobOrderFn(interfaceA interface{}, interfaceB interface{}, sHandle *plugin.ScheduleHandler) int {
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
	aFragile := isFragileJob(jobInfoA, sHandle)
	bFragile := isFragileJob(jobInfoB, sHandle)
	if aFragile && !bFragile {
		klog.V(util.LogInfoLev).Infof("jobOrderFn: job<%s> is fragile, job<%s> is not, A high priority",
			jobInfoA.Name, jobInfoB.Name)
		return util.JobOrderHighPriority
	}
	if !aFragile && bFragile {
		klog.V(util.LogInfoLev).Infof("jobOrderFn: job<%s> is not fragile, job<%s> is fragile, A low priority",
			jobInfoA.Name, jobInfoB.Name)
		return util.JobOrderLowPriority
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
	}
	return util.JobOrderSamePriority
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
