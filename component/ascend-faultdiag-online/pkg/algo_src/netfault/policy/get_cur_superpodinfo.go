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
	"reflect"
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

	// NpuType 网络拓扑类型
	NpuType = "npu_type"
	// ServerIdMap nodeName与serverId的映射
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
	case DiagVersionA5:
		typeStr := getNetWorkType(superPodPath, superPodInfo)
		if typeStr != "1D" && typeStr != "2D" {
			return nil, nil, nil
		}
		npuFmlink, linkPath, _ := GetA5CurSuperPod1D2DNpuInfo(superPodPath, superPodInfo)
		return superPodInfo, npuFmlink, linkPath
	case DiagVersionA3:
		fullMesh, linkPath := GetCurSuperPodInfoFromMapA3(superPodInfo)
		return superPodInfo, fullMesh, linkPath
	case DiagVersionServer:
		linkPath := getA51D2DSuperPodNpuLinkPath(superPodInfo, "reasoningServer")
		return superPodInfo, nil, linkPath
	default:
		hwlog.RunLog.Errorf("[NETFAULT ALGO]%s version info error, the value %s",
			superPodJsonFile, superPodInfo.Version)
		return nil, nil, nil
	}
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
	success := false
	switch superPodInfo.Version {
	case DiagVersionA5:
		npuInfoMap, success = handleA5NpuMapInfo(superPodInfo, superPodFilePath)
		if !success {
			return false, nil
		}
	case DiagVersionA3:
		_, npuNetplaneInfo = GetCurSuperPodInfoFromMapA3(superPodInfo)
		if len(npuNetplaneInfo) == 0 {
			hwlog.RunLog.Error("[NETFAULT ALGO]npu netplane link info is empty!")
			return false, nil
		}
		npuInfoMap = ExtractNPUMapA3(npuNetplaneInfo)
	case DiagVersionServer:
		npuInfoMap, success = handleReasoningServer(superPodInfo, superPodFilePath)
		if !success {
			return false, nil
		}
	default:
		hwlog.RunLog.Errorf("[NETFAULT ALGO]%s version info error, the value %s", superPodFile, superPodInfo.Version)
		return false, nil

	}
	return true, npuInfoMap
}

/* loop wait file (used in RoCE plane scenarios between non-super nodes)*/
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

func getNpuServerIdFromRackMap(npuId int, rack *RackInfo) string {
	if len(rack.ServerMap) == 0 {
		hwlog.RunLog.Error("empty super pod server map!")
		return ""
	}
	for _, server := range rack.ServerMap {
		if len(server.NpuMap) == 0 {
			hwlog.RunLog.Error("empty super pod npu map!")
			return ""
		}
		if _, exist := server.NpuMap[strconv.Itoa(npuId)]; exist {
			return server.ServerIndex
		}
	}
	return ""
}

func storeA51D2DNpuFmLink(param *npuMapParam, npuFmLink *[]string, srcAddr string, dstAddr string, srcId string) {
	if npuFmLink == nil || param == nil {
		hwlog.RunLog.Error("invalid npuFmLink or npuMap!")
		return
	}
	/* fm direct connection*/
	var link string
	link = fmt.Sprintf("%s:0#%s:0", srcAddr, dstAddr)
	if param.rackNpuMap != nil &&
		param.rackNpuMap[srcId] {
		*npuFmLink = append(*npuFmLink, link)
	}
}

func getNpuMapValueInfoUnit(rackAndServerIds [][]string,
	index int,
	slotId string,
	srcId int,
	serverId string) algo.NpuInfo {
	var rackName string
	if len(rackAndServerIds) == 0 {
		rackName = ""
	} else {
		rackName = "Rack-" + rackAndServerIds[0][index]
	}
	npuInfo := algo.NpuInfo{
		RackName:   rackName,
		SlotName:   "NSlot-" + slotId,
		NpuNumber:  srcId,
		NetPlaneId: "",
		OsName:     serverId,
	}
	return npuInfo
}

