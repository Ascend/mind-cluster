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

// Package algo 网络连通性检测算法
package algo

import (
	"math"

	"ascend-common/common-utils/hwlog"
)

// 返回当前超节点的滑窗
func (nd *NetDetect) getCurSlideWindow() []map[string]any {
	return nd.curSlideWindows
}

// 更新滑动窗口的数据
func (nd *NetDetect) updateCurSlideWindows(newData []map[string]any) {
	hwlog.RunLog.Infof("[ALGO] begin to updateCurSlideWindows, superPodId: %v", nd.curSuperPodId)
	nd.curSlideWindows = MergeAndDeduplicate(nd.curSlideWindows, newData)
	nd.curSlideWindowsMaxTs = getMaxTimeStamp(nd.curSlideWindows)

	totalTimeSpan := int64(millisecondNum * saveLenNum * nd.curPingPeriod)
	var validItems []map[string]any
	for _, item := range nd.curSlideWindows {
		if ts, ok := item[timestampConstant].(int64); ok {
			if nd.curSlideWindowsMaxTs-ts <= totalTimeSpan {
				validItems = append(validItems, item)
			}
		}
	}
	nd.curSlideWindows = validItems
	hwlog.RunLog.Infof("[ALGO] updateCurSlideWindows finished, widnowData size: %v, superPodId: %v",
		len(nd.curSlideWindows), nd.curSuperPodId)
}

// 消费queue数据
func (nd *NetDetect) consumeQueueData() {
	updateTime := nd.curPingPeriod // 更新的时间数

	// 移除最晚的指定时间数数据，并更新最大时间戳
	hwlog.RunLog.Infof("[ALGO] begin to consumeQueueData, superPodId: %v", nd.curSuperPodId)
	totalTimeSpan := int64(nd.curPingPeriod*millisecondNum*saveLenNum - updateTime*millisecondNum)
	var validItems []map[string]any
	for _, item := range nd.curSlideWindows {
		if ts, ok := item[timestampConstant].(int64); ok {
			if nd.curSlideWindowsMaxTs-ts <= totalTimeSpan {
				validItems = append(validItems, item)
			}
		}
	}
	nd.curSlideWindows = validItems

	// 从curConsumedQueue取走指定时间数的数据
	var windowItems []map[string]any
	var queueItems []map[string]any
	for _, item := range nd.curConsumedQueue {
		if ts, ok := item[timestampConstant].(int64); ok {
			if ts-nd.curSlideWindowsMaxTs <= int64(updateTime*millisecondNum) &&
				ts-nd.curSlideWindowsMaxTs > 0 {
				windowItems = append(windowItems, item)
			}
			if ts-nd.curSlideWindowsMaxTs > int64(updateTime*millisecondNum) {
				queueItems = append(queueItems, item)
			}
		}
	}
	nd.curSlideWindows = append(nd.curSlideWindows, windowItems...)
	nd.curConsumedQueue = queueItems
	nd.curSlideWindowsMaxTs += int64(updateTime * millisecondNum)
	hwlog.RunLog.Infof("[ALGO] consumeQueueData finished, queueData size: %v, widnowData size: %v, "+
		"superPodId: %v", len(nd.curConsumedQueue), len(nd.curSlideWindows), nd.curSuperPodId)
}

/*
 * 在滑窗里取指定周期的数据
 *
 * 新<-----------------旧
 *
 * 0 1 2 3 4 5 6 7 8 9 10
 * —————————————————————
 * | | | | | | | | | | |
 * —————————————————————
 */
func (nd *NetDetect) getWindowData(windows []map[string]any, startPeriod int,
	endPeriod int) []map[string]any {
	if startPeriod >= endPeriod {
		return []map[string]any{}
	}

	// 预计算常量
	startBound := nd.curSlideWindowsMaxTs - int64(millisecondNum*startPeriod*nd.curPingPeriod)
	endBound := nd.curSlideWindowsMaxTs - int64(millisecondNum*endPeriod*nd.curPingPeriod)

	// 预分配结果切片容量
	result := make([]map[string]any, 0, len(windows)*(endPeriod-startPeriod)/saveLenNum)

	for _, curItem := range windows {
		eachTimestamp, ok := curItem[timestampConstant].(int64)
		if !ok {
			continue
		}

		if eachTimestamp > startBound || eachTimestamp <= endBound {
			continue
		}
		result = append(result, curItem)
	}

	return result
}

// 寻找相同故障路径
func (nd *NetDetect) findSamePath(windows []map[string]any,
	path map[string]any) []map[string]any {
	// 预分配结果切片（最多nd.curPingPeriod个相同路径）
	res := make([]map[string]any, nd.curPingPeriod)

	for _, item := range windows {
		if isSamePath(item, path) {
			res = append(res, item)
		}
	}

	return res
}

// 计算窗口数据指定路径的动态阈值
func calDynamicThresholds(samePaths []map[string]any, faultType string) float64 {
	// 预过滤有效数据，避免重复类型断言
	values := make([]float64, 0, len(samePaths))
	for _, item := range samePaths {
		if v, ok := item[faultType].(float64); ok {
			if faultType == avgDelayConstant {
				v = math.Round(v / convertFactor)
			}
			values = append(values, v)
		}
	}
	if len(values) == 0 {
		return 0.0
	}

	// 计算均值（单次遍历）
	avg, variance := 0.0, 0.0
	for _, v := range values {
		avg += v
	}
	avg /= float64(len(values))

	// 计算方差
	for _, v := range values {
		diff := v - avg
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(values)))

	// 动态阈值
	minStdDev := 1.0
	if stdDev > minStdDev {
		minStdDev = stdDev
	}
	return avg + coefficientNum*minStdDev
}

// 判断是否是同一条探测路径
func isSamePath(pathA map[string]any, pathB map[string]any) bool {
	for _, key := range globalPathKeys {
		if pathA[key] != pathB[key] {
			return false
		}
	}

	return true
}

// 获取窗口数据的周期数
func (nd *NetDetect) getPeriodNum(windows []map[string]any) int {
	if len(windows) == 0 {
		return 0
	}

	minTimestamp := int64(math.MaxInt64)
	for i := 0; i < len(windows); i++ {
		curItem := windows[i]
		eachTimestamp, ok := curItem[timestampConstant].(int64)
		if !ok {
			continue
		}

		if eachTimestamp < minTimestamp {
			minTimestamp = eachTimestamp
		}
	}

	periodCount := float64(nd.curSlideWindowsMaxTs-minTimestamp) / float64(millisecondNum*nd.curPingPeriod)
	return int(math.Ceil(periodCount))
}

// 获取窗口数据的最大时间戳
func getMaxTimeStamp(windows []map[string]any) int64 {
	result := int64(math.MinInt64)
	for i := 0; i < len(windows); i++ {
		curItem := windows[i]
		curTimestamp, ok := curItem[timestampConstant].(int64)
		if !ok {
			continue
		}

		if curTimestamp > result {
			result = curTimestamp
		}
	}

	return result
}
