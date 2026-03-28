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
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

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
	taskLevels, err := util.GetTaskTreeLevels(mh.AffinityBlocks, mh.NPUTaskNum)
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
	if *job.JobReadyTag && len(job.SuperPods) != 0 {
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes %s: job is ready, skip", mh.GetPluginName(),
			task.Name)
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
	updateSuperNodesForPodLevelRescheduling(selectedNodes, task, job)
	klog.V(util.LogInfoLev).Infof("select nodes by multilevel policy successfully")
	// caching logic level1 group information to superpods structure
	*job.JobReadyTag = true
	job.SuperPods = selectedNodes
	return nil
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
	resourceTrees, getErr := plugin.GetResourceTrees(plugin.GetHealthyNPUNodes(mh.Nodes, nodes),
		mh.FrameAttr.ResourceLevelsInfo, mh.taskLevels)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("[%s] GetResourceTrees failed for task %s: %v", mh.GetPluginName(), task.Name, getErr)
		return nil, fmt.Errorf("[%s] GetResourceTrees failed: %v", mh.GetPluginName(), getErr)
	}

	var selectedTaskTree *util.TaskTree
	for _, resourceTree := range resourceTrees {
		newTaskTree, err := mh.tryScheduleTaskInSingleTree(task, resourceTree)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("[%s] failed to schedule task %s in topotree %s, %v",
				mh.GetPluginName(), task.Name, resourceTree.Name, err)
			continue
		}
		klog.V(util.LogInfoLev).Infof("[%s] successfully scheduled task %s in topotree %s with fragment score %d",
			mh.GetPluginName(), task.Name, resourceTree.Name, newTaskTree.FragmentScore)
		if selectedTaskTree != nil && selectedTaskTree.FragmentScore <= newTaskTree.FragmentScore {
			klog.V(util.LogInfoLev).Infof("[%s] existing fragment score %d, new score %d, choosing last task tree",
				mh.GetPluginName(), selectedTaskTree.FragmentScore, newTaskTree.FragmentScore)
			continue
		}
		selectedTaskTree = newTaskTree
	}
	if selectedTaskTree == nil {
		klog.V(util.LogErrorLev).Infof("[%s] no valid task tree found for task %s", mh.GetPluginName(), task.Name)
		return nil, fmt.Errorf("[%s] no valid task tree found for task %s", mh.GetPluginName(), task.Name)
	}

	supernode, err := plugin.GetSuperNodeMapFromTaskTree(selectedTaskTree)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("[%s] GetSuperNodeMapFromTaskTree failed: %v", mh.GetPluginName(), err)
		return nil, fmt.Errorf("[%s] GetSuperNodeMapFromTaskTree failed: %v", mh.GetPluginName(), err)
	}

	// reset fault node cache scheduling succeed
	mh.SuperPodInfo.SuperPodMapFaultTaskNodes[task.Job] = map[string]string{}
	klog.V(util.LogInfoLev).Infof("[%s] successfully get supernode map for task %s", mh.GetPluginName(), task.Name)
	return supernode, nil
}

