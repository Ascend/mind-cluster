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
Package metricpool 包提供了指标池相关的能力。
*/
package metricpool

import (
	"ascend-faultdiag-online/pkg/context/diagcontext"
	"ascend-faultdiag-online/pkg/model/diagmodel/metricmodel"
	"ascend-faultdiag-online/pkg/utils/slicetool"
)

// DomainMetrics 在同一个域下面不同指标的汇聚结构
type DomainMetrics struct {
	Domain       *diagcontext.Domain
	ItemGroupMap map[string]*ItemGroup // 指标字典，key为指标名，value为指标值结构
}

type MetricQueryBuilder struct {
	tempTreeNodes []*TreeNode
	Found         bool
}

func NewQueryBuilder(metricPool *MetricPool) *MetricQueryBuilder {
	return &MetricQueryBuilder{tempTreeNodes: []*TreeNode{{
		ChildrenNodesMap: metricPool.poolRootNodesMap},
	}}
}

// QueryByDomainItem 根据单个域信息查找下一级域
func (p *MetricQueryBuilder) QueryByDomainItem(domainItem *metricmodel.DomainItem) *MetricQueryBuilder {
	if !p.Found {
		return p
	}
	nextLevelTreeNodes := make([]*TreeNode, len(p.tempTreeNodes))
	for _, curNode := range p.tempTreeNodes {
		nextLevelNodesPart, ok := curNode.ChildrenNodesMap[domainItem.DomainType]
		if !ok {
			continue
		}
		// 域值不为空时，则进行比对，保留相同的选项
		if len(domainItem.Value) != 0 {
			nextLevelNodesPart = slicetool.Filter(nextLevelNodesPart, func(node *TreeNode) bool {
				return node.DomainValue == domainItem.Value
			})
		}
		nextLevelTreeNodes = append(nextLevelTreeNodes, nextLevelNodesPart...)
	}
	if len(nextLevelTreeNodes) == 0 {
		p.Found = false
	}
	p.tempTreeNodes = nextLevelTreeNodes
	return p
}

// QueryByDomain 根据整个域查找
func (p *MetricQueryBuilder) QueryByDomain(domain *diagcontext.Domain) *MetricQueryBuilder {
	for _, item := range domain.DomainItems {
		p.QueryByDomainItem(item)
	}
	return p
}

// CollectDomainMetrics 收集域列表下同个域多个指标的结果
func (p *MetricQueryBuilder) CollectDomainMetrics(metricNames []string) []*DomainMetrics {
	results := make([]*DomainMetrics, 0)
	for _, node := range p.tempTreeNodes {
		var domain *diagcontext.Domain
		itemGroupMap := make(map[string]*ItemGroup)
		for _, name := range metricNames {
			metricItem, ok := node.MetricMap[name]
			if !ok {
				continue
			}
			itemGroupMap[name] = metricItem
			domain = metricItem.Metric.Domain
		}
		if len(itemGroupMap) == 0 {
			continue
		}
		metrics := &DomainMetrics{
			Domain:       domain,
			ItemGroupMap: itemGroupMap,
		}
		results = append(results, metrics)
	}
	return results
}
