/* Copyright(C) 2026-2026. Huawei Technologies Co.,Ltd. All rights reserved.
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
	// linkStatusDesc npu udie port link status
	linkStatusDesc []*prometheus.Desc
	// bandwidth
	bandwidthTxDesc []*prometheus.Desc
	bandwidthRxDesc []*prometheus.Desc
	// linkSpeed
	npuChipPortLinkSpeedDesc []*prometheus.Desc

	notSupportedNetworkA5Devices = map[uint32]bool{
		api.Atlas3501PMainBoardID: true,
	}
)

const (
	// maxDieId max udie id
	maxDieId = 2
	// maxPortId max port id
	maxPortId = 9
	// bandwidthTime bandwidth's time param
	bandwidthTime = 100
)

func init() {
	for dieID := 0; dieID < maxDieId; dieID++ {
		for portID := 0; portID < maxPortId; portID++ {
			colcommon.BuildDescSlice(&linkStatusDesc, fmt.Sprint(api.MetricsPrefix, "link_status_", strconv.Itoa(dieID), "_", strconv.Itoa(portID)),
				fmt.Sprint("the npu link status ", "dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&bandwidthTxDesc, fmt.Sprint(api.MetricsPrefix, "bandwidth_tx_", strconv.Itoa(dieID), "_", strconv.Itoa(portID)),
				fmt.Sprint("the npu port transport speed, unit is 'MB/s' ", "dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&bandwidthRxDesc, fmt.Sprint(api.MetricsPrefix, "bandwidth_rx_", strconv.Itoa(dieID), "_", strconv.Itoa(portID)),
				fmt.Sprint("the npu port receive speed, unit is 'MB/s' ", "dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&npuChipPortLinkSpeedDesc, fmt.Sprint(api.MetricsPrefix, "link_speed_", strconv.Itoa(dieID), "_", strconv.Itoa(portID)),
				fmt.Sprint("the npu port link speed, unit is 'G' ", "dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
		}
	}
}

type netInfoA5Cache struct {
	chip      colcommon.HuaWeiAIChip
	timestamp time.Time
	extInfo   []*common.NpuNetInfo
}

// NetworkA5Collector collects the network info
type NetworkNPUCollector struct {
	colcommon.MetricsCollectorAdapter
}

// IsSupported check if the collector is supported
func (c *NetworkNPUCollector) IsSupported(n *colcommon.NpuCollector) bool {
	devType := n.Dmgr.GetDevType()
	if devType != api.Ascend910A5 {
		logForUnSupportDevice(false, devTypeMap[devType], colcommon.GetCacheKey(c), "")
		return false
	}
	mainBoardID := n.Dmgr.GetMainBoardId()
	if notSupportedNetworkA5Devices[mainBoardID] {
		logForUnSupportDevice(false, devTypeMap[devType], colcommon.GetCacheKey(c),
			fmt.Sprint("this mainBoardId:", mainBoardID, " is not supported"))
		return false
	}
	return true
}

// Describe description of the metric
func (c *NetworkNPUCollector) Describe(ch chan<- *prometheus.Desc) {
	// linkstatus
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
}

// CollectToCache collect the metric to cache
func (c *NetworkNPUCollector) CollectToCache(n *colcommon.NpuCollector, chipList []colcommon.HuaWeiAIChip) {
	for _, chip := range chipList {
		var netInfos []*common.NpuNetInfo
		netInfos = collectNetworkA5Info(chip.LogicID)
		c.LocalCache.Store(chip.PhyId, netInfoA5Cache{chip: chip, timestamp: time.Now(), extInfo: netInfos})
	}
	colcommon.UpdateCache[netInfoA5Cache](n, colcommon.GetCacheKey(c), &c.LocalCache)
}

// UpdatePrometheus update prometheus metrics
func (c *NetworkNPUCollector) UpdatePrometheus(ch chan<- prometheus.Metric, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) {

	updateSingleChip := func(chipWithVnpu colcommon.HuaWeiAIChip, cache netInfoA5Cache, cardLabel []string) {
		timestamp := cache.timestamp
		promUpdateNetInfo(ch, cache, timestamp, cardLabel)
	}
	updateFrame[netInfoA5Cache](colcommon.GetCacheKey(c), n, containerMap, chips, updateSingleChip)
}

func promUpdateNetInfo(ch chan<- prometheus.Metric, cache netInfoA5Cache,
	timestamp time.Time, cardLabel []string) {
	netInfo := cache.extInfo
	if netInfo == nil {
		return
	}
	for i := 0; i < (maxDieId * maxPortId); i++ {
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

// UpdateTelegraf update telegraf metrics
func (c *NetworkNPUCollector) UpdateTelegraf(fieldsMap map[string]map[string]interface{}, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) map[string]map[string]interface{} {

	caches := colcommon.GetInfoFromCache[netInfoA5Cache](n, colcommon.GetCacheKey(c))
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

func telegrafUpdateNetInfo(cache netInfoA5Cache, fieldMap map[string]interface{}) {
	netInfo := cache.extInfo
	if netInfo == nil {
		return
	}
	for i := 0; i < (maxDieId * maxPortId); i++ {
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

func collectNetworkA5Info(logicID int32) []*common.NpuNetInfo {
	var newNetInfo []*common.NpuNetInfo
	for dieID := 0; dieID < maxDieId; dieID++ {
		for portID := 0; portID < maxPortId; portID++ {
			netInfo := common.NpuNetInfo{
				LinkStatusInfo: &common.LinkStatusInfo{},
				BandwidthInfo:  &common.BandwidthInfo{},
				LinkSpeedInfo:  &common.LinkSpeedInfo{},
			}
			if linkState, err := hccn.GetNPULinkStatusA5(logicID, int32(dieID), int32(portID)); err == nil {
				hwlog.RunLog.Debugf("hccn_tool get npu link status: %s", linkState)
				netInfo.LinkStatusInfo.LinkState = linkState
				hwlog.ResetErrCnt(fmt.Sprint(colcommon.DomainForLinkState, dieID, portID), logicID)
			} else {
				logWarnMetricsWithLimit(fmt.Sprint(colcommon.DomainForLinkState, dieID, portID), logicID, dieID, portID, err)
				netInfo.LinkStatusInfo.LinkState = colcommon.Abnormal
			}
			if tx, rx, err := hccn.GetNPUInterfaceTrafficA5(logicID, int32(dieID), int32(portID), bandwidthTime); err == nil {
				netInfo.BandwidthInfo.RxValue = rx
				netInfo.BandwidthInfo.TxValue = tx
				hwlog.ResetErrCnt(fmt.Sprint(colcommon.DomainForBandwidth, dieID, portID), logicID)
			} else {
				netInfo.BandwidthInfo = nil
				logWarnMetricsWithLimit(fmt.Sprint(colcommon.DomainForBandwidth, dieID, portID), logicID, dieID, portID, err)
			}
			if speed, err := hccn.GetNPULinkSpeedA5(logicID, int32(dieID), int32(portID)); err == nil {
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
