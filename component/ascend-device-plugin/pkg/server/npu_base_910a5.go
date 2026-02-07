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

// Package server holds the implementation of registration to kubelet, k8s pod resource interface.
package server

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	apiCommon "ascend-common/devmanager/common"
)

var hcclTopoFilePathMap = map[int8]string{
	common.ProductTypeServer:    common.Server8PTopoPath,
	common.ProductType1D:        common.Pod1DTopoPath,
	common.ProductType2D:        common.Pod2DTopoPath,
	common.ProductType16PServer: common.Server16PTopoPath,
	common.ProductType32PServer: common.Server32PTopoPath,
	common.ProductType1PCard:    common.Card1PTopoPath,
	common.ProductType4PCard:    common.Card4PTopoPath,
}

var rankLevelInfoKeyArrMap = map[string][]string{
	common.A5300ICardName: {
		api.LevelInfoTypeIgnore, api.LevelInfoTypeIgnore, api.LevelInfoTypeIgnore, api.LevelInfoTypeRoCE,
	},
	common.A54P300ICardName: {
		api.LevelInfoTypeUB, api.LevelInfoTypeIgnore, api.LevelInfoTypeIgnore, api.LevelInfoTypeRoCE,
	},
}

const (
	size50M        = 50 * 1024 * 1024
	addrTypeEID    = "EID"
	addrTypeIPV4   = "IPV4"
	decimal        = 10
	hexadecimal    = 16
	addrNumsLength = 2
	dieIdMaskNum   = 0x04
	portIdMaskNum  = 0x7F
	rightShiftLen  = 3
)

var npuBase *NpuBase

func init() {
	npuBase = NewNpuBase()
}

type netTypeAndFeIdList struct {
	netType  string
	feIdList []uint
}

// ProductBase for product info in os domain
type ProductBase struct {
	superPodSize   uint32
	superPodID     uint32
	serverIndex    uint32
	chassisID      uint32
	superPodType   uint8
	nodeInternalIP string
	topoFilePath   string
	cardType       string
	topoInfo       *TopoInfo
}

// getID for get the level id in rank table
func (p *ProductBase) getID(level int) string {
	if p == nil {
		hwlog.RunLog.Errorf("product info is empty")
		return ""
	}
	// when return empty, it means that the level is no need exist in the final rank table file
	switch level {
	case api.RankLevel0:
		// standard card
		if p.cardType == common.A5300ICardName || p.cardType == common.A54P300ICardName {
			return p.nodeInternalIP
		}
		// pod
		if !p.isServer() {
			return fmt.Sprintf("%s_%s", strconv.Itoa(int(p.superPodID)), strconv.Itoa(int(p.chassisID)))
		}
		// server super pod
		if p.isSuperServer() {
			return fmt.Sprintf("%s_%s", strconv.Itoa(int(p.superPodID)), strconv.Itoa(int(p.serverIndex)))
		}
		// server not super pod
		return p.nodeInternalIP
	case api.RankLevel1:
		// server not super pod not have level 1
		if p.isServer() && !p.isSuperServer() {
			return ""
		}
		// super pod
		return strconv.Itoa(int(p.superPodID))
	case api.RankLevel2:
		return api.DefaultClusterName
	case api.RankLevel3:
		return api.DefaultClusterName
	default:
		// when return empty, it means that the level is no need exist in the final rank table file
		return ""
	}
}

func (p *ProductBase) isPodScene() bool {
	if p == nil {
		hwlog.RunLog.Errorf("product info is empty")
		return false
	}
	superPodType := p.superPodType
	return superPodType == common.ProductType1D || superPodType == common.ProductType2D
}

func (p *ProductBase) isSuperServer() bool {
	if p == nil {
		hwlog.RunLog.Errorf("product info is empty")
		return false
	}

	if !p.isServer() {
		return false
	}

	if p.superPodType == common.ProductType16PServer || p.superPodType == common.ProductType32PServer {
		return false
	}

	superPodID := p.superPodID
	superPodSize := p.superPodSize
	return int(superPodSize) != common.InvalidSuperPodSize && int(superPodID) != common.InvalidSuperPodID
}

