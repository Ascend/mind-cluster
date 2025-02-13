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
Package condition 提供场景匹配条件。
*/
package condition

import (
	"ascend-faultdiag-online/pkg/context"
	"ascend-faultdiag-online/pkg/model/diagmodel"
	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/utils/slice"
)

func getChipTypeCondition(chipTypes []enum.ChipType) *diagmodel.Condition {
	return &diagmodel.Condition{
		Data: chipTypes,
		MatchingFunc: func(ctx *context.FaultDiagContext, i interface{}) (bool, error) {
			chipTypes, ok := i.([]enum.ChipType)
			if !ok {
				return false, nil
			}
			err := slice.ValueIn(ctx.NodeStatus.ChipType, chipTypes)
			if err != nil {
				return false, err
			}
			return true, nil
		},
	}
}

// Ascend910A2Condition 表示Ascend910A2芯片的诊断条件
func Ascend910A2Condition() *diagmodel.Condition {
	return getChipTypeCondition([]enum.ChipType{enum.Ascend910A2})
}
