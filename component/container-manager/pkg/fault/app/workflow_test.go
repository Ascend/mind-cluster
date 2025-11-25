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

// Package app test for fault manager workflow
package app

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/devmanager/common"
	"container-manager/pkg/devmgr"
)

func TestFaultMgr(t *testing.T) {
	convey.Convey("test method 'Name' success", t, testMethodName)
	convey.Convey("test method 'Init' success", t, testMethodInit)
	convey.Convey("test method 'Work' success", t, testMethodWork)
	convey.Convey("test method 'ShutDown' success", t, testMethodShutDown)
}

func testMethodName() {
	convey.So(mockFaultMgr.Name(), convey.ShouldEqual, "fault manager")
}

func testMethodInit() {
	convey.Convey("test method 'Init' success", func() {
		var patches = gomonkey.ApplyFuncReturn(loadFaultCodeFromFile, nil)
		defer patches.Reset()
		convey.So(mockFaultMgr.Init(), convey.ShouldBeNil)
	})
	convey.Convey("test method 'Init' failed, load file error", func() {
		var patches = gomonkey.ApplyFuncReturn(loadFaultCodeFromFile, testErr)
		defer patches.Reset()
		err := mockFaultMgr.Init()
		expErr := errors.New("load fault code from file failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func testMethodWork() {
	var hasExecuted bool
	var p1 = gomonkey.ApplyPrivateMethod(&FaultMgr{}, "getAllFaultInfo", func() {}).
		ApplyMethodReturn(&devmgr.HwDevMgr{}, "SubscribeFaultEvent",
			func(callback func(devFaultInfo common.DevFaultInfo)) error { return nil }).
		ApplyMethod(&FaultMgr{}, "ProcessDCMIFault", func(fm *FaultMgr, ctx context.Context) {}).
		ApplyPrivateMethod(&FaultMgr{}, "checkMoreThanFiveMinFaults", func(_ context.Context) {}).
		ApplyMethod(&FaultMgr{}, "Work", func(fm *FaultMgr, ctx context.Context) {
			hasExecuted = true
			return
		})
	defer p1.Reset()
	mockFaultMgr.Work(context.Background())
	convey.So(hasExecuted, convey.ShouldBeTrue)
}

func testMethodShutDown() {
	var hasExecuted bool
	var p2 = gomonkey.ApplyMethod(&FaultMgr{}, "ShutDown", func(hdm *FaultMgr) {
		hasExecuted = true
		return
	})
	defer p2.Reset()
	mockFaultMgr.ShutDown()
	convey.So(hasExecuted, convey.ShouldBeTrue)
}
