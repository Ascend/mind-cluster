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

// Package devmgr test for hwDevMgr
package devmgr

import (
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/gogo/protobuf/sortkeys"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/devmanager"
	ascommon "ascend-common/devmanager/common"
	"container-manager/pkg/common"
)

func TestHwDevMgrBasicMethod(t *testing.T) {
	resetDevMgr()
	convey.Convey("test method 'GetDevType' success", t, func() {
		devType := mockDevMgr.GetDevType()
		convey.So(devType, convey.ShouldEqual, mockDevMgr.devType)
	})
	convey.Convey("test method 'GetDevUsage' success", t, func() {
		devUsage := mockDevMgr.GetDevUsage()
		convey.So(devUsage, convey.ShouldEqual, mockDevMgr.devUsage)
	})
	convey.Convey("test method 'GetDevNum' success", t, func() {
		devNum := mockDevMgr.GetDevNum()
		convey.So(devNum, convey.ShouldEqual, len(mockDevMgr.npuInfos))
	})
	convey.Convey("test method 'GetDevNum' success", t, func() {
		phyIds := mockDevMgr.GetPhyIds()
		sortkeys.Int32s(phyIds)
		expPhyIds := []int32{dev0, dev1, dev2, dev3, dev4, dev5, dev6, dev7}
		sortkeys.Int32s(expPhyIds)
		convey.So(phyIds, convey.ShouldResemble, expPhyIds)
	})
}

