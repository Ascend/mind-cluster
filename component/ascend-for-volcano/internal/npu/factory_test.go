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
Package npu is using for HuaWei Ascend pin affinity schedule.
*/
package npu

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// InitPolicyHandlerTest
type InitPolicyHandlerTest struct {
	name        string
	attr        util.SchedulerJobAttr
	env         plugin.ScheduleEnv
	wantHandler plugin.SchedulerPluginNeed
	wantBool    bool
}

// buildInitPolicyHandlerTestCases
func buildInitPolicyHandlerTestCases() []InitPolicyHandlerTest {
	return []InitPolicyHandlerTest{
		{
			name: "01 NPU910CardName - return handler",
			attr: util.SchedulerJobAttr{ComJob: util.ComJob{Label: map[string]string{}},
				NPUJob: &util.NPUJob{ReqNPUName: util.NPU910CardName}},
			env:         plugin.ScheduleEnv{},
			wantHandler: nil,
			wantBool:    true,
		},
		{
			name: "02 NPU310CardName - return handler",
			attr: util.SchedulerJobAttr{ComJob: util.ComJob{Label: map[string]string{}},
				NPUJob: &util.NPUJob{ReqNPUName: util.NPU310CardName}},
			env:         plugin.ScheduleEnv{},
			wantHandler: nil,
			wantBool:    true,
		},
		{
			name: "03 NPU310PCardName - return handler",
			attr: util.SchedulerJobAttr{ComJob: util.ComJob{Label: map[string]string{}},
				NPUJob: &util.NPUJob{ReqNPUName: util.NPU310PCardName}},
			env:         plugin.ScheduleEnv{},
			wantHandler: nil,
			wantBool:    true,
		},
		{
			name: "04 unknown plugin - return nil and false",
			attr: util.SchedulerJobAttr{
				ComJob: util.ComJob{Label: map[string]string{}},
				NPUJob: &util.NPUJob{ReqNPUName: "unknown-plugin-name"}},
			env:         plugin.ScheduleEnv{},
			wantHandler: nil,
			wantBool:    false,
		},
	}
}

func TestInitPolicyHandler(t *testing.T) {
	initCard910Factory()
	initCard310Factory()
	initCard310PFactory()
	tests := buildInitPolicyHandlerTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHandler, gotBool := InitPolicyHandler(tt.attr, tt.env)
			if gotBool != tt.wantBool {
				t.Errorf("InitPolicyHandler() gotBool = %v, want %v", gotBool, tt.wantBool)
			}
			if tt.wantBool && gotHandler == nil {
				t.Errorf("InitPolicyHandler() expected non-nil handler, got nil")
			}
			if !tt.wantBool && gotHandler != nil {
				t.Errorf("InitPolicyHandler() expected nil handler, got %v", gotHandler)
			}
		})
	}
}
