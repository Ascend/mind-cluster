// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package statistics test for statistic funcs about node
package statistics

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const (
	nodeSN     = "nodeSN"
	notExistSN = "not exist node sn"
	nodeName   = "nodeName"
)

func TestGetNodeNameBySN(t *testing.T) {
	nodeSNAndNameCache[nodeSN] = nodeName
	defer resetSNCache()
	convey.Convey("test func GetNodeNameBySN, sn exist", t, func() {
		name, exist := GetNodeNameBySN(nodeSN)
		convey.So(name, convey.ShouldEqual, nodeName)
		convey.So(exist, convey.ShouldBeTrue)
	})
	convey.Convey("test func GetNodeNameBySN, sn does not exist", t, func() {
		name, exist := GetNodeNameBySN(notExistSN)
		convey.So(name, convey.ShouldEqual, "")
		convey.So(exist, convey.ShouldBeFalse)
	})

	convey.Convey("test func GetNodeSNAndNameCache", t, func() {
		cache := GetNodeSNAndNameCache()
		convey.So(len(cache), convey.ShouldEqual, 1)
	})
}

func resetSNCache() {
	nodeSNAndNameCache = make(map[string]string)
}
