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

package v1

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"ascend-common/common-utils/hwlog"
	apiv1 "infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

// InferServiceSetReconciler reconciles a InferServiceSet object
type InferServiceSetReconciler struct {
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
	counter  atomic.Uint64
}

// NewInferServiceSetReconciler returns a new InferServiceSetReconciler
func NewInferServiceSetReconciler(mgr ctrl.Manager) *InferServiceSetReconciler {
	return &InferServiceSetReconciler{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor(common.InferServiceSetControllerName),
	}
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *InferServiceSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	startTime := time.Now()
	requestID := r.counter.Add(1)
	hwlog.RunLog.Debugf("start InferServiceSet reconcile %s (%v) requestID %d",
		req.NamespacedName, startTime, requestID)
	defer func() {
		hwlog.RunLog.Debugf("finish InferServiceSet reconcile %s (%v) requestID %d",
			req.NamespacedName, time.Since(startTime), requestID)
	}()

	iss, err := r.getInferServiceSet(ctx, req)
	if err != nil {
		return ctrl.Result{RequeueAfter: common.DefaultReEnqueueInterval}, err
	}
	if iss == nil {
		return ctrl.Result{}, nil
	}

	if err := r.validate(ctx, iss, req); err != nil {
		return ctrl.Result{RequeueAfter: common.NonRetriableRequeInterval}, err
	}

	isList, selector, err := r.listInferServices(ctx, iss)
	if err != nil {
		return ctrl.Result{RequeueAfter: common.DefaultReEnqueueInterval}, err
	}

	existedIsList := r.buildInferServiceMap(isList)

	isToCreate, isToUpdate, isToDelete := r.calculateInferServiceOperations(iss, existedIsList)

	if err := r.manageInferServices(ctx, iss, isToDelete, isToUpdate, isToCreate); err != nil {
		return ctrl.Result{RequeueAfter: common.DefaultReEnqueueInterval}, err
	}

	if err := r.updateInferServiceSetStatus(ctx, iss, selector); err != nil {
		return ctrl.Result{RequeueAfter: common.DefaultReEnqueueInterval}, err
	}

	return ctrl.Result{}, nil
}

// getInferServiceSet retrieves the InferServiceSet object from the cluster
func (r *InferServiceSetReconciler) getInferServiceSet(ctx context.Context, req ctrl.Request) (*apiv1.InferServiceSet, error) {
	iss := &apiv1.InferServiceSet{}
	if err := r.client.Get(ctx, req.NamespacedName, iss); err != nil {
		hwlog.RunLog.Errorf("unable to get InferServiceSet %s: %v", req.NamespacedName, err)
		if errors.IsNotFound(err) {
			hwlog.RunLog.Infof("InferServiceSet %s not found, skip reconcile", req.NamespacedName)
			return nil, nil
		}
		return nil, err
	}

	if !(iss.DeletionTimestamp == nil || iss.DeletionTimestamp.IsZero()) {
		hwlog.RunLog.Infof("InferServiceSet %s is being deleted", req.NamespacedName)
		return nil, nil
	}

	return iss, nil
}

// validateAndCheckStatus validates the InferServiceSet and checks its status
func (r *InferServiceSetReconciler) validate(ctx context.Context, iss *apiv1.InferServiceSet, req ctrl.Request) error {
	if iss.Spec.Replicas == nil {
		iss.Spec.Replicas = new(int32)
		*iss.Spec.Replicas = 1
	}
	if *iss.Spec.Replicas > common.MaxInferServiceReplicas || *iss.Spec.Replicas < 1 {
		hwlog.RunLog.Errorf("InferServiceSet %s replicas %d exceeds max %d or less than 1",
			req.NamespacedName, *iss.Spec.Replicas, common.MaxInferServiceReplicas)
		return fmt.Errorf("inferServiceSet %s replicas %d exceeds max %d or less than 1",
			req.NamespacedName, *iss.Spec.Replicas, common.MaxInferServiceReplicas)
	}

	return nil
}

