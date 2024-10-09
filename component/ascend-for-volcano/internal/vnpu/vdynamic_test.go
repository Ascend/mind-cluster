/*
Copyright(C)2024. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package vnpu is using for HuaWei Ascend pin vnpu allocation.
*/
package vnpu

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

type checkNodeNPUByDyTaskFields struct {
	StaticByConf bool
	VT           VTemplate
	StaticVNPU   StaticVNPU
	DynamicVNPU  DynamicVNPU
}

type checkNodeNPUByDyTaskArgs struct {
	task       *api.TaskInfo
	node       plugin.NPUNode
	taskResReq util.VResource
}

type checkNodeNPUByDyTaskTestCase struct {
	name    string
	fields  checkNodeNPUByDyTaskFields
	args    checkNodeNPUByDyTaskArgs
	wantErr bool
}

func buildCheckNodeNPUByDyTaskTestCase01() checkNodeNPUByDyTaskTestCase {
	return checkNodeNPUByDyTaskTestCase{
		name:    "01 will return error when tp is nil or task is nil",
		fields:  checkNodeNPUByDyTaskFields{},
		args:    checkNodeNPUByDyTaskArgs{},
		wantErr: true,
	}
}

func buildCheckNodeNPUByDyTaskTestCase02() checkNodeNPUByDyTaskTestCase {
	node := plugin.NPUNode{}
	node.ValidVNode = false
	return checkNodeNPUByDyTaskTestCase{
		name:    "02 will return error when node is not vNode",
		fields:  checkNodeNPUByDyTaskFields{StaticByConf: false},
		args:    checkNodeNPUByDyTaskArgs{task: &api.TaskInfo{}, node: node},
		wantErr: true,
	}
}

func buildCheckNodeNPUByDyTaskTestCase03() checkNodeNPUByDyTaskTestCase {
	arg := checkNodeNPUByDyTaskArgs{}
	arg.taskResReq = util.VResource{Aicore: 6, Aicpu: 4}
	arg.node.Chips = make(map[int]*plugin.VChip)
	arg.node.ValidVNode = true
	arg.node.Chips[0] = &plugin.VChip{TotalRes: util.VResource{Aicore: 5}}
	return checkNodeNPUByDyTaskTestCase{
		name:   "03 will return error when node res is not meet job require",
		fields: checkNodeNPUByDyTaskFields{StaticByConf: false},
		args: checkNodeNPUByDyTaskArgs{task: &api.TaskInfo{}, node: arg.node,
			taskResReq: arg.taskResReq},
		wantErr: true,
	}
}

func buildCheckNodeNPUByDyTaskTestCase04() checkNodeNPUByDyTaskTestCase {
	tp := checkNodeNPUByDyTaskFields{}
	tp.StaticByConf = false
	tp.DynamicVNPU.DowngradeCache = make(map[string][]string)
	arg := checkNodeNPUByDyTaskArgs{}
	arg.taskResReq = util.VResource{Aicore: 2, Aicpu: 2}
	arg.node.Chips = make(map[int]*plugin.VChip)
	arg.node.ValidVNode = true
	arg.node.Chips[0] = &plugin.VChip{TotalRes: util.VResource{Aicore: 1}}
	return checkNodeNPUByDyTaskTestCase{
		name:   "04 will return error when task can be down grade and node res is not meet job require",
		fields: tp,
		args: checkNodeNPUByDyTaskArgs{task: &api.TaskInfo{}, node: arg.node,
			taskResReq: arg.taskResReq},
		wantErr: true,
	}
}

func buildCheckNodeNPUByDyTaskTestCase05() checkNodeNPUByDyTaskTestCase {
	tp := checkNodeNPUByDyTaskFields{}
	tp.StaticByConf = false
	tp.DynamicVNPU.DowngradeCache = make(map[string][]string)
	arg := checkNodeNPUByDyTaskArgs{}
	arg.taskResReq = util.VResource{Aicore: 2, Aicpu: 2}
	arg.node.Chips = make(map[int]*plugin.VChip)
	arg.node.ValidVNode = true
	tmpRes := util.VResource{Aicore: 2, Aicpu: 2}
	arg.node.Chips[0] = &plugin.VChip{FreeRes: tmpRes, TotalRes: tmpRes}
	return checkNodeNPUByDyTaskTestCase{
		name:   "05 will return nil when  node res is  meet job require",
		fields: tp,
		args: checkNodeNPUByDyTaskArgs{task: &api.TaskInfo{}, node: arg.node,
			taskResReq: arg.taskResReq},
		wantErr: false,
	}
}

