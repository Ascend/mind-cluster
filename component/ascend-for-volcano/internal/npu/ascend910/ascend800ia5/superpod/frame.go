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
Package superpod is using for HuaWei ascend 800I A5 SuperPod affinity schedule.
*/
package superpod

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

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
	m := &module800SuperPod{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetMaxNodeNPUNum(nodeNPUNumber)
	m.netUnhealthyKey = networkUnhealthyNPU
	m.nodeVPodId = map[string]string{}
	return m
}

// ValidNPUJob verify the validity of job parameters
func (tp *module800SuperPod) ValidNPUJob() *api.ValidateResult {
	if res := tp.checkSpBlock(); res != nil {
		return res
	}
	return tp.checkRequireNPU()
}

func (tp *module800SuperPod) checkSpBlock() *api.ValidateResult {
	if tp.SpBlockNPUNum <= 0 {
		return &api.ValidateResult{
			Pass:    false,
			Reason:  spBlockInvalidReason,
			Message: fmt.Sprintf("sp-block(%d) is invalid", tp.SpBlockNPUNum),
		}
	}
	if tp.SpBlockNPUNum < nodeNPUNumber {
		klog.V(util.LogWarningLev).Info("sp-block less than 8, set default value 1")
		tp.spBlock = 1
	} else {
		if tp.SpBlockNPUNum%nodeNPUNumber != 0 {
			return &api.ValidateResult{
				Pass:    false,
				Reason:  spBlockInvalidReason,
				Message: fmt.Sprintf("sp-block(%d) is not mutiple of node npu (%d)", tp.SpBlockNPUNum, nodeNPUNumber),
			}
		}
		tp.spBlock = tp.SpBlockNPUNum / nodeNPUNumber
	}

	if tp.spBlock > tp.FrameAttr.SuperPodSize {
		return &api.ValidateResult{
			Pass:   false,
			Reason: spBlockInvalidReason,
			Message: fmt.Sprintf("sp-block(%d/8=%d) is bigger than size of super-pod(%d)",
				tp.SpBlockNPUNum, tp.spBlock, tp.FrameAttr.SuperPodSize),
		}
	}
	return nil
}

func (tp *module800SuperPod) checkRequireNPU() *api.ValidateResult {
	if tp.NPUTaskNum == 1 {
		if tp.ReqNPUNum == 1 || tp.ReqNPUNum <= nodeNPUNumber {
			if tp.ReqNPUNum != tp.SpBlockNPUNum {
				return &api.ValidateResult{
					Pass:    false,
					Reason:  jobCheckFailedReason,
					Message: "single super-pod job sp-block annotation should equal require npu num",
				}
			}
			return nil
		}
		return &api.ValidateResult{
			Pass:    false,
			Reason:  jobCheckFailedReason,
			Message: fmt.Sprintf("single super-pod job require npu [1, 2*n], instead of %d", tp.ReqNPUNum),
		}
	}

	// distributed job required npu must be multiple of sp-block
	if tp.ReqNPUNum%tp.SpBlockNPUNum != 0 {
		return &api.ValidateResult{
			Pass:   false,
			Reason: jobCheckFailedReason,
			Message: fmt.Sprintf("distributed super-pod job require npu(%d) should be multiple of sp-block",
				tp.ReqNPUNum),
		}
	}
	return tp.checkReqNPUEqualNodeNPU()
}

func (tp *module800SuperPod) checkReqNPUEqualNodeNPU() *api.ValidateResult {
	for _, task := range tp.Tasks {
		// npu num required by task in distributed job must be node npu num
		if task.ReqNPUNum != nodeNPUNumber {
			return &api.ValidateResult{
				Pass:    false,
				Reason:  jobCheckFailedReason,
				Message: fmt.Sprintf("distributed super-pod job require npu 8*n, instead of %d", task.ReqNPUNum),
			}
		}
	}
	return nil
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *module800SuperPod) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err)
		return err
	}
	// node in super-pod has super-podID which is not less than 0
	if node.SuperPodID < 0 {
		return fmt.Errorf("node %s is not super-pod node or superPodID is not set", node.Name)
	}

	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	nodeTop, err := tp.GetUsableTopFromNode(node, tp.NPUTaskNum/tp.spBlock > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof(getNPUFromPodFailedPattern, tp.GetPluginName(), err.Error())
		return err
	}

	if err = tp.NPUHandler.JudgeNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogErrorLev).Infof("%s JudgeNodeAndTaskNPU err: %s", tp.GetPluginName(), err.Error())
		return fmt.Errorf("checkNodeNPUByTask %s err: %s", util.NodeNotMeetTopologyWarning, err.Error())
	}
	return nil
}

// ScoreBestNPUNodes get best nodes score for job
func (tp *module800SuperPod) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo,
	sMap map[string]float64) error {
	if tp == nil || task == nil || len(nodes) == 0 || len(sMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %s.", err)
		return err
	}

	job, ok := tp.ScheduleEnv.Jobs[task.Job]
	if !ok {
		return fmt.Errorf("%s ScoreBestNPUNodes %s: job is not exist", tp.GetPluginName(), task.Name)
	}

	defer func() {
		tp.ScheduleEnv.Jobs[task.Job] = job
	}()

	if !*job.JobReadyTag {
		return nil
	}

	defer func() {
		if *job.JobReadyTag {
			tp.scoreNodeForReadyJob(task, job, sMap)
		}
	}()

	if *job.JobReadyTag && len(job.SuperPods) != 0 {
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes %s: job is ready, skip", tp.GetPluginName(),
			task.Name)
		return nil
	}

	if tp.NPUTaskNum == 1 {
		nodes = tp.selectNodesWithLeastResourceForSingle(nodes)
	}

	if tp.NPUTaskNum > len(nodes) && tp.SchedulingTaskNum == len(tp.Tasks) {
		*job.JobReadyTag = false
		return fmt.Errorf("select node failed by not enough node")
	}

	selectedNodes, err := tp.selectSuperPodForJob(task, nodes, sMap)
	if err != nil {
		*job.JobReadyTag = false
		return err
	}
	*job.JobReadyTag = true
	job.SuperPods = selectedNodes
	for id, sp := range selectedNodes {
		for _, node := range sp {
			tp.nodeVPodId[node.Name] = id
		}
	}
	klog.V(util.LogDebugLev).Infof("update SuperPods with selectedNodes: %#v", selectedNodes)
	return nil
}

