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
	"ascend-dynamic-resource-allocation/pkg/consts"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"

	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/utils/ptr"

	"ascend-common/api"
	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
)

// =============================================================================
// Test case data (decoupled from the test process below)
// =============================================================================

// listNpuDevicesCase is a data-driven case for ListNpuDevices on both 910 and
// 950 generations. The runner maps `scenario` to a specific gomonkey setup.
type listNpuDevicesCase struct {
	name      string
	scenario  string // "ok", "getDeviceListErr", "buildNpuDeviceErr"
	devNum    int32
	devList   []int32
	expectLen int
	expectErr bool
}

// ascend910ListCases covers ListNpuDevices branches for 910:
//   - GetDeviceList error
//   - zero devices (empty result, no error)
//   - happy path with several devices
//   - buildNpuDevice error mid-loop
var ascend910ListCases = []listNpuDevicesCase{
	{name: "GetDeviceList returns error", scenario: "getDeviceListErr", expectErr: true},
	{name: "zero devices", scenario: "ok", devNum: 0, devList: nil, expectLen: 0},
	{
		name:      "two devices happy path",
		scenario:  "ok",
		devNum:    2,
		devList:   []int32{0, 1},
		expectLen: 2,
	},
	{
		name:      "buildNpuDevice error mid-loop",
		scenario:  "buildNpuDeviceErr",
		devNum:    1,
		devList:   []int32{0},
		expectErr: true,
	},
}

// ascend950ListCases mirrors ascend910ListCases for the 950 generation.
var ascend950ListCases = []listNpuDevicesCase{
	{name: "GetDeviceList returns error", scenario: "getDeviceListErr", expectErr: true},
	{name: "zero devices", scenario: "ok", devNum: 0, devList: nil, expectLen: 0},
	{
		name:      "one device happy path",
		scenario:  "ok",
		devNum:    1,
		devList:   []int32{0},
		expectLen: 1,
	},
	{
		name:      "buildNpuDevice error mid-loop",
		scenario:  "buildNpuDeviceErr",
		devNum:    1,
		devList:   []int32{0},
		expectErr: true,
	},
}

// buildNpuDevice910Case covers the private 910 buildNpuDevice, which also
// exercises getDeviceIP branches through its success path.
type buildNpuDevice910Case struct {
	name      string
	scenario  string // "phyErr","cardErr","ipv4Ok","ipv4ErrIpv6Ok","ipv6LinkLocal","ipv6Err"
	logicID   int32
	expectIP  string
	expectErr bool
}

// buildNpu910Cases covers every branch of buildNpuDevice and getDeviceIP:
//   - GetPhysicIDFromLogicID error
//   - GetCardIDDeviceID error
//   - IPv4 path returns an address
//   - IPv4 error, IPv6 fallback returns an address
//   - IPv6 link-local address rejected (getDeviceIP errors, buildNpuDevice sets "")
//   - both IPv4 and IPv6 error (getDeviceIP errors, buildNpuDevice sets "")
var buildNpu910Cases = []buildNpuDevice910Case{
	{name: "GetPhysicIDFromLogicID error", scenario: "phyErr", expectErr: true},
	{name: "GetCardIDDeviceID error", scenario: "cardErr", expectErr: true},
	{name: "IPv4 address returned", scenario: "ipv4Ok", logicID: 0, expectIP: "192.168.0.1"},
	{name: "IPv4 error IPv6 fallback ok", scenario: "ipv4ErrIpv6Ok", logicID: 0, expectIP: "2001:db8::1"},
	{name: "IPv6 link-local rejected", scenario: "ipv6LinkLocal", logicID: 0, expectIP: ""},
	{name: "both IPv4 and IPv6 error", scenario: "ipv6Err", logicID: 0, expectIP: ""},
}

// getDeviceIP910Case covers getDeviceIP directly (white-box).
type getDeviceIP910Case struct {
	name      string
	scenario  string // "ipv4Ok","ipv4ErrIpv6Ok","ipv6LinkLocal","ipv6Err"
	logicID   int32
	expectIP  string
	expectErr bool
}

var getDeviceIP910Cases = []getDeviceIP910Case{
	{name: "IPv4 ok", scenario: "ipv4Ok", logicID: 0, expectIP: "192.168.0.1"},
	{name: "IPv4 error IPv6 ok", scenario: "ipv4ErrIpv6Ok", logicID: 0, expectIP: "2001:db8::1"},
	{name: "IPv6 link-local returns error", scenario: "ipv6LinkLocal", logicID: 0, expectErr: true},
	{name: "both IPv4 and IPv6 error", scenario: "ipv6Err", logicID: 0, expectErr: true},
}

// buildNpuDevice950Case covers the private 950 buildNpuDevice.
type buildNpuDevice950Case struct {
	name      string
	scenario  string // "phyErr","ok"
	logicID   int32
	expectErr bool
}

