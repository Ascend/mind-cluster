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
Package metricapi provides API
*/
package metricapi

import (
	"ascend-faultdiag-online/pkg/service/serviceapi"
)

const apiMetric = "metric"

// GetMetricApi 获取指标相关api
func GetMetricApi() *serviceapi.Api {
	return serviceapi.NewApi(apiMetric, nil, []*serviceapi.Api{
		GetAddMetricApi(),
	})
}
