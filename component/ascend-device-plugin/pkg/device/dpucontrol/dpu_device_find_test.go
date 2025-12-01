/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package dpucontrol is used for find dpu.
package dpucontrol

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager"
)

const (
	npuIdFir = 1
	npuIdSec = 2
	npuIdThi = 3
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		panic(err)
	}
}

func TestSaveDpuConfigToNode(t *testing.T) {
	patcheIsExist := gomonkey.ApplyFunc(utils.IsExist, func(path string) bool { return true })
	defer patcheIsExist.Reset()
	convey.Convey("TestSaveDpuConfigToNode1", t, func() {
		expectedErr := "config load error"
		patches := gomonkey.ApplyPrivateMethod(&DpuFilter{}, "loadDpuConfigFromFile", func(_ *DpuFilter) error {
			return errors.New(expectedErr)
		})
		defer patches.Reset()
		df := DpuFilter{}
		err := df.SaveDpuConfToNode(&devmanager.DeviceManagerMock{})
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, expectedErr)
	})
	convey.Convey("TestSaveDpuConfigToNode2", t, func() {
		expectedErr := "dir read error"
		df := DpuFilter{}
		patche := gomonkey.ApplyPrivateMethod(&DpuFilter{}, "loadDpuConfigFromFile", func(_ *DpuFilter) error {
			df.UserConfig.BusType = busTypeUb
			return nil
		})
		defer patche.Reset()
		patche.ApplyFunc(os.ReadDir, func(path string) ([]os.DirEntry, error) {
			return nil, errors.New(expectedErr)
		})
		err := df.SaveDpuConfToNode(&devmanager.DeviceManagerMock{})
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, expectedErr)
	})
	convey.Convey("TestSaveDpuConfigToNode3", t, func() {
		expectedErr := "filter dpu err"
		df := DpuFilter{}
		var testNics []os.DirEntry
		patche := gomonkey.ApplyPrivateMethod(&DpuFilter{}, "loadDpuConfigFromFile", func(_ *DpuFilter) error {
			df.UserConfig.BusType = busTypeUb
			return nil
		})
		defer patche.Reset()
		patche.ApplyFunc(os.ReadDir, func(path string) ([]os.DirEntry, error) {
			return testNics, nil
		})
		patche.ApplyPrivateMethod(&df, "filterDpu", func(_ *DpuFilter) ([]BaseDpuInfo, error) {
			return nil, errors.New(expectedErr)
		})
		err := df.SaveDpuConfToNode(&devmanager.DeviceManagerMock{})
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, expectedErr)
	})
}

