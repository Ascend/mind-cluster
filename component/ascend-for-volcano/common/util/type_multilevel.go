/*
Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package util is using for the total variable.
*/
package util

import (
	"errors"
	"fmt"
	"strconv"

	"k8s.io/klog"
)

// TaskTree task node tree
type TaskTree struct {
	*TaskNode
	FragmentScore  int
	ResourceLevels []ResourceTreeLevel
	Levels         []TaskTreeLevel
}

// TaskNode a single pod or a group of pods
type TaskNode struct {
	Index            int
	Parent           *TaskNode
	Children         map[int]*TaskNode
	ResourceNodeName string
}

// TaskTreeLevel level configs for task
type TaskTreeLevel struct {
	Name    string // level1 level2 ...
	ReqNode int
}

// ResourceTree resource tree
type ResourceTree struct {
	*ResourceNode
	Name   string
	Levels []ResourceTreeLevel
}

// ResourceNode a single node or a group of nodes
type ResourceNode struct {
	Name     string
	Parent   *ResourceNode
	Children map[string]*ResourceNode
}

// ResourceTreeLevel level configs for nodes
type ResourceTreeLevel struct {
	Label        string    `json:"label,omitempty"` // equals resource resource tree name
	ReservedNode int       `json:"reservedNode,omitempty"`
	Type         LevelType `json:"-"`
}

type LevelType string

const (
	// LevelTypeTree resource node tree node
	LevelTypeTree LevelType = "tree"
	// LevelTypeMiddle resource node network hyper node
	LevelTypeMiddle LevelType = "middle"
	// LevelTypeNode resource node physic k8s node
	LevelTypeNode LevelType = "node"
)

// GetSubTree gets subTree
func (rt *ResourceTree) GetSubTree(node *ResourceNode) (*ResourceTree, error) {
	if rt == nil || rt.ResourceNode == nil || !node.CheckNotNil() || !rt.checkLevel(0) {
		return nil, errors.New("input node or resource tree is invalid")
	}

	var (
		depth       int
		currentNode = node
	)
	for currentNode != rt.ResourceNode {
		depth++
		currentNode = currentNode.Parent
		if currentNode == nil || !rt.checkLevel(depth) {
			return nil, errors.New("input node is not belonging target resource tree")
		}
	}
	return &ResourceTree{
		ResourceNode: node,
		Levels:       rt.Levels[depth:],
	}, nil
}

func (rt *ResourceTree) checkLevel(depth int) bool {
	if depth < len(rt.Levels) && depth >= 0 {
		return true
	}

	klog.V(LogErrorLev).Infof("resource node depth check failed")
	return false
}

// FindNodeByTask finds resource node in level by task node
func (rn *ResourceNode) FindNodeByTask(task *TaskNode) (*ResourceNode, error) {
	if !rn.CheckNotNil() || !task.checkNotNil() {
		return nil, errors.New("input task or resource tree is invalid")
	}

	var (
		ancestors   []string
		currentTask = task
	)
	// find resource node's ancestors by task node
	for currentTask.ResourceNodeName != rn.Name {
		ancestors = append(ancestors, currentTask.ResourceNodeName)
		currentTask = currentTask.Parent
		if currentTask == nil {
			return nil, errors.New("input task is not matched target resource tree")
		}
	}

	// find resource node by ancestors
	currentNode := rn
	for i := len(ancestors) - 1; i >= 0; i-- {
		childNode, ok := currentNode.Children[ancestors[i]]
		if !ok || !childNode.CheckNotNil() {
			return nil, errors.New("resource tree does not have any available node for task")
		}
		currentNode = childNode
	}
	if currentNode.Name != task.ResourceNodeName {
		return nil, fmt.Errorf("resource node name[%v] not match task node name[%v]",
			currentNode.Name, task.ResourceNodeName)
	}
	return currentNode, nil
}

// RemoveBaseNodesByTask removes all nodes used by task
func (rn *ResourceNode) RemoveBaseNodesByTask(task *TaskNode) {
	if !rn.CheckNotNil() || !task.checkNotNil() {
		return
	}

	targetNode, err := rn.FindNodeByTask(task)
	if err != nil {
		return
	}
	targetNode.RemoveBaseNodesByTaskRecursively(task, 0)
}

func (rn *ResourceNode) RemoveBaseNodesByTaskRecursively(task *TaskNode, iterCount int) {
	if !rn.CheckNotNil() || !task.checkNotNil() {
		return
	}
	if iterCount == MaxLevel {
		klog.V(LogErrorLev).Infof("RemoveBaseNodesByTaskRecursively error, out of max level limit %v", MaxLevel)
		return
	}

	if len(rn.Children) == 0 {
		if rn.Parent != nil {
			delete(rn.Parent.Children, task.ResourceNodeName)
		}
		return
	}
	for _, subTask := range task.Children {
		if !subTask.checkNotNil() {
			continue
		}

		childNode, ok := rn.Children[subTask.ResourceNodeName]
		if !ok {
			continue
		}
		childNode.RemoveBaseNodesByTaskRecursively(subTask, iterCount+1)
	}
}

