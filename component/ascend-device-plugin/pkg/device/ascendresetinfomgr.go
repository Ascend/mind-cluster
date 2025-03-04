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

// Package device a series of device function
package device

import (
	"encoding/json"
	"strconv"
	"sync"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/common-utils/hwlog"
)

// ResetInfoMgr mgr for npu reset
type ResetInfoMgr struct {
	client    *kubeclient.ClientK8s
	resetInfo *ResetInfo
	busyDevs  sync.Map // map[int32]struct to record which device is busy
	resetCnt  sync.Map // map[int32]int to record the reset count of each device
	mu        sync.RWMutex
}

// ResetInfo information of npu reset
type ResetInfo struct {
	// ThirdPartyResetDevs devices waits for third party to reset
	ThirdPartyResetDevs []ResetDevice
	// ManualResetDevs devices waits for manually reset
	ManualResetDevs []ResetDevice
}

// ResetDevice device that fail to be reset
type ResetDevice struct {
	// CardId npu card id
	CardId int32
	// DeviceId npu device id
	DeviceId int32
	// AssociatedCardId card id of the associated npu
	AssociatedCardId int32
	// PhyId npu physic id
	PhyID int32
	// LogicID npu logic id
	LogicID int32
}

// WriteMode the mode determines how the content is written
type WriteMode int

const (
	// WMOverwrite write mode which will overwrite content
	WMOverwrite WriteMode = iota
	// WMAppend write mode which will append to content
	WMAppend
	// WMDelete write mode which will delete from content
	WMDelete
)

var (
	mgr  *ResetInfoMgr
	once sync.Once
)

// InitResetInfoMgr initialize ResetInfoMgr globally
func InitResetInfoMgr(client *kubeclient.ClientK8s) {
	once.Do(func() {
		infoMgr := ResetInfoMgr{
			client:    client,
			resetInfo: &ResetInfo{},
		}
		curNode, err := client.GetNode()
		if err != nil {
			hwlog.RunLog.Errorf("fail to get node from k8s, err: %v", err)
			mgr = &infoMgr
			return
		}
		if curNode.Annotations == nil {
			mgr = &infoMgr
			return
		}
		infoMgr.resetInfo = readAnnotation(curNode.Annotations, common.ResetInfoAnnotationKey)
		mgr = &infoMgr
	})
}

// GetResetInfoMgr return the single instance of reset mgr, load reset info from node annotation
func GetResetInfoMgr() *ResetInfoMgr {
	return mgr
}

// WriteResetInfo write reset info into cache and node annotation
func WriteResetInfo(resetInfo ResetInfo, writeMode WriteMode) {
	mgr.mu.Lock()
	mgr.resetInfo.ThirdPartyResetDevs = mergeFailDevs(mgr.resetInfo.ThirdPartyResetDevs,
		resetInfo.ThirdPartyResetDevs, writeMode)
	mgr.resetInfo.ManualResetDevs = mergeFailDevs(mgr.resetInfo.ManualResetDevs,
		resetInfo.ManualResetDevs, writeMode)
	hwlog.RunLog.Infof("reset info change: %v", *mgr.resetInfo)
	dataBytes, err := json.Marshal(*mgr.resetInfo)
	if err != nil {
		hwlog.RunLog.Errorf("marshal reset info error, data: %v, err: %v", *mgr.resetInfo, err)
		mgr.mu.Unlock()
		return
	}
	mgr.mu.Unlock()
	writeNodeAnnotation(string(dataBytes))
}

// ReadResetInfo read reset info from cache
func ReadResetInfo() ResetInfo {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	return *mgr.resetInfo
}

// IsDevBusy check whether one device is busy, for example in reset, wait third party reset or wait manually reset
func IsDevBusy(cardID, deviceID int32) bool {
	_, exist := mgr.busyDevs.Load(combineToString(cardID, deviceID))
	return exist
}

// AddBusyDev add a new busy device
func AddBusyDev(cardID, deviceID int32) {
	mgr.busyDevs.Store(combineToString(cardID, deviceID), struct{}{})
}

// FreeBusyDev remove a device from busy map
func FreeBusyDev(cardID, deviceID int32) {
	mgr.busyDevs.Delete(combineToString(cardID, deviceID))
}

// GetResetCnt get device reset count by physic ID
func GetResetCnt(cardID, deviceID int32) int {
	cnt, exist := mgr.resetCnt.Load(combineToString(cardID, deviceID))
	if !exist {
		return 0
	}

	ret, ok := cnt.(int)
	if !ok {
		hwlog.RunLog.Warnf("reset cnt map invalid value, val: %v", cnt)
		mgr.resetCnt.Store(combineToString(cardID, deviceID), 0)
		return 0
	}
	return ret
}

// AddResetCnt add device reset count
func AddResetCnt(cardID, deviceID int32) {
	cnt := GetResetCnt(cardID, deviceID)
	SetResetCnt(cardID, deviceID, cnt+1)
}

// SetResetCnt set device reset count
func SetResetCnt(cardID, deviceID int32, cnt int) {
	mgr.resetCnt.Store(combineToString(cardID, deviceID), cnt)
}

func writeNodeAnnotation(resetStr string) {
	if err := mgr.client.AddAnnotation(common.ResetInfoAnnotationKey, resetStr); err != nil {
		hwlog.RunLog.Errorf("fail to write reset info to node annotation, err: %v", err)
	}
}

func mergeFailDevs(curDevs, newDevs []ResetDevice, writeMode WriteMode) []ResetDevice {
	switch writeMode {
	case WMOverwrite:
		return newDevs
	case WMAppend:
		return mergeAndDeduplicate(curDevs, newDevs)
	case WMDelete:
		return deleteFrom(curDevs, newDevs)
	default:
		hwlog.RunLog.Errorf("write mode %v is invalid", writeMode)
		return curDevs
	}
}

func mergeAndDeduplicate(curArr, NewArr []ResetDevice) []ResetDevice {
	seen := make(map[int32]struct{})
	result := make([]ResetDevice, 0)

	for _, v := range curArr {
		if _, exists := seen[v.PhyID]; !exists {
			seen[v.PhyID] = struct{}{}
			result = append(result, v)
		}
	}

	for _, v := range NewArr {
		if _, exists := seen[v.PhyID]; !exists {
			seen[v.PhyID] = struct{}{}
			result = append(result, v)
		}
	}

	return result
}

func deleteFrom(curArr, delArr []ResetDevice) []ResetDevice {
	ret := make([]ResetDevice, 0, len(curArr))
	toDelMap := make(map[int32]struct{})
	for _, dev := range delArr {
		if _, exist := toDelMap[dev.PhyID]; !exist {
			toDelMap[dev.PhyID] = struct{}{}
		}
	}
	for _, dev := range curArr {
		if _, exist := toDelMap[dev.PhyID]; !exist {
			ret = append(ret, dev)
		}
	}
	return ret
}

func readAnnotation(annotation map[string]string, key string) *ResetInfo {
	if _, exist := annotation[key]; !exist {
		return &ResetInfo{}
	}
	var ret ResetInfo
	if err := json.Unmarshal([]byte(annotation[key]), &ret); err != nil {
		hwlog.RunLog.Errorf("unmarshal node annotation failed, err: %v", err)
		return &ResetInfo{}
	}
	return &ret
}

func combineToString(a, b int32) string {
	return strconv.Itoa(int(a)) + common.UnderLine + strconv.Itoa(int(b))
}
