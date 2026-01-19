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
	"fmt"
	"strconv"
	"strings"

	"ascend-common/common-utils/hwlog"
)

const (
	// IpPartNum ： ip位数
	IpPartNum = 4
	// PartsLimit 提取NPU ip分割部分数
	PartsLimit = 4
	// PartsByHashLimit 以#为分隔符提取rack  slot npu信息分隔符
	PartsByHashLimit = 3
	// NpuStringLimit ： NPU信息提取阶段后部分数
	NpuStringLimit = 2
	// NpuStringPartOne ： NPU信息提取阶段后第二部分 下标为1
	NpuStringPartOne = 1
)

func storeNpuBoard2DFullMeshIp(npuFullMeshInfo map[int]string, portInfo *VnicInfo, npuId int) {
	if npuFullMeshInfo == nil {
		return
	}
	if portInfo == nil {
		hwlog.RunLog.Error("port info is empty")
		return
	}
	npuFullMeshInfo[npuId] = portInfo.VnicIp
}

func storeNpuNetPlaneLink(rackId int, slotId int, npuPhyId int, portInfo *VnicInfo,
	npuNetplaneInfo map[string][]string) {
	if portInfo == nil || len(portInfo.PortId) == 0 || len(portInfo.VnicIp) == 0 {
		hwlog.RunLog.Error("port info is invalid")
		return
	}
	portId := portInfo.PortId
	portIp := portInfo.VnicIp
	/* net plane started by 0, port id 1 to net plane 0 */
	npuNetPlaneStr := fmt.Sprintf("NA.L2-LogicPort:0#Rack-%d.NA:0#Rack-%d.NSlot-%d:0#NPU-%d.%s:0",
		rackId, rackId, slotId, npuPhyId, portIp)

	switch portId {
	case npuFirstPort:
		if npuNetplaneInfo != nil {
			npuNetplaneInfo[npuFirstPort] = append(npuNetplaneInfo[npuFirstPort], npuNetPlaneStr)
		}
		break
	case npuSecondPort:
		if npuNetplaneInfo != nil {
			npuNetplaneInfo[npuSecondPort] = append(npuNetplaneInfo[npuSecondPort], npuNetPlaneStr)
		}
		break
	case npuThirdPort:
		if npuNetplaneInfo != nil {
			npuNetplaneInfo[npuThirdPort] = append(npuNetplaneInfo[npuThirdPort], npuNetPlaneStr)
		}
		break
	case npuFourthPort:
		if npuNetplaneInfo != nil {
			npuNetplaneInfo[npuFourthPort] = append(npuNetplaneInfo[npuFourthPort], npuNetPlaneStr)
		}

		break
	default:
		hwlog.RunLog.Warnf("unknown npu port net plane: %s", portId)
		return
	}
}

func check2DFullMeshLinkNotExist(array []string, aToB string, bToA string) bool {
	for _, item := range array {
		if item == aToB || item == bToA {
			return false
		}
	}
	return true
}

func storeRack2DFullMeshLink(npu2DXFullMesh map[int]string,
	npu2DYFullMesh map[int]string,
	serverNpusMap map[string]int) []string {
	var curRackTotalNpus int = 0
	for _, curServerNpus := range serverNpusMap {
		curRackTotalNpus += curServerNpus
	}
	if len(serverNpusMap) <= 0 || curRackTotalNpus <= 0 {
		hwlog.RunLog.Error("Invalid rack info!!!")
		return nil
	}
	fullMeshArrayAll := make([]string, 0)
	resInside := fullMeshInsideBoard(npu2DXFullMesh)
	if resInside != nil && len(resInside) != 0 {
		fullMeshArrayAll = append(fullMeshArrayAll, resInside...)
	}
	/* full mesh Cross-board coaxial */
	resCross := fullMeshCrossBoard(npu2DYFullMesh)
	if resCross != nil || len(resCross) != 0 {
		fullMeshArrayAll = append(fullMeshArrayAll, resCross...)
	}

	return fullMeshArrayAll
}

