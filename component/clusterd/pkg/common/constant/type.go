// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package constant a series of para
package constant

import (
	"ascend-common/common-utils/hwlog"
)

// FaultTimeAndLevel of each fault code
// some fault may not have accurate fault time and level,
// for example: duration fault use current time as `FaultTime`
type FaultTimeAndLevel struct {
	FaultTime  int64  `json:"fault_time"`
	FaultLevel string `json:"fault_level"`
}

// DeviceFault device or network fault info
type DeviceFault struct {
	FaultType            string                       `json:"fault_type"`
	NPUName              string                       `json:"npu_name"`
	LargeModelFaultLevel string                       `json:"large_model_fault_level"`
	FaultLevel           string                       `json:"fault_level"`
	FaultHandling        string                       `json:"fault_handling"`
	FaultCode            string                       `json:"fault_code"`
	FaultTimeAndLevelMap map[string]FaultTimeAndLevel `json:"fault_time_and_level_map"`
}

// NodeInfoCM the config map struct of node info
type NodeInfoCM struct {
	NodeInfo  NodeInfoNoName
	CheckCode string
}

// NodeInfoNoName node info without cm name
type NodeInfoNoName struct {
	FaultDevList []*FaultDev
	NodeStatus   string
}

// NodeInfo node info
type NodeInfo struct {
	NodeInfoNoName
	CmName string
}

// FaultDev fault device struct
type FaultDev struct {
	DeviceType string
	DeviceId   int64
	FaultCode  []string
	FaultLevel string
}

// DeviceInfo record node NPU device information. Will be solidified into cm
type DeviceInfo struct {
	DeviceInfoNoName
	CmName      string
	SuperPodID  int32
	ServerIndex int32
}

// SwitchInfo record switch info
type SwitchInfo struct {
	SwitchFaultInfo
	CmName string
}

// SwitchFaultInfo switch info detail
type SwitchFaultInfo struct {
	FaultCode  []string
	FaultLevel string
	UpdateTime int64
	NodeStatus string
}

// DeviceInfoCM record node NPU device information
type DeviceInfoCM struct {
	DeviceInfo  DeviceInfoNoName
	SuperPodID  int32
	ServerIndex int32
	CheckCode   string
}

// DeviceInfoNoName record node NPU device information. Will be solidified into cm
type DeviceInfoNoName struct {
	DeviceList map[string]string
	UpdateTime int64
}

// CurrJobStatistic current job statistic information
type CurrJobStatistic struct {
	JobStatistic map[string]JobStatistic
}

// JobNotifyMsg notify msg
type JobNotifyMsg struct {
	Operator string
	JobKey   string
}

// JobStatistic job statistic information
type JobStatistic struct {
	K8sJobID            string `json:"id"`                 // k8s job id
	CustomJobID         string `json:"customID,omitempty"` // custom job id
	CardNums            int64  `json:"cardNum,omitempty"`
	PodFirstRunningTime int64  `json:"podFirstRunTime,omitempty"`
	StopTime            int64  `json:"stopTime,omitempty"` // stop time when job failed or complete
	PodLastRunningTime  int64  `json:"podLastRunTime,omitempty"`
	PodLastFaultTime    int64  `json:"podLastFaultTime,omitempty"`
	PodFaultTimes       int64  `json:"podFaultTimes,omitempty"`
	Status              string `json:"-"`
	Name                string `json:"-"`
	NameSpace           string `json:"-"`
}

// JobInfo : normal job info
type JobInfo struct {
	JobType           string
	Framework         string
	NameSpace         string
	Name              string
	Key               string
	Replicas          int
	Status            string
	IsPreDelete       bool
	JobRankTable      RankTable // when job is preDelete or status is pending, jobRankTable is nil
	AddTime           int64
	DeleteTime        int64
	TotalCmNum        int
	LastUpdatedCmTime int64
	PreServerList     []ServerHccl
	SharedTorIp       string
	MasterAddr        string
	ResourceType      string
	CustomJobID       string
}

// RankTable rank table info
type RankTable struct {
	Status      string       `json:"status"`
	ServerList  []ServerHccl `json:"server_list"`
	ServerCount string       `json:"server_count"`
	Total       int          `json:"total"`
}

// ServerHccl to hccl
type ServerHccl struct {
	DeviceList   []Device `json:"device"`
	ServerID     string   `json:"server_id"`
	PodID        string   `json:"-"`
	PodNameSpace string   `json:"-"`
	ServerName   string   `json:"server_name"`
}

// Device to hccl with rankId
type Device struct {
	DeviceID string `json:"device_id"`
	DeviceIP string `json:"device_ip"`
	RankID   string `json:"rank_id"` // rank id
}

