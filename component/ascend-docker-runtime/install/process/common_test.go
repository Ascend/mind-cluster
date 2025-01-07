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

package process

import "testing"

// TestCheckParamAndGetBehavior test the function checkParamAndGetBehavior
func TestCheckParamAndGetBehavior(t *testing.T) {
	type args struct {
		action  string
		command []string
	}
	addCmds := []string{"0", "0", "0", "0", "0", "0", "0"}
	rmCmds := []string{"0", "0", "0", "0", "0", "0"}
	var tests = []struct {
		name  string
		args  args
		want  bool
		want1 string
	}{
		{
			name:  "01-add command,should return behavior install",
			args:  args{action: addCommand, command: addCmds},
			want:  true,
			want1: "install",
		}, {
			name:  "01-add command,should return behavior uninstall",
			args:  args{action: rmCommand, command: rmCmds},
			want:  true,
			want1: "uninstall",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := checkParamAndGetBehavior(tt.args.action, tt.args.command)
			if got != tt.want {
				t.Errorf("checkParamAndGetBehavior() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("checkParamAndGetBehavior() got = %v, want %v", got1, tt.want1)
			}
		})
	}
}