// AddOrUpdateChild appends a child to node by name
func (rn *ResourceNode) AddOrUpdateChild(child *ResourceNode) {
	if !rn.CheckNotNil() || !child.CheckNotNil() {
		return
	}

	if rn.Children == nil {
		rn.Children = map[string]*ResourceNode{}
	}
	rn.Children[child.Name] = child
	child.Parent = rn
}

func (rn *ResourceNode) CheckNotNil() bool {
	if rn != nil {
		return true
	}
	klog.V(LogErrorLev).Infof("resource node nil check failed")
	return false
}

// GetSubTree gets subTree
func (tt *TaskTree) GetSubTree(task *TaskNode) (*TaskTree, error) {
	if tt == nil || tt.TaskNode == nil || !task.checkNotNil() || !tt.checkLevel(0) {
		return nil, errors.New("input task or task tree is invalid")
	}
	var (
		depth       int
		currentTask = task
	)
	for currentTask != tt.TaskNode {
		depth++
		currentTask = currentTask.Parent
		if currentTask == nil || !tt.checkLevel(depth) {
			return nil, errors.New("input task is not belonging target task tree")
		}
	}
	return &TaskTree{
		TaskNode: task,
		Levels:   tt.Levels[depth:],
	}, nil
}

// GetAllBaseTasks gets all base tasks
func (tt *TaskTree) GetAllBaseTasks() []*TaskNode {
	if tt == nil || tt.TaskNode == nil {
		return nil
	}

	return tt.getAllBaseTasksRecursively(nil, len(tt.Levels), 1)
}

func (tt *TaskTree) checkLevel(depth int) bool {
	if depth < len(tt.Levels) && depth >= 0 {
		return true
	}

	klog.V(LogErrorLev).Infof("task node depth check failed")
	return false
}

func (tn *TaskNode) getAllBaseTasksRecursively(tasks []*TaskNode, maxLevel, currentLevel int) []*TaskNode {
	if !tn.checkNotNil() {
		return tasks
	}

	if maxLevel == currentLevel {
		return append(tasks, tn)
	}
	for _, subTask := range tn.Children {
		tasks = subTask.getAllBaseTasksRecursively(tasks, maxLevel, currentLevel+1)
	}
	return tasks
}

func (tn *TaskNode) checkNotNil() bool {
	if tn != nil {
		return true
	}

	klog.V(LogErrorLev).Infof("task node nil check failed")
	return false
}

// AddOrUpdateChild appends a child to node by index
func (tn *TaskNode) AddOrUpdateChild(child *TaskNode) {
	if !tn.checkNotNil() || !child.checkNotNil() {
		return
	}

	if tn.Children == nil {
		tn.Children = map[int]*TaskNode{}
	}
	tn.Children[child.Index] = child
	child.Parent = tn
}

// GetTaskTreeLevels builds task tree levels
func GetTaskTreeLevels(affinityBlocks map[string]int, nupTaskNum int) ([]TaskTreeLevel, error) {
	// at least one level
	if len(affinityBlocks) == 0 {
		return nil, fmt.Errorf("affinity blocks is empty")
	}

	// Add tree level to task level configuration
	taskLevel := []TaskTreeLevel{{Name: JobLevelName, ReqNode: nupTaskNum}}
	for i := len(affinityBlocks); i >= Level1Number; i-- {
		// 1. Make sure level numbers are continuous and task level number equals affinity blocks number
		levelKey := TopoLevelPrefix + strconv.Itoa(i)
		logicUnitNum, ok := affinityBlocks[levelKey]
		if !ok {
			return nil, fmt.Errorf(
				"affinity block %s is not exist, please check if level numbers are continuous", levelKey)
		}
		// 2. Make sure level(n-1) is a multiple of level(n)
		if taskLevel[len(taskLevel)-1].ReqNode%logicUnitNum != 0 {
			return nil, fmt.Errorf("affinity block %s's value %d can not be divided by upper level's value %d",
				levelKey, logicUnitNum, taskLevel[len(taskLevel)-1].ReqNode)
		}
		taskLevel = append(taskLevel, TaskTreeLevel{
			Name:    levelKey,
			ReqNode: logicUnitNum,
		})
	}
	taskLevel = append(taskLevel, TaskTreeLevel{Name: NodeLevelName, ReqNode: 1})

	return taskLevel, nil
}
