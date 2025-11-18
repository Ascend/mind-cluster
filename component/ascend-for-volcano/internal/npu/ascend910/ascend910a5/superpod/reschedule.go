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

// Package superpod for reschedule functions
package superpod

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// initialize nodes and make selection
func (tp *module910a5SuperPod) selectNodesForFaultJob(task *api.TaskInfo, totalNodes map[int32]superPod,
	spBlockIDs map[string]bool, selectNodes map[string][]plugin.SuperNode) error {
	// get fault job, may be no exist
	fJob := tp.getFaultJob(task)
	if fJob == nil || !fJob.IsFaultJob {
		return nil
	}
	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes %s: is fault job, superPods: %v",
		tp.GetPluginName(), fJob.JobName, fJob.SuperPods)
	if !tp.isPodLevelRescheduling(fJob) {
		klog.V(util.LogInfoLev).Infof("job level rescheduling start")
		return nil
	}

	// refuse rescheduling till grace deletion finished.
	for _, fTask := range fJob.FaultTasks {
		if fTask.IsBeingGracefulDeleted {
			klog.V(util.LogWarningLev).Infof("rescheduling: pod <%s> is being graceful delete, unable to reschedule",
				fTask.TaskName)
			return fmt.Errorf("pod <%s> is being graceful delete, unable to reschedule", fTask.TaskName)
		}
	}

	// select nodes in original spBlocks, maybe failed
	notReadySpBlock, err := tp.selectNodeFromOriginSpBlock(fJob, selectNodes, totalNodes, spBlockIDs)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("the selecting node from original spBlocks err: %v", err)
		return err
	}
	klog.V(util.LogInfoLev).Infof("tpBlock=%d,WhetherBackToVspSchedule=%v,PendingSessionNum=%d,"+
		"notReadySpBlock=%v", tp.tpBlock, fJob.WhetherBackToVspSchedule,
		fJob.PendingSessionNum, notReadySpBlock)

	if !fJob.WhetherBackToVspSchedule {
		return tp.selectNodesByRack(fJob, notReadySpBlock, totalNodes, spBlockIDs, selectNodes)
	}

	klog.V(util.LogInfoLev).Infof("selectNodesForFaultJob finished, tp.whetherBackToVspSchedule is %v, "+
		"reset to: false", tp.whetherBackToVspSchedule)
	tp.whetherBackToVspSchedule = false
	return nil
}

func (tp *module910a5SuperPod) getFaultJob(task *api.TaskInfo) *rescheduling.FaultJob {
	rescheduleCache := rescheduling.GetReSchedulerCache()
	if rescheduleCache == nil || rescheduleCache.FaultJobs == nil {
		klog.V(util.LogInfoLev).Infof("getFaultJob faultJobs are empty")
		return nil
	}
	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes %s: reScheduler is not nil", tp.GetPluginName(), task.Name)
	fJob := rescheduleCache.FaultJobs[task.Job]
	return fJob
}

// select node for the rest of fault job by rack
func (tp *module910a5SuperPod) selectNodesByRack(fJob *rescheduling.FaultJob,
	notReadySuperPod map[string]struct{}, totalNodes map[int32]superPod,
	spBlockIDs map[string]bool, selectNodes map[string][]plugin.SuperNode) error {
	if selectNodes == nil || spBlockIDs == nil {
		return nil
	}

	klog.V(util.LogInfoLev).Infof("selectNodesByRack start")
	defer klog.V(util.LogInfoLev).Infof("selectNodesByRack end")

	faultNodeNameMap := getFaultNodeNameMap(fJob, tp.tpBlock)
	klog.V(util.LogInfoLev).Infof("faultNodeNameMap:%v", faultNodeNameMap)

	for vSuperPodId := range notReadySuperPod {
		superNode := fJob.SuperPods[vSuperPodId][0]
		if fJob.PendingSessionNum < FirstRescheduleStage && tp.Jobs[fJob.JobUID].Label[util.ProcessRecoverEnable] !=
			util.EnableFunc {
			selectNodes[vSuperPodId] = getSameRackNodes(faultNodeNameMap, fJob.SuperPods[vSuperPodId],
				totalNodes[superNode.SuperPodID])
			spBlockIDs[vSuperPodId] = true
			continue
		}
		superPodWithRackId := transferSuperPodToRackIdMap(totalNodes[superNode.SuperPodID])
		selectRackId, err := tp.getRackId(superPodWithRackId, faultNodeNameMap, fJob.SuperPods[vSuperPodId], fJob)
		if err != nil {
			return err
		}
		selectNodes[vSuperPodId] = getAnotherRackNodes(faultNodeNameMap, fJob.SuperPods[vSuperPodId],
			totalNodes[superNode.SuperPodID], superPodWithRackId[selectRackId])
		spBlockIDs[vSuperPodId] = true
	}
	if !tp.checkSelectedNodesValid(selectNodes) {
		klog.V(util.LogErrorLev).Infof("selected nodes %v is invalid", selectNodes)
		return fmt.Errorf("selected nodes %v is invalid", selectNodes)
	}
	return nil
}

