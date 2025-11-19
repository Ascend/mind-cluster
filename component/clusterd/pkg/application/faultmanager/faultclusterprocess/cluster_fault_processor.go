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

// Package faultclusterprocess is used to return cluster faults
package faultclusterprocess

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/cmprocess"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/interface/grpc/fault"
	"clusterd/pkg/interface/kube"
)

// ClusterFaultCenter used to cache cluster fault
var ClusterFaultCenter *ClusterFaultProcessor

// faultLevelMap to indicate fault level num
var faultLevelMap = make(map[string]int)

const (
	clusterFaultCacheInterval = 3
	healthyLevelNum           = 0
	subHealthyLevelNum        = 1
)

func init() {
	ClusterFaultCenter = newClusterFaultProcessor()
	faultLevelMap[constant.NotHandleFault] = healthyLevelNum
	faultLevelMap[constant.SubHealthFault] = subHealthyLevelNum
}

func newClusterFaultProcessor() *ClusterFaultProcessor {
	return &ClusterFaultProcessor{
		ClusterFaultCache: &fault.FaultMsgSignal{
			NodeFaultInfo: make([]*fault.NodeFaultInfo, 0),
		},
	}
}

// ClusterFaultProcessor is the processor to deal with cluster level fault
type ClusterFaultProcessor struct {
	lock              sync.Mutex
	ClusterFaultCache *fault.FaultMsgSignal
	lastUpdateTime    time.Time
}

// GatherClusterFaultInfo main func of this processor
func (cfp *ClusterFaultProcessor) GatherClusterFaultInfo() *fault.FaultMsgSignal {
	cfp.lock.Lock()
	hwlog.RunLog.Infof("%v", time.Now().Sub(cfp.lastUpdateTime))
	if time.Now().Sub(cfp.lastUpdateTime) < clusterFaultCacheInterval*time.Second {
		cfp.lock.Unlock()
		hwlog.RunLog.Infof("cache is whitin %d seconds", clusterFaultCacheInterval)
		return cfp.ClusterFaultCache
	}
	cfp.lock.Unlock()

	content := constant.AllConfigmapContent{
		DeviceCm: cmprocess.DeviceCenter.GetProcessedCm(),
		SwitchCm: cmprocess.SwitchCenter.GetProcessedCm(),
		NodeCm:   cmprocess.NodeCenter.GetProcessedCm(),
	}
	clusterNodeFaults := getAllKindsFaults(content)

	cfp.lock.Lock()
	defer cfp.lock.Unlock()
	cfp.ClusterFaultCache = &fault.FaultMsgSignal{
		NodeFaultInfo: make([]*fault.NodeFaultInfo, 0),
	}
	cfp.ClusterFaultCache.NodeFaultInfo = clusterNodeFaults
	cfp.ClusterFaultCache.SignalType = constant.SignalTypeNormal
	if len(cfp.ClusterFaultCache.NodeFaultInfo) > 0 {
		cfp.ClusterFaultCache.SignalType = constant.SignalTypeFault
	}
	cfp.lastUpdateTime = time.Now()
	return cfp.ClusterFaultCache
}

func getAllKindsFaults(content constant.AllConfigmapContent) []*fault.NodeFaultInfo {
	clusterNodeFaults := make([]*fault.NodeFaultInfo, 0)
	for cmName, deviceinfo := range content.DeviceCm {
		nodeName := strings.TrimPrefix(cmName, constant.DeviceInfoPrefix)
		NodeFaultInfo := fault.NodeFaultInfo{
			NodeName: nodeName,
			NodeIP:   node.GetNodeIpByName(nodeName),
			NodeSN:   node.GetNodeSNByName(nodeName),
		}
		npuFaults := getNpuDeviceFaultInfo(deviceinfo)
		if npuFaults != nil {
			NodeFaultInfo.FaultDevice = append(NodeFaultInfo.FaultDevice, npuFaults...)
		}
		switchFaults := getSwitchFaultInfo(content.SwitchCm[constant.SwitchInfoPrefix+nodeName])
		if switchFaults != nil {
			NodeFaultInfo.FaultDevice = append(NodeFaultInfo.FaultDevice, switchFaults...)
		}
		nodeFaults := getNodeFaultInfo(content.NodeCm[constant.NodeInfoPrefix+nodeName])
		if nodeFaults != nil {
			NodeFaultInfo.FaultDevice = append(NodeFaultInfo.FaultDevice, nodeFaults...)
		}
		nodeReadyFault := getNodeReadyFaultInfo(nodeName)
		if nodeReadyFault != nil {
			NodeFaultInfo.FaultDevice = append(NodeFaultInfo.FaultDevice, nodeReadyFault)
		}
		if len(NodeFaultInfo.FaultDevice) == 0 {
			continue
		}
		NodeFaultInfo.FaultLevel = getMaxLevel(NodeFaultInfo)
		clusterNodeFaults = append(clusterNodeFaults, &NodeFaultInfo)
	}
	return clusterNodeFaults
}

func getMaxLevel(NodeFaultInfo fault.NodeFaultInfo) string {
	maxFaultLevel := constant.HealthyState
	for _, faultdev := range NodeFaultInfo.FaultDevice {
		if faultdev.FaultLevel == constant.UnHealthyState {
			maxFaultLevel = constant.UnHealthyState
			break
		}
		if faultdev.FaultLevel == constant.SubHealthyState {
			maxFaultLevel = constant.SubHealthyState
		}
	}
	return maxFaultLevel
}

