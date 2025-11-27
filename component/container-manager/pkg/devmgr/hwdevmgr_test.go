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
	"sort"
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

func TestGetPhyIdOnRing(t *testing.T) {
	resetDevMgr()

	testAtlas300IDuoScenarios(t)
	testDevicesWithoutRingScenario(t)
	testAscend910A3Scenario(t)
	testAscend910RingScenario(t)
	testGetDevNumPerRingFailScenario(t)
}

func testAtlas300IDuoScenarios(t *testing.T) {
	convey.Convey("test method 'GetPhyIdOnRing' for Atlas 300I Duo", t, func() {
		testAtlas300IDuoSuccess(t)
		testAtlas300IDuoGetCardIDDeviceIDFail(t)
	})
}

func testAtlas300IDuoSuccess(t *testing.T) {
	convey.Convey("should return coupled phy ids when product type is Atlas 300I Duo", func() {
		patches := gomonkey.ApplyMethod(&devmanager.DeviceManagerMock{}, "GetCardIDDeviceID",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (int32, int32, error) {
				return 0, 0, nil
			})
		defer patches.Reset()

		patches.ApplyPrivateMethod(&HwDevMgr{}, "isAtlas300IDuo",
			func(_ *HwDevMgr, cardId, deviceId int32) bool {
				return true
			})

		patches.ApplyPrivateMethod(&HwDevMgr{}, "getCoupledPhyIdsFrom310pDuo",
			func(_ *HwDevMgr, phyId int32) ([]int32, error) {
				return []int32{0, 1}, nil
			})

		result, err := mockDevMgr.GetPhyIdOnRing(0)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldResemble, []int32{0, 1})
	})
}

func testAtlas300IDuoGetCardIDDeviceIDFail(t *testing.T) {
	convey.Convey("should return error when GetCardIDDeviceID fails", func() {
		patches := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetCardIDDeviceID",
			int32(0), int32(0), testErr)
		defer patches.Reset()

		result, err := mockDevMgr.GetPhyIdOnRing(0)
		convey.So(err, convey.ShouldResemble, testErr)
		convey.So(result, convey.ShouldBeNil)
	})
}

func testDevicesWithoutRingScenario(t *testing.T) {
	convey.Convey("test method 'GetPhyIdOnRing' for devices without ring", t, func() {
		patches := gomonkey.ApplyMethod(&devmanager.DeviceManagerMock{}, "GetCardIDDeviceID",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (int32, int32, error) {
				return 0, 0, nil
			})
		defer patches.Reset()

		patches.ApplyPrivateMethod(&HwDevMgr{}, "isAtlas300IDuo",
			func(_ *HwDevMgr, cardId, deviceId int32) bool {
				return false
			})

		patches.ApplyMethodReturn(&HwDevMgr{}, "GetDevNumPerRing", common.NoRingNum, nil)

		result, err := mockDevMgr.GetPhyIdOnRing(0)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldResemble, []int32{0})
	})
}

func testAscend910A3Scenario(t *testing.T) {
	convey.Convey("test method 'GetPhyIdOnRing' for Ascend910A3", t, func() {
		patches := gomonkey.ApplyMethod(&devmanager.DeviceManagerMock{}, "GetCardIDDeviceID",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (int32, int32, error) {
				return 0, 0, nil
			})
		defer patches.Reset()

		patches.ApplyPrivateMethod(&HwDevMgr{}, "isAtlas300IDuo",
			func(_ *HwDevMgr, cardId, deviceId int32) bool {
				return false
			})

		patches.ApplyMethodReturn(&HwDevMgr{}, "GetDevNumPerRing", common.Ascend910RingsNum, nil)
		patches.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDevType", api.Ascend910A3)
		patches.ApplyPrivateMethod(&HwDevMgr{}, "getPhyIdOn910A3Ring",
			func(phyId, cardId, deviceId int32) ([]int32, error) {
				return []int32{0, 1, 2, 3}, nil
			})

		result, err := mockDevMgr.GetPhyIdOnRing(0)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldResemble, []int32{0, 1, 2, 3})
	})
}

