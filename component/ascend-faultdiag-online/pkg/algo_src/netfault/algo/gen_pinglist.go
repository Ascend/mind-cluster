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
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"ascend-common/common-utils/hwlog"
)

// SetFaultDetectParam 拨测参数入口
func (nd *NetDetect) SetFaultDetectParam(paramsMap map[string]any, npuInfoMap map[string]NpuInfo) bool {
	if nd == nil {
		return false
	}
	if len(paramsMap) == 0 {
		hwlog.RunLog.Error("[ALGO] paramsMap is empty")
		return false
	}

	var res bool
	res = checkParamsMap(paramsMap)
	if !res {
		hwlog.RunLog.Error("[ALGO] wrong params of paramsMap")
		return false
	}

	res = checkNpuAInfoMap(npuInfoMap)
	if !res {
		hwlog.RunLog.Error("[ALGO] wrong params of npuInfoMap")
		return false
	}

	res = nd.setParams(paramsMap, npuInfoMap)
	if !res {
		hwlog.RunLog.Error("[ALGO] failed to SetFaultDetectParam")
		return false
	}

	return true
}

// GenPingStrategy 拨测算法入口
func (nd *NetDetect) GenPingStrategy(input map[string]any) map[string]any {
	if nd == nil {
		return nil
	}
	// 1，初始化拨测策略和分组结构体
	aiPingStrategy := new(AiPingStrategy)
	initAiPingStrategy(aiPingStrategy)

	// 2，处理算法输入
	if !nd.processInput(input, aiPingStrategy) {
		hwlog.RunLog.Error("[ALGO] processInput failed")
		return nil
	}

	// 3，生成拨测策略
	dfChainsMap := make(map[string]*DataFrame)
	if !setDfChainMap(aiPingStrategy.chainList, dfChainsMap) {
		hwlog.RunLog.Error("[ALGO] setDfChainMap failed")
		return nil
	}

	if !nd.setPingDict(aiPingStrategy, dfChainsMap) {
		hwlog.RunLog.Error("[ALGO] setPingDict failed")
		return nil
	}

	// 4，处理拨测算法输出
	output := make(map[string]any)
	if !nd.processOutput(output, aiPingStrategy) {
		hwlog.RunLog.Error("[ALGO] processOutput failed")
		return nil
	}

	// 5，返回拨测策略结果
	return output
}

// 检测paramsMap合法性
func checkParamsMap(paramsMap map[string]any) bool {
	// 定义需要检查的键
	requiredKeys := []string{argsPeriod, argsSPeriod, argsPingObjType}

	// 遍历检查每个键是否存在
	for _, key := range requiredKeys {
		if _, exists := paramsMap[key]; !exists {
			return false
		}
	}

	return true
}

// 检测npuInfoMap合法性
func checkNpuAInfoMap(npuInfoMap map[string]NpuInfo) bool {
	return len(npuInfoMap) > 0
}

// 设置算法具体参数
func (nd *NetDetect) setParams(paramsMap map[string]any, npuInfoMap map[string]NpuInfo) bool {
	res1 := nd.setNecessaryParams(paramsMap, npuInfoMap)
	res2 := nd.setDefaultParams(paramsMap)
	return res1 && res2
}

// 设置算法必须参数
func (nd *NetDetect) setNecessaryParams(paramsMap map[string]any, npuInfoMap map[string]NpuInfo) bool {
	curPingPeriod, ok1 := paramsMap[argsPeriod].(int)
	curlSuppressedPeriod, ok2 := paramsMap[argsSPeriod].(int)
	curlPingObjType, ok3 := paramsMap[argsPingObjType].(int)
	curServerIdMap, ok4 := paramsMap[argsServerIdMap].(map[string]string)
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return false
	}

	nd.curPingPeriod = curPingPeriod
	nd.curSuppressedPeriod = curlSuppressedPeriod
	nd.curPingObjType = curlPingObjType
	nd.curServerIdMap = curServerIdMap
	nd.curNpuInfo = npuInfoMap

	return true
}

// 设置算法默认参数
func (nd *NetDetect) setDefaultParams(paramsMap map[string]any) bool {
	// 不传选轴策略的话，默认使用跨轴策略（可传"cross_axis", "both_axis", "same_axis"）
	if paramsMap[argsAxisStrategy] != nil {
		axisStrategy, ok := paramsMap[argsAxisStrategy].(string)
		if !ok || (axisStrategy != crossAxisConstant && axisStrategy != bothAxisConstant &&
			axisStrategy != sameAxisConstant) {
			hwlog.RunLog.Error("[ALGO] unexpected arg of axisStrategy")
			return false
		}

		nd.curAxisStrategy = axisStrategy
	} else {
		nd.curAxisStrategy = crossAxisConstant
	}

	// If npu_type is not passed, the default is A5 topology (alternate parameters: A5, A3)
	if paramsMap[argsNpuType] != nil {
		npuTopoType, ok := paramsMap[argsNpuType].(string)
		if !ok || (npuTopoType != a3NpuTypeConstant && npuTopoType != a5NpuTypeConstant) {
			hwlog.RunLog.Error("[ALGO] unexpected arg of npu_type")
			return false
		}

		nd.curNpuType = npuTopoType
	} else {
		nd.curNpuType = a5NpuTypeConstant
	}

	// 不传superPodJobFlag的话，默认是超节点内探测任务拓扑（可传true, false）
	if paramsMap[argsSuperPodJobFlag] != nil {
		superPodJobFlag, ok := paramsMap[argsSuperPodJobFlag].(bool)
		if !ok {
			hwlog.RunLog.Error("[ALGO] unexpected arg of superPodJobFlag")
			return false
		}

		nd.curSuperPodJobFlag = superPodJobFlag
	}

	// 不传superPodArr的话，默认是空切片
	if paramsMap[argsSuperPodArr] != nil {
		superPodArr, ok := paramsMap[argsSuperPodArr].([]string)
		if !ok {
			hwlog.RunLog.Error("[ALGO] unexpected arg of superPodArr")
			return false
		}

		nd.curSuperPodArr = superPodArr
	}

	return true
}

