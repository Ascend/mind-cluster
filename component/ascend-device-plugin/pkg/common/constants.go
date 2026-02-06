/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package common a series of common function
package common

import (
	"time"

	"ascend-common/api"
)

const (
	// Component component name
	Component = "device-plugin"
	// MaxBackups log file max backup
	MaxBackups = 30
	// MaxAge the log file last time
	MaxAge = 7

	// DevA3 is A3 device type
	DevA3 = "A3"

	// KubeEnvMaxLength k8s env name max length
	KubeEnvMaxLength = 230
	// PodNameMaxLength pod name max length
	PodNameMaxLength = 253
	// PodNameSpaceMaxLength pod name space max length
	PodNameSpaceMaxLength = 63
	// MaxPodLimit max pod num
	MaxPodLimit = 10000
	// MaxContainerLimit max container num
	MaxContainerLimit = 300000
	// RetryUpdateCount is max number of retry resource update
	RetryUpdateCount = 3
	// GetPodFromInformerTime is max number of get pod from informer
	GetPodFromInformerTime = 3
	// MaxDeviceNameLen max length of device name, like "Ascend310P-4c.3cpu-100-0"
	MaxDeviceNameLen = 50
	// MaxGRPCRecvMsgSize 4MB
	MaxGRPCRecvMsgSize = 4 * 1024 * 1024
	// MaxGRPCConcurrentStreams limit on the number of concurrent streams to each ServerTransport.
	MaxGRPCConcurrentStreams = 64
	// MaxConcurrentLimit limit over listener
	MaxConcurrentLimit = 64
	// MaxIPConnectionLimit limit over ip
	MaxIPConnectionLimit = 64
	// CacheSize cache for ip
	CacheSize = 128
	// MaxVirtualDeviceNum max num of virtual device
	MaxVirtualDeviceNum = 1024
	// CMDataMaxLength configMap max data size 1MB
	CMDataMaxLength = 1024 * 1024
	// PodAnnotationMaxLength pod annotation max data length 1MB
	PodAnnotationMaxLength = 1024 * 1024
	// UpdatePodWaitTime default try update pod wait time 200 millisecond
	UpdatePodWaitTime = 200

	// DeviceInfoCMNamePrefix device info configmap name prefix
	DeviceInfoCMNamePrefix = "mindx-dl-deviceinfo-"
	// DeviceInfoCMManuallySeparateNPUKey for deviceinfo configmap ManuallySeparateNPU key
	DeviceInfoCMManuallySeparateNPUKey = "ManuallySeparateNPU"
	// DeviceInfoCmUpgradeFaultReasonKey for deviceinfo configmap UpgradeFaultReason key
	DeviceInfoCmUpgradeFaultReasonKey = "UpgradeFaultReason"
	// EventReleaseUpgradeFault the event reason of release upgrade fault
	EventReleaseUpgradeFault = "EventReleaseUpgradeFault"
	// DescriptionKey for deviceinfo configmap Description key
	DescriptionKey = "Description"
	// DescriptionValue for deviceinfo configmap Description value
	DescriptionValue = "The field[xxx.com/xxx] is sunset. Volcano has the ability to maintain the usage status of " +
		"chips on its own, and can temporarily use this field by configuring the Volcano parameter 'self-maintain-available-card'"
	// SlowNodeNoticeCMName the name for slow node notice configmap
	SlowNodeNoticeCMName = "steptime-dtpgroup"

	// CmConsumerValue the value only for true
	CmConsumerValue = "true"

	runtimeEnvNum = 3
	// AscendVisibleDevicesEnv visible devices env
	AscendVisibleDevicesEnv = api.AscendVisibleDevicesEnv
	// ascendRuntimeOptionsEnv virtual runtime option env
	ascendRuntimeOptionsEnv = api.AscendRuntimeOptionsEnv
	// ascendAllowLinkEnv a500a2 need mount softlink
	ascendAllowLinkEnv = api.AscendAllowLinkEnv
	// PodPredicateTime pod predicate time
	PodPredicateTime = "predicate-time"
	// Pod2kl pod annotation key, means kubelet allocate device
	Pod2kl = "kltDev"
	// SuperPodIDKey super node id
	SuperPodIDKey = "superPodID"
	// ChipNameLabel label value is card type, eg. 910A
	ChipNameLabel = "node.kubernetes.io/npu.chip.name"
	// MetaDataAnnotation downward api which map annotation from volcano to container's env
	MetaDataAnnotation = "metadata.annotations"
	// MetaData is meta data of pod
	MetaData = "metadata"
	// ResetInfoAnnotationKey is the key of reset fail information in node annotation
	ResetInfoAnnotationKey = "ResetInfo"
	// DefaultScanDelay default delay time before scanning devices reset by third party, seconds
	DefaultScanDelay = 300

	// SlowNodeStepTimeEnvNum is the number of environment value for step time cm
	SlowNodeStepTimeEnvNum = 2
	// PerfDumpPathEnv is an environment variable for slow node step time configmap
	PerfDumpPathEnv = "PERF_DUMP_PATH"
	// PerfDumpConfigEnv is an environment variable for slow node step time configmap
	PerfDumpConfigEnv = "PERF_DUMP_CONFIG"

	// VirtualDev Virtual device tag
	VirtualDev = "VIRTUAL"
	// PhyDeviceLen like Ascend910-0 split length is 2
	PhyDeviceLen = 2
	// VirDeviceLen like Ascend910-2c-100-1 split length is 4
	VirDeviceLen = 4
	// MaxDevicesNum max device num
	MaxDevicesNum = 100
	// MaxCardNum max card num
	MaxCardNum = 64
	// MaxDevNumInCard max device num in card
	MaxDevNumInCard = 4
	// MaxRequestVirtualDeviceNum max request device num
	MaxRequestVirtualDeviceNum = 1
	// LabelDeviceLen like Ascend910-0 split length is 2
	LabelDeviceLen = 2
	// DefaultDeviceIP device ip address
	DefaultDeviceIP = "127.0.0.1"

	// SocketChmod socket file mode
	SocketChmod = 0600

	// AMPMode for AMP chip work mode
	AMPMode = "AMP"
	// SMPMode for SMP chip work mode
	SMPMode = "SMP"

	// Interval interval time
	Interval = 1
	// Timeout time
	Timeout = 10
	// BaseDec base
	BaseDec = 10
	// BitSize base size
	BitSize = 64
	// BitSize32 base size 32
	BitSize32 = 32
	// SleepTime The unit is seconds
	SleepTime = 5

	// GeneralMapSize general map size
	GeneralMapSize = 8
	// MapSizeTwo map size two
	MapSizeTwo = 2
	// GeneralSubscribeTime general subscribe try time
	GeneralSubscribeTime = 3
	// Hex hexadecimal
	Hex = 16
	// SecondMagnification is second-level unit magnification
	SecondMagnification = 1000
	// SecondMagnificationFloat is second-level unit magnification float
	SecondMagnificationFloat = 1000.0
)