func (tp *module910a5SuperPod) checkSelectedNodesValid(selectedNodes map[string][]plugin.SuperNode) bool {
	for _, nodes := range selectedNodes {
		if len(nodes) < tp.spBlock {
			return false
		}
	}
	return true
}

// select node for fault job in another rack
func getAnotherRackNodes(faultNodeNameMap map[string]struct{},
	vSuperPod []plugin.SuperNode, totalNodes map[string]nodeBaseInfo,
	superPodWithRackId []nodeBaseInfo) []plugin.SuperNode {
	if len(superPodWithRackId) == 0 {
		return nil
	}

	klog.V(util.LogInfoLev).Infof("getAnotherRackNodes start")
	defer klog.V(util.LogInfoLev).Infof("getAnotherRackNodes end")

	newNodes := make([]plugin.SuperNode, 0)
	for _, nodeOfFJob := range vSuperPod {
		if _, ok := faultNodeNameMap[nodeOfFJob.Name]; !ok {
			newNodes = append(newNodes, nodeOfFJob)
			continue
		}

		if len(superPodWithRackId) <= 0 {
			continue
		}

		nodeOfAnotherRack := superPodWithRackId[0]
		newNodes = append(newNodes, plugin.SuperNode{
			Name:       nodeOfAnotherRack.name,
			SuperPodID: nodeOfAnotherRack.superPodID,
			RackID:     nodeOfAnotherRack.rackID,
		})
		delete(totalNodes, nodeOfAnotherRack.name)
		superPodWithRackId = superPodWithRackId[1:]
	}
	klog.V(util.LogInfoLev).Infof("getAnotherRackNodes newNodes: %v", newNodes)

	return newNodes
}

func (tp *module910a5SuperPod) backToVspSchedule(fJob *rescheduling.FaultJob) error {
	const errInsufficientNodes = "there is no enough nodes for whole rack schedule"
	klog.V(util.LogErrorLev).Infof(errInsufficientNodes)

	klog.V(util.LogInfoLev).Infof("back to vsp schedule, tp.whetherBackToVspSchedule is %v, set to true",
		tp.whetherBackToVspSchedule)
	tp.whetherBackToVspSchedule = true

	return errors.New(errInsufficientNodes)
}

func (tp *module910a5SuperPod) getRackId(superPodWithRackId map[int32][]nodeBaseInfo,
	faultNodeNameMap map[string]struct{}, vSuperPod []plugin.SuperNode, fJob *rescheduling.FaultJob) (int32, error) {
	filterRackIdByTpBlock(superPodWithRackId, tp.tpBlock)
	restRackLenMapId := getOriginRackId(superPodWithRackId, faultNodeNameMap, vSuperPod)
	if restRackLenMapId == UninitializedRestRackLenMapId {
		rackIdOrder := sortRackIdByLengthInOneSuperPod(superPodWithRackId)
		restRackLenMapId = tp.getRestRackId(rackIdOrder, superPodWithRackId)
		if restRackLenMapId == UninitializedRestRackLenMapId {
			return 0, tp.backToVspSchedule(fJob)
		}
	}
	return restRackLenMapId, nil
}

func (tp *module910a5SuperPod) getRestRackId(rackIdOrder []int32, superPodWithRackId map[int32][]nodeBaseInfo) int32 {
	for _, rackId := range rackIdOrder {
		if len(superPodWithRackId[rackId]) > tp.tpBlock {
			return rackId
		}
	}
	return UninitializedRestRackLenMapId
}

func (tp *module910a5SuperPod) getRestRackLenMapId(restRackLenMap map[int][]int32) int32 {
	for nodeNumOfRack := tp.tpBlock; nodeNumOfRack <= rackNodeNum; nodeNumOfRack++ {
		if _, ok := restRackLenMap[nodeNumOfRack]; ok {
			return restRackLenMap[nodeNumOfRack][0]
		}
	}
	return UninitializedRestRackLenMapId
}

// get fault node map
func getFaultNodeNameMap(fJob *rescheduling.FaultJob, tpBlock int) map[string]struct{} {
	faultNodeNameMap := map[string]struct{}{}
	if fJob == nil || fJob.FaultTasks == nil {
		return faultNodeNameMap
	}
	for _, faultTask := range fJob.FaultTasks {
		if faultTask.IsFaultTask || !judgeSatisfiedRackAffinity(fJob, faultTask.NodeName, tpBlock) {
			faultNodeNameMap[faultTask.NodeName] = struct{}{}
		}
	}
	return faultNodeNameMap
}

