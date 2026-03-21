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

// Package chip1softsharedev is using for HuaWei chip1softsharedev schedule.
package chip1softsharedev

import (
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910b"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

const (
	pluginName           = "chip1softsharedev"
	testNodeName         = "test-node"
	testTaskName         = "test-task"
	npuChipMemoryValue   = "32G"
	aiCoreQuotaValue     = "50"
	hbmQuotaValue        = "16384"
	maxNodeNPUNum        = 16
	softShareEnableValue = "true"
)

func TestCheckNodeNPUByTask(t *testing.T) {
	tp := &chip1softsharedev{Base910b: ascend910b.Base910b{
		NPUHandler: base.NPUHandler{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{}}}}}
	task := &api.TaskInfo{Name: testTaskName}
	node := plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name:       testNodeName,
			Label:      map[string]string{},
			Tasks:      map[api.TaskID]*api.TaskInfo{},
			Annotation: map[string]string{util.SchedulePolicyAnnoKey: util.Chip1ShareShareDev},
		},
	}
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&base.NPUHandler{}), "GetUsableTopFromNode",
		func(_ *base.NPUHandler, node plugin.NPUNode, disFlag bool) ([]int, error) { return []int{0}, nil }).
		ApplyFunc((*chip1softsharedev).checkNodeUsableResourceForTask, func(_ *chip1softsharedev, _ plugin.NPUNode,
			_ []int, _ int, _ softShareDevResource, _ *api.TaskInfo) bool {
			return true
		})
	defer patch.Reset()
	err := tp.CheckNodeNPUByTask(task, node)
	if err == nil {
		t.Error("CheckNodeNPUByTask() should return error for missing soft share label")
	}
	node.Label[util.SchedulerSoftShareDevEnableNodeLabel] = softShareEnableValue
	node.Label[util.NPUChipMemoryLabelKey] = npuChipMemoryValue
	tp.Label = map[string]string{
		util.SchedulerSoftShareDevAicoreQuotaKey: aiCoreQuotaValue,
		util.SchedulerSoftShareDevHbmQuotaKey:    hbmQuotaValue,
		util.SchedulerSoftShareDevPolicyKey:      util.SoftShareDevPolicyFixedShare,
	}
	tp.SetMaxNodeNPUNum(maxNodeNPUNum)
	err = tp.CheckNodeNPUByTask(task, node)
	if err != nil {
		t.Errorf("CheckNodeNPUByTask() unexpected error: %v", err)
	}
}

func TestScoreBestNPUNodes(t *testing.T) {
	tp, ok := New(pluginName).(*chip1softsharedev)
	if !ok {
		t.Error("New() should return chip1softsharedev plugin")
	}
	task := &api.TaskInfo{Name: testTaskName}
	node := &api.NodeInfo{
		Name: testNodeName,
		Node: &v1.Node{ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{util.NPUChipMemoryLabelKey: npuChipMemoryValue},
		}},
	}
	nodes := []*api.NodeInfo{node}
	sMap := map[string]float64{testNodeName: 0}
	tp.Label = map[string]string{
		util.SchedulerSoftShareDevAicoreQuotaKey: aiCoreQuotaValue,
		util.SchedulerSoftShareDevHbmQuotaKey:    hbmQuotaValue,
		util.SchedulerSoftShareDevPolicyKey:      util.SoftShareDevPolicyFixedShare,
	}
	tp.Nodes = map[string]plugin.NPUNode{
		testNodeName: {CommonNode: plugin.CommonNode{
			Name:       testNodeName,
			Label:      node.Node.Labels,
			Tasks:      map[api.TaskID]*api.TaskInfo{},
			Annotation: map[string]string{util.SchedulerSoftShareDevEnableNodeLabel: softShareEnableValue},
		}},
	}
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&base.NPUHandler{}), "GetUsableTopFromNode",
		func(_ *base.NPUHandler, node plugin.NPUNode, disFlag bool) ([]int, error) { return []int{0}, nil })
	defer patch.Reset()
	err := tp.ScoreBestNPUNodes(task, nodes, sMap)
	if err != nil {
		t.Errorf("ScoreBestNPUNodes() error = %v", err)
	}
	if len(sMap) == 0 {
		t.Error("ScoreBestNPUNodes() should populate score map")
	}
}

func TestNew(t *testing.T) {
	plugin := New(pluginName)
	if plugin == nil {
		t.Error("New() should return non-nil plugin")
	}
	if plugin.GetPluginName() != pluginName {
		t.Errorf("New() plugin name = %v, want %s", plugin.GetPluginName(), pluginName)
	}
}
