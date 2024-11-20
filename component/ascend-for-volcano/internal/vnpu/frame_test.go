/*
Copyright(C)2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package vnpu is using for HuaWei Ascend pin vnpu allocation.
*/
package vnpu

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

type virtualNPUTestFields struct {
	DynamicByConf bool
	VT            VTemplate
	StaticVNPU    StaticVNPU
	DynamicVNPU   DynamicVNPU
}

type getTaskResourceArgs struct {
	task *api.TaskInfo
	node plugin.NPUNode
}

type getTaskResourceTest struct {
	name    string
	fields  virtualNPUTestFields
	args    getTaskResourceArgs
	want    util.VResource
	wantErr error
}

var testVT = map[string]util.VResource{
	plugin.VNPUTempVir01:        {Aicore: 1, Aicpu: 1, DVPP: plugin.AscendDVPPEnabledNull},
	plugin.VNPUTempVir02:        {Aicore: util.NPUIndex2, Aicpu: util.NPUIndex2, DVPP: plugin.AscendDVPPEnabledNull},
	plugin.VNPUTempVir02C1:      {Aicore: util.NPUIndex2, Aicpu: 1, DVPP: plugin.AscendDVPPEnabledNull},
	plugin.VNPUTempVir04:        {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex4, DVPP: plugin.AscendDVPPEnabledNull},
	plugin.VNPUTempVir04C3:      {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex3, DVPP: plugin.AscendDVPPEnabledNull},
	plugin.VNPUTempVir04C3NDVPP: {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex3, DVPP: plugin.AscendDVPPEnabledOff},
	plugin.VNPUTempVir04C4cDVPP: {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex4, DVPP: plugin.AscendDVPPEnabledOn},
}

func buildGetTaskResourceTestCase01() getTaskResourceTest {
	tempTask := test.FakeNormalTestTask("pod1", "node1", "vcjob")
	return getTaskResourceTest{
		name:    "01-task no ai-core test.",
		fields:  virtualNPUTestFields{},
		args:    getTaskResourceArgs{task: tempTask},
		want:    util.VResource{},
		wantErr: fmt.Errorf("task %s AscendNPUCore read failed", tempTask.Name),
	}
}

func buildGetTaskResourceTestCase02() getTaskResourceTest {
	tempTask := test.FakeVNPUTestTask("pod1", "node1", "vcjob", util.NPUIndex8)
	tempNode := plugin.NPUNode{CommonNode: plugin.CommonNode{Name: "testNode2", Idle: map[v1.ResourceName]float64{
		util.NPU910CardName: util.NPUIndex8 * util.NPUHexKilo}}}
	return getTaskResourceTest{
		name:    "02-node no ai-core test.",
		fields:  virtualNPUTestFields{},
		args:    getTaskResourceArgs{task: tempTask, node: tempNode},
		want:    util.VResource{},
		wantErr: fmt.Errorf("%s not inital for Aicore is 0", tempNode.Name),
	}
}

func buildGetTaskResourceTestCase03() getTaskResourceTest {
	tempTask := test.FakeVNPUTestTask("pod1", "node1", "vcjob", util.NPUIndex8)
	tempNode := plugin.NPUNode{CommonNode: plugin.CommonNode{Name: "testNode2"},
		VNode: plugin.VNode{ServerType: "Ascend310P-8", TotalRes: util.VResource{Aicore: util.NPUIndex8}}}
	return getTaskResourceTest{
		name:    "03-task whole card test.",
		fields:  virtualNPUTestFields{},
		args:    getTaskResourceArgs{task: tempTask, node: tempNode},
		want:    util.VResource{Aicore: util.NPUIndex8, Aicpu: 0, DVPP: plugin.AscendDVPPEnabledNull},
		wantErr: nil,
	}
}

func buildGetTaskResourceTestCase04() getTaskResourceTest {
	tempTask := test.FakeVNPUTestTask("pod1", "node1", "vcjob", util.NPUIndex4)
	test.AddTestTaskLabel(tempTask, plugin.AscendVNPUDVPP, "haha")
	tempNode := plugin.NPUNode{CommonNode: plugin.CommonNode{Name: "testNode2"},
		VNode: plugin.VNode{ServerType: "Ascend310P-8", TotalRes: util.VResource{Aicore: util.NPUIndex8}}}
	return getTaskResourceTest{
		name:    "04-illegal task dvpp test.",
		fields:  virtualNPUTestFields{},
		args:    getTaskResourceArgs{task: tempTask, node: tempNode},
		want:    util.VResource{},
		wantErr: errors.New("err dvpp value:haha"),
	}
}

