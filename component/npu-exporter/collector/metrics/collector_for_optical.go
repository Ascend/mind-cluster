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

const (
	txPower0 = "Tx_Power0"
	txPower1 = "Tx_Power1"
	txPower2 = "Tx_Power2"
	txPower3 = "Tx_Power3"

	rxPower0 = "Rx_Power0"
	rxPower1 = "Rx_Power1"
	rxPower2 = "Rx_Power2"
	rxPower3 = "Rx_Power3"

	notPresent  = "not present"
	present     = "present"
	temperature = "temperature"
	voltage     = "Vcc"

	// Npu specific constants
	txNpuPower0 = "TxPower Lane0(dBm)"
	txNpuPower1 = "TxPower Lane1(dBm)"
	txNpuPower2 = "TxPower Lane2(dBm)"
	txNpuPower3 = "TxPower Lane3(dBm)"

	rxNpuPower0 = "RxPower Lane0(dBm)"
	rxNpuPower1 = "RxPower Lane1(dBm)"
	rxNpuPower2 = "RxPower Lane2(dBm)"
	rxNpuPower3 = "RxPower Lane3(dBm)"

	opticalIndex = "optical_index"
)

var (
	// optical
	descOpticalState    = colcommon.BuildDesc("npu_chip_optical_state", "the npu interface receive optical-state")
	descOpticalVcc      = colcommon.BuildDesc("npu_chip_optical_vcc", "the npu interface receive optical-vcc")
	descOpticalTemp     = colcommon.BuildDesc("npu_chip_optical_temp", "the npu interface receive optical-temperature")
	descOpticalTxPower0 = colcommon.BuildDesc("npu_chip_optical_tx_power_0", "npu interface receive optical-tx-power-0")
	descOpticalTxPower1 = colcommon.BuildDesc("npu_chip_optical_tx_power_1", "npu interface receive optical-tx-power-1")
	descOpticalTxPower2 = colcommon.BuildDesc("npu_chip_optical_tx_power_2", "npu interface receive optical-tx-power-2")
	descOpticalTxPower3 = colcommon.BuildDesc("npu_chip_optical_tx_power_3", "npu interface receive optical-tx-power-3")

	descOpticalRxPower0 = colcommon.BuildDesc("npu_chip_optical_rx_power_0", "npu interface receive optical-rx-power-0")
	descOpticalRxPower1 = colcommon.BuildDesc("npu_chip_optical_rx_power_1", "npu interface receive optical-rx-power-1")
	descOpticalRxPower2 = colcommon.BuildDesc("npu_chip_optical_rx_power_2", "npu interface receive optical-rx-power-2")
	descOpticalRxPower3 = colcommon.BuildDesc("npu_chip_optical_rx_power_3", "npu interface receive optical-rx-power-3")

	// Npu specific metrics
	opticalIndexDesc    []*prometheus.Desc
	opticalTxPower0Desc []*prometheus.Desc
	opticalTxPower1Desc []*prometheus.Desc
	opticalTxPower2Desc []*prometheus.Desc
	opticalTxPower3Desc []*prometheus.Desc
	opticalRxPower0Desc []*prometheus.Desc
	opticalRxPower1Desc []*prometheus.Desc
	opticalRxPower2Desc []*prometheus.Desc
	opticalRxPower3Desc []*prometheus.Desc

	notSupportedOpticalNpuDevices = map[uint32]bool{
		api.Atlas3501PMainBoardID: true,
		api.Atlas3502PMainBoardID: true,
		api.Atlas3504PMainBoardID: true,
		api.Atlas9501DMainBoardID: true,
		api.Atlas950MainBoardID:   true,
	}
)

type opticalCache struct {
	chip      colcommon.HuaWeiAIChip
	timestamp time.Time
	// extInfo indicates the optical module information
	extInfo *common.OpticalInfo
}

type opticalNpuCache struct {
	chip      colcommon.HuaWeiAIChip
	timestamp time.Time
	// extInfo indicates the optical module information
	extInfo []*common.OpticalNpuInfo
}

// OpticalCollector collect the optical metrics
type OpticalCollector struct {
	colcommon.MetricsCollectorAdapter
}

