/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package device

import "ascend-dynamic-resource-allocation/pkg/consts"

// NpuDeviceKind identifies how a device is exposed to Kubernetes.
type NpuDeviceKind string

const (
	// PhysicalDevice is a complete physical NPU.
	PhysicalDevice NpuDeviceKind = consts.PhysicalNPUDeviceKind
	// StaticVNPUDevice was created before the DRA driver started.
	StaticVNPUDevice NpuDeviceKind = consts.StaticVNPUDeviceKind
)

// NpuAllInfo aggregates all discovered NPU devices and their distinct types.
type NpuAllInfo struct {
	AllDevTypes []string
	AllDevs     []*NpuDevice
	AICoreDevs  []*NpuDevice
}

// NpuDevice is the in-memory representation of a single Ascend device.
type NpuDevice struct {
	DevType      string
	DeviceName   string
	IP           string
	LogicID      int32
	PhyID        int32
	CardID       int32
	DeviceID     int32
	Kind         NpuDeviceKind
	VDevID       uint32
	TemplateName string
	VNPUType     string
	AICore       int64
}
