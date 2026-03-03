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
	"sync/atomic"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
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

// InferServiceReconciler reconciles a InferService object
type InferServiceReconciler struct {
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
	counter  atomic.Uint64
}

// NewInferServiceReconciler returns a new InferServiceReconciler
func NewInferServiceReconciler(mgr ctrl.Manager) *InferServiceReconciler {
	return &InferServiceReconciler{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor(common.InferServiceControllerName),
	}
}

// validateRoles validates the roles in an InferService
// It checks:
// 1. If the number of roles exceeds roletypecount
// 2. If there are duplicate role names
// 3. If each role name complies with IsDNS1123Label standards
func (r *InferServiceReconciler) validateRoles(is *apiv1.InferService) error {
	if is == nil {
		return fmt.Errorf("inferService cannot be nil")
	}

	// Check if the number of roles exceeds roletypecount
	if len(is.Spec.Roles) > common.MaxRoleTypeCount {
		return fmt.Errorf("number of roles %d exceeds maximum allowed %d", len(is.Spec.Roles), common.MaxRoleTypeCount)
	}

	// Check for duplicate role names
	roleNameMap := make(map[string]bool)
	for i, role := range is.Spec.Roles {
		if role.Name == "" {
			return fmt.Errorf("role at index %d has empty name", i)
		}

		if roleNameMap[role.Name] {
			return fmt.Errorf("duplicate role name %s", role.Name)
		}
		roleNameMap[role.Name] = true

		// Check if role name complies with IsDNS1123Label standards
		if errs := validation.IsDNS1123Label(role.Name); len(errs) > 0 {
			return fmt.Errorf("role name %s is invalid: %v", role.Name, errs)
		}
	}

	return nil
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *InferServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	startTime := time.Now()
	requestID := r.counter.Add(1)
	hwlog.RunLog.Debugf("start InferService reconcile %s (%v) requestID %d",
		req.NamespacedName, startTime, requestID)
	defer func() {
		hwlog.RunLog.Debugf("finish InferService reconcile %s (%v) requestID %d",
			req.NamespacedName, time.Since(startTime), requestID)
	}()

	is, err := r.getInferService(ctx, req)
	if err != nil {
		return ctrl.Result{RequeueAfter: common.DefaultReEnqueueInterval}, err
	}
	if is == nil {
		return ctrl.Result{}, nil
	}

	if err := r.validate(ctx, is, req); err != nil {
		return ctrl.Result{RequeueAfter: common.NonRetriableRequeInterval}, err
	}

	instanceSetList, selector, err := r.listInstanceSets(ctx, is)
	if err != nil {
		return ctrl.Result{RequeueAfter: common.DefaultReEnqueueInterval}, err
	}

	existedInstanceSetMap := r.buildInstanceSetMap(instanceSetList)

	instanceSetsToCreate, instanceSetsToUpdate, instanceSetsToDelete := r.calculateInstanceSetOperations(is, existedInstanceSetMap)

	if err := r.manageInstanceSets(ctx, is, instanceSetsToDelete, instanceSetsToUpdate, instanceSetsToCreate); err != nil {
		return ctrl.Result{RequeueAfter: common.DefaultReEnqueueInterval}, err
	}

	if err := r.updateInferServiceStatus(ctx, is, selector); err != nil {
		return ctrl.Result{RequeueAfter: common.DefaultReEnqueueInterval}, err
	}

	return ctrl.Result{}, nil
}

// getInferService retrieves the InferService object from the cluster
func (r *InferServiceReconciler) getInferService(ctx context.Context, req ctrl.Request) (*apiv1.InferService, error) {
	is := &apiv1.InferService{}
	if err := r.client.Get(ctx, req.NamespacedName, is); err != nil {
		hwlog.RunLog.Errorf("unable to get InferService %s: %v", req.NamespacedName, err)
		if errors.IsNotFound(err) {
			hwlog.RunLog.Infof("InferService %s not found, skip reconcile", req.NamespacedName)
			return nil, nil
		}
		return nil, err
	}

	if !(is.DeletionTimestamp == nil || is.DeletionTimestamp.IsZero()) {
		hwlog.RunLog.Infof("InferService %s is being deleted", req.NamespacedName)
		return nil, nil
	}

	return is, nil
}

// validateAndCheckStatus validates the InferService and checks its status
func (r *InferServiceReconciler) validate(ctx context.Context, is *apiv1.InferService, req ctrl.Request) error {
	// Validate roles in InferService
	if err := r.validateRoles(is); err != nil {
		hwlog.RunLog.Errorf("validation failed for InferService %s: %v", req.NamespacedName, err)
		return err
	}

	return nil
}

