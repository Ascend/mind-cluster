/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package manager for fault function
package manager

import (
	"strings"
	"sync"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
)

// FaultManager manage fault device info
type FaultManager interface {
	GetFaultDevInfo() *common.FaultDevInfo
	GetFaultDevList() []*common.FaultDev
	GetNodeStatus() string
	SetFaultDevInfo(*common.FaultDevInfo)
	SetFaultDevList([]*common.FaultDev)
	SetNodeStatus(string)
}

// NewFaultManager create
func NewFaultManager() FaultManager {
	return &FaultTools{
		faultDevInfo: &common.FaultDevInfo{},
		devInfoLock:  &sync.Mutex{},
	}
}

// FaultTools the fault tool definition
type FaultTools struct {
	faultDevInfo *common.FaultDevInfo
	devInfoLock  *sync.Mutex
}

// GetFaultDevInfo return fault device info
func (f *FaultTools) GetFaultDevInfo() *common.FaultDevInfo {
	if f.faultDevInfo == nil {
		hwlog.RunLog.Error("fault dev info is nil when get fault dev info")
		return nil
	}
	return f.faultDevInfo
}

// GetFaultDevList return fault device list
func (f *FaultTools) GetFaultDevList() []*common.FaultDev {
	if f.faultDevInfo == nil {
		hwlog.RunLog.Error("fault dev info is nil when get fault device list")
		return nil
	}
	return f.faultDevInfo.FaultDevList
}

// GetNodeStatus return node status
func (f *FaultTools) GetNodeStatus() string {
	if f.faultDevInfo == nil {
		hwlog.RunLog.Error("fault dev info is nil when get node status")
		return ""
	}
	return f.faultDevInfo.NodeStatus
}

// SetFaultDevInfo set fault device info
func (f *FaultTools) SetFaultDevInfo(faultDevInfo *common.FaultDevInfo) {
	if f.faultDevInfo == nil {
		hwlog.RunLog.Error("fault dev info is nil when set fault device info")
		return
	}
	f.SetFaultDevList(faultDevInfo.FaultDevList)
	f.SetNodeStatus(faultDevInfo.NodeStatus)
}

// SetFaultDevList set fault device list
func (f *FaultTools) SetFaultDevList(faultDevList []*common.FaultDev) {
	if f.faultDevInfo == nil {
		hwlog.RunLog.Error("fault dev info is nil when set fault device list")
		return
	}
	f.faultDevInfo.FaultDevList = make([]*common.FaultDev, len(faultDevList))
	for i, faultDev := range faultDevList {
		faultDevTmp := &common.FaultDev{}
		f.DeepCopyFaultDev(faultDevTmp, faultDev)
		f.faultDevInfo.FaultDevList[i] = faultDevTmp
	}
}

// SetNodeStatus set node status
func (f *FaultTools) SetNodeStatus(status string) {
	if f.faultDevInfo == nil {
		hwlog.RunLog.Error("fault dev info is nil when set node status")
		return
	}
	f.faultDevInfo.NodeStatus = status
}

// DeepCopyFaultDev deep copy fault device
func (f *FaultTools) DeepCopyFaultDev(oldFaultDev, newFaultDev *common.FaultDev) {
	oldFaultDev.FaultCode = common.CopyStringSlice(newFaultDev.FaultCode)
	oldFaultDev.DeviceType = strings.Clone(newFaultDev.DeviceType)
	oldFaultDev.DeviceId = newFaultDev.DeviceId
	oldFaultDev.FaultLevel = strings.Clone(newFaultDev.FaultLevel)
}
