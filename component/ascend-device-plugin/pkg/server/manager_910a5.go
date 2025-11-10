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
	"fmt"

	"k8s.io/api/core/v1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	npuCommon "ascend-common/devmanager/common"
)

func (hdm *HwDevManager) getCardType() (string, error) {
	boardInfo, err := hdm.manager.GetDmgr().GetBoardInfo(hdm.allInfo.AllDevs[common.FirstDevice].LogicID)
	if err != nil {
		return "", err
	}

	if boardInfo.BoardId != npuCommon.A5300IBoardId && boardInfo.BoardId != npuCommon.A5300IBoardId2 {
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
	return &ProductBase{
		superPodSize:   uint32(hdm.manager.GetSuperPodSize()),
		superPodID:     uint32(hdm.manager.GetSuperPodID()),
		serverIndex:    uint32(hdm.manager.GetServerIndex()),
		chassisID:      uint32(hdm.manager.GetRackID()),
		superPodType:   uint8(hdm.manager.GetSuperPodType()),
		nodeInternalIP: hdm.manager.GetNodeInternalIPInK8s(),
		cardType:       common.ParamOption.CardType,
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
		hwlog.RunLog.Errorf("set urma device info by hdm failed for cardID(%d) deviceID(%d) phyID(%d), err: %v",
			dev.CardID, dev.DeviceID, dev.PhyID, err)
		return nil
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
			hwlog.RunLog.Warnf("rank addr list is empty for cardID(%d) deviceID(%d) phyID(%d) level(%d) netType(%s)",
				dev.CardID, dev.DeviceID, dev.PhyID, level, infoKey)
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
	if level == api.RankLevel3 {
		return hdm.getROCEAddrList(dev)
	}
	netType, feIdList := npuBase.getNetTypeAndFeIDListByRankLevel(level)
	rankAddrList := make([]api.RankAddrItem, 0)
	for _, feId := range feIdList {
		addrs := npuBase.getRandAddrByFuncEntityID(dev.PhyID, feId, netType, level)
		rankAddrList = append(rankAddrList, addrs...)
	}
	return rankAddrList
}

// getROCEAddrList get RoCE addr list of device in A5
func (hdm *HwDevManager) getROCEAddrList(dev *common.NpuDevice) []api.RankAddrItem {
	dpuIPList, err := hdm.getNpuCorrespDpuInfo(dev)
	if err != nil {
		hwlog.RunLog.Errorf("get roce addr list failed, err: %v", err)
		return []api.RankAddrItem{}
	}

	rankAddrList := make([]api.RankAddrItem, 0)
	for _, ip := range dpuIPList {
		rankAddrList = append(rankAddrList, api.RankAddrItem{
			AddrType: "IPV4",
			Addr:     ip,
			Ports:    []string{},
			PlaneId:  api.DefaultRandAddrPlaneID,
		})
	}
	return rankAddrList
}

// GetDevManager get device manger instance
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
