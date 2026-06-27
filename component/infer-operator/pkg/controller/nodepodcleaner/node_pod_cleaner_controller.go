/*
Copyright(C) 2026-2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package nodepodcleaner provides a controller that force-deletes infer-operator
// managed pods on nodes that become NotReady (e.g. power off / reboot). This
// avoids the long Terminating window caused by pod-eviction-timeout and lets
// StatefulSet/Deployment controllers recreate pods on healthy nodes quickly.
package nodepodcleaner

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/common"
)

// NodePodCleanerReconciler watches Node status. When a node turns NotReady, it
// force-deletes (grace-period=0) all infer-operator managed pods scheduled on
// that node so that the owning workload controllers can recreate them on
// healthy nodes without waiting for kubelet recovery or pod-eviction-timeout.
type NodePodCleanerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// NewNodePodCleanerReconciler returns a new NodePodCleanerReconciler.
func NewNodePodCleanerReconciler(mgr manager.Manager) *NodePodCleanerReconciler {
	return &NodePodCleanerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
}

// Reconcile handles a node that has just turned NotReady. It force-deletes all
// infer-operator managed pods running on that node.
func (r *NodePodCleanerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	node := &corev1.Node{}
	if err := r.Get(ctx, req.NamespacedName, node); err != nil {
		if apierrors.IsNotFound(err) {
			hwlog.RunLog.Infof("node %s not found, skip force-delete pods", req.Name)
			return ctrl.Result{}, nil
		}
		hwlog.RunLog.Errorf("unable to fetch node %s: %v", req.Name, err)
		return ctrl.Result{}, err
	}

	if isNodeReady(node) {
		// node recovered or transient flip; nothing to do
		return ctrl.Result{}, nil
	}

	hwlog.RunLog.Infof("node %s is NotReady, start force-deleting infer-operator managed pods",
		req.Name)

	if err := r.forceDeletePodsOnNode(ctx, node.Name); err != nil {
		hwlog.RunLog.Errorf("force-delete pods on node %s failed: %v", node.Name, err)
		// requeue to retry; pods stuck on a dead node will keep blocking
		// StatefulSet recreation until we succeed. Return nil error so the
		// workqueue treats this as a normal requeue instead of an error.
		return ctrl.Result{RequeueAfter: common.DefaultReEnqueueInterval}, nil
	}

	hwlog.RunLog.Infof("force-delete finished for infer-operator managed pods on node %s",
		node.Name)
	return ctrl.Result{}, nil
}

// forceDeletePodsOnNode lists all infer-operator managed pods scheduled on the
// given node and force-deletes (grace-period=0) each of them. Deletion is
// idempotent (already-deleted pods return NotFound and are ignored).
func (r *NodePodCleanerReconciler) forceDeletePodsOnNode(ctx context.Context, nodeName string) error {
	podList := &corev1.PodList{}
	if err := r.List(ctx, podList,
		client.MatchingLabels{common.OperatorNameKey: common.TrueBool},
		client.InNamespace(""),
		client.MatchingFields{"spec.nodeName": nodeName}); err != nil {
		return fmt.Errorf("list pods on node %s failed: %w", nodeName, err)
	}

	if len(podList.Items) == 0 {
		hwlog.RunLog.Infof("no infer-operator managed pods on node %s, skip", nodeName)
		return nil
	}

	deleteOpts := []client.DeleteOption{
		client.GracePeriodSeconds(0),
		client.PropagationPolicy(metav1.DeletePropagationBackground),
	}

	var failed int
	for i := range podList.Items {
		pod := &podList.Items[i]
		if err := r.Delete(ctx, pod, deleteOpts...); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			hwlog.RunLog.Errorf("force-delete pod %s/%s on node %s failed: %v",
				pod.Namespace, pod.Name, nodeName, err)
			failed++
			continue
		}
		hwlog.RunLog.Infof("force-deleted pod %s/%s on node %s",
			pod.Namespace, pod.Name, nodeName)
	}

	if failed > 0 {
		return fmt.Errorf("%d pods failed to force-delete on node %s", failed, nodeName)
	}
	return nil
}

// isNodeReady returns true only when the node has a Ready condition with status
// True. Any other state (NotReady, Unknown, or no condition) is treated as not
// ready.
func isNodeReady(node *corev1.Node) bool {
	for i := range node.Status.Conditions {
		c := &node.Status.Conditions[i]
		if c.Type == corev1.NodeReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

// nodeNotReadyPredicate only lets through node events where the node turns
// NotReady. Create events pass when the node is already NotReady at creation;
// update events pass only on the Ready->NotReady transition. This keeps the
// work queue focused on nodes that need force-delete.
func nodeNotReadyPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			node, ok := e.Object.(*corev1.Node)
			if !ok {
				return false
			}
			return !isNodeReady(node)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldNode, okOld := e.ObjectOld.(*corev1.Node)
			newNode, okNew := e.ObjectNew.(*corev1.Node)
			if !okOld || !okNew {
				return false
			}
			// only enqueue when node transitions into NotReady
			return isNodeReady(oldNode) && !isNodeReady(newNode)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

// SetupWithManager sets up the controller with the Manager. It only watches
// Nodes; when a node transitions to NotReady, the reconciler lists and
// force-deletes the affected pods on demand. A field index on spec.nodeName is
// registered so the pod list can be filtered server-side.
func (r *NodePodCleanerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(),
		&corev1.Pod{}, "spec.nodeName", func(obj client.Object) []string {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return nil
			}
			if pod.Spec.NodeName == "" {
				return nil
			}
			return []string{pod.Spec.NodeName}
		}); err != nil {
		return fmt.Errorf("register pod nodeName field index failed: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}, builder.WithPredicates(nodeNotReadyPredicate())).
		Named("node-pod-cleaner-controller").
		Complete(r)
}
