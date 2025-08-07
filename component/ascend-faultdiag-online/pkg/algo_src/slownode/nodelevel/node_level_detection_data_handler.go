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

// Package nodelevel is used for file reading and writing, as well as data processing.
package nodelevel

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"sync"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/nodeleveldatarecorder"
	"ascend-faultdiag-online/pkg/algo_src/slownode/spacedetector"
	"ascend-faultdiag-online/pkg/algo_src/slownode/timedetector"
	"ascend-faultdiag-online/pkg/core/model"
)

/* callback */
var callbackFunc model.CallbackFunc = nil

/* report sync lock */
var nodeReportSyncLock sync.Mutex

/* 调用同质化聚类算法 */
func useHomogenizeDetection(haveDataRanks []int,
	detectionDatas [][]float64,
	conf config.AlgoInputConfig,
	slowCalculateRanks *[]int) {
	curTpGroupAbnormalRanks := spacedetector.HomogenizationComparisonFunc(conf,
		haveDataRanks, detectionDatas)
	if len(curTpGroupAbnormalRanks) != 0 {
		*slowCalculateRanks = append(*slowCalculateRanks, curTpGroupAbnormalRanks...)
	}
}

/* 调用时间维度检测算法 */
func useTimeDetection(haveDataRanks []int,
	detectionDatas [][]float64,
	conf config.AlgoInputConfig,
	slowCalculateRanks *[]int) {
	slowRanks := timedetector.DetectionAbnormalCard(haveDataRanks,
		detectionDatas, conf, zpDataColumn)
	if len(slowRanks) != 0 {
		*slowCalculateRanks = append(*slowCalculateRanks, slowRanks...)
	}
}

/* 用于获取当前npus中哪些npu卡是存在数据的 */
func getValidNpusAndData(npus []int, datas map[int][]float64) ([]int, [][]float64) {
	detectionDatas := make([][]float64, 0)
	haveDataRanks := make([]int, 0)
	for _, npuId := range npus {
		data := datas[npuId]
		/* 数据长度为0表示本轮检测增量数据为0，及该卡的comm.csv可能在读取时相较于上一次没有更新 */
		if len(data) == 0 {
			continue
		}
		/* haveDataRanks 对应 detectionDatas */
		detectionDatas = append(detectionDatas, data)
		haveDataRanks = append(haveDataRanks, npuId)
	}
	return haveDataRanks, detectionDatas
}

/* 判断当前job中是否存在tp并行通信域 */
func checkTpParallelExist(detectionGroups [][]int) bool {
	if len(detectionGroups) == 0 {
		hwlog.RunLog.Warn("[SLOWNODE ALGO]Empty detection tp parallel groups!")
		return false
	}
	groups := 0
	for _, detectionGroup := range detectionGroups {
		if len(detectionGroup) <= 1 {
			groups++
		}
	}
	if groups == len(detectionGroups) {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Tp parallel domain not exist!")
		return false
	}
	return true
}

