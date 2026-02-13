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
Package vnpu is using for Ascend vnpu affinity schedule.
*/
package vnpu

import (
	"errors"
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/vnpu"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func (tp *virtual310NPU) checkDyVJobReq() *api.ValidateResult {
	if !tp.IsVJob() {
		return &api.ValidateResult{Pass: false, Reason: util.InvalidArgumentReason, Message: "not vnpu job"}
	}
	if tp.vHandle.StaticByConf {
		return &api.ValidateResult{Pass: false, Reason: util.InvalidArgumentReason, Message: fmt.Sprintf("volcano configuration %s true, only support static vnpu", util.SegmentEnable)}
	}
	helper := util.NewTaskValidateHelper()
	for _, vT := range tp.Tasks {
		if helper.HasTask(vT.TaskSpecKey) {
			continue
		}
		if vT.ReqNPUNum == 1 || vT.ReqNPUNum == util.NPUIndex2 || vT.ReqNPUNum == util.NPUIndex4 ||
			vT.ReqNPUNum%util.NPUIndex8 == 0 || vT.ReqNPUNum%util.NPUIndex10 == 0 {
			helper.AddValidateTask(vT.TaskSpecKey)
			continue
		}
		helper.AddInvalidResourceRequest(vT.TaskSpecKey, vT.ReqNPUNum)
	}

	return helper.TaskValidResult("task req npu should be [1,2,4,8,10]")
}

func (tp *virtual310NPU) validDyVNPUTaskDVPPLabel(vT util.NPUTask) error {
	if !vT.IsVNPUTask() {
		return errors.New("not vNPU task")
	}

	dvppValue := vnpu.GetVNPUTaskDVPP(vT)

	if vT.ReqNPUNum > 0 && (vT.ReqNPUNum%util.NPUIndex8 == 0 || vT.ReqNPUNum%util.NPUIndex10 == 0) {
		return nil
	}
	switch vT.ReqNPUNum {
	case 1, util.NPUIndex2:
		if dvppValue != plugin.AscendDVPPEnabledNull {
			return fmt.Errorf("%s req %d ai-core, but dvpp label is:%s", vT.TaskSpecKey, vT.ReqNPUNum, dvppValue)
		}
	case util.NPUIndex4:
		return nil
	default:
		return fmt.Errorf("err require number:%d", vT.ReqNPUNum)
	}
	return nil
}

// vNPU tasks, must abide by the following conventions:
// 1.vir01: no vpu-dvpp, if has error; vpu-level ignore.
// 2.vir02: no vpu-dvpp, if has error; vpu-level ignore.
// 3.vir04: ignore vpu-level and vpu-dvpp.
// 4.every task must be vNPU Task.
func (tp *virtual310NPU) validDyVNPUJobLabel() *api.ValidateResult {
	if !tp.IsVJob() {
		return &api.ValidateResult{Pass: false, Reason: util.InvalidArgumentReason, Message: "not vnpu job"}
	}
	helper := util.NewTaskValidateHelper()
	for _, vT := range tp.Tasks {
		if helper.HasTask(vT.TaskSpecKey) {
			continue
		}
		if tErr := tp.validDyVNPUTaskDVPPLabel(vT); tErr != nil {
			helper.AddInvalidTask(vT.TaskSpecKey, tErr)
		}
	}

	return helper.TaskValidResult("")
}

func (tp *virtual310NPU) validDyVNPUJob() *api.ValidateResult {
	if tp.Status == util.PodGroupRunning {
		klog.V(util.LogDebugLev).Infof("virtual310NPU %s's pg is running", tp.ComJob.Name)
		return nil
	}
	if reqErr := tp.checkDyVJobReq(); reqErr != nil && !reqErr.Pass {
		return reqErr
	}
	return tp.validDyVNPUJobLabel()
}
