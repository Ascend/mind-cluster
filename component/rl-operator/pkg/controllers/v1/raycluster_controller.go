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
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	mindxdlv1 "rl-operator/pkg/api/v1"
	"rl-operator/pkg/common"
)

// RayClusterReconciler reconciles a RayCluster object
type RayClusterReconciler struct {
	*JobReconciler
	EnableGangScheduling bool
}

type RayClusterJobInfo struct {
	rayCluster    *mindxdlv1.RayCluster
	podGroupList  []*v1beta1.PodGroup
	rayHeadList   []*mindxdlv1.RayHead
	rayWorkerList []*mindxdlv1.RayWorker
}

// NewRayClusterReconciler new reconciler for RayCluster
func NewRayClusterReconciler(mgr manager.Manager, enableGangScheduling bool) *RayClusterReconciler {
	baseReconciler := NewJobReconciler(mgr, common.RayClusterControllerName)
	r := &RayClusterReconciler{
		JobReconciler:        baseReconciler,
		EnableGangScheduling: enableGangScheduling,
	}
	r.ConfigInfo = r
	return r
}

func (r *RayClusterReconciler) GetAPIGroupVersionKind() schema.GroupVersionKind {
	return mindxdlv1.GroupVersion.WithKind(common.RayClusterResourceKind)
}

func (r *RayClusterReconciler) GetAPIGroupVersion() schema.GroupVersion {
	return mindxdlv1.GroupVersion
}

func (r *RayClusterReconciler) GetRunPolicy(obj client.Object) (*commonv1.RunPolicy, error) {
	rayCluster, ok := obj.(*mindxdlv1.RayCluster)
	if !ok {
		return nil, fmt.Errorf("expect RayCluster, but got %v", reflect.TypeOf(obj))
	}
	return &rayCluster.Spec.RunPolicy, nil
}

