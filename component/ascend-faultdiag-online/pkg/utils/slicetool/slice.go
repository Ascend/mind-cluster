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
Package slicetool 提供切片相关的能力
*/
package slicetool

import (
	"fmt"
)

// In 判断在切片中是否存在某个元素，如果不存在则返回错误信息
func In[T comparable](value *T, slice []*T) error {
	for _, v := range slice {
		if v == nil || value == nil {
			continue
		}
		if *v == *value {
			return nil
		}
	}
	return fmt.Errorf("the parameter %v is not in the list: %v", value, slice)
}

// ValueIn 判断在切片中是否存在某个元素，如果不存在则返回错误信息
func ValueIn[T comparable](value T, slice []T) error {
	var contains = func(slice []T, value T) bool {
		for _, v := range slice {
			if v == value {
				return true
			}
		}
		return false
	}
	if contains(slice, value) {
		return nil
	}
	return fmt.Errorf("the parameter %v is not in the list: %v", value, slice)
}

// Map 切片映射函数
func Map[T, U any](slice []*T, mapper func(*T) *U) []*U {
	result := make([]*U, len(slice))
	for i, item := range slice {
		if mapper == nil {
			continue
		}
		result[i] = mapper(item)
	}
	return result
}

// FilterPtr 切片过滤函数
func Filter[T any](slice []*T, predicate func(*T) bool) []*T {
	var result []*T
	for _, item := range slice {
		if predicate == nil {
			continue
		}
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// MapToValue convert the data in a array to a new array by function provides
func MapToValue[T, U any](slice []*T, mapper func(*T) U) []U {
	result := make([]U, len(slice))
	for i, item := range slice {
		result[i] = mapper(item)
	}
	return result
}

// Any 切片任意元素满足条件函数
func Any[T any](slice []*T, predicate func(*T) bool) bool {
	for _, item := range slice {
		if predicate == nil {
			continue
		}
		if predicate(item) {
			return true
		}
	}
	return false
}

// All 切片所有元素都满足条件函数
func All[T any](slice []*T, predicate func(*T) bool) bool {
	for _, item := range slice {
		if predicate == nil {
			return false
		}
		if !predicate(item) {
			return false
		}
	}
	return true
}

// Chain convert the [][]any data to []any
func Chain[T any](slices [][]*T) []*T {
	var result []*T
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
}
