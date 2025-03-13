// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package common is grpc common types and functions
package common

// event type
const (
	// FaultOccurEvent fault occur event
	FaultOccurEvent = "faultOccur"

	// NotifyFinishEvent notify finish event
	NotifyFinishEvent = "notifyFinish"
	// NotifySuccessEvent notify success event
	NotifySuccessEvent = "notifySuccess"
	// NotifyFailEvent notify fail event
	NotifyFailEvent = "notifyFail"
	// NotifyRetrySuccessEvent notify retry strategy success
	NotifyRetrySuccessEvent = "notifyRetryStrategySuccess"
	// NotifyRecoverSuccessEvent notify recover strategy success
	NotifyRecoverSuccessEvent = "notifyRecoverStrategySuccess"
	// NotifyDumpSuccessEvent notify dump strategy success
	NotifyDumpSuccessEvent = "notifyDumpStrategySuccess"
	// NotifyExitSuccessEvent notify exit strategy success
	NotifyExitSuccessEvent = "notifyExitStrategySuccess"

	// ReceiveReportEvent receive report
	ReceiveReportEvent = "receiveReport"
	// ReportTimeoutEvent receive timeout
	ReportTimeoutEvent = "receiveTimeout"
	// ProcessNotReadyEvent process not ready
	ProcessNotReadyEvent = "processNotReadyEvent"

	// FaultFlushFinishedEvent fault flush finish event
	FaultFlushFinishedEvent = "flushFinished"

	// DumpForFaultEvent dump for fault
	DumpForFaultEvent = "dumpForFault"

	// ScheduleTimeoutEvent schedule timeout
	ScheduleTimeoutEvent = "scheduleTimeout"
	// ScheduleSuccessEvent schedule success
	ScheduleSuccessEvent = "scheduleSuccess"

	// ClearConfigMapFaultFailEvent clear config map fail
	ClearConfigMapFaultFailEvent = "clearFail"
	// ClearConfigMapFaultSuccessEvent clear config map success
	ClearConfigMapFaultSuccessEvent = "clearSuccess"
	// ClearConfigMapFaultFinishEvent clear config map finish
	ClearConfigMapFaultFinishEvent = "clearFinish"

	// RecoverSuccessEvent recover success
	RecoverSuccessEvent = "recoverSuccess"
	// RecoverFailEvent recover fail
	RecoverFailEvent = "recoverFail"
	// RecoverableRetryErrorEvent recoverable retry error
	RecoverableRetryErrorEvent = "recoverableRetryError"
	// UnRecoverableRetryErrorEvent unrecoverable retry error
	UnRecoverableRetryErrorEvent = "unRecoverableRetryError"

	// FinishKillPodEvent finish kill pod
	FinishKillPodEvent = "finishKillPod"

	// CheckResultFinishEvent check result finish
	CheckResultFinishEvent = "checkFinish"
	// RestartProcessFinishEvent restart process finish event
	RestartProcessFinishEvent = "restartFinish"

	// ChangeProcessSchedulingModePauseErrorEvent change process-scheduling pause error
	ChangeProcessSchedulingModePauseErrorEvent = "changeSwitchPauseError"
	// ChangeProcessSchedulingModeEnableErrorEvent change process-scheduling on error
	ChangeProcessSchedulingModeEnableErrorEvent = "changeSwitchEnableError"

	// FinishEvent finish machine state
	FinishEvent = "finish"

	// WaitPlatStrategyTimeoutEvent wait plat strategy timeout
	WaitPlatStrategyTimeoutEvent = "waitPlatStrategyTimeout"
	// WriteConfirmFaultOrWaitResultFaultTimeoutEvent write confirm fault or wait result fault error
	WriteConfirmFaultOrWaitResultFaultTimeoutEvent = "writeConfirmFaultOrWaitResultFaultTimeout"
	// WaitRankTableReadyTimeoutEvent wait rank table ready timeout
	WaitRankTableReadyTimeoutEvent = "waitRankTableTimeout"
)