func (p *ProductBase) isServer() bool {
	if p == nil {
		hwlog.RunLog.Error("product info is empty")
		return false
	}
	return !p.isPodScene()
}

func (p *ProductBase) isStandCard() bool {
	if p == nil {
		hwlog.RunLog.Error("product info is empty")
		return false
	}
	return p.cardType == common.A5300ICardName || p.cardType == common.A54P300ICardName
}

// getTopoFileInfo get topo info
func (p *ProductBase) getTopoFileInfo() (*TopoInfo, error) {
	if p.topoInfo != nil {
		hwlog.RunLog.Debugf("topo info is already loaded from file %s", p.topoFilePath)
		return p.topoInfo, nil
	}
	// get topo path
	path, err := p.getTopoPath()
	if err != nil {
		return nil, fmt.Errorf("get topo path failed, err:<%v>", err)
	}
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("check topo file exist failed, path:<%v>; err:<%v>", path, err)
	}
	topoData, err := utils.ReadLimitBytes(path, size50M)
	if err != nil {
		return nil, fmt.Errorf("read topo file failed, path:<%v>; err:<%v>", path, err)
	}
	// check json and unmarshal
	if !json.Valid(topoData) {
		return nil, fmt.Errorf("topo file is not json, path:<%v>", path)
	}
	var topoInfo TopoInfo
	if err := json.Unmarshal(topoData, &topoInfo); err != nil {
		return nil, fmt.Errorf("topo info json unmarshal failed, err:<%v>", err)
	}
	// put in cache
	p.topoFilePath = path
	p.topoInfo = &topoInfo
	return &topoInfo, nil
}

// getTopoPath get topo path by card type and super pod type
func (p *ProductBase) getTopoPath() (string, error) {
	if p.cardType == common.A5300ICardName {
		return hcclTopoFilePathMap[common.ProductType1PCard], nil
	}
	if p.cardType == common.A54P300ICardName {
		return hcclTopoFilePathMap[common.ProductType4PCard], nil
	}
	path, exist := hcclTopoFilePathMap[int8(p.superPodType)]
	if !exist {
		return "", fmt.Errorf("super pod type:<%d> topo path not exist", p.superPodType)
	}
	return path, nil
}

// NpuBase save npu base info
type NpuBase struct {
	productInfo    *ProductBase
	eidPortMap     map[string][]string
	portMapMutex   sync.RWMutex
	urmaDevInfoMap map[int32][]apiCommon.UrmaDeviceInfo
}

// NewNpuBase for new NpuBase instance
func NewNpuBase() *NpuBase {
	return &NpuBase{
		eidPortMap:     make(map[string][]string),
		portMapMutex:   sync.RWMutex{},
		urmaDevInfoMap: make(map[int32][]apiCommon.UrmaDeviceInfo),
	}
}

// SetUrmaDeviceInfoByHdm setting for urma device info
func (n *NpuBase) SetUrmaDeviceInfoByHdm(hdm *HwDevManager, dev *common.NpuDevice) error {
	if dev == nil {
		return errors.New("input parameter dev is nil")
	}
	if _, exist := n.urmaDevInfoMap[dev.PhyID]; exist {
		hwlog.RunLog.Infof("cardID(%d) deviceID(%d) phyID(%d) urma devie info already exist", dev.CardID,
			dev.DeviceID, dev.PhyID)
		return nil
	}

	if hdm == nil {
		return errors.New("input parameter hdm is nil")
	}
	if hdm.manager == nil {
		return errors.New("input parameter hdm.manager is nil")
	}
	dmgr := hdm.manager.GetDmgr()
	if dmgr == nil {
		return errors.New("input parameter dmgr is nil")
	}
	urmaDevInfoAll, err := dmgr.GetUrmaDevEidListAll(dev.CardID, dev.DeviceID)
	if err != nil {
		hwlog.RunLog.Errorf("get urma device info failed, err: %v", err)
		return err
	}
	n.urmaDevInfoMap[dev.PhyID] = urmaDevInfoAll
	return nil
}

func (n *NpuBase) getID(level int) string {
	if n.productInfo == nil {
		hwlog.RunLog.Error("product info is empty")
		return ""
	}
	return n.productInfo.getID(level)
}

