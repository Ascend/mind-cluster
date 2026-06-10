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
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/consts"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

const (
	testPluginName = "multilevel"
	testNodeName   = "test-node"
	testTaskName   = "test-task"
)

type handlerTestCase struct {
	name    string
	mh      *MultilevelHandler
	task    *api.TaskInfo
	node    plugin.NPUNode
	wantErr bool
}

type jobTestCase struct {
	name    string
	job     plugin.SchedulerJob
	task    *api.TaskInfo
	sMap    map[string]float64
	wantErr bool
}

type rankIndexTestCase struct {
	name    string
	task    *api.TaskInfo
	job     *plugin.SchedulerJob
	want    int
	wantErr bool
}

func newTestHandler() *MultilevelHandler {
	return &MultilevelHandler{
		NPUHandler: base.NPUHandler{
			SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{}}},
		},
	}
}

func newTestHandlerWithNodes() *MultilevelHandler {
	mh := newTestHandler()
	mh.Nodes = map[string]plugin.NPUNode{
		"node0": {CommonNode: plugin.CommonNode{Name: "node0"}},
		"node1": {CommonNode: plugin.CommonNode{Name: "node1"}},
	}
	return mh
}

func newTestTask(name string) *api.TaskInfo {
	return &api.TaskInfo{
		Name: name,
		Pod:  &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}}},
	}
}

func newTestNode(name string) plugin.NPUNode {
	return plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name: name, Label: map[string]string{}, Annotation: map[string]string{"a": "b"},
		},
	}
}

func newTestJobWithTasks(tasks map[api.TaskID]util.NPUTask) plugin.SchedulerJob {
	return plugin.SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: tasks}},
	}
}

func newTestSchedulerJob() plugin.SchedulerJob {
	return plugin.SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{}},
	}
}

func TestNew(t *testing.T) {
	p := New(testPluginName)
	if p == nil || p.GetPluginName() != testPluginName {
		t.Errorf("New() failed, got plugin=%v, name=%s", p, p.GetPluginName())
	}
}

func TestCheckTaskNPU(t *testing.T) {
	cases := []struct {
		name    string
		task    util.NPUTask
		wantErr bool
	}{
		{"valid", util.NPUTask{Name: testTaskName, ReqNPUNum: 8}, false},
		{"zero_req", util.NPUTask{Name: testTaskName, ReqNPUNum: 0}, true},
		{"skip_anno", util.NPUTask{ReqNPUNum: 0, Annotation: map[string]string{util.TaskSpecAnno: util.SchedulerType}}, false},
		{"skip_plugin", util.NPUTask{ReqNPUNum: 0, Annotation: map[string]string{util.SkipAscendPluginAnno: util.SkipEnabled}}, false},
	}
	for _, tc := range cases {
		mh := newTestHandler()
		mh.NPUJob.Tasks["task1"] = tc.task
		res := mh.checkTaskNPU()
		if (res != nil) != tc.wantErr {
			t.Errorf("%s: got err=%v, wantErr=%v", tc.name, res, tc.wantErr)
		}
	}
}

func TestCheckLevels(t *testing.T) {
	cases := []struct {
		name    string
		job     *util.NPUJob
		mockErr bool
		wantErr bool
	}{
		{"success", &util.NPUJob{NPUTaskNum: 2, AffinityBlocks: map[string]int{"topo": 1}}, false, false},
		{"error", &util.NPUJob{NPUTaskNum: 2, AffinityBlocks: map[string]int{}}, true, true},
	}
	for _, tc := range cases {
		mh := newTestHandler()
		mh.NPUJob = tc.job
		patch := gomonkey.ApplyFunc(util.GetTaskTreeLevels,
			func(map[string]int, int) ([]util.TaskTreeLevel, error) {
				if tc.mockErr {
					return nil, errors.New("test error")
				}
				return []util.TaskTreeLevel{{Name: "level", ReqNode: 2}}, nil
			})
		defer patch.Reset()
		res := mh.checkLevels()
		if (res != nil) != tc.wantErr {
			t.Errorf("%s: got err=%v, wantErr=%v", tc.name, res, tc.wantErr)
		}
	}
}

func TestValidNPUJob(t *testing.T) {
	mh := newTestHandler()
	mh.NPUJob.Tasks["task1"] = util.NPUTask{Name: testTaskName, ReqNPUNum: 0}
	if res := mh.ValidNPUJob(); res == nil {
		t.Error("ValidNPUJob() should return error for zero req npu")
	}
}

func TestCheckNodeNPUByTask(t *testing.T) {
	cases := []handlerTestCase{
		{"nil_params", nil, nil, plugin.NPUNode{}, true},
		{"empty_anno", newTestHandler(), newTestTask(testTaskName), plugin.NPUNode{}, true},
		{"topo_error", newTestHandler(), newTestTask(testTaskName), newTestNode(testNodeName), true},
	}
	mh := newTestHandler()
	mh.FrameAttr.ResourceLevelsInfo = map[string][]util.ResourceTreeLevel{
		util.DefaultTopoTree: {{Type: util.LevelTypeTree}, {Type: util.LevelTypeNode}},
	}
	cases = append(cases, handlerTestCase{"success", mh, newTestTask(testTaskName), newTestNode(testNodeName), false})
	for _, tc := range cases {
		patch := gomonkey.ApplyMethod(reflect.TypeOf(&base.NPUHandler{}), "GetTaskReqNPUNum",
			func(*base.NPUHandler, *api.TaskInfo) (int, error) { return 8, nil }).
			ApplyMethod(reflect.TypeOf(&base.NPUHandler{}), "GetUsableTopFromNode",
				func(*base.NPUHandler, plugin.NPUNode, bool) ([]int, error) { return []int{0, 1, 2, 3, 4, 5, 6, 7}, nil })
		err := tc.mh.CheckNodeNPUByTask(tc.task, tc.node)
		patch.Reset()
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: got err=%v, wantErr=%v", tc.name, err, tc.wantErr)
		}
	}
}

func TestCheckNodeNPUByTask_MiddleLevel(t *testing.T) {
	mh := newTestHandler()
	mh.FrameAttr.ResourceLevelsInfo = map[string][]util.ResourceTreeLevel{
		"custom": {{Type: util.LevelTypeTree}, {Type: util.LevelTypeMiddle, Label: "mid"}, {Type: util.LevelTypeNode}},
	}
	node := newTestNode(testNodeName)
	node.Label[util.TopoTreeLabel] = "custom"
	node.Label["mid"] = "val"
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&base.NPUHandler{}), "GetTaskReqNPUNum",
		func(*base.NPUHandler, *api.TaskInfo) (int, error) { return 8, nil }).
		ApplyMethod(reflect.TypeOf(&base.NPUHandler{}), "GetUsableTopFromNode",
			func(*base.NPUHandler, plugin.NPUNode, bool) ([]int, error) { return []int{0, 1, 2, 3, 4, 5, 6, 7}, nil })
	defer patch.Reset()
	if err := mh.CheckNodeNPUByTask(newTestTask(testTaskName), node); err != nil {
		t.Errorf("CheckNodeNPUByTask() unexpected error: %v", err)
	}
}

