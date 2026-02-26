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
	tasks map[api.TaskID]*api.TaskInfo) map[int]softShareDevResource {
	usedMap := make(map[int]softShareDevResource)
	for _, taskInfo := range tasks {
		ascendReal, existAscend := taskInfo.Pod.Annotations[util.AscendNPUPodRealUse]
		aicoreAnno, existAicore := taskInfo.Pod.Annotations[util.SchedulerSoftShareDevAicoreQuotaKey]
		hbmAnno, existHbm := taskInfo.Pod.Annotations[util.SchedulerSoftShareDevHbmQuotaKey]
		policyAnno, existPolicy := taskInfo.Pod.Annotations[util.SchedulerSoftShareDevPolicyKey]
		if !existAscend || !existAicore || !existHbm || !existPolicy {
			continue
		}
		cardInt, err := strconv.Atoi(strings.TrimPrefix(ascendReal, tp.GetAnnoPreVal(tp.ReqNPUName)))
		if err != nil {
			klog.V(util.LogErrorLev).Infof("invalid card number: %s, err: %v", ascendReal, err)
			continue
		}
		aicore, err := strconv.Atoi(aicoreAnno)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("invalid aicore quota: %s, err: %v", aicoreAnno, err)
			continue
		}
		hbm, err := strconv.Atoi(hbmAnno)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("invalid hbm quota: %s, err: %v", hbmAnno, err)
			continue
		}
		existing := usedMap[cardInt]
		existing.aicoreQuota += aicore
		existing.hbmQuota += hbm
		existing.schedulingPolicy = policyAnno
		usedMap[cardInt] = existing
	}
	return usedMap
}
