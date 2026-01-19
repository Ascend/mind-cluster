// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package node test for statistic funcs about node
package node

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/node"
)

const (
	nodeSN         = "nodeSN"
	nodeName       = "nodeName"
	nodeAnnotation = "product-serial-number"
)

var nodeInfo *v1.Node

func TestUpdateNodeInfoCache(t *testing.T) {
	nodeInfo = &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nodeName,
			Annotations: map[string]string{nodeAnnotation: nodeSN},
		},
	}

	convey.Convey("test func UpdateNodeInfoCache, node is nil", t, testUpdateCache)
	convey.Convey("test func UpdateNodeInfoCache, add node when node does not exist", t, testUpdateCacheNotExistWhenAdd)
	convey.Convey("test func UpdateNodeInfoCache, add node when node exist", t, testUpdateCacheExistWhenAdd)
	convey.Convey("test func UpdateNodeInfoCache, delete node when node exist", t, testUpdateCacheExistWhenDel)
	convey.Convey("test func UpdateNodeInfoCache, delete node when node does not exist", t, testUpdateCacheNotExistWhenDel)
	convey.Convey("test func UpdateNodeInfoCache, invalid operator", t, testUpdateCacheInvalidOpera)
	convey.Convey("test func UpdateNodeInfoCache, label does not exist", t, testLabelNotExist)
}

func testUpdateCache() {
	UpdateNodeInfoCache(nil, nil, constant.AddOperator)
	name, exist := node.GetNodeNameBySN(nodeSN)
	convey.So(name, convey.ShouldEqual, "")
	convey.So(exist, convey.ShouldEqual, false)
}

func testUpdateCacheNotExistWhenAdd() {
	name, exist := node.GetNodeNameBySN(nodeSN)
	convey.So(name, convey.ShouldEqual, "")
	convey.So(exist, convey.ShouldEqual, false)
	UpdateNodeInfoCache(nil, nodeInfo, constant.AddOperator)
	name, exist = node.GetNodeNameBySN(nodeSN)
	convey.So(name, convey.ShouldEqual, nodeName)
	convey.So(exist, convey.ShouldEqual, true)
}

func testUpdateCacheExistWhenAdd() {
	name, exist := node.GetNodeNameBySN(nodeSN)
	convey.So(name, convey.ShouldEqual, nodeName)
	convey.So(exist, convey.ShouldEqual, true)
	UpdateNodeInfoCache(nil, nodeInfo, constant.AddOperator)
	name, exist = node.GetNodeNameBySN(nodeSN)
	convey.So(name, convey.ShouldEqual, nodeName)
	convey.So(exist, convey.ShouldEqual, true)
}

func testUpdateCacheExistWhenDel() {
	name, exist := node.GetNodeNameBySN(nodeSN)
	convey.So(name, convey.ShouldEqual, nodeName)
	convey.So(exist, convey.ShouldEqual, true)
	UpdateNodeInfoCache(nil, nodeInfo, constant.DeleteOperator)
	name, exist = node.GetNodeNameBySN(nodeSN)
	convey.So(name, convey.ShouldEqual, "")
	convey.So(exist, convey.ShouldEqual, false)
}

func testUpdateCacheNotExistWhenDel() {
	name, exist := node.GetNodeNameBySN(nodeSN)
	convey.So(name, convey.ShouldEqual, "")
	convey.So(exist, convey.ShouldEqual, false)
	UpdateNodeInfoCache(nil, nodeInfo, constant.DeleteOperator)
	name, exist = node.GetNodeNameBySN(nodeSN)
	convey.So(name, convey.ShouldEqual, "")
	convey.So(exist, convey.ShouldEqual, false)
}

func testUpdateCacheInvalidOpera() {
	UpdateNodeInfoCache(nil, nodeInfo, "invalid operator")
	name, exist := node.GetNodeNameBySN(nodeSN)
	convey.So(name, convey.ShouldEqual, "")
	convey.So(exist, convey.ShouldEqual, false)
}

func testLabelNotExist() {
	node1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nodeName,
			Annotations: map[string]string{"label not exist": nodeSN},
		},
	}
	UpdateNodeInfoCache(nil, node1, constant.AddOperator)
	name, exist := node.GetNodeNameBySN(nodeSN)
	convey.So(name, convey.ShouldEqual, "")
	convey.So(exist, convey.ShouldEqual, false)
}
