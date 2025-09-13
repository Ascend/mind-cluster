// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package constant a series of para
package constant

import (
	"math"
	"time"
)

const (
	// MaxLogLineLength max log line length
	MaxLogLineLength = 2047

	// RetryTime is the retry time loading configmap
	RetryTime = 3
	// RetrySleepTime is the sleep time retry loading configmap
	RetrySleepTime = 50 * time.Millisecond

	// MaxGRPCRecvMsgSize 4MB
	MaxGRPCRecvMsgSize = 4 * 1024 * 1024
	// MaxGRPCSendMsgSize 8MB
	MaxGRPCSendMsgSize = 8 * 1024 * 1024
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

	// DefaultNamespace represents the default value of namespace
	DefaultNamespace = "default"
	// TestName represents the default value of name
	TestName = "test"
	// NoResourceOnServer represents no resources on the server
	NoResourceOnServer = "the server could not find the requested resource"

	// MaxSupportNodeNum max support node num
	MaxSupportNodeNum = 16000
	// MaxCmQueueLen max cm queue len support
	MaxCmQueueLen = 5
	// DefaultLogLevel default log level
	DefaultLogLevel = 0
	// MaxNotifyChanLen max support notify chan
	MaxNotifyChanLen = 1000

	// InvalidResult invalid result
	InvalidResult = -1
)

// fault code const
const (
	UceFaultCode            = "80E01801"
	AicFaultCode            = "80C98009"
	AivFaultCode            = "80CB8009"
	LinkDownFaultCode       = "81078603"
	SwitchLinkDownFaultCode = "[0x08520003,na,L2,na]"
	DevCqeFaultCode         = "8C1F8608"
	HostCqeFaultCode        = "4C1F8608"
	HcclRetryFaultCode      = "8C1F860B"
	StressTestHighLevelCode = "80818C05"
	StressTestLowLevelCode  = "80818C06"
)

// fault processor const
const (
	JobNotRecover               = int64(math.MaxInt64) // Cannot be used for calculation, only for comparison.
	JobNotRecoverComplete       = int64(math.MaxInt64) // Cannot be used for calculation, only for comparison.
	DeviceNotFault              = int64(math.MaxInt64) // Cannot be used for calculation, only for comparison.
	JobShouldReportFault        = int64(math.MaxInt64) // Cannot be used for calculation, only for comparison.
	DiagnosisAccompanyTimeout   = 5 * 1000
	JobReportRecoverTimeout     = 10 * 1000
	JobReportInfoExpiredTimeout = 10 * 1000
	JobReportCompleteTimeout    = 30 * 1000
	JobRestartInPlaceTimeout    = 60 * 1000
	MaxFaultCenterSubscriber    = 10
	UnknownFaultTime            = -1
	DeviceRetryFault            = "retry"
	DeviceNormalFault           = "normal"
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
)

// public fault assertion
const (
	// AssertionRecover recover fault assertion
	AssertionRecover = "recover"
	// AssertionOccur occur fault assertion
	AssertionOccur = "occur"
	// AssertionOnce once fault assertion
	AssertionOnce = "once"
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

// jobStc notify msg
const (
	PGAdd         = "PGAdd"
	PGUpdate      = "PGUpdate"
	PGDelete      = "PGDelete"
	JobInfoDelete = "JobInfoDelete"

	ACJobCreate = "ACJobCreate"
	ACJobUpdate = "ACJobUpdate"
	ACJobDelete = "ACJobDelete"

	VCJobCreate = "VCJobCreate"
	VCJobDelete = "VCJobDelete"
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
	// StatusRankTableCompleted is the complete rankTable status
	StatusRankTableCompleted = "completed"
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
)

const (
	// RankTableDataType rankTable data type
	RankTableDataType = "rankTable"
	// FaultMsgDataType faultMsg data type
	FaultMsgDataType = "faultMsg"
	// ProfilingDataType profiling data type
	ProfilingDataType = "profiling"
	// FaultTypeSwitch Types of switch faults
	FaultTypeSwitch = "Switch"
	// HealthyLevel int value of Healthy state
	HealthyLevel = 0
	// SubHealthyLevel int value of SubHealthy state
	SubHealthyLevel = 1
	// UnHealthyLevel int value of UnHealthy state
	UnHealthyLevel = 2
	// SignalTypeNormal signal type that fault can be ignored or no fault has occurred
	SignalTypeNormal = "normal"
	// SignalTypeFault signal type that fault occur
	SignalTypeFault = "fault"
	// Comma comma
	Comma = ","
	// Minus minus
	Minus = "-"
	// EmptyDeviceId device id for node or switch fault
	EmptyDeviceId = "-1"
	// FormatBase The base number used to convert int to string
	FormatBase = 10
	// DefaultJobId default job id for cluster dimension
	DefaultJobId = "-1"
)

// ras feature const
const (
	// RasNetDetectOff ras net fault detect status is off
	RasNetDetectOff = 0
	// RasNetDetectOn ras net fault detect status is on
	RasNetDetectOn = 1

	// RasGlobalKey one of the key of pingmesh config
	RasGlobalKey = "global"
	// RasNetDetectOnStr the string of detect on
	RasNetDetectOnStr = "on"
	// RasNetDetectOffStr the string of detect off
	RasNetDetectOffStr = "off"

	// PingMeshCMNamespace is the namespace of pingmesh configmap
	PingMeshCMNamespace = "cluster-system"
	// PingMeshConfigCm is the name of pingmesh configmap
	PingMeshConfigCm = "pingmesh-config"
)

const (
	// HCCLRoutingConvergenceTimeout is the timeout for HCCL routing convergence
	HCCLRoutingConvergenceTimeout = 3
	// StepRetryTimeout is the timeout for step retry
	StepRetryTimeout = 30
	// HCCLStepRetryTimeout is the timeout for HCCL step retry
	HCCLStepRetryTimeout = 1000 * 60
)

const (
	// ReleaseTimeOut release timeout
	ReleaseTimeOut = 10 * time.Second
)
