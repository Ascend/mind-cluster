// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package utils is common utils
package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-operator/pkg/api/v1"
)

// TestGetScaleOutTypeFromJob test case for GetScaleOutTypeFromJob
func TestGetScaleOutTypeFromJob(t *testing.T) {
	convey.Convey("TestGetScaleOutTypeFromJob", t, func() {
		convey.Convey("01-should return false when job is nil", func() {
			_, exist := GetScaleOutTypeFromJob(nil)
			convey.So(exist, convey.ShouldBeFalse)
		})
		convey.Convey("02-should return false when job labels is nil", func() {
			acJob := newCommonAscendJob()
			acJob.Labels = nil
			_, exist := GetScaleOutTypeFromJob(acJob)
			convey.So(exist, convey.ShouldBeFalse)
		})
		convey.Convey("03-should return true when scaleout-type=roce label exists", func() {
			acJob := newCommonAscendJob()
			acJob.Labels = map[string]string{
				v1.ScaleOutTypeLabel: "roCe",
			}
			res, exist := GetScaleOutTypeFromJob(acJob)
			convey.So(exist, convey.ShouldBeTrue)
			convey.So(res, convey.ShouldEqual, "ROCE")
		})
		convey.Convey("04-should return true when scaleout-type=uboe label exists", func() {
			acJob := newCommonAscendJob()
			acJob.Labels = map[string]string{
				v1.ScaleOutTypeLabel: "uboE",
			}
			res, exist := GetScaleOutTypeFromJob(acJob)
			convey.So(exist, convey.ShouldBeTrue)
			convey.So(res, convey.ShouldEqual, "UBOE")
		})
	})
}

// TestIsScaleOutTypeValid test case for IsScaleOutTypeValid
func TestIsScaleOutTypeValid(t *testing.T) {
	tests := []struct {
		scaleOutType string
		expected     bool
	}{
		{v1.ScaleOutTypeRoCE, true},
		{v1.ScaleOutTypeUBoE, true},
		{"other", false},
	}
	convey.Convey("TestIsScaleOutTypeValid", t, func() {
		for _, tt := range tests {
			res := IsScaleOutTypeValid(tt.scaleOutType)
			convey.So(res, convey.ShouldEqual, tt.expected)
		}
	})
}