func TestCheckNodeNPUByTask_NodeNpuNotMatch(t *testing.T) {
	mh := newTestHandler()
	mh.FrameAttr.ResourceLevelsInfo = map[string][]util.ResourceTreeLevel{
		util.DefaultTopoTree: {{Type: util.LevelTypeTree}, {Type: util.LevelTypeNode}},
	}
	node := newTestNode(testNodeName)
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&base.NPUHandler{}), "GetTaskReqNPUNum",
		func(*base.NPUHandler, *api.TaskInfo) (int, error) { return 8, nil }).
		ApplyMethod(reflect.TypeOf(&base.NPUHandler{}), "GetUsableTopFromNode",
			func(*base.NPUHandler, plugin.NPUNode, bool) ([]int, error) { return []int{0, 1, 2, 3}, nil })
	defer patch.Reset()
	err := mh.CheckNodeNPUByTask(newTestTask(testTaskName), node)
	if err == nil {
		t.Error("CheckNodeNPUByTask() should return error for node npu not match")
	}
}

func TestScoreBestNPUNodes(t *testing.T) {
	cases := []jobTestCase{
		{"nil_params", plugin.SchedulerJob{}, nil, nil, true},
		{"job_not_exist", plugin.SchedulerJob{}, &api.TaskInfo{Job: "job2"}, map[string]float64{"n": 0}, true},
	}
	for _, tc := range cases {
		mh := newTestHandler()
		mh.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{"job1": tc.job}
		err := mh.ScoreBestNPUNodes(tc.task, []*api.NodeInfo{{Name: testNodeName}}, tc.sMap)
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: got err=%v, wantErr=%v", tc.name, err, tc.wantErr)
		}
	}
}

func TestScoreBestNPUNodes_JobReady(t *testing.T) {
	mh := newTestHandlerWithNodes()
	jobReady := true
	job := newTestSchedulerJob()
	job.JobReadyTag = &jobReady
	job.SuperPods = map[string][]plugin.SuperNode{"0": {{Name: "node0"}}}
	mh.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{"job1": job}
	task := &api.TaskInfo{Name: testTaskName, Job: "job1", Pod: &v1.Pod{}}
	err := mh.ScoreBestNPUNodes(task, []*api.NodeInfo{{Name: "node0"}}, map[string]float64{"node0": 0})
	if err != nil {
		t.Errorf("ScoreBestNPUNodes() unexpected error: %v", err)
	}
}

func TestScoreBestNPUNodes_NotEnoughNodes(t *testing.T) {
	mh := newTestHandler()
	jobReady := true
	job := newTestSchedulerJob()
	job.JobReadyTag = &jobReady
	job.SuperPods = map[string][]plugin.SuperNode{}
	mh.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{"job1": job}
	mh.SuperPodInfo = &plugin.SuperPodInfo{SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{}}
	mh.NPUTaskNum = 4
	mh.SchedulingTaskNum = 4
	mh.NPUJob.Tasks = map[api.TaskID]util.NPUTask{"t1": {}, "t2": {}, "t3": {}, "t4": {}}
	task := &api.TaskInfo{Name: testTaskName, Job: "job1", Pod: &v1.Pod{}}
	err := mh.ScoreBestNPUNodes(task, []*api.NodeInfo{{Name: "node0"}}, map[string]float64{"node0": 0})
	if err == nil {
		t.Error("ScoreBestNPUNodes() should return error for not enough nodes")
	}
}

func TestScoreBestNPUNodes_SelectNodesFailed(t *testing.T) {
	mh := newTestHandlerWithNodes()
	jobReady := true
	job := newTestSchedulerJob()
	job.JobReadyTag = &jobReady
	job.SuperPods = map[string][]plugin.SuperNode{}
	mh.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{"job1": job}
	mh.SuperPodInfo = &plugin.SuperPodInfo{SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{}}
	mh.taskLevels = []util.TaskTreeLevel{{Name: "l0", ReqNode: 2}}
	mh.FrameAttr.ResourceLevelsInfo = map[string][]util.ResourceTreeLevel{
		util.DefaultTopoTree: {{Type: util.LevelTypeTree}, {Type: util.LevelTypeNode}},
	}
	task := &api.TaskInfo{Name: testTaskName, Job: "job1", Pod: &v1.Pod{}}
	patch := gomonkey.ApplyFunc(plugin.GetResourceTrees,
		func(map[string]plugin.NPUNode, map[string][]util.ResourceTreeLevel, []util.TaskTreeLevel) ([]*util.ResourceTree, error) {
			return nil, errors.New("mock error")
		})
	defer patch.Reset()
	err := mh.ScoreBestNPUNodes(task, []*api.NodeInfo{{Name: "node0"}, {Name: "node1"}}, map[string]float64{"node0": 0})
	if err == nil {
		t.Error("ScoreBestNPUNodes() should return error when select nodes failed")
	}
}

func TestSelectNodesForMultiLevelJob(t *testing.T) {
	mh := newTestHandlerWithNodes()
	mh.taskLevels = []util.TaskTreeLevel{{Name: "l0", ReqNode: 2}, {Name: "l1", ReqNode: 1}, {Name: "l2", ReqNode: 1}}
	mh.FrameAttr.ResourceLevelsInfo = map[string][]util.ResourceTreeLevel{
		util.DefaultTopoTree: {{Type: util.LevelTypeTree}, {Type: util.LevelTypeMiddle}, {Type: util.LevelTypeNode}},
	}
	task := newTestTask(testTaskName)
	patch := gomonkey.ApplyFunc(plugin.GetResourceTrees,
		func(map[string]plugin.NPUNode, map[string][]util.ResourceTreeLevel, []util.TaskTreeLevel) ([]*util.ResourceTree, error) {
			return []*util.ResourceTree{{Name: "t", ResourceNode: &util.ResourceNode{Name: "r"},
				Levels: []util.ResourceTreeLevel{{Type: util.LevelTypeNode}}}}, nil
		}).ApplyFunc(Schedule, func(*util.ResourceTree, []util.TaskTreeLevel) (*util.TaskTree, error) {
		return &util.TaskTree{TaskNode: &util.TaskNode{ResourceNodeName: "r"}}, nil
	}).ApplyFunc(plugin.GetSuperNodeMapFromTaskTree, func(*util.TaskTree, map[string][]plugin.SuperNode) (map[string][]plugin.SuperNode, error) {
		return map[string][]plugin.SuperNode{"0": {{Name: "node0"}}}, nil
	})
	defer patch.Reset()
	mh.SuperPodInfo = &plugin.SuperPodInfo{SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{}}
	_, err := mh.selectNodesForMultiLevelJob(task, []*api.NodeInfo{{Name: "node0"}})
	if err != nil {
		t.Errorf("selectNodesForMultiLevelJob() unexpected error: %v", err)
	}
}

