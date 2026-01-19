/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile Job.
*/

package v1

import (
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/controller.v1/common"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	_ "ascend-operator/pkg/testtool"
)

// TestSetPodLabels test setPodLabels
func TestSetPodLabels(t *testing.T) {
	convey.Convey("setPodLabels", t, func() {
		job := newCommonAscendJob()
		rc := newCommonReconciler()
		rc.JobController = common.JobController{Controller: rc}
		podTemp := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			Containers: make([]corev1.Container, 1),
		}}
		rt := mindxdlv1.ReplicaTypeWorker
		rtype := strings.ToLower(string(rt))
		index := "1"
		convey.Convey("pod labels should set normal", func() {
			patch := gomonkey.ApplyMethod(new(ASJobReconciler), "GenLabels",
				func(_ *ASJobReconciler, _ string) map[string]string { return map[string]string{} })
			defer patch.Reset()
			rc.setPodLabels(job, podTemp, rt, index)
			convey.So(podTemp.Labels, convey.ShouldResemble, map[string]string{
				commonv1.OperatorNameLabel:           "job-controller",
				commonv1.GroupNameLabelDeprecated:    "mindxdl.gitee.com",
				commonv1.JobNameLabel:                "ascendjob-test",
				commonv1.JobNameLabelDeprecated:      "ascendjob-test",
				commonv1.ReplicaTypeLabel:            rtype,
				commonv1.ReplicaTypeLabelDeprecated:  rtype,
				commonv1.ReplicaIndexLabel:           index,
				commonv1.ReplicaIndexLabelDeprecated: index,
				"version":                            "0",
			})
		})
	})
}
