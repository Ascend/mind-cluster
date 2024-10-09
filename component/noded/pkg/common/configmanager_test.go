/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common for common function
package common

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func mockWrongConfigManager() ConfigManager {
	return &ConfigTools{}
}

// TestSetFaultConfig test the function of set fault config
func TestSetFaultConfig(t *testing.T) {
	convey.Convey("test set fault config", t, func() {
		convey.Convey("config manager set fault config", func() {
			configManager := NewConfigManager()
			faultConfig := &FaultConfig{FaultTypeCode: &FaultTypeCode{
				NotHandleFaultCodes:   []string{"00000001"},
				PreSeparateFaultCodes: []string{"00000002"},
				SeparateFaultCodes:    []string{"00000003"},
			}}
			configManager.SetFaultConfig(faultConfig)
			FaultConfigEqual(configManager.GetFaultConfig(), faultConfig)
		})
		convey.Convey("wrong config manager set fault config", func() {
			configManager := mockWrongConfigManager()
			faultConfig := &FaultConfig{FaultTypeCode: &FaultTypeCode{
				NotHandleFaultCodes:   []string{"00000001"},
				PreSeparateFaultCodes: []string{"00000002"},
				SeparateFaultCodes:    []string{"00000003"},
			}}
			configManager.SetFaultConfig(faultConfig)
			convey.So(configManager.GetFaultConfig(), convey.ShouldBeNil)
		})
	})
}

// TestSetFaultTypeCodes test the function of set fault type codes
func TestSetFaultTypeCodes(t *testing.T) {
	convey.Convey("test set fault type codes", t, func() {
		convey.Convey("config manager set fault type codes", func() {
			configManager := NewConfigManager()
			faultTypeCodes := &FaultTypeCode{
				NotHandleFaultCodes:   []string{"00000001"},
				PreSeparateFaultCodes: []string{"00000002"},
				SeparateFaultCodes:    []string{"00000003"},
			}
			configManager.SetFaultTypeCode(faultTypeCodes)
			FaultTypeCodesEqual(configManager.GetFaultTypeCode(), faultTypeCodes)
		})
		convey.Convey("wrong config manager set fault type codes", func() {
			configManager := mockWrongConfigManager()
			faultTypeCodes := &FaultTypeCode{
				NotHandleFaultCodes:   []string{"00000001"},
				PreSeparateFaultCodes: []string{"00000002"},
				SeparateFaultCodes:    []string{"00000003"},
			}
			configManager.SetFaultTypeCode(faultTypeCodes)
			convey.So(configManager.GetFaultTypeCode(), convey.ShouldBeNil)
		})
	})
}
