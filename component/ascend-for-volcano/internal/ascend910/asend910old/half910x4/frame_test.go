/*
Copyright(C)2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package half910x4 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.
*/
package half910x4

import (
	"errors"
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	itest "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// TestNew
func TestNew(t *testing.T) {
	t.Run("test New", func(t *testing.T) {
		npu := New(SchedulerName)
		if npu.GetPluginName() != SchedulerName {
			t.Errorf("New() npu Name: %s, wantName: %s.", npu.GetPluginName(), SchedulerName)
		}
		if npu.GetAnnoName() != util.NPU910CardName {
			t.Errorf("New() npu annoName: %s, wantAnnoName: %s.", npu.GetPluginName(), util.NPU910CardName)
		}
		if npu.GetAnnoPreVal() != util.NPU910CardNamePre {
			t.Errorf("New() npu annoNamePre: %s, wantAnnoNamePre: %s.",
				npu.GetPluginName(), util.NPU910CardNamePre)
		}
	})
}

func buildValidNPUJobTestCase01() []itest.ValidNPUJobTestCase {
	job01 := test.FakeNormalTestJob("job01", 1)
	test.SetFakeJobResRequest(job01, util.NPU910CardName, "0")
	attr1 := itest.FakeSchedulerJobAttrByJob(job01)
	job02 := test.FakeNormalTestJob("job02", 1)
	test.SetFakeJobResRequest(job02, util.NPU910CardName, "3")
	attr2 := itest.FakeSchedulerJobAttrByJob(job02)
	job03 := test.FakeNormalTestJob("job03", 1)
	test.SetFakeJobResRequest(job03, util.NPU910CardName, "2")
	attr3 := itest.FakeSchedulerJobAttrByJob(job03)
	errMsg1 := "huawei.com/Ascend910half checkSingleTrainMode vcjob/job01 req npu not in [1,2,4]"
	errMsg2 := "huawei.com/Ascend910half checkSingleTrainMode vcjob/job02 req npu not in [1,2,4]"
	return []itest.ValidNPUJobTestCase{
		{
			Name: "01-ValidNPUJob should return error when job request no npu",
			Attr: attr1,
			WantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  errMsg1,
				Message: errMsg1,
			},
		},
		{
			Name: "02-ValidNPUJob should return error when tasks request is not 1-2-4",
			Attr: attr2,
			WantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  errMsg2,
				Message: errMsg2,
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
	const npuNum5 = 5
	job04 := test.FakeNormalTestJob("job04", util.NPUIndex2)
	test.SetFakeJobResRequest(job04, util.NPU910CardName, "4")
	attr4 := itest.FakeSchedulerJobAttrByJob(job04)
	task1 := util.NPUTask{Name: "pod0", ReqNPUNum: npuNum5}
	attr4.Tasks[test.FakeTaskName0] = task1
	job05 := test.FakeNormalTestJob("job05", util.NPUIndex2)
	test.SetFakeJobResRequest(job05, util.NPU910CardName, "4")
	attr5 := itest.FakeSchedulerJobAttrByJob(job05)
	task2 := util.NPUTask{Name: "pod0", ReqNPUNum: 1}
	attr5.Tasks[test.FakeTaskName0] = task2
	job06 := test.FakeNormalTestJob("job06", util.NPUIndex2)
	test.SetFakeJobResRequest(job06, util.NPU910CardName, "4")
	attr6 := itest.FakeSchedulerJobAttrByJob(job06)
	return []itest.ValidNPUJobTestCase{
		{
			Name:    "04-ValidNPUJob should return nil when task request npu in 1-2-4",
			Attr:    attr4,
			WantErr: nil,
		},
		{
			Name:    "05-ValidNPUJob should return nil when task request npu in 1-2-4",
			Attr:    attr5,
			WantErr: nil,
		},
		{
			Name:    "06-ValidNPUJob should return nil when tasks request is valid",
			Attr:    attr6,
			WantErr: nil,
		},
	}
}

// TestValidNPUJob
func TestValidNPUJob(t *testing.T) {
	npu := New(SchedulerName)
	testCases := buildValidNPUJobTestCase02()
	testCases = append(testCases, buildValidNPUJobTestCase01()...)
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			npu.SetSchedulerAttr(tt.Attr)
			if err := npu.ValidNPUJob(); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("ValidNPUJob() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

func buildCheckNodeNPUByTaskTestCases01() []itest.CheckNodeNPUByTaskTestCase {
	return []itest.CheckNodeNPUByTaskTestCase{
		{
			Name: "01-CheckNodeNPUByTask return nil when node npu meet task req",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3",
						networkUnhealthyNPU: "Ascend910-0",
					},
				},
			},
			WantErr: nil,
		},
		{
			Name: "02-CheckNodeNPUByTask return err when task is not npu task",
			Task: test.FakeTaskWithResReq("pod1", util.NPU910CardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3"},
				},
			},
			WantErr: errors.New("task<pod1> is not npu task"),
		},
		{
			Name: "03-CheckNodeNPUByTask return err when node has no req npu",
			Task: test.FakeTaskWithResReq("pod0", util.NPU310PCardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-1,Ascend310-2"},
				},
			},
			WantErr: errors.New("getUsableTopFromNode node1 don't have huawei.com/Ascend910"),
		},
	}

}

