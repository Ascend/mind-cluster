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

// Package chip8node8ra64sp for schedule strategy
package chip8node8ra64sp

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// all the strategies in this file we define
var strategyMap map[strategyKey]scheduleStrategy

type strategyFactoryCallBack func(*chip8node8ra64sp, []string, map[string][]plugin.SuperNode)

var strategyInitFactory strategyFactoryCallBack

// init all strategies which will be used
func newScheduleStrategy() {
	schedule := &strategy{}
	rackSchedule := &oneRackStrategy{
		strategy: schedule,
		name:     RackSchedule,
	}
	uBMemSchedule := &oneUBMemStrategy{
		strategy: schedule,
		name:     UBMemSchedule,
	}
	superPodSchedule := &oneSuperPodStrategy{
		strategy: schedule,
		name:     SuperPodSchedule,
	}
	mulSuperPodsSchedule := &mulSuperPodsStrategy{
		strategy: schedule,
		name:     MulSuperPodsSchedule,
	}
	strategyMap = map[strategyKey]scheduleStrategy{
		RackSchedule:         rackSchedule,
		UBMemSchedule:        uBMemSchedule,
		SuperPodSchedule:     superPodSchedule,
		MulSuperPodsSchedule: mulSuperPodsSchedule,
	}

	strategyInitFactory = func(handler *chip8node8ra64sp, unReadyIds []string,
		selectedNodes map[string][]plugin.SuperNode) {
		schedule.chip8node8ra64sp = handler
		schedule.unReadyIds = unReadyIds
		schedule.selectedNodes = selectedNodes
	}
}

// entrySelect begin to select in one rack.
// if return true means it should try the next scheduling strategy
func (tp *oneRackStrategy) entrySelect(superPodTopo *superPodsInfo) (bool, error) {
	if tp == nil {
		return false, errors.New("strategy is nil")
	}

	if superPodTopo.spCount < len(tp.unReadyIds) && !tp.isSoftSuperPodAffinity {
		return false, fmt.Errorf("before scheduling, required sp-block count=%d, but total sp-block count=%d",
			len(tp.unReadyIds), superPodTopo.spCount)
	}
	tp.totalCount = len(tp.unReadyIds)
	requireTpBlockCount := tp.NPUTaskNum / tp.tpBlock

	tp.searchTable(superPodTopo.superPodTable)

	// select success
	if tp.totalCount == 0 && len(tp.selectedNodes) != 0 {
		klog.V(util.LogInfoLev).Info("select nodes in one Rack success.")
		return false, nil
	}

	// select failed and do not need to try the next strategy
	if len(tp.selectedNodes) == 0 && requireTpBlockCount == 1 && !tp.isSoftSuperPodAffinity {
		return false, fmt.Errorf("not found the 1 count of tp-block:%d in all racks of every super-pod, "+
			"exit select process", tp.TpBlockNPUNum)
	}

	return true, fmt.Errorf("scheduling failed in %s strategy", RackSchedule)
}

func (tp *oneRackStrategy) searchTable(superPodTable superPodOrderTable) {
	if tp.totalCount == 0 {
		return
	}
	row, col := getPositionInTop(tp.NPUTaskNum, tp.spBlock)

	for j := col; j < len(superPodTable[0]); j++ {
		for i := row; i < len(superPodTable); i++ {
			if len(superPodTable[i][j]) == 0 {
				continue
			}
			tp.iterateEverySuperPod(superPodTable[i][j])

			if tp.totalCount == 0 {
				return
			}
		}
		row = 0
	}
}

// iterateEverySuperPod Iterate every superPod
func (tp *oneRackStrategy) iterateEverySuperPod(superPods []superPod) {
	for i := 0; i < len(superPods); i++ {
		// init rack topo from superPod
		rackGroup := transferSuperPodToRackIdMap(superPods[i])

		tp.selectOneSpBlock(rackGroup)
		if tp.totalCount == 0 {
			return
		}
	}
	return
}

