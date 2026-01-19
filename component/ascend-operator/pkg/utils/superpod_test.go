/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// Package utils is common utils
package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestGetLogicSuperPodNodes(t *testing.T) {
	convey.Convey("TestGetLogicSuperPodNodes", t, func() {
		convey.Convey("01-when spBlock is smaller than chipsPerNode, should return 1", func() {
			spBlock, chipsPerNode := 1, 2
			res := getLogicSuperPodNodes(spBlock, chipsPerNode)
			convey.So(res, convey.ShouldEqual, 1)
		})
		convey.Convey("02-when spBlock is bigger than chipsPerNode, should return quotient", func() {
			expected := 2
			spBlock, chipsPerNode := 4, 2
			res := getLogicSuperPodNodes(spBlock, chipsPerNode)
			convey.So(res, convey.ShouldEqual, expected)
		})
	})
}

func TestGetLogicSuperPodId(t *testing.T) {
	convey.Convey("TestGetLogicSuperPodId", t, func() {
		convey.Convey("01-when spBlock is 0, should return 0", func() {
			res := GetLogicSuperPodId(1, 0, 0)
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("02-when pod rank is 1, spBlock is 1 and chipsPerNode is 1, should return 1", func() {
			res := GetLogicSuperPodId(1, 1, 1)
			convey.So(res, convey.ShouldEqual, 1)
		})
	})
}

func TestGetSpBlock(t *testing.T) {
	convey.Convey("TestGetSpBlock", t, func() {
		convey.Convey("01-job is nil will return 0", func() {
			res := GetSpBlock(nil)
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("02-job without sp-block annotation should return 0", func() {
			res := GetSpBlock(newCommonAscendJob())
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("03-job with invalid sp-block annotation should return 0", func() {
			job := newCommonAscendJob()
			job.Annotations[AnnoKeyOfSuperPod] = "xx"
			res := GetSpBlock(newCommonAscendJob())
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("04-job with valid sp-block annotation should return 1", func() {
			job := newCommonAscendJob()
			job.Annotations[AnnoKeyOfSuperPod] = "1"
			res := GetSpBlock(job)
			convey.So(res, convey.ShouldEqual, 1)
		})
	})
}
