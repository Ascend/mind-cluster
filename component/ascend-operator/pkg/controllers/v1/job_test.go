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
	"reflect"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/controller.v1/common"
	"github.com/kubeflow/common/pkg/util"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	mindxdlv1 "ascend-operator/pkg/api/v1"
)

var (
	resultShouldreturnErr = "should return err"
)

// TestReconcileJobs test ReconcileJobs
func TestReconcileJobs(t *testing.T) {
	convey.Convey("ReconcileJobs", t, func() {
		job := newCommonAscendJob()
		replicaTypes := make(map[commonv1.ReplicaType]*commonv1.ReplicaSpec)
		jobStatus := commonv1.JobStatus{}
		runPolicy := &commonv1.RunPolicy{}
		rc := &ASJobReconciler{}
		rc.Controller = rc
		convey.Convey("01-create jobInfo failed, should return err", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "newJobInfo", func(
				_ *ASJobReconciler,
				_ interface{},
				_ map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
				_ *commonv1.JobStatus,
				_ *commonv1.RunPolicy) (*jobInfo, error) {
				return nil, errors.New("create jobInfo failed")
			})
			defer patch1.Reset()
			err := rc.ReconcileJobs(job, replicaTypes, jobStatus, runPolicy)
			convey.So(err, convey.ShouldResemble, errors.New("create jobInfo failed"))
		})
		convey.Convey("02-reconcile job failed, should return err", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "newJobInfo", func(
				_ *ASJobReconciler,
				_ interface{},
				_ map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
				_ *commonv1.JobStatus,
				_ *commonv1.RunPolicy) (*jobInfo, error) {
				return nil, nil
			})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "reconcileJob",
				func(_ *jobInfo) error { return errors.New("reconcile job failed") })
			defer patch2.Reset()
			err := rc.ReconcileJobs(job, replicaTypes, jobStatus, runPolicy)
			convey.So(err, convey.ShouldResemble, errors.New("reconcile job failed"))
		})
	})
}

// TestReconcileJob test ReconcileJob
func TestReconcileJob(t *testing.T) {
	rc := newCommonReconciler()
	ttlSecondsAfterFinished := int32(3)
	ji := &jobInfo{
		job: newCommonAscendJob(),
		status: &commonv1.JobStatus{
			Conditions: []commonv1.JobCondition{{
				Type:   commonv1.JobSucceeded,
				Status: corev1.ConditionTrue,
			}},
		},
		mtObj: &mindxdlv1.AscendJob{},
		runPolicy: &commonv1.RunPolicy{
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
		},
	}
	ji.mtObj.SetUID("test")
	rc.versions["test"] = 0
	rc.backoffLimits["test"] = 5
	convey.Convey("reconcileJob", t, func() {
		testReconcileJobUpdateFailed(rc, ji)
		testReconcileJobPodGroupAndRepSyncedSuccess(rc, ji)
	})
}

func testReconcileJobUpdateFailed(rc *ASJobReconciler, ji *jobInfo) {
	convey.Convey("01-success job update condition failed, should return err", func() {
		patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "handleFinishedJob",
			func(_ *ASJobReconciler, _ *jobInfo, _ bool, _ conditionInfo) error {
				return errors.New("handle finished Job failed")
			})
		defer patch1.Reset()
		err := rc.reconcileJob(ji)
		convey.So(err, convey.ShouldResemble, errors.New("handle finished Job failed"))
	})
	ji.status = &commonv1.JobStatus{
		Conditions: []commonv1.JobCondition{},
	}
	convey.Convey("02-normal job's pod-group has not synced and update status in server failed, "+
		resultShouldreturnErr, func() {
		patch2 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "isPodGroupSynced",
			func(_ *ASJobReconciler, _ *jobInfo) bool { return false })
		defer patch2.Reset()
		err := rc.reconcileJob(ji)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("03-normal job's pod group has synced and sync replicas failed, "+
		resultShouldreturnErr, func() {
		patch2 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "isPodGroupSynced",
			func(_ *ASJobReconciler, _ *jobInfo) bool { return true })
		defer patch2.Reset()
		patch3 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "syncReplicas",
			func(_ *ASJobReconciler, _ *jobInfo) error { return errors.New("sync replicas failed") })
		defer patch3.Reset()
		err := rc.reconcileJob(ji)
		convey.So(err, convey.ShouldResemble, errors.New("sync replicas failed"))
	})
}

