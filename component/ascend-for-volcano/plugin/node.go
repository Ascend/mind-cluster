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
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// NodeDeviceInfo like node annotation.
type NodeDeviceInfo struct {
	DeviceList map[string]string
	UpdateTime int64
}

// NodeDeviceInfoWithTime is node device info with time
type NodeDeviceInfoWithTime struct {
	NodeDeviceInfoWithID
	HostUpdateTime int64
}

// NodeDeviceInfoWithID is node the information reported by cm.
type NodeDeviceInfoWithID struct {
	NodeDeviceInfo
	SuperPodID int32
}

// NodeDNodeInfo is node the information reported by noded
type NodeDNodeInfo struct {
	FaultDevList []struct {
		DeviceType string
		DeviceId   int
		FaultCode  []string
		FaultLevel string
	}
	HeartbeatTime     int64
	HeartbeatInterval int
	NodeStatus        string
}

// NodeInfoWithNodeD is node the node information and checkCode reported by noded
type NodeInfoWithNodeD struct {
	NodeInfo  NodeDNodeInfo
	CheckCode string
}

// NodeDeviceInfoWithDevPlugin a node has one by cm.
type NodeDeviceInfoWithDevPlugin struct {
	DeviceInfo  NodeDeviceInfo
	CheckCode   string
	SuperPodID  int32 `json:"SuperPodID,omitempty"`
	ServerIndex int32 `json:"ServerIndex,omitempty"`
}

// SwitchFaultInfo Switch Fault Info
type SwitchFaultInfo struct {
	FaultCode  []string
	FaultLevel string
	UpdateTime int64
	NodeStatus string
}

// NPUNode the plugin define node info.
type NPUNode struct {
	CommonNode
	VNode
	IsUnhealthy       bool
	devInfoUpdateTime int64
}

