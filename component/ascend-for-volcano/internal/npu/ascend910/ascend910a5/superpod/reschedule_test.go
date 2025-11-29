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

// Package superpod for reschedule ut
package superpod

import (
	"fmt"
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

type getSameRackNodesTestArgs struct {
	faultNodeNameMap map[string]struct{}
	vSuperPod        []plugin.SuperNode
	spNodeMaps       map[string]nodeBaseInfo
}

type getSameRackNodesTestResult struct {
	spNodeMaps       map[string]nodeBaseInfo
	nodes            []plugin.SuperNode
	faultNodeNameMap map[string]struct{}
}

type getSameRackNodesTest struct {
	name string
	args getSameRackNodesTestArgs
	want getSameRackNodesTestResult
}

func buildGetSameRackNodesTest1() getSameRackNodesTest {
	return getSameRackNodesTest{
		name: "01-getSameRackNodes nil spNodeMaps",
		args: getSameRackNodesTestArgs{
			faultNodeNameMap: nil,
			vSuperPod:        make([]plugin.SuperNode, 0),
			spNodeMaps:       nil,
		},
		want: getSameRackNodesTestResult{
			spNodeMaps: nil,
			nodes:      nil,
		},
	}
}

func getSameRackNodesTest2VSuperPod() []plugin.SuperNode {
	return []plugin.SuperNode{
		{
			Name:       "work1",
			RackID:     0,
			SuperPodID: 0,
		},
		{
			Name:       "work2",
			RackID:     0,
			SuperPodID: 0,
		},
		{
			Name:       "work3",
			RackID:     0,
			SuperPodID: 0,
		},
		{
			Name:       "work4",
			RackID:     0,
			SuperPodID: 0,
		},
	}
}

func getSameRackNodesTest2WantNodes() []plugin.SuperNode {
	return []plugin.SuperNode{
		{
			Name:       "work2",
			RackID:     0,
			SuperPodID: 0,
		},
		{
			Name:       "work3",
			RackID:     0,
			SuperPodID: 0,
		},
		{
			Name:       "work4",
			RackID:     0,
			SuperPodID: 0,
		},
		{
			Name:       "work8",
			RackID:     0,
			SuperPodID: 0,
		},
	}
}

func buildGetSameRackNodesTest2() getSameRackNodesTest {
	return getSameRackNodesTest{
		name: "02-getSameRackNodes success",
		args: getSameRackNodesTestArgs{
			faultNodeNameMap: map[string]struct{}{
				"work1": {},
			},
			vSuperPod: getSameRackNodesTest2VSuperPod(),
			spNodeMaps: map[string]nodeBaseInfo{
				"work6": {
					name:       "work6",
					rackID:     1,
					superPodID: 0,
				},
				"work8": {
					name:       "work8",
					rackID:     0,
					superPodID: 0,
				},
			},
		},
		want: getSameRackNodesTestResult{
			spNodeMaps: map[string]nodeBaseInfo{
				"work6": {
					name:       "work6",
					rackID:     1,
					superPodID: 0,
				},
			},
			nodes: getSameRackNodesTest2WantNodes(),
			faultNodeNameMap: map[string]struct{}{
				"work1": {},
				"work8": {},
			},
		},
	}
}

func buildGetSameRackNodesTests() []getSameRackNodesTest {
	test1 := buildGetSameRackNodesTest1()
	test2 := buildGetSameRackNodesTest2()
	return []getSameRackNodesTest{test1, test2}
}

func superNodeListToMap(list []plugin.SuperNode) map[string]bool {
	m := make(map[string]bool)
	for _, node := range list {
		m[node.Name] = true
	}
	return m
}

func TestGetSameRackNodes(t *testing.T) {
	tests := buildGetSameRackNodesTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes := getSameRackNodes(tt.args.faultNodeNameMap, tt.args.vSuperPod, tt.args.spNodeMaps)
			if !reflect.DeepEqual(tt.args.spNodeMaps, tt.want.spNodeMaps) {
				t.Errorf("Test %s failed: expected result: %v, got: %v", tt.name, tt.want.spNodeMaps, tt.args.spNodeMaps)
			}
			if !reflect.DeepEqual(superNodeListToMap(nodes), superNodeListToMap(tt.want.nodes)) {
				t.Errorf("Test %s failed: expected result: %v, got: %v", tt.name, tt.want.nodes, nodes)
			}
			if !reflect.DeepEqual(tt.args.faultNodeNameMap, tt.want.faultNodeNameMap) {
				t.Errorf("Test %s failed: expected result: %v, got: %v", tt.name, tt.args.faultNodeNameMap, tt.want.faultNodeNameMap)
			}
		})
	}
}