func spliceInsideBoardNpuLink(npu2DXFullMesh map[int]string,
	boardXId int) []string {
	fullMeshArrayRet := make([]string, 0)
	for i := 0; i < perBoardNpus; i++ {
		srcId := boardXId*perBoardNpus + i
		/* 不存在说明卡缺失或损坏， topo更新 */
		ipSrc, exist := npu2DXFullMesh[srcId]
		if !exist {
			continue
		}
		for j := 1; j < perBoardNpus; j++ {
			var dstId int
			if (srcId + j) >= ((boardXId + 1) * perBoardNpus) {
				dstId = (srcId+j)%((boardXId+1)*perBoardNpus) + boardXId*perBoardNpus
			} else {
				dstId = (srcId + j) % ((boardXId + 1) * perBoardNpus)
			}
			ipDst, exist := npu2DXFullMesh[dstId]
			if !exist {
				continue
			}
			aToB := fmt.Sprintf("%s:0#%s:0", ipSrc, ipDst)
			bToA := fmt.Sprintf("%s:0#%s:0", ipDst, ipSrc)
			/* check if ping exist */
			if check2DFullMeshLinkNotExist(fullMeshArrayRet, aToB, bToA) {
				fullMeshArrayRet = append(fullMeshArrayRet, aToB)
			}
		}
	}
	return fullMeshArrayRet
}

func fullMeshInsideBoard(npu2DXFullMesh map[int]string) []string {
	fullMeshArrayResAll := make([]string, 0)
	/* full mesh inside the board */
	for boardXId := 0; boardXId < perRackMaxNpuBoard; boardXId++ {
		fullMeshArrayRes := spliceInsideBoardNpuLink(npu2DXFullMesh, boardXId)
		if fullMeshArrayRes == nil || len(fullMeshArrayRes) == 0 {
			continue
		}
		fullMeshArrayResAll = append(fullMeshArrayResAll, fullMeshArrayRes...)
	}
	return fullMeshArrayResAll
}

func spliceCrossBoardNpuLink(npu2DYFullMesh map[int]string,
	boardInsideNpuId int) []string {
	/* boardInsideNpuId is logic id in per npu board */
	fullMeshArrayAll := make([]string, 0)
	for i := 0; i < perRackMaxNpuBoard; i++ {
		/* the i is current logic npu board id*/
		srcId := boardInsideNpuId + i*perBoardNpus
		ipSrc, exist := npu2DYFullMesh[srcId]
		if !exist {
			continue
		}
		/* j */
		for j := 1; j < perRackMaxNpuBoard; j++ {
			dstId := (perBoardNpus*j + srcId) % (perBoardNpus * perRackMaxNpuBoard)
			ipDst, exist := npu2DYFullMesh[dstId]
			if !exist {
				continue
			}
			aToB := fmt.Sprintf("%s:0#%s:0", ipSrc, ipDst)
			bToA := fmt.Sprintf("%s:0#%s:0", ipDst, ipSrc)
			/* check if ping exist */
			if check2DFullMeshLinkNotExist(fullMeshArrayAll, aToB, bToA) {
				fullMeshArrayAll = append(fullMeshArrayAll, aToB)
			}
		}
	}
	return fullMeshArrayAll
}

func fullMeshCrossBoard(npu2DYFullMesh map[int]string) []string {
	fullMeshArrayAll := make([]string, 0)
	for boardInsideNpuId := 0; boardInsideNpuId < perBoardNpus; boardInsideNpuId++ {
		fullMeshArrayRes := spliceCrossBoardNpuLink(npu2DYFullMesh, boardInsideNpuId)
		if fullMeshArrayRes == nil || len(fullMeshArrayRes) == 0 {
			continue
		}
		fullMeshArrayAll = append(fullMeshArrayAll, fullMeshArrayRes...)
	}
	return fullMeshArrayAll
}

func getCurNpusInfo(rackId int,
	npuInfo *NpuInfo,
	npu2DXFullMeshIps map[int]string,
	npu2DYFullMeshIps map[int]string,
	npuNetplaneInfo map[string][]string) bool {
	/* npu physical id */
	if npuInfo == nil {
		hwlog.RunLog.Error("npu info is nil")
		return false
	}
	npuId, err := strconv.Atoi(npuInfo.PhyId)
	if err != nil {
		hwlog.RunLog.Errorf("npu physical id formatted error")
		return false
	}
	if len(npuInfo.VnicIpMap) == 0 {
		hwlog.RunLog.Error("the VnicIpMap of npu info is empty")
		return false
	}
	/* calculate slot id */
	slotId := npuId / perBoardNpus
	/* port ip map */
	for portId, portInfo := range npuInfo.VnicIpMap {
		if portInfo == nil {
			hwlog.RunLog.Errorf("the portInfo of npu %s VnicIpMap is empty", npuInfo.PhyId)
			return false
		}
		/* check port ip id */
		switch portId {
		case npuInsideBoardPort:
			storeNpuBoard2DFullMeshIp(npu2DXFullMeshIps, portInfo, npuId)
			break
		case npuAcrossBoardPort:
			storeNpuBoard2DFullMeshIp(npu2DYFullMeshIps, portInfo, npuId)
			break
		default: /* net plane port ip */
			storeNpuNetPlaneLink(rackId, slotId, npuId, portInfo, npuNetplaneInfo)
			break
		}
	}
	return true
}

