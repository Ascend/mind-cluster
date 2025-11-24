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

// Package domain test for fault cache
package domain

import (
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"container-manager/pkg/common"
)

var mockFaultCache = &FaultCache{}

func resetFaultCache() {
	mockFaultCache = &FaultCache{
		faults:     make(map[int32]map[int64]map[string]*common.DevFaultInfo),
		UpdateChan: make(chan struct{}, 1),
		mutex:      sync.Mutex{},
	}
}

func TestAddFault(t *testing.T) {
	convey.Convey("receive occur fault", t, testFaultOccur)
	convey.Convey("receive recover fault", t, testFaultRecover)
	convey.Convey("receive recover fault", t, testFaultOnce)
}

func testFaultOccur() {
	resetFaultCache()
	mockFaultCache.AddFault(*mockFault1)
	mockFaultCache.AddFault(*mockFault2)
	mockFaultCache.AddFault(*mockFault3)
	mockFaultCache.AddFault(*mockFault4)
	convey.So(len(mockFaultCache.faults), convey.ShouldEqual, len2)
	convey.So(len(mockFaultCache.faults[devId0]), convey.ShouldEqual, len2)
	convey.So(len(mockFaultCache.faults[devId0][eventId1]), convey.ShouldEqual, len2)
}

func testFaultRecover() {
	resetFaultCache()
	mockFaultCache.AddFault(*mockFault1)
	mockFaultCache.AddFault(*mockFault2)
	mockFaultCache.AddFault(*mockFault3)
	mockFaultCache.AddFault(*mockFault4)
	mockFaultCache.AddFault(*mockFault5)
	convey.So(len(mockFaultCache.faults), convey.ShouldEqual, len2)
	convey.So(len(mockFaultCache.faults[devId0]), convey.ShouldEqual, len2)
	convey.So(len(mockFaultCache.faults[devId0][eventId1]), convey.ShouldEqual, len1)
	mockFaultCache.AddFault(*mockFault6)
	convey.So(len(mockFaultCache.faults), convey.ShouldEqual, len1)
}

func testFaultOnce() {
	resetFaultCache()
	mockFaultCache.AddFault(*mockFault7)
	convey.So(len(mockFaultCache.faults), convey.ShouldEqual, len0)
}

func TestDeepCopy(t *testing.T) {
	convey.Convey("test method 'DeepCopy'", t, func() {
		resetFaultCache()
		mockFaultCache.AddFault(*mockFault1)
		cpFault, err := mockFaultCache.DeepCopy()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(cpFault), convey.ShouldEqual, len1)

		p1 := gomonkey.ApplyFuncReturn(common.DeepCopy, testErr)
		defer p1.Reset()
		_, err = mockFaultCache.DeepCopy()
		convey.So(err, convey.ShouldResemble, testErr)
	})
}

func TestUpdateFaultsOnDev(t *testing.T) {
	convey.Convey("test func 'UpdateFaultsOnDev'", t, func() {
		resetFaultCache()
		mockFaultCache.AddFault(*mockFault1)
		mockFaultCache.AddFault(*mockFault2)
		mockFaultCache.AddFault(*mockFault3)
		mockFaultCache.AddFault(*mockFault5)
		mockFaultCache.UpdateFaultsOnDev(devId0, []int64{eventId0})
		convey.So(len(mockFaultCache.faults), convey.ShouldEqual, len1)
	})
	convey.Convey("test func 'UpdateFaultsOnDev', new fault codes is nil", t, func() {
		resetFaultCache()
		mockFaultCache.AddFault(*mockFault1)
		mockFaultCache.UpdateFaultsOnDev(devId0, []int64{})
		convey.So(len(mockFaultCache.faults[devId1]), convey.ShouldEqual, len0)
	})
}

func TestConstructMockModuleFault(t *testing.T) {
	convey.Convey("test func 'ConstructMockModuleFault'", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(GetFaultLevelByCode, common.SeparateNPU)
		defer p1.Reset()
		res := ConstructMockModuleFault(devId0, eventId0)
		expRes := &common.DevFaultInfo{
			EventID:       eventId0,
			LogicID:       mockFaultAttr,
			ModuleType:    mockFaultAttr,
			ModuleID:      mockFaultAttr,
			SubModuleType: mockFaultAttr,
			SubModuleID:   mockFaultAttr,
			Assertion:     common.FaultOccur,
			PhyID:         devId0,
			FaultLevel:    common.SeparateNPU,
			ReceiveTime:   time.Now().Unix(),
		}
		convey.So(res.EventID, convey.ShouldEqual, expRes.EventID)
		convey.So(res.LogicID, convey.ShouldEqual, expRes.LogicID)
		convey.So(res.ModuleType, convey.ShouldEqual, expRes.ModuleType)
		convey.So(res.ModuleID, convey.ShouldEqual, expRes.ModuleID)
		convey.So(res.SubModuleType, convey.ShouldEqual, expRes.SubModuleType)
		convey.So(res.SubModuleID, convey.ShouldEqual, expRes.SubModuleID)
		convey.So(res.Assertion, convey.ShouldEqual, expRes.Assertion)
		convey.So(res.PhyID, convey.ShouldEqual, expRes.PhyID)
		convey.So(res.FaultLevel, convey.ShouldEqual, expRes.FaultLevel)
	})
}
