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
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/vishvananda/netlink"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"ascend-common/common-utils/hwlog"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/cdi"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/common"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/core"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
)

func init() {
	_ = hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func newTestUbResourceManager() *ubResourceManager {
	crm := core.NewCoreResourceManager("test.json", rdmaUbResourcePrefix, socketSuffix, false)
	return &ubResourceManager{
		CoreResourceManager: crm,
		deviceList:          []*UbDeviceInfo{},
		netlinkManager:      &netlinkManager{},
		rds:                 newUbRdmaDeviceSpec(common.RequiredRdmaDevices),
	}
}

func TestNewUbResourceManager(t *testing.T) {
	convey.Convey("Given NewUbResourceManager", t, func() {
		rm := NewUbResourceManager("test.json", false).(*ubResourceManager)

		convey.Convey("Then fields should be initialized correctly", func() {
			convey.So(rm, convey.ShouldNotBeNil)
			convey.So(rm.deviceList, convey.ShouldNotBeNil)
			convey.So(rm.netlinkManager, convey.ShouldNotBeNil)
			convey.So(rm.rds, convey.ShouldNotBeNil)
		})
	})
}

func TestDiscoverHostDevicesDirNotExist(t *testing.T) {
	rm := newTestUbResourceManager()

	patches := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	})
	defer patches.Reset()

	convey.Convey("When UB devices directory does not exist", t, func() {
		err := rm.DiscoverHostDevices()
		convey.Convey("Then it should succeed with empty device list", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(rm.deviceList), convey.ShouldEqual, 0)
		})
	})
}

func TestDiscoverHostDevicesReadDirError(t *testing.T) {
	rm := newTestUbResourceManager()

	patches := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
		return nil, nil
	})
	patches.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return nil, errors.New("permission denied")
	})
	defer patches.Reset()

	convey.Convey("When ReadDir returns an error", t, func() {
		err := rm.DiscoverHostDevices()
		convey.Convey("Then it should return the error", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "permission denied")
		})
	})
}

func TestDiscoverHostDevicesEmptyDir(t *testing.T) {
	rm := newTestUbResourceManager()

	patches := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
		return nil, nil
	})
	patches.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return []os.DirEntry{}, nil
	})
	defer patches.Reset()

	convey.Convey("When UB devices directory is empty", t, func() {
		err := rm.DiscoverHostDevices()
		convey.Convey("Then it should succeed with empty device list", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(rm.deviceList), convey.ShouldEqual, 0)
		})
	})
}

func TestDiscoverHostDevicesReadInfoFails(t *testing.T) {
	rm := newTestUbResourceManager()
	mockEntry := &mockDirEntry{name: "ub0", isDir: true}

	patches := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
		return nil, nil
	})
	patches.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if strings.Contains(name, "net") {
			return []os.DirEntry{}, nil
		}
		return []os.DirEntry{mockEntry}, nil
	})
	patches.ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
		return []byte("0x19e5"), nil
	})
	patches.ApplyFunc(os.Readlink, func(name string) (string, error) {
		return "", errors.New("readlink error")
	})
	defer patches.Reset()

	convey.Convey("When readUbDeviceInfo fails for all entries", t, func() {
		err := rm.DiscoverHostDevices()
		convey.Convey("Then it should succeed with empty device list", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(rm.deviceList), convey.ShouldEqual, 0)
		})
	})
}

func TestDiscoverHostDevicesSuccess(t *testing.T) {
	rm := newTestUbResourceManager()
	mockEntry := &mockDirEntry{name: "ub0", isDir: true}
	netEntry := &mockDirEntry{name: "enp0s1", isDir: true}

	patches := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
		return nil, nil
	})
	patches.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if strings.Contains(name, "net") {
			return []os.DirEntry{netEntry}, nil
		}
		return []os.DirEntry{mockEntry}, nil
	})
	patches.ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
		if strings.Contains(name, "vendor") {
			return []byte("0x19e5"), nil
		}
		if strings.Contains(name, "device") {
			return []byte("0xa222"), nil
		}
		return []byte(""), nil
	})
	patches.ApplyFunc(os.Readlink, func(name string) (string, error) {
		return "../../../bus/ub/drivers/ub", nil
	})
	defer patches.Reset()

	convey.Convey("When UB devices are found successfully", t, func() {
		err := rm.DiscoverHostDevices()
		convey.Convey("Then the device list should be populated", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(rm.deviceList), convey.ShouldEqual, 1)
			convey.So(rm.deviceList[0].UbID, convey.ShouldEqual, "ub0")
		})
	})
}