func TestSaveDpuConfigToNode2(t *testing.T) {
	patcheIsExist := gomonkey.ApplyFunc(utils.IsExist, func(path string) bool { return true })
	defer patcheIsExist.Reset()
	convey.Convey("TestSaveDpuConfigToNodeUBCase4", t, func() {
		const expectedErr = "get dpu info err"
		df := DpuFilter{}
		var testNics []os.DirEntry
		patche := gomonkey.ApplyPrivateMethod(&DpuFilter{}, "loadDpuConfigFromFile", func(_ *DpuFilter) error {
			df.UserConfig.BusType = busTypeUb
			return nil
		})
		defer patche.Reset()
		patche.ApplyFunc(os.ReadDir, func(path string) ([]os.DirEntry, error) {
			return testNics, nil
		})
		patche.ApplyPrivateMethod(&df, "filterDpu", func(_ *DpuFilter) ([]BaseDpuInfo, error) {
			return []BaseDpuInfo{}, nil
		})
		patche.ApplyPrivateMethod(&df, "getNpuCorrespDpuInfo", func() error {
			return errors.New(expectedErr)
		})
		err := df.SaveDpuConfToNode(&devmanager.DeviceManagerMock{})
		convey.So(err.Error(), convey.ShouldContainSubstring, expectedErr)
	})
	convey.Convey("TestSaveDpuConfigToNodePcieCase", t, func() {
		const expectedErr = "get dpu with pcie switch error"
		df := DpuFilter{}
		patche := gomonkey.ApplyPrivateMethod(&DpuFilter{}, "loadDpuConfigFromFile", func(_ *DpuFilter) error {
			df.UserConfig.BusType = busTypePcie
			return nil
		})
		defer patche.Reset()
		patche.ApplyPrivateMethod(&df, "getDpuWithNpuPcieSwitch", func() error {
			return errors.New(expectedErr)
		})
		err := df.SaveDpuConfToNode(&devmanager.DeviceManagerMock{})
		convey.So(err.Error(), convey.ShouldContainSubstring, expectedErr)
	})
	convey.Convey("TestSaveDpuConfigToNodeNoDpuInfos", t, func() {
		df := DpuFilter{}
		patche := gomonkey.ApplyPrivateMethod(&DpuFilter{}, "loadDpuConfigFromFile", func(_ *DpuFilter) error {
			df.UserConfig.BusType = busTypePcie
			return nil
		})
		defer patche.Reset()
		patche.ApplyPrivateMethod(&df, "getDpuWithNpuPcieSwitch", func() error {
			return nil
		})
		err := df.SaveDpuConfToNode(&devmanager.DeviceManagerMock{})
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestSaveDpuConfigToNode3(t *testing.T) {
	convey.Convey("TestNoDPUConfig", t, func() {
		df := DpuFilter{}
		err := df.SaveDpuConfToNode(&devmanager.DeviceManagerMock{})
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestSaveDpuConfigToNodeSuccess", t, func() {
		df := DpuFilter{}
		patcheIsExist := gomonkey.ApplyFunc(utils.IsExist, func(path string) bool { return true })
		defer patcheIsExist.Reset()
		patche := gomonkey.ApplyPrivateMethod(&DpuFilter{}, "loadDpuConfigFromFile", func(_ *DpuFilter) error {
			df.UserConfig.BusType = busTypePcie
			return nil
		})
		defer patche.Reset()
		patche.ApplyPrivateMethod(&df, "getDpuWithNpuPcieSwitch", func() error {
			return nil
		})
		df.NpuWithDpuInfos = []NpuWithDpuInfo{
			{
				NpuId:   1,
				DpuInfo: []BaseDpuInfo{},
			},
		}
		err := df.SaveDpuConfToNode(&devmanager.DeviceManagerMock{})
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGetDpuWithNpuPcieSwitch(t *testing.T) {
	df := &DpuFilter{}
	dmgr := &devmanager.DeviceManagerMock{}
	convey.Convey("TestGetDpuWithNpuPcieSwitch", t, func() {
		convey.Convey("TestGetCardListFailed", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyMethodReturn(dmgr, "GetCardList", nil, nil, errors.New("func failed"))
			err := df.getDpuWithNpuPcieSwitch(dmgr)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "func failed")
		})
		convey.Convey("TestGetPCIeBusInfoFailed", func() {
			const cardNum = int32(8)
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyMethodReturn(dmgr, "GetCardList", cardNum, []int32{0, 1, 2, 3, 4, 5, 6, 7}, nil)
			patches.ApplyMethodReturn(dmgr, "GetPCIeBusInfo", nil, errors.New("func failed"))
			err := df.getDpuWithNpuPcieSwitch(dmgr)
			convey.So(err.Error(), convey.ShouldContainSubstring, "func failed")
		})
	})
}

func TestAddDpuByNpuId(t *testing.T) {
	convey.Convey("TestAddDpuByNpuId", t, func() {
		df := &DpuFilter{NpuWithDpuInfos: make([]NpuWithDpuInfo, 0)}
		convey.Convey("When dpuinfos len is 1", func() {
			dpuInfos := []BaseDpuInfo{{}}
			df.addDpuByNpuId(npuIdFir, dpuIndexFir, dpuInfos)
			convey.So(len(df.NpuWithDpuInfos), convey.ShouldEqual, 1)
			convey.So(df.NpuWithDpuInfos[0].NpuId, convey.ShouldEqual, 1)
			convey.So(len(df.NpuWithDpuInfos[0].DpuInfo), convey.ShouldEqual, 1)
		})
		convey.Convey("When dpuinfos len is 2, dpu index is 0", func() {
			dpuInfos := []BaseDpuInfo{{}, {}}
			df.addDpuByNpuId(npuIdSec, dpuIndexFir, dpuInfos)
			convey.So(len(df.NpuWithDpuInfos), convey.ShouldEqual, 1)
			convey.So(df.NpuWithDpuInfos[0].NpuId, convey.ShouldEqual, 2)
			convey.So(len(df.NpuWithDpuInfos[0].DpuInfo), convey.ShouldEqual, 1)
		})
		convey.Convey("When dpuinfos len is 2, dpu index is 1", func() {
			dpuInfos := []BaseDpuInfo{{}, {}}
			df.addDpuByNpuId(npuIdThi, dpuIndexSec, dpuInfos)
			convey.So(len(df.NpuWithDpuInfos), convey.ShouldEqual, 1)
			convey.So(df.NpuWithDpuInfos[0].NpuId, convey.ShouldEqual, 3)
			convey.So(len(df.NpuWithDpuInfos[0].DpuInfo), convey.ShouldEqual, 1)
		})
	})
}

func TestGetDpuByPcieBusInfo(t *testing.T) {
	convey.Convey("TestGetDpuByPcieBusInfo", t, func() {
		df := &DpuFilter{}
		testPcieBusInfo := "test"
		convey.Convey("When func getPcieswByBusId failed", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyPrivateMethod(df, "getPcieswByBusId", func(_ *DpuFilter, _ string) (string, error) {
				return "", fmt.Errorf("getPcieswByBusId failed")
			})
			sw, dpuInfos, err := df.getDpuByPcieBusInfo(testPcieBusInfo)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(sw, convey.ShouldBeEmpty)
			convey.So(dpuInfos, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "getPcieswByBusId failed")
		})
		convey.Convey("When getNicsByPcieSw failed", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyPrivateMethod(df, "getPcieswByBusId", func(_ *DpuFilter, _ string) (string, error) {
				return "test-switch", nil
			})
			patches.ApplyPrivateMethod(df, "getNicsByPcieSw", func(_ *DpuFilter, _ string) ([]os.DirEntry, error) {
				return nil, fmt.Errorf("getNicsByPcieSw failed")
			})
			sw, dpuInfos, err := df.getDpuByPcieBusInfo(testPcieBusInfo)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(sw, convey.ShouldEqual, "test-switch")
			convey.So(dpuInfos, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "getNicsByPcieSw failed")
		})
	})
}

func TestGetDpuByPcieBusInfo2(t *testing.T) {
	convey.Convey("TestGetDpuByPcieBusInfo2", t, func() {
		df := &DpuFilter{}
		testPcieBusInfo := "test"
		var testNics []os.DirEntry
		testDpuInfos := []BaseDpuInfo{{}}
		convey.Convey("When filterDpu failed", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyPrivateMethod(df, "getPcieswByBusId", func(_ *DpuFilter, _ string) (string, error) {
				return "test-switch", nil
			})
			patches.ApplyPrivateMethod(df, "getNicsByPcieSw", func(_ *DpuFilter, _ string) ([]os.DirEntry, error) {
				return testNics, nil
			})
			patches.ApplyPrivateMethod(df, "filterDpu", func(_ *DpuFilter) ([]BaseDpuInfo, error) {
				return nil, fmt.Errorf("filterDpu failed")
			})
			sw, dpuInfos, err := df.getDpuByPcieBusInfo(testPcieBusInfo)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(sw, convey.ShouldEqual, "test-switch")
			convey.So(dpuInfos, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "filterDpu failed")
		})
		convey.Convey("When TestGetDpuByPcieBusInfo success", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyPrivateMethod(df, "getPcieswByBusId", func(_ *DpuFilter, _ string) (string, error) {
				return "test-switch", nil
			})
			patches.ApplyPrivateMethod(df, "getNicsByPcieSw", func(_ *DpuFilter, _ string) ([]os.DirEntry, error) {
				return testNics, nil
			})
			patches.ApplyPrivateMethod(df, "filterDpu", func(_ *DpuFilter) ([]BaseDpuInfo, error) {
				return testDpuInfos, nil
			})
			sw, dpuInfos, err := df.getDpuByPcieBusInfo(testPcieBusInfo)
			convey.So(err, convey.ShouldBeNil)
			convey.So(sw, convey.ShouldEqual, "test-switch")
			convey.So(dpuInfos, convey.ShouldNotBeEmpty)
			convey.So(len(dpuInfos), convey.ShouldEqual, 1)
		})
	})
}

