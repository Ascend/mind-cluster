// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault test for public fault limiter util
package publicfault

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"golang.org/x/time/rate"

	"clusterd/pkg/domain/publicfault"
)

func TestLimiter(t *testing.T) {
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
		err := LimitByResource(testResource1)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test func LimiterWaitByResource failed, resource does not exist", t, func() {
		invalidResource := "abc"
		err := LimitByResource(invalidResource)
		expErr := fmt.Errorf("resource <%s> limiter does not exist", invalidResource)
		convey.So(err, convey.ShouldResemble, expErr)
	})
	convey.Convey("test func LimiterWaitByResource failed, req exceeds the upper limit", t, func() {
		p1 := gomonkey.ApplyMethodReturn(&rate.Limiter{}, "Allow", false)
		defer p1.Reset()
		err := LimitByResource(testResource1)
		expErr := fmt.Errorf("request exceeds the upper limit, resource: <%s>", testResource1)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func resetResourceCache() {
	publicfault.PubFaultResource = nil
}
