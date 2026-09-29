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
	"fmt"
	"math"
)

// KPAAlgorithm implements Knative-style autoscaling with stable and panic windows.
type KPAAlgorithm struct{}

var _ ScalingAlgorithm = (*KPAAlgorithm)(nil)

// ComputeRecommendation computes a KPA recommendation for one metric source.
// The observed value is a per-Pod average, so the replica ratio must be applied
// to the current replica count: ceil(currentReplicas * observed / target).
func (algorithm *KPAAlgorithm) ComputeRecommendation(_ context.Context,
	request ScalingRequest) (*ScalingRecommendation, error) {
	if request.AggregatedMetrics == nil || request.Policy == nil || request.RuntimeState == nil {
		return nil, fmt.Errorf("aggregated metrics, scaling policy and runtime state are required")
	}

	metricName := request.AggregatedMetrics.MetricKey.MetricName
	targetValue, found := request.Policy.TargetValue(metricName)
	if !found || targetValue <= 0 {
		return nil, fmt.Errorf("target value for metric %q is not configured", metricName)
	}

	stableValue := request.AggregatedMetrics.StableValue
	panicValue := request.AggregatedMetrics.PanicValue

	inPanic := panicValue >= targetValue*request.Policy.PanicThreshold

	request.RuntimeState.PanicMode = inPanic
	request.RuntimeState.StableValue = stableValue
	request.RuntimeState.PanicValue = panicValue

	observed := stableValue
	mode := "stable"
	if inPanic {
		observed = panicValue
		mode = "panic"
	}

	rawDesired := desiredReplicas(request.CurrentReplicas, observed, targetValue,
		request.Policy.ScaleUpTolerance, request.Policy.ScaleDownTolerance)
	rateLimited := applyRateLimits(request.CurrentReplicas, rawDesired,
		request.Policy.MaxScaleUpRate, request.Policy.MaxScaleDownRate)
	desired := rateLimited

	if inPanic {
		if desired < request.RuntimeState.MaxPanicReplicas {
			desired = request.RuntimeState.MaxPanicReplicas
		} else {
			request.RuntimeState.MaxPanicReplicas = desired
		}
	} else {
		request.RuntimeState.MaxPanicReplicas = 0
	}

	return &ScalingRecommendation{
		DesiredReplicas:     constrain(desired, *request.Policy),
		RawDesiredReplicas:  rawDesired,
		RateLimitedReplicas: rateLimited,
		Confidence:          request.AggregatedMetrics.Confidence,
		Reason:              fmt.Sprintf("%s mode scaling", mode),
		Algorithm:           algorithm.GetAlgorithmType(),
		Mode:                mode,
		MetricValue:         observed,
		CurrentValue:        request.AggregatedMetrics.CurrentValue,
		StableValue:         stableValue,
		PanicValue:          panicValue,
		TargetValue:         targetValue,
		ScaleValid:          true,
		Metadata: map[string]any{
			"mode":          mode,
			"stable_value":  stableValue,
			"panic_value":   panicValue,
			"current_value": request.AggregatedMetrics.CurrentValue,
		},
	}, nil
}

// GetAlgorithmType returns the strategy name represented by this implementation.
func (algorithm *KPAAlgorithm) GetAlgorithmType() string { return "KPA" }

func desiredReplicas(currentReplicas int32, observedValue, targetValue, upTolerance, downTolerance float64) int32 {
	upValue := targetValue * (1 + upTolerance)
	downValue := targetValue * (1 - downTolerance)

	if observedValue <= upValue && observedValue >= downValue {
		return currentReplicas
	}

	return int32(math.Ceil(float64(currentReplicas) * observedValue / targetValue))
}

func applyRateLimits(currentReplicas, desiredReplicas int32, maxUpRate, maxDownRate float64) int32 {
	if currentReplicas <= 0 {
		return desiredReplicas
	}
	maxReplicas := int32(math.Ceil(float64(currentReplicas) * maxUpRate))
	minReplicas := int32(math.Floor(float64(currentReplicas) / maxDownRate))
	if desiredReplicas > maxReplicas {
		return maxReplicas
	}
	if desiredReplicas < minReplicas {
		return minReplicas
	}
	return desiredReplicas
}