func (tp *module800SuperPod) scoreNodeForReadyJob(task *api.TaskInfo, job plugin.SchedulerJob,
	sMap map[string]float64) {
	if len(sMap) == 0 {
		klog.V(util.LogErrorLev).Infof("scoreNodeForReadyJob: sMap is nil")
		return
	}
	var rank int
	var err error
	rankIndex, ok := task.Pod.Annotations[plugin.PodRankIndexKey]
	if ok {
		rank, err = strconv.Atoi(rankIndex)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s %s ScoreBestNPUNodes %s: rankIndex is not int",
				tp.GetPluginName(), task.Name, task.Name)
			return
		}
	} else {
		klog.V(util.LogWarningLev).Infof("%s %s ScoreBestNPUNodes %s: rankIndex is not exist",
			tp.GetPluginName(), task.Name, task.Name)
		nTask, ok := job.Tasks[task.UID]
		if !ok {
			klog.V(util.LogErrorLev).Infof("%s scoreNodeForReadyJob %s: task is not exist", tp.GetPluginName(),
				task.Name)
			return
		}
		rank = nTask.Index
	}
	if !tp.checkSpBlockGtZero() {
		return
	}
	superPodRankIndex, localRank := getSuperPodRanks(job, rank)
	if superPodRankIndex == "" {
		klog.V(util.LogErrorLev).Infof("CalculateRankIndex failed: %v", err)
		return
	}
	klog.V(util.LogInfoLev).Infof("superPodRank: %v, localRank: %d", superPodRankIndex, localRank)
	if localRank >= len(job.SuperPods[superPodRankIndex]) {
		klog.V(util.LogErrorLev).Infof("superPodRank: %v, localRank: %d out of rank", superPodRankIndex, localRank)
		return
	}
	spn := job.SuperPods[superPodRankIndex][localRank]
	if _, ok = sMap[spn.Name]; ok {
		klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes %s: node<%s/%s> is exist in "+
			"SuperPodID: %d, select success", tp.GetPluginName(), task.Name, spn.Name, superPodRankIndex,
			spn.SuperPodID)
		sMap[spn.Name] += scoreForNode
	}
}

func (tp *module800SuperPod) selectNodesWithLeastResourceForSingle(nodes []*api.NodeInfo) []*api.NodeInfo {
	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes : len(tp.Tasks) == 1", tp.GetPluginName())
	resourceName := v1.ResourceName(tp.ReqNPUName)
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Idle.ScalarResources[resourceName] < nodes[j].Idle.ScalarResources[resourceName]
	})
	n0 := nodes[0]
	preFitNodes := make([]*api.NodeInfo, 0)
	for _, node := range nodes {
		if node.Idle.ScalarResources[resourceName] > n0.Idle.ScalarResources[resourceName] {
			break
		}
		preFitNodes = append(preFitNodes, node)
	}

	rescheduleCache := rescheduling.GetReSchedulerCache()
	if rescheduleCache == nil {
		klog.V(util.LogDebugLev).Info("rescheduleCache is nil")
		return preFitNodes
	}

	fitNodes := make([]*api.NodeInfo, 0, len(preFitNodes))
	for _, node := range preFitNodes {
		if fNode, exist := rescheduleCache.FaultNodes[node.Name]; exist &&
			(fNode.HasCardSubHealthFault || fNode.HasSwitchSubHealthFault) {
			klog.V(util.LogDebugLev).Infof("try to filter subHealthy node: %s", fNode.NodeName)
			continue
		}
		fitNodes = append(fitNodes, node)
	}
	if len(fitNodes) > 0 {
		return fitNodes
	}

	return preFitNodes
}

func (tp *module800SuperPod) selectSuperPodForJob(task *api.TaskInfo, nodes []*api.NodeInfo,
	sMap map[string]float64) (map[string][]plugin.SuperNode, error) {
	klog.V(util.LogInfoLev).Infof("input nodes num(%d) for task %s", len(nodes), task.Name)
	totalNodes := tp.getSuperPodTop(nodes)
	if !tp.checkSpBlockGtZero() {
		return nil, fmt.Errorf("select super pod failed, sp-block less than 0")
	}
	tp.isSoftSuperPodAffinity = tp.Label[superPodAffinity] == softRequire
	totalRequiredSuperPod := tp.NPUTaskNum / tp.spBlock
	vSuperPodID := make(map[string]bool, totalRequiredSuperPod)
	for i := 0; i < totalRequiredSuperPod; i++ {
		vSuperPodID[strconv.Itoa(i)] = false
	}
	selectNodes, err := tp.selectNodesForFaultJob(task, totalNodes, vSuperPodID, sMap, nodes)
	if err != nil {
		return nil, err
	}
	spi, err := tp.classifySuperPod(totalNodes)
	if err != nil {
		return nil, err
	}
	var unReadyID []string
	for id, ready := range vSuperPodID {
		if !ready {
			unReadyID = append(unReadyID, id)
		}
	}

	isSuperPodRescheduling := tp.SchedulingTaskNum < len(tp.Tasks) && tp.SchedulingTaskNum > util.NPUIndex1

	if spi.countVSuperPod < len(unReadyID) && (!tp.isSoftSuperPodAffinity ||
		(tp.isSoftSuperPodAffinity && isSuperPodRescheduling)) {
		return nil, fmt.Errorf("select super pod failed, required %d, total %d", len(unReadyID), spi.countVSuperPod)
	}
	tp.selectNodes(unReadyID, &spi, selectNodes, len(vSuperPodID))

	return selectNodes, nil
}

