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

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
)

// ubDevice implements UbDevice interface to get UB device specific information
type ubDevice struct {
	ubID       string
	deviceName string
	vendor     string
	deviceID   string
	driver     string
	ifName     string
	linkType   string
	rdmaSpec   []*pluginapi.DeviceSpec
}

// NewUbDevice returns an instance of UbDevice interface
func NewUbDevice(ubID, deviceName, vendor, deviceID, driver, ifName, linkType string, rds types.RdmaDeviceSpec) (types.UbDevice, error) {
	rdmaSpec := rds.Get(ubID)
	if err := rds.VerifyRdmaSpec(rdmaSpec); err != nil {
		return nil, fmt.Errorf("missing RDMA device spec for UB device %s, %v", ubID, err)
	}

	return &ubDevice{
		ubID:       ubID,
		deviceName: deviceName,
		vendor:     vendor,
		deviceID:   deviceID,
		driver:     driver,
		ifName:     ifName,
		linkType:   linkType,
		rdmaSpec:   rdmaSpec,
	}, nil
}

func (ud *ubDevice) GetUbID() string {
	return ud.ubID
}

func (ud *ubDevice) GetDeviceName() string {
	return ud.deviceName
}

func (ud *ubDevice) GetVendor() string {
	return ud.vendor
}

func (ud *ubDevice) GetDeviceID() string {
	return ud.deviceID
}

func (ud *ubDevice) GetDriver() string {
	return ud.driver
}

func (ud *ubDevice) GetIfName() string {
	return ud.ifName
}

func (ud *ubDevice) GetLinkType() string {
	return ud.linkType
}

func (ud *ubDevice) GetRdmaSpec() []*pluginapi.DeviceSpec {
	return ud.rdmaSpec
}

func (ud *ubDevice) GetName() string {
	// Return the UB device ID as the name
	return ud.ubID
}
