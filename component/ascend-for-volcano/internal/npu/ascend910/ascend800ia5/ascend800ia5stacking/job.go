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
Package ascend800ia5stacking is using for HuaWei Ascend800ia5x8stacking pin affinity schedule.
*/
package ascend800ia5stacking

import (
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// JudgeNodeAndTaskNPU determine if the number of NPUs required by the task can be satisfied
func (tp *module800ia5stacking) JudgeNodeAndTaskNPU(taskNPU int, nodeTop []int) error {
	if taskNPU <= len(nodeTop) {
		return nil
	}
	meetErr := fmt.Errorf("%v not meet req npu(%d)", nodeTop, taskNPU)
	klog.V(util.LogErrorLev).Infof("cardIDs:<%v> not meet task reqNum<%d>.", nodeTop, taskNPU)
	return meetErr
}

func (ab *module800ia5stacking) Valid800ia5NPUJob() *api.ValidateResult {
	vResult := &api.ValidateResult{}
	var vErr error
	defer func() {
		if vErr != nil {
			vResult.Pass = false
			vResult.Reason = vErr.Error()
			vResult.Message = vErr.Error()
			return
		}
	}()

	// check parameter.
	if ab == nil {
		vErr = fmt.Errorf("nil plugin %s", ab.GetPluginName())
		klog.V(util.LogErrorLev).Infof("ValidNPUJob err: %s.", vErr)
		return vResult
	}

	if ab.SpBlockNPUNum != 0 {
		klog.V(util.LogWarningLev).Infof("There is no need to set sp-block in standard cluster server.")
	}
	if ab.TpBlockNPUNum != util.LeastTpBlock {
		klog.V(util.LogWarningLev).Infof("There is no need to set tp-block in standard cluster server.")
	}

	// check job mode:distribute and single.
	if vErr = ab.checkJobMode(); vErr != nil {
		klog.V(util.LogErrorLev).Infof("checkJobTrainMode: %s.", vErr)
		return vResult
	}

	return nil
}

func (ab *module800ia5stacking) checkJobMode() error {
	if ab.NPUTaskNum == 0 {
		klog.V(util.LogErrorLev).Infof("GetVTaskNumInVJob %s has no npu tasks.", ab.Name)
		return fmt.Errorf("%s no npu job", ab.Name)
	}
	klog.V(util.LogDebugLev).Infof("checkJobMode job(%s) has %d tasks.", ab.Name, len(ab.Tasks))
	nTaskReqNpuNum := ab.ReqNPUNum / ab.NPUTaskNum
	if ab.CheckJobAllowNum(nTaskReqNpuNum) {
		return nil
	}
	return fmt.Errorf("%s checkJobMode %s req npu is invalid", ab.GetPluginName(), ab.Name)
}

// CheckJobAllowNum check the single job require is valid.
func (ab *module800ia5stacking) CheckJobAllowNum(value int) bool {
	_, ok := ab.NpuNumInvalidMap[value]
	return !ok && value <= ab.MaxNodeNPUNum
}
