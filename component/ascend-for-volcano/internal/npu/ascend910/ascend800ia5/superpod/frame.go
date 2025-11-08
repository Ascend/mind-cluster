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
Package superpod is using for HuaWei ascend 800I A5 SuperPod affinity schedule.
*/
package superpod

import (
	"errors"
	"fmt"
	"time"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// New return npu plugin
func New(name string) base.AscendHandler {
	m := &module800SuperPod{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetMaxNodeNPUNum(nodeNPUNumber)
	m.netUnhealthyKey = networkUnhealthyNPU
	m.nodeVPodId = map[string]string{}
	return m
}

// ValidNPUJob verify the validity of job parameters
func (tp *module800SuperPod) ValidNPUJob() *api.ValidateResult {
	if res := tp.checkSpBlock(); res != nil {
		return res
	}
	return tp.checkRequireNPU()
}

func (tp *module800SuperPod) checkSpBlock() *api.ValidateResult {
	if tp.SpBlockNPUNum <= 0 {
		return &api.ValidateResult{
			Pass:    false,
			Reason:  spBlockInvalidReason,
			Message: fmt.Sprintf("sp-block(%d) is invalid", tp.SpBlockNPUNum),
		}
	}
	if tp.SpBlockNPUNum < nodeNPUNumber {
		klog.V(util.LogWarningLev).Info("sp-block less than 8, set default value 1")
		tp.spBlock = 1
	} else {
		if tp.SpBlockNPUNum%nodeNPUNumber != 0 {
			return &api.ValidateResult{
				Pass:    false,
				Reason:  spBlockInvalidReason,
				Message: fmt.Sprintf("sp-block(%d) is not mutiple of node npu (%d)", tp.SpBlockNPUNum, nodeNPUNumber),
			}
		}
		tp.spBlock = tp.SpBlockNPUNum / nodeNPUNumber
	}

	if tp.spBlock > tp.FrameAttr.SuperPodSize {
		return &api.ValidateResult{
			Pass:   false,
			Reason: spBlockInvalidReason,
			Message: fmt.Sprintf("sp-block(%d/8=%d) is bigger than size of super-pod(%d)",
				tp.SpBlockNPUNum, tp.spBlock, tp.FrameAttr.SuperPodSize),
		}
	}
	return nil
}

func (tp *module800SuperPod) checkRequireNPU() *api.ValidateResult {
	if tp.NPUTaskNum == 1 {
		if tp.ReqNPUNum == 1 || tp.ReqNPUNum <= nodeNPUNumber {
			if tp.ReqNPUNum != tp.SpBlockNPUNum {
				return &api.ValidateResult{
					Pass:    false,
					Reason:  jobCheckFailedReason,
					Message: "single super-pod job sp-block annotation should equal require npu num",
				}
			}
			return nil
		}
		return &api.ValidateResult{
			Pass:    false,
			Reason:  jobCheckFailedReason,
			Message: fmt.Sprintf("single super-pod job require npu [1, 2*n], instead of %d", tp.ReqNPUNum),
		}
	}

	// distributed job required npu must be multiple of sp-block
	if tp.ReqNPUNum%tp.SpBlockNPUNum != 0 {
		return &api.ValidateResult{
			Pass:   false,
			Reason: jobCheckFailedReason,
			Message: fmt.Sprintf("distributed super-pod job require npu(%d) should be multiple of sp-block",
				tp.ReqNPUNum),
		}
	}
	return tp.checkReqNPUEqualNodeNPU()
}

func (tp *module800SuperPod) checkReqNPUEqualNodeNPU() *api.ValidateResult {
	for _, task := range tp.Tasks {
		// npu num required by task in distributed job must be node npu num
		if task.ReqNPUNum != nodeNPUNumber {
			return &api.ValidateResult{
				Pass:    false,
				Reason:  jobCheckFailedReason,
				Message: fmt.Sprintf("distributed super-pod job require npu 8*n, instead of %d", task.ReqNPUNum),
			}
		}
	}
	return nil
}

// CheckNodeNPUByTask check nod npu meet task req
func (tp *module800SuperPod) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("CheckNodeNPUByTask err: %s", err)
		return err
	}
	// node in super-pod has super-podID which is not less than 0
	if node.SuperPodID < 0 {
		return fmt.Errorf("node %s is not super-pod node or superPodID is not set", node.Name)
	}

	taskNPUNum, err := tp.GetTaskReqNPUNum(task)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s GetTaskReqNPUNum err: %s", tp.GetPluginName(), err.Error())
		return err
	}

	nodeTop, err := tp.GetUsableTopFromNode(node, tp.NPUTaskNum/tp.spBlock > 1)
	if err != nil {
		klog.V(util.LogErrorLev).Infof(getNPUFromPodFailedPattern, tp.GetPluginName(), err.Error())
		return err
	}

	if err = tp.NPUHandler.JudgeNodeAndTaskNPU(taskNPUNum, nodeTop); err != nil {
		klog.V(util.LogErrorLev).Infof("%s JudgeNodeAndTaskNPU err: %s", tp.GetPluginName(), err.Error())
		return fmt.Errorf("checkNodeNPUByTask %s err: %s", util.NodeNotMeetTopologyWarning, err.Error())
	}
	return nil
}

// isDelayingJob checks if the job should continue waiting for resource release
// Returns true if:
//   - The waiting time exceeds the delayingTime threshold (10s)
//   - All normal nodes used by the job have been released
//
// Returns false if any normal node is still occupied
func (tp *module800SuperPod) isDelayingJob(fJob *rescheduling.FaultJob, nodes []*api.NodeInfo) bool {
	if tp == nil || fJob == nil || fJob.FaultTasks == nil {
		return false
	}
	// Check if waiting time exceeds threshold
	if time.Now().Unix()-fJob.RescheduleTime > delayingTime {
		klog.V(util.LogWarningLev).Infof("job %s wait used resource release time over 10s, skip wait", fJob.JobName)
		return true
	}

	// Convert nodes to map for quick lookup
	nodeMaps := util.ChangeNodesToNodeMaps(nodes)

	// Check all non-fault tasks to see if their nodes are released
	for _, task := range fJob.FaultTasks {
		if task.IsFaultTask {
			continue
		}
		// If node is not in available nodes list, it's still occupied
		if _, ok := nodeMaps[task.NodeName]; !ok {
			klog.V(util.LogWarningLev).Infof("job used %s normal node %s is not release", fJob.JobName, task.NodeName)
			return false
		}
	}
	return true
}
