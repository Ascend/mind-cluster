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

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	v12 "k8s.io/kube-scheduler/extender/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// NodeDeviceInfo like node annotation.
type NodeDeviceInfo struct {
	DeviceList map[string]string
	UpdateTime int64
}

// NodeDNodeInfo is node the information reported by noded
type NodeDNodeInfo struct {
	HeartbeatTime     int64
	HeartbeatInterval int
}

// NodeInfoWithNodeD is node the node information and checkCode reported by noded
type NodeInfoWithNodeD struct {
	NodeInfo  NodeDNodeInfo
	CheckCode string
}

// NodeDeviceInfoWithDevPlugin a node has one by cm.
type NodeDeviceInfoWithDevPlugin struct {
	DeviceInfo NodeDeviceInfo
	CheckCode  string
}

// NPUNode the plugin define node info.
type NPUNode struct {
	CommonNode
	VNode
	devInfoUpdateTime int64
}

// CommonNode common npu node properties
type CommonNode struct {
	Name       string
	Capability map[v1.ResourceName]float64
	Allocate   map[v1.ResourceName]float64
	Idle       map[v1.ResourceName]float64
	Annotation map[string]string
	Label      map[string]string
	Address    string
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

func getNodeDeviceInfoFromCM(cmData *v1.ConfigMap) (NodeDeviceInfo, error) {

	devInf := &NodeDeviceInfoWithDevPlugin{}
	data, ok := cmData.Data[util.DevInfoCMKey]
	if !ok {
		return devInf.DeviceInfo, fmt.Errorf("configmap<%s> has no %s", cmData.Name, util.DevInfoCMKey)
	}
	if unmarshalErr := json.Unmarshal([]byte(data), &devInf); unmarshalErr != nil {
		klog.V(util.LogInfoLev).Infof("convertToReSchedulerJobsMapFromCM Unmarshal: %s.", util.SafePrint(unmarshalErr))
		return devInf.DeviceInfo, unmarshalErr
	}

	if checkErr := checkNodeDeviceInfo(devInf); checkErr != nil {
		klog.V(util.LogInfoLev).Infof("checkNodeDeviceInfo failed :%s.", util.SafePrint(checkErr))
		return devInf.DeviceInfo, checkErr
	}
	return devInf.DeviceInfo, nil
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
		klog.V(util.LogErrorLev).Infof("convertToReSchedulerJobsMapFromCM Unmarshal: %s.", util.SafePrint(unmarshalErr))
		return NodeInfoWithNodeD{}, unmarshalErr
	}
	// check got information correction with check code from configmap
	if nodeInfo.CheckCode != util.MakeDataHash(nodeInfo.NodeInfo) {
		return nodeInfo, errors.New("checkCode is not match")
	}
	return nodeInfo, unmarshalErr
}