type selectNodesByRackTestArgs struct {
	fJob             *rescheduling.FaultJob
	notReadySuperPod map[string]struct{}
	totalNodes       map[int32]superPod
	virtualIdArr     map[string]bool
	selectNodes      map[string][]plugin.SuperNode
	tp               module910a5SuperPod
}

type selectNodesByRackTestResult struct {
	ret         error
	selectNodes map[string][]plugin.SuperNode
}

type selectNodesByRackTest struct {
	name string
	args selectNodesByRackTestArgs
	want selectNodesByRackTestResult
}

func buildSelectNodesByRackTest1() selectNodesByRackTest {
	return selectNodesByRackTest{
		name: "01-selectNodesByRackTest1 nil input",
		args: selectNodesByRackTestArgs{
			fJob:             &rescheduling.FaultJob{},
			notReadySuperPod: make(map[string]struct{}),
			totalNodes:       nil,
			virtualIdArr:     nil,
			selectNodes:      nil,
			tp:               module910a5SuperPod{},
		},

		want: selectNodesByRackTestResult{
			ret:         nil,
			selectNodes: nil,
		},
	}
}

func buildSelectNodesByRackTest2() selectNodesByRackTest {
	return selectNodesByRackTest{
		name: "03-selectNodesByRackTest3 getSameRackNodes",
		args: selectNodesByRackTestArgs{
			fJob: &rescheduling.FaultJob{
				JobUID:            "job1",
				PendingSessionNum: 1,
				SuperPods: map[string][]plugin.SuperNode{
					"0": {
						{SuperPodID: 1},
					},
				},
			},
			notReadySuperPod: map[string]struct{}{
				"0": {},
			},
			totalNodes: map[int32]superPod{
				1: {},
			},
			virtualIdArr: make(map[string]bool),
			selectNodes:  make(map[string][]plugin.SuperNode),
			tp:           module910a5SuperPod{},
		},
		want: selectNodesByRackTestResult{
			ret: nil,
			selectNodes: map[string][]plugin.SuperNode{
				"0": {
					{SuperPodID: 1},
				},
			},
		},
	}
}

func buildSelectNodesByRackTest3() selectNodesByRackTest {
	return selectNodesByRackTest{
		name: "04-selectNodesByRackTest3 getSameSuperPod",
		args: selectNodesByRackTestArgs{
			fJob: &rescheduling.FaultJob{
				JobUID:            "job1",
				PendingSessionNum: 8,
				SuperPods: map[string][]plugin.SuperNode{
					"0": {
						{SuperPodID: 1},
					},
				},
			},
			notReadySuperPod: map[string]struct{}{
				"0": {},
			},
			totalNodes: map[int32]superPod{
				1: {},
			},
			virtualIdArr: make(map[string]bool),
			selectNodes:  make(map[string][]plugin.SuperNode),
			tp:           module910a5SuperPod{},
		},
		want: selectNodesByRackTestResult{
			ret:         fmt.Errorf("there is no enough nodes for whole rack schedule"),
			selectNodes: map[string][]plugin.SuperNode{},
		},
	}
}

func buildSelectNodesByRackTests() []selectNodesByRackTest {
	test1 := buildSelectNodesByRackTest1()
	test2 := buildSelectNodesByRackTest2()
	test3 := buildSelectNodesByRackTest3()
	return []selectNodesByRackTest{test1, test2, test3}
}

