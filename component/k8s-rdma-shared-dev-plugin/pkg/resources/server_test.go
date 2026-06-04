// Copyright 2026 Huawei Technologies Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package resources for rdma device
package resources

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	registerapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/common"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
)

type mockDevice struct {
	name     string
	rdmaSpec []*pluginapi.DeviceSpec
}

func (d *mockDevice) GetName() string                      { return d.name }
func (d *mockDevice) GetVendor() string                    { return "" }
func (d *mockDevice) GetDeviceID() string                  { return "" }
func (d *mockDevice) GetDriver() string                    { return "" }
func (d *mockDevice) GetRdmaSpec() []*pluginapi.DeviceSpec { return d.rdmaSpec }
func (d *mockDevice) GetIfName() string                    { return "" }
func (d *mockDevice) GetLinkType() string                  { return "" }

func newTestResourceServer(config *types.UserConfig, devices []types.Device, watchMode bool, useCdi bool) *resourceServer {
	return &resourceServer{
		resourceName:    "test-prefix/test-resource",
		watchMode:       watchMode,
		socketName:      "test.sock",
		socketPath:      "/tmp/test.sock",
		stopWatcher:     make(chan bool, 1),
		updateResource:  make(chan bool, 1),
		health:          make(chan *pluginapi.Device, 1),
		rsConnector:     &resourcesServerPort{},
		rdmaHcaMax:      config.RdmaHcaMax,
		devs:            []*pluginapi.Device{{ID: "0", Health: pluginapi.Healthy}},
		deviceSpec:      []*pluginapi.DeviceSpec{{HostPath: "/dev/test", ContainerPath: "/dev/test", Permissions: "rw"}},
		devices:         devices,
		useCdi:          useCdi,
		cdiResourceName: config.ResourceName,
	}
}

func TestNewResourceServerInvalidRdmaHcaMax(t *testing.T) {
	convey.Convey("When rdmaHcaMax is negative", t, func() {
		config := &types.UserConfig{
			ResourceName:   "test",
			ResourcePrefix: "test.io",
			RdmaHcaMax:     -1,
		}

		rs, err := newResourceServer(config, nil, false, "sock", false)

		convey.Convey("Then error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "Invalid value for rdmaHcaMax < 0")
			convey.So(rs, convey.ShouldBeNil)
		})
	})
}

func TestNewResourceServerEmptyResourcePrefix(t *testing.T) {
	convey.Convey("When resourcePrefix is empty", t, func() {
		config := &types.UserConfig{
			ResourceName:   "test",
			ResourcePrefix: "",
			RdmaHcaMax:     100,
		}

		rs, err := newResourceServer(config, nil, false, "sock", false)

		convey.Convey("Then error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "Empty resourcePrefix")
			convey.So(rs, convey.ShouldBeNil)
		})
	})
}

func TestNewResourceServerSuccess(t *testing.T) {
	convey.Convey("When valid config is provided", t, func() {
		config := &types.UserConfig{
			ResourceName:   "test-resource",
			ResourcePrefix: "test.io",
			RdmaHcaMax:     5,
		}
		devices := []types.Device{&mockDevice{name: "test0", rdmaSpec: []*pluginapi.DeviceSpec{{HostPath: "/dev/test0", ContainerPath: "/dev/test0", Permissions: "rw"}}}}

		rs, err := newResourceServer(config, devices, false, "sock", false)

		convey.Convey("Then server should be created successfully", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(rs, convey.ShouldNotBeNil)
		})
	})
}

func TestNewResourceServerEmptyDevices(t *testing.T) {
	convey.Convey("When devices list is empty", t, func() {
		config := &types.UserConfig{
			ResourceName:   "test-resource",
			ResourcePrefix: "test.io",
			RdmaHcaMax:     5,
		}

		rs, err := newResourceServer(config, nil, false, "sock", false)

		convey.Convey("Then server should be created with empty devs", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(rs, convey.ShouldNotBeNil)
		})
	})
}

