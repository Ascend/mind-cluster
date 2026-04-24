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
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/dcmi"
)

const (
	testCardID       = 2
	testDeviceID     = 3
	testErrCount     = 5
	testRetError     = -1
	errGetCardDevice = "failed to get cardID and deviceID in get device error code by logicID(%d)"
	errGetErrorCode  = "failed to get device error code by logicID(%d)"
	timeout          = 5 * time.Second
)

const (
	testLogicID = int32(100)
	testPortID  = int32(10)
	testTime    = 10
)

var (
	mockDcmiVersion           = "24.0.rc2"
	mockCardNum         int32 = 16
	mockDeviceNumInCard int32 = 1
	mockCardList              = []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	mockProductType           = ""
	mockErr                   = errors.New("test error")
	mockLogicId               = int32(0)
	mock910ChipInfo           = &common.ChipInfo{
		Type:    "Ascend",
		Name:    "910_7591",
		Version: "V1",
	}
	mockBoardInfo = common.BoardInfo{
		BoardId: 0x28,
	}
	mockMainBoardId = uint32(0x01)
)

var (
	mockDeviceNum   int32 = 8
	mockDeviceList        = []int32{0, 1, 2, 3, 4, 5, 6, 7}
	mock950ChipInfo       = &common.ChipInfo{
		Type:    "Ascend",
		Name:    "Ascend950PR",
		Version: "V1",
	}
	mock950BoardInfo = common.BoardInfo{
		BoardId: 0x1c,
	}
	mock950MainBoardId = uint32(0x01)
)

var testAutoInitWhenDcmiMap = map[string]interface{}{
	"auto init success":                              testAutoInitSuccessWithDcmi,
	"auto init failed, get card list failed":         testGetCardListFailedWithDcmi,
	"auto init failed, get chip info failed":         testGetChipInfoFailedWithDcmi,
	"auto init failed, get device board info failed": testGetDeviceBoardInfoFailedWithDcmi,
	"auto init failed, get main board info failed":   testGetMainBoardInfoFailedWithDcmi,
	"auto init success, get empty all product ype":   testGetDeviceLogicIDFailedWithDcmi,
}

var testAutoInitWhenDcmiV2Map = map[string]interface{}{
	"auto init success":                              testAutoInitSuccessWithDcmiV2,
	"auto init failed, get device list failed":       testGetDeviceListFailedWithDcmiV2,
	"auto init failed, get chip info failed":         testGetChipInfoFailedWithDcmiV2,
	"auto init failed, get device board info failed": testGetDeviceBoardInfoFailedWithDcmiV2,
	"auto init failed, get main board info failed":   testGetMainBoardInfoFailedWithDcmiV2,
}

func TestDeviceManagerGetDeviceAllErrorCodeWithTimeOut(t *testing.T) {
	testCases := []struct {
		name           string
		logicID        int32
		timeout        time.Duration
		mockCardID     int32
		mockDeviceID   int32
		mockGetIDErr   error
		mockErrCount   int32
		mockErrCodes   []int64
		mockGetCodeErr error
		expectedCount  int32
		expectedCodes  []int64
		expectedErr    error
	}{
		{
			name:    "should return error codes successfully when get card and device id success",
			logicID: testLogicID, timeout: timeout, mockCardID: testCardID,
			mockDeviceID: testDeviceID, mockGetIDErr: nil, mockErrCount: testErrCount,
			mockErrCodes: []int64{1, 2, 3}, mockGetCodeErr: nil, expectedCount: testErrCount,
			expectedCodes: []int64{1, 2, 3}, expectedErr: nil,
		},
		{
			name:    "should return error when get card and device id failed",
			logicID: testLogicID, timeout: timeout, mockCardID: 0, mockDeviceID: 0,
			mockGetIDErr: errors.New("get id failed"),
			mockErrCount: 0, mockErrCodes: nil, mockGetCodeErr: nil, expectedCount: testRetError,
			expectedCodes: nil, expectedErr: errors.New(errGetCardDevice),
		},
		{
			name:    "should return error when get error code failed",
			logicID: testLogicID, timeout: timeout, mockCardID: testCardID,
			mockDeviceID: testDeviceID, mockGetIDErr: nil, mockErrCount: 0,
			mockErrCodes: nil, mockGetCodeErr: errors.New("get code failed"),
			expectedCount: testRetError, expectedCodes: nil,
			expectedErr: errors.New(errGetErrorCode),
		},
	}

	doTestGetDeviceAllErrorCodeWithTimeOut(t, testCases)
}

