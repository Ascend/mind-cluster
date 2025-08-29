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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/nodeleveldatarecorder"
)

/* 对齐多个切片数据 */
func alignData(localDataSlices map[int][]float64) map[int][]float64 {
	// 假设所有切片的长度相同，基于第一个切片的长度进行对齐
	if len(localDataSlices) == 0 {
		return nil
	}
	/* 获取所有切片中最小长度 */
	minLength := -1
	for _, ppData := range localDataSlices {
		if minLength == -1 {
			minLength = len(ppData)
			continue
		}
		if len(ppData) < minLength {
			minLength = len(ppData)
		}
	}
	/* 对齐各个切片数据量 */
	alignedData := make(map[int][]float64)
	for npuId, ppData := range localDataSlices {
		alignedData[npuId] = append(alignedData[npuId], ppData[:minLength]...)
	}
	return alignedData
}

func parseCsvDataToType(records [][]string, columnIndex int) ([]float64, []int, error) {
	var validData = []float64{}
	var steps = []int{}
	/* 跳过标题行 */
	for _, record := range records[1:] {
		if columnIndex < 0 || columnIndex >= len(record) || len(record) == 0 {
			return nil, nil, fmt.Errorf("record column index out of range")
		}
		value, err := strconv.ParseFloat(record[columnIndex], byteLength)
		if err != nil {
			return nil, nil, fmt.Errorf("[SLOWNODE ALGO]transfer float64 failed:%v", err)
		}
		step, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, nil, fmt.Errorf("[SLOWNODE ALGO]transfer int failed:%d", err)
		}
		// 有效值判断
		if value == math.MaxFloat64 || step == math.MinInt32 || math.IsInf(value, 0) ||
			value == 0 {
			continue
		}
		validData = append(validData, value)
		/* 保存到最新的step */
		steps = append(steps, step)
	}
	return validData, steps, nil
}

/* 读取comm.csv文件中指定列数据 */
func readTargetColumnFromCsv(filePath string, columnName string) ([]float64, []int, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err = file.Close(); err != nil {
			// 处理关闭文件时的错误
			hwlog.RunLog.Errorf("[SLOWNODE ALGO] %v", err)
		}
	}()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(records) <= 1 {
		return nil, nil, fmt.Errorf("[SLOWNODE ALGO]CSV file %s is empty", filePath)
	}
	headers := records[0]
	// 查找指定的列名对应的索引
	var columnIndex int
	columnFound := false
	for i, header := range headers {
		if strings.ToLower(strings.ReplaceAll(header, "_", "")) ==
			strings.ToLower(strings.ReplaceAll(columnName, "_", "")) {
			columnIndex = i
			columnFound = true
			break
		}
	}
	// 如果未找到列名，返回错误
	if !columnFound {
		return nil, nil, fmt.Errorf("[SLOWNODE ALGO]Column %s not found in file %s", columnName, filePath)
	}

	return parseCsvDataToType(records, columnIndex)
}

func getDetectionData(curDetectedStepIndex int,
	allDatas []float64,
	conf config.AlgoInputConfig,
	npuId int,
	column string) []float64 {
	returnData := make([]float64, 0)
	/* 未被覆盖取增量部分， 增量不够用历史数据 */
	if curDetectedStepIndex > len(allDatas)-1 {
		return returnData
	}
	increment := allDatas[curDetectedStepIndex+1:]
	historyData := conf.NconsecAnomaliesSignifySlow - len(increment)
	if historyData > 0 {
		nodeleveldatarecorder.SetJobDetectionRecorderHistoryData(conf.JobName, npuId, column, historyData)
		for i := 0; i < historyData && (curDetectedStepIndex-i >= 0); i++ {
			returnData = append(returnData, allDatas[curDetectedStepIndex-i])
		}
		/* 历史数据量不够 */
		if len(returnData) < historyData {
			hwlog.RunLog.Warnf("[SLOWNODE ALGO] %s, %d history data not enough", conf.JobName, npuId)
			return []float64{}
		}
	}
	returnData = append(returnData, increment...)
	return returnData
}

