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

// Package policy is used for processing superpod infromation
package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/netfault/algo"
)

const (
	// PrivilegeMode 用户权限模式
	PrivilegeMode    = 0600
	pkgSize          = 28
	basePortConstant = ":0"
	algoInputObjNums = 6
)

// PingInfo 用户权限模式
type PingInfo struct {
	// SrcType src type
	SrcType int `json:"srcType"`
	// DstType dst type
	DstType int `json:"dstType"`
	// PktSize pkt size
	PktSize int `json:"pktSize"`
	// SrcCardPhyId src card physical id
	SrcCardPhyId int `json:"srcCardPhyId"`
	// SrcIp src ip
	SrcIp string `json:"srcAddr"`
	// DstIp dst ip
	DstIp string `json:"dstAddr"`
	// DstCardPhyId
	DstCardPhyId int `json:"dstCardPhyId"`
}

/* 超节点内探测任务写文件 */
func writeServerIdPingList(resPingList []PingInfo, fileName string, superPodPath string) error {
	jsonMap := make(map[string]any)
	jsonMap["pingList"] = resPingList
	jsonMapArray := make([]map[string]any, 0)
	jsonMapArray = append(jsonMapArray, jsonMap)

	jsonStr, err := json.MarshalIndent(jsonMapArray, "", "  ")
	if err != nil {
		hwlog.RunLog.Errorf("json marshal fail, err: %v", err)
		return err
	}

	serverFilePath := superPodPath + "/" + fileName
	file, err := os.Create(serverFilePath)
	if err != nil {
		hwlog.RunLog.Errorf("create file err: %v", err)
		return err
	}

	defer file.Close()
	//修改文件权限
	if err := file.Chmod(PrivilegeMode); err != nil {
		hwlog.RunLog.Errorf("chmod file err, err: %v", err)
		return err
	}

	_, err = file.Write(jsonStr)
	if err != nil {
		hwlog.RunLog.Errorf("write string to file err: %v", err)
		return err
	}

	return nil
}

/* 从全部探测任务中当前超节点内os级别探测任务 */
func getA5ServerLevel1D2DPingList(allPingList []interface{}, npuMap map[string]algo.NpuInfo,
	serverInfo *ServerInfo, isReasoningServer int) []PingInfo {
	resPingList := make([]PingInfo, 0)
	/* 适配1D，从superpodInfo中根据server下有哪些npuId去匹配pingList中srcCardId来生成ping_list文件 */
	for _, pingUnit := range allPingList {
		pingItem, ok := pingUnit.(map[string]interface{})
		if !ok {
			hwlog.RunLog.Errorf("transfer ping list item fail: %v", pingUnit)
			continue
		}

		srcAddrStr, srcOk := pingItem["srcAddr"].(string)
		dstAddrStr, dstOk := pingItem["dstAddr"].(string)
		srcCardId, srcCardOk := pingItem["srcCardPhyId"].(int)
		dstCardId, dstCardOk := pingItem["dstCardPhyId"].(int)
		if !srcOk || !dstOk || !srcCardOk || !dstCardOk {
			hwlog.RunLog.Errorf("transfer ping list item fail: %v", pingUnit)
			continue
		}
		if isReasoningServer != 0 {
			srcCardId = srcCardId % isReasoningServer
			dstCardId = dstCardId % isReasoningServer
		}
		/* npuId在rack下唯一 */
		if value, exist := npuMap[srcAddrStr]; exist && value.OsName == serverInfo.ServerIndex {
			resPingList = append(resPingList, PingInfo{
				SrcIp:        srcAddrStr,
				DstIp:        dstAddrStr,
				SrcType:      algo.EidType,
				DstType:      algo.EidType,
				PktSize:      pkgSize,
				SrcCardPhyId: srcCardId,
				DstCardPhyId: dstCardId})
		}
	}
	return resPingList
}

