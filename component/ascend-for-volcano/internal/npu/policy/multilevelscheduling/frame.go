/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package multilevelscheduling for scheduling NPU job with general abstract network topology configuration.
package multilevelscheduling

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"k8s.io/api/core/v1"
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/consts"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// rescheduleContext bundles historical allocation data resolved for rescheduling.
type rescheduleContext struct {
	task         *api.TaskInfo
	superPods    map[string][]plugin.SuperNode
	missingNodes []string
	fJob         *rescheduling.FaultJob
}

// New return npu plugin
func New(name string) base.AscendHandler {
	m := &MultilevelHandler{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetMaxNodeNPUNum(maxNodeNpu)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetIsNetworkFaultAttention(true)
	return m
}

// ValidNPUJob verify the validity of job parameters
func (mh *MultilevelHandler) ValidNPUJob() *api.ValidateResult {
	res := mh.checkTaskNPU()
	if res != nil {
		return res
	}
	return mh.checkLevels()
}

// sample task level  [{name: level1, reqNode: 2}, {name: level2, reqNode: 4}]
func (mh *MultilevelHandler) checkLevels() *api.ValidateResult {
	taskLevels, err := util.GetTaskTreeLevels(mh.AffinityBlocks, mh.NPUTaskNum-mh.CountBackupTasks())
	if err != nil {
		return &api.ValidateResult{
			Pass:    false,
			Reason:  blockInvalidReason,
			Message: err.Error(),
		}
	}
	mh.taskLevels = taskLevels
	return nil
}

// checkTaskNPU check the distributed job require npu num must equal node npu num
func (mh *MultilevelHandler) checkTaskNPU() *api.ValidateResult {
	for _, task := range mh.Tasks {
		if task.ReqNPUNum != 0 {
			continue
		}
		if task.ReqNPUNum == 0 && (task.Annotation[util.TaskSpecAnno] == util.SchedulerType ||
			task.Annotation[util.SkipAscendPluginAnno] == util.SkipEnabled) {
			continue
		}
		return &api.ValidateResult{
			Pass:    false,
			Reason:  jobCheckFailedReason,
			Message: fmt.Sprintf("distributed job require full node npu, instead of %d", task.ReqNPUNum),
		}
	}
	return nil
}

// CheckNodeNPUByTask check nod npu meet task req
func (mh *MultilevelHandler) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if mh == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %v", err)
		return err
	}

	if err := mh.checkNodeForHotSwitch(task, node); err != nil {
		return err
	}

	topo, exist := node.Label[util.TopoTreeLabel]
	if !exist {
		topo = util.DefaultTopoTree
	}
	// filter nodes with incorrect multilevel scheduling labels
	resourceLevels, configExist := mh.FrameAttr.ResourceLevelsInfo[topo]
	if !configExist {
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %v", util.TopoTreeLabelError)
		return errors.New(util.TopoTreeLabelError)
	}
	// filter nodes with complete labels
	for _, level := range resourceLevels {
		if level.Type == util.LevelTypeTree || level.Type == util.LevelTypeNode {
			continue
		}
		if _, ok := node.Label[level.Label]; !ok {
			klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %v", util.TopoTreeLabelError)
			return errors.New(util.TopoTreeLabelError)
		}
	}

	taskNPUNum, err := mh.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %s", mh.GetPluginName(), err.Error())
		return err
	}

	nodeTop, err := mh.GetUsableTopFromNode(node, true)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s getUsableTopFromNode err: %s", task.Name, err.Error())
		return err
	}

	if len(nodeTop) != taskNPUNum {
		klog.V(util.LogErrorLev).Infof("%s JudgeNodeAndTaskNPU err: %s", task.Name, nodeNpuNotMatchError)
		return fmt.Errorf("checkNodeNPUByTask %s err: %s", util.NodeNotMeetTopologyWarning, nodeNpuNotMatchError)
	}
	return nil
}

// ScoreBestNPUNodes get best nodes score for job
func (mh *MultilevelHandler) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo,
	sMap map[string]float64) error {
	if mh == nil || task == nil || len(nodes) == 0 || len(sMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %s.", err)
		return err
	}
	job, ok := mh.ScheduleEnv.Jobs[task.Job]
	if !ok {
		return fmt.Errorf("%s ScoreBestNPUNodes %s: job is not exist", mh.GetPluginName(), task.Name)
	}
	defer func() {
		mh.ScheduleEnv.Jobs[task.Job] = job
	}()
	if !*job.JobReadyTag {
		return nil
	}
	defer mh.selectNodeFromCache(&job, task, sMap)
	if mh.tryUseCachedSuperPods(&job, task, nodes) {
		return nil
	}
	if mh.NPUTaskNum > len(nodes) && mh.SchedulingTaskNum == len(mh.Tasks) {
		*job.JobReadyTag = false
		return fmt.Errorf("select node failed by not enough node")
	}
	selectedNodes, err := mh.selectNodesForMultiLevelJob(task, nodes)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("select nodes failed, %v", err)
		*job.JobReadyTag = false
		return err
	}
	klog.V(util.LogInfoLev).Infof("select nodes by multilevel policy successfully")
	// caching logic level1 group information to superpods structure
	*job.JobReadyTag = true
	job.SuperPods = selectedNodes
	job.SuperPodsVerified = true
	return nil
}