func TestSelectNodesForMultiLevelJob_OnlyL1(t *testing.T) {
	mh := newTestHandlerWithNodes()
	mh.taskLevels = []util.TaskTreeLevel{{Name: "l0", ReqNode: 2}, {Name: "l1", ReqNode: 1}, {Name: "l2", ReqNode: 1}}
	mh.FrameAttr.ResourceLevelsInfo = map[string][]util.ResourceTreeLevel{
		util.DefaultTopoTree: {{Type: util.LevelTypeTree}, {Type: util.LevelTypeMiddle}, {Type: util.LevelTypeNode}},
	}
	task := newTestTask(testTaskName)
	patch := gomonkey.ApplyFunc(plugin.GetResourceTrees,
		func(map[string]plugin.NPUNode, map[string][]util.ResourceTreeLevel, []util.TaskTreeLevel) ([]*util.ResourceTree, error) {
			return []*util.ResourceTree{{Name: "t", ResourceNode: &util.ResourceNode{Name: "r"},
				Levels: []util.ResourceTreeLevel{{Type: util.LevelTypeNode}}}}, nil
		}).ApplyFunc(Schedule, func(*util.ResourceTree, []util.TaskTreeLevel) (*util.TaskTree, error) {
		return &util.TaskTree{TaskNode: &util.TaskNode{ResourceNodeName: "r"}}, nil
	}).ApplyFunc(plugin.GetSuperNodeMapFromTaskTree, func(*util.TaskTree, map[string][]plugin.SuperNode) (map[string][]plugin.SuperNode, error) {
		return map[string][]plugin.SuperNode{"0": {{Name: "node0"}}}, nil
	})
	defer patch.Reset()
	mh.SuperPodInfo = &plugin.SuperPodInfo{SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{}}
	_, err := mh.selectNodesForMultiLevelJob(task, []*api.NodeInfo{{Name: "node0"}})
	if err != nil {
		t.Errorf("selectNodesForMultiLevelJob() unexpected error: %v", err)
	}
}

func TestTryScheduleInStrictRules(t *testing.T) {
	mh := newTestHandlerWithNodes()
	mh.taskLevels = []util.TaskTreeLevel{{Name: "l0", ReqNode: 2}, {Name: "l1", ReqNode: 1}}
	mh.FrameAttr.ResourceLevelsInfo = map[string][]util.ResourceTreeLevel{
		util.DefaultTopoTree: {{Type: util.LevelTypeTree}, {Type: util.LevelTypeMiddle}, {Type: util.LevelTypeNode}},
	}
	mh.SuperPodInfo = &plugin.SuperPodInfo{SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{}}
	task := newTestTask(testTaskName)
	patch := gomonkey.ApplyFunc(plugin.GetResourceTrees,
		func(map[string]plugin.NPUNode, map[string][]util.ResourceTreeLevel, []util.TaskTreeLevel) ([]*util.ResourceTree, error) {
			return []*util.ResourceTree{{Name: "t", ResourceNode: &util.ResourceNode{Name: "r"},
				Levels: []util.ResourceTreeLevel{{Type: util.LevelTypeNode}}}}, nil
		}).ApplyFunc(Schedule, func(*util.ResourceTree, []util.TaskTreeLevel) (*util.TaskTree, error) {
		return &util.TaskTree{TaskNode: &util.TaskNode{ResourceNodeName: "r"}}, nil
	}).ApplyFunc(plugin.GetSuperNodeMapFromTaskTree, func(*util.TaskTree, map[string][]plugin.SuperNode) (map[string][]plugin.SuperNode, error) {
		return map[string][]plugin.SuperNode{"0": {{Name: "node0"}}}, nil
	})
	defer patch.Reset()
	_, err := mh.tryScheduleInStrictRules(task, []*api.NodeInfo{{Name: "node0"}})
	if err != nil {
		t.Logf("tryScheduleInStrictRules() returned error: %v", err)
	}
}

func TestScheduleMultipleLevelPodsForJob(t *testing.T) {
	mh := newTestHandlerWithNodes()
	mh.taskLevels = []util.TaskTreeLevel{{Name: "l0", ReqNode: 2}}
	mh.FrameAttr.ResourceLevelsInfo = map[string][]util.ResourceTreeLevel{
		util.DefaultTopoTree: {{Type: util.LevelTypeTree}, {Type: util.LevelTypeNode}},
	}
	task := newTestTask(testTaskName)
	patch := gomonkey.ApplyFunc(plugin.GetResourceTrees,
		func(map[string]plugin.NPUNode, map[string][]util.ResourceTreeLevel, []util.TaskTreeLevel) ([]*util.ResourceTree, error) {
			return []*util.ResourceTree{{Name: "t", ResourceNode: &util.ResourceNode{Name: "r"},
				Levels: []util.ResourceTreeLevel{{Type: util.LevelTypeNode}}}}, nil
		}).ApplyFunc(Schedule, func(*util.ResourceTree, []util.TaskTreeLevel) (*util.TaskTree, error) {
		return &util.TaskTree{TaskNode: &util.TaskNode{ResourceNodeName: "r"}}, nil
	}).ApplyFunc(plugin.GetSuperNodeMapFromTaskTree, func(*util.TaskTree, map[string][]plugin.SuperNode) (map[string][]plugin.SuperNode, error) {
		return map[string][]plugin.SuperNode{"0": {{Name: "node0"}}}, nil
	})
	defer patch.Reset()
	mh.SuperPodInfo = &plugin.SuperPodInfo{SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{}}
	_, err := mh.scheduleMultipleLevelPodsForJob(task, []*api.NodeInfo{{Name: "node0"}})
	if err != nil {
		t.Errorf("scheduleMultipleLevelPodsForJob() unexpected error: %v", err)
	}
}

func TestScheduleMultipleLevelPodsForJob_GetResourceTreesFailed(t *testing.T) {
	mh := newTestHandlerWithNodes()
	mh.taskLevels = []util.TaskTreeLevel{{Name: "l0", ReqNode: 2}}
	mh.FrameAttr.ResourceLevelsInfo = map[string][]util.ResourceTreeLevel{
		util.DefaultTopoTree: {{Type: util.LevelTypeTree}, {Type: util.LevelTypeNode}},
	}
	task := newTestTask(testTaskName)
	patch := gomonkey.ApplyFunc(plugin.GetResourceTrees,
		func(map[string]plugin.NPUNode, map[string][]util.ResourceTreeLevel, []util.TaskTreeLevel) ([]*util.ResourceTree, error) {
			return nil, errors.New("mock error")
		})
	defer patch.Reset()
	_, err := mh.scheduleMultipleLevelPodsForJob(task, []*api.NodeInfo{{Name: "node0"}})
	if err == nil {
		t.Error("scheduleMultipleLevelPodsForJob() should return error when GetResourceTrees failed")
	}
}

