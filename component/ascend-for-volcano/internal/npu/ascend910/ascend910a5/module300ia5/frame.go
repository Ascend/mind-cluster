/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package module300ia5 for 300I server scheduling
package module300ia5

import (
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
)

// New return npu plugin for ascend300IA5
func New(name string) base.AscendHandler {
	m := &ascend300IA5{}
	klog.V(util.LogInfoLev).Infof("ascend300IA5 card type =%s", name)
	num := getNPUNumByHandler(name)
	m.SetMaxNodeNPUNum(num)
	m.SetPluginName(name)
	m.SetAnnoName(util.NPU910CardName)
	m.SetAnnoPreVal(util.NPU910CardNamePre)
	m.affScoreList = createAffScoreList(m.MaxNodeNPUNum)
	return m
}