// tryUseCachedSuperPods validates cached SuperPods and returns true if the job
// can skip scheduling this session. SuperPodsVerified is set before validation
// so that subsequent pods in the same session skip the check and fall through
// to recompute only once. Stale SuperPods are kept for partial reschedule in
// tryScheduleTaskInSingleTree.
func (mh *MultilevelHandler) tryUseCachedSuperPods(job *plugin.SchedulerJob, task *api.TaskInfo,
	nodes []*api.NodeInfo) bool {
	if len(job.SuperPods) == 0 {
		return false
	}
	if job.SuperPodsVerified {
		return true
	}
	job.SuperPodsVerified = true
	if mh.isCachedSuperPodsValid(job, nodes) {
		return true
	}
	klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes %s: stale SuperPods, recomputing",
		mh.GetPluginName(), task.Name)
	return false
}

func (mh *MultilevelHandler) selectNodesForMultiLevelJob(task *api.TaskInfo,
	nodes []*api.NodeInfo) (map[string][]plugin.SuperNode, error) {
	var selectedNodes map[string][]plugin.SuperNode
	var err error
	const onlyL1ConfigLen = 3
	if len(mh.taskLevels) == onlyL1ConfigLen {
		// if a job's multi-level config has only level1 block, try padding l2 to job for better network performance
		selectedNodes, err = mh.tryScheduleInStrictRules(task, nodes)
		if err != nil {
			klog.V(util.LogInfoLev).Info("try scheduling all level1 in one level2 unit failed, back to normal")
			selectedNodes, err = mh.scheduleMultipleLevelPodsForJob(task, nodes)
		}
	} else {
		selectedNodes, err = mh.scheduleMultipleLevelPodsForJob(task, nodes)
	}
	return selectedNodes, err
}

func (mh *MultilevelHandler) tryScheduleInStrictRules(task *api.TaskInfo,
	nodes []*api.NodeInfo) (map[string][]plugin.SuperNode, error) {
	const insertIndex = 1
	originTaskLevel := mh.taskLevels
	paddingLevel2 := util.TaskTreeLevel{
		Name:    "level2",
		ReqNode: mh.taskLevels[0].ReqNode,
	}
	mh.taskLevels = append(originTaskLevel[:insertIndex],
		append([]util.TaskTreeLevel{paddingLevel2}, originTaskLevel[insertIndex:]...)...)
	selectedNodes, err := mh.scheduleMultipleLevelPodsForJob(task, nodes)
	if err == nil {
		return selectedNodes, nil
	}
	mh.taskLevels = append(mh.taskLevels[:insertIndex], mh.taskLevels[insertIndex+1:]...)
	return nil, errors.New("try scheduling all level1 in one level2 unit failed")
}

func (mh *MultilevelHandler) scheduleMultipleLevelPodsForJob(task *api.TaskInfo,
	nodes []*api.NodeInfo) (map[string][]plugin.SuperNode, error) {
	klog.V(util.LogInfoLev).Infof("[%s] input nodes num(%d) for task %s", mh.GetPluginName(), len(nodes), task.Name)
	rctx := mh.resolveSuperPodsForReschedule(task, nodes)
	if rctx == nil {
		return mh.scheduleFromAllNodes(task, nodes)
	}
	sm, err := mh.tryScheduleWithHistory(rctx, nodes)
	if err == nil {
		return sm, nil
	}
	return nil, err
}

// tryScheduleWithHistory attempts reschedule using historical SuperPods.
func (mh *MultilevelHandler) tryScheduleWithHistory(rctx *rescheduleContext,
	nodes []*api.NodeInfo) (map[string][]plugin.SuperNode, error) {
	if len(rctx.missingNodes) == 0 {
		klog.V(util.LogInfoLev).Infof("[%s] all historical nodes available for task %s, reuse directly",
			mh.GetPluginName(), rctx.task.Name)
		mh.SuperPodInfo.SuperPodMapFaultTaskNodes[rctx.task.Job] = map[string]string{}
		return rctx.superPods, nil
	}

	taskTree, err := mh.tryRescheduleWithHistory(rctx, nodes)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("[%s] tryRescheduleWithHistory failed: %v", mh.GetPluginName(), err)
		return nil, err
	}
	sm, getErr := plugin.GetSuperNodeMapFromTaskTree(taskTree, mh.getCachedJobSuperPods(rctx.task))
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("[%s] GetSuperNodeMapFromTaskTree failed: %v", mh.GetPluginName(), getErr)
		return nil, getErr
	}
	mh.SuperPodInfo.SuperPodMapFaultTaskNodes[rctx.task.Job] = map[string]string{}
	klog.V(util.LogInfoLev).Infof("[%s] reschedule succeeded for task %s", mh.GetPluginName(), rctx.task.Name)
	return sm, nil
}

// scheduleFromAllNodes builds resource trees from all healthy nodes and picks the
// best Schedule result across all topologies.
func (mh *MultilevelHandler) scheduleFromAllNodes(task *api.TaskInfo,
	nodes []*api.NodeInfo) (map[string][]plugin.SuperNode, error) {
	resourceTrees, err := plugin.GetResourceTrees(plugin.GetHealthyNPUNodes(mh.Nodes, nodes),
		mh.FrameAttr.ResourceLevelsInfo, mh.taskLevels)
	if err != nil {
		return nil, fmt.Errorf("[%s] GetResourceTrees failed: %v", mh.GetPluginName(), err)
	}

	var best *util.TaskTree
	for _, t := range resourceTrees {
		tt, e := Schedule(t, mh.taskLevels)
		if e != nil {
			continue
		}
		if best == nil || best.FragmentScore > tt.FragmentScore {
			best = tt
		}
	}
	if best == nil {
		return nil, fmt.Errorf("[%s] no valid task tree found for task %s", mh.GetPluginName(), task.Name)
	}

	sm, convErr := plugin.GetSuperNodeMapFromTaskTree(best, mh.getCachedJobSuperPods(task))
	if convErr != nil {
		return nil, convErr
	}
	mh.SuperPodInfo.SuperPodMapFaultTaskNodes[task.Job] = map[string]string{}
	return sm, nil
}

