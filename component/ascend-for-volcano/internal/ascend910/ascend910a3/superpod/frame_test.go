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
Package superpod is using for HuaWei Atlas 900 A3 SuperPod affinity schedule.
*/
package superpod

import (
	"errors"
	"reflect"
	"strconv"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

type selectSuperPodForJobTestCase struct {
	name       string
	nodes      []*api.NodeInfo
	tasks      map[api.TaskID]util.NPUTask
	npuTaskNum int
	frameAttr  plugin.VolcanoFrame
	spBlock    int
	want       map[int32]struct{}
	wantErr    error
}

func newNPUNodeWithSuperPodID(nodeName string, superPodID int32) plugin.NPUNode {
	return plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name:       nodeName,
			SuperPodID: superPodID,
		},
	}
}

func newNPUNodes(n int, sp int) map[string]plugin.NPUNode {
	nodes := make(map[string]plugin.NPUNode)
	for i := 0; i < n; i++ {
		nodeName := "node" + strconv.Itoa(i)
		nodes[nodeName] = newNPUNodeWithSuperPodID(nodeName, int32(i/sp))
	}
	return nodes
}

func newNPUTasks(n int) map[api.TaskID]util.NPUTask {
	tasks := make(map[api.TaskID]util.NPUTask)
	for i := 0; i < n; i++ {
		tasks[api.TaskID(strconv.Itoa(i))] = util.NPUTask{Name: "task" + strconv.Itoa(i)}
	}
	return tasks
}

var (
	node0  = &api.NodeInfo{Name: "node0"}
	node1  = &api.NodeInfo{Name: "node1"}
	node2  = &api.NodeInfo{Name: "node2"}
	node3  = &api.NodeInfo{Name: "node3"}
	node4  = &api.NodeInfo{Name: "node4"}
	node10 = &api.NodeInfo{Name: "node10"}
	node11 = &api.NodeInfo{Name: "node11"}
	node12 = &api.NodeInfo{Name: "node12"}
	node13 = &api.NodeInfo{Name: "node13"}
	node14 = &api.NodeInfo{Name: "node14"}
	node20 = &api.NodeInfo{Name: "node20"}
	node21 = &api.NodeInfo{Name: "node21"}
	node22 = &api.NodeInfo{Name: "node22"}
	node23 = &api.NodeInfo{Name: "node23"}
	node24 = &api.NodeInfo{Name: "node24"}
	node25 = &api.NodeInfo{Name: "node25"}
	node26 = &api.NodeInfo{Name: "node26"}
	node27 = &api.NodeInfo{Name: "node27"}
)

const (
	superPodSize10  = 10
	reservePodSize2 = 2
	reservePodSize4 = 4
	spBlockNum1     = 1
	spBlockNum2     = 2
	npuTaskNum1     = 1
	npuTaskNum4     = 4
	superPodId2     = 2
)

