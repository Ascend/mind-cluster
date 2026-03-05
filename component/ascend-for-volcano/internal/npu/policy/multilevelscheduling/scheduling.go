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
	"sort"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// Schedule schedules a series of pods on a resource tree
func Schedule(resourceTree *util.ResourceTree, tasksLevels []util.TaskTreeLevel) (*util.TaskTree, error) {
	tree, err := createSchedulingTree(resourceTree, tasksLevels)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduling tree, %v", err)
	}
	taskTree, ok := tree.schedule()
	if !ok {
		return nil, errors.New("failed to schedule tasks")
	}
	taskTree.ResourceLevels = resourceTree.Levels
	return taskTree, nil
}

func createSchedulingTree(resourceTree *util.ResourceTree, taskLevels []util.TaskTreeLevel) (*schedulingTree, error) {
	if resourceTree == nil || !resourceTree.CheckNotNil() {
		return nil, errors.New("resource tree or root resource node is nil")
	}
	if len(resourceTree.Levels) != len(taskLevels) {
		return nil, errors.New("number of levels does not match the number of task levels")
	}

	tree := schedulingTree{
		root:   &schedulingTreeNode{node: resourceTree.ResourceNode, children: make(map[string]*schedulingTreeNode)},
		levels: make([]*schedulingTreeLevel, 0, len(taskLevels)),
	}
	for depth := 0; depth < len(taskLevels); depth++ {
		tree.levels = append(tree.levels, &schedulingTreeLevel{
			nodes:         tree.initNodeTopoForLevel(depth),
			taskLevel:     taskLevels[depth],
			resourceLevel: resourceTree.Levels[depth],
		})
	}

	for depth := len(tree.levels) - 1; depth >= 0; depth-- {
		for _, treeNode := range tree.levels[depth].nodes {
			tree.initNode(treeNode)
		}
	}
	return &tree, nil
}

func (t *schedulingTree) initNodeTopoForLevel(depth int) []*schedulingTreeNode {
	var allNodes []*schedulingTreeNode
	if depth == 0 {
		allNodes = append(allNodes, t.root)
		return allNodes
	}

	parentLevel := t.levels[depth-1]
	for _, treeNode := range parentLevel.nodes {
		for name, resourceNode := range treeNode.node.Children {
			if !resourceNode.CheckNotNil() {
				continue
			}
			childNode := &schedulingTreeNode{
				depth:    depth,
				node:     resourceNode,
				parent:   treeNode,
				children: make(map[string]*schedulingTreeNode),
			}
			allNodes = append(allNodes, childNode)
			treeNode.children[name] = childNode
		}
	}
	return allNodes
}

func (t *schedulingTree) initNode(treeNode *schedulingTreeNode) {
	treeNode.isReserved = false
	treeNode.hasSufficientReservedResource = true
	treeNode.hasTraversed = false
	treeNode.allocatableTaskCount = 1
	treeNode.fragmentScore = 0
	if t.isBaseNode(treeNode) {
		return
	}

	var (
		maxReservedChildren      = t.levels[treeNode.depth].resourceLevel.ReservedNode
		reservedChildrenCount    int
		accumulatedFragmentScore int
		allocatableSubTaskCount  int
	)
	for _, childTreeNode := range treeNode.children {
		if reservedChildrenCount < maxReservedChildren {
			reservedChildrenCount++
			childTreeNode.isReserved = true
		}
		// skip nodes that are reserved or have sufficient reserved resource
		if childTreeNode.isReserved || !childTreeNode.hasSufficientReservedResource {
			continue
		}
		childTreeNode.isReserved = false
		accumulatedFragmentScore += childTreeNode.fragmentScore
		allocatableSubTaskCount += childTreeNode.allocatableTaskCount
	}
	if reservedChildrenCount < maxReservedChildren {
		treeNode.hasSufficientReservedResource = false
	}
	var (
		maxNodesForChildTask = t.levels[treeNode.depth+1].taskLevel.ReqNode
		maxChildrenTaskCount = t.getMaxChildrenTaskCount(treeNode.depth)
	)

	treeNode.allocatableTaskCount = allocatableSubTaskCount / maxChildrenTaskCount
	treeNode.fragmentScore = (allocatableSubTaskCount%maxChildrenTaskCount)*maxNodesForChildTask +
		accumulatedFragmentScore*defaultFragmentScoreScale
	klog.V(util.LogDebugLev).Infof("fragment score for %s is %d", treeNode.node.Name, treeNode.fragmentScore)

}