func siftFromPinglist(serverInfo *ServerInfo, superPodPingList map[string]interface{},
	superPodPath string, npuEidMap map[string]algo.NpuInfo, superPodInfo *SuperPodInfo) {
	if superPodInfo == nil || serverInfo.NpuMap == nil ||
		len(superPodPingList) == 0 || superPodInfo == nil || npuEidMap == nil {
		hwlog.RunLog.Error("invalid super pod info or ping list!")
		return
	}
	allPingList, ok := superPodPingList["pingList"].([]interface{})
	if !ok {
		hwlog.RunLog.Error("get pingList fail!")
		return
	}
	resPingList := make([]PingInfo, 0)
	/* 1D\2D\推理服务器 */
	if superPodInfo.Version == DiagVersionServer {
		/* 取出每台推理服务器有多少张卡 */
		var cardPerServer int
		num, err := fmt.Sscanf(DiagVersionServer, "800I-SuperPod-A5-%d", &cardPerServer)
		if err != nil || num != 1 {
			hwlog.RunLog.Error("Unexpected format of server name!")
			return
		}
		resPingList = getA5ServerLevel1D2DPingList(allPingList, npuEidMap, serverInfo, cardPerServer)
	} else {
		if value := getNetWorkType(superPodPath, serverInfo); value == "1D" || value == "2D" {
			resPingList = getA5ServerLevel1D2DPingList(allPingList, npuEidMap, serverInfo, 0)
		} else {
			hwlog.RunLog.Error("Unexpected situation of rack level topo type!")
			return
		}
	}
	if !isPureNumber(serverInfo.ServerIndex) {
		hwlog.RunLog.Errorf("error server index string:%s", serverInfo.ServerIndex)
		return
	}
	isOk := writeServerIdPingList(resPingList,
		fmt.Sprintf("ping_list_%s.json", serverInfo.ServerIndex), superPodPath)
	if isOk != nil {
		hwlog.RunLog.Error("writeServerIdPingList fail!")
	}
}

func handlePingList(allPingList []any, srcIp string, phyIdStr string) []PingInfo {
	var newPingListRes = make([]PingInfo, 0)
	for _, item := range allPingList {
		pingItem, pingOK := item.(map[string]any)
		if !pingOK {
			hwlog.RunLog.Errorf("get npu ping item fail")
			return newPingListRes
		}

		srcAddrStr, srcOK := pingItem["srcAddr"].(string)
		if !srcOK {
			hwlog.RunLog.Errorf("get srcAddr fail, not exist")
			continue
		}

		if srcAddrStr == srcIp {
			srcIpTmp := srcIp
			srcIpTmp = strings.ReplaceAll(srcIpTmp, basePortConstant, "")
			dstAddrStr, okConvert := pingItem["dstAddr"].(string)
			if !okConvert {
				hwlog.RunLog.Errorf("get dstAddr fail, not exist")
				continue
			}
			dstIp := strings.ReplaceAll(dstAddrStr, basePortConstant, "")
			phyId, err := strconv.Atoi(phyIdStr)
			if err != nil {
				hwlog.RunLog.Errorf("strconv.Atoi(%s) err: %v", phyIdStr, err)
				return nil
			}
			newPingListRes = append(newPingListRes,
				PingInfo{SrcIp: srcIpTmp, DstIp: dstIp, SrcType: algo.IpType, DstType: algo.IpType, PktSize: pkgSize, SrcCardPhyId: phyId})
		}
	}
	return newPingListRes
}

/* 超节点内探测任务划分为os级别写文件 */
func siftFromConfigMap(superPodInfo *SuperPodInfo,
	superPodPingList map[string]interface{},
	curSuperPodPath string) bool {
	if superPodInfo == nil || len(superPodInfo.RackMap) == 0 {
		hwlog.RunLog.Error("invalid super pod info!")
		return false
	}
	id, err := strconv.Atoi(superPodInfo.SuperPodID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return false
	}
	success, npuEidMap := GetTargetSuperPodNpuMap(curSuperPodPath, id)
	if !success {
		hwlog.RunLog.Error("get npu eid map failed!")
		return false
	}
	for rackID, rackInfo := range superPodInfo.RackMap {
		if rackInfo == nil || len(rackInfo.ServerMap) == 0 {
			hwlog.RunLog.Errorf("invalid rack info: %s", rackID)
			return false
		}
		for serverIdStr, serverInfo := range rackInfo.ServerMap {
			_, err := strconv.Atoi(serverIdStr)
			if err != nil {
				hwlog.RunLog.Error("the server id format is invalid")
				return false
			}
			if serverInfo == nil {
				hwlog.RunLog.Errorf("server info is empty of server id %s", serverIdStr)
				return false
			}
			if len(serverInfo.NpuMap) == 0 {
				hwlog.RunLog.Errorf("get npu map failed!")
				return false
			}
			siftFromPinglist(serverInfo, superPodPingList, curSuperPodPath, npuEidMap, superPodInfo)
		}
	}
	return true
}