// getCachedJobSuperPods returns the cached SuperPods for the task's job, preferring
// FaultJob SuperPods over the job's own SuperPods for sortNodesByCachedRank.
func (mh *MultilevelHandler) getCachedJobSuperPods(task *api.TaskInfo) map[string][]plugin.SuperNode {
	if fJob, exist := getFaultJob(task.Job); exist && fJob.IsFaultJob {
		return fJob.SuperPods
	}
	if job, ok := mh.ScheduleEnv.Jobs[task.Job]; ok {
		return job.SuperPods
	}
	return nil
}

// tryRescheduleWithHistory builds resource trees from pre-filtered candidate nodes
// and runs Reschedule at escalating fault scopes. The caller should have already
// returned cached superPods directly when no nodes are missing.
func (mh *MultilevelHandler) tryRescheduleWithHistory(rctx *rescheduleContext,
	nodes []*api.NodeInfo) (*util.TaskTree, error) {
	maxEscalation := mh.getMaxEscalationLevel(rctx)
	klog.V(util.LogInfoLev).Infof("[%s] maxEscalation=%d for task %s, missingNodes=%v",
		mh.GetPluginName(), maxEscalation, rctx.task.Name, rctx.missingNodes)
	return mh.rescheduleWithSuperPods(rctx, nodes, maxEscalation)
}

// resolveSuperPodsForReschedule builds a rescheduleContext from either the fault
// rescheduling cache or the job's previous allocation. missingNodes are filtered
// against the current candidate node set so they represent truly unavailable nodes.
// Returns nil if no historical data is available.
func (mh *MultilevelHandler) resolveSuperPodsForReschedule(task *api.TaskInfo,
	nodes []*api.NodeInfo) *rescheduleContext {
	// Source 1: FaultJob — always active.
	fJob, exist := getFaultJob(task.Job)
	if exist && fJob.IsFaultJob {
		rctx := &rescheduleContext{task: task, superPods: fJob.SuperPods, fJob: fJob}
		var err error
		rctx.missingNodes, err = mh.getFaultNodes(fJob.JobUID)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("[%s] getFaultNodes failed for %s: %v",
				mh.GetPluginName(), task.Name, err)
		}
		rawCount := len(rctx.missingNodes)
		rctx.missingNodes = filterOutAvailableNodes(rctx.missingNodes, nodes)
		klog.V(util.LogInfoLev).Infof("[%s] FaultJob hit for task %s: groups=%d, missingNodes raw=%d filtered=%d",
			mh.GetPluginName(), task.Name, len(rctx.superPods), rawCount, len(rctx.missingNodes))
		return rctx
	}

	// Source 2: job's previous allocation — gated by PreferPreviousNode.
	if !mh.FrameAttr.PreferPreviousNode {
		return nil
	}
	job, ok := mh.ScheduleEnv.Jobs[task.Job]
	if !ok || job.Owner.UID == "" || len(job.SuperPods) == 0 {
		return nil
	}
	rctx := &rescheduleContext{task: task, superPods: job.SuperPods}
	rctx.missingNodes = mh.getMissingNodesFromJob(job.SuperPods, job.Tasks)
	rawCount := len(rctx.missingNodes)
	rctx.missingNodes = filterOutAvailableNodes(rctx.missingNodes, nodes)
	klog.V(util.LogInfoLev).Infof("[%s] previous allocation hit for task=%s, groups=%d, missingNodes raw=%d filtered=%d",
		mh.GetPluginName(), task.Name, len(rctx.superPods), rawCount, len(rctx.missingNodes))
	return rctx
}

// OnBackupPodAllocated implements plugin.BackupPodAllocatedHook.
// When a hot-switch backup pod is allocated to a replacement node, this updates
// the in-memory SuperPods so the cache reflects the backup pod's new node
// instead of the deleted fault pod's node.
func (mh *MultilevelHandler) OnBackupPodAllocated(task *api.TaskInfo, job *plugin.SchedulerJob, nodeName string) {
	rank, err := getHcclRankIndex(task, *job)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("[%s] OnBackupPodAllocated: getHcclRankIndex failed for %s: %v",
			mh.GetPluginName(), task.Name, err)
		return
	}
	logicL1Rank, localRank, err := getL1Ranks(job.SuperPods, rank)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("[%s] OnBackupPodAllocated: getL1Ranks failed for %s (rank=%d): %v",
			mh.GetPluginName(), task.Name, rank, err)
		return
	}
	group, ok := job.SuperPods[logicL1Rank]
	if !ok || localRank >= len(group) {
		klog.V(util.LogWarningLev).Infof("[%s] OnBackupPodAllocated: invalid position L1=%s localRank=%d for %s",
			mh.GetPluginName(), logicL1Rank, localRank, task.Name)
		return
	}
	oldName := group[localRank].Name
	group[localRank].Name = nodeName
	klog.V(util.LogInfoLev).Infof("[%s] OnBackupPodAllocated: updated SuperPods[%s][%d] %s→%s for backup pod %s",
		mh.GetPluginName(), logicL1Rank, localRank, oldName, nodeName, task.Name)
}

