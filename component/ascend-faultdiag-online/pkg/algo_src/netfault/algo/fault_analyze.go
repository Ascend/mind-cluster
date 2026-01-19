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
	"fmt"
	"math"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"ascend-common/common-utils/hwlog"
)

// StartFaultDetect 故障检测算法入口
func (nd *NetDetect) StartFaultDetect(input []map[string]any) []any {
	rootCauseAlarmAll := make([]any, 0)

	// 1, 更新告警抑制ttl
	nd.updateHistoryAlarmMap()

	// 2, 格式化处理输入数据
	nd.formatInputData(input)

	// 3, 填充检测算法核心数据
	nd.fillDetectCoreData(input)

	// 4, 滑窗数据是否填满
	windowPeriod := nd.getPeriodNum(nd.curSlideWindows)
	if windowPeriod < saveLenNum {
		hwlog.RunLog.Infof("[ALGO] windows num is: %v, data is not full in superPodId: %v",
			windowPeriod, nd.curSuperPodId)
		return rootCauseAlarmAll
	} else {
		nd.curOpenQueueFlag = true
	}

	// 5, 待检测窗口数据是否为空
	startPeriod := 1 // 检测数据起始窗口数
	endPeriod := 2   // 检测数据结尾窗口
	detectData := nd.getWindowData(nd.curSlideWindows, startPeriod, endPeriod)
	if len(detectData) == 0 {
		hwlog.RunLog.Infof("[ALGO] no detect data found in superPodId: %s", nd.curSuperPodId)
		return rootCauseAlarmAll
	} else {
		hwlog.RunLog.Infof("[ALGO] %v detect data found in superPodId: %s", len(detectData), nd.curSuperPodId)
	}

	// 6, 获取最终告警
	rootCauseAlarmAll = nd.getFinalAlarm(detectData)
	return rootCauseAlarmAll
}

// 包导入时立马启动
func init() {
	if globalRootCauseEventNpu == nil || len(globalRootCauseEventNpu) == 0 {
		globalRootCauseEventNpu = make(map[string]string)
		globalRootCauseEventNpu[layer1Constant] = npuConstant
		globalRootCauseEventNpu[layer2Constant] = rackConstant
		globalRootCauseEventNpu[layer3Constant] = l1Constant
		globalRootCauseEventNpu[layer4Constant] = l2Constant
	}

	if globalRootCauseEventCpu == nil || len(globalRootCauseEventCpu) == 0 {
		globalRootCauseEventCpu = make(map[string]string)
		globalRootCauseEventCpu[layer1Constant] = cpuConstant
		globalRootCauseEventCpu[layer2Constant] = unionConstant
	}
}

// 模拟ttl实现，更新GlobalHistoryAlarms
func (nd *NetDetect) updateHistoryAlarmMap() {
	curStamp := time.Now().UnixMilli()
	for key, value := range globalHistoryAlarms {
		oldStamp, ok := value.(int64)
		if !ok {
			continue
		}

		// 直接比较时间差
		if int(curStamp-oldStamp) >= nd.curSuppressedPeriod*millisecondNum {
			hwlog.RunLog.Infof("[ALGO] alarm disappeared: %v, superPodId: %v", key, nd.curSuperPodId)
			delete(globalHistoryAlarms, key)
		}
	}
}

// 将待检测数据处理成算法需要的数据格式
func (nd *NetDetect) formatInputData(input []map[string]any) {
	for _, item := range input {
		nd.formatLayer(item)
		formatLossRate(item)
		formatDelay(item)
		formatTimestamp(item)
	}
}

// 将待检测数据填充到检测算法的核心数据结构里去
func (nd *NetDetect) fillDetectCoreData(input []map[string]any) {
	if nd.curOpenQueueFlag {
		// 开启数据队列策略
		nd.curConsumedQueue = MergeAndDeduplicate(nd.curConsumedQueue, input)
		nd.consumeQueueData()
	} else {
		// 刚开始填滑窗数据的时候不开启数据队列策略
		nd.updateCurSlideWindows(input)
	}
}

// 格式化layer信息
func (nd *NetDetect) formatLayer(item map[string]any) {
	if item == nil {
		return
	}

	srcAddr, srcOK := item[srcAddrConstant].(string)
	dstAddr, dstOk := item[dstAddrConstant].(string)
	if !srcOK || !dstOk {
		hwlog.RunLog.Errorf("[[NETFAULT ALGO]]convert srcAddrConstant: %v or dstAddrConstant: %v to string failed.",
			item[srcAddrConstant], item[dstAddrConstant])
		return
	}

	fromLayer := nd.findFullLayerPath(nd.curNpuInfo[srcAddr].RackName + ":" + srcAddr)
	if fromLayer != "" {
		item[fromLayerConstant] = fromLayer
	}

	toLayer := nd.findFullLayerPath(nd.curNpuInfo[dstAddr].RackName + ":" + dstAddr)
	if toLayer != "" {
		item[toLayerConstant] = toLayer
	}
}

// 格式化丢包率信息
func formatLossRate(item map[string]any) {
	if item == nil {
		return
	}

	avgLossRateStr, avgLossOk := item[avgLoseRateConstant].(string)
	minLossRateStr, minLossOk := item[minLoseRateConstant].(string)
	maxLossRateStr, maxLossOk := item[maxLoseRateConstant].(string)
	if !avgLossOk || !minLossOk || !maxLossOk {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]convert avgLoseRateConstant: %v or minLoseRateConstant: %v or "+
			"maxLoseRateConstant: %v to string failed.",
			item[avgLoseRateConstant], item[minLoseRateConstant], item[maxLoseRateConstant])
		return
	}

	if avgLossRateStr != "" {
		avgLossRateFloat, err := strconv.ParseFloat(avgLossRateStr, 64)
		if err == nil {
			item[avgLoseRateConstant] = basePercentNum * avgLossRateFloat
		}
	}

	if minLossRateStr != "" {
		minLossRateFloat, err := strconv.ParseFloat(minLossRateStr, 64)
		if err == nil {
			item[minLoseRateConstant] = basePercentNum * minLossRateFloat
		}
	}

	if maxLossRateStr != "" {
		maxLossRateFloat, err := strconv.ParseFloat(maxLossRateStr, 64)
		if err == nil {
			item[maxLoseRateConstant] = basePercentNum * maxLossRateFloat
		}
	}
}

// 格式化延迟信息
func formatDelay(item map[string]any) {
	if item == nil {
		return
	}

	avgDelayStr, avgDelayOk := item[avgDelayConstant].(string)
	minDelayStr, minDelayOk := item[minDelayConstant].(string)
	maxDelayStr, maxDelayOk := item[maxDelayConstant].(string)
	if !avgDelayOk || !minDelayOk || !maxDelayOk {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]convert avgDelayConstant: %v or minDelayConstant: %v or "+
			"maxDelayConstant: %v to string failed.",
			item[avgDelayConstant], item[minDelayConstant], item[maxDelayConstant])
		return
	}

	if avgDelayStr != "" {
		avgDelayFloat, err := strconv.ParseFloat(avgDelayStr, 64)
		if err != nil {
			hwlog.RunLog.Errorf("[NETFAULT ALGO]convert avgDelayStr: %s to float64 failed", avgDelayStr)
		} else {
			item[avgDelayConstant] = avgDelayFloat
		}
	}

	if minDelayStr != "" {
		minDelayFloat, err := strconv.ParseFloat(minDelayStr, 64)
		if err != nil {
			hwlog.RunLog.Errorf("[NETFAULT ALGO]convert minDelayStr: %s to float64 failed", minDelayStr)
		} else {
			item[minDelayConstant] = minDelayFloat
		}
	}

	if maxDelayStr != "" {
		maxDelayFloat, err := strconv.ParseFloat(maxDelayStr, 64)
		if err != nil {
			hwlog.RunLog.Errorf("[NETFAULT ALGO]convert maxDelayStr: %s to float64 failed", maxDelayStr)
		} else {
			item[maxDelayConstant] = maxDelayFloat
		}
	}
}

