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

package resources

import (
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/core"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
)

// NewVendorSelector returns a DeviceSelector interface for vendor list
func NewVendorSelector(vendors []string) types.DeviceSelector {
	return core.NewVendorSelector(vendors)
}

// NewDeviceSelector returns a DeviceSelector interface for device id list
func NewDeviceSelector(devices []string) types.DeviceSelector {
	return core.NewDeviceSelector(devices)
}

// NewIfNameSelector returns a DeviceSelector interface for ifName list
func NewIfNameSelector(ifNames []string) types.DeviceSelector {
	return core.NewIfNameSelector(ifNames)
}

// NewDriverSelector returns a DeviceSelector interface for driver list
func NewDriverSelector(drivers []string) types.DeviceSelector {
	return core.NewDriverSelector(drivers)
}

// NewLinkTypeSelector returns a interface for netDev list
func NewLinkTypeSelector(linkTypes []string) types.DeviceSelector {
	return core.NewLinkTypeSelector(linkTypes)
}