// getMaxEscalationLevel determines the maximum allowed escalation depth for
// rescheduleWithSuperPods, measured as len(faultSubTree.Levels) — 1 means pod-level,
// larger values mean broader group-level scope. expandFaultLevel returns the actual
// depth; if it exceeds maxEscalation, the expansion is rejected and the loop exits.
//
//	ReScheduleLimit == "pod":             0 (pod-level only, never escalate)
//	Other non-pod rescheduling (job-level): len(taskLevels) (allow any escalation)
//	Single/process rescheduling job:      escalates by session count:
//	  Rank-0 fault                         len(taskLevels)
//	  session < 6                          0 (pod-level)
//	  session 6-11                         1 (allow up to L1-group)
//	  session >= 12                        len(taskLevels) (allow full job)
func (mh *MultilevelHandler) getMaxEscalationLevel(rctx *rescheduleContext) int {
	fJob := rctx.fJob
	if fJob == nil || fJob.ReScheduleLimit == util.ReschedulingUpperLimitPod {
		return 0
	}
	if isFaultRankZero(rctx.task, mh.ScheduleEnv.Jobs) || fJob.PendingSessionNum >= util.PendingTimes ||
		(fJob.Labels[util.SinglePodTag] != util.EnableFunc && fJob.Labels[util.ProcessRecoverEnable] != util.EnableFunc) {
		return len(mh.taskLevels)
	}
	return fJob.PendingSessionNum / util.SpPendingTimes
}

// getResourceLevelsFromSuperPods returns the resource levels matching the topology
// of the first non-empty SuperNode.
func (mh *MultilevelHandler) getResourceLevelsFromSuperPods(
	superPods map[string][]plugin.SuperNode) ([]util.ResourceTreeLevel, error) {
	for _, nodes := range superPods {
		for _, node := range nodes {
			if node.TopoTreeName == "" {
				continue
			}
			if levels, ok := mh.FrameAttr.ResourceLevelsInfo[node.TopoTreeName]; ok {
				return levels, nil
			}
		}
	}
	return nil, fmt.Errorf("no matching resource levels found in superPods")
}

// rescheduleWithSuperPods builds filtered resource trees at each escalation level,
// excluding pinned (healthy historical) nodes so Reschedule can only place
// replacement pods on truly free nodes. At the last escalation level all nodes are
// unpinned, which is equivalent to a full Schedule() on all healthy nodes.
func (mh *MultilevelHandler) rescheduleWithSuperPods(rctx *rescheduleContext, nodes []*api.NodeInfo,
	maxEscalation int) (*util.TaskTree, error) {

	refLevels, err := mh.getResourceLevelsFromSuperPods(rctx.superPods)
	if err != nil {
		return nil, err
	}
	refTaskTree, err := plugin.GetTaskTreeFromSuperNodeMap(rctx.superPods, mh.taskLevels, refLevels, mh.Nodes)
	if err != nil {
		return nil, fmt.Errorf("[%s] build refTaskTree from historical superPods failed: %v", mh.GetPluginName(), err)
	}

	pinnedNodes := getHistoricalHealthyNodeNames(rctx.superPods, rctx.missingNodes)
	for {
		taskTree, err := plugin.GetTaskTreeFromSuperNodeMap(rctx.superPods, mh.taskLevels, refLevels, mh.Nodes)
		if err != nil {
			return nil, fmt.Errorf("[%s] build taskTree failed: %v", mh.GetPluginName(), err)
		}
		resultTree, err := mh.tryRescheduleOnFilteredTrees(nodes, taskTree, rctx.missingNodes, pinnedNodes)
		if err == nil {
			return resultTree, nil
		}
		newNodes, expandedLevel, didExpand := expandFaultLevel(refTaskTree, rctx.missingNodes)
		if expandedLevel > maxEscalation || !didExpand {
			break
		}
		rctx.missingNodes = newNodes
		pinnedNodes = removeAll(pinnedNodes, rctx.missingNodes)
	}
	return nil, fmt.Errorf("rescheduleWithSuperPods: all escalation levels exhausted")
}

func (mh *MultilevelHandler) tryRescheduleOnFilteredTrees(
	nodes []*api.NodeInfo, taskTree *util.TaskTree,
	missingNodes []string, pinnedNodes map[string]struct{}) (*util.TaskTree, error) {

	filtered := plugin.GetHealthyNPUNodes(mh.Nodes, nodes)
	for name := range pinnedNodes {
		delete(filtered, name)
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no nodes left after excluding pinned")
	}

	trees, getErr := plugin.GetResourceTrees(filtered, mh.FrameAttr.ResourceLevelsInfo, mh.taskLevels)
	if getErr != nil || len(trees) == 0 {
		return nil, fmt.Errorf("GetResourceTrees after filtering failed: %v", getErr)
	}

	for _, tree := range trees {
		resultTree, err := Reschedule(tree, taskTree, missingNodes)
		if err == nil {
			klog.V(util.LogInfoLev).Infof("[%s] reschedule succeeded on topology %s", mh.GetPluginName(), tree.Name)
			return resultTree, nil
		}
	}
	return nil, fmt.Errorf("reschedule failed on all topologies")
}

