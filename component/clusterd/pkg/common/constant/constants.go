// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package constant a series of para
package constant

import (
	"math"
	"time"
)

const (
	// ComponentName is the name of component
	ComponentName = "clusterd"

	// MaxLogLineLength max log line length
	MaxLogLineLength = 2047

	// RetryTime is the retry time loading configmap
	RetryTime = 3
	// RetrySleepTime is the sleep time retry loading configmap
	RetrySleepTime = 50 * time.Millisecond

	// MaxGRPCRecvMsgSize 4MB
	MaxGRPCRecvMsgSize = 4 * 1024 * 1024
	// MaxGRPCConcurrentStreams limit on the number of concurrent streams to each ServerTransport.
	MaxGRPCConcurrentStreams = 64
	// MaxConcurrentLimit limit over listener
	MaxConcurrentLimit = 1024
	// MaxIPConnectionLimit limit over ip
	MaxIPConnectionLimit = 512
	// CacheSize cache for ip
	CacheSize = 1024

	// GrpcPort is the grpc port
	GrpcPort = ":8899"
	// JobErrCount err count return
	JobErrCount = 3

	// DefaultNamespace represents the default value of namespace
	DefaultNamespace = "default"
	// TestName represents the default value of name
	TestName = "test"
	// NoResourceOnServer represents no resources on the server
	NoResourceOnServer = "the server could not find the requested resource"

	// CheckFaultGapSecond check fault gap seconds
	CheckFaultGapSecond = 10
	// JobRefKind reference kind is Job
	JobRefKind = "Job"
	// AscendJobRefKind reference kind is AscendJob
	AscendJobRefKind = "AscendJob"
	// MaxSupportNodeNum max support node num
	MaxSupportNodeNum = 5000
	// MaxSupportJobNum max support job num
	MaxSupportJobNum = 10000
	// MaxCmQueueLen max cm queue len support
	MaxCmQueueLen = 5
	// DefaultLogLevel default log level
	DefaultLogLevel = 0
	// MaxNotifyChanLen max support notify chan
	MaxNotifyChanLen = 1000
)

// fault code const
const (
	UceFaultCode      = "80E01801"
	AicFaultCode      = "80C98009"
	AivFaultCode      = "80CB8009"
	LinkDownFaultCode = "81078603"
	DevCqeFaultCode   = "8C1F8608"
	HostCqeFaultCode  = "4C1F8608"
)

// fault processor const
const (
	JobNotRecover               = int64(math.MaxInt64) // Cannot be used for calculation, only for comparison.
	JobNotRecoverComplete       = int64(math.MaxInt64) // Cannot be used for calculation, only for comparison.
	DeviceNotFault              = int64(math.MaxInt64) // Cannot be used for calculation, only for comparison.
	DiagnosisAccompanyTimeout   = 5 * 1000
	JobReportRecoverTimeout     = 10 * 1000
	JobReportInfoExpiredTimeout = 10 * 1000
	JobReportCompleteTimeout    = 30 * 1000
	FaultCenterProcessPeriod    = 3 * 1000
	MaxFaultCenterSubscriber    = 10
	UnknownFaultTime            = -1
)

// fault center
const (
	AllProcessType = iota
	DeviceProcessType
	NodeProcessType
	SwitchProcessType
)

const (
	// SeparateFaultStrategy separate fault strategy
	SeparateFaultStrategy = "Separate"
	// SubHealthFaultStrategy subhealth fault strategy
	SubHealthFaultStrategy = "SubHealth"
	// SwitchFault switch fault strategy
	SwitchFault = "switchFault"
	// UnHealthy
	UnHealthy = "UnHealthy"
)

// public fault assertion
const (
	// AssertionRecover recover fault assertion
	AssertionRecover = "recover"
	// AssertionOccur occur fault assertion
	AssertionOccur = "occur"
)

// public fault type
const (
	FaultTypeNPU     = "NPU"
	FaultTypeNode    = "Node"
	FaultTypeNetwork = "Network"
	FaultTypeStorage = "Storage"
	PublicFaultType  = "PublicFault"
)

// public fault file path and name
const (
	PubFaultCodeFilePath      = "/home/hwMindX/publicFaultConfiguration.json"
	PubFaultCodeFileName      = "publicFaultConfiguration.json"
	PubFaultCustomizationPath = "/user1/mindx-dl/clusterd/publicCustomization.json"
	PubFaultCustomizationName = "publicCustomization.json"
)

const (
	// NPUPreName npu pre name
	NPUPreName = "huawei.com/Ascend"
)

// jobStc notify msg
const (
	JobInfoAdd       = "JobInfoAdd"
	JobInfoUpdate    = "JobInfoUpdate"
	JobInfoPreDelete = "JobPreDelete"
	JobInfoDelete    = "JobDelete"
)

const (
	// MaxFaultNum max detailed fault number form statistic
	MaxFaultNum = 4500
)

const (
	// MindIeJobIdLabelKey mindie job id label key
	MindIeJobIdLabelKey = "jobID"
	// MindIeAppTypeLabelKey mindie job type label key
	MindIeAppTypeLabelKey = "app"

	// ControllerAppType controller app type
	ControllerAppType = "mindie-ms-controller"
	// CoordinatorAppType coordinator app type
	CoordinatorAppType = "mindie-ms-coordinator"
	// ServerAppType server app type
	ServerAppType = "mindie-ms-server"
	// MetaDataName metadata name
	MetaDataName = "metadata.name"
	// StatusRankTableCompleted is the complete rankTable status
	StatusRankTableCompleted = "completed"
	// RankTableVersion is the rank table version
	RankTableVersion = "1.0"
	// MaxRetryTime max retry time
	MaxRetryTime = 3
	// QueueInitDelay queue init delay
	QueueInitDelay = 2 * time.Second
	// QueueMaxDelay queue max delay
	QueueMaxDelay = 16 * time.Second
	// GroupId0 group id 0
	GroupId0 = "0"
	// GroupId1 group id 1
	GroupId1 = "1"
	// GroupId2 group id 2
	GroupId2 = "2"
	// GroupIdOffset group id offset
	GroupIdOffset = 2
	// ServerCountGroupId0Or1 server count group id 0 or 1
	ServerCountGroupId0Or1 = "1"
)
