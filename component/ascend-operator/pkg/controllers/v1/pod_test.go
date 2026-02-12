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
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/util"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"ascend-common/api"
	mindxdlv1 "ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable"
	"ascend-operator/pkg/ranktable/common"
	"ascend-operator/pkg/ranktable/generator"
	"ascend-operator/pkg/ranktable/utils"
	_ "ascend-operator/pkg/testtool"
)

const (
	num2 = 2
	num1 = 1
	num6 = 6
)

// TestReconcilePods test ReconcilePods function
func TestReconcilePods(t *testing.T) {
	convey.Convey("ReconcilePods", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		job.Labels = make(map[string]string)
		jobStatus := &commonv1.JobStatus{}
		var pods []*corev1.Pod
		rtype := mindxdlv1.ReplicaTypeWorker
		spec := &commonv1.ReplicaSpec{
			Replicas: newReplicas(1),
			Template: corev1.PodTemplateSpec{},
		}
		replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
		convey.Convey("01-nil reconciler should return error", func() {
			var fakeReconciler *ASJobReconciler
			err := fakeReconciler.ReconcilePods(job, jobStatus, pods, rtype, spec, replicas)
			convey.So(err, convey.ShouldResemble, errors.New("nil pointer"))
		})
		convey.Convey("02-not ascendJob should return err", func() {
			pod := &corev1.Pod{}
			err := rc.ReconcilePods(pod, jobStatus, pods, rtype, spec, replicas)
			convey.So(err, convey.ShouldResemble, errors.New("<nil> is not a type of Job"))
		})
		convey.Convey("03-get job framework failed should return err", func() {
			err := rc.ReconcilePods(job, jobStatus, pods, rtype, spec, replicas)
			convey.So(err, convey.ShouldResemble, errors.New("framework label is not set, please set label framework as one of <pytorch,mindspore,tensorflow>"))
		})
	})
}

func TestReconcilePods02(t *testing.T) {
	convey.Convey("02-reconcilePods", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		job.Labels = make(map[string]string)
		jobStatus := &commonv1.JobStatus{}
		var pods []*corev1.Pod
		rtype := mindxdlv1.ReplicaTypeWorker
		spec := &commonv1.ReplicaSpec{
			Replicas: newReplicas(1),
			Template: corev1.PodTemplateSpec{},
		}
		replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
		job.Labels[mindxdlv1.FrameworkKey] = mindxdlv1.PytorchFrameworkName
		convey.Convey("04-create pod info failed should return err", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "newPodInfo", func(_ *ASJobReconciler,
				_ *mindxdlv1.AscendJob, _ commonv1.ReplicaType, _ *commonv1.ReplicaSpec, _ string) (*podInfo, error) {
				return nil, errors.New("create pod info failed")
			})
			defer patch.Reset()
			err := rc.ReconcilePods(job, jobStatus, pods, rtype, spec, replicas)
			convey.So(err, convey.ShouldResemble, errors.New("create pod info failed"))
		})
		convey.Convey("05-reconcile pod failed should return err", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "newPodInfo", func(_ *ASJobReconciler,
				_ *mindxdlv1.AscendJob, _ commonv1.ReplicaType, _ *commonv1.ReplicaSpec, _ string) (*podInfo, error) {
				return nil, nil
			})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "reconcilePods", func(_ *ASJobReconciler,
				_ *podInfo, _ []*corev1.Pod, _ *commonv1.JobStatus, _ map[commonv1.ReplicaType]*commonv1.
					ReplicaSpec) error {
				return errors.New("reconcile pod failed")
			})
			defer patch2.Reset()
			err := rc.ReconcilePods(job, jobStatus, pods, rtype, spec, replicas)
			convey.So(err, convey.ShouldResemble, errors.New("reconcile pod failed"))
		})
	})
}

func TestNewPodInfo01(t *testing.T) {
	convey.Convey("01-newPodInfo", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		rtype := mindxdlv1.ReplicaTypeWorker
		spec := newCommonSpec()
		framework := "pytorch"
		convey.Convey("01-getMngSvcIpAndPort failed should return err", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getMngSvcIpAndPort",
				func(_ *ASJobReconciler, _ *mindxdlv1.AscendJob, _ string, _ commonv1.ReplicaType) (string, string,
					error) {
					return "", "", errors.New("getMngSvcIpAndPort failed")
				})
			defer patch.Reset()
			_, err := rc.newPodInfo(job, rtype, spec, framework)
			convey.So(err, convey.ShouldResemble, errors.New("getMngSvcIpAndPort failed"))
		})
	})
}

