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
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/fileutils"
)

/* 比较当前任务在各个节点上的劣化等级，取出最大 */
func getCurDetectionJobMaxDegradationLevel(mergedData *config.ClusterJobResult, curDegradation string) {
	curPercentStr := strings.TrimSuffix(curDegradation, "%")
	curNumberPercent, err := strconv.ParseFloat(curPercentStr, decimalLen)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]%s:%v", curDegradation, err)
		return
	}
	curDegradationLevel := mergedData.DegradationLevel
	curDegradationStr := strings.TrimSuffix(curDegradationLevel, "%")
	curDegradationNumber, err := strconv.ParseFloat(curDegradationStr, decimalLen)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]%v", err)
		return
	}
	if curDegradationNumber < curNumberPercent {
		mergedData.DegradationLevel = curDegradation
	}
}

/* 从集群路径下汇聚当前任务所有节点侧检测结果 */
func getGatherData(mergedData *config.ClusterJobResult,
	nodeResult config.NodeDetectionResult) {
	/* 每个节点的结果文件中只会存在一个键值对 */
	if len(nodeResult) != 1 {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]Error node level detection main result:%v", nodeResult)
		return
	}
	for _, result := range nodeResult {
		/* 节点的结果文件中只会存在一个键值对， 否则错误 */
		if len(result) != 1 {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO]Error node level detection minor result:%v", result)
			return
		}
		for _, resStruct := range result {
			/* current节点侧结果无异常 */
			if resStruct.IsSlow == 0 {
				return
			}
			mergedData.SlowCalculateRanks =
				append(mergedData.SlowCalculateRanks, resStruct.SlowCalculateRanks...)
			mergedData.SlowIORanks = append(mergedData.SlowIORanks, resStruct.SlowIORanks...)
			/* 先收集全部节点侧tp慢通信域 */
			mergedData.SlowCommunicationDomains =
				append(mergedData.SlowCommunicationDomains, resStruct.SlowCommunicationDomains...)
			mergedData.SlowHostNodes = append(mergedData.SlowHostNodes, resStruct.SlowHostNodes...)
			mergedData.SlowCommunicationRanks =
				append(mergedData.SlowCommunicationRanks, resStruct.SlowSendRanks...)
			/* 取出最大劣化等级 */
			getCurDetectionJobMaxDegradationLevel(mergedData, resStruct.DegradationLevel)
		}
	}
}

/* 读取节点级结果文件内容并转化为相应的格式 */
func getNodeLevelDetectionResult(filePath string) (bool, config.NodeDetectionResult) {
	/* 获取结果 */
	fileContent, err := fileutils.ReadLimitBytes(filePath, constants.Size500M)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]%s:%v", filePath, err)
		return false, config.NodeDetectionResult{}
	}
	var result config.NodeDetectionResult
	err = json.Unmarshal(fileContent, &result)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]%s:%v", filePath, err)
		return false, config.NodeDetectionResult{}
	}
	return true, result
}

/* 比较两个tp并行域组是否相同 */
func compareTwoSliceSame(tpGroupA []int, tpGroupB []int) bool {
	if len(tpGroupA) != len(tpGroupB) {
		return false
	}
	sort.Ints(tpGroupA)
	sort.Ints(tpGroupB)
	for i := range tpGroupA {
		if tpGroupA[i] != tpGroupB[i] {
			return false
		}
	}
	return true
}

func checkAllInNodeSlowTpCommunications(tpMap map[int]bool, tpGroup []int) bool {
	for _, rankId := range tpGroup {
		/* 存在必为true */
		if _, exist := tpMap[rankId]; !exist {
			return false
		}
	}
	return true
}

/* 切片转map bool类型以方便判断是否存在某个元素 */
func transferSliceToMapBool(tpGroup []int) map[int]bool {
	ret := make(map[int]bool)
	for _, value := range tpGroup {
		ret[value] = true
	}
	return ret
}

/* 节点侧tp并行域可能不完整，根据集群侧进行调整 */
func nodeLevelSlowTpCommunicateDomainIntegral(tpInfo [][]int, nodeSlowTpDomains *[][]int) [][]int {
	if len(tpInfo) == 0 || len(*nodeSlowTpDomains) == 0 {
		return [][]int{}
	}
	tpSlowCommunications := make([][]int, 0)
	tpAllRanks := make([]int, 0)
	/* 将tp并行域结果汇总 */
	for _, nodeTpGroup := range *nodeSlowTpDomains {
		tpAllRanks = append(tpAllRanks, nodeTpGroup...)
	}
	tpAllRanksMap := transferSliceToMapBool(tpAllRanks)
	/* 查看每一个集群侧tp组中卡是否都存在于结果中 */
	for _, tpGroup := range tpInfo {
		if checkAllInNodeSlowTpCommunications(tpAllRanksMap, tpGroup) {
			tpSlowCommunications = append(tpSlowCommunications, tpGroup)
		}
	}
	return tpSlowCommunications
}