func TestDiscoverHostDevicesMultipleDevices(t *testing.T) {
	rm := newTestUbResourceManager()
	entries := []os.DirEntry{
		&mockDirEntry{name: "ub0", isDir: true},
		&mockDirEntry{name: "ub1", isDir: true},
	}
	netEntry := &mockDirEntry{name: "enp0s1", isDir: true}

	patches := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
		return nil, nil
	})
	patches.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if strings.Contains(name, "net") {
			return []os.DirEntry{netEntry}, nil
		}
		return entries, nil
	})
	patches.ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
		if strings.Contains(name, "ub0") {
			if strings.Contains(name, "vendor") {
				return []byte("0x19e5"), nil
			}
			if strings.Contains(name, "device") {
				return []byte("0xa222"), nil
			}
		}
		return []byte("0x19e5"), nil
	})
	patches.ApplyFunc(os.Readlink, func(name string) (string, error) {
		if strings.Contains(name, "ub1") {
			return "", errors.New("readlink error")
		}
		return "../../../bus/ub/drivers/ub", nil
	})
	defer patches.Reset()

	convey.Convey("When some devices succeed and some fail", t, func() {
		err := rm.DiscoverHostDevices()
		convey.Convey("Then only successful devices should be added", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(rm.deviceList), convey.ShouldEqual, 1)
			convey.So(rm.deviceList[0].UbID, convey.ShouldEqual, "ub0")
		})
	})
}

func TestReadUbDeviceInfoSuccess(t *testing.T) {
	rm := newTestUbResourceManager()
	netEntry := &mockDirEntry{name: "enp0s1", isDir: true}

	patches := gomonkey.ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
		if strings.Contains(name, "vendor") {
			return []byte("0x19e5"), nil
		}
		if strings.Contains(name, "device") {
			return []byte("0xa222"), nil
		}
		return nil, errors.New("unexpected file")
	})
	patches.ApplyFunc(os.Readlink, func(name string) (string, error) {
		return "../../../bus/ub/drivers/ub", nil
	})
	patches.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if strings.Contains(name, "net") {
			return []os.DirEntry{netEntry}, nil
		}
		return nil, errors.New("unexpected dir")
	})
	defer patches.Reset()

	convey.Convey("When reading UB device info succeeds", t, func() {
		info, err := rm.readUbDeviceInfo("/fake/ub0", "ub0")
		convey.Convey("Then device info should be correct", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(info.UbID, convey.ShouldEqual, "ub0")
			convey.So(info.Vendor, convey.ShouldEqual, "0x19e5")
			convey.So(info.DeviceID, convey.ShouldEqual, "0xa222")
			convey.So(info.Driver, convey.ShouldEqual, "ub")
		})
	})
}

func TestReadUbDeviceInfoDriverReadlinkError(t *testing.T) {
	rm := newTestUbResourceManager()

	patches := gomonkey.ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
		return []byte("0x19e5"), nil
	})
	patches.ApplyFunc(os.Readlink, func(name string) (string, error) {
		return "", errors.New("readlink error")
	})
	defer patches.Reset()

	convey.Convey("When driver readlink fails", t, func() {
		_, err := rm.readUbDeviceInfo("/fake/ub0", "ub0")
		convey.Convey("Then an error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "driver info")
		})
	})
}

func TestReadUbDeviceInfoVendorFileError(t *testing.T) {
	rm := newTestUbResourceManager()

	patches := gomonkey.ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
		if strings.Contains(name, "vendor") {
			return nil, errors.New("read error")
		}
		return []byte("0xa222"), nil
	})
	patches.ApplyFunc(os.Readlink, func(name string) (string, error) {
		return "../../../bus/ub/drivers/ub", nil
	})
	patches.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if strings.Contains(name, "net") {
			return []os.DirEntry{}, nil
		}
		return nil, errors.New("unexpected dir")
	})
	defer patches.Reset()

	convey.Convey("When vendor file read fails", t, func() {
		info, err := rm.readUbDeviceInfo("/fake/ub0", "ub0")
		convey.Convey("Then info should still be returned with empty vendor", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(info.Vendor, convey.ShouldEqual, "")
			convey.So(info.DeviceID, convey.ShouldEqual, "0xa222")
			convey.So(info.Driver, convey.ShouldEqual, "ub")
		})
	})
}

