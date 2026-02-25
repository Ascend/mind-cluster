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
Package module910a3x16 is using for A3 x16 affinity schedule.
*/
package module910a3x16

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910a3"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// New return npu plugin
func New(name string) base.AscendHandler {
	m := &module910a3x16{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetMaxNodeNPUNum(ascend910a3.NodeNPUNumber16)
	m.SetMaxCardNPUNum(ascend910a3.DieNPUNumber)
	m.SetIsNetworkFaultAttention(true)
	m.setAffinityScore()
	return m
}

func (tp *module910a3x16) setAffinityScore() {
	tp.AffScoreList = make([][]int, tp.MaxNodeNPUNum)
	for i := 0; i < tp.MaxNodeNPUNum; i++ {
		tp.AffScoreList[i] = make([]int, tp.MaxNodeNPUNum)
		for j := 0; j < tp.MaxNodeNPUNum; j++ {
			if i > j {
				tp.AffScoreList[i][j] = util.AffScore15
			} else {
				tp.AffScoreList[i][j] = j - i
			}
		}
	}
}

// ValidNPUJob check job req npu num and mode
func (tp *module910a3x16) ValidNPUJob() *api.ValidateResult {
	if tp.ReqNPUNum == 1 {
		if tp.NPUTaskNum != 1 {
			return &api.ValidateResult{
				Pass:    false,
				Reason:  util.InvalidResourceRequestReason,
				Message: fmt.Sprintf("only single task job can request single npu"),
			}
		}
		return nil
	}
	if tp.ReqNPUNum%taskNPUMultiple != 0 {
		return &api.ValidateResult{
			Pass:   false,
			Reason: util.InvalidResourceRequestReason,
			Message: fmt.Sprintf("except for single task job can request single npu, "+
				"job require npu num should be even number, instead of %d", tp.ReqNPUNum),
		}
	}
	return tp.CheckTaskNPU()
}

// CheckTaskNPU check the job require npu num
func (tp *module910a3x16) CheckTaskNPU() *api.ValidateResult {
	helper := util.NewTaskValidateHelper()
	for _, task := range tp.Tasks {
		if helper.HasTask(task.TaskSpecKey) {
			continue
		}

		// except for single task job can request single npu, npu num required by task in a3×16 job must be 2n, and <= 16
		if task.ReqNPUNum != 0 && task.ReqNPUNum%taskNPUMultiple == 0 && task.ReqNPUNum <= tp.MaxNodeNPUNum {
			helper.AddValidateTask(task.TaskSpecKey)
			continue
		}

		if task.ReqNPUNum == 0 && (task.Annotation[ascend910a3.TaskSpecAnno] == ascend910a3.SchedulerType ||
			task.Annotation[ascend910a3.SkipAscendPluginAnno] == ascend910a3.SkipEnabled) {
			helper.AddValidateTask(task.TaskSpecKey)
			continue
		}
		helper.AddInvalidResourceRequest(task.TaskSpecKey, task.ReqNPUNum)
	}
	return helper.TaskValidResult(fmt.Sprintf("only standalone job can request single npu, npu num required "+
		"by task in a3×16 job must be 2n, and <= %d", tp.MaxNodeNPUNum))
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *module910a3x16) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogDebugLev).Infof("CheckNodeNPUByTask err: %v", err)
		return err
	}
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s GetTaskReqNPUNum err: %v", task.Name, err)
		return err
	}
	nodeTop, err := tp.GetUsableTopFromNode(node, tp.NPUTaskNum > 1 || tp.IsInstanceOfJobGroup())
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s getUsableTopFromNode err: %v", task.Name, err)
		return err
	}
	if err = tp.JudgeNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogDebugLev).Infof("%s JudgeNodeAndTaskNPU err: %v", task.Name, err)
		return err
	}
	return nil
}

