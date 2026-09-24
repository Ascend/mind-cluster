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

	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
)

func TestSilentFaultCmInfo(t *testing.T) {
	convey.Convey("silent fault cm info upsert and get", t, func() {
		c := NewSilentFaultCmInfo()
		c.Upsert("n1", "silent-fault-n1", "clusterd", []string{"Ascend910-0", "Ascend910-1"})
		convey.So(c.Len(), convey.ShouldEqual, 1)

		info, ok := c.Get("n1")
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(info.FaultCode, convey.ShouldEqual, constant.SilentFaultCode)
		convey.So(info.FaultLevel, convey.ShouldEqual, constant.SilentFault)
		convey.So(len(info.DevList), convey.ShouldEqual, 2)
		convey.So(len(info.Sources), convey.ShouldEqual, 1)
	})

	convey.Convey("silent fault cm info refresh and multi source", t, func() {
		c := NewSilentFaultCmInfo()
		c.Upsert("n1", "silent-fault-n1", "clusterd", []string{"Ascend910-0"})
		c.Upsert("n1", "ext-fault-1", "CCAE", []string{"Ascend910-1"})
		info, _ := c.Get("n1")
		convey.So(len(info.Sources), convey.ShouldEqual, 2)
		convey.So(len(info.DevList), convey.ShouldEqual, 1)

		// duplicate same source: Sources stays deduplicated
		c.Upsert("n1", "silent-fault-n1", "clusterd", []string{"Ascend910-2"})
		info, _ = c.Get("n1")
		convey.So(len(info.Sources), convey.ShouldEqual, 2)
		convey.So(info.DevList[0], convey.ShouldEqual, "Ascend910-2")
	})

	convey.Convey("silent fault cm info remove source", t, func() {
		c := NewSilentFaultCmInfo()
		c.Upsert("n1", "silent-fault-n1", "clusterd", []string{"Ascend910-0"})
		c.Upsert("n1", "ext-fault-1", "CCAE", []string{"Ascend910-1"})
		c.RemoveSource("n1", "clusterd"+"silent-fault-n1")
		info, _ := c.Get("n1")
		convey.So(len(info.Sources), convey.ShouldEqual, 1)

		c.RemoveSource("n1", "CCAE"+"ext-fault-1")
		convey.So(c.Len(), convey.ShouldEqual, 0)
	})

	convey.Convey("silent fault cm info restore", t, func() {
		c := NewSilentFaultCmInfo()
		c.Restore("n1", "silent-fault-n1", "clusterd", []string{"Ascend910-0"}, 1000)
		info, ok := c.Get("n1")
		convey.So(ok, convey.ShouldBeTrue)
		// restore keeps the historical LastSeparateTime without resetting to the current time
		convey.So(info.LastSeparateTime, convey.ShouldEqual, 1000)
		convey.So(len(info.Sources), convey.ShouldEqual, 1)
		convey.So(info.DevList[0], convey.ShouldEqual, "Ascend910-0")

		// multiple sources on the same node: Sources merged, LastSeparateTime keeps the larger value
		c.Restore("n1", "ext-1", "CCAE", []string{"Ascend910-1"}, 2000)
		info, _ = c.Get("n1")
		convey.So(len(info.Sources), convey.ShouldEqual, 2)
		convey.So(info.LastSeparateTime, convey.ShouldEqual, 2000)

		// earlier time does not overwrite the later LastSeparateTime
		c.Restore("n1", "ext-1", "CCAE", []string{"Ascend910-1"}, 500)
		info, _ = c.Get("n1")
		convey.So(len(info.Sources), convey.ShouldEqual, 2)
		convey.So(info.LastSeparateTime, convey.ShouldEqual, 2000)
	})

	convey.Convey("silent fault cm info getAll nodes reset", t, func() {
		c := NewSilentFaultCmInfo()
		c.Upsert("n1", "silent-fault-n1", "clusterd", []string{"Ascend910-0"})
		c.Upsert("n2", "ext-1", "CCAE", []string{"Ascend910-1"})

		all := c.GetAll()
		convey.So(len(all), convey.ShouldEqual, 2)
		convey.So(len(all["n1"].Sources), convey.ShouldEqual, 1)

		nodes := c.Nodes()
		convey.So(len(nodes), convey.ShouldEqual, 2)

		c.Reset()
		convey.So(c.Len(), convey.ShouldEqual, 0)
	})

	convey.Convey("silent fault cm info expired", t, func() {
		c := NewSilentFaultCmInfo()
		c.Upsert("n1", "f1", "clusterd", []string{"Ascend910-0"})
		now := time.Now().UnixMilli()
		expired := c.Expired(now+200*1000, 100)
		convey.So(len(expired), convey.ShouldEqual, 1)
		convey.So(expired[0], convey.ShouldEqual, "n1")

		// -1 means no auto release
		expired = c.Expired(now+200*1000, -1)
		convey.So(len(expired), convey.ShouldEqual, 0)
	})
}
