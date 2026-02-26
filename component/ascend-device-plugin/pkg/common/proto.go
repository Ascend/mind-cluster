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
	"sync"

	"github.com/fsnotify/fsnotify"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/api"
)

var (
	// ParamOption for option
	ParamOption Option
	// DpStartReset for reset configmap
	DpStartReset sync.Once
)

// NodeDeviceInfoCache record node NPU device information. Will be solidified into cm.
type NodeDeviceInfoCache struct {
	DeviceInfo  NodeDeviceInfo
	SuperPodID  int32
	ServerIndex int32
	CheckCode   string
	RackID      *int32 `json:"RackID,omitempty"`
}

// SwitchFaultEvent is the struct for switch reported fault
type SwitchFaultEvent struct {
	EventType uint
	// SubType fault subtype used for id a fault
	SubType uint
	// FaultID the fault id for switch fault
	FaultID string
	// AssembledFaultCode is to assemble cgo lq.struct_LqDcmiEvent to a device-plugin recognized fault code type
	// such as : [0x00f103b0,155904,na,na] in config file: SwitchFaultCode
	AssembledFaultCode string
	// PeerPortDevice used to tell what kind of device connected to
	PeerPortDevice uint
	PeerPortId     uint
	SwitchChipId   uint
	SwitchPortId   uint
	// Severity used to tell how serious is the fault
	Severity uint
	// Assertion tell what kind of fault, recover, happen or once
	Assertion       uint
	EventSerialNum  int
	NotifySerialNum int
	AlarmRaisedTime int64
	AdditionalParam string
	AdditionalInfo  string
}

// SwitchFaultInfo Switch Fault Info
type SwitchFaultInfo struct {
	FaultCode            []string
	FaultLevel           string
	UpdateTime           int64
	NodeStatus           string
	FaultTimeAndLevelMap map[string]FaultTimeAndLevel
}

// NodeDeviceInfo record node NPU device information. Will be solidified into cm.
type NodeDeviceInfo struct {
	DeviceList map[string]string
	UpdateTime int64
}

// DeviceHealth health status of device
type DeviceHealth struct {
	FaultCodes    []int64
	Health        string
	NetworkHealth string
}

// NpuAllInfo all npu infos
type NpuAllInfo struct {
	AllDevTypes []string
	AllDevs     []NpuDevice
	AICoreDevs  []*NpuDevice
}

// NpuDevice npu device description
type NpuDevice struct {
	FaultCodes             []int64
	AlarmRaisedTime        int64
	NetworkFaultCodes      []int64
	NetworkAlarmRaisedTime int64
	FaultTimeMap           map[int64]int64
	DevType                string
	DeviceName             string
	Health                 string
	NetworkHealth          string
	DpuHealth              string
	CardDrop               bool
	IP                     string
	LogicID                int32
	PhyID                  int32
	CardID                 int32
	DeviceID               int32
	SuperDeviceID          uint32
	Status                 string
	PodUsed                bool            // PodUsed indicates whether this chip is used by kubelet
	LevelList              []api.RankLevel // rank level info in rank table for A5
	UsedAicoreQuota        int             // used aicore quota
	UsedHbmQuota           int             // used hbm quota
}

// DavinCiDev davinci device
type DavinCiDev struct {
	IP       string
	LogicID  int32
	PhyID    int32
	CardID   int32
	DeviceID int32
}

// Device id for Instcance
type Device struct { // Device
	DeviceID      string `json:"device_id"` // device id
	DeviceIP      string `json:"device_ip"` // device ip
	SuperDeviceID string `json:"super_device_id,omitempty"`
	// rank level info in rank table for A5
	LevelList []api.RankLevel `json:"levelList,omitempty"`
}

