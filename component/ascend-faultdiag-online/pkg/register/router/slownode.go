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

// Package router provides some api relevant to slownode
package router

import (
	"ascend-faultdiag-online/pkg/core/api"
	"ascend-faultdiag-online/pkg/service/serviceapi/slownodeapi"
)

const (
	slowClusterApi = "cluster"
	slowNodetApi   = "node"
)

// getSlowNodeApi 获取指标相关api
func GetSlowNodeApi() *api.Api {
	return api.NewApi("slownode", nil, []*api.Api{
		getClusterSlowNodeApi(),
		getNodeSlowNodeApi(),
	})
}

func getClusterSlowNodeApi() *api.Api {
	return api.NewApi(slowClusterApi, nil, []*api.Api{
		slownodeapi.ClusterStart(),
		slownodeapi.ClusterStop(),
		slownodeapi.ClusterReload(),
	})
}

func getNodeSlowNodeApi() *api.Api {
	return api.NewApi(slowNodetApi, nil, []*api.Api{
		slownodeapi.NodeStart(),
		slownodeapi.NodeStop(),
		slownodeapi.NodeReload(),
	})
}
