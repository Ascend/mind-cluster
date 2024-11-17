/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package rescheduling

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// createFaultCardHandlers initialise FaultCard struct == getInoperableNPUCards
func (fNode *FaultNode) createFaultCardHandlers(node *plugin.NPUNode) ([]FaultCard, error) {
	klog.V(util.LogInfoLev).Infof("create new fault card handlers for node %s", node.Name)
	var faultCards []FaultCard
	for _, card := range fNode.AllCards {
		faultCard := FaultCard{
			IsFaultCard: false,
			NPUName:     card,
			NodeName:    node.Name,
			FaultType:   CardHealthy,
		}

		if faultCard.isCardUnhealthy(fNode.UnhealthyNPU) {
			klog.V(util.LogDebugLev).Infof("card %s is unhealthy", faultCard.NPUName)
			faultCard.setIsFaultCard(true)
			faultCard.setFaultType(CardUnhealthy)
			faultCards = append(faultCards, faultCard)
			continue
		}
		if faultCard.isCardNetworkUnhealthy(fNode.NetworkUnhealthyNPU) {
			klog.V(util.LogDebugLev).Infof("card %s is network unhealthy", faultCard.NPUName)
			faultCard.setIsFaultCard(true)
			faultCard.setFaultType(CardNetworkUnhealthy)
			faultCards = append(faultCards, faultCard)
			continue
		}
		faultCards = append(faultCards, faultCard)
	}

	return faultCards, nil
}

// getNodeNPUsByKey get the npu list from node.DeviceInfo
func (fNode *FaultNode) getNodeNPUsByKey(node *plugin.NPUNode, deviceKey string) ([]string, error) {
	npuStr, ok := node.Annotation[deviceKey]
	if !ok || len(npuStr) == 0 {
		return nil, fmt.Errorf("%s get nil npus", node.Name)
	}
	npus := strings.Split(npuStr, ",")

	return npus, nil
}

func (fNode *FaultNode) getNodeHeartbeatByKey(node *plugin.NPUNode, hbKey string) (string, error) {
	intervalStr, ok := node.Annotation[hbKey]
	if !ok || len(intervalStr) == 0 {
		klog.V(util.LogDebugLev).Infof("isNodeHealth %s no [%s].", node.Name, nodeHeartbeat)
		return "", fmt.Errorf("getFaultNodeState %s nil", node.Name)
	}
	return intervalStr, nil
}

// getAllNPUCardsFromDeviceInfo get un-allocated healthy card from device info
func (fNode *FaultNode) getAllNPUCardsFromDeviceInfo(node *plugin.NPUNode, cardName string) ([]string, error) {
	var allCard []string
	healthyCard, err := fNode.getNodeNPUsByKey(node, cardName) // ["Ascend910-0", ...]
	allCard = append(allCard, healthyCard...)
	allCard = append(allCard, fNode.UnhealthyNPU...)
	allCard = append(allCard, fNode.NetworkUnhealthyNPU...)
	allCard = util.RemoveSliceDuplicateElement(allCard)
	if err != nil {
		return allCard, err
	}
	return allCard, nil
}

// getUnhealthyCardsFromDeviceInfo get unhealthyCard from device info
func (fNode *FaultNode) getUnhealthyCardsFromDeviceInfo(node *plugin.NPUNode, cardName string) ([]string, error) {
	unhealthyCardName := fmt.Sprintf("%s-%s", cardName, CardUnhealthy) // ["Ascend910-1"]
	return fNode.getNodeNPUsByKey(node, unhealthyCardName)
}

// getNetworkUnhealthyCardsFromDeviceInfo get networkUnhealthyCard from device info
func (fNode *FaultNode) getNetworkUnhealthyCardsFromDeviceInfo(
	node *plugin.NPUNode, cardName string) ([]string, error) {
	networkUnhealthyCardName := fmt.Sprintf("%s-%s", cardName, CardNetworkUnhealthy) // ["Ascend910-1"]
	return fNode.getNodeNPUsByKey(node, networkUnhealthyCardName)
}

