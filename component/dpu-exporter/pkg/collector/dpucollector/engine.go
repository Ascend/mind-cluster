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
	"context"
	"sync"
	"time"

	"huawei.com/dpu-exporter/pkg/collector/metricscollector"
	"huawei.com/dpu-exporter/pkg/configmanager"
	"huawei.com/dpu-exporter/pkg/device"
	"huawei.com/dpu-exporter/utils/logger"
)

const (
	// defaultDpuListRefreshInterval is the default interval for refreshing DPU device list
	defaultDpuListRefreshInterval = 60 * time.Second
)

var (
	// chainsMu protects the collector chains from concurrent access during hot-reload
	chainsMu sync.RWMutex

	// ChainForDpuMetrics collectors for global DPU-level metrics (per-card parallel)
	ChainForDpuMetrics []metricscollector.MetricsCollector
	// ChainForInterfaceMetrics collectors for interface-level metrics (single goroutine)
	ChainForInterfaceMetrics []metricscollector.MetricsCollector

	// dpuListInitOnce ensures DPU list is initialized at least once before collecting
	dpuListInitOnce sync.Once

	// Collector is the global DpuCollector instance
	Collector *DpuCollector
)

// DpuCollector is the central collection engine for dpu-exporter.
// Per the class diagram, it holds cache, device manager, and orchestrates collection goroutines.
// It implements metricscollector.CollectorContext interface.
type DpuCollector struct {
	cache *DpuCache
	dmgr  device.DeviceManager
}

// NewDpuCollector creates and initializes a DpuCollector instance
func NewDpuCollector(dmgr device.DeviceManager) *DpuCollector {
	c := &DpuCollector{
		cache: NewDpuCache(),
		dmgr:  dmgr,
	}
	Collector = c
	return c
}

// GetDmgr returns the device manager (implements CollectorContext)
func (n *DpuCollector) GetDmgr() device.DeviceManager {
	return n.dmgr
}

// GetCache returns the cache instance (implements CollectorContext)
func (n *DpuCollector) GetCache() metricscollector.CacheAccessor {
	return n.cache
}

// GetDpuList returns the DPU list from cache
func (n *DpuCollector) GetDpuList() []device.DPU {
	return n.cache.GetDpuList()
}

// GetDpuMetrics returns global metrics for a DPU card
func (n *DpuCollector) GetDpuMetrics(cardName string) map[string]float64 {
	return n.cache.GetDpuMetrics(cardName)
}

// GetIfaceMetrics returns metrics for a network interface
func (n *DpuCollector) GetIfaceMetrics(ethName string) map[string]float64 {
	return n.cache.GetIfaceMetrics(ethName)
}

// SetChains atomically replaces the collector chains (called during config reload)
func SetChains(dpuMetrics, ifaceMetrics []metricscollector.MetricsCollector) {
	chainsMu.Lock()
	defer chainsMu.Unlock()
	ChainForDpuMetrics = dpuMetrics
	ChainForInterfaceMetrics = ifaceMetrics
}

// GetChainsSnapshot returns a shallow copy of the current collector chains
func GetChainsSnapshot() (dpuMetrics, ifaceMetrics []metricscollector.MetricsCollector) {
	chainsMu.RLock()
	defer chainsMu.RUnlock()
	dpuMetrics = append([]metricscollector.MetricsCollector(nil), ChainForDpuMetrics...)
	ifaceMetrics = append([]metricscollector.MetricsCollector(nil), ChainForInterfaceMetrics...)
	return dpuMetrics, ifaceMetrics
}

// InitDpuList starts a background goroutine that periodically refreshes the DPU device list.
// The refresh interval can be hot-reloaded via config change.
func InitDpuList(group *sync.WaitGroup, ctx context.Context, n *DpuCollector) {
	group.Add(1)
	go func() {
		defer group.Done()

		interval := configmanager.GetDpuListRefreshInterval(defaultDpuListRefreshInterval)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		reloadCh := configmanager.SubscribeReload()
		defer configmanager.UnsubscribeReload(reloadCh)

		// Initial refresh
		refreshDpuList(n)

		for {
			select {
			case <-ctx.Done():
				logger.Info("received stop signal, stop dpu list collect")
				return
			case <-reloadCh:
				configmanager.DrainReloadSignal(reloadCh)
				newInterval := configmanager.GetDpuListRefreshInterval(defaultDpuListRefreshInterval)
				if newInterval != interval {
					ticker.Stop()
					ticker = time.NewTicker(newInterval)
					interval = newInterval
					logger.Infof("dpu list refresh interval updated to %v", newInterval)
				}
			case <-ticker.C:
				refreshDpuList(n)
			}
		}
	}()
}

