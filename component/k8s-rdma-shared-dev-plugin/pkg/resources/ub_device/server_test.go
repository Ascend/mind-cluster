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
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	registerapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"

	"ascend-common/common-utils/hwlog"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/common"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
)

func init() {
	_ = hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func newTestUbServer() *ubResourceServer {
	return &ubResourceServer{
		resourceName: "rdma-ub/ub_dev",
		watchMode:    false,
		socketName:   "ub_dev.sock",
		socketPath:   "/tmp/ub_dev.sock",
		stopWatcher:  make(chan bool, 1),
		deviceSpec: []*pluginapi.DeviceSpec{
			{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
		},
		ubDevices: []types.Device{
			&ubDevice{ubID: "ub0", ifName: "enp0s1"},
		},
		devs: []*pluginapi.Device{
			{ID: "0", Health: pluginapi.Healthy},
		},
		updateResource: make(chan bool, 1),
		rsConnector:    &resourcesServerPort{},
	}
}

func TestCreateUbVirtualDevicesEmptySpec(t *testing.T) {
	convey.Convey("When deviceSpec is empty", t, func() {
		devs := createUbVirtualDevices(100, []*pluginapi.DeviceSpec{}, "test")
		convey.Convey("Then empty device slice should be returned", func() {
			convey.So(len(devs), convey.ShouldEqual, 0)
		})
	})
}

func TestCreateUbVirtualDevicesWithSpec(t *testing.T) {
	spec := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
	}

	convey.Convey("When deviceSpec is not empty", t, func() {
		devs := createUbVirtualDevices(3, spec, "test")
		convey.Convey("Then rdmaHcaMax number of devices should be created", func() {
			convey.So(len(devs), convey.ShouldEqual, 3)
			convey.So(devs[0].ID, convey.ShouldEqual, "0")
			convey.So(devs[1].ID, convey.ShouldEqual, "1")
			convey.So(devs[2].ID, convey.ShouldEqual, "2")
			convey.So(devs[0].Health, convey.ShouldEqual, pluginapi.Healthy)
		})
	})
}

func TestNewUbResourceServerInvalidRdmaHcaMax(t *testing.T) {
	config := &types.UserConfig{RdmaHcaMax: -1}

	convey.Convey("When rdmaHcaMax is negative", t, func() {
		_, err := NewUbResourceServer(config, nil, false, "sock", false)
		convey.Convey("Then an error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "rdmaHcaMax")
		})
	})
}

func TestNewUbResourceServerEmptyResourcePrefix(t *testing.T) {
	config := &types.UserConfig{RdmaHcaMax: 100, ResourcePrefix: ""}

	convey.Convey("When resourcePrefix is empty", t, func() {
		_, err := NewUbResourceServer(config, nil, false, "sock", false)
		convey.Convey("Then an error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "resourcePrefix")
		})
	})
}

func TestNewUbResourceServerSuccess(t *testing.T) {
	config := &types.UserConfig{
		ResourceName:   "ub_dev",
		ResourcePrefix: "mindx.com",
		RdmaHcaMax:     100,
	}
	devices := []types.Device{&ubDevice{ubID: "ub0", ifName: "enp0s1"}}

	patches := gomonkey.ApplyFunc(getUbDevicesSpec,
		func(devices []types.Device) []*pluginapi.DeviceSpec {
			return []*pluginapi.DeviceSpec{
				{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
			}
		})
	defer patches.Reset()

	convey.Convey("When valid config is provided", t, func() {
		rs, err := NewUbResourceServer(config, devices, false, "sock", false)
		convey.Convey("Then server should be created successfully", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(rs, convey.ShouldNotBeNil)
		})
	})
}

func TestNewUbResourceServerWatchMode(t *testing.T) {
	config := &types.UserConfig{
		ResourceName:   "ub_dev",
		ResourcePrefix: "mindx.com",
		RdmaHcaMax:     50,
	}

	patches := gomonkey.ApplyFunc(getUbDevicesSpec,
		func(devices []types.Device) []*pluginapi.DeviceSpec {
			return []*pluginapi.DeviceSpec{}
		})
	defer patches.Reset()

	convey.Convey("When watchMode is true", t, func() {
		rs, err := NewUbResourceServer(config, nil, true, "sock", false)
		convey.Convey("Then server should be created in watch mode", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(rs, convey.ShouldNotBeNil)
		})
	})
}

func TestNewUbResourceServerWithCDI(t *testing.T) {
	config := &types.UserConfig{
		ResourceName:   "ub_dev",
		ResourcePrefix: "mindx.com",
		RdmaHcaMax:     100,
	}

	patches := gomonkey.ApplyFunc(getUbDevicesSpec,
		func(devices []types.Device) []*pluginapi.DeviceSpec {
			return []*pluginapi.DeviceSpec{}
		})
	defer patches.Reset()

	convey.Convey("When useCdi is true", t, func() {
		rs, err := NewUbResourceServer(config, nil, false, "sock", true)
		convey.Convey("Then server should be created with CDI enabled", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(rs, convey.ShouldNotBeNil)
		})
	})
}

