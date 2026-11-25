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

import "testing"

func TestTimeGateAllows(t *testing.T) {
	tests := []struct {
		name      string
		afterTime int64
		elapsed   int64
		want      bool
	}{
		{"forbidden never allows", ToleranceForbidden, 999, false},
		{"zero allows immediately", 0, 0, true},
		{"negative (not forbidden) allows", -5, 0, true},
		{"elapsed before after-time", DefaultSilentUnhealthyToleranceAfterTime, DefaultSilentUnhealthyToleranceAfterTime - 1, false},
		{"elapsed equals after-time", DefaultSilentUnhealthyToleranceAfterTime, DefaultSilentUnhealthyToleranceAfterTime, true},
		{"elapsed after after-time", DefaultSilentUnhealthyToleranceAfterTime, DefaultSilentUnhealthyToleranceAfterTime + 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimeGateAllows(tt.afterTime, tt.elapsed); got != tt.want {
				t.Fatalf("TimeGateAllows(%d, %d) = %v, want %v", tt.afterTime, tt.elapsed, got, tt.want)
			}
		})
	}
}