// GenSuperPodServersPingList 生成超节点内探测任务csv文件
func GenSuperPodServersPingList(superPodPath string, detectObj *algo.NetDetect) bool {
	if detectObj == nil {
		hwlog.RunLog.Error("[NETFAULT ALGO]invalid nil detectObj")
		return false
	}
	superPodPath = filepath.Clean(superPodPath)
	/* get config map info and pingList */
	superPodInfo, superPodPingList := getCurrentSuperPodInfo(superPodPath, detectObj)
	if superPodPingList == nil || superPodInfo == nil {
		return false
	}
	/* 将当前超节点pingList拆分成每个serverId pingList */
	return siftFromConfigMapInterface(superPodInfo, superPodPingList, superPodPath)
}

// 生成超节点间探测任务csv文件
func GenRoceSuperPodLevelPingList(superPodRocePath string,
	detectObj *algo.NetDetect,
	npuLinkPaths map[string]interface{},
	npuMap map[string]algo.NpuInfo) bool {
	// 算法入参
	algoPingListInput := make(map[string]interface{}, algoInputObjNums)
	algoPingListInput["npu_superpod"] = npuLinkPaths
	algoPingListInput["npu_npu"] = []string{}
	algoPingListInput["npu_netplane"] = map[string]interface{}{}
	jsonPingList := detectObj.GenPingStrategy(algoPingListInput)
	if jsonPingList == nil {
		return false
	}
	// 根据npu与port address映射关系拆分成拨测ping_list_superPodId_serverId.csv
	return siftRoceTaskFromNpuMapInterface(superPodRocePath, npuMap, jsonPingList)
}

func writeRocePingListOsInfo(path string, jsonStr []byte) {
	file, err := os.Create(path)
	if err != nil {
		hwlog.RunLog.Errorf("create file err: %v", err)
		return
	}
	defer file.Close()
	// 修改文件权限
	if err := file.Chmod(PrivilegeMode); err != nil {
		hwlog.RunLog.Errorf("chmod file err: %v", err)
		return
	}
	_, err = file.Write(jsonStr)
	if err != nil {
		hwlog.RunLog.Errorf("write string to file err: %v", err)
		return
	}
}

func writeRocePingList(rocePingList map[string][]PingInfo, rocePath string, info map[string][]string) bool {
	// roce ping list 信息
	jsonStr, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		hwlog.RunLog.Errorf("transfer %v failed!", info)
	} else {
		writeRocePingListOsInfo(filepath.Join(rocePath, "ping_list_range.json"), jsonStr)
	}
	if len(rocePingList) == 0 {
		hwlog.RunLog.Error("unmarched any ping task from ping list!")
		return false
	}
	count := 0
	for k, v := range rocePingList {
		fileName := fmt.Sprintf("ping_list_%s.json", k)
		err := writeServerIdPingList(v, fileName, rocePath)
		if err != nil {
			continue
		}
		count++
	}
	if count == 0 {
		hwlog.RunLog.Error("write ping list json file failed!")
		return false
	}
	return true
}

