/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package ascend800ia5stacking provides tests for Huawei Ascend800i-A5 pin affinity scheduling.
*/
package ascend800ia5stacking

import (
	"errors"
	"fmt"

	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	itest "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

const (
	MockTaskNumOne     = 1
	MockResourceNumOne = 1
	MockNeedOne        = "1"
	MockNeedThree      = "3"
	MockJobName        = "job"
	MockPodZeroName    = "pod0"
	MockNodeOneName    = "node1"
	MockNodeType       = "800I-Stacking-A5-8"
	MockFullCards      = "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7"
	MockExCards        = "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6," +
		"Ascend910-7,Ascend910-8"
)

// createValidTask creates a valid task
func createValidTask(name string) *api.TaskInfo {
	return &api.TaskInfo{
		Name: name,
	}
}

func initNPU(attr util.SchedulerJobAttr, env plugin.ScheduleEnv) base.AscendHandler {
	npu := New(SchedulerName)
	npu.SetSchedulerAttr(attr)
	npu.SetSchedulerEnv(env)
	return npu
}

func initAttr(jobName string, taskNum int, need string) (util.SchedulerJobAttr, api.JobID) {
	// 创建一个已有一个task的job，即默认为扩容场景，该task在node0上，占满8卡，为pod0，pg0
	job := test.FakeNormalTestJob(jobName, taskNum)
	// Not sure what this 'need' parameter is for yet
	test.SetFakeJobResRequest(job, util.NPU910CardName, need)
	return itest.FakeSchedulerJobAttrByJob(job), job.UID
}

func initNode1(nodeName string, annoKey string, annoVal string) plugin.NPUNode {
	return plugin.NPUNode{CommonNode: plugin.CommonNode{
		Name:       nodeName,
		Annotation: map[string]string{annoKey: annoVal, util.AcceleratorType: MockNodeType},
	}}
}

func initNode2(nodeName string, annoKey string, annoVal string, unhealthyVal string) plugin.NPUNode {
	return plugin.NPUNode{CommonNode: plugin.CommonNode{
		Name: nodeName,
		Annotation: map[string]string{annoKey: annoVal, util.AcceleratorType: MockNodeType,
			networkUnhealthyNPU: unhealthyVal},
	}}
}

func initStackNode(nodeName string, annoKey string, annoVal string, superPodID int32) plugin.NPUNode {
	return plugin.NPUNode{CommonNode: plugin.CommonNode{
		Name:       nodeName,
		Annotation: map[string]string{annoKey: annoVal, util.AcceleratorType: MockNodeType, "superPodID": string(superPodID)},
	}}
}

func TestNew(t *testing.T) {
	handler := New("testPlugin")
	if handler.GetPluginName() != "testPlugin" {
		t.Errorf("expected plugin name 'testPlugin', got %v", handler.GetPluginName())
	}
	if handler.GetAnnoName() != util.NPU910CardName {
		t.Errorf("expected anno name '%v', got %v", util.NPU910CardName, handler.GetAnnoName())
	}
	if handler.GetAnnoPreVal() != util.NPU910CardNamePre {
		t.Errorf("expected anno pre value '%v', got %v", util.NPU910CardNamePre, handler.GetAnnoPreVal())
	}
}

func TestReleaseAnnotation(t *testing.T) {
	handler := New("testPlugin")
	task := createValidTask("task1")
	node := plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name:       "node1",
			Annotation: map[string]string{"dummy": "value"},
		},
	}
	newNode := handler.ReleaseAnnotation(task, node)
	if newNode == nil {
		t.Error("ReleaseAnnotation returned nil")
	}
	if newNode.Annotation["dummy"] != "value" {
		t.Error("ReleaseAnnotation modified node annotation unexpectedly")
	}
}

