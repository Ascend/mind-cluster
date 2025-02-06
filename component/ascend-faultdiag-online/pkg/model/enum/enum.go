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

type DeployMode string

const (
	Cluster DeployMode = "cluster"
	Node    DeployMode = "node"
)

var DeployModes = []DeployMode{Cluster, Node}

type LogLevel string

const (
	LgInfo  LogLevel = "info"
	LgDebug LogLevel = "debug"
	LgWarn  LogLevel = "warn"
	LgError LogLevel = "error"
)

var LogLevels = []LogLevel{LgInfo, LgDebug, LgWarn, LgError}

type RequestType string

const (
	Event  RequestType = "event"
	Metric RequestType = "metric_diag"
)

type ResponseBodyStatus string

const (
	Success ResponseBodyStatus = "success"
	Error   ResponseBodyStatus = "error"
)

var ResponseBodyStatuses = []ResponseBodyStatus{Success, Error}

type FaultType string

const (
	NodeFault   FaultType = "node"
	ChipFault   FaultType = "chip"
	SwitchFault FaultType = "switch"
)

type FaultState string

const (
	OccurState    FaultState = "occur"
	RecoveryState FaultState = "recovery"
)
