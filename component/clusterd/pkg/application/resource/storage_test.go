// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

//go:build !race

// Package resource a series of resource test function
package resource

import (
	"context"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/switchinfo"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

//go:noinline
func TestSaveDeviceInfoCM(t *testing.T) {
	convey.Convey("Test saveDeviceInfoCM with different data", t, func() {
		patch := gomonkey.ApplyFunc(device.BusinessDataIsNotEqual,
			func(_ *constant.DeviceInfo, _ *constant.DeviceInfo) bool { return false })
		defer patch.Reset()
		saveDeviceInfoCM(&constant.DeviceInfo{})
		convey.So(len(cmManager.deviceInfoMap), convey.ShouldEqual, 1)
	})
}

//go:noinline
func TestSaveSwitchInfoCM(t *testing.T) {
	convey.Convey("Test saveSwitchInfoCM with different data", t, func() {
		patch := gomonkey.ApplyFunc(switchinfo.BusinessDataIsNotEqual,
			func(_, _ *constant.SwitchInfo) bool { return false })
		defer patch.Reset()
		saveSwitchInfoCM(&constant.SwitchInfo{})
		convey.So(len(cmManager.switchInfoMap), convey.ShouldEqual, 1)
	})
}

//go:noinline
func TestSaveNodeInfoCM(t *testing.T) {
	convey.Convey("Test saveNodeInfoCM with different data", t, func() {
		patch := gomonkey.ApplyFunc(node.BusinessDataIsNotEqual,
			func(_, _ *constant.NodeInfo) bool { return false })
		defer patch.Reset()
		saveNodeInfoCM(&constant.NodeInfo{})
		convey.So(len(cmManager.nodeInfoMap), convey.ShouldEqual, 1)
	})
}
