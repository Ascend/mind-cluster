// Copyright 2025 NVIDIA CORPORATION & AFFILIATES
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

package types

import (
	"net"
	"os"

	"github.com/vishvananda/netlink"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// Selectors contains common device selectors fields
type Selectors struct {
	Buses     []string `json:"buses,omitempty"`
	Vendors   []string `json:"vendors,omitempty"`
	DeviceIDs []string `json:"deviceIDs,omitempty"`
	Drivers   []string `json:"drivers,omitempty"`
	IfNames   []string `json:"ifNames,omitempty"`
	LinkTypes []string `json:"linkTypes,omitempty"`
}

// UserConfig configuration for device plugin
type UserConfig struct {
	ResourceName   string    `json:"resourceName"`
	ResourcePrefix string    `json:"resourcePrefix"`
	RdmaHcaMax     int       `json:"rdmaHcaMax"`
	Devices        []string  `json:"devices"`
	Selectors      Selectors `json:"selectors"`
}

// UserConfigList config list for servers
type UserConfigList struct {
	PeriodicUpdateInterval *int         `json:"periodicUpdateInterval"`
	ConfigList             []UserConfig `json:"configList"`
}

// ResourceServer is gRPC server implements K8s device plugin api
type ResourceServer interface {
	pluginapi.DevicePluginServer
	Start() error
	Stop() error
	Restart() error
	Watch()
	UpdateDevices([]Device)
}

// ResourceManager manager multi plugins
type ResourceManager interface {
	ReadConfig() error
	ValidateConfigs() error
	ValidateRdmaSystemMode() error
	DiscoverHostDevices() error
	GetDevices() []Device
	InitServers() error
	StartAllServers() error
	StopAllServers() error
	RestartAllServers() error
	GetFilteredDevices(devices []Device, selector *Selectors) []Device
	PeriodicUpdate() func()
}

// ResourceServerPort to connect the resources server to k8s
type ResourceServerPort interface {
	GetServer() *grpc.Server
	CreateServer()
	DeleteServer()
	Listen(string, string) (net.Listener, error)
	Serve(net.Listener)
	Stop()
	Close(*grpc.ClientConn)
	Register(pluginapi.RegistrationClient, *pluginapi.RegisterRequest) error
	GetClientConn(string) (*grpc.ClientConn, error)
}

// NotifierFactory register signals to listen for
type SignalNotifier interface {
	Notify() chan os.Signal
}

// RdmaDeviceSpec used to find the rdma devices
type RdmaDeviceSpec interface {
	Get(string) []*pluginapi.DeviceSpec
	VerifyRdmaSpec([]*pluginapi.DeviceSpec) error
}

// Device is a generic interface for all device types
type Device interface {
	GetName() string
	GetVendor() string
	GetDeviceID() string
	GetDriver() string
	GetRdmaSpec() []*pluginapi.DeviceSpec
	GetIfName() string
	GetLinkType() string
}

// PciNetDevice provides an interface to get PCI network device specific information
type PciNetDevice interface {
	Device
	GetPciAddr() string
}

// UbDevice provides an interface to get UB device specific information
type UbDevice interface {
	Device
	GetUbID() string
	GetDeviceName() string
}

// DeviceSelector provides an interface for filtering a list of devices
type DeviceSelector interface {
	Filter([]Device) []Device
}

// NetlinkManager is an interface to mock netlink library
type NetlinkManager interface {
	LinkByName(string) (netlink.Link, error)
	LinkSetUp(netlink.Link) error
}
