// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestInitClientK8s(t *testing.T) {
	convey.Convey("test stop report", t, func() {
		gomonkey.ApplyFunc(newClientK8s, func() (*K8sClient, error) {
			return nil, nil
		})
		err := InitClientK8s()
		convey.So(err, convey.ShouldBeNil)
	})
}
