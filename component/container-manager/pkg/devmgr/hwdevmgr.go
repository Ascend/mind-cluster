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

// Package devmgr hwDevMgr function
package devmgr

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	ascommon "ascend-common/devmanager/common"
	"container-manager/pkg/common"
)

var (
	useIpv4 = true
)

const (
	ipv4Type           = 0
	ipv6Type           = 1
	ipv6LinkTypePrefix = "fe80"
	virMark            = "vir"
	defaultDeviceIP    = "127.0.0.1"
)

// SetDmgr set devmanager
func (hdm *HwDevMgr) SetDmgr(dmgr devmanager.DeviceInterface) {
	hdm.dmgr = dmgr
}

// GetDmgr get devmanager
func (hdm *HwDevMgr) GetDmgr() devmanager.DeviceInterface {
	return hdm.dmgr
}

// GetDevType get device type
func (hdm *HwDevMgr) GetDevType() string {
	return hdm.devType
}

// GetDevUsage get device usage
func (hdm *HwDevMgr) GetDevUsage() string {
	return hdm.devUsage
}

// GetDevNum get device number
func (hdm *HwDevMgr) GetDevNum() int {
	return len(hdm.npuInfos)
}

// GetPhyIds get phy ids
func (hdm *HwDevMgr) GetPhyIds() []int32 {
	var ids []int32
	for id := range hdm.npuInfos {
		ids = append(ids, id)
	}
	return ids
}

func (hdm *HwDevMgr) setDeviceUsage(phyId int32) error {
	if strings.HasPrefix(hdm.devType, api.Ascend310) {
		hdm.devUsage = common.Infer
		return nil
	}

	boardId, err := hdm.GetBoardId(phyId)
	if err != nil {
		hwlog.RunLog.Errorf("get board id failed, error: %v", err)
		return fmt.Errorf("get board id failed")
	}

	// A800IA2 without hccs can be auto set usage as infer
	if hdm.devType == api.Ascend910B && (boardId == common.A300IA2BoardId || boardId == common.A300IA2GB64BoardId ||
		boardId == common.A800IA2NoneHccsBoardId || boardId == common.A800IA2NoneHccsBoardIdOld) {
		hdm.devUsage = common.Infer
		return nil
	}

	hdm.devUsage = common.Train
	return nil
}

func (hdm *HwDevMgr) setRingInfo() error {
	for id, info := range hdm.npuInfos {
		devsOnRing, err := hdm.GetPhyIdOnRing(id)
		if err != nil {
			return err
		}
		info.DevsOnRing = devsOnRing
	}
	return nil
}

// GetPhyIdOnRing get phy ids on ring
func (hdm *HwDevMgr) GetPhyIdOnRing(phyId int32) ([]int32, error) {
	devNumPerRing, err := hdm.GetDevNumPerRing(phyId)
	if err != nil {
		return nil, err
	}
	if devNumPerRing == 0 {
		return nil, errors.New("dev num per ring is zero")
	}
	ringIdx := hdm.GetLogicIdByPhyId(phyId) / int32(devNumPerRing)
	startDevIdx := ringIdx * int32(devNumPerRing)
	endDevIdx := startDevIdx + int32(devNumPerRing)

	var phyIdsOnRing []int32
	for i := startDevIdx; i < endDevIdx; i++ {
		phyIdsOnRing = append(phyIdsOnRing, hdm.GetPhyIdByLogicId(i))
	}
	return phyIdsOnRing, nil
}

// GetDevNumPerRing get device number per ring
func (hdm *HwDevMgr) GetDevNumPerRing(phyID int32) (int, error) {
	boardId, err := hdm.GetBoardId(phyID)
	if err != nil {
		hwlog.RunLog.Errorf("get board id failed: %v", err)
		return 0, errors.New("get board id failed")
	}
	return common.GetDevNumPerRing(hdm.GetDevType(), hdm.GetDevUsage(), hdm.GetDevNum(), boardId), nil
}

func (hdm *HwDevMgr) setBoardId(logicId int32) error {
	boardInfo, err := hdm.dmgr.GetBoardInfo(logicId)
	if err != nil {
		hwlog.RunLog.Errorf("get board id failed, error: %v", err)
		return fmt.Errorf("get board id failed")
	}
	hdm.boardId = boardInfo.BoardId
	return nil
}

// GetBoardId get board id
func (hdm *HwDevMgr) GetBoardId(phyId int32) (uint32, error) {
	if hdm.boardId != common.EmptyBoardId {
		return hdm.boardId, nil
	}
	boardInfo, err := hdm.dmgr.GetBoardInfo(hdm.GetLogicIdByPhyId(phyId))
	if err != nil {
		return common.EmptyBoardId, fmt.Errorf("get board info failed, error: %v", err)
	}
	hdm.boardId = boardInfo.BoardId
	return hdm.boardId, nil
}

// GetNodeNPUInfo get node npu info
func (hdm *HwDevMgr) GetNodeNPUInfo() map[int32]*common.NPUInfo {
	return hdm.npuInfos
}

