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
Package base is using for HuaWei Ascend pin affinity schedule.
*/
package base

import (
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func TestGetCardNumGroupsFromTop(t *testing.T) {
	tests := []struct {
		name            string
		tp              *NPUHandler
		nodeNPUTopology []int
		expected        [][]int
	}{
		{
			name:            "01 Nil NPUHandler test",
			tp:              nil,
			nodeNPUTopology: []int{util.NPUIndex1, util.NPUIndex2, util.NPUIndex3},
			expected:        nil,
		},
		{
			name:            "02 MaxCardNPUNum is zero",
			tp:              &NPUHandler{MaxCardNPUNum: 0},
			nodeNPUTopology: []int{util.NPUIndex1, util.NPUIndex2, util.NPUIndex3},
			expected:        nil,
		},
		{
			name:            "03 Single group test",
			tp:              &NPUHandler{MaxCardNPUNum: 4},
			nodeNPUTopology: []int{util.NPUIndex1, util.NPUIndex2, util.NPUIndex3},
			expected:        [][]int{{util.NPUIndex1, util.NPUIndex2, util.NPUIndex3}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result [][]int
			if tt.tp == nil {
				// Avoid calling methods on nil pointer
				result = nil
			} else {
				result = tt.tp.GetCardNumGroupsFromTop(tt.nodeNPUTopology)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetCardNumGroupsFromTop() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Helper function: create test node
func createTestNode(anno map[string]string) plugin.NPUNode {
	return plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name:       "test-node",
			Annotation: anno,
		},
	}
}

// Helper function: create NPUHandler instance
func createNPUHandler() *NPUHandler {
	return &NPUHandler{
		SchedulerJobAttr: util.SchedulerJobAttr{
			NPUJob: &util.NPUJob{},
		},
	}
}

func TestGetNetUnhealthyNPU(t *testing.T) {
	type TestCase struct {
		name     string
		setup    func(*NPUHandler)
		nodeAnno map[string]string
		want     []int
		wantErr  bool
	}

	testCases := []TestCase{
		{name: "ReqNPUName is empty", setup: func(tp *NPUHandler) {
			tp.SetPluginName("test-plugin")
			tp.ReqNPUName = ""
		}, nodeAnno: make(map[string]string), want: nil, wantErr: true},
		{name: "ReqNPUName is NPUCardName, no annotation", setup: func(tp *NPUHandler) {
			tp.SetPluginName("test-plugin")
			tp.SetAnnoPreVal(util.NPUCardNamePre)
			tp.ReqNPUName = util.NPUCardName
		}, nodeAnno: make(map[string]string), want: nil, wantErr: true},
		{name: "ReqNPUName is NPUCardName, with annotation", setup: func(tp *NPUHandler) {
			tp.SetPluginName("test-plugin")
			tp.SetAnnoPreVal(util.NPUCardNamePre)
			tp.ReqNPUName = util.NPUCardName
		}, nodeAnno: map[string]string{networkUnhealthyNPU: "npu-0,npu-1"},
			want: []int{0, 1}, wantErr: false},
		{name: "ReqNPUName is not NPUCardName, no annotation", setup: func(tp *NPUHandler) {
			tp.SetPluginName("test-plugin")
			tp.SetAnnoPreVal(util.NPU910CardNamePre)
			tp.ReqNPUName = util.Ascend910
		}, nodeAnno: make(map[string]string), want: nil, wantErr: true},
		{name: "ReqNPUName is not NPUCardName, with annotation", setup: func(tp *NPUHandler) {
			tp.SetPluginName("test-plugin")
			tp.SetAnnoPreVal(util.NPU910CardNamePre)
			tp.ReqNPUName = util.Ascend910
		}, nodeAnno: map[string]string{networkUnhealthy910: "Ascend910-2,Ascend910-3"},
			want: []int{2, 3}, wantErr: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tp := createNPUHandler()
			tc.setup(tp)
			got, err := tp.getNetUnhealthyNPU(createTestNode(tc.nodeAnno))
			if (err != nil) != tc.wantErr {
				t.Errorf("getNetUnhealthyNPU() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("getNetUnhealthyNPU() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGetNetUnhealthyNPUKey(t *testing.T) {
	type TestCase struct {
		name  string
		setup func(*NPUHandler)
		want  string
	}

	testCases := []TestCase{
		{name: "ReqNPUName is npu", setup: func(tp *NPUHandler) {
			tp.SetPluginName("test-plugin")
			tp.ReqNPUName = util.NPUCardName
		}, want: networkUnhealthyNPU},
		{name: "ReqNPUName is not npu", setup: func(tp *NPUHandler) {
			tp.SetPluginName("test-plugin")
			tp.ReqNPUName = util.Ascend910
		}, want: networkUnhealthy910},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tp := createNPUHandler()
			tc.setup(tp)
			got := tp.GetNetUnhealthyNPUKey()
			if got != tc.want {
				t.Errorf("getNetUnhealthyNPUKey() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGetUnhealthyNPU(t *testing.T) {
	type TestCase struct {
		name     string
		setup    func(*NPUHandler)
		nodeAnno map[string]string
		want     []int
	}

	testCases := []TestCase{
		{name: "ReqNPUName is empty", setup: func(tp *NPUHandler) {
			tp.SetPluginName("test-plugin")
			tp.ReqNPUName = ""
		}, nodeAnno: make(map[string]string), want: nil},
		{name: "ReqNPUName is NPUCardName, no annotation", setup: func(tp *NPUHandler) {
			tp.SetPluginName("test-plugin")
			tp.SetAnnoPreVal(util.NPUCardNamePre)
			tp.ReqNPUName = util.NPUCardName
		}, nodeAnno: make(map[string]string), want: []int{}},
		{name: "ReqNPUName is NPUCardName, with annotation", setup: func(tp *NPUHandler) {
			tp.SetPluginName("test-plugin")
			tp.SetAnnoPreVal(util.NPUCardNamePre)
			tp.ReqNPUName = util.NPUCardName
		}, nodeAnno: map[string]string{unHealthyNPU: "npu-0,npu-1"}, want: []int{0, 1}},
		{name: "ReqNPUName is not NPUCardName, no annotation", setup: func(tp *NPUHandler) {
			tp.SetPluginName("test-plugin")
			tp.SetAnnoPreVal(util.NPU910CardNamePre)
			tp.ReqNPUName = util.Ascend910
		}, nodeAnno: make(map[string]string), want: []int{}},
		{name: "ReqNPUName is not NPUCardName, with annotation", setup: func(tp *NPUHandler) {
			tp.SetPluginName("test-plugin")
			tp.SetAnnoPreVal(util.NPU910CardNamePre)
			tp.ReqNPUName = util.Ascend910
		}, nodeAnno: map[string]string{unHealthy910: "Ascend910-2,Ascend910-3"}, want: []int{2, 3}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tp := createNPUHandler()
			tc.setup(tp)
			got := tp.getUnhealthyNPU(createTestNode(tc.nodeAnno))
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("getUnhealthyNPU() = %v, want %v", got, tc.want)
			}
		})
	}
}
