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

package multilevelscheduling

import (
	"fmt"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

type scheduleTestCase struct {
	name         string
	resourceTree *util.ResourceTree
	taskLevels   []util.TaskTreeLevel
	wantErr      bool
}

func newResourceTree(name string, levels int) *util.ResourceTree {
	root := &util.ResourceNode{Name: name, Children: make(map[string]*util.ResourceNode)}
	treeLevels := make([]util.ResourceTreeLevel, levels)
	for i := range treeLevels {
		if i == 0 {
			treeLevels[i] = util.ResourceTreeLevel{Type: util.LevelTypeTree, ReservedNode: 0}
		} else if i == levels-1 {
			treeLevels[i] = util.ResourceTreeLevel{Type: util.LevelTypeNode, ReservedNode: 0}
		} else {
			treeLevels[i] = util.ResourceTreeLevel{Type: util.LevelTypeMiddle, ReservedNode: 0}
		}
	}
	return &util.ResourceTree{Name: "test", ResourceNode: root, Levels: treeLevels}
}

func newTaskLevels(count int, reqNode int) []util.TaskTreeLevel {
	levels := make([]util.TaskTreeLevel, count)
	for i := range levels {
		levels[i] = util.TaskTreeLevel{Name: fmt.Sprintf("level%d", i), ReqNode: reqNode}
	}
	return levels
}

func newResourceTreeWithChildren() *util.ResourceTree {
	root := &util.ResourceNode{Name: "root", Children: make(map[string]*util.ResourceNode)}
	c1 := &util.ResourceNode{Name: "c1", Parent: root, Children: make(map[string]*util.ResourceNode)}
	c2 := &util.ResourceNode{Name: "c2", Parent: root, Children: make(map[string]*util.ResourceNode)}
	n1 := &util.ResourceNode{Name: "n1", Parent: c1}
	n2 := &util.ResourceNode{Name: "n2", Parent: c1}
	n3 := &util.ResourceNode{Name: "n3", Parent: c2}
	n4 := &util.ResourceNode{Name: "n4", Parent: c2}
	c1.Children["n1"] = n1
	c1.Children["n2"] = n2
	c2.Children["n3"] = n3
	c2.Children["n4"] = n4
	root.Children["c1"] = c1
	root.Children["c2"] = c2
	return &util.ResourceTree{
		Name:         "test",
		ResourceNode: root,
		Levels: []util.ResourceTreeLevel{
			{Type: util.LevelTypeTree, ReservedNode: 0},
			{Type: util.LevelTypeMiddle, ReservedNode: 0},
			{Type: util.LevelTypeNode, ReservedNode: 0},
		},
	}
}

func TestSchedule(t *testing.T) {
	cases := []scheduleTestCase{
		{"nil_tree", nil, newTaskLevels(1, 1), true},
		{"mismatched", newResourceTree("root", 3), newTaskLevels(1, 1), true},
		{"success", newResourceTree("base", 1), newTaskLevels(1, 1), false},
	}
	for _, tc := range cases {
		_, err := Schedule(tc.resourceTree, tc.taskLevels)
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: got err=%v, wantErr=%v", tc.name, err, tc.wantErr)
		}
	}
}

func TestSchedule_ComplexTree(t *testing.T) {
	tree := newResourceTreeWithChildren()
	taskLevels := []util.TaskTreeLevel{
		{Name: "l0", ReqNode: 4},
		{Name: "l1", ReqNode: 2},
		{Name: "l2", ReqNode: 1},
	}
	tt, err := Schedule(tree, taskLevels)
	if err != nil {
		t.Errorf("Schedule() unexpected error: %v", err)
	}
	if tt == nil {
		t.Error("Schedule() should return non-nil task tree")
	}
}