func TestGetPcieswByBusId(t *testing.T) {
	df := DpuFilter{}
	convey.Convey("TestGetPcieswByBusId", t, func() {
		patch := gomonkey.ApplyFunc(
			os.Readlink, func(path string) (string, error) {
				return "", os.ErrNotExist
			},
		)
		defer patch.Reset()
		pcieSw, err := df.getPcieswByBusId("invalid_bus_id")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(pcieSw, convey.ShouldBeEmpty)
	},
	)

	convey.Convey("TestGetPcieswByBusId2", t, func() {
		patch := gomonkey.ApplyFunc(
			os.Readlink, func(path string) (string, error) {
				return "/sys/bus/pci/devices/0000:01:00.0", nil
			},
		)
		defer patch.Reset()
		absPatch := gomonkey.ApplyFunc(
			filepath.Abs, func(path string) (string, error) {
				return "", fmt.Errorf("filepath.Abs error")
			},
		)
		defer absPatch.Reset()

		pcieSw, err := df.getPcieswByBusId("valid_bus_id")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(pcieSw, convey.ShouldBeEmpty)
	},
	)
}

func TestGetPcieswByBusIdSuccessAndErrorLength(t *testing.T) {
	df := DpuFilter{}
	convey.Convey("TestGetPcieswByBusId successfully", t, func() {
		readlinkPatch := gomonkey.ApplyFunc(
			os.Readlink, func(path string) (string, error) {
				return "/devices/pci0003:c0/0000:c0:00.0", nil
			},
		)
		defer readlinkPatch.Reset()
		absPatch := gomonkey.ApplyFunc(
			filepath.Abs, func(path string) (string, error) {
				return "/devices/0000:01:00.0/0000:c0:00.0", nil
			},
		)
		defer absPatch.Reset()
		pcieSw, err := df.getPcieswByBusId("valid_bus_id")
		convey.So(err, convey.ShouldBeNil)
		convey.So(pcieSw, convey.ShouldEqual, "/sys/devices/0000:01:00.0")
	},
	)

	convey.Convey("TestGetPcieswByBusId error path length", t, func() {
		readlinkPatch := gomonkey.ApplyFunc(
			os.Readlink, func(path string) (string, error) {
				return "/devices", nil
			},
		)
		defer readlinkPatch.Reset()
		absPatch := gomonkey.ApplyFunc(
			filepath.Abs, func(path string) (string, error) {
				return "/devices", nil
			},
		)
		defer absPatch.Reset()

		result, err := df.getPcieswByBusId("error_bus_id")
		convey.So(result, convey.ShouldBeEmpty)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestGetNicsByPcieSw(t *testing.T) {
	df := DpuFilter{}
	convey.Convey("TestGetNicsByPcieSw", t, func() {
		patch := gomonkey.ApplyFunc(
			os.ReadDir, func(path string) ([]os.DirEntry, error) {
				return nil, os.ErrNotExist
			},
		)
		defer patch.Reset()
		nics, err := df.getNicsByPcieSw("invalid_bus_id")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(nics, convey.ShouldBeEmpty)
	},
	)
}

type FakeDirEntry struct {
	name string
	typ  fs.FileMode
	info fs.FileInfo
}

func (f FakeDirEntry) Name() string               { return f.name }
func (f FakeDirEntry) IsDir() bool                { return false }
func (f FakeDirEntry) Type() fs.FileMode          { return f.typ }
func (f FakeDirEntry) Info() (fs.FileInfo, error) { return f.info, nil }

func TestGetNicsByPcieWithMockEntries(t *testing.T) {
	df := DpuFilter{}
	entries := []os.DirEntry{
		FakeDirEntry{name: "lo", typ: fs.ModeSymlink},
		FakeDirEntry{name: "eth0", typ: fs.ModeSymlink},
	}
	convey.Convey("test error read link", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(os.ReadDir, func(path string) ([]os.DirEntry, error) {
			return entries, nil
		})
		patches.ApplyFunc(os.Readlink, func(path string) (string, error) {
			return "exist_bus_id", fmt.Errorf("readlink error")
		})
		nics, err := df.getNicsByPcieSw("exist_bus_id")
		convey.So(err, convey.ShouldBeNil)
		convey.So(nics, convey.ShouldBeEmpty)
	})
	convey.Convey("test no link contains bus id", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(os.ReadDir, func(path string) ([]os.DirEntry, error) {
			return entries, nil
		})
		patches.ApplyFunc(os.Readlink, func(path string) (string, error) {
			return "", fmt.Errorf("readlink error")
		})
		nics, err := df.getNicsByPcieSw("nonexistent_bus_id")
		convey.So(err, convey.ShouldBeNil)
		convey.So(nics, convey.ShouldBeEmpty)
	})
	convey.Convey("test normal case", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(os.ReadDir, func(path string) ([]os.DirEntry, error) {
			return entries, nil
		})
		patches.ApplyFunc(os.Readlink, func(path string) (string, error) {
			return "exist_bus_id", fmt.Errorf("readlink error")
		})
		nics, err := df.getNicsByPcieSw("exist_bus_id")
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(nics), convey.ShouldNotBeEmpty)
	})
}