// listInferServices lists all child InferServices for the InferServiceSet
func (r *InferServiceSetReconciler) listInferServices(ctx context.Context,
	iss *apiv1.InferServiceSet) (*apiv1.InferServiceList, labels.Selector, error) {
	isList := &apiv1.InferServiceList{}
	selector := labels.SelectorFromSet(labels.Set{
		common.InferServiceSetNameLabelKey: iss.Name,
	})
	if err := r.client.List(ctx, isList,
		client.InNamespace(iss.Namespace),
		client.MatchingLabelsSelector{Selector: selector}); err != nil {
		hwlog.RunLog.Errorf("Failed to list child InferServices: %v", err)
		return nil, nil, err
	}

	return isList, selector, nil
}

// buildInferServiceMap builds a map of InferServices by index
func (r *InferServiceSetReconciler) buildInferServiceMap(isList *apiv1.InferServiceList) map[int]*apiv1.InferService {
	existedIsList := make(map[int]*apiv1.InferService)

	for i := range isList.Items {
		is := &isList.Items[i]
		index, ok := is.Labels[common.InferServiceIndexLabelKey]
		if !ok {
			continue
		}
		indexInt, err := strconv.Atoi(index)
		if err != nil {
			continue
		}
		existedIsList[indexInt] = is
	}

	return existedIsList
}

// calculateInferServiceOperations calculates which InferServices need to be created, updated, or deleted
func (r *InferServiceSetReconciler) calculateInferServiceOperations(iss *apiv1.InferServiceSet,
	existedIsList map[int]*apiv1.InferService) ([]*apiv1.InferService, []*apiv1.InferService, []*apiv1.InferService) {
	desiredReplicas := int(*iss.Spec.Replicas)
	var isToCreate []*apiv1.InferService
	var isToUpdate []*apiv1.InferService
	var isToDelete []*apiv1.InferService

	for i := 0; i < desiredReplicas; i++ {
		if is, ok := existedIsList[i]; ok {
			if r.inferServiceUpdated(iss, is) {
				isToUpdate = append(isToUpdate, is)
			}
			delete(existedIsList, i)
		} else {
			is := newInferService(iss, i)
			isToCreate = append(isToCreate, is)
		}
	}

	for index, is := range existedIsList {
		if index >= desiredReplicas {
			hwlog.RunLog.Warnf("InferService %s/%s index %d is out of desired replicas %d, will be deleted",
				is.Namespace, is.Name, index, desiredReplicas)
		} else {
			hwlog.RunLog.Warnf("InferService %s/%s missing index label or index is not a number, will be deleted",
				is.Namespace, is.Name)
		}
		isToDelete = append(isToDelete, is)
	}

	return isToCreate, isToUpdate, isToDelete
}

// manageInferServices handles the creation, update, and deletion of InferServices
func (r *InferServiceSetReconciler) manageInferServices(ctx context.Context,
	iss *apiv1.InferServiceSet, isToDelete []*apiv1.InferService, isToUpdate []*apiv1.InferService, isToCreate []*apiv1.InferService) error {
	if err := r.scaleDown(ctx, iss, isToDelete); err != nil {
		hwlog.RunLog.Errorf("Failed to scale down InferServiceSet %s/%s: %v", iss.Namespace, iss.Name, err)
		return err
	}

	if err := r.updateExistInferServices(ctx, iss, isToUpdate); err != nil {
		hwlog.RunLog.Errorf("Failed to update exist InferServices for InferServiceSet %s/%s: %v", iss.Namespace, iss.Name, err)
		return err
	}

	if err := r.scaleUp(ctx, iss, isToCreate); err != nil {
		hwlog.RunLog.Errorf("Failed to scale up InferServiceSet %s/%s: %v", iss.Namespace, iss.Name, err)
		return err
	}

	return nil
}