func testReconcileJobPodGroupAndRepSyncedSuccess(rc *ASJobReconciler, ji *jobInfo) {
	convey.Convey("01-normal job's pod group has synced and sync replicas success, but update status failed"+
		resultShouldreturnErr, func() {
		patch2 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "isPodGroupSynced",
			func(_ *ASJobReconciler, _ *jobInfo) bool { return true })
		defer patch2.Reset()
		patch3 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "syncReplicas",
			func(_ *ASJobReconciler, _ *jobInfo) error { return nil })
		defer patch3.Reset()
		err := rc.reconcileJob(ji)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("02-normal job's pod group has synced and sync replicas success, "+
		"but update in api-server failed should return err", func() {
		patch2 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "isPodGroupSynced",
			func(_ *ASJobReconciler, _ *jobInfo) bool { return true })
		defer patch2.Reset()
		patch3 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "syncReplicas",
			func(_ *ASJobReconciler, _ *jobInfo) error { return nil })
		defer patch3.Reset()
		patch4 := gomonkey.ApplyMethod(new(ASJobReconciler), "UpdateJobStatus",
			func(_ *ASJobReconciler, _ interface{}, _ map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
				_ *commonv1.JobStatus) error {
				return nil
			})
		defer patch4.Reset()
		err := rc.reconcileJob(ji)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("04-out of limit job update condition failed, should return err", func() {
		patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "handleFinishedJob",
			func(_ *ASJobReconciler, _ *jobInfo, _ bool, _ conditionInfo) error {
				return errors.New(
					"handle out of limit Job failed")
			})
		defer patch1.Reset()
		rc.versions = make(map[types.UID]int32)
		rc.backoffLimits = make(map[types.UID]int32)
		err := rc.reconcileJob(ji)
		convey.So(err, convey.ShouldResemble, errors.New("handle out of limit Job failed"))
	})
}

// TestUpdateJobStatus test update job status
func TestUpdateJobStatus(t *testing.T) {
	convey.Convey("UpdateJobStatus", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		replicas := make(map[commonv1.ReplicaType]*commonv1.ReplicaSpec)
		status := &commonv1.JobStatus{}
		convey.Convey("02-update spec status failed, should return err", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "updateSpecStatus",
				func(_ *ASJobReconciler, _ *mindxdlv1.AscendJob, _ map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
					_ *commonv1.JobStatus) error {
					return errors.New("update spec status failed")
				})
			defer patch.Reset()
			err := rc.UpdateJobStatus(job, replicas, status)
			convey.So(err, convey.ShouldResemble, errors.New("update spec status failed"))
		})
	})
}

// TestUpdateSpecStatus test update spec status
func TestUpdateSpecStatus(t *testing.T) {
	convey.Convey("updateSpecStatus", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
			mindxdlv1.PytorchReplicaTypeMaster: {
				Replicas: newReplicas(1),
			},
			mindxdlv1.ReplicaTypeWorker: {
				Replicas: newReplicas(1),
			},
		}
		status := &commonv1.JobStatus{
			ReplicaStatuses: map[commonv1.ReplicaType]*commonv1.ReplicaStatus{
				mindxdlv1.PytorchReplicaTypeMaster: {},
				mindxdlv1.ReplicaTypeWorker:        {},
			},
		}
		convey.Convey("01-replicas is empty should return nil", func() {
			repl := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
			err := rc.updateSpecStatus(job, repl, status)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("02-check spec status failed, should return err", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "checkSpecStatus",
				func(_ *ASJobReconciler, _ specInfo, _ *commonv1.JobStatus, _ func(*conditionInfo) error) error {
					return errors.New("check spec status failed")
				})
			defer patch1.Reset()
			err := rc.updateSpecStatus(job, replicas, status)
			convey.So(err, convey.ShouldResemble, errors.New("check spec status failed"))
		})
	})
}

