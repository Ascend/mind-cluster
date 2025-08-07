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

// Package policy is used for processing superpod information
package policy

import (
	"sort"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-faultdiag-online/pkg/algo_src/netfault/algo"
)

func sortStringMap(m map[string][]string) map[string][]string {
	result := make(map[string][]string)
	for k, v := range m {
		temp := append([]string{}, v...)
		sort.Strings(temp)
		result[k] = temp
	}
	return result
}

func TestParseNodeDevice(t *testing.T) {
	convey.Convey("Test ParseNodeDevice", t, func() {
		convey.Convey("should return nil When input invalid", func() {
			NodeDeviceMap := make(map[string]*NodeDevice)
			fullMeshInfo, npuNetPlaneLink := parseNodeDeviceMap(NodeDeviceMap)
			convey.So(fullMeshInfo, convey.ShouldBeNil)
			convey.So(npuNetPlaneLink, convey.ShouldBeNil)
			fullMeshInfo, npuNetPlaneLink = parseNodeDeviceMap(nil)
			convey.So(fullMeshInfo, convey.ShouldBeNil)
			convey.So(npuNetPlaneLink, convey.ShouldBeNil)
		})

		nodeDeviceMap := map[string]*NodeDevice{
			"node-1": {DeviceMap: map[string]string{"0": "1", "1": "2"}, ServerID: "0"},
			"node-2": {DeviceMap: map[string]string{"0": "3", "1": "4"}, ServerID: "1"},
		}
		convey.Convey("should return correct when input valid", func() {
			_, npuNetPlaneInfo := parseNodeDeviceMap(nodeDeviceMap)
			expectNpuNetPlaneInfo := map[string][]string{
				"netplane_0": {
					"L2:0#Work-0:0#work-0.NSlot-0:0#NPU0-1:0",
					"L2:0#Work-0:0#work-0.NSlot-0:0#NPU1-2:0",
					"L2:0#Work-1:0#work-1.NSlot-0:0#NPU0-3:0",
					"L2:0#Work-1:0#work-1.NSlot-0:0#NPU1-4:0",
				},
			}
			convey.So(sortStringMap(npuNetPlaneInfo), convey.ShouldResemble, sortStringMap(expectNpuNetPlaneInfo))
		})
	})
}

func TestGetCurWorkInfo(t *testing.T) {
	convey.Convey("Test GetCurWorkInfo", t, func() {
		convey.Convey("should return nil when npu npuNetPlane nil", func() {
			npuNetPlaneInfo := getCurWorkInfo(nil, nil)
			convey.So(npuNetPlaneInfo, convey.ShouldBeNil)
		})
		convey.Convey("should return nil when DeviceMap doesn't exist", func() {
			npuNetPlaneInfo := make(map[string][]string)
			npuNetPlaneInfo = getCurWorkInfo(npuNetPlaneInfo, nil)
			convey.So(npuNetPlaneInfo, convey.ShouldBeNil)
		})
		convey.Convey("should return nil when ServiceId doesn't exist", func() {
			workInfo := &NodeDevice{
				DeviceMap: map[string]string{},
			}
			npuNetPlaneInfo := getCurWorkInfo(map[string][]string{}, workInfo)
			convey.So(npuNetPlaneInfo, convey.ShouldBeNil)
		})

		convey.Convey("should return nil when ServiceID format error", func() {
			workInfo := &NodeDevice{
				DeviceMap: map[string]string{},
				ServerID:  "0",
			}
			npuNetPlaneInfo := getCurWorkInfo(map[string][]string{}, workInfo)
			convey.So(npuNetPlaneInfo, convey.ShouldBeNil)
		})

		convey.Convey("should return correct result when valid", func() {
			workInfo := &NodeDevice{
				DeviceMap: map[string]string{
					"0": "1",
					"1": "2",
				},
				ServerID: "1",
			}
			exepectNetPlaneInfo := map[string][]string{
				"netplane_0": {
					"L2:0#Work-1:0#work-1.NSlot-0:0#NPU0-1:0",
					"L2:0#Work-1:0#work-1.NSlot-0:0#NPU1-2:0",
				},
			}
			npuNetPlaneInfo := getCurWorkInfo(map[string][]string{}, workInfo)
			convey.So(sortStringMap(npuNetPlaneInfo), convey.ShouldResemble, sortStringMap(exepectNetPlaneInfo))
		})
	})
}

func TestExtractNPUMapA3(t *testing.T) {
	convey.Convey("Test ExtractNPUMapA3", t, func() {

		convey.Convey("should return nil when input invalid", func() {
			npuNetPlaneInfo := map[string][]string{
				"netplane": {
					"L2:0#Work-0:0#work-0.NSlot-0:0#NPU0-1:0",
				},
			}
			NPUInfo := ExtractNPUMapA3(npuNetPlaneInfo)
			convey.So(NPUInfo, convey.ShouldBeNil)
		})

		convey.Convey("", func() {
			npuNetPlaneInfo := map[string][]string{
				"netplane_0": {
					"L2:0#Work-0:0#work-0.NSlot-0:0#NPU0-100:0",
					"L2:0#Work-1:0#work-1.NSlot-0:0#NPU1-200:0",
				},
			}
			NPUInfo := ExtractNPUMapA3(npuNetPlaneInfo)
			expectNpuInfo := map[string]algo.NpuInfo{
				"100": {RackName: "Work-0", SlotName: "work-0.NSlot-0", NpuNumber: 0, NetPlaneId: "netplane_0", IP: "100"},
				"200": {RackName: "Work-1", SlotName: "work-1.NSlot-0", NpuNumber: 1, NetPlaneId: "netplane_0", IP: "200"},
			}
			convey.So(NPUInfo, convey.ShouldResemble, expectNpuInfo)

		})
	})
}

func TestExtractNetStrInfo(t *testing.T) {
	convey.Convey("Test extractNetStrInfo", t, func() {
		convey.Convey("When input is valid", func() {
			netStr := "L2:0#Work-0:0#work-0.Netplane-0:0#NPU0-1:0"
			expectedNpuInfo := algo.NpuInfo{
				RackName:   "Work-0",
				SlotName:   "work-0.Netplane-0",
				NetPlaneId: "netplane_0",
				IP:         "1",
				NpuNumber:  0,
			}
			sdIdStr, npuInfo := extractNetStrInfo(netStr)
			convey.So(sdIdStr, convey.ShouldEqual, "1")
			convey.So(npuInfo, convey.ShouldResemble, expectedNpuInfo)
		})

		convey.Convey("When first regex has insufficient matches", func() {
			netStr := "#rack01:#slot02" // 缺少第三个匹配项
			sdIdStr, _ := extractNetStrInfo(netStr)
			convey.So(sdIdStr, convey.ShouldEqual, "")

		})

		convey.Convey("When second regex fails to match", func() {
			netStr := "#rack01:#slot02:#invalid-npu-id"
			sdIdStr, _ := extractNetStrInfo(netStr)
			convey.So(sdIdStr, convey.ShouldEqual, "")
		})

		convey.Convey("When sdIdStr is not a number", func() {
			netStr := "#rack01:#slot02:#npu123-NPU100-ABC"
			sdIdStr, _ := extractNetStrInfo(netStr)
			convey.So(sdIdStr, convey.ShouldEqual, "")
		})

		convey.Convey("When input is completely invalid", func() {
			netStr := "invalid_string"
			sdIdStr, _ := extractNetStrInfo(netStr)
			convey.So(sdIdStr, convey.ShouldEqual, "")
		})
	})
}