// select node for fault job by rack
func getSameRackNodes(faultNodeNameMap map[string]struct{}, vSuperPod []plugin.SuperNode,
	spNodeMaps map[string]nodeBaseInfo) []plugin.SuperNode {
	if spNodeMaps == nil {
		return nil
	}

	klog.V(util.LogInfoLev).Infof("getSameRackNodes start")
	defer klog.V(util.LogInfoLev).Infof("getSameRackNodes end")

	newNodes := make([]plugin.SuperNode, 0)
	for _, nodeOfFJob := range vSuperPod {
		if _, ok := faultNodeNameMap[nodeOfFJob.Name]; !ok {
			newNodes = append(newNodes, nodeOfFJob)
			spNodeMaps = updateSpNodeMaps(spNodeMaps, nodeOfFJob)
			continue
		}
		for _, nodeOfTotalNodes := range spNodeMaps {
			if faultNodeNameMap == nil || nodeOfTotalNodes.rackID != nodeOfFJob.RackID {
				continue
			}
			faultNodeNameMap[nodeOfTotalNodes.name] = struct{}{}
			newNodes = append(newNodes, plugin.SuperNode{
				Name:       nodeOfTotalNodes.name,
				SuperPodID: nodeOfTotalNodes.superPodID,
				RackID:     nodeOfTotalNodes.rackID,
			})
			delete(spNodeMaps, nodeOfTotalNodes.name)
			break
		}
	}
	klog.V(util.LogInfoLev).Infof("getSameRackNodes newNodes: %v", newNodes)

	return newNodes
}

func updateSpNodeMaps(spNodeMaps map[string]nodeBaseInfo,
	nodeOfFJob plugin.SuperNode) map[string]nodeBaseInfo {
	newSpNodeMaps := make(map[string]nodeBaseInfo)
	for id, nodeOfTotalNodes := range spNodeMaps {
		if nodeOfTotalNodes.name != nodeOfFJob.Name {
			newSpNodeMaps[id] = nodeOfTotalNodes
		}
	}
	return newSpNodeMaps
}

// judge whether using pod level reschedule
func (tp *module910a5SuperPod) isPodLevelRescheduling(fJob *rescheduling.FaultJob) bool {
	job, ok := tp.Jobs[fJob.JobUID]
	if !ok {
		return false
	}
	if fJob.PendingSessionNum > SecondRescheduleStage {
		return false
	}
	klog.V(util.LogInfoLev).Infof("label pod-rescheduling is: %s", job.Label[util.SinglePodTag])
	return job.Label[util.SinglePodTag] == util.EnableFunc && !fJob.IsMasterFault
}

// select node for fault job from origin vSuperPod
func (tp *module910a5SuperPod) selectNodeFromOriginSpBlock(fJob *rescheduling.FaultJob,
	selectNodes map[string][]plugin.SuperNode, totalNodes map[int32]superPod,
	spBlockIDs map[string]bool) (map[string]struct{}, error) {
	if selectNodes == nil || spBlockIDs == nil || len(spBlockIDs) == 0 {
		return nil, nil
	}
	notReadySuperPod := make(map[string]struct{})

	for vSuperPodId, superNodes := range fJob.SuperPods {
		count := 0
		for _, superNode := range superNodes {
			if judgeLasTimeTaskIsHealthy(fJob, superNode.Name) &&
				judgeSatisfiedRackAffinity(fJob, superNode.Name, tp.tpBlock) {
				count++
			}
		}
		if count < len(superNodes) {
			notReadySuperPod[vSuperPodId] = struct{}{}
			continue
		}
		klog.V(util.LogInfoLev).Infof("vSuperPodId: %s is satisfied vSuperPod: %v", vSuperPodId, superNodes)
		for _, superNode := range superNodes {
			delete(totalNodes[superNode.SuperPodID], superNode.Name)
		}
		selectNodes[vSuperPodId] = superNodes
		spBlockIDs[vSuperPodId] = true
	}
	return notReadySuperPod, nil
}

// check if task is healthy at last time
func judgeLasTimeTaskIsHealthy(fJob *rescheduling.FaultJob, nodeName string) bool {
	for _, task := range fJob.FaultTasks {
		if task.NodeName == nodeName {
			if task.IsFaultTask {
				return false
			}
			break
		}
	}
	return true
}

func judgeSatisfiedRackAffinity(fJob *rescheduling.FaultJob, nodeName string, tpBlock int) bool {
	if tpBlock <= tpBlock1 {
		// 未开启框亲和性或者还没到整框调度阶段，返回 true
		return true
	}
	for _, task := range fJob.FaultTasks {
		if task.NodeName != nodeName {
			continue
		}
		if task.IsSatisfiedRackAffinity {
			return true
		}
		break
	}

	return false
}