// CommonNode common npu node properties
type CommonNode struct {
	Name           string
	Capability     map[v1.ResourceName]float64
	Allocate       map[v1.ResourceName]float64
	Idle           map[v1.ResourceName]float64
	BaseDeviceInfo string
	Annotation     map[string]string
	Label          map[string]string
	Address        string
	SuperPodID     int32
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

// checkNodeDeviceInfo will be add more later
func checkNodeDeviceInfo(nodeData *NodeDeviceInfoWithDevPlugin) error {
	if nodeData == nil {
		return errors.New("nil parameters")
	}

	if nodeData.CheckCode == "" {
		return errors.New("checkCode is empty")
	}

	nodeDeviceInfo := nodeData.DeviceInfo
	if nodeData.CheckCode != util.MakeDataHash(nodeDeviceInfo) {
		return errors.New("checkCode is not match")
	}
	return nil
}

func getDeviceClusterInfoFromCM(cmData *v1.ConfigMap) (map[string]NodeDeviceInfoWithID, error) {
	deviceInfoMap := map[string]NodeDeviceInfoWithID{}
	data, ok := cmData.Data[cmData.Name]
	if !ok {
		return deviceInfoMap, fmt.Errorf("configmap<%s> has no %s", cmData.Name, cmData.Name)
	}
	if unmarshalErr := json.Unmarshal([]byte(data), &deviceInfoMap); unmarshalErr != nil {
		klog.V(util.LogInfoLev).Infof("convertToReSchedulerJobsMapFromCM Unmarshal: %s.", util.SafePrint(unmarshalErr))
		return deviceInfoMap, unmarshalErr
	}
	return deviceInfoMap, nil
}

func getNodeDeviceInfoFromCM(cmData *v1.ConfigMap) (*NodeDeviceInfoWithDevPlugin, error) {
	devInf := &NodeDeviceInfoWithDevPlugin{}
	data, ok := cmData.Data[util.DevInfoCMKey]
	if !ok {
		return nil, fmt.Errorf("configmap<%s> has no %s", cmData.Name, util.DevInfoCMKey)
	}
	if unmarshalErr := json.Unmarshal([]byte(data), devInf); unmarshalErr != nil {
		klog.V(util.LogInfoLev).Infof("convertToReSchedulerJobsMapFromCM Unmarshal: %s.", util.SafePrint(unmarshalErr))
		return nil, unmarshalErr
	}

	if checkErr := checkNodeDeviceInfo(devInf); checkErr != nil {
		klog.V(util.LogInfoLev).Infof("checkNodeDeviceInfo failed :%s.", util.SafePrint(checkErr))
		return nil, checkErr
	}
	return devInf, nil
}

// getNodeClusterInfoFromCM try to init NodeInfoWithNodeD with cm information reported by nodeD
func getNodeClusterInfoFromCM(cmData *v1.ConfigMap) (map[string]NodeDNodeInfo, error) {
	nodeInfoMap := map[string]NodeDNodeInfo{}
	if cmData == nil {
		return nodeInfoMap, errors.New("nil parameters")
	}
	data, ok := cmData.Data[cmData.Name]
	if !ok {
		return nodeInfoMap, fmt.Errorf("configmap<%s> has no %s", cmData.Name, cmData.Name)
	}
	if unmarshalErr := json.Unmarshal([]byte(data), &nodeInfoMap); unmarshalErr != nil {
		klog.V(util.LogInfoLev).Infof("convertToReSchedulerJobsMapFromCM Unmarshal: %s.",
			util.SafePrint(unmarshalErr))
		return nodeInfoMap, unmarshalErr
	}
	return nodeInfoMap, nil
}

func getSwitchClusterInfoFromCM(cmData *v1.ConfigMap) (map[string]SwitchFaultInfo, error) {
	switchInfoMap := map[string]SwitchFaultInfo{}
	if cmData == nil {
		return switchInfoMap, errors.New("nil parameters")
	}
	data, ok := cmData.Data[cmData.Name]
	if !ok {
		return switchInfoMap, fmt.Errorf("configmap<%s> has no %s", cmData.Name, cmData.Name)
	}
	if unmarshalErr := json.Unmarshal([]byte(data), &switchInfoMap); unmarshalErr != nil {
		klog.V(util.LogInfoLev).Infof("unmarshal switch info failed, err: %s",
			util.SafePrint(unmarshalErr))
		return switchInfoMap, unmarshalErr
	}
	return switchInfoMap, nil
}

// getNodeInfoFromCM try to init NodeInfoWithNodeD with cm information reported by nodeD
func getNodeInfoFromCM(cmData *v1.ConfigMap) (NodeInfoWithNodeD, error) {
	nodeInfo := NodeInfoWithNodeD{}
	if cmData == nil {
		return NodeInfoWithNodeD{}, errors.New("nil parameters")
	}
	data, ok := cmData.Data[util.NodeInfoCMKey]
	if !ok {
		return NodeInfoWithNodeD{}, fmt.Errorf("configmap<%s> has no %s", cmData.Name, util.NodeInfoCMKey)
	}
	unmarshalErr := json.Unmarshal([]byte(data), &nodeInfo)
	if unmarshalErr != nil {
		klog.V(util.LogInfoLev).Infof("convertToReSchedulerJobsMapFromCM Unmarshal: %s.", util.SafePrint(unmarshalErr))
		return NodeInfoWithNodeD{}, unmarshalErr
	}
	// check got information correction with check code from configmap
	if nodeInfo.CheckCode != util.MakeDataHash(nodeInfo.NodeInfo) {
		return nodeInfo, errors.New("checkCode is not match")
	}
	return nodeInfo, unmarshalErr
}

// initNPUNodeByNodeInf init NPU node from node info and cm.
func (n *NPUNode) initNPUNodeByNodeInf(npuNode *api.NodeInfo, deviceInfos map[string]NodeDeviceInfoWithID,
	nodeInfoOfNodeD map[string]NodeDNodeInfo, switchInfos map[string]SwitchFaultInfo,
	vJobTemplate map[string]map[string]util.VResource) error {
	if n == nil || npuNode == nil {
		klog.V(util.LogInfoLev).Infof("InitNPUNodeByNodeInf failed: %s.", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	n.IsUnhealthy = false
	data, getErr := deviceInfos[npuNode.Name]
	if !getErr || data.DeviceList == nil {
		return fmt.Errorf("getNodeDeviceInfoFromCM %s failed", npuNode.Name)
	}

	capability := getNPUNodeCapacity(npuNode)
	if !util.IsMapHasNPUResource(capability, util.HwPreName) {
		return fmt.Errorf(normalNodeErr)
	}
	n.Name = npuNode.Name
	n.Capability = capability
	if baseDevInfo, ok := npuNode.Node.Annotations[util.BaseDeviceInfoKey]; ok {
		klog.V(util.LogDebugLev).Infof("base device infos: %s", baseDevInfo)
		n.BaseDeviceInfo = baseDevInfo
	}

	n.Allocate = npuNode.Allocatable.ScalarResources
	n.Idle = npuNode.Idle.ScalarResources
	n.Label = npuNode.Node.Labels

	for _, addr := range npuNode.Node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			n.Address = addr.Address
			break
		}
	}

	// sync last session device infos in cache while device infos's updateTime is not update
	n.syncOldDeviceInfoFromCache()

	n.syncAnnotationFromSsnNode(npuNode, nodeInfoOfNodeD[npuNode.Name], switchInfos[npuNode.Name])

	if !n.updateNPUNodeDeviceInfos(data) {
		return nil
	}

	if setVNPUErr := n.setNodeVNPUInfo(npuNode, vJobTemplate); setVNPUErr != nil {
		klog.V(util.LogDebugLev).Infof("setNodeVNPUInfo %s %s", npuNode.Name, setVNPUErr)
	}
	return nil
}

func getNPUNodeCapacity(npuNode *api.NodeInfo) map[v1.ResourceName]float64 {
	valueOfP := reflect.ValueOf(*npuNode)

	if valueOfP.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < valueOfP.NumField(); i++ {
		if valueOfP.Type().Field(i).Name != oldCapacity && valueOfP.Type().Field(i).Name != newCapacity {
			continue
		}
		if capacity, ok := valueOfP.Field(i).Interface().(*api.Resource); ok {
			return capacity.ScalarResources
		}
		klog.V(util.LogErrorLev).Info("get capacity failed by not meet the resource type")
		return nil
	}
	return nil
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
		}
	}
	newTopStr := strings.Join(newTopStrArray, ",")
	return newTopStr, nil
}

