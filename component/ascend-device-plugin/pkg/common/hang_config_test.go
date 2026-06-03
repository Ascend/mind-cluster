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

package common

import (
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		panic(err)
	}
}

// TestLoadHangDetectionConfigFromFile for test LoadHangDetectionConfigFromFile
func TestLoadHangDetectionConfigFromFile(t *testing.T) {
	globalHangConfig := hangConfig
	convey.Convey("test LoadHangDetectionConfigFromFile", t, func() {
		convey.Convey("01-when load file failed, should keep default config", func() {
			hangConfig = globalHangConfig
			patches := gomonkey.ApplyFuncReturn(utils.LoadFile, nil, fmt.Errorf("file not found"))
			defer patches.Reset()

			LoadHangDetectionConfigFromFile()
			convey.So(hangConfig.HangDetection.Threshold.AICoreUtilization, convey.ShouldEqual, defaultUtilizationLow)
		})

		convey.Convey("02-when load file success with valid config, should update config", func() {
			hangConfig = globalHangConfig
			validJSON := `{"HangDetection":{"Enabled":false,"Threshold":{"AICoreUtilization":-1,"HbmMemoryDelta":-1,"TrafficDelta":-1,"CPUTimeDelta":-1,"DetectDuration":0}}}`
			patches := gomonkey.ApplyFuncReturn(utils.LoadFile, []byte(validJSON), nil)
			defer patches.Reset()

			LoadHangDetectionConfigFromFile()
			convey.So(hangConfig.HangDetection.Enabled, convey.ShouldBeFalse)
			convey.So(hangConfig.HangDetection.Threshold.AICoreUtilization, convey.ShouldEqual, defaultUtilizationLow)
			convey.So(hangConfig.HangDetection.Threshold.HbmMemoryDelta, convey.ShouldEqual, defaultMemoryLow)
			convey.So(hangConfig.HangDetection.Threshold.TrafficDelta, convey.ShouldEqual, defaultTrafficLow)
			convey.So(hangConfig.HangDetection.Threshold.CPUTimeDelta, convey.ShouldEqual, defaultCPUTimeLow)
			convey.So(hangConfig.HangDetection.Threshold.DetectDuration, convey.ShouldEqual, defaultDetectDuration)
		})
	})
	hangConfig = globalHangConfig
}

// TestIsHangDetectionEnabled for test IsHangDetectionEnabled
func TestIsHangDetectionEnabled(t *testing.T) {
	convey.Convey("test IsHangDetectionEnabled", t, func() {
		convey.Convey("01-when Enabled is true, should return true", func() {
			hangConfig.HangDetection.Enabled = true
			convey.So(IsHangDetectionEnabled(), convey.ShouldBeTrue)
		})
	})
}

// TestGetHangDetectionThreshold for test GetHangDetectionThreshold
func TestGetHangDetectionThreshold(t *testing.T) {
	convey.Convey("test GetHangDetectionThreshold", t, func() {
		convey.Convey("01-should return current threshold config", func() {
			threshold := GetHangDetectionThreshold()
			convey.So(threshold.AICoreUtilization, convey.ShouldEqual, defaultUtilizationLow)
		})

		convey.Convey("02-should return updated threshold after config change", func() {
			utilizationLow := int32(20)
			hangConfig.HangDetection.Threshold = HangThreshold{AICoreUtilization: utilizationLow}
			threshold := GetHangDetectionThreshold()
			convey.So(threshold.AICoreUtilization, convey.ShouldEqual, utilizationLow)
		})
	})
}