func testAscend910RingScenario(t *testing.T) {
	convey.Convey("test method 'GetPhyIdOnRing' for Ascend910 ring", t, func() {
		patches := gomonkey.ApplyMethod(&devmanager.DeviceManagerMock{}, "GetCardIDDeviceID",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (int32, int32, error) {
				return 0, 0, nil
			})
		defer patches.Reset()

		patches.ApplyPrivateMethod(&HwDevMgr{}, "isAtlas300IDuo",
			func(_ *HwDevMgr, cardId, deviceId int32) bool {
				return false
			})

		patches.ApplyMethodReturn(&HwDevMgr{}, "GetDevNumPerRing", common.Ascend910BRingsNumTrain, nil)
		patches.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDevType", api.Ascend910)
		patches.ApplyPrivateMethod(&HwDevMgr{}, "getPhyIdOn910Ring",
			func(phyId, cardId, deviceId int32) ([]int32, error) {
				return []int32{0, 1, 2, 3, 4, 5, 6, 7}, nil
			})

		result, err := mockDevMgr.GetPhyIdOnRing(0)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldResemble, []int32{0, 1, 2, 3, 4, 5, 6, 7})
	})
}

func testGetDevNumPerRingFailScenario(t *testing.T) {
	convey.Convey("test method 'GetPhyIdOnRing' should return error when GetDevNumPerRing fails", t, func() {
		patches := gomonkey.ApplyMethod(&devmanager.DeviceManagerMock{}, "GetCardIDDeviceID",
			func(_ *devmanager.DeviceManagerMock, logicID int32) (int32, int32, error) {
				return 0, 0, nil
			})
		defer patches.Reset()

		patches.ApplyPrivateMethod(&HwDevMgr{}, "isAtlas300IDuo",
			func(_ *HwDevMgr, cardId, deviceId int32) bool {
				return false
			})

		patches.ApplyMethodReturn(&HwDevMgr{}, "GetDevNumPerRing", 0, testErr)

		result, err := mockDevMgr.GetPhyIdOnRing(0)
		convey.So(err, convey.ShouldResemble, testErr)
		convey.So(result, convey.ShouldBeNil)
	})
}

