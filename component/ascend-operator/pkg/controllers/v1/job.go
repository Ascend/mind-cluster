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
	"reflect"
	"strconv"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/controller.v1/common"
	"github.com/kubeflow/common/pkg/util"
	"github.com/kubeflow/common/pkg/util/k8sutil"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	mindxdlv1 "ascend-operator/pkg/api/v1"
)

type conditionInfo struct {
	condType        commonv1.JobConditionType
	reason, message string
}

// ReconcileJobs is used to reconcile the job related pod and service
func (r *ASJobReconciler) ReconcileJobs(
	job interface{},
	replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
	jobStatus commonv1.JobStatus,
	runPolicy *commonv1.RunPolicy) error {
	if r == nil {
		return errors.New("nil pointer")
	}
	hwlog.RunLog.Debugf("start reconcile AscendJob, job status: %v, runpolicy: %v", jobStatus, runPolicy)
	ji, err := r.newJobInfo(job, replicas, &jobStatus, runPolicy)
	if err != nil {
		return err
	}
	return r.reconcileJob(ji)
}

func (r *ASJobReconciler) reconcileJob(ji *jobInfo) error {
	oldStatus := ji.status.DeepCopy()
	var err error
	defer func() {
		if err == nil && !reflect.DeepEqual(*oldStatus, ji.status) {
			err = r.Controller.UpdateJobStatusInApiServer(ji.job, ji.status)
		}
	}()

	if util.IsSucceeded(*ji.status) || util.IsFailed(*ji.status) {
		err = r.handleFinishedJob(ji, false, conditionInfo{})
		return err
	}

	version, ok := r.versions[ji.mtObj.GetUID()]
	backoffLimit, backoffLimitOk := r.backoffLimits[ji.mtObj.GetUID()]
	if !ok || (backoffLimitOk && backoffLimit > 0 && version > backoffLimit) {
		hwlog.RunLog.Warnf("Job %s has failed because it has reached the specified backoff limit", ji.name)
		err = r.handleFinishedJob(ji, true, conditionInfo{
			condType: commonv1.JobFailed,
			reason:   util.JobFailedReason,
			message:  fmt.Sprintf("Job %s has failed because it has reached the specified backoff limit", ji.name),
		})
		return err
	}
	if r.Config.EnableGangScheduling && !r.isPodGroupSynced(ji) {
		now := metav1.Now()
		ji.status.LastReconcileTime = &now
		return nil
	}
	if err = r.syncReplicas(ji); err != nil {
		return err
	}

	if err = r.Controller.UpdateJobStatus(ji.job, ji.rpls, ji.status); err != nil {
		hwlog.RunLog.Warnf("UpdateJobStatus error %v", err)
		return err
	}
	return nil
}

// UpdateJobStatus update job status which in cache
func (r *ASJobReconciler) UpdateJobStatus(
	job interface{},
	replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
	jobStatus *commonv1.JobStatus) error {
	if r == nil {
		return errors.New("nil pointer")
	}

	ascendJob, ok := job.(*mindxdlv1.AscendJob)
	if !ok {
		return fmt.Errorf("%v is not a type of AscendJob", ascendJob)
	}
	if jobStatus.StartTime == nil {
		now := metav1.Now()
		jobStatus.StartTime = &now
	}

	if err := r.updateSpecStatus(ascendJob, replicas, jobStatus); err != nil {
		return err
	}

	/*
		we assign the jobStatus to the msJob.Status for testing purpose
		it won't effect the main reconcile logic
		because we already use oldStatus := jobStatus.DeepCopy() to record the oldStatus
		and use !reflect.DeepEqual(*oldStatus, jobStatus) to decide whether to update the msJob or not
	*/
	ascendJob.Status = *jobStatus.DeepCopy()

	return nil
}

func (r *ASJobReconciler) checkMasterStatus(si specInfo) *conditionInfo {
	if si.status.Active > 0 {
		return &conditionInfo{
			condType: commonv1.JobRunning,
			reason:   util.JobRunningReason,
			message:  fmt.Sprintf("AscendJob %s/%s is running.", si.job.Namespace, si.job.Name),
		}
	}
	if *(si.spec.Replicas) == si.status.Succeeded {
		return &conditionInfo{
			condType: commonv1.JobSucceeded,
			reason:   util.JobRunningReason,
			message:  fmt.Sprintf("AscendJob<%s-%s> successfully completed.", si.job.Namespace, si.job.Name),
		}
	}

	return nil
}