func getNodeReadyFaultInfo(nodeName string) *fault.DeviceFaultInfo {
	nodeStatus := kube.GetNode(nodeName)
	if nodeStatus == nil || !faultdomain.IsNodeReady(nodeStatus) {
		return &fault.DeviceFaultInfo{
			DeviceId:   constant.EmptyDeviceId,
			DeviceType: constant.FaultTypeNode,
			FaultCodes: nil,
			FaultLevel: constant.UnHealthyState,
		}
	}
	return nil
}

func getNodeFaultInfo(nodeInfo *constant.NodeInfo) []*fault.DeviceFaultInfo {
	if nodeInfo == nil {
		return nil
	}
	faultsOnNode := make([]*fault.DeviceFaultInfo, 0)
	for _, device := range nodeInfo.FaultDevList {
		deviceFault := fault.DeviceFaultInfo{
			DeviceId:   strconv.Itoa(int(device.DeviceId)),
			DeviceType: device.DeviceType,
			FaultCodes: device.FaultCode,
		}
		switch nodeInfo.NodeStatus {
		case constant.NotHandleFaultLevelStr:
			deviceFault.FaultLevel = constant.HealthyState
		case constant.SubHealthFaultLevelStr:
			deviceFault.FaultLevel = constant.SubHealthyState
		case constant.SeparateFaultLevelStr:
			deviceFault.FaultLevel = constant.UnHealthyState
		case constant.PreSeparateFaultLevelStr:
			deviceFault.FaultLevel = constant.PreSeparateState
		default:
			deviceFault.FaultLevel = constant.HealthyState
		}
		faultsOnNode = append(faultsOnNode, &deviceFault)
	}
	return faultsOnNode
}

func getSwitchFaultInfo(switchInfo *constant.SwitchInfo) []*fault.DeviceFaultInfo {
	if switchInfo == nil {
		return nil
	}
	allFault := fault.DeviceFaultInfo{
		DeviceId:   constant.EmptyDeviceId,
		DeviceType: constant.FaultTypeSwitch,
	}
	faultCodes := make([]string, 0, len(switchInfo.SwitchFaultInfo.FaultInfo))
	switchFaultInfos := make([]*fault.SwitchFaultInfo, 0, len(switchInfo.SwitchFaultInfo.FaultInfo))
	for _, switchFault := range switchInfo.SwitchFaultInfo.FaultInfo {
		faultCodes = append(faultCodes, switchFault.AssembledFaultCode)
		switchChipId := strconv.Itoa(int(switchFault.SwitchChipId))
		switchPortId := strconv.Itoa(int(switchFault.SwitchPortId))
		faultTimeAndLevelKey := switchFault.AssembledFaultCode + "_" + switchChipId + "_" + switchPortId
		levelInfo, exists := switchInfo.FaultTimeAndLevelMap[faultTimeAndLevelKey]
		if !exists {
			levelInfo = constant.FaultTimeAndLevel{}
		}
		switchFaultInfos = append(switchFaultInfos, &fault.SwitchFaultInfo{
			FaultCode:    switchFault.AssembledFaultCode,
			SwitchChipId: switchChipId,
			SwitchPortId: switchPortId,
			FaultTime:    strconv.Itoa(int(switchFault.AlarmRaisedTime)),
			FaultLevel:   levelInfo.FaultLevel,
		})
	}
	allFault.FaultCodes = faultCodes
	allFault.FaultLevel = switchInfo.NodeStatus
	allFault.SwitchFaultInfos = switchFaultInfos
	return []*fault.DeviceFaultInfo{&allFault}
}

func getNpuDeviceFaultInfo(deviceCm *constant.AdvanceDeviceFaultCm) []*fault.DeviceFaultInfo {
	if deviceCm == nil {
		return nil
	}
	faultsOnDevice := make([]*fault.DeviceFaultInfo, 0)
	for deviceName, deviceFaults := range deviceCm.FaultDeviceList {
		deviceId := constant.EmptyDeviceId
		if len(strings.Split(deviceName, "-")) > 1 {
			deviceId = strings.Split(deviceName, "-")[1]
		}
		deviceFault := fault.DeviceFaultInfo{
			DeviceId:    deviceId,
			DeviceType:  constant.FaultTypeNPU,
			FaultLevel:  constant.HealthyState,
			FaultReason: nil,
		}
		deviceFaultCodes, deviceFaultLevels, maxFaultLevel := getDeviceFaultInfo(deviceFaults)
		deviceFault.FaultCodes = deviceFaultCodes
		deviceFault.FaultLevel = maxFaultLevel
		deviceFault.FaultLevels = deviceFaultLevels
		faultsOnDevice = append(faultsOnDevice, &deviceFault)
	}
	return faultsOnDevice
}

func getDeviceFaultInfo(deviceFaults []constant.DeviceFault) ([]string, []string, string) {
	maxFaultLevel := constant.HealthyState
	faultCodes := make([]string, 0, len(deviceFaults))
	faultLevels := make([]string, 0, len(deviceFaults))
	for _, faultMsg := range deviceFaults {
		faultCodes = append(faultCodes, faultMsg.FaultCode)
		faultLevels = append(faultLevels, faultMsg.FaultTimeAndLevelMap[faultMsg.FaultCode].FaultLevel)
		faultLevel, ok := faultLevelMap[faultMsg.FaultLevel]
		if !ok {
			maxFaultLevel = constant.UnHealthyState
		}
		if faultLevel == subHealthyLevelNum {
			maxFaultLevel = constant.SubHealthyState
		}
	}
	return faultCodes, faultLevels, maxFaultLevel
}
