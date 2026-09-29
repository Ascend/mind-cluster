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

package algorithm

import (
	"context"
	"testing"

	apiv1 "infer-operator/pkg/api/v1"
	autoscalingcontext "infer-operator/pkg/autoscaling/context"
	autoscalingtypes "infer-operator/pkg/autoscaling/types"
)

func TestKPAReplicaFormulaAndLimits(t *testing.T) {
	tests := []struct {
		name          string
		current       int32
		stable, panic float64
		want          int32
	}{
		{name: "proportional scale up", current: 2, stable: 15, panic: 15, want: 3},
		{name: "scale up rate limit", current: 2, stable: 30, panic: 30, want: 4},
		{name: "scale down", current: 2, stable: 4, panic: 4, want: 1},
		{name: "tolerance", current: 2, stable: 10.5, panic: 10.5, want: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			evaluation := autoscalingcontext.NewEvaluationContext()
			pa := &apiv1.PodAutoscaler{Spec: apiv1.PodAutoscalerSpec{MinReplicas: 1, MaxReplicas: 20,
				MetricsSources: []apiv1.MetricSource{{TargetName: "qps", TargetValue: "10"}}}}
			if err := evaluation.Refresh(pa); err != nil {
				t.Fatal(err)
			}
			result, err := (&KPAAlgorithm{}).ComputeRecommendation(context.Background(), ScalingRequest{CurrentReplicas: test.current,
				Policy: &evaluation.Policy, RuntimeState: evaluation.StateFor("qps"),
				AggregatedMetrics: &autoscalingtypes.AggregatedMetrics{
					MetricKey: autoscalingtypes.MetricKey{MetricName: "qps"}, StableValue: test.stable, PanicValue: test.panic}})
			if err != nil {
				t.Fatal(err)
			}
			if result.DesiredReplicas != test.want {
				t.Fatalf("got %d replicas, want %d", result.DesiredReplicas, test.want)
			}
		})
	}
}

func TestKPAPanicUsesPanicValue(t *testing.T) {
	evaluation := autoscalingcontext.NewEvaluationContext()
	pa := &apiv1.PodAutoscaler{Spec: apiv1.PodAutoscalerSpec{MinReplicas: 1, MaxReplicas: 20,
		MetricsSources: []apiv1.MetricSource{{TargetName: "qps", TargetValue: "10"}}}}
	_ = evaluation.Refresh(pa)
	result, err := (&KPAAlgorithm{}).ComputeRecommendation(context.Background(), ScalingRequest{CurrentReplicas: 2,
		Policy: &evaluation.Policy, RuntimeState: evaluation.StateFor("qps"),
		AggregatedMetrics: &autoscalingtypes.AggregatedMetrics{
			MetricKey: autoscalingtypes.MetricKey{MetricName: "qps"}, CurrentValue: 30,
			StableValue: 10, PanicValue: 25}})
	if err != nil {
		t.Fatal(err)
	}
	if result.DesiredReplicas != 4 || !result.ScaleValid || result.Algorithm != "KPA" ||
		result.Metadata["mode"] != "panic" || !evaluation.StateFor("qps").PanicMode {
		t.Fatalf("unexpected panic recommendation: %+v", result)
	}
	if result.CurrentValue != 30 || result.StableValue != 10 || result.PanicValue != 25 ||
		result.TargetValue != 10 || result.Mode != "panic" || result.RawDesiredReplicas != 5 ||
		result.RateLimitedReplicas != 4 {
		t.Fatalf("unexpected observable decision stages: %+v", result)
	}
}

func TestKPAAlgorithmType(t *testing.T) {
	if algorithmType := (&KPAAlgorithm{}).GetAlgorithmType(); algorithmType != "KPA" {
		t.Fatalf("algorithm type=%q, want KPA", algorithmType)
	}
}