/* 通过tp并行域和当前节点侧任务级卡信息，获取检测组 */
func getDetectionGroups(tpranks [][]int, nodeGlobalRank []int) [][]int {
	// 将 nodeGlobalRank 转换为 map，便于快速查找 判断TP域中的Rank是否在本地，主要是为了处理：将TP设置的较大，出现跨节点的情况
	rankMap := make(map[int]bool)
	for _, rank := range nodeGlobalRank {
		rankMap[rank] = true
	}

	var DetectionGroups [][]int
	/* 并行域文件中不存在tp, 则检测组为所有卡本身 */
	if len(tpranks) == 0 {
		for _, rank := range nodeGlobalRank {
			DetectionGroups = append(DetectionGroups, []int{rank})
		}
		return DetectionGroups
	}
	/* 遍历 tp域 中的每个tp组 */
	for _, subRankList := range tpranks {
		var validRanks []int
		// 检查TP通信域中的每个 rank 是否在 nodeGlobalRank 中
		/* 防止错误的rankID */
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

/* 获取tp通信域中的慢计算卡 */
func getTpSlowCalculateRanks(detectionGroups [][]int,
	conf config.AlgoInputConfig,
	alignedData map[int][]float64) []int {
	if alignedData == nil || len(alignedData) == 0 {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] %s empty aligned ZP_device map data", conf.JobName)
		return nil
	}
	/* 慢计算结果 */
	slowCalculateRanks := make([]int, 0)
	/* 不存在tp并行域情况：检测组么每个组len==1或都为0，则不使用同质化聚类算法, 使用时间维度检测算法 */
	flag := checkTpParallelExist(detectionGroups)
	/* 遍历当前job tp域中的tp组 */
	for _, npuGroup := range detectionGroups {
		/* 用于检测的数据 */
		detectionDatas := make([][]float64, 0)
		haveDataRanks := make([]int, 0)
		/* 从当前job全部的rank数据中取出当前tp检测组中的当前npu算子时延数据中的 "ZP"数据 */
		for _, npuId := range npuGroup {
			data := alignedData[npuId]
			/* 不存在说明该job下该卡没有采集到数据 */
			if len(data) == 0 {
				continue
			}
			/* haveDataRanks 对应 detectionDatas */
			detectionDatas = append(detectionDatas, data)
			haveDataRanks = append(haveDataRanks, npuId)
		}
		/* 判断进行同质化聚类算法检测还是时间维度算法检测 */
		if flag {
			if len(haveDataRanks) < minRanksInGroup {
				continue
			}
			useHomogenizeDetection(haveDataRanks, detectionDatas, conf, &slowCalculateRanks)
		} else {
			useTimeDetection(haveDataRanks, detectionDatas, conf, &slowCalculateRanks)
		}
	}
	return slowCalculateRanks
}

/* 获取tp慢通信域结果 */
func getTpSlowCommunicateDomains(detectionGroups [][]int,
	conf config.AlgoInputConfig,
	alignedData map[int][]float64) [][]int {
	if alignedData == nil || len(alignedData) == 0 {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Empty aligned ZP_device data")
		return nil
	}
	slowCommunicateDomains := make([][]int, 0)
	/* 遍历当前job tp域中的tp组 */
	for _, npuGroup := range detectionGroups {
		/* 卡间不存在tp并行不做tp慢通信域检测 */
		if len(npuGroup) <= 1 {
			continue
		}
		/* 从当前job全部的rank数据中取出当前tp检测组中的npu "ZP_device"数据 */
		haveDataRanks, detectionDatas := getValidNpusAndData(npuGroup, alignedData)
		curTpSlowDomain :=
			timedetector.DetectAbnormalDomain(haveDataRanks, detectionDatas, conf, zpDataColumn)
		if len(curTpSlowDomain) == 0 {
			continue
		}
		slowCommunicateDomains = append(slowCommunicateDomains, curTpSlowDomain)
	}
	return slowCommunicateDomains
}

/* 获取pp通信域慢send卡结果 */
func getPpSlowSendRanks(npus []int,
	conf config.AlgoInputConfig,
	alignedData map[int][]float64) []int {
	if alignedData == nil || len(alignedData) == 0 {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Empty aligned PP_device data")
		return nil
	}
	haveDataRanks, detectionDatas := getValidNpusAndData(npus, alignedData)
	slowSendRanks := timedetector.DetectAbnormalSend(haveDataRanks, detectionDatas, conf, ppDataColumn)
	return slowSendRanks
}

/* 当前节点上当前任务的慢卡结果 */
func getAllSlowHostRanks(npus []int,
	conf config.AlgoInputConfig,
	alignedData map[int][]float64) []int {
	if alignedData == nil || len(alignedData) == 0 {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Empty aligned ZP_host data")
		return nil
	}
	curNodeSlowRanks := make([]int, 0)
	haveDataRanks, detectionDatas := getValidNpusAndData(npus, alignedData)
	curNodeSlowRanks =
		timedetector.DetectAbnormalDomain(haveDataRanks, detectionDatas, conf, zpHostDataColumn)
	return curNodeSlowRanks
}

/* 慢host结果 */
func getSlowHostNode(slowRanks []int) []string {
	slowNodes := make([]string, 0)
	if len(slowRanks) > 0 {
		hostIp, err := config.GetLocalIP()
		if err != nil {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO] %v", err)
			return slowNodes
		}
		slowNodes = append(slowNodes, hostIp)
	}
	return slowNodes
}

// contains 检查元素是否在切片中
func contains(slice []int, value int) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