var buildNpu950Cases = []buildNpuDevice950Case{
	{name: "GetPhysicIDFromLogicID error", scenario: "phyErr", expectErr: true},
	{name: "happy path", scenario: "ok", logicID: 3, expectErr: false},
}

// deviceAttrCase covers DeviceAttributes for both generations.
type deviceAttrCase struct {
	name string
	dev  NpuDevice
}

var deviceAttrCases = []deviceAttrCase{
	{name: "populated device", dev: NpuDevice{DevType: "Ascend910", PhyID: 7, LogicID: 3}},
	{name: "zero-value device", dev: NpuDevice{}},
}

// commonGenCase covers AscendCommonGeneration SetDmgr/GetDevType/GetProductTypes.
type commonGenCase struct {
	name           string
	devType        string
	productTypes   []string
	expectDevType  string
	expectProducts int
}

var commonGenCases = []commonGenCase{
	{name: "910 type", devType: "Ascend910", productTypes: []string{"Ascend910"}, expectDevType: "Ascend910", expectProducts: 1},
	{name: "950 type", devType: "Ascend950", productTypes: []string{"Ascend950"}, expectDevType: "Ascend950", expectProducts: 1},
	{name: "empty product types", devType: "Ascend910", productTypes: nil, expectDevType: "Ascend910", expectProducts: 0},
}

// =============================================================================
// Shared gomonkey helpers (the "process" half)
// =============================================================================

var dmmType = reflect.TypeOf(&devmanager.DeviceManagerMock{})

// patchGetDeviceList stubs DeviceManagerMock.GetDeviceList.
func patchGetDeviceList(p *gomonkey.Patches, devNum int32, devList []int32, err error) {
	p.ApplyMethod(dmmType, "GetDeviceList",
		func(_ *devmanager.DeviceManagerMock) (int32, []int32, error) {
			return devNum, devList, err
		})
}

// patchGetProductTypeArray stubs DeviceManagerMock.GetProductTypeArray.
func patchGetProductTypeArray(p *gomonkey.Patches, products []string) {
	p.ApplyMethod(dmmType, "GetProductTypeArray",
		func(_ *devmanager.DeviceManagerMock) []string { return products })
}

// patchGetChipInfoErr stubs DeviceManagerMock.GetChipInfo to fail.
func patchGetChipInfoErr(p *gomonkey.Patches) {
	p.ApplyMethod(dmmType, "GetChipInfo",
		func(_ *devmanager.DeviceManagerMock, _ int32) (*common.ChipInfo, error) {
			return nil, errSentinel
		})
}

// patchGetDevType stubs DeviceManagerMock.GetDevType.
func patchGetDevType(p *gomonkey.Patches, devType string) {
	p.ApplyMethod(dmmType, "GetDevType",
		func(_ *devmanager.DeviceManagerMock) string { return devType })
}

// patch910IPScenario stubs GetDeviceIPAddress for the named IP scenario. The
// double dispatches on ipType so a single stub serves both the v4 and v6
// fallback call paths that getDeviceIP performs.
func patch910IPScenario(p *gomonkey.Patches, scenario string) {
	p.ApplyMethod(dmmType, "GetDeviceIPAddress",
		func(_ *devmanager.DeviceManagerMock, _ int32, ipType int32) (string, error) {
			switch scenario {
			case "ipv4Ok":
				return "192.168.0.1", nil
			case "ipv4ErrIpv6Ok":
				if ipType == ipAddrTypeV4 {
					return "", errSentinel
				}
				return "2001:db8::1", nil
			case "ipv6LinkLocal":
				if ipType == ipAddrTypeV4 {
					return "", errSentinel
				}
				return "fe80::1", nil
			case "ipv6Err":
				return "", errSentinel
			}
			return "", nil
		})
}

// setup910BuildScenario applies the gomonkey stubs needed for one 910
// buildNpuDevice branch, identified by scenario. It is the bridge between the
// data-driven cases (buildNpu910Cases) and the actual test execution.
func setup910BuildScenario(p *gomonkey.Patches, scenario string) {
	phy := func(_ *devmanager.DeviceManagerMock, _ int32) (int32, error) { return 1, nil }
	card := func(_ *devmanager.DeviceManagerMock, _ int32) (int32, int32, error) { return 0, 0, nil }
	switch scenario {
	case "phyErr":
		p.ApplyMethod(dmmType, "GetPhysicIDFromLogicID",
			func(_ *devmanager.DeviceManagerMock, _ int32) (int32, error) { return 0, errSentinel })
		return
	case "cardErr":
		p.ApplyMethod(dmmType, "GetPhysicIDFromLogicID", phy)
		p.ApplyMethod(dmmType, "GetCardIDDeviceID",
			func(_ *devmanager.DeviceManagerMock, _ int32) (int32, int32, error) { return 0, 0, errSentinel })
		return
	}
	// success-oriented branches: phy + card succeed, IP behaviour varies.
	p.ApplyMethod(dmmType, "GetPhysicIDFromLogicID", phy)
	p.ApplyMethod(dmmType, "GetCardIDDeviceID", card)
	patch910IPScenario(p, scenario)
}

