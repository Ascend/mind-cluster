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

// Package plugins for custom metrics
package plugins

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"ascend-common/api"
	commonutils "ascend-common/common-utils/utils"
	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/utils"
	"huawei.com/npu-exporter/v6/utils/logger"
)

var (
	machineInfoCardDesc = prometheus.NewDesc("machine_card_nums", "Amount of card installed on the machine.", nil, nil)

	supportedCardNumDevices = map[string]bool{
		api.Ascend310P:  true,
		api.Ascend910A:  true,
		api.Ascend910B:  true,
		api.Ascend910A3: true,
	}
)

const (
	machineInfoCardDescKey   = "machineCardNum"
	machineCardNumPluginName = "machineCardNum"
)

// MachineCardNumPluginInfoCollector collect machine_card_num plugin info
type MachineCardNumPluginInfoCollector struct {
	common.MetricsCollectorAdapter
	Cache sync.Map
}

// IsSupported check whether the metric is supported
func (c *MachineCardNumPluginInfoCollector) IsSupported(n *common.NpuCollector) bool {
	isSupport := supportedCardNumDevices[n.Dmgr.GetDevType()]
	if !isSupport {
		logger.Infof("devType %v does not support [%v], %v",
			commonutils.MaskDevType(n.Dmgr.GetDevType()), common.GetCacheKey(c), "")
	}
	return isSupport
}

// Describe description of the metric
func (c *MachineCardNumPluginInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	// add desc
	ch <- machineInfoCardDesc
}

// CollectToCache collect the metric to cache
func (c *MachineCardNumPluginInfoCollector) CollectToCache(n *common.NpuCollector, chipList []common.HuaWeiAIChip) {
	// collect metric to cache
	cardNum, _, err := n.Dmgr.GetCardList()
	if err != nil {
		logger.Error(err)
		return
	}
	c.Cache.Store(machineInfoCardDescKey, cardNum)
}

// UpdatePrometheus update prometheus metric
func (c *MachineCardNumPluginInfoCollector) UpdatePrometheus(ch chan<- prometheus.Metric, n *common.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []common.HuaWeiAIChip) {
	// update machine_card_nums
	machineInfoCardCache, ok := c.Cache.Load(machineInfoCardDescKey)
	if ok {
		value := float64(machineInfoCardCache.(int32))
		ch <- prometheus.MustNewConstMetric(machineInfoCardDesc, prometheus.GaugeValue, value)
	}
}

// UpdateTelegraf update telegraf metric
func (c *MachineCardNumPluginInfoCollector) UpdateTelegraf(fieldsMap map[string]map[string]interface{}, n *common.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []common.HuaWeiAIChip) map[string]map[string]interface{} {
	// update machine_card_nums
	machineInfoCardCache, ok := c.Cache.Load(machineInfoCardDescKey)
	if ok {
		value := float64(machineInfoCardCache.(int32))
		if fieldsMap[common.GeneralDevTagKey] == nil {
			fieldsMap[common.GeneralDevTagKey] = make(map[string]interface{})
		}
		utils.DoUpdateTelegraf(fieldsMap[common.GeneralDevTagKey], machineInfoCardDesc, value, "")
	}
	return fieldsMap
}
