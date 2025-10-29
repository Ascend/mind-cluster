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

// Package applicablescenejudgment 用于判断算法的适用场景，当前暂不支持Moe和CP场景
package applicablescenejudgment

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/fileutils"
)

// Data 解析topo文件中的并行域信息
type Data struct {
	TP [][]int `json:"TP"`
	CP [][]int `json:"CP"`
	PP [][]int `json:"PP"`
	DP [][]int `json:"DP"`
	EP [][]int `json:"EP"`
}

const (
	task         = "task"
	topoFileName = "topo.json"
)

func checkEPContent(filePath string) (bool, error) {
	// 读文件的内容
	data, err := utils.LoadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("error reading file: %v", err)
	}

	// 解析文件
	var jsonData Data
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return false, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	// 如果没有EP字段，返回false
	if jsonData.EP == nil {
		return false, nil
	}

	// 检测EP通信域是否存在只有一张卡的情况，如果只有一张卡那么就返回FALSE，说明当前并不存在EP通信域
	for _, sublist := range jsonData.EP {
		if len(sublist) <= 1 {
			return false, nil
		}
	}

	// 检测EP和TP是否完全相同，如果EP和TP是完全相同的，说明不存在EP，直接返回FALSE； 存在不同，则认为存在EP
	if len(jsonData.EP) != len(jsonData.TP) {
		return true, nil
	}

	for i := range jsonData.EP {
		if len(jsonData.EP[i]) != len(jsonData.TP[i]) {
			return true, nil
		}
		for j := range jsonData.EP[i] {
			if jsonData.EP[i][j] != jsonData.TP[i][j] {
				return true, nil
			}
		}
	}

	// EP和TP的结果是完全相同的，则认为不存在EP
	return false, nil
}

// 检测是否存CP，如果存在会返回true，如果不存在会返回FALSE
func checkCPContent(filePath string) (bool, error) {
	// 读文件的内容
	data, err := fileutils.ReadLimitBytes(filePath, constants.Size10M)
	if err != nil {
		return false, fmt.Errorf("error reading file: %v", err)
	}

	// 解析文件
	var jsonData Data
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return false, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	// 如果没有CP字段，返回false
	if jsonData.CP == nil {
		return false, nil
	}

	// 检测CP通信域是否存在只有一张卡的情况，如果只有一张卡那么就返回FALSE，说明当前并不存在CP通信域
	for _, sublist := range jsonData.CP {
		if len(sublist) <= 1 {
			return false, nil
		}
	}

	// CP的二维矩阵的第二个维度都是大于1的，现在是存在CP通信的
	return true, nil
}

// CheckApplicableScene 判断场景是否使用
func CheckApplicableScene(sndConfig *config.DetectionConfig, taskName string) bool {
	if sndConfig == nil {
		hwlog.RunLog.Error("[SLOWNODE ALGO]invalid nil sndConfig")
		return false
	}
	// 获取当前文件名和行号
	topoPath := filepath.Join(sndConfig.SharedFilePath, task, taskName, topoFileName)
	result, err := checkEPContent(topoPath)
	if err != nil {
		hwlog.RunLog.Error("Error:", err)
		return false
	}

	if result {
		hwlog.RunLog.Error("training uses EP and slownode detection currently does not support EP parallelism")
		return false
	}

	resultCp, err := checkCPContent(topoPath)
	if err != nil {
		hwlog.RunLog.Error("Error:", err)
		return false
	}

	if resultCp {
		hwlog.RunLog.Error("training uses CP and slownode detection currently does not support CP parallelism")
		return false
	}

	return true
}
