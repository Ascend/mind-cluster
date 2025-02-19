/* Copyright(C) 2021-2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package npu this for parse and pack
package npu

import (
	_ "embed"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"ascend-common/devmanager/common"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/utils/logger"
)

//go:embed sample.conf
var sampleConfig string

// WatchNPU npu watch struct
type WatchNPU struct {
	collector *colcommon.NpuCollector
}

// SampleConfig used to return sampleConfig
func (*WatchNPU) SampleConfig() string {
	return sampleConfig
}

// Gather used to gather information from dcmi info and hccn tool info
func (npu *WatchNPU) Gather(acc telegraf.Accumulator) error {

	fieldsMap := make(map[int]map[string]interface{})
	const devName = "ascend"

	devTagValue := "unsupported"
	if cardType := npu.collector.Dmgr.GetDevType(); cardType == common.Ascend910A3 || cardType == common.Ascend910B ||
		cardType == common.Ascend910 {
		devTagValue = common.Chip910
	}
	logger.Logger.DynamicConfigure(logger.Config{Acc: acc})

	containerMap := make(map[int32]container.DevicesInfo)
	chips := colcommon.GetChipListWithVNPU(npu.collector)

	fieldsMap = npu.gatherChain(fieldsMap, colcommon.ChainForSingleGoroutine, containerMap, chips)
	fieldsMap = npu.gatherChain(fieldsMap, colcommon.ChainForMultiGoroutine, containerMap, chips)

	generalFields := fieldsMap[colcommon.GeneralDevTagKey]
	acc.AddFields(devName, generalFields, map[string]string{"device": devName + devTagValue})

	// after the report is completed, deleted to avoid repeated reporting in the for loop
	delete(fieldsMap, colcommon.GeneralDevTagKey)
	for logicID, fields := range fieldsMap {
		devTag := map[string]string{"device": devName + devTagValue + "-" + strconv.Itoa(logicID)}
		acc.AddFields(devName, fields, devTag)
	}

	return nil
}

func (npu *WatchNPU) gatherChain(fieldsMap map[int]map[string]interface{}, chain []colcommon.MetricsCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) map[int]map[string]interface{} {

	for _, collector := range chain {
		fieldsMap = collector.UpdateTelegraf(fieldsMap, npu.collector, containerMap, chips)
	}
	return fieldsMap
}

func init() {
	inputs.Add("npu", func() telegraf.Input {
		return &WatchNPU{
			collector: colcommon.Collector,
		}
	})
}
