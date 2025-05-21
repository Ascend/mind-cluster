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
	"ascend-faultdiag-online/pkg/context/contextdata"
)

// DiagContext 是一个诊断上下文的结构体
type DiagContext struct {
	MetricPool      *MetricPool            // 指标池
	DiagItemMap     map[string]*DiagItem   // 诊断项字典
	DiagRecordStore *DiagRecordStore       // 当前诊断结果记录
	DomainFactory   *DomainFactory         // 指标域工厂类
	tickerMap       map[string]*DiagTicker // 诊断周期任务字典
}

// NewDiagContext 创建一个新的 DiagContext 实例，并初始化 MetricPool 和 DiagItemMap 字段
func NewDiagContext() *DiagContext {
	return &DiagContext{
		MetricPool:      NewMetricPool(),
		DiagItemMap:     make(map[string]*DiagItem),
		DiagRecordStore: NewDiagRecordStore(),
		DomainFactory:   NewDomainFactory(),
		tickerMap:       make(map[string]*DiagTicker),
	}
}

// UpdateDiagItems 更新上下文中的诊断项列表，将新的诊断项添加到现有列表中
func (ctx *DiagContext) UpdateDiagItems(diagItems []*DiagItem) []*DiagTicker {
	tickers := make([]*DiagTicker, 0)
	if ctx == nil {
		return tickers
	}
	for _, item := range diagItems {
		ctx.DiagItemMap[item.Name] = item
		ticker := NewDiagTicker(item)
		ctx.tickerMap[item.Name] = ticker
		tickers = append(tickers, ticker)
	}
	return tickers
}

// StartDiag 开始诊断
func (ctx *DiagContext) StartDiag(ctxData *contextdata.CtxData) {
	if ctx != nil {
		for _, ticker := range ctx.tickerMap {
			ticker.Start(ctxData, ctx)
		}
	}

}

// CloseDiagItem 关闭
func (ctx *DiagContext) CloseDiagItem(itemName string) {
	if ctx == nil {
		return
	}
	ticker, ok := ctx.tickerMap[itemName]
	if !ok {
		return
	}
	close(ticker.StopChan)
	delete(ctx.tickerMap, itemName)
}