func (r *RayClusterReconciler) GetStatus(obj client.Object) (*commonv1.JobStatus, error) {
	rayCluster, ok := obj.(*mindxdlv1.RayCluster)
	if !ok {
		return nil, fmt.Errorf("expect RayCluster, but got %v", reflect.TypeOf(obj))
	}
	return &rayCluster.Status, nil
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the RayCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
func (r *RayClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r == nil {
		return ctrl.Result{}, fmt.Errorf("RayClusterReconciler nil pointer")
	}
	rayCluster := &mindxdlv1.RayCluster{}
	if err := r.FetchResource(ctx, req.NamespacedName, rayCluster); err != nil {
		return ctrl.Result{}, err
	}
	if common.IsObjectEmpty(rayCluster) {
		hwlog.RunLog.Warn("RayClusterReconciler object is empty. Skip reconcile.")
		return ctrl.Result{}, nil
	}

	if rayCluster.GetDeletionTimestamp() != nil {
		hwlog.RunLog.Infof("reconcile cancelled，job<%s> has been deleted", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	oldStatus := rayCluster.Status.DeepCopy()
	defer func() {
		var err error
		if !reflect.DeepEqual(*oldStatus, rayCluster.Status) {
			hwlog.RunLog.Debugf("RayCluster Update status, from %v to %v", *oldStatus, rayCluster.Status)
			if rayCluster.Status.ReplicaStatuses == nil {
				rayCluster.Status.ReplicaStatuses = map[commonv1.ReplicaType]*commonv1.ReplicaStatus{}
			}
			err = r.UpdateResourceStatus(rayCluster)
		}
		if err != nil {
			r.Recorder.Eventf(rayCluster, v1.EventTypeWarning, common.StatusUpdateFailedReason, err.Error())
			hwlog.RunLog.Warnf("failed to update RayCluster status in api server, err: %s", err)
		}
	}()

	err := r.ReconcileRayCluster(rayCluster)
	if err == nil {
		return ctrl.Result{}, nil
	}
	if k8serr.IsConflict(err) || common.IfNeedRequeue(err) {
		return ctrl.Result{Requeue: true}, nil
	}
	r.Recorder.Eventf(rayCluster, v1.EventTypeWarning, util.JobFailedReason, err.Error())
	hwlog.RunLog.Errorf("Reconcile RayCluster<%s> failed, %s", req.NamespacedName, err)
	err = util.UpdateJobConditions(&rayCluster.Status, commonv1.JobFailed,
		common.StatusUpdateFailedReason, err.Error())
	if err != nil {
		hwlog.RunLog.Warnf("failed to update job status conditions: %v", err)
	}
	return ctrl.Result{}, nil
}

func (r *RayClusterReconciler) ReconcileRayCluster(rayCluster *mindxdlv1.RayCluster) error {
	if err := r.ValidateJob(*rayCluster); err != nil {
		return fmt.Errorf("RayCluster %s Validation Error: %s", rayCluster.Name, err.Error())
	}
	// Set default priorities to RayCluster
	r.Scheme.Default(rayCluster)

	if err := r.ReconcileConfigMap(rayCluster); err != nil {
		return err
	}

	ji, err := r.getRayClusterJobInfo(rayCluster)
	if err != nil {
		return err
	}

	if r.EnableGangScheduling {
		isReady, err := r.ReconcilePodGroups(rayCluster)
		if err != nil {
			return err
		}
		if !isReady {
			return nil
		}
	}

	if err := r.ReconcileRayRoles(ji); err != nil {
		return err
	}

	return nil
}

func (r *RayClusterReconciler) ReconcileConfigMap(rayCluster *mindxdlv1.RayCluster) error {
	cmName := common.GenConfigInfoConfigMapName(rayCluster.Name)
	cm, err := r.GetConfigMapWithRetry(rayCluster.Namespace, cmName)
	if err != nil {
		hwlog.RunLog.Errorf("Get ConfigMap %s Error: %s", cmName, err.Error())
		return err
	}
	if cm != nil {
		hwlog.RunLog.Infof("ConfigMap %s already exists", cmName)
		return nil
	}
	cm = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: rayCluster.Namespace,
		},
		Data: map[string]string{},
	}
	if err := r.CreateConfigMap(cm); err != nil {
		hwlog.RunLog.Errorf("Failed to create ConfigMap %s/%s: %v", cm.Namespace, cm.Name, err)
		return err
	}
	hwlog.RunLog.Infof("ConfigMap %s/%s created", cm.Namespace, cm.Name)
	return nil
}

func (r *RayClusterReconciler) getRayClusterJobInfo(rayCluster *mindxdlv1.RayCluster) (*RayClusterJobInfo, error) {
	podGroupList := &v1beta1.PodGroupList{}
	if err := r.FetchChildrenForResource(rayCluster, podGroupList, common.RayClusterIdentification); err != nil {
		hwlog.RunLog.Errorf("list PodGroupList of resource<%s/%s> error: %v", rayCluster.GetNamespace(),
			rayCluster.GetName(), err)
		return nil, err
	}

	rayHeadList := &mindxdlv1.RayHeadList{}
	if err := r.FetchChildrenForResource(rayCluster, rayHeadList, common.RayClusterIdentification); err != nil {
		hwlog.RunLog.Errorf("list RayHeadList of resource<%s/%s> error: %v", rayCluster.GetNamespace(),
			rayCluster.GetName(), err)
		return nil, err
	}

	rayWorkerList := &mindxdlv1.RayWorkerList{}
	if err := r.FetchChildrenForResource(rayCluster, rayWorkerList, common.RayClusterIdentification); err != nil {
		hwlog.RunLog.Errorf("list RayWorkerList of resource<%s/%s> error: %v", rayCluster.GetNamespace(),
			rayCluster.GetName(), err)
		return nil, err
	}

	return &RayClusterJobInfo{
		rayCluster:    rayCluster,
		podGroupList:  common.ConvertToPointerSlice(podGroupList.Items),
		rayHeadList:   common.ConvertToPointerSlice(rayHeadList.Items),
		rayWorkerList: common.ConvertToPointerSlice(rayWorkerList.Items),
	}, nil
}

func (r *RayClusterReconciler) ReconcilePodGroups(rayCluster *mindxdlv1.RayCluster) (bool, error) {
	replicas := rayCluster.Spec.ReplicaSpecs
	result := true
	for rType := range replicas {
		isReady, err := r.ReconcilePodGroup(rayCluster, rType)
		result = result && isReady
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func (r *RayClusterReconciler) ReconcileRayRoles(ji *RayClusterJobInfo) error {
	rayCluster := ji.rayCluster
	rayHeadList := ji.rayHeadList
	if rayHeadReady, err := r.ReconcileRayHead(rayHeadList, rayCluster); !rayHeadReady {
		hwlog.RunLog.Debug("RayHead not exists or is not running")
		if err != nil {
			return err
		}
		return r.cleanUpRayWorker(ji)
	}

	rayWorkerList := ji.rayWorkerList
	isWorkerReady, err := r.ReconcileRayWorker(rayWorkerList, rayCluster)
	if err != nil {
		return err
	}
	if !isWorkerReady {
		return nil
	}
	if err = util.UpdateJobConditions(&rayCluster.Status, commonv1.JobRunning,
		util.JobRunningReason, ""); err != nil {
		hwlog.RunLog.Errorf("failed to update job status conditions: %v", err)
	}
	return err
}

func (r *RayClusterReconciler) ReconcilePodGroup(
	rayCluster *mindxdlv1.RayCluster,
	rType commonv1.ReplicaType) (bool, error) {
	jobNamespace := rayCluster.GetNamespace()
	jobName := rayCluster.GetName()
	pgName := common.GetPodGroupName(jobNamespace, jobName, string(rType))
	var pg *v1beta1.PodGroup
	pg, err := r.VolcanoControl.GetPodGroup(jobNamespace, pgName)
	if err != nil {
		// PodGroup not exists
		hwlog.RunLog.Debugf("get RayCluster<%s/%s> pg failed, try to create pg",
			jobNamespace, jobName)
		pgSpec := r.newPodGroupSpec(rayCluster, rType)
		if pgSpec == nil {
			return false, fmt.Errorf("RayCluster<%s/%s> new PodGroupSpec failed", jobNamespace, jobName)
		}
		// create podGroup for gang scheduling by volcano
		podGroupLabel := r.getPodGroupLabel(rayCluster)
		createPodGroup := &v1beta1.PodGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:        pgName,
				Namespace:   rayCluster.GetNamespace(),
				Annotations: rayCluster.GetAnnotations(),
				Labels:      podGroupLabel,
				OwnerReferences: []metav1.OwnerReference{
					*r.GenOwnerReference(rayCluster),
				},
			},
			Spec: *pgSpec,
		}
		pg, err = r.VolcanoControl.CreatePodGroup(createPodGroup)
	}
	if err != nil {
		hwlog.RunLog.Warnf("create PodGroup %v failed: %v", pgName, err)
		return false, err
	}

	// Delay pods creation until podgroup status is inqueue
	if pg == nil || pg.Status.Phase == "" || pg.Status.Phase == v1beta1.PodGroupPending {
		hwlog.RunLog.Warnf("PodGroup %v unschedulable", pgName)
		return false, nil
	}
	return true, nil
}

func (r *RayClusterReconciler) getPodGroupLabel(rayCluster *mindxdlv1.RayCluster) map[string]string {
	podGroupLabel := common.DeepCopyMap(rayCluster.GetLabels())
	extraLabels := r.GenLabels(rayCluster, common.RayClusterIdentification)
	for key, value := range extraLabels {
		podGroupLabel[key] = value
	}
	return podGroupLabel
}

func (r *RayClusterReconciler) ReconcileRayHead(
	rayHeads []*mindxdlv1.RayHead,
	rayCluster *mindxdlv1.RayCluster) (bool, error) {
	if common.IsSliceEmpty(rayHeads) {
		hwlog.RunLog.Infof("RayHead not exists, create it")
		createRayHead := r.newRayHead(rayCluster)
		if err := r.CreateChildForResource(rayCluster, createRayHead, common.RayClusterIdentification); err != nil {
			hwlog.RunLog.Errorf("failed to create resource RayHead<%s>, err: %s",
				createRayHead.GetName(), err)
			return false, err
		}
		return false, nil
	}

	if len(rayHeads) > 1 {
		hwlog.RunLog.Warn("more than 1 RayHead exists, maybe there is something wrong")
	}

	rayHead := rayHeads[0]
	if !common.IsRunning(rayHead.Status) {
		hwlog.RunLog.Debugf("expect RayHead<%s> is running, but %s",
			rayHead.Name, common.GetStatusString(rayHead.Status))
		common.SyncStatus(&rayCluster.Status, &rayHead.Status)
		return false, nil
	}
	return true, nil
}

func (r *RayClusterReconciler) newRayHead(rayCluster *mindxdlv1.RayCluster) *mindxdlv1.RayHead {
	headSpec := rayCluster.Spec.ReplicaSpecs[common.RayHeadType]
	headTemplate := headSpec.Template.DeepCopy()
	r.SetRestartPolicy(rayCluster, headTemplate, headSpec.RestartPolicy)
	return &mindxdlv1.RayHead{
		TypeMeta: metav1.TypeMeta{
			Kind:       common.RayHeadResourceKind,
			APIVersion: mindxdlv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        rayCluster.Name,
			Namespace:   rayCluster.Namespace,
			Labels:      common.DeepCopyMap(rayCluster.Labels),
			Annotations: common.DeepCopyMap(rayCluster.Annotations),
		},
		Spec: mindxdlv1.RayHeadSpec{
			Template:      *headTemplate,
			SchedulerName: rayCluster.Spec.SchedulerName,
		},
	}
}

func (r *RayClusterReconciler) cleanUpRayWorker(ji *RayClusterJobInfo) error {
	rayWorkerList := ji.rayWorkerList
	for _, rayWorker := range rayWorkerList {
		err := r.DeleteResource(rayWorker)
		if err != nil {
			hwlog.RunLog.Errorf("cleanUpRayWorker: delete resource<%s/%s> error: %v",
				rayWorker.GetNamespace(), rayWorker.GetName(), err)
			return err
		}
	}
	return nil
}

func (r *RayClusterReconciler) ReconcileRayWorker(
	rayWorkers []*mindxdlv1.RayWorker,
	rayCluster *mindxdlv1.RayCluster) (bool, error) {
	if common.IsSliceEmpty(rayWorkers) {
		hwlog.RunLog.Infof("RayWorker not exists, create it")
		createRayWorker := newRayWorker(rayCluster)
		if err := r.CreateChildForResource(rayCluster, createRayWorker, common.RayClusterIdentification); err != nil {
			hwlog.RunLog.Errorf("failed to create resource %s<%s>, err: %s",
				reflect.TypeOf(createRayWorker).Name(), createRayWorker.GetName(), err)
			return false, err
		}
		return false, nil
	}

	if len(rayWorkers) > 1 {
		hwlog.RunLog.Warn("more than 1 RayWorker exists, maybe there is something wrong")
	}

	rayWorker := rayWorkers[0]
	if !common.IsRunning(rayWorker.Status) {
		hwlog.RunLog.Infof("expect RayHead<%s> is running, but %s",
			rayWorker.Name, common.GetStatusString(rayWorker.Status))
		common.SyncStatus(&rayCluster.Status, &rayWorker.Status)
		return false, nil
	}
	return true, nil
}

func newRayWorker(rayCluster *mindxdlv1.RayCluster) *mindxdlv1.RayWorker {
	rayClusterCopy := rayCluster.Spec.DeepCopy()
	delete(rayClusterCopy.ReplicaSpecs, common.RayHeadType)
	labels := common.DeepCopyMap(rayCluster.Labels)
	labels[common.ServiceNameLabelKey] = common.GetServiceName(rayCluster.Namespace, rayCluster.Name)
	return &mindxdlv1.RayWorker{
		TypeMeta: metav1.TypeMeta{
			Kind:       common.RayWorkerResourceKind,
			APIVersion: mindxdlv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        rayCluster.Name,
			Namespace:   rayCluster.Namespace,
			Labels:      labels,
			Annotations: common.DeepCopyMap(rayCluster.Annotations),
		},
		Spec: mindxdlv1.RayWorkerSpec{
			ReplicaSpecs:  rayClusterCopy.ReplicaSpecs,
			SchedulerName: rayCluster.Spec.SchedulerName,
		},
	}
}

func (r *RayClusterReconciler) newPodGroupSpec(
	rayCluster *mindxdlv1.RayCluster,
	rType commonv1.ReplicaType) *v1beta1.PodGroupSpec {
	currentSpec, ok := rayCluster.Spec.ReplicaSpecs[rType]
	if !ok {
		hwlog.RunLog.Warnf("replica %s not found when new PodGroup", rType)
		return nil
	}
	minMember := *currentSpec.Replicas
	queue := ""
	priorityClass := ""
	var minResources *corev1.ResourceList = nil

	runPolicy := rayCluster.Spec.RunPolicy
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
		currentSpecMap := make(map[commonv1.ReplicaType]*commonv1.ReplicaSpec)
		currentSpecMap[rType] = currentSpec
		minResources = flowcommon.CalcPGMinResources(minMember, currentSpecMap, r.PriorityClassLister.Get)
	}

	minTaskMember := make(map[string]int32)
	minTaskMember[strings.ToLower(string(rType))] = common.DefaultMinMember
	if currentSpec.Replicas != nil {
		minTaskMember[strings.ToLower(string(rType))] = *currentSpec.Replicas
	}

	return &v1beta1.PodGroupSpec{
		MinMember:         minMember,
		MinTaskMember:     minTaskMember,
		Queue:             queue,
		PriorityClassName: priorityClass,
		MinResources:      minResources,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *RayClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return fmt.Errorf("RayClusterReconciler nil pointer")
	}

	c, err := controller.New(common.RayClusterControllerName, mgr, controller.Options{
		Reconciler: r,
		RateLimiter: workqueue.NewMaxOfRateLimiter(
			workqueue.NewItemExponentialFailureRateLimiter(common.WorkQueueBaseDelay, common.WorkQueueMaxDelay),
			&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(common.WorkQueueQps), common.WorkQueueBurst)},
		),
	})
	if err != nil {
		return err
	}
	return r.watchRelatedResource(c, mgr)
}