func TestGetJobStatus(t *testing.T) {
	convey.Convey("getJobStatus", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
			mindxdlv1.PytorchReplicaTypeMaster: {
				Replicas: newReplicas(1),
			},
			mindxdlv1.ReplicaTypeWorker: {
				Replicas: newReplicas(1),
			},
		}
		jobStatus := &commonv1.JobStatus{
			ReplicaStatuses: map[commonv1.ReplicaType]*commonv1.ReplicaStatus{
				mindxdlv1.PytorchReplicaTypeMaster: {Active: 0, Succeeded: 0, Failed: 1},
				mindxdlv1.ReplicaTypeWorker:        {Active: 0, Succeeded: 1, Failed: 0},
			},
		}
		convey.Convey("01-get job status", func() {
			expect := &commonv1.ReplicaStatus{
				Active:    0,
				Succeeded: 1,
				Failed:    1,
			}
			st := rc.getJobStatus(job, replicas, jobStatus)
			convey.So(st, convey.ShouldResemble, expect)
		})
	})
}

func TestCheckSpecStatus(t *testing.T) {
	rc := newCommonReconciler()
	job := newCommonAscendJob()
	job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
		mindxdlv1.PytorchReplicaTypeMaster: {
			Replicas: newReplicas(1),
		},
		mindxdlv1.ReplicaTypeWorker: {
			Replicas: newReplicas(1),
		},
	}
	jobStatus := &commonv1.JobStatus{
		Conditions: []commonv1.JobCondition{},
	}
	updateFunc := func(ci *conditionInfo) error {
		if ci == nil {
			return nil
		}
		if ci.condType == commonv1.JobSucceeded || ci.condType == commonv1.JobFailed {
			if jobStatus.CompletionTime == nil {
				now := metav1.Now()
				jobStatus.CompletionTime = &now
			}
		}
		err := util.UpdateJobConditions(jobStatus, ci.condType, ci.reason, ci.message)
		if err != nil {
			hwlog.RunLog.Errorf("Append ascendJob<%s-%s> condition err: %v",
				job.Namespace, job.Name, err)
			return err
		}
		return nil
	}
	convey.Convey("checkSpecStatus", t, func() {
		testCheckSpecStatusNoError(rc, job, jobStatus, updateFunc)
		testCheckSpecStatusWithError(rc, job, jobStatus, updateFunc)
	})
}

func testCheckSpecStatusWithError(rc *ASJobReconciler, job *mindxdlv1.AscendJob,
	jobStatus *commonv1.JobStatus, updateFunc func(ci *conditionInfo) error) {
	convey.Convey("02-when condition is running, and update condition failed should return err", func() {
		st := &commonv1.ReplicaStatus{
			Active:    1,
			Succeeded: 1,
			Failed:    0,
		}
		patch2 := gomonkey.ApplyFunc(util.UpdateJobConditions, func(_ *commonv1.JobStatus,
			_ commonv1.JobConditionType, _, _ string) error {
			return errors.New("update job conditions failed")
		})
		defer patch2.Reset()
		err := rc.checkSpecStatus(job, st, jobStatus, updateFunc)
		convey.So(err, convey.ShouldResemble, errors.New("update job conditions failed"))
	})
}

