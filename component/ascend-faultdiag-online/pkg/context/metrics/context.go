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
Package metrics 包提供了与监控和诊断相关的功能。
*/
package metrics

import (
	"ascend-faultdiag-online/pkg/context"
	"ascend-faultdiag-online/pkg/model/diagmodel"
	"ascend-faultdiag-online/pkg/utils/slice"
)

// DiagRule 是一个诊断规则的结构体
type DiagRule struct {
	MetricName  string                               // 监控指标名称
	Threshold   float64                              // 阈值
	Unit        string                               // 单位
	CompareFunc func(float64, float64) (bool, error) // 当前值比较函数
}

// IsMatching 方法用于判断给定的指标值是否匹配诊断规则
func (rule *DiagRule) IsMatching(metricValue float64) (bool, error) {
	if res, err := rule.CompareFunc(metricValue, rule.Threshold); err != nil {
		return false, err
	} else {
		return res, nil
	}
}

// DiagItem 结构体用于表示一个诊断项
type DiagItem struct {
	Name            string                                                                   // 名称
	Interval        int                                                                      // 检查间隔时间，单位为秒
	Rules           []*DiagRule                                                              // 诊断规则
	CustomRules     []func(ctx *context.FaultDiagContext, metricValue float64) (bool, error) // 自定义诊断规则
	ConditionGroups [][]*diagmodel.Condition                                                 // 诊断触发条件
	Description     string                                                                   // 描述信息
	DiagFlag        chan bool                                                                // 启用诊断标志
}

// Context 是一个诊断内容的结构体
type Context struct {
	MetricPool *MetricPool // 指标池
	DiagItems  []*DiagItem // 诊断项
}

// NewMetricContext 创建一个新的 Context 实例，并初始化 MetricPool 和 DiagItems 字段
func NewMetricContext() *Context {
	return &Context{
		MetricPool: NewMetricPool(),
		DiagItems:  make([]*DiagItem, 0),
	}
}

// UpdateDiagItems 更新上下文中的诊断项列表，将新的诊断项添加到现有列表中
func (ctx *Context) UpdateDiagItems(diagItems []*DiagItem) {
	ctx.DiagItems = slice.Extend(ctx.DiagItems, diagItems)
}
