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

package configmanager

import (
	"sync"
	"time"
)

// collectorIntervalMap stores per-collector intervals, keyed by cache key (collector struct name).
// Written by config manager during config load/reload, read by scheduler.
var collectorIntervalMap sync.Map // map[string]time.Duration

const (
	// CollectOnceInterval sentinel: collect only once then disable
	CollectOnceInterval = -1 * time.Second
	// DisabledInterval sentinel: collector is disabled
	DisabledInterval = -2 * time.Second
	// DefaultGroupInterval is the default interval when not configured
	DefaultGroupInterval = 60 * time.Second

	// CacheKeyHinicadm5 is the cache key for hinicadm5 collectors
	CacheKeyHinicadm5 = "Hinicadm5Collector"
	// CacheKeySysfs is the cache key for sysfs collectors
	CacheKeySysfs = "SysfsCollector"
	// CacheKeyDpuListRefresh is the cache key for DPU list refresh
	CacheKeyDpuListRefresh = "DpuListRefresh"
)

// SetCollectorInterval stores a collector's interval in the global map
func SetCollectorInterval(cacheKey string, interval time.Duration) {
	if cacheKey == "" {
		return
	}
	if interval == CollectOnceInterval {
		collectorIntervalMap.Store(cacheKey, interval)
		return
	}
	if interval <= 0 {
		return
	}
	collectorIntervalMap.Store(cacheKey, interval)
}

// GetCollectorInterval loads a collector's interval from the global map
func GetCollectorInterval(cacheKey string, fallback time.Duration) time.Duration {
	if cacheKey == "" {
		return fallback
	}
	v, ok := collectorIntervalMap.Load(cacheKey)
	if !ok {
		return fallback
	}
	d, ok := v.(time.Duration)
	if !ok {
		return fallback
	}
	if d == CollectOnceInterval {
		return d
	}
	if d <= 0 {
		return fallback
	}
	return d
}

// IntervalConfig holds the configured collection intervals
type IntervalConfig struct {
	// Hinicadm5CollectorInterval for hinicadm5 collectors, in seconds
	Hinicadm5CollectorInterval int
	// SysfsCollectorInterval for sysfs collectors, in seconds
	SysfsCollectorInterval int
	// DpuListRefreshInterval for DPU device list refresh, in seconds
	DpuListRefreshInterval int
}

// Apply writes the interval config into the global collectorIntervalMap
func (c *IntervalConfig) Apply() {
	SetCollectorInterval(CacheKeyHinicadm5,
		time.Duration(c.Hinicadm5CollectorInterval)*time.Second)
	SetCollectorInterval(CacheKeySysfs,
		time.Duration(c.SysfsCollectorInterval)*time.Second)
	SetCollectorInterval(CacheKeyDpuListRefresh,
		time.Duration(c.DpuListRefreshInterval)*time.Second)
}

// GetDpuListRefreshInterval returns the DPU list refresh interval
func GetDpuListRefreshInterval(fallback time.Duration) time.Duration {
	return GetCollectorInterval(CacheKeyDpuListRefresh, fallback)
}
