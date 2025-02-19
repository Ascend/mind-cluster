// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault test for public fault limiter util
package publicfault

import (
	"context"
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"golang.org/x/time/rate"

	"clusterd/pkg/domain/publicfault"
)

func TestLimiter(t *testing.T) {
	const testResource1 = "resource1"
	const testResource2 = "resource2"
	publicfault.PubFaultResource = []string{testResource1, testResource2}
	defer resetResourceCache()
	limiterMap.Store(testResource2, rate.NewLimiter(1, 1))
	convey.Convey("test func UpdateLimiter", t, func() {
		UpdateLimiter()
		mapLen := 0
		limiterMap.Range(func(key, value interface{}) bool {
			mapLen++
			return true
		})
		convey.So(mapLen, convey.ShouldEqual, len(publicfault.PubFaultResource))
	})

	convey.Convey("test func LimiterWaitByResource success", t, func() {
		err := LimiterWaitByResource(testResource1, context.Background())
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test func LimiterWaitByResource failed, resource does not exist", t, func() {
		err := LimiterWaitByResource("abc", context.Background())
		convey.So(err, convey.ShouldResemble, errors.New("resource limiter does not exist"))
	})
}

func resetResourceCache() {
	publicfault.PubFaultResource = nil
}