func TestSetDeviceUsage(t *testing.T) {
	resetDevMgr()
	convey.Convey("test method 'setDeviceUsage' success, devType is 310", t, func() {
		mockDevMgr.devType = api.Ascend310
		err := mockDevMgr.setDeviceUsage(0)
		convey.So(err, convey.ShouldBeNil)
		convey.So(mockDevMgr.devUsage, convey.ShouldEqual, common.Infer)
	})
	convey.Convey("test method 'setDeviceUsage' success, usage is infer", t, func() {
		mockDevMgr.devType = api.Ascend910B
		mockDevMgr.boardId = common.A800IA2NoneHccsBoardIdOld
		err := mockDevMgr.setDeviceUsage(0)
		convey.So(err, convey.ShouldBeNil)
		convey.So(mockDevMgr.devUsage, convey.ShouldEqual, common.Infer)

		mockDevMgr.devType = api.Ascend910
		err = mockDevMgr.setDeviceUsage(0)
		convey.So(err, convey.ShouldBeNil)
		convey.So(mockDevMgr.devUsage, convey.ShouldEqual, common.Train)
	})
	convey.Convey("test method 'setDeviceUsage' failed, get board id error", t, func() {
		var patches = gomonkey.ApplyMethodReturn(&HwDevMgr{}, "GetBoardId", uint32(0), testErr)
		defer patches.Reset()
		mockDevMgr.devType = api.Ascend910
		err := mockDevMgr.setDeviceUsage(0)
		expErr := fmt.Errorf("get board id failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestSetRingInfo(t *testing.T) {
	resetDevMgr()
	convey.Convey("test method 'setRingInfo' success", t, func() {
		err := mockDevMgr.setRingInfo()
		convey.So(err, convey.ShouldBeNil)
		convey.So(mockDevMgr.npuInfos[dev0].DevsOnRing, convey.ShouldResemble,
			[]int32{dev0, dev1, dev2, dev3, dev4, dev5, dev6, dev7})
	})
	convey.Convey("test method 'setRingInfo' failed, GetBoardId error", t, func() {
		var patches = gomonkey.ApplyMethodReturn(&HwDevMgr{}, "GetBoardId", uint32(0), testErr)
		defer patches.Reset()
		err := mockDevMgr.setRingInfo()
		expErr := errors.New("get board id failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestSetBoardId(t *testing.T) {
	resetDevMgr()
	convey.Convey("test method 'setBoardId' success", t, func() {
		var patches = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetBoardInfo",
			ascommon.BoardInfo{}, nil)
		defer patches.Reset()
		err := mockDevMgr.setBoardId(0)
		convey.So(err, convey.ShouldBeNil)
		convey.So(mockDevMgr.boardId, convey.ShouldEqual, int32(0))
	})
	convey.Convey("test method 'setBoardId' failed, GetBoardInfo error", t, func() {
		var patches = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetBoardInfo",
			ascommon.BoardInfo{}, testErr)
		defer patches.Reset()
		err := mockDevMgr.setBoardId(0)
		expErr := errors.New("get board id failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestGetBoardId(t *testing.T) {
	resetDevMgr()
	convey.Convey("test method 'GetBoardId' success", t, func() {
		boardId, err := mockDevMgr.GetBoardId(0)
		convey.So(err, convey.ShouldBeNil)
		convey.So(boardId, convey.ShouldEqual, mockDevMgr.boardId)
	})
	convey.Convey("test method 'GetBoardId' failed, GetBoardInfo error", t, func() {
		mockDevMgr.boardId = common.EmptyBoardId
		var patches = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetBoardInfo",
			ascommon.BoardInfo{}, testErr)
		defer patches.Reset()
		boardId, err := mockDevMgr.GetBoardId(0)
		expErr := fmt.Errorf("get board info failed, error: %v", testErr)
		convey.So(err, convey.ShouldResemble, expErr)
		convey.So(boardId, convey.ShouldEqual, common.EmptyBoardId)
	})
	convey.Convey("test method 'GetBoardId' success, GetBoardInfo success", t, func() {
		mockDevMgr.boardId = common.EmptyBoardId
		var patches = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetBoardInfo",
			ascommon.BoardInfo{BoardId: 0}, nil)
		defer patches.Reset()
		boardId, err := mockDevMgr.GetBoardId(0)
		convey.So(err, convey.ShouldBeNil)
		convey.So(boardId, convey.ShouldEqual, 0)
	})
}

func TestGetNodeNPUInfo(t *testing.T) {
	resetDevMgr()
	convey.Convey("test method 'GetNodeNPUInfo' success", t, func() {
		var patches = gomonkey.ApplyMethod(&devmanager.DeviceManagerMock{}, "GetPhysicIDFromLogicID",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (int32, error) {
				return logicID, nil
			}).ApplyMethod(&devmanager.DeviceManagerMock{}, "GetCardIDDeviceID",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (int32, int32, error) {
				return logicID, logicID, nil
			}).ApplyMethod(&devmanager.DeviceManagerMock{}, "GetChipInfo",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (*ascommon.ChipInfo, error) {
				return &ascommon.ChipInfo{
					Name: api.Ascend910,
				}, nil
			}).ApplyMethod(&devmanager.DeviceManagerMock{}, "GetDeviceIPAddress",
			func(_ *devmanager.DeviceManagerMock, logicID int32, ipType int32) (string, error) {
				return defaultDeviceIP, nil
			})
		defer patches.Reset()
		logicIds := []int32{dev0, dev1, dev2, dev3, dev4, dev5, dev6, dev7}
		npuInfo, err := mockDevMgr.setNodeNPUInfo(logicIds, len8)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(npuInfo), convey.ShouldEqual, len8)
	})
}

func TestSubscribeFaultEvent(t *testing.T) {
	convey.Convey("test method 'SubscribeFaultEvent' success", t, func() {
		var patches = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "SetFaultEventCallFunc", nil)
		defer patches.Reset()
		callback := func(devFaultInfo ascommon.DevFaultInfo) {}
		err := mockDevMgr.SubscribeFaultEvent(callback)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGetFaultCodesMap(t *testing.T) {
	resetDevMgr()
	convey.Convey("test method 'GetFaultCodesMap' success", t, func() {
		var patches = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDeviceAllErrorCode",
			int32(0), []int64{123456789, 987654321}, nil)
		defer patches.Reset()
		codesMap := mockDevMgr.GetFaultCodesMap()
		convey.So(len(codesMap), convey.ShouldEqual, len8)
		convey.So(codesMap[0], convey.ShouldResemble, []int64{123456789, 987654321})
	})
	convey.Convey("test method 'GetFaultCodesMap' failed, GetDeviceAllErrorCode error", t, func() {
		var patches = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDeviceAllErrorCode",
			int32(0), []int64{0}, testErr)
		defer patches.Reset()
		codesMap := mockDevMgr.GetFaultCodesMap()
		convey.So(len(codesMap), convey.ShouldEqual, 0)
	})
}
