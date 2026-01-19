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
	"ascend-faultdiag-online/pkg/core/api"
	"ascend-faultdiag-online/pkg/core/context/contextdata"
	"ascend-faultdiag-online/pkg/core/context/diagcontext"
	"ascend-faultdiag-online/pkg/core/model"
	"ascend-faultdiag-online/pkg/core/model/diagmodel/metricmodel"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/utils/slicetool"
)

const apiAddMetric = "add"

// GetAddMetricApi 获取添加指标的api
func GetAddMetricApi() *api.Api {
	return api.BuildApi(apiAddMetric, &metricmodel.MetricReqData{}, apiAddMetricFunc, nil)
}

func apiAddMetricFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext,
	reqCtx *model.RequestContext, model *metricmodel.MetricReqData) error {
	for _, metric := range model.Metrics {
		if slicetool.ValueIn(metric.ValueType, []enum.MetricValueType{enum.FloatMetric, enum.StringMetric}) != nil {
			ctxData.Framework.Logger.Println("Unknown Metric Type", metric.ValueType)
		}
		domain := diagCtx.DomainFactory.GetInstance(metric.Domain)
		diagCtx.MetricPool.AddMetric(&diagcontext.Metric{Domain: domain, Name: metric.Name}, metric.Value, metric.ValueType)
	}
	reqCtx.Response.Status = enum.Success
	reqCtx.Response.Msg = "add metric success"
	return nil
}
