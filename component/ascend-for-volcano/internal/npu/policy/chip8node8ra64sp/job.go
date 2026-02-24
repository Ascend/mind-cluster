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
Package chip8node8ra64sp is using for job check
*/
package chip8node8ra64sp

import (
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// the basic check of SpBlock & SuperPodSize
func (tp *chip8node8ra64sp) checkSpBlock() *api.ValidateResult {
	if tp.SpBlockNPUNum <= 0 {
		return &api.ValidateResult{
			Pass:    false,
			Reason:  spBlockInvalidReason,
			Message: fmt.Sprintf("Parameter sp-block(%d) is invalid.", tp.SpBlockNPUNum),
		}
	}

	if tp.SpBlockNPUNum < tp.MaxNodeNPUNum {
		tp.spBlock = 1
	} else {
		if tp.SpBlockNPUNum%tp.MaxNodeNPUNum != 0 {
			return &api.ValidateResult{
				Pass:   false,
				Reason: spBlockInvalidReason,
				Message: fmt.Sprintf("Parameter sp-block(%d) is not multiple of node npu (%d)",
					tp.SpBlockNPUNum, tp.MaxNodeNPUNum),
			}
		}
		tp.spBlock = tp.SpBlockNPUNum / tp.MaxNodeNPUNum
	}

	// distributed job required npu must be multiple of sp-block
	if tp.NPUTaskNum%tp.spBlock != 0 {
		return &api.ValidateResult{
			Pass:   false,
			Reason: "job task num is invalid",
			Message: fmt.Sprintf("job require total Pod(%d) should be multiple of a sp-block size %d",
				tp.NPUTaskNum, tp.spBlock),
		}
	}

	return nil
}

func (tp *chip8node8ra64sp) checkSuperPodSizeValid() *api.ValidateResult {
	// getting super-pod-size in volcano.yaml instead of new value changed by process which is different with 910a3
	SuperPodSizeFromConf := tp.FrameAttr.SuperPodSizeFromConf

	//  Max(super-pod-size) * 8 <= 8192
	if SuperPodSizeFromConf <= 0 || SuperPodSizeFromConf*tp.MaxNodeNPUNum > maxSuperPodNPUNum {
		return &api.ValidateResult{
			Pass:   false,
			Reason: superPodSizeInvalidReason,
			Message: fmt.Sprintf("Parameter super-pod-size(%d) in volcano.yaml is invalid "+
				"which should be in range [1,1024]",
				SuperPodSizeFromConf),
		}
	}

	if tp.spBlock > tp.FrameAttr.SuperPodSizeFromConf {
		return &api.ValidateResult{
			Pass:   false,
			Reason: superPodSizeInvalidReason,
			Message: fmt.Sprintf("Parameter spBlock(%d/8=%d) is bigger than size of super-pod-size(%d)",
				tp.SpBlockNPUNum, tp.spBlock, tp.FrameAttr.SuperPodSizeFromConf),
		}
	}
	return nil
}

// check the validation of tp-block
func (tp *chip8node8ra64sp) checkTpBlockNum() *api.ValidateResult {
	// the tp-block value must in range [1,64]
	if tp.TpBlockNPUNum > rackNPUNumber || tp.TpBlockNPUNum < miniTpBlockNum {
		return &api.ValidateResult{
			Pass:   false,
			Reason: tpBlockInvalidReason,
			Message: fmt.Sprintf("Parameter tp-block is invalid, it should be a number in the range "+
				"from %d to %d", miniTpBlockNum, rackNPUNumber),
		}
	}

	// check if tp-block is power of 2 by bitwise operation
	if (tp.TpBlockNPUNum & (tp.TpBlockNPUNum - 1)) != 0 {
		return &api.ValidateResult{
			Pass:    false,
			Reason:  tpBlockInvalidReason,
			Message: fmt.Sprintf("Parameter tp-block(%d) must be the power of 2", tp.TpBlockNPUNum),
		}
	}

	return nil
}

// calculate the tp-block and check if it's valid
func (tp *chip8node8ra64sp) calculateTpBlockAndCheck() *api.ValidateResult {
	// tp-block=1 -> tpBlock=1
	// tp-block=8 -> tpBlock=1
	// tp-block=16 -> tpBlock=2
	// tp-block=32 -> tpBlock=4
	// tp-block=64 -> tpBlock=8
	const (
		plusTpBlockNum = 7
	)
	tp.tpBlock = (tp.TpBlockNPUNum + plusTpBlockNum) / tp.MaxNodeNPUNum

	if tp.tpBlock > tp.spBlock {
		return &api.ValidateResult{
			Pass:   false,
			Reason: tpBlockInvalidReason,
			Message: fmt.Sprintf("Parameter tp-block(%d)/%d could not be bigger than sp-block(%d)/%d",
				tp.TpBlockNPUNum, tp.MaxNodeNPUNum, tp.SpBlockNPUNum, tp.MaxNodeNPUNum),
		}
	}

	if tp.NPUTaskNum%tp.tpBlock != 0 {
		return &api.ValidateResult{
			Pass:   false,
			Reason: tpBlockInvalidReason,
			Message: fmt.Sprintf("number of tasks(%d) must be multiple of "+
				"nodes occupied by tp-block(%d)", tp.NPUTaskNum, tp.tpBlock),
		}
	}

	if tp.spBlock%tp.tpBlock != 0 {
		return &api.ValidateResult{
			Pass:   false,
			Reason: tpBlockInvalidReason,
			Message: fmt.Sprintf("spBlock= %d / 8 must be multiple of tpBlock= %d / 8",
				tp.SpBlockNPUNum, tp.TpBlockNPUNum),
		}
	}

	return nil
}

// check if task ReqNPUNum is valid
func (tp *chip8node8ra64sp) checkJobReqNpuNum() *api.ValidateResult {
	// single job
	if tp.NPUTaskNum == 1 {
		if tp.ReqNPUNum <= tp.MaxNodeNPUNum && tp.ReqNPUNum > 0 {
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
			Message: fmt.Sprintf("single super-pod job require npu [1, 8], instead of %d", tp.ReqNPUNum),
		}
	}
	// distributed job required npu must be multiple of sp-block
	if tp.ReqNPUNum%tp.SpBlockNPUNum != 0 {
		return &api.ValidateResult{
			Pass:   false,
			Reason: jobCheckFailedReason,
			Message: fmt.Sprintf("distributed super-pod job require npu should be multiple of sp-block, instead of %d",
				tp.ReqNPUNum),
		}
	}
	// distributed job required npu must be multiple of tp-block
	if tp.ReqNPUNum%tp.TpBlockNPUNum != 0 {
		return &api.ValidateResult{
			Pass:   false,
			Reason: jobCheckFailedReason,
			Message: fmt.Sprintf("distributed super-pod job require npu(%d) should be multiple of tp-block",
				tp.ReqNPUNum),
		}
	}
	// check the distributed job require npu num must be equal to node npu num
	for _, task := range tp.Tasks {
		if task.ReqNPUNum == tp.MaxNodeNPUNum {
			continue
		}

		return &api.ValidateResult{
			Pass:   false,
			Reason: jobCheckFailedReason,
			Message: fmt.Sprintf("distributed job require npu %d, instead of %d",
				tp.MaxNodeNPUNum, task.ReqNPUNum),
		}
	}
	return nil
}

// check whether the job is invalid in UBMem scene
func (tp *chip8node8ra64sp) checkJobInUBMemScene() *api.ValidateResult {
	if !tp.isUBMemScene {
		return nil
	}
	jobReqNPUNum := 0
	klog.V(util.LogInfoLev).Infof("task is in ubmemory scene.")
	for _, task := range tp.Tasks {
		jobReqNPUNum += task.ReqNPUNum
		if jobReqNPUNum > maxNpuNumInUBMemScene {
			return &api.ValidateResult{
				Pass:   false,
				Reason: jobCheckFailedReason,
				Message: fmt.Sprintf("in ubmem scene, task require npu nums should smaller than %d, but recevice %d",
					maxNpuNumInUBMemScene, jobReqNPUNum),
			}
		}
	}
	return nil
}

func (tp *chip8node8ra64sp) isJobCacheSuperPod(job *plugin.SchedulerJob, task *api.TaskInfo) bool {
	if *job.JobReadyTag && len(job.SuperPods) != 0 {
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes %s: job is ready, skip Schedule",
			tp.GetPluginName(), task.Name)
		return true
	}

	return false
}