// updateInferServiceSetStatus updates the status of the InferServiceSet
func (r *InferServiceSetReconciler) updateInferServiceSetStatus(ctx context.Context, iss *apiv1.InferServiceSet, selector labels.Selector) error {
	isList := &apiv1.InferServiceList{}
	if err := r.client.List(ctx, isList,
		client.InNamespace(iss.Namespace),
		client.MatchingLabelsSelector{Selector: selector}); err != nil {
		hwlog.RunLog.Errorf("Failed to list child InferServices for InferServiceSet %s/%s: %v", iss.Namespace, iss.Name, err)
		return err
	}

	if err := r.updateStatus(ctx, iss, isList); err != nil {
		hwlog.RunLog.Errorf("Failed to update InferServiceSet %s/%s status: %v", iss.Namespace, iss.Name, err)
		return err
	}

	return nil
}

func (r *InferServiceSetReconciler) updateStatus(ctx context.Context, iss *apiv1.InferServiceSet, isList *apiv1.InferServiceList) error {
	if iss == nil {
		return nil
	}

	newStatus := r.calculateStatus(iss, isList)

	if reflect.DeepEqual(iss.Status, newStatus) {
		return nil
	}

	return r.updateStatusWithRetry(ctx, iss, newStatus)
}

// calculateStatus calculates the new status for the InferServiceSet
func (r *InferServiceSetReconciler) calculateStatus(iss *apiv1.InferServiceSet, isList *apiv1.InferServiceList) apiv1.InferServiceSetStatus {
	newStatus := *iss.Status.DeepCopy()
	newStatus.Replicas = *iss.Spec.Replicas
	newStatus.ObservedGeneration = iss.Generation

	newStatus.ReadyReplicas = r.calculateReadyReplicas(isList)

	condition := r.createReadyCondition(newStatus.ReadyReplicas, newStatus.Replicas)
	meta.SetStatusCondition(&newStatus.Conditions, condition)

	return newStatus
}

// calculateReadyReplicas calculates the number of ready InferServices
func (r *InferServiceSetReconciler) calculateReadyReplicas(isList *apiv1.InferServiceList) int32 {
	readyReplicas := int32(0)
	for _, is := range isList.Items {
		if meta.IsStatusConditionTrue(is.Status.Conditions, string(common.InferServiceSetReady)) {
			readyReplicas++
		}
	}
	return readyReplicas
}

// createReadyCondition creates the Ready condition for the InferServiceSet
func (r *InferServiceSetReconciler) createReadyCondition(readyReplicas, totalReplicas int32) metav1.Condition {
	if readyReplicas >= totalReplicas {
		return metav1.Condition{
			Type:               string(common.InferServiceSetReady),
			Status:             metav1.ConditionTrue,
			Reason:             "AllInferServicesReady",
			Message:            "All InferService replicas are ready",
			LastTransitionTime: metav1.Now(),
		}
	}

	return metav1.Condition{
		Type:   string(common.InferServiceSetReady),
		Status: metav1.ConditionFalse,
		Reason: "ReplicasNotReady",
		Message: fmt.Sprintf("%d of %d InferService replicas are ready",
			readyReplicas, totalReplicas),
		LastTransitionTime: metav1.Now(),
	}
}

// updateStatusWithRetry updates the status of the InferServiceSet with retry on conflict
func (r *InferServiceSetReconciler) updateStatusWithRetry(ctx context.Context, iss *apiv1.InferServiceSet, newStatus apiv1.InferServiceSetStatus) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestIss := &apiv1.InferServiceSet{}
		if err := r.client.Get(ctx, types.NamespacedName{
			Name:      iss.Name,
			Namespace: iss.Namespace,
		}, latestIss); err != nil {
			return err
		}
		latestIss.Status = newStatus
		return r.client.Status().Update(ctx, latestIss)
	})
}

