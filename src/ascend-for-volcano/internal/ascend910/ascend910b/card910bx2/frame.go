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
package card910bx2

import (
	"errors"
	"fmt"

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
	m := &card910bx2{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetDefaultJobSchedulerConfig(nil)
	m.SetMaxNodeNPUNum(nodeNPUNumber)
	m.SetAcceleratorValue(util.JobKind910BValue)
	m.SetArch(util.HuaweiArchX86 + util.HuaweiArchArm)
	m.AffScoreList = [][]int{
		{util.AffScore0, util.AffScore1},
		{util.AffScore2, util.AffScore0},
	}
	return m
}

// ValidNPUJob check job req npu num and mode
func (tp *card910bx2) ValidNPUJob() *api.ValidateResult {
	return tp.Valid910bNPUJob()
}

// PreStartAction pre-processing actions for rescheduling
func (tp *card910bx2) PreStartAction(i interface{}, _ *framework.Session) error {
	k, ok := i.(*rescheduling.ReScheduler)
	if !ok {
		return fmt.Errorf("PreStartAction failed %s, interface is not ReScheduler", SchedulerName)
	}
	tp.reHandle = k
	return nil
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *card910bx2) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err)
		return err
	}
	return nil
}

// ScoreBestNPUNodes core node by calculate task req npu num and node npu top
func (tp *card910bx2) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, sMap map[string]float64) error {
	return tp.ScoreAscendNPUNodes(task, nodes, sMap)
}

// UseAnnotation select npu for task from node
func (tp *card910bx2) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return tp.Use910bAnnotation(task, node)
}

// ReleaseAnnotation Release used resource.
func (tp *card910bx2) ReleaseAnnotation(_ *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	return &node
}