func buildGetTaskResourceTestCase05() getTaskResourceTest {
	tempTask := test.FakeVNPUTestTask("pod1", "node1", "vcjob", 1)
	test.AddTestTaskLabel(tempTask, plugin.AscendVNPUDVPP, plugin.AscendDVPPEnabledNull)
	tempNode := plugin.NPUNode{CommonNode: plugin.CommonNode{Name: "testNode2"},
		VNode: plugin.VNode{ServerType: "Ascend310P-8", TotalRes: util.VResource{Aicore: util.NPUIndex16}}}
	return getTaskResourceTest{
		name:    "05-task coreNum 01 test.",
		fields:  virtualNPUTestFields{VT: VTemplate{Data: testVT}},
		args:    getTaskResourceArgs{task: tempTask, node: tempNode},
		want:    util.VResource{Aicore: 1, Aicpu: 1, DVPP: "null"},
		wantErr: nil,
	}
}

func buildGetTaskResourceTestCase06() getTaskResourceTest {
	tempTask := test.FakeVNPUTestTask("pod1", "node1", "vcjob", util.NPUIndex2)
	test.AddTestTaskLabel(tempTask, plugin.AscendVNPUDVPP, plugin.AscendDVPPEnabledNull)
	tempNode := plugin.NPUNode{CommonNode: plugin.CommonNode{Name: "testNode2"},
		VNode: plugin.VNode{ServerType: "Ascend310P-8", TotalRes: util.VResource{Aicore: util.NPUIndex16}}}
	return getTaskResourceTest{
		name:    "06-task coreNum 02 test.",
		fields:  virtualNPUTestFields{VT: VTemplate{Data: testVT}},
		args:    getTaskResourceArgs{task: tempTask, node: tempNode},
		want:    util.VResource{Aicore: util.NPUIndex2, Aicpu: 1, DVPP: "null"},
		wantErr: nil,
	}
}

func buildGetTaskResourceTestCase07() getTaskResourceTest {
	tempTask := test.FakeVNPUTestTask("pod1", "node1", "vcjob", util.NPUIndex3)
	test.AddTestTaskLabel(tempTask, plugin.AscendVNPUDVPP, plugin.AscendDVPPEnabledNull)
	tempNode := plugin.NPUNode{CommonNode: plugin.CommonNode{Name: "testNode2"},
		VNode: plugin.VNode{ServerType: "Ascend310P-8", TotalRes: util.VResource{Aicore: util.NPUIndex16}}}
	return getTaskResourceTest{
		name:    "07-task coreNum 03 test.",
		fields:  virtualNPUTestFields{VT: VTemplate{Data: testVT}},
		args:    getTaskResourceArgs{task: tempTask, node: tempNode},
		want:    util.VResource{Aicore: 0, Aicpu: 0},
		wantErr: nil,
	}
}

func buildGetTaskResourceTestCase08() getTaskResourceTest {
	tempTask := test.FakeVNPUTestTask("pod1", "node1", "vcjob", util.NPUIndex4)
	test.AddTestTaskLabel(tempTask, plugin.AscendVNPUDVPP, plugin.AscendDVPPEnabledOn)
	tempNode := plugin.NPUNode{CommonNode: plugin.CommonNode{Name: "testNode2"},
		VNode: plugin.VNode{ServerType: "Ascend310P-8", TotalRes: util.VResource{Aicore: util.NPUIndex16}}}
	return getTaskResourceTest{
		name:    "08-task coreNum 04,dvpp on test.",
		fields:  virtualNPUTestFields{VT: VTemplate{Data: testVT}},
		args:    getTaskResourceArgs{task: tempTask, node: tempNode},
		want:    util.VResource{Aicore: util.NPUIndex4, Aicpu: util.NPUIndex4, DVPP: plugin.AscendDVPPEnabledOn},
		wantErr: nil,
	}
}

func buildGetTaskResourceTestCase09() getTaskResourceTest {
	tempTask := test.FakeVNPUTestTask("pod1", "node1", "vcjob", util.NPUIndex4)
	test.AddTestTaskLabel(tempTask, plugin.AscendVNPUDVPP, plugin.AscendDVPPEnabledOff)
	tempNode := plugin.NPUNode{CommonNode: plugin.CommonNode{Name: "testNode2"},
		VNode: plugin.VNode{ServerType: "Ascend310P-8", TotalRes: util.VResource{Aicore: util.NPUIndex16}}}
	return getTaskResourceTest{
		name:    "09-task coreNum 04,dvpp off test.",
		fields:  virtualNPUTestFields{VT: VTemplate{Data: testVT}},
		args:    getTaskResourceArgs{task: tempTask, node: tempNode},
		want:    util.VResource{Aicore: util.NPUIndex4, Aicpu: util.NPUIndex3, DVPP: plugin.AscendDVPPEnabledOff},
		wantErr: nil,
	}
}

