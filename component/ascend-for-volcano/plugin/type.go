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
	Ascend910               = "Ascend910"
	maxTorAffinityNodeScore = float64(200)
	cardHealthySuffix       = ""
	unhealthyCardSuffix     = "-Unhealthy"
	notNPUNodeError         = "getNodeDeviceInfoFromCM"
	notNPUJobError          = "nil npu"
	basePlugin              = "base"

	convertFailedErrorPattern = "Cannot convert to ConfigMap:%#v"
)

// SchedulerJob the plugin define job info
type SchedulerJob struct {
	util.SchedulerJobAttr
	RankIndexInfo
	handler     ISchedulerPlugin
	ServerList  []*Tor
	JobReadyTag bool
}

// RankIndexInfo the info of job used rank
type RankIndexInfo struct {
	HealthTorRankIndex map[string]string
	FaultRankIndex     map[int]struct{}
}

// VolcanoFrame passed in by the volcano frame.
type VolcanoFrame struct {
	UID          types.UID
	Confs        []config.Configuration
	KubeClient   kubernetes.Interface
	VJobTemplate map[string]map[string]util.VResource
}

// ScheduleCache the plugin defined caches saving cm data
type ScheduleCache struct {
	// special, name, value
	Names, Namespaces map[string]string
	Data              map[string]map[string]string
	FaultConfigMaps   map[api.JobID]*FaultRankIdData
}

type FaultRankIdData struct {
	Name, Namespace string
	Data            map[string]string
}

// ScheduleEnv for job scheduler context.
type ScheduleEnv struct {
	Jobs            map[api.JobID]SchedulerJob
	Nodes           map[string]NPUNode
	JobSeverInfos   map[api.JobID]struct{}
	DeviceInfos     *DeviceInfosWithMutex
	DeleteJobInfos  map[api.JobID]*api.JobInfo
	NodeInfosFromCm *NodeInfosFromCmWithMutex
	FrameAttr       VolcanoFrame
	Cache           ScheduleCache
	Tors            *TorList
}

// DeviceInfosWithMutex information for the current plugin
type DeviceInfosWithMutex struct {
	sync.Mutex
	Devices map[string]NodeDeviceInfo
}

// NodeInfosFromCmWithMutex node info with mutex
type NodeInfosFromCmWithMutex struct {
	sync.Mutex
	Nodes map[string]NodeDNodeInfo
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
