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
	"k8s.io/apimachinery/pkg/types"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable"
	"ascend-operator/pkg/ranktable/generator"
	"ascend-operator/pkg/ranktable/utils"
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
			Replicas: newReplicas(1),
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
			if _, ok := annotatedPod.Annotations[rankIndexKey]; !ok {
				t.Errorf("fail to write rank index for annotated pod")
			}
		})
	}
}

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
			convey.ShouldEqual(rtg.GetConfigmapStatus(), utils.UnknownStatus)
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
			convey.ShouldEqual(rtg.GetConfigmapStatus(), utils.UnknownStatus)
		})
	})
}

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
			convey.ShouldEqual(rtg.GetConfigmapStatus(), utils.InitialRTStatus)
		})
	})
}

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
			convey.ShouldBeError(err)
		})
		convey.Convey("02-write success, should return nil", func() {
			patch := gomonkey.ApplyPrivateMethod(r, "writeRanktableToCm",
				func(_ *ASJobReconciler, jobName, namespace string, uid types.UID) error {
					return nil
				})
			defer patch.Reset()
			err := r.tryWriteCm("", "", testUid)
			convey.ShouldBeNil(err)
		})
	})
}
