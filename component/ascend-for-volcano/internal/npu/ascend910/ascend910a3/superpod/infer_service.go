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

/*
Package superpod is using for A3 SuperPod affinity schedule.
*/
package superpod

import (
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base/inferservice"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// isInferServiceJobCheck checks whether current job is an infer service job by reading
// the inferServiceID label. If it is, the inferServiceID is cached on the handler.
func (tp *module910SuperPod) isInferServiceJobCheck() bool {
	if id := inferservice.GetInferServiceID(tp.Label); id != "" {
		tp.inferServiceID = id
		return true
	}
	return false
}

// selectNodesForInferService selects nodes for an infer service job with same-super-pod
// soft affinity, implemented by the shared inferservice package.
func (tp *module910SuperPod) selectNodesForInferService(task *api.TaskInfo,
	nodes []*api.NodeInfo) (map[string][]plugin.SuperNode, error) {
	return inferservice.SelectNodesForInferService(inferservice.InferServiceReq{
		Jobs:           tp.ScheduleEnv.Jobs,
		JobName:        tp.Name,
		InferServiceID: tp.inferServiceID,
		SpBlock:        tp.spBlock,
		ReqNPUNum:      tp.ReqNPUNum,
		SpBlockNPUNum:  tp.SpBlockNPUNum,
		SuperPodTop:    tp.getSuperPodTop(nodes),
	})
}