/* 仅获取增量部分(无增量取历史)，steptime.csv更新，comm.csv不判断更新 */
func updateCalculatorLatencyStepRecord(allDatas []float64,
	allSteps []int,
	npuId int,
	column string,
	conf config.AlgoInputConfig) []float64 {
	if len(allDatas) == 0 || len(allSteps) == 0 {
		return []float64{}
	}
	detectedMaxStep := nodeleveldatarecorder.GetJobDetectionRecorderMaxDetectedStep(conf.JobName, npuId, column)
	curDetectedStepIndex := -1
	/* 第一次检测 */
	if detectedMaxStep == 0 {
		if len(allDatas) < (conf.NormalNumber + conf.NconsecAnomaliesSignifySlow) {
			hwlog.RunLog.Warnf("[SLOWNODE ALGO] %s,  %d calculator data not enough", conf.JobName, npuId)
			return []float64{}
		}
		nodeleveldatarecorder.SetJobDetectionRecorderMaxDetectedStep(
			conf.JobName, npuId, allSteps[len(allSteps)-1], column)
		return allDatas
	}
	/* 非第一次检测 */
	for i := 0; i < len(allSteps); i++ {
		if allSteps[i] == detectedMaxStep {
			curDetectedStepIndex = i
		}
	}
	/* 全覆盖为全新数据（不考虑中间截断情况） */
	if curDetectedStepIndex == -1 &&
		len(allDatas) >= conf.NconsecAnomaliesSignifySlow {
		nodeleveldatarecorder.SetJobDetectionRecorderMaxDetectedStep(
			conf.JobName, npuId, allSteps[len(allSteps)-1], column)
		return allDatas
	}
	if curDetectedStepIndex == -1 && len(allDatas) < conf.NconsecAnomaliesSignifySlow {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO] %s,  %d calculator data not enough", conf.JobName, npuId)
		return []float64{}
	}
	returnData := getDetectionData(curDetectedStepIndex, allDatas, conf, npuId, column)
	nodeleveldatarecorder.SetJobDetectionRecorderMaxDetectedStep(
		conf.JobName, npuId, allSteps[len(allSteps)-1], column)
	return returnData
}

/* 按npus顺序获取当前Job下所有rank下zp\pp...时延数据并进行对齐 */
func getCurJobAllRanksStepData(conf config.AlgoInputConfig, jobPath string, npus []int) map[string]map[int][]float64 {
	allRanksStepData := make(map[string]map[int][]float64)
	/* 防止越界访问 */
	allRanksStepData[zpDataColumn] = make(map[int][]float64)
	allRanksStepData[dataLoaderDataColumn] = make(map[int][]float64)
	allRanksStepData[ppDataColumn] = make(map[int][]float64)
	allRanksStepData[zpHostDataColumn] = make(map[int][]float64)
	for _, npuId := range npus { // npu ID集群唯一标识
		rankPath := filepath.Join(jobPath, strconv.Itoa(npuId), nodeLevelNpuLatencyDataFile)
		/* get ZP */
		zpData, steps, err := readTargetColumnFromCsv(rankPath, zpDataColumn)
		if err != nil {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO] jobPath: %s, error: %v", jobPath, err)
			continue
		}
		allRanksStepData[zpDataColumn][npuId] =
			updateCalculatorLatencyStepRecord(zpData, steps, npuId, zpDataColumn, conf)
		/* get PP */
		ppData, steps, err := readTargetColumnFromCsv(rankPath, ppDataColumn)
		if err != nil {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO] jobPath: %s, error: %v", jobPath, err)
			continue
		}
		allRanksStepData[ppDataColumn][npuId] =
			updateCalculatorLatencyStepRecord(ppData, steps, npuId, ppDataColumn, conf)
		/* get ZP_host */
		zpHostData, steps, err := readTargetColumnFromCsv(rankPath, zpHostDataColumn)
		if err != nil {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO] jobPath: %s, error: %v", jobPath, err)
			continue
		}
		allRanksStepData[zpHostDataColumn][npuId] =
			updateCalculatorLatencyStepRecord(zpHostData, steps, npuId, zpHostDataColumn, conf)
		/* get dataloader_host */
		dataLoaderData, steps, err := readTargetColumnFromCsv(rankPath, dataLoaderDataColumn)
		if err != nil {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO] jobPath: %s, error: %v", jobPath, err)
			continue
		}
		allRanksStepData[dataLoaderDataColumn][npuId] =
			updateCalculatorLatencyStepRecord(dataLoaderData, steps, npuId, dataLoaderDataColumn, conf)
	}
	/* 当前仅将ZP开头的列数据进行对齐 */
	allRanksStepData[zpDataColumn] = alignData(allRanksStepData[zpDataColumn])
	allRanksStepData[zpHostDataColumn] = alignData(allRanksStepData[zpHostDataColumn])
	return allRanksStepData
}

