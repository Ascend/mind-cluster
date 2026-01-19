// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
)

func TestGetNode(t *testing.T) {
	testGetNodeFromIndexerSuccess(t)
	testGetNodeFromIndexerAndClientFail(t)
}

func testGetNodeFromIndexerSuccess(t *testing.T) {
	convey.Convey("When GetNodeFromIndexer succeeds", t, func() {
		node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1"}}

		patches := gomonkey.ApplyFunc(GetNodeFromIndexer, func(name string) (*v1.Node, error) {
			return node, nil
		})
		defer patches.Reset()

		result := GetNode("node1")
		convey.So(result, convey.ShouldEqual, node)
	})
}

func testGetNodeFromIndexerAndClientFail(t *testing.T) {
	convey.Convey("When both GetNodeFromIndexer and client fail", t, func() {
		patches := gomonkey.ApplyFunc(GetNodeFromIndexer, func(name string) (*v1.Node, error) {
			return nil, errors.New("indexer error")
		})
		defer patches.Reset()

		patches.ApplyFunc(hwlog.RunLog.Warnf, func(format string, args ...interface{}) {})
		patches.ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {})

		result := GetNode("node1")
		convey.So(result, convey.ShouldBeNil)
	})
}
