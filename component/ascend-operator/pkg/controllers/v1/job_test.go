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
	"github.com/kubeflow/common/pkg/util"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	mindxdlv1 "ascend-operator/pkg/api/v1"
	_ "ascend-operator/pkg/testtool"
)

var (
	resultShouldreturnErr       = "should return err"
	updateStatusApiServerFailed = "update status in api-server failed"
)

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
			convey.ShouldEqual(err, errors.New("create jobInfo failed"))
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
			convey.ShouldEqual(err, errors.New("reconcile job failed"))
		})
	})
}

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
				return errors.New(
					"handle finished Job failed")
			})
		defer patch1.Reset()
		err := rc.reconcileJob(ji)
		convey.ShouldEqual(err, errors.New("handle finished Job failed"))
	})
	ji.status = &commonv1.JobStatus{
		Conditions: []commonv1.JobCondition{},
	}
	convey.Convey("02-normal job's pod-group has not synced and update status in server fauled, "+
		resultShouldreturnErr, func() {
		patch2 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "isPodGroupSynced",
			func(_ *ASJobReconciler, _ *jobInfo) bool { return false })
		defer patch2.Reset()
		patch3 := gomonkey.ApplyMethod(new(ASJobReconciler), "UpdateJobStatusInApiServer",
			func(_ *ASJobReconciler, _ interface{}, _ *commonv1.JobStatus) error {
				return errors.New(updateStatusApiServerFailed)
			})
		defer patch3.Reset()
		err := rc.reconcileJob(ji)
		convey.ShouldEqual(err, errors.New(updateStatusApiServerFailed))
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
		convey.ShouldEqual(err, errors.New("sync replicas failed"))
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
		convey.ShouldEqual(err, errors.New("handle out of limit Job failed"))
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
		patch4 := gomonkey.ApplyMethod(new(ASJobReconciler), "UpdateJobStatus",
			func(_ *ASJobReconciler, _ interface{}, _ map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
				_ *commonv1.JobStatus) error {
				return errors.New("update status failed")
			})
		defer patch4.Reset()
		err := rc.reconcileJob(ji)
		convey.ShouldEqual(err, errors.New("update status failed"))
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
		patch5 := gomonkey.ApplyMethod(new(ASJobReconciler), "UpdateJobStatusInApiServer",
			func(_ *ASJobReconciler, _ interface{}, _ *commonv1.JobStatus) error {
				return errors.New(updateStatusApiServerFailed)
			})
		defer patch5.Reset()
		err := rc.reconcileJob(ji)
		convey.ShouldEqual(err, errors.New("update status in api-server  failed"))
	})
}

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
			convey.ShouldEqual(err, errors.New("update spec status failed"))
		})
	})
}

func TestUpdateSpecStatus(t *testing.T) {
	convey.Convey("updateSpecStatus", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
			mindxdlv1.PytorchReplicaTypeMaster: {
				Replicas: defaultReplicas(),
			},
			mindxdlv1.ReplicaTypeWorker: {
				Replicas: defaultReplicas(),
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
			convey.ShouldBeNil(err)
		})
		convey.Convey("02-check spec status failed, should return err", func() {
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "checkSpecStatus",
				func(_ *ASJobReconciler, _ specInfo, _ *commonv1.JobStatus, _ func(*conditionInfo) error) error {
					return errors.New("check spec status failed")
				})
			defer patch1.Reset()
			err := rc.updateSpecStatus(job, replicas, status)
			convey.ShouldEqual(err, errors.New("check spec status failed"))
		})
	})
}

func TestGetJobStatus(t *testing.T) {
	convey.Convey("getJobStatus", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
			mindxdlv1.PytorchReplicaTypeMaster: {
				Replicas: defaultReplicas(),
			},
			mindxdlv1.ReplicaTypeWorker: {
				Replicas: defaultReplicas(),
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
			convey.ShouldEqual(st, expect)
		})
	})
}

func TestCheckSpecStatus(t *testing.T) {
	rc := newCommonReconciler()
	job := newCommonAscendJob()
	job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
		mindxdlv1.PytorchReplicaTypeMaster: {
			Replicas: defaultReplicas(),
		},
		mindxdlv1.ReplicaTypeWorker: {
			Replicas: defaultReplicas(),
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
		patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "checkMasterStatus",
			func(_ *ASJobReconciler, _ specInfo, _ *commonv1.JobStatus) *conditionInfo {
				return &conditionInfo{
					condType: commonv1.JobRunning,
				}
			})
		defer patch1.Reset()
		patch2 := gomonkey.ApplyFunc(util.UpdateJobConditions, func(_ *commonv1.JobStatus,
			_ commonv1.JobConditionType, _, _ string) error {
			return errors.New("update job conditions failed")
		})
		defer patch2.Reset()
		err := rc.checkSpecStatus(job, st, jobStatus, updateFunc)
		convey.ShouldEqual(err, errors.New("update job conditions failed"))
	})
}