func getCurOsInfo(rackId int,
	serverInfo *ServerInfo,
	npu2DXFullMeshIps map[int]string,
	npu2DYFullMeshIps map[int]string,
	npuNetplaneInfo map[string][]string) (bool, int) {
	if serverInfo == nil {
		hwlog.RunLog.Error("server info is empty")
		return false, -1
	}
	/* npu map */
	perServerNpus := len(serverInfo.NpuMap)
	if perServerNpus == 0 {
		hwlog.RunLog.Errorf("the NpuMap of node %s is empty", serverInfo.NodeName)
		return false, -1
	}
	/* npu info */
	for _, npuInfo := range serverInfo.NpuMap {
		if !getCurNpusInfo(rackId,
			npuInfo,
			npu2DXFullMeshIps,
			npu2DYFullMeshIps,
			npuNetplaneInfo) {
			return false, -1
		}
	}
	return true, perServerNpus
}

func getCurRackInfo(npuNetplaneInfo map[string][]string, rackInfo *RackInfo) []string {
	if rackInfo == nil {
		hwlog.RunLog.Error("the rack info is empty")
		return nil
	}
	/* store current npu 2D full mesh ip */
	npu2DXFullMeshIps := make(map[int]string)
	npu2DYFullMeshIps := make(map[int]string)
	/* rack id */
	rackId, err := strconv.Atoi(rackInfo.RackID)
	if err != nil {
		hwlog.RunLog.Errorf("rackId formatted error")
		return nil
	}
	/* server Map */
	serverNpusMap := make(map[string]int)
	for serverId, osMap := range rackInfo.ServerMap {
		flag, perServerNpuNums := getCurOsInfo(rackId, osMap, npu2DXFullMeshIps, npu2DYFullMeshIps, npuNetplaneInfo)
		if !flag {
			return nil
		}
		serverNpusMap[serverId] = perServerNpuNums
	}
	resStore := storeRack2DFullMeshLink(npu2DXFullMeshIps, npu2DYFullMeshIps, serverNpusMap)
	if resStore == nil || len(resStore) == 0 {
		return nil
	}
	return resStore
}

func parseRackMap(racksInfo map[string]*RackInfo) ([]string, map[string][]string) {
	if racksInfo == nil {
		hwlog.RunLog.Errorf("rackMap is nil")
		return nil, nil
	}
	/* store all rack npu-npu net 2D full mesh */
	npu2DFullMeshInfo := make([]string, 0)
	/* store all rack npu-net planes(2D 4 net plane), key 1-4(2D) or 1-9(1D), value npuNetPlaneInfos */
	npuNetplaneInfo := make(map[string][]string)

	for rackID, rackInfo := range racksInfo {
		res := getCurRackInfo(npuNetplaneInfo, rackInfo)
		if res == nil || len(res) == 0 {
			hwlog.RunLog.Warnf("rack %s full mesh result is empty", rackID)
			continue
		}
		npu2DFullMeshInfo = append(npu2DFullMeshInfo, res...)
	}

	return npu2DFullMeshInfo, npuNetplaneInfo
}

func getNPUNum(s string) int {
	resRet := -1
	parts := strings.Split(s, "-")
	if len(parts) < NpuStringLimit {
		return resRet
	}
	num := strings.Split(parts[0], "NPU")
	if len(num) < NpuStringLimit {
		return resRet
	}
	res, err := strconv.Atoi(num[NpuStringPartOne])
	if err != nil {
		hwlog.RunLog.Errorf("string into int failed：%v", err)
		return resRet
	}
	resRet = res

	return resRet
}

func removeTail(s string) string {
	parts := strings.Split(s, ":")
	return parts[0]
}