const (
	// DistributedJob annotation indicates that the job is distributed
	DistributedJob = "distributed-job"

	// Ascend310PV 310P-V
	Ascend310PV = api.Ascend310P + "-V"
	// Ascend310PVPro 310P-VPro
	Ascend310PVPro = api.Ascend310P + "-VPro"
	// Ascend310PIPro 310P-IPro
	Ascend310PIPro = api.Ascend310P + "-IPro"
	// Ascend310Pc1 Ascend310P 1 core
	Ascend310Pc1 = api.Ascend310P + "-" + Core1
	// Ascend310Pc2 Ascend310P 2 core
	Ascend310Pc2 = api.Ascend310P + "-" + Core2
	// Ascend310Pc4 Ascend310P 4 core
	Ascend310Pc4 = api.Ascend310P + "-" + Core4
	// Ascend310Pc4Cpu3 Ascend310P 4core 3cpu
	Ascend310Pc4Cpu3 = api.Ascend310P + "-" + Core4Cpu3
	// Ascend310Pc2Cpu1 Ascend310P 2core 1cpu
	Ascend310Pc2Cpu1 = api.Ascend310P + "-" + Core2Cpu1
	// Ascend310Pc4Cpu4Dvpp Ascend310P 4core 4cpu dvpp
	Ascend310Pc4Cpu4Dvpp = api.Ascend310P + "-" + Core4Cpu4Dvpp
	// Ascend310Pc4Cpu3Ndvpp Ascend310P 4core 3cpu ndvpp
	Ascend310Pc4Cpu3Ndvpp = api.Ascend310P + "-" + Core4Cpu3Ndvpp

	// Ascend910vir2  Ascend910 2core
	Ascend910vir2 = api.Ascend910 + "-" + Core2
	// Ascend910vir4 Ascend910 4core
	Ascend910vir4 = api.Ascend910 + "-" + Core4
	// Ascend910vir8 Ascend910 8core
	Ascend910vir8 = api.Ascend910 + "-" + Core8
	// Ascend910vir16 Ascend910 16core
	Ascend910vir16 = api.Ascend910 + "-" + Core16
	// Ascend910vir5Cpu1Gb8 Ascend910 5core 1cpu 8 Gb memory
	Ascend910vir5Cpu1Gb8 = api.Ascend910 + "-" + Core5Cpu1Gb8
	// Ascend910vir5Cpu1Gb16 Ascend910 5core 1cpu 16Gb memory
	Ascend910vir5Cpu1Gb16 = api.Ascend910 + "-" + Core5Cpu1Gb16
	// Ascend910vir6Cpu1Gb16 Ascend910 6core 1cpu 16Gb memory
	Ascend910vir6Cpu1Gb16 = api.Ascend910 + "-" + Core6Cpu1Gb16
	// Ascend910vir10Cpu3Gb16 Ascend910 10core 3cpu 16Gb memory
	Ascend910vir10Cpu3Gb16 = api.Ascend910 + "-" + Core10Cpu3Gb16

	// Ascend910vir10Cpu3Gb16Ndvpp Ascend910 10core 3cpu 16Gb memory ndvpp
	Ascend910vir10Cpu3Gb16Ndvpp = api.Ascend910 + "-" + Core10Cpu3Gb16Ndvpp
	// Ascend910vir10Cpu3Gb32 Ascend910 10core 3cpu 32Gb memory
	Ascend910vir10Cpu3Gb32 = api.Ascend910 + "-" + Core10Cpu3Gb32
	// Ascend910vir10Cpu4Gb16Dvpp Ascend910 10core 4cpu 16Gb memory dvpp
	Ascend910vir10Cpu4Gb16Dvpp = api.Ascend910 + "-" + Core10Cpu4Gb16Dvpp

	// Ascend910vir12Cpu3Gb32 Ascend910 12core 3cpu 32Gb memory
	Ascend910vir12Cpu3Gb32 = api.Ascend910 + "-" + Core12Cpu3Gb32

	// Ascend910vir3Cpu1Gb8 Ascend910 3core 1cpu 8Gb memory
	Ascend910vir3Cpu1Gb8 = api.Ascend910 + "-" + Core3Cpu1Gb8

	// AscendfdPrefix use in fd
	AscendfdPrefix = "davinci-mini"

	// HuaweiNetworkUnHealthAscend910 910 network unhealthy
	HuaweiNetworkUnHealthAscend910 = api.HuaweiAscend910 + "-NetworkUnhealthy"
	// HuaweiDpuUnHealthAscend910 910 DPU unhealthy
	HuaweiDpuUnHealthAscend910 = api.HuaweiAscend910 + "-DPUUnhealthy"
	// HuaweiUnHealthAscend910 unhealthy
	HuaweiUnHealthAscend910 = api.HuaweiAscend910 + "-Unhealthy"
	// HuaweiRecoveringAscend910 recovering
	HuaweiRecoveringAscend910 = api.HuaweiAscend910 + "-Recovering"
	// HuaweiUnHealthAscend310P 310p unhealthy
	HuaweiUnHealthAscend310P = api.HuaweiAscend310P + "-Unhealthy"
	// HuaweiUnHealthAscend310 310 unhealthy
	HuaweiUnHealthAscend310 = api.HuaweiAscend310 + "-Unhealthy"
	// HuaweiNetworkRecoverAscend910 910 network recover
	HuaweiNetworkRecoverAscend910 = api.HuaweiAscend910 + "-NetworkRecover"
	// HuaweiRecoverAscend910 910 recover
	HuaweiRecoverAscend910 = api.HuaweiAscend910 + "-Recover"

	// HuaweiFaultCodeAscend910 910 fault code
	HuaweiFaultCodeAscend910 = api.HuaweiAscend910 + "-Fault"
	// HuaweiFaultCodeAscend310P 310p fault code
	HuaweiFaultCodeAscend310P = api.HuaweiAscend310P + "-Fault"
	// HuaweiFaultCodeAscend310 310 fault code
	HuaweiFaultCodeAscend310 = api.HuaweiAscend310P + "-Fault"

	// AiCoreResourceName resource name for virtual device
	AiCoreResourceName = "npu-core"

	// Core1 1 core
	Core1 = "1c"
	// Core2 2 core
	Core2 = "2c"
	// Core2Cpu1 2core 1cpu
	Core2Cpu1 = "2c.1cpu"

	// Core3Cpu1Gb8 3 core, 1 cpu and 8GB memory
	Core3Cpu1Gb8 = "3c.1cpu.8g"
	// Core4 4 core
	Core4 = "4c"
	// Core4Cpu3 4core 3cpu
	Core4Cpu3 = "4c.3cpu"
	// Core4Cpu3Ndvpp 4core 3cpu ndvpp
	Core4Cpu3Ndvpp = "4c.3cpu.ndvpp"
	// Core4Cpu4Dvpp 4core 4cpu dvpp
	Core4Cpu4Dvpp = "4c.4cpu.dvpp"
	// Core5Cpu1Gb8 5 core, 1 cpu and 8GB memory
	Core5Cpu1Gb8 = "5c.1cpu.8g"
	// Core5Cpu1Gb16 5 core, 1 cpu and 16GB memory
	Core5Cpu1Gb16 = "5c.1cpu.16g"

	// Core6Cpu1Gb16 6 core, 1 cpu and 16GB memory
	Core6Cpu1Gb16 = "6c.1cpu.16g"

	// Core8 8 core
	Core8 = "8c"
	// Core10Cpu3Gb16 10 core, 3 cpu and 16Gb memory
	Core10Cpu3Gb16 = "10c.3cpu.16g"

	// Core10Cpu3Gb16Ndvpp 10 core, 3 cpu, 16Gb memory and ndvpp
	Core10Cpu3Gb16Ndvpp = "10c.3cpu.16g.ndvpp"
	// Core10Cpu3Gb32 10 core, 3 cpu and 32GB memory
	Core10Cpu3Gb32 = "10c.3cpu.32g"
	// Core10Cpu4Gb16Dvpp 10 core, 4 cpu, 16Gb memory and dvpp
	Core10Cpu4Gb16Dvpp = "10c.4cpu.16g.dvpp"

	// Core12Cpu3Gb32 12 core, 3 cpu and 32GB memory
	Core12Cpu3Gb32 = "12c.3cpu.32g"

	// Core16 16 core
	Core16 = "16c"

	// Vir01 template name vir01
	Vir01 = "vir01"
	// Vir02 template name vir02
	Vir02 = "vir02"
	// Vir02C1 template name vir02_1c
	Vir02C1 = "vir02_1c"
	// Vir03C1G8 template name vir03_1c_8g
	Vir03C1G8 = "vir03_1c_8g"
	// Vir04 template name vir04
	Vir04 = "vir04"
	// Vir04C3 template name vir04_3c
	Vir04C3 = "vir04_3c"
	// Vir04C4Dvpp template name vir04_4c_dvpp
	Vir04C4Dvpp = "vir04_4c_dvpp"
	// Vir04C3Ndvpp template name vir04_3c_ndvpp
	Vir04C3Ndvpp = "vir04_3c_ndvpp"
	// Vir05C1G8 template name vir05_1c_8g
	Vir05C1G8 = "vir05_1c_8g"
	// Vir05C1G16 template name vir05_1c_16g
	Vir05C1G16 = "vir05_1c_16g"
	// Vir06C1G16 template name vir06_1c_16g
	Vir06C1G16 = "vir06_1c_16g"
	// Vir08 template name vir08
	Vir08 = "vir08"
	// Vir10C3G16 template name vir10_3c_16g
	Vir10C3G16 = "vir10_3c_16g"
	// Vir10C3G16NM template name vir10_3c_16g_nm
	Vir10C3G16NM = "vir10_3c_16g_nm"
	// Vir10C3G32 template name vir10_3c_32g
	Vir10C3G32 = "vir10_3c_32g"
	// Vir10C4G16M template name vir10_4c_16g_m
	Vir10C4G16M = "vir10_4c_16g_m"
	// Vir12C3G32 template name vir12_3c_32g
	Vir12C3G32 = "vir12_3c_32g"
	// Vir16 template name vir16
	Vir16 = "vir16"

	// VirMark the mark of virtual device
	VirMark = "vir"

	// AnnotationVNPUInfoSplitLen length of pod annotation for allocate vnpu info
	AnnotationVNPUInfoSplitLen = 2

	// MaxAICoreNum max ai core num, A5 max 36
	MaxAICoreNum = 36
	// MinAICoreNum min ai core num
	MinAICoreNum = 8
	// DefaultIDForCreateVNPU default id for creating vnpu
	DefaultIDForCreateVNPU = 0xFFFFFFFF

	// ServerTypeInfoMinLen the min len of server type split data
	ServerTypeInfoMinLen = 2
	// VGroupAndDevLen a list only contain virtual group and device
	VGroupAndDevLen = 2
	// MaxShareDevCount open share device function, max share count is 100
	MaxShareDevCount = 100
)