// --- helper methods for rescheduleWithSuperPods ---

// getHistoricalHealthyNodeNames returns the set of node names in superPods that are NOT in missingNodes.
func getHistoricalHealthyNodeNames(superPods map[string][]plugin.SuperNode,
	missingNodes []string) map[string]struct{} {
	faultSet := make(map[string]struct{}, len(missingNodes))
	for _, n := range missingNodes {
		faultSet[n] = struct{}{}
	}
	pinned := make(map[string]struct{})
	for _, group := range superPods {
		for _, sn := range group {
			if sn.Name == "" {
				continue
			}
			if _, isFault := faultSet[sn.Name]; !isFault {
				pinned[sn.Name] = struct{}{}
			}
		}
	}
	klog.V(util.LogDebugLev).Infof("getHistoricalHealthyNodeNames: missing=%d, pinned=%d",
		len(missingNodes), len(pinned))
	return pinned
}

// expandFaultLevel walks up from each fault node's largest complete fault subtree
// and adds all siblings at the parent level to the fault set. Returns the expanded
// node list, the maximum len(largestFaultTree.Levels) across all fault nodes
// (1=pod level, >1=group levels), and whether any expansion occurred.
func expandFaultLevel(taskTree *util.TaskTree, currentFaultNodes []string) ([]string, int, bool) {
	faultSet := make(map[string]struct{}, len(currentFaultNodes))
	for _, n := range currentFaultNodes {
		faultSet[n] = struct{}{}
	}
	expanded := make(map[string]struct{}, len(currentFaultNodes)*2)
	for _, n := range currentFaultNodes {
		expanded[n] = struct{}{}
	}

	maxLevel := 0
	didExpand := false
	for _, faultNode := range currentFaultNodes {
		largestFaultTree, err := findLargestFaultSubTree(taskTree, faultNode, faultSet)
		if err != nil || largestFaultTree == nil || largestFaultTree.TaskNode == nil {
			continue
		}
		if len(largestFaultTree.Levels) > maxLevel {
			maxLevel = len(largestFaultTree.Levels)
		}
		parentOfLargest := largestFaultTree.Parent
		if parentOfLargest == nil {
			continue
		}
		for _, sibling := range parentOfLargest.Children {
			expanded[sibling.ResourceNodeName] = struct{}{}
			didExpand = true
		}
	}

	result := make([]string, 0, len(expanded))
	for n := range expanded {
		result = append(result, n)
	}
	klog.V(util.LogDebugLev).Infof("expandFaultLevel: faultNodes %d→%d, maxLevel=%d, didExpand=%v",
		len(currentFaultNodes), len(result), maxLevel, didExpand)
	return result, maxLevel, didExpand
}

// filterOutAvailableNodes filters nodeNames to those NOT in the candidate node list.
func filterOutAvailableNodes(nodeNames []string, nodes []*api.NodeInfo) []string {
	nodeSet := make(map[string]struct{}, len(nodes))
	for _, n := range nodes {
		nodeSet[n.Name] = struct{}{}
	}
	var result []string
	for _, name := range nodeNames {
		if _, ok := nodeSet[name]; !ok {
			result = append(result, name)
		}
	}
	klog.V(util.LogDebugLev).Infof("filterOutAvailableNodes: input=%d, filtered=%d", len(nodeNames), len(result))
	return result
}

// isFaultRankZero returns true if the fault task is rank 0. Rank-0 faults escalate
// immediately because the leader pod cannot be replaced within a partial group.
func isFaultRankZero(task *api.TaskInfo, jobs map[api.JobID]plugin.SchedulerJob) bool {
	job, ok := jobs[task.Job]
	if !ok {
		return false
	}
	rank, err := getHcclRankIndex(task, job)
	if err != nil {
		return false
	}
	return rank == 0
}

// getMissingNodesFromJob collects historical node names from superPods for all
// pending tasks in the job. This gives the complete set of nodes that may need
// to be reconsidered when pods are rescheduled en masse (e.g. preemption/eviction).
func (mh *MultilevelHandler) getMissingNodesFromJob(superPods map[string][]plugin.SuperNode,
	tasks map[api.TaskID]util.NPUTask) []string {
	var missingNodes []string
	for _, t := range tasks {
		if t.PodStatus != v1.PodPending {
			continue
		}
		rankIndexStr, ok := t.Annotation[plugin.PodRankIndexKey]
		if !ok {
			continue
		}
		rank, err := strconv.Atoi(rankIndexStr)
		if err != nil {
			continue
		}
		logicL1Rank, localRank, err := getL1Ranks(superPods, rank)
		if err != nil {
			continue
		}
		group, ok := superPods[logicL1Rank]
		if !ok || localRank >= len(group) {
			continue
		}
		nodeName := group[localRank].Name
		if nodeName != "" {
			missingNodes = append(missingNodes, nodeName)
		}
	}
	return missingNodes
}

// removeAll removes all entries in toRemove from src and returns a new map.
func removeAll(src map[string]struct{}, toRemove []string) map[string]struct{} {
	result := make(map[string]struct{}, len(src))
	for k, v := range src {
		result[k] = v
	}
	for _, name := range toRemove {
		delete(result, name)
	}
	return result
}