// 初始化拨测策略相关参数
func initAiPingStrategy(aiPingStrategy *AiPingStrategy) {
	aiPingStrategy.npuNpuList = make([]string, 0)
	aiPingStrategy.chainList = make(map[string][]string)
	aiPingStrategy.pingList = make([]string, 0)
	aiPingStrategy.layersIps = make(map[string]any)
	aiPingStrategy.dfGrouped = new(DataFrameGroupBy)
	aiPingStrategy.pingDict = make(map[string]any)
}

// 初始化链路拓扑数据
func initDfChain() *DataFrame {
	dfChains := new(DataFrame)
	dfChains.chains = make(map[string]any)
	dfChains.columnNames = make([]string, 0)

	return dfChains
}

// 处理算法输入
func (nd *NetDetect) processInput(input map[string]any, aiPingStrategy *AiPingStrategy) bool {
	if aiPingStrategy == nil {
		return false
	}

	npu2NetPlane, ok1 := input[argsNpu2NetPlane].(map[string]any)
	npu2Npu, ok2 := input[argsNpu2Npu].([]string)
	npu2SuperPod, ok3 := input[argsNpu2SuperPod].(map[string]any)
	if !ok1 || !ok2 || !ok3 {
		hwlog.RunLog.Error("[ALGO] type assertion failed")
		return false
	}

	for netplaneName, fullPath := range npu2NetPlane {
		planeFullPath, res := fullPath.([]string)
		if !res {
			hwlog.RunLog.Error("[ALGO] type assertion failed")
			return false
		}
		nd.setNpuFullPath(netplaneName, planeFullPath, aiPingStrategy)
	}

	for superPodName, podPath := range npu2SuperPod {
		superPodPath, res := podPath.([]string)
		if !res {
			hwlog.RunLog.Error("[ALGO] type assertion failed")
			return false
		}
		nd.setNpuFullPath(superPodName, superPodPath, aiPingStrategy)
	}

	setNpu2Npu(npu2Npu, aiPingStrategy)
	return true
}

// 将npu出框的路径里的数据填充到ai_ping_strategy里的chain_list
func (nd *NetDetect) setNpuFullPath(planeName string, fullPaths []string, aiPingStrategy *AiPingStrategy) {
	netplaneLen := len(fullPaths)
	if netplaneLen == 0 {
		return
	}

	for _, eachPath := range fullPaths {
		// 当前需要满足 L1:0#L0:0#Slot:0#Ip:0 这样的格式
		requiredCharCount1 := 3 // '#' 字符个数
		requiredCharCount2 := 4 // ':' 字符个数
		if strings.Count(eachPath, layerIntervalChar) != requiredCharCount1 ||
			strings.Count(eachPath, portIntervalChar) != requiredCharCount2 {
			continue
		}

		eachPathBefore := strings.TrimSuffix(eachPath, ":0")
		pathArr := strings.Split(eachPathBefore, "#")
		eachPathAfter := setLayerPort(pathArr)
		eachPathAfter = cropNaStr(eachPathAfter)
		if eachPathAfter != "" {
			aiPingStrategy.chainList[planeName] = append(aiPingStrategy.chainList[planeName], eachPathAfter)
			nd.curTopo = append(nd.curTopo, eachPathAfter)
		}
	}
}

// 删除字符串中所有的 ".NA" 和 "NA."等无意义的对象层级
func cropNaStr(s string) string {
	s = strings.ReplaceAll(s, ".NA", "")
	s = strings.ReplaceAll(s, "NA.", "")
	return s
}

// 设置每个层级的网口信息
func setLayerPort(infoStr []string) string {
	infoLayerNum := len(infoStr)
	finalArr := make([]string, 0)
	for i := 0; i < infoLayerNum-1; i++ {
		if strings.Contains(infoStr[i+1], nSlotConstant) {
			npuLayerArr := strings.Split(infoStr[infoLayerNum-1], dotIntervalChar)
			part := strings.Split(npuLayerArr[0], normalIntervalChar)
			if len(part) != baseSegmentNum {
				return ""
			}
			curLayerStr := strings.ReplaceAll(infoStr[i], ":0", ":"+part[1])
			finalArr = append(finalArr, curLayerStr)
			finalArr = append(finalArr, infoStr[i+1])
			finalArr = append(finalArr, strings.Join(npuLayerArr[1:], ".")) // 兼容ip和非ip的组装
			break
		}
		if strings.Contains(infoStr[i], roceSwitchConstant) {
			finalArr = append(finalArr, infoStr[i])
			continue
		}
		curLayerStr := strings.ReplaceAll(infoStr[i], ":0", "")
		childLayerStr := strings.ReplaceAll(infoStr[i+1], ":0", "")
		childLayerInfoArr := strings.Split(childLayerStr, dotIntervalChar)
		// 获取有实际意义的目标层级字符串
		var targetStr string
		if childLayerInfoArr[0] != "NA" {
			targetStr = childLayerInfoArr[0]
		} else {
			if len(childLayerInfoArr) < baseSegmentNum {
				hwlog.RunLog.Errorf("[NETFAULT ALGO]the length of childLayerInfoArr is less than: %d", baseSegmentNum)
				return ""
			}
			targetStr = childLayerInfoArr[1]
		}
		part := strings.Split(targetStr, normalIntervalChar)
		if len(part) != baseSegmentNum {
			return ""
		}
		curLayerStr = curLayerStr + portIntervalChar + part[1]
		finalArr = append(finalArr, curLayerStr)
	}
	result := strings.Join(finalArr, "#")
	return result
}

// 将npu_npu里的数据填充到ai_ping_strategy里的npu_npu_list
func setNpu2Npu(npu2Npu []string, aiPingStrategy *AiPingStrategy) {
	npuLen := len(npu2Npu)
	for i := 0; i < npuLen; i++ {
		aiPingStrategy.npuNpuList = append(aiPingStrategy.npuNpuList, npu2Npu[i])
	}
}

