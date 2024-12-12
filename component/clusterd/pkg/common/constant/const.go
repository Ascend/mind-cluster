// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package common is grpc common types and functions
package constant

// process signal type
const (
	// KillMasterSignalType kill master agent
	KillMasterSignalType = "killMaster"
	// StopTrainSignalType stop train signal type
	StopTrainSignalType = "stopTrain"
	// GlobalFaultSignalType global fault ranks signal type
	GlobalFaultSignalType = "globalFault"
	// ChangeStrategySignalType change strategy signal type
	ChangeStrategySignalType = "changeStrategy"
	// KeepAliveSignalType keep alive signal type
	KeepAliveSignalType = "keep-alive"
)

// recover strategy name
const (
	// RecoverStrategies config in pod group label for supported strategy
	RecoverStrategies = "recover-strategy"
	// ProcessRetryStrategyName strategy name of HBM fault step retry
	ProcessRetryStrategyName = "retry"
	// ProcessRecoverStrategyName strategy name of process online recover
	ProcessRecoverStrategyName = "recover"
	// ProcessDumpStrategyName strategy name of save check point
	ProcessDumpStrategyName = "dump"
	// ProcessExitStrategyName strategy name of directly exit
	ProcessExitStrategyName = "exit"
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
	// ProcessRecoverEnableLabel the process recover label of pg
	ProcessRecoverEnableLabel = "process-recover-enable"
	// ProcessRecoverEnable open process recover
	ProcessRecoverEnable = "on"
	// ProcessRecoverPause close process recover temporarily
	ProcessRecoverPause = "pause"
)

// write reset configmap operation
const (
	// RestartAllProcessOperation add reset.json retryTimes which trigger agent restart all process
	RestartAllProcessOperation = "restartAllProcess"
	// ClearOperation reset resetConfigMap
	ClearOperation = "clear"
	// NotifyFaultListOperation write fault list to reset.json
	NotifyFaultListOperation = "fault"
	// NotifyFaultFlushingOperation notify agent fault occur and wait fault flush finished
	NotifyFaultFlushingOperation = "notifyFaultFlushing"
)

const (
	// MaxUuidRandomLength max uuid random length
	MaxUuidRandomLength = 32
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
	// GetPodGroupTimes get pod group times
	GetPodGroupTimes = 3
	// UpdatePodGroupTimes get pod group times
	UpdatePodGroupTimes = 3
	// MaxServeJobs max serve job num for fault recover
	MaxServeJobs = 10000
	// QpsLimit max qps for grpc service
	QpsLimit = 1000
)

const (
	// UceFaultType uce fault type
	UceFaultType = "0"
	// NormalFaultType other uce type
	NormalFaultType = "1"
	// HotResetPolicy hot reset policy
	HotResetPolicy = "reset"
)
