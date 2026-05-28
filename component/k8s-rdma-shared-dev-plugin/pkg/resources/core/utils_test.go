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

package core

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/vishvananda/netlink"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
)

var errMock = errors.New("mock error")

// testDevice implements types.Device for testing
type testDevice struct {
	name     string
	vendor   string
	deviceID string
	driver   string
	ifName   string
	linkType string
}

func (d *testDevice) GetName() string                      { return d.name }
func (d *testDevice) GetVendor() string                    { return d.vendor }
func (d *testDevice) GetDeviceID() string                  { return d.deviceID }
func (d *testDevice) GetDriver() string                    { return d.driver }
func (d *testDevice) GetRdmaSpec() []*pluginapi.DeviceSpec { return nil }
func (d *testDevice) GetIfName() string                    { return d.ifName }
func (d *testDevice) GetLinkType() string                  { return d.linkType }

// testPciDevice implements types.PciNetDevice for testing
type testPciDevice struct {
	testDevice
	pciAddr string
}

func (d *testPciDevice) GetPciAddr() string { return d.pciAddr }

// testUbDevice implements types.UbDevice for testing
type testUbDevice struct {
	testDevice
	ubID       string
	deviceName string
}

func (d *testUbDevice) GetUbID() string       { return d.ubID }
func (d *testUbDevice) GetDeviceName() string { return d.deviceName }

// fakeLink implements netlink.Link for testing
type fakeLink struct {
	netlink.LinkAttrs
}

func (l *fakeLink) Attrs() *netlink.LinkAttrs { return &l.LinkAttrs }
func (l *fakeLink) Type() string              { return "fakeLink" }

func TestNewNetlinkManager(t *testing.T) {
	c.Convey("Given NewNetlinkManager function", t, func() {
		nm := NewNetlinkManager()
		c.Convey("When creating a new netlink manager", func() {
			_, ok := nm.(*netlinkManager)
			c.Convey("Then it should return a non-nil netlinkManager instance", func() {
				c.So(nm, c.ShouldNotBeNil)
				c.So(ok, c.ShouldBeTrue)
			})
		})
	})
}

func TestCreateNetlinkManager(t *testing.T) {
	c.Convey("Given CreateNetlinkManager function", t, func() {
		nm := CreateNetlinkManager()
		c.Convey("When creating a new netlink manager via CreateNetlinkManager", func() {
			_, ok := nm.(*netlinkManager)
			c.Convey("Then it should return a non-nil netlinkManager instance", func() {
				c.So(nm, c.ShouldNotBeNil)
				c.So(ok, c.ShouldBeTrue)
			})
		})
	})
}

func TestNetlinkManagerLinkByName(t *testing.T) {
	c.Convey("Given a netlinkManager", t, func() {
		nm := NewNetlinkManager()

		c.Convey("When LinkByName succeeds", func() {
			link := &fakeLink{}
			patches := gomonkey.ApplyFunc(netlink.LinkByName,
				func(name string) (netlink.Link, error) {
					c.So(name, c.ShouldEqual, "eth0")
					return link, nil
				})
			defer patches.Reset()

			result, err := nm.LinkByName("eth0")

			c.Convey("Then it should return the link without error", func() {
				c.So(err, c.ShouldBeNil)
				c.So(result, c.ShouldEqual, link)
			})
		})

		c.Convey("When LinkByName fails", func() {
			patches := gomonkey.ApplyFunc(netlink.LinkByName,
				func(name string) (netlink.Link, error) {
					return nil, errMock
				})
			defer patches.Reset()

			result, err := nm.LinkByName("eth1")

			c.Convey("Then it should return an error", func() {
				c.So(err, c.ShouldEqual, errMock)
				c.So(result, c.ShouldBeNil)
			})
		})
	})
}