func testCheckSpecStatusNoError(rc *ASJobReconciler, job *mindxdlv1.AscendJob,
	jobStatus *commonv1.JobStatus, updateFunc func(ci *conditionInfo) error) {
	convey.Convey("01-status with running and condition is nil should do nothing nil", func() {
		const fakeActive = 2
		st := &commonv1.ReplicaStatus{
			Active:    fakeActive,
			Succeeded: 0,
			Failed:    0,
		}
		err := rc.checkSpecStatus(job, st, jobStatus, updateFunc)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("03-status with failed and condition is failed, "+
		"when update condition success should return nil", func() {
		st := &commonv1.ReplicaStatus{
			Active:    0,
			Succeeded: 1,
			Failed:    1,
		}
		patch2 := gomonkey.ApplyFunc(util.UpdateJobConditions, func(_ *commonv1.JobStatus,
			_ commonv1.JobConditionType, _, _ string) error {
			return nil
		})
		defer patch2.Reset()
		err := rc.checkSpecStatus(job, st, jobStatus, updateFunc)
		convey.So(err, convey.ShouldBeNil)
		convey.So(job.CreationTimestamp, convey.ShouldNotBeNil)
	})
}

func TestSyncReplicas(t *testing.T) {
	convey.Convey("syncReplicas", t, func() {
		rc := newCommonReconciler()
		ji := &jobInfo{
			job:   newCommonAscendJob(),
			mtObj: &mindxdlv1.AscendJob{},
			status: &commonv1.JobStatus{
				Conditions: []commonv1.JobCondition{{
					Type:   commonv1.JobSucceeded,
					Status: corev1.ConditionTrue,
				}},
			},
			rpls: map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
				mindxdlv1.PytorchReplicaTypeMaster: {},
				mindxdlv1.ReplicaTypeWorker:        {},
			},
		}
		convey.Convey("reconcile pods failed should return err", func() {
			patch1 := gomonkey.ApplyMethod(new(ASJobReconciler), "ReconcileServices", func(_ *ASJobReconciler, _ metav1.Object,
				_ []*corev1.Service, _ commonv1.ReplicaType, _ *commonv1.ReplicaSpec) error {
				return nil
			})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyMethod(new(ASJobReconciler), "ReconcilePods", func(_ *ASJobReconciler,
				_ interface{}, _ *commonv1.JobStatus, pods []*corev1.Pod, _ commonv1.ReplicaType, _ *commonv1.ReplicaSpec,
				_ map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
				return errors.New("reconcile pods failed")
			})
			defer patch2.Reset()
			err := rc.syncReplicas(ji)
			convey.So(err, convey.ShouldResemble, errors.New("reconcile pods failed"))
		})
	})
}

func TestNewPodGroupSpec(t *testing.T) {
	convey.Convey("newPodGroupSpec", t, func() {
		rc := newCommonReconciler()
		ji := &jobInfo{
			job: newCommonAscendJob(),
			runPolicy: &commonv1.RunPolicy{
				SchedulingPolicy: &commonv1.SchedulingPolicy{
					Queue: "default",
				},
			},
			rpls: map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
				mindxdlv1.ReplicaTypeWorker: {
					Replicas: newReplicas(1),
				},
			},
		}
		convey.Convey("when run policy is set, pod group spec should be corresponding to it", func() {
			patch1 := gomonkey.ApplyFunc(common.CalcPGMinResources, func(_ int32,
				_ map[commonv1.ReplicaType]*commonv1.ReplicaSpec, _ common.PriorityClassGetFunc) *corev1.
				ResourceList {
				return nil
			})
			defer patch1.Reset()
			pgSpec := rc.newPodGroupSpec(ji)
			convey.So(pgSpec, convey.ShouldResemble, v1beta1.PodGroupSpec{
				MinMember:         1,
				MinTaskMember:     map[string]int32{strings.ToLower(string(mindxdlv1.ReplicaTypeWorker)): 1},
				Queue:             "default",
				PriorityClassName: "",
				MinResources:      nil,
			})
		})
	})
}

