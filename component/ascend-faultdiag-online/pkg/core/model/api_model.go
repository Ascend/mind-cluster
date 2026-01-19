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
Package model 服务相关模型。
*/
package model

// RequestContext 包含请求和响应信息以及结束标记。
type RequestContext struct {
	Api        string // 请求接口
	ReqJson    string // 请求json字符串
	Response   *ResponseBody
	FinishChan chan struct{} // 完成标记
}
