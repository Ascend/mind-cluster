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

// Package constant a package for constant
package constant

import (
	"sync"
	"time"
)

const (
	// LogFilePathEnv for log file path environment
	LogFilePathEnv = "TASKD_LOG_PATH"
	// LogFileName default log file name
	LogFileName = "taskd.log"
)

const (
	// DefaultLogFilePath default log file
	DefaultLogFilePath = "./taskd_log/"
	// DefaultLogLevel default log level
	DefaultLogLevel = 0
	// DefaultMaxBackups max backup log file num
	DefaultMaxBackups = 10
	// DefaultMaxAge max age backup file exist
	DefaultMaxAge = 7
	// DefaultMaxLineLength max line length in log
	DefaultMaxLineLength = 1023
)

const (
	// TaskThreadHold the threshold of task queue
	TaskThreadHold = 0.8
	// DiskUsageCheckInterval the interval between each check
	DiskUsageCheckInterval = 60
)

const (
	// SizeLimitPerProfilingFile the size of each profiling
	SizeLimitPerProfilingFile = 10 * 1024 * 1024
	// BytesPerMB default bytes per in 1 MB
	BytesPerMB = 1024 * 1024
	// DefaultDiskUpperLimitInMB the initial size of upper limit
	DefaultDiskUpperLimitInMB = 5 * 1024
	// LeastBufferSize is the buffer size for flush file
	LeastBufferSize = 4096
	// CheckProfilingCacheInterval every interval to check cache
	CheckProfilingCacheInterval = 5 * time.Second
	// DomainCheckInterval the interval between each check of domain change
	DomainCheckInterval = 1 * time.Second
	// ProfilingFileMode the mode of profiling file
	ProfilingFileMode = 0644
	// ProfilingDirMode the mode of profiling directory
	ProfilingDirMode = 0755

	// NumberOfParts the number of parts
	NumberOfParts = 5
	// MinProfilingFileNum the number of profiling file to reserve
	MinProfilingFileNum = 3
	// TenBase is 10 base number
	TenBase = 10
	// BitSize64 is the 64 bit size
	BitSize64 = 64
	// PathLengthLimit is the max length of path
	PathLengthLimit = 1024
	// DefaultRecordLength the default length of records to estimate the size of buffer
	DefaultRecordLength = 200
)

const (
	// TaskBufferSize is the buffer size for each rank
	TaskBufferSize = 20
	// NormalBufferSizeInBytes is the buffer size used in common scenario
	NormalBufferSizeInBytes = 2048 * 1024
	// MaxRequestBufferNum is the max num of buffer at same time
	MaxRequestBufferNum = 100
)

const (
	// TaskUidKey is the uid of acjob which is in env wrote by ascend-operator
	TaskUidKey = "MINDX_TASK_ID"
	// ProfilingBaseDir is the path store all profiling data
	ProfilingBaseDir = "/user/cluster-info/profiling"
	// MsptiLibPath the path of mspti
	// MsptiLibPath the path of mspti so need to consider other path by user
	MsptiLibPath = "/usr/local/Ascend/ascend-toolkit/latest/lib64/"
	// ProfilingSwitchFilePath the path of the switch controller, wrote by device-plugin
	ProfilingSwitchFilePath = "/user/cluster-info/datatrace-config/profilingSwitch"
	// LineSeperator is the line separator for each record
	LineSeperator = '\n'
)

const (
	// DefaultDomainName default domain name
	DefaultDomainName = "default"
	// CommunicationDomainName communication domain name
	CommunicationDomainName = "communication"
	// SwitchOFF off status
	SwitchOFF = "off"
	// SwitchON on status
	SwitchON = "on"
)

// MuMark mutex locker for marker
var MuMark sync.Mutex

// MuApi mutex locker for api
var MuApi sync.Mutex

// MuKernal mutex locker for kernel
var MuKernal sync.Mutex
