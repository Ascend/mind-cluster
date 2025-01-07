/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package plugin is using for HuaWei Ascend pin affinity schedule.
package plugin

import (
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/config"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const (
	// PluginName the HuaWei NPU 's plugin name.
	PluginName = "huaweiNPU"

	nodesNoMeetNPUReqError = "insufficient npus on the schedulable nodes in cluster"
	objectNilError         = "object or argument is nil"
	podRankIndex           = "hccl/rankIndex"

	// FormatIncorrectError format incorrect error
	FormatIncorrectError = "format incorrect"

	// AscendVNPULevel vnpu level
	AscendVNPULevel = "vnpu-level"
	// AscendVNPULevelLow low
	AscendVNPULevelLow = "low"
	// AscendVNPULevelHigh high
	AscendVNPULevelHigh = "high"
	// AscendVNPUPrefix vir
	AscendVNPUPrefix = "vir"
	// AscendVNPUDVPP dvpp enable
	AscendVNPUDVPP = "vnpu-dvpp"
	// AscendDVPPEnabledOff off
	AscendDVPPEnabledOff = "no"
	// AscendDVPPEnabledNull null
	AscendDVPPEnabledNull = "null"
	// AscendDVPPEnabledOn on
	AscendDVPPEnabledOn = "yes"
	// AscendNDVPPValue value
	AscendNDVPPValue = "ndvpp"
	// AscendDVPPValue value
	AscendDVPPValue = "dvpp"
	// VNPUTempVir01 vir01
	VNPUTempVir01 = "vir01"
	// VNPUTempVir02 vir02
	VNPUTempVir02 = "vir02"
	// VNPUTempVir02C1 vir02_1c
	VNPUTempVir02C1 = "vir02_1c"
	// VNPUTempVir04  vir04
	VNPUTempVir04 = "vir04"
	// VNPUTempVir04C3 vir04_3c
	VNPUTempVir04C3 = "vir04_3c"
	// VNPUTempVir04C3NDVPP vir04_3c_ndvpp
	VNPUTempVir04C3NDVPP = "vir04_3c_ndvpp"
	// VNPUTempVir04C4cDVPP vir04_4c_dvpp
	VNPUTempVir04C4cDVPP = "vir04_4c_dvpp"
	// VNPUTempVir08  vir08 only 910
	VNPUTempVir08 = "vir08"
	// VNPUTempVir16  vir16 only 910
	VNPUTempVir16 = "vir16"
	// Ascend310P 310P template name
	Ascend310P = "Ascend310P"
	// Ascend910 910 template name
	Ascend910                  = "Ascend910"
	maxTorAffinityNodeScore    = float64(200)
	halfTorAffinityNodeScore   = float64(100)
	sharedTorAffinityNodeScore = float64(99)
	cardHealthySuffix          = ""
	unhealthyCardSuffix        = "-Unhealthy"
	notNPUNodeError            = "getNodeDeviceInfoFromCM"
	notNPUJobError             = "nil npu"
	basePlugin                 = "base"
	oneTor                     = 1
	twoTor                     = 2
	defaultResyncTime          = 30
	// ResetInfoCMNamePrefix for reset configmap name prefix
	ResetInfoCMNamePrefix = "reset-config-"
	// ResetInfoCMDataKey for reset configmap data key
	ResetInfoCMDataKey = "reset.json"
	// ResetInfoTypeKey for reset configmap type key
	ResetInfoTypeKey = "restartType"
	// PodRescheduleRestartType for hot reset restart type
	PodRescheduleRestartType = "podReschedule"
	normalNodeErr            = "not NPU node"
	oldCapacity              = "Capability"
	newCapacity              = "Capacity"
)

// SchedulerJob the plugin define job info
type SchedulerJob struct {
	util.SchedulerJobAttr
	RankIndexInfo
	UnschedulableReason
	handler      ISchedulerPlugin
	ServerList   []*Tor
	TorBlackMaps map[string]struct{}
	JobReadyTag  bool
	SuperPods    map[string][]SuperNode
	Owner        OwnerInfo
}

// OwnerInfo the owner info of job
type OwnerInfo struct {
	v1.OwnerReference
	Annotations map[string]string
	Replicas    *int32
}

// UnschedulableReason the message of pod pending
type UnschedulableReason struct {
	Reason map[string]map[string]struct{}
	*sync.Mutex
}