// PodDevice pod annotation device info
type PodDevice struct {
	Devices  []Device `json:"devices"`
	PodName  string   `json:"pod_name"`
	ServerID string   `json:"server_id"`
}

// JobServerInfoMap to store job server info
type JobServerInfoMap struct {
	InfoMap      map[string]map[string]ServerHccl
	UceTolerate  map[string]bool
	ResourceType map[string]string
}

// UceDeviceInfo uce device info
type UceDeviceInfo struct {
	// DeviceName has prefix Ascend910
	DeviceName   string
	FaultTime    int64
	RecoverTime  int64
	CompleteTime int64
}

// UceNodeInfo uce node info
type UceNodeInfo struct {
	NodeName string
	// DeviceName->DeviceInfo
	DeviceInfo map[string]UceDeviceInfo
}

// UceJobInfo uce job info
type UceJobInfo struct {
	// UceNode node->nodeInfo
	UceNode map[string]UceNodeInfo
	JobId   string
}

// ReportInfo train process report uce info
type ReportInfo struct {
	RecoverTime  int64
	CompleteTime int64
}

// FaultProcessor a interface of fault process
type FaultProcessor interface {
	Process(info any) any
}

// AdvanceDeviceFaultCm more structure device info
type AdvanceDeviceFaultCm struct {
	ServerType       string
	CmName           string
	SuperPodID       int32
	ServerIndex      int32
	FaultDeviceList  map[string][]DeviceFault
	CardUnHealthy    []string
	NetworkUnhealthy []string
	UpdateTime       int64
}

// InformerCmItem informer configmap item of queue or buffer
type InformerCmItem[T ConfigMapInterface] struct {
	IsAdd bool
	Data  T
}

// OneConfigmapContent contains one kind of configmap content
type OneConfigmapContent[T ConfigMapInterface] struct {
	AllConfigmap    map[string]T
	UpdateConfigmap []InformerCmItem[T]
}

// AllConfigmapContent contains all kind of configmap content
type AllConfigmapContent struct {
	DeviceCm map[string]*DeviceInfo
	SwitchCm map[string]*SwitchInfo
	NodeCm   map[string]*NodeInfo
}

// ConfigMapInterface configmap interface
type ConfigMapInterface interface {
	GetCmName() string
	IsSame(another ConfigMapInterface) bool
}

// GetCmName get configmap name of device info
func (cm *DeviceInfo) GetCmName() string {
	return cm.CmName
}

// GetCmName get configmap name of switch info
func (cm *SwitchInfo) GetCmName() string {
	return cm.CmName
}

// GetCmName get configmap name of node info
func (cm *NodeInfo) GetCmName() string {
	return cm.CmName
}

// IsSame compare with another cm
func (cm *DeviceInfo) IsSame(another ConfigMapInterface) bool {
	anotherDeviceInfo, ok := another.(*DeviceInfo)
	if !ok {
		hwlog.RunLog.Warnf("compare with cm which is not DeviceInfo")
		return false
	}
	return !DeviceInfoBusinessDataIsNotEqual(cm, anotherDeviceInfo)
}

// IsSame compare with another cm
func (cm *SwitchInfo) IsSame(another ConfigMapInterface) bool {
	anotherSwitchInfo, ok := another.(*SwitchInfo)
	if !ok {
		hwlog.RunLog.Warnf("compare with cm which is not SwitchInfo")
		return false
	}
	return !SwitchInfoBusinessDataIsNotEqual(cm, anotherSwitchInfo)
}

// IsSame compare with another cm
func (cm *NodeInfo) IsSame(another ConfigMapInterface) bool {
	anotherNodeInfo, ok := another.(*NodeInfo)
	if !ok {
		hwlog.RunLog.Warnf("compare with cm which is not NodeInfo")
		return false
	}
	return !NodeInfoBusinessDataIsNotEqual(cm, anotherNodeInfo)
}

// DeviceInfoBusinessDataIsNotEqual determine the business data is not equal
func DeviceInfoBusinessDataIsNotEqual(oldDevInfo *DeviceInfo, devInfo *DeviceInfo) bool {
	if oldDevInfo == nil && devInfo == nil {
		hwlog.RunLog.Debug("both oldDevInfo and devInfo are nil")
		return false
	}
	if oldDevInfo == nil || devInfo == nil {
		hwlog.RunLog.Debug("one of oldDevInfo and devInfo is not empty, and the other is empty")
		return true
	}
	if len(oldDevInfo.DeviceList) != len(devInfo.DeviceList) {
		hwlog.RunLog.Debug("the length of the deviceList of oldDevInfo is not equal to that of the deviceList of devInfo")
		return true
	}
	for nKey, nValue := range oldDevInfo.DeviceList {
		oValue, exists := devInfo.DeviceList[nKey]
		if !exists || nValue != oValue {
			hwlog.RunLog.Debug("neither oldDevInfo nor devInfo is empty, but oldDevInfo is not equal to devInfo")
			return true
		}
	}
	hwlog.RunLog.Debug("oldDevInfo is equal to devInfo")
	return false
}

