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
Package enum 提供枚举类
*/
package enum

// DeployMode 定义部署模式枚举类型
type DeployMode string

const (
	Cluster DeployMode = "cluster"
	Node    DeployMode = "node"
)

// DeployModes 所以的部署模式
var DeployModes = []DeployMode{Cluster, Node}

// LogLevel 定义日志级别枚举类型
type LogLevel string

const (
	LgInfo  LogLevel = "info"
	LgDebug LogLevel = "debug"
	LgWarn  LogLevel = "warn"
	LgError LogLevel = "error"
)

var LogLevels = []LogLevel{LgInfo, LgDebug, LgWarn, LgError}

// RequestType 定义请求类型
type RequestType string

const (
	Event  RequestType = "event"
	Metric RequestType = "metricdiag"
)

// ResponseBodyStatus 返回请求体状态
type ResponseBodyStatus string

const (
	Success ResponseBodyStatus = "success"
	Error   ResponseBodyStatus = "error"
)

// ResponseBodyStatuses 返回所有可能的响应体状态
var ResponseBodyStatuses = []ResponseBodyStatus{Success, Error}

// FaultType 故障类型
type FaultType string

const (
	NodeFault   FaultType = "node"
	ChipFault   FaultType = "chip"
	SwitchFault FaultType = "switch"
)

// FaultState 故障状态
type FaultState string

const (
	OccurState    FaultState = "occur"
	RecoveryState FaultState = "recovery"
)
