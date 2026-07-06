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

// Package snapshot for pod reconcile
package snapshot

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/common"
)

type PodSnapshotReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func NewPodSnapshotReadinessReconciler(mgr ctrl.Manager) *PodSnapshotReconciler {
	return &PodSnapshotReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
}

func (r *PodSnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	pod := &corev1.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.setPodSnapshotModeAnnotation(ctx, pod); err != nil {
		hwlog.RunLog.Errorf("Failed to set snapshot mode annotation for pod %s/%s: %v",
			pod.Namespace, pod.Name, err)
		return ctrl.Result{}, nil
	}

	mode := pod.Annotations[common.SnapshotModeAnnotationKey]
	if mode != common.SnapshotSaveMode {
		// bind service for PD instance which doesn't need to save snapshot
		if err := r.setPodActiveLabel(ctx, pod); err != nil {
			hwlog.RunLog.Errorf("Failed to set active label for pod %s/%s: %v",
				pod.Namespace, pod.Name, err)
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *PodSnapshotReconciler) setPodSnapshotModeAnnotation(ctx context.Context, pod *corev1.Pod) error {
	hostSnapshotPath := GetHostSnapshotPath(pod)
	if hostSnapshotPath == "" {
		hwlog.RunLog.Errorf("Pod %s/%s has no hostSnapshotPath configed", pod.Namespace, pod.Name)
		return fmt.Errorf("pod %s/%s has no hostSnapshotPath configed", pod.Namespace, pod.Name)
	}

	snapshotMode := common.SnapshotSaveMode
	if common.IsSnapshotStatusExists(hostSnapshotPath) {
		snapshotMode = common.SnapshotLoadMode
		// change metadata configmap GrusSnapshotRestoredFlag key to true
		r.updateSnapshotConfigMap(ctx, pod)
	} else if instanceIndex := pod.Labels[common.InstanceIndexLabelKey]; "0" != instanceIndex {
		// save mode only apply to the first instance of P/D instanceset
		return nil
	}

	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}

	currentMode, exists := pod.Annotations[common.SnapshotModeAnnotationKey]
	if exists && currentMode == snapshotMode {
		return nil
	}

	updatedPod := pod.DeepCopy()
	updatedPod.Annotations[common.SnapshotModeAnnotationKey] = snapshotMode
	if err := r.Patch(ctx, updatedPod, client.MergeFrom(pod)); err != nil {
		return fmt.Errorf("failed to patch pod annotations: %v", err)
	}
	pod.Annotations[common.SnapshotModeAnnotationKey] = snapshotMode
	return nil
}

func (r *PodSnapshotReconciler) setPodActiveLabel(ctx context.Context, pod *corev1.Pod) error {
	updatedPod := pod.DeepCopy()
	updatedPod.Labels[common.ActiveLabelKey] = common.TrueBool
	if err := r.Patch(ctx, updatedPod, client.MergeFrom(pod)); err != nil {
		return fmt.Errorf("failed to patch pod Labels: %v", err)
	}
	return nil
}

func (r *PodSnapshotReconciler) updateSnapshotConfigMap(ctx context.Context, pod *corev1.Pod) {
	instanceSetName := fmt.Sprintf("%s-%s",
		common.GetInstanceSetNameFromLabels(pod.Labels), pod.Labels[common.InstanceIndexLabelKey])
	cmName := common.SnapshotMetadataPrefix + instanceSetName

	cm := &corev1.ConfigMap{}
	var err error

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Millisecond * 1000 * time.Duration(attempt))
		}

		err = r.Get(ctx, client.ObjectKey{Namespace: pod.Namespace, Name: cmName}, cm)
		if err == nil {
			break
		}

		if !apierrors.IsNotFound(err) {
			hwlog.RunLog.Warnf("Attempt %d: Failed to get configmap %s/%s: %v", attempt+1, pod.Namespace, cmName, err)
		}
	}

	if err != nil {
		if apierrors.IsNotFound(err) {
			hwlog.RunLog.Warnf("ConfigMap %s/%s not found after retries, it may not be created yet or cache not synced",
				pod.Namespace, cmName)
		} else {
			hwlog.RunLog.Warnf("Failed to get configmap %s/%s after retries: %v", pod.Namespace, cmName, err)
		}
		return
	}

	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}

	currentValue, exists := cm.Data[common.GrusSnapshotRestoredFlagKey]
	if exists && currentValue == "true" {
		return
	}

	cm.Data[common.GrusSnapshotRestoredFlagKey] = "true"
	if err := r.Update(ctx, cm); err != nil {
		hwlog.RunLog.Warnf("Failed to update configmap %s/%s: %v", pod.Namespace, cmName, err)
		return
	}

	hwlog.RunLog.Infof("Updated snapshot configmap %s/%s, set GrusSnapshotRestoredFlag to true", pod.Namespace, cmName)
	return
}

func (r *PodSnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(podSnapshotReadinessPredicate()).
		Complete(r)
}

func podSnapshotReadinessPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			pod, ok := e.Object.(*corev1.Pod)
			if !ok {
				return false
			}
			return shouldProcessPod(pod)
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

func shouldProcessPod(pod *corev1.Pod) bool {
	operatorName, ok := pod.Labels[common.OperatorNameKey]
	if !ok {
		return false
	}
	if _, ok := pod.Labels[common.InferServiceNameLabelKey]; !ok {
		return false
	}
	if _, ok := pod.Labels[common.InstanceSetNameLabelKey]; !ok {
		return false
	}
	if _, ok := pod.Labels[common.InstanceIndexLabelKey]; !ok {
		return false
	}
	isContainerSnapshotOn, ok := pod.Labels[common.ContainerSnapshotLabelKey]
	if !ok {
		return false
	}
	return operatorName == common.TrueBool && isContainerSnapshotOn == common.TrueBool
}