func TestSelectNodesByRack(t *testing.T) {
	tests := buildSelectNodesByRackTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &tt.args.tp
			err := tp.selectNodesByRack(tt.args.fJob,
				tt.args.notReadySuperPod,
				tt.args.totalNodes,
				tt.args.virtualIdArr,
				tt.args.selectNodes)

			if (err != nil && tt.want.ret == nil) || (err == nil && tt.want.ret != nil) ||
				(err != nil && tt.want.ret != nil && err.Error() != tt.want.ret.Error()) {
				t.Errorf("Test %s failed: expected error: %v, got: %v", tt.name, tt.want.ret, err)
			}

			if !reflect.DeepEqual(tt.args.selectNodes, tt.want.selectNodes) {
				t.Errorf("Test %s failed: expected selectNodes: %v, got: %v",
					tt.name, tt.want.selectNodes, tt.args.selectNodes)
			}
		})
	}
}

type getRestRackLenMapIdTestArgs struct {
	tpBlock        int
	restRackLenMap map[int][]int32
}

type getRestRackLenMapIdTest struct {
	name string
	args getRestRackLenMapIdTestArgs
	want int32
}

func buildGetRestRackLenMapIdTest1() getRestRackLenMapIdTest {
	return getRestRackLenMapIdTest{
		name: "01-getRestRackLenMapIdTest1 found match at tpBlock",
		args: getRestRackLenMapIdTestArgs{
			tpBlock: 4,
			restRackLenMap: map[int][]int32{
				4: {1001, 1002},
				5: {2001},
			},
		},
		want: 1001,
	}
}

func buildGetRestRackLenMapIdTest2() getRestRackLenMapIdTest {
	return getRestRackLenMapIdTest{
		name: "02-getRestRackLenMapIdTest2 found match after tpBlock",
		args: getRestRackLenMapIdTestArgs{
			tpBlock: 3,
			restRackLenMap: map[int][]int32{
				5: {3001},
			},
		},
		want: 3001,
	}
}

func buildGetRestRackLenMapIdTest3() getRestRackLenMapIdTest {
	return getRestRackLenMapIdTest{
		name: "03-getRestRackLenMapIdTest3 no match found",
		args: getRestRackLenMapIdTestArgs{
			tpBlock:        6,
			restRackLenMap: map[int][]int32{2: {4001}},
		},
		want: UninitializedRestRackLenMapId,
	}
}

func buildGetRestRackLenMapIdTests() []getRestRackLenMapIdTest {
	return []getRestRackLenMapIdTest{
		buildGetRestRackLenMapIdTest1(),
		buildGetRestRackLenMapIdTest2(),
		buildGetRestRackLenMapIdTest3(),
	}
}

func TestGetRestRackLenMapId(t *testing.T) {
	tests := buildGetRestRackLenMapIdTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &module910a5SuperPod{
				jobParams: jobParams{
					tpBlock: tt.args.tpBlock,
				},
			}
			got := tp.getRestRackLenMapId(tt.args.restRackLenMap)
			if got != tt.want {
				t.Errorf("Test %s failed: expected %v, got %v", tt.name, tt.want, got)
			}
		})
	}
}

type getAnotherRackNodesTestArgs struct {
	faultNodeNameMap   map[string]struct{}
	vSuperPod          []plugin.SuperNode
	totalNodes         map[string]nodeBaseInfo
	superPodWithRackId []nodeBaseInfo
}

type getAnotherRackNodesWant struct {
	ret              []plugin.SuperNode
	faultNodeNameMap map[string]struct{}
	totalNodes       map[string]nodeBaseInfo
}
type getAnotherRackNodesTest struct {
	name string
	args getAnotherRackNodesTestArgs
	want getAnotherRackNodesWant
}

func buildGetAnotherRackNodesTest1() getAnotherRackNodesTest {
	return getAnotherRackNodesTest{
		name: "01-getAnotherRackNodesTest1 len(superPodWithRackId) = 0",
		args: getAnotherRackNodesTestArgs{
			faultNodeNameMap: map[string]struct{}{},
			vSuperPod: []plugin.SuperNode{
				{Name: "node1", SuperPodID: 1, RackID: 1},
				{Name: "node2", SuperPodID: 1, RackID: 1},
			},
			totalNodes:         map[string]nodeBaseInfo{},
			superPodWithRackId: []nodeBaseInfo{},
		},
		want: getAnotherRackNodesWant{
			ret:              nil,
			faultNodeNameMap: map[string]struct{}{},
			totalNodes:       map[string]nodeBaseInfo{},
		},
	}
}

