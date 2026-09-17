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
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/hccn"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/utils/logger"
)

type netInfoStatusCache struct {
	chip      colcommon.HuaWeiAIChip
	timestamp time.Time
	extInfo   *common.NpuNetStatusInfo
}

type netInfoNPUStatusCache struct {
	chip      colcommon.HuaWeiAIChip
	timestamp time.Time
	extInfo   []*common.NpuNetStatusInfo
}

// NetworkLinkCollector collects the network link status info.
type NetworkLinkCollector struct {
	colcommon.MetricsCollectorAdapter
}

// IsSupported check if the collector is supported
func (c *NetworkLinkCollector) IsSupported(n *colcommon.NpuCollector) bool {
	return isNetworkSupported(n, colcommon.GetCacheKey(c))
}

// IsParallel returns true so NetworkLinkCollector always runs in parallel goroutines.
func (c *NetworkLinkCollector) IsParallel(n *colcommon.NpuCollector) bool {
	c.DcmiSupported = false
	logger.Infof("[NetworkLinkCollector] isParallel: true")
	return true
}

// Describe description of the metric
func (c *NetworkLinkCollector) Describe(ch chan<- *prometheus.Desc) {
	if colcommon.DevType == api.Ascend910A5 {
		initDesc(ch, linkStatusDesc)
		initDesc(ch, npuChipPortLinkSpeedDesc)
		addNetWorkStatusLegacyMetricsDesc(ch)
		return
	}
	ch <- npuChipLinkSpeed
	ch <- npuChipLinkUpNum
	ch <- descLinkStatus
}

// CollectToCache collect the metric to cache
func (c *NetworkLinkCollector) CollectToCache(n *colcommon.NpuCollector, chipList []colcommon.HuaWeiAIChip) {
	if colcommon.DevType == api.Ascend910A5 {
		for _, chip := range chipList {
			netInfos := collectNetworkNpuStatusInfo(chip.LogicID)
			c.LocalCache.Store(chip.PhyId, netInfoNPUStatusCache{chip: chip, timestamp: time.Now(), extInfo: netInfos})
		}
		colcommon.UpdateCache[netInfoNPUStatusCache](n, colcommon.GetCacheKey(c), &c.LocalCache)
		return
	}
	for _, chip := range chipList {
		netInfo := collectNetworkStatusInfo(chip.PhyId)
		c.LocalCache.Store(chip.PhyId, netInfoStatusCache{chip: chip, timestamp: time.Now(), extInfo: netInfo})
	}
	colcommon.UpdateCache[netInfoStatusCache](n, colcommon.GetCacheKey(c), &c.LocalCache)
}

// UpdatePrometheus update prometheus metrics
func (c *NetworkLinkCollector) UpdatePrometheus(ch chan<- prometheus.Metric, n *colcommon.NpuCollector,
	containerMap map[int32][]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) {
	if colcommon.DevType == api.Ascend910A5 {
		updateSingleChipNpu := func(chipWithVnpu colcommon.HuaWeiAIChip, cache netInfoNPUStatusCache, cardLabel []string) {
			promUpdateNetInfoStatus(ch, cache, cardLabel)
		}
		updateFrame[netInfoNPUStatusCache](colcommon.GetCacheKey(c), n, containerMap, chips, updateSingleChipNpu)
		return
	}
	updateSingleChip := func(chipWithVnpu colcommon.HuaWeiAIChip, cache netInfoStatusCache, cardLabel []string) {
		netInfo := cache.extInfo
		if netInfo == nil {
			return
		}
		if validateNotNilForEveryElement(netInfo.LinkSpeedInfo) {
			doUpdateMetricWithValidateNum(ch, cache.timestamp, netInfo.LinkSpeedInfo.Speed, cardLabel, npuChipLinkSpeed)
		}
		if validateNotNilForEveryElement(netInfo.LinkStatInfo) {
			doUpdateMetricWithValidateNum(ch, cache.timestamp, netInfo.LinkStatInfo.LinkUPNum, cardLabel, npuChipLinkUpNum)
		}
		if validateNotNilForEveryElement(netInfo.LinkStatusInfo) {
			doUpdateMetricWithValidateNum(ch, cache.timestamp, float64(getLinkStatusCode(netInfo.LinkStatusInfo.LinkState)),
				cardLabel, descLinkStatus)
		}
	}
	updateFrame[netInfoStatusCache](colcommon.GetCacheKey(c), n, containerMap, chips, updateSingleChip)
}

