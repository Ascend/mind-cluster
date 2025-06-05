// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault test for public fault cache
package publicfault

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestPubFaultNeedDelete(t *testing.T) {
	convey.Convey("test method 'Push', 'Pop' and 'Len' success", t, testBasicMethod)
	convey.Convey("test method 'DealDelete' success", t, testDealDelete)
}

func testBasicMethod() {
	const (
		len4 = 4
		len3 = 3
		len2 = 2
		len1 = 1
		len0 = 0
	)
	resetQueueCache()
	deleteTime := time.Now().Unix()
	PubFaultNeedDelete.Push(deleteTime, testNodeName1, faultKey1)
	PubFaultNeedDelete.Push(deleteTime, testNodeName2, faultKey2)
	PubFaultNeedDelete.Push(deleteTime, testNodeName1, faultKey3)
	PubFaultNeedDelete.Push(deleteTime, testNodeName2, faultKey4)
	convey.So(len(PubFaultNeedDelete.faults), convey.ShouldEqual, len4)

	item := PubFaultNeedDelete.Pop()
	convey.So(item.faultKey, convey.ShouldEqual, faultKey1)
	convey.So(PubFaultNeedDelete.Len(), convey.ShouldEqual, len3)

	item = PubFaultNeedDelete.Pop()
	convey.So(item.faultKey, convey.ShouldEqual, faultKey2)
	convey.So(PubFaultNeedDelete.Len(), convey.ShouldEqual, len2)

	item = PubFaultNeedDelete.Pop()
	convey.So(item.faultKey, convey.ShouldEqual, faultKey3)
	convey.So(PubFaultNeedDelete.Len(), convey.ShouldEqual, len1)

	item = PubFaultNeedDelete.Pop()
	convey.So(item.faultKey, convey.ShouldEqual, faultKey4)
	convey.So(PubFaultNeedDelete.Len(), convey.ShouldEqual, len0)

	item = PubFaultNeedDelete.Pop()
	convey.So(item, convey.ShouldResemble, needDeleteFault{})
	convey.So(PubFaultNeedDelete.Len(), convey.ShouldEqual, len0)
}

func testDealDelete() {
	const (
		diffTime                  = 2
		waitGoroutineFinishedTime = 200 * time.Millisecond
		waitDeleteFinishedTime    = 2 * time.Second
	)

	// prepare data
	resetQueueCache()
	deleteTime := time.Now().Unix()
	PubFaultNeedDelete.Push(deleteTime-1, testNodeName1, faultKey1)
	PubFaultNeedDelete.Push(deleteTime, testNodeName2, faultKey2)
	PubFaultNeedDelete.Push(deleteTime+1, testNodeName1, faultKey3)
	PubFaultNeedDelete.Push(deleteTime+diffTime, testNodeName2, faultKey4)

	// ctx stop
	ctx, cancel := context.WithCancel(context.Background())
	haveStopped := atomic.Bool{}
	go func() {
		PubFaultNeedDelete.DealDelete(ctx)
		haveStopped.Store(true)
	}()
	cancel()
	time.Sleep(waitGoroutineFinishedTime)
	convey.So(haveStopped.Load(), convey.ShouldBeTrue)

	// delete faults
	ctx, cancel = context.WithCancel(context.Background())
	go func() {
		PubFaultNeedDelete.DealDelete(ctx)
	}()
	time.Sleep(waitDeleteFinishedTime)
	cancel()
	time.Sleep(waitGoroutineFinishedTime)
	convey.So(PubFaultNeedDelete.Len(), convey.ShouldEqual, 0)
}

func resetQueueCache() {
	PubFaultNeedDelete = &needDeleteQueue{
		faults: make([]needDeleteFault, 0),
		mutex:  sync.Mutex{},
	}
}