// listInstanceSets lists all child InstanceSets for the InferService
func (r *InferServiceReconciler) listInstanceSets(ctx context.Context, is *apiv1.InferService) (*apiv1.InstanceSetList, labels.Selector, error) {
	instanceSetList := &apiv1.InstanceSetList{}
	selector := labels.SelectorFromSet(labels.Set{
		common.InferServiceNameLabelKey: is.Name,
	})
	if err := r.client.List(ctx, instanceSetList,
		client.InNamespace(is.Namespace),
		client.MatchingLabelsSelector{Selector: selector}); err != nil {
		hwlog.RunLog.Errorf("Failed to list child InstanceSets for InferService %s/%s: %v", is.Namespace, is.Name, err)
		return nil, nil, err
	}

	return instanceSetList, selector, nil
}

// buildInstanceSetMap builds a map of InstanceSets by role name
func (r *InferServiceReconciler) buildInstanceSetMap(instanceSetList *apiv1.InstanceSetList) map[string]*apiv1.InstanceSet {
	existedInstanceSetMap := make(map[string]*apiv1.InstanceSet)
	for i := range instanceSetList.Items {
		instanceSet := &instanceSetList.Items[i]
		roleName, ok := instanceSet.Labels[common.InstanceSetNameLabelKey]
		if !ok {
			hwlog.RunLog.Errorf("InstanceSet %s/%s missing role name label", instanceSet.Namespace, instanceSet.Name)
			continue
		}
		existedInstanceSetMap[roleName] = instanceSet
	}

	return existedInstanceSetMap
}

// calculateInstanceSetOperations calculates which InstanceSets need to be created, updated, or deleted
func (r *InferServiceReconciler) calculateInstanceSetOperations(is *apiv1.InferService,
	existedInstanceSetMap map[string]*apiv1.InstanceSet) ([]*apiv1.InstanceSet, []*apiv1.InstanceSet, []*apiv1.InstanceSet) {
	var instanceSetsToCreate []*apiv1.InstanceSet
	var instanceSetsToUpdate []*apiv1.InstanceSet

	for _, role := range is.Spec.Roles {
		if instanceSet, ok := existedInstanceSetMap[role.Name]; ok {
			if r.instanceSetUpdated(instanceSet, role) {
				// Update the InstanceSet's Spec to the new role's Spec before updating
				instanceSet.Spec = role
				instanceSetsToUpdate = append(instanceSetsToUpdate, instanceSet)
			}
			delete(existedInstanceSetMap, role.Name)
		} else {
			instanceSet := r.newInstanceSet(is, role)
			instanceSetsToCreate = append(instanceSetsToCreate, instanceSet)
		}
	}

	var instanceSetsToDelete []*apiv1.InstanceSet
	for _, instanceSet := range existedInstanceSetMap {
		instanceSetsToDelete = append(instanceSetsToDelete, instanceSet)
	}

	return instanceSetsToCreate, instanceSetsToUpdate, instanceSetsToDelete
}

// manageInstanceSets handles the creation, update, and deletion of InstanceSets
func (r *InferServiceReconciler) manageInstanceSets(ctx context.Context, is *apiv1.InferService,
	instanceSetsToDelete []*apiv1.InstanceSet, instanceSetsToUpdate []*apiv1.InstanceSet, instanceSetsToCreate []*apiv1.InstanceSet) error {
	if err := r.deleteInstanceSets(ctx, is, instanceSetsToDelete); err != nil {
		hwlog.RunLog.Errorf("Failed to delete InstanceSets for InferService %s/%s: %v", is.Namespace, is.Name, err)
		return err
	}

	if err := r.updateExistInstanceSets(ctx, is, instanceSetsToUpdate); err != nil {
		hwlog.RunLog.Errorf("Failed to update exist InstanceSets for InferService %s/%s: %v", is.Namespace, is.Name, err)
		return err
	}

	if err := r.createInstanceSets(ctx, is, instanceSetsToCreate); err != nil {
		hwlog.RunLog.Errorf("Failed to scale up InferService %s/%s: %v", is.Namespace, is.Name, err)
		return err
	}

	return nil
}

