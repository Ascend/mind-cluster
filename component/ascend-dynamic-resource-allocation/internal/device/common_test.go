/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package device

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// =============================================================================
// Test case data (decoupled from the test process below)
// =============================================================================

// classifyDevicesCase describes one input/expectation pair for ClassifyDevices.
type classifyDevicesCase struct {
	name     string
	allDevs  []*NpuDevice
	devTypes []string
	// expectCounts maps each devType to the expected number of classified
	// devices. A missing key means the key must not appear in the result.
	expectCounts map[string]int
}

// classifyDevicesCases enumerates the branches of ClassifyDevices:
//   - nil/empty devTypes (loop body never runs, empty map)
//   - devTypes present but allDevs empty/nil (empty slice per type)
//   - exact match per type
//   - devType with no matching device
//   - multiple types with mixed membership
//   - device whose type is not listed in devTypes
var classifyDevicesCases = []classifyDevicesCase{
	{
		name:         "test scenario where devs and types are nil, result is an empty map",
		allDevs:      nil,
		devTypes:     nil,
		expectCounts: map[string]int{},
	},
	{
		name:         "test scenario where dev slice is empty with one type, result is count 0 for that type",
		allDevs:      []*NpuDevice{},
		devTypes:     []string{"Ascend910"},
		expectCounts: map[string]int{"Ascend910": 0},
	},
	{
		name: "test scenario with single type single match, result is count 1 for that type",
		allDevs: []*NpuDevice{
			{DevType: "Ascend910", PhyID: 0},
		},
		devTypes:     []string{"Ascend910"},
		expectCounts: map[string]int{"Ascend910": 1},
	},
	{
		name: "test scenario with single type no match, result is count 0 for that type",
		allDevs: []*NpuDevice{
			{DevType: "Ascend910B", PhyID: 1},
		},
		devTypes:     []string{"Ascend910"},
		expectCounts: map[string]int{"Ascend910": 0},
	},
	{
		name: "test scenario with two types mixed membership, result is counts 2 and 1 per type",
		allDevs: []*NpuDevice{
			{DevType: "Ascend910", PhyID: 0},
			{DevType: "Ascend910B", PhyID: 2},
			{DevType: "Ascend910", PhyID: 3},
		},
		devTypes:     []string{"Ascend910", "Ascend910B"},
		expectCounts: map[string]int{"Ascend910": 2, "Ascend910B": 1},
	},
	{
		name: "test scenario where device type is absent from devTypes list, result counts only listed types",
		allDevs: []*NpuDevice{
			{DevType: "Ascend910", PhyID: 0},
			{DevType: "Ascend950", PhyID: 9},
		},
		devTypes:     []string{"Ascend910"},
		expectCounts: map[string]int{"Ascend910": 1},
	},
}

// classifyDevByTypeCase describes one input/expectation pair for the private
// classifyDevByType helper. It is exercised white-box below.
type classifyDevByTypeCase struct {
	name    string
	allDevs []*NpuDevice
	suffix  string
	expect  int
}

// classifyDevByTypeCases covers classifyDevByType branches:
//   - nil allDevs
//   - none match the suffix
//   - one match
//   - partial match (some devices match)
//   - all match
var classifyDevByTypeCases = []classifyDevByTypeCase{
	{name: "nil allDevs", allDevs: nil, suffix: "Ascend910", expect: 0},
	{
		name:    "none match suffix",
		allDevs: []*NpuDevice{{DevType: "Ascend910B", PhyID: 1}},
		suffix:  "Ascend910",
		expect:  0,
	},
	{
		name:    "one match",
		allDevs: []*NpuDevice{{DevType: "Ascend910", PhyID: 0}},
		suffix:  "Ascend910",
		expect:  1,
	},
	{
		name: "partial match",
		allDevs: []*NpuDevice{
			{DevType: "Ascend910", PhyID: 0},
			{DevType: "Ascend910B", PhyID: 2},
			{DevType: "Ascend910", PhyID: 3},
		},
		suffix: "Ascend910",
		expect: 2,
	},
	{
		name: "all match",
		allDevs: []*NpuDevice{
			{DevType: "Ascend910", PhyID: 0},
			{DevType: "Ascend910", PhyID: 1},
		},
		suffix: "Ascend910",
		expect: 2,
	},
}

// =============================================================================
// Test process (uses the case data above)
// =============================================================================

// TestClassifyDevices walks the data-driven cases above and asserts the
// resulting per-type counts.
func TestClassifyDevices(t *testing.T) {
	Convey("ClassifyDevices groups devices by requested type", t, func() {
		for idx, tc := range classifyDevicesCases {
			Convey(fmt.Sprintf("case#%d %s", idx, tc.name), func() {
				got := ClassifyDevices(tc.allDevs, tc.devTypes)
				Convey("result map has the expected number of keys", func() {
					So(len(got), ShouldEqual, len(tc.expectCounts))
				})
				for devType, cnt := range tc.expectCounts {
					devType := devType
					cnt := cnt
					Convey(fmt.Sprintf("type %q has %d devices", devType, cnt), func() {
						So(len(got[devType]), ShouldEqual, cnt)
					})
				}
			})
		}
	})
}

// TestClassifyDevByType exercises the private helper directly, covering the
// loop branch where DevType matches the suffix and where it does not.
func TestClassifyDevByType(t *testing.T) {
	Convey("classifyDevByType filters devices by suffix", t, func() {
		for idx, tc := range classifyDevByTypeCases {
			Convey(fmt.Sprintf("case#%d %s", idx, tc.name), func() {
				got := classifyDevByType(tc.allDevs, tc.suffix)
				So(len(got), ShouldEqual, tc.expect)
			})
		}
	})
}
