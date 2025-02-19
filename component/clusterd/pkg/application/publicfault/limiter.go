// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault utils about limiter
package publicfault

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/time/rate"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/domain/publicfault"
)

// map key: resource; value: limiter
var limiterMap sync.Map

// UpdateLimiter update limiterMap
func UpdateLimiter() {
	// map initialize
	limiterMap.Range(func(key, value interface{}) bool {
		limiterMap.Delete(key)
		return true
	})

	const (
		limit = 100
		burst = 200
	)
	for _, resource := range publicfault.PubFaultResource {
		limiterMap.Store(resource, rate.NewLimiter(limit, burst))
	}
}

// LimiterWaitByResource limiter wait by resource
func LimiterWaitByResource(resource string, ctx context.Context) error {
	limiter, ok := getLimiterByResource(resource)
	if !ok {
		hwlog.RunLog.Error("resource limiter does not exist")
		return errors.New("resource limiter does not exist")
	}
	if err := limiter.Wait(ctx); err != nil {
		hwlog.RunLog.Errorf("%s limiter wait failed, error: %v", resource, err)
		return fmt.Errorf("%s limiter wait failed", resource)
	}
	return nil
}

func getLimiterByResource(resource string) (*rate.Limiter, bool) {
	limiter, ok := limiterMap.Load(resource)
	if !ok {
		return &rate.Limiter{}, false
	}
	return limiter.(*rate.Limiter), true
}