// 将chain_list结构转化为dataFrame结构
func setDfChainMap(chainLists map[string][]string, dfChainsMap map[string]*DataFrame) bool {
	if dfChainsMap == nil {
		return false
	}

	for planeName, chainList := range chainLists {
		if len(chainList) == 0 {
			continue
		}
		tmpRowStr := chainList[0]
		modifyRowStr := strings.ReplaceAll(tmpRowStr, layerIntervalChar, portIntervalChar)
		tmpRowStrArr := strings.Split(modifyRowStr, portIntervalChar)
		colNum := len(tmpRowStrArr)
		if colNum < minimumColNum || colNum%baseEvenNum == 0 {
			hwlog.RunLog.Errorf("[ALGO] unexpected topo, colNum: %d, minimumColNum: %d, baseEvenNum: %d",
				colNum, minimumColNum, baseEvenNum)
			return false
		}

		dfChains := initDfChain()
		setDfChainColumn(colNum, dfChains)
		setDfChainRow(chainList, dfChains)
		dfChainsMap[planeName] = dfChains
	}

	return true
}

// 设置dfChains的列
func setDfChainColumn(colNum int, dfChains *DataFrame) {
	for i := 0; i < colNum; i++ {
		layerNum := (colNum - i) / baseEvenNum
		if i == colNum-1 {
			dfChains.columnNames = append(dfChains.columnNames, ipConstant)
			dfChains.chains[ipConstant] = []string{}
		} else if i%baseEvenNum == 0 {
			layerStr := fmt.Sprintf("%s%s%d", layerConstant, objectIntervalChar, layerNum)
			dfChains.columnNames = append(dfChains.columnNames, layerStr)
			dfChains.chains[layerStr] = []string{}
		} else {
			portStr := fmt.Sprintf("%s%s%d", portConstant, objectIntervalChar, layerNum)
			dfChains.columnNames = append(dfChains.columnNames, portStr)
			dfChains.chains[portStr] = []string{}
		}
	}
}

// 设置dfChains的行
func setDfChainRow(chainList []string, dfChains *DataFrame) {
	dfChains.rowNum = len(chainList)
	for i := 0; i < dfChains.rowNum; i++ {
		tmpRowStr := chainList[i]
		modifyRowStr := strings.ReplaceAll(tmpRowStr, layerIntervalChar, portIntervalChar)
		tmpRowStrArr := strings.Split(modifyRowStr, portIntervalChar)
		for j := 0; j < len(tmpRowStrArr); j++ {
			key := dfChains.columnNames[j]
			if chain, ok := dfChains.chains[key].([]string); ok {
				chain = append(chain, tmpRowStrArr[j])
				dfChains.chains[key] = chain
			}
		}
	}
}

// 获得拨测策略
func (nd *NetDetect) setPingDict(aiPingStrategy *AiPingStrategy, dfChainsMap map[string]*DataFrame) bool {
	if len(dfChainsMap) == 0 {
		hwlog.RunLog.Error("[ALGO] dfChainsMap is null")
		return false
	}

	for _, dfChains := range dfChainsMap {
		if dfChains == nil {
			continue
		}
		layers := extractLayers(dfChains)
		nd.doPingStrategy(layers, aiPingStrategy, dfChains)
	}

	return true
}

// 提取列名
func extractLayers(dfChains *DataFrame) []string {
	var layers []string

	for _, str := range dfChains.columnNames {
		if strings.Contains(str, layerConstant) {
			layers = append(layers, str)
		}
	}

	return layers
}

// 拨测策略核心步骤
func (nd *NetDetect) doPingStrategy(layers []string, aiPingStrategy *AiPingStrategy, dfChains *DataFrame) {
	for i := 0; i < len(layers); i++ {
		layer := layers[i]
		if values, ok := dfChains.chains[layer].([]string); ok {
			// 分组
			uniqueValues := uniqueSlice(values)
			aiPingStrategy.layersIps[layer] = uniqueValues
			groupBy(dfChains, layer, aiPingStrategy.dfGrouped)

			// 分层拨测
			if i == len(layers)-1 && !nd.curFullPingFlag {
				npuFullPing(aiPingStrategy)
				nd.curFullPingFlag = true
			} else if i < len(layers)-1 {
				childLayer := layers[i+1]
				nd.npuRingPing(aiPingStrategy, layer, childLayer)
			}
		}
	}
}

// 对dataFrame进行分组操作（根据layer列分组），分组的结果存到dfGrouped里
func groupBy(df *DataFrame, layer string, dfGrouped *DataFrameGroupBy) {
	// 清空分组数据
	dfGrouped.groupNums = 0
	dfGrouped.groups = make([]*Group, 0)

	// 获取列的数据
	var columnData []string
	if chains, ok := df.chains[layer].([]string); ok {
		columnData = chains
	} else {
		columnData = []string{}
	}

	// 遍历列数据并分组
	for i := 0; i < len(columnData); i++ {
		value := columnData[i]

		// 检查这个值是否已经存在于分组中
		var groupIdx = -1
		for j := 0; j < dfGrouped.groupNums; j++ {
			if dfGrouped.groups[j].key == value {
				groupIdx = j
				break
			}
		}

		// 如果不存在该分组，创建新的分组
		if groupIdx == -1 {
			groupIdx = dfGrouped.groupNums
			dfGrouped.groupNums++
			tmpGroup := new(Group)
			tmpGroup.key = value
			tmpGroup.groupData = new(DataFrame)
			tmpGroup.groupData.columnNames = df.columnNames
			tmpGroup.groupData.chains = make(map[string]any)
			tmpGroup.groupData.rowNum = 0
			dfGrouped.groups = append(dfGrouped.groups, tmpGroup)
		}

		// 将当前行的数据添加到对应的分组中
		group := dfGrouped.groups[groupIdx]
		groupDf := group.groupData
		groupDf.rowNum++

		// 将行数据按分组的列添加到分组中的DataFrame
		for j := 0; j < len(groupDf.columnNames); j++ {
			key := groupDf.columnNames[j]
			var tempList []string
			if chains, ok := df.chains[key].([]string); ok {
				tempList = chains
			} else {
				tempList = []string{}
			}
			column, exists := groupDf.chains[key].([]string)
			if !exists {
				column = make([]string, 0, groupDf.rowNum)
			}
			column = append(column, tempList[i])
			groupDf.chains[key] = column
		}
	}
}