func (r *ASJobReconciler) checkWorkerStatus(si specInfo) *conditionInfo {
	job := si.job

	if *(si.spec.Replicas) == si.status.Succeeded {
		return &conditionInfo{
			condType: commonv1.JobSucceeded,
			reason:   util.JobSucceededReason,
			message:  fmt.Sprintf("AscendJob <%s/%s> successfully completed.", job.Namespace, job.Name),
		}
	}
	if si.status.Active > 0 {
		return &conditionInfo{
			condType: commonv1.JobRunning,
			reason:   util.JobRunningReason,
			message:  fmt.Sprintf("AscendJob <%s/%s> is running.", job.Namespace, job.Name),
		}
	}
	return nil
}

func (r *ASJobReconciler) reconcileWithFailed(job *mindxdlv1.AscendJob, jobStatus *commonv1.JobStatus) *conditionInfo {
	restartCondition := getRestartCondition(jobStatus.Conditions)

	if restartCondition != nil {
		return &conditionInfo{
			condType: commonv1.JobRestarting,
			reason:   restartCondition.Reason,
			message:  restartCondition.Message,
		}
	}

	var ci = &conditionInfo{
		condType: commonv1.JobFailed,
		reason:   util.JobFailedReason,
		message: fmt.Sprintf("AscendJob <%s/%s> has failed because has pod failed.", job.Namespace,
			job.Name),
	}

	if !r.isUnconditionalRetryJob(job) {
		return ci
	}

	if rt, err := r.getJobRemainRetryTimes(job); err != nil {
		return ci
	} else if rt < 1 {
		ci.message = fmt.Sprintf("AscendJob <%s/%s> has failed because pod failed and remain retry times is 0.",
			job.Namespace, job.Name)
		return ci
	} else {
		return &conditionInfo{
			condType: commonv1.JobRestarting,
			reason:   util.JobRestartingReason,
			message: fmt.Sprintf("AscendJob <%s/%s> is unconditional retry job and remain retry times is <%d>.",
				job.Namespace, job.Name, rt),
		}
	}
}

func (r *ASJobReconciler) isLeaderRole(rType commonv1.ReplicaType) bool {
	return rType == mindxdlv1.MindSporeReplicaTypeScheduler ||
		rType == mindxdlv1.PytorchReplicaTypeMaster ||
		rType == mindxdlv1.TensorflowReplicaTypeChief
}

func (r *ASJobReconciler) updateSpecStatus(ascendJob *mindxdlv1.AscendJob,
	replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
	jobStatus *commonv1.JobStatus) error {

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

		r.recorder.Event(ascendJob, corev1.EventTypeNormal, ci.reason, ci.message)
		hwlog.RunLog.Infof("Append ascendJob<%s/%s> condition: %#v", ascendJob.Namespace, ascendJob.Name, ci)
		err := util.UpdateJobConditions(jobStatus, ci.condType, ci.reason, ci.message)
		if err != nil {
			hwlog.RunLog.Errorf("Append ascendJob<%s-%s> condition err: %v",
				ascendJob.Namespace, ascendJob.Name, err)
			return err
		}
		return nil
	}

	st := r.getJobStatus(ascendJob, replicas, jobStatus)

	return r.checkSpecStatus(ascendJob, st, jobStatus, updateFunc)
}
func (r *ASJobReconciler) checkSpecStatus(job *mindxdlv1.AscendJob, status *commonv1.ReplicaStatus,
	jobStatus *commonv1.JobStatus,
	updateFunc func(*conditionInfo) error) error {
	if status.Failed > 0 {
		return updateFunc(r.reconcileWithFailed(job, jobStatus))
	}
	if status.Active > 0 {
		return updateFunc(&conditionInfo{
			condType: commonv1.JobRunning,
			reason:   util.JobRunningReason,
			message:  fmt.Sprintf("AscendJob %s/%s is running.", job.Namespace, job.Name),
		})
	}
	if getTotalReplicas(job) == status.Succeeded {
		return updateFunc(&conditionInfo{
			condType: commonv1.JobSucceeded,
			reason:   util.JobRunningReason,
			message:  fmt.Sprintf("AscendJob<%s-%s> successfully completed.", job.Namespace, job.Name),
		})
	}
	return nil
}