func buildCheckNodeNPUByTaskTestCases02() []itest.CheckNodeNPUByTaskTestCase {
	return []itest.CheckNodeNPUByTaskTestCase{
		{
			Name: "04-CheckNodeNPUByTask return err when node has no req npu",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: map[string]string{util.NPU910CardName: "Ascend910-0, Ascend910-1"},
				},
			},
			WantErr: errors.New("npu topology not meet job require,network unhealthy card is [  ]"),
		},
		{
			Name: "05-CheckNodeNPUByTask return err when node has no req npu",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-4",
					},
				},
			},
			WantErr: errors.New("npu topology not meet job require,network unhealthy card is [  ]"),
		},
		{
			Name: "06-CheckNodeNPUByTask return err when task is nil",
			Task: nil,
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
				},
			},
			WantErr: errors.New(util.ArgumentError),
		},
		{
			Name: "07-CheckNodeNPUByTask return err when node annotation is nil",
			Task: test.FakeTaskWithResReq("pod1", util.NPU310PCardName, util.NPUIndex2),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: nil,
				},
			},
			WantErr: errors.New("node<node1> annotation is empty"),
		},
	}
}

func buildCheckNodeNPUByTaskTestCases03() []itest.CheckNodeNPUByTaskTestCase {
	return []itest.CheckNodeNPUByTaskTestCase{
		{
			Name: "01-CheckNodeNPUByTask return nil when node has enough npu",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3",
						networkUnhealthyNPU: ""},
				},
			},
			WantErr: nil,
		},
		{
			Name: "02-CheckNodeNPUByTask return err when node has no enough npu",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-2",
						networkUnhealthyNPU: ""},
				},
			},
			WantErr: errors.New("npu topology not meet job require,network unhealthy card is [  ]"),
		},
		{
			Name: "03-CheckNodeNPUByTask return err when node has no enough npu",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3",
						networkUnhealthyNPU: "Ascend910-0"},
				},
			},
			WantErr: errors.New("npu topology not meet job require,network unhealthy card is [ Ascend910-0 ]"),
		},
	}
}