func TestFilterDpuErrorCase(t *testing.T) {
	convey.Convey("TestFilterDpu", t, func() {
		df := DpuFilter{}
		df.UserConfig = UserDpuConfig{
			Selectors: &DeviceSelectors{},
		}
		convey.Convey("error entries length", func() {
			df.entries = []os.DirEntry{}
			_, err := df.filterDpu()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "the lengh of df.entries is invalid: 0")
		})
		convey.Convey("test error read link", func() {
			df.entries = []os.DirEntry{FakeDirEntry{}}
			patch := gomonkey.ApplyFunc(os.Readlink, func(path string) (string, error) {
				return "", fmt.Errorf("readlink error")
			})
			defer patch.Reset()
			dpuInfos, err := df.filterDpu()
			convey.So(err, convey.ShouldBeNil)
			convey.So(dpuInfos, convey.ShouldBeEmpty)
		})
	})
}

func TestDpuFilterSuccessCase(t *testing.T) {
	df := DpuFilter{}
	df.UserConfig = UserDpuConfig{
		Selectors: &DeviceSelectors{},
	}
	convey.Convey("test normal case", t, func() {
		df.entries = []os.DirEntry{
			FakeDirEntry{name: "lo", typ: fs.ModeSymlink},
			FakeDirEntry{name: "eth0", typ: fs.ModeSymlink},
		}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(os.Readlink, func(path string) (string, error) {
			return "", nil
		})
		patches.ApplyPrivateMethod(&df, "shouldFilterByVendor",
			func(string, []string) (bool, string) {
				return false, "VendorValue"
			})
		patches.ApplyPrivateMethod(&df, "shouldFilterByDeviceID",
			func(string, []string) (bool, string) {
				return false, "DeviceID"
			})
		patches.ApplyPrivateMethod(&df, "shouldFilterByDeviceName",
			func(string, []string) (bool, string) {
				return false, "DeviceName"
			})
		patches.ApplyFunc(getInterfaceIPs, func(string) string {
			return ""
		})
		dpuInfos, err := df.filterDpu()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(dpuInfos), convey.ShouldEqual, len(df.entries))
	})
}

