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

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// New return npu plugin
func New(name string) base.AscendHandler {
	m := &module910bx8{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetMaxNodeNPUNum(nodeNPUNumber)
	m.SetAcceleratorValue(util.JobKind910BValue)
	m.SetIsNetworkFaultAttention(true)
	m.InitVNPU()
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
	for _, handler := range tp.PolicyHandler {
		if validResult := handler.ValidNPUJob(); validResult != nil && !validResult.Pass {
			return validResult
		}
	}
	return tp.Valid910bNPUJob()
}

// PreStartAction pre-processing actions for rescheduling
func (tp *module910bx8) PreStartAction(i interface{}, ssn *framework.Session) error {
	k, ok := i.(*rescheduling.ReScheduler)
	if !ok {
		return fmt.Errorf("preStartAction failed %s, interface is not ReScheduler", SchedulerName)
	}
	tp.ReHandle = k
	if vErr := tp.PreStartVNPU(ssn); vErr != nil {
		return fmt.Errorf("preStartVNPU failed %s, err is %s", SchedulerName, vErr)
	}
	for _, handler := range tp.PolicyHandler {
		if err := handler.PreStartAction(i, ssn); err != nil {
			return fmt.Errorf("preStartAction %s", err.Error())
		}
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
		if err := tp.NPUHandler.CheckNodeNPUByTask(task, node); err != nil {
			return err
		}
	default:
		return nil
	}
	for _, handler := range tp.PolicyHandler {
		if err := handler.CheckNodeNPUByTask(task, node); err != nil {
			return fmt.Errorf("checkNodeNPUByTask %s", err.Error())
		}
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
	for _, handler := range tp.PolicyHandler {
		if err := handler.ScoreBestNPUNodes(task, nodes, sMap); err != nil {
			return err
		}
	}
	return tp.NPUHandler.ScoreBestNPUNodes(task, nodes, sMap)
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
	for _, handler := range tp.PolicyHandler {
		handler.UseAnnotation(task, node)
	}
	return tp.NPUHandler.UseAnnotation(task, node)
}