func (tp *module800SuperPod) getSuperPodTop(nodes []*api.NodeInfo) map[int32]superPod {
	totalNodes := make(map[int32]superPod)
	for _, node := range nodes {
		nNode, ok := tp.Nodes[node.Name]
		if !ok {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes %s is not npu node",
				tp.GetPluginName(), node.Name)
			continue
		}
		_, exist := totalNodes[nNode.SuperPodID]
		if !exist {
			totalNodes[nNode.SuperPodID] = superPod{}
		}
		totalNodes[nNode.SuperPodID][node.Name] = nNode
	}
	return totalNodes
}

func (tp *module800SuperPod) selectNodesForFaultJob(task *api.TaskInfo, totalNodes map[int32]superPod,
	vSuperPodID map[string]bool, sMap map[string]float64,
	nodes []*api.NodeInfo) (map[string][]plugin.SuperNode, error) {
	if tp == nil || task == nil {
		return nil, fmt.Errorf("selectNodesForFaultJob task is nil")
	}

	selectNodes := make(map[string][]plugin.SuperNode)
	rescheduleCache := rescheduling.GetReSchedulerCache()
	if rescheduleCache == nil {
		return selectNodes, nil
	}

	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes %s: reScheduler is not nil", tp.GetPluginName(), task.Name)
	fJob := rescheduleCache.FaultJobs[task.Job]
	if fJob == nil || !fJob.IsFaultJob {
		return selectNodes, nil
	}

	if fJob.RescheduleTime == 0 {
		fJob.RescheduleTime = time.Now().Unix()
	}

	if !tp.isDelayingJob(fJob, nodes) && tp.SchedulingTaskNum == len(tp.Tasks) {
		return selectNodes, fmt.Errorf("selectNode failed, wait for normal node resource release")
	}
	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes %s: is fault job, superPods: %v",
		tp.GetPluginName(), fJob.JobName, fJob.SuperPods)

	if !tp.schedulable(fJob, totalNodes) {
		return selectNodes, fmt.Errorf("selectNodeFromOriginVSuperPod failed, unschedulable")
	}

	notReadySuperPod, err := tp.selectNodeFromOriginVSuperPod(fJob, sMap, selectNodes, totalNodes, vSuperPodID)
	if err != nil {
		return selectNodes, nil
	}

	tp.selectNodeFromOriginSuperPod(fJob, notReadySuperPod, totalNodes, vSuperPodID, selectNodes)
	if tp.ifPodLevelRescheduling(fJob) {
		tp.selectNodeForPodLevelRescheduling(fJob, notReadySuperPod, totalNodes, vSuperPodID, selectNodes)
	}

	return selectNodes, nil
}

// isDelayingJob checks if the job should continue waiting for resource release
// Returns true if:
//   - The waiting time exceeds the delayingTime threshold (10s)
//   - All normal nodes used by the job have been released
//
// Returns false if any normal node is still occupied
func (tp *module800SuperPod) isDelayingJob(fJob *rescheduling.FaultJob, nodes []*api.NodeInfo) bool {
	if tp == nil || fJob == nil || fJob.FaultTasks == nil {
		return false
	}
	// Check if waiting time exceeds threshold
	if time.Now().Unix()-fJob.RescheduleTime > delayingTime {
		klog.V(util.LogWarningLev).Infof("job %s wait used resource release time over 10s, skip wait", fJob.JobName)
		return true
	}

	// Convert nodes to map for quick lookup
	nodeMaps := util.ChangeNodesToNodeMaps(nodes)

	// Check all non-fault tasks to see if their nodes are released
	for _, task := range fJob.FaultTasks {
		if task.IsFaultTask {
			continue
		}
		// If node is not in available nodes list, it's still occupied
		if _, ok := nodeMaps[task.NodeName]; !ok {
			klog.V(util.LogWarningLev).Infof("job used %s normal node %s is not release", fJob.JobName, task.NodeName)
			return false
		}
	}
	return true
}

func (tp *module800SuperPod) selectNodeFromOriginVSuperPod(fJob *rescheduling.FaultJob, sMap map[string]float64,
	selectNodes map[string][]plugin.SuperNode, totalNodes map[int32]superPod,
	vSuperPodID map[string]bool) (map[string]struct{}, error) {
	if tp == nil || fJob == nil {
		return nil, fmt.Errorf("selectNodeFromOriginVSuperPod task is nil")
	}

	if selectNodes == nil || vSuperPodID == nil {
		return nil, nil
	}
	if _, ok := tp.SuperPodInfo.SuperPodReschdInfo[fJob.JobUID]; ok {
		fJob.SuperPods = tp.SuperPodInfo.SuperPodReschdInfo[fJob.JobUID]
	}
	if tp.ifPodLevelRescheduling(fJob) {
		return tp.selectForPodRescheduling(fJob, selectNodes, vSuperPodID)
	}
	return tp.selectForJobRescheduling(fJob, sMap, selectNodes, totalNodes, vSuperPodID)
}