// getNodeHeartbeatIntervalFromDeviceInfo get nodeHeartbeatInterval from device info
func (fNode *FaultNode) getNodeHeartbeatIntervalFromNodeDInfo(node *plugin.NPUNode) (int, error) {
	var heartbeatInterval = nodeUpdateTime
	heartbeatIntervalStr, getErr := fNode.getNodeHeartbeatByKey(node, util.NodeDNodeHeartbeatIntervalKey)
	if getErr != nil {
		return heartbeatInterval, getErr
	}
	var err error
	heartbeatInterval, err = strconv.Atoi(heartbeatIntervalStr)
	if err != nil {
		klog.V(util.LogInfoLev).Infof("%s convert %s to int64 failed [%s].",
			node.Name, heartbeatIntervalStr, util.SafePrint(err))
		return nodeUpdateTime, err
	}

	if heartbeatInterval > maxIntervalTime || heartbeatInterval < 1 {
		klog.V(util.LogInfoLev).Infof("%s's HeartbeatInterval %d over limit, will use %d.",
			node.Name, heartbeatInterval, nodeUpdateTime)
		return nodeUpdateTime, nil
	}
	klog.V(util.LogInfoLev).Infof("%s heartbeatTimeInterval: %d", node.Name, heartbeatInterval)
	return heartbeatInterval, nil
}

// getNodeHeartbeatFromDeviceInfo get nodeHeartbeat from device info
func (fNode *FaultNode) getNodeHeartbeatFromNodeDInfo(node *plugin.NPUNode) (int64, error) {
	heartbeatTimeStr, getErr := fNode.getNodeHeartbeatByKey(node, util.NodedHeartbeatTimeKey)
	if getErr != nil {
		return 0, getErr
	}
	heartbeatTime, err := strconv.ParseInt(heartbeatTimeStr, util.Base10, util.BitSize64)
	if err != nil {
		klog.V(util.LogInfoLev).Infof("%s cover %s to int64 failed [%s].", node.Name, heartbeatTimeStr, util.SafePrint(err))
		return 0, err
	}
	klog.V(util.LogInfoLev).Infof("%s heartbeatTime: %d", node.Name, heartbeatTime)
	return heartbeatTime, nil
}

func (fCard *FaultCard) isCardUnhealthy(unHealthyList []string) bool {
	return util.IsSliceContain(fCard.NPUName, unHealthyList)
}

func (fCard *FaultCard) isCardNetworkUnhealthy(networkUnhealthyList []string) bool {
	return util.IsSliceContain(fCard.NPUName, networkUnhealthyList)
}

func (fNode *FaultNode) updateFaultNodesFromDeviceInfo(node *plugin.NPUNode, cardName string) {
	klog.V(util.LogInfoLev).Infof("update information from device info for node %s", node.Name)

	tmpHBTime, err := fNode.getNodeHeartbeatFromNodeDInfo(node)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("getNodeHeartbeatFromDeviceInfo: %s", util.SafePrint(err))
	}

	klog.V(util.LogDebugLev).Infof(
		"getNodeHeartbeatFromDeviceInfo: former heartbeat time %d, new heartbeat time %d",
		fNode.OldHeartbeatTime, tmpHBTime)
	if fNode.OldHeartbeatTime != tmpHBTime {
		fNode.UpdateHeartbeatTime = time.Now().Unix()
	}
	fNode.setNewNodeHeartbeatTime(tmpHBTime)

	tmpHBIntervalTime, err := fNode.getNodeHeartbeatIntervalFromNodeDInfo(node)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("getNodeHeartbeatIntervalFromDeviceInfo: %s", util.SafePrint(err))
	}
	fNode.setNodeHeartbeatInterval(tmpHBIntervalTime)

	tmpUnhealthyNPUs, err := fNode.getUnhealthyCardsFromDeviceInfo(node, cardName)
	if err != nil {
		klog.V(util.LogInfoLev).Infof("getUnhealthyCardsFromDeviceInfo: %s", util.SafePrint(err))
	}
	fNode.setUnhealthyNPUList(tmpUnhealthyNPUs)
	klog.V(util.LogInfoLev).Infof("Unhealthy cards from device info: %v", tmpUnhealthyNPUs)

	tmpNetworkUnhealthyNPUs, err := fNode.getNetworkUnhealthyCardsFromDeviceInfo(node, cardName)
	if err != nil {
		klog.V(util.LogInfoLev).Infof("getNetworkUnhealthyCardsFromDeviceInfo: %s", util.SafePrint(err))
	}
	fNode.setNetworkUnhealthyNPUList(tmpNetworkUnhealthyNPUs)
	klog.V(util.LogInfoLev).Infof("Network unhealthy cards from device info: %v", tmpUnhealthyNPUs)

	tmpAllCardsList, err := fNode.getAllNPUCardsFromDeviceInfo(node, cardName)
	if err != nil {
		klog.V(util.LogInfoLev).Infof("getAllNPUCardsFromDeviceInfo: %s", util.SafePrint(err))
	}
	fNode.setAllCardList(tmpAllCardsList)
	klog.V(util.LogDebugLev).Infof("Unallocated and fault cards from device info: %v", tmpAllCardsList)
	DeviceFaultReason, err := GetNodeDeviceFaultFromDeviceInfo(node)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("GetNodeDeviceFaultFromDeviceInfo: %s", util.SafePrint(err))
	}
	fNode.setFaultDeviceList(DeviceFaultReason)

}

