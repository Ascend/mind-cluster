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

// Package chip8node8ra64sp for base function ut
package chip8node8ra64sp

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

const (
	testSchedulerName = "test"
	tpBlock4          = 4
	tpBlock8          = 8
)

// NewPluginTestCase use for New func test
type NewPluginTestCase struct {
	Name    string
	WantErr error

	ScheduleName  string
	PluginName    string
	AnnoName      string
	AnnoPreValue  string
	MaxNodeNPUNum int
}

func buildNewTestCase() []NewPluginTestCase {
	return []NewPluginTestCase{
		{
			Name:          "01-NewTest should return nil when Schedule Name is huawei.com/Ascend910900SuperPod-A5-8",
			ScheduleName:  util.SuperPodx8SchedulerName,
			PluginName:    util.SuperPodx8SchedulerName,
			AnnoName:      util.NPU910CardName,
			AnnoPreValue:  util.NPU910CardNamePre,
			MaxNodeNPUNum: npuNumber8,
			WantErr:       nil,
		},
		{
			Name:          "02-NewTest should return nil when Schedule Name is test",
			ScheduleName:  testSchedulerName,
			PluginName:    testSchedulerName,
			AnnoName:      util.NPU910CardName,
			AnnoPreValue:  util.NPU910CardNamePre,
			MaxNodeNPUNum: npuNumber8,
			WantErr:       nil,
		},
	}
}

func TestNew(t *testing.T) {
	testCases := buildNewTestCase()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			npu := New(tt.ScheduleName)
			if npu.GetPluginName() != tt.PluginName {
				t.Errorf("New() npu Name: %s, wantName: %s.", npu.GetPluginName(), tt.PluginName)
			}
			if npu.GetAnnoName(util.NPU910CardName) != tt.AnnoName {
				t.Errorf("New() npu annoName: %s, wantAnnoName: %s.", npu.GetPluginName(), tt.AnnoName)
			}
			if npu.GetAnnoPreVal(util.NPU910CardName) != tt.AnnoPreValue {
				t.Errorf("New() npu annoNamePre: %s, wantAnnoNamePre: %s.",
					npu.GetPluginName(), tt.AnnoPreValue)
			}
			if npu.MaxNodeNPUNum != tt.MaxNodeNPUNum {
				t.Errorf("New() npu MaxNodeNPUNum: %d, wantMaxNodeNPUNum: %d.",
					npu.MaxNodeNPUNum, tt.MaxNodeNPUNum)
			}
		})
	}
}

func buildTestSelectNodesForJobCases0() []*selectScoreBestNPUNodesTestCase {
	nodeInfo := buildDefaultNodeInfoList()
	selectSuperPodForJobTestCases := []*selectScoreBestNPUNodesTestCase{
		{
			name:             "only one pod scheduled, 1Pod * 8; superPodSize=64, npuTaskNum=1,  but spBlock=1",
			npuTaskNum:       1,
			nodes:            nodeInfo,
			superPodSize:     superPodSize64,
			scheduleStrategy: RackSchedule,
			spBlock:          spBlock1,
			tpBlock:          tpBlock1,
			wantRes: map[int32]*selectedRackInfo{
				0: {
					selectedRacksInSuperPod: 1,
					selectedNodesInRack:     []int{1},
				},
			},
			wantErr: nil,
		},
		{
			name: "OneRack strategy success, 8Pod * 8; SuperPodSize=64, npuTaskNum=8,  but spBlock=1 " +
				"use SuperPodSchedule will success, one SuperPod will be scheduled",
			npuTaskNum:       npuTaskNum8,
			nodes:            nodeInfo,
			superPodSize:     superPodSize64,
			scheduleStrategy: RackSchedule,
			spBlock:          spBlock1,
			tpBlock:          tpBlock1,
			wantRes: map[int32]*selectedRackInfo{
				0: {
					selectedRacksInSuperPod: 1,
					selectedNodesInRack:     []int{8},
				},
			},
			wantErr: nil,
		},
	}
	return selectSuperPodForJobTestCases
}

func buildTestSelectNodesForJobCases1() []*selectScoreBestNPUNodesTestCase {
	nodeInfo := buildDefaultNodeInfoList()
	selectSuperPodForJobTestCases := []*selectScoreBestNPUNodesTestCase{
		{
			name: "OneSuperPod strategy success, case Test 16Pod * 8; SuperPodSize=64, npuTaskNum=16, " +
				"use SuperPodSchedule will success ",
			nodes:            nodeInfo,
			superPodSize:     superPodSize64,
			npuTaskNum:       npuTaskNum16,
			scheduleStrategy: SuperPodSchedule,
			spBlock:          spBlock4,
			tpBlock:          tpBlock2,
			wantRes: map[int32]*selectedRackInfo{
				1: {
					selectedRacksInSuperPod: 2,
					selectedNodesInRack:     []int{8, 8},
				},
			},
			wantErr: nil,
		},
	}
	return selectSuperPodForJobTestCases
}