const (
	// ServerTypeLabelKey the node label key of server type
	ServerTypeLabelKey = "servertype"
	// AcceleratorTypeKey the node label key of accelerator type
	AcceleratorTypeKey = "accelerator-type"
	// ServerUsageLabelKey is to indicate the usage of server
	// is infer or training, currently only related to A800IA2 infer server
	ServerUsageLabelKey = "server-usage"
	// InferCardKey the node label key of infer card
	InferCardKey = "infer-card-type"
)

const (
	// HiAIHDCDevice hisi_hdc
	HiAIHDCDevice = "/dev/hisi_hdc"
	// HiAIManagerDevice davinci_manager
	HiAIManagerDevice = "/dev/davinci_manager"
	// HiAIManagerDeviceDocker davinci_manager for docker
	HiAIManagerDeviceDocker = "/dev/davinci_manager_docker"
	// HiAISVMDevice devmm_svm
	HiAISVMDevice = "/dev/devmm_svm"
	// HiAi200RCSVM0 svm0
	HiAi200RCSVM0 = "/dev/svm0"
	// HiAi200RCLog log_drv
	HiAi200RCLog = "/dev/log_drv"
	// HiAi200RCEventSched event_sched
	HiAi200RCEventSched = "/dev/event_sched"
	// HiAi200RCUpgrade upgrade
	HiAi200RCUpgrade = "/dev/upgrade"
	// HiAi200RCHiDvpp hi_dvpp
	HiAi200RCHiDvpp = "/dev/hi_dvpp"
	// HiAi200RCMemoryBandwidth memory_bandwidth
	HiAi200RCMemoryBandwidth = "/dev/memory_bandwidth"
	// HiAi200RCTsAisle ts_aisle
	HiAi200RCTsAisle = "/dev/ts_aisle"
)