func TestScheduleMultipleLevelPodsForJob_NoValidTaskTree(t *testing.T) {
	mh := newTestHandlerWithNodes()
	mh.taskLevels = []util.TaskTreeLevel{{Name: "l0", ReqNode: 2}}
	mh.FrameAttr.ResourceLevelsInfo = map[string][]util.ResourceTreeLevel{
		util.DefaultTopoTree: {{Type: util.LevelTypeTree}, {Type: util.LevelTypeNode}},
	}
	task := newTestTask(testTaskName)
	patch := gomonkey.ApplyFunc(plugin.GetResourceTrees,
		func(map[string]plugin.NPUNode, map[string][]util.ResourceTreeLevel, []util.TaskTreeLevel) ([]*util.ResourceTree, error) {
			return []*util.ResourceTree{{Name: "t", ResourceNode: &util.ResourceNode{Name: "r"},
				Levels: []util.ResourceTreeLevel{{Type: util.LevelTypeNode}}}}, nil
		}).ApplyFunc(Schedule, func(*util.ResourceTree, []util.TaskTreeLevel) (*util.TaskTree, error) {
		return nil, errors.New("mock error")
	})
	defer patch.Reset()
	_, err := mh.scheduleMultipleLevelPodsForJob(task, []*api.NodeInfo{{Name: "node0"}})
	if err == nil {
		t.Error("scheduleMultipleLevelPodsForJob() should return error when no valid task tree")
	}
}

func TestTryScheduleTaskInSingleTree_Schedule(t *testing.T) {
	mh := newTestHandlerWithNodes()
	mh.taskLevels = []util.TaskTreeLevel{{Name: "l0", ReqNode: 1}}
	mh.FrameAttr.ResourceLevelsInfo = map[string][]util.ResourceTreeLevel{
		util.DefaultTopoTree: {{Type: util.LevelTypeTree, ReservedNode: 0}, {Type: util.LevelTypeNode, ReservedNode: 0}},
	}
	jobReady := true
	job := plugin.SchedulerJob{JobReadyTag: &jobReady}
	mh.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{"job1": job}
	mh.SuperPodInfo = &plugin.SuperPodInfo{SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{}}
	task := &api.TaskInfo{Name: testTaskName, Job: "job1", Pod: &v1.Pod{}}
	nodes := []*api.NodeInfo{{Name: "node0"}}
	_, err := mh.scheduleMultipleLevelPodsForJob(task, nodes)
	if err != nil {
		t.Errorf("scheduleMultipleLevelPodsForJob() unexpected error: %v", err)
	}
}

func TestRescheduleWithSuperPods(t *testing.T) {
	mh := newTestHandler()
	mh.taskLevels = []util.TaskTreeLevel{{Name: "l0", ReqNode: 1}}
	mh.Nodes = map[string]plugin.NPUNode{
		"node0": {CommonNode: plugin.CommonNode{Name: "node0"}},
		"node2": {CommonNode: plugin.CommonNode{Name: "node2"}},
	}
	mh.FrameAttr.ResourceLevelsInfo = map[string][]util.ResourceTreeLevel{
		util.DefaultTopoTree: {{Type: util.LevelTypeTree, ReservedNode: 0}, {Type: util.LevelTypeNode, ReservedNode: 0}},
	}
	tree := &util.ResourceTree{Name: "t", ResourceNode: &util.ResourceNode{Name: "r"},
		Levels: []util.ResourceTreeLevel{{Type: util.LevelTypeNode, ReservedNode: 0}}}
	ctx := &rescheduleContext{
		task:         &api.TaskInfo{Name: testTaskName, Job: "job1", Pod: &v1.Pod{}},
		superPods:    map[string][]plugin.SuperNode{"0": {{Name: "node0", TopoTreeName: util.DefaultTopoTree}, {TopoTreeName: util.DefaultTopoTree}}},
		missingNodes: []string{"node1"},
	}
	nodes := []*api.NodeInfo{{Name: "node0"}, {Name: "node2"}}
	patch := gomonkey.ApplyFunc(plugin.GetResourceTrees,
		func(map[string]plugin.NPUNode, map[string][]util.ResourceTreeLevel, []util.TaskTreeLevel) ([]*util.ResourceTree, error) {
			return []*util.ResourceTree{tree}, nil
		}).ApplyFunc(plugin.GetTaskTreeFromSuperNodeMap,
		func(map[string][]plugin.SuperNode, []util.TaskTreeLevel, []util.ResourceTreeLevel, map[string]plugin.NPUNode) (*util.TaskTree, error) {
			return &util.TaskTree{TaskNode: &util.TaskNode{}}, nil
		}).ApplyFunc(Reschedule,
		func(*util.ResourceTree, *util.TaskTree, []string) (*util.TaskTree, error) {
			return &util.TaskTree{TaskNode: &util.TaskNode{}}, nil
		})
	defer patch.Reset()
	_, err := mh.rescheduleWithSuperPods(ctx, nodes, 0)
	if err != nil {
		t.Errorf("rescheduleWithSuperPods() unexpected error: %v", err)
	}
}

func TestGetHcclRankIndex(t *testing.T) {
	cases := []rankIndexTestCase{
		{"from_anno", &api.TaskInfo{Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{plugin.PodRankIndexKey: "5"}}}},
			&plugin.SchedulerJob{}, 5, false},
		{"invalid_anno", &api.TaskInfo{Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{plugin.PodRankIndexKey: "invalid"}}}},
			&plugin.SchedulerJob{}, 0, true},
		{"from_task", &api.TaskInfo{UID: "t1", Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}}}},
			&plugin.SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{"t1": {Index: 3}}}}}, 3, false},
		{"task_not_exist", &api.TaskInfo{UID: "t1", Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}}}},
			&plugin.SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{}}}}, 0, true},
	}
	for _, tc := range cases {
		got, err := getHcclRankIndex(tc.task, *tc.job)
		if (err != nil) != tc.wantErr || got != tc.want {
			t.Errorf("%s: got=%d, err=%v, want=%d, wantErr=%v", tc.name, got, err, tc.want, tc.wantErr)
		}
	}
}

func TestGetL1Ranks(t *testing.T) {
	nodes := map[string][]plugin.SuperNode{
		"0": {{Name: "n0"}, {Name: "n1"}},
		"1": {{Name: "n2"}, {Name: "n3"}},
	}
	cases := []struct {
		name      string
		nodes     map[string][]plugin.SuperNode
		rank      int
		wantKey   string
		wantLocal int
		wantErr   bool
	}{
		{"empty", map[string][]plugin.SuperNode{}, 0, "", 0, true},
		{"success", nodes, 1, "0", 1, false},
		{"boundary", nodes, 2, "1", 0, false},
		{"exceeds", nodes, 10, "", 0, true},
	}
	for _, tc := range cases {
		key, local, err := getL1Ranks(tc.nodes, tc.rank)
		if (err != nil) != tc.wantErr || key != tc.wantKey || local != tc.wantLocal {
			t.Errorf("%s: key=%s, local=%d, err=%v, wantKey=%s, wantLocal=%d, wantErr=%v",
				tc.name, key, local, err, tc.wantKey, tc.wantLocal, tc.wantErr)
		}
	}
}

func TestGetFaultJob(t *testing.T) {
	patch := gomonkey.ApplyFunc(rescheduling.GetReSchedulerCache,
		func() *rescheduling.DealReSchedulerCache { return nil })
	_, ok := getFaultJob("job1")
	patch.Reset()
	if ok {
		t.Error("getFaultJob() should return false for nil cache")
	}
}