/* 获取当前npu卡的topo信息 */
func getCurRankTopoInfoA3(rankPath string) map[string]any {
	file := filepath.Join(rankPath, rankTopofileName)
	data, err := utils.LoadFile(file)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] error: %v", err)
		return nil
	}
	var rankJson map[string]any
	err = json.Unmarshal(data, &rankJson)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] error: %v", err)
		return nil
	}
	return rankJson
}

/* 检查当前rank的并行域信息是否已经通过其他同一个并行域中的rank并行域信息添加 */
func checkRankParallelExist(parallelInfo *map[int]map[int]bool, rankId int, npuGroup []int) bool {
	/* 检查并行域中是否已经存在当前并行域组 */
	for _, parallelDomainGroups := range *parallelInfo {
		/* 存在必为true */
		_, exist := parallelDomainGroups[rankId]
		if exist {
			return true
		}
	}
	/* 添加到当前并行域信息中 */
	_, exist := (*parallelInfo)[rankId]
	if !exist {
		(*parallelInfo)[rankId] = make(map[int]bool)
	}
	for _, npu := range npuGroup {
		(*parallelInfo)[rankId][npu] = true
	}

	return false
}

/* 添加tp并行域 */
func addTpGroupToArray(npuGroup []any,
	tpGroups *[][]int,
	rankId int,
	tpParallelInfo *map[int]map[int]bool) {
	/* json中数字将解析为float64类型 */
	npuParallel := config.TransferFloatArrayToInt(npuGroup)
	if npuParallel == nil {
		return
	}
	/* 如果在同一个并行域中则不需要再添加 */
	if !checkRankParallelExist(tpParallelInfo, rankId, npuParallel) {
		*tpGroups = append(*tpGroups, npuParallel)
	}
}

/* 节点侧任务级topo中整合tp并行域信息 */
func getParallelGroupInfoA3(rankId int,
	rankInfo map[string]any,
	tpGroups *[][]int,
	tpParallelInfo *map[int]map[int]bool) bool {
	/* key值为group_name_xx */
	for _, parallelInfo := range rankInfo {
		parallelMap, ok := parallelInfo.(map[string]any)
		if !ok {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO]Invalid rank %d parallel domain info!", rankId)
			return false
		}
		/* 并行域类型 */
		groupName, exist := parallelMap[dataFIleFieldGroupName]
		if !exist {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO]Rank %d parallel domain without goup_name!", rankId)
			return false
		}
		name, ok := groupName.(string)
		/* 节点侧仅检测tp并行域信息 */
		if !ok || name != tpParallelDomainName {
			continue
		}
		/* 并行域npu卡信息 */
		parallelNpus, exist := parallelMap[dataFIleFieldGlobalRanks]
		if !exist {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO]Rank %d parallel domain without global_ranks!", rankId)
			return false
		}
		npuGroup, ok := parallelNpus.([]any)
		if !ok {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO]Invalid rank %d parallel domain global_ranks:%v!",
				rankId, parallelNpus)
			return false
		}
		addTpGroupToArray(npuGroup, tpGroups, rankId, tpParallelInfo)
	}
	return true
}

/* 获取当前JOB TP并行域信息：PP可能存在跨节点，放在集群侧进行检测 */
func getJobTpParallelsInfoA3(curJobRanksTopo map[int]any) [][]int {
	if len(curJobRanksTopo) == 0 {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]Invalid job rank info!")
		return nil
	}
	/* 辅助用于避免重复添加并行域信息 */
	tpParallelInfo := make(map[int]map[int]bool, 0)
	tpGroups := make([][]int, 0)
	for rankId, topoInfo := range curJobRanksTopo {
		rankInfo, ok := topoInfo.(map[string]any)
		if !ok {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO]Invalid rank %d info!", rankId)
			return nil
		}
		if !getParallelGroupInfoA3(rankId, rankInfo, &tpGroups, &tpParallelInfo) {
			continue
		}
	}
	hwlog.RunLog.Infof("[SLOWNODE ALGO]TP Parallel info: %v", tpGroups)
	return tpGroups
}