func TestNewPodInfo02(t *testing.T) {
	convey.Convey("02-newPodInfo", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		rtype := mindxdlv1.ReplicaTypeWorker
		spec := newCommonSpec()
		framework := "pytorch"
		patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getMngSvcIpAndPort",
			func(_ *ASJobReconciler, _ *mindxdlv1.AscendJob, _ string, _ commonv1.ReplicaType) (string, string, error) {
				return "127.0.0.1", "2222", nil
			})
		defer patch.Reset()
		patch1 := gomonkey.ApplyFunc(getNpuReqInfoPerPod, func(job *mindxdlv1.AscendJob) (string, int) {
			return "", 1
		})
		defer patch1.Reset()
		convey.Convey("04-get all info success should return pod info", func() {
			job.Spec.ReplicaSpecs = make(map[commonv1.ReplicaType]*commonv1.ReplicaSpec)
			spec.Replicas = newReplicas(1)
			job.Spec.ReplicaSpecs[mindxdlv1.ReplicaTypeWorker] = spec
			patch3 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getClusterDSvcIp",
				func(_ *ASJobReconciler) string { return "127.0.0.1" })
			defer patch3.Reset()
			_, err := rc.newPodInfo(job, rtype, spec, framework)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestConfigmapExist(t *testing.T) {
	convey.Convey("configmapExist", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		rtg := ranktable.NewGenerator(job)
		convey.Convey("01-configmap not exist should return false", func() {
			rtg.SetConfigmapExist(utils.ConfigmapExsit)
			exist := rc.configmapExist(rtg, job.Namespace, job.Name)
			convey.So(exist, convey.ShouldEqual, true)
		})
		convey.Convey("02-configmap not exist should return false", func() {
			rtg.SetConfigmapExist(utils.ConfigmapNotExist)
			exist := rc.configmapExist(rtg, job.Namespace, job.Name)
			convey.So(exist, convey.ShouldEqual, false)
		})
		rtg.SetConfigmapExist("")
		convey.Convey("03-configmap exist should return true", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getConfigmapFromApiserver",
				func(_ *ASJobReconciler, _, _ string) (*corev1.ConfigMap, error) {
					return nil, errors.New("get configmap from apiserver failed")
				})
			defer patch1.Reset()
			exist := rc.configmapExist(rtg, job.Namespace, job.Name)
			convey.So(exist, convey.ShouldEqual, false)
		})
		convey.Convey("04-configmap exist should return true", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getConfigmapFromApiserver",
				func(_ *ASJobReconciler, _, _ string) (*corev1.ConfigMap, error) { return nil, nil })
			defer patch1.Reset()
			exist := rc.configmapExist(rtg, job.Namespace, job.Name)
			convey.So(exist, convey.ShouldEqual, true)
		})
	})
}

