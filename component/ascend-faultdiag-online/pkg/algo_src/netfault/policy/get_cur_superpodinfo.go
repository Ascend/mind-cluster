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

// Package policy is used for processing superpod information
package policy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/netfault/algo"
	"ascend-faultdiag-online/pkg/algo_src/netfault/controllerflags"
	"ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/fileutils"
)

const (
	numSplits  = 2
	configFile = "cathelper.conf"

	minSuperPodRoceDetectionNums = 5
	minSuperPodRoceNpuNumsRate   = 0.01

	// DiagVersionA3 is the fault detection A3 string
	DiagVersionA3 = "A3"
	// DiagVersionA5 is the fault detection A5 string
	DiagVersionA5 = "A5"
	// DiagVersionServer is the inference server
	DiagVersionServer = "800I-SuperPod-A5-8"

	logPrintInterval = 10
	// NpuType is the Network Topology Type
	NpuType = "npu_type"
	// ServerIdMap is the mapping between nodeName and serverId
	ServerIdMap = "serverIdMap"

	level3 = 3
	level1 = 1
	level0 = 0
)

/* 解析server级别topo需要的参数 */
type superPodParam struct {
	superPodId     string
	protocol       string
	rack           *RackInfo
	server         *ServerInfo
	npu            *NpuInfo
	protocolPorts  *PortInfo
	protocolLevels *LevelElement
}

type npuMapParam struct {
	superPodInfo   *SuperPodInfo
	typeStr        string
	rackNpuMap     map[string]bool
	serverTopology *RackTopology
}

type parseTopoParam struct {
	topoServerDirPath  []string
	superPodInfo       *SuperPodInfo
	superPodRackNpuMap map[string]map[string]bool
	typeStr            string
	rackAndServerInfo  [][]string
	superPodPath       string
}

func spliceSuperPodFilePath(superPodPath string) string {
	copyPath := superPodPath
	/* 从/xx/xx/super-pod-0/ 获取当前超节点id */
	// 获取路径的最后一部分
	lastLevel := filepath.Base(copyPath)

	// 去除末尾的斜杠
	superPodJsonFile := strings.TrimSuffix(lastLevel, "/")

	fileName := superPodJsonFile + ".json"
	retStr := superPodPath + "/" + fileName
	retStr = filepath.Clean(retStr)
	hwlog.RunLog.Infof("[NETFAULT ALGO]Read superPodJsonFile: %s", retStr)
	return retStr
}

/* 获取超级点内探测pingList和super-pod-i.json内容 */
func getCurrentSuperPodInfo(
	superPodPath string,
	detectObj *algo.NetDetect) (*SuperPodInfo, map[string]any) {
	if superPodPath == "" {
		hwlog.RunLog.Error("[NETFAULT ALGO]Invalid config path")
		return nil, nil
	}

	superPodJsonFile := spliceSuperPodFilePath(superPodPath)
	superPodInfo, fullMesh, linkPath := processSuperPodJson(superPodJsonFile, superPodPath)
	if superPodInfo == nil && linkPath == nil &&
		len(fullMesh) == 0 {
		return nil, nil
	}
	/* 拼接算法生成pingList接口的入参 */
	algoPingListInput := spliceAlgorithmInput(fullMesh, linkPath)
	if algoPingListInput == nil {
		return nil, nil
	}
	jsonPingList := detectObj.GenPingStrategy(algoPingListInput)
	if jsonPingList == nil {
		return nil, nil
	}
	return superPodInfo, jsonPingList
}

// 解析superPodJsonFile文件
func processSuperPodJson(superPodJsonFile string, superPodPath string) (*SuperPodInfo, []string, map[string][]string) {
	if !loopWaitFile(superPodJsonFile, superPodPath) {
		return nil, nil, nil
	}
	superPodInfo := readConfigMap(superPodJsonFile)
	if superPodInfo == nil {
		return nil, nil, nil
	}
	switch superPodInfo.Version {
	case DiagVersionA3:
		fullMesh, linkPath := GetCurSuperPodInfoFromMapA3(superPodInfo)
		return superPodInfo, fullMesh, linkPath
	default:
		hwlog.RunLog.Errorf("[NETFAULT ALGO]%s version info error, the value %s",
			superPodJsonFile, superPodInfo.Version)
		return nil, nil, nil
	}
}