func (tp *oneRackStrategy) selectOneSpBlock(rackGroup map[int32][]nodeBaseInfo) {
	bestRackId := int32(1)
	bestRackIdLen := rackNodeNum + 1
	// finding the best rackId will be selected
	for rackId, nodes := range rackGroup {
		if len(nodes) == tp.NPUTaskNum {
			bestRackId = rackId
			bestRackIdLen = len(nodes)
			break
		}
		if len(nodes) > tp.NPUTaskNum && bestRackIdLen > len(nodes) {
			bestRackId = rackId
			bestRackIdLen = len(nodes)
		}
	}
	if bestRackIdLen == rackNodeNum+1 {
		return
	}

	spIndex := 0
	count := 0
	for _, nNode := range rackGroup[bestRackId] {
		klog.V(util.LogInfoLev).Infof("select nNode %s, super-pod ID: %d, rack ID: %d",
			nNode.name, nNode.superPodID, nNode.rackID)
		if _, ok := tp.selectedNodes[tp.unReadyIds[spIndex]]; !ok {
			tp.selectedNodes[tp.unReadyIds[spIndex]] = make([]plugin.SuperNode, 0)
		}
		tp.selectedNodes[tp.unReadyIds[spIndex]] = append(tp.selectedNodes[tp.unReadyIds[spIndex]],
			plugin.SuperNode{
				Name:       nNode.name,
				SuperPodID: nNode.superPodID,
				RackID:     nNode.rackID,
			})
		count++

		// already selected a sp-block, skip to next sp-block
		if count > 0 && count%tp.spBlock == 0 {
			tp.totalCount--
			spIndex++
		}
		// already selected enough nodes, saving the rest of nodes
		if spIndex >= len(tp.unReadyIds) {
			if nodes, ok := rackGroup[bestRackId]; ok && nodes != nil {
				rackGroup[bestRackId] = nodes[count:]
			}
			return
		}
	}
}

func (tp *oneUBMemStrategy) entrySelect(superPodTopo *superPodsInfo) (bool, error) {
	if tp == nil {
		return false, fmt.Errorf("strategy is nil")
	}
	if tp.spBlock < tp.tpBlock {
		return false, fmt.Errorf("parameter tp-block(%d) could not be bigger than sp-block(%d)", tp.tpBlock, tp.spBlock)
	}
	klog.V(util.LogInfoLev).Infof("select nodes in one UBMem start.")
	tp.totalCount = len(tp.unReadyIds)

	tp.searchTable(superPodTopo.superPodTable)

	if tp.totalCount == 0 && len(tp.selectedNodes) != 0 {
		klog.V(util.LogInfoLev).Info("select nodes in one UBMem success.")
		return false, nil
	}

	return false, fmt.Errorf("scheduling failed in %s strategy", UBMemSchedule)
}

func (tp *oneUBMemStrategy) searchTable(superPodTable superPodOrderTable) {
	if tp.totalCount == 0 {
		return
	}
	row, col := getPositionInTop(tp.NPUTaskNum, tp.spBlock)
	for j := col; j < len(superPodTable[0]); j++ {
		for i := row; i < len(superPodTable); i++ {
			if len(superPodTable[i][j]) == 0 {
				continue
			}
			tp.iterateEverySuperPod(superPodTable[i][j])

			if tp.totalCount == 0 {
				return
			}
		}
		row = 0
	}
}

func (tp *oneUBMemStrategy) iterateEverySuperPod(superPods []superPod) {
	const notFoundID = -1
	var bestCount int
	var bestUBMemID int32 = notFoundID
	var index int
	for i := 0; i < len(superPods); i++ {
		nodesInUbMems := getNodesInUbMem(superPods[i])

		for uBMemId, nodes := range nodesInUbMems {
			// before we decide to select nodes in someone ubmem, we must check whether it can contain all the pods
			ok, nodesCount := tp.checkUBMemIsSatisfied(nodes)
			if !ok {
				continue
			}
			if bestUBMemID == -1 || bestCount > nodesCount {
				bestUBMemID = uBMemId
				bestCount = nodesCount
			}
		}
		// find the best choice in the superPod
		if bestUBMemID > notFoundID {
			klog.V(util.LogInfoLev).Infof("start select nodes in the ubmem, ubmem id: %d.", bestUBMemID)
			tp.selectAllInOneUBMem(nodesInUbMems[bestUBMemID], superPods[index])
			return
		}
	}
}