func TestIsPodGroupSyncedFalseScene(t *testing.T) {
	convey.Convey("isPodGroupSynced", t, func() {
		rc := newCommonReconciler()
		ji := &jobInfo{
			job:    newCommonAscendJob(),
			status: &commonv1.JobStatus{},
		}
		convey.Convey("01-sync pod group failed, should return false", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "newPodGroupSpec", func(_ *ASJobReconciler,
				_ *jobInfo) v1beta1.PodGroupSpec {
				return v1beta1.PodGroupSpec{}
			})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyMethod(new(ASJobReconciler), "SyncPodGroup", func(_ *ASJobReconciler,
				_ metav1.Object, _ v1beta1.PodGroupSpec) (*v1beta1.PodGroup, error) {
				return nil, errors.New("sync pod group failed")
			})
			defer patch2.Reset()
			res := rc.isPodGroupSynced(ji)
			convey.So(res, convey.ShouldEqual, false)
		})
		convey.Convey("02-pod group phase is pending, should return false", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "newPodGroupSpec", func(_ *ASJobReconciler,
				_ *jobInfo) v1beta1.PodGroupSpec {
				return v1beta1.PodGroupSpec{}
			})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyMethod(new(ASJobReconciler), "SyncPodGroup", func(_ *ASJobReconciler,
				_ metav1.Object, _ v1beta1.PodGroupSpec) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{
					Status: v1beta1.PodGroupStatus{
						Phase: v1beta1.PodGroupPending,
					},
				}, nil
			})
			defer patch2.Reset()
			res := rc.isPodGroupSynced(ji)
			convey.So(res, convey.ShouldEqual, false)
		})
	})
}

func TestIsPodGroupSyncedTrueScene(t *testing.T) {
	convey.Convey("isPodGroupSynced", t, func() {
		rc := newCommonReconciler()
		ji := &jobInfo{
			job: newCommonAscendJob(),
		}
		convey.Convey("03-pod group phase is Inqueue, should return true", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "newPodGroupSpec", func(_ *ASJobReconciler,
				_ *jobInfo) v1beta1.PodGroupSpec {
				return v1beta1.PodGroupSpec{}
			})
			defer patch1.Reset()
			patch := gomonkey.ApplyMethod(new(ASJobReconciler), "SyncPodGroup", func(_ *ASJobReconciler,
				_ metav1.Object, _ v1beta1.PodGroupSpec) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{
					Status: v1beta1.PodGroupStatus{
						Phase: v1beta1.PodGroupInqueue,
					},
				}, nil
			})
			defer patch.Reset()
			res := rc.isPodGroupSynced(ji)
			convey.So(res, convey.ShouldEqual, true)
		})
	})
}

func TestHandleFinishedJob(t *testing.T) {
	convey.Convey("handleFinishedJob", t, func() {
		rc := newCommonReconciler()
		ji := &jobInfo{
			job: newCommonAscendJob(),
			status: &commonv1.JobStatus{
				Conditions: []commonv1.JobCondition{{Type: commonv1.JobSucceeded}},
				ReplicaStatuses: map[commonv1.ReplicaType]*commonv1.ReplicaStatus{
					mindxdlv1.PytorchReplicaTypeMaster: {Active: 1},
				},
			},
			pods:      []*corev1.Pod{},
			runPolicy: &commonv1.RunPolicy{},
		}
		convey.Convey("01-delete pods and service failed, should return err", func() {
			patch := gomonkey.ApplyMethod(new(common.JobController), "DeletePodsAndServices",
				func(_ *common.JobController, _ *commonv1.RunPolicy, _ interface{}, _ []*corev1.Pod) error {
					return errors.New("delete pods and service failed")
				})
			defer patch.Reset()
			err := rc.handleFinishedJob(ji, false, conditionInfo{})
			convey.So(err, convey.ShouldResemble, errors.New("delete pods and service failed"))
		})
		convey.Convey("02-clean up job failed, should return err", func() {
			patch := gomonkey.ApplyMethod(new(common.JobController), "CleanupJob",
				func(_ *common.JobController, _ *commonv1.RunPolicy, _ commonv1.JobStatus, _ interface{}) error {
					return errors.New("clean up job failed")
				})
			defer patch.Reset()
			err := rc.handleFinishedJob(ji, false, conditionInfo{})
			convey.So(err, convey.ShouldResemble, errors.New("clean up job failed"))
		})
		convey.Convey("03-delete podgroup failed, should return err", func() {
			patch := gomonkey.ApplyMethod(new(ASJobReconciler), "DeletePodGroup",
				func(_ *ASJobReconciler, _ metav1.Object) error {
					return errors.New("delete podgroup failed")
				})
			defer patch.Reset()
			err := rc.handleFinishedJob(ji, false, conditionInfo{})
			convey.So(err, convey.ShouldResemble, errors.New("delete podgroup failed"))
		})
	})
}