// CheckNPUResourceStable check resource stabilize.
func (n NPUNode) CheckNPUResourceStable(vcJob SchedulerJob) error {
	if vcJob.IsVJob() {
		klog.V(util.LogDebugLev).Infof("%s is vNPU job no need check %s stable in frame.", vcJob.Name, n.Name)
		return nil
	}

	k, err := vcJob.GetAnnoName()
	if err != nil {
		return err
	}
	iNum, iOK := n.Idle[v1.ResourceName(k)]
	nodeA, aOK := n.Annotation[k]
	if iOK != true || aOK != true {
		return fmt.Errorf("not has(or not same) %s", k)
	}

	sSlice := strings.Split(nodeA, ",")
	length := len(sSlice)
	if length == 1 && sSlice[0] == "" {
		length = 0
	}
	if length != int(iNum/util.NPUHexKilo) {
		return fmt.Errorf("%s not statble:device-info is <%d> but k8s is <%d>", k, length, int(iNum/util.NPUHexKilo))
	}
	return nil
}

// CheckNPUResourceStableReScheduling check resource stabilize.
func (n NPUNode) CheckNPUResourceStableReScheduling(vcJob util.SchedulerJobAttr) error {
	k := vcJob.ReqNPUName
	iNum, iOK := n.Idle[v1.ResourceName(k)]
	nodeA, aOK := n.Annotation[k]
	if iOK != true || aOK != true {
		return fmt.Errorf("not has(or not same) %s", k)
	}

	sSlice := strings.Split(nodeA, ",")
	length := len(sSlice)
	if length == 1 && sSlice[0] == "" {
		length = 0
	}
	iNumInt := int(iNum / util.NPUHexKilo)
	if length != iNumInt && iNumInt >= 0 {
		return fmt.Errorf("%s not stable:%d=>%d", k, length, iNumInt)
	}
	return nil
}

func (n *NPUNode) syncOldDeviceInfoFromCache() {
	if n.Annotation == nil {
		n.Annotation = make(map[string]string, util.MapInitNum)
	}

	existAnno := make(map[string]string)
	for annoKey, annoValue := range n.Annotation {
		if strings.Contains(annoKey, util.HwPreName) {
			existAnno[annoKey] = annoValue
			continue
		}
	}
	n.Annotation = existAnno
}