// SetCallAlgorithmParamInfo 设置算法参数
func SetCallAlgorithmParamInfo(superPodId int, superPodFilePath string,
	callAlgorithmParam map[string]any) error {
	if callAlgorithmParam == nil {
		return errors.New("callAlgorithmParam is nullptr")
	}

	superPodFile := fmt.Sprintf("super-pod-%d.json", superPodId)
	superPodFile = superPodFilePath + "/" + superPodFile
	superPodFile = filepath.Clean(superPodFile)

	if !loopWaitFile(superPodFile, superPodFilePath) {
		return errors.New("loop wait failed")
	}
	superPodInfo := readConfigMap(superPodFile)
	if superPodInfo == nil {
		return errors.New("super pod info is nil")
	}

	if superPodInfo.Version != DiagVersionA3 {
		return fmt.Errorf("unexpected %s version", superPodFile)
	}
	callAlgorithmParam[NpuType] = superPodInfo.Version
	/* A3 */
	callAlgorithmParam["pingObjType"] = 1

	// A3网络结构设置nodeName与serverId的映射
	return getWorKMapping(callAlgorithmParam, superPodInfo)
}

func getWorKMapping(callAlgorithmParam map[string]any, superPodInfo *SuperPodInfo) error {
	if superPodInfo == nil {
		return errors.New("the superPodInfo is empty")
	}
	if superPodInfo.NodeDeviceMap == nil {
		return errors.New("the NodeDeviceMap is empty")
	}
	serverIdMap, ok := callAlgorithmParam[ServerIdMap].(map[string]string)
	if !ok {
		return errors.New("callAlgorithmParam ServerId Map format error")
	}
	for workId, workInfo := range superPodInfo.NodeDeviceMap {
		if workInfo == nil || len(workInfo.NodeName) == 0 {
			return fmt.Errorf("get work %s NodeName error", workId)
		}
		if len(workInfo.ServerID) == 0 {
			return fmt.Errorf("get work %s ServerId error", workId)
		}
		serverIdMap[workInfo.ServerID] = workInfo.NodeName
	}
	return nil
}

// GetTargetSuperPodNpuMap get target super pod npu eid or ip map
func GetTargetSuperPodNpuMap(superPodFilePath string,
	superPodId int) (bool, map[string]algo.NpuInfo) {
	superPodFile := fmt.Sprintf("super-pod-%d.json", superPodId)
	superPodFile = superPodFilePath + "/" + superPodFile
	superPodFile = filepath.Clean(superPodFile)
	var superPodInfo *SuperPodInfo
	var npuNetplaneInfo map[string][]string
	if !loopWaitFile(superPodFile, superPodFilePath) {
		return false, nil
	}
	superPodInfo = readConfigMap(superPodFile)
	if superPodInfo == nil {
		hwlog.RunLog.Error("[NETFAULT ALGO]read config map failed: superPodInfo is nil")
		return false, nil
	}
	var npuInfoMap = make(map[string]algo.NpuInfo)
	switch superPodInfo.Version {
	case DiagVersionA3:
		_, npuNetplaneInfo = GetCurSuperPodInfoFromMapA3(superPodInfo)
		if len(npuNetplaneInfo) == 0 {
			hwlog.RunLog.Error("[NETFAULT ALGO]npu netplane link info is empty!")
			return false, nil
		}
		npuInfoMap = ExtractNPUMapA3(npuNetplaneInfo)
	default:
		hwlog.RunLog.Errorf("[NETFAULT ALGO]%s version info error, the value %s", superPodFile, superPodInfo.Version)
		return false, nil

	}
	return true, npuInfoMap
}