func initNpuOpticalDesc() {
	// Initialize Npu specific metrics descriptions
	for dieID, portIDs := range colcommon.NpuDevPortInfos.GetPortMap() {
		for _, portID := range portIDs {
			colcommon.BuildDescSlice(&opticalIndexDesc, fmt.Sprint(api.MetricsPrefix, "optical_index_num_",
				strconv.Itoa(dieID), "_", strconv.Itoa(portID)), fmt.Sprint("the npu link optical index num ",
				"dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))

			colcommon.BuildDescSlice(&opticalTxPower0Desc, fmt.Sprint(api.MetricsPrefix, "optical_tx_power_0_",
				strconv.Itoa(dieID), "_", strconv.Itoa(portID)), fmt.Sprint("npu interface receive optical_tx_power_0_ ",
				"dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&opticalTxPower1Desc, fmt.Sprint(api.MetricsPrefix, "optical_tx_power_1_",
				strconv.Itoa(dieID), "_", strconv.Itoa(portID)), fmt.Sprint("npu interface receive optical_tx_power_1 ",
				"dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&opticalTxPower2Desc, fmt.Sprint(api.MetricsPrefix, "optical_tx_power_2_",
				strconv.Itoa(dieID), "_", strconv.Itoa(portID)), fmt.Sprint("npu interface receive optical_tx_power_2 ",
				"dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&opticalTxPower3Desc, fmt.Sprint(api.MetricsPrefix, "optical_tx_power_3_",
				strconv.Itoa(dieID), "_", strconv.Itoa(portID)), fmt.Sprint("npu interface receive optical_tx_power_3 ",
				"dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))

			colcommon.BuildDescSlice(&opticalRxPower0Desc, fmt.Sprint(api.MetricsPrefix, "optical_rx_power_0_",
				strconv.Itoa(dieID), "_", strconv.Itoa(portID)), fmt.Sprint("npu interface receive optical_rx_power_0 ",
				"dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&opticalRxPower1Desc, fmt.Sprint(api.MetricsPrefix, "optical_rx_power_1_",
				strconv.Itoa(dieID), "_", strconv.Itoa(portID)), fmt.Sprint("npu interface receive optical_rx_power_1 ",
				"dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&opticalRxPower2Desc, fmt.Sprint(api.MetricsPrefix, "optical_rx_power_2_",
				strconv.Itoa(dieID), "_", strconv.Itoa(portID)), fmt.Sprint("npu interface receive optical_rx_power_2 ",
				"dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
			colcommon.BuildDescSlice(&opticalRxPower3Desc, fmt.Sprint(api.MetricsPrefix, "optical_rx_power_3_",
				strconv.Itoa(dieID), "_", strconv.Itoa(portID)), fmt.Sprint("npu interface receive optical_rx_power_3 ",
				"dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
		}
	}
}

// IsSupported judge whether the collector is supported
func (c *OpticalCollector) IsSupported(n *colcommon.NpuCollector) bool {
	mainBoardID := n.Dmgr.GetMainBoardId()
	// For Npu devices, check if it's a supported optical model
	if colcommon.DevType == api.Ascend910A5 {
		if !notSupportedOpticalNpuDevices[mainBoardID] {
			initNpuOpticalDesc()
			return true
		}
		logForUnSupportDevice(false, colcommon.DevType, colcommon.GetCacheKey(c),
			fmt.Sprint("this mainBoardId:", mainBoardID, " is not supported"))
		return false
	}

	isSupport := n.Dmgr.IsTrainingCard()
	logForUnSupportDevice(isSupport, colcommon.DevType, colcommon.GetCacheKey(c),
		"only training card supports optical related info")
	return isSupport
}

// Describe description of the metric
func (c *OpticalCollector) Describe(ch chan<- *prometheus.Desc) {
	if colcommon.DevType == api.Ascend910A5 {
		// Npu specific optical metrics
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
		return
	}
	// Regular optical metrics
	ch <- descOpticalState
	ch <- descOpticalTxPower0
	ch <- descOpticalTxPower1
	ch <- descOpticalTxPower2
	ch <- descOpticalTxPower3
	ch <- descOpticalRxPower0
	ch <- descOpticalRxPower1
	ch <- descOpticalRxPower2
	ch <- descOpticalRxPower3
	ch <- descOpticalVcc
	ch <- descOpticalTemp
}

// CollectToCache collect the metric to cache
func (c *OpticalCollector) CollectToCache(n *colcommon.NpuCollector, chipList []colcommon.HuaWeiAIChip) {
	if colcommon.DevType == api.Ascend910A5 {
		for _, chip := range chipList {
			// Collect Npu specific optical info
			opticalInfos := collectOpticalNpuInfo(chip.LogicID)
			c.LocalCache.Store(chip.PhyId, opticalNpuCache{chip: chip, timestamp: time.Now(), extInfo: opticalInfos})
		}
		colcommon.UpdateCache[opticalNpuCache](n, colcommon.GetCacheKey(c), &c.LocalCache)
		return
	}
	for _, chip := range chipList {
		// Collect regular optical info
		opticalInfo, err := hccn.GetNPUOpticalInfo(chip.PhyId)
		if err != nil {
			logErrMetricsWithLimit(colcommon.DomainForOptical, chip.PhyId, err)
			continue
		}
		hwlog.ResetErrCnt(colcommon.DomainForOptical, chip.PhyId)
		info := getMainOptInfo(opticalInfo)
		c.LocalCache.Store(chip.PhyId, opticalCache{chip: chip, timestamp: time.Now(), extInfo: info})
	}
	colcommon.UpdateCache[opticalCache](n, colcommon.GetCacheKey(c), &c.LocalCache)
}

// UpdatePrometheus update prometheus metrics
func (c *OpticalCollector) UpdatePrometheus(ch chan<- prometheus.Metric, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) {
	if colcommon.DevType == api.Ascend910A5 {
		// Update Npu specific optical metrics
		updateSingleChipNpu := func(chipWithVnpu colcommon.HuaWeiAIChip, cache opticalNpuCache, cardLabel []string) {
			timestamp := cache.timestamp
			promUpdateOpticalInfo(ch, cache, timestamp, cardLabel)
		}
		updateFrame[opticalNpuCache](colcommon.GetCacheKey(c), n, containerMap, chips, updateSingleChipNpu)
		return
	}
	// Update regular optical metrics
	updateSingleChip := func(chipWithVnpu colcommon.HuaWeiAIChip, cache opticalCache, cardLabel []string) {
		opticalInfo := cache.extInfo
		if opticalInfo == nil {
			return
		}
		timestamp := cache.timestamp
		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo.OpticalState, cardLabel, descOpticalState)
		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo.OpticalVcc, cardLabel, descOpticalVcc)
		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo.OpticalTemp, cardLabel, descOpticalTemp)

		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo.OpticalTxPower0, cardLabel, descOpticalTxPower0)
		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo.OpticalTxPower1, cardLabel, descOpticalTxPower1)
		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo.OpticalTxPower2, cardLabel, descOpticalTxPower2)
		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo.OpticalTxPower3, cardLabel, descOpticalTxPower3)

		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo.OpticalRxPower0, cardLabel, descOpticalRxPower0)
		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo.OpticalRxPower1, cardLabel, descOpticalRxPower1)
		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo.OpticalRxPower2, cardLabel, descOpticalRxPower2)
		doUpdateMetricWithValidateNum(ch, timestamp, opticalInfo.OpticalRxPower3, cardLabel, descOpticalRxPower3)
	}
	updateFrame[opticalCache](colcommon.GetCacheKey(c), n, containerMap, chips, updateSingleChip)
}

