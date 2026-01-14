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

// Package externalbridge is a DT collection for func in data_patse_bridge_test
package externalbridge

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/business/handlejson"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/sdk"
	"ascend-faultdiag-online/pkg/core/model"
)

func TestStartDataParse(t *testing.T) {
	convey.Convey("test StartDataParse", t, func() {
		callCount := 0
		patch := gomonkey.ApplyFunc(sdk.StartParse, func(config.DataParseModel) {
			callCount++
		})
		defer patch.Reset()
		StartDataParse(config.DataParseModel{})
		convey.So(callCount, convey.ShouldEqual, 1)
	})
}

func TestStopDataParse(t *testing.T) {
	convey.Convey("test StopDataParse", t, func() {
		callCount := 0
		patch := gomonkey.ApplyFunc(sdk.StopParse, func(config.DataParseModel) {
			callCount++
		})
		defer patch.Reset()
		StopDataParse(config.DataParseModel{})
		convey.So(callCount, convey.ShouldEqual, 1)
	})
}

func TestReloadDataParse(t *testing.T) {
	convey.Convey("test ReloadDataParse", t, func() {
		callCount := 0
		patch := gomonkey.ApplyFunc(sdk.ReloadParse, func(config.DataParseModel) {
			callCount++
		})
		defer patch.Reset()
		ReloadDataParse(config.DataParseModel{})
		convey.So(callCount, convey.ShouldEqual, 1)
	})
}

func TestRegisterDataParse(t *testing.T) {
	convey.Convey("test RegisterDataParse", t, func() {
		callCount := 0
		patch := gomonkey.ApplyFunc(sdk.RegisterParseCallback, func(model.CallbackFunc) {
			callCount++
		})
		defer patch.Reset()
		RegisterDataParse(func(string) {})
		convey.So(callCount, convey.ShouldEqual, 1)
	})
}

func TestStartMergeParGroupInfo(t *testing.T) {
	convey.Convey("test StartMergeParGroupInfo", t, func() {
		patch := gomonkey.ApplyFuncReturn(handlejson.MergeParallelGroupInfo, nil)
		defer patch.Reset()
		convey.So(func() {
			StartMergeParGroupInfo(config.DataParseModel{})
		}, convey.ShouldNotPanic)
	})
}

func TestRegisterMergeParGroup(t *testing.T) {
	convey.Convey("test RegisterMergeParGroup", t, func() {
		callCount := 0
		patch := gomonkey.ApplyFunc(handlejson.RegisterParGroupCallback, func(model.CallbackFunc) {
			callCount++
		})
		defer patch.Reset()
		RegisterMergeParGroup(func(string) {})
		convey.So(callCount, convey.ShouldEqual, 1)
	})
}