// 获取dataFrame里其中某个组
func getGroup(dfGrouped *DataFrameGroupBy, key string) *DataFrame {
	if dfGrouped == nil {
		return nil
	}

	for i := 0; i < dfGrouped.groupNums; i++ {
		group := dfGrouped.groups[i]
		if group.key == key {
			return group.groupData
		}
	}

	return nil
}

// 板号/框号按编号从小到大来排序
func sortLayerList(childLayerList []string) {
	sort.Slice(childLayerList, func(i, j int) bool {
		// 提取板号/框号字符串最后一个"-"后面的数字
		lastI := strings.LastIndex(childLayerList[i], "-")
		lastJ := strings.LastIndex(childLayerList[j], "-")
		if lastI == -1 || lastJ == -1 {
			hwlog.RunLog.Error("[ALGO] wrong format of board/rack")
			return false
		}

		numI, err := strconv.Atoi(childLayerList[i][lastI+1:])
		numJ, err := strconv.Atoi(childLayerList[j][lastJ+1:])
		if err != nil {
			hwlog.RunLog.Errorf("[ALGO] wrong format of board/rack, err: %v", err)
			return false
		}

		return numI < numJ
	})
}

// 板号/框号按编号从小到大来排序
func sortLayerList2(childLayerList []string) {
	sort.Slice(childLayerList, func(i, j int) bool {
		// 用 "." 分割字符串以获取第一个部分
		partsI := strings.Split(childLayerList[i], ".")
		partsJ := strings.Split(childLayerList[j], ".")

		// 如果第一个部分相同，使用第二个部分的数字排序
		if partsI[0] == partsJ[0] {
			if len(partsI) != baseSegmentNum || len(partsJ) != baseSegmentNum {
				return false
			}
			return sortBySuffix(partsI[1], partsJ[1])
		}
		// 否则使用第一个部分的数字排序
		return sortBySuffix(partsI[0], partsJ[0])
	})
}

// 根据 "-" 后面的数字进行排序
func sortBySuffix(s1, s2 string) bool {
	lastI := strings.LastIndex(s1, "-")
	lastJ := strings.LastIndex(s2, "-")

	if lastI == -1 || lastJ == -1 {
		hwlog.RunLog.Error("[ALGO] wrong format of board/rack")
		return false
	}

	numI, errI := strconv.Atoi(s1[lastI+1:])
	numJ, errJ := strconv.Atoi(s2[lastJ+1:])

	if errI != nil || errJ != nil {
		hwlog.RunLog.Errorf("[ALGO] wrong format of board/rack, err: %v, %v", errI, errJ)
		return false
	}

	return numI < numJ
}

// sdid按编号从小到大排序
func sortSdidList(numStrings []string) {
	sort.Slice(numStrings, func(i, j int) bool {
		// 将字符串转换为整数进行比较
		numI, err := strconv.Atoi(numStrings[i])
		numJ, err := strconv.Atoi(numStrings[j])
		if err != nil {
			hwlog.RunLog.Errorf("[ALGO] wrong format of sdid, err: %v", err)
			return false
		}

		return numI < numJ
	})
}

// sortIpList sort the ip list by npu id
func (nd *NetDetect) sortIpList(ipList []string) {
	// create a ip and number struct for order
	type npuWithNumber struct {
		ip     string
		number int
	}

	// prepare the initial data
	toSort := make([]npuWithNumber, 0, len(ipList))
	for _, ip := range ipList {
		info, ok := nd.curNpuInfo[ip]
		if !ok || isEmptyNpuInfo(info) {
			// if no data or data is empyt, the default number is the max value
			// so that this value will be the last order
			toSort = append(toSort, npuWithNumber{ip: ip, number: math.MaxInt32})
			continue
		}
		toSort = append(toSort, npuWithNumber{ip: ip, number: info.NpuNumber})
	}

	sort.Slice(toSort, func(i, j int) bool {
		return toSort[i].number < toSort[j].number
	})
	if len(toSort) != len(ipList) {
		return
	}
	// write back the data to ipList
	for i, item := range toSort {
		ipList[i] = item.ip
	}
}

// 判断NpuInfo结构体是否为空
func isEmptyNpuInfo(info NpuInfo) bool {
	if info.IP == "" && info.NetPlaneId == "" && info.SlotName == "" && info.RackName == "" && info.NpuNumber == 0 {
		return true
	}
	return false
}

// 过滤指定行
func filterAndExtractIps(dfIp *DataFrame, childLayerName string, childLayer string) []string {
	ipList := make([]string, 0)

	// 获取指定列的数据（child_layer_name）
	var childLayerCol []string
	if col, ok := dfIp.chains[childLayerName].([]string); ok {
		childLayerCol = col
	} else {
		childLayerCol = []string{}
	}

	// 获取 ip 列的数据
	var ipCol []string
	if col, ok := dfIp.chains[ipConstant].([]string); ok {
		ipCol = col
	} else {
		ipCol = []string{}
	}

	if len(childLayerCol) != dfIp.rowNum || len(ipCol) != dfIp.rowNum {
		return ipList
	}

	// 遍历child_layer列，筛选出符合条件的行
	for i := 0; i < dfIp.rowNum; i++ {
		if childLayerCol[i] == childLayer {
			// 如果该行的child_layer列值符合条件，将对应行的ip列值添加到ip_list中
			ipList = append(ipList, ipCol[i])
		}
	}

	return ipList
}

