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

// Package policy for processing superpod information
package policy

import "ascend-faultdiag-online/pkg/algo_src/netfault/algo"

// SuperPodInfo super node device info, key is superPodID, value is RackInfo
type SuperPodInfo struct {
	// Version represents the version of the super pod (A3 or A5)
	Version string
	// SuperPodID is the unique identifier of the super pod
	SuperPodID string
	// NodeDeviceMap is a mapping of node devices
	NodeDeviceMap map[string]*NodeDevice `json:"NodeDeviceMap,omitempty"`
	// RackMap is a mapping of rack information
	RackMap map[string]*RackInfo `json:"RackMap,omitempty"`
}

// NodeDevice node device info
type NodeDevice struct {
	// NodeName is the name of the node
	NodeName string
	// ServerID is the identifier of the server
	ServerID string
	// ServerType indicates the type of the server
	ServerType string `json:"-"`
	// DeviceMap is a mapping of device information
	DeviceMap map[string]string // key: dev phyID, value: superPod device id
	// RackID is the identifier of the rack
	RackID string `json:"RackID,omitempty"`
	// NpuInfoMap is a mapping of NPU information
	NpuInfoMap map[string]*NpuInfo `json:"NpuInfoMap,omitempty"`
}

// RackInfo rack info
type RackInfo struct {
	// RackID is the identifier of the rack
	RackID string
	// ServerMap is a mapping of server information
	ServerMap map[string]*ServerInfo
}

// ServerInfo server info
type ServerInfo struct {
	// ServerIndex is the index identifier of the server
	ServerIndex string
	// NodeName is the name of the node
	NodeName string
	// NpuMap is a mapping of NPU information
	NpuMap map[string]*NpuInfo
}

// NpuInfo npu info for device
type NpuInfo struct { /* 新1D、2D */
	// Ports is a slice of port information
	Ports []PortInfo `json:"ports"`
	// PhyId is the physical identifier
	PhyId string
	// VnicIpMap is a mapping of virtual NIC IP information
	VnicIpMap map[string]*VnicInfo
	// LevelList is a leve list
	LevelList []LevelElement
}

// LevelElement for the concrete level info in rank table
type LevelElement struct {
	NetLayer      int    `json:"net_layer"`       // generate by operator, tentatively increase from 0
	NetInstanceID string `json:"net_instance_id"` // from annotation, relying on super_pod_id field
	NetType       string `json:"net_type"`        // generate by operator, level=0 tentatively empty; level=1,2 clos
	NetAttr       string `json:"net_attr"`        // generate by operator, tentatively empty /
	// generate by operator, tentatively level=0,3 nil; level=1,2 from 9th and 18th eid
	RankAddrList []RankAddrItem `json:"rank_addr_list"`
}

// RankAddrItem for item info in LevelElement
type RankAddrItem struct {
	// AddrType address type
	AddrType string `json:"addr_type"`
	// Addr rank address
	Addr string `json:"addr"`
	// Ports rank ports
	Ports []string `json:"ports"`
	// PlanId plan id
	PlaneId string `json:"plane_id"`
}

// VnicInfo vnic ip info for device
type VnicInfo struct {
	// PortId is the identifier of the port
	PortId string
	// VnicIp is the IP address of the virtual NIC
	VnicIp string
}

// PortInfo out of rack detection, eid for device
type PortInfo struct {
	// Position represents the position information of the port
	Position string `json:"position"`
	// AddrType indicates the type of address associated with the port
	AddrType string `json:"addrType"`
	// Addresses is a slice of address information for the port
	Addresses []string `json:"addrs"`
}

// EidNpuMap mapping between NPU and EID
type EidNpuMap struct {
	// Map is a mapping where the key is an EID and the value is the corresponding NPU information
	Map map[string]algo.NpuInfo
}

// EndPoint NPU end-to-end in rack-level topology relationship
type EndPoint struct {
	// Type indicates the type of the endpoint
	Type string `json:"type"`
	// Id is the identifier of the endpoint
	Id int `json:"id"`
	// Addr is the address of the endpoint
	Addr string `json:"addr"`
	// Position represents the position information of the endpoint
	Position string `json:"position"`
}

// NpuPeer rack-level NPU card ID
type NpuPeer struct {
	// Id is the rack-level NPU card identifier
	Id int `json:"id"`
}

// PeerToPeer NPU direct connection information in rack-level topology
type PeerToPeer struct {
	// Level indicates the hierarchy information
	Level int `json:"level"`
	// Protocol represents the protocol information
	Protocol string `json:"protocol"`
	// SrcPoint is the source endpoint
	SrcPoint EndPoint `json:"u_endpoint"`
	// DstPoint is the destination endpoint
	DstPoint EndPoint `json:"v_endpoint"`
}

// RackTopology rack-level topology information
type RackTopology struct {
	// Version represents the version information
	Version string `json:"version"`
	// HardwareType indicates the type of hardware
	HardwareType string `json:"hardware_type"`
	// PeerCount is the number of peers
	PeerCount int `json:"peer_count"`
	// PeerList is a slice of peer
	PeerList []NpuPeer `json:"peer_list"`
	// EdgeCount is the number of edges
	EdgeCount int `json:"edge_count"`
	// EdgeList is a slice of PeerToPeer instances
	EdgeList []PeerToPeer `json:"edge_list"`
}
