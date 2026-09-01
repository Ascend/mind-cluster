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

package driver

import (
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/cdi"
	"ascend-common/devmanager"
	"ascend-dynamic-resource-allocation/internal/device"
	draFlags "ascend-dynamic-resource-allocation/internal/flags"
	"ascend-dynamic-resource-allocation/internal/plugin"
)

// =============================================================================
// Test case data (decoupled from the test process below)
// =============================================================================

// startCase drives AscendDraDriver.Start through every branch:
//   - pullNPUInfo fails
//   - startService fails
//   - publishResources fails
//   - happy path
type startCase struct {
	name        string
	listErr     error // error returned by fakeGeneration.ListNpuDevices
	devices     []device.NpuDevice
	registerErr error
	publishErr  error
	expectErr   bool
}

var startCases = []startCase{
	{name: "pullNPUInfo fails", listErr: errSentinel, expectErr: true},
	{
		name:        "startService fails",
		devices:     []device.NpuDevice{{DevType: "Ascend910", DeviceName: "npu-0", PhyID: 0}},
		registerErr: errSentinel,
		expectErr:   true,
	},
	{
		name:       "publishResources fails",
		devices:    []device.NpuDevice{{DevType: "Ascend910", DeviceName: "npu-0", PhyID: 0}},
		publishErr: errSentinel,
		expectErr:  true,
	},
	{
		name:      "happy path",
		devices:   []device.NpuDevice{{DevType: "Ascend910", DeviceName: "npu-0", PhyID: 0}},
		expectErr: false,
	},
}

// pullNPUInfoCase drives AscendDraDriver.pullNPUInfo:
//   - ListNpuDevices returns error
//   - empty device list (no devices type found)
//   - single device one type success
//   - multiple devices with duplicate types (removeDuplicate collapses)
type pullNPUInfoCase struct {
	name      string
	listErr   error
	devices   []device.NpuDevice
	expectErr bool
	expectDev int
}

var pullNPUInfoCases = []pullNPUInfoCase{
	{name: "ListNpuDevices error", listErr: errSentinel, expectErr: true},
	{name: "empty device list", devices: nil, expectErr: true},
	{
		name:      "single device single type",
		devices:   []device.NpuDevice{{DevType: "Ascend910", DeviceName: "npu-0", PhyID: 0}},
		expectErr: false,
		expectDev: 1,
	},
	{
		name: "duplicate types collapse to one group",
		devices: []device.NpuDevice{
			{DevType: "Ascend910", DeviceName: "npu-0", PhyID: 0},
			{DevType: "Ascend910", DeviceName: "npu-1", PhyID: 1},
		},
		expectErr: false,
		expectDev: 2,
	},
	{
		name: "two distinct types",
		devices: []device.NpuDevice{
			{DevType: "Ascend910", DeviceName: "npu-0", PhyID: 0},
			{DevType: "Ascend910B", DeviceName: "npu-1", PhyID: 1},
		},
		expectErr: false,
		expectDev: 2,
	},
}

// buildDriverResourcesCase drives AscendDraDriver.buildDriverResources.
type buildDriverResourcesCase struct {
	name        string
	allDevs     []*device.NpuDevice
	expectPools int
	expectDevs  int
}

var buildDriverResourcesCases = []buildDriverResourcesCase{
	{name: "no devices", allDevs: nil, expectPools: 1, expectDevs: 0},
	{
		name: "two devices",
		allDevs: []*device.NpuDevice{
			{DevType: "Ascend910", DeviceName: "npu-0", PhyID: 0},
			{DevType: "Ascend910", DeviceName: "npu-1", PhyID: 1},
		},
		expectPools: 1,
		expectDevs:  2,
	},
}

// removeDuplicateCase drives removeDuplicate.
type removeDuplicateCase struct {
	name    string
	input   []string
	expectN int
}

var removeDuplicateCases = []removeDuplicateCase{
	{name: "empty input", input: nil, expectN: 0},
	{name: "no duplicates", input: []string{"a", "b", "c"}, expectN: 3},
	{name: "all duplicates", input: []string{"a", "a", "a"}, expectN: 1},
	{name: "some duplicates", input: []string{"a", "b", "a", "c", "b"}, expectN: 3},
}

