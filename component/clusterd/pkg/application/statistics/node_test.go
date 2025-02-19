// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package statistics test for statistic funcs about node
package statistics

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/statistics"
)

const (
	nodeSN   = "nodeSN"
	nodeName = "nodeName"
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_xode = %v\n", code)
}

func setup() error {
	return initLog()
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return errors.New("init hwlog failed")
	}
	return nil
}

func TestUpdateNodeSNAndNameCache(t *testing.T) {
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nodeName,
			Annotations: map[string]string{nodeAnnotation: nodeSN},
		},
	}

	convey.Convey("test func UpdateNodeSNAndNameCache, node is nil", t, func() {
		UpdateNodeSNAndNameCache(nil, constant.AddOperator)
	})

	convey.Convey("test func UpdateNodeSNAndNameCache, add node when node does not exist", t, func() {
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 0)
		UpdateNodeSNAndNameCache(node, constant.AddOperator)
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 1)
	})

	convey.Convey("test func UpdateNodeSNAndNameCache, add node when node exist", t, func() {
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 1)
		UpdateNodeSNAndNameCache(node, constant.AddOperator)
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 1)
	})

	convey.Convey("test func UpdateNodeSNAndNameCache, delete node when node exist", t, func() {
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 1)
		UpdateNodeSNAndNameCache(node, constant.DeleteOperator)
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 0)
	})

	convey.Convey("test func UpdateNodeSNAndNameCache, delete node when node does not exist", t, func() {
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 0)
		UpdateNodeSNAndNameCache(node, constant.DeleteOperator)
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 0)
	})

	convey.Convey("test func UpdateNodeSNAndNameCache, invalid operator", t, func() {
		UpdateNodeSNAndNameCache(node, "invalid operator")
		convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 0)
	})

	convey.Convey("test func UpdateNodeSNAndNameCache, label does not exist", t, testLabelNotExist)
}

func testLabelNotExist() {
	node1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nodeName,
			Annotations: map[string]string{"label not exist": nodeSN},
		},
	}
	UpdateNodeSNAndNameCache(node1, constant.AddOperator)
	convey.So(len(statistics.GetNodeSNAndNameCache()), convey.ShouldEqual, 0)
}
