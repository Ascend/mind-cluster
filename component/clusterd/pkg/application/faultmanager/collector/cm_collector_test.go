// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package collector

import (
	"testing"

	"clusterd/pkg/common/constant"
)

func TestDeviceInfoCollector(t *testing.T) {
	InitCmCollectBuffer()
	t.Run("TestDeviceInfoCollector", func(t *testing.T) {
		deviceInfo := &constant.DeviceInfo{
			CmName: "test",
		}
		DeviceInfoCollector(nil, deviceInfo, constant.AddOperator)
		if len(DeviceCmCollectBuffer.buffer) == 0 {
			t.Error("TestDeviceInfoCollector failed")
		}
	})
	t.Run("TestNodeInfoCollector", func(t *testing.T) {
		nodeInfo := &constant.NodeInfo{
			CmName: "test",
		}
		NodeCollector(nil, nodeInfo, constant.AddOperator)
		if len(NodeCmCollectBuffer.buffer) == 0 {
			t.Error("TestNodeInfoCollector failed")
		}
	})
	t.Run("TestSwitchInfoCollector", func(t *testing.T) {
		switchInfo := &constant.SwitchInfo{
			CmName: "test",
		}
		SwitchInfoCollector(nil, switchInfo, constant.AddOperator)
		if len(SwitchCmCollectBuffer.buffer) == 0 {
			t.Error("TestSwitchInfoCollector failed")
		}
	})
}