func storeA51D2DNpuFmLinkAndNpuEidMapInfo(index int, rackAndServerIds [][]string,
	param *npuMapParam) (map[string]algo.NpuInfo, []string) {
	npuFmLink := make([]string, 0)
	npuInfoMap := make(map[string]algo.NpuInfo)
	/* extract the direct connection of NPU, the mapping relationship between directly connected NPU and EID,
	and concatenate the algorithm input */
	for i := 0; i < len(param.serverTopology.EdgeList); i++ {
		/* Parameter validity check: The value under each rack cannot be less than 0 */
		if param.serverTopology.EdgeList[i].LocalA < 0 || param.serverTopology.EdgeList[i].LocalB < 0 {
			hwlog.RunLog.Error("error topology")
			return map[string]algo.NpuInfo{}, []string{}
		}
		/* add the npu direct connection and the mapping relationship between npu and eid,
		and check whether it is peer to peer and layer 0 */
		if !(param.serverTopology.EdgeList[i].NetLayer == level0 &&
			param.serverTopology.EdgeList[i].LinkType == "PEER2PEER") {
			continue
		}
		/* find eid from levelList according to localAPorts/localBPorts */
		serverIDA := getNpuServerIdFromRackMap(param.serverTopology.EdgeList[i].LocalA,
			param.superPodInfo.RackMap[rackAndServerIds[0][index]])
		serverIDB := getNpuServerIdFromRackMap(param.serverTopology.EdgeList[i].LocalB,
			param.superPodInfo.RackMap[rackAndServerIds[0][index]])
		if serverIDA == "" || serverIDB == "" {
			continue
		}
		localAEid := findEid(serverIDA, param.serverTopology.EdgeList[i].LocalA,
			param.serverTopology.EdgeList[i].LocalAPorts, param.superPodInfo.RackMap[rackAndServerIds[0][index]])
		localBEid := findEid(serverIDB, param.serverTopology.EdgeList[i].LocalB,
			param.serverTopology.EdgeList[i].LocalBPorts, param.superPodInfo.RackMap[rackAndServerIds[0][index]])
		if localAEid == "" || localBEid == "" {
			continue
		}
		storeA51D2DNpuFmLink(param, &npuFmLink, localAEid, localBEid,
			strconv.Itoa(param.serverTopology.EdgeList[i].LocalA))

		/* mapping relationship between npu and eid */
		if _, exist := npuInfoMap[localAEid]; !exist && npuInfoMap != nil {
			npuInfoMap[localAEid] =
				getNpuMapValueInfoUnit(rackAndServerIds, index,
					strconv.Itoa(param.serverTopology.EdgeList[i].LocalA/perBoardNpus),
					param.serverTopology.EdgeList[i].LocalA, serverIDA)
		}

		if _, exist := npuInfoMap[localBEid]; !exist && npuInfoMap != nil {
			npuInfoMap[localBEid] = getNpuMapValueInfoUnit(
				rackAndServerIds, index, strconv.Itoa(param.serverTopology.EdgeList[i].LocalB/perBoardNpus),
				param.serverTopology.EdgeList[i].LocalB, serverIDB)
		}
	}
	return npuInfoMap, npuFmLink
}

func findEid(serverId string, npuId int, localPorts []string, rack *RackInfo) string {
	levelLists := rack.ServerMap[serverId].NpuMap[strconv.Itoa(npuId)].LevelList
	var eid string
	for _, levelList := range levelLists {
		if levelList.NetLayer != level0 {
			continue
		}
		for _, addrList := range levelList.RankAddrList {
			if reflect.DeepEqual(addrList.Ports, localPorts) {
				eid = addrList.Addr
				return eid
			}
		}
	}
	hwlog.RunLog.Errorf("find eid failed: serverId %s, npuId %d, localPorts %v, rack %v",
		serverId, npuId, localPorts, rack)
	return eid
}