func TestGetFaultNodes(t *testing.T) {
	mh := newTestHandler()
	mh.SuperPodInfo = &plugin.SuperPodInfo{SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{}}
	if _, err := mh.getFaultNodes("job1"); err == nil {
		t.Error("getFaultNodes() should return error for job not exist")
	}
	mh.SuperPodInfo.SuperPodMapFaultTaskNodes["job1"] = map[string]string{"t1": "n0", "t2": "n1"}
	nodes, err := mh.getFaultNodes("job1")
	if err != nil || len(nodes) != 2 {
		t.Errorf("getFaultNodes() got len=%d, err=%v, want len=2", len(nodes), err)
	}
}

func TestObtainBatchScoreRank(t *testing.T) {
	cases := []struct {
		name string
		task *api.TaskInfo
		job  *plugin.SchedulerJob
		want int
	}{
		{"nil_task", nil, &plugin.SchedulerJob{}, 0},
		{"nil_job", newTestTask(testTaskName), nil, 0},
		{"no_anno", newTestTask(testTaskName), &plugin.SchedulerJob{}, 0},
		{"valid", &api.TaskInfo{Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{util.TaskSpecAnno: "spec"}}}},
			&plugin.SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{"t1": {
				Annotation: map[string]string{util.TaskSpecAnno: "spec", plugin.PodRankIndexKey: "0"},
				ReqNPUName: "huawei.com/Ascend910", PodStatus: v1.PodPending}}}}}, 1},
	}
	for _, tc := range cases {
		mh := newTestHandler()
		result := mh.obtainBatchScoreRank(tc.task, tc.job)
		if len(result) != tc.want {
			t.Errorf("%s: got len=%d, want %d", tc.name, len(result), tc.want)
		}
	}
}

func TestScoreNodeForReadyJob(t *testing.T) {
	mh := newTestHandler()
	job := plugin.SchedulerJob{SuperPods: map[string][]plugin.SuperNode{"0": {{Name: "node0"}}}}
	sMap := map[string]float64{"node0": 0}
	patch := gomonkey.ApplyFunc(getHcclRankIndex, func(*api.TaskInfo, plugin.SchedulerJob) (int, error) { return 0, nil }).
		ApplyFunc(getL1Ranks, func(map[string][]plugin.SuperNode, int) (string, int, error) { return "0", 0, nil })
	defer patch.Reset()
	mh.scoreNodeForReadyJob(newTestTask(testTaskName), job, sMap)
	if sMap["node0"] != float64(scoreForNode) {
		t.Errorf("scoreNodeForReadyJob() sMap[node0]=%f, want %d", sMap["node0"], scoreForNode)
	}
}

func TestScoreNodeForReadyJob_NilSMap(t *testing.T) {
	mh := newTestHandler()
	job := plugin.SchedulerJob{SuperPods: map[string][]plugin.SuperNode{"0": {{Name: "node0"}}}}
	mh.scoreNodeForReadyJob(newTestTask(testTaskName), job, nil)
}

func TestScoreNodeForReadyJob_GetRankFailed(t *testing.T) {
	mh := newTestHandler()
	job := plugin.SchedulerJob{SuperPods: map[string][]plugin.SuperNode{"0": {{Name: "node0"}}}}
	sMap := map[string]float64{"node0": 0}
	patch := gomonkey.ApplyFunc(getHcclRankIndex, func(*api.TaskInfo, plugin.SchedulerJob) (int, error) { return 0, errors.New("mock") })
	defer patch.Reset()
	mh.scoreNodeForReadyJob(newTestTask(testTaskName), job, sMap)
}

func TestScoreNodeBatchForReadyJob(t *testing.T) {
	mh := newTestHandler()
	mh.taskLevels = []util.TaskTreeLevel{{Name: "l0", ReqNode: 2}, {Name: "l1", ReqNode: 1}}
	jobReady := true
	job := &plugin.SchedulerJob{JobReadyTag: &jobReady, SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{"t1": {
		Annotation: map[string]string{util.TaskSpecAnno: "spec", plugin.PodRankIndexKey: "0"},
		ReqNPUName: "huawei.com/Ascend910", PodStatus: v1.PodPending}}}}}
	job.SuperPods = map[string][]plugin.SuperNode{"0": {{Name: "node0"}}}
	task := &api.TaskInfo{Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{util.TaskSpecAnno: "spec"}}}}
	mh.scoreNodeBatchForReadyJob(task, job, map[string]float64{"node0": 0})
}

func TestScoreNodeBatchForReadyJob_NilParams(t *testing.T) {
	mh := newTestHandler()
	mh.scoreNodeBatchForReadyJob(nil, nil, nil)
}

func TestSelectNodeFromCache(t *testing.T) {
	mh := newTestHandler()
	jobReady := true
	job := &plugin.SchedulerJob{JobReadyTag: &jobReady, SuperPods: map[string][]plugin.SuperNode{"0": {{Name: "node0"}}}}
	sMap := map[string]float64{"node0": 0}
	patch := gomonkey.ApplyFunc(getHcclRankIndex, func(*api.TaskInfo, plugin.SchedulerJob) (int, error) { return 0, nil }).
		ApplyFunc(getL1Ranks, func(map[string][]plugin.SuperNode, int) (string, int, error) { return "0", 0, nil })
	defer patch.Reset()
	mh.selectNodeFromCache(job, newTestTask(testTaskName), sMap)
}

func TestSelectNodeFromCache_WithPodGroup(t *testing.T) {
	mh := newTestHandler()
	mh.taskLevels = []util.TaskTreeLevel{{Name: "l0", ReqNode: 2}, {Name: "l1", ReqNode: 1}}
	jobReady := true
	job := &plugin.SchedulerJob{
		JobReadyTag: &jobReady,
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{Label: map[string]string{plugin.PodGroupScheduleKey: plugin.PodGroupScheduleValue}},
			NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{"t1": {
				Annotation: map[string]string{util.TaskSpecAnno: "spec", plugin.PodRankIndexKey: "0"},
				ReqNPUName: "huawei.com/Ascend910", PodStatus: v1.PodPending}}},
		},
		SuperPods: map[string][]plugin.SuperNode{"0": {{Name: "node0"}}},
	}
	sMap := map[string]float64{"node0": 0}
	task := &api.TaskInfo{Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{util.TaskSpecAnno: "spec"}}}}
	mh.selectNodeFromCache(job, task, sMap)
}

func TestMultilevelHandler_SetPluginName(t *testing.T) {
	mh := &MultilevelHandler{}
	mh.SetPluginName(testPluginName)
	if mh.GetPluginName() != testPluginName {
		t.Errorf("SetPluginName() failed, got %s", mh.GetPluginName())
	}
}

type checkNodeForHotSwitchCase struct {
	name    string
	mh      *MultilevelHandler
	task    *api.TaskInfo
	node    plugin.NPUNode
	wantErr bool
}