var selectSuperPodForJobTestCases = []selectSuperPodForJobTestCase{
	{
		name:  "01-total nodes is not fit for job require, should return err",
		tasks: newNPUTasks(npuTaskNum4),
		nodes: []*api.NodeInfo{node0, node1},
		frameAttr: plugin.VolcanoFrame{
			SuperPodSize:   superPodSize10,
			ReservePodSize: reservePodSize2,
		},
		npuTaskNum: npuTaskNum4,
		spBlock:    spBlockNum2,
		want:       nil,
		wantErr:    errors.New("select super pod failed, required 2, total 1"),
	},

	{
		name:       "02-2*2 job will select 4 and 5 from super pods with node 4, 5 or 6",
		tasks:      newNPUTasks(npuTaskNum4),
		npuTaskNum: npuTaskNum4,
		nodes: []*api.NodeInfo{node0, node1, node2, node3, node10, node11, node12, node13, node14, node20,
			node21, node22, node23, node24, node25,
		},

		frameAttr: plugin.VolcanoFrame{
			SuperPodSize:   superPodSize10,
			ReservePodSize: reservePodSize2,
		},
		spBlock: spBlockNum2,
		want:    map[int32]struct{}{0: {}, 1: {}},
		wantErr: nil,
	},

	{
		name:  "03-2*2 job will select 6 from super pods with node 5 or 6",
		tasks: newNPUTasks(npuTaskNum4),
		nodes: []*api.NodeInfo{node10, node11, node12, node13, node14, node20,
			node21, node22, node23, node24, node25,
		},
		npuTaskNum: npuTaskNum4,
		frameAttr: plugin.VolcanoFrame{
			SuperPodSize:   superPodSize10,
			ReservePodSize: reservePodSize2,
		},
		spBlock: spBlockNum2,
		want:    map[int32]struct{}{superPodId2: {}},
		wantErr: nil,
	},

	{
		name:  "04-2*2 job will select 5 and 8 from super pods with node 5 or 8",
		tasks: newNPUTasks(npuTaskNum4),
		nodes: []*api.NodeInfo{node10, node11, node12, node13, node14, node20,
			node21, node22, node23, node24, node25, node26, node27,
		},
		npuTaskNum: npuTaskNum4,
		frameAttr: plugin.VolcanoFrame{
			SuperPodSize:   superPodSize10,
			ReservePodSize: reservePodSize2,
		},
		spBlock: spBlockNum2,
		want:    map[int32]struct{}{1: {}, superPodId2: {}},
		wantErr: nil,
	},
	{
		name:  "05-2*2 job will select 5 and 8 from super pods with node 3, 5 or 8",
		tasks: newNPUTasks(npuTaskNum4),
		nodes: []*api.NodeInfo{node0, node1, node2, node10, node11, node12, node13, node14, node20,
			node21, node22, node23, node24, node25, node26, node27,
		},
		npuTaskNum: npuTaskNum4,
		frameAttr: plugin.VolcanoFrame{
			SuperPodSize:   superPodSize10,
			ReservePodSize: reservePodSize2,
		},
		spBlock: spBlockNum2,
		want:    map[int32]struct{}{1: {}, superPodId2: {}},
		wantErr: nil,
	},
	{
		name:  "06-2*2 job will select 3 and 5 from super pods with node 3, 5",
		tasks: newNPUTasks(npuTaskNum4),
		nodes: []*api.NodeInfo{node0, node1, node2, node10, node11, node12, node13, node14},
		frameAttr: plugin.VolcanoFrame{
			SuperPodSize:   superPodSize10,
			ReservePodSize: reservePodSize2,
		},
		npuTaskNum: npuTaskNum4,
		spBlock:    spBlockNum2,
		want:       map[int32]struct{}{1: {}, 0: {}},
		wantErr:    nil,
	},

	{
		name:  "07-2*2 job will select 5 and 5 from super pods with node 5, 5",
		tasks: newNPUTasks(npuTaskNum4),
		nodes: []*api.NodeInfo{node0, node1, node2, node3, node4, node10, node11, node12, node13, node14},
		frameAttr: plugin.VolcanoFrame{
			SuperPodSize:   superPodSize10,
			ReservePodSize: reservePodSize2,
		},
		npuTaskNum: npuTaskNum4,
		spBlock:    spBlockNum2,
		want:       map[int32]struct{}{1: {}, 0: {}},
		wantErr:    nil,
	},
	{
		name:  "08-2*2 job will select 3 from super pods with node 1, 3",
		tasks: newNPUTasks(npuTaskNum1),
		nodes: []*api.NodeInfo{node0, node10, node11, node12},
		frameAttr: plugin.VolcanoFrame{
			SuperPodSize:   superPodSize10,
			ReservePodSize: reservePodSize4,
		},
		npuTaskNum: npuTaskNum1,
		spBlock:    spBlockNum1,
		want:       map[int32]struct{}{1: {}},
		wantErr:    nil,
	},
}

func TestSelectSuperPodForJob(t *testing.T) {
	plg, _ := New(SchedulerName).(*module910SuperPod)
	plg.Name = "job1"
	plg.SchedulerJobAttr = util.SchedulerJobAttr{
		ComJob: util.ComJob{},
		NPUJob: &util.NPUJob{},
	}
	plg.ScheduleEnv = plugin.ScheduleEnv{}
	task := &api.TaskInfo{
		UID:  "0",
		Job:  "job1",
		Name: "task0",
	}
	const npuNodes = 30
	plg.Nodes = newNPUNodes(npuNodes, superPodSize10)
	scoreMap := make(map[string]float64, 0)
	for _, cs := range selectSuperPodForJobTestCases {
		t.Run(cs.name, func(t *testing.T) {
			plg.spBlock = cs.spBlock
			plg.Tasks = cs.tasks
			plg.FrameAttr = cs.frameAttr
			plg.NPUTaskNum = cs.npuTaskNum
			selectedNodes, err := plg.selectSuperPodForJob(task, cs.nodes, scoreMap)
			if !reflect.DeepEqual(err, cs.wantErr) {
				t.Errorf("InitMyJobPlugin() error = %v, wantErr %v", err, cs.wantErr)
			}
			if !reflect.DeepEqual(getSelectedNodesSuperPodID(selectedNodes), cs.want) {
				t.Errorf("InitMyJobPlugin() selectedNodes = %v, want: %v", selectedNodes, cs.want)
			}
		})

	}
}

func getSelectedNodesSuperPodID(selectedNodes map[string][]plugin.SuperNode) map[int32]struct{} {
	if selectedNodes == nil {
		return nil
	}
	selectedNodesID := make(map[int32]struct{}, 0)
	for _, sp := range selectedNodes {
		for _, node := range sp {
			selectedNodesID[node.SuperPodID] = struct{}{}
		}
	}
	return selectedNodesID
}
