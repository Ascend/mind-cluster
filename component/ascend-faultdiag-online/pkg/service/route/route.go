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
Package route 路由
*/
package route

import (
	"errors"
	"fmt"
	"strings"

	"ascend-faultdiag-online/pkg/service/serviceapi"
	"ascend-faultdiag-online/pkg/service/servicecore"
	"ascend-faultdiag-online/pkg/utils/constants"
)

// Router 路由
type Router struct {
	RootApi *servicecore.Api // Api根节点
}

// NewRouter 创建路由实例
func NewRouter() *Router {
	return &Router{
		RootApi: serviceapi.GetRootApi(),
	}
}

// HandleApi 处理请求
func (router *Router) HandleApi(api string) (servicecore.ApiFunc, error) {
	if router == nil {
		return nil, errors.New("the router is nil")
	}
	tempApiNode := router.RootApi
	routeParts := strings.Split(api, constants.ApiSeparator)
	for _, routePart := range routeParts {
		nextApiNode, ok := tempApiNode.SubApiMap[routePart]
		if !ok {
			return nil, fmt.Errorf("api %s is not existed", api)
		}
		tempApiNode = nextApiNode
	}
	return tempApiNode.ApiFunc, nil
}
