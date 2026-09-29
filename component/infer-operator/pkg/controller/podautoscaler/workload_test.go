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

package podautoscaler

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	apiv1 "infer-operator/pkg/api/v1"
)

func TestInstanceSetScale(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := apiv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	replicas := int32(2)
	target := &apiv1.InstanceSet{ObjectMeta: metav1.ObjectMeta{Name: "target", Namespace: "default"},
		Spec: apiv1.InstanceSetSpec{Replicas: &replicas}, Status: apiv1.InstanceSetStatus{LabelSelector: "app=demo"}}
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(target).Build()
	scale := NewInstanceSetScale(client)
	pa := &apiv1.PodAutoscaler{ObjectMeta: metav1.ObjectMeta{Namespace: "default"},
		Spec: apiv1.PodAutoscalerSpec{ScaleTargetRef: apiv1.ScaleTargetRef{Name: "target"}}}
	loaded, err := scale.GetTarget(context.Background(), pa)
	if err != nil {
		t.Fatal(err)
	}
	current, err := scale.GetCurrentReplicas(loaded)
	if err != nil || current != 2 {
		t.Fatalf("current=%d err=%v", current, err)
	}
	selector, err := scale.GetPodSelector(loaded)
	if err != nil || !selector.Matches(labels.Set{"app": "demo"}) {
		t.Fatalf("selector=%v err=%v", selector, err)
	}
	if err := scale.UpdateReplicas(context.Background(), pa, 4); err != nil {
		t.Fatal(err)
	}
	updated, _ := scale.GetTarget(context.Background(), pa)
	if updated.Spec.Replicas == nil || *updated.Spec.Replicas != 4 {
		t.Fatalf("replicas were not patched")
	}
}
