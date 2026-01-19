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

// Package spacedetector performs space dimension detection by homogenizing the data.
package spacedetector

// 查找二维切片中的公共部分, 故障卡的异常会持续一段时间，通过寻找公共部分找出异常卡
func findCommonElements(data [][]int) []int {
	if len(data) == 0 {
		return nil
	}

	// 先将第一行的元素作为初始的公共部分
	common := make(map[int]bool, len(data[0]))
	for _, val := range data[0] {
		common[val] = true
	}
	// 遍历剩下的每一行，更新公共部分
	for i := 1; i < len(data); i++ {
		currentRow := make(map[int]bool, len(data[i]))
		for _, val := range data[i] {
			currentRow[val] = true
		}

		// 更新公共部分，只保留在每一行中都存在的元素
		for key := range common {
			if !currentRow[key] {
				delete(common, key)
			}
		}
	}
	// 将公共部分转为切片
	var result []int
	for key := range common {
		result = append(result, key)
	}
	return result
}

// 查找二维切片中是否有非公共元素
func checkCommonAndNonCommon(data [][]int, commonRank []int) ([]int, bool) {
	// 创建一个 map 来存储公共rank元素，预估大小
	commonMap := make(map[int]bool, len(commonRank))
	for _, val := range commonRank {
		commonMap[val] = true
	}
	// 遍历二维切片中的每一行
	for _, row := range data {
		// 检查当前行是否包含非公共元素
		for _, val := range row {
			if !commonMap[val] {
				// 存在非公共元素，返回公共部分并返回false
				return commonRank, false
			}
		}
	}
	// 如果所有行的元素都属于公共rank，返回公共部分并返回true
	return commonRank, true
}

// 输入二维切片，输出公共元素并检查是否所有行都包含公共元素
func findCommonAndCheck(data [][]int) ([]int, bool) {
	// 查找公共元素
	commonRank := findCommonElements(data)
	// 检查是否所有行都包含非公共元素
	return checkCommonAndNonCommon(data, commonRank)
}
