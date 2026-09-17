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

type netInfoBandwidthCache struct {
	chip      colcommon.HuaWeiAIChip
	timestamp time.Time
	extInfo   *common.NpuNetBandwidthInfo
}

type netInfoNPUBandwidthCache struct {
	chip      colcommon.HuaWeiAIChip
	timestamp time.Time
	extInfo   []*common.NpuNetBandwidthInfo
}

// NetworkBandwidthCollector collects the network real-time bandwidth info.
type NetworkBandwidthCollector struct {
	colcommon.MetricsCollectorAdapter
}

// IsSupported check if the collector is supported
func (c *NetworkBandwidthCollector) IsSupported(n *colcommon.NpuCollector) bool {
	return isNetworkSupported(n, colcommon.GetCacheKey(c))
}

// IsParallel returns true so NetworkBandwidthCollector always runs in parallel goroutines.
func (c *NetworkBandwidthCollector) IsParallel(n *colcommon.NpuCollector) bool {
	c.DcmiSupported = false
	logger.Infof("[NetworkBandwidthCollector] isParallel: true")
	return true
}

// Describe description of the metric
func (c *NetworkBandwidthCollector) Describe(ch chan<- *prometheus.Desc) {
	if colcommon.DevType == api.Ascend910A5 {
		initDesc(ch, bandwidthTxDesc)
		initDesc(ch, bandwidthRxDesc)
		addNetWorkBandwidthLegacyMetricsDesc(ch)
		return
	}
	ch <- descBandwidthTx
	ch <- descBandwidthRx
}

// CollectToCache collect the metric to cache
func (c *NetworkBandwidthCollector) CollectToCache(n *colcommon.NpuCollector, chipList []colcommon.HuaWeiAIChip) {
	if colcommon.DevType == api.Ascend910A5 {
		for _, chip := range chipList {
			netInfos := collectNetworkNpuBandwidthInfo(chip.LogicID)
			c.LocalCache.Store(chip.PhyId, netInfoNPUBandwidthCache{chip: chip, timestamp: time.Now(), extInfo: netInfos})
		}
		colcommon.UpdateCache[netInfoNPUBandwidthCache](n, colcommon.GetCacheKey(c), &c.LocalCache)
		return
	}
	for _, chip := range chipList {
		netInfo := collectNetworkBandwidthInfo(chip.PhyId)
		c.LocalCache.Store(chip.PhyId, netInfoBandwidthCache{chip: chip, timestamp: time.Now(), extInfo: netInfo})
	}
	colcommon.UpdateCache[netInfoBandwidthCache](n, colcommon.GetCacheKey(c), &c.LocalCache)
}

// UpdatePrometheus update prometheus metrics
func (c *NetworkBandwidthCollector) UpdatePrometheus(ch chan<- prometheus.Metric, n *colcommon.NpuCollector,
	containerMap map[int32][]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) {
	if colcommon.DevType == api.Ascend910A5 {
		updateSingleChipNpu := func(chipWithVnpu colcommon.HuaWeiAIChip, cache netInfoNPUBandwidthCache,
			cardLabel []string) {
			promUpdateNetInfoBandwidth(ch, cache, cardLabel)
		}
		updateFrame[netInfoNPUBandwidthCache](colcommon.GetCacheKey(c), n, containerMap, chips, updateSingleChipNpu)
		return
	}
	updateSingleChip := func(chipWithVnpu colcommon.HuaWeiAIChip, cache netInfoBandwidthCache, cardLabel []string) {
		netInfo := cache.extInfo
		if netInfo == nil || netInfo.BandwidthInfo == nil {
			return
		}
		doUpdateMetricWithValidateNum(ch, cache.timestamp, netInfo.BandwidthInfo.TxValue, cardLabel, descBandwidthTx)
		doUpdateMetricWithValidateNum(ch, cache.timestamp, netInfo.BandwidthInfo.RxValue, cardLabel, descBandwidthRx)
	}
	updateFrame[netInfoBandwidthCache](colcommon.GetCacheKey(c), n, containerMap, chips, updateSingleChip)
}

