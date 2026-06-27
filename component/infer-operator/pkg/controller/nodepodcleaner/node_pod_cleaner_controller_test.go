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

package nodepodcleaner

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/common"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

// newFakeClient builds a fake client with corev1 scheme and a spec.nodeName
// index that mirrors the one registered in SetupWithManager.
func newFakeClient(objs ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	builder := fake.NewClientBuilder().
		WithScheme(scheme).
		WithIndex(&corev1.Pod{}, "spec.nodeName", func(obj client.Object) []string {
			pod, ok := obj.(*corev1.Pod)
			if !ok || pod.Spec.NodeName == "" {
				return nil
			}
			return []string{pod.Spec.NodeName}
		})
	for _, o := range objs {
		builder = builder.WithObjects(o)
	}
	return builder.Build()
}

func readyNode(name string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{
			{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
		}},
	}
}

func notReadyNode(name string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{
			{Type: corev1.NodeReady, Status: corev1.ConditionFalse},
		}},
	}
}

// podOnNode builds a pod managed by infer-operator and bound to nodeName.
func podOnNode(name, ns, nodeName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    map[string]string{common.OperatorNameKey: common.TrueBool},
		},
		Spec: corev1.PodSpec{NodeName: nodeName},
	}
}

// podWithoutOperatorLabel builds a pod not managed by this operator.
func podWithoutOperatorLabel(name, ns, nodeName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec:       corev1.PodSpec{NodeName: nodeName},
	}
}

// pendingPod is managed by the operator but not scheduled yet.
func pendingPod(name, ns string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    map[string]string{common.OperatorNameKey: common.TrueBool},
		},
		Spec: corev1.PodSpec{NodeName: ""},
	}
}

// TestIsNodeReady covers the ready/not-ready/unknown conditions.
func TestIsNodeReady(t *testing.T) {
	convey.Convey("isNodeReady", t, func() {
		convey.Convey("ready node returns true", func() {
			convey.So(isNodeReady(readyNode("n1")), convey.ShouldBeTrue)
		})
		convey.Convey("notReady node returns false", func() {
			convey.So(isNodeReady(notReadyNode("n1")), convey.ShouldBeFalse)
		})
		convey.Convey("node without conditions returns false", func() {
			n := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1"}}
			convey.So(isNodeReady(n), convey.ShouldBeFalse)
		})
		convey.Convey("unknown status returns false", func() {
			n := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1"},
				Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{
					{Type: corev1.NodeReady, Status: corev1.ConditionUnknown},
				}}}
			convey.So(isNodeReady(n), convey.ShouldBeFalse)
		})
	})
}

// TestNodeNotReadyPredicate covers create/update/delete filtering.
func TestNodeNotReadyPredicate(t *testing.T) {
	p := nodeNotReadyPredicate()

	convey.Convey("predicate", t, func() {
		convey.Convey("create notReady node passes", func() {
			e := event.CreateEvent{Object: notReadyNode("n1")}
			convey.So(p.Create(e), convey.ShouldBeTrue)
		})
		convey.Convey("create ready node is filtered", func() {
			e := event.CreateEvent{Object: readyNode("n1")}
			convey.So(p.Create(e), convey.ShouldBeFalse)
		})
		convey.Convey("ready->notReady transition passes", func() {
			e := event.UpdateEvent{ObjectOld: readyNode("n1"), ObjectNew: notReadyNode("n1")}
			convey.So(p.Update(e), convey.ShouldBeTrue)
		})
		convey.Convey("notReady->ready transition is filtered", func() {
			e := event.UpdateEvent{ObjectOld: notReadyNode("n1"), ObjectNew: readyNode("n1")}
			convey.So(p.Update(e), convey.ShouldBeFalse)
		})
		convey.Convey("notReady->notReady stays filtered", func() {
			e := event.UpdateEvent{ObjectOld: notReadyNode("n1"), ObjectNew: notReadyNode("n1")}
			convey.So(p.Update(e), convey.ShouldBeFalse)
		})
		convey.Convey("delete is always filtered", func() {
			e := event.DeleteEvent{Object: notReadyNode("n1")}
			convey.So(p.Delete(e), convey.ShouldBeFalse)
		})
	})
}

