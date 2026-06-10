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
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"ascend-common/common-utils/utils"
	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/utils/logger"
)

var (
	notSupportedVectorUtilDevices = map[string]bool{
		common.Ascend910: true,
	}
	supportedOverallUtilDevices = map[string]bool{
		common.Ascend910B:  true,
		common.Ascend910A3: true,
		common.Ascend910A5: true,
	}
	supportedCubeDevices = map[string]bool{
		common.Ascend910B:  true,
		common.Ascend910A3: true,
	}
)

var (
	descUtil       = colcommon.BuildDesc("npu_chip_info_utilization", "the ai core utilization")
	descOverUtil   = colcommon.BuildDesc("npu_chip_info_overall_utilization", "the overall utilization of npu")
	descVectorUtil = colcommon.BuildDesc("npu_chip_info_vector_utilization", "the vector utilization")
	descCubeUtil   = colcommon.BuildDesc("npu_chip_info_cube_utilization", "the cube utilization")

	// container (vnpu not support this metrics), only report to prometheus
	npuCtrUtilization = colcommon.BuildDesc("container_npu_utilization",
		"npu ai core utilization in container, unit is '%'")
)

type chipUtilizationCache struct {
	chip colcommon.HuaWeiAIChip
	// Utilization the ai core utilization of the chip
	Utilization int `json:"utilization"`
	// OverallUtilization the overall utilization of the chip
	OverallUtilization int `json:"overall_utilization"`
	// VectorUtilization the vector utilization of the chip
	VectorUtilization int `json:"vector_utilization"`
	// CubeUtilization the cube utilization of the chip
	CubeUtilization int `json:"cube_utilization"`
	timestamp       time.Time
}

// UtilizationCollector collects the base info of the chip
type UtilizationCollector struct {
	colcommon.MetricsCollectorAdapter
	realGetDeviceUtilizationRateInfoFunc func(logicID int32, dmgr devmanager.DeviceInterface, chip *chipUtilizationCache)
}

func (c *UtilizationCollector) PreCollect(n *colcommon.NpuCollector, chipList []colcommon.HuaWeiAIChip) {
	if n.Dmgr.GetDevType() != common.Ascend910B &&
		n.Dmgr.GetDevType() != common.Ascend910A3 &&
		n.Dmgr.GetDevType() != common.Ascend910A5 {
		// only A2、A3 and A5 support use new api (dcmi_get_device_multi_utilization_rate、dcmi_get_device_multi_utilization_rate)
		c.realGetDeviceUtilizationRateInfoFunc = collectUtilV1
		logger.Infof("devType %v does not support get device utilization by v2 api, "+
			"will use v1 api to get utilization info", utils.MaskDevType(n.Dmgr.GetDevType()))
		return
	}
	c.realGetDeviceUtilizationRateInfoFunc = collectUtilCommon
}

// Describe collects the base info of the chip
func (c *UtilizationCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- descUtil
	ch <- descVectorUtil
	ch <- descCubeUtil
	ch <- descOverUtil
	ch <- npuCtrUtilization
}

// CollectToCache collects the base info of the chip
func (c *UtilizationCollector) CollectToCache(n *colcommon.NpuCollector, chipList []colcommon.HuaWeiAIChip) {
	for _, chip := range chipList {
		logicID := chip.LogicID
		cache := &chipUtilizationCache{chip: chip}
		collectUtil(c, logicID, n.Dmgr, cache)

		cache.timestamp = time.Now()
		c.LocalCache.Store(chip.PhyId, *cache)
	}
	colcommon.UpdateCache[chipUtilizationCache](n, colcommon.GetCacheKey(c), &c.LocalCache)
}

// UpdatePrometheus updates the base info of the chip
func (c *UtilizationCollector) UpdatePrometheus(ch chan<- prometheus.Metric, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) {

	updateSingleChip := func(chipWithVnpu colcommon.HuaWeiAIChip, cache chipUtilizationCache, cardLabel []string) {
		containerInfo := geenContainerInfo(&chipWithVnpu, containerMap)
		timestamp := cache.timestamp
		doUpdateMetricWithValidateNum(ch, timestamp, float64(cache.Utilization), cardLabel, descUtil)
		doUpdateMetricWithValidateNum(ch, timestamp, float64(cache.OverallUtilization), cardLabel, descOverUtil)
		doUpdateMetricWithValidateNum(ch, timestamp, float64(cache.VectorUtilization), cardLabel, descVectorUtil)
		doUpdateMetricWithValidateNum(ch, timestamp, float64(cache.CubeUtilization), cardLabel, descCubeUtil)

		updateContainerUtilization(ch, containerInfo, cardLabel, &cache, chipWithVnpu)

	}
	updateFrame[chipUtilizationCache](colcommon.GetCacheKey(c), n, containerMap, chips, updateSingleChip)
}

