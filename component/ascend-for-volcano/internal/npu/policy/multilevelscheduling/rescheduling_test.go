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
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

type rescheduleTestCase struct {
	name         string
	resourceTree *util.ResourceTree
	taskTree     *util.TaskTree
	faultNodes   []string
	wantErr      bool
}

func createMockResourceTree() *util.ResourceTree {
	root := &util.ResourceNode{Name: "root", Children: make(map[string]*util.ResourceNode)}
	child1 := &util.ResourceNode{Name: "child1", Parent: root, Children: make(map[string]*util.ResourceNode)}
	child2 := &util.ResourceNode{Name: "child2", Parent: root, Children: make(map[string]*util.ResourceNode)}
	root.Children["child1"] = child1
	root.Children["child2"] = child2
	return &util.ResourceTree{
		Name:         "test-tree",
		ResourceNode: root,
		Levels: []util.ResourceTreeLevel{
			{Type: util.LevelTypeTree, ReservedNode: 0},
			{Type: util.LevelTypeMiddle, ReservedNode: 0},
			{Type: util.LevelTypeNode, ReservedNode: 0},
		},
	}
}

func createTaskTree(rootName string, children ...*util.TaskNode) *util.TaskTree {
	root := &util.TaskNode{ResourceNodeName: rootName, Children: make(map[int]*util.TaskNode)}
	for i, child := range children {
		child.Parent = root
		root.Children[i] = child
	}
	return &util.TaskTree{TaskNode: root, Levels: []util.TaskTreeLevel{{Name: "l0", ReqNode: 1}}}
}

func createTaskNode(name string) *util.TaskNode {
	return &util.TaskNode{ResourceNodeName: name, Children: make(map[int]*util.TaskNode)}
}

func createDeepTaskTree() *util.TaskTree {
	grandchild1 := createTaskNode("gc1")
	grandchild2 := createTaskNode("gc2")
	child1 := createTaskNode("c1")
	child2 := createTaskNode("c2")
	root := createTaskNode("root")
	child1.Children[0] = grandchild1
	child1.Children[1] = grandchild2
	grandchild1.Parent = child1
	grandchild2.Parent = child1
	root.Children[0] = child1
	root.Children[1] = child2
	child1.Parent = root
	child2.Parent = root
	return &util.TaskTree{TaskNode: root, Levels: []util.TaskTreeLevel{
		{Name: "l0", ReqNode: 2}, {Name: "l1", ReqNode: 2}, {Name: "l2", ReqNode: 1}}}
}

func TestReschedule(t *testing.T) {
	cases := []rescheduleTestCase{
		{"nil_resource", nil, createTaskTree("root"), []string{"n0"}, true},
		{"nil_task", createMockResourceTree(), nil, []string{"n0"}, true},
		{"empty_fault", createMockResourceTree(), createTaskTree("root"), []string{}, false},
	}
	for _, tc := range cases {
		_, err := Reschedule(tc.resourceTree, tc.taskTree, tc.faultNodes)
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: got err=%v, wantErr=%v", tc.name, err, tc.wantErr)
		}
	}
}

func TestReschedule_AllTasksFailed(t *testing.T) {
	tree := createMockResourceTree()
	root := createTaskNode("root")
	child1 := createTaskNode("child1")
	child2 := createTaskNode("child2")
	gc1 := createTaskNode("gc1")
	gc2 := createTaskNode("gc2")
	root.Children[0] = child1
	root.Children[1] = child2
	child1.Parent = root
	child2.Parent = root
	child1.Children[0] = gc1
	child2.Children[0] = gc2
	gc1.Parent = child1
	gc2.Parent = child2
	tt := &util.TaskTree{TaskNode: root, Levels: []util.TaskTreeLevel{
		{Name: "l0", ReqNode: 1}, {Name: "l1", ReqNode: 1}, {Name: "l2", ReqNode: 1}}}
	patch := gomonkey.ApplyFunc(Schedule, func(*util.ResourceTree, []util.TaskTreeLevel) (*util.TaskTree, error) {
		return &util.TaskTree{TaskNode: &util.TaskNode{}}, nil
	})
	defer patch.Reset()
	_, err := Reschedule(tree, tt, []string{"gc1", "gc2"})
	if err != nil {
		t.Errorf("Reschedule() unexpected error: %v", err)
	}
}

