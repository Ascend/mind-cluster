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

// Package superpod for base function
package superpod

import (
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// scheduleStrategy the interface need to be implemented by any schedule strategy
type scheduleStrategy interface {
	entrySelect(*superPodsInfo) (bool, error) // select entry
	getStrategyName() strategyKey
}

// strategy for basic and sp-block scheduling
type strategy struct {
	*module910a5SuperPod
	selectedNodes map[string][]plugin.SuperNode
	unReadyIds    []string
}

// oneRackStrategy select nodes in one rack
type oneRackStrategy struct {
	*strategy
	name strategyKey
}

// oneUBMemStrategy select nodes in one UBMem
type oneUBMemStrategy struct {
	*strategy
	name strategyKey
}

// oneSuperPodStrategy select nodes in one superpod
type oneSuperPodStrategy struct {
	*strategy
	name strategyKey
}

// mulSuperPodsStrategy select nodes in multiple superpods
type mulSuperPodsStrategy struct {
	*strategy
	name strategyKey
}

// doSelect select nodes from ubmem level or superpod level, when we have selected one sp-block count nodes, return result
func (tp *strategy) selectOneSpBlock(rackGroup map[int32][]nodeBaseInfo, superPod superPod) map[int32][]nodeBaseInfo {
	spIndex := tp.unReadyIds[tp.totalCount-1]
	if tp.tpBlock == 0 {
		klog.V(util.LogErrorLev).Infof("invalid tp-block, select nodes failed")
		return rackGroup
	}
	tpCountRemain := tp.spBlock / tp.tpBlock

	rackIdGroup := sortRackIdByLengthInOneSuperPod(rackGroup)
	// iterate all RackId sorted by nodes length
	for _, rackId := range rackIdGroup {
		// already selected success
		if tpCountRemain == 0 {
			break
		}
		// how many tp-block count can be selected in this rack
		tpBlockNum := len(rackGroup[rackId]) / tp.tpBlock
		if tpCountRemain < tpBlockNum {
			tpBlockNum = tpCountRemain
		}
		tpCountRemain -= tpBlockNum
		_, ok := tp.selectedNodes[spIndex]
		if !ok {
			tp.selectedNodes[spIndex] = make([]plugin.SuperNode, 0)
		}
		// append node to selectedNodes
		for i := 0; i < tpBlockNum*tp.tpBlock; i++ {
			nNode := rackGroup[rackId][i]
			tp.selectedNodes[spIndex] = append(tp.selectedNodes[spIndex], plugin.SuperNode{
				Name:       nNode.name,
				SuperPodID: nNode.superPodID,
				RackID:     nNode.rackID,
			})
			delete(superPod, nNode.name)
		}
		// remove selected nodes from rackGroup, making totalNodes less
		rackGroup[rackId] = rackGroup[rackId][tpBlockNum*tp.tpBlock:]
	}
	tp.totalCount--

	return rackGroup
}

func (tp *oneRackStrategy) getStrategyName() strategyKey {
	return tp.name
}

func (tp *oneUBMemStrategy) getStrategyName() strategyKey {
	return tp.name
}

func (tp *oneSuperPodStrategy) getStrategyName() strategyKey {
	return tp.name
}

func (tp *mulSuperPodsStrategy) getStrategyName() strategyKey {
	return tp.name
}
