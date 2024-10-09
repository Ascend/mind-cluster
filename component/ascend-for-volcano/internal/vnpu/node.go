/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package vnpu is using for HuaWei Ascend pin vnpu allocation.
*/
package vnpu

import (
	"sort"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (tp *DynamicVNPU) orderVNodesByFreeResource(nodes []*api.NodeInfo) []*api.NodeInfo {
	tempVNodes := vNodesList(nodes)
	sort.Sort(tempVNodes)
	return tempVNodes
}

type vNodesList []*api.NodeInfo

// Len for order.
func (vNodes vNodesList) Len() int {
	return len(vNodes)
}

// Less for order.
func (vNodes vNodesList) Less(i, j int) bool {
	if i > vNodes.Len() || j > vNodes.Len() {
		return false
	}
	iIdleAiCore, ok := vNodes[i].Idle.ScalarResources[util.AscendNPUCore]
	if !ok {
		return false
	}
	jIdleAiCore, ok := vNodes[j].Idle.ScalarResources[util.AscendNPUCore]
	if !ok {
		return true
	}
	return iIdleAiCore < jIdleAiCore
}

// Swap for order.
func (vNodes vNodesList) Swap(i, j int) {
	if i > vNodes.Len() || j > vNodes.Len() {
		return
	}
	vNodes[i], vNodes[j] = vNodes[j], vNodes[i]
}

type vChipsList []*plugin.VChip

// Len for order.
func (vChips vChipsList) Len() int {
	return len(vChips)
}

// Less for order.
func (vChips vChipsList) Less(i, j int) bool {
	if i > vChips.Len() || j > vChips.Len() {
		return false
	}
	return !vChips[i].FreeRes.BeGreater(vChips[j].FreeRes)
}

// Swap for order.
func (vChips vChipsList) Swap(i, j int) {
	if i > vChips.Len() || j > vChips.Len() {
		return
	}
	vChips[i], vChips[j] = vChips[j], vChips[i]
}
