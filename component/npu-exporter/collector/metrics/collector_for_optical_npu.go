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
	"ascend-common/common-utils/hwlog"
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"ascend-common/api"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/hccn"

	colcommon "huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
)

var (
	// npu udie port optical index
	opticalIndexDesc []*prometheus.Desc
	// optical tx power info
	opticalTxPower0Desc []*prometheus.Desc
	opticalTxPower1Desc []*prometheus.Desc
	opticalTxPower2Desc []*prometheus.Desc
	opticalTxPower3Desc []*prometheus.Desc
	// optical rx power info
	opticalRxPower0Desc []*prometheus.Desc
	opticalRxPower1Desc []*prometheus.Desc
	opticalRxPower2Desc []*prometheus.Desc
	opticalRxPower3Desc []*prometheus.Desc

	supportedOpticalNpuDevices = map[uint32]bool{
		api.Atlas850MainBoardID:  true,
		api.Atlas850MainBoardID2: true,
	}
)

const (
	txNpuPower0 = "Tx_Power Lane0(dBm)"
	txNpuPower1 = "Tx_Power Lane1(dBm)"
	txNpuPower2 = "Tx_Power Lane2(dBm)"
	txNpuPower3 = "Tx_Power Lane3(dBm)"

	rxNpuPower0 = "Rx_Power Lane0(dBm)"
	rxNpuPower1 = "Rx_Power Lane1(dBm)"
	rxNpuPower2 = "Rx_Power Lane2(dBm)"
	rxNpuPower3 = "Rx_Power Lane3(dBm)"

	opticalIndex = "optical_index"
)

