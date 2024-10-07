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

// Package common for common function
package common

import (
	"strings"
	"sync"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
)

// FaultManager manage fault device info
type FaultManager interface {
	GetFaultDevInfo() *FaultDevInfo
	GetFaultDevList() []*FaultDev
	GetHeartbeatTime() int64
	GetHeartbeatInterval() int
	GetNodeStatus() string
	SetFaultDevInfo(*FaultDevInfo)
	SetFaultDevList([]*FaultDev)
	SetHeartbeatTime(int64)
	SetHeartbeatInterval(int)
	SetNodeStatus(string)
}

// NewFaultManager create
func NewFaultManager() FaultManager {
	return &FaultTools{
		faultDevInfo: &FaultDevInfo{},
		devInfoLock:  &sync.Mutex{},
	}
}

// FaultTools the fault tool definition
type FaultTools struct {
	faultDevInfo *FaultDevInfo
	devInfoLock  *sync.Mutex
}

// GetFaultDevInfo return fault device info
func (f *FaultTools) GetFaultDevInfo() *FaultDevInfo {
	if f.faultDevInfo == nil {
		hwlog.RunLog.Error("fault dev info is nil when get fault dev info")
		return nil
	}
	return f.faultDevInfo
}

// GetFaultDevList return fault device list
func (f *FaultTools) GetFaultDevList() []*FaultDev {
	if f.faultDevInfo == nil {
		hwlog.RunLog.Error("fault dev info is nil when get fault device list")
		return nil
	}
	return f.faultDevInfo.FaultDevList
}

// GetHeartbeatTime return heartbeat time
func (f *FaultTools) GetHeartbeatTime() int64 {
	if f.faultDevInfo == nil {
		hwlog.RunLog.Error("fault dev info is nil when get heartbeat time")
		return -1
	}
	return f.faultDevInfo.HeartbeatTime
}

// GetHeartbeatInterval return heartbeat interval
func (f *FaultTools) GetHeartbeatInterval() int {
	if f.faultDevInfo == nil {
		hwlog.RunLog.Error("fault dev info is nil when get heartbeat interval")
		return -1
	}
	return f.faultDevInfo.HeartbeatInterval
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
func (f *FaultTools) SetFaultDevInfo(faultDevInfo *FaultDevInfo) {
	if f.faultDevInfo == nil {
		hwlog.RunLog.Error("fault dev info is nil when set fault device info")
		return
	}
	f.SetFaultDevList(faultDevInfo.FaultDevList)
	f.SetHeartbeatTime(faultDevInfo.HeartbeatTime)
	f.SetHeartbeatInterval(faultDevInfo.HeartbeatInterval)
	f.SetNodeStatus(faultDevInfo.NodeStatus)
}

// SetFaultDevList set fault device list
func (f *FaultTools) SetFaultDevList(faultDevList []*FaultDev) {
	if f.faultDevInfo == nil {
		hwlog.RunLog.Error("fault dev info is nil when set fault device list")
		return
	}
	f.faultDevInfo.FaultDevList = make([]*FaultDev, len(faultDevList))
	for i, faultDev := range faultDevList {
		faultDevTmp := &FaultDev{}
		f.DeepCopyFaultDev(faultDevTmp, faultDev)
		f.faultDevInfo.FaultDevList[i] = faultDevTmp
	}
}

// SetHeartbeatTime set heartbeat time
func (f *FaultTools) SetHeartbeatTime(heartbeatTime int64) {
	if f.faultDevInfo == nil {
		hwlog.RunLog.Error("fault dev info is nil when set heartbeat time")
		return
	}
	f.faultDevInfo.HeartbeatTime = heartbeatTime
}

// SetHeartbeatInterval set heartbeat interval
func (f *FaultTools) SetHeartbeatInterval(heartbeatInterval int) {
	if f.faultDevInfo == nil {
		hwlog.RunLog.Error("fault dev info is nil when set heartbeat interval")
		return
	}
	f.faultDevInfo.HeartbeatInterval = heartbeatInterval
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
func (f *FaultTools) DeepCopyFaultDev(oldFaultDev, newFaultDev *FaultDev) {
	oldFaultDev.FaultCode = CopyStringSlice(newFaultDev.FaultCode)
	oldFaultDev.DeviceType = strings.Clone(newFaultDev.DeviceType)
	oldFaultDev.DeviceId = newFaultDev.DeviceId
	oldFaultDev.FaultLevel = strings.Clone(newFaultDev.FaultLevel)
}