func (n *NpuBase) getRankLevelInfoKeyArr() []string {
	if n.productInfo == nil {
		hwlog.RunLog.Error("product info is empty")
		return []string{}
	}
	if arr, ok := rankLevelInfoKeyArrMap[n.productInfo.cardType]; ok {
		return arr
	}
	if n.productInfo.isPodScene() {
		return []string{
			api.LevelInfoTypeUB, api.LevelInfoTypeUB, api.LevelInfoTypeUBG, api.LevelInfoTypeRoCE,
		}
	}
	// server scene with super pod
	if n.productInfo.isSuperServer() {
		return []string{
			api.LevelInfoTypeUB, api.LevelInfoTypeUB, api.LevelInfoTypeUBoE, api.LevelInfoTypeRoCE,
		}
	}
	// server scene without super pod
	if n.productInfo.isServer() {
		return []string{
			api.LevelInfoTypeUB, api.LevelInfoTypeIgnore, api.LevelInfoTypeUBoE, api.LevelInfoTypeRoCE,
		}
	}
	return []string{}
}

// getNetTypeForLevel for get the fabric type in rank table for A5
func (n *NpuBase) getNetTypeForLevel(level int) string {
	if level == api.RankLevel0 {
		return api.NetTypeTopo
	}
	return api.NetTypeCLOS
}

func (n *NpuBase) getNetTypeAndFeIDListByRankLevel(rankLevel int) (string, []uint) {
	if n.productInfo == nil {
		hwlog.RunLog.Warn("product info is nil")
		return "", []uint{}
	}
	if rankLevel < 0 || rankLevel >= api.RankLevelCnt {
		hwlog.RunLog.Errorf("rank level is %d, should be in range [0, %d)", rankLevel, api.RankLevelCnt)
		return "", []uint{}
	}
	switch rankLevel {
	case api.RankLevel0:
		if n.productInfo.isStandCard() {
			return api.LevelInfoTypeUB, []uint{common.UrmaFeId0}
		}
		return api.LevelInfoTypeUB, []uint{common.UrmaFeId1}
	case api.RankLevel1:
		return api.LevelInfoTypeUB, []uint{common.UrmaFeId0}
	case api.RankLevel2:
		if !n.productInfo.isPodScene() {
			return api.LevelInfoTypeUBoE, []uint{common.UrmaFeId8, common.UrmaFeId9}
		}
		return api.LevelInfoTypeUBG, []uint{common.UrmaFeId3}
	case api.RankLevel3:
		return api.LevelInfoTypeRoCE, []uint{}
	default:
		return "", nil
	}
}

func (n *NpuBase) getRandAddrByFuncEntityID(phyID int32, feID uint, netType string, rankLevel int) []api.RankAddrItem {
	urmaDevInfoAll, exist := n.urmaDevInfoMap[phyID]
	if !exist {
		hwlog.RunLog.Errorf("get urma device info failed, phyID(%d) not exist in cache map", phyID)
		return nil
	}

	rankAddrList := make([]api.RankAddrItem, 0)
	for _, devInfo := range urmaDevInfoAll {
		eidList := n.getEidListByFeIDAndRankLevel(feID, &devInfo, rankLevel)
		for i := 0; i < len(eidList); i++ {
			eid := eidList[i].Eid
			eidStr := hex.EncodeToString(eid.Raw[:])
			portList, err := n.GetPortListByEid(phyID, eidStr, rankLevel)
			if err != nil {
				hwlog.RunLog.Warnf("get port list by eid for phyID=%d feID=%d netType=%s rankLevel=%d eid=%s "+
					"failed, err: %v", phyID, feID, netType, rankLevel, eidStr, err)
				continue
			}
			item := n.createRankAddrItem(netType, eid, portList)
			rankAddrList = append(rankAddrList, item)
		}
	}
	return rankAddrList
}

func (n *NpuBase) GetPortListByEid(phyId int32, eid string, rLevel int) ([]string, error) {
	n.portMapMutex.Lock()
	defer n.portMapMutex.Unlock()
	if common.ParamOption.RealCardType != api.Ascend910A5 {
		return nil, fmt.Errorf("get port list by eid error, device type is not A5")
	}
	// hit cache
	eidPortMapKey := getEidPortMapKey(phyId, eid)
	if ports, exist := n.eidPortMap[eidPortMapKey]; exist {
		hwlog.RunLog.Infof("get port list success, hit cache key:<%s>", eidPortMapKey)
		return ports, nil
	}
	// update cache
	return n.getPortsList(phyId, eid, rLevel)
}