func TestShouldFilterByVendor(t *testing.T) {
	df := DpuFilter{}
	convey.Convey("TestShouldFilterByVendor", t, func() {
		patch := gomonkey.ApplyFunc(
			readFileContent, func(path string) (string, error) {
				return "", fmt.Errorf("read file error")
			},
		)
		defer patch.Reset()
		shouldFilter, vendorValue := df.shouldFilterByVendor("/fake/path", []string{"vendor1"})
		convey.So(shouldFilter, convey.ShouldBeTrue)
		convey.So(vendorValue, convey.ShouldBeEmpty)
	},
	)

	convey.Convey("TestShouldFilterByVendor2", t, func() {
		patch := gomonkey.ApplyFunc(
			readFileContent, func(path string) (string, error) {
				return "vendor2", nil
			},
		)
		defer patch.Reset()
		shouldFilter, vendorValue := df.shouldFilterByVendor("/fake/path", []string{"vendor1"})
		convey.So(shouldFilter, convey.ShouldBeTrue)
		convey.So(vendorValue, convey.ShouldBeEmpty)
	},
	)

	convey.Convey("TestShouldFilterByVendor3", t, func() {
		patch := gomonkey.ApplyFunc(
			readFileContent, func(path string) (string, error) {
				return "vendor3", nil
			},
		)
		defer patch.Reset()
		shouldFilter, vendorValue := df.shouldFilterByVendor("/fake/path", []string{"vendor3"})
		convey.So(shouldFilter, convey.ShouldBeFalse)
		convey.So(vendorValue, convey.ShouldEqual, "vendor3")
	},
	)

	convey.Convey("TestShouldFilterByVendor4", t, func() {
		patch := gomonkey.ApplyFunc(
			readFileContent, func(path string) (string, error) {
				return "vendor4", nil
			},
		)
		defer patch.Reset()
		shouldFilter, vendorValue := df.shouldFilterByVendor("/fake/path", []string{})
		convey.So(shouldFilter, convey.ShouldBeFalse)
		convey.So(vendorValue, convey.ShouldEqual, "vendor4")
	},
	)
}

