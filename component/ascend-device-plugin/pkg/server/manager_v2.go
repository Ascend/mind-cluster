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
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"k8s.io/api/core/v1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	npuCommon "ascend-common/devmanager/common"
)

const npuNicMappingConfigPath = "/user/mindx-dl/npu/npu-nic-mapping.json"

// NpuNicMapping is npu nic mapping config
type NpuNicMapping struct {
	NpuNics []NpuNicItem `json:"npuNics"`
}

// NpuNicItem is npu nic item
type NpuNicItem struct {
	NpuId    int      `json:"npuId"`
	NicNames []string `json:"nicNames"`
}

var (
	npuNicMappingCache     *NpuNicMapping
	npuNicMappingCacheOnce sync.Once
	npuNicMappingErr       error
)

func (hdm *HwDevManager) getCardType() (string, error) {
	boardInfo, err := hdm.manager.GetDmgr().GetBoardInfo(hdm.allInfo.AllDevs[common.FirstDevice].LogicID)
	if err != nil {
		return "", err
	}

	if boardInfo.BoardId != npuCommon.A5300IBoardId && boardInfo.BoardId != npuCommon.A5300IBoardId2 &&
		boardInfo.BoardId != npuCommon.A5300IBoardId3 {
		return "", nil
	}

	mainBoardId := hdm.manager.GetDmgr().GetMainBoardId()

	if mainBoardId == common.A5300IMainBoardId {
		return common.A5300ICardName, nil
	}
	if mainBoardId == common.A5300I4PMainBoardId {
		return common.A54P300ICardName, nil
	}

	return "", nil
}

func (hdm *HwDevManager) getProductInfo() *ProductBase {
	if hdm.manager == nil {
		return nil
	}
	dmgr := hdm.manager.GetDmgr()
	if dmgr == nil {
		return nil
	}
	mainBoardId := int(dmgr.GetMainBoardId())
	return &ProductBase{
		superPodSize:   uint32(hdm.manager.GetSuperPodSize()),
		superPodID:     uint32(hdm.manager.GetSuperPodID()),
		serverIndex:    uint32(hdm.manager.GetServerIndex()),
		chassisID:      uint32(hdm.manager.GetRackID()),
		superPodType:   uint8(hdm.manager.GetSuperPodType()),
		nodeInternalIP: hdm.manager.GetNodeInternalIPInK8s(),
		cardType:       common.ParamOption.CardType,
		mainBoardId:    mainBoardId,
		maxNpuCount:    npuCommon.GetMaxNpuCountPerNode(mainBoardId),
	}
}

// getLevelList get node baseDeviceInfo levelList info for rank table
func (hdm *HwDevManager) getLevelList(dev *common.NpuDevice) []api.RankLevel {
	if common.ParamOption.RealCardType != api.Ascend910A5 {
		hwlog.RunLog.Debugf("real card type is %v, no levelList information", common.ParamOption.RealCardType)
		return nil
	}
	if dev == nil {
		hwlog.RunLog.Error("input parameter dev is empty")
		return nil
	}
	npuBase.productInfo = hdm.getProductInfo()
	if err := npuBase.SetUrmaDeviceInfoByHdm(hdm, dev); err != nil {
		hwlog.RunLog.Errorf("set urma device info by hdm failed for LogicID(%d) phyID(%d), err: %v",
			dev.LogicID, dev.PhyID, err)
		// a5 standard card no mesh scene, there is no urma device and eid info, should generate rank table level_list
	}

	infoKeyArr := npuBase.getRankLevelInfoKeyArr()
	levelList := make([]api.RankLevel, 0)
	for level := 0; level < len(infoKeyArr); level++ {
		infoKey := infoKeyArr[level]
		if infoKey == "" {
			continue
		}
		rankAddrList := hdm.getRankAddrList(level, dev)
		if len(rankAddrList) == 0 {
			hwlog.RunLog.Warnf("rank addr list is empty for LogicID(%d) phyID(%d) level(%d) netType(%s)",
				dev.LogicID, dev.PhyID, level, infoKey)
			continue
		}
		info := map[string]api.LevelElement{
			infoKey: {
				NetLayer:      level,
				NetInstanceID: npuBase.getID(level),
				NetType:       npuBase.getNetTypeForLevel(level),
				NetAttr:       api.NetAttrEmpty,
				RankAddrList:  rankAddrList,
			},
		}
		levelList = append(levelList, api.RankLevel{Level: level, Info: info})
	}

	return levelList
}

