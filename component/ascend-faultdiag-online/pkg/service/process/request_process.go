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
Package process 提供事务处理
*/
package process

import (
	"ascend-faultdiag-online/pkg/context"
	"ascend-faultdiag-online/pkg/service/request"
)

const (
	// EventType 事件类型：event
	EventType = "event"
	// MetricType 指标类型：metricdiag
	MetricType = "metricdiag"
)

// RequestProcess 处理请求的函数，根据请求类型调用不同的处理函数
func RequestProcess(ctx *context.FaultDiagContext, reqCtx *request.Context) error {
	switch reqCtx.Request.RequestType {
	case EventType:
		return EventProcess(ctx, reqCtx)
	case MetricType:
		return MetricProcess(ctx, reqCtx)
	default:
		return nil
	}
}
