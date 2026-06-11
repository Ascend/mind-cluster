/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package config for general collector
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/metrics"
	"huawei.com/npu-exporter/v6/utils/logger"
)

var (
	// singleGoroutineMap metrics in this map will be collected in single goroutine
	singleGoroutineMap = map[string]common.MetricsCollector{
		groupHccs:        &metrics.HccsCollector{},
		groupNpu:         &metrics.BaseInfoCollector{},
		groupUtilization: &metrics.UtilizationCollector{},
		groupSio:         &metrics.SioCollector{},
		groupVersion:     &metrics.VersionCollector{},
		groupHbm:         &metrics.HbmCollector{},
		groupDDR:         &metrics.DdrCollector{},
		groupVnpu:        &metrics.VnpuCollector{},
		groupPcie:        &metrics.PcieCollector{},
		groupNodeBase:    &metrics.NodeBaseCollector{},
	}
	// multiGoroutineMap metrics in this map will be collected in multi goroutine
	multiGoroutineMap = map[string]common.MetricsCollector{
		groupNetwork: &metrics.NetworkCollector{},
		groupRoce:    &metrics.RoceCollector{},
		groupOptical: &metrics.OpticalCollector{},
		groupUb:      &metrics.UbCollector{},
	}
	// pluginCollectorMap metrics in this map will be collected in plugin goroutine
	pluginCollectorMap = map[string]common.MetricsCollector{}
	presetConfigs      = make([]MetricsGroupConfig, 0)
	pluginConfigs      = make([]MetricsGroupConfig, 0)
)

const (
	defaultIntervalSeconds = 60
	intervalSeconds1       = 1
	intervalSeconds5       = 5
	intervalSeconds10      = 10
	intervalSeconds30      = 30
	maxIntervalSeconds     = 86400 // 1 day = 24 * 60 * 60 seconds

	groupDDR     = "ddr"
	groupHccs    = "hccs"
	groupNpu     = "npu"
	groupNetwork = "network"
	groupPcie    = "pcie"
	groupRoce    = "roce"
	groupSio     = "sio"
	groupVnpu    = "vnpu"
	groupVersion = "version"
	groupOptical = "optical"
	groupHbm     = "hbm"
	// groupText represents text-based metrics collected by plugin collectors
	groupText        = "text"
	groupUb          = "ub"
	groupNodeBase    = "nodeBase"
	groupUtilization = "utilization"

	stateOn  = "ON"
	stateOFF = "OFF"
)

var (
	defaultPresetConfigs = []MetricsGroupConfig{
		// dcmi
		buildDefaultConfig(groupVersion, stateOn, -1),
		buildDefaultConfig(groupUtilization, stateOn, 1),
		buildDefaultConfig(groupNpu, stateOn, intervalSeconds5),
		buildDefaultConfig(groupDDR, stateOn, intervalSeconds10),
		buildDefaultConfig(groupSio, stateOn, defaultIntervalSeconds),
		buildDefaultConfig(groupHbm, stateOn, defaultIntervalSeconds),
		buildDefaultConfig(groupHccs, stateOn, defaultIntervalSeconds),
		buildDefaultConfig(groupPcie, stateOn, defaultIntervalSeconds),
		buildDefaultConfig(groupVnpu, stateOn, defaultIntervalSeconds),
		buildDefaultConfig(groupNodeBase, stateOn, maxIntervalSeconds),
		// hccn_tool
		buildDefaultConfig(groupRoce, stateOn, defaultIntervalSeconds),
		buildDefaultConfig(groupOptical, stateOn, defaultIntervalSeconds),
		buildDefaultConfig(groupNetwork, stateOn, defaultIntervalSeconds),
		buildDefaultConfig(groupUb, stateOn, defaultIntervalSeconds),
	}
	defaultPluginConfigs = []MetricsGroupConfig{
		buildDefaultConfig(groupText, stateOn, defaultIntervalSeconds),
	}
)

const (
	// PresetConfigPath is the path to the preset metrics configuration file.
	// NOTE: Changed from "/usr/local/metricConfiguration.json" to "/user/mind-cluster/npu-exporter-config/metricConfiguration.json"
	// to support ConfigMap hot-reload without subPath or items. This is a compatibility change that needs to be documented
	// in detailed design documents.
	PresetConfigPath = "/user/mind-cluster/npu-exporter-config/metricConfiguration.json"
	// PluginConfigPath is the path to the plugin metrics configuration file.
	// NOTE: Same compatibility change as PresetConfigPath.
	PluginConfigPath = "/user/mind-cluster/npu-exporter-config/pluginConfiguration.json"
)

