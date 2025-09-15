// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package collector collect information to process
package collector

import (
	"sync"
	"testing"

	"clusterd/pkg/common/constant"
)

const (
	CmName  = "CmName"
	CmName2 = "CmName2"
)

func resetDeviceCmCollector() {
	DeviceCmCollectBuffer = &ConfigmapCollectBuffer[*constant.AdvanceDeviceFaultCm]{
		mutex:    sync.Mutex{},
		buffer:   make(map[string]*[]constant.InformerCmItem[*constant.AdvanceDeviceFaultCm]),
		lastItem: make(map[string]constant.InformerCmItem[*constant.AdvanceDeviceFaultCm]),
	}
}

func TestCmInfoCollector(t *testing.T) {
	testDeviceInfoCollector(t)
	testNodeCollector(t)
	testSwitchInfoCollector(t)
}

func testDeviceInfoCollector(t *testing.T) {
	t.Run("TestDeviceInfoCollector", func(t *testing.T) {
		DeviceInfoCollector(nil, nil, constant.AddOperator)
		if len(DeviceCmCollectBuffer.buffer) != 0 {
			t.Error("TestDeviceInfoCollector failed, when newDevInfo is nil")
		}
		deviceInfo := &constant.DeviceInfo{
			CmName: CmName,
		}
		DeviceInfoCollector(nil, deviceInfo, constant.AddOperator)
		if len(*DeviceCmCollectBuffer.buffer[CmName]) != 1 {
			t.Error("TestDeviceInfoCollector failed")
		}
		DeviceInfoCollector(nil, deviceInfo, constant.DeleteOperator)
		if len(*DeviceCmCollectBuffer.buffer[CmName]) != 2 {
			t.Error("TestDeviceInfoCollector failed")
		}
	})
}

func testNodeCollector(t *testing.T) {
	t.Run("TestNodeInfoCollector", func(t *testing.T) {
		NodeCollector(nil, nil, constant.AddOperator)
		if len(NodeCmCollectBuffer.buffer) != 0 {
			t.Error("TestNodeInfoCollector failed, when newNodeInfo is nil")
		}
		nodeInfo := &constant.NodeInfo{
			CmName: CmName,
		}
		NodeCollector(nil, nodeInfo, constant.AddOperator)
		if len(*NodeCmCollectBuffer.buffer[CmName]) != 1 {
			t.Error("TestNodeInfoCollector failed")
		}
		NodeCollector(nil, nodeInfo, constant.DeleteOperator)
		if len(*NodeCmCollectBuffer.buffer[CmName]) != 2 {
			t.Error("TestNodeInfoCollector failed")
		}
	})
}

func testSwitchInfoCollector(t *testing.T) {
	t.Run("TestSwitchInfoCollector", func(t *testing.T) {
		SwitchInfoCollector(nil, nil, constant.AddOperator)
		if len(SwitchCmCollectBuffer.buffer) != 0 {
			t.Error("TestSwitchInfoCollector failed, when newSwitchInfo is nil")
		}
		switchInfo := &constant.SwitchInfo{
			CmName: CmName,
		}
		SwitchInfoCollector(nil, switchInfo, constant.AddOperator)
		if len(*SwitchCmCollectBuffer.buffer[CmName]) != 1 {
			t.Error("TestSwitchInfoCollector failed")
		}
		SwitchInfoCollector(nil, switchInfo, constant.DeleteOperator)
		if len(*SwitchCmCollectBuffer.buffer[CmName]) != 2 {
			t.Error("TestSwitchInfoCollector failed")
		}
	})
}

func TestUpdateInfoSuccess(t *testing.T) {
	resetDeviceCmCollector()
	t.Run("push info with update success", func(t *testing.T) {
		deviceInfo := &constant.DeviceInfo{
			CmName: CmName,
		}
		DeviceInfoCollector(nil, deviceInfo, constant.AddOperator)
		if len(*DeviceCmCollectBuffer.buffer[CmName]) != 1 {
			t.Error("TestDeviceInfoCollector failed")
		}
		DeviceInfoCollector(nil, deviceInfo, constant.DeleteOperator)
		if len(*DeviceCmCollectBuffer.buffer[CmName]) != 2 {
			t.Error("TestDeviceInfoCollector failed")
		}
	})
}

func TestPopInfoCorrect(t *testing.T) {
	resetDeviceCmCollector()
	t.Run("push info and pop info correctly", func(t *testing.T) {
		deviceInfo := &constant.DeviceInfo{
			CmName: CmName,
		}
		deviceInfo2 := &constant.DeviceInfo{
			CmName: CmName2,
		}
		DeviceInfoCollector(nil, deviceInfo, constant.AddOperator)
		DeviceInfoCollector(nil, deviceInfo2, constant.AddOperator)
		DeviceInfoCollector(nil, deviceInfo, constant.DeleteOperator)
		popInfo := DeviceCmCollectBuffer.Pop()
		if len(popInfo) != 2 {
			t.Error("TestPopInfoCorrect failed")
		}
		popInfo = DeviceCmCollectBuffer.Pop()
		if len(popInfo) != 1 {
			t.Error("TestPopInfoCorrect failed")
		}
		popInfo = DeviceCmCollectBuffer.Pop()
		if len(popInfo) != 0 {
			t.Error("TestPopInfoCorrect failed")
		}
	})
}
