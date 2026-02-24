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

// Package chip8node8ra64sp for
package chip8node8ra64sp

import (
	"fmt"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

const (
	superPodSize128 = 128
)

func buildOneRackStrategyCase0() []*selectScoreBestNPUNodesTestCase {
	selectSuperPodForJobTestCases := []*selectScoreBestNPUNodesTestCase{
		{
			name: "test_case_001 1Pod * 8卡; SuperPodSize=128, npuTaskNum=1, spBlock=1; rack schedule failed" +
				"no nodes to be scheduled",
			superPodMap:      buildSuperPodMapWithUBMemByParams(map[int32]map[int32]int32{}),
			npuTaskNum:       npuTaskNum1,
			superPodSize:     superPodSize128,
			scheduleStrategy: RackSchedule,
			spBlock:          spBlock1,
			tpBlock:          tpBlock1,
			wantRes:          make(map[int32]*selectedRackInfo),
			wantErr:          fmt.Errorf("before scheduling, required sp-block count=1, but total sp-block count=0"),
		},
		{
			name: "test_case_002 8Pod * 8卡; SuperPodSize=128, npuTaskNum=8, spBlock=8, tpBlock=4; rack schedule success",
			superPodMap: buildSuperPodMapWithUBMemByParams(map[int32]map[int32]int32{
				0: {0: 15,
					1: 8},
				1: {0: 5},
			}),
			npuTaskNum:       npuTaskNum8,
			superPodSize:     superPodSize128,
			scheduleStrategy: RackSchedule,
			spBlock:          spBlock8,
			tpBlock:          spBlock8,
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

func buildOneUBMemStrategyCase0() []*selectScoreBestNPUNodesTestCase {
	selectSuperPodForJobTestCases := []*selectScoreBestNPUNodesTestCase{
		{
			name: "test_case_003 32Pod * 8卡; SuperPodSize=64, npuTaskNum=32, spBlock=4; ubMem schedule success" +
				"with priority given to the SuperPod with fewer remaining numbers.",
			superPodMap: buildSuperPodMapWithUBMemByParams(map[int32]map[int32]int32{
				0: {0: 32,
					1: 32},
				1: {0: 32,
					1: 8},
			}),
			npuTaskNum:       npuTaskNum32,
			superPodSize:     superPodSize64,
			scheduleStrategy: UBMemSchedule,
			spBlock:          spBlock4,
			tpBlock:          tpBlock4,
			wantRes: map[int32]*selectedRackInfo{
				1: {
					selectedRacksInSuperPod: 4,
					selectedNodesInRack:     []int{8, 8, 8, 8},
				},
			},
			wantErr: nil,
		},
		{
			name: "test_case_004 32Pod * 8卡; SuperPodSize=128, npuTaskNum=32, spBlock=4; ubMem schedule success" +
				"with priority given to the SuperPod with fewer remaining numbers.",
			superPodMap: buildSuperPodMapWithUBMemByParams(map[int32]map[int32]int32{
				0: {0: 32,
					1: 32},
				1: {0: 32,
					1: 48},
			}),
			npuTaskNum:       npuTaskNum32,
			superPodSize:     superPodSize128,
			scheduleStrategy: UBMemSchedule,
			spBlock:          spBlock4,
			tpBlock:          tpBlock4,
			wantRes: map[int32]*selectedRackInfo{
				0: {
					selectedRacksInSuperPod: 4,
					selectedNodesInRack:     []int{8, 8, 8, 8},
				},
			},
			wantErr: nil,
		},
	}
	return selectSuperPodForJobTestCases
}

func buildOneUBMemStrategyCase1() []*selectScoreBestNPUNodesTestCase {
	selectSuperPodForJobTestCases := []*selectScoreBestNPUNodesTestCase{
		{
			name: "test_case_005 32Pod * 8卡; SuperPodSize=64, npuTaskNum=32, spBlock=4, tpBlock=4; ubMem schedule success" +
				"Supernode priority is sorted from smallest to largest according to the number of nodes, giving priority to those that can perform UBMem scheduling.",
			superPodMap: buildSuperPodMapWithUBMemByParams(map[int32]map[int32]int32{
				0: {0: 8,
					1: 24},
				1: {0: 32,
					1: 32},
			}),
			npuTaskNum:       npuTaskNum32,
			superPodSize:     superPodSize64,
			scheduleStrategy: UBMemSchedule,
			spBlock:          spBlock4,
			tpBlock:          tpBlock4,
			wantRes: map[int32]*selectedRackInfo{
				1: {
					selectedRacksInSuperPod: 4,
					selectedNodesInRack:     []int{8, 8, 8, 8},
				},
			},
			wantErr: nil,
		},
		{
			name: "test_case_007 40Pod * 8卡; SuperPodSize=64, npuTaskNum=40, spBlock=8, tpBlock=4; ubMem schedule failed" +
				"spBlock is greater than size of UBMem",
			superPodMap: buildSuperPodMapWithUBMemByParams(map[int32]map[int32]int32{
				0: {0: 8,
					1: 24},
				1: {0: 32,
					1: 32},
			}),
			npuTaskNum:       npuTaskNum40,
			superPodSize:     superPodSize64,
			scheduleStrategy: UBMemSchedule,
			spBlock:          spBlock8,
			tpBlock:          tpBlock4,
			wantRes:          map[int32]*selectedRackInfo{},
			wantErr:          fmt.Errorf("scheduling failed in UBMemSchedule strategy"),
		},
	}
	return selectSuperPodForJobTestCases
}

func getAllStrategyCases() []*selectScoreBestNPUNodesTestCase {
	cases := make([]*selectScoreBestNPUNodesTestCase, 0)
	cases = append(cases, buildOneRackStrategyCase0()...)
	cases = append(cases, buildOneUBMemStrategyCase0()...)
	cases = append(cases, buildOneUBMemStrategyCase1()...)
	return cases
}

func TestSelectNodesFromSuperPods(t *testing.T) {
	allCases := getAllStrategyCases()

	tasks := getTaskInfos(npuTaskNum2, "job1")
	getUnReadyIDs := func(taskNums int) []string {
		s := make([]string, 0)
		for i := 0; i < taskNums; i++ {
			s = append(s, fmt.Sprintf("%d", i))
		}
		return s
	}
	for _, cs := range allCases {
		t.Run(cs.name, func(t *testing.T) {
			selectedNodes := make(map[string][]plugin.SuperNode)
			// call New to init chip8node8ra64sp
			plg := superPodModelForTest(tasks, cs)
			unReadyID := getUnReadyIDs(cs.npuTaskNum / plg.spBlock)
			strategyInitFactory(plg, unReadyID, selectedNodes)
			// set ubMem tag
			plg.uBMemParams = uBMemParams{
				isUBMemScene: true,
				uBMemRackNum: uBMemRackNumber, // 16
			}
			err := plg.selectNodesFromSuperPods(cs.superPodMap, unReadyID, selectedNodes)
			if !reflect.DeepEqual(err, cs.wantErr) {
				t.Errorf("EntrySelect error = %v but wantErr %v", err, cs.wantErr)
			}
			if !reflect.DeepEqual(checkScoreBestNPUNodesResult(selectedNodes), cs.wantRes) {
				t.Errorf("ScoreBestNPUNodes() fault result = %v, want %#v", selectedNodes, cs.wantRes)
			}
		})
	}
}

func buildSuperPodMapWithUBMemByParams(superPodMap map[int32]map[int32]int32) map[int32]superPod {
	result := make(map[int32]superPod)

	for superPodID, uBMemMap := range superPodMap {
		podNodes := make(superPod)
		for uBMemId, nodeCount := range uBMemMap {
			var i int32
			for i = 0; i < nodeCount; i++ {
				rackID := uBMemId*uBMemRackNumber + i/rackNodeNum
				nodeName := fmt.Sprintf("node-%d-%d-%d", superPodID, uBMemId, i)
				info := nodeBaseInfo{
					name:       nodeName,
					superPodID: superPodID,
					ubMemID:    uBMemId,
					rackID:     rackID,
				}
				podNodes[nodeName] = info
			}
		}
		result[superPodID] = podNodes
	}

	return result
}