const (
	// Atlas200ISoc 200 soc env
	Atlas200ISoc = "Atlas 200I SoC A1"
	// Atlas200ISocXSMEM is xsmem_dev
	Atlas200ISocXSMEM = "/dev/xsmem_dev"
	// Atlas200ISocSYS is sys
	Atlas200ISocSYS = "/dev/sys"
	// Atlas200ISocVDEC is vdec
	Atlas200ISocVDEC = "/dev/vdec"
	// Atlas200ISocVPC is vpc
	Atlas200ISocVPC = "/dev/vpc"
	// Atlas200ISocSpiSmbus is spi_smbus
	Atlas200ISocSpiSmbus = "/dev/spi_smbus"
	// Atlas200ISocUserConfig is user_config
	Atlas200ISocUserConfig = "/dev/user_config"
)

const (
	// DvppCmdlistDevice is dvpp_cmdlist
	DvppCmdlistDevice = "/dev/dvpp_cmdlist"
	// Atlas310BPngd is pngd
	Atlas310BPngd = "/dev/pngd"
	// Atlas310BVenc is venc
	Atlas310BVenc = "/dev/venc"
)

// Audio and video dependent device for Atlas310B
const (
	Atlas310BAcodec = "/dev/acodec"
	Atlas310BAi     = "/dev/ai"
	Atlas310BAo     = "/dev/ao"
	Atlas310BVo     = "/dev/vo"
	Atlas310BHdmi   = "/dev/hdmi"
)

