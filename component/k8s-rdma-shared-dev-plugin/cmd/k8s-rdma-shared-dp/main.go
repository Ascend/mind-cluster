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
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"ascend-common/common-utils/hwlog"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/fault"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/common"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/core"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/ub_device"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
)

const (
	// Default log file path
	defaultLogFile = "/var/log/mindx-dl/k8s-rdma-shared-dp/k8s-rdma-shared-dp.log"

	// Max log line length
	maxLogLineLength = 1024

	// Default values for log parameters
	defaultLogLevel      = 0
	defaultLogMaxBackups = 3
	defaultLogMaxAge     = 7
)

var (
	// Single variable for both -version and -v flags
	versionOpt bool

	// Other flag variables
	configFilePath string
	useCdi         bool
	logLevel       int
	logMaxBackups  int
	logMaxAge      int
	logFile        string

	// Build info
	version = "master@git"
	commit  = "unknown commit"
	date    = "unknown date"
)

func printVersionString() string {
	return fmt.Sprintf("k8s-rdma-shared-dev-plugin version:%s, commit:%s, date:%s", version, commit, date)
}

func initFlags() {
	// Init command line flags to clear vendor packages' flags, especially in init()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Bind both -version and -v to the same variable
	flag.BoolVar(&versionOpt, "version", false, "Show application version")
	flag.BoolVar(&versionOpt, "v", false, "Show application version")

	// Other command line flags using value variables
	flag.StringVar(&configFilePath, "config-file", common.DefaultConfigFilePath, "Path to device plugin config file")
	flag.BoolVar(&useCdi, "use-cdi", false, "Use Container Device Interface to expose devices in containers")

	// Log related flags
	flag.IntVar(&logLevel, "logLevel", defaultLogLevel, "Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical (default 0)")
	flag.IntVar(&logMaxBackups, "maxBackups", defaultLogMaxBackups, "Maximum number of backup log files, range is (0, 30]")
	flag.IntVar(&logMaxAge, "maxAge", defaultLogMaxAge, "Maximum number of days for backup log files, range [7, 700]")
	flag.StringVar(&logFile, "logFile", defaultLogFile, "The log file path, if the file size exceeds 20MB, will be rotate")
}

func initLogModule(ctx context.Context) error {
	hwLogConfig := &hwlog.LogConfig{
		LogFileName:   logFile,
		LogLevel:      logLevel,
		MaxBackups:    logMaxBackups,
		MaxAge:        logMaxAge,
		MaxLineLength: maxLogLineLength,
	}

	if err := hwlog.InitRunLogger(hwLogConfig, ctx); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return err
	}

	return nil
}

func checkLogParams() bool {
	// Check log level: -1 <= logLevel <= 3
	if logLevel < -1 || logLevel > 3 {
		fmt.Printf("Invalid logLevel %d, range should be [-1, 3]\n", logLevel)
		return false
	}

	// Check maxBackups: 0 < maxBackups <= 30
	if logMaxBackups <= 0 || logMaxBackups > 30 {
		fmt.Printf("Invalid maxBackups %d, range should be (0, 30]\n", logMaxBackups)
		return false
	}

	// Check maxAge: 7 <= maxAge <= 700
	if logMaxAge < 7 || logMaxAge > 700 {
		fmt.Printf("Invalid maxAge %d, range should be [7, 700]\n", logMaxAge)
		return false
	}

	return true
}

