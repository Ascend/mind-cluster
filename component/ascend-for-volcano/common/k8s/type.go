/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package k8s is using for the k8s operation.
*/
package k8s

import "sync"

// ClusterInfoWitchCm cluster info whit different configmap
type ClusterInfoWitchCm struct {
	deviceInfos       *DeviceInfosWithMutex
	nodeInfosFromCm   *NodeInfosFromCmWithMutex   // NodeInfos is get from kube-system/node-info- configmap
	switchInfosFromCm *SwitchInfosFromCmWithMutex // switchInfosFromCm is get from mindx-dl/device-info- configmap
	dpuInfosFromCm    *DpuInfosWithMutex          // dpuInfosFromCm is get from dpuinfo- configmap
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

// DpuInfosWithMutex dpu info with mutex
type DpuInfosWithMutex struct {
	sync.Mutex
	Dpus map[string]DpuInfoWithNode
}

// DpuInfoWithNode is the dpu information reported by cm with cache update time
type DpuInfoWithNode struct {
	DpuInfoCfg
	CacheUpdateTime int64
}

// DpuInfoCfg represents the DPU information configuration structure, aligned with dpu-dp cm
type DpuInfoCfg struct {
	DPUInfo    DPUInfoBody `json:"DPUInfo"`
	UpdateTime int64       `json:"UpdateTime"`
}

// DPUInfoBody is the body of DPU info, containing DPU list and node event
type DPUInfoBody struct {
	DPUList   []DPUItem     `json:"DPUList"`
	NodeEvent *DpuNodeEvent `json:"NodeEvent"`
}

// DPUItem represents a single DPU device information
type DPUItem struct {
	HcaName     string           `json:"HcaName"`
	EthName     string           `json:"EthName"`
	IpAddr      string           `json:"IpAddr,omitempty"`
	DeviceID    string           `json:"DeviceID"`
	VendorID    string           `json:"VendorID"`
	FaultList   []DpuFaultDetail `json:"FaultList"`
	AffectedNPU []int            `json:"AffectedNPU"`
}

// DpuNodeEvent represents node-level dpu fault events (e.g. dpu card drop)
type DpuNodeEvent struct {
	NodeName  string           `json:"NodeName"`
	FaultList []DpuFaultDetail `json:"FaultList"`
}

// DpuFaultDetail represents detailed information about a dpu fault
type DpuFaultDetail struct {
	FaultCode   string `json:"FaultCode"`
	Time        int64  `json:"Time"`
	Description string `json:"Description"`
	FaultLevel  string `json:"FaultLevel"`
}

// NodeDeviceInfo like node annotation.
type NodeDeviceInfo struct {
	DeviceList map[string]string
	UpdateTime int64
}

// NodeDeviceInfoWithID is node the information reported by cm.
type NodeDeviceInfoWithID struct {
	NodeDeviceInfo
	CacheUpdateTime int64
	SuperPodID      int32
	RackID          int32
}

// NodeDNodeInfo is node the information reported by noded
type NodeDNodeInfo struct {
	FaultDevList []FaultDevList
	NodeStatus   string
}

// FaultDevList is node fault device list information
type FaultDevList struct {
	DeviceType string
	DeviceId   int
	FaultCode  []string
	FaultLevel string
}

// SwitchFaultInfo Switch Fault Info
type SwitchFaultInfo struct {
	FaultCode  []string
	FaultLevel string
	UpdateTime int64
	NodeStatus string
}

// NodeDeviceInfoWithDevPlugin a node has one by cm.
type NodeDeviceInfoWithDevPlugin struct {
	DeviceInfo  NodeDeviceInfo
	CheckCode   string
	SuperPodID  int32 `json:"SuperPodID,omitempty"`
	ServerIndex int32 `json:"ServerIndex,omitempty"`
	RackID      int32 `json:"RackID,omitempty"`
}

// NodeInfoWithNodeD is node the node information and checkCode reported by noded
type NodeInfoWithNodeD struct {
	NodeInfo  NodeDNodeInfo
	CheckCode string
}

// NewClusterInfoWitchCm new empty cluster info with cm
func NewClusterInfoWitchCm() ClusterInfoWitchCm {
	return ClusterInfoWitchCm{
		deviceInfos: &DeviceInfosWithMutex{
			Mutex:   sync.Mutex{},
			Devices: map[string]NodeDeviceInfoWithID{},
		},
		nodeInfosFromCm: &NodeInfosFromCmWithMutex{
			Mutex: sync.Mutex{},
			Nodes: map[string]NodeDNodeInfo{},
		},
		switchInfosFromCm: &SwitchInfosFromCmWithMutex{
			Mutex:    sync.Mutex{},
			Switches: map[string]SwitchFaultInfo{},
		},
		dpuInfosFromCm: &DpuInfosWithMutex{
			Mutex: sync.Mutex{},
			Dpus:  map[string]DpuInfoWithNode{},
		},
	}
}
