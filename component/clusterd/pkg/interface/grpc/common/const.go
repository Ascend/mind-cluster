// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package common is grpc common types and functions
package common

const (
	// StopTrainSignalType stop train signal type
	StopTrainSignalType = "stopTrain"
	// GlobalFaultSignalType global fault ranks signal type
	GlobalFaultSignalType = "globalFault"
	// ChangeStrategySignalType change strategy signal type
	ChangeStrategySignalType = "changeStrategy"

	// ProcessArfStrategy online process recover strategy
	ProcessArfStrategy = "arf"
	// ProcessDumpStrategy save checkpoint and grace restart process recover strategy
	ProcessDumpStrategy = "dump"
	// ProcessExitStrategy directly restart process recover strategy
	ProcessExitStrategy = "exit"

	// ArfRecoverLevel arf strategy level
	ArfRecoverLevel = 0
	// DumpRecoverLevel dump strategy level
	DumpRecoverLevel = 1
	// ExitRecoverLevel exit strategy level
	ExitRecoverLevel = 2
	// RandLength random number length
	RandLength = 32
)

// MachineState is recover state
type MachineState int

const (
	// INIT init state
	INIT MachineState = iota

	// SentStopTrain sent stop train
	SentStopTrain
	// ReceiveStopFinish receive stop finish
	ReceiveStopFinish
	// SentGlobalFault sent global fault ranks
	SentGlobalFault
	// ReceiveSupportStrategy sent decide strategy
	ReceiveSupportStrategy
	// StartListenSchedule start listen schedule result
	StartListenSchedule
	// GetJobScheduleResult already get job schedule result
	GetJobScheduleResult
	// ListenCheckPointSave listen checkpoint save result
	ListenCheckPointSave
	// ListenOnlineRecoverStatus listen online recover status
	ListenOnlineRecoverStatus
	// ReceiveRecoverStatus receive process recover status
	ReceiveRecoverStatus

	// ReceiveStepRetry receive step retry request
	ReceiveStepRetry
	// ReceiveStepRetryStatus receive step retry status
	ReceiveStepRetryStatus

	// StartPodReschedule start pod reschedule
	StartPodReschedule
)

// RecoverMode recover mode.
type RecoverMode int

const (
	// InitMode init mode
	InitMode RecoverMode = iota
	// HbmFaultStepRetryMode hbm step retry mode
	HbmFaultStepRetryMode
	// ProcessFaultRecoverMode process recover mode
	ProcessFaultRecoverMode
	// PodRescheduleMode pod reschedule mode
	PodRescheduleMode
)

const (
	// ResetInfoDir dir for reset info
	ResetInfoDir = "/user/restore/reset/"
	// ResetInfoCMNamePrefix for reset configmap name prefix
	ResetInfoCMNamePrefix = "reset-config-"
	// ResetInfoCMDataKey for reset configmap data key
	ResetInfoCMDataKey = "reset.json"
	// ResetInfoCMCheckCodeKey for reset configmap checkcode key
	ResetInfoCMCheckCodeKey = "checkCode"
	// ResetTaskNameKey for obtain the reset task name
	ResetTaskNameKey = "volcano.sh/job-name"
	// ResetTaskNameKeyInLabel for obtain the reset task name when using operator
	ResetTaskNameKeyInLabel = "training.kubeflow.org/job-name"
)

const (
	// FaultRankStatus rank status is fault
	FaultRankStatus = "fault"
	// RestartAllProcess flush reset.json and restart all process
	RestartAllProcess = "restartAllProcess"
	// PodReschedulingLabel the pod rescheduling label of pg
	PodReschedulingLabel = "pod-rescheduling"
	// ProcessReschedulingLabel the process rescheduling label of pg
	ProcessReschedulingLabel = "process-rescheduling"
	// ProcessReschedulingEnable open process rescheduling
	ProcessReschedulingEnable = "on"
	// ProcessReschedulingPause close process rescheduling temporarily
	ProcessReschedulingPause = "pause"
)

const (
	// StateTimeoutSecond state time out second
	StateTimeoutSecond = 600
	// CheckPGRunningRetryTimes check pg change running state retry times
	CheckPGRunningRetryTimes = 54
	// SleepSecondBeforeCheckPGRunning check pg state interval
	SleepSecondBeforeCheckPGRunning = 5
	// WriteResetInfoRetryTimes retry set reset configmap
	WriteResetInfoRetryTimes = 3
	// WaitProcessRestart sleep 60 second
	WaitProcessRestart = 60
	// ProcessRecoverStrategy pg label control process recover continue
	ProcessRecoverStrategy = "ProcessRecoverStrategy"
	// ProcessConfirmFaultKey pg annotation key store fault rank
	ProcessConfirmFaultKey = "ProcessConfirmFault"
	// ProcessResultFaultKey pg annotation key store final fault rank
	ProcessResultFaultKey = "ProcessResultFault"
	// ProcessRecoverStatusKey process recover status
	ProcessRecoverStatusKey = "ProcessRecoverStatus"
	// RankTableReadyKey pg annotation key store whether rank table ready
	RankTableReadyKey = "RankTableReady"
	// CheckPeriod sleep when process not ready
	CheckPeriod = 3
	// ProcessControlTimeout wait process annotation until timeout
	ProcessControlTimeout = 300

	// PlatFormArfStrategyName plat arf strategy name
	PlatFormArfStrategyName = "recover"
	// PlatFormDumpStrategyName plat dump strategy name
	PlatFormDumpStrategyName = "dump"
	// PlatFormExitStrategyName plat exit strategy name
	PlatFormExitStrategyName = "none"

	// RecoverSuccess process recover success
	RecoverSuccess = "recover-success"
	// RecoverFailed process recover failed
	RecoverFailed = "recover-failed"
	// DumpSuccess save ckpt success
	DumpSuccess = "dump-success"
	// DumpFailed save ckpt fail
	DumpFailed = "dump-failed"
	// ExitCompleted exit strategy finish
	ExitCompleted = "exit-completed"
)

const (
	// UnknownEventId unknown event id
	UnknownEventId = "unknown_event_id"
	// GetPodGroupTimes get pod group times
	GetPodGroupTimes = 3
	// UpdatePodGroupTimes get pod group times
	UpdatePodGroupTimes = 3
	// MaxChangeStrategyTimes max changeStrategy Times
	MaxChangeStrategyTimes = 2
	// MaxServeJobs max serve job num for fault recover
	MaxServeJobs = 10000
	// QpsLimit max qps for grpc service
	QpsLimit = 1000
)
