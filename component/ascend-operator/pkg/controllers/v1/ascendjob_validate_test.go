/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	_ "ascend-operator/pkg/testtool"
)

func TestValidateBasicInfo(t *testing.T) {
	convey.Convey("reconciler valid basic info", t, func() {
		rc := &ASJobReconciler{}
		job := newCommonAscendJob()
		convey.Convey("01-job ReplicaSpecs is nil, should return err", func() {
			err := rc.validateBasicInfo(job)
			convey.ShouldEqual(err, &validateError{
				reason:  "SpecsError",
				message: "job spec is not valid",
			})
		})
		convey.Convey("02-job SuccessPolicy is invalid, should return err", func() {
			fakeSuccessPolicy := mindxdlv1.SuccessPolicy("xxx")
			job.Spec.SuccessPolicy = &fakeSuccessPolicy
			err := rc.validateBasicInfo(job)
			convey.ShouldEqual(err, &validateError{
				reason:  "SuccessPolicyError",
				message: `job success policy is invalid, it must be one of <"", AllWorkers>`,
			})
		})
		convey.Convey("03-job queue is not exist, should return err", func() {
			rc.Config.EnableGangScheduling = true
			schedulingPolicy := &commonv1.SchedulingPolicy{
				Queue: "XXX",
			}
			job.Spec.RunPolicy.SchedulingPolicy = schedulingPolicy
			patch := gomonkey.ApplyMethod(new(ASJobReconciler), "Get",
				func(_ *ASJobReconciler, _ context.Context, _ client.ObjectKey, _ client.Object) error {
					return errors.New("not found")
				})
			defer patch.Reset()
			err := rc.validateBasicInfo(job)
			convey.ShouldEqual(err, &validateError{
				reason:  "QueueGetFailed",
				message: "not found",
			})
		})
	})
}

func TestValidateSpec(t *testing.T) {
	convey.Convey("reconciler valid job spec", t, func() {
		rc := &ASJobReconciler{}
		job := newCommonAscendJob()
		spec := make(map[commonv1.ReplicaType]*commonv1.ReplicaSpec)
		convey.Convey("01-job framework label is not set, should return err", func() {
			err := rc.validateSpec(job, spec)
			convey.ShouldEqual(err, &validateError{
				reason:  "FrameworkLabelError",
				message: "framework label  is not set",
			})
		})
		convey.Convey("02-job framework label is invalid, should return err", func() {
			job.Labels = map[string]string{mindxdlv1.FrameworkKey: "xxx"}
			err := rc.validateSpec(job, spec)
			convey.ShouldEqual(err, &validateError{
				reason:  "FrameworkLabelError",
				message: "framework<xxx> is not supported, must be one of <mindspore, pytorch, tensorflow>",
			})
		})
	})
}

func TestJobTotalRequest(t *testing.T) {
	convey.Convey("get ms job total require npu", t, func() {
		convey.Convey("01-job require 8 npu, should return 8", func() {
			replicas := int32(1)
			expectResult := 8
			specs := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
				mindxdlv1.MindSporeReplicaTypeScheduler: nil,
				mindxdlv1.ReplicaTypeWorker: {
					Replicas: &replicas,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: mindxdlv1.DefaultContainerName}},
						},
					},
				},
			}
			patch1 := gomonkey.ApplyFunc(getContainerResourceReq, func(_ corev1.Container) int {
				return expectResult
			})
			defer patch1.Reset()
			res := jobTotalRequest(specs)
			convey.ShouldEqual(res, expectResult)
		})
	})
}