// UB device
const (
	UburmaDevicePath = "/dev/uburma"
	UmmuDevicePath   = "/dev/ummu"
)

const (
	// RootUID is root user id
	RootUID = 0
	// RootGID is root group id
	RootGID = 0

	// KeySliceLength is the length of key slice check
	KeySliceLength = 2

	// DotSepDev if the separator between devices on labels
	DotSepDev = "."

	// CommaSepDev if the separator between devices on annotation
	CommaSepDev = ","
	// MiddelLine if the separator between devices for split id
	MiddelLine = "-"
	// UnderLine the separator between ids
	UnderLine = "_"

	// NoNPUResource means allocated some devices that don't exist
	NoNPUResource = "NoNPUResource"
	// NPUSegmentFailed means create vnpu device failed
	NPUSegmentFailed = "NPUSegmentFailed"
	// CenterScene deploy the device-plugin component on the central side
	CenterScene = "center"
	// EdgeScene deploy the device-plugin component on the edge side
	EdgeScene = "edge"
	// A300IA2BoardId board id of A300I A2 32GB
	A300IA2BoardId = 0x28
	// A300IA2GB64BoardId board id of A300I A2 64GB
	A300IA2GB64BoardId = 0x29
	// A800IA2NoneHccsBoardIdOld is the boardid of a800i a2 device,0x33 is server without hccs
	A800IA2NoneHccsBoardIdOld = 0x33
	// A800IA2NoneHccsBoardId 0x33 changed to 0x3c , and compatible with the old boardId ,since 2024.9.4
	A800IA2NoneHccsBoardId = 0x3c
	// EmptyBoardId is the boardid of device before initialized
	EmptyBoardId = 0x00
	// FirstDevice the first device id
	FirstDevice = 0
	// Infer means device for inference
	Infer = "infer"
	// Train means device for training
	Train = "train"
)

// Special scene for invoking the dcmi interface
const (
	DeviceNotSupport = 8255
	// DefaultAiCoreNum set a default value of aicore number
	DefaultAiCoreNum = 1
)

const (
	// Atlas300IDuo for hot reset function, sync chip healthy state
	Atlas300IDuo = "Atlas 300I Duo"
	// HotResetClose not using chip hot reset function
	HotResetClose = -1
	// HotResetInfer using infer chip hot reset
	HotResetInfer = 0
	// HotResetTrainOnLine using train chip hot reset online
	HotResetTrainOnLine = 1
	// HotResetTrainOffLine using train chip hot reset offline
	HotResetTrainOffLine = 2
	// BootStartFinish chip hot reset finish
	BootStartFinish = 16
	// SleepMinutesForA3Reset sleep minutes before a3 card reset case hotReset=2
	SleepMinutesForA3Reset = 5
)

