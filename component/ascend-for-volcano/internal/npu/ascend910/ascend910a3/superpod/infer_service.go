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
	"strconv"

	"k8s.io/klog/v2"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base/inferservice"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
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
// When the job is a fault job with pod-level rescheduling enabled, it dispatches to
// the pod-level branch which keeps healthy super pods and only refills fault slots.
func (tp *module910SuperPod) selectNodesForInferService(task *api.TaskInfo,
	nodes []*api.NodeInfo) (map[string][]plugin.SuperNode, error) {
	// pod-level rescheduling: keep healthy super pods, refill fault slots only
	rescheduleCache := rescheduling.GetReSchedulerCache()
	if rescheduleCache != nil {
		fJob := rescheduleCache.FaultJobs[task.Job]
		if fJob != nil && fJob.IsFaultJob && tp.ifPodLevelRescheduling(fJob) {
			return tp.selectNodesForInferServicePodLevel(task, nodes, fJob)
		}
	}
	// full super-pod selection for job-level rescheduling or first scheduling
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

// selectNodesForInferServicePodLevel selects nodes for a fault infer service job in
// pod-level rescheduling mode. It reuses Stage 1-3 sub-functions of the fault-job
// upgrading chain (keep healthy SP -> same physical SP new nodes -> replace fault
// slots) and falls back to the infer service priority queue in Stage 4. The
// schedulable gate of selectNodesForFaultJob is skipped on purpose: it validates
// resource totals with job-level semantics and would block pod-level Stage 2-4.
func (tp *module910SuperPod) selectNodesForInferServicePodLevel(task *api.TaskInfo,
	nodes []*api.NodeInfo, fJob *rescheduling.FaultJob) (map[string][]plugin.SuperNode, error) {
	totalNodes := tp.getSuperPodTop(nodes)
	totalRequiredSuperPod := tp.NPUTaskNum / tp.spBlock
	vSuperPodID := make(map[string]bool, totalRequiredSuperPod)
	for i := 0; i < totalRequiredSuperPod; i++ {
		vSuperPodID[strconv.Itoa(i)] = false
	}
	selectNodes := make(map[string][]plugin.SuperNode)
	klog.V(util.LogInfoLev).Infof("infer service pod-level: job %s start selecting, spBlock=%d, requiredSP=%d, superPodTop=%d, faultSP=%d",
		tp.Name, tp.spBlock, totalRequiredSuperPod, len(totalNodes), len(fJob.SuperPods))

	// Stage 1: keep healthy super pods. selectNodeFromOriginVSuperPod restores
	// SuperPodReschdInfo first, then dispatches by ifPodLevelRescheduling
	// (the pod-level branch does not read sMap, passing nil is safe).
	notReadySuperPod, _ := tp.selectNodeFromOriginVSuperPod(fJob, nil,
		selectNodes, totalNodes, vSuperPodID)

	// Stage 2: select brand-new node group within the same physical super pod
	tp.selectNodeFromOriginSuperPod(fJob, notReadySuperPod, totalNodes, vSuperPodID, selectNodes)

	// Stage 3: replace fault node slots within the same physical super pod
	tp.selectNodeForPodLevelRescheduling(fJob, notReadySuperPod, totalNodes, vSuperPodID, selectNodes)

	var unReadyID []string
	for id, ready := range vSuperPodID {
		if !ready {
			unReadyID = append(unReadyID, id)
		}
	}
	if len(unReadyID) == 0 {
		klog.V(util.LogInfoLev).Infof("infer service pod-level: job %s all sp-blocks ready after stage 1-3", tp.Name)
		return selectNodes, nil
	}
	util.SortByNumericValue(unReadyID)
	klog.V(util.LogInfoLev).Infof("infer service pod-level: job %s stage 4 fallback for unready sp-blocks %v",
		tp.Name, unReadyID)

	// Stage 4: infer service priority queue (same-service SP first), replacing the
	// plain selectNodes of selectSuperPodForJob which has no affinity
	if err := inferservice.SelectInferServiceSPForPodLevel(tp.ScheduleEnv.Jobs, tp.Name,
		tp.inferServiceID, tp.spBlock, unReadyID, totalNodes, selectNodes, vSuperPodID); err != nil {
		return nil, err
	}
	return selectNodes, nil
}
