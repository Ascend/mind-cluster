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
	"fmt"
	"sync"
	"time"

	"huawei.com/dpu-exporter/utils/logger"
)

// collectorIntervalMap stores per-collector intervals, keyed by cache key (collector struct name).
// Written by config manager during config load/reload, read by scheduler.
var collectorIntervalMap sync.Map // map[string]time.Duration

const (
	// CollectOnceInterval sentinel: collect only once then disable
	CollectOnceInterval = -1 * time.Second
	// DisabledInterval sentinel: collector is disabled
	DisabledInterval = -2 * time.Second
	// DefaultGroupInterval is the default interval when not configured or invalid
	DefaultGroupInterval = 60 * time.Second

	// MinIntervalSeconds is the minimum valid collection interval in seconds
	MinIntervalSeconds = 20
	// MaxIntervalSeconds is the maximum valid collection interval in seconds
	MaxIntervalSeconds = 3600

	// DefaultHinicadm5IntervalSeconds is the default hinicadm5 collection interval in seconds
	DefaultHinicadm5IntervalSeconds = 40
	// DefaultSysfsIntervalSeconds is the default sysfs collection interval in seconds
	DefaultSysfsIntervalSeconds = 20
	// DefaultDpuListRefreshIntervalSeconds is the default DPU list refresh interval in seconds
	DefaultDpuListRefreshIntervalSeconds = 60

	// CacheKeyHinicadm5 is the cache key for hinicadm5 collectors
	CacheKeyHinicadm5 = "Hinicadm5Collector"
	// CacheKeySysfs is the cache key for sysfs collectors
	CacheKeySysfs = "SysfsCollector"
	// CacheKeyDpuListRefresh is the cache key for DPU list refresh
	CacheKeyDpuListRefresh = "DpuListRefresh"
)

// ValidateInterval validates an interval value (in seconds) and returns
// the corresponding time.Duration.
//
// Valid values:
//   - -1: CollectOnceInterval (collect only once)
//   - 20~3600: normal collection interval in seconds
//
// Invalid values return an error. The caller should use the default
// interval when validation fails.
func ValidateInterval(seconds int) (time.Duration, error) {
	switch {
	case seconds == -1:
		return CollectOnceInterval, nil
	case seconds < MinIntervalSeconds || seconds > MaxIntervalSeconds:
		return 0, fmt.Errorf(
			"interval %d is invalid, valid values: -1 or [%d, %d]",
			seconds, MinIntervalSeconds, MaxIntervalSeconds)
	default:
		return time.Duration(seconds) * time.Second, nil
	}
}

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

// Apply writes the interval config into the global collectorIntervalMap.
// Invalid values are replaced with each collector's own default and a warning is logged.
func (c *IntervalConfig) Apply() {
	applyValidatedInterval(CacheKeyHinicadm5, c.Hinicadm5CollectorInterval, DefaultHinicadm5IntervalSeconds)
	applyValidatedInterval(CacheKeySysfs, c.SysfsCollectorInterval, DefaultSysfsIntervalSeconds)
	applyValidatedInterval(CacheKeyDpuListRefresh, c.DpuListRefreshInterval, DefaultDpuListRefreshIntervalSeconds)
}

func applyValidatedInterval(cacheKey string, seconds int, defaultSeconds int) {
	d, err := ValidateInterval(seconds)
	if err != nil {
		logger.Warnf("invalid interval for %s: %v, using default %ds", cacheKey, err, defaultSeconds)
		SetCollectorInterval(cacheKey, time.Duration(defaultSeconds)*time.Second)
		return
	}
	SetCollectorInterval(cacheKey, d)
}

// GetDpuListRefreshInterval returns the DPU list refresh interval
func GetDpuListRefreshInterval(fallback time.Duration) time.Duration {
	return GetCollectorInterval(CacheKeyDpuListRefresh, fallback)
}
