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

// Package register provides some api for registration
package register

import (
	"ascend-faultdiag-online/pkg/core/api"
	"ascend-faultdiag-online/pkg/core/context"
	"ascend-faultdiag-online/pkg/register/router"
)

func getFeatureApi() *api.Api {
	return api.NewApi("feature", nil, []*api.Api{
		router.GetNetFaultApi(),
		router.GetSlowNodeApi(),
	})
}

// routes register the APIs to fd ctx
func routes(fdCtx *context.FaultDiagContext) {
	featureAPI := api.NewApi("", nil, []*api.Api{
		router.GetMetricApi(),
		router.GetEventApi(),
		getFeatureApi(),
	})
	fdCtx.RegisterRouter(featureAPI)
}