func (r *InferServiceSetReconciler) updateExistInferServices(ctx context.Context, iss *apiv1.InferServiceSet, isList []*apiv1.InferService) error {
	if iss == nil {
		return nil
	}
	for _, is := range isList {
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			latestInferService := &apiv1.InferService{}
			if err := r.client.Get(ctx, types.NamespacedName{
				Name:      is.Name,
				Namespace: is.Namespace,
			}, latestInferService); err != nil {
				return err
			}

			latestInferService.Spec = iss.Spec.InferServiceTemplate
			return r.client.Update(ctx, latestInferService)
		})
		if err != nil {
			hwlog.RunLog.Errorf("Failed to update InferService %s/%s: %v", is.Namespace, is.Name, err)
			return err
		}
	}
	return nil
}

func (r *InferServiceSetReconciler) inferServiceUpdated(iss *apiv1.InferServiceSet, is *apiv1.InferService) bool {
	if iss == nil || is == nil {
		return false
	}
	if !reflect.DeepEqual(is.Spec, iss.Spec.InferServiceTemplate) {
		return true
	}
	return false
}

func (r *InferServiceSetReconciler) scaleDown(ctx context.Context, iss *apiv1.InferServiceSet, isList []*apiv1.InferService) error {
	if iss == nil {
		return nil
	}
	for _, is := range isList {
		if err := r.client.Delete(ctx, is); err != nil {
			if !errors.IsNotFound(err) {
				hwlog.RunLog.Errorf("Failed to delete InferService %s/%s: %v", is.Namespace, is.Name, err)
				return err
			}
			hwlog.RunLog.Infof("InferService %s/%s not found, skip delete", is.Namespace, is.Name)
		}
	}
	return nil
}

func (r *InferServiceSetReconciler) scaleUp(ctx context.Context, iss *apiv1.InferServiceSet, isList []*apiv1.InferService) error {
	if iss == nil {
		return nil
	}
	for _, is := range isList {
		if err := controllerutil.SetControllerReference(iss, is, r.scheme); err != nil {
			hwlog.RunLog.Errorf("Failed to set controller reference for InferService %s/%s: %v", is.Namespace, is.Name, err)
			return err
		}
		err := r.client.Create(ctx, is)
		if err == nil {
			hwlog.RunLog.Infof("InferService %s/%s created", is.Namespace, is.Name)
		} else if errors.IsAlreadyExists(err) {
			hwlog.RunLog.Infof("InferService %s/%s already exists, skip create", is.Namespace, is.Name)
			continue
		} else {
			hwlog.RunLog.Errorf("Failed to create InferService %s/%s: %v", is.Namespace, is.Name, err)
			return err
		}
	}
	return nil
}

func newInferService(iss *apiv1.InferServiceSet, index int) *apiv1.InferService {
	return &apiv1.InferService{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      iss.Name + "-" + strconv.Itoa(index),
			Namespace: iss.Namespace,
			Labels: map[string]string{
				common.InferServiceSetNameLabelKey: iss.Name,
				common.InferServiceIndexLabelKey:   strconv.Itoa(index),
			},
		},
		Spec: iss.Spec.InferServiceTemplate,
	}
}

// filterInferServiceEvents filters InferService events to only process those belonging to the InferServiceSet
func filterInferServiceEvents() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Only process InferService events
			oldIs, okOld := e.ObjectOld.(*apiv1.InferService)
			newIs, okNew := e.ObjectNew.(*apiv1.InferService)
			if !okOld || !okNew {
				return false
			}

			// Filter out updates where spec and status are identical
			if reflect.DeepEqual(oldIs.Spec, newIs.Spec) && reflect.DeepEqual(oldIs.Status, newIs.Status) {
				return false
			}

			// Only process InferService events that have InferServiceSetNameLabelKey label
			return newIs.Labels != nil && newIs.Labels[common.InferServiceSetNameLabelKey] != ""
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if is, ok := e.Object.(*apiv1.InferService); ok {
				// Only process InferService events that have InferServiceSetNameLabelKey label
				return is.Labels != nil && is.Labels[common.InferServiceSetNameLabelKey] != ""
			}
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *InferServiceSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.InferServiceSet{}).
		Owns(&apiv1.InferService{}, builder.WithPredicates(filterInferServiceEvents())).
		Named(common.InferServiceSetControllerName).
		Complete(r)
}
