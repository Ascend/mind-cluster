// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package l2fault test for l2 fault cache util
package l2fault

import (
	"reflect"
	"sync"
	"testing"

	"clusterd/pkg/common/constant"
)

const (
	node1 = "node1"
	job1  = "job1"
)

func getMockAdvanceDeviceFaultCmForTest() *constant.AdvanceDeviceFaultCm {
	return &constant.AdvanceDeviceFaultCm{
		DeviceType:          "",
		CmName:              "CmName-" + node1,
		SuperPodID:          0,
		ServerIndex:         0,
		FaultDeviceList:     make(map[string][]constant.DeviceFault),
		AvailableDeviceList: []string{"xxx"},
		Recovering:          []string{"xxx"},
		CardUnHealthy:       []string{"xxx"},
		NetworkUnhealthy:    []string{"xxx"},
		UpdateTime:          0,
	}
}

func getMockSwitchInfoForTest() *constant.SwitchInfo {
	return &constant.SwitchInfo{
		SwitchFaultInfo: constant.SwitchFaultInfo{},
		CmName:          "CmName-" + node1,
	}
}

// TestGetDeletedDevL2FaultCmForNodeMap tests GetDeletedDevL2FaultCmForNodeMap method
func TestGetDeletedDevL2FaultCmForNodeMap(t *testing.T) {
	cache := &l2FaultCache{
		deletedDevL2FaultCm:      make(map[string]*constant.AdvanceDeviceFaultCm),
		deletedJobFaultDeviceMap: make(map[string][]constant.FaultDevice),
		rwMutex:                  sync.RWMutex{},
	}
	t.Run("get empty map", func(t *testing.T) {
		result := cache.GetDeletedDevL2FaultCmForNodeMap()
		if result == nil {
			t.Errorf("GetDeletedDevL2FaultCmForNodeMap() = nil, want empty map")
		}
		if len(result) != 0 {
			t.Errorf("GetDeletedDevL2FaultCmForNodeMap() length = %v, want 0", len(result))
		}
	})
	t.Run("get map with data", func(t *testing.T) {

		testData := map[string]*constant.AdvanceDeviceFaultCm{
			node1: getMockAdvanceDeviceFaultCmForTest(),
		}
		cache.SetDeletedDevL2FaultCmForNodeMap(testData)
		result := cache.GetDeletedDevL2FaultCmForNodeMap()
		if !reflect.DeepEqual(result, testData) {
			t.Errorf("GetDeletedDevL2FaultCmForNodeMap() = %v, want %v", result, testData)
		}
	})
}

// TestGetDeletedSwitchL2FaultCmForNodeMap tests GetDeletedSwitchL2FaultCmForNodeMap method
func TestGetDeletedSwitchL2FaultCmForNodeMap(t *testing.T) {
	cache := &l2FaultCache{
		deletedSwitchL2FaultCm:   make(map[string]*constant.SwitchInfo),
		deletedJobFaultDeviceMap: make(map[string][]constant.FaultDevice),
		rwMutex:                  sync.RWMutex{},
	}
	t.Run("get empty map", func(t *testing.T) {
		result := cache.GetDeletedSwitchL2FaultCmForNodeMap()
		if result == nil {
			t.Errorf("GetDeletedSwitchL2FaultCmForNodeMap() = nil, want empty map")
		}
		if len(result) != 0 {
			t.Errorf("GetDeletedSwitchL2FaultCmForNodeMap() length = %v, want 0", len(result))
		}
	})
	t.Run("get map with data", func(t *testing.T) {
		testData := map[string]*constant.SwitchInfo{
			node1: getMockSwitchInfoForTest(),
		}
		cache.SetDeletedSwitchL2FaultCmForNodeMap(testData)
		result := cache.GetDeletedSwitchL2FaultCmForNodeMap()
		if !reflect.DeepEqual(result, testData) {
			t.Errorf("GetDeletedSwitchL2FaultCmForNodeMap() = %v, want %v", result, testData)
		}
	})
}

func getMockFaultDeviceListForTest() []constant.FaultDevice {
	return []constant.FaultDevice{
		{ServerName: "node1", ServerId: "1", DeviceId: "0", FaultLevel: constant.SeparateNPU,
			DeviceType: constant.FaultTypeNPU},
		{ServerName: "node1", ServerId: "1", DeviceId: "1", FaultLevel: constant.RestartNPU,
			DeviceType: constant.FaultTypeNPU},
		{ServerName: "node1", ServerId: "1", DeviceId: "2", FaultLevel: constant.SubHealthFault,
			DeviceType: constant.FaultTypeNPU},
		{ServerName: "node1", ServerId: "1", DeviceId: "3", FaultLevel: constant.NotHandleFault,
			DeviceType: constant.FaultTypeNPU},
		{ServerName: "node0", ServerId: "0", DeviceId: "0", FaultLevel: constant.SeparateNPU,
			DeviceType: constant.FaultTypeSwitch},
		{ServerName: "node2", ServerId: "2", DeviceId: "0", FaultLevel: constant.SubHealthFault,
			DeviceType: constant.FaultTypeNPU},
		{ServerName: "node3", ServerId: "3", DeviceId: "0", FaultLevel: constant.NotHandleFault,
			DeviceType: constant.FaultTypeNPU},
	}
}