// updateNPUNodeDeviceInfos return true if device info was updated, else return false
func (n *NPUNode) updateNPUNodeDeviceInfos(data NodeDeviceInfoWithID) bool {
	if n.devInfoUpdateTime >= data.UpdateTime {
		klog.V(util.LogDebugLev).Infof("device info is not update, skip refresh cache")
		return false
	}
	n.SuperPodID = data.SuperPodID

	n.updateNPUNodeDeviceInfosWithVolcanoCache(data, data.UpdateTime)

	n.devInfoUpdateTime = data.UpdateTime
	klog.V(util.LogDebugLev).Infof("update device info for node<%s> annotations: %v", n.Name, n.Annotation)
	return true
}

func (n *NPUNode) updateNPUNodeDeviceInfosWithVolcanoCache(data NodeDeviceInfoWithID, updateTime int64) {
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
		n.Annotation[k] = n.getRealHealthyDeviceList(k, n.Annotation[k], v)
	}
}

func (n *NPUNode) getRealHealthyDeviceList(deviceKey, oldList, newList string) string {
	// if cache card list is empty or device info is empty. update by device info
	if len(oldList) == 0 || len(newList) == 0 {
		return newList
	}
	newDeviceList := strings.Split(newList, ",")
	oldDeviceList := strings.Split(oldList, ",")

	// if cache is not equal k8s or device info is equal k8s. update by device info
	if int(n.Idle[v1.ResourceName(deviceKey)]/util.NPUHexKilo) != len(oldDeviceList) ||
		int(n.Idle[v1.ResourceName(deviceKey)]/util.NPUHexKilo) == len(newDeviceList) {
		return newList
	}
	oldDevices := make(map[string]struct{})
	for _, device := range oldDeviceList {
		oldDevices[device] = struct{}{}
	}
	var deviceListCache []string
	for _, newDevice := range newDeviceList {
		if _, ok := oldDevices[newDevice]; !ok {
			continue
		}
		deviceListCache = append(deviceListCache, newDevice)
	}
	klog.V(util.LogWarningLev).Infof("update device info for node<%s> annotations: %#v", n.Name, deviceListCache)
	return strings.Join(deviceListCache, ",")
}

func (n *NPUNode) syncAnnotationFromSsnNode(npuNode *api.NodeInfo, nodeInfoOfNodeD NodeDNodeInfo,
	switchInfo SwitchFaultInfo) {
	for k, v := range npuNode.Node.Annotations {
		n.Annotation[k] = v
	}
	// adding noded reported info into NPUNode.Annotation including node heartbeat time and healthy status
	n.Annotation[util.NodedHeartbeatTimeKey] = strconv.FormatInt(nodeInfoOfNodeD.HeartbeatTime, util.Base10)
	n.Annotation[util.NodedNodeHealtyStatuskey] = nodeInfoOfNodeD.NodeStatus
	n.Annotation[util.NodeDNodeHeartbeatIntervalKey] = strconv.Itoa(nodeInfoOfNodeD.HeartbeatInterval)

	n.Annotation[util.SwitchNodeHealtyStatuskey] = switchInfo.NodeStatus
}

// InitNodesFromSsn init all nodes in ssn.
func (sHandle *ScheduleHandler) InitNodesFromSsn(ssn *framework.Session) {
	if sHandle == nil || sHandle.FrameAttr.KubeClient == nil {
		return
	}
	// 1.nodes not in session cannot keep in npu node cache
	sHandle.delNPUNodeNotInSsn(ssn)

	// 2.obtain device infos ,and if node not in session, its device info should not keep in cache
	deviceInfos := sHandle.syncDeviceInfosBySsn(ssn)
	// 3. obtain node infos of noded configmap
	nodeInfosOfNodeD := sHandle.syncNodeInfosBySsn(ssn)
	// 4. obtain switch infos of switch configmap
	switchInfos := sHandle.syncSwitchInfosBySsn(ssn)

	// 4.init NPU Nodes by  ssn.Nodes and deviceInfos
	sHandle.initNodesFromSsn(ssn, deviceInfos, nodeInfosOfNodeD, switchInfos)
	return
}

