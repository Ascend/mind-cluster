/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package device a series of device function
package device

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager"
)

func mockNumaNodeManager() *NumaNodeManager {
	return NewNumaNodeManager(&devmanager.DeviceManagerMock{})
}

// TestInitAllNumaNodeInfo_GlobFails test InitAllNumaNodeInfo when filepath.Glob fails
func TestInitAllNumaNodeInfo_GlobFails(t *testing.T) {
	convey.Convey("01-should return error when filepath.Glob fails", t, func() {
		mgr := mockNumaNodeManager()
		patchGlob := gomonkey.ApplyFunc(filepath.Glob,
			func(pattern string) ([]string, error) {
				return nil, errors.New("glob error")
			})
		defer patchGlob.Reset()

		err := mgr.InitAllNumaNodeInfo(&common.NpuAllInfo{})
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "glob error")
	})
}

// TestInitAllNumaNodeInfo_NoNumaNodes test InitAllNumaNodeInfo when no numa nodes found
func TestInitAllNumaNodeInfo_NoNumaNodes(t *testing.T) {
	convey.Convey("02-should return nil when no numa nodes found", t, func() {
		mgr := mockNumaNodeManager()
		patchGlob := gomonkey.ApplyFunc(filepath.Glob,
			func(pattern string) ([]string, error) {
				return []string{}, nil
			})
		defer patchGlob.Reset()

		err := mgr.InitAllNumaNodeInfo(&common.NpuAllInfo{})
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestInitAllNumaNodeInfo_TooManyNumaNodes test InitAllNumaNodeInfo when too many numa nodes
func TestInitAllNumaNodeInfo_TooManyNumaNodes(t *testing.T) {
	convey.Convey("03-should return error when too many numa nodes", t, func() {
		mgr := mockNumaNodeManager()
		manyNodes := make([]string, maxNumaNodePathCount+1)
		for i := 0; i < maxNumaNodePathCount+1; i++ {
			manyNodes[i] = "/sys/devices/system/node/node" + string(rune('0'+i%10))
		}
		patchGlob := gomonkey.ApplyFunc(filepath.Glob,
			func(pattern string) ([]string, error) {
				return manyNodes, nil
			})
		defer patchGlob.Reset()

		err := mgr.InitAllNumaNodeInfo(&common.NpuAllInfo{})
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "too many numa node paths")
	})
}

// TestInitAllNumaNodeInfo_SkipInvalidNodeID test InitAllNumaNodeInfo skips invalid node id
func TestInitAllNumaNodeInfo_SkipInvalidNodeID(t *testing.T) {
	convey.Convey("04-should skip invalid node id and continue", t, func() {
		mgr := mockNumaNodeManager()
		patchGlob := gomonkey.ApplyFunc(filepath.Glob,
			func(pattern string) ([]string, error) {
				return []string{
					"/sys/devices/system/node/node0",
					"/sys/devices/system/node/nodeabc",
					"/sys/devices/system/node/node1",
				}, nil
			})
		defer patchGlob.Reset()

		readCallCount := 0
		patchRead := gomonkey.ApplyFunc(utils.ReadLimitBytes,
			func(filePath string, limit int) ([]byte, error) {
				readCallCount++
				if filePath == "/sys/devices/system/node/node0/cpulist" {
					return []byte("0-15\n"), nil
				}
				if filePath == "/sys/devices/system/node/node1/cpulist" {
					return []byte("16-31\n"), nil
				}
				return nil, errors.New("unexpected path")
			})
		defer patchRead.Reset()

		err := mgr.InitAllNumaNodeInfo(&common.NpuAllInfo{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(readCallCount, convey.ShouldEqual, 2)
		mgr.cpuListToNumaNodeMu.RLock()
		convey.So(mgr.cpuListToNumaNode, convey.ShouldContainKey, "0-15")
		convey.So(mgr.cpuListToNumaNode, convey.ShouldContainKey, "16-31")
		convey.So(len(mgr.cpuListToNumaNode), convey.ShouldEqual, 2)
		mgr.cpuListToNumaNodeMu.RUnlock()
	})
}

// TestInitAllNumaNodeInfo_SkipReadCpuListFails test InitAllNumaNodeInfo skips when read cpulist fails
func TestInitAllNumaNodeInfo_SkipReadCpuListFails(t *testing.T) {
	convey.Convey("05-should skip when read cpulist fails", t, func() {
		mgr := mockNumaNodeManager()
		patchGlob := gomonkey.ApplyFunc(filepath.Glob,
			func(pattern string) ([]string, error) {
				return []string{
					"/sys/devices/system/node/node0",
					"/sys/devices/system/node/node1",
				}, nil
			})
		defer patchGlob.Reset()

		patchRead := gomonkey.ApplyFunc(utils.ReadLimitBytes,
			func(filePath string, limit int) ([]byte, error) {
				if filePath == "/sys/devices/system/node/node0/cpulist" {
					return []byte("0-15\n"), nil
				}
				return nil, errors.New("read error")
			})
		defer patchRead.Reset()

		err := mgr.InitAllNumaNodeInfo(&common.NpuAllInfo{})
		convey.So(err, convey.ShouldBeNil)
		mgr.cpuListToNumaNodeMu.RLock()
		convey.So(len(mgr.cpuListToNumaNode), convey.ShouldEqual, 1)
		convey.So(mgr.cpuListToNumaNode["0-15"], convey.ShouldEqual, int64(0))
		mgr.cpuListToNumaNodeMu.RUnlock()
	})
}

// TestInitAllNumaNodeInfo_PopulateAllValidNodes test InitAllNumaNodeInfo populates all valid nodes
func TestInitAllNumaNodeInfo_PopulateAllValidNodes(t *testing.T) {
	convey.Convey("06-should correctly populate map for all valid nodes", t, func() {
		mgr := mockNumaNodeManager()
		patchGlob := gomonkey.ApplyFunc(filepath.Glob,
			func(pattern string) ([]string, error) {
				return []string{
					"/sys/devices/system/node/node0",
					"/sys/devices/system/node/node1",
					"/sys/devices/system/node/node2",
				}, nil
			})
		defer patchGlob.Reset()

		patchRead := gomonkey.ApplyFunc(utils.ReadLimitBytes,
			func(filePath string, limit int) ([]byte, error) {
				switch filePath {
				case "/sys/devices/system/node/node0/cpulist":
					return []byte("0-15\n"), nil
				case "/sys/devices/system/node/node1/cpulist":
					return []byte("16-31\n"), nil
				case "/sys/devices/system/node/node2/cpulist":
					return []byte("32-47\n"), nil
				}
				return nil, errors.New("unexpected path")
			})
		defer patchRead.Reset()

		err := mgr.InitAllNumaNodeInfo(&common.NpuAllInfo{})
		convey.So(err, convey.ShouldBeNil)
		mgr.cpuListToNumaNodeMu.RLock()
		convey.So(mgr.cpuListToNumaNode["0-15"], convey.ShouldEqual, int64(0))
		convey.So(mgr.cpuListToNumaNode["16-31"], convey.ShouldEqual, int64(1))
		convey.So(mgr.cpuListToNumaNode["32-47"], convey.ShouldEqual, int64(2))
		convey.So(len(mgr.cpuListToNumaNode), convey.ShouldEqual, 3)
		mgr.cpuListToNumaNodeMu.RUnlock()
	})
}

// TestSetNumaNodes_CpuListMatches test setNumaNodes via affinity cpu (path A)
func TestSetNumaNodes_CpuListMatches(t *testing.T) {
	convey.Convey("01-should set cache via affinity cpu", t, func() {
		mgr := mockNumaNodeManager()
		mgr.cpuListToNumaNodeMu.Lock()
		mgr.cpuListToNumaNode["0-15"] = int64(0)
		mgr.cpuListToNumaNodeMu.Unlock()

		dev := common.NpuDevice{LogicID: 0, PhyID: 0, DeviceName: "Ascend910-0"}

		patchGetAffinity := gomonkey.ApplyMethod(reflect.TypeOf(&devmanager.DeviceManagerMock{}),
			"GetAffinityCpuInfo",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (string, error) {
				return "0-15", nil
			})
		defer patchGetAffinity.Reset()

		mgr.setNumaNodes(&dev)
		nodes := mgr.getNumaNodesByPhyID(0)
		convey.So(len(nodes), convey.ShouldEqual, 1)
		convey.So(nodes[0], convey.ShouldEqual, int64(0))
	})
}

// TestSetNumaNodes_GetAffinityFails test setNumaNodes when both affinity and pcie fallback fail
func TestSetNumaNodes_GetAffinityFails(t *testing.T) {
	convey.Convey("02-should not set cache when both affinity and pcie fail", t, func() {
		mgr := mockNumaNodeManager()
		dev := common.NpuDevice{LogicID: 1, PhyID: 1, DeviceName: "Ascend910-1"}

		patchAffinity := gomonkey.ApplyMethod(reflect.TypeOf(&devmanager.DeviceManagerMock{}),
			"GetAffinityCpuInfo",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (string, error) {
				return "", errors.New("dcmi error")
			})
		defer patchAffinity.Reset()
		patchPcie := gomonkey.ApplyMethod(reflect.TypeOf(&devmanager.DeviceManagerMock{}),
			"GetPCIeBusInfo",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (string, error) {
				return "", errors.New("pcie error")
			})
		defer patchPcie.Reset()

		mgr.setNumaNodes(&dev)
		convey.So(mgr.getNumaNodesByPhyID(1), convey.ShouldBeNil)
	})
}

// TestSetNumaNodes_CpuListNotFound test setNumaNodes when cpu list not found & pcie fallback fails
func TestSetNumaNodes_CpuListNotFound(t *testing.T) {
	convey.Convey("03-should not set cache when cpu list not found & pcie fails", t, func() {
		mgr := mockNumaNodeManager()
		dev := common.NpuDevice{LogicID: 2, PhyID: 2, DeviceName: "Ascend910-2"}

		patchAffinity := gomonkey.ApplyMethod(reflect.TypeOf(&devmanager.DeviceManagerMock{}),
			"GetAffinityCpuInfo",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (string, error) {
				return "64-79", nil
			})
		defer patchAffinity.Reset()
		patchPcie := gomonkey.ApplyMethod(reflect.TypeOf(&devmanager.DeviceManagerMock{}),
			"GetPCIeBusInfo",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (string, error) {
				return "", errors.New("pcie error")
			})
		defer patchPcie.Reset()

		mgr.setNumaNodes(&dev)
		convey.So(mgr.getNumaNodesByPhyID(2), convey.ShouldBeNil)
	})
}

// TestSetNumaNodes_EmptyCpuList test setNumaNodes handles empty cpu list & pcie fallback fails
func TestSetNumaNodes_EmptyCpuList(t *testing.T) {
	convey.Convey("04-should not set cache when empty cpu list & pcie fails", t, func() {
		mgr := mockNumaNodeManager()
		dev := common.NpuDevice{LogicID: 3, PhyID: 3, DeviceName: "Ascend910-3"}

		patchAffinity := gomonkey.ApplyMethod(reflect.TypeOf(&devmanager.DeviceManagerMock{}),
			"GetAffinityCpuInfo",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (string, error) {
				return "", nil
			})
		defer patchAffinity.Reset()
		patchPcie := gomonkey.ApplyMethod(reflect.TypeOf(&devmanager.DeviceManagerMock{}),
			"GetPCIeBusInfo",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (string, error) {
				return "", errors.New("pcie error")
			})
		defer patchPcie.Reset()

		mgr.setNumaNodes(&dev)
		convey.So(mgr.getNumaNodesByPhyID(3), convey.ShouldBeNil)
	})
}

// TestSetNumaNodes_PcieFallbackSuccess test setNumaNodes pcie fallback succeeds
func TestSetNumaNodes_PcieFallbackSuccess(t *testing.T) {
	convey.Convey("05-should set cache via pcie fallback when affinity fails", t, func() {
		mgr := mockNumaNodeManager()
		dev := common.NpuDevice{LogicID: 0, PhyID: 0, DeviceName: "Ascend910-0"}

		patchAffinity := gomonkey.ApplyMethod(reflect.TypeOf(&devmanager.DeviceManagerMock{}),
			"GetAffinityCpuInfo",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (string, error) {
				return "", errors.New("dcmi error")
			})
		defer patchAffinity.Reset()
		patchRead := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink,
			func(path string, limit int, validator func(string) bool) ([]byte, error) {
				return []byte("1\n"), nil
			})
		defer patchRead.Reset()

		mgr.setNumaNodes(&dev)
		nodes := mgr.getNumaNodesByPhyID(0)
		convey.So(len(nodes), convey.ShouldEqual, 1)
		convey.So(nodes[0], convey.ShouldEqual, int64(1))
	})
}

// TestGetNumaNodesByPhyID_NotFound test getNumaNodesByPhyID returns nil for unknown phyID
func TestGetNumaNodesByPhyID_NotFound(t *testing.T) {
	convey.Convey("should return nil for unknown phyID", t, func() {
		mgr := mockNumaNodeManager()
		convey.So(mgr.getNumaNodesByPhyID(999), convey.ShouldBeNil)
	})
}

// TestGetNumaNodesByPhyID_Found test getNumaNodesByPhyID returns cached value
func TestGetNumaNodesByPhyID_Found(t *testing.T) {
	convey.Convey("should return cached value for known phyID", t, func() {
		mgr := mockNumaNodeManager()
		mgr.phyIDToNumaNodesMu.Lock()
		mgr.phyIDToNumaNodes[0] = []int64{0, 1}
		mgr.phyIDToNumaNodesMu.Unlock()

		nodes := mgr.getNumaNodesByPhyID(0)
		convey.So(len(nodes), convey.ShouldEqual, 2)
		convey.So(nodes[0], convey.ShouldEqual, int64(0))
		convey.So(nodes[1], convey.ShouldEqual, int64(1))
	})
}
