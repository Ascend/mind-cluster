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

// Package domain test for shared cache
package domain

import (
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"container-manager/pkg/common"
)

const invalidNPUFaultNum = 200

var mockSharedCache = &sharedCache{}

func resetSharedCache() {
	mockSharedCache = &sharedCache{
		faults:     make(map[int32][]*common.DevFaultInfo),
		UpdateChan: make(chan struct{}, 1),
		mutex:      sync.Mutex{},
	}
}

func TestAddFaultForSharedCache(t *testing.T) {
	convey.Convey("receive occur fault, recover fault and once fault", t, func() {
		resetSharedCache()
		mockSharedCache.AddFault(mockFault1)
		mockSharedCache.AddFault(mockFault2)
		mockSharedCache.AddFault(mockFault3)
		mockSharedCache.AddFault(mockFault4)
		mockSharedCache.AddFault(mockFault5)
		mockSharedCache.AddFault(mockFault6)
		mockSharedCache.AddFault(mockFault7)
		convey.So(len(mockSharedCache.faults), convey.ShouldEqual, len2)
	})
	convey.Convey("new fault is nil", t, func() {
		resetSharedCache()
		mockSharedCache.AddFault(nil)
		convey.So(len(mockSharedCache.faults), convey.ShouldEqual, len0)
	})
	convey.Convey("fault length is exceed the limit", t, func() {
		resetSharedCache()
		var p1 = gomonkey.ApplyPrivateMethod(&sharedCache{},
			"getNPUFaultNum", func(_ int32) int { return invalidNPUFaultNum })
		defer p1.Reset()
		mockSharedCache.AddFault(mockFault1)
		convey.So(len(mockSharedCache.faults), convey.ShouldEqual, len0)
	})
}

func TestGetAndClean(t *testing.T) {
	convey.Convey("test method 'GetAndClean'", t, func() {
		resetSharedCache()
		mockSharedCache.AddFault(mockFault1)
		mockSharedCache.AddFault(mockFault2)
		mockSharedCache.AddFault(mockFault3)
		mockSharedCache.AddFault(mockFault4)
		res := mockSharedCache.GetAndClean()
		convey.So(len(mockSharedCache.faults), convey.ShouldEqual, len0)
		convey.So(len(res[devId0]), convey.ShouldEqual, len3)
		convey.So(len(res[devId1]), convey.ShouldEqual, len1)
		convey.So(len(res), convey.ShouldEqual, len2)
	})
}

func TestDeepCopyForSharedCache(t *testing.T) {
	convey.Convey("test method 'DeepCopy'", t, func() {
		resetSharedCache()
		mockSharedCache.AddFault(mockFault1)
		cpFault, err := mockSharedCache.DeepCopy()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(cpFault), convey.ShouldEqual, len1)

		p1 := gomonkey.ApplyFuncReturn(common.DeepCopy, testErr)
		defer p1.Reset()
		_, err = mockSharedCache.DeepCopy()
		convey.So(err, convey.ShouldResemble, testErr)
	})
}