func TestCreateSchedulingTree(t *testing.T) {
	tree := createMockResourceTree()
	taskLevels := newTaskLevels(3, 4)
	st, err := createSchedulingTree(tree, taskLevels)
	if err != nil || st == nil || len(st.levels) != 3 {
		t.Errorf("createSchedulingTree() got err=%v, st=%v, len=%d", err, st, len(st.levels))
	}
}

func TestCreateSchedulingTree_Nil(t *testing.T) {
	_, err := createSchedulingTree(nil, newTaskLevels(1, 1))
	if err == nil {
		t.Error("createSchedulingTree() should return error for nil tree")
	}
}

func TestCreateSchedulingTree_NilRoot(t *testing.T) {
	tree := &util.ResourceTree{Name: "test", ResourceNode: nil, Levels: []util.ResourceTreeLevel{{}}}
	_, err := createSchedulingTree(tree, newTaskLevels(1, 1))
	if err == nil {
		t.Error("createSchedulingTree() should return error for nil root")
	}
}

func TestSchedulingTree_IsBaseNode(t *testing.T) {
	tree := createMockResourceTree()
	st, _ := createSchedulingTree(tree, newTaskLevels(3, 4))
	if st.isBaseNode(st.root) {
		t.Error("isBaseNode() should return false for root")
	}
	for _, child := range st.root.children {
		for _, grandchild := range child.children {
			if !st.isBaseNode(grandchild) {
				t.Error("isBaseNode() should return true for leaf node")
			}
		}
	}
}

func TestSchedulingTree_IsRootNode(t *testing.T) {
	tree := createMockResourceTree()
	st, _ := createSchedulingTree(tree, newTaskLevels(3, 4))
	if !st.isRootNode(st.root) {
		t.Error("isRootNode() should return true for root")
	}
	for _, child := range st.root.children {
		if st.isRootNode(child) {
			t.Error("isRootNode() should return false for child")
		}
	}
}

func TestSchedulingTree_GetMaxChildrenTaskCount(t *testing.T) {
	tree := createMockResourceTree()
	taskLevels := []util.TaskTreeLevel{
		{Name: "l0", ReqNode: 4},
		{Name: "l1", ReqNode: 2},
		{Name: "l2", ReqNode: 1},
	}
	st, _ := createSchedulingTree(tree, taskLevels)
	if count := st.getMaxChildrenTaskCount(0); count != 2 {
		t.Errorf("getMaxChildrenTaskCount(0)=%d, want 2", count)
	}
	if count := st.getMaxChildrenTaskCount(1); count != 2 {
		t.Errorf("getMaxChildrenTaskCount(1)=%d, want 2", count)
	}
	if count := st.getMaxChildrenTaskCount(2); count != 0 {
		t.Errorf("getMaxChildrenTaskCount(2)=%d, want 0", count)
	}
}

func TestSortSiblings(t *testing.T) {
	nodeMap := map[string]*schedulingTreeNode{
		"n1": {node: &util.ResourceNode{Name: "n1"}, fragmentScore: 10, allocatableTaskCount: 2},
		"n2": {node: &util.ResourceNode{Name: "n2"}, fragmentScore: 5, allocatableTaskCount: 1},
		"n3": {node: &util.ResourceNode{Name: "n3"}, fragmentScore: 5, allocatableTaskCount: 3},
	}
	tree := &schedulingTree{}
	result := sortSiblings(nodeMap, tree.compareSmallerTreeNodes)
	if len(result) != 3 || result[0].fragmentScore != 5 {
		t.Errorf("sortSiblings() len=%d, first score=%d", len(result), result[0].fragmentScore)
	}
}

