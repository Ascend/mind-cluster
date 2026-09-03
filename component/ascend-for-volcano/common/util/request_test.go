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

func TestParseScheduleMode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want ScheduleMode
	}{
		{"hard", "hard", HardScheduleMode},
		{"empty defaults soft", "", SoftScheduleMode},
		{"soft", "soft", SoftScheduleMode},
		{"invalid value", "harder", SoftScheduleMode},
		{"mixed case is invalid", "HARD", SoftScheduleMode},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseScheduleMode(tt.in); got != tt.want {
				t.Errorf("ParseScheduleMode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestScheduleModeConstants pins along which lane a request must go.
func TestScheduleModeConstants(t *testing.T) {
	if HardScheduleMode != "hard" || SoftScheduleMode != "soft" {
		t.Errorf("ScheduleMode constants changed: hard=%q soft=%q", HardScheduleMode, SoftScheduleMode)
	}
}
