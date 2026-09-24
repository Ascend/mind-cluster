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

package hardwarefault

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestFaultHardwareLogAppend(t *testing.T) {
	convey.Convey("fault log append sorted and dedup", t, func() {
		l := NewFaultHardwareLog()
		l.Append("n1", 200)
		l.Append("n1", 100)
		l.Append("n1", 200)
		convey.So(l.HasFaultIn("n1", 0, 500), convey.ShouldBeTrue)
		convey.So(l.HasFaultIn("n1", 0, 99), convey.ShouldBeFalse)
		convey.So(l.HasFaultIn("n1", 250, 500), convey.ShouldBeFalse)
	})
}

func TestFaultHardwareLogPruneExpired(t *testing.T) {
	convey.Convey("fault log prune expired", t, func() {
		l := NewFaultHardwareLog()
		l.Append("n1", 100)
		l.Append("n1", 5000)
		l.PruneExpired(5000, 100)
		convey.So(l.HasFaultIn("n1", 0, 6000), convey.ShouldBeTrue)
		convey.So(l.HasFaultIn("n1", 0, 10), convey.ShouldBeFalse)
	})
}

func TestFaultHardwareLogReset(t *testing.T) {
	convey.Convey("fault log reset", t, func() {
		l := NewFaultHardwareLog()
		l.Append("n1", 100)
		l.Reset()
		convey.So(l.HasFaultIn("n1", 0, 6000), convey.ShouldBeFalse)
		convey.So(l.Len(), convey.ShouldEqual, 0)
	})
}

func TestFaultHardwareLogCap(t *testing.T) {
	convey.Convey("fault log caps total records and drops oldest", t, func() {
		l := NewFaultHardwareLog()
		for i := 1; i <= 5000; i++ {
			l.Append("n1", int64(i))
		}
		convey.So(l.Len(), convey.ShouldEqual, maxFaultRecords)
		convey.So(l.HasFaultIn("n1", 1, 1), convey.ShouldBeFalse)
		convey.So(l.HasFaultIn("n1", 5000, 5000), convey.ShouldBeTrue)
	})
}

func TestFaultHardwareLogPruneDaily(t *testing.T) {
	convey.Convey("fault log prune drops records older than 1 day", t, func() {
		l := NewFaultHardwareLog()
		l.Append("n1", 1000)
		l.PruneExpired(2*dailyRetention, dailyRetention)
		convey.So(l.Len(), convey.ShouldEqual, 0)

		l.Append("n1", 2*dailyRetention+500)
		l.PruneExpired(2*dailyRetention+1000, dailyRetention)
		convey.So(l.Len(), convey.ShouldEqual, 1)
	})
}

func TestFaultHardwareLogRunDailyCleanup(t *testing.T) {
	convey.Convey("fault log run daily cleanup returns on canceled context", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		NewFaultHardwareLog().RunDailyCleanup(ctx)
	})
}
