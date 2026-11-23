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

import (
	"sync"
)

// MergeParallelGroupInfoInput is the input structure for parallel domain merging
type MergeParallelGroupInfoInput struct {
	// FileMu is a mutex for file access
	FileMu sync.Mutex
	// FilePaths is a list of file paths
	FilePaths []string
	// FileSavePath is the file save path
	FileSavePath string
	// DeleteFileFlag is the flag for deleting files
	DeleteFileFlag bool
}

// MergeParallelGroupInfoResult callback result of parallel domain information callback
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

// ParseJobInfo stores stop signals and job-related information
type ParseJobInfo struct {
	// JobName is the name of the task
	JobName string
	// JobId is the identifier of the task
	JobId string
	// JobStatus indicates the status of the task
	JobStatus string
	// StopParseFlag is a channel for the stop signal of the parsing task
	StopParseFlag chan struct{}
	// JobWg is the wait group for the task
	JobWg *sync.WaitGroup
	// StopWg is the wait group for stopping the task
	StopWg *sync.WaitGroup
	// TimeStamp is the timestamp of the task
	TimeStamp int64
}
