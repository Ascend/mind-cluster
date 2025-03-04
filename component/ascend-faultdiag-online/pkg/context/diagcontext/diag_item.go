package diagcontext

import (
	"time"

	"ascend-faultdiag-online/pkg/context/contextdata"
	"ascend-faultdiag-online/pkg/utils/slicetool"
)

// DiagFunc 诊断函数
type DiagFunc func(diagItem *DiagItem, thresholds []*MetricThreshold, domainMetrics []*DomainMetrics) []*MetricDiagRes

// CustomRuleFunc 自定义规则函数
type CustomRuleFunc func(ctxData *contextdata.CtxData, item *DiagItem) []*MetricDiagRes

// MetricPoolQueryFunc 指标池查找规则
type MetricPoolQueryFunc func(pool *MetricPool) []*DomainMetrics

// MetricCompareFunc 指标比较函数
type MetricCompareFunc func(metric, threshold interface{}) *CompareRes

// CustomRule 自定义诊断规则结构体
type CustomRule struct {
	CustomRuleFunc CustomRuleFunc // 自定义规则函数
	Description    string         // 描述
}

// MetricThreshold 指标预置结构
type MetricThreshold struct {
	Name  string
	Value interface{}
	Unit  string // 单位
}

// DiagRule 是一个诊断规则的结构体
type DiagRule struct {
	QueryFunc   MetricPoolQueryFunc //查找规则
	DiagFunc    DiagFunc            // 诊断函数
	Thresholds  []*MetricThreshold  // 阈值列表
	Description string              //描述
}

// Diag 方法用于判断给定的指标值是否匹配诊断规则
func (rule *DiagRule) Diag(diagItem *DiagItem, pool *MetricPool) []*MetricDiagRes {
	domainMetrics := rule.QueryFunc(pool)
	return rule.DiagFunc(diagItem, rule.Thresholds, domainMetrics)
}

// MetricDiagRes 诊断结果结构体
type MetricDiagRes struct {
	Metric      *Metric     // 指标项
	Value       interface{} // 指标值
	Threshold   interface{} // 阈值
	Time        time.Time   // 时间戳
	Unit        string      // 单位
	IsAbnormal  bool        //是否异常
	Description string      // 诊断规则描述
}

// CompareRes 比较结果
type CompareRes struct {
	IsAbnormal  bool   //是否异常
	Description string //描述
}

// DiagItem 结构体用于表示一个诊断项
type DiagItem struct {
	Name           string          // 名称
	Interval       int             // 检查间隔时间，单位为秒
	Rules          []*DiagRule     // 诊断规则
	CustomRules    []*CustomRule   // 自定义诊断规则
	ConditionGroup *ConditionGroup // 诊断触发条件
	Description    string          // 描述信息
}

// Diag 方法用于执行诊断逻辑
func (d *DiagItem) Diag(ctxData *contextdata.CtxData, diagCtx *DiagContext) []*MetricDiagRes {
	matching := d.ConditionGroup.IsDynamicMatching(ctxData)
	if !matching {
		return nil
	}
	pool := diagCtx.MetricPool
	return append(d.ruleDiag(pool), d.customRulesDiag(ctxData)...)
}

// ruleDiag 构建诊断结果
func (d *DiagItem) ruleDiag(pool *MetricPool) []*MetricDiagRes {
	if len(d.Rules) == 0 {
		return nil
	}
	results := slicetool.MapToValue(d.Rules, func(rule *DiagRule) []*MetricDiagRes {
		return rule.Diag(d, pool)
	})
	return slicetool.Chain(results)
}

// customRulesDiag 自定义诊断规则匹配
func (d *DiagItem) customRulesDiag(ctxData *contextdata.CtxData) []*MetricDiagRes {
	if len(d.CustomRules) == 0 {
		return nil
	}
	resLists := slicetool.MapToValue(d.CustomRules, func(rule *CustomRule) []*MetricDiagRes {
		return rule.CustomRuleFunc(ctxData, d)
	})
	return slicetool.Chain(resLists)
}