func TestUbResourceServerGetDevicePluginOptions(t *testing.T) {
	rs := newTestUbServer()

	convey.Convey("When GetDevicePluginOptions is called", t, func() {
		opts, err := rs.GetDevicePluginOptions(context.Background(), &pluginapi.Empty{})
		convey.Convey("Then PreStartRequired should be false", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(opts, convey.ShouldNotBeNil)
			convey.So(opts.PreStartRequired, convey.ShouldBeFalse)
		})
	})
}

func TestUbResourceServerPreStartContainer(t *testing.T) {
	rs := newTestUbServer()

	convey.Convey("When PreStartContainer is called", t, func() {
		resp, err := rs.PreStartContainer(context.Background(), &pluginapi.PreStartContainerRequest{})
		convey.Convey("Then empty response should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
		})
	})
}

func TestUbResourceServerGetPreferredAllocation(t *testing.T) {
	rs := newTestUbServer()

	convey.Convey("When GetPreferredAllocation is called", t, func() {
		resp, err := rs.GetPreferredAllocation(context.Background(), &pluginapi.PreferredAllocationRequest{})
		convey.Convey("Then empty response should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
		})
	})
}

func TestUbResourceServerGetInfo(t *testing.T) {
	rs := newTestUbServer()

	convey.Convey("When GetInfo is called", t, func() {
		info, err := rs.GetInfo(context.Background(), &registerapi.InfoRequest{})
		convey.Convey("Then plugin info should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(info.Type, convey.ShouldEqual, "DevicePlugin")
			convey.So(info.Name, convey.ShouldEqual, rs.resourceName)
			convey.So(info.Endpoint, convey.ShouldEqual, rs.socketName)
		})
	})
}

func TestUbResourceServerNotifyRegistrationStatusRegistered(t *testing.T) {
	rs := newTestUbServer()
	status := &registerapi.RegistrationStatus{PluginRegistered: true}

	convey.Convey("When plugin is registered", t, func() {
		resp, err := rs.NotifyRegistrationStatus(context.Background(), status)
		convey.Convey("Then success response should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
		})
	})
}

func TestUbResourceServerNotifyRegistrationStatusNotRegistered(t *testing.T) {
	rs := newTestUbServer()
	status := &registerapi.RegistrationStatus{
		PluginRegistered: false,
		Error:            "registration failed",
	}

	convey.Convey("When plugin is not registered", t, func() {
		resp, err := rs.NotifyRegistrationStatus(context.Background(), status)
		convey.Convey("Then it should still return without error", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
		})
	})
}

func TestUbResourceServerAllocate(t *testing.T) {
	rs := newTestUbServer()
	rs.deviceSpec = []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
	}

	convey.Convey("When Allocate is called", t, func() {
		reqs := &pluginapi.AllocateRequest{
			ContainerRequests: []*pluginapi.ContainerAllocateRequest{{}, {}},
		}
		resp, err := rs.Allocate(context.Background(), reqs)
		convey.Convey("Then device specs should be allocated to each container", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(resp.ContainerResponses), convey.ShouldEqual, 2)
			convey.So(len(resp.ContainerResponses[0].Devices), convey.ShouldEqual, 1)
		})
	})
}

func TestUbResourceServerUpdateDevicesChanged(t *testing.T) {
	rs := newTestUbServer()
	oldSpec := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
	}
	rs.deviceSpec = oldSpec
	newDevices := []types.Device{&ubDevice{ubID: "ub1", ifName: "enp0s2"}}

	patches := gomonkey.ApplyFunc(getUbDevicesSpec,
		func(devices []types.Device) []*pluginapi.DeviceSpec {
			return []*pluginapi.DeviceSpec{
				{HostPath: "/dev/ub1", ContainerPath: "/dev/ub1", Permissions: "rw"},
			}
		})
	patches.ApplyFunc(common.DevicesChanged,
		func(deviceList, newDeviceList []*pluginapi.DeviceSpec) bool {
			return true
		})
	defer patches.Reset()

	convey.Convey("When devices have changed", t, func() {
		rs.UpdateDevices(newDevices)
		convey.Convey("Then deviceSpec should be updated", func() {
			convey.So(len(rs.deviceSpec), convey.ShouldEqual, 1)
			convey.So(rs.deviceSpec[0].HostPath, convey.ShouldEqual, "/dev/ub1")
		})
	})
}

func TestUbResourceServerUpdateDevicesNotChanged(t *testing.T) {
	rs := newTestUbServer()
	oldSpec := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
	}
	rs.deviceSpec = oldSpec

	patches := gomonkey.ApplyFunc(getUbDevicesSpec,
		func(devices []types.Device) []*pluginapi.DeviceSpec {
			return oldSpec
		})
	patches.ApplyFunc(common.DevicesChanged,
		func(deviceList, newDeviceList []*pluginapi.DeviceSpec) bool {
			return false
		})
	defer patches.Reset()

	convey.Convey("When devices have not changed", t, func() {
		rs.UpdateDevices(nil)
		convey.Convey("Then updateResource channel should not be signaled", func() {
			convey.So(len(rs.updateResource), convey.ShouldEqual, 0)
		})
	})
}

