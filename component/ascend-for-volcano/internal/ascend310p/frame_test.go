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
Package ascend310p is using for HuaWei Ascend pin affinity schedule.
*/
package ascend310p

import (
	"errors"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"
	itest "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// TestNew
func TestNew(t *testing.T) {
	t.Run("test New", func(t *testing.T) {
		npu := New(PluginName)
		if npu.GetPluginName() != PluginName {
			t.Errorf("New() npu Name: %s, wantName: %s.", npu.GetPluginName(), PluginName)
		}
		if npu.GetAnnoName() != util.NPU310PCardName {
			t.Errorf("New() npu annoName: %s, wantAnnoName: %s.", npu.GetPluginName(), util.NPU310PCardName)
		}
		if npu.GetAnnoPreVal() != util.NPU310PCardNamePre {
			t.Errorf("New() npu annoNamePre: %s, wantAnnoNamePre: %s.",
				npu.GetPluginName(), util.NPU310PCardNamePre)
		}
	})
}

func buildValidNPUJobTestCase01() []itest.ValidNPUJobTestCase {
	job01 := test.FakeNormalTestJob("job01", 1)
	test.SetFakeJobResRequest(job01, util.NPU310PCardName, "1")
	attr1 := itest.FakeSchedulerJobAttrByJob(job01)
	job02 := test.FakeNormalTestJob("job02", 1)
	test.SetFakeJobResRequest(job02, util.NPU310PCardName, "5")
	attr2 := itest.FakeSchedulerJobAttrByJob(job02)
	job03 := test.FakeNormalTestJob("job02", 1)
	test.SetFakeJobResRequest(job03, util.NPU310PCardName, "2")
	attr3 := itest.FakeSchedulerJobAttrByJob(job03)
	return []itest.ValidNPUJobTestCase{
		{
			Name:    "01-ValidNPUJob should return nil when job request no npu",
			Attr:    attr1,
			WantErr: nil,
		},
		{
			Name:    "02-ValidNPUJob should return error when tasks request npu more than 4",
			Attr:    attr2,
			WantErr: nil,
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
	test.SetFakeJobResRequest(job04, util.NPU310PCardName, "1")
	attr4 := itest.FakeSchedulerJobAttrByJob(job04)
	task := util.NPUTask{ReqNPUNum: 1}
	attr4.Tasks[test.FakeTaskName1] = task
	job05 := test.FakeNormalTestJob("job05", util.NPUIndex2)
	test.SetFakeJobResRequest(job05, util.NPU310PCardName, "5")
	attr5 := itest.FakeSchedulerJobAttrByJob(job05)
	attr5.Tasks[test.FakeTaskName1] = task
	job06 := test.FakeNormalTestJob("job06", util.NPUIndex2)
	test.SetFakeJobResRequest(job06, util.NPU310PCardName, "2")
	attr6 := itest.FakeSchedulerJobAttrByJob(job06)
	return []itest.ValidNPUJobTestCase{
		{
			Name:    "04-ValidNPUJob should return nil when task request no npu",
			Attr:    attr4,
			WantErr: nil,
		},
		{
			Name:    "05-ValidNPUJob should return error when task request npu more than 4",
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

func TestValidNPUJob(t *testing.T) {
	n := New(PluginName)
	npu, ok := n.(*ascend310P)
	if !ok {
		return
	}
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

func buildCheckNodeNPUByTaskTestCases01() []itest.CheckNodeNPUByTaskTestCase {
	return []itest.CheckNodeNPUByTaskTestCase{
		{
			Name: "01-CheckNodeNPUByTask return err when task is nil",
			Task: nil,
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1,Ascend310P-3,Ascend310P-4"},
				},
			},
			Attr:    util.SchedulerJobAttr{NPUJob: &util.NPUJob{VJob: &util.VJob{Type: util.JobTypeWhole}}},
			WantErr: errors.New(util.ArgumentError),
		},
		{
			Name: "02-CheckNodeNPUByTask return err when node annotation is nil",
			Task: test.FakeTaskWithResReq("pod1", util.NPU310PCardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: nil,
				},
			},
			Attr:    util.SchedulerJobAttr{NPUJob: &util.NPUJob{VJob: &util.VJob{Type: util.JobTypeWhole}}},
			WantErr: errors.New(util.ArgumentError),
		},
		{
			Name: "03-CheckNodeNPUByTask return err when Vjob is nil",
			Task: test.FakeTaskWithResReq("pod1", util.NPU310PCardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1,Ascend310P-3,Ascend310P-4"},
				},
			},
			Attr:    util.SchedulerJobAttr{NPUJob: &util.NPUJob{}},
			WantErr: errors.New("task<pod1> is not npu task"),
		},
	}
}

func buildCheckNodeNPUByTaskTestCases02() []itest.CheckNodeNPUByTaskTestCase {
	return []itest.CheckNodeNPUByTaskTestCase{
		{
			Name: "04-CheckNodeNPUByTask return err when tp.Type is util.JobTypeStCut",
			Task: test.FakeTaskWithResReq("pod1", util.NPU310PCardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1,Ascend310P-3,Ascend310P-4"},
				},
			},
			Attr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{
				test.FakeTaskName1: {VTask: &util.VTask{Type: util.JobTypeStCut}},
			},
				VJob: &util.VJob{Type: util.JobTypeStCut}}},
			WantErr: errors.New("rescheduling CheckNodeNPUByTask invalid argument"),
		},
		{
			Name: "05-CheckNodeNPUByTask return err when ty.Type is util.JobTypeDyCut",
			Task: test.FakeTaskWithResReq("pod1", util.NPU310PCardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1,Ascend310P-3,Ascend310P-4"},
				},
			},
			Attr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{
				test.FakeTaskName1: {VTask: &util.VTask{Type: util.JobTypeDyCut}},
			},
				VJob: &util.VJob{Type: util.JobTypeDyCut}}},
			WantErr: errors.New("task pod1 AscendNPUCore read failed"),
		},
		{
			Name: "06-CheckNodeNPUByTask return err when ty.Type is other",
			Task: test.FakeTaskWithResReq("pod1", util.NPU310PCardName, util.NPUIndex4),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1,Ascend310P-3,Ascend310P-4"},
				},
			},
			Attr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{
				test.FakeTaskName1: {VTask: &util.VTask{Type: 3}},
			}, VJob: &util.VJob{Type: 3}}},
			WantErr: errors.New(" no type 3"),
		},
	}
}

