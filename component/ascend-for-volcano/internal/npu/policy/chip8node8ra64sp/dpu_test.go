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

// Package chip8node8ra64sp is using for Huawei Ascend pin affinity schedule
package chip8node8ra64sp

import (
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func TestFilterDpuFault(t *testing.T) {
	tests := []filterDpuFaultTestCase{
		buildFilterDpuFaultTestCase1(), buildFilterDpuFaultTestCase2(),
	}
	tp := New(SuperPodx8SchedulerName)
	// Set up NPUJob with ReqNPUName to avoid nil pointer
	tp.SetSchedulerAttr(util.SchedulerJobAttr{
		NPUJob: &util.NPUJob{
			ReqNPUName: util.NPUCardName,
		},
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tp.filterDpuFault(tt.npuCardIdList, tt.node)
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
		name:          "No dpuUnhealthyNPU info, returns input list",
		npuCardIdList: []int{0, 1, 2, 3, 4, 5, 6, 7},
		node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name: "work1",
			},
		},
		want: []int{0, 1, 2, 3, 4, 5, 6, 7},
	}
}

func buildFilterDpuFaultTestCase2() filterDpuFaultTestCase {
	return filterDpuFaultTestCase{
		name:          "If dpuUnhealthyNPU is not empty, it will be filtered out.",
		npuCardIdList: []int{0, 1, 2, 3, 4, 5, 6, 7},
		node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name: "work2",
				Annotation: map[string]string{
					dpuUnhealthyNPU: "npu-2,npu-3,npu-4,npu-5,npu-6,npu-7",
				},
			},
		},
		want: []int{0, 1},
	}
}