func TestNetlinkManagerLinkSetUp(t *testing.T) {
	c.Convey("Given a netlinkManager", t, func() {
		nm := NewNetlinkManager()

		c.Convey("When LinkSetUp succeeds", func() {
			link := &fakeLink{}
			patches := gomonkey.ApplyFunc(netlink.LinkSetUp,
				func(l netlink.Link) error {
					c.So(l, c.ShouldEqual, link)
					return nil
				})
			defer patches.Reset()

			err := nm.LinkSetUp(link)

			c.Convey("Then it should return nil error", func() {
				c.So(err, c.ShouldBeNil)
			})
		})

		c.Convey("When LinkSetUp fails", func() {
			link := &fakeLink{}
			patches := gomonkey.ApplyFunc(netlink.LinkSetUp,
				func(l netlink.Link) error {
					return errMock
				})
			defer patches.Reset()

			err := nm.LinkSetUp(link)

			c.Convey("Then it should return the error", func() {
				c.So(err, c.ShouldEqual, errMock)
			})
		})
	})
}

func TestNewVendorSelector(t *testing.T) {
	c.Convey("Given NewVendorSelector function", t, func() {
		vendors := []string{"vendor-a", "vendor-b"}
		selector := NewVendorSelector(vendors)
		c.Convey("When creating a vendor selector", func() {
			_, ok := selector.(*vendorSelector)
			c.Convey("Then it should return a non-nil vendorSelector instance", func() {
				c.So(selector, c.ShouldNotBeNil)
				c.So(ok, c.ShouldBeTrue)
			})
		})
	})
}

func TestVendorSelectorFilter(t *testing.T) {
	c.Convey("Given a vendorSelector with vendors [vendor-a, vendor-b]", t, func() {
		selector := NewVendorSelector([]string{"vendor-a", "vendor-b"})

		c.Convey("When filtering devices with matching vendors", func() {
			devices := []types.Device{
				&testDevice{name: "dev1", vendor: "vendor-a"},
				&testDevice{name: "dev2", vendor: "vendor-c"},
				&testDevice{name: "dev3", vendor: "vendor-b"},
			}
			result := selector.Filter(devices)

			c.Convey("Then only matching devices should be returned", func() {
				c.So(len(result), c.ShouldEqual, 2)
				c.So(result[0].GetName(), c.ShouldEqual, "dev1")
				c.So(result[1].GetName(), c.ShouldEqual, "dev3")
			})
		})

		c.Convey("When filtering devices with no matching vendors", func() {
			devices := []types.Device{
				&testDevice{name: "dev1", vendor: "vendor-c"},
				&testDevice{name: "dev2", vendor: "vendor-d"},
			}
			result := selector.Filter(devices)

			c.Convey("Then an empty slice should be returned", func() {
				c.So(len(result), c.ShouldEqual, 0)
			})
		})

		c.Convey("When filtering an empty device list", func() {
			result := selector.Filter([]types.Device{})

			c.Convey("Then an empty slice should be returned", func() {
				c.So(len(result), c.ShouldEqual, 0)
			})
		})
	})
}

func TestNewDeviceSelector(t *testing.T) {
	c.Convey("Given NewDeviceSelector function", t, func() {
		deviceIDs := []string{"id-1", "id-2"}
		selector := NewDeviceSelector(deviceIDs)
		c.Convey("When creating a device selector", func() {
			_, ok := selector.(*deviceIDSelector)
			c.Convey("Then it should return a non-nil deviceIDSelector instance", func() {
				c.So(selector, c.ShouldNotBeNil)
				c.So(ok, c.ShouldBeTrue)
			})
		})
	})
}

func TestDeviceIDSelectorFilter(t *testing.T) {
	c.Convey("Given a deviceIDSelector with ids [id-1, id-2]", t, func() {
		selector := NewDeviceSelector([]string{"id-1", "id-2"})

		c.Convey("When filtering devices with matching device IDs", func() {
			devices := []types.Device{
				&testDevice{name: "dev1", deviceID: "id-1"},
				&testDevice{name: "dev2", deviceID: "id-3"},
				&testDevice{name: "dev3", deviceID: "id-2"},
			}
			result := selector.Filter(devices)

			c.Convey("Then only matching devices should be returned", func() {
				c.So(len(result), c.ShouldEqual, 2)
				c.So(result[0].GetName(), c.ShouldEqual, "dev1")
				c.So(result[1].GetName(), c.ShouldEqual, "dev3")
			})
		})

		c.Convey("When filtering devices with no matching device IDs", func() {
			devices := []types.Device{
				&testDevice{name: "dev1", deviceID: "id-5"},
				&testDevice{name: "dev2", deviceID: "id-6"},
			}
			result := selector.Filter(devices)

			c.Convey("Then an empty slice should be returned", func() {
				c.So(len(result), c.ShouldEqual, 0)
			})
		})

		c.Convey("When filtering an empty device list", func() {
			result := selector.Filter([]types.Device{})

			c.Convey("Then an empty slice should be returned", func() {
				c.So(len(result), c.ShouldEqual, 0)
			})
		})
	})
}