// 生成排列组合
func generatePermutations(ipList []string, aiPingStrategy *AiPingStrategy, layer string, layerIp string) {
	if len(ipList) == 0 || aiPingStrategy == nil || layer == "" || layerIp == "" {
		// log
		return
	}

	// 检查 ip_ls 是否为空
	minLen := 2
	if len(ipList) < minLen {
		return
	}

	// 创建 IP 对并将其存储在 ping_dict 中
	for i := 0; i < len(ipList); i++ {
		for j := 0; j < len(ipList); j++ {
			if i != j {
				// 组合 x 和 y 为 {'from': x, 'to': y}
				fromIp := ipList[i]
				toIp := ipList[j]

				// 使用 ping_dict_key 将结果添加到 ping_dict
				pingDictKey := fmt.Sprintf("%s%s%s", layer, portIntervalChar, layerIp)

				// 确保动态创建的结构体在内存中进行释放
				addPingPair(aiPingStrategy, fromIp, toIp, pingDictKey)
			}
		}
	}
}

// 用于将IP地址对添加到cJSON对象中的辅助函数
func addPingPair(aiPingStrategy *AiPingStrategy, fromIp string, toIp string, pingDictKey string) {
	if aiPingStrategy == nil || fromIp == "" || toIp == "" || pingDictKey == "" {
		// log
		return
	}

	// 创建一个对象 {'from': from_ip, 'to': to_ip}
	pingPair := make(map[string]any)
	pingPair[fromConstant] = fromIp
	pingPair[toConstant] = toIp

	// 获取当前层的 key
	value, exists := aiPingStrategy.pingDict[pingDictKey]
	var layerKey []any

	if !exists {
		layerKey = make([]any, 0)
	} else {
		var ok bool
		layerKey, ok = value.([]any)
		if !ok {
			layerKey = make([]any, 0)
		}
	}

	layerKey = append(layerKey, pingPair)
	aiPingStrategy.pingDict[pingDictKey] = layerKey
}

// 从ip_ls中随机选择SWITCH_POINTS个元素, 如果ip_ls的长度小于等于SWITCH_POINTS，则返回全部IP
func sampleIPs(ipList []string, sampleSize int) []string {
	if len(ipList) <= sampleSize {
		// 如果 ipList 中的元素个数小于等于需要的个数，直接返回所有 IP
		return append([]string(nil), ipList...) // 返回一个新的切片
	}
	// 创建一个切片用于存储结果
	return randomSample(ipList, sampleSize)
}

// 随机选择n个不重复的IP地址
func randomSample(ipList []string, n int) []string {
	// 创建 ipList 的副本
	copyList := make([]string, len(ipList))
	copy(copyList, ipList)

	// 初始化随机数生成器
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Fisher-Yates 洗牌算法
	for i := len(copyList) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		copyList[i], copyList[j] = copyList[j], copyList[i]
	}

	// 返回前 n 个元素
	if n > len(copyList) {
		n = len(copyList)
	}
	return copyList[:n]
}

// 根据给定的srcIp, 获得相同编号的npu_ip
func (nd *NetDetect) getSameNumDstIp(srcIp string, dstIps []string) string {
	var srcNpuInfo NpuInfo
	if npuInfo, ok := nd.curNpuInfo[srcIp]; ok {
		srcNpuInfo = npuInfo
	} else {
		hwlog.RunLog.Errorf("[ALGO] can't find srcIp: %s in NpuInfo, superPodId: %s", srcIp, nd.curSuperPodId)
		return ""
	}

	for i := 0; i < len(dstIps); i++ {
		var dstNpuInfo NpuInfo
		if npuInfo, ok := nd.curNpuInfo[dstIps[i]]; ok {
			dstNpuInfo = npuInfo
		} else {
			hwlog.RunLog.Errorf("[ALGO] can't find dstIp: %s in NpuInfo, superPodId: %s",
				dstIps[i], nd.curSuperPodId)
			continue
		}

		if srcNpuInfo.NpuNumber != dstNpuInfo.NpuNumber {
			continue
		} else {
			return dstIps[i]
		}
	}

	return ""
}

// 获得同轴编号的npu_ip
func (nd *NetDetect) getAlignNumDstIp(srcIp string, dstIps []string) string {
	var srcNpuInfo NpuInfo
	if npuInfo, ok := nd.curNpuInfo[srcIp]; ok {
		srcNpuInfo = npuInfo
	} else {
		hwlog.RunLog.Errorf("[ALGO] can't find srcIp: %s in NpuInfo, superPodId: %s", srcIp, nd.curSuperPodId)
		return ""
	}

	for i := 0; i < len(dstIps); i++ {
		var dstNpuInfo NpuInfo
		if npuInfo, ok := nd.curNpuInfo[dstIps[i]]; ok {
			dstNpuInfo = npuInfo
		} else {
			hwlog.RunLog.Errorf("[ALGO] can't find dstIp: %s in NpuInfo, superPodId: %s",
				dstIps[i], nd.curSuperPodId)
			continue
		}

		if (dstNpuInfo.NpuNumber-srcNpuInfo.NpuNumber)%baseNpuNum != 0 {
			continue
		} else {
			return dstIps[i]
		}
	}

	return ""
}

// 获得跨轴编号的npu_ip
func (nd *NetDetect) getCrossNumDstIp(srcIp string, dstIps []string) string {
	var srcNpuInfo NpuInfo
	if npuInfo, ok := nd.curNpuInfo[srcIp]; ok {
		srcNpuInfo = npuInfo
	} else {
		hwlog.RunLog.Errorf("[ALGO] can't find srcIp: %s in NpuInfo, superPodId: %s", srcIp, nd.curSuperPodId)
		return ""
	}

	for i := 0; i < len(dstIps); i++ {
		var dstNpuInfo NpuInfo
		if npuInfo, ok := nd.curNpuInfo[dstIps[i]]; ok {
			dstNpuInfo = npuInfo
		} else {
			hwlog.RunLog.Errorf("[ALGO] can't find dstIp: %s in NpuInfo, superPodId: %s",
				dstIps[i], nd.curSuperPodId)
			continue
		}

		diffValueForAdjacentNpu := 1     // the difference in npu between adjacent plates across axes
		diffValueForNonadjacentNpu := -7 // the difference in npu between nonadjacent plates across axes
		if (dstNpuInfo.NpuNumber-srcNpuInfo.NpuNumber)%baseNpuNum == diffValueForAdjacentNpu ||
			(dstNpuInfo.NpuNumber-srcNpuInfo.NpuNumber)%baseNpuNum == diffValueForNonadjacentNpu {
			return dstIps[i]
		} else {
			continue
		}
	}

	return ""
}

