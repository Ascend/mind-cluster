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
Package ascend310 is using for HuaWei A800/9000 Ascend910 pin affinity schedule.
*/
package ascend310

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/ascend310/chip310x4"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

const (
	MockTaskNumTwo   = 2
	MockJobNamespace = "vcjob"
	MockPodOneName   = "pod1"
	MockNodeOneName  = "node1"
)

type nilTPTestCase struct {
	name               string
	tp                 *asend310
	wantNode           *plugin.NPUNode
	wantArr            []int
	wantValidateResult *api.ValidateResult
	wantErr            error
}

// TestNew
func TestNew(t *testing.T) {
	t.Run("test New", func(t *testing.T) {
		npu := New(PluginName)
		if npu.GetPluginName() != PluginName {
			t.Errorf("New() npu Name: %s, wantName: %s.", npu.GetPluginName(), PluginName)
		}
		if npu.GetAnnoName() != util.NPU310CardName {
			t.Errorf("New() npu annoName: %s, wantAnnoName: %s.", npu.GetPluginName(), util.NPU310CardName)
		}
		if npu.GetAnnoPreVal() != util.NPU310CardNamePre {
			t.Errorf("New() npu annoNamePre: %s, wantAnnoNamePre: %s.",
				npu.GetPluginName(), util.NPU310CardNamePre)
		}
	})
}

type initMyJobPluginTestCase struct {
	name    string
	attr    util.SchedulerJobAttr
	env     plugin.ScheduleEnv
	handler base.AscendHandler
	wantErr error
}

func initNode(nodeName string, annoKey string, annoVal string) plugin.NPUNode {
	return plugin.NPUNode{CommonNode: plugin.CommonNode{
		Name:       nodeName,
		Annotation: map[string]string{annoKey: annoVal},
	}}
}

func buildNilTPTestCase(funcName string) nilTPTestCase {
	return nilTPTestCase{
		name: "00-" + funcName + " will return when tp is nil",
		wantValidateResult: &api.ValidateResult{
			Pass:    false,
			Reason:  "invalid argument",
			Message: "invalid argument"},
		wantErr: fmt.Errorf("nil plugin %s", PluginName),
	}
}

func buildInitMyJobPluginTestCases() []initMyJobPluginTestCase {
	return []initMyJobPluginTestCase{
		{
			name: "01-InitMyJobPlugin return nil when define accelerator the handler will be define as card",
			attr: util.SchedulerJobAttr{
				ComJob: util.ComJob{Label: map[string]string{Accelerator310Key: Card310AcceleratorValue}},
				NPUJob: &util.NPUJob{ReqNPUName: util.NPU310CardName},
			},
			env:     plugin.ScheduleEnv{},
			wantErr: nil,
		},
		{
			name: "02-InitMyJobPlugin return nil when not define accelerator the handler will be define as chip",
			attr: util.SchedulerJobAttr{
				ComJob: util.ComJob{Label: map[string]string{}},
				NPUJob: &util.NPUJob{ReqNPUName: util.NPU310CardName},
			},
			env:     plugin.ScheduleEnv{},
			handler: chip310x4.New(chip310x4.SchedulerName),
			wantErr: nil,
		},
		{
			name: "03-InitMyJobPlugin return error when not define accelerator the handler will be define as chip",
			attr: util.SchedulerJobAttr{
				ComJob: util.ComJob{Label: map[string]string{}},
				NPUJob: &util.NPUJob{ReqNPUName: util.NPU310PCardName},
			},
			env:     plugin.ScheduleEnv{},
			handler: chip310x4.New(chip310x4.SchedulerName),
			wantErr: fmt.Errorf("not support %s", util.NPU310PCardName+Chip310AcceleratorValue),
		},
	}
}

func TestInitMyJobPlugin(t *testing.T) {
	testCase := buildNilTPTestCase("InitMyJobPlugin")
	t.Run(testCase.name, func(t *testing.T) {
		if err := testCase.tp.InitMyJobPlugin(util.SchedulerJobAttr{},
			plugin.ScheduleEnv{}); !reflect.DeepEqual(err, testCase.wantErr) {
			t.Errorf("InitMyJobPlugin() = %v, want %v", err, testCase.wantErr)
		}
	})
	npu := New(PluginName)
	testCases := buildInitMyJobPluginTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := npu.InitMyJobPlugin(tt.attr, tt.env)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("ValidNPUJob() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// checkNodeNPUByTaskTestCase CheckNodeNPUByTask test case
type checkNodeNPUByTaskTestCase struct {
	Task    *api.TaskInfo
	Name    string
	Attr    util.SchedulerJobAttr
	Node    plugin.NPUNode
	WantErr error
}

func buildCheckNodeNPUByTaskTestCases() []checkNodeNPUByTaskTestCase {
	return []checkNodeNPUByTaskTestCase{
		{
			Name:    "01-CheckNodeNPUByTask return err when task is nil",
			Task:    nil,
			Node:    initNode(MockNodeOneName, util.NPU310CardName, "Ascend310-0,Ascend310-1"),
			WantErr: errors.New(util.ArgumentError),
		},
		{
			Name: "02-CheckNodeNPUByTask return err when node annotation is nil",
			Task: test.FakeTaskWithResReq(MockPodOneName, util.NPU310CardName, MockTaskNumTwo),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       MockNodeOneName,
					Annotation: nil,
				}},
			WantErr: errors.New(util.ArgumentError),
		},
	}
}

