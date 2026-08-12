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

// Package device a series of device function
package device

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager"
)

const (
	maxCpuListFileSize   = 10 * 1024
	maxNumaNodeFileSize  = 10 * 1024
	maxNumaNodePathCount = 10 * 1024
)

var (
	numaMgr *NumaNodeManager
)

// NumaNodeManager manages NUMA node discovery and caching for NPU devices.
type NumaNodeManager struct {
	dmgr                devmanager.DeviceInterface
	cpuListToNumaNode   map[string]int64
	cpuListToNumaNodeMu sync.RWMutex
	phyIDToNumaNodes    map[int32][]int64
	phyIDToNumaNodesMu  sync.RWMutex
}

// NewNumaNodeManager creates a new NumaNodeManager.
func NewNumaNodeManager(dmgr devmanager.DeviceInterface) *NumaNodeManager {
	return &NumaNodeManager{
		dmgr:              dmgr,
		cpuListToNumaNode: make(map[string]int64, common.GeneralMapSize),
		phyIDToNumaNodes:  make(map[int32][]int64, common.GeneralMapSize),
	}
}

// GetNumaNodesByPhyID returns cached numa nodes for the given phyID.
func GetNumaNodesByPhyID(phyID int32) []int64 {
	return numaMgr.getNumaNodesByPhyID(phyID)
}

// getNumaNodesByPhyID returns cached numa nodes for the given phyID.
func (n *NumaNodeManager) getNumaNodesByPhyID(phyID int32) []int64 {
	n.phyIDToNumaNodesMu.RLock()
	defer n.phyIDToNumaNodesMu.RUnlock()
	return n.phyIDToNumaNodes[phyID]
}

func validateSysfsPath(resolvedPath string) bool {
	return strings.HasPrefix(resolvedPath, "/sys/")
}

// buildCpuListToNumaNodeMap builds the cpuListToNumaNode map from sysfs numa node paths.
func (n *NumaNodeManager) buildCpuListToNumaNodeMap() error {
	nodePaths, err := filepath.Glob("/sys/devices/system/node/node*")
	if err != nil {
		hwlog.RunLog.Errorf("get numa node failed, err: %v", err)
		return err
	}
	nodeCnt := len(nodePaths)
	if nodeCnt == 0 {
		hwlog.RunLog.Warnf("the numa is not supported, nodes path count is zero")
		return nil
	}

	if nodeCnt > maxNumaNodePathCount {
		hwlog.RunLog.Warnf("there are too many numa node paths, which is %d should be in range [0,%d]",
			nodeCnt, maxNumaNodePathCount)
		return errors.New("too many numa node paths")
	}

	n.cpuListToNumaNodeMu.Lock()
	defer n.cpuListToNumaNodeMu.Unlock()
	for _, nodePath := range nodePaths {
		// "/sys/devices/system/node/node0" -> 0
		nodeIDStr := strings.TrimPrefix(nodePath, "/sys/devices/system/node/node")
		nodeID, errConv := strconv.Atoi(nodeIDStr)
		if errConv != nil {
			hwlog.RunLog.Warnf("parse numa node id from node path failed, err: %v", errConv)
			continue
		}
		cpuListPath := filepath.Join(nodePath, "cpulist")
		cpuData, errRead := utils.ReadLimitBytes(cpuListPath, maxCpuListFileSize)
		if errRead != nil {
			hwlog.RunLog.Warnf("read cpulist file info failed, err: %v", errRead)
			continue
		}
		cpuList := strings.TrimSpace(string(cpuData))
		n.cpuListToNumaNode[cpuList] = int64(nodeID)
	}
	hwlog.RunLog.Infof("cpuListToNumaNodeMap: %v", n.cpuListToNumaNode)
	return nil
}

