/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package v1

import (
	"testing"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	_ "ascend-operator/pkg/testtool"
)

func TestSetPodAnnotation(t *testing.T) {
	convey.Convey("set pod Annotation", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		podTemplate := &corev1.PodTemplateSpec{}
		convey.Convey("01-index is invalid, err is not nil", func() {
			err := rc.setPodAnnotation(job, podTemplate, "worker", "xxx")
			convey.ShouldNotBeNil(err)
		})
		convey.Convey("02-job has no chief or master, hccl/rankIndex should equal index", func() {
			err := rc.setPodAnnotation(job, podTemplate, "worker", "0")
			convey.ShouldBeNil(err)
			convey.ShouldEqual(podTemplate.Annotations, map[string]string{rankIndexKey: "0"})
		})
		convey.Convey("03-job has chief or master, and rtype is master, hccl/rankIndex should equal index", func() {
			job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{mindxdlv1.
				PytorchReplicaTypeMaster: nil}
			err := rc.setPodAnnotation(job, podTemplate, "worker", "1")
			convey.ShouldBeNil(err)
			convey.ShouldEqual(podTemplate.Annotations, map[string]string{rankIndexKey: "1"})
		})
		convey.Convey("04-job has chief or master, and rtype is worker, hccl/rankIndex should equal index + 1",
			func() {
				job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{mindxdlv1.
					PytorchReplicaTypeMaster: nil}
				err := rc.setPodAnnotation(job, podTemplate, "worker", "1")
				convey.ShouldBeNil(err)
				convey.ShouldEqual(podTemplate.Annotations, map[string]string{rankIndexKey: "2"})
			})
	})
}
