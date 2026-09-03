//go:build volcano_v115

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

package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"k8s.io/klog/v2"
	k8sfwk "k8s.io/kube-scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

const simulateNodeKey = "ascend-simulate-vcnode"

const (
	preemptActionName          = "preempt"
	topologyAwarePreemptionKey = "enableTopologyAwarePreemption"
)

func topologyAwarePreemptActive(ssn *framework.Session) bool {
	if ssn == nil {
		return false
	}
	args := framework.GetArgOfActionFromConf(ssn.Configurations, preemptActionName)
	if args == nil {
		return false
	}
	var active bool
	args.GetBool(&active, topologyAwarePreemptionKey)
	klog.V(util.LogInfoLev).Infof("topologyAwarePreemptActive: enabled=%t", active)
	return active
}

func addSimulateFns(ssn *framework.Session, tp *huaweiNPUPlugin) {
	ssn.AddSimulateRemoveTaskFn(tp.Name(), func(ctx context.Context, state k8sfwk.CycleState,
		taskToSchedule *api.TaskInfo, taskToRemove *api.TaskInfo, nodeInfo *api.NodeInfo) error {
		simNode, ok := getSimNode(tp, nodeInfo)
		if !ok {
			return nil
		}
		delete(simNode.Tasks, taskToRemove.UID)
		if simNode.ChipTopo != nil && taskToRemove.Pod != nil {
			if err := simNode.ChipTopo.Rollback(string(taskToRemove.Pod.UID)); err != nil {
				klog.V(util.LogWarningLev).Infof("simulateRemoveTaskFn: node<%s> remove task<%s>: %v",
					nodeInfo.Name, taskToRemove.Name, err)
			}
		}
		updateSimNodeIdleChips(&simNode, taskToRemove, true)
		nodeInfo.Others[simulateNodeKey] = simNode
		return nil
	})
	ssn.AddSimulateAddTaskFn(tp.Name(), func(ctx context.Context, state k8sfwk.CycleState,
		taskToSchedule *api.TaskInfo, taskToAdd *api.TaskInfo, nodeInfo *api.NodeInfo) error {
		simNode, ok := getSimNode(tp, nodeInfo)
		if !ok {
			return nil
		}
		simNode.Tasks[taskToAdd.UID] = taskToAdd
		if simNode.ChipTopo != nil && taskToAdd.Pod != nil {
			_, _, ids := taskChipNames(taskToAdd)
			if err := simNode.ChipTopo.TryAllocate(string(taskToAdd.Pod.UID), ids); err != nil {
				klog.V(util.LogWarningLev).Infof("simulateAddTaskFn: node<%s> restore task<%s>: %v",
					nodeInfo.Name, taskToAdd.Name, err)
			}
		}
		updateSimNodeIdleChips(&simNode, taskToAdd, false)
		nodeInfo.Others[simulateNodeKey] = simNode
		return nil
	})
	ssn.AddSimulatePredicateFn(tp.Name(), func(ctx context.Context, state k8sfwk.CycleState,
		taskInfo *api.TaskInfo, nodeInfo *api.NodeInfo) error {
		klog.V(util.LogInfoLev).Infof("simulatePredicateFn: task<%s> on node<%s>", taskInfo.Name, nodeInfo.Name)
		simNode, ok := getSimNode(tp, nodeInfo)
		if !ok {
			klog.V(util.LogDebugLev).Infof("simulatePredicateFn: node<%s> not in cache, pass", nodeInfo.Name)
			return nil
		}
		simNode.Idle = nodeInfo.Idle.ScalarResources
		predicateErr := tp.Scheduler.NodePredicateOnVCNode(taskInfo, simNode)
		if predicateErr != nil {
			klog.V(util.LogInfoLev).Infof("simulatePredicateFn: task<%s> on node<%s> failed: %s",
				taskInfo.Name, nodeInfo.Name, predicateErr.Error())
		}
		return predicateErr
	})
}

func getSimNode(tp *huaweiNPUPlugin, nodeInfo *api.NodeInfo) (plugin.NPUNode, bool) {
	if v, ok := nodeInfo.Others[simulateNodeKey]; ok {
		if sim, ok := v.(plugin.NPUNode); ok {
			return sim, true
		}
	}
	vcNode, ok := tp.Scheduler.Nodes[nodeInfo.Name]
	if !ok {
		return plugin.NPUNode{}, false
	}
	simNode := vcNode
	simNode.Idle = nodeInfo.Idle.ScalarResources
	simNode.Tasks = make(map[api.TaskID]*api.TaskInfo, len(nodeInfo.Tasks))
	for tid, t := range nodeInfo.Tasks {
		simNode.Tasks[tid] = t
	}
	simNode.Annotation = make(map[string]string, len(vcNode.Annotation))
	for k, v := range vcNode.Annotation {
		simNode.Annotation[k] = v
	}
	simNode.ChipTopo = vcNode.ChipTopo.Clone()
	nodeInfo.Others[simulateNodeKey] = simNode
	return simNode, true
}

func taskChipNames(task *api.TaskInfo) ([]string, string, []int) {
	if task == nil || task.Pod == nil {
		return nil, "", nil
	}
	devStr, ok := task.Pod.Annotations[util.AscendNPUPodRealUse]
	if !ok {
		devStr, ok = task.Pod.Annotations[util.NPU910CardName]
		if !ok {
			devStr, ok = task.Pod.Annotations[util.NPUCardName]
			if !ok {
				return nil, "", nil
			}
		}
	}
	prefix := util.NPU910CardNamePre
	if strings.HasPrefix(devStr, util.NPUCardNamePre) {
		prefix = util.NPUCardNamePre
	}
	ids := util.ChangeTopToIntArray(devStr, prefix)
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		names = append(names, fmt.Sprintf("%s%d", prefix, id))
	}
	return names, util.HwPreName + strings.TrimSuffix(prefix, "-"), ids
}

func updateSimNodeIdleChips(simNode *plugin.NPUNode, task *api.TaskInfo, release bool) {
	if simNode == nil || task == nil || task.Pod == nil || simNode.Annotation == nil {
		return
	}
	names, annoKey, _ := taskChipNames(task)
	if len(names) == 0 {
		return
	}
	val, ok := simNode.Annotation[annoKey]
	if !ok {
		return
	}
	set := make(map[string]struct{})
	for _, card := range strings.Split(val, ",") {
		if card != "" {
			set[card] = struct{}{}
		}
	}
	for _, n := range names {
		if release {
			set[n] = struct{}{}
		} else {
			delete(set, n)
		}
	}
	if len(set) == 0 {
		simNode.Annotation[annoKey] = ""
		return
	}
	cards := make([]string, 0, len(set))
	for card := range set {
		cards = append(cards, card)
	}
	sort.Strings(cards)
	simNode.Annotation[annoKey] = strings.Join(cards, ",")
}