func (tp *module800SuperPod) selectForJobRescheduling(fJob *rescheduling.FaultJob, sMap map[string]float64,
	selectNodes map[string][]plugin.SuperNode, totalNodes map[int32]superPod,
	vSuperPodID map[string]bool) (map[string]struct{}, error) {
	if selectNodes == nil || vSuperPodID == nil {
		return nil, nil
	}
	notReadySuperPod := make(map[string]struct{})
	for superPodId, superPod := range fJob.SuperPods {
		count := 0
		for _, spn := range superPod {
			if _, ok := sMap[spn.Name]; ok && judgeLasTimeTaskIsHealthy(fJob, spn.Name) {
				count++
			}
		}
		if count < len(superPod) {
			notReadySuperPod[superPodId] = struct{}{}
			continue
		}
		klog.V(util.LogInfoLev).Infof("superPodId: %s is satisfied superPod: %v", superPodId, superPod)
		for _, spn := range superPod {
			delete(totalNodes[spn.SuperPodID], spn.Name)
		}
		selectNodes[superPodId] = superPod
		vSuperPodID[superPodId] = true
	}
	return notReadySuperPod, nil
}

func (tp *module800SuperPod) selectForPodRescheduling(fJob *rescheduling.FaultJob,
	selectNodes map[string][]plugin.SuperNode, vSuperPodID map[string]bool) (map[string]struct{}, error) {
	if selectNodes == nil || vSuperPodID == nil {
		return nil, nil
	}
	notReadySuperPod := make(map[string]struct{})
	for superPodId, superPod := range fJob.SuperPods {
		count := 0
		for _, spn := range superPod {
			if judgeLasTimeTaskIsHealthy(fJob, spn.Name) {
				count++
			}
		}
		if count < len(superPod) {
			notReadySuperPod[superPodId] = struct{}{}
			continue
		}
		klog.V(util.LogInfoLev).Infof("superPodId: %s is satisfied superPod: %v", superPodId, superPod)
		selectNodes[superPodId] = superPod
		vSuperPodID[superPodId] = true
	}
	return notReadySuperPod, nil
}

