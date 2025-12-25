/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	flowcommon "github.com/kubeflow/common/pkg/controller.v1/common"
	"github.com/kubeflow/common/pkg/util"
	"github.com/kubeflow/common/pkg/util/labels"
	"golang.org/x/time/rate"
	v1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"ascend-common/common-utils/hwlog"
	mindxdlv1 "rl-operator/pkg/api/v1"
	"rl-operator/pkg/common"
)

type RayWorkerReconciler struct {
	*JobReconciler
	EnableGangScheduling bool
	Started              bool
}

func NewRayWorkerReconciler(mgr manager.Manager, enableGangScheduling bool) *RayWorkerReconciler {
	baseReconciler := NewJobReconciler(mgr, common.RayWorkerControllerName)
	r := &RayWorkerReconciler{
		JobReconciler:        baseReconciler,
		EnableGangScheduling: enableGangScheduling,
	}
	r.ConfigInfo = r
	return r
}

func (r *RayWorkerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return fmt.Errorf("RayHeadReconciler nil pointer")
	}

	c, err := controller.New(common.RayWorkerControllerName, mgr, controller.Options{
		Reconciler: r,
		RateLimiter: workqueue.NewMaxOfRateLimiter(
			workqueue.NewItemExponentialFailureRateLimiter(common.WorkQueueBaseDelay, common.WorkQueueMaxDelay),
			&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(common.WorkQueueQps), common.WorkQueueBurst)},
		),
	})
	if err != nil {
		return err
	}

	return r.watchRelatedResource(c)
}

func (r *RayWorkerReconciler) watchRelatedResource(c controller.Controller) error {
	if err := c.Watch(&source.Kind{Type: &mindxdlv1.RayWorker{}}, &handler.EnqueueRequestForObject{},
		predicate.Funcs{CreateFunc: r.onOwnerCreateFunc(), DeleteFunc: r.onOwnerDeleteFunc()},
	); err != nil {
		return err
	}
	resourceOptions := []*common.ResourceOption{
		{Kind: &source.Kind{Type: &v1.Pod{}},
			PredicateFunc: predicate.Funcs{DeleteFunc: r.onPodDeleteFunc()}},
	}

	for _, src := range resourceOptions {
		if err := c.Watch(src.Kind, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &mindxdlv1.RayWorker{},
		}, src.PredicateFunc); err != nil {
			return err
		}
	}
	return nil
}

func (r *RayWorkerReconciler) GetAPIGroupVersionKind() schema.GroupVersionKind {
	return mindxdlv1.GroupVersion.WithKind(common.RayWorkerResourceKind)
}

func (r *RayWorkerReconciler) GetAPIGroupVersion() schema.GroupVersion {
	return mindxdlv1.GroupVersion
}

func (r *RayWorkerReconciler) GetRunPolicy(obj client.Object) (*commonv1.RunPolicy, error) {
	rayWorker, ok := obj.(*mindxdlv1.RayWorker)
	if !ok {
		return nil, fmt.Errorf("expect RayWorker, but got %v", reflect.TypeOf(obj))
	}
	return &rayWorker.Spec.RunPolicy, nil
}

