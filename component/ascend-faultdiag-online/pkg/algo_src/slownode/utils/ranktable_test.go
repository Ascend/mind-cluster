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

// Package utils is a DT collection for func in ranktable
package utils

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/utils"
)

func TestGetNodeRanksFromRanktable(t *testing.T) {
	// ranktable is not exist
	var filePath = "not_exist"
	assert.Equal(t, []int{}, getNodeRanksFromRanktable(filePath))

	mockLoadFile := gomonkey.ApplyFunc(utils.LoadFile, func(string) ([]byte, error) {
		return []byte{}, errors.New("load file error")
	})
	sourceData := ""
	filePath = "ranktable.json"
	assert.Nil(t, generateFile(sourceData, filePath))
	// load file failed, empty
	assert.Equal(t, []int{}, getNodeRanksFromRanktable(filePath))
	mockLoadFile.Reset()

	absPath, err := filepath.Abs(filePath)
	assert.Nil(t, err)
	resoledPath, err := filepath.EvalSymlinks(absPath)
	assert.Nil(t, err)

	// wrong json data
	sourceData = "{"
	assert.Nil(t, generateFile(sourceData, resoledPath))
	assert.Equal(t, []int{}, getNodeRanksFromRanktable(resoledPath))

	// no serverList
	sourceData = `{"server_listssssssssssss":[1,2,3,4]}`
	assert.Nil(t, generateFile(sourceData, resoledPath))
	assert.Equal(t, []int{}, getNodeRanksFromRanktable(resoledPath))

	// wrong serverList
	sourceData = `{"server_list":"string"}`
	assert.Nil(t, generateFile(sourceData, resoledPath))
	assert.Equal(t, []int{}, getNodeRanksFromRanktable(resoledPath))

	sourceData = `{"server_list":[{"server_id":"server_id1",` +
		`"device":[{"rank_id":"123"},{},{"rank_id":"string"}]},` +
		`{"server_id":"server_id2","device":[{"rank_id":"123"},{"rank_id":"456"}]}]}`
	assert.Nil(t, generateFile(sourceData, resoledPath))

	// normal data with no xdlIP
	err = os.Setenv(xdlIpField, "server_id0")
	assert.Nil(t, err)
	assert.Equal(t, []int{}, getNodeRanksFromRanktable(resoledPath))

	// normal data with exist xdlIP
	err = os.Setenv(xdlIpField, "server_id1")
	assert.Nil(t, err)
	assert.Equal(t, []int{123, 0}, getNodeRanksFromRanktable(resoledPath))

	err = clearFile(resoledPath)
	assert.Nil(t, err)
}

func TestBuildIp2Ranks(t *testing.T) {

	testCases := []struct {
		serverList []any
		expect     map[string][]int
	}{
		// empty serverList
		{[]any{}, map[string][]int{}},
		// wrong type of serverData
		{[]any{"test"}, map[string][]int{}},
		// wrong type of serverID
		{[]any{map[string]any{
			"server_id": []int{1, 2, 3},
		}},
			map[string][]int{}},
		// wrong type of deviceList
		{[]any{
			map[string]any{
				"server_id": "server_id_data",
				"device":    "string",
			}},
			map[string][]int{}},
		// wrong type of device which in deviceList
		{[]any{
			map[string]any{
				"server_id": "server_id_data",
				"device":    []string{"string"},
			}},
			map[string][]int{}},
		// wrong type of rank_id
		{[]any{
			map[string]any{
				"server_id": "server_id_data",
				"device": []any{
					map[string]any{"rank_id": "123"},
				},
			}},
			map[string][]int{
				"server_id_data": {123},
			}},
		// mixed wrong type data and correct data
		{[]any{
			map[string]any{
				"server_id": "server_id_data",
				"device": []any{
					map[string]any{"rank_id": "123"},
					map[string]any{"rank_id": "string"},
				},
			}},
			map[string][]int{
				"server_id_data": {123, 0},
			}},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expect, buildIp2Ranks(tc.serverList))
	}
}

func TestStringToInt(t *testing.T) {
	testCases := []struct {
		param    string
		expected int
	}{
		{"1", 1},
		{"test", 0},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, stringToInt(tc.param))
	}
}