// check whether this ubmem can be selected and return total nodes
func (tp *oneUBMemStrategy) checkUBMemIsSatisfied(racks map[int32][]nodeBaseInfo) (bool, int) {
	if racks == nil {
		return false, 0
	}
	// total nodes count
	count := 0
	// the real available node count
	totalAvailableNodes := 0

	for rackId := range racks {
		count += len(racks[rackId])
		totalAvailableNodes += (len(racks[rackId]) / tp.tpBlock) * tp.tpBlock
	}

	return totalAvailableNodes >= tp.spBlock*tp.totalCount, count
}

// select total sp-block in this ubmem
func (tp *oneUBMemStrategy) selectAllInOneUBMem(rackTopoInUbMem map[int32][]nodeBaseInfo, superPod superPod) {
	maxTimes := len(tp.unReadyIds)
	curTimes := 0
	for curTimes <= maxTimes {
		curTimes++
		if tp.totalCount == 0 {
			return
		}
		if tp.totalCount-1 >= len(tp.unReadyIds) {
			klog.V(util.LogErrorLev).Infof("index out of range, totalCount: %d, unReadyIds: %d",
				tp.totalCount-1, len(tp.unReadyIds))
			return
		}
		rackTopoInUbMem = tp.selectOneSpBlock(rackTopoInUbMem, superPod)
		if tp.totalCount == 0 {
			return
		}
	}
	klog.V(util.LogErrorLev).Info("select nodes failed by iterating ubmem but it should be success")
}

// entrySelect begin to select in one superPod
func (tp *oneSuperPodStrategy) entrySelect(superPodTopo *superPodsInfo) (bool, error) {
	if tp == nil {
		return false, errors.New("strategy is nil")
	}

	if tp.spBlock < tp.tpBlock {
		return false, fmt.Errorf("parameter tp-block(%d) could not be bigger than sp-block(%d)", tp.tpBlock, tp.spBlock)
	}
	klog.V(util.LogInfoLev).Info("select nodes in one superpod start")
	tp.totalCount = len(tp.unReadyIds)

	tp.searchTable(superPodTopo.superPodTable)

	if tp.totalCount == 0 && len(tp.selectedNodes) != 0 {
		klog.V(util.LogInfoLev).Info("select nodes in one superpod success.")
		return false, nil
	}

	return true, fmt.Errorf("scheduling failed in %s strategy", SuperPodSchedule)
}

func (tp *oneSuperPodStrategy) searchTable(superPodTable superPodOrderTable) {
	if tp.totalCount == 0 {
		return
	}
	row, col := getPositionInTop(tp.NPUTaskNum, tp.spBlock)
	for j := col; j < len(superPodTable[0]); j++ {
		for i := row; i < len(superPodTable); i++ {
			if len(superPodTable[i][j]) == 0 {
				continue
			}

			tp.iterateEverySuperPod(superPodTable[i][j])

			if tp.totalCount == 0 {
				return
			}
		}
		row = 0
	}
}

// iterateEverySuperPod Iterate every superPod
func (tp *oneSuperPodStrategy) iterateEverySuperPod(superPods []superPod) {
	maxLoopTimes := len(tp.unReadyIds)
	time := 0
	for i := 0; i < len(superPods); i++ {
		rackGroup := transferSuperPodToRackIdMap(superPods[i])

		// check this superPod can be scheduled
		if !tp.checkSuperPodIsSatisfied(rackGroup) {
			continue
		}
		for time <= maxLoopTimes {
			time++
			// select one sp-block each loop time
			tp.selectOneSpBlock(rackGroup, superPods[i])
			if tp.totalCount == 0 {
				return
			}
		}
		klog.V(util.LogErrorLev).Info("select nodes failed by iterating superPod but it should be success")
		break
	}
	return
}