/* 获取当前任务节点侧所有rank的parallel domain信息 */
func getCurJobAllRanksTopo(jobPath string, ranks []string) (map[int]any, []int) {
	curJobRanksTopo := make(map[int]any, 0)
	validRanks := make([]int, 0)
	for i := 0; i < len(ranks); i++ {
		/* 若当前rank级目录不存在则跳过 */
		_, err := os.Stat(filepath.Join(jobPath, ranks[i]))
		if err != nil {
			continue
		}
		/* 获取当前Rank topo信息 */
		curRankTopo := getCurRankTopoInfoA3(filepath.Join(jobPath, ranks[i]))
		if curRankTopo != nil {
			id, err := strconv.Atoi(ranks[i])
			if err != nil {
				hwlog.RunLog.Errorf("[SLOWNODE ALGO]error rank ID: %s", ranks[i])
				continue
			}
			/* 有效值判断 */
			if id < 0 || id >= math.MaxInt32 {
				continue
			}
			curJobRanksTopo[id] = curRankTopo
			validRanks = append(validRanks, id)
		}
	}
	if len(validRanks) == 0 {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] %v no valid rank!", ranks)
		return nil, nil
	}
	return curJobRanksTopo, validRanks
}

/* 获取当前job npu信息、tp并行域信息并判断npu之间是否存在tp并行域 */
func isCurJobEnableDetectionA3(jobPath string, ranks []string) ([][]int, []int) {
	if len(ranks) == 0 {
		hwlog.RunLog.Error("[SLOWNODE ALGO]empty job ranks!")
		return nil, nil
	}
	curJobRanksTopo, validRanks := getCurJobAllRanksTopo(jobPath, ranks)
	if curJobRanksTopo == nil || validRanks == nil {
		return nil, nil
	}
	/* 整合当前job tp并行域信息 */
	tpParallel := getJobTpParallelsInfoA3(curJobRanksTopo)
	/* tp并行域在文件中不存在，认为是仅单张卡为一个组，发生错误才返回nil */
	if tpParallel == nil {
		return nil, nil
	}
	/* 对rank进行排序，便于后续寻找最小rank */
	sort.Ints(validRanks)
	return tpParallel, validRanks
}

/* 仅用于读取stepTime.csv文件 */
func readCsvFile(filePath string) [][]string {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]failed to open file:%s", filePath)
		return nil
	}
	defer func() {
		if err = file.Close(); err != nil {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO] %v", err)
		}
	}()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]failed to read CSV file: %v", err)
		return nil
	}
	return records
}

/* 检查steptime.csv file是否有更新 */
func checkStepTimeFileUpdate(conf config.AlgoInputConfig, npuId int, tail []string, dataLen int) bool {
	/* 获取历史检测最大的step */
	detectedMaxStep :=
		nodeleveldatarecorder.GetJobDetectionRecorderMaxDetectedStep(conf.JobName, npuId, stepTimeData)
	/* 末尾step判断 */
	if len(tail) <= 0 {
		hwlog.RunLog.Error("[SLOWNODE ALGO]empty last line")
		return false
	}
	end, err := strconv.Atoi(tail[0])
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]parse step: %v", err)
		return false
	}
	/* 第一次检测情况，数据数量是否足够 */
	if detectedMaxStep == 0 && (dataLen < conf.NormalNumber+conf.NconsecAnomaliesSignifySlow) {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]rank %v first detection data is not enough!", npuId)
		return false
	}
	/* 一行数据时不考虑检测step==0==start==end==detectedMaxStep */
	if detectedMaxStep < end {
		return true
	}
	/* 文件未更新 */
	nodeleveldatarecorder.AddJobDetectionContinuousNotUpdateTimes(conf.JobName, npuId, stepTimeData)
	times := nodeleveldatarecorder.GetJobDetectionContinuousNotUpdateTimes(conf.JobName, npuId, stepTimeData)
	if times >= maxContinuousNotUpdate {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] %s npu: %v %s is not update or training complete or blocked!",
			conf.JobName, npuId, stepTimeFileName)
	} else {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO] %s npu: %v %s no increment!(%d)",
			conf.JobName, npuId, stepTimeFileName, detectedMaxStep)
	}
	return false
}

func getStepTimeIncrementData(records [][]string, detectedStep int) ([]float64, int) {
	incrementData := make([]float64, 0)
	curDetectedStepIndex := -1
	for index, record := range records[1:] {
		step, err := strconv.Atoi(record[0])
		if err != nil {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO] %v", err)
			return nil, -1
		}
		/* 用于有增量但不够时 */
		if step == detectedStep {
			curDetectedStepIndex = index
		}
		/* 跳过已检测step */
		if step <= detectedStep && detectedStep != 0 {
			continue
		}
		if len(record) < stepTimeFileMinColumns {
			hwlog.RunLog.Error("[SLOWNODE ALGO]invalid CSV format: missing columns")
			return nil, -1
		}
		duration, err := strconv.ParseFloat(record[1], byteLength)
		if err != nil {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO] %v", err)
			return nil, -1
		}
		/* 有效值判断 */
		if duration == 0 || duration == math.MaxFloat64 || math.IsInf(duration, 0) {
			continue
		}
		incrementData = append(incrementData, duration)
	}
	return incrementData, curDetectedStepIndex
}