// TestReconcilePodNeedCreateOrDelete test ReconcilePodNeedCreateOrDelete
func TestReconcilePodNeedCreateOrDelete(t *testing.T) {
	convey.Convey("reconcilePods", t, func() {
		rc := newCommonReconciler()
		jobStatus := &commonv1.JobStatus{}
		si := &podInfo{
			job: newCommonAscendJob(),
			spec: &commonv1.ReplicaSpec{
				Replicas:      newReplicas(1),
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
			convey.So(err, convey.ShouldResemble, errors.New("failed to create pods of type<>: create pod failed"))
		})
		convey.Convey("03-need delete pod and return nil", func() {
			patch1 := gomonkey.ApplyMethod(new(ASJobReconciler), "GetPodSlices", func(_ *ASJobReconciler,
				_ []*corev1.Pod, _ int) [][]*corev1.Pod {
				return [][]*corev1.Pod{{{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{commonv1.ReplicaIndexLabel: "1"},
					},
				}}}
			})
			defer patch1.Reset()
			err := rc.reconcilePods(si, pods, jobStatus, replicas)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestCheckPodStatus test checkPodStatus
func TestCheckPodStatus(t *testing.T) {
	convey.Convey("checkPodStatus", t, func() {
		rc := newCommonReconciler()
		pod := &corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: make([]corev1.ContainerStatus, 1)}}
		containerStatus := corev1.ContainerStatus{
			Name:  "ascend",
			State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 1}},
		}
		pi := &podInfo{spec: newCommonSpec(), job: newCommonAscendJob()}
		convey.Convey("01-restart policy is not ExitCode should return nil", func() {
			pi.spec.RestartPolicy = commonv1.RestartPolicyAlways
			err := rc.checkPodStatus(pi, pod, nil)
			convey.So(err, convey.ShouldBeNil)
		})
		pi.spec.RestartPolicy = commonv1.RestartPolicyExitCode
		convey.Convey("02-restart policy is ExitCode, but pod is not failed, should return nil", func() {
			pod.Status.Phase = corev1.PodRunning
			err := rc.checkPodStatus(pi, pod, nil)
			convey.So(err, convey.ShouldBeNil)
		})
		pod.Status.Phase = corev1.PodFailed
		convey.Convey("03-restart policy is ExitCode and pod is failed, "+
			"but exit code is not retryable should return nil", func() {
			pod.Status.ContainerStatuses[0] = containerStatus
			err := rc.checkPodStatus(pi, pod, nil)
			convey.So(err, convey.ShouldBeNil)
		})
		containerStatus.State.Terminated.ExitCode = 128
		convey.Convey("04-update job condition failed should return err", func() {
			patch := gomonkey.ApplyFunc(util.UpdateJobConditions, func(_ *commonv1.JobStatus,
				_ commonv1.JobConditionType, _, _ string) error {
				return errors.New("update job condition failed")
			})
			defer patch.Reset()
			err := rc.checkPodStatus(pi, pod, nil)
			convey.So(err, convey.ShouldResemble, errors.New("update job condition failed"))
		})
		convey.Convey("04-update job condition success should return nil", func() {
			patch := gomonkey.ApplyFunc(util.UpdateJobConditions, func(_ *commonv1.JobStatus,
				_ commonv1.JobConditionType, _, _ string) error {
				return nil
			})
			defer patch.Reset()
			err := rc.checkPodStatus(pi, pod, nil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestCreateNewPod test create new pod
func TestCreateNewPod(t *testing.T) {
	convey.Convey("createNewPod", t, func() {
		rc := newCommonReconciler()
		pi := &podInfo{spec: newCommonSpec(), job: newCommonAscendJob()}
		replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
		convey.Convey("01-nil reconciler should return err", func() {
			var r *ASJobReconciler
			err := r.createNewPod(pi, replicas)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-create pod spec failed should return err", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "createPodSpec", func(_ *ASJobReconciler,
				_ *podInfo, _ *commonv1.ReplicaSpec) (*corev1.PodTemplateSpec, error) {
				return nil, errors.New("create pod spec failed")
			})
			defer patch.Reset()
			err := rc.createNewPod(pi, replicas)
			convey.So(err, convey.ShouldResemble, errors.New("create pod spec failed"))
		})
		convey.Convey("03-create pod spec success should return nil", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "createPodSpec", func(_ *ASJobReconciler,
				_ *podInfo, _ *commonv1.ReplicaSpec) (*corev1.PodTemplateSpec, error) {
				return &corev1.PodTemplateSpec{}, nil
			})
			defer patch.Reset()
			err := rc.createNewPod(pi, replicas)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestCreatePodSpec test create pod spec
func TestCreatePodSpec(t *testing.T) {
	convey.Convey("createPodSpec", t, func() {
		rc := newCommonReconciler()
		pi := &podInfo{
			spec:  newCommonSpec(),
			job:   newCommonAscendJob(),
			rtype: mindxdlv1.ReplicaTypeWorker,
			index: math.MaxInt,
		}
		replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
		pi.job.Spec.ReplicaSpecs = replicas
		pi.job.Annotations = map[string]string{
			nonWorkerPodMountChipStatus: "true",
		}
		convey.Convey("01-nil ReplicaSpecs should return err", func() {
			_, err := rc.createPodSpec(pi, replicas)
			convey.ShouldEqual(err.Error(), fmt.Errorf("job or job specs is nil"))
		})

		convey.Convey("02-max int index should return err", func() {
			_, err := rc.createPodSpec(pi, replicas)
			convey.So(err, convey.ShouldResemble, errors.New("rank is the max int"))
		})
		pi.index = 0
		convey.Convey("03-set env failed should return error", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "setEnv", func(_ *ASJobReconciler,
				_ *podInfo, _ *corev1.PodTemplateSpec) error {
				return errors.New("set env failed")
			})
			defer patch1.Reset()
			_, err := rc.createPodSpec(pi, replicas)
			convey.So(err, convey.ShouldResemble, errors.New("set env failed"))
		})
		patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "setEnv", func(_ *ASJobReconciler,
			_ *podInfo, _ *corev1.PodTemplateSpec) error {
			return nil
		})
		defer patch1.Reset()
		convey.Convey("04-set set pod annotation failed should return error", func() {
			_, err := rc.createPodSpec(pi, replicas)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestReconcilePodNotNeedCreateOrDelete test reconcile pods
func TestReconcilePodNotNeedCreateOrDelete(t *testing.T) {
	convey.Convey("reconcilePods", t, func() {
		rc := newCommonReconciler()
		jobStatus := &commonv1.JobStatus{}
		si := &podInfo{
			job: newCommonAscendJob(),
			spec: &commonv1.ReplicaSpec{
				Replicas:      newReplicas(1),
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
			convey.So(err, convey.ShouldBeNil)
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
			convey.So(err, convey.ShouldResemble, errors.New("check pod status failed"))
		})
	})
}

func newReplicaSpecs() map[commonv1.ReplicaType]*commonv1.ReplicaSpec {
	var replicas int32 = 1
	return map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
		"test": {Replicas: &replicas},
	}
}

func newJobInfo(job interface{}) *jobInfo {
	metaObj, ok := job.(metav1.Object)
	if !ok {
		return nil
	}
	ascendJob, ok := job.(*mindxdlv1.AscendJob)
	if !ok {
		return nil
	}
	ascendJob.Spec.ReplicaSpecs = newReplicaSpecs()
	return &jobInfo{
		job:    job,
		jobKey: "",
		name:   "test",
		rtObj:  nil,
		mtObj:  metaObj,
		pods: []*corev1.Pod{
			&corev1.Pod{},
		},
		status:        &ascendJob.Status,
		runPolicy:     &ascendJob.Spec.RunPolicy,
		rpls:          ascendJob.Spec.ReplicaSpecs,
		totalReplicas: getTotalReplicas(ascendJob),
	}
}

// TestGenRankTable test the function of genRankTable
func TestGenRankTable(t *testing.T) {
	type args struct {
		ji *jobInfo
	}
	const testUid types.UID = "for_test_only"
	acjob := newCommonAscendJob()
	acjob.UID = testUid
	rtg := ranktable.NewGenerator(acjob)
	tests := []struct {
		name string
		args args
	}{
		{name: "01-base func test, should change status",
			args: args{ji: newJobInfo(acjob)}},
	}
	r := newCommonReconciler()
	r.rtGenerators = map[types.UID]generator.RankTableGenerator{testUid: rtg}
	patch1 := gomonkey.ApplyPrivateMethod(r, "saveRankTable", func(r *ASJobReconciler,
		rtg generator.RankTableGenerator,
		jobName, namespace string, uid types.UID) {
		return
	})
	defer patch1.Reset()
	patch2 := gomonkey.ApplyFunc(utils.PodHasAllocated, func(pod *corev1.Pod) bool {
		return true
	})
	defer patch2.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.genRankTable(tt.args.ji)
			if rtg.GetStatus() != utils.CompletedRTStatus {
				t.Errorf("update rank table failed")
			}
		})
	}
}

// TestUpdateRandIndex test the function of updateRandIndex
func TestUpdateRandIndex(t *testing.T) {
	type args struct {
		allocatedPods []*corev1.Pod
	}
	const testUid types.UID = "for_test_only"
	acjob := newCommonAscendJob()
	acjob.UID = testUid
	rtg := ranktable.NewGenerator(acjob)
	emptyPod := corev1.Pod{}
	annotatedPod := corev1.Pod{}
	annotatedPod.Annotations = map[string]string{"test": "test"}
	tests := []struct {
		name string
		args args
	}{
		{name: "01-base function test,should change rank table",
			args: args{allocatedPods: []*corev1.Pod{
				&emptyPod,
				&annotatedPod,
			}}},
	}
	r := newCommonReconciler()
	r.rtGenerators = map[types.UID]generator.RankTableGenerator{testUid: rtg}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.updateRandIndex(tt.args.allocatedPods)
			if emptyPod.Annotations == nil {
				t.Errorf("fail to create annotation for empty pod")
			}
			if _, ok := annotatedPod.Annotations[api.PodRankIndexAnno]; !ok {
				t.Errorf("fail to write rank index for annotated pod")
			}
		})
	}
}

// TestCheckPodDelete test the function of checkPodDelete
func TestCheckPodDelete(t *testing.T) {
	type args struct {
		rtg generator.RankTableGenerator
		ji  *jobInfo
	}
	const testUid types.UID = "for_test_only"
	const minReplicas int32 = -1
	const maxReplicas int32 = 10
	acjob := newCommonAscendJob()
	acjob.UID = testUid
	rtg := ranktable.NewGenerator(acjob)
	rtg.SetStatus(utils.CompletedRTStatus)
	jiNotDelete := newJobInfo(acjob)
	jiDelete := newJobInfo(acjob)
	jiNotDelete.totalReplicas = minReplicas // not trigger pod delete
	jiDelete.totalReplicas = maxReplicas    // trigger pod delete
	tests := []struct {
		name       string
		args       args
		wantStatus utils.RankTableStatus
	}{
		{
			name:       "01-not trigger pod delete, should not change status",
			args:       args{rtg: rtg, ji: jiNotDelete},
			wantStatus: utils.CompletedRTStatus,
		},
		{
			name:       "02-trigger pod delete, should change status",
			args:       args{rtg: rtg, ji: jiDelete},
			wantStatus: utils.InitialRTStatus,
		},
	}
	r := newCommonReconciler()
	patch1 := gomonkey.ApplyPrivateMethod(r, "saveRankTable",
		func(r *ASJobReconciler, rtg generator.RankTableGenerator,
			jobName, namespace string, uid types.UID) {
			return
		})
	defer patch1.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.checkPodDelete(tt.args.rtg, tt.args.ji)
			if rtg.GetStatus() != tt.wantStatus {
				t.Errorf("fail to check pod delete, status wrong")
			}
		})
	}
}