// TestCheckNodeNPUByTask
func TestCheckNodeNPUByTask(t *testing.T) {
	npu := New(SchedulerName)
	job := test.FakeNormalTestJob("job", 1)
	test.SetFakeJobResRequest(job, util.NPU910CardName, "4")
	attr := itest.FakeSchedulerJobAttrByJob(job)
	sJob := plugin.SchedulerJob{}
	sJob.SchedulerJobAttr = attr
	env := plugin.ScheduleEnv{
		Jobs: map[api.JobID]plugin.SchedulerJob{job.UID: sJob},
	}
	npu.SetSchedulerAttr(attr)
	npu.SetSchedulerEnv(env)
	testCases := buildCheckNodeNPUByTaskTestCases01()
	testCases = append(testCases, buildCheckNodeNPUByTaskTestCases02()...)
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.CheckNodeNPUByTask(tt.Task, tt.Node); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

// TestCheckNodeNPUByTaskDis
func TestCheckNodeNPUByTaskDis(t *testing.T) {
	npu := New(SchedulerName)
	job := test.FakeNormalTestJob("job", util.NPUIndex2)
	test.SetFakeJobResRequest(job, util.NPU910CardName, "4")
	attr := itest.FakeSchedulerJobAttrByJob(job)
	sJob := plugin.SchedulerJob{}
	sJob.SchedulerJobAttr = attr
	env := plugin.ScheduleEnv{
		Jobs: map[api.JobID]plugin.SchedulerJob{job.UID: sJob},
	}
	npu.SetSchedulerAttr(attr)
	npu.SetSchedulerEnv(env)
	testCases := buildCheckNodeNPUByTaskTestCases03()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.CheckNodeNPUByTask(tt.Task, tt.Node); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("ValidNPUJob() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

func buildScoreBestNPUNodesTestCases01() []itest.ScoreBestNPUNodesTestCase {
	return []itest.ScoreBestNPUNodesTestCase{
		{
			Name:     "01-ScoreBestNPUNodes return err when task is not this job npu task ",
			Task:     test.FakeTaskWithResReq("pod1", util.NPU910CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node1"}, {Name: "node2"}},
			ScoreMap: map[string]float64{"node1": 0, "node2": 0},
			WantSMap: map[string]float64{"node1": 0, "node2": 0},
			WantErr:  errors.New("task<pod1> is not npu task"),
		},
		{
			Name:     "02-ScoreBestNPUNodes scoreMap no refresh when node is not this job npu node",
			Task:     test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node6"}},
			ScoreMap: map[string]float64{"node6": 0},
			WantSMap: map[string]float64{"node6": 0},
			WantErr:  nil,
		},
		{
			Name:     "03-ScoreBestNPUNodes scoreMap no refresh when node netUnhealthyNPU not define",
			Task:     test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node7"}},
			ScoreMap: map[string]float64{"node7": 0},
			WantSMap: map[string]float64{"node7": 0},
			WantErr:  nil,
		},
	}
}

func buildScoreBestNPUNodesTestCases02() []itest.ScoreBestNPUNodesTestCase {
	const (
		score104 = 104
		score112 = 112
		score120 = 120
		score128 = 128
	)
	return []itest.ScoreBestNPUNodesTestCase{
		{
			Name:     "04-ScoreBestNPUNodes scoreMap no refresh when node has no npu",
			Task:     test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node8"}},
			ScoreMap: map[string]float64{"node8": 0},
			WantSMap: map[string]float64{"node8": 0},
			WantErr:  nil,
		},
		{
			Name:     "05-ScoreBestNPUNodes return nil when node npu meet task req",
			Task:     test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
			Nodes:    []*api.NodeInfo{{Name: "node1"}, {Name: "node3"}, {Name: "node4"}, {Name: "node5"}},
			ScoreMap: map[string]float64{"node1": 0, "node3": 0, "node4": 0, "node5": 0},
			WantSMap: map[string]float64{"node1": score128, "node3": score120, "node4": score112, "node5": score104},
			WantErr:  nil,
		},
	}
}

func buildFakeScheduleEnv() plugin.ScheduleEnv {
	const allocateNPUNum4 = 4
	return plugin.ScheduleEnv{
		Nodes: map[string]plugin.NPUNode{
			"node1": {
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU910CardName: "Ascend910-0", networkUnhealthyNPU: ""},
					Allocate:   map[v1.ResourceName]float64{util.NPU910CardName: allocateNPUNum4 * util.NPUHexKilo},
				},
			},
			"node2": {
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1"},
				},
			},
			"node3": {
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-2",
						networkUnhealthyNPU: ""},
					Allocate: map[v1.ResourceName]float64{util.NPU910CardName: allocateNPUNum4 * util.NPUHexKilo}},
			},
			"node4": {
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1",
						networkUnhealthyNPU: ""},
					Allocate: map[v1.ResourceName]float64{util.NPU910CardName: allocateNPUNum4 * util.NPUHexKilo}},
			},
			"node5": {CommonNode: plugin.CommonNode{
				Annotation: map[string]string{util.NPU910CardName: "Ascend910-0,Ascend910-1,Ascend910-2," +
					"Ascend910-3", networkUnhealthyNPU: ""},
				Allocate: map[v1.ResourceName]float64{util.NPU910CardName: allocateNPUNum4 * util.NPUHexKilo}},
			},
			"node6": {CommonNode: plugin.CommonNode{Annotation: map[string]string{}}},
			"node7": {CommonNode: plugin.CommonNode{Annotation: map[string]string{util.NPU910CardName: "Ascend910-0"}}},
			"node8": {CommonNode: plugin.CommonNode{Annotation: map[string]string{util.NPU910CardName: "",
				networkUnhealthyNPU: ""}}},
			"node9": {CommonNode: plugin.CommonNode{Annotation: map[string]string{util.NPU910CardName: "",
				networkUnhealthyNPU: ""}}},
		},
	}
}