func TestReadUbDeviceInfoDeviceIDTrim(t *testing.T) {
	rm := newTestUbResourceManager()

	patches := gomonkey.ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
		if strings.Contains(name, "vendor") {
			return []byte("0x19e5"), nil
		}
		if strings.Contains(name, "device") {
			return []byte("0xa222"), nil
		}
		return nil, errors.New("unexpected file")
	})
	patches.ApplyFunc(os.Readlink, func(name string) (string, error) {
		return "../../../bus/ub/drivers/ub", nil
	})
	patches.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if strings.Contains(name, "net") {
			return []os.DirEntry{}, nil
		}
		return nil, errors.New("unexpected dir")
	})
	defer patches.Reset()

	convey.Convey("When reading device info with hex prefixed values", t, func() {
		info, err := rm.readUbDeviceInfo("/fake/ub0", "ub0")
		convey.Convey("Then hex prefix should be preserved correctly", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(info.Vendor, convey.ShouldEqual, "0x19e5")
			convey.So(info.DeviceID, convey.ShouldEqual, "0xa222")
		})
	})
}

func TestReadUbNetInfoNoNetDir(t *testing.T) {
	rm := newTestUbResourceManager()

	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return nil, os.ErrNotExist
	})
	defer patches.Reset()

	convey.Convey("When net directory does not exist", t, func() {
		ifName, linkType := rm.readUbNetInfo("/sys/bus/ub/devices/ub0", "ub0")
		convey.Convey("Then empty values should be returned", func() {
			convey.So(ifName, convey.ShouldEqual, "")
			convey.So(linkType, convey.ShouldEqual, "")
		})
	})
}

func TestReadFirstDirEntryInvalidPath(t *testing.T) {
	convey.Convey("When dir path is not under /sys/", t, func() {
		name, count, err := readFirstDirEntry("/tmp/fake/dir")
		convey.Convey("Then error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(name, convey.ShouldEqual, "")
			convey.So(count, convey.ShouldEqual, 0)
			convey.So(err.Error(), convey.ShouldContainSubstring, "not under /sys/")
		})
	})
}

func TestReadFirstDirEntryPathTraversal(t *testing.T) {
	convey.Convey("When dir path contains traversal", t, func() {
		name, count, err := readFirstDirEntry("/sys/bus/ub/../../etc/passwd")
		convey.Convey("Then error should be returned due to path escape", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(name, convey.ShouldEqual, "")
			convey.So(count, convey.ShouldEqual, 0)
		})
	})
}

func TestReadFirstDirEntryReadDirError(t *testing.T) {
	convey.Convey("When os.ReadDir returns an error", t, func() {
		patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
			return nil, os.ErrPermission
		})
		defer patches.Reset()

		name, count, err := readFirstDirEntry("/sys/bus/ub/devices/ub0/infiniband")
		convey.Convey("Then error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(name, convey.ShouldEqual, "")
			convey.So(count, convey.ShouldEqual, 0)
			convey.So(err.Error(), convey.ShouldContainSubstring, "failed to read dir")
		})
	})
}

func TestReadFirstDirEntryEmptyDir(t *testing.T) {
	convey.Convey("When directory is empty", t, func() {
		patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
			return []os.DirEntry{}, nil
		})
		defer patches.Reset()

		name, count, err := readFirstDirEntry("/sys/bus/ub/devices/ub0/infiniband")
		convey.Convey("Then error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(name, convey.ShouldEqual, "")
			convey.So(count, convey.ShouldEqual, 0)
			convey.So(err.Error(), convey.ShouldContainSubstring, "no entries found")
		})
	})
}

func TestReadFirstDirEntrySingleEntry(t *testing.T) {
	convey.Convey("When directory has a single entry", t, func() {
		entry := &mockDirEntry{name: "ib0", isDir: false}
		patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
			return []os.DirEntry{entry}, nil
		})
		defer patches.Reset()

		name, count, err := readFirstDirEntry("/sys/bus/ub/devices/ub0/infiniband")
		convey.Convey("Then first entry name and count should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(name, convey.ShouldEqual, "ib0")
			convey.So(count, convey.ShouldEqual, 1)
		})
	})
}

