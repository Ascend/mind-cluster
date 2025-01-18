/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package v1

import (
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	_ "ascend-operator/pkg/testtool"
)

// TestNewJobInfo01 test newJobInfo
func TestNewJobInfo01(t *testing.T) {
	convey.Convey("01-newJobInfo", t, func() {
		job := newCommonAscendJob()
		replicaTypes := make(map[commonv1.ReplicaType]*commonv1.ReplicaSpec)
		jobStatus := &commonv1.JobStatus{}
		runPolicy := &commonv1.RunPolicy{}
		rc := newCommonReconciler()
		convey.Convey("01-get job ref pods failed, should return err", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getPodsForJob",
				func(_ *ASJobReconciler, _ interface{}) ([]*corev1.Pod, error) {
					return nil, errors.New("not found pods")
				})
			defer patch1.Reset()
			_, err := rc.newJobInfo(job, replicaTypes, jobStatus, runPolicy)
			convey.So(err, convey.ShouldResemble, errors.New("not found pods"))
		})
		convey.Convey("02-get job ref pod and svc success, should return right job-info", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getPodsForJob",
				func(_ *ASJobReconciler, _ interface{}) ([]*corev1.Pod, error) { return nil, nil })
			defer patch1.Reset()
			expected := &jobInfo{
				job:           job,
				jobKey:        "ascendjob-test",
				name:          "ascendjob-test",
				rtObj:         interface{}(job).(runtime.Object),
				mtObj:         interface{}(job).(metav1.Object),
				pods:          nil,
				status:        jobStatus,
				runPolicy:     runPolicy,
				rpls:          replicaTypes,
				totalReplicas: 0,
			}
			ji, err := rc.newJobInfo(job, replicaTypes, jobStatus, runPolicy)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ji, convey.ShouldResemble, expected)
		})
	})
}

// TestNewJobInfo02 test newJobInfo
func TestNewJobInfo02(t *testing.T) {
	convey.Convey("02-newJobInfo", t, func() {
		job := newCommonAscendJob()
		rc := newCommonReconciler()
		replicaTypes := make(map[commonv1.ReplicaType]*commonv1.ReplicaSpec)
		jobStatus := &commonv1.JobStatus{}
		runPolicy := &commonv1.RunPolicy{}
		convey.Convey("04-job which is not Object should return err ", func() {
			_, err := rc.newJobInfo("job", replicaTypes, jobStatus, runPolicy)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("job<%v> is not of type metav1.Object", "job"))
		})
		convey.Convey("05-job which is not AscendJob should return err ", func() {
			_, err := rc.newJobInfo(&corev1.Pod{}, replicaTypes, jobStatus, runPolicy)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("job<%v> is not of type AscendJob", &corev1.Pod{}))
		})
		convey.Convey("06-job which is not AscendJob should return err ", func() {
			patch := gomonkey.ApplyFunc(cache.DeletionHandlingMetaNamespaceKeyFunc,
				func(_ interface{}) (string, error) { return "", errors.New("not found") })
			defer patch.Reset()
			_, err := rc.newJobInfo(job, replicaTypes, jobStatus, runPolicy)
			convey.So(err, convey.ShouldResemble, errors.New("not found"))
		})
	})
}

