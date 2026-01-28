// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package constant a series of para
package constant

// FaultTimeAndLevel of each fault code
// some fault may not have accurate fault time and level,
// for example: duration fault use current time as `FaultTime`
type FaultTimeAndLevel struct {
	FaultTime         int64  `json:"fault_time"`
	FaultReceivedTime int64  `json:"-"`
	FaultLevel        string `json:"fault_level"`
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
	ForceAdd             bool                         `json:"-"`
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
	RackID      int32
	ServerIndex int32
}

// SwitchInfo record switch info
type SwitchInfo struct {
	SwitchFaultInfo
	CmName string
}

// SwitchFaultInfo switch info detail
type SwitchFaultInfo struct {
	FaultInfo            []SimpleSwitchFaultInfo
	FaultLevel           string
	UpdateTime           int64
	NodeStatus           string
	FaultTimeAndLevelMap map[string]FaultTimeAndLevel
}

type SwitchInfoFromCM struct {
	SwitchFaultInfoFromCm
	CmName string
}

// SwitchFaultInfoFromCm switch info detail from cm
type SwitchFaultInfoFromCm struct {
	FaultCode            []string
	FaultLevel           string
	UpdateTime           int64
	NodeStatus           string
	FaultTimeAndLevelMap map[string]FaultTimeAndLevel
}

// DeviceInfoCM record node NPU device information
type DeviceInfoCM struct {
	DeviceInfo  DeviceInfoNoName
	SuperPodID  int32
	RackID      int32
	ServerIndex int32
	CheckCode   string
}

// DeviceInfoNoName record node NPU device information. Will be solidified into cm
type DeviceInfoNoName struct {
	DeviceList map[string]string
	UpdateTime int64
}

// DpuCMDataList data structures of DPUList in dpu cm
type DpuCMDataList []DpuCMDataItem

// DpuInfoCM data structures of dpu cm in clusterd
type DpuInfoCM struct {
	BusType      string
	DPUList      DpuCMDataList
	NpuToDpusMap map[string][]string
	UpdateTime   int64
	CmName       string
}

// DpuCMDataItem data structures of DPUListItem in dpu cm
type DpuCMDataItem struct {
	Name      string
	Operstate string
	DeviceID  string
	VendorID  string
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
	ScheduleProcess     string `json:"-"`
	ScheduleFailReason  string `json:"-"`
	Status              string `json:"-"`
	Name                string `json:"-"`
	Namespace           string `json:"-"`
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
	// MultiInstanceJobId is the job id of multi-instance acjob, unique identification of a reasoning task
	MultiInstanceJobId string
	// AppType is the app type of acjob
	// for multi-instance job, the value is controller coordinator or server
	AppType   string
	NodeNames map[string]string
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
	ServerID     string   `json:"server_id"` // host ip
	HostIp       string   `json:"host_ip"`   // host ip
	SuperPodId   int      `json:"super_pod_id"`
	PodID        string   `json:"-"` // actual pod name
	PodNameSpace string   `json:"-"`
	ServerName   string   `json:"server_name"` // node name
	ServerSN     string   `json:"server_sn"`   // node sn
}

// Device to hccl with rankId
type Device struct {
	DeviceID      string `json:"device_id"`
	DeviceIP      string `json:"device_ip"`
	RankID        string `json:"rank_id"` // rank id
	SuperDeviceID string `json:"super_device_id,omitempty"`
}

// PodDevice pod annotation device info
type PodDevice struct {
	Devices    []Device `json:"devices"`
	PodName    string   `json:"pod_name"`
	ServerID   string   `json:"server_id"`
	HostIp     string   `json:"host_ip"`
	SuperPodId int      `json:"super_pod_id"`
}

// JobServerInfoMap to store job server info
type JobServerInfoMap struct {
	InfoMap       map[string]map[string]ServerHccl
	RetryTolerate map[string]bool
	ResourceType  map[string]string
}

// DeviceFaultDetail device fault detail
type DeviceFaultDetail struct {
	FaultCodeLevel map[string]string
	PodRankStr     string
	FaultTime      int64
	RecoverTime    int64
	CompleteTime   int64
	ReportTime     int64 // ReportTime: notify the data center to report the fault directly to cm
	FaultType      string
}

// RetryDeviceInfo retry device info
type RetryDeviceInfo struct {
	// DeviceName has prefix Ascend910
	DeviceName  string
	FaultDetail DeviceFaultDetail
}

// RetryNodeInfo retry node info
type RetryNodeInfo struct {
	NodeName string
	// DeviceName->DeviceInfo
	DeviceInfo map[string]RetryDeviceInfo
}

