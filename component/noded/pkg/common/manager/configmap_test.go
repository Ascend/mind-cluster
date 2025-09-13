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
package manager

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
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
			faultConfig := &common.FaultConfig{FaultTypeCode: &common.FaultTypeCode{
				NotHandleFaultCodes:   []string{"00000001"},
				PreSeparateFaultCodes: []string{"00000002"},
				SeparateFaultCodes:    []string{"00000003"},
			}}
			configManager.SetFaultConfig(faultConfig)
			FaultConfigEqual(configManager.GetFaultConfig(), faultConfig)
		})
		convey.Convey("wrong config manager set fault config", func() {
			configManager := mockWrongConfigManager()
			faultConfig := &common.FaultConfig{FaultTypeCode: &common.FaultTypeCode{
				NotHandleFaultCodes:   []string{"00000001"},
				PreSeparateFaultCodes: []string{"00000002"},
				SeparateFaultCodes:    []string{"00000003"},
			}}
			configManager.SetFaultConfig(faultConfig)
			convey.So(configManager.GetFaultConfig(), convey.ShouldBeNil)
		})
	})
}

// FaultConfigEqual judge if fault config is equal
func FaultConfigEqual(oldFaultConfig, newFaultConfig *common.FaultConfig) {
	convey.So(oldFaultConfig.FaultTypeCode, convey.ShouldNotBeNil)
	convey.So(newFaultConfig.FaultTypeCode, convey.ShouldNotBeNil)
	FaultTypeCodesEqual(oldFaultConfig.FaultTypeCode, newFaultConfig.FaultTypeCode)
}

// FaultTypeCodesEqual judge if fault type code is equal
func FaultTypeCodesEqual(oldFaultTypeCode, newFaultTypeCode *common.FaultTypeCode) {
	SliceStrEqual(oldFaultTypeCode.NotHandleFaultCodes, newFaultTypeCode.NotHandleFaultCodes)
	SliceStrEqual(oldFaultTypeCode.PreSeparateFaultCodes, newFaultTypeCode.PreSeparateFaultCodes)
	SliceStrEqual(oldFaultTypeCode.SeparateFaultCodes, newFaultTypeCode.SeparateFaultCodes)
}

// SliceStrEqual judge string slice is equal
func SliceStrEqual(slice1, slice2 []string) {
	convey.So(len(slice1), convey.ShouldEqual, len(slice2))
	if len(slice1) != len(slice2) {
		return
	}
	for i := range slice1 {
		convey.So(slice1[i], convey.ShouldEqual, slice2[i])
	}
}
