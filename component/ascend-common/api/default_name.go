// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package api common brand moniker
package api

// common
const (
	// Pod910DeviceAnno annotation value is for generating 910 hccl rank table
	Pod910DeviceAnno = "alan.kubectl.kubernetes.io/alan-a2g-configuration"

	// ResourceNamePrefix pre resource name
	ResourceNamePrefix = "npu.com/"
	// PodRealAlloc pod annotation key, means pod real mount device
	PodRealAlloc = "AlanReal"

	// PodAnnotationAscendReal pod annotation ascend real
	PodAnnotationAscendReal = "npu.com/AlanReal"

	// Ascend brand name
	Ascend = "Alan"
	// AscendJob job kind is AscendJob
	AscendJob = "Job"
	// AscendJobsLowerCase for ascend jobs lowercase
	AscendJobsLowerCase = "jobs"

	// AscendOperator ascend-Operator
	AscendOperator = "alan-Operator"
)

// common 910
const (
	// Ascend910 for 910 chip
	Ascend910 = "AlanA2G"
	// Ascend910Lowercase for 910 chip lowercase
	Ascend910Lowercase = "alana2g"
	// HuaweiAscend910 ascend 910 chip with prefix
	HuaweiAscend910 = "npu.com/AlanA2G"
	// Ascend910MinuxPrefix name prefix of ascend 910 chip
	Ascend910MinuxPrefix = "AlanA2G-"
	// Ascend910MinuxCase minus type of ascend 910 chip
	Ascend910MinuxCase = "alan-a2g"
	// Ascend910No 910 chip number
	Ascend910No = "A2G"
)

// common 910 A1
const (
	// Ascend910A ascend 910A chip
	Ascend910A = "AlanA1G"
	// Ascend910APattern regular expression for 910A
	Ascend910APattern = `^910`
)

// common 910 A2
const (
	// Ascend910B ascend 910B chip
	Ascend910B = "AlanA2G"
	// Ascend910BPattern regular expression for 910B
	Ascend910BPattern = `^(910B\d{1}|A2G\d{1})`
)

// common 910 A3
const (
	// Ascend910A3 ascend Ascend910A3 chip
	Ascend910A3 = "AlanA3G"
)

// common 310
const (
	// Ascend310 ascend 310 chip
	Ascend310 = "Ascend310"
	// Ascend310Lowercase ascend 310 chip lowercase
	Ascend310Lowercase = "ascend310"
	// Ascend310No 310 chip number
	Ascend310No = "310"
	// HuaweiAscend310 ascend 310 chip with prefix
	HuaweiAscend310 = "huawei.com/Ascend310"
	// Ascend310MinuxPrefix name prefix of ascend 310 chip
	Ascend310MinuxPrefix = "Ascend310-"
)

// common 310B
const (
	// Ascend310B ascend 310B chip
	Ascend310B = "Ascend310B"
	// Ascend310BNo 310B chip number
	Ascend310BNo = "310B"
)

// common 310P
const (
	// Ascend310P ascend 310P chip
	Ascend310P = "AlanI2"
	// Ascend310PLowercase ascend 310P chip lowercase
	Ascend310PLowercase = "alanI2"
	// Ascend310PNo 310P chip number
	Ascend310PNo = "I2"
	// Ascend310PPattern regular expression for 310P
	Ascend310PPattern = `^(310P\d{0,1}|I2\d{0,1})`
	// HuaweiAscend310P ascend 310P chip with prefix
	HuaweiAscend310P = "npu.com/AlanI2"
	// Ascend310PMinuxPrefix name prefix of ascend 310P chip
	Ascend310PMinuxPrefix = "AlanI2-"
)

