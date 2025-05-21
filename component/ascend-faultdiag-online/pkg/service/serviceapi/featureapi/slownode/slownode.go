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
Package slownode provides API
*/
package slownode

import (
	"ascend-faultdiag-online/pkg/service/servicecore"
)

const (
	slownodeApi    = "slownode"
	slowClusterApi = "cluster"
	slowNodetApi   = "node"
)

// GetSlowNodeApi 获取指标相关api
func GetSlowNodeApi() *servicecore.Api {
	return servicecore.NewApi(slownodeApi, nil, []*servicecore.Api{
		getSlownodeClusterApi(),
		getSlownodeApi(),
	})
}

func getSlownodeClusterApi() *servicecore.Api {
	return servicecore.NewApi(slowClusterApi, nil, []*servicecore.Api{
		ClusterStart(),
		ClusterStop(),
		ClusterReload(),
	})
}

func getSlownodeApi() *servicecore.Api {
	return servicecore.NewApi(slowNodetApi, nil, []*servicecore.Api{
		NodeStart(),
		NodeStop(),
		NodeReload(),
	})
}
