package diagcontext

import (
	"ascend-faultdiag-online/pkg/context"
	"ascend-faultdiag-online/pkg/utils/slicetool"
)

// Condition 表示一个诊断条件，包含数据和匹配函数。
type Condition struct {
	Data         interface{}
	MatchingFunc func(ctx *context.FaultDiagContext, data interface{}) bool
}

// IsMatching 检查当前条件是否与给定的数据匹配。
func (condition *Condition) IsMatching(ctx *context.FaultDiagContext) bool {
	return condition.MatchingFunc(ctx, condition.Data)
}

// ConditionGroup 条件组
type ConditionGroup struct {
	StaticConditions  []*Condition // 静态条件，启动阶段过滤
	DynamicConditions []*Condition // 动态条件，每次诊断前判断
}

// IsStaticMatching 检查当前条件是否与给定的数据匹配。
func (group *ConditionGroup) IsStaticMatching(ctx *context.FaultDiagContext) bool {
	if len(group.StaticConditions) == 0 {
		return true
	}
	return slicetool.All(group.StaticConditions, func(c *Condition) bool {
		return c.IsMatching(ctx)
	})
}

// IsDynamicMatching 检查当前条件是否与给定的数据匹配。
func (group *ConditionGroup) IsDynamicMatching(ctx *context.FaultDiagContext) bool {
	if len(group.DynamicConditions) == 0 {
		return true
	}
	return slicetool.All(group.DynamicConditions, func(c *Condition) bool {
		return c.IsMatching(ctx)
	})
}
