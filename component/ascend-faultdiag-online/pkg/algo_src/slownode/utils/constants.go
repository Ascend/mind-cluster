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

/*
Package utils is used for file reading and writing, as well as data processing.
*/
package utils

const (
	serverListField = "server_list"
	xdlIpField      = "XDL_IP"
	serverIdField   = "server_id"
	deviceField     = "device"
	rankIdField     = "rank_id"
	task            = "task"
	topoName        = "topo.json"
	ranktableName   = "ranktable.json"
	minColumns      = 2
	zp              = "ZP" // 自我抽象统一的并行策略
	tp              = "TP"
	pp              = "PP"
	fileMode        = 0644
)

const (
	// Stop sign for command which is in user's input
	Stop = "stop"
	// Start sign for command which is in user's input
	Start = "start"
	// Reload sign for command which is in user's input
	Reload = "reload"
	// RegisterCallBack sign for command which is in user's input
	RegisterCallBack = "registerCallBack"
	// Node sign for target which is in user's input
	Node = "node"
	// Cluster sign for target which is in user's input
	Cluster = "cluster"
)