// initNPUNodeByNodeInf init NPU node from node info and cm.
func (n *NPUNode) initNPUNodeByNodeInf(npuNode *api.NodeInfo, deviceInfos map[string]NodeDeviceInfo,
	nodeInfoOfNodeD map[string]NodeDNodeInfo, vJobTemplate map[string]map[string]util.VResource) error {
	if n == nil || npuNode == nil {
		klog.V(util.LogInfoLev).Infof("InitNPUNodeByNodeInf failed: %s.", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	data, getErr := deviceInfos[npuNode.Name]
	if !getErr || data.DeviceList == nil {
		return fmt.Errorf("getNodeDeviceInfoFromCM %s failed", npuNode.Name)
	}

	capability := npuNode.Capability.ScalarResources
	if !util.IsMapHasNPUResource(capability, util.HwPreName) {
		return fmt.Errorf("%s not NPU node", npuNode.Name)
	}
	n.Name = npuNode.Name
	n.Capability = capability

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

	n.syncAnnotationFromSsnNode(npuNode, nodeInfoOfNodeD[npuNode.Name])

	if err := n.updateNPUNodeDeviceInfos(data); err != nil {
		return nil
	}

	if setVNPUErr := n.setNodeVNPUInfo(npuNode, vJobTemplate); setVNPUErr != nil {
		klog.V(util.LogDebugLev).Infof("setNodeVNPUInfo %s %s", npuNode.Name, setVNPUErr)
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
		return fmt.Errorf("%s not has(or not same) %s", n.Name, k)
	}

	sSlice := strings.Split(nodeA, ",")
	length := len(sSlice)
	if length == 1 && sSlice[0] == "" {
		length = 0
	}
	if length != int(iNum/util.NPUHexKilo) {
		return fmt.Errorf("%s's %s not statble:%d=>%d", n.Name, k, length, int(iNum/util.NPUHexKilo))
	}
	return nil
}

// CheckNPUResourceStableReScheduling check resource stabilize.
func (n NPUNode) CheckNPUResourceStableReScheduling(vcJob SchedulerJob) error {
	k := vcJob.handler.GetAnnoName()
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

func (n *NPUNode) updateNPUNodeDeviceInfos(data NodeDeviceInfo) error {
	if n.devInfoUpdateTime == data.UpdateTime {
		klog.V(util.LogDebugLev).Infof("device info is not update, skip refresh cache")
		return fmt.Errorf("device info is not update, skip refresh cache")
	}
	n.devInfoUpdateTime = data.UpdateTime
	for k, v := range data.DeviceList {
		n.Annotation[k] = v
	}
	return nil
}

func (n *NPUNode) syncAnnotationFromSsnNode(npuNode *api.NodeInfo, nodeInfoOfNodeD NodeDNodeInfo) {
	for k, v := range npuNode.Node.Annotations {
		n.Annotation[k] = v
	}
	// adding noded reported info into NPUNode.Annotation including node heartbeat time and healthy status
	n.Annotation[util.NodedHeartbeatTimeKey] = strconv.FormatInt(nodeInfoOfNodeD.HeartbeatTime, util.Base10)
	n.Annotation[util.NodeDNodeHeartbeatIntervalKey] = strconv.Itoa(nodeInfoOfNodeD.HeartbeatInterval)
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

	// 4.init NPU Nodes by  ssn.Nodes and deviceInfos
	sHandle.initNodesFromSsn(ssn, deviceInfos, nodeInfosOfNodeD)
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

	if !IsNPUTask(taskInfo) {
		return nil
	}
	if err := vcNode.CheckNPUResourceStable(vcJob); err != nil {
		return err
	}
	if err := vcJob.CheckNodeNum(taskInfo, vcNode); err != nil {
		return err
	}
	if err := vcJob.handler.CheckNodeNPUByTask(taskInfo, vcNode); err != nil {
		// node doesn't have enough npu for the task
		klog.V(util.LogDebugLev).Infof("checkNodeNPUByTask %s:%s ,cannot be selected.", vcNode.Name, util.SafePrint(err))
		return fmt.Errorf("checkNodeNPUByTask  %s : %s", vcNode.Name, err)
	}
	klog.V(util.LogDebugLev).Infof("%s NodePredicate %s select successes.", PluginName, vcNode.Name)
	return nil
}

func (sHandle *ScheduleHandler) delNPUNodeNotInSsn(ssn *framework.Session) {
	existNodes := make(map[string]NPUNode)
	for nodeName, nNode := range sHandle.Nodes {
		if _, exist := ssn.Nodes[nodeName]; !exist {
			klog.V(util.LogWarningLev).Infof("node init <%s> is not in session,"+
				"maybe node is deleted or not ready", nodeName)
			continue
		}
		existNodes[nodeName] = nNode
	}
	sHandle.Nodes = existNodes
}

func (sHandle *ScheduleHandler) syncDeviceInfosBySsn(ssn *framework.Session) map[string]NodeDeviceInfo {
	deviceInfos := make(map[string]NodeDeviceInfo)
	sHandle.DeviceInfos.Lock()
	for nodeName := range ssn.Nodes {
		deviceInfos[nodeName] = sHandle.DeviceInfos.Devices[nodeName]
	}
	sHandle.DeviceInfos.Unlock()
	return deviceInfos
}

// syncNodeInfosBySsm is to get a copy of nodeInfo which is get from configmap of nodeD
func (sHandle *ScheduleHandler) syncNodeInfosBySsn(ssn *framework.Session) map[string]NodeDNodeInfo {
	nodeInfos := make(map[string]NodeDNodeInfo, util.MapInitNum)
	sHandle.NodeInfosFromCm.Lock()
	for nodeName := range ssn.Nodes {
		nodeInfos[nodeName] = sHandle.NodeInfosFromCm.Nodes[nodeName]
	}
	sHandle.NodeInfosFromCm.Unlock()
	return nodeInfos
}

func (sHandle *ScheduleHandler) initNodesFromSsn(ssn *framework.Session, deviceInfos map[string]NodeDeviceInfo,
	nodeInfoOfNodeD map[string]NodeDNodeInfo) {
	newNodes := make(map[string]NPUNode)
	for nodeName, nodeInf := range ssn.Nodes {
		// get npu node in map sHandle.Nodes, if exist get old node, if not exist get NPUNode{} for new node init
		node := sHandle.Nodes[nodeName]
		if err := node.initNPUNodeByNodeInf(nodeInf, deviceInfos,
			nodeInfoOfNodeD, sHandle.FrameAttr.VJobTemplate); err != nil {
			klog.V(util.LogErrorLev).Infof("init node(%s) failed, err: %v", nodeName, err)
			if !strings.Contains(err.Error(), notNPUNodeError) {
				klog.V(util.LogDebugLev).Infof("InitNodesFromSsn %s %s, not put in nodes.", nodeName, err)
			}
			continue
		}
		newNodes[nodeName] = node
	}
	sHandle.Nodes = newNodes
}

func initScoreMap(nodes []*api.NodeInfo, interPodAffinityScore v12.HostPriorityList) map[string]float64 {
	for _, node := range nodes {
		if reflect.ValueOf(node).IsNil() {
			continue
		}

		interPodAffinityScore = append(interPodAffinityScore, v12.HostPriority{
			Host:  node.Name,
			Score: 0,
		})
	}
	scoreMap := make(map[string]float64, len(interPodAffinityScore))
	for _, host := range interPodAffinityScore {
		scoreMap[host.Host] = float64(host.Score)
	}

	return scoreMap
}
