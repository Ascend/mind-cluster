// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

//go:build !race

// Package jobv2 a series of job test function
package jobv2

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"golang.org/x/time/rate"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

const (
	time50ms  = 50 * time.Millisecond
	key1001   = 1001
	value1001 = 1001
)

func TestPreDeleteToDelete(t *testing.T) {
	uniqueQueue = sync.Map{}
	convey.Convey("test preDeleteToDelete", t, func() {
		convey.Convey("test deleteKeys len is 0", func() {
			preDeleteToDelete()
			messageLength := 0
			uniqueQueue.Range(func(key, value interface{}) bool {
				messageLength++
				return true
			})
			convey.So(messageLength, convey.ShouldEqual, 0)
		})
		convey.Convey("test delete Key is 123", func() {
			mockGetShouldDeleteJobKey := gomonkey.ApplyFunc(job.GetShouldDeleteJobKey, func() []string {
				return []string{jobUid1}
			})
			defer mockGetShouldDeleteJobKey.Reset()
			preDeleteToDelete()
			value, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(value, convey.ShouldEqual, queueOperatorDelete)
		})
	})
}

func TestAddUpdateMessageIfOutdated(t *testing.T) {
	uniqueQueue = sync.Map{}
	convey.Convey("test addUpdateMessageIfOutdated", t, func() {
		convey.Convey("test should update job len is 0", func() {
			addUpdateMessageIfOutdated()
			messageLength := 0
			uniqueQueue.Range(func(key, value interface{}) bool {
				messageLength++
				return true
			})
			convey.So(messageLength, convey.ShouldEqual, 0)
		})
		convey.Convey("test should update job is 123", func() {
			mockGetShouldUpdateJobKey := gomonkey.ApplyFunc(job.GetShouldUpdateJobKey, func() []string {
				return []string{jobUid1}
			})
			defer mockGetShouldUpdateJobKey.Reset()
			addUpdateMessageIfOutdated()
			value, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(value, convey.ShouldEqual, queueOperatorUpdate)
		})
	})
}

func TestCheckQueueBlock(t *testing.T) {
	uniqueQueue = sync.Map{}
	convey.Convey("test checkQueueBlock", t, func() {
		convey.Convey("test queue len is 0", func() {
			for i := 0; i < messageNumThreshold; i++ {
				uniqueQueue.Store(i, i)
			}
			isTooLarge := checkQueueBlock()
			convey.So(isTooLarge, convey.ShouldEqual, false)
		})
		convey.Convey("test queue len is equals 1000", func() {
			for i := 0; i < messageNumThreshold; i++ {
				uniqueQueue.Store(i, i)
			}
			isTooLarge := checkQueueBlock()
			convey.So(isTooLarge, convey.ShouldEqual, false)
		})
		convey.Convey("test queue len is more than 1000", func() {
			uniqueQueue.Store(key1001, value1001)
			isTooLarge := checkQueueBlock()
			convey.So(isTooLarge, convey.ShouldEqual, true)
		})
	})
}

func TestHandlerQueueIsNil(t *testing.T) {
	uniqueQueue = sync.Map{}
	convey.Convey("test Handler test queue len is 0", t, func() {
		go Handler(context.TODO())
		messageLength := 0
		uniqueQueue.Range(func(key, value interface{}) bool {
			messageLength++
			return true
		})
		convey.So(messageLength, convey.ShouldEqual, 0)
	})
}

func TestHandlerQueueIsNotNil(t *testing.T) {
	uniqueQueue = sync.Map{}
	convey.Convey("test Handler test queue len is 1", t, func() {
		convey.Convey("test queue len is 1", func() {
			uniqueQueue.Store(jobUid1, queueOperatorDelete)
			mockDeleteJob := gomonkey.ApplyFunc(deleteJob, func(jobUniqueKey string) {
			})
			defer mockDeleteJob.Reset()
			go Handler(context.TODO())
			time.Sleep(time50ms)
			_, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
		})
	})
}

func TestHandlerLimiterIsError(t *testing.T) {
	uniqueQueue = sync.Map{}
	uniqueQueue.Store(jobUid1, queueOperatorDelete)
	convey.Convey("test Handler test limiter is failed", t, func() {
		mockLimiter := gomonkey.ApplyMethod(limiter, "Wait",
			func(_ *rate.Limiter, ctx context.Context) error {
				return errors.New("test error")
			})
		go Handler(context.TODO())
		time.Sleep(time50ms)
		_, ok := uniqueQueue.Load(jobUid1)
		convey.So(ok, convey.ShouldEqual, true)
		mockLimiter.Reset()
	})
}

