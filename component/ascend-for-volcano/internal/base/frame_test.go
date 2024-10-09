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
Package base is using for HuaWei Ascend pin affinity schedule.
*/
package base

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
	maxNodeNPUNum = 64
	maxCardNPUNum = 4
)

func buildInitMyJobPluginTestCases() []itest.InitMyJobPluginTestCase {
	return []itest.InitMyJobPluginTestCase{
		{
			Name: "01-InitMyJobPlugin return nil",
			Attr: util.SchedulerJobAttr{
				ComJob: util.ComJob{Label: map[string]string{}},
				NPUJob: &util.NPUJob{ReqNPUName: util.NPU310CardName},
			},
			Env:     plugin.ScheduleEnv{},
			WantErr: nil,
		},
	}
}

// TestInitMyJobPlugin
func TestInitMyJobPlugin(t *testing.T) {
	npu := NPUHandler{}
	testCases := buildInitMyJobPluginTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			err := npu.InitMyJobPlugin(tt.Attr, tt.Env)
			if !reflect.DeepEqual(tt.Attr, npu.SchedulerJobAttr) ||
				!reflect.DeepEqual(tt.Env, npu.ScheduleEnv) ||
				!reflect.DeepEqual(tt.WantErr, err) {
				t.Errorf("InitMyJobPlugin() err: %v, wantErr: %v", err, tt.WantErr)
			}
		})
	}
}

func buildValidNPUJobTestCase01() []itest.ValidNPUJobTestCase {
	job01 := test.FakeNormalTestJob("job01", 1)
	test.SetFakeJobResRequest(job01, util.NPU310CardName, "1")
	attr1 := itest.FakeSchedulerJobAttrByJob(job01)
	job02 := test.FakeNormalTestJob("job02", 1)
	test.SetFakeJobResRequest(job02, util.NPU310CardName, "65")
	attr2 := itest.FakeSchedulerJobAttrByJob(job02)
	job03 := test.FakeNormalTestJob("job02", 1)
	test.SetFakeJobResRequest(job03, util.NPU310CardName, "2")
	attr3 := itest.FakeSchedulerJobAttrByJob(job03)
	return []itest.ValidNPUJobTestCase{
		{
			Name:    "01-ValidNPUJob should return error when job request no npu",
			Attr:    attr1,
			WantErr: nil,
		},
		{
			Name: "02-ValidNPUJob should return error when tasks request npu more than 64",
			Attr: attr2,
			WantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  "task req npu num is invalid",
				Message: "task<vcjob/job02-pod0> req npu num<65> is invalid",
			},
		},
		{
			Name:    "03-ValidNPUJob should return nil when tasks request is valid",
			Attr:    attr3,
			WantErr: nil,
		},
	}
}

// TestValidNPUJob
func TestValidNPUJob(t *testing.T) {
	npu := NPUHandler{
		MaxCardNPUNum: maxCardNPUNum,
		MaxNodeNPUNum: maxNodeNPUNum,
	}
	testCases := buildValidNPUJobTestCase01()
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
		Task: test.FakeTaskWithResReq("pod0", util.NPU310PCardName, util.NPUIndex2),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
			},
		},
		WantErr: nil,
	}
}

func buildCheckNodeNPUByTaskTestCase2() itest.CheckNodeNPUByTaskTestCase {
	return itest.CheckNodeNPUByTaskTestCase{
		Name: "02-CheckNodeNPUByTask return err when task is not npu task",
		Task: test.FakeTaskWithResReq("pod1", util.NPU310PCardName, util.NPUIndex2),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
			},
		},
		WantErr: errors.New("task<pod1> is not npu task"),
	}
}

func buildCheckNodeNPUByTaskTestCase3() itest.CheckNodeNPUByTaskTestCase {
	return itest.CheckNodeNPUByTaskTestCase{
		Name: "03-CheckNodeNPUByTask return err when node has no req npu",
		Task: test.FakeTaskWithResReq("pod0", util.NPU310PCardName, util.NPUIndex2),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU910CardName: "Ascend310P-0,Ascend310P-1"},
			},
		},
		WantErr: errors.New("getUsableTopFromNode node1 don't have huawei.com/Ascend310P"),
	}
}

