/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/dcmi"
)

// TestGetCardIdAndDeviceId test the getCardIdAndDeviceId function
func TestGetCardIdAndDeviceId(t *testing.T) {

	var (
		cardId, deviceId = int32(0), int32(0)
		err              error
		returnValue      = int32(0)
		errReturnValue   = int32(-1)
	)
	manager := &DeviceManager{DcMgr: &dcmi.DcManager{}}
	convey.Convey("failed to get info by dcmi", t, func() {
		mk2 := gomonkey.ApplyMethodReturn(manager.DcMgr, "DcGetCardIDDeviceID",
			errReturnValue, errReturnValue, errors.New("mock err"))
		defer mk2.Reset()
		cardId, deviceId, err = manager.getCardIdAndDeviceId(0)

		convey.So(cardId, convey.ShouldEqual, common.RetError)
		convey.So(deviceId, convey.ShouldEqual, common.RetError)
		convey.So(err, convey.ShouldNotBeNil)

	})

	mk := gomonkey.ApplyMethodReturn(manager.DcMgr, "DcGetCardIDDeviceID", returnValue, returnValue, nil)
	defer mk.Reset()

	convey.Convey("get info from dcmi", t, func() {
		testGetCardIdAndDeviceId(t, cardId, deviceId, err, manager)
	})
	convey.Convey("get info from cache", t, func() {
		testGetCardIdAndDeviceId(t, cardId, deviceId, err, manager)
	})

}

func testGetCardIdAndDeviceId(t *testing.T, cardId int32, deviceId int32, err error, manager *DeviceManager) {
	cardId, deviceId, err = manager.getCardIdAndDeviceId(0)

	convey.So(cardId, convey.ShouldEqual, 0)
	convey.So(deviceId, convey.ShouldEqual, 0)
	convey.So(err, convey.ShouldBeNil)

}
func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&config, nil)
}

const (
	testLogicID0   = int32(0)
	testCardID0    = int32(0)
	testDeviceID0  = int32(0)
	testAicAvgUtil = uint32(50)
	testAivAvgUtil = uint32(60)
	testAicoreUtil = uint32(70)
	testNpuUtil    = uint32(80)
	getIdFailedMsg = "get id failed"
	dcmiFailedMsg  = "dcmi failed"
)

type dcGetDeviceUtilizationRateV2TestCase struct {
	name         string
	logicID      int32
	setupPatches func(*DeviceManager) *gomonkey.Patches
	expectError  bool
}

func buildDcGetDeviceUtilizationRateV2TestCases() []dcGetDeviceUtilizationRateV2TestCase {

	return []dcGetDeviceUtilizationRateV2TestCase{
		{name: "should return error when getCardIdAndDeviceId failed",
			logicID: testLogicID0,
			setupPatches: func(dm *DeviceManager) *gomonkey.Patches {
				return gomonkey.ApplyPrivateMethod(reflect.TypeOf(dm), "getCardIdAndDeviceId",
					func(*DeviceManager, int32) (int32, int32, error) {
						return common.RetError, common.RetError, errors.New(getIdFailedMsg)
					})
			},
			expectError: true},
		{name: "should return error when DcMgr.GetDeviceUtilizationRateV2 failed",
			logicID: testLogicID0,
			setupPatches: func(dm *DeviceManager) *gomonkey.Patches {
				return gomonkey.NewPatches().
					ApplyPrivateMethod(reflect.TypeOf(dm), "getCardIdAndDeviceId",
						func(*DeviceManager, int32) (int32, int32, error) {
							return testCardID0, testDeviceID0, nil
						}).
					ApplyMethodReturn(dm.DcMgr, "GetDeviceUtilizationRateV2",
						dcmi.BuildErrNpuMultiUtilizationInfo(), errors.New(dcmiFailedMsg))
			},
			expectError: true},
		{name: "should return success when all operations succeed",
			logicID: testLogicID0,
			setupPatches: func(dm *DeviceManager) *gomonkey.Patches {
				return gomonkey.NewPatches().
					ApplyPrivateMethod(reflect.TypeOf(dm), "getCardIdAndDeviceId",
						func(*DeviceManager, int32) (int32, int32, error) {
							return testCardID0, testDeviceID0, nil
						}).
					ApplyMethodReturn(dm.DcMgr, "GetDeviceUtilizationRateV2",
						common.NpuMultiUtilizationInfo{
							AicAvgUtil: testAicAvgUtil,
							AivAvgUtil: testAivAvgUtil,
							AicoreUtil: testAicoreUtil,
							NpuUtil:    testNpuUtil,
						}, nil)
			},
			expectError: false,
		},
	}
}

func TestDcGetDeviceUtilizationRateV2(t *testing.T) {
	convey.Convey("TestDcGetDeviceUtilizationRateV2", t, func() {
		for _, tt := range buildDcGetDeviceUtilizationRateV2TestCases() {
			convey.Convey(tt.name, func() {
				dm := &DeviceManager{DcMgr: &dcmi.DcManager{}}
				var patches *gomonkey.Patches
				if tt.setupPatches != nil {
					patches = tt.setupPatches(dm)
					defer patches.Reset()
				}
				result, err := dm.GetDeviceUtilizationRateV2(tt.logicID)
				if tt.expectError {
					convey.So(err, convey.ShouldNotBeNil)
				} else {
					convey.So(err, convey.ShouldBeNil)
					convey.So(result.AicAvgUtil, convey.ShouldEqual, testAicAvgUtil)
					convey.So(result.AivAvgUtil, convey.ShouldEqual, testAivAvgUtil)
					convey.So(result.AicoreUtil, convey.ShouldEqual, testAicoreUtil)
					convey.So(result.NpuUtil, convey.ShouldEqual, testNpuUtil)
				}
			})
		}
	})
}