func judgeLasTimeTaskIsHealthy(fJob *rescheduling.FaultJob, nodeName string) bool {
	if fJob == nil || fJob.FaultTasks == nil {
		return false
	}

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

func (tp *module800SuperPod) schedulable(fJob *rescheduling.FaultJob, totalNodes map[int32]superPod) bool {
	if tp == nil || fJob == nil {
		return false
	}

	count := make(map[int32]int)
	for _, sp := range fJob.SuperPods {
		faultTasks := 0
		for _, task := range fJob.FaultTasks {
			_, podExists := tp.Tasks[task.TaskUID]
			klog.V(util.LogDebugLev).Infof("task.IsFaultTask: %v, podExists: %v, task: %#v",
				task.IsFaultTask, podExists, task)
			if !task.IsFaultTask && podExists {
				continue
			}
			if ok, _ := tp.isContain(sp, task.TaskName, fJob.JobUID); ok {
				faultTasks++
			}
		}
		klog.V(util.LogDebugLev).Infof("the number of faultTasks is %v, len(sp) is: %v", faultTasks, len(sp))
		if faultTasks == len(sp) {
			klog.V(util.LogDebugLev).Info("schedulable return true")
			return true
		}
		if value, ok := count[sp[0].SuperPodID]; ok {
			count[sp[0].SuperPodID] = value + faultTasks
		} else {
			count[sp[0].SuperPodID] = faultTasks
		}
	}
	return ifSchedule(count, totalNodes)
}

func ifSchedule(count map[int32]int, totalNodes map[int32]map[string]plugin.NPUNode) bool {
	if len(count) == 0 || len(totalNodes) == 0 {
		return false
	}

	for id, res := range count {
		if len(totalNodes[id]) < res {
			return false
		}
	}
	return true
}

func (tp *module800SuperPod) isContain(superPod []plugin.SuperNode, name string, jobId api.JobID) (bool, int) {
	if tp == nil {
		return false, -1
	}
	klog.V(util.LogDebugLev).Infof("tp.SuperPodInfo.SuperPodMapFaultTaskNodes[%v][%s]: %s",
		jobId, name, tp.SuperPodInfo.SuperPodMapFaultTaskNodes[jobId][name])
	klog.V(util.LogDebugLev).Infof("superPod is: %#v", superPod)
	for id, each := range superPod {
		if each.Name == tp.SuperPodInfo.SuperPodMapFaultTaskNodes[jobId][name] {
			klog.V(util.LogDebugLev).Infof("the id of node in superPod is: %d", id)
			return true, id
		}
	}
	return false, -1
}

func (tp *module800SuperPod) selectNodeFromOriginSuperPod(fJob *rescheduling.FaultJob,
	notReadySuperPod map[string]struct{}, totalNodes map[int32]superPod,
	vSuperPodID map[string]bool, selectNodes map[string][]plugin.SuperNode) {
	if tp == nil || fJob == nil || fJob.SuperPods == nil || len(fJob.SuperPods) == 0 {
		return
	}

	if selectNodes == nil || vSuperPodID == nil {
		return
	}
	faultNodeNameMap := getFaultNodeNameMap(fJob)
	for superPodId := range notReadySuperPod {
		spn := fJob.SuperPods[superPodId][0]
		if len(totalNodes[spn.SuperPodID]) < tp.spBlock {
			continue
		}
		klog.V(util.LogInfoLev).Infof("superPodId: %s is satisfied superPod: %v in super-pod: %d",
			superPodId, fJob.SuperPods[superPodId], spn.SuperPodID)
		vSuperPodID[superPodId] = true
		selectNodes[superPodId] = getSelectNodes(faultNodeNameMap,
			fJob.SuperPods[superPodId], totalNodes[spn.SuperPodID])
		for _, node := range selectNodes[superPodId] {
			delete(totalNodes[node.SuperPodID], node.Name)
		}
	}
}

func getFaultNodeNameMap(job *rescheduling.FaultJob) map[string]struct{} {
	faultNodeNameMap := map[string]struct{}{}
	if job == nil || job.FaultTasks == nil {
		return faultNodeNameMap
	}
	for _, faultTask := range job.FaultTasks {
		if faultTask.IsFaultTask {
			faultNodeNameMap[faultTask.NodeName] = struct{}{}
		}
	}
	return faultNodeNameMap
}

func (tp *module800SuperPod) ifPodLevelRescheduling(fJob *rescheduling.FaultJob) bool {
	if tp == nil || fJob == nil {
		return false
	}

	job, ok := tp.Jobs[fJob.JobUID]
	if !ok {
		return false
	}
	klog.V(util.LogInfoLev).Infof("label pod-rescheduling is: %s", job.Label[util.SinglePodTag])
	return job.Label[util.SinglePodTag] == util.EnableFunc
}

func (tp *module800SuperPod) selectNodeForPodLevelRescheduling(fJob *rescheduling.FaultJob,
	notReadySuperPod map[string]struct{}, totalNodes map[int32]map[string]plugin.NPUNode,
	vSuperPodID map[string]bool, selectNodes map[string][]plugin.SuperNode) {
	if fJob == nil || fJob.SuperPods == nil {
		return
	}

	if selectNodes == nil || vSuperPodID == nil {
		return
	}
	for superPodId := range notReadySuperPod {
		if len(fJob.SuperPods[superPodId]) == 0 {
			return
		}
		spn := fJob.SuperPods[superPodId][0]
		ids := tp.getLogicSuperPodFaultTaskIds(fJob, superPodId)
		if len(ids) == 0 {
			klog.V(util.LogInfoLev).Infof("superPodId: %s is satisfied superPod: %v in super-pod: %d and ids: %v",
				superPodId, fJob.SuperPods[superPodId], spn.SuperPodID, ids)
			selectNodes[superPodId] = fJob.SuperPods[superPodId]
			vSuperPodID[superPodId] = true
			continue
		}
		// If the number of failed pods in the logical supernode exceeds the number of available nodes in the physical supernode,
		// or if the number of pods to be scheduled exceeds the size of the logical supernode
		if len(ids) > len(totalNodes[spn.SuperPodID]) || tp.SchedulingTaskNum >= len(fJob.SuperPods[superPodId]) {
			continue
		}
		selectNodesForFaultPod(fJob, ids, totalNodes, spn, superPodId)
		klog.V(util.LogInfoLev).Infof("superPodId: %s is satisfied superPod: %v in super-pod: %d and ids: %v",
			superPodId, fJob.SuperPods[superPodId], spn.SuperPodID, ids)
		selectNodes[superPodId] = fJob.SuperPods[superPodId]
		vSuperPodID[superPodId] = true
	}
}

func selectNodesForFaultPod(fJob *rescheduling.FaultJob, ids []int, totalNodes map[int32]map[string]plugin.NPUNode,
	spn plugin.SuperNode, superPodId string) {
	for _, id := range ids {
		for _, node := range totalNodes[spn.SuperPodID] {
			if inSuperPods(fJob, superPodId, node) {
				delete(totalNodes[spn.SuperPodID], node.Name)
				break
			}
			fJob.SuperPods[superPodId][id].Name = node.Name
			delete(totalNodes[spn.SuperPodID], node.Name)
			break
		}
	}
	return
}

func inSuperPods(fJob *rescheduling.FaultJob, superPodId string, node plugin.NPUNode) bool {
	if fJob == nil || fJob.SuperPods == nil || len(fJob.SuperPods) == 0 {
		return false
	}

	for _, spNode := range fJob.SuperPods[superPodId] {
		if spNode.Name == node.Name {
			return true
		}
	}
	return false
}

func (tp *module800SuperPod) getLogicSuperPodFaultTaskIds(fJob *rescheduling.FaultJob, superPodId string) []int {
	if tp == nil || fJob == nil {
		return nil
	}

	var ids []int
	for _, task := range fJob.FaultTasks {
		if task.IsFaultTask {
			if ok, index := tp.isContain(fJob.SuperPods[superPodId], task.TaskName, fJob.JobUID); ok {
				ids = append(ids, index)
			}
		}
	}
	return ids
}

func getSelectNodes(faultNodeNameMap map[string]struct{}, spNodes []plugin.SuperNode,
	spNodeMaps map[string]plugin.NPUNode) []plugin.SuperNode {
	if spNodeMaps == nil {
		return nil
	}
	reserveIndex := make([]int, 0)
	newNodes := make([]plugin.SuperNode, len(spNodes))
	// 1. use last time healthy node first
	for idx, spNode := range spNodes {
		if _, ok := faultNodeNameMap[spNode.Name]; ok {
			reserveIndex = append(reserveIndex, idx)
			continue
		}
		if _, ok := spNodeMaps[spNode.Name]; ok {
			newNodes[idx] = spNode
			delete(spNodeMaps, spNode.Name)
			continue
		}
		reserveIndex = append(reserveIndex, idx)
	}
	i := 0
	getOtherNodeFunc := func(notUseFault bool) {
		for _, node := range spNodeMaps {
			if i == len(reserveIndex) {
				return
			}
			if _, ok := faultNodeNameMap[node.Name]; ok == notUseFault {
				continue
			}
			newNodes[reserveIndex[i]] = plugin.SuperNode{
				Name:       node.Name,
				SuperPodID: node.SuperPodID,
			}
			i++
		}
	}
	// 2. use last time not used nodes second
	getOtherNodeFunc(true)
	// 3. use last time fault nodes third
	getOtherNodeFunc(false)
	return newNodes
}

func (tp *module800SuperPod) initRemainderTop() [][][]superPod {
	maxMultiple := tp.FrameAttr.SuperPodSize/tp.spBlock + 1
	rmd := make([][][]superPod, tp.spBlock)
	for i := range rmd {
		rmd[i] = make([][]superPod, maxMultiple)
	}
	return rmd
}

func (tp *module800SuperPod) classifySuperPod(totalNodes map[int32]superPod) (superPodInfo, error) {
	firstLevelRemainTop := tp.initRemainderTop()
	countVSuperPod := 0
	column := 0
	remainder := 0
	if !tp.checkSpBlockGtZero() {
		return superPodInfo{}, fmt.Errorf("classify super pod failed, sp-block less than 0")
	}
	for index, sp := range totalNodes {
		klog.V(util.LogInfoLev).Infof("super-pod: %d, len: %d", index, len(sp))
		if len(sp) < tp.spBlock && !tp.isSoftSuperPodAffinity {
			continue
		}
		if tp.FrameAttr.SuperPodSize < len(sp) {
			klog.V(util.LogErrorLev).Infof("please adjust super-pod-size, now super-pod-size(%d) "+
				"less than superPod's node size(%d)", tp.FrameAttr.SuperPodSize, len(sp))
			return superPodInfo{}, fmt.Errorf("super-pod-size is smaller than superPod's node size")
		}
		countVSuperPod += len(sp) / tp.spBlock
		nodesExceptReserve := len(sp) - tp.FrameAttr.ReservePodSize
		if nodesExceptReserve < 0 {
			column = 0
			remainder = len(sp) % tp.spBlock
		} else {
			column = nodesExceptReserve / tp.spBlock
			remainder = nodesExceptReserve % tp.spBlock
		}
		klog.V(util.LogInfoLev).Infof("super-pod: %d, column: %d, remainder: %d", index, column,
			remainder)
		if len(firstLevelRemainTop[remainder][column]) == 0 {
			firstLevelRemainTop[remainder][column] = make([]superPod, 0, 1)
		}
		firstLevelRemainTop[remainder][column] = append(firstLevelRemainTop[remainder][column], sp)
	}

	return superPodInfo{
		firstLevel:     firstLevelRemainTop,
		countVSuperPod: countVSuperPod,
	}, nil
}

func (tp *module800SuperPod) selectNodes(unReadyID []string, spi *superPodInfo,
	selectNodes map[string][]plugin.SuperNode, maxId int) {
	totalCount := len(unReadyID)
	if spi == nil {
		klog.V(util.LogErrorLev).Info("select nodes failed, super pod info is nil")
		return
	}
	tp.selectFromSmallerSuperPods(unReadyID, spi, selectNodes, &totalCount)
	tp.selectFromBiggerSuperPods(unReadyID, spi, selectNodes, &totalCount)
	tp.selectFromSuperPodsWithReserve(unReadyID, spi, selectNodes, &totalCount)
	tp.selectFromSuperPodsWithSoftStrategy(unReadyID, spi, selectNodes, &totalCount, maxId)
}

func (tp *module800SuperPod) selectFromSmallerSuperPods(unReadyID []string, spi *superPodInfo,
	selectNodes map[string][]plugin.SuperNode, totalCount *int) {
	klog.V(util.LogInfoLev).Infof("select from smaller super pods, totalCount: %d", *totalCount)
	for i := 0; i < len(spi.firstLevel); i++ {
		for j := 1; j < len(spi.firstLevel[0]) && j <= *totalCount; j++ {
			if *totalCount == 0 {
				return
			}
			spi.firstLevel[i][j] = tp.selectNodesFromSuperPodsExceptReserve(unReadyID, totalCount,
				spi.firstLevel[i][j], selectNodes)
		}
	}
}

func (tp *module800SuperPod) selectFromBiggerSuperPods(unReadyID []string, spi *superPodInfo,
	selectNodes map[string][]plugin.SuperNode, totalCount *int) {
	klog.V(util.LogInfoLev).Infof("select from bigger super pods, totalCount: %d", *totalCount)
	for j := *totalCount + 1; j < len(spi.firstLevel[0]); j++ {
		for i := 0; i < len(spi.firstLevel); i++ {
			if *totalCount == 0 {
				return
			}
			spi.firstLevel[i][j] = tp.selectNodesFromSuperPodsExceptReserve(unReadyID, totalCount,
				spi.firstLevel[i][j], selectNodes)
		}
	}
}

func (tp *module800SuperPod) selectFromSuperPodsWithReserve(unReadyID []string, spi *superPodInfo,
	selectNodes map[string][]plugin.SuperNode, totalCount *int) {
	klog.V(util.LogInfoLev).Infof("select from super pods which is less than sp block when except reserve, "+
		"totalCount: %d", *totalCount)
	for i := tp.spBlock - 1; i > -1; i-- {
		for j := 0; j < len(spi.firstLevel[0]); j++ {
			if *totalCount == 0 {
				return
			}
			tp.selectNodesFromSuperPods(unReadyID, totalCount, spi.firstLevel[i][j], selectNodes)
		}
	}
}

func (tp *module800SuperPod) selectNodesFromSuperPodsExceptReserve(unReadyID []string, totalCount *int,
	superPods []superPod, selectNodes map[string][]plugin.SuperNode) []superPod {
	for i := 0; i < len(superPods); i++ {
		klog.V(util.LogInfoLev).Infof("totalCount: %d, len of superPods: %d", *totalCount, len(superPods))
		for {
			if *totalCount == 0 {
				return superPods
			}
			if *totalCount-1 >= len(unReadyID) {
				klog.V(util.LogErrorLev).Infof("index out of range, totalCount: %d, unReadyID: %d",
					*totalCount-1, len(unReadyID))
				return superPods
			}
			if len(superPods[i])-tp.FrameAttr.ReservePodSize < tp.spBlock {
				klog.V(util.LogInfoLev).Infof("num(%d) of superPods[%d] is less than spBlock when except reserve, "+
					"skip this superPod", len(superPods[i]), i)
				break
			}
			superPods[i] = tp.selectNodesFromSuperPod(unReadyID[*totalCount-1], superPods[i], selectNodes)
			klog.V(util.LogInfoLev).Infof("after select, len of reserveNodes: %d", len(superPods[i]))
			*totalCount--
			if len(superPods[i])-tp.FrameAttr.ReservePodSize < tp.spBlock {
				break
			}
		}
	}
	return superPods
}

func (tp *module800SuperPod) selectNodesFromSuperPods(unReadyID []string, totalCount *int,
	superPods []superPod, selectNodes map[string][]plugin.SuperNode) []superPod {
	sort.Slice(superPods, func(i, j int) bool {
		return len(superPods[i]) > len(superPods[j])
	})
	k := 0
	for {
		if *totalCount == 0 || k == len(superPods) {
			return superPods
		}
		if *totalCount-1 >= len(unReadyID) {
			klog.V(util.LogErrorLev).Infof("index out of range, totalCount: %d, unReadyID: %d",
				*totalCount-1, len(unReadyID))
			return superPods
		}
		if len(superPods[k]) < tp.spBlock {
			k++
			continue
		}
		superPods[k] = tp.selectNodesFromSuperPod(unReadyID[*totalCount-1], superPods[k], selectNodes)
		*totalCount--
		k++
	}
}

func (tp *module800SuperPod) selectNodesFromSuperPod(vid string, superPod map[string]plugin.NPUNode,
	selectNodes map[string][]plugin.SuperNode) map[string]plugin.NPUNode {
	count := 0
	reserveNode := make(map[string]plugin.NPUNode, len(superPod)-tp.spBlock)
	if selectNodes == nil {
		return reserveNode
	}
	if len(superPod) < tp.spBlock && !tp.isSoftSuperPodAffinity {
		return superPod
	}
	rescheduleCache := rescheduling.GetReSchedulerCache()
	subHealthyNodes := make([]plugin.NPUNode, 0, len(superPod))
	for _, nNode := range superPod {
		subHealthy := false
		if rescheduleCache != nil {
			fNode, exist := rescheduleCache.FaultNodes[nNode.Name]
			if exist && (fNode.HasCardSubHealthFault || fNode.HasSwitchSubHealthFault) {
				subHealthy = true
			}
		}
		if count >= tp.spBlock || subHealthy {
			reserveNode[nNode.Name] = nNode
			subHealthyNodes = append(subHealthyNodes, nNode)
			continue
		}
		klog.V(util.LogInfoLev).Infof("select nNode %s, super-pod ID: %d", nNode.Name, nNode.SuperPodID)
		_, ok := selectNodes[vid]
		if !ok {
			selectNodes[vid] = make([]plugin.SuperNode, 0)
		}
		selectNodes[vid] = append(selectNodes[vid], plugin.SuperNode{
			Name:       nNode.Name,
			SuperPodID: nNode.SuperPodID,
		})
		count++
	}
	if count >= tp.spBlock {
		return reserveNode
	}
	for i := 0; i < len(subHealthyNodes) && count < tp.spBlock; i++ {
		selectNodes[vid] = append(selectNodes[vid], plugin.SuperNode{
			Name:       subHealthyNodes[i].Name,
			SuperPodID: subHealthyNodes[i].SuperPodID,
		})
		count++
		delete(reserveNode, subHealthyNodes[i].Name)
	}
	return reserveNode
}

func (tp *module800SuperPod) selectFromSuperPodsWithSoftStrategy(unReadyID []string, spi *superPodInfo,
	selectNodes map[string][]plugin.SuperNode, totalCount *int, maxId int) {
	selectNodeNum := 0
	for _, spNodes := range selectNodes {
		selectNodeNum += len(spNodes)
	}
	needNode := tp.NPUTaskNum - selectNodeNum
	if needNode <= 0 {
		return
	}
	recorder := &vPodIdRecorder{unReadyId: unReadyID, leftIndex: *totalCount - 1, rightIndex: maxId}
	klog.V(util.LogWarningLev).Infof("select from super pods which is less than sp block, totalNodes: %d", needNode)
	klog.V(util.LogWarningLev).Infof("job <%s> will scheduling as soft strategy", tp.Name)
	// select super pod which is less than sp block and more than reserve nodes
	for i := tp.spBlock - 1; i >= 0 && i >= tp.FrameAttr.ReservePodSize; i-- {
		for j := 0; j < len(spi.firstLevel[0]); j++ {
			tp.selectNodesForSoftStrategy(recorder, &needNode, spi.firstLevel[i][j], selectNodes)
		}
	}

	// select super pod which is less than reserve nodes
	for i := 0; i < tp.spBlock && i < tp.FrameAttr.ReservePodSize; i++ {
		for j := 0; j < len(spi.firstLevel[0]); j++ {
			tp.selectNodesForSoftStrategy(recorder, &needNode, spi.firstLevel[i][j], selectNodes)
		}
	}
}

func (tp *module800SuperPod) selectNodesForSoftStrategy(recorder *vPodIdRecorder, totalNode *int,
	superPods []superPod, selectNodes map[string][]plugin.SuperNode) []superPod {
	sort.Slice(superPods, func(i, j int) bool {
		return len(superPods[i]) > len(superPods[j])
	})
	for k := 0; k < len(superPods); k++ {
		if *totalNode <= 0 {
			return superPods
		}
		if len(superPods[k]) == 0 {
			break
		}
		nodeNum := len(superPods[k])
		vid := recorder.getVPodID()
		if vid == "" {
			break
		}
		superPods[k] = tp.selectNodesFromSuperPod(vid, superPods[k], selectNodes)
		*totalNode = *totalNode - nodeNum + len(superPods[k])
	}
	return superPods
}

func (r *vPodIdRecorder) getVPodID() string {
	if r.unReadyId == nil || r.leftIndex >= len(r.unReadyId) {
		return ""
	}
	if r.leftIndex < 0 {
		ans := strconv.Itoa(r.rightIndex)
		r.rightIndex++
		return ans
	}
	ans := r.unReadyId[r.leftIndex]
	r.leftIndex--
	return ans
}

// UseAnnotation select npu for task from node
func (tp *module800SuperPod) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("UseAnnotation %s.", err)
		return nil
	}
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation task<%s> node<%s> resource<%s> Annotation: %#v",
		tp.GetPluginName(), task.Name, node.Name, tp.GetAnnoName(), node.Annotation)
	selectedNPU, err := tp.selectNPUFromNode(task, node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation err:%s.", tp.GetPluginName(), err)
		return nil
	}
	klog.V(util.LogInfoLev).Infof("%s UseAnnotation %s select %v from node %s.", tp.GetPluginName(), task.Name,
		selectedNPU, node.Name)

	tp.SetNPUTopologyToPodFn(task, selectedNPU, node)
	newNode := tp.UpdateNodeInfo(node, selectedNPU)
	task.Pod.Annotations[superPodRankKey] = tp.nodeVPodId[node.Name]
	task.Pod.Annotations[superPodIdKey] = strconv.Itoa(int(node.SuperPodID))

	return newNode
}