// NodePredicate Predicate nodes.
func (sHandle *ScheduleHandler) NodePredicate(taskInfo *api.TaskInfo, nodeInfo *api.NodeInfo) error {
	if sHandle == nil || taskInfo == nil || nodeInfo == nil {
		klog.V(util.LogErrorLev).Infof("NodePredicate got null parameter(s), which is invalid.")
		return fmt.Errorf("got null parameter(s)")
	}
	klog.V(util.LogDebugLev).Infof("enter node(%s) predicate", nodeInfo.Name)
	defer klog.V(util.LogDebugLev).Infof("leave node(%s) predicate", nodeInfo.Name)
	vcJob, ok := sHandle.Jobs[taskInfo.Job]
	if !ok {
		klog.V(util.LogDebugLev).Infof("NodePredicate not support job:%s.", util.SafePrint(taskInfo.Job))
		return nil
	}
	// check vcjob is npu job
	if !vcJob.IsNPUJob() {
		klog.V(util.LogDebugLev).Infof("NodePredicate vc-job:%#v is not npu job.", vcJob)
		return nil
	}

	vcNode, ok := sHandle.Nodes[nodeInfo.Name]
	if !ok {
		klog.V(util.LogDebugLev).Infof("NodePredicate %s not in.", nodeInfo.Name)
		return nil
	}
	if vcNode.IsUnhealthy {
		return fmt.Errorf("node is unhealthy")
	}
	if !IsNPUTask(taskInfo) {
		return nil
	}
	if err := vcJob.preCheckNodePredicate(taskInfo, vcNode); err != nil {
		return err
	}
	if err := vcJob.handler.CheckNodeNPUByTask(taskInfo, vcNode); err != nil {
		// node doesn't have enough npu for the task
		klog.V(util.LogDebugLev).Infof("checkNodeNPUByTask %s:%s ,cannot be selected.", vcNode.Name, util.SafePrint(err))
		return fmt.Errorf("checkNodeNPUByTask : %s", err)
	}
	klog.V(util.LogDebugLev).Infof("%s NodePredicate %s select successes.", PluginName, vcNode.Name)
	if !(vcJob.IsJobSinglePodDelete() && vcJob.IsTorAffinityJob()) {
		return nil
	}
	if sHandle.getNSLBVsersion() == defaultNSLBVersion {
		return vcJob.CheckTorJobSinglePodDeleteV1(sHandle, taskInfo, vcNode)
	}
	return vcJob.CheckTorJobSinglePodDeleteV2(sHandle, vcNode)
}

