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

// Package diagcontext defines some structs relevant the metric
package diagcontext

import (
	"ascend-faultdiag-online/pkg/utils/constants"
)

// Metric 指标结构体, 抽象的指标，包含指标域和指标名，不包含具体的值
type Metric struct {
	Domain *Domain // 指标域
	Name   string  // 指标名
}

// GetMetricKey get the key of Metric
func (item *Metric) GetMetricKey() string {
	if item == nil {
		return ""
	}
	return item.Domain.GetDomainKey() + constants.TypeSeparator + item.Name
}
