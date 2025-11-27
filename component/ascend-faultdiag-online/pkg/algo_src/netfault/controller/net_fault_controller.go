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

// Package controller
package controller

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-faultdiag-online/pkg/algo_src/netfault/algo"
	"ascend-faultdiag-online/pkg/algo_src/netfault/controllerflags"
	"ascend-faultdiag-online/pkg/algo_src/netfault/policy"
	"ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/fileutils"
)

const (
	maxSuperPodDetectionNums             = 1000
	maxLoopTimes                         = 10000000
	loopInterval                         = 10
	logPrintOutInterval                  = 10
	maxSuperPodId                        = 65535
	pattern                              = `^super-pod-[0-9]+$`
	detectionResult                      = "network_fault.json"
	roceDirName                          = "super-pod-roce"
	roceConfigFile                       = "super-pod-roce.json"
	fileMode                 os.FileMode = 0644
	defaultDetectionInterval             = 15
	decimalTen                           = 10
	decimalLen                           = 64
	readConfFailedReTryNums              = 180
	/* 读取csv文件最大协程并发数量 */
	maxAsyncGoRoutineToReadCsvFiles = 16
	/* 默认一个超节点下的csv文件数量 */
	defaultPerSuperPodCsvFiles = 32
	/* used to represent RoCE plane detection between super nodes */
	roceSuperPodIdConst = -999
)

/* 记录superPod 检测 */
var superPodDetectionRecorder = make(map[string]bool)

/* 并发修改map锁 */
var modifyRecorderSyncLock sync.Mutex

type roceConfig struct {
	RoceSuperPods []int `json:"superPodList"`
}

type detectionParam struct {
	detectionSuperPod     string
	detectionSuperPodPath string
	detectionStartTime    int64
	detectionInterval     int64
	algoObj               *algo.NetDetect
}

/* 并发掉算法接口同步锁 */
var callAlgoInterfaceSyncLock sync.Mutex

/* 将检测标记为false */
func markFalseDetection(superPodAbsolutePath string) {
	modifyRecorderSyncLock.Lock()
	if _, exist := superPodDetectionRecorder[superPodAbsolutePath]; exist {
		superPodDetectionRecorder[superPodAbsolutePath] = false
	}
	modifyRecorderSyncLock.Unlock()
}

func writeNetFaultResult(result []byte, superPodPath string, curDetectionCount int) {
	filePath := filepath.Join(superPodPath, detectionResult)
	/* reload will clean file */
	if _, err := os.Stat(filePath); err == nil && !os.IsNotExist(err) && curDetectionCount == 0 {
		if err = os.Truncate(filePath, 0); err != nil {
			hwlog.RunLog.Errorf("[NETFAULT ALGO]Clean file failed: %v", err)
		}
	}
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileMode)
	if err != nil {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]Error opening file: %v", err)
		return
	}
	defer file.Close()
	// check the symlink
	isSoftlink, err := utils.IsSoftlink(filePath)
	if err != nil {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]check file: %s is softlink or not failed: %v", filePath, err)
		return
	}
	if isSoftlink {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]file: %s is symlink, unsupported", filePath)
		return
	}
	now := time.Now()
	year := now.Year()
	month := now.Month()
	day := now.Day()
	hour := now.Hour()
	minute := now.Minute()
	second := now.Second()
	detectionTimesInfo := fmt.Sprintf("Detection %d result(%d-%d-%d %d:%d:%d):\n",
		curDetectionCount,
		year, month, day, hour, minute, second)
	str := detectionTimesInfo + string(result) + "\n"
	if _, err := file.WriteString(str); err != nil {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]Error writing to file: %v", err)
		return
	}
}

