/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// Package v1 is using for reconcile AscendJob.
package v1

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"

	mindxdlv1 "ascend-operator/pkg/api/v1"
)

// TestValidateJobForA5 test case for ValidateJob in A5
func TestValidateJobForA5(t *testing.T) {
	convey.Convey("reconciler validate job for A5", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
		convey.Convey("01-valid basic info failed should return err when invalid scaleout-type label", func() {
			job.Labels = map[string]string{mindxdlv1.ScaleOutTypeLabel: "xx"}
			fakeErr := &validateError{
				reason: invalidScaleOutConfigReason,
				message: fmt.Sprintf("the value of label %s is invalid, which should be %s or %s",
					mindxdlv1.ScaleOutTypeLabel, mindxdlv1.ScaleOutTypeRoCE, mindxdlv1.ScaleOutTypeUBoE),
			}
			err := rc.validateJob(job)
			convey.So(err, convey.ShouldResemble, fakeErr)
		})
		convey.Convey("02-valid basic info success should return err when scaleout-type=roce label", func() {
			job.Labels = map[string]string{mindxdlv1.ScaleOutTypeLabel: "RoCe"}
			basicInfoPatch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "validateBasicInfo",
				func(_ *ASJobReconciler, _ *mindxdlv1.AscendJob) *validateError {
					return nil
				})
			defer basicInfoPatch.Reset()
			specPatch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "validateSpec",
				func(_ *ASJobReconciler, _ *mindxdlv1.AscendJob,
					_ map[commonv1.ReplicaType]*commonv1.ReplicaSpec) *validateError {
					return nil
				})
			defer specPatch.Reset()
			err := rc.validateJob(job)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("03-valid basic info success should return err when scaleout-type=uboe label", func() {
			job.Labels = map[string]string{mindxdlv1.ScaleOutTypeLabel: "uboE"}
			basicInfoPatch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "validateBasicInfo",
				func(_ *ASJobReconciler, _ *mindxdlv1.AscendJob) *validateError {
					return nil
				})
			defer basicInfoPatch.Reset()
			specPatch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "validateSpec",
				func(_ *ASJobReconciler, _ *mindxdlv1.AscendJob,
					_ map[commonv1.ReplicaType]*commonv1.ReplicaSpec) *validateError {
					return nil
				})
			defer specPatch.Reset()
			err := rc.validateJob(job)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
