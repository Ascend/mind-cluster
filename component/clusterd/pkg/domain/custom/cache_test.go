// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package custom cache utils for fault filtered by custom_filter_fault_processor
package custom

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

// TestGetDeletedDevFaultCmForNodeMap tests GetDeletedDevFaultCmForNodeMap method
func TestGetDeletedDevFaultCmForNodeMap(t *testing.T) {
	cache := &faultCache{
		deletedDevFaultCm:        make(map[string]*constant.AdvanceDeviceFaultCm),
		deletedJobFaultDeviceMap: make(map[string][]constant.FaultDevice),
		rwMutex:                  sync.RWMutex{},
	}
	t.Run("get empty map", func(t *testing.T) {
		result := cache.GetDeletedDevFaultCmForNodeMap()
		if result == nil {
			t.Errorf("GetDeletedDevFaultCmForNodeMap() = nil, want empty map")
		}
		if len(result) != 0 {
			t.Errorf("GetDeletedDevFaultCmForNodeMap() length = %v, want 0", len(result))
		}
	})
	t.Run("get map with data", func(t *testing.T) {

		testData := map[string]*constant.AdvanceDeviceFaultCm{
			node1: getMockAdvanceDeviceFaultCmForTest(),
		}
		cache.SetDeletedDevFaultCmForNodeMap(testData)
		result := cache.GetDeletedDevFaultCmForNodeMap()
		if !reflect.DeepEqual(result, testData) {
			t.Errorf("GetDeletedDevFaultCmForNodeMap() = %v, want %v", result, testData)
		}
	})
}

// TestGetDeletedSwitchFaultCmForNodeMap tests GetDeletedSwitchFaultCmForNodeMap method
func TestGetDeletedSwitchFaultCmForNodeMap(t *testing.T) {
	cache := &faultCache{
		deletedSwitchFaultCm:     make(map[string]*constant.SwitchInfo),
		deletedJobFaultDeviceMap: make(map[string][]constant.FaultDevice),
		rwMutex:                  sync.RWMutex{},
	}
	t.Run("get empty map", func(t *testing.T) {
		result := cache.GetDeletedSwitchFaultCmForNodeMap()
		if result == nil {
			t.Errorf("GetDeletedSwitchFaultCmForNodeMap() = nil, want empty map")
		}
		if len(result) != 0 {
			t.Errorf("GetDeletedSwitchFaultCmForNodeMap() length = %v, want 0", len(result))
		}
	})
	t.Run("get map with data", func(t *testing.T) {
		testData := map[string]*constant.SwitchInfo{
			node1: getMockSwitchInfoForTest(),
		}
		cache.SetDeletedSwitchFaultCmForNodeMap(testData)
		result := cache.GetDeletedSwitchFaultCmForNodeMap()
		if !reflect.DeepEqual(result, testData) {
			t.Errorf("GetDeletedSwitchFaultCmForNodeMap() = %v, want %v", result, testData)
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
	cache := &faultCache{
		deletedDevFaultCm:        make(map[string]*constant.AdvanceDeviceFaultCm),
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

// TestSetDeletedDevFaultCmForNodeMap tests SetDeletedDevFaultCmForNodeMap method
func TestSetDeletedDevFaultCmForNodeMap(t *testing.T) {
	cache := &faultCache{
		deletedDevFaultCm:        make(map[string]*constant.AdvanceDeviceFaultCm),
		deletedJobFaultDeviceMap: make(map[string][]constant.FaultDevice),
		rwMutex:                  sync.RWMutex{},
	}
	t.Run("set normal map", func(t *testing.T) {
		testData := map[string]*constant.AdvanceDeviceFaultCm{
			node1: getMockAdvanceDeviceFaultCmForTest(),
		}
		cache.SetDeletedDevFaultCmForNodeMap(testData)
		result := cache.GetDeletedDevFaultCmForNodeMap()
		if !reflect.DeepEqual(result, testData) {
			t.Errorf("SetDeletedDevFaultCmForNodeMap() failed, got %v, want %v", result, testData)
		}
	})
	t.Run("set nil map", func(t *testing.T) {
		cache.SetDeletedDevFaultCmForNodeMap(nil)
		result := cache.GetDeletedDevFaultCmForNodeMap()
		if result != nil {
			t.Errorf("SetDeletedDevFaultCmForNodeMap(nil) failed, got %v, want nil", result)
		}
	})
}

// TestSetDeletedSwitchFaultCmForNodeMap tests SetDeletedSwitchFaultCmForNodeMap method
func TestSetDeletedSwitchFaultCmForNodeMap(t *testing.T) {
	cache := &faultCache{
		deletedSwitchFaultCm:     make(map[string]*constant.SwitchInfo),
		deletedJobFaultDeviceMap: make(map[string][]constant.FaultDevice),
		rwMutex:                  sync.RWMutex{},
	}
	t.Run("set normal map", func(t *testing.T) {
		testData := map[string]*constant.SwitchInfo{
			node1: getMockSwitchInfoForTest(),
		}
		cache.SetDeletedSwitchFaultCmForNodeMap(testData)
		result := cache.GetDeletedSwitchFaultCmForNodeMap()
		if !reflect.DeepEqual(result, testData) {
			t.Errorf("SetDeletedSwitchFaultCmForNodeMap() failed, got %v, want %v", result, testData)
		}
	})
	t.Run("set nil map", func(t *testing.T) {
		cache.SetDeletedSwitchFaultCmForNodeMap(nil)
		result := cache.GetDeletedSwitchFaultCmForNodeMap()
		if result != nil {
			t.Errorf("SetDeletedSwitchFaultCmForNodeMap(nil) failed, got %v, want nil", result)
		}
	})
}

// TestSetDeletedJobFaultDeviceMap tests SetDeletedJobFaultDeviceMap method
func TestSetDeletedJobFaultDeviceMap(t *testing.T) {
	cache := &faultCache{
		deletedDevFaultCm:        make(map[string]*constant.AdvanceDeviceFaultCm),
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
