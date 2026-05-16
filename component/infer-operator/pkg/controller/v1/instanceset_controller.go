/*
Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package v1 contains API Schema definitions for the mindcluster v1 API group
package v1

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	apiv1 "infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
	util "infer-operator/pkg/common/client-go"
	"infer-operator/pkg/controller/schedule"
	"infer-operator/pkg/controller/workload"
)

// InstanceSetReconciler reconciles a InstanceSet object
type InstanceSetReconciler struct {
	client.Client
	workload.PodGroupManager
	Scheme             *runtime.Scheme
	WorkLoadReconciler *workload.WorkLoadReconciler
	Recorder           record.EventRecorder
	SupportPodGroup    bool
}

type WorkloadRegister func(mgr ctrl.Manager, reconciler *workload.WorkLoadReconciler)

// Reconcile reconciles the InstanceSet.
func (r *InstanceSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// 1. fetch InstanceSet
	instanceSet := &apiv1.InstanceSet{}
	if err := r.Get(ctx, req.NamespacedName, instanceSet); err != nil {
		if apierrors.IsNotFound(err) {
			hwlog.RunLog.Infof("InstanceSet does not exist %s/%s", req.Namespace, req.Name)
		} else {
			hwlog.RunLog.Errorf("unable to fetch InstanceSet %s/%s, error: %v",
				req.Namespace, req.Name, err)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if !(instanceSet.DeletionTimestamp == nil || instanceSet.DeletionTimestamp.IsZero()) {
		hwlog.RunLog.Infof("instanceSet %s is being deleted", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	// 2. validate InstanceSet
	if err := r.validate(instanceSet); err != nil {
		hwlog.RunLog.Errorf("InstanceSet %s/%s is invalid, error: %v",
			req.Namespace, req.Name, err)
		return ctrl.Result{}, nil
	}

	// 3. reconcile workloads
	err := r.reconcileWorkLoads(ctx, instanceSet)
	if common.IsRequeueError(err) || err == nil {
		// 4. update status
		if err := r.updateStatus(ctx, instanceSet); err != nil {
			hwlog.RunLog.Errorf("unable to update status %s/%s, error: %v", req.Namespace, req.Name, err)
			return ctrl.Result{}, err
		}
		if common.IsRequeueError(err) {
			hwlog.RunLog.Warnf("reconcile workloads of InstanceSet %s/%s error: %v, will requeue this request",
				req.Namespace, req.Name, err)
			return ctrl.Result{RequeueAfter: common.DefaultReEnqueueInterval}, nil
		}
		return ctrl.Result{}, nil
	}

	if apierrors.IsConflict(err) {
		hwlog.RunLog.Warnf("conflict error when reconcile workloads of InstanceSet %s/%s, "+
			"will requeue immediately, error: %v", req.Namespace, req.Name, err)
		return ctrl.Result{Requeue: true}, nil
	}
	hwlog.RunLog.Errorf("reconcile workloads of InstanceSet %s/%s error: %v",
		req.Namespace, req.Name, err)
	return ctrl.Result{}, err
}

func (r *InstanceSetReconciler) reconcileWorkLoads(ctx context.Context, instanceSet *apiv1.InstanceSet) error {
	indexer := common.InstanceIndexer{
		Namespace:      instanceSet.Namespace,
		ServiceName:    instanceSet.Labels[common.InferServiceNameLabelKey],
		InstanceSetKey: instanceSet.Labels[common.InstanceSetNameLabelKey],
	}

	if err := r.deleteExtraWorkloads(ctx, instanceSet, indexer); err != nil {
		return err
	}

	if err := r.reconcileServices(ctx, instanceSet, indexer); err != nil {
		return err
	}

	if shouldSchedule, err := schedule.ShouldSchedule(ctx, r.Client, instanceSet); err != nil {
		return err
	} else if !shouldSchedule {
		return nil
	}

	return r.reconcileWorkloadInstances(ctx, instanceSet, indexer)
}

func (r *InstanceSetReconciler) checkOrCreateService(
	ctx context.Context,
	instanceSet *apiv1.InstanceSet,
	serviceSpec apiv1.ServiceSpec,
	indexer common.InstanceIndexer) error {
	service := &corev1.Service{}
	customServiceName := fmt.Sprintf("%s-%s-%s", serviceSpec.Name, indexer.ServiceName, indexer.InstanceSetKey)
	err := r.Client.Get(ctx, types.NamespacedName{Name: customServiceName, Namespace: instanceSet.Namespace}, service)
	if err != nil && !apierrors.IsNotFound(err) {
		hwlog.RunLog.Errorf("Failed to get service %s: %v", customServiceName, err)
		return err
	}

	if serviceSpec.Spec.Type == corev1.ServiceTypeNodePort {
		// For NodePort type, add instanceSet index offset to each instanceset's nodePort
		if err := handleConflictNodePort(&serviceSpec, indexer); err != nil {
			return err
		}
	}
	if apierrors.IsNotFound(err) {
		labels := common.AddLabelsFromIndexer(instanceSet.Labels, indexer)
		newService := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        customServiceName,
				Namespace:   instanceSet.Namespace,
				Annotations: instanceSet.Annotations,
				Labels:      labels,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(instanceSet, instanceSet.GroupVersionKind()),
				},
			},
			Spec: serviceSpec.Spec,
		}
		if err := r.Client.Create(ctx, newService); err != nil {
			r.Recorder.Eventf(instanceSet, corev1.EventTypeWarning, common.ServiceCreateReason,
				fmt.Sprintf("Failed to create service %s: %v", serviceSpec.Name, err))
			hwlog.RunLog.Errorf("Failed to create service %s: %v", serviceSpec.Name, err)
			return err
		}
	}
	return nil
}

func (r *InstanceSetReconciler) validate(instanceSet *apiv1.InstanceSet) error {
	// 1. validate gang-schedule
	enableGangScheduling := instanceSet.Labels[common.GangScheduleLabelKey]
	if enableGangScheduling == common.TrueBool && !r.SupportPodGroup {
		errorMsg := "label gang-schedule is true, but PodGroup not supported now, please check if " +
			"volcano installed in k8s"
		r.Recorder.Eventf(instanceSet, corev1.EventTypeWarning, common.ValidateErrorReason, errorMsg)
		return errors.New(errorMsg)
	}
	// 2. validate replicas
	replicas := instanceSet.Spec.Replicas
	if replicas == nil {
		errorMsg := "jobSpec is not valid: replicas cannot be nil"
		r.Recorder.Eventf(instanceSet, corev1.EventTypeWarning, common.ValidateErrorReason, errorMsg)
		return errors.New(errorMsg)
	}
	if *replicas < 0 {
		errorMsg := fmt.Sprintf("jobSpec is not valid: replicas can not be negative num, but but got %d",
			*replicas)
		r.Recorder.Eventf(instanceSet, corev1.EventTypeWarning, common.ValidateErrorReason, errorMsg)
		return errors.New(errorMsg)
	}
	// 3. validate labels
	if _, ok := instanceSet.Labels[common.InferServiceNameLabelKey]; !ok {
		errorMsg := fmt.Sprintf("label <%s> is missing", common.InferServiceNameLabelKey)
		r.Recorder.Eventf(instanceSet, corev1.EventTypeWarning, common.ValidateErrorReason, errorMsg)
		return errors.New(errorMsg)
	}
	if _, ok := instanceSet.Labels[common.InstanceSetNameLabelKey]; !ok {
		errorMsg := fmt.Sprintf("label <%s> is missing", common.InstanceSetNameLabelKey)
		r.Recorder.Eventf(instanceSet, corev1.EventTypeWarning, common.ValidateErrorReason, errorMsg)
		return errors.New(errorMsg)
	}
	// 4. validate nodePort
	if err := validateServices(instanceSet); err != nil {
		r.Recorder.Eventf(instanceSet, corev1.EventTypeWarning, common.ValidateErrorReason, err.Error())
		return err
	}
	// 5. validate workload
	err := r.WorkLoadReconciler.Validate(instanceSet)
	if err != nil {
		r.Recorder.Eventf(instanceSet, corev1.EventTypeWarning, common.ValidateErrorReason, err.Error())
		return err
	}
	return nil
}

func (r *InstanceSetReconciler) updateStatus(ctx context.Context, instanceSet *apiv1.InstanceSet) error {
	if instanceSet == nil {
		return nil
	}
	indexer := common.InstanceIndexer{
		ServiceName:    instanceSet.Labels[common.InferServiceNameLabelKey],
		InstanceSetKey: instanceSet.Labels[common.InstanceSetNameLabelKey],
	}
	newStatus, err := r.getNewStatus(ctx, instanceSet, indexer)
	if err != nil {
		return err
	}
	if reflect.DeepEqual(instanceSet.Status, newStatus) {
		return nil
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestInstanceSet := &apiv1.InstanceSet{}
		if err := r.Get(ctx, types.NamespacedName{Name: instanceSet.Name, Namespace: instanceSet.Namespace},
			latestInstanceSet); err != nil {
			hwlog.RunLog.Errorf("get latestInstanceSet %s/%s error: %v",
				instanceSet.Namespace, instanceSet.Name, err)
			return err
		}
		latestInstanceSet.Status = newStatus
		if err := r.Status().Update(ctx, latestInstanceSet); err != nil {
			hwlog.RunLog.Errorf("update latestInstanceSet %s/%s status error: %v",
				instanceSet.Namespace, instanceSet.Name, err)
			return err
		}
		hwlog.RunLog.Infof("update latestInstanceSet %s/%s status successfully",
			instanceSet.Namespace, instanceSet.Name)
		return nil
	})
}

func (r *InstanceSetReconciler) getNewStatus(
	ctx context.Context,
	instanceSet *apiv1.InstanceSet,
	indexer common.InstanceIndexer) (newStatus apiv1.InstanceSetStatus, err error) {
	newStatus = *instanceSet.Status.DeepCopy()
	newStatus.Replicas = *instanceSet.Spec.Replicas
	newStatus.ObservedGeneration = instanceSet.Generation
	readyReplicas, err := r.WorkLoadReconciler.InstanceReady(ctx, instanceSet, indexer)
	if err != nil {
		hwlog.RunLog.Errorf("get ready replicas of instanceSet %s/%s error: %v",
			instanceSet.Namespace, instanceSet.Name, err)
		return newStatus, err
	}
	newStatus.ReadyReplicas = int32(readyReplicas)

	var condition metav1.Condition
	if newStatus.ReadyReplicas >= newStatus.Replicas {
		condition = metav1.Condition{
			Type:               string(common.InstanceSetReady),
			Status:             metav1.ConditionTrue,
			Reason:             common.AllWorkloadReadyReason,
			Message:            common.AllWorkloadReadyMessage,
			LastTransitionTime: metav1.Now(),
		}
	} else {
		condition = metav1.Condition{
			Type:   string(common.InstanceSetReady),
			Status: metav1.ConditionFalse,
			Reason: common.WorkloadNotReadyReason,
			Message: fmt.Sprintf("%d of %d InferService replicas are ready",
				newStatus.ReadyReplicas, newStatus.Replicas),
			LastTransitionTime: metav1.Now(),
		}
	}
	meta.SetStatusCondition(&newStatus.Conditions, condition)
	return newStatus, nil
}

func validateServices(instanceSet *apiv1.InstanceSet) error {
	for _, serviceSpec := range instanceSet.Spec.Services {
		if serviceSpec.Spec.Type != corev1.ServiceTypeNodePort {
			// skip non-NodePort service
			continue
		}
		for _, port := range serviceSpec.Spec.Ports {
			if port.NodePort == 0 {
				errorMsg := fmt.Sprintf("nodePort of service<%s> should config as non-zero", serviceSpec.Name)
				return errors.New(errorMsg)
			}
		}
	}
	return nil
}

// For NodePort type, add instanceSet index offset to each instanceset's nodePort
func handleConflictNodePort(serviceSpec *apiv1.ServiceSpec, indexer common.InstanceIndexer) error {
	splitResult := strings.Split(indexer.ServiceName, "-")
	if len(splitResult) < common.InferServiceNameSplitNum {
		return fmt.Errorf("service name %s format error, should be <service-name>-<index>",
			indexer.ServiceName)
	}
	offsetInt32, err := strconv.ParseInt(splitResult[len(splitResult)-1], common.BaseDec, common.BitSize)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to parse service index %v: %v", offsetInt32, err)
		return err
	}
	offset := int32(offsetInt32)
	for i := range serviceSpec.Spec.Ports {
		if serviceSpec.Spec.Ports[i].NodePort > 0 {
			// Add instanceSet index offset to each instanceset's nodePort, avoid port conflict
			serviceSpec.Spec.Ports[i].NodePort += offset
		} else {
			errMsg := fmt.Sprintf("nodePort of service<%s> should config as non-zero", serviceSpec.Name)
			hwlog.RunLog.Error(errMsg)
			return errors.New(errMsg)
		}
	}
	return nil
}

// NewInstanceSetReconciler creates a new InstanceSetReconciler.
func NewInstanceSetReconciler(
	mgr manager.Manager,
	workloadRegister WorkloadRegister) *InstanceSetReconciler {
	workLoadReconciler := workload.NewWorkLoadReconciler(mgr.GetClient())
	workloadRegister(mgr, workLoadReconciler)
	recorder := mgr.GetEventRecorderFor(common.InstanceSetControllerName)
	return &InstanceSetReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		PodGroupManager:    workload.NewVolcanoPodGroupManager(mgr.GetClient()),
		WorkLoadReconciler: workLoadReconciler,
		Recorder:           recorder,
		SupportPodGroup:    false,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *InstanceSetReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	controller := ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.InstanceSet{}).
		Owns(&appsv1.Deployment{}, builder.WithPredicates(WorkLoadPredicate())).
		Owns(&appsv1.StatefulSet{}, builder.WithPredicates(WorkLoadPredicate())).
		Owns(&corev1.Service{}, builder.WithPredicates(WorkLoadPredicate())).
		Named(common.InstanceSetControllerName)

	// if PodGroup exists, support PodGroup
	if err := util.CRDExists(ctx, mgr.GetAPIReader(), common.VolcanoPodGroupCrdName); err == nil {
		hwlog.RunLog.Info("Volcano PodGroup CRD exists, support PodGroup")
		controller.Owns(&v1beta1.PodGroup{}, builder.WithPredicates(PodGroupPredicate()))
		r.SupportPodGroup = true
	} else {
		hwlog.RunLog.Infof("Volcano PodGroup CRD not exists, gang schedule is disabled, err: %v",
			err.Error())
	}

	return controller.Complete(r)
}

func WorkLoadPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			// skip create event
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
	}
}

func PodGroupPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			// skip create event
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			// skip update event
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
	}
}

func (r *InstanceSetReconciler) deleteExtraWorkloads(ctx context.Context, instanceSet *apiv1.InstanceSet, indexer common.InstanceIndexer) error {
	return r.WorkLoadReconciler.DeleteExtraInstances(ctx, instanceSet, indexer)
}

func (r *InstanceSetReconciler) reconcileServices(ctx context.Context, instanceSet *apiv1.InstanceSet, indexer common.InstanceIndexer) error {
	for _, serviceSpec := range instanceSet.Spec.Services {
		if err := r.checkOrCreateService(ctx, instanceSet, serviceSpec, indexer); err != nil {
			return err
		}
	}
	return nil
}

func (r *InstanceSetReconciler) reconcileWorkloadInstances(ctx context.Context, instanceSet *apiv1.InstanceSet, indexer common.InstanceIndexer) error {
	replicaNum := int(*instanceSet.Spec.Replicas)
	for instanceIndex := 0; instanceIndex < replicaNum; instanceIndex++ {
		indexer.InstanceIndex = strconv.Itoa(instanceIndex)
		if err := r.WorkLoadReconciler.Reconcile(ctx, instanceSet, indexer); err != nil {
			hwlog.RunLog.Errorf("reconcile WorkLoads of InstanceSet %s/%s, error: %v",
				instanceSet.Namespace, instanceSet.Name, err)
			return err
		}
	}
	return nil
}
