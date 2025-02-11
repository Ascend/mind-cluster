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

// Context 包含请求和响应信息以及结束标记。
type Context struct {
	Request    *Body
	Response   *ResponseBody
	FinishChan chan struct{} // 结束标记
}

// NewRequestContext 创建一个新的请求上下文。
func NewRequestContext(req *Body) *Context {
	return &Context{
		Request:    req,
		Response:   &ResponseBody{},
		FinishChan: make(chan struct{}),
	}
}