// updateInferServiceStatus updates the status of the InferService
func (r *InferServiceReconciler) updateInferServiceStatus(ctx context.Context, is *apiv1.InferService, selector labels.Selector) error {
	instanceSetList := &apiv1.InstanceSetList{}
	if err := r.client.List(ctx, instanceSetList,
		client.InNamespace(is.Namespace),
		client.MatchingLabelsSelector{Selector: selector}); err != nil {
		hwlog.RunLog.Errorf("Failed to list child InstanceSets for InferService %s/%s: %v", is.Namespace, is.Name, err)
		return err
	}

	if err := r.updateStatus(ctx, is, instanceSetList); err != nil {
		hwlog.RunLog.Errorf("Failed to update InferService %s/%s status: %v", is.Namespace, is.Name, err)
		return err
	}

	return nil
}

func (r *InferServiceReconciler) updateStatus(ctx context.Context, is *apiv1.InferService, instanceSetList *apiv1.InstanceSetList) error {
	if is == nil {
		return nil
	}

	newStatus := r.calculateStatus(is, instanceSetList)

	if reflect.DeepEqual(is.Status, newStatus) {
		return nil
	}

	return r.updateStatusWithRetry(ctx, is, newStatus)
}

// calculateStatus calculates the new status for the InferService
func (r *InferServiceReconciler) calculateStatus(is *apiv1.InferService, instanceSetList *apiv1.InstanceSetList) apiv1.InferServiceStatus {
	newStatus := *is.Status.DeepCopy()

	newStatus.Replicas = int32(len(is.Spec.Roles))
	newStatus.ReadyReplicas = r.calculateReadyReplicas(instanceSetList)

	condition := r.createReadyCondition(newStatus.ReadyReplicas, newStatus.Replicas)
	meta.SetStatusCondition(&newStatus.Conditions, condition)

	newStatus.ObservedGeneration = is.Generation

	return newStatus
}

// calculateReadyReplicas calculates the number of ready InstanceSets
func (r *InferServiceReconciler) calculateReadyReplicas(instanceSetList *apiv1.InstanceSetList) int32 {
	readyReplicas := int32(0)
	for _, instanceSet := range instanceSetList.Items {
		// Check if InstanceSet is ready based on condition
		if meta.IsStatusConditionTrue(instanceSet.Status.Conditions, string(common.InferServiceReady)) {
			readyReplicas += 1
		}
	}
	return readyReplicas
}

// createReadyCondition creates the Ready condition for the InferService
func (r *InferServiceReconciler) createReadyCondition(readyReplicas, totalReplicas int32) metav1.Condition {
	if readyReplicas >= totalReplicas && totalReplicas > 0 {
		return metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			Reason:             "AllInstanceSetsReady",
			Message:            "All InstanceSet replicas are ready",
			LastTransitionTime: metav1.Now(),
		}
	}

	return metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		Reason:             "ReplicasNotReady",
		Message:            fmt.Sprintf("%d of %d InstanceSet replicas are ready", readyReplicas, totalReplicas),
		LastTransitionTime: metav1.Now(),
	}
}

// updateStatusWithRetry updates the status of the InferService with retry on conflict
func (r *InferServiceReconciler) updateStatusWithRetry(ctx context.Context, is *apiv1.InferService, newStatus apiv1.InferServiceStatus) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestInferService := &apiv1.InferService{}
		if err := r.client.Get(ctx, types.NamespacedName{
			Name:      is.Name,
			Namespace: is.Namespace,
		}, latestInferService); err != nil {
			return err
		}
		latestInferService.Status = newStatus
		return r.client.Status().Update(ctx, latestInferService)
	})
}

func (r *InferServiceReconciler) updateExistInstanceSets(ctx context.Context, is *apiv1.InferService, instanceSets []*apiv1.InstanceSet) error {
	if is == nil {
		return nil
	}
	for _, instanceSet := range instanceSets {
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			latestInstanceSet := &apiv1.InstanceSet{}
			if err := r.client.Get(ctx, types.NamespacedName{
				Name:      instanceSet.Name,
				Namespace: instanceSet.Namespace,
			}, latestInstanceSet); err != nil {
				return err
			}

			latestInstanceSet.Spec = instanceSet.Spec
			return r.client.Update(ctx, latestInstanceSet)
		})
		if err != nil {
			hwlog.RunLog.Errorf("Failed to update InstanceSet %s/%s for InferService %s/%s: %v",
				instanceSet.Namespace, instanceSet.Name, is.Namespace, is.Name, err)
			return err
		}
		hwlog.RunLog.Infof("updated InstanceSet %s/%s for InferService %s/%s",
			instanceSet.Namespace, instanceSet.Name, is.Namespace, is.Name)
	}
	return nil
}

func (r *InferServiceReconciler) instanceSetUpdated(instanceSet *apiv1.InstanceSet, role apiv1.InstanceSetSpec) bool {
	if instanceSet == nil {
		return false
	}
	if !reflect.DeepEqual(instanceSet.Spec, role) {
		return true
	}
	return false
}