// Instance is for annotation
type Instance struct { // Instance
	PodName     string   `json:"pod_name"`  // pod Name
	ServerID    string   `json:"server_id"` // serverdId
	HostIP      string   `json:"host_ip"`   // hostIp
	SuperPodId  int32    `json:"super_pod_id"`
	Devices     []Device `json:"devices"` // dev
	RackId      int32    `json:"rack_id"`
	ServerIndex string   `json:"server_index"` // serverIndex for A5
	ServerIP    string   `json:"server_ip"`
}

// Option option
type Option struct {
	GetFdFlag             bool     // to describe FdFlag
	UseAscendDocker       bool     // UseAscendDocker to choose docker type
	UseVolcanoType        bool     // use volcano mode
	AutoStowingDevs       bool     // auto stowing fixes devices or not
	PresetVDevice         bool     // preset virtual device
	Use310PMixedInsert    bool     // chose 310P mixed insert mode
	GraceToleranceOn      bool     // check if grace tolerance is on
	ListAndWatchPeriod    int      // set listening device state period
	HotReset              int      // unhealthy chip hot reset
	ShareCount            uint     // share device count
	AiCoreCount           int32    // found by dcmi interface
	BuildScene            string   // build scene judge device-plugin start scene
	ProductTypes          []string // all product types
	RealCardType          string   // real card type
	CardType              string   // standard card type
	LinkdownTimeout       int64    // linkdown timeout duration
	DealWatchHandler      bool     // update pod cache when receiving pod informer watch errors
	EnableSwitchFault     bool     // if enable switch fault
	CheckCachedPods       bool     // check cached pods periodically
	EnableSlowNode        bool     // switch of set slow node notice environment
	ThirdPartyScanDelay   int      // delay time (seconds) for scanning devices when third party resetting
	DeviceResetTimeout    int      // device reset timeout
	SoftShareDevConfigDir string   // soft share device config dir
}

// GetAllDeviceInfoTypeList Get All Device Info Type List
func GetAllDeviceInfoTypeList() map[string]struct{} {
	return map[string]struct{}{
		GetAscend910Key(""):                               {},
		GetAscend910Key(api.CmCardUnhealthySuffix):        {},
		GetAscend910Key(api.CmCardNetworkUnhealthySuffix): {},
		HuaweiUnHealthAscend910:                           {}, HuaweiNetworkUnHealthAscend910: {},
		api.ResourceNamePrefix + api.Ascend910: {}, api.ResourceNamePrefix + Ascend910vir2: {},
		api.ResourceNamePrefix + Ascend910vir4: {}, api.ResourceNamePrefix + Ascend910vir8: {},
		api.ResourceNamePrefix + Ascend910vir16: {}, api.ResourceNamePrefix + Ascend910vir5Cpu1Gb8: {},
		api.ResourceNamePrefix + Ascend910vir5Cpu1Gb16: {}, api.ResourceNamePrefix + Ascend910vir6Cpu1Gb16: {},
		api.ResourceNamePrefix + Ascend910vir10Cpu3Gb16: {}, api.ResourceNamePrefix + Ascend910vir3Cpu1Gb8: {},
		api.ResourceNamePrefix + Ascend910vir10Cpu3Gb16Ndvpp: {}, api.ResourceNamePrefix + Ascend910vir10Cpu3Gb32: {},
		api.ResourceNamePrefix + Ascend910vir10Cpu4Gb16Dvpp: {},
		api.ResourceNamePrefix + Ascend910vir12Cpu3Gb32:     {}, api.ResourceNamePrefix + api.Ascend310: {},
		api.ResourceNamePrefix + api.Ascend310P: {}, api.ResourceNamePrefix + Ascend310Pc1: {},
		api.ResourceNamePrefix + Ascend310Pc2: {}, api.ResourceNamePrefix + Ascend310Pc4: {},
		api.ResourceNamePrefix + Ascend310Pc2Cpu1: {}, api.ResourceNamePrefix + Ascend310Pc4Cpu3: {},
		api.ResourceNamePrefix + Ascend310Pc4Cpu3Ndvpp: {}, api.ResourceNamePrefix + Ascend310Pc4Cpu4Dvpp: {},
		HuaweiUnHealthAscend310P: {}, HuaweiUnHealthAscend310: {}, api.ResourceNamePrefix + AiCoreResourceName: {}}
}

