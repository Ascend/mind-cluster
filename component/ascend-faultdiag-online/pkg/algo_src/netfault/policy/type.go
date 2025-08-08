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
	Version       string
	SuperPodID    string
	NodeDeviceMap map[string]*NodeDevice `json:"NodeDeviceMap,omitempty"`
	RackMap       map[string]*RackInfo   `json:"RackMap,omitempty"`
}

// NodeDevice node device info
type NodeDevice struct {
	NodeName   string
	ServerID   string
	ServerType string              `json:"-"`
	DeviceMap  map[string]string   // key: dev phyID, value: superPod device id
	RackID     string              `json:"RackID,omitempty"`
	NpuInfoMap map[string]*NpuInfo `json:"NpuInfoMap,omitempty"`
}

// RackInfo rack info
type RackInfo struct {
	RackID    string
	ServerMap map[string]*ServerInfo
}

// ServerInfo server info
type ServerInfo struct {
	ServerIndex string
	NodeName    string
	NpuMap      map[string]*NpuInfo
}

// NpuInfo npu info for device
type NpuInfo struct {
	/* 新1D、2D */
	Ports     []PortInfo `json:"ports"`
	PhyId     string
	VnicIpMap map[string]*VnicInfo
}

// VnicInfo vnic ip info for device
type VnicInfo struct {
	PortId string
	VnicIp string
}

// PortInfo out of rack detection, eid for device
type PortInfo struct {
	Position  string   `json:"position"`
	AddrType  string   `json:"addrType"`
	Addresses []string `json:"addrs"`
}

// EidNpuMap npu与eid映射关系
type EidNpuMap struct {
	Map map[string]algo.NpuInfo
}

// EndPoint rack级topo关系npu端到端
type EndPoint struct {
	Type     string `json:"type"`
	Id       int    `json:"id"`
	Addr     string `json:"addr"`
	Position string `json:"position"`
}

// NpuPeer rack级npu卡id
type NpuPeer struct {
	Id int `json:"id"`
}

// PeerToPeer rack级topo中npu直连信息
type PeerToPeer struct {
	Level    int      `json:"level"`
	Protocol string   `json:"protocol"`
	SrcPoint EndPoint `json:"u_endpoint"`
	DstPoint EndPoint `json:"v_endpoint"`
}

// RackTopology rack级topo信息
type RackTopology struct {
	Version      string       `json:"version"`
	HardwareType string       `json:"hardware_type"`
	PeerCount    int          `json:"peer_count"`
	PeerList     []NpuPeer    `json:"peer_list"`
	EdgeCount    int          `json:"edge_count"`
	EdgeList     []PeerToPeer `json:"edge_list"`
}
