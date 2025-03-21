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
Package diagcontext some test case for the metric.
*/
package diagcontext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetric_GetMetricKey(t *testing.T) {
	// 指标域的type1:指标域的name1-指标域的type2:指标域的name2_指标名
	metricKey := "domain_type_string:domain_item_1-domain_type_string:domain_item_2-metric_name"
	assert.Equal(t, metricKey, metric.GetMetricKey())
}
