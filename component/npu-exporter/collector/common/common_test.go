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

// Package common for general constants
package common

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/devmanager/common"
	"ascend-common/devmanager/hccn"
)

// TestGetNpuDevNetPortInfos test getNpuDevNetPortInfos function
func TestGetNpuDevNetPortInfos(t *testing.T) {
	convey.Convey("TestGetNpuDevNetPortInfos", t, func() {
		// Use existing mock
		n := mockNewNpuCollector()

		// Setup mocks for success case
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		// Mock device list with one device
		patches.ApplyMethodReturn(n.Dmgr, "GetDeviceList", int32(0), []int32{0}, nil)
		// Mock port info
		patches.ApplyFuncReturn(hccn.GetNpuDevNetPortInfo, map[int][]common.NpuDevPortInfo{
			0: {{PortID: 0, PortType: "UB", LinkStatus: "UP"}},
		}, nil)

		// Test function
		err := getNpuDevNetPortInfos(n)
		convey.So(err, convey.ShouldBeNil)
	})
}

// test constants for NpuDevPortsInfo tests
const (
	testLogicID0 = int32(0)
	testLogicID1 = int32(1)
	testDieID0   = 0
	testDieID1   = 1
)

// portIDsOf extracts the PortID of each port in order, for assertions
func portIDsOf(ports []common.NpuDevPortInfo) []int {
	ids := make([]int, 0, len(ports))
	for _, port := range ports {
		ids = append(ids, port.PortID)
	}
	return ids
}

// TestSetPortMap tests SetPortMap stores ports per logicID and sorts by PortID
func TestSetPortMap(t *testing.T) {
	convey.Convey("TestSetPortMap", t, func() {
		const (
			missingLogicID          = int32(2)
			expectedDiesForLogicID0 = 1
		)
		// 乱序输入用于验证排序，期望输出按 PortID 升序排列
		sortedDie0PortIDs := []int{1, 2, 3}
		sortedDie1PortIDs := []int{4, 5}

		infos := NpuDevPortsInfo{}
		infos.SetPortMap(testLogicID1, map[int][]common.NpuDevPortInfo{
			testDieID0: {{PortID: 3}, {PortID: 1}, {PortID: 2}},
		})
		infos.SetPortMap(testLogicID0, map[int][]common.NpuDevPortInfo{
			testDieID1: {{PortID: 5}, {PortID: 4}},
		})

		portMap0 := infos.GetPortMap(testLogicID0)
		convey.So(len(portMap0), convey.ShouldEqual, expectedDiesForLogicID0)
		convey.So(portIDsOf(portMap0[testDieID1]), convey.ShouldResemble, sortedDie1PortIDs)

		portMap1 := infos.GetPortMap(testLogicID1)
		convey.So(portIDsOf(portMap1[testDieID0]), convey.ShouldResemble, sortedDie0PortIDs)

		// 不存在的 logicID 返回 nil
		convey.So(infos.GetPortMap(missingLogicID), convey.ShouldBeNil)
	})
}

// TestInitAndGetCount tests total port count is summed across multiple logicIDs
func TestInitAndGetCount(t *testing.T) {
	convey.Convey("TestInitAndGetCount", t, func() {
		// logicID0 共 3 个端口（die0 2 个 + die1 1 个），logicID1 共 1 个端口
		const expectedTotalPorts = 4

		infos := NpuDevPortsInfo{}
		infos.SetPortMap(testLogicID0, map[int][]common.NpuDevPortInfo{
			testDieID0: {{PortID: 0}, {PortID: 1}},
			testDieID1: {{PortID: 0}},
		})
		infos.SetPortMap(testLogicID1, map[int][]common.NpuDevPortInfo{
			testDieID0: {{PortID: 0}},
		})
		infos.Init()
		convey.So(infos.GetCount(), convey.ShouldEqual, expectedTotalPorts)
	})
}

// TestGetMergedPortMap tests merging dieID->ports across all logicIDs
func TestGetMergedPortMap(t *testing.T) {
	convey.Convey("TestGetMergedPortMap", t, func() {
		convey.Convey("merge same-structured chips across logicIDs", func() {
			const (
				expectedDieCount      = 2
				expectedDie0PortCount = 2
				expectedDie1PortCount = 1
			)
			portMap := map[int][]common.NpuDevPortInfo{
				testDieID0: {{PortID: 0}, {PortID: 1}},
				testDieID1: {{PortID: 0}},
			}
			infos := NpuDevPortsInfo{}
			infos.SetPortMap(testLogicID0, portMap)
			infos.SetPortMap(testLogicID1, portMap)

			merged := infos.GetMergedPortMap()
			convey.So(len(merged), convey.ShouldEqual, expectedDieCount)
			convey.So(len(merged[testDieID0]), convey.ShouldEqual, expectedDie0PortCount)
			convey.So(len(merged[testDieID1]), convey.ShouldEqual, expectedDie1PortCount)
		})

		convey.Convey("returns empty map when no port map set", func() {
			infos := NpuDevPortsInfo{}
			merged := infos.GetMergedPortMap()
			convey.So(merged, convey.ShouldBeEmpty)
		})
	})
}
