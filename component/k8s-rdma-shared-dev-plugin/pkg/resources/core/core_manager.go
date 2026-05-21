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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/vishvananda/netlink"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/utils"
)

const (
	// Default periodic update interval
	defaultPeriodicUpdateInterval = 60 * time.Second
	// RDMA subsystem network namespace mode
	rdmaExclusive = "exclusive"
)

var (
	activeSockDir = "/var/lib/kubelet/plugins_registry"
)

// CoreResourceManager defines the core functionality that all resource managers should have
type CoreResourceManager interface {
	// ReadConfig reads configuration from file
	ReadConfig() error
	// ValidateConfigs validates the configuration
	ValidateConfigs() error
	// ValidateRdmaSystemMode ensures RDMA subsystem network namespace mode is shared
	ValidateRdmaSystemMode() error
	// InitServers initializes the resource servers
	InitServers() error
	// StartAllServers starts all resource servers
	StartAllServers() error
	// StopAllServers stops all resource servers
	StopAllServers() error
	// RestartAllServers restarts all resource servers
	RestartAllServers() error
	// PeriodicUpdate returns a function that performs periodic updates
	PeriodicUpdate() func()

	// GetConfigList returns the list of configurations
	GetConfigList() []*types.UserConfig
	// SetConfigList sets the list of configurations
	SetConfigList(configs []*types.UserConfig)
	// GetResourceServers returns the list of resource servers
	GetResourceServers() []types.ResourceServer
	// AddResourceServer adds a resource server
	AddResourceServer(server types.ResourceServer)
	// GetUseCdi returns whether CDI is enabled
	GetUseCdi() bool
	// GetPeriodicUpdateInterval returns the periodic update interval
	GetPeriodicUpdateInterval() time.Duration
}

// coreResourceManager implements CoreResourceManager interface
type coreResourceManager struct {
	configFile             string
	defaultResourcePrefix  string
	socketSuffix           string
	watchMode              bool
	configList             []*types.UserConfig
	resourceServers        []types.ResourceServer
	netlinkManager         types.NetlinkManager
	PeriodicUpdateInterval time.Duration
	useCdi                 bool
}

// NewCoreResourceManager returns a new instance of CoreResourceManager
func NewCoreResourceManager(configFile string, defaultResourcePrefix string, socketSuffix string,
	useCdi bool) CoreResourceManager {
	return &coreResourceManager{
		configFile:             configFile,
		defaultResourcePrefix:  defaultResourcePrefix,
		socketSuffix:           socketSuffix,
		watchMode:              detectPluginWatchMode(activeSockDir),
		resourceServers:        []types.ResourceServer{},
		PeriodicUpdateInterval: defaultPeriodicUpdateInterval,
		useCdi:                 useCdi,
	}
}

// ReadConfig reads configuration from file
func (crm *coreResourceManager) ReadConfig() error {
	log.Println("Reading", crm.configFile)
	raw, err := os.ReadFile(crm.configFile)
	if err != nil {
		log.Printf("Warning: Failed to read config file %s: %v", crm.configFile, err)
		log.Println("Using default configuration")
		return crm.useDefaultConfig()
	}

	config := &types.UserConfigList{}
	if err := json.Unmarshal(raw, config); err != nil {
		log.Printf("Warning: Failed to parse config file %s: %v", crm.configFile, err)
		log.Println("Using default configuration")
		return crm.useDefaultConfig()
	}

	log.Printf("loaded config: %+v \n", config.ConfigList)

	// if periodic update is not set then use the default value
	if config.PeriodicUpdateInterval == nil {
		log.Println("no periodic update interval is set, use default interval 60 seconds")
		crm.PeriodicUpdateInterval = defaultPeriodicUpdateInterval
	} else {
		PeriodicUpdateInterval := *config.PeriodicUpdateInterval
		if PeriodicUpdateInterval == 0 {
			log.Println("warning: periodic update interval is 0, no periodic update will run")
		} else {
			log.Printf("periodic update interval: %+d \n", PeriodicUpdateInterval)
		}
		crm.PeriodicUpdateInterval = time.Duration(PeriodicUpdateInterval) * time.Second
	}

	crm.configList = make([]*types.UserConfig, len(config.ConfigList))
	for i := range config.ConfigList {
		crm.configList[i] = &config.ConfigList[i]
	}
	return nil
}

// useDefaultConfig sets up a default configuration when config file is missing or invalid
func (crm *coreResourceManager) useDefaultConfig() error {
	defaultConfig := &types.UserConfig{
		ResourceName:   "rdma",
		ResourcePrefix: crm.defaultResourcePrefix,
		RdmaHcaMax:     1000,
		Devices:        []string{},
		Selectors: types.Selectors{
			Vendors:   []string{},
			DeviceIDs: []string{},
			Drivers:   []string{},
		},
	}
	crm.configList = []*types.UserConfig{defaultConfig}
	return nil
}