func (t *schedulingTree) schedule() (*util.TaskTree, bool) {
	// 1. find a task tree without using any reserved resource
	klog.V(util.LogDebugLev).Infof("start to traverse resource tree from root node")
	taskTree, scheduled := t.traverseTree(t.root, 1)
	if scheduled {
		return taskTree, true
	}
	// 2. find a task tree with the reserved resource
	klog.V(util.LogDebugLev).Info("start to traverse remaining nodes")
	taskTree, scheduled = t.traverseLevelForRemainingNodes(len(t.levels) - 1)
	if scheduled {
		return taskTree, true
	}

	return nil, false
}

// traverseTree select nodes from resource tree
func (t *schedulingTree) traverseTree(treeNode *schedulingTreeNode, unscheduledTaskCount int) (*util.TaskTree, bool) {
	if t.isBaseNode(treeNode) {
		return t.buildTaskTree(treeNode)
	}

	// filter nodes that are reserved or has insufficient reserved resource
	nodeMap := make(map[string]*schedulingTreeNode)
	for name, childNode := range treeNode.children {
		if childNode.isReserved || !childNode.hasSufficientReservedResource || childNode.allocatableTaskCount <= 0 {
			continue
		}
		nodeMap[name] = childNode
	}

	unscheduledTaskCount *= t.getMaxChildrenTaskCount(treeNode.depth)
	return t.traverseSiblings(nodeMap, unscheduledTaskCount)
}

// traverseLevelForRemainingNodes select all unvisited nodes for single depth
func (t *schedulingTree) traverseLevelForRemainingNodes(depth int) (*util.TaskTree, bool) {
	var (
		parentLevel                  = t.levels[depth-1]
		childrenGroupByReservedNodes = make([][][]*schedulingTreeNode, parentLevel.resourceLevel.ReservedNode+1)
	)
	// find all remaining nodes group by their parents
	for _, parentTreeNode := range parentLevel.nodes {
		var (
			reservedNodeCount      int
			nonReservedNodes       []*schedulingTreeNode
			remainingChildrenNodes *[]*schedulingTreeNode
		)
		for _, treeNode := range parentTreeNode.children {
			if treeNode.hasTraversed {
				continue
			}
			if treeNode.isReserved {
				reservedNodeCount++
				childrenGroupByReservedNodes[reservedNodeCount] =
					append(childrenGroupByReservedNodes[reservedNodeCount], []*schedulingTreeNode{treeNode})
				groupLen := len(childrenGroupByReservedNodes[reservedNodeCount])
				remainingChildrenNodes = &childrenGroupByReservedNodes[reservedNodeCount][groupLen-1]
			} else {
				nonReservedNodes = append(nonReservedNodes, treeNode)
			}
		}
		if len(nonReservedNodes) > 0 {
			if remainingChildrenNodes == nil {
				childrenGroupByReservedNodes[reservedNodeCount] =
					append(childrenGroupByReservedNodes[reservedNodeCount], nonReservedNodes)
			} else {
				*remainingChildrenNodes = append(*remainingChildrenNodes, nonReservedNodes...)
			}
		}
	}
	// visit these nodes in order of the number of the reserved children node of the parent node
	for reservedNodeCount := len(childrenGroupByReservedNodes) - 1; reservedNodeCount >= 0; reservedNodeCount-- {
		for _, remainingChildrenNodes := range childrenGroupByReservedNodes[reservedNodeCount] {
			var (
				allocatableTaskCount int
				nodeMap              = make(map[string]*schedulingTreeNode, len(remainingChildrenNodes))
			)
			for _, treeNode := range remainingChildrenNodes {
				allocatableTaskCount += treeNode.allocatableTaskCount
				nodeMap[treeNode.node.Name] = treeNode
			}
			taskTree, scheduled := t.traverseSiblings(nodeMap, allocatableTaskCount)
			if scheduled {
				return taskTree, true
			}
		}
	}
	return nil, false
}

func (t *schedulingTree) traverseSiblings(
	siblings map[string]*schedulingTreeNode, unscheduledTaskCount int) (*util.TaskTree, bool) {
	for _, treeNode := range sortSiblings(siblings, t.compareSmallerTreeNodes) {
		if treeNode.allocatableTaskCount > unscheduledTaskCount {
			continue
		}
		delete(siblings, treeNode.node.Name)
		if taskTree, scheduled := t.traverseNode(treeNode, &unscheduledTaskCount); scheduled || unscheduledTaskCount <= 0 {
			return taskTree, scheduled
		}
	}
	for _, treeNode := range sortSiblings(siblings, t.compareBiggerTreeNodes) {
		if taskTree, scheduled := t.traverseNode(treeNode, &unscheduledTaskCount); scheduled || unscheduledTaskCount <= 0 {
			return taskTree, scheduled
		}
	}
	return nil, false
}