// TestCheckNodeNPUByTask
func TestCheckNodeNPUByTask(t *testing.T) {
	npu := New(PluginName)
	testCases := buildCheckNodeNPUByTaskTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.CheckNodeNPUByTask(tt.Task, tt.Node); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

// scoreBestNPUNodesTestCase scoreBestNPUNodes test case
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
	return []scoreBestNPUNodesTestCase{
		{
			Name:     "01-ScoreBestNPUNodes return err when task is nil",
			Task:     nil,
			Nodes:    []*api.NodeInfo{test.FakeNormalTestNode(MockNodeOneName)},
			ScoreMap: map[string]float64{MockNodeOneName: 0},
			WantSMap: map[string]float64{MockNodeOneName: 0},
			WantErr:  errors.New(util.ArgumentError),
		},
		{
			Name:     "02-ScoreBestNPUNodes return err when nodes is empty",
			Task:     test.FakeNormalTestTask(MockPodOneName, MockNodeOneName, MockJobNamespace),
			Nodes:    []*api.NodeInfo{},
			ScoreMap: map[string]float64{MockNodeOneName: 0},
			WantSMap: map[string]float64{MockNodeOneName: 0},
			WantErr:  errors.New(util.ArgumentError),
		},
		{
			Name:     "03-ScoreBestNPUNodes return err when scoreMap is empty",
			Task:     test.FakeNormalTestTask(MockPodOneName, MockNodeOneName, MockJobNamespace),
			Nodes:    []*api.NodeInfo{test.FakeNormalTestNode(MockNodeOneName)},
			ScoreMap: map[string]float64{},
			WantSMap: map[string]float64{},
			WantErr:  errors.New(util.ArgumentError),
		},
	}
}

// TestCheckNodeNPUByTask
func TestScoreBestNPUNodes(t *testing.T) {
	npu := New(PluginName)
	testCases := buildScoreBestNPUNodesTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.ScoreBestNPUNodes(tt.Task, tt.Nodes, tt.ScoreMap); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

// useAnnotationTestCase useAnnotation test case
type useAnnotationTestCase struct {
	Task     *api.TaskInfo
	WantNode *plugin.NPUNode
	Name     string
	Node     plugin.NPUNode
	PodAnno  string
	Attr     util.SchedulerJobAttr
}

func buildUseAnnotationTestCases() []useAnnotationTestCase {
	return []useAnnotationTestCase{
		{
			Name:     "01-ScoreBestNPUNodes return nil when task is nil",
			Task:     nil,
			Node:     initNode(MockNodeOneName, util.NPU310CardName, "Ascend310-0,Ascend310-1"),
			WantNode: nil,
		},
		{
			Name:     "02-ScoreBestNPUNodes return nil when node annotation is nil",
			Task:     test.FakeNormalTestTask(MockPodOneName, MockNodeOneName, MockJobNamespace),
			Node:     plugin.NPUNode{CommonNode: plugin.CommonNode{}},
			WantNode: nil,
		},
	}
}

// TestUseAnnotation
func TestUseAnnotation(t *testing.T) {
	npu := New(PluginName)
	testCases := buildUseAnnotationTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := npu.UseAnnotation(tt.Task, tt.Node); !reflect.DeepEqual(err, tt.WantNode) {
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.WantNode)
			}
		})
	}
}

type ValidNPUJobTest struct {
	name string
	tp   *asend310
	want *api.ValidateResult
}

func buildValidNPUJobTestCase() []ValidNPUJobTest {
	tests := []ValidNPUJobTest{
		{
			name: "01-ValidNPUJob will return when tp is nil",
			tp:   nil,
			want: &api.ValidateResult{Pass: false, Reason: "invalid argument", Message: "invalid argument"},
		},
		{
			name: "02-ValidNPUJob will return when tp is nil",
			tp:   &asend310{},
			want: nil,
		},
	}
	return tests
}

func TestAscend310ValidNPUJob(t *testing.T) {
	tests := buildValidNPUJobTestCase()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tp.ValidNPUJob(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidNPUJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAscend310Name(t *testing.T) {
	tests := []struct {
		Name string
		want string
	}{
		{
			Name: "01-Name will return PluginName",
			want: PluginName,
		},
	}
	tp := &asend310{}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if got := tp.Name(); got != tt.want {
				t.Errorf("Name() = %s, want %s", got, tt.want)
			}
		})
	}
}

// releaseAnnotationTestCase releaseAnnotation test case
type releaseAnnotationTestCase struct {
	Task     *api.TaskInfo
	WantNode *plugin.NPUNode
	Name     string
	Node     plugin.NPUNode
	PodAnno  string
	Attr     util.SchedulerJobAttr
}

func buildReleaseAnnotationTestCases() []releaseAnnotationTestCase {
	node := initNode(MockNodeOneName, util.NPU310CardName, "Ascend310-0,Ascend310-1")
	return []releaseAnnotationTestCase{{
		Name:     "01-ReleaseAnnotation return nil when call this fn",
		Node:     node,
		WantNode: &node,
	}}
}

func TestReleaseAnnotation(t *testing.T) {
	tests := buildReleaseAnnotationTestCases()
	tp := &asend310{}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if got := tp.ReleaseAnnotation(tt.Task, tt.Node); !reflect.DeepEqual(got, tt.WantNode) {
				t.Errorf("ReleaseAnnotation() = %v, want %v", got, tt.WantNode)
			}
		})
	}
}