// UpdateTelegraf update telegraf metrics
func (c *NetworkBandwidthCollector) UpdateTelegraf(ch chan<- colcommon.TelegrafMetric, n *colcommon.NpuCollector,
	containerMap map[int32][]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) {
	if colcommon.DevType == api.Ascend910A5 {
		caches := colcommon.GetInfoFromCache[netInfoNPUBandwidthCache](n, colcommon.GetCacheKey(c))
		for _, chip := range chips {
			cache, ok := caches[chip.PhyId]
			if !ok {
				continue
			}
			metric := colcommon.NewDeviceMetric(cache.chip.LogicID)
			telegrafUpdateNetInfoBandwidth(cache, metric.Fields)
			ch <- metric
		}
		return
	}
	caches := colcommon.GetInfoFromCache[netInfoBandwidthCache](n, colcommon.GetCacheKey(c))
	for _, chip := range chips {
		cache, ok := caches[chip.PhyId]
		if !ok {
			continue
		}
		netInfo := cache.extInfo
		if netInfo == nil || netInfo.BandwidthInfo == nil {
			continue
		}
		metric := colcommon.NewDeviceMetric(cache.chip.LogicID)
		doUpdateTelegrafWithValidateNum(metric.Fields, descBandwidthTx, netInfo.BandwidthInfo.TxValue, "")
		doUpdateTelegrafWithValidateNum(metric.Fields, descBandwidthRx, netInfo.BandwidthInfo.RxValue, "")
		ch <- metric
	}
}

// isNetworkSupported check whether the current hardware supports network metrics.
func isNetworkSupported(n *colcommon.NpuCollector, cacheKey string) bool {
	if colcommon.DevType == api.Ascend910A5 {
		mainBoardID := n.Dmgr.GetMainBoardId()
		if notSupportedNetworkNpuDevices[mainBoardID] {
			logForUnSupportDevice(false, colcommon.DevType, cacheKey,
				fmt.Sprint("this mainBoardId:", mainBoardID, " is not supported"))
			return false
		}
		initNpuNetWorkDesc()
		return true
	}
	isSupport := n.Dmgr.IsTrainingCard()
	logForUnSupportDevice(isSupport, colcommon.DevType, cacheKey, "only training card supports network related info")
	return isSupport
}

// collectNetworkBandwidthInfo collects real-time bandwidth info for non-Npu devices.
func collectNetworkBandwidthInfo(phyID int32) *common.NpuNetBandwidthInfo {
	bandwidthInfo := &common.NpuNetBandwidthInfo{BandwidthInfo: &common.BandwidthInfo{}}
	if tx, rx, err := hccn.GetNPUInterfaceTraffic(phyID); err == nil {
		bandwidthInfo.BandwidthInfo.RxValue = rx
		bandwidthInfo.BandwidthInfo.TxValue = tx
		hwlog.ResetErrCnt(colcommon.DomainForBandwidth, phyID)
	} else {
		bandwidthInfo.BandwidthInfo = nil
		logErrMetricsWithLimit(colcommon.DomainForBandwidth, phyID, err)
	}
	return bandwidthInfo
}

