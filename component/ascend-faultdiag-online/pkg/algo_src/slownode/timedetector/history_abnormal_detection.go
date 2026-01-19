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
Package timedetector is used for time dimension detection by comparing data
with itself to identify significantly abnormal data points in time series data.
*/
package timedetector

import (
	"math"
	"sort"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/nodeleveldatarecorder"
)

// 常量定义
const (
	defaultSigmaMultiplier = 3.0 // 默认 Sigma 的乘数
	dataCountLarge         = 1.5
	/* normalNumber 和 nConsecAbnormal关系倍数阈值 */
	algoInitialParam   int = 4
	defaultStandardNum     = 3
)

type findAbnormalRet struct {
	abnormalIndexes []int
	mean            float64
	stdDev          float64
	validCount      int
}

// calculateMean 计算数据列表的均值
func calculateMean(dataList []float64) float64 {
	if len(dataList) == 0 {
		return 0.0
	}
	var sum float64
	for _, value := range dataList {
		sum += value
	}
	return sum / float64(len(dataList))
}

// calculateStandardDeviation 计算 float64 切片的标准差
func calculateStandardDeviation(data []float64, mean float64) float64 {
	if len(data) == 0 {
		return 0
	}

	var sumOfSquares float64
	for _, value := range data {
		diff := value - mean
		sumOfSquares += diff * diff
	}
	variance := sumOfSquares / float64(len(data))
	stdDev := math.Sqrt(variance)
	return stdDev
}

// 获取阈值的上界
// 存在两种阈值的计算方法，1、硬阈值；2、NSigma；
// 硬阈值的使用优先级高于NSigma，在两种阈值参数均设置为0时，默认使用3Sigma；
func getLowerBound(Nsigma int, degradationPercentage float64, mean float64, stdDev float64) float64 {
	var lowerBound float64
	if Nsigma == 0 {
		if degradationPercentage == 0 {
			lowerBound = mean - defaultSigmaMultiplier*stdDev
		} else {
			lowerBound = mean * (1.0 - degradationPercentage)
		}
	} else {
		if degradationPercentage == 0 {
			lowerBound = mean - float64(Nsigma)*stdDev
		} else {
			lowerBound = mean * (1.0 - degradationPercentage)
		}
	}
	return lowerBound
}

// 获取阈值的上界
// 存在两种阈值的计算方法，1、硬阈值；2、NSigma；
// 硬阈值的使用优先级高于NSigma，在两种阈值参数均设置为0时，默认使用3Sigma；
func getUpperBound(Nsigma int, degradationPercentage float64, mean float64, stdDev float64) float64 {
	var upperBound float64
	if Nsigma == 0 {
		if degradationPercentage == 0 {
			upperBound = mean + defaultSigmaMultiplier*stdDev
		} else {
			upperBound = mean * (1.0 + degradationPercentage)
		}
	} else {
		if degradationPercentage == 0 {
			upperBound = mean + float64(Nsigma)*stdDev
		} else {
			upperBound = mean * (1.0 + degradationPercentage)
		}
	}

	return upperBound
}

/* 更新均值、标准差 */
func updateMeanAndStdN(mean float64,
	stdDev float64,
	dataCount int,
	newOne float64) (float64, float64, int) {
	newMean := (float64(dataCount)*mean + newOne) / (float64(dataCount) + 1.0)
	// 校验是否越界
	if math.IsInf(newMean, 0) {
		hwlog.RunLog.Error("[SLOWNODE ALGO] mean calculation overflow")
		return 0, 0, 0
	}
	mean = newMean
	curVariance := stdDev * stdDev
	if math.IsInf(curVariance, 0) {
		hwlog.RunLog.Error("[SLOWNODE ALGO] curVariance calculation overflow")
		return 0, 0, 0
	}
	sumSqDiff := curVariance*float64(dataCount) + (newOne-mean)*(newOne-newMean)
	/* 判断数据有效性 */
	if math.IsNaN(sumSqDiff) || math.IsInf(sumSqDiff, 0) {
		return -1, 0, 0
	}
	// 更新方差
	varianceNew := sumSqDiff / float64(dataCount+1)

	// 新的标准差是方差的平方根
	stdDevNew := math.Sqrt(varianceNew)
	/* 判断数据有效性 */
	if math.IsNaN(stdDevNew) || math.IsInf(stdDevNew, 0) {
		return -1, 0, 0
	}
	stdDev = stdDevNew

	dataCount += 1
	return mean, stdDev, dataCount
}

