// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"k8s.io/apimachinery/pkg/util/sets"

	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
)

// FaultProcessCenter processes the faults and coordinates the fault handling among different components.
type FaultProcessCenter struct {
	deviceCenter      *map[string]*constant.DeviceInfo
	nodeCenter        *map[string]*constant.SwitchInfo
	switchCenter      *map[string]*constant.NodeInfo
	faultJobCenter    *faultJobProcessCenter
	notifyProcessChan chan int
}

type faultJobProcessCenter struct {
	jobServerInfoMap job.ServerInfoMap
	lastProcessTime  int64
	deviceInfoCm     map[string]*constant.DeviceInfo
	switchInfoCm     map[string]*constant.SwitchInfo
	FaultJobs        map[string]*FaultJob
}

type FaultJob struct {
	IsA3Job             bool
	NameSpace           string
	PodNames            map[string]string
	RelationFaults      []*faultInfo
	TriggerFault        []faultInfo
	processedFaultInfo  []faultInfo
	FaultStrategy       FaultStrategy
	SeparateNodes       sets.String
	AllFaultCode        sets.String
	ProcessingFaultCode sets.String
	PodStrategiesMaps   map[string]string
	FindNPUUnderSwitch  bool
}

type faultInfo struct {
	FaultUid         string
	FaultType        string
	NodeName         string
	NPUName          string
	FaultCode        string
	FaultLevel       string
	FaultTime        int64
	ExecutedStrategy string
	DealMaxTime      int64
}

type simpleSwitchFaultInfo struct {
	EventType          uint
	AssembledFaultCode string
	PeerPortDevice     uint
	PeerPortId         uint
	SwitchChipId       uint
	SwitchPortId       uint
	Severity           uint
	Assertion          uint
	AlarmRaisedTime    int64
}

// AdvanceDeviceFaultCm more structure device info
type AdvanceDeviceFaultCm struct {
	ServerType       string
	CmName           string
	SuperPodID       int32
	ServerIndex      int32
	FaultDeviceList  map[string][]constant.DeviceFault
	CardUnHealthy    []string
	NetworkUnhealthy []string
	UpdateTime       int64
}

// FaultRank defines the structure for storing fault rank information.
// It includes the rank ID and fault code.
type FaultRank struct {
	RankId      string
	FaultCode   string
	FaultLevel  string
	DoStepRetry bool
}

// JobFaultInfo job fault rank info
type JobFaultInfo struct {
	JobId     string
	FaultList []FaultRank
}

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
	SubHealthFault       = "SubHealthFault"
)

// cluster support server
const (
	Ascend910Server  = "Ascend910"
	Ascend310PServer = "Ascend310P"
	Ascend310Server  = "Ascend310"
)

const (
	invalidSuperPodIndex    = -2
	patchPodTimes           = 3
	faultJobProcessInterval = 5 * 1000
	allCardId               = "FF"
	switchFaultType         = "switchFault"
	deviceFaultType         = "deviceFault"
	nodeFaultType           = "nodeFault"
	nodeUnhealthy           = "UnHealthy"
	triggerFaultType        = "TriggerFault"
	relationFaultType       = "RelationFaultCodes"
	taskFaultKey            = "fault-type"
	kilo                    = 1000
	faultCustomizationPath  = "/home/hwMindX/relationFaultCustomization.json"
	faultDuration           = "/home/hwMindX/faultDuration.json"
)

// FaultStrategy fault strategies
type FaultStrategy struct {
	NodeLvList   map[string]string
	DeviceLvList map[string][]DeviceStrategy
}

// RelationFaultStrategy relation fault strategy
type RelationFaultStrategy struct {
	TriggerFault   string
	RelationFaults []string
	FaultStrategy  string
}

// FaultDuration fault duration config
type FaultDuration struct {
	FaultCode       string
	FaultType       string
	TimeOutInterval int64
}

// DeviceStrategy device fault strategy
type DeviceStrategy struct {
	Strategy string
	NPUName  string
}