// 格式化时间戳信息
func formatTimestamp(item map[string]any) {
	if item == nil {
		return
	}

	timestampStr, timeOK := item[timestampConstant].(string)
	if !timeOK {
		return
	}

	if timestampStr != "" {
		numberSystem := 10 // 10进制
		timestampInt, err := strconv.ParseInt(timestampStr, numberSystem, 64)
		if err == nil {
			item[timestampConstant] = timestampInt
		}
	}
}

// 根据ip找到该ip的topo路径
func (nd *NetDetect) findFullLayerPath(ip string) string {
	for _, item := range nd.curTopo {
		if strings.Contains(item, ip) {
			return item
		}
	}

	return ""
}

// 获取当前超节点的根因告警列表
func (nd *NetDetect) getFinalAlarm(input []map[string]any) []any {
	rootCauseAlarmAll := make([]any, 0)

	calStartPeriod := 2 // 计算起始窗口数
	calEndPeriod := 10  // 计算结束窗口数
	calWindows := nd.getWindowData(nd.curSlideWindows, calStartPeriod, calEndPeriod)
	nd.updatePathIndex(calWindows)

	lossFaultPathList := make([]any, 0)
	delayFaultPathList := make([]any, 0)

	// 获取异常路径
	hwlog.RunLog.Infof("[ALGO] begin to getFaultPathList, superPodId: %v", nd.curSuperPodId)
	nd.getFaultPathList(input, &lossFaultPathList, &delayFaultPathList)
	hwlog.RunLog.Infof("[ALGO] success to getFaultPathList, superPodId: %v", nd.curSuperPodId)

	for detectIdx := 0; detectIdx < len(globalDetectTypes); detectIdx += 1 {
		// 检测类型
		detectType := globalDetectTypes[detectIdx]

		var faultPathList []any
		if detectType == avgLoseRateConstant {
			faultPathList = lossFaultPathList
		} else {
			faultPathList = delayFaultPathList
		}

		// npu直连的异常路径告警、非npu直连的其他异常路径告警
		npuDireAlarmList, otherAlarmList := nd.diffFaultPathList(faultPathList, detectType)

		// npu直连根因告警
		if len(npuDireAlarmList) > 0 {
			hwlog.RunLog.Infof("[ALGO] %v npuDireAlarm path found, detectType: %v, superPodId: %v",
				len(npuDireAlarmList), detectType, nd.curSuperPodId)
			getNpuDireFaultAlarm(npuDireAlarmList, &rootCauseAlarmAll, detectType)
		} else {
			hwlog.RunLog.Infof("[ALGO] no npuDireAlarm path found, detectType: %v, superPodId: %v",
				detectType, nd.curSuperPodId)
		}

		// 其他异常路径根因告警
		if len(otherAlarmList) > 0 {
			hwlog.RunLog.Infof("[ALGO] %v otherAlarmList path found, detectType: %v, superPodId: %v",
				len(otherAlarmList), detectType, nd.curSuperPodId)
			nd.getOtherFaultAlarm(otherAlarmList, &rootCauseAlarmAll, detectType)
		} else {
			hwlog.RunLog.Infof("[ALGO] no otherAlarmList path found, detectType: %v, superPodId: %v",
				detectType, nd.curSuperPodId)
		}
	}

	return rootCauseAlarmAll
}

// 获取异常路径
func (nd *NetDetect) getFaultPathList(input []map[string]any, lossFaultPathList *[]any,
	delayFaultPathList *[]any) {
	for i := 0; i < len(input); i += 1 {
		path := input[i]
		samePaths := nd.findSamePathFast(path)
		if len(samePaths) == 0 {
			hwlog.RunLog.Infof("[ALGO] no same path found in window data, src: %v, dst: %v, superPodId: %s",
				path[srcAddrConstant], path[dstAddrConstant], nd.curSuperPodId)
			continue
		}

		lossDynamicThresholds, delayDynamicThresholds := getDynamicThresholds(samePaths)
		lossIndicator, delayIndicator := getIndicators(input, path)

		if delayIndicator >= delayDynamicThresholds {
			hwlog.RunLog.Infof("[ALGO] curSuperPodId: %v, detectType: %s, src: %v, dst: %v, samePathsNum: %v, "+
				"dynamicThreshold: %v, indicator: %v", nd.curSuperPodId, avgDelayConstant, path[srcAddrConstant],
				path[dstAddrConstant], len(samePaths), delayDynamicThresholds, delayIndicator)
			*delayFaultPathList = append(*delayFaultPathList, path)
		}

		if lossIndicator >= lossDynamicThresholds {
			hwlog.RunLog.Infof("[ALGO] curSuperPodId: %v, detectType: %s, src: %v, dst: %v, samePathsNum: %v, "+
				"dynamicThreshold: %v, indicator: %v", nd.curSuperPodId, avgLoseRateConstant, path[srcAddrConstant],
				path[dstAddrConstant], len(samePaths), lossDynamicThresholds, lossIndicator)
			*lossFaultPathList = append(*lossFaultPathList, path)
		}
	}
}

// calWindows更新时调用，更新路径的索引编号
func (nd *NetDetect) updatePathIndex(calWindows []map[string]any) {
	nd.pathIndex = make(map[string][]map[string]any)
	for _, path := range calWindows {
		key := getPathHashKey(path)
		if _, exists := nd.pathIndex[key]; !exists {
			nd.pathIndex[key] = make([]map[string]any, 0, 1)
		}
		nd.pathIndex[key] = append(nd.pathIndex[key], path)
	}
}

// 按globalPathKeys顺序拼接所有关键字段
func getPathHashKey(path map[string]any) string {
	var builder strings.Builder
	for _, key := range globalPathKeys {
		builder.WriteString(fmt.Sprintf("%v|", path[key]))
	}

	return builder.String()
}

// 直接返回预存结果，若无匹配返回nil
func (nd *NetDetect) findSamePathFast(path map[string]any) []map[string]any {
	key := getPathHashKey(path)
	return nd.pathIndex[key]
}

// 分别返回丢包动态阈值、时延动态阈值
func getDynamicThresholds(samePaths []map[string]any) (float64, float64) {
	lossDynamicThresholds := calDynamicThresholds(samePaths, avgLoseRateConstant)
	delayDynamicThresholds := calDynamicThresholds(samePaths, avgDelayConstant)
	return lossDynamicThresholds, delayDynamicThresholds
}

// 分别返回丢包检测值、时延检测值
func getIndicators(input []map[string]any, path map[string]any) (float64, float64) {
	lossIndicator := getCurPathIndicator(input, path, avgLoseRateConstant)
	delayIndicator := getCurPathIndicator(input, path, avgDelayConstant)
	return lossIndicator, delayIndicator
}

func isDuplicatedLinkCurDetectionPeriod(uniqueLinkPaths []interface{}, curLinkPath map[string]interface{}) bool {
	for i := 0; i < len(uniqueLinkPaths); i++ {
		path, ok := uniqueLinkPaths[i].(map[string]interface{})
		if !ok {
			hwlog.RunLog.Warn("[ALGO] transfer fault link obj failed!")
			continue
		}
		src, ok1 := path[srcAddrConstant].(string)
		dst, ok2 := path[dstAddrConstant].(string)
		curSrc, ok3 := curLinkPath[srcAddrConstant].(string)
		curDst, ok4 := curLinkPath[dstAddrConstant].(string)
		if !ok1 || !ok2 || !ok3 || !ok4 {
			hwlog.RunLog.Warn("[ALGO] transfer fault link address string failed!")
			continue
		}
		if src != curSrc || dst != curDst {
			continue
		}
		hwlog.RunLog.Infof("[ALGO] remove duplicated link path, src: %v, dst: %v", src, dst)
		return true
	}
	return false
}