// MetricsGroupConfig represents the configuration of a metrics group
type MetricsGroupConfig struct {
	MetricsGroup string `json:"metricsGroup"`
	State        string `json:"state"`
	// IntervalSeconds is the collection interval in seconds.
	// Uses pointer type to distinguish "not configured" from "configured as 0":
	//   - nil: not configured, use default value
	//   - points to 0: explicitly set to 0, treated as invalid
	//   - points to -1: collect only once
	//   - points to other values: use the configured value
	IntervalSeconds *int `json:"intervalSeconds,omitempty"`
}

// buildDefaultConfig creates a MetricsGroupConfig with the specified metrics group name, state, and interval.
func buildDefaultConfig(metricsGroup string, state string, intervalSeconds int) MetricsGroupConfig {
	return MetricsGroupConfig{
		MetricsGroup:    metricsGroup,
		State:           state,
		IntervalSeconds: &intervalSeconds,
	}
}

func loadConfiguration() {
	if fileBytes := loadFromFile(PresetConfigPath); fileBytes == nil {
		logger.Warnf("load config from file %s failed, use default config", PresetConfigPath)
		presetConfigs = defaultPresetConfigs
	} else {
		initConfiguration(fileBytes, &presetConfigs)
	}
	if fileBytes := loadFromFile(PluginConfigPath); fileBytes == nil {
		logger.Warnf("load config from file %s failed, use default config", PluginConfigPath)
		pluginConfigs = defaultPluginConfigs
	} else {
		initConfiguration(fileBytes, &pluginConfigs)
	}
}

func loadFromFile(filePath string) []byte {
	// K8s ConfigMap uses symlink mechanism:
	//   metricConfiguration.json -> ..data/metricConfiguration.json
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		logger.Warnf("read config file %s failed: %v", filePath, err)
		return nil
	}
	return fileBytes
}

func initConfiguration(fileBytes []byte, configs *[]MetricsGroupConfig) {
	if err := json.Unmarshal(fileBytes, configs); err != nil {
		logger.Errorf("unmarshal config byte failed: %v", err)
		return
	}
}

// AddPluginCollector add plugin collector to cache
func AddPluginCollector(name string, collector common.MetricsCollector) error {
	if _, exist := pluginCollectorMap[name]; exist {
		logger.Errorf("plugin collector %v already exist", name)
		return fmt.Errorf("plugin collector %v already exist", name)
	}
	logger.Infof("add plugin collector %v ok", name)
	pluginCollectorMap[name] = collector
	return nil
}

// DeletePluginCollector delete plugin collector from cache
func DeletePluginCollector(name string) {
	if _, exist := pluginCollectorMap[name]; !exist {
		logger.Warnf("plugin collector %v does not exist", name)
		return
	}
	logger.Infof("delete plugin collector %v ok", name)
	delete(pluginCollectorMap, name)
}

// collectorIntervalEntry records the name and collection interval of a single collector for merged logging
type collectorIntervalEntry struct {
	name     string
	interval time.Duration
}

func logCollectorIntervals(entries []collectorIntervalEntry) {
	if len(entries) == 0 {
		return
	}
	// Group by interval, key is interval, value is list of collector names
	groups := make(map[time.Duration][]string)
	for _, e := range entries {
		groups[e.interval] = append(groups[e.interval], e.name)
	}
	// Extract and sort all intervals for ordered output
	intervals := make([]time.Duration, 0, len(groups))
	for iv := range groups {
		intervals = append(intervals, iv)
	}
	sort.Slice(intervals, func(i, j int) bool { return intervals[i] < intervals[j] })
	for _, iv := range intervals {
		names := groups[iv]
		sort.Strings(names)
		if iv == common.CollectOnceInterval() {
			logger.Infof("collect once: %v", strings.Join(names, ", "))
		} else {
			logger.Infof("collect interval %v: %v", iv, strings.Join(names, ", "))
		}
	}
}