// collectNetworkNpuBandwidthInfo collects real-time bandwidth info for Npu devices.
func collectNetworkNpuBandwidthInfo(logicID int32) []*common.NpuNetBandwidthInfo {
	var newNetInfo []*common.NpuNetBandwidthInfo
	// udie only has 0 and 1
	dieIDs := []int{0, 1}
	for _, dieID := range dieIDs {
		portIDs, ok := colcommon.NpuDevPortInfos.GetPortMap()[dieID]
		if !ok || len(portIDs) == 0 {
			continue
		}
		for _, port := range portIDs {
			netInfo := &common.NpuNetBandwidthInfo{
				BandwidthInfo: &common.BandwidthInfo{},
				Udie:          dieID,
				Port:          port.PortID,
			}
			if tx, rx, err := hccn.GetNPUInterfaceTrafficNpu(logicID, int32(dieID), port); err == nil {
				netInfo.BandwidthInfo.RxValue = rx
				netInfo.BandwidthInfo.TxValue = tx
				hwlog.ResetErrCnt(fmt.Sprint(colcommon.DomainForBandwidth, dieID, port.PortID), logicID)
			} else {
				netInfo.BandwidthInfo = nil
				logWarnMetricsWithLimit(fmt.Sprint(colcommon.DomainForBandwidth, dieID, port.PortID), logicID,
					dieID, port.PortID, err)
			}
			newNetInfo = append(newNetInfo, netInfo)
		}
	}
	return newNetInfo
}

func promUpdateNetInfoBandwidth(ch chan<- prometheus.Metric, cache netInfoNPUBandwidthCache,
	cardLabel []string) {
	netInfo := cache.extInfo
	if netInfo == nil {
		return
	}
	timestamp := cache.timestamp
	for i := 0; i < len(netInfo); i++ {
		extendedLabel := append(cardLabel, strconv.Itoa(netInfo[i].Udie), strconv.Itoa(netInfo[i].Port))
		if validateNotNilForEveryElement(netInfo[i].BandwidthInfo) {
			doUpdateMetricWithValidateNum(ch, timestamp, netInfo[i].BandwidthInfo.TxValue, extendedLabel, bandwidthTxDesc)
			doUpdateMetricWithValidateNum(ch, timestamp, netInfo[i].BandwidthInfo.RxValue, extendedLabel, bandwidthRxDesc)
		}
		promUpdateNetInfoBandwidthLegacy(ch, timestamp, netInfo[i], extendedLabel, i)
	}
}

func telegrafUpdateNetInfoBandwidth(cache netInfoNPUBandwidthCache, fieldMap map[string]interface{}) {
	netInfo := cache.extInfo
	if netInfo == nil {
		return
	}
	for i := 0; i < len(netInfo); i++ {
		extInfo := fmt.Sprint("_", netInfo[i].Udie, "_", netInfo[i].Port)
		if validateNotNilForEveryElement(netInfo[i].BandwidthInfo) {
			doUpdateTelegrafWithValidateNum(fieldMap, bandwidthTxDesc, netInfo[i].BandwidthInfo.TxValue, extInfo)
			doUpdateTelegrafWithValidateNum(fieldMap, bandwidthRxDesc, netInfo[i].BandwidthInfo.RxValue, extInfo)
		}
	}
}

// promUpdateNetInfoBandwidthLegacy emits legacy bandwidth metrics for backward compatibility.
func promUpdateNetInfoBandwidthLegacy(ch chan<- prometheus.Metric, timestamp time.Time,
	netInfo *common.NpuNetBandwidthInfo, extendedLabel []string, i int) {
	if !colcommon.EnableLegacyMetrics {
		return
	}
	if validateNotNilForEveryElement(netInfo.BandwidthInfo) {
		tryEmitNetworkLegacyMetric(ch, timestamp, netInfo.BandwidthInfo.TxValue, extendedLabel, bandwidthTxLegacyDescs, i)
		tryEmitNetworkLegacyMetric(ch, timestamp, netInfo.BandwidthInfo.RxValue, extendedLabel, bandwidthRxLegacyDescs, i)
	}
}

func addNetWorkBandwidthLegacyMetricsDesc(ch chan<- *prometheus.Desc) {
	if !colcommon.EnableLegacyMetrics {
		return
	}
	for _, desc := range bandwidthTxLegacyDescs {
		ch <- desc
	}
	for _, desc := range bandwidthRxLegacyDescs {
		ch <- desc
	}
}
