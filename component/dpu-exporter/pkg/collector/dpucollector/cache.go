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

package dpucollector

import (
	"sync"

	"huawei.com/dpu-exporter/pkg/device"
)

// DpuData holds the metrics for all DPU cards (global metrics from hinicadm5).
// Per the class diagram, DpuData has its own RWMutex for fine-grained locking.
type DpuData struct {
	Mu          sync.RWMutex
	CardMetrics map[string]map[string]float64 // keyed by card name → metric name → value
}

// IfaceData holds the metrics for all network interfaces (from sysfs).
// Per the class diagram, IfaceData has its own RWMutex for fine-grained locking.
type IfaceData struct {
	Mu           sync.RWMutex
	IfaceMetrics map[string]map[string]float64 // keyed by eth name → metric name → value
}

// DpuCache is the central in-memory cache for dpu-exporter.
// Per the class diagram, it holds DpuData, IfaceData, and DPU list with a top-level RWMutex.
// It implements metricscollector.CacheAccessor interface.
type DpuCache struct {
	mu        sync.RWMutex
	dpuList   []device.DPU
	dpuData   *DpuData
	ifaceData *IfaceData
}

// NewDpuCache creates a new DpuCache instance
func NewDpuCache() *DpuCache {
	return &DpuCache{
		dpuData: &DpuData{
			CardMetrics: make(map[string]map[string]float64),
		},
		ifaceData: &IfaceData{
			IfaceMetrics: make(map[string]map[string]float64),
		},
	}
}

// --- CacheAccessor interface implementation ---

// GetDpuList returns the DPU device list (read, called by Collect goroutine)
func (c *DpuCache) GetDpuList() []device.DPU {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.dpuList == nil {
		return make([]device.DPU, 0)
	}
	return c.dpuList
}

// SetDpuList replaces the DPU device list (write, called by InitDpuList goroutine)
func (c *DpuCache) SetDpuList(dpuList []device.DPU) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dpuList = dpuList
}

// GetDpuMetrics returns global metrics for a DPU card (read)
func (c *DpuCache) GetDpuMetrics(cardName string) map[string]float64 {
	c.dpuData.Mu.RLock()
	defer c.dpuData.Mu.RUnlock()

	data, ok := c.dpuData.CardMetrics[cardName]
	if !ok {
		return make(map[string]float64)
	}
	result := make(map[string]float64, len(data))
	for k, v := range data {
		result[k] = v
	}
	return result
}

// SetDpuMetrics stores global metrics for a DPU card (write, called by hinicadm5 collector)
func (c *DpuCache) SetDpuMetrics(cardName string, metrics map[string]float64) {
	c.dpuData.Mu.Lock()
	defer c.dpuData.Mu.Unlock()

	existing, ok := c.dpuData.CardMetrics[cardName]
	if !ok {
		existing = make(map[string]float64)
		c.dpuData.CardMetrics[cardName] = existing
	}
	for k, v := range metrics {
		existing[k] = v
	}
}

// GetIfaceMetrics returns metrics for a network interface (read)
func (c *DpuCache) GetIfaceMetrics(ethName string) map[string]float64 {
	c.ifaceData.Mu.RLock()
	defer c.ifaceData.Mu.RUnlock()

	data, ok := c.ifaceData.IfaceMetrics[ethName]
	if !ok {
		return make(map[string]float64)
	}
	result := make(map[string]float64, len(data))
	for k, v := range data {
		result[k] = v
	}
	return result
}

// SetIfaceMetrics stores metrics for a network interface (write, called by sysfs collector)
func (c *DpuCache) SetIfaceMetrics(ethName string, metrics map[string]float64) {
	c.ifaceData.Mu.Lock()
	defer c.ifaceData.Mu.Unlock()

	existing, ok := c.ifaceData.IfaceMetrics[ethName]
	if !ok {
		existing = make(map[string]float64)
		c.ifaceData.IfaceMetrics[ethName] = existing
	}
	for k, v := range metrics {
		existing[k] = v
	}
}
