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
	FaultDevList      []*FaultDev
	HeartbeatTime     int64
	HeartbeatInterval int
	NodeStatus        string
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
