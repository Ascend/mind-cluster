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
Package plugin is using for HuaWei Ascend pin affinity schedule frame.
*/
package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/k8s"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/affinity/chip/topo"
)

// NPUNode the plugin define node info.
type NPUNode struct {
	CommonNode
	VNode
}

type softShareDevResourceQuota struct {
	aicoreQuota int
	hbmQuota    int
}

// CommonNode common npu node properties
type CommonNode struct {
	Name           string
	Capability     map[v1.ResourceName]float64
	Allocate       map[v1.ResourceName]float64
	Idle           map[v1.ResourceName]float64
	Tasks          map[api.TaskID]*api.TaskInfo
	ChipTopo       *topo.ChipNode
	ChipPods       map[int]map[string]*v1.Pod
	BaseDeviceInfo string
	// convert phy id to device id at ascend950
	PhyIDToDeviceIDMap map[int32]int32
	// node annotation and device info + switch info + node info
	Annotation map[string]string
	// node fault device list
	NodeFaultList []k8s.FaultDevList
	// switch fault code
	SwitchFaultCode []string
	// switch fault level
	SwitchFaultLevel  string
	Label             map[string]string
	Address           string
	SuperPodID        int32
	devInfoUpdateTime int64
	//  [Ascend950] rack id
	RackID int32
	// [Ascend950] server index
	ServerIndex string
}

// VNode vnpu node class
type VNode struct {
	// Chips map chipID to VChip class
	Chips map[int]*VChip
	// ChipKind Ascend910/310/310p
	ChipKind string
	// UnhealthyChipIds the card unhealthy chip ids in this node
	UnhealthyChipIds map[int]struct{}
	// ServerType Ascend310p-10-dual cardType-cardCoreNum-duo
	ServerType string
	// TotalChipNum num of total chips, get from capacity
	TotalChipNum int
	// AiCorePerChip num of aicore on each chip
	AiCorePerChip int
	// FreeChipNum num of free chips get from device-info
	FreeChipNum int
	// TotalRes total resource on node
	TotalRes util.VResource
	// ValidVNode node init success
	ValidVNode bool
	// Chip type 910B1/910B2C/910B3/910B4
	ChipType string
}

// GetChipCount get chip count of npu node.
func (n *NPUNode) GetChipCount(npuResourceName v1.ResourceName) (free, total, occupied int) {
	total = int(n.Allocate[npuResourceName] / util.NPUHexKilo)
	free = int(n.Idle[npuResourceName] / util.NPUHexKilo)
	occupied = total - free
	return
}

// VChip vnpu chip class
type VChip struct {
	PodMap map[string]*v1.Pod
	ID     []string
	// Name Ascend910-0
	Name string
	// Kind Ascend910/Ascend310/Ascend310P
	Kind        string
	IsDual      bool
	Unstable    bool
	CoreNum     int
	SegmentFlag bool
	TotalRes    util.VResource
	UsedRes     util.VResource
	FreeRes     util.VResource
}

// InitNodesFromSsn init all nodes in ssn.
func (sHandle *ScheduleHandler) InitNodesFromSsn(ssn *framework.Session) {
	if sHandle == nil {
		return
	}
	// 1.obtain need init node info list
	nodeList := sHandle.getNeedInitNodeList(ssn)
	// 2.init NPU Nodes by enable node list
	sHandle.initNodesFromSsn(nodeList)
	return
}

// NodePredicate Predicate nodes.
func (sHandle *ScheduleHandler) NodePredicate(taskInfo *api.TaskInfo, nodeInfo *api.NodeInfo) error {
	if sHandle == nil || taskInfo == nil || nodeInfo == nil {
		klog.V(util.LogErrorLev).Infof("NodePredicate got null parameter(s), which is invalid.")
		return fmt.Errorf("got null parameter(s)")
	}

	vcNode, ok := sHandle.Nodes[nodeInfo.Name]
	if !ok {
		klog.V(util.LogDebugLev).Infof("NodePredicate %s not in.", nodeInfo.Name)
		return nil
	}

	return sHandle.NodePredicateOnVCNode(taskInfo, vcNode)
}