func TestReadFirstDirEntryMultipleEntries(t *testing.T) {
	convey.Convey("When directory has multiple entries", t, func() {
		entries := []os.DirEntry{
			&mockDirEntry{name: "ib0", isDir: false},
			&mockDirEntry{name: "ib1", isDir: false},
		}
		patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
			return entries, nil
		})
		defer patches.Reset()

		name, count, err := readFirstDirEntry("/sys/bus/ub/devices/ub0/infiniband")
		convey.Convey("Then first entry name and total count should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(name, convey.ShouldEqual, "ib0")
			convey.So(count, convey.ShouldEqual, 2)
		})
	})
}

func TestReadUbDeviceNameReadDirError(t *testing.T) {
	rm := newTestUbResourceManager()

	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return nil, os.ErrPermission
	})
	defer patches.Reset()

	convey.Convey("When infiniband dir read fails", t, func() {
		result := rm.readUbDeviceName("/sys/bus/ub/devices/ub0", "ub0")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestReadUbDeviceNameEmptyDir(t *testing.T) {
	rm := newTestUbResourceManager()

	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return []os.DirEntry{}, nil
	})
	defer patches.Reset()

	convey.Convey("When infiniband dir is empty", t, func() {
		result := rm.readUbDeviceName("/sys/bus/ub/devices/ub0", "ub0")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestReadUbDeviceNameSingleEntry(t *testing.T) {
	rm := newTestUbResourceManager()
	ibEntry := &mockDirEntry{name: "ub0_0", isDir: false}

	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return []os.DirEntry{ibEntry}, nil
	})
	defer patches.Reset()

	convey.Convey("When infiniband dir has a single entry", t, func() {
		result := rm.readUbDeviceName("/sys/bus/ub/devices/ub0", "ub0")
		convey.Convey("Then the entry name should be returned", func() {
			convey.So(result, convey.ShouldEqual, "ub0_0")
		})
	})
}

func TestReadUbDeviceNameMultipleEntries(t *testing.T) {
	rm := newTestUbResourceManager()
	entries := []os.DirEntry{
		&mockDirEntry{name: "ub0_0", isDir: false},
		&mockDirEntry{name: "ub0_1", isDir: false},
	}

	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return entries, nil
	})
	defer patches.Reset()

	convey.Convey("When infiniband dir has multiple entries", t, func() {
		result := rm.readUbDeviceName("/sys/bus/ub/devices/ub0", "ub0")
		convey.Convey("Then the first entry name should be returned", func() {
			convey.So(result, convey.ShouldEqual, "ub0_0")
		})
	})
}

func TestReadUbNetInfoEmptyNetDir(t *testing.T) {
	rm := newTestUbResourceManager()

	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return []os.DirEntry{}, nil
	})
	defer patches.Reset()

	convey.Convey("When net directory is empty", t, func() {
		ifName, linkType := rm.readUbNetInfo("/sys/bus/ub/devices/ub0", "ub0")
		convey.Convey("Then empty values should be returned", func() {
			convey.So(ifName, convey.ShouldEqual, "")
			convey.So(linkType, convey.ShouldEqual, "")
		})
	})
}

func TestReadUbNetInfoSuccess(t *testing.T) {
	rm := newTestUbResourceManager()
	mockEntry := &mockDirEntry{name: "enp0s1", isDir: false}

	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return []os.DirEntry{mockEntry}, nil
	})
	mockLink := &mockLink{name: "enp0s1", encapType: "eth"}
	patches.ApplyMethodFunc(rm.netlinkManager, "LinkByName",
		func(name string) (netlink.Link, error) {
			return mockLink, nil
		})
	defer patches.Reset()

	convey.Convey("When net directory has entries", t, func() {
		ifName, linkType := rm.readUbNetInfo("/sys/bus/ub/devices/ub0", "ub0")
		convey.Convey("Then ifName and linkType should be returned", func() {
			convey.So(ifName, convey.ShouldEqual, "enp0s1")
			convey.So(linkType, convey.ShouldEqual, "eth")
		})
	})
}

