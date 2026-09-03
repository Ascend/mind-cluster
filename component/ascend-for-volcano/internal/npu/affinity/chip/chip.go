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

package chip

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/affinity/chip/topo"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

const PolicyName = "chip-affinity"

const maxNodeNPUNum = 64

type chipHandler struct {
	base.NPUHandler
}

func New() base.AscendHandler {
	h := &chipHandler{}
	h.SetPluginName(PolicyName)
	h.SetMaxNodeNPUNum(maxNodeNPUNum)
	h.SetMaxCardNPUNum(1)
	return h
}

func ShouldUseAffinity(annotation, selector map[string]string) bool {
	if _, ok := annotation[util.SchedulePolicyAnnoKey]; ok {
		return false
	}
	if _, ok := selector[util.AcceleratorTypeKeyDeprecated]; ok {
		return false
	}
	return true
}

func (tp *chipHandler) ValidNPUJob() *api.ValidateResult {
	if tp == nil {
		err := errors.New(util.ArgumentError)
		return &api.ValidateResult{Pass: false, Reason: err.Error(), Message: err.Error()}
	}
	return nil
}

func (tp *chipHandler) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil {
		return errors.New(util.ArgumentError)
	}
	reqName, reqNum := util.GetNPURequestFromTask(task)
	if reqNum < 1 {
		return fmt.Errorf("task<%s> req npu num<%d> is invalid", task.Name, reqNum)
	}

	root := node.ChipTopo
	if root == nil {
		return fmt.Errorf("%s node<%s> has no topology tree", util.NPUResourceUnavailableError, node.Name)
	}

	result := root.Fit(&util.Request{
		ReqNPUName:        reqName,
		ReqNPUNum:         reqNum,
		Mode:              tp.ScheduleMode,
		AllowNetUnhealthy: tp.ParameterPlaneUnhealthyTolerance,
	})
	switch result {
	case topo.FitNormal:
		return nil
	case topo.FitEvict:
		return fmt.Errorf("%s node<%s> don't have enough free npu, req<%d>, can be satisfied by eviction",
			util.NPUResourceShortageError, node.Name, reqNum)
	default: // topo.FitFailed
		return fmt.Errorf("%s node<%s> cannot satisfy task req<%d> even with eviction",
			util.NPUResourceUnavailableError, node.Name, reqNum)
	}
}

func (tp *chipHandler) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	if tp == nil || task == nil || len(nodes) == 0 || len(scoreMap) == 0 {
		return errors.New(util.ArgumentError)
	}
	_, req := util.GetNPURequestFromTask(task)
	if req <= 0 {
		klog.V(util.LogDebugLev).Infof("%s ScoreBestNPUNodes task<%s> req npu num is invalid", tp.GetPluginName(), task.Name)
		return errors.New(util.ArgumentError)
	}
	// job-level tolerance: promoted from the embedded NPUJob attr, mirrors chip.go.
	allow := tp.ParameterPlaneUnhealthyTolerance
	for _, node := range nodes {
		if node == nil {
			continue
		}
		nNode, ok := tp.Nodes[node.Name]
		if !ok {
			continue
		}
		root := nNode.ChipTopo
		if root == nil {
			continue
		}
		scoreMap[node.Name] = root.Score(req, allow)
	}
	return nil
}

func (tp *chipHandler) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if tp == nil || task == nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation err: %s.", tp.GetPluginName(), util.ArgumentError)
		return nil
	}
	_, req := util.GetNPURequestFromTask(task)
	root := node.ChipTopo
	if root == nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation task<%s> node<%s> has no topology tree",
			tp.GetPluginName(), task.Name, node.Name)
		return nil
	}
	selected := root.SelectChips(&util.Request{
		ReqNPUNum:         req,
		Mode:              tp.ScheduleMode,
		AllowNetUnhealthy: tp.ParameterPlaneUnhealthyTolerance,
	})
	if len(selected) < req {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation task<%s> node<%s> cannot select %d chips",
			tp.GetPluginName(), task.Name, node.Name, req)
		return nil
	}
	klog.V(util.LogInfoLev).Infof("%s UseAnnotation task<%s> select %v", tp.GetPluginName(), task.Name, selected)
	if err := root.TryAllocate(string(task.Pod.UID), selected); err != nil {
		klog.V(util.LogErrorLev).Infof("%s UseAnnotation task<%s> node<%s> register chips %v: %v",
			tp.GetPluginName(), task.Name, node.Name, selected, err)
		return nil
	}
	tp.SetNPUTopologyToPodFn(task, selected, node)
	return tp.UpdateNodeInfo(node, selected)
}

// ReleaseAnnotation release annotation: undo the in-memory chip occupancy recorded
// by UseAnnotation so the chips are re-usable in the next scheduling round.
func (tp *chipHandler) ReleaseAnnotation(task *api.TaskInfo, node plugin.NPUNode) *plugin.NPUNode {
	if task == nil {
		klog.V(util.LogErrorLev).Infof("%s ReleaseAnnotation err: %s.", tp.GetPluginName(), util.ArgumentError)
		return nil
	}
	if node.ChipTopo == nil {
		klog.V(util.LogWarningLev).Infof("%s ReleaseAnnotation task<%s> node<%s> has no topology tree",
			tp.GetPluginName(), task.Name, node.Name)
		return &node
	}
	if err := node.ChipTopo.Rollback(string(task.Pod.UID)); err != nil {
		klog.V(util.LogWarningLev).Infof("%s ReleaseAnnotation task<%s> node<%s> rollback: %v",
			tp.GetPluginName(), task.Name, node.Name, err)
	}
	return &node
}
