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

package npu

import "testing"

func TestJudgeSchedulePolicyIsNodeAffinityRequired(t *testing.T) {
	tests := []struct {
		name   string
		policy string
		want   bool
	}{
		{"ra and sp affinity", "chip8-node8-ra64-sp", true},
		{"sp affinity only", "chip8-node8-sp", true},
		{"ra affinity only", "chip8-node8-ra64", true},
		{"no affinity", "chip8-node8", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := judgeSchedulePolicyIsNodeAffinityRequired(tt.policy); got != tt.want {
				t.Fatalf("judgeSchedulePolicyIsNodeAffinityRequired(%q) = %v, want %v", tt.policy, got, tt.want)
			}
		})
	}
}

func TestStripAffinitySegments(t *testing.T) {
	tests := []struct {
		name   string
		policy string
		want   string
	}{
		{"ra and sp", "chip8-node8-ra64-sp", "chip8-node8"},
		{"ra only", "chip8-node8-ra64", "chip8-node8"},
		{"sp only", "chip8-node8-sp", "chip8-node8"},
		{"no affinity", "chip8-node8", "chip8-node8"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripAffinitySegments(tt.policy); got != tt.want {
				t.Fatalf("stripAffinitySegments(%q) = %q, want %q", tt.policy, got, tt.want)
			}
		})
	}
}