// UpdateTelegraf update telegraf metrics
func (c *NetworkLinkCollector) UpdateTelegraf(ch chan<- colcommon.TelegrafMetric, n *colcommon.NpuCollector,
	containerMap map[int32][]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) {
	if colcommon.DevType == api.Ascend910A5 {
		caches := colcommon.GetInfoFromCache[netInfoNPUStatusCache](n, colcommon.GetCacheKey(c))
		for _, chip := range chips {
			cache, ok := caches[chip.PhyId]
			if !ok {
				continue
			}
			metric := colcommon.NewDeviceMetric(cache.chip.LogicID)
			telegrafUpdateNetInfoStatus(cache, metric.Fields)
			ch <- metric
		}
		return
	}
	caches := colcommon.GetInfoFromCache[netInfoStatusCache](n, colcommon.GetCacheKey(c))
	for _, chip := range chips {
		cache, ok := caches[chip.PhyId]
		if !ok {
			continue
		}
		netInfo := cache.extInfo
		if netInfo == nil {
			continue
		}
		metric := colcommon.NewDeviceMetric(cache.chip.LogicID)
		if validateNotNilForEveryElement(netInfo.LinkSpeedInfo) {
			doUpdateTelegrafWithValidateNum(metric.Fields, npuChipLinkSpeed, netInfo.LinkSpeedInfo.Speed, "")
		}
		if validateNotNilForEveryElement(netInfo.LinkStatInfo) {
			doUpdateTelegrafWithValidateNum(metric.Fields, npuChipLinkUpNum, netInfo.LinkStatInfo.LinkUPNum, "")
		}
		if validateNotNilForEveryElement(netInfo.LinkStatusInfo) {
			doUpdateTelegrafWithValidateNum(metric.Fields, descLinkStatus,
				float64(getLinkStatusCode(netInfo.LinkStatusInfo.LinkState)), "")
		}
		ch <- metric
	}
}

// collectNetworkStatusInfo collects link status info for non-Npu devices.
func collectNetworkStatusInfo(phyID int32) *common.NpuNetStatusInfo {
	statusInfo := &common.NpuNetStatusInfo{LinkStatusInfo: &common.LinkStatusInfo{}}
	if linkState, err := hccn.GetNPULinkStatus(phyID); err == nil {
		statusInfo.LinkStatusInfo.LinkState = linkState
		hwlog.ResetErrCnt(colcommon.DomainForLinkState, phyID)
	} else {
		logErrMetricsWithLimit(colcommon.DomainForLinkState, phyID, err)
		statusInfo.LinkStatusInfo.LinkState = colcommon.Unknown
	}
	if linkUpNum, err := hccn.GetNPULinkUpNum(phyID); err == nil {
		statusInfo.LinkStatInfo = &common.LinkStatInfo{LinkUPNum: float64(linkUpNum)}
		hwlog.ResetErrCnt(colcommon.DomainForLinkStat, phyID)
	} else {
		statusInfo.LinkStatInfo = nil
		logErrMetricsWithLimit(colcommon.DomainForLinkStat, phyID, err)
	}
	if speed, err := hccn.GetNPULinkSpeed(phyID); err == nil {
		statusInfo.LinkSpeedInfo = &common.LinkSpeedInfo{Speed: float64(speed)}
		hwlog.ResetErrCnt(colcommon.DomainForLinkSpeed, phyID)
	} else {
		statusInfo.LinkSpeedInfo = nil
		logErrMetricsWithLimit(colcommon.DomainForLinkSpeed, phyID, err)
	}
	return statusInfo
}