// NodePredicateOnVCNode
func (sHandle *ScheduleHandler) NodePredicateOnVCNode(taskInfo *api.TaskInfo, vcNode NPUNode) error {
	if sHandle == nil || taskInfo == nil {
		return fmt.Errorf("got null parameter(s)")
	}

	// Bypass DRA tasks before fault check: DRA pods are allocated by the DRA
	// driver, they should never be rejected by NPU fault handling.
	if util.IsDRATask(taskInfo) {
		klog.V(util.LogInfoLev).Infof("NodePredicate bypass DRA task <%s> on node <%s>.",
			taskInfo.Name, vcNode.Name)
		return nil
	}

	if sHandle.FaultHandle != nil {
		if err := sHandle.FaultHandle.CheckNodeNPUByTask(taskInfo, &vcNode); err != nil {
			return err
		}
	}

	if !util.IsNPUTask(taskInfo) {
		return nil
	}
	klog.V(util.LogDebugLev).Infof("enter node(%s) predicate", vcNode.Name)
	defer klog.V(util.LogDebugLev).Infof("leave node(%s) predicate", vcNode.Name)
	vcJob, ok := sHandle.Jobs[taskInfo.Job]
	if !ok {
		klog.V(util.LogDebugLev).Infof("NodePredicate not support job:%s.", util.SafePrint(taskInfo.Job))
		return nil
	}
	// check vcjob is npu job
	if !vcJob.isNPUJob() {
		klog.V(util.LogDebugLev).Infof("NodePredicate vc-job:%#v is not npu job.", vcJob)
		return nil
	}

	if err := vcJob.preCheckNodePredicate(taskInfo, vcNode); err != nil {
		return err
	}

	if err := vcJob.policyHandler.CheckNodeNPUByTask(taskInfo, vcNode); err != nil {
		// node doesn't have enough npu for the task
		klog.V(util.LogDebugLev).Infof("checkNodeNPUByTask %s:%s ,cannot be selected.", vcNode.Name,
			util.SafePrint(err))
		return err
	}
	return nil
}

// build PhyIDToDeviceIDMap that convert phy id to device id
func (n *NPUNode) buildPhyIdToDeviceIdMap() error {
	deviceInfoMap := make(map[string]*util.NpuBaseInfo)
	err := json.Unmarshal([]byte(n.BaseDeviceInfo), &deviceInfoMap)
	if err != nil {
		klog.Infof("buildPhyIdToDeviceIdMap unmarshal err, origin value: %v", n.BaseDeviceInfo)
		return fmt.Errorf("unmarshal annotation baseDeviceInfos failed")
	}
	n.PhyIDToDeviceIDMap = make(map[int32]int32)
	for device, deviceInfo := range deviceInfoMap {
		if deviceInfo.DeviceID == nil {
			return fmt.Errorf("lack device id in baseDeviceInfos at node %v", n.Name)
		}
		parts := strings.Split(device, "-")
		if len(parts) != SplitedLength {
			continue
		}
		phyID, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("parse phy id failed: %s", parts[1])
		}
		n.PhyIDToDeviceIDMap[int32(phyID)] = *deviceInfo.DeviceID
	}
	if len(n.PhyIDToDeviceIDMap) != len(deviceInfoMap) {
		// clear
		n.PhyIDToDeviceIDMap = make(map[int32]int32)
		return fmt.Errorf("phyIDToDeviceIDMap length not equal deviceInfoMap length")
	}
	return nil
}

