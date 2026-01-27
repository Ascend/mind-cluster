/*
Copyright(C) 2021-2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package prometheus for prometheus collector
package prom

import (
	"strconv"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/collector/metrics"
	"huawei.com/npu-exporter/v6/utils"
	"huawei.com/npu-exporter/v6/utils/logger"
)

const (
	maxMetricsCount         = 2000
	num2                    = 2
	num5                    = 5
	mockContainerName       = "mockContainerName"
	maxChipNum        int32 = 8
)

func TestDescribe(t *testing.T) {
	convey.Convey("test prometheus desc ", t, func() {
		collector := NewPrometheusCollector(nil)

		convey.Convey("test prometheus desc when ch is nil", func() {
			collector.Describe(nil)
		})
		convey.Convey("test prometheus desc when ch is not nil", func() {
			ch := make(chan *prometheus.Desc, maxMetricsCount)
			collector.Describe(ch)
			t.Logf("Describe len(ch):%v", len(ch))

			convey.So(ch, convey.ShouldNotBeEmpty)
		})
	})
}

func TestDescribeNilDesc(t *testing.T) {
	convey.Convey("should skip nil desc when received from tempCh", t, func() {
		collector := NewPrometheusCollector(nil)
		ch := make(chan *prometheus.Desc, maxMetricsCount)

		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(describeChain, func(ch chan<- *prometheus.Desc, chain []common.MetricsCollector) {
			ch <- nil
		})

		collector.Describe(ch)
		close(ch)

		descCount := 0
		for range ch {
			descCount++
		}
		convey.So(descCount, convey.ShouldEqual, 0)
	})
}

func TestDescribeDuplicateDesc(t *testing.T) {
	convey.Convey("should filter duplicate desc and keep first one", t, func() {
		collector := NewPrometheusCollector(nil)
		ch := make(chan *prometheus.Desc, maxMetricsCount)

		patches := gomonkey.NewPatches()
		defer patches.Reset()
		desc1 := prometheus.NewDesc("test_metric", "help1", nil, nil)
		desc2 := prometheus.NewDesc("test_metric", "help2", nil, nil)
		desc3 := prometheus.NewDesc("another_metric", "help3", nil, nil)
		sentDescs := []*prometheus.Desc{desc1, desc2, desc3}
		index := 0
		patches.ApplyFunc(describeChain, func(ch chan<- *prometheus.Desc, chain []common.MetricsCollector) {
			if index < len(sentDescs) {
				ch <- sentDescs[index]
				index++
			}
		})
		collector.Describe(ch)
		close(ch)

		descCount := 0
		seenNames := make(map[string]bool)
		for desc := range ch {
			if desc != nil {
				name := utils.GetDescName(desc)
				seenNames[name] = true
				descCount++
			}
		}

		convey.So(descCount, convey.ShouldEqual, num2)
		convey.So(seenNames["test_metric"], convey.ShouldBeTrue)
		convey.So(seenNames["another_metric"], convey.ShouldBeTrue)
	})
}

func TestCollect(t *testing.T) {
	convey.Convey("test prometheus collect ", t, func() {
		npuCollector := mockNewNpuCollector()
		collector := NewPrometheusCollector(npuCollector)

		convey.Convey("test prometheus collect when ch is nil", func() {
			collector.Collect(nil)
		})
		convey.Convey("test prometheus collect when ch is not nil", func() {

			ch := make(chan prometheus.Metric, maxMetricsCount)

			patches := gomonkey.NewPatches()
			collector.Collect(ch)

			patches.ApplyFuncReturn(common.GetChipListWithVNPU, mockGetNPUChipList())
			patches.ApplyFuncReturn(common.GetContainerNPUInfo, mockGetContainerNPUInfo())

			t.Logf("Describe len(ch):%v", len(ch))
			convey.So(ch, convey.ShouldNotBeEmpty)
		})
	})
}

func mockNewNpuCollector() *common.NpuCollector {
	tc := newNpuCollectorTestCase{
		cacheTime:    time.Duration(num5),
		updateTime:   time.Duration(num5),
		deviceParser: &container.DevicesParser{},
		dmgr:         &devmanager.DeviceManager{},
	}
	c := common.NewNpuCollector(tc.cacheTime, tc.updateTime, tc.deviceParser, tc.dmgr)
	return c
}

type newNpuCollectorTestCase struct {
	cacheTime    time.Duration
	updateTime   time.Duration
	deviceParser *container.DevicesParser
	dmgr         *devmanager.DeviceManager
}

func mockGetNPUChipList() []common.HuaWeiAIChip {
	chips := make([]common.HuaWeiAIChip, 0)
	for id := int32(0); id < maxChipNum; id++ {
		chip := common.HuaWeiAIChip{
			CardId:   id,
			PhyId:    id,
			DeviceID: id,
			LogicID:  id,
		}

		chips = append(chips, chip)
	}
	return chips
}

func mockGetContainerNPUInfo() map[int32]container.DevicesInfo {
	containsInfo := make(map[int32]container.DevicesInfo)
	for id := int32(0); id < maxChipNum; id++ {

		containerInfo := container.DevicesInfo{
			ID:      strconv.Itoa(int(id)),
			Name:    mockContainerName,
			Devices: []int{int(id)},
		}
		containsInfo[id] = containerInfo
	}
	return containsInfo
}

func init() {
	logger.HwLogConfig = &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	logger.InitLogger("Prometheus")

	initChain()
}

func initChain() {
	common.ChainForSingleGoroutine = []common.MetricsCollector{
		&metrics.HccsCollector{},
		&metrics.BaseInfoCollector{},
		&metrics.SioCollector{},
		&metrics.VersionCollector{},
		&metrics.HbmCollector{},
		&metrics.DdrCollector{},
		&metrics.VnpuCollector{},
		&metrics.PcieCollector{},
	}
	common.ChainForMultiGoroutine = []common.MetricsCollector{
		&metrics.NetworkCollector{},
		&metrics.RoceCollector{},
		&metrics.OpticalCollector{},
	}
}