func updateContainerUtilization(ch chan<- prometheus.Metric, containerInfo container.DevicesInfo,
	cardLabel []string, chip *chipUtilizationCache, chipWithVnpu colcommon.HuaWeiAIChip) {
	containerName := getContainerNameArray(containerInfo)
	if len(containerName) != colcommon.ContainerNameLen {
		return
	}

	// vnpu not support this metrics
	vDevActivityInfo := chipWithVnpu.VDevActivityInfo
	if vDevActivityInfo != nil && common.IsValidVDevID(vDevActivityInfo.VDevID) {
		return
	}

	doUpdateMetricWithValidateNum(ch, chip.timestamp, float64(chip.Utilization), cardLabel, npuCtrUtilization)
}

// UpdateTelegraf updates the base info of the chip
func (c *UtilizationCollector) UpdateTelegraf(fieldsMap map[string]map[string]interface{}, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) map[string]map[string]interface{} {
	caches := colcommon.GetInfoFromCache[chipUtilizationCache](n, colcommon.GetCacheKey(c))
	for _, chip := range chips {
		cache, ok := caches[chip.PhyId]
		if !ok {
			continue
		}
		fieldMap := getFieldMap(fieldsMap, cache.chip.LogicID)

		doUpdateTelegrafWithValidateNum(fieldMap, descUtil, float64(cache.Utilization), "")
		doUpdateTelegrafWithValidateNum(fieldMap, descVectorUtil, float64(cache.VectorUtilization), "")
		doUpdateTelegrafWithValidateNum(fieldMap, descCubeUtil, float64(cache.CubeUtilization), "")
		doUpdateTelegrafWithValidateNum(fieldMap, descOverUtil, float64(cache.OverallUtilization), "")
	}
	return fieldsMap
}

func collectUtil(c *UtilizationCollector, logicID int32, dmgr devmanager.DeviceInterface, chip *chipUtilizationCache) {
	if c.realGetDeviceUtilizationRateInfoFunc != nil {
		c.realGetDeviceUtilizationRateInfoFunc(logicID, dmgr, chip)
		return
	}
	buildDefaultMultiUtilInfo(chip)
	err := fmt.Errorf("realGetDeviceUtilizationRateInfoFunc is nil when get utilization info ")
	handleErr(err, "utilization", 0)
}

func buildDefaultMultiUtilInfo(chip *chipUtilizationCache) {
	chip.Utilization = -1
	chip.OverallUtilization = -1
	chip.VectorUtilization = -1
	chip.CubeUtilization = -1
}

func collectUtilV1(logicID int32, dmgr devmanager.DeviceInterface, chip *chipUtilizationCache) {
	buildDefaultMultiUtilInfo(chip)
	// aicore
	util, err := dmgr.GetDeviceUtilizationRate(logicID, common.AICore)
	handleErr(err, colcommon.DomainForAICoreUtilization, logicID)
	chip.Utilization = int(util)

	devType := dmgr.GetDevType()
	// ai vector
	if !notSupportedVectorUtilDevices[devType] {
		// only 910A does not support input type 12
		vecUtil, err := dmgr.GetDeviceUtilizationRate(logicID, common.VectorCore)
		handleErr(err, colcommon.DomainForVectorCoreUtilization, logicID)
		chip.VectorUtilization = int(vecUtil)
	} else {
		logger.LogfWithOptions(logger.WarnLevel, logger.LogOptions{Domain: "vectorUtil", ID: devType, MaxCounts: 1},
			"%v does not support utilization of vector", utils.MaskDevType(devType))
	}

	// overall
	if supportedOverallUtilDevices[devType] {
		// only A2/A3 support input type 13
		// A5 some product type support 13 , and some product type does not
		overAllUtil, err := dmgr.GetDeviceUtilizationRate(logicID, common.Overall)
		handleErr(err, colcommon.DomainForOverallUtilization, logicID)
		chip.OverallUtilization = int(overAllUtil)
	} else {
		logger.LogfWithOptions(logger.WarnLevel, logger.LogOptions{Domain: "overallUtil", ID: devType, MaxCounts: 1},
			"%v does not support utilization of overall", utils.MaskDevType(devType))
	}

	// ai cube
	msg := ""
	if supportedCubeDevices[devType] {
		// input type 14 is not supported when v2 api is not available
		msg = "%v does not support utilization of cube when v2 api is not available"
	} else {
		msg = "%v does not support utilization of cube"
	}
	logger.LogfWithOptions(logger.WarnLevel,
		logger.LogOptions{Domain: "cubeUtil", ID: devType, MaxCounts: 1}, msg, utils.MaskDevType(devType))
}

func collectUtilCommon(logicID int32, dmgr devmanager.DeviceInterface, chip *chipUtilizationCache) {
	multiUtilInfo, err := dmgr.GetDeviceUtilizationRateCommon(logicID)
	handleErr(err, "multiUtilInfoPeriod", logicID)
	chip.Utilization = int(multiUtilInfo.AicoreUtil)
	chip.OverallUtilization = int(multiUtilInfo.NpuUtil)
	chip.VectorUtilization = int(multiUtilInfo.AivUtil)
	chip.CubeUtilization = int(multiUtilInfo.AicUtil)
}