func (n *NpuBase) getEidListByFeIDAndRankLevel(feID uint, urmaDevInfo *apiCommon.UrmaDeviceInfo,
	rankLevel int) []apiCommon.UrmaEidInfo {
	if urmaDevInfo == nil {
		return []apiCommon.UrmaEidInfo{}
	}

	eidList := make([]apiCommon.UrmaEidInfo, 0)
	for i := 0; i < int(urmaDevInfo.EidCount); i++ {
		if n.getFeIDByEid(&urmaDevInfo.EidInfos[i].Eid) != feID {
			continue
		}
		if rankLevel == api.RankLevel0 && !n.checkEidIsUsedForD2D(&urmaDevInfo.EidInfos[i].Eid) {
			continue
		}
		eidList = append(eidList, urmaDevInfo.EidInfos[i])
	}
	return eidList
}

func (n *NpuBase) getFeIDByEid(eid *apiCommon.Eid) uint {
	if eid == nil {
		return math.MaxUint
	}
	return uint(binary.BigEndian.Uint64(eid.Raw[0:common.FeIdIndexByte]) >> common.FeIdIndexBit & common.FeIdMask)
}

func (n *NpuBase) checkEidIsUsedForD2D(eid *apiCommon.Eid) bool {
	if eid == nil {
		return false
	}
	const twoBytes = 2
	const usageBitIndex = 11
	const usageBitVal = 0
	const usageBitMask = 0x01
	return (binary.BigEndian.Uint16(eid.Raw[len(eid.Raw)-twoBytes:]) >> usageBitIndex & usageBitMask) == usageBitVal
}

func (n *NpuBase) createRankAddrItem(netType string, eid apiCommon.Eid, ports []string) api.RankAddrItem {
	planeId := api.DefaultRandAddrPlaneID
	if netType == api.LevelInfoTypeUB {
		val := int(eid.Raw[len(eid.Raw)-1])
		dieId := 0
		if val <= common.PhyLimit {
			dieId = ((val - 1) / common.PhyPortNumPerDie) % common.DieNumPerDev
		} else if common.LogicLowerLimit <= val && val <= common.LogicUpperLimit {
			dieId = ((val - common.LogicLowerLimit) / common.LogicPortNumPerDie) % common.DieNumPerDev
		}
		planeId = strconv.Itoa(dieId)
	}

	if netType == api.LevelInfoTypeUBG || netType == api.LevelInfoTypeUB {
		return api.RankAddrItem{
			AddrType: addrTypeEID,
			Addr:     hex.EncodeToString(eid.Raw[:]),
			Ports:    ports,
			PlaneId:  planeId,
		}
	}

	if netType == api.LevelInfoTypeUBoE {
		ipv4Bytes := eid.Raw[len(eid.Raw)-apiCommon.MaxUBCNAByteLen : len(eid.Raw)]
		ipv4Str := fmt.Sprintf("%d.%d.%d.%d", ipv4Bytes[0], ipv4Bytes[1], ipv4Bytes[2], ipv4Bytes[3])
		return api.RankAddrItem{
			AddrType: addrTypeIPV4,
			Addr:     ipv4Str,
			Ports:    ports,
			PlaneId:  planeId,
		}
	}
	return api.RankAddrItem{}
}

func getEidPortMapKey(phyId int32, eid string) string {
	return fmt.Sprintf("%d_%s", phyId, eid)
}