func TestNewResourceServerWatchMode(t *testing.T) {
	convey.Convey("When watch mode is enabled", t, func() {
		config := &types.UserConfig{
			ResourceName:   "test-resource",
			ResourcePrefix: "test.io",
			RdmaHcaMax:     3,
		}

		rs, err := newResourceServer(config, nil, true, "sock", false)

		convey.Convey("Then server should be created successfully", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(rs, convey.ShouldNotBeNil)
		})
	})
}

func TestDetectPluginWatchModeSockDirExists(t *testing.T) {
	convey.Convey("When socket directory exists", t, func() {
		tempDir, _ := os.MkdirTemp("", "test-watch-mode")
		defer os.RemoveAll(tempDir)

		result := detectPluginWatchMode(tempDir)

		convey.Convey("Then it should return true", func() {
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

func TestDetectPluginWatchModeSockDirNotExists(t *testing.T) {
	convey.Convey("When socket directory does not exist", t, func() {
		result := detectPluginWatchMode("/nonexistent/path/that/should/not/exist")

		convey.Convey("Then it should return false", func() {
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestResourceServerGetDevicePluginOptions(t *testing.T) {
	convey.Convey("When GetDevicePluginOptions is called", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)

		opts, err := rs.GetDevicePluginOptions(context.Background(), &pluginapi.Empty{})

		convey.Convey("Then PreStartRequired should be false", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(opts, convey.ShouldNotBeNil)
			convey.So(opts.PreStartRequired, convey.ShouldBeFalse)
		})
	})
}

func TestResourceServerPreStartContainer(t *testing.T) {
	convey.Convey("When PreStartContainer is called", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)

		resp, err := rs.PreStartContainer(context.Background(), &pluginapi.PreStartContainerRequest{})

		convey.Convey("Then empty response should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
		})
	})
}

func TestResourceServerGetPreferredAllocation(t *testing.T) {
	convey.Convey("When GetPreferredAllocation is called", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)

		resp, err := rs.GetPreferredAllocation(context.Background(), &pluginapi.PreferredAllocationRequest{})

		convey.Convey("Then nil response should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldBeNil)
		})
	})
}

func TestResourceServerGetInfo(t *testing.T) {
	convey.Convey("When GetInfo is called", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)

		info, err := rs.GetInfo(context.Background(), &registerapi.InfoRequest{})

		convey.Convey("Then plugin info should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(info, convey.ShouldNotBeNil)
			convey.So(info.Type, convey.ShouldEqual, registerapi.DevicePlugin)
			convey.So(info.Name, convey.ShouldEqual, rs.resourceName)
		})
	})
}

func TestResourceServerNotifyRegistrationStatusRegistered(t *testing.T) {
	convey.Convey("When plugin is registered successfully", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)
		rs.rsConnector = &resourcesServerPort{}

		resp, err := rs.NotifyRegistrationStatus(context.Background(), &registerapi.RegistrationStatus{PluginRegistered: true})

		convey.Convey("Then empty response should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
		})
	})
}

func TestResourceServerNotifyRegistrationStatusFailed(t *testing.T) {
	convey.Convey("When plugin registration fails", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)
		rsc := &resourcesServerPort{}
		rsc.CreateServer()
		rs.rsConnector = rsc

		resp, err := rs.NotifyRegistrationStatus(context.Background(), &registerapi.RegistrationStatus{PluginRegistered: false, Error: "test error"})

		convey.Convey("Then empty response should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
		})
	})
}

func TestResourceServerAllocateWithoutCdi(t *testing.T) {
	convey.Convey("When Allocate is called without CDI", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)
		rs.deviceSpec = []*pluginapi.DeviceSpec{{HostPath: "/dev/test0", ContainerPath: "/dev/test0", Permissions: "rw"}}

		req := &pluginapi.AllocateRequest{
			ContainerRequests: []*pluginapi.ContainerAllocateRequest{{}, {}},
		}

		resp, err := rs.Allocate(context.Background(), req)

		convey.Convey("Then device specs should be allocated to each container", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(resp.ContainerResponses), convey.ShouldEqual, 2)
			convey.So(len(resp.ContainerResponses[0].Devices), convey.ShouldEqual, 1)
		})
	})
}

