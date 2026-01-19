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

// Package algo 网络连通性检测算法
package algo

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
)

// 字符串切片翻转
func reverseSlice(s []string) []string {
	// 创建一个新的切片，长度与原切片相同
	reversed := make([]string, len(s))

	// 反向填充新的切片
	for i, v := range s {
		reversed[len(s)-1-i] = v
	}

	return reversed
}

// 字符串切片去重
func uniqueSlice(s []string) []string {
	if len(s) == 0 {
		return []string{}
	}

	uniqueMap := make(map[string]struct{})
	var uniqueValues []string

	for _, str := range s {
		if _, exists := uniqueMap[str]; !exists {
			uniqueMap[str] = struct{}{}
			// 只有在未存在时才添加
			uniqueValues = append(uniqueValues, str)
		}
	}

	return uniqueValues
}

// 获取出现一次的字符串
func appearOnce(input []string) []string {
	if len(input) == 0 {
		return []string{}
	}

	countMap := make(map[string]int)

	// 统计每个字符串的出现次数
	for _, str := range input {
		countMap[str]++
	}

	var result []string
	for str, count := range countMap {
		if count == 1 {
			result = append(result, str)
		}
	}

	return result
}

// 字符串切片移除子切片里的内容
func removeElements(a, b []string) []string {
	if len(a) == 0 {
		return []string{}
	}

	// 创建一个 map 来存储 b 中的元素
	bMap := make(map[string]struct{})
	for _, item := range b {
		bMap[item] = struct{}{}
	}

	// 创建一个切片来存储结果
	var result []string
	for _, item := range a {
		if _, exists := bMap[item]; !exists {
			result = append(result, item)
		}
	}

	return result
}

// 返回一个新的切片，整体元素向左移动一位
func moveSliceLeftOneStep(slice []string) []string {
	if len(slice) == 0 {
		return slice
	}

	newSlice := make([]string, len(slice))

	copy(newSlice, slice[1:])
	newSlice[len(newSlice)-1] = slice[0]

	return newSlice
}

// 字符串切片是否包含指定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 获取字符串在切片里的索引值
func getIndex(slice []string, target string) int {
	for index, value := range slice {
		if value == target {
			return index
		}
	}
	return -1
}

// 检查 map 中是否包含特定键
func containsKey(m map[string]any, key string) bool {
	_, exists := m[key]
	return exists
}

// 对input去重
func deduplicateSlice(input []map[string]any) []map[string]any {
	seen := make(map[string]struct{})
	result := make([]map[string]any, 0, len(input))

	for _, item := range input {
		// 生成唯一指纹（JSON序列化）
		bytes, err := json.Marshal(item)
		if err != nil {
			continue
		}
		fingerprint := string(bytes)

		// 检查是否已存在
		if _, exists := seen[fingerprint]; !exists {
			seen[fingerprint] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// 使用 math.Round 函数进行四舍五入，保留3位小数
func roundToThreeDecimal(value float64) float64 {
	const expectedRate = float64(1000) // 预期保留的小数位数比例
	roundedValue := math.Round(value*expectedRate) / expectedRate
	return roundedValue
}

// MergeAndDeduplicate 合并两个 []map[string]any 的函数
func MergeAndDeduplicate(a, b []map[string]any) []map[string]any {
	// 快速路径检查（减少嵌套深度）
	if len(a) == 0 && len(b) == 0 {
		return a
	}
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}

	result := make([]map[string]any, 0, len(a)+len(b))
	unique := make(map[uint64]struct{}, len(a)+len(b))
	hasher := fnv.New64a()

	// 提取哈希计算逻辑为独立函数（降低深度）
	getHash := func(item map[string]any) (uint64, error) {
		hasher.Reset()
		keys := make([]string, 0, len(item))
		for k := range item {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			if _, err := hasher.Write([]byte(k)); err != nil {
				return 0, err
			}
			if _, err := fmt.Fprintf(hasher, "%v", item[k]); err != nil {
				return 0, err
			}
		}
		return hasher.Sum64(), nil
	}

	// 合并逻辑复用
	for _, items := range [][]map[string]any{a, b} {
		for _, item := range items {
			hashKey, err := getHash(item)
			if err != nil {
				continue // 或记录日志
			}
			if _, exists := unique[hashKey]; !exists {
				unique[hashKey] = struct{}{}
				result = append(result, item)
			}
		}
	}

	return result
}

// NewNetDetect 创建NetDetect对象
func NewNetDetect(superPodId string) *NetDetect {
	return &NetDetect{
		curSuperPodId:        superPodId,
		curNpuType:           "",
		curServerIdMap:       make(map[string]string),
		curFullPingFlag:      false,
		curOpenQueueFlag:     false,
		curSuperPodJobFlag:   false,
		curSuperPodArr:       make([]string, 0),
		curAxisStrategy:      crossAxisConstant,
		curTopo:              make([]string, 0),
		curPingPeriod:        defaultPingPeriod,
		curSuppressedPeriod:  defaultSPeriod,
		curDetectParams:      make(map[string]any, 2),
		curNpuInfo:           make(map[string]NpuInfo, sampleNum),
		curSlideWindows:      make([]map[string]any, 0),
		curConsumedQueue:     make([]map[string]any, 0),
		curSlideWindowsMaxTs: int64(0),
		pathIndex:            make(map[string][]map[string]any),
	}
}