// getRankAddrList for get the rank addr list in rant table for A5
func (hdm *HwDevManager) getRankAddrList(level int, dev *common.NpuDevice) []api.RankAddrItem {
	if dev == nil {
		return nil
	}
	product := hdm.getProductInfo()
	if product == nil {
		return nil
	}

	// RoCE：same logic for both server and pod
	if level == api.RankLevel3 {
		return hdm.getROCEAddrList(dev)
	}

	// Standcard: keep the original logic
	if product.isStandCard() {
		return hdm.getRankAddrListOriginal(level, dev)
	}

	// Parse all URMA devices once (die/fe/port/PG/UBOE/IP)
	urmaList := hdm.GetUrmaDeviceList(dev)
	if len(urmaList) == 0 {
		return nil
	}
	// Parse URMA（die/fe/port/PG/UBOE/IP）
	parsed := ParseUrmaDevices(urmaList)
	// Pod scene
	if product.isPodScene() {
		return npuBase.buildPodRankAddrListParsed(level, dev, parsed)
	}
	// Server scene
	if product.isServer() {
		return npuBase.buildServerRankAddrListParsed(level, parsed)
	}
	return nil
}

// GetUrmaDeviceList get urma device eid list from dcmi
func (hdm *HwDevManager) GetUrmaDeviceList(dev *common.NpuDevice) []*UrmaDevice {
	dmgr := hdm.manager.GetDmgr()
	if dmgr == nil {
		return nil
	}

	infoList, err := dmgr.GetUrmaDevEidListAll(dev.LogicID)
	if err != nil {
		return nil
	}
	result := make([]*UrmaDevice, 0)
	for _, info := range infoList {
		u := &UrmaDevice{
			EidList: make([]string, 0),
		}
		for i := 0; i < int(info.EidCount); i++ {
			raw := info.EidInfos[i].Eid.Raw[:]
			eid := RawBytesToEidString(raw)
			u.EidList = append(u.EidList, eid)
		}
		result = append(result, u)
	}
	return result
}

// getRankAddrListOriginal for get the rank addr list in rant table for A5 stand card
func (hdm *HwDevManager) getRankAddrListOriginal(level int, dev *common.NpuDevice) []api.RankAddrItem {
	netType, feIdList := npuBase.getNetTypeAndFeIDListByRankLevel(level)
	rankAddrList := make([]api.RankAddrItem, 0)
	for _, feId := range feIdList {
		addrs := npuBase.getRandAddrByFuncEntityID(dev.PhyID, feId, netType, level)
		rankAddrList = append(rankAddrList, addrs...)
	}
	return rankAddrList
}

func getNpuNicMappingCache() (*NpuNicMapping, error) {
	npuNicMappingCacheOnce.Do(func() {
		data, err := utils.LoadFile(npuNicMappingConfigPath)
		if err != nil {
			npuNicMappingErr = fmt.Errorf("read config file error: %v", err)
			return
		}

		if data == nil {
			hwlog.RunLog.Warnf("npu-nic-mapping config file not found: %s", npuNicMappingConfigPath)
			npuNicMappingCache = nil
			return
		}

		var mapping NpuNicMapping
		if err = json.Unmarshal(data, &mapping); err != nil {
			npuNicMappingErr = fmt.Errorf("parse config file error: %v", err)
			return
		}

		npuNicMappingCache = &mapping
		hwlog.RunLog.Infof("npu-nic-mapping config loaded: %v", mapping)
	})

	return npuNicMappingCache, npuNicMappingErr
}

func getIPAddressType(ip string) string {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return addrTypeIPV4 // default IPv4
	}
	if parsedIP.To4() != nil {
		return addrTypeIPV4
	}
	return addrTypeIPV6
}

func getInterfaceIPsByPriority(nicNames []string) (string, error) {
	for _, nicName := range nicNames {
		ips := getInterfaceIPs(nicName)
		if len(ips) > 0 {
			return ips[0], nil
		}
		hwlog.RunLog.Warnf("interface %s has no valid IP address, checking next interface", nicName)
	}
	return "", fmt.Errorf("no valid IP address found for any interface: %v", nicNames)
}

func getInterfaceIPs(nicName string) []string {
	var ips []string
	iface, err := net.InterfaceByName(nicName)
	if err != nil {
		hwlog.RunLog.Errorf("get interface %s error: %v", nicName, err)
		return ips
	}

	addrs, err := iface.Addrs()
	if err != nil || len(addrs) == 0 {
		hwlog.RunLog.Errorf("get interface %s addrs error: %v", nicName, err)
		return ips
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipNet.IP
		// IsLoopback: IPv4(127.0.0.1) and IPv6(::1)
		// IsLinkLocalUnicast: IPv4(169.254.0.0/16) and IPv6(fe80::/10)
		// IsLinkLocalMulticast: IPv6(ff02::/16)
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			continue
		}
		ips = append(ips, ip.String())
	}
	return ips
}

