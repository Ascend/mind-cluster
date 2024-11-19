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
	MaxLogLineLength = 1023

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
	// MaxSupportJobNum
	MaxSupportJobNum = 10000
)

// fault code const
const (
	UCE_FAULT_CODE = "80E01801"
	AIC_FAULT_CODE = "80C98009"
	AIV_FAULT_CODE = "80CB8009"
)

// fault processor const
const (
	JobNotRecover               = int64(math.MaxInt64) // Cannot be used for calculation, only for comparison.
	JobNotRecoverComplete       = int64(math.MaxInt64) // Cannot be used for calculation, only for comparison.
	DeviceNotFault              = int64(math.MaxInt64) // Cannot be used for calculation, only for comparison.
	DiagnosisAccompanyTimeout   = 5 * 1000
	JobReportRecoverTimeout     = 10 * 1000
	JobReportCompleteTimeout    = 30 * 1000
	FaultCenterProcessPeriod    = 3 * 1000
	MAX_FAULT_CENTER_SUBSCRIBER = 10
)

// fault center
const (
	ALL_FAULT = iota
	DEVICE_FAULT
	NODE_FAULT
	SWITCH_FAULT
)
