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

// Package ub_device for ub device info
package ub_device

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vishvananda/netlink"

	"ascend-common/common-utils/hwlog"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/cdi"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/common"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/core"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
)

const (
	socketSuffix         = "sock"
	rdmaUbResourcePrefix = "rdma-ub"

	// CDI related constants
	cdiResourcePrefix = "huawei.com"
)

// UbResourceManager for UB device plugin
type UbResourceManager interface {
	types.ResourceManager
	// Additional UB-specific methods can be added here
}

// ubResourceManager implements UbResourceManager interface
type ubResourceManager struct {
	core.CoreResourceManager
	deviceList     []*UbDeviceInfo
	netlinkManager types.NetlinkManager
	rds            types.RdmaDeviceSpec
}

// UbDeviceInfo holds information about a UB device
type UbDeviceInfo struct {
	UbID       string
	DeviceName string
	Vendor     string
	DeviceID   string
	Driver     string
	IfName     string
	LinkType   string
}

// NewUbResourceManager returns a new instance of UbResourceManager
func NewUbResourceManager(configFile string, useCdi bool) UbResourceManager {
	coreManager := core.NewCoreResourceManager(configFile, rdmaUbResourcePrefix, socketSuffix, useCdi)

	return &ubResourceManager{
		CoreResourceManager: coreManager,
		deviceList:          []*UbDeviceInfo{},
		netlinkManager:      &netlinkManager{},
		rds:                 newUbRdmaDeviceSpec(common.RequiredRdmaDevices),
	}
}

// DiscoverHostDevices discovers UB devices on the host
func (rm *ubResourceManager) DiscoverHostDevices() error {
	rm.deviceList = []*UbDeviceInfo{}

	if _, err := os.Stat(common.SysBusUb); os.IsNotExist(err) {
		hwlog.RunLog.Infof("UB devices directory %s does not exist, no UB devices found", common.SysBusUb)
		return nil
	}

	entries, err := os.ReadDir(common.SysBusUb)
	if err != nil {
		return fmt.Errorf("error reading UB devices directory %s: %v", common.SysBusUb, err)
	}

	for _, entry := range entries {
		deviceDir := filepath.Join(common.SysBusUb, entry.Name())

		if info, err := rm.readUbDeviceInfo(deviceDir, entry.Name()); err == nil {
			hwlog.RunLog.Infof("DiscoverHostDevices(): UB device found: %-12s	%-8s	%-8s	%-20s	%-8s",
				info.UbID, info.Vendor, info.DeviceID, info.Driver, info.IfName)
			rm.deviceList = append(rm.deviceList, info)
		}
	}

	if len(rm.deviceList) == 0 {
		hwlog.RunLog.Warn("DiscoverHostDevices(): no UB devices found")
	}

	return nil
}

func (rm *ubResourceManager) readUbDeviceInfo(deviceDir, ubID string) (*UbDeviceInfo, error) {
	vendor := ""
	deviceID := ""

	if data, err := os.ReadFile(filepath.Join(deviceDir, "vendor")); err == nil {
		vendor = strings.TrimSpace(string(data))
	}

	if data, err := os.ReadFile(filepath.Join(deviceDir, "device")); err == nil {
		deviceID = strings.TrimSpace(string(data))
	}

	driverInfo, err := os.Readlink(filepath.Join(deviceDir, "driver"))
	if err != nil {
		return nil, fmt.Errorf("error getting driver info for UB device %s: %v", ubID, err)
	}
	driver := filepath.Base(driverInfo)

	ifName, linkType := rm.readUbNetInfo(deviceDir, ubID)

	return &UbDeviceInfo{
		UbID:       ubID,
		DeviceName: ubID,
		Vendor:     vendor,
		DeviceID:   deviceID,
		Driver:     driver,
		IfName:     ifName,
		LinkType:   linkType,
	}, nil
}

func (rm *ubResourceManager) readUbNetInfo(deviceDir, ubID string) (string, string) {
	netDir := filepath.Join(deviceDir, "net")
	netEntries, err := os.ReadDir(netDir)
	if err != nil || len(netEntries) == 0 {
		return "", ""
	}

	ifName := netEntries[0].Name()
	if len(netEntries) > 1 {
		hwlog.RunLog.Warnf("found several net names for UB device %s, using first name %s", ubID, ifName)
	}

	link, err := rm.netlinkManager.LinkByName(ifName)
	if err != nil {
		hwlog.RunLog.Warnf("unable to get link info for UB device %s net %s: %v", ubID, ifName, err)
		return ifName, ""
	}

	return ifName, link.Attrs().EncapType
}