func getNpuToNicNames(npuId int) ([]string, error) {
	mapping, err := getNpuNicMappingCache()
	if err != nil {
		return nil, err
	}

	if mapping == nil {
		return nil, nil
	}

	for _, item := range mapping.NpuNics {
		if item.NpuId == npuId {
			return item.NicNames, nil
		}
	}

	return nil, fmt.Errorf("npuId %d not found in mapping", npuId)
}

// getROCEAddrList get RoCE addr list of device in A5
func (hdm *HwDevManager) getROCEAddrList(dev *common.NpuDevice) []api.RankAddrItem {
	if dev == nil {
		hwlog.RunLog.Error("device is nil")
		return []api.RankAddrItem{}
	}

	npuId := int(dev.PhyID % common.NpuNum)
	nicNames, err := getNpuToNicNames(npuId)
	if err != nil {
		hwlog.RunLog.Warnf("get npu %d nic names failed: %v, using legacy dpu info", npuId, err)
		return hdm.getROCEAddrListLegacy(dev)
	}

	if nicNames == nil {
		hwlog.RunLog.Warnf("npu-nic-mapping config not found, using legacy dpu info")
		return hdm.getROCEAddrListLegacy(dev)
	}

	ip, err := getInterfaceIPsByPriority(nicNames)
	if err != nil {
		hwlog.RunLog.Errorf("get roce addr list failed: %v", err)
		return []api.RankAddrItem{}
	}

	addrType := getIPAddressType(ip)
	hwlog.RunLog.Infof("get RoCE addr for NPU %d: %s (type: %s)", npuId, ip, addrType)

	return []api.RankAddrItem{
		{
			AddrType: addrType,
			Addr:     ip,
			Ports:    []string{},
			PlaneId:  api.DefaultRandAddrPlaneID,
		},
	}
}

func (hdm *HwDevManager) getROCEAddrListLegacy(dev *common.NpuDevice) []api.RankAddrItem {
	dpuIPList, err := hdm.getNpuCorrespDpuInfo(dev)
	if err != nil {
		hwlog.RunLog.Errorf("get roce addr list failed, err: %v", err)
		return []api.RankAddrItem{}
	}

	rankAddrList := make([]api.RankAddrItem, 0)
	for _, ip := range dpuIPList {
		addrType := getIPAddressType(ip)
		rankAddrList = append(rankAddrList, api.RankAddrItem{
			AddrType: addrType,
			Addr:     ip,
			Ports:    []string{},
			PlaneId:  api.DefaultRandAddrPlaneID,
		})
	}
	return rankAddrList
}

// GetDevManager get device manager instance
func (hdm *HwDevManager) GetDevManager() device.DevManager {
	return hdm.manager
}

// GetRackID get id of rack
func (hdm *HwDevManager) GetRackID() int32 {
	return hdm.manager.GetRackID()
}

// GetSuperPodID get id of current super pod
func (hdm *HwDevManager) GetSuperPodID() int32 {
	return hdm.manager.GetSuperPodID()
}

// GetSuperPodType get type of current super pod
func (hdm *HwDevManager) GetSuperPodType() int8 {
	return hdm.manager.GetSuperPodType()
}

// SetNodeInternalIPInK8s get super pod info then cache it
func (hdm *HwDevManager) SetNodeInternalIPInK8s(node *v1.Node) {
	if common.ParamOption.RealCardType != api.Ascend910A5 {
		hwlog.RunLog.Infof("real card type is %v, no need server ip in k8s", common.ParamOption.RealCardType)
		return
	}
	if node == nil {
		hwlog.RunLog.Error("node is empty")
		return
	}
	internalIP := ""
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			internalIP = addr.Address
			break
		}
	}
	hdm.manager.SetNodeInternalIPInK8s(internalIP)
	return
}

// getNpuCorrespDpuInfo get npu dpu info
func (hdm *HwDevManager) getNpuCorrespDpuInfo(dev *common.NpuDevice) ([]string, error) {
	if hdm.dpuManager.NpuWithDpuInfos == nil {
		return nil, fmt.Errorf("dpu infos is empty")
	}
	npuPhyId := dev.PhyID
	npuId := npuPhyId % common.NpuNum
	var ipAddrs []string
	for _, NpuWithDpuInfo := range hdm.dpuManager.NpuWithDpuInfos {
		if NpuWithDpuInfo.NpuId == npuId {
			for _, DpuInfo := range NpuWithDpuInfo.DpuInfo {
				ipAddrs = append(ipAddrs, DpuInfo.DpuIP)
			}
			return ipAddrs, nil
		}
	}
	return nil, fmt.Errorf("get npu %d correspond dpuinfos error", npuId)
}
