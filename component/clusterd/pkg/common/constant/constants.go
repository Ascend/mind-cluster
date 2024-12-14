// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package constant a series of para
package constant

import (
	"time"
)

const (
	// ComponentName is the name of component
	ComponentName = "clusterd"

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
)

const (
	// SeparateFaultStrategy separate fault strategy
	SeparateFaultStrategy = "Separate"
	// SubHealthFaultStrategy subhealth fault strategy
	SubHealthFaultStrategy = "SubHealth"
	// SwitchFault switch fault strategy
	SwitchFault = "switchFault"
)
