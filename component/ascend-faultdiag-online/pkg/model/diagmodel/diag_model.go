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
Package diagmodel 提供诊断相关模型实体
*/
package diagmodel

import "ascend-faultdiag-online/pkg/context"

// Condition 表示一个诊断条件，包含数据和匹配函数。
type Condition struct {
	Data         interface{}
	MatchingFunc func(*context.FaultDiagContext, interface{}) (bool, error)
}

// IsMatching 检查当前条件是否与给定的数据匹配。
func (condition *Condition) IsMatching(ctx *context.FaultDiagContext) (bool, error) {
	return condition.MatchingFunc(ctx, condition.Data)
}