// GetDevices returns the list of UB devices
func (rm *ubResourceManager) GetDevices() []types.Device {
	devices := make([]types.Device, 0)
	for _, deviceInfo := range rm.deviceList {
		if device, err := NewUbDevice(
			deviceInfo.UbID,
			deviceInfo.DeviceName,
			deviceInfo.Vendor,
			deviceInfo.DeviceID,
			deviceInfo.Driver,
			deviceInfo.IfName,
			deviceInfo.LinkType,
			rm.rds,
		); err == nil {
			devices = append(devices, device)
		} else {
			hwlog.RunLog.Infof("Error creating UB device: %v", err)
		}
	}
	return devices
}

// InitServers initializes the resource servers for UB devices
func (rm *ubResourceManager) InitServers() error {
	for _, config := range rm.GetConfigList() {
		hwlog.RunLog.Infof("UB Resource Config: %+v\n", config)
		devices := rm.GetDevices()
		filteredDevices := rm.GetFilteredDevices(devices, &config.Selectors)
		hwlog.RunLog.Infof("UB resource %s: total devices=%d, filtered devices=%d", config.ResourceName,
			len(devices), len(filteredDevices))

		rm.setUbNicsUp(filteredDevices)

		if len(filteredDevices) == 0 {
			hwlog.RunLog.Warnf("no UB devices in device pool, creating empty resource server for %s",
				config.ResourceName)
		}

		if rm.GetUseCdi() {
			if err := cdi.CleanupSpecs(cdiResourcePrefix); err != nil {
				return err
			}
		}

		rs, err := NewUbResourceServer(config, filteredDevices, false, socketSuffix, rm.GetUseCdi())
		if err != nil {
			return err
		}
		rm.AddResourceServer(rs)
	}

	return nil
}

func (rm *ubResourceManager) setUbNicsUp(devices []types.Device) {
	for _, device := range devices {
		ubDev, ok := device.(types.UbDevice)
		if !ok || ubDev.GetIfName() == "" {
			continue
		}
		link, err := rm.netlinkManager.LinkByName(ubDev.GetIfName())
		if err != nil {
			hwlog.RunLog.Warnf("InitServers(): unable to get NIC info for UB device %s: %s", ubDev.GetUbID(), err)
			continue
		}
		if err := rm.netlinkManager.LinkSetUp(link); err != nil {
			hwlog.RunLog.Warnf("InitServers(): unable to set NIC %s to up state: %s", ubDev.GetIfName(), err)
		}
	}
}

// GetFilteredDevices filters UB devices based on selectors
func (rm *ubResourceManager) GetFilteredDevices(devices []types.Device, selector *types.Selectors) []types.Device {
	// Use core package's GetFilteredDevices function
	return core.GetFilteredDevices(devices, selector)
}

// PeriodicUpdate returns a function that updates UB devices periodically
func (rm *ubResourceManager) PeriodicUpdate() func() {
	stopChan := make(chan interface{}, 1)
	done := make(chan struct{})
	interval := rm.GetPeriodicUpdateInterval()

	if interval > 0 {
		hwlog.RunLog.Infof("Starting periodic update for UB devices with interval %v", interval)
		go rm.runPeriodicUpdate(interval, stopChan, done)
	} else {
		close(done)
	}

	return func() {
		if rm.GetPeriodicUpdateInterval() > 0 {
			stopChan <- struct{}{}
		}
		<-done
	}
}

func (rm *ubResourceManager) runPeriodicUpdate(interval time.Duration, stopChan chan interface{}, done chan struct{}) {
	defer close(done)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hwlog.RunLog.Info("Performing periodic update for UB devices")
			if err := rm.DiscoverHostDevices(); err != nil {
				hwlog.RunLog.Errorf("Error discovering UB devices during periodic update: %v", err)
				continue
			}

			resourceServers := rm.GetResourceServers()
			configList := rm.GetConfigList()
			for index, rs := range resourceServers {
				devices := rm.GetDevices()
				filteredDevices := rm.GetFilteredDevices(devices, &configList[index].Selectors)
				rs.UpdateDevices(filteredDevices)
			}
		case <-stopChan:
			hwlog.RunLog.Info("Stopping periodic update for UB devices")
			return
		}
	}
}

// netlinkManager implements types.NetlinkManager interface
type netlinkManager struct{}

// LinkByName gets a link by name
func (nlm *netlinkManager) LinkByName(name string) (netlink.Link, error) {
	return netlink.LinkByName(name)
}

// LinkSetUp sets a link up
func (nlm *netlinkManager) LinkSetUp(link netlink.Link) error {
	return netlink.LinkSetUp(link)
}