func TestReschedule_SubTaskFailed(t *testing.T) {
	tree := createMockResourceTree()
	tt := createDeepTaskTree()
	patch := gomonkey.ApplyFunc(Schedule, func(*util.ResourceTree, []util.TaskTreeLevel) (*util.TaskTree, error) {
		return &util.TaskTree{TaskNode: &util.TaskNode{}}, nil
	}).ApplyMethod(reflect.TypeOf(&util.ResourceTree{}), "FindNodeByTask",
		func(*util.ResourceTree, *util.TaskNode) (*util.ResourceNode, error) {
			return &util.ResourceNode{Name: "parent", Children: make(map[string]*util.ResourceNode)}, nil
		}).ApplyMethod(reflect.TypeOf(&util.ResourceTree{}), "GetSubTree",
		func(*util.ResourceTree, *util.ResourceNode) (*util.ResourceTree, error) {
			return tree, nil
		})
	defer patch.Reset()
	_, err := Reschedule(tree, tt, []string{"gc1"})
	if err != nil {
		t.Logf("Reschedule() returned error: %v", err)
	}
}

func TestReschedule_RescheduleSubTaskFailed(t *testing.T) {
	tree := createMockResourceTree()
	tt := createDeepTaskTree()
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&util.ResourceTree{}), "FindNodeByTask",
		func(*util.ResourceTree, *util.TaskNode) (*util.ResourceNode, error) {
			return nil, errors.New("mock error")
		})
	defer patch.Reset()
	_, err := Reschedule(tree, tt, []string{"gc1"})
	if err == nil {
		t.Error("Reschedule() should return error when rescheduleSubTask failed")
	}
}

func TestFindLargestFaultSubTree(t *testing.T) {
	child := createTaskNode("child")
	root := createTaskNode("root")
	root.Children[0] = child
	child.Parent = root
	tt := &util.TaskTree{TaskNode: root, Levels: []util.TaskTreeLevel{{Name: "l0", ReqNode: 1}, {Name: "l1", ReqNode: 1}}}
	cases := []struct {
		name       string
		tree       *util.TaskTree
		faultNode  string
		faultNodes map[string]struct{}
		wantErr    bool
	}{
		{"nil_tree", nil, "n0", map[string]struct{}{"n0": {}}, true},
		{"not_found", createTaskTree("root"), "missing", map[string]struct{}{"missing": {}}, true},
		{"child", tt, "child", map[string]struct{}{"child": {}}, false},
	}
	for _, tc := range cases {
		_, err := findLargestFaultSubTree(tc.tree, tc.faultNode, tc.faultNodes)
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: got err=%v, wantErr=%v", tc.name, err, tc.wantErr)
		}
	}
}

func TestFindLargestFaultSubTree_Root(t *testing.T) {
	child1 := createTaskNode("child1")
	child2 := createTaskNode("child2")
	root := createTaskNode("root")
	root.Children[0] = child1
	root.Children[1] = child2
	child1.Parent = root
	child2.Parent = root
	tt := &util.TaskTree{TaskNode: root, Levels: []util.TaskTreeLevel{{Name: "l0", ReqNode: 2}, {Name: "l1", ReqNode: 1}}}
	faultNodes := map[string]struct{}{"child1": {}, "child2": {}}
	result, err := findLargestFaultSubTree(tt, "child1", faultNodes)
	if err != nil || result == nil {
		t.Errorf("findLargestFaultSubTree() got err=%v, result=%v", err, result)
	}
}

func TestFindLargestFaultSubTree_Parent(t *testing.T) {
	child1 := createTaskNode("child1")
	child2 := createTaskNode("child2")
	root := createTaskNode("root")
	root.Children[0] = child1
	root.Children[1] = child2
	child1.Parent = root
	child2.Parent = root
	tt := &util.TaskTree{TaskNode: root, Levels: []util.TaskTreeLevel{{Name: "l0", ReqNode: 2}, {Name: "l1", ReqNode: 1}}}
	faultNodes := map[string]struct{}{"child1": {}}
	result, err := findLargestFaultSubTree(tt, "child1", faultNodes)
	if err != nil || result == nil {
		t.Errorf("findLargestFaultSubTree() got err=%v, result=%v", err, result)
	}
}

