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
Package card910bx2 is using for HuaWei Ascend 910B(Atlas 300T A2) card pin affinity schedule.
*/
package card910bx2infer

import (
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
	m := &card910bx2infer{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetDefaultJobSchedulerConfig(nil)
	m.SetMaxNodeNPUNum(nodeNPUNumber)
	m.SetAcceleratorValue(util.JobKind910BValue)
	return m
}

// PreStartAction pre-processing actions for rescheduling
func (tp *card910bx2infer) PreStartAction(i interface{}, _ *framework.Session) error {
	k, ok := i.(*rescheduling.ReScheduler)
	if !ok {
		return fmt.Errorf("preStartAction failed %s, interface is not ReScheduler", SchedulerName)
	}
	tp.reHandle = k
	return nil
}

// ValidNPUJob check job req npu num and mode
func (tp *card910bx2infer) ValidNPUJob() *api.ValidateResult {
	if tp.NPUTaskNum != 1 {
		klog.V(util.LogErrorLev).Infof("GetVTaskNumInVJob %s has %d npu tasks, only support 1.", tp.Name, tp.NPUTaskNum)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "ValidNPUJob failed",
			Message: fmt.Sprintf("%s task num is invalid", tp.Name),
		}
	}

	return tp.Valid910bNPUJob()
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *card910bx2infer) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	return tp.NPUHandler.CheckNodeNPUByTask(task, node)
}

// ScoreBestNPUNodes core node by calculate task req npu num and node npu top
func (tp *card910bx2infer) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, sMap map[string]float64) error {
	return tp.NPUHandler.ScoreBestNPUNodes(task, nodes, sMap)
}

// UseAnnotation select npu for task from node
func (tp *card910bx2infer) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return tp.NPUHandler.UseAnnotation(task, node)
}

// ReleaseAnnotation Release used resource.
func (tp *card910bx2infer) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
