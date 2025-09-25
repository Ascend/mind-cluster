// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package constant is grpc common types and functions
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
	// FaultNodesExitSignalType fault nodes exit signal type
	FaultNodesExitSignalType = "faultNodesExit"
	// HotSwitchSignalType hot switch
	HotSwitchSignalType = "hot-switch"
	// WaitStartAgentSignalType pause start train signal type
	WaitStartAgentSignalType = "pauseStartAgent"
	// ContinueStartAgentSignalType continue start train signal type
	ContinueStartAgentSignalType = "continueStartAgent"
)

// recover strategy name
const (
	// RecoverStrategies config in pod group label for supported strategy
	RecoverStrategies = "recover-strategy"
	// ProcessRetryStrategyName strategy name of HBM fault step retry
	ProcessRetryStrategyName = "retry"
	// ProcessRecoverStrategyName strategy name of process online recover
	ProcessRecoverStrategyName = "recover"
	// ProcessRecoverInPlaceStrategyName strategy name of recover in place with only restarting fault processes
	ProcessRecoverInPlaceStrategyName = "recover-in-place"
	// ProcessDumpStrategyName strategy name of save check point
	ProcessDumpStrategyName = "dump"
	// ProcessExitStrategyName strategy name of directly exit
	ProcessExitStrategyName = "exit"
	// ProcessContinueTrain continue train
	ProcessContinueTrain = "continue"
	// ElasticTrainingStrategyName strategy name of elastic-training
	ElasticTrainingStrategyName = "elastic-training"
	// ScaleInStrategyName strategy name of DP level scale-in training
	ScaleInStrategyName = "downgrade"
	// ScaleOutStrategyName strategy name of DP level scale-out recover training
	ScaleOutStrategyName = "upgrade"
	// JobReschedulingStrategyName is the name of job level rescheduling
	JobReschedulingStrategyName = "job-rescheduling"
	// JobReschedulingStrategyKey the key of job rescheduling strategy
	JobReschedulingStrategyKey = "fault-scheduling"
	// JobReschedulingStrategyGraceValue one of job rescheduling strategies' value
	JobReschedulingStrategyGraceValue = "grace"
	// JobReschedulingStrategyForceValue one of job rescheduling strategies' value
	JobReschedulingStrategyForceValue = "force"
	// PodReschedulingStrategyName is the name of pod level rescheduling
	PodReschedulingStrategyName = "pod-rescheduling"
	// PodReschedulingStrategyKey is the key of pod level rescheduling label
	PodReschedulingStrategyKey = "pod-rescheduling"
	// PodReschedulingStrategyOpenValue is the value of pod level rescheduling label that stands open
	PodReschedulingStrategyOpenValue = "on"
	// ProcessMigration	migration , strategy used in hotswitch flow
	ProcessMigration = "migration"
)

const (
	// SubHealthyStrategy config in pod group label for subHealthy fault strategy
	SubHealthyStrategy = "subHealthyStrategy"
	// SubHealthyGraceExit strategy name of grace exit
	SubHealthyGraceExit = "graceExit"
	// SubHealthyIngore strategy name of ignore
	SubHealthyIngore = "ignore"
	// SubHealthyHotSwitch strategy name of hot switch
	SubHealthyHotSwitch = "hotSwitch"
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
	// CheckPodReschedulingTimes check pg change running state for exit strategy
	CheckPodReschedulingTimes = 4
	// SleepSecondBeforeCheckPGRunning check pg state interval
	SleepSecondBeforeCheckPGRunning = 5
	// WaitAgentGetFaultRank wait agent get fault rank interval
	WaitAgentGetFaultRank = 5
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
	// JobFaultDisappearRetryTimes wait fault disappear retry times
	JobFaultDisappearRetryTimes = 5
	// JobFaultCheckPeriod job fault check period
	JobFaultCheckPeriod = 3
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
	// MaxEventChanLen max event chan len
	MaxEventChanLen = 100
	// DumpExit dump exit
	DumpExit = "dump_exit"
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
	// RecoverGrpcProbe grpc probe of Recover service
	RecoverGrpcProbe = "/Recover/HealthCheck"
	// BusinessGrpcReq business grpc request for all service
	BusinessGrpcReq = "BusinessGrpcReq"
)

