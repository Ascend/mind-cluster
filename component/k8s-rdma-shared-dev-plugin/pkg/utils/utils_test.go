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

package utils

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/Mellanox/rdmamap"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
)

// ubDirEntry is a fake dir entry whose IsDir reports true.
type ubDirEntry struct{ name string }

func (f ubDirEntry) Name() string               { return f.name }
func (f ubDirEntry) IsDir() bool                { return true }
func (f ubDirEntry) Type() fs.FileMode          { return fs.ModeDir }
func (f ubDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

// TestGetPciAddress tests GetPciAddress for valid symlinks, missing dirs and non-symlink devices.
func TestGetPciAddress(t *testing.T) {
	convey.Convey("Given a net device whose device dir is a kernel-style symlink", t, func() {
		netDir := filepath.Join(t.TempDir(), "net")
		convey.So(os.MkdirAll(filepath.Join(netDir, "ib0"), 0700), convey.ShouldBeNil)
		reset := gomonkey.ApplyGlobalVar(&sysNetDevices, netDir)
		defer reset.Reset()
		convey.So(os.Symlink("../../../0000:03:00.0", filepath.Join(netDir, "ib0", "device")), convey.ShouldBeNil)

		pciAddr, err := GetPciAddress("ib0")
		convey.Convey("Then the pci address is parsed from the link tail", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(pciAddr, convey.ShouldEqual, "0000:03:00.0")
		})
	})
	convey.Convey("Given a net device without a device dir", t, func() {
		netDir := filepath.Join(t.TempDir(), "net")
		convey.So(os.MkdirAll(filepath.Join(netDir, "ib0"), 0700), convey.ShouldBeNil)
		reset := gomonkey.ApplyGlobalVar(&sysNetDevices, netDir)
		defer reset.Reset()

		_, err := GetPciAddress("ib0")
		convey.Convey("Then an error is returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
	convey.Convey("Given a net device whose device dir is not a symlink", t, func() {
		netDir := filepath.Join(t.TempDir(), "net")
		convey.So(os.MkdirAll(filepath.Join(netDir, "ib0", "device"), 0700), convey.ShouldBeNil)
		reset := gomonkey.ApplyGlobalVar(&sysNetDevices, netDir)
		defer reset.Reset()

		_, err := GetPciAddress("ib0")
		convey.Convey("Then an error is returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestGetRdmaDevices tests GetRdmaDevices aggregating char devices from rdmamap.
func TestGetRdmaDevices(t *testing.T) {
	convey.Convey("Given rdmamap returns resources and char devices", t, func() {
		patches := gomonkey.ApplyFunc(rdmamap.GetRdmaDevicesForPcidev, func(string) []string {
			return []string{"uverbs0", "rdma_cm"}
		}).ApplyFunc(rdmamap.GetRdmaCharDevices, func(resource string) []string {
			return []string{resource}
		})
		defer patches.Reset()

		devices := GetRdmaDevices("0000:03:00.0")
		convey.Convey("Then all char devices are aggregated", func() {
			convey.So(devices, convey.ShouldResemble, []string{"uverbs0", "rdma_cm"})
		})
	})
	convey.Convey("Given rdmamap returns no resources", t, func() {
		patches := gomonkey.ApplyFuncReturn(rdmamap.GetRdmaDevicesForPcidev, []string(nil))
		defer patches.Reset()

		devices := GetRdmaDevices("0000:03:00.0")
		convey.Convey("Then an empty slice is returned", func() {
			convey.So(devices, convey.ShouldNotBeNil)
			convey.So(devices, convey.ShouldBeEmpty)
		})
	})
}

// TestGetRdmaDevicesForUbdev tests GetRdmaDevicesForUbdev reading the UB infiniband dir via mocked sysfs.
func TestGetRdmaDevicesForUbdev(t *testing.T) {
	convey.Convey("Given a UB device whose infiniband dir contains sub-dirs", t, func() {
		patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
			return []os.DirEntry{ubDirEntry{name: "ub0_0"}, ubDirEntry{name: "ub0_1"}}, nil
		}).ApplyFunc(rdmamap.GetRdmaCharDevices, func(dev string) []string {
			return []string{dev + "u"}
		})
		defer patches.Reset()

		devices := GetRdmaDevicesForUbdev("ub0")
		convey.Convey("Then char devices of sub-dirs are aggregated", func() {
			convey.So(devices, convey.ShouldResemble, []string{"ub0_0u", "ub0_1u"})
		})
	})
	convey.Convey("Given the infiniband dir does not exist", t, func() {
		patches := gomonkey.ApplyFuncReturn(os.ReadDir, []os.DirEntry(nil), errors.New("no such file"))
		defer patches.Reset()

		devices := GetRdmaDevicesForUbdev("ub0")
		convey.Convey("Then nil is returned", func() {
			convey.So(devices, convey.ShouldBeNil)
		})
	})
	convey.Convey("Given the infiniband dir has no sub-dirs", t, func() {
		patches := gomonkey.ApplyFuncReturn(os.ReadDir, []os.DirEntry{}, nil)
		defer patches.Reset()

		devices := GetRdmaDevicesForUbdev("ub0")
		convey.Convey("Then an empty slice is returned", func() {
			convey.So(devices, convey.ShouldNotBeNil)
			convey.So(devices, convey.ShouldBeEmpty)
		})
	})
}

// TestIsEmptySelector tests IsEmptySelector for nil, partially and fully set selectors.
func TestIsEmptySelector(t *testing.T) {
	convey.Convey("Given an all-empty selector", t, func() {
		convey.So(IsEmptySelector(&types.Selectors{}), convey.ShouldBeTrue)
	})
	convey.Convey("Given a selector with a set field", t, func() {
		convey.So(IsEmptySelector(&types.Selectors{Vendors: []string{"15b3"}}), convey.ShouldBeFalse)
	})
	convey.Convey("Given a selector with an empty slice field", t, func() {
		convey.So(IsEmptySelector(&types.Selectors{Vendors: []string{}}), convey.ShouldBeTrue)
	})
}

// TestGetNetNames tests GetNetNames for valid dirs and missing dirs.
func TestGetNetNames(t *testing.T) {
	convey.Convey("Given a pci device with a net dir", t, func() {
		reset := redirectSysBusPciForTest(t)
		defer reset.Reset()

		names, err := GetNetNames("0000:03:00.0")
		convey.Convey("Then all net interface names are returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(names, convey.ShouldResemble, []string{"eth0", "eth1"})
		})
	})
	convey.Convey("Given a pci device without a net dir", t, func() {
		reset := redirectSysBusPciForTest(t)
		defer reset.Reset()

		_, err := GetNetNames("0000:09:00.0")
		convey.Convey("Then an error is returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestGetPCIDevDriver tests GetPCIDevDriver for valid links and missing drivers.
func TestGetPCIDevDriver(t *testing.T) {
	convey.Convey("Given a pci device bound to a driver", t, func() {
		pciDir := filepath.Join(t.TempDir(), "devices")
		convey.So(os.MkdirAll(filepath.Join(pciDir, "0000:03:00.0"), 0700), convey.ShouldBeNil)
		reset := gomonkey.ApplyGlobalVar(&SysBusPci, pciDir)
		defer reset.Reset()
		convey.So(os.Symlink("../../../../drivers/hns3", filepath.Join(pciDir, "0000:03:00.0", "driver")), convey.ShouldBeNil)

		driver, err := GetPCIDevDriver("0000:03:00.0")
		convey.Convey("Then the driver name is the link base", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(driver, convey.ShouldEqual, "hns3")
		})
	})
	convey.Convey("Given a pci device without a driver link", t, func() {
		pciDir := filepath.Join(t.TempDir(), "devices")
		convey.So(os.MkdirAll(filepath.Join(pciDir, "0000:03:00.0"), 0700), convey.ShouldBeNil)
		reset := gomonkey.ApplyGlobalVar(&SysBusPci, pciDir)
		defer reset.Reset()

		_, err := GetPCIDevDriver("0000:03:00.0")
		convey.Convey("Then an error is returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestGetNodeName tests GetNodeName for set and empty env values.
func TestGetNodeName(t *testing.T) {
	convey.Convey("Given NODE_NAME is set", t, func() {
		t.Setenv("NODE_NAME", "node-1")
		name, err := GetNodeName()
		convey.Convey("Then the node name is returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(name, convey.ShouldEqual, "node-1")
		})
	})
	convey.Convey("Given NODE_NAME is empty", t, func() {
		t.Setenv("NODE_NAME", "")
		_, err := GetNodeName()
		convey.Convey("Then an error is returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// redirectSysBusPciForTest points SysBusPci to a temp dir containing one device with a net dir.
func redirectSysBusPciForTest(t *testing.T) *gomonkey.Patches {
	pciDir := filepath.Join(t.TempDir(), "devices")
	for _, dir := range []string{
		filepath.Join(pciDir, "0000:03:00.0", "net", "eth0"),
		filepath.Join(pciDir, "0000:03:00.0", "net", "eth1"),
	} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	return gomonkey.ApplyGlobalVar(&SysBusPci, pciDir)
}
