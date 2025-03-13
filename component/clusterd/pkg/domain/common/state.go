// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package common is grpc common types and functions
package common

// state name of state MachineState
const (
	// InitState init state
	InitState = "INIT"

	// NotifyWaitFaultFlushingState wait notify agent wait fault flushing
	NotifyWaitFaultFlushingState = "NotifyWaitFaultFlushingState"

	// NotifyStopTrainState notify process controller stop train
	NotifyStopTrainState = "NotifyStopTrainState"

	// WaitReportStopCompleteState wait report stop complete
	WaitReportStopCompleteState = "WaitReportStopCompleteState"

	// WaitFaultFlushFinishedState wait fault flush finished
	WaitFaultFlushFinishedState = "WaitFaultFlushFinishedState"

	// NotifyGlobalFaultState notify fault flush finished
	NotifyGlobalFaultState = "NotifyGlobalFaultState"

	// WaitReportRecoverStrategyState wait agent report supported recover strategies
	WaitReportRecoverStrategyState = "WaitReportRecoverStrategyState"

	// NotifyDecidedStrategyState notify process controller use decided strategy recover training
	NotifyDecidedStrategyState = "NotifyDecidedStrategyState"

	// WaitReportStepRetryStatusState wait report step retry recover result
	WaitReportStepRetryStatusState = "WaitReportStepRetryStatusState"

	// WaitReportProcessRecoverStatusState wait report online process recover result
	WaitReportProcessRecoverStatusState = "WaitReportProcessRecoverStatusState"

	// WaitReportDumpStatusState wait report check point save status
	WaitReportDumpStatusState = "WaitReportDumpStatusState"

	// WaitProcessRestartResultState wait process restart result
	WaitProcessRestartResultState = "WaitProcessRestartResultState"

	// FaultClearState clear reset configmap fault list
	FaultClearState = "FaultClearState"

	// FaultRetryState fault retry for volcano
	FaultRetryState = "FaultRetryState"

	// CheckRecoverResultState write recover result
	CheckRecoverResultState = "CheckRecoverResultState"

	// ListenScheduleResultState check schedule result
	ListenScheduleResultState = "ListenScheduleResultState"

	// NotifyRestartAllProcessState notify restart all process
	NotifyRestartAllProcessState = "NotifyRestartAllProcessState"

	// WaitRestartAllProcessState notify restart all process
	WaitRestartAllProcessState = "WaitRestartAllProcessState"

	// NotifyKillJobState send kill job signal to agent for job reschedule
	NotifyKillJobState = "NotifyKillJobState"

	// KillPodForUnrecoverableRetryState KillPodForUnrecoverableRetryError kill pod for unrecoverable retry error
	KillPodForUnrecoverableRetryState = "KillPodForUnrecoverableRetryState"

	// NotifyDumpState notify dump
	NotifyDumpState = "NotifyDumpState"
)
