/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package card310x4 is using for HuaWei A300T Ascend pin affinity schedule.
*/
package card310x4

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	itest "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const (
	npuNum4 = 4
)

func buildValidNPUJobTestCase01() []itest.ValidNPUJobTestCase {
	job01 := test.FakeNormalTestJob("job01", 1)
	test.SetFakeJobResRequest(job01, util.NPU310CardName, "1")
	attr1 := itest.FakeSchedulerJobAttrByJob(job01)
	job02 := test.FakeNormalTestJob("job02", 1)
	test.SetFakeJobResRequest(job02, util.NPU310CardName, "5")
	attr2 := itest.FakeSchedulerJobAttrByJob(job02)
	job03 := test.FakeNormalTestJob("job02", 1)
	test.SetFakeJobResRequest(job03, util.NPU310CardName, "2")
	attr3 := itest.FakeSchedulerJobAttrByJob(job03)
	return []itest.ValidNPUJobTestCase{
		{
			Name:    "01-ValidNPUJob should return nil when job request no npu",
			Attr:    attr1,
			WantErr: nil,
		},
		{
			Name: "02-ValidNPUJob should return error when tasks request npu more than 4",
			Attr: attr2,
			WantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  "job require npu num is invalid",
				Message: "task <vcjob/job02-pod0> req npu <5> is invalid",
			},
		},
		{
			Name:    "03-ValidNPUJob should return nil when tasks request is valid",
			Attr:    attr3,
			WantErr: nil,
		},
	}
}

func buildValidNPUJobTestCase02() []itest.ValidNPUJobTestCase {
	job04 := test.FakeNormalTestJob("job04", util.NPUIndex2)
	test.SetFakeJobResRequest(job04, util.NPU310CardName, "1")
	attr4 := itest.FakeSchedulerJobAttrByJob(job04)
	task := util.NPUTask{ReqNPUNum: 1}
	attr4.Tasks[test.FakeTaskName1] = task
	job05 := test.FakeNormalTestJob("job05", util.NPUIndex2)
	test.SetFakeJobResRequest(job05, util.NPU310CardName, "5")
	attr5 := itest.FakeSchedulerJobAttrByJob(job05)
	attr5.Tasks[test.FakeTaskName1] = task
	job06 := test.FakeNormalTestJob("job06", util.NPUIndex2)
	test.SetFakeJobResRequest(job06, util.NPU310CardName, "2")
	attr6 := itest.FakeSchedulerJobAttrByJob(job06)
	return []itest.ValidNPUJobTestCase{
		{
			Name:    "04-ValidNPUJob should return nil when task request no npu",
			Attr:    attr4,
			WantErr: nil,
		},
		{
			Name: "05-ValidNPUJob should return error when task request npu more than 4",
			Attr: attr5,
			WantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  "job require npu num is invalid",
				Message: "task <vcjob/job05-pod0> req npu <5> is invalid",
			},
		},
		{
			Name:    "03-ValidNPUJob should return nil when tasks request is valid",
			Attr:    attr6,
			WantErr: nil,
		},
	}
}

func TestValidNPUJob(t *testing.T) {
	npu := New(SchedulerName)
	testCases := buildValidNPUJobTestCase01()
	testCases = append(testCases, buildValidNPUJobTestCase02()...)
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			npu.SetSchedulerAttr(tt.Attr)
			if err := npu.ValidNPUJob(); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("ValidNPUJob() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

func buildCheckNodeNPUByTaskTestCase1() itest.CheckNodeNPUByTaskTestCase {
	return itest.CheckNodeNPUByTaskTestCase{
		Name: "01-CheckNodeNPUByTask when return nil node npu meet task req",
		Task: test.FakeTaskWithResReq("pod0", util.NPU310CardName, util.NPUIndex3),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-1,Ascend310-2"},
			},
		},
		WantErr: nil,
	}
}

