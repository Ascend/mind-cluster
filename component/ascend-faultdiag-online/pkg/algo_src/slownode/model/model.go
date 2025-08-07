/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package model is used for providing slow node related model entities
package model

// ClusterResult 集群结果
type ClusterResult struct {
	// SlowCalculateRanks 慢计算卡
	SlowCalculateRanks []int `json:"slowCalculateRanks"`
	// SlowCommunicationDomains 慢通信域
	SlowCommunicationDomains [][]int `json:"slowCommunicationDomains"`
	// SlowHostNodes hosts侧慢卡
	SlowHostNodes []string `json:"slowHostNodes"`
	// SlowIONodes 慢IO卡的节点
	SlowIORanks []int `json:"slowIORanks"`
	// SlowCommunicationRanks  慢通信卡
	SlowCommunicationRanks []int `json:"slowCommunicationRanks"`
	// IsSlow 是否存在慢节点
	IsSlow int `json:"isSlow"`
	// DegradationLevel 劣化百分点
	DegradationLevel string `json:"degradationLevel"`
}

// NodeResult 定义数据结构，确保与文件中的 JSON 格式匹配
type NodeResult struct {
	// SlowCalculateRank 节点级慢计算卡
	SlowCalculateRank []int `json:"slowCalculateRanks"`
	// SlowCommunicationDomain 节点级慢通信域
	SlowCommunicationDomain [][]int `json:"slowCommunicationDomains"`
	// SlowSendRanks 节点级slow send
	SlowSendRanks []int `json:"slowSendRanks"`
	// SlowHostNodes 当前节点是否为慢节点
	SlowHostNodes []string `json:"slowHostNodes"`
	// SlowIORanks 节点级慢IO卡
	SlowIORanks []int `json:"slowIORanks"`
	// FileName 节点级结果路径
	FileName string `json:"FileName"`
}
