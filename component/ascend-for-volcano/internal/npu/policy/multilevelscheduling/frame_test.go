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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
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
	mh := newTestHandler()
	jobReady := true
	job := newTestSchedulerJob()
	job.JobReadyTag = &jobReady
	job.SuperPods = map[string][]plugin.SuperNode{"0": {{Name: "node0"}}}
	mh.ScheduleEnv.Jobs = map[api.JobID]plugin.SchedulerJob{"job1": job}
	task := &api.TaskInfo{Name: testTaskName, Job: "job1", Pod: &v1.Pod{}}
	patch := gomonkey.ApplyFunc(getHcclRankIndex, func(*api.TaskInfo, plugin.SchedulerJob) (int, error) { return 0, nil }).
		ApplyFunc(getL1Ranks, func(map[string][]plugin.SuperNode, int) (string, int, error) { return "0", 0, nil })
	defer patch.Reset()
	err := mh.ScoreBestNPUNodes(task, []*api.NodeInfo{{Name: testNodeName}}, map[string]float64{"node0": 0})
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
	}).ApplyFunc(plugin.GetSuperNodeMapFromTaskTree, func(*util.TaskTree) (map[string][]plugin.SuperNode, error) {
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
	}).ApplyFunc(plugin.GetSuperNodeMapFromTaskTree, func(*util.TaskTree) (map[string][]plugin.SuperNode, error) {
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
	}).ApplyFunc(plugin.GetSuperNodeMapFromTaskTree, func(*util.TaskTree) (map[string][]plugin.SuperNode, error) {
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
	}).ApplyFunc(plugin.GetSuperNodeMapFromTaskTree, func(*util.TaskTree) (map[string][]plugin.SuperNode, error) {
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

func TestTryScheduleTaskInSingleTree(t *testing.T) {
	mh := newTestHandler()
	mh.taskLevels = []util.TaskTreeLevel{{Name: "l0", ReqNode: 1}}
	tree := &util.ResourceTree{Name: "t", ResourceNode: &util.ResourceNode{Name: "r"},
		Levels: []util.ResourceTreeLevel{{Type: util.LevelTypeNode, ReservedNode: 0}}}
	patch := gomonkey.ApplyFunc(Schedule, func(*util.ResourceTree, []util.TaskTreeLevel) (*util.TaskTree, error) {
		return &util.TaskTree{TaskNode: &util.TaskNode{}}, nil
	})
	defer patch.Reset()
	tt, err := mh.tryScheduleTaskInSingleTree(newTestTask(testTaskName), tree)
	if err != nil || tt == nil {
		t.Errorf("tryScheduleTaskInSingleTree() got err=%v, tt=%v", err, tt)
	}
}

func TestTryScheduleTaskInSingleTree_Reschedule(t *testing.T) {
	mh := newTestHandler()
	mh.taskLevels = []util.TaskTreeLevel{{Name: "l0", ReqNode: 1}}
	mh.SuperPodInfo = &plugin.SuperPodInfo{
		SuperPodReschdInfo:        map[api.JobID]map[string][]plugin.SuperNode{},
		SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{"job1": {"t1": "node0"}},
	}
	tree := &util.ResourceTree{Name: "t", ResourceNode: &util.ResourceNode{Name: "r"},
		Levels: []util.ResourceTreeLevel{{Type: util.LevelTypeNode, ReservedNode: 0}}}
	task := &api.TaskInfo{Name: testTaskName, Job: "job1", Pod: &v1.Pod{}}
	patch := gomonkey.ApplyFunc(getFaultJob, func(api.JobID) (*rescheduling.FaultJob, bool) {
		return &rescheduling.FaultJob{IsFaultJob: true, JobUID: "job1"}, true
	}).ApplyFunc(plugin.GetTaskTreeFromSuperNodeMap,
		func(map[string][]plugin.SuperNode, []util.TaskTreeLevel, []util.ResourceTreeLevel, map[string]plugin.NPUNode) (*util.TaskTree, error) {
			return &util.TaskTree{TaskNode: &util.TaskNode{}}, nil
		}).ApplyFunc(Reschedule, func(*util.ResourceTree, *util.TaskTree, []string) (*util.TaskTree, error) {
		return &util.TaskTree{TaskNode: &util.TaskNode{}}, nil
	})
	defer patch.Reset()
	tt, err := mh.tryScheduleTaskInSingleTree(task, tree)
	if err != nil || tt == nil {
		t.Errorf("tryScheduleTaskInSingleTree() got err=%v, tt=%v", err, tt)
	}
}

func TestRescheduleHandler(t *testing.T) {
	mh := newTestHandler()
	mh.taskLevels = []util.TaskTreeLevel{{Name: "l0", ReqNode: 1}}
	mh.SuperPodInfo = &plugin.SuperPodInfo{
		SuperPodReschdInfo:        map[api.JobID]map[string][]plugin.SuperNode{},
		SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{"job1": {"t1": "node0"}},
	}
	tree := &util.ResourceTree{Name: "t", ResourceNode: &util.ResourceNode{Name: "r"},
		Levels: []util.ResourceTreeLevel{{Type: util.LevelTypeNode, ReservedNode: 0}}}
	task := &api.TaskInfo{Name: testTaskName, Job: "job1", Pod: &v1.Pod{}}
	fJob := &rescheduling.FaultJob{IsFaultJob: true, JobUID: "job1", SuperPods: map[string][]plugin.SuperNode{}}
	patch := gomonkey.ApplyFunc(plugin.GetTaskTreeFromSuperNodeMap,
		func(map[string][]plugin.SuperNode, []util.TaskTreeLevel, []util.ResourceTreeLevel, map[string]plugin.NPUNode) (*util.TaskTree, error) {
			return &util.TaskTree{TaskNode: &util.TaskNode{}}, nil
		}).ApplyFunc(Reschedule, func(*util.ResourceTree, *util.TaskTree, []string) (*util.TaskTree, error) {
		return &util.TaskTree{TaskNode: &util.TaskNode{}}, nil
	})
	defer patch.Reset()
	_, err := mh.reschedule(fJob, task, tree)
	if err != nil {
		t.Errorf("reschedule() unexpected error: %v", err)
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

func TestIfPodLevelRescheduling_False(t *testing.T) {
	fJob := &rescheduling.FaultJob{PendingSessionNum: 0}
	sJob := &plugin.SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{"t1": {}}}}}
	sJob.SchedulingTaskNum = 1
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.FaultJob{}), "IsJobSingleRescheduling",
		func(*rescheduling.FaultJob, *plugin.SchedulerJob) bool { return false }).
		ApplyMethod(reflect.TypeOf(&rescheduling.FaultJob{}), "IsProcessReschedulingJob",
			func(*rescheduling.FaultJob, *plugin.SchedulerJob) bool { return false })
	defer patch.Reset()
	if ifPodLevelRescheduling(fJob, sJob) {
		t.Error("ifPodLevelRescheduling() should return false")
	}
}

