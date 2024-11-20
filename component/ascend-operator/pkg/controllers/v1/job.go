/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package controllers

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/controller.v1/common"
	"github.com/kubeflow/common/pkg/controller.v1/control"
	"github.com/kubeflow/common/pkg/core"
	commonutil "github.com/kubeflow/common/pkg/util"
	"github.com/kubeflow/common/pkg/util/k8sutil"
	"github.com/kubeflow/training-operator/pkg/common/util"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
)

func (r *ASJobReconciler) ReconcileJobs(
	job interface{},
	replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
	jobStatus commonv1.JobStatus,
	runPolicy *commonv1.RunPolicy) error {
	if r == nil {
		return errors.New("nil pointer")
	}

	hwlog.RunLog.Debugf("start reconcile AscendJob, job status: %v, runpolicy: %v", jobStatus, runPolicy)

	metaObject, ok := job.(metav1.Object)
	if !ok {
		return fmt.Errorf("job is not of type metav1.Object")
	}
	jobName := metaObject.GetName()

	runtimeObject, ok := job.(runtime.Object)
	if !ok {
		return fmt.Errorf("job is not of type runtime.Object")
	}
	jobKey, err := common.KeyFunc(job)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for job object %#v: %v", job, err))
		return err
	}
	// Reset expectations
	// 1. Since `ReconcileJobs` is called, we expect that previous expectations are all satisfied,
	//    and it's safe to reset the expectations
	// 2. Reset expectations can avoid dirty data such as `expectedDeletion = -1`
	//    (pod or service was deleted unexpectedly)
	r.ResetExpectations(jobKey, replicas)

	hwlog.RunLog.Infof("Reconciling for job %s", metaObject.GetName())
	pods, err := r.Controller.GetPodsForJob(job)
	if err != nil {
		hwlog.RunLog.Warnf("GetPodsForJob error %v", err)
		return err
	}

	services, err := r.Controller.GetServicesForJob(job)
	if err != nil {
		hwlog.RunLog.Warnf("GetServicesForJob error %v", err)
		return err
	}

	oldStatus := jobStatus.DeepCopy()
	if commonutil.IsSucceeded(jobStatus) || commonutil.IsFailed(jobStatus) {
		// If the Job is succeed or failed, delete all pods and services.
		if err := r.DeletePodsAndServices(runPolicy, job, pods); err != nil {
			hwlog.RunLog.Errorf("job<%s> delete pods and services failed, err: %s", jobName, err)
			return err
		}

		if err := r.CleanupJob(runPolicy, jobStatus, job); err != nil {
			hwlog.RunLog.Errorf("clean up job<%s> failed, err: %s", jobName, err)
			return err
		}

		if r.Config.EnableGangScheduling {
			r.Recorder.Event(runtimeObject, corev1.EventTypeNormal, "JobTerminated", "Job has been terminated. Deleting PodGroup")
			if err := r.DeletePodGroup(metaObject); err != nil {
				hwlog.RunLog.Errorf("delete pg failed, err: %s", err)
				r.Recorder.Eventf(runtimeObject, corev1.EventTypeWarning, "FailedDeletePodGroup", "Error deleting: %v", err)
				return err
			} else {
				r.Recorder.Eventf(runtimeObject, corev1.EventTypeNormal, "SuccessfulDeletePodGroup", "Deleted PodGroup: %v", jobName)
			}
		}

		// At this point the pods may have been deleted.
		// 1) If the job succeeded, we manually set the replica status.
		// 2) If any replicas are still active, set their status to succeeded.
		if commonutil.IsSucceeded(jobStatus) {
			for rtype := range jobStatus.ReplicaStatuses {
				jobStatus.ReplicaStatuses[rtype].Succeeded += jobStatus.ReplicaStatuses[rtype].Active
				jobStatus.ReplicaStatuses[rtype].Active = 0
			}
		}

		// No need to update the job status if the status hasn't changed since last time.
		if !reflect.DeepEqual(*oldStatus, jobStatus) {
			return r.Controller.UpdateJobStatusInApiServer(job, &jobStatus)
		}

		return nil
	}

	// retrieve the previous number of retry
	activePods := k8sutil.FilterActivePods(pods)

	core.RecordAbnormalPods(activePods, runtimeObject, r.Recorder)

	totalReplicas := k8sutil.GetTotalReplicas(replicas)
	var failureMessage string
	jobExceedsLimit := false
	exceedsBackoffLimit := false

	version, ok := r.versions[metaObject.GetUID()]
	backoffLimit, backoffLimitOk := r.backoffLimits[metaObject.GetUID()]
	if !ok || (backoffLimitOk && backoffLimit > 0 && version > backoffLimit) {
		exceedsBackoffLimit = true
		hwlog.RunLog.Warnf("Job %s has failed because it has reached the specified backoff limit", jobName)
	}

	if exceedsBackoffLimit {
		// check if the number of pod restart exceeds backoff (for restart OnFailure only)
		// OR if the number of failed jobs increased since the last syncJob
		jobExceedsLimit = true
		failureMessage = fmt.Sprintf("Job %s has failed because it has reached the specified backoff limit", jobName)
	} else if r.PastActiveDeadline(runPolicy, jobStatus) {
		failureMessage = fmt.Sprintf("Job %s has failed because it was active longer than specified deadline", jobName)
		jobExceedsLimit = true
	}

	if jobExceedsLimit {
		// Set job completion time before resource cleanup
		if jobStatus.CompletionTime == nil {
			now := metav1.Now()
			jobStatus.CompletionTime = &now
		}

		// If the Job exceeds backoff limit or is past active deadline
		// delete all pods and services, then set the status to failed
		if err := r.DeletePodsAndServices(runPolicy, job, pods); err != nil {
			hwlog.RunLog.Errorf("job<%s> delete pods and services failed, err: %s", jobName, err)
			return err
		}

		if err := r.CleanupJob(runPolicy, jobStatus, job); err != nil {
			hwlog.RunLog.Errorf("clean up job<%s> failed, err: %s", jobName, err)
			return err
		}

		if r.Config.EnableGangScheduling {
			r.Recorder.Event(runtimeObject, corev1.EventTypeNormal, "JobTerminated", "Job has been terminated. Deleting PodGroup")
			if err := r.DeletePodGroup(metaObject); err != nil {
				hwlog.RunLog.Errorf("delete pg failed, err: %s", err)
				r.Recorder.Eventf(runtimeObject, corev1.EventTypeWarning, "FailedDeletePodGroup", "Error deleting: %v", err)
				return err
			} else {
				r.Recorder.Eventf(runtimeObject, corev1.EventTypeNormal, "SuccessfulDeletePodGroup", "Deleted PodGroup: %v", jobName)
			}
		}

		r.Recorder.Event(runtimeObject, corev1.EventTypeNormal, commonutil.JobFailedReason, failureMessage)

		if err := commonutil.UpdateJobConditions(&jobStatus, commonv1.JobFailed, commonutil.JobFailedReason, failureMessage); err != nil {
			hwlog.RunLog.Errorf("Append job condition error: %v", err)
			return err
		}

		return r.Controller.UpdateJobStatusInApiServer(job, &jobStatus)
	} else {
		// General cases which need to reconcile
		if r.Config.EnableGangScheduling {
			minMember := totalReplicas
			queue := ""
			priorityClass := ""
			var minResources *corev1.ResourceList

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
				minResources = common.CalcPGMinResources(minMember, replicas, r.PriorityClassLister.Get)
			}

			pgSpec := v1beta1.PodGroupSpec{
				MinMember:         minMember,
				Queue:             queue,
				PriorityClassName: priorityClass,
				MinResources:      minResources,
			}

			syncReplicas := true
			pg, err := r.SyncPodGroup(metaObject, pgSpec)
			if err != nil {
				hwlog.RunLog.Warnf("Sync PodGroup %v: %v", jobKey, err)
				syncReplicas = false
			}

			// Delay pods creation until podgroup status is inqueue
			if pg == nil || pg.Status.Phase == "" || pg.Status.Phase == v1beta1.PodGroupPending {
				hwlog.RunLog.Warnf("PodGroup %v unschedulable", jobKey)
				syncReplicas = false
			}

			if !syncReplicas {
				now := metav1.Now()
				jobStatus.LastReconcileTime = &now

				// Update job status here to trigger a new reconciliation
				return r.Controller.UpdateJobStatusInApiServer(job, &jobStatus)
			}
		}

		for rtype, spec := range replicas {
			if err = r.Controller.ReconcileServices(metaObject, services, rtype, spec); err != nil {
				hwlog.RunLog.Errorf("ReconcileServices error %v", err)
				return err
			}

			if err = r.Controller.ReconcilePods(metaObject, &jobStatus, pods, rtype, spec, replicas); err != nil {
				hwlog.RunLog.Errorf("ReconcilePods error %v", err)
				return err
			}
		}
	}

	err = r.Controller.UpdateJobStatus(job, replicas, &jobStatus)
	if err != nil {
		hwlog.RunLog.Warnf("UpdateJobStatus error %v", err)
		return err
	}
	// No need to update the job status if the status hasn't changed since last time.
	if !reflect.DeepEqual(*oldStatus, jobStatus) {
		return r.Controller.UpdateJobStatusInApiServer(job, &jobStatus)
	}
	return nil
}

