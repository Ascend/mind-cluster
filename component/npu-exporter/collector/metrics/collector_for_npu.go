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
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/utils/logger"
)

var (
	errorCodeDescs        []*prometheus.Desc
	cardLabelForProcess   = append(colcommon.CardLabel, "process_id", "container_id")
	cardLabelForContainer []string
	cardLabelForNpuName   = make([]string, len(colcommon.CardLabel))
)

var (
	machineInfoNPUDesc = colcommon.BuildDescWithLabel("machine_npu_nums", "Amount of npu installed on the machine.", nil)

	descDevProcessNum = colcommon.BuildDesc("npu_chip_info_process_info_num",
		"the npu process num")

	descDevProcessInfo = colcommon.BuildDescWithLabel("npu_chip_info_process_info",
		"the npu process info, unit is 'MB'. if process run on host, container_id and container_name will be empty",
		cardLabelForProcess)

	// container, only report to prometheus
	npuCtrUtilization = colcommon.BuildDesc("container_npu_utilization",
		"npu ai core utilization in container, unit is '%'")
	npuCtrTotalMemory = colcommon.BuildDesc("container_npu_total_memory",
		"npu total memory in container, unit is 'MB'")
	npuCtrUsedMemory = colcommon.BuildDesc("container_npu_used_memory",
		"the npu used memory in container, unit is 'MB'")

	npuCtrInfo *prometheus.Desc = nil
)

func init() {

	colcommon.BuildDescSlice(&errorCodeDescs, "npu_chip_info_error_code", "the npu error code")
	for i := 1; i < common.MaxErrorCodeLen; i++ {
		colcommon.BuildDescSlice(&errorCodeDescs, "npu_chip_info_error_code_"+strconv.Itoa(i), "the npu error code")
	}

	copy(cardLabelForNpuName, colcommon.CardLabel)
	cardLabelForNpuName[1] = "name"
}

type chipCache struct {
	chip      colcommon.HuaWeiAIChip
	timestamp time.Time

	// the healthy status of the  AI chip
	HealthStatus string `json:"health_status"`
	// the all error codes of the chip
	ErrorCodes []int64 `json:"error_codes"`
	// the utilization of the chip
	Utilization int `json:"utilization"`
	// the overall utilization of the chip
	OverallUtilization int `json:"overall_utilization"`
	// the vector utilization of the chip
	VectorUtilization int `json:"vector_utilization"`
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

// CollectToCache collects the base info of the chip
func (c *BaseInfoCollector) CollectToCache(n *colcommon.NpuCollector, chipList []colcommon.HuaWeiAIChip) {
	for _, chip := range chipList {
		logicID := chip.LogicID

		dmgr := n.Dmgr

		cache := &chipCache{
			chip:         chip,
			HealthStatus: getHealth(logicID, dmgr),
		}
		setNetHealthStatus(logicID, dmgr, cache)

		cache.timestamp = time.Now()
		c.LocalCache.Store(chip.PhyId, *cache)
	}
	colcommon.UpdateCache[chipCache](n, colcommon.GetCacheKey(c), &c.LocalCache)
}

func updateContainerInfo(ch chan<- prometheus.Metric, containerInfo container.DevicesInfo, timestamp time.Time,
	cardLabel []string, chip *chipCache) {
	containerName := getContainerNameArray(containerInfo)
	if len(containerName) != colcommon.ContainerNameLen {
		return
	}
	// based on chipType , container_npu_total_memory、container_npu_used_memory reported in hbm or ddr group
	doUpdateMetricWithValidateNum(ch, timestamp, float64(chip.Utilization), cardLabel, npuCtrUtilization)
	doUpdateMetric(ch, timestamp, 1, append(cardLabel, containerInfo.ID, strings.Join(containerName, "_")),
		npuCtrInfo)
}

func updateErrorCodesInfo(ch chan<- prometheus.Metric, chip *chipCache, timestamp time.Time, cardLabel []string) {
	if len(chip.ErrorCodes) > common.MaxErrorCodeLen {
		logger.Logger.Logf(logger.Warn, "Error code number is larger than %v, only the first %v will be reported, "+
			"all errorCode is: %v", common.MaxErrorCodeLen, common.MaxErrorCodeLen, chip.ErrorCodes)
	}
	for i := 0; i < len(chip.ErrorCodes) && i < len(errorCodeDescs); i++ {
		doUpdateMetricWithValidateNum(ch, timestamp, float64(chip.ErrorCodes[i]), cardLabel, errorCodeDescs[i])
	}
}

// UpdateTelegraf updates the base info of the chip
func (c *BaseInfoCollector) UpdateTelegraf(fieldsMap map[int]map[string]interface{}, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) map[int]map[string]interface{} {
	caches := colcommon.GetInfoFromCache[chipCache](n, colcommon.GetCacheKey(c))
	for _, chip := range chips {
		cache, ok := caches[chip.PhyId]
		if !ok {
			continue
		}
		fieldMap := getFieldMap(fieldsMap, cache.chip.LogicID)

		updateErrorCode(&cache, fieldMap)
	}

	return fieldsMap
}

func updateErrorCode(chip *chipCache, fieldMap map[string]interface{}) {
	for i := 0; i < len(chip.ErrorCodes); i++ {
		extInfo := ""
		if i != 0 {
			extInfo = "_" + strconv.Itoa(i)
		}
		doUpdateTelegrafWithValidateNum(fieldMap, errorCodeDescs[i], float64(chip.ErrorCodes[i]), extInfo)
	}
}

func setNetHealthStatus(logicID int32, dmgr devmanager.DeviceInterface, chip *chipCache) {
	chip.NetHealthStatus = colcommon.Abnormal
	if !dmgr.IsTrainingCard() {
		return
	}

	netCode, err := dmgr.GetDeviceNetWorkHealth(logicID)
	logger.Logger.Logf(logger.Debug, "chip %d network healthy code is %d", logicID, netCode)
	if err != nil {
		netCode = math.MaxUint32
	}
	chip.NetHealthStatus = getNetworkHealthy(netCode)
}

func getNetworkHealthy(netCode uint32) string {
	if netCode == math.MaxUint32 {
		return colcommon.Abnormal
	}

	if netCode == common.NetworkInit || netCode == common.NetworkSuccess {
		return colcommon.Healthy
	}

	return colcommon.UnHealthy
}

func getHealth(logicID int32, dmgr devmanager.DeviceInterface) string {
	health, err := dmgr.GetDeviceHealth(logicID)
	if err != nil || health != 0 {
		return colcommon.UnHealthy
	}
	return colcommon.Healthy
}