// 组合异常路径
func (nd *NetDetect) diffFaultPathList(faultPathList []any, detectType string) ([]any, []any) {
	npuDireAlarmList := make([]any, 0)
	otherAlarmList := make([]any, 0)

	faultPathLen := len(faultPathList)
	for i := 0; i < faultPathLen; i += 1 {
		path, ok := faultPathList[i].(map[string]any)
		if !ok {
			continue
		}
		path[faultTypeConstant] = detectType

		// 挑选出npu直连的异常路径（npu直连的没有完整路径，只有ip->ip，因此单独拿出来进行分析）
		if !containsKey(path, fromLayerConstant) || !containsKey(path, toLayerConstant) {
			if !isDuplicatedLinkCurDetectionPeriod(npuDireAlarmList, path) {
				npuDireAlarmList = append(npuDireAlarmList, path)
			}
			continue
		}
		if isDuplicatedLinkCurDetectionPeriod(otherAlarmList, path) {
			continue
		}

		fromLayerStr, ok := path[fromLayerConstant].(string)
		toLayerStr, ok := path[toLayerConstant].(string)
		if !ok {
			continue
		}
		desc := fmt.Sprintf("%s%s%s%s", "S-", fromLayerStr, " TO D-", toLayerStr)
		path[descriptionConstant] = desc

		srcNodes := getLayerObject(fromLayerStr)
		dstNodes := getLayerObject(toLayerStr)
		srcNodes = reverseSlice(srcNodes)
		dstNodes = reverseSlice(dstNodes)
		if len(srcNodes) != len(dstNodes) {
			continue
		}

		layerNum := 0
		for j := 0; j < len(srcNodes); j++ {
			if srcNodes[j] != dstNodes[j] {
				layerNum++
			}
		}
		if nd.curNpuType == a3NpuTypeConstant && layerNum == 0 {
			layerNum++
		}
		info := fmt.Sprintf("%s%s%d%s%s", layerConstant, objectIntervalChar, layerNum, portIntervalChar, srcNodes[layerNum])
		path[informationConstant] = info
		otherAlarmList = append(otherAlarmList, path)
	}

	return npuDireAlarmList, otherAlarmList
}

