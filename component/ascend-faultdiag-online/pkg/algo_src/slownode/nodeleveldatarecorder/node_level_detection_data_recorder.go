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

// Package nodeleveldatarecorder is used for recording node level detection data.
package nodeleveldatarecorder

import (
	"sync"

	"ascend-common/common-utils/hwlog"
)

/* 记录当前节点级检测任务卡级算子时延数据中每种数据的均值、标准差、已检测数据行信息 */
type dataRecord struct {
	mean   *float64
	stdDev *float64
	/* 对应均值和标准差的数据计数 */
	healthDataCount *int
	/* 已检测的最大stepId */
	maxDetectedStep *int
	/* 本轮检测数据已经利用更新过mean和stdDev了，避免重复利用同一组数据进行更新 */
	curRoundUpdated *bool
	/* 本轮检测数据中的历史数据个数（增量数据不够时） */
	curRoundHistoryDatas *int
	/* 连续未更新次数(用于判断卡死，当前仅用于劣化感知(steptime.csv)) */
	nContinuousNotUpdate *int
}

const zpDevice string = "ZP_device"

const zpHost string = "ZP_host"

const ppDevice string = "PP_device"

const dataLoaderHost string = "dataloader_host"

const stepTime string = "stepTime"

var recorderSyncLock sync.RWMutex

/* nodeLevel data Recorder; key:jobName value:(key:npuId, value:(key:算子时延类型, value:记录数据))*/
var nodeLevelDataRecorder = make(map[string]map[int]map[string]dataRecord)

// InitCurJobDetectionRecorder 初始化当前任务的recorder
func InitCurJobDetectionRecorder(jobName string, ranks []int) bool {
	recorderSyncLock.Lock()
	defer recorderSyncLock.Unlock()
	if _, exist := nodeLevelDataRecorder[jobName]; exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s has already been started or not stopped!", jobName)
		return false
	}
	nodeLevelDataRecorder[jobName] = make(map[int]map[string]dataRecord, len(ranks))
	for _, npuId := range ranks {
		nodeLevelDataRecorder[jobName][npuId] = map[string]dataRecord{
			zpDevice: dataRecord{
				new(float64), new(float64), new(int), new(int),
				new(bool), new(int), new(int)},
			zpHost: dataRecord{new(float64), new(float64),
				new(int), new(int),
				new(bool), new(int), new(int)},
			ppDevice: dataRecord{new(float64), new(float64),
				new(int), new(int),
				new(bool), new(int), new(int)},
			dataLoaderHost: dataRecord{new(float64), new(float64),
				new(int), new(int),
				new(bool), new(int), new(int)},
			stepTime: dataRecord{new(float64), new(float64),
				new(int), new(int),
				new(bool), new(int), new(int)},
		}
	}
	return true
}

// DeleteJobDetectionRecorder 删除target job的recorder
func DeleteJobDetectionRecorder(jobName string) {
	recorderSyncLock.Lock()
	defer recorderSyncLock.Unlock()
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] delete %s: data recorder has already been deleted!", jobName)
		return
	}
	delete(nodeLevelDataRecorder, jobName)
}

// SetJobDetectionRecorderMean 更新均值
func SetJobDetectionRecorderMean(jobName string, npuId int, value float64, column string) {
	recorderSyncLock.Lock()
	defer recorderSyncLock.Unlock()
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s mean not exist!", jobName)
		return
	}
	if _, exist := nodeLevelDataRecorder[jobName][npuId]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s card: %v not exist!", jobName, npuId)
		return
	}
	/* 初始化必然存在 */
	addr := nodeLevelDataRecorder[jobName][npuId][column].mean
	*addr = value
}

// GetJobDetectionRecorderMean 获取均值
func GetJobDetectionRecorderMean(jobName string, npuId int, column string) float64 {
	recorderSyncLock.RLock()
	defer recorderSyncLock.RUnlock()
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s mean not exist!", jobName)
		return -1
	}
	if _, exist := nodeLevelDataRecorder[jobName][npuId]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s card: %v not exist!", jobName, npuId)
		return -1
	}
	/* 初始化必然存在 */
	value := *(nodeLevelDataRecorder[jobName][npuId][column].mean)
	return value
}

// SetJobDetectionRecorderStdDev 更新标准差
func SetJobDetectionRecorderStdDev(jobName string, npuId int, value float64, column string) {
	recorderSyncLock.Lock()
	defer recorderSyncLock.Unlock()
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s stdDev not exist!", jobName)
		return
	}
	if _, exist := nodeLevelDataRecorder[jobName][npuId]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s card: %v not exist!", jobName, npuId)
		return
	}
	/* 初始化必然存在 */
	addr := nodeLevelDataRecorder[jobName][npuId][column].stdDev
	*addr = value
}