func (r *RayWorkerReconciler) GetStatus(obj client.Object) (*commonv1.JobStatus, error) {
	rayWorker, ok := obj.(*mindxdlv1.RayWorker)
	if !ok {
		return nil, fmt.Errorf("expect RayWorker, but got %v", reflect.TypeOf(obj))
	}
	return &rayWorker.Status, nil
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the RayCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
func (r *RayWorkerReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if r == nil {
		return ctrl.Result{}, fmt.Errorf("RayWorkerReconciler nil pointer")
	}
	rayWorker := &mindxdlv1.RayWorker{}
	if err := r.FetchResource(ctx, req.NamespacedName, rayWorker); err != nil {
		hwlog.RunLog.Warn("RayWorkerReconciler object is empty. Skip reconcile.")
		return ctrl.Result{}, err
	}
	if common.IsObjectEmpty(rayWorker) {
		return ctrl.Result{}, nil
	}

	if rayWorker.GetDeletionTimestamp() != nil {
		hwlog.RunLog.Infof("reconcile cancelled，job<%s> has been deleted", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	oldStatus := rayWorker.Status.DeepCopy()
	defer func() {
		var err error
		if !reflect.DeepEqual(*oldStatus, rayWorker.Status) {
			if rayWorker.Status.ReplicaStatuses == nil {
				rayWorker.Status.ReplicaStatuses = map[commonv1.ReplicaType]*commonv1.ReplicaStatus{}
			}
			err = r.UpdateResourceStatus(rayWorker)
		}
		if err != nil {
			r.Recorder.Eventf(rayWorker, v1.EventTypeWarning, common.StatusUpdateFailedReason, err.Error())
			hwlog.RunLog.Warnf("failed to update RayWorker status in api server, err: %s", err)
		}
	}()

	err := r.ReconcileRayWorker(rayWorker)
	if err == nil {
		return ctrl.Result{}, nil
	}
	if k8serr.IsConflict(err) || common.IfNeedRequeue(err) {
		return ctrl.Result{Requeue: true}, nil
	}
	r.Recorder.Eventf(rayWorker, v1.EventTypeWarning, util.JobFailedReason, err.Error())
	hwlog.RunLog.Errorf("Reconcile RayWorker<%s> failed, %s", req.NamespacedName, err)
	err = util.UpdateJobConditions(&rayWorker.Status, commonv1.JobFailed,
		common.StatusUpdateFailedReason, err.Error())
	if err != nil {
		hwlog.RunLog.Errorf("failed to update job status conditions: %v", err)
	}
	return ctrl.Result{}, nil
}

func (r *RayWorkerReconciler) ReconcileRayWorker(rayWorker *mindxdlv1.RayWorker) error {
	r.Mutex.RLock()
	version, ok := r.Versions[rayWorker.GetUID()]
	backoffLimit, backoffLimitOk := r.BackoffLimits[rayWorker.GetUID()]
	r.Mutex.RUnlock()
	if !ok || (backoffLimitOk && backoffLimit > 0 && version > backoffLimit) {
		msg := fmt.Sprintf("RayWorker %s has failed because it has reached the specified backoff limit",
			rayWorker.Name)
		hwlog.RunLog.Warn(msg)
		err := util.UpdateJobConditions(&rayWorker.Status, commonv1.JobFailed, util.JobFailedReason, msg)
		return err
	}

	isRunning, err := r.ReconcileReplicas(rayWorker)
	if err != nil {
		return err
	}
	if isRunning {
		if !r.Started {
			// first running
			r.Started = true
			hwlog.RunLog.Infof("RayWorker %s finish starting", rayWorker.GetName())
		}
		if err := util.UpdateJobConditions(&rayWorker.Status, commonv1.JobRunning,
			util.JobRunningReason, ""); err != nil {
			hwlog.RunLog.Errorf("failed to update job status conditions: %v", err)
			return err
		}
	} else if r.Started {
		// maybe some of the pods are restarting
		hwlog.RunLog.Infof("RayWorker %s is restarting some pods", rayWorker.GetName())
		if err := util.UpdateJobConditions(&rayWorker.Status, commonv1.JobRestarting,
			util.JobRestartingReason, "reschedule ray worker pod"); err != nil {
			hwlog.RunLog.Errorf("failed to update job status conditions: %v", err)
			return err
		}
	}
	return nil
}

func (r *RayWorkerReconciler) ReconcileReplicas(rayWorker *mindxdlv1.RayWorker) (bool, error) {
	podList := &v1.PodList{}
	if err := r.FetchChildrenForResource(rayWorker, podList, common.RayWorkerIdentification); err != nil {
		hwlog.RunLog.Errorf("list PodList of resource<%s/%s> error: %v", rayWorker.GetNamespace(),
			rayWorker.GetName(), err)
		return false, err
	}
	hwlog.RunLog.Debugf("RayWorker pods: %v", len(podList.Items))
	pods := common.ConvertToPointerSlice(podList.Items)

	replica := rayWorker.Spec.ReplicaSpecs
	isRunning := true
	for rType, spec := range replica {
		common.InitializeReplicaStatuses(&rayWorker.Status, rType)
		pi, err := r.newPodInfo(rayWorker, rType, spec)
		if err != nil {
			hwlog.RunLog.Errorf("create PodInfo<%s> error: %v", rType, err)
			return false, err
		}
		rtypeStr := strings.ToLower(string(rType))
		filterPods := common.FilterPodsByReplicaType(pods, rtypeStr)
		isReplicaRunning, err := r.ReconcilePods(pi, filterPods, rayWorker)
		isRunning = isRunning && isReplicaRunning
		if err != nil {
			return false, err
		}
	}
	return isRunning, nil
}

func (r *RayWorkerReconciler) ReconcilePods(pi *PodInfo, pods []*v1.Pod, rayWorker *mindxdlv1.RayWorker) (bool, error) {
	// GetPodSlices will return enough information here to make decision to add/remove/update resources.
	//
	// For example, let's assume we have pods with replica-index 0, 1, 2
	// If replica is 4, return a slice with size 4. [[0],[1],[2],[]], a pod with replica-index 3 will be created.
	//
	// If replica is 1, return a slice with size 3. [[0],[1],[2]], pod with replica-index 1 and 2 are out of range
	// and will be deleted.
	podSlices := r.GetPodSlices(pods, int(*pi.spec.Replicas))
	isRunning := true

	var podToCreate []*PodInfo
	for index, podSlice := range podSlices {
		if len(podSlice) > 1 {
			hwlog.RunLog.Warnf("We have too many pods for %s %d", pi.rtype, index)
			isRunning = false
		} else if len(podSlice) == 0 {
			hwlog.RunLog.Debugf("Need to create new pod: %s-%d", pi.rtype, index)
			p := pi.DeepCopy()
			p.index = index
			podToCreate = append(podToCreate, p)
			isRunning = false
		} else {
			hwlog.RunLog.Debugf("Need to check pod: %s-%d", pi.rtype, index)
			isPodRunning, err := r.checkExistPod(pi, index, podSlice[0], rayWorker)
			isRunning = isPodRunning && isRunning
			if err != nil {
				return isRunning, err
			}
		}
	}

	return isRunning, r.createPods(podToCreate, rayWorker)
}

// GetPodSlices sorts the list of pods by label
func (r *RayWorkerReconciler) GetPodSlices(pods []*v1.Pod, replicas int) map[int][]*v1.Pod {
	if r == nil {
		return nil
	}
	podSlices := make(map[int][]*v1.Pod)
	for i := 0; i < replicas; i++ {
		podSlices[i] = nil
	}

	for _, pod := range pods {
		index, err := labels.ReplicaIndex(pod.Labels)
		if err != nil {
			hwlog.RunLog.Warnf("Error obtaining replica index from Pod %s/%s: %v", pod.Namespace, pod.Name, err)
			continue
		}
		podSlices[index] = append(podSlices[index], pod)
	}
	return podSlices
}

func (r *RayWorkerReconciler) checkExistPod(pi *PodInfo, index int, pod *v1.Pod, rayWorker *mindxdlv1.RayWorker) (bool, error) {
	// check if the index is in the valid range, if not, we should kill the pod
	if index < 0 || index >= int(*pi.spec.Replicas) {
		if err := r.PodControl.DeletePod(pod.Namespace, pod.Name, rayWorker); err != nil {
			return false, err
		}
	}
	if pod.GetDeletionTimestamp() != nil {
		hwlog.RunLog.Warnf("Pod %s/%s is being deleted", pod.Namespace, pod.Name)
		return false, &common.ReQueueError{
			Message: fmt.Sprintf("Pod %s/%s is Terminating, try to requeue", pod.Namespace, pod.Name),
		}
	}
	if err := r.SetRayEnv(pod, rayWorker.Name); err != nil {
		return false, err
	}
	isRunning := pod.Status.Phase == v1.PodRunning
	// the pod of ray node should never succeed or failed
	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		hwlog.RunLog.Infof("Pod %s/%s status is failed or unexpectedly completed, will restart it",
			pod.Namespace, pod.Name)
		if r.IsReachBackoffLimit(rayWorker) {
			return false, fmt.Errorf("RayWorker pod<%s> reach backoff limit, it will not restart again",
				rayWorker.Name)
		}
		if err := r.PodControl.DeletePod(pod.Namespace, pod.Name, rayWorker); err != nil {
			hwlog.RunLog.Warnf("error delete Pod<%s> of RayHead when restarting: %v", pod.Name, err)
			return false, err
		}
	}
	common.UpdateJobReplicaStatuses(&rayWorker.Status, pi.rtype, pod)
	if pod.Status.Phase != v1.PodRunning {
		hwlog.RunLog.Debugf("expect pod<%s> of RayHead is running, but %s", pod.Name, pod.Status.Phase)
	}
	return isRunning, nil
}

func (r *RayWorkerReconciler) createPods(pods []*PodInfo, rayWorker *mindxdlv1.RayWorker) error {
	if len(pods) == 0 {
		return nil
	}

	appendMutex := sync.RWMutex{}
	var createErr []error
	appendErr := func(err error) {
		appendMutex.Lock()
		defer appendMutex.Unlock()
		createErr = append(createErr, err)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(pods))
	start := time.Now()
	for _, pInfo := range pods {
		go func(p *PodInfo) {
			defer wg.Done()
			if err := r.createNewPod(p, rayWorker); err != nil {
				appendErr(err)
			}
		}(pInfo)
	}
	wg.Wait()
	hwlog.RunLog.Infof("create job all pods use time (%v)", time.Since(start))
	if len(createErr) > 0 {
		return fmt.Errorf("failed to create pods: %v", createErr)
	}
	return nil
}

func (r *RayWorkerReconciler) createNewPod(pi *PodInfo, rayWorker *mindxdlv1.RayWorker) error {
	podTemplate, err := r.createPodSpec(pi, rayWorker)
	if err != nil {
		return err
	}
	err = r.PodControl.CreatePodsWithControllerRef(rayWorker.Namespace, podTemplate,
		rayWorker, r.GenOwnerReference(rayWorker))
	if err != nil && k8serr.IsTimeout(err) {
		// Pod is created but its initialization has timed out.
		// If the initialization is successful eventually, the
		// controller will observe the creation via the informer.
		// If the initialization fails, or if the pod keeps
		// uninitialized for a long time, the informer will not
		// receive any update, and the controller will create a new
		// pod when the expectation expires.
		return nil
	} else if err != nil {
		// Decrement the expected number of creates because the informer won't observe this pod
		hwlog.RunLog.Debugf("Failed creation, decrementing expectations for rayWorker %s/%s",
			rayWorker.Namespace, rayWorker.Name)
		return err
	}
	hwlog.RunLog.Infof("create pod: %s/%s success", rayWorker.Namespace, podTemplate.Name)
	return nil
}

func (r *RayWorkerReconciler) createPodSpec(pi *PodInfo,
	rayWorker *mindxdlv1.RayWorker) (*v1.PodTemplateSpec, error) {
	if rayWorker.Spec.ReplicaSpecs == nil {
		return nil, fmt.Errorf("job or job specs is nil")
	}

	podTemplate := pi.spec.Template.DeepCopy()
	indexStr := strconv.Itoa(pi.index)
	podTemplate.Name = flowcommon.GenGeneralName(rayWorker.Name, strings.ToLower(string(pi.rtype)), indexStr)
	r.setPodLabels(rayWorker, podTemplate, pi.rtype, indexStr)
	common.SetCommonEnv(podTemplate)
	common.SetVolumes(podTemplate, rayWorker.Name)
	podTemplate.Annotations = common.DeepCopyMap(rayWorker.Annotations)
	setWorkerPodCmdArgs(podTemplate, pi, rayWorker.Labels)
	r.SetRestartPolicy(rayWorker, podTemplate, pi.spec.RestartPolicy)

	// if gang-scheduling is enabled:
	// 1. if user has specified other scheduler, we report a warning without overriding any fields.
	// 2. if no SchedulerName is set for pods, then we set the SchedulerName to "volcano".
	if r.EnableGangScheduling {
		rtypeStr := strings.ToLower(string(pi.rtype))
		r.setGangScheduleInfo(rayWorker, podTemplate, rayWorker.Spec.ReplicaSpecs, rtypeStr)
	}
	return podTemplate, nil
}

func (r *RayWorkerReconciler) setPodLabels(job *mindxdlv1.RayWorker, podTemplate *v1.PodTemplateSpec,
	rt commonv1.ReplicaType, index string) {
	labelsMap := r.GenLabels(job, common.RayWorkerIdentification)
	rtypeStr := strings.ToLower(string(rt))
	labels.SetReplicaType(labelsMap, rtypeStr)
	labels.SetReplicaIndexStr(labelsMap, index)

	if podTemplate.Labels == nil {
		podTemplate.Labels = make(map[string]string)
	}
	for key, value := range labelsMap {
		podTemplate.Labels[key] = value
	}
	r.SetCommonPodLabel(podTemplate, job)
}

func (r *RayWorkerReconciler) setGangScheduleInfo(job *mindxdlv1.RayWorker, podTemplate *v1.PodTemplateSpec,
	replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec, rType string) {
	jobSchedulerName := job.Spec.SchedulerName
	if len(jobSchedulerName) == 0 || jobSchedulerName == common.GangSchedulerName {
		jobSchedulerName = common.GangSchedulerName
	} else {
		errMsg := "Another job scheduler is specified when gang-scheduling is enabled and it will not be overwritten"
		hwlog.RunLog.Warn(errMsg)
		r.Recorder.Event(job, v1.EventTypeWarning, common.JobSchedulerNameReason, errMsg)
	}
	podSchedulerName := common.GetSchedulerName(replicas)
	if len(podSchedulerName) == 0 {
		podTemplate.Spec.SchedulerName = jobSchedulerName
	} else if strings.Compare(podSchedulerName, common.GangSchedulerName) != 0 {
		errMsg := "Another scheduler is specified when gang-scheduling is enabled and it will not be overwritten"
		hwlog.RunLog.Warn(errMsg)
		r.Recorder.Event(job, v1.EventTypeWarning, common.PodTemplateSchedulerNameReason, errMsg)
	}
	if podTemplate.Annotations == nil {
		podTemplate.Annotations = map[string]string{}
	}
	podTemplate.Annotations[common.GangSchedulingPodGroupAnnotation] =
		common.GetPodGroupName(job.Namespace, job.Name, rType)
	podTemplate.Annotations[common.VolcanoTaskSpecKey] = rType
}
