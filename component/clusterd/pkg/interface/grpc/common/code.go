// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package common is grpc common types and functions
package common

// response code for grpc fault service
type RespCode int32

const (
	OK RespCode = 0

	/*
		4xx is client error which is not retryable
	*/
	// UnRegistry jobId unregistered
	UnRegistry RespCode = 400
	// UnRegistry state machine rules not support
	OrderMix RespCode = 401
	// JobNotExist jobId not exist
	JobNotExist RespCode = 402
	// ProcessRescheduleOff not open the switch of process-rescheduling
	ProcessRescheduleOff RespCode = 403
	// StopDeviceError stop device error
	StopDeviceError RespCode = 405
	// CleanDeviceError clean device error
	CleanDeviceError RespCode = 406
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
	// ServerInnerError server common error
	ServerInnerError RespCode = 599
)