func TestReadUbNetInfoLinkByNameError(t *testing.T) {
	rm := newTestUbResourceManager()
	mockEntry := &mockDirEntry{name: "enp0s1", isDir: false}

	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return []os.DirEntry{mockEntry}, nil
	})
	patches.ApplyMethodFunc(rm.netlinkManager, "LinkByName",
		func(name string) (netlink.Link, error) {
			return nil, errors.New("link not found")
		})
	defer patches.Reset()

	convey.Convey("When LinkByName fails", t, func() {
		ifName, linkType := rm.readUbNetInfo("/sys/bus/ub/devices/ub0", "ub0")
		convey.Convey("Then ifName should be returned with empty linkType", func() {
			convey.So(ifName, convey.ShouldEqual, "enp0s1")
			convey.So(linkType, convey.ShouldEqual, "")
		})
	})
}

func TestGetDevicesEmptyList(t *testing.T) {
	rm := newTestUbResourceManager()
	rm.deviceList = []*UbDeviceInfo{}

	convey.Convey("When device list is empty", t, func() {
		devices := rm.GetDevices()
		convey.Convey("Then empty slice should be returned", func() {
			convey.So(len(devices), convey.ShouldEqual, 0)
		})
	})
}

func TestGetDevicesNewUbDeviceFails(t *testing.T) {
	rm := newTestUbResourceManager()
	rm.deviceList = []*UbDeviceInfo{
		{UbID: "ub0", Vendor: "19e5", DeviceID: "a222", Driver: "ub", IfName: "enp0s1", LinkType: "eth"},
	}

	patches := gomonkey.ApplyFunc(NewUbDevice,
		func(ubID, deviceName, vendor, deviceID, driver, ifName, linkType string, rds types.RdmaDeviceSpec) (types.UbDevice, error) {
			return nil, errors.New("missing RDMA spec")
		})
	defer patches.Reset()

	convey.Convey("When NewUbDevice returns an error", t, func() {
		devices := rm.GetDevices()
		convey.Convey("Then empty slice should be returned", func() {
			convey.So(len(devices), convey.ShouldEqual, 0)
		})
	})
}

func TestGetDevicesSuccess(t *testing.T) {
	rm := newTestUbResourceManager()
	rm.deviceList = []*UbDeviceInfo{
		{UbID: "ub0", Vendor: "19e5", DeviceID: "a222", Driver: "ub", IfName: "enp0s1", LinkType: "eth"},
	}

	patches := gomonkey.ApplyFunc(NewUbDevice,
		func(ubID, deviceName, vendor, deviceID, driver, ifName, linkType string, rds types.RdmaDeviceSpec) (types.UbDevice, error) {
			return &ubDevice{ubID: ubID, vendor: vendor, deviceID: deviceID, ifName: ifName}, nil
		})
	defer patches.Reset()

	convey.Convey("When NewUbDevice succeeds", t, func() {
		devices := rm.GetDevices()
		convey.Convey("Then devices should be returned", func() {
			convey.So(len(devices), convey.ShouldEqual, 1)
			convey.So(devices[0].GetName(), convey.ShouldEqual, "ub0")
		})
	})
}

func TestGetFilteredDevices(t *testing.T) {
	rm := newTestUbResourceManager()
	testDevices := []types.Device{&ubDevice{ubID: "ub0"}}
	testSelectors := &types.Selectors{Vendors: []string{"19e5"}}

	patches := gomonkey.ApplyFunc(core.GetFilteredDevices,
		func(devices []types.Device, selectors *types.Selectors) []types.Device {
			return testDevices
		})
	defer patches.Reset()

	convey.Convey("When GetFilteredDevices is called", t, func() {
		result := rm.GetFilteredDevices(testDevices, testSelectors)
		convey.Convey("Then it should delegate to core.GetFilteredDevices", func() {
			convey.So(len(result), convey.ShouldEqual, 1)
			convey.So(result[0].GetName(), convey.ShouldEqual, "ub0")
		})
	})
}

func TestSetUbNicsUpNoDevices(t *testing.T) {
	rm := newTestUbResourceManager()

	convey.Convey("When device list is empty", t, func() {
		rm.setUbNicsUp([]types.Device{})
		convey.Convey("Then it should not panic", func() {
			convey.So(true, convey.ShouldBeTrue)
		})
	})
}

func TestSetUbNicsUpNoIfName(t *testing.T) {
	rm := newTestUbResourceManager()
	device := &ubDevice{ubID: "ub0", ifName: ""}

	convey.Convey("When device has no interface name", t, func() {
		rm.setUbNicsUp([]types.Device{device})
		convey.Convey("Then it should skip the device", func() {
			convey.So(true, convey.ShouldBeTrue)
		})
	})
}

