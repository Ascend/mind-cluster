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
	"errors"
	"fmt"
	"time"

	apiv1 "infer-operator/pkg/api/v1"
	"infer-operator/pkg/autoscaling/algorithm"
	autoscalingcontext "infer-operator/pkg/autoscaling/context"
	"infer-operator/pkg/autoscaling/metrics"
	autoscalingtypes "infer-operator/pkg/autoscaling/types"
)

const (
	defaultStableWindow = 180 * time.Second
	defaultPanicWindow  = 60 * time.Second
)

// ReplicaComputeRequest contains one reconciliation's immutable inputs.
type ReplicaComputeRequest struct {
	PodAutoscaler   *apiv1.PodAutoscaler
	Evaluation      *autoscalingcontext.EvaluationContext
	CurrentReplicas int32
	Observations    []autoscalingtypes.MetricObservation
	Timestamp       time.Time
}

// ReplicaComputeResult reports the selected recommendation and its evidence.
type ReplicaComputeResult struct {
	DesiredReplicas int32
	Algorithm       string
	Reason          string
	MetricValues    map[string]float64
	MetricDecisions []MetricDecision
	SelectedMetric  string
	Valid           bool
}

// RecommendationEngine computes replica recommendations and manages the
// metric-window state associated with each PodAutoscaler.
type RecommendationEngine interface {
	Recommend(context.Context, ReplicaComputeRequest) (*ReplicaComputeResult, error)
	Forget(namespace, name string)
}

// MetricDecision records the observable stages for one metric.
type MetricDecision struct {
	Name                string
	CurrentValue        float64
	StableValue         float64
	PanicValue          float64
	TargetValue         float64
	Mode                string
	RawReplicas         int32
	RateLimitedReplicas int32
	BoundedReplicas     int32
}

// Engine owns only metric-window state and replica recommendation logic.
// Pod discovery and metric endpoint access happen before data reaches it.
type Engine struct{ windows *metrics.MetricsClient }

var _ RecommendationEngine = (*Engine)(nil)

// NewEngine creates an in-memory recommendation engine.
func NewEngine() *Engine { return &Engine{windows: metrics.NewMetricsClient(time.Second)} }

// Recommend evaluates all metric sources. Invalid sources are collected as
// errors while successful sources continue to participate in the decision.
func (engine *Engine) Recommend(ctx context.Context,
	request ReplicaComputeRequest) (*ReplicaComputeResult, error) {
	if err := validateRequest(request); err != nil {
		return &ReplicaComputeResult{}, err
	}

	result := &ReplicaComputeResult{MetricValues: make(map[string]float64)}
	failures := make([]error, 0)
	for _, observation := range request.Observations {
		recommendation, err := engine.evaluateObservation(ctx, request, observation)
		if err != nil {
			failures = append(failures, fmt.Errorf("metric %q: %w", observation.Name, err))
			continue
		}
		result.record(observation.Name, recommendation)
	}
	if !result.Valid {
		return result, fmt.Errorf("all metric sources failed: %w", errors.Join(failures...))
	}
	return result, nil
}

func validateRequest(request ReplicaComputeRequest) error {
	if request.PodAutoscaler == nil || request.Evaluation == nil {
		return fmt.Errorf("PodAutoscaler and evaluation context are required")
	}
	if len(request.Observations) == 0 {
		return fmt.Errorf("metric observations are empty")
	}
	return nil
}