func getAnotherRackNodesTestVSuperPod() []plugin.SuperNode {
	return []plugin.SuperNode{
		{
			Name:   "work1",
			RackID: 0,

			SuperPodID: 0,
		},
		{
			Name:   "work2",
			RackID: 0,

			SuperPodID: 0,
		},
		{
			Name:   "work3",
			RackID: 0,

			SuperPodID: 0,
		},
		{
			Name:   "work4",
			RackID: 0,

			SuperPodID: 0,
		},
	}
}

func buildGetAnotherRackNodesTest2() getAnotherRackNodesTest {
	return getAnotherRackNodesTest{
		name: "02-getAnotherRackNodesTest2 getAnotherRackNodes success",
		args: getAnotherRackNodesTestArgs{
			faultNodeNameMap: map[string]struct{}{"work1": {}},
			vSuperPod:        getAnotherRackNodesTestVSuperPod(),
			totalNodes: map[string]nodeBaseInfo{
				"work9":  {name: "work9", superPodID: 0, rackID: 1},
				"work15": {name: "work15", superPodID: 1, rackID: 2},
				"work16": {name: "work16", superPodID: 1, rackID: 2},
				"work18": {name: "work18", superPodID: 1, rackID: 2},
			},
			superPodWithRackId: []nodeBaseInfo{
				{name: "work9", superPodID: 0, rackID: 1},
			},
		},
		want: getAnotherRackNodesWant{
			ret: []plugin.SuperNode{
				{Name: "work9", RackID: 1, SuperPodID: 0},
				{Name: "work2", RackID: 0, SuperPodID: 0},
				{Name: "work3", RackID: 0, SuperPodID: 0},
				{Name: "work4", RackID: 0, SuperPodID: 0},
			},
			faultNodeNameMap: map[string]struct{}{"work1": {}},
			totalNodes: map[string]nodeBaseInfo{
				"work15": {name: "work15", superPodID: 1, rackID: 2},
				"work16": {name: "work16", superPodID: 1, rackID: 2},
				"work18": {name: "work18", superPodID: 1, rackID: 2},
			},
		},
	}
}

func buildGetAnotherRackNodesTest3() getAnotherRackNodesTest {
	return getAnotherRackNodesTest{
		name: "02-getAnotherRackNodesTest2 not enough nodes",
		args: getAnotherRackNodesTestArgs{
			faultNodeNameMap: map[string]struct{}{"work1": {}, "work2": {}},
			vSuperPod:        getAnotherRackNodesTestVSuperPod(),
			totalNodes: map[string]nodeBaseInfo{
				"work9":  {name: "work9", superPodID: 0, rackID: 1},
				"work15": {name: "work15", superPodID: 1, rackID: 2},
				"work16": {name: "work16", superPodID: 1, rackID: 2},
				"work18": {name: "work18", superPodID: 1, rackID: 2},
			},
			superPodWithRackId: []nodeBaseInfo{
				{name: "work9", superPodID: 0, rackID: 1},
			},
		},
		want: getAnotherRackNodesWant{
			ret: []plugin.SuperNode{
				{
					Name:       "work9",
					RackID:     1,
					SuperPodID: 0,
				},
				{
					Name:       "work3",
					RackID:     0,
					SuperPodID: 0,
				},
				{
					Name:       "work4",
					RackID:     0,
					SuperPodID: 0,
				},
			},
			faultNodeNameMap: map[string]struct{}{"work1": {}, "work2": {}},
			totalNodes: map[string]nodeBaseInfo{
				"work15": {name: "work15", superPodID: 1, rackID: 2},
				"work16": {name: "work16", superPodID: 1, rackID: 2},
				"work18": {name: "work18", superPodID: 1, rackID: 2},
			},
		},
	}
}