/* 使用循环检测间隔，并判断本次执行检测用时是否超过配置间隔 */
func loopDetectionIntervalCheckSwitch(interval int64,
	detectionStartTime int64, superPodPath string) {
	curTime := time.Now().Unix()
	detectionUsed := curTime - detectionStartTime
	if detectionUsed >= interval {
		return
	}
	curSleep := interval - detectionUsed
	for i := 0; i < int(curSleep); i++ {
		if controllerflags.IsControllerExited.GetState() ||
			!policy.CheckCurSuperPodConfigSwitch(superPodPath) {
			break
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func readWorkRoutine(wg *sync.WaitGroup, csvFileQueue <-chan string,
	input *[]map[string]any, startTime int64) {
	defer wg.Done()
	for path := range csvFileQueue {
		content, err := readCSVFile(path, startTime)
		if err != nil {
			hwlog.RunLog.Warnf("[NETFAULT ALGO]Failed to read csv file(%s): %v", path, err)
			continue
		}
		callAlgoInterfaceSyncLock.Lock()
		*input = algo.MergeAndDeduplicate(*input, content)
		callAlgoInterfaceSyncLock.Unlock()
	}
}

/* 并发读取文件：加快效率 */
func asyncReadCsvFile(files []string, startTime int64) []map[string]any {
	input := make([]map[string]any, 0)
	var wg sync.WaitGroup
	csvFileQueue := make(chan string, len(files))
	for _, file := range files {
		csvFileQueue <- file
	}
	close(csvFileQueue)
	maxAsyncNums := maxAsyncGoRoutineToReadCsvFiles
	if len(files) < maxAsyncGoRoutineToReadCsvFiles {
		maxAsyncNums = len(files)
	}
	for i := 0; i < maxAsyncNums; i++ {
		wg.Add(1)
		go readWorkRoutine(&wg, csvFileQueue, &input, startTime)
	}
	wg.Wait()
	return input
}

/* add switch on detection new detection go routine */
func addNewSuperPodDetection(wg *sync.WaitGroup, superPodPaths []string, superPodIds []int) {
	if len(superPodPaths) == 0 {
		return
	}
	if len(superPodIds) >= maxSuperPodDetectionNums {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]Overload Max super pod detection:%d, %d",
			len(superPodIds), len(superPodPaths))
		return
	}
	for j := 0; j < len(superPodPaths) && j < len(superPodIds); j++ {
		if _, exist := superPodDetectionRecorder[superPodPaths[j]]; !exist &&
			policy.CheckCurSuperPodConfigSwitch(superPodPaths[j]) {
			modifyRecorderSyncLock.Lock()
			superPodDetectionRecorder[superPodPaths[j]] = true
			modifyRecorderSyncLock.Unlock()
			/* 新建协程 */
			wg.Add(1)
			hwlog.RunLog.Infof("[NETFAULT ALGO][Add detection]%s", superPodPaths[j])
			if superPodIds[j] == roceSuperPodIdConst {
				startRoCERoutine(wg)
				continue
			}
			go func(id int, path string) {
				defer wg.Done()
				detectionCurSuperPod(id, path)
			}(superPodIds[j], superPodPaths[j])
		}
	}
}

func startRoCERoutine(wg *sync.WaitGroup) {
	go func(path string) {
		defer wg.Done()
		startRoceDetection(path)
	}(clusterLevelPath)
}

func getFalseFlagDetection(getAll bool) []string {
	ret := make([]string, 0)
	for key, value := range superPodDetectionRecorder {
		if value && !getAll {
			continue
		}
		ret = append(ret, key)
	}
	return ret
}

/* loop check super pod directory config file switch which under cluster directory */
func ifAddNewSuperPodDetection(clusterPath string, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < maxLoopTimes && !controllerflags.IsControllerExited.GetState(); i++ {
		/* 删除为false的超节点检测记录 */
		modifyRecorderSyncLock.Lock()
		recordFalseKey := getFalseFlagDetection(false)
		modifyRecorderSyncLock.Unlock()
		for j := 0; j < len(recordFalseKey); j++ {
			modifyRecorderSyncLock.Lock()
			hwlog.RunLog.Infof("[NETFAULT ALGO][Delete detection]%s", recordFalseKey[j])
			delete(superPodDetectionRecorder, recordFalseKey[j])
			modifyRecorderSyncLock.Unlock()
		}
		superPodIds, superPodPaths := getSuperPodDirInfo(clusterPath)
		if len(superPodPaths) != 0 {
			addNewSuperPodDetection(wg, superPodPaths, superPodIds)
		}
		for j := 0; j < loopInterval && !controllerflags.IsControllerExited.GetState(); j++ {
			time.Sleep(time.Duration(1) * time.Second)
		}
	}
	if controllerflags.IsControllerExited.GetState() {
		recordFalseKey := getFalseFlagDetection(true)
		for _, key := range recordFalseKey {
			modifyRecorderSyncLock.Lock()
			log.Printf("[Delete detection]%s", key)
			delete(superPodDetectionRecorder, key)
			modifyRecorderSyncLock.Unlock()
		}
	}
}

func deletePingListFile(superPodPath string) {
	// 构建文件模式
	pingListPattern := filepath.Join(superPodPath, "ping_list_*")

	// 使用Glob匹配所有符合模式的文件
	files, err := filepath.Glob(pingListPattern)
	if err != nil {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]Error matching files: %v", err)
	}

	// 遍历匹配到的文件并删除
	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			hwlog.RunLog.Errorf("[NETFAULT ALGO]Removing file %s: %v", file, err)
		}
	}
}