func parseA5ServerLevelTopologyFile(topoParam *parseTopoParam) ([]string, map[string]algo.NpuInfo) {
	if len(topoParam.topoServerDirPath) == 0 {
		hwlog.RunLog.Error("no server topology file  exist!")
		return nil, nil
	}
	npuFmLink := make([]string, 0)
	var npuInfoMap map[string]algo.NpuInfo
	for index, file := range topoParam.topoServerDirPath {
		/* loop waiting for reading file*/
		if !loopWaitFile(file, topoParam.superPodPath) {
			return nil, nil
		}
		data, err := fileutils.ReadLimitBytes(file, constants.Size10M)
		if err != nil {
			hwlog.RunLog.Errorf("[NETFAULT ALGO]Open:%v", err)
			return nil, nil
		}
		if controllerflags.IsControllerExited.GetState() {
			return nil, nil
		}
		if len(data) == 0 {
			hwlog.RunLog.Errorf("[CONTROLLER]empty or not exist:%s", file)
			continue
		}
		var obj RackTopology
		err = json.Unmarshal(data, &obj)
		if err != nil {
			hwlog.RunLog.Error(err)
			return nil, nil
		}
		if len(obj.EdgeList) == 0 {
			hwlog.RunLog.Error("not found edge list")
		}
		param := npuMapParam{superPodInfo: topoParam.superPodInfo, typeStr: topoParam.typeStr,
			rackNpuMap:     topoParam.superPodRackNpuMap[topoParam.rackAndServerInfo[0][index]],
			serverTopology: &obj}
		npuInfoMapTmp, npuFmLinkTmp :=
			storeA51D2DNpuFmLinkAndNpuEidMapInfo(index, topoParam.rackAndServerInfo, &param)
		npuInfoMap = mergeNpuEidMap(npuInfoMapTmp, npuInfoMap)
		npuFmLink = append(npuFmLink, npuFmLinkTmp...)
		if npuInfoMap == nil {
			return nil, nil
		}
	}
	return npuFmLink, npuInfoMap
}

func getA51D2DNpuLinkPath(npuNetPlanePaths map[string][]string, npu *NpuInfo, rackId string, typeStr string) {
	if npuNetPlanePaths == nil {
		return
	}
	var format string
	if typeStr == "1D" {
		format = "NA.L2-LogicPort%d:0#Rack-%s.L1-LogicPort%d:0#Rack-%s.NSlot-%d:0#NPU-%s.%s:0"
	} else { // "2D"
		format = "NA.L2-LogicPort:0#Rack-%s.NA:0#Rack-%s.NSlot-%d:0#NPU-%s.%s:0"
	}
	for _, levelInfo := range npu.LevelList {
		if levelInfo.NetLayer != level1 {
			continue
		}
		/* eid numbers equals plane numbers */
		for i := 0; i < len(levelInfo.RankAddrList); i++ {
			index := strconv.Itoa(i + 1)
			if _, exist := npuNetPlanePaths[index]; !exist {
				npuNetPlanePaths[index] = make([]string, 0)
			}
			/* string concatenation */
			npuId, err := strconv.Atoi(npu.PhyId)
			if err != nil {
				hwlog.RunLog.Error(err)
				continue
			}
			planeId, err := strconv.Atoi(levelInfo.RankAddrList[i].PlaneId)
			if err != nil {
				hwlog.RunLog.Error(err)
				continue
			}
			slotId := npuId / perBoardNpus
			var link string
			if typeStr == "1D" {
				link = fmt.Sprintf(format, planeId, rackId, planeId, rackId, slotId, npu.PhyId, levelInfo.RankAddrList[i].Addr)
			} else {
				link = fmt.Sprintf(format, rackId, rackId, slotId, npu.PhyId, levelInfo.RankAddrList[i].Addr)
			}
			npuNetPlanePaths[index] = append(npuNetPlanePaths[index], link)
		}
	}
}

/* determine if it's the new 1D and check if ports exist*/
func checkIfNew1D(rackInfo map[string]*RackInfo) bool {
	if rackInfo == nil || len(rackInfo) == 0 {
		hwlog.RunLog.Error("new 1D rack info is nil")
		return false
	}
	var rackI *RackInfo = nil
	for _, rack := range rackInfo {
		rackI = rack
		break
	}
	if rackI == nil || rackI.ServerMap == nil || len(rackI.ServerMap) == 0 {
		hwlog.RunLog.Error("new 1D rack info is nil")
		return false
	}
	var serverI *ServerInfo = nil
	for _, server := range rackI.ServerMap {
		serverI = server
		break
	}
	if serverI == nil || serverI.NpuMap == nil || len(serverI.NpuMap) == 0 {
		hwlog.RunLog.Error("new 1D server info is nil")
		return false
	}
	var npuI *NpuInfo = nil
	for _, npu := range serverI.NpuMap {
		npuI = npu
		break
	}
	if npuI == nil || npuI.LevelList == nil || len(npuI.LevelList) == 0 {
		return false
	}
	return true
}

