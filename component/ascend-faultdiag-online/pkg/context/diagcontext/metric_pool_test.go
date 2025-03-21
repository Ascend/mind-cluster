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
Package diagcontext some test case for the metric pool.
*/
package diagcontext

import (
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/model/diagmodel/metricmodel"
	"ascend-faultdiag-online/pkg/utils/constants"
)

var (
	// 指标域单项
	domainItemOne = &metricmodel.DomainItem{
		DomainType: "domain_type_string",
		Value:      "domain_item_1",
	}
	domainItems = []*metricmodel.DomainItem{
		domainItemOne,
		{
			DomainType: "domain_type_string",
			Value:      "domain_item_2",
		},
	}
	// DomainItems，多个指标域
	domain = &Domain{DomainItems: domainItems}
	// 多个指标项（包括指标域和指标名）
	metric = &Metric{
		Domain: domain,
		Name:   "metric_name",
	}
	// 1、指标group（指标域和指标名 + 指标值和时间）
	itemGroup = NewMetricPoolItemGroup(metric)
	// 具体的指标项（指标值和时间）
	itemFirst = &Item{
		Value:     "item_value_first",
		Timestamp: time.Now(),
	}
	itemLast = &Item{
		Value:     "item_value_last",
		Timestamp: time.Now(),
	}
	// 2、指标树
	parentNode = &TreeNode{}
	treeNode   = NewMetricPoolTreeNode(domainItemOne, parentNode)
	// 3、指标池（指标group + 指标树）
	metricPool = NewMetricPool()
)

func TestNewMetricPoolItemGroup(t *testing.T) {
	assert.NotNil(t, itemGroup)
}

func TestAdd(t *testing.T) {
	itemGroup.Add(itemFirst)
	for i := 1; i < maxMetricRecordSize; i++ {
		itemTemp := &Item{
			Value:     "item_value" + strconv.Itoa(i),
			Timestamp: time.Now(),
		}
		itemGroup.Add(itemTemp)
	}
	assert.Equal(t, len(itemGroup.Items), maxMetricRecordSize)

	itemGroup.Add(itemLast)
	assert.Equal(t, len(itemGroup.Items), maxMetricRecordSize)
	if slices.Contains(itemGroup.Items, itemFirst) {
		assert.Fail(t, "如果超过最大记录数，则移除最旧的一个")
	}
}

func TestGetLatestMetricPoolItem(t *testing.T) {
	itemGroup.Add(itemFirst)
	itemGroup.Add(itemLast)
	assert.Equal(t, itemLast, itemGroup.GetLatestMetricPoolItem(), "获取到最新的指标项")
}

func TestGetItemGroup(t *testing.T) {
	// treeNode.MetricMap中不存在metric，添加并返回
	_, notExit := treeNode.MetricMap[metric.Name]
	assert.False(t, notExit)
	itemGroupGet := treeNode.GetItemGroup(metric)
	itemGroupInMap, ok := treeNode.MetricMap[metric.Name]
	assert.True(t, ok)
	// treeNode.MetricMap中存在metric，直接返回
	assert.Equal(t, itemGroupInMap, itemGroupGet)
}

func TestNewMetricPoolTreeNode(t *testing.T) {
	assert.NotNil(t, treeNode)
}

func TestNewMetricPool(t *testing.T) {
	assert.NotNil(t, metricPool)
}

func TestAddMetric(t *testing.T) {
	// metricPool.metricMap中不存在，则添加key:ItemGroup，并在该ItemGroup中添加Item
	key := metric.GetMetricKey() // 指标域的type1:指标域的name1-指标域的type2:指标域的name2_指标名
	itemValue := "item_value"

	_, notExit := metricPool.metricMap[key]
	assert.False(t, notExit, "未添加时不存在")

	metricPool.AddMetric(metric, itemValue)

	group, exit := metricPool.metricMap[key]
	assert.True(t, exit, "添加后，指标名称到指标项的映射中含有该指标group")
	assert.Equal(t, group.Items[0].Value, itemValue)
	// todo验证addToMetricTree
}

func TestGetMetricByMetricKey(t *testing.T) {
	groupsNil := metricPool.GetMetricByMetricKey(metric)
	assert.Nil(t, groupsNil, "未查找到返回nil")

	metricPool.AddMetric(metric, "item_value")
	groups := metricPool.GetMetricByMetricKey(metric)
	// 精确查找最新的指标项，返回切片，统一查询类接口返回切片。
	key := metric.GetMetricKey()
	group, exit := metricPool.metricMap[key]
	assert.True(t, exit, "添加后，精确查找到指标项")
	assert.Equal(t, groups[0], group)
}

func TestGetDomainMetrics(t *testing.T) {
	resultsNil := metricPool.GetDomainMetrics(domain)
	assert.Nil(t, resultsNil, "添加前，未查找到返回[]*ItemGroup的零值nil")

	metricPool.AddMetric(metric, "item_value")
	results := metricPool.GetDomainMetrics(domain)
	domainKey := domain.GetDomainKey()
	key := domainKey + constants.TypeSeparator + metric.Name
	group, exit := metricPool.metricMap[key]
	assert.True(t, exit, "添加后，根据指标域精确查找到数据")
	assert.Equal(t, results[0], group)
}