/* loop read ping result csv file and call detection */
func loopCsvCallDetection(param detectionParam) {
	hwlog.RunLog.Infof("[NETFAULT ALGO]start super pod %s network detection", param.detectionSuperPod)
	var curDetectionCount int = 0
	for {
		if controllerflags.IsControllerExited.GetState() || curDetectionCount >= maxLoopTimes ||
			!policy.CheckCurSuperPodConfigSwitch(param.detectionSuperPodPath) {
			markFalseDetection(param.detectionSuperPodPath)
			break
		}
		// 获得轮询文件列表
		curTime := time.Now().Unix()
		pingResultFiles, err := findCSVFiles(param.detectionSuperPodPath)
		hwlog.RunLog.Infof("[NETFAULT ALGO]super pod %s csv files(%d):%v", param.detectionSuperPod,
			len(pingResultFiles), pingResultFiles)
		if err != nil {
			hwlog.RunLog.Errorf("[NETFAULT ALGO]Error finding super-pod ping result csv files: %v", err)
		}
		input := asyncReadCsvFile(pingResultFiles, param.detectionStartTime)
		endReadFile := time.Now().Unix()
		hwlog.RunLog.Infof("[NETFAULT ALGO][CONTROLLER]Read csv files duration time: %d", endReadFile-curTime)
		rootCauseAlarmAll := param.algoObj.StartFaultDetect(input)
		endCallAlgo := time.Now().Unix()
		hwlog.RunLog.Infof("[NETFAULT ALGO][CONTROLLER]Call algorithm duration time: %d", endCallAlgo-endReadFile)
		jsonData, err := json.MarshalIndent(rootCauseAlarmAll, "", "  ")
		if err != nil {
			hwlog.RunLog.Errorf("[NETFAULT ALGO]transfer json: %v", err)
			continue
		}
		hwlog.RunLog.Infof("[NETFAULT ALGO]super pod %s net fault detection result: %s",
			param.detectionSuperPod, string(jsonData))
		if callbackFunc != nil {
			go callbackFunc(string(jsonData))
		}
		writeNetFaultResult(jsonData, param.detectionSuperPodPath, curDetectionCount)
		curDetectionCount++
		loopDetectionIntervalCheckSwitch(param.detectionInterval, curTime, param.detectionSuperPodPath)
	}
	/* 删除ping_list文件 */
	deletePingListFile(param.detectionSuperPodPath)
	hwlog.RunLog.Infof("[NETFAULT ALGO]stop %s network detection", param.detectionSuperPodPath)
}