func (r *ASJobReconciler) getJobStatus(ascendJob *mindxdlv1.AscendJob,
	replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
	jobStatus *commonv1.JobStatus) *commonv1.ReplicaStatus {
	st := &commonv1.ReplicaStatus{}
	for rtype, spec := range replicas {
		status := jobStatus.ReplicaStatuses[rtype]
		hwlog.RunLog.Infof("AscendJob=%s/%s, ReplicaType=%s expected=%d, running=%d, failed=%d",
			ascendJob.Namespace, ascendJob.Name, rtype, *(spec.Replicas)-status.Succeeded, status.Active, status.Failed)
		st.Succeeded += status.Succeeded
		st.Active += status.Active
		st.Failed += status.Failed
	}
	hwlog.RunLog.Infof("count ascendJob<%s/%s> status<%#v>", ascendJob.Namespace, ascendJob.Name, st)
	return st
}

func (r *ASJobReconciler) syncReplicas(ji *jobInfo) error {
	status := checkNonWorkerRplMountChips(ji)
	annotations := ji.mtObj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[nonWorkerPodMountChipStatus] = strconv.FormatBool(status)
	ji.mtObj.SetAnnotations(annotations)
	for rtype, spec := range ji.rpls {
		if err := r.Controller.ReconcilePods(ji.mtObj, ji.status, ji.pods, rtype, spec, ji.rpls); err != nil {
			hwlog.RunLog.Errorf("ReconcilePods type<%s> error %v", rtype, err)
			return err
		}
	}
	r.genRankTable(ji)
	return nil
}

func (r *ASJobReconciler) newPodGroupSpec(ji *jobInfo) v1beta1.PodGroupSpec {
	minMember := k8sutil.GetTotalReplicas(ji.rpls)
	queue := ""
	priorityClass := ""
	var minResources *corev1.ResourceList = nil

	runPolicy := ji.runPolicy

	if runPolicy.SchedulingPolicy != nil {
		if runPolicy.SchedulingPolicy.MinAvailable != nil {
			minMember = *runPolicy.SchedulingPolicy.MinAvailable
		}

		if runPolicy.SchedulingPolicy.Queue != "" {
			queue = runPolicy.SchedulingPolicy.Queue
		}

		if runPolicy.SchedulingPolicy.PriorityClass != "" {
			priorityClass = runPolicy.SchedulingPolicy.PriorityClass
		}

		if runPolicy.SchedulingPolicy.MinResources != nil {
			minResources = runPolicy.SchedulingPolicy.MinResources
		}
	}

	if minResources == nil {
		minResources = common.CalcPGMinResources(minMember, ji.rpls, r.PriorityClassLister.Get)
	}

	return v1beta1.PodGroupSpec{
		MinMember:         minMember,
		Queue:             queue,
		PriorityClassName: priorityClass,
		MinResources:      minResources,
	}
}

func (r *ASJobReconciler) isPodGroupSynced(ji *jobInfo) bool {
	pg, err := r.SyncPodGroup(ji.mtObj, r.newPodGroupSpec(ji))
	if err != nil {
		hwlog.RunLog.Warnf("Sync PodGroup %v: %v", ji.jobKey, err)
		return false
	}

	// Delay pods creation until podgroup status is inqueue
	if pg == nil || pg.Status.Phase == "" || pg.Status.Phase == v1beta1.PodGroupPending {
		hwlog.RunLog.Warnf("PodGroup %v unschedulable", ji.jobKey)
		return false
	}

	return true
}

