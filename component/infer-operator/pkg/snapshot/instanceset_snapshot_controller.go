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

package snapshot

import (
	"context"
	"encoding/json"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"ascend-common/common-utils/hwlog"
	v1 "infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

// InstanceSetSnapshotReconciler reconciles InstanceSet snapshot operations
type InstanceSetSnapshotReconciler struct {
	client.Client
	// Scheme is the runtime scheme
	Scheme *runtime.Scheme
	// SnapshotChecker checks snapshot completion status
	SnapshotChecker *SnapshotChecker
}

// NewInstanceSetSnapshotReconciler creates a new InstanceSetSnapshotReconciler
func NewInstanceSetSnapshotReconciler(mgr ctrl.Manager) *InstanceSetSnapshotReconciler {
	return &InstanceSetSnapshotReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		SnapshotChecker: NewSnapshotChecker(mgr.GetClient()),
	}
}

// Reconcile reconciles InstanceSet snapshot operations
func (r *InstanceSetSnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instanceSet := &v1.InstanceSet{}
	if err := r.Get(ctx, req.NamespacedName, instanceSet); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	selectLabels := map[string]string{
		common.InferServiceNameLabelKey: instanceSet.Labels[common.InferServiceNameLabelKey],
		common.InstanceSetNameLabelKey:  instanceSet.Labels[common.InstanceSetNameLabelKey],
	}

	stsSpec, err := getStsSpecFromInstanceSpec(instanceSet.Spec.InstanceSpec)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to get sts Spec from InstanceSet %s/%s: %v",
			instanceSet.Namespace, instanceSet.Name, err)
		return ctrl.Result{}, nil
	}

	hostSnapshotPath := common.GetHostSnapshotPathFromPodTemplate(&stsSpec.Template, instanceSet)
	if hostSnapshotPath == "" {
		hwlog.RunLog.Warnf("InstanceSet %s/%s has no host-snapshot volume mounts, skip snapshot tracking",
			instanceSet.Namespace, instanceSet.Name)
		return ctrl.Result{}, nil
	}
	if common.IsSnapshotStatusExists(hostSnapshotPath) &&
		common.IsSnapshotValid(hostSnapshotPath) {
		hwlog.RunLog.Infof("Host snapshot path %s for InstanceSet %s/%s is valid, will load snapshot",
			hostSnapshotPath, instanceSet.Namespace, instanceSet.Name)
		return ctrl.Result{}, nil
	}

	if !r.SnapshotChecker.IsRunning() {
		r.SnapshotChecker.Start(ctx)
	}
	if err := r.SnapshotChecker.TrackInstanceSet(instanceSet, selectLabels, *stsSpec.Replicas); err != nil {
		hwlog.RunLog.Warnf("Failed to track InstanceSet %s/%s for snapshot: %v",
			instanceSet.Namespace, instanceSet.Name, err)
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the manager
func (r *InstanceSetSnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.InstanceSet{}).
		WithEventFilter(snapshotPredicate()).
		Complete(r)
}

func snapshotPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			instanceSet, ok := e.Object.(*v1.InstanceSet)
			if !ok {
				return false
			}
			return shouldProcessInstanceset(instanceSet)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

func shouldProcessInstanceset(instanceSet *v1.InstanceSet) bool {
	// only track instanceset with container snapshot enabled
	if instanceSet.Spec.Replicas == nil || *instanceSet.Spec.Replicas <= 0 {
		return false
	}

	return common.IsContainerSnapshotOn(instanceSet)
}

// Start starts the snapshot checker
func (r *InstanceSetSnapshotReconciler) Start(ctx context.Context) {
	r.SnapshotChecker.Start(ctx)
}

// Stop stops the snapshot checker
func (r *InstanceSetSnapshotReconciler) Stop() {
	r.SnapshotChecker.Stop()
}

func getStsSpecFromInstanceSpec(raw runtime.RawExtension) (*appsv1.StatefulSetSpec, error) {
	if len(raw.Raw) == 0 {
		return nil, fmt.Errorf("instance spec is empty")
	}

	var spec appsv1.StatefulSetSpec
	if err := json.Unmarshal(raw.Raw, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal instance spec: %w", err)
	}

	return &spec, nil
}