/* return super pod ids and directory paths */
func getSuperPodDirInfo(clusterPath string) ([]int, []string) {
	/* traverse below cluster directory named super-pod-i dir*/
	var superPodIds []int = nil
	var superPodPaths []string = nil
	regexFileName := regexp.MustCompile(pattern)
	fileCount := 0
	err := filepath.Walk(clusterPath, func(path string, info os.FileInfo, err error) error {
		fileCount++
		if fileCount >= constants.MaxFileCount {
			hwlog.RunLog.Errorf("[NETFAULT ALGO]Too many files under: %s, exceed max file count: %d",
				clusterPath, constants.MaxFileCount)
			return errors.New("too many files")
		}
		if err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
		if !info.IsDir() {
			return err
		}
		dirName := filepath.Base(path)
		superPodFilePath := dirName
		matches := regexFileName.FindStringSubmatch(dirName)
		if matches == nil {
			return nil
		}
		/* split super pod id */
		superPodId, err := strconv.Atoi(strings.Split(matches[0], "-")[2])
		if err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
		/* 限制范围0-65535 */
		if superPodId < 0 || superPodId > maxSuperPodId {
			hwlog.RunLog.Warnf("[NETFAULT ALGO]Invalid super-pod-%d.json file(out of range)", superPodId)
			return nil
		}
		superPodIds = append(superPodIds, superPodId)
		superPodPaths = append(superPodPaths, filepath.Join(clusterPath, superPodFilePath))
		return nil
	})
	if err != nil || len(superPodIds) == 0 {
		hwlog.RunLog.Error("[NETFAULT ALGO]No policy id found")
		return nil, nil
	}
	// roce logic
	roceDirPath := filepath.Join(clusterPath, roceDirName)
	if _, err := os.Stat(roceDirPath); err == nil || os.IsNotExist(err) {
		superPodIds = append(superPodIds, roceSuperPodIdConst)
		superPodPaths = append(superPodPaths, roceDirPath)
	}
	return superPodIds, superPodPaths
}

// 检查并启动网络故障检测任务
func startSuperPodsDetectionAsync(clusterPath string) {
	/* async for different super pod directory */
	var wg sync.WaitGroup
	/* 轮询检查各个超节点开关和roce探测开关 */
	wg.Add(1)
	go ifAddNewSuperPodDetection(clusterPath, &wg)
	wg.Wait()
	/* call stop jump out of loop */
	controllerSyncOperatorLock.Lock()
	if controllerExitCond != nil {
		controllerExitCond.Signal()
	}
	controllerSyncOperatorLock.Unlock()
	hwlog.RunLog.Info("[NETFAULT ALGO]net fault detection complete!")
}

/* 在检查文件之后使用 */
func checkDiffConfig(superPodFilePath string) map[string]any {
	confFilePath := filepath.Join(superPodFilePath, configFile)
	fileContent, err := fileutils.ReadLimitBytes(confFilePath, constants.Size10M)
	if err != nil {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]error: %v", err)
		return nil
	}
	// 定义一个map来保存解析结果
	targetKeys := []string{"networkType", "pingType", "pingTimes", "pingInterval", "suppressedPeriod", "period"}
	callAlgorithmParam := policy.ReadConfigFromFile(fileContent, targetKeys)

	return callAlgorithmParam
}

func getCurSuperPodDetectionInterval(conf map[string]any, superPodFilePath string) int {
	/* 获取检测间隔 */
	var detectionInterval int
	interval, exist := conf["period"]
	if !exist {
		hwlog.RunLog.Warnf("[NETFAULT ALGO]%s config without period field", superPodFilePath)
		detectionInterval = defaultDetectionInterval
	} else {
		intervalDigit, ok := interval.(int)
		if !ok || intervalDigit <= 0 {
			hwlog.RunLog.Warnf("[NETFAULT ALGO]%s config period field format error", superPodFilePath)
			detectionInterval = defaultDetectionInterval
		} else {
			detectionInterval = intervalDigit
		}
	}
	return detectionInterval
}