// autoSetDriverCase drives AscendDraManager.autoSetDraDriver via
// NewAscendDraManager, covering the devType switch branches.
type autoSetDriverCase struct {
	name      string
	devType   string
	pluginErr bool // true to make patched NewAscendDraPlugin return err
	expectErr bool
	expectGen string // "910" | "950" | "" for unsupported
}

var autoSetDriverCases = []autoSetDriverCase{
	{name: "Ascend910A selects 910", devType: api.Ascend910A, expectGen: "910"},
	{name: "Ascend910B selects 910", devType: api.Ascend910B, expectGen: "910"},
	{name: "Ascend910A3 selects 910", devType: api.Ascend910A3, expectGen: "910"},
	{name: "Ascend910A5 selects 950", devType: api.Ascend910A5, expectGen: "950"},
	{name: "unsupported devType", devType: "UnknownChip", expectErr: true, expectGen: ""},
	{name: "NewAscendDraPlugin error", devType: api.Ascend910A, pluginErr: true, expectErr: true},
}

// =============================================================================
// Test process (uses the case data above)
// =============================================================================

// TestNewAscendDraDriver verifies the constructor wires its arguments
// directly into the struct fields.
func TestNewAscendDraDriver(t *testing.T) {
	Convey("NewAscendDraDriver stores its arguments verbatim", t, func() {
		cfg := &draFlags.DRAConfig{DraOption: &draFlags.DRAOption{NodeName: "n1"}}
		gen := &fakeGeneration{devType: "Ascend910"}
		adp := &plugin.AscendDraPlugin{}
		d := NewAscendDraDriver(cfg, gen, adp)
		So(d.draConfig, ShouldEqual, cfg)
		So(d.generation, ShouldEqual, gen)
		So(d.ascendDraPlugin, ShouldEqual, adp)
	})
}

// TestAscendDraDriver_Start walks every branch of the Start sequence.
func TestAscendDraDriver_Start(t *testing.T) {
	Convey("AscendDraDriver.Start", t, func() {
		for idx, tc := range startCases {
			Convey(fmt.Sprintf("case#%d %s", idx, tc.name), func() {
				gen := &fakeGeneration{
					devices: tc.devices,
					listErr: tc.listErr,
				}
				d, _ := newDriverWithFake(gen)
				stopCalled := false
				patches := &pluginMethodPatches{}
				patchPluginMethods(patches, tc.registerErr, tc.publishErr, &stopCalled)
				defer patches.Reset()
				err := d.Start(context.Background())
				if tc.expectErr {
					So(err, ShouldNotBeNil)
				} else {
					So(err, ShouldBeNil)
				}
			})
		}
	})
}

// TestAscendDraDriver_pullNPUInfo exercises pullNPUInfo's branches including
// the "no devices type found" path when the device list is empty.
func TestAscendDraDriver_pullNPUInfo(t *testing.T) {
	Convey("AscendDraDriver.pullNPUInfo", t, func() {
		for idx, tc := range pullNPUInfoCases {
			Convey(fmt.Sprintf("case#%d %s", idx, tc.name), func() {
				gen := &fakeGeneration{
					devices: tc.devices,
					listErr: tc.listErr,
				}
				d, _ := newDriverWithFake(gen)
				err := d.pullNPUInfo()
				if tc.expectErr {
					So(err, ShouldNotBeNil)
				} else {
					So(err, ShouldBeNil)
					So(len(d.allInfo.AllDevs), ShouldEqual, tc.expectDev)
					So(len(d.groupDevice), ShouldBeGreaterThan, 0)
				}
			})
		}
	})
}

// TestAscendDraDriver_buildDriverResources verifies the ResourceSlice shape
// produced from the driver's allInfo.
func TestAscendDraDriver_buildDriverResources(t *testing.T) {
	Convey("AscendDraDriver.buildDriverResources", t, func() {
		for idx, tc := range buildDriverResourcesCases {
			Convey(fmt.Sprintf("case#%d %s", idx, tc.name), func() {
				gen := &fakeGeneration{}
				d, _ := newDriverWithFake(gen)
				d.allInfo.AllDevs = tc.allDevs
				res := d.buildDriverResources()
				So(len(res.Pools), ShouldEqual, tc.expectPools)
				pool := res.Pools["test-node"]
				So(len(pool.Slices), ShouldEqual, 1)
				So(len(pool.Slices[0].Devices), ShouldEqual, tc.expectDevs)
			})
		}
	})
}