/* 慢npu IO结果 */
func getSlowIORanks(npus []int,
	conf config.AlgoInputConfig,
	alignedData map[int][]float64,
	slowHostRanks []int) []int {
	if alignedData == nil || len(alignedData) == 0 {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Empty aligned dataloader_host data")
		return nil
	}
	slowIoRanks := make([]int, 0)
	haveDataRanks, detectionDatas := getValidNpusAndData(npus, alignedData)
	slowHostOrIORanks :=
		timedetector.DetectAbnormalSend(haveDataRanks, detectionDatas, conf, dataLoaderDataColumn)
	/* 去除slowHostOrIORanks中已经存在于slowJobRanks中的元素 */
	if len(slowHostRanks) != 0 {
		for _, slowRank := range slowHostOrIORanks {
			if !contains(slowHostRanks, slowRank) {
				slowIoRanks = append(slowIoRanks, slowRank)
			}
		}
	}
	return slowIoRanks
}

/* 获取定界结果 */
func getEnsureBoundaryResult(nodeResult *config.NodeJobResult,
	detectionGroups [][]int,
	conf config.AlgoInputConfig,
	curJobnpus []int) {
	/* 如果不存在劣化结果，则不做定界检测 */
	if nodeResult.IsSlow == 0 {
		return
	}
	/* 获取当前节点上当前job的rank的算子时延数据 */
	jobPath := filepath.Join(conf.FilePath, conf.JobName)
	/* 读文件时全部读取 */
	curJobRanksStepData := getCurJobAllRanksStepData(conf, jobPath, curJobnpus)
	if curJobRanksStepData == nil {
		return
	}
	/* tp通信域慢计算结果(没有tp并行域也检测) */
	nodeResult.SlowCalculateRanks =
		getTpSlowCalculateRanks(detectionGroups, conf, curJobRanksStepData[zpDataColumn])
	/* 节点侧只做slow tp并行域检测，pp可能跨节点，由集群侧做 */
	nodeResult.SlowCommunicationDomains =
		getTpSlowCommunicateDomains(detectionGroups, conf, curJobRanksStepData[zpDataColumn])
	/* 根据PP列数据检测pp慢send卡结果 */
	nodeResult.SlowSendRanks = getPpSlowSendRanks(curJobnpus, conf, curJobRanksStepData[ppDataColumn])
	/* 当前任务（当前节点）所有卡是否都变慢,不是所有卡的话就返回空 */
	slowHostRanks := getAllSlowHostRanks(curJobnpus, conf, curJobRanksStepData[zpHostDataColumn])
	/* 慢host结果:所有卡慢才是host慢 */
	nodeResult.SlowHostNodes = getSlowHostNode(slowHostRanks)
	/* 慢npu IO结果 */
	nodeResult.SlowIORanks = getSlowIORanks(curJobnpus,
		conf, curJobRanksStepData[dataLoaderDataColumn], slowHostRanks)
}

/* 查看当前节点当前任务最小rank是否更新（以此判断是否所有rank更新）*/
func checkRanksUpdate(curjobnpus []int, jobPath string, conf config.AlgoInputConfig) bool {
	if curjobnpus == nil || len(curjobnpus) == 0 {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]empty %s used npus", conf.JobName)
		return false
	}
	sort.Ints(curjobnpus)
	stepTimeFile := filepath.Join(jobPath, strconv.Itoa(curjobnpus[0]), stepTimeFileName)
	if !config.CheckExistDirectoryOrFile(stepTimeFile, false, "node", conf.JobName) {
		return false
	}
	records := readCsvFile(stepTimeFile)
	if records == nil {
		return false
	}
	if len(records) <= 1 {
		hwlog.RunLog.Error("[SLOWNODE ALGO]stepTime file is empty!")
		return false
	}
	if !checkStepTimeFileUpdate(conf, curjobnpus[0], records[len(records)-1], len(records)-1) {
		return false
	}
	return true
}