func TestUbResourceServerCleanupSuccess(t *testing.T) {
	rs := newTestUbServer()

	patches := gomonkey.ApplyFunc(os.Remove, func(name string) error {
		return nil
	})
	defer patches.Reset()

	convey.Convey("When socket file exists", t, func() {
		err := rs.cleanup()
		convey.Convey("Then no error should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUbResourceServerCleanupFileNotExist(t *testing.T) {
	rs := newTestUbServer()

	patches := gomonkey.ApplyFunc(os.Remove, func(name string) error {
		return os.ErrNotExist
	})
	defer patches.Reset()

	convey.Convey("When socket file does not exist", t, func() {
		err := rs.cleanup()
		convey.Convey("Then no error should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUbResourceServerCleanupError(t *testing.T) {
	rs := newTestUbServer()

	patches := gomonkey.ApplyFunc(os.Remove, func(name string) error {
		return errors.New("permission denied")
	})
	defer patches.Reset()

	convey.Convey("When os.Remove fails", t, func() {
		err := rs.cleanup()
		convey.Convey("Then the error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestUbResourceServerStopNilConnector(t *testing.T) {
	rs := &ubResourceServer{
		resourceName: "test",
		rsConnector:  nil,
	}

	convey.Convey("When rsConnector is nil", t, func() {
		err := rs.Stop()
		convey.Convey("Then no error should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUbResourceServerStopNilServer(t *testing.T) {
	rs := newTestUbServer()
	rs.rsConnector = &resourcesServerPort{server: nil}

	convey.Convey("When grpc server is nil", t, func() {
		err := rs.Stop()
		convey.Convey("Then no error should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUbResourceServerRestartNilConnector(t *testing.T) {
	rs := &ubResourceServer{
		resourceName: "test",
		rsConnector:  nil,
	}

	convey.Convey("When rsConnector is nil during Restart", t, func() {
		err := rs.Restart()
		convey.Convey("Then an error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "grpc server instance not found")
		})
	})
}

func TestUbResourceServerRestartNilServer(t *testing.T) {
	rs := newTestUbServer()
	rs.rsConnector = &resourcesServerPort{server: nil}

	convey.Convey("When grpc server is nil during Restart", t, func() {
		err := rs.Restart()
		convey.Convey("Then an error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestResourcesServerPortGetServer(t *testing.T) {
	rsc := &resourcesServerPort{}

	convey.Convey("When server is nil", t, func() {
		s := rsc.GetServer()
		convey.Convey("Then nil should be returned", func() {
			convey.So(s, convey.ShouldBeNil)
		})
	})
}

func TestResourcesServerPortCreateServer(t *testing.T) {
	rsc := &resourcesServerPort{}

	convey.Convey("When CreateServer is called", t, func() {
		rsc.CreateServer()
		convey.Convey("Then server should not be nil", func() {
			convey.So(rsc.server, convey.ShouldNotBeNil)
		})
	})
}

func TestResourcesServerPortDeleteServer(t *testing.T) {
	rsc := &resourcesServerPort{}
	rsc.CreateServer()

	convey.Convey("When DeleteServer is called", t, func() {
		rsc.DeleteServer()
		convey.Convey("Then server should be nil", func() {
			convey.So(rsc.server, convey.ShouldBeNil)
		})
	})
}

func TestResourcesServerPortStopNilServer(t *testing.T) {
	rsc := &resourcesServerPort{}

	convey.Convey("When server is nil during Stop", t, func() {
		rsc.Stop()
		convey.Convey("Then it should return immediately", func() {
			convey.So(true, convey.ShouldBeTrue)
		})
	})
}

func TestResourcesServerPortCloseListener(t *testing.T) {
	rsc := &resourcesServerPort{}

	convey.Convey("When listener is nil", t, func() {
		rsc.CloseListener(nil)
		convey.Convey("Then it should not panic", func() {
			convey.So(true, convey.ShouldBeTrue)
		})
	})
}

func TestResourcesServerPortClose(t *testing.T) {
	rsc := &resourcesServerPort{}

	convey.Convey("When conn is nil", t, func() {
		rsc.Close(nil)
		convey.Convey("Then it should not panic", func() {
			convey.So(true, convey.ShouldBeTrue)
		})
	})
}

func TestGetUbDevicesSpecEmpty(t *testing.T) {
	convey.Convey("When devices list is empty", t, func() {
		spec := getUbDevicesSpec([]types.Device{})
		convey.Convey("Then empty spec should be returned", func() {
			convey.So(len(spec), convey.ShouldEqual, 0)
		})
	})
}

func TestGetUbDevicesSpecWithRdmaSpec(t *testing.T) {
	device := &ubDevice{
		ubID: "ub0",
		rdmaSpec: []*pluginapi.DeviceSpec{
			{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
		},
	}

	convey.Convey("When device has RDMA spec", t, func() {
		spec := getUbDevicesSpec([]types.Device{device})
		convey.Convey("Then device specs should be returned", func() {
			convey.So(len(spec), convey.ShouldEqual, 1)
			convey.So(spec[0].HostPath, convey.ShouldEqual, "/dev/ub0")
		})
	})
}