const (
	// Ascend910RingsNum indicates the number of devices in a ring
	Ascend910RingsNum = 4
	// Ascend910BRingsNumTrain indicates the number of devices in a ring
	Ascend910BRingsNumTrain = 8
	// Ascend910BRingsNumInfer indicates the number of devices in a ring
	Ascend910BRingsNumInfer = 1
	// Ascend910A3RingsNum indicates the number of devices in a ring
	Ascend910A3RingsNum = 16
	// Ascend910A5RingsNum indicates the number of devices in a ring
	Ascend910A5RingsNum = 1
	// RingSum indicates the max number of ring
	RingSum = 2
	// InferRankIndex indecates the rank index of infer situation (rank index is meaningless in infer situation)
	InferRankIndex = "-1"
	// ResetRetryTimes for max retry times when reset failed
	ResetRetryTimes = 4
	// A200TA2RingsNum indicates the number of devices in a ring
	A200TA2RingsNum = 16
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
	// ResetInfoTypeKey for reset configmap type key
	ResetInfoTypeKey = "restartType"
	// HotResetRestartType for hot reset restart type
	HotResetRestartType = "hotReset"
	// ResetTaskNameKey for obtain the reset task name
	ResetTaskNameKey = "volcano.sh/job-name"
	// ResetTaskNameKeyInLabel for obtain the reset task name when using operator
	ResetTaskNameKeyInLabel = "training.kubeflow.org/job-name"
)

const (
	// FaultInfoCMNamePrefix for fault configmap name prefix
	FaultInfoCMNamePrefix = "fault-config-"
	// FaultInfoCMDataKey for fault configmap data key
	FaultInfoCMDataKey = "fault-npus"
	// FaultInfoCMCheckCodeKey for fault configmap checkcode key
	FaultInfoCMCheckCodeKey = "checkCode"
)

const (
	// EmptyError indicates that there is no fault
	EmptyError = "empty"
	// IgnoreError indicates that the current fault can be ignored
	IgnoreError = "ignore"
	// RestartRequestError indicates that the task only needs to re-execute this request
	RestartRequestError = "restart_request"
	// RestartError indicates that the training needs to be re-executed for the current fault
	RestartError = "restart"
	// FreeResetError indicates the fault level of the device to be reset whenever there is no task on NPU
	FreeResetError = "free_reset"
	// ResetError indicates that the current fault requires resetting the chip and re-executing the training
	ResetError = "reset"
	// IsolateError indicates that the device needs to be isolated due to the current fault
	IsolateError = "isolate"
)

const (
	// EmptyErrorLevel indicates the level of no fault state
	EmptyErrorLevel = iota
	// IgnoreErrorLevel indicates the level of a fault that can be ignored
	IgnoreErrorLevel
	// RestartRequestErrorLevel indicates that the task only needs to re-execute this request
	RestartRequestErrorLevel
	// RestartErrorLevel indicates the level of the fault that needs to be re-executed
	RestartErrorLevel
	// FreeResetErrorLevel indicates the fault level of the device to be reset whenever there is no task on NPU
	FreeResetErrorLevel
	// ResetErrorLevel indicates the fault level of the device to be reset
	ResetErrorLevel
	// IsolateErrorLevel indicates the fault level of the device to be isolated
	IsolateErrorLevel
)

const (
	// UnrecoveredStatus indicates the status before recovery
	UnrecoveredStatus = "unrecovered"
	// RecoveredStatus indicates that the recovery is successful
	RecoveredStatus = "recovered"
	// RecoverFailedStatus indicates that the recovery fails
	RecoverFailedStatus = "failed"
)

const (
	// MaxResetWaitRecoverTime max reset wait chip recover time is 150s
	MaxResetWaitRecoverTime = 150
)

const (
	// AssertionRecovery the name of assertion 0
	AssertionRecovery = "Recovery"
	// AssertionOccur the name of assertion 1
	AssertionOccur = "Occur"
	// AssertionNotice the name of assertion 2
	AssertionNotice = "Notice"

	// TimeFormat the format for time
	TimeFormat = "2006-01-02 15:04:05"

	// ResourceKindPod the kind pod of resource
	ResourceKindPod = "pod"
)

