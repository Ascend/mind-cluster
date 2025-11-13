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

// Package superpod for schedule strategy
package superpod

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// init all strategies which will be used
func newScheduleStrategy(handler *module910a5SuperPod, unReadyIds []string,
	selectedNodes map[string][]plugin.SuperNode) scheduleStrategy {
	schedule := &strategy{
		module910a5SuperPod: handler,
		selectedNodes:       selectedNodes,
		unReadyIds:          unReadyIds,
	}
	rackSchedule := &oneRackStrategy{
		strategy: *schedule,
	}
	superPodSchedule := &oneSuperPodStrategy{
		strategy: *schedule,
	}
	mulSuperPodsSchedule := &mulSuperPodsStrategy{
		strategy: *schedule,
	}
	rackSchedule.nextStrategy = superPodSchedule

	superPodSchedule.nextStrategy = mulSuperPodsSchedule
	handler.chooseWhichStrategyByNPUTaskNum()
	switch handler.getScheduleStrategy() {
	case RackSchedule:
		return rackSchedule
	case SuperPodSchedule:
		return superPodSchedule
	case MulSuperPodsSchedule:
		return mulSuperPodsSchedule
	default:
		return superPodSchedule
	}
}

// entrySelect begin to select in one rack
func (tp *oneRackStrategy) entrySelect(superPodTopo *superPodsInfo) (bool, error) {
	if tp == nil {
		return false, errors.New("strategy is nil")
	}
	if superPodTopo.spCount < len(tp.unReadyIds) && !tp.isSoftSuperPodAffinity {
		return false, fmt.Errorf("select super pod failed, required sp-block count=%d, but total sp-block count=%d",
			len(tp.unReadyIds), superPodTopo.spCount)
	}
	tp.totalCount = len(tp.unReadyIds)
	requireTpBlockCount := tp.NPUTaskNum / tp.tpBlock

	tp.searchTable(superPodTopo.superPodTable)

	if len(tp.selectedNodes) == 0 && requireTpBlockCount == 1 && !tp.isSoftSuperPodAffinity {
		return false, fmt.Errorf("not found the %d count of the tp-block:%d in all racks of every super-pod,"+
			" exit select process", requireTpBlockCount, tp.TpBlockNPUNum)
	}

	if len(tp.selectedNodes) == 0 && tp.TpBlockNPUNum >= npuNumber8 && !tp.isSoftSuperPodAffinity {
		klog.V(util.LogWarningLev).Infof("try to schedule job in multiple rack in one superpod " +
			"after scheduling in one rack failed,")
	}

	// select success
	if tp.totalCount == 0 && len(tp.selectedNodes) != 0 {
		return true, nil
	}

	// select failed, then choose to be superPod schedule strategy or ub
	tp.inNextStrategy = true
	tp.scheduleStrategy = SuperPodSchedule
	return tp.nextStrategy.entrySelect(superPodTopo)
}

// handleSelectResult handle the one rack select result
func (tp *oneRackStrategy) handleSelectResult(jobName string, ok bool, err error) error {
	if tp.inNextStrategy {
		return tp.nextStrategy.handleSelectResult(jobName, ok, err)
	}
	if err != nil {
		klog.V(util.LogErrorLev).Infof("Job: %s, select nodes in one Rack failed. error info: [%s]. ",
			jobName, util.SafePrint(err))
		return err
	}
	if ok {
		// select success
		klog.V(util.LogInfoLev).Infof("Job: %s, select nodes in one Rack success.", jobName)
		return nil
	}
	return err
}

func (tp *oneRackStrategy) searchTable(superPodTable superPodOrderTable) {
	row, col := getPositionInTop(tp.NPUTaskNum, tp.spBlock)

	for j := col; j < len(superPodTable[0]); j++ {
		for i := row; i < len(superPodTable); i++ {
			if tp.totalCount == 0 {
				return
			}
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

		tp.doSelect(rackGroup)
		if tp.totalCount == 0 {
			return
		}
	}
	return
}

func (tp *oneRackStrategy) doSelect(rackGroup map[int32][]nodeBaseInfo) {
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
			rackGroup[bestRackId] = rackGroup[bestRackId][count:]
			return
		}
	}
}

// entrySelect begin to select in one superPod
func (tp *oneSuperPodStrategy) entrySelect(superPodTopo *superPodsInfo) (bool, error) {
	if tp == nil {
		return false, errors.New("strategy is nil")
	}
	if tp.spBlock < tp.tpBlock {
		return false, fmt.Errorf("parameter tp-block(%d) could not be bigger than sp-block(%d)", tp.tpBlock, tp.spBlock)
	}
	klog.V(util.LogInfoLev).Infof("select nodes in one superpod start.")
	tp.totalCount = len(tp.unReadyIds)

	tp.searchTable(superPodTopo.superPodTable)

	if tp.totalCount == 0 && len(tp.selectedNodes) != 0 {
		return true, nil
	}

	// select nodes failed, choose to Multiple superPod strategy
	tp.inNextStrategy = true
	tp.scheduleStrategy = MulSuperPodsSchedule
	return tp.nextStrategy.entrySelect(superPodTopo)
}

// handleSelectResult handle the one superpod select result
func (tp *oneSuperPodStrategy) handleSelectResult(jobName string, ok bool, err error) error {
	if tp.inNextStrategy {
		return tp.nextStrategy.handleSelectResult(jobName, ok, err)
	}

	if err != nil {
		klog.V(util.LogErrorLev).Infof("Job: %s, select nodes in one superpod failed. error info: [%s]. ",
			jobName, util.SafePrint(err))
		return err
	}

	if ok {
		klog.V(util.LogInfoLev).Infof("Job: %s, select nodes in one superpod success.", jobName)
		return nil
	}
	return err
}

func (tp *oneSuperPodStrategy) searchTable(superPodTable superPodOrderTable) {
	row, col := getPositionInTop(tp.NPUTaskNum, tp.spBlock)
	for j := col; j < len(superPodTable[0]); j++ {
		for i := row; i < len(superPodTable); i++ {
			if tp.totalCount == 0 {
				return
			}
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
			tp.doSelect(rackGroup, superPods[i])
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
			return true, nil
		} else {
			return false, fmt.Errorf("soft schedule job failed, need <%d> nodes, could only schedule <%d> nodes",
				needNode, former-last)
		}
	}

	return false, fmt.Errorf("schedule job failed, need <%d> spBlock, could only schedule <%d> sp-block",
		len(tp.unReadyIds), len(tp.unReadyIds)-tp.totalCount)
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

// handleSelectResult handle multiple superpods select result
func (tp *mulSuperPodsStrategy) handleSelectResult(jobName string, ok bool, err error) error {
	if err != nil {
		klog.V(util.LogErrorLev).Infof("Job: %s, select nodes in multiple superpods failed. error info: [%s]. ",
			jobName, util.SafePrint(err))
		return err
	}
	if ok {
		klog.V(util.LogInfoLev).Infof("Job: %s, select nodes in multiple superpods success.", jobName)
		return nil
	}
	return err
}

// iterateEverySuperPod Iterate every SuperPod
func (tp *mulSuperPodsStrategy) iterateEverySuperPod(superPods []superPod) {
	for i := 0; i < len(superPods); i++ {
		rackGroup := transferSuperPodToRackIdMap(superPods[i])
		for {
			if !tp.checkSuperPodIsSatisfied(rackGroup) {
				break
			}

			tp.doSelect(rackGroup, superPods[i])
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
