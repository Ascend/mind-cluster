/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package v1

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/controller.v1/common"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"

	_ "ascend-operator/pkg/testtool"
)

func TestNewJobInfo(t *testing.T) {
	convey.Convey("newJobInfo", t, func() {
		job := newCommonAscendJob()
		replicaTypes := make(map[commonv1.ReplicaType]*commonv1.ReplicaSpec)
		jobStatus := &commonv1.JobStatus{}
		runPolicy := &commonv1.RunPolicy{}
		rc := newCommonReconciler()
		convey.Convey("01-get job ref pods failed, should return err", func() {
			patch1 := gomonkey.ApplyMethod(new(common.JobController), "GetPodsForJob",
				func(_ *common.JobController, _ interface{}) ([]*corev1.Pod, error) {
					return nil, errors.New("not found pods")
				})
			defer patch1.Reset()
			_, err := rc.newJobInfo(job, replicaTypes, jobStatus, runPolicy)
			convey.ShouldEqual(err, errors.New("not found pods"))
		})
		convey.Convey("02-get job ref svc failed, should return err", func() {
			patch1 := gomonkey.ApplyMethod(new(common.JobController), "GetPodsForJob",
				func(_ *common.JobController, _ interface{}) ([]*corev1.Pod, error) { return nil, nil })
			defer patch1.Reset()
			_, err := rc.newJobInfo(job, replicaTypes, jobStatus, runPolicy)
			convey.ShouldEqual(err, errors.New("not found services"))
		})
		convey.Convey("03-get job ref pod and svc success, should return right GetPodsForJob", func() {
			patch1 := gomonkey.ApplyMethod(new(common.JobController), "GetPodsForJob",
				func(_ *common.JobController, _ interface{}) ([]*corev1.Pod, error) { return nil, nil })
			defer patch1.Reset()
			_, err := rc.newJobInfo(job, replicaTypes, jobStatus, runPolicy)
			convey.ShouldBeNil(err)
		})
	})
}
