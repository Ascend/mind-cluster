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
	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/config"
	"huawei.com/npu-exporter/v6/utils/logger"
)

// RegisterPlugin register plugin collector
func RegisterPlugin() {
	registerPlugin("text", &TextMetricsInfoCollector{})

	// Add custom plugin，pluginName should be consistent with the name in pluginConfiguration.json
	// The path of the pluginConfiguration.json file must be /usr/local/pluginConfiguration.json
}

func registerPlugin(pluginName string, c common.MetricsCollector) {
	err := config.AddPluginCollector(pluginName, c)
	if err != nil {
		logger.Errorf("%v", err)
	}
}