func buildGetAnotherRackNodesTests() []getAnotherRackNodesTest {
	test1 := buildGetAnotherRackNodesTest1()
	test2 := buildGetAnotherRackNodesTest2()
	test3 := buildGetAnotherRackNodesTest3()
	return []getAnotherRackNodesTest{test1, test2, test3}
}

func TestGetAnotherRackNodes(t *testing.T) {
	tests := buildGetAnotherRackNodesTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getAnotherRackNodes(tt.args.faultNodeNameMap, tt.args.vSuperPod, tt.args.totalNodes, tt.args.superPodWithRackId)

			if !reflect.DeepEqual(got, tt.want.ret) {
				t.Errorf("Test %s failed: expected %+v, got %+v", tt.name, tt.want.ret, got)
			}

			if !reflect.DeepEqual(tt.args.faultNodeNameMap, tt.want.faultNodeNameMap) {
				t.Errorf("Test %s failed: expected result: %v, got: %v", tt.name, tt.want.faultNodeNameMap, tt.args.faultNodeNameMap)
			}
			if !reflect.DeepEqual(tt.args.totalNodes, tt.want.totalNodes) {
				t.Errorf("Test %s failed: expected result: %v, got: %v", tt.name, tt.want.totalNodes, tt.args.totalNodes)
			}

		})
	}
}

type selectNodeFromOriginSpBlockTestArgs struct {
	fJob         *rescheduling.FaultJob
	selectNodes  map[string][]plugin.SuperNode
	totalNodes   map[int32]superPod
	virtualIdArr map[string]bool
	tp           *module910a5SuperPod
}

type selectNodeFromOriginSpBlockTest struct {
	name string
	args selectNodeFromOriginSpBlockTestArgs
	want map[string]struct{}
}

func buildSelectNodeFromOriginSpBlockTest1() selectNodeFromOriginSpBlockTest {
	jobID := api.JobID("job1")
	return selectNodeFromOriginSpBlockTest{
		name: "01-selectNodeFromOriginSpBlock all superNodes healthy and satisfy affinity",
		args: selectNodeFromOriginSpBlockTestArgs{
			fJob: &rescheduling.FaultJob{
				JobUID:            jobID,
				PendingSessionNum: 1,
				SuperPods: map[string][]plugin.SuperNode{
					"vsp1": {
						{Name: "work1", SuperPodID: 1},
						{Name: "work2", SuperPodID: 1},
					},
				},
				FaultTasks: []rescheduling.FaultTask{
					{NodeName: "work1", IsFaultTask: false,
						FaultTaskA5Field: rescheduling.FaultTaskA5Field{IsSatisfiedRackAffinity: true}},
					{NodeName: "work2", IsFaultTask: false,
						FaultTaskA5Field: rescheduling.FaultTaskA5Field{IsSatisfiedRackAffinity: true}},
				},
			},
			selectNodes: map[string][]plugin.SuperNode{},
			totalNodes: map[int32]superPod{
				1: {
					"work1": {name: "work1", superPodID: 1},
					"work2": {name: "work2", superPodID: 1},
				},
			},
			virtualIdArr: map[string]bool{
				"vsp1": false,
			},
			tp: &module910a5SuperPod{
				jobParams: jobParams{tpBlock: tpBlock2},
			},
		},
		want: map[string]struct{}{},
	}
}

func buildSelectNodeFromOriginSpBlockTest2() selectNodeFromOriginSpBlockTest {
	jobID := api.JobID("job1")
	return selectNodeFromOriginSpBlockTest{
		name: "02-selectNodeFromOriginSpBlock some superNodes not healthy or not satisfy affinity",
		args: selectNodeFromOriginSpBlockTestArgs{
			fJob: &rescheduling.FaultJob{
				JobUID:            jobID,
				PendingSessionNum: 2,
				SuperPods: map[string][]plugin.SuperNode{
					"vsp1": {
						{Name: "work1", SuperPodID: 1},
						{Name: "work2", SuperPodID: 1},
					},
				},
				FaultTasks: []rescheduling.FaultTask{
					{NodeName: "work1", IsFaultTask: true,
						FaultTaskA5Field: rescheduling.FaultTaskA5Field{IsSatisfiedRackAffinity: true}},
					{NodeName: "work2", IsFaultTask: false,
						FaultTaskA5Field: rescheduling.FaultTaskA5Field{IsSatisfiedRackAffinity: false}},
				},
			},
			selectNodes: map[string][]plugin.SuperNode{},
			totalNodes: map[int32]superPod{
				1: {
					"work1": {name: "work1", superPodID: 1},
					"work2": {name: "work2", superPodID: 1},
				},
			},
			virtualIdArr: map[string]bool{
				"vsp1": false,
			},
			tp: &module910a5SuperPod{
				jobParams: jobParams{tpBlock: tpBlock2},
			},
		},
		want: map[string]struct{}{
			"vsp1": {},
		},
	}
}