// GetJobDetectionRecorderStdDev 获取标准差
func GetJobDetectionRecorderStdDev(jobName string, npuId int, column string) float64 {
	recorderSyncLock.RLock()
	defer recorderSyncLock.RUnlock()
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s stdDev not exist!", jobName)
		return -1
	}
	if _, exist := nodeLevelDataRecorder[jobName][npuId]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s card: %v not exist!", jobName, npuId)
		return -1
	}
	/* 初始化必然存在 */
	value := *(nodeLevelDataRecorder[jobName][npuId][column].stdDev)
	return value
}

// SetJobDetectionRecorderHealthDataCount 更新计算均值的数据总数
func SetJobDetectionRecorderHealthDataCount(jobName string, npuId int, value int, column string) {
	recorderSyncLock.Lock()
	defer recorderSyncLock.Unlock()
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s healthDataCount not exist!", jobName)
		return
	}
	if _, exist := nodeLevelDataRecorder[jobName][npuId]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s card: %v not exist!", jobName, npuId)
		return
	}
	/* 初始化必然存在 */
	addr := nodeLevelDataRecorder[jobName][npuId][column].healthDataCount
	*addr = value
}

// GetJobDetectionRecorderHealthDataCount 计算均值的数据总数
func GetJobDetectionRecorderHealthDataCount(jobName string, npuId int, column string) int {
	recorderSyncLock.RLock()
	defer recorderSyncLock.RUnlock()
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s healthDataCount not exist!", jobName)
		return -1
	}
	if _, exist := nodeLevelDataRecorder[jobName][npuId]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s card: %v not exist!", jobName, npuId)
		return -1
	}
	/* 初始化必然存在 */
	value := *(nodeLevelDataRecorder[jobName][npuId][column].healthDataCount)
	return value
}

// SetJobDetectionRecorderRoundUpdateFlag 设置当前检测数据是否已经用于更新均值标准差flag
func SetJobDetectionRecorderRoundUpdateFlag(jobName string, npuId int, value bool, column string) {
	recorderSyncLock.Lock()
	defer recorderSyncLock.Unlock()
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s curRoundUpdated not exist!", jobName)
		return
	}
	if _, exist := nodeLevelDataRecorder[jobName][npuId]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s card: %v not exist!", jobName, npuId)
		return
	}
	/* 初始化必然存在 */
	addr := nodeLevelDataRecorder[jobName][npuId][column].curRoundUpdated
	*addr = value
}

// SetJobDetectionRecorderAllUpdateFlag 设置当前检测数据所有数据列更新均值标准差flag
func SetJobDetectionRecorderAllUpdateFlag(jobName string, value bool) {
	recorderSyncLock.Lock()
	defer recorderSyncLock.Unlock()
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s AllRoundUpdateFlag not exist!", jobName)
		return
	}
	/* 将当前job每张卡的每个数据列的本轮检测更新flag进行设置 */
	for _, subMap := range nodeLevelDataRecorder[jobName] {
		for _, data := range subMap {
			addr := data.curRoundUpdated
			*addr = value
		}
	}
}

// SetJobDetectionRecorderAllHistoryDatas 仅用于初始化历史使用数据
func SetJobDetectionRecorderAllHistoryDatas(jobName string, value int) {
	recorderSyncLock.Lock()
	defer recorderSyncLock.Unlock()
	/* 仅能设置为0 */
	if value > 0 || value < 0 {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]Job: %s history data only can set 0!", jobName)
		value = 0
	}
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s ALLHistoryData not exist!", jobName)
		return
	}
	/* 将当前job每张卡的每个数据列的本轮检测更新flag进行设置 */
	for _, subMap := range nodeLevelDataRecorder[jobName] {
		for _, data := range subMap {
			addr := data.curRoundHistoryDatas
			*addr = value
		}
	}
}

// SetJobDetectionRecorderHistoryData 设置本轮检测对应npu的数据列的历史数据使用数量
func SetJobDetectionRecorderHistoryData(jobName string, npuId int, column string, value int) {
	recorderSyncLock.Lock()
	defer recorderSyncLock.Unlock()
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s curRoundHistoryData not exist!", jobName)
		return
	}
	if _, exist := nodeLevelDataRecorder[jobName][npuId]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s card: %v not exist!", jobName, npuId)
		return
	}
	addr := nodeLevelDataRecorder[jobName][npuId][column].curRoundHistoryDatas
	*addr = value
}

