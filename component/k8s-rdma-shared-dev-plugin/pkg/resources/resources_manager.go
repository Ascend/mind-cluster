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

/*----------------------------------------------------

  2023 NVIDIA CORPORATION & AFFILIATES

  Licensed under the Apache License, Version 2.0 (the License);
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an AS IS BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

----------------------------------------------------*/

package resources

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/jaypipes/ghw"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/cdi"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/common"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/core"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
)

const (
	// General constants
	kubeEndPoint          = "kubelet.sock"
	socketSuffix          = "sock"
	rdmaHcaResourcePrefix = "rdma"

	// PCI related constants
	netClass             = 0x02 // Device class - Network controller
	maxVendorNameLength  = 20
	maxProductNameLength = 40
)

var (
	activeSockDir = "/var/lib/kubelet/plugins_registry"
)

// resourceManager for PCI device plugin
type resourceManager struct {
	core.CoreResourceManager
	deviceList     []*ghw.PCIDevice
	netlinkManager types.NetlinkManager
	rds            types.RdmaDeviceSpec
}

func NewResourceManager(configFile string, useCdi bool) types.ResourceManager {
	// Create core resource manager
	coreManager := core.NewCoreResourceManager(configFile, rdmaHcaResourcePrefix, socketSuffix, useCdi)

	// Create PCI-specific resource manager
	return &resourceManager{
		CoreResourceManager: coreManager,
		deviceList:          []*ghw.PCIDevice{},
		netlinkManager:      core.NewNetlinkManager(),
		rds:                 NewRdmaDeviceSpec(common.RequiredRdmaDevices),
	}
}

// InitServers init server
func (rm *resourceManager) InitServers() error {
	// Use core method to get config list
	for _, config := range rm.GetConfigList() {
		log.Printf("Resource: %+v\n", config)
		devices := rm.GetDevices()
		filteredDevices := rm.GetFilteredDevices(devices, &config.Selectors)
		// NOTE: it's a temporary workaround to bring interfaces in UP state until
		// Network Operator will be able to do it.
		for _, device := range filteredDevices {
			if pciDev, ok := device.(types.PciNetDevice); ok {
				link, err := rm.netlinkManager.LinkByName(pciDev.GetIfName())
				if err != nil {
					log.Printf("Warning: InitServers(): unable to get NIC info: %s", err)
					continue
				}

				if rm.netlinkManager.LinkSetUp(link) != nil {
					log.Printf("Warning: InitServers(): unable to set NIC %s to up state: %s", pciDev.GetIfName(), err)
					continue
				}
			}
		}

		if len(filteredDevices) == 0 {
			log.Printf("Warning: no devices in device pool, creating empty resource server for %s", config.ResourceName)
		}

		if rm.GetUseCdi() {
			err := cdi.CleanupSpecs(cdiResourcePrefix)
			if err != nil {
				return err
			}
		}
		// TODO: Need to get watchMode and socketSuffix from core or add methods to core
		rs, err := newResourceServer(config, filteredDevices, false, socketSuffix, rm.GetUseCdi())
		if err != nil {
			return err
		}
		// Use core method to add resource server
		rm.AddResourceServer(rs)
	}
	return nil
}

func validResourceName(name string) bool {
	// name regex
	var validString = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return validString.MatchString(name)
}

func (rm *resourceManager) DiscoverHostDevices() error {
	log.Println("discovering host network devices")

	// Discover PCI devices only
	pci, err := ghw.PCI()
	if err != nil {
		return fmt.Errorf("error getting PCI info: %v", err)
	}

	devices := pci.Devices
	if len(devices) == 0 {
		log.Println("Warning: DiscoverHostDevices(): no PCI network device found")
	}

	// cleanup deviceList as this method is also called during periodic update for the resources
	rm.deviceList = []*ghw.PCIDevice{}

	for _, device := range devices {
		devClass, err := strconv.ParseInt(device.Class.ID, 16, 64)
		if err != nil {
			log.Printf("Warning: DiscoverHostDevices(): unable to parse device class for device %+v %q", device,
				err)
			continue
		}

		if devClass != netClass {
			continue
		}

		vendor := device.Vendor
		vendorName := vendor.Name
		if len(vendor.Name) > maxVendorNameLength {
			vendorName = string([]byte(vendorName)[0:17]) + "..."
		}
		product := device.Product
		productName := product.Name
		if len(product.Name) > maxProductNameLength {
			productName = string([]byte(productName)[0:37]) + "..."
		}
		log.Printf("DiscoverHostDevices(): PCI device found: %-12s\t%-12s\t%-20s\t%-40s", device.Address,
			device.Class.ID, vendorName, productName)

		rm.deviceList = append(rm.deviceList, device)
	}

	return nil
}

func (rm *resourceManager) GetDevices() []types.Device {
	newDevices := make([]types.Device, 0)

	// Add PCI devices only
	for _, device := range rm.deviceList {
		if newDevice, err := NewPciNetDevice(device, rm.rds, rm.netlinkManager); err == nil {
			newDevices = append(newDevices, newDevice)
		} else {
			log.Printf("error creating PCI device: %q", err)
		}
	}

	return newDevices
}

func (rm *resourceManager) GetFilteredDevices(devices []types.Device,
	selector *types.Selectors) []types.Device {
	// Use core package's GetFilteredDevices function
	return core.GetFilteredDevices(devices, selector)
}

func (rm *resourceManager) PeriodicUpdate() func() {
	stopChan := make(chan interface{})
	interval := rm.GetPeriodicUpdateInterval()
	if interval > 0 {
		ticker := time.NewTicker(interval)
		go rm.runPeriodicUpdate(ticker, stopChan)
	}
	return func() {
		if rm.GetPeriodicUpdateInterval() > 0 {
			stopChan <- true
			close(stopChan)
		}
	}
}

func (rm *resourceManager) runPeriodicUpdate(ticker *time.Ticker, stopChan chan interface{}) {
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := rm.DiscoverHostDevices(); err != nil {
				log.Printf("error: failed to discover host devices: %v", err)
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
			return
		}
	}
}