func TestHandleFinishedJobUpdateStep(t *testing.T) {
	convey.Convey("handleFinishedJob", t, func() {
		rc := newCommonReconciler()
		ji := &jobInfo{
			job: newCommonAscendJob(),
			status: &commonv1.JobStatus{
				Conditions: []commonv1.JobCondition{{Type: commonv1.JobSucceeded}},
				ReplicaStatuses: map[commonv1.ReplicaType]*commonv1.ReplicaStatus{
					mindxdlv1.PytorchReplicaTypeMaster: {Active: 1},
				},
			},
			pods:      []*corev1.Pod{},
			runPolicy: &commonv1.RunPolicy{},
		}
		convey.Convey("01-update condition failed, should return err", func() {
			patch1 := gomonkey.ApplyMethod(new(ASJobReconciler), "DeletePodGroup",
				func(_ *ASJobReconciler, _ metav1.Object) error { return nil })
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(util.UpdateJobConditions, func(_ *commonv1.JobStatus,
				_ commonv1.JobConditionType, _, _ string) error {
				return errors.New("update condition failed")
			})
			defer patch2.Reset()
			err := rc.handleFinishedJob(ji, true, conditionInfo{})
			convey.So(err, convey.ShouldResemble, errors.New("update condition failed"))
		})
		convey.Convey("02-when all update process finish and job is success, "+
			"all replica type should set to success", func() {
			patch1 := gomonkey.ApplyMethod(new(ASJobReconciler), "DeletePodGroup",
				func(_ *ASJobReconciler, _ metav1.Object) error { return nil })
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(util.UpdateJobConditions, func(_ *commonv1.JobStatus,
				_ commonv1.JobConditionType, _, _ string) error {
				return nil
			})
			defer patch2.Reset()
			err := rc.handleFinishedJob(ji, true, conditionInfo{})
			convey.So(err, convey.ShouldBeNil)
			convey.So(ji.status.ReplicaStatuses[mindxdlv1.PytorchReplicaTypeMaster], convey.ShouldResemble,
				&commonv1.ReplicaStatus{Active: 1})
		})
	})
}

func TestSyncPodGroup(t *testing.T) {
	convey.Convey("SyncPodGroup", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		pgSpec := v1beta1.PodGroupSpec{}
		fakePodGroup := &v1beta1.PodGroup{}
		convey.Convey("01-get podGroup success, should return podGroup and nil err", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getPodGroup",
				func(_ *ASJobReconciler, _ metav1.Object) (*v1beta1.PodGroup, error) { return fakePodGroup, nil })
			defer patch.Reset()
			pg, err := rc.SyncPodGroup(job, pgSpec)
			convey.So(err, convey.ShouldBeNil)
			convey.So(pg, convey.ShouldResemble, fakePodGroup)
		})
		convey.Convey("02-get podGroup failed and create pg failed, should return err", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getPodGroup",
				func(_ *ASJobReconciler, _ metav1.Object) (*v1beta1.PodGroup, error) {
					return nil,
						errors.New("get podGroup failed")
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "createPodGroup",
				func(_ *ASJobReconciler, _ metav1.Object, _ v1beta1.PodGroupSpec) (*v1beta1.PodGroup, error) {
					return nil,
						errors.New("create podGroup failed")
				})
			defer patch2.Reset()
			_, err := rc.SyncPodGroup(job, pgSpec)
			convey.So(err, convey.ShouldResemble, errors.New("create podGroup failed"))
		})
		convey.Convey("03-get podGroup failed and create pg success, should return podGroup and nil err", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getPodGroup",
				func(_ *ASJobReconciler, _ metav1.Object) (*v1beta1.PodGroup, error) {
					return nil,
						errors.New("get podGroup failed")
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "createPodGroup",
				func(_ *ASJobReconciler, _ metav1.Object, _ v1beta1.PodGroupSpec) (*v1beta1.PodGroup,
					error) {
					return fakePodGroup, nil
				})
			defer patch2.Reset()
			pg, err := rc.SyncPodGroup(job, pgSpec)
			convey.So(err, convey.ShouldBeNil)
			convey.So(pg, convey.ShouldResemble, fakePodGroup)
		})
	})
}