// InitAllNumaNodeInfo builds cpuListToNumaNodeMap from sysfs, then discovers and caches
// NUMA nodes for all given devices, and logs the result.
func (n *NumaNodeManager) InitAllNumaNodeInfo(allInfo *common.NpuAllInfo) error {
	if err := n.buildCpuListToNumaNodeMap(); err != nil {
		return err
	}

	for i := range allInfo.AllDevs {
		n.setNumaNodes(&allInfo.AllDevs[i])
	}

	for _, dev := range allInfo.AllDevs {
		hwlog.RunLog.Infof("npu name: %s, npu id: %d, npu phyId: %d, numa nodes : %v",
			dev.DeviceName, dev.LogicID, dev.PhyID, n.getNumaNodesByPhyID(dev.PhyID))
	}
	for _, dev := range allInfo.AICoreDevs {
		hwlog.RunLog.Infof("npu name: %s, npu id: %d, npu phyId: %d, numa nodes : %v",
			dev.DeviceName, dev.LogicID, dev.PhyID, n.getNumaNodesByPhyID(dev.PhyID))
	}
	return nil
}

// setNumaNodes discovers and caches the NUMA node for the given device.
func (n *NumaNodeManager) setNumaNodes(dev *common.NpuDevice) {
	n.phyIDToNumaNodesMu.RLock()
	_, ok := n.phyIDToNumaNodes[dev.PhyID]
	n.phyIDToNumaNodesMu.RUnlock()
	if ok {
		return
	}

	err := n.setNumaNodesByAffinityCpu(dev)
	if err == nil {
		return
	}
	hwlog.RunLog.Warnf("set numa nodes by AffinityCpuInfo failed, err: %s, try set by pcie bus info", err)
	pcieBusIdStr, errPcie := n.dmgr.GetPCIeBusInfo(dev.LogicID)
	if errPcie != nil {
		hwlog.RunLog.Warnf("get pcie bus info failed, err: %s", errPcie)
		return
	}
	pcieBusIdStr = strings.ToLower(pcieBusIdStr)
	nodePath := fmt.Sprintf("/sys/bus/pci/devices/%s/numa_node", pcieBusIdStr)
	nodeIdBytes, err := utils.ReadLimitBytesWithSymlink(nodePath, maxNumaNodeFileSize, validateSysfsPath)
	if err != nil {
		hwlog.RunLog.Warnf("read pcie numa_node file %s failed, err: %s", nodePath, err)
		return
	}
	nodeIdStr := strings.TrimSpace(string(nodeIdBytes))
	nodeId, err := strconv.Atoi(nodeIdStr)
	if err != nil {
		hwlog.RunLog.Errorf("convert numa node id failed, err: %s", err)
		return
	}
	if nodeId < 0 {
		hwlog.RunLog.Warnf("numa node id %d is invalid", nodeId)
		return
	}
	n.phyIDToNumaNodesMu.Lock()
	n.phyIDToNumaNodes[dev.PhyID] = []int64{int64(nodeId)}
	n.phyIDToNumaNodesMu.Unlock()
}

func (n *NumaNodeManager) setNumaNodesByAffinityCpu(dev *common.NpuDevice) error {
	cpuList, errCpu := n.dmgr.GetAffinityCpuInfo(dev.LogicID)
	if errCpu != nil {
		hwlog.RunLog.Errorf("get affinity cpu info for npu %d failed, error: %v", dev.LogicID, errCpu)
		return errCpu
	}
	n.cpuListToNumaNodeMu.RLock()
	numaNode, ok := n.cpuListToNumaNode[cpuList]
	n.cpuListToNumaNodeMu.RUnlock()
	if !ok {
		errNotFond := fmt.Errorf("no numa node for cpu list %s for logicID(%d)", cpuList, dev.LogicID)
		hwlog.RunLog.Warn(errNotFond)
		return errNotFond
	}
	n.phyIDToNumaNodesMu.Lock()
	n.phyIDToNumaNodes[dev.PhyID] = []int64{numaNode}
	n.phyIDToNumaNodesMu.Unlock()
	return nil
}

// InitAllNumaNodeInfo builds cpuListToNumaNodeMap from sysfs, then discovers and caches
// NUMA nodes for all given devices, and logs the result.
func InitAllNumaNodeInfo(dmgr devmanager.DeviceInterface, allInfo *common.NpuAllInfo) error {
	if numaMgr == nil {
		numaMgr = NewNumaNodeManager(dmgr)
	}
	return numaMgr.InitAllNumaNodeInfo(allInfo)
}
