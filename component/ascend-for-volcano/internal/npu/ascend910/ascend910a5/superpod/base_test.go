/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package superpod for base test function
package superpod

import (
	"fmt"
	"sort"
	"strconv"
	"testing"

	"k8s.io/api/core/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

const (
	testSuperPodID0 = 0
	testSuperPodID1 = 1
	testSuperPodID2 = 2
	testSpBlock4    = 4
	testSpBlock8    = 8
	testTpBlock4    = 4
	testTpBlock8    = 8
	testTotalCount1 = 1
	testNodeCount   = 12
	superPodSize32  = 32
)

const (
	spBlock1 = 1
	spBlock4 = 4
	spBlock6 = 6
	spBlock8 = 8

	npuTaskNum0  = 0
	npuTaskNum6  = 6
	npuTaskNum16 = 16
	npuTaskNum32 = 32
	npuTaskNum64 = 64

	superPodSize64 = 64

	tpBlock1 = 1
	tpBlock2 = 2
	tpBlock8 = 8
)

// for test cases use
const (
	nodeInfoIdx0  = 0
	nodeInfoIdx1  = 1
	nodeInfoIdx2  = 2
	nodeInfoIdx3  = 3
	nodeInfoIdx7  = 7
	nodeInfoIdx8  = 8
	nodeInfoIdx10 = 10
	nodeInfoIdx11 = 11
	nodeInfoIdx16 = 16
	nodeInfoIdx19 = 19
	nodeInfoIdx24 = 24
	nodeInfoIdx27 = 27

	nodeInfoIdx64  = 64
	nodeInfoIdx79  = 79
	nodeInfoIdx80  = 80
	nodeInfoIdx128 = 128
	nodeInfoIdx151 = 151
	nodeInfoIdx192 = 192
	nodeInfoIdx223 = 223
	nodeInfoIdx256 = 256
	nodeInfoIdx295 = 295
	nodeInfoIdx368 = 368
	nodeInfoIdx320 = 320
	nodeInfoIdx367 = 367
)

type doSelectForStrategyTestCaseType struct {
	name            string
	nodesInSuperPod superPod
	jobParams       jobParams
	selectRes       map[int32][]nodeBaseInfo
}

func buildDoSelectForStrategyTestCases() []doSelectForStrategyTestCaseType {
	return []doSelectForStrategyTestCaseType{
		{
			name: "case1 test need 4 nodes with tp-block=32, sp-block=32",
			jobParams: jobParams{
				spBlock:    testSpBlock4,
				tpBlock:    testTpBlock4,
				totalCount: testTotalCount1,
			},
			nodesInSuperPod: buildSuperPodsByParams(map[int32]int32{testSuperPodID0: testNodeCount})[testSuperPodID0],
		},
		{
			name: "case2 test need 8 nodes with tp-block=64, sp-block=64",
			jobParams: jobParams{
				spBlock:    testSpBlock8,
				tpBlock:    testTpBlock8,
				totalCount: testTotalCount1,
			},
			nodesInSuperPod: buildSuperPodsByParams(map[int32]int32{testSuperPodID0: testNodeCount})[testSuperPodID0],
		},
	}
}

// TestDoSelectForStrategy test doSelect can select nodes from given conditions
func TestDoSelectForStrategy(t *testing.T) {
	for _, testCase := range buildDoSelectForStrategyTestCases() {
		t.Run(testCase.name, func(t *testing.T) {
			scheduler := New(util.SuperPodx8SchedulerName)
			scheduler.jobParams = testCase.jobParams
			selectNodes := make(map[string][]plugin.SuperNode)
			unReadyIds := make([]string, 0)
			for i := 0; i < testCase.jobParams.totalCount; i++ {
				unReadyIds = append(unReadyIds, fmt.Sprintf("%d", i))
			}
			s := strategy{
				module910a5SuperPod: scheduler,
				selectedNodes:       selectNodes,
				unReadyIds:          unReadyIds,
				inNextStrategy:      false,
			}
			superPodWithRackId := transferSuperPodToRackIdMap(testCase.nodesInSuperPod)

			s.doSelect(superPodWithRackId, testCase.nodesInSuperPod)

			if len(selectNodes["0"]) != testCase.jobParams.spBlock {
				t.Errorf("doSelect result not match, expect get nodes length =%d, got selecteNodes: %v",
					testCase.jobParams.spBlock, selectNodes)
			}
		})
	}

}

type selectScoreBestNPUNodesTestCase struct {
	name             string
	npuTaskNum       int
	nodes            []*api.NodeInfo
	npuNodes         map[string]plugin.NPUNode
	superPodMap      map[int32]superPod
	tasks            map[api.TaskID]util.NPUTask
	scheduleStrategy int
	superPodSize     int
	spBlock          int
	tpBlock          int
	isNeedAlgoAlign  bool
	wantRes          map[int32]*selectedRackInfo
	wantErr          error
}

type selectedRackInfo struct {
	// how many racks have been selected, like: 1, 2, 3...
	selectedRacksInSuperPod int
	// how many nodes have been selected in every racks based on selectedRacksInSuperPod
	selectedNodesInRack []int
}

// total 168 nodes
// default nodes:
// 1、every rack with 8 nodes
// superPod0 - 1 rack
// superPod1 - 2 rack
// superPod2 - 3 rack
// superPod3 - 4 rack
// superPod4 - 5 rack
// superPod5 - 6 rack
func buildDefaultNodeInfoList() []*api.NodeInfo {
	var nodeInfo []*api.NodeInfo
	nodeInfo = append(nodeInfo, buildNodeInfos(nodeInfoIdx0, nodeInfoIdx7, nil)...)
	nodeInfo = append(nodeInfo, buildNodeInfos(nodeInfoIdx64, nodeInfoIdx79, nil)...)
	nodeInfo = append(nodeInfo, buildNodeInfos(nodeInfoIdx128, nodeInfoIdx151, nil)...)
	nodeInfo = append(nodeInfo, buildNodeInfos(nodeInfoIdx192, nodeInfoIdx223, nil)...)
	nodeInfo = append(nodeInfo, buildNodeInfos(nodeInfoIdx256, nodeInfoIdx295, nil)...)
	nodeInfo = append(nodeInfo, buildNodeInfos(nodeInfoIdx320, nodeInfoIdx367, nil)...)
	return nodeInfo
}

func buildNodeInfos(start int, end int, nodeInfo []*api.NodeInfo, exclude ...int) []*api.NodeInfo {
	if start > end {
		temp := start
		start = end
		end = temp
	}
	// exclude some nodes
	set := make(map[int]struct{})
	if len(exclude) > 0 {
		for _, i := range exclude {
			set[i] = struct{}{}
		}
	}
	for i := start; i <= end; i++ {
		_, ok := set[i]
		if ok {
			continue
		}
		nodeName := "node" + strconv.Itoa(i)
		nodeInfo = append(nodeInfo, &api.NodeInfo{Name: nodeName, Idle: &api.Resource{
			ScalarResources: map[v1.ResourceName]float64{
				util.NPU910CardName: 8,
			},
		}})
	}

	return nodeInfo
}

func newNPUTasks(n int, reqNpuNum int) map[api.TaskID]util.NPUTask {
	tasks := make(map[api.TaskID]util.NPUTask)
	for i := 0; i < n; i++ {
		tasks[api.TaskID(strconv.Itoa(i))] = util.NPUTask{Name: "task" + strconv.Itoa(i), ReqNPUNum: reqNpuNum}
	}
	return tasks
}

func setSuperPodSizeFrame(superpodSize int) plugin.VolcanoFrame {
	return plugin.VolcanoFrame{
		ConfigParameters: plugin.ConfigParameters{
			DynamicParameters: plugin.DynamicParameters{
				SuperPodSize: superpodSize,
			},
		},
	}
}

func getNPUNodes(start int, end int, sp int, rack int) map[string]plugin.NPUNode {
	nodes := make(map[string]plugin.NPUNode)
	for i := start; i < end; i++ {
		nodeName := "node" + strconv.Itoa(i)
		nodes[nodeName] = newNPUNodeWithNPUNum(nodeName, int32(i/sp), int32(i/rack), npuList8)
	}
	return nodes
}

func newNPUNodeWithNPUNum(nodeName string, superPodID int32, rackID int32,
	npuList []int) plugin.NPUNode {
	return plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name:       nodeName,
			SuperPodID: superPodID,
			RackID:     rackID,
			Annotation: map[string]string{
				"huawei.com/Ascend910": newNodeAnnotation(npuList),
				networkUnhealthyNPU:    "",
				faultNPU:               "",
			},
			Label: map[string]string{
				util.AcceleratorType: SuperPodx8,
			},
		},
	}
}