const (
	// UceFaultType uce fault type
	UceFaultType = "0"
	// HcclFaultType uce fault type
	HcclFaultType = "2"
	// NormalFaultType other uce type
	NormalFaultType = "1"
	// HotResetPolicy hot reset policy
	HotResetPolicy = "reset"
	// RestartPolicy restart process policy
	RestartPolicy = "restart"
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
	// NotHandleFaultLevelStr NotHandle Fault Level Str
	NotHandleFaultLevelStr = "NotHandle"
	// PreSeparateFaultLevelStr PreSeparate Fault Level Str
	PreSeparateFaultLevelStr = "PreSeparate"
	// SeparateFaultLevelStr Separate Fault Level Str
	SeparateFaultLevelStr = "Separate"
	// PreSeparateFault pre-separate fault
	PreSeparateFault = "PreSeparateFault"
	// SeparateFault separate fault
	SeparateFault = "SeparateFault"
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
	UnknownResourceType = "unknown"
)

const (
	// InvalidSuperPodIndex invalid super pod index
	InvalidSuperPodIndex = -2
	// PatchPodTimes patch pod retry times
	PatchPodTimes = 3
	// PatchNodeTimes patch node retry times
	PatchNodeTimes = 3
	// AllCardId all card id
	AllCardId = "FF"
	// SwitchFaultType is switchFault
	SwitchFaultType = "switchFault"
	// DeviceFaultType is deviceFault
	DeviceFaultType = "deviceFault"
	// TaskFaultKey is fault-type
	TaskFaultKey = "fault-type"
	// Kilo is 1000
	Kilo = 1000
	// FaultCustomizationPath fault customization path
	FaultCustomizationPath = "/home/hwMindX/relationFaultCustomization.json"
	// FaultDurationPath fault duration path
	FaultDurationPath = "/home/hwMindX/faultDuration.json"
)

const (
	PtFramework = "pytorch"
	MsFramework = "mindspore"
)

const (
	Success = "success"
	Failed  = "failed"
	Start   = "start"
)

const (
	// CardDropFault is the fault code of card drop fault
	CardDropFault = "40F84E00"
)

const (
	// NodeHealthyStatusKey node healthy status key
	NodeHealthyStatusKey = "NodeHealthyStatus"
	// NodeUnHealthy in this case pod will be rescheduling
	NodeUnHealthy = "UnHealthy"
	// StressTestOK stress test ok
	StressTestOK = "0"
	// StressTestExecFail stress test exec fail
	StressTestExecFail = "1"
	// StressTestFindFault stress test find fault
	StressTestFindFault = "2"
	// StressTestTimeout value of stress test timeout
	StressTestTimeout = "3"
	// StressTestVolRecoverFail voltage recovery failed
	StressTestVolRecoverFail = "4"
)

const (
	// FaultNodesExitAction action to notify fault nodes to exit
	FaultNodesExitAction = "fault_nodes_exit"
	// FaultNodesRestartAction action to notify fault nodes to restart
	FaultNodesRestartAction = "fault_nodes_restart"
	// OnGlobalRankAction on_global_rank action
	OnGlobalRankAction = "on_global_rank"
	// ContinueStartAgent continue start agent action
	ContinueStartAgent = "continue_start_agent"
	// PauseStartAgent pause start agent action
	PauseStartAgent = "pause_start_agent"
	// StopAction stop_train action
	StopAction = "stop_train"
	// PauseTrainAction pause_train action
	PauseTrainAction = "pause_train"
	// ChangeStrategyAction change_strategy action
	ChangeStrategyAction = "change_strategy"
	// DefaultWaitRescheduleTimeout default reschedule timeout before executing arf or dp scale-in strategy
	// (Unit: second)
	DefaultWaitRescheduleTimeout = 270
	// MinWaitRescheduleTimeout min reschedule timeout before executing arf or dp scale-in strategy (Unit: second)
	MinWaitRescheduleTimeout = 30
	// WaitRescheduleTimeoutKey is the key of WaitRescheduleTimeout
	WaitRescheduleTimeoutKey = "wait-reschedule-timeout"
	// DefaultWaitRescheduleTimeoutBeforeDeployStrategy is the waiting pod reschedule timeout when ARF is closed but
	// scale-train and pod/job reschedule strategy is open
	DefaultWaitRescheduleTimeoutBeforeDeployStrategy = 20
	// MindIOWaitTimeKey is the key of wait time before deploy strategy
	MindIOWaitTimeKey = "MINDIO_WAIT_MINDX_TIME"
	// MindIOWaitTimeMax is the max wait time (Unit: second)
	MindIOWaitTimeMax = 3600
	// DifferenceTime is the difference with timeout env
	DifferenceTime = 10
	// RankZeroNodeId is "0"
	RankZeroNodeId = "0"
	// PodRankIndexAnno annotation value is rank index of the pod
	PodRankIndexAnno = "hccl/rankIndex"
)