/* 第一次获取均值和标准差 */
func firstGetMeanAndStdDev(
	conf config.AlgoInputConfig,
	originalSlice []float64,
	npuId int,
	column string) (float64, float64, bool) {
	if len(originalSlice) < conf.NormalNumber+conf.NconsecAnomaliesSignifySlow {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO] %s: data is not enough to detect", conf.JobName)
		return -1, -1, false
	}
	/* 取前部分数据计算平均值和方差 */
	bigNormalDataSlice := append([]float64{}, originalSlice[:conf.NormalNumber]...)
	sort.Slice(bigNormalDataSlice, func(i, j int) bool {
		return bigNormalDataSlice[i] <= bigNormalDataSlice[j]
	})
	normalDataSlice := getNormalDataSlice(bigNormalDataSlice, conf.NormalNumber, conf.NconsecAnomaliesSignifySlow)
	mean := calculateMean(normalDataSlice)
	stdDev := calculateStandardDeviation(normalDataSlice, mean)
	/* 记录第一次计算均值使用的数据量中正常数据量个数 */
	nodeleveldatarecorder.SetJobDetectionRecorderHealthDataCount(conf.JobName, npuId, len(normalDataSlice), column)
	return mean, stdDev, true
}

/* 获取均值和方差 */
func getMeanAndStdDev(conf config.AlgoInputConfig,
	npuId int,
	column string,
	dataSlice []float64) (float64, float64, bool) {
	if nodeleveldatarecorder.GetJobDetectionRecorderMean(conf.JobName, npuId, column) == 0 {
		return firstGetMeanAndStdDev(conf, dataSlice, npuId, column)
	}
	/* 非第一次处理数据,取历史记录数据（在调用FirstNPointsShouldBeNormal检测之前已判断有无增量） */
	mean := nodeleveldatarecorder.GetJobDetectionRecorderMean(conf.JobName, npuId, column)
	stdDev := nodeleveldatarecorder.GetJobDetectionRecorderStdDev(conf.JobName, npuId, column)
	return mean, stdDev, true
}

/* 查找异常值并就当前数据更新均值、标准差 */
func findAbnormalValues(
	increment []float64,
	conf config.AlgoInputConfig,
	npuId int,
	column string,
	updateFlag bool) (findAbnormalRet, bool) {
	mean, stdDev, ok := getMeanAndStdDev(conf, npuId, column, increment)
	if !ok {
		return findAbnormalRet{}, false
	}
	validDataCount := nodeleveldatarecorder.GetJobDetectionRecorderHealthDataCount(conf.JobName, npuId, column)
	/* 获取增量中历史数据个数（仅增量不够时会>0） */
	history := nodeleveldatarecorder.GetJobDetectionRecorderHistoryData(conf.JobName, npuId, column)
	/* 如果是第一次检测，前conf.Normal个数据已用于计算均值, history一定==0 */
	if nodeleveldatarecorder.GetJobDetectionRecorderMaxDetectedStep(conf.JobName, npuId, column) == 0 {
		history = conf.NormalNumber
	}
	/* 记录异常指标 */
	var abnormalIndexs = []int{}
	/* 用增量数据更新均值，标准差 */
	for index, value := range increment {
		if value > getUpperBound(conf.Nsigma, conf.DegradationPercentage, mean, stdDev) {
			abnormalIndexs = append(abnormalIndexs, index)
		} else if updateFlag && index+1 > history {
			/* 若本轮检测数据是第一次检测，则非异常指标更新均值、正常数据++ */
			tmpMean, tmpStdDev, tmpValidDataCount := updateMeanAndStdN(mean, stdDev, validDataCount, value)
			if tmpMean > 0 {
				mean = tmpMean
				stdDev = tmpStdDev
				validDataCount = tmpValidDataCount
			}
		}
	}
	ret := findAbnormalRet{
		abnormalIndexes: abnormalIndexs,
		mean:            mean,
		stdDev:          stdDev,
		validCount:      validDataCount,
	}
	return ret, true
}