func (t *schedulingTree) traverseNode(treeNode *schedulingTreeNode, unscheduledTaskCount *int) (*util.TaskTree, bool) {
	klog.V(util.LogDebugLev).Infof("start to traverse node %s in level-%d", treeNode.node.Name, treeNode.depth)
	treeNode.hasTraversed = true
	allocatableTaskCount := treeNode.allocatableTaskCount
	if *unscheduledTaskCount < allocatableTaskCount {
		allocatableTaskCount = *unscheduledTaskCount
	}
	taskTree, scheduled := t.traverseTree(treeNode, allocatableTaskCount)
	if scheduled {
		return taskTree, true
	}
	*unscheduledTaskCount -= allocatableTaskCount
	return nil, false
}

func (t *schedulingTree) compareSmallerTreeNodes(left, right *schedulingTreeNode) bool {
	if left.fragmentScore != right.fragmentScore {
		return left.fragmentScore < right.fragmentScore
	}

	return left.allocatableTaskCount < right.allocatableTaskCount
}

func (t *schedulingTree) compareBiggerTreeNodes(left, right *schedulingTreeNode) bool {
	if left.allocatableTaskCount != right.allocatableTaskCount {
		return left.allocatableTaskCount < right.allocatableTaskCount
	}

	return left.fragmentScore < right.fragmentScore
}

func sortSiblings(nodeMap map[string]*schedulingTreeNode, compareFn func(left, right *schedulingTreeNode) bool,
) []*schedulingTreeNode {
	treeNodes := make([]*schedulingTreeNode, 0, len(nodeMap))
	for _, treeNode := range nodeMap {
		treeNodes = append(treeNodes, treeNode)
	}
	sort.Slice(treeNodes, func(i, j int) bool {
		return compareFn(treeNodes[i], treeNodes[j])
	})
	return treeNodes
}

func (t *schedulingTree) buildTaskTree(treeNode *schedulingTreeNode) (*util.TaskTree, bool) {
	for {
		maxSubTaskCount := t.getMaxChildrenTaskCount(treeNode.depth)
		klog.V(util.LogDebugLev).Infof("building task tree for node %s in depth %d, max subtask count: %d, current free subtasks: %d",
			treeNode.node.Name, treeNode.depth, maxSubTaskCount, len(treeNode.freeSubTasks))

		if len(treeNode.freeSubTasks) < maxSubTaskCount {
			klog.V(util.LogDebugLev).Infof("not enough free subtasks for node %s, need %d but have %d",
				treeNode.node.Name, maxSubTaskCount, len(treeNode.freeSubTasks))
			return nil, false
		}

		taskNode := &util.TaskNode{
			ResourceNodeName: treeNode.node.Name,
		}
		for i := 0; i < maxSubTaskCount; i++ {
			childNode := treeNode.freeSubTasks[i]
			childNode.Index = i
			taskNode.AddOrUpdateChild(childNode)
		}
		treeNode.freeSubTasks = treeNode.freeSubTasks[maxSubTaskCount:]

		klog.V(util.LogDebugLev).Infof(
			"allocate new task node from node %s in depth %d", treeNode.node.Name, treeNode.depth)

		if t.isRootNode(treeNode) {
			return t.createTaskTree(taskNode), true
		}
		parentTreeNode := treeNode.parent
		if parentTreeNode == nil {
			klog.V(util.LogDebugLev).Infof("parent node of %s is nil", treeNode.node.Name)
			return nil, false
		}
		parentTreeNode.freeSubTasks = append(parentTreeNode.freeSubTasks, taskNode)
		treeNode = parentTreeNode
	}
}

func (t *schedulingTree) isBaseNode(node *schedulingTreeNode) bool {
	return node.depth == len(t.levels)-1
}

func (t *schedulingTree) isRootNode(node *schedulingTreeNode) bool {
	return node.depth == 0
}

func (t *schedulingTree) getMaxChildrenTaskCount(depth int) int {
	if depth >= len(t.levels)-1 {
		return 0
	}
	return t.levels[depth].taskLevel.ReqNode / t.levels[depth+1].taskLevel.ReqNode
}

func (t *schedulingTree) createTaskTree(rootTask *util.TaskNode) *util.TaskTree {
	taskTree := &util.TaskTree{
		TaskNode:      rootTask,
		FragmentScore: t.root.fragmentScore,
		Levels:        make([]util.TaskTreeLevel, 0, len(t.levels)),
	}
	for _, level := range t.levels {
		taskTree.Levels = append(taskTree.Levels, level.taskLevel)
	}
	return taskTree
}
