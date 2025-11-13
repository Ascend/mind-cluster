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

// Package rescheduling is using for Huawei Ascend pin fault rescheduling
package rescheduling

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/k8s"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

func TestGetTaskHealthStateByNodeDpu(t *testing.T) {
	tests := []getTaskHealthStateByNodeDpuTestCase{
		buildGetTaskHealthStateByNodeDpuTestCase1(), buildGetTaskHealthStateByNodeDpuTestCase2(),
		buildGetTaskHealthStateByNodeDpuTestCase3(), buildGetTaskHealthStateByNodeDpuTestCase4(),
		buildGetTaskHealthStateByNodeDpuTestCase6(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the FaultNode and ReScheduler
			dpuList := make([]k8s.DpuListItem, 0, len(tt.args.dpuList))
			for _, d := range tt.args.dpuList {
				dpuList = append(dpuList, k8s.DpuListItem{Name: d.Name, Operstate: d.Operstate})
			}
			oneNodeDpuInfoCm := k8s.DpuCMInfo{
				BusType:      tt.args.busType,
				NpuToDpusMap: tt.args.npuToDpusMap,
				DpuList:      dpuList,
			}
			faultNode := &FaultNode{
				NodeName:  tt.args.nodeName,
				dpuCMInfo: oneNodeDpuInfoCm,
			}
			reScheduler := &ReScheduler{
				DealReSchedulerCache: &DealReSchedulerCache{
					FaultNodes: map[string]*FaultNode{
						tt.args.nodeName: faultNode,
					},
				},
			}
			gotFault := reScheduler.getTaskHealthStateByNodeDpu(tt.args.fTask)
			if gotFault != tt.wantFault {
				t.Errorf("getTaskHealthStateByNodeDpu() = (%v), want (%v)", gotFault, tt.wantFault)
			}
		})
	}
}

type getTaskHealthStateByNodeDpuTestCase struct {
	name string
	args struct {
		fTask        *FaultTask
		nodeName     string
		busType      string
		npuToDpusMap map[string][]string
		dpuList      []k8s.DpuListItem
	}
	wantFault bool
}

func buildGetTaskHealthStateByNodeDpuTestCase1() getTaskHealthStateByNodeDpuTestCase {
	return getTaskHealthStateByNodeDpuTestCase{
		name: "UBType - as long as one DPU status is up, the result should be healthy",
		args: struct {
			fTask        *FaultTask
			nodeName     string
			busType      string
			npuToDpusMap map[string][]string
			dpuList      []k8s.DpuListItem
		}{
			fTask: &FaultTask{
				TaskName: "task1",
				NodeName: "node1",
				UseCardName: []string{
					"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3", "Ascend910-4", "Ascend910-5",
					"Ascend910-6", "Ascend910-7",
				},
			},
			nodeName: "node1",
			busType:  util.UbType,
			npuToDpusMap: map[string][]string{
				"0": {"enps0", "enps2"},
				"1": {"enps0", "enps2"},
				"2": {"enps0", "enps2"},
				"3": {"enps0", "enps2"},
				"4": {"enps1", "enps3"},
				"5": {"enps1", "enps3"},
				"6": {"enps1", "enps3"},
				"7": {"enps1", "enps3"},
			},
			dpuList: []k8s.DpuListItem{
				{Name: "enps0", Operstate: util.ActiveStatus},
				{Name: "enps1", Operstate: util.ActiveStatus},
				{Name: "enps2", Operstate: util.ActiveStatus},
				{Name: "enps3", Operstate: "down"},
			},
		},
		wantFault: false,
	}
}

func buildGetTaskHealthStateByNodeDpuTestCase2() getTaskHealthStateByNodeDpuTestCase {
	return getTaskHealthStateByNodeDpuTestCase{
		name: "UBType - both DPUs faulty",
		args: struct {
			fTask        *FaultTask
			nodeName     string
			busType      string
			npuToDpusMap map[string][]string
			dpuList      []k8s.DpuListItem
		}{
			fTask: &FaultTask{
				TaskName:    "task1",
				NodeName:    "node1",
				UseCardName: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3"},
			},
			nodeName: "node1",
			busType:  util.UbType,
			npuToDpusMap: map[string][]string{
				"0": {"enps0", "enps2"},
				"1": {"enps0", "enps2"},
				"2": {"enps0", "enps2"},
				"3": {"enps0", "enps2"},
			},
			dpuList: []k8s.DpuListItem{
				{Name: "enps0", Operstate: "down"},
				{Name: "enps2", Operstate: "down"},
			},
		},
		wantFault: true,
	}
}

