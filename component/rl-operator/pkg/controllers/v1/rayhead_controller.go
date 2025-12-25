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
	"strings"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	flowcommon "github.com/kubeflow/common/pkg/controller.v1/common"
	"github.com/kubeflow/common/pkg/util"
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

type RayHeadReconciler struct {
	*JobReconciler
	EnableGangScheduling bool
	Started              bool
}

// NewRayHeadReconciler new reconciler for RayHead
func NewRayHeadReconciler(mgr manager.Manager, enableGangScheduling bool) *RayHeadReconciler {
	baseReconciler := NewJobReconciler(mgr, common.RayHeadControllerName)
	r := &RayHeadReconciler{
		JobReconciler:        baseReconciler,
		EnableGangScheduling: enableGangScheduling,
	}
	r.ConfigInfo = r
	return r
}

// SetupWithManager sets up the controller with the Manager.
func (r *RayHeadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return fmt.Errorf("RayHeadReconciler nil pointer")
	}

	c, err := controller.New(common.RayHeadControllerName, mgr, controller.Options{
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

func (r *RayHeadReconciler) watchRelatedResource(c controller.Controller) error {
	if err := c.Watch(&source.Kind{Type: &mindxdlv1.RayHead{}}, &handler.EnqueueRequestForObject{},
		predicate.Funcs{CreateFunc: r.onOwnerCreateFunc(), DeleteFunc: r.onOwnerDeleteFunc()},
	); err != nil {
		return err
	}
	resourceOptions := []*common.ResourceOption{
		{Kind: &source.Kind{Type: &v1.Service{}}, PredicateFunc: predicate.Funcs{}},
		{Kind: &source.Kind{Type: &v1.Pod{}},
			PredicateFunc: predicate.Funcs{DeleteFunc: r.onPodDeleteFunc()}},
	}

	for _, src := range resourceOptions {
		if err := c.Watch(src.Kind, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &mindxdlv1.RayHead{},
		}, src.PredicateFunc); err != nil {
			return err
		}
	}
	return nil
}

func (r *RayHeadReconciler) GetAPIGroupVersionKind() schema.GroupVersionKind {
	return mindxdlv1.GroupVersion.WithKind(common.RayHeadResourceKind)
}

func (r *RayHeadReconciler) GetAPIGroupVersion() schema.GroupVersion {
	return mindxdlv1.GroupVersion
}

func (r *RayHeadReconciler) GetRunPolicy(obj client.Object) (*commonv1.RunPolicy, error) {
	rayHead, ok := obj.(*mindxdlv1.RayHead)
	if !ok {
		return nil, fmt.Errorf("expect RayHead, but got %v", reflect.TypeOf(obj))
	}
	return &rayHead.Spec.RunPolicy, nil
}

func (r *RayHeadReconciler) GetStatus(obj client.Object) (*commonv1.JobStatus, error) {
	rayHead, ok := obj.(*mindxdlv1.RayHead)
	if !ok {
		return nil, fmt.Errorf("expect RayHead, but got %v", reflect.TypeOf(obj))
	}
	return &rayHead.Status, nil
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the RayHead object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
func (r *RayHeadReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if r == nil {
		return ctrl.Result{}, fmt.Errorf("RayHeadReconciler nil pointer")
	}
	rayHead := &mindxdlv1.RayHead{}
	if err := r.FetchResource(ctx, req.NamespacedName, rayHead); err != nil {
		return ctrl.Result{}, err
	}
	if common.IsObjectEmpty(rayHead) {
		hwlog.RunLog.Warn("RayHeadReconciler object is empty. Skip reconcile.")
		return ctrl.Result{}, nil
	}

	if rayHead.GetDeletionTimestamp() != nil {
		hwlog.RunLog.Infof("reconcile cancelled, job<%s> has been deleted", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	oldStatus := rayHead.Status.DeepCopy()
	defer func() {
		var err error
		if !reflect.DeepEqual(*oldStatus, rayHead.Status) {
			if rayHead.Status.ReplicaStatuses == nil {
				rayHead.Status.ReplicaStatuses = map[commonv1.ReplicaType]*commonv1.ReplicaStatus{}
			}
			err = r.UpdateResourceStatus(rayHead)
		}
		if err != nil {
			r.Recorder.Eventf(rayHead, v1.EventTypeWarning, common.StatusUpdateFailedReason, err.Error())
			hwlog.RunLog.Warnf("failed to update RayCluster status in api server, err: %s", err)
		}
	}()

	err := r.ReconcileRayHead(rayHead)
	if err == nil {
		return ctrl.Result{}, nil
	}
	if k8serr.IsConflict(err) || common.IfNeedRequeue(err) {
		return ctrl.Result{Requeue: true}, nil
	}
	r.Recorder.Eventf(rayHead, v1.EventTypeWarning, util.JobFailedReason, err.Error())
	hwlog.RunLog.Errorf("Reconcile RayCluster<%s> failed, %s", req.NamespacedName, err)
	err = util.UpdateJobConditions(&rayHead.Status, commonv1.JobFailed,
		common.StatusUpdateFailedReason, err.Error())
	if err != nil {
		hwlog.RunLog.Errorf("failed to update job status conditions: %v", err)
	}
	return ctrl.Result{}, nil
}

func (r *RayHeadReconciler) ReconcileRayHead(rayHead *mindxdlv1.RayHead) error {
	r.Mutex.RLock()
	version, ok := r.Versions[rayHead.GetUID()]
	backoffLimit, backoffLimitOk := r.BackoffLimits[rayHead.GetUID()]
	r.Mutex.RUnlock()
	if !ok || (backoffLimitOk && backoffLimit > 0 && version > backoffLimit) {
		msg := fmt.Sprintf("RayHead %s has failed because it has reached the specified backoff limit",
			rayHead.Name)
		hwlog.RunLog.Warn(msg)
		err := util.UpdateJobConditions(&rayHead.Status, commonv1.JobFailed, util.JobFailedReason, msg)
		return err
	}

	if err := r.ReconcileService(rayHead); err != nil {
		hwlog.RunLog.Errorf("Reconcile service of RayHead<%s> failed, %s", rayHead.Name, err)
		return err
	}

	isRunning, err := r.ReconcileHeadPod(rayHead)
	if err != nil {
		return err
	}
	if isRunning {
		if !r.Started {
			// first running
			r.Started = true
			hwlog.RunLog.Infof("RayHead %s finish starting", rayHead.GetName())
		}
		if err := util.UpdateJobConditions(&rayHead.Status, commonv1.JobRunning,
			util.JobRunningReason, ""); err != nil {
			hwlog.RunLog.Errorf("failed to update job status conditions: %v", err)
			return err
		}
	} else if r.Started {
		// maybe some of the pods are restarting
		hwlog.RunLog.Infof("RayHead %s is restarting some pods", rayHead.GetName())
		if err := util.UpdateJobConditions(&rayHead.Status, commonv1.JobRestarting,
			util.JobRestartingReason, "reschedule ray worker pod"); err != nil {
			hwlog.RunLog.Errorf("failed to update job status conditions: %v", err)
			return err
		}
	}
	return nil
}

func (r *RayHeadReconciler) ReconcileHeadPod(rayHead *mindxdlv1.RayHead) (bool, error) {
	podList := &v1.PodList{}
	if err := r.FetchChildrenForResource(rayHead, podList, common.RayHeadIdentification); err != nil {
		hwlog.RunLog.Errorf("list PodList of resource<%s/%s> error: %v", rayHead.GetNamespace(),
			rayHead.GetName(), err)
		return false, err
	}
	pods := podList.Items

	if common.IsSliceEmpty(pods) {
		hwlog.RunLog.Infof("RayHead not exists, create it")
		podTemplate, err := r.createPodSpec(rayHead)
		if err != nil {
			return false, err
		}
		err = r.PodControl.CreatePodsWithControllerRef(rayHead.GetNamespace(), podTemplate,
			rayHead, r.GenOwnerReference(rayHead))
		if err != nil {
			hwlog.RunLog.Errorf("Failed create pod of RayHead %s/%s",
				rayHead.GetNamespace(), rayHead.GetName())
			return false, err
		}
		hwlog.RunLog.Infof("create pod: %s/%s success", rayHead.GetNamespace(), podTemplate.Name)
		return false, nil
	}

	if len(pods) > 1 {
		hwlog.RunLog.Warn("more than 1 ray head pod exists, maybe there is something wrong")
		return false, nil
	}

	headPod := pods[0]
	if headPod.GetDeletionTimestamp() != nil {
		hwlog.RunLog.Warnf("Pod %s/%s is being deleted", headPod.Namespace, headPod.Name)
		return false, &common.ReQueueError{
			Message: fmt.Sprintf("Pod %s/%s is Terminating, try to requeue", headPod.Namespace, headPod.Name),
		}
	}
	if err := r.SetRayEnv(&headPod, rayHead.Name); err != nil {
		return false, err
	}

	return r.HandlePodByStatus(&headPod, rayHead)
}

func (r *RayHeadReconciler) HandlePodByStatus(headPod *v1.Pod, rayHead *mindxdlv1.RayHead) (bool, error) {
	if headPod.Status.Phase == v1.PodRunning {
		return true, nil
	}
	if headPod.Status.Phase == v1.PodFailed || headPod.Status.Phase == v1.PodSucceeded {
		hwlog.RunLog.Infof("Pod %s/%s is failed or unexpectedly completed, will restart it",
			headPod.Namespace, headPod.Name)
		if r.IsReachBackoffLimit(rayHead) {
			return false, fmt.Errorf("RayHead pod<%s> reach backoff limit, it will not restart again",
				rayHead.Name)
		}
		err := r.PodControl.DeletePod(headPod.Namespace, headPod.Name, rayHead)
		if err != nil {
			hwlog.RunLog.Warnf("error delete Pod<%s> of RayHead when restarting: %v", headPod.Name, err)
		}
		return false, err
	}
	hwlog.RunLog.Debugf("expect pod<%s> of RayHead is running, but %s", rayHead.Name, headPod.Status.Phase)
	return false, nil
}

func (r *RayHeadReconciler) createPodSpec(rayHead *mindxdlv1.RayHead) (*v1.PodTemplateSpec, error) {
	podTemplate := rayHead.Spec.Template.DeepCopy()
	podTemplate.Name = flowcommon.GenGeneralName(rayHead.Namespace, common.RayHeadType, common.DefaultPodIndex)

	podTemplate.Labels = r.GenLabels(rayHead, common.RayHeadIdentification)
	podTemplate.Labels[commonv1.ReplicaTypeLabel] = common.RayHeadType
	r.SetCommonPodLabel(podTemplate, rayHead)
	common.SetCommonEnv(podTemplate)
	common.SetVolumes(podTemplate, rayHead.Name)
	podTemplate.Annotations = common.DeepCopyMap(rayHead.Annotations)
	setHeadPodCmdArgs(podTemplate, rayHead.Labels)

	// if gang-scheduling is enabled:
	// 1. if user has specified other scheduler, we report a warning without overriding any fields.
	// 2. if no SchedulerName is set for pods, then we set the SchedulerName to "volcano".
	if r.EnableGangScheduling {
		r.setGangScheduleInfo(rayHead, podTemplate)
	}
	return podTemplate, nil
}

func (r *RayHeadReconciler) setGangScheduleInfo(job *mindxdlv1.RayHead, podTemplate *v1.PodTemplateSpec) {
	jobSchedulerName := job.Spec.SchedulerName
	if len(jobSchedulerName) == 0 || strings.Compare(jobSchedulerName, common.GangSchedulerName) == 0 {
		jobSchedulerName = common.GangSchedulerName
	} else {
		errMsg := "Another job scheduler is specified when gang-scheduling is enabled and it will not be overwritten"
		hwlog.RunLog.Warn(errMsg)
		r.Recorder.Event(job, v1.EventTypeWarning, common.JobSchedulerNameReason, errMsg)
	}
	podSchedulerName := podTemplate.Spec.SchedulerName
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
		common.GetPodGroupName(job.Namespace, job.Name, common.RayHeadType)
	podTemplate.Annotations[common.VolcanoTaskSpecKey] = strings.ToLower(common.RayHeadType)
}