/* 从super pod.json中获取port中sddrType为target的eid或ip */
func getEidInfoOrIpFromPorts(npuEidMap map[string]algo.NpuInfo, param superPodParam) {
	for _, levelInfo := range param.npu.LevelList {
		if levelInfo.NetLayer != level1 {
			continue
		}
		for i := 0; i < len(levelInfo.RankAddrList); i++ {
			id, err := strconv.Atoi(param.npu.PhyId)
			if err != nil {
				hwlog.RunLog.Errorf("[CONTROLLER]%s:%v", param.npu.PhyId, err)
				continue
			}
			slotId := id / perBoardNpus
			npuInfo := algo.NpuInfo{
				RackName:     "Rack-" + param.rack.RackID,
				NpuNumber:    id,
				SlotName:     fmt.Sprintf("NSlot-%d", slotId),
				NetPlaneId:   fmt.Sprintf("netplane_%s", levelInfo.RankAddrList[i].PlaneId),
				SuperPodName: fmt.Sprintf("SuperPod-%s", param.superPodId),
				OsName:       param.server.ServerIndex}
			if _, exist := npuEidMap[levelInfo.RankAddrList[i].Addr]; !exist && npuEidMap != nil {
				npuEidMap[levelInfo.RankAddrList[i].Addr] = npuInfo
			}
		}
	}
}

func getReasoningServerNpuLinkPath(npuNetPlanePaths map[string][]string,
	serverIds []int, serverMap map[string]*ServerInfo) {
	for index, serverId := range serverIds {
		strId := strconv.Itoa(serverId)
		/* must exist */
		server := serverMap[strId]
		for _, npu := range server.NpuMap {
			if npu == nil || len(npu.LevelList) == 0 {
				continue
			}
			getReasoningServerNpuLinkPathStr(npuNetPlanePaths, npu, index)
		}
	}
}

func getReasoningServerNpuLinkPathStr(npuNetPlanePaths map[string][]string,
	npu *NpuInfo, serverIndex int) {
	if npuNetPlanePaths == nil {
		return
	}
	format := "NA.L2-LogicPort%d:0#Node-%d.L1-LogicPort%d:0#Node-%d.NSlot-%d:0#NPU-%s.%s:0"
	for _, levelInfo := range npu.LevelList {
		if levelInfo.NetLayer != level1 {
			continue
		}
		/* eid numbers equals plane numbers */
		for i := 0; i < len(levelInfo.RankAddrList); i++ {
			index := strconv.Itoa(i + 1)
			if _, exist := npuNetPlanePaths[index]; !exist {
				npuNetPlanePaths[index] = make([]string, 0)
			}
			/* string concatenation */
			npuId, err := strconv.Atoi(npu.PhyId)
			if err != nil {
				hwlog.RunLog.Error(err)
				continue
			}
			planeId, err := strconv.Atoi(levelInfo.RankAddrList[i].PlaneId)
			if err != nil {
				hwlog.RunLog.Error(err)
				continue
			}
			/* calculate real npu id according to server index */
			npuPhyId := serverIndex*perBoardNpus + npuId
			slotId := npuPhyId / perBoardNpus
			link := fmt.Sprintf(format, planeId, serverIndex, planeId, serverIndex, slotId,
				strconv.Itoa(npuPhyId), levelInfo.RankAddrList[i].Addr)
			npuNetPlanePaths[index] = append(npuNetPlanePaths[index], link)
		}
	}
}

func getA51D2DServerLevelInfo(npuNetPlanePaths map[string][]string, rack *RackInfo, typeStr string) {
	/* sort serverId, serverId is unique string */
	if typeStr == "reasoningServer" {
		serverIds := make([]int, 0)
		for _, server := range rack.ServerMap {
			serverId, err := strconv.Atoi(server.ServerIndex)
			if err != nil {
				hwlog.RunLog.Error(err)
				continue
			}
			serverIds = append(serverIds, serverId)
		}
		sort.Ints(serverIds)
		getReasoningServerNpuLinkPath(npuNetPlanePaths, serverIds, rack.ServerMap)
		return
	}
	for _, server := range rack.ServerMap {
		if server == nil || len(server.NpuMap) == 0 {
			continue
		}
		for _, npu := range server.NpuMap {
			if npu == nil || len(npu.LevelList) == 0 {
				continue
			}
			getA51D2DNpuLinkPath(npuNetPlanePaths, npu, rack.RackID, typeStr)
		}
	}
}