func buildSelectNodeFromOriginSpBlockTest3() selectNodeFromOriginSpBlockTest {
	return selectNodeFromOriginSpBlockTest{
		name: "03-selectNodeFromOriginSpBlock input nil or empty virtualIdArr",
		args: selectNodeFromOriginSpBlockTestArgs{
			fJob:         &rescheduling.FaultJob{},
			selectNodes:  nil,
			totalNodes:   nil,
			virtualIdArr: nil,
			tp:           &module910a5SuperPod{},
		},
		want: nil,
	}
}

func buildSelectNodeFromOriginSpBlockTests() []selectNodeFromOriginSpBlockTest {
	return []selectNodeFromOriginSpBlockTest{
		buildSelectNodeFromOriginSpBlockTest1(),
		buildSelectNodeFromOriginSpBlockTest2(),
		buildSelectNodeFromOriginSpBlockTest3(),
	}
}

func TestSelectNodeFromOriginSpBlock(t *testing.T) {
	tests := buildSelectNodeFromOriginSpBlockTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := tt.args.tp.selectNodeFromOriginSpBlock(
				tt.args.fJob, tt.args.selectNodes, tt.args.totalNodes, tt.args.virtualIdArr)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.want, got)
			}
		})
	}
}

func createSuperPodWithRackIdMap(rackIdOrder []int32, nodeCounts map[int32]int) map[int32][]nodeBaseInfo {
	superPodWithRackId := make(map[int32][]nodeBaseInfo)
	for _, rackId := range rackIdOrder {
		nodes := make([]nodeBaseInfo, nodeCounts[rackId])
		superPodWithRackId[rackId] = nodes
	}
	return superPodWithRackId
}

func TestGetRestRackId(t *testing.T) {
	testCases := []struct {
		name        string
		rackIdOrder []int32
		nodeCounts  map[int32]int
		tpBlock     int
		want        int32
	}{
		{
			name:        "single rackId with enough nodes",
			rackIdOrder: []int32{1},
			nodeCounts:  map[int32]int{1: 2},
			tpBlock:     2,
			want:        1,
		},
		{
			name:        "single rackId with not enough nodes",
			rackIdOrder: []int32{1},
			nodeCounts:  map[int32]int{1: 1},
			tpBlock:     2,
			want:        UninitializedRestRackLenMapId,
		},
		{
			name:        "multiple rackIds with one having enough nodes",
			rackIdOrder: []int32{1, 2, 3},
			nodeCounts:  map[int32]int{1: 1, 2: 2, 3: 3},
			tpBlock:     2,
			want:        2,
		},
		{
			name:        "multiple rackIds with none having enough nodes",
			rackIdOrder: []int32{1, 2, 3},
			nodeCounts:  map[int32]int{1: 1, 2: 1, 3: 1},
			tpBlock:     2,
			want:        UninitializedRestRackLenMapId,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tp := &module910a5SuperPod{}
			tp.tpBlock = tt.tpBlock
			superPodWithRackId := createSuperPodWithRackIdMap(tt.rackIdOrder, tt.nodeCounts)
			restRackId := tp.getRestRackId(tt.rackIdOrder, superPodWithRackId)
			if restRackId != tt.want {
				t.Errorf("Test %s failed: expected result: %v, got: %v", tt.name, tt.want, restRackId)
			}
		})
	}
}