// 处理ipPairs对
func processIpPairs(aiPingStrategy *AiPingStrategy, pingDictKey string, ipPairs [][]string) {
	// 如果 pingDictKey 不存在，则初始化一个空的 []map[string]any
	if _, exists := aiPingStrategy.pingDict[pingDictKey]; !exists {
		aiPingStrategy.pingDict[pingDictKey] = make([]map[string]any, 0)
	}

	// 获取 pingDictKey 对应的值
	value, ok := aiPingStrategy.pingDict[pingDictKey].([]map[string]any)
	if !ok {
		value = make([]map[string]any, 0)
	}

	// 遍历 ipPairs，填充 value
	for i := 0; i < len(ipPairs); i++ {
		if len(ipPairs[i]) < fromToNum {
			continue
		}
		tempMap := make(map[string]any)
		tempMap[fromConstant] = ipPairs[i][0]
		tempMap[toConstant] = ipPairs[i][1]
		value = append(value, tempMap)
	}

	// 将更新后的 value 同步回 aiPingStrategy.pingDict
	aiPingStrategy.pingDict[pingDictKey] = value
}

// 框间配对
func (nd *NetDetect) processIpsRack(aiPingStrategy *AiPingStrategy, srcIps []string, dstIps []string,
	pingDictKey string) {
	ipPairs := make([][]string, 0)
	for i := 0; i < len(srcIps); i++ {
		dstIp := nd.getSameNumDstIp(srcIps[i], dstIps)
		if dstIp != "" {
			// ipPair存储每一对ip，ipPairs存储所有的配对
			ipPair := make([]string, 0)
			ipPair = append(ipPair, srcIps[i])
			ipPair = append(ipPair, dstIp)
			ipPairs = append(ipPairs, ipPair)
		}
	}

	processIpPairs(aiPingStrategy, pingDictKey, ipPairs)
}

// 板间配对
func (nd *NetDetect) processIpsSlot(aiPingStrategy *AiPingStrategy, srcIps []string, dstIps []string,
	pingDictKey string) {
	ipPairs := make([][]string, 0)
	for i := 0; i < len(srcIps); i++ {
		alignDstIp := nd.getAlignNumDstIp(srcIps[i], dstIps)
		if alignDstIp != "" {
			// ipPair存储每一对ip，ipPairs存储所有的配对
			ipPair := make([]string, 0)
			ipPair = append(ipPair, srcIps[i])
			ipPair = append(ipPair, alignDstIp)
			ipPairs = append(ipPairs, ipPair)
		}

		crossDstIp := nd.getCrossNumDstIp(srcIps[i], dstIps)
		if crossDstIp != "" {
			// ipPair存储每一对ip，ipPairs存储所有的配对
			ipPair := make([]string, 0)
			ipPair = append(ipPair, srcIps[i])
			ipPair = append(ipPair, crossDstIp)
			ipPairs = append(ipPairs, ipPair)
		}
	}

	processIpPairs(aiPingStrategy, pingDictKey, ipPairs)
}

// 批量处理ip配对(框间一一对应 + 板间同轴 + 板间跨轴)
func (nd *NetDetect) processBothAxis(aiPingStrategy *AiPingStrategy, srcIps []string, dstIps []string,
	pingDictKey string) {
	if strings.Contains(pingDictKey, layer3Constant) {
		nd.processIpsRack(aiPingStrategy, srcIps, dstIps, pingDictKey)
	}

	if strings.Contains(pingDictKey, layer2Constant) {
		nd.processIpsSlot(aiPingStrategy, srcIps, dstIps, pingDictKey)
	}
}

// 批量处理ip配对(框间一一对应 + 板间跨轴)
func processCrossAxis(aiPingStrategy *AiPingStrategy, srcIps []string, dstIps []string, pingDictKey string) {
	// 确保长度一致
	if len(srcIps) != len(dstIps) {
		hwlog.RunLog.Warn("[ALGO] length between srcIps and dstIps is not equal")
		return
	}

	srcIpsLen := len(srcIps)
	ipPairs := make([][]string, srcIpsLen)
	for i := 0; i < srcIpsLen; i++ {
		ipPairs[i] = make([]string, 2)
		if strings.Contains(pingDictKey, layer3Constant) {
			ipPairs[i][0] = srcIps[i]
			ipPairs[i][1] = dstIps[i]
		} else {
			ipPairs[i][0] = srcIps[i]
			ipPairs[i][1] = dstIps[(i+1)%len(dstIps)]
		}
	}

	processIpPairs(aiPingStrategy, pingDictKey, ipPairs)
}

// 批量处理ip配对(同轴)
func processSameAxis(aiPingStrategy *AiPingStrategy, srcIps []string, dstIps []string, pingDictKey string) {
	// 确保长度一致
	if len(srcIps) != len(dstIps) {
		hwlog.RunLog.Warn("[ALGO] length between srcIps and dstIps is not equal")
		return
	}

	srcIpsLen := len(srcIps)
	ipPairs := make([][]string, srcIpsLen)
	for i := 0; i < srcIpsLen; i++ {
		ipPairs[i] = make([]string, 2)
		ipPairs[i][0] = srcIps[i]
		ipPairs[i][1] = dstIps[i]
	}

	processIpPairs(aiPingStrategy, pingDictKey, ipPairs)
}