/*In super-pod-i.json, the eid of the UBC protocol type in the ports is netplane_I*/
func getA51D2DSuperPodNpuLinkPath(superPodInfo *SuperPodInfo, typeStr string) map[string][]string {
	if superPodInfo == nil {
		hwlog.RunLog.Error("invalid super pod information")
		return nil
	}
	npuNetPlanePaths := make(map[string][]string)
	if len(superPodInfo.RackMap) == 0 {
		hwlog.RunLog.Error("[CONTROLLER]empty rack info")
		return nil
	}
	for _, rack := range superPodInfo.RackMap {
		if rack == nil || len(rack.ServerMap) == 0 {
			continue
		}
		getA51D2DServerLevelInfo(npuNetPlanePaths, rack, typeStr)
	}
	return npuNetPlanePaths
}

/*
return all npu pyhId of each rack of current super pod,
ensure that all the npu direct connection information src passed to the algorithm exists in super-pod.json
*/
func getSuperPodRackLevelNpuMap(superPodInfo *SuperPodInfo) map[string]map[string]bool {
	if superPodInfo == nil || len(superPodInfo.RackMap) == 0 {
		hwlog.RunLog.Error("invalid super pod info")
		return nil
	}
	/* rackId is unique in super pod, npu phyId is unique in rack */
	ret := make(map[string]map[string]bool)
	count := 0
	for _, rack := range superPodInfo.RackMap {
		if len(rack.ServerMap) == 0 {
			continue
		}
		ret[rack.RackID] = make(map[string]bool)
		for _, server := range rack.ServerMap {
			if len(server.NpuMap) == 0 {
				continue
			}
			for _, npu := range server.NpuMap {
				ret[rack.RackID][npu.PhyId] = true
			}
			count++
		}
	}
	if len(ret) == 0 {
		hwlog.RunLog.Error("unmatched any npu in rack")
		return nil
	}
	return ret
}

// GetA5CurSuperPod1D2DNpuInfo get npu direct connection, mapping, link info in topology of rack dir under super pod
func GetA5CurSuperPod1D2DNpuInfo(superPodPath string,
	superPodInfo *SuperPodInfo) ([]string, map[string][]string, map[string]algo.NpuInfo) {
	/* get all rack-I directories under the current path */
	rackNums := len(superPodInfo.RackMap)
	if rackNums == 0 {
		hwlog.RunLog.Error(superPodInfo, " has no rack")
		return nil, nil, nil
	}
	topoServerDirPath := make([]string, 0)
	rackAndServerInfo := make([][]string, 0)
	rackAndServerInfo = append(rackAndServerInfo, make([]string, 0))
	rackAndServerInfo = append(rackAndServerInfo, make([]string, 0))
	superPodRackNpuMap := getSuperPodRackLevelNpuMap(superPodInfo)
	if superPodRackNpuMap == nil {
		return nil, nil, nil
	}
	for _, rack := range superPodInfo.RackMap {
		rackPath := filepath.Join(superPodPath, "rack-"+rack.RackID)
		if len(rack.ServerMap) == 0 {
			continue
		}
		/* get one serverId as topo file suffix of each rack */
		for _, server := range rack.ServerMap {
			topoPath := filepath.Join(rackPath, "topo_"+server.ServerIndex+".json")
			/* judge file exist when loop wait and read file */
			topoServerDirPath = append(topoServerDirPath, topoPath)
			rackAndServerInfo[0] = append(rackAndServerInfo[0], rack.RackID)
			rackAndServerInfo[1] = append(rackAndServerInfo[1], server.ServerIndex)
			break
		}
	}
	typeStr := getNetWorkType(superPodPath, superPodInfo)
	if typeStr == "" {
		return nil, nil, nil
	}
	/* get npu link out of rack info from super-pod.json */
	npuNetPlanePaths := getA51D2DSuperPodNpuLinkPath(superPodInfo, typeStr)
	/* parse ech topo and get direct connection, mapping info */
	param := parseTopoParam{superPodInfo: superPodInfo, typeStr: typeStr,
		topoServerDirPath: topoServerDirPath, superPodRackNpuMap: superPodRackNpuMap,
		rackAndServerInfo: rackAndServerInfo, superPodPath: superPodPath}
	npuFmLink, npuEidMap := parseA5ServerLevelTopologyFile(&param)
	return npuFmLink, npuNetPlanePaths, npuEidMap
}

