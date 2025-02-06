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

import "ascend-faultdiag-online/pkg/model/enum"

type ResponseBody struct {
	Status string      `json:"status"` // 状态
	Msg    string      `json:"msg"`    // 消息
	Data   interface{} `json:"data"`   // 数据
}

type Influence struct {
	NodeIp string `json:"nodeIp"`
	PhyIds []int  `json:"phyIds"`
}

type Fault struct {
	FaultType      enum.FaultType  `json:"faultType"`
	FaultCode      string          `json:"faultCode"`
	FaultState     enum.FaultState `json:"faultState"`
	FaultOccurTime int64           `json:"faultOccurTime"`
	FaultId        string          `json:"faultId"`
	Influence      []*Influence    `json:"influence"`
}

type FaultBody struct {
	Producer string `json:"producer"` // 生产者
	Time     int64  `json:"time"`     // 时间戳
	Faults   Fault  `json:"faults"`   // 故障列表
}
