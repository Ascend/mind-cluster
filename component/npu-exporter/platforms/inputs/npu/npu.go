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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"ascend-common/api"
	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/utils/logger"
)

//go:embed sample.conf
var sampleConfig string

const (
	devName       = "ascend"
	deviceTagKey  = "device"
	vDevTagKey    = "vdev_id"
	chanCacheSize = 128
)

// WatchNPU npu watch struct
type WatchNPU struct {
	collector *common.NpuCollector
}

// SampleConfig used to return sampleConfig
func (*WatchNPU) SampleConfig() string {
	return sampleConfig
}

// Gather used to gather information from dcmi info and hccn tool info
func (npu *WatchNPU) Gather(acc telegraf.Accumulator) error {
	devTagValue := getDevTagValue(npu.collector.Dmgr.GetDevType())
	logger.DynamicConfigure(logger.Config{Acc: acc})

	containerMap := common.GetContainerNPUInfo(npu.collector)
	chips := common.GetChipListWithVNPU(npu.collector)
	single, multi, plugin := common.GetChainsSnapshot()

	ch := make(chan common.TelegrafMetric, chanCacheSize)
	done := make(chan struct{})
	go func() {
		consumeAndReport(acc, ch, devTagValue)
		close(done)
	}()

	for _, chain := range [][]common.MetricsCollector{single, multi, plugin} {
		npu.collectChain(ch, chain, containerMap, chips)
	}
	close(ch)
	<-done

	return nil
}

func getDevTagValue(cardType string) string {
	if cardType == api.Ascend910A3 || cardType == api.Ascend910B || cardType == api.Ascend910A {
		return strings.ToLower(api.Ascend910)
	}
	if cardType == api.Ascend910A5 {
		return api.NPULowerCase
	}
	return strings.ToLower(cardType)
}

func (npu *WatchNPU) collectChain(ch chan<- common.TelegrafMetric, chain []common.MetricsCollector,
	containerMap map[int32]container.DevicesInfo, chips []common.HuaWeiAIChip) {
	for _, collector := range chain {
		if collector == nil {
			continue
		}
		collector.UpdateTelegraf(ch, npu.collector, containerMap, chips)
	}
}

type metricGroup struct {
	measurement string
	tags        map[string]string
	fields      map[string]interface{}
	timestamp   time.Time
}

// consumeAndReport drains the channel, applies the default device tag, aggregates data by
// measurement + labels, dedups duplicated fields (keeping the first declaration) and reports.
func consumeAndReport(acc telegraf.Accumulator, ch <-chan common.TelegrafMetric, devTagValue string) {
	groups := make(map[string]*metricGroup)
	order := make([]string, 0)
	for metric := range ch {
		if len(metric.Fields) == 0 {
			continue
		}
		measurement := metric.Measurement
		if measurement == "" {
			measurement = devName
		}
		tags := buildTags(metric, devTagValue)
		key := groupKey(measurement, tags)
		group, ok := groups[key]
		if !ok {
			group = &metricGroup{
				measurement: measurement,
				tags:        tags,
				fields:      make(map[string]interface{}),
				timestamp:   metric.Timestamp,
			}
			groups[key] = group
			order = append(order, key)
		}
		mergeFields(group.fields, metric.Fields)
	}

	for _, key := range order {
		group := groups[key]
		if len(group.fields) == 0 {
			continue
		}
		if group.timestamp.IsZero() {
			acc.AddFields(group.measurement, group.fields, group.tags)
			continue
		}
		acc.AddFields(group.measurement, group.fields, group.tags, group.timestamp)
	}
}

func buildTags(data common.TelegrafMetric, devTagValue string) map[string]string {
	if data.Labels != nil {
		return data.Labels
	}
	if data.DeviceID < 0 {
		return map[string]string{deviceTagKey: devTagValue}
	}
	tags := map[string]string{deviceTagKey: devTagValue + "-" + strconv.Itoa(int(data.DeviceID))}
	if data.VDevID >= 0 {
		tags[vDevTagKey] = strconv.Itoa(int(data.VDevID))
	}
	return tags
}

func groupKey(measurement string, tags map[string]string) string {
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var builder strings.Builder
	builder.WriteString(measurement)
	for _, k := range keys {
		builder.WriteString("|")
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(tags[k])
	}
	return builder.String()
}

func mergeFields(dst, src map[string]interface{}) {
	for name, value := range src {
		if _, exists := dst[name]; exists {
			logger.Warnf("duplicate metric field detected, keeping first declaration, ignoring duplicate: %s", name)
			continue
		}
		dst[name] = value
	}
}

func init() {
	inputs.Add("npu", func() telegraf.Input {
		return &WatchNPU{
			collector: common.Collector,
		}
	})
}
