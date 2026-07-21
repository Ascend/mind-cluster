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
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/consts"
)

// NPUAllocateFunc Allocate npu and called by volcano frame.
func (sHandle ScheduleHandler) NPUAllocateFunc(task *api.TaskInfo) {
	if task == nil {
		klog.V(util.LogErrorLev).Infof("NPUAllocateFunc %s.", util.ArgumentError)
		return
	}

	if !sHandle.isTaskNeedNPUAllocated(task) {
		klog.V(util.LogDebugLev).Infof("NPUAllocateFunc %s no need to set pod annotation.", task.Name)
		return
	}

	vcJob, ok := sHandle.Jobs[task.Job]
	if !ok {
		klog.V(util.LogDebugLev).Infof("NPUAllocateFunc %s not req npu.", task.Name)
		return
	}
	if !vcJob.isNPUJob() {
		klog.V(util.LogDebugLev).Infof("NPUAllocateFunc vc-job:%#v is not npu job.", vcJob)
		return
	}
	if !*vcJob.JobReadyTag {
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
	vcNode := vcJob.policyHandler.UseAnnotation(task, node)
	if vcNode != nil {
		npuResName := v1.ResourceName(vcJob.ReqNPUName)
		sHandle.updateChipCountAfterAllocate(task, vcNode, npuResName)
		sHandle.Nodes[nodeName] = *vcNode
	}
	if sHandle.FaultHandle != nil {
		sHandle.FaultHandle.UseAnnotation(task)
	}

	// Record pod-to-node mapping for "prefer previous node" scheduling.
	// The cache internally saves the old Node as a rollback anchor (Previous field).
	if sHandle.AffinityCache != nil && vcJob.Owner.UID != "" {
		if rankIndex, ok := task.Pod.Annotations[PodRankIndexKey]; ok && rankIndex != "" {
			sHandle.AffinityCache.RecordAssignment(vcJob.Owner.UID, rankIndex, nodeName)
		}
	}

	// For hot-switch backup pods, notify the policy handler so it can update
	// internal state (e.g. SuperPods) with the backup pod's new node.
	if _, isBackup := task.Pod.Annotations[consts.BackupSourcePodNameKey]; isBackup {
		if hook, ok := vcJob.policyHandler.(BackupPodAllocatedHook); ok {
			hook.OnBackupPodAllocated(task, &vcJob, nodeName)
		}
	}

	klog.V(util.LogDebugLev).Infof("%s %s useAnnotation node [%s]'s top.", PluginName, util.SafePrint(task.Name), nodeName)
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
	if !vcJob.isNPUJob() {
		klog.V(util.LogDebugLev).Infof("NPUDeallocateFunc vc-job:%#v is not npu job.", vcJob)
		return
	}
	nodeName := task.NodeName
	node, found := sHandle.Nodes[nodeName]
	if !found {
		klog.V(util.LogWarningLev).Infof("%s npuAllocateFunc NOT EXIST node [%s].", PluginName, nodeName)
		return
	}
	sHandle.releaseAnnotation(task, vcJob, node)
	npuResName := v1.ResourceName(vcJob.ReqNPUName)
	sHandle.updateChipCountAfterDeallocate(task, nodeName, npuResName)

	// When the task status is Pending (allocation rollback / unpipeline), the
	// cache entry written by NPUAllocateFunc is stale — the pod never actually
	// ran on this node. RollbackAssignment restores the previous assignment
	// if one exists, or removes the entry entirely if this was a first-time assignment.
	// When the task is Releasing (evicted by preempt/reclaim), keep the cache
	// so the pod can be rescheduled back to its original node to reuse images.
	if task.Status == api.Pending {
		if sHandle.AffinityCache != nil && vcJob.Owner.UID != "" {
			if rankIndex, ok := task.Pod.Annotations[PodRankIndexKey]; ok && rankIndex != "" {
				sHandle.AffinityCache.RollbackAssignment(vcJob.Owner.UID, rankIndex)
			}
		}
	}

	klog.V(util.LogDebugLev).Infof("%s %s NPUDeallocateFunc node [%s]'s top.",
		PluginName, util.SafePrint(task.Name), nodeName)
}

// isTaskNeedNPUAllocated to judge the task is static cut. true is dynamic cut.
func (sHandle ScheduleHandler) isTaskNeedNPUAllocated(task *api.TaskInfo) bool {
	if !util.IsNPUTask(task) {
		klog.V(util.LogDebugLev).Infof("isTaskNeedNPUAllocated %s not npu task.", task.Name)
		return false
	}
	return true
}

func (sHandle *ScheduleHandler) releaseAnnotation(task *api.TaskInfo, vcJob SchedulerJob, vcNode NPUNode) {
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
	if task.Status == api.Pending {
		delete(task.Pod.Annotations, util.AscendNPUPodRealUse)
		delete(task.Pod.Annotations, vcTask.ReqNPUName)
		delete(task.Pod.Annotations, util.Pod910DeviceKey)
		return
	}
	tmpNode := vcJob.policyHandler.ReleaseAnnotation(task, vcNode)
	if tmpNode != nil {
		sHandle.Nodes[vcNode.Name] = *tmpNode
	}
	delete(task.Pod.Annotations, util.AscendNPUPodRealUse)
	delete(task.Pod.Annotations, vcTask.ReqNPUName)
	delete(task.Pod.Annotations, util.Pod910DeviceKey)
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

func (sHandle *ScheduleHandler) updateChipCountAfterAllocate(task *api.TaskInfo, vcNode *NPUNode,
	npuResName v1.ResourceName) {
	if sHandle == nil || task == nil || vcNode == nil {
		return
	}
	chipIDs := getAllocatedChipIDsFromPod(task.Pod, vcNode)
	for _, chipID := range chipIDs {
		chip, ok := vcNode.Chips[chipID]
		if !ok {
			continue
		}
		chip.PodMap[string(task.Pod.UID)] = task.Pod
	}
	if len(chipIDs) > 0 && npuResName != "" {
		if vcNode.Idle == nil {
			vcNode.Idle = make(map[v1.ResourceName]float64)
		}
	}
}

func (sHandle *ScheduleHandler) updateChipCountAfterDeallocate(task *api.TaskInfo, nodeName string,
	npuResName v1.ResourceName) {
	if sHandle == nil || task == nil {
		return
	}
	vcNode, ok := sHandle.Nodes[nodeName]
	if !ok {
		return
	}
	chipIDs := getAllocatedChipIDsFromPod(task.Pod, &vcNode)
	podUID := string(task.Pod.UID)
	for _, chipID := range chipIDs {
		chip, chipOK := vcNode.Chips[chipID]
		if !chipOK {
			continue
		}
		delete(chip.PodMap, podUID)
	}
	if len(chipIDs) > 0 && npuResName != "" {
		if vcNode.Idle == nil {
			vcNode.Idle = make(map[v1.ResourceName]float64)
		}
	}
	sHandle.Nodes[nodeName] = vcNode
}

func getAllocatedChipIDsFromPod(pod *v1.Pod, vcNode *NPUNode) []int {
	chipIDs := make([]int, 0)
	if pod == nil || vcNode == nil || pod.Annotations == nil {
		return chipIDs
	}
	annoPrefixToNpuPre := map[string]string{
		util.NPU910CardName:  util.NPU910CardNamePre,
		util.NPU310PCardName: util.NPU310PCardNamePre,
		util.NPU310CardName:  util.NPU310CardNamePre,
		util.Ascend910bName:  util.NPU910CardNamePre,
		util.NPUCardName:     util.NPUCardNamePre,
	}
	for annoKey, npuPre := range annoPrefixToNpuPre {
		if topStr, ok := pod.Annotations[annoKey]; ok && topStr != "" {
			chipIDs = util.ChangeTopToIntArray(topStr, npuPre)
			if len(chipIDs) > 0 {
				return chipIDs
			}
		}
	}
	return chipIDs
}

// CalcCardFreeCount calculates free chip count per card after preempting the given preemptees.
func CalcCardFreeCount(vcNode *NPUNode, preemptees []*api.TaskInfo, maxCardNPUNum int,
	availableChipIDs []int) map[int]int {
	klog.V(util.LogInfoLev).Infof("CalcCardFreeCount: preemptees=%v", preemptees)
	if vcNode == nil || maxCardNPUNum <= 0 {
		klog.V(util.LogInfoLev).Infof("CalcCardFreeCount: invalid args, vcNode nil=%v maxCardNPUNum=%d",
			vcNode == nil, maxCardNPUNum)
		return nil
	}
	cardFreeCount := make(map[int]int)

	// preemptee pod UID set
	preempteeSet := make(map[string]struct{})
	for _, pe := range preemptees {
		if pe != nil && pe.Pod != nil {
			preempteeSet[string(pe.Pod.UID)] = struct{}{}
		}
	}

	// allSchedulableChips = idle chips (from annotation) + chips occupied by any task.
	allSchedulableChips := make(map[int]struct{})
	for _, cid := range availableChipIDs {
		allSchedulableChips[cid] = struct{}{}
	}

	// collect chips occupied by non-preemptee pods (cannot be freed)
	nonPreempteeChips := make(map[int]struct{})
	for _, t := range vcNode.Tasks {
		if t == nil || t.Pod == nil {
			continue
		}
		_, isPE := preempteeSet[string(t.Pod.UID)]
		for _, cid := range getAllocatedChipIDsFromPod(t.Pod, vcNode) {
			allSchedulableChips[cid] = struct{}{}
			if !isPE {
				nonPreempteeChips[cid] = struct{}{}
			}
		}
	}

	// iterate the full schedulable chip set, exclude unhealthy and non-preemptee occupied
	for cid := range allSchedulableChips {
		if _, unhealthy := vcNode.UnhealthyChipIds[cid]; unhealthy {
			continue
		}
		if _, occupied := nonPreempteeChips[cid]; occupied {
			continue
		}
		cardID := cid / maxCardNPUNum
		cardFreeCount[cardID]++
	}
	klog.V(util.LogInfoLev).Infof("CalcCardFreeCount: node<%s> maxCardNPUNum<%d> preemptees<%d> "+
		"availableChipIDs=%v cardFreeCount=%v", vcNode.Name, maxCardNPUNum, len(preemptees),
		availableChipIDs, cardFreeCount)
	return cardFreeCount
}

func FilterPreempteesByFeasibleCards(vcNode *NPUNode, preemptees []*api.TaskInfo,
	feasibleCards map[int]struct{}, maxCardNPUNum int) []*api.TaskInfo {

	if vcNode == nil || len(feasibleCards) == 0 || maxCardNPUNum <= 0 {
		klog.V(util.LogInfoLev).Infof("FilterPreempteesByFeasibleCards: invalid args, vcNode nil=%v "+
			"feasibleCards=%d maxCardNPUNum=%d", vcNode == nil, len(feasibleCards), maxCardNPUNum)
		return nil
	}
	var filtered []*api.TaskInfo
	for _, pe := range preemptees {
		chipIDs := getAllocatedChipIDsFromPod(pe.Pod, vcNode)
		onFeasibleCard := false
		for _, cid := range chipIDs {
			if _, ok := feasibleCards[cid/maxCardNPUNum]; ok {
				onFeasibleCard = true
				break
			}
		}
		if onFeasibleCard {
			filtered = append(filtered, pe)
		}
	}
	klog.V(util.LogInfoLev).Infof("FilterPreempteesByFeasibleCards: node<%s> feasibleCards=%d "+
		"preemptees<%d> filtered<%d>", vcNode.Name, len(feasibleCards), len(preemptees), len(filtered))
	return filtered
}
