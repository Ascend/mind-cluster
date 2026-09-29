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

package autoscaler

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "infer-operator/pkg/api/v1"
	autoscalingcontext "infer-operator/pkg/autoscaling/context"
	autoscalingtypes "infer-operator/pkg/autoscaling/types"
)

func TestComputeUsesHighestMetricRecommendation(t *testing.T) {
	engine := NewEngine()
	pa := &apiv1.PodAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: "pa", Namespace: "default"}, Spec: apiv1.PodAutoscalerSpec{
		ScaleTargetRef: apiv1.ScaleTargetRef{APIVersion: apiv1.GroupVersion.String(), Kind: "InstanceSet", Name: "target"},
		MinReplicas:    1, MaxReplicas: 10, ScalingStrategy: apiv1.ScalingStrategyKPA,
		MetricsSources: []apiv1.MetricSource{{TargetName: "qps", TargetValue: "10", Port: "8000"},
			{TargetName: "waiting", TargetValue: "10", Port: "8000"}}}}
	evaluation := autoscalingcontext.NewEvaluationContext()
	if err := evaluation.Refresh(pa); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	result, err := engine.Recommend(context.Background(), ReplicaComputeRequest{PodAutoscaler: pa,
		Evaluation: evaluation, CurrentReplicas: 2, Timestamp: now,
		Observations: []autoscalingtypes.MetricObservation{
			{Name: "qps", Values: []float64{15}, Timestamp: now},
			{Name: "waiting", Values: []float64{25}, Timestamp: now},
		}})
	if err != nil {
		t.Fatal(err)
	}
	if result.DesiredReplicas != 4 || result.Algorithm != "KPA" || !result.Valid || len(result.MetricValues) != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.SelectedMetric != "waiting" || len(result.MetricDecisions) != 2 ||
		result.MetricDecisions[0].CurrentValue != 15 || result.MetricDecisions[1].CurrentValue != 25 {
		t.Fatalf("unexpected metric decisions: %+v", result.MetricDecisions)
	}
	engine.Forget("default", "pa")
}

func TestComputeRejectsInvalidRequest(t *testing.T) {
	engine := NewEngine()
	if _, err := engine.Recommend(context.Background(), ReplicaComputeRequest{}); err == nil {
		t.Fatal("empty request accepted")
	}
	pa := &apiv1.PodAutoscaler{}
	if _, err := engine.Recommend(context.Background(), ReplicaComputeRequest{PodAutoscaler: pa,
		Evaluation: autoscalingcontext.NewEvaluationContext()}); err == nil {
		t.Fatal("empty observations accepted")
	}
}
