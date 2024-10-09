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
Package plugin is using for HuaWei Ascend pin affinity schedule frame.
*/
package plugin

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// IsTaskNeedNPUAllocated to judge the task is static cut. true is dynamic cut.
func (sHandle ScheduleHandler) IsTaskNeedNPUAllocated(task *api.TaskInfo) bool {
	if !IsNPUTask(task) {
		klog.V(util.LogDebugLev).Infof("IsTaskNeedNPUAllocated %s not npu task.", task.Name)
		return false
	}

	vcJob, ok := sHandle.Jobs[task.Job]
	if !ok {
		klog.V(util.LogDebugLev).Infof("IsTaskNeedNPUAllocated %s not in npu jobs.", task.Job)
		return false
	}
	nTask, ok := vcJob.Tasks[task.UID]
	if !ok {
		klog.V(util.LogDebugLev).Infof("IsTaskNeedNPUAllocated %s not in npu tasks.", task.Name)
		return false
	}
	// static cut job no need allocated,it followed by kubelet in device-plugin.
	if nTask.Type == util.JobTypeStCut {
		klog.V(util.LogDebugLev).Infof("IsTaskNeedNPUAllocated %s is static cut job.", task.Name)
		return false
	}
	return true
}

// NPUAllocateFunc Allocate npu and called by volcano frame.
func (sHandle ScheduleHandler) NPUAllocateFunc(task *api.TaskInfo) {
	if task == nil {
		klog.V(util.LogErrorLev).Infof("NPUAllocateFunc %s.", util.ArgumentError)
		return
	}

	if !sHandle.IsTaskNeedNPUAllocated(task) {
		klog.V(util.LogDebugLev).Infof("NPUAllocateFunc %s no need to set pod annotation.", task.Name)
		return
	}

	vcJob, ok := sHandle.Jobs[task.Job]
	if !ok {
		klog.V(util.LogDebugLev).Infof("NPUAllocateFunc %s not req npu.", task.Name)
		return
	}
	if !vcJob.JobReadyTag {
		klog.V(util.LogDebugLev).Infof("NPUAllocateFunc %s not allow allocate npu.", task.Name)
		return
	}
	nodeName := task.NodeName
	node, found := sHandle.Nodes[nodeName]
	if !found {
		klog.V(util.LogWarningLev).Infof("%s npuAllocateFunc %s not exist.", PluginName, nodeName)
		return
	}
	if vcJob.NPUTaskNum > 1 {
		task.Pod.Annotations[util.DistributedJobKey] = util.DistributedJobValue
	} else {
		task.Pod.Annotations[util.DistributedJobKey] = util.StandaloneJobValue
	}
	vcNode := vcJob.handler.UseAnnotation(task, node)
	if vcNode != nil {
		// update node.
		sHandle.Nodes[nodeName] = *vcNode
	}
	if vcJob.IsTorAffinityJob() && sHandle.getNSLBVsersion() == NSLB2Version {
		sHandle.setNslbV2PodAnnotation(task, vcJob, nodeName)
	}
	klog.V(util.LogDebugLev).Infof("%s %s useAnnotation node [%s]'s top.", PluginName, util.SafePrint(task.Name), nodeName)
}

func (sHandle ScheduleHandler) setNslbV2PodAnnotation(task *api.TaskInfo, vcJob SchedulerJob, nodeName string) {
	task.Pod.Annotations[isSharedTor] = strconv.Itoa(freeTor)
	task.Pod.Annotations[isHealthy] = strconv.Itoa(healthyTor)
	torIp := sHandle.Tors.torIpMap[nodeName]
	tor, getTor := sHandle.Tors.torMaps[torIp]
	if getTor && len(vcJob.ServerList) > 1 {
		task.Pod.Annotations[isSharedTor] = strconv.Itoa(tor.IsSharedTor)
		task.Pod.Annotations[isHealthy] = strconv.Itoa(tor.IsHealthy)
	}
	defer func() {
		if task.Pod.Annotations[isSharedTor] == strconv.Itoa(sharedTor) {
			task.Pod.Annotations[SharedTorIp] = tor.IP
		}
	}()
	if vcJob.SchedulingTaskNum == len(vcJob.Tasks) {
		return
	}
	usedTorInfo := vcJob.getUsedTorInfos(&sHandle)
	if usedTorInfo.isSingleTorJob {
		return
	}
	if tor.IsSharedTor != freeTor {
		task.Pod.Annotations[isSharedTor] = strconv.Itoa(tor.IsSharedTor)
		task.Pod.Annotations[isHealthy] = strconv.Itoa(tor.IsHealthy)
		return
	}
	if usedTorInfo.sharedTorNum <= 0 {
		task.Pod.Annotations[isSharedTor] = strconv.Itoa(exclusiveTor)
		tor.IsSharedTor = exclusiveTor
		task.Pod.Annotations[isHealthy] = strconv.Itoa(healthyTor)
	}
}