func TestShouldFilterByDeviceID(t *testing.T) {
	df := DpuFilter{}
	convey.Convey("TestShouldFilterByDeviceID", t, func() {
		patch := gomonkey.ApplyFunc(
			readFileContent, func(path string) (string, error) {
				return "", fmt.Errorf("read file error")
			},
		)
		defer patch.Reset()
		shouldFilter, deviceIdValue := df.shouldFilterByDeviceID("/fake/path", []string{"deviceId1"})
		convey.So(shouldFilter, convey.ShouldBeTrue)
		convey.So(deviceIdValue, convey.ShouldBeEmpty)
	},
	)
}

func TestShouldFilterByDeviceName(t *testing.T) {
	df := DpuFilter{}
	convey.Convey("TestShouldFilterByDeviceName Successfully", t, func() {
		deviceNames := []string{"device1", "device2"}
		convey.So(df.shouldFilterByDeviceName("device0", deviceNames), convey.ShouldBeTrue)
	})

	convey.Convey("TestShouldFilterByDeviceName Failed", t, func() {
		deviceNames := []string{"device1", "device2"}
		convey.So(df.shouldFilterByDeviceName("device0", []string{}), convey.ShouldBeFalse)
		convey.So(df.shouldFilterByDeviceName("device1", deviceNames), convey.ShouldBeFalse)
	})
}

func TestGetSlotId(t *testing.T) {
	convey.Convey("TestGetSlotId", t, func() {
		df := DpuFilter{}
		patch := gomonkey.ApplyFunc(
			os.Readlink, func(path string) (string, error) {
				return "", fmt.Errorf("readlink error")
			},
		)
		defer patch.Reset()
		slotIdValue, err := df.getSlotId("eth0")
		convey.So(slotIdValue, convey.ShouldBeEmpty)
		convey.So(err, convey.ShouldBeError)
	},
	)
	convey.Convey("TestGetSlotId2", t, func() {
		df := DpuFilter{
			UserConfig: UserDpuConfig{
				BusType:   busTypePcie,
				Selectors: &DeviceSelectors{},
			},
		}
		patch := gomonkey.ApplyFunc(
			os.Readlink, func(path string) (string, error) {
				return "device", nil
			},
		)
		defer patch.Reset()
		slotIdValue, err := df.getSlotId("eth0")
		convey.So(slotIdValue, convey.ShouldBeEmpty)
		convey.So(err, convey.ShouldBeError)
		convey.So(err.Error(), convey.ShouldEqual, "busType is pcie not ub")
	},
	)
}

