/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package ascend800ia5stacking provides node management and NPU resource allocation functionality
for Ascend 800i A5 stacking scenarios. It handles NPU topology information, network health checks,
and resource allocation across stacked nodes.
*/
package ascend800ia5stacking

import (
	"fmt"
	"math"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// getUsableTopFromNode get available NPU topology information on the node and decide whether to filter out network unhealthy NPUs based on the disFlag parameter
func (tp *module800ia5stacking) getUsableTopFromNode(node plugin.NPUNode, disFlag bool) ([]int, error) {
	nodeTop, err := tp.GetUsableTopFromNode(node, disFlag)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getUsableTopFromNode err: %s", err)
		return nil, err
	}

	if !disFlag {
		return nodeTop, nil
	}
	networkUnhealthyTopStr, ok := node.Annotation[tp.netUnhealthyKey]
	if !ok {
		err := fmt.Errorf("node<%s> don't have resource<%s>", node.Name, tp.netUnhealthyKey)
		klog.V(util.LogWarningLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	networkUnhealthyTop := util.ChangeTopToIntArray(networkUnhealthyTopStr, tp.GetAnnoPreVal())
	if len(networkUnhealthyTop) > tp.MaxNodeNPUNum {
		err := fmt.Errorf("node<%s> npu networkUnhealthy top<%v> is invalid", node.Name, networkUnhealthyTop)
		klog.V(util.LogWarningLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	res := util.RemoveCommonElement(nodeTop, networkUnhealthyTop)
	return res, nil
}

func (tp *module800ia5stacking) getUsableTopFromStack(node plugin.NPUNode, disFlag bool) ([]int, error) {
	if stackingNpuList, exists := tp.SuperPodCache[node.SuperPodID]; !exists {
		return []int{}, nil
	} else {
		availableList := tp.getTheAvailableCommonTop(stackingNpuList)
		if !disFlag {
			return availableList, nil
		}
		networkUnhealthyTopStr, ok := node.Annotation[tp.netUnhealthyKey]
		if !ok {
			err := fmt.Errorf("node<%s> don't have resource<%s>", node.Name, tp.netUnhealthyKey)
			klog.V(util.LogWarningLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
			return nil, err
		}
		networkUnhealthyTop := util.ChangeTopToIntArray(networkUnhealthyTopStr, tp.GetAnnoPreVal())
		if len(networkUnhealthyTop) > tp.MaxNodeNPUNum {
			err := fmt.Errorf("node<%s> npu networkUnhealthy top<%v> is invalid", node.Name, networkUnhealthyTop)
			klog.V(util.LogWarningLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
			return nil, err
		}
		res := util.RemoveCommonElement(availableList, networkUnhealthyTop)
		return res, nil
	}
}

// getTheAvailableCommonTop get the set of cards that are free in all stacks
func (tp *module800ia5stacking) getTheAvailableCommonTop(stackingNodeList []plugin.NPUNode) []int {
	commonAvailableNpu := sets.NewInt()
	// If there's only one machine left in the stacking for scoring queue, it means the corresponding machines no longer meet the card number requirement and will never be selected again, so stacknpu is 0
	if len(stackingNodeList) == 1 {
		return commonAvailableNpu.List()
	}
	for _, stackNode := range stackingNodeList {
		cardIds, err := tp.GetUsableTopFromNode(stackNode, tp.NPUTaskNum > 1)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes getErr: %s", tp.GetPluginName(), err)
			continue
		}
		cardSet := sets.NewInt(cardIds...)
		if len(commonAvailableNpu) == 0 {
			commonAvailableNpu = cardSet
		} else {
			commonAvailableNpu = commonAvailableNpu.Intersection(cardSet)
		}
	}
	return commonAvailableNpu.List()
}

// getNodeBestScoreInStack find the corresponding score from AffScoreList based on the number of NPUs requested by the task and the actual available NPUs on the node
func (tp *module800ia5stacking) getNodeBestScoreInStack(taskNPUNum int, stackNPUNum int, cardNum int) (int, error) {
	if taskNPUNum < 1 || taskNPUNum > nodeNPUNumber {
		return 0, fmt.Errorf("task req npu num<%d> is invalid", taskNPUNum)
	}
	if stackNPUNum < 0 || stackNPUNum > tp.MaxNodeNPUNum {
		return 0, fmt.Errorf("stacking npu num<%d> is invalid", stackNPUNum)
	}
	if cardNum < 1 || cardNum > tp.MaxNodeNPUNum {
		return 0, fmt.Errorf("node npu num<%d> is invalid", cardNum)
	}
	if stackNPUNum < taskNPUNum {
		return tp.AffScoreList[taskNPUNum-1][cardNum-1] + nodeNPUNumber, nil
	}
	return tp.AffScoreList[taskNPUNum-1][cardNum-1], nil
}

func (tp *module800ia5stacking) getBestScoreInComplicatedStack(taskNPUNum int, potentialNPUNum int,
	cardNum int) (int, error) {
	if taskNPUNum < 1 || taskNPUNum > nodeNPUNumber {
		return 0, fmt.Errorf("task req npu num<%d> is invalid", taskNPUNum)
	}
	if potentialNPUNum < 1 || potentialNPUNum > tp.MaxNodeNPUNum {
		return 0, fmt.Errorf("potential stacking npu num<%d> is invalid", potentialNPUNum)
	}
	if cardNum < 1 || cardNum > tp.MaxNodeNPUNum {
		return 0, fmt.Errorf("node npu num<%d> is invalid", cardNum)
	}
	if potentialNPUNum < taskNPUNum {
		return tp.AffScoreList[taskNPUNum-1][cardNum-1] + nodeNPUNumber, nil
	} else {
		return tp.AffScoreList[taskNPUNum-1][cardNum-1] - nodeNPUNumber, nil
	}
}

// getTheSameSuperPodIdNodeNpuNum get stacknpu based on the current npunode and all nodeinfo
func (tp *module800ia5stacking) getTheSameSuperPodIdNodeNpuNum(curNode plugin.NPUNode) ([]int, []int) {
	var stackNPUList []plugin.NPUNode
	cardIds, err := tp.GetUsableTopFromNode(curNode, tp.NPUTaskNum > 1)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes getErr: %s", tp.GetPluginName(), err)
	}
	if stackingNpuList, exists := tp.SuperPodCache[curNode.SuperPodID]; !exists ||
		len(stackingNpuList) != stackingNodeNumber {
		klog.V(util.LogWarningLev).Infof("%s stacking status not available or number not illegal", tp.GetPluginName())
		return []int{}, []int{}
	} else {
		stackNPUList = stackingNpuList
	}
	commonAvailableNpu := tp.getTheAvailableCommonTop(stackNPUList)

	return commonAvailableNpu, cardIds
}

// initSelectedCardCache get previously selected card information for the current job (scaling scenario), and collect all node information with tasks into the cache
func (tp *module800ia5stacking) initSelectedCardCache(task *api.TaskInfo) {
	if len(tp.NPUSelectedCache) != 0 {
		return
	}
	vcJob, ok := tp.ScheduleEnv.Jobs[task.Job]
	if !ok {
		klog.V(util.LogWarningLev).Infof("Cannot get job info with nTask %s.", task.Name)
		return
	}
	tp.NPUSelectedCache[task.Job] = make(map[int32][]int)
	for _, nTask := range vcJob.Tasks {
		nodeName := nTask.NodeName
		if nodeName == "" || nTask.PodStatus != v1.PodRunning {
			continue
		}
		tp.PickedNodeCache[nodeName] = true
		node := tp.ScheduleEnv.Nodes[nodeName]
		cardAnnotation, exists := nTask.Annotation[util.NPU910CardName]
		if !exists {
			continue
		}
		cardList := util.ChangeTopToIntArray(cardAnnotation, tp.GetAnnoPreVal())
		if stackSelected, ok := tp.NPUSelectedCache[task.Job][node.SuperPodID]; ok {
			tp.NPUSelectedCache[task.Job][node.SuperPodID] = util.MergeUnique(stackSelected, cardList)
		} else {
			tp.NPUSelectedCache[task.Job][node.SuperPodID] = cardList
		}
	}
}

// initSelectedNodeCache put selected and selectable nodes into the cache as an additional scoring basis
func (tp *module800ia5stacking) initSelectedNodeCache(nodes []*api.NodeInfo) {
	if len(tp.PickedNodeCache) != 0 {
		return
	}
	for _, node := range nodes {
		tp.PickedNodeCache[node.Name] = true
	}
}

func (tp *module800ia5stacking) getHealthyNpu(superPodID int32) float64 {
	healthyTops := float64(nodeNPUNumber * util.NPUHexKilo)
	if len(tp.SuperPodCache) != 0 {
		for _, nNode := range tp.SuperPodCache[superPodID] {
			healthyNPUNum, _ := nNode.Allocate[v1.ResourceName(tp.GetAnnoName())]
			healthyTops = math.Min(healthyNPUNum, healthyTops)
		}
	}

	return healthyTops
}