// Fault customization const
const (
	// PollFaultCodeCMInterval is the default interval(second) of polling fault code CM
	PollFaultCodeCMInterval = 300
	// PollFaultCodeCMMaxInterval is the max interval(second) of polling fault code CM
	PollFaultCodeCMMaxInterval = 3600
	// PollFaultCodeCMMinInterval is the min interval(second) of polling fault code CM
	PollFaultCodeCMMinInterval = 30
	// GetSwitchFaultCodeInterval is the interval(second) of get all fault code by get interface
	GetSwitchFaultCodeInterval = 300
	// MaxLengthOfFaultCode [0x00f103b0,155904,na,na] must contain at most 50 characters
	MaxLengthOfFaultCode = 50
	// PartNumOfFaultCode [0x00f103b0,155904,na,na] must have 4 parts
	PartNumOfFaultCode = 4
	// FaultCodeCMName is the name of the configmap that is used to save fault code
	FaultCodeCMName = "mindx-dl-fault-config"
	// FaultCodeKey is the key to find fault code in cm
	FaultCodeKey = "faultCode.json"
	// SwitchFaultCodeKey is the key of the switch fault code
	SwitchFaultCodeKey = "SwitchFaultCode.json"
	// FaultCustomizationKey is the key to find fault customization in cm
	FaultCustomizationKey = "faultCustomization.json"
	// PollIntervalKey is the key to find poll interval in cm
	PollIntervalKey = "PollInterval"
	// DefaultProcessReadCMTime is the default time for process read configmap
	DefaultProcessReadCMTime = 30
	// DefaultWaitFaultSelfHealingTime for waiting for fault self-healing
	DefaultWaitFaultSelfHealingTime = 15
	// MinWaitFaultSelfHealingTime for min time of waiting for fault self-healing
	MinWaitFaultSelfHealingTime = 1
	// MaxWaitFaultSelfHealingTime for max time of waiting for fault self-healing
	MaxWaitFaultSelfHealingTime = 30
	// DefaultPollingInterval  represents the time between polls of the dcmi interface
	DefaultPollingInterval = 1
	// MaxWaitProcessReadCMTime for max time waiting for process to read cm
	MaxWaitProcessReadCMTime = 90
	// MinWaitProcessReadCMTime for min time waiting for process to read cm
	MinWaitProcessReadCMTime = 5
	// DefaultWaitDeviceResetTime is the default time used in waiting device reset
	DefaultWaitDeviceResetTime = 150
	// MaxWaitDeviceResetTime is the max time used in waiting device reset
	MaxWaitDeviceResetTime = 180
	// MinWaitDeviceResetTime is the min time used in waiting device reset
	MinWaitDeviceResetTime = 60
	// MaxFaultFrequencyTimeWindow is the max time for the time window of fault frequency
	MaxFaultFrequencyTimeWindow = 864000
	// MinFaultFrequencyTimeWindow is the min time for the time window of fault frequency
	MinFaultFrequencyTimeWindow = 60
	// MaxReleaseTimeWindow is the max time for the time window of release upgrade fault
	MaxReleaseTimeWindow = (1 << 63) - 1
	// MinReleaseTimeWindow is the min time for the time window of release upgrade fault
	MinReleaseTimeWindow = 60
	// MaxFaultFrequencyTimes is the max count for the fault occurrence time of fault frequency
	MaxFaultFrequencyTimes = 100
	// MinFaultFrequencyTimes is the min count for the fault occurrence time of fault frequency
	MinFaultFrequencyTimes = 1
	// MaxFaultTimeout is the max time(s) for the fault duration time of fault duration
	MaxFaultTimeout = 600
	// MinFaultTimeout is the min time(s) for the fault duration time of fault duration
	MinFaultTimeout = 0
	// MaxRecoverTimeout is the max time(s) for the fault recover duration time of fault duration
	MaxRecoverTimeout = 86400
	// MinRecoverTimeout is the min time(s) for the fault recover duration time of fault duration
	MinRecoverTimeout = 0
	// DefaultSubscribeToPollingTime is the default time from subscribe to polling
	DefaultSubscribeToPollingTime = 5
	// MaxLogicID is the maximum logic ID
	MaxLogicID = 15
	// MinLogicID is the minimum logic ID
	MinLogicID = 0
	// MaxResetTimes the max reset times of a device while error happened,
	// setting to 30 to avoid manually reset on host machine
	MaxResetTimes = 3
)

// the severity level of fault
const (
	FaultSeveritySuggestion = iota
	FaultSeverityMinor
	FaultSeverityMajor
	FaultSeverityCritical
)

// peer device type of switch
const (
	// PeerDeviceChipOrCpuPort if peer device is whole chip or cpu the given value should be 0
	PeerDeviceChipOrCpuPort = 0
	// PeerDeviceChipOrCpuPortName the name of  peer device is whole chip or cpu
	PeerDeviceChipOrCpuPortName = "cpu"
	// PeerDeviceNpuPort 1 means switch contact peer device is npu
	PeerDeviceNpuPort = 1
	// PeerDeviceNpuPortName the name of peer device is npu
	PeerDeviceNpuPortName = "npu"
	// PeerDeviceL2Port 1 means switch contact peer device is L2
	PeerDeviceL2Port = 2
	// PeerDeviceL2PortName the name of peer device is L2
	PeerDeviceL2PortName = "L2"
	// PeerDeviceNAPortName the name of switch peer device is not valid
	PeerDeviceNAPortName = "na"
)

// LogicID list for reset, get id list of ring
const (
	ManuallySeparateNpuFirstHandle = "FirstHandle"
	ManuallySeparateNpuHandled     = "Handled"
	ManuallySeparateNpuAll         = "All"
)