func TestGetSlotId2(t *testing.T) {
	convey.Convey("TestGetSlotId3", t, func() {
		patch := gomonkey.ApplyFunc(
			os.Readlink, func(path string) (string, error) {
				return "eth0", nil
			},
		)
		patch.ApplyFunc(
			readFileContent, func(path string) (string, error) {
				return "slot1", nil
			},
		)
		defer patch.Reset()

		df := DpuFilter{
			UserConfig: UserDpuConfig{
				BusType:   busTypeUb,
				Selectors: &DeviceSelectors{},
			},
		}
		slotIdValue, err := df.getSlotId("eth0")
		convey.So(slotIdValue, convey.ShouldEqual, "slot1")
		convey.So(err, convey.ShouldBeNil)
	},
	)
}

func TestLoadConfigFromFile(t *testing.T) {
	convey.Convey("TestLoadConfigFromFile", t, func() {
		patch := gomonkey.ApplyFunc(utils.LoadFile, func(path string) ([]byte, error) {
			return nil, fmt.Errorf("load file error")
		})
		defer patch.Reset()
		df := DpuFilter{}
		err := df.loadDpuConfigFromFile()
		convey.So(err, convey.ShouldBeError)
		convey.So(err.Error(), convey.ShouldEqual, "load config from file error:load file error")
	})

	convey.Convey("TestLoadConfigFromFile2", t, func() {
		patch := gomonkey.ApplyFunc(utils.LoadFile, func(path string) ([]byte, error) {
			return []byte("invalid json"), nil
		})
		defer patch.Reset()
		df := DpuFilter{}
		err := df.loadDpuConfigFromFile()
		convey.So(err, convey.ShouldBeError)
		convey.So(err.Error(), convey.ShouldStartWith, "parse config from file error:")
	})

	convey.Convey("TestLoadConfigFromFile3", t, func() {
		config := ConfigList{
			UserDpuConfigList: []UserDpuConfig{
				{
					Selectors: &DeviceSelectors{},
					BusType:   "",
				},
			},
		}
		jsonContent, err0 := json.Marshal(config)
		convey.So(err0, convey.ShouldBeNil)
		patch := gomonkey.ApplyFunc(utils.LoadFile, func(path string) ([]byte, error) {
			return jsonContent, nil
		})
		defer patch.Reset()
		df := DpuFilter{}
		err := df.loadDpuConfigFromFile()
		convey.So(err, convey.ShouldBeError)
		convey.So(err.Error(), convey.ShouldEqual, "config missing parameter, dpu devices find is not enable")
	})
}

func TestPrivateGetDpuPair(t *testing.T) {
	convey.Convey("TestGetDpuPair1", t, func() {
		df := &DpuFilter{}
		dpus := df.getDpuPair("slot1", "slot2")
		convey.So(dpus, convey.ShouldBeNil)
	})

	convey.Convey("TestGetDpuPair2", t, func() {
		const OneNpuCorresTwoDpu = 2
		patch := gomonkey.ApplyPrivateMethod(&DpuFilter{}, "getSlotId",
			func(_ DpuFilter, ifaceName string) (string, error) {
				return "slot1", nil
			})
		defer patch.Reset()
		df := DpuFilter{
			dpuInfos: []BaseDpuInfo{
				{DeviceName: "eth0"},
				{DeviceName: "eth1"},
			},
		}
		dpus := df.getDpuPair("slot1", "slot2")
		convey.So(len(dpus), convey.ShouldEqual, OneNpuCorresTwoDpu)
		convey.So(dpus[0].DeviceName, convey.ShouldEqual, "eth0")
	})

	convey.Convey("TestGetDpuPair3", t, func() {
		patch := gomonkey.ApplyPrivateMethod(&DpuFilter{}, "getSlotId",
			func(_ DpuFilter, DeviceName string) (string, error) {
				return "slot9", nil
			})
		defer patch.Reset()
		df := DpuFilter{
			dpuInfos: []BaseDpuInfo{
				{DeviceName: "eth0"},
				{DeviceName: "eth1"},
			},
		}
		dpus := df.getDpuPair("slot1", "slot2")
		convey.So(len(dpus), convey.ShouldEqual, 0)
	})
}