/* loop wait file */
func loopWaitFile(filePath string, superPodDirPath string) bool {
	for i := 0; i < maxRetryTime && !controllerflags.IsControllerExited.GetState() &&
		CheckCurSuperPodConfigSwitch(superPodDirPath); i++ {
		_, err := os.Stat(filePath)
		/* 不管错误类型 */
		if err != nil && os.IsNotExist(err) {
			if i == maxRetryTime-1 {
				hwlog.RunLog.Errorf("[NETFAULT ALGO]%s retry max time failed!", filePath)
				return false
			}
			if i%logPrintInterval == 0 {
				hwlog.RunLog.Warnf("[NETFAULT ALGO]retry: %d, failed: %v", i+1, err)
			}
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	/* 总体开关检查 */
	if controllerflags.IsControllerExited.GetState() {
		hwlog.RunLog.Info("[NETFAULT ALGO]network detection off")
		return false
	}
	/* 当前超节点开关检查 */
	if !CheckCurSuperPodConfigSwitch(superPodDirPath) {
		hwlog.RunLog.Infof("[NETFAULT ALGO]%s detection switch(off)", superPodDirPath)
		return false
	}
	return true
}

func isAlphanumeric(s string) bool {
	curPattern := "^[a-zA-Z0-9]+$"
	regex := regexp.MustCompile(curPattern)
	return regex.MatchString(s)
}

func containsElement(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func isPureNumber(str string) bool {
	matched, err := regexp.MatchString(`^\d+$`, str)
	if err != nil {
		return false
	}
	return matched
}

func isPureLetter(str string) bool {
	matched, err := regexp.MatchString(`^[a-zA-Z]+$`, str)
	if err != nil {
		return false
	}
	return matched
}

// ReadConfigFromFile 从key=value配置文件中获取指定的所有key
func ReadConfigFromFile(fileContent []byte, targetKeys []string) map[string]any {
	callAlgorithmParam := make(map[string]any)
	scanner := bufio.NewScanner(bytes.NewReader(fileContent))
	for scanner.Scan() {
		line := scanner.Text()
		// 跳过空行和注释行
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		// 解析键值对
		parts := strings.SplitN(line, "=", numSplits)
		if len(parts) != numSplits {
			hwlog.RunLog.Errorf("[NETFAULT ALGO]Invalid line format: %s", line)
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		/* 非预期值 */
		if !isAlphanumeric(value) {
			continue
		}
		if isPureNumber(value) && containsElement(targetKeys, key) {
			intValue, err := strconv.Atoi(value)
			if err == nil {
				callAlgorithmParam[key] = intValue
			}
		} else if isPureLetter(value) && containsElement(targetKeys, key) {
			callAlgorithmParam[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]Error reading file: %v", err)
		return nil
	}
	return callAlgorithmParam
}

// CheckCurSuperPodConfigSwitch 判断某个超节点检测开关
func CheckCurSuperPodConfigSwitch(superPodPath string) bool {
	configPath := filepath.Join(superPodPath, configFile)
	/* 需要文件权限、存在、软链接检查等 */
	fileContent, err := fileutils.ReadLimitBytes(configPath, constants.Size10M)
	if err != nil {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]Open:%v", err)
		return false
	}
	target := []string{"netFault"}
	configParam := ReadConfigFromFile(fileContent, target)
	if len(configParam) == 0 {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]netfault field is not exist in %s", configPath)
		return false
	}
	/* 检查开关, 上面接口中取的是唯一目标 */
	flag := configParam["netFault"]
	if value, ok := flag.(string); ok && value == "on" {
		return true
	}
	return false
}

func getReasoningServerSuperPodNpuMap(superPodInfo *SuperPodInfo) map[string]algo.NpuInfo {
	npuEidMap := make(map[string]algo.NpuInfo)
	rackMap := superPodInfo.RackMap
	if len(rackMap) != 1 {
		hwlog.RunLog.Error("[NETFAULT ALGO]error rack numbers")
		return nil
	}
	var rackSingle *RackInfo
	for _, rack := range rackMap {
		rackSingle = rack
		break
	}
	if rackSingle == nil || len(rackSingle.ServerMap) == 0 {
		hwlog.RunLog.Error("[NETFAULT ALGO]error rack inner server mumbers")
		return nil
	}
	serverIds := make([]int, 0)
	for _, server := range rackSingle.ServerMap {
		serverId, err := strconv.Atoi(server.ServerIndex)
		if err != nil {
			hwlog.RunLog.Errorf("[NETFAULT ALGO]%v", err)
			continue
		}
		serverIds = append(serverIds, serverId)
	}
	sort.Ints(serverIds)
	for index, serverId := range serverIds {
		strId := strconv.Itoa(serverId)
		server := rackSingle.ServerMap[strId]
		for _, npu := range server.NpuMap {
			if npu == nil || npu.LevelList == nil || len(npu.LevelList) == 0 {
				continue
			}
			getReasoningServerNpuInfo(npuEidMap, index, npu, server.ServerIndex, superPodInfo.SuperPodID)
		}
	}
	return npuEidMap
}

func handleReasoningServer(superPodInfo *SuperPodInfo, superPodPath string) (map[string]algo.NpuInfo, bool) {
	if superPodInfo == nil || len(superPodInfo.RackMap) == 0 {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]%s: invalid or empty superPodInfo", superPodPath)
		return nil, false
	}
	npuMap := getReasoningServerSuperPodNpuMap(superPodInfo)
	if len(npuMap) == 0 {
		return nil, false
	}
	return npuMap, true
}

/* 根据协议构建算法需要的链路形式(1D、2D与roce链路格式都不相同) */
func formatDifferentByProtocol(npuNetPlanePaths map[string][]string,
	param superPodParam,
	id int) bool {
	if npuNetPlanePaths == nil {
		hwlog.RunLog.Error("[NETFAULT ALGO]empty npuNetPlanePaths!")
		return false
	}
	// key使用rack-os
	key := param.rack.RackID + "-" + param.server.ServerIndex
	if _, exist := npuNetPlanePaths[key]; !exist {
		npuNetPlanePaths[key] = make([]string, 0)
	}
	var link string
	switch param.protocol {
	case "ROCE":
		link =
			fmt.Sprintf("NA.ROCESwitch:0#NA.SuperPod-%s:0#NA.NSlot-0:0#NPU-%s.%s:0",
				param.superPodId, param.npu.PhyId, param.protocolLevels.RankAddrList[id].Addr)
	default:
		hwlog.RunLog.Errorf("[NETFAULT ALGO]undefined protocol:%s", param.protocol)
		return false
	}
	npuNetPlanePaths[key] = append(npuNetPlanePaths[key], link)
	return true
}

func getRoceNpuLinkPathInfo(npuNetPlanePaths map[string][]string, param superPodParam,
	npuMap map[string]string) {
	if npuNetPlanePaths == nil || npuMap == nil {
		return
	}
	for _, levelInfo := range param.npu.LevelList {
		if levelInfo.NetLayer != level3 {
			continue
		}
		/* roce每个os下每个npu出口ip相同 */
		for i := 0; i < len(levelInfo.RankAddrList); i++ {
			param.protocolLevels = &levelInfo
			flag := formatDifferentByProtocol(npuNetPlanePaths, param, i)
			/* 超节点间每张卡取一个ip就可以了 */
			if _, exist := npuMap[levelInfo.RankAddrList[i].Addr]; !exist && param.protocol == "ROCE" && flag {
				npuMap[levelInfo.RankAddrList[i].Addr] = param.superPodId + "_" +
					param.rack.RackID + "_" + param.server.ServerIndex
				break
			}
		}
	}
}

func getRoceServerLevelNpuLinkPathInfo(npuNetPlanePaths map[string][]string,
	param superPodParam) map[string]string {
	// key:address; value:=superPod_rackId_osId, roce每个os下所有npu出口ip相同
	npuMap := make(map[string]string)
	for _, server := range param.rack.ServerMap {
		if server == nil || server.NpuMap == nil || len(server.NpuMap) == 0 {
			continue
		}
		for _, npu := range server.NpuMap {
			if npu == nil || npu.LevelList == nil || len(npu.LevelList) == 0 {
				continue
			}
			param.server = server
			param.npu = npu
			getRoceNpuLinkPathInfo(npuNetPlanePaths, param, npuMap)
		}
	}
	return npuMap
}

func mergeRoceNpuInfoMap(dst map[string]string, src map[string]string) {
	if src == nil || dst == nil {
		return
	}
	for k, v := range src {
		if _, exist := dst[k]; !exist {
			dst[k] = v
		}
	}
}

func getRoceNpuLinkPathFromSuperPodJson(superPodInfo *SuperPodInfo,
	protocol string) (map[string][]string, map[string]string) {
	if superPodInfo == nil {
		hwlog.RunLog.Error("[NETFAULT ALGO]invalid super pod information")
		return nil, nil
	}

	if superPodInfo.RackMap == nil || len(superPodInfo.RackMap) == 0 {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]super pod %s empty rack info", superPodInfo.SuperPodID)
		return nil, nil
	}
	npuRoceInfoMap := make(map[string]string)
	npuNetPlanePaths := make(map[string][]string)
	for _, rack := range superPodInfo.RackMap {
		if rack == nil || rack.ServerMap == nil || len(rack.ServerMap) == 0 {
			continue
		}
		param := superPodParam{}
		param.superPodId = superPodInfo.SuperPodID
		param.protocol = protocol
		param.rack = rack
		npuMap := getRoceServerLevelNpuLinkPathInfo(npuNetPlanePaths, param)
		mergeRoceNpuInfoMap(npuRoceInfoMap, npuMap)
	}
	if len(npuNetPlanePaths) == 0 {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]super pod %s empty npu link path", superPodInfo.SuperPodID)
	}
	return npuNetPlanePaths, npuRoceInfoMap
}