// initNPUNodeByNodeInf init NPU node from node info and cm.
func (n *NPUNode) initNPUNodeByNodeInf(npuNode *api.NodeInfo, deviceInfo k8s.NodeDeviceInfoWithID,
	nodeInfoOfNodeD k8s.NodeDNodeInfo, switchInfo k8s.SwitchFaultInfo, dpuInfo k8s.DpuInfoWithNode,
	vJobTemplate map[string]map[string]util.VResource) error {
	if n == nil || npuNode == nil {
		klog.V(util.LogInfoLev).Infof("InitNPUNodeByNodeInf failed: %s.", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	n.Name = npuNode.Name
	n.BaseDeviceInfo, _ = util.GetAnnotationValue(npuNode.Node.Annotations,
		util.NPUBaseDevInfosAnnotation, util.BaseDeviceInfoKeyDeprecated)
	n.Allocate = npuNode.Allocatable.ScalarResources
	n.Idle = npuNode.Idle.ScalarResources
	n.Label = npuNode.Node.Labels
	n.Address = getNPUNodeAddress(npuNode)
	n.Tasks = npuNode.Tasks
	n.syncAnnotation(npuNode, nodeInfoOfNodeD, switchInfo)
	capability := npuNode.Capacity.ScalarResources
	// serverIndex key of node annotations for A5
	// Prefer topotree.serverid label, fallback to serverIndex annotation
	var ok bool
	n.ServerIndex, ok = util.GetLabelValue(npuNode.Node.Labels, util.TopoLabelServerId)
	if !ok {
		n.ServerIndex, _ = util.GetNodeAnnotation(npuNode.Node, util.ServerIndexKeyDeprecated)
	}
	if capability == nil {
		return fmt.Errorf("node %s capability is invalid", npuNode.Name)
	}
	n.Capability = capability
	if !util.IsMapHasNPUResource(capability, util.HwPreName) {
		return fmt.Errorf("node %s npu resource is not enable", npuNode.Name)
	}
	chipName, ok := util.GetLabelValue(n.Label, util.NPUChipNameLabel, util.ChipTypeKeyDeprecated)
	// build PhyIdToDeviceIdMap in Ascend 950
	if ok && strings.HasPrefix(chipName, Ascend950Prefix) {
		if err := n.buildPhyIdToDeviceIdMap(); err != nil {
			return fmt.Errorf("build phyIdToDeviceIdMap failed: %s", err.Error())
		}
	}
	if deviceInfo.DeviceList == nil {
		return fmt.Errorf("node %s device info or clusterd info is not enable", npuNode.Name)
	}
	n.updateNPUNodeDeviceInfos(deviceInfo)
	n.updateNPUNodeDpuInfos(dpuInfo)

	if setVNPUErr := n.setNodeVNPUInfo(npuNode, vJobTemplate); setVNPUErr != nil {
		klog.V(util.LogDebugLev).Infof("setNodeVNPUInfo %s %s", npuNode.Name, setVNPUErr)
	}
	n.ParseChipTopology(npuNode)
	klog.V(util.LogDebugLev).Infof("initNPUNodeByNodeInf <%s> success %#v", npuNode.Name, n.CommonNode)
	return nil
}

func getNPUCapacity(node *api.NodeInfo) int {
	for name, res := range node.Capacity.ScalarResources {
		if name == util.HwPreName+util.Ascend910 || name == util.NPUCardName {
			return int(res / util.NPUHexKilo)
		}
	}
	return 0
}

// ParseChipTopology parse node huawei.com/npu.topology annotation
func (n *NPUNode) ParseChipTopology(node *api.NodeInfo) {
	raw, exist := n.Annotation[util.TopologyAnnoKey]
	if !exist {
		raw = topo.BuildFlatTopology(getNPUCapacity(node))
	}
	if n.ChipTopo == nil || n.ChipTopo.Raw != raw {
		n.ChipTopo = topo.ParseTopology(raw)
	}
	if n.ChipTopo == nil {
		return
	}
	n.ChipPods = make(map[int]map[string]*v1.Pod)
	for _, ti := range node.Tasks {
		if !util.IsNPUTask(ti) || ti.Pod == nil {
			continue
		}
		for _, chipID := range getAllocatedChipIDsFromPod(ti.Pod, n) {
			if n.ChipPods[chipID] == nil {
				n.ChipPods[chipID] = make(map[string]*v1.Pod, util.MapInitNum)
			}
			n.ChipPods[chipID][string(ti.Pod.UID)] = ti.Pod
		}
	}

	var faulty, netUnh map[int]struct{}
	maxCardID := 0
	for key, val := range n.Annotation {
		pre, ok := cardAnnoNpuPre(key)
		if !ok {
			continue
		}
		for id := range util.ParseDevList(val, pre) {
			if id > maxCardID {
				maxCardID = id
			}
			switch {
			case strings.HasSuffix(key, util.NPUUnhealthySuffix):
				if faulty == nil {
					faulty = make(map[int]struct{})
				}
				faulty[id] = struct{}{}
			case strings.HasSuffix(key, util.NPUNetworkUnhealthy):
				if netUnh == nil {
					netUnh = make(map[int]struct{})
				}
				netUnh[id] = struct{}{}
			}
		}
	}
	for id := range n.ChipPods {
		if id > maxCardID {
			maxCardID = id
		}
	}
	if n.chipTopoOversize(maxCardID) {
		n.ChipTopo = nil
		return
	}
	owners := make(map[string][]int)
	for id, pods := range n.ChipPods {
		for uid, pod := range pods {
			if pod != nil && pod.Status.Phase != v1.PodFailed && pod.Status.Phase != v1.PodSucceeded {
				owners[uid] = append(owners[uid], id)
			}
		}
	}
	n.ChipTopo.Init(faulty, netUnh, owners)
}

// cardAnnoNpuPre
func cardAnnoNpuPre(key string) (string, bool) {
	family := strings.TrimPrefix(key, util.HwPreName)
	family = strings.TrimSuffix(family, util.NPUUnhealthySuffix)
	family = strings.TrimSuffix(family, util.NPUNetworkUnhealthy)
	switch family {
	case "Ascend910", "Ascend910b":
		return util.NPU910CardNamePre, true
	case "Ascend310P":
		return util.NPU310PCardNamePre, true
	case "Ascend310":
		return util.NPU310CardNamePre, true
	case "npu":
		return util.NPUCardNamePre, true
	}
	return "", false
}

// chipTopoOversize reports whether the node owns a card whose id exceeds the
// topology's max chip id. maxCardID is recorded once in ParseChipTopology while
// the annotations are already being walked, instead of re-parsing every card id here.
func (n *NPUNode) chipTopoOversize(maxCardID int) bool {
	maxID := n.ChipTopo.MaxChipID()
	if maxCardID > maxID {
		klog.V(util.LogErrorLev).Infof(
			"node<%s> topo<%s> abnormal: node card id<%d> exceeds topo max chip id<%d>, init failed",
			n.Name, n.ChipTopo.Raw, maxCardID, maxID)
		return true
	}
	return false
}

// getNPUNodeAddress get npu node address
func getNPUNodeAddress(npuNode *api.NodeInfo) string {
	for _, addr := range npuNode.Node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			return addr.Address
		}
	}
	return ""
}

