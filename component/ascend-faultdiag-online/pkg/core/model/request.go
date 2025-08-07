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
Package model 提供请求上下文管理功能。
*/
package model

import (
	"encoding/json"
	"time"

	"ascend-faultdiag-online/pkg/core/model/enum"
)

// Body 表示一个请求体
type Body struct {
	Component   string           `json:"component"`    // 消息来源组件，如 noded
	RequestType enum.RequestType `json:"request_type"` // 消息类型：event（事件）或 metric（指标）
	Name        string           `json:"name"`         // 消息名称，事件名或指标名
	SendTime    string           `json:"send_time"`    // 发送时间
	ReceiveTime time.Time        `json:"receive_time"` // 接收时间
	Msg         string           `json:"msg"`          // 消息文本
	Data        any              `json:"data"`         // 详细结构化信息，根据消息类型决定
}

// NewRequestBodyFromJson 从 JSON 字符串中解析创建 Body 请求体
func NewRequestBodyFromJson(msgJson string) (*Body, error) {
	request := Body{}
	if err := json.Unmarshal([]byte(msgJson), &request); err != nil {
		return nil, err
	}
	request.ReceiveTime = time.Now()
	return &request, nil
}

// NewRequestContext get a new struct object of request context
func NewRequestContext(api string, reqJson string) *RequestContext {
	return &RequestContext{
		Api:        api,
		ReqJson:    reqJson,
		Response:   &ResponseBody{},
		FinishChan: make(chan struct{}, 1),
	}
}

// CommonReqModel is a common req model for empty req
type CommonReqModel struct{}

// CallbackFunc for the data callback from algo_src
type CallbackFunc func(string)

// Input is an input data in the execute function of so.
type Input struct {
	// Command what the operation the so package should do: start/stop
	Command enum.Command
	// Target cluster/node
	Target enum.DeployMode
	// Model is the data which so package will process
	Model any
	// Func is the uintptr of the callback func, using uint replace uintptr
	// the so package will call this func if there is callback data
	Func CallbackFunc
	// EventType is the enum of execution type, slowNode or dataParse
	EventType string
}