func (mh *MultilevelHandler) tryScheduleTaskInSingleTree(task *api.TaskInfo,
	resourceTree *util.ResourceTree) (*util.TaskTree, error) {
	klog.V(util.LogInfoLev).Infof("[%s] trying to schedule task %s on topotree %s", mh.GetPluginName(), task.Name, resourceTree.Name)
	var (
		fJob     *rescheduling.FaultJob
		taskTree *util.TaskTree
		err      error
	)
	fJob, exist := getFaultJob(task.Job)
	if exist && fJob.IsFaultJob {
		taskTree, err = mh.reschedule(fJob, task, resourceTree)
	} else {
		taskTree, err = Schedule(resourceTree, mh.taskLevels)
	}
	if err != nil {
		klog.V(util.LogErrorLev).Infof("[%s] schedule task %s in topotree %s failed: %v",
			mh.GetPluginName(), task.Name, resourceTree.Name, err)
		return nil, err
	}
	return taskTree, nil
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

func (mh *MultilevelHandler) reschedule(fJob *rescheduling.FaultJob, task *api.TaskInfo,
	resourceTree *util.ResourceTree) (*util.TaskTree, error) {
	klog.V(util.LogInfoLev).Infof("[%s] rescheduling job %s on topotree %s", mh.GetPluginName(), task.Job, resourceTree.Name)
	if _, ok := mh.SuperPodInfo.SuperPodReschdInfo[fJob.JobUID]; ok {
		fJob.SuperPods = mh.SuperPodInfo.SuperPodReschdInfo[fJob.JobUID]
		klog.V(util.LogInfoLev).Infof("[%s] loaded rescheduling cache for job %s, logic level1 group: %v",
			mh.GetPluginName(), task.Job, fJob.SuperPods)
	}
	faultNodes, getFaultNodeErr := mh.getFaultNodes(fJob.JobUID)
	if getFaultNodeErr != nil {
		klog.V(util.LogErrorLev).Infof("[%s] loaded fault tasks node for job %s failed, %v",
			mh.GetPluginName(), task.Job, getFaultNodeErr)
		return nil, getFaultNodeErr
	}
	klog.V(util.LogDebugLev).Infof("job: %v/%v, faultNodes: %v", fJob.JobNamespace, fJob.JobName, faultNodes)

	taskTree, err := plugin.GetTaskTreeFromSuperNodeMap(fJob.SuperPods, mh.taskLevels,
		resourceTree.Levels, mh.Nodes)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("[%s] GetTaskTreeFromSuperNodeMap for job %s failed, %v",
			mh.GetPluginName(), task.Job, err)
		return nil, err
	}
	taskTree, err = Reschedule(resourceTree, taskTree, faultNodes)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("[%s] Reschedule for job %s failed, %v",
			mh.GetPluginName(), task.Job, err)
		return nil, err
	}
	return taskTree, nil
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

func updateSuperNodesForPodLevelRescheduling(currentSelectedNodes map[string][]plugin.SuperNode,
	task *api.TaskInfo, job plugin.SchedulerJob) {
	if currentSelectedNodes == nil {
		klog.V(util.LogWarningLev).Infof("updateSuperNodesForPodLevelRescheduling: currentSelectedNodes is nil.")
		return
	}
	fJob, exist := getFaultJob(task.Job)
	if !exist || !ifPodLevelRescheduling(fJob, &job) {
		return
	}
	klog.V(util.LogInfoLev).Infof("select nodes for pod level rescheduling")
	// get logical L1 group rank of fault node
	rank, err := getHcclRankIndex(task, job)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("getHcclRankIndex %s failed: %v", task.Name, err)
		return
	}
	// get cached L1 group and records the nodes that have been used in the past
	logicL1Rank, localRank, err := getL1Ranks(fJob.SuperPods, rank)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("logic L1 ranks %s from cache[%v] failed: %v",
			task.Name, currentSelectedNodes, err)
		return
	}
	superNodesCache := fJob.SuperPods[logicL1Rank]
	klog.V(util.LogDebugLev).Infof("cached last scheduling superNodes: %v", superNodesCache)
	lastSelectedNodes := make(map[string]plugin.SuperNode, len(superNodesCache))
	for index, node := range superNodesCache {
		if index == localRank {
			continue
		}
		lastSelectedNodes[node.Name] = node
	}
	for _, node := range currentSelectedNodes[logicL1Rank] {
		_, isHistorySelected := lastSelectedNodes[node.Name]
		if isHistorySelected {
			continue
		}
		superNodesCache[localRank] = node
	}
	currentSelectedNodes[logicL1Rank] = superNodesCache
	klog.V(util.LogDebugLev).Infof("updated selectedNodes[%v]: %v", logicL1Rank, currentSelectedNodes[logicL1Rank])
}

func ifPodLevelRescheduling(fJob *rescheduling.FaultJob, sJob *plugin.SchedulerJob) bool {
	// for multilevel rescheduling, only single pod need to individually update fault node in L1 group
	return sJob.SchedulingTaskNum != len(sJob.Tasks) && fJob.PendingSessionNum < sessionsForSinglePod &&
		(fJob.IsJobSingleRescheduling(sJob) || fJob.IsProcessReschedulingJob(sJob))
}
