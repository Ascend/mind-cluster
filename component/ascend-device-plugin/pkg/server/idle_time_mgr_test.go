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

package server

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		return
	}
}

func TestIdleTimeMgrRecordIdleTime(t *testing.T) {
	convey.Convey("test IdleTimeMgr RecordIdleTime", t, func() {
		convey.Convey("first record should store idle time", func() {
			mgr := NewIdleTimeMgr()
			mgr.RecordIdleTime(0)
			_, ok := mgr.GetIdleTime(0)
			convey.So(ok, convey.ShouldBeTrue)
		})
		convey.Convey("second record should not update idle time", func() {
			mgr := NewIdleTimeMgr()
			mgr.RecordIdleTime(0)
			firstTime, _ := mgr.GetIdleTime(0)
			time.Sleep(10 * time.Millisecond)
			mgr.RecordIdleTime(0)
			secondTime, _ := mgr.GetIdleTime(0)
			convey.So(secondTime.Equal(firstTime), convey.ShouldBeTrue)
		})
	})
}

func TestIdleTimeMgrDeleteIdleTime(t *testing.T) {
	convey.Convey("test IdleTimeMgr DeleteIdleTime", t, func() {
		convey.Convey("delete should remove idle time record", func() {
			mgr := NewIdleTimeMgr()
			mgr.RecordIdleTime(0)
			mgr.DeleteIdleTime(0)
			_, ok := mgr.GetIdleTime(0)
			convey.So(ok, convey.ShouldBeFalse)
		})
		convey.Convey("delete non-existent record should not panic", func() {
			mgr := NewIdleTimeMgr()
			mgr.DeleteIdleTime(99)
			_, ok := mgr.GetIdleTime(99)
			convey.So(ok, convey.ShouldBeFalse)
		})
	})
}

func TestIdleTimeMgrIsIdleTimeExceeded(t *testing.T) {
	convey.Convey("test IdleTimeMgr IsIdleTimeExceeded", t, func() {
		convey.Convey("no record should return false", func() {
			mgr := NewIdleTimeMgr()
			convey.So(mgr.IsIdleTimeExceeded(0, 60), convey.ShouldBeFalse)
		})
		convey.Convey("idle time not exceeded should return false", func() {
			mgr := NewIdleTimeMgr()
			mgr.RecordIdleTime(0)
			convey.So(mgr.IsIdleTimeExceeded(0, 60), convey.ShouldBeFalse)
		})
	})
}

func TestIdleTimeMgrIsIdleTimeExceededAfterWait(t *testing.T) {
	convey.Convey("test IdleTimeMgr IsIdleTimeExceeded after wait", t, func() {
		convey.Convey("idle time exceeded should return true", func() {
			mgr := NewIdleTimeMgr()
			mgr.idleTimes.Store(int32(0), time.Now().Add(-61*time.Second))
			convey.So(mgr.IsIdleTimeExceeded(0, 60), convey.ShouldBeTrue)
		})
		convey.Convey("idle time exactly at boundary should return true", func() {
			mgr := NewIdleTimeMgr()
			mgr.idleTimes.Store(int32(0), time.Now().Add(-60*time.Second))
			convey.So(mgr.IsIdleTimeExceeded(0, 60), convey.ShouldBeTrue)
		})
	})
}

func TestIdleTimeMgrGetIdleTime(t *testing.T) {
	convey.Convey("test IdleTimeMgr GetIdleTime", t, func() {
		convey.Convey("no record should return false", func() {
			mgr := NewIdleTimeMgr()
			idleTime, ok := mgr.GetIdleTime(0)
			convey.So(ok, convey.ShouldBeFalse)
			convey.So(idleTime.IsZero(), convey.ShouldBeTrue)
		})
		convey.Convey("recorded idle time should return correct value", func() {
			mgr := NewIdleTimeMgr()
			beforeRecord := time.Now()
			mgr.RecordIdleTime(0)
			idleTime, ok := mgr.GetIdleTime(0)
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(idleTime.After(beforeRecord) || idleTime.Equal(beforeRecord), convey.ShouldBeTrue)
		})
		convey.Convey("different logicIDs should have independent records", func() {
			mgr := NewIdleTimeMgr()
			mgr.RecordIdleTime(0)
			mgr.RecordIdleTime(1)
			_, ok0 := mgr.GetIdleTime(0)
			_, ok1 := mgr.GetIdleTime(1)
			convey.So(ok0, convey.ShouldBeTrue)
			convey.So(ok1, convey.ShouldBeTrue)
		})
	})
}
