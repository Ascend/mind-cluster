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
	"ascend-faultdiag-online/pkg/context/contextdata"
	"ascend-faultdiag-online/pkg/context/diagcontext"
	"ascend-faultdiag-online/pkg/model/diagmodel/metricmodel"
	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/service/request"
	"ascend-faultdiag-online/pkg/service/servicecore"
	"ascend-faultdiag-online/pkg/utils"
)

const apiAddMetric = "add"

// GetAddMetricApi 获取添加指标的api
func GetAddMetricApi() *servicecore.Api {
	return servicecore.BuildApi(apiAddMetric, &metricmodel.MetricReqData{}, apiAddMetricFunc, nil)
}

func apiAddMetricFunc(_ *contextdata.CtxData, diagCtx *diagcontext.DiagContext, reqCtx *request.Context, model *metricmodel.MetricReqData) error {
	for _, metric := range model.Metrics {
		var metricValue interface{}
		switch metric.ValueType {
		case enum.FloatMetric:
			metricValue = utils.ToFloat64(metric.Value, 0)
		case enum.StringMetric:
			metricValue = utils.ToString(metric.Value)
		}
		domain := diagCtx.DomainFactory.GetInstance(metric.Domain)
		diagCtx.MetricPool.AddMetric(&diagcontext.Metric{Domain: domain, Name: metric.Name}, metricValue)
	}
	reqCtx.Response.Status = enum.Success
	reqCtx.Response.Msg = "add metric success"
	return nil
}