// 从按约定的格式的字符串中取出相关信息
func getNpuInfoFromNpuLinkPath(npuMap map[string]algo.NpuInfo, npuLinkPath string) {
	if npuMap == nil {
		return
	}
	var superPodId, npuId, ip1, ip2, ip3, ip4 int
	lenParam := 6
	num, err := fmt.Sscanf(npuLinkPath, "NA.ROCESwitch:0#NA.SuperPod-%d:0#NA.NSlot-0:0#NPU-%d.%d.%d.%d.%d:0",
		&superPodId, &npuId, &ip1, &ip2, &ip3, &ip4)
	if err != nil || num != lenParam {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]roceIp:%s(%v)", npuLinkPath, err)
		return
	}
	slotId := npuId / perBoardNpus
	ip := strconv.Itoa(ip1) + "." + strconv.Itoa(ip2) + "." + strconv.Itoa(ip3) + "." + strconv.Itoa(ip4)
	npuMap[ip] = algo.NpuInfo{
		NpuNumber:    npuId,
		SuperPodName: "SuperPod-" + strconv.Itoa(superPodId),
		SlotName:     "NSlot-" + strconv.Itoa(slotId),
	}
}

// 随机选出一些NPU LINK PATH信息，并返回这些链路的npu与ip或eid的映射关系
func getRoceLimitNpuNumsPerSuperPod(npuNetPlanePaths map[string][]string) ([]string, map[string]algo.NpuInfo, bool) {
	if len(npuNetPlanePaths) == 0 {
		return nil, nil, false
	}
	// 超节点os总数
	total := len(npuNetPlanePaths)
	tmp := float64(total) * minSuperPodRoceNpuNumsRate
	ret := make([]string, 0)
	npuMap := make(map[string]algo.NpuInfo)
	if tmp < minSuperPodRoceDetectionNums && total >= minSuperPodRoceDetectionNums {
		tmp = minSuperPodRoceDetectionNums
	}
	uniqueMap := make(map[string]bool)
	for key, path := range npuNetPlanePaths {
		if _, exist := uniqueMap[key]; !exist && len(path) > 0 {
			uniqueMap[key] = true
			ret = append(ret, path[0])
			getNpuInfoFromNpuLinkPath(npuMap, path[0])
		}
		if len(ret) == int(tmp) {
			break
		}
	}
	if len(ret) < minSuperPodRoceDetectionNums {
		hwlog.RunLog.Error("[NETFAULT ALGO]not enough npu roce eid nums")
		return nil, nil, false
	}
	return ret, npuMap, true
}

