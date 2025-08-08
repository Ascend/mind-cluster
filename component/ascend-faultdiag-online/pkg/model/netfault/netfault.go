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
Package netfault 特性类型定义
*/
package netfault

// ClusterResult 集群中心侧的结果模型
type ClusterResult struct {
	// TaskID 网络故障检测的任务Id
	TaskID string `json:"taskId"`
	// TimeStamp 时间戳
	TimeStamp int `json:"timestamp"`
	// MinLossRate 最小丢包率
	MinLossRate float64 `json:"minLossRate"`
	// MaxLossRate 最大丢包率
	MaxLossRate float64 `json:"maxLossRate"`
	// AvgLossRate 平均丢包率
	AvgLossRate float64 `json:"avgLossRate"`
	// MinDelay 最小网络延时
	MinDelay float64 `json:"minDelay"`
	// MaxDelay 最大网络延时
	MaxDelay float64 `json:"maxDelay"`
	// AvgDelay 平均网络时延
	AvgDelay float64 `json:"avgDelay"`
	// SrcID 源地址
	SrcID string `json:"srcId"`
	// SrcType 源地址类型
	SrcType int `json:"srcType"`
	// DstID 目标地址
	DstID string `json:"dstId"`
	// DstType 目标地址类型
	DstType int `json:"dstType"`
	// Level 故障等级
	Level int `json:"level"`
	// FaultType 故障类型
	FaultType int `json:"faultType"`
}