func buildGetTaskResourceTestCase10() getTaskResourceTest {
	tempTask := test.FakeVNPUTestTask("pod1", "node1", "vcjob", util.NPUIndex4)
	test.AddTestTaskLabel(tempTask, plugin.AscendVNPUDVPP, plugin.AscendDVPPEnabledOn)
	test.AddTestTaskLabel(tempTask, plugin.AscendVNPULevel, plugin.AscendVNPULevelHigh)
	tempNode := plugin.NPUNode{CommonNode: plugin.CommonNode{Name: "testNode2"},
		VNode: plugin.VNode{ServerType: "Ascend310P-8", TotalRes: util.VResource{Aicore: util.NPUIndex16}}}
	return getTaskResourceTest{
		name:    "10-task coreNum 04,level high, dvpp on test.",
		fields:  virtualNPUTestFields{VT: VTemplate{Data: testVT}},
		args:    getTaskResourceArgs{task: tempTask, node: tempNode},
		want:    util.VResource{Aicore: util.NPUIndex4, Aicpu: util.NPUIndex4, DVPP: plugin.AscendDVPPEnabledOn},
		wantErr: nil,
	}
}

func buildGetTaskResourceTestCase11() getTaskResourceTest {
	tempTask := test.FakeVNPUTestTask("pod1", "node1", "vcjob", util.NPUIndex4)
	test.AddTestTaskLabel(tempTask, plugin.AscendVNPUDVPP, plugin.AscendDVPPEnabledNull)
	test.AddTestTaskLabel(tempTask, plugin.AscendVNPULevel, plugin.AscendVNPULevelHigh)
	tempNode := plugin.NPUNode{CommonNode: plugin.CommonNode{Name: "testNode2"},
		VNode: plugin.VNode{ServerType: "Ascend310P-8", TotalRes: util.VResource{Aicore: util.NPUIndex16}}}
	return getTaskResourceTest{
		name:    "11-task coreNum 04,level high, dvpp null test.",
		fields:  virtualNPUTestFields{VT: VTemplate{Data: testVT}},
		args:    getTaskResourceArgs{task: tempTask, node: tempNode},
		want:    util.VResource{Aicore: util.NPUIndex4, Aicpu: util.NPUIndex4, DVPP: plugin.AscendDVPPEnabledNull},
		wantErr: nil,
	}
}

func buildGetTaskResourceTestCase12() getTaskResourceTest {
	tempTask := test.FakeVNPUTestTask("pod1", "node1", "vcjob", util.NPUIndex4)
	test.AddTestTaskLabel(tempTask, plugin.AscendVNPUDVPP, plugin.AscendDVPPEnabledNull)
	test.AddTestTaskLabel(tempTask, plugin.AscendVNPULevel, plugin.AscendVNPULevelLow)
	tempNode := plugin.NPUNode{CommonNode: plugin.CommonNode{Name: "testNode2"},
		VNode: plugin.VNode{ServerType: "Ascend310P-8", TotalRes: util.VResource{Aicore: util.NPUIndex16}}}
	return getTaskResourceTest{
		name:    "12-task coreNum 04,level low, dvpp null test.",
		fields:  virtualNPUTestFields{VT: VTemplate{Data: testVT}},
		args:    getTaskResourceArgs{task: tempTask, node: tempNode},
		want:    util.VResource{Aicore: util.NPUIndex4, Aicpu: util.NPUIndex3, DVPP: plugin.AscendDVPPEnabledNull},
		wantErr: nil,
	}
}

func buildGetTaskResourceTestCases() []getTaskResourceTest {
	testCases := []getTaskResourceTest{
		buildGetTaskResourceTestCase01(),
		buildGetTaskResourceTestCase02(),
		buildGetTaskResourceTestCase03(),
		buildGetTaskResourceTestCase04(),
		buildGetTaskResourceTestCase05(),
		buildGetTaskResourceTestCase06(),
		buildGetTaskResourceTestCase07(),
		buildGetTaskResourceTestCase08(),
		buildGetTaskResourceTestCase09(),
		buildGetTaskResourceTestCase10(),
		buildGetTaskResourceTestCase11(),
		buildGetTaskResourceTestCase12(),
	}
	return testCases
}

// TestGetTaskResource test GetTaskResource of TestVirtualNPU.
func TestGetTaskResource(t *testing.T) {
	tests := buildGetTaskResourceTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tp = &VirtualNPU{
				StaticByConf: tt.fields.DynamicByConf,
				VT:           tt.fields.VT,
				StaticVNPU:   tt.fields.StaticVNPU,
				DynamicVNPU:  tt.fields.DynamicVNPU,
			}
			got, err := tp.GetTaskResource(tt.args.task, tt.args.node)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetTaskResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTaskResource() got = %v, want %v", got, tt.want)
			}
		})
	}
}
