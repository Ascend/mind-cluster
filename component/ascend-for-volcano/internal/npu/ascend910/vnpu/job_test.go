/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package vnpu

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/vnpu"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

type checkDyJobRequireTestCase struct {
	name string
	vT   util.NPUTask
	want bool
}

func buildCheckB41DyJobRequireTestCases() []checkDyJobRequireTestCase {
	return []checkDyJobRequireTestCase{
		{
			name: "01-checkB41DyJobRequire return true when ReqNPUNum is util.CoreNum5",
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum5},
			want: true,
		},
		{
			name: "02-checkB41DyJobRequire return true when ReqNPUNum is util.CoreNum10",
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum10},
			want: true,
		},
		{
			name: "03-checkB41DyJobRequire return true when ReqNPUNum is util.CoreNum20",
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum20},
			want: true,
		},
		{
			name: "04-checkB41DyJobRequire return false when ReqNPUNum is 7",
			vT:   util.NPUTask{ReqNPUNum: 7},
			want: false,
		},
	}
}

func TestCheckB41DyJobRequire(t *testing.T) {
	testCases := buildCheckB41DyJobRequireTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkB41DyJobRequire(tt.vT); got != tt.want {
				t.Errorf("checkB41DyJobRequire() = %v, want %v", got, tt.want)
			}
		})
	}
}

func buildCheckA3X20DyJobRequireTestCases() []checkDyJobRequireTestCase {
	return []checkDyJobRequireTestCase{
		{
			name: "01-checkA3X20DyJobRequire return true when ReqNPUNum is util.CoreNum5",
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum5},
			want: true,
		},
		{
			name: "02-checkA3X20DyJobRequire return true when ReqNPUNum is util.CoreNum10",
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum10},
			want: true,
		},
		{
			name: "03-checkA3X20DyJobRequire return true when ReqNPUNum is util.CoreNum20",
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum20},
			want: true,
		},
		{
			name: "04-checkA3X20DyJobRequire return true when ReqNPUNum is 40",
			vT:   util.NPUTask{ReqNPUNum: 40},
			want: true,
		},
		{
			name: "05-checkA3X20DyJobRequire return false when ReqNPUNum is 6",
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum6},
			want: false,
		},
		{
			name: "06-checkA3X20DyJobRequire return false when ReqNPUNum is 7",
			vT:   util.NPUTask{ReqNPUNum: 7},
			want: false,
		},
	}
}

func TestCheckA3X20DyJobRequire(t *testing.T) {
	testCases := buildCheckA3X20DyJobRequireTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkA3X20DyJobRequire(tt.vT); got != tt.want {
				t.Errorf("checkA3X20DyJobRequire() = %v, want %v", got, tt.want)
			}
		})
	}
}

func buildCheckA3X24DyJobRequireTestCases() []checkDyJobRequireTestCase {
	return []checkDyJobRequireTestCase{
		{
			name: "01-checkA3X24DyJobRequire return true when ReqNPUNum is util.CoreNum6",
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum6},
			want: true,
		},
		{
			name: "02-checkA3X24DyJobRequire return true when ReqNPUNum is util.CoreNum12",
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum12},
			want: true,
		},
		{
			name: "03-checkA3X24DyJobRequire return true when ReqNPUNum is util.CoreNum24",
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum24},
			want: true,
		},
		{
			name: "04-checkA3X24DyJobRequire return true when ReqNPUNum is 48",
			vT:   util.NPUTask{ReqNPUNum: 48},
			want: true,
		},
		{
			name: "05-checkA3X24DyJobRequire return false when ReqNPUNum is 5",
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum5},
			want: false,
		},
		{
			name: "06-checkA3X24DyJobRequire return false when ReqNPUNum is 7",
			vT:   util.NPUTask{ReqNPUNum: 7},
			want: false,
		},
	}
}

func TestCheckA3X24DyJobRequire(t *testing.T) {
	testCases := buildCheckA3X24DyJobRequireTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkA3X24DyJobRequire(tt.vT); got != tt.want {
				t.Errorf("checkA3X24DyJobRequire() = %v, want %v", got, tt.want)
			}
		})
	}
}

type checkDyVJobReqByTempTestCase struct {
	name string
	temp string
	vT   util.NPUTask
	want bool
}

