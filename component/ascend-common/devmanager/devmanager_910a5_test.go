/*
 *    Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

// package devmanager for test auto init
package devmanager

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/dcmi"
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
	mockChipInfo              = &common.ChipInfo{
		Type:    "Ascend",
		Name:    "Ascend910 7591",
		Version: "V1",
	}
	mockBoardInfo = common.BoardInfo{
		BoardId: common.A900A5SuperPodBin1BoardId,
	}
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(&hwLogConfig, context.Background()); err != nil {
		fmt.Printf("init log failed, %v\n", err)
		return
	}
}

// TestAutoInit test auto init
func TestAutoInit(t *testing.T) {
	p := gomonkey.ApplyMethodReturn(&dcmi.DcManager{}, "DcInit", nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetDcmiVersion", mockDcmiVersion, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetCardList", mockCardNum, mockCardList, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetDeviceNumInCard", mockDeviceNumInCard, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetChipInfo", mockChipInfo, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetDeviceBoardInfo", mockBoardInfo, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetProductType", mockProductType, nil)
	defer p.Reset()

	convey.Convey("auto init success", t, testAutoInitSuccess)
	convey.Convey("auto init failed, get card list failed", t, testGetCardListFailed)
	convey.Convey("auto init failed, get chip info failed", t, testGetChipInfoFailed)
	convey.Convey("auto init failed, get device board info failed", t, testDeviceBoardInfoFailed)
}

func testAutoInitSuccess() {
	devM, err := AutoInit("", api.DefaultDeviceResetTimeout)
	convey.So(err, convey.ShouldBeNil)
	convey.So(devM.DevType, convey.ShouldEqual, api.Ascend910A5)
	convey.So(devM.dcmiVersion, convey.ShouldEqual, mockDcmiVersion)
	convey.So(devM.isTrainingCard, convey.ShouldBeTrue)
	convey.So(devM.ProductTypes, convey.ShouldResemble, []string{mockProductType})
	convey.So(devM.DcMgr, convey.ShouldResemble, &A910Manager{})
}

func testGetCardListFailed() {
	patch := gomonkey.ApplyMethodReturn(&dcmi.DcManager{}, "DcGetCardList", mockCardNum, mockCardList, mockErr)
	defer patch.Reset()

	expectErr := errors.New("auto init failed, err: get card list failed for init")
	errDevM, err := AutoInit("", api.DefaultDeviceResetTimeout)
	convey.So(err, convey.ShouldResemble, expectErr)
	convey.So(errDevM, convey.ShouldBeNil)
}

func testGetChipInfoFailed() {
	patch := gomonkey.ApplyMethodReturn(&dcmi.DcManager{}, "DcGetChipInfo", mockChipInfo, mockErr)
	defer patch.Reset()

	expectErr := errors.New("auto init failed, err: cannot get valid chip info")
	errDevM, err := AutoInit("", api.DefaultDeviceResetTimeout)
	convey.So(err, convey.ShouldResemble, expectErr)
	convey.So(errDevM, convey.ShouldBeNil)
}

func testDeviceBoardInfoFailed() {
	patch := gomonkey.ApplyMethodReturn(&dcmi.DcManager{}, "DcGetDeviceBoardInfo", mockBoardInfo, mockErr)
	defer patch.Reset()

	expectErr := errors.New("auto init failed, err: cannot get valid board info")
	errDevM, err := AutoInit("", api.DefaultDeviceResetTimeout)
	convey.So(err, convey.ShouldResemble, expectErr)
	convey.So(errDevM, convey.ShouldBeNil)
}

// TestDeviceManagerInitError test device manager init error (GetDeviceManager)
func TestDeviceManagerInitError(t *testing.T) {
	patch := gomonkey.ApplyMethodReturn(&dcmi.DcManager{}, "DcInit", errors.New("init error"))
	defer patch.Reset()
	devManagerOnce = sync.Once{} // reset singleton
	devManager = nil
	manager, err := GetDeviceManager(testTime)
	convey.Convey("GetDeviceManager returns error when DcInit fails", t, func() {
		convey.So(manager, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
	})
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
