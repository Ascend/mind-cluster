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
	for _, diePortMap := range e.devPortMap {
		for _, ports := range diePortMap {
			totalPorts += len(ports)
		}
	}
	e.totalPort = totalPorts
	logger.Infof("[NpuDevPortInfos] Init succeeded, totalPort=%d", totalPorts)
}

// GetCount get npu total ports
func (e *NpuDevPortsInfo) GetCount() int {
	return e.totalPort
}

// GetPortMap get npu ports info for the specified logicID
func (e *NpuDevPortsInfo) GetPortMap(logicID int32) map[int][]common.NpuDevPortInfo {
	return e.devPortMap[logicID]
}

// GetMergedPortMap returns the union of dieID->ports across all logicIDs.
// Same-type chips share identical die/port structure, so this merged view is
// used to build legacy metric descriptors which do not carry a logicID.
func (e *NpuDevPortsInfo) GetMergedPortMap() map[int][]common.NpuDevPortInfo {
	merged := make(map[int][]common.NpuDevPortInfo)
	for _, diePortMap := range e.devPortMap {
		for dieID, ports := range diePortMap {
			merged[dieID] = ports
		}
	}
	return merged
}

// SetPortMap init set npu ports info for the specified logicID
func (e *NpuDevPortsInfo) SetPortMap(logicID int32, devMap map[int][]common.NpuDevPortInfo) {
	if e.devPortMap == nil {
		e.devPortMap = make(map[int32]map[int][]common.NpuDevPortInfo)
	}
	// Sort port list for each die to ensure consistent order
	for dieID, ports := range devMap {
		sort.Slice(ports, func(i, j int) bool {
			return ports[i].PortID < ports[j].PortID
		})
		portIDs := make([]int, 0, len(ports))
		for _, port := range ports {
			portIDs = append(portIDs, port.PortID)
		}
		logger.Infof("[NpuDevPortInfos] set port map, logicID=%d, dieID=%d, portIDs=%v", logicID, dieID, portIDs)
	}
	e.devPortMap[logicID] = devMap
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
			logger.Warnf("[NpuDevPortInfos] get port info for logicID=%d failed: %v", logicID, err)
			continue
		}
		NpuDevPortInfos.SetPortMap(logicID, devInfo)
		isGetPortInfo = true
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