func getSuperPodIdAndOsIdFromRoceMap(npuMap map[string]algo.NpuInfo, roceMap map[string]string) {
	if npuMap == nil || roceMap == nil {
		return
	}
	for k, v := range roceMap {
		if npuInfo, exist := npuMap[k]; exist {
			var superPodId, rackId, osId int
			lenParam := 6
			num, err := fmt.Sscanf(v, "%d_%d_%d", &superPodId, &rackId, &osId)
			if err != nil || num != lenParam {
				hwlog.RunLog.Error(err)
				continue
			}
			npuMap[k] = algo.NpuInfo{
				NpuNumber:    npuInfo.NpuNumber,
				SuperPodName: npuInfo.SuperPodName,
				OsName:       strconv.Itoa(osId),
				RackName:     "Rack-" + strconv.Itoa(rackId),
				SlotName:     npuInfo.SlotName,
			}
		}
	}
}

// GetSuperPodsRoceNpuInfo 获取超节点间探测npu信息和npu与port映射关系
func GetSuperPodsRoceNpuInfo(superPodInfoPaths []string) (map[string]algo.NpuInfo, map[string]interface{}) {
	/* 超节点间Id需要从小到大排序，超节点内rack需要从小到大排序，便于算法区分使用 */
	if len(superPodInfoPaths) == 0 {
		hwlog.RunLog.Error("[NETFAULT ALGO]empty roce super pod paths")
		return nil, nil
	}
	npuLinkPaths := make(map[string]interface{})
	npuLinkPaths["netplane_0"] = make([]string, 0)
	allSuperPodNpuMap := make(map[string]algo.NpuInfo)
	count := 0
	for _, path := range superPodInfoPaths {
		superPodInfo := readConfigMap(path)
		if superPodInfo == nil {
			continue
		}
		/* roce仅会有一个值"netplane_0" */
		npuNetPlanePaths, npuRoceMap := getRoceNpuLinkPathFromSuperPodJson(superPodInfo, "ROCE")
		if npuNetPlanePaths == nil || len(npuNetPlanePaths) == 0 || npuRoceMap == nil {
			continue
		}
		/* 取%1或至少5个 */
		npuLinks, npuMap, flag := getRoceLimitNpuNumsPerSuperPod(npuNetPlanePaths)
		if !flag {
			continue
		}
		/* 补充os id和rack id信息 */
		getSuperPodIdAndOsIdFromRoceMap(npuMap, npuRoceMap)
		if paths, ok := npuLinkPaths["netplane_0"].([]string); ok {
			npuLinkPaths["netplane_0"] = append(paths, npuLinks...)
		}
		allSuperPodNpuMap = mergeNpuEidMap(allSuperPodNpuMap, npuMap)
		count++
	}
	if count <= 1 {
		hwlog.RunLog.Error("[NETFAULT ALGO]valid roce super pod is not enough(at least two super pods)")
		return nil, nil
	}
	return allSuperPodNpuMap, npuLinkPaths
}