func getFaultJob(jobID api.JobID) (*rescheduling.FaultJob, bool) {
	rescheduleCache := rescheduling.GetReSchedulerCache()
	if rescheduleCache == nil {
		return nil, false
	}
	fJob, fJobExist := rescheduleCache.FaultJobs[jobID]
	if !fJobExist || fJob == nil {
		return nil, false
	}
	return fJob, true
}

func (mh *MultilevelHandler) getFaultNodes(jobID api.JobID) ([]string, error) {
	var faultNodes []string
	faultTasksNodesInfo, ok := mh.SuperPodInfo.SuperPodMapFaultTaskNodes[jobID]
	if !ok {
		return nil, fmt.Errorf("failed jobID [%v] not exist", jobID)
	}
	for _, NodeName := range faultTasksNodesInfo {
		faultNodes = append(faultNodes, NodeName)
	}
	return faultNodes, nil
}

// isCachedSuperPodsValid checks that all cached SuperPod nodes exist in the
// current candidate node set. Nodes where pods are still running are excluded
// because they are naturally occupied and absent from the candidate set.
func (mh *MultilevelHandler) isCachedSuperPodsValid(job *plugin.SchedulerJob, nodes []*api.NodeInfo) bool {
	nodeSet := make(map[string]struct{}, len(nodes))
	for _, n := range nodes {
		nodeSet[n.Name] = struct{}{}
	}
	runningNodes := make(map[string]struct{})
	for _, task := range job.Tasks {
		if task.NodeName != "" {
			runningNodes[task.NodeName] = struct{}{}
		}
	}
	for _, sp := range job.SuperPods {
		for _, sn := range sp {
			if _, running := runningNodes[sn.Name]; running {
				continue
			}
			if _, ok := nodeSet[sn.Name]; !ok {
				return false
			}
		}
	}
	return true
}

func (mh *MultilevelHandler) selectNodeFromCache(job *plugin.SchedulerJob, task *api.TaskInfo, sMap map[string]float64) {
	if *job.JobReadyTag {
		if podGroupEnable, exist := job.Label[plugin.PodGroupScheduleKey]; exist && podGroupEnable == plugin.PodGroupScheduleValue {
			mh.scoreNodeBatchForReadyJob(task, job, sMap)
			return
		}
		mh.scoreNodeForReadyJob(task, *job, sMap)
	}
}

func (mh *MultilevelHandler) scoreNodeBatchForReadyJob(task *api.TaskInfo, job *plugin.SchedulerJob,
	sMap map[string]float64) {
	if task == nil || job == nil || len(sMap) == 0 {
		klog.V(util.LogErrorLev).Infof("scoreNodeBatchForReadyJob %s", errors.New(util.ArgumentError))
		return
	}

	if _, isBackup := task.Pod.Annotations[consts.BackupSourcePodNameKey]; isBackup {
		mh.scoreNodeForHotSwitchBackupPod(sMap)
		return
	}

	rankIdMap := mh.obtainBatchScoreRank(task, job)
	if len(rankIdMap) == 0 {
		klog.V(util.LogErrorLev).Infof("%s scoreNodeBatchForReadyJob %s: rankIdMap empty", mh.GetPluginName(), task.Name)
		*job.JobReadyTag = false
		return
	}
	for rankId := range rankIdMap {
		nodeDepth := len(mh.taskLevels) - 1
		level1Depth := nodeDepth - util.Level1Number
		logicL1Rank := rankId / mh.taskLevels[level1Depth].ReqNode
		localRank := rankId % mh.taskLevels[level1Depth].ReqNode
		klog.V(util.LogInfoLev).Infof("logicL1Rank: %d, localRank: %d", logicL1Rank, localRank)
		logicL1RankIndex := strconv.Itoa(logicL1Rank)
		if localRank >= len(job.SuperPods[logicL1RankIndex]) {
			klog.V(util.LogErrorLev).Infof("logicL1Rank: %d, localRank: %d out of rank", logicL1Rank, localRank)
			*job.JobReadyTag = false
			break
		}
		spn := job.SuperPods[logicL1RankIndex][localRank]
		if _, ok := sMap[spn.Name]; !ok {
			klog.V(util.LogErrorLev).Infof("%s scoreNodeBatchForReadyJob %s: node<%s> not in sMap, select fail",
				mh.GetPluginName(), task.Name, spn.Name)
			*job.JobReadyTag = false
			break
		}
		klog.V(util.LogInfoLev).Infof("%s scoreNodeBatchForReadyJob %s: node<%s logicL1 rank index %s> is exist, select success",
			mh.GetPluginName(), task.Name, spn.Name, logicL1RankIndex)
		sMap[spn.Name] = float64(scoreForNode - rankId)
	}
}

