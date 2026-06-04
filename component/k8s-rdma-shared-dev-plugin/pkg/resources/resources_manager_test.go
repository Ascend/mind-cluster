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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/jaypipes/ghw"
	"github.com/smartystreets/goconvey/convey"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
)

func TestNewResourceManager(t *testing.T) {
	convey.Convey("When NewResourceManager is called", t, func() {
		rm := NewResourceManager("/tmp/test.json", false)

		convey.Convey("Then it should return a valid resourceManager", func() {
			convey.So(rm, convey.ShouldNotBeNil)

			// Verify type assertion works
			resourceMgr, ok := rm.(*resourceManager)
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(resourceMgr.deviceList, convey.ShouldBeEmpty)
			convey.So(resourceMgr.netlinkManager, convey.ShouldNotBeNil)
			convey.So(resourceMgr.rds, convey.ShouldNotBeNil)
		})
	})
}

func TestNewResourceManagerWithCDI(t *testing.T) {
	convey.Convey("When NewResourceManager is called with useCdi=true", t, func() {
		rm := NewResourceManager("/tmp/test.json", true)

		convey.Convey("Then it should return a valid resourceManager with CDI enabled", func() {
			convey.So(rm, convey.ShouldNotBeNil)

			resourceMgr, ok := rm.(*resourceManager)
			convey.So(ok, convey.ShouldBeTrue)

			// Verify CDI is enabled in core manager
			convey.So(resourceMgr.GetUseCdi(), convey.ShouldBeTrue)
		})
	})
}

func TestValidResourceName_Success(t *testing.T) {
	convey.Convey("When validResourceName is called with valid names", t, func() {
		testCases := []string{
			"rdma",
			"rdma_dev",
			"RDMA",
			"rdma123",
			"RDMA_DEVICE_1",
		}

		for _, name := range testCases {
			result := validResourceName(name)
			convey.Convey("Then it should return true for "+name, func() {
				convey.So(result, convey.ShouldBeTrue)
			})
		}
	})
}

func TestValidResourceName_Invalid(t *testing.T) {
	convey.Convey("When validResourceName is called with invalid names", t, func() {
		testCases := []string{
			"rdma-dev",
			"rdma dev",
			"rdma@dev",
			"rdma.dev",
			"rdma/dev",
			"rdma\\dev",
			"",
		}

		for _, name := range testCases {
			result := validResourceName(name)
			convey.Convey("Then it should return false for '"+name+"'", func() {
				convey.So(result, convey.ShouldBeFalse)
			})
		}
	})
}

func TestResourceManagerPeriodicUpdateZeroInterval(t *testing.T) {
	convey.Convey("When PeriodicUpdate is called with zero interval", t, func() {
		rm := NewResourceManager("/tmp/test.json", false)

		stopFn := rm.PeriodicUpdate()

		convey.Convey("Then stop function should not block and return nil", func() {
			convey.So(stopFn, convey.ShouldNotBeNil)
			// Calling stop should not panic
			stopFn()
		})
	})
}

func TestGetDevices_EmptyList(t *testing.T) {
	convey.Convey("When GetDevices is called on empty deviceList", t, func() {
		rm := NewResourceManager("/tmp/test.json", false)
		resourceMgr, _ := rm.(*resourceManager)

		devices := resourceMgr.GetDevices()

		convey.Convey("Then it should return empty slice", func() {
			convey.So(devices, convey.ShouldBeEmpty)
		})
	})
}

func TestGetFilteredDevices_EmptyInput(t *testing.T) {
	convey.Convey("When GetFilteredDevices is called with empty devices", t, func() {
		rm := NewResourceManager("/tmp/test.json", false)
		resourceMgr, _ := rm.(*resourceManager)

		selector := &types.Selectors{}
		result := resourceMgr.GetFilteredDevices(nil, selector)

		convey.Convey("Then it should return empty slice", func() {
			convey.So(result, convey.ShouldBeEmpty)
		})
	})
}

func TestDiscoverHostDevices_EmptyDeviceList(t *testing.T) {
	convey.Convey("When DiscoverHostDevices is called and no devices found", t, func() {
		rm := NewResourceManager("/tmp/test.json", false)
		resourceMgr, _ := rm.(*resourceManager)

		patches := gomonkey.ApplyFunc(ghw.PCI, func() (*ghw.PCIInfo, error) {
			return &ghw.PCIInfo{Devices: []*ghw.PCIDevice{}}, nil
		})
		defer patches.Reset()

		err := resourceMgr.DiscoverHostDevices()

		convey.Convey("Then it should not return error", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