func GetNodeDeviceFaultFromDeviceInfo(node *plugin.NPUNode) ([]FaultDeviceList, error) {
	deviceFaultList, ok := node.Annotation[DeviceFaultCmKey]
	if !ok {
		return nil, fmt.Errorf("GetNodeDeviceFaultFromDeviceInfo failed")
	}
	var deviceFault []FaultDeviceList
	if unmarshalErr := json.Unmarshal([]byte(deviceFaultList), &deviceFault); unmarshalErr != nil {
		klog.V(util.LogWarningLev).Infof("convertToDeviceFaultListFromCM Unmarshal: %s.", util.SafePrint(unmarshalErr))
		return nil, unmarshalErr
	}
	return deviceFault, nil
}

// updateFaultNodesAttr update Information from device Info
func (fNode *FaultNode) updateFaultNodesAttr(node *plugin.NPUNode) error {
	klog.V(util.LogInfoLev).Infof("Update node %s attributes", node.Name)
	// 1. create fault Card Object
	tmpFaultCards, err := fNode.createFaultCardHandlers(node)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("Getting node card failed: %s", util.SafePrint(err))
		return err
	}
	fNode.setFaultCards(tmpFaultCards)

	fNode.setNodeHealthStateValue(NodeHealthy)
	fNode.setIsFaultNodeValue(false)

	// 2. judge if node is unhealthy by NodeD
	fNode.setNodeHealthyByNodeD(node)
	if fNode.NodeHealthState == NodeUnhealthy {
		return nil
	}

	// 3. set node health state by card unhealthy
	fNode.setNodeHealthyByCardHealth(node)
	return nil
}

func (fNode *FaultNode) setNodeHealthyByNodeD(node *plugin.NPUNode) {
	if !fNode.isNodeDEnabled(node) {
		klog.V(util.LogDebugLev).Infof("node %s nodeD not enabled", node.Name)
		fNode.setNodeDValue(false)
		return
	}
	fNode.setNodeDValue(true)
	// 1. last node heartbeat update time until now being greater than maxInterval indicates unhealthy
	if !fNode.isNodeHealthyByHeartbeat() {
		fNode.setIsFaultNodeValue(true)
		fNode.setNodeHealthStateValue(NodeUnhealthy)
		klog.V(util.LogInfoLev).Infof("Node %s health state set %s for wrong heartbeat", node.Name, NodeUnhealthy)
	}
}

func (fNode *FaultNode) setNodeHealthyByCardHealth(node *plugin.NPUNode) {
	for _, card := range fNode.FaultCards {
		if !card.IsFaultCard {
			continue
		}
		fNode.setIsFaultNodeValue(true)
		switch card.FaultType {
		case CardUnhealthy:
			fNode.setNodeHealthStateValue(NodeCardUnhealthy)
			klog.V(util.LogInfoLev).Infof("Node %s health state set to %s", node.Name, NodeCardUnhealthy)
		case CardNetworkUnhealthy:
			fNode.setNodeHealthStateValue(NodeCardNetworkUnhealthy)
			klog.V(util.LogInfoLev).Infof("Node %s health state set to %s", node.Name, NodeCardNetworkUnhealthy)
		default:
			klog.V(util.LogInfoLev).Infof("card health state %s illegal", card.FaultType)
		}
	}
}

func (fNode *FaultNode) isNodeDEnabled(node *plugin.NPUNode) bool {
	value, ok := node.Label[nodeDEnableKey]
	if !ok {
		return false
	}

	switch value {
	case nodeDEnableOnValue:
		return true
	case nodeDEnableOffValue:
		return false
	default:
		klog.V(util.LogErrorLev).Infof("isEnableFaultNode not support %s.", value)
		return false
	}
}