// UpdateTelegraf update telegraf metrics
func (c *OpticalCollector) UpdateTelegraf(fieldsMap map[string]map[string]interface{}, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) map[string]map[string]interface{} {
	if colcommon.DevType == api.Ascend910A5 {
		// Update Npu specific optical metrics
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
	// Update regular optical metrics
	caches := colcommon.GetInfoFromCache[opticalCache](n, colcommon.GetCacheKey(c))
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
		doUpdateTelegrafWithValidateNum(fieldMap, descOpticalState, extInfo.OpticalState, "")
		doUpdateTelegrafWithValidateNum(fieldMap, descOpticalVcc, extInfo.OpticalVcc, "")
		doUpdateTelegrafWithValidateNum(fieldMap, descOpticalTemp, extInfo.OpticalTemp, "")

		doUpdateTelegrafWithValidateNum(fieldMap, descOpticalTxPower0, extInfo.OpticalTxPower0, "")
		doUpdateTelegrafWithValidateNum(fieldMap, descOpticalTxPower1, extInfo.OpticalTxPower1, "")
		doUpdateTelegrafWithValidateNum(fieldMap, descOpticalTxPower2, extInfo.OpticalTxPower2, "")
		doUpdateTelegrafWithValidateNum(fieldMap, descOpticalTxPower3, extInfo.OpticalTxPower3, "")

		doUpdateTelegrafWithValidateNum(fieldMap, descOpticalRxPower0, extInfo.OpticalRxPower0, "")
		doUpdateTelegrafWithValidateNum(fieldMap, descOpticalRxPower1, extInfo.OpticalRxPower1, "")
		doUpdateTelegrafWithValidateNum(fieldMap, descOpticalRxPower2, extInfo.OpticalRxPower2, "")
		doUpdateTelegrafWithValidateNum(fieldMap, descOpticalRxPower3, extInfo.OpticalRxPower3, "")
	}
	return fieldsMap
}

