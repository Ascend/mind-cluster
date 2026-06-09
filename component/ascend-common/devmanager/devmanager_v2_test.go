/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

// Package devmanager for device driver manager
package devmanager

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/devmanager/common"
	"ascend-common/devmanager/dcmi"
)

const (
	testV2LogicID    = int32(0)
	testV2AicUtil    = uint32(50)
	testV2AivUtil    = uint32(60)
	testV2AicoreUtil = uint32(70)
	testV2NpuUtil    = uint32(80)
	dcmiV2FailedMsg  = "dcmi v2 failed"
)

type getDeviceUtilizationRateV2PeriodV2TestCase struct {
	name         string
	logicID      int32
	setupPatches func(*DeviceManagerV2) *gomonkey.Patches
	expectError  bool
	expectedInfo common.DcmiMultiUtilizationInfo
}

func buildGetDeviceUtilizationRateV2PeriodV2TestCases() []getDeviceUtilizationRateV2PeriodV2TestCase {
	return []getDeviceUtilizationRateV2PeriodV2TestCase{
		{
			name:    "should return error when DcGetDeviceUtilizationRateV2Period failed",
			logicID: testV2LogicID,
			setupPatches: func(dm *DeviceManagerV2) *gomonkey.Patches {
				return gomonkey.ApplyMethodReturn(dm.DcMgr, "DcGetDeviceUtilizationRateV2Period",
					dcmi.BuildErrNpuMultiUtilizationInfo(), errors.New(dcmiV2FailedMsg))
			},
			expectError: true,
		},
		{
			name:    "should return success when DcGetDeviceUtilizationRateV2Period succeeds",
			logicID: testV2LogicID,
			setupPatches: func(dm *DeviceManagerV2) *gomonkey.Patches {
				return gomonkey.ApplyMethodReturn(dm.DcMgr, "DcGetDeviceUtilizationRateV2Period",
					common.DcmiMultiUtilizationInfo{
						AicUtil:    testV2AicUtil,
						AivUtil:    testV2AivUtil,
						AicoreUtil: testV2AicoreUtil,
						NpuUtil:    testV2NpuUtil,
					}, nil)
			},
			expectError: false,
			expectedInfo: common.DcmiMultiUtilizationInfo{
				AicUtil:    testV2AicUtil,
				AivUtil:    testV2AivUtil,
				AicoreUtil: testV2AicoreUtil,
				NpuUtil:    testV2NpuUtil,
			},
		},
	}
}

func TestGetDeviceUtilizationRateV2PeriodV2(t *testing.T) {
	for _, tt := range buildGetDeviceUtilizationRateV2PeriodV2TestCases() {
		t.Run(tt.name, func(t *testing.T) {
			dm := &DeviceManagerV2{DcMgr: &dcmi.DcV2Manager{}}
			var patches *gomonkey.Patches
			if tt.setupPatches != nil {
				patches = tt.setupPatches(dm)
				defer patches.Reset()
			}
			result, err := dm.GetDeviceUtilizationRateV2Period(tt.logicID)
			if tt.expectError {
				convey.Convey("", t, func() {
					convey.So(err, convey.ShouldNotBeNil)
				})
			} else {
				convey.Convey("", t, func() {
					convey.So(err, convey.ShouldBeNil)
					convey.So(result.AicUtil, convey.ShouldEqual, tt.expectedInfo.AicUtil)
					convey.So(result.AivUtil, convey.ShouldEqual, tt.expectedInfo.AivUtil)
					convey.So(result.AicoreUtil, convey.ShouldEqual, tt.expectedInfo.AicoreUtil)
					convey.So(result.NpuUtil, convey.ShouldEqual, tt.expectedInfo.NpuUtil)
				})
			}
		})
	}
}
