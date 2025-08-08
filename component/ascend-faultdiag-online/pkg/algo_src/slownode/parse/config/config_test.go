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

// Package config is a DT collection for func in config.
package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateConfig(configPath string, config map[string]any) error {
	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}
	var perm os.FileMode = 0644
	return os.WriteFile(configPath, jsonData, perm)

}

func clearConfig(path string) error {
	return os.Remove(path)
}

func TestNewSlowNodeParserConfig(t *testing.T) {
	// wrong config path
	_, err := NewSlowNodeParserConfig("notExist")
	assert.Equal(t, "open notExist: no such file or directory", err.Error())

	// error config
	errConfig := map[string]any{
		"db_file_path":                         "",
		"global_rank_csv_file_path":            "",
		"step_time_csv_file_path":              "",
		"parallel_group_json_input_file_path":  "",
		"parallel_group_json_output_file_path": "",
		"traffic":                              "ttt",
	}
	errConfigPath := "err_config.json"
	err = generateConfig(errConfigPath, errConfig)
	assert.Nil(t, err)
	_, err = NewSlowNodeParserConfig(errConfigPath)
	assert.ErrorContains(t, err, "cannot unmarshal string into Go struct field SlowNodeParserConfig")
	err = clearConfig(errConfigPath)
	assert.Nil(t, err)
	// correct config
	correctConfig := map[string]any{
		"db_file_path":                         "",
		"global_rank_csv_file_path":            "",
		"step_time_csv_file_path":              "",
		"parallel_group_json_input_file_path":  "",
		"parallel_group_json_output_file_path": "",
		"traffic":                              123,
	}
	correctConfigPath := "config.json"
	err = generateConfig(correctConfigPath, correctConfig)
	assert.Nil(t, err)
	config, err := NewSlowNodeParserConfig(correctConfigPath)
	assert.Nil(t, err)
	assert.Equal(t, "", config.DbFilePath)
	assert.Equal(t, "", config.GlobalRankCsvFilePath)
	assert.Equal(t, "", config.StepTimeCsvFilePath)
	assert.Equal(t, "", config.ParGroupJsonInputFilePath)
	assert.Equal(t, "", config.ParGroupJsonOutputFilePath)
	assert.Equal(t, int64(123), config.Traffic)
	err = clearConfig(correctConfigPath)
	assert.Nil(t, err)
}
