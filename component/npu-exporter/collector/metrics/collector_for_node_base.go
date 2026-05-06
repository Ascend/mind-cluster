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

// Package metrics for general collector
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/utils"
	"huawei.com/npu-exporter/v6/utils/logger"
	"huawei.com/npu-exporter/v6/versions"
)

var (
	nodeInfoDesc = common.BuildDescWithLabel("node_base_info", "the common information of this node",
		[]string{exporterVersionLabel, driverVersionLabel})
)

const (
	measurementForNodeBaseInfo = "ascend-nodeBaseInfo"
	exporterVersionLabel       = "exporterVersion"
	driverVersionLabel         = "driverVersion"
)

// NodeBaseCollector collect node base info
type NodeBaseCollector struct {
	common.MetricsCollectorAdapter
}

type nodeBaseInfoCache struct {
	timestamp       time.Time
	exporterVersion string
	driverVersion   string
}

// Describe description of the metric
func (c *NodeBaseCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- nodeInfoDesc
}

// CollectToCache collect the metric to cache
func (c *NodeBaseCollector) CollectToCache(n *common.NpuCollector, chipList []common.HuaWeiAIChip) {
	c.LocalCache.Store(common.GetCacheKey(c), nodeBaseInfoCache{
		timestamp:       time.Now(),
		exporterVersion: versions.BuildVersion,
		driverVersion:   n.Dmgr.GetDcmiVersion(),
	})
}

// UpdatePrometheus update prometheus metric
func (c *NodeBaseCollector) UpdatePrometheus(ch chan<- prometheus.Metric, n *common.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []common.HuaWeiAIChip) {
	nodeBaseInfo, ok := c.LocalCache.Load(common.GetCacheKey(c))
	if !ok {
		logger.Debugf("cacheKey(%v) not found", common.GetCacheKey(c))
		return
	}
	cache, ok := nodeBaseInfo.(nodeBaseInfoCache)
	if !ok {
		logger.Error("cache type mismatch")
		return
	}
	doUpdateMetric(ch, cache.timestamp, 1, []string{cache.exporterVersion, cache.driverVersion}, nodeInfoDesc)
}

// UpdateTelegraf update telegraf metric
func (c *NodeBaseCollector) UpdateTelegraf(fieldsMap map[string]map[string]interface{}, n *common.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []common.HuaWeiAIChip) map[string]map[string]interface{} {
	nodeBaseInfo, ok := c.LocalCache.Load(common.GetCacheKey(c))
	if !ok {
		logger.Debugf("cacheKey(%v) not found", common.GetCacheKey(c))
		return fieldsMap
	}
	cache, ok := nodeBaseInfo.(nodeBaseInfoCache)
	if !ok {
		logger.Error("cache type mismatch")
		return fieldsMap
	}

	labelsMap := make(map[string]string)
	labelsMap[exporterVersionLabel] = cache.exporterVersion
	labelsMap[driverVersionLabel] = cache.driverVersion

	tetegrafData := common.TelegrafData{
		Measurement: measurementForNodeBaseInfo,
		Labels:      labelsMap,
		Metrics:     map[string]interface{}{utils.GetDescName(nodeInfoDesc): 1},
		Timestamp:   cache.timestamp,
	}
	fieldsMap[common.KeyForMetricsWithCustomLabels][tetegrafData.Measurement] = tetegrafData
	return fieldsMap
}