func getStepTimeHistoryData(records [][]string,
	lenIncrement int,
	conf config.AlgoInputConfig,
	npuId int,
	curDetectedStepIndex int) []float64 {
	var historyData []float64 = nil
	/* 覆盖但数量不够情况 */
	if !(curDetectedStepIndex == -1) && (len(records)-1 < conf.NconsecAnomaliesSignifySlow) {
		return []float64{}
	}
	/* 未覆盖，有更新，但未找到之前记录的最新的step, 且数据数量足够检测 */
	neededData := conf.NconsecAnomaliesSignifySlow - lenIncrement
	if neededData <= 0 {
		return []float64{}
	}
	var startIndex int
	if curDetectedStepIndex == -1 {
		startIndex = len(records) - 1 - neededData
	} else {
		startIndex = curDetectedStepIndex
	}
	nodeleveldatarecorder.SetJobDetectionRecorderHistoryData(conf.JobName, npuId, stepTimeData, neededData)
	for i := startIndex + 1; i >= 1 && neededData > 0; i-- {
		if i >= len(records) {
			break
		}
		/* 回退有效值判断 */
		duration, err := strconv.ParseFloat(records[i][1], byteLength)
		if err != nil {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO]failed to parse durationtime as int64 : %v", err)
			return nil
		}
		/* 有效值判断 */
		if duration == 0 || math.IsInf(duration, 1) {
			continue
		}
		historyData = append(historyData, duration)
		neededData--
	}
	/* 历史有效数据不够 */
	if len(historyData) < conf.NconsecAnomaliesSignifySlow-lenIncrement {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO] %s npu: %v %v valid history data is not enough!",
			conf.JobName, npuId, stepTimeData)
		return nil
	}
	return historyData
}

/* 读取 steptime.csv 文件中的 durationtime 列并返回 []int64 类型的切片 */
func parseStepTimeCsvFile(jobPath string, npuId int, conf config.AlgoInputConfig) []float64 {
	stepTimeFile := filepath.Join(jobPath, strconv.Itoa(npuId), stepTimeFileName)
	if !config.CheckExistDirectoryOrFile(stepTimeFile, false, "node", conf.JobName) {
		return nil
	}
	/* 获取已检测的step序号 */
	detectedStep := nodeleveldatarecorder.GetJobDetectionRecorderMaxDetectedStep(conf.JobName, npuId, stepTimeData)
	records := readCsvFile(stepTimeFile)
	if records == nil {
		return nil
	}
	if len(records) <= conf.NconsecAnomaliesSignifySlow {
		hwlog.RunLog.Error("[SLOWNODE ALGO]stepTime file data is not enough!")
		return nil
	}
	/* 判断当前卡的steptime数据是否有更新 */
	if !checkStepTimeFileUpdate(conf, npuId, records[len(records)-1], len(records)-1) {
		return nil
	}
	maxStep, err := strconv.Atoi(records[len(records)-1][0])
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]parse step: %v", err)
		return nil
	}
	nodeleveldatarecorder.CleanJobDetectionContinuousNotUpdateTimes(conf.JobName, npuId, stepTimeData)
	incrementData, curDetectedStepIndex := getStepTimeIncrementData(records, detectedStep)
	if incrementData == nil {
		return nil
	}
	/* 判断是否是第一次检测，有效数据量是否足够，或非第一次数据量不够 */
	if detectedStep == 0 && (len(incrementData) < (conf.NconsecAnomaliesSignifySlow+conf.NormalNumber)) ||
		len(records)-1 < conf.NconsecAnomaliesSignifySlow {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO] %s npu: %v %v valid data is not enough!",
			conf.JobName, npuId, stepTimeData)
		return nil
	}
	historyData := getStepTimeHistoryData(records, len(incrementData), conf, npuId, curDetectedStepIndex)
	if historyData == nil {
		return nil
	}
	var returnData []float64
	returnData = append(returnData, historyData...)
	returnData = append(returnData, incrementData...)
	/* 更新已检测最大step */
	nodeleveldatarecorder.SetJobDetectionRecorderMaxDetectedStep(conf.JobName, npuId, maxStep, stepTimeData)
	return returnData
}