/* 获取劣化感知结果 */
func getDegradationPerceptionResult(nodeResult *config.NodeJobResult,
	jobPath string,
	curJobnpus []int,
	conf config.AlgoInputConfig) bool {
	/* 检测每一张卡的steptime数据 */
	var maxDegradationLevel float64 = -1
	/* 检查最小的卡的steptime数据是否有更新,作为是否进行检测的依据 */
	if !checkRanksUpdate(curJobnpus, jobPath, conf) {
		return false
	}
	/* 对当前节点上当前job使用的npu卡进行劣化感知:检测当前任务节点上某张卡是否劣化 */
	failedCount := 0
	for _, npuId := range curJobnpus {
		/* 读文件时全部读取，未更新的文件进行记录,并更新最新检测行数，返回增量部分（或可达到数量部分） */
		incrementData := parseStepTimeCsvFile(jobPath, npuId, conf)
		if incrementData == nil {
			failedCount++
			continue
		}
		isSlow, curDegradationLevel := timedetector.FirstNPointsShouldBeNormal(
			incrementData, conf, npuId, stepTimeData)
		if isSlow == 1 && curDegradationLevel > maxDegradationLevel {
			maxDegradationLevel = curDegradationLevel
		}
	}
	if failedCount == len(curJobnpus) {
		return false
	}
	if maxDegradationLevel > 0 {
		nodeResult.IsSlow = 1
		nodeResult.DegradationLevel = fmt.Sprintf("%.2f%%", maxDegradationLevel*config.DegradationPercent)
	}
	return true
}

/* 格式化节点侧任务级检测结果 */
func getFormatDetectionResult(nodeResult config.NodeJobResult, conf config.AlgoInputConfig) string {
	/* get local ip */
	ip, err := config.GetLocalIP()
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] %v", err)
		return ""
	}
	/* 若劣化感知结果为非slow */
	if nodeResult.IsSlow == 0 {
		nodeResult.DegradationLevel = "0.0%"
		nodeResult.SlowHostNodes = []string{}
		nodeResult.SlowIORanks = []int{}
		nodeResult.SlowIORanks = []int{}
		nodeResult.SlowCommunicationDomains = [][]int{}
		nodeResult.SlowSendRanks = []int{}
		nodeResult.SlowCalculateRanks = []int{}
	}
	/* 大key */
	mainKey := "slownode" + "_" + conf.JobName
	minorKey := ip
	nodeResult.JobName = conf.JobName
	nodeResult.NodeRank = ip
	result := make(config.NodeDetectionResult)
	result[mainKey] = map[string]config.NodeJobResult{minorKey: nodeResult}
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] %v", err)
		return ""
	}
	return string(jsonStr)
}

// RegisterNodeLevelCallback 注册回调节点侧回调
func RegisterNodeLevelCallback(callback model.CallbackFunc) {
	callbackFunc = callback
}

/* 节点侧慢节点A3检测算法流程 */
func jobLevelDetectionA3(tpInfo [][]int,
	curNodeJobnpus []int,
	conf config.AlgoInputConfig) {
	/* 初始化本轮检测数据是否已经用于更新历史数据flag */
	nodeleveldatarecorder.SetJobDetectionRecorderAllUpdateFlag(conf.JobName, false)
	/* 初始化本轮检测数据历史使用数据量 */
	nodeleveldatarecorder.SetJobDetectionRecorderAllHistoryDatas(conf.JobName, 0)
	/* 获取tp并行域中的并行检测组 */
	detectionGroups := getDetectionGroups(tpInfo, curNodeJobnpus)
	/* 获取当前Job下所有rank的step数据并进行数据对齐 */
	jobPath := filepath.Join(conf.FilePath, conf.JobName)
	nodeResult := config.NodeJobResult{
		IsSlow: 0,
	}
	/* 劣化感知结果,所有卡更新才检测上报 */
	if !getDegradationPerceptionResult(&nodeResult, jobPath, curNodeJobnpus, conf) {
		return
	}
	/* 定界结果 */
	getEnsureBoundaryResult(&nodeResult, detectionGroups, conf, curNodeJobnpus)
	/* 格式化检测结果 */
	jsonStr := getFormatDetectionResult(nodeResult, conf)
	if len(jsonStr) == 0 {
		return
	}
	/* debug */
	hwlog.RunLog.Infof("[SLOWNODE ALGO]Node detection result:%s", jsonStr)
	/* call callback report */
	if callbackFunc != nil {
		go callbackFunc(jsonStr)
	}
}
