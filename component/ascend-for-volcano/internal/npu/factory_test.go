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
Package npu is using for HuaWei Ascend pin affinity schedule.
*/
package npu

import (
	"fmt"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

func TestInit910CardPolicyHandler(t *testing.T) {
	configs := []string{
		util.Chip4Node8,
		util.Chip1Node2,
		util.Chip4Node4,
		util.Chip8Node8,
		util.Chip8Node16,
		util.Chip2Node16,
		util.Chip2Node16Sp,
	}

	for _, config := range configs {
		name := fmt.Sprintf("When schedule policy is %s then handleName is %s",
			config, policy910HandlerMap[config])
		t.Run(name, func(t *testing.T) {
			attr := util.SchedulerJobAttr{
				ComJob: util.ComJob{
					Annotation: map[string]string{
						util.SchedulePolicyAnnoKey: config,
					},
				},
			}
			handlerName := get910CardHandlerName(attr)
			if handlerName != policy910HandlerMap[config] {
				t.Errorf("Expect handler name to be %s, got %s", policy910HandlerMap[config], handlerName)
			}
		})
	}
}
