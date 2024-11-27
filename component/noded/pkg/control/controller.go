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

// Package control for fault handling
package control

import (
	"fmt"
	"sync"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
)

// NodeController process fault device info based on configuration
type NodeController struct {
	kubeClient          *kubeclient.ClientK8s
	faultManager        common.FaultManager
	configManager       common.ConfigManager
	nextConfigProcessor common.ConfigProcessor
	nextFaultProcessor  common.FaultProcessor
	faultLevelMap       map[string]int
	faultLevelMapLock   *sync.Mutex
}

// NewNodeController create a node controller
func NewNodeController(client *kubeclient.ClientK8s) *NodeController {
	return &NodeController{
		kubeClient:        client,
		faultManager:      common.NewFaultManager(),
		configManager:     common.NewConfigManager(),
		faultLevelMap:     map[string]int{},
		faultLevelMapLock: &sync.Mutex{},
	}
}

// Init initialize node controller
func (nc *NodeController) Init() error {
	return nil
}

// Execute process fault device info and send message to next fault processor
func (nc *NodeController) Execute(faultDevInfo *common.FaultDevInfo) {
	nc.faultManager.SetFaultDevInfo(faultDevInfo)
	nc.updateFaultDevInfo()
	nc.nextFaultProcessor.Execute(nc.faultManager.GetFaultDevInfo())
}

// SetNextFaultProcessor set the next fault processor
func (nc *NodeController) SetNextFaultProcessor(faultProcessor common.FaultProcessor) {
	nc.nextFaultProcessor = faultProcessor
}

// UpdateConfig update config and update fault level map
func (nc *NodeController) UpdateConfig(faultConfig *common.FaultConfig) error {
	if err := nc.updateFaultLevelMap(faultConfig.FaultTypeCode); err != nil {
		hwlog.RunLog.Errorf("update fault level map failed, err is %v", err)
		return err
	}
	nc.configManager.SetFaultConfig(faultConfig)
	return nil
}

// SetNextConfigProcessor set next config processor
func (nc *NodeController) SetNextConfigProcessor(configProcessor common.ConfigProcessor) {
	nc.nextConfigProcessor = configProcessor
}

// updateFaultDevInfo update fault device info
func (nc *NodeController) updateFaultDevInfo() {
	// filter not support fault device
	nc.filterNotSupportFaultDev()

	faultDevInfo := nc.faultManager.GetFaultDevInfo()
	var nodeFaultLevel int64
	// update fault level
	for _, faultDev := range faultDevInfo.FaultDevList {
		faultLevelStr, faultLevelInt := nc.getFaultLevel(faultDev.FaultCode)
		faultDev.FaultLevel = faultLevelStr
		if faultLevelInt > nodeFaultLevel {
			nodeFaultLevel = faultLevelInt
		}
	}
	// update node status
	nc.faultManager.SetNodeStatus(nc.getNodeStatus(nodeFaultLevel))
}

// filterNotSupportFaultDev filter not support fault devs
func (nc *NodeController) filterNotSupportFaultDev() {
	faultDevInfo := nc.faultManager.GetFaultDevInfo()
	newFaultDevInfo := &common.FaultDevInfo{
		FaultDevList:      make([]*common.FaultDev, 0),
		HeartbeatTime:     faultDevInfo.HeartbeatTime,
		HeartbeatInterval: faultDevInfo.HeartbeatInterval,
		NodeStatus:        faultDevInfo.NodeStatus,
	}
	for _, faultDev := range faultDevInfo.FaultDevList {
		tmpFaultDev := &common.FaultDev{
			DeviceType: faultDev.DeviceType,
			DeviceId:   faultDev.DeviceId,
			FaultCode:  nc.filterNotSupportFaultCodes(faultDev.FaultCode),
			FaultLevel: faultDev.FaultLevel,
		}
		if len(tmpFaultDev.FaultCode) > 0 {
			newFaultDevInfo.FaultDevList = append(newFaultDevInfo.FaultDevList, tmpFaultDev)
		}
	}
	nc.faultManager.SetFaultDevInfo(newFaultDevInfo)
}

// filterNotSupportFaultCodes filter not support fault codes
func (nc *NodeController) filterNotSupportFaultCodes(faultCodes []string) []string {
	newFaultCodes := make([]string, 0)
	for _, faultCode := range faultCodes {
		if _, ok := nc.faultLevelMap[faultCode]; ok {
			newFaultCodes = append(newFaultCodes, faultCode)
		}
	}
	return newFaultCodes
}

// updateFaultLevelMap update fault level map when update config
func (nc *NodeController) updateFaultLevelMap(faultTypeCode *common.FaultTypeCode) error {
	newFaultLevelMap := make(map[string]int, 0)
	for _, notHandleFaultCode := range faultTypeCode.NotHandleFaultCodes {
		if _, ok := newFaultLevelMap[notHandleFaultCode]; ok {
			hwlog.RunLog.Errorf("update not handle fault code failed, code %s is conflict", notHandleFaultCode)
			return fmt.Errorf("not handle code %s is conflict, "+
				"please check whether the code not just configured at not handle level", notHandleFaultCode)
		}
		newFaultLevelMap[notHandleFaultCode] = common.NotHandleFaultLevel
	}
	for _, preSeparateFaultCode := range faultTypeCode.PreSeparateFaultCodes {
		if _, ok := newFaultLevelMap[preSeparateFaultCode]; ok {
			hwlog.RunLog.Errorf("update pre separate fault code failed, "+
				"code %s is conflict", preSeparateFaultCode)
			return fmt.Errorf("pre separate code %s is conflict, "+
				"please check whether the code not just configured at pre separate level", preSeparateFaultCode)
		}
		newFaultLevelMap[preSeparateFaultCode] = common.PreSeparateFaultLevel
	}
	for _, separateFaultCode := range faultTypeCode.SeparateFaultCodes {
		if _, ok := newFaultLevelMap[separateFaultCode]; ok {
			hwlog.RunLog.Errorf("update separate fault code failed, "+
				"code %s is conflict", separateFaultCode)
			return fmt.Errorf("separate fault code %s is conflict, "+
				"please check whether the code not just configured at separate level", separateFaultCode)
		}
		newFaultLevelMap[separateFaultCode] = common.SeparateFaultLevel
	}
	nc.faultLevelMapLock.Lock()
	nc.faultLevelMap = newFaultLevelMap
	nc.faultLevelMapLock.Unlock()
	hwlog.RunLog.Debugf("new fault level map is %v", newFaultLevelMap)
	return nil
}

// getFaultLevel get fault level
func (nc *NodeController) getFaultLevel(faultCodes []string) (string, int64) {
	maxLevel := 0
	for _, faultCode := range faultCodes {
		if level, ok := nc.faultLevelMap[faultCode]; ok {
			if level > maxLevel {
				maxLevel = level
			}
		}
	}
	switch maxLevel {
	case common.PreSeparateFaultLevel:
		return common.PreSeparateFault, common.PreSeparateFaultLevel
	case common.SeparateFaultLevel:
		return common.SeparateFault, common.SeparateFaultLevel
	default:
		return common.NotHandleFault, common.NotHandleFaultLevel
	}
}

// getNodeStatus get node status
func (nc *NodeController) getNodeStatus(nodeFaultLevel int64) string {
	switch nodeFaultLevel {
	case common.NodeUnHealthyLevel:
		return common.NodeUnHealthy
	case common.NodeSubHealthyLevel:
		return common.PreSeparate
	case common.NodeHealthyLevel:
		return common.NodeHealthy
	default:
		return ""
	}
}
