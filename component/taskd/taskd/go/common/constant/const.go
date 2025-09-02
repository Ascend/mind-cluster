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
	LogFileName          = "taskd.log"
	WorkerLogPathPattern = "taskd-worker-%s.log"
	// ProxyLogPathPattern for proxy log path pattern
	ProxyLogPathPattern = "taskd-proxy-%s.log"
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
	// HalfSize the halfsize
	HalfSize = 2
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

// All MsgBody's Code must be defined here
const (
	RegisterCode                 = 101
	RestartTimeCode              = 201
	FaultRankCode                = 202
	ExitAgentCode                = 203
	SwitchNicCode                = 204
	StressTestCode               = 205
	ProfilingAllCloseCmdCode     = 700
	ProfilingDefaultDomainOnCode = 710
	ProfilingCommDomainOnCode    = 701
	ProfilingAllOnCmdCode        = 711
)

// RequestChanNum message handler chan number
const RequestChanNum = 100

// MaxMsgQueueLength max length of message queue
const MaxMsgQueueLength = 40000

// All Mtype or MsgType must be defined here
const (
	REGISTER          = "REGISTER"
	STATUS            = "STATUS"
	Action            = "ACTION"
	ReportFaultRank   = "REPORT_FAULT_RANK"
	KeepAlive         = "KEEP_ALIVE"
	ReportRestartTime = "REPORT_RESTART_TIME"
	Exit              = "EXIT"
)

const (
	SwitchNic  = "SWITCH_NIC"
	StressTest = "STRESS_TEST"
)

// All num const must be defined here
const (
	Ten     = 10
	Hundred = 100
)

const (
	// MaxResendTimes the max times to resend message
	MaxResendTimes = 5
	// ResendSeconds time interval for resending messages
	ResendSeconds = 3
)

// All cluster info type must be defined here
const (
	ClusterRole  = "Cluster"
	ClusterDRank = "ClusterD"
	TaskDRank    = "TaskD"
)

// All cluster command must be defined here
const (
	DefaultDomainCmd = "DefaultDomainCmd"
	CommDomainCmd    = "CommDomainCmd"
)

// All status must be defined here
const (
	DefaultDomainStatus = "DefaultDomainStatus"
	CommDomainStatus    = "CommDomainStatus"
)

// All worker profiling execute status must be defined here
const (
	OffCode = 0
	OnCode  = 1
	ExpCode = 2
	On      = "On"
	Off     = "Off"
	Unknown = "Unknown"
	Exp     = "Exception"
)

// MuMark mutex locker for marker
var MuMark sync.Mutex

// MuApi mutex locker for api
var MuApi sync.Mutex

// MuKernal mutex locker for kernel
var MuKernal sync.Mutex

// All profiling execute result code definition
const (
	ProfilingAllCloseCode   = 1400
	ProfilingDefaultOpenInc = 10
	ProfilingDefaultExpInc  = 20
	ProfilingCommOpenInc    = 1
	ProfilingCommExpInc     = 2
)

// All kind of ProfilingExecRes
var ProfilingUnknownStatus = NewProfilingExecRes(Unknown)
var ProfilingOnStatus = NewProfilingExecRes(On)
var ProfilingOffStatus = NewProfilingExecRes(Off)
var ProfilingExpStatus = NewProfilingExecRes(Exp)

// All grpc ip must be defined here
const (
	// DefaultIP grpc manager ListenAddr default ip
	DefaultIP = "127.0.0.1"
	// MgrPort grpc manager ListenAddr port
	MgrPort      = ":9601"
	ProxyPort    = ":9602"
	ClusterdPort = ":8899"
)

const (
	// HandleStageInit indicate plugin handle state in start
	HandleStageInit = "Init"
	// HandleStageProcess indicate plugin handle state in process
	HandleStageProcess = "Process"
	// HandleStageFinal indicate plugin handle state in final
	HandleStageFinal = "Final"
	// HandleStageException indicate plugin handle state in exce
	HandleStageException = "Exception"
)

const (
	// CandidateStatus indicate plugin request stream
	CandidateStatus = "candidate"
	// UnselectStatus indicate plugin request stream
	UnselectStatus = "unselect"
)

// All kind of ProfilingWorkerState
var ProfilingWorkerOpenedState = NewWorkerProfilingState(Opened)
var ProfilingWorkerClosedState = NewWorkerProfilingState(Closed)
var ProfilingWorkerWaitOpenState = NewWorkerProfilingState(WaitOpen)
var ProfilingWorkerExceptionState = NewWorkerProfilingState(Exception)
var ProfilingWorkerWaitCloseState = NewWorkerProfilingState(WaitClose)

// Profiling states
const (
	Opened    = "opened"
	Closed    = "closed"
	WaitOpen  = "waitOpen"
	Exception = "exception"
	WaitClose = "waitClose"
	Invalid   = "Invalid"
)

// All env variables
const (
	MindxServerIp = "MINDX_SERVER_IP"
)

// StreamName and PluginName
const (
	ProfilingStream     = "ProfilingCollect"
	ProfilingPluginName = "ProfilingPlugin"
)

const (
	// OMSwitchNicPluginName name of OMSwitchNicPlugin
	OMSwitchNicPluginName = "OMSwitchNicPlugin"
	// OMSwitchNicStreamName name of OMStream
	OMSwitchNicStreamName = "OMSwitchNicStream"
	// GlobalRankKey key of global ranks
	GlobalRankKey = "globalRankIDs"
	// GlobalOpKey key of global ops
	GlobalOpKey = "globalOps"
	// SwitchNicUUID key of switch nic uuid
	SwitchNicUUID = "switchNicUUID"
	// SwitchJobID key of switch job id
	SwitchJobID = "switchJobID"
	// SwitchOK value of switch ok
	SwitchOK = "switchOK"
	// SwitchFail value of switch fail
	SwitchFail = "switchFail"
)

const (
	// LocalProxyIP local proxy ip
	LocalProxyIP = "127.0.0.1"
	// LocalProxyEnableEnv whether enable local proxy
	LocalProxyEnableEnv = "LOCAL_PROXY_ENABLE"
	// LocalProxyEnableOn local proxy enable value
	LocalProxyEnableOn = "on"
)

const (
	// OMStressTestPluginName name of OMStressTestPlugin
	OMStressTestPluginName = "OMStressTestPlugin"
	// OMStressTestStreamName name of OMStream
	OMStressTestStreamName = "OMStressTestStream"
	// StressTestRankOPStr key of stress test rank op str
	StressTestRankOPStr = "StressTestRankOPStr"
	// StressTestUUID key of stress test uuid
	StressTestUUID = "stressTestUUID"
	// StressTestJobID key of stress test job id
	StressTestJobID = "stressTestJobID"
	// StressTestOK value of stress test ok
	StressTestOK = "0"
	// StressTestExecFail value of stress test exec fail
	StressTestExecFail = "1"
	// StressTestFindFault value of stress test find fault
	StressTestFindFault = "2"
	// StressTestTimeout value of stress test timeout
	StressTestTimeout = "3"
	// StressTestVolRecoverFail voltage recovery failed
	StressTestVolRecoverFail = "4"
)
