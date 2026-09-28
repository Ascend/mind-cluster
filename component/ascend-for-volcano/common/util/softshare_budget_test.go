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

package util

import (
	"math"
	"testing"
)

func TestSoftShareAicoreBudget(t *testing.T) {
	defer SetSoftShareCoreScaling(1)

	tests := []struct {
		name     string
		scaling  float64
		want     int
		wantKeep float64
	}{
		{name: "default", scaling: 1, want: SoftShareDevCount, wantKeep: 1},
		{name: "oversell 1.5", scaling: 1.5, want: 150, wantKeep: 1.5},
		{name: "oversell 2", scaling: 2, want: 200, wantKeep: 2},
		{name: "clamp below one", scaling: 0.5, want: SoftShareDevCount, wantKeep: 1},
		{name: "clamp zero", scaling: 0, want: SoftShareDevCount, wantKeep: 1},
		{name: "clamp nan", scaling: math.NaN(), want: SoftShareDevCount, wantKeep: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetSoftShareCoreScaling(tt.scaling)
			if SoftShareCoreScaling() != tt.wantKeep {
				t.Fatalf("SoftShareCoreScaling() = %v, want %v", SoftShareCoreScaling(), tt.wantKeep)
			}
			if got := SoftShareAicoreBudget(); got != tt.want {
				t.Fatalf("SoftShareAicoreBudget() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSoftShareAicoreBudgetPacking(t *testing.T) {
	defer SetSoftShareCoreScaling(1)
	SetSoftShareCoreScaling(1.5)
	budget := SoftShareAicoreBudget()
	const perTask = 40
	fitted := 0
	used := 0
	for used+perTask <= budget {
		used += perTask
		fitted++
	}
	if fitted != 3 {
		t.Fatalf("with budget %d, expected 3x%d tasks, got %d (used=%d)", budget, perTask, fitted, used)
	}
	if used+perTask <= budget {
		t.Fatalf("fourth %d%% task should not fit after used=%d budget=%d", perTask, used, budget)
	}
}