func (n *NPUNode) getUsedResourceQuotaMap(topStrArray []string) map[string]softShareDevResourceQuota {
	topUsedResourceQuotaMap := make(map[string]softShareDevResourceQuota)
	for _, cardStr := range topStrArray {
		topUsedResourceQuotaMap[cardStr] = softShareDevResourceQuota{aicoreQuota: 0, hbmQuota: 0}
	}
	for _, taskInfo := range n.Tasks {
		aicoreQuotaAnno, existAicoreQuotaAnno := taskInfo.Pod.Annotations[util.SchedulerSoftShareDevAicoreQuotaKey]
		hbmQuotaAnno, existHbmQuotaAnno := taskInfo.Pod.Annotations[util.SchedulerSoftShareDevHbmQuotaKey]
		ascendReal, existAscendReal := taskInfo.Pod.Annotations[util.AscendNPUPodRealUse]
		if !existAicoreQuotaAnno || !existHbmQuotaAnno || !existAscendReal {
			continue
		}
		aicoreQuota, err := strconv.Atoi(aicoreQuotaAnno)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("GetNewNPUNodeAnnotation err: %s.", err.Error())
			continue
		}
		hbmQuota, err := strconv.Atoi(hbmQuotaAnno)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("GetNewNPUNodeAnnotation err: %s.", err.Error())
			continue
		}
		tmpUsedResourceQuota := topUsedResourceQuotaMap[ascendReal]
		topUsedResourceQuotaMap[ascendReal] = softShareDevResourceQuota{
			aicoreQuota: tmpUsedResourceQuota.aicoreQuota + aicoreQuota,
			hbmQuota:    tmpUsedResourceQuota.hbmQuota + hbmQuota}
	}
	return topUsedResourceQuotaMap
}

func (n *NPUNode) getChipMemoryFromNodeLabel() int {
	chipMemory, ok := util.GetLabelValue(n.Label, util.NPUChipMemoryLabelKey, util.NPUChipMemoryLabelKeyDeprecated)
	if !ok {
		return 0
	}
	npuChipMemoryLabel := strings.Replace(chipMemory, "G", "", -1)
	npuChipMemory, err := strconv.Atoi(npuChipMemoryLabel)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("GetNewNPUNodeAnnotation err: %s.", err.Error())
		return 0
	}
	return npuChipMemory
}