func (r *InferServiceReconciler) deleteInstanceSets(ctx context.Context, is *apiv1.InferService, instanceSets []*apiv1.InstanceSet) error {
	if is == nil {
		return nil
	}
	for _, instanceSet := range instanceSets {
		if err := r.client.Delete(ctx, instanceSet); err != nil {
			if !errors.IsNotFound(err) {
				hwlog.RunLog.Errorf("Failed to delete InstanceSet %s/%s for InferService %s/%s: %v",
					instanceSet.Namespace, instanceSet.Name, is.Namespace, is.Name, err)
				return err
			}
			hwlog.RunLog.Infof("InstanceSet %s/%s not found for InferService %s/%s, skip delete",
				instanceSet.Namespace, instanceSet.Name, is.Namespace, is.Name)
		}
		hwlog.RunLog.Infof("deleted InstanceSet %s/%s for InferService %s/%s",
			instanceSet.Namespace, instanceSet.Name, is.Namespace, is.Name)
	}
	return nil
}

func (r *InferServiceReconciler) createInstanceSets(ctx context.Context, is *apiv1.InferService, instanceSets []*apiv1.InstanceSet) error {
	if is == nil {
		return nil
	}
	for _, instanceSet := range instanceSets {
		if err := controllerutil.SetControllerReference(is, instanceSet, r.scheme); err != nil {
			hwlog.RunLog.Errorf("Failed to set controller reference for InstanceSet %s/%s for InferService %s/%s: %v",
				instanceSet.Namespace, instanceSet.Name, is.Namespace, is.Name, err)
			return err
		}
		err := r.client.Create(ctx, instanceSet)
		if err == nil {
			hwlog.RunLog.Infof("InstanceSet %s/%s created for InferService %s/%s",
				instanceSet.Namespace, instanceSet.Name, is.Namespace, is.Name)
		} else if errors.IsAlreadyExists(err) {
			hwlog.RunLog.Infof("InstanceSet %s/%s already exists for InferService %s/%s, skip create",
				instanceSet.Namespace, instanceSet.Name, is.Namespace, is.Name)
			continue
		} else {
			hwlog.RunLog.Errorf("Failed to create InstanceSet %s/%s for InferService %s/%s: %v",
				instanceSet.Namespace, instanceSet.Name, is.Namespace, is.Name, err)
			return err
		}
	}
	return nil
}

func (r *InferServiceReconciler) newInstanceSet(is *apiv1.InferService, role apiv1.InstanceSetSpec) *apiv1.InstanceSet {
	labels := make(map[string]string)
	if role.WorkloadObjectMeta.Labels != nil {
		for k, v := range role.WorkloadObjectMeta.Labels {
			labels[k] = v
		}
	}
	labels[common.InferServiceNameLabelKey] = is.Name
	labels[common.InstanceSetNameLabelKey] = role.Name

	annotations := make(map[string]string)
	if role.WorkloadObjectMeta.Annotations != nil {
		for k, v := range role.WorkloadObjectMeta.Annotations {
			annotations[k] = v
		}
	}

	return &apiv1.InstanceSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        is.Name + "-" + role.Name,
			Namespace:   is.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: role,
	}
}

// filterInstanceSetEvents filters InstanceSet events to only process those belonging to an InferService
func filterInstanceSetEvents() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			// Directly return false to skip InstanceSet create events
			// These events are already handled during InferService reconciliation
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Only process InstanceSet events
			oldIs, okOld := e.ObjectOld.(*apiv1.InstanceSet)
			newIs, okNew := e.ObjectNew.(*apiv1.InstanceSet)
			if !okOld || !okNew {
				return false
			}

			// Filter out updates where spec and status are identical
			if reflect.DeepEqual(oldIs.Spec, newIs.Spec) && reflect.DeepEqual(oldIs.Status, newIs.Status) {
				return false
			}

			// Only process InstanceSet events that have InferServiceNameLabelKey label
			return newIs.Labels != nil && newIs.Labels[common.InferServiceNameLabelKey] != ""
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if iset, ok := e.Object.(*apiv1.InstanceSet); ok {
				// Only process InstanceSet events that have InferServiceNameLabelKey label
				return iset.Labels != nil && iset.Labels[common.InferServiceNameLabelKey] != ""
			}
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *InferServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.InferService{}).
		Owns(&apiv1.InstanceSet{}, builder.WithPredicates(filterInstanceSetEvents())).
		Named(common.InferServiceControllerName).
		Complete(r)
}