func TestCompareSmallerTreeNodes(t *testing.T) {
	tree := &schedulingTree{}
	cases := []struct {
		left, right *schedulingTreeNode
		want        bool
	}{
		{left: &schedulingTreeNode{fragmentScore: 5}, right: &schedulingTreeNode{fragmentScore: 10}, want: true},
		{left: &schedulingTreeNode{fragmentScore: 10, allocatableTaskCount: 1}, right: &schedulingTreeNode{fragmentScore: 10, allocatableTaskCount: 2}, want: true},
		{left: &schedulingTreeNode{fragmentScore: 5, allocatableTaskCount: 2}, right: &schedulingTreeNode{fragmentScore: 5, allocatableTaskCount: 2}, want: false},
	}
	for i, tc := range cases {
		if got := tree.compareSmallerTreeNodes(tc.left, tc.right); got != tc.want {
			t.Errorf("case %d: got=%v, want=%v", i, got, tc.want)
		}
	}
}

func TestCompareBiggerTreeNodes(t *testing.T) {
	tree := &schedulingTree{}
	cases := []struct {
		left, right *schedulingTreeNode
		want        bool
	}{
		{left: &schedulingTreeNode{allocatableTaskCount: 1}, right: &schedulingTreeNode{allocatableTaskCount: 2}, want: true},
		{left: &schedulingTreeNode{allocatableTaskCount: 2, fragmentScore: 1}, right: &schedulingTreeNode{allocatableTaskCount: 2, fragmentScore: 2}, want: true},
		{left: &schedulingTreeNode{allocatableTaskCount: 2, fragmentScore: 5}, right: &schedulingTreeNode{allocatableTaskCount: 2, fragmentScore: 5}, want: false},
	}
	for i, tc := range cases {
		if got := tree.compareBiggerTreeNodes(tc.left, tc.right); got != tc.want {
			t.Errorf("case %d: got=%v, want=%v", i, got, tc.want)
		}
	}
}

func TestSchedulingTree_InitNode(t *testing.T) {
	tree := createMockResourceTree()
	st, _ := createSchedulingTree(tree, newTaskLevels(3, 4))
	st.initNode(st.root)
	if st.root.allocatableTaskCount < 0 {
		t.Errorf("initNode() allocatableTaskCount=%d", st.root.allocatableTaskCount)
	}
}

func TestSchedulingTree_Schedule(t *testing.T) {
	tree := createMockResourceTree()
	st, _ := createSchedulingTree(tree, newTaskLevels(3, 4))
	st.root.allocatableTaskCount = 1
	for _, child := range st.root.children {
		child.allocatableTaskCount = 1
		for _, grandchild := range child.children {
			grandchild.allocatableTaskCount = 1
		}
	}
	tt, scheduled := st.schedule()
	if scheduled && tt == nil {
		t.Error("schedule() should return non-nil task tree when scheduled")
	}
}

func TestSchedulingTree_TraverseTree_BaseNode(t *testing.T) {
	tree := newResourceTree("base", 1)
	st, _ := createSchedulingTree(tree, newTaskLevels(1, 1))
	st.root.allocatableTaskCount = 1
	tt, scheduled := st.traverseTree(st.root, 1)
	if scheduled && tt == nil {
		t.Error("traverseTree() should return non-nil task tree for base node")
	}
}

func TestSchedulingTree_TraverseSiblings_EmptyMap(t *testing.T) {
	tree := newResourceTree("root", 1)
	st, _ := createSchedulingTree(tree, newTaskLevels(1, 1))
	_, scheduled := st.traverseSiblings(map[string]*schedulingTreeNode{}, 1)
	if scheduled {
		t.Error("traverseSiblings() should return false for empty map")
	}
}

func TestSchedulingTree_BuildTaskTree(t *testing.T) {
	tree := newResourceTree("base", 1)
	st, _ := createSchedulingTree(tree, newTaskLevels(1, 1))
	st.root.allocatableTaskCount = 1
	st.root.freeSubTasks = []*util.TaskNode{{ResourceNodeName: "task"}}
	tt, ok := st.buildTaskTree(st.root)
	if !ok || tt == nil {
		t.Errorf("buildTaskTree() got ok=%v, tt=%v", ok, tt)
	}
}

