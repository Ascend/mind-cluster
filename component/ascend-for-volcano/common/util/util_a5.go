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

// Package util provides utility functions for Ascend 800i A5 related operations,
// including list operations, label checking, and other common utilities used
// across the ascend-for-volcano component.
package util

import (
	"k8s.io/klog"
)

// MergeUnique is the union function but keep only one same item
func MergeUnique(list1, list2 []int) []int {
	seen := make(map[int]bool)
	var result []int
	for _, v := range list1 {
		seen[v] = true
		result = append(result, v)
	}
	for _, v := range list2 {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// HasCommonElement is to find the common element in two int list equal to the function HasAny bug return the common
func HasCommonElement(s1, s2 []int) ([]int, bool) {
	appearance := make(map[int]bool)
	var result []int
	for _, value := range s1 {
		appearance[value] = true
	}
	for _, value := range s2 {
		if _, exists := appearance[value]; exists {
			result = append(result, value)
			delete(appearance, value)
		}
	}
	klog.V(LogDebugLev).Infof("HasCommonElement s1->%v  s2->%v, result->%v", s1, s2, result)
	return result, len(result) != 0
}
