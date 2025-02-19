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
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"ascend-common/devmanager/common"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/utils/logger"
)

var (
	cardLabelForVNpuName                      = make([]string, len(colcommon.CardLabel))
	podAiCoreUtilizationRate *prometheus.Desc = nil
	podTotalMemory           *prometheus.Desc = nil
	podUsedMemory            *prometheus.Desc = nil
)

var (
	supportedVnpuDevices = map[string]bool{
		common.Ascend310P: true,
	}
)

const (
	vNpuUUID  = "v_dev_id"
	aiCoreCnt = "aicore_count"
	isVirtual = "is_virtual"
)

func init() {
	cardLabelForVNpuName = append(colcommon.CardLabel, isVirtual)
	cardLabelForVNpuName[2] = vNpuUUID
	cardLabelForVNpuName[3] = aiCoreCnt

	podAiCoreUtilizationRate = colcommon.BuildDescWithLabel("vnpu_pod_aicore_utilization",
		"the vnpu aicore utilization rate, unit is '%'", cardLabelForVNpuName)
	podTotalMemory = colcommon.BuildDescWithLabel("vnpu_pod_total_memory",
		"the vnpu total memory on pod, unit is 'KB'", cardLabelForVNpuName)
	podUsedMemory = colcommon.BuildDescWithLabel("vnpu_pod_used_memory",
		"the vnpu used memory on pod, unit is 'KB'", cardLabelForVNpuName)

}

// VnpuCollector collect vnpu info
type VnpuCollector struct {
	colcommon.MetricsCollectorAdapter
}

// UpdatePrometheus update prometheus metrics
func (c *VnpuCollector) UpdatePrometheus(ch chan<- prometheus.Metric, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) {

	updateSingleChip := func(cache chipCache, cardLabel []string) {
		aiChip := cache.chip

		containerName := getContainerNameArray(containerMap[aiChip.DeviceID])
		if len(containerName) != colcommon.ContainerNameLen {
			return
		}

		cardLabel = getPodDisplayInfo(&aiChip, containerName)
	}

	updateFrame[chipCache](colcommon.GetCacheKey(c), n, containerMap, chips, updateSingleChip)
}

// UpdateTelegraf update telegraf metrics
func (c *VnpuCollector) UpdateTelegraf(fieldsMap map[int]map[string]interface{}, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) map[int]map[string]interface{} {

	caches := colcommon.GetInfoFromCache[chipCache](n, colcommon.GetCacheKey(c))
	for _, chip := range chips {
		cache, ok := caches[chip.PhyId]
		if !ok {
			continue
		}
		fieldMap := getFieldMap(fieldsMap, cache.chip.LogicID)


		doUpdateTelegraf(fieldMap, podUsedMemory, nil, "")
	}
	return fieldsMap
}

func getPodDisplayInfo(chip *colcommon.HuaWeiAIChip, containerName []string) []string {
	if len(containerName) != colcommon.ContainerNameLen {
		logger.Logger.Logf(logger.Error, "container name length %v is not %v", len(containerName), colcommon.ContainerNameLen)
		return nil
	}

	chipInfo := common.DeepCopyChipInfo(chip.ChipInfo)

	return []string{
		strconv.Itoa(int(chip.DeviceID)),
		common.GetNpuName(chipInfo),
		containerName[colcommon.NameSpaceIdx],
		containerName[colcommon.PodNameIdx],
		containerName[colcommon.ConNameIdx],
	}
}
