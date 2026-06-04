/*
   Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package ub_device for ub device info
package ub_device

import (
	"context"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"ascend-common/common-utils/hwlog"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/utils"
)

func init() {
	_ = hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestNewUbRdmaDeviceSpec(t *testing.T) {
	convey.Convey("When newUbRdmaDeviceSpec is called", t, func() {
		rds := newUbRdmaDeviceSpec([]string{"rdma_cm", "uverbs"})
		convey.Convey("Then it should return a valid RdmaDeviceSpec", func() {
			convey.So(rds, convey.ShouldNotBeNil)
		})
	})
}

func TestNewUbRdmaDeviceSpecEmpty(t *testing.T) {
	convey.Convey("When rdmaDevs is empty", t, func() {
		rds := newUbRdmaDeviceSpec([]string{})
		convey.Convey("Then it should return a valid RdmaDeviceSpec", func() {
			convey.So(rds, convey.ShouldNotBeNil)
		})
	})
}

func TestUbRdmaDeviceSpecGetSuccess(t *testing.T) {
	rds := newUbRdmaDeviceSpec([]string{"rdma_cm", "uverbs"})
	ubID := "ub0"

	patches := gomonkey.ApplyFunc(utils.GetRdmaDevicesForUbdev,
		func(ubID string) []string {
			return []string{"/dev/infiniband/rdma_cm", "/dev/infiniband/uverbs0"}
		})
	defer patches.Reset()

	convey.Convey("When GetRdmaDevicesForUbdev returns devices", t, func() {
		specs := rds.Get(ubID)
		convey.Convey("Then device specs should be returned", func() {
			convey.So(len(specs), convey.ShouldEqual, 2)
			convey.So(specs[0].HostPath, convey.ShouldEqual, "/dev/infiniband/rdma_cm")
			convey.So(specs[0].Permissions, convey.ShouldEqual, "rwm")
		})
	})
}

func TestUbRdmaDeviceSpecGetEmpty(t *testing.T) {
	rds := newUbRdmaDeviceSpec([]string{"rdma_cm"})
	ubID := "ub0"

	patches := gomonkey.ApplyFunc(utils.GetRdmaDevicesForUbdev,
		func(ubID string) []string {
			return []string{}
		})
	defer patches.Reset()

	convey.Convey("When GetRdmaDevicesForUbdev returns empty", t, func() {
		specs := rds.Get(ubID)
		convey.Convey("Then empty device specs should be returned", func() {
			convey.So(len(specs), convey.ShouldEqual, 0)
		})
	})
}

func TestUbRdmaDeviceSpecGetNil(t *testing.T) {
	rds := newUbRdmaDeviceSpec([]string{"rdma_cm"})
	ubID := "ub0"

	patches := gomonkey.ApplyFunc(utils.GetRdmaDevicesForUbdev,
		func(ubID string) []string {
			return nil
		})
	defer patches.Reset()

	convey.Convey("When GetRdmaDevicesForUbdev returns nil", t, func() {
		specs := rds.Get(ubID)
		convey.Convey("Then empty device specs should be returned", func() {
			convey.So(len(specs), convey.ShouldEqual, 0)
		})
	})
}

func TestUbRdmaDeviceSpecVerifyRdmaSpecSuccess(t *testing.T) {
	rds := newUbRdmaDeviceSpec([]string{"rdma_cm", "uverbs"})
	devSpecs := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/infiniband/rdma_cm", ContainerPath: "/dev/infiniband/rdma_cm", Permissions: "rwm"},
		{HostPath: "/dev/infiniband/uverbs0", ContainerPath: "/dev/infiniband/uverbs0", Permissions: "rwm"},
	}

	convey.Convey("When all required RDMA devices are present", t, func() {
		err := rds.VerifyRdmaSpec(devSpecs)
		convey.Convey("Then verification should succeed", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUbRdmaDeviceSpecVerifyRdmaSpecMissing(t *testing.T) {
	rds := newUbRdmaDeviceSpec([]string{"rdma_cm", "uverbs", "missing"})
	devSpecs := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/infiniband/rdma_cm", ContainerPath: "/dev/infiniband/rdma_cm", Permissions: "rwm"},
		{HostPath: "/dev/infiniband/uverbs0", ContainerPath: "/dev/infiniband/uverbs0", Permissions: "rwm"},
	}

	convey.Convey("When a required RDMA device is missing", t, func() {
		err := rds.VerifyRdmaSpec(devSpecs)
		convey.Convey("Then verification should fail", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "missing")
		})
	})
}

func TestUbRdmaDeviceSpecVerifyRdmaSpecEmptyRequired(t *testing.T) {
	rds := newUbRdmaDeviceSpec([]string{})
	devSpecs := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/infiniband/uverbs0", ContainerPath: "/dev/infiniband/uverbs0", Permissions: "rwm"},
	}

	convey.Convey("When no required RDMA devices", t, func() {
		err := rds.VerifyRdmaSpec(devSpecs)
		convey.Convey("Then verification should succeed", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUbRdmaDeviceSpecVerifyRdmaSpecEmptySpecs(t *testing.T) {
	rds := newUbRdmaDeviceSpec([]string{"rdma_cm"})
	devSpecs := []*pluginapi.DeviceSpec{}

	convey.Convey("When rdmaDevSpecs is empty and required devices exist", t, func() {
		err := rds.VerifyRdmaSpec(devSpecs)
		convey.Convey("Then verification should fail", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestContainsUbRdmaDevExists(t *testing.T) {
	devSpecs := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/infiniband/rdma_cm", ContainerPath: "/dev/infiniband/rdma_cm", Permissions: "rwm"},
		{HostPath: "/dev/infiniband/uverbs0", ContainerPath: "/dev/infiniband/uverbs0", Permissions: "rwm"},
	}

	convey.Convey("When rdmaDev is contained in devSpecs", t, func() {
		result := containsUbRdmaDev(devSpecs, "rdma_cm")
		convey.Convey("Then it should return true", func() {
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

func TestContainsUbRdmaDevExistsUverbs(t *testing.T) {
	devSpecs := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/infiniband/rdma_cm", ContainerPath: "/dev/infiniband/rdma_cm", Permissions: "rwm"},
		{HostPath: "/dev/infiniband/uverbs0", ContainerPath: "/dev/infiniband/uverbs0", Permissions: "rwm"},
	}

	convey.Convey("When rdmaDev matches with suffix", t, func() {
		result := containsUbRdmaDev(devSpecs, "uverbs")
		convey.Convey("Then it should return true", func() {
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

func TestContainsUbRdmaDevNotExists(t *testing.T) {
	devSpecs := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/infiniband/rdma_cm", ContainerPath: "/dev/infiniband/rdma_cm", Permissions: "rwm"},
	}

	convey.Convey("When rdmaDev is not contained in devSpecs", t, func() {
		result := containsUbRdmaDev(devSpecs, "uverbs")
		convey.Convey("Then it should return false", func() {
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestContainsUbRdmaDev_EmptySpecs(t *testing.T) {
	devSpecs := []*pluginapi.DeviceSpec{}

	convey.Convey("When devSpecs is empty", t, func() {
		result := containsUbRdmaDev(devSpecs, "rdma_cm")
		convey.Convey("Then it should return false", func() {
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}
