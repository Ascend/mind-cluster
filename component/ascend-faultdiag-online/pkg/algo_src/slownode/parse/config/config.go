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
Package config provides some funcs relevant to the config
*/
package config

import (
	"encoding/json"
	"os"
)

// SlowNodeParserConfig 慢节点清洗配置
type SlowNodeParserConfig struct {
	RankDir                    string `json:"rank_dir"`                             // rank文件位置
	DbFilePath                 string `json:"db_file_path"`                         // db文件位置
	GlobalRankCsvFilePath      string `json:"global_rank_csv_file_path"`            // comm.csv文件目录
	StepTimeCsvFilePath        string `json:"step_time_csv_file_path"`              // steptime.csv文件目录
	ParGroupJsonInputFilePath  string `json:"parallel_group_json_input_file_path"`  // parallel_group.json输入路径
	ParGroupJsonOutputFilePath string `json:"parallel_group_json_output_file_path"` // parallel_group.json输出路径
	Traffic                    int64  `json:"traffic"`                              // 通信量
}

// NewSlowNodeParserConfig 慢节点配置构造函数
func NewSlowNodeParserConfig(configPath string) (*SlowNodeParserConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var config SlowNodeParserConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
