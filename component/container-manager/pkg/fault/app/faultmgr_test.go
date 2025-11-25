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

// Package app test for fault manager
package app

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	ascommon "ascend-common/devmanager/common"
	"container-manager/pkg/common"
	"container-manager/pkg/devmgr"
	"container-manager/pkg/fault/domain"
)

func TestFaultMgr2(t *testing.T) {
	convey.Convey("test method 'ProcessDCMIFault' success", t, testProcessDCMIFault)
	convey.Convey("test method 'getAllFaultInfo' success", t, testGetAllFaultInfo)
	convey.Convey("test method 'saveDevFaultInfo' success", t, testSaveDevFaultInfo)
	convey.Convey("test method 'doCheck' success", t, testDoCheck)
}

func testProcessDCMIFault() {
	var patches = gomonkey.ApplyMethod(&domain.FaultCache{}, "AddFault",
		func(fc *domain.FaultCache, newFault common.DevFaultInfo) {
			return
		})
	defer patches.Reset()
	const (
		waitGoroutineFinishedTime = 200 * time.Millisecond
		waitDeleteFinishedTime    = 2 * time.Second
	)

	// prepare data
	resetQueueCache()
	QueueCache.Push(newItem1)
	QueueCache.Push(newItem2)
	QueueCache.Push(newItem3)
	QueueCache.Push(newItem4)

	// ctx stop
	ctx, cancel := context.WithCancel(context.Background())
	haveStopped := atomic.Bool{}
	go func() {
		mockFaultMgr.ProcessDCMIFault(ctx)
		haveStopped.Store(true)
	}()
	cancel()
	time.Sleep(waitGoroutineFinishedTime)
	convey.So(haveStopped.Load(), convey.ShouldBeTrue)

	// delete faults
	ctx, cancel = context.WithCancel(context.Background())
	go func() {
		mockFaultMgr.ProcessDCMIFault(ctx)
	}()
	time.Sleep(waitDeleteFinishedTime)
	cancel()
	time.Sleep(waitGoroutineFinishedTime)
	convey.So(QueueCache.Len(), convey.ShouldEqual, 0)
}

func testGetAllFaultInfo() {
	resetQueueCache()
	var patches = gomonkey.ApplyMethodReturn(&devmgr.HwDevMgr{}, "GetFaultCodesMap",
		map[int32][]int64{
			devId0: {eventId0, eventId1, eventId2},
			devId1: {eventId3},
		}).ApplyFuncReturn(domain.GetFaultLevelByCode, common.SeparateNPU)
	defer patches.Reset()
	mockFaultMgr.getAllFaultInfo()
	convey.So(QueueCache.Len(), convey.ShouldEqual, len4)
}

func testSaveDevFaultInfo() {
	resetQueueCache()
	var patches = gomonkey.ApplyFuncReturn(domain.GetFaultLevelByCode, common.SeparateNPU)
	defer patches.Reset()
	mockDevFault1 := ascommon.DevFaultInfo{
		EventID: eventId0,
		LogicID: devId0,
	}
	mockDevFault2 := ascommon.DevFaultInfo{
		EventID: eventId1,
		LogicID: devId1,
	}
	mockDevFault3 := ascommon.DevFaultInfo{
		EventID: 0,
		LogicID: devId1,
	}
	saveDevFaultInfo(mockDevFault1)
	saveDevFaultInfo(mockDevFault2)
	saveDevFaultInfo(mockDevFault3)
	convey.So(QueueCache.Len(), convey.ShouldEqual, len2)
}

func testDoCheck() {
	mockFault1 := common.DevFaultInfo{
		EventID:       eventId0,
		PhyID:         devId0,
		ModuleType:    moduleId0,
		ModuleID:      moduleId0,
		SubModuleType: moduleId0,
		SubModuleID:   moduleId0,
		Assertion:     common.FaultOccur,
		ReceiveTime:   time.Now().Unix(),
	}
	mockFault2 := common.DevFaultInfo{
		EventID:       eventId1,
		PhyID:         devId1,
		ModuleType:    moduleId1,
		ModuleID:      moduleId1,
		SubModuleType: moduleId1,
		SubModuleID:   moduleId1,
		Assertion:     common.FaultOccur,
		ReceiveTime:   time.Now().Unix() - faultExistedDuration,
	}
	mockFaultMgr.faultInfo.AddFault(mockFault1)
	mockFaultMgr.faultInfo.AddFault(mockFault2)
	var patches = gomonkey.ApplyMethodReturn(&devmgr.HwDevMgr{}, "GetDeviceErrCode", int32(1), []int64{eventId1}, nil)
	defer patches.Reset()
	mockFaultMgr.doCheck()
	faults, err := mockFaultMgr.faultInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(faults), convey.ShouldEqual, len2)
}
