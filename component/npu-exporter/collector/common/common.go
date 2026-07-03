/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common for general constants
package common

import (
	"fmt"
	"sort"

	"ascend-common/api"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/hccn"
	"huawei.com/npu-exporter/v6/utils/logger"
)

// Init init npu total ports num
func (e *NpuDevPortsInfo) Init() {
	totalPorts := 0
	for _, v := range e.devPortMap {
		totalPorts += len(v)
	}
	e.totalPort = totalPorts
}

// GetCount get npu total ports
func (e *NpuDevPortsInfo) GetCount() int {
	return e.totalPort
}

// GetPortMap get npu ports info
func (e *NpuDevPortsInfo) GetPortMap() map[int][]common.NpuDevPortInfo {
	return e.devPortMap
}

// SetPortMap init set npu ports info
func (e *NpuDevPortsInfo) SetPortMap(devMap map[int][]common.NpuDevPortInfo) {
	// Sort port list for each die to ensure consistent order
	for _, ports := range devMap {
		sort.Slice(ports, func(i, j int) bool {
			return ports[i].PortID < ports[j].PortID
		})
	}
	e.devPortMap = devMap
}

func getNpuDevNetPortInfos(n *NpuCollector) error {
	_, npuList, err := n.Dmgr.GetDeviceList()
	if err != nil {
		return fmt.Errorf("failed to detect any NPU")
	}
	isGetPortInfo := false
	for _, logicID := range npuList {
		devInfo, err := hccn.GetNpuDevNetPortInfo(logicID)
		if err != nil {
			continue
		}
		NpuDevPortInfos.SetPortMap(devInfo)
		isGetPortInfo = true
		break
	}
	if !isGetPortInfo {
		return fmt.Errorf("failed to detect any queryable NPU")
	}
	NpuDevPortInfos.Init()
	return nil
}

// InitNpuDevNetPortInfos init npu net port infos
func InitNpuDevNetPortInfos(n *NpuCollector) {
	DevType = n.Dmgr.GetDevType()
	if DevType != api.Ascend910A5 {
		return
	}
	err := getNpuDevNetPortInfos(n)
	if err != nil {
		logger.Errorf("getNpuDevNetPortInfos failed, %v", err)
	}
}