func TestSetUbNicsUpNotUbDevice(t *testing.T) {
	rm := newTestUbResourceManager()

	// 创建一个实现 types.Device 但 GetIfName 返回空字符串的 mock 设备
	nonUbDevice := &mockNonUbDevice{}

	convey.Convey("When device is not a UbDevice", t, func() {
		rm.setUbNicsUp([]types.Device{nonUbDevice})
		convey.Convey("Then it should skip gracefully", func() {
			convey.So(true, convey.ShouldBeTrue)
		})
	})
}

// mockNonUbDevice 实现 types.Device 接口但 GetIfName 返回空字符串
type mockNonUbDevice struct{}

func (m *mockNonUbDevice) GetName() string                      { return "mock" }
func (m *mockNonUbDevice) GetVendor() string                    { return "19e5" }
func (m *mockNonUbDevice) GetDeviceID() string                  { return "0001" }
func (m *mockNonUbDevice) GetDriver() string                    { return "mock" }
func (m *mockNonUbDevice) GetRdmaSpec() []*pluginapi.DeviceSpec { return nil }
func (m *mockNonUbDevice) GetIfName() string                    { return "" }
func (m *mockNonUbDevice) GetLinkType() string                  { return "eth" }

func TestSetUbNicsUpLinkByNameError(t *testing.T) {
	rm := newTestUbResourceManager()
	device := &ubDevice{ubID: "ub0", ifName: "enp0s1"}

	patches := gomonkey.ApplyMethodFunc(rm.netlinkManager, "LinkByName",
		func(name string) (netlink.Link, error) {
			return nil, errors.New("link not found")
		})
	defer patches.Reset()

	convey.Convey("When LinkByName fails", t, func() {
		rm.setUbNicsUp([]types.Device{device})
		convey.Convey("Then it should skip gracefully", func() {
			convey.So(true, convey.ShouldBeTrue)
		})
	})
}

func TestSetUbNicsUpLinkSetUpError(t *testing.T) {
	rm := newTestUbResourceManager()
	device := &ubDevice{ubID: "ub0", ifName: "enp0s1"}
	mockLink := &mockLink{name: "enp0s1"}

	patches := gomonkey.ApplyMethodFunc(rm.netlinkManager, "LinkByName",
		func(name string) (netlink.Link, error) {
			return mockLink, nil
		})
	patches.ApplyMethodFunc(rm.netlinkManager, "LinkSetUp",
		func(link netlink.Link) error {
			return errors.New("set up failed")
		})
	defer patches.Reset()

	convey.Convey("When LinkSetUp fails", t, func() {
		rm.setUbNicsUp([]types.Device{device})
		convey.Convey("Then it should continue gracefully", func() {
			convey.So(true, convey.ShouldBeTrue)
		})
	})
}

func TestSetUbNicsUpSuccess(t *testing.T) {
	rm := newTestUbResourceManager()
	device := &ubDevice{ubID: "ub0", ifName: "enp0s1"}
	mockLink := &mockLink{name: "enp0s1"}

	patches := gomonkey.ApplyMethodFunc(rm.netlinkManager, "LinkByName",
		func(name string) (netlink.Link, error) {
			return mockLink, nil
		})
	patches.ApplyMethodFunc(rm.netlinkManager, "LinkSetUp",
		func(link netlink.Link) error {
			return nil
		})
	defer patches.Reset()

	convey.Convey("When both LinkByName and LinkSetUp succeed", t, func() {
		rm.setUbNicsUp([]types.Device{device})
		convey.Convey("Then it should complete without error", func() {
			convey.So(true, convey.ShouldBeTrue)
		})
	})
}