func (engine *Engine) evaluateObservation(ctx context.Context, request ReplicaComputeRequest,
	observation autoscalingtypes.MetricObservation) (*algorithm.ScalingRecommendation, error) {
	key := metricKey(request.PodAutoscaler, observation.Name)
	stableWindow, panicWindow := MetricWindows(request.PodAutoscaler)
	if err := engine.windows.UpdateMetrics(observation.Timestamp, key,
		stableWindow, panicWindow, observation.Values...); err != nil {
		return nil, fmt.Errorf("update metric windows: %w", err)
	}
	stableValue, panicValue, err := engine.windows.GetMetricValues(key)
	if err != nil {
		return nil, fmt.Errorf("read metric windows: %w", err)
	}
	aggregated := &autoscalingtypes.AggregatedMetrics{MetricKey: key,
		CurrentValue: average(observation.Values), StableValue: stableValue,
		PanicValue: panicValue, Confidence: 1, LastUpdated: observation.Timestamp}

	recommendation, err := (&algorithm.KPAAlgorithm{}).ComputeRecommendation(ctx, algorithm.ScalingRequest{
		Target: autoscalingtypes.ScaleTarget{
			Namespace: request.PodAutoscaler.Namespace, Name: request.PodAutoscaler.Spec.ScaleTargetRef.Name,
			Kind:       request.PodAutoscaler.Spec.ScaleTargetRef.Kind,
			APIVersion: request.PodAutoscaler.Spec.ScaleTargetRef.APIVersion, MetricKey: key,
		},
		CurrentReplicas: request.CurrentReplicas, AggregatedMetrics: aggregated,
		Policy: &request.Evaluation.Policy, RuntimeState: request.Evaluation.StateFor(observation.Name),
		Timestamp: request.Timestamp,
	})
	if err != nil {
		return nil, fmt.Errorf("compute recommendation: %w", err)
	}
	if !recommendation.ScaleValid {
		return nil, fmt.Errorf("invalid scaling recommendation")
	}
	return recommendation, nil
}

func metricKey(pa *apiv1.PodAutoscaler, metricName string) autoscalingtypes.MetricKey {
	return autoscalingtypes.MetricKey{Namespace: pa.Namespace, Name: pa.Spec.ScaleTargetRef.Name,
		MetricName: metricName, PANamespace: pa.Namespace, PAName: pa.Name}
}

func (result *ReplicaComputeResult) record(metricName string,
	recommendation *algorithm.ScalingRecommendation) {
	result.MetricValues[metricName] = recommendation.MetricValue
	result.MetricDecisions = append(result.MetricDecisions, MetricDecision{
		Name: metricName, CurrentValue: recommendation.CurrentValue,
		StableValue: recommendation.StableValue, PanicValue: recommendation.PanicValue,
		TargetValue: recommendation.TargetValue, Mode: recommendation.Mode,
		RawReplicas:         recommendation.RawDesiredReplicas,
		RateLimitedReplicas: recommendation.RateLimitedReplicas,
		BoundedReplicas:     recommendation.DesiredReplicas,
	})
	if result.Valid && recommendation.DesiredReplicas <= result.DesiredReplicas {
		return
	}
	result.Valid = true
	result.DesiredReplicas = recommendation.DesiredReplicas
	result.Algorithm = recommendation.Algorithm
	result.Reason = recommendation.Reason
	result.SelectedMetric = metricName
}

func average(values []float64) float64 {
	var sum float64
	for _, value := range values {
		sum += value
	}
	if len(values) == 0 {
		return 0
	}
	return sum / float64(len(values))
}

// Forget removes metric history owned by one PodAutoscaler.
func (engine *Engine) Forget(namespace, name string) {
	engine.windows.DeleteForPodAutoscaler(namespace, name)
}

// MetricWindows returns the configured stable and panic aggregation windows.
func MetricWindows(pa *apiv1.PodAutoscaler) (time.Duration, time.Duration) {
	stable, panicDuration := defaultStableWindow, defaultPanicWindow
	if pa.Spec.StableWindowSeconds != nil {
		stable = time.Duration(*pa.Spec.StableWindowSeconds) * time.Second
	}
	if pa.Spec.PanicWindowSeconds != nil {
		panicDuration = time.Duration(*pa.Spec.PanicWindowSeconds) * time.Second
	}
	return stable, panicDuration
}