func TestGetNpuCorrespondDpuInfo(t *testing.T) {
	convey.Convey("TestGetNpuCorrespDpuInfo1", t, func() {
		patch := gomonkey.ApplyPrivateMethod(&DpuFilter{}, "getDpuPair", func(_ DpuFilter, _, _ string) []BaseDpuInfo {
			return []BaseDpuInfo{
				{DeviceName: "eth0"},
				{DeviceName: "eth1"},
			}
		})
		defer patch.Reset()
		df := DpuFilter{}
		err := df.getNpuCorrespDpuInfo()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(df.NpuWithDpuInfos), convey.ShouldEqual, api.NpuCountPerNode)
	})

	convey.Convey("TestGetNpuCorrespDpuInfo2", t, func() {
		patch := gomonkey.ApplyPrivateMethod(&DpuFilter{}, "getDpuPair", func(_ DpuFilter, _, _ string) []BaseDpuInfo {
			return []BaseDpuInfo{
				{DeviceName: "eth0"},
			}
		})
		defer patch.Reset()
		df := DpuFilter{}
		err := df.getNpuCorrespDpuInfo()
		convey.So(err, convey.ShouldBeError)
		convey.So(err.Error(), convey.ShouldEqual, "get npu 0 correspond dpuinfos error")
	})
}

func TestReadFileContent(t *testing.T) {
	convey.Convey("TestReadFileContent", t, func() {
		patch := gomonkey.ApplyFunc(utils.LoadFile, func(filename string) ([]byte, error) {
			return nil, os.ErrNotExist
		})
		defer patch.Reset()
		content, err := readFileContent("nonexistent.txt")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(content, convey.ShouldBeEmpty)
	})

	convey.Convey("TestReadFileContent2", t, func() {
		patch := gomonkey.ApplyFunc(utils.LoadFile, func(filename string) ([]byte, error) {
			return []byte{}, nil
		})
		defer patch.Reset()
		content, err := readFileContent("empty.txt")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(content, convey.ShouldBeEmpty)
		convey.So(err.Error(), convey.ShouldEqual, "file empty.txt is empty")
	})

	convey.Convey("Test ReadFileContent normally", t, func() {
		patch := gomonkey.ApplyFunc(utils.LoadFile, func(filename string) ([]byte, error) {
			return []byte(" hello world "), nil
		})
		defer patch.Reset()
		content, err := readFileContent("normal.txt")
		convey.So(content, convey.ShouldEqual, "hello world")
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGetInterfaceIps(t *testing.T) {
	convey.Convey("TestGetInterfaceIps2", t, func() {
		mockIface := net.Interface{Index: 1, Name: "single_addr_iface"}

		patch := gomonkey.ApplyFunc(net.InterfaceByName, func(name string) (*net.Interface, error) {
			return &mockIface, nil
		})
		defer patch.Reset()
		const (
			mockIP   = 127
			mockMask = 255
		)
		addr := &net.IPNet{IP: net.IPv4(mockIP, 0, 0, 1),
			Mask: net.IPv4Mask(mockMask, mockMask, mockMask, mockMask)}
		addrsPatch := gomonkey.ApplyFunc(mockIface.Addrs, func() ([]net.Addr, error) {
			return []net.Addr{addr}, nil
		})
		defer addrsPatch.Reset()
		ip := getInterfaceIPs("single_addr_iface")
		convey.So(ip, convey.ShouldEqual, "127.0.0.1")
	})

	convey.Convey("Error get interface", t, func() {
		patch := gomonkey.ApplyFunc(net.InterfaceByName, func(name string) (*net.Interface, error) {
			return nil, fmt.Errorf("get interface error")
		})
		defer patch.Reset()

		ip := getInterfaceIPs("single_addr_iface")
		convey.So(ip, convey.ShouldBeEmpty)
	})

	convey.Convey("Error get address", t, func() {
		mockIface := net.Interface{}
		patch := gomonkey.ApplyFunc(net.InterfaceByName, func(name string) (*net.Interface, error) {
			return &mockIface, nil
		})
		defer patch.Reset()

		addrsPatch := gomonkey.ApplyFunc(mockIface.Addrs, func() ([]net.Addr, error) {
			return []net.Addr{}, fmt.Errorf("get addrs error")
		})
		defer addrsPatch.Reset()

		ip := getInterfaceIPs("single_addr_iface")
		convey.So(ip, convey.ShouldBeEmpty)
	})
}
