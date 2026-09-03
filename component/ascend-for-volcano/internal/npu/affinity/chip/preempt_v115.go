//go:build volcano_v115

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
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/affinity/chip/topo"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func (tp *chipHandler) routeEviction(preemptor *api.TaskInfo, preemptees []*api.TaskInfo,
	vcNode *plugin.NPUNode) ([]*api.TaskInfo, bool) {
	if TopologyAwarePreemptActive {
		return tp.preemptOrReclaim(preemptor, preemptees, vcNode)
	}
	return tp.preemptOrReclaimSelect(preemptor, preemptees, vcNode)
}

func (tp *chipHandler) preemptOrReclaim(preemptor *api.TaskInfo, preemptees []*api.TaskInfo,
	vcNode *plugin.NPUNode) ([]*api.TaskInfo, bool) {
	if tp == nil || preemptor == nil || vcNode == nil {
		return nil, false
	}
	_, req := util.GetNPURequestFromTask(preemptor)
	if req <= 0 {
		return nil, false
	}
	root := vcNode.ChipTopo
	if root == nil {
		return nil, false
	}
	allow := tp.ParameterPlaneUnhealthyTolerance
	hard := tp.ScheduleMode == util.HardScheduleMode
	if hard {
		if root.Fit(&util.Request{ReqNPUNum: req,
			Mode: util.HardScheduleMode, AllowNetUnhealthy: allow}) == topo.FitNormal {
			klog.V(util.LogInfoLev).Infof("Preemptable(chip): task<%s> node<%s> hard zero eviction",
				preemptor.Name, vcNode.Name)
			return nil, true
		}
	} else if root.CountUsable(allow) >= req {
		klog.V(util.LogInfoLev).Infof("Preemptable(chip): task<%s> node<%s> zero eviction, usable=%d",
			preemptor.Name, vcNode.Name, root.CountUsable(allow))
		return nil, true
	}
	reclaim := reclaimableIDsOf(preemptees, vcNode)
	if len(reclaim) == 0 || !evictableFor(root, req, allow, reclaim, hard) {
		klog.V(util.LogInfoLev).Infof("Preemptable(chip): task<%s> node<%s> not satisfiable after "+
			"full eviction (reclaim=%d), abstain", preemptor.Name, vcNode.Name, len(reclaim))
		return nil, false
	}
	return preemptees, true
}

func reclaimableIDsOf(preemptees []*api.TaskInfo, vcNode *plugin.NPUNode) map[int]struct{} {
	res := make(map[int]struct{})
	for _, pe := range preemptees {
		for _, id := range util.GetAllocatedChipIDsFromPod(pe.Pod) {
			res[id] = struct{}{}
		}
	}
	return res
}

func evictableFor(n *topo.ChipNode, req int, allow bool, reclaim map[int]struct{}, hard bool) bool {
	if hard {
		return n.CanFitHardFor(req, allow, reclaim) != topo.FitFailed
	}
	return n.CountEvictSet(allow, reclaim) >= req
}
