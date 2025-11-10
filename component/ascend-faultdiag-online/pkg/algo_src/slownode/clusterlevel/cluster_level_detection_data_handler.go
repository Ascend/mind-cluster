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

// Package clusterlevel is used for file reading and writing, as well as data processing.
package clusterlevel

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/core/model"
	"ascend-faultdiag-online/pkg/utils/constants"
)

/* callback */
var callbackFunc model.CallbackFunc = nil

func getCurJobAllNodeResultFile(nodeLevelResultPath string, recorder map[string]int64) []string {
	resultPaths := make([]string, 0)
	/* 正则表达式匹配 */
	resultFile := regexp.MustCompile(config.NodeJobDetectionResultFileName)
	/* 确认目录可读 */
	if !config.CheckFileOrDirectoryReadMode(nodeLevelResultPath) {
		return nil
	}
	/* 遍历文件 */
	var fileCount = 0
	err := filepath.Walk(nodeLevelResultPath, func(path string, info os.FileInfo, err error) error {
		fileCount++
		if fileCount >= constants.MaxFileCount {
			return fmt.Errorf("too many files under: %s, exceed max file count: %d", path, constants.MaxFileCount)
		}
		if err != nil {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO]Traverse %s failed!", path)
			return err
		}
		/* check if dir */
		if info.IsDir() {
			return nil
		}
		if !resultFile.MatchString(info.Name()) {
			return nil
		}
		/* 检查文件时间戳，仅保存更新的文件 */
		if lastTime, exists := recorder[info.Name()]; exists && lastTime == info.ModTime().Unix() {
			return nil
		}
		if recorder != nil {
			recorder[info.Name()] = info.ModTime().Unix()
		}
		/* 存储完整路径 */
		resultPaths = append(resultPaths, filepath.Join(nodeLevelResultPath, info.Name()))
		return nil
	})
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]loop check all files under: %s failed: %v", nodeLevelResultPath, err)
		return nil
	}
	return resultPaths
}

/* 检查并添加慢通信域 */
func checkAndAddSlowDomain(slowSendRank int, PPranks []int, allSlowDomainOneNode *[][]int) bool {
	for index, PPrank := range PPranks {
		/* 当前卡在pp通信域中存在且不为当前pp域通信组的最后一个则将当前卡以及后一个卡加入到慢通信结果中 */
		if slowSendRank == PPrank && index != len(PPranks)-1 {
			xx := []int{PPranks[index], PPranks[index+1]}
			*allSlowDomainOneNode = append(*allSlowDomainOneNode, xx)
			return true
		}
	}
	return false
}

/* 处理pp慢通信域的添加 */
func addSlowDomainIfNecessary(slowSendRank int, ppRankss [][]int, allSlowDomainOneNode *[][]int) {
	for _, PPranks := range ppRankss {
		// 根据慢的Send Rank，遍历PP通信域，寻找PP通信域
		if foundAndAdded := checkAndAddSlowDomain(slowSendRank, PPranks, allSlowDomainOneNode); foundAndAdded {
			break
		}
	}
}

func getSlowPpParallelDomains(mergedData *config.ClusterJobResult,
	mergedSlowSendRanks []int,
	ppInfo [][]int) {
	if len(ppInfo) == 0 || len(mergedSlowSendRanks) == 0 {
		hwlog.RunLog.Warn("[SLOWNODE ALGO]Cluster detection empty pp info!")
		return
	}
	ppSlowCommunicateDomains := make([][]int, 0)
	for _, rankI := range mergedSlowSendRanks {
		addSlowDomainIfNecessary(rankI, ppInfo, &ppSlowCommunicateDomains)
	}
	mergedData.SlowCommunicationDomains = append(mergedData.SlowCommunicationDomains, ppSlowCommunicateDomains...)
}

/* 组中每张卡的连接数， 中间的卡左右各连接一张卡，因此连接数为2 */
func initializeConnectNumbers(length int) ([]int, []int, error) {
	if length <= 0 {
		hwlog.RunLog.Error("[SLOWNODE ALGO]length is zero")
		return nil, nil, errors.New("[SLOWNODE ALGO]PPranks len is zero")
	}
	connectNumbers := make([]int, length)
	badConnectNumbers := make([]int, length)

	for index := range connectNumbers {
		/* 将Rank所有数据设置左右两个连接 */
		connectNumbers[index] = config.MaxNpuLinkNumsInDomain
	}
	connectNumbers[0] = 1        // 两端设置值为1
	connectNumbers[length-1] = 1 // 两端设置值为1
	return connectNumbers, badConnectNumbers, nil
}

// 更新坏连接数，对慢Send两边的Rank进行加1
func updateBadConnectNumbers(slowSendRanks, PPranks []int, badConnectNumbers []int) {
	for _, abnormalCard := range slowSendRanks {
		processAbnormalCard(abnormalCard, PPranks, badConnectNumbers)
	}
}

// 处理每个异常Send
func processAbnormalCard(abnormalCard int, PPranks []int, badConnectNumbers []int) {
	// Iterate through PPranks and update badConnectNumbers
	for index, value := range PPranks {
		if abnormalCard == value {
			updateBadConnectNumbersForIndex(index, badConnectNumbers)
		}
	}
}

// 更新指定索引的坏连接数
func updateBadConnectNumbersForIndex(index int, badConnectNumbers []int) {
	if index >= len(badConnectNumbers)-1 {
		return
	}
	// Update for current and next index
	badConnectNumbers[index]++
	badConnectNumbers[index+1]++
}

