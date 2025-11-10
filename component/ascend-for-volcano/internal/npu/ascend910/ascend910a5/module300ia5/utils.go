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

// Package module300ia5 for any common functions used in module300ia5
package module300ia5

import (
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// createAffScoreList for 300I-A5-8 or 300I-A5-16
/*
	The horizontal axis represents the number of available cards, and the vertical axis represents the number of required cards.
	The scoring sheet is generated as follows:
	[
	  [ 0,  1,  2,  3,  4,  5,  6,  7,  8,  9, 10, 11, 12, 13, 14, 15],
	  [16,  0,  1,  2,  3,  4,  5,  6,  7,  8,  9, 10, 11, 12, 13, 14],
	  [16, 16,  0,  1,  2,  3,  4,  5,  6,  7,  8,  9, 10, 11, 12, 13],
	  [16, 16, 16,  0,  1,  2,  3,  4,  5,  6,  7,  8,  9, 10, 11, 12],
	  [16, 16, 16, 16,  0,  1,  2,  3,  4,  5,  6,  7,  8,  9, 10, 11],
	  [16, 16, 16, 16, 16,  0,  1,  2,  3,  4,  5,  6,  7,  8,  9, 10],
	  [16, 16, 16, 16, 16, 16,  0,  1,  2,  3,  4,  5,  6,  7,  8,  9],
	  [16, 16, 16, 16, 16, 16, 16,  0,  1,  2,  3,  4,  5,  6,  7,  8],
	  [16, 16, 16, 16, 16, 16, 16, 16,  0,  1,  2,  3,  4,  5,  6,  7],
	  [16, 16, 16, 16, 16, 16, 16, 16, 16,  0,  1,  2,  3,  4,  5,  6],
	  [16, 16, 16, 16, 16, 16, 16, 16, 16, 16,  0,  1,  2,  3,  4,  5],
	  [16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16,  0,  1,  2,  3,  4],
	  [16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16,  0,  1,  2,  3],
	  [16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16,  0,  1,  2],
	  [16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16,  0,  1],
	  [16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16,  0]
	]
*/
func createAffScoreList(maxCardNum int) [][]int {
	affScoreList := make([][]int, maxCardNum)
	for i := 0; i < maxCardNum; i++ {
		affScoreList[i] = make([]int, maxCardNum)
		for j := 0; j < maxCardNum; j++ {
			if i > j {
				affScoreList[i][j] = maxCardNum
				continue
			}
			affScoreList[i][j] = j - i
		}
	}
	return affScoreList
}

// the first int is the max node nup num like 8 or 16, if return 0 means split failed
// getNPUNumByHandler split the name string return values
func getNPUNumByHandler(name string) int {
	switch name {
	case Ascend300I4Px8Label:
		return maxNodeNPUNumX8
	case Ascend300I4Px16Label:
		return maxNodeNPUNumX16
	case Ascend300Ix8Label:
		return maxNodeNPUNumX8
	case Ascend300Ix16Label:
		return maxNodeNPUNumX16
	default:
		klog.V(util.LogErrorLev).Infof("found an unsupported handler name %s", name)
		return 0
	}
}
