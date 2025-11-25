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

// Package app test for queue cache
package app

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"container-manager/pkg/common"
)

func TestQueueCache(t *testing.T) {
	convey.Convey("test method 'Push', 'Pop' and 'Len' success", t, func() {
		resetQueueCache()
		QueueCache.Push(newItem1)
		QueueCache.Push(newItem2)
		QueueCache.Push(newItem3)
		QueueCache.Push(newItem4)
		QueueCache.Push(newItem5)
		convey.So(len(QueueCache.faults), convey.ShouldEqual, len4)

		item := QueueCache.Pop()
		convey.So(item.EventID, convey.ShouldEqual, eventId0)
		convey.So(QueueCache.Len(), convey.ShouldEqual, len3)

		item = QueueCache.Pop()
		convey.So(item.EventID, convey.ShouldEqual, eventId1)
		convey.So(QueueCache.Len(), convey.ShouldEqual, len2)

		item = QueueCache.Pop()
		convey.So(item.EventID, convey.ShouldEqual, eventId2)
		convey.So(QueueCache.Len(), convey.ShouldEqual, len1)

		item = QueueCache.Pop()
		convey.So(item.EventID, convey.ShouldEqual, eventId3)
		convey.So(QueueCache.Len(), convey.ShouldEqual, len0)

		item = QueueCache.Pop()
		convey.So(item, convey.ShouldResemble, common.DevFaultInfo{})
		convey.So(QueueCache.Len(), convey.ShouldEqual, len0)
	})
	convey.Convey("test method 'Push' success, when queue length is exceed the limit", t, func() {
		resetQueueCache()
		var patches = gomonkey.ApplyMethodReturn(&FaultQueue{}, "Len", invalidQueueLen)
		defer patches.Reset()
		QueueCache.Push(newItem1)
		convey.So(len(QueueCache.faults), convey.ShouldEqual, len0)
	})
}
