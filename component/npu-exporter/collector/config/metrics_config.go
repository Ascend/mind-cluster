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

	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/metrics"
	"huawei.com/npu-exporter/v6/utils/logger"
)

var (

	// singleGoroutineMap metrics in this map will be collected in single goroutine
	singleGoroutineMap = map[string]common.MetricsCollector{
		groupHccs:    &metrics.HccsCollector{},
		groupNpu:     &metrics.BaseInfoCollector{},
		groupSio:     &metrics.SioCollector{},
		groupVersion: &metrics.VersionCollector{},
		groupHbm:     &metrics.HbmCollector{},
		groupDDR:     &metrics.DdrCollector{},
		groupVnpu:    &metrics.VnpuCollector{},
		groupPcie:    &metrics.PcieCollector{},
	}
	// multiGoroutineMap metrics in this map will be collected in multi goroutine
	multiGoroutineMap = map[string]common.MetricsCollector{
		groupNetwork: &metrics.NetworkCollector{},
		groupRoce:    &metrics.RoceCollector{},
		groupOptical: &metrics.OpticalCollector{},
	}
	configs = []map[string]string{
		{metricsGroup: groupDDR, state: stateOn},
		{metricsGroup: groupHccs, state: stateOn},
		{metricsGroup: groupNpu, state: stateOn},
		{metricsGroup: groupNetwork, state: stateOn},
		{metricsGroup: groupPcie, state: stateOn},
		{metricsGroup: groupRoce, state: stateOn},
		{metricsGroup: groupSio, state: stateOn},
		{metricsGroup: groupVnpu, state: stateOn},
		{metricsGroup: groupVersion, state: stateOn},
		{metricsGroup: groupOptical, state: stateOn},
		{metricsGroup: groupHbm, state: stateOn},
	}
)

const (
	metricsGroup = "metricsGroup"
	state        = "state"

	groupDDR     = "ddr"
	groupHccs    = "hccs"
	groupNpu     = "npu"
	groupNetwork = "network"
	groupPcie    = "pcie"
	groupRoce    = "roce"
	groupSio     = "sio"
	groupVnpu    = "vnpu"
	groupVersion = "version"
	groupOptical = "optical"
	groupHbm     = "hbm"

	stateOn  = "ON"
	stateOFF = "OFF"
)

// Register register collector to cache
func Register(n *common.NpuCollector) {

	for _, config := range configs {
		metricsGroupName := config[metricsGroup]

		if config[state] != stateOn {
			logger.Infof("metricsGroup [%v] is off", metricsGroupName)
			continue
		}
		collector, exist := singleGoroutineMap[metricsGroupName]
		if exist && collector.IsSupported(n) {
			common.ChainForSingleGoroutine = append(common.ChainForSingleGoroutine, collector)
		}

		collector, exist = multiGoroutineMap[metricsGroupName]
		if exist && collector.IsSupported(n) {
			common.ChainForMultiGoroutine = append(common.ChainForMultiGoroutine, collector)
		}
	}
	logger.Debugf("ChainForSingleGoroutine:%#v", common.ChainForSingleGoroutine)
	logger.Debugf("ChainForMultiGoroutine:%#v", common.ChainForMultiGoroutine)
}

// UnRegister delete collector from chain
func UnRegister(worker reflect.Type) {
	logger.Debugf("unRegister collector:%v", worker)
	unRegisterChain(worker, &common.ChainForSingleGoroutine)
	unRegisterChain(worker, &common.ChainForMultiGoroutine)
}

func unRegisterChain(worker reflect.Type, chain *[]common.MetricsCollector) {
	newChain := make([]common.MetricsCollector, 0)
	for _, collector := range *chain {
		if reflect.TypeOf(collector) != worker {
			newChain = append(newChain, collector)
		}
	}
	*chain = newChain
}