/* 循环等待超节点目录生成和配置文件生成,并检查开关 superPodDirPath绝对路径。（needConfig为配置开关文件） */
func loopWaitSuperPodDirAndCheckConfigFile(
	superPodDirPath string,
	jsonFile string,
	configFile string,
	needConfig bool,
) bool {
	/* 默认开启超节点间roce探测，文件夹不存在轮询等待 */
	for i := 0; i < readConfFailedReTryNums &&
		!controllerflags.IsControllerExited.GetState(); i++ {
		if needConfig && policy.CheckCurSuperPodConfigSwitch(superPodDirPath) {
			hwlog.RunLog.Infof("[NetFault Algo]%s network detection switch off", superPodDirPath)
			return false
		}
		/* path */
		_, err1 := os.Stat(superPodDirPath)
		/* cathelper.conf */
		_, err2 := os.Stat(filepath.Join(superPodDirPath, configFile))
		if !needConfig {
			err2 = nil
		}
		/* super-pod-x.json */
		_, err3 := os.Stat(filepath.Join(superPodDirPath, jsonFile))
		/* 错误具体情况不管 */
		if err1 != nil && os.IsNotExist(err1) ||
			err2 != nil && os.IsNotExist(err2) ||
			err3 != nil && os.IsNotExist(err3) {
			if i == readConfFailedReTryNums-1 {
				hwlog.RunLog.Errorf("[NETFAULT ALGO]roce %s detection retry max time failed(%s)!",
					superPodDirPath, jsonFile)
				return false
			}
			/* 控制日志输出数量 */
			if i%logPrintOutInterval == 0 {
				hwlog.RunLog.Infof("[NETFAULT ALGO]roce %s detection(%s) retry:%d",
					superPodDirPath, jsonFile, i+1)
			}
			time.Sleep(time.Duration(1) * time.Second)
			continue
		}
		break
	}
	/* 总体开关检查(超节点间已去除每个超节点开关检查) */
	if controllerflags.IsControllerExited.GetState() {
		hwlog.RunLog.Info("[NETFAULT ALGO]network detection off")
		return false
	}
	return true
}

func detectionCurSuperPod(superPodId int, superPodFilePath string) {
	if !loopWaitSuperPodDirAndCheckConfigFile(
		superPodFilePath,
		fmt.Sprintf("super-pod-%d.json", superPodId),
		configFile,
		true) {
		markFalseDetection(superPodFilePath)
		return
	}
	callAlgorithmParam := checkDiffConfig(superPodFilePath)
	if callAlgorithmParam == nil {
		markFalseDetection(superPodFilePath)
		return
	}
	callAlgorithmParam[policy.ServerIdMap] = make(map[string]string)
	/* 适配算法 */
	callAlgorithmParam["superPodJobFlag"] = false
	callAlgorithmParam["superPodArr"] = []string{}
	err := policy.SetCallAlgorithmParamInfo(superPodId, superPodFilePath, callAlgorithmParam)
	if err != nil {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]set call algorithm param info failed: %v", err)
		markFalseDetection(superPodFilePath)
		return
	}

	interval := getCurSuperPodDetectionInterval(callAlgorithmParam, superPodFilePath)
	flag, npuInfoMap := policy.GetTargetSuperPodNpuMap(superPodFilePath, superPodId)
	if !flag {
		markFalseDetection(superPodFilePath)
		return
	}
	/* set algorithm current supper pod cards info */
	superPodIdStr := strconv.Itoa(superPodId)
	detectObj := algo.NewNetDetect(superPodIdStr)
	detectObj.SetFaultDetectParam(callAlgorithmParam, npuInfoMap)
	/* generate current super pod server level pinglist */
	if !policy.GenSuperPodServersPingList(superPodFilePath, detectObj) {
		markFalseDetection(superPodFilePath)
		return
	}
	milliTimestamp := time.Now().UnixMilli()
	hwlog.RunLog.Infof("[NETFAULT ALGO]Super pod %d detection started timestamp: %d", superPodId, milliTimestamp)
	param := detectionParam{
		detectionInterval:     int64(interval),
		detectionSuperPodPath: superPodFilePath,
		detectionStartTime:    milliTimestamp,
		detectionSuperPod:     superPodIdStr,
		algoObj:               detectObj}
	loopCsvCallDetection(param)
}