// sdidList的环ping
func processSdidList(aiPingStrategy *AiPingStrategy, srcSdidList []string, dstSdidList []string, pingDictKey string) {
	// 确保长度一致
	if len(srcSdidList) != len(dstSdidList) {
		hwlog.RunLog.Error("[ALGO] length between srcSdidList and dstSdidList is not equal")
		return
	}

	srcSdidListLen := len(srcSdidList)
	ipPairs := make([][]string, srcSdidListLen)
	for i := 0; i < srcSdidListLen; i++ {
		ipPairs[i] = make([]string, 2)
		ipPairs[i][0] = srcSdidList[i]
		ipPairs[i][1] = dstSdidList[i]
	}

	processIpPairs(aiPingStrategy, pingDictKey, ipPairs)
}

// npu的full_mesh策略
func npuFullPing(aiPingStrategy *AiPingStrategy) {
	if aiPingStrategy == nil || len(aiPingStrategy.npuNpuList) == 0 {
		return
	}

	// 传了npu的直连关系，直接添加到ping_dict中
	length := len(aiPingStrategy.npuNpuList)
	for i := 0; i < length; i++ {
		npuPairStr := aiPingStrategy.npuNpuList[i]
		npuList := strings.Split(npuPairStr, layerIntervalChar)
		if len(npuList) != baseSegmentNum {
			continue
		}

		npuList[0] = strings.ReplaceAll(npuList[0], ":0", "")
		npuList[1] = strings.ReplaceAll(npuList[1], ":0", "")

		// 正反向都需要
		addPingPair(aiPingStrategy, npuList[0], npuList[1], argsNpu2Npu)
		addPingPair(aiPingStrategy, npuList[1], npuList[0], argsNpu2Npu)
	}
}

// npu的环ping策略
func (nd *NetDetect) npuRingPing(aiPingStrategy *AiPingStrategy, layer string, childLayerName string) {
	switch nd.curNpuType {
	case a3NpuTypeConstant:
		nd.a3NpuRingPing(aiPingStrategy, layer, childLayerName)
	case a5NpuTypeConstant:
		nd.a5NpuRingPing(aiPingStrategy, layer, childLayerName)
	default:
		hwlog.RunLog.Error("[ALGO] unexpected detection type!")
	}
}

// A5 scenario: NPU loop ping strategy
func (nd *NetDetect) a5NpuRingPing(aiPingStrategy *AiPingStrategy, layer string, childLayerName string) {
	layerIps := getCurLayerIps(aiPingStrategy, layer)
	// create a temporary map to store randomly selected IPs
	randomIps := make(map[string]interface{})
	for i := 0; i < len(layerIps); i++ {
		layerIp := layerIps[i]
		// get the DataFrame corresponding to the layer_ip
		dfIp := getGroup(aiPingStrategy.dfGrouped, layerIp)
		// get the unique values of the child_layer_name column
		childLayerList := getChildColUniqueList(childLayerName, dfIp)
		// skip loop ping for single slot or single board
		if len(childLayerList) <= 1 {
			continue
		}
		// sort each board/slot by number
		sortLayerList2(childLayerList)
		// select random IPs for each sub-layer and store in random_ips
		nd.setRandomIps(childLayerList, childLayerName, dfIp, randomIps)
		// generate and process IP pairs
		nd.setPingPair(layer, layerIp, childLayerList, randomIps, aiPingStrategy)
	}
}

// A3场景npu的环ping策略
func (nd *NetDetect) a3NpuRingPing(aiPingStrategy *AiPingStrategy, layer string, childLayerName string) {
	layerIps := getCurLayerIps(aiPingStrategy, layer)

	// 创建一个临时的map用于存储随机选中的IP
	randomIps := make(map[string]any)
	for i := 0; i < len(layerIps); i++ {
		layerIp := layerIps[i]

		// 获取与该layer_ip对应的DataFrame
		dfIp := getGroup(aiPingStrategy.dfGrouped, layerIp)

		// 获取child_layer_name列的唯一值
		childLayerList := getChildColUniqueList(childLayerName, dfIp)

		if len(childLayerList) > 1 {
			// 为每个节点按编号排序
			sortLayerList(childLayerList)

			// 为每个子层选择随机IP并存储在random_ips中
			nd.setRandomIps(childLayerList, childLayerName, dfIp, randomIps)

			// 生成并处理IP对
			nd.setPingPair(layer, layerIp, childLayerList, randomIps, aiPingStrategy)
		} else if layer == layer3Constant {
			// 单一节点，不进行节点间环ping
			continue
		} else if layer == layer2Constant {
			// 节点内sdid进行环ping
			nd.a3NodeRingPing(layer, layerIp, childLayerList, aiPingStrategy)
		}
	}
}

// A3场景下的节点内环ping
func (nd *NetDetect) a3NodeRingPing(layer string, layerIp string, childLayerList []string,
	aiPingStrategy *AiPingStrategy) {
	if len(childLayerList) == 0 {
		return
	}

	// 获取sdid列表并排序
	srcSdidList := nd.findIPsBySlotName(childLayerList[0])
	sortSdidList(srcSdidList)

	// 获取环ping list
	dstSdidList := moveSliceLeftTwoStep(srcSdidList)

	// 生成ping_dict_key
	pingDictKey := fmt.Sprintf("%s%s%s", layer, portIntervalChar, layerIp)

	// 直接加入pingDict
	processSdidList(aiPingStrategy, srcSdidList, dstSdidList, pingDictKey)
}

// 查找所有 SlotName 等于 targetSlotName 的 IP
func (nd *NetDetect) findIPsBySlotName(targetSlotName string) []string {
	var result []string

	for _, slotInfo := range nd.curNpuInfo {
		if slotInfo.SlotName == targetSlotName {
			result = append(result, slotInfo.IP)
		}
	}

	return result
}

