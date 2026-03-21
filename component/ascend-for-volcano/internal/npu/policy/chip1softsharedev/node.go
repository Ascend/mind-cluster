/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package chip1softsharedev is using for HuaWei chip1softsharedev schedule.
package chip1softsharedev

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func (tp *chip1softsharedev) getChipMemoryFromNodeLabel(nodeLabel map[string]string) (int, error) {
	memLabel, ok := nodeLabel[util.NPUChipMemoryLabelKey]
	if !ok {
		return 0, errors.New("missing npu-chip-memory label")
	}
	memStr := strings.Replace(memLabel, "G", "", -1)
	mem, err := strconv.Atoi(memStr)
	if err != nil {
		return 0, fmt.Errorf("invalid npu-chip-memory value: %w", err)
	}
	return mem * util.MBPerGB, nil
}

func (tp *chip1softsharedev) getUsedResourceMapFromNodeTasks(
	tasks map[api.TaskID]*api.TaskInfo, skipTargetJob bool, targetJob *api.TaskInfo) map[int]softShareDevResource {
	usedMap := make(map[int]softShareDevResource)
	for _, taskInfo := range tasks {
		if skipTargetJob && targetJob != nil && taskInfo.Job == targetJob.Job {
			klog.V(util.LogDebugLev).Infof("[chip1softsharedev] skip target task %s (job: %s)",
				taskInfo.Name, targetJob.Job)
			continue
		}
		cardID, resource, ok := tp.parseTaskResource(taskInfo, tp.GetAnnoPreVal(tp.ReqNPUName))
		if !ok {
			continue
		}
		usedCard := usedMap[cardID]
		usedCard.aicoreQuota += resource.aicoreQuota
		usedCard.hbmQuota += resource.hbmQuota
		usedCard.schedulingPolicy = resource.schedulingPolicy
		usedMap[cardID] = usedCard
	}
	return usedMap
}

func (tp *chip1softsharedev) parseTaskResource(taskInfo *api.TaskInfo, annoPrefix string) (
	int, softShareDevResource, bool) {
	if taskInfo == nil || taskInfo.Pod == nil || taskInfo.Pod.Annotations == nil {
		klog.V(util.LogErrorLev).Infof("[chip1softsharedev] parseTaskResource: taskInfo, taskInfo.Pod or " +
			"taskInfo.Pod.Annotations is nil")
		return 0, softShareDevResource{}, false
	}
	annotations := taskInfo.Pod.Annotations
	ascendReal, existAscend := annotations[util.AscendNPUPodRealUse]
	if !existAscend {
		ascendReal, existAscend = taskInfo.Pod.Annotations[tp.GetAnnoName(tp.ReqNPUName)]
	}
	aicoreAnno, existAicore := annotations[util.SchedulerSoftShareDevAicoreQuotaKey]
	hbmAnno, existHbm := annotations[util.SchedulerSoftShareDevHbmQuotaKey]
	policyAnno, existPolicy := annotations[util.SchedulerSoftShareDevPolicyKey]

	if !existAscend || !existAicore || !existHbm || !existPolicy {
		return 0, softShareDevResource{}, false
	}
	cardStr := strings.TrimPrefix(ascendReal, annoPrefix)
	cardID, err := strconv.Atoi(cardStr)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("[chip1softsharedev] task %s invalid card number: %s (prefix: %s), "+
			"err: %v", taskInfo.Name, ascendReal, annoPrefix, err)
		return 0, softShareDevResource{}, false
	}
	aicoreQuota, err := strconv.Atoi(aicoreAnno)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("[chip1softsharedev] task %s invalid aicore quota: %s, err: %v",
			taskInfo.Name, aicoreAnno, err)
		return 0, softShareDevResource{}, false
	}
	hbmQuota, err := strconv.Atoi(hbmAnno)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("[chip1softsharedev] task %s invalid hbm quota: %s, err: %v",
			taskInfo.Name, hbmAnno, err)
		return 0, softShareDevResource{}, false
	}
	resource := softShareDevResource{
		aicoreQuota:      aicoreQuota,
		hbmQuota:         hbmQuota,
		schedulingPolicy: policyAnno,
	}
	return cardID, resource, true
}

func (tp *chip1softsharedev) checkNodeUsableResourceForTask(node plugin.NPUNode, nodeTop []int, chipMemory int,
	reqResourceCfg softShareDevResource, task *api.TaskInfo) bool {
	if tp == nil || len(node.Annotation) == 0 || task == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("checkNodeUsableResource err: %s", err)
		return false
	}
	nodeUsedResourceMap := tp.getUsedResourceMapFromNodeTasks(node.Tasks, true, task)
	taskTotalReqResource := softShareDevResource{
		aicoreQuota:      reqResourceCfg.aicoreQuota * tp.NPUTaskNum,
		hbmQuota:         reqResourceCfg.hbmQuota * tp.NPUTaskNum,
		schedulingPolicy: reqResourceCfg.schedulingPolicy,
	}
	nodeUsableResourceForTask := softShareDevResource{schedulingPolicy: taskTotalReqResource.schedulingPolicy}
	for _, cardIdx := range nodeTop {
		used, exists := nodeUsedResourceMap[cardIdx]
		if !exists {
			nodeUsableResourceForTask.aicoreQuota += util.MaxAicoreQuota / reqResourceCfg.aicoreQuota *
				reqResourceCfg.aicoreQuota
			nodeUsableResourceForTask.hbmQuota += chipMemory / reqResourceCfg.hbmQuota * reqResourceCfg.hbmQuota
			continue
		}
		if used.schedulingPolicy != taskTotalReqResource.schedulingPolicy {
			continue
		}
		nodeUsableResourceForTask.aicoreQuota += (util.MaxAicoreQuota - used.aicoreQuota) / reqResourceCfg.aicoreQuota *
			reqResourceCfg.aicoreQuota
		nodeUsableResourceForTask.hbmQuota += (chipMemory - used.hbmQuota) / reqResourceCfg.hbmQuota *
			reqResourceCfg.hbmQuota
	}
	if taskTotalReqResource.aicoreQuota > nodeUsableResourceForTask.aicoreQuota ||
		taskTotalReqResource.hbmQuota > nodeUsableResourceForTask.hbmQuota {
		return false
	}
	return true
}
