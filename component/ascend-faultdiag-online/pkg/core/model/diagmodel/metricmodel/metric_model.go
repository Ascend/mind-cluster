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

// Package metricmodel defines the data models for the metric domain.
package metricmodel

import (
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/utils/constants"
)

// DomainItem 指标域单项
type DomainItem struct {
	DomainType enum.MetricDomainType `json:"domain_type"`
	Value      string                `json:"value"`
}

// GetDomainItemKey get the key of DomainItem
func (item *DomainItem) GetDomainItemKey() string {
	if item == nil {
		return ""
	}
	return string(item.DomainType) + constants.ValueSeparator + item.Value
}

// MetricReqModel 指标请求数据模型
type MetricReqModel struct {
	Domain    []*DomainItem        `json:"domain"`
	Name      string               `json:"name"`
	ValueType enum.MetricValueType `json:"value_type"`
	Value     string               `json:"value"`
}

// MetricReqData 指标请求data
type MetricReqData struct {
	Metrics []*MetricReqModel `json:"metrics"`
}