func TestIsProcessRecoverJob(t *testing.T) {
	convey.Convey("isProcessRecoverJob", t, func() {
		convey.Convey("01-job has recover strategy annotation, should return true", func() {
			job := &mindxdlv1.AscendJob{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{api.RecoverStrategyKey: "some-strategy"},
				},
			}
			result := isProcessRecoverJob(job)
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("02-job does not have recover strategy annotation, should return false", func() {
			job := &mindxdlv1.AscendJob{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			}
			result := isProcessRecoverJob(job)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("03-job has nil annotations, should return false", func() {
			job := &mindxdlv1.AscendJob{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: nil,
				},
			}
			result := isProcessRecoverJob(job)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

// TestGetJobRecoverStrategy test getJobRecoverStrategy
func TestGetJobRecoverStrategy(t *testing.T) {
	convey.Convey("getJobRecoverStrategy", t, func() {
		convey.Convey("01-ascendJob is nil, should return empty string", func() {
			result := getJobRecoverStrategy(nil)
			convey.So(result, convey.ShouldEqual, "")
		})

		convey.Convey("02-ascendJob has recover strategy annotation, should return the strategy", func() {
			expectedStrategy := "some-strategy"
			ascendJob := &mindxdlv1.AscendJob{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						api.RecoverStrategyKey: expectedStrategy,
					},
				},
			}
			result := getJobRecoverStrategy(ascendJob)
			convey.So(result, convey.ShouldEqual, expectedStrategy)
		})

		convey.Convey("03-ascendJob does not have recover strategy annotation, should return empty string", func() {
			ascendJob := &mindxdlv1.AscendJob{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}}}
			result := getJobRecoverStrategy(ascendJob)
			convey.So(result, convey.ShouldEqual, "")
		})

		convey.Convey("04-ascendJob has nil annotations, should return empty string", func() {
			ascendJob := &mindxdlv1.AscendJob{ObjectMeta: metav1.ObjectMeta{Annotations: nil}}
			result := getJobRecoverStrategy(ascendJob)
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

// TestIsPodScheduleStrategy test isPodScheduleStrategy
func TestIsPodScheduleStrategy(t *testing.T) {
	convey.Convey("01-ascendJob is nil, should return false", t, func() {
		result := isPodScheduleStrategy(nil)
		convey.So(result, convey.ShouldBeFalse)
	})
	convey.Convey("02-ascendJob has pod schedule label with enable value, should return true", t, func() {
		ascendJob := &mindxdlv1.AscendJob{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{api.PodScheduleLabel: "on"}}}
		result := isPodScheduleStrategy(ascendJob)
		convey.So(result, convey.ShouldBeTrue)
	})

	convey.Convey("03-ascendJob has pod schedule label with disable value, should return false", t, func() {
		ascendJob := &mindxdlv1.AscendJob{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{api.PodScheduleLabel: "disable"}},
		}
		result := isPodScheduleStrategy(ascendJob)
		convey.So(result, convey.ShouldBeFalse)
	})

	convey.Convey("04-ascendJob does not have pod schedule label, should return false", t, func() {
		ascendJob := &mindxdlv1.AscendJob{
			ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}}}
		result := isPodScheduleStrategy(ascendJob)
		convey.So(result, convey.ShouldBeFalse)
	})

	convey.Convey("05-ascendJob has nil labels, should return false", t, func() {
		ascendJob := &mindxdlv1.AscendJob{
			ObjectMeta: metav1.ObjectMeta{Labels: nil},
		}
		result := isPodScheduleStrategy(ascendJob)
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestASJobReconcilerDeletePendingPodsNilJobInfo(t *testing.T) {
	r := &ASJobReconciler{}
	err := r.deletePendingPods(nil)
	assert.NoError(t, err)
	ji := &jobInfo{
		pods: []*corev1.Pod{nil},
	}
	err = r.deletePendingPods(ji)
	assert.NoError(t, err)
}

func TestASJobReconcilerDeletePendingPodsNoPendingPods(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	ji := &jobInfo{
		pods: []*corev1.Pod{pod},
	}
	r := &ASJobReconciler{}
	err := r.deletePendingPods(ji)
	assert.NoError(t, err)
}

type mockClient struct {
	client.Client
}

func TestASJobReconcilerDeletePendingPodsSuccessfulDelete(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
		},
	}
	ji := &jobInfo{
		pods: []*corev1.Pod{pod},
	}
	r := &ASJobReconciler{
		Client: mockClient{},
	}

	patches := gomonkey.ApplyMethod(reflect.TypeOf(&mockClient{}), "Delete", func(_ *mockClient, ctx context.Context,
		obj client.Object, opts ...client.DeleteOption) error {
		return nil
	})
	defer patches.Reset()
	err := r.deletePendingPods(ji)
	assert.NoError(t, err)
}

func TestASJobReconcilerDeletePendingPodsDeleteFailure(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
		},
	}
	ji := &jobInfo{
		pods: []*corev1.Pod{pod},
	}
	r := &ASJobReconciler{
		Client: mockClient{},
	}
	expectedErr := assert.AnError
	patches := gomonkey.ApplyMethod(reflect.TypeOf(&mockClient{}), "Delete", func(_ *mockClient, ctx context.Context,
		obj client.Object, opts ...client.DeleteOption) error {
		return expectedErr
	})
	defer patches.Reset()

	err := r.deletePendingPods(ji)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestGetSubHealthyStrategy(t *testing.T) {
	cases := []struct {
		name           string
		labels         map[string]string
		expectedResult string
	}{
		{name: "Nil job",
			labels:         nil,
			expectedResult: ""},
		{name: "Nil labels",
			labels:         nil,
			expectedResult: ""},
		{name: "Empty labels",
			labels:         map[string]string{},
			expectedResult: ""},
		{name: "SubHealthyStrategy annotation not present",
			labels:         map[string]string{"other-annotation": "value"},
			expectedResult: ""},
		{name: "SubHealthyStrategy annotation present with empty value",
			labels:         map[string]string{api.SubHealthyStrategy: ""},
			expectedResult: ""},
		{name: "SubHealthyStrategy annotation present with value",
			labels:         map[string]string{api.SubHealthyStrategy: api.SubHealthyHotSwitch},
			expectedResult: api.SubHealthyHotSwitch},
		{name: "SubHealthyStrategy annotation present with custom value",
			labels:         map[string]string{api.SubHealthyStrategy: "custom-value"},
			expectedResult: "custom-value"},
	}

	for _, tc := range cases {
		convey.Convey("When "+tc.name, t, func() {
			var job *mindxdlv1.AscendJob
			if tc.name == "Nil job" {
				job = nil
			} else {
				job = &mindxdlv1.AscendJob{ObjectMeta: metav1.ObjectMeta{Labels: tc.labels}}
			}
			result := getSubHealthyStrategy(job)
			convey.So(result, convey.ShouldEqual, tc.expectedResult)
		})
	}
}