// SuperNode node with SuperPodID
type SuperNode struct {
	Name       string
	SuperPodID int32
}

// RankIndexInfo the info of job used rank
type RankIndexInfo struct {
	HealthTorRankIndex map[string]string
	FaultRankIndex     map[int]struct{}
}

// VolcanoFrame passed in by the volcano frame.
type VolcanoFrame struct {
	UID            types.UID
	Confs          []config.Configuration
	KubeClient     kubernetes.Interface
	VJobTemplate   map[string]map[string]util.VResource
	SuperPodSize   int
	ReservePodSize int
}

// NslbParameters the Parameters os nslb
type NslbParameters struct {
	nslbVersion  string
	sharedTorNum int
}

// ScheduleCache the plugin defined caches saving cm data
type ScheduleCache struct {
	// special, name, value
	Names, Namespaces map[string]string
	Data              map[string]map[string]string
}

// ScheduleEnv for job scheduler context.
type ScheduleEnv struct {
	IsFirstSession    *bool // scheduler first session message is unreliable
	Jobs              map[api.JobID]SchedulerJob
	JobReplicas       map[api.JobID]int32
	Nodes             map[string]NPUNode
	NodesNotInSsn     map[string]*corev1.Node
	JobSinglePodFlag  map[api.JobID]bool
	JobSeverInfos     map[api.JobID]struct{}
	JobDeleteFlag     map[api.JobID]struct{}
	DeviceInfos       *DeviceInfosWithMutex
	DeleteJobInfos    map[api.JobID]*api.JobInfo
	NodeInfosFromCm   *NodeInfosFromCmWithMutex   // NodeInfos is get from kube-system/node-info- configmap
	SwitchInfosFromCm *SwitchInfosFromCmWithMutex // SwitchInfosFromCm is get from mindx-dl/device-info- configmap
	FrameAttr         VolcanoFrame
	Cache             ScheduleCache
	Tors              *TorList
	NslbAttr          *NslbParameters
	SuperPodInfo      *SuperPodInfo
	JobPendingMessage map[api.JobID]map[string]map[string]struct{}
}

// SuperPodInfo cache super pod info for pod rescheduling
type SuperPodInfo struct {
	SuperPodReschdInfo        map[api.JobID]map[string][]SuperNode // cache super pod re-schd info
	SuperPodFaultTaskNodes    map[api.JobID][]string               // cache fault task nodes info
	SuperPodMapFaultTaskNodes map[api.JobID]map[string]string      // cache task and nodes for stage2
}

// DeviceInfosWithMutex information for the current plugin
type DeviceInfosWithMutex struct {
	sync.Mutex
	Devices map[string]NodeDeviceInfoWithID
}

// NodeInfosFromCmWithMutex node info with mutex
type NodeInfosFromCmWithMutex struct {
	sync.Mutex
	Nodes map[string]NodeDNodeInfo
}

// SwitchInfosFromCmWithMutex SwitchInfos From Cm WithMutex
type SwitchInfosFromCmWithMutex struct {
	sync.Mutex
	Switches map[string]SwitchFaultInfo
}

// ScheduleHandler information for the current plugin
type ScheduleHandler struct {
	NPUPlugins map[string]NPUBuilder
	ScheduleEnv
	BaseHandle ISchedulerPlugin
	sync.Once
}

// AllocNodeRankOccurrence object recording node rankIndex and whether index re-allocated to new node
type AllocNodeRankOccurrence struct {
	NodeName   string
	RankIndex  string
	IsFault    bool
	Occurrence int
}

type jobUsedNodeInfos struct {
	NodeInfos string
	JobName   string
}

type jobServerInfos struct {
	IsSharedTor bool
	Nodes       []jobUsedNodeInfos
}

type jobTorInfos struct {
	usedHealthyTor []*Tor
	otherTor       []*Tor
	torNums        map[string]int
	usedAllTorNum  int
}

type usedTorInfos struct {
	sharedTorNum   int
	isSingleTorJob bool
	usedTors       map[string]*Tor
}

// TaskResetInfo record task reset device information
type TaskResetInfo struct {
	RankList      []*TaskDevInfo
	UpdateTime    int64
	RetryTime     int
	FaultFlushing bool
	GracefulExit  int
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
