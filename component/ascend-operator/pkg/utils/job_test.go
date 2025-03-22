// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package utils is common utils
package utils

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-operator/pkg/api/v1"
)

func newCommonAscendJob() *v1.AscendJob {
	return &v1.AscendJob{
		TypeMeta: metav1.TypeMeta{
			Kind: "AscendJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "ascendjob-test",
			UID:         "1111",
			Annotations: map[string]string{},
		},
		Spec: v1.AscendJobSpec{},
	}
}

// TestIsMindIEEPJob test IsMindIEEPJob
func TestIsMindIEEPJob(t *testing.T) {
	convey.Convey("isMindIEEPJob", t, func() {
		job := newCommonAscendJob()
		convey.Convey("01-job nil will return false", func() {
			res := IsMindIEEPJob(nil)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("02-label nil will return false", func() {
			res := IsMindIEEPJob(job)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey(fmt.Sprintf("03-label %s not exist will return false", v1.JodIdLabelKey), func() {
			job.SetLabels(map[string]string{v1.AppLabelKey: ""})
			res := IsMindIEEPJob(job)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey(fmt.Sprintf("04-label %s not exist will return false", v1.AppLabelKey), func() {
			job.SetLabels(map[string]string{v1.JodIdLabelKey: ""})
			res := IsMindIEEPJob(job)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey(
			fmt.Sprintf("05-label %s and %s exist will return true", v1.JodIdLabelKey, v1.AppLabelKey),
			func() {
				job.SetLabels(map[string]string{v1.JodIdLabelKey: "", v1.AppLabelKey: ""})
				res := IsMindIEEPJob(job)
				convey.So(res, convey.ShouldBeTrue)
			})
	})
}