// collectNetworkNpuStatusInfo collects link status info for Npu devices.
func collectNetworkNpuStatusInfo(logicID int32) []*common.NpuNetStatusInfo {
	var newNetInfo []*common.NpuNetStatusInfo
	// udie only has 0 and 1
	dieIDs := []int{0, 1}
	for _, dieID := range dieIDs {
		portIDs, ok := colcommon.NpuDevPortInfos.GetPortMap()[dieID]
		if !ok || len(portIDs) == 0 {
			continue
		}
		for _, port := range portIDs {
			netInfo := &common.NpuNetStatusInfo{
				LinkStatusInfo: &common.LinkStatusInfo{},
				LinkSpeedInfo:  &common.LinkSpeedInfo{},
				Udie:           dieID,
				Port:           port.PortID,
			}
			if linkState, err := hccn.GetNPULinkStatusNpu(logicID, int32(dieID), int32(port.PortID)); err == nil {
				netInfo.LinkStatusInfo.LinkState = linkState
				hwlog.ResetErrCnt(fmt.Sprint(colcommon.DomainForLinkState, dieID, port.PortID), logicID)
			} else {
				logWarnMetricsWithLimit(fmt.Sprint(colcommon.DomainForLinkState, dieID, port.PortID), logicID,
					dieID, port.PortID, err)
				netInfo.LinkStatusInfo.LinkState = colcommon.Unknown
			}
			if speed, err := hccn.GetNPULinkSpeedNpu(logicID, int32(dieID), int32(port.PortID)); err == nil {
				netInfo.LinkSpeedInfo.Speed = float64(speed)
				hwlog.ResetErrCnt(fmt.Sprint(colcommon.DomainForLinkSpeed, dieID, port.PortID), logicID)
			} else {
				netInfo.LinkSpeedInfo = nil
				logWarnMetricsWithLimit(fmt.Sprint(colcommon.DomainForLinkSpeed, dieID, port.PortID), logicID,
					dieID, port.PortID, err)
			}
			newNetInfo = append(newNetInfo, netInfo)
		}
	}
	return newNetInfo
}

func promUpdateNetInfoStatus(ch chan<- prometheus.Metric, cache netInfoNPUStatusCache, cardLabel []string) {
	netInfo := cache.extInfo
	if netInfo == nil {
		return
	}
	timestamp := cache.timestamp
	for i := 0; i < len(netInfo); i++ {
		extendedLabel := append(cardLabel, strconv.Itoa(netInfo[i].Udie), strconv.Itoa(netInfo[i].Port))
		if validateNotNilForEveryElement(netInfo[i].LinkStatusInfo) {
			doUpdateMetricWithValidateNum(ch, timestamp, float64(getLinkStatusCode(netInfo[i].LinkStatusInfo.LinkState)),
				extendedLabel, linkStatusDesc)
		}
		if validateNotNilForEveryElement(netInfo[i].LinkSpeedInfo) {
			doUpdateMetricWithValidateNum(ch, timestamp, netInfo[i].LinkSpeedInfo.Speed, extendedLabel,
				npuChipPortLinkSpeedDesc)
		}
		promUpdateNetInfoStatusLegacy(ch, timestamp, netInfo[i], extendedLabel, i)
	}
}

func telegrafUpdateNetInfoStatus(cache netInfoNPUStatusCache, fieldMap map[string]interface{}) {
	netInfo := cache.extInfo
	if netInfo == nil {
		return
	}
	for i := 0; i < len(netInfo); i++ {
		extInfo := fmt.Sprint("_", netInfo[i].Udie, "_", netInfo[i].Port)
		if validateNotNilForEveryElement(netInfo[i].LinkStatusInfo) {
			doUpdateTelegrafWithValidateNum(fieldMap, linkStatusDesc,
				float64(getLinkStatusCode(netInfo[i].LinkStatusInfo.LinkState)), extInfo)
		}
		if validateNotNilForEveryElement(netInfo[i].LinkSpeedInfo) {
			doUpdateTelegrafWithValidateNum(fieldMap, npuChipPortLinkSpeedDesc, netInfo[i].LinkSpeedInfo.Speed, extInfo)
		}
	}
}

// promUpdateNetInfoStatusLegacy emits legacy link status/speed metrics for backward compatibility.
func promUpdateNetInfoStatusLegacy(ch chan<- prometheus.Metric, timestamp time.Time,
	netInfo *common.NpuNetStatusInfo, extendedLabel []string, i int) {
	if !colcommon.EnableLegacyMetrics {
		return
	}
	if validateNotNilForEveryElement(netInfo.LinkStatusInfo) {
		tryEmitNetworkLegacyMetric(ch, timestamp, float64(getLinkStatusCode(netInfo.LinkStatusInfo.LinkState)),
			extendedLabel, linkStatusLegacyDescs, i)
	}
	if validateNotNilForEveryElement(netInfo.LinkSpeedInfo) {
		tryEmitNetworkLegacyMetric(ch, timestamp, netInfo.LinkSpeedInfo.Speed, extendedLabel,
			npuChipPortLinkSpeedLegacyDescs, i)
	}
}

func addNetWorkStatusLegacyMetricsDesc(ch chan<- *prometheus.Desc) {
	if !colcommon.EnableLegacyMetrics {
		return
	}
	for _, desc := range linkStatusLegacyDescs {
		ch <- desc
	}
	for _, desc := range npuChipPortLinkSpeedLegacyDescs {
		ch <- desc
	}
}