// ValidateConfigs validates the configuration
func (crm *coreResourceManager) ValidateConfigs() error {
	resourceName := make(map[string]string)

	if crm.PeriodicUpdateInterval < 0 {
		return fmt.Errorf("invalid \"periodicUpdateInterval\" configuration \"%d\"", crm.PeriodicUpdateInterval)
	}

	if len(crm.configList) < 1 {
		return fmt.Errorf("no resources configuration found")
	}

	for _, conf := range crm.configList {
		// check if name contains acceptable characters
		if !validResourceName(conf.ResourceName) {
			return fmt.Errorf("error: resource name \"%s\" contains invalid characters", conf.ResourceName)
		}
		// check resource names are unique
		_, ok := resourceName[conf.ResourceName]
		if ok {
			// resource name already exist
			return fmt.Errorf("error: resource name \"%s\" already exists", conf.ResourceName)
		}
		// If prefix is not configured - use the default one
		if conf.ResourcePrefix == "" {
			conf.ResourcePrefix = crm.defaultResourcePrefix
		}

		if !validResourcePrefix(conf.ResourcePrefix) {
			return fmt.Errorf("error: resource prefix \"%s\" contains invalid characters, "+
				"must be a valid DNS subdomain (lowercase alphanumeric, hyphens, dots only)", conf.ResourcePrefix)
		}

		if conf.RdmaHcaMax < 0 {
			return fmt.Errorf("error: Invalid value for rdmaHcaMax < 0: %d", conf.RdmaHcaMax)
		}

		isEmptySelector := utils.IsEmptySelector(&(conf.Selectors))
		if isEmptySelector && len(conf.Devices) == 0 {
			return fmt.Errorf("error: configuration mismatch. neither \"selectors\" nor \"devices\" fields exits," +
				" it is recommended to use the new “selectors” field")
		}

		// If both "selectors" and "devices" fields are provided then fail with confusion
		if !isEmptySelector && len(conf.Devices) > 0 {
			return fmt.Errorf("configuration mismatch. Cannot specify both \"selectors\" and \"devices\" fields")
		} else if isEmptySelector { // If no "selector" then use devices as IfNames selector
			log.Println("Warning: \"devices\" field is deprecated, it is recommended to use the new “selectors” field")
			conf.Selectors.IfNames = conf.Devices
		}

		resourceName[conf.ResourceName] = conf.ResourceName
	}

	return nil
}

// ValidateRdmaSystemMode ensure RDMA subsystem network namespace mode is shared
func (crm *coreResourceManager) ValidateRdmaSystemMode() error {
	mode, err := netlink.RdmaSystemGetNetnsMode()
	if err != nil {
		if err.Error() == "invalid argument" {
			log.Printf("too old kernel to get RDMA subsystem")
			return nil
		}

		return fmt.Errorf("can not get RDMA subsystem network namespace mode")
	}

	if mode == rdmaExclusive {
		return fmt.Errorf("incorrect RDMA subsystem network namespace")
	}
	return nil
}

// InitServers initializes the resource servers
func (crm *coreResourceManager) InitServers() error {
	// This is a placeholder implementation
	// Device-specific resource managers should override this method
	log.Println("CoreResourceManager.InitServers() called")
	return nil
}

// StartAllServers starts all resource servers
func (crm *coreResourceManager) StartAllServers() error {
	for _, rs := range crm.resourceServers {
		if err := rs.Start(); err != nil {
			return err
		}

		// start watcher
		if !crm.watchMode {
			go rs.Watch()
		}
	}
	return nil
}

// StopAllServers stops all resource servers
func (crm *coreResourceManager) StopAllServers() error {
	for _, rs := range crm.resourceServers {
		if err := rs.Stop(); err != nil {
			return err
		}
	}
	return nil
}

// RestartAllServers restarts all resource servers
func (crm *coreResourceManager) RestartAllServers() error {
	if err := crm.StopAllServers(); err != nil {
		return err
	}
	return crm.StartAllServers()
}

// PeriodicUpdate returns a function that performs periodic updates
func (crm *coreResourceManager) PeriodicUpdate() func() {
	stopChan := make(chan interface{})
	if crm.PeriodicUpdateInterval > 0 {
		ticker := time.NewTicker(crm.PeriodicUpdateInterval)
		go func() {
			for {
				select {
				case <-ticker.C:
					log.Println("periodic update triggered")
					// This should be overridden by device-specific managers
				case <-stopChan:
					ticker.Stop()
					return
				}
			}
		}()
	}

	return func() {
		close(stopChan)
	}
}

// GetConfigList returns the list of configurations
func (crm *coreResourceManager) GetConfigList() []*types.UserConfig {
	return crm.configList
}

// SetConfigList sets the list of configurations
func (crm *coreResourceManager) SetConfigList(configs []*types.UserConfig) {
	crm.configList = configs
}

// GetResourceServers returns the list of resource servers
func (crm *coreResourceManager) GetResourceServers() []types.ResourceServer {
	return crm.resourceServers
}

// AddResourceServer adds a resource server
func (crm *coreResourceManager) AddResourceServer(server types.ResourceServer) {
	crm.resourceServers = append(crm.resourceServers, server)
}

// GetUseCdi returns whether CDI is enabled
func (crm *coreResourceManager) GetUseCdi() bool {
	return crm.useCdi
}

// GetPeriodicUpdateInterval returns the periodic update interval
func (crm *coreResourceManager) GetPeriodicUpdateInterval() time.Duration {
	return crm.PeriodicUpdateInterval
}

// validResourceName returns true if the name is valid
func validResourceName(name string) bool {
	var validString = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return validString.MatchString(name)
}

// validResourcePrefix returns true if the prefix is a valid DNS subdomain
func validResourcePrefix(prefix string) bool {
	var validString = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*$`)
	return validString.MatchString(prefix)
}

// detectPluginWatchMode detects if the plugin should use watch mode
func detectPluginWatchMode(activeSockDir string) bool {
	if _, err := os.Stat(activeSockDir); err == nil {
		return true
	}
	return false
}