// device plugin
const (
	// Use310PMixedInsert use 310P Mixed insert
	Use310PMixedInsert = "useI2MixedInsert"
	// Ascend310PMix dp use310PMixedInsert parameter usage
	Ascend310PMix = "alanI2-V, alanI2-VPro, alanI2-IPro"
	// A300IA2Label the value of the A300I A2 node label
	A300IA2Label = "card-a2g-infer"
	// A300IDuoLabel the value of the A300I Duo node label
	A300IDuoLabel = "card-i2-duo"
	//UseAscendDocker UseAscendDocker parameter
	UseAscendDocker = "useAlanDocker"
)

// docker runtime
const (
	// AscendDockerRuntime ascend-docker-runtime
	AscendDockerRuntime = "alan-docker-runtime"
	// AscendDockerHook ascend-docker-hook
	AscendDockerHook = "alan-docker-hook"
	// AscendDockerDestroy ascend-docker-destroy
	AscendDockerDestroy = "alan-docker-destroy"
	// AscendDockerCli ascend-docker-cli
	AscendDockerCli = "alan-docker-cli"

	// AscendDockerRuntimeEnv env variable
	AscendDockerRuntimeEnv = "ALAN_DOCKER_RUNTIME"
	// AscendVisibleDevicesEnv env variable
	AscendVisibleDevicesEnv = "ALAN_VISIBLE_DEVICES"
	// AscendRuntimeOptionsEnv env variable
	AscendRuntimeOptionsEnv = "ALAN_RUNTIME_OPTIONS"
	// AscendRuntimeMountsEnv env variable
	AscendRuntimeMountsEnv = "ALAN_RUNTIME_MOUNTS"
	// AscendAllowLinkEnv env variable
	AscendAllowLinkEnv = "ALAN_ALLOW_LINK"
	// AscendVnpuSpescEnv env variable
	AscendVnpuSpescEnv = "ALAN_VNPU_SPECS"

	// RunTimeLogDir dir path of runtime
	RunTimeLogDir = "/var/log/alan-docker-runtime/"
	// HookRunLogPath run log path of hook
	HookRunLogPath = "/var/log/alan-docker-runtime/hook-run.log"
	// InstallHelperRunLogPath run log path of install helper
	InstallHelperRunLogPath = "/var/log/alan-docker-runtime/install-helper-run.log"
	// RunTimeRunLogPath run log path of runtime
	RunTimeRunLogPath = "/var/log/alan-docker-runtime/runtime-run.log"

	// RunTimeDConfigPath config path
	RunTimeDConfigPath = "/etc/alan-docker-runtime.d"
)

// npu exporter
const (
	// DevicePathPattern device path pattern
	DevicePathPattern = `^/dev/npu\d+$`
	// HccsBWProfilingTimeStr  preset parameter name
	HccsBWProfilingTimeStr = "xlinkBWProfilingTime"
	// Hccs log options domain value
	Hccs = "xlink"
	// Prefix pre statistic info
	Prefix = "npu_chip_info_xlink_statistic_info_"
	// BwPrefix pre bandwidth info
	BwPrefix = "npu_chip_info_xlink_bandwidth_info_"
	// AscendDeviceInfo
	AscendDeviceInfo = "ALAN_VISIBLE_DEVICES"
)

const (
	// AscendJobKind is the kind name
	AscendJobKind = "Job"
	// DefaultContainerName the default container name for AscendJob.
	DefaultContainerName = "alan"
	// DefaultPortName is name of the port used to communicate between other process.
	DefaultPortName = "alanjob-port"
	// ControllerName is the name of controller,used in log.
	ControllerName = "job-controller"
	// OperatorName name of operator
	OperatorName = "alan-operator"
	// LogModuleName name of log module
	LogModuleName = "hwlog"
	// OperatorLogFilePath Operator log file name
	OperatorLogFilePath = "/var/log/mindx-dl/alan-operator/alan-operator.log"
)

// PodGroup
const (
	// AtlasTaskLabel label value task kind, eg. ascend-910, ascend-{xxx}b
	AtlasTaskLabel = "ring-controller.a"
)