func buildCheckNodeNPUByTaskTestCase4() itest.CheckNodeNPUByTaskTestCase {
	return itest.CheckNodeNPUByTaskTestCase{
		Name: "04-CheckNodeNPUByTask return err when node has no req npu",
		Task: test.FakeTaskWithResReq("pod0", util.NPU310PCardName, util.NPUIndex2),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0, Ascend310P-1"},
			},
		},
		WantErr: fmt.Errorf("judgeNodeAndTaskNPU node don't have enough resource, req<2>, idle<0>"),
	}
}

func buildCheckNodeNPUByTaskTestCase5() itest.CheckNodeNPUByTaskTestCase {
	return itest.CheckNodeNPUByTaskTestCase{
		Name: "05-CheckNodeNPUByTask return err when node has no enough npu",
		Task: test.FakeTaskWithResReq("pod0", util.NPU310PCardName, util.NPUIndex2),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0"},
			},
		},
		WantErr: errors.New("judgeNodeAndTaskNPU node don't have enough resource, req<2>, idle<1>"),
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
	npu := NPUHandler{}
	npu.SetAnnoName(util.NPU310PCardName)
	npu.SetAnnoPreVal(util.NPU310PCardNamePre)
	job := test.FakeNormalTestJob("job", 1)
	test.SetFakeJobResRequest(job, util.NPU310PCardName, "2")
	attr1 := itest.FakeSchedulerJobAttrByJob(job)
	npu.SetMaxNodeNPUNum(maxNodeNPUNum)
	npu.SetMaxCardNPUNum(maxCardNPUNum)
	npu.SetSchedulerAttr(attr1)
	env := plugin.ScheduleEnv{
		Jobs: map[api.JobID]plugin.SchedulerJob{test.FakeJobName: {SchedulerJobAttr: attr1}},
	}
	npu.SetSchedulerEnv(env)
	testCases := buildCheckNodeNPUByTaskTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.CheckNodeNPUByTask(tt.Task, tt.Node); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

// TestScoreBestNPUNodes
func TestScoreBestNPUNodes(t *testing.T) {
	npu1 := NPUHandler{}
	task := &api.TaskInfo{}
	var nodes []*api.NodeInfo
	scoreMap := make(map[string]float64, util.MapInitNum)
	wantErr := errors.New(util.ArgumentError)
	t.Run("scoreBestNPU will return err", func(t *testing.T) {
		if err := npu1.ScoreBestNPUNodes(task, nodes, scoreMap); !reflect.DeepEqual(err, wantErr) {
			t.Errorf("ScoreBestNPUNodes() error = %v, wantErr %v", err, nil)
		}
	})
}

func buildUseAnnotationTestCases01() []itest.UseAnnotationTestCase {
	return []itest.UseAnnotationTestCase{
		{
			Name: "01-UseAnnotation success when node resource meet task req",
			Task: test.FakeTaskWithResReq("pod0", util.NPU310PCardName, util.NPUIndex2),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
					Allocate:   map[v1.ResourceName]float64{util.NPU310PCardName: util.NPUIndex2 * util.NPUHexKilo},
				},
			},
			PodAnno: "Ascend310P-0,Ascend310P-1",
			WantNode: &plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Allocate:   map[v1.ResourceName]float64{util.NPU310PCardName: util.NPUIndex2 * util.NPUHexKilo},
					Annotation: map[string]string{util.NPU310PCardName: ""},
				},
			},
		},
		{
			Name: "02-UseAnnotation return err when task is nil",
			Task: nil,
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310PCardName: ""},
					Allocate:   map[v1.ResourceName]float64{util.NPU310PCardName: 0},
				},
			},
			PodAnno:  "",
			WantNode: nil,
		},
		{
			Name: "03-UseAnnotation return err when node annotation is nil",
			Task: new(api.TaskInfo),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: nil,
					Allocate:   map[v1.ResourceName]float64{util.NPU310PCardName: 0},
				},
			},
			PodAnno:  "",
			WantNode: nil,
		},
	}
}

func buildUseAnnotationTestCases02() []itest.UseAnnotationTestCase {
	return []itest.UseAnnotationTestCase{
		{
			Name: "04-UseAnnotation return err when node annotation resource less than task req npu",
			Task: test.FakeTaskWithResReq("pod0", util.NPU310PCardName, util.NPUIndex2),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{util.NPU310PCardName: ""},
					Allocate:   map[v1.ResourceName]float64{util.NPU310PCardName: 0},
				},
			},
			PodAnno:  "",
			WantNode: nil,
		},
	}
}

