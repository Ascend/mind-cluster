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

// Package common for common info
package common

import pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

// hasHostPath reports whether any device in the list uses the given host path
func hasHostPath(deviceList []*pluginapi.DeviceSpec, hostPath string) bool {
	for _, dev := range deviceList {
		if dev.HostPath == hostPath {
			return true
		}
	}

	return false
}

// DevicesChanged check whether the host path sets of the two device lists differ
func DevicesChanged(deviceList, newDeviceList []*pluginapi.DeviceSpec) bool {
	if len(deviceList) != len(newDeviceList) {
		return true
	}

	for _, oldDev := range deviceList {
		if !hasHostPath(newDeviceList, oldDev.HostPath) {
			return true
		}
	}

	return false
}