func (tp *module800SuperPod) selectNPUFromNode(task *api.TaskInfo, node plugin.NPUNode) ([]int, error) {
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	npuTop, err := tp.GetUsableTopFromNode(node, tp.NPUTaskNum/tp.spBlock > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof(getNPUFromPodFailedPattern, tp.GetPluginName(), err.Error())
		return nil, err
	}
	if tp.NPUTaskNum > 1 {
		if len(npuTop) == nodeNPUNumber {
			return npuTop, nil
		}
		return nil, fmt.Errorf("node<%s> top<%v> can not meet task req<%d>", node.Name, len(npuTop), taskNPUNum)
	}

	return tp.selectNPUForStandaloneJob(taskNPUNum, npuTop, node)
}

func (tp *module800SuperPod) selectNPUForStandaloneJob(taskNPUNum int, npuTop []int,
	node plugin.NPUNode) ([]int, error) {
	sort.Ints(npuTop)
	klog.V(util.LogInfoLev).Infof("%s select %d NPU Node(%s) nodeTop<%v>", tp.GetPluginName(), taskNPUNum,
		node.Name, npuTop)
	return npuTop[:taskNPUNum], nil
}

// ReleaseAnnotation Release used resource.
func (tp *module800SuperPod) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}

// prevent division by zero
func (tp *module800SuperPod) checkSpBlockGtZero() bool {
	if tp.spBlock > 0 {
		return true
	}
	klog.V(util.LogErrorLev).Infof("sp-block is less than 0")
	return false
}

func getSuperPodRanks(job plugin.SchedulerJob, rank int) (string, int) {
	// 1. Collect and sort all SuperPod ranks
	sortedRanks := make([]int, 0, len(job.SuperPods))
	for key := range job.SuperPods {
		rankVal, err := strconv.Atoi(key)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("Invalid SuperPod key: %s", key)
			continue
		}
		sortedRanks = append(sortedRanks, rankVal)
	}
	sort.Ints(sortedRanks)

	// 2. Calculate cumulative node count and find matching SuperPod
	cumulativeNodes := 0
	for _, spRank := range sortedRanks {
		spKey := strconv.Itoa(spRank)
		nodeCount := len(job.SuperPods[spKey])

		// 3. Check if rank falls within current SuperPod range
		if rank < cumulativeNodes+nodeCount {
			localRank := rank - cumulativeNodes
			return spKey, localRank
		}
		cumulativeNodes += nodeCount
	}

	// 4. No matching SuperPod found
	klog.V(util.LogErrorLev).Infof(
		"Rank %d exceeds total SuperPod nodes (%d)",
		rank, cumulativeNodes,
	)
	return "", 0
}