/* 获取当前任务所有节点侧检测结果 */
func getCurJobAllNodeResultInfo(filePaths []string, tpInfo [][]int) config.ClusterJobResult {
	mergedData := config.ClusterJobResult{
		SlowCalculateRanks:       []int{},
		SlowCommunicationDomains: [][]int{},
		SlowCommunicationRanks:   []int{},
		SlowHostNodes:            []string{},
		SlowIORanks:              []int{},
	}
	mergedData.DegradationLevel = degradationLevelZero
	for _, filePath := range filePaths {
		flag, curData := getNodeLevelDetectionResult(filePath)
		if !flag {
			continue
		}
		getGatherData(&mergedData, curData)
	}
	/* 根据集群侧完整tp并行域进行处理节点侧tp slow通信域(当前节点侧只会检测慢tp通信域) */
	mergedData.SlowCommunicationDomains =
		nodeLevelSlowTpCommunicateDomainIntegral(tpInfo, &mergedData.SlowCommunicationDomains)
	/* 节点侧有检测结果一定有劣化等级，有劣化等级一定slow，集群侧亦 */
	if mergedData.DegradationLevel != degradationLevelZero {
		mergedData.IsSlow = 1
	}
	return mergedData
}

/* 合并集群侧topo中pp或tp并行域信息 */
func mergeTopoPpParallelInfo(topoData map[string]any, Parallel *[][]int, target string) bool {
	for _, parallelMap := range topoData {
		parallelInfo, ok := parallelMap.(map[string]any)
		if !ok {
			hwlog.RunLog.Error("[SLOWNODE ALGO]Parallel get failed!")
			return false
		}
		groupNameStr, exist := parallelInfo[dataFIleFieldGroupName]
		if !exist {
			hwlog.RunLog.Error("[SLOWNODE ALGO]group_name not exist!")
			return false
		}
		groupName, ok := groupNameStr.(string)
		if !ok {
			hwlog.RunLog.Error("[SLOWNODE ALGO]Transfer groupName failed!")
			return false
		}
		/* pp或tp并行域信息 */
		if groupName != target {
			continue
		}
		/* 存储pp或tp并行域信息 */
		GroupInfo, exist := parallelInfo[dataFIleFieldGlobalRanks]
		if !exist {
			hwlog.RunLog.Error("[SLOWNODE ALGO]global_ranks not exist!")
			return false
		}
		GroupStr, ok := GroupInfo.([]any)
		if !ok {
			hwlog.RunLog.Error("[SLOWNODE ALGO]transfer global_ranks failed!")
			return false
		}
		GroupInt := config.TransferFloatArrayToInt(GroupStr)
		if GroupInt == nil {
			continue
		}
		*Parallel = append(*Parallel, GroupInt)
	}
	if len(*Parallel) == 0 {
		/* 空表示没有并行域，该并行域中仅单卡自己或文件中没有该内容，不停止检测 */
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]empty %s parallel info!", target)
		return true
	}
	return true
}

/* 从集群侧topo文件中获取指定并行域信息 */
func getParallelDomain(filePath string, target string) [][]int {
	data, err := fileutils.ReadLimitBytes(filePath, constants.Size10M)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]Read topo failed:%v", err)
		return nil
	}
	parallel := make([][]int, 0)
	/* 解析集群侧任务级完整topology文件JSON 数据到结构体 */
	var topoData map[string]any
	err = json.Unmarshal(data, &topoData)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]%v", err)
		return nil
	}
	if !mergeTopoPpParallelInfo(topoData, &parallel, target) {
		return nil
	}
	return parallel
}

/* 获取指定路径下的topo中pp并行域信息， 若存在EP或CP则不进行检测 */
func getJobLevelPpParallelDomain(jobPath string) (bool, [][]int) {
	topoFile := filepath.Join(jobPath, config.JobLevelTopologyFileName)
	ppInfo := getParallelDomain(topoFile, ppParallelDomainName)
	if ppInfo == nil {
		return false, nil
	}
	return true, ppInfo
}

func getJobLevelTpParallelDomain(jobPath string) (bool, [][]int) {
	topoFile := filepath.Join(jobPath, config.JobLevelTopologyFileName)
	tpInfo := getParallelDomain(topoFile, tpParallelDomainName)
	if tpInfo == nil {
		return false, nil
	}
	return true, tpInfo
}
