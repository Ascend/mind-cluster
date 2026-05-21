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
// Package main
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/common"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/core"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
)

var (
	version = "master@git"
	commit  = "unknown commit"
	date    = "unknown date"
)

func printVersionString() string {
	return fmt.Sprintf("k8s-rdma-shared-dev-plugin version:%s, commit:%s, date:%s", version, commit, date)
}

func main() {
	// Init command line flags to clear vendor packages' flags, especially in init()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// add version flag
	versionOpt := false
	var configFilePath string
	flag.BoolVar(&versionOpt, "version", false, "Show application version")
	flag.BoolVar(&versionOpt, "v", false, "Show application version")
	flag.StringVar(
		&configFilePath, "config-file", common.DefaultConfigFilePath, "path to device plugin config file")
	useCdi := false
	flag.BoolVar(&useCdi, "use-cdi", false,
		"Use Container Device Interface to expose devices in containers")
	flag.Parse()
	if versionOpt {
		fmt.Printf("%s\n", printVersionString())
		return
	}

	log.Println("Starting K8s RDMA Shared Device Plugin version=", version)

	// First, read the config file to determine which device types to enable
	log.Println("Reading configuration to determine device types")
	var enablePci, enableUb bool

	// Create a temporary core manager to read the config
	tempCoreManager := core.NewCoreResourceManager(configFilePath, "rdma", "sock", useCdi)
	if err := tempCoreManager.ReadConfig(); err != nil {
		log.Fatalln(err.Error())
	}

	// Check each config to determine device type
	configList := tempCoreManager.GetConfigList()
	for _, config := range configList {
		buses := config.Selectors.Buses
		log.Println("Found buses:", buses)
		// Check if it's a UB device config
		if len(buses) > 0 && strings.Contains(strings.ToLower(buses[0]), "ub") {
			log.Println("only enable ub devices")
			enableUb = true
		} else {
			// Default to PCI device
			log.Println("only enable pci devices")
			enablePci = true
		}
	}

	if useCdi {
		log.Println("CDI enabled")
		if enableUb {
			useCdi = false
			log.Println("UB devices not supported cdi, will not enable")
		}
	}

	// Initialize resource manager and stop function
	var rm types.ResourceManager
	var stopPeriodicUpdate func()

	// Initialize and start PCI device manager if enabled
	if enablePci {
		log.Println("Initializing PCI device resource manager")
		rm = resources.NewResourceManager(configFilePath, useCdi)

		log.Println("Reading PCI device configs")
		if err := rm.ReadConfig(); err != nil {
			log.Fatalln(err.Error())
		}

		log.Println("Validating PCI device configs")
		if err := rm.ValidateConfigs(); err != nil {
			log.Fatalf("Exiting.. one or more invalid PCI configuration(s) given: %v", err)
		}

		log.Println("Validating RDMA system mode")
		if err := rm.ValidateRdmaSystemMode(); err != nil {
			log.Fatalf("Exiting.. can not change RDMA system mode: %v", err)
		}

		log.Println("Discovering PCI host devices")
		if err := rm.DiscoverHostDevices(); err != nil {
			log.Fatalf("Error: error discovering PCI host devices %v \n", err)
		}

		log.Println("Initializing PCI resource servers")
		if err := rm.InitServers(); err != nil {
			log.Fatalf("Error: initializing PCI resource servers %v \n", err)
		}

		log.Println("Starting PCI servers...")
		if err := rm.StartAllServers(); err != nil {
			log.Fatalf("Error: starting PCI resource servers %v\n", err.Error())
		}

		stopPeriodicUpdate = rm.PeriodicUpdate()
	}

	//// Initialize and start UB device manager if enabled
	//if enableUb {
	//	log.Println("Initializing UB device resource manager")
	//	rm = ub_device.NewUbResourceManager(configFilePath, useCdi)
	//
	//	log.Println("Reading UB device configs")
	//	if err := rm.ReadConfig(); err != nil {
	//		log.Fatalln(err.Error())
	//	}
	//
	//	log.Println("Validating UB device configs")
	//	if err := rm.ValidateConfigs(); err != nil {
	//		log.Fatalf("Exiting.. one or more invalid UB configuration(s) given: %v", err)
	//	}
	//
	//	log.Println("Discovering UB host devices")
	//	if err := rm.DiscoverHostDevices(); err != nil {
	//		log.Fatalf("Error: error discovering UB host devices %v \n", err)
	//	}
	//
	//	log.Println("Initializing UB resource servers")
	//	if err := rm.InitServers(); err != nil {
	//		log.Fatalf("Error: initializing UB resource servers %v \n", err)
	//	}
	//
	//	log.Println("Starting UB servers...")
	//	if err := rm.StartAllServers(); err != nil {
	//		log.Fatalf("Error: starting UB resource servers %v\n", err.Error())
	//	}
	//
	//	stopPeriodicUpdate = rm.PeriodicUpdate()
	//}

	log.Println("Enabled servers started.")

	log.Println("Listening for term signals")
	log.Println("Starting OS watcher.")
	signalsNotifier := resources.NewSignalNotifier(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sigs := signalsNotifier.Notify()

	for {
		s := <-sigs
		switch s {
		case syscall.SIGHUP:
			log.Println("Received SIGHUP, restarting servers.")
			if err := rm.RestartAllServers(); err != nil {
				log.Fatalf("unable to restart servers %v", err)
			}
		default:
			log.Printf("Received signal \"%v\", shutting down.", s)
			if stopPeriodicUpdate != nil {
				stopPeriodicUpdate()
			}
			_ = rm.StopAllServers()
			return
		}
	}
}
