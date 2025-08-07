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

/*
Package handlejson is a DT that read parallel group information from JSON data
*/
package handlejson

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
)

func TestDealParallelGroupInfo(t *testing.T) {
	// 异常校验
	currentDir, err := os.Getwd()
	assert.NoError(t, err)
	modelNotExit := config.DataParseModel{
		FilePath:          currentDir,
		JobName:           "job_name_1",
		JobId:             "job_id_1",
		ParallelGroupPath: []string{"not_exit_file.json"},
	}
	err = MergeParallelGroupInfo(modelNotExit)
	assert.Error(t, err, "文件路径不存在，导致读取json报错")

	modelExit := config.DataParseModel{
		FilePath:          currentDir,
		JobName:           "job_name_2",
		JobId:             "",
		ParallelGroupPath: getJsonPath(t),
	}
	err = MergeParallelGroupInfo(modelExit)
	assert.NoError(t, err)

	savePath := filepath.Join(currentDir, "parallel_group_global.json")
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		assert.NoError(t, err)
	}
	realData := getRealData()
	res := make(map[string]model.OpGroupInfo, 3)
	readData, err := os.ReadFile(savePath)
	assert.NoError(t, err)
	err = json.Unmarshal(readData, &res)
	assert.NoError(t, err)

	for name, info := range res {
		realInfo, ok := realData[name]
		assert.True(t, ok)
		assert.ElementsMatch(t, info.GlobalRanks, realInfo.GlobalRanks)
	}

	// 测试用例执行后删除生成的json文件
	err = os.Remove(savePath)
	assert.NoError(t, err)

}

func getRealData() map[string]model.OpGroupInfo {
	return map[string]model.OpGroupInfo{
		"operator_domain_def": {
			GroupName:   "default_group",
			GroupRank:   0,
			GlobalRanks: []int64{0, 1, 2, 4, 5, 7, 9, 22},
		},
		"operator_domain_dp": {
			GroupName:   "dp",
			GroupRank:   1,
			GlobalRanks: []int64{1, 2, 3, 6, 7, 8},
		},
		"operator_domain_cp": {
			GroupName:   "cp",
			GroupRank:   2,
			GlobalRanks: []int64{},
		},
	}
}

func getJsonPath(t *testing.T) []string {
	data1 := map[string]model.OpGroupInfo{
		"operator_domain_def": {
			GroupName:   "default_group",
			GroupRank:   0,
			GlobalRanks: []int64{0, 1, 2, 22},
		},
		"operator_domain_dp": {
			GroupName:   "dp",
			GroupRank:   1,
			GlobalRanks: []int64{1, 2, 3},
		},
	}
	data2 := map[string]model.OpGroupInfo{
		"operator_domain_def": {
			GroupName:   "default_group",
			GroupRank:   0,
			GlobalRanks: []int64{5, 4, 22},
		},
		"operator_domain_cp": {
			GroupName:   "cp",
			GroupRank:   2,
			GlobalRanks: []int64{},
		},
	}
	data3 := map[string]model.OpGroupInfo{
		"operator_domain_def": {
			GroupName:   "default_group",
			GroupRank:   0,
			GlobalRanks: []int64{9, 7},
		},
		"operator_domain_dp": {
			GroupName:   "dp",
			GroupRank:   1,
			GlobalRanks: []int64{7, 6, 8},
		},
	}

	filePath1 := creatJsonFile(t, data1, "example_1.json")
	filePath2 := creatJsonFile(t, data2, "example_2.json")
	filePath3 := creatJsonFile(t, data3, "example_3.json")
	return []string{filepath.Base(filePath1), filepath.Base(filePath2), filepath.Base(filePath3)}
}

func creatJsonFile(t *testing.T, data map[string]model.OpGroupInfo, name string) string {
	// creating a temporary json file
	currentDir, err := os.Getwd()
	assert.NoError(t, err)
	tmpFile, err := os.CreateTemp(filepath.Join(currentDir, ""), name)
	assert.NoError(t, err)

	// encode to JSON and write to a file
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	assert.NoError(t, err)

	_, err = tmpFile.Write(jsonBytes)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)
	return tmpFile.Name()
}
