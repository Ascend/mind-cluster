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

// Package utils is a DT collection for func in get_detection_groups
package utils

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
)

type condition string

const (
	cd1 condition = "condition1"
	cd2 condition = "condition2"
	cd3 condition = "condition3"
	cd4 condition = "condition4"
)

func TestGetNodeGlobalRanks(t *testing.T) {
	var dirPath = "."
	// mock Walk
	mockWalk := gomonkey.ApplyFunc(filepath.Walk, func(root string, fn filepath.WalkFunc) error {
		return errors.New("filepath walk error")
	})

	data, err := getNodeGlobalRanks(dirPath)
	assert.Equal(t, "filepath walk error", err.Error())
	assert.Nil(t, data)
	mockWalk.Reset()

	// generate file
	fileNames := make([]string, 0)
	var batchGenerateFile = func(fileIDs []int) {
		for _, id := range fileIDs {
			fileName := fmt.Sprintf("global_rank_%d.csv", id)
			assert.Nil(t, generateFile("", fileName))
			fileNames = append(fileNames, fileName)
		}
	}

	correctIds := []int{0, 1, 10, 9999999}
	incorrectIds := []int{999999999, -1, 1111111111}
	batchGenerateFile(correctIds)
	batchGenerateFile(incorrectIds)

	data, err = getNodeGlobalRanks(dirPath)
	assert.Nil(t, err)
	assert.ElementsMatch(t, correctIds, data)

	// clear file
	for _, fileName := range fileNames {
		assert.Nil(t, clearFile(fileName))
	}

}

func TestGetDetectionGroups(t *testing.T) {
	var testCases = []struct {
		tpranks        [][]int
		nodeGlobalRank []int
		expect         [][]int
	}{
		{[][]int{}, []int{}, [][]int{}},
		{[][]int{{1, 2, 3}, {4, 5}, {7, 8}}, []int{1, 10, 100}, [][]int{{1}}},
		{[][]int{{1, 2, 3}, {4, 5, 100}, {7, 8, 10}}, []int{1, 10, 100}, [][]int{{1}, {100}, {10}}},
		{[][]int{{1, 2, 3}, {4, 5, 100}, {7, 8, 10}}, []int{99, 9999, 99999}, [][]int{}},
	}

	for _, tc := range testCases {
		assert.ElementsMatch(t, tc.expect, getDetectionGroups(tc.tpranks, tc.nodeGlobalRank))
	}
}

func TestGetGloRanksAndDetGroups(t *testing.T) {
	var getNodeGlobalRanksFailed = false
	var cd = cd1
	patches := mockFunc(&cd, &getNodeGlobalRanksFailed)
	defer patches.Reset()
	var sndConfig = &config.DetectionConfig{CardsOneNode: 1}

	// getNodeGlobalRanksFailed
	getNodeGlobalRanksFailed = true
	globalRanksNode, detectionGroups := GetGloRanksAndDetGroups(sndConfig, "")
	assert.Nil(t, globalRanksNode)
	assert.Nil(t, detectionGroups)
	getNodeGlobalRanksFailed = false

	// len(nodeRanksFromRanktable) is greater than 0
	// and len(globalRanksNode) is greater than len(nodeRanksFromRanktable)
	cd = cd1
	globalRanksNode, detectionGroups = GetGloRanksAndDetGroups(sndConfig, "")
	assert.Equal(t, []int{1, 2, 3, 4, 5}, globalRanksNode)
	assert.Equal(t, [][]int{{1}, {2, 3}, {4, 5}}, detectionGroups)

	// len(nodeRanksFromRanktable) is greater than 0 and len(globalRanksNode) equals len(nodeRanksFromRanktable)
	cd = cd2
	globalRanksNode, detectionGroups = GetGloRanksAndDetGroups(sndConfig, "")
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, globalRanksNode)
	assert.Equal(t, [][]int{{1}, {2, 3}, {4, 5, 6}}, detectionGroups)

	// len(nodeRanksFromRanktable) is greater than 0 and len(globalRanksNode) is less than len(nodeRanksFromRanktable)
	cd = cd3
	globalRanksNode, detectionGroups = GetGloRanksAndDetGroups(sndConfig, "")
	assert.ElementsMatch(t, []int{}, globalRanksNode)
	assert.ElementsMatch(t, [][]int{}, detectionGroups)

	// len(nodeRanksFromRanktable) equals 0 and len(tpParallelRanks) equals 0
	cd = cd3
	globalRanksNode, detectionGroups = GetGloRanksAndDetGroups(sndConfig, "")
	assert.ElementsMatch(t, []int{}, globalRanksNode)
	assert.ElementsMatch(t, [][]int{}, detectionGroups)
}

func mockFunc(cd *condition, getNodeGlobalRanksFailed *bool) *gomonkey.Patches {
	// mock func appear in GetGloRanksAndDetGroups
	patches := gomonkey.ApplyFunc(getNodeGlobalRanks, func(string) ([]int, error) {
		if *getNodeGlobalRanksFailed {
			return nil, errors.New("getNodeGlobalRanksFailed")
		}
		return []int{1, 2, 3, 4, 5, 6}, nil
	})

	patches.ApplyFunc(getNodeRanksFromRanktable, func(string) []int {
		switch *cd {
		case cd1:
			return []int{1, 2, 3, 4, 5}
		case cd2:
			return []int{1, 2, 3, 4, 5, 6}
		case cd3:
			return []int{1, 2, 3, 4, 5, 6, 7}
		default:
			return []int{}
		}
	})

	patches.ApplyFunc(getTPParallel, func(string) ([][]int, error) {
		switch *cd {
		case cd1, cd2, cd3:
			return [][]int{{1}, {2, 3}, {4, 5, 6}}, nil
		default:
			return nil, errors.New("getTPParallel faild")
		}
	})
	return patches
}
