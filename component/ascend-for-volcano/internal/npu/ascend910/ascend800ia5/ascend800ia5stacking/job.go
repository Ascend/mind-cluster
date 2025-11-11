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