/* 生成超节点间探测任务：根据npu与Eid或者ip关系筛选出跨超节点拨测任务 */
func siftRoceTaskFromNpuMapInterface(rocePath string,
	npuMap map[string]algo.NpuInfo,
	superPodPingList map[string]interface{}) bool {
	// 遍历每一个pingList，根据arcAddr在npuMap中找对应的superPodId-serverId
	if npuMap == nil || superPodPingList == nil {
		hwlog.RunLog.Errorf("error npu map:%v, ping list:%v", npuMap, superPodPingList)
		return false
	}
	// key:superPodId-serverId value:pingTask
	rocePingList := make(map[string][]PingInfo)
	allPingList, ok := superPodPingList["pingList"].([]interface{})
	if !ok {
		hwlog.RunLog.Error("get pingList failed!")
		return false
	}
	// 记录生成那些超节点哪些os的ping_list
	uniqueSuperPodOs := make(map[string]bool)
	superPodPingListInfo := make(map[string][]string)
	for _, pingUnit := range allPingList {
		pingItem, pingOk := pingUnit.(map[string]interface{})
		if !pingOk {
			hwlog.RunLog.Errorf("transfer ping list item fail:%v", pingUnit)
			continue
		}
		srcAddrStr, srcOk := pingItem["srcAddr"].(string)
		dstAddrStr, dstOk := pingItem["dstAddr"].(string)
		srcCardId, srcCardOk := pingItem["srcCardPhyId"].(int)
		dstCardId, dstCardOk := pingItem["dstCardPhyId"].(int)
		if !srcOk || !dstOk || !srcCardOk || !dstCardOk {
			hwlog.RunLog.Errorf("transfer ping list item fail:%v", pingUnit)
			continue
		}
		if _, exist := npuMap[srcAddrStr]; !exist || len(npuMap[srcAddrStr].SuperPodName) <= len("SuperPod-") {
			continue
		}
		superPodId := string(npuMap[srcAddrStr].SuperPodName[len("SuperPod-"):])
		key := superPodId + "_" + npuMap[srcAddrStr].OsName
		if _, exist := uniqueSuperPodOs[key]; !exist {
			uniqueSuperPodOs[key] = true
			if _, existKey := superPodPingListInfo[superPodId]; !existKey {
				superPodPingListInfo[superPodId] = make([]string, 0)
			}
			superPodPingListInfo[superPodId] = append(superPodPingListInfo[superPodId], npuMap[srcAddrStr].OsName)
		}
		if _, exist := rocePingList[key]; !exist {
			rocePingList[key] = make([]PingInfo, 0)
		}
		rocePingList[key] = append(rocePingList[key], PingInfo{
			SrcIp:        srcAddrStr,
			DstIp:        dstAddrStr,
			SrcType:      algo.IpType,
			DstType:      algo.IpType,
			PktSize:      pkgSize,
			SrcCardPhyId: srcCardId,
			DstCardPhyId: dstCardId})
	}
	return writeRocePingList(rocePingList, rocePath, superPodPingListInfo)
}

/* 超节点探测任务划分为os级别任务文件 */
func siftFromConfigMapInterface(superPodInfo *SuperPodInfo, superPodPingList map[string]any, curSuperPodPath string) bool {
	switch superPodInfo.Version {
	case DiagVersionA3:
		return siftFromConfigMapA3(superPodInfo, superPodPingList, curSuperPodPath)
	default:
		hwlog.RunLog.Errorf("unknown detection version!")
		return false
	}
}

func siftFromConfigMapA3(configMap *SuperPodInfo, superPodPingList map[string]any, curSuperPodPath string) bool {
	if configMap == nil || len(configMap.NodeDeviceMap) == 0 {
		hwlog.RunLog.Errorf("get DodeDeviceMap map failed")
		return false
	}
	for _, workInfo := range configMap.NodeDeviceMap {
		if workInfo == nil {
			hwlog.RunLog.Errorf("get target work failed")
			return false
		}
		if len(workInfo.DeviceMap) == 0 {
			hwlog.RunLog.Errorf("get device map failed")
			return false
		}
		if len(workInfo.ServerID) == 0 {
			hwlog.RunLog.Errorf("get server id failed")
			return false
		}
		serverID, err := strconv.Atoi(workInfo.ServerID)
		if err != nil {
			hwlog.RunLog.Errorf("workId Atoi err: %s", err)
			return false
		}
		siftFromPinglistA3(workInfo.DeviceMap, superPodPingList, serverID, curSuperPodPath)
	}
	return true
}

func siftFromPinglistA3(NodeDevMap map[string]string, superPodPingList map[string]any, workId int, superPodPath string) {
	if NodeDevMap == nil || superPodPingList == nil {
		return
	}
	allPingList, ok := superPodPingList["pingList"].([]any)
	if !ok {
		hwlog.RunLog.Errorf("get pingList failed")
		return
	}
	var resPingList []PingInfo = nil
	for phyID, superDeviceId := range NodeDevMap {
		if len(superDeviceId) == 0 {
			hwlog.RunLog.Errorf("get superDeviceId failed")
			return
		}
		resPingListRet := handlePingList(allPingList, superDeviceId, phyID)
		if resPingListRet == nil || len(resPingListRet) == 0 {
			continue
		}
		resPingList = append(resPingList, resPingListRet...)
	}
	isOk := writeServerIdPingList(resPingList, fmt.Sprintf("ping_list_%d.json", workId), superPodPath)
	if isOk != nil {
		hwlog.RunLog.Errorf("writeServerIdPingList fail")
	}
}
