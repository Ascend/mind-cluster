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
	"path"
	"strings"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/utils"
)

type ubRdmaDeviceSpec struct {
	rdmaDevs []string
}

func newUbRdmaDeviceSpec(rdmaDevs []string) types.RdmaDeviceSpec {
	return &ubRdmaDeviceSpec{rdmaDevs: rdmaDevs}
}

func (rf *ubRdmaDeviceSpec) Get(ubID string) []*pluginapi.DeviceSpec {
	rdmaDevices := utils.GetRdmaDevicesForUbdev(ubID)
	deviceSpec := make([]*pluginapi.DeviceSpec, 0, len(rdmaDevices))
	for _, device := range rdmaDevices {
		deviceSpec = append(deviceSpec, &pluginapi.DeviceSpec{
			HostPath:      device,
			ContainerPath: device,
			Permissions:   "rwm",
		})
	}

	return deviceSpec
}

func (rf *ubRdmaDeviceSpec) VerifyRdmaSpec(rdmaDevSpecs []*pluginapi.DeviceSpec) error {
	for _, rdmaDev := range rf.rdmaDevs {
		if !containsUbRdmaDev(rdmaDevSpecs, rdmaDev) {
			return fmt.Errorf("RDMA device %q not found", rdmaDev)
		}
	}

	return nil
}

func containsUbRdmaDev(devSpecs []*pluginapi.DeviceSpec, rdmaDev string) bool {
	for _, devSpec := range devSpecs {
		_, devSpecName := path.Split(devSpec.HostPath)
		if strings.Contains(devSpecName, rdmaDev) {
			return true
		}
	}

	return false
}