func getNpuEidMapInfo(npuEidMap map[string]algo.NpuInfo, param superPodParam) {
	for _, npu := range param.server.NpuMap {
		if npu == nil || npu.LevelList == nil || len(npu.LevelList) == 0 {
			continue
		}
		param.npu = npu
		getEidInfoOrIpFromPorts(npuEidMap, param)
	}
}

func getServerNpuEidMapInfo(npuEidMap map[string]algo.NpuInfo, param superPodParam) {
	for _, server := range param.rack.ServerMap {
		if server == nil || server.NpuMap == nil || len(server.NpuMap) == 0 {
			continue
		}
		param.server = server
		getNpuEidMapInfo(npuEidMap, param)
	}
}

/* 获取super-pod.json中npu ports中目标协议类型出框port与npu映射关系 */
func getNpuEidMapOutOfRack(rackInfo map[string]*RackInfo, protocol string, superPodId string) map[string]algo.NpuInfo {
	npuEidMap := EidNpuMap{
		Map: make(map[string]algo.NpuInfo),
	}
	if rackInfo == nil || len(rackInfo) == 0 {
		hwlog.RunLog.Error("[CONTROLLER]empty rack info")
		return nil
	}
	for _, rack := range rackInfo {
		if rack == nil || rack.ServerMap == nil || len(rack.ServerMap) == 0 {
			continue
		}
		param := superPodParam{}
		param.rack = rack
		param.protocol = protocol
		param.superPodId = superPodId
		getServerNpuEidMapInfo(npuEidMap.Map, param)
	}
	if len(npuEidMap.Map) == 0 {
		hwlog.RunLog.Error("[CONTROLLER]empty npu ports info")
		return nil
	}
	return npuEidMap.Map
}