// TestSaveRankTable test the function of saveRankTable
func TestSaveRankTable(t *testing.T) {
	type args struct {
		rtg       generator.RankTableGenerator
		jobName   string
		namespace string
		uid       types.UID
	}
	const testUid types.UID = "for_test_only"
	acjob := newCommonAscendJob()
	acjob.UID = testUid
	rtg := ranktable.NewGenerator(acjob)

	tests := []struct {
		name string
		args args
	}{
		{
			name: "01-base function test, should change timestamp",
			args: args{
				rtg:       rtg,
				jobName:   "",
				namespace: "",
				uid:       testUid,
			},
		},
	}
	r := newCommonReconciler()
	patch1 := gomonkey.ApplyPrivateMethod(r, "saveRankTableFile",
		func(r *ASJobReconciler, rtg generator.RankTableGenerator) {
			return
		})
	defer patch1.Reset()
	patch2 := gomonkey.ApplyPrivateMethod(r, "saveRankTableConfigmap",
		func(r *ASJobReconciler, rtg generator.RankTableGenerator,
			jobName, namespace string, uid types.UID) {
			return
		})
	defer patch2.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.saveRankTable(tt.args.rtg, tt.args.jobName, tt.args.namespace, tt.args.uid)
			if rtg.GetTimeStamp() <= 0 {
				t.Errorf("saveRankTable failed, timestamp not change")
			}
		})
	}
}