func TestFindLargestFaultSubTree_AllChildrenFault(t *testing.T) {
	child1 := createTaskNode("child1")
	child2 := createTaskNode("child2")
	root := createTaskNode("root")
	root.Children[0] = child1
	root.Children[1] = child2
	child1.Parent = root
	child2.Parent = root
	tt := &util.TaskTree{TaskNode: root, Levels: []util.TaskTreeLevel{{Name: "l0", ReqNode: 2}, {Name: "l1", ReqNode: 1}}}
	faultNodes := map[string]struct{}{"child1": {}, "child2": {}}
	result, err := findLargestFaultSubTree(tt, "child1", faultNodes)
	if err != nil || result == nil {
		t.Errorf("findLargestFaultSubTree() got err=%v, result=%v", err, result)
	}
}

func TestFindLargestFaultSubTree_DeepTree(t *testing.T) {
	grandchild := createTaskNode("grandchild")
	child := createTaskNode("child")
	root := createTaskNode("root")
	child.Children[0] = grandchild
	grandchild.Parent = child
	root.Children[0] = child
	child.Parent = root
	tt := &util.TaskTree{TaskNode: root, Levels: []util.TaskTreeLevel{
		{Name: "l0", ReqNode: 1}, {Name: "l1", ReqNode: 1}, {Name: "l2", ReqNode: 1}}}
	faultNodes := map[string]struct{}{"grandchild": {}}
	result, err := findLargestFaultSubTree(tt, "grandchild", faultNodes)
	if err != nil || result == nil {
		t.Errorf("findLargestFaultSubTree() got err=%v, result=%v", err, result)
	}
}

func TestFindLargestFaultSubTree_GetSubTreeFailed(t *testing.T) {
	child := createTaskNode("child")
	root := createTaskNode("root")
	root.Children[0] = child
	child.Parent = root
	tt := &util.TaskTree{TaskNode: root, Levels: []util.TaskTreeLevel{{Name: "l0", ReqNode: 1}, {Name: "l1", ReqNode: 1}}}
	faultNodes := map[string]struct{}{"child": {}, "other": {}}
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&util.TaskTree{}), "GetSubTree",
		func(*util.TaskTree, *util.TaskNode) (*util.TaskTree, error) {
			return nil, errors.New("mock error")
		})
	defer patch.Reset()
	_, err := findLargestFaultSubTree(tt, "child", faultNodes)
	if err == nil {
		t.Error("findLargestFaultSubTree() should return error when GetSubTree failed")
	}
}

func TestRescheduleSubTask(t *testing.T) {
	tree := createMockResourceTree()
	faultSubTree := &util.TaskTree{
		TaskNode: &util.TaskNode{ResourceNodeName: "child1", Children: make(map[int]*util.TaskNode)},
		Levels:   []util.TaskTreeLevel{{Name: "l0", ReqNode: 1}},
	}
	faultSubTree.Parent = &util.TaskNode{ResourceNodeName: "root", Children: make(map[int]*util.TaskNode)}
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&util.ResourceTree{}), "FindNodeByTask",
		func(*util.ResourceTree, *util.TaskNode) (*util.ResourceNode, error) {
			return &util.ResourceNode{Name: "parent", Children: make(map[string]*util.ResourceNode)}, nil
		}).ApplyMethod(reflect.TypeOf(&util.ResourceTree{}), "GetSubTree",
		func(*util.ResourceTree, *util.ResourceNode) (*util.ResourceTree, error) {
			return tree, nil
		}).ApplyFunc(Schedule, func(*util.ResourceTree, []util.TaskTreeLevel) (*util.TaskTree, error) {
		return &util.TaskTree{TaskNode: &util.TaskNode{Children: map[int]*util.TaskNode{0: {ResourceNodeName: "new"}}}}, nil
	})
	defer patch.Reset()
	err := rescheduleSubTask(tree, faultSubTree)
	if err != nil {
		t.Errorf("rescheduleSubTask() unexpected error: %v", err)
	}
}

