/* Copyright(C) 2025-2025. Huawei Technologies Co.,Ltd. All rights reserved.
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
)

var (
	// bandwidth
	descBandwidthTx = colcommon.BuildDesc("npu_chip_info_bandwidth_tx",
		"the npu interface transport speed, unit is 'MB/s'")
	descBandwidthRx = colcommon.BuildDesc("npu_chip_info_bandwidth_rx",
		"the npu interface receive speed, unit is 'MB/s'")

	// linkspeed
	npuChipLinkSpeed = colcommon.BuildDesc("npu_chip_link_speed",
		"the npu interface receive link speed, unit is 'Mb/s'")

	// linkupNum
	npuChipLinkUpNum = colcommon.BuildDesc("npu_chip_link_up_num", "the npu interface receive link-up num")

	// linkstatus
	descLinkStatus = colcommon.BuildDesc("npu_chip_info_link_status", "the npu link status")

	// npu specific metrics
	linkStatusDesc           []*prometheus.Desc
	bandwidthTxDesc          []*prometheus.Desc
	bandwidthRxDesc          []*prometheus.Desc
	npuChipPortLinkSpeedDesc []*prometheus.Desc

	notSupportedNetworkNpuDevices = map[uint32]bool{
		api.Atlas3501PMainBoardID: true,
	}
)

const (
	// bandwidthTime bandwidth's time param
	bandwidthTime = 100
)

type netInfoCache struct {
	chip      colcommon.HuaWeiAIChip
	timestamp time.Time
	extInfo   *common.NpuNetInfo
}

type netInfoNPUCache struct {
	chip      colcommon.HuaWeiAIChip
	timestamp time.Time
	extInfo   []*common.NpuNetInfo
}

// NetworkCollector collects the network info
type NetworkCollector struct {
	colcommon.MetricsCollectorAdapter
}

// IsSupported check if the collector is supported
func (c *NetworkCollector) IsSupported(n *colcommon.NpuCollector) bool {
	// For Npu devices, check if it's a supported model
	if colcommon.DevType == api.Ascend910A5 {
		mainBoardID := n.Dmgr.GetMainBoardId()
		if notSupportedNetworkNpuDevices[mainBoardID] {
			logForUnSupportDevice(false, colcommon.DevType, colcommon.GetCacheKey(c),
				fmt.Sprint("this mainBoardId:", mainBoardID, " is not supported"))
			return false
		}
		initNpuNetWorkDesc()
		return true
	}

	// For other devices, check if it's a training card
	isSupport := n.Dmgr.IsTrainingCard()
	logForUnSupportDevice(isSupport, colcommon.DevType, colcommon.GetCacheKey(c),
		"only training card supports network related info")
	return isSupport
}

// Describe description of the metric
func (c *NetworkCollector) Describe(ch chan<- *prometheus.Desc) {
	if colcommon.DevType == api.Ascend910A5 {
		// Npu specific metrics
		for _, desc := range linkStatusDesc {
			ch <- desc
		}
		for _, desc := range bandwidthTxDesc {
			ch <- desc
		}
		for _, desc := range bandwidthRxDesc {
			ch <- desc
		}
		for _, desc := range npuChipPortLinkSpeedDesc {
			ch <- desc
		}
		return
	}
	// Non-Npu metrics
	ch <- descBandwidthTx
	ch <- descBandwidthRx
	ch <- npuChipLinkSpeed
	ch <- npuChipLinkUpNum
	ch <- descLinkStatus
}

// CollectToCache collect the metric to cache
func (c *NetworkCollector) CollectToCache(n *colcommon.NpuCollector, chipList []colcommon.HuaWeiAIChip) {
	if colcommon.DevType == api.Ascend910A5 {
		for _, chip := range chipList {
			// Collect Npu specific network info
			netInfos := collectNetworkNpuInfo(chip.LogicID)
			c.LocalCache.Store(chip.PhyId, netInfoNPUCache{chip: chip, timestamp: time.Now(), extInfo: netInfos})
		}
		colcommon.UpdateCache[netInfoNPUCache](n, colcommon.GetCacheKey(c), &c.LocalCache)
		return
	}
	for _, chip := range chipList {
		// Collect regular network info
		netInfo := collectNetworkInfo(chip.PhyId)
		c.LocalCache.Store(chip.PhyId, netInfoCache{chip: chip, timestamp: time.Now(), extInfo: &netInfo})
	}
	colcommon.UpdateCache[netInfoCache](n, colcommon.GetCacheKey(c), &c.LocalCache)
}

// UpdatePrometheus update prometheus metrics
func (c *NetworkCollector) UpdatePrometheus(ch chan<- prometheus.Metric, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) {
	if colcommon.DevType == api.Ascend910A5 {
		// Update Npu specific metrics
		updateSingleChipNpu := func(chipWithVnpu colcommon.HuaWeiAIChip, cache netInfoNPUCache, cardLabel []string) {
			timestamp := cache.timestamp
			promUpdateNetInfo(ch, cache, timestamp, cardLabel)
		}
		updateFrame[netInfoNPUCache](colcommon.GetCacheKey(c), n, containerMap, chips, updateSingleChipNpu)
		return
	}
	// Update regular metrics
	updateSingleChip := func(chipWithVnpu colcommon.HuaWeiAIChip, cache netInfoCache, cardLabel []string) {
		netInfo := cache.extInfo
		if netInfo == nil {
			return
		}
		timestamp := cache.timestamp
		if validateNotNilForEveryElement(netInfo.BandwidthInfo) {
			doUpdateMetricWithValidateNum(ch, timestamp, netInfo.BandwidthInfo.TxValue, cardLabel, descBandwidthTx)
			doUpdateMetricWithValidateNum(ch, timestamp, netInfo.BandwidthInfo.RxValue, cardLabel, descBandwidthRx)
		}
		if validateNotNilForEveryElement(netInfo.LinkSpeedInfo) {
			doUpdateMetricWithValidateNum(ch, timestamp, netInfo.LinkSpeedInfo.Speed, cardLabel, npuChipLinkSpeed)
		}
		if validateNotNilForEveryElement(netInfo.LinkStatInfo) {
			doUpdateMetricWithValidateNum(ch, timestamp, netInfo.LinkStatInfo.LinkUPNum, cardLabel, npuChipLinkUpNum)
		}
		if validateNotNilForEveryElement(netInfo.LinkStatusInfo) {
			doUpdateMetricWithValidateNum(ch, timestamp, float64(hccn.GetLinkStatusCode(netInfo.LinkStatusInfo.LinkState)),
				cardLabel, descLinkStatus)
		}
	}
	updateFrame[netInfoCache](colcommon.GetCacheKey(c), n, containerMap, chips, updateSingleChip)
}

// UpdateTelegraf update telegraf metrics
func (c *NetworkCollector) UpdateTelegraf(fieldsMap map[string]map[string]interface{}, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) map[string]map[string]interface{} {
	if colcommon.DevType == api.Ascend910A5 {
		// Update Npu specific metrics
		caches := colcommon.GetInfoFromCache[netInfoNPUCache](n, colcommon.GetCacheKey(c))
		for _, chip := range chips {
			cache, ok := caches[chip.PhyId]
			if !ok {
				continue
			}
			fieldMap := getFieldMap(fieldsMap, cache.chip.LogicID)
			telegrafUpdateNetInfo(cache, fieldMap)
		}
		return fieldsMap
	}
	// Update regular metrics
	caches := colcommon.GetInfoFromCache[netInfoCache](n, colcommon.GetCacheKey(c))
	for _, chip := range chips {
		cache, ok := caches[chip.PhyId]
		if !ok {
			continue
		}
		fieldMap := getFieldMap(fieldsMap, cache.chip.LogicID)
		extInfo := cache.extInfo
		if extInfo == nil {
			continue
		}
		if validateNotNilForEveryElement(extInfo.BandwidthInfo) {
			doUpdateTelegrafWithValidateNum(fieldMap, descBandwidthTx, extInfo.BandwidthInfo.TxValue, "")
			doUpdateTelegrafWithValidateNum(fieldMap, descBandwidthRx, extInfo.BandwidthInfo.RxValue, "")
		}
		if validateNotNilForEveryElement(extInfo.LinkSpeedInfo) {
			doUpdateTelegrafWithValidateNum(fieldMap, npuChipLinkSpeed, extInfo.LinkSpeedInfo.Speed, "")
		}
		if validateNotNilForEveryElement(extInfo.LinkStatInfo) {
			doUpdateTelegrafWithValidateNum(fieldMap, npuChipLinkUpNum, extInfo.LinkStatInfo.LinkUPNum, "")
		}
		if validateNotNilForEveryElement(extInfo.LinkStatusInfo) {
			doUpdateTelegrafWithValidateNum(fieldMap, descLinkStatus,
				float64(hccn.GetLinkStatusCode(extInfo.LinkStatusInfo.LinkState)), "")
		}
	}
	return fieldsMap
}

func collectNetworkInfo(phyID int32) common.NpuNetInfo {
	newNetInfo := common.NpuNetInfo{}
	newNetInfo.LinkStatusInfo = &common.LinkStatusInfo{}
	if linkState, err := hccn.GetNPULinkStatus(phyID); err == nil {
		newNetInfo.LinkStatusInfo.LinkState = linkState
		hwlog.ResetErrCnt(colcommon.DomainForLinkState, phyID)
	} else {
		logErrMetricsWithLimit(colcommon.DomainForLinkState, phyID, err)
		newNetInfo.LinkStatusInfo.LinkState = colcommon.Abnormal
	}

	if tx, rx, err := hccn.GetNPUInterfaceTraffic(phyID); err == nil {
		newNetInfo.BandwidthInfo = &common.BandwidthInfo{}
		newNetInfo.BandwidthInfo.RxValue = rx
		newNetInfo.BandwidthInfo.TxValue = tx
		hwlog.ResetErrCnt(colcommon.DomainForBandwidth, phyID)
	} else {
		newNetInfo.BandwidthInfo = nil
		logErrMetricsWithLimit(colcommon.DomainForBandwidth, phyID, err)
	}
	if linkUpNum, err := hccn.GetNPULinkUpNum(phyID); err == nil {
		newNetInfo.LinkStatInfo = &common.LinkStatInfo{}
		newNetInfo.LinkStatInfo.LinkUPNum = float64(linkUpNum)
		hwlog.ResetErrCnt(colcommon.DomainForLinkStat, phyID)
	} else {
		newNetInfo.LinkStatInfo = nil
		logErrMetricsWithLimit(colcommon.DomainForLinkStat, phyID, err)
	}

	if speed, err := hccn.GetNPULinkSpeed(phyID); err == nil {
		newNetInfo.LinkSpeedInfo = &common.LinkSpeedInfo{}
		newNetInfo.LinkSpeedInfo.Speed = float64(speed)
		hwlog.ResetErrCnt(colcommon.DomainForLinkSpeed, phyID)
	} else {
		newNetInfo.LinkSpeedInfo = nil
		logErrMetricsWithLimit(colcommon.DomainForLinkSpeed, phyID, err)
	}

	return newNetInfo
}

// Npu specific collection functions
func collectNetworkNpuInfo(logicID int32) []*common.NpuNetInfo {
	var newNetInfo []*common.NpuNetInfo
	for dieID, portIDs := range colcommon.NpuDevPortInfos.GetPortMap() {
		for _, portID := range portIDs {
			netInfo := common.NpuNetInfo{
				LinkStatusInfo: &common.LinkStatusInfo{},
				BandwidthInfo:  &common.BandwidthInfo{},
				LinkSpeedInfo:  &common.LinkSpeedInfo{},
			}
			if linkState, err := hccn.GetNPULinkStatusNpu(logicID, int32(dieID), int32(portID)); err == nil {
				hwlog.RunLog.Debugf("hccn_tool get npu link status: %s", linkState)
				netInfo.LinkStatusInfo.LinkState = linkState
				hwlog.ResetErrCnt(fmt.Sprint(colcommon.DomainForLinkState, dieID, portID), logicID)
			} else {
				logWarnMetricsWithLimit(fmt.Sprint(colcommon.DomainForLinkState, dieID, portID), logicID, dieID, portID, err)
				netInfo.LinkStatusInfo.LinkState = colcommon.Abnormal
			}
			if tx, rx, err := hccn.GetNPUInterfaceTrafficNpu(logicID, int32(dieID), int32(portID), int32(bandwidthTime)); err == nil {
				netInfo.BandwidthInfo.RxValue = rx
				netInfo.BandwidthInfo.TxValue = tx
				hwlog.ResetErrCnt(fmt.Sprint(colcommon.DomainForBandwidth, dieID, portID), logicID)
			} else {
				netInfo.BandwidthInfo = nil
				logWarnMetricsWithLimit(fmt.Sprint(colcommon.DomainForBandwidth, dieID, portID), logicID, dieID, portID, err)
			}
			if speed, err := hccn.GetNPULinkSpeedNpu(logicID, int32(dieID), int32(portID)); err == nil {
				netInfo.LinkSpeedInfo.Speed = float64(speed)
				hwlog.ResetErrCnt(fmt.Sprint(colcommon.DomainForLinkSpeed, dieID, portID), logicID)
			} else {
				netInfo.LinkSpeedInfo = nil
				logWarnMetricsWithLimit(fmt.Sprint(colcommon.DomainForLinkSpeed, dieID, portID), logicID, dieID, portID, err)
			}
			newNetInfo = append(newNetInfo, &netInfo)
		}
	}

	return newNetInfo
}

func promUpdateNetInfo(ch chan<- prometheus.Metric, cache netInfoNPUCache, timestamp time.Time, cardLabel []string) {
	netInfo := cache.extInfo
	if netInfo == nil {
		return
	}
	for i := 0; i < colcommon.NpuDevPortInfos.GetCount(); i++ {
		if validateNotNilForEveryElement(netInfo[i].LinkStatusInfo) {
			doUpdateMetricWithValidateNum(ch, timestamp, float64(hccn.GetLinkStatusCode(netInfo[i].LinkStatusInfo.LinkState)),
				cardLabel, linkStatusDesc[i])
		}
		if validateNotNilForEveryElement(netInfo[i].BandwidthInfo) {
			doUpdateMetricWithValidateNum(ch, timestamp, netInfo[i].BandwidthInfo.TxValue, cardLabel, bandwidthTxDesc[i])
			doUpdateMetricWithValidateNum(ch, timestamp, netInfo[i].BandwidthInfo.RxValue, cardLabel, bandwidthRxDesc[i])
		}
		if validateNotNilForEveryElement(netInfo[i].LinkSpeedInfo) {
			doUpdateMetricWithValidateNum(ch, timestamp, netInfo[i].LinkSpeedInfo.Speed, cardLabel, npuChipPortLinkSpeedDesc[i])
		}
	}
}

func telegrafUpdateNetInfo(cache netInfoNPUCache, fieldMap map[string]interface{}) {
	netInfo := cache.extInfo
	if netInfo == nil {
		return
	}
	for i := 0; i < colcommon.NpuDevPortInfos.GetCount(); i++ {
		if validateNotNilForEveryElement(netInfo[i].LinkStatusInfo) {
			doUpdateTelegrafWithValidateNum(fieldMap, linkStatusDesc[i],
				float64(hccn.GetLinkStatusCode(netInfo[i].LinkStatusInfo.LinkState)), "")
		}
		if validateNotNilForEveryElement(netInfo[i].BandwidthInfo) {
			doUpdateTelegrafWithValidateNum(fieldMap, bandwidthTxDesc[i], netInfo[i].BandwidthInfo.TxValue, "")
			doUpdateTelegrafWithValidateNum(fieldMap, bandwidthRxDesc[i], netInfo[i].BandwidthInfo.RxValue, "")
		}
		if validateNotNilForEveryElement(netInfo[i].LinkSpeedInfo) {
			doUpdateTelegrafWithValidateNum(fieldMap, npuChipPortLinkSpeedDesc[i], netInfo[i].LinkSpeedInfo.Speed, "")
		}
	}
}

func initNpuNetWorkDesc() {
	for dieID, portIDs := range colcommon.NpuDevPortInfos.GetPortMap() {
		for _, portID := range portIDs {
			colcommon.BuildDescSlice(&linkStatusDesc, fmt.Sprint(api.MetricsPrefix, "link_status_", strconv.Itoa(dieID),
				"_", strconv.Itoa(portID)), fmt.Sprint("the npu link status ", "dieId:", strconv.Itoa(dieID),
				" portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&bandwidthTxDesc, fmt.Sprint(api.MetricsPrefix, "bandwidth_tx_", strconv.Itoa(dieID),
				"_", strconv.Itoa(portID)), fmt.Sprint("the npu port transport speed, unit is 'MB/s' ", "dieId:",
				strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&bandwidthRxDesc, fmt.Sprint(api.MetricsPrefix, "bandwidth_rx_", strconv.Itoa(dieID),
				"_", strconv.Itoa(portID)), fmt.Sprint("the npu port receive speed, unit is 'MB/s' ", "dieId:",
				strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&npuChipPortLinkSpeedDesc, fmt.Sprint(api.MetricsPrefix, "link_speed_",
				strconv.Itoa(dieID), "_", strconv.Itoa(portID)), fmt.Sprint("the npu port link speed, unit is 'G' ",
				"dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
		}
	}
}
