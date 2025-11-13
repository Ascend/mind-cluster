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

// Package superpod is using for Huawei Ascend pin affinity schedule
package superpod

import (
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/k8s"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func TestFilterDpuFault(t *testing.T) {
	tests := []filterDpuFaultTestCase{buildFilterDpuFaultTestCase1(), buildFilterDpuFaultTestCase2(),
		buildFilterDpuFaultTestCase3()}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterDpuFault(tt.npuCardIdList, tt.node)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterDpuFault() = %v, want %v", got, tt.want)
			}
		})
	}
}

type filterDpuFaultTestCase struct {
	name          string
	npuCardIdList []int
	node          plugin.NPUNode
	want          []int
	wantErr       bool
}

func buildFilterDpuFaultTestCase1() filterDpuFaultTestCase {
	return filterDpuFaultTestCase{
		name:          "No DPU info, returns input list",
		npuCardIdList: []int{0, 1, 2, 3, 4, 5, 6, 7},
		node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name: "work1",
				DpuInfo: k8s.DpuCMInfo{
					DpuList: []k8s.DpuListItem{},
				},
			},
		},
		want: []int{0, 1, 2, 3, 4, 5, 6, 7},
	}
}
func buildFilterDpuFaultTestCase2() filterDpuFaultTestCase {
	return filterDpuFaultTestCase{
		name:          "[UB type] If one of the two DPUs is in the up state, it will be returned. If neither is up, it will be filtered out.",
		npuCardIdList: []int{0, 1, 2, 3, 4, 5, 6, 7},
		node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name: "work2",
				DpuInfo: k8s.DpuCMInfo{
					BusType: util.UbType,
					DpuList: []k8s.DpuListItem{
						{Name: "enps0", Operstate: util.ActiveStatus},
						{Name: "enps1", Operstate: "down"},
						{Name: "enps2", Operstate: "down"},
						{Name: "enps3", Operstate: "down"},
					},
					NpuToDpusMap: map[string][]string{
						"0": {"enps0", "enps2"},
						"1": {"enps0", "enps2"},
						"4": {"enps1", "enps3"},
						"5": {"enps1", "enps3"},
					},
				},
			},
		},
		want: []int{0, 1},
	}
}
func buildFilterDpuFaultTestCase3() filterDpuFaultTestCase {
	return filterDpuFaultTestCase{
		name:          "[PCIe type] The dpu status associated with the returned npu must be up.",
		npuCardIdList: []int{0, 1, 2, 3, 4, 5, 6, 7},
		node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name: "work2",
				DpuInfo: k8s.DpuCMInfo{
					BusType: util.PcieType,
					DpuList: []k8s.DpuListItem{
						{Name: "eth1", Operstate: util.ActiveStatus},
						{Name: "eth2", Operstate: util.ActiveStatus},
						{Name: "eth3", Operstate: "down"},
						{Name: "eth4", Operstate: "down"},
					},
					NpuToDpusMap: map[string][]string{
						"0": {"eth1"},
						"1": {"eth2"},
						"2": {"eth3"},
						"3": {"eth4"},
					},
				},
			},
		},
		want: []int{0, 1},
	}
}
