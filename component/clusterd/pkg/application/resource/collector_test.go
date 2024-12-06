// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

//go:build !race

// Package resource a series of resource test function
package resource

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

const (
	testCmName = "test-node-name"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

//go:noinline
func TestDeviceInfoCollector(t *testing.T) {
	convey.Convey("test DeviceInfoCollector", t, func() {
		convey.Convey("add new device info", func() {
			oldDevInfo := &constant.DeviceInfo{}
			newDevInfo := &constant.DeviceInfo{
				CmName: testCmName,
			}
			DeviceInfoCollector(oldDevInfo, newDevInfo, constant.AddOperator)
			convey.So(len(cmManager.deviceInfoMap), convey.ShouldEqual, 1)
		})
		convey.Convey("update device info", func() {
			oldDevInfo := &constant.DeviceInfo{}
			newDevInfo := &constant.DeviceInfo{
				CmName: testCmName,
			}
			DeviceInfoCollector(oldDevInfo, newDevInfo, constant.UpdateOperator)
			convey.So(cmManager.deviceInfoMap[testCmName], convey.ShouldEqual, newDevInfo)
		})
		convey.Convey("delete device info", func() {
			oldDevInfo := &constant.DeviceInfo{}
			newDevInfo := &constant.DeviceInfo{
				CmName: testCmName,
			}
			DeviceInfoCollector(oldDevInfo, newDevInfo, constant.AddOperator)
			convey.So(len(cmManager.deviceInfoMap), convey.ShouldEqual, 1)
			DeviceInfoCollector(oldDevInfo, newDevInfo, constant.DeleteOperator)
			convey.So(len(cmManager.deviceInfoMap), convey.ShouldEqual, 0)
		})
	})
}

//go:noinline
func TestNodeCollector(t *testing.T) {
	convey.Convey("TestNodeCollector", t, func() {
		convey.Convey("add new node info", func() {
			oldDevInfo := &constant.NodeInfo{}
			newDevInfo := &constant.NodeInfo{
				CmName: testCmName,
			}
			NodeCollector(oldDevInfo, newDevInfo, constant.AddOperator)
			convey.So(len(cmManager.nodeInfoMap), convey.ShouldEqual, 1)
		})
		convey.Convey("update node info", func() {
			oldDevInfo := &constant.NodeInfo{}
			newDevInfo := &constant.NodeInfo{
				CmName: testCmName,
			}
			NodeCollector(oldDevInfo, newDevInfo, constant.UpdateOperator)
			convey.So(cmManager.nodeInfoMap[testCmName], convey.ShouldEqual, newDevInfo)
		})
		convey.Convey("delete node info", func() {
			oldDevInfo := &constant.NodeInfo{}
			newDevInfo := &constant.NodeInfo{
				CmName: testCmName,
			}
			NodeCollector(oldDevInfo, newDevInfo, constant.AddOperator)
			convey.So(len(cmManager.nodeInfoMap), convey.ShouldEqual, 1)
			NodeCollector(oldDevInfo, newDevInfo, constant.DeleteOperator)
			convey.So(len(cmManager.nodeInfoMap), convey.ShouldEqual, 0)
		})
	})
}