func TestRescheduleSubTask_FindNodeFailed(t *testing.T) {
	tree := createMockResourceTree()
	faultSubTree := &util.TaskTree{
		TaskNode: &util.TaskNode{ResourceNodeName: "child1"},
		Levels:   []util.TaskTreeLevel{{Name: "l0", ReqNode: 1}},
	}
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&util.ResourceTree{}), "FindNodeByTask",
		func(*util.ResourceTree, *util.TaskNode) (*util.ResourceNode, error) {
			return nil, errors.New("mock error")
		})
	defer patch.Reset()
	err := rescheduleSubTask(tree, faultSubTree)
	if err == nil {
		t.Error("rescheduleSubTask() should return error when FindNodeByTask failed")
	}
}

func TestRescheduleSubTask_GetSubTreeFailed(t *testing.T) {
	tree := createMockResourceTree()
	faultSubTree := &util.TaskTree{
		TaskNode: &util.TaskNode{ResourceNodeName: "child1"},
		Levels:   []util.TaskTreeLevel{{Name: "l0", ReqNode: 1}},
	}
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&util.ResourceTree{}), "FindNodeByTask",
		func(*util.ResourceTree, *util.TaskNode) (*util.ResourceNode, error) {
			return &util.ResourceNode{Name: "parent"}, nil
		}).ApplyMethod(reflect.TypeOf(&util.ResourceTree{}), "GetSubTree",
		func(*util.ResourceTree, *util.ResourceNode) (*util.ResourceTree, error) {
			return nil, errors.New("mock error")
		})
	defer patch.Reset()
	err := rescheduleSubTask(tree, faultSubTree)
	if err == nil {
		t.Error("rescheduleSubTask() should return error when GetSubTree failed")
	}
}

func TestRescheduleSubTask_ScheduleFailed(t *testing.T) {
	tree := createMockResourceTree()
	faultSubTree := &util.TaskTree{
		TaskNode: &util.TaskNode{ResourceNodeName: "child1"},
		Levels:   []util.TaskTreeLevel{{Name: "l0", ReqNode: 1}},
	}
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&util.ResourceTree{}), "FindNodeByTask",
		func(*util.ResourceTree, *util.TaskNode) (*util.ResourceNode, error) {
			return &util.ResourceNode{Name: "parent"}, nil
		}).ApplyMethod(reflect.TypeOf(&util.ResourceTree{}), "GetSubTree",
		func(*util.ResourceTree, *util.ResourceNode) (*util.ResourceTree, error) {
			return tree, nil
		}).ApplyFunc(Schedule, func(*util.ResourceTree, []util.TaskTreeLevel) (*util.TaskTree, error) {
		return nil, errors.New("mock error")
	})
	defer patch.Reset()
	err := rescheduleSubTask(tree, faultSubTree)
	if err == nil {
		t.Error("rescheduleSubTask() should return error when Schedule failed")
	}
}

func TestRescheduleSubTask_NoChildren(t *testing.T) {
	tree := createMockResourceTree()
	faultSubTree := &util.TaskTree{
		TaskNode: &util.TaskNode{ResourceNodeName: "child1"},
		Levels:   []util.TaskTreeLevel{{Name: "l0", ReqNode: 1}},
	}
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&util.ResourceTree{}), "FindNodeByTask",
		func(*util.ResourceTree, *util.TaskNode) (*util.ResourceNode, error) {
			return &util.ResourceNode{Name: "parent"}, nil
		}).ApplyMethod(reflect.TypeOf(&util.ResourceTree{}), "GetSubTree",
		func(*util.ResourceTree, *util.ResourceNode) (*util.ResourceTree, error) {
			return tree, nil
		}).ApplyFunc(Schedule, func(*util.ResourceTree, []util.TaskTreeLevel) (*util.TaskTree, error) {
		return &util.TaskTree{TaskNode: &util.TaskNode{Children: map[int]*util.TaskNode{}}}, nil
	})
	defer patch.Reset()
	err := rescheduleSubTask(tree, faultSubTree)
	if err == nil {
		t.Error("rescheduleSubTask() should return error when no children in result")
	}
}