// TestGenLabels test genLabels
func TestGenLabels(t *testing.T) {
	convey.Convey("genLabels", t, func() {
		job := newCommonAscendJob()
		convey.Convey("01-job which is  not AscendJob should return err", func() {
			_, err := genLabels(&corev1.Pod{}, job.Name)
			convey.ShouldNotBeNil(err)
		})
		convey.Convey("02-job which is AscendJob should return right labels", func() {
			expected := map[string]string{commonv1.JobNameLabel: job.Name}
			label, err := genLabels(job, job.Name)
			convey.So(err, convey.ShouldBeNil)
			convey.So(label, convey.ShouldResemble, expected)
		})
		convey.Convey("03-job which is deploy should return right labels", func() {
			expected := map[string]string{deployLabelKey: job.Name}
			job.Kind = deployKind
			label, err := genLabels(job, job.Name)
			convey.So(err, convey.ShouldBeNil)
			convey.So(label, convey.ShouldResemble, expected)
		})
		convey.Convey("04-job which is vcjob should return right labels", func() {
			expected := map[string]string{vcjobLabelKey: job.Name}
			job.Kind = vcjobKind
			label, err := genLabels(job, job.Name)
			convey.So(err, convey.ShouldBeNil)
			convey.So(label, convey.ShouldResemble, expected)
		})
		convey.Convey("05-job which is ptjob should return right labels", func() {
			job.Kind = "ptjob"
			_, err := genLabels(job, job.Name)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestGetPodsForJob test getPodsForJob
func TestGetPodsForJob(t *testing.T) {
	convey.Convey("getPodsForJob", t, func() {
		job := newCommonAscendJob()
		rc := newCommonReconciler()
		convey.Convey("01-not Object should return error", func() {
			_, err := rc.getPodsForJob("job")
			convey.ShouldEqual(err, fmt.Errorf("job<%v> is not of type metav1.Object", job))
		})
		convey.Convey("02-gen labels failed should return error", func() {
			patch := gomonkey.ApplyFunc(genLabels, func(_ interface{}, _ string) (map[string]string, error) {
				return nil, errors.New("gen labels failed")
			})
			defer patch.Reset()
			_, err := rc.getPodsForJob(job)
			convey.So(err, convey.ShouldResemble, errors.New("gen labels failed"))
		})
		convey.Convey("03-get selector failed should return error", func() {
			patch := gomonkey.ApplyFunc(metav1.LabelSelectorAsSelector,
				func(_ *metav1.LabelSelector) (labels.Selector, error) {
					return nil, errors.New("get selector failed")
				})
			defer patch.Reset()
			_, err := rc.getPodsForJob(job)
			convey.So(err, convey.ShouldResemble, errors.New("couldn't convert Job selector: get selector failed"))
		})
		convey.Convey("04-get pods success should return nil", func() {
			_, err := rc.getPodsForJob(job)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestGetOrCreateSvc test getOrCreateSvc
func TestGetOrCreateSvc(t *testing.T) {
	convey.Convey("01-getOrCreateSvc", t, func() {
		job := newCommonAscendJob()
		rc := newCommonReconciler()
		job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
			mindxdlv1.PytorchReplicaTypeMaster: {},
		}
		convey.Convey("01-get svc from api-server success should return svc", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getSvcFromApiserver", func(_,
				_ string) (*corev1.Service, error) {
				return &corev1.Service{}, nil
			})
			defer patch.Reset()
			svc, err := rc.getOrCreateSvc(job)
			convey.So(err, convey.ShouldBeNil)
			convey.So(svc, convey.ShouldNotBeNil)
		})
		convey.Convey("02-get svc from api-server error occur should return error", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getSvcFromApiserver", func(_,
				_ string) (*corev1.Service, error) {
				return nil, errors.New("get svc failed")
			})
			defer patch.Reset()
			_, err := rc.getOrCreateSvc(job)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-gen svc failed should return error", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getSvcFromApiserver",
				func(_ *ASJobReconciler, _, _ string) (*corev1.Service, error) {
					return nil, &notFoundError{err: "get svc failed"}
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "genService", func(_ *ASJobReconciler,
				_ metav1.Object, _ commonv1.ReplicaType, _ *commonv1.ReplicaSpec) (*corev1.Service, error) {
				return nil, errors.New("gen svc failed")
			})
			defer patch2.Reset()
			_, err := rc.getOrCreateSvc(job)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestGetOrCreateSvc02 test getOrCreateSvc
func TestGetOrCreateSvc02(t *testing.T) {
	convey.Convey("02-getOrCreateSvc", t, func() {
		job := newCommonAscendJob()
		rc := newCommonReconciler()
		patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getSvcFromApiserver",
			func(_ *ASJobReconciler, _, _ string) (*corev1.Service, error) {
				return nil, &notFoundError{err: "get svc failed"}
			})
		defer patch1.Reset()
		patch2 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "genService", func(_ *ASJobReconciler,
			_ metav1.Object, _ commonv1.ReplicaType, _ *commonv1.ReplicaSpec) (*corev1.Service, error) {
			return &corev1.Service{}, nil
		})
		defer patch2.Reset()
		convey.Convey("04-create svc failed should return error", func() {
			patch3 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "createService", func(_ *ASJobReconciler,
				_ string, _ *corev1.Service) (*corev1.Service, error) {
				return nil, errors.New("create svc failed")
			})
			defer patch3.Reset()
			_, err := rc.getOrCreateSvc(job)
			convey.So(err, convey.ShouldResemble, errors.New("create svc failed"))
		})
		convey.Convey("05-create svc success should return svc", func() {
			patch3 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "createService", func(_ *ASJobReconciler,
				_ string, _ *corev1.Service) (*corev1.Service, error) {
				return &corev1.Service{}, nil
			})
			defer patch3.Reset()
			_, err := rc.getOrCreateSvc(job)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

type notFoundError struct {
	err string
}

func (ne *notFoundError) Error() string {
	return ne.err
}

func (ne *notFoundError) Status() metav1.Status {
	return metav1.Status{
		Reason: metav1.StatusReasonNotFound,
	}
}
