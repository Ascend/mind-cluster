/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package config for general collector
package config

import (
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/metrics"
	"huawei.com/npu-exporter/v6/utils/logger"
)

func init() {
	logger.HwLogConfig = &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	logger.InitLogger("Prometheus")
	initChain()
}

func initChain() {
	common.ChainForSingleGoroutine = []common.MetricsCollector{}
	common.ChainForMultiGoroutine = []common.MetricsCollector{}
}

func TestRegister(t *testing.T) {
	convey.Convey("TestRegister", t, func() {
		n := &common.NpuCollector{}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		// Mock IsSupported method to always return true
		patches.ApplyMethodReturn(&metrics.HccsCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.BaseInfoCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.SioCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.VersionCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.HbmCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.DdrCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.VnpuCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.PcieCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.NetworkCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.RoceCollector{}, "IsSupported", true)
		patches.ApplyMethodReturn(&metrics.OpticalCollector{}, "IsSupported", true)
		configs = append(configs, map[string]string{metricsGroup: "mockGroup", state: stateOFF})

		Register(n)
		convey.Convey("Should add collectors to ChainForSingleGoroutine", func() {
			convey.So(len(common.ChainForSingleGoroutine), convey.ShouldBeGreaterThan, 0)
		})
		convey.Convey("Should add collectors to ChainForMultiGoroutine", func() {
			convey.So(len(common.ChainForMultiGoroutine), convey.ShouldBeGreaterThan, 0)
		})
	})
}

func TestUnRegister(t *testing.T) {
	convey.Convey("TestUnRegister", t, func() {
		// Initialize chains with some collectors
		common.ChainForSingleGoroutine = []common.MetricsCollector{
			&metrics.HccsCollector{},
			&metrics.BaseInfoCollector{},
		}
		common.ChainForMultiGoroutine = []common.MetricsCollector{
			&metrics.NetworkCollector{},
			&metrics.RoceCollector{},
		}

		convey.Convey("When UnRegister is called with HccsCollector type", func() {
			UnRegister(reflect.TypeOf(&metrics.HccsCollector{}))

			convey.Convey("Should remove HccsCollector from ChainForSingleGoroutine", func() {
				expected := []common.MetricsCollector{
					&metrics.BaseInfoCollector{},
				}
				convey.So(len(common.ChainForSingleGoroutine), convey.ShouldEqual, len(expected))
				for i, collector := range common.ChainForSingleGoroutine {
					convey.So(reflect.TypeOf(collector), convey.ShouldEqual, reflect.TypeOf(expected[i]))
				}
			})

			convey.Convey("Should not affect ChainForMultiGoroutine", func() {
				expected := []common.MetricsCollector{
					&metrics.NetworkCollector{},
					&metrics.RoceCollector{},
				}
				convey.So(len(common.ChainForMultiGoroutine), convey.ShouldEqual, len(expected))
				for i, collector := range common.ChainForMultiGoroutine {
					convey.So(reflect.TypeOf(collector), convey.ShouldEqual, reflect.TypeOf(expected[i]))
				}
			})
		})
	})
}

func TestUnRegisterChain(t *testing.T) {
	convey.Convey("TestUnRegisterChain", t, func() {
		// Initialize a chain with some collectors
		chain := []common.MetricsCollector{
			&metrics.HccsCollector{},
			&metrics.BaseInfoCollector{},
			&metrics.NetworkCollector{},
		}

		convey.Convey("When unRegisterChain is called with BaseInfoCollector type", func() {
			unRegisterChain(reflect.TypeOf(&metrics.BaseInfoCollector{}), &chain)
			convey.Convey("Should remove BaseInfoCollector from the chain", func() {
				expected := []common.MetricsCollector{
					&metrics.HccsCollector{},
					&metrics.NetworkCollector{},
				}
				convey.So(len(chain), convey.ShouldEqual, len(expected))
				for i, collector := range chain {
					convey.So(reflect.TypeOf(collector), convey.ShouldEqual, reflect.TypeOf(expected[i]))
				}
			})
		})
	})
}