func buildCheckNodeNPUByTaskTestCase2() itest.CheckNodeNPUByTaskTestCase {
	return itest.CheckNodeNPUByTaskTestCase{
		Name: "02-CheckNodeNPUByTask return err when task is not npu task",
		Task: test.FakeTaskWithResReq("pod1", util.NPU310CardName, util.NPUIndex3),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-1,Ascend310-2"},
			},
		},
		WantErr: errors.New("task<pod1> is not npu task"),
	}
}

func buildCheckNodeNPUByTaskTestCase3() itest.CheckNodeNPUByTaskTestCase {
	return itest.CheckNodeNPUByTaskTestCase{
		Name: "03-CheckNodeNPUByTask return err when node has no req npu",
		Task: test.FakeTaskWithResReq("pod0", util.NPU310PCardName, util.NPUIndex3),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU310PCardName: "Ascend310-0,Ascend310-1,Ascend310-2"},
			},
		},
		WantErr: errors.New("getUsableTopFromNode node1 don't have huawei.com/Ascend310"),
	}
}

func buildCheckNodeNPUByTaskTestCase4() itest.CheckNodeNPUByTaskTestCase {
	return itest.CheckNodeNPUByTaskTestCase{
		Name: "04-CheckNodeNPUByTask return err when node has no req npu",
		Task: test.FakeTaskWithResReq("pod0", util.NPU310CardName, util.NPUIndex3),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU310CardName: "Ascend310-0, Ascend310-1"},
			},
		},
		WantErr: errors.New("checkNodeNPUByTask the npus on this node don't satisfy the schedulable topology : req" +
			" npu(3) illegal not meet node top<[]>"),
	}
}

func buildCheckNodeNPUByTaskTestCase5() itest.CheckNodeNPUByTaskTestCase {
	return itest.CheckNodeNPUByTaskTestCase{
		Name: "05-CheckNodeNPUByTask return err when node has no req npu",
		Task: test.FakeTaskWithResReq("pod0", util.NPU310CardName, util.NPUIndex3),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-1,Ascend310-4"},
			},
		},
		WantErr: fmt.Errorf("checkNodeNPUByTask the npus on this node don't satisfy the schedulable " +
			"topology : req npu(3) illegal not meet node top<[0 1 4]>"),
	}
}

func buildCheckNodeNPUByTaskTestCases() []itest.CheckNodeNPUByTaskTestCase {
	return []itest.CheckNodeNPUByTaskTestCase{
		buildCheckNodeNPUByTaskTestCase1(),
		buildCheckNodeNPUByTaskTestCase2(),
		buildCheckNodeNPUByTaskTestCase3(),
		buildCheckNodeNPUByTaskTestCase4(),
		buildCheckNodeNPUByTaskTestCase5(),
	}
}

