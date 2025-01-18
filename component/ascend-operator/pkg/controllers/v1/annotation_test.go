/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package v1

import (
	"math"
	"strconv"
	"testing"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	_ "ascend-operator/pkg/testtool"
)

// TestSetPodAnnotation is test for setPodAnnotation
func TestSetPodAnnotation(t *testing.T) {
	convey.Convey("set pod Annotation", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		podTemplate := &corev1.PodTemplateSpec{}
		convey.Convey("01-index is invalid, err is not nil", func() {
			err := rc.setPodAnnotation(job, podTemplate, "worker", "xxx")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-job has no chief、 master or scheduler, or job has scheduler without npu, "+
			"hccl/rankIndex should equal index", func() {
			err := rc.setPodAnnotation(job, podTemplate, "worker", "0")
			convey.So(err, convey.ShouldBeNil)
			convey.So(podTemplate.Annotations[rankIndexKey], convey.ShouldEqual, "0")
		})
		job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{mindxdlv1.ReplicaTypeWorker: nil}
		job.SetAnnotations(map[string]string{nonWorkerPodMountChipStatus: "true"})
		convey.Convey("03-job has chief or master, or job has scheduler with npu,"+
			" and rtype is master, hccl/rankIndex should equal index", func() {
			err := rc.setPodAnnotation(job, podTemplate, "master", "1")
			convey.So(err, convey.ShouldBeNil)
			convey.So(podTemplate.Annotations[rankIndexKey], convey.ShouldEqual, "1")
		})
		convey.Convey("04-job has chief or master, or job has scheduler with npu, "+
			"and rtype is worker, hccl/rankIndex should equal index + 1",
			func() {
				err := rc.setPodAnnotation(job, podTemplate, "worker", "1")
				convey.So(err, convey.ShouldBeNil)
				convey.So(podTemplate.Annotations[rankIndexKey], convey.ShouldEqual, "2")
			})
		convey.Convey("05-index is equal to MaxInt, err is not nil", func() {
			err := rc.setPodAnnotation(job, podTemplate, "worker", strconv.Itoa(math.MaxInt))
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}