func buildGetTaskHealthStateByNodeDpuTestCase3() getTaskHealthStateByNodeDpuTestCase {
	return getTaskHealthStateByNodeDpuTestCase{
		name: "PcieType - DPU healthy",
		args: struct {
			fTask        *FaultTask
			nodeName     string
			busType      string
			npuToDpusMap map[string][]string
			dpuList      []k8s.DpuListItem
		}{
			fTask: &FaultTask{
				TaskName:    "task1",
				NodeName:    "node1",
				UseCardName: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3"},
			},
			nodeName: "node1",
			busType:  util.PcieType,
			npuToDpusMap: map[string][]string{
				"0": {"eth1"},
				"1": {"eth2"},
				"2": {"eth3"},
				"3": {"eth4"},
			},
			dpuList: []k8s.DpuListItem{
				{Name: "eth1", Operstate: util.ActiveStatus},
				{Name: "eth2", Operstate: util.ActiveStatus},
				{Name: "eth3", Operstate: util.ActiveStatus},
				{Name: "eth4", Operstate: util.ActiveStatus},
			},
		},
		wantFault: false,
	}
}

func buildGetTaskHealthStateByNodeDpuTestCase4() getTaskHealthStateByNodeDpuTestCase {
	return getTaskHealthStateByNodeDpuTestCase{
		name: "PcieType - DPU faulty",
		args: struct {
			fTask        *FaultTask
			nodeName     string
			busType      string
			npuToDpusMap map[string][]string
			dpuList      []k8s.DpuListItem
		}{
			fTask: &FaultTask{
				TaskName:    "task1",
				NodeName:    "node1",
				UseCardName: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3"},
			},
			nodeName: "node1",
			busType:  util.PcieType,
			npuToDpusMap: map[string][]string{
				"0": {"eth1"},
				"1": {"eth2"},
				"2": {"eth3"},
				"3": {"eth4"},
			},
			dpuList: []k8s.DpuListItem{
				{Name: "eth1", Operstate: "down"},
				{Name: "eth2", Operstate: util.ActiveStatus},
				{Name: "eth3", Operstate: util.ActiveStatus},
				{Name: "eth4", Operstate: util.ActiveStatus},
			},
		},
		wantFault: true,
	}
}

func buildGetTaskHealthStateByNodeDpuTestCase6() getTaskHealthStateByNodeDpuTestCase {
	return getTaskHealthStateByNodeDpuTestCase{
		name: "No DPU info, should be healthy",
		args: struct {
			fTask        *FaultTask
			nodeName     string
			busType      string
			npuToDpusMap map[string][]string
			dpuList      []k8s.DpuListItem
		}{
			fTask: &FaultTask{
				TaskName:    "task1",
				NodeName:    "node1",
				UseCardName: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3"},
			},
			nodeName:     "node1",
			busType:      util.PcieType,
			npuToDpusMap: map[string][]string{},
			dpuList:      []k8s.DpuListItem{},
		},
		wantFault: false,
	}
}

func TestIsNpuBeUsed(t *testing.T) {
	tests := append(
		buildIsNpuBeUsedTestCasesPart1(),
		buildIsNpuBeUsedTestCasesPart2()...,
	)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNpuBeUsed(tt.args.strs, tt.args.target); got != tt.want {
				t.Errorf("isNpuBeUsed() = %v, want %v", got, tt.want)
			}
		})
	}
}

type isNpuBeUsedArgs struct {
	strs   []string
	target string
}

type isNpuBeUsedTestCase struct {
	name string
	args isNpuBeUsedArgs
	want bool
}

func buildIsNpuBeUsedTestCasesPart1() []isNpuBeUsedTestCase {
	return []isNpuBeUsedTestCase{
		{
			name: "01-should return true when target matches cardIdStr",
			args: isNpuBeUsedArgs{
				strs:   []string{"Ascend910-0", "Ascend910-1"},
				target: "1",
			},
			want: true,
		},
		{
			name: "02-should return false when target does not match any cardIdStr",
			args: isNpuBeUsedArgs{
				strs:   []string{"Ascend910-0", "Ascend910-1"},
				target: "3",
			},
			want: false,
		},
		{
			name: "03-should return false when input slice is empty",
			args: isNpuBeUsedArgs{
				strs:   []string{},
				target: "0",
			},
			want: false,
		},
		{
			name: "04-should return false when string does not contain '-'",
			args: isNpuBeUsedArgs{
				strs:   []string{"Ascend910"},
				target: "0",
			},
			want: false,
		},
	}
}

func buildIsNpuBeUsedTestCasesPart2() []isNpuBeUsedTestCase {
	return []isNpuBeUsedTestCase{
		{
			name: "05-should return false when cardId is not an integer",
			args: isNpuBeUsedArgs{
				strs:   []string{"Ascend910-abc"},
				target: "0",
			},
			want: false,
		},
		{
			name: "06-should handle cardIdStr with modulo NodeNum8",
			args: isNpuBeUsedArgs{
				strs:   []string{"Ascend910-9"}, // 9 % 8 = 1
				target: "1",
			},
			want: true,
		},
		{
			name: "07-should return false if modulo does not match target",
			args: isNpuBeUsedArgs{
				strs:   []string{"Ascend910-10"}, // 10 % 8 = 2
				target: "3",
			},
			want: false,
		},
	}
}