// TestCheckNodeNPUByTask
func TestScoreBestNPUNodes(t *testing.T) {
	npu := New(SchedulerName)
	job := test.FakeNormalTestJob("job", 1)
	test.SetFakeJobResRequest(job, util.NPU910CardName, "1")
	attr := itest.FakeSchedulerJobAttrByJob(job)
	npu.SetSchedulerAttr(attr)
	env := buildFakeScheduleEnv()
	env.Jobs = map[api.JobID]plugin.SchedulerJob{test.FakeJobName: {SchedulerJobAttr: attr}}
	npu.SetSchedulerEnv(env)
	testCases := buildScoreBestNPUNodesTestCases01()
	testCases = append(testCases, buildScoreBestNPUNodesTestCases02()...)
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
			Name: "01-UseAnnotation task will select the npu which is the only one on the ring",
			Task: test.FakeTaskWithResReq("pod0", util.NPU910CardName, 1),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU910CardName: "Ascend910-0",
						networkUnhealthyNPU: ""},
				},
			},
			PodAnno: "Ascend910-0",
			WantNode: &plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU910CardName: "", networkUnhealthyNPU: ""},
				},
			},
		},
	}
}

// TestUseAnnotation
func TestUseAnnotation(t *testing.T) {
	npu := New(SchedulerName)
	job := test.FakeNormalTestJob("job", 1)
	test.SetFakeJobResRequest(job, util.NPU910CardName, "1")
	attr := itest.FakeSchedulerJobAttrByJob(job)
	npu.SetSchedulerAttr(attr)
	env := plugin.ScheduleEnv{
		Jobs: map[api.JobID]plugin.SchedulerJob{test.FakeJobName: {SchedulerJobAttr: attr}},
	}
	npu.SetSchedulerEnv(env)
	testCases := buildUseAnnotationTestCases01()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			node := npu.UseAnnotation(tt.Task, tt.Node)
			if tt.Task != nil && tt.Node.Annotation != nil && (!reflect.DeepEqual(node.Annotation,
				tt.Node.Annotation)) || !reflect.DeepEqual(tt.Task.Pod.Annotations[util.NPU910CardName], tt.PodAnno) {
				t.Errorf("UseAnnotation() node: %v, wantNode: %v, anno %v, wantAnno %v",
					node, tt.WantNode, tt.Task.Pod.Annotations, tt.PodAnno)
			}
			if (tt.Task == nil || tt.Node.Annotation == nil) || !reflect.DeepEqual(node, tt.WantNode) {
				t.Errorf("UseAnnotation() node: %v, wantNode: %v", node, tt.WantNode)
			}
		})
	}
}
