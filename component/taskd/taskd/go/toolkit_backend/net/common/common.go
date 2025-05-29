/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package common defines common constants and types used by the toolkit backend.
package common

import "time"

const (
	// MgrRole represents the manager role.
	MgrRole = "Mgr"

	// ProxyRole represents the proxy role.
	ProxyRole = "Proxy"

	// AgentRole represents the agent role.
	AgentRole = "Agent"

	// WorkerRole represents the worker role.
	WorkerRole = "Worker"
)

const (
	MgrLevel    = 3
	ProxyLevel  = 2
	AgentLevel  = 1
	WorkerLevel = 1
)

// roleLevelMap maps roles to their corresponding levels.
var roleLevelMap = map[string]int{
	MgrRole:    MgrLevel,
	ProxyRole:  ProxyLevel,
	AgentRole:  AgentLevel,
	WorkerRole: WorkerLevel,
}

// MaxRoleLevel represents the maximum role level.
const MaxRoleLevel = 3

// MinRoleLevel represents the minimum role level.
const MinRoleLevel = 1

// RoleLevel returns the level of a given role. If the role does not exist, it returns -1.
func RoleLevel(role string) int {
	if level, exist := roleLevelMap[role]; exist {
		return level
	}
	return -1
}

// roleProcessProperty maps roles to their process properties.
var roleProcessProperty = map[string]bool{
	MgrRole:    false,
	ProxyRole:  false,
	AgentRole:  false,
	WorkerRole: true,
}

// RoleHasProcessProperty checks if a given role has process properties.
func RoleHasProcessProperty(role string) bool {
	haveProcessRank, exist := roleProcessProperty[role]
	if !exist {
		return false
	}
	return haveProcessRank
}

const (
	MgrBufSize    = 1024
	ProxyBufSize  = 32
	AgentBufSize  = 8
	WorkerBufSize = 8
)

// roleRecvBuf maps roles to their receive buffer sizes.
var roleRecvBuf = map[string]int{
	MgrRole:    MgrBufSize,
	ProxyRole:  ProxyBufSize,
	AgentRole:  AgentBufSize,
	WorkerRole: WorkerBufSize,
}

// RoleRecvBuffer returns the receive buffer size of a given role. If the role does not exist, it returns -1.
func RoleRecvBuffer(role string) int {
	bufSize, exist := roleRecvBuf[role]
	if !exist {
		return -1
	}
	return bufSize
}

// MetaRoleKey is the key for the role in metadata.
const (
	// MetaRoleKey is the key for the role in metadata.
	MetaRoleKey = "role"

	// MetaServerRankKey is the key for the server rank in metadata.
	MetaServerRankKey = "serverRank"

	// MetaProcessRankKey is the key for the process rank in metadata.
	MetaProcessRankKey = "processRank"

	// BroadCastPos represents the broadcast position.
	BroadCastPos = "All"

	// NoneProcessRank represents the non-existent process rank.
	NoneProcessRank = "None"

	// Dst2Self represents the destination is the same as the source.
	Dst2Self = "dst2self"

	// Dst2SameLevel represents the destination is at the same level as the source.
	Dst2SameLevel = "dst2sameLevel"

	// Dst2LowerLevel represents the destination is at a lower level than the source.
	Dst2LowerLevel = "dst2lowerLevel"

	// Dst2UpperLevel represents the destination is at a higher level than the source.
	Dst2UpperLevel = "dst2upperLevel"

	// DataFromUpper indicates the data comes from an upper level.
	DataFromUpper = "dataFromUpper"

	// DataFromLower indicates the data comes from a lower level.
	DataFromLower = "dataFromLower"

	// DataFromSelf indicates the data comes from the same level.
	DataFromSelf = "dataFromSelf"
)

const (
	MgrGrNum    = 128
	ProxyGrNum  = 4
	AgentGrNum  = 1
	WorkerGrNum = 1
)

// roleWorkerNum maps roles to their worker numbers.
var roleWorkerNum = map[string]int{
	MgrRole:    MgrGrNum,
	ProxyRole:  ProxyGrNum,
	AgentRole:  AgentGrNum,
	WorkerRole: WorkerGrNum,
}

// RoleWorkerNum returns the number of workers for a given role. If the role does not exist, it returns -1.
func RoleWorkerNum(role string) int {
	if num, exist := roleWorkerNum[role]; exist {
		return num
	}
	return -1
}

const (
	// MaxGRPCRecvMsgSize is the maximum size of a gRPC receive message.
	MaxGRPCRecvMsgSize = 8 * 1024 * 1024

	// MaxGRPCSendMsgSize is the maximum size of a gRPC send message.
	MaxGRPCSendMsgSize = 8 * 1024 * 1024

	// GrpcQps is the QPS limit for gRPC.
	GrpcQps = 10000

	// MaxRegistryNum is the maximum number of registries.
	MaxRegistryNum = 5000

	// NetworkRetryTimes is the number of network retry attempts.
	NetworkRetryTimes = 3

	// RetryPeriod is the period between network retry attempts.
	RetryPeriod = 10 * time.Millisecond

	// AckTimeout is the timeout for ACK messages.
	AckTimeout = 3 * time.Second

	// KeepAlivePeriod is the period for keep-alive messages.
	KeepAlivePeriod = 5 * time.Second

	// KeepAliveTimeout is the timeout for keep-alive messages.
	KeepAliveTimeout = time.Second
)
