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

// Package spacedetector is a DT collection for func in check_uncommon_elements
package spacedetector

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
)

const logLineLength = 256

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout:  true,
		MaxLineLength: logLineLength,
	}
	err := hwlog.InitRunLogger(&config, context.TODO())
	if err != nil {
		fmt.Println(err)
	}
}

func TestFindCommonElements(t *testing.T) {
	var testCases = []struct {
		data   [][]int
		expect []int
	}{
		{[][]int{}, []int{}},
		{[][]int{{1}, {2, 3}}, []int{}},
		{[][]int{{1}, {2, 3, 1}, {9, 9, 1}}, []int{1}},
		{[][]int{{1}, {2, 3, 1}, {9, 9, 1, 2}}, []int{1}},
		{[][]int{{1}, {2, 3, 1}, {9, 9, 2}}, []int{}},
	}

	for _, tc := range testCases {
		assert.ElementsMatch(t, tc.expect, findCommonElements(tc.data))
	}
}

func TestCheckCommonAndNonCommon(t *testing.T) {
	var testCases = []struct {
		data             [][]int
		commonRank       []int
		expectCommonRank []int
		expectBool       bool
	}{
		{[][]int{}, []int{1, 2, 3}, []int{1, 2, 3}, true},
		{[][]int{{1}}, []int{1, 2, 3}, []int{1, 2, 3}, true},
		{[][]int{{4}}, []int{1, 2, 3}, []int{1, 2, 3}, false},
	}

	for _, tc := range testCases {
		actualCommonRank, actualBool := checkCommonAndNonCommon(tc.data, tc.commonRank)
		assert.ElementsMatch(t, tc.expectCommonRank, actualCommonRank)
		assert.Equal(t, tc.expectBool, actualBool)
	}
}

func TestFindCommonAndCheck(t *testing.T) {
	var testCases = []struct {
		data             [][]int
		expectCommonRank []int
		expectBool       bool
	}{
		{[][]int{}, []int{}, true},
		{[][]int{{1, 2, 3}}, []int{1, 2, 3}, true},
		{[][]int{{1, 2, 3}, {3}}, []int{3}, false},
		{[][]int{{1, 2, 3}, {1, 2, 3}, {4, 5, 6}}, []int{}, false},
		{[][]int{{1, 2, 3}, {1, 2, 3}, {3, 3, 4}}, []int{3}, false},
	}

	for _, tc := range testCases {
		actualCommonRank, actualBool := findCommonAndCheck(tc.data)
		assert.ElementsMatch(t, tc.expectCommonRank, actualCommonRank)
		assert.Equal(t, tc.expectBool, actualBool)
	}
}
