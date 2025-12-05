// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package api structs for SuperPodDevice
package api

import "k8s.io/apimachinery/pkg/util/sets"

// NpuBaseInfo is the base info of npu
type NpuBaseInfo struct {
	IP            string
	SuperDeviceID uint32
	// LevelList info for A5
	LevelList []RankLevel `json:"levelList,omitempty"`
}

// NodeDevice node device info
type NodeDevice struct {
	NodeName   string
	ServerID   string
	ServerType string            `json:"-"`
	DeviceMap  map[string]string // key: dev phyID, value: superPod device id
	// RackID for A5 ras netfault
	RackID string `json:"RackID,omitempty"`
	// NpuInfoMap for A5 ras netfault
	NpuInfoMap map[string]*NpuInfo `json:"NpuInfoMap,omitempty"`
	// AcceleratorType for A5 ras netfault
	AcceleratorType string `json:"AcceleratorType,omitempty"`
}

// SuperPodDevice super node device info, key is superPodID, value is NodeDevice
type SuperPodDevice struct {
	Version       string
	SuperPodID    string
	NodeDeviceMap map[string]*NodeDevice
	// RackMap  for A5 ras
	RackMap map[string]*RackInfo `json:"RackMap,omitempty"`
	// AcceleratorType for A5 ras
	AcceleratorType string `json:"AcceleratorType,omitempty"`
}

// SuperPodFaultInfos super pod fault info
type SuperPodFaultInfos struct {
	SdIds      []string
	FaultNodes sets.String
	NodeNames  []string
	FaultTimes int64
	JobId      string `json:"JobId,omitempty"`
}