// 计算ppValue
func calculatePPValue(connectNumbers, badConnectNumbers []int) []float64 {
	ppValue := make([]float64, len(connectNumbers))
	for index := range connectNumbers {
		if index >= len(badConnectNumbers) {
			continue
		}
		/* 当前卡坏掉的链路数/当前卡总的链路数 */
		if connectNumbers[index] != 0 {
			ppValue[index] = float64(badConnectNumbers[index]) / float64(connectNumbers[index])
		} else {
			ppValue[index] = 0
		}
	}
	return ppValue
}

/* 查找异常卡 */
func findAbnormalRanks(PPranks []int, ppValue []float64) []int {
	var abnormalRanks = []int{}
	for index, value := range ppValue {
		/* 仅检测所有链路都坏了了卡 */
		if value != 1 {
			continue
		}
		/* 仅有一条链路的卡和有两条链路的卡分别判断异常卡 */
		if index == 0 && ppValue[index+1] == linkHalfStandard {
			abnormalRanks = append(abnormalRanks, PPranks[index])
		} else if index == len(ppValue)-1 && ppValue[index-1] == linkHalfStandard {
			abnormalRanks = append(abnormalRanks, PPranks[index])
		} else if index > 0 && index < len(ppValue)-1 &&
			(ppValue[index-1] == linkHalfStandard ||
				ppValue[index+1] == linkHalfStandard) {
			abnormalRanks = append(abnormalRanks, PPranks[index])
		}
	}
	return abnormalRanks
}

/* 通过PP通信域和节点级慢send检测结果得出慢网络卡 */
func ppNetworkDetection(slowSendRanks []int, PPrankss [][]int) []int {
	if len(PPrankss) == 0 || len(slowSendRanks) == 0 {
		return []int{}
	}
	var abnormalRanks = []int{}
	for _, PPranks := range PPrankss {
		length := len(PPranks)
		if length <= 0 {
			continue
		}
		// 在此PP通信域内，初始化连接数和坏连接数
		connectNumbers, badConnectNumbers, err := initializeConnectNumbers(length)
		if err != nil {
			continue
		}
		/* 根据慢send卡统计每张卡的异常链路数 */
		updateBadConnectNumbers(slowSendRanks, PPranks, badConnectNumbers)
		/* 计算每张卡坏链路系数 */
		ranksBadLinkValue := calculatePPValue(connectNumbers, badConnectNumbers)
		abnormalRanks = append(abnormalRanks, findAbnormalRanks(PPranks, ranksBadLinkValue)...)
	}
	return abnormalRanks
}

/* 格式化集群侧任务级检测结果 */
func getFormatDetectionResult(clusterResult config.ClusterJobResult,
	conf config.AlgoInputConfig) string {
	/* 暂时不上报此结果 */
	clusterResult.SlowCommunicationRanks = []int{}
	/* get cluster local ip */
	ip, err := config.GetLocalIP()
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]Get local ip failed: %v", err)
		return ""
	}
	/* 若劣化感知结果为非slow */
	if clusterResult.IsSlow == 0 {
		clusterResult.DegradationLevel = "0.0%"
		clusterResult.SlowHostNodes = []string{}
		clusterResult.SlowIORanks = []int{}
		clusterResult.SlowCommunicationDomains = [][]int{}
		clusterResult.SlowCommunicationRanks = []int{}
	}
	/* 大key */
	mainKey := "slownode" + "_" + conf.JobName
	minorKey := ip
	clusterResult.JobName = conf.JobName
	result := make(config.ClusterDetectionResult)
	result[mainKey] = map[string]config.ClusterJobResult{minorKey: clusterResult}
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]Marshal cluster detection result failed: %v", err)
		return ""
	}
	return string(jsonStr)
}

// RegisterClusterLevelCallback 注册集群侧回调
func RegisterClusterLevelCallback(callback model.CallbackFunc) {
	callbackFunc = callback
}

/* 集群侧慢节点A3检测算法流程 */
func jobLevelDetectionA3(ppInfo [][]int, tpInfo [][]int, conf config.AlgoInputConfig,
	recorder map[string]int64) {
	/* 当前job 节点级检测结果 */
	nodeLevelResultPath := filepath.Join(conf.FilePath, conf.JobName, config.NodeDetectionResultDirName)
	/* check 路径 */
	if !config.CheckExistDirectoryOrFile(nodeLevelResultPath, true, "cluster", conf.JobName) {
		return
	}
	/* get all current job result */
	files := getCurJobAllNodeResultFile(nodeLevelResultPath, recorder)
	if files == nil {
		return
	}
	if len(files) == 0 {
		hwlog.RunLog.Warn("[SLOWNODE ALGO]Cluster detection no result file updated!")
		return
	}
	/* 获取所有节点侧检测结果 */
	mergedData := getCurJobAllNodeResultInfo(files, tpInfo)
	/* 通过pp通信域检测节点级慢算子send卡结果 */
	slowPpranks := ppNetworkDetection(mergedData.SlowCommunicationRanks, ppInfo)
	/* 更新慢网络卡 */
	mergedSlowSendRanks := mergedData.SlowCommunicationRanks
	mergedData.SlowCommunicationRanks = slowPpranks
	/* 慢PP并行域检测 */
	getSlowPpParallelDomains(&mergedData, mergedSlowSendRanks, ppInfo)
	/* 格式化检测结果 */
	jsonStr := getFormatDetectionResult(mergedData, conf)
	/* debug */
	hwlog.RunLog.Infof("[SLOWNODE ALGO]Cluster detection result: %s", jsonStr)
	/* call callback report */
	if callbackFunc != nil {
		go callbackFunc(jsonStr)
	}
}