// SwitchInfoBusinessDataIsNotEqual judge is the faultcode and fault level is the same as known, if is not same returns true
func SwitchInfoBusinessDataIsNotEqual(oldSwitch, newSwitch *SwitchInfo) bool {
	if oldSwitch == nil && newSwitch == nil {
		return false
	}
	if (oldSwitch != nil && newSwitch == nil) || (oldSwitch == nil && newSwitch != nil) {
		return true
	}
	if newSwitch.FaultLevel != oldSwitch.FaultLevel || newSwitch.NodeStatus != oldSwitch.NodeStatus ||
		len(newSwitch.FaultCode) != len(oldSwitch.FaultCode) {
		return true
	}
	hwlog.RunLog.Debug("oldSwitch is equal to newSwitch")
	return false
}

// NodeInfoBusinessDataIsNotEqual determine the business data is not equal
func NodeInfoBusinessDataIsNotEqual(oldNodeInfo *NodeInfo, newNodeInfo *NodeInfo) bool {
	if oldNodeInfo == nil && newNodeInfo == nil {
		hwlog.RunLog.Debug("both oldNodeInfo and newNodeInfo are nil")
		return false
	}
	if oldNodeInfo == nil || newNodeInfo == nil {
		hwlog.RunLog.Debug("one of oldNodeInfo and newNodeInfo is not empty, and the other is empty")
		return true
	}
	if oldNodeInfo.NodeStatus != newNodeInfo.NodeStatus ||
		len(oldNodeInfo.FaultDevList) != len(newNodeInfo.FaultDevList) {
		hwlog.RunLog.Debug("neither oldNodeInfo nor newNodeInfo is empty, but oldNodeInfo is not equal to newNodeInfo")
		return true
	}
	hwlog.RunLog.Debug("oldNodeInfo is equal to newNodeInfo")
	return false
}

// FaultRank defines the structure for storing fault rank information.
// It includes the rank ID and fault code.
type FaultRank struct {
	RankId      string
	PodUid      string
	PodRank     string
	FaultCode   string
	FaultLevel  string
	DoStepRetry bool
}

// JobFaultInfo job fault rank info
type JobFaultInfo struct {
	JobId        string
	FaultList    []FaultRank
	HealthyState string
}

// FaultStrategy fault strategies
type FaultStrategy struct {
	NodeLvList   map[string]string
	DeviceLvList map[string][]DeviceStrategy
}

// DeviceStrategy device fault strategy
type DeviceStrategy struct {
	Strategy string
	NPUName  string
}

// FaultInfo fault info of relation fault process
type FaultInfo struct {
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

// FaultDuration fault duration config
type FaultDuration struct {
	FaultCode       string
	FaultType       string
	TimeOutInterval int64
}

// RelationFaultStrategy relation fault strategy
type RelationFaultStrategy struct {
	TriggerFault   string
	RelationFaults []string
	FaultStrategy  string
}

// SimpleSwitchFaultInfo simple switch fault info
type SimpleSwitchFaultInfo struct {
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

// ReportRecoverInfo cluster grpc should call back for report uce fault
type ReportRecoverInfo struct {
	JobId       string
	Rank        string
	RecoverTime int64
}

// PubFaultCache public fault in cache for node
type PubFaultCache struct {
	FaultDevIds   []int32
	FaultDevNames []string
	FaultId       string
	FaultType     string
	FaultCode     string
	FaultLevel    string
	FaultTime     int64
	Assertion     string
	FaultAddTime  int64
}

// NodeFault node fault info
type NodeFault struct {
	FaultResource string  `json:"resource,omitempty"`
	FaultDevIds   []int32 `json:"devIds,omitempty"`
	FaultId       string  `json:"faultId,omitempty"`
	FaultType     string  `json:"type,omitempty"`
	FaultCode     string  `json:"faultCode,omitempty"`
	FaultLevel    string  `json:"level,omitempty"`
	FaultTime     int64   `json:"faultTime,omitempty"`
}

// FaultNum faults number
type FaultNum struct {
	TotalFaultNum      int `json:"-"`
	DevFaultNum        int `json:"-"`
	DevNetworkFaultNum int `json:"-"`
	NodeFaultNum       int `json:"-"`
	PubFaultNum        int `json:"publicFaultNum"`
}
