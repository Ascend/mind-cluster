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
	"sync"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/common-utils/hwlog"
)

// ResetInfoMgr mgr for npu reset
type ResetInfoMgr struct {
	client    *kubeclient.ClientK8s
	resetInfo *ResetInfo
	mu        sync.RWMutex
}

// ResetInfo information of npu reset
type ResetInfo struct {
	// ThirdPartyResetDevs devices waits for third party to reset
	ThirdPartyResetDevs []ResetFailDevice
	// ManualResetDevs devices waits for manually reset
	ManualResetDevs []ResetFailDevice
}

// ResetFailDevice device that fail to be reset
type ResetFailDevice struct {
	// CardId npu card id
	CardId int32
	// DeviceId npu device id
	DeviceId int32
	// AssociatedCardId card id of the associated npu
	AssociatedCardId int32
	// PhyId npu physic id
	PhyID int32
}

// WriteMode the mode determines how the content is written
type WriteMode int

const (
	// WMOverwrite write mode which will overwrite content
	WMOverwrite WriteMode = iota
	// WMAppend write mode which will append to content
	WMAppend
)

var (
	instance *ResetInfoMgr
	once     sync.Once
)

// GetResetInfoMgr return the single instance of reset mgr, load reset info from node annotation
func GetResetInfoMgr(client *kubeclient.ClientK8s) *ResetInfoMgr {
	once.Do(func() {
		infoMgr := ResetInfoMgr{
			client:    client,
			resetInfo: &ResetInfo{},
		}
		curNode, err := client.GetNode()
		if err != nil {
			hwlog.RunLog.Errorf("fail to get node from k8s, err: %v", err)
			instance = &infoMgr
			return
		}
		if curNode.Annotations == nil {
			instance = &infoMgr
			return
		}
		infoMgr.resetInfo = readAnnotation(curNode.Annotations, common.ResetInfoAnnotationKey)
		instance = &infoMgr
	})
	return instance
}

// WriteResetInfo write reset info into cache and node annotation
func (mgr *ResetInfoMgr) WriteResetInfo(resetInfo ResetInfo, writeMode WriteMode) {
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
	mgr.writeNodeAnnotation(string(dataBytes))
}

// ReadResetInfo read reset info from cache
func (mgr *ResetInfoMgr) ReadResetInfo() ResetInfo {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	return *mgr.resetInfo
}

func (mgr *ResetInfoMgr) writeNodeAnnotation(resetStr string) {
	if err := mgr.client.AddAnnotation(common.ResetInfoAnnotationKey, resetStr); err != nil {
		hwlog.RunLog.Errorf("fail to write reset info to node annotation, err: %v", err)
	}
}

func mergeFailDevs(curDevs []ResetFailDevice, newDevs []ResetFailDevice, writeMode WriteMode) []ResetFailDevice {
	if writeMode == WMOverwrite {
		return newDevs
	}
	if writeMode == WMAppend {
		curDevs = append(curDevs, newDevs...)
		return curDevs
	}
	hwlog.RunLog.Errorf("write mode %v is invalid", writeMode)
	return curDevs
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