func TestUseAnnotation(t *testing.T) {
	attr, _ := initAttr(MockJobName, MockTaskNumOne, MockNeedThree)
	env := plugin.ScheduleEnv{}
	env.Jobs = map[api.JobID]plugin.SchedulerJob{test.FakeJobName: {SchedulerJobAttr: attr}}
	handler := initNPU(attr, env)
	task := test.FakeTaskWithResReq(MockPodZeroName, util.NPU910CardName, util.NPUIndex3)
	node := initStackNode("work1", util.NPU910CardName, MockFullCards, 1)
	newNode := handler.UseAnnotation(task, node)
	if newNode == nil {
		t.Error("UseAnnotation returned nil")
	}
}

// checkNodeNPUByTaskTestCase represents test cases for CheckNodeNPUByTask
type checkNodeNPUByTaskTestCase struct {
	Task    *api.TaskInfo
	Name    string
	Attr    util.SchedulerJobAttr
	Node    plugin.NPUNode
	WantErr error
}

func TestCheckNodeNPUByTask1(t *testing.T) {
	attr, _ := initAttr(MockJobName, MockTaskNumOne, MockNeedThree)
	env := plugin.ScheduleEnv{}
	env.Jobs = map[api.JobID]plugin.SchedulerJob{test.FakeJobName: {SchedulerJobAttr: attr}}
	handler := initNPU(attr, env)
	task := test.FakeTaskWithResReq(MockPodZeroName, util.NPU910CardName, util.NPUIndex3)
	errSuppose := errors.New("example err")
	node := initNode1(MockNodeOneName, util.NPU910CardName,
		"Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,,Ascend910-5,,Ascend910-6,,Ascend910-7")
	patch1 := gomonkey.ApplyMethod(reflect.TypeOf(new(base.NPUHandler)), "GetTaskReqNPUNum",
		func(_ *base.NPUHandler, task *api.TaskInfo) (int, error) { return 0, errSuppose })
	defer patch1.Reset()
	err := handler.CheckNodeNPUByTask(task, node)
	if err != errSuppose {
		t.Errorf("CheckNodeNpuByTask returned not expected err %v", err)
	}
}

func buildCheckNodeNPUByTaskTestCases() []checkNodeNPUByTaskTestCase {
	var exArray = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	unhealthyNode := initNode2(MockNodeOneName, util.NPU910CardName, MockFullCards, MockExCards)
	abnormalNode := initNode1(MockNodeOneName, util.NPU910CardName, MockExCards)
	return []checkNodeNPUByTaskTestCase{
		{
			Name:    "01-getUsableTopFromNode return err when GetUsableTopFromNode return err",
			Task:    test.FakeTaskWithResReq(MockPodZeroName, util.NPU910CardName, util.NPUIndex3),
			Node:    abnormalNode,
			WantErr: fmt.Errorf("node npu top<%v> is invalid", exArray),
		},
		{
			Name:    "02-getUsableTopFromNode return err when node annotation huawei.com/Ascend910-NetworkUnhealthy not exists",
			Task:    test.FakeTaskWithResReq(MockPodZeroName, util.NPU910CardName, util.NPUIndex3),
			Node:    initNode1(MockNodeOneName, util.NPU910CardName, "Ascend910-0, Ascend910-1"),
			WantErr: fmt.Errorf("node<%s> don't have resource<%s>", MockNodeOneName, networkUnhealthyNPU),
		},
		{
			Name:    "03-getUsableTopFromNode return err when networkUnhealthyTop > max",
			Task:    test.FakeTaskWithResReq(MockPodZeroName, util.NPU910CardName, util.NPUIndex3),
			Node:    unhealthyNode,
			WantErr: fmt.Errorf("node<%s> npu networkUnhealthy top<%v> is invalid", MockNodeOneName, exArray),
		},
		{
			Name:    "04-JudgeNodeAndTaskNPU return err when node npu meet task req",
			Task:    test.FakeTaskWithResReq(MockPodZeroName, util.NPU310CardName, util.NPUIndex8),
			Node:    initNode2(MockNodeOneName, util.NPU910CardName, "Ascend910-0", ""),
			WantErr: fmt.Errorf("npu topology not meet job require,network unhealthy card is [  ]"),
		},
	}
}

