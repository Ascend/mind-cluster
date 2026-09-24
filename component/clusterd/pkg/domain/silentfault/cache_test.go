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

	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/domain/conf"
)

func TestPendingCache(t *testing.T) {
	convey.Convey("pending cache add and take due", t, func() {
		policy := conf.SilentFaultPolicy{}
		policy.Detect.HardwareFaultWindowSecond = 30
		conf.SetSilentFaultPolicy(policy)

		c := NewPendingCache()
		c.Add(&PendingEvent{JobID: "j1", FailNode: "n1", TaskNodes: []string{"n1", "n2"}, Timestamp: 100})
		c.Add(&PendingEvent{JobID: "j1", FailNode: "n1", TaskNodes: []string{"n1"}, Timestamp: 200})

		due := c.TakeDue(135)
		convey.So(len(due), convey.ShouldEqual, 1)
		convey.So(due[0].Timestamp, convey.ShouldEqual, 100)
	})

	convey.Convey("pending cache processed prune", t, func() {
		c := NewPendingCache()
		c.MarkProcessed("j1", 1)
		c.MarkProcessed("j1", 2)
		c.MarkProcessed("j2", 3)
		convey.So(c.HasProcessed("j1", 1), convey.ShouldBeTrue)

		c.PruneProcessed("j1", map[int64]struct{}{2: {}})
		convey.So(c.HasProcessed("j1", 1), convey.ShouldBeFalse)
		convey.So(c.HasProcessed("j1", 2), convey.ShouldBeTrue)

		c.PruneOrphanProcessed(map[string]struct{}{"j1": {}})
		convey.So(c.HasProcessed("j2", 3), convey.ShouldBeFalse)
	})

	convey.Convey("pending cache reset", t, func() {
		c := NewPendingCache()
		c.Add(&PendingEvent{JobID: "j1", Timestamp: 1})
		c.MarkProcessed("j1", 1)
		c.ResetEvents()
		c.ResetProcessed()
		convey.So(len(c.TakeDue(100)), convey.ShouldEqual, 0)
		convey.So(c.HasProcessed("j1", 1), convey.ShouldBeFalse)
	})
}

func TestFirstFaultMgr(t *testing.T) {
	convey.Convey("first fault mgr add and query", t, func() {
		m := NewFirstFaultMgr()
		m.AddEvent("n1", &FirstFaultEvent{JobID: "j1", IsFirst: true, Timestamp: 1000})
		m.AddEvent("n1", &FirstFaultEvent{JobID: "j2", IsFirst: false, Timestamp: 2000})
		m.AddEvent("n2", &FirstFaultEvent{JobID: "j1", IsFirst: false, Timestamp: 1500})

		convey.So(len(m.Nodes()), convey.ShouldEqual, 2)
		events := m.GetEvents("n1")
		convey.So(len(events), convey.ShouldEqual, 2)
		convey.So(events[0].IsFirst, convey.ShouldBeTrue)
		convey.So(events[1].IsFirst, convey.ShouldBeFalse)
	})

	convey.Convey("first fault mgr prune all", t, func() {
		m := NewFirstFaultMgr()
		m.AddEvent("n1", &FirstFaultEvent{JobID: "j1", Timestamp: 1000})
		m.AddEvent("n2", &FirstFaultEvent{JobID: "j1", Timestamp: 5000})
		m.PruneExpiredAll(6000, 2000)
		convey.So(len(m.Nodes()), convey.ShouldEqual, 1)
		convey.So(m.GetEvents("n2")[0].Timestamp, convey.ShouldEqual, 5000)
	})

	convey.Convey("first fault mgr clear node and reset", t, func() {
		m := NewFirstFaultMgr()
		m.AddEvent("n1", &FirstFaultEvent{JobID: "j1", Timestamp: 1000})
		m.AddEvent("n2", &FirstFaultEvent{JobID: "j1", Timestamp: 2000})

		m.ClearNode("n1")
		convey.So(len(m.Nodes()), convey.ShouldEqual, 1)

		m.Reset()
		convey.So(len(m.Nodes()), convey.ShouldEqual, 0)
	})
}