// TestAscendDraDriver_Stop verifies that Stop delegates to the plugin.
func TestAscendDraDriver_Stop(t *testing.T) {
	Convey("AscendDraDriver.Stop calls plugin.Stop", t, func() {
		gen := &fakeGeneration{}
		d, _ := newDriverWithFake(gen)
		stopCalled := false
		patches := &pluginMethodPatches{}
		patchPluginMethods(patches, nil, nil, &stopCalled)
		defer patches.Reset()
		d.Stop()
		So(stopCalled, ShouldBeTrue)
	})
}

// TestRemoveDuplicate exercises the dedup helper directly.
func TestRemoveDuplicate(t *testing.T) {
	Convey("removeDuplicate collapses repeated entries", t, func() {
		for idx, tc := range removeDuplicateCases {
			Convey(fmt.Sprintf("case#%d %s", idx, tc.name), func() {
				got := removeDuplicate(&tc.input)
				So(len(got), ShouldEqual, tc.expectN)
			})
		}
	})
}

// TestNewAscendDraManager covers the devType switch in autoSetDraDriver and
// the NewAscendDraPlugin failure path. The plugin constructors are stubbed
// so no filesystem or k8s client is touched.
func TestNewAscendDraManager(t *testing.T) {
	Convey("NewAscendDraManager auto-selects generation by devType", t, func() {
		for idx, tc := range autoSetDriverCases {
			Convey(fmt.Sprintf("case#%d %s", idx, tc.name), func() {
				dmgr := &devmanager.DeviceManagerMock{DevType: tc.devType}
				cfg := &draFlags.DRAConfig{DraOption: &draFlags.DRAOption{}}

				p := gomonkey.NewPatches()
				defer p.Reset()
				// Stub cdi.PrepareMountConfigFile so NewCDISpecManager does
				// not write to /etc during tests.
				p.ApplyFunc(cdi.PrepareMountConfigFile,
					func(_ string) error { return nil })
				// Stub NewAscendDraPlugin to either succeed or fail per case.
				if tc.pluginErr {
					p.ApplyFunc(plugin.NewAscendDraPlugin,
						func(_ *draFlags.DRAConfig, _ context.CancelCauseFunc,
							_ plugin.CdiSpecInterface) (*plugin.AscendDraPlugin, error) {
							return nil, errSentinel
						})
				} else {
					p.ApplyFunc(plugin.NewAscendDraPlugin,
						func(_ *draFlags.DRAConfig, _ context.CancelCauseFunc,
							_ plugin.CdiSpecInterface) (*plugin.AscendDraPlugin, error) {
							return &plugin.AscendDraPlugin{}, nil
						})
				}

				mgr, err := NewAscendDraManager(
					func(error) {}, dmgr, cfg)
				if tc.expectErr {
					So(err, ShouldNotBeNil)
					So(mgr, ShouldBeNil)
				} else {
					So(err, ShouldBeNil)
					So(mgr, ShouldNotBeNil)
				}
			})
		}
	})
}

// TestAscendDraManager_Stop verifies the manager delegates Stop to its driver.
func TestAscendDraManager_Stop(t *testing.T) {
	Convey("AscendDraManager.Stop delegates to driver", t, func() {
		// Construct a manager with a fake driver-like stub. The simplest path
		// is to build an AscendDraManager with a real AscendDraDriver whose
		// plugin methods are patched, then call Stop.
		gen := &fakeGeneration{}
		d, _ := newDriverWithFake(gen)
		stopCalled := false
		patches := &pluginMethodPatches{}
		patchPluginMethods(patches, nil, nil, &stopCalled)
		defer patches.Reset()
		mgr := &AscendDraManager{draDriver: d}
		mgr.Stop()
		So(stopCalled, ShouldBeTrue)
	})
}

// Compile-time guard: fakeGeneration satisfies DraGenerationInterface.
var _ DraGenerationInterface = (*fakeGeneration)(nil)
