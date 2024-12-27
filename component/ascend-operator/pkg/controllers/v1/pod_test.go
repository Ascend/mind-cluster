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
	"github.com/kubeflow/common/pkg/core"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	_ "ascend-operator/pkg/testtool"
)

func TestReconcilePods(t *testing.T) {
	convey.Convey("ReconcilePods", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		jobStatus := &commonv1.JobStatus{}
		var pods []*corev1.Pod
		rtype := mindxdlv1.ReplicaTypeWorker
		spec := &commonv1.ReplicaSpec{
			Replicas: defaultReplicas(),
			Template: corev1.PodTemplateSpec{},
		}
		replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
		convey.Convey("01-filter pod for replica type failed, should return err", func() {
			patch := gomonkey.ApplyFunc(core.FilterPodsForReplicaType,
				func(_ []*corev1.Pod, _ string) ([]*corev1.Pod, error) {
					return nil, errors.New("filter pod failed")
				})
			defer patch.Reset()
			err := rc.ReconcilePods(job, jobStatus, pods, rtype, spec, replicas)
			convey.ShouldEqual(err, errors.New("filter pod failed"))
		})
		convey.Convey("02-filter pod for replica type failed, should return err", func() {
			patch1 := gomonkey.ApplyMethod(new(common.JobController), "FilterPodsForReplicaType",
				func(_ *common.JobController, _ []*corev1.Pod, _ string) ([]*corev1.Pod, error) { return nil, nil })
			defer patch1.Reset()
			patch2 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "reconcilePods",
				func(_ *ASJobReconciler, _ specInfo, _ []*corev1.Pod, _ *commonv1.JobStatus,
					_ map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
					return errors.New("reconcile pods failed")
				})
			defer patch2.Reset()
			err := rc.ReconcilePods(job, jobStatus, pods, rtype, spec, replicas)
			convey.ShouldEqual(err, errors.New("reconcile pods failed"))
		})
	})
}

func TestReconcilePodNeedCreateOrDelete(t *testing.T) {
	convey.Convey("reconcilePods", t, func() {
		rc := newCommonReconciler()
		jobStatus := &commonv1.JobStatus{}
		si := &podInfo{
			job: newCommonAscendJob(),
			spec: &commonv1.ReplicaSpec{
				Replicas:      defaultReplicas(),
				Template:      corev1.PodTemplateSpec{},
				RestartPolicy: "",
			},
			status: &commonv1.ReplicaStatus{},
		}
		replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
		pods := []*corev1.Pod{{}}
		convey.Convey("02-need create pod, but failed, should return err", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "createNewPod", func(_ *ASJobReconciler,
				_ *mindxdlv1.AscendJob, _ podInfo, _ map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
				return errors.New("create pod failed")
			})
			defer patch.Reset()
			err := rc.reconcilePods(si, pods, jobStatus, replicas)
			convey.ShouldBeNil(err, errors.New("create pod failed"))
		})
		convey.Convey("03-need delete pod, but failed, should return err", func() {
			patch1 := gomonkey.ApplyMethod(new(ASJobReconciler), "GetPodSlices", func(_ *ASJobReconciler,
				_ []*corev1.Pod, _ int) [][]*corev1.Pod {
				return [][]*corev1.Pod{{{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{commonv1.ReplicaIndexLabel: "1"},
					},
				}}}
			})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyPrivateMethod(new(fakePodController), "DeletePod", func(_ *ASJobReconciler,
				_ string, _ string, _ runtime.Object) error {
				return errors.New("delete pod failed")
			})
			defer patch2.Reset()
			err := rc.reconcilePods(si, pods, jobStatus, replicas)
			convey.ShouldBeNil(err, errors.New("delete pod failed"))
		})
	})
}

func TestReconcilePodNotNeedCreateOrDelete(t *testing.T) {
	convey.Convey("reconcilePods", t, func() {
		rc := newCommonReconciler()
		jobStatus := &commonv1.JobStatus{}
		si := &podInfo{
			job: newCommonAscendJob(),
			spec: &commonv1.ReplicaSpec{
				Replicas:      defaultReplicas(),
				Template:      corev1.PodTemplateSpec{},
				RestartPolicy: "",
			},
			status: &commonv1.ReplicaStatus{},
		}
		replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
		pods := []*corev1.Pod{{}}
		convey.Convey("01-pod slice is empty, should do nothing and return nil", func() {
			patch := gomonkey.ApplyMethod(new(ASJobReconciler), "GetPodSlices", func(_ *ASJobReconciler,
				_ []*corev1.Pod, _ int) [][]*corev1.Pod {
				return nil
			})
			defer patch.Reset()
			err := rc.reconcilePods(si, pods, jobStatus, replicas)
			convey.ShouldBeNil(err)
		})
		convey.Convey("04-check pod status failed, should return err", func() {
			pods[0] = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{commonv1.ReplicaIndexLabel: "0"},
				},
			}
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "checkPodStatus", func(_ *ASJobReconciler,
				_ specInfo, _ *corev1.Pod, _ *commonv1.JobStatus) error {
				return errors.New("check pod status failed")
			})
			defer patch.Reset()
			err := rc.reconcilePods(si, pods, jobStatus, replicas)
			convey.ShouldBeNil(err, errors.New("check pod status failed"))
		})
	})
}

func TestGenRankTable(t *testing.T) {
	convey.Convey("genRankTable", t, func() {
		rc := newCommonReconciler()
		jobStatus := &commonv1.JobStatus{}
		si := &podInfo{
			job: newCommonAscendJob(),
			spec: &commonv1.ReplicaSpec{
				Replicas:      defaultReplicas(),
				Template:      corev1.PodTemplateSpec{},
				RestartPolicy: "",
			},
			status: &commonv1.ReplicaStatus{},
		}
		replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
		pods := []*corev1.Pod{{}}
		convey.Convey("01-need create pod, but failed, should return err", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "createNewPod", func(_ *ASJobReconciler,
				_ *mindxdlv1.AscendJob, _ podInfo, _ map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
				return errors.New("create pod failed")
			})
			defer patch.Reset()
			err := rc.reconcilePods(si, pods, jobStatus, replicas)
			convey.ShouldBeNil(err, errors.New("create pod failed"))
		})
		convey.Convey("02-need delete pod, but failed, should return err", func() {
			patch1 := gomonkey.ApplyMethod(new(ASJobReconciler), "GetPodSlices", func(_ *ASJobReconciler,
				_ []*corev1.Pod, _ int) [][]*corev1.Pod {
				return [][]*corev1.Pod{{{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{commonv1.ReplicaIndexLabel: "1"},
					},
				}}}
			})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyPrivateMethod(new(fakePodController), "DeletePod", func(_ *ASJobReconciler,
				_ string, _ string, _ runtime.Object) error {
				return errors.New("delete pod failed")
			})
			defer patch2.Reset()
			err := rc.reconcilePods(si, pods, jobStatus, replicas)
			convey.ShouldBeNil(err, errors.New("delete pod failed"))
		})
	})
}
