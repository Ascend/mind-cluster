/* Copyright(C) 2025-2026. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
	utils2 "huawei.com/npu-exporter/v6/utils"
	"huawei.com/npu-exporter/v6/utils/logger"
)

var (
	errorCodeDescs        []*prometheus.Desc
	cardLabelForProcess   = append(colcommon.CardLabel, "process_id", "container_id")
	cardLabelForSN        = append(colcommon.CardLabel, "serial_number")
	cardLabelForProduct   = append(colcommon.CardLabel, "product_type")
	cardLabelForNpuName   = make([]string, len(colcommon.CardLabel))
	cardLabelForContainer []string

	notSupportedNetworkHealthDevices = map[uint32]bool{
		api.Atlas3501PMainBoardID: true,
		api.Atlas3502PMainBoardID: true,
		api.Atlas3504PMainBoardID: true,
	}

	supportedCardNumDevices = map[string]bool{
		api.Ascend310P:  true,
		api.Ascend910A:  true,
		api.Ascend910B:  true,
		api.Ascend910A3: true,
	}
)

var (
	machineInfoNPUDesc  = colcommon.BuildDescWithLabel("machine_npu_nums", "Amount of npu installed on the machine.", nil)
	machineInfoCardDesc = prometheus.NewDesc("machine_card_nums", "Amount of card installed on the machine.", nil, nil)

	descTemp    = colcommon.BuildDesc("npu_chip_info_temperature", "the npu temperature")
	descPower   = colcommon.BuildDesc("npu_chip_info_power", "the npu power")
	descVoltage = colcommon.BuildDesc("npu_chip_info_voltage", "the npu voltage")

	descAICoreFreq = colcommon.BuildDesc("npu_chip_info_aicore_current_freq",
		"the npu ai core current frequency, unit is 'MHz'")
	descHealthStatus  = colcommon.BuildDesc("npu_chip_info_health_status", "the npu health status")
	descDevProcessNum = colcommon.BuildDesc("npu_chip_info_process_info_num",
		"the npu process num")

	descDevProcessInfo = colcommon.BuildDescWithLabel("npu_chip_info_process_info",
		"the npu process info, unit is 'MB'. if process run on host, container_id and container_name will be empty",
		cardLabelForProcess)

	// net status
	descNetworkStatus = colcommon.BuildDesc("npu_chip_info_network_status", "the npu network health status")
	// NPU serial number
	descNPUSerialNumber = colcommon.BuildDescWithLabel("npu_chip_info_serial_number",
		"the npu serial number information", cardLabelForSN)
	// product type
	descNPUProduct = colcommon.BuildDescWithLabel("npu_chip_info_product_type", "the npu product_type information",
		cardLabelForProduct)

	npuCtrTotalMemory = colcommon.BuildDesc("container_npu_total_memory",
		"npu total memory in container, unit is 'MB'")
	npuCtrUsedMemory = colcommon.BuildDesc("container_npu_used_memory",
		"the npu used memory in container, unit is 'MB'")

	npuCtrInfo  *prometheus.Desc = nil
	descNpuName *prometheus.Desc = nil
)

func init() {

	colcommon.BuildDescSlice(&errorCodeDescs, "npu_chip_info_error_code", "the npu error code")
	for i := 1; i < common.MaxErrorCodeLen; i++ {
		colcommon.BuildDescSlice(&errorCodeDescs, "npu_chip_info_error_code_"+strconv.Itoa(i), "the npu error code")
	}

	cardLabelForContainer = append(colcommon.CardLabel, "containerID", "containerName")
	cardLabelForContainer[0] = "npuID"
	npuCtrInfo = colcommon.BuildDescWithLabel("npu_container_info", "the container name and deviceID relationship",
		cardLabelForContainer)

	copy(cardLabelForNpuName, colcommon.CardLabel)
	cardLabelForNpuName[1] = "name"
	descNpuName = colcommon.BuildDescWithLabel("npu_chip_info_name", "the Ascend npu name with value '1'",
		cardLabelForNpuName)
}

type chipCache struct {
	chip      colcommon.HuaWeiAIChip
	timestamp time.Time

	// the healthy status of the  AI chip
	HealthStatus string `json:"health_status"`
	// the all error codes of the chip
	ErrorCodes []int64 `json:"error_codes"`
	// the temperature of the chip
	Temperature int `json:"temperature"`
	// the work power of the chip
	Power float32 `json:"power"`
	// the work voltage of the chip
	Voltage float32 `json:"voltage"`
	// the AI core current frequency of the chip
	AICoreCurrentFreq uint32 `json:"aicore_current_freq"`
	// NetHealthStatus chip network health status
	NetHealthStatus string `json:"net_health_status"`
	// DevProcessInfo chip process info
	DevProcessInfo *common.DevProcessInfo
}

// BaseInfoCollector collects the base info of the chip
type BaseInfoCollector struct {
	colcommon.MetricsCollectorAdapter
}

// Describe collects the base info of the chip
func (c *BaseInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	// base info
	ch <- machineInfoNPUDesc
	ch <- machineInfoCardDesc
	ch <- descTemp
	ch <- descPower
	ch <- descVoltage
	ch <- descHealthStatus
	ch <- descNpuName
	ch <- descAICoreFreq
	ch <- descNPUSerialNumber
	ch <- descDevProcessInfo
	ch <- descNPUProduct
	// status
	ch <- descNetworkStatus
	// container
	ch <- npuCtrInfo
	ch <- npuCtrTotalMemory
	ch <- npuCtrUsedMemory

	// error code
	for _, desc := range errorCodeDescs {
		ch <- desc
	}
}

// CollectToCache collects the base info of the chip
func (c *BaseInfoCollector) CollectToCache(n *colcommon.NpuCollector, chipList []colcommon.HuaWeiAIChip) {
	collectCardNum(n, c)
	for _, chip := range chipList {
		logicID := chip.LogicID

		dmgr := n.Dmgr

		freq, err := dmgr.GetDeviceFrequency(logicID, common.AICoreCurrentFreq)
		if err != nil {
			freq = common.UnRetError
		}
		temp, err := dmgr.GetDeviceTemperature(logicID)
		if err != nil {
			temp = common.RetError
		}
		vol, err := dmgr.GetDeviceVoltage(logicID)
		if err != nil {
			vol = common.UnRetError
		}

		_, errCodes, err := dmgr.GetDeviceAllErrorCode(logicID)
		if err != nil {
			errCodes = make([]int64, 0)
		}
		if len(errCodes) > 0 {
			var errCodesInHex []string
			for _, errCode := range errCodes {
				errCodesInHex = append(errCodesInHex, strings.ToUpper(strconv.FormatInt(errCode, 16)))
			}
			logger.Warnf("there are error(s) on chip(logicID: %v), errorCodes: %v", logicID, errCodesInHex)
		}

		cache := &chipCache{
			chip:              chip,
			AICoreCurrentFreq: freq,
			Temperature:       int(temp),
			Voltage:           vol,
			HealthStatus:      getHealth(logicID, dmgr),
			ErrorCodes:        errCodes,
		}
		collectPower(logicID, dmgr, cache)
		if isSupportNetworkHealthDevices(n.Dmgr.GetDevType(), chip.MainBoardId) {
			setNetHealthStatus(logicID, dmgr, cache)
		}
		setProcessInfo(logicID, dmgr, cache)

		cache.timestamp = time.Now()
		c.LocalCache.Store(chip.PhyId, *cache)
	}
	colcommon.UpdateCache[chipCache](n, colcommon.GetCacheKey(c), &c.LocalCache)
}

func collectCardNum(n *colcommon.NpuCollector, c *BaseInfoCollector) {
	if !supportedCardNumDevices[n.Dmgr.GetDevType()] {
		logger.LogfWithOptions(logger.ErrorLevel, logger.LogOptions{Domain: colcommon.DomainForCardNum, ID: 0, MaxCounts: 1},
			fmt.Sprintf("devType %v does not support [%v]",
				utils.MaskDevType(n.Dmgr.GetDevType()), utils2.GetDescName(machineInfoCardDesc)))
		return
	}

	cardNum, _, err := n.Dmgr.GetCardList()
	if err != nil {
		logger.LogfWithOptions(logger.ErrorLevel, logger.LogOptions{Domain: colcommon.DomainForCardNum, ID: 0}, err.Error())
		return
	}
	hwlog.ResetErrCnt(colcommon.DomainForCardNum, 0)
	c.LocalCache.Store(colcommon.MachineInfoCardDescKey, cardNum)
}

func collectPower(logicID int32, dmgr devmanager.DeviceInterface, chip *chipCache) {
	if dmgr.GetDevType() == api.Ascend310P {
		cardPower, err := dmgr.GetMcuPowerInfo(chip.chip.CardId)
		handleErr(err, colcommon.DomainForMcuPower, chip.chip.CardId)
		// Ascend310P use cardPower to replace chipPower
		chip.Power = cardPower
	} else {
		power, err := dmgr.GetDevicePowerInfo(logicID)
		handleErr(err, colcommon.DomainForChipPower, logicID)
		chip.Power = power
	}
}

// UpdatePrometheus updates the base info of the chip
func (c *BaseInfoCollector) UpdatePrometheus(ch chan<- prometheus.Metric, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) {

	updateSingleChip := func(chipWithVnpu colcommon.HuaWeiAIChip, cache chipCache, cardLabel []string) {
		containerInfo := geenContainerInfo(&chipWithVnpu, containerMap)
		timestamp := cache.timestamp
		doUpdateMetricWithValidateNum(ch, timestamp, float64(cache.Power), cardLabel, descPower)
		doUpdateMetricWithValidateNum(ch, timestamp, float64(cache.Voltage), cardLabel, descVoltage)
		doUpdateMetricWithValidateNum(ch, timestamp, float64(cache.AICoreCurrentFreq), cardLabel, descAICoreFreq)
		doUpdateMetricWithValidateNum(ch, timestamp, float64(cache.Temperature), cardLabel, descTemp)
		doUpdateMetricWithValidateNum(ch, timestamp, 1, cardLabel, descNpuName)
		doUpdateMetricWithValidateNum(ch, timestamp, float64(getHealthCode(cache.HealthStatus)), cardLabel, descHealthStatus)
		if isSupportNetworkHealthDevices(n.Dmgr.GetDevType(), cache.chip.MainBoardId) {
			doUpdateMetricWithValidateNum(ch, timestamp, float64(getHealthCode(cache.NetHealthStatus)),
				cardLabel, descNetworkStatus)
		}

		updateContainerInfo(ch, containerInfo, cardLabel, &cache, chipWithVnpu)

		updateProcessInfoForPrometheus(ch, &cache, containerInfo, timestamp, cardLabel)
		updateErrorCodesInfo(ch, &cache, timestamp, cardLabel)
		// Update NPU serial number info
		if cache.chip.ElabelInfo != nil {
			snLabel := append(cardLabel, cache.chip.ElabelInfo.SerialNumber)
			doUpdateMetricWithValidateNum(ch, timestamp, 1, snLabel, descNPUSerialNumber)
		}
		if cache.chip.ProductType != "" {
			doUpdateMetricWithValidateNum(ch, timestamp, 1, append(cardLabel, cache.chip.ProductType), descNPUProduct)
		}
	}
	updateFrame[chipCache](colcommon.GetCacheKey(c), n, containerMap, chips, updateSingleChip)
	updateMachineInfoCardMetric(ch, &c.LocalCache)
	ch <- prometheus.MustNewConstMetric(machineInfoNPUDesc, prometheus.GaugeValue, float64(len(chips)))
}

func updateMachineInfoCardMetric(ch chan<- prometheus.Metric, localCache *sync.Map) {
	machineInfoCardCache, ok := localCache.Load(colcommon.MachineInfoCardDescKey)
	if !ok {
		return
	}
	cardNum, ok := machineInfoCardCache.(int32)
	if !ok {
		logger.Errorf("machineInfoCardCache type assertion failed, got %T", machineInfoCardCache)
		return
	}
	ch <- prometheus.MustNewConstMetric(machineInfoCardDesc, prometheus.GaugeValue, float64(cardNum))
}

func updateContainerInfo(ch chan<- prometheus.Metric, containerInfo container.DevicesInfo,
	cardLabel []string, chip *chipCache, chipWithVnpu colcommon.HuaWeiAIChip) {
	containerName := getContainerNameArray(containerInfo)
	if len(containerName) != colcommon.ContainerNameLen {
		return
	}
	// based on chipType , container_npu_total_memory、container_npu_used_memory reported in hbm or ddr group
	doUpdateMetric(ch, chip.timestamp, 1, append(cardLabel, containerInfo.ID, strings.Join(containerName, "_")),
		npuCtrInfo)
}

func updateErrorCodesInfo(ch chan<- prometheus.Metric, chip *chipCache, timestamp time.Time, cardLabel []string) {
	if len(chip.ErrorCodes) > common.MaxErrorCodeLen {
		logger.Warnf("Error code number is larger than %v, only the first %v will be reported, "+
			"all errorCode is: %v", common.MaxErrorCodeLen, common.MaxErrorCodeLen, chip.ErrorCodes)
	}
	for i := 0; i < len(chip.ErrorCodes) && i < len(errorCodeDescs); i++ {
		doUpdateMetricWithValidateNum(ch, timestamp, float64(chip.ErrorCodes[i]), cardLabel, errorCodeDescs[i])
	}
}

func updateProcessInfoForPrometheus(ch chan<- prometheus.Metric, chip *chipCache,
	containerInfo container.DevicesInfo, timestamp time.Time, cardLabel []string) {
	devProcessInfo := chip.DevProcessInfo
	if devProcessInfo == nil {
		return
	}
	doUpdateMetric(ch, timestamp, devProcessInfo.ProcNum, cardLabel, descDevProcessNum)

	containerID := ""
	containerName := ""
	cNameArray := getContainerNameArray(containerInfo)
	if len(cNameArray) == colcommon.ContainerNameLen {
		containerID = containerInfo.ID
		containerName = strings.Join(cNameArray, "_")
	}

	newCardLabel := make([]string, len(cardLabel))
	copy(newCardLabel, cardLabel)
	// containerName in process info is namespace_podName_containerName
	newCardLabel[len(newCardLabel)-1] = containerName

	if devProcessInfo.ProcNum == 0 {
		doUpdateMetric(ch, timestamp, 0, append(newCardLabel, "", containerID), descDevProcessInfo)
		return
	}

	for i := int32(0); i < devProcessInfo.ProcNum; i++ {
		procInfo := devProcessInfo.DevProcArray[i]
		doUpdateMetric(ch, timestamp, procInfo.MemUsage,
			append(newCardLabel, strconv.FormatInt(int64(procInfo.Pid), colcommon.Base), containerID), descDevProcessInfo)
	}
}

// UpdateTelegraf updates the base info of the chip
func (c *BaseInfoCollector) UpdateTelegraf(ch chan<- colcommon.TelegrafMetric, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) {
	caches := colcommon.GetInfoFromCache[chipCache](n, colcommon.GetCacheKey(c))
	for _, chip := range chips {
		cache, ok := caches[chip.PhyId]
		if !ok {
			continue
		}
		metric := colcommon.NewDeviceMetric(cache.chip.LogicID)

		doUpdateTelegrafWithValidateNum(metric.Fields, descTemp, float64(cache.Temperature), "")
		doUpdateTelegrafWithValidateNum(metric.Fields, descPower, float64(cache.Power), "")
		doUpdateTelegrafWithValidateNum(metric.Fields, descVoltage, float64(cache.Voltage), "")
		doUpdateTelegrafWithValidateNum(metric.Fields, descAICoreFreq, float64(cache.AICoreCurrentFreq), "")
		doUpdateTelegrafWithValidateNum(metric.Fields, descHealthStatus, float64(getHealthCode(cache.HealthStatus)), "")
		if isSupportNetworkHealthDevices(n.Dmgr.GetDevType(), chip.MainBoardId) {
			doUpdateTelegrafWithValidateNum(metric.Fields, descNetworkStatus, float64(getHealthCode(cache.NetHealthStatus)), "")
		}
		doUpdateTelegraf(metric.Fields, descNpuName, chip.ChipInfo.Name, "")

		updateProcessInfoForTelegraf(&cache, metric.Fields)
		updateErrorCode(&cache, metric.Fields)
		// Update NPU serial number info
		if cache.chip.ElabelInfo != nil {
			doUpdateTelegraf(metric.Fields, descNPUSerialNumber, cache.chip.ElabelInfo.SerialNumber, "")
		}
		if cache.chip.ProductType != "" {
			doUpdateTelegraf(metric.Fields, descNPUProduct, cache.chip.ProductType, "")
		}
		ch <- metric
	}

	metric := colcommon.NewGeneralMetric()
	machineInfoCardCache, ok := c.LocalCache.Load(colcommon.MachineInfoCardDescKey)
	if ok {
		doUpdateTelegraf(metric.Fields, machineInfoCardDesc, machineInfoCardCache, "")
	}
	doUpdateTelegraf(metric.Fields, machineInfoNPUDesc, len(chips), "")
	ch <- metric
}

func updateErrorCode(chip *chipCache, fieldMap map[string]interface{}) {
	if len(errorCodeDescs) == 0 {
		return
	}
	descErrorCode := errorCodeDescs[0]
	for i := 0; i < len(chip.ErrorCodes); i++ {
		extInfo := ""
		if i != 0 {
			extInfo = "_" + strconv.Itoa(i)
		}
		doUpdateTelegrafWithValidateNum(fieldMap, descErrorCode, float64(chip.ErrorCodes[i]), extInfo)
	}
}

func updateProcessInfoForTelegraf(chip *chipCache, fieldMap map[string]interface{}) {
	devProcessInfo := chip.DevProcessInfo
	doUpdateTelegraf(fieldMap, descDevProcessNum, devProcessInfo.ProcNum, "")
	if devProcessInfo.ProcNum == 0 {
		doUpdateTelegraf(fieldMap, descDevProcessInfo, 0, "")
		return
	}
	for i := int32(0); i < devProcessInfo.ProcNum; i++ {
		procInfo := devProcessInfo.DevProcArray[i]
		doUpdateTelegraf(fieldMap, descDevProcessInfo, procInfo.MemUsage, "_"+strconv.Itoa(int(procInfo.Pid)))
	}
}

func setNetHealthStatus(logicID int32, dmgr devmanager.DeviceInterface, chip *chipCache) {
	chip.NetHealthStatus = colcommon.NotReport
	if !dmgr.IsTrainingCard() {
		return
	}
	chip.NetHealthStatus = getNetworkHealthy(logicID, dmgr)
}

func getNetworkHealthy(logicID int32, dmgr devmanager.DeviceInterface) string {
	netCode, err := dmgr.GetDeviceNetWorkHealth(logicID)
	logger.Debugf("chip %d network healthy code is %d", logicID, netCode)
	if err != nil {
		return colcommon.Unknown
	}
	if netCode == common.NetworkInit || netCode == common.NetworkSuccess {
		return colcommon.Healthy
	}

	return colcommon.UnHealthy
}

func getHealth(logicID int32, dmgr devmanager.DeviceInterface) string {
	health, err := dmgr.GetDeviceHealth(logicID)
	if err != nil {
		return colcommon.Unknown
	}
	if health != 0 {
		return colcommon.UnHealthy
	}
	return colcommon.Healthy
}

func getHealthCode(health string) int {
	if health == colcommon.NotReport {
		return common.UnRetError
	}
	if health == colcommon.Unknown {
		return common.FailedValue
	}
	if colcommon.Healthy == health {
		return colcommon.HealthyCode
	}
	return colcommon.UnhealthyCode
}

func setProcessInfo(logicID int32, dmgr devmanager.DeviceInterface, hwChip *chipCache) {
	productTypes := dmgr.GetProductTypeArray()
	info, err := dmgr.GetDevProcessInfo(logicID)
	if err != nil {
		if len(productTypes) == 1 && productTypes[0] == common.Atlas200ISoc {
			logger.Debugf("process info is not supported on %s", common.Atlas200ISoc)
			hwChip.DevProcessInfo = &common.DevProcessInfo{}
			return
		}
		handleErr(err, colcommon.DomainForProcess, logicID)
		info = &common.DevProcessInfo{}
	}
	hwChip.DevProcessInfo = info
}

func isSupportNetworkHealthDevices(devType string, mainBoardId uint32) bool {
	if devType == api.Ascend910A5 && notSupportedNetworkHealthDevices[mainBoardId] {
		return false
	}
	return true
}