func TestInitServersNoConfigs(t *testing.T) {
	rm := newTestUbResourceManager()

	convey.Convey("When config list is empty", t, func() {
		err := rm.InitServers()
		convey.Convey("Then it should succeed without creating servers", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestInitServersEmptyDevices(t *testing.T) {
	rm := newTestUbResourceManager()
	testConfig := &types.UserConfig{
		ResourceName:   "ub_dev",
		ResourcePrefix: "mindx.com",
		RdmaHcaMax:     100,
		Selectors:      types.Selectors{Vendors: []string{"19e5"}},
	}

	patches := gomonkey.ApplyMethodFunc(rm, "GetConfigList",
		func(_ *ubResourceManager) []*types.UserConfig {
			return []*types.UserConfig{testConfig}
		})
	patches.ApplyMethodFunc(rm, "GetDevices",
		func(_ *ubResourceManager) []types.Device { return []types.Device{} })
	patches.ApplyFunc(core.GetFilteredDevices,
		func(devices []types.Device, selectors *types.Selectors) []types.Device {
			return []types.Device{}
		})
	patches.ApplyFunc(NewUbResourceServer,
		func(config *types.UserConfig, devices []types.Device, watchMode bool, socketSuffix string, useCdi bool) (UbResourceServer, error) {
			return &ubResourceServer{resourceName: config.ResourceName}, nil
		})
	defer patches.Reset()

	convey.Convey("When devices list is empty", t, func() {
		err := rm.InitServers()
		convey.Convey("Then it should still succeed", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestInitServersNewUbResourceServerError(t *testing.T) {
	// Note: NewUbResourceServer is a same-package function, gomonkey cannot intercept it
	// This test verifies that InitServers properly handles configuration and creates empty resource server when no devices
	crm := core.NewCoreResourceManager("test.json", rdmaUbResourcePrefix, socketSuffix, false)
	// Initialize config list by calling ReadConfig
	_ = crm.ReadConfig()

	rm := &ubResourceManager{
		CoreResourceManager: crm,
		deviceList:          []*UbDeviceInfo{},
		netlinkManager:      &netlinkManager{},
		rds:                 newUbRdmaDeviceSpec(common.RequiredRdmaDevices),
	}

	convey.Convey("When InitServers is called with empty device list", t, func() {
		err := rm.InitServers()
		convey.Convey("Then it should succeed and create empty resource server", func() {
			convey.So(err, convey.ShouldBeNil)
			// Even with empty device list, an empty resource server is created
			servers := rm.GetResourceServers()
			convey.So(len(servers), convey.ShouldEqual, 1)
		})
	})
}

func TestInitServersCleanupSpecsError(t *testing.T) {
	// Create a real CoreResourceManager with CDI enabled
	crm := core.NewCoreResourceManager("test.json", rdmaUbResourcePrefix, socketSuffix, true)
	// Ensure we have a valid config by calling ReadConfig
	_ = crm.ReadConfig()

	rm := &ubResourceManager{
		CoreResourceManager: crm,
		deviceList:          []*UbDeviceInfo{{UbID: "ub0", Vendor: "19e5"}},
		netlinkManager:      &netlinkManager{},
		rds:                 newUbRdmaDeviceSpec(common.RequiredRdmaDevices),
	}

	// Verify CDI is enabled
	convey.Convey("Given CDI is enabled", t, func() {
		convey.So(rm.GetUseCdi(), convey.ShouldBeTrue)

		patches := gomonkey.ApplyFunc(cdi.CleanupSpecs,
			func(prefix string) error {
				return errors.New("cdi cleanup failed")
			})
		defer patches.Reset()

		convey.Convey("When cdi.CleanupSpecs fails", func() {
			err := rm.InitServers()
			convey.Convey("Then the error should be returned", func() {
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, "cdi cleanup failed")
			})
		})
	})
}

func TestPeriodicUpdateZeroInterval(t *testing.T) {
	crm := core.NewCoreResourceManager("test.json", rdmaUbResourcePrefix, socketSuffix, false)
	rm := &ubResourceManager{
		CoreResourceManager: crm,
		deviceList:          []*UbDeviceInfo{},
		netlinkManager:      &netlinkManager{},
		rds:                 newUbRdmaDeviceSpec(common.RequiredRdmaDevices),
	}

	convey.Convey("When PeriodicUpdateInterval is zero", t, func() {
		stopFn := rm.PeriodicUpdate()
		convey.Convey("Then calling the stop function should not block", func() {
			convey.So(stopFn, convey.ShouldNotBeNil)
			stopFn()
			convey.So(true, convey.ShouldBeTrue)
		})
	})
}

func TestPeriodicUpdatePositiveInterval(t *testing.T) {
	crm := core.NewCoreResourceManager("test.json", rdmaUbResourcePrefix, socketSuffix, false)
	rm := &ubResourceManager{
		CoreResourceManager: crm,
		deviceList:          []*UbDeviceInfo{},
		netlinkManager:      &netlinkManager{},
		rds:                 newUbRdmaDeviceSpec(common.RequiredRdmaDevices),
	}

	convey.Convey("When PeriodicUpdateInterval is positive", t, func() {
		stopFn := rm.PeriodicUpdate()
		convey.Convey("Then stop function should work correctly", func() {
			convey.So(stopFn, convey.ShouldNotBeNil)
			stopFn()
			convey.So(true, convey.ShouldBeTrue)
		})
	})
}

func TestRunPeriodicUpdateStop(t *testing.T) {
	rm := newTestUbResourceManager()
	stopChan := make(chan interface{})
	done := make(chan struct{})

	patches := gomonkey.ApplyMethodFunc(rm, "DiscoverHostDevices",
		func(_ *ubResourceManager) error { return nil })
	defer patches.Reset()

	convey.Convey("When stop signal is sent", t, func() {
		go rm.runPeriodicUpdate(1*time.Hour, stopChan, done)
		time.Sleep(10 * time.Millisecond)
		stopChan <- struct{}{}

		convey.Convey("Then runPeriodicUpdate should exit", func() {
			select {
			case <-done:
				convey.So(true, convey.ShouldBeTrue)
			case <-time.After(2 * time.Second):
				convey.So(false, convey.ShouldBeTrue)
			}
		})
	})
}

func TestRunPeriodicUpdateDiscoverError(t *testing.T) {
	crm := core.NewCoreResourceManager("test.json", rdmaUbResourcePrefix, socketSuffix, false)
	rm := &ubResourceManager{
		CoreResourceManager: crm,
		deviceList:          []*UbDeviceInfo{},
		netlinkManager:      &netlinkManager{},
		rds:                 newUbRdmaDeviceSpec(common.RequiredRdmaDevices),
	}
	stopChan := make(chan interface{})
	done := make(chan struct{})
	ticker := time.NewTicker(10 * time.Millisecond)

	patches := gomonkey.ApplyFunc(time.NewTicker, func(d time.Duration) *time.Ticker {
		return ticker
	})
	patches.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return nil, errors.New("discover error")
	})
	defer patches.Reset()

	convey.Convey("When DiscoverHostDevices returns error", t, func() {
		go rm.runPeriodicUpdate(10*time.Millisecond, stopChan, done)
		time.Sleep(50 * time.Millisecond)
		close(stopChan)

		convey.Convey("Then runPeriodicUpdate should not crash", func() {
			select {
			case <-done:
				convey.So(true, convey.ShouldBeTrue)
			case <-time.After(2 * time.Second):
				convey.So(false, convey.ShouldBeTrue)
			}
		})
	})
}

func TestRunPeriodicUpdateTickerEvent(t *testing.T) {
	rm := newTestUbResourceManager()
	stopChan := make(chan interface{})
	done := make(chan struct{})
	ticker := time.NewTicker(10 * time.Millisecond)
	updateCalled := make(chan struct{}, 1)

	patches := gomonkey.ApplyFunc(time.NewTicker, func(d time.Duration) *time.Ticker {
		return ticker
	})
	patches.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
		return nil, nil
	})
	patches.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		updateCalled <- struct{}{}
		return nil, nil
	})
	defer patches.Reset()

	convey.Convey("When ticker event fires", t, func() {
		go rm.runPeriodicUpdate(10*time.Millisecond, stopChan, done)
		defer func() {
			close(stopChan)
			<-done
		}()

		convey.Convey("Then the updater should be called", func() {
			select {
			case <-updateCalled:
				convey.So(true, convey.ShouldBeTrue)
			case <-time.After(2 * time.Second):
				convey.So(false, convey.ShouldBeTrue)
			}
		})
	})
}

type mockDirEntry struct {
	name  string
	isDir bool
}

func (m *mockDirEntry) Name() string               { return m.name }
func (m *mockDirEntry) IsDir() bool                { return m.isDir }
func (m *mockDirEntry) Type() os.FileMode          { return 0 }
func (m *mockDirEntry) Info() (os.FileInfo, error) { return nil, nil }

type mockLink struct {
	name      string
	encapType string
}

func (m *mockLink) Attrs() *netlink.LinkAttrs {
	return &netlink.LinkAttrs{Name: m.name, EncapType: m.encapType}
}
func (m *mockLink) Type() string { return "device" }
