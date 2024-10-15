// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package util a series of util function
package util

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	fakeNs   = "test-ns"
	fakeName = "test-name"
)

// TestGetServiceIpWithRetry test case GetServiceIpWithRetry
func TestGetServiceIpWithRetry(t *testing.T) {
	convey.Convey("TestGetServiceIpWithRetry", t, func() {
		fakeClient := fake.NewSimpleClientset()
		convey.So(GetServiceIpWithRetry(fakeClient, fakeNs, fakeName), convey.ShouldEqual, "")
	})
}
