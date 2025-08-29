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

// GenSuperPodServersPingList 生成超节点内探测任务csv文件
func GenSuperPodServersPingList(superPodPath string, detectObj *algo.NetDetect) bool {
	superPodPath = filepath.Clean(superPodPath)
	/* get config map info and pingList */
	superPodInfo, superPodPingList := getCurrentSuperPodInfo(superPodPath, detectObj)
	if superPodPingList == nil || superPodInfo == nil {
		return false
	}
	/* 将当前超节点pingList拆分成每个serverId pingList */
	return siftFromConfigMapInterface(superPodInfo, superPodPingList, superPodPath)
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