// TestCheckNodeNPUByTask
func TestCheckNodeNPUByTask(t *testing.T) {
	npu := New(SchedulerName)
	job := test.FakeNormalTestJob("job", 1)
	test.SetFakeJobResRequest(job, util.NPU310CardName, "3")
	attr1 := itest.FakeSchedulerJobAttrByJob(job)
	npu.SetSchedulerAttr(attr1)
	npu.SetSchedulerEnv(plugin.ScheduleEnv{
		Jobs: map[api.JobID]plugin.SchedulerJob{test.FakeJobName: {SchedulerJobAttr: attr1}},
	})
	testCases := buildCheckNodeNPUByTaskTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.CheckNodeNPUByTask(tt.Task, tt.Node); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("ValidNPUJob() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

func buildScoreBestNPUNodesTestCases() []itest.ScoreBestNPUNodesTestCase {
	const (
		score32 = 32
		score24 = 24
		score16 = 16
		score8  = 8
	)
	return []itest.ScoreBestNPUNodesTestCase{
		{
			Name:     "01-ScoreBestNPUNodes when return nil node npu meet task req",
			Task:     test.FakeTaskWithResReq("pod1", util.NPU310CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}, {Name: "node3"}, {Name: "node4"}},
			ScoreMap: map[string]float64{"node1": 0, "node2": 0, "node3": 0, "node4": 0},
			WantSMap: map[string]float64{"node1": 0, "node2": 0, "node3": 0, "node4": 0},
			WantErr:  errors.New("task<pod1> is not npu task"),
		},
		{
			Name:     "02-ScoreBestNPUNodes when return nil node npu meet task req",
			Task:     test.FakeTaskWithResReq("pod0", util.NPU310CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}, {Name: "node3"}, {Name: "node4"}},
			ScoreMap: map[string]float64{"node1": 0, "node2": 0, "node3": 0, "node4": 0},
			WantSMap: map[string]float64{"node1": score32, "node2": score16, "node3": score24, "node4": score8},
			WantErr:  nil,
		},
		{
			Name:     "03-ScoreBestNPUNodes return nil when node npu meet task req",
			Task:     test.FakeTaskWithResReq("pod0", util.NPU310CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}, {Name: "node3"}, {Name: "node5"}},
			ScoreMap: map[string]float64{"node1": 0, "node2": 0, "node3": 0, "node5": 0},
			WantSMap: map[string]float64{"node1": score32, "node2": score16, "node3": score24, "node5": 0},
			WantErr:  nil,
		},
		{
			Name:     "04-ScoreBestNPUNodes return err when node npu not meet task req",
			Task:     test.FakeTaskWithResReq("pod0", util.NPU310CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}, {Name: "node3"}, {Name: "node6"}},
			ScoreMap: map[string]float64{"node1": 0, "node2": 0, "node3": 0, "node6": 0},
			WantSMap: map[string]float64{"node1": score32, "node2": score16, "node3": score24, "node6": 0},
			WantErr:  nil,
		},
		{
			Name:     "05-ScoreBestNPUNodes return nil when node npu not meet task req",
			Task:     test.FakeTaskWithResReq("pod0", util.NPU310CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}, {Name: "node3"}, {Name: "node7"}},
			ScoreMap: map[string]float64{"node1": 0, "node2": 0, "node3": 0, "node7": 0},
			WantSMap: map[string]float64{"node1": score32, "node2": score16, "node3": score24, "node7": score32},
			WantErr:  nil,
		},
	}
}

// TestCheckNodeNPUByTask
func TestScoreBestNPUNodes(t *testing.T) {
	npu := New(SchedulerName)
	job := test.FakeNormalTestJob("job", 1)
	test.SetFakeJobResRequest(job, util.NPU310CardName, "1")
	attr := itest.FakeSchedulerJobAttrByJob(job)
	env := plugin.ScheduleEnv{
		Jobs: map[api.JobID]plugin.SchedulerJob{test.FakeJobName: {SchedulerJobAttr: attr}},
		Nodes: map[string]plugin.NPUNode{
			"node1": {CommonNode: plugin.CommonNode{Annotation: map[string]string{util.NPU310CardName: "Ascend310-0"}}},
			"node2": {CommonNode: plugin.CommonNode{Annotation: map[string]string{util.NPU310CardName: "Ascend310-0," +
				"Ascend310-1"}}},
			"node3": {CommonNode: plugin.CommonNode{Annotation: map[string]string{util.NPU310CardName: "Ascend310-0," +
				"Ascend310-1,Ascend310-2"}}},
			"node4": {CommonNode: plugin.CommonNode{Annotation: map[string]string{util.NPU310CardName: "Ascend310-0," +
				"Ascend310-1,Ascend310-2,Ascend310-3"}}},
			"node5": {CommonNode: plugin.CommonNode{Annotation: map[string]string{util.NPU310CardName: "Ascend310-0, " +
				"Ascend310-4"}}},
			"node6": {CommonNode: plugin.CommonNode{Annotation: map[string]string{util.NPU310CardName: ""}}},
			"node7": {CommonNode: plugin.CommonNode{Annotation: map[string]string{util.NPU310CardName: "Ascend310-4"}}},
		},
	}
	npu.SetSchedulerAttr(attr)
	npu.SetSchedulerEnv(env)
	testCases := buildScoreBestNPUNodesTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			err := npu.ScoreBestNPUNodes(tt.Task, tt.Nodes, tt.ScoreMap)
			if !reflect.DeepEqual(err, tt.WantErr) || !reflect.DeepEqual(tt.ScoreMap, tt.WantSMap) {
				t.Errorf("ScoreBestNPUNodes() scoreMap: %v, wantSMap: %v, error = %v, wantErr %v",
					tt.ScoreMap, tt.WantSMap, err, tt.WantErr)
			}
		})
	}
}