func TestUpdateSuperNodesForPodLevelRescheduling(t *testing.T) {
	nodes := map[string][]plugin.SuperNode{"0": {{Name: "n0"}, {Name: "n1"}}}
	task := &api.TaskInfo{Name: testTaskName, Job: "job1"}
	job := plugin.SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{"t1": {}}}}}
	patch := gomonkey.ApplyFunc(getFaultJob, func(api.JobID) (*rescheduling.FaultJob, bool) {
		return &rescheduling.FaultJob{IsFaultJob: true, PendingSessionNum: 0, SuperPods: nodes}, true
	}).ApplyFunc(getHcclRankIndex, func(*api.TaskInfo, plugin.SchedulerJob) (int, error) { return 0, nil }).
		ApplyFunc(getL1Ranks, func(map[string][]plugin.SuperNode, int) (string, int, error) { return "0", 0, nil }).
		ApplyMethod(reflect.TypeOf(&rescheduling.FaultJob{}), "IsJobSingleRescheduling",
			func(*rescheduling.FaultJob, *plugin.SchedulerJob) bool { return true })
	defer patch.Reset()
	updateSuperNodesForPodLevelRescheduling(nodes, task, job)
}

func TestUpdateSuperNodesForPodLevelRescheduling_NilNodes(t *testing.T) {
	task := &api.TaskInfo{Name: testTaskName, Job: "job1"}
	job := plugin.SchedulerJob{}
	updateSuperNodesForPodLevelRescheduling(nil, task, job)
}

func TestMultilevelHandler_SetPluginName(t *testing.T) {
	mh := &MultilevelHandler{}
	mh.SetPluginName(testPluginName)
	if mh.GetPluginName() != testPluginName {
		t.Errorf("SetPluginName() failed, got %s", mh.GetPluginName())
	}
}