func buildCheckNodeNPUByDyTaskTestCase() []checkNodeNPUByDyTaskTestCase {
	return []checkNodeNPUByDyTaskTestCase{
		buildCheckNodeNPUByDyTaskTestCase01(),
		buildCheckNodeNPUByDyTaskTestCase02(),
		buildCheckNodeNPUByDyTaskTestCase03(),
		buildCheckNodeNPUByDyTaskTestCase04(),
		buildCheckNodeNPUByDyTaskTestCase05(),
	}
}

func TestVirtualNPUCheckNodeNPUByDyTask(t *testing.T) {
	tests := buildCheckNodeNPUByDyTaskTestCase()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &VirtualNPU{
				StaticByConf: tt.fields.StaticByConf,
				VT:           tt.fields.VT,
				StaticVNPU:   tt.fields.StaticVNPU,
				DynamicVNPU:  tt.fields.DynamicVNPU,
			}
			if err := tp.CheckNodeNPUByDyTask(tt.args.task, tt.args.node, tt.args.taskResReq); (err != nil) != tt.wantErr {
				t.Errorf("CheckNodeNPUByDyTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type scoreBestNPUNodesFields struct {
	vnpuHandler    vnpuHandler
	DowngradeCache map[string][]string
	ConCache       map[string]map[string]map[api.TaskID]struct{}
}

type scoreBestNPUNodesArgs struct {
	task     *api.TaskInfo
	nodes    []*api.NodeInfo
	scoreMap map[string]float64
}

type scoreBestNPUNodesTest struct {
	name    string
	fields  scoreBestNPUNodesFields
	args    scoreBestNPUNodesArgs
	wantErr bool
}

func buildScoreBestNPUNodesTestCase01() scoreBestNPUNodesTest {
	return scoreBestNPUNodesTest{
		name:    "01 will return err when tp is nil",
		fields:  scoreBestNPUNodesFields{},
		args:    scoreBestNPUNodesArgs{},
		wantErr: true,
	}
}

func buildScoreBestNPUNodesTestCase02() scoreBestNPUNodesTest {
	return scoreBestNPUNodesTest{
		name:    "02 will return err when scoreMap is nil",
		fields:  scoreBestNPUNodesFields{DowngradeCache: map[string][]string{}},
		args:    scoreBestNPUNodesArgs{task: &api.TaskInfo{}, nodes: []*api.NodeInfo{{}}},
		wantErr: true,
	}
}

func buildScoreBestNPUNodesTestCase03() scoreBestNPUNodesTest {
	return scoreBestNPUNodesTest{
		name:   "03 will return nil when node is not in down grade map ",
		fields: scoreBestNPUNodesFields{DowngradeCache: map[string][]string{}},
		args: scoreBestNPUNodesArgs{task: &api.TaskInfo{}, nodes: []*api.NodeInfo{{}},
			scoreMap: map[string]float64{"node1": 200}},
		wantErr: false,
	}
}

func buildScoreBestNPUNodesTestCase04() scoreBestNPUNodesTest {
	return scoreBestNPUNodesTest{
		name:   "04 will return nil when node is in down grade map ",
		fields: scoreBestNPUNodesFields{DowngradeCache: map[string][]string{"task01": {"node1"}}},
		args: scoreBestNPUNodesArgs{task: &api.TaskInfo{Name: "task01"}, nodes: []*api.NodeInfo{{Name: "node1"}},
			scoreMap: map[string]float64{"node1": 200}},
		wantErr: false,
	}
}

func buildScoreBestNPUNodesTestCase() []scoreBestNPUNodesTest {
	return []scoreBestNPUNodesTest{
		buildScoreBestNPUNodesTestCase01(),
		buildScoreBestNPUNodesTestCase02(),
		buildScoreBestNPUNodesTestCase03(),
		buildScoreBestNPUNodesTestCase04(),
	}
}

func TestDynamicVNPUScoreBestNPUNodes(t *testing.T) {
	tests := buildScoreBestNPUNodesTestCase()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &DynamicVNPU{
				vnpuHandler:    tt.fields.vnpuHandler,
				DowngradeCache: tt.fields.DowngradeCache,
				ConCache:       tt.fields.ConCache,
			}
			if err := tp.ScoreBestNPUNodes(tt.args.task, tt.args.nodes, tt.args.scoreMap); (err != nil) != tt.wantErr {
				t.Errorf("ScoreBestNPUNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