// TestCheckNodeNPUByTask
func TestCheckNodeNPUByTask(t *testing.T) {
	n := New(PluginName)
	npu, ok := n.(*ascend310P)
	if !ok {
		return
	}
	testCases := buildCheckNodeNPUByTaskTestCases01()
	testCases = append(testCases, buildCheckNodeNPUByTaskTestCases02()...)
	for _, tt := range testCases {
		npu.SchedulerJobAttr = tt.Attr
		npu.SetSchedulerEnv(plugin.ScheduleEnv{
			Jobs: map[api.JobID]plugin.SchedulerJob{
				test.FakeJobName: {SchedulerJobAttr: tt.Attr},
			},
		})
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.CheckNodeNPUByTask(tt.Task, tt.Node); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

func buildScoreBestNPUNodesTestCases01() []itest.ScoreBestNPUNodesTestCase {
	return []itest.ScoreBestNPUNodesTestCase{
		{
			Name:     "01-ScoreBestNPUNodes return err when task is nil",
			Task:     nil,
			Nodes:    []*api.NodeInfo{test.FakeNormalTestNode("node1")},
			ScoreMap: map[string]float64{"node1": 0},
			WantSMap: map[string]float64{"node1": 0},
			Attr:     util.SchedulerJobAttr{NPUJob: &util.NPUJob{}},
			WantErr:  errors.New(util.ArgumentError),
		},
		{
			Name:     "02-ScoreBestNPUNodes return err when nodes is empty",
			Task:     test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Nodes:    []*api.NodeInfo{},
			ScoreMap: map[string]float64{"node1": 0},
			WantSMap: map[string]float64{"node1": 0},
			Attr:     util.SchedulerJobAttr{NPUJob: &util.NPUJob{}},
			WantErr:  errors.New(util.ArgumentError),
		},
		{
			Name:     "03-ScoreBestNPUNodes return err when scoreMap is empty",
			Task:     test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Nodes:    []*api.NodeInfo{test.FakeNormalTestNode("node1")},
			ScoreMap: map[string]float64{},
			WantSMap: map[string]float64{},
			Attr:     util.SchedulerJobAttr{NPUJob: &util.NPUJob{}},
			WantErr:  errors.New(util.ArgumentError),
		},
		{
			Name:     "04-ScoreBestNPUNodes return nil when tp.VJob is nil",
			Task:     test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Nodes:    []*api.NodeInfo{test.FakeNormalTestNode("node1")},
			ScoreMap: map[string]float64{"node1": 0},
			WantSMap: map[string]float64{"node1": 0},
			Attr:     util.SchedulerJobAttr{NPUJob: &util.NPUJob{}},
			WantErr:  nil,
		},
	}
}

func buildScoreBestNPUNodesTestCases02() []itest.ScoreBestNPUNodesTestCase {
	return []itest.ScoreBestNPUNodesTestCase{
		{
			Name:     "05-ScoreBestNPUNodes return nil when tp.Type is JobTypeWhole",
			Task:     test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Nodes:    []*api.NodeInfo{test.FakeNormalTestNode("node1")},
			ScoreMap: map[string]float64{"node1": 0},
			WantSMap: map[string]float64{"node1": 0},
			Attr:     util.SchedulerJobAttr{NPUJob: &util.NPUJob{VJob: &util.VJob{Type: util.JobTypeWhole}}},
			WantErr:  nil,
		},
		{
			Name:     "06-ScoreBestNPUNodes return nil when tp.Type is JobTypeStCut",
			Task:     test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Nodes:    []*api.NodeInfo{test.FakeNormalTestNode("node1")},
			ScoreMap: map[string]float64{"node1": 0},
			WantSMap: map[string]float64{"node1": 0},
			Attr:     util.SchedulerJobAttr{NPUJob: &util.NPUJob{VJob: &util.VJob{Type: util.JobTypeStCut}}},
			WantErr:  nil,
		},
		{
			Name:     "07-ScoreBestNPUNodes return nil when tp.Type is JobTypeDyCut",
			Task:     test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Nodes:    []*api.NodeInfo{test.FakeNormalTestNode("node1")},
			ScoreMap: map[string]float64{"node1": 0},
			WantSMap: map[string]float64{"node1": 0},
			Attr:     util.SchedulerJobAttr{NPUJob: &util.NPUJob{VJob: &util.VJob{Type: util.JobTypeDyCut}}},
			WantErr:  nil,
		},
		{
			Name:     "08-ScoreBestNPUNodes return nil when tp.Type is other",
			Task:     test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Nodes:    []*api.NodeInfo{test.FakeNormalTestNode("node1")},
			ScoreMap: map[string]float64{"node1": 0},
			WantSMap: map[string]float64{"node1": 0},
			Attr:     util.SchedulerJobAttr{NPUJob: &util.NPUJob{VJob: &util.VJob{Type: 4}}},
			WantErr:  errors.New(" no type 4"),
		},
	}
}

func TestScoreBestNPUNodes(t *testing.T) {
	n := New(PluginName)
	npu, ok := n.(*ascend310P)
	if !ok {
		return
	}
	testCases := buildScoreBestNPUNodesTestCases01()
	testCases = append(testCases, buildScoreBestNPUNodesTestCases02()...)
	for _, tt := range testCases {
		npu.SchedulerJobAttr = tt.Attr
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.ScoreBestNPUNodes(tt.Task, tt.Nodes, tt.ScoreMap); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("ScoreBestNPUNodes() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

func buildUseAnnotationTestCases01() []itest.UseAnnotationTestCase {
	return []itest.UseAnnotationTestCase{
		{
			Name: "01-UseAnnotation return nil when task is nil",
			Task: nil,
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
				},
			},
			Attr:     util.SchedulerJobAttr{NPUJob: &util.NPUJob{}},
			WantNode: nil,
		},
		{
			Name: "02-UseAnnotation return nil when node annotation is nil",
			Task: test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: nil,
				},
			},
			Attr:     util.SchedulerJobAttr{NPUJob: &util.NPUJob{}},
			WantNode: nil,
		},
		{
			Name: "03-UseAnnotation return nil when tp.VJob is nil",
			Task: test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1,Ascend310P-3,Ascend310P-4"},
				},
			},
			Attr:     util.SchedulerJobAttr{NPUJob: &util.NPUJob{}},
			WantNode: nil,
		},
	}
}