func TestIsAtlas300IDuo(t *testing.T) {
	resetDevMgr()

	convey.Convey("test method 'isAtlas300IDuo'", t, func() {
		convey.Convey("should return true when product type is Atlas 300I Duo", func() {
			patches := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetProductType",
				common.ProductTypeAtlas300IDuo, nil)
			defer patches.Reset()

			result := mockDevMgr.isAtlas300IDuo(0, 0)
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("should return false when product type is not Atlas 300I Duo", func() {
			patches := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetProductType",
				"OtherProduct", nil)
			defer patches.Reset()

			result := mockDevMgr.isAtlas300IDuo(0, 0)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("should return false when GetProductType fails", func() {
			patches := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetProductType",
				"", testErr)
			defer patches.Reset()

			result := mockDevMgr.isAtlas300IDuo(0, 0)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestGetCoupledPhyIdsFrom310pDuo(t *testing.T) {
	resetDevMgr()

	convey.Convey("test method 'getCoupledPhyIdsFrom310pDuo'", t, func() {
		convey.Convey("should return coupled phy ids successfully", func() {
			mockDevMgr.npuInfos = map[int32]*common.NPUInfo{
				0: {PhyID: 0, LogicID: 0, DeviceID: 0},
				1: {PhyID: 1, LogicID: 1, DeviceID: 0}, // Same DeviceID as phyId 0
				2: {PhyID: 2, LogicID: 2, DeviceID: 1},
			}

			result, err := mockDevMgr.getCoupledPhyIdsFrom310pDuo(0)
			var sortSlice []int
			for _, val := range result {
				sortSlice = append(sortSlice, int(val))
			}
			sort.Ints(sortSlice)
			convey.So(err, convey.ShouldBeNil)
			convey.So(sortSlice, convey.ShouldResemble, []int{0, 1})
		})

		convey.Convey("should return error when npuInfos is nil", func() {
			mockDevMgr.npuInfos = nil

			result, err := mockDevMgr.getCoupledPhyIdsFrom310pDuo(0)
			convey.So(err, convey.ShouldResemble, errors.New("npuInfos is nil"))
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("should return error when phyId not found in npuInfos", func() {
			mockDevMgr.npuInfos = map[int32]*common.NPUInfo{
				1: {PhyID: 1, LogicID: 1, DeviceID: 0},
			}

			result, err := mockDevMgr.getCoupledPhyIdsFrom310pDuo(0)
			convey.So(err, convey.ShouldResemble, errors.New("npuInfos is nil"))
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

func TestGetPhyIdOn910A3Ring(t *testing.T) {
	resetDevMgr()

	convey.Convey("test method 'getPhyIdOn910A3Ring'", t, func() {
		testSuccessScenario(t)
		testGetBrotherCardIDFail(t)
		testGetDeviceLogicIDFailForDevice0(t)
		testGetDeviceLogicIDFailForDevice1(t)
		testGetDeviceLogicIDFailForOtherDevice(t)
	})
}

func testSuccessScenario(t *testing.T) {
	convey.Convey("should return phy ids on 910A3 ring successfully", func() {
		patches := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetBrotherCardID",
			int32(1), nil)
		defer patches.Reset()

		patches.ApplyMethodSeq(&devmanager.DeviceManagerMock{}, "GetDeviceLogicID",
			[]gomonkey.OutputCell{
				{Values: gomonkey.Params{int32(10), nil}}, // First call for device 0
				{Values: gomonkey.Params{int32(11), nil}}, // Second call for device 1
				{Values: gomonkey.Params{int32(12), nil}}, // Third call for other device
			})

		patches.ApplyMethod(&HwDevMgr{}, "GetPhyIdByLogicId",
			func(_ *HwDevMgr, logicId int32) int32 {
				return logicId // Simple mapping for testing
			})

		result, err := mockDevMgr.getPhyIdOn910A3Ring(0, 0, 0)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldResemble, []int32{0, 12, 10, 11})
	})
}

func testGetBrotherCardIDFail(t *testing.T) {
	convey.Convey("should return error when GetBrotherCardID fails", func() {
		patches := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetBrotherCardID",
			int32(0), testErr)
		defer patches.Reset()

		result, err := mockDevMgr.getPhyIdOn910A3Ring(0, 0, 0)
		convey.So(err, convey.ShouldResemble, testErr)
		convey.So(result, convey.ShouldBeNil)
	})
}

func testGetDeviceLogicIDFailForDevice0(t *testing.T) {
	convey.Convey("should return error when GetDeviceLogicID for device 0 fails", func() {
		patches := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetBrotherCardID",
			int32(1), nil)
		defer patches.Reset()

		patches.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDeviceLogicID",
			int32(0), testErr)

		result, err := mockDevMgr.getPhyIdOn910A3Ring(0, 0, 0)
		convey.So(err, convey.ShouldResemble, testErr)
		convey.So(result, convey.ShouldBeNil)
	})
}

func testGetDeviceLogicIDFailForDevice1(t *testing.T) {
	convey.Convey("should return error when GetDeviceLogicID for device 1 fails", func() {
		patches := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetBrotherCardID",
			int32(1), nil)
		defer patches.Reset()

		patches.ApplyMethodSeq(&devmanager.DeviceManagerMock{}, "GetDeviceLogicID",
			[]gomonkey.OutputCell{
				{Values: gomonkey.Params{int32(10), nil}},    // First call succeeds
				{Values: gomonkey.Params{int32(0), testErr}}, // Second call fails
			})

		result, err := mockDevMgr.getPhyIdOn910A3Ring(0, 0, 0)
		convey.So(err, convey.ShouldResemble, testErr)
		convey.So(result, convey.ShouldBeNil)
	})
}

func testGetDeviceLogicIDFailForOtherDevice(t *testing.T) {
	convey.Convey("should return error when GetDeviceLogicID for other device fails", func() {
		patches := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetBrotherCardID",
			int32(1), nil)
		defer patches.Reset()

		patches.ApplyMethodSeq(&devmanager.DeviceManagerMock{}, "GetDeviceLogicID",
			[]gomonkey.OutputCell{
				{Values: gomonkey.Params{int32(10), nil}},
				{Values: gomonkey.Params{int32(11), nil}},
				{Values: gomonkey.Params{int32(0), testErr}}, // Third call fails
			})

		result, err := mockDevMgr.getPhyIdOn910A3Ring(0, 0, 0)
		convey.So(err, convey.ShouldResemble, testErr)
		convey.So(result, convey.ShouldBeNil)
	})
}