// SetCallAlgorithmParamInfo 设置算法参数
func SetCallAlgorithmParamInfo(superPodId int, superPodFilePath string,
	callAlgorithmParam map[string]interface{}) error {
	if callAlgorithmParam == nil {
		return errors.New("callAlgorithmParam is nullptr")
	}

	superPodFile := fmt.Sprintf("super-pod-%d.json", superPodId)
	superPodFile = filepath.Join(superPodFilePath, superPodFile)
	superPodFile = filepath.Clean(superPodFile)

	if !loopWaitFile(superPodFile, superPodFilePath) {
		return errors.New("loop wait failed")
	}
	superPodInfo := readConfigMap(superPodFile)
	if superPodInfo == nil {
		return errors.New("super pod info is nil")
	}

	if superPodInfo.Version != DiagVersionA5 && superPodInfo.Version != DiagVersionA3 &&
		superPodInfo.Version != DiagVersionServer {
		return fmt.Errorf("get %s version info error", superPodFile)
	}
	if superPodInfo.Version == DiagVersionServer {
		callAlgorithmParam[NpuType] = DiagVersionA5
		callAlgorithmParam["pingObjType"] = algo.EidType
		return nil
	}
	callAlgorithmParam[NpuType] = superPodInfo.Version
	if superPodInfo.Version == DiagVersionA5 {
		if checkIfNew1D(superPodInfo.RackMap) {
			callAlgorithmParam["pingObjType"] = algo.EidType
		} else {
			callAlgorithmParam["pingObjType"] = algo.IpType
		}
		return nil
	}
	/* A3 */
	callAlgorithmParam["pingObjType"] = 1

	// A3网络结构设置nodeName与serverId的映射
	return getWorKMapping(callAlgorithmParam, superPodInfo)
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

// CheckDiffConfig 获取超节点目录下cathelper.conf配置文件内容
func CheckDiffConfig(superPodFilePath string) map[string]interface{} {
	confFilePath := filepath.Join(superPodFilePath, configFile)
	var fileContent []byte
	var err error
	for retryCount := 0; retryCount < maxRetryTime &&
		!controllerflags.IsControllerExited.GetState(); retryCount++ {
		fileContent, err = fileutils.ReadLimitBytes(confFilePath, constants.Size10M)
		if err == nil {
			break
		}
		hwlog.RunLog.Warnf("[NETFAULT ALGO]Opening file:%v", err)
		time.Sleep(time.Duration(1) * time.Second)
	}
	if fileContent == nil {
		hwlog.RunLog.Error(superPodFilePath, "config file read failed!")
		return nil
	}

	// 定义一个map来保存解析结果
	targetKeys := []string{"networkType", "pingType", "pingTimes", "pingInterval", "suppressedPeriod", "period"}
	callAlgorithmParam := ReadConfigFromFile(fileContent, targetKeys)

	return callAlgorithmParam
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
	file := filepath.Join(superPod, "rack-"+rackId, "topo_"+serverId+".json")
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

/* 获取超级点内探测Npu与EID映射关系 */
func handleA5NpuMapInfo(superPodInfo *SuperPodInfo, superPodPath string) (map[string]algo.NpuInfo, bool) {
	npuInfoMap := make(map[string]algo.NpuInfo)
	var npuNetplaneInfo map[string][]string
	/* 判断是否是1D还是2D */
	typeStr := getNetWorkType(superPodPath, superPodInfo)
	if typeStr == "1D" || typeStr == "2D" {
		npuEidMapOutRack := getNpuEidMapOutOfRack(superPodInfo.RackMap, "UBC", superPodInfo.SuperPodID)
		_, _, npuEidMapFromTopo := GetA5CurSuperPod1D2DNpuInfo(superPodPath, superPodInfo)
		if npuEidMapFromTopo == nil ||
			(len(npuEidMapFromTopo) == 0 && len(npuNetplaneInfo) == 0) {
			return npuInfoMap, false
		}
		npuInfoMap = mergeNpuEidMap(npuEidMapOutRack, npuEidMapFromTopo)
	} else {
		return npuInfoMap, false
	}
	if npuInfoMap == nil {
		return npuInfoMap, false
	}
	return npuInfoMap, true
}

func getReasoningServerSuperPodNpuMap(superPodInfo *SuperPodInfo) map[string]algo.NpuInfo {
	if superPodInfo == nil {
		hwlog.RunLog.Error("[NETFAULT ALGO]invalid or empty superPodInfo")
		return nil
	}
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
		hwlog.RunLog.Error("[NETFAULT ALGO]error rack inner server numbers")
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
			if npu == nil || len(npu.LevelList) == 0 {
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
	key := fmt.Sprintf("%s-%s", param.rack.RackID, param.server.ServerIndex)
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
				npuMap[levelInfo.RankAddrList[i].Addr] = fmt.Sprintf("%s_%s_%s", param.superPodId,
					param.rack.RackID, param.server.ServerIndex)
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
		if server == nil || len(server.NpuMap) == 0 {
			continue
		}
		for _, npu := range server.NpuMap {
			if npu == nil || len(npu.LevelList) == 0 {
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

	if len(superPodInfo.RackMap) == 0 {
		hwlog.RunLog.Errorf("[NETFAULT ALGO]super pod %s empty rack info", superPodInfo.SuperPodID)
		return nil, nil
	}
	npuRoceInfoMap := make(map[string]string)
	npuNetPlanePaths := make(map[string][]string)
	for _, rack := range superPodInfo.RackMap {
		if rack == nil || len(rack.ServerMap) == 0 {
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
	ip := fmt.Sprintf("%d.%d.%d.%d", ip1, ip2, ip3, ip4)
	npuMap[ip] = algo.NpuInfo{
		NpuNumber:    npuId,
		SuperPodName: fmt.Sprintf("SuperPod-%d", superPodId),
		SlotName:     fmt.Sprintf("NSlot-%d", slotId),
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
				RackName:     fmt.Sprintf("Rack-%d", rackId),
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
		if len(npuNetPlanePaths) == 0 || npuRoceMap == nil {
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

func getReasoningServerNpuInfo(npuEidMap map[string]algo.NpuInfo,
	serverIndex int, npu *NpuInfo, serverId, superPodId string) {
	if npu == nil {
		return
	}
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
				SlotName:     fmt.Sprintf("NSlot-%d", slotId),
				NetPlaneId:   fmt.Sprintf("netplane_%s", levelInfo.RankAddrList[i].PlaneId),
				SuperPodName: fmt.Sprintf("SuperPod-%s", superPodId),
				OsName:       serverId}
			if _, exist := npuEidMap[levelInfo.RankAddrList[i].Addr]; !exist && npuEidMap != nil {
				npuEidMap[levelInfo.RankAddrList[i].Addr] = npuInfo
			}
		}
	}
}
