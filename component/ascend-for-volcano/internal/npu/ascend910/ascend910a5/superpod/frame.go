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

// Package superpod for a5 schedule handler
package superpod

import (
	"errors"
	"fmt"
	"strconv"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// New return npu plugin
func New(name string) *module910a5SuperPod {
	m := &module910a5SuperPod{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetMaxNodeNPUNum(nodeNPUNum)
	m.netUnhealthyKey = networkUnhealthyNPU
	m.faultNPUKey = faultNPU
	newScheduleStrategy()
	return m
}

// ValidNPUJob check jobs' required NPU number and mode.
// ssn.AddJobValidFn -> JobValid -> Job.ValidJobFn -> ValidNPUJob
func (tp *module910a5SuperPod) ValidNPUJob() *api.ValidateResult {
	errStr := "check npu job failed"
	if tp == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("%s, err is %v", errStr, err)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  err.Error(),
			Message: errStr,
		}
	}
	// register all check func in order
	checkers := []jobCheckerFunc{
		tp.checkSpBlock,
		tp.checkSuperPodSizeValid,
		tp.checkTpBlockNum,
		tp.calculateTpBlockAndCheck,
		tp.checkJobReqNpuNum,
	}
	for _, checker := range checkers {
		if err := checker(); err != nil {
			klog.V(util.LogErrorLev).Infof("%s %s", errStr, err.Message)
			return err
		}
	}

	return nil
}

// CheckNodeNPUByTask to check node NPU for each task
// ssn.AddPredicateFn -> NodePredicate -> CheckNodeNPUByTask -> filter node for score
func (tp *module910a5SuperPod) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	errStr := "check npu node by task failed"
	// valid argument
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("%s, err is %v", errStr, err)
		return err
	}
	checkers := []nodeCheckerFunc{
		tp.checkNodeStaticParams,
		tp.checkNodeNPUNums,
	}

	for _, checker := range checkers {
		if err := checker(task, node); err != nil {
			klog.V(util.LogErrorLev).Infof("%s %s", errStr, err.Error())
			return err
		}
	}
	return nil
}

// choose one schedule strategy will use by NPUTaskNum
func (tp *module910a5SuperPod) getStrategyNameByNPUTaskNum() strategyKey {
	var ret strategyKey = RackSchedule
	if tp.NPUTaskNum <= npuTaskNum8 {
		ret = RackSchedule
	} else if tp.NPUTaskNum <= tp.FrameAttr.SuperPodSize {
		ret = SuperPodSchedule
	} else {
		ret = MulSuperPodsSchedule
	}
	return ret
}

// ScoreBestNPUNodes to score the nodes for Jobs, we should know all nodes to get rack topo and superpod topo
func (tp *module910a5SuperPod) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo,
	sMap map[string]float64) error {
	if tp == nil || task == nil || len(nodes) == 0 || len(sMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("score best NPU nodes err: %s.", err)
		return err
	}

	job, ok := tp.ScheduleEnv.Jobs[task.Job]
	if !ok {
		return fmt.Errorf("%s score best  NPU nodes %s: job does not exist", tp.GetPluginName(), task.Name)
	}

	defer func() {
		tp.ScheduleEnv.Jobs[task.Job] = job
	}()

	if !*job.JobReadyTag {
		return nil
	}

	defer tp.scoreNodesForJob(&job, task, sMap)

	// already selected nodes for this job, don't do that again
	if tp.isJobCacheSuperPod(&job, task) {
		return nil
	}

	if tp.NPUTaskNum == 1 {
		nodes = tp.selectNodeForStandaloneJob(nodes)
	}

	tp.isSoftSuperPodAffinity = tp.Label[superPodAffinity] == softRequire
	selectedSpBlock, err := tp.selectNodesForJob(task, nodes)
	job.TpBlock = tp.tpBlock
	job.WhetherBackToVspSchedule = tp.whetherBackToVspSchedule
	if err != nil {
		*job.JobReadyTag = false
		return err
	}
	*job.JobReadyTag = true
	job.SuperPods = selectedSpBlock
	klog.V(util.LogInfoLev).Infof("selectedNodes in every sp-block information:%v", selectedSpBlock)

	return nil
}