// GetJobDetectionRecorderHistoryData 获取本轮检测对应npu的数据列的历史数据使用数量
func GetJobDetectionRecorderHistoryData(jobName string, npuId int, column string) int {
	recorderSyncLock.RLock()
	defer recorderSyncLock.RUnlock()
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s curRoundHistoryData not exist!", jobName)
		return 0
	}
	if _, exist := nodeLevelDataRecorder[jobName][npuId]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s card: %v not exist!", jobName, npuId)
		return 0
	}
	value := *(nodeLevelDataRecorder[jobName][npuId][column].curRoundHistoryDatas)
	return value
}

// GetJobDetectionRecorderRoundUpdateFlag 获取当前检测数据是否已经用于更新均值标准差flag
func GetJobDetectionRecorderRoundUpdateFlag(jobName string, npuId int, column string) (bool, bool) {
	recorderSyncLock.RLock()
	defer recorderSyncLock.RUnlock()
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s curRoundUpdated not exist!", jobName)
		return false, false
	}
	if _, exist := nodeLevelDataRecorder[jobName][npuId]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s card: %v not exist!", jobName, npuId)
		return false, false
	}
	/* 初始化必然存在 */
	value := *(nodeLevelDataRecorder[jobName][npuId][column].curRoundUpdated)
	return value, true
}

// SetJobDetectionRecorderMaxDetectedStep 设置已检测数据行数
func SetJobDetectionRecorderMaxDetectedStep(jobName string, npuId int, value int, column string) {
	recorderSyncLock.Lock()
	defer recorderSyncLock.Unlock()
	if column != zpDevice && column != zpHost && column != dataLoaderHost &&
		column != stepTime && column != ppDevice {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s not exist!", column)
		return
	}
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s not exist!", jobName)
		return
	}
	if _, exist := nodeLevelDataRecorder[jobName][npuId]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s card: %v not exist!", jobName, npuId)
		return
	}
	/* 初始化必然存在 */
	addr := nodeLevelDataRecorder[jobName][npuId][column].maxDetectedStep
	*addr = value
}

// GetJobDetectionRecorderMaxDetectedStep 获取已检测数据行数
func GetJobDetectionRecorderMaxDetectedStep(jobName string, npuId int, column string) int {
	recorderSyncLock.RLock()
	defer recorderSyncLock.RUnlock()
	if column != zpDevice && column != zpHost && column != dataLoaderHost &&
		column != stepTime && column != ppDevice {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s not exist!", column)
		return -1
	}
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s - detected line not exist!", jobName)
		return -1
	}
	if _, exist := nodeLevelDataRecorder[jobName][npuId]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s card: %v not exist!", jobName, npuId)
		return -1
	}
	/* 初始化必然存在 */
	value := *(nodeLevelDataRecorder[jobName][npuId][column].maxDetectedStep)
	return value
}

// AddJobDetectionContinuousNotUpdateTimes 记录时延文件未更新次数
func AddJobDetectionContinuousNotUpdateTimes(jobName string, npuId int, column string) {
	recorderSyncLock.Lock()
	defer recorderSyncLock.Unlock()
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s - continuous not update flag not exist!", jobName)
		return
	}
	if _, exist := nodeLevelDataRecorder[jobName][npuId]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s card: %v not exist!", jobName, npuId)
		return
	}
	addr := nodeLevelDataRecorder[jobName][npuId][column].nContinuousNotUpdate
	*addr = *addr + 1
}

// CleanJobDetectionContinuousNotUpdateTimes 归零时延文件未更新次数
func CleanJobDetectionContinuousNotUpdateTimes(jobName string, npuId int, column string) {
	recorderSyncLock.Lock()
	defer recorderSyncLock.Unlock()
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s - continuous not update flag not exist!", jobName)
		return
	}
	if _, exist := nodeLevelDataRecorder[jobName][npuId]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s card: %v not exist!", jobName, npuId)
		return
	}
	addr := nodeLevelDataRecorder[jobName][npuId][column].nContinuousNotUpdate
	*addr = 0
}

// GetJobDetectionContinuousNotUpdateTimes 获取时延文件未更新次数
func GetJobDetectionContinuousNotUpdateTimes(jobName string, npuId int, column string) int {
	recorderSyncLock.RLock()
	defer recorderSyncLock.RUnlock()
	if _, exist := nodeLevelDataRecorder[jobName]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s - continuous not update flag not exist!", jobName)
		return -1
	}
	if _, exist := nodeLevelDataRecorder[jobName][npuId]; !exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO] Job: %s card: %v not exist!", jobName, npuId)
		return -1
	}
	/* 初始化必然存在 */
	value := *(nodeLevelDataRecorder[jobName][npuId][column].nContinuousNotUpdate)
	return value
}
