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
	// NotifyContinueSuccessEvent notify process continue strategy success
	NotifyContinueSuccessEvent = "notifyContinueSuccessEvent"

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

	// WaitHCCLRoutingConvergenceFail wait hccl routing convergence fail
	WaitHCCLRoutingConvergenceFail = "waitHCCLRoutingConvergenceFail"

	// ProcessPauseFailEvent process not ready
	ProcessPauseFailEvent = "processPauseFail"
	// ContinueTrainFailEvent continue train fail event
	ContinueTrainFailEvent = "continueTrainFail"

	// StartSwitchNic start switch nic
	StartSwitchNic = "startSwitchNic"
	// WaitSwitchNicRecvFaultEvent wait switch nic and receive fault
	WaitSwitchNicRecvFaultEvent = "waitSwitchNicRecvFault"
	// SwitchNicFailEvent switch nic fail
	SwitchNicFailEvent = "switchNicFail"
	// SwitchNicRecvPauseEvent receive pause event before switch nic
	SwitchNicRecvPauseEvent = "switchNicRecvPause"
	// SwitchNicRecvContinueEvent receive continue event after switch nic
	SwitchNicRecvContinueEvent = "switchNicRecvContinue"

	// KillPodAfterRestartProcessEvent kill pod when cant not restart process
	KillPodAfterRestartProcessEvent = "KillPodAfterRestartProcessEvent"

	// StartStressTest start stress test
	StartStressTest = "startStressTest"
	// StressTestRecvPauseEvent receive pause event before stress test
	StressTestRecvPauseEvent = "stressTestRecvPause"
	// StressTestFailEvent stress test fail
	StressTestFailEvent = "stressTestFail"
	// StressTestRecvContinueEvent receive continue event after stress test
	StressTestRecvContinueEvent = "stressTestRecvContinue"

	// NotifyScaleInStrategySuccessEvent notify scale-in strategy success
	NotifyScaleInStrategySuccessEvent = "notifyScaleInStrategySuccessEvent"
	// ScaleInSuccessEvent scale-in success
	ScaleInSuccessEvent = "scaleInSuccessEvent"
	// ScaleOutSuccessEvent scale-out success
	ScaleOutSuccessEvent = "scaleOutSuccessEvent"
	// NeedTryScaleInStrategyEvent try to change strategy to scale-in
	NeedTryScaleInStrategyEvent = "needTryScaleStrategyInEvent"
	// NotifyScaleOutStrategySuccessEvent notify scale-out strategy success
	NotifyScaleOutStrategySuccessEvent = "notifyScaleOutStrategySuccessEvent"
	// NotifyFaultNodesExitSuccessEvent notify fault nodes to exit successfully
	NotifyFaultNodesExitSuccessEvent = "notifyFaultNodesExitSuccessEvent"

	// BeginHotSwitchEvent begin hotswitch
	BeginHotSwitchEvent = "BeginHotSwitchEvent"
	// NewPodTimeoutEvent new pod time out
	NewPodTimeoutEvent = "NewPodTimeoutEvent"
	// NewPodRunningEvent new pod running
	NewPodRunningEvent = "NewPodRunningEvent"
	// MigrationEvent migration event
	MigrationEvent = "MigrationEvent"
	// ExitEvent exit event
	ExitEvent = "ExitEvent"
	// OldPodDeletedEvent old pod deleted
	OldPodDeletedEvent = "OldPodDeletedEvent"
	// RestartSuccessEvent restart success
	RestartSuccessEvent = "RestartSuccessEvent"
	// RestartFaildEvent restart faild
	RestartFaildEvent = "RestartFaildEvent"
)