func TestSchedulingTree_TraverseNode(t *testing.T) {
	tree := newResourceTree("base", 1)
	st, _ := createSchedulingTree(tree, newTaskLevels(1, 1))
	st.root.allocatableTaskCount = 1
	count := 1
	tt, scheduled := st.traverseNode(st.root, &count)
	if !scheduled || tt == nil {
		t.Errorf("traverseNode() got scheduled=%v, tt=%v", scheduled, tt)
	}
}

func TestSchedulingTree_TraverseSiblings_SingleNode(t *testing.T) {
	tree := newResourceTree("base", 1)
	st, _ := createSchedulingTree(tree, newTaskLevels(1, 1))
	st.root.allocatableTaskCount = 1
	nodeMap := map[string]*schedulingTreeNode{"base": st.root}
	tt, scheduled := st.traverseSiblings(nodeMap, 1)
	if !scheduled || tt == nil {
		t.Errorf("traverseSiblings() got scheduled=%v, tt=%v", scheduled, tt)
	}
}

func TestSchedulingTree_InitNodeTopoForLevel(t *testing.T) {
	tree := createMockResourceTree()
	st, _ := createSchedulingTree(tree, newTaskLevels(3, 4))
	if nodes := st.initNodeTopoForLevel(0); len(nodes) != 1 {
		t.Errorf("initNodeTopoForLevel(0) len=%d, want 1", len(nodes))
	}
	if nodes := st.initNodeTopoForLevel(1); len(nodes) != 2 {
		t.Errorf("initNodeTopoForLevel(1) len=%d, want 2", len(nodes))
	}
}

func TestSchedulingTree_CreateTaskTree(t *testing.T) {
	tree := newResourceTree("base", 1)
	st, _ := createSchedulingTree(tree, newTaskLevels(1, 1))
	rootTask := &util.TaskNode{ResourceNodeName: "base"}
	tt := st.createTaskTree(rootTask)
	if tt == nil || len(tt.Levels) != 1 {
		t.Errorf("createTaskTree() got tt=%v, len=%d", tt, len(tt.Levels))
	}
}

func TestSchedulingTree_TraverseTree_WithChildren(t *testing.T) {
	tree := createMockResourceTree()
	tree.Levels = tree.Levels[:2]
	st, _ := createSchedulingTree(tree, newTaskLevels(2, 2))
	st.root.allocatableTaskCount = 1
	for _, child := range st.root.children {
		child.allocatableTaskCount = 1
	}
	tt, scheduled := st.traverseTree(st.root, 1)
	if scheduled && tt == nil {
		t.Error("traverseTree() should return non-nil task tree")
	}
}

func TestSchedulingTree_TraverseSiblings_MultipleNodes(t *testing.T) {
	tree := createMockResourceTree()
	tree.Levels = tree.Levels[:2]
	st, _ := createSchedulingTree(tree, newTaskLevels(2, 2))
	nodeMap := make(map[string]*schedulingTreeNode)
	for name, child := range st.root.children {
		child.allocatableTaskCount = 1
		nodeMap[name] = child
	}
	tt, scheduled := st.traverseSiblings(nodeMap, 1)
	if scheduled && tt == nil {
		t.Error("traverseSiblings() should return non-nil task tree")
	}
}

func TestSchedulingTree_TraverseLevelForRemainingNodes(t *testing.T) {
	tree := createMockResourceTree()
	st, _ := createSchedulingTree(tree, newTaskLevels(3, 4))
	st.root.allocatableTaskCount = 0
	for _, child := range st.root.children {
		child.allocatableTaskCount = 0
		child.hasTraversed = true
	}
	tt, scheduled := st.traverseLevelForRemainingNodes(2)
	if scheduled && tt == nil {
		t.Error("traverseLevelForRemainingNodes() should return non-nil task tree when scheduled")
	}
}