func TestPodGroupMessage(t *testing.T) {
	convey.Convey("test podGroupMessage", t, func() {
		pg := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
		convey.Convey("test operator is add", func() {
			uniqueQueue = sync.Map{}
			podGroupMessage(pg, constant.AddOperator)
			value, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(value, convey.ShouldEqual, queueOperatorAdd)
		})
		convey.Convey("test operator is delete", func() {
			uniqueQueue = sync.Map{}
			podGroupMessage(pg, constant.DeleteOperator)
			value, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(value, convey.ShouldEqual, queueOperatorPreDelete)
		})
		convey.Convey("test operator is update", func() {
			uniqueQueue = sync.Map{}
			podGroupMessage(pg, constant.UpdateOperator)
			value, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(value, convey.ShouldEqual, queueOperatorUpdate)
		})
		convey.Convey("test operator is illegal", func() {
			uniqueQueue = sync.Map{}
			podGroupMessage(pg, "illegal")
			_, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
		})
	})
}

func TestChecker(t *testing.T) {
	convey.Convey("Checker function test suite", t, func() {
		testContextCancelation()
		testHourTimerTrigger()
		testCheckQueueBlockTrueCase()
		testMinuteTimerTrigger()
	})
}

func testContextCancelation() {
	convey.Convey("Should exit when context is canceled", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		hourCh := make(chan time.Time, 1)
		minuteCh := make(chan time.Time, 1)
		patches := gomonkey.NewPatches()
		patches.ApplyFunc(time.NewTicker, func(d time.Duration) *time.Ticker {
			if d == time.Hour {
				return &time.Ticker{C: hourCh}
			}
			return &time.Ticker{C: minuteCh}
		})
		defer patches.Reset()

		exit := false
		go func() {
			Checker(ctx)
			exit = true
		}()

		cancel()
		time.Sleep(time.Second)
		convey.ShouldBeTrue(exit)
	})
}

func testHourTimerTrigger() {
	convey.Convey("Should execute hour logic when hour timer triggers", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		hourCh := make(chan time.Time, 1)
		patches := gomonkey.NewPatches()
		patches.ApplyFunc(time.NewTicker, func(d time.Duration) *time.Ticker {
			if d == time.Hour {
				return &time.Ticker{C: hourCh}
			}
			return &time.Ticker{C: make(chan time.Time)}
		})
		defer patches.Reset()

		checkCalled := false
		addUpdateCalled := false
		patches.ApplyFunc(checkQueueBlock, func() bool {
			checkCalled = true
			return false
		})
		patches.ApplyFunc(addUpdateMessageIfOutdated, func() {
			addUpdateCalled = true
		})

		go Checker(ctx)
		hourCh <- time.Now()
		time.Sleep(time.Second)

		convey.So(checkCalled, convey.ShouldBeTrue)
		convey.So(addUpdateCalled, convey.ShouldBeTrue)
	})
}

func testCheckQueueBlockTrueCase() {
	convey.Convey("Should not call addUpdate when checkQueueBlock returns true", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		hourCh := make(chan time.Time, 1)
		patches := gomonkey.NewPatches()
		patches.ApplyFunc(time.NewTicker, func(d time.Duration) *time.Ticker {
			return &time.Ticker{C: hourCh}
		})
		defer patches.Reset()

		addUpdateCalled := false
		patches.ApplyFunc(checkQueueBlock, func() bool {
			return true
		})
		patches.ApplyFunc(addUpdateMessageIfOutdated, func() {
			addUpdateCalled = true
		})

		go Checker(ctx)
		hourCh <- time.Now()
		time.Sleep(time.Second)

		convey.So(addUpdateCalled, convey.ShouldBeFalse)
	})
}

func testMinuteTimerTrigger() {
	convey.Convey("Should call preDeleteToDelete when minute timer triggers", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		minuteCh := make(chan time.Time, 1)
		patches := gomonkey.NewPatches()
		patches.ApplyFunc(time.NewTicker, func(d time.Duration) *time.Ticker {
			if d == time.Minute {
				return &time.Ticker{C: minuteCh}
			}
			return &time.Ticker{C: make(chan time.Time)}
		})
		defer patches.Reset()

		preDeleteCalled := false
		patches.ApplyFunc(preDeleteToDelete, func() {
			preDeleteCalled = true
		})

		go Checker(ctx)
		minuteCh <- time.Now()
		time.Sleep(time.Second)

		convey.So(preDeleteCalled, convey.ShouldBeTrue)
	})
}