// ApiServerPort is port of API server
const ApiServerPort = "443"

const (
	// InitialProcNum represents the initial value of the number of remaining processes
	InitialProcNum = 1
)

const (
	// SdIdAbnormal represents super pod sdid abnormal value
	SdIdAbnormal = -2
	// ScaleTypeAbnormal represents super pod scaleType abnormal value
	ScaleTypeAbnormal = -2
	// SuperPodIdAbnormal represents super pod superPodId abnormal value
	SuperPodIdAbnormal = -2
	// ServerIdAbnormal represents super pod serverId abnormal value
	ServerIdAbnormal = -2
)

const (
	// TimeoutProcess represents fault timeout process
	TimeoutProcess = "fault timeout"
	// TimeoutRecoverProcess represents fault timeout recover process
	TimeoutRecoverProcess = "fault timeout recover"
)

const (
	// ChipFaultMode represents chip fault mode
	ChipFaultMode = "chip fault mode"
	// NetworkFaultMode represents network fault mode
	NetworkFaultMode = "network fault mode"
)

const (
	// Polling represents subscribe mode invalid and polling is used scenario
	Polling = "polling"
	// Subscribe represents subscribe mode
	Subscribe = "subscribe"
)

const (
	// NPUNormalStatus represents normal status
	NPUNormalStatus = "normal"
	// NPUResettingStatus represents resetting status
	NPUResettingStatus = "resetting"
)

const (
	// HbmDoubleBitFaultCode indicate 0x80E01801
	HbmDoubleBitFaultCode = 2162169857
	// HbmDoubleBitFaultCodeStr indicate 80e01801
	HbmDoubleBitFaultCodeStr = "80e01801"
	// AivBusFaultCode indicate 0x80CB8009
	AivBusFaultCode = 2160820233
	// AicBusFaultCode indicate 0x80C98009
	AicBusFaultCode = 2160689161
	// AssociatedFaultDiagnosisTime associated fault diagnosis
	AssociatedFaultDiagnosisTime = 5
	// TimeMilliseconds indicate how many milliseconds are there in a second
	TimeMilliseconds = 1000
)

const (
	// DataTraceCmPrefix is the prefix string for profiling confingmap
	DataTraceCmPrefix = "data-trace-"
	// DataTraceConfigDir is the directory containing the configuration
	DataTraceConfigDir = "/user/cluster-info/datatrace-config"
	// DataTraceCmProfilingSwitchKey is the key in the configuration map for enabling profiling switch
	DataTraceCmProfilingSwitchKey = "profilingSwitch"
)

const (
	// FailureCountThresholdForRestart threshold number of consecutive send failures for restart dp
	FailureCountThresholdForRestart = 15
	// FailureCountThresholdForReRegistry threshold number of consecutive send failures for reRegistry to kubelet
	FailureCountThresholdForReRegistry = 5
	// CheckFailurePeriodSecond period of check connection between kubelet and device-plugin
	CheckFailurePeriodSecond = time.Second * 5
	// EmptyStrategy do nothing for send result
	EmptyStrategy = ""
	// ReRegistryStrategy a strategy registry to kubelet
	ReRegistryStrategy = "reRegistry"
	// ReStartDevicePluginStrategy a strategy restart device-plugin
	ReStartDevicePluginStrategy = "restart"
	// MaxSendRecordLength max length record send result
	MaxSendRecordLength = 1024
	// DefaultSendRecordLength default length record send result
	DefaultSendRecordLength = 128
	// GrpcKeepAliveTime keep alive time for grpc connection
	GrpcKeepAliveTime = 5 * time.Minute
	// GrpcKeepAliveTimeout grpc timeout for keep-alive ping response
	GrpcKeepAliveTimeout = 5 * time.Minute
)

const (
	// DefaultPerm default perm for creating dir
	DefaultPerm = 0666
)

const (
	// MaxPodEventRetryTimes max try time for pod add event while cache none
	MaxPodEventRetryTimes = 4
)

const (
	// WriteEventRateLimit upper limit rate of write fault to k8s event per minute
	WriteEventRateLimit = 10
	// FaultCallBackRateLimit  upper limit rate of call back receive fault from driver per minute
	FaultCallBackRateLimit = 1000
	// WriteEventChanLenLimit upper limit of length of cache event
	WriteEventChanLenLimit = 100
)

const (
	// L1ToL2PlanePortNum l1 switch port number link to l2
	L1ToL2PlanePortNum = 16
	// ResetFaultToleranceTimeInterval fault tolerance time interval
	ResetFaultToleranceTimeInterval = 60
)
const (
	// DefaultSuperPodID default super pod id -1
	DefaultSuperPodID = -1
	// DefaultServerIndex default server index -1
	DefaultServerIndex = -1
)
