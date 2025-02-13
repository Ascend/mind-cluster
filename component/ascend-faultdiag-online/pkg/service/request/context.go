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
Package request 提供请求上下文管理功能。
*/
package request

import "ascend-faultdiag-online/pkg/model/servicemodel"

// Context 包含请求和响应信息以及结束标记。
type Context struct {
	Api        string // 请求接口
	ReqJson    string //请求json字符串
	Response   *servicemodel.ResponseBody
	FinishChan chan struct{} // 完成标记
}

// NewRequestContext 创建一个新的请求上下文。
func NewRequestContext(api string, reqJson string) *Context {
	return &Context{
		Api:        api,
		ReqJson:    reqJson,
		Response:   &servicemodel.ResponseBody{},
		FinishChan: make(chan struct{}),
	}
}
