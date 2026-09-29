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
	"time"

	autoscalingcontext "infer-operator/pkg/autoscaling/context"
	autoscalingtypes "infer-operator/pkg/autoscaling/types"
)

// ScalingRequest contains the metric state and policy used by a scaling algorithm.
type ScalingRequest struct {
	Target            autoscalingtypes.ScaleTarget
	CurrentReplicas   int32
	AggregatedMetrics *autoscalingtypes.AggregatedMetrics
	Policy            *autoscalingcontext.ScalingPolicy
	RuntimeState      *autoscalingcontext.RuntimeState
	Timestamp         time.Time
}

// ScalingRecommendation describes one algorithm decision for one metric source.
type ScalingRecommendation struct {
	DesiredReplicas     int32
	RawDesiredReplicas  int32
	RateLimitedReplicas int32
	Confidence          float64
	Reason              string
	Algorithm           string
	Mode                string
	MetricValue         float64
	CurrentValue        float64
	StableValue         float64
	PanicValue          float64
	TargetValue         float64
	ScaleValid          bool
	Metadata            map[string]any
}

// ScalingAlgorithm computes a replica recommendation from aggregated metrics.
type ScalingAlgorithm interface {
	ComputeRecommendation(context.Context, ScalingRequest) (*ScalingRecommendation, error)
	GetAlgorithmType() string
}

func constrain(value int32, policy autoscalingcontext.ScalingPolicy) int32 {
	if value < policy.MinReplicas {
		return policy.MinReplicas
	}
	if value > policy.MaxReplicas {
		return policy.MaxReplicas
	}
	return value
}