func (r *ASJobReconciler) handleFinishedJob(ji *jobInfo, needUpdateCond bool, cond conditionInfo) error {
	// If the Job is succeed or failed, delete all pods and services.
	if err := r.DeletePodsAndServices(ji.runPolicy, ji.job, ji.pods); err != nil {
		hwlog.RunLog.Errorf("job<%s> delete pods and services failed, err: %s", ji.name, err)
		return err
	}

	if err := r.CleanupJob(ji.runPolicy, *ji.status, ji.job); err != nil {
		hwlog.RunLog.Errorf("clean up job<%s> failed, err: %s", ji.name, err)
		return err
	}

	if r.Config.EnableGangScheduling {
		r.Recorder.Event(ji.rtObj, corev1.EventTypeNormal, "JobTerminated", "Job has been terminated. Deleting PodGroup")
		if err := r.DeletePodGroup(ji.mtObj); err != nil {
			hwlog.RunLog.Errorf("delete pg failed, err: %s", err)
			r.Recorder.Eventf(ji.rtObj, corev1.EventTypeWarning, "FailedDeletePodGroup", "Error deleting: %v", err)
			return err
		} else {
			r.Recorder.Eventf(ji.rtObj, corev1.EventTypeNormal, "SuccessfulDeletePodGroup", "Deleted PodGroup: %v", ji.name)
		}
	}

	if needUpdateCond {
		r.Recorder.Event(ji.rtObj, corev1.EventTypeNormal, cond.reason, cond.message)
		if err := util.UpdateJobConditions(ji.status, cond.condType, cond.reason, cond.message); err != nil {
			hwlog.RunLog.Errorf("Append job condition error: %v", err)
			return err
		}
	}

	// At this point the pods may have been deleted.
	// 1) If the job succeeded, we manually set the replica status.
	// 2) If any replicas are still active, set their status to succeeded.
	if util.IsSucceeded(*ji.status) {
		for rtype := range ji.status.ReplicaStatuses {
			ji.status.ReplicaStatuses[rtype].Succeeded += ji.status.ReplicaStatuses[rtype].Active
			ji.status.ReplicaStatuses[rtype].Active = 0
		}
	}
	return nil
}

// SyncPodGroup synchronizes the PodGroup
func (r *ASJobReconciler) SyncPodGroup(job metav1.Object, pgSpec v1beta1.PodGroupSpec) (*v1beta1.PodGroup, error) {
	if r == nil {
		return nil, errors.New("nil pointer")
	}

	if pg, err := r.getPodGroup(job); err == nil {
		return pg, nil
	}

	hwlog.RunLog.Debugf("get job<%s/%s> pg failed, try to create pg", job.GetNamespace(), job.GetName())
	return r.createPodGroup(job, pgSpec)
}

func (r *ASJobReconciler) getPodGroup(job metav1.Object) (*v1beta1.PodGroup, error) {
	pgName := job.GetName() + "-" + string(job.GetUID())

	// Check whether podGroup exists or not
	return r.VolcanoClientSet.SchedulingV1beta1().PodGroups(job.GetNamespace()).Get(context.TODO(), pgName,
		metav1.GetOptions{})
}

func (r *ASJobReconciler) createPodGroup(job metav1.Object, pgSpec v1beta1.PodGroupSpec) (*v1beta1.PodGroup, error) {
	pgName := job.GetName() + "-" + string(job.GetUID())
	// create podGroup for gang scheduling by volcano
	createPodGroup := &v1beta1.PodGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:        pgName,
			Namespace:   job.GetNamespace(),
			Annotations: job.GetAnnotations(),
			Labels:      job.GetLabels(),
			OwnerReferences: []metav1.OwnerReference{
				*r.GenOwnerReference(job),
			},
		},
		Spec: pgSpec,
	}
	createdPodGroup, err := r.VolcanoClientSet.SchedulingV1beta1().PodGroups(job.GetNamespace()).Create(context.TODO(),
		createPodGroup, metav1.CreateOptions{})
	if err != nil {
		return createdPodGroup, fmt.Errorf("unable to create PodGroup: %v", err)
	}
	return createdPodGroup, nil
}

// DeletePodGroup delete PodGroup
func (r *ASJobReconciler) DeletePodGroup(job metav1.Object) error {
	if r == nil {
		return errors.New("nil pointer")
	}

	volcanoClientSet := r.VolcanoClientSet
	pgName := job.GetName() + "-" + string(job.GetUID())
	// Check whether podGroup exists or not
	_, err := volcanoClientSet.SchedulingV1beta1().PodGroups(job.GetNamespace()).Get(context.TODO(),
		pgName, metav1.GetOptions{})
	if err != nil && k8serrors.IsNotFound(err) {
		return nil
	}

	hwlog.RunLog.Infof("Deleting PodGroup %s", pgName)

	// Delete podGroup
	err = volcanoClientSet.SchedulingV1beta1().PodGroups(job.GetNamespace()).Delete(context.TODO(),
		pgName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("unable to delete PodGroup: %v", err)
	}
	return nil
}