// 获取该条路径的平均值
func getCurPathIndicator(data []map[string]any, path map[string]any, detectType string) float64 {
	sum := float64(0)
	count := 0
	length := len(data)
	for i := 0; i < length; i += 1 {
		item := data[i]
		if isSamePath(item, path) {
			curValue, ok := item[detectType].(float64)
			if !ok {
				continue
			}
			sum += curValue
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// 获取字符串里的拓扑对象
func getLayerObject(str string) []string {
	result := make([]string, 0)

	tmpArr1 := strings.Split(str, layerIntervalChar)
	for i := 0; i < len(tmpArr1); i++ {
		tmpArr2 := strings.Split(tmpArr1[i], portIntervalChar)
		result = append(result, tmpArr2[0])
	}

	return result
}

// 对字符串切片进行计数，返回出现最多的字符串和计数map
func countForSlice(target []string) (string, map[string]int) {
	countMap := make(map[string]int)
	var maxNum int
	var maxNumKey string

	for _, key := range target {
		countMap[key]++
		if countMap[key] > maxNum {
			maxNum = countMap[key]
			maxNumKey = key
		}
	}

	return maxNumKey, countMap
}

// 告警里公共指标的格式化
func setCommonAlarmFormat(faultPath []any, rootCauseAlarm map[string]any) {
	setIndicators(faultPath, rootCauseAlarm)
	setFaultType(rootCauseAlarm)
}

// 设置告警指标值
func setIndicators(faultPath []any, rootCauseAlarm map[string]any) {
	if rootCauseAlarm == nil || len(faultPath) == 0 {
		return
	}

	var maxLoss, maxDelay, sumLoss, sumDelay float64 = 0, 0, 0, 0
	var minDelay = math.MaxFloat64
	var minLoss = math.MaxFloat64
	var timeStamp int64 = 0
	var taskId string
	for i := 0; i < len(faultPath); i++ {
		curPath, ok1 := faultPath[i].(map[string]any)
		tmpMinDelay, ok2 := curPath[minDelayConstant].(float64)
		tmpMaxDelay, ok3 := curPath[maxDelayConstant].(float64)
		tmpMinLoss, ok4 := curPath[minLoseRateConstant].(float64)
		tmpMaxLoss, ok5 := curPath[maxLoseRateConstant].(float64)
		tmpTimeStamp, ok6 := curPath[timestampConstant].(int64)
		tmpDelay, ok7 := curPath[avgDelayConstant].(float64)
		tmpLossRate, ok8 := curPath[avgLoseRateConstant].(float64)
		tmpTaskId, ok9 := curPath[pingTaskIDConstant].(string)
		if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 || !ok6 || !ok7 || !ok8 || !ok9 {
			continue
		}
		if tmpMinDelay < minDelay {
			minDelay = tmpMinDelay
		}
		if tmpMaxDelay > maxDelay {
			maxDelay = tmpMaxDelay
		}
		if tmpMinLoss < minLoss {
			minLoss = tmpMinLoss
		}
		if tmpMaxLoss > maxLoss {
			maxLoss = tmpMaxLoss
		}
		if tmpTimeStamp > timeStamp {
			timeStamp = tmpTimeStamp
		}
		sumDelay += tmpDelay
		sumLoss += tmpLossRate
		taskId = tmpTaskId
	}
	rootCauseAlarm[taskIDConstant] = taskId
	rootCauseAlarm[timestampConstant] = timeStamp
	rootCauseAlarm[minLoseRateConstant] = roundToThreeDecimal(minLoss)
	rootCauseAlarm[maxLoseRateConstant] = roundToThreeDecimal(maxLoss)
	rootCauseAlarm[avgLoseRateConstant] = roundToThreeDecimal(sumLoss / (float64(len(faultPath))))
	rootCauseAlarm[minDelayConstant] = roundToThreeDecimal(minDelay)
	rootCauseAlarm[maxDelayConstant] = roundToThreeDecimal(maxDelay)
	rootCauseAlarm[avgDelayConstant] = roundToThreeDecimal(sumDelay / float64(len(faultPath)))
}

// 设置检测类型
func setFaultType(rootCauseAlarm map[string]any) {
	if rootCauseAlarm == nil {
		return
	}

	detectType, ok := rootCauseAlarm[faultTypeConstant].(string)
	if !ok {
		return
	}

	if detectType == globalDetectTypes[1] {
		rootCauseAlarm[faultTypeConstant] = delayType
	} else if rootCauseAlarm[avgLoseRateConstant].(float64) >= lossThreshold {
		rootCauseAlarm[faultTypeConstant] = disconnectType
	} else {
		rootCauseAlarm[faultTypeConstant] = lossRateType
	}
}

// 判断告警是否被抑制，true->被抑制 false->没被抑制
func isAlarmSuppressed(rootCauseAlarm map[string]any) bool {
	srcId, srcIdRes := rootCauseAlarm[srcIdConstant].(string)
	dstId, dstIdRes := rootCauseAlarm[dstIdConstant].(string)
	faultType, ftRes := rootCauseAlarm[faultTypeConstant].(int)
	level, levelRes := rootCauseAlarm[levelConstant].(int)
	if !srcIdRes || !dstIdRes || !ftRes || !levelRes {
		hwlog.RunLog.Error("[ALGO] type assertion failed when execute isAlarmSuppressed")
		return false
	}

	key := fmt.Sprintf("%s-%s-%d-%d", srcId, dstId, faultType, level)
	if _, exists := globalHistoryAlarms[key]; exists {
		curTimeStamp := time.Now().UnixMilli()
		globalHistoryAlarms[key] = curTimeStamp
		hwlog.RunLog.Infof("[ALGO] the same alarm appeared and suppressed: %v", key)
		return true
	} else {
		curTimeStamp := time.Now().UnixMilli()
		globalHistoryAlarms[key] = curTimeStamp
		hwlog.RunLog.Infof("[ALGO] add alarm to globalHistoryAlarms: %v", key)
		return false
	}
}

// 获取npu直连的根因告警，并添加到总根因告警里
func getNpuDireFaultAlarm(npuDireAlarmList []any, rootCauseAlarmAll *[]any, faultType string) {
	// 剩余的ip集合初始化
	leftIps := initLeftIps(npuDireAlarmList)

	for {
		rootCauseIps := getRootCauseIps(leftIps)

		// 格式化告警信息并添加到总告警列表rootCauseAlarmAll里去
		tmpList := make([]string, 0)
		for {
			rootCauseAlarm := make(map[string]any)
			rootCauseAlarm[faultTypeConstant] = faultType

			faultPath := getCurFaultPathInfo(npuDireAlarmList, rootCauseIps, &tmpList)
			tmpList = uniqueSlice(tmpList)
			setCommonAlarmFormat(faultPath, rootCauseAlarm)
			setNpuDireAlarmFormat(rootCauseIps, tmpList, rootCauseAlarm)

			// 告警抑制
			if !isAlarmSuppressed(rootCauseAlarm) {
				*rootCauseAlarmAll = append(*rootCauseAlarmAll, rootCauseAlarm)
			}

			rootCauseIps = removeElements(rootCauseIps, tmpList)

			// 内循环结束条件
			if len(rootCauseIps) == 0 {
				break
			}
		}

		// 循环遍历条件，逐渐减少list的元素个数
		leftIps = removeElements(leftIps, tmpList)

		// 外循环结束条件
		if len(leftIps) == 0 {
			break
		}
	}
}

// 初始化剩余的ip
func initLeftIps(npuDireAlarmList []any) []string {
	leftIps := make([]string, 0)

	npuDireLen := len(npuDireAlarmList)
	for i := 0; i < npuDireLen; i += 1 {
		path, ok := npuDireAlarmList[i].(map[string]any)
		srcAddr, ok := path[srcAddrConstant].(string)
		dstAddr, ok := path[dstAddrConstant].(string)
		if !ok {
			continue
		}

		leftIps = append(leftIps, srcAddr)
		leftIps = append(leftIps, dstAddr)
	}

	return leftIps
}

// 从列表中获取根因ip
func getRootCauseIps(ips []string) []string {
	rootCauseIps := make([]string, 0)

	mostAppearIp, countMap := countForSlice(ips)
	mostNum := countMap[mostAppearIp]
	for key, value := range countMap {
		if value == mostNum {
			rootCauseIps = append(rootCauseIps, key)
		}
	}

	return rootCauseIps
}

// 获取当前层级故障路径信息
func getCurFaultPathInfo(npuDireAlarmList []any, rootCauseIps []string, tmpList *[]string) []any {
	faultPath := make([]any, 0)

	if len(rootCauseIps) == 0 {
		hwlog.RunLog.Info("[ALGO] rootCauseIps is empty in npuDireAlarmList")
		return faultPath
	}

	for i := 0; i < len(npuDireAlarmList); i++ {
		path, ok := npuDireAlarmList[i].(map[string]any)
		srcAddr, ok := path[srcAddrConstant].(string)
		dstAddr, ok := path[dstAddrConstant].(string)
		if !ok {
			continue
		}

		if rootCauseIps[0] == srcAddr || rootCauseIps[0] == dstAddr {
			faultPath = append(faultPath, path)
			*tmpList = append(*tmpList, srcAddr)
			*tmpList = append(*tmpList, dstAddr)
		}
	}

	return faultPath
}

// 设置npu直连告警信息
func setNpuDireAlarmFormat(rootCauseIps []string, curIps []string, rootCauseAlarm map[string]any) {
	if rootCauseAlarm == nil {
		return
	}

	if len(rootCauseIps) == 0 {
		hwlog.RunLog.Error("[NETFAULT ALGO]rootCauseIps is empty")
		return
	}

	if len(rootCauseIps) == 1 {
		rootCauseAlarm[srcIdConstant] = rootCauseIps[0]
		rootCauseAlarm[srcTypeConstant] = npuType
		rootCauseAlarm[dstIdConstant] = rootCauseIps[0]
		rootCauseAlarm[dstTypeConstant] = npuType
		rootCauseAlarm[levelConstant] = majorType
	} else {
		dstIp := ""
		for i := 1; i < len(rootCauseIps); i++ {
			if contains(curIps, rootCauseIps[i]) {
				dstIp = rootCauseIps[i]
				break
			}
		}

		if dstIp != "" {
			rootCauseAlarm[srcIdConstant] = rootCauseIps[0]
			rootCauseAlarm[srcTypeConstant] = npuType
			rootCauseAlarm[dstIdConstant] = dstIp
			rootCauseAlarm[dstTypeConstant] = npuType
			rootCauseAlarm[levelConstant] = minorType
		} else {
			rootCauseAlarm[srcIdConstant] = rootCauseIps[0]
			rootCauseAlarm[srcTypeConstant] = npuType
			rootCauseAlarm[dstIdConstant] = rootCauseIps[0]
			rootCauseAlarm[dstTypeConstant] = npuType
			rootCauseAlarm[levelConstant] = majorType
		}
	}
}

// 获取其他异常路径的根因告警，并添加到总根因告警里
func (nd *NetDetect) getOtherFaultAlarm(npuOtherAlarmList []any, rootCauseAlarmAll *[]any,
	detectType string) {

	// 区分平面
	netplaneAlarmList := make(map[string][]any)
	nd.diffNetplaneAlarmList(netplaneAlarmList, npuOtherAlarmList)

	for netplane, npuAlarmList := range netplaneAlarmList {
		nd.getNetplaneRootCauseAlarm(netplane, npuAlarmList, npuOtherAlarmList, rootCauseAlarmAll, detectType)
	}
}

func (nd *NetDetect) getNetplaneRootCauseAlarm(netplane string, npuAlarmList []any,
	npuOtherAlarmList []any, rootCauseAlarmAll *[]any, detectType string) {
	// 获取根因故障
	rootCauseFault := nd.getRootCauseAlarm(npuAlarmList)

	// 遍历rootCauseFault对象的所有键
	for layer := range globalRootCauseEventNpu {
		if eventLayer, exists := rootCauseFault[layer]; exists {
			for i := 0; i < len(eventLayer); i++ {
				rootCauseObj := eventLayer[i]
				rootCauseAlarm := make(map[string]any)
				rootCauseAlarm[faultTypeConstant] = detectType
				rootCauseAlarm[rootCauseConstant] = rootCauseObj
				setCommonAlarmFormat(npuOtherAlarmList, rootCauseAlarm)
				nd.setSingleFaultAlarm(rootCauseAlarm, netplane, npuOtherAlarmList)
				updateRootCauseAlarmAll(rootCauseAlarmAll, rootCauseAlarm)
			}
		}
	}
}

// 将告警路径按平面区分
func (nd *NetDetect) diffNetplaneAlarmList(netplaneAlarmList map[string][]any,
	npuOtherAlarmList []any) {
	if netplaneAlarmList == nil {
		return
	}

	for i := 0; i < len(npuOtherAlarmList); i++ {
		item, ok := npuOtherAlarmList[i].(map[string]any)
		if !ok {
			hwlog.RunLog.Error("[NETFAULT ALGO] wrong format of alarm list")
			return
		}

		srcAddr, ok := item[srcAddrConstant].(string)
		if !ok {
			hwlog.RunLog.Error("[NETFAULT ALGO] wrong format of srcAddr")
			return
		}

		netplaneName := nd.findNetplaneByPingObj(srcAddr)
		// 检查 netplaneName 是否存在于 netplaneAlarmList 中
		if _, exists := netplaneAlarmList[netplaneName]; !exists {
			netplaneAlarmList[netplaneName] = make([]any, 0)
		}
		netplaneAlarmList[netplaneName] = append(netplaneAlarmList[netplaneName], item)
	}
}

// 根据故障路径获取根因故障
func (nd *NetDetect) getRootCauseAlarm(npuOtherAlarmList []any) map[string][]string {
	rootCauseAlarm := make(map[string][]string)

	// 将list_alarm_all按layer分类
	dfAlarmAll := classifyByLayer(npuOtherAlarmList)
	lastPath := make([]map[string]any, 0)

	// 模糊告警标识（信息缺失场景使用）
	fuzzyAlarmFlag := false

	for layer, layerItem := range dfAlarmAll {
		// 排除上一层的故障路径
		excludeLastPaths(&layerItem, lastPath)
		if len(layerItem) == 0 {
			continue
		}

		// 对每一层再分类（L0层分框，L1层不用分）
		childDfAlarmAll := make(map[string]any)
		if layer == layer2Constant {
			childDfAlarmAll = classifyByRack(layerItem)
		} else {
			childDfAlarmAll[layer] = layerItem
		}

		rcaObjArray := make([]string, 0)
		for _, eachChildDfAlarm := range childDfAlarmAll {
			eachChildDfAlarmAll, ok := eachChildDfAlarm.([]map[string]any)
			if !ok {
				continue
			}

			rootCauseObj := nd.getRootCauseObj(eachChildDfAlarmAll, &rootCauseAlarm, &fuzzyAlarmFlag, layer)

			// 添加到根因对象集合
			rcaObjArray = append(rcaObjArray, rootCauseObj...)
		}

		// 组合根因对象
		rootCauseAlarm[layer] = rcaObjArray

		// 标记上一层故障路径
		if !fuzzyAlarmFlag {
			lastPath = append(lastPath, layerItem...)
		}
	}

	return rootCauseAlarm
}

// 单节点故障告警输出格式化
func (nd *NetDetect) setSingleFaultAlarm(rootCauseAlarm map[string]any, netplane string,
	npuOtherAlarmList []any) {
	if rootCauseAlarm == nil {
		return
	}

	rootCauseObj, ok := rootCauseAlarm[rootCauseConstant].(string)
	if !ok {
		return
	}

	delete(rootCauseAlarm, rootCauseConstant)

	if strings.Contains(rootCauseObj, fuzzyAlarmFlagChar) {
		setFuzzyFaultAlarm(rootCauseObj, rootCauseAlarm)
	} else if net.ParseIP(rootCauseObj) != nil {
		setNpuFaultAlarm(rootCauseObj, rootCauseAlarm)
	} else if strings.Contains(rootCauseObj, rackConstant) {
		nd.setRackFaultAlarm(rootCauseObj, rootCauseAlarm, netplane)
	} else if strings.Contains(rootCauseObj, nodeConstant) {
		nd.setNodeFaultAlarm(rootCauseObj, rootCauseAlarm, netplane)
	} else if strings.Contains(rootCauseObj, l2Constant) {
		if nd.curNpuType == a3NpuTypeConstant {
			nd.setA3L2FaultAlarm(rootCauseObj, rootCauseAlarm)
		} else {
			hwlog.RunLog.Error("[ALGO] unexpected detection type!")
		}
	} else if strings.Contains(rootCauseObj, superPodConstant) {
		nd.setSuperPodFaultAlarm(rootCauseObj, rootCauseAlarm, npuOtherAlarmList)
	} else if strings.Contains(rootCauseObj, roceSwitchConstant) {
		nd.setRoceSwitchFaultAlarm(rootCauseObj, rootCauseAlarm, npuOtherAlarmList)
	} else {
		hwlog.RunLog.Infof("[ALGO] rootCauseObj is: %v", rootCauseObj)
	}
}

// 更新最终告警（判断当前告警是否需要抑制）
func updateRootCauseAlarmAll(rootCauseAlarmAll *[]any, rootCauseAlarm map[string]any) {
	if !isAlarmSuppressed(rootCauseAlarm) {
		*rootCauseAlarmAll = append(*rootCauseAlarmAll, rootCauseAlarm)
	}
}

// 兼容模糊告警（缺少告警信息时，不展示具体的根因，只展示对应链路故障）
func setFuzzyFaultAlarm(rootCauseObj string, rootCauseAlarm map[string]any) {
	if rootCauseAlarm == nil {
		return
	}

	temp := strings.Split(rootCauseObj, fuzzyAlarmFlagChar)
	if len(temp) != baseSegmentNum {
		return
	}
	rootCauseAlarm[srcIdConstant] = temp[0]
	rootCauseAlarm[srcTypeConstant] = npuType
	rootCauseAlarm[dstIdConstant] = temp[1]
	rootCauseAlarm[dstTypeConstant] = npuType
	rootCauseAlarm[levelConstant] = majorType
}

// 单npu故障
func setNpuFaultAlarm(rootCauseObj string, rootCauseAlarm map[string]any) {
	if rootCauseAlarm == nil {
		return
	}

	rootCauseAlarm[srcIdConstant] = rootCauseObj
	rootCauseAlarm[srcTypeConstant] = npuType
	rootCauseAlarm[dstIdConstant] = rootCauseObj
	rootCauseAlarm[dstTypeConstant] = npuType
	rootCauseAlarm[levelConstant] = majorType
}

func (nd *NetDetect) setSrcAndDstForA3(rootCauseAlarm map[string]interface{}, rootCauseObj string, setSrc bool,
	setDst bool) {
	if rootCauseAlarm == nil {
		return
	}
	// A3场景是一个worker抽象成一个rack，所以这里rack级根因在A3场景下应该为npu卡上行到节点worker故障
	splitFirstSlice := strings.Split(rootCauseObj, "-") // 取出worker id
	if len(splitFirstSlice) != baseSegmentNum {
		hwlog.RunLog.Errorf("[ALGO] error split woker name: %s", rootCauseObj)
		return
	}
	// A3上升到worker情况
	if strings.Contains(splitFirstSlice[1], ":") {
		splitSecondSlice := strings.Split(splitFirstSlice[1], ":")
		if len(splitSecondSlice) != baseSegmentNum {
			hwlog.RunLog.Errorf("[ALGO] error split woker name: %s, %s", splitFirstSlice[1], rootCauseObj)
			return
		}
		splitFirstSlice[1] = splitSecondSlice[0]
	}
	if workerName, exit := nd.curServerIdMap[splitFirstSlice[1]]; exit {
		if setDst {
			rootCauseAlarm[dstIdConstant] = workerName
			rootCauseAlarm[dstTypeConstant] = workNodeType
		}
		if setSrc {
			rootCauseAlarm[srcIdConstant] = workerName
			rootCauseAlarm[srcTypeConstant] = workNodeType
		}
	} else {
		hwlog.RunLog.Errorf("[ALGO] error to get worker name: %s form map: %v", rootCauseObj, nd.curServerIdMap)
	}
}

// 框里交换机故障
func (nd *NetDetect) setRackFaultAlarm(rootCauseObj string, rootCauseAlarm map[string]any, netplane string) {
	if rootCauseAlarm == nil {
		return
	}

	// 框里交换机的下行网口故障
	if strings.Contains(rootCauseObj, portIntervalChar) {
		npuRackList := strings.Split(rootCauseObj, portIntervalChar)
		if len(npuRackList) != baseSegmentNum {
			return
		}
		// A3场景下Rack-i:Sdid而非npu物理ID，Sdid是一个string
		npuNumber, err := strconv.Atoi(npuRackList[1])
		if err != nil {
			return
		}

		rackName := strings.Split(npuRackList[0], dotIntervalChar)[0]
		srcId := nd.findPingObjByNpuNumber(rackName, npuNumber, netplane)
		rootCauseAlarm[srcIdConstant] = srcId
		rootCauseAlarm[srcTypeConstant] = npuType
		if nd.curNpuType == a3NpuTypeConstant {
			nd.setSrcAndDstForA3(rootCauseAlarm, rootCauseObj, false, true)
		} else {
			rootCauseAlarm[dstIdConstant] = npuRackList[0]
			rootCauseAlarm[dstTypeConstant] = rackNetplaneType
		}
		rootCauseAlarm[levelConstant] = minorType
	} else {
		// 框里交换机本身故障
		if nd.curNpuType == a3NpuTypeConstant {
			nd.setSrcAndDstForA3(rootCauseAlarm, rootCauseObj, true, true)
		} else {
			rootCauseAlarm[srcIdConstant] = rootCauseObj
			rootCauseAlarm[dstIdConstant] = rootCauseObj
			rootCauseAlarm[srcTypeConstant] = rackNetplaneType
			rootCauseAlarm[dstTypeConstant] = rackNetplaneType
		}
		rootCauseAlarm[levelConstant] = criticalType
	}
}

func (nd *NetDetect) setNodeFaultAlarm(rootCauseObj string, rootCauseAlarm map[string]any, netplane string) {
	if rootCauseAlarm == nil {
		return
	}

	// 节点所属L1的下行网口故障
	if strings.Contains(rootCauseObj, portIntervalChar) {
		npuRackList := strings.Split(rootCauseObj, portIntervalChar)
		if len(npuRackList) != baseSegmentNum {
			return
		}
		npuNumber, err := strconv.Atoi(npuRackList[1])
		if err != nil {
			return
		}

		tmpNodeName := strings.Split(npuRackList[0], dotIntervalChar)[0]
		nodeName := strings.ReplaceAll(tmpNodeName, "Node-", "")
		srcId := nd.findPingObjByNpuNumberWithOsName(nodeName, npuNumber, netplane)
		rootCauseAlarm[srcIdConstant] = srcId
		rootCauseAlarm[srcTypeConstant] = npuType
		rootCauseAlarm[dstIdConstant] = npuRackList[0]
		rootCauseAlarm[dstTypeConstant] = workNodeType
		rootCauseAlarm[levelConstant] = minorType
	} else {
		// 节点所属L1本身故障
		rootCauseAlarm[srcIdConstant] = rootCauseObj
		rootCauseAlarm[srcTypeConstant] = workNodeType
		rootCauseAlarm[dstIdConstant] = rootCauseObj
		rootCauseAlarm[dstTypeConstant] = workNodeType
		rootCauseAlarm[levelConstant] = criticalType
	}
}

// 框外交换机故障
func (nd *NetDetect) setA3L2FaultAlarm(rootCauseObj string, rootCauseAlarm map[string]any) {
	if rootCauseAlarm == nil {
		return
	}

	// 框外交换机的下行网口故障
	if strings.Contains(rootCauseObj, portIntervalChar) {
		rackL1List := strings.Split(rootCauseObj, portIntervalChar)
		if len(rackL1List) != baseSegmentNum {
			return
		}

		srcId, exists := nd.curServerIdMap[rackL1List[1]]
		if !exists {
			hwlog.RunLog.Errorf("[ALGO] can't find key: %s in nd.curServerIdMap", rackL1List[1])
			return
		}

		rootCauseAlarm[srcIdConstant] = srcId
		rootCauseAlarm[srcTypeConstant] = workNodeType
		rootCauseAlarm[dstIdConstant] = rackL1List[0]
		rootCauseAlarm[dstTypeConstant] = l2NetplaneType
		rootCauseAlarm[levelConstant] = majorType
	} else {
		// 框外交换机本身故障
		rootCauseAlarm[srcIdConstant] = rootCauseObj
		rootCauseAlarm[srcTypeConstant] = l2NetplaneType
		rootCauseAlarm[dstIdConstant] = rootCauseObj
		rootCauseAlarm[dstTypeConstant] = l2NetplaneType
		rootCauseAlarm[levelConstant] = criticalType
	}
}

// 超节点内下行故障
func (nd *NetDetect) setSuperPodFaultAlarm(rootCauseObj string, rootCauseAlarm map[string]any,
	npuOtherAlarmList []any) {
	if rootCauseAlarm == nil {
		return
	}

	rootCauseObjArr := strings.Split(rootCauseObj, portIntervalChar)

	const rootCauseObjNum1 = 1 // 根因对象个数为1
	const rootCauseObjNum2 = 2 // 根因对象个数为2
	const rootCauseObjNum3 = 4 // 根因对象个数为4

	if len(rootCauseObjArr) == rootCauseObjNum1 {
		// 大网下行超节点内上行故障
		rootCauseAlarm[srcIdConstant] = rootCauseObj
		rootCauseAlarm[srcTypeConstant] = superPodType
		rootCauseAlarm[dstIdConstant] = roceSwitchConstant
		rootCauseAlarm[dstTypeConstant] = roceSwitchType
		rootCauseAlarm[levelConstant] = criticalType
		desc := nd.getRoceSwitchDesc(npuOtherAlarmList, rootCauseObj)
		rootCauseAlarm[descriptionConstant] = rootCauseObj + desc
		return
	}

	if len(rootCauseObjArr) == rootCauseObjNum2 {
		// ip上行到超节点故障
		npuNumber, err := strconv.Atoi(rootCauseObjArr[1])
		if err != nil {
			return
		}

		superPodName := strings.Split(rootCauseObjArr[0], dotIntervalChar)[0]
		srcId := nd.findPingObjBySuperPod(superPodName, npuNumber)
		rootCauseAlarm[srcIdConstant] = srcId
		rootCauseAlarm[srcTypeConstant] = nd.curPingObjType
		rootCauseAlarm[dstIdConstant] = rootCauseObjArr[0]
		rootCauseAlarm[dstTypeConstant] = superPodType
		rootCauseAlarm[levelConstant] = minorType
		return
	}

	if len(rootCauseObjArr) == rootCauseObjNum3 {
		// 超节点间链路故障
		rootCauseAlarm[srcIdConstant] = rootCauseObjArr[0]
		rootCauseAlarm[srcTypeConstant] = superPodType
		rootCauseAlarm[dstIdConstant] = rootCauseObjArr[1]
		rootCauseAlarm[dstTypeConstant] = superPodType
		rootCauseAlarm[levelConstant] = criticalType
		srcIpDesc := "srcIps: " + rootCauseObjArr[2]
		dstIpsDesc := "dstIps: " + rootCauseObjArr[3]
		rootCauseAlarm[descriptionConstant] = srcIpDesc + ", " + dstIpsDesc
		return
	}
}

// 大网内下行超节点上行故障
func (nd *NetDetect) setRoceSwitchFaultAlarm(rootCauseObj string, rootCauseAlarm map[string]any,
	npuOtherAlarmList []any) {
	if rootCauseAlarm == nil {
		return
	}

	// 大网下行超节点内上行故障
	if strings.Contains(rootCauseObj, portIntervalChar) {
		npuRoceSwitchList := strings.Split(rootCauseObj, portIntervalChar)
		if len(npuRoceSwitchList) != baseSegmentNum {
			return
		}

		superPodName := superPodConstant + normalIntervalChar + npuRoceSwitchList[1]
		rootCauseAlarm[srcIdConstant] = superPodName
		rootCauseAlarm[srcTypeConstant] = superPodType
		rootCauseAlarm[dstIdConstant] = npuRoceSwitchList[0]
		rootCauseAlarm[dstTypeConstant] = roceSwitchType
		rootCauseAlarm[levelConstant] = criticalType
		desc := nd.getRoceSwitchDesc(npuOtherAlarmList, superPodName)
		rootCauseAlarm[descriptionConstant] = superPodName + desc
	} else {
		// 超节点本身故障
		rootCauseAlarm[srcIdConstant] = rootCauseObj
		rootCauseAlarm[srcTypeConstant] = roceSwitchType
		rootCauseAlarm[dstIdConstant] = rootCauseObj
		rootCauseAlarm[dstTypeConstant] = roceSwitchType
		rootCauseAlarm[levelConstant] = criticalType
	}
}

// 大网下行故障描述信息
func (nd *NetDetect) getRoceSwitchDesc(npuOtherAlarmList []any, superPodName string) string {
	allIpList := make([]string, 0)
	rootCauseIpList := make([]string, 0)

	for i := 0; i < len(npuOtherAlarmList); i++ {
		curPath, ok := npuOtherAlarmList[i].(map[string]any)
		if !ok {
			hwlog.RunLog.Error("[ALGO] type assertion failed")
		}

		allIpList = append(allIpList, curPath[srcAddrConstant].(string))
		allIpList = append(allIpList, curPath[dstAddrConstant].(string))
	}

	mostAppearIp, countMap := countForSlice(allIpList)
	mostNum := countMap[mostAppearIp]

	// 将次数最大的path加入到rootCauseIpList
	for key, value := range countMap {
		if value == mostNum && nd.curNpuInfo[key].SuperPodName == superPodName {
			rootCauseIpList = append(rootCauseIpList, key)
		}
	}

	desc := " fault ips: [" + strings.Join(rootCauseIpList, ", ") + "]"
	return desc
}

// 按层分类cJSON对象并输出，key为类名
func classifyByLayer(input []any) map[string][]map[string]any {
	output := make(map[string][]map[string]any)
	layer1Arr := make([]map[string]any, 0)
	layer2Arr := make([]map[string]any, 0)
	layer3Arr := make([]map[string]any, 0)

	for i := 0; i < len(input); i++ {
		// 区分分类条件
		item, ok := input[i].(map[string]any)
		info, ok := item[informationConstant].(string)
		if !ok {
			continue
		}

		infoArr := strings.Split(info, portIntervalChar)
		layer := infoArr[0]

		if layer == layer1Constant {
			item[layerConstant] = layer1Constant
			layer1Arr = append(layer1Arr, item)
		} else if layer == layer2Constant {
			item[layerConstant] = layer2Constant
			layer2Arr = append(layer2Arr, item)
		} else if layer == layer3Constant {
			item[layerConstant] = layer3Constant
			layer3Arr = append(layer3Arr, item)
		}
	}

	// 注意下面的添加顺序，会影响检测结果，因为这个添加顺序就是后面的遍历顺序
	output[layer1Constant] = layer1Arr
	output[layer2Constant] = layer2Arr
	output[layer3Constant] = layer3Arr
	return output
}

// 按框分类cJSON对象并输出，key为框的名字
func classifyByRack(input []map[string]any) map[string]any {
	output := make(map[string]any)
	for i := 0; i < len(input); i++ {
		item := input[i]
		fromLayer, ok := item[fromLayerConstant].(string)
		if !ok {
			continue
		}

		tmp := strings.Split(fromLayer, layerIntervalChar)
		if len(tmp) < baseSegmentNum {
			continue
		}

		rackName, ok := getCurPathRackName(tmp)
		if !ok {
			continue
		}

		if value, exists := output[rackName]; exists {
			value = append(value.([]map[string]any), item)
			output[rackName] = value
		} else {
			newRack := make([]map[string]any, 0)
			newRack = append(newRack, item)
			output[rackName] = newRack
		}
	}

	return output
}

func getCurPathRackName(rackArr []string) (string, bool) {
	var rackName string
	for j := 0; j < len(rackArr); j++ {
		if strings.Contains(rackArr[j], rackConstant) {
			tmpRackArr := strings.Split(rackArr[j], portIntervalChar)
			if len(tmpRackArr) != baseSegmentNum {
				return "", false
			}

			rackName = tmpRackArr[0]
			break
		}
	}

	return rackName, true
}

// 排除上一层故障路径
func excludeLastPaths(curPath *[]map[string]any, lastPath []map[string]any) {
	rackPath := make([]string, 0)
	for i := 0; i < len(lastPath); i++ {
		eachLastPath := lastPath[i]
		from, ok := eachLastPath[fromLayerConstant].(string)
		to, ok := eachLastPath[toLayerConstant].(string)
		if !ok {
			continue
		}

		fromTmp := strings.Split(from, layerIntervalChar)
		toTmp := strings.Split(to, layerIntervalChar)
		if len(fromTmp) < baseSegmentNum || len(toTmp) < baseSegmentNum {
			continue
		}

		rackPath = append(rackPath, fromTmp[1])
		rackPath = append(rackPath, toTmp[1])
	}

	for i := 0; i < len(*curPath); {
		eachCurItem := (*curPath)[i]
		from, ok := eachCurItem[fromLayerConstant].(string)
		to, ok := eachCurItem[toLayerConstant].(string)
		if !ok {
			continue
		}

		fromRack := strings.Split(from, layerIntervalChar)[1]
		toRack := strings.Split(to, layerIntervalChar)[1]
		if contains(rackPath, fromRack) || contains(rackPath, toRack) {
			removeAt(curPath, i)
		} else {
			i++
		}
	}
}

// 移除切片中指定索引的元素
func removeAt(slice *[]map[string]any, index int) {
	if index < 0 || index >= len(*slice) {
		return
	}

	// 使用 append 函数移除指定索引的元素
	*slice = append((*slice)[:index], (*slice)[index+1:]...)
	return
}

// 获取故障路径（faultPathList：链路级别的故障， swFaultPathList：对象级别的故障）
func setFaultPathList(eachChildDfAlarmAll []map[string]any, faultPathList *[][]string,
	swFaultPathList *[][]string) {
	length := len(eachChildDfAlarmAll)
	for i := 0; i < length; i++ {
		(*faultPathList)[i] = make([]string, 0)
		(*swFaultPathList)[i] = make([]string, 0)
	}

	index := 0
	for i := 0; i < length; i++ {
		item := eachChildDfAlarmAll[i]
		desc, ok := item[descriptionConstant].(string)
		if !ok {
			continue
		}

		desc = strings.ReplaceAll(desc, "S-", "")
		desc = strings.ReplaceAll(desc, "D-", "")
		tmpDesc := strings.Split(desc, " TO ")
		descList := make([]string, 0)
		for j := 0; j < len(tmpDesc); j++ {
			tmpDescArr := strings.Split(tmpDesc[j], layerIntervalChar)
			descList = append(descList, tmpDescArr...)
		}

		uniqueElements := appearOnce(descList)
		(*faultPathList)[index] = uniqueElements
		index++
	}

	// 交换机对象打分，同一条路径去掉port后，相同的交换机去重
	for i := 0; i < length; i++ {
		tmpFaultPathItem := make([]string, 0)
		faultPathItem := (*faultPathList)[i]
		for j := 0; j < len(faultPathItem); j++ {
			obj := faultPathItem[j]
			arr := strings.Split(obj, portIntervalChar)
			tmpFaultPathItem = append(tmpFaultPathItem, arr[0])
		}
		uniqueList := uniqueSlice(tmpFaultPathItem)
		(*swFaultPathList)[i] = uniqueList
	}
}

// 设置根因对象列表
func (nd *NetDetect) setRootCauseList(rootCauseList *[]string, faultPathList [][]string, swFaultPathList [][]string) {
	if len(faultPathList) != len(swFaultPathList) {
		return
	}

	faultPathListBak := make([]string, 0)
	length := len(faultPathList)
	for i := 0; i < length; i++ {
		faultPathListBak = append(faultPathListBak, faultPathList[i]...)
	}

	mostAppearIp, countMap := countForSlice(faultPathListBak)
	mostNum := countMap[mostAppearIp]
	if mostNum < length {
		for i := 0; i < length; i++ {
			tempList := swFaultPathList[i]
			faultPathListBak = append(faultPathListBak, tempList...)
		}
		mostAppearIp, countMap = countForSlice(faultPathListBak)
		mostNum = countMap[mostAppearIp]
	}

	// 将次数最大的path加入到rootCauseList
	for key, value := range countMap {
		if value == mostNum {
			*rootCauseList = append(*rootCauseList, key)
		}
	}
}

// 将字符串中的数字减1，例如输入"layer_3"，输出"layer_2"
func getLastLayer(input string) string {
	// 使用正则表达式查找字符串中的数字
	re := regexp.MustCompile(`\d+`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		// 将匹配到的数字转换为整数
		num, err := strconv.Atoi(match)
		if err != nil {
			return match // 如果转换出错，返回原始匹配
		}
		// 减 1 并返回字符串形式
		return strconv.Itoa(num - 1)
	})
}

// 获取故障根因对象
func (nd *NetDetect) getRootCauseObj(eachChildDfAlarmAll []map[string]any, rootCauseAlarm *map[string][]string,
	fuzzyAlarmFlag *bool, layer string) []string {
	rootCauseObj := make([]string, 0)

	// 分层分框后只有一条故障路径，标识模糊告警，不打分聚合
	if len(eachChildDfAlarmAll) == 1 {
		*fuzzyAlarmFlag = true
		onceAlarm := eachChildDfAlarmAll[0]
		srcAddr, ok := onceAlarm[srcAddrConstant].(string)
		dstAddr, ok := onceAlarm[dstAddrConstant].(string)
		if !ok {
			return rootCauseObj
		}

		rootCauseObj = append(rootCauseObj, fmt.Sprintf("%s%s%s", srcAddr, fuzzyAlarmFlagChar, dstAddr))
	} else {
		length := len(eachChildDfAlarmAll)
		rootCauseList := make([]string, 0)
		faultPathList := make([][]string, length)
		swFaultPathList := make([][]string, length)
		setFaultPathList(eachChildDfAlarmAll, &faultPathList, &swFaultPathList)
		nd.setRootCauseList(&rootCauseList, faultPathList, swFaultPathList)

		// 根因节点定界的时候区分一下是否是超节点间探测任务
		if nd.curSuperPodJobFlag {
			rootCauseList = nd.getSuperPodRootCauseList(rootCauseList)
			rootCauseObj = append(rootCauseObj, rootCauseList...)
		} else {
			rootCauseList = nd.getCurRootCauseList(rootCauseList, rootCauseAlarm, layer)
			tmpRootCauseObj := getRootCauseEvent(rootCauseList, globalRootCauseEventNpu)
			rootCauseObj = append(rootCauseObj, tmpRootCauseObj)
		}
	}

	return rootCauseObj
}

// 获取超节点间探测任务的根因列表
func (nd *NetDetect) getSuperPodRootCauseList(input []string) []string {
	// 大网下行超节点上行，只有一个的话，直接作为根因列表返回
	if len(input) == 1 {
		return input
	}

	superPodArr := make([]string, 0)
	ipArr := make([]string, 0)
	superPodMap := make(map[string][]string)
	result := make([]string, 0)

	for _, str := range input {
		if strings.Contains(str, superPodConstant) {
			superPodArr = append(superPodArr, str)
		} else {
			ipArr = append(ipArr, str)
		}
	}

	// 返回"SuperPod-1:1"或"SuperPod-1", ip上行到超节点
	if len(superPodArr) == 1 {
		return superPodArr
	}

	for i := 0; i < len(superPodArr); i++ {
		nd.matchRootCauseArr(superPodArr[i], &ipArr, superPodMap)
	}

	// 返回"SuperPod-0:SuperPod-1", 超节点链路故障
	sort.Strings(superPodArr)
	for i := 0; i < len(superPodArr); i++ {
		var curSuperPodName string
		var nextSuperPodName string
		if i == len(superPodArr)-1 {
			curSuperPodName = superPodArr[i]
			nextSuperPodName = superPodArr[0]
		} else {
			curSuperPodName = superPodArr[i]
			nextSuperPodName = superPodArr[i+1]
		}

		if getIndex(nd.curSuperPodArr, curSuperPodName)+1 == getIndex(nd.curSuperPodArr, nextSuperPodName) ||
			(getIndex(nd.curSuperPodArr, curSuperPodName) == len(nd.curSuperPodArr)-1 &&
				getIndex(nd.curSuperPodArr, nextSuperPodName) == 0) {
			srcIpsDesc := "[" + strings.Join(superPodMap[curSuperPodName], ", ") + "]"
			dstIpsDesc := "[" + strings.Join(superPodMap[nextSuperPodName], ", ") + "]"
			result = append(result, curSuperPodName+portIntervalChar+nextSuperPodName+
				portIntervalChar+srcIpsDesc+portIntervalChar+dstIpsDesc)
		}
	}

	return result
}

func (nd *NetDetect) matchRootCauseArr(superPodName string, ipArr *[]string, superPodMap map[string][]string) {
	if superPodMap == nil {
		return
	}

	for j := 0; j < len(*ipArr); j++ {
		if nd.curNpuInfo[(*ipArr)[j]].SuperPodName == superPodName {
			// 如果 superPodMap 还没有这个键，初始化为一个空切片
			if _, exists := superPodMap[superPodName]; !exists {
				superPodMap[superPodName] = []string{}
			}
			superPodMap[superPodName] = append(superPodMap[superPodName], (*ipArr)[j])
		}
	}

	*ipArr = removeElements(*ipArr, superPodMap[superPodName])
}

// 获取当前层的根因列表
func (nd *NetDetect) getCurRootCauseList(rootCauseList []string, rootCauseAlarm *map[string][]string,
	layer string) []string {
	rootCauseListTmp := make([]string, 0)

	lastLayer := getLastLayer(layer)
	if _, exists := (*rootCauseAlarm)[lastLayer]; exists {
		for _, item := range rootCauseList {
			if !contains(rootCauseListTmp, item) {
				rootCauseListTmp = append(rootCauseListTmp, item)
			}
		}
		return rootCauseListTmp
	}

	for i := 0; i < len(rootCauseList); i++ {
		value, exist := globalRootCauseEventNpu[lastLayer]
		if !exist {
			rootCauseListTmp = append(rootCauseListTmp, rootCauseList[i])
			continue
		}

		if value == npuConstant {
			if _, exists := nd.curNpuInfo[value]; !exists {
				rootCauseListTmp = append(rootCauseListTmp, rootCauseList[i])
			}
			continue
		}

		if !strings.Contains(rootCauseList[i], value) {
			rootCauseListTmp = append(rootCauseListTmp, rootCauseList[i])
		}
	}

	return rootCauseListTmp
}

// 获取根因对象
func getRootCauseEvent(rootCauseList []string, rootCauseEvent map[string]string) string {
	for _, eventValue := range rootCauseEvent {
		for _, rootCause := range rootCauseList {
			// 跳过包含 NSlot 和 portIntervalChar 的 rootCause
			if strings.Contains(rootCause, nSlotConstant) && strings.Contains(rootCause, portIntervalChar) {
				continue
			}

			// 分割 rootCause
			parts := strings.Split(rootCause, portIntervalChar)
			if len(parts) == 0 {
				continue
			}

			// 检查 eventValue 是否为 npuConstant 且 parts[0] 是合法 IP
			if eventValue == npuConstant && net.ParseIP(parts[0]) != nil {
				return rootCause
			}

			// 检查 parts[0] 是否包含 eventValue
			if strings.Contains(parts[0], eventValue) {
				return rootCause
			}
		}
	}

	return ""
}

// 根据npu编号和框号获取ping探测目标(ip、eid、sdid)
func (nd *NetDetect) findPingObjByNpuNumber(rackName string, npuNumber int, netplane string) string {
	if len(nd.curNpuInfo) == 0 {
		hwlog.RunLog.Infof("[ALGO] no npuInfoMap found in superPodId: %s", nd.curSuperPodId)
		return ""
	}

	for pingObj, npuInfo := range nd.curNpuInfo {
		if nd.curNpuType == a3NpuTypeConstant {
			npuSdId, err := strconv.Atoi(npuInfo.IP)
			if err != nil {
				hwlog.RunLog.Errorf("[ALGO] string to int failed: %v", err)
				return ""
			}
			if npuInfo.RackName == rackName && npuSdId == npuNumber && npuInfo.NetPlaneId == netplane {
				return pingObj
			}
		}
		if nd.curNpuType != a3NpuTypeConstant &&
			npuInfo.RackName == rackName && npuInfo.NpuNumber == npuNumber && npuInfo.NetPlaneId == netplane {
			return pingObj
		}
	}

	return ""
}

// 根据npu编号和os编号获取ping探测目标(ip、eid、sdid)
func (nd *NetDetect) findPingObjByNpuNumberWithOsName(osName string, npuNumber int, netplane string) string {
	if len(nd.curNpuInfo) == 0 {
		hwlog.RunLog.Infof("[ALGO] no npuInfoMap found in superPodId: %s", nd.curSuperPodId)
		return ""
	}

	for pingObj, npuInfo := range nd.curNpuInfo {
		if npuInfo.OsName == osName && npuInfo.NpuNumber == npuNumber && npuInfo.NetPlaneId == netplane {
			return pingObj
		}
	}

	return ""
}

// 根据npu编号和超节点号获取ping探测目标(ip、eid、sdid)
func (nd *NetDetect) findPingObjBySuperPod(superPodName string, npuNumber int) string {
	if len(nd.curNpuInfo) == 0 {
		hwlog.RunLog.Infof("[ALGO] no npuInfoMap found in superPodId: %s", nd.curSuperPodId)
		return ""
	}

	for pingObj, npuInfo := range nd.curNpuInfo {
		if npuInfo.SuperPodName == superPodName && npuInfo.NpuNumber == npuNumber {
			return pingObj
		}
	}

	return ""
}

// 根据ping探测目标(ip、eid、sdid)获取平面信息
func (nd *NetDetect) findNetplaneByPingObj(pingObj string) string {
	if len(nd.curNpuInfo) == 0 {
		hwlog.RunLog.Infof("[ALGO] no npuInfoMap found in superPodId: %s", nd.curSuperPodId)
		return ""
	}

	if npuInfo, exists := nd.curNpuInfo[pingObj]; exists {
		return npuInfo.NetPlaneId
	}

	return ""
}