func (mh *MultilevelHandler) obtainBatchScoreRank(taskInfo *api.TaskInfo, job *plugin.SchedulerJob) map[int]struct{} {
	if taskInfo == nil || job == nil {
		klog.V(util.LogErrorLev).Infof("obtainBatchScoreRank %s", errors.New(util.ArgumentError))
		return nil
	}
	spec, ok := taskInfo.Pod.Annotations[util.TaskSpecAnno]
	if !ok {
		klog.V(util.LogErrorLev).Infof("obtainBatchScoreRank %s: (%s/%s) obtain annotation %s failed, skip",
			mh.GetPluginName(), taskInfo.Namespace, taskInfo.Name, util.TaskSpecAnno)
		return nil
	}
	klog.V(util.LogDebugLev).Infof("obtainOriginalRankIdMap job (%s/%s), len(job.Tasks) %d",
		job.NameSpace, job.Name, len(job.Tasks))
	m := make(map[int]struct{}, len(job.Tasks))
	for _, npuTask := range job.Tasks {
		if !npuTask.IsNPUTask() || npuTask.Annotation[util.TaskSpecAnno] != spec {
			continue
		}
		if npuTask.PodStatus != v1.PodPending {
			continue
		}
		rankIndex, ok := npuTask.Annotation[plugin.PodRankIndexKey]
		if !ok {
			klog.V(util.LogWarningLev).Infof("obtainBatchScoreRank (%s/%s): rankIndex not exist",
				npuTask.NameSpace, npuTask.Name)
			continue
		}
		rank, err := strconv.Atoi(rankIndex)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("obtainBatchScoreRank (%s/%s): rankIndex is not int",
				npuTask.NameSpace, npuTask.Name)
			continue
		}
		m[rank] = struct{}{}
	}
	klog.V(util.LogInfoLev).Infof("obtainBatchScoreRank job (%s/%s), len(rankMap) %d",
		job.NameSpace, job.Name, len(m))
	return m
}

func (mh *MultilevelHandler) scoreNodeForReadyJob(task *api.TaskInfo, job plugin.SchedulerJob,
	sMap map[string]float64) {
	if sMap == nil {
		klog.V(util.LogWarningLev).Infof("%s scoreNodeForReadyJob %s: sMap is nil.", mh.GetPluginName(), task.Name)
		return
	}

	if _, isBackup := task.Pod.Annotations[consts.BackupSourcePodNameKey]; isBackup {
		mh.scoreNodeForHotSwitchBackupPod(sMap)
		return
	}

	rank, err := getHcclRankIndex(task, job)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getHcclRankIndex %s failed: %v", task.Name, err)
		return
	}
	logicL1Rank, localRank, err := getL1Ranks(job.SuperPods, rank)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getL1Ranks %s failed: %v", task.Name, err)
		return
	}
	spn := job.SuperPods[logicL1Rank][localRank]
	if _, ok := sMap[spn.Name]; ok {
		klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodes %s: node<%s/%s> is exist, select success",
			mh.GetPluginName(), task.Name, spn.Name, logicL1Rank)
		sMap[spn.Name] += scoreForNode
	}
}

func getHcclRankIndex(task *api.TaskInfo, job plugin.SchedulerJob) (int, error) {
	var rank int
	var err error
	rankIndex, ok := task.Pod.Annotations[plugin.PodRankIndexKey]
	if ok {
		rank, err = strconv.Atoi(rankIndex)
		if err != nil {
			return 0, errors.New("rankIndex is not int")
		}
	} else {
		klog.V(util.LogWarningLev).Infof("getHcclRankIndex %s, rankIndex not exist, use task index", task.Name)
		nTask, ok := job.Tasks[task.UID]
		if !ok {
			return 0, errors.New("task not exist")
		}
		rank = nTask.Index
	}
	return rank, nil
}

func getL1Ranks(logicL1Nodes map[string][]plugin.SuperNode, rank int) (string, int, error) {
	// 1. Collect and sort all L1 ranks
	sortedRanks := make([]int, 0, len(logicL1Nodes))
	for key := range logicL1Nodes {
		rankVal, err := strconv.Atoi(key)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("Invalid L1 rank key: %s", key)
			continue
		}
		sortedRanks = append(sortedRanks, rankVal)
	}
	sort.Ints(sortedRanks)

	// 2. Calculate cumulative node count and find matching L1
	cumulativeNodes := 0
	for _, L1Rank := range sortedRanks {
		spKey := strconv.Itoa(L1Rank)
		nodeCount := len(logicL1Nodes[spKey])

		// 3. Check if rank falls within current SuperPod range
		if rank < cumulativeNodes+nodeCount {
			localRank := rank - cumulativeNodes
			return spKey, localRank, nil
		}
		cumulativeNodes += nodeCount
	}

	// 4. No matching L1 rank found
	return "", 0, fmt.Errorf("rank %d exceeds total L1 rank nodes (%d)", rank, cumulativeNodes)
}

type hotSwitchContext struct {
	logicL1Rank  string
	localRank    int
	TopoTreeName string
	LabelKey     string
	LabelValue   string
}

// resolveHotSwitchSuperPods returns SuperPods for hot switch, falling back to SuperPodReschdInfo cache
// when job.SuperPods is empty.
func (mh *MultilevelHandler) resolveHotSwitchSuperPods(job plugin.SchedulerJob) map[string][]plugin.SuperNode {
	if len(job.SuperPods) > 0 {
		return job.SuperPods
	}
	if mh.SuperPodInfo == nil {
		return nil
	}
	cached, ok := mh.SuperPodInfo.SuperPodReschdInfo[job.Name]
	if !ok {
		return nil
	}
	return cached
}

