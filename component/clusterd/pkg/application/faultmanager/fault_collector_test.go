// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"testing"

	"clusterd/pkg/common/constant"
)

func TestDeviceInfoCollector(t *testing.T) {
	GlobalFaultProcessCenter = &FaultProcessCenter{
		deviceCenter:      newDeviceFaultProcessCenter(),
		nodeCenter:        newNodeFaultProcessCenter(),
		switchCenter:      newSwitchFaultProcessCenter(),
		faultJobCenter:    newFaultJobProcessCenter(),
		notifyProcessChan: make(chan int, 1000),
	}
	t.Run("TestDeviceInfoCollector", func(t *testing.T) {
		deviceInfo := &constant.DeviceInfo{
			CmName: "test",
		}
		DeviceInfoCollector(nil, deviceInfo, constant.AddOperator)
		if len(GlobalFaultProcessCenter.deviceCenter.getOriginalCm()) == 0 {
			t.Error("TestDeviceInfoCollector failed")
		}
		DeviceInfoCollector(nil, deviceInfo, constant.DeleteOperator)
		if len(GlobalFaultProcessCenter.deviceCenter.getOriginalCm()) != 0 {
			t.Error("TestDeviceInfoCollector failed")
		}
	})
	t.Run("TestNodeInfoCollector", func(t *testing.T) {
		nodeInfo := &constant.NodeInfo{
			CmName: "test",
		}
		NodeCollector(nil, nodeInfo, constant.AddOperator)
		if len(GlobalFaultProcessCenter.nodeCenter.getOriginalCm()) == 0 {
			t.Error("TestNodeInfoCollector failed")
		}
		NodeCollector(nil, nodeInfo, constant.DeleteOperator)
		if len(GlobalFaultProcessCenter.nodeCenter.getOriginalCm()) != 0 {
			t.Error("TestNodeInfoCollector failed")
		}
	})
	t.Run("TestSwitchInfoCollector", func(t *testing.T) {
		switchInfo := &constant.SwitchInfo{
			CmName: "test",
		}
		SwitchInfoCollector(nil, switchInfo, constant.AddOperator)
		if len(GlobalFaultProcessCenter.switchCenter.getOriginalCm()) == 0 {
			t.Error("TestSwitchInfoCollector failed")
		}
		SwitchInfoCollector(nil, switchInfo, constant.DeleteOperator)
		if len(GlobalFaultProcessCenter.switchCenter.getOriginalCm()) != 0 {
			t.Error("TestSwitchInfoCollector failed")
		}
	})
}

func TestShouldNotInformer(t *testing.T) {
	GlobalFaultProcessCenter = &FaultProcessCenter{
		deviceCenter:      newDeviceFaultProcessCenter(),
		nodeCenter:        newNodeFaultProcessCenter(),
		switchCenter:      newSwitchFaultProcessCenter(),
		faultJobCenter:    newFaultJobProcessCenter(),
		notifyProcessChan: make(chan int, 1000),
	}
	t.Run("TestShouldNotInformer", func(t *testing.T) {
		informInfoUpdate(nil, -1, true)
		select {
		case <-GlobalFaultProcessCenter.notifyProcessChan:
			t.Error("TestShouldNotInformer fail")
		default:
		}
	})
}