func buildHotSwitchCheckHandler() *MultilevelHandler {
	jobReady := true
	mh := newTestHandler()
	mh.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{
		"job1": {
			JobReadyTag: &jobReady,
			SuperPods: map[string][]plugin.SuperNode{
				"0": {{Name: "node0", TopoTreeName: "topoA"}, {Name: "node1", TopoTreeName: "topoA"}},
			},
		},
	}
	mh.FrameAttr.ResourceLevelsInfo = map[string][]util.ResourceTreeLevel{
		"topoA": {
			{Type: util.LevelTypeTree, Label: util.TopoTreeLabel},
			{Type: util.LevelTypeMiddle, Label: "rack-id"},
			{Type: util.LevelTypeNode},
		},
	}
	mh.Nodes = map[string]plugin.NPUNode{
		"node0": {CommonNode: plugin.CommonNode{
			Name: "node0", Label: map[string]string{util.TopoTreeLabel: "topoA", "rack-id": "rack1"},
		}},
	}
	return mh
}

func buildCheckNodeForHotSwitchCases() []checkNodeForHotSwitchCase {
	mh := buildHotSwitchCheckHandler()
	backupTask := &api.TaskInfo{
		Name: "backup-pod", Job: "job1",
		Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{consts.BackupSourcePodNameKey: "origin-pod"},
		}},
	}
	normalTask := newTestTask("normal-pod")
	normalTask.Job = "job1"
	sameL1Node := plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name: "node-same-l1", Label: map[string]string{util.TopoTreeLabel: "topoA", "rack-id": "rack1"},
			Annotation: map[string]string{"a": "b"},
		},
	}
	diffL1Node := plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name: "node-diff-l1", Label: map[string]string{util.TopoTreeLabel: "topoA", "rack-id": "rack2"},
			Annotation: map[string]string{"a": "b"},
		},
	}
	diffTopoNode := plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name: "node-diff", Label: map[string]string{util.TopoTreeLabel: "topoB"},
			Annotation: map[string]string{"a": "b"},
		},
	}
	noTopoNode := newTestNode("node-notopo")
	mhNoJob := newTestHandler()
	mhNoJob.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{}
	mhNotReady := newTestHandler()
	mhNotReady.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{
		"job1": {SuperPods: map[string][]plugin.SuperNode{}},
	}
	return []checkNodeForHotSwitchCase{
		{"normal_pod_skip", mh, normalTask, diffTopoNode, false},
		{"backup_pod_same_l1", mh, backupTask, sameL1Node, false},
		{"backup_pod_diff_l1_same_topo", mh, backupTask, diffL1Node, true},
		{"backup_pod_diff_topo", mh, backupTask, diffTopoNode, true},
		{"backup_pod_no_topo_label", mh, backupTask, noTopoNode, true},
		{"backup_pod_job_not_exist", mhNoJob, backupTask, sameL1Node, true},
		{"backup_pod_job_not_ready", mhNotReady, backupTask, sameL1Node, false},
	}
}

func TestCheckNodeForHotSwitch(t *testing.T) {
	cases := buildCheckNodeForHotSwitchCases()
	for _, tc := range cases {
		patch := gomonkey.ApplyFunc(getHcclRankIndex,
			func(*api.TaskInfo, plugin.SchedulerJob) (int, error) { return 0, nil }).
			ApplyFunc(getL1Ranks,
				func(map[string][]plugin.SuperNode, int) (string, int, error) { return "0", 0, nil })
		err := tc.mh.checkNodeForHotSwitch(tc.task, tc.node)
		patch.Reset()
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: got err=%v, wantErr=%v", tc.name, err, tc.wantErr)
		}
	}
}

func TestCheckNodeForHotSwitch_GetRankFailed(t *testing.T) {
	jobReady := true
	mh := newTestHandler()
	mh.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{
		"job1": {JobReadyTag: &jobReady, SuperPods: map[string][]plugin.SuperNode{"0": {{Name: "n0"}}}},
	}
	task := &api.TaskInfo{
		Name: "backup-pod", Job: "job1",
		Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{consts.BackupSourcePodNameKey: "origin-pod"},
		}},
	}
	patch := gomonkey.ApplyFunc(getHcclRankIndex,
		func(*api.TaskInfo, plugin.SchedulerJob) (int, error) { return 0, errors.New("mock") })
	defer patch.Reset()
	err := mh.checkNodeForHotSwitch(task, newTestNode("node0"))
	if err == nil {
		t.Error("checkNodeForHotSwitch() should return error when getHcclRankIndex failed")
	}
}

func TestCheckNodeForHotSwitch_GetL1RanksFailed(t *testing.T) {
	jobReady := true
	mh := newTestHandler()
	mh.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{
		"job1": {JobReadyTag: &jobReady, SuperPods: map[string][]plugin.SuperNode{"0": {{Name: "n0"}}}},
	}
	task := &api.TaskInfo{
		Name: "backup-pod", Job: "job1",
		Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{consts.BackupSourcePodNameKey: "origin-pod"},
		}},
	}
	patch := gomonkey.ApplyFunc(getHcclRankIndex,
		func(*api.TaskInfo, plugin.SchedulerJob) (int, error) { return 0, nil }).
		ApplyFunc(getL1Ranks,
			func(map[string][]plugin.SuperNode, int) (string, int, error) { return "", 0, errors.New("mock") })
	defer patch.Reset()
	err := mh.checkNodeForHotSwitch(task, newTestNode("node0"))
	if err == nil {
		t.Error("checkNodeForHotSwitch() should return error when getL1Ranks failed")
	}
}

func TestCheckNodeForHotSwitch_EmptyTopoTree(t *testing.T) {
	jobReady := true
	mh := newTestHandler()
	mh.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{
		"job1": {
			JobReadyTag: &jobReady,
			SuperPods:   map[string][]plugin.SuperNode{"0": {{Name: "n0", TopoTreeName: ""}}},
		},
	}
	task := &api.TaskInfo{
		Name: "backup-pod", Job: "job1",
		Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{consts.BackupSourcePodNameKey: "origin-pod"},
		}},
	}
	node := plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name: "node-any", Label: map[string]string{util.TopoTreeLabel: "topoB"},
			Annotation: map[string]string{"a": "b"},
		},
	}
	patch := gomonkey.ApplyFunc(getHcclRankIndex,
		func(*api.TaskInfo, plugin.SchedulerJob) (int, error) { return 0, nil }).
		ApplyFunc(getL1Ranks,
			func(map[string][]plugin.SuperNode, int) (string, int, error) { return "0", 0, nil })
	defer patch.Reset()
	err := mh.checkNodeForHotSwitch(task, node)
	if err == nil {
		t.Error("checkNodeForHotSwitch() should return error when cached TopoTreeName is empty")
	}
}

func TestScoreNodeForHotSwitchBackupPod(t *testing.T) {
	mh := newTestHandler()
	sMap := map[string]float64{"nodeA": 1.0, "nodeB": 2.0}
	mh.scoreNodeForHotSwitchBackupPod(sMap)
	expectedBonus := float64(scoreForNode)
	if sMap["nodeA"] != 1.0+expectedBonus || sMap["nodeB"] != 2.0+expectedBonus {
		t.Errorf("scoreNodeForHotSwitchBackupPod() sMap=%v, want bonus=%f", sMap, expectedBonus)
	}
}