func buildTestSelectNodesForJobCases2() []*selectScoreBestNPUNodesTestCase {
	nodes := buildDefaultNodeInfoList()
	selectSuperPodForJobTestCases := []*selectScoreBestNPUNodesTestCase{
		{
			name: "MulSuperPod strategy success, case Test 64Pod * 8npus; SuperPodSize=64, npuTaskNum=64;" +
				"use mulSuperPodsSchedule will success; Multiple SuperPods will be scheduled",
			npuTaskNum:       npuTaskNum64,
			nodes:            nodes,
			superPodSize:     superPodSize64,
			scheduleStrategy: SuperPodSchedule,
			spBlock:          spBlock4,
			tpBlock:          tpBlock2,
			wantRes: map[int32]*selectedRackInfo{
				5: {
					selectedRacksInSuperPod: 6,
					selectedNodesInRack:     []int{8, 8, 8, 8, 8, 8},
				},
				4: {
					selectedRacksInSuperPod: 2,
					selectedNodesInRack:     []int{8, 8},
				},
			},
			wantErr: nil,
		},
	}
	return selectSuperPodForJobTestCases
}

func buildTestSelectNodesForJobCases3() []*selectScoreBestNPUNodesTestCase {
	var nodes []*api.NodeInfo
	// one superPod with 2 racks, 1 rack with 2 nodes, and 1 rack with 3 nodes
	nodes = append(nodes, buildNodeInfos(nodeInfoIdx0, nodeInfoIdx1, nil)...)
	nodes = append(nodes, buildNodeInfos(nodeInfoIdx8, nodeInfoIdx10, nil)...)
	selectSuperPodForJobTestCases := []*selectScoreBestNPUNodesTestCase{
		{
			name: "RackSchedule failed, OneSuperPod success. case Test 4Pod * 8npus;  SuperPodSize=64, npuTaskNum=4;" +
				"3 nodes in one rack ,and 2 nodes in another rack; use RackSchedule will failed; " +
				"use SuperPodSchedule will success ,One SuperPod will be scheduled",
			npuTaskNum:       npuTaskNum4,
			nodes:            nodes,
			superPodSize:     superPodSize64,
			scheduleStrategy: RackSchedule,
			spBlock:          spBlock4,
			tpBlock:          tpBlock2,
			wantRes: map[int32]*selectedRackInfo{
				0: {
					selectedRacksInSuperPod: 2,
					selectedNodesInRack:     []int{2, 2},
				},
			},
			wantErr: nil,
		},
		{
			name: "RackSchedule failed. case Test 4Pod * 8npus;  SuperPodSize=64, npuTaskNum=4;" +
				"3 nodes in one rack ,and 2 nodes in another rack; use RackSchedule will failed",
			npuTaskNum:       npuTaskNum4,
			nodes:            nodes,
			superPodSize:     superPodSize64,
			scheduleStrategy: RackSchedule,
			spBlock:          spBlock4,
			tpBlock:          tpBlock4,
			wantRes:          make(map[int32]*selectedRackInfo),
			wantErr:          errors.New("not found the 1 count of tp-block:0 in all racks of every super-pod, exit select process"),
		},
	}
	return selectSuperPodForJobTestCases
}

func buildTestSelectNodesForJobCases4() []*selectScoreBestNPUNodesTestCase {
	var nodes []*api.NodeInfo
	// one superPod with 4 racks: 3/3/1/1
	nodes = append(nodes, buildNodeInfos(nodeInfoIdx0, nodeInfoIdx2, nil)...)
	nodes = append(nodes, buildNodeInfos(nodeInfoIdx8, nodeInfoIdx10, nil)...)
	nodes = append(nodes, buildNodeInfos(nodeInfoIdx16, nodeInfoIdx16, nil)...)
	nodes = append(nodes, buildNodeInfos(nodeInfoIdx24, nodeInfoIdx24, nil)...)
	selectSuperPodForJobTestCases := []*selectScoreBestNPUNodesTestCase{
		{
			name: "case Test 6Pod * 8npus;  SuperPodSize=64, npuTaskNum=6;" +
				"return err: schedule job failed, not enough nodes available for this job",
			npuTaskNum:       npuTaskNum6,
			nodes:            nodes,
			superPodSize:     superPodSize64,
			scheduleStrategy: SuperPodSchedule,
			spBlock:          spBlock6,
			tpBlock:          tpBlock2,
			wantRes:          map[int32]*selectedRackInfo{},
			wantErr:          fmt.Errorf("scheduling failed in %s strategy", MulSuperPodsSchedule),
		},
	}
	return selectSuperPodForJobTestCases
}