func (sHandle *ScheduleHandler) delNPUNodeNotInSsn(ssn *framework.Session) {
	existNodes := make(map[string]NPUNode)
	// not in node informer and notready node
	nodesNotInSsn := make(map[string]*v1.Node)
	indexer := ssn.InformerFactory().Core().V1().Nodes().Informer().GetIndexer()
	for nodeName, nNode := range sHandle.Nodes {
		if _, exist := ssn.Nodes[nodeName]; exist {
			existNodes[nodeName] = nNode
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
		nodesNotInSsn[nodeName] = vNode
	}
	sHandle.Nodes = existNodes
	sHandle.NodesNotInSsn = nodesNotInSsn
}

func (sHandle *ScheduleHandler) syncDeviceInfosBySsn(ssn *framework.Session) map[string]NodeDeviceInfoWithID {
	deviceInfos := make(map[string]NodeDeviceInfoWithID)
	sHandle.DeviceInfos.Lock()
	for nodeName, devInfo := range sHandle.DeviceInfos.Devices {
		if _, exist := ssn.Nodes[nodeName]; exist {
			continue
		}
		if info, ok := sHandle.DevInfoNotInSession[nodeName]; !ok || info.UpdateTime < devInfo.UpdateTime {
			now := time.Now().Unix()
			klog.V(util.LogWarningLev).Infof("node<%s> is not in session, add device-info in cache, time: %d",
				nodeName, now)
			sHandle.DevInfoNotInSession[nodeName] = NodeDeviceInfoWithTime{
				NodeDeviceInfoWithID: devInfo,
				HostUpdateTime:       now,
			}
		}
	}

	var needDealNodes []string
	for nodeName := range ssn.Nodes {
		needDealNodes = append(needDealNodes, nodeName)
	}
	for nodeName := range sHandle.NodesNotInSsn {
		needDealNodes = append(needDealNodes, nodeName)
	}
	for _, nodeName := range needDealNodes {
		deviceInfos[nodeName] = NodeDeviceInfoWithID{
			NodeDeviceInfo: NodeDeviceInfo{
				DeviceList: make(map[string]string),
				UpdateTime: sHandle.DeviceInfos.Devices[nodeName].UpdateTime,
			},
			SuperPodID: sHandle.DeviceInfos.Devices[nodeName].SuperPodID,
		}
		for deviceName, deviceID := range sHandle.DeviceInfos.Devices[nodeName].DeviceList {
			deviceInfos[nodeName].DeviceList[deviceName] = deviceID
		}

	}
	sHandle.DeviceInfos.Unlock()
	return deviceInfos
}

// syncNodeInfosBySsn is to get a copy of nodeInfo which is get from configmap of nodeD
func (sHandle *ScheduleHandler) syncNodeInfosBySsn(ssn *framework.Session) map[string]NodeDNodeInfo {
	nodeInfos := make(map[string]NodeDNodeInfo, util.MapInitNum)
	sHandle.NodeInfosFromCm.Lock()
	for nodeName := range ssn.Nodes {
		nodeInfos[nodeName] = sHandle.NodeInfosFromCm.Nodes[nodeName]
	}
	for nodeName := range sHandle.NodesNotInSsn {
		nodeInfos[nodeName] = sHandle.NodeInfosFromCm.Nodes[nodeName]
	}
	sHandle.NodeInfosFromCm.Unlock()
	return nodeInfos
}

// syncSwitchInfosBySsn is to get a copy of switchInfo which is get from configmap of switch info
func (sHandle *ScheduleHandler) syncSwitchInfosBySsn(ssn *framework.Session) map[string]SwitchFaultInfo {
	switchInfos := make(map[string]SwitchFaultInfo, util.MapInitNum)
	sHandle.SwitchInfosFromCm.Lock()
	for nodeName := range ssn.Nodes {
		switchInfos[nodeName] = sHandle.SwitchInfosFromCm.Switches[nodeName]
	}
	for nodeName := range sHandle.NodesNotInSsn {
		switchInfos[nodeName] = sHandle.SwitchInfosFromCm.Switches[nodeName]
	}
	sHandle.SwitchInfosFromCm.Unlock()
	return switchInfos
}

func (sHandle *ScheduleHandler) initNodesFromSsn(ssn *framework.Session, deviceInfos map[string]NodeDeviceInfoWithID,
	nodeInfoOfNodeD map[string]NodeDNodeInfo, switchInfos map[string]SwitchFaultInfo) {
	newNodes := make(map[string]NPUNode)
	// apiNode: type is *api.NodeInfo
	for nodeName, apiNode := range ssn.Nodes {
		// get npu node in map sHandle.Nodes, if exist get old node, if not exist get NPUNode{} for new node init
		node := sHandle.Nodes[nodeName]
		if err := node.initNPUNodeByNodeInf(apiNode, deviceInfos, nodeInfoOfNodeD, switchInfos,
			sHandle.FrameAttr.VJobTemplate); err != nil {
			if !strings.Contains(err.Error(), notNPUNodeError) && err.Error() != normalNodeErr {
				klog.V(util.LogErrorLev).Infof("InitNodesFromSsn %s %s, not put in nodes.", nodeName, err)
			}
			continue
		}
		newNodes[nodeName] = node
	}

	// init deviceInfos for node in NodesNotInSsn
	for nodeName, notInSsnNode := range sHandle.NodesNotInSsn {
		if _, ok := newNodes[nodeName]; ok {
			continue
		}
		nNode, ok := sHandle.Nodes[nodeName]
		if ok {
			continue
		}
		if err := nNode.initNPUNodeByNodeInf(api.NewNodeInfo(notInSsnNode), deviceInfos, nodeInfoOfNodeD, switchInfos,
			sHandle.FrameAttr.VJobTemplate); err != nil {
			if !strings.Contains(err.Error(), notNPUNodeError) && err.Error() != normalNodeErr {
				klog.V(util.LogErrorLev).Infof("InitNodesFromSsn %s %s, not put in nodes.", nodeName, err)
			}
			continue
		}
		newNodes[nodeName] = nNode
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
