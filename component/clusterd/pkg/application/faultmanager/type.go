// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"sync"

	"clusterd/pkg/common/constant"
)

type faultProcessor interface {
	process()
}

type baseFaultCenter[T constant.ConfigMapInterface] struct {
	processorList        []faultProcessor
	lastProcessTime      int64
	subscribeChannelList []chan int
	mutex                sync.Mutex
	processPeriod        int64
	jobServerInfoMap     constant.JobServerInfoMap
	cmManager            *faultCenterCmManager[T]
	centerType           int
}

// deviceFaultProcessCenter
type deviceFaultProcessCenter struct {
	baseFaultCenter[*constant.DeviceInfo]
}

// GlobalFaultProcessCenter is a global instance of FaultProcessCenter used for processing faults.
var GlobalFaultProcessCenter *FaultProcessCenter

// FaultProcessCenter processes the faults and coordinates the fault handling among different components.
type FaultProcessCenter struct {
	deviceCenter      *deviceFaultProcessCenter
	nodeCenter        *nodeFaultProcessCenter
	switchCenter      *switchFaultProcessCenter
	notifyProcessChan chan int
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

type linkDownCqeFaultProcessCenter struct {
	deviceCenter        *deviceFaultProcessCenter
	linkDownCqeFaults   map[string]map[string]map[string]cqeLinkDownFaultRank // job, node, device faultInfo
	nodeDeviceFaultInfo map[string]AdvanceDeviceFaultCm
	cqeFaultTimeList    map[string][]int64
}

type cqeLinkDownFaultRank struct {
	LinkDownFaultTime int64
	DeviceName        string
	IsLinkDown        bool
	IsCqe             bool
}

type jobRankFaultInfoProcessor struct {
	deviceCenter    *deviceFaultProcessCenter
	jobFaultInfoMap map[string]JobFaultInfo
	mutex           sync.RWMutex
}

// nodeFaultProcessCenter
type nodeFaultProcessCenter struct {
	baseFaultCenter[*constant.NodeInfo]
}

type switchFaultProcessCenter struct {
	baseFaultCenter[*constant.SwitchInfo]
}

// uceAccompanyFaultProcessor:
// aic aiv fault can be 1) accompanied by uce fault, also can 2) curr alone.
// if 1) aic aiv fault should be filtered. Once find aic fault, check if there is an uce fault 5s ago
// if 2) aic aiv fault should not be retained.
type uceAccompanyFaultProcessor struct {
	deviceCenter *deviceFaultProcessCenter
	// maintain 5s ago device info
	DiagnosisAccompanyTimeout int64
	// nodeName -> deviceName -> faultQue
	uceAccompanyFaultQue map[string]map[string][]constant.DeviceFault
	// uceFaultTime
	uceFaultTime       map[string]map[string]int64
	deviceCmForNodeMap map[string]AdvanceDeviceFaultCm
}

/*
The uceFaultProcessor process uce fault reporting information.
If the device fault is UCE fault, then determine whether the job running on the device can tolerate UCE faults.
If they can tolerate it, the reporting of the UCE fault should be delayed by 10 seconds.
*/
type uceFaultProcessor struct {
	deviceCenter             *deviceFaultProcessCenter
	JobReportRecoverTimeout  int64
	JobReportCompleteTimeout int64

	reportInfo *reportInfosForAllJobs
	// uceJob->jobInfo
	uceDevicesOfUceJob map[string]uceJobInfo
	// node->DeviceName->uceDeviceInfo
	uceDeviceOfNode  map[string]uceNodeInfo
	jobServerInfoMap constant.JobServerInfoMap
	nodeDeviceCmMap  map[string]AdvanceDeviceFaultCm
}

// JobId->node->device->report_info
type reportInfosForAllJobs struct {
	InfoMap map[string]map[string]map[string]reportInfo
	RwMutex sync.RWMutex
}

type uceDeviceInfo struct {
	// DeviceName has prefix Ascend910
	DeviceName   string
	FaultTime    int64
	RecoverTime  int64
	CompleteTime int64
}

type uceNodeInfo struct {
	NodeName string
	// DeviceName->DeviceInfo
	DeviceInfo map[string]uceDeviceInfo
}

type uceJobInfo struct {
	// UceNode node->nodeInfo
	UceNode map[string]uceNodeInfo
	JobId   string
}

type reportInfo struct {
	RecoverTime  int64
	CompleteTime int64
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

type configMap[T constant.ConfigMapInterface] struct {
	configmap map[string]T
}

type faultCenterCmManager[T constant.ConfigMapInterface] struct {
	mutex        sync.RWMutex
	originalCm   configMap[T]
	processingCm configMap[T]
	processedCm  configMap[T]
}
