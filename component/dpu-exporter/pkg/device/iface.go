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

package device

// Interface represents a network interface belonging to a DPU.
// Per the class diagram, Interface carries its own metrics map and accessor method.
type Interface struct {
	// EthName e.g. "ens1f0"
	EthName string
	// HcaName the HCA name associated with this interface
	HcaName string
	// SysfsPath e.g. "/sys/class/net/ens1f0"
	SysfsPath string
	// MacAddr MAC address (optional, may be empty)
	MacAddr string
	// metrics holds interface-level metrics (from sysfs)
	metrics map[string]float64
}

// NewInterface creates an Interface with initialized metrics map
func NewInterface(ethName, hcaName, sysfsPath string) Interface {
	return Interface{
		EthName:   ethName,
		HcaName:   hcaName,
		SysfsPath: sysfsPath,
		metrics:   make(map[string]float64),
	}
}

// GetMetric returns the metric value for the given key, or 0 if not found
func (i *Interface) GetMetric(key string) float64 {
	return i.metrics[key]
}

// SetMetrics replaces all metrics for this interface
func (i *Interface) SetMetrics(metrics map[string]float64) {
	i.metrics = metrics
}

// GetMetrics returns a copy of all metrics
func (i *Interface) GetMetrics() map[string]float64 {
	cp := make(map[string]float64, len(i.metrics))
	for k, v := range i.metrics {
		cp[k] = v
	}
	return cp
}
