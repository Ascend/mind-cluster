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

package context

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "infer-operator/pkg/api/v1"
	autoscalingtypes "infer-operator/pkg/autoscaling/types"
)

func TestContextConfigurationAndState(t *testing.T) {
	pa := &apiv1.PodAutoscaler{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
		autoscalingtypes.MaxScaleUpRateLabel: "3", autoscalingtypes.MaxScaleDownRateLabel: "4",
		autoscalingtypes.PanicThresholdLabel: "1.5", autoscalingtypes.ScaleUpToleranceLabel: "0.2",
		autoscalingtypes.ScaleDownToleranceLabel: "0.3", autoscalingtypes.ScaleUpCooldownWindowLabel: "10s",
		autoscalingtypes.ScaleDownCooldownWindowLabel: "20s"}}, Spec: apiv1.PodAutoscalerSpec{
		MinReplicas: 1, MaxReplicas: 9,
		MetricsSources: []apiv1.MetricSource{{TargetName: "qps", TargetValue: "12"}}}}
	evaluation := NewEvaluationContext()
	if err := evaluation.Refresh(pa); err != nil {
		t.Fatal(err)
	}
	policy := evaluation.Policy
	value, found := policy.TargetValue("qps")
	if !found || value != 12 || policy.MaxScaleUpRate != 3 || policy.MaxScaleDownRate != 4 ||
		policy.PanicThreshold != 1.5 || policy.ScaleUpTolerance != 0.2 ||
		policy.ScaleDownTolerance != 0.3 || policy.ScaleUpCooldownWindow != 10*time.Second ||
		policy.ScaleDownCooldownWindow != 20*time.Second {
		t.Fatalf("unexpected policy: %+v", policy)
	}
	state := evaluation.StateFor("qps")
	*state = RuntimeState{StableValue: 2, PanicValue: 3, PanicMode: true, MaxPanicReplicas: 5}
	if state.StableValue != 2 || state.PanicValue != 3 ||
		!state.PanicMode || state.MaxPanicReplicas != 5 {
		t.Fatal("runtime state was not retained")
	}
}

func TestContextRejectsInvalidConfiguration(t *testing.T) {
	evaluation := NewEvaluationContext()
	pa := &apiv1.PodAutoscaler{Spec: apiv1.PodAutoscalerSpec{MetricsSources: []apiv1.MetricSource{{TargetName: "qps", TargetValue: "bad"}}}}
	if err := evaluation.Refresh(pa); err == nil {
		t.Fatal("invalid target accepted")
	}
	pa.Spec.MinReplicas, pa.Spec.MaxReplicas = 1, 2
	pa.Spec.MetricsSources[0].TargetValue = "1"
	pa.Annotations = map[string]string{autoscalingtypes.MaxScaleUpRateLabel: "bad"}
	if err := evaluation.Refresh(pa); err == nil {
		t.Fatal("invalid annotation accepted")
	}
}

func TestRefreshRestoresDefaultsAfterAnnotationRemoval(t *testing.T) {
	evaluation := NewEvaluationContext()
	pa := &apiv1.PodAutoscaler{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
		autoscalingtypes.MaxScaleUpRateLabel: "5"}}, Spec: apiv1.PodAutoscalerSpec{
		MinReplicas: 1, MaxReplicas: 10,
		MetricsSources: []apiv1.MetricSource{{TargetName: "qps", TargetValue: "1"}}}}
	if err := evaluation.Refresh(pa); err != nil {
		t.Fatal(err)
	}
	evaluation.StateFor("qps").PanicMode = true
	delete(pa.Annotations, autoscalingtypes.MaxScaleUpRateLabel)
	if err := evaluation.Refresh(pa); err != nil {
		t.Fatal(err)
	}
	if evaluation.Policy.MaxScaleUpRate != defaultScaleRate {
		t.Fatalf("scale-up rate=%v, want default %v", evaluation.Policy.MaxScaleUpRate, defaultScaleRate)
	}
	if !evaluation.StateFor("qps").PanicMode {
		t.Fatal("refresh discarded the metric runtime state")
	}
}