func TestNewDriverSelector(t *testing.T) {
	c.Convey("Given NewDriverSelector function", t, func() {
		drivers := []string{"mlx5_core", "bnxt_en"}
		selector := NewDriverSelector(drivers)
		c.Convey("When creating a driver selector", func() {
			_, ok := selector.(*driverSelector)
			c.Convey("Then it should return a non-nil driverSelector instance", func() {
				c.So(selector, c.ShouldNotBeNil)
				c.So(ok, c.ShouldBeTrue)
			})
		})
	})
}

func TestDriverSelectorFilter(t *testing.T) {
	c.Convey("Given a driverSelector with drivers [mlx5_core, bnxt_en]", t, func() {
		selector := NewDriverSelector([]string{"mlx5_core", "bnxt_en"})

		c.Convey("When filtering devices with matching drivers", func() {
			devices := []types.Device{
				&testDevice{name: "dev1", driver: "mlx5_core"},
				&testDevice{name: "dev2", driver: "i40e"},
				&testDevice{name: "dev3", driver: "bnxt_en"},
			}
			result := selector.Filter(devices)

			c.Convey("Then only matching devices should be returned", func() {
				c.So(len(result), c.ShouldEqual, 2)
				c.So(result[0].GetName(), c.ShouldEqual, "dev1")
				c.So(result[1].GetName(), c.ShouldEqual, "dev3")
			})
		})

		c.Convey("When filtering devices with no matching drivers", func() {
			devices := []types.Device{
				&testDevice{name: "dev1", driver: "i40e"},
				&testDevice{name: "dev2", driver: "ixgbe"},
			}
			result := selector.Filter(devices)

			c.Convey("Then an empty slice should be returned", func() {
				c.So(len(result), c.ShouldEqual, 0)
			})
		})

		c.Convey("When filtering an empty device list", func() {
			result := selector.Filter([]types.Device{})

			c.Convey("Then an empty slice should be returned", func() {
				c.So(len(result), c.ShouldEqual, 0)
			})
		})
	})
}

func TestNewIfNameSelector(t *testing.T) {
	c.Convey("Given NewIfNameSelector function", t, func() {
		ifNames := []string{"eth0", "ib0"}
		selector := NewIfNameSelector(ifNames)
		c.Convey("When creating an ifName selector", func() {
			_, ok := selector.(*ifNameSelector)
			c.Convey("Then it should return a non-nil ifNameSelector instance", func() {
				c.So(selector, c.ShouldNotBeNil)
				c.So(ok, c.ShouldBeTrue)
			})
		})
	})
}