// GetNewNPUNodeAnnotation get new annotation after allocate
func (n *NPUNode) GetNewNPUNodeAnnotation(usedTop []int, resourceName, resourceNamePre string) (string, error) {
	if n == nil || len(usedTop) == 0 || resourceName == "" || resourceNamePre == "" {
		klog.V(util.LogInfoLev).Infof("GetNewNPUNodeAnnotation failed: %s.", util.ArgumentError)
		return "", errors.New(util.ArgumentError)
	}
	annotation, ok := n.Annotation[resourceName]
	if !ok {
		err := fmt.Errorf("node<%s> not have resource<%s>", n.Name, resourceName)
		klog.V(util.LogErrorLev).Infof("GetNewNPUNodeAnnotation err: %s.", err.Error())
		return "", err
	}
	if annotation == "" {
		return "", nil
	}
	usedSet := sets.NewInt(usedTop...)
	topStrArray := strings.Split(annotation, ",")
	topUsedResourceQuotaMap := n.getUsedResourceQuotaMap(topStrArray)
	npuChipMemory := n.getChipMemoryFromNodeLabel()
	var newTopStrArray []string
	for _, cardStr := range topStrArray {
		v := strings.TrimPrefix(cardStr, resourceNamePre)
		cardInt, err := strconv.Atoi(v)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("ChangeTopToIntArray conv failed %v.", err)
			return "", err
		}
		if !usedSet.Has(cardInt) {
			newTopStrArray = append(newTopStrArray, cardStr)
		} else {
			softShareDevEnable, ok := n.Label[util.SchedulerSoftShareDevEnableNodeLabel]
			if !ok || softShareDevEnable != "true" {
				continue
			}
			if topUsedResourceQuota, exists := topUsedResourceQuotaMap[cardStr]; exists &&
				(topUsedResourceQuota.aicoreQuota >= util.MaxAicoreQuota ||
					topUsedResourceQuota.hbmQuota >= npuChipMemory*util.MBPerGB) {
				continue
			}
			newTopStrArray = append(newTopStrArray, cardStr)
		}
	}
	newTopStr := strings.Join(newTopStrArray, ",")
	return newTopStr, nil
}

// checkNPUResourceStable check resource stabilize.
func (n NPUNode) checkNPUResourceStable(vcJob SchedulerJob) error {
	if vcJob.IsVJob() {
		klog.V(util.LogDebugLev).Infof("%s is vNPU job no need check %s stable in frame.", vcJob.Name, n.Name)
		return nil
	}

	k := vcJob.ReqNPUName
	iNum, iOK := n.Idle[v1.ResourceName(k)]
	cNum, cOK := n.Capability[v1.ResourceName(k)]
	nodeA, aOK := n.Annotation[k]
	if iOK != true || aOK != true || cOK != true {
		return fmt.Errorf("not has(or not same) %s", k)
	}

	sSlice := strings.Split(nodeA, ",")
	length := len(sSlice)
	if length == 1 && sSlice[0] == "" {
		length = 0
	}
	// public fault occurred, device info <= k8s
	if length > int(iNum/util.NPUHexKilo) || length > int(cNum/util.NPUHexKilo) {
		return fmt.Errorf("%s not stable:device-info is <%d> but k8s is <%d>", k, length, int(iNum/util.NPUHexKilo))
	}
	if int(iNum/util.NPUHexKilo) > int(cNum/util.NPUHexKilo) {
		return fmt.Errorf("%s is not stable because of capability=%d < allocatable=%d", n.Name,
			int(cNum/util.NPUHexKilo), int(iNum/util.NPUHexKilo))
	}
	return nil
}

func (n *NPUNode) updateNPUNodeDpuInfos(dpuInfo k8s.DpuInfoWithNode) {
	if dpuInfo.UpdateTime == 0 {
		klog.V(util.LogDebugLev).Infof("node %s dpu info is empty, skip", n.Name)
		return
	}
	dpuData, err := json.Marshal(dpuInfo.DpuInfoCfg)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("marshal dpu info for node %s failed: %s", n.Name, err)
		return
	}
	if n.Annotation == nil {
		n.Annotation = make(map[string]string)
	}
	n.Annotation[util.DpuInfoAnnoKey] = string(dpuData)
	klog.V(util.LogDebugLev).Infof("update dpu info for node<%s>, updateTime: %d", n.Name, dpuInfo.UpdateTime)
}

