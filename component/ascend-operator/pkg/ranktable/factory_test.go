/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package ranktable is using for reconcile AscendJob.
*/
package ranktable

import (
	"ascend-common/api"
	"testing"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	v1 "ascend-operator/pkg/ranktable/v1"
	"ascend-operator/pkg/ranktable/v1dot2"
	_ "ascend-operator/pkg/testtool"
	"ascend-operator/pkg/utils"
)

func TestNewGenerator(t *testing.T) {
	convey.Convey("TestNewGenerator", t, func() {
		job := &mindxdlv1.AscendJob{}
		convey.Convey("01-job without sp-block annotation should return v1 ranktable", func() {
			generator := NewGenerator(job)
			_, ok := generator.(*v1.RankTable)
			convey.So(ok, convey.ShouldEqual, true)
		})
		convey.Convey("02-job with sp-block annotation should return v1.2 ranktable", func() {
			job.Annotations = map[string]string{utils.AnnoKeyOfSuperPod: "16"}
			generator := NewGenerator(job)
			_, ok := generator.(*v1dot2.RankTable)
			convey.So(ok, convey.ShouldEqual, true)
		})
		convey.Convey("03-job without schedule policy annotation and with accelerator-type A3"+
			" should return v1.2 ranktable", func() {
			selector := map[string]string{api.AcceleratorTypeKey: api.AcceleratorTypeModule910A3SuperPod}
			rpls := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{"": &commonv1.ReplicaSpec{
				Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{NodeSelector: selector}},
			}}
			job.Spec.ReplicaSpecs = rpls
			generator := NewGenerator(job)
			_, ok := generator.(*v1dot2.RankTable)
			convey.So(ok, convey.ShouldEqual, true)
		})
	})
}
