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

// Package multilevelscheduling for scheduling NPU job with general abstract network topology configuration.
package multilevelscheduling

import (
	"errors"
	"fmt"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// Reschedule reschedules a series of pods based on affinity
func Reschedule(resourceTree *util.ResourceTree, taskTree *util.TaskTree, faultNodes []string) (*util.TaskTree, error) {
	faultNodeMap := make(map[string]struct{}, len(faultNodes))
	for _, nodeName := range faultNodes {
		faultNodeMap[nodeName] = struct{}{}
	}
	for _, faultNode := range faultNodes {
		if _, ok := faultNodeMap[faultNode]; !ok {
			continue
		}
		klog.V(util.LogDebugLev).Infof("start to reschedule faultNode: %s", faultNode)
		faultSubTree, err := findLargestFaultSubTree(taskTree, faultNode, faultNodeMap)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("findLargestFaultSubTree failed, %v", err)
			break
		}
		klog.V(util.LogDebugLev).Infof("findLargestFaultSubTree success, faultSubTree: %v, root node: %v",
			faultSubTree, faultSubTree.TaskNode)
		if len(faultSubTree.Levels) == len(resourceTree.Levels) {
			klog.V(util.LogDebugLev).Infof("all tasks failed, start to schedule")
			return Schedule(resourceTree, taskTree.Levels)
		}
		if err := rescheduleSubTask(resourceTree, faultSubTree); err != nil {
			klog.V(util.LogErrorLev).Infof("reschedule fault subtask failed, %v", err)
			return nil, err
		}
		for _, task := range faultSubTree.GetAllBaseTasks() {
			klog.V(util.LogDebugLev).Infof("reschedule fault node %s successfully", task.ResourceNodeName)
			delete(faultNodeMap, task.ResourceNodeName)
		}
	}
	if len(faultNodeMap) == 0 {
		taskTree.ResourceLevels = resourceTree.Levels
		return taskTree, nil
	}
	return taskTree, errors.New("failed to reschedule tasks")
}

func findLargestFaultSubTree(taskTree *util.TaskTree, faultNode string, faultNodes map[string]struct{}) (*util.TaskTree, error) {
	var faultTask *util.TaskNode
	// find the taskNode of fault node
	for _, task := range taskTree.GetAllBaseTasks() {
		if task.ResourceNodeName == faultNode {
			faultTask = task
			break
		}
	}
	if faultTask == nil {
		return nil, fmt.Errorf("find fault task from node[%v] failed", faultNode)
	}
	for depth := len(taskTree.Levels) - 1; depth >= 0; depth-- {
		// root task node, return whole tree
		if faultTask.Parent == nil {
			return taskTree, nil
		}
		parentSubTree, err := taskTree.GetSubTree(faultTask.Parent)
		if err != nil {
			return nil, fmt.Errorf("find fault task [%v] parent subtree failed, %v", faultTask, err)
		}
		for _, task := range parentSubTree.GetAllBaseTasks() {
			if _, ok := faultNodes[task.ResourceNodeName]; !ok {
				// get the largest fault subtree that includes the fault node
				return taskTree.GetSubTree(faultTask)
			}
		}
		faultTask = faultTask.Parent
	}
	return nil, errors.New("find largest fault subtree failed")
}

func rescheduleSubTask(resourceTree *util.ResourceTree, faultSubTree *util.TaskTree) error {
	// find fault resource node
	parentNode, err := resourceTree.FindNodeByTask(faultSubTree.Parent)
	if err != nil {
		return fmt.Errorf("failed to find fault resource nodes, %v", err)
	}
	// find fault resource node's parent node for scheduling
	parentSubtree, err := resourceTree.GetSubTree(parentNode)
	if err != nil {
		return fmt.Errorf("failed to find fault sub tree, %v", err)
	}

	taskLevels := make([]util.TaskTreeLevel, len(faultSubTree.Levels)+1)
	copy(taskLevels[1:], faultSubTree.Levels)
	taskLevels[0].ReqNode = taskLevels[1].ReqNode
	rescheduledTaskSubTree, err := Schedule(parentSubtree, taskLevels)
	if err != nil {
		return fmt.Errorf("failed to schedule fault subtree from parent resource subtree, %v", err)
	}
	rescheduledTask, ok := rescheduledTaskSubTree.Children[0]
	if !ok {
		return errors.New("failed to find rescheduled task")
	}

	rescheduledTask.Index = faultSubTree.Index
	faultSubTree.Parent.AddOrUpdateChild(rescheduledTask)
	parentNode.RemoveBaseNodesByTask(rescheduledTask)
	return nil
}
