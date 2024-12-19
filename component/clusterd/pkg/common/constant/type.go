// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package constant a series of para
package constant

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
	DeviceList []Device `json:"device"`
	ServerID   string   `json:"server_id"`
	PodID      string   `json:"-"`
	ServerName string   `json:"server_name"`
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
	InfoMap     map[string]map[string]ServerHccl
	UceTolerate map[string]bool
}