// schedule failed because of tp-block = 64 in every Rack
func buildTestSelectNodesForJobCases5() []*selectScoreBestNPUNodesTestCase {
	var nodes []*api.NodeInfo
	// one superPod with four racks：4 node、 4 node、 4 node、 4 node
	nodes = append(nodes, buildNodeInfos(nodeInfoIdx0, nodeInfoIdx3, nil)...)
	nodes = append(nodes, buildNodeInfos(nodeInfoIdx8, nodeInfoIdx11, nil)...)
	nodes = append(nodes, buildNodeInfos(nodeInfoIdx16, nodeInfoIdx19, nil)...)
	nodes = append(nodes, buildNodeInfos(nodeInfoIdx24, nodeInfoIdx27, nil)...)
	selectSuperPodForJobTestCases := []*selectScoreBestNPUNodesTestCase{
		{
			name: "case Test 8Pod * 8npus;  SuperPodSize=64, npuTaskNum=8;" +
				"return err: schedule job failed, not enough nodes available for this job",
			npuTaskNum:       npuTaskNum16,
			nodes:            nodes,
			superPodSize:     superPodSize64,
			scheduleStrategy: SuperPodSchedule,
			spBlock:          spBlock8,
			tpBlock:          tpBlock8,
			wantRes:          map[int32]*selectedRackInfo{},
			wantErr:          fmt.Errorf("scheduling failed in %s strategy", MulSuperPodsSchedule),
		},
	}
	return selectSuperPodForJobTestCases
}

func buildTestSelectNodesForJobCases6() []*selectScoreBestNPUNodesTestCase {
	nodes := buildDefaultNodeInfoList()
	selectSuperPodForJobTestCases := []*selectScoreBestNPUNodesTestCase{
		{
			name:             "case Test 32Pod * 8npus; ; SuperPodSize=64; use SuperPodSchedule will failed",
			npuTaskNum:       npuTaskNum32,
			nodes:            nodes,
			npuNodes:         getNPUNodes(nodeInfoIdx0, nodeInfoIdx80, superPodSize64, rackNodeNum),
			superPodSize:     superPodSize64,
			scheduleStrategy: SuperPodSchedule,
			spBlock:          spBlock8,
			tpBlock:          tpBlock2,
			wantRes:          map[int32]*selectedRackInfo{},
			wantErr:          fmt.Errorf("scheduling failed in %s strategy", MulSuperPodsSchedule),
		},
	}
	return selectSuperPodForJobTestCases
}

func buildTestSelectNodesForJobCases() []*selectScoreBestNPUNodesTestCase {
	// get all test cases for testing
	selectNodesForJobCases := buildTestSelectNodesForJobCases0()
	selectNodesForJobCases = append(selectNodesForJobCases,
		buildTestSelectNodesForJobCases1()...)
	selectNodesForJobCases = append(selectNodesForJobCases,
		buildTestSelectNodesForJobCases2()...)
	selectNodesForJobCases = append(selectNodesForJobCases,
		buildTestSelectNodesForJobCases3()...)
	selectNodesForJobCases = append(selectNodesForJobCases,
		buildTestSelectNodesForJobCases4()...)
	selectNodesForJobCases = append(selectNodesForJobCases,
		buildTestSelectNodesForJobCases5()...)
	selectNodesForJobCases = append(selectNodesForJobCases,
		buildTestSelectNodesForJobCases6()...)
	return selectNodesForJobCases
}

func TestSelectNodesForJob(t *testing.T) {
	selectNodesForJobCases := buildTestSelectNodesForJobCases()
	tasks := getTaskInfos(npuTaskNum2, "job1")
	for _, cs := range selectNodesForJobCases {
		t.Run(cs.name, func(t *testing.T) {
			plg := superPodModelForTest(tasks, cs)
			res, err := plg.selectNodesForJob(tasks[npuTaskNum0], cs.nodes)
			if !reflect.DeepEqual(err, cs.wantErr) {
				t.Errorf("ScoreBestNPUNodes() error = %v, wantErr %v", err, cs.wantErr)
			}
			selectedSuperPodInfo := checkScoreBestNPUNodesResult(res)
			if !reflect.DeepEqual(selectedSuperPodInfo, cs.wantRes) {
				t.Errorf("ScoreBestNPUNodes() fault result = %v, want %v;", selectedSuperPodInfo, cs.wantRes)
			}
		})
	}
}

