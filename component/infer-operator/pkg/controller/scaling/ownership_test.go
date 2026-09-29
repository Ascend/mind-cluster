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

package scaling

import (
	"context"
	"strings"
	"testing"
	"time"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	apiv1 "infer-operator/pkg/api/v1"
)

func TestOwnershipResolver(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := apiv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := autoscalingv2.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	target := &apiv1.InstanceSet{ObjectMeta: metav1.ObjectMeta{Name: "target", Namespace: "default"}}
	older := metav1.NewTime(time.Unix(1, 0))
	newer := metav1.NewTime(time.Unix(2, 0))
	first := testPA("first", older)
	second := testPA("second", newer)
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(target, second, first).Build()
	resolver := NewOwnershipResolver(client)
	allowed, message, err := resolver.CanManageWithKPA(context.Background(), second)
	if err != nil || allowed || !strings.Contains(message, "first") {
		t.Fatalf("allowed=%v message=%q err=%v", allowed, message, err)
	}
	allowed, _, err = resolver.CanManageWithKPA(context.Background(), first)
	if err != nil || !allowed {
		t.Fatalf("oldest KPA should win: allowed=%v err=%v", allowed, err)
	}
	allowed, message, err = resolver.CanManageWithHPA(context.Background(), target)
	if err != nil || allowed || !strings.Contains(message, "first") {
		t.Fatalf("HPA should be blocked: %v %q %v", allowed, message, err)
	}
}

func TestExistingHPAWins(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = apiv1.AddToScheme(scheme)
	_ = autoscalingv2.AddToScheme(scheme)
	pa := testPA("kpa", metav1.Now())
	hpa := &autoscalingv2.HorizontalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: "target-scaler", Namespace: "default"},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
			APIVersion: apiv1.GroupVersion.String(), Kind: "InstanceSet", Name: "target"}}}
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pa, hpa).Build()
	allowed, message, err := NewOwnershipResolver(client).CanManageWithKPA(context.Background(), pa)
	if err != nil || allowed || !strings.Contains(message, "target-scaler") {
		t.Fatalf("allowed=%v message=%q err=%v", allowed, message, err)
	}
}

func TestKPANameBreaksCreationTimestampTie(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = apiv1.AddToScheme(scheme)
	_ = autoscalingv2.AddToScheme(scheme)
	created := metav1.NewTime(time.Unix(1, 0))
	first := testPA("a-owner", created)
	second := testPA("z-candidate", created)
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(second, first).Build()

	allowed, message, err := NewOwnershipResolver(client).CanManageWithKPA(context.Background(), second)
	if err != nil || allowed || !strings.Contains(message, first.Name) {
		t.Fatalf("same-time owner was not selected deterministically: allowed=%v message=%q err=%v",
			allowed, message, err)
	}
}

func testPA(name string, created metav1.Time) *apiv1.PodAutoscaler {
	return &apiv1.PodAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default", CreationTimestamp: created},
		Spec: apiv1.PodAutoscalerSpec{ScalingStrategy: apiv1.ScalingStrategyKPA,
			ScaleTargetRef: apiv1.ScaleTargetRef{APIVersion: apiv1.GroupVersion.String(), Kind: "InstanceSet", Name: "target"}}}
}