func TestResourceServerUpdateDevicesChanged(t *testing.T) {
	convey.Convey("When devices have changed", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)
		rs.deviceSpec = []*pluginapi.DeviceSpec{{HostPath: "/dev/old", ContainerPath: "/dev/old", Permissions: "rw"}}

		patches := gomonkey.ApplyFunc(common.DevicesChanged,
			func(deviceList, newDeviceList []*pluginapi.DeviceSpec) bool {
				return true
			})
		defer patches.Reset()

		newDevices := []types.Device{&mockDevice{name: "new0", rdmaSpec: []*pluginapi.DeviceSpec{{HostPath: "/dev/new", ContainerPath: "/dev/new", Permissions: "rw"}}}}
		rs.UpdateDevices(newDevices)

		convey.Convey("Then deviceSpec should be updated", func() {
			convey.So(len(rs.deviceSpec), convey.ShouldEqual, 1)
			convey.So(rs.deviceSpec[0].HostPath, convey.ShouldEqual, "/dev/new")
			convey.So(len(rs.updateResource), convey.ShouldEqual, 1)
		})
	})
}

func TestResourceServerUpdateDevicesNotChanged(t *testing.T) {
	convey.Convey("When devices have not changed", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)
		rs.deviceSpec = []*pluginapi.DeviceSpec{{HostPath: "/dev/test0", ContainerPath: "/dev/test0", Permissions: "rw"}}

		patches := gomonkey.ApplyFunc(common.DevicesChanged,
			func(deviceList, newDeviceList []*pluginapi.DeviceSpec) bool {
				return false
			})
		defer patches.Reset()

		rs.UpdateDevices(nil)

		convey.Convey("Then deviceSpec should not be updated", func() {
			convey.So(len(rs.updateResource), convey.ShouldEqual, 0)
		})
	})
}

func TestResourceServerUpdateDevicesEmptyDevices(t *testing.T) {
	convey.Convey("When new devices list is empty", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)
		rs.deviceSpec = []*pluginapi.DeviceSpec{{HostPath: "/dev/test0", ContainerPath: "/dev/test0", Permissions: "rw"}}
		rs.devs = []*pluginapi.Device{{ID: "0", Health: pluginapi.Healthy}}

		patches := gomonkey.ApplyFunc(common.DevicesChanged,
			func(deviceList, newDeviceList []*pluginapi.DeviceSpec) bool {
				return true
			})
		patches.ApplyFunc(getDevicesSpec, func(devices []types.Device) []*pluginapi.DeviceSpec {
			return []*pluginapi.DeviceSpec{}
		})
		defer patches.Reset()

		rs.UpdateDevices(nil)

		convey.Convey("Then devs should be empty", func() {
			convey.So(len(rs.deviceSpec), convey.ShouldEqual, 0)
			convey.So(len(rs.devs), convey.ShouldEqual, 0)
		})
	})
}

func TestResourceServerCleanupSuccess(t *testing.T) {
	convey.Convey("When socket file exists", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)
		rs.socketPath = "/tmp/test-cleanup.sock"
		os.Remove(rs.socketPath)
		_, _ = os.Create(rs.socketPath)

		err := rs.cleanup()

		convey.Convey("Then no error should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
			_, statErr := os.Stat(rs.socketPath)
			convey.So(os.IsNotExist(statErr), convey.ShouldBeTrue)
		})
	})
}

func TestResourceServerCleanupFileNotExist(t *testing.T) {
	convey.Convey("When socket file does not exist", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)
		rs.socketPath = "/tmp/nonexistent-file-12345.sock"

		err := rs.cleanup()

		convey.Convey("Then no error should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestResourceServerCleanupError(t *testing.T) {
	convey.Convey("When os.Remove fails", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)

		patches := gomonkey.ApplyFunc(os.Remove, func(name string) error {
			return errors.New("permission denied")
		})
		defer patches.Reset()

		err := rs.cleanup()

		convey.Convey("Then error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestResourceServerStopNilConnector(t *testing.T) {
	convey.Convey("When rsConnector is nil", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)
		rs.rsConnector = nil

		err := rs.Stop()

		convey.Convey("Then no error should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestResourceServerRestartNilConnector(t *testing.T) {
	convey.Convey("When rsConnector is nil during Restart", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)
		rs.rsConnector = nil

		err := rs.Restart()

		convey.Convey("Then error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "grpc server instance not found")
		})
	})
}