func TestIfNameSelectorFilter(t *testing.T) {
	c.Convey("Given an ifNameSelector with names [eth0, ib0]", t, func() {
		selector := NewIfNameSelector([]string{"eth0", "ib0"})

		c.Convey("When filtering PciNetDevices with matching ifNames", func() {
			devices := []types.Device{
				&testPciDevice{testDevice: testDevice{name: "pci1", ifName: "eth0"}, pciAddr: "0000:01:00.0"},
				&testPciDevice{testDevice: testDevice{name: "pci2", ifName: "eth1"}, pciAddr: "0000:02:00.0"},
				&testPciDevice{testDevice: testDevice{name: "pci3", ifName: "ib0"}, pciAddr: "0000:03:00.0"},
			}
			result := selector.Filter(devices)

			c.Convey("Then only matching PciNetDevices should be returned", func() {
				c.So(len(result), c.ShouldEqual, 2)
				c.So(result[0].GetName(), c.ShouldEqual, "pci1")
				c.So(result[1].GetName(), c.ShouldEqual, "pci3")
			})
		})

		c.Convey("When filtering UbDevices with matching ifNames", func() {
			devices := []types.Device{
				&testUbDevice{testDevice: testDevice{name: "ub1", ifName: "eth0"}, ubID: "ub-1", deviceName: "dname1"},
				&testUbDevice{testDevice: testDevice{name: "ub2", ifName: "eth2"}, ubID: "ub-2", deviceName: "dname2"},
			}
			result := selector.Filter(devices)

			c.Convey("Then only matching UbDevices should be returned", func() {
				c.So(len(result), c.ShouldEqual, 1)
				c.So(result[0].GetName(), c.ShouldEqual, "ub1")
			})
		})

		c.Convey("When filtering plain Devices (not PciNetDevice or UbDevice)", func() {
			devices := []types.Device{
				&testDevice{name: "dev1", ifName: "eth0"},
				&testDevice{name: "dev2", ifName: "ib0"},
			}
			result := selector.Filter(devices)

			c.Convey("Then no devices should be returned (plain Device has no ifName via type assertion)", func() {
				c.So(len(result), c.ShouldEqual, 0)
			})
		})

		c.Convey("When filtering mixed device types", func() {
			devices := []types.Device{
				&testPciDevice{testDevice: testDevice{name: "pci1", ifName: "eth0"}, pciAddr: "0000:01:00.0"},
				&testUbDevice{testDevice: testDevice{name: "ub1", ifName: "ib0"}, ubID: "ub-1", deviceName: "dname1"},
				&testPciDevice{testDevice: testDevice{name: "pci2", ifName: "eth1"}, pciAddr: "0000:02:00.0"},
			}
			result := selector.Filter(devices)

			c.Convey("Then both PciNetDevice and UbDevice with matching ifNames should be returned", func() {
				c.So(len(result), c.ShouldEqual, 2)
				c.So(result[0].GetName(), c.ShouldEqual, "pci1")
				c.So(result[1].GetName(), c.ShouldEqual, "ub1")
			})
		})

		c.Convey("When filtering an empty device list", func() {
			result := selector.Filter([]types.Device{})

			c.Convey("Then an empty slice should be returned", func() {
				c.So(len(result), c.ShouldEqual, 0)
			})
		})
	})
}

func TestNewLinkTypeSelector(t *testing.T) {
	c.Convey("Given NewLinkTypeSelector function", t, func() {
		linkTypes := []string{"ether", "infiniband"}
		selector := NewLinkTypeSelector(linkTypes)
		c.Convey("When creating a linkType selector", func() {
			_, ok := selector.(*linkTypeSelector)
			c.Convey("Then it should return a non-nil linkTypeSelector instance", func() {
				c.So(selector, c.ShouldNotBeNil)
				c.So(ok, c.ShouldBeTrue)
			})
		})
	})
}

