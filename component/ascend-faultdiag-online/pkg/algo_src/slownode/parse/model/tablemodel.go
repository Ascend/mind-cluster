/* Copyright 2025 Huawei Technologies Co., Ltd.

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
Package model.
*/
package model

// CAnnApi is the structure corresponding to the CANN_API database
type CAnnApi struct {
	// StartNs is the start timestamp
	StartNs int64 `db:"startNs"`
	// EndNs is the end timestamp
	EndNs int64 `db:"endNs"`
	// ApiType indicates the hierarchy level the API belongs to
	ApiType int `db:"type"`
	// GlobalTid is the global TID that this API belongs to
	GlobalTid int64 `db:"globalTid"`
	// ConnectionId is the primary key ID
	ConnectionId int64 `db:"connectionId"`
	// Name is the name identifier of this API
	Name int64 `db:"name"`
}

// CommOp is the structure corresponding to the COMMUNICATION_OP database
type CommOp struct {
	// OpName is the operator name identifier
	OpName int64 `db:"opName"`
	// StartNs is the start timestamp
	StartNs int64 `db:"startNs"`
	// EndNs is the end timestamp
	EndNs int64 `db:"endNs"`
	// ConnectionId is the unique identifier
	ConnectionId int64 `db:"connectionId"`
	// GroupName is the communication domain name identifier
	GroupName int64 `db:"groupName"`
	// OpId is the operator identifier
	OpId int64 `db:"opId"`
	// Relay is the relay communication flag
	Relay int `db:"relay"`
	// Retry is the retransmission flag
	Retry int `db:"retry"`
	// DataType is the data type identifier
	DataType int64 `db:"dataType"`
	// AlgType is the algorithm type
	AlgType int `db:"algType"`
	// Count is the number of data items of DataType transferred by the operator
	Count int64 `db:"count"`
	// OpType is the operator type identifier
	OpType int64 `db:"opType"`
}

// MSTXEvents is the structure corresponding to the MSTX_EVENTS database
type MSTXEvents struct {
	// StartNs is the start timestamp
	StartNs int64 `db:"startNs"`
	// EndNs is the end timestamp
	EndNs int64 `db:"endNs"`
	// EventType is the type of TX data
	EventType int `db:"eventType"`
	// RangeId is the range ID corresponding to range-type TX data
	RangeId int64 `db:"rangeId"`
	// Category is the category ID that the TX data belongs to
	Category int `db:"category"`
	// Message is the information carried by the TX data
	Message int64 `db:"message"`
	// GlobalTid is the global TID of the thread where the TX data starts
	GlobalTid int64 `db:"globalTid"`
	// EndGlobalTid is the global TID of the thread where the TX data ends
	EndGlobalTid int64 `db:"endGlobalTid"`
	// DomainId is the domain ID of the domain that the TX data belongs to
	DomainId int64 `db:"domainId"`
	// ConnectionId is the unique identifier
	ConnectionId int64 `db:"connectionId"`
}

// StepTime is the structure for step time table, corresponding to the TSTEP_TIME database
type StepTime struct {
	// Id is the unique identifier
	Id int64 `db:"id"`
	// StartNs is the start timestamp
	StartNs int64 `db:"startNs"`
	// EndNs is the end timestamp
	EndNs int64 `db:"endNs"`
}

// Task is the structure corresponding to the TASK database
type Task struct {
	// StartNs is the start timestamp
	StartNs int64 `db:"startNs"`
	// EndNs is the end timestamp
	EndNs int64 `db:"endNs"`
	// DeviceId is the device identifier
	DeviceId int64 `db:"deviceId"`
	// ConnectionId is the unique identifier
	ConnectionId int64 `db:"connectionId"`
	// GlobalTaskId is the unique identifier of the global operator task
	GlobalTaskId int64 `db:"globalTaskId"`
	// GlobalPid is the PID during the execution of the operator task
	GlobalPid int64 `db:"globalPid"`
	// TaskType is the type of accelerator that executes the operator on the device
	TaskType int `db:"taskType"`
	// ContextId is the identifier for distinguishing subgraph small operators
	ContextId int64 `db:"contextId"`
	// StreamId is the stream identifier corresponding to the operator task
	StreamId int64 `db:"streamId"`
	// TaskId is the operator task identifier
	TaskId int64 `db:"taskId"`
	// ModelId is the model identifier corresponding to the operator task
	ModelId int64 `db:"modelId"`
}