func testCheckSpecStatusNoError(rc *ASJobReconciler, job *mindxdlv1.AscendJob,
	jobStatus *commonv1.JobStatus, updateFunc func(ci *conditionInfo) error) {
	convey.Convey("01-status with running and condition is nil should do nothing nil", func() {
		st := &commonv1.ReplicaStatus{
			Active:    2,
			Succeeded: 0,
			Failed:    0,
		}
		patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "checkMasterStatus",
			func(_ *ASJobReconciler, _ specInfo, _ *commonv1.JobStatus) *conditionInfo { return nil })
		defer patch.Reset()
		err := rc.checkSpecStatus(job, st, jobStatus, updateFunc)
		convey.ShouldBeNil(err)
	})
	convey.Convey("03-status with failed and condition is failed, "+
		"when update condition success should return nil", func() {
		st := &commonv1.ReplicaStatus{
			Active:    0,
			Succeeded: 1,
			Failed:    1,
		}
		patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "checkMasterStatus",
			func(_ *ASJobReconciler, _ specInfo, _ *commonv1.JobStatus) *conditionInfo {
				return &conditionInfo{
					condType: commonv1.JobFailed,
				}
			})
		defer patch1.Reset()
		patch2 := gomonkey.ApplyFunc(util.UpdateJobConditions, func(_ *commonv1.JobStatus,
			_ commonv1.JobConditionType, _, _ string) error {
			return nil
		})
		defer patch2.Reset()
		err := rc.checkSpecStatus(job, st, jobStatus, updateFunc)
		convey.ShouldBeNil(err)
		convey.ShouldNotBeNil(job.CreationTimestamp)
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
		convey.Convey("reconcile services failed should return err", func() {
			patch := gomonkey.ApplyMethod(new(ASJobReconciler), "ReconcileServices", func(_ *ASJobReconciler, _ metav1.Object,
				_ []*corev1.Service, _ commonv1.ReplicaType, _ *commonv1.ReplicaSpec) error {
				return errors.New(
					"reconcile services failed")
			})
			defer patch.Reset()
			err := rc.syncReplicas(ji)
			convey.ShouldEqual(err, errors.New("reconcile services failed"))
		})
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
			convey.ShouldEqual(err, errors.New("reconcile pods failed"))
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
					Replicas: defaultReplicas(),
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
			convey.ShouldEqual(pgSpec, v1beta1.PodGroupSpec{
				MinMember:         *defaultReplicas(),
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
			job: newCommonAscendJob(),
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
			convey.ShouldEqual(res, false)
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
			convey.ShouldEqual(res, false)
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
			convey.ShouldEqual(res, true)
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
			convey.ShouldEqual(err, errors.New("delete pods and service failed"))
		})
		convey.Convey("02-clean up job failed, should return err", func() {
			patch := gomonkey.ApplyMethod(new(common.JobController), "CleanupJob",
				func(_ *common.JobController, _ *commonv1.RunPolicy, _ commonv1.JobStatus, _ interface{}) error {
					return errors.New("clean up job failed")
				})
			defer patch.Reset()
			err := rc.handleFinishedJob(ji, false, conditionInfo{})
			convey.ShouldEqual(err, errors.New("clean up job failed"))
		})
		convey.Convey("03-delete podgroup failed, should return err", func() {
			patch := gomonkey.ApplyMethod(new(ASJobReconciler), "DeletePodGroup",
				func(_ *ASJobReconciler, _ metav1.Object) error {
					return errors.New("delete podgroup failed")
				})
			defer patch.Reset()
			err := rc.handleFinishedJob(ji, false, conditionInfo{})
			convey.ShouldEqual(err, errors.New("delete podgroup failed"))
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
			convey.ShouldEqual(err, errors.New("update condition failed"))
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
			convey.ShouldBeNil(err)
			convey.ShouldEqual(ji.status.ReplicaStatuses[mindxdlv1.PytorchReplicaTypeMaster], &commonv1.ReplicaStatus{
				Succeeded: 1,
			})
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
			convey.ShouldBeNil(err)
			convey.ShouldEqual(pg, fakePodGroup)
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
			convey.ShouldEqual(err, errors.New("create podGroup failed"))
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
			convey.ShouldBeNil(err)
			convey.ShouldEqual(pg, fakePodGroup)
		})
	})
}