/* 动态更新mean 和 stdDev 并返回异常值下标 */
func dynamicUpdateMeanAndStdDev(conf config.AlgoInputConfig, npuId int, column string, increment []float64) []int {
	update, exist := nodeleveldatarecorder.GetJobDetectionRecorderRoundUpdateFlag(conf.JobName, npuId, column)
	/* 获取当前检测数据中最新的异常数据和均值等信息 */
	ret, ok :=
		findAbnormalValues(increment, conf, npuId, column, !update && exist)
	if !ok {
		return nil
	}
	/* 判断是否需要更新当前npu卡的recorder数据,避免同一组数据重复更新 */
	if !update && exist {
		/* detectedLine设置放在读取数据确认起始行时 */
		nodeleveldatarecorder.SetJobDetectionRecorderStdDev(conf.JobName, npuId, ret.stdDev, column)
		nodeleveldatarecorder.SetJobDetectionRecorderMean(conf.JobName, npuId, ret.mean, column)
		nodeleveldatarecorder.SetJobDetectionRecorderHealthDataCount(conf.JobName, npuId, ret.validCount, column)
		nodeleveldatarecorder.SetJobDetectionRecorderRoundUpdateFlag(conf.JobName, npuId, true, column)
	}
	return ret.abnormalIndexes
}

func checkIsAsExpectedAbnormalData(abnormalIndexes []int,
	dataSlice []float64,
	conf config.AlgoInputConfig,
	npuId int,
	column string) bool {
	/* 判断异常指标是否足够 */
	if len(abnormalIndexes) < conf.NconsecAnomaliesSignifySlow {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO] npu: %v %s abnormal data not enough!", npuId, column)
		return false
	}
	/* 判断变慢是否已恢复:最后一个异常值下标是否等于当前所有数据最后一个 */
	lastAbnormalIndex := abnormalIndexes[len(abnormalIndexes)-1]
	if len(dataSlice) > (lastAbnormalIndex + 1) {
		hwlog.RunLog.Infof("[SLOWNODE ALGO] %s npu: %v column: %s - abnormal data recovered!",
			conf.JobName, npuId, column)
		return false
	}
	/* 判断是否满足变慢最少 异常数量（且需要连续step的数据） */
	if !(abnormalIndexes[len(abnormalIndexes)-1]-
		abnormalIndexes[len(abnormalIndexes)-
			conf.NconsecAnomaliesSignifySlow]+1 ==
		conf.NconsecAnomaliesSignifySlow) {
		return false
	}
	return true
}

// FirstNPointsShouldBeNormal 时间维度检测，限制条件是：开头的数据是正常的，无故障发生
// 检测当前卡是否变慢并且检测的对应列数据是否是最新的n个数据， 入参incrementSlice是增量部分 >= NconsecAbnormal
func FirstNPointsShouldBeNormal(
	incrementData []float64,
	conf config.AlgoInputConfig,
	npuId int,
	column string) (int, float64) {
	/* 异常值下标 */
	abnormalIndexes := dynamicUpdateMeanAndStdDev(conf, npuId, column, incrementData)
	if abnormalIndexes == nil || len(abnormalIndexes) == 0 {
		return 0, 0.0
	}
	/* 检测异常下标是否符合预期 */
	if !checkIsAsExpectedAbnormalData(abnormalIndexes,
		incrementData, conf, npuId, column) {
		return 0, 0.0
	}
	var isSlow int = 1
	sumOfLastNCData := 0.0
	/* 取出最新的n个数据计算变慢程度 */
	for _, value := range incrementData[(len(incrementData) - conf.NconsecAnomaliesSignifySlow):] {
		sumOfLastNCData += value
	}
	meanOfLastNCData := sumOfLastNCData / float64(conf.NconsecAnomaliesSignifySlow)
	degradationLevel :=
		meanOfLastNCData/nodeleveldatarecorder.GetJobDetectionRecorderMean(conf.JobName, npuId, column) - 1.0
		/* 判断数据有效性 */
	if math.IsNaN(degradationLevel) || math.IsInf(degradationLevel, 0) {
		return 1, 0.0
	}
	return isSlow, degradationLevel
}

/* 根据normalNumber与NConsecAbnormal关心选取中间数据进行计算 */
func getNormalDataSlice(bigNormalDataSlice []float64, normalNumber int, nConsecAbnormal int) []float64 {
	start := 0
	end := 0
	if normalNumber >= algoInitialParam*nConsecAbnormal {
		start = nConsecAbnormal
		end = len(bigNormalDataSlice) - nConsecAbnormal
	} else {
		start = defaultStandardNum
		end = len(bigNormalDataSlice) - defaultStandardNum
	}
	bigNormalDataSlice = bigNormalDataSlice[start:end]
	return bigNormalDataSlice
}