func doTestGetDeviceAllErrorCodeWithTimeOut(t *testing.T, testCases []struct {
	name           string
	logicID        int32
	timeout        time.Duration
	mockCardID     int32
	mockDeviceID   int32
	mockGetIDErr   error
	mockErrCount   int32
	mockErrCodes   []int64
	mockGetCodeErr error
	expectedCount  int32
	expectedCodes  []int64
	expectedErr    error
}) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dm := &DeviceManager{
				DcMgr: &dcmi.DcManager{},
			}

			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyPrivateMethod(dm, "getCardIdAndDeviceId",
				func(*DeviceManager, int32) (int32, int32, error) {
					return tc.mockCardID, tc.mockDeviceID, tc.mockGetIDErr
				})

			patches.ApplyMethodReturn(
				dm.DcMgr,
				"DcGetDeviceAllErrorCodeWithTimeout",
				tc.mockErrCount,
				tc.mockErrCodes,
				tc.mockGetCodeErr,
			)

			count, codes, err := dm.GetDeviceAllErrorCodeWithTimeOut(tc.logicID, tc.timeout)

			convey.Convey("", t, func() {
				convey.So(count, convey.ShouldEqual, tc.expectedCount)
				convey.So(codes, convey.ShouldResemble, tc.expectedCodes)
				if tc.expectedErr != nil {
					convey.So(err, convey.ShouldNotBeNil)
				} else {
					convey.So(err, convey.ShouldBeNil)
				}
			})
		})
	}
}

// TestGetCardIdAndDeviceId test the getCardIdAndDeviceId function
func TestGetCardIdAndDeviceId(t *testing.T) {
	idCache.Clear()
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
	if err := hwlog.InitRunLogger(&config, context.Background()); err != nil {
		fmt.Printf("init log failed, %v\n", err)
		return
	}
}

const (
	testLogicID0   = int32(0)
	testCardID0    = int32(0)
	testDeviceID0  = int32(0)
	testAicUtil    = uint32(50)
	testAivUtil    = uint32(60)
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
					ApplyMethodReturn(dm.DcMgr, "DcGetDeviceUtilizationRateV2",
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
					ApplyMethodReturn(dm.DcMgr, "DcGetDeviceUtilizationRateV2",
						common.DcmiMultiUtilizationInfo{
							AicUtil:    testAicUtil,
							AivUtil:    testAivUtil,
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
					convey.So(result.AicUtil, convey.ShouldEqual, testAicUtil)
					convey.So(result.AivUtil, convey.ShouldEqual, testAivUtil)
					convey.So(result.AicoreUtil, convey.ShouldEqual, testAicoreUtil)
					convey.So(result.NpuUtil, convey.ShouldEqual, testNpuUtil)
				}
			})
		}
	})
}

// TestAutoInit test auto init
func TestAutoInit(t *testing.T) {
	convey.Convey("auto init when dcmi version is dcmi", t, testAutoInitWhenDcmi)
	convey.Convey("auto init when dcmi version is dcmiv2", t, testAutoInitWhenDcmiV2)
}

func testAutoInitWhenDcmi() {
	p := getDcmiFullPatches()
	defer p.Reset()
	for tip, testCase := range testAutoInitWhenDcmiMap {
		autoInitSetup()
		convey.Convey(tip, testCase)
	}
}

func autoInitSetup() {
	deviceCommonSetManagerList = []DeviceCommonSetInterface{
		&deviceCommonInitManager{
			DeviceManager: DeviceManager{
				DcMgr:          &dcmi.DcManager{},
				dcmiApiVersion: DcmiApiV1,
			},
		},
		&deviceCommonInitManagerV2{
			DeviceManagerV2: DeviceManagerV2{
				DcMgr:          &dcmi.DcV2Manager{},
				dcmiApiVersion: DcmiApiV2,
			},
		},
	}
}