// select nodes depending on network topology type
func (tp *module910a5SuperPod) selectNodesForJob(task *api.TaskInfo,
	nodes []*api.NodeInfo) (map[string][]plugin.SuperNode, error) {
	if tp.spBlock == 0 {
		return nil, errors.New("the spBlock value is zero")
	}

	var err error
	selectedNodes := make(map[string][]plugin.SuperNode)

	klog.V(util.LogInfoLev).Infof("input nodes num(%d) for task %s", len(nodes), task.Name)

	superPodMap := getSuperPodMap(tp.Nodes, nodes, tp.GetPluginName())

	spBlockCount := tp.NPUTaskNum / tp.spBlock
	spBlockIDs := make(map[string]bool, spBlockCount)
	for i := 0; i < spBlockCount; i++ {
		spBlockIDs[strconv.Itoa(i)] = false
	}

	err = tp.selectNodesForFaultJob(task, superPodMap, spBlockIDs, selectedNodes)
	if err != nil {
		return nil, err
	}

	var unReadyIds []string
	for id, ready := range spBlockIDs {
		if !ready {
			unReadyIds = append(unReadyIds, id)
		}
	}
	util.SortByNumericValue(unReadyIds)

	strategyInitFactory(tp, unReadyIds, selectedNodes)

	err = tp.selectNodesFromSuperPods(superPodMap, unReadyIds, selectedNodes)

	if err != nil {
		klog.V(util.LogErrorLev).Infof("get error when selecting nodes for job, error: %v", err)
		return nil, err
	}

	return selectedNodes, nil
}

// select nodes from original superpods
func (tp *module910a5SuperPod) selectNodesFromSuperPods(superPodMap map[int32]superPod,
	unReadyIds []string, selectNodes map[string][]plugin.SuperNode) error {
	klog.V(util.LogInfoLev).Infof("selectNodes after selectNodesForFaultJob:%v", selectNodes)

	superPodTopoInfo, err := getSuperPodsInfo(superPodMap, tp.FrameAttr.SuperPodSize, tp.spBlock)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("classify super pod failed!")
		return err
	}
	if len(unReadyIds) == 0 {
		klog.V(util.LogInfoLev).Infof("all nodes have been selected before basic scheduling: %v", selectNodes)
		return nil
	}
	strategyName := tp.getStrategyNameByNPUTaskNum()
	strategySpec, ok := strategyMap[strategyName]
	if !ok {
		return fmt.Errorf("scheduling strategy not found, the key is %s", strategyName)
	}

	scheduleErr := tp.selectNodesBySpecStrategy(strategySpec, &superPodTopoInfo)

	return scheduleErr
}

func (tp *module910a5SuperPod) selectNodesBySpecStrategy(
	strategySpec scheduleStrategy, superPodTopoInfo *superPodsInfo) error {
	if strategySpec == nil {
		return errors.New("the scheduling strategy is nil")
	}

	continueTag, err := strategySpec.entrySelect(superPodTopoInfo)
	if err == nil {
		return nil
	}

	klog.V(util.LogErrorLev).Infof("enforce scheduling strategy failed: error: %v; schedule strategy: %s;",
		err, strategySpec.getStrategyName())

	if !continueTag {
		klog.V(util.LogInfoLev).Infof("stop trying next strategy at the strategy: %s", strategySpec.getStrategyName())
		return err
	}
	// check the validation of the next strategy in the chain
	nextStrategyKey, ok := nextStrategyChain[strategySpec.getStrategyName()]
	if !ok {
		klog.V(util.LogInfoLev).Infof("not found next strategy at the strategy: %s", strategySpec.getStrategyName())
		return err
	}

	return tp.selectNodesBySpecStrategy(strategyMap[nextStrategyKey], superPodTopoInfo)
}

// the real place where we score for nodes, and sMap should change
func (tp *module910a5SuperPod) scoreNodesForJob(job *plugin.SchedulerJob, task *api.TaskInfo, sMap map[string]float64) {
	if !*job.JobReadyTag {
		klog.V(util.LogWarningLev).Infof("job %s has not been ready", job.Name)
		return
	}
	if podGroupEnable, exist := job.Label[plugin.PodGroupScheduleKey]; exist &&
		podGroupEnable == plugin.PodGroupScheduleValue {
		tp.scoreNodeBatchForReadyJob(task, job, sMap)
		return
	}
	tp.scoreNodeForReadyJob(task, job, sMap)
}