func TestCheckNodeNPUByTask2(t *testing.T) {
	attr, _ := initAttr(MockJobName, MockTaskNumOne, MockNeedThree)
	env := plugin.ScheduleEnv{}
	env.Jobs = map[api.JobID]plugin.SchedulerJob{test.FakeJobName: {SchedulerJobAttr: attr}}
	npu := initNPU(attr, env)
	npu.(*module800ia5stacking).NPUTaskNum = 2
	testCases := buildCheckNodeNPUByTaskTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.CheckNodeNPUByTask(tt.Task, tt.Node); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

// scoreBestNPUNodesTestCase represents test cases for ScoreBestNPUNodes
type scoreBestNPUNodesTestCase struct {
	Task     *api.TaskInfo
	Nodes    []*api.NodeInfo
	ScoreMap map[string]float64
	WantSMap map[string]float64
	Name     string
	WantErr  error
	Attr     util.SchedulerJobAttr
}

func buildScoreBestNPUNodesTestCases() []scoreBestNPUNodesTestCase {
	const (
		score160 = 160
		score159 = 159
		score158 = 158
		score165 = 165
	)
	return []scoreBestNPUNodesTestCase{
		{
			Name:     "01-ScoreBestNPUNodes when has potential card getBestScoreInComplicatedStack > task require",
			Task:     test.FakeTaskWithResReq("pod0", util.NPU910CardName, MockResourceNumOne),
			Nodes:    []*api.NodeInfo{{Name: "node0"}, {Name: "node1"}},
			ScoreMap: map[string]float64{"node0": 0, "node1": 0},
			WantSMap: map[string]float64{"node0": score160, "node1": score159},
			WantErr:  nil,
		},
		{
			Name:     "02-ScoreBestNPUNodes when return nil node npu meet task req",
			Task:     test.FakeTaskWithResReq("pod0", util.NPU910CardName, MockResourceNumOne),
			Nodes:    []*api.NodeInfo{nil, {Name: "node0"}, {Name: "node1"}, {Name: "node2"}, {Name: "node3"}},
			ScoreMap: map[string]float64{"node0": 0, "node1": 0, "node2": 0, "node3": 0},
			WantSMap: map[string]float64{"node0": score160, "node1": score159, "node2": score158, "node3": score165},
			WantErr:  nil,
		},
	}
}

func initScheduleEnv(attr util.SchedulerJobAttr) plugin.ScheduleEnv {
	return plugin.ScheduleEnv{
		ClusterCache: plugin.ClusterCache{
			Jobs: map[api.JobID]plugin.SchedulerJob{test.FakeJobName: {SchedulerJobAttr: attr}},
			Nodes: map[string]plugin.NPUNode{
				"node0": {CommonNode: plugin.CommonNode{Name: "node0", SuperPodID: 1, Annotation: map[string]string{
					util.NPU910CardName: "Ascend910-0"},
					Allocate: map[v1.ResourceName]float64{"huawei.com/Ascend910": 8000}}},
				"node1": {CommonNode: plugin.CommonNode{Name: "node1", SuperPodID: 1, Annotation: map[string]string{
					util.NPU910CardName: "Ascend910-0,Ascend910-1"},
					Allocate: map[v1.ResourceName]float64{"huawei.com/Ascend910": 8000}}},
				"node2": {CommonNode: plugin.CommonNode{Name: "node2", SuperPodID: 2, Annotation: map[string]string{
					util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-2"},
					Allocate: map[v1.ResourceName]float64{"huawei.com/Ascend910": 8000}}},
				"node3": {CommonNode: plugin.CommonNode{Name: "node3", SuperPodID: 2, Annotation: map[string]string{
					util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3"},
					Allocate: map[v1.ResourceName]float64{"huawei.com/Ascend910": 8000}}},
				"node4": {CommonNode: plugin.CommonNode{Name: "node4", SuperPodID: 3, Annotation: map[string]string{
					util.NPU910CardName: "Ascend910-0,Ascend910-4"},
					Allocate: map[v1.ResourceName]float64{"huawei.com/Ascend910": 8000}}},
				"node5": {CommonNode: plugin.CommonNode{Name: "node5", SuperPodID: 3, Annotation: map[string]string{
					util.NPU910CardName: ""},
					Allocate: map[v1.ResourceName]float64{"huawei.com/Ascend910": 8000}}},
				"node6": {CommonNode: plugin.CommonNode{Name: "node6", SuperPodID: 4, Annotation: map[string]string{
					util.NPU910CardName: "Ascend910-4"},
					Allocate: map[v1.ResourceName]float64{"huawei.com/Ascend910": 8000}}},
				"node7": {CommonNode: plugin.CommonNode{Name: "node7", SuperPodID: 4, Annotation: map[string]string{
					util.NPU910CardName: "Ascend910-4"},
					Allocate: map[v1.ResourceName]float64{"huawei.com/Ascend910": 7000}}},
			},
		},
	}
}

func initNodeCache(env plugin.ScheduleEnv) (map[int32][]plugin.NPUNode, map[string]bool) {
	node0 := env.ClusterCache.Nodes["node0"]
	node1 := env.ClusterCache.Nodes["node1"]
	stacking1 := []plugin.NPUNode{node0, node1}
	node2 := env.ClusterCache.Nodes["node2"]
	node3 := env.ClusterCache.Nodes["node3"]
	stacking2 := []plugin.NPUNode{node2, node3}
	node4 := env.ClusterCache.Nodes["node4"]
	node5 := env.ClusterCache.Nodes["node5"]
	stacking3 := []plugin.NPUNode{node4, node5}
	node6 := env.ClusterCache.Nodes["node6"]
	node7 := env.ClusterCache.Nodes["node7"]
	stacking4 := []plugin.NPUNode{node6, node7}
	return map[int32][]plugin.NPUNode{
			1: stacking1,
			2: stacking2,
			3: stacking3,
			4: stacking4,
		}, map[string]bool{
			"node0": true,
			"node1": true,
			"node2": true,
			"node3": true,
			"node4": true,
			"node5": true,
			"node6": true,
			"node7": true,
		}
}

func initSelectedCache(jobId api.JobID) map[api.JobID]map[int32][]int {
	return map[api.JobID]map[int32][]int{
		jobId: {
			1: {},
			2: {3},
			3: {5, 6},
			4: {7, 8},
		},
	}
}

func TestScoreBestNPUNodes(t *testing.T) {
	handler := New("nil handler")
	t.Run("00-ScoreBestNPUNodes return nil", func(t *testing.T) {
		if err := handler.ScoreBestNPUNodes(nil, nil,
			nil); !reflect.DeepEqual(err, errors.New("invalid argument")) {
			t.Errorf("ScoreBestNPUNodes() = %v, want %v", err, errors.New("invalid argument"))
		}
	})
	attr, jobId := initAttr(MockJobName, MockTaskNumOne, MockNeedOne)
	env := initScheduleEnv(attr)
	npu := initNPU(attr, env)
	npu.(*module800ia5stacking).SuperPodCache, npu.(*module800ia5stacking).PickedNodeCache = initNodeCache(env)

	testCases := buildScoreBestNPUNodesTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			err := npu.ScoreBestNPUNodes(tt.Task, tt.Nodes, tt.ScoreMap)
			if !reflect.DeepEqual(err, tt.WantErr) || !reflect.DeepEqual(tt.ScoreMap, tt.WantSMap) {
				t.Errorf("ScoreBestNPUNodes() scoreMap: %v, wantSMap: %v, error = %v, wantErr %v",
					tt.ScoreMap, tt.WantSMap, err, tt.WantErr)
			}
			npu.(*module800ia5stacking).NPUSelectedCache = initSelectedCache(jobId)
		})
	}
}