// ScoreBestNPUNodes score best npu nodes
func (tp *module910a3x16) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo,
	sMap map[string]float64) error {
	if tp == nil || task == nil || len(nodes) == 0 || len(sMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %s.", err.Error())
		return err
	}
	taskNPUNum, getErr := tp.GetTaskReqNPUNum(task)
	if getErr != nil {
		klog.V(util.LogDebugLev).Infof("%s GetTaskReqNPUNum %s: %s",
			tp.GetPluginName(), task.Name, getErr.Error())
		return getErr
	}
	for _, node := range nodes {
		if node == nil {
			continue
		}
		bestScore, getSuccess := tp.getBestScoreAndHealthyNPUNum(task, node, taskNPUNum)
		if !getSuccess {
			continue
		}

		sMap[node.Name] = float64(tp.MaxNodeNPUNum * (tp.MaxCardNPUNum*tp.MaxNodeNPUNum - bestScore))
		// fast break for large k8s cluster, if best score is 0, we can not find better score than 0
		if bestScore == util.AffScore0 {
			break
		}
	}
	klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodes task<%s> sMap<%v>", tp.GetPluginName(),
		task.Name, sMap)
	return nil
}

func (tp *module910a3x16) getBestScoreAndHealthyNPUNum(task *api.TaskInfo,
	node *api.NodeInfo, taskNPUNum int) (int, bool) {
	var bestScore = 0
	nNode, ok := tp.Nodes[node.Name]
	if !ok {
		klog.V(util.LogDebugLev).Infof("%s %s ScoreBestNPUNodes %s is not npu node",
			tp.GetPluginName(), task.Name, node.Name)
		return bestScore, false
	}
	cardIds, err := tp.GetUsableTopFromNode(nNode, tp.NPUTaskNum > 1)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodes getErr: %#v", tp.GetPluginName(), err)
		return bestScore, false
	}
	// the most matching node with a score of 0
	bestScore, err = tp.getNodeBestScore(taskNPUNum, cardIds)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodes getErr: %#v", tp.GetPluginName(), err)
		return bestScore, false
	}

	if taskNPUNum != 1 {
		return bestScore, true
	}

	// single die schedule, we should select has single die node first
	npuTopology := tp.GetNodeCardTopology(cardIds)
	hasSingleDie := false
	for _, npu := range npuTopology {
		if len(npu) != tp.MaxCardNPUNum {
			hasSingleDie = true
			break
		}
	}
	if !hasSingleDie {
		bestScore += tp.MaxNodeNPUNum
	}
	return bestScore, true
}

func (tp *module910a3x16) getNodeBestScore(taskNPUNum int, usableNPU []int) (int, error) {
	if taskNPUNum < 1 || taskNPUNum > tp.MaxNodeNPUNum {
		return 0, fmt.Errorf("task req npu num<%d> is invalid", taskNPUNum)
	}
	npuNum := len(usableNPU)
	if npuNum < 1 || npuNum > tp.MaxNodeNPUNum {
		return 0, fmt.Errorf("node npu num<%d> is invalid", npuNum)
	}
	return tp.AffScoreList[taskNPUNum-1][npuNum-1], nil
}

// UseAnnotation select npu for task from node
func (tp *module910a3x16) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("UseAnnotation %s.", err)
		return nil
	}
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation task<%s> node<%s> resource<%s> Annotation: %#v",
		tp.GetPluginName(), task.Name, node.Name, tp.GetAnnoName(tp.ReqNPUName), node.Annotation)
	selectedNPU, err := tp.SelectNPUFromNode(task, node, tp.NPUTaskNum > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation err:%v.", task.Name, err)
		return nil
	}
	klog.V(util.LogDebugLev).Infof("%s UseAnnotation %s select %v from node %s.", tp.GetPluginName(), task.Name,
		selectedNPU, node.Name)

	tp.SetNPUTopologyToPodFn(task, selectedNPU, node)
	newNode := tp.UpdateNodeInfo(node, selectedNPU)
	return newNode
}