// TestReconcileForceDeletePods verifies that a NotReady node triggers
// force-deletion of all operator-managed pods on it, while leaving pods on
// other nodes, non-operator pods, and pending pods untouched.
func TestReconcileForceDeletePods(t *testing.T) {
	convey.Convey("reconcile notReady node", t, func() {
		// pods on the dead node
		deadPod := podOnNode("p-dead", "ns1", "dead-node")
		otherNsPod := podOnNode("p-other-ns", "ns2", "dead-node")
		// non-operator pod on the dead node: must NOT be deleted
		alienPod := podWithoutOperatorLabel("p-alien", "ns1", "dead-node")
		// operator pod on a healthy node: must NOT be deleted
		healthyPod := podOnNode("p-healthy", "ns1", "healthy-node")
		// operator pod not scheduled yet: must NOT be deleted
		pending := pendingPod("p-pending", "ns1")

		cli := newFakeClient(notReadyNode("dead-node"),
			deadPod, otherNsPod, alienPod, healthyPod, pending)
		r := &NodePodCleanerReconciler{Client: cli}

		_, err := r.Reconcile(context.Background(),
			ctrl.Request{NamespacedName: types.NamespacedName{Name: "dead-node"}})
		convey.So(err, convey.ShouldBeNil)

		// operator pods on the dead node are gone (both namespaces)
		convey.So(cli.Get(context.Background(),
			client.ObjectKey{Namespace: "ns1", Name: "p-dead"}, &corev1.Pod{}),
			convey.ShouldNotBeNil)
		convey.So(cli.Get(context.Background(),
			client.ObjectKey{Namespace: "ns2", Name: "p-other-ns"}, &corev1.Pod{}),
			convey.ShouldNotBeNil)

		// non-operator pod is untouched
		convey.So(cli.Get(context.Background(),
			client.ObjectKey{Namespace: "ns1", Name: "p-alien"}, &corev1.Pod{}),
			convey.ShouldBeNil)
		// pod on healthy node is untouched
		convey.So(cli.Get(context.Background(),
			client.ObjectKey{Namespace: "ns1", Name: "p-healthy"}, &corev1.Pod{}),
			convey.ShouldBeNil)
		// pending pod is untouched
		convey.So(cli.Get(context.Background(),
			client.ObjectKey{Namespace: "ns1", Name: "p-pending"}, &corev1.Pod{}),
			convey.ShouldBeNil)
	})
}

// TestReconcileReadyNodeSkips verifies that a ready node triggers no deletion.
func TestReconcileReadyNodeSkips(t *testing.T) {
	convey.Convey("reconcile ready node", t, func() {
		p := podOnNode("p1", "ns1", "n1")
		cli := newFakeClient(readyNode("n1"), p)
		r := &NodePodCleanerReconciler{Client: cli}

		_, err := r.Reconcile(context.Background(),
			ctrl.Request{NamespacedName: types.NamespacedName{Name: "n1"}})
		convey.So(err, convey.ShouldBeNil)

		convey.So(cli.Get(context.Background(),
			client.ObjectKey{Namespace: "ns1", Name: "p1"}, &corev1.Pod{}),
			convey.ShouldBeNil)
	})
}

// TestReconcileNodeNotFound verifies that a missing node is a no-op.
func TestReconcileNodeNotFound(t *testing.T) {
	convey.Convey("reconcile missing node", t, func() {
		cli := newFakeClient()
		r := &NodePodCleanerReconciler{Client: cli}

		_, err := r.Reconcile(context.Background(),
			ctrl.Request{NamespacedName: types.NamespacedName{Name: "ghost"}})
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestReconcileNoPodsOnNode verifies that a notReady node with no operator
// pods does not error.
func TestReconcileNoPodsOnNode(t *testing.T) {
	convey.Convey("notReady node with no operator pods", t, func() {
		cli := newFakeClient(notReadyNode("empty-node"))
		r := &NodePodCleanerReconciler{Client: cli}

		_, err := r.Reconcile(context.Background(),
			ctrl.Request{NamespacedName: types.NamespacedName{Name: "empty-node"}})
		convey.So(err, convey.ShouldBeNil)
	})
}
