/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package silentfault

import (
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/domain/silentfault"
)

func TestValidatePending(t *testing.T) {
	convey.Convey("validate pending dispatch to first fault mgr", t, func() {
		pendingCache.ResetEvents()
		pendingCache.ResetProcessed()
		firstFaultMgr.Reset()
		faultHardwareLog.Reset()
		enableSilentFault()
		pendingCache.Add(&silentfault.PendingEvent{
			JobID: "j1", FailNode: "n1", TaskNodes: []string{"n1", "n2"}, Timestamp: 100,
		})
		p := gomonkey.ApplyFunc(time.Now, func() time.Time { return time.Unix(150, 0) })
		defer p.Reset()
		validatePending()
		convey.So(len(firstFaultMgr.Nodes()), convey.ShouldEqual, 2)
		events := firstFaultMgr.GetEvents("n1")
		convey.So(len(events), convey.ShouldEqual, 1)
		convey.So(events[0].IsFirst, convey.ShouldBeTrue)
		events2 := firstFaultMgr.GetEvents("n2")
		convey.So(events2[0].IsFirst, convey.ShouldBeFalse)
	})
}

func TestValidatePendingStale(t *testing.T) {
	convey.Convey("validate pending drops stale event", t, func() {
		pendingCache.ResetEvents()
		pendingCache.ResetProcessed()
		firstFaultMgr.Reset()
		faultHardwareLog.Reset()
		enableSilentFault()
		now := int64(10000)
		p := gomonkey.ApplyFunc(time.Now, func() time.Time { return time.Unix(now, 0) })
		defer p.Reset()

		pendingCache.Add(&silentfault.PendingEvent{
			JobID: "j1", FailNode: "n1", TaskNodes: []string{"n1"}, Timestamp: now - 8000,
		})
		validatePending()
		convey.So(len(firstFaultMgr.Nodes()), convey.ShouldEqual, 0)
	})
}

func TestValidatePendingHardwarePresent(t *testing.T) {
	convey.Convey("validate pending skips when hardware fault present", t, func() {
		pendingCache.ResetEvents()
		pendingCache.ResetProcessed()
		firstFaultMgr.Reset()
		faultHardwareLog.Reset()
		enableSilentFault()
		now := int64(1000)
		p := gomonkey.ApplyFunc(time.Now, func() time.Time { return time.Unix(now, 0) })
		defer p.Reset()

		ts := now - 10
		faultHardwareLog.Append("n1", ts*1000)
		pendingCache.Add(&silentfault.PendingEvent{
			JobID: "j1", FailNode: "n1", TaskNodes: []string{"n1"}, Timestamp: ts,
		})
		validatePending()
		convey.So(len(firstFaultMgr.Nodes()), convey.ShouldEqual, 0)
	})
}
