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
	// SaveAndExitSignalType save and exit signal type
	SaveAndExitSignalType = "saveAndExit"
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
	// SubHealthyStrategy config in pod group label for subHealthy fault strategy
	SubHealthyStrategy = "subHealthyStrategy"
	// SubHealthyGraceExit strategy name of grace exit
	SubHealthyGraceExit = "graceExit"
	// HealthyState state of Healthy
	HealthyState = "Healthy"
	// UnHealthyState state of unHealthy
	UnHealthyState = "UnHealthy"
	// SubHealthyState state of subHealthy
	SubHealthyState = "SubHealthy"
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
	// ProcessRecoverInit init state before real open process-recover-enable
	ProcessRecoverInit = ""
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
	// RetrySuccess retry success
	RetrySuccess = "retry-success"
	// RetryFailed retry failed
	RetryFailed = "retry-failed"
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

// FaultLevel string describe
const (
	// NotHandleFault not handle fault
	NotHandleFault = "NotHandleFault"
	// RestartRequest restart request
	RestartRequest = "RestartRequest"
	// RestartBusiness restart business
	RestartBusiness = "RestartBusiness"
	// RestartNPU restart NPU
	RestartNPU = "RestartNPU"
	// FreeRestartNPU wait free and restart NPU
	FreeRestartNPU = "FreeRestartNPU"
	// SeparateNPU separate NPU
	SeparateNPU = "SeparateNPU"
	// NormalNPU normal NPU
	NormalNPU = "NormalNPU"
	// NormalNetwork normal network
	NormalNetwork = "NormalNetwork"
	// PreSeparateNPU pre separate NPU
	PreSeparateNPU = "PreSeparateNPU"
	// ManuallySeparateNPU Manually Separate NPU
	ManuallySeparateNPU = "ManuallySeparateNPU"
	// CardUnhealthy fault is caused by card unhealthy
	CardUnhealthy = "CardUnhealthy"
	// CardNetworkUnhealthy  fault is caused by card network unhealthy
	CardNetworkUnhealthy = "CardNetworkUnhealthy"
	// SubHealthFault  sub healthy fault
	SubHealthFault = "SubHealthFault"
)

// About cm keys
const (
	// CmRecoveringSuffix Recovering Suffix
	CmRecoveringSuffix = "-Recovering"
	// CmCardUnhealthySuffix CardUnhealthy Suffix
	CmCardUnhealthySuffix = "-Unhealthy"
	// CmCardNetworkUnhealthySuffix NetworkUnhealthy Suffix
	CmCardNetworkUnhealthySuffix = "-NetworkUnhealthy"
	// CmFaultListSuffix FaultList Suffix
	CmFaultListSuffix = "-Fault"
)

// support device type
const (
	Ascend910           = "Ascend910"
	Ascend310P          = "Ascend310P"
	Ascend310           = "Ascend310"
	UnknownResourceType = "unknown"
)

const (
	// InvalidSuperPodIndex invalid super pod index
	InvalidSuperPodIndex = -2
	// PatchPodTimes patch pod retry times
	PatchPodTimes = 3
	// AllCardId all card id
	AllCardId = "FF"
	// SwitchFaultType is switchFault
	SwitchFaultType = "switchFault"
	// DeviceFaultType is deviceFault
	DeviceFaultType = "deviceFault"
	// NodeFaultType is nodeFault
	NodeFaultType = "nodeFault"
	// TriggerFaultType is TriggerFault
	TriggerFaultType = "TriggerFault"
	// RelationFaultType is RelationFaultCodes
	RelationFaultType = "RelationFaultCodes"
	// TaskFaultKey is fault-type
	TaskFaultKey = "fault-type"
	// Kilo is 1000
	Kilo = 1000
	// FaultCustomizationPath fault customization path
	FaultCustomizationPath = "/home/hwMindX/relationFaultCustomization.json"
	// FaultDurationPath fault duration path
	FaultDurationPath = "/home/hwMindX/faultDuration.json"
)
