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

// Package chip8node8ra64sp for utils functions ut cases
package chip8node8ra64sp

import (
	"fmt"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

const (
	nodeNum8       = 8
	nodeNum15      = 15
	testRackMapLen = 2
)

func TestTransferSuperPodToRackIdMap(t *testing.T) {
	superPods := buildSuperPodsByParams(map[int32]int32{testSuperPodID1: nodeNum8,
		testSuperPodID2: nodeNum15})
	rackMap := transferSuperPodToRackIdMap(superPods[testSuperPodID1])
	if len(rackMap) != 1 {
		t.Errorf("rackID map len get %d; want 1", len(rackMap))
	}
	rackMap = transferSuperPodToRackIdMap(superPods[testSuperPodID2])
	if len(rackMap) != testRackMapLen {
		t.Errorf("rackID map len get %d; want 2", len(rackMap))
	}
}

func TestSortRackIdByLengthInOneSuperPod(t *testing.T) {
	superPods := buildSuperPodsByParams(map[int32]int32{testSuperPodID1: nodeNum8,
		testSuperPodID2: nodeNum15})
	rackMap := transferSuperPodToRackIdMap(superPods[testSuperPodID1])
	rackIDOrder := sortRackIdByLengthInOneSuperPod(rackMap)
	want := []int32{0}
	if !reflect.DeepEqual(want, rackIDOrder) {
		t.Errorf("sortRackIdByLengthInOneSuperPod result failed: %v", rackIDOrder)
	}
	rackMap = transferSuperPodToRackIdMap(superPods[testSuperPodID2])
	rackIDOrder = sortRackIdByLengthInOneSuperPod(rackMap)
	want = []int32{1, 0}
	if !reflect.DeepEqual(want, rackIDOrder) {
		t.Errorf("sortRackIdByLengthInOneSuperPod result failed: %v", rackIDOrder)
	}
}

func TestGetRackLengthInOneSuperPod(t *testing.T) {
	superPods := buildSuperPodsByParams(map[int32]int32{testSuperPodID1: nodeNum8,
		testSuperPodID2: nodeNum15})
	rackMap := transferSuperPodToRackIdMap(superPods[testSuperPodID1])
	ret := getRackLengthInOneSuperPod(rackMap)
	want := map[int][]int32{
		8: []int32{0},
	}
	if !reflect.DeepEqual(want, ret) {
		t.Errorf("getRackLengthInOneSuperPod result failed: %v", ret)
	}
}

func TestGetRackNPUTop(t *testing.T) {
	tp := New(util.SuperPodx8SchedulerName)
	tp.Nodes = getNPUNodes(nodeInfoIdx0, nodeInfoIdx185, superPodSize32, rackOsNum)
	t.Run("test getRackNPUTop err getting empty top", func(t *testing.T) {
		res := tp.getRackNPUTop(buildNodeBaseInfoArr(nodeInfoIdx9))
		if !reflect.DeepEqual(res, testRackNPUTop) {
			t.Errorf("getUsableNPUIndex fail, getting empty top")
		}
		res1 := tp.getRackNPUTop(buildNodeBaseInfoArr(nodeInfoIdx0))
		if !reflect.DeepEqual(res1, testRackNPUTop) {
			t.Errorf("getUsableNPUIndex fail, getting empty top")
		}
	})
	t.Run("test getRackNPUTop get success", func(t *testing.T) {
		res := tp.getRackNPUTop(buildNodeBaseInfoArr(nodeInfoIdx4))
		if !reflect.DeepEqual(res, testRackNPUTop1) {
			t.Errorf("getUsableNPUIndex fail, the result is %v", res)
		}
		res1 := tp.getRackNPUTop(buildNodeBaseInfoArr(nodeInfoIdx8))
		if !reflect.DeepEqual(res1, testRackNPUTop2) {
			t.Errorf("getUsableNPUIndex fail, the result is %v", res1)
		}
	})
}

// superPodMap key is superPodID, value is how many nodes in the superPod
func buildSuperPodsByParams(superPodMap map[int32]int32) map[int32]superPod {
	result := make(map[int32]superPod)

	for superPodID, nodeCount := range superPodMap {
		podNodes := make(superPod)

		for i := int32(0); i < nodeCount; i++ {
			rackID := i / rackNodeNum
			nodeName := fmt.Sprintf("node-%d-%d", superPodID, i)

			info := nodeBaseInfo{
				name:       nodeName,
				superPodID: superPodID,
				rackID:     rackID,
			}
			podNodes[nodeName] = info
		}
		result[superPodID] = podNodes
	}

	return result
}
