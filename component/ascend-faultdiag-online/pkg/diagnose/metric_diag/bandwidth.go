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
Package metric_diag 提供了一些诊断项的实现。
*/
package metric_diag

import (
	"ascend-faultdiag-online/pkg/context/metrics"
	"ascend-faultdiag-online/pkg/diagnose/condition"
	"ascend-faultdiag-online/pkg/model/diag_model"
)

// Ascend910A2BandwidthDiagItem 提供开发样例，仅供参考，非实际诊断项。
var Ascend910A2BandwidthDiagItem = &metrics.DiagItem{
	Interval:        60,
	ConditionGroups: [][]*diag_model.Condition{{condition.Ascend910A2}},
	Description:     "910A2带宽诊断",
	Rules: []*metrics.DiagRule{
		{
			MetricName: "bandwidth",
			Threshold:  1.0,
			Unit:       "Gb/s",
			CompareFunc: func(metric float64, threshold float64) (bool, error) {
				return metric > threshold, nil
			},
		},
	},
}
