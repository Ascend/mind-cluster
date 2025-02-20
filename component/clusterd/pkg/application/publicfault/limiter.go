// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault utils about limiter
package publicfault

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"

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

	const qpsLim = 100
	for _, resource := range publicfault.PubFaultResource {
		limiterMap.Store(resource, rate.NewLimiter(rate.Every(time.Second), qpsLim))
	}
}

// LimitByResource limiter work by resource
func LimitByResource(resource string) error {
	limiter, ok := getLimiterByResource(resource)
	if !ok {
		return fmt.Errorf("resource <%s> limiter does not exist", resource)
	}
	if !limiter.Allow() {
		return fmt.Errorf("request exceeds the upper limit, resource: <%s>", resource)
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
