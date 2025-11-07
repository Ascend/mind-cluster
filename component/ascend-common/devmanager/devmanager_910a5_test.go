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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/dcmi"
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