func TestLinkTypeSelectorFilter(t *testing.T) {
	c.Convey("Given a linkTypeSelector with types [ether, infiniband]", t, func() {
		selector := NewLinkTypeSelector([]string{"ether", "infiniband"})

		c.Convey("When filtering PciNetDevices with matching linkTypes", func() {
			devices := []types.Device{
				&testPciDevice{testDevice: testDevice{name: "pci1", linkType: "ether"}, pciAddr: "0000:01:00.0"},
				&testPciDevice{testDevice: testDevice{name: "pci2", linkType: "other"}, pciAddr: "0000:02:00.0"},
				&testPciDevice{testDevice: testDevice{name: "pci3", linkType: "infiniband"}, pciAddr: "0000:03:00.0"},
			}
			result := selector.Filter(devices)

			c.Convey("Then only matching PciNetDevices should be returned", func() {
				c.So(len(result), c.ShouldEqual, 2)
				c.So(result[0].GetName(), c.ShouldEqual, "pci1")
				c.So(result[1].GetName(), c.ShouldEqual, "pci3")
			})
		})

		c.Convey("When filtering UbDevices with matching linkTypes", func() {
			devices := []types.Device{
				&testUbDevice{testDevice: testDevice{name: "ub1", linkType: "infiniband"}, ubID: "ub-1", deviceName: "dname1"},
				&testUbDevice{testDevice: testDevice{name: "ub2", linkType: "other"}, ubID: "ub-2", deviceName: "dname2"},
			}
			result := selector.Filter(devices)

			c.Convey("Then only matching UbDevices should be returned", func() {
				c.So(len(result), c.ShouldEqual, 1)
				c.So(result[0].GetName(), c.ShouldEqual, "ub1")
			})
		})

		c.Convey("When filtering plain Devices (not PciNetDevice or UbDevice)", func() {
			devices := []types.Device{
				&testDevice{name: "dev1", linkType: "ether"},
				&testDevice{name: "dev2", linkType: "infiniband"},
			}
			result := selector.Filter(devices)

			c.Convey("Then no devices should be returned", func() {
				c.So(len(result), c.ShouldEqual, 0)
			})
		})

		c.Convey("When filtering mixed device types", func() {
			devices := []types.Device{
				&testPciDevice{testDevice: testDevice{name: "pci1", linkType: "ether"}, pciAddr: "0000:01:00.0"},
				&testUbDevice{testDevice: testDevice{name: "ub1", linkType: "infiniband"}, ubID: "ub-1", deviceName: "dname1"},
				&testPciDevice{testDevice: testDevice{name: "pci2", linkType: "other"}, pciAddr: "0000:02:00.0"},
			}
			result := selector.Filter(devices)

			c.Convey("Then both PciNetDevice and UbDevice with matching linkTypes should be returned", func() {
				c.So(len(result), c.ShouldEqual, 2)
				c.So(result[0].GetName(), c.ShouldEqual, "pci1")
				c.So(result[1].GetName(), c.ShouldEqual, "ub1")
			})
		})

		c.Convey("When filtering an empty device list", func() {
			result := selector.Filter([]types.Device{})

			c.Convey("Then an empty slice should be returned", func() {
				c.So(len(result), c.ShouldEqual, 0)
			})
		})
	})
}