// updateNPUNodeDeviceInfos return true if device info was updated, else return false
func (n *NPUNode) updateNPUNodeDeviceInfos(data k8s.NodeDeviceInfoWithID) {
	if n.devInfoUpdateTime >= data.UpdateTime {
		klog.V(util.LogDebugLev).Infof("device info is not update, skip refresh cache")
		return
	}
	n.SuperPodID = data.SuperPodID
	n.RackID = data.RackID
	n.updateNPUNodeDeviceInfosWithVolcanoCache(data, data.UpdateTime)

	n.devInfoUpdateTime = data.UpdateTime
	klog.V(util.LogDebugLev).Infof("update device info for node<%s> annotations: %v", n.Name, n.Annotation)
	return
}

func (n *NPUNode) updateNPUNodeDeviceInfosWithVolcanoCache(data k8s.NodeDeviceInfoWithID, updateTime int64) {
	unhealthyCard := ""
	for k, v := range n.Annotation {
		if strings.HasSuffix(k, util.NPUUnhealthySuffix) {
			unhealthyCard = v
			break
		}
	}
	for k, v := range data.DeviceList {
		// if k does not represent huawei.com/Ascend910/310/310P continue
		if len(strings.Split(k, "-")) > 1 {
			n.Annotation[k] = v
			continue
		}
		// if time interval over 10s continue
		if updateTime-n.devInfoUpdateTime > deviceInfoForceUpdateInterval {
			n.Annotation[k] = v
			continue
		}
		n.Annotation[k] = n.getRealHealthyDeviceList(k, n.Annotation[k], v, unhealthyCard)
	}
}

func (n *NPUNode) getRealHealthyDeviceList(deviceKey, oldList, newList, oldUnhealthyCard string) string {
	// if cache card list is empty or device info is empty. update by device info
	if len(oldList) == 0 || len(newList) == 0 {
		return newList
	}
	newDeviceList := strings.Split(newList, ",")
	oldDeviceList := strings.Split(oldList, ",")
	oldUnhealthyList := strings.Split(oldUnhealthyCard, ",")

	// if cache is not equal k8s or device info is equal k8s. update by device info
	if int(n.Idle[v1.ResourceName(deviceKey)]/util.NPUHexKilo) != len(oldDeviceList) ||
		int(n.Idle[v1.ResourceName(deviceKey)]/util.NPUHexKilo) == len(newDeviceList) {
		return newList
	}
	oldDevices := make(map[string]struct{})
	for _, device := range oldDeviceList {
		oldDevices[device] = struct{}{}
	}
	oldUnhealthyDevices := make(map[string]struct{})
	for _, device := range oldUnhealthyList {
		oldUnhealthyDevices[device] = struct{}{}
	}

	var deviceListCache []string
	for _, newDevice := range newDeviceList {
		_, existInOld := oldDevices[newDevice]
		_, existInUnhealthy := oldUnhealthyDevices[newDevice]
		if !existInOld && !existInUnhealthy {
			continue
		}
		deviceListCache = append(deviceListCache, newDevice)
	}
	klog.V(util.LogWarningLev).Infof("update device info for node<%s> annotations: %#v", n.Name, deviceListCache)
	return strings.Join(deviceListCache, ",")
}

// syncAnnotation 4 parts, 1 v1.node annotations, 2 last session device infos, 3 switch info, 4 noded info
func (n *NPUNode) syncAnnotation(npuNode *api.NodeInfo, nodeInfoOfNodeD k8s.NodeDNodeInfo,
	switchInfo k8s.SwitchFaultInfo) {
	klog.V(util.LogDebugLev).Infof("nodeInfoOfNodeD: %v, switchInfo: %v", nodeInfoOfNodeD, switchInfo)
	existAnno := make(map[string]string)
	// 1. sync v1.node annotations
	for k, v := range npuNode.Node.Annotations {
		existAnno[k] = v
	}
	// 2. last session device infos
	for annoKey, annoValue := range n.Annotation {
		if strings.Contains(annoKey, util.HwPreName) {
			existAnno[annoKey] = annoValue
			continue
		}
	}
	// 3. switch info
	existAnno[util.SwitchNodeHealtyStatuskey] = switchInfo.NodeStatus
	n.SwitchFaultCode = switchInfo.FaultCode
	n.SwitchFaultLevel = switchInfo.FaultLevel
	// 4. noded info. adding noded reported info into NPUNode.Annotation including node healthy status
	// when there are no faults on the node, node info cm does not exist
	n.NodeFaultList = make([]k8s.FaultDevList, 0, len(nodeInfoOfNodeD.FaultDevList))
	for _, faultDev := range nodeInfoOfNodeD.FaultDevList {
		var tmpList k8s.FaultDevList
		tmpList.FaultLevel = faultDev.FaultLevel
		tmpList.FaultCode = faultDev.FaultCode
		tmpList.DeviceType = faultDev.DeviceType
		tmpList.DeviceId = faultDev.DeviceId
		n.NodeFaultList = append(n.NodeFaultList, tmpList)
	}
	if existAnno[util.NodeHealthyStatusKey] == util.NodeUnHealthy {
		n.Annotation = existAnno
		return
	}
	if nodeInfoOfNodeD.NodeStatus != "" {
		existAnno[util.NodeHealthyStatusKey] = nodeInfoOfNodeD.NodeStatus
	} else {
		existAnno[util.NodeHealthyStatusKey] = util.NodeHealthyByNodeD
	}
	n.Annotation = existAnno
}