func mergeNpuEidMap(npuEidMapOutRack map[string]algo.NpuInfo,
	npuEidMapFromTopo map[string]algo.NpuInfo) map[string]algo.NpuInfo {
	npuInfoMap := make(map[string]algo.NpuInfo)
	if npuEidMapOutRack != nil {
		for key, value := range npuEidMapOutRack {
			npuInfoMap[key] = value
		}
	}
	if npuEidMapFromTopo != nil {
		for key, value := range npuEidMapFromTopo {
			if _, exist := npuInfoMap[key]; !exist {
				npuInfoMap[key] = value
			}
		}
	}
	return npuInfoMap
}

func getReasoningServerNpuInfo(npuEidMap map[string]algo.NpuInfo,
	serverIndex int, npu *NpuInfo, serverId, superPodId string) {
	for _, levelInfo := range npu.LevelList {
		if levelInfo.NetLayer != level1 {
			continue
		}
		for i := 0; i < len(levelInfo.RankAddrList); i++ {
			id, err := strconv.Atoi(npu.PhyId)
			if err != nil {
				hwlog.RunLog.Errorf("[CONTROLLER]%s:%v", npu.PhyId, err)
				continue
			}
			npuPhyId := serverIndex*perBoardNpus + id
			slotId := npuPhyId / perBoardNpus
			npuInfo := algo.NpuInfo{
				RackName:     "Rack-0", // 推理服务器都写0
				NpuNumber:    npuPhyId,
				SlotName:     "NSlot-" + strconv.Itoa(slotId),
				NetPlaneId:   "netplane_" + levelInfo.RankAddrList[i].PlaneId,
				SuperPodName: "SuperPod-" + superPodId,
				OsName:       serverId}
			if _, exist := npuEidMap[levelInfo.RankAddrList[i].Addr]; !exist && npuEidMap != nil {
				npuEidMap[levelInfo.RankAddrList[i].Addr] = npuInfo
			}
		}
	}
}