func main() {
	// Initialize flags first - this must be done before flag.Parse()
	initFlags()

	// Parse command line arguments
	flag.Parse()

	// Show version information (single variable for both -version and -v)
	if versionOpt {
		fmt.Printf("%s\n", printVersionString())
		return
	}

	// Check log parameters
	if !checkLogParams() {
		return
	}

	// Initialize log module
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := initLogModule(ctx); err != nil {
		return
	}

	hwlog.RunLog.Infof("Starting K8s RDMA Shared Device Plugin version=%s", version)

	// First, read the config file to determine which device types to enable
	hwlog.RunLog.Info("Reading configuration to determine device types")
	var enablePci, enableUb bool

	// Create a temporary core manager to read the config
	tempCoreManager := core.NewCoreResourceManager(configFilePath, "rdma", "sock", useCdi)
	if err := tempCoreManager.ReadConfig(); err != nil {
		hwlog.RunLog.Errorf("%s", err.Error())
	}

	// Check each config to determine device type
	configList := tempCoreManager.GetConfigList()
	for _, config := range configList {
		buses := config.Selectors.Buses
		hwlog.RunLog.Infof("Found buses: %v", buses)
		// Check if it's a UB device config
		if len(buses) > 0 && strings.Contains(strings.ToLower(buses[0]), "ub") {
			hwlog.RunLog.Info("Only enable UB devices")
			enableUb = true
		} else {
			// Default to PCI device
			hwlog.RunLog.Info("Only enable PCI devices")
			enablePci = true
		}
	}

	if useCdi {
		hwlog.RunLog.Info("CDI enabled")
		if enableUb {
			useCdi = false
			hwlog.RunLog.Info("UB devices not supported CDI, will not enable")
		}
	}

	// Initialize resource manager and stop function
	var rm types.ResourceManager
	var stopPeriodicUpdate func()

	// Initialize and start PCI device manager if enabled
	if enablePci {
		rm, stopPeriodicUpdate = initAndStartDevices("PCI", func() types.ResourceManager {
			return resources.NewResourceManager(configFilePath, useCdi)
		})
	}

	// Initialize and start UB device manager if enabled
	if enableUb {
		rm, stopPeriodicUpdate = initAndStartDevices("UB", func() types.ResourceManager {
			return ub_device.NewUbResourceManager(configFilePath, useCdi)
		})

		if ubRm, ok := rm.(ub_device.UbResourceManager); ok {
			startFaultDetection(ctx, ubRm, tempCoreManager.GetFaultDetectPeriod())
		} else {
			hwlog.RunLog.Error("Resource manager is not of type UbResourceManager, skipping fault detection")
		}
	}

	hwlog.RunLog.Info("Enabled servers started.")

	hwlog.RunLog.Info("Listening for term signals")
	hwlog.RunLog.Info("Starting OS watcher.")
	signalsNotifier := resources.NewSignalNotifier(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sigs := signalsNotifier.Notify()

	for {
		s := <-sigs
		switch s {
		case syscall.SIGHUP:
			hwlog.RunLog.Info("Received SIGHUP, restarting servers.")
			if err := rm.RestartAllServers(); err != nil {
				hwlog.RunLog.Errorf("Unable to restart servers %v", err)
			}
		default:
			hwlog.RunLog.Infof("Received signal \"%v\", shutting down.", s)
			cancel() // Cancel ctx to stop all goroutines
			if stopPeriodicUpdate != nil {
				stopPeriodicUpdate()
			}
			_ = rm.StopAllServers()
			return
		}
	}
}

func initAndStartDevices(deviceType string, createRm func() types.ResourceManager) (types.ResourceManager, func()) {
	hwlog.RunLog.Infof("Initializing %s device resource manager", deviceType)
	rm := createRm()

	if err := rm.ReadConfig(); err != nil {
		hwlog.RunLog.Errorf("%s", err.Error())
	}

	if err := rm.ValidateConfigs(); err != nil {
		hwlog.RunLog.Errorf("Exiting.. one or more invalid %s configuration(s) given: %v", deviceType, err)
	}

	if deviceType == "PCI" {
		if err := rm.ValidateRdmaSystemMode(); err != nil {
			hwlog.RunLog.Errorf("Exiting.. can not change RDMA system mode: %v", err)
		}
	}

	if err := rm.DiscoverHostDevices(); err != nil {
		hwlog.RunLog.Errorf("Error: error discovering %s host devices %v \n", deviceType, err)
	}

	if err := rm.InitServers(); err != nil {
		hwlog.RunLog.Errorf("Error: initializing %s resource servers %v \n", deviceType, err)
	}

	if err := rm.StartAllServers(); err != nil {
		hwlog.RunLog.Errorf("Error: starting %s resource servers %v\n", deviceType, err.Error())
	}

	return rm, rm.PeriodicUpdate()
}

func startFaultDetection(ctx context.Context, ubRm ub_device.UbResourceManager, faultDetectPeriod int) {
	hwlog.RunLog.Infof("Fault detection HCA list from UB devices: %v", ubRm.GetHcaNames())

	if faultDetectPeriod < common.MinFaultDetectionPeriod {
		hwlog.RunLog.Warnf("Fault detection disabled, period %d below min %d",
			faultDetectPeriod, common.MinFaultDetectionPeriod)
		return
	}

	hwlog.RunLog.Infof("Fault detection period: %d seconds, starting...", faultDetectPeriod)

	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		hwlog.RunLog.Errorf("Failed to get in-cluster config for fault reporting: %v", err)
		return
	}
	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to create k8s client for fault reporting: %v", err)
		return
	}

	go fault.StartFaultDetection(ctx, func() []string {
		return ubRm.GetHcaNames()
	}, ubRm.GetHcaDiscoverChan(), faultDetectPeriod)
	go fault.StartFaultReporting(ctx, k8sClient)
}
