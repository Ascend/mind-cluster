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

// Package dpucontrol is used for find dpu.
package dpucontrol

import (
	"os"
)

const (
	dpuSlotIdx1                 = "1"
	dpuSlotIdx2                 = "2"
	dpuSlotIdx9                 = "9"
	dpuSlotIdx10                = "10"
	dpuIpAddrsLen               = 2
	npuIdxCorrespDpuRangeMiddle = 4
	busTypeUb                   = "ub"
	busTypePcie                 = "pcie"
	deviceDir                   = "device"
	deviceFile                  = "device"
	vendorFile                  = "vendor"
	slotIdFile                  = "slot_id"
	pcieSwitchDir               = "/sys/bus/pci/devices"
	// netPath dpu file path
	netPath = "/sys/class/net"
	// dpuConfigPath dpu config file path
	dpuConfigPath = "/user/mindx-dl/dpu/dpu-config.json"
	// dpuIndexFir third party dpu first card
	dpuIndexFir = 0
	// dpuIndexSec third party dpu second card
	dpuIndexSec = 1
	// pcieDirLen pcie head dir length
	pcieDirLen = 4
	// onlyOneDpu only have one third party dpu
	onlyOneDpu = 1
)

// DpuFilter dpu filter info
type DpuFilter struct {
	NpuWithDpuInfos []NpuWithDpuInfo
	userConfig      UserDpuConfig
	entries         []os.DirEntry
	dpuInfos        []BaseDpuInfo
}

// BaseDpuInfo base dpu info struct
type BaseDpuInfo struct {
	Operstate  string
	DeviceName string
	DpuIP      string
	Vendor     string
	DeviceId   string
}

// DeviceSelectors user dpu config selector
type DeviceSelectors struct {
	Vendor      []string `json:"vendor"`
	DeviceIds   []string `json:"deviceIds"`
	DeviceNames []string `json:"devices"`
}

// UserDpuConfig user dpu config file configList struct
type UserDpuConfig struct {
	BusType   string           `json:"busType"`
	Selectors *DeviceSelectors `json:"selectors"`
}

// ConfigList user dpu config file struct
type ConfigList struct {
	UserDpuConfigList []UserDpuConfig `json:"configList"`
}

// NpuWithDpuInfo npu correspond dpu infos
type NpuWithDpuInfo struct {
	NpuId   int32
	DpuInfo []BaseDpuInfo
}