func getOneTopoFilePath(superPod string, superPodInfo *SuperPodInfo) string {
	var rackId string
	var serverId string
	success := false
	for _, v := range superPodInfo.RackMap {
		rackId = v.RackID
		if len(v.ServerMap) == 0 {
			continue
		}
		for _, server := range v.ServerMap {
			serverId = server.ServerIndex
			success = true
			break
		}
		break
	}
	if !success {
		hwlog.RunLog.Errorf("not found server level topo in %s", superPod)
		return ""
	}
	file := filepath.Join(superPod, fmt.Sprintf("rack-%s", rackId), fmt.Sprintf("topo_%s.json", serverId))
	return file
}

func getNetWorkType(superPod string, superPodInfo *SuperPodInfo) string {
	if superPodInfo == nil || len(superPodInfo.RackMap) == 0 {
		hwlog.RunLog.Error("empty super pod info, get network type failed!")
		return ""
	}
	topoFile := getOneTopoFilePath(superPod, superPodInfo)
	if topoFile == "" {
		return ""
	}
	var data []byte
	for i := 0; i < maxRetryTime && !controllerflags.IsControllerExited.GetState() &&
		CheckCurSuperPodConfigSwitch(superPod); i++ {
		fileData, err := os.ReadFile(topoFile)
		if err != nil && os.IsNotExist(err) {
			if i%logPrintInterval == 0 {
				hwlog.RunLog.Warnf("%v (retry:%d)", err, i+1)
			}
			time.Sleep(time.Duration(1) * time.Second)
			continue
		} else if err != nil {
			hwlog.RunLog.Error(err)
			return ""
		}
		data = fileData
		break
	}
	if controllerflags.IsControllerExited.GetState() || !CheckCurSuperPodConfigSwitch(superPod) {
		return ""
	}
	if len(data) == 0 {
		hwlog.RunLog.Errorf("[CONTROLLER]empty %s", topoFile)
		return ""
	}
	var obj RackTopology
	err := json.Unmarshal(data, &obj)
	if err != nil {
		hwlog.RunLog.Error(err)
		return ""
	}
	if obj.HardwareType == "Atlas 950 SuperPod 2D" {
		return "2D"
	} else if obj.HardwareType == "Atlas 950 SuperPod 1D" {
		return "1D"
	} else {
		hwlog.RunLog.Errorf("Unknown hardware type %s", obj.HardwareType)
		return ""
	}
}
