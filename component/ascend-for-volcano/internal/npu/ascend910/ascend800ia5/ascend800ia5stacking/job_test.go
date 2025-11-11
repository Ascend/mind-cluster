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
package ascend800ia5stacking is using for HuaWei Ascend800ia5 pin affinity schedule.
*/
package ascend800ia5stacking

import (
	"strings"
	"testing"
)

const RequiredNodesForTestThree = 3
const RequiredNodesForTestFour = 4

func TestJudgeNodeAndTaskNPU(t *testing.T) {
	mod := &module800ia5stacking{}
	if err := mod.JudgeNodeAndTaskNPU(RequiredNodesForTestThree, []int{0, 1, 2}); err != nil {
		t.Errorf("expected no error when required=3 and nodeTop length is 3, got: %v", err)
	}
	if err := mod.JudgeNodeAndTaskNPU(RequiredNodesForTestFour, []int{0, 1, 2}); err == nil {
		t.Error("expected error when required=4 and nodeTop length is 3, got nil")
	} else if !strings.Contains(err.Error(), "not meet req npu") {
		t.Errorf("unexpected error message: %v", err)
	}
}