func newNodeAnnotation(npuList []int) string {
	var str string
	for index, npu := range npuList {
		if index == 0 {
			str = fmt.Sprintf("%s%d", util.NPU910CardNamePre, npu)
		} else {
			str = fmt.Sprintf("%s,%s%d", str, util.NPU910CardNamePre, npu)
		}
	}
	return str
}

func getTaskInfos(taskNum int, jobId string) []*api.TaskInfo {
	tasks := make([]*api.TaskInfo, taskNum)
	for i := 0; i < taskNum; i++ {
		indexStr := strconv.Itoa(i)
		task := &api.TaskInfo{
			UID:  api.TaskID(indexStr),
			Job:  api.JobID(jobId),
			Name: "task" + strconv.Itoa(i+1),
			Pod:  test.BuildNPUPod(test.NPUPod{}),
		}
		task.Pod.Annotations[plugin.PodRankIndexKey] = indexStr
		tasks[i] = task
	}
	return tasks
}

func superPodModelForTest(tasks []*api.TaskInfo, cs *selectScoreBestNPUNodesTestCase) *module910a5SuperPod {
	plg := New(SuperPodx8SchedulerName)
	jobs := make(map[api.JobID]plugin.SchedulerJob)
	job := plugin.SchedulerJob{
		JobReadyTag: new(bool),
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{
				Annotation: map[string]string{},
			},
			NPUJob: &util.NPUJob{},
		},
		SuperPods: map[string][]plugin.SuperNode{},
	}
	*job.JobReadyTag = true
	jobs[tasks[0].Job] = job
	plg.SchedulerJobAttr = util.SchedulerJobAttr{
		ComJob: util.ComJob{},
		NPUJob: &util.NPUJob{},
	}
	plg.ScheduleEnv = plugin.ScheduleEnv{}
	plg.spBlock = cs.spBlock
	plg.tpBlock = cs.tpBlock
	if len(cs.tasks) > 0 {
		plg.Tasks = cs.tasks
	} else {
		plg.Tasks = newNPUTasks(cs.npuTaskNum, nodeNPUNum)
	}
	plg.FrameAttr = setSuperPodSizeFrame(cs.superPodSize)
	plg.NPUTaskNum = cs.npuTaskNum
	plg.scheduleStrategy = cs.scheduleStrategy
	plg.Jobs = jobs
	if len(cs.npuNodes) > 0 {
		plg.Nodes = cs.npuNodes
	} else {
		plg.Nodes = getNPUNodes(nodeInfoIdx0, nodeInfoIdx368, cs.superPodSize, rackNodeNum)
	}
	plg.isNeedAlgoAlign = cs.isNeedAlgoAlign

	for _, task := range plg.Tasks {
		job.ReqNPUNum += task.ReqNPUNum
	}
	plg.ReqNPUNum = job.ReqNPUNum
	plg.ClusterCache.Jobs[job.Name] = job
	return plg
}