// =============================================================================
// Test process (uses the case data above)
// =============================================================================

func TestAscendCommonGeneration(t *testing.T) {
	Convey("AscendCommonGeneration SetDmgr/GetDevType/GetProductTypes", t, func() {
		for idx, tc := range commonGenCases {
			Convey(fmt.Sprintf("case#%d %s", idx, tc.name), func() {
				g := NewAscend910Generation() // embeds AscendCommonGeneration
				mock := &devmanager.DeviceManagerMock{DevType: tc.devType}
				p := gomonkey.NewPatches()
				defer p.Reset()
				patchGetProductTypeArray(p, tc.productTypes)
				g.SetDmgr(mock)
				So(g.GetDevType(), ShouldEqual, tc.expectDevType)
				So(len(g.GetProductTypes()), ShouldEqual, tc.expectProducts)
			})
		}
	})
}

func TestAscend910Generation_ListNpuDevices(t *testing.T) {
	Convey("Ascend910 ListNpuDevices", t, func() {
		for idx, tc := range ascend910ListCases {
			Convey(fmt.Sprintf("case#%d %s", idx, tc.name), func() {
				g, _ := new910WithMock()
				p := gomonkey.NewPatches()
				defer p.Reset()
				switch tc.scenario {
				case "getDeviceListErr":
					patchGetDeviceList(p, 0, nil, errSentinel)
				case "ok":
					patchGetDeviceList(p, tc.devNum, tc.devList, nil)
					setup910BuildScenario(p, "ipv4Ok")
				case "buildNpuDeviceErr":
					patchGetDeviceList(p, tc.devNum, tc.devList, nil)
					setup910BuildScenario(p, "phyErr")
				}
				devs, err := g.ListNpuDevices()
				if tc.expectErr {
					So(err, ShouldNotBeNil)
					So(devs, ShouldBeNil)
				} else {
					So(err, ShouldBeNil)
					So(len(devs), ShouldEqual, tc.expectLen)
				}
			})
		}
	})
}

func TestAscend910Generation_buildNpuDevice(t *testing.T) {
	Convey("Ascend910 buildNpuDevice", t, func() {
		for idx, tc := range buildNpu910Cases {
			Convey(fmt.Sprintf("case#%d %s", idx, tc.name), func() {
				g, _ := new910WithMock()
				p := gomonkey.NewPatches()
				defer p.Reset()
				setup910BuildScenario(p, tc.scenario)
				dev, err := g.buildNpuDevice(tc.logicID)
				if tc.expectErr {
					So(err, ShouldNotBeNil)
					So(dev, ShouldResemble, NpuDevice{})
				} else {
					So(err, ShouldBeNil)
					So(dev.IP, ShouldEqual, tc.expectIP)
					So(dev.DevType, ShouldEqual, api.Ascend910)
					So(dev.LogicID, ShouldEqual, tc.logicID)
				}
			})
		}
	})
}

func TestAscend910Generation_getDeviceIP(t *testing.T) {
	Convey("Ascend910 getDeviceIP", t, func() {
		for idx, tc := range getDeviceIP910Cases {
			Convey(fmt.Sprintf("case#%d %s", idx, tc.name), func() {
				g, _ := new910WithMock()
				p := gomonkey.NewPatches()
				defer p.Reset()
				patch910IPScenario(p, tc.scenario)
				ip, err := g.getDeviceIP(tc.logicID)
				if tc.expectErr {
					So(err, ShouldNotBeNil)
					So(ip, ShouldEqual, "")
				} else {
					So(err, ShouldBeNil)
					So(ip, ShouldEqual, tc.expectIP)
				}
			})
		}
	})
}

func TestAscend910Generation_DeviceAttributes(t *testing.T) {
	Convey("Ascend910 DeviceAttributes reports deviceType, physicId and chipName", t, func() {
		g, _ := new910WithMock()
		for idx, tc := range deviceAttrCases {
			Convey(fmt.Sprintf("case#%d %s", idx, tc.name), func() {
				attrs := g.DeviceAttributes(tc.dev)
				So(len(attrs), ShouldEqual, 3)
				So(attrs[attrKeyType], ShouldResemble,
					resourceapi.DeviceAttribute{StringValue: ptr.To(consts.NPUNamePrefix)})
				So(attrs[attrKeyPhysicID], ShouldResemble,
					resourceapi.DeviceAttribute{IntValue: ptr.To(int64(tc.dev.PhyID))})
				So(attrs[attrKeyChipName], ShouldResemble,
					resourceapi.DeviceAttribute{StringValue: ptr.To(common.Chip910)})
			})
		}
		Convey("GetChipInfo error yields empty chipName", func() {
			p := gomonkey.NewPatches()
			defer p.Reset()
			patchGetChipInfoErr(p)
			attrs := g.DeviceAttributes(NpuDevice{LogicID: 1})
			So(attrs[attrKeyChipName], ShouldResemble,
				resourceapi.DeviceAttribute{StringValue: ptr.To("")})
		})
	})
}

