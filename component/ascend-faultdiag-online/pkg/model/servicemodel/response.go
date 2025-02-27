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
package servicemodel

import "ascend-faultdiag-online/pkg/model/enum"

// ResponseBody 返回报文结构体，包含状态、消息和数据。
type ResponseBody struct {
	Status enum.ResponseBodyStatus `json:"status"` // 状态
	Msg    string                  `json:"msg"`    // 消息
	Data   interface{}             `json:"data"`   // 数据
}

// ErrorResponse 异常回报
func ErrorResponse(errMsg string) *ResponseBody {
	return &ResponseBody{
		Status: enum.Error,
		Msg:    errMsg,
	}
}

// Influence 故障影响范围的结构体
type Influence struct {
	NodeIp string `json:"nodeIp"`
	PhyIds []int  `json:"phyIds"`
}

// Fault 故障列表信息
type Fault struct {
	FaultType      enum.FaultType  `json:"faultType"`
	FaultCode      string          `json:"faultCode"`
	FaultState     enum.FaultState `json:"faultState"`
	FaultOccurTime int64           `json:"faultOccurTime"`
	FaultId        string          `json:"faultId"`
	Influence      []*Influence    `json:"influence"`
}

// FaultBody 返回故障结构体，包括生产者、时间戳和故障列表
type FaultBody struct {
	Producer string `json:"producer"` // 生产者
	Time     int64  `json:"time"`     // 时间戳
	Faults   Fault  `json:"faults"`   // 故障列表
}
