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
Package diagcontext 包提供了指标池相关的能力。
*/
package diagcontext

import (
	"strings"
	"sync"
	"time"

	"ascend-faultdiag-online/pkg/model/diagmodel/metricmodel"
	"ascend-faultdiag-online/pkg/model/enum"
)

// Item 表示一个具体的指标项。
type Item struct {
	Value     interface{} // 指标值
	Timestamp time.Time   // 时间戳
}

// maxMetricRecordSize 表示一个指标最多记录的数量。
const maxMetricRecordSize = 10

// ItemGroup 表示指标项，留存10条历史记录。
type ItemGroup struct {
	Metric *Metric      // 指标项
	Items  []*Item      // 指标历史记录
	mu     sync.RWMutex // 读写锁，保证并发安全
}

// NewMetricPoolItemGroup 创建一个新的 MetricPoolItemGroup 实例。
func NewMetricPoolItemGroup(metric *Metric) *ItemGroup {
	return &ItemGroup{
		Metric: metric,
		Items:  make([]*Item, 0),
		mu:     sync.RWMutex{},
	}
}

// Add 添加一个新的指标项。如果超过最大记录数，则移除最旧的一个。
func (group *ItemGroup) Add(item *Item) {
	group.mu.Lock()
	defer group.mu.Unlock()
	if len(group.Items) >= maxMetricRecordSize {
		group.Items = append(group.Items[1:], item)
	} else {
		group.Items = append(group.Items, item)
	}
}

// GetLatestMetricPoolItem 获取最新的指标项。
func (group *ItemGroup) GetLatestMetricPoolItem() *Item {
	group.mu.RLock()
	defer group.mu.RUnlock()
	return group.Items[len(group.Items)-1]
}

// TreeNode 表示一个指标树
type TreeNode struct {
	DomainType       enum.MetricDomainType
	DomainValue      string
	ParentNode       *TreeNode                             // 父节点
	ChildrenNodesMap map[enum.MetricDomainType][]*TreeNode // 子节点
	MetricMap        map[string]*ItemGroup                 // 指标名称到指标项的映射
}

// NewMetricPoolTreeNode 新建指标池的指标树
func NewMetricPoolTreeNode(domainItem *metricmodel.DomainItem, parentNode *TreeNode) *TreeNode {
	return &TreeNode{
		DomainType:       domainItem.DomainType,
		DomainValue:      domainItem.Value,
		ParentNode:       parentNode,
		ChildrenNodesMap: make(map[enum.MetricDomainType][]*TreeNode),
		MetricMap:        make(map[string]*ItemGroup),
	}
}

// MetricPool 表示一个指标池，用于存储和管理多个指标项。
type MetricPool struct {
	metricMap        map[string]*ItemGroup // 指标名称到指标项的映射
	poolRootNodesMap map[enum.MetricDomainType][]*TreeNode
	mu               sync.RWMutex // 读写锁，保证并发安全
}

// NewMetricPool 创建一个新的指标池
func NewMetricPool() *MetricPool {
	return &MetricPool{
		metricMap: make(map[string]*ItemGroup),
	}
}

// AddMetric 添加指标项
func (p *MetricPool) AddMetric(metric *Metric, value interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	poolItem := &Item{
		Value:     value,
		Timestamp: time.Now(),
	}
	p.addToMetricMap(metric, poolItem)
	p.addToMetricTree(metric, poolItem)
}

// 添加到指标字典
func (p *MetricPool) addToMetricMap(metric *Metric, poolItem *Item) {
	key := metric.GetMetricKey()

	if _, ok := p.metricMap[key]; !ok {
		p.metricMap[key] = NewMetricPoolItemGroup(metric)
	}
	p.metricMap[key].Add(poolItem)
}

// 添加到指标树
func (p *MetricPool) addToMetricTree(metric *Metric, poolItem *Item) {
	var curNodesMap map[enum.MetricDomainType][]*TreeNode
	var lastNode *TreeNode
	curNodesMap = p.poolRootNodesMap
	for _, domainItem := range metric.Domain.DomainItems {
		var node *TreeNode
		nodes, ok := curNodesMap[domainItem.DomainType]
		if ok {
			node = exitNode(nodes, domainItem)
		} else {
			curNodesMap[domainItem.DomainType] = make([]*TreeNode, 0)
		}
		if node == nil {
			node = NewMetricPoolTreeNode(domainItem, lastNode)
		}
		curNodesMap[domainItem.DomainType] = append(curNodesMap[domainItem.DomainType], node)
		curNodesMap = node.ChildrenNodesMap
		lastNode = node
	}
	lastNode.MetricMap[metric.Name].Add(poolItem)
}

// 获取指标树
func exitNode(nodes []*TreeNode, domainItem *metricmodel.DomainItem) *TreeNode {
	for _, treeNode := range nodes {
		if treeNode.DomainValue == domainItem.Value {
			return treeNode
		}
	}
	return nil
}

// GetMetricByMetricKey 精确查找最新的指标项
func (p *MetricPool) GetMetricByMetricKey(metric *Metric) []*ItemGroup {
	p.mu.RLock()
	defer p.mu.RUnlock()
	group, ok := p.metricMap[metric.GetMetricKey()]
	if !ok {
		return nil
	}
	return []*ItemGroup{group}
}

// GetDomainMetrics 根据指标域精确查找数据
func (p *MetricPool) GetDomainMetrics(domain *Domain) []*ItemGroup {
	p.mu.RLock()
	defer p.mu.RUnlock()
	domainKey := domain.GetDomainKey()
	var results []*ItemGroup
	for key, group := range p.metricMap {
		if strings.HasPrefix(key, domainKey) {
			results = append(results, group)
		}
	}
	return results
}
