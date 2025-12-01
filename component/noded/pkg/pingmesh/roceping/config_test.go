/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package roceping for ping by icmp in RoCE mesh net between super pods in A5
package roceping

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestSeqIdSetMethods(t *testing.T) {
	convey.Convey("test SeqIdSet methods", t, func() {
		convey.Convey("01-test Add should be success", func() {
			seq := NewSeqIdSet()
			seq.Add(1, 1)
			convey.So(len(seq.m), convey.ShouldEqual, 1)
		})
		convey.Convey("02-test Delete should be success", func() {
			seq := NewSeqIdSet()
			seq.Add(1, 1)
			seq.Delete(1)
			convey.So(len(seq.m), convey.ShouldEqual, 0)
		})
		convey.Convey("03-test Contains should be false when not contain k", func() {
			seq := NewSeqIdSet()
			seq.Add(1, 1)
			convey.So(seq.Contains(0), convey.ShouldBeFalse)
		})
		convey.Convey("04-test Contains should be true when contain k", func() {
			seq := NewSeqIdSet()
			seq.Add(1, 1)
			convey.So(seq.Contains(1), convey.ShouldBeTrue)
		})
		convey.Convey("05-test Len should be success", func() {
			seq := NewSeqIdSet()
			seq.Add(1, 1)
			convey.So(seq.Len(), convey.ShouldEqual, len(seq.m))
		})
		convey.Convey("06-test DeleteLessValues should return 0 when no element less than target", func() {
			seq := NewSeqIdSet()
			seq.Add(1, 0)
			ret, deletedItems := seq.DeleteLessValues(0)
			convey.So(ret, convey.ShouldEqual, 0)
			convey.So(len(deletedItems), convey.ShouldEqual, 0)
		})
		convey.Convey("07-test DeleteLessValues should return 1 when element less than target exist", func() {
			seq := NewSeqIdSet()
			seq.Add(1, 0)
			ret, deletedItems := seq.DeleteLessValues(1)
			convey.So(ret, convey.ShouldEqual, 1)
			convey.So(deletedItems, convey.ShouldResemble, []int{1})
		})
	})
}

func TestWaitingPktCacheMethods01(t *testing.T) {
	convey.Convey("test WaitingPktCache Methods 01", t, func() {
		convey.Convey("01-test addPktSeqIdToWaitingSet should be success when not exist", func() {
			c := NewWaitingPktCache()
			convey.So(c, convey.ShouldNotBeNil)
			pktData := &IcmpBodyData{
				Uid:   "a",
				SeqId: 1,
			}
			c.addPktSeqIdToWaitingSet(pktData.Uid, pktData.SeqId, time.Now().UnixMilli())
			convey.So(c.Len(), convey.ShouldEqual, 1)
		})

		convey.Convey("02-test addPktSeqIdToWaitingSet should no change when exist", func() {
			c := NewWaitingPktCache()
			pktData := &IcmpBodyData{
				Uid:   "a",
				SeqId: 1,
			}
			c.addPktSeqIdToWaitingSet(pktData.Uid, pktData.SeqId, time.Now().UnixMilli())
			convey.So(c.Len(), convey.ShouldEqual, 1)
			c.addPktSeqIdToWaitingSet(pktData.Uid, pktData.SeqId, time.Now().UnixMilli())
			convey.So(c.Len(), convey.ShouldEqual, 1)
		})

		convey.Convey("03-test isPacketValid should return err when not exist", func() {
			c := NewWaitingPktCache()
			pktData := &IcmpBodyData{
				Uid:   "a",
				SeqId: 1,
			}
			convey.So(c.isPacketValid(nil), convey.ShouldNotBeNil)
			convey.So(c.isPacketValid(pktData), convey.ShouldNotBeNil)
		})

		convey.Convey("04-test isPacketValid should return nil when exist", func() {
			c := NewWaitingPktCache()
			pktData := &IcmpBodyData{
				Uid:   "a",
				SeqId: 1,
			}
			c.addPktSeqIdToWaitingSet(pktData.Uid, pktData.SeqId, time.Now().UnixMilli())
			convey.So(c.isPacketValid(pktData), convey.ShouldBeNil)
		})
	})
}

func TestWaitingPktCacheMethods02(t *testing.T) {
	convey.Convey("test WaitingPktCache Methods 02", t, func() {
		convey.Convey("01-test deletePacketInWaitingSet should return err when input nil", func() {
			c := NewWaitingPktCache()
			pktData := &IcmpBodyData{
				Uid:   "a",
				SeqId: 1,
			}
			c.addPktSeqIdToWaitingSet(pktData.Uid, pktData.SeqId, time.Now().UnixMilli())
			ret := c.deletePacketInWaitingSet(nil)
			convey.So(ret, convey.ShouldNotBeNil)
		})

		convey.Convey("02-test deletePacketInWaitingSet should return nil when input exist", func() {
			c := NewWaitingPktCache()
			pktData := &IcmpBodyData{
				Uid:   "a",
				SeqId: 1,
			}
			c.addPktSeqIdToWaitingSet(pktData.Uid, pktData.SeqId, time.Now().UnixMilli())
			cntBefore := c.Len()
			ret := c.deletePacketInWaitingSet(pktData)
			cntAfter := c.Len()
			convey.So(ret, convey.ShouldBeNil)
			convey.So(cntAfter+1, convey.ShouldEqual, cntBefore)
		})
	})
}

func TestWaitingPktCacheMethods03(t *testing.T) {
	convey.Convey("test WaitingPktCache Methods 03 - calcMayLostPktCnt", t, func() {
		convey.Convey("01-test calcMayLostPktCnt should return 0 when pkt no timeout", func() {
			c := NewWaitingPktCache()
			pktData := &IcmpBodyData{
				Uid:   "a",
				SeqId: 1,
			}
			sendTimestamp := time.Now().UnixMilli()
			c.addPktSeqIdToWaitingSet(pktData.Uid, pktData.SeqId, sendTimestamp)
			ret, detailItems := c.calcMayLostPktCnt()
			convey.So(ret, convey.ShouldEqual, 0)
			convey.So(len(detailItems), convey.ShouldEqual, 0)
		})

		convey.Convey("02-test calcMayLostPktCnt should return 1 when pkt timeout", func() {
			c := NewWaitingPktCache()
			pktData := &IcmpBodyData{
				Uid:   "a",
				SeqId: 1,
			}
			sendTimestamp := time.Now().UnixMilli() - waitTimeoutMilliSec - 1
			c.addPktSeqIdToWaitingSet(pktData.Uid, pktData.SeqId, sendTimestamp)
			ret, detailItems := c.calcMayLostPktCnt()
			convey.So(ret, convey.ShouldEqual, 1)
			convey.So(len(detailItems), convey.ShouldEqual, 1)
			convey.So(c.Len(), convey.ShouldEqual, 0)
		})
	})
}