func buildUseAnnotationTestCases02() []itest.UseAnnotationTestCase {
	return []itest.UseAnnotationTestCase{
		{
			Name: "04-UseAnnotation return nil when tp.VJob is JobTypeWhole",
			Task: test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1,Ascend310P-3,Ascend310P-4"},
				},
			},
			Attr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{
				test.FakeTaskName1: {VTask: &util.VTask{Type: util.JobTypeWhole}},
			}, VJob: &util.VJob{Type: util.JobTypeWhole}}},
			WantNode: nil,
		},
		{
			Name: "05-UseAnnotation return node when tp.VJob is JobTypeStCut",
			Task: test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1,Ascend310P-3,Ascend310P-4"},
				},
			},
			Attr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{VJob: &util.VJob{Type: util.JobTypeStCut}}},
			WantNode: &plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1,Ascend310P-3,Ascend310P-4"},
				},
			},
		},
		{
			Name: "06-UseAnnotation return node when tp.VJob is JobTypeWhole",
			Task: test.FakeNormalTestTask("pod1", "node1", "vcjob"),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1,Ascend310P-3,Ascend310P-4"},
				},
			},
			Attr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{VJob: &util.VJob{Type: util.JobTypeDyCut}}},
			WantNode: &plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1,Ascend310P-3,Ascend310P-4"},
				},
			},
		},
	}
}