func (r *ASJobReconciler) SyncPodGroup(job metav1.Object, pgSpec v1beta1.PodGroupSpec) (*v1beta1.PodGroup, error) {
	if r == nil {
		return nil, errors.New("nil pointer")
	}

	pgName := job.GetName() + "-" + string(job.GetUID())

	volcanoClientSet := r.VolcanoClientSet
	// Check whether podGroup exists or not
	podGroup, err := volcanoClientSet.SchedulingV1beta1().PodGroups(job.GetNamespace()).Get(context.TODO(), pgName, metav1.GetOptions{})
	if err == nil {
		return podGroup, nil
	}

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
	createdPodGroup, err := volcanoClientSet.SchedulingV1beta1().PodGroups(job.GetNamespace()).Create(context.TODO(), createPodGroup, metav1.CreateOptions{})
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
	_, err := volcanoClientSet.SchedulingV1beta1().PodGroups(job.GetNamespace()).Get(context.TODO(), pgName, metav1.GetOptions{})
	if err != nil && k8serrors.IsNotFound(err) {
		return nil
	}

	hwlog.RunLog.Infof("Deleting PodGroup %s", pgName)

	// Delete podGroup
	err = volcanoClientSet.SchedulingV1beta1().PodGroups(job.GetNamespace()).Delete(context.TODO(), pgName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("unable to delete PodGroup: %v", err)
	}
	return nil
}