// 在给定的超节点目录中查找所有 CSV 文件，并返回它们的完整路径
func findCSVFiles(dir string) ([]string, error) {
	csvFiles := make([]string, 0, defaultPerSuperPodCsvFiles)
	// 使用 filepath.Walk 遍历目录
	fileCount := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		fileCount++
		if fileCount >= constants.MaxFileCount {
			return fmt.Errorf("too many files under: %s, exceed max file count: %d", path, constants.MaxFileCount)
		}
		if err != nil {
			return err
		}
		if !info.IsDir() && (filepath.Ext(info.Name()) == ".csv" || filepath.Ext(info.Name()) == ".csv-bak") {
			csvFiles = append(csvFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return csvFiles, nil
}

// 读取csv文件内容，返回一个map[string]string类型的数组
func readCSVFile(filePath string, startTime int64) ([]map[string]any, error) {
	fileContent, err := fileutils.ReadLimitBytes(filePath, constants.Size10M)
	if err != nil {
		hwlog.RunLog.Warnf("[NETFAULT ALGO]failed to open CSV file: %v", err)
		return nil, err
	}

	reader := csv.NewReader(bytes.NewReader(fileContent))
	records, err := reader.ReadAll()
	if err != nil {
		hwlog.RunLog.Warnf("[NETFAULT ALGO]failed to read CSV file: %v", err)
		return nil, err
	}

	if len(records) == 0 {
		hwlog.RunLog.Error("[NETFAULT ALGO]CSV file is empty")
		return nil, errors.New("CSV file is empty")
	}
	/* 标题行 */
	headers := records[0]
	/* 数据行 */
	data := make([]map[string]any, 0, len(records))
	for _, record := range records[1:] {
		/* 判断时间戳 */
		timeStampStr := record[len(record)-1]
		timeStamp, err := strconv.ParseInt(timeStampStr, decimalTen, decimalLen)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		if timeStamp < startTime {
			continue
		}
		row := make(map[string]any, len(record))
		for idx, value := range record {
			if idx < len(headers) {
				row[headers[idx]] = value
			}
		}
		data = append(data, row)
	}
	return data, nil
}

func makeAlgoRockParam(clusterPath string, superPodNames []string) map[string]any {
	confFile := filepath.Join(clusterPath, roceDirName)
	paramMap := checkDiffConfig(confFile)
	if paramMap == nil {
		return nil
	}
	// the parameters for other algo
	paramMap[policy.ServerIdMap] = make(map[string]string)
	paramMap["axisStrategy"] = "same_axis"
	paramMap["superPodArr"] = superPodNames
	paramMap["superPodJobFlag"] = true
	paramMap["npu_type"] = "A5"
	// 0 stands for ip
	paramMap["pingObjType"] = 0
	return paramMap
}

func getRoceDetectionSuperPodFiles(clusterPath, roceDirPath string) ([]string, []string) {
	// parse the detection config file
	date, err := fileutils.ReadLimitBytes(filepath.Join(roceDirPath, roceConfigFile), constants.Size10M)
	if err != nil {
		hwlog.RunLog.Errorf("[NetFault Algo]Read roce config file failed: %v", err)
		return nil, nil
	}
	var roceConf roceConfig
	if err = json.Unmarshal(date, &roceConf); err != nil || len(roceConf.RoceSuperPods) == 0 {
		hwlog.RunLog.Errorf("[NetFault Algo]Invalid roce detection info: %v or invalid config data", roceConf)
		return nil, nil
	}
	// remove duplicate super pod ids
	uniqueNums := make(map[int]bool, len(roceConf.RoceSuperPods))
	uniqueNumsArr := make([]int, 0, len(roceConf.RoceSuperPods))
	for _, num := range roceConf.RoceSuperPods {
		if !uniqueNums[num] {
			uniqueNums[num] = true
			uniqueNumsArr = append(uniqueNumsArr, num)
		}
	}
	roceConf.RoceSuperPods = uniqueNumsArr
	roceSuperPodInfoFiles := make([]string, len(roceConf.RoceSuperPods))
	superPodNames := make([]string, len(roceConf.RoceSuperPods))
	// sort by ascending order
	sort.Ints(roceConf.RoceSuperPods)
	for i, id := range roceConf.RoceSuperPods {
		superPodDirName := fmt.Sprintf("super-pod-%d", id)
		algoSuperPodName := fmt.Sprintf("SuperPod-%d", id)
		superPodPath := filepath.Join(clusterPath, superPodDirName)
		jsonFile := fmt.Sprintf("super-pod-%d.json", id)
		superPodJsonInfo := filepath.Join(clusterPath, superPodDirName, jsonFile)
		// sync waiting till the super pod dir and json file ready
		if !loopWaitSuperPodDirAndCheckConfigFile(superPodPath, jsonFile, configFile, false) {
			return nil, nil
		}
		superPodNames[i] = algoSuperPodName
		roceSuperPodInfoFiles[i] = superPodJsonInfo
	}
	return superPodNames, roceSuperPodInfoFiles
}

// startRoceDetection start roce of super pods detection
func startRoceDetection(clusterPath string) {
	roceDirPath := filepath.Join(clusterPath, roceDirName)
	// loop for waiting roce dir and config file
	if !loopWaitSuperPodDirAndCheckConfigFile(roceDirPath, roceConfigFile, configFile, true) {
		markFalseDetection(roceDirPath)
		return
	}
	// wait all the required files ready, return the super pod path
	superPodNames, roceSuperPodInfoFiles := getRoceDetectionSuperPodFiles(clusterPath, roceDirPath)
	if superPodNames == nil || roceSuperPodInfoFiles == nil {
		markFalseDetection(roceDirPath)
		return
	}

	// get npu map and link info of super pod roce
	npuMap, linkPaths := policy.GetSuperPodsRoceNpuInfo(roceSuperPodInfoFiles)
	if npuMap == nil || linkPaths == nil {
		markFalseDetection(roceDirPath)
		return
	}
	algoParam := makeAlgoRockParam(clusterPath, superPodNames)
	// empty string needs for the detection of super pod roce
	detectObj := algo.NewNetDetect("roce")
	detectObj.SetFaultDetectParam(algoParam, npuMap)
	// generate the ping list for super pod roce detection
	if !policy.GenRoceSuperPodLevelPingList(filepath.Join(clusterPath, roceDirName), detectObj, linkPaths, npuMap) {
		markFalseDetection(roceDirPath)
		return
	}
	// detect interval
	interval := getCurSuperPodDetectionInterval(algoParam, filepath.Join(clusterPath, roceDirName))
	milliTimestamp := time.Now().UnixMilli()
	hwlog.RunLog.Infof("[NetFault Algo]Super pod roce detetion started timestamps: %d", milliTimestamp)
	param := detectionParam{
		detectionInterval:     int64(interval),
		detectionSuperPodPath: roceDirPath,
		detectionStartTime:    milliTimestamp,
		detectionSuperPod:     "roce",
		algoObj:               detectObj,
	}
	loopCsvCallDetection(param)
}