// 8sp soft schedule
func buildSelectScoreBestNPUNodesTestCases12() []*selectScoreBestNPUNodesTestCase {
	var nodeInfo1 []*api.NodeInfo
	nodeInfo1 = append(nodeInfo1, buildNodeInfos(nodeInfoIdx0, nodeInfoIdx0, nil)...)
	nodeInfo1 = append(nodeInfo1, buildNodeInfos(nodeInfoIdx8, nodeInfoIdx8, nil)...)
	selectSuperPodForJobTestCases := []*selectScoreBestNPUNodesTestCase{
		{
			name: "case Test 2Pod * 8卡; ; SuperPodSize=64, npuTaskNum=2;" +
				"use MultiSuperPodSchedule will failed, use soft schedule will success",
			tasks:            newNPUTasks(npuTaskNum2, nodeNPUNum),
			npuTaskNum:       npuTaskNum2,
			nodes:            nodeInfo1,
			superPodSize:     superPodSize64,
			npuNodes:         getNPUNodes(nodeInfoIdx0, nodeInfoIdx80, superPodSize64, rackNodeNum),
			scheduleStrategy: MulSuperPodsSchedule,
			spBlock:          tpBlock2,
			tpBlock:          tpBlock2,
			wantRes: map[int32]*selectedRackInfo{
				0: {
					selectedRacksInSuperPod: 2,
					selectedNodesInRack:     []int{1, 1},
				},
			},
			wantErr: nil,
		},
	}
	return selectSuperPodForJobTestCases
}

func TestScoreBestNPUNode2(t *testing.T) {
	selectScoreBestNPUNodesTestCases := buildSelectScoreBestNPUNodesTestCases12()

	tasks := getTaskInfos(npuTaskNum2, "job1")
	scoreMap := make(map[string]float64)
	for _, cs := range selectScoreBestNPUNodesTestCases {
		t.Run(cs.name, func(t *testing.T) {
			plg := packageModuleSuperPod4Soft(tasks, cs)
			plg.Label = map[string]string{superPodAffinity: softRequire}
			plg.tpBlock = cs.tpBlock
			suitableNode := cs.nodes
			for _, index := range cs.nodes {
				scoreMap[index.Name] = 0.0
			}
			err := plg.ScoreBestNPUNodes(tasks[npuTaskNum0], suitableNode, scoreMap)
			if !reflect.DeepEqual(err, cs.wantErr) {
				t.Errorf("ScoreBestNPUNodes() error = %v, wantErr %v", err, cs.wantErr)
			}
			checkScoreBestNPUNodesResult(plg.Jobs[tasks[npuTaskNum0].Job].SuperPods)
		})
	}
}

func buildTestScoreBestNPUNodes0() []*selectScoreBestNPUNodesTestCase {
	nodeInfo := buildDefaultNodeInfoList()
	selectSuperPodForJobTestCases := []*selectScoreBestNPUNodesTestCase{
		{
			name:             "only one pod scheduled, 1Pod * 8; superPodSize=64, npuTaskNum=1, spBlock=1",
			npuTaskNum:       1,
			nodes:            nodeInfo,
			superPodSize:     superPodSize64,
			scheduleStrategy: RackSchedule,
			spBlock:          spBlock1,
			tpBlock:          tpBlock1,
			wantRes: map[int32]*selectedRackInfo{
				0: {
					selectedRacksInSuperPod: 1,
					selectedNodesInRack:     []int{1},
				},
			},
			wantErr: nil,
		},
		{
			name:             "nodes len is 0 should return error",
			npuTaskNum:       4,
			nodes:            make([]*api.NodeInfo, 0),
			superPodSize:     superPodSize64,
			scheduleStrategy: RackSchedule,
			spBlock:          spBlock1,
			tpBlock:          tpBlock1,
			wantRes:          make(map[int32]*selectedRackInfo),
			wantErr:          errors.New("invalid argument"),
		},
	}
	return selectSuperPodForJobTestCases
}

func TestScoreBestNPUNodes(t *testing.T) {
	selectNodesForJobCases := buildTestScoreBestNPUNodes0()
	tasks := getTaskInfos(npuTaskNum2, "job1")
	scoreMap := make(map[string]float64)

	for _, cs := range selectNodesForJobCases {
		t.Run(cs.name, func(t *testing.T) {
			for _, index := range cs.nodes {
				scoreMap[index.Name] = 0.0
			}
			plg := superPodModelForTest(tasks, cs)
			err := plg.ScoreBestNPUNodes(tasks[npuTaskNum0], cs.nodes, scoreMap)
			if !reflect.DeepEqual(err, cs.wantErr) {
				t.Errorf("ScoreBestNPUNodes() error = %v, wantErr %v", err, cs.wantErr)
			}
			selectedSuperPodInfo := checkScoreBestNPUNodesResult(plg.Jobs[tasks[npuTaskNum0].Job].SuperPods)
			if !reflect.DeepEqual(selectedSuperPodInfo, cs.wantRes) {
				t.Errorf("ScoreBestNPUNodes() fault result = %v, want %v;", selectedSuperPodInfo, cs.wantRes)
			}
		})
	}
}
