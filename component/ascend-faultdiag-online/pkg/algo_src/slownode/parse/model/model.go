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

import (
	"sync"
)

// MergeParallelGroupInfoInput 并行域合并输入结构体
type MergeParallelGroupInfoInput struct {
	// FileMu 访问文件时加锁
	FileMu sync.Mutex
	// FilePaths 文件路径列表
	FilePaths []string
	// FileSavePath 文件保存路径
	FileSavePath string
	// DeleteFileFlag 删除文件标记
	DeleteFileFlag bool
}

// MergeParallelGroupInfoResult 并行域信息回调结果
type MergeParallelGroupInfoResult struct {
	// JobName the name of job
	JobName string `json:"jobName"`
	// jobId the unique id of a job
	JobId string `json:"jobId"`
	// IsFinished whether the data parse finished or not
	IsFinished bool `json:"isFinished"`
	// FinishedTime the data parse finished time, sample: 1745567190000
	FinishedTime int64 `json:"finishedTime"`
}

// NodeDataParseResult the model of data parse result which callback from slownode in node
type NodeDataParseResult struct {
	// JobName the name of job
	JobName string `json:"jobName"`
	// jobId the unique id of a job
	JobId string `json:"jobId"`
	// IsFinished whether the data parse finished or not
	IsFinished bool `json:"isFinished"`
	// FinishedTime the data parse finished time, sample: 1745567190000
	FinishedTime int64 `json:"finishedTime"`
	// StepCount is the step data steptime.csv
	StepCount int64 `json:"stepCount"`
	// RankIds is the rank ids slice
	RankIds []string `json:"rankIds"`
}

// ParseJobInfo 保存停止信号
type ParseJobInfo struct {
	JobName       string
	JobId         string
	JobStatus     string
	StopParseFlag chan struct{}
	JobWg         *sync.WaitGroup
	StopWg        *sync.WaitGroup
	TimeStamp     int64
}