// getNeedInitNodeList init all nodes in ssn.
func (sHandle *ScheduleHandler) getNeedInitNodeList(ssn *framework.Session) []*api.NodeInfo {
	if sHandle == nil || sHandle.FrameAttr.KubeClient == nil {
		return ssn.NodeList
	}
	nodeList := make([]*api.NodeInfo, 0)
	indexer := sHandle.FrameAttr.informerFactory.Core().V1().Nodes().Informer().GetIndexer()
	for nodeName := range sHandle.Nodes {
		if _, exist := ssn.Nodes[nodeName]; exist {
			continue
		}
		klog.V(util.LogWarningLev).Infof("node <%s> is not in session when initializing,"+
			"maybe node is deleted or not ready", nodeName)
		obj, exist, err := indexer.GetByKey(nodeName)
		if err != nil || !exist {
			klog.V(util.LogWarningLev).Infof("node <%s> is not in informer indexer, maybe is deleted", nodeName)
			continue
		}
		// nNode: type is NPUNode; vNode: type is *v1.Node
		vNode, ok := obj.(*v1.Node)
		if !ok || !util.IsNodeReady(vNode) {
			klog.V(util.LogWarningLev).Infof("node <%s> is real notready", nodeName)
			continue
		}
		nodeList = append(nodeList, api.NewNodeInfo(vNode))
	}

	for _, node := range ssn.NodeList {
		if !util.IsNodeReady(node.Node) {
			klog.V(util.LogWarningLev).Infof("node <%s> is real notready", node.Name)
			continue
		}
		nodeList = append(nodeList, node)
	}

	return nodeList
}

func (sHandle *ScheduleHandler) initNodesFromSsn(nodeList []*api.NodeInfo) {
	// 1.obtain device infos, and if node not in session, its device info should not keep in cache
	deviceInfos := k8s.GetDeviceInfosAndSetInformerStart(nodeList, sHandle.FrameAttr.UseClusterD,
		sHandle.FrameAttr.SelfMaintainAvailCard)
	// 2. obtain node infos of noded configmap
	nodeInfosOfNodeD := k8s.GetNodeDInfos(nodeList)
	// 3. obtain switch infos of switch configmap
	switchInfos := k8s.GetSwitchInfos(nodeList)
	// 4. obtain dpu infos of dpu-dp or clusterd configmap
	dpuInfos := k8s.GetDpuInfos(nodeList)

	newNodes := make(map[string]NPUNode)
	// apiNode: type is *api.NodeInfo
	// init node by node list and config infos
	for _, apiNode := range nodeList {
		// get npu node in map sHandle.Nodes, if exist get old node, if not exist get NPUNode{} for new node init
		node := sHandle.Nodes[apiNode.Name]
		if err := node.initNPUNodeByNodeInf(apiNode, deviceInfos[apiNode.Name], nodeInfosOfNodeD[apiNode.Name],
			switchInfos[apiNode.Name], dpuInfos[apiNode.Name], sHandle.FrameAttr.VJobTemplate); err != nil &&
			!strings.Contains(err.Error(), noneResourceErr) {
			klog.V(util.LogErrorLev).Infof("InitNodesFromSsn %s %s, not put in nodes.", apiNode.Name, err)
			continue
		}
		newNodes[apiNode.Name] = node
	}
	sHandle.Nodes = newNodes
}

func initScoreMap(nodes []*api.NodeInfo) map[string]float64 {
	scoreMap := make(map[string]float64, len(nodes))
	for _, node := range nodes {
		if reflect.ValueOf(node).IsNil() {
			continue
		}
		scoreMap[node.Name] = 0.0
	}
	return scoreMap
}
