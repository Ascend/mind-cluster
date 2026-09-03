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
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

var TopologyAwarePreemptActive bool

func (tp *chipHandler) Preemptable(preemptor *api.TaskInfo, preemptees []*api.TaskInfo,
	vcNode *plugin.NPUNode) ([]*api.TaskInfo, bool) {
	return tp.routeEviction(preemptor, preemptees, vcNode)
}

func (tp *chipHandler) Reclaimable(reclaimer *api.TaskInfo, reclaimees []*api.TaskInfo,
	vcNode *plugin.NPUNode) ([]*api.TaskInfo, bool) {
	// Reclaim always goes through selective eviction regardless of
	// TopologyAwarePreemptActive: the topology-aware path returns the full
	// candidate set, which is acceptable for preemption but too aggressive
	// for reclaim.
	return tp.preemptOrReclaimSelect(reclaimer, reclaimees, vcNode)
}

func (tp *chipHandler) preemptOrReclaimSelect(preemptor *api.TaskInfo, preemptees []*api.TaskInfo,
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
	if tp.ScheduleMode == util.HardScheduleMode {
		return root.SelectEvictHard(req, allow, preemptees)
	}
	return root.SelectEvictSoft(req, allow, preemptees)
}
