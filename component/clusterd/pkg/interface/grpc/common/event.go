// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package common is grpc common types and functions
package common

// event type
const (
	FaultOccurEvent = "faultOccur"

	NotifyFinishEvent         = "notifyFinish"
	NotifySuccessEvent        = "notifySuccess"
	NotifyFailEvent           = "notifyFail"
	NotifyRetrySuccessEvent   = "notifyRetryStrategySuccess"
	NotifyRecoverSuccessEvent = "notifyRecoverStrategySuccess"
	NotifyDumpSuccessEvent    = "notifyDumpStrategySuccess"
	NotifyExitSuccessEvent    = "notifyExitStrategySuccess"

	ReceiveReportEvent = "receiveReport"
	ReportTimeoutEvent = "receiveTimeout"

	FaultFlushFinishedEvent = "flushFinished"

	ScheduleTimeoutEvent = "scheduleTimeout"
	ScheduleSuccessEvent = "scheduleSuccess"

	ClearConfigMapFaultFailEvent    = "clearFail"
	ClearConfigMapFaultSuccessEvent = "clearSuccess"
	ClearConfigMapFaultFinishEvent  = "clearFinish"

	RecoverSuccessEvent  = "recoverSuccess"
	RecoverFailEvent     = "recoverFail"
	DeviceCleanFailEvent = "deviceCleanFail"

	CheckResultFinishEvent    = "checkFinish"
	RestartProcessFinishEvent = "restartFinish"

	ChangeProcessSchedulingModePauseErrorEvent  = "changeSwitchPauseError"
	ChangeProcessSchedulingModeEnableErrorEvent = "changeSwitchEnableError"

	FinishEvent = "finish"
)
