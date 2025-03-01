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
Package ascend910b is using for HuaWei Ascend 910B pin affinity schedule.
*/
package ascend910b

import (
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// Valid910bNPUJob check the 910b job req npu num and mode
func (ab *Base910b) Valid910bNPUJob() *api.ValidateResult {
	vResult := &api.ValidateResult{}
	var vErr error = nil
	defer func() {
		if vErr != nil {
			vResult.Pass = false
			vResult.Reason = vErr.Error()
			vResult.Message = vErr.Error()
			return
		}
	}()

	// 1. check parameter.
	if ab == nil {
		vErr = fmt.Errorf("nil plugin %s", ab.GetPluginName())
		klog.V(util.LogErrorLev).Infof("ValidNPUJob err: %s.", vErr)
		return vResult
	}

	// 2.check ring-controller.atlas
	if vErr = ab.CheckJobForm(); vErr != nil {
		klog.V(util.LogErrorLev).Infof("checkJobForm: %s.", vErr)
		return vResult
	}

	return ab.NPUHandler.ValidNPUJob()
}

// CheckJobForm to check job ring-controller.atlas for future unification.
func (ab *Base910b) CheckJobForm() error {
	// for vcJob and deployment.
	lValue, ok := ab.Label[util.JobKindKey]
	if !ok {
		return fmt.Errorf("%s not has no label:%s", ab.Name, util.JobKindKey)
	}

	if lValue != ab.GetAcceleratorValue() {
		return fmt.Errorf("%s label:%s not right(%s)", ab.Name, lValue, ab.GetAcceleratorValue())
	}
	return nil
}
