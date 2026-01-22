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

// Package devmgr test for hwDevMgr workflow
package devmgr

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/devmanager"
	ascommon "ascend-common/devmanager/common"
	"container-manager/pkg/common"
)

func TestHwDevMgr(t *testing.T) {
	convey.Convey("test method 'Name' success", t, testMethodName)
	convey.Convey("test method 'Init' success", t, testMethodInit)
	convey.Convey("test method 'Work' success", t, testMethodWork)
	convey.Convey("test method 'ShutDown' success", t, testMethodShutDown)
	convey.Convey("test method 'initDmgr' success", t, testMethodInitDmgr)
	convey.Convey("test method 'initInfoRelatedDev'", t, testInitInfoRelatedDev)
	convey.Convey("test method 'initInfoRelatedNode'", t, testInitInfoRelatedNode)
}

func testMethodName() {
	convey.So(mockDevMgr.Name(), convey.ShouldEqual, "hwDev manager")
}

func testMethodInit() {
	convey.So(mockDevMgr.Init(), convey.ShouldBeNil)
}

func testMethodWork() {
	var hasExecuted bool
	var p1 = gomonkey.ApplyMethod(&HwDevMgr{}, "Work",
		func(hdm *HwDevMgr, ctx context.Context) {
			hasExecuted = true
			return
		})
	defer p1.Reset()
	mockDevMgr.Work(context.Background())
	convey.So(hasExecuted, convey.ShouldBeTrue)
}

func testMethodShutDown() {
	var hasExecuted bool
	var p2 = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "ShutDown", nil).
		ApplyMethod(&HwDevMgr{}, "ShutDown", func(hdm *HwDevMgr) {
			hasExecuted = true
			return
		})
	defer p2.Reset()
	mockDevMgr.ShutDown()
	convey.So(hasExecuted, convey.ShouldBeTrue)
}

func testMethodInitDmgr() {
	var p1 = gomonkey.ApplyFuncReturn(devmanager.AutoInit, nil, nil)
	err := mockDevMgr.initDmgr()
	convey.So(err, convey.ShouldBeNil)
	p1.Reset()

	var p2 = gomonkey.ApplyFuncReturn(devmanager.AutoInit, nil, testErr)
	err = mockDevMgr.initDmgr()
	expErr := errors.New("init devmanager failed")
	convey.So(err, convey.ShouldResemble, expErr)
	p2.Reset()
}

func testInitInfoRelatedDev() {
	convey.Convey("test method 'initInfoRelatedDev' success", func() {
		resetDevMgr()
		var mockNPUInfo = map[int32]*common.NPUInfo{0: {PhyID: 0, LogicID: 0}}
		var patches = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetDeviceList", int32(1), []int32{0}, nil).
			ApplyPrivateMethod(&HwDevMgr{}, "setNodeNPUInfo",
				func(logicIds []int32, devNum int32) (map[int32]*common.NPUInfo, error) { return mockNPUInfo, nil })
		defer patches.Reset()
		err := mockDevMgr.initInfoRelatedDev()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test method 'initInfoRelatedDev' failed, GetDeviceList error", func() {
		var patches = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDeviceList",
			int32(len8), []int32{dev0, dev1, dev2, dev3, dev4, dev5, dev6, dev7}, testErr)
		defer patches.Reset()
		err := mockDevMgr.initInfoRelatedDev()
		convey.So(err, convey.ShouldResemble, testErr)
	})
	convey.Convey("test method 'initInfoRelatedDev' failed, GetNodeNPUInfo error", func() {
		var patches = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetDeviceList", int32(1), []int32{0}, nil).
			ApplyPrivateMethod(&HwDevMgr{}, "setNodeNPUInfo",
				func(logicIds []int32, devNum int32) (map[int32]*common.NPUInfo, error) { return nil, testErr })
		defer patches.Reset()
		err := mockDevMgr.initInfoRelatedDev()
		convey.So(err, convey.ShouldResemble, testErr)
	})
	convey.Convey("test method 'initInfoRelatedDev' failed, invalid npu info", func() {
		var patches = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetDeviceList", int32(1), []int32{0}, nil).
			ApplyPrivateMethod(&HwDevMgr{}, "setNodeNPUInfo",
				func(logicIds []int32, devNum int32) (map[int32]*common.NPUInfo, error) { return nil, nil })
		defer patches.Reset()
		err := mockDevMgr.initInfoRelatedDev()
		expErr := errors.New("npu info is nil")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func testInitInfoRelatedNode() {
	convey.Convey("test method 'initInfoRelatedNode' success", func() {
		resetDevMgr()
		var patches = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDevType", api.Ascend910).
			ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetNpuWorkMode", ascommon.AMPMode).
			ApplyPrivateMethod(&HwDevMgr{}, "setBoardId", func(_ int32) error { return nil }).
			ApplyPrivateMethod(&HwDevMgr{}, "setDeviceUsage", func(_ int32) error { return nil }).
			ApplyPrivateMethod(&HwDevMgr{}, "setRingInfo", func() error { return nil })
		defer patches.Reset()
		err := mockDevMgr.initInfoRelatedNode()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test method 'initInfoRelatedNode' failed, setBoardId error", func() {
		resetDevMgr()
		var patches = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDevType", api.Ascend910).
			ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetNpuWorkMode", ascommon.AMPMode).
			ApplyPrivateMethod(&HwDevMgr{}, "setBoardId", func(_ int32) error { return testErr })
		defer patches.Reset()
		err := mockDevMgr.initInfoRelatedNode()
		convey.So(err, convey.ShouldResemble, testErr)
	})
	convey.Convey("test method 'initInfoRelatedNode' failed, setDeviceUsage error", func() {
		resetDevMgr()
		var patches = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDevType", api.Ascend910).
			ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetNpuWorkMode", ascommon.AMPMode).
			ApplyPrivateMethod(&HwDevMgr{}, "setBoardId", func(_ int32) error { return nil }).
			ApplyPrivateMethod(&HwDevMgr{}, "setDeviceUsage", func(_ int32) error { return testErr })
		defer patches.Reset()
		err := mockDevMgr.initInfoRelatedNode()
		convey.So(err, convey.ShouldResemble, testErr)
	})
	convey.Convey("test method 'initInfoRelatedNode' failed, setRingInfo error", func() {
		resetDevMgr()
		var patches = gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetDevType", api.Ascend910).
			ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetNpuWorkMode", ascommon.AMPMode).
			ApplyPrivateMethod(&HwDevMgr{}, "setBoardId", func(_ int32) error { return nil }).
			ApplyPrivateMethod(&HwDevMgr{}, "setDeviceUsage", func(_ int32) error { return nil }).
			ApplyPrivateMethod(&HwDevMgr{}, "setRingInfo", func() error { return testErr })
		defer patches.Reset()
		err := mockDevMgr.initInfoRelatedNode()
		convey.So(err, convey.ShouldResemble, testErr)
	})
}