func (fNode *FaultNode) isNodeHealthyByHeartbeat() bool {
	maxInterval := int64(fNode.HeartbeatInterval) * util.MapInitNum
	nowTime := time.Now().Unix()
	latestInterval := nowTime - fNode.UpdateHeartbeatTime

	klog.V(util.LogDebugLev).Infof(
		"node %s latestInterval: %d, nowTime: %d", fNode.NodeName, latestInterval, nowTime)
	if latestInterval < 0 {
		klog.V(util.LogErrorLev).Infof(" isNodeHealth %s cache Time is newer[%d-%d], confused, skip.",
			fNode.NodeName, nowTime, fNode.UpdateHeartbeatTime)
	}
	if latestInterval > maxInterval {
		klog.V(util.LogErrorLev).Infof(" %s Time over %d [%d-%d],not health,last heartbeat is %d.",
			fNode.NodeName, maxInterval, nowTime, fNode.UpdateHeartbeatTime, fNode.OldHeartbeatTime)
		return false
	}
	return true
}

func (fNode *FaultNode) getFaultCardIds(cardName string) ([]int, error) {
	if fNode.UnhealthyNPU == nil && fNode.NetworkUnhealthyNPU == nil {
		return nil, fmt.Errorf("no fault card on node")
	}
	allFaultCards := append(fNode.UnhealthyNPU, fNode.NetworkUnhealthyNPU...)
	faultCardIds := util.ChangeTopToIntArray(strings.Join(allFaultCards, ","), cardName)
	return faultCardIds, nil
}

// isNodeInSessionByNpuNodes judge if node is sent in session
func (fNode *FaultNode) isNodeInSessionByNpuNodes(nodes map[string]plugin.NPUNode) bool {
	_, ok := nodes[fNode.NodeName]
	return ok
}

func (fNode *FaultNode) setNodeDValue(value bool) {
	fNode.NodeDEnable = value
}

func (fNode *FaultNode) setIsFaultNodeValue(value bool) {
	fNode.IsFaultNode = value
}

func (fNode *FaultNode) setNodeHealthStateValue(nodeHealthState string) {
	fNode.NodeHealthState = nodeHealthState
}

func (fNode *FaultNode) setAllCardList(value []string) {
	fNode.AllCards = value
}

func (fNode *FaultNode) setUnhealthyNPUList(value []string) {
	fNode.UnhealthyNPU = value
}

func (fNode *FaultNode) setNetworkUnhealthyNPUList(value []string) {
	fNode.NetworkUnhealthyNPU = value
}

func (fNode *FaultNode) setUpdateTime(value int64) {
	fNode.UpdateTime = value
}

func (fNode *FaultNode) setFaultCards(value []FaultCard) {
	fNode.FaultCards = value
}

func (fNode *FaultNode) setOldNodeHeartbeatTime(value int64) {
	fNode.OldHeartbeatTime = value
}

func (fNode *FaultNode) setNewNodeHeartbeatTime(value int64) {
	fNode.NewHeartbeatTime = value
}

func (fCard *FaultCard) setFaultType(value string) {
	fCard.FaultType = value
}

func (fCard *FaultCard) setIsFaultCard(value bool) {
	fCard.IsFaultCard = value
}

func (fNode *FaultNode) setNodeHeartbeatInterval(value int) {
	fNode.HeartbeatInterval = value
}

func (fNode *FaultNode) setFaultDeviceList(value []FaultDeviceList) {
	fNode.FaultDeviceList = value
}

func newFaultNodeDefault(nodeName string, updateTime int64) FaultNode {
	faultNode := FaultNode{
		NodeName:            nodeName,
		UpdateTime:          updateTime,
		UnhealthyNPU:        nil,
		NetworkUnhealthyNPU: nil,
		IsFaultNode:         false,
		NodeDEnable:         false,
		NodeHealthState:     NodeHealthy,
		AllCards:            nil,
		FaultCards:          nil,
		HeartbeatInterval:   0,
		OldHeartbeatTime:    0,
		UpdateHeartbeatTime: 0,
		FaultDeviceList:     []FaultDeviceList{},
	}
	return faultNode
}

func isNodeInSessionByNodeName(NodeName string, nodes map[string]plugin.NPUNode) bool {
	_, ok := nodes[NodeName]
	return ok
}

func initSimpleFNodeInfoByFNode(node *FaultNode) SimpleFNodeInfo {
	return SimpleFNodeInfo{
		NodeName:    node.NodeName,
		IsFaultNode: node.IsFaultNode,
	}
}
