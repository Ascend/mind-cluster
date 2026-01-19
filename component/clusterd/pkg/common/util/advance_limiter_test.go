// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package util a series of util function
package util

import (
	"context"
	"sync"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"golang.org/x/time/rate"
)

const (
	r, burst, maxWaitQueue = 10, 5, 20
	smallQuelen            = 2
	resultLen              = 3
)

// TestNewAdvancedRateLimiter test basic new limit
func TestNewAdvancedRateLimiter(t *testing.T) {
	convey.Convey("test NewAdvancedRateLimiter basic info", t, func() {
		limiter := NewAdvancedRateLimiter(r, burst, maxWaitQueue)
		convey.So(limiter, convey.ShouldNotBeNil)
		convey.So(limiter.limiter.Limit(), convey.ShouldEqual, rate.Limit(r))
		convey.So(limiter.limiter.Burst(), convey.ShouldEqual, burst)
		convey.So(limiter.maxWaitQueue, convey.ShouldEqual, maxWaitQueue)
		convey.So(limiter.waitQueue, convey.ShouldEqual, 0)
	})
}

// TestAdvancedRateLimiterAllow test allow secen
func TestAdvancedRateLimiterAllow(t *testing.T) {
	convey.Convey("test AdvancedRateLimiter Allow function", t, func() {
		limiter := NewAdvancedRateLimiter(r, burst, maxWaitQueue)
		ctx := context.Background()

		convey.Convey("to test there is token", func() {
			for i := 0; i < burst; i++ {
				result := limiter.Allow(ctx)
				convey.So(result, convey.ShouldBeTrue)
			}
		})

		convey.Convey("to test there is no token, but queue is not full", func() {
			for i := 0; i < burst; i++ {
				limiter.Allow(ctx)
			}
			var wg sync.WaitGroup
			results := make(chan bool, r)
			for i := 0; i < r; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					results <- limiter.Allow(ctx)
				}()
			}
			wg.Wait()
			close(results)
			allowedCount := 0
			for result := range results {
				if result {
					allowedCount++
				}
			}
			convey.So(allowedCount, convey.ShouldBeGreaterThan, 0)
		})
	})
}

// TestQueueFull test queue full
func TestQueueFull(t *testing.T) {
	convey.Convey("to test queue is full", t, func() {
		ctx := context.Background()
		smallQueueLimiter := NewAdvancedRateLimiter(r, burst, smallQuelen)
		for i := 0; i < burst; i++ {
			smallQueueLimiter.Allow(ctx)
		}
		var wg sync.WaitGroup
		results := make(chan bool, resultLen)
		for i := 0; i < resultLen; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				results <- smallQueueLimiter.Allow(ctx)
			}()
		}
		wg.Wait()
		close(results)
		deniedCount := 0
		for result := range results {
			if !result {
				deniedCount++
			}
		}
		convey.So(deniedCount, convey.ShouldBeGreaterThan, 1)
	})
}
