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
Package model.
*/
package model

// CAnnApi 对应CANN_API数据库的结构体
type CAnnApi struct {
	StartNs      int64 `db:"startNs"`
	EndNs        int64 `db:"endNs"`
	ApiType      int   `db:"type"`
	GlobalTid    int64 `db:"globalTid"`
	ConnectionId int64 `db:"connectionId"`
	Name         int64 `db:"name"`
}

// CommOp 对应COMMUNICATION_OP数据库的结构体
type CommOp struct {
	OpName       int64 `db:"opName"`
	StartNs      int64 `db:"startNs"`
	EndNs        int64 `db:"endNs"`
	ConnectionId int64 `db:"connectionId"`
	GroupName    int64 `db:"groupName"`
	OpId         int64 `db:"opId"`
	Relay        int   `db:"relay"`
	Retry        int   `db:"retry"`
	DataType     int64 `db:"dataType"`
	AlgType      int   `db:"algType"`
	Count        int64 `db:"count"`
	OpType       int64 `db:"opType"`
}

// MSTXEvents 对应MSTX_EVENTS数据库的结构体
type MSTXEvents struct {
	StartNs      int64 `db:"startNs"`
	EndNs        int64 `db:"endNs"`
	EventType    int   `db:"eventType"`
	RangeId      int64 `db:"rangeId"`
	Category     int   `db:"category"`
	Message      int64 `db:"message"`
	GlobalTid    int64 `db:"globalTid"`
	EndGlobalTid int64 `db:"endGlobalTid"`
	DomainId     int64 `db:"domainId"`
	ConnectionId int64 `db:"connectionId"`
}

// StepTime step时间表，对应TSTEP_TIME数据库的结构体
type StepTime struct {
	Id      int64 `db:"id"`
	StartNs int64 `db:"startNs"`
	EndNs   int64 `db:"endNs"`
}

// Task 对应TASK数据库的结构体
type Task struct {
	StartNs      int64 `db:"startNs"`
	EndNs        int64 `db:"endNs"`
	DeviceId     int64 `db:"deviceId"`
	ConnectionId int64 `db:"connectionId"`
	GlobalTaskId int64 `db:"globalTaskId"`
	GlobalPid    int64 `db:"globalPid"`
	TaskType     int   `db:"taskType"`
	ContextId    int64 `db:"contextId"`
	StreamId     int64 `db:"streamId"`
	TaskId       int64 `db:"taskId"`
	ModelId      int64 `db:"modelId"`
}