func (r *RayClusterReconciler) watchRelatedResource(c controller.Controller, mgr ctrl.Manager) error {
	if err := c.Watch(&source.Kind{Type: &mindxdlv1.RayCluster{}}, &handler.EnqueueRequestForObject{},
		predicate.Funcs{CreateFunc: r.onOwnerCreateFunc(), DeleteFunc: r.onOwnerDeleteFunc()},
	); err != nil {
		return err
	}
	resourceOptions := []*common.ResourceOption{
		{Kind: &source.Kind{Type: &mindxdlv1.RayHead{}}, PredicateFunc: predicate.Funcs{}},
		{Kind: &source.Kind{Type: &mindxdlv1.RayWorker{}}},
	}

	if r.EnableGangScheduling {
		_, mapErr := mgr.GetRESTMapper().RESTMapping(schema.GroupKind{Group: v1beta1.SchemeGroupVersion.Group,
			Kind: "PodGroup"},
			v1beta1.SchemeGroupVersion.Version)
		if mapErr != nil {
			hwlog.RunLog.Errorf("enableGangScheduling is true, but PodGroup is not in cluster, " +
				"maybe volcano is not started on your cluster")
			return mapErr
		}
		resourceOptions = append(resourceOptions,
			&common.ResourceOption{Kind: &source.Kind{Type: &v1beta1.PodGroup{}}})
	}

	for _, src := range resourceOptions {
		if err := c.Watch(src.Kind, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &mindxdlv1.RayCluster{},
		}, src.PredicateFunc); err != nil {
			return err
		}
	}
	return nil
}

// onOwnerDeleteFunc clean up configmap created before
func (r *RayClusterReconciler) onOwnerDeleteFunc() func(deleteEvent event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		cmName := common.GenConfigInfoConfigMapName(e.Object.GetName())
		cm, err := r.GetConfigMapWithRetry(e.Object.GetNamespace(), cmName)
		if err != nil {
			hwlog.RunLog.Error("Failed to get ConfigMap, ConfigMap will not be deleted if exists: %v", err)
			return r.JobReconciler.onOwnerDeleteFunc()(e)
		}
		if cm == nil {
			return r.JobReconciler.onOwnerDeleteFunc()(e)
		}
		if err := r.DeleteConfigMap(cm); err != nil {
			hwlog.RunLog.Errorf("Failed to delete ConfigMap %s/%s: %v", cm.Namespace, cm.Name, err)
			return r.JobReconciler.onOwnerDeleteFunc()(e)
		}
		hwlog.RunLog.Infof("ConfigMap %s/%s deleted", cm.Namespace, cm.Name)
		return r.JobReconciler.onOwnerDeleteFunc()(e)
	}
}