// checkSuperPodIsSatisfied whether this superPod can be scheduled
func (tp *oneSuperPodStrategy) checkSuperPodIsSatisfied(rackGroup map[int32][]nodeBaseInfo) bool {
	tpBlockNum := 0
	for rackId := range rackGroup {
		// how many tpBlock in this superPod
		tpBlockNum += len(rackGroup[rackId]) / tp.tpBlock
		if tpBlockNum >= tp.totalCount*tp.spBlock/tp.tpBlock {
			return true
		}
	}
	return false
}

// entrySelect begin to select in multiple superPod
func (tp *mulSuperPodsStrategy) entrySelect(superPodTopo *superPodsInfo) (bool, error) {
	if tp == nil {
		return false, errors.New("strategy is nil")
	}

	if tp.spBlock < tp.tpBlock {
		return false, fmt.Errorf("parameter tp-block(%d) could not be bigger than sp-block(%d)", tp.tpBlock, tp.spBlock)
	}
	klog.V(util.LogInfoLev).Infof("select nodes in multiple superpods start.")
	tp.totalCount = len(tp.unReadyIds)

	tp.searchTable(superPodTopo.superPodTable)

	if tp.totalCount == 0 && len(tp.selectedNodes) != 0 {
		klog.V(util.LogInfoLev).Infof("select nodes in multiple superpods success.")
		return true, nil
	}

	// to break sp tp limit, use soft schedule
	if tp.isSoftSuperPodAffinity {
		former := tp.getSelectedNodesNum()
		needNode := tp.NPUTaskNum - former
		klog.V(util.LogWarningLev).Infof("multiple superpods schedule failed, job <%s> will scheduling as"+
			" soft strategy", tp.Name)
		last := tp.selectFromSuperPodsWithSoftStrategy(superPodTopo.superPodTable, needNode)
		if last == 0 {
			return false, nil
		} else {
			return true, fmt.Errorf("soft schedule job failed, need <%d> nodes, could only schedule <%d> nodes",
				needNode, former-last)
		}
	}
	klog.V(util.LogErrorLev).Infof("schedule job failed, need <%d> spBlock, could only schedule <%d> sp-block",
		len(tp.unReadyIds), len(tp.unReadyIds)-tp.totalCount)
	return true, fmt.Errorf("scheduling failed in %s strategy", MulSuperPodsSchedule)
}

// getSelectedNodesNum
func (tp *mulSuperPodsStrategy) getSelectedNodesNum() int {
	amount := 0
	for _, sp := range tp.selectedNodes {
		amount += len(sp)
	}
	return amount
}

func (tp *mulSuperPodsStrategy) searchTable(superPodTable superPodOrderTable) {
	if tp.totalCount == 0 {
		return
	}
	row, col := getPositionInTop(tp.FrameAttr.SuperPodSize, tp.spBlock)

	for j := col; j >= 0; j-- {
		for i := row; i >= 0; i-- {
			if len(superPodTable[i][j]) == 0 {
				continue
			}

			tp.iterateEverySuperPod(superPodTable[i][j])

			if tp.totalCount == 0 {
				return
			}
		}
		row = len(superPodTable) - 1
	}
}

// iterateEverySuperPod Iterate every SuperPod
func (tp *mulSuperPodsStrategy) iterateEverySuperPod(superPods []superPod) {
	for i := 0; i < len(superPods); i++ {
		rackGroup := transferSuperPodToRackIdMap(superPods[i])
		for {
			if !tp.checkSuperPodIsSatisfied(rackGroup) {
				break
			}

			tp.selectOneSpBlock(rackGroup, superPods[i])
			if tp.totalCount == 0 {
				return
			}
		}
	}
	return
}

// checkSuperPodIsSatisfied whether this superPod can be scheduled
func (tp *mulSuperPodsStrategy) checkSuperPodIsSatisfied(rackGroup map[int32][]nodeBaseInfo) bool {
	tpBlockNum := 0
	for rackId := range rackGroup {
		tpBlockNum += len(rackGroup[rackId]) / tp.tpBlock

		// here is different from check func in one SuperPod strategy
		if tpBlockNum >= tp.spBlock/tp.tpBlock {
			return true
		}
	}
	return false
}