func testAutoInitWhenDcmiV2() {
	p := getDcmiV2FullPatches()
	defer p.Reset()
	for tip, testCase := range testAutoInitWhenDcmiV2Map {
		autoInitSetup()
		convey.Convey(tip, testCase)
	}
}

func getDcmiFullPatches() *gomonkey.Patches {
	p := gomonkey.ApplyMethodReturn(&dcmi.DcManager{}, "DcInit", nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcShutDown", nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetCardList", mockCardNum, mockCardList, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetDeviceNumInCard", mockDeviceNumInCard, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetChipInfo", mock910ChipInfo, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetDcmiVersion", mockDcmiVersion, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetDeviceBoardInfo", mockBoardInfo, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetDeviceMainBoardInfo", mockMainBoardId, nil).
		ApplyMethod(reflect.TypeOf(&dcmi.DcManager{}), "DcGetCardIDDeviceID", func(_ *dcmi.DcManager, logicId int32) (int32, int32, error) {
			return logicId, 0, nil
		}).
		ApplyMethod(reflect.TypeOf(&dcmi.DcManager{}), "DcGetDeviceLogicID", func(_ *dcmi.DcManager, cardId, _ int32) (int32, error) {
			return cardId, nil
		}).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetProductType", mockProductType, nil)
	return p
}

func getDcmiV2FullPatches() *gomonkey.Patches {
	p := gomonkey.ApplyMethodReturn(&dcmi.DcV2Manager{}, "DcInit", nil).
		ApplyMethodReturn(&dcmi.DcV2Manager{}, "DcShutDown", nil).
		ApplyMethodReturn(&dcmi.DcV2Manager{}, "DcGetDeviceList", mockDeviceNum, mockDeviceList, nil).
		ApplyMethodReturn(&dcmi.DcV2Manager{}, "DcGetAllDeviceCount", mockDeviceNum, nil).
		ApplyMethodReturn(&dcmi.DcV2Manager{}, "DcGetChipInfo", mock950ChipInfo, nil).
		ApplyMethodReturn(&dcmi.DcV2Manager{}, "DcGetDcmiVersion", mockDcmiVersion, nil).
		ApplyMethodReturn(&dcmi.DcV2Manager{}, "DcGetDeviceBoardInfo", mock950BoardInfo, nil).
		ApplyMethodReturn(&dcmi.DcV2Manager{}, "DcGetDeviceMainBoardInfo", mock950MainBoardId, nil)
	return p
}

func testAutoInitSuccessWithDcmi() {
	devm, err := AutoInit("", api.DefaultDeviceResetTimeout)
	devM, ok := devm.(*DeviceManager)
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(err, convey.ShouldBeNil)
	convey.So(devM.DevType, convey.ShouldEqual, api.Ascend910)
	convey.So(devM.dcmiApiVersion, convey.ShouldEqual, DcmiApiV1)
	convey.So(devM.dcmiVersion, convey.ShouldEqual, mockDcmiVersion)
	convey.So(devM.isTrainingCard, convey.ShouldBeTrue)
	convey.So(devM.ProductTypes, convey.ShouldResemble, []string{mockProductType})
	convey.So(devM.DcMgr, convey.ShouldResemble, &A910Manager{})
}

func testAutoInitSuccessWithDcmiV2() {
	devm, err := AutoInit("", api.DefaultDeviceResetTimeout)
	devM, ok := devm.(*DeviceManagerV2)
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(err, convey.ShouldBeNil)
	convey.So(devM.DevType, convey.ShouldEqual, api.Ascend910A5)
	convey.So(devM.dcmiApiVersion, convey.ShouldEqual, DcmiApiV2)
	convey.So(devM.dcmiVersion, convey.ShouldEqual, mockDcmiVersion)
	convey.So(devM.isTrainingCard, convey.ShouldBeTrue)
	convey.So(devM.ProductTypes, convey.ShouldResemble, []string{})
	convey.So(devM.DcMgr, convey.ShouldResemble, &A950Manager{})
}

func testGetCardListFailedWithDcmi() {
	patch := gomonkey.ApplyMethodReturn(&dcmi.DcManager{}, "DcGetCardList", mockCardNum, mockCardList, mockErr)
	defer patch.Reset()
	expectErr := errors.New("auto init failed when get device info, err: get card list failed for init")
	errDevM, err := AutoInit("", api.DefaultDeviceResetTimeout)
	convey.So(err, convey.ShouldResemble, expectErr)
	convey.So(errDevM, convey.ShouldBeNil)
}

func testGetDeviceListFailedWithDcmiV2() {
	patch := gomonkey.ApplyMethodReturn(&dcmi.DcV2Manager{}, "DcGetDeviceList", mockDeviceNum, mockDeviceList, mockErr)
	defer patch.Reset()
	expectErr := errors.New("auto init failed when get device info, err: get device list failed for init")
	errDevM, err := AutoInit("", api.DefaultDeviceResetTimeout)
	convey.So(err, convey.ShouldResemble, expectErr)
	convey.So(errDevM, convey.ShouldBeNil)
}

func testGetChipInfoFailedWithDcmi() {
	patch := gomonkey.ApplyMethodReturn(&dcmi.DcManager{}, "DcGetChipInfo", mock910ChipInfo, mockErr)
	defer patch.Reset()

	expectErr := errors.New("auto init failed when get device info, err: cannot get valid chip info")
	errDevM, err := AutoInit("", api.DefaultDeviceResetTimeout)
	convey.So(err, convey.ShouldResemble, expectErr)
	convey.So(errDevM, convey.ShouldBeNil)
}

func testGetChipInfoFailedWithDcmiV2() {
	patch := gomonkey.ApplyMethodReturn(&dcmi.DcV2Manager{}, "DcGetChipInfo", mock950ChipInfo, mockErr)
	defer patch.Reset()

	expectErr := errors.New("auto init failed when get device info, err: cannot get valid chip info")
	errDevM, err := AutoInit("", api.DefaultDeviceResetTimeout)
	convey.So(err, convey.ShouldResemble, expectErr)
	convey.So(errDevM, convey.ShouldBeNil)
}

func testGetDeviceBoardInfoFailedWithDcmi() {
	patch := gomonkey.ApplyMethodReturn(&dcmi.DcManager{}, "DcGetDeviceBoardInfo", mockBoardInfo, mockErr)
	defer patch.Reset()
	expectErr := errors.New("auto init failed when get device info, err: cannot get valid board info")
	errDevM, err := AutoInit("", api.DefaultDeviceResetTimeout)
	convey.So(err, convey.ShouldResemble, expectErr)
	convey.So(errDevM, convey.ShouldBeNil)
}

func testGetDeviceBoardInfoFailedWithDcmiV2() {
	patch := gomonkey.ApplyMethodReturn(&dcmi.DcV2Manager{}, "DcGetDeviceBoardInfo", mock950BoardInfo, mockErr)
	defer patch.Reset()
	expectErr := errors.New("auto init failed when get device info, err: cannot get valid board info")
	errDevM, err := AutoInit("", api.DefaultDeviceResetTimeout)
	convey.So(err, convey.ShouldResemble, expectErr)
	convey.So(errDevM, convey.ShouldBeNil)
}

func testGetMainBoardInfoFailedWithDcmi() {
	patch := gomonkey.ApplyMethodReturn(&dcmi.DcManager{}, "DcGetDeviceMainBoardInfo", mockMainBoardId, mockErr)
	defer patch.Reset()
	DevM, err := AutoInit("", api.DefaultDeviceResetTimeout)
	convey.So(DevM, convey.ShouldNotBeNil)
	convey.So(DevM.GetMainBoardId(), convey.ShouldEqual, 0)
	convey.So(err, convey.ShouldBeNil)
}

func testGetMainBoardInfoFailedWithDcmiV2() {
	patch := gomonkey.ApplyMethodReturn(&dcmi.DcV2Manager{}, "DcGetDeviceMainBoardInfo", mockMainBoardId, mockErr)
	defer patch.Reset()
	DevM, err := AutoInit("", api.DefaultDeviceResetTimeout)
	convey.So(DevM, convey.ShouldNotBeNil)
	convey.So(DevM.GetMainBoardId(), convey.ShouldEqual, 0)
	convey.So(err, convey.ShouldBeNil)
}

func testGetDeviceLogicIDFailedWithDcmi() {
	patch := gomonkey.ApplyMethodReturn(&dcmi.DcManager{}, "DcGetDeviceLogicID", mockLogicID, mockErr)
	defer patch.Reset()
	DevM, err := AutoInit("", api.DefaultDeviceResetTimeout)
	convey.So(DevM, convey.ShouldNotBeNil)
	allProductType, _ := DevM.GetAllProductType()
	convey.So(allProductType, convey.ShouldResemble, []string{})
	convey.So(err, convey.ShouldBeNil)
}

// TestDeviceManagerGetDeviceHealthErrorPath test GetDeviceHealth
func TestDeviceManagerGetDeviceHealthErrorPath(t *testing.T) {
	manager := &DeviceManager{DcMgr: &dcmi.DcManager{}}
	patch := gomonkey.ApplyMethodReturn(manager.DcMgr, "DcGetCardIDDeviceID",
		int32(-1), int32(-1), errors.New("mock err"))
	defer patch.Reset()
	_, err := manager.GetDeviceHealth(testLogicID)
	convey.Convey("GetDeviceHealth returns error if getCardIdAndDeviceId fails", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestDeviceManagerGetDeviceHealthErrorPath test GetDeviceVoltage  device voltage error path
func TestDeviceManagerGetDeviceVoltageErrorPath(t *testing.T) {
	manager := &DeviceManager{DcMgr: &dcmi.DcManager{}}
	patch := gomonkey.ApplyMethodReturn(manager.DcMgr, "DcGetCardIDDeviceID",
		int32(-1), int32(-1), errors.New("mock err"))
	defer patch.Reset()
	_, err := manager.GetDeviceVoltage(testLogicID)
	convey.Convey("GetDeviceVoltage returns error if getCardIdAndDeviceId fails", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestDeviceManagerCreateVirtualDeviceInvalidTemplateName test CreateVirtualDevice invalid template name
func TestDeviceManagerCreateVirtualDeviceInvalidTemplateName(t *testing.T) {
	manager := &DeviceManager{DevType: "Ascend910", DcMgr: &dcmi.DcManager{}}
	vDevInfo := common.CgoCreateVDevRes{TemplateName: "invalid"}
	convey.Convey("CreateVirtualDevice returns error for invalid template name", t, func() {
		_, err := manager.CreateVirtualDevice(1, vDevInfo)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestDeviceManagerSetFaultEventCallFuncNilFunc test set fault event func is nil
func TestDeviceManagerSetFaultEventCallFuncNilFunc(t *testing.T) {
	manager := &DeviceManager{DcMgr: &dcmi.DcManager{}}
	convey.Convey("SetFaultEventCallFunc returns error if func is nil", t, func() {
		err := manager.SetFaultEventCallFunc(nil)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestDeviceManagerGetNpuWorkModeAMPMode test get npu work mode
func TestDeviceManagerGetNpuWorkModeAMPMode(t *testing.T) {
	manager := &DeviceManager{DevType: "Ascend910B", DcMgr: &dcmi.DcManager{}}
	convey.Convey("GetNpuWorkMode returns AMPMode for Ascend910B", t, func() {
		mode := manager.GetNpuWorkMode()
		convey.So(mode, convey.ShouldEqual, common.AMPMode)
	})
}

// TestDeviceManagerIsTrainingCard should return true
func TestDeviceManagerIsTrainingCard(t *testing.T) {
	manager := &DeviceManager{isTrainingCard: true}
	convey.Convey("IsTrainingCard returns true", t, func() {
		convey.So(manager.IsTrainingCard(), convey.ShouldBeTrue)
	})
}

// TestDeviceManagerGetProductTypeArray test GetProductTypeArray return array
func TestDeviceManagerGetProductTypeArray(t *testing.T) {
	manager := &DeviceManager{ProductTypes: []string{"A", "B"}}
	convey.Convey("GetProductTypeArray returns correct product types", t, func() {
		types := manager.GetProductTypeArray()
		convey.So(types, convey.ShouldResemble, []string{"A", "B"})
	})
}

// TestDeviceManagerGetDevType get dev type should return Ascend910
func TestDeviceManagerGetDevType(t *testing.T) {
	manager := &DeviceManager{DevType: "Ascend910"}
	convey.Convey("GetDevType returns correct type", t, func() {
		convey.So(manager.GetDevType(), convey.ShouldEqual, "Ascend910")
	})
}