// TestUseAnnotation
func TestUseAnnotation(t *testing.T) {
	npu := NPUHandler{}
	npu.SetAnnoName(util.NPU310PCardName)
	npu.SetAnnoPreVal(util.NPU310PCardNamePre)
	npu.SetMaxNodeNPUNum(maxNodeNPUNum)
	job := test.FakeNormalTestJob("job", 1)
	test.SetFakeJobResRequest(job, util.NPU910CardName, "2")
	attr := itest.FakeSchedulerJobAttrByJob(job)
	env := plugin.ScheduleEnv{Jobs: map[api.JobID]plugin.SchedulerJob{test.FakeJobName: {SchedulerJobAttr: attr}}}
	if err := npu.InitMyJobPlugin(attr, env); err != nil {
		t.Errorf("InitMyJobPlugin failed err: %s", err.Error())
	}
	testCases := buildUseAnnotationTestCases01()
	testCases = append(testCases, buildUseAnnotationTestCases02()...)
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			node := npu.UseAnnotation(tt.Task, tt.Node)
			if tt.Task != nil && tt.Node.Annotation != nil && !reflect.DeepEqual(
				tt.Task.Pod.Annotations[util.NPU310PCardName], tt.PodAnno) {
				t.Errorf("UseAnnotation() anno %v, wantAnno %v", tt.Task.Pod.Annotations, tt.PodAnno)
			}
			if !reflect.DeepEqual(node, tt.WantNode) {
				t.Errorf("UseAnnotation() node: %v, wantNode: %v", node, tt.WantNode)
			}
		})
	}
}

func buildJudgeNodeAndTaskNPUTestCases() []itest.JudgeNodeAndTaskNPUTestCase {
	const npuNum65 = 65
	return []itest.JudgeNodeAndTaskNPUTestCase{
		{
			Name:    "01-JudgeNodeAndTaskNPU return err when task npu nun is 0",
			TaskNPU: 0,
			NodeTop: []int{0, 1},
			WantErr: errors.New("judgeNodeAndTaskNPU task req num<0> is invalid"),
		},
		{
			Name:    "02-JudgeNodeAndTaskNPU return err when task npu num is 65",
			TaskNPU: npuNum65,
			NodeTop: []int{0, 1},
			WantErr: errors.New("judgeNodeAndTaskNPU task req num<65> is invalid"),
		},
		{
			Name:    "03-JudgeNodeAndTaskNPU return err when node not meet task npu num",
			TaskNPU: util.NPUIndex2,
			NodeTop: []int{0},
			WantErr: errors.New("judgeNodeAndTaskNPU node don't have enough resource, req<2>, idle<1>"),
		},
	}
}

// TestJudgeNodeAndTaskNPU
func TestJudgeNodeAndTaskNPU(t *testing.T) {
	npu := NPUHandler{}
	npu.SetMaxNodeNPUNum(maxNodeNPUNum)
	testCases := buildJudgeNodeAndTaskNPUTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.JudgeNodeAndTaskNPU(tt.TaskNPU, tt.NodeTop); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("JudgeNodeAndTaskNPU() err: %s, wanterr: %s", err, tt.WantErr)
			}
		})
	}
}

func buildSetMaxNodeNPUNumTestCase() []itest.SetMaxNodeNPUNumTestCase {
	return []itest.SetMaxNodeNPUNumTestCase{
		{
			Name:    "01-SetMaxNodeNPUNum the get num not equal wantNum when num is invalid",
			Num:     -1,
			WantNum: 0,
		},
		{
			Name:    "02-SetMaxNodeNPUNum the get num equal wantNum when num is valid",
			Num:     1,
			WantNum: 1,
		},
	}
}

// TestSetMaxNodeNPUNum
func TestSetMaxNodeNPUNum(t *testing.T) {
	npu := NPUHandler{}
	testCases := buildSetMaxNodeNPUNumTestCase()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			npu.SetMaxNodeNPUNum(tt.Num)
			if npu.MaxNodeNPUNum != tt.WantNum {
				t.Errorf("SetMaxNodeNPUNum() num: %d, wanterr: %d", npu.MaxNodeNPUNum, tt.WantNum)
			}
		})
	}
}