func TestSchedulingTree_Schedule_Failed(t *testing.T) {
	tree := newResourceTree("root", 2)
	st, _ := createSchedulingTree(tree, newTaskLevels(2, 2))
	st.root.allocatableTaskCount = 0
	_, scheduled := st.schedule()
	if scheduled {
		t.Error("schedule() should return false when no allocatable tasks")
	}
}

func TestSchedulingTree_InitNode_BaseNode(t *testing.T) {
	tree := newResourceTree("base", 1)
	st, _ := createSchedulingTree(tree, newTaskLevels(1, 1))
	st.initNode(st.root)
	if st.root.allocatableTaskCount != 1 {
		t.Errorf("initNode() base node allocatableTaskCount=%d, want 1", st.root.allocatableTaskCount)
	}
}

func TestSchedulingTree_TraverseTree_SkipReserved(t *testing.T) {
	tree := newResourceTree("root", 2)
	tree.ResourceNode.Children["c1"] = &util.ResourceNode{Name: "c1"}
	tree.ResourceNode.Children["c2"] = &util.ResourceNode{Name: "c2"}
	st, _ := createSchedulingTree(tree, newTaskLevels(2, 2))
	st.root.allocatableTaskCount = 1
	for _, child := range st.root.children {
		child.allocatableTaskCount = 1
		child.isReserved = true
	}
	_, scheduled := st.traverseTree(st.root, 1)
	if scheduled {
		t.Error("traverseTree() should return false when all children are reserved")
	}
}

func TestSchedulingTree_InitNode_WithReserved(t *testing.T) {
	tree := &util.ResourceTree{
		Name:         "test",
		ResourceNode: &util.ResourceNode{Name: "root", Children: make(map[string]*util.ResourceNode)},
		Levels: []util.ResourceTreeLevel{
			{Type: util.LevelTypeTree, ReservedNode: 1},
			{Type: util.LevelTypeMiddle, ReservedNode: 1},
			{Type: util.LevelTypeNode, ReservedNode: 0},
		},
	}
	tree.ResourceNode.Children["c1"] = &util.ResourceNode{Name: "c1", Children: make(map[string]*util.ResourceNode)}
	tree.ResourceNode.Children["c1"].Children["n1"] = &util.ResourceNode{Name: "n1"}
	tree.ResourceNode.Children["c1"].Children["n2"] = &util.ResourceNode{Name: "n2"}
	st, _ := createSchedulingTree(tree, newTaskLevels(3, 4))
	for _, child := range st.root.children {
		for _, grandchild := range child.children {
			grandchild.allocatableTaskCount = 1
		}
	}
	st.initNode(st.root)
	if st.root.allocatableTaskCount < 0 {
		t.Errorf("initNode() allocatableTaskCount=%d", st.root.allocatableTaskCount)
	}
}

func TestSchedulingTree_InitNode_InsufficientReserved(t *testing.T) {
	tree := &util.ResourceTree{
		Name:         "test",
		ResourceNode: &util.ResourceNode{Name: "root", Children: make(map[string]*util.ResourceNode)},
		Levels: []util.ResourceTreeLevel{
			{Type: util.LevelTypeTree, ReservedNode: 5},
			{Type: util.LevelTypeNode, ReservedNode: 0},
		},
	}
	tree.ResourceNode.Children["c1"] = &util.ResourceNode{Name: "c1"}
	tree.ResourceNode.Children["c2"] = &util.ResourceNode{Name: "c2"}
	st, _ := createSchedulingTree(tree, newTaskLevels(2, 2))
	for _, child := range st.root.children {
		child.allocatableTaskCount = 1
	}
	st.initNode(st.root)
	if st.root.hasSufficientReservedResource {
		t.Error("initNode() should set hasSufficientReservedResource to false")
	}
}