func buildB41TestCases() []checkDyVJobReqByTempTestCase {
	return []checkDyVJobReqByTempTestCase{
		{
			name: "01-checkDyVJobReqByTemp return true when temp is ChipTypeB41 and ReqNPUNum is util.CoreNum5",
			temp: plugin.ChipTypeB41,
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum5},
			want: true,
		},
		{
			name: "02-checkDyVJobReqByTemp return false when temp is ChipTypeB41 and ReqNPUNum is 7",
			temp: plugin.ChipTypeB41,
			vT:   util.NPUTask{ReqNPUNum: 7},
			want: false,
		},
	}
}

func buildA3X20TestCases() []checkDyVJobReqByTempTestCase {
	return []checkDyVJobReqByTempTestCase{
		{
			name: "03-checkDyVJobReqByTemp return true when temp is ServerTypeA3X20 and ReqNPUNum is util.CoreNum5",
			temp: plugin.ServerTypeA3X20,
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum5},
			want: true,
		},
		{
			name: "04-checkDyVJobReqByTemp return true when temp is ServerTypeA3X20 and ReqNPUNum is util.CoreNum10",
			temp: plugin.ServerTypeA3X20,
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum10},
			want: true,
		},
		{
			name: "05-checkDyVJobReqByTemp return true when temp is ServerTypeA3X20 and ReqNPUNum is util.CoreNum20",
			temp: plugin.ServerTypeA3X20,
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum20},
			want: true,
		},
		{
			name: "06-checkDyVJobReqByTemp return false when temp is ServerTypeA3X20 and ReqNPUNum is util.CoreNum6",
			temp: plugin.ServerTypeA3X20,
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum6},
			want: false,
		},
		{
			name: "07-checkDyVJobReqByTemp return false when temp is ServerTypeA3X20 and ReqNPUNum is util.CoreNum12",
			temp: plugin.ServerTypeA3X20,
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum12},
			want: false,
		},
		{
			name: "08-checkDyVJobReqByTemp return false when temp is ServerTypeA3X20 and ReqNPUNum is 7",
			temp: plugin.ServerTypeA3X20,
			vT:   util.NPUTask{ReqNPUNum: 7},
			want: false,
		},
	}
}

func buildA3X24TestCases() []checkDyVJobReqByTempTestCase {
	return []checkDyVJobReqByTempTestCase{
		{
			name: "09-checkDyVJobReqByTemp return true when temp is ServerTypeA3X24 and ReqNPUNum is util.CoreNum6",
			temp: plugin.ServerTypeA3X24,
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum6},
			want: true,
		},
		{
			name: "10-checkDyVJobReqByTemp return true when temp is ServerTypeA3X24 and ReqNPUNum is util.CoreNum12",
			temp: plugin.ServerTypeA3X24,
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum12},
			want: true,
		},
		{
			name: "11-checkDyVJobReqByTemp return true when temp is ServerTypeA3X24 and ReqNPUNum is util.CoreNum24",
			temp: plugin.ServerTypeA3X24,
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum24},
			want: true,
		},
		{
			name: "12-checkDyVJobReqByTemp return false when temp is ServerTypeA3X24 and ReqNPUNum is util.CoreNum5",
			temp: plugin.ServerTypeA3X24,
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum5},
			want: false,
		},
		{
			name: "13-checkDyVJobReqByTemp return false when temp is ServerTypeA3X24 and ReqNPUNum is util.CoreNum10",
			temp: plugin.ServerTypeA3X24,
			vT:   util.NPUTask{ReqNPUNum: util.CoreNum10},
			want: false,
		},
	}
}

func buildUnknownTempTestCases() []checkDyVJobReqByTempTestCase {
	return []checkDyVJobReqByTempTestCase{
		{
			name: "14-checkDyVJobReqByTemp return false when temp is unknown",
			temp: "unknown",
			vT:   util.NPUTask{ReqNPUNum: 7},
			want: false,
		},
	}
}

func buildCheckDyVJobReqByTempTestCases() []checkDyVJobReqByTempTestCase {
	var testCases []checkDyVJobReqByTempTestCase
	testCases = append(testCases, buildB41TestCases()...)
	testCases = append(testCases, buildA3X20TestCases()...)
	testCases = append(testCases, buildA3X24TestCases()...)
	testCases = append(testCases, buildUnknownTempTestCases()...)
	return testCases
}

func TestCheckDyVJobReqByTemp(t *testing.T) {
	testCases := buildCheckDyVJobReqByTempTestCases()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tp := &virtual910NPU{
				vHandle: &vnpu.VirtualNPU{
					VT: vnpu.VTemplate{Temp: tt.temp},
				},
			}
			if got := tp.checkDyVJobReqByTemp(tt.vT); got != tt.want {
				t.Errorf("checkDyVJobReqByTemp() = %v, want %v", got, tt.want)
			}
		})
	}
}
