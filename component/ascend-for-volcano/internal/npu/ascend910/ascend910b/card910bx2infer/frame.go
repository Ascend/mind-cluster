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
Package card910bx2infer is using for HuaWei Ascend 910B(Atlas 300T A2) card pin affinity schedule.
*/
package card910bx2infer

import (
	"fmt"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/base"
)

// New return npu plugin
func New(name string) base.AscendHandler {
	m := &card910bx2infer{}
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.SetMaxNodeNPUNum(nodeNPUNumber)
	m.SetAcceleratorValue(util.JobKind910BValue)
	return m
}

// ValidNPUJob check job req npu num and mode
func (tp *card910bx2infer) ValidNPUJob() *api.ValidateResult {
	if tp.NPUTaskNum != 1 {
		klog.V(util.LogErrorLev).Infof("GetVTaskNumInVJob %s has %d npu tasks, only support 1.", tp.Name, tp.NPUTaskNum)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  "ValidNPUJob failed",
			Message: fmt.Sprintf("%s task num is invalid", tp.Name),
		}
	}

	return tp.Valid910bNPUJob()
}