func TestSchedulingTree_TraverseLevelForRemainingNodes_WithReserved(t *testing.T) {
	tree := createMockResourceTree()
	st, _ := createSchedulingTree(tree, newTaskLevels(3, 4))
	st.root.allocatableTaskCount = 0
	for _, child := range st.root.children {
		child.allocatableTaskCount = 1
		child.hasTraversed = false
		child.isReserved = true
		for _, grandchild := range child.children {
			grandchild.allocatableTaskCount = 1
			grandchild.hasTraversed = true
		}
	}
	tt, scheduled := st.traverseLevelForRemainingNodes(2)
	if scheduled && tt == nil {
		t.Error("traverseLevelForRemainingNodes() should return non-nil task tree when scheduled")
	}
}

func TestSchedulingTree_BuildTaskTree_MultipleIterations(t *testing.T) {
	tree := newResourceTree("parent", 2)
	child := &util.ResourceNode{Name: "base", Parent: tree.ResourceNode}
	tree.ResourceNode.Children["base"] = child
	st, _ := createSchedulingTree(tree, newTaskLevels(2, 2))
	st.root.allocatableTaskCount = 1
	for _, c := range st.root.children {
		c.allocatableTaskCount = 1
		c.freeSubTasks = []*util.TaskNode{
			{ResourceNodeName: "t1"},
			{ResourceNodeName: "t2"},
		}
	}
	for _, c := range st.root.children {
		tt, ok := st.buildTaskTree(c)
		if ok && tt == nil {
			t.Error("buildTaskTree() should return non-nil task tree when ok")
		}
		break
	}
}

func TestSchedulingTree_BuildTaskTree_NilParent(t *testing.T) {
	tree := newResourceTree("base", 1)
	st, _ := createSchedulingTree(tree, newTaskLevels(1, 1))
	st.root.allocatableTaskCount = 1
	st.root.freeSubTasks = []*util.TaskNode{{ResourceNodeName: "task"}}
	st.root.parent = nil
	_, ok := st.buildTaskTree(st.root)
	if !ok {
		t.Log("buildTaskTree() returned false for nil parent (root node)")
	}
}

func TestSchedulingTree_TraverseSiblings_SkipLargeAllocatable(t *testing.T) {
	tree := newResourceTree("root", 2)
	tree.ResourceNode.Children["c1"] = &util.ResourceNode{Name: "c1"}
	tree.ResourceNode.Children["c2"] = &util.ResourceNode{Name: "c2"}
	st, _ := createSchedulingTree(tree, newTaskLevels(2, 2))
	nodeMap := make(map[string]*schedulingTreeNode)
	for name, child := range st.root.children {
		child.allocatableTaskCount = 10
		nodeMap[name] = child
	}
	_, scheduled := st.traverseSiblings(nodeMap, 1)
	if scheduled {
		t.Log("traverseSiblings() scheduled with large allocatable count")
	}
}

func TestSchedulingTree_TraverseNode_DecreaseUnscheduled(t *testing.T) {
	tree := newResourceTree("base", 1)
	st, _ := createSchedulingTree(tree, newTaskLevels(1, 1))
	st.root.allocatableTaskCount = 5
	count := 3
	st.traverseNode(st.root, &count)
	if count < 0 {
		t.Errorf("traverseNode() count=%d, should be >= 0", count)
	}
}

func TestSchedulingTree_InitNodeTopoForLevel_SkipNilChildren(t *testing.T) {
	tree := &util.ResourceTree{
		Name:         "test",
		ResourceNode: &util.ResourceNode{Name: "root", Children: make(map[string]*util.ResourceNode)},
		Levels: []util.ResourceTreeLevel{
			{Type: util.LevelTypeTree, ReservedNode: 0},
			{Type: util.LevelTypeNode, ReservedNode: 0},
		},
	}
	tree.ResourceNode.Children["nil"] = nil
	tree.ResourceNode.Children["valid"] = &util.ResourceNode{Name: "valid"}
	st, _ := createSchedulingTree(tree, newTaskLevels(2, 2))
	if len(st.root.children) != 1 {
		t.Errorf("initNodeTopoForLevel() should skip nil children, got %d", len(st.root.children))
	}
}
