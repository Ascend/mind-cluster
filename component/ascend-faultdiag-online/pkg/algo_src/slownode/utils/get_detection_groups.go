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
Package utils is used for file reading and writing, as well as data processing.
*/
package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/utils/constants"
)

// getNodeGlobalRanks 遍历文件夹，获取所有符合格式 global_rank_i.csv 的文件，并提取 i 值
func getNodeGlobalRanks(dirPath string) ([]int, error) {
	// 存储 i 值的切片
	var iValues = []int{}

	// 正则表达式，匹配 "global_rank_i.csv" 格式的文件
	re := regexp.MustCompile(`global_rank_(\d{1,7})\.csv`)

	// 遍历目录中的所有文件
	fileCount := 0
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		fileCount++
		if fileCount >= constants.MaxFileCount {
			return fmt.Errorf("too many files in: %s, exceed max file count: %d", path, constants.MaxFileCount)
		}
		if err != nil {
			hwlog.RunLog.Warnf("Error accessing path: %v", err)
			return nil
		}
		// 检查是否是文件，并且符合命名规则
		if info.IsDir() {
			return nil
		}
		if !re.MatchString(info.Name()) {
			return nil
		}
		// 提取文件名中的 i 值
		matches := re.FindStringSubmatch(info.Name())
		if len(matches) == 0 {
			return nil
		}
		i, err := strconv.Atoi(matches[1])
		if err != nil {
			// 输出错误信息并跳过该文件
			hwlog.RunLog.Errorf("Error parsing global_rank value from file: %s, error: %v",
				info.Name(), err)
			return nil // or continue if you prefer to continue processing
		}
		// 将 i 添加到切片中
		iValues = append(iValues, i)
		return nil
	})

	// 如果有错误，返回错误
	if err != nil {
		return nil, err
	}
	// 返回提取的所有 i 值
	return iValues, nil
}

// 获取测试组
func getDetectionGroups(tpranks [][]int, nodeGlobalRank []int) [][]int {

	// 将 nodeGlobalRank 转换为 map，便于快速查找 判断TP域中的Rank是否在本地，主要是为了处理：将TP设置的较大，出现跨节点的情况
	rankMap := make(map[int]bool)
	for _, rank := range nodeGlobalRank {
		rankMap[rank] = true
	}

	var DetectionGroups [][]int
	// 遍历 tps 中的每个tp, 这是一个二维数组，每一个数组表示一个TP域
	for _, subRankList := range tpranks {
		var validRanks []int
		// 检查TP通信域中的每个 rank 是否在 nodeGlobalRank 中
		for _, rank := range subRankList {
			if rankMap[rank] {
				validRanks = append(validRanks, rank)
			}
		}
		// 如果该TP通信域中有有效的 rank，加入到 DetectionGroups 中
		if len(validRanks) > 0 {
			DetectionGroups = append(DetectionGroups, validRanks)
		}
	}
	return DetectionGroups
}

// GetGloRanksAndDetGroups获取节点上的全局Ranks一维列表，检测组的二维列表
func GetGloRanksAndDetGroups(sndConfig *config.DetectionConfig, taskName string) ([]int, [][]int) {
	if sndConfig == nil {
		hwlog.RunLog.Error("[SLOWNODE ALGO]invalid nil sndConfig")
		return nil, nil
	}
	// 某一个task本地数据的路径
	taskLocalDataPath := filepath.Join(sndConfig.LocalFilePath, taskName)

	// 某一个task的并行域信息
	topoPath := filepath.Join(sndConfig.SharedFilePath, task, taskName, topoName)
	hwlog.RunLog.Infof("topoPath: %s", topoPath)

	globalRanksNode, err := getNodeGlobalRanks(taskLocalDataPath)
	if err != nil {
		hwlog.RunLog.Errorf("getNodeGlobalRanks err %v", err)
		return nil, nil
	}
	hwlog.RunLog.Infof("The ranks on the local node of the task %s are %v", taskName, globalRanksNode)

	// Ranktable文件尝试读取，不同的代际，生成的ranktable文件内容可能不同，
	rankTablePath := filepath.Join(sndConfig.SharedFilePath, task, taskName, ranktableName)
	nodeRanksFromRanktable := getNodeRanksFromRanktable(rankTablePath)
	hwlog.RunLog.Infof("Global rank of this node read from ranktable: %v", nodeRanksFromRanktable)

	// 读取task_x的topo！  topo文件会存储在共享文件路径下
	tpParallelRanks, _ := getTPParallel(topoPath)
	hwlog.RunLog.Infof("TPranks: %v", tpParallelRanks)

	if len(nodeRanksFromRanktable) > 0 {
		if len(globalRanksNode) > len(nodeRanksFromRanktable) {
			hwlog.RunLog.Infof("Is the operator data mistakenly stored in the shared path? Task: %s, "+
				"Ranks on the node should be: %v",
				taskName, nodeRanksFromRanktable)
			globalRanksNode = nodeRanksFromRanktable
		} else if len(globalRanksNode) == len(nodeRanksFromRanktable) {
			hwlog.RunLog.Info("Operator data is stored in the local directory...")
		} else {
			hwlog.RunLog.Warn("Some cards on this node have not collected data.")
			return nil, nil
		}
	} else {
		// 没有ranktable文件，或者ranktable文件无法解析（可能是不同代际的硬件生成不同的ranktable文件所致）
		if len(tpParallelRanks) == 0 || len(tpParallelRanks[0]) == 0 {
			hwlog.RunLog.Error("Error: tpParallelRanks data is incomplete.")
		}
		if len(globalRanksNode) > sndConfig.CardsOneNode {
			hwlog.RunLog.Errorf("globalRanksNode: %+v", globalRanksNode)
			hwlog.RunLog.Errorf("CardsOneNode: %+v", sndConfig.CardsOneNode)
			hwlog.RunLog.Errorf("Confirm whether the operator data is stored in the local path."+
				" There are %d cards locally, but there are %d operator data.",
				sndConfig.CardsOneNode, len(globalRanksNode))
		}
	}
	detectionGroups := getDetectionGroups(tpParallelRanks, globalRanksNode)
	hwlog.RunLog.Infof("Detection group: %v", detectionGroups)
	return globalRanksNode, detectionGroups
}
