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

// Package common for common info
package common

import (
	"testing"

	c "github.com/smartystreets/goconvey/convey"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func TestDevicesChanged_DifferentLength(t *testing.T) {
	deviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
	}
	newDeviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
		{HostPath: "/dev/ub1", ContainerPath: "/dev/ub1", Permissions: "rw"},
	}

	c.Convey("When device lists have different lengths", t, func() {
		result := DevicesChanged(deviceList, newDeviceList)
		c.Convey("Then it should return true", func() {
			c.So(result, c.ShouldBeTrue)
		})
	})
}

func TestDevicesChanged_SameDevices(t *testing.T) {
	deviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
		{HostPath: "/dev/ub1", ContainerPath: "/dev/ub1", Permissions: "rw"},
	}
	newDeviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
		{HostPath: "/dev/ub1", ContainerPath: "/dev/ub1", Permissions: "rw"},
	}

	c.Convey("When device lists are identical", t, func() {
		result := DevicesChanged(deviceList, newDeviceList)
		c.Convey("Then it should return false", func() {
			c.So(result, c.ShouldBeFalse)
		})
	})
}

func TestDevicesChanged_DifferentOrder(t *testing.T) {
	deviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
		{HostPath: "/dev/ub1", ContainerPath: "/dev/ub1", Permissions: "rw"},
	}
	newDeviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub1", ContainerPath: "/dev/ub1", Permissions: "rw"},
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
	}

	c.Convey("When device lists have different order but same content", t, func() {
		result := DevicesChanged(deviceList, newDeviceList)
		c.Convey("Then it should return false", func() {
			c.So(result, c.ShouldBeFalse)
		})
	})
}

func TestDevicesChanged_NewDeviceAdded(t *testing.T) {
	deviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
	}
	newDeviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
		{HostPath: "/dev/ub2", ContainerPath: "/dev/ub2", Permissions: "rw"},
	}

	c.Convey("When a new device is added", t, func() {
		result := DevicesChanged(deviceList, newDeviceList)
		c.Convey("Then it should return true", func() {
			c.So(result, c.ShouldBeTrue)
		})
	})
}

func TestDevicesChanged_DeviceRemoved(t *testing.T) {
	deviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
		{HostPath: "/dev/ub1", ContainerPath: "/dev/ub1", Permissions: "rw"},
	}
	newDeviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
	}

	c.Convey("When a device is removed", t, func() {
		result := DevicesChanged(deviceList, newDeviceList)
		c.Convey("Then it should return true", func() {
			c.So(result, c.ShouldBeTrue)
		})
	})
}

func TestDevicesChanged_EmptyBoth(t *testing.T) {
	c.Convey("When both device lists are empty", t, func() {
		result := DevicesChanged([]*pluginapi.DeviceSpec{}, []*pluginapi.DeviceSpec{})
		c.Convey("Then it should return false", func() {
			c.So(result, c.ShouldBeFalse)
		})
	})
}

func TestDevicesChanged_EmptyFirst(t *testing.T) {
	newDeviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
	}

	c.Convey("When first list is empty and second has devices", t, func() {
		result := DevicesChanged([]*pluginapi.DeviceSpec{}, newDeviceList)
		c.Convey("Then it should return true", func() {
			c.So(result, c.ShouldBeTrue)
		})
	})
}

func TestDevicesChanged_EmptySecond(t *testing.T) {
	deviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
	}

	c.Convey("When second list is empty and first has devices", t, func() {
		result := DevicesChanged(deviceList, []*pluginapi.DeviceSpec{})
		c.Convey("Then it should return true", func() {
			c.So(result, c.ShouldBeTrue)
		})
	})
}

func TestDevicesChanged_DifferentPermissions(t *testing.T) {
	deviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
	}
	newDeviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "r"},
	}

	c.Convey("When permissions differ but paths are same", t, func() {
		result := DevicesChanged(deviceList, newDeviceList)
		c.Convey("Then it should return false", func() {
			c.So(result, c.ShouldBeFalse)
		})
	})
}

func TestDevicesChanged_DifferentContainerPath(t *testing.T) {
	deviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0", Permissions: "rw"},
	}
	newDeviceList := []*pluginapi.DeviceSpec{
		{HostPath: "/dev/ub0", ContainerPath: "/dev/ub0-custom", Permissions: "rw"},
	}

	c.Convey("When container paths differ but host paths are same", t, func() {
		result := DevicesChanged(deviceList, newDeviceList)
		c.Convey("Then it should return false", func() {
			c.So(result, c.ShouldBeFalse)
		})
	})
}