// refreshDpuList queries the device manager and updates the cache
func refreshDpuList(n *DpuCollector) {
	begin := time.Now()
	dpuList := n.dmgr.GetDpuList()
	n.cache.SetDpuList(dpuList)
	logger.Infof("refresh dpu list, count: %d, time cost: %v", len(dpuList), time.Since(begin))
}

// dpuListInitAtFirstTime ensures the DPU list is populated before collecting starts
func dpuListInitAtFirstTime(n *DpuCollector) {
	dpuListInitOnce.Do(func() {
		dpuList := n.cache.GetDpuList()
		if len(dpuList) == 0 {
			logger.Debug("no dpu list in cache, start to refresh")
			refreshDpuList(n)
		}
	})
}

// StartCollect starts all collection goroutines.
// It launches two types of collectors:
//  1. DPU metrics collectors — one goroutine per DPU card (parallel)
//  2. Interface metrics collectors — single goroutine (reads are very fast)
func StartCollect(group *sync.WaitGroup, ctx context.Context, n *DpuCollector) {
	dpuListInitAtFirstTime(n)
	startCollectDpuMetrics(group, ctx, n)
	startCollectInterfaceMetrics(group, ctx, n)
}

// filterSupported returns only collectors that support the current card type
func filterSupported(collectors []metricscollector.MetricsCollector, ctx metricscollector.CollectorContext) []metricscollector.MetricsCollector {
	supported := make([]metricscollector.MetricsCollector, 0, len(collectors))
	for _, c := range collectors {
		if c != nil && c.IsSupported(ctx) {
			supported = append(supported, c)
		}
	}
	return supported
}

// startCollectDpuMetrics starts one goroutine per DPU card for global DPU metrics collectors.
// Only collectors whose IsSupported() returns true for the current card type will run.
func startCollectDpuMetrics(group *sync.WaitGroup, ctx context.Context, n *DpuCollector) {
	dpuList := n.cache.GetDpuList()
	group.Add(len(dpuList))
	for i := range dpuList {
		go func(cardName string) {
			defer group.Done()
			runCollectorLoop(ctx, n, collectorLoopOptions{
				loadChain: func() []metricscollector.MetricsCollector {
					dpuMetricsChain, _ := GetChainsSnapshot()
					return filterSupported(dpuMetricsChain, n)
				},
				onStop: func() {
					logger.Infof("received stop signal, stop dpu metrics collect for card %s", cardName)
				},
				runDueCollectors: func(dueCollectors []metricscollector.MetricsCollector) {
					logger.Debugf("start dpu metrics collect for card %s, collectors: %v",
						cardName, getCollectorNames(dueCollectors))
					begin := time.Now()
					for _, c := range dueCollectors {
						c.CollectToCache(n)
					}
					logger.Debugf("end dpu metrics collect for card %s, time cost: %v",
						cardName, time.Since(begin))
				},
			})
		}(dpuList[i].CardName)
	}
}

// startCollectInterfaceMetrics starts a single goroutine for interface-level metrics collectors.
// Only collectors whose IsSupported() returns true for the current card type will run.
func startCollectInterfaceMetrics(group *sync.WaitGroup, ctx context.Context, n *DpuCollector) {
	group.Add(1)
	go func() {
		defer group.Done()
		runCollectorLoop(ctx, n, collectorLoopOptions{
			loadChain: func() []metricscollector.MetricsCollector {
				_, ifaceMetricsChain := GetChainsSnapshot()
				return filterSupported(ifaceMetricsChain, n)
			},
			onStop: func() {
				logger.Info("received stop signal, stop interface metrics collect")
			},
			runDueCollectors: func(dueCollectors []metricscollector.MetricsCollector) {
				logger.Debugf("start interface metrics collect, collectors: %v", getCollectorNames(dueCollectors))
				begin := time.Now()
				for _, c := range dueCollectors {
					c.CollectToCache(n)
				}
				logger.Debugf("end interface metrics collect, time cost: %v", time.Since(begin))
			},
		})
	}()
}

// ReloadConfig triggers a config reload notification to all collector goroutines
func ReloadConfig() {
	configmanager.NotifyConfigReload()
}