func resolveInterval(configIntervalSeconds *int, fallbackInterval time.Duration) time.Duration {
	if fallbackInterval > 0 {
		return fallbackInterval
	}
	if configIntervalSeconds == nil {
		// Not configured, use default interval
		return time.Duration(defaultIntervalSeconds) * time.Second
	}
	if *configIntervalSeconds == -1 {
		return common.CollectOnceInterval()
	}
	if *configIntervalSeconds == 0 {
		// Explicitly set to 0, treated as invalid
		return common.DisabledInterval()
	}
	if *configIntervalSeconds > maxIntervalSeconds || *configIntervalSeconds < -1 {
		return common.DisabledInterval()
	}
	return time.Duration(*configIntervalSeconds) * time.Second
}

// validatedConfig holds the validated result of a single metrics group config
type validatedConfig struct {
	name     string
	interval time.Duration
}

// validateConfigs validates configs and resolves collection intervals.
// It logs state and interval info, and skips invalid or disabled configs.
func validateConfigs(configs []MetricsGroupConfig, fallbackInterval time.Duration, logPrefix string) []validatedConfig {
	results := make([]validatedConfig, 0, len(configs))
	for _, config := range configs {
		name := config.MetricsGroup
		if strings.ToUpper(config.State) != stateOn {
			logger.Infof("%-18s [%-13v] is off", logPrefix, name)
			continue
		}
		var intervalStr string
		if config.IntervalSeconds == nil {
			intervalStr = fmt.Sprintf("not set, will use default %v", defaultIntervalSeconds)
		} else {
			intervalStr = fmt.Sprintf("%d", *config.IntervalSeconds)
		}
		logger.Infof("%-18s [%-13v] is on, collect interval is %vs", logPrefix, name, intervalStr)
		interval := resolveInterval(config.IntervalSeconds, fallbackInterval)
		if interval == common.DisabledInterval() {
			logger.Warnf("%-18s [%-13v] has invalid intervalSeconds %v, disabled", logPrefix, name, intervalStr)
			continue
		}
		results = append(results, validatedConfig{name: name, interval: interval})
	}
	return results
}

// matchCollectors matches validated configs against the collector map and returns
// the matched collectors and their interval entries.
func matchCollectors(validated []validatedConfig, collectorMap map[string]common.MetricsCollector,
	n *common.NpuCollector) ([]common.MetricsCollector, []collectorIntervalEntry) {

	collectors := make([]common.MetricsCollector, 0)
	entries := make([]collectorIntervalEntry, 0)
	for _, vc := range validated {
		collector, exist := collectorMap[vc.name]
		if exist && collector.IsSupported(n) {
			common.SetCollectorInterval(common.GetCacheKey(collector), vc.interval)
			collectors = append(collectors, collector)
			entries = append(entries, collectorIntervalEntry{name: vc.name, interval: vc.interval})
		}
	}
	return collectors, entries
}

// Register registers collectors to cache. It loads configuration files, determines the collection interval
func Register(n *common.NpuCollector) {
	loadConfiguration()

	fallbackInterval := n.GetUpdateTime()

	presetValidated := validateConfigs(presetConfigs, fallbackInterval, "metricsGroup")
	pluginValidated := validateConfigs(pluginConfigs, fallbackInterval, "plugin collector")

	newSingle, singleEntries := matchCollectors(presetValidated, singleGoroutineMap, n)
	newMulti, multiEntries := matchCollectors(presetValidated, multiGoroutineMap, n)
	newPlugin, pluginEntries := matchCollectors(pluginValidated, pluginCollectorMap, n)

	allEntries := append(append(singleEntries, multiEntries...), pluginEntries...)
	logCollectorIntervals(allEntries)
	common.SetChains(newSingle, newMulti, newPlugin)
	common.NotifyConfigReload()

	logger.Infof("ChainForSingleGoroutine:%#v", newSingle)
	logger.Infof("ChainForMultiGoroutine:%#v", newMulti)
	logger.Infof("ChainForCustomPlugin:%#v", newPlugin)
}

// UnRegister delete collector from chain
func UnRegister(worker reflect.Type) {
	logger.Debugf("unRegister collector:%v", worker)
	unRegisterChain(worker, &common.ChainForSingleGoroutine)
	unRegisterChain(worker, &common.ChainForMultiGoroutine)
	unRegisterChain(worker, &common.ChainForCustomPlugin)
}

func unRegisterChain(worker reflect.Type, chain *[]common.MetricsCollector) {
	newChain := make([]common.MetricsCollector, 0)
	for _, collector := range *chain {
		if reflect.TypeOf(collector) != worker {
			newChain = append(newChain, collector)
		}
	}
	*chain = newChain
}
