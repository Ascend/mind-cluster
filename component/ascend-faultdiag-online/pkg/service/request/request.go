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

import (
	"encoding/json"
	"time"

	"ascend-faultdiag-online/pkg/model/enum"
)

// Body 表示一个请求体
type Body struct {
	Component   string           `json:"component"`    // 消息来源组件，如 noded
	RequestType enum.RequestType `json:"request_type"` // 消息类型：event（事件）或 metricdiag（指标）
	Name        string           `json:"name"`         // 消息名称，事件名或指标名
	SendTime    string           `json:"send_time"`    // 发送时间
	ReceiveTime time.Time        `json:"receive_time"` // 接收时间
	Msg         string           `json:"msg"`          // 消息文本
	Data        interface{}      `json:"data"`         // 详细结构化信息，根据消息类型决定
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