func (tp *mulSuperPodsStrategy) selectFromSuperPodsWithSoftStrategy(superPodTable superPodOrderTable, remainingNodesSelecting int) int {
	row, col := getPositionInTop(tp.FrameAttr.SuperPodSize, tp.spBlock)
	recorder := &vPodIdRecorder{unReadyId: tp.unReadyIds, leftIndex: tp.totalCount - 1, rightIndex: tp.NPUTaskNum}
	for j := col; j >= 0; j-- {
		for i := row; i >= 0; i-- {
			if len(superPodTable[i][j]) == 0 {
				continue
			}
			superPodTable[i][j], remainingNodesSelecting = tp.IterateEverySuperPodWithoutFilter(superPodTable[i][j], remainingNodesSelecting, recorder)
			if remainingNodesSelecting == 0 {
				return remainingNodesSelecting
			}
		}
		row = len(superPodTable) - 1
	}
	return remainingNodesSelecting
}

// IterateEverySuperPodWithoutFilter soft ver of IterateEverySuperPod
func (tp *mulSuperPodsStrategy) IterateEverySuperPodWithoutFilter(superPods []superPod, remainingNodesSelecting int, recorder *vPodIdRecorder) ([]superPod, int) {
	for i := 0; i < len(superPods); i++ {
		selectedOneSp := 0
		rackGroup := transferSuperPodToRackIdMap(superPods[i])
		rackIdGroup := sortRackIdByLengthInOneSuperPod(rackGroup)
		superPods[i], selectedOneSp = tp.doNoFilterSelect(superPods[i], rackIdGroup, remainingNodesSelecting, recorder)
		remainingNodesSelecting -= selectedOneSp
	}
	return superPods, remainingNodesSelecting
}

type vPodIdRecorder struct {
	unReadyId  []string
	leftIndex  int
	rightIndex int
}

func (r *vPodIdRecorder) getVPodID() string {
	if r.leftIndex >= len(r.unReadyId) {
		return ""
	}
	if r.leftIndex < 0 {
		ans := strconv.Itoa(r.rightIndex)
		r.rightIndex--
		return ans
	}
	ans := r.unReadyId[r.leftIndex]
	r.leftIndex--
	return ans
}

// doNoFilterSelect soft ver of doSelect
func (tp *mulSuperPodsStrategy) doNoFilterSelect(superPod map[string]nodeBaseInfo,
	rackIdGroup []int32, nodeNeed int, recorder *vPodIdRecorder) (map[string]nodeBaseInfo, int) {
	count := 0
	reserveNode := make(map[string]nodeBaseInfo, len(superPod)-tp.spBlock)
	if tp.selectedNodes == nil {
		klog.V(util.LogErrorLev).Infof("inner error with selected nodes")
		return superPod, count
	}
	superPodWithRackId := transferSuperPodToRackIdMap(superPod)
	for _, rackId := range rackIdGroup {
		if len(superPodWithRackId[rackId]) == 0 {
			continue
		}
		spIndex := recorder.getVPodID()
		for _, nNode := range superPodWithRackId[rackId] {
			if count >= nodeNeed {
				reserveNode[nNode.name] = nNode
				continue
			}
			klog.V(util.LogInfoLev).Infof("select nNode %s, super-pod ID: %d, rack ID: %d",
				nNode.name, nNode.superPodID, nNode.rackID)
			_, ok := tp.selectedNodes[spIndex]
			if !ok {
				klog.V(util.LogErrorLev).Infof("inner error with selected nodes")
				tp.selectedNodes[spIndex] = make([]plugin.SuperNode, 0)
			}
			tp.selectedNodes[spIndex] = append(tp.selectedNodes[spIndex], plugin.SuperNode{
				Name:       nNode.name,
				SuperPodID: nNode.superPodID,
				RackID:     nNode.rackID,
			})
			count++
		}
	}
	if count == 0 {
		klog.V(util.LogInfoLev).Infof("select nNode in Racks failed")
		return superPod, count
	}
	tp.totalCount--
	return reserveNode, count
}
