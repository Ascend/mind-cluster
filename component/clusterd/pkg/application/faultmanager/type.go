// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"

	"clusterd/pkg/application/faultmanager/collector"
	"clusterd/pkg/common/constant"
)

type baseFaultCenter[T constant.ConfigMapInterface] struct {
	processorList        []constant.FaultProcessor
	lastProcessTime      int64
	subscribeChannelList []chan int
	mutex                sync.Mutex
	processPeriod        int64
	cmManager            *faultCenterCmManager[T]
	centerType           int
}

// DeviceFaultProcessCenter
type DeviceFaultProcessCenter struct {
	baseFaultCenter[*constant.DeviceInfo]
}

// GlobalFaultProcessCenter is a global instance of FaultProcessCenter used for processing faults.
var GlobalFaultProcessCenter *FaultProcessCenter

// FaultProcessCenter processes the faults and coordinates the fault handling among different components.
type FaultProcessCenter struct {
	DeviceCenter      *DeviceFaultProcessCenter
	NodeCenter        *NodeFaultProcessCenter
	SwitchCenter      *SwitchFaultProcessCenter
	FaultJobCenter    *faultJobProcessCenter
	faultJobProcessor *faultProcessorImpl
	jobServerInfoMap  constant.JobServerInfoMap
	NotifyProcessChan chan int
}

type faultJobProcessCenter struct {
	jobServerInfoMap constant.JobServerInfoMap
	lastProcessTime  int64
	deviceInfoCm     map[string]*constant.DeviceInfo
	switchInfoCm     map[string]*constant.SwitchInfo
	nodeInfoCm       map[string]*constant.NodeInfo
	FaultJobs        map[string]*FaultJob
}

// FaultJob contain some fault info about a fault job
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
	deviceCenter        *DeviceFaultProcessCenter
	linkDownCqeFaults   map[string]map[string]map[string]cqeLinkDownFaultRank // job, node, device faultInfo
	nodeDeviceFaultInfo map[string]constant.AdvanceDeviceFaultCm
	cqeFaultTimeList    map[string][]int64
}

type cqeLinkDownFaultRank struct {
	LinkDownFaultTime int64
	DeviceName        string
	IsLinkDown        bool
	IsCqe             bool
}

type jobRankFaultInfoProcessor struct {
	deviceCenter    *DeviceFaultProcessCenter
	jobFaultInfoMap map[string]JobFaultInfo
	mutex           sync.RWMutex
}

// NodeFaultProcessCenter
type NodeFaultProcessCenter struct {
	baseFaultCenter[*constant.NodeInfo]
}

type SwitchFaultProcessCenter struct {
	baseFaultCenter[*constant.SwitchInfo]
}

// uceAccompanyFaultProcessor:
// aic aiv fault can be 1) accompanied by uce fault, also can 2) curr alone.
// if 1) aic aiv fault should be filtered. Once find aic fault, check if there is an uce fault 5s ago
// if 2) aic aiv fault should not be retained.
type uceAccompanyFaultProcessor struct {
	deviceCenter *DeviceFaultProcessCenter
	// maintain 5s ago device info
	DiagnosisAccompanyTimeout int64
	// nodeName -> deviceName -> faultQue
	uceAccompanyFaultQue map[string]map[string][]constant.DeviceFault
	// uceFaultTime
	uceFaultTime       map[string]map[string]int64
	deviceCmForNodeMap map[string]constant.AdvanceDeviceFaultCm
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

type configMap[T constant.ConfigMapInterface] struct {
	configmap map[string]T
}

type faultCenterCmManager[T constant.ConfigMapInterface] struct {
	mutex        sync.RWMutex
	cmBuffer     *collector.ConfigmapCollectBuffer[T]
	originalCm   configMap[T]
	processingCm configMap[T]
	processedCm  configMap[T]
}

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