// buildHotSwitchContext resolves rank, L1 group, and L1 label context for the given task.
// Used by checkNodeForHotSwitch for node validation during the predicate phase.
func (mh *MultilevelHandler) buildHotSwitchContext(task *api.TaskInfo, job plugin.SchedulerJob,
	superPods map[string][]plugin.SuperNode) (*hotSwitchContext, error) {
	rank, err := getHcclRankIndex(task, job)
	if err != nil {
		return nil, err
	}
	logicL1Rank, localRank, err := getL1Ranks(superPods, rank)
	if err != nil {
		return nil, err
	}
	superNodes, ok := superPods[logicL1Rank]
	if !ok {
		return nil, fmt.Errorf("level1 group %s not found in SuperPods", logicL1Rank)
	}
	if localRank >= len(superNodes) {
		return nil, fmt.Errorf("localRank %d out of range for level1 group %s", localRank, logicL1Rank)
	}
	var level1TopoTree string
	for _, sn := range superNodes {
		if sn.TopoTreeName != "" {
			level1TopoTree = sn.TopoTreeName
			break
		}
	}
	cachedNode := superNodes[localRank]
	level1LabelKey, level1LabelValue, err := mh.getL1LabelFromCache(cachedNode)
	if err != nil {
		return nil, err
	}
	return &hotSwitchContext{
		logicL1Rank:  logicL1Rank,
		localRank:    localRank,
		TopoTreeName: level1TopoTree,
		LabelKey:     level1LabelKey,
		LabelValue:   level1LabelValue,
	}, nil
}

// scoreNodeForHotSwitchBackupPod scores nodes for a backup pod in hot switch scenario.
// All nodes in sMap have already passed TopoTree and L1 label validation in checkNodeForHotSwitch,
// so this function simply adds a uniform score bonus to each candidate node.
func (mh *MultilevelHandler) scoreNodeForHotSwitchBackupPod(sMap map[string]float64) {
	if sMap == nil {
		return
	}
	for nodeName := range sMap {
		sMap[nodeName] += float64(scoreForNode)
	}
}

// checkNodeForHotSwitch validates node constraint for backup pod during hot switch in multi-level scheduling.
// For non-hot-switch pods, returns nil without any validation.
func (mh *MultilevelHandler) checkNodeForHotSwitch(task *api.TaskInfo, node plugin.NPUNode) error {
	if _, ok := task.Pod.Annotations[consts.BackupSourcePodNameKey]; !ok {
		return nil
	}
	job, ok := mh.ScheduleEnv.Jobs[task.Job]
	if !ok {
		return fmt.Errorf("%s checkNodeForHotSwitch %s: job is not exist", mh.GetPluginName(), task.Name)
	}
	superPods := mh.resolveHotSwitchSuperPods(job)
	if job.JobReadyTag == nil || !*job.JobReadyTag || len(superPods) == 0 {
		klog.V(util.LogInfoLev).Infof("%s checkNodeForHotSwitch: job not ready or SuperPods empty, "+
			"skip validation for pod %s", mh.GetPluginName(), task.Name)
		return nil
	}
	ctx, err := mh.buildHotSwitchContext(task, job, superPods)
	if err != nil {
		return fmt.Errorf("%s checkNodeForHotSwitch %s: %v", mh.GetPluginName(), task.Name, err)
	}
	nodeTopoTree := node.Label[util.TopoTreeLabel]
	if nodeTopoTree == "" {
		nodeTopoTree = util.DefaultTopoTree
	}
	if ctx.TopoTreeName != nodeTopoTree {
		return fmt.Errorf("%s checkNodeForHotSwitch: node %s topoTree %s mismatch L1 topoTree %s",
			mh.GetPluginName(), node.Name, nodeTopoTree, ctx.TopoTreeName)
	}
	nodeL1Value, l1Exist := node.Label[ctx.LabelKey]
	if !l1Exist || nodeL1Value != ctx.LabelValue {
		return fmt.Errorf("%s checkNodeForHotSwitch: node %s L1 label %s=%s mismatch fault L1 %s=%s",
			mh.GetPluginName(), node.Name, ctx.LabelKey, nodeL1Value, ctx.LabelKey, ctx.LabelValue)
	}
	klog.V(util.LogInfoLev).Infof("%s checkNodeForHotSwitch: backup pod %s passed validation, "+
		"logicL1Rank=%s, node=%s", mh.GetPluginName(), task.Name, ctx.logicL1Rank, node.Name)
	return nil
}

// getL1LabelFromCache returns the L1 label key and its value by looking up the fault node in the cluster cache.
// The cachedNode.TopoTreeName is guaranteed non-empty (at least DefaultTopoTree) because it was set
// during reBuildMultiLevelSchedulingCache from the node's actual label.
func (mh *MultilevelHandler) getL1LabelFromCache(cachedNode plugin.SuperNode) (string, string, error) {
	resourceLevels, ok := mh.FrameAttr.ResourceLevelsInfo[cachedNode.TopoTreeName]
	if !ok {
		return "", "", fmt.Errorf("topoTree %s not found in ResourceLevelsInfo", cachedNode.TopoTreeName)
	}
	l1Index := len(resourceLevels) - 2
	if l1Index < 1 || resourceLevels[l1Index].Type != util.LevelTypeMiddle {
		return "", "", fmt.Errorf("no valid L1 label config for topoTree %s", cachedNode.TopoTreeName)
	}
	l1LabelKey := resourceLevels[l1Index].Label
	faultNode, ok := mh.Nodes[cachedNode.Name]
	if !ok {
		return "", "", fmt.Errorf("cached node %s not found in cluster nodes", cachedNode.Name)
	}
	l1LabelValue, ok := faultNode.Label[l1LabelKey]
	if !ok {
		return "", "", fmt.Errorf("label of L1:%s not found on node %s", l1LabelKey, cachedNode.Name)
	}
	return l1LabelKey, l1LabelValue, nil
}