func checkScoreBestNPUNodesResult(selectedNodes map[string][]plugin.SuperNode) map[int32]*selectedRackInfo {
	if selectedNodes == nil {
		return make(map[int32]*selectedRackInfo)
	}
	selectedSuperPodInfo := make(map[int32]*selectedRackInfo)
	selectedRacksInSuperPod := make(map[int32]map[int32][]string)
	for _, sp := range selectedNodes {
		for _, node := range sp {
			if selectedRacksInSuperPod[node.SuperPodID] == nil {
				selectedRacksInSuperPod[node.SuperPodID] = make(map[int32][]string)
			}
			selectedRacksInSuperPod[node.SuperPodID][node.RackID] =
				append(selectedRacksInSuperPod[node.SuperPodID][node.RackID], node.Name)
		}
	}
	for sp, racks := range selectedRacksInSuperPod {
		if selectedSuperPodInfo[sp] == nil {
			selectedSuperPodInfo[sp] = &selectedRackInfo{}
		}
		selectedSuperPodInfo[sp].selectedRacksInSuperPod = len(racks)
		for _, nodes := range racks {
			selectedSuperPodInfo[sp].selectedNodesInRack =
				append(selectedSuperPodInfo[sp].selectedNodesInRack, len(nodes))
		}
		sort.Ints(selectedSuperPodInfo[sp].selectedNodesInRack)
	}
	return selectedSuperPodInfo
}
