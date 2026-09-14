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
	"fmt"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/affinity/chip"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/policy/chip8node8sp"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910a3/superpod"
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

func TestInit910CardPolicyHandler(t *testing.T) {
	configs := []string{
		util.Chip1Node2,
		util.Chip4Node4,
		util.Chip8Node8,
		util.Chip8Node16,
		util.Chip2Node16,
		util.Chip2Node16Sp,
		util.Chip8Node16Sp,
		util.Chip4Node8,
		util.Chip4Node16,
		util.Chip1Node8,
		util.Chip1Node16,
	}

	for _, config := range configs {
		name := fmt.Sprintf("When schedule policy is %s then handleName is %s",
			config, policy910HandlerMap[config])
		t.Run(name, func(t *testing.T) {
			attr := util.SchedulerJobAttr{
				NPUJob: &util.NPUJob{
					ReqNPUName: util.NPU910CardName,
				},
				ComJob: util.ComJob{
					Annotation: map[string]string{
						util.SchedulePolicyAnnoKey: config,
					},
				},
			}
			handlerName := get910CardHandlerName(attr)
			if handlerName != policy910HandlerMap[config] {
				t.Errorf("Expect handler name to be %s, got %s", policy910HandlerMap[config], handlerName)
			}
		})
	}
}

type getA3AcceleratorTypeTest struct {
	selector    string
	wantHandler string
}

func TestGetA3AcceleratorType(t *testing.T) {
	configs := []getA3AcceleratorTypeTest{
		{
			selector:    util.Module910A3x8SuperPodAcceleratorType,
			wantHandler: superpod.A3x8SchedulerName,
		},
		{
			selector:    util.Module910A3x16SuperPodAcceleratorType,
			wantHandler: superpod.A3x16SchedulerName,
		},
	}
	for _, config := range configs {
		name := fmt.Sprintf("When node selector is %s then handleName is %s",
			config.selector, config.wantHandler)
		t.Run(name, func(t *testing.T) {
			attr := util.SchedulerJobAttr{
				ComJob: util.ComJob{
					Selector: map[string]string{
						util.AcceleratorTypeKeyDeprecated: config.selector,
					},
				},
			}
			handlerName := getA3HandlerNameWithSpBlock(attr)
			if handlerName != config.wantHandler {
				t.Errorf("Expect handler name to be %s, got %s", config.wantHandler, handlerName)
			}
		})
	}
}

func TestChipAffinityRegistered(t *testing.T) {
	factory := card910Factory[chip.PolicyName]
	if factory == nil {
		t.Fatalf("chip.PolicyName(%q) not registered in card910Factory", chip.PolicyName)
	}
	if h := factory(); h == nil {
		t.Fatal("chip factory returned nil handler")
	}
}

func TestGet910CardHandlerNameRoutesToChipAffinity(t *testing.T) {
	attr := util.SchedulerJobAttr{
		NPUJob: &util.NPUJob{ReqNPUName: util.NPU910CardName},
		ComJob: util.ComJob{Annotation: map[string]string{}, Selector: map[string]string{}},
	}
	if got := get910CardHandlerName(attr); got != chip.PolicyName {
		t.Errorf("default 910 routing = %q, want %q", got, chip.PolicyName)
	}
}

// TestGet910CardHandlerNameSingleTaskRoutesToChip verifies that a single-node
// (NPUTaskNum==1) job under the 8p-16-sp schedule policy is routed to
// chip-affinity, while single-task 8p-8-sp and all multi-task jobs keep their
// super-pod handler.
func TestGet910CardHandlerNameSingleTaskRoutesToChip(t *testing.T) {
	tests := []struct {
		name        string
		policy      string
		npuTaskNum  int
		wantHandler string
	}{
		{
			name:        "single task 8p-8-sp keeps super-pod handler",
			policy:      util.Chip8Node8Sp,
			npuTaskNum:  1,
			wantHandler: chip8node8sp.SchedulePolicy8Px8Sp,
		},
		{
			name:        "single task 8p-16-sp routes to chip",
			policy:      util.Chip8Node16Sp,
			npuTaskNum:  1,
			wantHandler: chip.PolicyName,
		},
		{
			name:        "multi task 8p-8-sp keeps super-pod handler",
			policy:      util.Chip8Node8Sp,
			npuTaskNum:  8,
			wantHandler: chip8node8sp.SchedulePolicy8Px8Sp,
		},
		{
			name:        "multi task 8p-16-sp keeps super-pod handler",
			policy:      util.Chip8Node16Sp,
			npuTaskNum:  16,
			wantHandler: chip8node8sp.SchedulePolicy8Px16Sp,
		},
		{
			name:        "single task non-super-pod policy keeps own handler",
			policy:      util.Chip8Node16,
			npuTaskNum:  1,
			wantHandler: policy910HandlerMap[util.Chip8Node16],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := util.SchedulerJobAttr{
				NPUJob: &util.NPUJob{
					ReqNPUName: util.NPU910CardName,
					NPUTaskNum: tt.npuTaskNum,
				},
				ComJob: util.ComJob{
					Annotation: map[string]string{
						util.SchedulePolicyAnnoKey: tt.policy,
					},
				},
			}
			got := get910CardHandlerName(attr)
			if got != tt.wantHandler {
				t.Errorf("handler = %q, want %q", got, tt.wantHandler)
			}
		})
	}
}
