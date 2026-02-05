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

// Package chip4nodex for any struct used in module300ia5
package chip4nodex

import "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910a5"

type chip4nodex struct {
	ascend910a5.Base910A5
	affScoreList [][]int
}

const (
	networkUnhealthyNPU = "huawei.com/Ascend910-NetworkUnhealthy"
	unhealthyNPU        = "huawei.com/Ascend910-Unhealthy"
	// SchedulePolicy4Px8 the label 4p-8
	SchedulePolicy4Px8 = "4p-8"
	// SchedulePolicy4Px16 the label 4p-16
	SchedulePolicy4Px16 = "4p-16"
	// SchedulePolicy1Px8 the label 1p-8
	SchedulePolicy1Px8 = "1p-8"
	// SchedulePolicy1Px16 the label 1p-16
	SchedulePolicy1Px16 = "1p-16"
	// maxNodeNPUNumX8  for the max cards num of one node is 8
	maxNodeNPUNumX8 = 8
	// maxNodeNPUNumX16 for the max cards num of one node is 16
	maxNodeNPUNumX16 = 16
	// cardsNumPerMesh for the cards num in per mesh is 4
	cardsNumPerMesh = 4
	// scoreWeightX16 for the score weight in 300I-A5-4p-8 is 16
	scoreWeightX16 = 16
	// scoreWeightX64 for the score weight in 300I-A5-4p-16 is 64
	scoreWeightX64 = 64
)
