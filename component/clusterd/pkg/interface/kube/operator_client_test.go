// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"ascend-common/api/ascend-operator/client/clientset/versioned"
)

func TestInitOperatorClient(t *testing.T) {
	convey.Convey("Test Init operator Client", t, func() {
		convey.Convey("Should return error when BuildConfigFromFlags fails", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyFunc(clientcmd.BuildConfigFromFlags, func(_, _ string) (*rest.Config, error) {
				return nil, errors.New("config error")
			})
			client, err := InitOperatorClient()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(client, convey.ShouldBeNil)
		})
		convey.Convey("Should return error when NewForConfig fails", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyFunc(versioned.NewForConfig, func(_ *rest.Config) (*versioned.Clientset, error) {
				return nil, errors.New("client error")
			})
			client, err := InitOperatorClient()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(client, convey.ShouldBeNil)
		})
		convey.Convey("Should return a client when init success", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyFunc(clientcmd.BuildConfigFromFlags, func(_, _ string) (*rest.Config, error) {
				return nil, nil
			})
			patches.ApplyFunc(versioned.NewForConfig, func(_ *rest.Config) (*versioned.Clientset, error) {
				return nil, nil
			})
			client, err := InitOperatorClient()
			convey.So(err, convey.ShouldBeNil)
			convey.So(client, convey.ShouldNotBeNil)
		})
	})
}
