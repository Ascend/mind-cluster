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

// Package chip4nodex is using for HuaWei 300I A5 affinity schedule.
package chip4nodex

import (
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

func TestCreateAffScoreList(t *testing.T) {
	affScoreList := [][]int{
		{util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3, util.AffScore4, util.AffScore5,
			util.AffScore6, util.AffScore7},
		{util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3, util.AffScore4,
			util.AffScore5, util.AffScore6},
		{util.AffScore8, util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3,
			util.AffScore4, util.AffScore5},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2,
			util.AffScore3, util.AffScore4},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore0, util.AffScore1,
			util.AffScore2, util.AffScore3},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore0,
			util.AffScore1, util.AffScore2},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore0, util.AffScore1},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore8, util.AffScore0},
	}
	if res := createAffScoreList(maxNodeNPUNumX8); !reflect.DeepEqual(res, affScoreList) {
		t.Errorf("The affinity scoring table does not match expectations; the actual scoring table is:: %v", res)
	}
}

func TestGetNPUNumByHandler(t *testing.T) {
	if n := getNPUNumByHandler(SchedulePolicy4Px8); n != maxNodeNPUNumX8 {
		t.Errorf("getNPUNumByHandler The results do not match the expectations; the actual results: %d", n)
	}
	if n := getNPUNumByHandler(SchedulePolicy4Px16); n != maxNodeNPUNumX16 {
		t.Errorf("getNPUNumByHandler The results do not match the expectations; the actual results: %d", n)
	}
	if n := getNPUNumByHandler("test"); n != 0 {
		t.Errorf("getNPUNumByHandler The results do not match the expectations; the actual results: %d", n)
	}
}