func TestScoreNodeForReadyJob_BackupPod(t *testing.T) {
	mh := newTestHandler()
	sMap := map[string]float64{"nodeA": 0}
	task := &api.TaskInfo{
		Name: "backup-pod",
		Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{consts.BackupSourcePodNameKey: "origin-pod"},
		}},
	}
	mh.scoreNodeForReadyJob(task, plugin.SchedulerJob{}, sMap)
	if sMap["nodeA"] <= 0 {
		t.Errorf("scoreNodeForReadyJob() should score backup pod's node, got %f", sMap["nodeA"])
	}
}

func TestScoreNodeBatchForReadyJob_BackupPod(t *testing.T) {
	mh := newTestHandler()
	jobReady := true
	job := &plugin.SchedulerJob{JobReadyTag: &jobReady}
	sMap := map[string]float64{"nodeA": 0}
	task := &api.TaskInfo{
		Name: "backup-pod",
		Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{consts.BackupSourcePodNameKey: "origin-pod"},
		}},
	}
	mh.scoreNodeBatchForReadyJob(task, job, sMap)
	if sMap["nodeA"] <= 0 {
		t.Errorf("scoreNodeBatchForReadyJob() should score backup pod's node, got %f", sMap["nodeA"])
	}
}

func TestIsCachedSuperPodsValid(t *testing.T) {
	mh := newTestHandlerWithNodes()
	cases := []struct {
		name      string
		superPods map[string][]plugin.SuperNode
		nodes     []*api.NodeInfo
		tasks     map[api.TaskID]util.NPUTask
		want      bool
	}{
		{
			"all_in_candidate_set",
			map[string][]plugin.SuperNode{"0": {{Name: "node0"}, {Name: "node1"}}},
			[]*api.NodeInfo{{Name: "node0"}, {Name: "node1"}},
			nil, true,
		},
		{
			"node_missing_from_candidates",
			map[string][]plugin.SuperNode{"0": {{Name: "node0"}}},
			[]*api.NodeInfo{{Name: "node1"}},
			nil, false,
		},
		{
			"running_pod_skipped",
			map[string][]plugin.SuperNode{"0": {{Name: "node0"}}},
			[]*api.NodeInfo{},
			map[api.TaskID]util.NPUTask{"t1": {NodeName: "node0"}},
			true,
		},
		{
			"empty_superpods",
			map[string][]plugin.SuperNode{},
			[]*api.NodeInfo{{Name: "node0"}},
			nil, true,
		},
		{
			"mixed_running_and_missing",
			map[string][]plugin.SuperNode{"0": {{Name: "node0"}, {Name: "node1"}}},
			[]*api.NodeInfo{{Name: "node0"}},
			map[api.TaskID]util.NPUTask{"t1": {NodeName: "node1"}},
			true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tasks := tc.tasks
			if tasks == nil {
				tasks = map[api.TaskID]util.NPUTask{}
			}
			job := &plugin.SchedulerJob{
				SuperPods:        tc.superPods,
				SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: tasks}},
			}
			if got := mh.isCachedSuperPodsValid(job, tc.nodes); got != tc.want {
				t.Errorf("isCachedSuperPodsValid() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRemoveAll(t *testing.T) {
	src := map[string]struct{}{"a": {}, "b": {}, "c": {}}
	got := removeAll(src, []string{"b", "c"})
	if len(got) != 1 {
		t.Fatalf("removeAll() len=%d, want 1", len(got))
	}
	if _, ok := got["a"]; !ok {
		t.Error("removeAll() should keep 'a'")
	}
}

func TestRemoveAll_EmptyRemove(t *testing.T) {
	src := map[string]struct{}{"a": {}, "b": {}}
	got := removeAll(src, nil)
	if len(got) != 2 {
		t.Errorf("removeAll(nil) len=%d, want 2", len(got))
	}
}

func TestRemoveAll_EmptySrc(t *testing.T) {
	got := removeAll(map[string]struct{}{}, []string{"a"})
	if len(got) != 0 {
		t.Errorf("removeAll(empty) len=%d, want 0", len(got))
	}
}

func TestTryUseCachedSuperPods(t *testing.T) {
	mh := newTestHandlerWithNodes()
	task := &api.TaskInfo{Name: testTaskName, Job: "job1", Pod: &v1.Pod{}}
	candidateNodes := []*api.NodeInfo{{Name: "node0"}, {Name: "node1"}}

	t.Run("empty_superpods", func(t *testing.T) {
		job := &plugin.SchedulerJob{}
		if mh.tryUseCachedSuperPods(job, task, candidateNodes) {
			t.Error("empty SuperPods should return false")
		}
	})
	t.Run("already_verified", func(t *testing.T) {
		job := &plugin.SchedulerJob{
			SuperPods:         map[string][]plugin.SuperNode{"0": {{Name: "node0"}}},
			SuperPodsVerified: true,
		}
		if !mh.tryUseCachedSuperPods(job, task, candidateNodes) {
			t.Error("Verified should return true")
		}
	})
	t.Run("not_verified_cache_valid", func(t *testing.T) {
		job := &plugin.SchedulerJob{
			SuperPods:        map[string][]plugin.SuperNode{"0": {{Name: "node0"}}},
			SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{}}},
		}
		if !mh.tryUseCachedSuperPods(job, task, candidateNodes) {
			t.Error("valid cache should return true and set Verified")
		}
		if !job.SuperPodsVerified {
			t.Error("should set SuperPodsVerified to true")
		}
	})
	t.Run("not_verified_cache_stale", func(t *testing.T) {
		job := &plugin.SchedulerJob{
			SuperPods:        map[string][]plugin.SuperNode{"0": {{Name: "nodeX"}}},
			SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{}}},
		}
		if mh.tryUseCachedSuperPods(job, task, candidateNodes) {
			t.Error("stale cache should return false")
		}
		if !job.SuperPodsVerified {
			t.Error("should set SuperPodsVerified to true even on stale")
		}
	})
}

func TestGetHistoricalHealthyNodeNames(t *testing.T) {
	superPods := map[string][]plugin.SuperNode{
		"0": {{Name: "node0"}, {Name: "node1"}},
		"1": {{Name: "node2"}, {Name: ""}},
	}
	faultNodes := []string{"node1"}
	pinned := getHistoricalHealthyNodeNames(superPods, faultNodes)
	if len(pinned) != 2 {
		t.Errorf("len=%d, want 2", len(pinned))
	}
	for _, n := range []string{"node0", "node2"} {
		if _, ok := pinned[n]; !ok {
			t.Errorf("missing healthy node %s", n)
		}
	}
}

func TestFilterOutAvailableNodes(t *testing.T) {
	nodes := []*api.NodeInfo{{Name: "node0"}, {Name: "node2"}}
	got := filterOutAvailableNodes([]string{"node0", "node1", "node2", "node3"}, nodes)
	if len(got) != 2 {
		t.Fatalf("len=%d, want 2", len(got))
	}
	remaining := map[string]struct{}{got[0]: {}, got[1]: {}}
	for _, expected := range []string{"node1", "node3"} {
		if _, ok := remaining[expected]; !ok {
			t.Errorf("missing %s", expected)
		}
	}
}

