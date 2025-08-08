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
Package diagcontext 包提供了与监控和诊断相关的上下文信息。
*/
package diagcontext

import (
	"ascend-faultdiag-online/pkg/core/context/contextdata"
	"ascend-faultdiag-online/pkg/utils/slicetool"
)

// Condition 表示一个诊断条件，包含数据和匹配函数。
type Condition struct {
	Data         any
	MatchingFunc func(ctxData *contextdata.CtxData, data any) bool
}

// IsMatching 检查当前条件是否与给定的数据匹配。
func (condition *Condition) IsMatching(ctxData *contextdata.CtxData) bool {
	if condition == nil {
		return false
	}
	return condition.MatchingFunc(ctxData, condition.Data)
}

// ConditionGroup 条件组
type ConditionGroup struct {
	StaticConditions  []*Condition // 静态条件，启动阶段过滤
	DynamicConditions []*Condition // 动态条件，每次诊断前判断
}

// IsStaticMatching 检查当前条件是否与给定的数据匹配。
func (group *ConditionGroup) IsStaticMatching(ctxData *contextdata.CtxData) bool {
	if group == nil {
		return false
	}
	if len(group.StaticConditions) == 0 {
		return true
	}
	return slicetool.All(group.StaticConditions, func(c *Condition) bool {
		return c.IsMatching(ctxData)
	})
}

// IsDynamicMatching 检查当前条件是否与给定的数据匹配。
func (group *ConditionGroup) IsDynamicMatching(ctxData *contextdata.CtxData) bool {
	if group == nil {
		return false
	}
	if len(group.DynamicConditions) == 0 {
		return true
	}
	return slicetool.All(group.DynamicConditions, func(c *Condition) bool {
		return c.IsMatching(ctxData)
	})
}