func getMainOptInfo(opticalInfo map[string]string) *common.OpticalInfo {
	mainOpticalInfo := common.OpticalInfo{}
	mainOpticalInfo.OpticalTxPower0 = hccn.GetFloatDataFromStr(opticalInfo[txPower0], txPower0)
	mainOpticalInfo.OpticalTxPower1 = hccn.GetFloatDataFromStr(opticalInfo[txPower1], txPower1)
	mainOpticalInfo.OpticalTxPower2 = hccn.GetFloatDataFromStr(opticalInfo[txPower2], txPower2)
	mainOpticalInfo.OpticalTxPower3 = hccn.GetFloatDataFromStr(opticalInfo[txPower3], txPower3)
	mainOpticalInfo.OpticalRxPower0 = hccn.GetFloatDataFromStr(opticalInfo[rxPower0], rxPower0)
	mainOpticalInfo.OpticalRxPower1 = hccn.GetFloatDataFromStr(opticalInfo[rxPower1], rxPower1)
	mainOpticalInfo.OpticalRxPower2 = hccn.GetFloatDataFromStr(opticalInfo[rxPower2], rxPower2)
	mainOpticalInfo.OpticalRxPower3 = hccn.GetFloatDataFromStr(opticalInfo[rxPower3], rxPower3)
	mainOpticalInfo.OpticalVcc = hccn.GetFloatDataFromStr(opticalInfo[voltage], voltage)
	mainOpticalInfo.OpticalTemp = hccn.GetFloatDataFromStr(opticalInfo[temperature], temperature)
	var optState float64
	if opticalInfo[present] == present {
		optState = 1.0
	} else if opticalInfo[present] == notPresent {
		optState = 0.0
	} else {
		optState = common.RetError
	}
	mainOpticalInfo.OpticalState = optState

	return &mainOpticalInfo
}

// Npu specific optical collection functions
func collectOpticalNpuInfo(logicID int32) []*common.OpticalNpuInfo {
	var opticalInfos []*common.OpticalNpuInfo
	for dieID, portIDs := range colcommon.NpuDevPortInfos.GetPortMap() {
		for _, portID := range portIDs {
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

func promUpdateOpticalInfo(ch chan<- prometheus.Metric, cache opticalNpuCache, timestamp time.Time, cardLabel []string) {
	opticalInfo := cache.extInfo
	if opticalInfo == nil {
		return
	}
	for i := 0; i < colcommon.NpuDevPortInfos.GetCount(); i++ {
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

func telegrafUpdateOpticalInfo(cache opticalNpuCache, fieldMap map[string]interface{}) {
	opticalInfo := cache.extInfo
	if opticalInfo == nil {
		return
	}
	for i := 0; i < colcommon.NpuDevPortInfos.GetCount(); i++ {
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
			hwlog.RunLog.Errorf("storeSingleOpticalNpuInfo failed,logicID: %d, udie:%d, port:%d, error is :%v",
				logicID, uDie, port, err)
			return data
		}
		return data
	case "float":
		var data float64
		var err error
		if data, err = hccn.GetFloatDataFromStrNpu(str); err != nil {
			hwlog.RunLog.Errorf("storeSingleOpticalNpuInfo failed,logicID: %d, udie:%d, port:%d, error is :%v",
				logicID, uDie, port, err)
			return data
		}
		return data
	default:
		hwlog.RunLog.Errorf("storeSingleOpticalNpuInfo failed,logicID: %d, udie:%d, port:%d,"+
			" error is : inputType error", logicID, uDie, port)
		return common.RetError
	}
}