func TestGetFilteredDevices(t *testing.T) {
	c.Convey("Given a list of devices and selectors", t, func() {
		devices := []types.Device{
			&testPciDevice{
				testDevice: testDevice{
					name: "pci1", vendor: "vendor-a", deviceID: "id-1",
					driver: "mlx5_core", ifName: "eth0", linkType: "ether",
				},
				pciAddr: "0000:01:00.0",
			},
			&testPciDevice{
				testDevice: testDevice{
					name: "pci2", vendor: "vendor-b", deviceID: "id-2",
					driver: "i40e", ifName: "eth1", linkType: "ether",
				},
				pciAddr: "0000:02:00.0",
			},
			&testUbDevice{
				testDevice: testDevice{
					name: "ub1", vendor: "vendor-a", deviceID: "id-3",
					driver: "mlx5_core", ifName: "ib0", linkType: "infiniband",
				},
				ubID: "ub-1", deviceName: "dname1",
			},
			&testPciDevice{
				testDevice: testDevice{
					name: "pci3", vendor: "vendor-c", deviceID: "id-4",
					driver: "mlx5_core", ifName: "ib1", linkType: "infiniband",
				},
				pciAddr: "0000:03:00.0",
			},
		}

		c.Convey("When no selectors are specified", func() {
			selector := &types.Selectors{}
			result := GetFilteredDevices(devices, selector)

			c.Convey("Then all devices should be returned", func() {
				c.So(len(result), c.ShouldEqual, 4)
			})
		})

		c.Convey("When filtering by vendors only", func() {
			selector := &types.Selectors{
				Vendors: []string{"vendor-a"},
			}
			result := GetFilteredDevices(devices, selector)

			c.Convey("Then only devices with matching vendor should be returned", func() {
				c.So(len(result), c.ShouldEqual, 2)
				c.So(result[0].GetName(), c.ShouldEqual, "pci1")
				c.So(result[1].GetName(), c.ShouldEqual, "ub1")
			})
		})

		c.Convey("When filtering by device IDs only", func() {
			selector := &types.Selectors{
				DeviceIDs: []string{"id-1", "id-4"},
			}
			result := GetFilteredDevices(devices, selector)

			c.Convey("Then only devices with matching device IDs should be returned", func() {
				c.So(len(result), c.ShouldEqual, 2)
				c.So(result[0].GetName(), c.ShouldEqual, "pci1")
				c.So(result[1].GetName(), c.ShouldEqual, "pci3")
			})
		})

		c.Convey("When filtering by drivers only", func() {
			selector := &types.Selectors{
				Drivers: []string{"i40e"},
			}
			result := GetFilteredDevices(devices, selector)

			c.Convey("Then only devices with matching driver should be returned", func() {
				c.So(len(result), c.ShouldEqual, 1)
				c.So(result[0].GetName(), c.ShouldEqual, "pci2")
			})
		})

		c.Convey("When filtering by ifNames only", func() {
			selector := &types.Selectors{
				IfNames: []string{"eth0", "ib0"},
			}
			result := GetFilteredDevices(devices, selector)

			c.Convey("Then only devices with matching ifNames should be returned", func() {
				c.So(len(result), c.ShouldEqual, 2)
				c.So(result[0].GetName(), c.ShouldEqual, "pci1")
				c.So(result[1].GetName(), c.ShouldEqual, "ub1")
			})
		})

		c.Convey("When filtering by linkTypes only", func() {
			selector := &types.Selectors{
				LinkTypes: []string{"infiniband"},
			}
			result := GetFilteredDevices(devices, selector)

			c.Convey("Then only devices with matching linkTypes should be returned", func() {
				c.So(len(result), c.ShouldEqual, 2)
				c.So(result[0].GetName(), c.ShouldEqual, "ub1")
				c.So(result[1].GetName(), c.ShouldEqual, "pci3")
			})
		})

		c.Convey("When filtering by multiple selectors combined", func() {
			selector := &types.Selectors{
				Vendors:   []string{"vendor-a"},
				Drivers:   []string{"mlx5_core"},
				LinkTypes: []string{"infiniband"},
			}
			result := GetFilteredDevices(devices, selector)

			c.Convey("Then only devices matching all selector criteria should be returned", func() {
				c.So(len(result), c.ShouldEqual, 1)
				c.So(result[0].GetName(), c.ShouldEqual, "ub1")
			})
		})

		c.Convey("When filtering with matching vendors but non-matching drivers", func() {
			selector := &types.Selectors{
				Vendors: []string{"vendor-c"},
				Drivers: []string{"i40e"},
			}
			result := GetFilteredDevices(devices, selector)

			c.Convey("Then no devices should be returned (no device matches both)", func() {
				c.So(len(result), c.ShouldEqual, 0)
			})
		})

		c.Convey("When filtering an empty device list", func() {
			selector := &types.Selectors{
				Vendors: []string{"vendor-a"},
			}
			result := GetFilteredDevices([]types.Device{}, selector)

			c.Convey("Then an empty slice should be returned", func() {
				c.So(len(result), c.ShouldEqual, 0)
			})
		})

		c.Convey("When filtering with all selectors set", func() {
			selector := &types.Selectors{
				Vendors:   []string{"vendor-a"},
				DeviceIDs: []string{"id-1"},
				Drivers:   []string{"mlx5_core"},
				IfNames:   []string{"eth0"},
				LinkTypes: []string{"ether"},
			}
			result := GetFilteredDevices(devices, selector)

			c.Convey("Then only the device matching all criteria should be returned", func() {
				c.So(len(result), c.ShouldEqual, 1)
				c.So(result[0].GetName(), c.ShouldEqual, "pci1")
			})
		})

		c.Convey("When result slice should be independent from the input", func() {
			selector := &types.Selectors{}
			result := GetFilteredDevices(devices, selector)

			result[0] = nil

			c.Convey("Then modifying the result should not affect the original devices", func() {
				c.So(devices[0], c.ShouldNotBeNil)
				c.So(len(result), c.ShouldEqual, 4)
				c.So(len(devices), c.ShouldEqual, 4)
			})
		})
	})
}