// TestSaveRankTableFile test the function of saveRankTableFile
func TestSaveRankTableFile(t *testing.T) {
	type args struct {
		rtg generator.RankTableGenerator
	}
	const testUid types.UID = "for_test_only"
	acjob := newCommonAscendJob()
	acjob.UID = testUid
	rtgA := ranktable.NewGenerator(acjob) // set status same to trigger condition check
	rtgA.SetFileStatus(utils.CompletedRTStatus)
	rtgA.SetStatus(utils.CompletedRTStatus)
	rtgB := ranktable.NewGenerator(acjob) // set status different to trigger condition check
	rtgB.SetFileStatus(utils.CompletedRTStatus)
	rtgB.SetStatus(utils.InitialRTStatus)
	tests := []struct {
		name       string
		args       args
		wantSattus utils.RankTableStatus
	}{
		{
			name:       "01-status not same, file status should not change",
			args:       args{rtg: rtgA},
			wantSattus: utils.CompletedRTStatus,
		},
		{
			name:       "02-status same, write file fail,status should be unknown",
			args:       args{rtg: rtgB},
			wantSattus: utils.UnknownStatus,
		},
	}
	r := newCommonReconciler()
	patch := gomonkey.ApplyMethodFunc(rtgA, "WriteToFile",
		func() error {
			return errors.New("write failed")
		})
	defer patch.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.saveRankTableFile(tt.args.rtg)
			if tt.args.rtg.GetFileStatus() != tt.wantSattus {
				t.Errorf("saveRankTableFile status wrong: %v, expected: %v", tt.args.rtg.GetFileStatus(),
					tt.wantSattus)
			}
		})
	}
}

// TestSaveRankTableFilePatch1 test the function of saveRankTableFile
func TestSaveRankTableFilePatch1(t *testing.T) {
	type args struct {
		rtg generator.RankTableGenerator
	}
	const testUid types.UID = "for_test_only"
	acjob := newCommonAscendJob()
	acjob.UID = testUid
	rtgB := ranktable.NewGenerator(acjob) // set status different to trigger condition check
	rtgB.SetFileStatus(utils.CompletedRTStatus)
	rtgB.SetStatus(utils.InitialRTStatus)
	tests := []struct {
		name       string
		args       args
		wantSattus utils.RankTableStatus
	}{
		{
			name:       "03-status same, write file success, file status should be initial",
			args:       args{rtg: rtgB},
			wantSattus: utils.InitialRTStatus,
		},
	}
	r := newCommonReconciler()
	patch := gomonkey.ApplyMethodFunc(rtgB, "WriteToFile",
		func() error {
			return nil
		})
	defer patch.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.saveRankTableFile(tt.args.rtg)
			if tt.args.rtg.GetFileStatus() != tt.wantSattus {
				t.Errorf("saveRankTableFile status wrong: %v, expected: %v", tt.args.rtg.GetFileStatus(),
					tt.wantSattus)
			}
		})
	}
}

