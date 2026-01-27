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

package plugins

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/config"
	"huawei.com/npu-exporter/v6/utils/logger"
)

const (
	testPluginName        = "text"
	registerTestErrorMsg  = "test error"
	testPluginNameDup     = "duplicate_plugin"
	pluginAlreadyExistMsg = "plugin collector already exist"
)

func TestRegisterPlugin(t *testing.T) {
	convey.Convey("should register text plugin successfully when AddPluginCollector succeeds", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFuncReturn(config.AddPluginCollector, nil)
		patches.ApplyFunc(logger.Errorf, func(format string, args ...interface{}) {})

		RegisterPlugin()
	})

	convey.Convey("should log error when AddPluginCollector fails", t, func() {
		testErr := errors.New(registerTestErrorMsg)
		var logCalled bool
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFuncReturn(config.AddPluginCollector, testErr)
		patches.ApplyFunc(logger.Errorf, func(format string, args ...interface{}) {
			logCalled = true
		})

		RegisterPlugin()
		convey.So(logCalled, convey.ShouldBeTrue)
	})
}

func TestRegisterPluginWithTable(t *testing.T) {
	testCases := []struct {
		name         string
		pluginName   string
		collector    common.MetricsCollector
		addPluginErr error
		shouldLogErr bool
	}{
		{
			name:         "should succeed when AddPluginCollector returns nil",
			pluginName:   testPluginName,
			collector:    &TextMetricsInfoCollector{},
			addPluginErr: nil,
			shouldLogErr: false,
		},
		{
			name:         "should log error when AddPluginCollector returns error",
			pluginName:   testPluginName,
			collector:    &TextMetricsInfoCollector{},
			addPluginErr: errors.New(registerTestErrorMsg),
			shouldLogErr: true,
		},
		{
			name:         "should log error when plugin already exists",
			pluginName:   testPluginNameDup,
			collector:    &TextMetricsInfoCollector{},
			addPluginErr: errors.New(pluginAlreadyExistMsg),
			shouldLogErr: true,
		},
	}

	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			var logCalled bool
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyFuncReturn(config.AddPluginCollector, tc.addPluginErr)
			patches.ApplyFunc(logger.Errorf, func(format string, args ...interface{}) {
				logCalled = true
			})

			registerPlugin(tc.pluginName, tc.collector)
			if tc.shouldLogErr {
				convey.So(logCalled, convey.ShouldBeTrue)
			} else {
				convey.So(logCalled, convey.ShouldBeFalse)
			}
		})
	}
}
