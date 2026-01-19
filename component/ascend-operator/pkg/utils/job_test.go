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

func newCommonAscendJob() *v1.Job {
	return &v1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind: "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "ascendjob-test",
			UID:         "1111",
			Annotations: map[string]string{},
		},
		Spec: v1.JobSpec{},
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

// TestIsSoftStrategyJob test IsSoftStrategyJob
func TestIsSoftStrategyJob(t *testing.T) {
	convey.Convey("isSoftStrategyJob", t, func() {
		job := newCommonAscendJob()
		convey.Convey("01-job nil will return false", func() {
			res := IsSoftStrategyJob(nil)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("02-label nil will return false", func() {
			res := IsSoftStrategyJob(job)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("03-label SuperPodAffinity not exist will return false", func() {
			job.SetLabels(map[string]string{"otherLabel": ""})
			res := IsSoftStrategyJob(job)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("04-label SuperPodAffinity not equal SoftStrategy will return false", func() {
			job.SetLabels(map[string]string{SuperPodAffinity: "otherValue"})
			res := IsSoftStrategyJob(job)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("05-label SuperPodAffinity equal SoftStrategy will return true", func() {
			job.SetLabels(map[string]string{SuperPodAffinity: SoftStrategy})
			res := IsSoftStrategyJob(job)
			convey.So(res, convey.ShouldBeTrue)
		})
	})
}
