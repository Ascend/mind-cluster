/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
+   Licensed under the Apache License, Version 2.0 (the "License");
+   you may not use this file except in compliance with the License.
+   You may obtain a copy of the License at
+
+   http://www.apache.org/licenses/LICENSE-2.0
+
+   Unless required by applicable law or agreed to in writing, software
+   distributed under the License is distributed on an "AS IS" BASIS,
+   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
+   See the License for the specific language governing permissions and
+   limitations under the License.
+*/

// Package server holds the implementation of registration to kubelet, k8s pod resource interface.
package server

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"ascend-common/devmanager"
	npuCommon "ascend-common/devmanager/common"
)

// TestGetCardType for test getCardType
func TestGetCardType(t *testing.T) {
	hdm := &HwDevManager{
		manager: device.NewHwAscend910Manager(),
		allInfo: common.NpuAllInfo{
			AllDevs: []common.NpuDevice{{LogicID: 0}},
		},
	}
	mockGetDmgr := gomonkey.ApplyMethod(reflect.TypeOf(new(device.HwAscend910Manager)), "GetDmgr",
		func(_ *device.HwAscend910Manager) devmanager.DeviceInterface { return &devmanager.DeviceManagerMock{} })
	defer mockGetDmgr.Reset()
	convey.Convey("test getCardType when get board info error", t, func() {
		mockGetBoardInfo := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
			"GetBoardInfo", func(_ *devmanager.DeviceManagerMock, _ int32) (npuCommon.BoardInfo, error) {
				return npuCommon.BoardInfo{}, fmt.Errorf("get board info error")
			})
		defer mockGetBoardInfo.Reset()
		cardType, _ := hdm.getCardType()
		convey.So(cardType, convey.ShouldBeEmpty)
	})
	convey.Convey("test getCardType success", t, func() {
		mockGetMainBoardId := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
			"GetMainBoardId", func(_ *devmanager.DeviceManagerMock) uint32 {
				return common.A5300IMainBoardId
			})
		defer mockGetMainBoardId.Reset()
		mockGetBoardInfo := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
			"GetBoardInfo", func(_ *devmanager.DeviceManagerMock, _ int32) (npuCommon.BoardInfo, error) {
				return npuCommon.BoardInfo{BoardId: npuCommon.A5300IBoardId}, nil
			})
		defer mockGetBoardInfo.Reset()
		cardType, _ := hdm.getCardType()
		convey.So(cardType, convey.ShouldEqual, common.A5300ICardName)
	})
	convey.Convey("test getCardType failed", t, func() {
		mockGetBoardInfo := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
			"GetBoardInfo", func(_ *devmanager.DeviceManagerMock, _ int32) (npuCommon.BoardInfo, error) {
				return npuCommon.BoardInfo{BoardId: common.A300IA2BoardId}, nil
			})
		defer mockGetBoardInfo.Reset()
		cardType, _ := hdm.getCardType()
		convey.So(cardType, convey.ShouldBeEmpty)
	})
}
