/* Copyright(C) 2021-2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common define common variable
package common

import (
	"math"

	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/api"
)

// DeviceType define device type
type DeviceType struct {
	// Code device type code
	Code int32
	// Name device type name
	Name string
}

var (
	// ProfilingTime for getting PCIe bandwidth
	ProfilingTime int

	// HccsBWProfilingTime for getting hccs bandwidth
	HccsBWProfilingTime int

	// a3BoardIds for A3 Board IDs
	a3BoardIds = sets.NewInt32(A900A3SuperPodBin1BoardId, A900A3SuperPodBin2BoardId,
		A900A3SuperPodBin3BoardId, A800IA3BoardId, A800IA3BoardId2, A3SuperPodZQBoardId, A3ServerZQBoardId,
		A3SuperPodZQNpuBoardId, A3ServerZQNpuBoardId)

	// a900A3SuperPodMainBoardIds for A900 A3 Super Pod Main Board IDs
	a900A3SuperPodMainBoardIds = sets.NewInt32(A900A3SuperPodMainBoardId1, A900A3SuperPodMainBoardId2)

	// a9000A3SuperPodMainBoardIds for A9000 A3 Super Pod Main Board IDs
	a9000A3SuperPodMainBoardIds = sets.NewInt32(A9000A3SuperPodMainBoardId1, A9000A3SuperPodMainBoardId2)

	// ParameterPlaneDownProtsNumToPreciseFaultCodeMap maps device type to a mapping from ParameterPlaneDownProtsNum
	// to precise fault code
	ParameterPlaneDownProtsNumToPreciseFaultCodeMap = map[string]map[int]int64{
		api.Ascend910A5: {
			NumberOne: UBOESubHealFaultCode,
			NumberTwo: UBOEPreSeparateFaultCode,
		},
		api.Ascend910: {
			NumberOne: LinkDownFaultCode,
		},
	}
	// DetailCustomFaultCodesSet detail custom fault codes set
	DetailCustomFaultCodesSet = sets.NewInt64(UBSeparateFaultCode, UBSubHealFaultCode)
	// DetailCustomParameterPlaneFaultCodesSet detail custom parameter plane fault codes set
	DetailCustomParameterPlaneFaultCodesSet = sets.NewInt64(UBOEPreSeparateFaultCode, UBOESubHealFaultCode)
)

// DeviceType for utilization
var (
	// AICore Ascend310 & Ascend910
	AICore = DeviceType{Code: 2, Name: "AICore"}
	// HbmUtilization utilization rate of hbm
	HbmUtilization = DeviceType{Code: 6, Name: "Hbm"}
	// VectorCore Ascend310P
	VectorCore = DeviceType{Code: 12, Name: "VectorCore"}
	// Overall Overall utilization rate of NPU
	Overall = DeviceType{Code: 13, Name: "Overall"}
	// AICube AICube utilization rate of NPU
	AICube = DeviceType{Code: 14, Name: "AICube"}
)

// DeviceType for frequency
var (
	// AICoreCurrentFreq Ascend310 & Ascend910 & Ascend910B & Ascend310P
	AICoreCurrentFreq = DeviceType{Code: 7, Name: "AICore Current"}
)

// Shared NPU and vNPU naming constants.
const (
	// Minus separates the fields of an NPU device name.
	Minus = "-"
)

// Public vNPU type suffixes shared by device-plugin and DRA.
const (
	// Core1 represents a one-core vNPU type.
	Core1 = "1c"
	// Core2 represents a two-core vNPU type.
	Core2 = "2c"
	// Core2Cpu1 represents a two-core, one-CPU vNPU type.
	Core2Cpu1 = "2c.1cpu"
	// Core3Cpu1Gb8 represents a three-core, one-CPU, 8-GB vNPU type.
	Core3Cpu1Gb8 = "3c.1cpu.8g"
	// Core4 represents a four-core vNPU type.
	Core4 = "4c"
	// Core4Cpu3 represents a four-core, three-CPU vNPU type.
	Core4Cpu3 = "4c.3cpu"
	// Core4Cpu3Ndvpp represents a four-core, three-CPU vNPU type without DVPP.
	Core4Cpu3Ndvpp = "4c.3cpu.ndvpp"
	// Core4Cpu4Dvpp represents a four-core, four-CPU vNPU type with DVPP.
	Core4Cpu4Dvpp = "4c.4cpu.dvpp"
	// Core5Cpu1Gb8 represents a five-core, one-CPU, 8-GB vNPU type.
	Core5Cpu1Gb8 = "5c.1cpu.8g"
	// Core5Cpu1Gb16 represents a five-core, one-CPU, 16-GB vNPU type.
	Core5Cpu1Gb16 = "5c.1cpu.16g"
	// Core6Cpu1Gb16 represents a six-core, one-CPU, 16-GB vNPU type.
	Core6Cpu1Gb16 = "6c.1cpu.16g"
	// Core8 represents an eight-core vNPU type.
	Core8 = "8c"
	// Core10Cpu3Gb16 represents a ten-core, three-CPU, 16-GB vNPU type.
	Core10Cpu3Gb16 = "10c.3cpu.16g"
	// Core10Cpu3Gb16Ndvpp represents a ten-core, three-CPU, 16-GB vNPU type without DVPP.
	Core10Cpu3Gb16Ndvpp = "10c.3cpu.16g.ndvpp"
	// Core10Cpu3Gb32 represents a ten-core, three-CPU, 32-GB vNPU type.
	Core10Cpu3Gb32 = "10c.3cpu.32g"
	// Core10Cpu4Gb16Dvpp represents a ten-core, four-CPU, 16-GB vNPU type with DVPP.
	Core10Cpu4Gb16Dvpp = "10c.4cpu.16g.dvpp"
	// Core12Cpu3Gb32 represents a twelve-core, three-CPU, 32-GB vNPU type.
	Core12Cpu3Gb32 = "12c.3cpu.32g"
	// Core16 represents a sixteen-core vNPU type.
	Core16 = "16c"
)

// DCMI vNPU template names shared by device-plugin and DRA.
const (
	// Vir01 is the vir01 DCMI template.
	Vir01 = "vir01"
	// Vir02 is the vir02 DCMI template.
	Vir02 = "vir02"
	// Vir02C1 is the vir02_1c DCMI template.
	Vir02C1 = "vir02_1c"
	// Vir03C1G8 is the vir03_1c_8g DCMI template.
	Vir03C1G8 = "vir03_1c_8g"
	// Vir04 is the vir04 DCMI template.
	Vir04 = "vir04"
	// Vir04C3 is the vir04_3c DCMI template.
	Vir04C3 = "vir04_3c"
	// Vir04C3Ndvpp is the vir04_3c_ndvpp DCMI template.
	Vir04C3Ndvpp = "vir04_3c_ndvpp"
	// Vir04C4Dvpp is the vir04_4c_dvpp DCMI template.
	Vir04C4Dvpp = "vir04_4c_dvpp"
	// Vir05C1G8 is the vir05_1c_8g DCMI template.
	Vir05C1G8 = "vir05_1c_8g"
	// Vir05C1G16 is the vir05_1c_16g DCMI template.
	Vir05C1G16 = "vir05_1c_16g"
	// Vir06C1G16 is the vir06_1c_16g DCMI template.
	Vir06C1G16 = "vir06_1c_16g"
	// Vir08 is the vir08 DCMI template.
	Vir08 = "vir08"
	// Vir10C3G16 is the vir10_3c_16g DCMI template.
	Vir10C3G16 = "vir10_3c_16g"
	// Vir10C3G16NM is the vir10_3c_16g_nm DCMI template.
	Vir10C3G16NM = "vir10_3c_16g_nm"
	// Vir10C3G32 is the vir10_3c_32g DCMI template.
	Vir10C3G32 = "vir10_3c_32g"
	// Vir10C4G16M is the vir10_4c_16g_m DCMI template.
	Vir10C4G16M = "vir10_4c_16g_m"
	// Vir12C3G32 is the vir12_3c_32g DCMI template.
	Vir12C3G32 = "vir12_3c_32g"
	// Vir16 is the vir16 DCMI template.
	Vir16 = "vir16"
)

const (
	// Success for interface return code
	Success = 0
	// DeviceNotReadyErrCodeStr for dcmi interface device not ready err code string
	DeviceNotReadyErrCodeStr = "-8012"
	// DeviceNotReadyErrCode for dcmi interface device not ready err code
	DeviceNotReadyErrCode = -8012
	// CardDropFaultCode card drop fault code
	CardDropFaultCode int64 = 0x40F84E00
	// HangFaultCode NPU hang fault code
	HangFaultCode int64 = 0x200001002
	// UBSeparateFaultCode UBOE separate fault code
	UBSeparateFaultCode int64 = 0x020001002
	// UBSubHealFaultCode UB sub heal fault code
	UBSubHealFaultCode int64 = 0x020000002
	// UBOEPreSeparateFaultCode UBOE pre separate fault code
	UBOEPreSeparateFaultCode int64 = 0x110001024
	// UBOESubHealFaultCode UBOE sub heal fault code
	UBOESubHealFaultCode int64 = 0x110000002
	// LinkDownFaultCode linkdown fault code
	LinkDownFaultCode int64 = 0x81078603
	// UBOEPortDownCode uboe port down fault code
	UBOEPortDownCode int64 = 0x81078607
	// UBPortDownCode uboe port down fault code
	UBPortDownCode int64 = 0x81B18603
	// RetError return error when the function failed
	RetError = -1
	// Percent constant of 100
	Percent = 100
	// MaxErrorCodeCount number of error codes
	MaxErrorCodeCount = 128
	// UnRetError return unsigned int error
	UnRetError = math.MaxUint32
	// Abnormal status of Abnormal
	Abnormal = "Abnormal"
	// ChannelStateOk means out band channel is ok for resetting
	ChannelStateOk = 1
	// DefaultUtilizationRatePeriod default period (1s) for querying device utilization rate
	DefaultUtilizationRatePeriod = 1

	// HiAIMaxCardID max card id for Ascend chip
	HiAIMaxCardID = math.MaxInt32

	// HiAIMaxCardNum max card number
	HiAIMaxCardNum = 64

	// HiAIMaxDeviceNum max device number
	HiAIMaxDeviceNum = 4

	// NpuType present npu chip
	NpuType = 0

	// ReduceOnePercent for calculation reduce one percent
	ReduceOnePercent = 0.01
	// ReduceTenth for calculation reduce one tenth
	ReduceTenth = 0.1
	// DefaultTemperatureWhenQueryFailed when get temperature failed, use this value
	DefaultTemperatureWhenQueryFailed = -275

	// Ascend310P ascend 310P chip
	Ascend310P = "Ascend310P"
	// Ascend910 ascend 910 chip
	Ascend910 = "Ascend910"
	// Ascend910B ascend 910B chip
	Ascend910B = "Ascend910B"
	// Ascend910A3 ascend Ascend910A3 chip
	Ascend910A3 = "Ascend910A3"
	// Atlas200ISoc 200 soc env
	Atlas200ISoc = "Atlas 200I SoC A1"

	// DcmiApiTimeout dcmi interface timeout seconds
	DcmiApiTimeout = 1

	// SubscribeAllDevice subscribe all device ID
	SubscribeAllDevice = -1
	// MinVDevID min value of virtual device id
	MinVDevID = 100
	// MaxVDevID max value of virtual device id
	MaxVDevID = 1124

	// InvalidID invalid ID
	InvalidID = 0xffffffff

	// FailedMetricValue for failed metric value
	FailedMetricValue = -1

	// FailedValue for failed value
	FailedValue = math.MaxInt32

	// MaxErrorCodeLen max length of error code for Prometheus
	MaxErrorCodeLen = 10

	// DcmiRetryInterval call dcmi retry interval
	DcmiRetryInterval = 5

	// NotSupportErrorCode for not support error code
	NotSupportErrorCode = "-8255"
	// FuncNotFoundErrorCode for function missing error code
	FuncNotFoundErrorCode = "-99998"
)

const (
	// BootStartFinish chip hot reset finish
	BootStartFinish = 16

	NumberOne = 1
	NumberTwo = 2
)

const (
	// FaultRecover device fault recover
	FaultRecover = int8(0)
	// FaultOccur device fault occur
	FaultOccur = int8(1)
	// FaultOnce once device fault
	FaultOnce = int8(2)
)

const (
	// AMPMode for AMP chip work mode
	AMPMode = "AMP"
	// SMPMode for SMP chip work mode
	SMPMode = "SMP"

	// NetworkInit init status
	NetworkInit = 6
	// NetworkSuccess chip network is healthy
	NetworkSuccess = 0

	// MaxProcNum process number in device side
	MaxProcNum = 64
	// UnitMB MB
	UnitMB float64 = 1024 * 1024

	// Chip910 chip name 910
	Chip910 = "910"

	// A300IA2BoardId board id of A300I A2 and 910proB
	A300IA2BoardId = 0x28

	// A300IA2GB64BoardId board id of A300I A2 64GB
	A300IA2GB64BoardId = 0x29

	// A800IA2NoneHccsBoardIdOld board id of Atlas 800I A2 server without HCCS (old id)
	A800IA2NoneHccsBoardIdOld = 0x33
	// A800IA2NoneHccsBoardId board id of Atlas 800I A2 server without HCCS.
	// 0x33 changed to 0x3c and both stay valid for compatibility (since 2024.9.4).
	A800IA2NoneHccsBoardId = 0x3c
	// Atlas200TA2BoardId1 board id of Atlas 200T A2 BOX
	Atlas200TA2BoardId1 = 0x51
	// Atlas200TA2BoardId2 board id of Atlas 200T A2 BOX
	Atlas200TA2BoardId2 = 0x53
	// Atlas200TA2BoardId3 board id of Atlas 200T A2 BOX
	Atlas200TA2BoardId3 = 0x54

	// A900A3SuperPodBin1BoardId board id of A900/A9000 A3 SuperPod Bin1
	A900A3SuperPodBin1BoardId = 0xb0

	// A900A3SuperPodBin2BoardId board id of A900/A9000 A3 SuperPod Bin2
	A900A3SuperPodBin2BoardId = 0xb1

	// A900A3SuperPodBin3BoardId board id of A900/A9000 A3 SuperPod Bin3
	A900A3SuperPodBin3BoardId = 0xb2

	// A800IA3BoardId board id of A800I A3
	A800IA3BoardId = 0xb3

	// A800IA3BoardId2 board id of A800I A3 additional SKU
	A800IA3BoardId2 = 0xb4

	// A900A3SuperPodMainBoardId1 board id of A900 A3 SuperPod MainBoard1
	A900A3SuperPodMainBoardId1 = 0x18

	// A900A3SuperPodMainBoardId2 board id of A900 A3 SuperPod MainBoard2
	A900A3SuperPodMainBoardId2 = 0x19

	// A800IA3MainBoardId A800I A3 MainBoardId
	A800IA3MainBoardId = 0x14

	// A9000A3SuperPodMainBoardId1 board id of A9000 A3 SuperPod MainBoard1
	A9000A3SuperPodMainBoardId1 = 0x1C

	// A9000A3SuperPodMainBoardId2 board id of A9000 A3 SuperPod MainBoard2
	A9000A3SuperPodMainBoardId2 = 0x1D

	// Atlas200LA2ZQBoardId board id of Atlas 200L A2 ZQ
	Atlas200LA2ZQBoardId = 0x69

	// A3SuperPodZQBoardId board id of A3 SuperPod ZQ (Zuque SuperPod)
	A3SuperPodZQBoardId = 0x81

	// A3ServerZQBoardId board id of A3 Server ZQ (Zuque Server)
	A3ServerZQBoardId = 0x83

	// A3SuperPodZQNpuBoardId is the board ID of A3 SuperPod ZQ (Zuque SuperPod).
	A3SuperPodZQNpuBoardId = 0xd1

	// A3ServerZQNpuBoardId is the board ID of A3 Server ZQ (Zuque Server).
	A3ServerZQNpuBoardId = 0xd3
)

// log limit domains for metrics
const (
	// DomainForLogicIdErr domain for failed to get cardId and deviceId by logicID
	DomainForLogicIdErr = "logicID"
)

// DcmiDeviceType used to represent the dcmi device type
type DcmiDeviceType int32

const (
	// DcmiDeviceTypeDDR represents the component type DCMI_DEVICE_TYPE_DDR
	DcmiDeviceTypeDDR DcmiDeviceType = 0
	// DcmiDeviceTypeSRAM represents the component type DCMI_DEVICE_TYPE_SRAM
	DcmiDeviceTypeSRAM DcmiDeviceType = 1
	// DcmiDeviceTypeHBM represents the component type DCMI_DEVICE_TYPE_HBM
	DcmiDeviceTypeHBM DcmiDeviceType = 2
	// DcmiDeviceTypeNPU represents the component type DCMI_DEVICE_TYPE_NPU
	DcmiDeviceTypeNPU DcmiDeviceType = 3
	// DcmiDeviceTypeNONE represents the component type DCMI_DEVICE_TYPE_NONE
	DcmiDeviceTypeNONE DcmiDeviceType = 0xff
)

const (
	// ErrMsgInitCardListFailed is used where initialization of the card list fails
	ErrMsgInitCardListFailed = "get card list failed for init"
	// ErrMsgInitDeviceListFailed is used where initialization of the device list fails
	ErrMsgInitDeviceListFailed = "get device list failed for init"
	// ErrMsgGetBoardInfoFailed is used where there is a failure in getting board info
	ErrMsgGetBoardInfoFailed = "get board info failed, no card found"
)

const (
	// MaxHccspingMeshAddr is the max number of hccsping addresses
	MaxHccspingMeshAddr = 1024
	// MinPktSize is the min packet size
	MinPktSize = 1792
	// MaxPktSize is the max packet size
	MaxPktSize = 3000
	// MinPktSendNum is the min packet send number
	MinPktSendNum = 1
	// MaxPktSendNum is the max packet send number
	MaxPktSendNum = 1000
	// MinPktInterval is the min packet interval
	MinPktInterval = 1
	// MaxPktInterval is the max packet interval
	MaxPktInterval = 1000
	// MinTaskInterval is the min task interval
	MinTaskInterval = 1
	// MaxTaskInterval is the max task interval
	MaxTaskInterval = 60
	// InternalPingMeshTaskID is the inner ping mesh task id; for A5 is the inner super pod task id
	InternalPingMeshTaskID uint = 0
	// ExternalPingMeshTaskID is the outer ping mesh task id
	ExternalPingMeshTaskID uint = 1
	// DefaultPingMeshPortID is the default ping mesh port
	DefaultPingMeshPortID = 0
	// DefaultPktSize is the default packet size
	DefaultPktSize = 1792
	// DefaultPktSendNum is the default packet send number
	DefaultPktSendNum = 10
	// DefaultPktInterval is the default packet interval
	DefaultPktInterval = 10
	// DefaultTimeout is the default timeout
	DefaultTimeout = 1
)

const (
	// NPUNetworkLinkDownStatus indicate the network status of down
	NPUNetworkLinkDownStatus = "DOWN"
	// NPUNetworkLinkUpStatus indicate the network status of up
	NPUNetworkLinkUpStatus = "UP"
	// PortNoDownCount indicate no port down count
	PortNoDownCount = 0
	// RoceParameterPlanePortAllDownCount indicate the network port all up count
	RoceParameterPlanePortAllDownCount = 1
	// UBOEParameterPlanePortAllDownCount indicate the network port all up count for UBOE
	UBOEParameterPlanePortAllDownCount = 2
)

// constants for k8s
const (
	// ReplaceOP is the replace operation
	ReplaceOP = "replace"
	// AddOP is the add operation
	AddOP = "add"

	// PathForAnnotations is the path for annotations
	PathForAnnotations = "/metadata/annotations/"
	// PathForLabels is the path for labels
	PathForLabels = "/metadata/labels/"
)