func TestGetMissingNodesFromJob(t *testing.T) {
	mh := newTestHandler()
	superPods := map[string][]plugin.SuperNode{
		"0": {{Name: "node0"}, {Name: "node1"}},
		"1": {{Name: "node2"}, {Name: "node3"}},
	}

	t.Run("all_pending", func(t *testing.T) {
		tasks := map[api.TaskID]util.NPUTask{
			"t0": {PodStatus: v1.PodPending, Annotation: map[string]string{plugin.PodRankIndexKey: "0"}},
			"t1": {PodStatus: v1.PodPending, Annotation: map[string]string{plugin.PodRankIndexKey: "1"}},
			"t3": {PodStatus: v1.PodPending, Annotation: map[string]string{plugin.PodRankIndexKey: "3"}},
		}
		patch := gomonkey.ApplyFunc(getL1Ranks, func(sp map[string][]plugin.SuperNode, rank int) (string, int, error) {
			switch rank {
			case 0:
				return "0", 0, nil
			case 1:
				return "0", 1, nil
			case 3:
				return "1", 1, nil
			}
			return "", 0, fmt.Errorf("unexpected rank %d", rank)
		})
		defer patch.Reset()
		got := mh.getMissingNodesFromJob(superPods, tasks)
		if len(got) != 3 {
			t.Fatalf("len=%d, want 3", len(got))
		}
		expect := map[string]struct{}{"node0": {}, "node1": {}, "node3": {}}
		for _, n := range got {
			if _, ok := expect[n]; !ok {
				t.Errorf("unexpected node %s", n)
			}
		}
	})

	t.Run("mixed_status", func(t *testing.T) {
		tasks := map[api.TaskID]util.NPUTask{
			"t0": {PodStatus: v1.PodPending, Annotation: map[string]string{plugin.PodRankIndexKey: "0"}},
			"t1": {PodStatus: v1.PodRunning, Annotation: map[string]string{plugin.PodRankIndexKey: "1"}},
		}
		patch := gomonkey.ApplyFunc(getL1Ranks, func(sp map[string][]plugin.SuperNode, rank int) (string, int, error) {
			return "0", rank, nil
		})
		defer patch.Reset()
		got := mh.getMissingNodesFromJob(superPods, tasks)
		if len(got) != 1 || got[0] != "node0" {
			t.Errorf("got=%v, want [node0]", got)
		}
	})

	t.Run("no_rank_annotation", func(t *testing.T) {
		tasks := map[api.TaskID]util.NPUTask{
			"t0": {PodStatus: v1.PodPending, Annotation: map[string]string{}},
		}
		got := mh.getMissingNodesFromJob(superPods, tasks)
		if len(got) != 0 {
			t.Errorf("len=%d, want 0", len(got))
		}
	})
}

func TestIsFaultRankZero(t *testing.T) {
	jobs := map[api.JobID]plugin.SchedulerJob{
		"job1": {SuperPods: map[string][]plugin.SuperNode{"0": {{Name: "n0"}}}}}

	t.Run("rank_zero", func(t *testing.T) {
		task := &api.TaskInfo{Name: testTaskName, Job: "job1", Pod: &v1.Pod{}}
		patch := gomonkey.ApplyFunc(getHcclRankIndex, func(*api.TaskInfo, plugin.SchedulerJob) (int, error) { return 0, nil })
		defer patch.Reset()
		if !isFaultRankZero(task, jobs) {
			t.Error("rank 0 should return true")
		}
	})

	t.Run("rank_nonzero", func(t *testing.T) {
		task := &api.TaskInfo{Name: testTaskName, Job: "job1", Pod: &v1.Pod{}}
		patch := gomonkey.ApplyFunc(getHcclRankIndex, func(*api.TaskInfo, plugin.SchedulerJob) (int, error) { return 3, nil })
		defer patch.Reset()
		if isFaultRankZero(task, jobs) {
			t.Error("rank 3 should return false")
		}
	})

	t.Run("job_not_found", func(t *testing.T) {
		task := &api.TaskInfo{Name: testTaskName, Job: "nonexistent", Pod: &v1.Pod{}}
		if isFaultRankZero(task, jobs) {
			t.Error("missing job should return false")
		}
	})
}

func TestGetCachedJobSuperPods(t *testing.T) {
	t.Run("fault_job_priority", func(t *testing.T) {
		mh := newTestHandler()
		fJob := &rescheduling.FaultJob{IsFaultJob: true, SuperPods: map[string][]plugin.SuperNode{"0": {{Name: "fault_node"}}}}
		faultCache := rescheduling.GetReSchedulerCache()
		if faultCache == nil {
			t.Skip("rescheduling cache not initialized")
		}
		faultCache.FaultJobs = map[api.JobID]*rescheduling.FaultJob{"job1": fJob}
		mh.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{
			"job1": {SuperPods: map[string][]plugin.SuperNode{"0": {{Name: "job_node"}}}},
		}
		defer func() { faultCache.FaultJobs = nil }()

		got := mh.getCachedJobSuperPods(&api.TaskInfo{Job: "job1"})
		if len(got) == 0 || got["0"][0].Name != "fault_node" {
			t.Errorf("should prefer FaultJob SuperPods, got %v", got)
		}
	})

	t.Run("falls_back_to_job", func(t *testing.T) {
		mh := newTestHandler()
		mh.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{
			"job1": {SuperPods: map[string][]plugin.SuperNode{"0": {{Name: "job_node"}}}},
		}
		got := mh.getCachedJobSuperPods(&api.TaskInfo{Job: "job1"})
		if len(got) == 0 || got["0"][0].Name != "job_node" {
			t.Errorf("should fall back to job SuperPods, got %v", got)
		}
	})

	t.Run("no_data", func(t *testing.T) {
		mh := newTestHandler()
		got := mh.getCachedJobSuperPods(&api.TaskInfo{Job: "job1"})
		if got != nil {
			t.Errorf("should return nil, got %v", got)
		}
	})
}

func TestScheduleFromAllNodes(t *testing.T) {
	mh := newTestHandlerWithNodes()
	mh.taskLevels = []util.TaskTreeLevel{{Name: "l0", ReqNode: 1}}
	mh.FrameAttr.ResourceLevelsInfo = map[string][]util.ResourceTreeLevel{
		util.DefaultTopoTree: {{Type: util.LevelTypeTree, ReservedNode: 0}, {Type: util.LevelTypeNode, ReservedNode: 0}},
	}
	mh.SuperPodInfo = &plugin.SuperPodInfo{SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{}}
	mh.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{"job1": {}}

	task := &api.TaskInfo{Name: testTaskName, Job: "job1", Pod: &v1.Pod{}}
	sm, err := mh.scheduleFromAllNodes(task, []*api.NodeInfo{{Name: "node0"}})
	if err != nil {
		t.Errorf("scheduleFromAllNodes() unexpected error: %v", err)
	}
	if sm == nil || len(sm) == 0 {
		t.Error("scheduleFromAllNodes() should return non-empty supernode map")
	}
}