// FileWatch is used to watch sock file
type FileWatch struct {
	FileWatcher *fsnotify.Watcher
}

// DevStatusSet contain different states devices
type DevStatusSet struct {
	UnHealthyDevice    sets.String
	NetUnHealthyDevice sets.String
	DpuUnHealthyDevice sets.String
	HealthDevices      sets.String
	RecoveringDevices  sets.String
	FreeHealthyDevice  map[string]sets.String
	DeviceFault        []DeviceFault
	AllDevices         sets.String
}

// FaultTimeAndLevel of each fault code
type FaultTimeAndLevel struct {
	FaultTime  int64  `json:"fault_time"`
	FaultLevel string `json:"fault_level"`
}

// DeviceFault  npu or network fault info
type DeviceFault struct {
	FaultType            string                       `json:"fault_type"`
	NPUName              string                       `json:"npu_name"`
	LargeModelFaultLevel string                       `json:"large_model_fault_level"`
	FaultLevel           string                       `json:"fault_level"`
	FaultHandling        string                       `json:"fault_handling"`
	FaultCode            string                       `json:"fault_code"`
	FaultTimeAndLevelMap map[string]FaultTimeAndLevel `json:"fault_time_and_level_map"`
}

// TaskResetInfoCache record task reset device information cache
type TaskResetInfoCache struct {
	ResetInfo *TaskResetInfo
	CheckCode string
}

// TaskResetInfo record task reset device information
type TaskResetInfo struct {
	RankList            []*TaskDevInfo
	UpdateTime          int64
	RetryTime           int
	FaultFlushing       bool
	GracefulExit        int
	RestartFaultProcess bool
}

// TaskDevInfo is the device info of a task
type TaskDevInfo struct {
	RankId int
	DevFaultInfo
}

// DevFaultInfo is the fault info of device
type DevFaultInfo struct {
	LogicId       int32
	Status        string
	Policy        string
	InitialPolicy string
	ErrorCode     []int64
	ErrorCodeHex  string
}

// TaskFaultInfoCache record task fault rank information cache
type TaskFaultInfoCache struct {
	FaultInfo *TaskFaultInfo
	CheckCode string
}

// TaskFaultInfo record task fault rank information
type TaskFaultInfo struct {
	FaultRank  []int
	UpdateTime int64
}

// SuperPodInfo is super pod info
type SuperPodInfo struct {
	ScaleType  int32
	SuperPodId int32
	ServerId   int32
	RackId     int32
	// SuperPodType for A5
	SuperPodType int8
	Reserve      []int8
}

// Get310PProductType get 310P product type
func Get310PProductType() map[string]string {
	return map[string]string{
		"Atlas 300V Pro": Ascend310PVPro,
		"Atlas 300V":     Ascend310PV,
		"Atlas 300I Pro": Ascend310PIPro,
	}
}

// HasOnChipMemory check if card has on-chip-memory
func HasOnChipMemory() bool {
	return ParamOption.RealCardType == api.Ascend910B ||
		ParamOption.RealCardType == api.Ascend910A ||
		ParamOption.RealCardType == api.Ascend910A3 ||
		ParamOption.RealCardType == api.Ascend910A5
}

// PodDeviceInfo define device info of pod, include kubelet allocate and real allocate device
type PodDeviceInfo struct {
	Pod        v1.Pod
	KltDevice  []string
	RealDevice []string
}

// NpuBaseInfo is the base info of npu
type NpuBaseInfo struct {
	IP            string
	SuperDeviceID uint32
	// LevelList info for A5
	LevelList []api.RankLevel `json:"levelList,omitempty"`
}
