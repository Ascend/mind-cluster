// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package common is grpc common types and functions
package common

// RespCode response code for grpc fault service
type RespCode int32

const (
	// OK is success
	OK RespCode = 0
	// SuccessCode when query is fine
	SuccessCode RespCode = 200

	/*
		4xx is client error which is not retryable
	*/

	// UnRegistry jobId unregistered
	UnRegistry RespCode = 400
	// OrderMix state machine rules not support
	OrderMix RespCode = 401
	// JobNotExist jobId not exist
	JobNotExist RespCode = 402
	// ProcessRecoverEnableOff not open the switch of process-recover-enable
	ProcessRecoverEnableOff RespCode = 403
	// ProcessNotReady process not ready
	ProcessNotReady = 404
	// RecoverableRetryError error can up to recover strategy
	RecoverableRetryError = 405
	// UnRecoverableRetryError error can not up to recover strategy
	UnRecoverableRetryError = 406
	// DumpError dump error
	DumpError = 407
	// UnInit not init
	UnInit = 408
	// InvalidReqParam invalid request parameters
	InvalidReqParam = 409
	// InvalidReqRate invalid request rate
	InvalidReqRate = 410
	// OMIsRunning om is running
	OMIsRunning RespCode = 411
	// OMParamInvalid param is invalid
	OMParamInvalid RespCode = 413
	// UnRecoverTrainError unrecoverable training errors
	UnRecoverTrainError = 499
	// RateLimitedCode limit rate
	RateLimitedCode = 429
	// ClientError common client error
	ClientError RespCode = 499

	/*
		5xx is server inner error which is retryable
	*/

	// OutOfMaxServeJobs out of max serve jobs number
	OutOfMaxServeJobs RespCode = 500
	// OperateConfigMapError operate config map error
	OperateConfigMapError RespCode = 501
	// OperatePodGroupError opearate pod group error
	OperatePodGroupError RespCode = 502
	// ScheduleTimeout job/pod schedule timeout
	ScheduleTimeout RespCode = 503
	// SignalQueueBusy signal queue busy
	SignalQueueBusy RespCode = 504
	// EventQueueBusy event queue busy
	EventQueueBusy RespCode = 505
	// ControllerEventCancel controller event cancel
	ControllerEventCancel RespCode = 506
	// WaitReportTimeout wait client report timeout
	WaitReportTimeout RespCode = 507
	// WaitPlatStrategyTimeout wait plat strategy error
	WaitPlatStrategyTimeout RespCode = 508
	// WriteConfirmFaultOrWaitPlatResultFault write confirm fault and wait result fault
	WriteConfirmFaultOrWaitPlatResultFault RespCode = 509
	// HCCLRoutingConvergenceFail wait client report timeout
	HCCLRoutingConvergenceFail RespCode = 510
	// ServerInnerError server common error
	ServerInnerError RespCode = 599
)