// TestSaveRankTableConfigmap test the function of saveRankTableConfigmap
func TestSaveRankTableConfigmap(t *testing.T) {
	convey.Convey("save rank table cm and change status", t, func() {
		const testUid types.UID = "for_test_only"
		acjob := newCommonAscendJob()
		acjob.UID = testUid
		rtg := ranktable.NewGenerator(acjob)
		r := newCommonReconciler()
		convey.Convey("01-configmap not exist, status not change", func() {
			patch := gomonkey.ApplyPrivateMethod(r, "configmapExist",
				func(_ *ASJobReconciler, rtg generator.RankTableGenerator, jobName, namespace string) bool {
					return false
				})
			defer patch.Reset()
			rtg.SetConfigmapStatus(utils.UnknownStatus)
			r.saveRankTableConfigmap(rtg, "", "", testUid)
			convey.So(rtg.GetConfigmapStatus(), convey.ShouldEqual, utils.UnknownStatus)
		})
		convey.Convey("02-configmap status same, should not change", func() {
			patch1 := gomonkey.ApplyPrivateMethod(r, "configmapExist",
				func(_ *ASJobReconciler, rtg generator.RankTableGenerator, jobName, namespace string) bool {
					return true
				})
			defer patch1.Reset()
			rtg.SetStatus(utils.InitialRTStatus)
			rtg.SetConfigmapStatus(utils.InitialRTStatus)
			r.saveRankTableConfigmap(rtg, "", "", testUid)
			convey.ShouldEqual(rtg.GetConfigmapStatus(), utils.InitialRTStatus)
		})
		convey.Convey("03-write cm fail, status should be unknown", func() {
			patch1 := gomonkey.ApplyPrivateMethod(r, "configmapExist",
				func(_ *ASJobReconciler, rtg generator.RankTableGenerator, jobName, namespace string) bool {
					return true
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyPrivateMethod(r, "tryWriteCm",
				func(_ *ASJobReconciler, jobName, namespace string, uid types.UID) error {
					return errors.New("write cm failed")
				})
			defer patch2.Reset()
			rtg.SetStatus(utils.InitialRTStatus)
			rtg.SetConfigmapStatus(utils.CompletedRTStatus)
			r.saveRankTableConfigmap(rtg, "", "", testUid)
			convey.So(rtg.GetConfigmapStatus(), convey.ShouldEqual, utils.UnknownStatus)
		})
	})
}

// TestSaveRankTableConfigmapPatch1 test the function of saveRankTableConfigmap
func TestSaveRankTableConfigmapPatch1(t *testing.T) {
	convey.Convey("save rank table cm and change status", t, func() {
		const testUid types.UID = "for_test_only"
		acjob := newCommonAscendJob()
		acjob.UID = testUid
		rtg := ranktable.NewGenerator(acjob)
		r := newCommonReconciler()
		convey.Convey("04-write cm success, status should change", func() {
			patch1 := gomonkey.ApplyPrivateMethod(r, "configmapExist",
				func(_ *ASJobReconciler, rtg generator.RankTableGenerator, jobName, namespace string) bool {
					return true
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyPrivateMethod(r, "tryWriteCm",
				func(_ *ASJobReconciler, jobName, namespace string, uid types.UID) error {
					return nil
				})
			defer patch2.Reset()
			rtg.SetStatus(utils.InitialRTStatus)
			rtg.SetConfigmapStatus(utils.CompletedRTStatus)
			r.saveRankTableConfigmap(rtg, "", "", testUid)
			convey.So(rtg.GetConfigmapStatus(), convey.ShouldEqual, utils.InitialRTStatus)
		})
	})
}

// TestTryWriteCm test the function of tryWriteCm
func TestTryWriteCm(t *testing.T) {
	convey.Convey("try write cm multiple times", t, func() {
		const testUid types.UID = "for_test_only"
		r := newCommonReconciler()
		convey.Convey("01-write fail, should return err", func() {
			patch := gomonkey.ApplyPrivateMethod(r, "writeRanktableToCm",
				func(_ *ASJobReconciler, jobName, namespace string, uid types.UID) error {
					return errors.New("write cm failed")
				})
			defer patch.Reset()
			err := r.tryWriteCm("", "", testUid)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-write success, should return nil", func() {
			patch := gomonkey.ApplyPrivateMethod(r, "writeRanktableToCm",
				func(_ *ASJobReconciler, jobName, namespace string, uid types.UID) error {
					return nil
				})
			defer patch.Reset()
			err := r.tryWriteCm("", "", testUid)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestSetEnv test the function of setEnv
func TestSetEnv(t *testing.T) {
	convey.Convey("set env", t, func() {
		rc := newCommonReconciler()
		podTemplate := &corev1.PodTemplateSpec{}
		job := newCommonAscendJob()
		pi := &podInfo{
			frame: mindxdlv1.MindSporeFrameworkName,
			job:   job,
			rtype: mindxdlv1.ReplicaTypeWorker,
			ctReq: 1,
		}
		convey.Convey("01-job with frame of mindspore and 1 replicas need not set env", func() {
			job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
				mindxdlv1.ReplicaTypeWorker: {},
			}
			err := rc.setEnv(pi, podTemplate)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("02-pod with invalid frame should return error", func() {
			pi.frame = "fake-frame"
			err := rc.setEnv(pi, podTemplate)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("frameworke<%s> is not support", pi.frame))
		})
		convey.Convey("03-pod without requiring npu should return nil", func() {
			pi.ctReq = 0
			err := rc.setEnv(pi, podTemplate)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestSetGangScheduleInfo test the function of setGangScheduleInfo
func TestSetGangScheduleInfo(t *testing.T) {
	convey.Convey("set gang schedule info", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		podTemplate := &corev1.PodTemplateSpec{}
		replicas := make(map[commonv1.ReplicaType]*commonv1.ReplicaSpec)
		rt := string(mindxdlv1.ReplicaTypeWorker)
		convey.Convey("01-volcano will be set when job and pod scheduler name is not set ", func() {
			rc.setGangScheduleInfo(job, podTemplate, replicas, rt)
			convey.So(podTemplate.Spec.SchedulerName, convey.ShouldEqual, "volcano")
		})
		job.Spec.SchedulerName = "job-scheduler"
		convey.Convey("02-scheduler name will be set when job scheduler name is set and pod scheduler name is not"+
			" set", func() {
			rc.setGangScheduleInfo(job, podTemplate, replicas, rt)
			convey.So(podTemplate.Spec.SchedulerName, convey.ShouldEqual, "job-scheduler")
		})
		convey.Convey("03-scheduler name will not be overwritten when job and pod scheduler name is set", func() {
			podTemplate.Spec.SchedulerName = "pod-scheduler"
			replicas[mindxdlv1.ReplicaTypeWorker] = &commonv1.ReplicaSpec{
				Template: *podTemplate,
			}
			rc.setGangScheduleInfo(job, podTemplate, replicas, rt)
			convey.So(podTemplate.Spec.SchedulerName, convey.ShouldEqual, "pod-scheduler")
			pgName := job.GetName() + "-" + string(job.GetUID())
			convey.So(podTemplate.Annotations[gangSchedulingPodGroupAnnotation], convey.ShouldEqual, pgName)
			convey.So(podTemplate.Annotations[volcanoTaskSpecKey], convey.ShouldEqual, rt)
		})
	})
}

// TestSetRestartPolicy test the function of setRestartPolicy
func TestSetRestartPolicy(t *testing.T) {
	convey.Convey("set restart policy", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		podTemplate := &corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{},
		}
		spec := &commonv1.ReplicaSpec{}
		convey.Convey("01-container restart policy will be set as Never while spec restart policy is ExitCode",
			func() {
				spec.RestartPolicy = commonv1.RestartPolicyExitCode
				rc.setRestartPolicy(job, podTemplate, spec)
				convey.So(podTemplate.Spec.RestartPolicy, convey.ShouldEqual, corev1.RestartPolicyNever)
			})
		convey.Convey("02-container restart policy will be set as spec restart policy", func() {
			spec.RestartPolicy = commonv1.RestartPolicyOnFailure
			rc.setRestartPolicy(job, podTemplate, spec)
			convey.So(podTemplate.Spec.RestartPolicy, convey.ShouldEqual, corev1.RestartPolicyOnFailure)
		})
		convey.Convey("02-pod template restart policy will be overwritten", func() {
			podTemplate.Spec.RestartPolicy = corev1.RestartPolicyNever
			spec.RestartPolicy = commonv1.RestartPolicyOnFailure
			rc.setRestartPolicy(job, podTemplate, spec)
			convey.So(podTemplate.Spec.RestartPolicy, convey.ShouldEqual, corev1.RestartPolicyOnFailure)
		})
	})
}

// TestBatchCreatePods test the function of batchCreatePods
func TestBatchCreatePods(t *testing.T) {
	convey.Convey("batch create pod", t, func() {
		rc := newCommonReconciler()
		info := newCommonPodInfo()
		specs := newReplicaSpecs()
		convey.Convey("01-batch create pods success", func() {
			patch := gomonkey.ApplyPrivateMethod(rc, "getPodsSlice", func(pods []*podInfo,
				replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) ([]corev1.PodList, error) {
				return []corev1.PodList{}, nil
			})
			defer patch.Reset()
			err := rc.batchCreatePods([]*podInfo{info}, specs, info.job)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestGetPodsSlice test the function of getPodsSlice
func TestGetPodsSlice(t *testing.T) {
	convey.Convey("get pods slice", t, func() {
		rc := newCommonReconciler()
		info := newCommonPodInfo()
		specs := newReplicaSpecs()
		convey.Convey("01-get slice success", func() {
			patch := gomonkey.ApplyPrivateMethod(rc, "createPodSpec", func(pi *podInfo,
				replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) (*corev1.PodTemplateSpec, error) {
				return &corev1.PodTemplateSpec{}, nil
			})
			defer patch.Reset()
			_, err := rc.getPodsSlice([]*podInfo{info}, specs)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("02-create podSpec fail", func() {
			patch := gomonkey.ApplyPrivateMethod(rc, "createPodSpec", func(pi *podInfo,
				replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) (*corev1.PodTemplateSpec, error) {
				return nil, errors.New("")
			})
			defer patch.Reset()
			_, err := rc.getPodsSlice([]*podInfo{info}, specs)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

type testCase struct {
	name     string
	input    []*corev1.Pod
	replicas int
	check    func(result [][]*corev1.Pod)
}

func buildTestCase(name string, replicas int, pods []*corev1.Pod, check func(result [][]*corev1.Pod)) testCase {
	return testCase{
		name:     name,
		input:    pods,
		replicas: replicas,
		check:    check,
	}

}

// TestGetPodSlices test GetPodSlices function
func TestGetPodSlices(t *testing.T) {
	convey.Convey("01-nil reconciler should return nil", t, func() {
		var r *ASJobReconciler
		result := r.GetPodSlices([]*corev1.Pod{}, 0)
		convey.So(result, convey.ShouldBeNil)
	})

	rc := newCommonReconciler()
	testCases := buildTestCasesd()
	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			result := rc.GetPodSlices(tc.input, tc.replicas)
			tc.check(result)
		})
	}
}

func buildTestCasesd() []testCase {
	return []testCase{
		buildTestCase("02-normal case with valid pods", num2, []*corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{commonv1.ReplicaIndexLabel: "0"}}},
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{commonv1.ReplicaIndexLabel: "1"}}},
		}, func(result [][]*corev1.Pod) {
			convey.So(len(result), convey.ShouldEqual, num2)
			convey.So(len(result[0]), convey.ShouldEqual, num1)
			convey.So(len(result[1]), convey.ShouldEqual, num1)
		}),
		buildTestCase("03-pod with invalid index label should be ignored", num1, []*corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{commonv1.ReplicaIndexLabel: "invalid"}}},
		}, func(result [][]*corev1.Pod) {
			convey.So(len(result), convey.ShouldEqual, num1)
			convey.So(len(result[0]), convey.ShouldEqual, 0)
		}),
		buildTestCase("04-pod with out of range index should be ignored", num1, []*corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{commonv1.ReplicaIndexLabel: "5"}}},
		}, func(result [][]*corev1.Pod) {
			convey.So(len(result), convey.ShouldEqual, num6)
			convey.So(len(result[0]), convey.ShouldEqual, 0)
		}),
		buildTestCase("05-pod in hot switch flow should be ignored", num1, []*corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{commonv1.ReplicaIndexLabel: "0"},
				Annotations: map[string]string{api.PodTypeKey: api.PodTypeBackup}}},
		}, func(result [][]*corev1.Pod) {
			convey.So(len(result), convey.ShouldEqual, num1)
			convey.So(len(result[0]), convey.ShouldEqual, 0)
		}),
		buildTestCase("06-replicas is 0 should return empty slice", 0, []*corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{commonv1.ReplicaIndexLabel: "0"}}},
		}, func(result [][]*corev1.Pod) {
			convey.So(len(result), convey.ShouldEqual, 0)
		}),
		buildTestCase("07-multiple pods with same index should be in same slice", 1, []*corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{commonv1.ReplicaIndexLabel: "0"}}},
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{commonv1.ReplicaIndexLabel: "0"}}},
		}, func(result [][]*corev1.Pod) {
			convey.So(len(result), convey.ShouldEqual, num1)
			convey.So(len(result[0]), convey.ShouldEqual, num2)
		}),
	}
}

