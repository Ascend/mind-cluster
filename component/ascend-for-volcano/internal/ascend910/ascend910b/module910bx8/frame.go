/*
Copyright(C)2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package module910bx8 is using for HuaWei Ascend910Bx8 pin affinity schedule.
*/
package module910bx8

import (
	"errors"
	"fmt"
	"reflect"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// New return npu plugin
func New(name string) base.AscendHandler {
	m := &module910bx8{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetDefaultJobSchedulerConfig(nil)
	m.SetMaxNodeNPUNum(nodeNPUNumber)
	m.SetAcceleratorValue(util.JobKind910BValue)
	m.InitVNPU()
	m.netUnhealthyKey = networkUnhealthyNPU
	m.AffScoreList = [][]int{
		{util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3, util.AffScore4, util.AffScore5,
			util.AffScore6, util.AffScore7},
		{util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3, util.AffScore4,
			util.AffScore5, util.AffScore6},
		{util.AffScore8, util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2, util.AffScore3,
			util.AffScore4, util.AffScore5},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore0, util.AffScore1, util.AffScore2,
			util.AffScore3, util.AffScore4},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore0, util.AffScore1,
			util.AffScore2, util.AffScore3},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore0,
			util.AffScore1, util.AffScore2},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore0, util.AffScore1},
		{util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8, util.AffScore8,
			util.AffScore8, util.AffScore0},
	}
	return m
}

// ValidNPUJob check job req npu num and mode
func (tp *module910bx8) ValidNPUJob() *api.ValidateResult {
	if tp.VJob.Type == util.JobTypeDyCut {
		return tp.ValidDyVNPUJob()
	}
	if err := tp.Valid910bNPUJob(); err != nil {
		return err
	}
	return tp.reHandle.ValidJobByReschedule(tp.SchedulerJobAttr)
}

// PreStartAction pre-processing actions for rescheduling
func (tp *module910bx8) PreStartAction(i interface{}, ssn *framework.Session) error {
	k, ok := i.(*rescheduling.ReScheduler)
	if !ok {
		return fmt.Errorf("preStartAction failed %s, interface is not ReScheduler", SchedulerName)
	}
	tp.reHandle = k
	if vErr := tp.PreStartVNPU(ssn); vErr != nil {
		return fmt.Errorf("preStartVNPU failed %s, err is %s", SchedulerName, vErr)
	}
	return nil
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *module910bx8) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err)
		return err
	}
	switch tp.VJob.Type {
	case util.JobTypeDyCut:
		taskRes, err := tp.VHandle.GetTaskResource(task, node)
		if err != nil {
			return err
		}
		if err := tp.VHandle.CheckNodeNPUByDyTask(task, node, taskRes); err != nil {
			return err
		}
	case util.JobTypeWhole:
		if err := tp.checkNodeNPUForWholeCard(task, node); err != nil {
			return err
		}
	default:
		return nil
	}

	if tp.reHandle != nil {
		if reErr := tp.reHandle.CheckNodeNPUByTask(task, node, tp.ReqNPUName); reErr != nil {
			return fmt.Errorf("rescheduling %s", reErr.Error())
		}
	}
	return nil
}

func (tp *module910bx8) checkNodeNPUForWholeCard(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err)
		return err
	}
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %s", tp.GetPluginName(), err.Error())
		return err
	}
	nodeTop, err := tp.getUsableTopFromNode(node, tp.NPUTaskNum > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	if err = tp.JudgeNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogErrorLev).Infof("%s JudgeNodeAndTaskNPU err: %s", tp.GetPluginName(), err.Error())
		return fmt.Errorf("npu topology not meet job require,network unhealthy card is [ %s ]",
			node.Annotation[tp.netUnhealthyKey])
	}
	return nil
}