// TestGetDeletedJobFaultDeviceMap tests GetDeletedJobFaultDeviceMap method
func TestGetDeletedJobFaultDeviceMap(t *testing.T) {
	cache := &l2FaultCache{
		deletedDevL2FaultCm:      make(map[string]*constant.AdvanceDeviceFaultCm),
		deletedJobFaultDeviceMap: make(map[string][]constant.FaultDevice),
		rwMutex:                  sync.RWMutex{},
	}
	t.Run("get empty map", func(t *testing.T) {
		result := cache.GetDeletedJobFaultDeviceMap()
		if result == nil {
			t.Errorf("GetDeletedJobFaultDeviceMap() = nil, want empty map")
		}
		if len(result) != 0 {
			t.Errorf("GetDeletedJobFaultDeviceMap() length = %v, want 0", len(result))
		}
	})
	t.Run("get map with data", func(t *testing.T) {
		testData := map[string][]constant.FaultDevice{
			job1: getMockFaultDeviceListForTest(),
		}
		cache.SetDeletedJobFaultDeviceMap(testData)
		result := cache.GetDeletedJobFaultDeviceMap()
		if !reflect.DeepEqual(result, testData) {
			t.Errorf("GetDeletedJobFaultDeviceMap() = %v, want %v", result, testData)
		}
	})
}

// TestSetDeletedDevL2FaultCmForNodeMap tests SetDeletedDevL2FaultCmForNodeMap method
func TestSetDeletedDevL2FaultCmForNodeMap(t *testing.T) {
	cache := &l2FaultCache{
		deletedDevL2FaultCm:      make(map[string]*constant.AdvanceDeviceFaultCm),
		deletedJobFaultDeviceMap: make(map[string][]constant.FaultDevice),
		rwMutex:                  sync.RWMutex{},
	}
	t.Run("set normal map", func(t *testing.T) {
		testData := map[string]*constant.AdvanceDeviceFaultCm{
			node1: getMockAdvanceDeviceFaultCmForTest(),
		}
		cache.SetDeletedDevL2FaultCmForNodeMap(testData)
		result := cache.GetDeletedDevL2FaultCmForNodeMap()
		if !reflect.DeepEqual(result, testData) {
			t.Errorf("SetDeletedDevL2FaultCmForNodeMap() failed, got %v, want %v", result, testData)
		}
	})
	t.Run("set nil map", func(t *testing.T) {
		cache.SetDeletedDevL2FaultCmForNodeMap(nil)
		result := cache.GetDeletedDevL2FaultCmForNodeMap()
		if result != nil {
			t.Errorf("SetDeletedDevL2FaultCmForNodeMap(nil) failed, got %v, want nil", result)
		}
	})
}

// TestSetDeletedSwitchL2FaultCmForNodeMap tests SetDeletedSwitchL2FaultCmForNodeMap method
func TestSetDeletedSwitchL2FaultCmForNodeMap(t *testing.T) {
	cache := &l2FaultCache{
		deletedSwitchL2FaultCm:   make(map[string]*constant.SwitchInfo),
		deletedJobFaultDeviceMap: make(map[string][]constant.FaultDevice),
		rwMutex:                  sync.RWMutex{},
	}
	t.Run("set normal map", func(t *testing.T) {
		testData := map[string]*constant.SwitchInfo{
			node1: getMockSwitchInfoForTest(),
		}
		cache.SetDeletedSwitchL2FaultCmForNodeMap(testData)
		result := cache.GetDeletedSwitchL2FaultCmForNodeMap()
		if !reflect.DeepEqual(result, testData) {
			t.Errorf("SetDeletedSwitchL2FaultCmForNodeMap() failed, got %v, want %v", result, testData)
		}
	})
	t.Run("set nil map", func(t *testing.T) {
		cache.SetDeletedSwitchL2FaultCmForNodeMap(nil)
		result := cache.GetDeletedSwitchL2FaultCmForNodeMap()
		if result != nil {
			t.Errorf("SetDeletedSwitchL2FaultCmForNodeMap(nil) failed, got %v, want nil", result)
		}
	})
}

// TestSetDeletedJobFaultDeviceMap tests SetDeletedJobFaultDeviceMap method
func TestSetDeletedJobFaultDeviceMap(t *testing.T) {
	cache := &l2FaultCache{
		deletedDevL2FaultCm:      make(map[string]*constant.AdvanceDeviceFaultCm),
		deletedJobFaultDeviceMap: make(map[string][]constant.FaultDevice),
		rwMutex:                  sync.RWMutex{},
	}
	t.Run("set normal map", func(t *testing.T) {
		testData := map[string][]constant.FaultDevice{
			job1: getMockFaultDeviceListForTest(),
		}
		cache.SetDeletedJobFaultDeviceMap(testData)
		result := cache.GetDeletedJobFaultDeviceMap()
		if !reflect.DeepEqual(result, testData) {
			t.Errorf("SetDeletedJobFaultDeviceMap() failed, got %v, want %v", result, testData)
		}
	})
	t.Run("set nil map", func(t *testing.T) {
		cache.SetDeletedJobFaultDeviceMap(nil)
		result := cache.GetDeletedJobFaultDeviceMap()
		if result != nil {
			t.Errorf("SetDeletedJobFaultDeviceMap(nil) failed, got %v, want nil", result)
		}
	})
}