// RetryJobInfo retry job info
type RetryJobInfo struct {
	// RetryNode node->nodeInfo
	RetryNode map[string]RetryNodeInfo
	JobId     string
}

// SingleProcessDeviceInfo single process fault info
type SingleProcessDeviceInfo struct {
	// DeviceName has prefix Ascend910
	DeviceName  string
	FaultDetail *DeviceFaultDetail // key is retry or normal
}

// SingleProcessNodeInfo single process node info
type SingleProcessNodeInfo struct {
	NodeName string
	// DeviceName->DeviceInfo
	DeviceInfo map[string]*SingleProcessDeviceInfo
}

// SingleProcessJobInfo single process job info
type SingleProcessJobInfo struct {
	Node  map[string]*SingleProcessNodeInfo
	JobId string
}

// SingleProcessJobPodInfo single process job pod info
type SingleProcessJobPodInfo struct {
	Pod   map[string]*DeviceFaultDetail // key: pod rank
	JobId string
}

// ReportInfo train process report retry info
type ReportInfo struct {
	RecoverTime  int64
	CompleteTime int64
	FaultType    string
}

// FaultProcessor a interface of fault process
type FaultProcessor interface {
	Process(info any) any
}

// AdvanceDeviceFaultCm more structure device info
type AdvanceDeviceFaultCm struct {
	DeviceType          string
	CmName              string
	SuperPodID          int32
	RackID              int32
	ServerIndex         int32
	FaultDeviceList     map[string][]DeviceFault
	AvailableDeviceList []string
	Recovering          []string
	CardUnHealthy       []string
	NetworkUnhealthy    []string
	DPUUnhealthy        []string
	UpdateTime          int64
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
	DeviceCm map[string]*AdvanceDeviceFaultCm
	SwitchCm map[string]*SwitchInfo
	NodeCm   map[string]*NodeInfo
}

// ConfigMapInterface configmap interface
type ConfigMapInterface interface {
	GetCmName() string
	IsSame(another ConfigMapInterface) bool
	UpdateFaultReceiveTime(oldInfo ConfigMapInterface)
}

// FaultRank defines the structure for storing fault rank information.
// It includes the rank ID and fault code.
type FaultRank struct {
	RankId           string
	PodUid           string
	PodRank          string
	FaultCode        string
	FaultLevel       string
	DoStepRetry      bool
	DoRestartInPlace bool
	DeviceId         string // This value will only be filled in when fault type is npu
}

// JobFaultInfo job fault rank info
type JobFaultInfo struct {
	JobId        string
	FaultList    []FaultRank
	FaultDevice  []FaultDevice
	HealthyState string
}

// FaultDevice fault device  info
type FaultDevice struct {
	ServerName      string
	ServerSN        string
	ServerId        string
	HostIp          string
	DeviceId        string
	FaultCode       string
	FaultLevel      string
	DeviceType      string
	SwitchChipId    string
	SwitchPortId    string
	SwitchFaultTime string
}

// PodFaultInfo simple pod rank fault
type PodFaultInfo struct {
	PodRank          string
	DoRestartInPlace bool
	FaultType        int
	ShouldExitPod    bool
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
	ForceAdd         bool
	SwitchChipId     uint
	SwitchPortId     uint
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
	ForceAdd           bool `json:"-"`
}

// ReportRecoverInfo cluster grpc should call back for report uce fault
type ReportRecoverInfo struct {
	JobId       string
	Rank        string
	RecoverTime int64
	FaultType   string
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

// HccspingMeshItem is the configuration for the pingmesh component
type HccspingMeshItem struct {
	Activate     string `json:"activate"`
	TaskInterval int    `json:"task_interval"`
}

// ConfigPingMesh the config of pingmesh set by user
type ConfigPingMesh map[string]*HccspingMeshItem

// CathelperConf config info for cathelper
type CathelperConf struct {
	SuppressedPeriod int    `json:"suppressedPeriod"`
	NetworkType      int    `json:"networkType"`
	PingType         int    `json:"pingType"`
	PingTimes        int    `json:"pingTimes"`
	PingInterval     int    `json:"pingInterval"`
	Period           int    `json:"period"`
	NetFault         string `json:"netFault"`
}

// CacheStatus cache the status
type CacheStatus struct {
	Inited bool
}

// NetFaultInfo the ras feature cm of fault network
type NetFaultInfo struct {
	NetFault int // the switch of fault network feature
}

// SimplePodInfo of Pod
type SimplePodInfo struct {
	PodUid   string
	PodRank  string
	NodeName string
}