func buildUseAnnotationTestCases01() []itest.UseAnnotationTestCase {
	return []itest.UseAnnotationTestCase{
		{
			Name: "01-UseAnnotation task will select the npu which is the only one on the card",
			Task: test.FakeTaskWithResReq("pod0", util.NPU310CardName, 1),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-4,Ascend310-5"},
				},
			},
			PodAnno: "Ascend310-0",
			WantNode: &plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Allocate: map[v1.ResourceName]float64{util.NPU310CardName: 0},
				},
			},
		},
		{
			Name: "02-UseAnnotation task will select the npu which is on the card that has 3 npu other than 2",
			Task: test.FakeTaskWithResReq("pod0", util.NPU310CardName, 1),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-1,Ascend310-2,Ascend310-4," +
						"Ascend310-5"},
				},
			},
			PodAnno: "Ascend310-0",
			WantNode: &plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Allocate: map[v1.ResourceName]float64{util.NPU310CardName: npuNum4 * util.NPUHexKilo},
				},
			},
		},
	}
}

func buildUseAnnotationTestCases02() []itest.UseAnnotationTestCase {
	return []itest.UseAnnotationTestCase{
		{
			Name: "03-UseAnnotation task will select the npu which is on the card that has 3 npu other than 2",
			Task: test.FakeTaskWithResReq("pod0", util.NPU310CardName, 1),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-1,Ascend310-4,Ascend310-5," +
						"Ascend310-6"},
				},
			},
			PodAnno: "Ascend310-4",
			WantNode: &plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-1,Ascend310-5,Ascend310-6"},
				},
			},
		},
		{
			Name: "04-UseAnnotation task will select the npu which is on the card that has 2 npu other than 4",
			Task: test.FakeTaskWithResReq("pod0", util.NPU310CardName, 1),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-1,Ascend310-4,Ascend310-5," +
						"Ascend310-6,Ascend310-7"},
				},
			},
			PodAnno: "Ascend310-0",
			WantNode: &plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310CardName: "Ascend310-1,Ascend310-4,Ascend310-5,Ascend310-6," +
						"Ascend310-0,Ascend310-7"},
				},
			},
		},
	}
}

func TestUseAnnotation(t *testing.T) {
	npu := New(SchedulerName)
	job := test.FakeNormalTestJob("job", 1)
	test.SetFakeJobResRequest(job, util.NPU310CardName, "1")
	attr := itest.FakeSchedulerJobAttrByJob(job)
	npu.SetSchedulerAttr(attr)
	env := plugin.ScheduleEnv{Jobs: map[api.JobID]plugin.SchedulerJob{test.FakeJobName: {SchedulerJobAttr: attr}}}
	npu.SetSchedulerEnv(env)
	testCases := buildUseAnnotationTestCases01()
	testCases = append(testCases, buildUseAnnotationTestCases02()...)
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			node := npu.UseAnnotation(tt.Task, tt.Node)
			if !reflect.DeepEqual(node.Annotation, tt.Node.Annotation) ||
				!reflect.DeepEqual(tt.Task.Pod.Annotations[util.NPU310CardName], tt.PodAnno) {
				t.Errorf("UseAnnotation() node: %v, wantNode: %v, anno %v, wantAnno %v",
					node, tt.WantNode, tt.Task.Pod.Annotations, tt.PodAnno)
			}
		})
	}
}