func init() {
	for dieID := 0; dieID < maxDieId; dieID++ {
		for portID := 0; portID < maxPortId; portID++ {
			colcommon.BuildDescSlice(&opticalIndexDesc, fmt.Sprint(api.MetricsPrefix, "optical_index_num_", strconv.Itoa(dieID), "_", strconv.Itoa(portID)),
				fmt.Sprint("the npu link optical index num ", "dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))

			colcommon.BuildDescSlice(&opticalTxPower0Desc, fmt.Sprint(api.MetricsPrefix, "optical_tx_power_0_", strconv.Itoa(dieID), "_", strconv.Itoa(portID)),
				fmt.Sprint("npu interface receive optical_tx_power_0_ ", "dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&opticalTxPower1Desc, fmt.Sprint(api.MetricsPrefix, "optical_tx_power_1_", strconv.Itoa(dieID), "_", strconv.Itoa(portID)),
				fmt.Sprint("npu interface receive optical_tx_power_1 ", "dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&opticalTxPower2Desc, fmt.Sprint(api.MetricsPrefix, "optical_tx_power_2_", strconv.Itoa(dieID), "_", strconv.Itoa(portID)),
				fmt.Sprint("npu interface receive optical_tx_power_2 ", "dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&opticalTxPower3Desc, fmt.Sprint(api.MetricsPrefix, "optical_tx_power_3_", strconv.Itoa(dieID), "_", strconv.Itoa(portID)),
				fmt.Sprint("npu interface receive optical_tx_power_3 ", "dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))

			colcommon.BuildDescSlice(&opticalRxPower0Desc, fmt.Sprint(api.MetricsPrefix, "optical_rx_power_0_", strconv.Itoa(dieID), "_", strconv.Itoa(portID)),
				fmt.Sprint("npu interface receive optical_rx_power_0 ", "dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&opticalRxPower1Desc, fmt.Sprint(api.MetricsPrefix, "optical_rx_power_1_", strconv.Itoa(dieID), "_", strconv.Itoa(portID)),
				fmt.Sprint("npu interface receive optical_rx_power_1 ", "dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&opticalRxPower2Desc, fmt.Sprint(api.MetricsPrefix, "optical_rx_power_2_", strconv.Itoa(dieID), "_", strconv.Itoa(portID)),
				fmt.Sprint("npu interface receive optical_rx_power_2 ", "dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&opticalRxPower3Desc, fmt.Sprint(api.MetricsPrefix, "optical_rx_power_3_", strconv.Itoa(dieID), "_", strconv.Itoa(portID)),
				fmt.Sprint("npu interface receive optical_rx_power_3 ", "dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
		}
	}
}

type opticalNpuCache struct {
	chip      colcommon.HuaWeiAIChip
	timestamp time.Time
	// extInfo indicates the optical module information
	extInfo []*common.OpticalNpuInfo
}

// OpticalNpuCollector collects the optical info
type OpticalNpuCollector struct {
	colcommon.MetricsCollectorAdapter
}

// IsSupported check if the collector is supported
func (c *OpticalNpuCollector) IsSupported(n *colcommon.NpuCollector) bool {
	devType := n.Dmgr.GetDevType()
	if devType != api.Ascend910A5 {
		logForUnSupportDevice(false, devType, colcommon.GetCacheKey(c), "")
		return false
	}
	mainBoardID := n.Dmgr.GetMainBoardId()
	if !supportedOpticalNpuDevices[mainBoardID] {
		logForUnSupportDevice(false, devTypeMap[devType], colcommon.GetCacheKey(c),
			fmt.Sprint("this mainBoardId:", mainBoardID, " is not supported"))
		return false
	}
	return true
}

// Describe description of the metric
func (c *OpticalNpuCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range opticalIndexDesc {
		ch <- desc
	}
	for _, desc := range opticalTxPower0Desc {
		ch <- desc
	}
	for _, desc := range opticalTxPower1Desc {
		ch <- desc
	}
	for _, desc := range opticalTxPower2Desc {
		ch <- desc
	}
	for _, desc := range opticalTxPower3Desc {
		ch <- desc
	}
	for _, desc := range opticalRxPower0Desc {
		ch <- desc
	}
	for _, desc := range opticalRxPower1Desc {
		ch <- desc
	}
	for _, desc := range opticalRxPower2Desc {
		ch <- desc
	}
	for _, desc := range opticalRxPower3Desc {
		ch <- desc
	}
}

// CollectToCache collect the metric to cache
func (c *OpticalNpuCollector) CollectToCache(n *colcommon.NpuCollector, chipList []colcommon.HuaWeiAIChip) {
	for _, chip := range chipList {
		var opticalInfos []*common.OpticalNpuInfo
		opticalInfos = collectOpticalNpuInfo(chip.LogicID)
		c.LocalCache.Store(chip.PhyId, opticalNpuCache{chip: chip, timestamp: time.Now(), extInfo: opticalInfos})
	}
	colcommon.UpdateCache[opticalNpuCache](n, colcommon.GetCacheKey(c), &c.LocalCache)
}

// UpdatePrometheus update prometheus metrics
func (c *OpticalNpuCollector) UpdatePrometheus(ch chan<- prometheus.Metric, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) {

	updateSingleChip := func(chipWithVnpu colcommon.HuaWeiAIChip, cache opticalNpuCache, cardLabel []string) {
		timestamp := cache.timestamp
		promUpdateOpticalInfo(ch, cache, timestamp, cardLabel)
	}
	updateFrame[opticalNpuCache](colcommon.GetCacheKey(c), n, containerMap, chips, updateSingleChip)
}

func promUpdateOpticalInfo(ch chan<- prometheus.Metric, cache opticalNpuCache,
	timestamp time.Time, cardLabel []string) {
	opticalInfo := cache.extInfo
	if opticalInfo == nil {
		return
	}
	for i := 0; i < (maxDieId * maxPortId); i++ {
		if opticalInfo[i] == nil {
			continue
		}
		doUpdateMetric(ch, timestamp, opticalInfo[i].OpticalIndex, cardLabel, opticalIndexDesc[i])

		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo[i].OpticalTxPower0, cardLabel, opticalTxPower0Desc[i])
		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo[i].OpticalTxPower1, cardLabel, opticalTxPower1Desc[i])
		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo[i].OpticalTxPower2, cardLabel, opticalTxPower2Desc[i])
		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo[i].OpticalTxPower3, cardLabel, opticalTxPower3Desc[i])

		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo[i].OpticalRxPower0, cardLabel, opticalRxPower0Desc[i])
		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo[i].OpticalRxPower1, cardLabel, opticalRxPower1Desc[i])
		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo[i].OpticalRxPower2, cardLabel, opticalRxPower2Desc[i])
		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo[i].OpticalRxPower3, cardLabel, opticalRxPower3Desc[i])
	}
}

// UpdateTelegraf update telegraf metrics
func (c *OpticalNpuCollector) UpdateTelegraf(fieldsMap map[string]map[string]interface{}, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) map[string]map[string]interface{} {
	caches := colcommon.GetInfoFromCache[opticalNpuCache](n, colcommon.GetCacheKey(c))
	for _, chip := range chips {
		cache, ok := caches[chip.PhyId]
		if !ok {
			continue
		}
		fieldMap := getFieldMap(fieldsMap, cache.chip.LogicID)

		telegrafUpdateOpticalInfo(cache, fieldMap)
	}

	return fieldsMap
}

func telegrafUpdateOpticalInfo(cache opticalNpuCache, fieldMap map[string]interface{}) {
	opticalInfo := cache.extInfo
	if opticalInfo == nil {
		return
	}
	for i := 0; i < (maxDieId * maxPortId); i++ {
		if opticalInfo[i] == nil {
			continue
		}
		doUpdateTelegraf(fieldMap, opticalIndexDesc[i], opticalInfo[i].OpticalIndex, "")

		doUpdateTelegrafWithValidateNum(fieldMap, opticalTxPower0Desc[i], opticalInfo[i].OpticalTxPower0, "")
		doUpdateTelegrafWithValidateNum(fieldMap, opticalTxPower1Desc[i], opticalInfo[i].OpticalTxPower1, "")
		doUpdateTelegrafWithValidateNum(fieldMap, opticalTxPower2Desc[i], opticalInfo[i].OpticalTxPower2, "")
		doUpdateTelegrafWithValidateNum(fieldMap, opticalTxPower3Desc[i], opticalInfo[i].OpticalTxPower3, "")

		doUpdateTelegrafWithValidateNum(fieldMap, opticalRxPower0Desc[i], opticalInfo[i].OpticalRxPower0, "")
		doUpdateTelegrafWithValidateNum(fieldMap, opticalRxPower1Desc[i], opticalInfo[i].OpticalRxPower1, "")
		doUpdateTelegrafWithValidateNum(fieldMap, opticalRxPower2Desc[i], opticalInfo[i].OpticalRxPower2, "")
		doUpdateTelegrafWithValidateNum(fieldMap, opticalRxPower3Desc[i], opticalInfo[i].OpticalRxPower3, "")
	}
}

func collectOpticalNpuInfo(logicID int32) []*common.OpticalNpuInfo {
	var opticalInfos []*common.OpticalNpuInfo
	for dieID := 0; dieID < maxDieId; dieID++ {
		for portID := 0; portID < maxPortId; portID++ {
			opticalInfo := &common.OpticalNpuInfo{}
			if info, err := hccn.GetNpuOpticalInfoNpu(logicID, int32(dieID), int32(portID)); err == nil {
				opticalInfo = storeOpticalNpuInfos(info, logicID, dieID, portID)
				hwlog.ResetErrCnt(fmt.Sprint(colcommon.DomainForOpticalV2, dieID, portID), logicID)
			} else {
				opticalInfo = nil
				logWarnMetricsWithLimit(fmt.Sprint(colcommon.DomainForOpticalV2, dieID, portID), logicID, dieID, portID, err)
			}
			opticalInfos = append(opticalInfos, opticalInfo)
		}
	}
	return opticalInfos
}

func storeOpticalNpuInfos(info map[string]string, logicID int32, dieID, portID int) *common.OpticalNpuInfo {
	opticalInfo := common.OpticalNpuInfo{}
	if val, ok := storeSingleOpticalNpuInfo(info[txNpuPower0], logicID, dieID, portID, "float").(float64); ok {
		opticalInfo.OpticalTxPower0 = val
	}
	if val, ok := storeSingleOpticalNpuInfo(info[txNpuPower1], logicID, dieID, portID, "float").(float64); ok {
		opticalInfo.OpticalTxPower1 = val
	}
	if val, ok := storeSingleOpticalNpuInfo(info[txNpuPower2], logicID, dieID, portID, "float").(float64); ok {
		opticalInfo.OpticalTxPower2 = val
	}
	if val, ok := storeSingleOpticalNpuInfo(info[txNpuPower3], logicID, dieID, portID, "float").(float64); ok {
		opticalInfo.OpticalTxPower3 = val
	}
	if val, ok := storeSingleOpticalNpuInfo(info[rxNpuPower0], logicID, dieID, portID, "float").(float64); ok {
		opticalInfo.OpticalRxPower0 = val
	}
	if val, ok := storeSingleOpticalNpuInfo(info[rxNpuPower1], logicID, dieID, portID, "float").(float64); ok {
		opticalInfo.OpticalRxPower1 = val
	}
	if val, ok := storeSingleOpticalNpuInfo(info[rxNpuPower2], logicID, dieID, portID, "float").(float64); ok {
		opticalInfo.OpticalRxPower2 = val
	}
	if val, ok := storeSingleOpticalNpuInfo(info[rxNpuPower3], logicID, dieID, portID, "float").(float64); ok {
		opticalInfo.OpticalRxPower3 = val
	}
	if val, ok := storeSingleOpticalNpuInfo(info[opticalIndex], logicID, dieID, portID, "int").(int); ok {
		opticalInfo.OpticalIndex = val
	}
	return &opticalInfo
}

func storeSingleOpticalNpuInfo(str string, logicID int32, uDie, port int, convertType string) interface{} {
	switch convertType {
	case "int":
		var data int
		var err error
		if data, err = hccn.GetIntDataFromStrNpu(str); err != nil {
			hwlog.RunLog.Errorf("storeSingleOpticalNpuInfo failed,logicID: %d, udie:%d, port:%d, error is :%v", logicID, uDie, port, err)
			return data
		}
		return data
	case "float":
		var data float64
		var err error
		if data, err = hccn.GetFloatDataFromStrNpu(str); err != nil {
			hwlog.RunLog.Errorf("storeSingleOpticalNpuInfo failed,logicID: %d, udie:%d, port:%d, error is :%v", logicID, uDie, port, err)
			return data
		}
		return data
	default:
		hwlog.RunLog.Errorf("storeSingleOpticalNpuInfo failed,logicID: %d, udie:%d, port:%d, error is : inputType error", logicID, uDie, port)
		return common.RetError
	}
}