func (sHandle *ScheduleHandler) releaseAnnotation(task *api.TaskInfo, vcJob SchedulerJob, vcNode NPUNode) {
	rankIndex, ok := task.Pod.Annotations[podRankIndex]
	klog.V(util.LogInfoLev).Infof("task %s node %s rankIndex annotation: %s", task.Name, vcNode.Name, rankIndex)
	if ok { // if pod rankIndex has been written, delete it
		klog.V(util.LogInfoLev).Infof("node %s bind failed, release rankIndex annotation: %s",
			vcNode.Name, rankIndex)
		delete(task.Pod.Annotations, podRankIndex)
	}
	vcTask, ok := vcJob.Tasks[task.UID]
	if !ok {
		klog.V(util.LogInfoLev).Infof("task %s not in vcjob %s", vcTask.Name, vcJob.Name)
		return
	}
	reqStr, ok := task.Pod.Annotations[util.AscendNPUPodRealUse]
	if !ok {
		reqStr, ok = task.Pod.Annotations[vcTask.ReqNPUName]
		if !ok {
			return
		}
	}
	reqSlice := strings.Split(reqStr, ",")
	if len(reqSlice) != vcTask.ReqNPUNum {
		return
	}
	value, ok := vcNode.Annotation[vcTask.ReqNPUName]
	if !ok {
		return
	}
	vcNode.Annotation[vcTask.ReqNPUName] = reqStr
	if value != "" {
		// if failed, reset by next session.
		if isEachStringContainsSameElement(value, reqStr, ",") {
			annErr := fmt.Errorf("%s:%s has same NPU used %s:%s", vcNode.Name, value, vcTask.Name, reqStr)
			klog.V(util.LogErrorLev).Infof("releaseAnnotation %s", annErr)
			return
		}
		vcNode.Annotation[vcTask.ReqNPUName] = reqStr + "," + value
	}
	sHandle.Nodes[vcNode.Name] = vcNode
	klog.V(util.LogDebugLev).Infof("%s releaseAnnotation %s's %s on %s,new top:[%s].", PluginName, task.Name,
		reqStr, vcNode.Name, reqStr+","+value)
	tmpNode := vcJob.handler.ReleaseAnnotation(task, vcNode)
	if tmpNode != nil {
		// update node.
		sHandle.Nodes[vcNode.Name] = *tmpNode
	}
}

// NPUDeallocateFunc Free assigned npu, if allocate failed by volcano frame.
func (sHandle *ScheduleHandler) NPUDeallocateFunc(task *api.TaskInfo) {
	if sHandle == nil || task == nil {
		klog.V(util.LogInfoLev).Infof("NPUDeallocateFunc failed: %s.", util.ArgumentError)
		return
	}
	vcJob, ok := sHandle.Jobs[task.Job]
	if !ok {
		klog.V(util.LogDebugLev).Infof("NPUDeallocateFunc %s not req npu.", task.Name)
		return
	}
	nodeName := task.NodeName
	node, found := sHandle.Nodes[nodeName]
	if !found {
		klog.V(util.LogWarningLev).Infof("%s npuAllocateFunc NOT EXIST node [%s].", PluginName, nodeName)
		return
	}
	sHandle.releaseAnnotation(task, vcJob, node)
	klog.V(util.LogDebugLev).Infof("%s %s NPUDeallocateFunc node [%s]'s top.",
		PluginName, util.SafePrint(task.Name), nodeName)
}

func updatePodPendingReason(task *api.TaskInfo, reasonTmp string) {
	condition := v1.PodCondition{
		Type:    v1.PodScheduled,
		Status:  v1.ConditionFalse,
		Reason:  v1.PodReasonUnschedulable,
		Message: reasonTmp,
	}
	for _, tmp := range task.Pod.Status.Conditions {
		if strings.Contains(tmp.Message, reasonTmp) {
			klog.V(util.LogDebugLev).Infof("%s has record the reason:%s ,skip.", task.Name, reasonTmp)
			return
		}
	}
	task.Pod.Status.Conditions = append(task.Pod.Status.Conditions, condition)
}

// IsNPUTask to judge the task either is NPU task or not.
func IsNPUTask(nT *api.TaskInfo) bool {
	for k := range nT.Resreq.ScalarResources {
		// must contain "huawei.com/"
		if strings.Contains(string(k), util.HwPreName) {
			return true
		}
	}
	return false
}