// 获取当前layer下的IP列表
func getCurLayerIps(aiPingStrategy *AiPingStrategy, layer string) []string {
	var layerIps []string

	if ips, ok := aiPingStrategy.layersIps[layer].([]string); ok {
		layerIps = ips
	} else {
		layerIps = []string{}
	}

	return layerIps
}

// 获取child_layer_name列的唯一值列表
func getChildColUniqueList(childLayerName string, dfIp *DataFrame) []string {
	if dfIp == nil {
		return []string{}
	}

	var values []string

	if chains, ok := dfIp.chains[childLayerName].([]string); ok {
		values = chains
	} else {
		values = []string{}
	}

	return uniqueSlice(values)
}

// 为每个子层选择随机IP并存储在random_ips中
func (nd *NetDetect) setRandomIps(childLayerList []string, childLayerName string, dfIp *DataFrame,
	randomIps map[string]any) {
	if randomIps == nil {
		return
	}

	for j := 0; j < len(childLayerList); j++ {
		childLayer := childLayerList[j]

		// 获取所有与该子层对应的ip_ls
		ipList := filterAndExtractIps(dfIp, childLayerName, childLayer)

		// 随机选择IP
		sampledIps := sampleIPs(ipList, sampleNum)

		// 按npu编号排序
		nd.sortIpList(sampledIps)

		// 将sampled_ips添加到random_ips中
		randomIps[childLayer] = sampledIps
	}
}

// 生成并处理IP对
func (nd *NetDetect) setPingPair(layer string, layerIp string, childLayerList []string,
	randomIps map[string]any, aiPingStrategy *AiPingStrategy) {
	if randomIps == nil {
		return
	}

	for j := 0; j < len(childLayerList); j++ {
		key := childLayerList[j]
		srcIps, ok := randomIps[key].([]string)
		if !ok {
			continue
		}
		dstIps := make([]string, 0)
		if j == len(childLayerList)-1 {
			dstIps, ok = randomIps[childLayerList[0]].([]string)
			if !ok {
				continue
			}
		} else {
			dstIps, ok = randomIps[childLayerList[j+1]].([]string)
			if !ok {
				continue
			}
		}

		// 生成ping_dict_key
		pingDictKey := fmt.Sprintf("%s%s%s", layer, portIntervalChar, layerIp)

		// 根据选轴策略批量处理IP配对
		switch nd.curAxisStrategy {
		case bothAxisConstant:
			nd.processBothAxis(aiPingStrategy, srcIps, dstIps, pingDictKey)
		case crossAxisConstant:
			processCrossAxis(aiPingStrategy, srcIps, dstIps, pingDictKey)
		case sameAxisConstant:
			processSameAxis(aiPingStrategy, srcIps, dstIps, pingDictKey)
		default:
			break
		}
	}
}

// 处理拨测策略算法的输出
func (nd *NetDetect) processOutput(output map[string]any, aiPingStrategy *AiPingStrategy) bool {
	if output == nil || aiPingStrategy == nil {
		hwlog.RunLog.Error("[ALGO] output or aiPingStrategy is null")
		return false
	}

	pingList := make([]any, 0)
	for _, layerValue := range aiPingStrategy.pingDict {
		switch v := layerValue.(type) {
		case []map[string]any:
			nd.processMapType(&pingList, v)
		case []any:
			nd.processSliceType(&pingList, v)
		default:
			hwlog.RunLog.Error("[ALGO] type assertion failed")
			return false
		}
	}

	output[pingListConstant] = pingList
	return true
}

// 处理map类型
func (nd *NetDetect) processMapType(pingList *[]any, v []map[string]any) {
	for _, item := range v {
		srcPingObj, srcOk := item[fromConstant].(string)
		dstPingObj, dstOk := item[toConstant].(string)
		if !srcOk || !dstOk {
			hwlog.RunLog.Error("[ALGO] type assertion failed")
			continue
		}
		srcPhyId := nd.findNpuNumberByPingObj(srcPingObj)
		dstPhyId := nd.findNpuNumberByPingObj(dstPingObj)
		item[srcTypeConstant] = nd.curPingObjType
		item[dstTypeConstant] = nd.curPingObjType
		item[srcCardPhyId] = srcPhyId
		item[dstCardPhyId] = dstPhyId
		item[pktSizeConstant] = pktSizeNum
		item[srcAddrConstant] = item[fromConstant]
		item[dstAddrConstant] = item[toConstant]
		delete(item, fromConstant)
		delete(item, toConstant)
		*pingList = append(*pingList, item)
	}
}

// 处理切片类型
func (nd *NetDetect) processSliceType(pingList *[]any, v []any) {
	for _, item := range v {
		if m, ok := item.(map[string]any); ok {
			srcPingObj, srcOk := m[fromConstant].(string)
			dstPingObj, dstOk := m[toConstant].(string)
			if !srcOk || !dstOk {
				hwlog.RunLog.Error("[ALGO] type assertion failed")
				continue
			}
			srcPhyId := nd.findNpuNumberByPingObj(srcPingObj)
			dstPhyId := nd.findNpuNumberByPingObj(dstPingObj)
			m[srcTypeConstant] = nd.curPingObjType
			m[dstTypeConstant] = nd.curPingObjType
			m[srcCardPhyId] = srcPhyId
			m[dstCardPhyId] = dstPhyId
			m[pktSizeConstant] = pktSizeNum
			m[srcAddrConstant] = m[fromConstant]
			m[dstAddrConstant] = m[toConstant]
			delete(m, fromConstant)
			delete(m, toConstant)
			*pingList = append(*pingList, m)
		} else {
			continue
		}
	}
}

// 根据ping探测目标(ip、eid、die)获取npu编号
func (nd *NetDetect) findNpuNumberByPingObj(obj string) int {
	if len(nd.curNpuInfo) == 0 {
		hwlog.RunLog.Infof("[ALGO] no npuInfoMap found in superPodId: %s", nd.curSuperPodId)
		return -1
	}

	if npuInfo, exist := nd.curNpuInfo[obj]; exist {
		return npuInfo.NpuNumber
	}

	return -1
}
