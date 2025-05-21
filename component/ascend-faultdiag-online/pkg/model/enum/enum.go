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
	// Cluster 集群部署模式
	Cluster DeployMode = "cluster"
	// Node 单节点部署模式
	Node DeployMode = "node"
)

// DeployModes 所有的部署模式
func DeployModes() []DeployMode {
	return []DeployMode{Cluster, Node}
}

// LogLevel 定义日志级别枚举类型
type LogLevel string

const (
	// LgInfo 表示信息级别的日志，用于记录一般的操作信息
	LgInfo LogLevel = "info"
	// LgDebug 表示调试级别的日志，用于记录详细的调试信息
	LgDebug LogLevel = "debug"
	// LgWarn 表示警告级别的日志，用于记录潜在的问题或异常情况
	LgWarn LogLevel = "warn"
	// LgError 表示错误级别的日志，用于记录错误信息
	LgError LogLevel = "error"
)

// LogLevels 所有的日志级别
func LogLevels() []LogLevel {
	return []LogLevel{LgInfo, LgDebug, LgWarn, LgError}
}

// RequestType 定义请求类型
type RequestType string

const (
	// EventRequest 请求类型：event
	EventRequest RequestType = "event"
	// MetricRequest 请求类型：metric
	MetricRequest RequestType = "metric"
)

// ResponseBodyStatus 返回请求体状态
type ResponseBodyStatus string

const (
	// Success 响应成功时的请求体状态
	Success ResponseBodyStatus = "success"
	// Error 响应失败时的请求体状态
	Error ResponseBodyStatus = "error"
)

// ResponseBodyStatuses 返回所有可能的响应体状态
func ResponseBodyStatuses() []ResponseBodyStatus {
	return []ResponseBodyStatus{Success, Error}
}

// FaultType 故障类型
type FaultType string

const (
	// NodeFault 节点故障
	NodeFault FaultType = "node"
	// ChipFault 芯片故障
	ChipFault FaultType = "chip"
	// SwitchFault 交换机故障
	SwitchFault FaultType = "switch"
)

// FaultState 故障状态
type FaultState string

const (
	// OccurState 故障状态为发生
	OccurState FaultState = "occur"
	// RecoveryState 故障状态为恢复
	RecoveryState FaultState = "recovery"
)

// MetricDomainType 指标域
type MetricDomainType string

const (
	// NpuDomain NPU中的指标域
	NpuDomain = "npu"
	// HostDomain 主机中的指标域
	HostDomain = "host"
	// NetworkDomain 网络中的指标域
	NetworkDomain = "network"
	// NpuChipDomain Npu芯片中的指标域
	NpuChipDomain = "npu_chip"
	// TrainingTaskDomain 训练任务
	TrainingTaskDomain = "task"
)

// GetMetricDomains get all the domains of metric
func GetMetricDomains() []MetricDomainType {
	return []MetricDomainType{NpuDomain, HostDomain, NetworkDomain}
}

// MetricValueType the type of metric value
type MetricValueType string

const (
	// FloatMetric 浮点指标值
	FloatMetric MetricValueType = "float"
	// StringMetric 字符串指标值
	StringMetric MetricValueType = "string"
)

// SlowNodeEventType is the type of slow node
type SlowNodeEventType string

const (
	// SlowNodeAlgo means slow node should do the algorithms to detect the node/cluster
	SlowNodeAlgo SlowNodeEventType = "slowNodeAlgo"
	// DataParse means slow node should do the data parse
	DataParse SlowNodeEventType = "dataParse"
)

// Command is the command for execute API in so package
type Command string

const (
	// Start
	Start Command = "start"
	// Stop
	Stop Command = "stop"
	// Register a callback function
	Register Command = "registerCallBack"
	// Reload
	Reload Command = "reload"
)

const (
	// SlowNode
	SlowNode string = "slowNode"
	// NetFault
	NetFault string = "netFault"
)