func TestHandleHotSwitch(t *testing.T) {
	convey.Convey("Test handleHotSwitch", t, func() {
		rc := newCommonReconciler()
		pi := newCommonPodInfo()
		pod := &corev1.Pod{}
		replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}

		convey.Convey("01-nil reconciler should return error", func() {
			var r *ASJobReconciler
			err := r.handleHotSwitch(pi, pod, replicas)
			convey.So(err, convey.ShouldResemble, errors.New("nil pointer"))
		})
		convey.Convey("02-pod without NeedOperatorOpeKey annotation should return nil", func() {
			err := rc.handleHotSwitch(pi, pod, replicas)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("03-pod with NeedOperatorOpeKey not equal to create should return nil", func() {
			pod.Annotations = map[string]string{api.NeedOperatorOpeKey: "other"}
			err := rc.handleHotSwitch(pi, pod, replicas)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("04-createHotSwitchPod failed should return error", func() {
			pod.Annotations = map[string]string{api.NeedOperatorOpeKey: api.OpeTypeCreate}
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(rc), "createHotSwitchPod",
				func(_ *ASJobReconciler, _ *corev1.Pod, _ *podInfo,
					_ map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
					return errors.New("create hot switch pod failed")
				})
			defer patch.Reset()
			err := rc.handleHotSwitch(pi, pod, replicas)
			convey.So(err, convey.ShouldResemble, errors.New("create hot switch pod failed"))
		})
		convey.Convey("05-handle hot switch success should return nil", func() {
			pod.Annotations = map[string]string{api.NeedOperatorOpeKey: api.OpeTypeCreate}
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(rc), "createHotSwitchPod",
				func(_ *ASJobReconciler, _ *corev1.Pod, _ *podInfo,
					_ map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
					return nil
				})
			defer patch.Reset()
			err := rc.handleHotSwitch(pi, pod, replicas)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

type testHotSwitchCase struct {
	name       string
	setupMocks func(*gomonkey.Patches, *ASJobReconciler)
	wantErr    bool
}

func buildTestHotSwitchCases() []testHotSwitchCase {
	return []testHotSwitchCase{
		{name: "01-nil reconciler should return error",
			setupMocks: func(p *gomonkey.Patches, r *ASJobReconciler) { mockCreateNewPodReturnNil(p, r) },
			wantErr:    true,
		}, {
			name: "02-createNewPod failed should return error",
			setupMocks: func(p *gomonkey.Patches, r *ASJobReconciler) {
				p.ApplyPrivateMethod(r, "createNewPod", func(*podInfo, map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
					return errors.New("create pod failed")
				})
			},
			wantErr: true,
		}, {
			name: "06-create hot switch pod success",
			setupMocks: func(p *gomonkey.Patches, r *ASJobReconciler) {
				mockCreateNewPodReturnNil(p, r)
				mockGetReturnNil(p, r).ApplyMethodReturn(r.Client, "Update", nil)
			},
			wantErr: false,
		},
	}
}
func TestCreateHotSwitchPod(t *testing.T) {
	tests := buildTestHotSwitchCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convey.Convey(tt.name, t, func() {
				patch := gomonkey.NewPatches()
				defer patch.Reset()

				var r *ASJobReconciler
				if tt.name != "01-nil reconciler should return error" {
					r = newCommonReconciler()
				}
				if tt.setupMocks != nil {
					tt.setupMocks(patch, r)
				}

				oldPod := buildOldPod()
				pi := newCommonPodInfo()
				replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}

				err := r.createHotSwitchPod(oldPod, pi, replicas)

				if tt.wantErr {
					convey.So(err, convey.ShouldNotBeNil)
				} else {
					convey.So(err, convey.ShouldBeNil)
				}
			})
		})
	}
}

func mockGetReturnNil(p *gomonkey.Patches, r *ASJobReconciler) *gomonkey.Patches {
	return p.ApplyMethodFunc(r.Client, "Get", func(_ context.Context, _ client.ObjectKey, obj client.Object) error {
		obj.SetAnnotations(map[string]string{})
		return nil
	})
}

func mockCreateNewPodReturnNil(p *gomonkey.Patches, r *ASJobReconciler) *gomonkey.Patches {
	return p.ApplyPrivateMethod(r, "createNewPod", func(*podInfo, map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
		return nil
	})
}

func buildOldPod() *corev1.Pod {
	oldPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Annotations: map[string]string{
				api.NeedOperatorOpeKey: api.OpeTypeCreate,
			},
		},
	}
	return oldPod
}

func TestSetOnePodOneNode(t *testing.T) {
	rc := newCommonReconciler()
	podsOnDiffNodes := []*corev1.Pod{{
		Spec: corev1.PodSpec{
			NodeName: "node-1",
		},
	}, {
		Spec: corev1.PodSpec{
			NodeName: "node-2",
		},
	}}
	rc.setOnePodOneNode(podsOnDiffNodes)
	for _, pod := range podsOnDiffNodes {
		if pod.Annotations[common.OnePodOneNode] != "true" {
			t.Errorf("when all pods are one different nodes, then OnePodOneNode should be true")
		}
	}

	podsOnSameNodes := []*corev1.Pod{{
		Spec: corev1.PodSpec{
			NodeName: "node-1",
		},
	}, {
		Spec: corev1.PodSpec{
			NodeName: "node-1",
		},
	}}

	rc.setOnePodOneNode(podsOnSameNodes)
	for _, pod := range podsOnSameNodes {
		if pod.Annotations[common.OnePodOneNode] == "true" {
			t.Errorf("when all pods are one same nodes, then OnePodOneNode should be not true")
		}
	}
}