func (n *NpuBase) getPortsList(phyId int32, eid string, rLevel int) ([]string, error) {
	ports := make([]string, 0)
	// one eid cannot have duplicate ports
	portsMap := make(map[string]struct{})
	eidPortMapKey := getEidPortMapKey(phyId, eid)
	// level 2 not need die id
	if rLevel == api.RankLevel2 {
		// get topo path info
		topoInfo, err := n.productInfo.getTopoFileInfo()
		if err != nil {
			return ports, fmt.Errorf("get port list by eid failed: get topo file info err:<%v>", err)
		}
		// save port result
		for _, edge := range topoInfo.EdgeList {
			if edge.NetLayer == rLevel && edge.LocalA == int(phyId) {
				ports = savePortsResult(edge, ports, portsMap)
			}
		}
		if len(ports) == 0 {
			hwlog.RunLog.Warnf("eid<%s> ports info is not found in topo file, edge list count %d",
				eid, len(topoInfo.EdgeList))
		}
		n.eidPortMap[eidPortMapKey] = ports
		return ports, nil
	}

	// full mesh calculate get ports
	dieId, portId, err := getDieIdAndPortId(eid)
	if err != nil {
		return ports, err
	}

	ports = append(ports, fmt.Sprintf("%d/%d", dieId, portId))
	// get topo path info
	topoInfo, err := n.productInfo.getTopoFileInfo()
	if err != nil {
		return ports, fmt.Errorf("get topoInfo failed: err:<%v>", err)
	}

	for _, edge := range topoInfo.EdgeList {
		if edge.NetLayer == rLevel && edge.LocalA == int(phyId) && edge.LinkType == common.Peer2Net {
			ports = savePortsResultByDieId(edge, dieId, ports, portsMap)
		}
	}
	hwlog.RunLog.Infof("get port list success, save cache key:<%s>, value:%v, rLevel:<%d>",
		eidPortMapKey, ports, rLevel)
	n.eidPortMap[eidPortMapKey] = ports
	return ports, nil
}

func getDieIdAndPortId(eid string) (byte, byte, error) {
	if len(eid) < common.DieIDOffset {
		return 0, 0, fmt.Errorf("eid:<%v> len is invalid, which should be greater equal than %d",
			eid, common.DieIDOffset)
	}

	var dieId, portId byte
	dieIdStr := eid[len(eid)-common.DieIDOffset : len(eid)-common.PortIDSuffixLen] // third-to-last character for dieId
	portIdStr := eid[len(eid)-common.PortIDSuffixLen:]                             // last two characters for portId calculation
	dieIdInt, err := strconv.ParseInt(dieIdStr, hexadecimal, 0)
	if err != nil {
		return 0, 0, fmt.Errorf("dieId:<%v> is invalid, parse to int failed, err: %v", dieIdInt, err)
	}

	// dieId & 0x04(0000 0100) to get the 3rd bit of dieId lower byte
	if (dieIdInt & dieIdMaskNum) != 0 {
		dieId = 1
	} else {
		dieId = 0
	}

	portInt, err := strconv.ParseInt(portIdStr, hexadecimal, 0)
	if err != nil {
		return 0, 0, fmt.Errorf("portId:<%v> is invalid, parse to int failed, err: %v", portInt, err)
	}

	// Mask with 0x7F (0111 1111) to discard the MSB (Most Significant Bit), then right shift by 3 bits to drop the 3 least significant bits.
	// Example: 0x28 (0010 1000)
	// 1. Masking: 0x28 & 0x7F -> 0010 1000
	// 2. Shifting: 0010 1000 >> 3 -> 0000 0101 (Decimal: 5)
	portId = byte((portInt & portIdMaskNum) >> rightShiftLen)

	return dieId, portId, nil
}

func savePortsResult(edge Edge, ports []string, portsMap map[string]struct{}) []string {
	for _, addr := range edge.LocalAPorts {
		if _, ok := portsMap[addr]; ok {
			continue
		}
		ports = append(ports, addr)
		portsMap[addr] = struct{}{}
	}
	hwlog.RunLog.Infof("savePortsResult finished, total ports count=%d", len(ports))
	return ports
}

func savePortsResultByDieId(edge Edge, dieId byte, ports []string, portsMap map[string]struct{}) []string {
	for _, addr := range edge.LocalAPorts {
		if _, ok := portsMap[addr]; ok {
			continue
		}
		addrNums := strings.Split(addr, "/")
		if len(addrNums) == addrNumsLength && addrNums[0] == strconv.Itoa(int(dieId)) {
			ports = append(ports, addr)
			portsMap[addr] = struct{}{}
		}
	}
	return ports
}