// TestUseAnnotation
func TestUseAnnotation(t *testing.T) {
	n := New(PluginName)
	npu, ok := n.(*ascend310P)
	if !ok {
		return
	}
	testCases := buildUseAnnotationTestCases01()
	testCases = append(testCases, buildUseAnnotationTestCases02()...)
	for _, tt := range testCases {
		npu.SchedulerJobAttr = tt.Attr
		env := plugin.ScheduleEnv{
			Jobs: map[api.JobID]plugin.SchedulerJob{
				"vcjob/vcjob": {SchedulerJobAttr: tt.Attr},
			},
		}
		npu.SetSchedulerEnv(env)
		t.Run(tt.Name, func(t *testing.T) {
			if got := npu.UseAnnotation(tt.Task, tt.Node); !reflect.DeepEqual(got, tt.WantNode) {
				t.Errorf("CheckNodeNPUByTask() got = %v, wantNode %v", got, tt.WantNode)
			}
		})
	}
}

func buildReleaseAnnotationTestCase01() itest.ReleaseAnnotationTestCase {
	test1 := itest.ReleaseAnnotationTestCase{
		Name: "01-ReleaseAnnotation return node when job is not VJob",
		Task: test.FakeNormalTestTask("pod1", "node1", "vcjob"),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
			},
		},
		Attr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{}},
		WantNode: &plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
			},
		},
	}
	return test1
}

func buildReleaseAnnotationTestCase02() itest.ReleaseAnnotationTestCase {
	test2 := itest.ReleaseAnnotationTestCase{
		Name: "02-ReleaseAnnotation return node  when type  is util.JobTypeWhole",
		Task: test.FakeNormalTestTask("pod1", "node1", "vcjob"),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
			},
		},
		Attr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{VJob: &util.VJob{Type: util.JobTypeWhole}}},
		WantNode: &plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
			},
		},
	}
	return test2
}

func buildReleaseAnnotationTestCase03() itest.ReleaseAnnotationTestCase {
	test3 := itest.ReleaseAnnotationTestCase{
		Name: "03-ReleaseAnnotation return node  when type  is util.JobTypeDyCut",
		Task: test.FakeNormalTestTask("pod1", "node1", "vcjob"),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
			},
		},
		Attr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{VJob: &util.VJob{Type: util.JobTypeDyCut}}},
		WantNode: &plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
			},
		},
	}
	return test3
}

func buildReleaseAnnotationTestCase04() itest.ReleaseAnnotationTestCase {
	test4 := itest.ReleaseAnnotationTestCase{
		Name: "04-ReleaseAnnotation return node  when type  is other",
		Task: test.FakeNormalTestTask("pod1", "node1", "vcjob"),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
			},
		},
		Attr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{VJob: &util.VJob{Type: 3}}},
		WantNode: &plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
			},
		},
	}
	return test4
}

func buildReleaseAnnotationTestCase() []itest.ReleaseAnnotationTestCase {
	tests := []itest.ReleaseAnnotationTestCase{
		buildReleaseAnnotationTestCase01(),
		buildReleaseAnnotationTestCase02(),
		buildReleaseAnnotationTestCase03(),
		buildReleaseAnnotationTestCase04(),
	}
	return tests
}

func TestReleaseAnnotation(t *testing.T) {
	n := New(PluginName)
	npu, ok := n.(*ascend310P)
	if !ok {
		return
	}
	testCases := buildReleaseAnnotationTestCase()
	for _, tt := range testCases {
		npu.SchedulerJobAttr = tt.Attr
		t.Run(tt.Name, func(t *testing.T) {
			if got := npu.ReleaseAnnotation(tt.Task, tt.Node); !reflect.DeepEqual(got, tt.WantNode) {
				t.Errorf("CheckNodeNPUByTask() got = %v, wantNode %v", got, tt.WantNode)
			}
		})
	}
}
