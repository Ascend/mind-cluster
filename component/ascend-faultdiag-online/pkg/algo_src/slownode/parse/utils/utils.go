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

// Package utils provides some common utils
package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/utils/fileutils"
)

const splitLen = 2

// ToStringList 将变量列表转为string列表
func ToStringList(srcList []any) []string {
	var res []string
	for _, srcItem := range srcList {
		res = append(res, fmt.Sprintf("%v", srcItem))
	}
	return res
}

// UniqueSlice 切片去重
func UniqueSlice[T comparable](slice []T) []T {
	uniqueMap := make(map[T]struct{}, len(slice))
	var res []T

	for _, v := range slice {
		if _, exists := uniqueMap[v]; !exists {
			uniqueMap[v] = struct{}{}
			res = append(res, v)
		}
	}
	return res
}

// InSlice 判断元素是否在切片里面
func InSlice[T comparable](slice []T, element T) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

// IndexOf 在切片s中查找value的首次出现位置，返回其索引；若未找到返回-1
func IndexOf[T comparable](s []T, value T) int {
	for i, v := range s {
		if v == value {
			return i
		}
	}
	return -1
}

// Poller 通用轮训器
func Poller(conditionFunc func() (bool, error), pollInterval, timeout time.Duration, stopChan chan struct{}) error {
	pollTicker := time.NewTicker(pollInterval)
	defer pollTicker.Stop()

	var timeoutChan <-chan time.Time
	if timeout > 0 {
		timeoutChan = time.After(timeout)
	}

	var tempStopChan = make(chan struct{})
	if stopChan != nil {
		tempStopChan = stopChan
	}
	for {
		select {
		case <-pollTicker.C:
			result, err := conditionFunc()
			if result {
				return nil
			} else if err != nil {
				return err
			}
		case <-timeoutChan:
			return fmt.Errorf("timeout: %v", timeout)
		case _, ok := <-tempStopChan:
			if !ok {
				return nil
			}
		}
	}
}

// StrContains 是否包含字符串
func StrContains(strSlice []string, txt string) bool {
	if strSlice == nil {
		return false
	}

	for _, str := range strSlice {
		if strings.Contains(str, txt) {
			return true
		}
	}
	return false
}

// RemoveAllFile 删除文件
func RemoveAllFile(filePaths []string) error {
	for _, filePath := range filePaths {
		if filePath == "" {
			continue
		}
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue
		}
		absFilePath, err := fileutils.CheckPath(filePath)
		if err != nil {
			continue
		}
		if err := os.Remove(absFilePath); err != nil {
			return fmt.Errorf("failed to remove %s: %v", filePath, err)
		}
	}
	return nil
}

// AllZero 判断所有值是否都为0（为空默认返回true）
func AllZero(nums []int64) bool {
	for _, n := range nums {
		if n != 0 {
			return false
		}
	}
	return true
}

// SubtractAndDedupe 实现切片相减
func SubtractAndDedupe[T comparable](slice1, slice2 []T) []T {
	exclude := make(map[T]bool, len(slice2))
	for _, v := range slice2 {
		exclude[v] = true
	}

	// 在相减的同时去重
	result := make([]T, 0)
	seen := make(map[T]bool, len(slice1))
	for _, v := range slice1 {
		if !exclude[v] && !seen[v] {
			result = append(result, v)
			seen[v] = true
		}
	}
	return result
}

// SplitNum 切分字符串获取第一个数字（e.g: "step 1" --> 1）
func SplitNum(data string) (int64, error) {
	parts := strings.Fields(data)
	if len(parts) < splitLen {
		return 0, fmt.Errorf("failed to split data: %s", data)
	}
	stepId, err := strconv.ParseInt(parts[1], constants.DecimalMark, constants.Base64Mark)
	if err != nil {
		return 0, fmt.Errorf("failed to convert data: %v", err)
	}
	return stepId, nil
}
