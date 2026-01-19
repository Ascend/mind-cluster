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

// Package constants provides some constant value
package constants

import "time"

const (
	// FlagMarkerStartWithHost 调用mstxRangeStartA接口传入stream设置为nullptr时，MSPTI_ACTIVITY_KIND_MARKER使用
	FlagMarkerStartWithHost = 1 << 1
	// FlagMarkerEndWithHost 调用mstxRangeEnd传入的id来自传入stream设置nullptr的mstxRangeStart，MSPTI_ACTIVITY_KIND_MARKER使用
	FlagMarkerEndWithHost = 1 << 2
	// FlagMarkerStartWithDevice 调用mstxRangeStartA接口传入有效stream时，对应的打点数据类型，MSPTI_ACTIVITY_KIND_MARKER使用
	FlagMarkerStartWithDevice = 1 << 4
	// FlagMarkerEndWithDevice 调用mstxRangeEnd传入的id来自传入有效stream时的mstxRangeStart，MSPTI_ACTIVITY_KIND_MARKER使用
	FlagMarkerEndWithDevice = 1 << 5
	// SourceKindHost 标记数据的来源是Host
	SourceKindHost = 0
	// SourceKindDevice 标记数据的来源是Device
	SourceKindDevice = 1

	// DbCAnnApi 数据库名称CANN_API
	DbCAnnApi = "CANN_API"
	// DbCommOp 数据库名称COMMUNICATION_OP
	DbCommOp = "COMMUNICATION_OP"
	// DbMSTXEvents 数据库名称MSTX_EVENTS
	DbMSTXEvents = "MSTX_EVENTS"
	// DbStepTime 数据库名称STEP_TIME
	DbStepTime = "STEP_TIME"
	// DbTask 数据库名称TASK
	DbTask = "TASK"
	// DbStringIds 数据库名称STRING_IDS
	DbStringIds = "STRING_IDS"

	// DomainComm Json数据中Domain属性字段选项：communication
	DomainComm = "communication"
	// DomainDefault Json数据中Domain属性字段默认选项：default
	DomainDefault = "default"

	// StepWord Json数据中Name属性字段选项：step
	StepWord = "step"
	// CKPTWord Json数据中Name属性字段选项：save_checkpoint
	CKPTWord = "save_checkpoint"
	// ForwardWord Json数据中Name属性字段选项：forward
	ForwardWord = "forward"
	// DataLoaderWord Json数据中Name属性字段选项：dataloader
	DataLoaderWord = "dataloader"

	// DefaultFilePermission 默认文件权限为八进制格式0640。 -rw-r-----
	DefaultFilePermission = 0640
	// InsertNumber 每一次插入SQL语句条数
	InsertNumber = 200
	// DefaultNameIndex 默认Name属性值
	DefaultNameIndex = -1
	// DecimalMark 表示十进制
	DecimalMark = 10
	// Base64Mark 表示64编码
	Base64Mark = 64
	// MaxRankNum 卡的最大数量
	MaxRankNum = 16
	// LoopCount 循环次数
	LoopCount = 60

	// PollTime 每10s轮询一次
	PollTime = 10 * time.Second
	// CallbackTime 每30s回调结果
	CallbackTime = 30 * time.Second
	// ParGroupTime 每20s合并一次并行域
	ParGroupTime = 20 * time.Second
	// WaitStopTime 等待job停止时间
	WaitStopTime = 10 * time.Minute
	// TimeoutFindFile 查找文件超时时间
	TimeoutFindFile = 2 * time.Hour
	// StopPoll 查找停止状态时间
	StopPoll = 5 * time.Second

	// GlobalRankCsvFileName 保存通信算子数据的csv文件
	GlobalRankCsvFileName = "comm.csv"
	// StepTimeCsvFileName 保存迭代时延信息的csv文件
	StepTimeCsvFileName = "steptime.csv"
	// ParGroupJsonFileName 保存并行域信息的json文件
	ParGroupJsonFileName = "parallel_group.json"
	// ClusterParGroupFileName 保存集群侧汇总并行域信息的json文件
	ClusterParGroupFileName = "parallel_group_global.json"
	// DbFileName 数据库db文件
	DbFileName = "database.db"

	// JobRunStatus job运行状态
	JobRunStatus = "running"
	// JobStopStatus job已停止状态
	JobStopStatus = "stopped"
	// JobStoppingStatus job正在停止状态
	JobStoppingStatus = "stopping"

	// FileMaxSize max size for a file
	FileMaxSize int64 = 1024 * 1024 * 1024 // 1GB
)
