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
	"sync"
	"time"
)

type MetricPoolItem struct {
	Name      string      // 指标名称
	Value     interface{} // 指标值（可以是任意类型）
	Timestamp time.Time   // 时间戳
}

type MetricPool struct {
	metrics map[string][]MetricPoolItem // 指标名称到指标项的映射
	mu      sync.RWMutex                // 读写锁，保证并发安全
}

// NewMetricPool 创建一个新的指标池
func NewMetricPool() *MetricPool {
	return &MetricPool{
		metrics: make(map[string][]MetricPoolItem),
	}
}

// AddMetric 添加指标项
func (p *MetricPool) AddMetric(name string, value interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	item := MetricPoolItem{
		Name:      name,
		Value:     value,
		Timestamp: time.Now(),
	}
	p.metrics[name] = append(p.metrics[name], item)
}

// GetLatestMetric 获取最新的指标项
func (p *MetricPool) GetLatestMetric(name string) (MetricPoolItem, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	items, ok := p.metrics[name]
	if !ok || len(items) == 0 {
		return MetricPoolItem{}, false
	}
	return items[len(items)-1], true
}

// GetMetricHistory 获取指标历史数据
func (p *MetricPool) GetMetricHistory(name string) ([]MetricPoolItem, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	items, ok := p.metrics[name]
	if !ok {
		return nil, false
	}
	return items, true
}
