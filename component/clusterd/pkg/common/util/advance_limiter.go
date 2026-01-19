// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package util a series of util function
package util

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"ascend-common/common-utils/hwlog"
)

// AdvancedRateLimiter limiter with queue
type AdvancedRateLimiter struct {
	limiter      *rate.Limiter
	maxWaitQueue int
	waitQueue    int
	mu           sync.Mutex
}

// NewAdvancedRateLimiter create a new limiter, rate.limiter with upper waiting queue
func NewAdvancedRateLimiter(r, burst, maxWaitQueue int) *AdvancedRateLimiter {
	return &AdvancedRateLimiter{
		limiter:      rate.NewLimiter(rate.Limit(r), burst),
		maxWaitQueue: maxWaitQueue,
	}
}

// Allow to check whether requeset is allow
func (arl *AdvancedRateLimiter) Allow(ctx context.Context) bool {
	if arl.limiter.Allow() {
		return true
	}

	arl.mu.Lock()
	if arl.waitQueue >= arl.maxWaitQueue {
		arl.mu.Unlock()
		return false
	}
	arl.waitQueue++
	arl.mu.Unlock()

	defer func() {
		arl.mu.Lock()
		arl.waitQueue--
		arl.mu.Unlock()
	}()
	upperWait := time.Duration(arl.maxWaitQueue / arl.limiter.Burst())
	ctx, cancel := context.WithTimeout(ctx, upperWait*time.Second)
	defer cancel()

	if err := arl.limiter.Wait(ctx); err != nil {
		hwlog.RunLog.Errorf("failed to wait for token, err: %v", err)
		return false
	}
	return true
}
