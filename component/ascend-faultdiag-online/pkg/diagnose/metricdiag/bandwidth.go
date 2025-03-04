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
Package metricdiag 提供了一些诊断项的实现。
*/
package metricdiag

import (
	"ascend-faultdiag-online/pkg/context/diagcontext"
	"ascend-faultdiag-online/pkg/context/diagcontext/diagtemplate"
	"ascend-faultdiag-online/pkg/diagnose/condition"
	"ascend-faultdiag-online/pkg/model/diagmodel/metricmodel"
	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/utils/constants"
)

func getBandwidthThreshold910A2() *diagcontext.MetricThreshold {
	return &diagcontext.MetricThreshold{
		Name:  constants.MetricBandwidth,
		Unit:  "GB/s",
		Value: 300.0,
	}
}

// ascend910A2BandwidthDiagItem 提供开发样例，仅供参考，非实际诊断项。
func ascend910A2BandwidthDiagItem() *diagcontext.DiagItem {
	threshold910A2 := getBandwidthThreshold910A2()
	return &diagcontext.DiagItem{
		Name:        "910A2 bandwidth",
		Interval:    60,
		Description: "910A2带宽诊断",
		ConditionGroup: &diagcontext.ConditionGroup{
			StaticConditions: []*diagcontext.Condition{condition.Ascend910A2Condition()},
		},
		Rules: []*diagcontext.DiagRule{
			{
				QueryFunc: func(pool *diagcontext.MetricPool) []*diagcontext.DomainMetrics {
					return diagcontext.NewQueryBuilder(pool).
						QueryByDomainItem(&metricmodel.DomainItem{DomainType: enum.NpuDomain}).
						CollectDomainMetrics([]string{constants.MetricBandwidth})
				},
				Thresholds: []*diagcontext.MetricThreshold{threshold910A2},
				DiagFunc: diagtemplate.SingleFloat64MetricDiagFunc(threshold910A2,
					func(metric, threshold float64) *diagcontext.CompareRes {
						return &diagcontext.CompareRes{
							IsAbnormal:  metric < threshold,
							Description: "metric less than lower threshold",
						}
					},
				),
			},
		},
	}
}

func GetBandWidthDiagItems() []*diagcontext.DiagItem {
	return []*diagcontext.DiagItem{
		ascend910A2BandwidthDiagItem(),
	}
}