func (hdm *HwDevMgr) setNodeNPUInfo(logicIds []int32, devNum int32) (map[int32]*common.NPUInfo, error) {
	var npuInfos = make(map[int32]*common.NPUInfo, devNum)
	for _, id := range logicIds {
		npuInfo, err := hdm.constructNPUInfo(id)
		if err != nil {
			return nil, err
		}
		npuInfos[id] = &npuInfo
	}
	return npuInfos, nil
}

func (hdm *HwDevMgr) constructNPUInfo(logicID int32) (common.NPUInfo, error) {
	phyID, err := hdm.dmgr.GetPhysicIDFromLogicID(logicID)
	if err != nil {
		return common.NPUInfo{}, err
	}
	cardID, deviceID, err := hdm.dmgr.GetCardIDDeviceID(logicID)
	if err != nil {
		return common.NPUInfo{}, err
	}
	ip, err := hdm.getDeviceIP(logicID)
	if err != nil {
		hwlog.RunLog.Warnf("get device ip failed, err: %v", err)
		ip = ""
	}
	return common.NPUInfo{
		LogicID:  logicID,
		PhyID:    phyID,
		CardID:   cardID,
		IP:       ip,
		DeviceID: deviceID,
	}, nil
}

func (hdm *HwDevMgr) getDeviceIP(logicID int32) (string, error) {
	chip, err := hdm.dmgr.GetChipInfo(logicID)
	if err != nil {
		return "", fmt.Errorf("get logicId(%d) chip info failed, error: %v", logicID, err)
	}
	if strings.Contains(chip.Name, virMark) {
		return defaultDeviceIP, nil
	}
	return hdm.getDcmiDeviceIP(logicID)
}

func (hdm *HwDevMgr) getDcmiDeviceIP(logicID int32) (string, error) {
	var deviceIp string
	var err error
	if useIpv4 {
		if deviceIp, err = hdm.dmgr.GetDeviceIPAddress(logicID, ipv4Type); err == nil {
			return deviceIp, nil
		}
		useIpv4 = false
	}

	if !useIpv4 {
		deviceIp, err = hdm.dmgr.GetDeviceIPAddress(logicID, ipv6Type)
		if err != nil {
			useIpv4 = true
			return "", err
		}
		if strings.Index(deviceIp, ipv6LinkTypePrefix) == 0 {
			return "", fmt.Errorf("logicID(%d) device ip %v is a link type ipv6 address", logicID, deviceIp)
		}
	}
	return deviceIp, nil
}

// GetLogicIdByPhyId get logic id by phy id
func (hdm *HwDevMgr) GetLogicIdByPhyId(logicId int32) int32 {
	for _, npuInfo := range hdm.npuInfos {
		if npuInfo.LogicID == logicId {
			return npuInfo.PhyID
		}
	}
	return 0
}

// GetPhyIdByLogicId get phy id by logic id
func (hdm *HwDevMgr) GetPhyIdByLogicId(phyId int32) int32 {
	for _, npuInfo := range hdm.npuInfos {
		if npuInfo.PhyID == phyId {
			return npuInfo.LogicID
		}
	}
	return 0
}

// SubscribeFaultEvent subscribe fault event
func (hdm *HwDevMgr) SubscribeFaultEvent(callback func(devFaultInfo ascommon.DevFaultInfo)) error {
	return hdm.subscribeNPUFaultEvent(callback)
}

func (hdm *HwDevMgr) subscribeNPUFaultEvent(callback func(devFaultInfo ascommon.DevFaultInfo)) error {
	if err := hdm.GetDmgr().SetFaultEventCallFunc(callback); err != nil {
		hwlog.RunLog.Errorf("set fault event call back function failed, error: %v", err)
		return errors.New("set fault event call back function failed")
	}

	const retryTime = 3
	var subSuc bool
	var err error
	for i := 0; i < retryTime; i++ {
		if err = hdm.GetDmgr().SubscribeDeviceFaultEvent(ascommon.SubscribeAllDevice); err == nil {
			subSuc = true
			break
		}
		hwlog.RunLog.Errorf("subscribe device fault event failed, error: %v, try again", err)
		time.Sleep(time.Second)
	}
	if !subSuc {
		hwlog.RunLog.Error("subscribe device fault event failed")
		return errors.New("subscribe device fault event failed")
	}
	hwlog.RunLog.Info("subscribe device fault event success")
	return nil
}

// GetDeviceErrCode get device error code by dcmi interface
func (hdm *HwDevMgr) GetDeviceErrCode(phyId int32) (int32, []int64, error) {
	logicId := hdm.GetLogicIdByPhyId(phyId)
	return hdm.GetDmgr().GetDeviceAllErrorCode(logicId)
}

// GetFaultCodesMap get fault codes map
func (hdm *HwDevMgr) GetFaultCodesMap() map[int32][]int64 {
	var idCodesMap = make(map[int32][]int64)
	for phyId := range hdm.npuInfos {
		_, codes, err := hdm.GetDeviceErrCode(phyId)
		if err != nil {
			hwlog.RunLog.Errorf("get device error code failed, error: %v", err)
			continue
		}
		if len(codes) == 0 {
			continue
		}
		for _, code := range codes {
			hwlog.RunLog.Infof("before fault subscribe, device(%d) had error code: %v",
				phyId, strconv.FormatInt(code, common.Hex))
		}
		idCodesMap[phyId] = codes
	}
	return idCodesMap
}
