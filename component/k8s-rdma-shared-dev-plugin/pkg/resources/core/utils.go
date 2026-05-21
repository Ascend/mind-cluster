// Copyright 2025 NVIDIA CORPORATION & AFFILIATES
// Modified by Huawei Technologies Co.,Ltd in 2026
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Package core for common func
package core

import (
	"github.com/vishvananda/netlink"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
)

// netlinkManager implements NetlinkManager interface
type netlinkManager struct{}

// NewNetlinkManager returns a new instance of NetlinkManager
func NewNetlinkManager() types.NetlinkManager {
	return &netlinkManager{}
}

// LinkByName returns link by name
func (nm *netlinkManager) LinkByName(name string) (netlink.Link, error) {
	return netlink.LinkByName(name)
}

// LinkSetUp sets link up
func (nm *netlinkManager) LinkSetUp(link netlink.Link) error {
	return netlink.LinkSetUp(link)
}

// CreateNetlinkManager returns a new netlink manager instance
func CreateNetlinkManager() types.NetlinkManager {
	return &netlinkManager{}
}

// NewVendorSelector selects devices by vendor
func NewVendorSelector(vendors []string) types.DeviceSelector {
	return &vendorSelector{vendors: vendors}
}

type vendorSelector struct {
	vendors []string
}

func (s *vendorSelector) Filter(inDevices []types.Device) []types.Device {
	filteredList := make([]types.Device, 0)
	for _, dev := range inDevices {
		if contains(s.vendors, dev.GetVendor()) {
			filteredList = append(filteredList, dev)
		}
	}
	return filteredList
}

// NewDeviceSelector selects devices by device ID
func NewDeviceSelector(deviceIDs []string) types.DeviceSelector {
	return &deviceIDSelector{deviceIDs: deviceIDs}
}

type deviceIDSelector struct {
	deviceIDs []string
}

// Filter filter vy deviceId
func (s *deviceIDSelector) Filter(inDevices []types.Device) []types.Device {
	filteredList := make([]types.Device, 0)
	for _, dev := range inDevices {
		if contains(s.deviceIDs, dev.GetDeviceID()) {
			filteredList = append(filteredList, dev)
		}
	}
	return filteredList
}

// NewDriverSelector driverSelector selects devices by driver
func NewDriverSelector(drivers []string) types.DeviceSelector {
	return &driverSelector{drivers: drivers}
}

type driverSelector struct {
	drivers []string
}

func (s *driverSelector) Filter(inDevices []types.Device) []types.Device {
	filteredList := make([]types.Device, 0)
	for _, dev := range inDevices {
		if contains(s.drivers, dev.GetDriver()) {
			filteredList = append(filteredList, dev)
		}
	}
	return filteredList
}

// contains checks if a slice contains a string
func contains(list []string, needle string) bool {
	for _, s := range list {
		if s == needle {
			return true
		}
	}
	return false
}

// GetFilteredDevices filters devices based on selectors
func GetFilteredDevices(devices []types.Device, selector *types.Selectors) []types.Device {
	filteredDevice := devices

	// filter by Vendors list
	if len(selector.Vendors) > 0 {
		filteredDevice = NewVendorSelector(selector.Vendors).Filter(filteredDevice)
	}

	// filter by DeviceIDs list
	if len(selector.DeviceIDs) > 0 {
		filteredDevice = NewDeviceSelector(selector.DeviceIDs).Filter(filteredDevice)
	}

	// filter by Driver list
	if len(selector.Drivers) > 0 {
		filteredDevice = NewDriverSelector(selector.Drivers).Filter(filteredDevice)
	}

	// filter by IfNames list - only applicable for PCI devices
	if len(selector.IfNames) > 0 {
		ifNamesSelector := NewIfNameSelector(selector.IfNames)
		filteredDevice = ifNamesSelector.Filter(filteredDevice)
	}

	// filter by LinkType list - only applicable for PCI devices
	if len(selector.LinkTypes) > 0 {
		linkTypeSelector := NewLinkTypeSelector(selector.LinkTypes)
		filteredDevice = linkTypeSelector.Filter(filteredDevice)
	}

	newDeviceList := make([]types.Device, len(filteredDevice))
	copy(newDeviceList, filteredDevice)

	return newDeviceList
}

// NewIfNameSelector selects devices by interface name
func NewIfNameSelector(ifNames []string) types.DeviceSelector {
	return &ifNameSelector{ifNames: ifNames}
}

type ifNameSelector struct {
	ifNames []string
}

func (s *ifNameSelector) Filter(inDevices []types.Device) []types.Device {
	filteredList := make([]types.Device, 0)
	for _, dev := range inDevices {
		var ifName string
		if pciDev, ok := dev.(types.PciNetDevice); ok {
			ifName = pciDev.GetIfName()
		} else if ubDev, ok := dev.(types.UbDevice); ok {
			ifName = ubDev.GetIfName()
		}
		if ifName != "" && contains(s.ifNames, ifName) {
			filteredList = append(filteredList, dev)
		}
	}
	return filteredList
}

// NewLinkTypeSelector selects devices by link type
func NewLinkTypeSelector(linkTypes []string) types.DeviceSelector {
	return &linkTypeSelector{linkTypes: linkTypes}
}

type linkTypeSelector struct {
	linkTypes []string
}

func (s *linkTypeSelector) Filter(inDevices []types.Device) []types.Device {
	filteredList := make([]types.Device, 0)
	for _, dev := range inDevices {
		var linkType string
		if pciDev, ok := dev.(types.PciNetDevice); ok {
			linkType = pciDev.GetLinkType()
		} else if ubDev, ok := dev.(types.UbDevice); ok {
			linkType = ubDev.GetLinkType()
		}
		if linkType != "" && contains(s.linkTypes, linkType) {
			filteredList = append(filteredList, dev)
		}
	}
	return filteredList
}