// GetPodsForJob returns the set of pods that this job should manage.
// It also reconciles ControllerRef by adopting/orphaning.
// Note that the returned Pods are pointers into the cache.
func (r *ASJobReconciler) GetPodsForJob(jobObject interface{}) ([]*corev1.Pod, error) {
	if r == nil {
		return nil, errors.New("nil pointer")
	}

	job, ok := jobObject.(metav1.Object)
	if !ok {
		return nil, fmt.Errorf("job is not of type metav1.Object")
	}

	// Create selector.
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: r.GenLabels(job.GetName()),
	})

	if err != nil {
		hwlog.RunLog.Errorf("couldn't convert Job selector: %s", err)
		return nil, fmt.Errorf("couldn't convert Job selector: %v", err)
	}
	// List all pods to include those that don't match the selector anymore
	// but have a ControllerRef pointing to this controller.
	podlist := &corev1.PodList{}
	err = r.List(context.Background(), podlist,
		client.MatchingLabelsSelector{Selector: selector}, client.InNamespace(job.GetNamespace()))
	if err != nil {
		hwlog.RunLog.Errorf("list job<%s> pods failed: %s", job.GetName(), err)
		return nil, err
	}

	pods := util.ConvertPodList(podlist.Items)

	// If any adoptions are attempted, we should first recheck for deletion
	// with an uncached quorum read sometime after listing Pods (see #42639).
	canAdoptFunc := common.RecheckDeletionTimestamp(func() (metav1.Object, error) {
		fresh, err := r.Controller.GetJobFromAPIClient(job.GetNamespace(), job.GetName())
		if err != nil {
			hwlog.RunLog.Errorf("get job<%s> for api client failed, err: %s", job.GetName(), err)
			return nil, err
		}
		if fresh.GetUID() != job.GetUID() {
			return nil, fmt.Errorf("original Job %v/%v is gone: got uid %v, wanted %v", job.GetNamespace(), job.GetName(), fresh.GetUID(), job.GetUID())
		}
		return fresh, nil
	})
	cm := control.NewPodControllerRefManager(r.PodControl, job, selector, r.Controller.GetAPIGroupVersionKind(), canAdoptFunc)
	return cm.ClaimPods(pods)
}

// GetServicesForJob returns the set of services that this job should manage.
// It also reconciles ControllerRef by adopting/orphaning.
// Note that the returned services are pointers into the cache.
func (r *ASJobReconciler) GetServicesForJob(jobObject interface{}) ([]*corev1.Service, error) {
	if r == nil {
		return nil, errors.New("nil pointer")
	}

	job, ok := jobObject.(metav1.Object)
	if !ok {
		return nil, fmt.Errorf("job is not of type metav1.Object")
	}

	// Create selector
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: r.GenLabels(job.GetName()),
	})

	if err != nil {
		return nil, fmt.Errorf("couldn't convert Job selector: %v", err)
	}
	// List all services to include those that don't match the selector anymore
	// but have a ControllerRef pointing to this controller.
	svclist := &corev1.ServiceList{}
	err = r.List(context.Background(), svclist,
		client.MatchingLabelsSelector{Selector: selector}, client.InNamespace(job.GetNamespace()))
	if err != nil {
		return nil, fmt.Errorf("couldn't get Service: %v", err)
	}

	// If any adoptions are attempted, we should first recheck for deletion
	// with an uncached quorum read sometime after listing services (see #42639).
	canAdoptFunc := common.RecheckDeletionTimestamp(func() (metav1.Object, error) {
		fresh, err := r.GetJobFromInformerCache(job.GetNamespace(), job.GetName())
		if err != nil {
			return nil, err
		}
		if fresh.GetUID() != job.GetUID() {
			return nil, fmt.Errorf("original Job %v/%v is gone: got uid %v, wanted %v", job.GetNamespace(), job.GetName(), fresh.GetUID(), job.GetUID())
		}
		return fresh, nil
	})
	cm := control.NewServiceControllerRefManager(r.ServiceControl, job, selector, r.Controller.GetAPIGroupVersionKind(), canAdoptFunc)

	services := util.ConvertServiceList(svclist.Items)
	return cm.ClaimServices(services)
}