func (tp *module910bx8) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, sMap map[string]float64) error {
	if tp == nil || task == nil || len(nodes) == 0 || len(sMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %s.", err)
		return err
	}
	if tp.VJob.Type == util.JobTypeDyCut {
		return tp.VHandle.DynamicVNPU.ScoreBestNPUNodes(task, nodes, sMap)
	}
	taskNPUNum, getErr := tp.GetTaskReqNPUNum(task)
	if getErr != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum %s: %s", tp.GetPluginName(), task.Name, getErr)
		return getErr
	}
	for _, node := range nodes {
		if reflect.ValueOf(node).IsNil() {
			continue
		}
		bestScore, healthyNPUNum, getSuccess := tp.getBestScoreAndHealthyNPUNum(task, node, taskNPUNum)
		if !getSuccess {
			continue
		}

		sMap[node.Name] = float64(tp.MaxNodeNPUNum * (int(healthyNPUNum/util.NPUHexKilo) - bestScore))
	}
	reErr := tp.reHandle.ScoreBestNPUNodes(task, sMap)
	if reErr != nil {
		klog.V(util.LogErrorLev).Infof(
			"%s rescheduling ScoreBestNPUNodes failed :%s.", SchedulerName, reErr.Error())
	}
	klog.V(util.LogInfoLev).Infof("%s ScoreBestNPUNodes task<%s> sMap<%v>", tp.GetPluginName(),
		task.Name, sMap)
	return nil
}

func (tp *module910bx8) getBestScoreAndHealthyNPUNum(task *api.TaskInfo,
	node *api.NodeInfo, taskNPUNum int) (int, float64, bool) {
	var bestScore int = 0
	var healthyNPUNum float64 = 0
	nNode, ok := tp.Nodes[node.Name]
	if !ok {
		klog.V(util.LogWarningLev).Infof("%s %s ScoreBestNPUNodes %s is not npu node",
			tp.GetPluginName(), task.Name, node.Name)
		return bestScore, healthyNPUNum, false
	}
	cardIds, err := tp.getUsableTopFromNode(nNode, tp.NPUTaskNum > 1)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes getErr: %s", tp.GetPluginName(), err)
		return bestScore, healthyNPUNum, false
	}
	bestScore, err = tp.getNodeBestScore(taskNPUNum, cardIds)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes getErr: %s", tp.GetPluginName(), err)
		return bestScore, healthyNPUNum, false
	}
	healthyNPUNum, ok = nNode.Allocate[v1.ResourceName(tp.GetAnnoName())]
	if !ok {
		klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes node<%s> get allocate npu failed",
			tp.GetPluginName(), node.Name)
		return bestScore, healthyNPUNum, false
	}
	return bestScore, healthyNPUNum, true
}

// UseAnnotation select npu for task from node
func (tp *module910bx8) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("UseAnnotation %s.", err)
		return nil
	}

	if tp.VJob.Type == util.JobTypeDyCut {
		taskRes, err := tp.VHandle.GetTaskResource(task, node)
		klog.V(util.LogDebugLev).Infof("task<%s> require resource<%#v>", task.Name, taskRes)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("%s UseAnnotation job(%s) get require task resource failed: %s",
				tp.GetPluginName(), tp.Name, err)
		}
		return tp.VHandle.DynamicVNPU.UseAnnotation(task, node, taskRes, tp.VHandle.VT)
	}

	klog.V(util.LogDebugLev).Infof("%s UseAnnotation task<%s> node<%s> resource<%s> Annotation: %s",
		tp.GetPluginName(), task.Name, node.Name, tp.GetAnnoName(), util.SafePrint(node.Annotation))
	selectedNPU, err := tp.selectNPUFromNode(task, node)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation err:%s.", tp.GetPluginName(), err)
		return nil
	}
	klog.V(util.LogInfoLev).Infof("%s UseAnnotation %s select %v.", tp.GetPluginName(), task.Name, selectedNPU)

	tp.SetNPUTopologyToPodFn(task, selectedNPU, node)
	newNode := tp.UpdateNodeInfo(node, selectedNPU)
	return newNode
}

func (tp *module910bx8) selectNPUFromNode(task *api.TaskInfo, node plugin.NPUNode) ([]int, error) {
	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	nodeTop, err := tp.getUsableTopFromNode(node, tp.NPUTaskNum > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s getUsableTopFromNode err: %s", tp.GetPluginName(), err.Error())
		return nil, err
	}
	if len(nodeTop) < taskNPUNum {
		err = fmt.Errorf("node<%s> top<%v> can not meet task req<%d>", node.Name, len(nodeTop), taskNPUNum)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes err: %s", err)
		return nil, err
	}
	return nodeTop[:taskNPUNum], nil
}

// ReleaseAnnotation Release used resource.
func (tp *module910bx8) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