func TestAscend950Generation_ListNpuDevices(t *testing.T) {
	Convey("Ascend950 ListNpuDevices", t, func() {
		for idx, tc := range ascend950ListCases {
			Convey(fmt.Sprintf("case#%d %s", idx, tc.name), func() {
				g, _ := new950WithMock()
				p := gomonkey.NewPatches()
				defer p.Reset()
				switch tc.scenario {
				case "getDeviceListErr":
					patchGetDeviceList(p, 0, nil, errSentinel)
				case "ok":
					patchGetDeviceList(p, tc.devNum, tc.devList, nil)
					p.ApplyMethod(dmmType, "GetPhysicIDFromLogicID",
						func(_ *devmanager.DeviceManagerMock, _ int32) (int32, error) { return 0, nil })
				case "buildNpuDeviceErr":
					patchGetDeviceList(p, tc.devNum, tc.devList, nil)
					p.ApplyMethod(dmmType, "GetPhysicIDFromLogicID",
						func(_ *devmanager.DeviceManagerMock, _ int32) (int32, error) { return 0, errSentinel })
				}
				devs, err := g.ListNpuDevices()
				if tc.expectErr {
					So(err, ShouldNotBeNil)
					So(devs, ShouldBeNil)
				} else {
					So(err, ShouldBeNil)
					So(len(devs), ShouldEqual, tc.expectLen)
				}
			})
		}
	})
}

func TestAscend950Generation_buildNpuDevice(t *testing.T) {
	Convey("Ascend950 buildNpuDevice", t, func() {
		for idx, tc := range buildNpu950Cases {
			Convey(fmt.Sprintf("case#%d %s", idx, tc.name), func() {
				g, _ := new950WithMock()
				p := gomonkey.NewPatches()
				defer p.Reset()
				switch tc.scenario {
				case "phyErr":
					p.ApplyMethod(dmmType, "GetPhysicIDFromLogicID",
						func(_ *devmanager.DeviceManagerMock, _ int32) (int32, error) { return 0, errSentinel })
				case "ok":
					p.ApplyMethod(dmmType, "GetPhysicIDFromLogicID",
						func(_ *devmanager.DeviceManagerMock, _ int32) (int32, error) { return 5, nil })
				}
				dev, err := g.buildNpuDevice(tc.logicID)
				if tc.expectErr {
					So(err, ShouldNotBeNil)
					So(dev, ShouldResemble, NpuDevice{})
				} else {
					So(err, ShouldBeNil)
					So(dev.PhyID, ShouldEqual, 5)
					So(dev.LogicID, ShouldEqual, tc.logicID)
					So(dev.DevType, ShouldEqual, api.Ascend910A5)
					So(dev.DeviceName, ShouldEqual, "npu-5")
				}
			})
		}
	})
}

func TestAscend950Generation_DeviceAttributes(t *testing.T) {
	Convey("Ascend950 DeviceAttributes reports deviceType, physicId and chipName", t, func() {
		g, _ := new950WithMock()
		for idx, tc := range deviceAttrCases {
			Convey(fmt.Sprintf("case#%d %s", idx, tc.name), func() {
				attrs := g.DeviceAttributes(tc.dev)
				So(len(attrs), ShouldEqual, 3)
				So(attrs[attrKeyType], ShouldResemble,
					resourceapi.DeviceAttribute{StringValue: ptr.To(consts.NPUNamePrefix)})
				So(attrs[attrKeyPhysicID], ShouldResemble,
					resourceapi.DeviceAttribute{IntValue: ptr.To(int64(tc.dev.PhyID))})
				So(attrs[attrKeyChipName], ShouldResemble,
					resourceapi.DeviceAttribute{StringValue: ptr.To(common.Chip910)})
			})
		}
		Convey("GetChipInfo error yields empty chipName", func() {
			p := gomonkey.NewPatches()
			defer p.Reset()
			patchGetChipInfoErr(p)
			attrs := g.DeviceAttributes(NpuDevice{LogicID: 1})
			So(attrs[attrKeyChipName], ShouldResemble,
				resourceapi.DeviceAttribute{StringValue: ptr.To("")})
		})
	})
}