func TestGetDevicesSpecWithDevices(t *testing.T) {
	convey.Convey("When devices have RDMA spec", t, func() {
		devices := []types.Device{
			&mockDevice{name: "dev0", rdmaSpec: []*pluginapi.DeviceSpec{{HostPath: "/dev/dev0", ContainerPath: "/dev/dev0", Permissions: "rw"}}},
			&mockDevice{name: "dev1", rdmaSpec: []*pluginapi.DeviceSpec{{HostPath: "/dev/dev1", ContainerPath: "/dev/dev1", Permissions: "rw"}}},
		}

		result := getDevicesSpec(devices)

		convey.Convey("Then combined specs should be returned", func() {
			convey.So(len(result), convey.ShouldEqual, 2)
		})
	})
}

func TestGetDevicesSpecEmptyDevices(t *testing.T) {
	convey.Convey("When devices list is empty", t, func() {
		result := getDevicesSpec(nil)

		convey.Convey("Then empty spec should be returned", func() {
			convey.So(result, convey.ShouldBeEmpty)
		})
	})
}

func TestGetDevicesSpecDeviceWithEmptySpec(t *testing.T) {
	convey.Convey("When device has empty RDMA spec", t, func() {
		devices := []types.Device{
			&mockDevice{name: "dev0", rdmaSpec: []*pluginapi.DeviceSpec{}},
		}

		result := getDevicesSpec(devices)

		convey.Convey("Then empty spec should be returned", func() {
			convey.So(result, convey.ShouldBeEmpty)
		})
	})
}

func TestResourcesServerPortGetServer(t *testing.T) {
	convey.Convey("When server is nil", t, func() {
		rsc := &resourcesServerPort{}

		result := rsc.GetServer()

		convey.Convey("Then nil should be returned", func() {
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

func TestResourcesServerPortCreateServer(t *testing.T) {
	convey.Convey("When CreateServer is called", t, func() {
		rsc := &resourcesServerPort{}

		rsc.CreateServer()

		convey.Convey("Then server should be created", func() {
			convey.So(rsc.server, convey.ShouldNotBeNil)
		})
	})
}

func TestResourcesServerPortDeleteServer(t *testing.T) {
	convey.Convey("When DeleteServer is called", t, func() {
		rsc := &resourcesServerPort{}
		rsc.CreateServer()

		rsc.DeleteServer()

		convey.Convey("Then server should be nil", func() {
			convey.So(rsc.server, convey.ShouldBeNil)
		})
	})
}

func TestResourcesServerPortStopNilServer(t *testing.T) {
	convey.Convey("When server is nil during Stop", t, func() {
		rsc := &resourcesServerPort{}
		rsc.CreateServer()

		rsc.Stop()

		convey.Convey("Then it should not panic", func() {
			convey.So(true, convey.ShouldBeTrue)
		})
	})
}

func TestResourcesServerPortCloseNilConn(t *testing.T) {
	convey.Convey("When conn is nil during Close", t, func() {
		rsc := &resourcesServerPort{}
		conn, err := grpc.NewClient("unix:///tmp/test-close.sock",
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Skip("cannot create grpc client connection")
		}

		rsc.Close(conn)

		convey.Convey("Then it should not panic", func() {
			convey.So(true, convey.ShouldBeTrue)
		})
	})
}

func TestUpdateCDISpecWithoutCdi(t *testing.T) {
	convey.Convey("When CDI is disabled", t, func() {
		config := &types.UserConfig{ResourceName: "test", ResourcePrefix: "test.io", RdmaHcaMax: 10}
		rs := newTestResourceServer(config, nil, false, false)

		err := rs.updateCDISpec()

		convey.Convey("Then no error should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
